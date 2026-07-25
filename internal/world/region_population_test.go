package world

import "testing"

func TestRegionRecalculatePopulationCombinesRuralAndSettlements(t *testing.T) {
	region := &Region{
		Population:      999,
		RuralPopulation: 700,
		Settlements: []Settlement{
			{ID: "city", Population: 200},
			{ID: "town", Population: 100},
		},
	}

	if got := region.RecalculatePopulation(); got != 1000 {
		t.Fatalf("bölge nüfusu kırsal ve yerleşim toplamı olmalıydı, got=%d", got)
	}
	if region.SettlementPopulation() != 300 || region.RuralPopulation != 700 {
		t.Fatalf("nüfus bileşenleri korunmalıydı: rural=%d settlement=%d", region.RuralPopulation, region.SettlementPopulation())
	}
}

func TestRegionRecalculatePopulationMigratesLegacyTotalToRural(t *testing.T) {
	region := &Region{
		Population: 520,
		Settlements: []Settlement{
			{ID: "city", Population: 20},
		},
	}

	if got := region.RecalculatePopulation(); got != 520 || region.RuralPopulation != 500 {
		t.Fatalf("eski toplam nüfusun yerleşim dışındaki kısmı kırsala taşınmalıydı: population=%d rural=%d", got, region.RuralPopulation)
	}
}
