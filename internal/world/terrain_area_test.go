package world

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestTerrainAreaPolygonContains(t *testing.T) {
	area := TerrainArea{Polygons: [][][2]int{{
		{0, 0}, {4, 0}, {4, 4}, {0, 4},
	}}}
	if !area.Contains(1, 1) || area.Contains(5, 5) {
		t.Fatal("polygon containment is incorrect")
	}
}

func TestTerrainAreasContainPoint(t *testing.T) {
	areas := []TerrainArea{{Polygons: [][][2]int{{
		{10, 10}, {20, 10}, {20, 20}, {10, 20},
	}}}}
	if !TerrainAreasContainPoint(areas, 15, 15) {
		t.Fatal("painted terrain area point was not detected")
	}
	if TerrainAreasContainPoint(areas, 25, 15) {
		t.Fatal("point outside painted terrain area was detected")
	}
}

func TestSortedRegionIDsReturnsAlphabeticalCopy(t *testing.T) {
	input := []RegionID{"zulu", "alpha", "area::first"}
	got := SortedRegionIDs(input)
	want := []RegionID{"alpha", "area::first", "zulu"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("region IDs were not sorted: got=%#v want=%#v", got, want)
	}
	if !reflect.DeepEqual(input, []RegionID{"zulu", "alpha", "area::first"}) {
		t.Fatalf("sorting changed the input slice: got=%#v", input)
	}
}

func TestLoadTerrainAreasTreatsEmptyJSONAsNoAreas(t *testing.T) {
	for _, content := range []string{"null", "[]"} {
		path := filepath.Join(t.TempDir(), "terrain_areas.json")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		areas, err := LoadTerrainAreas(path, nil)
		if err != nil {
			t.Fatalf("content %q: %v", content, err)
		}
		if areas != nil {
			t.Fatalf("content %q: expected no areas, got %#v", content, areas)
		}
	}
}

func TestTerrainAreaCenterStaysInsidePolygon(t *testing.T) {
	area := TerrainArea{Polygons: [][][2]int{{
		{0, 0}, {6, 0}, {0, 2},
	}}}
	x, y := area.Center()
	if !PointInPolygon(float64(x)+0.5, float64(y)+0.5, area.Polygons[0]) {
		t.Fatalf("center is outside polygon: %d,%d", x, y)
	}
}

func TestMergeTerrainAreaPolygonsCombinesIntersectingRings(t *testing.T) {
	existing := [][][2]int{{{0, 0}, {4, 0}, {4, 4}, {0, 4}}}
	added := [][2]int{{2, 2}, {6, 2}, {6, 6}, {2, 6}}

	merged, intersects := MergeTerrainAreaPolygons(existing, added)
	if !intersects || len(merged) != 1 {
		t.Fatalf("expected one merged contour, got intersects=%v contours=%d", intersects, len(merged))
	}
	if !PointInPolygon(1.5, 1.5, merged[0]) || !PointInPolygon(5.5, 5.5, merged[0]) {
		t.Fatalf("merged contour does not contain both source areas: %#v", merged[0])
	}
}

func TestMergeTerrainAreaPolygonsKeepsDisjointRingsSeparate(t *testing.T) {
	existing := [][][2]int{{{0, 0}, {2, 0}, {2, 2}, {0, 2}}}
	added := [][2]int{{4, 4}, {6, 4}, {6, 6}, {4, 6}}

	merged, intersects := MergeTerrainAreaPolygons(existing, added)
	if intersects || merged != nil {
		t.Fatalf("disjoint contours should not be merged: intersects=%v merged=%#v", intersects, merged)
	}
}

func TestTerrainAreaPassabilityUsesMoveCostOnly(t *testing.T) {
	if !TerrainAreaIsPassable(TerrainArea{Terrain: TerrainDesert, MoveCost: -1}) {
		t.Fatal("desert with a non-zero move cost should be passable")
	}
	if TerrainAreaIsPassable(TerrainArea{Terrain: TerrainDesert, MoveCost: 0}) {
		t.Fatal("zero move cost should block the area")
	}
	if !TerrainAreaIsPassable(TerrainArea{Terrain: TerrainMountain, MoveCost: -1}) {
		t.Fatal("mountain label with a non-zero move cost should be passable")
	}
	if !TerrainAreaIsPassable(TerrainArea{Terrain: TerrainLake, MoveCost: -2}) {
		t.Fatal("lake label with a non-zero move cost should be passable")
	}
}

