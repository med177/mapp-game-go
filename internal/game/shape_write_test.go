package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

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
			ID    string       `json:"id"`
			Name  string       `json:"name"`
			Rings [][][2]int   `json:"rings"`
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