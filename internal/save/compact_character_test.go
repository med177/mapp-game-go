package save

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestLoadScenarioDefinitionIncludesCouncilRolesAndTraits(t *testing.T) {
	scenarioPath := t.TempDir()
	dataDir := filepath.Join(scenarioPath, "data")
	if err := os.Mkdir(dataDir, 0o700); err != nil {
		t.Fatalf("scenario data dizini oluşturulamadı: %v", err)
	}
	files := map[string]string{
		"scenario.json": `{"id":"test","name":"Test","period":"medieval"}`,
		"character_council.json": `{
			"roles": [
				{"id":"military_responsible","name_tr":"Serasker"},
				{"id":"treasurer","name_tr":"Hazinedar"}
			]
		}`,
		"character_traits.json": `{
			"traits": {
				"administrator": {"name_tr":"İdareci"}
			}
		}`,
	}
	for filename, contents := range files {
		if err := os.WriteFile(filepath.Join(dataDir, filename), []byte(contents), 0o600); err != nil {
			t.Fatalf("%s yazılamadı: %v", filename, err)
		}
	}

	loaded, err := loadScenarioDefinition(scenarioPath)
	if err != nil {
		t.Fatalf("senaryo tanımı yüklenemedi: %v", err)
	}
	for roleID, want := range map[string]string{
		"military_responsible": "Serasker",
		"treasurer":            "Hazinedar",
	} {
		role, ok := loaded.CharacterCouncil.Role(roleID)
		if !ok || role.NameTR != want {
			t.Errorf("rol %q = %+v, found=%t; want name_tr %q", roleID, role, ok, want)
		}
	}
	if got := loaded.CharacterCouncil.Traits["administrator"].NameTR; got != "İdareci" {
		t.Errorf("administrator trait name = %q, want %q", got, "İdareci")
	}
}

func TestAdoptedNeutralCharacterOwnershipPersistsCompactAndLegacySaves(t *testing.T) {
	character := &army.Commander{
		ID: "chr_bronn", Type: army.CharacterTypeCharacter, OwnerID: "player",
		Name: "Bronn", Traits: []army.CharacterTrait{"mercenary"},
		AdoptionCostGold: 500, GoldUpkeep: 6,
	}
	gs := &state.GameState{
		ScenarioID:      "test_scenario",
		PlayerFactionID: faction.FactionID("player"),
		Commanders:      map[string]*army.Commander{character.ID: character},
	}

	compact := campaignSaveState{
		ScenarioID:      gs.ScenarioID,
		PlayerFactionID: gs.PlayerFactionID,
		Commanders:      cloneCommanders(gs.Commanders),
	}
	encoded, err := json.Marshal(compact)
	if err != nil {
		t.Fatalf("compact save serialize edilemedi: %v", err)
	}
	var decoded campaignSaveState
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("compact save deserialize edilemedi: %v", err)
	}
	loaded := &state.GameState{}
	applyCampaignSaveState(loaded, decoded)
	if got := loaded.Commanders[character.ID]; got == nil || got.OwnerID != "player" ||
		got.AdoptionCostGold != 500 || got.GoldUpkeep != 6 {
		t.Fatalf("compact save/load character owner = %#v, want player", got)
	}

	debug := makeDebugCampaignSaveState(gs)
	legacy := convertLegacyCampaignSaveState(debug)
	if got := legacy.Commanders[character.ID]; got == nil || got.OwnerID != "player" ||
		got.AdoptionCostGold != 500 || got.GoldUpkeep != 6 {
		t.Fatalf("debug/legacy save conversion character owner = %#v, want player", got)
	}
}

