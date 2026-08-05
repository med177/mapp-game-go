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
		AmountPerTurn: 1, GoldPerUnit: 5,
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
				ID: "merchant", OwnerID: "venice", RegionID: "med", IsNaval: true,
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
	if got := gs.MerchantFleetTradeRouteBonus(gs.Armies["merchant"], route); got != 3 {
		t.Fatalf("uygun konumdaki üç merchant gemisi gemi başına +1 bonus vermeliydi, got=%d", got)
	}
	gs.RefreshMerchantTradeBonuses()
	if route.MerchantAmountBonus != 3 || route.EffectiveAmountPerTurn() != 4 {
		t.Fatalf("üç merchant gemisi rota hacmini 4/tur sınırına kadar doldurmalıydı: %+v", route)
	}
	logs := economy.ApplyTradeRoutes(gs.Factions, gs.TradeRoutes)
	if len(logs) != 0 || gs.Factions["venice"].Spice != 6 || gs.Factions["venice"].Gold != 20 || gs.Factions["mamluk"].Spice != 4 || gs.Factions["mamluk"].Gold != 80 {
		t.Fatalf("merchant hacmi gerçek mal ve altın transferi üretmeliydi: venice=%+v mamluk=%+v logs=%v", gs.Factions["venice"], gs.Factions["mamluk"], logs)
	}
}

func TestMerchantTradeRoutesForFleetFiltersByOwnerAndActiveCenter(t *testing.T) {
	gs, route := merchantTradeTestState()
	gs.TradeRoutes = append(gs.TradeRoutes,
		&economy.TradeRoute{FromFactionID: "mamluk", ToFactionID: "venice", Good: economy.GoodSpice, AmountPerTurn: 1, GoldPerUnit: 6},
		&economy.TradeRoute{FromFactionID: "genoa", ToFactionID: "mamluk", Good: economy.GoodCloth, AmountPerTurn: 1, GoldPerUnit: 6},
		&economy.TradeRoute{FromFactionID: "venice", ToFactionID: "mamluk", Good: economy.GoodIron, AmountPerTurn: 1, GoldPerUnit: 9, SuspendedTurns: 1},
	)
	gs.Factions["genoa"] = &faction.Faction{ID: "genoa"}

	routes := gs.MerchantTradeRoutesForFleet(gs.Armies["merchant"])
	if len(routes) != 1 || routes[0] != route {
		t.Fatalf("filo yalnız kendi aktif merkez rotasını görmeliydi: %+v", routes)
	}
	if !gs.SetMerchantTradeRoute("merchant", route.AssignmentKey()) || gs.Armies["merchant"].TradeRouteKey != route.AssignmentKey() {
		t.Fatalf("geçerli rota merchant filosuna atanmalıydı")
	}
	gs.Armies["merchant"].TradeRouteKey = gs.TradeRoutes[1].AssignmentKey()
	if gs.MerchantFleetSupportsTradeRoute(gs.Armies["merchant"], gs.TradeRoutes[1]) {
		t.Fatalf("merchant filosu karşı tarafın ihracat rotasını desteklememeli")
	}
	gs.Armies["merchant"].TradeRouteKey = route.AssignmentKey()
	if gs.SetMerchantTradeRoute("merchant", "genoa->mamluk") {
		t.Fatalf("başka fraksiyonun rotası merchant filosuna atanamamalıydı")
	}
	if !gs.SetMerchantTradeRoute("merchant", "") || gs.Armies["merchant"].TradeRouteKey != "" {
		t.Fatalf("boş rota anahtarı görevi kaldırmalıydı")
	}
}

