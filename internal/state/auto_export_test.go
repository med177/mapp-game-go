package state

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestApplyAutomaticExportsSellsConfiguredPercentOfSurplus(t *testing.T) {
	playerID := faction.FactionID("player")
	buyerID := faction.FactionID("buyer")
	gs := &GameState{
		PlayerFactionID: playerID,
		Factions: map[faction.FactionID]*faction.Faction{
			playerID: {ID: playerID, Iron: 500},
			buyerID:  {ID: buyerID, Gold: 1000},
		},
		MarketPrices: economy.CurrentMarketPrice{economy.GoodIron: 5},
		MarketOrders: MarketOrderBook{
			BuyOrders: map[faction.FactionID]map[economy.GoodType]int{
				buyerID: {economy.GoodIron: 200},
			},
		},
		AutoExportPolicies: map[economy.GoodType]AutoExportPolicy{
			economy.GoodIron: {Enabled: true, Percent: 20},
		},
	}

	results := gs.ApplyAutomaticExports()
	if got := results[economy.GoodIron].Sold; got != 96 {
		t.Fatalf("surplusun yüzde 20si satılmalıydı, got=%d", got)
	}
	if got := gs.Factions[playerID].Iron; got != 404 {
		t.Fatalf("oyuncu rezervi korunarak stok 404 olmalıydı, got=%d", got)
	}
	if got := gs.Factions[buyerID].Iron; got != 96 {
		t.Fatalf("alıcıya 96 demir aktarılmalıydı, got=%d", got)
	}
	if got := gs.MarketBuyOrder(buyerID, economy.GoodIron, 5); got != 104 {
		t.Fatalf("alıcı talebi 104 azalmalıydı, got=%d", got)
	}
}

func TestAutoExportReserveUsesTwoTurnsOfProductionForOtherGoods(t *testing.T) {
	playerID := faction.FactionID("player")
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			playerID: {ID: playerID, Iron: 500},
		},
		Regions: map[world.RegionID]*world.Region{
			"mine": {ID: "mine", OwnerID: string(playerID), BaseIronOutput: 100},
		},
	}

	if got := gs.AutoExportReserve(playerID, economy.GoodIron); got != 200 {
		t.Fatalf("demir rezervi iki tur uretim olan 200 olmaliydi, got=%d", got)
	}
	if got := gs.AutoExportSurplus(playerID, economy.GoodIron); got != 300 {
		t.Fatalf("demir fazlasi rezerv sonrasi 300 olmaliydi, got=%d", got)
	}
}
