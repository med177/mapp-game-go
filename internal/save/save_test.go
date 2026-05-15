package save

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestLoadFromPathRehydratesScenarioRuntimeFromScenarioID(t *testing.T) {
	tmp := t.TempDir()
	oldBaseDir := scenarioBaseDir
	scenarioBaseDir = filepath.Join(tmp, "scenarios")
	defer func() { scenarioBaseDir = oldBaseDir }()

	scenarioID := "test_scenario"
	scenarioPath := filepath.Join(scenarioBaseDir, scenarioID)
	if err := os.MkdirAll(filepath.Join(scenarioPath, "data"), 0755); err != nil {
		t.Fatalf("scenario data dir olusmadi: %v", err)
	}

	writeJSONFile(t, filepath.Join(scenarioPath, "scenario.json"), scenario.Scenario{
		ID:    scenarioID,
		Year:  1453,
		Month: 4,
		MapConfig: scenario.MapConfig{
			WorldWidth: intPtr(64),
		},
		VictoryConditions: []scenario.VictoryOptionDef{
			{ID: "economic", Type: "economic", TargetGoldIncome: 500, GoldHoldTurns: 5},
		},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "regions.json"), []*world.Region{
		{ID: "r1", NameTR: "R1", OwnerID: "player", ShapeID: "AAA", TaxRate: 50, Satisfaction: 50},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "units.json"), []map[string]any{
		{"id": "militia", "name": "Militia", "name_tr": "Milis", "category": "infantry", "attack": 5, "defense": 4, "morale": 10, "hp": 100, "gold_cost": 10, "grain_upkeep": 1, "turns_required": 1},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "buildings.json"), []map[string]any{
		{"id": "market", "name": "Market", "name_tr": "Pazar", "gold_cost": 50, "turns_required": 1, "gold_mod": 1.2, "grain_mod": 1.0, "sat_bonus": 0, "def_bonus": 0, "max_per_region": 1},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "technologies.json"), []map[string]any{
		{"id": "tax", "name_tr": "Vergi", "category": "economy", "description_tr": "vergi", "gold_cost": 20, "turns_required": 1, "requires": []string{}, "effects": map[string]any{"gold_per_region": 5}},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "country_shapes.json"), map[string]any{
		"shapes": []map[string]any{
			{"id": "AAA", "name": "AAA", "rings": [][][]int{{{1, 1}, {4, 1}, {4, 4}, {1, 4}}}},
		},
	})

	savePath := filepath.Join(tmp, "slot.json")
	writeJSONFile(t, savePath, &state.GameState{
		ScenarioID:      scenarioID,
		ScenarioPath:    "",
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", NameTR: "R1", OwnerID: "player", ShapeID: "AAA", TaxRate: 50, Satisfaction: 50},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
		},
		Armies: map[army.ArmyID]*army.Army{},
	})

	gs, err := loadFromPath(savePath)
	if err != nil {
		t.Fatalf("loadFromPath hata verdi: %v", err)
	}

	if gs.ScenarioPath != scenarioPath {
		t.Fatalf("scenario path resolve olmadi: got=%q want=%q", gs.ScenarioPath, scenarioPath)
	}
	if len(gs.AvailableVictories) != 1 || gs.AvailableVictories[0].ID != "economic" {
		t.Fatalf("victory metadata geri yuklenmedi: %+v", gs.AvailableVictories)
	}
	if gs.MapConfig.WorldWidth == nil || *gs.MapConfig.WorldWidth != 64 {
		t.Fatalf("map config geri yuklenmedi: %+v", gs.MapConfig)
	}
	if len(gs.RegionOrder) != 1 || gs.RegionOrder[0] != "r1" {
		t.Fatalf("region order geri yuklenmedi: %+v", gs.RegionOrder)
	}
	if gs.UnitTypes["militia"] == nil || gs.BuildingTypes["market"] == nil || gs.TechTypes["tax"] == nil {
		t.Fatal("runtime tipleri geri yuklenmedi")
	}
	if len(gs.ShapeData.Shapes["AAA"]) == 0 {
		t.Fatal("shape data geri yuklenmedi")
	}
}

func writeJSONFile(t *testing.T, path string, payload any) {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal hatasi: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("json dosyasi yazilamadi: %v", err)
	}
}

func intPtr(v int) *int {
	return &v
}
