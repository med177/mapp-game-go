package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func siegeDefenderRetreatTestState(safeOwner faction.FactionID, safeRelation faction.DiplomaticStance, vassal bool) (*state.GameState, *army.Army) {
	defenderFaction := &faction.Faction{ID: "defender"}
	safeFaction := &faction.Faction{ID: safeOwner}
	if vassal {
		safeFaction.OverlordID = defenderFaction.ID
	}
	target := &world.Region{
		ID: "sieged", OwnerID: string(defenderFaction.ID), Neighbors: []world.RegionID{"safe"},
		WorldX: 0, WorldY: 0,
	}
	safe := &world.Region{
		ID: "safe", OwnerID: string(safeOwner), Neighbors: []world.RegionID{"sieged"},
		WorldX: 10, WorldY: 0,
	}
	attacker := &army.Army{
		ID: "besieger", OwnerID: "besieger", RegionID: target.ID,
		Units: []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}, {TypeID: "infantry", CurrentHP: army.MaxUnitHP}},
	}
	defender := &army.Army{
		ID: "defender_army", OwnerID: string(defenderFaction.ID), RegionID: target.ID,
		Units: []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}}, MovePoints: 1,
	}
	factions := map[faction.FactionID]*faction.Faction{
		defenderFaction.ID: defenderFaction,
		"besieger":         {ID: "besieger"},
	}
	if safeOwner != defenderFaction.ID {
		factions[safeFaction.ID] = safeFaction
	}
	gs := &state.GameState{
		Regions:  map[world.RegionID]*world.Region{target.ID: target, safe.ID: safe},
		Factions: factions,
		Armies:   map[army.ArmyID]*army.Army{attacker.ID: attacker, defender.ID: defender},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Attack: 10, Morale: 10, HP: army.MaxUnitHP},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("besieger", defenderFaction.ID): {
				FactionA: "besieger", FactionB: defenderFaction.ID, Stance: faction.StanceWar,
			},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			target.ID: {RegionID: target.ID, AttackerArmyID: attacker.ID, AttackerFactionID: attacker.OwnerID},
		},
	}
	if safeRelation != "" {
		gs.Relations[faction.RelationKey(defenderFaction.ID, safeOwner)] = &faction.Relation{
			FactionA: defenderFaction.ID, FactionB: safeOwner, Stance: safeRelation,
		}
	}
	return gs, defender
}

func TestApplyRetreatAssignmentsMovesWeakSiegeDefenderToOwnRegion(t *testing.T) {
	gs, defender := siegeDefenderRetreatTestState("defender", "", false)
	ctx := &StrategicContext{
		FactionID:       "defender",
		ArmyAssignments: make(map[army.ArmyID]AIArmyAssignment),
		gs:              gs,
	}

	applyRetreatAssignments(ctx)

	assignment, ok := ctx.ArmyAssignments[defender.ID]
	if !ok || assignment.Role != AIArmyRoleRetreat || assignment.AnchorRegionID != "safe" {
		t.Fatalf("weak siege defender was not assigned to its own safe region: %#v", assignment)
	}
}

func TestApplyRetreatAssignmentsAcceptsVassalAndAlliedSafeRegions(t *testing.T) {
	tests := []struct {
		name      string
		safeOwner faction.FactionID
		relation  faction.DiplomaticStance
		vassal    bool
	}{
		{name: "vassal", safeOwner: "vassal", vassal: true},
		{name: "ally", safeOwner: "ally", relation: faction.StanceAllied},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gs, defender := siegeDefenderRetreatTestState(test.safeOwner, test.relation, test.vassal)
			ctx := &StrategicContext{
				FactionID:       "defender",
				ArmyAssignments: make(map[army.ArmyID]AIArmyAssignment),
				gs:              gs,
			}

			applyRetreatAssignments(ctx)

			assignment, ok := ctx.ArmyAssignments[defender.ID]
			if !ok || assignment.Role != AIArmyRoleRetreat || assignment.AnchorRegionID != "safe" {
				t.Fatalf("weak siege defender was not assigned to %s safe region: %#v", test.name, assignment)
			}
			if next := aiRetreatNextStep(ctx, defender); next != "safe" {
				t.Fatalf("retreat route did not reach %s region: got %q", test.name, next)
			}
		})
	}
}

func TestApplyRetreatAssignmentsKeepsStrongSiegeDefender(t *testing.T) {
	gs, defender := siegeDefenderRetreatTestState("defender", "", false)
	attacker := gs.Armies["besieger"]
	attacker.Units = attacker.Units[:1]
	ctx := &StrategicContext{
		FactionID:       "defender",
		ArmyAssignments: make(map[army.ArmyID]AIArmyAssignment),
		gs:              gs,
	}

	applyRetreatAssignments(ctx)

	if _, ok := ctx.ArmyAssignments[defender.ID]; ok {
		t.Fatal("a sufficiently strong siege defender was incorrectly assigned to retreat")
	}
}

func TestAIMovementIntoVassalRetreatRegionDoesNotConquerIt(t *testing.T) {
	gs, defender := siegeDefenderRetreatTestState("vassal", "", true)
	gs.Sieges = nil

	outcome := executeMoveWithNavalPatrolAndContact(gs, defender, "safe", "defender", false, false)

	if defender.RegionID != "safe" || outcome.step.Kind != TurnStepMove {
		t.Fatalf("expected movement into vassal region, got region=%q step=%s", defender.RegionID, outcome.step.Kind)
	}
	if gs.Regions["safe"].OwnerID != "vassal" {
		t.Fatalf("AI retreat incorrectly changed vassal region owner to %q", gs.Regions["safe"].OwnerID)
	}
}
