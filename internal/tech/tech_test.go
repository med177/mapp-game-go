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

func TestIsUnlockedForContextRequiresYearAndRegions(t *testing.T) {
	rs := faction.ResearchState{Completed: map[string]bool{"gunpowder": true}}
	target := &Technology{
		ID:              "cast_bronze_cannon",
		Requires:        []string{"gunpowder"},
		RequiredRegions: []string{"thrace", "bursa"},
		MinYear:         1420,
	}

	if IsUnlockedForContext(&rs, target, 1419, map[string]bool{"thrace": true, "bursa": true}) {
		t.Fatal("minimum yıldan önce teknoloji açılmamalı")
	}
	if IsUnlockedForContext(&rs, target, 1420, map[string]bool{"thrace": true}) {
		t.Fatal("eksik bölge varken teknoloji açılmamalı")
	}
	if !IsUnlockedForContext(&rs, target, 1420, map[string]bool{"thrace": true, "bursa": true}) {
		t.Fatal("yıl ve tüm bölgeler sağlandığında teknoloji açılmalı")
	}
}

func TestNextResearchableTechIDForContextSkipsUnavailableTech(t *testing.T) {
	rs := faction.ResearchState{Completed: map[string]bool{}}
	allTechs := map[string]*Technology{
		"early": {ID: "early", GoldCost: 1},
		"late":  {ID: "late", GoldCost: 1, MinYear: 1420, RequiredRegions: []string{"bursa"}},
	}

	if id, ok := NextResearchableTechIDForContext(&rs, allTechs, 1, 1419, map[string]bool{"bursa": true}); !ok || id != "early" {
		t.Fatalf("erken teknoloji seçilmeliydi: id=%q ok=%v", id, ok)
	}
	if id, ok := NextResearchableTechIDForContext(&rs, allTechs, 1, 1420, map[string]bool{}); ok && id == "late" {
		t.Fatal("bölge şartı sağlanmayan teknoloji seçilmemeli")
	}
}
