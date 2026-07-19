package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAICompositionTargetsFollowApprovedPlanShares(t *testing.T) {
	tests := []struct {
		kind state.AIObjectiveKind
		want aiCompositionTarget
	}{
		{state.AIObjectiveExpand, aiCompositionTarget{Infantry: 55, Cavalry: 25, Siege: 20}},
		{state.AIObjectiveDefend, aiCompositionTarget{Infantry: 75, Cavalry: 15, Siege: 10}},
		{state.AIObjectiveConsolidate, aiCompositionTarget{Infantry: 65, Cavalry: 25, Siege: 10}},
	}
	for _, test := range tests {
		got := aiCompositionTargetForPlan(&state.AIPlanState{Kind: test.kind})
		if got != test.want {
			t.Fatalf("%s kompozisyon hedefi yanlış: got=%+v want=%+v", test.kind, got, test.want)
		}
		if got.Infantry+got.Cavalry+got.Siege != 100 {
			t.Fatalf("kompozisyon toplamı 100 olmalıydı: %+v", got)
		}
	}
}

func TestAIStrategicRecruitmentFillsLargestCompositionDeficit(t *testing.T) {
	gs, ctx := aiUnitCompositionTestState(state.AIObjectiveExpand)
	gs.UnitTypes = map[string]*army.UnitType{
		"inf":   aiTestLandUnitType("inf", army.CategoryInfantry, 10, 10, 50, 80, 2),
		"cav":   aiTestLandUnitType("cav", army.CategoryCavalry, 10, 10, 50, 80, 2),
		"siege": aiTestLandUnitType("siege", army.CategorySiege, 10, 10, 50, 80, 2),
	}
	gs.Armies["field"].Units = append(aiUnitsOfType("inf", 6), aiUnitsOfType("cav", 4)...)

	got := aiSelectBestUnitForStrategicContext(gs, gs.Factions["ai"], nil, ctx)
	if got != "siege" {
		t.Fatalf("en büyük kategori açığı kuşatmadaydı: got=%s", got)
	}
}

func TestAIFortifiedObjectiveFillsEachOffensiveSiegeShortfall(t *testing.T) {
	gs, ctx := aiUnitCompositionTestState(state.AIObjectiveExpand)
	gs.UnitTypes = map[string]*army.UnitType{
		"inf":   aiTestLandUnitType("inf", army.CategoryInfantry, 14, 12, 60, 80, 2),
		"siege": aiTestLandUnitType("siege", army.CategorySiege, 18, 4, 35, 100, 3),
	}
	gs.Regions["target"].Buildings = []string{"walls"}
	gs.Armies["field"].Units = aiUnitsOfType("inf", 8)
	gs.Armies["second"] = &army.Army{ID: "second", OwnerID: "ai", RegionID: "home", Units: aiUnitsOfType("inf", 4)}
	gs.Armies["reserve_siege"] = &army.Army{ID: "reserve_siege", OwnerID: "ai", RegionID: "home", IsGarrison: true, Units: aiUnitsOfType("siege", 5)}
	ctx.ArmyAssignments = map[army.ArmyID]AIArmyAssignment{
		"field":  {Role: AIArmyRoleAssault, AnchorRegionID: "target"},
		"second": {Role: AIArmyRoleAssault, AnchorRegionID: "target"},
	}

	needs := aiBuildRecruitmentBattleNeeds(gs, "ai", ctx)
	if !needs.FortifiedTarget || needs.SiegeShortfall != 2 {
		t.Fatalf("iki kuşatma birimsiz hücum ordusu tespit edilmeliydi: %+v", needs)
	}
	got := aiSelectBestUnitForStrategicContext(gs, gs.Factions["ai"], nil, ctx)
	if got != "siege" {
		t.Fatalf("global kuşatma oranı yüksek olsa bile hücum açığı kapatılmalıydı: got=%s", got)
	}

	gs.ProductionQueue = append(gs.ProductionQueue, state.ProductionOrder{Kind: aiProductionKindUnit, FactionID: "ai", RegionID: "home", TypeID: "siege"})
	needs = aiBuildRecruitmentBattleNeeds(gs, "ai", ctx)
	if needs.SiegeShortfall != 1 {
		t.Fatalf("bekleyen kuşatma üretimi açık sayısından düşmeliydi: %+v", needs)
	}
}

