package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiPathfindingTestState() *state.GameState {
	return &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Regions:    make(map[world.RegionID]*world.Region),
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":    {ID: "ai"},
			"enemy": {ID: "enemy"},
			"ally":  {ID: "ally"},
			"peace": {ID: "peace"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "enemy"): {FactionA: "ai", FactionB: "enemy", Stance: faction.StanceWar},
			faction.RelationKey("ai", "ally"):  {FactionA: "ai", FactionB: "ally", Stance: faction.StanceAllied},
			faction.RelationKey("ai", "peace"): {FactionA: "ai", FactionB: "peace", Stance: faction.StancePeace},
		},
		Armies: map[army.ArmyID]*army.Army{
			"mover": {
				ID: "mover", OwnerID: "ai", RegionID: "start",
				Units: []army.Unit{{TypeID: "inf", CurrentHP: army.MaxUnitHP}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50, GrainUpkeep: 2},
		},
	}
}

func TestWeightedLongRangeMovePrefersLowerTerrainCost(t *testing.T) {
	gs := aiPathfindingTestState()
	gs.Regions = map[world.RegionID]*world.Region{
		"start":   {ID: "start", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"pass", "plain_a"}},
		"pass":    {ID: "pass", OwnerID: "ai", Terrain: world.TerrainPass, Neighbors: []world.RegionID{"start", "goal"}},
		"plain_a": {ID: "plain_a", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"start", "plain_b"}},
		"plain_b": {ID: "plain_b", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"plain_a", "goal"}},
		"goal":    {ID: "goal", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"pass", "plain_b"}},
	}

	if got := chooseBestMove(gs, gs.Armies["mover"]); got != "plain_a" {
		t.Fatalf("ağırlıklı rota kısa geçit yerine düşük maliyetli ova hattını seçmeliydi: %s", got)
	}
}

func TestWeightedRouteAvoidsLocallyOutmatchedTransit(t *testing.T) {
	gs := aiPathfindingTestState()
	gs.Regions = map[world.RegionID]*world.Region{
		"start":      {ID: "start", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"danger", "safe_a"}},
		"danger":     {ID: "danger", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"start", "goal", "enemy_base"}},
		"safe_a":     {ID: "safe_a", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"start", "safe_b"}},
		"safe_b":     {ID: "safe_b", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"safe_a", "goal"}},
		"goal":       {ID: "goal", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"danger", "safe_b"}},
		"enemy_base": {ID: "enemy_base", OwnerID: "enemy", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"danger"}},
	}
	gs.Armies["enemy"] = &army.Army{
		ID: "enemy", OwnerID: "enemy", RegionID: "enemy_base",
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: army.MaxUnitHP},
			{TypeID: "inf", CurrentHP: army.MaxUnitHP},
			{TypeID: "inf", CurrentHP: army.MaxUnitHP},
		},
	}

	routes := aiWeightedLandRoutes(gs, gs.Armies["mover"], "start", aiRouteGeneral, 0, nil)
	if got := routes.nextStep("goal"); got != "safe_a" {
		t.Fatalf("rota kısa fakat ezici düşman tehdidindeki hattı kullanmamalıydı: %s", got)
	}
}

