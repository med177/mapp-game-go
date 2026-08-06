package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func strategicPlanTestState() *state.GameState {
	return &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Turn:       4,
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {
				ID:                 "ottoman",
				AIAggressiveness:   62,
				AIExpansionTargets: []faction.FactionID{"east_rome", "germiyan_bey"},
			},
			"east_rome":    {ID: "east_rome"},
			"germiyan_bey": {ID: "germiyan_bey"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ottoman_home": {
				ID:           "ottoman_home",
				OwnerID:      "ottoman",
				Neighbors:    []world.RegionID{"germiyan_border"},
				Satisfaction: 50,
			},
			"germiyan_border": {
				ID:           "germiyan_border",
				OwnerID:      "germiyan_bey",
				Neighbors:    []world.RegionID{"ottoman_home"},
				TaxRate:      60,
				Satisfaction: 50,
			},
			"constantinople": {
				ID:           "constantinople",
				OwnerID:      "east_rome",
				TaxRate:      80,
				Satisfaction: 50,
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ottoman_army": {
				ID:       "ottoman_army",
				OwnerID:  "ottoman",
				RegionID: "ottoman_home",
				Units:    []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
			},
			"east_army": {
				ID:       "east_army",
				OwnerID:  "east_rome",
				RegionID: "constantinople",
				Units:    []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 10, Defense: 10, Morale: 50},
		},
		Relations: map[string]*faction.Relation{},
	}
}

func TestEnsureStrategicPlanPersistsIntentUntilReassessment(t *testing.T) {
	gs := strategicPlanTestState()
	ctx := buildStrategicContext(gs, "ottoman")

	plan := ensureStrategicPlan(gs, "ottoman", ctx)
	if plan == nil {
		t.Fatal("1300 AI planı üretilmedi")
	}
	if plan.Kind != state.AIObjectiveExpand || plan.TargetFactionID != "germiyan_bey" {
		t.Fatalf("sınırdaki zayıf genişleme hedefi seçilmeliydi: %+v", plan)
	}
	if len(plan.TargetRegionIDs) == 0 || plan.TargetRegionIDs[0] != "germiyan_border" {
		t.Fatalf("hedef bölge önceliği yanlış: %+v", plan.TargetRegionIDs)
	}
	if plan.StartedTurn != 4 || plan.ReassessTurn != 10 || plan.Commitment != 62 || plan.Reason == "" {
		t.Fatalf("kalıcı plan metadata'sı eksik: %+v", plan)
	}

	gs.Turn = 7
	kept := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if kept != plan {
		t.Fatalf("yeniden değerlendirme turundan önce plan korunmalı: old=%p new=%p", plan, kept)
	}
}

func TestEnsureStrategicPlanReplacesInvalidTarget(t *testing.T) {
	gs := strategicPlanTestState()
	first := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	gs.Turn = first.ReassessTurn
	gs.Factions["germiyan_bey"].IsEliminated = true
	gs.Regions["germiyan_border"].OwnerID = "ottoman"

	replaced := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if replaced == first {
		t.Fatal("geçersiz hedefte plan yeniden üretilmeliydi")
	}
	if replaced.TargetFactionID != "east_rome" || replaced.Kind != state.AIObjectiveExpand {
		t.Fatalf("kalan genişleme hedefi seçilmedi: %+v", replaced)
	}
	if replaced.StartedTurn != gs.Turn {
		t.Fatalf("yenilenen plan başlangıç turu yanlış: %+v", replaced)
	}
}

func TestScenarioObjectiveTargetsCurrentClaimOwner(t *testing.T) {
	gs := strategicPlanTestState()
	gs.AIStrategies = map[string]scenario.AIFactionStrategy{
		"ottoman": {
			FactionID: "ottoman",
			Objectives: []scenario.AIObjectiveDef{{
				ID:                "recover_constantinople",
				Kind:              "expand",
				TerritorialClaims: []scenario.AITerritorialClaimDef{{RegionID: "constantinople", Value: 100}},
				Priority:          100,
			}},
		},
	}
	first := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if first == nil || first.TargetFactionID != "east_rome" {
		t.Fatalf("ilk claim sahibi hedefi yanlış: %+v", first)
	}

	gs.Regions["constantinople"].OwnerID = "germiyan_bey"
	gs.Turn = first.ReassessTurn
	retargeted := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if retargeted == nil || retargeted.TargetFactionID != "germiyan_bey" {
		t.Fatalf("claim bölgesi el değiştirince hedef yeni sahibine taşınmadı: %+v", retargeted)
	}
}

