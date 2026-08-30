package state

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestTradePowerUsesCapacityRoutesAndOwnedCenters(t *testing.T) {
	gs := &GameState{
		Year: 1300,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", Spice: 10}, "b": {ID: "b", Gold: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"center": {ID: "center", OwnerID: "a", TradeCapacity: 4},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{{ID: "center", Tier: world.TradeCenterPrimary}}},
		TradeRoutes:  []*economy.TradeRoute{{FromFactionID: "a", ToFactionID: "b", AmountPerTurn: 2, GoldPerUnit: 5}},
	}

	if got := gs.TradePowerForFaction("a"); got != 100 {
		t.Fatalf("a ticaret gücü 100 olmalıydı, got=%d", got)
	}
	if got := gs.TradePowerForFaction("b"); got != 10 {
		t.Fatalf("b rota ticaret gücü 10 olmalıydı, got=%d", got)
	}
	if got := gs.TradePowerSharePercent("a"); got != 90 {
		t.Fatalf("a ticaret gücü payı %%90 olmalıydı, got=%d", got)
	}
}

func TestTradePowerShareCanRaiseCustomsRate(t *testing.T) {
	gs := &GameState{
		Year: 1300,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", Spice: 10}, "b": {ID: "b", Gold: 100},
		},
		Regions:     map[world.RegionID]*world.Region{"center": {ID: "center", OwnerID: "b", TradeCapacity: 4}},
		TradeRoutes: []*economy.TradeRoute{{FromFactionID: "a", ToFactionID: "b", Good: economy.GoodSpice, AmountPerTurn: 2, GoldPerUnit: 10}},
	}
	_, transfers := economy.ApplyTradeRoutesWithTransfersAndCustoms(gs.Factions, gs.TradeRoutes, func(importer faction.FactionID) int {
		return economy.TradeRouteCustomsRatePercent + gs.TradePowerSharePercent(importer)/10
	})
	if len(transfers) != 1 || transfers[0].CustomsAmount != 3 {
		t.Fatalf("ticaret gücü gümrüğü artırmalıydı: %+v", transfers)
	}
}

func TestTradePowerCommerceIncomeUsesCenterShare(t *testing.T) {
	gs := &GameState{
		Year:         1300,
		Factions:     map[faction.FactionID]*faction.Faction{"a": {ID: "a", Spice: 10}, "b": {ID: "b", Gold: 100}},
		Regions:      map[world.RegionID]*world.Region{"center": {ID: "center", OwnerID: "a", TradeCapacity: 4}},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{{ID: "center", Tier: world.TradeCenterPrimary}}},
		TradeRoutes:  []*economy.TradeRoute{{FromFactionID: "a", ToFactionID: "b", Good: economy.GoodSpice, AmountPerTurn: 2, GoldPerUnit: 10}},
	}
	if got := gs.TradePowerCommerceIncome("a"); got != 3 {
		t.Fatalf("merkez payı geliri 3 olmalıydı, got=%d", got)
	}
}
