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

func TestOrderedAIFactionsUsesFactionOrder(t *testing.T) {
	g := &Game{
		gs: &state.GameState{
			PlayerFactionID: "player",
			FactionOrder:    []faction.FactionID{"venice", "player", "mamluk"},
			Factions: map[faction.FactionID]*faction.Faction{
				"player":  {ID: "player"},
				"venice":  {ID: "venice"},
				"mamluk":  {ID: "mamluk"},
				"england": {ID: "england"},
			},
		},
	}

	got := g.orderedAIFactions()
	want := []faction.FactionID{"venice", "mamluk", "england"}
	if len(got) != len(want) {
		t.Fatalf("beklenen %v, got=%v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("beklenen %v, got=%v", want, got)
		}
	}
}

func TestRegionNearPlayerUsesOwnedRegionsAndArmies(t *testing.T) {
	g := &Game{
		gs: &state.GameState{
			PlayerFactionID: "player",
			Regions: map[world.RegionID]*world.Region{
				"home":     {ID: "home", OwnerID: "player", Neighbors: []world.RegionID{"frontier"}},
				"frontier": {ID: "frontier", Neighbors: []world.RegionID{"home", "outer"}},
				"outer":    {ID: "outer", Neighbors: []world.RegionID{"frontier", "deep"}},
				"deep":     {ID: "deep", Neighbors: []world.RegionID{"outer"}},
				"fleet":    {ID: "fleet", IsSea: true, Neighbors: []world.RegionID{"deep"}},
			},
			Armies: map[army.ArmyID]*army.Army{
				"player_stack": {ID: "player_stack", OwnerID: "player", RegionID: "fleet"},
			},
		},
	}

	if !g.regionNearPlayer("outer", 2) {
		t.Fatal("outer bölgesi oyuncu toprağından iki sıçrama içinde görünür olmalı")
	}
	if !g.regionNearPlayer("deep", 1) {
		t.Fatal("deep bölgesi oyuncu ordusunun bulunduğu denizden bir sıçrama içinde görünür olmalı")
	}
	if g.regionNearPlayer("deep", 0) {
		t.Fatal("sıfır derinlikte yalnız başlangıç bölgeleri görünür olmalı")
	}
}

func TestAITurnSequenceWaitsWhilePlayerOfferPending(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Phase:           state.PhaseAITurn,
		FactionOrder:    []faction.FactionID{"ai_1", "player"},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1"},
		},
		DiplomaticOffers: []state.DiplomaticOffer{
			{FromFactionID: "ai_1", ToFactionID: "player", Action: "propose_peace"},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}
	g.startAITurnSequence()
	g.updateAITurnSequence()

	if g.aiTurn == nil {
		t.Fatal("AI sıra durumu korunmalıydı")
	}
	if g.aiTurn.stepper != nil {
		t.Fatal("oyuncu cevabı beklenirken AI stepper ilerlememeli")
	}
	if g.aiTurn.index != 0 {
		t.Fatalf("aktif AI index ilerlememeliydi, got=%d", g.aiTurn.index)
	}
}

func TestAcceptedOfferEndsCurrentAIFactionTurn(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Phase:           state.PhaseAITurn,
		FactionOrder:    []faction.FactionID{"ai_1", "ai_2", "player"},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1"},
			"ai_2":   {ID: "ai_2", NameTR: "AI 2"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Stance: faction.StanceWar, Score: -80},
		},
		DiplomaticOffers: []state.DiplomaticOffer{
			{FromFactionID: "ai_1", ToFactionID: "player", Action: "propose_peace"},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}
	g.startAITurnSequence()
	g.aiTurn.stepper = nil
	g.aiTurn.index = 0
	g.aiTurn.stepper = ai.NewTurnStepper(gs, "ai_1")

	g.handleAITurnOfferResponse(0, true)

	if len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("kabul sonrası teklif kuyruktan düşmeliydi, got=%d", len(gs.DiplomaticOffers))
	}
	rel := gs.Relations[faction.RelationKey("ai_1", "player")]
	if rel == nil || rel.Stance != faction.StancePeace {
		t.Fatalf("barış uygulanmalıydı, got=%+v", rel)
	}
	if g.aiTurn.index != 1 {
		t.Fatalf("kabul sonrası aktif AI turu kapanmalıydı, got index=%d", g.aiTurn.index)
	}
	if g.aiTurn.stepper != nil {
		t.Fatal("aktif AI stepper'ı temizlenmeliydi")
	}
}
