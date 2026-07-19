package ai

import (
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiRetreatUnits(count, hp int) []army.Unit {
	units := make([]army.Unit, count)
	for index := range units {
		units[index] = army.Unit{TypeID: "inf", CurrentHP: hp}
	}
	return units
}

func aiRetreatTestState() *state.GameState {
	return &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Difficulty: 2,
		Turn:       4,
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":    {ID: "ai", AIAggressiveness: 60},
			"enemy": {ID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			"safe":       {ID: "safe", OwnerID: "ai", Neighbors: []world.RegionID{"rear"}, Satisfaction: 50},
			"rear":       {ID: "rear", OwnerID: "ai", Neighbors: []world.RegionID{"safe", "front"}, Satisfaction: 50},
			"front":      {ID: "front", OwnerID: "ai", Neighbors: []world.RegionID{"rear", "enemy_gate"}, Satisfaction: 50},
			"enemy_gate": {ID: "enemy_gate", OwnerID: "enemy", Neighbors: []world.RegionID{"front", "enemy_rear"}, Satisfaction: 50},
			"enemy_rear": {ID: "enemy_rear", OwnerID: "enemy", Neighbors: []world.RegionID{"enemy_gate"}, Satisfaction: 50},
		},
		Armies: map[army.ArmyID]*army.Army{
			"field": {
				ID: "field", OwnerID: "ai", RegionID: "front", MovePoints: 2, MaxMovePoints: 2,
				Units: aiRetreatUnits(4, army.MaxUnitHP),
			},
			"enemy": {
				ID: "enemy", OwnerID: "enemy", RegionID: "enemy_gate", MovePoints: 2, MaxMovePoints: 2,
				Units: aiRetreatUnits(1, army.MaxUnitHP),
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 10, Defense: 10, GrainUpkeep: 1},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "enemy"): {FactionA: "ai", FactionB: "enemy", Stance: faction.StanceWar},
		},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"ai": {
				ObjectiveID: "expand:enemy", Kind: state.AIObjectiveExpand, TargetFactionID: "enemy",
				TargetRegionIDs: []world.RegionID{"enemy_gate", "enemy_rear"}, StartedTurn: 1, ReassessTurn: 10,
			},
		},
	}
}

func TestRetreatStrengthThresholdIsStrictlyBelowFortyFivePercent(t *testing.T) {
	gs := aiRetreatTestState()
	armyRef := gs.Armies["field"]
	armyRef.Units = aiRetreatUnits(1, 45)
	if aiArmyStrengthBelowPercent(armyRef, gs.UnitTypes, aiRetreatStrengthPercent) {
		t.Fatal("tam yüzde 45 güç geri çekilme eşiğinin altında sayılmamalıydı")
	}
	armyRef.Units[0].CurrentHP = 44
	if !aiArmyStrengthBelowPercent(armyRef, gs.UnitTypes, aiRetreatStrengthPercent) {
		t.Fatal("yüzde 44 güç geri çekilme eşiğini tetiklemeliydi")
	}
}

func TestWornArmyRetreatsThroughFriendlyLandToSafeSupply(t *testing.T) {
	gs := aiRetreatTestState()
	gs.Armies["field"].Units = aiRetreatUnits(4, 40)

	ctx := prepareStrategicContext(gs, "ai")
	assignment := ctx.ArmyAssignments["field"]
	if assignment.Role != AIArmyRoleRetreat || assignment.AnchorRegionID != "rear" {
		t.Fatalf("yıpranmış ordu en yakın güvenli ikmal bölgesine çekilmeliydi: %+v", assignment)
	}
	if target := chooseBestMoveWithStrategicContext(gs, gs.Armies["field"], ctx); target != "rear" {
		t.Fatalf("retreat rotasının ilk adımı yalnız dost toprağa çıkmalıydı: %s", target)
	}
}

func TestHealthyArmyRetreatsWhenLocalEnemyPowerReachesOneThirtyFivePercent(t *testing.T) {
	gs := aiRetreatTestState()
	gs.Armies["field"].Units = aiRetreatUnits(2, army.MaxUnitHP)
	gs.Armies["enemy"].Units = []army.Unit{
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "heavy", CurrentHP: 100},
	}
	gs.UnitTypes["heavy"] = &army.UnitType{ID: "heavy", Category: army.CategoryInfantry, Attack: 7, Defense: 7, GrainUpkeep: 1}

	ctx := prepareStrategicContext(gs, "ai")
	assignment := ctx.ArmyAssignments["field"]
	if assignment.Role != AIArmyRoleRetreat || !strings.Contains(assignment.Reason, "135") {
		t.Fatalf("yerel düşman gücü tam yüzde 135'te retreat tetiklemeliydi: %+v", assignment)
	}
}

