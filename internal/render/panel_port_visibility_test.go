package render

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestBuildingVisibleByRegionRules_PortUsesCoastalPredicate(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"coastal_plain": {
				ID:        "coastal_plain",
				Terrain:   world.TerrainPlain,
				Neighbors: []world.RegionID{"sea_1"},
			},
			"inland_coast": {
				ID:      "inland_coast",
				Terrain: world.TerrainCoast,
			},
			"sea_1": {
				ID:    "sea_1",
				IsSea: true,
			},
		},
	}
	port := &city.Building{ID: "port", RequiredTerrain: "coast"}

	if !buildingVisibleByRegionRules(gs, gs.Regions["coastal_plain"], "port", port) {
		t.Fatalf("denize komsu plain bolgede port gorunur olmali")
	}
	if buildingVisibleByRegionRules(gs, gs.Regions["inland_coast"], "port", port) {
		t.Fatalf("denize komsu olmayan coast terrain bolgede port gorunmemeli")
	}
}
