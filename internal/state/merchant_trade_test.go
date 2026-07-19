package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func merchantTradeTestState() (*GameState, *economy.TradeRoute) {
	route := &economy.TradeRoute{
		FromFactionID: "venice", ToFactionID: "mamluk", Good: economy.GoodSpice,
		AmountPerTurn: 2, GoldPerUnit: 5,
	}
	return &GameState{
		Year: 1300,
		Factions: map[faction.FactionID]*faction.Faction{
			"venice": {ID: "venice", Spice: 10},
			"mamluk": {ID: "mamluk", Gold: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"venice":   {ID: "venice", OwnerID: "venice", Neighbors: []world.RegionID{"adriatic"}, TradeCapacity: 9},
			"egypt":    {ID: "egypt", OwnerID: "mamluk", Neighbors: []world.RegionID{"med"}, TradeCapacity: 6},
			"adriatic": {ID: "adriatic", IsSea: true, Neighbors: []world.RegionID{"venice", "med"}},
			"med":      {ID: "med", IsSea: true, Neighbors: []world.RegionID{"adriatic", "egypt"}},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{
			{ID: "venice", Links: []world.RegionID{"egypt"}, Tier: world.TradeCenterPrimary},
			{ID: "egypt", Links: []world.RegionID{"venice"}, Tier: world.TradeCenterSecondary},
		}},
		TradeRoutes: []*economy.TradeRoute{route},
		UnitTypes: map[string]*army.UnitType{
			"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade},
		},
		Armies: map[army.ArmyID]*army.Army{
			"merchant": {
				ID: "merchant", OwnerID: "venice", RegionID: "adriatic", IsNaval: true,
				TradeRouteKey: route.AssignmentKey(),
				Units: []army.Unit{
					{TypeID: "merchant_ship", CurrentHP: 100},
					{TypeID: "merchant_ship", CurrentHP: 100},
					{TypeID: "merchant_ship", CurrentHP: 100},
				},
			},
		},
	}, route
}

func TestMerchantTradeBonusUsesAssignmentLocationAndRouteCap(t *testing.T) {
	gs, route := merchantTradeTestState()
	gs.RefreshMerchantTradeBonuses()
	if route.MerchantAmountBonus != 2 || route.EffectiveAmountPerTurn() != 4 {
		t.Fatalf("üç merchant gemisi rota başına +2 sınırında kalmalıydı: %+v", route)
	}
	logs := economy.ApplyTradeRoutes(gs.Factions, gs.TradeRoutes)
	if len(logs) != 0 || gs.Factions["venice"].Spice != 6 || gs.Factions["venice"].Gold != 20 || gs.Factions["mamluk"].Spice != 4 || gs.Factions["mamluk"].Gold != 80 {
		t.Fatalf("merchant hacmi gerçek mal ve altın transferi üretmeliydi: venice=%+v mamluk=%+v logs=%v", gs.Factions["venice"], gs.Factions["mamluk"], logs)
	}
}

func TestMerchantTradeBonusRequiresActiveConnectedCenterSea(t *testing.T) {
	gs, route := merchantTradeTestState()
	gs.Armies["merchant"].RegionID = "unknown_sea"
	route.MerchantAmountBonus = 2
	gs.RefreshMerchantTradeBonuses()
	if route.MerchantAmountBonus != 0 {
		t.Fatalf("ticaret merkezi denizinde olmayan filo bonus vermemeliydi: %d", route.MerchantAmountBonus)
	}

	gs.Armies["merchant"].RegionID = "adriatic"
	gs.TradeCenters.Centers[0].Links = nil
	gs.TradeCenters.Centers[1].Links = nil
	gs.RefreshMerchantTradeBonuses()
	if route.MerchantAmountBonus != 0 {
		t.Fatalf("tarihsel merkez bağlantısı kopuk rota bonus vermemeliydi: %d", route.MerchantAmountBonus)
	}
}

func TestSuspendedTradeRouteTransfersNeitherBaseNorMerchantVolume(t *testing.T) {
	gs, route := merchantTradeTestState()
	route.SuspendedTurns = 2
	route.MerchantAmountBonus = 2
	gs.RefreshMerchantTradeBonuses()
	economy.ApplyTradeRoutes(gs.Factions, gs.TradeRoutes)
	if route.MerchantAmountBonus != 0 || gs.Factions["venice"].Spice != 10 || gs.Factions["mamluk"].Gold != 100 {
		t.Fatalf("askıdaki rota tamamen hareketsiz kalmalıydı: route=%+v", route)
	}
}