func TestRecoveryAnchorSkipsNearestOverloadedRegion(t *testing.T) {
	gs := aiRetreatTestState()
	gs.Armies["field"].Units = aiRetreatUnits(1, 40)
	gs.Armies["rear_stack"] = &army.Army{
		ID: "rear_stack", OwnerID: "ai", RegionID: "rear", MovePoints: 2,
		Units: aiRetreatUnits(4, army.MaxUnitHP),
	}

	ctx := prepareStrategicContext(gs, "ai")
	assignment := ctx.ArmyAssignments["field"]
	if assignment.Role != AIArmyRoleRetreat || assignment.AnchorRegionID != "safe" {
		t.Fatalf("kapasitesi dolu yakın bölge atlanıp sonraki güvenli ikmal noktası seçilmeliydi: %+v", assignment)
	}
	if target := aiRetreatNextStep(ctx, gs.Armies["field"]); target != "rear" {
		t.Fatalf("overload bölgesi hedef değil yalnız transit adımı olmalıydı: %s", target)
	}
}

func TestActiveSiegeNeedsBothSupplyOverloadAndReliefSuperiorityToWithdraw(t *testing.T) {
	tests := []struct {
		name              string
		overCapacityTurns int
		reliefUnits       int
		wantRetreat       bool
	}{
		{name: "only relief superiority", reliefUnits: 4},
		{name: "only supply overload", overCapacityTurns: 1},
		{name: "exactly one hundred fifty percent", overCapacityTurns: 1, reliefUnits: 3},
		{name: "both risks", overCapacityTurns: 1, reliefUnits: 4, wantRetreat: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gs := aiRetreatTestState()
			delete(gs.Armies, "enemy")
			gs.Armies["field"].Units = aiRetreatUnits(2, army.MaxUnitHP)
			gs.Armies["field"].OverCapacityTurns = test.overCapacityTurns
			gs.Armies["relief"] = &army.Army{
				ID: "relief", OwnerID: "enemy", RegionID: "enemy_rear", MovePoints: 2,
				Units: aiRetreatUnits(test.reliefUnits, army.MaxUnitHP),
			}
			gs.Sieges = map[world.RegionID]*state.SiegeState{
				"enemy_gate": {
					RegionID: "enemy_gate", AttackerArmyID: "field", AttackerHomeRegionID: "front",
					AttackerFactionID: "ai", FortLevel: 1,
				},
			}

			ctx := prepareStrategicContext(gs, "ai")
			gotRetreat := ctx.ArmyAssignments["field"].Role == AIArmyRoleRetreat
			if gotRetreat != test.wantRetreat {
				t.Fatalf("kuşatma retreat kararı hatalı: want=%t assignment=%+v", test.wantRetreat, ctx.ArmyAssignments["field"])
			}
		})
	}
}

func TestRiskySiegeWithdrawalClearsSiegeAndEmitsStep(t *testing.T) {
	gs := aiRetreatTestState()
	delete(gs.Armies, "enemy")
	gs.Armies["field"].Units = aiRetreatUnits(2, army.MaxUnitHP)
	gs.Armies["field"].OverCapacityTurns = 1
	gs.Armies["relief"] = &army.Army{
		ID: "relief", OwnerID: "enemy", RegionID: "enemy_rear", Units: aiRetreatUnits(4, army.MaxUnitHP),
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"enemy_gate": {
			RegionID: "enemy_gate", AttackerArmyID: "field", AttackerHomeRegionID: "front",
			AttackerFactionID: "ai", FortLevel: 1,
		},
	}

	ctx := prepareStrategicContext(gs, "ai")
	step, withdrew := executeStrategicSiegeWithdrawal(gs, gs.Armies["field"], "ai", ctx)
	if !withdrew || gs.SiegeAt("enemy_gate") != nil {
		t.Fatalf("riskli kuşatma kaldırılmalıydı: withdrew=%t siege=%+v", withdrew, gs.SiegeAt("enemy_gate"))
	}
	if step.ArmyID != "field" || step.FocusRegion != "enemy_gate" || !strings.Contains(step.Message, "geri çekiliyor") {
		t.Fatalf("kuşatma geri çekilmesi görünür turn step üretmeliydi: %+v", step)
	}
}

func TestAIStartedSiegeRecordsHomeRegionForWithdrawal(t *testing.T) {
	gs := aiRetreatTestState()
	field := gs.Armies["field"]
	aiStartSiege(gs, field, gs.Regions["enemy_gate"], nil)

	siege := gs.SiegeAt("enemy_gate")
	if siege == nil || siege.AttackerHomeRegionID != "front" {
		t.Fatalf("AI kuşatması geri çekilme evi olarak başlangıç bölgesini kaydetmeliydi: %+v", siege)
	}
}

