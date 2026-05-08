package victory

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestConquerCityVictoryTriggersWhenTargetOwned(t *testing.T) {
	gs := &state.GameState{
		Turn:            1,
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: faction.FactionID("ottoman"),
		Victory: state.VictoryCondition{
			Type:            state.VictoryConquerCity,
			RequiredRegions: []world.RegionID{"constantinople"},
		},
		Regions: map[world.RegionID]*world.Region{
			"bithynia": {
				ID:      "bithynia",
				OwnerID: "ottoman",
			},
			"constantinople": {
				ID:      "constantinople",
				OwnerID: "ottoman",
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman"},
		},
	}

	Check(gs)

	if gs.Phase != state.PhaseGameOver {
		t.Fatalf("expected game over, got %s", gs.Phase)
	}
	if gs.WinnerID != gs.PlayerFactionID {
		t.Fatalf("expected winner %s, got %s", gs.PlayerFactionID, gs.WinnerID)
	}
}

func TestConquerCityVictoryWaitsForTargetOwnership(t *testing.T) {
	gs := &state.GameState{
		Turn:            1,
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: faction.FactionID("ottoman"),
		Victory: state.VictoryCondition{
			Type:            state.VictoryConquerCity,
			RequiredRegions: []world.RegionID{"constantinople"},
		},
		Regions: map[world.RegionID]*world.Region{
			"bithynia": {
				ID:      "bithynia",
				OwnerID: "ottoman",
			},
			"constantinople": {
				ID:      "constantinople",
				OwnerID: "byzantine",
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman"},
		},
	}

	Check(gs)

	if gs.Phase == state.PhaseGameOver {
		t.Fatal("expected game to continue before target ownership")
	}
}
