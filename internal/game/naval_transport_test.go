package game

import (
	"math/rand"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func testTransportType() *army.UnitType {
	return &army.UnitType{ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10}
}

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
			"transport": testTransportType(),
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

func TestMoveArmyEmbarkTransfersLandCommanderInsteadOfFleetCommander(t *testing.T) {
	landCommander := army.NewCommander("cmd_land", "Kara Komutanı")
	fleetCommander := army.NewCommander("cmd_fleet", "Donanma Komutanı")
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"land_a": {ID: "land_a", OwnerID: "p1", Neighbors: []world.RegionID{"sea_1"}},
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_a"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_p1_1": {
				ID: "army_p1_1", OwnerID: "p1", RegionID: "land_a",
				Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}}, MovePoints: 2, MaxMovePoints: 2,
				Commander: landCommander,
			},
			"fleet_p1_1": {
				ID: "fleet_p1_1", OwnerID: "p1", RegionID: "sea_1",
				Units: []army.Unit{{TypeID: "transport", CurrentHP: 100}}, MovePoints: 3, MaxMovePoints: 3,
				IsNaval: true, Commander: fleetCommander,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{"p1": {ID: "p1"}},
		UnitTypes: map[string]*army.UnitType{
			"infantry":  {ID: "infantry", Embarkable: true},
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("army_p1_1", "sea_1")

	fleet := gs.Armies["fleet_p1_1"]
	if fleet.EmbarkedCommander != landCommander {
		t.Fatalf("kara komutanı filoda taşınmalıydı, got=%v", fleet.EmbarkedCommander)
	}
	if fleet.Commander != fleetCommander {
		t.Fatalf("filo komutanı filoda kalmalıydı, got=%v", fleet.Commander)
	}
	if gs.AmphibiousCommander(fleet.ID) != landCommander {
		t.Fatal("çıkarma komutanı taşınan kara komutanı olmalıydı")
	}
}

func TestMoveArmyEmbarkRejectsInsufficientTransportCapacity(t *testing.T) {
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
				Units:         []army.Unit{{TypeID: "infantry", CurrentHP: 100}, {TypeID: "infantry", CurrentHP: 100}},
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
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 1},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("army_p1_1", "sea_1")

	if _, exists := gs.Armies["army_p1_1"]; !exists {
		t.Fatalf("kapasite yetmiyorsa kara ordusu silinmemeliydi")
	}
	if got := len(gs.Armies["fleet_p1_1"].EmbarkedUnits); got != 0 {
		t.Fatalf("kapasite yetmiyorsa filoya yük alınmamalı, got=%d", got)
	}
}

func TestMoveArmyEmbarkAppendsIntoExistingFleetCargo(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"land_a": {ID: "land_a", OwnerID: "p1", Neighbors: []world.RegionID{"sea_1"}},
			"land_b": {ID: "land_b", OwnerID: "p1", Neighbors: []world.RegionID{"sea_1"}},
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_a", "land_b"}},
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
			"army_p1_2": {
				ID:            "army_p1_2",
				OwnerID:       "p1",
				RegionID:      "land_b",
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
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 2},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("army_p1_1", "sea_1")
	g.moveArmy("army_p1_2", "sea_1")

	fleet := gs.Armies["fleet_p1_1"]
	if got := len(fleet.EmbarkedUnits); got != 2 {
		t.Fatalf("iki embark sonrası filoda iki kara birimi beklenirdi, got=%d", got)
	}
	if _, exists := gs.Armies["army_p1_1"]; exists {
		t.Fatalf("ilk embark eden ordu silinmeliydi")
	}
	if _, exists := gs.Armies["army_p1_2"]; exists {
		t.Fatalf("ikinci embark eden ordu silinmeliydi")
	}
}

