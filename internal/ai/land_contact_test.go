package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func landContactTestState(stance faction.DiplomaticStance) (*state.GameState, *army.Army, *army.Army) {
	from := &world.Region{ID: "bursa", Terrain: world.TerrainPlain}
	target := &world.Region{ID: "bithynia", Terrain: world.TerrainPlain, OwnerID: "ottoman"}
	attacker := &army.Army{
		ID: "east_rome_army", OwnerID: "east_rome", RegionID: from.ID,
		Units: []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}}, MovePoints: 1,
	}
	defender := &army.Army{
		ID: "ottoman_army", OwnerID: "ottoman", RegionID: target.ID,
		Units: []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}}, MovePoints: 1,
	}
	return &state.GameState{
		PlayerFactionID: "ottoman",
		Regions:         map[world.RegionID]*world.Region{from.ID: from, target.ID: target},
		Armies:          map[army.ArmyID]*army.Army{attacker.ID: attacker, defender.ID: defender},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("east_rome", "ottoman"): {
				FactionA: "east_rome", FactionB: "ottoman", Stance: stance,
			},
		},
	}, attacker, defender
}

func TestAIMovementCreatesPlayerLandContactBeforeBattle(t *testing.T) {
	gs, attacker, defender := landContactTestState(faction.StanceWar)

	outcome := executeMoveWithNavalPatrolAndContact(gs, attacker, defender.RegionID, "east_rome", false, false)
	if outcome.step.Kind != TurnStepBattle {
		t.Fatalf("expected a battle step for the pending contact, got %s", outcome.step.Kind)
	}
	if gs.PendingLandContact == nil || gs.PendingLandContact.PlayerArmyID != defender.ID {
		t.Fatalf("expected the player army to receive a land-contact decision: %#v", gs.PendingLandContact)
	}
	if _, exists := gs.Armies[defender.ID]; !exists {
		t.Fatal("the defender was removed before the player could resolve contact")
	}
}

func TestAIMovementDoesNotAttackForeignArmyWithoutWar(t *testing.T) {
	gs, attacker, defender := landContactTestState(faction.StancePeace)

	executeMoveWithNavalPatrolAndContact(gs, attacker, defender.RegionID, "east_rome", false, false)
	if gs.PendingLandContact != nil {
		t.Fatal("a peaceful foreign army unexpectedly created a land contact")
	}
	if _, exists := gs.Armies[defender.ID]; !exists {
		t.Fatal("a peaceful foreign army was removed by the AI movement fallback")
	}
}