func TestWeightedRouteAvoidsProjectedLogisticsOverload(t *testing.T) {
	gs := aiPathfindingTestState()
	gs.Regions = map[world.RegionID]*world.Region{
		"start":   {ID: "start", OwnerID: "ai", Terrain: world.TerrainPlain, BaseGrainOutput: 10, Neighbors: []world.RegionID{"crowded", "safe_a"}},
		"crowded": {ID: "crowded", OwnerID: "ai", Terrain: world.TerrainPlain, BaseGrainOutput: 1, Neighbors: []world.RegionID{"start", "goal"}},
		"safe_a":  {ID: "safe_a", OwnerID: "ai", Terrain: world.TerrainPlain, BaseGrainOutput: 10, Neighbors: []world.RegionID{"start", "safe_b"}},
		"safe_b":  {ID: "safe_b", OwnerID: "ai", Terrain: world.TerrainPlain, BaseGrainOutput: 10, Neighbors: []world.RegionID{"safe_a", "goal"}},
		"goal":    {ID: "goal", OwnerID: "ai", Terrain: world.TerrainPlain, BaseGrainOutput: 10, Neighbors: []world.RegionID{"crowded", "safe_b"}},
	}
	gs.Armies["crowded_stack"] = &army.Army{
		ID: "crowded_stack", OwnerID: "ai", RegionID: "crowded",
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: army.MaxUnitHP},
			{TypeID: "inf", CurrentHP: army.MaxUnitHP},
			{TypeID: "inf", CurrentHP: army.MaxUnitHP},
			{TypeID: "inf", CurrentHP: army.MaxUnitHP},
		},
	}

	routes := aiWeightedLandRoutes(gs, gs.Armies["mover"], "start", aiRouteGeneral, 0, nil)
	if got := routes.nextStep("goal"); got != "safe_a" {
		t.Fatalf("rota kısa fakat ikmal aşımı yaratacak hattı kullanmamalıydı: %s", got)
	}
}

func TestWeightedRouteUsesAlliedTransitButBlocksPeaceTransit(t *testing.T) {
	gs := aiPathfindingTestState()
	gs.Regions = map[world.RegionID]*world.Region{
		"start": {ID: "start", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"peace_gate", "ally_gate"}},
		"peace_gate": {
			ID: "peace_gate", OwnerID: "peace", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"start", "goal"},
		},
		"ally_gate": {
			ID: "ally_gate", OwnerID: "ally", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"start", "ally_rear"},
		},
		"ally_rear": {
			ID: "ally_rear", OwnerID: "ally", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"ally_gate", "goal"},
		},
		"goal": {ID: "goal", OwnerID: "enemy", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"peace_gate", "ally_rear"}},
	}

	routes := aiWeightedLandRoutes(gs, gs.Armies["mover"], "start", aiRouteGeneral, 0, nil)
	if got := routes.nextStep("goal"); got != "ally_gate" {
		t.Fatalf("barıştaki üçüncü taraf bloke, müttefik transit açık olmalıydı: %s", got)
	}
}

func TestWeightedRouteTieBreaksByRegionID(t *testing.T) {
	gs := aiPathfindingTestState()
	gs.Regions = map[world.RegionID]*world.Region{
		"start": {ID: "start", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"b", "a"}},
		"a":     {ID: "a", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"start", "goal"}},
		"b":     {ID: "b", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"start", "goal"}},
		"goal":  {ID: "goal", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"b", "a"}},
	}

	routes := aiWeightedLandRoutes(gs, gs.Armies["mover"], "start", aiRouteGeneral, 0, nil)
	if got := routes.nextStep("goal"); got != "a" {
		t.Fatalf("eşit maliyetli rotada deterministik küçük region ID seçilmeliydi: %s", got)
	}
}

func TestFriendlyRouteDoesNotUseAlliedTerritory(t *testing.T) {
	gs := aiPathfindingTestState()
	gs.Regions = map[world.RegionID]*world.Region{
		"start": {ID: "start", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"ally_gate"}},
		"ally_gate": {
			ID: "ally_gate", OwnerID: "ally", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"start", "goal"},
		},
		"goal": {ID: "goal", OwnerID: "ai", Terrain: world.TerrainPlain, Neighbors: []world.RegionID{"ally_gate"}},
	}

	routes := aiWeightedLandRoutes(gs, gs.Armies["mover"], "start", aiRouteFriendly, 0, nil)
	if _, reachable := routes.distance("goal"); reachable {
		t.Fatal("retreat/security dost-toprak rotası müttefik bölgesini transit kullanmamalıydı")
	}
}
