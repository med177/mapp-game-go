package render

import (
	"testing"

	"mapp-game-go/internal/world"
)

func TestCoastalSettlementPointUsesActualLandSeaRasterBoundary(t *testing.T) {
	land := &world.Region{
		ID:        "land",
		WorldX:    800,
		WorldY:    350,
		Neighbors: []world.RegionID{"sea"},
	}
	sea := &world.Region{ID: "sea", IsSea: true, WorldX: 900, WorldY: 350}
	landPixelX, landPixelY := 1191, 529
	seaPixelX := landPixelX + 1

	wm := &WorldMap{
		regionAt:  make([]uint16, WorldW*WorldH),
		regionIDs: []world.RegionID{"", "land", "sea"},
		regionPx:  map[world.RegionID][]int{"land": {landPixelY*WorldW + landPixelX}},
	}
	wm.regionAt[landPixelY*WorldW+landPixelX] = 1
	wm.regionAt[landPixelY*WorldW+seaPixelX] = 2

	x, y, ok := wm.CoastalSettlementPoint(land, map[world.RegionID]*world.Region{
		land.ID: land,
		sea.ID:  sea,
	})
	if !ok {
		t.Fatal("kara-deniz raster sınırından liman noktası üretilmeliydi")
	}
	if x < 845 || x > 850 || y < 348 || y > 352 {
		t.Fatalf("liman gerçek kıyı sınırının kara tarafında olmalıydı: got=(%d,%d)", x, y)
	}
}
