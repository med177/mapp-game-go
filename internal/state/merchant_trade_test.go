package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/world"
)

func TestMerchantFleetTradeStatusesMatchesIndividualEvaluation(t *testing.T) {
	const routeKey = "from->to"
	route := &economy.TradeRoute{
		FromFactionID: "from",
		ToFactionID:   "to",
		AmountPerTurn: 3,
	}
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"from_port": {
				ID:          "from_port",
				OwnerID:     "from",
				Neighbors:   []world.RegionID{"sea_from"},
				Settlements: []world.Settlement{{ID: "from_harbor", Type: world.SettlementPort}},
			},
			"to_port": {
				ID:          "to_port",
				OwnerID:     "to",
				Neighbors:   []world.RegionID{"sea_to"},
				Settlements: []world.Settlement{{ID: "to_harbor", Type: world.SettlementPort}},
			},
			"sea_from": {ID: "sea_from", IsSea: true, Neighbors: []world.RegionID{"from_port", "sea_to"}},
			"sea_to":   {ID: "sea_to", IsSea: true, Neighbors: []world.RegionID{"sea_from", "to_port"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"active_1":    merchantTestFleet("active_1", "from", "sea_to", routeKey, 2),
			"active_2":    merchantTestFleet("active_2", "from", "sea_to", routeKey, 2),
			"pending":     merchantTestFleet("pending", "from", "sea_from", routeKey, 1),
			"wrong_owner": merchantTestFleet("wrong_owner", "other", "sea_to", routeKey, 1),
		},
		UnitTypes: map[string]*army.UnitType{
			"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade},
		},
		TradeRoutes: []*economy.TradeRoute{route},
	}

	statuses := gs.MerchantFleetTradeStatuses(nil)
	if len(statuses) != len(gs.Armies) {
		t.Fatalf("status count = %d, want %d", len(statuses), len(gs.Armies))
	}
	for fleetID, fleet := range gs.Armies {
		got := statuses[fleetID]
		wantBonus := gs.MerchantFleetTradeRouteBonus(fleet, route)
		wantPending := !gs.MerchantFleetSupportsTradeRoute(fleet, route)
		if got.Bonus != wantBonus || got.Pending != wantPending {
			t.Errorf("fleet %s status = %+v, want bonus=%d pending=%v", fleetID, got, wantBonus, wantPending)
		}
	}

	reused := gs.MerchantFleetTradeStatuses(map[army.ArmyID]MerchantFleetTradeStatus{
		"stale": {Bonus: 99},
	})
	if _, ok := reused["stale"]; ok {
		t.Error("reused status map retained stale entry")
	}
}

func merchantTestFleet(id army.ArmyID, owner string, region world.RegionID, routeKey string, ships int) *army.Army {
	units := make([]army.Unit, ships)
	for i := range units {
		units[i] = army.Unit{TypeID: "merchant_ship", CurrentHP: army.MaxUnitHP}
	}
	return &army.Army{
		ID:            id,
		OwnerID:       owner,
		RegionID:      region,
		Units:         units,
		IsNaval:       true,
		TradeRouteKey: routeKey,
	}
}
