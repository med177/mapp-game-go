package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAILogisticsUsesSharedGrainProductionDemandAndUpkeepRules(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 0},
		},
		Regions: map[world.RegionID]*world.Region{
			"farm": {
				ID:              "farm",
				OwnerID:         "player",
				Population:      200,
				BaseGrainOutput: 100,
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"field": {ID: "field", OwnerID: "player", RegionID: "farm", Units: []army.Unit{{TypeID: "inf"}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 2},
		},
		ActiveRegionEvents: []state.RegionEventStatus{{
			RegionID:               "farm",
			TurnsLeft:              2,
			GrainProductionPercent: -50,
			GrainDemandPercent:     100,
		}},
	}

	demand, eventCapacity, _ := aiRegionLogistics(gs, gs.Regions["farm"], "player")
	if demand != gs.EffectiveArmyGrainUpkeep(gs.Armies["field"]) {
		t.Fatalf("AI ordu talebi ortak efektif bakım kuralını kullanmalıydı: got=%d", demand)
	}
	if got := gs.RegionMilitaryGrainProduction(gs.Regions["farm"]); got != 26 {
		t.Fatalf("AI lojistik için ortak askeri üretim seam'i yanlış: got=%d", got)
	}

	gs.ActiveRegionEvents = nil
	_, normalCapacity, _ := aiRegionLogistics(gs, gs.Regions["farm"], "player")
	if eventCapacity >= normalCapacity {
		t.Fatalf("aktif tahıl olayı AI ikmal kapasitesini azaltmalıydı: event=%d normal=%d", eventCapacity, normalCapacity)
	}
}

func TestAIProcuresMilitaryIronFromConnectedTradeNetwork(t *testing.T) {
	gs := aiTestState()
	gs.ScenarioID = "1300_ottoman_rise"
	gs.Factions["ai_1"].Gold = 1000
	gs.Factions["ai_1"].Iron = 5
	gs.Factions["ai_2"].Iron = 100
	gs.TradeRoutes = []*economy.TradeRoute{{
		FromFactionID: "ai_1", ToFactionID: "ai_2", Good: economy.GoodCloth, AmountPerTurn: 1,
	}}
	gs.Relations[faction.RelationKey("ai_1", "player")].Stance = faction.StanceWar
	gs.MarketPrices = economy.CurrentMarketPrice{economy.GoodIron: 5}

	if got := aiProcureMilitaryIron(gs, "ai_1"); got != 35 {
		t.Fatalf("savaşta demir açığı olan AI ticaret ağından rezervini tamamlamalıydı: got=%d", got)
	}
	if gs.Factions["ai_1"].Iron != 40 || gs.Factions["ai_1"].Gold != 825 {
		t.Fatalf("askerî demir alımında alıcı stoğu/altını yanlış: %+v", gs.Factions["ai_1"])
	}
	if gs.Factions["ai_2"].Iron != 65 || gs.Factions["ai_2"].Gold != 275 {
		t.Fatalf("askerî demir alımında tedarikçi stoğu/geliri yanlış: %+v", gs.Factions["ai_2"])
	}
}

func TestAIProcuresEveryMissingProductionResource(t *testing.T) {
	gs := aiTestState()
	gs.ScenarioID = "1300_ottoman_rise"
	gs.Factions["ai_1"].Gold = 2000
	gs.Factions["ai_1"].Grain = 0
	gs.Factions["ai_1"].Iron = 0
	gs.Factions["ai_1"].Timber = 0
	gs.Factions["ai_1"].Stone = 0
	gs.Factions["ai_1"].Spice = 0
	gs.Factions["ai_1"].Cloth = 0
	gs.Factions["ai_2"].Grain = 200
	gs.Factions["ai_2"].Iron = 200
	gs.Factions["ai_2"].Timber = 200
	gs.Factions["ai_2"].Stone = 200
	gs.Factions["ai_2"].Spice = 200
	gs.Factions["ai_2"].Cloth = 200
	gs.Regions["a1"].Satisfaction = 100
	gs.Regions["a1"].Buildings = []string{"barracks"}
	gs.UnitTypes["inf"] = &army.UnitType{
		ID: "inf", Category: army.CategoryInfantry, RequiredBldg: "barracks", RequiredBldgLevel: 1,
		GoldCost: 100, GrainCost: 30, IronCost: 40, TimberCost: 25,
		StoneCost: 10, SpiceCost: 5, ClothCost: 6,
	}
	gs.TradeRoutes = []*economy.TradeRoute{{
		FromFactionID: "ai_1", ToFactionID: "ai_2", Good: economy.GoodCloth, AmountPerTurn: 1,
	}}
	gs.MarketPrices = economy.CurrentMarketPrice{
		economy.GoodGrain: 2, economy.GoodIron: 5, economy.GoodTimber: 3,
		economy.GoodStone: 4, economy.GoodSpice: 12, economy.GoodCloth: 8,
	}

	demand := economy.ResourceCost{Grain: 30, Iron: 40, Timber: 25, Stone: 10, Spice: 5, Cloth: 6}
	purchased := aiProcureStrategicResources(gs, "ai_1", nil)
	if purchased != demand {
		t.Fatalf("AI maliyetin tüm eksik ticari kaynaklarını tamamlamalıydı: got=%+v want=%+v", purchased, demand)
	}
	if gs.Factions["ai_1"].Gold != 1517 {
		t.Fatalf("kaynak alımı yalnız gerekli altını harcamalıydı: gold=%d", gs.Factions["ai_1"].Gold)
	}
	for _, kind := range economy.CostResourceKinds()[1:] {
		if got := economy.FactionResourceAmount(gs.Factions["ai_1"], kind); got != demand.Amount(kind) {
			t.Fatalf("%s açığı tamamlanmadı: got=%d want=%d", kind, got, demand.Amount(kind))
		}
	}
}
