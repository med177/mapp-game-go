package state

import (
	"testing"

	"mapp-game-go/internal/army"
)

func TestDismissCommanderRemovesAssignedCommanderFromState(t *testing.T) {
	commander := &army.Commander{ID: "commander_vassal_1", OwnerID: "player", Name: "Commander"}
	currentArmy := &army.Army{ID: "army_1", OwnerID: "player", Commander: commander}
	gs := &GameState{
		Armies:     map[army.ArmyID]*army.Army{currentArmy.ID: currentArmy},
		Commanders: map[string]*army.Commander{commander.ID: commander},
	}

	if !gs.DismissCommander(commander.ID) {
		t.Fatal("DismissCommander() rejected an existing commander")
	}
	if currentArmy.Commander != nil {
		t.Fatal("dismissed commander remained assigned to the army")
	}
	if commander.AssignedArmyID != "" {
		t.Fatalf("dismissed commander kept army assignment %q", commander.AssignedArmyID)
	}
	if _, exists := gs.Commanders[commander.ID]; exists {
		t.Fatal("dismissed commander remained in the canonical commander pool")
	}
}

func TestDismissCommanderSuppressesScenarioTemplateReappearance(t *testing.T) {
	commander := &army.Commander{ID: "commander_template", OwnerID: "player", Name: "Tarihsel Komutan"}
	gs := &GameState{
		Year:                    1300,
		PlayerFactionID:         "player",
		Commanders:              map[string]*army.Commander{commander.ID: commander},
		CommanderTemplates:      map[string][]*army.Commander{"player": {commander}},
		CommanderArrivalNotices: map[string]bool{commander.ID: true},
	}

	if !gs.DismissCommander(commander.ID) {
		t.Fatal("DismissCommander() rejected a template commander")
	}
	gs.SyncCommanderAvailability()
	if _, exists := gs.Commanders[commander.ID]; exists {
		t.Fatal("dismissed template commander reappeared during availability sync")
	}
	if !gs.DismissedCommanderIDs[commander.ID] {
		t.Fatal("dismissed template commander was not persisted as dismissed")
	}
}

func TestUpdateCommanderProfileChangesGeneratedCommanderOnlyProfile(t *testing.T) {
	commander := &army.Commander{
		ID:            "commander_vassal_2",
		OwnerID:       "vassal",
		Name:          "Commander",
		PortraitAsset: army.DefaultPortraitAsset,
		Experience:    300,
	}
	currentArmy := &army.Army{ID: "army_vassal", OwnerID: "vassal", Commander: commander}
	gs := &GameState{
		Armies:     map[army.ArmyID]*army.Army{currentArmy.ID: currentArmy},
		Commanders: map[string]*army.Commander{commander.ID: commander},
	}
	gs.TransferArmyOwnership(currentArmy, "player")
	if commander.OwnerID != "player" {
		t.Fatalf("transferred commander owner = %q", commander.OwnerID)
	}

	if !gs.IsGeneratedCommander(commander.ID) {
		t.Fatal("runtime commander was not identified as generated")
	}
	if !gs.UpdateCommanderProfile(commander.ID, "Köse Mihal", "commander_ottoman_kose.png") {
		t.Fatal("UpdateCommanderProfile() rejected a valid profile")
	}
	if commander.Name != "Köse Mihal" || commander.PortraitAsset != "commander_ottoman_kose.png" {
		t.Fatalf("profile was not updated: %#v", commander)
	}
	if commander.Experience != 300 {
		t.Fatal("profile update changed commander career progression")
	}
}
