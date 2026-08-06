package state

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
)

func TestMarketOrdersClampTransactionsToOfferAndGoldBackedDemand(t *testing.T) {
	playerID := faction.FactionID("player")
	sellerID := faction.FactionID("seller")
	buyerID := faction.FactionID("buyer")
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			playerID: {ID: playerID, Iron: 100, Gold: 1000},
			sellerID: {ID: sellerID, Iron: 500, Gold: 100},
			buyerID:  {ID: buyerID, Iron: 10, Gold: 12},
		},
	}
	gs.SetMarketSellOffer(sellerID, economy.GoodIron, 80)
	gs.SetMarketBuyOrder(buyerID, economy.GoodIron, 10)

	if got := gs.MarketSellOffer(sellerID, economy.GoodIron); got != 80 {
		t.Fatalf("satış arzı stoktan bağımsız büyümemeli ama tam kota görünmeli: got=%d", got)
	}
	if got := gs.MarketBuyOrder(buyerID, economy.GoodIron, 5); got != 2 {
		t.Fatalf("alım talebi güncel altınla sınırlandırılmalı: got=%d want=2", got)
	}
	gs.ConsumeMarketSellOffer(sellerID, economy.GoodIron, 30)
	gs.ConsumeMarketBuyOrder(buyerID, economy.GoodIron, 3)
	if got := gs.MarketSellOffer(sellerID, economy.GoodIron); got != 50 {
		t.Fatalf("başarılı alım satış arzını azaltmalı: got=%d want=50", got)
	}
	gs.Factions[buyerID].Gold = 100
	if got := gs.MarketBuyOrder(buyerID, economy.GoodIron, 5); got != 7 {
		t.Fatalf("başarılı satış alım talebini azaltmalı: got=%d want=7", got)
	}

	gs.Factions[sellerID].Iron = 20
	if got := gs.MarketSellOffer(sellerID, economy.GoodIron); got != 20 {
		t.Fatalf("eski emir gerçek stoktan büyükse stokla kırpılmalı: got=%d want=20", got)
	}
}

func TestMarketOrdersCloneIsDeepCopy(t *testing.T) {
	gs := &GameState{}
	gs.SetMarketSellOffer("seller", economy.GoodGrain, 42)
	gs.SetMarketBuyOrder("buyer", economy.GoodGrain, 18)
	clone := gs.CloneMarketOrders()
	if clone == nil || clone.SellOffers["seller"][economy.GoodGrain] != 42 || clone.BuyOrders["buyer"][economy.GoodGrain] != 18 {
		t.Fatalf("emir defteri kopyalanmadi: %+v", clone)
	}
	clone.SellOffers["seller"][economy.GoodGrain] = 1
	if got := gs.MarketOrders.SellOffers["seller"][economy.GoodGrain]; got != 42 {
		t.Fatalf("emir defteri kopyası kaynak state'i değiştirmemeli: got=%d", got)
	}
}

func TestOpenMarketSupplyExcludesReservedFactionStock(t *testing.T) {
	sellerID := faction.FactionID("seller")
	otherSellerID := faction.FactionID("other_seller")
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			sellerID:      {ID: sellerID, Grain: 9000},
			otherSellerID: {ID: otherSellerID, Grain: 500},
		},
	}
	gs.SetMarketSellOffer(sellerID, economy.GoodGrain, 0)
	gs.SetMarketSellOffer(otherSellerID, economy.GoodGrain, 25)

	if got := gs.OpenMarketSupplyByGood()[economy.GoodGrain]; got != 25 {
		t.Fatalf("pazar arzı toplam stok değil kalan satış emirleri olmalıydı: got=%d want=25", got)
	}
}
