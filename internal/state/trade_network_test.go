package state

import (
	"testing"

	"mapp-game-go/internal/world"
)

func TestTradeNetworkModifiersAggregateByCenterAndRegion(t *testing.T) {
	gs := &GameState{
		TradeNetworkModifiers: []TradeNetworkModifier{
			{ID: "old_spice", CenterIDs: []world.RegionID{"egypt", "basra"}, RegionIDs: []world.RegionID{"egypt"}, TradeIncomePercent: -35, SpicePercent: -30},
			{ID: "portuguese_gain", CenterIDs: []world.RegionID{"portugal"}, TradeIncomePercent: 40},
		},
	}

	if got := gs.TradeNetworkIncomeModifier("egypt"); got != -35 {
		t.Fatalf("Mısır ticaret geliri modifier'ı yanlış: got=%d", got)
	}
	if got := gs.RegionSpiceProductionModifier("egypt"); got != -30 {
		t.Fatalf("Mısır baharat modifier'ı yanlış: got=%d", got)
	}
	if got := gs.TradeNetworkIncomeModifier("portugal"); got != 40 {
		t.Fatalf("Portekiz ticaret geliri modifier'ı yanlış: got=%d", got)
	}
	if got := gs.TradeNetworkIncomeModifier("venice"); got != 0 {
		t.Fatalf("ilgisiz merkez etkilenmemeli: got=%d", got)
	}
}

func TestTradeNetworkModifierReplacementByID(t *testing.T) {
	gs := &GameState{}
	modifiers := []TradeNetworkModifier{
		{ID: "spice_monopoly", CenterIDs: []world.RegionID{"egypt"}, TradeIncomePercent: -35},
	}
	gs.TradeNetworkModifiers = modifiers
	gs.TradeNetworkModifiers[0].TradeIncomePercent = -45

	if got := gs.TradeNetworkIncomeModifier("egypt"); got != -45 {
		t.Fatalf("aynı event modifier'ı güncel değerle uygulanmalı: got=%d", got)
	}
}
