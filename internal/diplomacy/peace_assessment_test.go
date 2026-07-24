package diplomacy

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func peaceTestState() *state.GameState {
	gs := &state.GameState{
		Turn:       1,
		ScenarioID: "1300_ottoman_rise",
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", Gold: 500, Grain: 500},
			"b": {ID: "b", Gold: 500, Grain: 500},
		},
		Regions: map[world.RegionID]*world.Region{
			"a1": {ID: "a1", OwnerID: "a"},
			"a2": {ID: "a2", OwnerID: "a"},
			"b1": {ID: "b1", OwnerID: "b"},
			"b2": {ID: "b2", OwnerID: "b"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"aa": {ID: "aa", OwnerID: "a", Units: []army.Unit{{TypeID: "u", CurrentHP: 100}}},
			"bb": {ID: "bb", OwnerID: "b", Units: []army.Unit{{TypeID: "u", CurrentHP: 100}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b", Stance: faction.StanceWar, Score: -80},
		},
	}
	gs.BeginWarLedger("a", "b")
	return gs
}

func TestPeaceAssessmentBlocksFirstThreeTurnsWithoutEmergency(t *testing.T) {
	gs := peaceTestState()
	gs.Turn = 3
	assessment := AssessPeaceDesire(gs, "a", "b")
	if assessment.Emergency || assessment.Eligible || assessment.ShouldPropose() {
		t.Fatalf("erken barış kapısı açık olmamalı: %+v", assessment)
	}
}

func TestPeaceAssessmentAllowsEarlyMilitaryCollapse(t *testing.T) {
	gs := peaceTestState()
	delete(gs.Armies, "aa")
	assessment := AssessPeaceDesire(gs, "a", "b")
	if !assessment.MilitaryCollapse || !assessment.Emergency || !assessment.ShouldPropose() {
		t.Fatalf("askeri çöküş erken barış kapısını açmalı: %+v", assessment)
	}
}

func TestPeaceAssessmentUsesLossesDurationObjectiveAndCooldown(t *testing.T) {
	gs := peaceTestState()
	ledger := gs.WarLedgerFor("a", "b")
	gs.Turn = 11
	gs.Regions["a2"].OwnerID = "b"
	gs.RecordWarRegionCapture("b", "a")
	gs.RecordWarCasualties("b", "a", 1, 5)
	ledger.LastBattleTurn = 1
	gs.AIPlans = map[faction.FactionID]*state.AIPlanState{
		"b": {
			Kind:            state.AIObjectiveExpand,
			TargetFactionID: "a",
			TargetRegionIDs: []world.RegionID{"a2"},
			Commitment:      60,
		},
	}

	aAssessment := AssessPeaceDesire(gs, "a", "b")
	bAssessment := AssessPeaceDesire(gs, "b", "a")
	if !aAssessment.ShouldPropose() {
		t.Fatalf("kaybeden taraf barış istemeli: %+v", aAssessment)
	}
	if !bAssessment.ObjectiveDone || !bAssessment.ShouldPropose() {
		t.Fatalf("hedefini tamamlayan taraf barışı kabul etmeli: %+v", bAssessment)
	}

	gs.MarkPeaceOffer("a", "b")
	if AssessPeaceDesire(gs, "a", "b").Eligible {
		t.Fatal("barış teklifi cooldown'u aynı turda yeni teklifi engellemedi")
	}
	ledger.LastPeaceOfferTurn = 0
	result := Execute(gs, "a", "b", ActionProposePeace)
	if !result.Applied || IsWar(gs, "a", "b") || gs.WarLedgerFor("a", "b") != nil {
		t.Fatalf("iki tarafın kabul ettiği barış uygulanmadı: %+v", result)
	}
}

func TestPeaceAssessmentRecognizesLongWarStalemate(t *testing.T) {
	gs := peaceTestState()
	gs.Turn = 20
	ledger := gs.WarLedgerFor("a", "b")
	ledger.LastBattleTurn = 5

	assessment := AssessPeaceDesire(gs, "a", "b")
	if !assessment.Stalemate || !assessment.ShouldPropose() {
		t.Fatalf("uzun süre sonuçsuz kalan savaş barış baskısı üretmeliydi: %+v", assessment)
	}
}

func TestPeaceAssessmentReportsWarScoreAndObjectiveProgress(t *testing.T) {
	gs := peaceTestState()
	gs.Turn = 12
	gs.AIPlans = map[faction.FactionID]*state.AIPlanState{
		"a": {
			Kind:            state.AIObjectiveExpand,
			TargetFactionID: "b",
			TargetRegionIDs: []world.RegionID{"b1", "b2"},
		},
	}
	gs.Regions["b1"].OwnerID = "a"
	ledger := gs.WarLedgerFor("a", "b")
	ledger.RegionsCapturedA = 1
	ledger.CasualtiesA = 2
	ledger.CasualtiesB = 8

	assessment := AssessPeaceDesire(gs, "a", "b")
	if assessment.WarScore <= 0 || assessment.ObjectiveHeld != 1 || assessment.ObjectiveTotal != 2 {
		t.Fatalf("savaş skoru/objective ilerlemesi yanlış: %+v", assessment)
	}
}

func TestPeaceAssessmentReportsWarExhaustionPressures(t *testing.T) {
	gs := peaceTestState()
	gs.Turn = 20
	gs.Factions["a"].Gold = 40
	gs.Factions["a"].Grain = 20
	gs.Regions["a1"].Satisfaction = 20
	gs.Regions["a2"].Satisfaction = 40
	gs.GrainEconomy = map[faction.FactionID]state.GrainEconomyStatus{
		"a": {FactionID: "a", SupplyLevel: state.GrainSupplyFamine, TotalDemand: 20},
	}
	ledger := gs.WarLedgerFor("a", "b")
	ledger.StartedTurn = 1
	ledger.CasualtiesA = 5
	ledger.LastBattleTurn = 19

	assessment := AssessPeaceDesire(gs, "a", "b")
	if assessment.WarExhaustion <= 0 || assessment.GoldPressure <= 0 || assessment.GrainPressure != 20 || assessment.SatisfactionPressure <= 0 || assessment.RelationshipPressure <= 0 {
		t.Fatalf("savaş yorgunluğu baskıları görünür raporlanmalıydı: %+v", assessment)
	}
}

func TestAssessPeaceSettlementSelectsCessionForClearWinner(t *testing.T) {
	gs := peaceTestState()
	gs.Turn = 12
	gs.Regions["b1"].OwnerID = "a"
	gs.Regions["a1"].Neighbors = []world.RegionID{"b2"}
	gs.Regions["b2"].Neighbors = []world.RegionID{"a1"}
	ledger := gs.WarLedgerFor("a", "b")
	ledger.RegionsCapturedA = 1
	ledger.CasualtiesA = 1
	ledger.CasualtiesB = 12

	settlement := AssessPeaceSettlement(gs, "a", "b")
	if settlement.Outcome != PeaceOutcomeCedeRegion || settlement.Winner != "a" || settlement.Loser != "b" || settlement.RegionID != "b2" {
		t.Fatalf("açık üstünlükte bölge bırakma seçilmeliydi: %+v", settlement)
	}
}

func TestExecuteAIPeaceAppliesReparations(t *testing.T) {
	gs := peaceTestState()
	gs.Turn = 50
	gs.Factions["a"].Gold = 200
	gs.Factions["b"].Gold = 100
	gs.Factions["a"].CapitalSettlementID = "a_cap"
	gs.Factions["b"].CapitalSettlementID = "b_cap"
	gs.Regions["a1"].Settlements = []world.Settlement{{ID: "a_cap", IsCapital: true}}
	gs.Regions["b1"].Settlements = []world.Settlement{{ID: "b_cap", IsCapital: true}}
	gs.Factions["b"].CapitalSettlementID = "b_cap"
	gs.Regions["b2"].OwnerID = "a"
	ledger := gs.WarLedgerFor("a", "b")
	ledger.RegionsCapturedA = 2
	ledger.CasualtiesA = 1
	ledger.CasualtiesB = 12

	result := ExecuteAIPeace(gs, "a", "b")
	if !result.Applied || result.Settlement == nil || result.Settlement.Outcome != PeaceOutcomeReparations {
		t.Fatalf("AI-AI barışında tazminat sonucu uygulanmalıydı: %+v", result)
	}
	if gs.Factions["a"].Gold <= 200 || gs.Factions["b"].Gold >= 100 {
		t.Fatalf("tazminat altınları aktarmalıydı: a=%d b=%d", gs.Factions["a"].Gold, gs.Factions["b"].Gold)
	}
}
