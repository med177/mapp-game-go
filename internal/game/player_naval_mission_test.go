package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestPlayerSeaRouteNextUsesDeterministicShortestRoute(t *testing.T) {
	gs := &state.GameState{Regions: map[world.RegionID]*world.Region{
		"start":  {ID: "start", IsSea: true, Neighbors: []world.RegionID{"long", "short"}},
		"long":   {ID: "long", IsSea: true, Neighbors: []world.RegionID{"start", "target"}},
		"short":  {ID: "short", IsSea: true, Neighbors: []world.RegionID{"start", "target"}},
		"target": {ID: "target", IsSea: true, Neighbors: []world.RegionID{"long", "short"}},
	}}
	next, ok := playerSeaRouteNext(gs, "start", "target")
	if !ok || next != "long" {
		t.Fatalf("eşit uzunluktaki rota deterministik seçilmedi: next=%s ok=%v", next, ok)
	}
}

func TestPlayerTransportMissionTargetsCoastAfterSeaRoute(t *testing.T) {
	gs := &state.GameState{Regions: map[world.RegionID]*world.Region{
		"start":      {ID: "start", IsSea: true, Neighbors: []world.RegionID{"sea_target"}},
		"sea_target": {ID: "sea_target", IsSea: true, Neighbors: []world.RegionID{"start", "coast"}},
		"coast":      {ID: "coast", Neighbors: []world.RegionID{"sea_target"}},
	}, Armies: map[army.ArmyID]*army.Army{
		"fleet": {ID: "fleet", OwnerID: "player", IsNaval: true, RegionID: "start", NavalMission: &army.NavalMission{Kind: army.NavalMissionTransport, TargetRegionID: "coast"}},
	}}
	next, ok := playerNavalMissionNextTarget(gs, gs.Armies["fleet"])
	if !ok || next != "sea_target" {
		t.Fatalf("nakliye filosu kıyının deniz rotasına yönlenmedi: next=%s ok=%v", next, ok)
	}
	gs.Armies["fleet"].RegionID = "sea_target"
	next, ok = playerNavalMissionNextTarget(gs, gs.Armies["fleet"])
	if !ok || next != "coast" {
		t.Fatalf("nakliye filosu kıyıya iniş adımı üretmedi: next=%s ok=%v", next, ok)
	}
}

func TestExecutePlayerNavalMissionsKeepsPatrolInCurrentSea(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		UnitTypes: map[string]*army.UnitType{
			"warship": {ID: "warship", Category: army.CategoryNavalWar, MovementPoints: 1},
		},
		Regions: map[world.RegionID]*world.Region{
			"start":  {ID: "start", IsSea: true, Neighbors: []world.RegionID{"target"}},
			"target": {ID: "target", IsSea: true, Neighbors: []world.RegionID{"start"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID: "fleet", OwnerID: "player", IsNaval: true, RegionID: "start",
				Units: []army.Unit{{TypeID: "warship"}}, MovePoints: 1, MaxMovePoints: 1,
				NavalMission: &army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "target"},
			},
		},
	}
	(&Game{gs: gs}).executePlayerNavalMissions()
	if gs.Armies["fleet"].RegionID != "start" {
		t.Fatalf("devriye filosu görev atandığı mevcut denizden ayrılmamalı: %s", gs.Armies["fleet"].RegionID)
	}
	if gs.ArmyMoveUsage["fleet"] {
		t.Fatal("mevcut denizde sabit kalan devriye filosu otomatik hareket kullanımı yazmamalı")
	}
}

func TestMovingEscortedFleetMovesEscortWithAvailableMovement(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		UnitTypes: map[string]*army.UnitType{
			"warship":   {ID: "warship", Category: army.CategoryNavalWar},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"sea_a": {ID: "sea_a", IsSea: true, Neighbors: []world.RegionID{"sea_b"}},
			"sea_b": {ID: "sea_b", IsSea: true, Neighbors: []world.RegionID{"sea_a"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"transport": {
				ID: "transport", OwnerID: "player", RegionID: "sea_a", IsNaval: true,
				Units: []army.Unit{{TypeID: "transport"}}, MovePoints: 1, MaxMovePoints: 1,
			},
			"escort": {
				ID: "escort", OwnerID: "player", RegionID: "sea_a", IsNaval: true,
				Units: []army.Unit{{TypeID: "warship"}}, MovePoints: 1, MaxMovePoints: 1,
				NavalMission: &army.NavalMission{Kind: army.NavalMissionEscort, TargetFleetID: "transport"},
			},
		},
	}

	(&Game{gs: gs, renderer: &render.Renderer{}}).moveArmy("transport", "sea_b")

	if got := gs.Armies["transport"].RegionID; got != "sea_b" {
		t.Fatalf("korunan filo hedef denize gidemedi: %s", got)
	}
	if got := gs.Armies["escort"].RegionID; got != "sea_b" {
		t.Fatalf("escort filosu korunan filoyu takip etmedi: %s", got)
	}
	if gs.Armies["escort"].MovePoints != 0 {
		t.Fatalf("escort hareketi kendi hareket puanını harcamalı: %d", gs.Armies["escort"].MovePoints)
	}
}

func TestMovingEscortedFleetLeavesEscortWhenMovementIsInsufficient(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		UnitTypes: map[string]*army.UnitType{
			"warship":   {ID: "warship", Category: army.CategoryNavalWar},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"sea_a": {ID: "sea_a", IsSea: true, Neighbors: []world.RegionID{"sea_b"}},
			"sea_b": {ID: "sea_b", IsSea: true, Neighbors: []world.RegionID{"sea_a"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"transport": {
				ID: "transport", OwnerID: "player", RegionID: "sea_a", IsNaval: true,
				Units: []army.Unit{{TypeID: "transport"}}, MovePoints: 1, MaxMovePoints: 1,
			},
			"escort": {
				ID: "escort", OwnerID: "player", RegionID: "sea_a", IsNaval: true,
				Units: []army.Unit{{TypeID: "warship"}}, MovePoints: 0, MaxMovePoints: 1,
				NavalMission: &army.NavalMission{Kind: army.NavalMissionEscort, TargetFleetID: "transport"},
			},
		},
	}

	(&Game{gs: gs, renderer: &render.Renderer{}}).moveArmy("transport", "sea_b")

	if got := gs.Armies["transport"].RegionID; got != "sea_b" {
		t.Fatalf("korunan filo hareket edebilmeliydi: %s", got)
	}
	if got := gs.Armies["escort"].RegionID; got != "sea_a" {
		t.Fatalf("hareket puanı olmayan escort yerinden ayrılmamalı: %s", got)
	}
}