func TestMerchantTradeRouteUsesConnectedOwnedPortsWithoutHistoricalCenters(t *testing.T) {
	route := &economy.TradeRoute{
		FromFactionID: "ottoman", ToFactionID: "golden_horde",
		Good: economy.GoodGrain, AmountPerTurn: 3, GoldPerUnit: 2,
	}
	gs := &GameState{
		Year: 1303,
		Regions: map[world.RegionID]*world.Region{
			"bursa":          {ID: "bursa", OwnerID: "ottoman", Buildings: []string{"port"}, Neighbors: []world.RegionID{"marmara"}},
			"dobruja":        {ID: "dobruja", OwnerID: "golden_horde", Buildings: []string{"port"}, Neighbors: []world.RegionID{"black_sea_port"}},
			"marmara":        {ID: "marmara", IsSea: true, Neighbors: []world.RegionID{"bursa", "black_sea"}},
			"black_sea":      {ID: "black_sea", IsSea: true, Neighbors: []world.RegionID{"marmara", "black_sea_port"}},
			"black_sea_port": {ID: "black_sea_port", IsSea: true, Neighbors: []world.RegionID{"black_sea", "dobruja"}},
		},
		TradeRoutes: []*economy.TradeRoute{route},
		UnitTypes: map[string]*army.UnitType{
			"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade},
		},
		Armies: map[army.ArmyID]*army.Army{
			"merchant": {
				ID: "merchant", OwnerID: "ottoman", RegionID: "marmara", IsNaval: true,
				Units: []army.Unit{{TypeID: "merchant_ship"}},
			},
		},
	}

	pairs := gs.MerchantTradeRoutePortPairs(route)
	if len(pairs) != 1 || pairs[0].FromRegionID != "bursa" || pairs[0].ToRegionID != "dobruja" {
		t.Fatalf("gerçek limanlar arasındaki rota çifti bulunmalıydı: %+v", pairs)
	}
	got := gs.MerchantTradeRouteSeaRegions(route)
	if len(got) != 1 || got[0] != "black_sea_port" {
		t.Fatalf("liman çifti rotanın hedef denizini döndürmeliydi: %+v", got)
	}
	if target, ok := gs.MerchantTradeRouteTargetSeaRegion(route); !ok || target != "black_sea_port" {
		t.Fatalf("kanonik hedef deniz yanlış: target=%s ok=%v", target, ok)
	}
	routes := gs.MerchantTradeRoutesForFleet(gs.Armies["merchant"])
	if len(routes) != 1 || routes[0] != route {
		t.Fatalf("tarihsel merkez olmadan aktif liman rotası merchant panelinde görünmeliydi: %+v", routes)
	}
}

func TestMerchantTradeRouteUsesDestinationPortSeaForGemlikOzi(t *testing.T) {
	route := &economy.TradeRoute{
		FromFactionID: "ottoman", ToFactionID: "golden_horde",
		Good: economy.GoodGrain, AmountPerTurn: 3, GoldPerUnit: 2,
	}
	gs := &GameState{
		Year: 1300,
		Regions: map[world.RegionID]*world.Region{
			"bursa": {
				ID: "bursa", OwnerID: "ottoman", Neighbors: []world.RegionID{"marmara"},
				Settlements: []world.Settlement{{ID: "gemlik", NameTR: "Gemlik", Type: world.SettlementPort}},
			},
			"lower_ukrainian_steppe": {
				ID: "lower_ukrainian_steppe", OwnerID: "golden_horde", Neighbors: []world.RegionID{"black_open_4"},
				Settlements: []world.Settlement{{ID: "ozi", NameTR: "Özi", Type: world.SettlementPort}},
			},
			"marmara":      {ID: "marmara", IsSea: true, Neighbors: []world.RegionID{"bursa", "black_sea"}},
			"black_sea":    {ID: "black_sea", IsSea: true, Neighbors: []world.RegionID{"marmara", "black_open_4"}},
			"black_open_4": {ID: "black_open_4", NameTR: "Karadeniz Açık 4", IsSea: true, Neighbors: []world.RegionID{"black_sea", "lower_ukrainian_steppe"}},
		},
		TradeRoutes: []*economy.TradeRoute{route},
		UnitTypes: map[string]*army.UnitType{
			"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade},
		},
		Armies: map[army.ArmyID]*army.Army{
			"merchant": {
				ID: "merchant", OwnerID: "ottoman", RegionID: "black_open_4", IsNaval: true,
				TradeRouteKey: route.AssignmentKey(), Units: []army.Unit{{TypeID: "merchant_ship"}},
			},
		},
	}

	pairs := gs.MerchantTradeRoutePortPairs(route)
	if len(pairs) != 1 || pairs[0].FromRegionID != "bursa" || pairs[0].ToRegionID != "lower_ukrainian_steppe" || pairs[0].ToSeaID != "black_open_4" {
		t.Fatalf("Gemlik-Özi liman çifti yanlış: %+v", pairs)
	}
	if target, ok := gs.MerchantTradeRouteTargetSeaRegion(route); !ok || target != "black_open_4" {
		t.Fatalf("Gemlik-Özi rotası Karadeniz Açık 4'ü hedeflemeli: target=%s ok=%v", target, ok)
	}
	if got := gs.MerchantFleetTradeRouteBonus(gs.Armies["merchant"], route); got != 1 {
		t.Fatalf("Özi denizindeki merchant filosu bonus vermeli: got=%d", got)
	}
	gs.Armies["merchant"].RegionID = "marmara"
	if got := gs.MerchantFleetTradeRouteBonus(gs.Armies["merchant"], route); got != 0 {
		t.Fatalf("Gemlik tarafındaki deniz tek hedef kuralında bonus vermemeli: got=%d", got)
	}
	reverse := &economy.TradeRoute{FromFactionID: "golden_horde", ToFactionID: "ottoman"}
	if target, ok := gs.MerchantTradeRouteTargetSeaRegion(reverse); !ok || target != "marmara" {
		t.Fatalf("Özi -> Gemlik yönü Marmara denizini hedeflemeli: target=%s ok=%v", target, ok)
	}
}

