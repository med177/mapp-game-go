package scenario

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func scenario1300IntegrityPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "assets", "scenarios", "1300_ottoman_rise")
}

func read1300JSON(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s okunamadi: %v", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("%s parse edilemedi: %v", path, err)
	}
}

func load1300IntegrityData(t *testing.T) (string, map[world.RegionID]*world.Region, map[faction.FactionID]*faction.Faction) {
	t.Helper()
	scenarioPath := scenario1300IntegrityPath(t)
	dataPath := filepath.Join(scenarioPath, "data")

	regions, err := world.LoadRegions(filepath.Join(dataPath, "regions.json"))
	if err != nil {
		t.Fatalf("1300 bölgeleri yüklenemedi: %v", err)
	}
	if err := world.LoadRegionSettlements(filepath.Join(dataPath, "settlements.json"), regions); err != nil {
		t.Fatalf("1300 yerleşimleri yüklenemedi: %v", err)
	}
	factions, err := faction.LoadFactions(filepath.Join(dataPath, "factions.json"))
	if err != nil {
		t.Fatalf("1300 devletleri yüklenemedi: %v", err)
	}
	return scenarioPath, regions, factions
}

func Test1300ScenarioArmyReferencesExist(t *testing.T) {
	scenarioPath, regions, factions := load1300IntegrityData(t)
	dataPath := filepath.Join(scenarioPath, "data")
	unitTypes, err := army.LoadUnitTypes(filepath.Join(dataPath, "units.json"))
	if err != nil {
		t.Fatalf("1300 birlik tipleri yüklenemedi: %v", err)
	}
	armies, err := army.LoadArmies(filepath.Join(dataPath, "armies.json"), unitTypes)
	if err != nil {
		t.Fatalf("1300 orduları yüklenemedi: %v", err)
	}

	for armyID, candidate := range armies {
		if candidate == nil {
			t.Errorf("nil ordu kaydi: %s", armyID)
			continue
		}
		if factions[faction.FactionID(candidate.OwnerID)] == nil {
			t.Errorf("ordu bilinmeyen devlete bağlı: army=%s faction=%s", armyID, candidate.OwnerID)
		}
		if regions[candidate.RegionID] == nil {
			t.Errorf("ordu bilinmeyen bölgeye bağlı: army=%s region=%s", armyID, candidate.RegionID)
		}
		for _, unit := range candidate.Units {
			if unitTypes[unit.TypeID] == nil {
				t.Errorf("ordu bilinmeyen birlik tipi taşıyor: army=%s unit=%s", armyID, unit.TypeID)
			}
		}
	}
}

func Test1300ScenarioVictoryRegionReferencesExist(t *testing.T) {
	scenarioPath, regions, _ := load1300IntegrityData(t)
	var definition Scenario
	read1300JSON(t, filepath.Join(scenarioPath, "scenario.json"), &definition)

	for _, option := range definition.VictoryConditions {
		for _, regionID := range option.RegionTargets() {
			if regions[world.RegionID(regionID)] == nil {
				t.Errorf("zafer hedefi bilinmeyen bölgeye bağlı: victory=%s region=%s", option.ID, regionID)
			}
		}
	}
}

func Test1300ScenarioFactionReferencesExist(t *testing.T) {
	scenarioPath, _, factions := load1300IntegrityData(t)

	for factionID, definition := range factions {
		for _, targetID := range definition.AIExpansionTargets {
			if factions[targetID] == nil {
				t.Errorf("AI genişleme hedefi bilinmeyen devlete bağlı: faction=%s target=%s", factionID, targetID)
			}
		}
	}

	var relations []*faction.Relation
	read1300JSON(t, filepath.Join(scenarioPath, "data", "relations.json"), &relations)
	for index, relation := range relations {
		if relation == nil {
			t.Errorf("nil ilişki kaydı: index=%d", index)
			continue
		}
		if factions[relation.FactionA] == nil {
			t.Errorf("ilişki bilinmeyen devlete bağlı: index=%d faction_a=%s", index, relation.FactionA)
		}
		if factions[relation.FactionB] == nil {
			t.Errorf("ilişki bilinmeyen devlete bağlı: index=%d faction_b=%s", index, relation.FactionB)
		}
		if relation.FactionA == relation.FactionB {
			t.Errorf("ilişki aynı devleti iki tarafta kullanıyor: index=%d faction=%s", index, relation.FactionA)
		}
	}
}

