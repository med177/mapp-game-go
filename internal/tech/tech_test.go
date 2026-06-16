package tech

import (
	"testing"

	"mapp-game-go/internal/faction"
)

func TestNextResearchableTechIDPrefersUnlockedAndShallowerTechs(t *testing.T) {
	rs := faction.ResearchState{
		Completed: map[string]bool{"root": true},
	}
	allTechs := map[string]*Technology{
		"root": {
			ID:       "root",
			Category: CategoryMilitary,
		},
		"smithing": {
			ID:       "smithing",
			Category: CategoryMilitary,
			Requires: []string{"root"},
			GoldCost: 10,
		},
		"trade": {
			ID:       "trade",
			Category: CategoryEconomy,
			Requires: []string{"root"},
			GoldCost: 10,
		},
		"advanced_siege": {
			ID:       "advanced_siege",
			Category: CategoryMilitary,
			Requires: []string{"smithing"},
			GoldCost: 10,
		},
	}

	id, ok := NextResearchableTechID(&rs, allTechs, 10)
	if !ok {
		t.Fatal("araştırılabilir bir teknoloji seçilebilmeliydi")
	}
	if id != "smithing" {
		t.Fatalf("beklenen ilk aday smithing idi, got=%q", id)
	}
}

func TestNextResearchableTechIDSkipsPausedResearch(t *testing.T) {
	rs := faction.ResearchState{
		PausedTurns: map[string]int{"smithing": 2},
		Completed:   map[string]bool{"root": true},
	}
	allTechs := map[string]*Technology{
		"root": {
			ID:       "root",
			Category: CategoryMilitary,
		},
		"smithing": {
			ID:       "smithing",
			Category: CategoryMilitary,
			Requires: []string{"root"},
			GoldCost: 10,
		},
	}

	if id, ok := NextResearchableTechID(&rs, allTechs, 10); ok || id != "" {
		t.Fatalf("paused research varken otomatik secim yapilmamali, got=%q ok=%v", id, ok)
	}
}
