package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiFrontTestState() *state.GameState {
	units := func(count int) []army.Unit {
		result := make([]army.Unit, count)
		for index := range result {
			result[index] = army.Unit{TypeID: "inf", CurrentHP: 100}
		}
		return result
	}
	return &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Difficulty: 2,
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":    {ID: "ai", CapitalSettlementID: "capital_city", AIAggressiveness: 60, Gold: 500, Grain: 500},
			"enemy": {ID: "enemy"},
			"other": {ID: "other"},
		},
		Regions: map[world.RegionID]*world.Region{
			"rear": {
				ID: "rear", OwnerID: "ai", Neighbors: []world.RegionID{"capital"}, Satisfaction: 50,
			},
			"capital": {
				ID: "capital", OwnerID: "ai", Neighbors: []world.RegionID{"rear", "front", "enemy_border"},
				Satisfaction: 50,
				Settlements:  []world.Settlement{{ID: "capital_city", Type: world.SettlementCity}},
			},
			"front": {
				ID: "front", OwnerID: "ai", Neighbors: []world.RegionID{"capital", "enemy_border"}, Satisfaction: 50,
			},
			"enemy_border": {
				ID: "enemy_border", OwnerID: "enemy", Neighbors: []world.RegionID{"capital", "front", "enemy_rear"}, Satisfaction: 50,
			},
			"enemy_rear": {
				ID: "enemy_rear", OwnerID: "enemy", Neighbors: []world.RegionID{"enemy_border"}, Satisfaction: 50,
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"strong": {ID: "strong", OwnerID: "ai", RegionID: "front", Units: units(7), MovePoints: 2},
			"weak":   {ID: "weak", OwnerID: "ai", RegionID: "rear", Units: units(3), MovePoints: 2},
			"enemy":  {ID: "enemy", OwnerID: "enemy", RegionID: "enemy_border", Units: units(2), MovePoints: 2},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 10, Defense: 10, Morale: 50},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "enemy"): {FactionA: "ai", FactionB: "enemy", Stance: faction.StancePeace},
		},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"ai": {
				ObjectiveID: "expand:enemy", Kind: state.AIObjectiveExpand, TargetFactionID: "enemy",
				TargetRegionIDs: []world.RegionID{"enemy_border", "enemy_rear"}, StartedTurn: 1, ReassessTurn: 9,
			},
		},
		AIDifficultyPolicy: scenario.AIDifficultyPolicy{Levels: map[string]scenario.AIDifficultyLevel{
			"2": {
				PlanHorizonTurns: 6, PlanTargetRegionLimit: 4, PathSearchDepth: 8, PlanMoveBonusPercent: 100,
				ProactiveWar: true, WarThreshold: 70, MinAttackPowerPercent: 115, WarCadenceTurns: 10, MaxConcurrentWars: 1,
			},
		}},
	}
}

func TestDynamicReserveKeepsWeakStackBackAndStrongStackOnObjective(t *testing.T) {
	gs := aiFrontTestState()
	ctx := prepareStrategicContext(gs, "ai")

	if ctx.ReservePercent != 15 || ctx.ReserveTargetPower <= 0 || ctx.ReserveAssignedPower < ctx.ReserveTargetPower {
		t.Fatalf("barış dönemi yüzde 15 rezerv üretmeliydi: percent=%d target=%d assigned=%d", ctx.ReservePercent, ctx.ReserveTargetPower, ctx.ReserveAssignedPower)
	}
	if got := ctx.ArmyAssignments["weak"]; got.Role != AIArmyRoleReserve || got.AnchorRegionID != "capital" {
		t.Fatalf("zayıf stack başkent rezervi olmalıydı: %+v", got)
	}
	if got := ctx.ArmyAssignments["strong"]; got.Role != AIArmyRoleAssault || got.AnchorRegionID != "enemy_border" {
		t.Fatalf("güçlü stack objective hücumunda kalmalıydı: %+v", got)
	}
}

