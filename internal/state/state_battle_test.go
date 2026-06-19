package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestSelectBattleDefenderPrefersStrongestDeterministically(t *testing.T) {
	gs := &GameState{
		Armies: map[army.ArmyID]*army.Army{
			"atk":      {ID: "atk", OwnerID: "p1", RegionID: "src"},
			"weak":     {ID: "weak", OwnerID: "p2", RegionID: "dst", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"strong":   {ID: "strong", OwnerID: "p3", RegionID: "dst", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
			"strong_b": {ID: "strong_b", OwnerID: "p4", RegionID: "dst", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 40},
		},
	}

	defender := gs.SelectBattleDefender(gs.Armies["atk"], world.RegionID("dst"), false)
	if defender == nil || defender.ID != "strong" {
		t.Fatalf("en güçlü ve tie-break'te en küçük ID bekleniyordu, got=%v", defender)
	}
}

func TestSelectBattleDefenderFiltersNavalTargetsByWarState(t *testing.T) {
	gs := &GameState{
		Armies: map[army.ArmyID]*army.Army{
			"atk":  {ID: "atk", OwnerID: "p1", RegionID: "sea_a", IsNaval: true},
			"ally": {ID: "ally", OwnerID: "p2", RegionID: "sea_b", IsNaval: true, Units: []army.Unit{{TypeID: "ship", CurrentHP: 100}}},
			"war":  {ID: "war", OwnerID: "p3", RegionID: "sea_b", IsNaval: true, Units: []army.Unit{{TypeID: "ship", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"ship": {ID: "ship", Attack: 10, Defense: 10, Morale: 40},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p1", "p2"): {FactionA: "p1", FactionB: "p2", Stance: faction.StancePeace},
			faction.RelationKey("p1", "p3"): {FactionA: "p1", FactionB: "p3", Stance: faction.StanceWar},
		},
	}

	defender := gs.SelectBattleDefender(gs.Armies["atk"], world.RegionID("sea_b"), true)
	if defender == nil || defender.ID != "war" {
		t.Fatalf("denizde sadece savaş halindeki filo hedef olmalı, got=%v", defender)
	}
}
