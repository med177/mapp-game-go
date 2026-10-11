package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestSuccessorStartingUnitTypeSelectsFirstProducibleLandUnit(t *testing.T) {
	successor := &faction.Faction{
		ID: "successor",
		Research: faction.ResearchState{Completed: map[string]bool{
			"iron_weapons": true,
		}},
	}
	region := &world.Region{Buildings: []string{"barracks"}}
	gs := &state.GameState{
		UnitTypeOrder: []string{"militia", "blocked", "naval", "infantry", "cavalry"},
		UnitTypes: map[string]*army.UnitType{
			"militia":  {ID: "militia", Category: army.CategoryInfantry},
			"blocked":  {ID: "blocked", Category: army.CategoryInfantry, AllowedFactions: []string{"another-faction"}},
			"naval":    {ID: "naval", Category: army.CategoryNavalWar},
			"infantry": {ID: "infantry", Category: army.CategoryInfantry, RequiredTech: []string{"iron_weapons"}, RequiredBuildings: []army.BuildingRequirement{{ID: "barracks", Level: 1}}},
			"cavalry":  {ID: "cavalry", Category: army.CategoryCavalry},
		},
	}

	if got := successorStartingUnitType(gs, region, successor); got != "infantry" {
		t.Fatalf("başlangıç birimi = %q, want infantry", got)
	}
}

func TestSuccessorStartingUnitTypeFallsBackWhenNoOtherUnitIsProducible(t *testing.T) {
	successor := &faction.Faction{ID: "successor"}
	region := &world.Region{Buildings: []string{"barracks"}}
	gs := &state.GameState{
		UnitTypeOrder: []string{"militia", "infantry"},
		UnitTypes: map[string]*army.UnitType{
			"militia":  {ID: "militia", Category: army.CategoryInfantry},
			"infantry": {ID: "infantry", Category: army.CategoryInfantry, RequiredTech: []string{"iron_weapons"}},
		},
	}

	if got := successorStartingUnitType(gs, region, successor); got != "militia" {
		t.Fatalf("yedek başlangıç birimi = %q, want militia", got)
	}
}
