package render

import (
	"testing"

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