func TestCouncilRoleAssignmentPersistsCompactAndLegacySaves(t *testing.T) {
	treasurer := &army.Commander{
		ID: "cmd_tywin", Type: army.CharacterTypeAdministrator, OwnerID: "player",
		Name: "Tywin Lannister", AssignedCouncilRole: state.CouncilRoleTreasurer,
		CouncilLevel: 3, CouncilExperience: 300,
	}
	diplomat := &army.Commander{
		ID: "chr_varys", Type: army.CharacterTypeCharacter, OwnerID: "player",
		Name: "Varys", AssignedCouncilRole: state.CouncilRoleDiplomat,
	}
	headOfHouse := &army.Commander{
		ID: "chr_joffrey", Type: army.CharacterTypeCharacter, OwnerID: "player",
		Name: "Joffrey Baratheon", AssignedCouncilRole: state.CouncilRoleHeadOfHouse,
	}
	heir := &army.Commander{
		ID: "chr_tommen", Type: army.CharacterTypeCharacter, OwnerID: "player",
		Name: "Tommen Baratheon", AssignedCouncilRole: state.CouncilRoleHeir,
	}
	militaryResponsible := &army.Commander{
		ID: "chr_stannis", Type: army.CharacterTypeCharacter, OwnerID: "player",
		Name: "Stannis Baratheon", AssignedCouncilRole: state.CouncilRoleMilitaryResponsible,
	}
	gs := &state.GameState{
		ScenarioID:      "test_scenario",
		PlayerFactionID: faction.FactionID("player"),
		Commanders: map[string]*army.Commander{
			treasurer.ID: treasurer, diplomat.ID: diplomat, headOfHouse.ID: headOfHouse,
			heir.ID: heir, militaryResponsible.ID: militaryResponsible,
		},
	}

	compact := campaignSaveState{
		ScenarioID:      gs.ScenarioID,
		PlayerFactionID: gs.PlayerFactionID,
		Commanders:      cloneCommanders(gs.Commanders),
	}
	encoded, err := json.Marshal(compact)
	if err != nil {
		t.Fatalf("compact save serialize edilemedi: %v", err)
	}
	var decoded campaignSaveState
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("compact save deserialize edilemedi: %v", err)
	}
	loaded := &state.GameState{}
	applyCampaignSaveState(loaded, decoded)
	if got := loaded.Commanders[treasurer.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleTreasurer {
		t.Fatalf("compact save/load council role = %#v, want treasurer", got)
	} else if got.CouncilLevel != 3 || got.CouncilExperience != 300 {
		t.Fatalf("compact save/load council progression = level %d, xp %d; want level 3, xp 300", got.CouncilLevel, got.CouncilExperience)
	}
	if got := loaded.Commanders[diplomat.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleDiplomat {
		t.Fatalf("compact save/load council role = %#v, want diplomat", got)
	}
	if got := loaded.Commanders[headOfHouse.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleHeadOfHouse {
		t.Fatalf("compact save/load council role = %#v, want head of house", got)
	}
	if got := loaded.Commanders[heir.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleHeir {
		t.Fatalf("compact save/load council role = %#v, want heir", got)
	}
	if got := loaded.Commanders[militaryResponsible.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleMilitaryResponsible {
		t.Fatalf("compact save/load council role = %#v, want military responsible", got)
	}

	debug := makeDebugCampaignSaveState(gs)
	legacy := convertLegacyCampaignSaveState(debug)
	if got := legacy.Commanders[treasurer.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleTreasurer {
		t.Fatalf("debug/legacy save conversion council role = %#v, want treasurer", got)
	} else if got.CouncilLevel != 3 || got.CouncilExperience != 300 {
		t.Fatalf("debug/legacy council progression = level %d, xp %d; want level 3, xp 300", got.CouncilLevel, got.CouncilExperience)
	}
	if got := legacy.Commanders[diplomat.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleDiplomat {
		t.Fatalf("debug/legacy save conversion council role = %#v, want diplomat", got)
	}
	if got := legacy.Commanders[headOfHouse.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleHeadOfHouse {
		t.Fatalf("debug/legacy save conversion council role = %#v, want head of house", got)
	}
	if got := legacy.Commanders[heir.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleHeir {
		t.Fatalf("debug/legacy save conversion council role = %#v, want heir", got)
	}
	if got := legacy.Commanders[militaryResponsible.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleMilitaryResponsible {
		t.Fatalf("debug/legacy save conversion council role = %#v, want military responsible", got)
	}
}

func TestActiveCharacterMissionPersistsCompactAndLegacySaves(t *testing.T) {
	character := &army.Commander{
		ID: "chr_varys", Type: army.CharacterTypeCharacter, OwnerID: "player",
		Name: "Varys", AssignedCouncilRole: state.CouncilRoleDiplomat,
		ActiveMissionID: "diplomatic_envoy", ActiveMissionTargetID: "target_house",
		ActiveMissionTurnsRemaining: 1, ActiveMissionOutcomeRoll: 37, ActiveMissionOutcome: "partial",
		AdventureAbilities: []army.CharacterAbility{{
			ID: "fieldcraft", NameTR: "Saha Keşfi", Scope: "faction", Effect: "military_power", Modifier: 0.02,
		}},
		PendingAdventureAbilities: []army.CharacterAbility{{
			ID: "shadowcraft", NameTR: "Gölgede Ustalık", Scope: "army", Effect: "army_power", Modifier: 0.03,
		}},
	}
	gs := &state.GameState{
		ScenarioID:      "test_scenario",
		PlayerFactionID: faction.FactionID("player"),
		Commanders:      map[string]*army.Commander{character.ID: character},
	}

	compact := campaignSaveState{
		ScenarioID:      gs.ScenarioID,
		PlayerFactionID: gs.PlayerFactionID,
		Commanders:      cloneCommanders(gs.Commanders),
	}
	encoded, err := json.Marshal(compact)
	if err != nil {
		t.Fatalf("compact save serialize edilemedi: %v", err)
	}
	var decoded campaignSaveState
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("compact save deserialize edilemedi: %v", err)
	}
	loaded := &state.GameState{}
	applyCampaignSaveState(loaded, decoded)
	assertActiveMissionSaveState(t, loaded.Commanders[character.ID])

	debug := makeDebugCampaignSaveState(gs)
	legacy := convertLegacyCampaignSaveState(debug)
	assertActiveMissionSaveState(t, legacy.Commanders[character.ID])
}

func assertActiveMissionSaveState(t *testing.T, character *army.Commander) {
	t.Helper()
	if character == nil || character.AssignedCouncilRole != state.CouncilRoleDiplomat ||
		character.ActiveMissionID != "diplomatic_envoy" || character.ActiveMissionTargetID != "target_house" ||
		character.ActiveMissionTurnsRemaining != 1 || character.ActiveMissionOutcomeRoll != 37 ||
		character.ActiveMissionOutcome != "partial" ||
		len(character.AdventureAbilities) != 1 || character.AdventureAbilities[0].ID != "fieldcraft" ||
		character.AdventureAbilities[0].Scope != "faction" ||
		character.AdventureAbilities[0].Effect != "military_power" ||
		character.AdventureAbilities[0].Modifier != 0.02 ||
		len(character.PendingAdventureAbilities) != 1 || character.PendingAdventureAbilities[0].ID != "shadowcraft" ||
		character.PendingAdventureAbilities[0].Scope != "army" ||
		character.PendingAdventureAbilities[0].Effect != "army_power" ||
		character.PendingAdventureAbilities[0].Modifier != 0.03 {
		t.Fatalf("active mission/ability save state = %+v, want mission result roll and retained abilities", character)
	}
}
