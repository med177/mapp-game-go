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
	rawTradeData, err := os.ReadFile(filepath.Join(dataDir, "trade_centers.json"))
	if err != nil {
		t.Fatalf("ticaret merkezi JSON çıktısı okunamadı: %v", err)
	}
	var rawTradeConfig struct {
		Centers []map[string]json.RawMessage `json:"centers"`
	}
	if err := json.Unmarshal(rawTradeData, &rawTradeConfig); err != nil {
		t.Fatalf("ticaret merkezi JSON çıktısı parse edilemedi: %v", err)
	}
	if len(rawTradeConfig.Centers) != 1 {
		t.Fatalf("ticaret merkezi JSON çıktısı beklenmeyen merkez sayısına sahip: %d", len(rawTradeConfig.Centers))
	}
	if _, ok := rawTradeConfig.Centers[0]["region_id"]; !ok {
		t.Fatal("ticaret merkezi JSON çıktısında region_id yok")
	}
	if _, ok := rawTradeConfig.Centers[0]["id"]; ok {
		t.Fatal("ticaret merkezi JSON çıktısında eski id alanı kaldı")
	}
}

func TestWriteScenarioSettlementsSkipsSeaRegions(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("scenario data dizini oluşturulamadı: %v", err)
	}

	gs := &state.GameState{
		ScenarioPath: root,
		RegionOrder:  []world.RegionID{"land", "sea"},
		Regions: map[world.RegionID]*world.Region{
			"land": {
				ID:          "land",
				Settlements: []world.Settlement{{ID: "land_city", IsCenter: true}},
			},
			"sea": {
				ID:          "sea",
				IsSea:       true,
				Settlements: []world.Settlement{{ID: "invalid_sea_settlement"}},
			},
		},
	}

	if err := writeScenarioSettlements(gs); err != nil {
		t.Fatalf("settlement verisi yazılamadı: %v", err)
	}

	var entries []world.SettlementListEntry
	readScenarioEditTestJSON(t, filepath.Join(dataDir, "settlements.json"), &entries)
	if len(entries) != 1 || entries[0].RegionID != "land" {
		t.Fatalf("deniz region settlement çıktısına yazıldı: %+v", entries)
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
