package render

import (
	"testing"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestCloneEditMapBuildStateIsolatedFromSource(t *testing.T) {
	region := &world.Region{
		ID:          "region_a",
		Neighbors:   []world.RegionID{"region_b"},
		Settlements: []world.Settlement{{ID: "settlement_a"}},
		Shape:       [][][2]float32{{{1, 2}, {3, 4}}},
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{region.ID: region},
		TerrainAreas: []world.TerrainArea{{
			ID:       "terrain_a",
			Cells:    [][2]int{{1, 2}},
			Polygons: [][][2]int{{{3, 4}, {5, 6}}},
		}},
		ShapeData: world.CountryShapeJSON{
			Shapes: map[string][][][2]float32{"shape_a": {{{1, 2}}}},
		},
		RegionPaintOverrides: map[int]world.RegionID{12: "region_a"},
	}

	clone := cloneEditMapBuildState(gs)
	clone.Regions[region.ID].Neighbors[0] = "changed"
	clone.Regions[region.ID].Settlements[0].ID = "changed"
	clone.Regions[region.ID].Shape[0][0][0] = 99
	clone.TerrainAreas[0].Cells[0][0] = 99
	clone.TerrainAreas[0].Polygons[0][0][0] = 99
	clone.ShapeData.Shapes["shape_a"][0][0][0] = 99
	clone.RegionPaintOverrides[12] = "changed"

	if region.Neighbors[0] != "region_b" || region.Settlements[0].ID != "settlement_a" {
		t.Fatal("cloning the map state mutated source region data")
	}
	if region.Shape[0][0][0] != 1 || gs.TerrainAreas[0].Cells[0][0] != 1 || gs.TerrainAreas[0].Polygons[0][0][0] != 3 {
		t.Fatal("cloning the map state mutated source geometry")
	}
	if gs.ShapeData.Shapes["shape_a"][0][0][0] != 1 || gs.RegionPaintOverrides[12] != "region_a" {
		t.Fatal("cloning the map state mutated source shape/override data")
	}
}

func TestPollEditMapBuildIgnoresStaleGeneration(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"current": {ID: "current"},
		},
	}
	resultCh := make(chan editMapBuildResult, 1)
	r := &Renderer{
		gs:                     gs,
		editMapBuildPending:    true,
		editMapBuildGeneration: 2,
		editMapBuildResult:     resultCh,
	}
	resultCh <- editMapBuildResult{
		generation: 1,
		snapshot: &state.GameState{Regions: map[world.RegionID]*world.Region{
			"stale": {ID: "stale"},
		}},
		worldMap: &WorldMap{},
	}

	r.pollEditMapBuild()

	if _, ok := r.gs.Regions["current"]; !ok {
		t.Fatal("stale worker result replaced the live regions")
	}
	if _, ok := r.gs.Regions["stale"]; ok {
		t.Fatal("stale worker result was applied")
	}
	if r.editMapBuildPending || r.editMapBuildResult != nil {
		t.Fatal("worker result channel was not cleared")
	}
}

func TestCancelEditMapBuildInvalidatesWorkerAndCompletion(t *testing.T) {
	cancelled := false
	r := &Renderer{
		editMapBuildPending:    true,
		editMapBuildGeneration: 4,
		editMapBuildCancel:     func() { cancelled = true },
		editMapBuildCompletion: func() {},
	}

	r.cancelEditMapBuild()

	if !cancelled {
		t.Fatal("worker cancellation callback was not called")
	}
	if r.editMapBuildPending || r.editMapBuildCancel != nil || r.editMapBuildCompletion != nil {
		t.Fatal("cancelled worker state was not cleared")
	}
	if r.editMapBuildGeneration != 5 {
		t.Fatalf("generation = %d, want 5", r.editMapBuildGeneration)
	}
}
