package ai

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIProcuresGrainFromOpenMarket(t *testing.T) {
	gs := aiTestState()
	gs.ScenarioID = "1300_ottoman_rise"
	gs.Factions["ai_1"].Gold = 600
	gs.Factions["ai_1"].Grain = 0
	gs.Factions["ai_2"].Gold = 100
	gs.Factions["ai_2"].Grain = 500
	gs.GrainEconomy = map[faction.FactionID]state.GrainEconomyStatus{
		"ai_1": {TotalDemand: 20},
	}
	gs.MarketPrices = economy.CurrentMarketPrice{economy.GoodGrain: 2}
	if got := aiProcureGrain(gs, "ai_1"); got != 40 {
		t.Fatalf("AI iki aylık tahıl penceresini satın almalıydı: got=%d", got)
	}
	if gs.Factions["ai_1"].Grain != 40 || gs.Factions["ai_1"].Gold != 520 {
		t.Fatalf("alıcının tahıl/altın stoğu yanlış: %+v", gs.Factions["ai_1"])
	}
	if gs.Factions["ai_2"].Grain != 460 || gs.Factions["ai_2"].Gold != 180 {
		t.Fatalf("tedarikçinin stoğu/geliri yanlış: %+v", gs.Factions["ai_2"])
	}
}

func TestRefreshMarketOrdersPublishesAIGrainSurplusAndGoldBoundDemand(t *testing.T) {
	gs := aiTestState()
	gs.ScenarioID = "1300_ottoman_rise"
	gs.Factions["ai_1"].Gold = 600
	gs.Factions["ai_1"].Grain = 0
	gs.Factions["ai_2"].Gold = 100
	gs.Factions["ai_2"].Grain = 500
	gs.GrainEconomy = map[faction.FactionID]state.GrainEconomyStatus{
		"ai_1": {TotalDemand: 20},
		"ai_2": {TotalDemand: 20},
	}
	gs.MarketPrices = economy.CurrentMarketPrice{economy.GoodGrain: 2}

	RefreshMarketOrders(gs)
	if got := gs.MarketSellOffer("ai_2", economy.GoodGrain); got != 400 {
		t.Fatalf("AI satış arzı güvenli rezerv üstündeki fazlayı göstermeli: got=%d want=400", got)
	}
	if got := gs.MarketBuyOrder("ai_1", economy.GoodGrain, 2); got != 60 {
		t.Fatalf("AI alım talebi stratejik üç aylık açığı göstermeli: got=%d want=60", got)
	}

	gs.Factions["ai_1"].Gold = 90
	RefreshMarketOrders(gs)
	if got := gs.MarketBuyOrder("ai_1", economy.GoodGrain, 2); got != 5 {
		t.Fatalf("AI alım talebi altın rezervini koruyarak sınırlandırılmalı: got=%d want=5", got)
	}
}

func TestAIOpenMarketNeverBuysFromEnemy(t *testing.T) {
	gs := aiTestState()
	gs.ScenarioID = "1300_ottoman_rise"
	gs.Factions["ai_1"].Gold = 600
	gs.Factions["ai_1"].Grain = 0
	gs.Factions["ai_2"].Grain = 500
	gs.Factions["enemy"] = &faction.Faction{ID: "enemy", Grain: 1000}
	gs.Relations[faction.RelationKey("ai_1", "enemy")] = &faction.Relation{FactionA: "ai_1", FactionB: "enemy", Stance: faction.StanceWar}
	gs.GrainEconomy = map[faction.FactionID]state.GrainEconomyStatus{
		"ai_1":  {TotalDemand: 20},
		"ai_2":  {TotalDemand: 20},
		"enemy": {TotalDemand: 20},
	}
	gs.MarketPrices = economy.CurrentMarketPrice{economy.GoodGrain: 2}

	if got := aiProcureGrain(gs, "ai_1"); got != 40 {
		t.Fatalf("barıştaki açık pazar satıcısından tahıl alınmalıydı: got=%d", got)
	}
	if got := gs.Factions["enemy"].Grain; got != 1000 {
		t.Fatalf("AI düşman devletten açık pazar alımı yapmamalıydı: enemy_grain=%d", got)
	}
}

func TestAIGrainProcurementUsesReserveInsteadOfStorageCapacityAsSupplierLimit(t *testing.T) {
	gs := aiTestState()
	gs.ScenarioID = "1300_ottoman_rise"
	gs.Factions["ai_1"].Gold = 600
	gs.Factions["ai_1"].Grain = 0
	gs.Factions["ai_2"].Gold = 100
	gs.Factions["ai_2"].Grain = 260
	gs.GrainEconomy = map[faction.FactionID]state.GrainEconomyStatus{
		"ai_1": {TotalDemand: 20},
		// Altı aylık depolama hedefi 1.000 olsa da satıcının üç aylık
		// ihtiyacı yalnız 180; kalan 80 tahıl piyasaya sunulabilmeli.
		"ai_2": {TotalDemand: 60, StorageCapacity: 1000},
	}
	gs.MarketPrices = economy.CurrentMarketPrice{economy.GoodGrain: 2}
	if got := aiProcureGrain(gs, "ai_1"); got != 40 {
		t.Fatalf("satıcı kendi üç aylık rezervini korurken alıcıya tahıl satmalıydı: got=%d", got)
	}
	if got := gs.Factions["ai_2"].Grain; got != 220 {
		t.Fatalf("satıcının tahıl bakiyesi yanlış: got=%d want=220", got)
	}
}

func TestAIEconomyPrioritizesFirstGranary(t *testing.T) {
	gs := aiTestState()
	gs.ScenarioID = "1300_ottoman_rise"
	gs.Factions["ai_1"].Gold = 500
	gs.Factions["ai_1"].Grain = 500
	gs.BuildingTypes = map[string]*city.Building{
		"granary": {ID: "granary", GoldCost: 160, GrainCost: 24, MaxPerRegion: 1, TurnsRequired: 3, StorageCapacity: 100},
	}

	candidate, ok := aiBestBuildingInvestment(gs, "ai_1", nil, nil)
	if !ok || candidate.BuildingID != "granary" {
		t.Fatalf("ambarı olmayan AI ilk ekonomi yatırımında ambarı seçmeliydi: candidate=%+v ok=%v", candidate, ok)
	}
}

func TestAIQueuesFirstBarracksBeforeManpowerCapIsNearFull(t *testing.T) {
	gs := aiTestState()
	gs.Regions["a2"] = &world.Region{ID: "a2", OwnerID: "ai_1"}
	gs.BuildingTypes = map[string]*city.Building{
		"barracks": {ID: "barracks", GoldCost: 150, MaxPerRegion: 1, TurnsRequired: 2},
	}
	gs.Factions["ai_1"].Gold = 500
	gs.Factions["ai_1"].Grain = 500

	aiRecruitAndBuildWithBudgetAndSteps(gs, "ai_1", nil, nil)

	if len(gs.ProductionQueue) != 1 || gs.ProductionQueue[0].TypeID != "barracks" {
		t.Fatalf("boş ordulu AI ilk kışlayı kuyruğa almalıydı: %+v", gs.ProductionQueue)
	}
}
