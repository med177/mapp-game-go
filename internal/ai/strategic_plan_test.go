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

func TestScenarioProfileCreatesSoftHistoricalObjectivePlan(t *testing.T) {
	gs := strategicPlanTestState()
	gs.Year = 1300
	gs.AIStrategies = map[string]scenario.AIFactionStrategy{
		"ottoman": {
			FactionID: "ottoman",
			Profile:   "frontier_expansion",
			Objectives: []scenario.AIObjectiveDef{
				{ID: "secure_bithynia", Kind: "expand", TargetFactions: []string{"east_rome"}, TargetRegions: []string{"constantinople"}, Priority: 100, Commitment: 74, AnnexRegionIDs: []string{"constantinople"}},
				{ID: "unite_anatolian_beyliks", Kind: "expand", TargetFactions: []string{"germiyan_bey"}, TargetRegions: []string{"germiyan_border"}, Priority: 70, AllowVassalization: true},
			},
		},
	}

	plan := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if plan == nil || plan.ObjectiveID != "secure_bithynia" || plan.TargetFactionID != "east_rome" {
		t.Fatalf("yüksek öncelikli tarihsel yön seçilmedi: %+v", plan)
	}
	if plan.Commitment != 74 || len(plan.AnnexRegionIDs) != 1 || plan.AnnexRegionIDs[0] != "constantinople" {
		t.Fatalf("objective metadata'sı dinamik plana taşınmadı: %+v", plan)
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
			FactionID: "ottoman",
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
