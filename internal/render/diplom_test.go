package render

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestSortedFactionsSkipsPlayerAndEliminated(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"a":      {ID: "a"},
			"b":      {ID: "b", IsEliminated: true},
			"c":      {ID: "c"},
		},
	}

	got := sortedFactions(gs)
	if len(got) != 2 {
		t.Fatalf("beklenen 2 aktif fraksiyon, got=%d (%v)", len(got), got)
	}
	if got[0] != "a" || got[1] != "c" {
		t.Fatalf("beklenen [a c], got=%v", got)
	}
}
