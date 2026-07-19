package ai

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIActiveObjectiveCancelsAllianceAndBlockingTrade(t *testing.T) {
	gs := &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Difficulty: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", Religion: religion.Catholic},
			"b": {ID: "b", Religion: religion.Catholic},
		},
		Regions: map[world.RegionID]*world.Region{
			"a1": {ID: "a1", OwnerID: "a", TradeCapacity: 6, Neighbors: []world.RegionID{"b1"}},
			"b1": {ID: "b1", OwnerID: "b", TradeCapacity: 6, Neighbors: []world.RegionID{"a1"}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b", Score: 40, Stance: faction.StanceAllied},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "a", ToFactionID: "b"},
			{FromFactionID: "b", ToFactionID: "a"},
		},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"a": {Kind: state.AIObjectiveExpand, TargetFactionID: "b", Commitment: 70},
		},
	}

	aiHandleDiplomacy(gs, "a")
	rel := gs.Relations[faction.RelationKey("a", "b")]
	if rel.Stance != faction.StancePeace || len(gs.TradeRoutes) != 0 {
		t.Fatalf("aktif objective ittifak ve engelleyici ticareti bitirmeliydi: rel=%+v routes=%+v", rel, gs.TradeRoutes)
	}
}

func TestAIKeepsFutureTargetAllianceUnderCommonThreat(t *testing.T) {
	gs := allianceStrategyAITestState()
	gs.Factions["a"].AIExpansionTargets = []faction.FactionID{"b"}
	if aiShouldCancelAlliance(gs, "a", "b") {
		t.Fatal("statik gelecek hedefi ortak tehdit varken ittifakı bozmamalıydı")
	}
}

func allianceStrategyAITestState() *state.GameState {
	gs := &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Factions: map[faction.FactionID]*faction.Faction{
			"a":      {ID: "a", Religion: religion.Catholic},
			"b":      {ID: "b", Religion: religion.Catholic},
			"threat": {ID: "threat", Religion: religion.Catholic},
		},
		Regions: map[world.RegionID]*world.Region{
			"a1": {ID: "a1", OwnerID: "a", TradeCapacity: 4, Neighbors: []world.RegionID{"b1"}},
			"b1": {ID: "b1", OwnerID: "b", TradeCapacity: 4, Neighbors: []world.RegionID{"a1", "t1"}},
			"t1": {ID: "t1", OwnerID: "threat", Neighbors: []world.RegionID{"b1"}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"):      {FactionA: "a", FactionB: "b", Score: 40, Stance: faction.StanceAllied},
			faction.RelationKey("a", "threat"): {FactionA: "a", FactionB: "threat", Score: -80, Stance: faction.StanceWar},
			faction.RelationKey("b", "threat"): {FactionA: "b", FactionB: "threat", Score: -80, Stance: faction.StanceWar},
		},
	}
	return gs
}
