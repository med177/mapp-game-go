package world

import "testing"

func TestSyncTerrainAreaRegionsBuildsChildNeighbors(t *testing.T) {
	regions := map[RegionID]*Region{
		"parent":  {ID: "parent", Terrain: TerrainPlain, Neighbors: []RegionID{"outside"}},
		"outside": {ID: "outside", IsSea: true},
	}
	areas := []TerrainArea{
		{ID: "north", ParentRegionID: "parent", Terrain: TerrainDesert, Cells: [][2]int{{10, 10}}},
		{ID: "south", ParentRegionID: "parent", Cells: [][2]int{{10, 11}}},
	}
	SyncTerrainAreaRegions(regions, areas)
	north := regions["terrain_area::north"]
	south := regions["terrain_area::south"]
	if north == nil || south == nil || !north.IsTerrainArea || north.ParentRegionID != "parent" {
		t.Fatalf("child regions were not materialized: %#v %#v", north, south)
	}
	if north.Terrain != TerrainDesert || north.WorldX != 10 || north.WorldY != 10 {
		t.Fatalf("child terrain/center = %q (%d,%d)", north.Terrain, north.WorldX, north.WorldY)
	}
	if !hasRegionID(north.Neighbors, "parent") || !hasRegionID(north.Neighbors, "terrain_area::south") {
		t.Fatalf("north neighbors = %#v", north.Neighbors)
	}
	if !hasRegionID(regions["parent"].Neighbors, north.ID) || !hasRegionID(regions["parent"].Neighbors, south.ID) {
		t.Fatalf("parent neighbors = %#v", regions["parent"].Neighbors)
	}
}

func hasRegionID(ids []RegionID, wanted RegionID) bool {
	for _, id := range ids {
		if id == wanted {
			return true
		}
	}
	return false
}
