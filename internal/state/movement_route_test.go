package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestMovementRouteUsesMoveBudgetAndDeterministicTieBreak(t *testing.T) {
	regions := map[world.RegionID]*world.Region{
		"start":  {ID: "start", OwnerID: "player", Neighbors: []world.RegionID{"via_c", "via_b"}},
		"via_b":  {ID: "via_b", OwnerID: "player", Neighbors: []world.RegionID{"start", "target"}},
		"via_c":  {ID: "via_c", OwnerID: "player", Neighbors: []world.RegionID{"start", "target"}},
		"target": {ID: "target", OwnerID: "player", Neighbors: []world.RegionID{"via_b", "via_c"}},
	}
	gs := &GameState{
		Regions: regions,
		Armies: map[army.ArmyID]*army.Army{
			"army": {ID: "army", OwnerID: "player", RegionID: "start", MovePoints: 2},
		},
	}

	got := gs.MovementRouteForArmy(gs.Armies["army"], "target")
	want := []world.RegionID{"start", "via_b", "target"}
	if len(got) != len(want) {
		t.Fatalf("route length = %v, want %v (%v)", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("route[%d] = %q, want %q (%v)", index, got[index], want[index], got)
		}
	}
}

func TestMovementRouteStopsTransitAtForeignRegion(t *testing.T) {
	regions := map[world.RegionID]*world.Region{
		"start":    {ID: "start", OwnerID: "player", Neighbors: []world.RegionID{"friendly"}},
		"friendly": {ID: "friendly", OwnerID: "player", Neighbors: []world.RegionID{"start", "enemy"}},
		"enemy":    {ID: "enemy", OwnerID: "enemy", Neighbors: []world.RegionID{"friendly", "beyond"}},
		"beyond":   {ID: "beyond", OwnerID: "enemy", Neighbors: []world.RegionID{"enemy"}},
	}
	gs := &GameState{
		Regions: regions,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy"},
		},
		Relations: map[string]*faction.Relation{},
	}
	a := &army.Army{ID: "army", OwnerID: "player", RegionID: "start", MovePoints: 3}
	gs.Armies = map[army.ArmyID]*army.Army{a.ID: a}

	if got := gs.MovementRouteForArmy(a, "enemy"); len(got) != 3 || got[2] != "enemy" {
		t.Fatalf("foreign terminal route = %v, want start,friendly,enemy", got)
	}
	if got := gs.MovementRouteForArmy(a, "beyond"); got != nil {
		t.Fatalf("route continued through foreign terminal: %v", got)
	}
}

func TestNavalMovementRouteReachesSeaAndPortWithinBudget(t *testing.T) {
	regions := map[world.RegionID]*world.Region{
		"sea_start": {ID: "sea_start", IsSea: true, Neighbors: []world.RegionID{"sea_mid"}},
		"sea_mid":   {ID: "sea_mid", IsSea: true, Neighbors: []world.RegionID{"sea_start", "sea_end", "port"}},
		"sea_end":   {ID: "sea_end", IsSea: true, Neighbors: []world.RegionID{"sea_mid"}},
		"port":      {ID: "port", OwnerID: "player", Buildings: []string{"port"}, Neighbors: []world.RegionID{"sea_mid"}},
	}
	gs := &GameState{
		Regions: regions,
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {ID: "fleet", OwnerID: "player", IsNaval: true, RegionID: "sea_start", MovePoints: 2},
		},
		Factions: map[faction.FactionID]*faction.Faction{"player": {ID: "player"}},
	}
	fleet := gs.Armies["fleet"]
	if got := gs.MovementRouteForArmy(fleet, "sea_end"); len(got) != 3 {
		t.Fatalf("sea route = %v, want three nodes", got)
	}
	if got := gs.MovementRouteForArmy(fleet, "port"); len(got) != 3 || got[2] != "port" {
		t.Fatalf("port route = %v, want sea_start,sea_mid,port", got)
	}
}

func TestNavalMovementRouteCanLeaveSeaWithEnemyFleet(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"sea_start": {ID: "sea_start", IsSea: true, Neighbors: []world.RegionID{"sea_mid"}},
			"sea_mid":   {ID: "sea_mid", IsSea: true, Neighbors: []world.RegionID{"sea_start", "sea_end"}},
			"sea_end":   {ID: "sea_end", IsSea: true, Neighbors: []world.RegionID{"sea_mid"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet":       {ID: "fleet", OwnerID: "player", IsNaval: true, RegionID: "sea_start", MovePoints: 2},
			"enemy_fleet": {ID: "enemy_fleet", OwnerID: "enemy", IsNaval: true, RegionID: "sea_mid", MovePoints: 2},
		},
	}

	fleet := gs.Armies["fleet"]
	reachability := gs.MovementReachableForArmy(fleet)
	if _, ok := reachability.Nodes["sea_end"]; !ok {
		t.Fatalf("expected fleet to pass a foreign fleet without war: %#v", reachability.Nodes)
	}
	if route := gs.MovementRouteForArmy(fleet, "sea_end"); len(route) != 3 {
		t.Fatalf("expected route through a foreign fleet without war, got %v", route)
	}
}