func TestStrategicWarReadyUsesBorderForceDuringDefensivePlan(t *testing.T) {
	gs := aiFrontTestState()
	gs.AIPlans["ai"] = &state.AIPlanState{
		ObjectiveID:     "hold_frontier",
		Kind:            state.AIObjectiveDefend,
		TargetFactionID: "enemy",
		TargetRegionIDs: []world.RegionID{"capital"},
		StartedTurn:     1,
		ReassessTurn:    20,
	}

	ctx := prepareStrategicContext(gs, "ai")
	if !aiStrategicWarReady(ctx, "enemy") {
		t.Fatalf("düşük tehditli savunma planında sınır kuvveti fırsat savaşına hazır olmalıydı: %+v", ctx)
	}
}

func TestAIDiagnosticSnapshotExposesFrontTargetAndRoles(t *testing.T) {
	gs := aiFrontTestState()
	gs.Turn = 30
	gs.Relations[faction.RelationKey("ai", "enemy")].Stance = faction.StanceWar
	gs.BeginWarLedger("ai", "enemy")

	snapshot := BuildAIDiagnosticSnapshot(gs, "ai")
	if snapshot.PlanTargetFactionID != "enemy" || len(snapshot.Fronts) != 1 {
		t.Fatalf("diagnostic snapshot plan/cephe bilgisini taşımalıydı: %+v", snapshot)
	}
	if snapshot.Fronts[0].TargetRegionID == "" {
		t.Fatalf("diagnostic snapshot aktif cephe hedefini taşımalıydı: %+v", snapshot.Fronts[0])
	}
	if snapshot.ArmyRoleCounts[AIArmyRoleAssault] == 0 && snapshot.ArmyRoleCounts[AIArmyRoleSiege] == 0 && snapshot.ArmyRoleCounts[AIArmyRoleDefense] == 0 {
		t.Fatalf("diagnostic snapshot ordu rollerini taşımalıydı: %+v", snapshot.ArmyRoleCounts)
	}
}

func TestAIDiagnosticSnapshotExposesNavalState(t *testing.T) {
	gs := aiMerchantTradeTestState()
	fleet := gs.Armies["merchant"]
	fleet.DockedRegionID = "venice"
	fleet.DockedSettlementID = "venice_port"

	snapshot := BuildAIDiagnosticSnapshot(gs, "venice")
	if snapshot.NavalFleetCount != 1 || snapshot.NavalDockedFleetCount != 1 {
		t.Fatalf("donanma sayıları diagnostic snapshot'a taşınmalıydı: %+v", snapshot)
	}
	if snapshot.NavalMissionKind != "patrol" {
		t.Fatalf("görevsiz filonun patrol görünmesi bekleniyordu: %+v", snapshot)
	}
	if len(snapshot.BlockReasons) == 0 || snapshot.BlockReasons[len(snapshot.BlockReasons)-1] != "donanmanın bir kısmı limanda; ilk hareket denize çıkış" {
		t.Fatalf("dock teşhis engeli görünmeliydi: %+v", snapshot.BlockReasons)
	}
}

func TestFrontTargetPrefersStrategicValueOverFirstRegion(t *testing.T) {
	gs := aiFrontTestState()
	gs.Regions["capital"].Neighbors = append(gs.Regions["capital"].Neighbors, "enemy_rear")
	gs.Regions["enemy_border"].BaseGrainOutput = 1
	gs.Regions["enemy_rear"].BaseGrainOutput = 100

	ctx := prepareStrategicContext(gs, "ai")
	for _, front := range ctx.Fronts {
		if front.EnemyFactionID == "enemy" {
			if front.TargetRegionID != "enemy_rear" {
				t.Fatalf("cephe ilk bölgeye değil stratejik değeri yüksek hedefe yönelmeli: %+v", front)
			}
			return
		}
	}
	t.Fatal("enemy cephesi bulunamadı")
}

func TestWarFrontTargetStaysLockedForShortWindow(t *testing.T) {
	gs := aiFrontTestState()
	gs.Regions["capital"].Neighbors = append(gs.Regions["capital"].Neighbors, "enemy_rear")
	gs.Regions["enemy_border"].BaseGrainOutput = 1
	gs.Regions["enemy_rear"].BaseGrainOutput = 100
	gs.Relations[faction.RelationKey("ai", "enemy")].Stance = faction.StanceWar
	gs.Turn = 20
	ledger := gs.BeginWarLedger("ai", "enemy")
	ledger.StartedTurn = 1
	ledger.TargetRegionID = "enemy_border"
	ledger.TargetLockedTurn = 19

	ctx := prepareStrategicContext(gs, "ai")
	for _, front := range ctx.Fronts {
		if front.EnemyFactionID == "enemy" && front.TargetRegionID != "enemy_border" {
			t.Fatalf("kısa hedef kilidi stratejik skorla hemen değişmemeli: %+v", front)
		}
	}
}

