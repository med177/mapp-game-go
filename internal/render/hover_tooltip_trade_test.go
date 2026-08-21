package render

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
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

func TestPortBuildingEffectLinesShowNavalCapacityIncrease(t *testing.T) {
	gs := &state.GameState{Regions: map[world.RegionID]*world.Region{
		"port": {ID: "port", OwnerID: "p1", Population: 1000, Buildings: []string{"port"}},
	}}
	lines := buildingNavalCapacityEffectLines(gs, gs.Regions["port"], &city.Building{ID: "port", MaxPerRegion: 3})

	if len(lines) != 2 || lines[0] != "Donanma kapasitesi: +2 gemi" || lines[1] != "Donanma sınırı: 4 → 6" {
		t.Fatalf("liman tooltip'i kapasite artışını ve yeni sınırı göstermeli: %v", lines)
	}
}

func TestBarracksBuildingEffectLinesShowLandCapacityIncrease(t *testing.T) {
	gs := &state.GameState{Regions: map[world.RegionID]*world.Region{
		"barracks": {ID: "barracks", OwnerID: "p1", Buildings: []string{"barracks"}},
	}}
	lines := buildingLandCapacityEffectLines(gs, gs.Regions["barracks"], &city.Building{ID: "barracks", MaxPerRegion: 3})

	if len(lines) != 2 || lines[0] != "Savaşçı sınırı: 20 (1 ordu × 20)" || lines[1] != "Kışla üretim limiti: 1 → 2 birim/tur" {
		t.Fatalf("kışla tooltip'i ordu başına savaşçı sınırını ve üretim etkisini göstermeli: %v", lines)
	}
}
