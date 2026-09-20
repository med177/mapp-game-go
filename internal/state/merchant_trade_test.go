package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestMerchantTradePortEndpointsPreferCapitalPort(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"genoa": {ID: "genoa", CapitalSettlementID: "genoa_capital"},
		},
		Regions: map[world.RegionID]*world.Region{
			"genoa": {
				ID:        "genoa",
				OwnerID:   "genoa",
				WorldX:    0,
				WorldY:    0,
				Neighbors: []world.RegionID{"ligurian_sea"},
				Settlements: []world.Settlement{
					{ID: "genoa_capital", IsCenter: true},
				},
				Buildings: []string{"port"},
			},
			"midilli": {
				ID:        "midilli",
				OwnerID:   "genoa",
				WorldX:    100,
				WorldY:    100,
				Neighbors: []world.RegionID{"aegean_sea"},
				Buildings: []string{"port"},
			},
			"ligurian_sea": {ID: "ligurian_sea", IsSea: true},
			"aegean_sea":   {ID: "aegean_sea", IsSea: true},
		},
	}

	endpoints := gs.merchantTradePortEndpoints("genoa")
	if len(endpoints) != 1 || endpoints[0].regionID != "genoa" {
		t.Fatalf("başkent liman endpoint'leri = %+v, want genoa", endpoints)
	}
}

func TestMerchantTradePortEndpointsChooseNearestPortToCapital(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"genoa": {ID: "genoa", CapitalSettlementID: "capital_city"},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {
				ID:        "capital",
				OwnerID:   "genoa",
				WorldX:    0,
				WorldY:    0,
				Neighbors: []world.RegionID{"inland"},
				Settlements: []world.Settlement{
					{ID: "capital_city", IsCenter: true},
				},
			},
			"near_port": {
				ID:        "near_port",
				OwnerID:   "genoa",
				WorldX:    3,
				WorldY:    4,
				Neighbors: []world.RegionID{"near_sea"},
				Buildings: []string{"port"},
			},
			"far_port": {
				ID:        "far_port",
				OwnerID:   "genoa",
				WorldX:    10,
				WorldY:    0,
				Neighbors: []world.RegionID{"far_sea"},
				Buildings: []string{"port"},
			},
			"inland":   {ID: "inland"},
			"near_sea": {ID: "near_sea", IsSea: true},
			"far_sea":  {ID: "far_sea", IsSea: true},
		},
	}

	endpoints := gs.merchantTradePortEndpoints("genoa")
	if len(endpoints) != 1 || endpoints[0].regionID != "near_port" {
		t.Fatalf("en yakın liman endpoint'leri = %+v, want near_port", endpoints)
	}
}

func TestMerchantTradePortEndpointsFollowCapitalChange(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"genoa": {ID: "genoa", CapitalSettlementID: "capital_a"},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital_a": {
				ID:        "capital_a",
				OwnerID:   "genoa",
				Neighbors: []world.RegionID{"sea_a"},
				Settlements: []world.Settlement{
					{ID: "capital_a", IsCenter: true},
				},
				Buildings: []string{"port"},
			},
			"capital_b": {
				ID:        "capital_b",
				OwnerID:   "genoa",
				Neighbors: []world.RegionID{"sea_b"},
				Settlements: []world.Settlement{
					{ID: "capital_b", IsCenter: true},
				},
				Buildings: []string{"port"},
			},
			"sea_a": {ID: "sea_a", IsSea: true},
			"sea_b": {ID: "sea_b", IsSea: true},
		},
	}

	if got := gs.merchantTradePortEndpoints("genoa")[0].regionID; got != "capital_a" {
		t.Fatalf("ilk başkent limanı = %s, want capital_a", got)
	}
	if !gs.SetFactionCapital("genoa", "capital_b") {
		t.Fatal("başkent capital_b'ye taşınamadı")
	}
	if got := gs.merchantTradePortEndpoints("genoa")[0].regionID; got != "capital_b" {
		t.Fatalf("taşınan başkent limanı = %s, want capital_b", got)
	}
}