func TestTradeRouteBlockadeReducesMerchantVolume(t *testing.T) {
	gs, route := merchantTradeTestState()
	gs.Factions["genoa"] = &faction.Faction{ID: "genoa"}
	gs.Relations = map[string]*faction.Relation{
		faction.RelationKey("genoa", "venice"): {
			FactionA: "genoa", FactionB: "venice", Stance: faction.StanceWar,
		},
	}
	gs.UnitTypes["warship"] = &army.UnitType{ID: "warship", Category: army.CategoryNavalWar}
	gs.Armies["blockader"] = &army.Army{
		ID: "blockader", OwnerID: "genoa", RegionID: "med", IsNaval: true,
		NavalMission: &army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "med"},
		Units:        []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}},
	}

	gs.RefreshTradeRouteBlockades()
	gs.RefreshMerchantTradeBonuses()
	if route.BlockadePercent != 50 || route.EffectiveAmountPerTurn() != 2 {
		t.Fatalf("tek savaş gemisi rota hacmini yarıya indirmeliydi: route=%+v", route)
	}

	economy.ApplyTradeRoutes(gs.Factions, gs.TradeRoutes)
	if gs.Factions["venice"].Spice != 8 || gs.Factions["venice"].Gold != 10 || gs.Factions["mamluk"].Spice != 2 || gs.Factions["mamluk"].Gold != 90 {
		t.Fatalf("abluka sonrası yarım rota hacmi uygulanmalıydı: venice=%+v mamluk=%+v", gs.Factions["venice"], gs.Factions["mamluk"])
	}

	gs.Armies["blockader"].Units = append(gs.Armies["blockader"].Units, army.Unit{TypeID: "warship", CurrentHP: army.MaxUnitHP})
	route.BlockadePercent = 0
	gs.RefreshTradeRouteBlockades()
	if route.BlockadePercent != 100 || route.EffectiveAmountPerTurn() != 0 {
		t.Fatalf("iki savaş gemisi rotayı tamamen kapatmalıydı: route=%+v", route)
	}
}

func TestPatrolCountersEnemyBlockadeForTradeRoute(t *testing.T) {
	gs, route := merchantTradeTestState()
	gs.Factions["genoa"] = &faction.Faction{ID: "genoa"}
	gs.Relations = map[string]*faction.Relation{
		faction.RelationKey("genoa", "venice"): {
			FactionA: "genoa", FactionB: "venice", Stance: faction.StanceWar,
		},
	}
	gs.UnitTypes["warship"] = &army.UnitType{ID: "warship", Category: army.CategoryNavalWar}
	gs.Armies["blockader"] = &army.Army{
		ID: "blockader", OwnerID: "genoa", RegionID: "med", IsNaval: true,
		NavalMission: &army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "med"},
		Units:        []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}},
	}
	gs.RefreshTradeRouteBlockades()
	if route.BlockadePercent != 50 {
		t.Fatalf("devriye eklenmeden düşman abluka etkisi %%50 olmalıydı: %d", route.BlockadePercent)
	}
	gs.Armies["patrol"] = &army.Army{
		ID: "patrol", OwnerID: "venice", RegionID: "med", IsNaval: true,
		NavalMission: &army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "med"},
		Units:        []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}},
	}
	gs.RefreshTradeRouteBlockades()
	if route.BlockadePercent != 0 {
		t.Fatalf("aynı denizdeki devriye bir düşman gemisinin ablukasını dengelemeliydi: %d", route.BlockadePercent)
	}
}

