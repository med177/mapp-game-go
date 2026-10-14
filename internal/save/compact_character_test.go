package save

import (
	"encoding/json"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestAdoptedNeutralCharacterOwnershipPersistsCompactAndLegacySaves(t *testing.T) {
	character := &army.Commander{
		ID: "chr_bronn", Type: army.CharacterTypeCharacter, OwnerID: "player",
		Name: "Bronn", Traits: []army.CharacterTrait{"mercenary"},
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
	if got := loaded.Commanders[character.ID]; got == nil || got.OwnerID != "player" {
		t.Fatalf("compact save/load character owner = %#v, want player", got)
	}

	debug := makeDebugCampaignSaveState(gs)
	legacy := convertLegacyCampaignSaveState(debug)
	if got := legacy.Commanders[character.ID]; got == nil || got.OwnerID != "player" {
		t.Fatalf("debug/legacy save conversion character owner = %#v, want player", got)
	}
}

func TestCouncilRoleAssignmentPersistsCompactAndLegacySaves(t *testing.T) {
	treasurer := &army.Commander{
		ID: "cmd_tywin", Type: army.CharacterTypeAdministrator, OwnerID: "player",
		Name: "Tywin Lannister", AssignedCouncilRole: state.CouncilRoleTreasurer,
	}
	diplomat := &army.Commander{
		ID: "chr_varys", Type: army.CharacterTypeCharacter, OwnerID: "player",
		Name: "Varys", AssignedCouncilRole: state.CouncilRoleDiplomat,
	}
	headOfHouse := &army.Commander{
		ID: "chr_joffrey", Type: army.CharacterTypeCharacter, OwnerID: "player",
		Name: "Joffrey Baratheon", AssignedCouncilRole: state.CouncilRoleHeadOfHouse,
	}
	gs := &state.GameState{
		ScenarioID:      "test_scenario",
		PlayerFactionID: faction.FactionID("player"),
		Commanders:      map[string]*army.Commander{treasurer.ID: treasurer, diplomat.ID: diplomat, headOfHouse.ID: headOfHouse},
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
	}
	if got := loaded.Commanders[diplomat.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleDiplomat {
		t.Fatalf("compact save/load council role = %#v, want diplomat", got)
	}
	if got := loaded.Commanders[headOfHouse.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleHeadOfHouse {
		t.Fatalf("compact save/load council role = %#v, want head of house", got)
	}

	debug := makeDebugCampaignSaveState(gs)
	legacy := convertLegacyCampaignSaveState(debug)
	if got := legacy.Commanders[treasurer.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleTreasurer {
		t.Fatalf("debug/legacy save conversion council role = %#v, want treasurer", got)
	}
	if got := legacy.Commanders[diplomat.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleDiplomat {
		t.Fatalf("debug/legacy save conversion council role = %#v, want diplomat", got)
	}
	if got := legacy.Commanders[headOfHouse.ID]; got == nil || got.AssignedCouncilRole != state.CouncilRoleHeadOfHouse {
		t.Fatalf("debug/legacy save conversion council role = %#v, want head of house", got)
	}
}
