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
	statusText := techResearchStatusText(f, gs.TechTypes)
	_, button := buildTechResearchStatus(techPanelLayoutForScreen(), statusText)

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

func TestTechResearchStatusGroupIsCentered(t *testing.T) {
	layout := techPanelLayoutForScreen()
	statusText := "> Araştırılıyor: Uzun Bir Teknoloji Adı  (12 tur kaldı)"
	textRect, button := buildTechResearchStatus(layout, statusText)
	groupCenter := (textRect.X + button.X + button.W) / 2
	statusCenter := layout.statusRect.X + layout.statusRect.W/2
	if delta := groupCenter - statusCenter; delta < -0.5 || delta > 0.5 {
		t.Fatalf("aktif araştırma ve iptal grubu merkezlenmedi: group=%v status=%v", groupCenter, statusCenter)
	}
	if button.X <= textRect.X+textRect.W {
		t.Fatalf("iptal düğmesi durum metninden sonra gelmeli: text=%+v button=%+v", textRect, button)
	}
}

func TestTechFilterTabsAreNarrowAndCentered(t *testing.T) {
	rect := techPanelLayoutForScreen().filterRect
	categories := tech.AllCategories()
	tabW, startX := techFilterTabDimensions(rect)
	totalW := tabW*float64(len(categories)) + techFilterTabGap*float64(len(categories)-1)
	if tabW >= rect.W/float64(len(categories)) {
		t.Fatalf("filtre sekmeleri alanı dolduracak kadar geniş: tabW=%v rowW=%v", tabW, rect.W)
	}
	if delta := startX + totalW/2 - (rect.X + rect.W/2); delta < -0.5 || delta > 0.5 {
		t.Fatalf("filtre grubu ekran ortasında değil: start=%v width=%v rect=%+v", startX, totalW, rect)
	}
	for idx, category := range categories {
		x := startX + float64(idx)*(tabW+techFilterTabGap) + tabW/2
		got, ok := techFilterTabHit(x, rect.Y+rect.H/2, rect)
		if !ok || got != category {
			t.Fatalf("sekme merkezi yanlış kategoriye gidiyor: idx=%d got=%q ok=%t want=%q", idx, got, ok, category)
		}
	}
}
