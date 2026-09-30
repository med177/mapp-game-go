package render

import (
	"testing"

	"mapp-game-go/internal/world"
)

func TestSplitEditRegionResourcesPreservesEconomicTotals(t *testing.T) {
	source := &world.Region{
		BaseGoldIncome:   101,
		BaseGrainOutput:  102,
		BaseIronOutput:   103,
		BaseTimberOutput: 104,
		BaseStoneOutput:  105,
		BaseSpiceOutput:  106,
		BaseClothOutput:  107,
		TradeCapacity:    9,
		Population:       201,
		RuralPopulation:  201,
	}

	original := *source
	split := splitEditRegionResources(source)

	if source.BaseGoldIncome != 50 || split.baseGoldIncome != 51 {
		t.Fatalf("altın payları = %d/%d, want 50/51", source.BaseGoldIncome, split.baseGoldIncome)
	}
	if source.BaseStoneOutput != 52 || split.baseStoneOutput != 53 {
		t.Fatalf("taş payları = %d/%d, want 52/53", source.BaseStoneOutput, split.baseStoneOutput)
	}
	if source.TradeCapacity != 4 || split.tradeCapacity != 5 {
		t.Fatalf("ticaret kapasitesi payları = %d/%d, want 4/5", source.TradeCapacity, split.tradeCapacity)
	}
	if source.Population != 100 || source.RuralPopulation != 100 || split.population != 101 || split.ruralPopulation != 101 {
		t.Fatalf("nüfus payları = source %d/%d, child %d/%d; want source 100/100, child 101/101", source.Population, source.RuralPopulation, split.population, split.ruralPopulation)
	}

	checks := []struct {
		name      string
		original  int
		remaining int
		newRegion int
	}{
		{"altın", original.BaseGoldIncome, source.BaseGoldIncome, split.baseGoldIncome},
		{"tahıl", original.BaseGrainOutput, source.BaseGrainOutput, split.baseGrainOutput},
		{"demir", original.BaseIronOutput, source.BaseIronOutput, split.baseIronOutput},
		{"kereste", original.BaseTimberOutput, source.BaseTimberOutput, split.baseTimberOutput},
		{"taş", original.BaseStoneOutput, source.BaseStoneOutput, split.baseStoneOutput},
		{"baharat", original.BaseSpiceOutput, source.BaseSpiceOutput, split.baseSpiceOutput},
		{"kumaş", original.BaseClothOutput, source.BaseClothOutput, split.baseClothOutput},
		{"ticaret", original.TradeCapacity, source.TradeCapacity, split.tradeCapacity},
		{"nüfus", original.Population, source.Population, split.population},
	}
	for _, check := range checks {
		if check.remaining+check.newRegion != check.original {
			t.Errorf("%s toplamı korumadı: %d + %d != %d", check.name, check.remaining, check.newRegion, check.original)
		}
	}
}

func TestSplitEditRegionResourcesKeepsSettlementPopulationIndependent(t *testing.T) {
	source := &world.Region{
		Population:      140,
		RuralPopulation: 100,
		Settlements:     []world.Settlement{{ID: "center"}},
	}

	split := splitEditRegionResources(source)

	if source.RuralPopulation != 50 || source.Population != 70 {
		t.Fatalf("kaynak nüfusu = %d/%d; want toplam 70, kırsal 50", source.Population, source.RuralPopulation)
	}
	if split.population != 70 || split.ruralPopulation != 50 {
		t.Fatalf("yeni bölge nüfusu = %d/%d, want toplam 70, kırsal 50", split.population, split.ruralPopulation)
	}
	if source.Population+split.population != 140 || source.RuralPopulation+split.ruralPopulation != 100 {
		t.Fatalf("toplam nüfus/kırsal korunmadı: %d + %d, %d + %d", source.Population, split.population, source.RuralPopulation, split.ruralPopulation)
	}
}
