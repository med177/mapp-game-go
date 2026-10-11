package render

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	gameui "mapp-game-go/internal/ui"
)

func TestTechCategoryIconsUseCategoryAssets(t *testing.T) {
	tests := []struct {
		category tech.Category
		want     gameui.IconID
	}{
		{category: tech.CategoryMilitary, want: gameui.IconSword},
		{category: tech.CategoryEconomy, want: gameui.IconGold},
		{category: tech.CategoryDiplomacy, want: gameui.IconDiplomacy},
		{category: tech.CategoryNaval, want: gameui.IconNaval},
		{category: tech.CategoryReligion, want: gameui.IconReligion},
	}

	for _, tt := range tests {
		if got := techCategoryIcon(tt.category); got != tt.want {
			t.Errorf("%s kategori ikonu = %q, want %q", tt.category, got, tt.want)
		}
	}
}

func TestCountTechByCategoryExcludesFactionRestrictedTechnologies(t *testing.T) {
	f := &faction.Faction{ID: "stark", Research: faction.ResearchState{Completed: map[string]bool{
		"available":  true,
		"restricted": true,
	}}}
	allTechs := map[string]*tech.Technology{
		"available": {ID: "available", Category: tech.CategoryMilitary},
		"restricted": {
			ID:              "restricted",
			Category:        tech.CategoryMilitary,
			AllowedFactions: []string{"lannister"},
		},
	}

	completed, total := countTechByCategory(f, tech.CategoryMilitary, allTechs)
	if completed != 1 || total != 1 {
		t.Fatalf("factiona kapalı teknoloji sayılmamalı: completed=%d total=%d", completed, total)
	}
}

func TestTechCancelResearchButtonDispatchesCancelAction(t *testing.T) {
	f := &faction.Faction{
		ID: "player",
		Research: faction.ResearchState{
			ActiveID:  "active",
			TurnsLeft: 3,
		},
	}
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions:        map[faction.FactionID]*faction.Faction{"player": f},
		TechTypes: map[string]*tech.Technology{
			"active": {ID: "active", NameTR: "Aktif", Category: tech.CategoryMilitary, TurnsRequired: 4},
		},
	}
	r := &Renderer{gs: gs}
	button := buildTechCancelResearchButton(techPanelLayoutForScreen())

	action := r.handleTechInput(f, gameui.InputState{
		MouseX:          button.X + button.W/2,
		MouseY:          button.Y + button.H/2,
		LeftJustPressed: true,
	})
	if action.Kind != ActionCancelResearch {
		t.Fatalf("iptal düğmesi action = %q, want %q", action.Kind, ActionCancelResearch)
	}
	if !r.techPanelPointerHit(button.X+button.W/2, button.Y+button.H/2) {
		t.Fatal("iptal düğmesi cursor hit-test tarafından etkileşimli tanınmalı")
	}
}