func TestMerchantTradeRouteUsesSelectedPortForFleetIncome(t *testing.T) {
	route := &economy.TradeRoute{FromFactionID: "genoa", ToFactionID: "partner", AmountPerTurn: 3}
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"genoa":   {ID: "genoa", CapitalSettlementID: "genoa_capital"},
			"partner": {ID: "partner", CapitalSettlementID: "partner_capital"},
		},
		Regions: map[world.RegionID]*world.Region{
			"genoa": {
				ID:          "genoa",
				OwnerID:     "genoa",
				Neighbors:   []world.RegionID{"genoa_sea"},
				Settlements: []world.Settlement{{ID: "genoa_capital", IsCenter: true}},
				Buildings:   []string{"port"},
			},
			"midilli": {
				ID:        "midilli",
				OwnerID:   "genoa",
				Neighbors: []world.RegionID{"old_trade_sea"},
				Buildings: []string{"port"},
			},
			"partner_port": {
				ID:          "partner_port",
				OwnerID:     "partner",
				Neighbors:   []world.RegionID{"partner_sea"},
				Settlements: []world.Settlement{{ID: "partner_capital", IsCenter: true}},
				Buildings:   []string{"port"},
			},
			"genoa_sea":     {ID: "genoa_sea", IsSea: true, Neighbors: []world.RegionID{"partner_sea"}},
			"partner_sea":   {ID: "partner_sea", IsSea: true, Neighbors: []world.RegionID{"genoa_sea", "partner_port"}},
			"old_trade_sea": {ID: "old_trade_sea", IsSea: true, Neighbors: []world.RegionID{"midilli"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"merchant": merchantTestFleet("merchant", "genoa", "old_trade_sea", route.AssignmentKey(), 1),
		},
		UnitTypes:   map[string]*army.UnitType{"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade}},
		TradeRoutes: []*economy.TradeRoute{route},
	}

	pairs := gs.MerchantTradeRoutePortPairs(route)
	if len(pairs) != 1 || pairs[0].FromRegionID != "genoa" || pairs[0].ToRegionID != "partner_port" {
		t.Fatalf("canonical rota limanları = %+v, want genoa -> partner_port", pairs)
	}
	if gs.MerchantFleetSupportsTradeRoute(gs.Armies["merchant"], route) {
		t.Fatal("eski Midilli denizindeki filo yeni Cenova rotasında gelir üretmemeli")
	}
	gs.Armies["merchant"].RegionID = pairs[0].ToSeaID
	if !gs.MerchantFleetSupportsTradeRoute(gs.Armies["merchant"], route) {
		t.Fatal("filo canonical hedef denize ulaştığında rotayı desteklemeli")
	}
}

func TestMerchantTradePortEndpointFollowsPortFacingSea(t *testing.T) {
	route := &economy.TradeRoute{FromFactionID: "athena_duk", ToFactionID: "partner", AmountPerTurn: 3}
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"athena_duk": {ID: "athena_duk", CapitalSettlementID: "athens"},
			"partner":    {ID: "partner", CapitalSettlementID: "partner_capital"},
		},
		Regions: map[world.RegionID]*world.Region{
			"greece": {
				ID:        "greece",
				OwnerID:   "athena_duk",
				WorldX:    953,
				WorldY:    515,
				Neighbors: []world.RegionID{"cretan_sea", "mediterranean_open_6", "otranto_strait"},
				Settlements: []world.Settlement{
					{ID: "athens", X: 953, Y: 514, Type: world.SettlementCity, IsCenter: true},
					{ID: "greece_piraeus", X: 953, Y: 518, Type: world.SettlementPort},
				},
				Buildings: []string{"port"},
			},
			"partner_port": {
				ID:          "partner_port",
				OwnerID:     "partner",
				Neighbors:   []world.RegionID{"partner_sea"},
				Settlements: []world.Settlement{{ID: "partner_capital", Type: world.SettlementCity, IsCenter: true}},
				Buildings:   []string{"port"},
			},
			"cretan_sea":           {ID: "cretan_sea", IsSea: true, WorldX: 971, WorldY: 547, Neighbors: []world.RegionID{"partner_sea"}},
			"mediterranean_open_6": {ID: "mediterranean_open_6", IsSea: true, WorldX: 948, WorldY: 486},
			"otranto_strait":       {ID: "otranto_strait", IsSea: true, WorldX: 863, WorldY: 488},
			"partner_sea":          {ID: "partner_sea", IsSea: true, Neighbors: []world.RegionID{"cretan_sea", "partner_port"}},
		},
		TradeRoutes: []*economy.TradeRoute{route},
	}

	if got := gs.MerchantTradePortSettlementID("athena_duk"); got != "greece_piraeus" {
		t.Fatalf("ana ticaret port settlement = %q, want greece_piraeus", got)
	}
	if got := gs.merchantTradePortEndpoints("athena_duk"); len(got) != 1 || got[0].seaID != "cretan_sea" {
		t.Fatalf("Atina ana port deniz endpoint'i = %+v, want cretan_sea", got)
	}
	pairs := gs.MerchantTradeRoutePortPairs(route)
	if len(pairs) != 1 || pairs[0].FromSeaID != "cretan_sea" {
		t.Fatalf("Atina rota endpoint'i = %+v, want cretan_sea", pairs)
	}
}

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

