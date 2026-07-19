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
