package render

import (
	"testing"

	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestShapePaintStrokeUpdatesShapeDataAndWorldMap(t *testing.T) {
	r := newLandShapeEditRenderer()
	r.editInspectorTab = editInspectorShape
	r.editSelectedRegion = "land_test"
	r.editShapeBrushRadius = 1

	if got := r.worldMap.RegionAt(14, 10); got != "" {
		t.Fatalf("paint oncesi piksel bos olmali, got=%q", got)
	}
	sx, sy := r.worldToScreen(wcX(14), wcY(10))
	if !r.beginShapePaintStroke(sx, sy) {
		t.Fatal("shape stroke baslatilamadi")
	}
	r.finishShapePaintStroke()

	if got := r.worldMap.RegionAt(14, 10); got != "land_test" {
		t.Fatalf("paint sonrasi piksel region'a baglanmadi: got=%q", got)
	}
	if len(r.gs.ShapeData.Shapes["land_shape"]) == 0 {
		t.Fatal("shape ringleri guncellenmedi")
	}
	if !r.editDirty {
		t.Fatal("shape edit dirty flag set etmedi")
	}
}

func TestShapePaintStrokeTracksLivePreviewDiff(t *testing.T) {
	r := newLandShapeEditRenderer()
	r.editInspectorTab = editInspectorShape
	r.editSelectedRegion = "land_test"
	r.editShapeBrushRadius = 1

	sx, sy := r.worldToScreen(wcX(14), wcY(10))
	if !r.beginShapePaintStroke(sx, sy) {
		t.Fatal("shape stroke baslatilamadi")
	}
	session := r.editShapeSession
	if session == nil {
		t.Fatal("shape session olusmadi")
	}
	idx := session.index(14, 10)
	if len(session.DiffList) == 0 {
		t.Fatal("canli preview diff kaydi olusmadi")
	}
	if session.DiffMask[idx] == 0 {
		t.Fatal("boyanan piksel diff mask'e islenmedi")
	}
	if session.Mask[idx] == 0 {
		t.Fatal("boyanan piksel mask'e islenmedi")
	}
}

func TestWorldSnapshotClonesShapeData(t *testing.T) {
	r := newLandShapeEditRenderer()
	snap := r.worldSnapshot()
	r.gs.ShapeData.Shapes["land_shape"][0][0][0] = 999

	if got := snap.ShapeData.Shapes["land_shape"][0][0][0]; got == 999 {
		t.Fatal("shape data snapshot clone edilmedi")
	}
}

func newLandShapeEditRenderer() *Renderer {
	worldW := 64
	worldH := 64
	offset := 0.0
	scale := 1.0
	rings := [][][2]float32{{
		{8, 8},
		{12, 8},
		{12, 12},
		{8, 12},
	}}
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
			"land_test": {
				ID:      "land_test",
				Name:    "Land Test",
				NameTR:  "Kara Test",
				Terrain: world.TerrainPlain,
				WorldX:  10,
				WorldY:  10,
				ShapeID: "land_shape",
				Shape:   cloneFloatRings(rings),
			},
		},
		RegionOrder: []world.RegionID{"land_test"},
		ShapeData: world.CountryShapeJSON{
			Shapes: map[string][][][2]float32{"land_shape": cloneFloatRings(rings)},
			Names:  map[string]string{"land_shape": "Land Shape"},
		},
	}
	return New(gs)
}
