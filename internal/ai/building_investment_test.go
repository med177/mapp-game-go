package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIBuildingInvestmentPrefersHigherROIRegion(t *testing.T) {
	gs := aiBuildingInvestmentTestState()
	gs.BuildingTypes = map[string]*city.Building{
		"market": aiTestBuilding("market", 120, 1.5, 1, 1.45, 2, 2),
	}
	gs.Regions["poor"] = aiTestEconomyRegion("poor", 20, 20, 2, 70)
	gs.Regions["rich"] = aiTestEconomyRegion("rich", 100, 20, 8, 70)

	candidate, ok := aiBestBuildingInvestment(gs, "ai", nil, nil)
	if !ok {
		t.Fatal("uygun market yatırımı bulunmalıydı")
	}
	if candidate.RegionID != "rich" || candidate.BuildingID != "market" {
		t.Fatalf("yüksek marjinal getirili bölge seçilmeliydi: %+v", candidate)
	}
	if candidate.ROIScore <= 0 {
		t.Fatalf("market ROI bileşeni pozitif olmalıydı: %+v", candidate)
	}
}

func TestAIBuildingInvestmentPrioritizesFarmDuringGrainBottleneck(t *testing.T) {
	gs := aiBuildingInvestmentTestState()
	gs.Factions["ai"].Grain = 20
	gs.BuildingTypes = map[string]*city.Building{
		"farm":   aiTestBuilding("farm", 80, 1.05, 1.65, 1, 4, 2),
		"market": aiTestBuilding("market", 120, 1.5, 1, 1.45, 2, 2),
	}
	gs.Regions["heartland"] = aiTestEconomyRegion("heartland", 90, 30, 6, 70)
	gs.UnitTypes = map[string]*army.UnitType{"hungry": {ID: "hungry", GrainUpkeep: 40}}
	gs.Armies = map[army.ArmyID]*army.Army{
		"field": {ID: "field", OwnerID: "ai", RegionID: "heartland", Units: []army.Unit{{TypeID: "hungry", CurrentHP: 100}}},
	}

	candidate, ok := aiBestBuildingInvestment(gs, "ai", nil, nil)
	if !ok || candidate.BuildingID != "farm" {
		t.Fatalf("tahıl darboğazında çiftlik seçilmeliydi: ok=%v candidate=%+v", ok, candidate)
	}
	if candidate.BottleneckScore < 100 {
		t.Fatalf("tahıl darboğazı skora açıkça yansımalıydı: %+v", candidate)
	}
}

func TestAIBuildingInvestmentCountsCivilianDemandForFarmPriority(t *testing.T) {
	gs := aiBuildingInvestmentTestState()
	gs.Factions["ai"].Grain = 200
	gs.BuildingTypes = map[string]*city.Building{
		"farm": {
			ID: "farm", NameTR: "farm", GoldCost: 80, GoldMod: 1, GrainMod: 1.3,
			TurnsRequired: 2, MaxPerRegion: 1,
		},
		"market": aiTestBuilding("market", 120, 1.5, 1, 1.45, 2, 2),
	}
	gs.Regions["heartland"] = aiTestEconomyRegion("heartland", 90, 30, 6, 70)
	gs.Regions["heartland"].Population = 1000 // 56 sivil talep, üretim ise yalnız 30.

	candidate, ok := aiBestBuildingInvestment(gs, "ai", nil, nil)
	if !ok || candidate.BuildingID != "farm" {
		t.Fatalf("sivil tahıl açığında çiftlik marketten önce gelmeliydi: ok=%v candidate=%+v", ok, candidate)
	}
}