func TestTurnStepperExecutesSiegeWithdrawalBeforeMovement(t *testing.T) {
	gs := aiRetreatTestState()
	delete(gs.Armies, "enemy")
	gs.Armies["field"].Units = aiRetreatUnits(2, army.MaxUnitHP)
	gs.Armies["field"].OverCapacityTurns = 1
	gs.Armies["relief"] = &army.Army{ID: "relief", OwnerID: "enemy", RegionID: "enemy_rear", Units: aiRetreatUnits(4, 100)}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"enemy_gate": {RegionID: "enemy_gate", AttackerArmyID: "field", AttackerFactionID: "ai", FortLevel: 1},
	}
	ctx := prepareStrategicContext(gs, "ai")
	stepper := &TurnStepper{gs: gs, fid: "ai", preludeDone: true, strategicContext: ctx}

	step, done := stepper.Step()
	if done || step.ArmyID != "field" || !strings.Contains(step.Message, "kuşatmayı kaldırdı") {
		t.Fatalf("turn stepper hareketten önce kuşatma withdrawal adımını döndürmeliydi: done=%t step=%+v", done, step)
	}
	if gs.SiegeAt("enemy_gate") != nil {
		t.Fatalf("turn stepper withdrawal adımında kuşatma kaydı kaldırılmalıydı: %+v", gs.SiegeAt("enemy_gate"))
	}
}

func TestRiskySiegeWithdrawalTransfersToRemainingArmyDeterministically(t *testing.T) {
	gs := aiRetreatTestState()
	delete(gs.Armies, "enemy")
	gs.Armies["field"].Units = aiRetreatUnits(2, army.MaxUnitHP)
	gs.Armies["field"].OverCapacityTurns = 1
	gs.Armies["relief"] = &army.Army{ID: "relief", OwnerID: "enemy", RegionID: "enemy_rear", Units: aiRetreatUnits(4, 100)}
	gs.Armies["support_b"] = &army.Army{ID: "support_b", OwnerID: "ai", RegionID: "enemy_gate", Units: aiRetreatUnits(1, 100)}
	gs.Armies["support_a"] = &army.Army{ID: "support_a", OwnerID: "ai", RegionID: "enemy_gate", Units: aiRetreatUnits(1, 100)}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"enemy_gate": {RegionID: "enemy_gate", AttackerArmyID: "field", AttackerFactionID: "ai", FortLevel: 1},
	}

	ctx := prepareStrategicContext(gs, "ai")
	step, withdrew := executeStrategicSiegeWithdrawal(gs, gs.Armies["field"], "ai", ctx)
	if !withdrew || gs.SiegeAt("enemy_gate") == nil || gs.SiegeAt("enemy_gate").AttackerArmyID != "support_a" {
		t.Fatalf("kuşatma alfabetik ilk uygun destek ordusuna devredilmeliydi: withdrew=%t siege=%+v", withdrew, gs.SiegeAt("enemy_gate"))
	}
	if !strings.Contains(step.Message, "devretti") {
		t.Fatalf("devir turn step mesajına yansımalıydı: %+v", step)
	}
}

func TestRiskySiegeIsHeldWhenNoSafeRecoveryRegionExists(t *testing.T) {
	gs := aiRetreatTestState()
	delete(gs.Armies, "enemy")
	gs.Armies["field"].Units = aiRetreatUnits(2, army.MaxUnitHP)
	gs.Armies["field"].OverCapacityTurns = 1
	gs.Armies["relief"] = &army.Army{ID: "relief", OwnerID: "enemy", RegionID: "enemy_rear", Units: aiRetreatUnits(4, 100)}
	gs.Armies["blocker"] = &army.Army{ID: "blocker", OwnerID: "enemy", RegionID: "enemy_gate", Units: aiRetreatUnits(1, 100)}
	gs.Regions["rear"].OwnerID = "enemy"
	gs.Regions["safe"].OwnerID = "enemy"
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"enemy_gate": {RegionID: "enemy_gate", AttackerArmyID: "field", AttackerFactionID: "ai", FortLevel: 1},
	}

	ctx := prepareStrategicContext(gs, "ai")
	if assignment := ctx.ArmyAssignments["field"]; assignment.Role != AIArmyRoleSiege {
		t.Fatalf("güvenli kaçış hattı yoksa kuşatma plansızca terk edilmemeliydi: %+v", assignment)
	}
}