func TestAIRecruitmentWeightsUseRealTerrainAndEnemyStats(t *testing.T) {
	gs, ctx := aiUnitCompositionTestState(state.AIObjectiveExpand)
	gs.Regions["target"].Terrain = world.TerrainMountain
	gs.Regions["target"].Buildings = []string{"walls"}
	gs.UnitTypes = map[string]*army.UnitType{
		"enemy_guard": aiTestLandUnitType("enemy_guard", army.CategoryInfantry, 5, 30, 70, 0, 1),
	}
	gs.Armies["enemy"] = &army.Army{ID: "enemy", OwnerID: "enemy", RegionID: "target", Units: aiUnitsOfType("enemy_guard", 2)}
	ctx.WarEnemies = []faction.FactionID{"enemy"}

	needs := aiBuildRecruitmentBattleNeeds(gs, "ai", ctx)
	if needs.TargetTerrain != world.TerrainMountain || needs.AttackWeight != 6 || needs.DefenseWeight != 2 || needs.MoraleWeight != 2 {
		t.Fatalf("dağ savunması ve yüksek düşman savunması hücum/moral ihtiyacına yansımalıydı: %+v", needs)
	}
	if !needs.FortifiedTarget {
		t.Fatal("düşman suru kuşatma ihtiyacı üretmeliydi")
	}

	gs.AIPlans["ai"] = &state.AIPlanState{Kind: state.AIObjectiveDefend, TargetRegionIDs: []world.RegionID{"home"}}
	gs.Regions["home"].Terrain = world.TerrainPass
	delete(gs.Armies, "enemy")
	ctx.WarEnemies = nil
	needs = aiBuildRecruitmentBattleNeeds(gs, "ai", ctx)
	if needs.AttackWeight != 2 || needs.DefenseWeight != 5 || needs.MoraleWeight != 2 || needs.FortifiedTarget {
		t.Fatalf("dost geçit savunması savunma/moral ağırlığı vermeliydi: %+v", needs)
	}
}

func TestAIStrategicRecruitmentUsesCostAndUpkeepWithinCategory(t *testing.T) {
	gs, ctx := aiUnitCompositionTestState(state.AIObjectiveDefend)
	gs.UnitTypes = map[string]*army.UnitType{
		"efficient": aiTestLandUnitType("efficient", army.CategoryInfantry, 14, 14, 60, 100, 2),
		"wasteful":  aiTestLandUnitType("wasteful", army.CategoryInfantry, 14, 14, 60, 400, 8),
	}
	gs.UnitTypes["efficient"].GrainCost = 10
	gs.UnitTypes["wasteful"].GrainCost = 40

	got := aiSelectBestUnitForStrategicContext(gs, gs.Factions["ai"], nil, ctx)
	if got != "efficient" {
		t.Fatalf("aynı savaş değerinde düşük maliyet/bakım seçilmeliydi: got=%s", got)
	}
}

func aiUnitCompositionTestState(kind state.AIObjectiveKind) (*state.GameState, *StrategicContext) {
	gs := &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Year:       1300,
		Month:      6,
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":    {ID: "ai", Gold: 5000, Grain: 2000, Iron: 2000, Timber: 2000, Stone: 2000, Research: faction.ResearchState{Completed: map[string]bool{}}},
			"enemy": {ID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home":   {ID: "home", OwnerID: "ai", Terrain: world.TerrainPlain, Buildings: []string{"barracks", "barracks", "barracks"}, BaseGrainOutput: 80, Satisfaction: 70, TaxRate: 30},
			"target": {ID: "target", OwnerID: "enemy", Terrain: world.TerrainPlain},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "enemy"): {FactionA: "ai", FactionB: "enemy", Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"field": {ID: "field", OwnerID: "ai", RegionID: "home"},
		},
		UnitTypes: map[string]*army.UnitType{},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"ai": {Kind: kind, TargetFactionID: "enemy", TargetRegionIDs: []world.RegionID{"target"}},
		},
	}
	ctx := &StrategicContext{
		FactionID: "ai", gs: gs, WarEnemies: []faction.FactionID{"enemy"},
		ArmyAssignments: map[army.ArmyID]AIArmyAssignment{"field": {Role: AIArmyRoleAssault, AnchorRegionID: "target"}},
	}
	return gs, ctx
}

func aiTestLandUnitType(id string, category army.UnitCategory, attack, defense, morale, goldCost, upkeep int) *army.UnitType {
	return &army.UnitType{
		ID: id, Category: category, Attack: attack, Defense: defense, Morale: morale,
		GoldCost: goldCost, GrainUpkeep: upkeep, TurnsRequired: 2, RequiredBldg: "barracks", RequiredBldgLevel: 1,
	}
}

func aiUnitsOfType(typeID string, count int) []army.Unit {
	units := make([]army.Unit, count)
	for index := range units {
		units[index] = army.Unit{TypeID: typeID, CurrentHP: 100}
	}
	return units
}
