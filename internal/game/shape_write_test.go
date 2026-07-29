package game

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"mapp-game-go/internal/render"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestWriteScenarioShapesWritesShapeFile(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, "data"), 0755); err != nil {
		t.Fatalf("data dir olusmadi: %v", err)
	}
	gs := &state.GameState{
		ScenarioPath: tmp,
		ShapeData: world.CountryShapeJSON{
			Shapes: map[string][][][2]float32{
				"AAA": {{{1, 2}, {4, 2}, {4, 5}, {1, 5}}},
			},
			Names: map[string]string{"AAA": "Test Shape"},
		},
	}

	if err := writeScenarioShapes(gs); err != nil {
		t.Fatalf("writeScenarioShapes hata verdi: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, "data", "country_shapes.json"))
	if err != nil {
		t.Fatalf("shape dosyasi okunamadi: %v", err)
	}
	var payload struct {
		Shapes []struct {
			ID    string         `json:"id"`
			Name  string         `json:"name"`
			Rings [][][2]float32 `json:"rings"`
		} `json:"shapes"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("shape dosyasi parse edilemedi: %v", err)
	}
	if len(payload.Shapes) != 1 {
		t.Fatalf("beklenen 1 shape, got=%d", len(payload.Shapes))
	}
	if payload.Shapes[0].ID != "AAA" || payload.Shapes[0].Name != "Test Shape" {
		t.Fatalf("beklenmeyen shape metadata: %+v", payload.Shapes[0])
	}
	if len(payload.Shapes[0].Rings) != 1 || len(payload.Shapes[0].Rings[0]) != 4 {
		t.Fatalf("beklenmeyen ring verisi: %+v", payload.Shapes[0].Rings)
	}
}

func TestWriteScenarioShapesPreservesSubPixelBoundaries(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, "data"), 0755); err != nil {
		t.Fatalf("data dir olusmadi: %v", err)
	}

	// Editor maskesi dunya pikseli sinirini, senaryo koordinatinda ondalikli
	// bir noktaya cevirir. 2.025 olcekte bunu tam sayiya yuvarlamak dunya
	// pikselini kaydirir; kayit/yukleme ayni koordinati korumali.
	scale := float32(2.025)
	boundaryX := float32(15) / scale
	boundaryY := float32(8) / scale
	ring := [][2]float32{
		{boundaryX, boundaryY},
		{float32(19) / scale, boundaryY},
		{float32(19) / scale, float32(12) / scale},
		{boundaryX, float32(12) / scale},
	}
	gs := &state.GameState{
		ScenarioPath: tmp,
		ShapeData: world.CountryShapeJSON{
			Shapes: map[string][][][2]float32{"AAA": {ring}},
			Names:  map[string]string{"AAA": "Hassas Shape"},
		},
	}

	if err := writeScenarioShapes(gs); err != nil {
		t.Fatalf("hassas shape yazilamadi: %v", err)
	}
	reloaded, err := world.LoadCountryShapes(filepath.Join(tmp, "data", "country_shapes.json"), nil)
	if err != nil {
		t.Fatalf("hassas shape tekrar yuklenemedi: %v", err)
	}
	got := reloaded.Shapes["AAA"][0]
	if len(got) != len(ring) {
		t.Fatalf("ring noktasi sayisi degisti: got=%d want=%d", len(got), len(ring))
	}
	for i := range ring {
		for axis := 0; axis < 2; axis++ {
			if math.Abs(float64(got[i][axis]-ring[i][axis])) > 0.00001 {
				t.Fatalf("sub-pixel koordinat kayboldu: point=%d axis=%d got=%v want=%v", i, axis, got[i][axis], ring[i][axis])
			}
		}
	}

	worldW, worldH := 64, 64
	offsetX, offsetY := 0.0, 0.0
	scaleX, scaleY := 2.025, 2.025
	regions := map[world.RegionID]*world.Region{
		"AAA": {
			ID:      "AAA",
			WorldX:  10,
			WorldY:  10,
			ShapeID: "AAA",
		},
	}
	gs.MapConfig = scenario.MapConfig{
		WorldWidth:   &worldW,
		WorldHeight:  &worldH,
		ShapeOffsetX: &offsetX,
		ShapeOffsetY: &offsetY,
		ShapeScaleX:  &scaleX,
		ShapeScaleY:  &scaleY,
	}
	gs.Regions = regions
	gs.RegionOrder = []world.RegionID{"AAA"}
	beforeMap := render.NewWorldMap(gs)

	reloadedRegions := map[world.RegionID]*world.Region{
		"AAA": {
			ID:      "AAA",
			WorldX:  10,
			WorldY:  10,
			ShapeID: "AAA",
		},
	}
	gsReload := &state.GameState{
		ScenarioPath: tmp,
		MapConfig:    gs.MapConfig,
		Regions:      reloadedRegions,
		RegionOrder:  []world.RegionID{"AAA"},
		ShapeData:    reloaded,
	}
	afterMap := render.NewWorldMap(gsReload)
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			if beforeMap.RegionAt(x, y) != afterMap.RegionAt(x, y) {
				t.Fatalf("yaz-kaydet-yukle sonrasi raster degisti: x=%d y=%d before=%q after=%q", x, y, beforeMap.RegionAt(x, y), afterMap.RegionAt(x, y))
			}
		}
	}
}

func TestWriteScenarioShapesSyncsLandRegionPaintIntoCountryShapes(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, "data"), 0755); err != nil {
		t.Fatalf("data dir olusmadi: %v", err)
	}
	worldW := 64
	worldH := 64
	offset := 0.0
	scale := 1.0
	regionID := world.RegionID("land_test")
	shapeID := "land_shape"
	gs := &state.GameState{
		ScenarioPath: tmp,
		MapConfig: scenario.MapConfig{
			WorldWidth:   &worldW,
			WorldHeight:  &worldH,
			ShapeOffsetX: &offset,
			ShapeOffsetY: &offset,
			ShapeScaleX:  &scale,
			ShapeScaleY:  &scale,
		},
		Regions: map[world.RegionID]*world.Region{
			regionID: {
				ID:      regionID,
				NameTR:  "Kara Test",
				Terrain: world.TerrainPlain,
				WorldX:  10,
				WorldY:  10,
				ShapeID: shapeID,
				Shape: [][][2]float32{{
					{8, 8}, {12, 8}, {12, 12}, {8, 12},
				}},
			},
		},
		RegionOrder: []world.RegionID{regionID},
		ShapeData: world.CountryShapeJSON{
			Shapes: map[string][][][2]float32{
				shapeID: {{{8, 8}, {12, 8}, {12, 12}, {8, 12}}},
			},
			Names: map[string]string{shapeID: "Land Shape"},
		},
		RegionPaintOverrides: map[int]world.RegionID{
			10*worldW + 14: regionID,
		},
	}

	if err := writeScenarioShapes(gs); err != nil {
		t.Fatalf("writeScenarioShapes hata verdi: %v", err)
	}

	reloadedRegions := map[world.RegionID]*world.Region{
		regionID: {
			ID:      regionID,
			NameTR:  "Kara Test",
			Terrain: world.TerrainPlain,
			WorldX:  10,
			WorldY:  10,
			ShapeID: shapeID,
		},
	}
	shapeData, err := world.LoadCountryShapes(filepath.Join(tmp, "data", "country_shapes.json"), reloadedRegions)
	if err != nil {
		t.Fatalf("yazilan shape dosyasi tekrar yuklenemedi: %v", err)
	}
	gsReload := &state.GameState{
		ScenarioPath: tmp,
		MapConfig: scenario.MapConfig{
			WorldWidth:   &worldW,
			WorldHeight:  &worldH,
			ShapeOffsetX: &offset,
			ShapeOffsetY: &offset,
			ShapeScaleX:  &scale,
			ShapeScaleY:  &scale,
		},
		Regions:     reloadedRegions,
		RegionOrder: []world.RegionID{regionID},
		ShapeData:   shapeData,
	}
	wm := render.NewWorldMap(gsReload)
	if got := wm.RegionAt(14, 10); got != regionID {
		t.Fatalf("country_shapes sync sonrasi dis piksel region'a baglanmali, got=%q", got)
	}
}
