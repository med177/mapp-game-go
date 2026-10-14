package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
)

func TestAIAdoptsOneNeutralCharacterToFillCouncilVacancy(t *testing.T) {
	const aiFaction faction.FactionID = "ai"
	first := &army.Commander{
		ID: "chr_a", Type: army.CharacterTypeCharacter, OwnerID: state.NeutralCharacterOwnerID, Name: "A",
	}
	second := &army.Commander{
		ID: "chr_b", Type: army.CharacterTypeAdministrator, OwnerID: state.NeutralCharacterOwnerID, Name: "B",
	}
	gs := &state.GameState{
		Year:            300,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":  {ID: "player"},
			aiFaction: {ID: aiFaction},
		},
		Commanders: map[string]*army.Commander{
			first.ID:  first,
			second.ID: second,
		},
		CouncilConfig: scenario.DefaultCharacterCouncilConfig(),
	}

	aiManageNeutralCharacters(gs, aiFaction)

	if first.OwnerID != string(aiFaction) || first.AssignedCouncilRole != state.CouncilRoleHeadOfHouse {
		t.Fatalf("AI-adopted character = owner %q, role %q; want AI faction and head of house", first.OwnerID, first.AssignedCouncilRole)
	}
	if second.OwnerID != state.NeutralCharacterOwnerID || second.AssignedCouncilRole != "" {
		t.Fatalf("second candidate was also adopted: owner %q, role %q", second.OwnerID, second.AssignedCouncilRole)
	}
	if roster := gs.CharacterRoster(string(aiFaction)); len(roster) != 1 || roster[0] != first {
		t.Fatalf("AI character roster = %#v, want only adopted character", roster)
	}
}

func TestAILeavesNeutralCharactersForPlayerControlledFaction(t *testing.T) {
	character := &army.Commander{
		ID: "chr_a", Type: army.CharacterTypeCharacter, OwnerID: state.NeutralCharacterOwnerID, Name: "A",
	}
	gs := &state.GameState{
		Year:            300,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Commanders:    map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.DefaultCharacterCouncilConfig(),
	}

	aiManageNeutralCharacters(gs, faction.FactionID("player"))

	if character.OwnerID != state.NeutralCharacterOwnerID || character.AssignedCouncilRole != "" {
		t.Fatalf("AI changed player candidate: owner %q, role %q", character.OwnerID, character.AssignedCouncilRole)
	}
}

func TestAIFactionEliminationReleasesLivingHouseholdCharacters(t *testing.T) {
	character := &army.Commander{
		ID: "chr_survivor", Type: army.CharacterTypeCharacter, OwnerID: "fallen",
		AssignedCouncilRole: state.CouncilRoleTreasurer,
	}
	gs := &state.GameState{
		Commanders: map[string]*army.Commander{character.ID: character},
		Factions: map[faction.FactionID]*faction.Faction{
			"fallen": {ID: "fallen"},
		},
		Armies:      make(map[army.ArmyID]*army.Army),
		TradeRoutes: nil,
	}

	aiEliminateFactionAfterPrivilegeRevoke(gs, "fallen")

	if !gs.Factions["fallen"].IsEliminated {
		t.Fatal("faction was not eliminated")
	}
	if character.OwnerID != state.NeutralCharacterOwnerID || character.AssignedCouncilRole != "" {
		t.Fatalf("survivor state = owner %q, role %q; want neutral and unassigned", character.OwnerID, character.AssignedCouncilRole)
	}
	if candidates := gs.NeutralCharacterCandidates(); len(candidates) != 1 || candidates[0] != character {
		t.Fatalf("neutral candidates = %#v, want the eliminated faction's survivor", candidates)
	}
}
