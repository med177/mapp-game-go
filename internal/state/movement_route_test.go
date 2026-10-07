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

func TestAdvanceTurnClearsArmyMoveUsageSnapshot(t *testing.T) {
	gs := &GameState{
		Turn:  4,
		Year:  1300,
		Month: 1,
		Armies: map[army.ArmyID]*army.Army{
			"army": {ID: "army", OwnerID: "player", MovePoints: 1, MaxMovePoints: 3},
		},
		ArmyMoveUsage: map[army.ArmyID]bool{"army": true},
	}

	gs.AdvanceTurn()

	if gs.ArmyMoveUsage != nil {
		t.Fatalf("ArmyMoveUsage tur geçişinden sonra temizlenmedi: %#v", gs.ArmyMoveUsage)
	}
	if got := gs.Armies["army"].MovePoints; got != gs.Armies["army"].MaxMovePoints {
		t.Fatalf("yeni tur hareket puanı = %d, max = %d olmalı", got, gs.Armies["army"].MaxMovePoints)
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

func TestAirMovementRouteCrossesSeaAndBlockedTerrain(t *testing.T) {
	const (
		start  = world.RegionID("start")
		sea    = world.RegionID("sea")
		target = world.RegionID("blocked_terrain")
	)
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			start:  {ID: start, OwnerID: "player", Neighbors: []world.RegionID{sea}},
			sea:    {ID: sea, IsSea: true, Neighbors: []world.RegionID{start, target}},
			target: {ID: target, IsTerrainArea: true, IsLocked: true, TerrainAreaID: "blocked", Neighbors: []world.RegionID{sea}},
		},
		UnitTypes: map[string]*army.UnitType{
			"dragon": {ID: "dragon", MovementType: army.MovementTypeAir},
		},
	}
	a := &army.Army{
		ID:         "dragon_army",
		OwnerID:    "player",
		RegionID:   start,
		MovePoints: 4,
		Units:      []army.Unit{{TypeID: "dragon"}},
	}

	route := gs.MovementRouteForArmy(a, target)
	if len(route) != 3 || route[1] != sea || route[2] != target {
		t.Fatalf("air route = %v, want start,sea,blocked_terrain", route)
	}
	if got := gs.MovementReachableForArmy(a).Nodes[target].Cost; got != 2 {
		t.Fatalf("air terrain cost = %d, want 2", got)
	}
}

func TestAirMovementRouteUsesScenarioAirspaceDiplomacyRule(t *testing.T) {
	regions := map[world.RegionID]*world.Region{
		"start":   {ID: "start", OwnerID: "player", Neighbors: []world.RegionID{"foreign"}},
		"foreign": {ID: "foreign", OwnerID: "enemy", Neighbors: []world.RegionID{"start", "beyond"}},
		"beyond":  {ID: "beyond", OwnerID: "enemy", Neighbors: []world.RegionID{"foreign"}},
	}
	a := &army.Army{
		ID:         "dragon_army",
		OwnerID:    "player",
		RegionID:   "start",
		MovePoints: 4,
		Units:      []army.Unit{{TypeID: "dragon"}},
	}
	base := GameState{
		Regions:   regions,
		UnitTypes: map[string]*army.UnitType{"dragon": {ID: "dragon", MovementType: army.MovementTypeAir}},
		Armies:    map[army.ArmyID]*army.Army{a.ID: a},
	}

	if route := base.MovementRouteForArmy(a, "beyond"); len(route) != 3 {
		t.Fatalf("unrestricted airspace route = %v, want route through foreign region", route)
	}
	base.AirspaceEnabled = true
	if route := base.MovementRouteForArmy(a, "beyond"); route != nil {
		t.Fatalf("restricted airspace route = %v, want blocked transit", route)
	}
}

func TestAirSortieUsesHalfRangeAndReturnsToOrigin(t *testing.T) {
	start := world.RegionID("start")
	sea := world.RegionID("sea")
	target := world.RegionID("target")
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			start:    {ID: start, OwnerID: "player", Neighbors: []world.RegionID{sea}},
			sea:      {ID: sea, IsSea: true, Neighbors: []world.RegionID{start, "middle"}},
			"middle": {ID: "middle", IsSea: true, Neighbors: []world.RegionID{sea, target}},
			target:   {ID: target, IsSea: true, Neighbors: []world.RegionID{"middle"}},
		},
		UnitTypes: map[string]*army.UnitType{
			"dragon": {ID: "dragon", Category: army.CategoryDragon},
		},
	}
	a := &army.Army{
		ID:         "dragon_army",
		OwnerID:    "player",
		RegionID:   start,
		MovePoints: 5,
		Units:      []army.Unit{{TypeID: "dragon"}},
	}
	reachability := gs.MovementReachableForArmy(a)
	if gs.AirSortieAllowed(a, target, reachability) {
		t.Fatalf("three-step sortie should exceed half-range")
	}
	if gs.AirSortieBudget(a) != 2 {
		t.Fatalf("sortie budget = %d, want 2", gs.AirSortieBudget(a))
	}

	shortTarget := sea
	if !gs.AirSortieAllowed(a, shortTarget, reachability) {
		t.Fatalf("one-step sortie should be allowed")
	}
	if !gs.BeginAirSortie(a, shortTarget) {
		t.Fatal("expected sortie to start")
	}
	a.RegionID = shortTarget
	a.MovePoints -= 1
	gs.RecordAirSortieStep(a, 1)
	if !gs.FinishAirSortie(a, shortTarget) || a.RegionID != start || a.MovePoints != 3 {
		t.Fatalf("sortie state = region %q, points %d; want start, 3", a.RegionID, a.MovePoints)
	}
}
