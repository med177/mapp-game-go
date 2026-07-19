package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIStrategicRecruitRegionBalancesThroughputAndQueue(t *testing.T) {
	gs, ctx, unitType := aiRecruitRegionTestState()
	gs.Regions["high"] = aiRecruitTestRegion("high", 2, "anchor")
	gs.Regions["low"] = aiRecruitTestRegion("low", 1, "anchor")
	gs.Regions["anchor"].Neighbors = []world.RegionID{"high", "low"}

	if got := aiFindRecruitRegionForStrategicContext(gs, "ai", unitType, ctx); got != "high" {
		t.Fatalf("boş iki slotlu kışla hattı seçilmeliydi: got=%s", got)
	}

	gs.ProductionQueue = append(gs.ProductionQueue, state.ProductionOrder{
		Kind: aiProductionKindUnit, FactionID: "ai", RegionID: "high", TypeID: unitType.ID, TurnsLeft: 2,
	})
	if got := aiFindRecruitRegionForStrategicContext(gs, "ai", unitType, ctx); got != "low" {
		t.Fatalf("kuyruk cezası eşit kalan throughput'ta boş hattı seçmeliydi: got=%s", got)
	}
}

func TestAIStrategicRecruitRegionUsesWeightedFrontDistance(t *testing.T) {
	gs, ctx, unitType := aiRecruitRegionTestState()
	gs.Regions["a_near"] = aiRecruitTestRegion("a_near", 1, "pass")
	gs.Regions["pass"] = &world.Region{ID: "pass", OwnerID: "ai", Terrain: world.TerrainPass, Satisfaction: 70, BaseGrainOutput: 30, Neighbors: []world.RegionID{"a_near", "anchor"}}
	gs.Regions["z_far"] = aiRecruitTestRegion("z_far", 1, "plain")
	gs.Regions["plain"] = &world.Region{ID: "plain", OwnerID: "ai", Terrain: world.TerrainPlain, Satisfaction: 70, BaseGrainOutput: 30, Neighbors: []world.RegionID{"z_far", "anchor"}}
	gs.Regions["anchor"].Neighbors = []world.RegionID{"pass", "plain"}

	if got := aiFindRecruitRegionForStrategicContext(gs, "ai", unitType, ctx); got != "z_far" {
		t.Fatalf("ağırlıklı Dijkstra iki hoplu geçit yerine iki hoplu ova hattını seçmeliydi: got=%s", got)
	}
}

func TestAIStrategicRecruitRegionRejectsUnsafeProductionLines(t *testing.T) {
	tests := []struct {
		name   string
		unsafe func(*state.GameState, *StrategicContext)
	}{
		{
			name: "kuşatma",
			unsafe: func(gs *state.GameState, _ *StrategicContext) {
				gs.Sieges["a_unsafe"] = &state.SiegeState{RegionID: "a_unsafe"}
			},
		},
		{
			name: "yabancı ordu",
			unsafe: func(gs *state.GameState, _ *StrategicContext) {
				gs.Armies["foreign"] = &army.Army{ID: "foreign", OwnerID: "enemy", RegionID: "a_unsafe"}
			},
		},
		{
			name: "isyan riski",
			unsafe: func(gs *state.GameState, _ *StrategicContext) {
				gs.Regions["a_unsafe"].Satisfaction = 29
			},
		},
		{
			name: "kritik tehdit",
			unsafe: func(_ *state.GameState, ctx *StrategicContext) {
				ctx.Fronts[0].CriticalThreat = true
				ctx.Fronts[0].FriendlyRegions = []world.RegionID{"a_unsafe"}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gs, ctx, unitType := aiRecruitRegionTestState()
			gs.Regions["a_unsafe"] = aiRecruitTestRegion("a_unsafe", 3, "anchor")
			gs.Regions["z_safe"] = aiRecruitTestRegion("z_safe", 1, "anchor")
			gs.Regions["anchor"].Neighbors = []world.RegionID{"a_unsafe", "z_safe"}
			test.unsafe(gs, ctx)

			if got := aiFindRecruitRegionForStrategicContext(gs, "ai", unitType, ctx); got != "z_safe" {
				t.Fatalf("güvensiz yüksek-throughput hattı elenmeliydi: got=%s", got)
			}
		})
	}
}

