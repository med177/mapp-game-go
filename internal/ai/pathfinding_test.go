package ai

import (
	"testing"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIRouteNeighborIDsIncludesLandPassageEndpoints(t *testing.T) {
	from := &world.Region{ID: "from"}
	to := &world.Region{ID: "to"}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			from.ID: from,
			to.ID:   to,
		},
		LandPassages: []world.LandPassage{{From: from.ID, To: to.ID, MoveCost: 1}},
	}

	neighbors := aiRouteNeighborIDs(gs, from)
	if len(neighbors) != 1 || neighbors[0] != to.ID {
		t.Fatalf("land passage endpoint was not added to AI route neighbors: %#v", neighbors)
	}
}

func TestAILandEntryMoveCostIncludesTerrainAndPassageCost(t *testing.T) {
	from := &world.Region{ID: "from"}
	target := &world.Region{ID: "target"}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			from.ID:   from,
			target.ID: target,
		},
		LandPassages: []world.LandPassage{{From: from.ID, To: target.ID, MoveCost: 3}},
	}

	cost, allowed := aiLandEntryMoveCost(gs, from.ID, target)
	if !allowed || cost != 3 {
		t.Fatalf("unexpected passage movement cost: allowed=%v cost=%d", allowed, cost)
	}
}

func TestAITerrainMovePenaltyIncludesAttrition(t *testing.T) {
	from := &world.Region{ID: "from"}
	target := &world.Region{ID: "terrain", IsTerrainArea: true, TerrainAreaID: "area", IsLocked: false}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			from.ID:   from,
			target.ID: target,
		},
		TerrainAreas: []world.TerrainArea{{ID: "area", MoveCost: -1, AttritionCost: 10}},
	}

	penalty := aiTerrainMovePenalty(gs, from.ID, target)
	if penalty != 30 {
		t.Fatalf("unexpected terrain risk penalty: %d", penalty)
	}
}
