package state

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestRegionProductionSummaryAppliesBuildingResourceModsAndBonuses(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		BuildingTypes: map[string]*city.Building{
			"forge": {
				ID:          "forge",
				IronMod:     1.5,
				IronBonus:   2,
				StoneMod:    2,
				StoneBonus:  3,
				ClothMod:    1.25,
				ClothBonus:  1,
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"mine": {
				ID:              "mine",
				OwnerID:         "player",
				Terrain:         world.TerrainPlain,
				BaseIronOutput:  10,
				BaseStoneOutput: 4,
				BaseClothOutput: 8,
				Buildings:       []string{"forge"},
			},
		},
	}

	production := gs.RegionProductionSummary(gs.Regions["mine"])
	if production.Iron != 17 {
		t.Fatalf("demir bina etkisi uygulanmadı: got=%d want=17", production.Iron)
	}
	if production.Stone != 11 {
		t.Fatalf("taş bina etkisi uygulanmadı: got=%d want=11", production.Stone)
	}
	if production.Cloth != 11 {
		t.Fatalf("kumaş bina etkisi uygulanmadı: got=%d want=11", production.Cloth)
	}
}
