package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestOpenAirTargetIntentDialogChoosesAttackOrLanding(t *testing.T) {
	const (
		playerID = faction.FactionID("player")
		enemyID  = "enemy"
		targetID = world.RegionID("target")
	)
	attacker := &army.Army{
		ID:            "dragon_carrier",
		OwnerID:       string(playerID),
		RegionID:      "origin",
		MovePoints:    4,
		Units:         []army.Unit{{TypeID: "dragon"}},
		EmbarkedUnits: []army.Unit{{TypeID: "soldier"}},
	}
	defender := &army.Army{
		ID:       "enemy_army",
		OwnerID:  enemyID,
		RegionID: targetID,
		Units:    []army.Unit{{TypeID: "soldier"}},
	}
	r := &Renderer{gs: &state.GameState{
		PlayerFactionID: playerID,
		Regions: map[world.RegionID]*world.Region{
			targetID: {ID: targetID, OwnerID: string(playerID)},
		},
		Armies: map[army.ArmyID]*army.Army{
			attacker.ID: attacker,
			defender.ID: defender,
		},
		UnitTypes: map[string]*army.UnitType{
			"dragon": {ID: "dragon", Category: army.CategoryDragon},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(playerID, faction.FactionID(enemyID)): {
				FactionA: playerID,
				FactionB: faction.FactionID(enemyID),
				Stance:   faction.StanceWar,
			},
		},
	}}

	if !r.openAirTargetIntentDialog(attacker, r.gs.Regions[targetID], defender) {
		t.Fatal("expected the enemy land target to open an intent dialog")
	}
	if got := r.confirmDialog.acceptLabel; got != "Saldır" {
		t.Fatalf("attack choice label = %q, want Saldır", got)
	}
	if got := r.confirmDialog.pendingAction; got.Kind != ActionMoveArmy || got.TargetArmyID != defender.ID || got.TargetRegion != targetID {
		t.Fatalf("attack action = %#v, want targeted attack on %q", got, defender.ID)
	}
	if got := r.confirmDialog.thirdLabel; got != "Buraya İn" {
		t.Fatalf("landing choice label = %q, want Buraya İn", got)
	}
	if got := r.confirmDialog.thirdAction; got.Kind != ActionMoveArmy || !got.AirLanding || got.TargetRegion != targetID {
		t.Fatalf("landing action = %#v, want land-and-stay at %q", got, targetID)
	}
	if !r.confirmDialog.declineActs || r.confirmDialog.declineLabel != "İptal" {
		t.Fatalf("cancel choice = %#v, want an explicit cancel button", r.confirmDialog)
	}
}

func TestOpenAirTargetIntentDialogOpensForEmptyEnemyRegion(t *testing.T) {
	const (
		playerID = "player"
		enemyID  = "enemy"
		targetID = world.RegionID("enemy_land")
	)
	attacker := &army.Army{ID: "dragon", OwnerID: playerID, RegionID: "origin", Units: []army.Unit{{TypeID: "dragon"}}}
	r := &Renderer{gs: &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			targetID: {ID: targetID, OwnerID: enemyID},
		},
		UnitTypes: map[string]*army.UnitType{
			"dragon": {ID: "dragon", Category: army.CategoryDragon},
		},
	}}

	if !airTargetIntentRequired(r.gs, attacker, r.gs.Regions[targetID], nil) {
		t.Fatal("empty hostile region should require explicit air intent")
	}
	if !r.openAirTargetIntentDialog(attacker, r.gs.Regions[targetID], nil) {
		t.Fatal("empty enemy land target should open the air intent dialog")
	}
	if got := r.confirmDialog.acceptLabel; got != "Saldır" {
		t.Fatalf("attack choice label = %q, want Saldır", got)
	}
	if got := r.confirmDialog.thirdAction; got.Kind != ActionMoveArmy || !got.AirLanding || got.TargetRegion != targetID {
		t.Fatalf("landing action = %#v, want land-and-stay at %q", got, targetID)
	}
}

func TestOpenAirTargetIntentDialogDoesNotOfferAttackOnFriendlyArmy(t *testing.T) {
	const (
		playerID = "player"
		targetID = world.RegionID("target")
	)
	attacker := &army.Army{ID: "dragon", OwnerID: playerID, RegionID: "origin", Units: []army.Unit{{TypeID: "dragon"}}}
	friendly := &army.Army{ID: "ally", OwnerID: playerID, RegionID: targetID}
	r := &Renderer{gs: &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			targetID: {ID: targetID, OwnerID: playerID},
		},
		UnitTypes: map[string]*army.UnitType{
			"dragon": {ID: "dragon", Category: army.CategoryDragon},
		},
	}}

	if r.openAirTargetIntentDialog(attacker, r.gs.Regions[targetID], friendly) {
		t.Fatal("friendly army should not open an attack-versus-landing dialog")
	}
	if airTargetIntentRequired(r.gs, attacker, r.gs.Regions[targetID], friendly) {
		t.Fatal("friendly region and army should not require an air intent choice")
	}
	if r.confirmDialog.show {
		t.Fatal("friendly army unexpectedly opened a confirmation dialog")
	}
}

func TestOpenQueuedWarConfirmKeepsSelectedAirAttackTarget(t *testing.T) {
	const (
		playerID = faction.FactionID("player")
		enemyID  = faction.FactionID("enemy")
		targetID = world.RegionID("target")
	)
	attacker := &army.Army{ID: "dragon", OwnerID: string(playerID), RegionID: "origin", Units: []army.Unit{{TypeID: "dragon"}}}
	firstDefender := &army.Army{ID: "first_enemy", OwnerID: string(enemyID), RegionID: targetID}
	selectedDefender := &army.Army{ID: "selected_enemy", OwnerID: string(enemyID), RegionID: targetID}
	r := &Renderer{gs: &state.GameState{
		PlayerFactionID: playerID,
		AirspaceEnabled: true,
		Factions: map[faction.FactionID]*faction.Faction{
			playerID: {ID: playerID},
			enemyID:  {ID: enemyID},
		},
		Regions: map[world.RegionID]*world.Region{
			targetID: {ID: targetID, OwnerID: string(playerID)},
		},
		Armies: map[army.ArmyID]*army.Army{
			attacker.ID:         attacker,
			firstDefender.ID:    firstDefender,
			selectedDefender.ID: selectedDefender,
		},
		UnitTypes: map[string]*army.UnitType{
			"dragon": {ID: "dragon", Category: army.CategoryDragon},
		},
	}}

	if !r.OpenQueuedWarConfirm(attacker, r.gs.Regions[targetID], selectedDefender.ID, true) {
		t.Fatal("expected a war confirmation for the foreign target")
	}
	if got := r.warConfirm.pendingEnemy; got != selectedDefender.ID {
		t.Fatalf("queued war target = %q, want explicitly selected army %q", got, selectedDefender.ID)
	}
	if got := faction.FactionID(r.warConfirm.factionID); got != enemyID {
		t.Fatalf("war target faction = %q, want selected enemy army owner %q", got, enemyID)
	}
	if !r.warConfirm.airLanding {
		t.Fatal("queued war confirmation lost the selected land-and-stay intent")
	}
}