func TestEnsureStrategicPlanOnlyRunsFor1300Scenario(t *testing.T) {
	gs := strategicPlanTestState()
	gs.ScenarioID = "other_scenario"

	if plan := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman")); plan != nil {
		t.Fatalf("1300 dışı senaryoya plan yazılmamalı: %+v", plan)
	}
	if len(gs.AIPlans) != 0 {
		t.Fatalf("1300 dışı senaryonun state'i değişmemeli: %+v", gs.AIPlans)
	}
}

func TestEnsureStrategicPlanUsesVictoryConditionsInAnyScenario(t *testing.T) {
	gs := strategicPlanTestState()
	gs.ScenarioID = "1455_wars_of_the_roses"
	gs.ScenarioVictories = []scenario.VictoryOptionDef{
		{ID: "general_hegemony", Type: "domination", RequiredRegions: []string{"germiyan_border"}},
	}

	plan := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if plan == nil || plan.ObjectiveID != "victory:general_hegemony" || plan.TargetFactionID != "germiyan_bey" {
		t.Fatalf("zafer koşulu olan her senaryoda AI planı üretilmeliydi: %+v", plan)
	}
}

func TestTurnPreludeReturnsVictoryStrategicContextInAnyScenario(t *testing.T) {
	gs := strategicPlanTestState()
	gs.ScenarioID = "1455_wars_of_the_roses"
	gs.ScenarioVictories = []scenario.VictoryOptionDef{
		{ID: "general_hegemony", Type: "domination", RequiredRegions: []string{"germiyan_border"}},
	}

	ctx := runTurnPrelude(gs, "ottoman", nil)
	if ctx == nil || gs.AIPlans["ottoman"] == nil || gs.AIPlans["ottoman"].ObjectiveID != "victory:general_hegemony" {
		t.Fatalf("zafer planı 1300 dışı senaryoda AI hareket context'ine taşınmalıydı: context=%+v plan=%+v", ctx, gs.AIPlans["ottoman"])
	}
}

func TestScenarioProfileCreatesSoftHistoricalObjectivePlan(t *testing.T) {
	gs := strategicPlanTestState()
	gs.Year = 1300
	gs.AIStrategies = map[string]scenario.AIFactionStrategy{
		"ottoman": {
			FactionID: "ottoman",
			Profile:   "frontier_expansion",
			Objectives: []scenario.AIObjectiveDef{
				{ID: "secure_bithynia", Kind: "expand", TargetFactions: []string{"east_rome"}, TargetRegions: []string{"constantinople"}, Priority: 100, Commitment: 74},
				{ID: "unite_anatolian_beyliks", Kind: "expand", TargetFactions: []string{"germiyan_bey"}, TargetRegions: []string{"germiyan_border"}, Priority: 70, AllowVassalization: true},
			},
		},
	}

	plan := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if plan == nil || plan.ObjectiveID != "secure_bithynia" || plan.TargetFactionID != "east_rome" {
		t.Fatalf("yüksek öncelikli tarihsel yön seçilmedi: %+v", plan)
	}
	if plan.Commitment != 74 {
		t.Fatalf("objective metadata'sı dinamik plana taşınmadı: %+v", plan)
	}
}