func TestEmbarkArmyOntoSpecificDockedFleet(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"land_a": {ID: "land_a", OwnerID: "p1", Neighbors: []world.RegionID{"sea_1"}},
			"land_b": {ID: "land_b", OwnerID: "p1", Neighbors: []world.RegionID{"sea_1"}},
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_a", "land_b"}},
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
			"fleet_here": {
				ID:                 "fleet_here",
				OwnerID:            "p1",
				RegionID:           "sea_1",
				DockedRegionID:     "land_a",
				DockedSettlementID: "port_a",
				Units:              []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				MovePoints:         3,
				MaxMovePoints:      3,
				IsNaval:            true,
			},
			"fleet_elsewhere": {
				ID:                 "fleet_elsewhere",
				OwnerID:            "p1",
				RegionID:           "sea_1",
				DockedRegionID:     "land_b",
				DockedSettlementID: "port_b",
				Units:              []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				MovePoints:         3,
				MaxMovePoints:      3,
				IsNaval:            true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry":  {ID: "infantry", Embarkable: true},
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.embarkArmyOntoFleet("army_p1_1", "fleet_here")

	if _, exists := gs.Armies["army_p1_1"]; exists {
		t.Fatalf("kara ordusu belirli filoya bindikten sonra silinmeliydi")
	}
	if got := len(gs.Armies["fleet_here"].EmbarkedUnits); got != 1 {
		t.Fatalf("secilen filoda bir cargo birimi beklenirdi, got=%d", got)
	}
	if got := len(gs.Armies["fleet_elsewhere"].EmbarkedUnits); got != 0 {
		t.Fatalf("diger filo etkilenmemeliydi, got=%d", got)
	}
}

func TestMoveArmyEmbarkPrefersFleetDockedAtSourceRegion(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"land_a": {ID: "land_a", OwnerID: "p1", Neighbors: []world.RegionID{"sea_1"}},
			"land_b": {ID: "land_b", OwnerID: "p1", Neighbors: []world.RegionID{"sea_1"}},
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_a", "land_b"}},
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
			"fleet_here": {
				ID:                 "fleet_here",
				OwnerID:            "p1",
				RegionID:           "sea_1",
				DockedRegionID:     "land_a",
				DockedSettlementID: "port_a",
				Units:              []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				MovePoints:         3,
				MaxMovePoints:      3,
				IsNaval:            true,
			},
			"fleet_elsewhere": {
				ID:                 "fleet_elsewhere",
				OwnerID:            "p1",
				RegionID:           "sea_1",
				DockedRegionID:     "land_b",
				DockedSettlementID: "port_b",
				Units:              []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				MovePoints:         3,
				MaxMovePoints:      3,
				IsNaval:            true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry":  {ID: "infantry", Embarkable: true},
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("army_p1_1", "sea_1")

	if got := len(gs.Armies["fleet_here"].EmbarkedUnits); got != 1 {
		t.Fatalf("aynı limandaki filo tercih edilmeliydi, got=%d", got)
	}
	if got := len(gs.Armies["fleet_elsewhere"].EmbarkedUnits); got != 0 {
		t.Fatalf("uzaktaki docked filo cargo almamaliydi, got=%d", got)
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
			"transport": testTransportType(),
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
			"transport": testTransportType(),
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
			"transport": testTransportType(),
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
			"transport": testTransportType(),
			"elite":     {ID: "elite", Embarkable: true, Attack: 100, Defense: 100, Morale: 100},
			"weak":      {ID: "weak", Attack: 1, Defense: 1, Morale: 1},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "land_e")

	if _, ok := gs.Armies["enemy_army"]; ok {
		t.Fatalf("kazanılan çıkarma savaşında düşman ordusu silinmeliydi")
	}
	if gs.Regions["land_e"].OwnerID != "p2" {
		t.Fatalf("son bölgede karar öncesi sahiplik korunmalıydı, got=%s", gs.Regions["land_e"].OwnerID)
	}
	if _, ok := gs.Armies["army_p1_12"]; !ok {
		t.Fatalf("başarılı çıkarma sonrası yeni kara ordusu bekleniyordu")
	}
	if len(gs.Armies["fleet_p1_1"].EmbarkedUnits) != 0 {
		t.Fatalf("savaş sonrası filo cargo'su boş olmalı")
	}
	if len(g.pendingConquestDecisions) != 1 {
		t.Fatalf("son bölgede savaş sonrası karar beklenmeliydi, got=%d", len(g.pendingConquestDecisions))
	}
	g.resolvePendingConquestDecision(false)
	if gs.Regions["land_e"].OwnerID != "p1" {
		t.Fatalf("ilhak kararı sonrası bölge ele geçirilmeli, got=%s", gs.Regions["land_e"].OwnerID)
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
			"transport": testTransportType(),
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

func TestMoveArmyDisembarkEnemyCoastNoArmyConquersOnWar(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     30,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_e"}},
			"land_e": {ID: "land_e", OwnerID: "p2", Religion: "catholic", Neighbors: []world.RegionID{"sea_1"}},
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
			"p1": {ID: "p1", Religion: "sunni"},
			"p2": {ID: "p2", Religion: "catholic"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p1", "p2"): {FactionA: "p1", FactionB: "p2", Score: -90, Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry":  {ID: "infantry", Embarkable: true},
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "land_e")

	if gs.Regions["land_e"].OwnerID != "p2" {
		t.Fatalf("son bölgede karar öncesi sahiplik korunmalıydı, got=%s", gs.Regions["land_e"].OwnerID)
	}
	if _, ok := gs.Armies["army_p1_31"]; !ok {
		t.Fatalf("çıkarma sonrası kara ordusu oluşmalıydı")
	}
	if len(gs.Armies["fleet_p1_1"].EmbarkedUnits) != 0 {
		t.Fatalf("çıkarma sonrası filo cargo'su boş olmalı")
	}
	if len(g.pendingConquestDecisions) != 1 {
		t.Fatalf("son bölgede savaş sonrası karar beklenmeliydi, got=%d", len(g.pendingConquestDecisions))
	}
	g.resolvePendingConquestDecision(false)
	if gs.Regions["land_e"].OwnerID != "p1" {
		t.Fatalf("ilhak kararı sonrası sahiplik değişmeliydi, got=%s", gs.Regions["land_e"].OwnerID)
	}
}

func fortifiedDisembarkTestState(withDefender bool) *state.GameState {
	landingCommander := army.NewCommander("cmd_land", "Çıkarma Komutanı")
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     30,
		Regions: map[world.RegionID]*world.Region{
			"sea_1": {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"fort"}},
			"fort": {
				ID: "fort", NameTR: "Kale Limanı", OwnerID: "p2", Neighbors: []world.RegionID{"sea_1"},
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fortress", Type: world.SettlementFortress, NameTR: "Kale"}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_p1_1": {
				ID: "fleet_p1_1", OwnerID: "p1", RegionID: "sea_1",
				Units:             []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits:     []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
				EmbarkedCommander: landingCommander,
				MovePoints:        3, MaxMovePoints: 3, IsNaval: true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"}, "p2": {ID: "p2"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p1", "p2"): {FactionA: "p1", FactionB: "p2", Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry":  {ID: "infantry", Embarkable: true, Attack: 10, Defense: 10, Morale: 50},
			"transport": testTransportType(),
		},
	}
	if withDefender {
		gs.Armies["defender"] = &army.Army{
			ID: "defender", OwnerID: "p2", RegionID: "fort",
			Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}}, MovePoints: 2, MaxMovePoints: 2,
		}
	}
	return gs
}

func TestMoveArmyDisembarkEnemyFortressStartsSiegeWithoutBattle(t *testing.T) {
	gs := fortifiedDisembarkTestState(true)
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "fort")

	siege := gs.SiegeAt("fort")
	if siege == nil || siege.AttackerArmyID != "army_p1_31" {
		t.Fatalf("kale kıyısı çıkarması aktif kuşatma oluşturmalıydı, got=%+v", siege)
	}
	if gs.Regions["fort"].OwnerID != "p2" {
		t.Fatalf("kuşatma başlarken bölge sahibi değişmemeli, got=%s", gs.Regions["fort"].OwnerID)
	}
	if _, ok := gs.Armies["defender"]; !ok {
		t.Fatal("kale savunucusu aynı hamlede savaşa sokulup silinmemeli")
	}
	landed := gs.Armies["army_p1_31"]
	if landed == nil || landed.Commander == nil || landed.Commander.Name != "Çıkarma Komutanı" {
		t.Fatalf("karaya çıkan ordu kara komutanını taşımalıydı, got=%+v", landed)
	}
	if len(gs.Armies["fleet_p1_1"].EmbarkedUnits) != 0 || len(g.pendingConquestDecisions) != 0 {
		t.Fatal("kuşatma çıkarma sırasında cargo veya fetih kararı kalmamalı")
	}
}

func TestMoveArmyDisembarkEnemyFortressWithoutDefenderStillStartsSiege(t *testing.T) {
	gs := fortifiedDisembarkTestState(false)
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "fort")

	if gs.SiegeAt("fort") == nil {
		t.Fatal("savunmasız kale kıyısında bile kuşatma state'i başlamalıydı")
	}
	if gs.Regions["fort"].OwnerID != "p2" {
		t.Fatalf("savunmasız kale aynı hamlede fethedilmemeli, got=%s", gs.Regions["fort"].OwnerID)
	}
	if len(g.pendingConquestDecisions) != 0 {
		t.Fatal("savunmasız kale çıkarma sırasında fetih kararı açılmamalı")
	}
}

