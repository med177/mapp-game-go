package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAINavalFishingAssignsIdleFleetOnOwnedCoastWhenGrainIsNeeded(t *testing.T) {
	const fid faction.FactionID = "ai"
	fleet := &army.Army{
		ID: "fleet", OwnerID: string(fid), IsNaval: true, RegionID: "sea",
		Units: []army.Unit{{TypeID: "fishing_ship", CurrentHP: army.MaxUnitHP}},
	}
	gs := &state.GameState{
		PlayerFactionID: fid,
		Factions: map[faction.FactionID]*faction.Faction{
			fid: {ID: fid, Grain: 0},
		},
		UnitTypes: map[string]*army.UnitType{
			"fishing_ship": {ID: "fishing_ship", Category: army.CategoryNavalTrade, FishingGrainIncome: 40, GrainUpkeep: 20},
		},
		Regions: map[world.RegionID]*world.Region{
			"coast": {ID: "coast", OwnerID: string(fid), Neighbors: []world.RegionID{"sea"}},
			"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
		},
		Armies: map[army.ArmyID]*army.Army{"fleet": fleet},
	}
	ctx := buildStrategicContext(gs, fid)

	target, handled := aiNavalMissionMove(gs, fleet, ctx)
	if !handled || target != "" {
		t.Fatalf("fishing mission move = %q, handled %v; want assigned in place", target, handled)
	}
	if fleet.NavalMission == nil || fleet.NavalMission.Kind != army.NavalMissionFishing || fleet.NavalMission.TargetRegionID != "sea" {
		t.Fatalf("idle fleet fishing mission = %#v, want fishing at own coastal sea", fleet.NavalMission)
	}
	if got, want := gs.NavalFishingGrainIncomeForFaction(string(fid)), 40; got != want {
		t.Fatalf("AI faction fishing income = %d, want %d", got, want)
	}
}

func TestAINavalFishingMovesDockedIdleFleetTowardCoastalSea(t *testing.T) {
	const fid faction.FactionID = "ai"
	fleet := &army.Army{
		ID: "fleet", OwnerID: string(fid), IsNaval: true, RegionID: "sea", DockedRegionID: "port",
		Units: []army.Unit{{TypeID: "fishing_ship", CurrentHP: army.MaxUnitHP}},
	}
	gs := &state.GameState{
		PlayerFactionID: fid,
		Factions:        map[faction.FactionID]*faction.Faction{fid: {ID: fid, Grain: 0}},
		UnitTypes: map[string]*army.UnitType{
			"fishing_ship": {ID: "fishing_ship", Category: army.CategoryNavalTrade, FishingGrainIncome: 40, GrainUpkeep: 20},
		},
		Regions: map[world.RegionID]*world.Region{
			"port":  {ID: "port", OwnerID: string(fid), Neighbors: []world.RegionID{"sea"}},
			"coast": {ID: "coast", OwnerID: string(fid), Neighbors: []world.RegionID{"sea"}},
			"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"port", "coast"}},
		},
		Armies: map[army.ArmyID]*army.Army{"fleet": fleet},
	}

	target, handled := aiNavalMissionMove(gs, fleet, buildStrategicContext(gs, fid))
	if !handled || target != "sea" {
		t.Fatalf("docked fleet fishing move = %q, handled %v; want to leave port for sea", target, handled)
	}
	if fleet.NavalMission != nil {
		t.Fatalf("docked fleet should receive fishing mission only after reaching open sea, got %#v", fleet.NavalMission)
	}
}

