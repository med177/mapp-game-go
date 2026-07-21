package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiNavalMissionTestState(unitCount int) *state.GameState {
	units := make([]army.Unit, unitCount)
	for index := range units {
		units[index] = army.Unit{TypeID: "inf", CurrentHP: 100}
	}
	return &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Turn:       1,
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":    {ID: "ai", Gold: 2000, Grain: 500, Timber: 500, Iron: 500, Stone: 500},
			"enemy": {ID: "enemy"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "enemy"): {FactionA: "ai", FactionB: "enemy", Stance: faction.StanceWar},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID: "home", OwnerID: "ai", Neighbors: []world.RegionID{"port"}, Satisfaction: 50,
			},
			"port": {
				ID: "port", OwnerID: "ai", Neighbors: []world.RegionID{"home", "sea_home"},
				Buildings: []string{"port", "port", "port"}, Satisfaction: 50,
			},
			"sea_home": {
				ID: "sea_home", IsSea: true, Neighbors: []world.RegionID{"port", "sea_aux", "sea_target"},
			},
			"sea_aux": {
				ID: "sea_aux", IsSea: true, Neighbors: []world.RegionID{"sea_home"},
			},
			"sea_target": {
				ID: "sea_target", IsSea: true, Neighbors: []world.RegionID{"sea_home", "target"},
			},
			"target": {
				ID: "target", OwnerID: "enemy", Neighbors: []world.RegionID{"sea_target"}, Satisfaction: 50,
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"field": {
				ID: "field", OwnerID: "ai", RegionID: "home", Units: units, MovePoints: 2, MaxMovePoints: 2,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {
				ID: "inf", Category: army.CategoryInfantry, Attack: 10, Defense: 10, Morale: 50, Embarkable: true,
			},
			"transport": {
				ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 5,
				RequiredBldg: "port", RequiredBldgLevel: 1, GoldCost: 100, TimberCost: 10, TurnsRequired: 2,
			},
		},
		BuildingTypes: map[string]*city.Building{
			"port": {ID: "port", MaxPerRegion: 3, GoldCost: 200, TimberCost: 20, TurnsRequired: 2},
		},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"ai": {
				ObjectiveID: "overseas", Kind: state.AIObjectiveExpand, TargetFactionID: "enemy",
				TargetRegionIDs: []world.RegionID{"target"}, StartedTurn: 1, ReassessTurn: 12,
			},
		},
	}
}

func Test1300NavalMissionMatchesExistingAndPendingTransportCapacity(t *testing.T) {
	gs := aiNavalMissionTestState(12)
	gs.Armies["transport"] = &army.Army{
		ID: "transport", OwnerID: "ai", RegionID: "sea_aux", IsNaval: true,
		Units: []army.Unit{{TypeID: "transport", CurrentHP: 100}}, MovePoints: 3,
	}
	gs.ProductionQueue = []state.ProductionOrder{{
		ID: "prod_1", Kind: aiProductionKindUnit, FactionID: "ai", RegionID: "port", TypeID: "transport", TurnsLeft: 1,
	}}

	ctx := prepareStrategicContext(gs, "ai")
	if ctx.navalMission == nil {
		t.Fatal("kara yolu olmayan aktif objective için deniz görevi üretilmeliydi")
	}
	if got := ctx.navalMission.AvailableCapacity; got != 10 {
		t.Fatalf("mevcut ve kuyruktaki kapasite birlikte sayılmalıydı, got=%d", got)
	}
	if got := ctx.navalMission.MissingCapacity; got != 2 {
		t.Fatalf("12 kişilik ordu için yalnız 2 kapasite eksik olmalıydı, got=%d", got)
	}

	aiNavalStrategyWithStrategicContextAndSteps(gs, "ai", nil, ctx, nil)
	if len(gs.ProductionQueue) != 2 || gs.ProductionQueue[1].TypeID != "transport" || gs.ProductionQueue[1].RegionID != "port" {
		t.Fatalf("eksik kapasite için yalnız bir transport emri bekleniyordu: %+v", gs.ProductionQueue)
	}

	ctx = prepareStrategicContext(gs, "ai")
	aiNavalStrategyWithStrategicContextAndSteps(gs, "ai", nil, ctx, nil)
	if len(gs.ProductionQueue) != 2 {
		t.Fatalf("yeterli boş kapasite varken yeni transport açılmamalıydı: %+v", gs.ProductionQueue)
	}
}