func TestMerchantTradeIncomeUsesCenterCapacityAndNavalSafety(t *testing.T) {
	const routeKey = "from->to"
	route := &economy.TradeRoute{
		FromFactionID: "from",
		ToFactionID:   "to",
		AmountPerTurn: 4,
		Good:          economy.GoodSpice,
	}
	gs := &GameState{
		Year: 1300,
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
		TradeCenters: world.TradeCenterConfig{
			Centers: []world.TradeCenterDef{
				{ID: "from_port", Tier: world.TradeCenterPrimary},
				{ID: "to_port", Tier: world.TradeCenterSecondary},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"merchant": merchantTestFleet("merchant", "from", "sea_to", routeKey, 2),
		},
		UnitTypes: map[string]*army.UnitType{
			"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade, MerchantTradeIncome: 8},
			"warship":       {ID: "warship", Category: army.CategoryNavalWar},
		},
		TradeRoutes: []*economy.TradeRoute{route},
	}

	if got, want := gs.MerchantBonusCapacity(route), 7; got != want {
		t.Fatalf("merchant capacity = %d, want %d", got, want)
	}
	if got, want := gs.MerchantTradeIncomeForRoute(route, 6), 15; got != want {
		t.Fatalf("unescorted merchant income = %d, want %d", got, want)
	}

	gs.Armies["patrol"] = &army.Army{
		ID:       "patrol",
		OwnerID:  "from",
		RegionID: "sea_to",
		IsNaval:  true,
		Units:    []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}},
		NavalMission: &army.NavalMission{
			Kind:           army.NavalMissionPatrol,
			TargetRegionID: "sea_to",
		},
	}
	if got, want := gs.MerchantTradeIncomeForRoute(route, 6), 20; got != want {
		t.Fatalf("escorted merchant income = %d, want %d", got, want)
	}
}

func TestMerchantTradeIncomeRequiresMerchantCargo(t *testing.T) {
	route := &economy.TradeRoute{FromFactionID: "from", ToFactionID: "to", AmountPerTurn: 4}
	gs := &GameState{
		UnitTypes: map[string]*army.UnitType{
			"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade, MerchantTradeIncome: 8},
		},
		TradeRoutes: []*economy.TradeRoute{route},
	}
	if got := gs.MerchantTradeIncomeForRoute(route, 4); got != 0 {
		t.Fatalf("base route volume produced merchant income %d, want 0", got)
	}
}

