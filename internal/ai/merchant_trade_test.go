package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiMerchantTradeTestState() *state.GameState {
	return &state.GameState{
		ScenarioID: "1300_ottoman_rise", Year: 1300,
		Factions: map[faction.FactionID]*faction.Faction{
			"venice": {ID: "venice", Gold: 2000, Grain: 500, Iron: 500, Timber: 500, Stone: 500, Research: faction.ResearchState{Completed: map[string]bool{"harbor_administration": true}}},
			"mamluk": {ID: "mamluk"},
		},
		Regions: map[world.RegionID]*world.Region{
			"venice":   {ID: "venice", OwnerID: "venice", Neighbors: []world.RegionID{"adriatic"}, Buildings: []string{"port", "port"}, TradeCapacity: 9},
			"egypt":    {ID: "egypt", OwnerID: "mamluk", Neighbors: []world.RegionID{"med"}, TradeCapacity: 6},
			"adriatic": {ID: "adriatic", IsSea: true, Neighbors: []world.RegionID{"venice", "med"}},
			"med":      {ID: "med", IsSea: true, Neighbors: []world.RegionID{"adriatic", "egypt"}},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{
			{ID: "venice", Links: []world.RegionID{"egypt"}, Tier: world.TradeCenterPrimary},
			{ID: "egypt", Links: []world.RegionID{"venice"}, Tier: world.TradeCenterSecondary},
		}},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "venice", ToFactionID: "mamluk", Good: economy.GoodCloth, AmountPerTurn: 2, GoldPerUnit: 8},
			{FromFactionID: "mamluk", ToFactionID: "venice", Good: economy.GoodSpice, AmountPerTurn: 2, GoldPerUnit: 12},
		},
		UnitTypes: map[string]*army.UnitType{
			"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade, RequiredBldg: "port", RequiredBldgLevel: 2, RequiredTech: []string{"harbor_administration"}, GoldCost: 180, TimberCost: 26, TurnsRequired: 3},
			"warship":       {ID: "warship", Category: army.CategoryNavalWar, Attack: 28, Defense: 18, Morale: 60, RequiredBldg: "port", RequiredBldgLevel: 3, RequiredTech: []string{"naval_doctrine"}, GoldCost: 400, TimberCost: 34, TurnsRequired: 4},
		},
		BuildingTypes: map[string]*city.Building{
			"port": {ID: "port", MaxPerRegion: 3, GoldCost: 200, TimberCost: 36, TurnsRequired: 2},
		},
		Armies: map[army.ArmyID]*army.Army{
			"merchant": {ID: "merchant", OwnerID: "venice", RegionID: "adriatic", IsNaval: true, Units: []army.Unit{{TypeID: "merchant_ship", CurrentHP: 100}}},
		},
	}
}

func Test1300MerchantAIKeepsAssignmentAndQueuesMissingShip(t *testing.T) {
	gs := aiMerchantTradeTestState()
	aiExecuteMerchantTradeStrategy(gs, "venice", nil, nil, nil)
	fleet := gs.Armies["merchant"]
	if fleet.TradeRouteKey != "venice->mamluk" {
		t.Fatalf("merchant filosu yalnız kendi ihracat rotasına atanmalıydı, got=%q", fleet.TradeRouteKey)
	}
	if len(gs.ProductionQueue) != 1 || gs.ProductionQueue[0].TypeID != "merchant_ship" || gs.ProductionQueue[0].RegionID != "venice" {
		t.Fatalf("eksik rota kapasitesi için bir merchant emri bekleniyordu: %+v", gs.ProductionQueue)
	}
}

func Test1300MerchantAIBuildsRequiredPortLevelBeforeShip(t *testing.T) {
	gs := aiMerchantTradeTestState()
	gs.Regions["venice"].Buildings = []string{"port"}
	aiExecuteMerchantTradeStrategy(gs, "venice", nil, nil, nil)
	if len(gs.ProductionQueue) != 1 || gs.ProductionQueue[0].Kind != aiProductionKindBuilding || gs.ProductionQueue[0].TypeID != "port" {
		t.Fatalf("merchant gemisinden önce ikinci liman seviyesi açılmalıydı: %+v", gs.ProductionQueue)
	}
}

func Test1300MerchantBudgetReservesPortAndFirstShipResources(t *testing.T) {
	gs := aiMerchantTradeTestState()
	gs.Regions["venice"].Buildings = []string{"port"}
	gs.Factions["venice"].Timber = 80
	budget := prepareAIBudget(gs, "venice", nil)
	if budget == nil {
		t.Fatal("1300 stratejik bütçesi üretilmeliydi")
	}
	if budget.ResourceReserve.Gold != 380 || budget.ResourceReserve.Timber != 62 {
		t.Fatalf("ilk merchant hattı liman yükseltmesi ve bir gemiyi birlikte rezerve etmeliydi: %+v", budget.ResourceReserve)
	}
	marketCost := economy.ResourceCost{Gold: 120, Timber: 28}
	if aiCanAffordForBudget(gs.Factions["venice"], marketCost, budget, aiBudgetEconomy) {
		t.Fatal("ekonomi yatırımı merchant için ayrılan keresteyi tüketmemeliydi")
	}
}

func Test1300MerchantFleetMovesToAssignedTradeCenterSea(t *testing.T) {
	gs := aiMerchantTradeTestState()
	gs.Regions["outer"] = &world.Region{ID: "outer", IsSea: true, Neighbors: []world.RegionID{"adriatic"}}
	gs.Regions["adriatic"].Neighbors = append(gs.Regions["adriatic"].Neighbors, "outer")
	fleet := gs.Armies["merchant"]
	fleet.RegionID = "outer"
	fleet.TradeRouteKey = "venice->mamluk"
	if got, handled := aiMerchantTradeFleetMove(gs, fleet, nil); !handled || got != "adriatic" {
		t.Fatalf("atanmış merchant filosu rota merkezi denizine ilerlemeliydi: got=%s handled=%v", got, handled)
	}
}