func TestMoveArmyDisembarkNeutralCoastClaimsRegion(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     30,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_n"}},
			"land_n": {ID: "land_n", Religion: "catholic", Neighbors: []world.RegionID{"sea_1"}},
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
			"p1": {ID: "p1", Religion: "sunni"},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry":  {ID: "infantry", Embarkable: true},
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "land_n")

	if gs.Regions["land_n"].OwnerID != "p1" {
		t.Fatalf("sahipsiz kıyı çıkarmasında sahiplik değişmeliydi, got=%s", gs.Regions["land_n"].OwnerID)
	}
	if _, ok := gs.Armies["army_p1_31"]; !ok {
		t.Fatalf("çıkarma sonrası kara ordusu oluşmalıydı")
	}
	if len(gs.Armies["fleet_p1_1"].EmbarkedUnits) != 0 {
		t.Fatalf("çıkarma sonrası filo cargo'su boş olmalı")
	}
}

func TestMoveNavalIntoEnemySeaAtPeaceNoWarDeclarationNoBattle(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"sea_a": {ID: "sea_a", IsSea: true, OwnerID: "p1", Neighbors: []world.RegionID{"sea_b"}},
			"sea_b": {ID: "sea_b", IsSea: true, OwnerID: "p2", Neighbors: []world.RegionID{"sea_a"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_p1": {
				ID:            "fleet_p1",
				OwnerID:       "p1",
				RegionID:      "sea_a",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
			"fleet_p2": {
				ID:            "fleet_p2",
				OwnerID:       "p2",
				RegionID:      "sea_b",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
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
			faction.RelationKey("p1", "p2"): {FactionA: "p1", FactionB: "p2", Score: 10, Stance: faction.StancePeace},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10, Attack: 10, Defense: 10, Morale: 50},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1", "sea_b")

	if gs.Armies["fleet_p1"].RegionID != "sea_b" {
		t.Fatalf("barışta donanma düşman deniz bölgesine girebilmeli, got=%s", gs.Armies["fleet_p1"].RegionID)
	}
	if gs.Armies["fleet_p1"].MovePoints != 2 {
		t.Fatalf("hareket sonrası move point 1 düşmeli, got=%d", gs.Armies["fleet_p1"].MovePoints)
	}
	if _, ok := gs.Armies["fleet_p2"]; !ok {
		t.Fatalf("barışta düşman donanma ile karşılaşmada savaş olmamalı, fleet_p2 silinmemeli")
	}
}

func TestMoveArmyUndockFleetToSeaCenterOnSameRegion(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"land_a": {ID: "land_a", OwnerID: "p1", Neighbors: []world.RegionID{"sea_1"}},
			"sea_1":  {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"land_a"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_p1_1": {
				ID:                 "fleet_p1_1",
				OwnerID:            "p1",
				RegionID:           "sea_1",
				DockedRegionID:     "land_a",
				DockedSettlementID: "port_a",
				Units:              []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				MovePoints:         3,
				MaxMovePoints:      3,
				IsNaval:            true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "sea_1")

	fleet := gs.Armies["fleet_p1_1"]
	if fleet.DockedRegionID != "" || fleet.DockedSettlementID != "" {
		t.Fatalf("undock sonrası liman bağı temizlenmeli, docked_region=%q docked_settlement=%q", fleet.DockedRegionID, fleet.DockedSettlementID)
	}
	if fleet.RegionID != "sea_1" {
		t.Fatalf("undock sonrası filo aynı deniz bölgesinde kalmalı, got=%s", fleet.RegionID)
	}
	if fleet.MovePoints != 2 {
		t.Fatalf("undock hareketi 1 puan tüketmeli, got=%d", fleet.MovePoints)
	}
}

func TestMoveArmyDockFleetAtOwnedPort(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"sea_1": {
				ID:        "sea_1",
				IsSea:     true,
				Neighbors: []world.RegionID{"land_a"},
			},
			"land_a": {
				ID:          "land_a",
				OwnerID:     "p1",
				Neighbors:   []world.RegionID{"sea_1"},
				Buildings:   []string{"port"},
				Settlements: []world.Settlement{{ID: "port_a", Type: world.SettlementPort, NameTR: "Liman"}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
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
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "land_a")

	fleet := gs.Armies["fleet_p1_1"]
	if fleet.RegionID != "sea_1" {
		t.Fatalf("dock sonrası filo deniz bölgesinde kalmalı, got=%s", fleet.RegionID)
	}
	if fleet.DockedRegionID != "land_a" || fleet.DockedSettlementID != "port_a" {
		t.Fatalf("filo limana bağlanmalı, got docked_region=%q docked_settlement=%q", fleet.DockedRegionID, fleet.DockedSettlementID)
	}
	if fleet.MovePoints != 2 {
		t.Fatalf("dock hareketi 1 puan tüketmeli, got=%d", fleet.MovePoints)
	}
}

func TestForceDisembarkFleetSkipsDockAndLandsArmy(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     30,
		Regions: map[world.RegionID]*world.Region{
			"sea_1": {
				ID:        "sea_1",
				IsSea:     true,
				Neighbors: []world.RegionID{"land_a"},
			},
			"land_a": {
				ID:          "land_a",
				OwnerID:     "p1",
				Neighbors:   []world.RegionID{"sea_1"},
				Buildings:   []string{"port"},
				Settlements: []world.Settlement{{ID: "port_a", Type: world.SettlementPort, NameTR: "Liman"}},
			},
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
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.forceDisembarkFleet("fleet_p1_1", "land_a")

	fleet := gs.Armies["fleet_p1_1"]
	if len(fleet.EmbarkedUnits) != 0 {
		t.Fatalf("zorunlu indirme sonrası cargo boş olmalı")
	}
	if fleet.DockedRegionID != "" || fleet.DockedSettlementID != "" {
		t.Fatalf("zorunlu indirme sonrası filo otomatik dock olmamalı, got docked_region=%q docked_settlement=%q", fleet.DockedRegionID, fleet.DockedSettlementID)
	}
	if fleet.RegionID != "sea_1" {
		t.Fatalf("zorunlu indirme sonrası filo denizde kalmalı, got=%s", fleet.RegionID)
	}
	if fleet.MovePoints != 2 {
		t.Fatalf("zorunlu indirme 1 hareket puanı tüketmeli, got=%d", fleet.MovePoints)
	}
	newArmy, ok := gs.Armies["army_p1_31"]
	if !ok {
		t.Fatalf("zorunlu indirme sonrası yeni kara ordusu beklenirdi")
	}
	if newArmy.RegionID != "land_a" || newArmy.IsNaval || len(newArmy.Units) != 1 {
		t.Fatalf("zorunlu indirme sonucu ordu hatalı: %+v", newArmy)
	}
}

func TestMoveArmyDisembarkFromDockedOwnedPort(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     20,
		Regions: map[world.RegionID]*world.Region{
			"sea_1": {
				ID:        "sea_1",
				IsSea:     true,
				Neighbors: []world.RegionID{"land_a"},
			},
			"land_a": {
				ID:          "land_a",
				OwnerID:     "p1",
				Neighbors:   []world.RegionID{"sea_1"},
				Buildings:   []string{"port"},
				Settlements: []world.Settlement{{ID: "port_a", Type: world.SettlementPort, NameTR: "Liman"}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_p1_1": {
				ID:                 "fleet_p1_1",
				OwnerID:            "p1",
				RegionID:           "sea_1",
				DockedRegionID:     "land_a",
				DockedSettlementID: "port_a",
				Units:              []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits:      []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
				MovePoints:         3,
				MaxMovePoints:      3,
				IsNaval:            true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry":  {ID: "infantry", Embarkable: true},
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "land_a")

	fleet := gs.Armies["fleet_p1_1"]
	if len(fleet.EmbarkedUnits) != 0 {
		t.Fatalf("dock edilmiş limandan indirme sonrası cargo boş olmalı")
	}
	if fleet.DockedRegionID != "land_a" || fleet.DockedSettlementID != "port_a" {
		t.Fatalf("indirme sonrası filo limana bağlı kalmalı, got docked_region=%q docked_settlement=%q", fleet.DockedRegionID, fleet.DockedSettlementID)
	}
	if fleet.MovePoints != 2 {
		t.Fatalf("indirme hareketi 1 puan tüketmeli, got=%d", fleet.MovePoints)
	}
	newArmy, ok := gs.Armies["army_p1_21"]
	if !ok {
		t.Fatalf("limandan indirme sonrası yeni kara ordusu beklenirdi")
	}
	if newArmy.RegionID != "land_a" || newArmy.IsNaval || len(newArmy.Units) != 1 {
		t.Fatalf("indirme sonucu ordu hatalı: %+v", newArmy)
	}
}

func TestMoveArmyDockFleetAtAlliedPort(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"sea_1": {
				ID:        "sea_1",
				IsSea:     true,
				Neighbors: []world.RegionID{"ally_land"},
			},
			"ally_land": {
				ID:          "ally_land",
				OwnerID:     "ally",
				Neighbors:   []world.RegionID{"sea_1"},
				Buildings:   []string{"port"},
				Settlements: []world.Settlement{{ID: "ally_port", Type: world.SettlementPort, NameTR: "Müttefik Limanı"}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
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
			"p1":   {ID: "p1"},
			"ally": {ID: "ally"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ally", "p1"): {FactionA: "ally", FactionB: "p1", Stance: faction.StanceAllied},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "ally_land")

	fleet := gs.Armies["fleet_p1_1"]
	if fleet.DockedRegionID != "ally_land" || fleet.DockedSettlementID != "ally_port" {
		t.Fatalf("müttefik limana dock olmalı, got docked_region=%q docked_settlement=%q", fleet.DockedRegionID, fleet.DockedSettlementID)
	}
	if fleet.MovePoints != 2 {
		t.Fatalf("dock hareketi 1 puan tüketmeli, got=%d", fleet.MovePoints)
	}
}

func TestMoveArmyDockFleetAtVassalPort(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "lord",
		Regions: map[world.RegionID]*world.Region{
			"sea_1": {
				ID:        "sea_1",
				IsSea:     true,
				Neighbors: []world.RegionID{"vassal_land"},
			},
			"vassal_land": {
				ID:          "vassal_land",
				OwnerID:     "vassal",
				Neighbors:   []world.RegionID{"sea_1"},
				Buildings:   []string{"port"},
				Settlements: []world.Settlement{{ID: "vassal_port", Type: world.SettlementPort, NameTR: "Vassal Limanı"}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_lord_1": {
				ID:            "fleet_lord_1",
				OwnerID:       "lord",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"lord":   {ID: "lord"},
			"vassal": {ID: "vassal", OverlordID: "lord"},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_lord_1", "vassal_land")

	fleet := gs.Armies["fleet_lord_1"]
	if fleet.DockedRegionID != "vassal_land" || fleet.DockedSettlementID != "vassal_port" {
		t.Fatalf("vassal limanına dock olmalı, got docked_region=%q docked_settlement=%q", fleet.DockedRegionID, fleet.DockedSettlementID)
	}
}

func TestMoveArmyDisembarkToVassalCoastWithoutWar(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "lord",
		NextArmySeq:     30,
		Regions: map[world.RegionID]*world.Region{
			"sea_1": {
				ID:        "sea_1",
				IsSea:     true,
				Neighbors: []world.RegionID{"vassal_land"},
			},
			"vassal_land": {
				ID:        "vassal_land",
				OwnerID:   "vassal",
				Neighbors: []world.RegionID{"sea_1"},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_lord_1": {
				ID:            "fleet_lord_1",
				OwnerID:       "lord",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"lord":   {ID: "lord", Religion: "sunni"},
			"vassal": {ID: "vassal", Religion: "sunni", OverlordID: "lord"},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry":  {ID: "infantry", Embarkable: true},
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_lord_1", "vassal_land")

	if _, ok := gs.Armies["army_lord_31"]; !ok {
		t.Fatalf("vassal kıyısına çıkarma sonrası kara ordusu oluşmalıydı")
	}
	if gs.Regions["vassal_land"].OwnerID != "vassal" {
		t.Fatalf("askeri geçişte vassal kıyısının sahibi değişmemeliydi, got=%s", gs.Regions["vassal_land"].OwnerID)
	}
	if len(gs.Armies["fleet_lord_1"].EmbarkedUnits) != 0 {
		t.Fatalf("çıkarma sonrası filo cargo'su boş olmalı")
	}
}

func TestMoveArmyDoesNotDockFleetAtPortSettlementWithoutPortBuilding(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"sea_1": {
				ID:        "sea_1",
				IsSea:     true,
				Neighbors: []world.RegionID{"land_a"},
			},
			"land_a": {
				ID:          "land_a",
				OwnerID:     "p1",
				Neighbors:   []world.RegionID{"sea_1"},
				Settlements: []world.Settlement{{ID: "port_a", Type: world.SettlementPort, NameTR: "Liman"}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
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
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "land_a")

	fleet := gs.Armies["fleet_p1_1"]
	if fleet.DockedRegionID != "" || fleet.DockedSettlementID != "" {
		t.Fatalf("port binasi yoksa filo dock olmamaliydi, got docked_region=%q docked_settlement=%q", fleet.DockedRegionID, fleet.DockedSettlementID)
	}
	if fleet.RegionID != "sea_1" {
		t.Fatalf("port binasi yoksa filo denizde kalmaliydi, got=%s", fleet.RegionID)
	}
	if fleet.MovePoints != 3 {
		t.Fatalf("gecersiz dock denemesinde hareket puani dusmemeli, got=%d", fleet.MovePoints)
	}
}

func TestMoveArmyDockFleetAtOwnedPortBuildingWithoutPortSettlement(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"sea_1": {
				ID:        "sea_1",
				IsSea:     true,
				Neighbors: []world.RegionID{"land_a"},
			},
			"land_a": {
				ID:          "land_a",
				OwnerID:     "p1",
				Neighbors:   []world.RegionID{"sea_1"},
				Buildings:   []string{"port"},
				Settlements: []world.Settlement{{ID: "izmir", Type: world.SettlementTown, NameTR: "İzmir"}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
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
			"transport": testTransportType(),
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmy("fleet_p1_1", "land_a")

	fleet := gs.Armies["fleet_p1_1"]
	if fleet.DockedRegionID != "land_a" || fleet.DockedSettlementID != "izmir" {
		t.Fatalf("port binasi olan bolgeye dock olmali, got docked_region=%q docked_settlement=%q", fleet.DockedRegionID, fleet.DockedSettlementID)
	}
}

func TestCompleteBuildingPortCreatesPortSettlement(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"land_a": {
				ID:      "land_a",
				OwnerID: "p1",
				WorldX:  90,
				WorldY:  120,
				Shape: [][][2]float32{
					{
						{70, 100},
						{100, 100},
						{100, 140},
						{70, 140},
					},
				},
				Neighbors:   []world.RegionID{"sea_1"},
				Settlements: []world.Settlement{{ID: "town_a", NameTR: "Kasaba", X: 92, Y: 118, Type: world.SettlementTown, IsCapital: true}},
			},
			"sea_1": {ID: "sea_1", IsSea: true, WorldX: 120, WorldY: 120},
		},
		BuildingTypes: map[string]*city.Building{
			"port": {ID: "port", MaxPerRegion: 1},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if !g.completeBuilding(gs.Regions["land_a"], "port") {
		t.Fatalf("port binasi tamamlanmaliydi")
	}
	region := gs.Regions["land_a"]
	if got := len(region.Buildings); got != 1 {
		t.Fatalf("port binasi kayda gecmeliydi, got=%d", got)
	}
	if got := len(region.Settlements); got != 2 {
		t.Fatalf("port binasi port settlement olusturmaliydi, got=%d", got)
	}
	portSettlement := region.Settlements[1]
	if portSettlement.Type != world.SettlementPort || portSettlement.NameTR != "Liman" {
		t.Fatalf("olusan settlement liman olmaliydi, got=%+v", portSettlement)
	}
	if portSettlement.ID != "land_a_port" {
		t.Fatalf("yeni port settlement id beklenmiyordu, got=%s", portSettlement.ID)
	}
	if portSettlement.X != 98 || portSettlement.Y != 120 {
		t.Fatalf("port settlement deniz sinirina yakin olmalıydı, got=(%d,%d)", portSettlement.X, portSettlement.Y)
	}
}
