package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestNavalContactClashOpensBattlePlan(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"sea": {ID: "sea", IsSea: true, NameTR: "Test Denizi"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player-fleet": {ID: "player-fleet", OwnerID: "p1", RegionID: "sea", IsNaval: true, Units: []army.Unit{{TypeID: "warship"}}},
			"enemy-fleet":  {ID: "enemy-fleet", OwnerID: "p2", RegionID: "sea", IsNaval: true, Units: []army.Unit{{TypeID: "warship"}}},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1", NameTR: "Oyuncu"},
			"p2": {ID: "p2", NameTR: "Düşman"},
		},
		UnitTypes: map[string]*army.UnitType{
			"warship": {ID: "warship", Category: army.CategoryNavalWar},
		},
	}
	r := &Renderer{gs: gs}

	if !r.ShowNavalContactBattlePlan("player-fleet", "enemy-fleet", "sea", false) {
		t.Fatal("temas çatışmaya dönüştüğünde deniz muharebesi planı açılmalı")
	}
	if !r.battlePlan.show || !r.battlePlan.navalContactResolved {
		t.Fatalf("temas sonrası battle plan state'i kurulmadı: %+v", r.battlePlan)
	}
}
