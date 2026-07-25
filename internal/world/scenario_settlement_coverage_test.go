package world

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test1300ScenarioEveryLandRegionHasSettlement(t *testing.T) {
	dataDir := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data")
	regions, err := LoadRegions(filepath.Join(dataDir, "regions.json"))
	if err != nil {
		t.Fatalf("1300 bölgeleri yüklenemedi: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dataDir, "settlements.json"))
	if err != nil {
		t.Fatalf("1300 yerleşimleri okunamadı: %v", err)
	}
	var entries []SettlementListEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("1300 yerleşimleri parse edilemedi: %v", err)
	}

	settlementCounts := make(map[RegionID]int, len(entries))
	for _, entry := range entries {
		if _, ok := regions[entry.RegionID]; !ok {
			t.Errorf("yerleşim kaydı bilinmeyen bölgeye bağlı: %s", entry.RegionID)
			continue
		}
		settlementCounts[entry.RegionID] += len(entry.Settlements)
	}

	for regionID, region := range regions {
		if region == nil || region.IsSea {
			continue
		}
		if settlementCounts[regionID] == 0 {
			t.Errorf("kara bölgesinde yerleşim yok: %s", regionID)
		}
	}
}

func Test1300ScenarioSettlementPopulationLeavesRuralPopulation(t *testing.T) {
	dataDir := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data")
	regions, err := LoadRegions(filepath.Join(dataDir, "regions.json"))
	if err != nil {
		t.Fatalf("1300 bölgeleri yüklenemedi: %v", err)
	}
	if err := LoadRegionSettlements(filepath.Join(dataDir, "settlements.json"), regions); err != nil {
		t.Fatalf("1300 yerleşimleri yüklenemedi: %v", err)
	}

	for regionID, region := range regions {
		if region == nil || region.IsSea {
			continue
		}
		settlementPopulation := region.SettlementPopulation()
		if region.Population != region.RuralPopulation+settlementPopulation {
			t.Errorf("%s nüfus toplamı tutarsız: population=%d rural=%d settlement=%d", regionID, region.Population, region.RuralPopulation, settlementPopulation)
		}
		if region.Population > 0 && region.RuralPopulation <= settlementPopulation {
			t.Errorf("%s kırsal nüfusu yerleşim toplamından büyük olmalı: rural=%d settlement=%d", regionID, region.RuralPopulation, settlementPopulation)
		}
	}
}
