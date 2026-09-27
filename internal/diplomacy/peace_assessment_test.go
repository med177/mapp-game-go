package diplomacy

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestInaccessibleStalemateCanEndUnreachableClaimWar(t *testing.T) {
	gs := &state.GameState{
		Turn: 10,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", TerritorialClaims: []faction.TerritorialClaim{{RegionID: "b_land", Value: 90}}},
			"b": {ID: "b"},
		},
		Regions: map[world.RegionID]*world.Region{
			"a_land": {ID: "a_land", OwnerID: "a"},
			"b_land": {ID: "b_land", OwnerID: "b"},
		},
		Relations: map[string]*faction.Relation{
			"a|b": {FactionA: "a", FactionB: "b", Stance: faction.StanceWar},
		},
		WarLedgers: map[string]*state.WarLedger{
			"a|b": {FactionA: "a", FactionB: "b", StartedTurn: 1},
		},
	}

	assessment := AssessPeaceDesire(gs, "a", "b")
	if !assessment.Stalemate || !assessment.Inaccessible {
		t.Fatalf("unreachable war stalemate was not classified: %#v", assessment)
	}
	if !assessment.ShouldPropose() {
		t.Fatal("unreachable claim war should allow a peace proposal")
	}
}
