package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiRallyTestState() *state.GameState {
	gs := aiFrontTestState()
	gs.Turn = 2
	gs.Regions["capital"].Neighbors = []world.RegionID{"rear", "front"}
	gs.Regions["enemy_border"].Neighbors = []world.RegionID{"front", "enemy_rear"}
	gs.Armies["strong"].RegionID = "front"
	gs.Armies["middle"] = &army.Army{
		ID: "middle", OwnerID: "ai", RegionID: "rear", MovePoints: 2,
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
		},
	}
	return gs
}

func TestRallyStatePersistsSafeBorderAndHoldsGatheredArmy(t *testing.T) {
	gs := aiRallyTestState()
	ctx := prepareStrategicContext(gs, "ai")
	plan := gs.AIPlans[faction.FactionID("ai")]

	if plan.RallyRegionID != "front" || plan.RallyDeadlineTurn != gs.Turn+aiRallyMaxWaitTurns {
		t.Fatalf("güvenli hedef sınırı üç turluk rally olmalıydı: %+v", plan)
	}
	if !ctx.RallyActive || ctx.RallyRegionID != "front" || ctx.RallyRequiredPower <= 0 {
		t.Fatalf("rally hazırlığı runtime context'e yansımadı: %+v", ctx)
	}
	if assignment := ctx.ArmyAssignments["strong"]; !assignment.Rallying || assignment.AnchorRegionID != "front" {
		t.Fatalf("rally noktasındaki güçlü ordu beklemeliydi: %+v", assignment)
	}
	if target := chooseBestMoveWithStrategicContext(gs, gs.Armies["strong"], ctx); target != "" {
		t.Fatalf("rally noktasındaki ordu hazırlık bitmeden saldırmamalıydı: %s", target)
	}
	if target := chooseBestMoveWithStrategicContext(gs, gs.Armies["middle"], ctx); target != "capital" {
		t.Fatalf("ikinci hücum ordusu rally noktasına yaklaşmalıydı: %s", target)
	}
	if aiStrategicWarReady(ctx, "enemy") {
		t.Fatal("aktif rally tamamlanmadan proaktif savaş açılmamalıydı")
	}
}

func TestRallyCompletesWhenTwoArmiesMeetPowerRequirement(t *testing.T) {
	gs := aiRallyTestState()
	first := prepareStrategicContext(gs, "ai")
	gs.Armies["middle"].RegionID = first.RallyRegionID

	ready := prepareStrategicContext(gs, "ai")
	plan := gs.AIPlans["ai"]
	if ready.RallyActive || !ready.RallyReady || plan.RallyDeadlineTurn != gs.Turn {
		t.Fatalf("iki ordu ve yeterli güç rally hazırlığını tamamlamalıydı: context=%+v plan=%+v", ready, plan)
	}
	if ready.ArmyAssignments["strong"].Rallying || ready.ArmyAssignments["middle"].Rallying {
		t.Fatalf("hazır ordular objective hedefine serbest bırakılmalıydı: %+v", ready.ArmyAssignments)
	}
	if !aiStrategicWarReady(ready, "enemy") {
		t.Fatal("tamamlanmış rally savaş hazırlığı kapısından geçmeliydi")
	}
}

func TestRallyDeadlineReleasesIncompleteForceAfterThreeTurns(t *testing.T) {
	gs := aiRallyTestState()
	first := prepareStrategicContext(gs, "ai")
	deadline := first.RallyDeadlineTurn
	gs.Turn = deadline

	released := prepareStrategicContext(gs, "ai")
	if released.RallyActive || !released.RallyReady || gs.AIPlans["ai"].RallyDeadlineTurn != deadline {
		t.Fatalf("üç tur sonunda eksik kuvvet de serbest bırakılmalıydı: context=%+v plan=%+v", released, gs.AIPlans["ai"])
	}
}

func TestRallyIsNotCreatedWithoutTwoOffensiveArmies(t *testing.T) {
	gs := aiRallyTestState()
	delete(gs.Armies, "middle")

	ctx := prepareStrategicContext(gs, "ai")
	if ctx.RallyActive || gs.AIPlans["ai"].RallyRegionID != "" {
		t.Fatalf("tek hücum ordusu gereksiz rally ile bekletilmemeliydi: context=%+v plan=%+v", ctx, gs.AIPlans["ai"])
	}
}

func TestRallyFollowsActiveWarInsteadOfPeaceTimePlanTarget(t *testing.T) {
	gs := aiRallyTestState()
	gs.Factions["war_enemy"] = &faction.Faction{ID: "war_enemy"}
	gs.Regions["war_border"] = &world.Region{ID: "war_border", OwnerID: "war_enemy", Neighbors: []world.RegionID{"rear"}}
	gs.Regions["rear"].Neighbors = append(gs.Regions["rear"].Neighbors, "war_border")
	gs.Relations[faction.RelationKey("ai", "war_enemy")] = &faction.Relation{FactionA: "ai", FactionB: "war_enemy", Stance: faction.StanceWar}

	ctx := prepareStrategicContext(gs, "ai")
	plan := gs.AIPlans["ai"]
	if plan.RallyRegionID != "rear" || !ctx.RallyActive {
		t.Fatalf("rally barıştaki plan hedefi yerine aktif savaş sınırına taşınmalıydı: context=%+v plan=%+v", ctx, plan)
	}
	if got := ctx.ArmyAssignments["strong"]; got.FrontFactionID != "war_enemy" || !got.Rallying || got.AnchorRegionID != "rear" {
		t.Fatalf("hücum rolü aktif savaş rally noktasına bağlanmalıydı: %+v", got)
	}
}

func TestUnsafeRallyRegionIsCleared(t *testing.T) {
	gs := aiRallyTestState()
	first := prepareStrategicContext(gs, "ai")
	if first.RallyRegionID != "front" {
		t.Fatalf("test önkoşulu rally bölgesini üretmedi: %+v", first)
	}
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"front": {RegionID: "front", AttackerArmyID: "enemy", AttackerFactionID: "enemy"},
	}

	unsafe := prepareStrategicContext(gs, "ai")
	if unsafe.RallyRegionID != "" || gs.AIPlans["ai"].RallyRegionID != "" {
		t.Fatalf("kuşatma altındaki rally iptal edilmeliydi: context=%+v plan=%+v", unsafe, gs.AIPlans["ai"])
	}
}
