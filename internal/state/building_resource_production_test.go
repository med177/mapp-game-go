package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/world"
)

func TestRegionProductionSummaryAppliesBuildingResourceModsAndBonuses(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		BuildingTypes: map[string]*city.Building{
			"forge": {
				ID:         "forge",
				IronMod:    1.5,
				IronBonus:  2,
				StoneMod:   2,
				StoneBonus: 3,
				ClothMod:   1.25,
				ClothBonus: 1,
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

func TestRegionProductionSummaryAppliesFactionCouncilResourceBonus(t *testing.T) {
	const owner faction.FactionID = "player"
	region := &world.Region{
		ID: "capital", OwnerID: string(owner), Terrain: world.TerrainPlain,
		BaseGrainOutput: 100, BaseIronOutput: 100, BaseTimberOutput: 100,
		BaseStoneOutput: 100, BaseSpiceOutput: 100, BaseClothOutput: 100,
	}
	buildState := func(withTreasurer bool) *GameState {
		character := &army.Commander{
			ID: "chr_treasurer", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		}
		gs := &GameState{
			Year:          1,
			Factions:      map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
			Commanders:    map[string]*army.Commander{character.ID: character},
			Regions:       map[world.RegionID]*world.Region{region.ID: region},
			CouncilConfig: scenario.DefaultCharacterCouncilConfig(),
		}
		if withTreasurer && !gs.AssignCharacterToCouncilRole(owner, character.ID, CouncilRoleTreasurer) {
			t.Fatal("could not assign treasurer")
		}
		return gs
	}
	base := buildState(false).RegionProductionSummary(region)
	bonus := buildState(true).RegionProductionSummary(region)
	if bonus.Grain != base.Grain*105/100 || bonus.Iron != base.Iron*105/100 ||
		bonus.Timber != base.Timber*105/100 || bonus.Stone != base.Stone*105/100 ||
		bonus.Spice != base.Spice*105/100 || bonus.Cloth != base.Cloth*105/100 {
		t.Fatalf("production with council resource bonus = %+v, base = %+v", bonus, base)
	}
}
