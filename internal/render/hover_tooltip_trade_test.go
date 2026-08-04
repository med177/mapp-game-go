package render

import (
	"testing"

	"mapp-game-go/internal/city"
)

func TestBuildingEffectLinesIncludeTradeCapacityModifier(t *testing.T) {
	lines := buildingEffectLines(&city.Building{ID: "granary", TradeCapacityMod: 1.05})
	for _, line := range lines {
		if line == "Ticaret kapasitesi: x1.05" {
			return
		}
	}
	t.Fatalf("ticaret kapasitesi etkisi tooltip satırlarında görünmeli: %v", lines)
}
