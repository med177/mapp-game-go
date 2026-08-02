package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestTradeBonusFleetVisualsFilterToActiveAssignedFleets(t *testing.T) {
	route := &economy.TradeRoute{FromFactionID: "from", ToFactionID: "to", GoldPerUnit: 5}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"from_port": {ID: "from_port", OwnerID: "from", Neighbors: []world.RegionID{"aegean"}},
			"to_port":   {ID: "to_port", OwnerID: "to", Neighbors: []world.RegionID{"marmara"}},
			"aegean":    {ID: "aegean", IsSea: true},
			"marmara":   {ID: "marmara", IsSea: true},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{
			{ID: "from_port", Links: []world.RegionID{"to_port"}},
			{ID: "to_port", Links: []world.RegionID{"from_port"}},
		}},
		TradeRoutes: []*economy.TradeRoute{route},
		Armies: map[army.ArmyID]*army.Army{
			"active": {
				ID: "active", OwnerID: "from", IsNaval: true, RegionID: "marmara",
				TradeRouteKey: route.AssignmentKey(), Units: []army.Unit{{TypeID: "merchant_ship"}},
			},
			"away": {
				ID: "away", OwnerID: "from", IsNaval: true, RegionID: "aegean",
				TradeRouteKey: route.AssignmentKey(), Units: []army.Unit{{TypeID: "merchant_ship"}},
			},
		},
	}
	r := &Renderer{
		gs:       gs,
		camScale: 1,
		worldMap: &WorldMap{regionAnchor: map[world.RegionID][2]int{"marmara": {100, 100}, "aegean": {40, 100}}},
	}

	visuals := r.tradeBonusFleetVisuals()
	if len(visuals) != 1 || visuals[0].fleet.ID != "active" || visuals[0].bonus != 1 {
		t.Fatalf("yalnız aktif hedef denizdeki merchant filosu görünmeli: %+v", visuals)
	}

	r.mapMode = MapModeTrade
	r.tradeCorridors = []tradeCorridorInfo{{
		sx: 0, sy: 0, cx: 50, cy: 60, dx: 100, dy: 0,
		routeKeys: []string{route.AssignmentKey()},
	}}
	badge := merchantTradeBonusBadgeRect(visuals[0].position.X, visuals[0].position.Y)
	if aid, ok := r.merchantTradeBonusHitAt(badge.X+badge.W/2, badge.Y+badge.H/2); !ok || aid != "active" {
		t.Fatalf("ticaret haritasındaki merchant rozeti hover hit-test'i aktif filoyu bulmalı: aid=%q hit=%t", aid, ok)
	}
}

func TestTradeBonusFleetConnectsToItsAssignedCorridor(t *testing.T) {
	routeKey := "from->to"
	r := &Renderer{
		tradeCorridors: []tradeCorridorInfo{
			{
				sx: 0, sy: 0, cx: 50, cy: 60, dx: 100, dy: 0,
				routeKeys: []string{routeKey},
			},
			{
				sx: 200, sy: 0, cx: 250, cy: 60, dx: 300, dy: 0,
				routeKeys: []string{"other->route"},
			},
		},
	}

	x, y, ok := r.tradeRouteConnectionPoint(routeKey, 50, 100)
	if !ok {
		t.Fatal("atanan merchant rotası için bağlantı noktası bulunmalı")
	}
	if x < 0 || x > 100 || y < 0 || y > 60 {
		t.Fatalf("bağlantı noktası yanlış koridorda: (%.2f, %.2f)", x, y)
	}
	if _, _, ok := r.tradeRouteConnectionPoint("missing", 50, 100); ok {
		t.Fatal("atanmamış rota için bağlantı noktası üretilmemeli")
	}
}