func Test1300DockedMerchantFleetUndocksBeforeTradeRouteMove(t *testing.T) {
	gs := aiMerchantTradeTestState()
	fleet := gs.Armies["merchant"]
	fleet.DockedRegionID = "venice"
	fleet.DockedSettlementID = "venice_port"
	fleet.MovePoints = 3

	aiExecuteMerchantTradeStrategy(gs, "venice", nil, nil, nil)
	ctx := prepareStrategicContext(gs, "venice")
	if !fleet.IsDocked() {
		t.Fatal("test merchant filosu başlangıçta limanda olmalıydı")
	}
	if got := chooseBestMoveWithStrategicContext(gs, fleet, ctx); got != "adriatic" {
		t.Fatalf("docked merchant filosu önce kendi deniz ankrajına çıkmalıydı, got=%s", got)
	}

	outcome := executeMove(gs, fleet, "adriatic", "venice")
	if !outcome.survived {
		t.Fatal("docked merchant filosunun denize çıkışı filoyu yok etmemeliydi")
	}
	if fleet.IsDocked() || !fleet.IsAtSea() {
		t.Fatalf("merchant filosu denize çıktıktan sonra docked kalmamalıydı: %+v", fleet)
	}
	if fleet.RegionID != "adriatic" || fleet.MovePoints != 2 {
		t.Fatalf("undock sonrası deniz ankrajı veya hareket puanı hatalı: region=%s move=%d", fleet.RegionID, fleet.MovePoints)
	}
}

func Test1300WarshipPatrolMovesTowardActiveTradeSea(t *testing.T) {
	gs := aiMerchantTradeTestState()
	gs.Regions["outer"] = &world.Region{ID: "outer", IsSea: true, Neighbors: []world.RegionID{"adriatic"}}
	gs.Regions["adriatic"].Neighbors = append(gs.Regions["adriatic"].Neighbors, "outer")
	gs.Armies["guard"] = &army.Army{
		ID: "guard", OwnerID: "venice", RegionID: "outer", IsNaval: true,
		Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}, MovePoints: 3, MaxMovePoints: 3,
	}
	aiExecuteMerchantTradeStrategy(gs, "venice", nil, nil, nil)
	ctx := prepareStrategicContext(gs, "venice")

	if got := chooseBestMoveWithStrategicContext(gs, gs.Armies["guard"], ctx); got != "adriatic" {
		t.Fatalf("görevsiz warship aktif ticaret denizine devriye yapmalıydı, got=%s", got)
	}
	assignment, ok := ctx.ArmyAssignments["guard"]
	if !ok || assignment.Role != AIArmyRolePatrol {
		t.Fatalf("görevsiz savaş filosu patrol rolü almalıydı: %+v", assignment)
	}
}

func Test1300ThreatenedTradeCenterQueuesEscortBeforeMerchant(t *testing.T) {
	gs := aiMerchantTradeTestState()
	// Yeni tek yönlü modelde Venice yalnızca kendi ihracat rotasını kullanır.
	// Bu testte o rotanın hedef denizini Venice limanına taşıyarak escort
	// önceliğini koru.
	gs.Regions["egypt"].Neighbors = []world.RegionID{"adriatic"}
	gs.Regions["adriatic"].Neighbors = append(gs.Regions["adriatic"].Neighbors, "egypt")
	gs.Factions["venice"].Research.Completed["naval_doctrine"] = true
	gs.Regions["venice"].Buildings = []string{"port", "port", "port"}
	gs.Factions["enemy"] = &faction.Faction{ID: "enemy"}
	gs.Relations = map[string]*faction.Relation{
		faction.RelationKey("venice", "enemy"): {FactionA: "venice", FactionB: "enemy", Stance: faction.StanceWar},
	}
	gs.Regions["threat"] = &world.Region{ID: "threat", IsSea: true, Neighbors: []world.RegionID{"adriatic"}}
	gs.Regions["adriatic"].Neighbors = append(gs.Regions["adriatic"].Neighbors, "threat")
	gs.Armies["enemy"] = &army.Army{
		ID: "enemy", OwnerID: "enemy", RegionID: "threat", IsNaval: true,
		Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}, {TypeID: "warship", CurrentHP: 100}},
	}

	aiExecuteMerchantTradeStrategy(gs, "venice", nil, nil, nil)
	if len(gs.ProductionQueue) < 2 {
		t.Fatalf("yüzde 110 eşiği tek savaş gemisiyle sağlanmıyorsa gereken stack birlikte kuyruğa alınmalıydı: %+v", gs.ProductionQueue)
	}
	for _, order := range gs.ProductionQueue {
		if order.TypeID != "warship" {
			t.Fatalf("tehditli ticaret merkezi merchant yerine önce escort istemeliydi: %+v", gs.ProductionQueue)
		}
	}
}

func Test1300AssignedMerchantFleetDoesNotMergeWithWarFleet(t *testing.T) {
	gs := aiMerchantTradeTestState()
	gs.Armies["merchant"].TradeRouteKey = "venice->mamluk"
	gs.Armies["war"] = &army.Army{
		ID: "war", OwnerID: "venice", RegionID: "adriatic", IsNaval: true,
		Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}},
	}
	aiConsolidateArmies(gs, "venice")
	if len(gs.Armies) != 2 {
		t.Fatalf("merchant görev filosu savaş filosuyla birleşmemeliydi: %+v", gs.Armies)
	}
}
