package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
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
	if !r.battlePlan.show || !r.battlePlan.contactResolved {
		t.Fatalf("temas sonrası battle plan state'i kurulmadı: %+v", r.battlePlan)
	}
}

func TestLandContactClashOpensBattlePlan(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"land": {ID: "land", NameTR: "Test Bölgesi"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player-army": {ID: "player-army", OwnerID: "p1", RegionID: "land", Units: []army.Unit{{TypeID: "inf"}}},
			"enemy-army":  {ID: "enemy-army", OwnerID: "p2", RegionID: "land", Units: []army.Unit{{TypeID: "inf"}}},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1", NameTR: "Oyuncu"},
			"p2": {ID: "p2", NameTR: "Düşman"},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry},
		},
	}
	r := &Renderer{gs: gs}

	if !r.ShowLandContactBattlePlan("player-army", "enemy-army", "land") {
		t.Fatal("kara temasında çatış seçilince kara muharebesi planı açılmalı")
	}
	if !r.battlePlan.show || !r.battlePlan.contactResolved || r.battlePlan.battleContext != combat.BattleContextLand {
		t.Fatalf("kara temas battle plan state'i kurulmadı: %+v", r.battlePlan)
	}
}

func TestFortifiedLandContactClashOpensBattlePlan(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"fort": {ID: "fort", NameTR: "Test Kalesi", OwnerID: "p2", Buildings: []string{"walls"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player-army": {ID: "player-army", OwnerID: "p1", RegionID: "fort", Units: []army.Unit{{TypeID: "inf"}}},
			"enemy-army":  {ID: "enemy-army", OwnerID: "p2", RegionID: "fort", Units: []army.Unit{{TypeID: "inf"}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry},
		},
	}
	r := &Renderer{gs: gs}

	if !r.ShowLandContactBattlePlan("player-army", "enemy-army", "fort") {
		t.Fatal("tahkimli kara temasında çatış seçilince kara muharebesi planı açılmalı")
	}
	if !r.battlePlan.show || !r.battlePlan.contactResolved || r.battlePlan.battleContext != combat.BattleContextLand {
		t.Fatalf("tahkimli temas sonrası battle plan state'i kurulmadı: %+v", r.battlePlan)
	}
}
