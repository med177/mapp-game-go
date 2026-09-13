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