func Test1300NavalMissionBuildsPortOnlyAtChosenEmbarkCoast(t *testing.T) {
	gs := aiNavalMissionTestState(4)
	gs.Regions["port"].Buildings = nil
	gs.Regions["unused_port"] = &world.Region{
		ID: "unused_port", OwnerID: "ai", Neighbors: []world.RegionID{"sea_aux"}, Satisfaction: 50,
	}
	gs.Regions["sea_aux"].Neighbors = append(gs.Regions["sea_aux"].Neighbors, "unused_port")

	ctx := prepareStrategicContext(gs, "ai")
	if ctx.navalMission == nil || ctx.navalMission.EmbarkRegionID != "port" {
		t.Fatalf("orduya en yakın kıyı çıkış hattı seçilmeliydi: %+v", ctx.navalMission)
	}
	aiNavalStrategyWithStrategicContextAndSteps(gs, "ai", nil, ctx, nil)
	if len(gs.ProductionQueue) != 1 {
		t.Fatalf("önce tek liman emri bekleniyordu: %+v", gs.ProductionQueue)
	}
	order := gs.ProductionQueue[0]
	if order.Kind != aiProductionKindBuilding || order.TypeID != "port" || order.RegionID != "port" {
		t.Fatalf("liman yalnız görev çıkış kıyısında kurulmalıydı: %+v", order)
	}
}

func Test1300NavalMissionGuidesArmyFleetAndLanding(t *testing.T) {
	gs := aiNavalMissionTestState(4)
	gs.Armies["transport"] = &army.Army{
		ID: "transport", OwnerID: "ai", RegionID: "sea_aux", IsNaval: true,
		Units: []army.Unit{{TypeID: "transport", CurrentHP: 100}}, MovePoints: 3,
	}
	ctx := prepareStrategicContext(gs, "ai")

	if got := chooseBestMoveWithStrategicContext(gs, gs.Armies["field"], ctx); got != "port" {
		t.Fatalf("seçilen kara ordusu çıkış limanına yönelmeliydi, got=%s", got)
	}
	if got := chooseBestMoveWithStrategicContext(gs, gs.Armies["transport"], ctx); got != "sea_home" {
		t.Fatalf("boş transport görev çıkış denizine yönelmeliydi, got=%s", got)
	}

	gs.Armies["field"].RegionID = "port"
	gs.Armies["transport"].RegionID = "sea_home"
	if got := chooseBestMoveWithStrategicContext(gs, gs.Armies["field"], ctx); got != "sea_home" {
		t.Fatalf("kapasite hazır olduğunda kara ordusu filoya binmeliydi, got=%s", got)
	}

	delete(gs.Armies, "field")
	gs.Armies["transport"].EmbarkedUnits = []army.Unit{
		{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100},
	}
	ctx = prepareStrategicContext(gs, "ai")
	if ctx.navalMission == nil || ctx.navalMission.FleetID != "transport" {
		t.Fatalf("yüklenmiş filo sonraki turda objective görevini korumalıydı: %+v", ctx.navalMission)
	}
	if got := chooseBestMoveWithStrategicContext(gs, gs.Armies["transport"], ctx); got != "sea_target" {
		t.Fatalf("yüklenmiş filo hedef kıyı denizine gitmeliydi, got=%s", got)
	}
	gs.Armies["transport"].RegionID = "sea_target"
	if got := chooseBestMoveWithStrategicContext(gs, gs.Armies["transport"], ctx); got != "target" {
		t.Fatalf("hedef denizine ulaşan filo objective kıyısına çıkmalıydı, got=%s", got)
	}
}

