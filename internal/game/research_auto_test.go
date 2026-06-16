package game

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
)

func TestAutoStartResearchIfIdleStartsNextResearchableTech(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {
				ID:   "player",
				Gold: 25,
				Research: faction.ResearchState{
					Completed: map[string]bool{"root": true},
				},
			},
		},
		TechTypes: map[string]*tech.Technology{
			"root": {
				ID:       "root",
				Category: tech.CategoryMilitary,
			},
			"smithing": {
				ID:            "smithing",
				NameTR:        "Demircilik",
				Category:      tech.CategoryMilitary,
				Requires:      []string{"root"},
				GoldCost:      20,
				TurnsRequired: 4,
			},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if started := g.autoStartResearchIfIdle(); !started {
		t.Fatal("uygun teknoloji varken otomatik baslatma calismaliydi")
	}

	player := gs.Factions["player"]
	if player.Research.ActiveID != "smithing" {
		t.Fatalf("beklenen aktif research smithing olmaliydi, got=%q", player.Research.ActiveID)
	}
	if player.Research.TurnsLeft != 4 {
		t.Fatalf("turn sayaci teknoloji turu kadar ayarlanmaliydi, got=%d", player.Research.TurnsLeft)
	}
	if player.Gold != 5 {
		t.Fatalf("gold otomatik baslatmada dusmeliydi, got=%d", player.Gold)
	}
}
