package render

import (
	"os"
	"path/filepath"
	"testing"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestLoadRegionPaintOverridesTreatsEmptyJSONAsNoOverrides(t *testing.T) {
	for _, content := range []string{"null", "[]"} {
		path := filepath.Join(t.TempDir(), "region_shapes.json")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if overrides := loadRegionPaintOverrides(path); len(overrides) != 0 {
			t.Fatalf("content %q: expected no overrides, got %#v", content, overrides)
		}
	}
}

func TestEnsureEditRegionPaintOverridesInitializesNilMap(t *testing.T) {
	r := &Renderer{}
	r.ensureEditRegionPaintOverrides()
	if r.editRegionPaintOverrides == nil {
		t.Fatal("expected edit region paint overrides map to be initialized")
	}
}

func TestRegionCenterAffectsRasterOnlyForSeaOrSharedShape(t *testing.T) {
	r := &Renderer{gs: &state.GameState{Regions: map[world.RegionID]*world.Region{
		"single":   {ID: "single", ShapeID: "single_shape"},
		"shared_a": {ID: "shared_a", ShapeID: "shared_shape"},
		"shared_b": {ID: "shared_b", ShapeID: "shared_shape"},
		"sea":      {ID: "sea", IsSea: true},
	}}}

	if r.regionCenterAffectsRaster(r.gs.Regions["single"]) {
		t.Fatal("a unique land shape center should not require raster rebuild")
	}
	if !r.regionCenterAffectsRaster(r.gs.Regions["shared_a"]) {
		t.Fatal("a shared-shape center should require raster rebuild")
	}
	if !r.regionCenterAffectsRaster(r.gs.Regions["sea"]) {
		t.Fatal("a sea center should require raster rebuild")
	}
}

func TestSetTerrainAreaTypeValueUpdatesOnlySelectedArea(t *testing.T) {
	first := &world.Region{ID: "first_area", IsTerrainArea: true, TerrainAreaID: "area_a"}
	second := &world.Region{ID: "second_area", IsTerrainArea: true, TerrainAreaID: "area_b"}
	r := &Renderer{gs: &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			first.ID:  first,
			second.ID: second,
		},
		TerrainAreas: []world.TerrainArea{
			{ID: "area_a", Terrain: world.TerrainPlain},
			{ID: "area_b", Terrain: world.TerrainForest},
		},
	}}

	r.setTerrainAreaTypeValue("area_a", world.TerrainMountain)
	if got := r.gs.TerrainAreas[0].Terrain; got != world.TerrainMountain {
		t.Fatalf("selected area terrain = %q, want %q", got, world.TerrainMountain)
	}
	if got := first.Terrain; got != world.TerrainMountain {
		t.Fatalf("selected child terrain = %q, want %q", got, world.TerrainMountain)
	}
	if got := r.gs.TerrainAreas[1].Terrain; got != world.TerrainForest {
		t.Fatalf("other area terrain changed to %q", got)
	}
	if got := second.Terrain; got != "" {
		t.Fatalf("other child terrain changed to %q", got)
	}
}

func TestCycleTerrainAreaCostSkipsMapRebuildForNonZeroTransitions(t *testing.T) {
	region := &world.Region{ID: "area_region", IsTerrainArea: true, TerrainAreaID: "area_a"}
	r := &Renderer{
		gs: &state.GameState{
			Regions:      map[world.RegionID]*world.Region{region.ID: region},
			TerrainAreas: []world.TerrainArea{{ID: "area_a", MoveCost: -1}},
		},
		editSelectedRegion:      region.ID,
		editTerrainAreaMoveCost: -1,
	}
	r.cycleEditTerrainAreaCost()
	if got := r.gs.TerrainAreas[0].MoveCost; got != -2 {
		t.Fatalf("move cost = %d, want -2", got)
	}

	r.cycleEditTerrainAreaCost()
	if got := r.gs.TerrainAreas[0].MoveCost; got != 0 {
		t.Fatalf("move cost = %d, want 0", got)
	}
}

func TestNextTerrainAreaMoveCostStartsWithDraftCost(t *testing.T) {
	if got := nextTerrainAreaMoveCost(0); got != -1 {
		t.Fatalf("cost 0 should cycle to -1, got %d", got)
	}
	if got := nextTerrainAreaMoveCost(-1); got != -2 {
		t.Fatalf("cost -1 should cycle to -2, got %d", got)
	}
	if got := nextTerrainAreaMoveCost(-2); got != 0 {
		t.Fatalf("cost -2 should cycle to 0, got %d", got)
	}
}

func TestScenarioTerrainAreasSplitPassablePolygonsByBaseRegion(t *testing.T) {
	root := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise")
	regions, order, err := world.LoadRegionsWithOrder(filepath.Join(root, "data", "regions.json"))
	if err != nil {
		t.Fatal(err)
	}
	areas, err := world.LoadTerrainAreas(filepath.Join(root, "data", "terrain_areas.json"), regions)
	if err != nil {
		t.Fatal(err)
	}
	shapes, err := world.LoadCountryShapes(filepath.Join(root, "data", "country_shapes.json"), regions)
	if err != nil {
		t.Fatal(err)
	}
	gs := &state.GameState{Regions: regions, RegionOrder: order, TerrainAreas: areas, ShapeData: shapes}
	prepareWorldMapData(gs, "", MapModeNormal, nil, false, true)

	fragmentsByArea := map[string]int{}
	for id, region := range gs.Regions {
		if region != nil && region.IsTerrainArea {
			fragmentsByArea[region.TerrainAreaID]++
			areaIndex := -1
			for i := range areas {
				if areas[i].ID == region.TerrainAreaID {
					areaIndex = i
					break
				}
			}
			if areaIndex >= 0 && id == world.TerrainAreaRegionID(region.TerrainAreaID) && world.TerrainAreaIsPassable(areas[areaIndex]) {
				t.Fatalf("passable area %s retained its unsplit root runtime region", region.TerrainAreaID)
			}
		}
	}
	for _, areaID := range []string{"area__6", "area__11"} {
		if fragmentsByArea[areaID] < 2 {
			t.Fatalf("passable area %s should split by base region, got %d fragment(s)", areaID, fragmentsByArea[areaID])
		}
	}
	t.Logf("runtime terrain fragments: %#v", fragmentsByArea)
}
