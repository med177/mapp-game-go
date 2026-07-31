package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIWarAssessmentIncludesAlliedAndVassalMilitaryPower(t *testing.T) {
	gs := aiTestState()
	gs.Factions["ally"] = &faction.Faction{ID: "ally", NameTR: "Müttefik", Religion: religion.Catholic}
	gs.Factions["ally_v"] = &faction.Faction{ID: "ally_v", NameTR: "Müttefik Vassalı", Religion: religion.Catholic, OverlordID: "ally"}
	gs.Regions["b1"].Neighbors = []world.RegionID{"a1", "ally_land"}
	gs.Regions["ally_land"] = &world.Region{ID: "ally_land", OwnerID: "ally", Neighbors: []world.RegionID{"b1", "ally_v_land"}}
	gs.Regions["ally_v_land"] = &world.Region{ID: "ally_v_land", OwnerID: "ally_v", Neighbors: []world.RegionID{"ally_land"}}
	gs.Relations[faction.RelationKey("ai_2", "ally")] = &faction.Relation{
		FactionA: "ai_2", FactionB: "ally", Score: 80, Stance: faction.StanceAllied,
	}
	gs.Armies["ally_army"] = &army.Army{
		ID: "ally_army", OwnerID: "ally", RegionID: "ally_land",
		Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
	}
	gs.Armies["ally_v_army"] = &army.Army{
		ID: "ally_v_army", OwnerID: "ally_v", RegionID: "ally_v_land",
		Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
	}

	risk := aiWarCoalitionAssessment(gs, "ai_1", "ai_2")
	if risk.AllyPower <= 0 {
		t.Fatalf("hedefin müttefik askerî gücü hesaba katılmadı: %+v", risk)
	}
	if risk.AllyVassalPower <= 0 {
		t.Fatalf("hedef müttefikinin vassal askerî gücü hesaba katılmadı: %+v", risk)
	}
	if risk.NearestAllyArmy != 1 {
		t.Fatalf("müttefik cephe mesafesi yanlış hesaplandı: got=%d", risk.NearestAllyArmy)
	}
	if risk.DefenderPower <= risk.TargetPower+risk.TargetVassalPower {
		t.Fatalf("savunma koalisyonuna müttefik katkısı eklenmedi: %+v", risk)
	}
}

func TestAIWarDecisionUsesAlliedDistanceAndCoalitionPower(t *testing.T) {
	closeState := aiWarAlliedTargetState(false)
	aiEvaluateWarOpportunities(closeState, "ai_1")
	if relation := closeState.Relations[faction.RelationKey("ai_1", "ai_2")]; relation.Stance == faction.StanceWar {
		t.Fatal("hedefe yakın güçlü müttefik varken AI savaş açmamalıydı")
	}

	farState := aiWarAlliedTargetState(true)
	aiEvaluateWarOpportunities(farState, "ai_1")
	if relation := farState.Relations[faction.RelationKey("ai_1", "ai_2")]; relation.Stance != faction.StanceWar {
		t.Fatal("uzak müttefik gücü mesafe nedeniyle yakın cephe gücü gibi sayılmamalıydı")
	}
}

func TestAIWarAssessmentIncludesCertainAttackingAlly(t *testing.T) {
	closeState := aiWarCertainAttackerAllyState(false)
	call := diplomacy.AssessWarCall(closeState, "ai_1", "ally", "ai_2")
	if !call.AutoJoin {
		t.Fatalf("test müttefiki kesin otomatik katılmalıydı: %+v", call)
	}

	closeRisk := aiWarCoalitionAssessment(closeState, "ai_1", "ai_2")
	farRisk := aiWarCoalitionAssessment(aiWarCertainAttackerAllyState(true), "ai_1", "ai_2")
	if closeRisk.CertainAttackerAllyPower <= 0 || closeRisk.NearestAttackerAllyArmy != 1 {
		t.Fatalf("kesin katılan saldıran müttefik hesaba katılmadı: %+v", closeRisk)
	}
	if closeRisk.AttackerPower <= farRisk.AttackerPower {
		t.Fatalf("yakın kesin müttefik gücü uzak müttefikten fazla olmalıydı: close=%+v far=%+v", closeRisk, farRisk)
	}
}

