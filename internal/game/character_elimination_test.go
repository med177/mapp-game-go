package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestFactionEliminationReleasesLivingHouseholdCharacters(t *testing.T) {
	character := &army.Commander{
		ID: "chr_survivor", Type: army.CharacterTypeCharacter, OwnerID: "fallen",
		AssignedCouncilRole: state.CouncilRoleTreasurer,
	}
	gs := &state.GameState{
		Commanders: map[string]*army.Commander{character.ID: character},
		Factions: map[faction.FactionID]*faction.Faction{
			"fallen":    {ID: "fallen"},
			"successor": {ID: "successor"},
		},
		Armies: make(map[army.ArmyID]*army.Army),
	}

	eliminateFaction(gs, "fallen", "successor")

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