func TestTerrainAreaJSONStoresPolygonsWithoutCells(t *testing.T) {
	data, err := json.Marshal(TerrainArea{
		ID: "sahara", ParentRegionID: "luxor", Terrain: TerrainDesert,
		MoveCost: 0, Polygons: [][][2]int{{{0, 0}, {2, 0}, {2, 2}}},
		Cells: [][2]int{{0, 0}, {1, 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"polygons"`) || strings.Contains(string(data), `"cells"`) || strings.Contains(string(data), `"parent_region_id"`) {
		t.Fatalf("unexpected terrain area JSON: %s", data)
	}
}

func TestSyncTerrainAreaRegionsKeepsOnlyParentNeighbor(t *testing.T) {
	regions := map[RegionID]*Region{
		"parent":   {ID: "parent", OwnerID: "faction", Neighbors: []RegionID{"neighbor"}},
		"neighbor": {ID: "neighbor"},
	}
	areas := []TerrainArea{{
		ID: "mountain", ParentRegionID: "parent", Terrain: TerrainMountain,
		Polygons: [][][2]int{{{0, 0}, {4, 0}, {4, 4}, {0, 4}}},
	}}

	SyncTerrainAreaRegions(regions, areas)
	child := regions["area::mountain"]
	if child == nil {
		t.Fatal("terrain child was not created")
	}
	if child.OwnerID != "" {
		t.Fatalf("terrain child unexpectedly inherited owner: %q", child.OwnerID)
	}
	if !child.IsLocked {
		t.Fatal("a move_cost=0 terrain area should start locked")
	}
	if len(child.Neighbors) != 1 || child.Neighbors[0] != "parent" {
		t.Fatalf("unexpected terrain child neighbors: %#v", child.Neighbors)
	}
}

func TestSyncTerrainAreaRegionsPreservesLoadedAreaNeighborOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "regions.json")
	data, err := json.Marshal([]*Region{
		{ID: "parent", Neighbors: []RegionID{
			"neighbor", "area::second", "area::first",
		}},
		{ID: "neighbor"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	regions, _, err := LoadRegionsWithOrder(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []RegionID{"area::second", "area::first"}
	if got := regions["parent"].AreaNeighborOrder; !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded area neighbor order changed: got %#v want %#v", got, want)
	}
	areas := []TerrainArea{
		{ID: "first", ParentRegionID: "parent", Polygons: [][][2]int{{{0, 0}, {2, 0}, {2, 2}}}},
		{ID: "second", ParentRegionID: "parent", Polygons: [][][2]int{{{3, 0}, {5, 0}, {5, 2}}}},
	}

	SyncTerrainAreaRegions(regions, areas)
	wantNeighbors := []RegionID{"neighbor", "area::second", "area::first"}
	if got := regions["parent"].Neighbors; !reflect.DeepEqual(got, wantNeighbors) {
		t.Fatalf("area neighbor order changed: got %#v want %#v", got, wantNeighbors)
	}
}

func TestSyncTerrainAreaRegionsAllowsTopLevelAreaWithoutParent(t *testing.T) {
	regions := map[RegionID]*Region{
		"plain": {ID: "plain", Terrain: TerrainPlain},
	}
	areas := []TerrainArea{{
		ID: "sahara", Terrain: "", Polygons: [][][2]int{{{0, 0}, {4, 0}, {4, 4}, {0, 4}}},
	}}

	SyncTerrainAreaRegions(regions, areas)
	child := regions["area::sahara"]
	if child == nil {
		t.Fatal("top-level terrain area runtime node was not created")
	}
	if child.Terrain != TerrainPlain {
		t.Fatalf("unexpected default terrain: %q", child.Terrain)
	}
	if child.ParentRegionID != "" {
		t.Fatalf("top-level terrain area unexpectedly has parent: %q", child.ParentRegionID)
	}
}
