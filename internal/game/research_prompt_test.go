package game

import (
	"testing"

	"mapp-game-go/internal/faction"
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