func Test1300ScenarioCapitalSettlementsExist(t *testing.T) {
	_, regions, factions := load1300IntegrityData(t)
	settlementRegions := make(map[string]world.RegionID)
	for regionID, region := range regions {
		if region == nil {
			continue
		}
		for _, settlement := range region.Settlements {
			if previous, duplicate := settlementRegions[settlement.ID]; duplicate {
				t.Errorf("settlement ID birden fazla bölgede kullanılıyor: settlement=%s first=%s second=%s", settlement.ID, previous, regionID)
				continue
			}
			settlementRegions[settlement.ID] = regionID
		}
	}

	for factionID, definition := range factions {
		if definition.CapitalSettlementID == "" {
			t.Errorf("devletin başkent settlement kimliği boş: faction=%s", factionID)
			continue
		}
		regionID, ok := settlementRegions[definition.CapitalSettlementID]
		if !ok {
			t.Errorf("devletin başkenti bilinmeyen settlement'a bağlı: faction=%s settlement=%s", factionID, definition.CapitalSettlementID)
			continue
		}
		if region := regions[regionID]; region == nil || region.OwnerID != string(factionID) {
			ownerID := ""
			if region != nil {
				ownerID = region.OwnerID
			}
			t.Errorf("başkent settlement'ı farklı devletin bölgesinde: faction=%s settlement=%s region=%s owner=%s", factionID, definition.CapitalSettlementID, regionID, ownerID)
		}
	}
}

func Test1300ScenarioTradeCenterReferencesExist(t *testing.T) {
	scenarioPath, regions, _ := load1300IntegrityData(t)
	var config world.TradeCenterConfig
	read1300JSON(t, filepath.Join(scenarioPath, "data", "trade_centers.json"), &config)

	centers := make(map[world.RegionID]world.TradeCenterDef, len(config.Centers))
	for _, center := range config.Centers {
		if center.ID == "" {
			t.Error("ticaret merkezi kimliği boş")
			continue
		}
		if _, duplicate := centers[center.ID]; duplicate {
			t.Errorf("ticaret merkezi kimliği tekrarlanıyor: center=%s", center.ID)
			continue
		}
		centers[center.ID] = center
		if !center.OffMap && regions[center.ID] == nil {
			t.Errorf("ticaret merkezi bilinmeyen bölgeye bağlı: center=%s", center.ID)
		}
	}

	for _, center := range config.Centers {
		for _, linkID := range center.Links {
			if _, ok := centers[linkID]; !ok {
				t.Errorf("ticaret merkezi bilinmeyen linke bağlı: center=%s link=%s", center.ID, linkID)
			}
			if linkID == center.ID {
				t.Errorf("ticaret merkezi kendisine bağlı: center=%s", center.ID)
			}
		}
	}
}

func Test1300ScenarioAIStrategyReferencesExist(t *testing.T) {
	scenarioPath, regions, factions := load1300IntegrityData(t)
	strategies, err := LoadAIStrategies(filepath.Join(scenarioPath, "data", "ai_strategies.json"))
	if err != nil {
		t.Fatalf("1300 AI stratejileri yüklenemedi: %v", err)
	}
	if len(strategies) == 0 {
		t.Fatal("1300 AI strateji profilleri boş")
	}
	for factionID, strategy := range strategies {
		if factions[faction.FactionID(factionID)] == nil {
			t.Errorf("AI profili bilinmeyen devlete bağlı: faction=%s", factionID)
		}
		for _, objective := range strategy.Objectives {
			for _, targetID := range objective.TargetFactions {
				if factions[faction.FactionID(targetID)] == nil {
					t.Errorf("AI objective bilinmeyen devleti hedefliyor: objective=%s target=%s", objective.ID, targetID)
				}
			}
			for _, regionIDs := range [][]string{objective.TargetRegions, objective.ReadinessRegions, objective.AnnexRegionIDs} {
				for _, regionID := range regionIDs {
					if regions[world.RegionID(regionID)] == nil {
						t.Errorf("AI objective bilinmeyen bölgeye bağlı: objective=%s region=%s", objective.ID, regionID)
					}
				}
			}
		}
	}
}
