package render

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestCanPlaceEditLandArmyRejectsBlockedTerrain(t *testing.T) {
	blocked := &world.Region{ID: "blocked", IsTerrainArea: true, TerrainAreaID: "area"}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{blocked.ID: blocked},
		TerrainAreas: []world.TerrainArea{{
			ID:       "area",
			MoveCost: 0,
			Polygons: [][][2]int{{{-1, -1}, {2, -1}, {2, 2}, {-1, 2}}},
		}},
	}
	if canPlaceEditLandArmy(gs, blocked) {
		t.Fatal("edit mode allowed an army on blocked terrain")
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

func TestRebuildShapeRegionAssignmentsOnlyTouchesTargetShape(t *testing.T) {
	originalWorldW, originalWorldH := WorldW, WorldH
	originalShapeOffX, originalShapeOffY := shapeOffX, shapeOffY
	originalShapeScaleX, originalShapeScaleY := shapeScaleX, shapeScaleY
	t.Cleanup(func() {
		WorldW, WorldH = originalWorldW, originalWorldH
		shapeOffX, shapeOffY = originalShapeOffX, originalShapeOffY
		shapeScaleX, shapeScaleY = originalShapeScaleX, originalShapeScaleY
	})
	WorldW, WorldH = 8, 2
	shapeOffX, shapeOffY = 0, 0
	shapeScaleX, shapeScaleY = 1, 1

	regions := map[world.RegionID]*world.Region{
		"shared_a": {ID: "shared_a", ShapeID: "shared", WorldX: 0, WorldY: 0, Terrain: world.TerrainPlain},
		"shared_b": {ID: "shared_b", ShapeID: "shared", WorldX: 4, WorldY: 0, Terrain: world.TerrainPlain},
		"other":    {ID: "other", ShapeID: "other_shape", WorldX: 7, WorldY: 0, Terrain: world.TerrainPlain},
	}
	gs := &state.GameState{Regions: regions}
	wm := &WorldMap{
		basePixels:        make([]byte, WorldW*WorldH*4),
		baseRegionAt:      []uint16{1, 1, 1, 1, 3, 3, 3, 3, 0, 0, 0, 0, 0, 0, 0, 0},
		regionAt:          []uint16{1, 1, 1, 1, 3, 3, 3, 3, 0, 0, 0, 0, 0, 0, 0, 0},
		regionIDs:         []world.RegionID{"", "shared_a", "shared_b", "other"},
		regionIdx:         map[world.RegionID]uint16{"shared_a": 1, "shared_b": 2, "other": 3},
		regionPx:          map[world.RegionID][]int{"shared_a": {0, 1, 2, 3}, "other": {4, 5, 6, 7}},
		shapeRasterPixels: make(map[string][]int),
	}
	wm.rebuildShapeRasterCache(gs)

	if !wm.rebuildShapeRegionAssignments(gs, "shared") {
		t.Fatal("target shape was not rebuilt")
	}
	wantTarget := []uint16{1, 1, 2, 2}
	for i, want := range wantTarget {
		if got := wm.baseRegionAt[i]; got != want {
			t.Fatalf("target pixel %d = region index %d, want %d", i, got, want)
		}
	}
	for i := 4; i < 8; i++ {
		if got := wm.baseRegionAt[i]; got != 3 {
			t.Fatalf("unrelated pixel %d changed to region index %d", i, got)
		}
	}
}

func TestWorldMapRenameRegionIDUpdatesOnlyIDCaches(t *testing.T) {
	oldID := world.RegionID("old_region")
	newID := world.RegionID("new_region")
	wm := &WorldMap{
		regionAt:          []uint16{1, 1, 0},
		regionIDs:         []world.RegionID{"", oldID},
		regionIdx:         map[world.RegionID]uint16{oldID: 1},
		regionPx:          map[world.RegionID][]int{oldID: {0, 1}},
		regionAnchor:      map[world.RegionID][2]int{oldID: {10, 20}},
		settlementAnchor:  map[settlementAnchorKey][2]int{{Region: oldID, Index: 0}: {11, 21}},
		primarySettlement: map[world.RegionID][2]int{oldID: {11, 21}},
	}

	wm.renameRegionID(oldID, newID)

	if got := wm.regionAt; got[0] != 1 || got[1] != 1 {
		t.Fatalf("region raster changed during ID rename: %v", got)
	}
	if wm.regionIDs[1] != newID || wm.regionIdx[newID] != 1 {
		t.Fatalf("region index cache was not moved: IDs=%v index=%v", wm.regionIDs, wm.regionIdx)
	}
	if _, ok := wm.regionIdx[oldID]; ok {
		t.Fatal("old region ID remained in the index cache")
	}
	if got := wm.regionPx[newID]; len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("region pixel cache was not moved: %v", got)
	}
	if x, y, ok := wm.RegionAnchor(newID); !ok || x != 10 || y != 20 {
		t.Fatalf("region anchor was not moved: %d, %d, %v", x, y, ok)
	}
	if x, y, ok := wm.SettlementAnchor(newID, 0); !ok || x != 11 || y != 21 {
		t.Fatalf("settlement anchor was not moved: %d, %d, %v", x, y, ok)
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
	wm := prepareWorldMapData(gs, "", MapModeNormal, nil, false, true)

	var targetShapeID string
	var targetRegion *world.Region
	for shapeID := range wm.shapeRasterPixels {
		count := 0
		for _, region := range gs.Regions {
			if region != nil && !region.IsTerrainArea && region.ShapeID == shapeID {
				count++
				if targetRegion == nil {
					targetRegion = region
				}
			}
		}
		if count > 1 {
			targetShapeID = shapeID
			break
		}
	}
	if targetShapeID == "" || targetRegion == nil {
		t.Fatal("scenario has no shared shape raster cache entry")
	}
	snapshot := cloneEditMapBuildState(gs)
	snapshot.Regions[targetRegion.ID].WorldX++
	partialStarted := time.Now()
	partialBaseMap := cloneWorldMapForEdit(wm)
	partialMap, partialTiming, err := buildEditMapSnapshot(nil, snapshot, nil, partialBaseMap, targetShapeID)
	if err != nil || partialMap == nil {
		t.Fatalf("partial shape rebuild failed: map=%v err=%v", partialMap != nil, err)
	}
	t.Logf("partial shape rebuild: %s (raster=%s, post=%s, shape=%s, pixels=%d)",
		time.Since(partialStarted), partialTiming.raster, partialTiming.post, targetShapeID, len(wm.shapeRasterPixels[targetShapeID]))

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
