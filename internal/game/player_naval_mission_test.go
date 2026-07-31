package game

import (
	"testing"

	"mapp-game-go/internal/army"
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

func TestExecutePlayerNavalMissionsMovesFleetTowardAssignedSea(t *testing.T) {
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
	if gs.Armies["fleet"].RegionID != "target" {
		t.Fatalf("görevli oyuncu filosu hedefe ilerlemedi: %s", gs.Armies["fleet"].RegionID)
	}
	if !gs.ArmyMoveUsage["fleet"] {
		t.Fatal("otomatik görev hareketi ekonomi hareket kullanımına işlenmedi")
	}
}
