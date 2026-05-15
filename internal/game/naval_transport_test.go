package game

import (
	"math/rand"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestMoveArmyEmbarkSuccess(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"land_a": {ID: "land_a", OwnerID: "p1", Neighbors: []world.RegionID{"sea_1"}},
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_a", "land_b"}},
			"land_b": {ID: "land_b", OwnerID: "p1", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_p1_1": {
				ID:            "army_p1_1",
				OwnerID:       "p1",
				RegionID:      "land_a",
				Units:         []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
			},
			"fleet_p1_1": {
				ID:            "fleet_p1_1",
				OwnerID:       "p1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry":  {ID: "infantry", Embarkable: true},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("army_p1_1", "sea_1")

	if _, exists := gs.Armies["army_p1_1"]; exists {
		t.Fatalf("kara ordusu embark sonrası silinmeliydi")
	}
	fleet := gs.Armies["fleet_p1_1"]
	if fleet == nil || len(fleet.EmbarkedUnits) != 1 {
		t.Fatalf("filoda tek embark birimi beklenirdi, got=%+v", fleet)
	}
	if fleet.MovePoints != 2 {
		t.Fatalf("filo hareket puanı 1 düşmeliydi, got=%d", fleet.MovePoints)
	}
}

func TestMoveArmyEmbarkRejectsNonEmbarkableUnits(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"land_a": {ID: "land_a", OwnerID: "p1", Neighbors: []world.RegionID{"sea_1"}},
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_a"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_p1_1": {
				ID:            "army_p1_1",
				OwnerID:       "p1",
				RegionID:      "land_a",
				Units:         []army.Unit{{TypeID: "cavalry", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
			},
			"fleet_p1_1": {
				ID:            "fleet_p1_1",
				OwnerID:       "p1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"cavalry":   {ID: "cavalry", Embarkable: false},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("army_p1_1", "sea_1")

	if _, exists := gs.Armies["army_p1_1"]; !exists {
		t.Fatalf("embark reddinde kara ordusu silinmemeliydi")
	}
	if len(gs.Armies["fleet_p1_1"].EmbarkedUnits) != 0 {
		t.Fatalf("embark reddinde filoya birim yüklenmemeliydi")
	}
}

func TestMoveArmyDisembarkSuccess(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     7,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_a"}},
			"land_a": {ID: "land_a", OwnerID: "p1", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_p1_1": {
				ID:            "fleet_p1_1",
				OwnerID:       "p1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry":  {ID: "infantry", Embarkable: true},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "land_a")

	fleet := gs.Armies["fleet_p1_1"]
	if len(fleet.EmbarkedUnits) != 0 {
		t.Fatalf("çıkarma sonrası filo cargo'su boş olmalı")
	}
	if fleet.MovePoints != 2 {
		t.Fatalf("çıkarma sonrası filo hareket puanı 1 düşmeli, got=%d", fleet.MovePoints)
	}
	newArmy, ok := gs.Armies["army_p1_8"]
	if !ok {
		t.Fatalf("çıkarma sonrası yeni kara ordusu beklenirdi")
	}
	if newArmy.RegionID != "land_a" || newArmy.IsNaval || len(newArmy.Units) != 1 {
		t.Fatalf("çıkarma sonucu ordu hatalı: %+v", newArmy)
	}
}

func TestMoveArmyDisembarkEnemyCoastRequiresWar(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     3,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_e"}},
			"land_e": {ID: "land_e", OwnerID: "p2", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_p1_1": {
				ID:            "fleet_p1_1",
				OwnerID:       "p1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
			"p2": {ID: "p2"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p1", "p2"): {FactionA: "p1", FactionB: "p2", Score: -10, Stance: faction.StancePeace},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry":  {ID: "infantry", Embarkable: true},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "land_e")

	fleet := gs.Armies["fleet_p1_1"]
	if len(fleet.EmbarkedUnits) != 1 {
		t.Fatalf("savaş yokken çıkarma olmamalıydı, cargo korunmalı")
	}
	if fleet.MovePoints != 3 {
		t.Fatalf("savaş yokken hareket puanı düşmemeli, got=%d", fleet.MovePoints)
	}
	if _, ok := gs.Armies["army_p1_4"]; ok {
		t.Fatalf("savaş yokken yeni kara ordusu oluşmamalıydı")
	}
}

func TestMoveArmyDisembarkEnemyArmyBattleWin(t *testing.T) {
	rand.Seed(1)
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     11,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_e"}},
			"land_e": {ID: "land_e", OwnerID: "p2", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_p1_1": {
				ID:            "fleet_p1_1",
				OwnerID:       "p1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "elite", CurrentHP: 100}, {TypeID: "elite", CurrentHP: 100}, {TypeID: "elite", CurrentHP: 100}, {TypeID: "elite", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
			"enemy_army": {
				ID:            "enemy_army",
				OwnerID:       "p2",
				RegionID:      "land_e",
				Units:         []army.Unit{{TypeID: "weak", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
			"p2": {ID: "p2"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p1", "p2"): {FactionA: "p1", FactionB: "p2", Score: -90, Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans},
			"elite":     {ID: "elite", Embarkable: true, Attack: 100, Defense: 100, Morale: 100},
			"weak":      {ID: "weak", Attack: 1, Defense: 1, Morale: 1},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "land_e")

	if _, ok := gs.Armies["enemy_army"]; ok {
		t.Fatalf("kazanılan çıkarma savaşında düşman ordusu silinmeliydi")
	}
	if gs.Regions["land_e"].OwnerID != "p1" {
		t.Fatalf("başarılı çıkarma sonrası bölge ele geçirilmeli, got=%s", gs.Regions["land_e"].OwnerID)
	}
	if _, ok := gs.Armies["army_p1_12"]; !ok {
		t.Fatalf("başarılı çıkarma sonrası yeni kara ordusu bekleniyordu")
	}
	if len(gs.Armies["fleet_p1_1"].EmbarkedUnits) != 0 {
		t.Fatalf("savaş sonrası filo cargo'su boş olmalı")
	}
}

func TestMoveArmyDisembarkEnemyArmyBattleLose(t *testing.T) {
	rand.Seed(2)
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     21,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_e"}},
			"land_e": {ID: "land_e", OwnerID: "p2", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_p1_1": {
				ID:            "fleet_p1_1",
				OwnerID:       "p1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "weak", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
			"enemy_army": {
				ID:            "enemy_army",
				OwnerID:       "p2",
				RegionID:      "land_e",
				Units:         []army.Unit{{TypeID: "elite", CurrentHP: 100}, {TypeID: "elite", CurrentHP: 100}, {TypeID: "elite", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
			"p2": {ID: "p2"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p1", "p2"): {FactionA: "p1", FactionB: "p2", Score: -90, Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans},
			"elite":     {ID: "elite", Attack: 100, Defense: 100, Morale: 100},
			"weak":      {ID: "weak", Embarkable: true, Attack: 1, Defense: 1, Morale: 1},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "land_e")

	if gs.Regions["land_e"].OwnerID != "p2" {
		t.Fatalf("başarısız çıkarma sonrası sahiplik değişmemeli, got=%s", gs.Regions["land_e"].OwnerID)
	}
	if _, ok := gs.Armies["army_p1_22"]; ok {
		t.Fatalf("başarısız çıkarma sonrası kara ordusu oluşmamalı")
	}
	if len(gs.Armies["fleet_p1_1"].EmbarkedUnits) != 0 {
		t.Fatalf("başarısız çıkarma sonrası cargo tüketilmeli")
	}
}
