package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestWriteScenarioEditDataWritesAIStrategiesAndTradeCenters(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("scenario data dizini oluşturulamadı: %v", err)
	}

	gs := &state.GameState{
		ScenarioPath: root,
		AIStrategies: map[string]scenario.AIFactionStrategy{
			"faction_b": {FactionID: "faction_b"},
			"faction_a": {FactionID: "faction_a", Objectives: []scenario.AIObjectiveDef{{TargetRegions: []string{"new_region"}}}},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{{ID: "new_region", Links: []world.RegionID{"faction_a"}}}},
	}

	if err := writeScenarioAIStrategies(gs); err != nil {
		t.Fatalf("AI stratejileri yazılamadı: %v", err)
	}
	if err := writeScenarioTradeCenters(gs); err != nil {
		t.Fatalf("ticaret merkezleri yazılamadı: %v", err)
	}

	var aiConfig scenario.AIStrategyConfig
	readScenarioEditTestJSON(t, filepath.Join(dataDir, "ai_strategies.json"), &aiConfig)
	if len(aiConfig.Factions) != 2 || aiConfig.Factions[0].FactionID != "faction_a" || aiConfig.Factions[1].FactionID != "faction_b" {
		t.Fatalf("AI stratejileri deterministik yazılmadı: %+v", aiConfig.Factions)
	}

	var tradeConfig world.TradeCenterConfig
	readScenarioEditTestJSON(t, filepath.Join(dataDir, "trade_centers.json"), &tradeConfig)
	if len(tradeConfig.Centers) != 1 || tradeConfig.Centers[0].ID != "new_region" || tradeConfig.Centers[0].Links[0] != "faction_a" {
		t.Fatalf("ticaret merkezleri yazılmadı: %+v", tradeConfig.Centers)
	}
}

func readScenarioEditTestJSON(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s okunamadı: %v", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("%s parse edilemedi: %v", path, err)
	}
}
