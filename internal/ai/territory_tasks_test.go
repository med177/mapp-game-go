package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAITerritoryTaskCanSetAmbushAndKeepArmyHidden(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"pass":       {ID: "pass", OwnerID: "enemy", Terrain: world.TerrainPass, Neighbors: []world.RegionID{"enemy_rear"}},
			"enemy_rear": {ID: "enemy_rear", OwnerID: "enemy", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"pass"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_army":    {ID: "ai_army", OwnerID: "ai", RegionID: "pass", MovePoints: 2, Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"enemy_army": {ID: "enemy_army", OwnerID: "enemy", RegionID: "enemy_rear", MovePoints: 2, Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "enemy"): {FactionA: "ai", FactionB: "enemy", Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50}},
	}
	step, handled := executeAITerritoryTask(gs, gs.Armies["ai_army"], "ai")
	if !handled || !gs.Armies["ai_army"].InAmbush {
		t.Fatalf("AI pusu görevi uygulanmalıydı: handled=%v army=%+v", handled, gs.Armies["ai_army"])
	}
	if step.Message == "" || gs.Armies["ai_army"].MovePoints != 0 {
		t.Fatalf("AI pusu görevi hareketi tüketip kayıt üretmeli: step=%+v army=%+v", step, gs.Armies["ai_army"])
	}
}

func TestAITerritoryTaskRaidsValuableRegionWithoutAmbushOpportunity(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"rich": {ID: "rich", OwnerID: "enemy", Terrain: world.TerrainPlain, BaseGoldIncome: 20, BaseGrainOutput: 15, BaseIronOutput: 8},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_army": {ID: "ai_army", OwnerID: "ai", RegionID: "rich", MovePoints: 2, Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "enemy"): {FactionA: "ai", FactionB: "enemy", Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50}},
	}

	step, handled := executeAITerritoryTask(gs, gs.Armies["ai_army"], "ai")
	if !handled || gs.Raids["rich"] == nil {
		t.Fatalf("değerli ve savunmasız düşman bölgesi yağmalanmalı: handled=%v raids=%+v", handled, gs.Raids)
	}
	if step.Message == "" || gs.Armies["ai_army"].MovePoints != 0 {
		t.Fatalf("yağma görevi hareketi tüketip görünür step üretmeli: step=%+v army=%+v", step, gs.Armies["ai_army"])
	}
}

func TestAITerritoryTaskDoesNotDelayPrimaryConquestTarget(t *testing.T) {
	gs := &state.GameState{
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"ai": {Kind: state.AIObjectiveExpand, TargetFactionID: "enemy", TargetRegionIDs: []world.RegionID{"objective"}},
		},
		Regions: map[world.RegionID]*world.Region{
			"objective": {ID: "objective", OwnerID: "enemy", Terrain: world.TerrainPass, BaseGoldIncome: 30},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_army": {ID: "ai_army", OwnerID: "ai", RegionID: "objective", MovePoints: 2, Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "enemy"): {FactionA: "ai", FactionB: "enemy", Stance: faction.StanceWar},
		},
	}

	if step, handled := executeAITerritoryTask(gs, gs.Armies["ai_army"], "ai"); handled || step.Message != "" {
		t.Fatalf("ana fetih hedefi yağma/pusu ile geciktirilmemeli: handled=%v step=%+v", handled, step)
	}
}
