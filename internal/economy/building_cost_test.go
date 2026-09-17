package economy

import (
	"testing"

	"mapp-game-go/internal/city"
)

func TestBuildingCostAtLevelCompoundsAllResources(t *testing.T) {
	building := &city.Building{
		GoldCost:              90,
		StoneCost:             800,
		UpgradeCostMultiplier: 1.25,
	}
	wantGold := []int{90, 113, 141, 176, 220}
	wantStone := []int{800, 1000, 1250, 1563, 1953}
	for level := 1; level <= len(wantGold); level++ {
		cost := BuildingCostAtLevel(building, level)
		if cost.Gold != wantGold[level-1] || cost.Stone != wantStone[level-1] {
			t.Fatalf("level %d cost = gold %d, stone %d; want gold %d, stone %d", level, cost.Gold, cost.Stone, wantGold[level-1], wantStone[level-1])
		}
	}
}

func TestBuildingCostAtLevelKeepsLegacyBaseCostWithoutMultiplier(t *testing.T) {
	building := &city.Building{GoldCost: 90, StoneCost: 800}
	for _, level := range []int{0, 1, 3} {
		cost := BuildingCostAtLevel(building, level)
		if cost.Gold != 90 || cost.Stone != 800 {
			t.Fatalf("level %d cost = %+v; missing multiplier should keep base cost", level, cost)
		}
	}
}
