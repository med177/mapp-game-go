package diplomacy

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestIndependentWarSatisfactionPenaltyUsesBorderType(t *testing.T) {
	tests := []struct {
		name        string
		regions     map[world.RegionID]*world.Region
		wantPenalty int
	}{
		{
			name: "uzak savaş",
			regions: map[world.RegionID]*world.Region{
				"a": {ID: "a", OwnerID: "a"},
				"b": {ID: "b", OwnerID: "b"},
			},
			wantPenalty: 1,
		},
		{
			name: "ortak deniz sınırı",
			regions: map[world.RegionID]*world.Region{
				"a":   {ID: "a", OwnerID: "a", Neighbors: []world.RegionID{"sea"}},
				"b":   {ID: "b", OwnerID: "b", Neighbors: []world.RegionID{"sea"}},
				"sea": {ID: "sea", IsSea: true},
			},
			wantPenalty: 2,
		},
		{
			name: "kara sınırı öncelikli",
			regions: map[world.RegionID]*world.Region{
				"a":   {ID: "a", OwnerID: "a", Neighbors: []world.RegionID{"b", "sea"}},
				"b":   {ID: "b", OwnerID: "b", Neighbors: []world.RegionID{"a", "sea"}},
				"sea": {ID: "sea", IsSea: true},
			},
			wantPenalty: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gs := &state.GameState{
				Regions: test.regions,
				Relations: map[string]*faction.Relation{
					faction.RelationKey("a", "b"): {
						FactionA: "a",
						FactionB: "b",
						Stance:   faction.StanceWar,
					},
				},
			}
			if got := IndependentWarSatisfactionPenalty(gs, "a"); got != test.wantPenalty {
				t.Fatalf("ceza = %d, beklenen %d", got, test.wantPenalty)
			}
		})
	}
}