func TestSameRealmWarFrontSharesTargetAndFriendlyPower(t *testing.T) {
	gs := aiFrontTestState()
	gs.Turn = 20
	gs.Factions["vassal"] = &faction.Faction{ID: "vassal", OverlordID: "ai"}
	gs.Relations[faction.RelationKey("ai", "enemy")].Stance = faction.StanceWar
	gs.Relations[faction.RelationKey("vassal", "enemy")] = &faction.Relation{
		FactionA: "vassal", FactionB: "enemy", Stance: faction.StanceWar,
	}
	gs.Armies["vassal_field"] = &army.Army{
		ID: "vassal_field", OwnerID: "vassal", RegionID: "front", Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100},
		},
	}
	gs.Regions["enemy_border"].BaseGrainOutput = 1
	gs.Regions["enemy_rear"].BaseGrainOutput = 100
	gs.Regions["capital"].Neighbors = append(gs.Regions["capital"].Neighbors, "enemy_rear")
	vassalLedger := gs.BeginWarLedger("vassal", "enemy")
	vassalLedger.TargetRegionID = "enemy_rear"
	vassalLedger.TargetLockedTurn = 19

	ctx := prepareStrategicContext(gs, "ai")
	for _, front := range ctx.Fronts {
		if front.EnemyFactionID != "enemy" {
			continue
		}
		if front.TargetRegionID != "enemy_rear" {
			t.Fatalf("aynı realm vassalının hedef kilidi overlord cephesine taşınmalıydı: %+v", front)
		}
		if front.FriendlyPower < 135 {
			t.Fatalf("vassal saha gücü ortak cephe gücüne eklenmeliydi: %+v", front)
		}
		return
	}
	t.Fatal("enemy cephesi bulunamadı")
}