func TestAIBuildingInvestmentPrioritizesWallsForCriticalDefenseObjective(t *testing.T) {
	gs := aiBuildingInvestmentTestState()
	gs.BuildingTypes = map[string]*city.Building{
		"farm":   aiTestBuilding("farm", 80, 1.05, 1.65, 1, 4, 2),
		"market": aiTestBuilding("market", 120, 1.5, 1, 1.45, 2, 2),
		"walls":  aiTestBuilding("walls", 280, 1.08, 1.03, 1, 6, 4),
	}
	gs.Regions["front"] = aiTestEconomyRegion("front", 50, 30, 3, 60)
	gs.Regions["front"].Neighbors = []world.RegionID{"enemy"}
	gs.Regions["safe"] = aiTestEconomyRegion("safe", 120, 40, 8, 70)
	gs.Regions["enemy"] = &world.Region{ID: "enemy", OwnerID: "enemy"}
	gs.Factions["enemy"] = &faction.Faction{ID: "enemy"}
	gs.AIPlans["ai"] = &state.AIPlanState{Kind: state.AIObjectiveDefend, TargetRegionIDs: []world.RegionID{"front"}}
	ctx := &StrategicContext{FactionID: "ai", gs: gs, Fronts: []AIFront{{
		EnemyFactionID: "enemy", FriendlyRegions: []world.RegionID{"front"}, AtWar: true,
		CriticalThreat: true, CapitalThreat: true, ThreatScore: 200,
	}}}

	candidate, ok := aiBestBuildingInvestment(gs, "ai", nil, ctx)
	if !ok || candidate.RegionID != "front" || candidate.BuildingID != "walls" {
		t.Fatalf("kritik savunma objective'i sur yatırımını seçmeliydi: ok=%v candidate=%+v", ok, candidate)
	}
	if candidate.ThreatScore < 300 || candidate.ObjectiveScore != 150 {
		t.Fatalf("tehdit ve objective bileşenleri sur skoruna yansımadı: %+v", candidate)
	}
}

func TestAIBuildingInvestmentUsesTempleForSevereInstability(t *testing.T) {
	gs := aiBuildingInvestmentTestState()
	gs.BuildingTypes = map[string]*city.Building{
		"market": aiTestBuilding("market", 120, 1.5, 1, 1.45, 2, 2),
		"temple": aiTestBuilding("temple", 100, 1.05, 1, 1, 10, 2),
	}
	gs.Regions["restive"] = aiTestEconomyRegion("restive", 70, 30, 4, 20)

	candidate, ok := aiBestBuildingInvestment(gs, "ai", nil, nil)
	if !ok || candidate.BuildingID != "temple" {
		t.Fatalf("ağır istikrarsızlıkta ibadet yeri seçilmeliydi: ok=%v candidate=%+v", ok, candidate)
	}
	if candidate.StabilityScore < 300 {
		t.Fatalf("memnuniyet açığı skora yansımadı: %+v", candidate)
	}
}

func TestAIBuildingInvestmentQueuePenaltyAndTieBreakAreDeterministic(t *testing.T) {
	gs := aiBuildingInvestmentTestState()
	gs.BuildingTypes = map[string]*city.Building{
		"market": aiTestBuilding("market", 120, 1.5, 1, 1.45, 2, 2),
	}
	gs.Regions["a"] = aiTestEconomyRegion("a", 60, 25, 4, 70)
	gs.Regions["b"] = aiTestEconomyRegion("b", 60, 25, 4, 70)

	candidate, ok := aiBestBuildingInvestment(gs, "ai", nil, nil)
	if !ok || candidate.RegionID != "a" {
		t.Fatalf("eşit skorda küçük region ID seçilmeliydi: ok=%v candidate=%+v", ok, candidate)
	}
	gs.ProductionQueue = []state.ProductionOrder{{
		ID: "prod_1", Kind: aiProductionKindBuilding, FactionID: "ai", RegionID: "a", TypeID: "market", TurnsLeft: 1,
	}}
	candidate, ok = aiBestBuildingInvestment(gs, "ai", nil, nil)
	if !ok || candidate.RegionID != "b" {
		t.Fatalf("kuyruk cezası yatırımı boş bölgeye kaydırmalıydı: ok=%v candidate=%+v", ok, candidate)
	}
}

