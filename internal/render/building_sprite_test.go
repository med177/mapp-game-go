package render

import (
	"path/filepath"
	"testing"

	"mapp-game-go/internal/city"
)

func TestBuildingSpritePathUsesBuildingID(t *testing.T) {
	oldScenarioPath := ActiveScenarioPath
	t.Cleanup(func() {
		ActiveScenarioPath = oldScenarioPath
		resetBuildingSpriteCache()
	})

	ActiveScenarioPath = filepath.Join("assets", "scenarios", "1300_ottoman_rise")
	tests := map[string]string{
		"market":   "market.png",
		"farm":     "farm.png",
		"barracks": "barracks.png",
		"port":     "port.png",
		"walls":    "walls.png",
		"temple":   "temple.png",
		"granary":  "granary.png",
	}
	for id, filename := range tests {
		want := filepath.Join(ActiveScenarioPath, "sprites", "buildings", filename)
		if got := buildingSpritePath(id); got != want {
			t.Errorf("%s bina sprite yolu yanlış: got=%q want=%q", id, got, want)
		}
	}
}

func TestBuildingSpriteImageLoadsScenarioAssetsByID(t *testing.T) {
	oldScenarioPath := ActiveScenarioPath
	t.Cleanup(func() {
		ActiveScenarioPath = oldScenarioPath
		resetBuildingSpriteCache()
	})

	ActiveScenarioPath, _ = filepath.Abs(filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	resetBuildingSpriteCache()
	buildings, err := city.LoadBuildings(filepath.Join(ActiveScenarioPath, "data", "buildings.json"))
	if err != nil {
		t.Fatalf("senaryo binaları yüklenemedi: %v", err)
	}
	for id := range buildings {
		if sprite := buildingSpriteImage(id); sprite == nil {
			t.Errorf("%s bina sprite'ı yüklenemedi", id)
		}
	}
}