func TestAINavalFishingMovesFleetToSafeOwnedCoast(t *testing.T) {
	const fid faction.FactionID = "ai"
	fleet := &army.Army{
		ID: "fleet", OwnerID: string(fid), IsNaval: true, RegionID: "far_sea",
		Units: []army.Unit{{TypeID: "fishing_ship", CurrentHP: army.MaxUnitHP}},
	}
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{fid: {ID: fid, Grain: 0}},
		UnitTypes: map[string]*army.UnitType{
			"fishing_ship": {ID: "fishing_ship", Category: army.CategoryNavalTrade, FishingGrainIncome: 40, GrainUpkeep: 20},
		},
		Regions: map[world.RegionID]*world.Region{
			"coast":       {ID: "coast", OwnerID: string(fid), Neighbors: []world.RegionID{"fishing_sea"}},
			"fishing_sea": {ID: "fishing_sea", IsSea: true, Neighbors: []world.RegionID{"coast", "far_sea"}},
			"far_sea":     {ID: "far_sea", IsSea: true, Neighbors: []world.RegionID{"fishing_sea"}},
		},
		Armies: map[army.ArmyID]*army.Army{"fleet": fleet},
	}
	ctx := buildStrategicContext(gs, fid)

	target, handled := aiNavalMissionMove(gs, fleet, ctx)
	if !handled || target != "fishing_sea" {
		t.Fatalf("fleet fishing route = %q, handled %v; want safe coastal sea", target, handled)
	}
	if fleet.NavalMission != nil {
		t.Fatalf("fleet should receive fishing mission after arriving, got %#v", fleet.NavalMission)
	}

	fleet.RegionID = "fishing_sea"
	target, handled = aiNavalMissionMove(gs, fleet, ctx)
	if !handled || target != "" || fleet.NavalMission == nil || fleet.NavalMission.Kind != army.NavalMissionFishing {
		t.Fatalf("arrival fishing assignment = target %q, handled %v, mission %#v", target, handled, fleet.NavalMission)
	}
}

func TestAINavalFishingPreservesStrategicAndMerchantMissions(t *testing.T) {
	const fid faction.FactionID = "ai"
	for _, mission := range []army.NavalMission{
		{Kind: army.NavalMissionPatrol, TargetRegionID: "sea"},
		{Kind: army.NavalMissionBlockade, TargetRegionID: "sea"},
	} {
		fleet := &army.Army{
			ID: "fleet", OwnerID: string(fid), IsNaval: true, RegionID: "sea",
			TradeRouteKey: "route", NavalMission: &mission,
			Units: []army.Unit{{TypeID: "fishing_ship", CurrentHP: army.MaxUnitHP}},
		}
		gs := &state.GameState{
			PlayerFactionID: fid,
			Factions:        map[faction.FactionID]*faction.Faction{fid: {ID: fid, Grain: 0}},
			UnitTypes: map[string]*army.UnitType{
				"fishing_ship": {ID: "fishing_ship", Category: army.CategoryNavalTrade, FishingGrainIncome: 40, GrainUpkeep: 20},
			},
			Regions: map[world.RegionID]*world.Region{
				"coast": {ID: "coast", OwnerID: string(fid), Neighbors: []world.RegionID{"sea"}},
				"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
			},
			Armies: map[army.ArmyID]*army.Army{"fleet": fleet},
		}
		target, handled := aiNavalFishingMissionMove(gs, fleet, buildStrategicContext(gs, fid))
		if handled || target != "" || fleet.NavalMission.Kind != mission.Kind {
			t.Fatalf("existing %s mission changed: target=%q handled=%v mission=%#v", mission.Kind, target, handled, fleet.NavalMission)
		}
	}
}

func TestAINavalFishingReleasesFleetWhenGrainStorageIsFull(t *testing.T) {
	const fid faction.FactionID = "ai"
	fleet := &army.Army{
		ID: "fleet", OwnerID: string(fid), IsNaval: true, RegionID: "sea",
		NavalMission: &army.NavalMission{Kind: army.NavalMissionFishing, TargetRegionID: "sea"},
		Units:        []army.Unit{{TypeID: "fishing_ship", CurrentHP: army.MaxUnitHP}},
	}
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{fid: {ID: fid, Grain: 100}},
		UnitTypes: map[string]*army.UnitType{
			"fishing_ship": {ID: "fishing_ship", Category: army.CategoryNavalTrade, FishingGrainIncome: 40, GrainUpkeep: 20},
		},
		Regions: map[world.RegionID]*world.Region{
			"coast": {ID: "coast", OwnerID: string(fid), Neighbors: []world.RegionID{"sea"}},
			"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
		},
		Armies: map[army.ArmyID]*army.Army{"fleet": fleet},
	}

	_, handled := aiNavalFishingMissionMove(gs, fleet, buildStrategicContext(gs, fid))
	if handled || fleet.NavalMission != nil {
		t.Fatalf("fleet should leave fishing duty at full storage: handled=%v mission=%#v", handled, fleet.NavalMission)
	}
}