func TestNavalAssignmentsUseTransportAndEscortRoles(t *testing.T) {
	gs := &state.GameState{
		Armies: map[army.ArmyID]*army.Army{
			"transport": {ID: "transport", OwnerID: "ai", RegionID: "sea_home", IsNaval: true, Units: []army.Unit{{TypeID: "transport_ship"}}},
			"escort":    {ID: "escort", OwnerID: "ai", RegionID: "sea_home", IsNaval: true, Units: []army.Unit{{TypeID: "warship"}}},
			"merchant":  {ID: "merchant", OwnerID: "ai", RegionID: "sea_home", IsNaval: true, Units: []army.Unit{{TypeID: "merchant_ship"}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport_ship": {ID: "transport_ship", Category: army.CategoryNavalTrans, CarryCapacity: 5},
			"warship":        {ID: "warship", Category: army.CategoryNavalWar},
			"merchant_ship":  {ID: "merchant_ship", Category: army.CategoryNavalTrade},
		},
	}
	ctx := &StrategicContext{gs: gs, FactionID: "ai", ArmyAssignments: make(map[army.ArmyID]AIArmyAssignment)}
	assignAINavalRoles(ctx)
	if got := ctx.ArmyAssignments["transport"]; got.Role != AIArmyRoleTransport || got.AnchorRegionID != "sea_home" {
		t.Fatalf("transport filosuna transport rolü verilmedi: %+v", got)
	}
	if got := ctx.ArmyAssignments["escort"]; got.Role != AIArmyRoleEscort || got.AnchorRegionID != "sea_home" {
		t.Fatalf("savaş filosuna escort rolü verilmedi: %+v", got)
	}
	if _, assigned := ctx.ArmyAssignments["merchant"]; assigned {
		t.Fatal("merchant filosu kara görev rolüne zorlanmamalıydı")
	}
}

func TestCapitalWarThreatRaisesReserveToThirtyPercent(t *testing.T) {
	gs := aiFrontTestState()
	gs.Relations[faction.RelationKey("ai", "enemy")].Stance = faction.StanceWar
	ctx := prepareStrategicContext(gs, "ai")

	if !ctx.CriticalThreat || ctx.ReservePercent != 30 {
		t.Fatalf("başkent savaş cephesi rezervi yüzde 30'a çıkarmalıydı: critical=%t percent=%d fronts=%+v", ctx.CriticalThreat, ctx.ReservePercent, ctx.Fronts)
	}
	if len(ctx.Fronts) == 0 || !ctx.Fronts[0].CapitalThreat {
		t.Fatalf("başkent tehdidi cephe snapshot'ına yazılmadı: %+v", ctx.Fronts)
	}
}

func TestMultipleActiveFrontsRaiseReserveWithoutCriticalThreat(t *testing.T) {
	gs := aiFrontTestState()
	ctx := &StrategicContext{
		gs:        gs,
		FactionID: "ai",
		Fronts: []AIFront{
			{EnemyFactionID: "enemy", AtWar: true, ThreatScore: -20},
			{EnemyFactionID: "other", AtWar: true, ThreatScore: -10},
		},
	}
	if got := aiReservePercentForFrontRisk(ctx); got != 25 {
		t.Fatalf("iki aktif cephede yedek oranı yüzde 25 olmalıydı: got=%d", got)
	}

	ctx.Fronts[0].AtWar = false
	ctx.Fronts[1].AtWar = false
	if got := aiReservePercentForFrontRisk(ctx); got != 15 {
		t.Fatalf("savaşsız durumda temel yedek oranı yüzde 15 olmalıydı: got=%d", got)
	}
}

func TestSingleArmyFactionDoesNotFreezeItsOnlyFieldArmy(t *testing.T) {
	gs := aiFrontTestState()
	delete(gs.Armies, "weak")
	ctx := prepareStrategicContext(gs, "ai")

	if ctx.ReserveTargetPower != 0 || ctx.ArmyAssignments["strong"].Role != AIArmyRoleAssault {
		t.Fatalf("tek saha ordusu rezervde dondurulmamalıydı: target=%d assignment=%+v", ctx.ReserveTargetPower, ctx.ArmyAssignments["strong"])
	}
}

func TestReserveRoleMovesTowardCapitalAndWillNotInvade(t *testing.T) {
	gs := aiFrontTestState()
	ctx := prepareStrategicContext(gs, "ai")
	reserve := gs.Armies["weak"]

	if target := chooseBestMoveWithStrategicContext(gs, reserve, ctx); target != "capital" {
		t.Fatalf("rezerv başkente yaklaşmalıydı: got=%s assignment=%+v", target, ctx.ArmyAssignments[reserve.ID])
	}
	reserve.RegionID = "capital"
	if target := chooseBestMoveWithStrategicContext(gs, reserve, ctx); target != "" {
		t.Fatalf("başkente ulaşan rezerv objective için sınırı geçmemeliydi: %s", target)
	}
}

func TestActiveSiegeAndFriendlyReliefReceiveDedicatedRoles(t *testing.T) {
	t.Run("active siege", func(t *testing.T) {
		gs := aiFrontTestState()
		gs.Sieges = map[world.RegionID]*state.SiegeState{
			"enemy_border": {RegionID: "enemy_border", AttackerArmyID: "strong", AttackerFactionID: "ai"},
		}
		ctx := prepareStrategicContext(gs, "ai")
		if got := ctx.ArmyAssignments["strong"]; got.Role != AIArmyRoleSiege || got.AnchorRegionID != "enemy_border" {
			t.Fatalf("aktif kuşatma ordusu siege rolünü korumalıydı: %+v", got)
		}
	})

	t.Run("friendly relief", func(t *testing.T) {
		gs := aiFrontTestState()
		gs.Armies["enemy"].RegionID = "capital"
		gs.Sieges = map[world.RegionID]*state.SiegeState{
			"capital": {RegionID: "capital", AttackerArmyID: "enemy", AttackerFactionID: "enemy"},
		}
		gs.Relations[faction.RelationKey("ai", "enemy")].Stance = faction.StanceWar
		ctx := prepareStrategicContext(gs, "ai")
		if got := ctx.ArmyAssignments["weak"]; got.Role != AIArmyRoleRelief || got.AnchorRegionID != "capital" {
			t.Fatalf("en yakın uygun ordu relief rolü almalıydı: %+v", got)
		}
	})
}

func TestStrategicWarReadinessBlocksNewFrontDuringCriticalThreat(t *testing.T) {
	gs := aiFrontTestState()
	peace := prepareStrategicContext(gs, "ai")
	if !aiStrategicWarReady(peace, "enemy") {
		t.Fatalf("rezervi ve hücum gücü hazır barış state'i savaş hazırlığından geçmeliydi: %+v", peace.ArmyAssignments)
	}

	gs.Relations[faction.RelationKey("ai", "enemy")].Stance = faction.StanceWar
	threatened := prepareStrategicContext(gs, "ai")
	if aiStrategicWarReady(threatened, "other") {
		t.Fatal("başkent kritik tehdit altındayken yeni cephe açılmamalıydı")
	}
}

func TestAssaultRoleFinishesActiveWarBeforePeaceTimeObjective(t *testing.T) {
	gs := aiFrontTestState()
	gs.Factions["war_enemy"] = &faction.Faction{ID: "war_enemy"}
	gs.Regions["war_border"] = &world.Region{ID: "war_border", OwnerID: "war_enemy", Neighbors: []world.RegionID{"rear"}}
	gs.Regions["rear"].Neighbors = append(gs.Regions["rear"].Neighbors, "war_border")
	gs.Relations[faction.RelationKey("ai", "war_enemy")] = &faction.Relation{FactionA: "ai", FactionB: "war_enemy", Stance: faction.StanceWar}

	ctx := prepareStrategicContext(gs, "ai")
	assignment := ctx.ArmyAssignments["strong"]
	if assignment.Role != AIArmyRoleAssault || assignment.FrontFactionID != "war_enemy" || assignment.AnchorRegionID != "war_border" {
		t.Fatalf("hücum ordusu barıştaki objective yerine aktif savaşı sonuçlandırmalıydı: %+v", assignment)
	}
}

func TestAssaultRoleSkipsAlreadyConqueredPriorityRegion(t *testing.T) {
	gs := aiFrontTestState()
	gs.Regions["enemy_border"].OwnerID = "ai"
	gs.Relations[faction.RelationKey("ai", "enemy")].Stance = faction.StanceWar

	ctx := prepareStrategicContext(gs, "ai")
	assignment := ctx.ArmyAssignments["strong"]
	if assignment.Role != AIArmyRoleAssault || assignment.AnchorRegionID != "enemy_rear" {
		t.Fatalf("ele geçirilmiş ilk hedef atlanıp düşmanda kalan bölge seçilmeliydi: %+v", assignment)
	}
}

func TestActiveWarOverridesDefensivePlanForNonCriticalFront(t *testing.T) {
	gs := aiFrontTestState()
	gs.Regions["capital"].Neighbors = []world.RegionID{"rear", "front"}
	gs.Regions["enemy_border"].Neighbors = []world.RegionID{"front", "enemy_rear"}
	delete(gs.Armies, "enemy")
	gs.AIPlans["ai"] = &state.AIPlanState{
		ObjectiveID:     "hold_frontier",
		Kind:            state.AIObjectiveDefend,
		TargetFactionID: "enemy",
		TargetRegionIDs: []world.RegionID{"rear"},
		ReassessTurn:    20,
	}
	gs.Relations[faction.RelationKey("ai", "enemy")].Stance = faction.StanceWar
	gs.Turn = 20
	ledger := gs.BeginWarLedger("ai", "enemy")
	ledger.StartedTurn = 1

	ctx := prepareStrategicContext(gs, "ai")
	assignment := ctx.ArmyAssignments["strong"]
	if assignment.Role != AIArmyRoleAssault || assignment.FrontFactionID != "enemy" || assignment.AnchorRegionID != "enemy_border" {
		t.Fatalf("aktif kritik olmayan savaş cephesi savunma planına rağmen hücum almalıydı: %+v", assignment)
	}
}
