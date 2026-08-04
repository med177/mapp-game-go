package game

import (
	"testing"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAITurnResolvesAIOnlyNavalContact(t *testing.T) {
	gs := &state.GameState{
		Phase:           state.PhaseAITurn,
		PlayerFactionID: "player",
		FactionOrder:    []faction.FactionID{"ai_a", "ai_b"},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ai_a":   {ID: "ai_a"},
			"ai_b":   {ID: "ai_b"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_a", "ai_b"): {
				FactionA: "ai_a", FactionB: "ai_b", Stance: faction.StanceWar,
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"sea":    {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"escape"}},
			"escape": {ID: "escape", IsSea: true, Neighbors: []world.RegionID{"sea"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_a_fleet": {ID: "ai_a_fleet", OwnerID: "ai_a", RegionID: "sea", IsNaval: true, MovePoints: 1, MaxMovePoints: 1, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
			"ai_b_fleet": {ID: "ai_b_fleet", OwnerID: "ai_b", RegionID: "sea", IsNaval: true, MovePoints: 1, MaxMovePoints: 1, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"warship": {ID: "warship", Category: army.CategoryNavalWar, Attack: 100, Defense: 100, Morale: 100},
		},
	}
	contact := gs.BeginNavalContact(gs.Armies["ai_a_fleet"], gs.Armies["ai_b_fleet"], "sea", "", state.NavalContactWarOpening)
	if contact == nil {
		t.Fatal("AI-AI savaş açılışında deniz teması oluşmalı")
	}
	ai.ResolveNavalContactDecision(gs, contact)
	g := &Game{
		gs:       gs,
		renderer: render.New(gs),
		aiTurn:   &aiTurnState{order: []faction.FactionID{"ai_a"}},
	}

	if err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if gs.PendingNavalContact != nil {
		t.Fatalf("AI-AI teması oyuncu scheduler'ını bloklamamalı: %+v", gs.PendingNavalContact)
	}
}