func TestAIBuildingInvestmentQueuesSingleBudgetedConstruction(t *testing.T) {
	gs := aiBuildingInvestmentTestState()
	gs.Factions["ai"].Gold = 500
	gs.BuildingTypes = map[string]*city.Building{
		"market": aiTestBuilding("market", 120, 1.5, 1, 1.45, 2, 2),
	}
	gs.Regions["capital"] = aiTestEconomyRegion("capital", 80, 30, 6, 70)
	budget := &aiBudget{
		EmergencyGold: 100,
		Remaining:     map[aiBudgetCategory]int{aiBudgetEconomy: 150},
		Spent:         map[aiBudgetCategory]int{},
	}

	aiEconomyBuildWithStrategicContextAndSteps(gs, "ai", budget, nil, nil)
	if len(gs.ProductionQueue) != 1 {
		t.Fatalf("tur başına tek bina emri bekleniyordu: %+v", gs.ProductionQueue)
	}
	order := gs.ProductionQueue[0]
	if order.RegionID != "capital" || order.TypeID != "market" || order.TurnsLeft != 2 {
		t.Fatalf("beklenmeyen bina emri: %+v", order)
	}
	if gs.Factions["ai"].Gold != 380 || budget.Spent[aiBudgetEconomy] != 120 || budget.Remaining[aiBudgetEconomy] != 30 {
		t.Fatalf("bütçeli bina maliyeti yanlış uygulandı: faction=%+v budget=%+v", gs.Factions["ai"], budget)
	}
}

func TestAILegacyEconomyBuildKeepsOtherScenarioBehavior(t *testing.T) {
	gs := aiBuildingInvestmentTestState()
	gs.ScenarioID = "legacy"
	gs.BuildingTypes = map[string]*city.Building{
		"farm":   aiTestBuilding("farm", 80, 1.05, 1.65, 1, 4, 2),
		"market": aiTestBuilding("market", 120, 1.5, 1, 1.45, 2, 2),
	}
	gs.Regions["low_grain"] = aiTestEconomyRegion("low_grain", 80, 10, 4, 70)

	aiEconomyBuild(gs, "ai")
	if len(gs.ProductionQueue) != 1 || gs.ProductionQueue[0].TypeID != "farm" {
		t.Fatalf("diğer senaryoların eski çiftlik eşiği korunmalıydı: %+v", gs.ProductionQueue)
	}
}

func aiBuildingInvestmentTestState() *state.GameState {
	return &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Year:       1300,
		Month:      6,
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", Gold: 2000, Grain: 500, Iron: 500, Timber: 500, Stone: 500},
		},
		Regions:       map[world.RegionID]*world.Region{},
		Relations:     map[string]*faction.Relation{},
		Armies:        map[army.ArmyID]*army.Army{},
		UnitTypes:     map[string]*army.UnitType{},
		BuildingTypes: map[string]*city.Building{},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"ai": {Kind: state.AIObjectiveConsolidate},
		},
	}
}

func aiTestEconomyRegion(id world.RegionID, gold, grain, trade, satisfaction int) *world.Region {
	return &world.Region{
		ID: id, OwnerID: "ai", Terrain: world.TerrainPlain, BaseGoldIncome: gold,
		BaseGrainOutput: grain, TradeCapacity: trade, Satisfaction: satisfaction, TaxRate: 100,
	}
}

func aiTestBuilding(id string, goldCost int, goldMod, grainMod, tradeMod float64, satBonus, turns int) *city.Building {
	return &city.Building{
		ID: id, NameTR: id, GoldCost: goldCost, GoldMod: goldMod, GrainMod: grainMod,
		TradeCapacityMod: tradeMod, SatBonus: satBonus, TurnsRequired: turns, MaxPerRegion: 3,
	}
}