func TestHistoricalVictoryConditionOverridesProfilePlan(t *testing.T) {
	gs := strategicPlanTestState()
	gs.ScenarioID = "1455_wars_of_the_roses"
	gs.AIStrategies = map[string]scenario.AIFactionStrategy{
		"ottoman": {
			FactionID: "ottoman",
			Objectives: []scenario.AIObjectiveDef{
				{ID: "hold_frontier", Kind: "defend", TargetFactions: []string{"germiyan_bey"}, TargetRegions: []string{"ottoman_home"}, Priority: 100},
			},
		},
	}
	gs.ScenarioVictories = []scenario.VictoryOptionDef{
		{ID: "ottoman_frontier_victory", Type: "conquer_city", RequiredRegions: []string{"germiyan_border"}, AllowedFactions: []string{"ottoman"}},
		{ID: "general_hegemony", Type: "domination", RequiredRegions: []string{"constantinople"}},
	}

	plan := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if plan == nil || plan.ObjectiveID != "victory:ottoman_frontier_victory" || plan.Kind != state.AIObjectiveExpand || plan.TargetFactionID != "germiyan_bey" {
		t.Fatalf("tarihsel zafer hedefi profil savunma planından önce seçilmeliydi: %+v", plan)
	}
}

func TestScenarioProfileExpansionBiasStartsAfterHistoricalOpening(t *testing.T) {
	gs := strategicPlanTestState()
	gs.AIStrategies = map[string]scenario.AIFactionStrategy{
		"ottoman": {
			FactionID: "ottoman",
			Objectives: []scenario.AIObjectiveDef{
				{ID: "hold_frontier", Kind: "defend", TargetFactions: []string{"germiyan_bey"}, TargetRegions: []string{"ottoman_home"}, Priority: 96},
				{ID: "press_border", Kind: "expand", TargetFactions: []string{"germiyan_bey"}, TargetRegions: []string{"germiyan_border"}, Priority: 70},
			},
		},
	}
	gs.Turn = 1

	opening := chooseScenarioObjectivePlan(gs, gs.Factions["ottoman"], buildStrategicContext(gs, "ottoman"))
	if opening == nil || opening.ObjectiveID != "hold_frontier" {
		t.Fatalf("tarihsel açılışta savunma objective'i korunmalıydı: %+v", opening)
	}

	gs.Turn = aiWarLogisticsActivationTurn
	postOpening := chooseScenarioObjectivePlan(gs, gs.Factions["ottoman"], buildStrategicContext(gs, "ottoman"))
	if postOpening == nil || postOpening.ObjectiveID != "press_border" {
		t.Fatalf("açılış sonrası genişleme objective'i proaktif bonus almalıydı: %+v", postOpening)
	}
}

func TestGeneralVictoryConditionGuidesFactionWithoutHistoricalGoal(t *testing.T) {
	gs := strategicPlanTestState()
	gs.ScenarioVictories = []scenario.VictoryOptionDef{
		{ID: "general_hegemony", Type: "domination", RequiredRegions: []string{"germiyan_border"}},
	}

	plan := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if plan == nil || plan.ObjectiveID != "victory:general_hegemony" || plan.Kind != state.AIObjectiveExpand || plan.TargetFactionID != "germiyan_bey" {
		t.Fatalf("tarihsel hedefi olmayan fraksiyon genel zafer hedefine yönelmeliydi: %+v", plan)
	}
}

func TestHistoricalPriorityDoesNotOverrideExtremePowerRisk(t *testing.T) {
	gs := strategicPlanTestState()
	gs.Year = 1300
	gs.Armies["east_army"].Units = make([]army.Unit, 80)
	for index := range gs.Armies["east_army"].Units {
		gs.Armies["east_army"].Units[index] = army.Unit{TypeID: "inf", CurrentHP: 100}
	}
	gs.AIStrategies = map[string]scenario.AIFactionStrategy{
		"ottoman": {
			FactionID: "ottoman",
			Objectives: []scenario.AIObjectiveDef{
				{ID: "secure_bithynia", Kind: "expand", TargetFactions: []string{"east_rome"}, TargetRegions: []string{"constantinople"}, Priority: 100},
				{ID: "unite_anatolian_beyliks", Kind: "expand", TargetFactions: []string{"germiyan_bey"}, TargetRegions: []string{"germiyan_border"}, Priority: 92, AllowVassalization: true},
			},
		},
	}

	plan := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if plan == nil || plan.ObjectiveID != "unite_anatolian_beyliks" || plan.TargetFactionID != "germiyan_bey" {
		t.Fatalf("soft tarihsel öncelik aşırı güç riskini bastırmamalıydı: %+v", plan)
	}
}

