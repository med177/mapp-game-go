package render

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestPaintCoordinatesUseContainingCellAndItsCenter(t *testing.T) {
	shapeX, shapeY := 14, 10
	shapeWX := float64(shapeX) + 0.9
	shapeWY := float64(shapeY) + 0.9
	if gotX, gotY := shapePaintCellFromWorld(shapeWX, shapeWY); gotX != shapeX || gotY != shapeY {
		t.Fatalf("shape mouse konumu bulundugu hucreye ait olmali: got=(%d,%d) want=(%d,%d)", gotX, gotY, shapeX, shapeY)
	}
	centerX, centerY := shapePaintCellCenterWorld(shapeX, shapeY)
	wantCenterX := float64(shapeX) + 0.5
	wantCenterY := float64(shapeY) + 0.5
	if math.Abs(centerX-wantCenterX) > 0.0001 || math.Abs(centerY-wantCenterY) > 0.0001 {
		t.Fatalf("shape preview hucenin merkezine oturmali: got=(%.4f,%.4f) want=(%.4f,%.4f)", centerX, centerY, wantCenterX, wantCenterY)
	}

	regionX, regionY := regionPaintCellFromWorld(14.9, 10.9)
	if regionX != 14 || regionY != 10 {
		t.Fatalf("region mouse konumu floor ile bulundugu hucreye ait olmali: got=(%d,%d)", regionX, regionY)
	}
	if centerX, centerY := regionPaintCellCenterWorld(regionX, regionY); centerX != 14.5 || centerY != 10.5 {
		t.Fatalf("region preview hucenin merkezine oturmali: got=(%.1f,%.1f)", centerX, centerY)
	}
}

func TestShapeOutlineUsesRasterizedWorldBoundary(t *testing.T) {
	point := [2]float32{1016.49, 436.49}
	gotX, gotY := shapeRasterWorldPoint(point)
	wantX := float64(int(shapeOffX + float64(point[0])*shapeScaleX))
	wantY := float64(int(shapeOffY + float64(point[1])*shapeScaleY))
	if gotX != wantX || gotY != wantY {
		t.Fatalf("shape outline raster siniriyle ayni donusumu kullanmali: got=(%.2f,%.2f) want=(%.2f,%.2f)", gotX, gotY, wantX, wantY)
	}
}

func TestShapeBrushRadiusUsesWorldPixelUnits(t *testing.T) {
	oldShapeScaleX, oldShapeScaleY := shapeScaleX, shapeScaleY
	defer func() {
		shapeScaleX, shapeScaleY = oldShapeScaleX, oldShapeScaleY
	}()
	shapeScaleX, shapeScaleY = 2, 2

	session := &shapeEditSession{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10, Width: 11, Height: 11, Mask: make([]byte, 121)}
	r := &Renderer{}
	if !r.applyShapeBrushCircle(session, 5, 5, 0.5, true) {
		t.Fatal("shape brush tek hucreyi boyamali")
	}
	count := 0
	for _, value := range session.Mask {
		count += int(value)
	}
	if count != 1 {
		t.Fatalf("0.5 dunya pikseli yaricapi tek hucre olmali: got=%d", count)
	}

	for i := range session.Mask {
		session.Mask[i] = 0
	}
	if !r.applyShapeBrushCircle(session, 5, 5, 1, true) {
		t.Fatal("shape brush bir dunya pikseli yaricapinda degismeli")
	}
	count = 0
	for _, value := range session.Mask {
		count += int(value)
	}
	if count != 5 {
		t.Fatalf("1 dunya pikseli yaricapi dort komsu ve merkezi icermeli: got=%d", count)
	}
}

func TestSingleShapeWorldPixelRoundTripsThroughRings(t *testing.T) {
	oldShapeOffX, oldShapeOffY := shapeOffX, shapeOffY
	oldShapeScaleX, oldShapeScaleY := shapeScaleX, shapeScaleY
	defer func() {
		shapeOffX, shapeOffY = oldShapeOffX, oldShapeOffY
		shapeScaleX, shapeScaleY = oldShapeScaleX, oldShapeScaleY
	}()
	shapeOffX, shapeOffY = -530, -180
	shapeScaleX, shapeScaleY = 2.025, 2.025

	session := &shapeEditSession{MinX: 1490, MinY: 690, MaxX: 1510, MaxY: 710, Width: 21, Height: 21, Mask: make([]byte, 21*21)}
	session.Mask[session.index(1500, 700)] = 1
	rings := shapeMaskToFloatRings(session)
	if len(rings) != 1 {
		t.Fatalf("tek piksel tek ring uretmeli: got=%d", len(rings))
	}

	roundTrip := &shapeEditSession{MinX: session.MinX, MinY: session.MinY, MaxX: session.MaxX, MaxY: session.MaxY, Width: session.Width, Height: session.Height, Mask: make([]byte, len(session.Mask))}
	for _, ring := range rings {
		rasterizeFloatRingToMask(roundTrip, ring)
	}
	if !roundTrip.filled(1500, 700) {
		t.Fatal("tek shape dunya pikseli ring donusumunde kayboldu")
	}
	count := 0
	for _, value := range roundTrip.Mask {
		count += int(value)
	}
	if count != 1 {
		t.Fatalf("tek shape dunya pikseli round-trip sonrasi tek kalmali: got=%d", count)
	}
}