func TestAIStrategicRecruitRegionProjectsPendingLogisticsDemand(t *testing.T) {
	gs, ctx, unitType := aiRecruitRegionTestState()
	unitType.GrainUpkeep = 2
	gs.Regions["a_overloaded"] = aiRecruitTestRegion("a_overloaded", 3, "anchor")
	gs.Regions["a_overloaded"].BaseGrainOutput = 0
	gs.Regions["z_supplied"] = aiRecruitTestRegion("z_supplied", 1, "anchor")
	gs.Regions["anchor"].Neighbors = []world.RegionID{"a_overloaded", "z_supplied"}
	gs.Armies["local"] = &army.Army{
		ID: "local", OwnerID: "ai", RegionID: "a_overloaded",
		Units: []army.Unit{{TypeID: unitType.ID, CurrentHP: army.MaxUnitHP}},
	}
	gs.ProductionQueue = append(gs.ProductionQueue, state.ProductionOrder{
		Kind: aiProductionKindUnit, FactionID: "ai", RegionID: "a_overloaded", TypeID: unitType.ID, TurnsLeft: 2,
	})

	demand, capacity, overload := aiProjectedRecruitRegionLogistics(gs, "ai", gs.Regions["a_overloaded"], unitType)
	if overload <= 0 || demand <= capacity {
		t.Fatalf("mevcut + pending + yeni birlik ikmal aşımı üretmeliydi: demand=%d capacity=%d overload=%d", demand, capacity, overload)
	}
	if got := aiFindRecruitRegionForStrategicContext(gs, "ai", unitType, ctx); got != "z_supplied" {
		t.Fatalf("projeksiyonda aşım yaratan yakın hat yerine ikmalli hat seçilmeliydi: got=%s", got)
	}
}

func TestAISiegeRecruitmentUsesMissingOffensiveFrontAnchor(t *testing.T) {
	gs, ctx, _ := aiRecruitRegionTestState()
	siegeType := aiTestLandUnitType("siege", army.CategorySiege, 15, 5, 40, 100, 2)
	gs.UnitTypes[siegeType.ID] = siegeType
	gs.Regions = map[world.RegionID]*world.Region{
		"a_rear":         aiRecruitTestRegion("a_rear", 1, "defense_anchor"),
		"defense_anchor": {ID: "defense_anchor", OwnerID: "ai", Terrain: world.TerrainPlain, Satisfaction: 70, BaseGrainOutput: 30, Neighbors: []world.RegionID{"a_rear", "bridge"}},
		"bridge":         {ID: "bridge", OwnerID: "ai", Terrain: world.TerrainPlain, Satisfaction: 70, BaseGrainOutput: 30, Neighbors: []world.RegionID{"defense_anchor", "offense_anchor"}},
		"offense_anchor": {ID: "offense_anchor", OwnerID: "ai", Terrain: world.TerrainPlain, Satisfaction: 70, BaseGrainOutput: 30, Neighbors: []world.RegionID{"bridge", "z_forward"}},
		"z_forward":      aiRecruitTestRegion("z_forward", 1, "offense_anchor"),
	}
	gs.Armies["assault"] = &army.Army{ID: "assault", OwnerID: "ai", RegionID: "a_rear", Units: aiUnitsOfType("inf", 2)}
	ctx.OwnedLandRegionIDs = []world.RegionID{"a_rear", "bridge", "defense_anchor", "offense_anchor", "z_forward"}
	ctx.Fronts = []AIFront{
		{EnemyFactionID: "other_enemy", AnchorRegionID: "defense_anchor", AtWar: true, ThreatScore: 200},
		{EnemyFactionID: "enemy", AnchorRegionID: "offense_anchor", AtWar: true, ObjectiveRelated: true},
	}
	ctx.ArmyAssignments = map[army.ArmyID]AIArmyAssignment{
		"assault": {Role: AIArmyRoleAssault, FrontFactionID: "enemy", AnchorRegionID: "enemy_target"},
	}

	if got := aiFindRecruitRegionForStrategicContext(gs, "ai", siegeType, ctx); got != "z_forward" {
		t.Fatalf("kuşatma desteği genel tehdit cephesine değil eksik hücum cephesine yakın üretilmeliydi: got=%s", got)
	}
}