func TestScenarioObjectiveMinYearIsHardGate(t *testing.T) {
	gs := strategicPlanTestState()
	gs.Year = 1300
	gs.Factions["safavid"] = &faction.Faction{ID: "safavid"}
	gs.Regions["ardabil"] = &world.Region{ID: "ardabil", OwnerID: "safavid"}
	gs.Factions["ottoman"].AIExpansionTargets = nil
	gs.AIStrategies = map[string]scenario.AIFactionStrategy{
		"ottoman": {
			FactionID:         "ottoman",
			TerritorialClaims: []scenario.AITerritorialClaimDef{{RegionID: "ardabil", Value: 80}},
			Objectives: []scenario.AIObjectiveDef{
				{ID: "eastern_imperial_rivalry", Kind: "expand", TargetFactions: []string{"safavid"}, Priority: 100, MinYear: 1501, RequiredEventFlags: []string{"safavid_rise"}},
			},
		},
	}

	before := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if before == nil || before.Kind != state.AIObjectiveConsolidate || before.TargetFactionID != "" {
		t.Fatalf("1501 öncesi geç hedef hard gate arkasında kalmalıydı: %+v", before)
	}

	gs.Year = 1501
	gs.Turn = before.ReassessTurn
	atYear := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if atYear == nil || atYear.Kind != state.AIObjectiveConsolidate {
		t.Fatalf("zorunlu event flag olmadan hedef kapalı kalmalıydı: %+v", atYear)
	}

	gs.FiredEventIDs = map[string]bool{"flag:safavid_rise": true}
	gs.Turn = atYear.ReassessTurn
	after := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if after == nil || after.ObjectiveID != "eastern_imperial_rivalry" || after.TargetFactionID != "safavid" {
		t.Fatalf("yıl eşiğinde geç hedef açılmalıydı: %+v", after)
	}
}

func TestScenarioObjectiveMaxYearIsInclusiveHardGateAndAddsUrgency(t *testing.T) {
	objective := scenario.AIObjectiveDef{MinYear: 1453, MaxYear: 1454}
	gs := &state.GameState{Year: 1453}

	if !scenarioObjectiveHardGateActive(gs, objective) {
		t.Fatal("objective min_year ile max_year arasındaki yılda aktif olmalı")
	}
	if got := objectiveDeadlineUrgency(gs, objective); got != 120 {
		t.Fatalf("son geçerli yıldan bir yıl önce aciliyet yanlış: got=%d want=120", got)
	}

	gs.Year = 1454
	if !scenarioObjectiveHardGateActive(gs, objective) {
		t.Fatal("max_year verilen yılın sonuna kadar objective aktif kalmalı")
	}
	if got := objectiveDeadlineUrgency(gs, objective); got != 180 {
		t.Fatalf("son geçerli yılda aciliyet yanlış: got=%d want=180", got)
	}

	gs.Year = 1455
	if scenarioObjectiveHardGateActive(gs, objective) {
		t.Fatal("max_year sonrasındaki yılda objective kapanmalı")
	}
}

func TestScenarioObjectiveCompletionUsesClaimsButKeepsDefendIntent(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"home":   {ID: "home", OwnerID: "ai"},
			"border": {ID: "border", OwnerID: "enemy"},
		},
	}
	claim := scenario.AITerritorialClaimDef{RegionID: "home", Value: 100}

	if !scenarioObjectiveCompleted(gs, "ai", scenario.AIObjectiveDef{
		Kind:              "consolidate",
		TerritorialClaims: []scenario.AITerritorialClaimDef{claim},
	}) {
		t.Fatal("AI'nin elindeki tüm claim bölgeleri consolidate objective'ini tamamlamalı")
	}
	if scenarioObjectiveCompleted(gs, "ai", scenario.AIObjectiveDef{
		Kind:              "defend",
		TerritorialClaims: []scenario.AITerritorialClaimDef{claim},
	}) {
		t.Fatal("defend objective'i yalnız claim sahipliğiyle tamamlanmamalı")
	}
	if scenarioObjectiveCompleted(gs, "ai", scenario.AIObjectiveDef{
		Kind:              "expand",
		TerritorialClaims: []scenario.AITerritorialClaimDef{{RegionID: "border", Value: 100}},
	}) {
		t.Fatal("düşmanın elindeki expand claim'i tamamlanmış sayılmamalı")
	}
}

