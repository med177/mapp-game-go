package world

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTerrainAreaMovementCost(t *testing.T) {
	areas := []TerrainArea{{ID: "mountain", ParentRegionID: "parent", MoveCost: -1, Cells: [][2]int{{4, 5}}}, {ID: "cliff", ParentRegionID: "parent", MoveCost: 0, Cells: [][2]int{{6, 7}}}}
	if cost, blocked := TerrainAreaMovementCost(areas, "parent", 4, 5); blocked || cost != -1 {
		t.Fatalf("penalty area = (%d, %v), want (-1, false)", cost, blocked)
	}
	if cost, blocked := TerrainAreaMovementCost(areas, "parent", 6, 7); !blocked || cost != 0 {
		t.Fatalf("blocked area = (%d, %v), want (0, true)", cost, blocked)
	}
}

func TestLoadTerrainAreasFiltersInvalidParents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "terrain_areas.json")
	data, _ := json.Marshal([]TerrainArea{{ID: "valid", ParentRegionID: "land", MoveCost: -2, Cells: [][2]int{{1, 1}}}, {ID: "sea-child", ParentRegionID: "sea", Cells: [][2]int{{2, 2}}}, {ID: "missing", ParentRegionID: "none", Cells: [][2]int{{3, 3}}}})
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	areas, err := LoadTerrainAreas(path, map[RegionID]*Region{"land": {}, "sea": {IsSea: true}})
	if err != nil || len(areas) != 1 || areas[0].ID != "valid" {
		t.Fatalf("loaded areas = %#v, err=%v", areas, err)
	}
}