func TestAIStrategicRecruitRegionTieBreaksByRegionID(t *testing.T) {
	gs, ctx, unitType := aiRecruitRegionTestState()
	gs.Regions["b"] = aiRecruitTestRegion("b", 1, "anchor")
	gs.Regions["a"] = aiRecruitTestRegion("a", 1, "anchor")
	gs.Regions["anchor"].Neighbors = []world.RegionID{"b", "a"}

	if got := aiFindRecruitRegionForStrategicContext(gs, "ai", unitType, ctx); got != "a" {
		t.Fatalf("eşit adaylarda küçük region ID seçilmeliydi: got=%s", got)
	}
}

func TestAIRecruitRegionPreservesLegacyScenarioSelection(t *testing.T) {
	gs, ctx, unitType := aiRecruitRegionTestState()
	gs.ScenarioID = "legacy_scenario"
	gs.Regions["a_unsafe"] = aiRecruitTestRegion("a_unsafe", 3, "anchor")
	gs.Regions["a_unsafe"].Satisfaction = 10
	gs.Regions["z_safe"] = aiRecruitTestRegion("z_safe", 1, "anchor")

	if got := aiFindRecruitRegionForStrategicContext(gs, "ai", unitType, ctx); got != "a_unsafe" {
		t.Fatalf("diğer senaryolar mevcut remaining-capacity/level seçimini korumalıydı: got=%s", got)
	}
}

func aiRecruitRegionTestState() (*state.GameState, *StrategicContext, *army.UnitType) {
	unitType := aiTestLandUnitType("inf", army.CategoryInfantry, 10, 10, 50, 80, 1)
	gs := &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":    {ID: "ai", Gold: 5000, Grain: 2000},
			"enemy": {ID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			"anchor": {ID: "anchor", OwnerID: "ai", Terrain: world.TerrainPlain, Satisfaction: 70, BaseGrainOutput: 30},
		},
		Armies:    make(map[army.ArmyID]*army.Army),
		UnitTypes: map[string]*army.UnitType{unitType.ID: unitType},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"ai": {Kind: state.AIObjectiveExpand, TargetFactionID: "enemy"},
		},
		Sieges: make(map[world.RegionID]*state.SiegeState),
	}
	ctx := &StrategicContext{
		FactionID: "ai", gs: gs,
		OwnedLandRegionIDs: []world.RegionID{"anchor"},
		Fronts:             []AIFront{{EnemyFactionID: "enemy", AnchorRegionID: "anchor", AtWar: true, ObjectiveRelated: true}},
		ArmyAssignments:    make(map[army.ArmyID]AIArmyAssignment),
	}
	return gs, ctx, unitType
}

func aiRecruitTestRegion(id world.RegionID, barracksLevel int, neighbors ...world.RegionID) *world.Region {
	buildings := make([]string, barracksLevel)
	for index := range buildings {
		buildings[index] = "barracks"
	}
	return &world.Region{
		ID: id, OwnerID: "ai", Terrain: world.TerrainPlain, Satisfaction: 70,
		BaseGrainOutput: 30, Buildings: buildings, Neighbors: neighbors,
	}
}