func TestConsolidationObjectiveRecoversLostClaimFromCurrentOwner(t *testing.T) {
	gs := strategicPlanTestState()
	gs.AIStrategies = map[string]scenario.AIFactionStrategy{
		"ottoman": {
			FactionID: "ottoman",
			Objectives: []scenario.AIObjectiveDef{{
				ID:                "secure_anatolian_core",
				Kind:              "consolidate",
				Priority:          100,
				TerritorialClaims: []scenario.AITerritorialClaimDef{{RegionID: "germiyan_border", Value: 100}},
			}},
		},
	}
	gs.Regions["germiyan_border"].OwnerID = "ottoman"

	initialPlan := chooseScenarioObjectivePlan(gs, gs.Factions["ottoman"], buildStrategicContext(gs, "ottoman"))
	if initialPlan == nil || initialPlan.ObjectiveID != "consolidate:ottoman" {
		t.Fatalf("başlangıçta eldeki consolidate claim'i hazırlık planı olarak tamamlanmalıydı: %+v", initialPlan)
	}
	if !scenarioObjectiveWasCompleted(gs, "ottoman", "secure_anatolian_core") {
		t.Fatal("başlangıçta tüm claim eldeyken objective tamamlanma geçmişi kaydedilmeliydi")
	}

	gs.Regions["germiyan_border"].OwnerID = "germiyan_bey"
	plan := chooseScenarioObjectivePlan(gs, gs.Factions["ottoman"], buildStrategicContext(gs, "ottoman"))
	if plan == nil || plan.ObjectiveID != "secure_anatolian_core" || plan.Kind != state.AIObjectiveExpand || plan.TargetFactionID != "germiyan_bey" {
		t.Fatalf("kaybedilen consolidate claim'i güncel sahibine karşı recovery expand planına dönüşmeliydi: %+v", plan)
	}
	if len(plan.TargetRegionIDs) != 1 || plan.TargetRegionIDs[0] != "germiyan_border" {
		t.Fatalf("recovery planı kaybedilen claim bölgesini hedeflemeliydi: %+v", plan.TargetRegionIDs)
	}

	gs.Regions["germiyan_border"].OwnerID = "ottoman"
	completedPlan := chooseScenarioObjectivePlan(gs, gs.Factions["ottoman"], buildStrategicContext(gs, "ottoman"))
	if completedPlan == nil || completedPlan.ObjectiveID != "consolidate:ottoman" || completedPlan.Kind != state.AIObjectiveConsolidate {
		t.Fatalf("claim geri alındığında tamamlanan objective hazırlık planına dönmeliydi: %+v", completedPlan)
	}
}

func TestDefensivePlanPullsArmyTowardPriorityRegion(t *testing.T) {
	gs := strategicPlanTestState()
	gs.Regions["reserve"] = &world.Region{ID: "reserve", OwnerID: "ottoman"}
	gs.AIPlans = map[faction.FactionID]*state.AIPlanState{
		"ottoman": {ObjectiveID: "hold_frontier", Kind: state.AIObjectiveDefend, TargetFactionID: "east_rome", TargetRegionIDs: []world.RegionID{"reserve"}},
	}
	army := gs.Armies["ottoman_army"]

	if score := scoreMove(gs, army, gs.Regions["reserve"]); score <= 0 {
		t.Fatalf("savunma objective bölgesi pozitif toplanma skoru üretmeliydi: %d", score)
	}
}