func TestMergeMerchantTradeFleetAtRouteKeepsExistingFleetAndHonorsCapacity(t *testing.T) {
	const routeKey = "from->to"
	route := &economy.TradeRoute{FromFactionID: "from", ToFactionID: "to", AmountPerTurn: 3}
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"from_port": {ID: "from_port", OwnerID: "from", Neighbors: []world.RegionID{"sea_from"}, Settlements: []world.Settlement{{ID: "from_harbor", Type: world.SettlementPort}}},
			"to_port":   {ID: "to_port", OwnerID: "to", Neighbors: []world.RegionID{"sea_to"}, Settlements: []world.Settlement{{ID: "to_harbor", Type: world.SettlementPort}}},
			"sea_from":  {ID: "sea_from", IsSea: true, Neighbors: []world.RegionID{"from_port", "sea_to"}},
			"sea_to":    {ID: "sea_to", IsSea: true, Neighbors: []world.RegionID{"sea_from", "to_port"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_old": merchantTestFleet("fleet_old", "from", "sea_to", routeKey, 18),
			"fleet_new": merchantTestFleet("fleet_new", "from", "sea_to", routeKey, 5),
		},
		UnitTypes:   map[string]*army.UnitType{"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade}},
		TradeRoutes: []*economy.TradeRoute{route},
	}

	survivorID, merged := gs.MergeMerchantTradeFleetAtRoute("fleet_new")
	if !merged || survivorID != "fleet_new" {
		t.Fatalf("kısmi birleşme = (%q, %v), want (fleet_new, true)", survivorID, merged)
	}
	if got := len(gs.Armies["fleet_old"].Units); got != army.MaxArmySize {
		t.Fatalf("önceki filo gemi sayısı = %d, want %d", got, army.MaxArmySize)
	}
	if got := len(gs.Armies["fleet_new"].Units); got != 3 {
		t.Fatalf("artan yeni filo gemi sayısı = %d, want 3", got)
	}

	if removed := gs.MergeMerchantTradeFleets(); removed != 0 {
		t.Fatalf("dolu hedeften sonra silinen filo = %d, want 0", removed)
	}
}

func TestMergeMerchantTradeFleetsConsolidatesThreeTurnEndFleets(t *testing.T) {
	const routeKey = "from->to"
	route := &economy.TradeRoute{FromFactionID: "from", ToFactionID: "to", AmountPerTurn: 3}
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"from_port": {ID: "from_port", OwnerID: "from", Neighbors: []world.RegionID{"sea_from"}, Settlements: []world.Settlement{{ID: "from_harbor", Type: world.SettlementPort}}},
			"to_port":   {ID: "to_port", OwnerID: "to", Neighbors: []world.RegionID{"sea_to"}, Settlements: []world.Settlement{{ID: "to_harbor", Type: world.SettlementPort}}},
			"sea_from":  {ID: "sea_from", IsSea: true, Neighbors: []world.RegionID{"from_port", "sea_to"}},
			"sea_to":    {ID: "sea_to", IsSea: true, Neighbors: []world.RegionID{"sea_from", "to_port"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_1": merchantTestFleet("fleet_1", "from", "sea_to", routeKey, 2),
			"fleet_2": merchantTestFleet("fleet_2", "from", "sea_to", routeKey, 2),
			"fleet_3": merchantTestFleet("fleet_3", "from", "sea_to", routeKey, 2),
		},
		UnitTypes:   map[string]*army.UnitType{"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade}},
		TradeRoutes: []*economy.TradeRoute{route},
	}

	if removed := gs.MergeMerchantTradeFleets(); removed != 2 {
		t.Fatalf("tur sonu birleşmesinde silinen filo = %d, want 2", removed)
	}
	if got := len(gs.Armies); got != 1 {
		t.Fatalf("tur sonu filo sayısı = %d, want 1", got)
	}
	if got := len(gs.Armies["fleet_1"].Units); got != 6 {
		t.Fatalf("birleşmiş filodaki gemi sayısı = %d, want 6", got)
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
