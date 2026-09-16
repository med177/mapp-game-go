package render

import (
	"testing"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestMapRegionDoubleClickUsesTerrainParentIdentity(t *testing.T) {
	parentID := world.RegionID("parent")
	fragmentID := world.RegionID("area::forest::parent")
	r := &Renderer{gs: &state.GameState{Regions: map[world.RegionID]*world.Region{
		parentID:   {ID: parentID, OwnerID: "target"},
		fragmentID: {ID: fragmentID, IsTerrainArea: true, ParentRegionID: parentID},
	}}}

	if r.mapRegionDoubleClicked(fragmentID) {
		t.Fatal("first terrain click unexpectedly counted as double click")
	}
	if !r.mapRegionDoubleClicked(parentID) {
		t.Fatal("terrain fragment and its parent were not treated as the same map click target")
	}
}
