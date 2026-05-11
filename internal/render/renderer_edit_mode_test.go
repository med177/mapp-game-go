package render

import (
	"testing"

	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestEditRegionAtAllowsSeaRegions(t *testing.T) {
	r := newSeaEditRenderer()
	sea := r.gs.Regions["sea_test"]
	sx, sy := r.worldToScreen(wcX(sea.WorldX), wcY(sea.WorldY))

	if got := r.editRegionAt(sx, sy); got != sea.ID {
		t.Fatalf("sea region secilemedi: got=%q want=%q", got, sea.ID)
	}
}

func TestAddRegionFromSourcePreservesSeaFlag(t *testing.T) {
	r := newSeaEditRenderer()
	r.addRegionFromSource("sea_test", 36, 38)

	if len(r.gs.Regions) != 2 {
		t.Fatalf("beklenen 2 region, got=%d", len(r.gs.Regions))
	}

	for rid, region := range r.gs.Regions {
		if rid == "sea_test" {
			continue
		}
		if !region.IsSea {
			t.Fatalf("yeni region deniz olmali: %+v", region)
		}
		if region.Terrain != world.TerrainSea {
			t.Fatalf("yeni deniz region terrain sea olmali: got=%q", region.Terrain)
		}
		return
	}

	t.Fatal("yeni region bulunamadi")
}

func TestMoveSelectedRegionCenterToAllowsSea(t *testing.T) {
	r := newSeaEditRenderer()
	r.editSelectedRegion = "sea_test"
	sx, sy := r.worldToScreen(wcX(40), wcY(42))

	r.moveSelectedRegionCenterTo(sx, sy)

	sea := r.gs.Regions["sea_test"]
	if sea.WorldX != 40 || sea.WorldY != 42 {
		t.Fatalf("sea center tasinmadi: got=(%d,%d)", sea.WorldX, sea.WorldY)
	}
}

func newSeaEditRenderer() *Renderer {
	worldW := 64
	worldH := 64
	offset := 0.0
	scale := 1.0
	gs := &state.GameState{
		MapConfig: scenario.MapConfig{
			WorldWidth:   &worldW,
			WorldHeight:  &worldH,
			ShapeOffsetX: &offset,
			ShapeOffsetY: &offset,
			ShapeScaleX:  &scale,
			ShapeScaleY:  &scale,
		},
		Regions: map[world.RegionID]*world.Region{
			"sea_test": {
				ID:      "sea_test",
				Name:    "Sea Test",
				NameTR:  "Deniz Test",
				Terrain: world.TerrainSea,
				WorldX:  20,
				WorldY:  20,
				ShapeID: "sea_shape",
				IsSea:   true,
			},
		},
		RegionOrder: []world.RegionID{"sea_test"},
	}
	return New(gs)
}