func TestAIWarDecisionUsesCertainAttackingAllyDistance(t *testing.T) {
	closeState := aiWarCertainAttackerAllyState(false)
	aiEvaluateWarOpportunities(closeState, "ai_1")
	if relation := closeState.Relations[faction.RelationKey("ai_1", "ai_2")]; relation.Stance != faction.StanceWar {
		t.Fatal("yakındaki kesin katılan güçlü müttefik ile AI savaş açabilmeliydi")
	}

	farState := aiWarCertainAttackerAllyState(true)
	aiEvaluateWarOpportunities(farState, "ai_1")
	if relation := farState.Relations[faction.RelationKey("ai_1", "ai_2")]; relation.Stance == faction.StanceWar {
		t.Fatal("uzaktaki kesin müttefik yakın cephe gücü gibi sayılarak savaş açılmamalıydı")
	}
}

func aiWarAlliedTargetState(farAlly bool) *state.GameState {
	gs := aiTestState()
	gs.Difficulty = 2
	gs.Month = 1
	gs.Factions["ai_1"].AIAggressiveness = 70
	gs.Relations[faction.RelationKey("ai_1", "ai_2")].Score = -60
	gs.Armies["ai1_army"].Units = []army.Unit{
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
	}
	gs.Factions["ally"] = &faction.Faction{ID: "ally", NameTR: "Müttefik", Religion: religion.Catholic}
	gs.Relations[faction.RelationKey("ai_2", "ally")] = &faction.Relation{
		FactionA: "ai_2", FactionB: "ally", Score: 80, Stance: faction.StanceAllied,
	}
	gs.Regions["a1"].Neighbors = []world.RegionID{"b1"}
	gs.Regions["b1"].Neighbors = []world.RegionID{"a1"}

	allyRegionID := world.RegionID("ally_close")
	if farAlly {
		previous := world.RegionID("b1")
		for i := 1; i <= 6; i++ {
			current := world.RegionID("ally_far_" + string(rune('0'+i)))
			gs.Regions[current] = &world.Region{ID: current, OwnerID: "ally", Neighbors: []world.RegionID{previous}}
			gs.Regions[previous].Neighbors = append(gs.Regions[previous].Neighbors, current)
			previous = current
		}
		allyRegionID = previous
	} else {
		gs.Regions["b1"].Neighbors = append(gs.Regions["b1"].Neighbors, allyRegionID)
		gs.Regions[allyRegionID] = &world.Region{ID: allyRegionID, OwnerID: "ally", Neighbors: []world.RegionID{"b1"}}
	}
	gs.Armies["ally_army"] = &army.Army{
		ID: "ally_army", OwnerID: "ally", RegionID: allyRegionID,
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
		},
	}
	return gs
}

func aiWarCertainAttackerAllyState(farAlly bool) *state.GameState {
	gs := aiTestState()
	gs.Difficulty = 3
	gs.Month = 1
	gs.Factions["ai_1"].AIAggressiveness = 70
	gs.Factions["ai_1"].AIExpansionTargets = []faction.FactionID{"ai_2"}
	gs.Relations[faction.RelationKey("ai_1", "ai_2")].Score = -60
	gs.Factions["ally"] = &faction.Faction{ID: "ally", NameTR: "Kesin Müttefik", Religion: religion.Catholic}
	gs.Relations[faction.RelationKey("ai_1", "ally")] = &faction.Relation{
		FactionA: "ai_1", FactionB: "ally", Score: 80, Stance: faction.StanceAllied,
	}
	gs.Relations[faction.RelationKey("ally", "ai_2")] = &faction.Relation{
		FactionA: "ally", FactionB: "ai_2", Score: -80, Stance: faction.StanceWar,
	}
	gs.Armies["ai2_army"].Units = []army.Unit{
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
	}
	gs.Regions["a1"].Neighbors = []world.RegionID{"b1"}
	gs.Regions["b1"].Neighbors = []world.RegionID{"a1"}

	allyRegionID := world.RegionID("ally_close")
	if farAlly {
		previous := world.RegionID("b1")
		for i := 1; i <= 6; i++ {
			current := world.RegionID("attacker_ally_far_" + string(rune('0'+i)))
			gs.Regions[current] = &world.Region{ID: current, OwnerID: "ally", Neighbors: []world.RegionID{previous}}
			gs.Regions[previous].Neighbors = append(gs.Regions[previous].Neighbors, current)
			previous = current
		}
		allyRegionID = previous
	} else {
		gs.Regions["b1"].Neighbors = append(gs.Regions["b1"].Neighbors, allyRegionID)
		gs.Regions[allyRegionID] = &world.Region{ID: allyRegionID, OwnerID: "ally", Neighbors: []world.RegionID{"b1"}}
	}
	gs.Armies["ally_army"] = &army.Army{
		ID: "ally_army", OwnerID: "ally", RegionID: allyRegionID,
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
		},
	}
	return gs
}
