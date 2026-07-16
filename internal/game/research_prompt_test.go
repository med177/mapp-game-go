package game

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
)

func TestPlayerHasResearchableTechsReturnsFalseWhenAllTechsCompleted(t *testing.T) {
	g := &Game{
		gs: &state.GameState{
			PlayerFactionID: "player",
			Factions: map[faction.FactionID]*faction.Faction{
				"player": {
					ID: "player",
					Research: faction.ResearchState{
						Completed: map[string]bool{
							"basic":    true,
							"advanced": true,
						},
					},
				},
			},
			TechTypes: map[string]*tech.Technology{
				"basic":    {ID: "basic"},
				"advanced": {ID: "advanced", Requires: []string{"basic"}},
			},
		},
	}

	if g.playerHasResearchableTechs() {
		t.Fatalf("tum teknolojiler tamamlanmissa arastirilabilir tech kalmamali")
	}
}

func TestPlayerHasResearchableTechsReturnsTrueWhenUnlockedTechRemains(t *testing.T) {
	g := &Game{
		gs: &state.GameState{
			PlayerFactionID: "player",
			Factions: map[faction.FactionID]*faction.Faction{
				"player": {
					ID: "player",
					Research: faction.ResearchState{
						Completed: map[string]bool{
							"basic": true,
						},
					},
				},
			},
			TechTypes: map[string]*tech.Technology{
				"basic":    {ID: "basic"},
				"advanced": {ID: "advanced", Requires: []string{"basic"}},
			},
		},
	}

	if !g.playerHasResearchableTechs() {
		t.Fatalf("kilidi acilmis tamamlanmamis teknoloji varken uyarinin devam etmesi gerekir")
	}
}

func TestPlayerHasResearchableTechsReturnsFalseWhenOnlyLockedTechsRemain(t *testing.T) {
	g := &Game{
		gs: &state.GameState{
			PlayerFactionID: "player",
			Factions: map[faction.FactionID]*faction.Faction{
				"player": {
					ID:       "player",
					Research: faction.ResearchState{Completed: map[string]bool{}},
				},
			},
			TechTypes: map[string]*tech.Technology{
				"advanced": {ID: "advanced", Requires: []string{"basic"}},
			},
		},
	}

	if g.playerHasResearchableTechs() {
		t.Fatalf("sadece kilitli teknolojiler kalmissa gereksiz uyarı gosterilmemeli")
	}
}

func TestEndTurnResearchNeedsConfirmationWhenAutoStartCannotSelectResearch(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {
				ID:   "player",
				Gold: 0,
				Research: faction.ResearchState{
					Completed: map[string]bool{},
				},
			},
		},
		TechTypes: map[string]*tech.Technology{
			"basic": {
				ID:            "basic",
				NameTR:        "Temel Askerlik",
				GoldCost:      20,
				TurnsRequired: 3,
			},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if !g.endTurnResearchNeedsConfirmation() {
		t.Fatal("otomatik arastirma baslatilamadiginda arastirilabilir teknoloji icin onay istenmeliydi")
	}
	if got := gs.Factions["player"].Research.ActiveID; got != "" {
		t.Fatalf("yetersiz altinda otomatik arastirma baslamamaliydi, got=%q", got)
	}
}

func TestEndTurnResearchDoesNotNeedConfirmationWhenAutoStartSucceeds(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {
				ID:   "player",
				Gold: 20,
				Research: faction.ResearchState{
					Completed: map[string]bool{},
				},
			},
		},
		TechTypes: map[string]*tech.Technology{
			"basic": {
				ID:            "basic",
				NameTR:        "Temel Askerlik",
				GoldCost:      20,
				TurnsRequired: 3,
			},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if g.endTurnResearchNeedsConfirmation() {
		t.Fatal("otomatik arastirma basladiginda tur sonu onayi istenmemeliydi")
	}
	if got := gs.Factions["player"].Research.ActiveID; got != "basic" {
		t.Fatalf("otomatik arastirma baslamaliydi, got=%q", got)
	}
}