func TestPatrolAndBlockadeMissionsHaveDifferentCommerceEffects(t *testing.T) {
	gs, route := merchantTradeTestState()
	gs.Factions["genoa"] = &faction.Faction{ID: "genoa"}
	gs.Relations = map[string]*faction.Relation{
		faction.RelationKey("genoa", "venice"): {
			FactionA: "genoa", FactionB: "venice", Stance: faction.StanceWar,
		},
	}
	gs.UnitTypes["warship"] = &army.UnitType{ID: "warship", Category: army.CategoryNavalWar}
	fleet := &army.Army{
		ID: "fleet", OwnerID: "genoa", RegionID: "med", IsNaval: true,
		Units: []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}},
	}
	gs.Armies["fleet"] = fleet
	fleet.NavalMission = &army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "med"}
	gs.RefreshTradeRouteBlockades()
	if route.BlockadePercent != 0 {
		t.Fatalf("devriye görevi ticaret ablukası oluşturmamalıydı: %d", route.BlockadePercent)
	}
	fleet.NavalMission = &army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "med"}
	gs.RefreshTradeRouteBlockades()
	if route.BlockadePercent != 50 {
		t.Fatalf("abluka görevi ticaret kesintisi oluşturmalıydı: %d", route.BlockadePercent)
	}
}

func TestRegionBlockadePercentUsesHostileWarshipsAtPort(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"port": {ID: "port", OwnerID: "venice", Neighbors: []world.RegionID{"sea"}, Settlements: []world.Settlement{{Type: world.SettlementPort}}},
			"sea":  {ID: "sea", IsSea: true},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("genoa", "venice"): {FactionA: "genoa", FactionB: "venice", Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"warship": {ID: "warship", Category: army.CategoryNavalWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"enemy_fleet":  {ID: "enemy_fleet", OwnerID: "genoa", RegionID: "sea", IsNaval: true, NavalMission: &army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "sea"}, Units: []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}}},
			"docked_enemy": {ID: "docked_enemy", OwnerID: "genoa", RegionID: "sea", IsNaval: true, DockedRegionID: "genoa_port", DockedSettlementID: "genoa_port_settlement", Units: []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}}},
		},
	}

	if got := gs.RegionBlockadePercent(gs.Regions["port"], "venice"); got != 50 {
		t.Fatalf("liman ablukası %%%d olmalıydı, got=%d", 50, got)
	}
}

func TestUnassignedFleetDoesNotCreateBlockade(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"port": {ID: "port", OwnerID: "venice", Neighbors: []world.RegionID{"sea"}, Settlements: []world.Settlement{{Type: world.SettlementPort}}},
			"sea":  {ID: "sea", IsSea: true},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("genoa", "venice"): {FactionA: "genoa", FactionB: "venice", Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"warship": {ID: "warship", Category: army.CategoryNavalWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"enemy": {ID: "enemy", OwnerID: "genoa", RegionID: "sea", IsNaval: true, Units: []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}}},
		},
	}
	if got := gs.RegionBlockadePercent(gs.Regions["port"], "venice"); got != 0 {
		t.Fatalf("görevsiz filo ekonomik abluka oluşturmamalıydı, got=%d", got)
	}
}

func TestPatrolProtectsPortLogisticsFromEnemyBlockade(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"port": {ID: "port", OwnerID: "venice", Neighbors: []world.RegionID{"sea"}, Settlements: []world.Settlement{{Type: world.SettlementPort}}},
			"sea":  {ID: "sea", IsSea: true},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("genoa", "venice"): {FactionA: "genoa", FactionB: "venice", Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"warship": {ID: "warship", Category: army.CategoryNavalWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"enemy":  {ID: "enemy", OwnerID: "genoa", RegionID: "sea", IsNaval: true, Units: []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}}},
			"patrol": {ID: "patrol", OwnerID: "venice", RegionID: "sea", IsNaval: true, NavalMission: &army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "sea"}, Units: []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}}},
		},
	}
	if got := gs.RegionBlockadePercent(gs.Regions["port"], "venice"); got != 0 {
		t.Fatalf("devriye dost liman lojistiğini korumalıydı, got=%d", got)
	}
}

func TestMerchantTradeBonusRequiresActiveConnectedCenterSea(t *testing.T) {
	gs, route := merchantTradeTestState()
	gs.Armies["merchant"].RegionID = "unknown_sea"
	if got := gs.MerchantFleetTradeRouteBonus(gs.Armies["merchant"], route); got != 0 {
		t.Fatalf("rota denizi dışındaki filo bonus vermemeliydi, got=%d", got)
	}
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
