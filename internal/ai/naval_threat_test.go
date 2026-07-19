package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiNavalThreatRouteState() *state.GameState {
	return &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai"}, "enemy": {ID: "enemy"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "enemy"): {FactionA: "ai", FactionB: "enemy", Stance: faction.StanceWar},
		},
		Regions: map[world.RegionID]*world.Region{
			"port":   {ID: "port", OwnerID: "ai", Buildings: []string{"port"}, Neighbors: []world.RegionID{"start"}},
			"start":  {ID: "start", IsSea: true, Neighbors: []world.RegionID{"port", "short", "safe_a"}},
			"short":  {ID: "short", IsSea: true, Neighbors: []world.RegionID{"start", "target"}},
			"safe_a": {ID: "safe_a", IsSea: true, Neighbors: []world.RegionID{"start", "safe_b"}},
			"safe_b": {ID: "safe_b", IsSea: true, Neighbors: []world.RegionID{"safe_a", "target"}},
			"target": {ID: "target", IsSea: true, Neighbors: []world.RegionID{"short", "safe_b"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"enemy_fleet": {
				ID: "enemy_fleet", OwnerID: "enemy", RegionID: "short", IsNaval: true,
				Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"warship": {ID: "warship", Category: army.CategoryNavalWar, Attack: 28, Defense: 18, Morale: 60},
		},
	}
}

func Test1300NavalRoutePrefersLongerThreatFreePath(t *testing.T) {
	gs := aiNavalThreatRouteState()
	route := aiThreatAwareSeaRoute(gs, "ai", "start", "target")
	if !route.Reachable || route.FirstStep != "safe_a" || route.Hops != 3 || route.MaxThreat != 0 {
		t.Fatalf("kısa düşman hattı yerine uzun güvenli rota seçilmeliydi: %+v", route)
	}
}

func Test1300NavalThreatSnapshotMarksApproachedPort(t *testing.T) {
	gs := aiNavalThreatRouteState()
	ctx := buildStrategicContext(gs, "ai")
	buildAINavalThreatSnapshot(ctx)

	if len(ctx.NavalThreats) != 1 || ctx.NavalThreats[0].SeaRegionID != "short" || ctx.NavalThreats[0].HostilePower <= 0 {
		t.Fatalf("düşman filosu deniz tehdit snapshot'ına girmeliydi: %+v", ctx.NavalThreats)
	}
	if len(ctx.ThreatenedPortIDs) != 1 || ctx.ThreatenedPortIDs[0] != "port" {
		t.Fatalf("bir deniz adımı uzaktaki düşman görev limanını tehdit etmeliydi: %+v", ctx.ThreatenedPortIDs)
	}
}

func Test1300NavalSafetyGateRequiresOneHundredTenPercentPower(t *testing.T) {
	gs := aiNavalThreatRouteState()
	gs.Armies["enemy_fleet"].RegionID = "target"
	gs.Armies["enemy_fleet"].Units = append(gs.Armies["enemy_fleet"].Units, army.Unit{TypeID: "warship", CurrentHP: 100})
	fleet := &army.Army{
		ID: "task", OwnerID: "ai", RegionID: "start", IsNaval: true,
		Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}, {TypeID: "warship", CurrentHP: 100}},
	}
	if aiNavalFleetMeetsSafetyGate(gs, fleet, "target") {
		t.Fatal("eşit güç yüzde 110 güvenlik eşiğini geçmemeliydi")
	}
	fleet.Units = append(fleet.Units, army.Unit{TypeID: "warship", CurrentHP: 100})
	if !aiNavalFleetMeetsSafetyGate(gs, fleet, "target") {
		t.Fatal("yüzde 110 üstündeki görev filosu tehditli denize girebilmeliydi")
	}
}

func Test1300MissionFleetWaitsUntilEscortPowerIsSafe(t *testing.T) {
	gs := aiNavalMissionTestState(4)
	gs.UnitTypes["warship"] = &army.UnitType{ID: "warship", Category: army.CategoryNavalWar, Attack: 28, Defense: 18, Morale: 60}
	gs.Armies["task"] = &army.Army{
		ID: "task", OwnerID: "ai", RegionID: "sea_home", IsNaval: true,
		Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}, {TypeID: "warship", CurrentHP: 100}},
		EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}, MovePoints: 3,
	}
	delete(gs.Armies, "field")
	gs.Armies["enemy_fleet"] = &army.Army{
		ID: "enemy_fleet", OwnerID: "enemy", RegionID: "sea_target", IsNaval: true,
		Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}, {TypeID: "warship", CurrentHP: 100}},
	}

	ctx := prepareStrategicContext(gs, "ai")
	if got := chooseBestMoveWithStrategicContext(gs, gs.Armies["task"], ctx); got != "" {
		t.Fatalf("yetersiz escort gücündeki yüklü filo tehdit önünde beklemeliydi, got=%s", got)
	}
	gs.Armies["task"].Units = append(gs.Armies["task"].Units,
		army.Unit{TypeID: "warship", CurrentHP: 100}, army.Unit{TypeID: "warship", CurrentHP: 100})
	ctx = prepareStrategicContext(gs, "ai")
	if got := chooseBestMoveWithStrategicContext(gs, gs.Armies["task"], ctx); got != "sea_target" {
		t.Fatalf("güvenlik eşiğini geçen görev filosu hedef denizine ilerlemeliydi, got=%s", got)
	}
}

func Test1300ThreatenedMissionQueuesEscortToSafetyRequirement(t *testing.T) {
	gs := aiNavalMissionTestState(4)
	gs.UnitTypes["transport"].Attack = 4
	gs.UnitTypes["transport"].Defense = 6
	gs.UnitTypes["transport"].Morale = 30
	gs.Factions["ai"].Research.Completed = map[string]bool{"naval_doctrine": true}
	gs.UnitTypes["warship"] = &army.UnitType{
		ID: "warship", Category: army.CategoryNavalWar, Attack: 28, Defense: 18, Morale: 60,
		RequiredBldg: "port", RequiredBldgLevel: 3, RequiredTech: "naval_doctrine",
		GoldCost: 100, TimberCost: 10, TurnsRequired: 2,
	}
	gs.Armies["transport"] = &army.Army{
		ID: "transport", OwnerID: "ai", RegionID: "sea_home", IsNaval: true,
		Units: []army.Unit{{TypeID: "transport", CurrentHP: 100}}, MovePoints: 3,
	}
	gs.Armies["enemy_fleet"] = &army.Army{
		ID: "enemy_fleet", OwnerID: "enemy", RegionID: "sea_target", IsNaval: true,
		Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}, {TypeID: "warship", CurrentHP: 100}},
	}

	ctx := prepareStrategicContext(gs, "ai")
	if ctx.navalMission == nil || ctx.navalMission.RouteThreatPower <= 0 {
		t.Fatalf("görev rotası düşman filo gücünü taşımalıydı: %+v", ctx.navalMission)
	}
	aiNavalStrategyWithStrategicContextAndSteps(gs, "ai", nil, ctx, nil)
	if len(gs.ProductionQueue) != 2 {
		t.Fatalf("yüzde 110 eşiğine ulaşmak için iki escort emri bekleniyordu: %+v", gs.ProductionQueue)
	}
	for _, order := range gs.ProductionQueue {
		if order.TypeID != "warship" || order.RegionID != "port" {
			t.Fatalf("escortlar tehdit edilen görev limanında üretilmeliydi: %+v", gs.ProductionQueue)
		}
	}
}