func Test1300LoadedFleetRetargetsNeutralObjectiveToReachableWarCoast(t *testing.T) {
	gs := aiNavalMissionTestState(4)
	gs.Factions["neutral"] = &faction.Faction{ID: "neutral"}
	gs.Relations[faction.RelationKey("ai", "neutral")] = &faction.Relation{
		FactionA: "ai", FactionB: "neutral", Stance: faction.StancePeace,
	}
	gs.Regions["objective_target"] = &world.Region{
		ID: "objective_target", OwnerID: "neutral", Neighbors: []world.RegionID{"sea_home"}, Satisfaction: 50,
	}
	gs.Regions["sea_home"].Neighbors = append(gs.Regions["sea_home"].Neighbors, "objective_target")
	gs.Regions["war_coast"] = &world.Region{
		ID: "war_coast", OwnerID: "enemy", Neighbors: []world.RegionID{"sea_home"}, Satisfaction: 50,
	}
	gs.Regions["sea_home"].Neighbors = append(gs.Regions["sea_home"].Neighbors, "war_coast")
	gs.AIPlans["ai"] = &state.AIPlanState{
		ObjectiveID: "secure_neutral", Kind: state.AIObjectiveExpand, TargetFactionID: "neutral",
		TargetRegionIDs: []world.RegionID{"objective_target"}, StartedTurn: 1, ReassessTurn: 12,
	}
	gs.Armies["transport"] = &army.Army{
		ID: "transport", OwnerID: "ai", RegionID: "sea_home", IsNaval: true,
		Units: []army.Unit{{TypeID: "transport", CurrentHP: 100}},
		EmbarkedUnits: []army.Unit{
			{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100},
		},
		MovePoints: 3, MaxMovePoints: 3,
	}

	ctx := prepareStrategicContext(gs, "ai")
	if ctx.navalMission == nil || ctx.navalMission.TargetRegionID != "objective_target" {
		t.Fatalf("test başlangıcında barış hedefi aktif mission olmalıydı: %+v", ctx.navalMission)
	}
	if got := chooseBestMoveWithStrategicContext(gs, gs.Armies["transport"], ctx); got != "war_coast" {
		t.Fatalf("yüklü filo en yakın çıkarılabilir savaş kıyısına retarget olmalıydı, got=%s mission=%+v", got, ctx.navalMission)
	}
	if ctx.navalMission.TargetFactionID != "enemy" || ctx.navalMission.TargetRegionID != "war_coast" {
		t.Fatalf("mission savaş kıyısına güncellenmeliydi: %+v", ctx.navalMission)
	}
}

func Test1300NavalStrategyDoesNothingWithoutConcreteMission(t *testing.T) {
	gs := aiNavalMissionTestState(4)
	gs.AIPlans["ai"] = &state.AIPlanState{
		ObjectiveID: "consolidate:ai", Kind: state.AIObjectiveConsolidate,
		TargetRegionIDs: []world.RegionID{"home"}, StartedTurn: 1, ReassessTurn: 12,
	}
	gs.Regions["port"].Buildings = nil
	gs.Armies["transport"] = &army.Army{
		ID: "transport", OwnerID: "ai", RegionID: "sea_aux", IsNaval: true,
		Units: []army.Unit{{TypeID: "transport", CurrentHP: 100}}, MovePoints: 3,
	}

	ctx := prepareStrategicContext(gs, "ai")
	if ctx.navalMission != nil {
		t.Fatalf("konsolidasyon planı deniz taşıma görevi üretmemeliydi: %+v", ctx.navalMission)
	}
	aiNavalStrategyWithStrategicContextAndSteps(gs, "ai", nil, ctx, nil)
	if len(gs.ProductionQueue) != 0 {
		t.Fatalf("görev yokken liman veya transport kuyruğu açılmamalıydı: %+v", gs.ProductionQueue)
	}
	if got := chooseBestMoveWithStrategicContext(gs, gs.Armies["transport"], ctx); got != "" {
		t.Fatalf("görevsiz boş filo uzak denizde dolaşmamalıydı, got=%s", got)
	}
}

func Test1300NavalMissionSkippedWhenSafeLandRouteExists(t *testing.T) {
	gs := aiNavalMissionTestState(4)
	gs.Regions["port"].Neighbors = append(gs.Regions["port"].Neighbors, "target")
	gs.Regions["target"].Neighbors = append(gs.Regions["target"].Neighbors, "port")

	ctx := prepareStrategicContext(gs, "ai")
	if ctx.navalMission != nil {
		t.Fatalf("güvenli kara rotası olan objective gereksiz deniz görevi üretmemeliydi: %+v", ctx.navalMission)
	}
}

func Test1300NavalMissionSupportsOverseasDefenseObjective(t *testing.T) {
	gs := aiNavalMissionTestState(4)
	gs.Regions["island"] = &world.Region{
		ID: "island", OwnerID: "ai", Neighbors: []world.RegionID{"sea_target"}, Satisfaction: 50,
	}
	gs.Regions["sea_target"].Neighbors = append(gs.Regions["sea_target"].Neighbors, "island")
	gs.AIPlans["ai"] = &state.AIPlanState{
		ObjectiveID: "defend_island", Kind: state.AIObjectiveDefend, TargetFactionID: "enemy",
		TargetRegionIDs: []world.RegionID{"island"}, StartedTurn: 1, ReassessTurn: 12,
	}

	ctx := prepareStrategicContext(gs, "ai")
	if ctx.navalMission == nil || ctx.navalMission.Kind != aiNavalMissionRelief || ctx.navalMission.TargetRegionID != "island" || ctx.navalMission.TargetFactionID != "enemy" {
		t.Fatalf("denizaşırı savunma objective'i relief görevi üretmeliydi: %+v", ctx.navalMission)
	}
}
