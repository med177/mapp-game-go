package diplomacy

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func allianceStrategyTestState() *state.GameState {
	return &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Factions: map[faction.FactionID]*faction.Faction{
			"a":      {ID: "a", Religion: religion.Catholic, Gold: 500, Grain: 300},
			"b":      {ID: "b", Religion: religion.Catholic, Gold: 500, Grain: 300},
			"threat": {ID: "threat", Religion: religion.Catholic, Gold: 900, Grain: 500},
		},
		Regions: map[world.RegionID]*world.Region{
			"a1": {ID: "a1", OwnerID: "a", TradeCapacity: 6, Neighbors: []world.RegionID{"b1"}},
			"b1": {ID: "b1", OwnerID: "b", TradeCapacity: 6, Neighbors: []world.RegionID{"a1", "t1"}},
			"t1": {ID: "t1", OwnerID: "threat", TradeCapacity: 4, Neighbors: []world.RegionID{"b1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"aa": {ID: "aa", OwnerID: "a", RegionID: "a1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
			"bb": {ID: "bb", OwnerID: "b", RegionID: "b1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
			"tt": {ID: "tt", OwnerID: "threat", RegionID: "t1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"):      {FactionA: "a", FactionB: "b", Score: 45, Stance: faction.StancePeace},
			faction.RelationKey("a", "threat"): {FactionA: "a", FactionB: "threat", Score: -80, Stance: faction.StanceWar},
		},
	}
}

func TestAllianceStrategyBlocksActiveObjectiveConflict(t *testing.T) {
	gs := allianceStrategyTestState()
	gs.AIPlans = map[faction.FactionID]*state.AIPlanState{
		"a": {Kind: state.AIObjectiveExpand, TargetFactionID: "b", Commitment: 70},
	}
	rel := Relation(gs, "a", "b")
	assessment := AssessAllianceProposal(gs, rel, "a", "b")
	if !assessment.TargetStrategic.ActiveObjectiveConflict || assessment.BlockReason != "Aktif stratejik hedefler ittifakla çakışıyor" {
		t.Fatalf("aktif objective ittifakı kesin engellemeliydi: %+v", assessment)
	}
	if result := Execute(gs, "a", "b", ActionProposeAlliance); result.Applied {
		t.Fatalf("objective çakışmasına rağmen ittifak uygulandı: %+v", result)
	}
}

func TestAllianceStrategyTreatsFutureExpansionAsOverridablePenalty(t *testing.T) {
	gs := allianceStrategyTestState()
	gs.Factions["a"].AIExpansionTargets = []faction.FactionID{"b"}
	gs.Relations[faction.RelationKey("b", "threat")] = &faction.Relation{
		FactionA: "b", FactionB: "threat", Score: -80, Stance: faction.StanceWar,
	}

	actorView := AssessStrategicAlliance(gs, "a", "b")
	if actorView.ExpansionTensionPenalty != 18 || !actorView.CommonEnemy || actorView.Score < 18 {
		t.Fatalf("gelecek hedef gerilimi ortak tehditle aşılabilmeliydi: %+v", actorView)
	}
	assessment := AssessAllianceProposal(gs, Relation(gs, "a", "b"), "a", "b")
	if assessment.BlockReason != "" || !assessment.Accepted() {
		t.Fatalf("ortak tehdide rağmen teklif reddedildi: %+v", assessment)
	}
}

func TestAllianceStrategyValuesBufferFrontAndTrade(t *testing.T) {
	gs := allianceStrategyTestState()
	assessment := AssessStrategicAlliance(gs, "a", "b")
	if assessment.BufferValue == 0 {
		t.Fatalf("tehdit sınırındaki aday tampon değeri üretmedi: %+v", assessment)
	}
	if assessment.FrontSupportValue == 0 {
		t.Fatalf("adayın tehdit cephesindeki ordusu destek değeri üretmedi: %+v", assessment)
	}
	if assessment.TradeValue == 0 {
		t.Fatalf("bağlanabilir ticaret hattı değer üretmedi: %+v", assessment)
	}
}

func TestAllianceProposalRejectsStrategicallyEmptyAITarget(t *testing.T) {
	gs := allianceStrategyTestState()
	delete(gs.Relations, faction.RelationKey("a", "threat"))
	delete(gs.Armies, "aa")
	delete(gs.Armies, "bb")
	gs.Regions["a1"].TradeCapacity = 0
	gs.Regions["b1"].TradeCapacity = 0
	gs.Regions["b1"].Neighbors = []world.RegionID{"a1"}
	gs.Regions["t1"].Neighbors = nil

	assessment := AssessAllianceProposal(gs, Relation(gs, "a", "b"), "a", "b")
	if assessment.BlockReason != "İttifak hedef devlet için yeterli stratejik değer üretmiyor" {
		t.Fatalf("stratejik değersiz AI ittifakı reddedilmeliydi: %+v", assessment)
	}
}
