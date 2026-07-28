package ai

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIProcuresGrainFromConnectedTradeNetwork(t *testing.T) {
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
	gs.TradeRoutes = []*economy.TradeRoute{{
		FromFactionID: "ai_1", ToFactionID: "ai_2", Good: economy.GoodCloth, AmountPerTurn: 1,
	}}

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
