package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiNavalContactTestState() *state.GameState {
	return &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"sea_a": {ID: "sea_a", IsSea: true, Neighbors: []world.RegionID{"sea_b"}},
			"sea_b": {ID: "sea_b", IsSea: true, Neighbors: []world.RegionID{"sea_a", "sea_c"}},
			"sea_c": {ID: "sea_c", IsSea: true, Neighbors: []world.RegionID{"sea_b"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"weak": {
				ID: "weak", OwnerID: "ai_a", RegionID: "sea_a", IsNaval: true,
				MovePoints: 2, MaxMovePoints: 2, Units: []army.Unit{{TypeID: "transport", CurrentHP: 100}},
			},
			"strong": {
				ID: "strong", OwnerID: "ai_b", RegionID: "sea_b", IsNaval: true,
				MovePoints: 2, MaxMovePoints: 2, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}},
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ai_a": {ID: "ai_a"}, "ai_b": {ID: "ai_b"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_a", "ai_b"): {FactionA: "ai_a", FactionB: "ai_b", Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, Attack: 1, Defense: 1, Morale: 1},
			"warship":   {ID: "warship", Category: army.CategoryNavalWar, Attack: 100, Defense: 100, Morale: 100},
		},
	}
}

func TestAIContactRetreatMovesWeakFleetBackAfterEntry(t *testing.T) {
	gs := aiNavalContactTestState()
	outcome := executeMoveWithNavalPatrolAndContact(gs, gs.Armies["weak"], "sea_b", "ai_a", false, false)

	if !outcome.survived || gs.PendingNavalContact != nil {
		t.Fatalf("AI-AI temasından sonra filo hayatta kalıp temas temizlenmeliydi: survived=%v contact=%+v", outcome.survived, gs.PendingNavalContact)
	}
	weak := gs.Armies["weak"]
	if weak.RegionID != "sea_c" || weak.MovePoints != 0 {
		t.Fatalf("zayıf AI filosu girişten sonra geri dönüp 2 hareket puanı harcamalı: region=%s mp=%d", weak.RegionID, weak.MovePoints)
	}
}

func TestAIContactCannotChooseWithdrawWithoutMovementPoint(t *testing.T) {
	gs := aiNavalContactTestState()
	gs.Armies["weak"].MovePoints = 0
	gs.Armies["weak"].RegionID = "sea_b"
	gs.Armies["strong"].RegionID = "sea_b"
	contact := gs.BeginNavalContact(gs.Armies["weak"], gs.Armies["strong"], "sea_b", "", state.NavalContactWarOpening)
	ResolveNavalContactDecision(gs, contact)

	if contact.AttackerDecision == state.NavalContactWithdraw {
		t.Fatal("hareket puanı olmayan AI filosu geri çekilme kararı verememeli")
	}
}

func TestAIContactRetreatPrefersUnoccupiedSea(t *testing.T) {
	gs := aiNavalContactTestState()
	gs.Regions["sea_a"].Neighbors = []world.RegionID{"sea_b"}
	gs.Regions["sea_b"].Neighbors = []world.RegionID{"sea_a", "sea_c"}
	gs.Regions["sea_c"] = &world.Region{ID: "sea_c", IsSea: true, Neighbors: []world.RegionID{"sea_b"}}
	gs.Armies["blocking"] = &army.Army{
		ID: "blocking", OwnerID: "ai_b", RegionID: "sea_a", IsNaval: true,
		MovePoints: 2, MaxMovePoints: 2, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}},
	}

	outcome := executeMoveWithNavalPatrolAndContact(gs, gs.Armies["weak"], "sea_b", "ai_a", false, false)
	if !outcome.survived || gs.Armies["weak"].RegionID != "sea_c" {
		t.Fatalf("AI geri çekilmede düşmansız denizi tercih etmeliydi: survived=%v region=%s", outcome.survived, gs.Armies["weak"].RegionID)
	}
}
