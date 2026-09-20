package render

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestSortedFactionsForMarketSortsSellersBySupply(t *testing.T) {
	gs := marketSortingTestState()
	gs.SetMarketSellOffer("seller_low", economy.GoodIron, 10)
	gs.SetMarketSellOffer("seller_high", economy.GoodIron, 30)

	got := sortedFactionsForMarket(gs, 1, TradeListSellers, TradeSortDistance)
	want := []faction.FactionID{"seller_high", "seller_low"}
	assertFactionIDs(t, got, want)
}

func TestSortedFactionsForMarketSortsBuyersByDemand(t *testing.T) {
	gs := marketSortingTestState()
	gs.SetMarketBuyOrder("buyer_low", economy.GoodIron, 10)
	gs.SetMarketBuyOrder("buyer_high", economy.GoodIron, 30)

	got := sortedFactionsForMarket(gs, 1, TradeListBuyers, TradeSortDistance)
	want := []faction.FactionID{"buyer_high", "buyer_low"}
	assertFactionIDs(t, got, want)
}

func marketSortingTestState() *state.GameState {
	return &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":      {ID: "player", Gold: 1000, Iron: 100},
			"seller_low":  {ID: "seller_low", Gold: 100, Iron: 10},
			"seller_high": {ID: "seller_high", Gold: 100, Iron: 30},
			"buyer_low":   {ID: "buyer_low", Gold: 1000, Iron: 0},
			"buyer_high":  {ID: "buyer_high", Gold: 1000, Iron: 0},
		},
		MarketOrders: state.MarketOrderBook{
			SellOffers: make(map[faction.FactionID]map[economy.GoodType]int),
			BuyOrders:  make(map[faction.FactionID]map[economy.GoodType]int),
		},
	}
}

func assertFactionIDs(t *testing.T, got, want []faction.FactionID) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("devlet sayısı = %d, want %d (%v)", len(got), len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("devlet sırası = %v, want %v", got, want)
		}
	}
}