func TestShapeInspectorToolButtonsToggleOff(t *testing.T) {
	r := &Renderer{}

	shapePaint := editInspectorButtonRect(editButtonShapePaint)
	if _, handled := r.handleEditShapeInspectorClick(shapePaint[0]+shapePaint[2]/2, shapePaint[1]+shapePaint[3]/2); !handled {
		t.Fatal("shape boya butonu tıklanmadı")
	}
	if r.editShapeTool != editShapeToolShape || r.editShapeBrushMode != editShapeBrushPaint {
		t.Fatalf("shape boya seçilmedi: tool=%d mode=%d", r.editShapeTool, r.editShapeBrushMode)
	}
	if _, handled := r.handleEditShapeInspectorClick(shapePaint[0]+shapePaint[2]/2, shapePaint[1]+shapePaint[3]/2); !handled {
		t.Fatal("shape boya ikinci tıklaması işlenmedi")
	}
	if r.editShapeTool != editShapeToolNone {
		t.Fatalf("shape boya ikinci tıklamada kapanmadı: tool=%d", r.editShapeTool)
	}

	regionErase := editInspectorButtonRect(editButtonShapeRegionErase)
	r.handleEditShapeInspectorClick(regionErase[0]+regionErase[2]/2, regionErase[1]+regionErase[3]/2)
	if r.editShapeTool != editShapeToolRegion || r.editShapeBrushMode != editShapeBrushErase {
		t.Fatalf("bölge sil seçilmedi: tool=%d mode=%d", r.editShapeTool, r.editShapeBrushMode)
	}
	r.handleEditShapeInspectorClick(regionErase[0]+regionErase[2]/2, regionErase[1]+regionErase[3]/2)
	if r.editShapeTool != editShapeToolNone {
		t.Fatalf("bölge sil ikinci tıklamada kapanmadı: tool=%d", r.editShapeTool)
	}
}

func TestShapeBrushSupportsTwoFineStepsBelowOnePixelRadius(t *testing.T) {
	r := &Renderer{editShapeBrushRadius: 1}
	brushMinus := editInspectorButtonRect(editButtonShapeBrushMinus)
	click := func() {
		r.handleEditShapeInspectorClick(brushMinus[0]+brushMinus[2]/2, brushMinus[1]+brushMinus[3]/2)
	}

	click()
	if r.editShapeBrushRadius != 0.75 {
		t.Fatalf("ilk ince kademe 0.75 olmali: got=%v", r.editShapeBrushRadius)
	}
	click()
	if r.editShapeBrushRadius != editShapeBrushMinRadius {
		t.Fatalf("ikinci ince kademe %.2f olmali: got=%v", editShapeBrushMinRadius, r.editShapeBrushRadius)
	}
	click()
	if r.editShapeBrushRadius != editShapeBrushMinRadius {
		t.Fatalf("minimum fırça yarıçapının altına inilmemeli: got=%v", r.editShapeBrushRadius)
	}
}

func TestShapeBrushKeepsWholePixelStepsAboveFineRange(t *testing.T) {
	r := &Renderer{editShapeBrushRadius: 6}
	brushMinus := editInspectorButtonRect(editButtonShapeBrushMinus)
	r.handleEditShapeInspectorClick(brushMinus[0]+brushMinus[2]/2, brushMinus[1]+brushMinus[3]/2)
	if r.editShapeBrushRadius != 5 {
		t.Fatalf("büyük fırçada mevcut tam piksel adımı korunmalı: got=%v", r.editShapeBrushRadius)
	}

	r.editShapeBrushRadius = editShapeBrushMinRadius
	brushPlus := editInspectorButtonRect(editButtonShapeBrushPlus)
	for _, want := range []float64{0.75, 1, 2} {
		r.handleEditShapeInspectorClick(brushPlus[0]+brushPlus[2]/2, brushPlus[1]+brushPlus[3]/2)
		if r.editShapeBrushRadius != want {
			t.Fatalf("Firca + kademesi yanlis: want=%v got=%v", want, r.editShapeBrushRadius)
		}
	}
}

func TestShapePaintStrokeUpdatesShapeDataAndWorldMap(t *testing.T) {
	r := newLandShapeEditRenderer()
	r.editInspectorTab = editInspectorShape
	r.editSelectedRegion = "land_test"
	r.editShapeTool = editShapeToolShape
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
	r.editShapeTool = editShapeToolShape
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
	if r.editPaintPreviewImage == nil {
		t.Fatal("canli preview goruntusu olusturulmadi")
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
	r.finishShapePaintStroke()

	if got := r.worldMap.RegionAt(14, 10); got != "land_test" {
		t.Fatalf("paint sonrasi piksel region'a baglanmadi: got=%q", got)
	}
	if len(r.editRegionPaintOverrides) == 0 {
		t.Fatal("region paint overrides kayit edilmedi")
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
