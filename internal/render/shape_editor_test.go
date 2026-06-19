package render

import (
	"os"
	"path/filepath"
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

func TestRegionPaintStrokeOverridesWorldMapRegionAt(t *testing.T) {
	r := newLandShapeEditRenderer()
	r.editInspectorTab = editInspectorShape
	r.editSelectedRegion = "land_test"
	r.editShapeTool = editShapeToolRegion
	r.editShapeBrushMode = editShapeBrushPaint
	r.editShapeBrushRadius = 1

	if got := r.worldMap.RegionAt(14, 10); got != "" {
		t.Fatalf("paint oncesi piksel bos olmali, got=%q", got)
	}
	sx, sy := r.worldToScreen(wcX(14), wcY(10))
	if !r.beginShapePaintStroke(sx, sy) {
		t.Fatal("region paint stroke baslatilamadi")
	}
	if r.editShapeSession != nil {
		t.Fatal("region paint stroke shape session olusturmamali")
	}
	if state := r.regionPaintPreviewStateAt(10*WorldW + 14); state != 1 {
		t.Fatalf("region paint canli preview yesil olmali, got=%d", state)
	}
	r.finishShapePaintStroke()

	if got := r.worldMap.RegionAt(14, 10); got != "land_test" {
		t.Fatalf("paint sonrasi piksel region'a baglanmadi: got=%q", got)
	}
	if len(r.editRegionPaintOverrides) == 0 {
		t.Fatal("region paint overrides kayit edilmedi")
	}
	if len(r.editRegionPaintStrokeList) != 0 {
		t.Fatal("stroke bitince region paint preview temizlenmeli")
	}
}

func TestRegionPaintStrokeAllowsSeaRegions(t *testing.T) {
	r := newSeaRegionPaintRenderer()
	r.editInspectorTab = editInspectorShape
	r.editSelectedRegion = "sea_left"
	r.editShapeTool = editShapeToolRegion
	r.editShapeBrushMode = editShapeBrushPaint
	r.editShapeBrushRadius = 1

	if got := r.worldMap.RegionAt(40, 20); got != "sea_right" {
		t.Fatalf("paint oncesi hedef piksel sea_right olmali, got=%q", got)
	}
	sx, sy := r.worldToScreen(wcX(40), wcY(20))
	if !r.beginShapePaintStroke(sx, sy) {
		t.Fatal("sea region paint stroke baslatilamadi")
	}
	beforeWM := r.worldMap
	r.finishShapePaintStroke()

	if got := r.worldMap.RegionAt(40, 20); got != "sea_left" {
		t.Fatalf("paint sonrasi piksel sea_left olmali, got=%q", got)
	}
	if len(r.editRegionPaintOverrides) == 0 {
		t.Fatal("sea region paint overrides kayit edilmedi")
	}
	if r.worldMap != beforeWM {
		t.Fatal("sea region paint full world map rebuild yapmamali")
	}
}

func TestRegionPaintStrokeKeepsExistingExternalOverridesAcrossMultipleStrokes(t *testing.T) {
	r := newLandShapeEditRenderer()
	r.editInspectorTab = editInspectorShape
	r.editSelectedRegion = "land_test"
	r.editShapeTool = editShapeToolRegion
	r.editShapeBrushMode = editShapeBrushPaint
	r.editShapeBrushRadius = 1

	firstX, firstY := 14, 10
	sx, sy := r.worldToScreen(wcX(firstX), wcY(firstY))
	if !r.beginShapePaintStroke(sx, sy) {
		t.Fatal("ilk region paint stroke baslatilamadi")
	}
	r.finishShapePaintStroke()

	firstIdx := firstY*WorldW + firstX
	if got := r.editRegionPaintOverrides[firstIdx]; got != "land_test" {
		t.Fatalf("ilk dis override korunmaliydi, got=%q", got)
	}

	secondX := 15
	sx2, sy2 := r.worldToScreen(wcX(secondX), wcY(firstY))
	if !r.beginShapePaintStroke(sx, sy) {
		t.Fatal("ikinci region paint stroke baslatilamadi")
	}
	r.continueShapePaintStroke(sx2, sy2)
	r.finishShapePaintStroke()

	if got := r.worldMap.RegionAt(firstX, firstY); got != "land_test" {
		t.Fatalf("ikinci stroke sonrasi ilk dis piksel secili region'da kalmali, got=%q", got)
	}
	if got := r.worldMap.RegionAt(secondX, firstY); got != "land_test" {
		t.Fatalf("ikinci stroke sonrasi yeni piksel secili region'a baglanmali, got=%q", got)
	}
}

func TestRegionPaintStrokeEraseClearsGameStateOverrides(t *testing.T) {
	r := newLandShapeEditRenderer()
	r.editInspectorTab = editInspectorShape
	r.editSelectedRegion = "land_test"
	r.editShapeTool = editShapeToolRegion
	r.editShapeBrushMode = editShapeBrushPaint
	r.editShapeBrushRadius = 1

	sx, sy := r.worldToScreen(wcX(14), wcY(10))
	if !r.beginShapePaintStroke(sx, sy) {
		t.Fatal("ilk region paint stroke baslatilamadi")
	}
	r.finishShapePaintStroke()

	r.editShapeBrushMode = editShapeBrushErase
	if !r.beginShapePaintStroke(sx, sy) {
		t.Fatal("erase stroke baslatilamadi")
	}
	r.finishShapePaintStroke()

	if len(r.gs.RegionPaintOverrides) != 0 {
		t.Fatalf("erase sonrasi game state overrides temizlenmeli, got=%d", len(r.gs.RegionPaintOverrides))
	}
}

func TestRegionPaintOverridesPersistOutsideBaseShapeAfterReload(t *testing.T) {
	r := newLandShapeEditRenderer()
	r.editInspectorTab = editInspectorShape
	r.editSelectedRegion = "land_test"
	r.editShapeTool = editShapeToolRegion
	r.editShapeBrushMode = editShapeBrushPaint
	r.editShapeBrushRadius = 1

	targetX, targetY := 14, 10
	sx, sy := r.worldToScreen(wcX(targetX), wcY(targetY))
	if !r.beginShapePaintStroke(sx, sy) {
		t.Fatal("region paint stroke baslatilamadi")
	}
	r.finishShapePaintStroke()

	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("data dir olusturulamadi: %v", err)
	}
	path := filepath.Join(dataDir, "region_shapes.json")
	if err := SaveRegionPaintOverrides(path, r.gs.RegionPaintOverrides); err != nil {
		t.Fatalf("region paint overrides kaydedilemedi: %v", err)
	}

	gs := newLandShapeEditRenderer().gs
	gs.ScenarioPath = tempDir
	gs.RegionPaintOverrides = nil
	wm := NewWorldMap(gs)
	if got := wm.RegionAt(targetX, targetY); got != "land_test" {
		t.Fatalf("reload sonrasi dis sinir override'i korunmali, got=%q", got)
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

func newSeaRegionPaintRenderer() *Renderer {
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
			"sea_left": {
				ID:      "sea_left",
				Name:    "Sea Left",
				NameTR:  "Sol Deniz",
				Terrain: world.TerrainSea,
				WorldX:  12,
				WorldY:  20,
				IsSea:   true,
			},
			"sea_right": {
				ID:      "sea_right",
				Name:    "Sea Right",
				NameTR:  "Sag Deniz",
				Terrain: world.TerrainSea,
				WorldX:  52,
				WorldY:  20,
				IsSea:   true,
			},
		},
		RegionOrder: []world.RegionID{"sea_left", "sea_right"},
	}
	return New(gs)
}
