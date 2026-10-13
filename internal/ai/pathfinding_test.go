package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIRouteNeighborIDsIncludesLandPassageEndpoints(t *testing.T) {
	from := &world.Region{ID: "from"}
	to := &world.Region{ID: "to"}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			from.ID: from,
			to.ID:   to,
		},
		LandPassages: []world.LandPassage{{From: from.ID, To: to.ID, MoveCost: 1}},
	}

	neighbors := aiRouteNeighborIDs(gs, from)
	if len(neighbors) != 1 || neighbors[0] != to.ID {
		t.Fatalf("land passage endpoint was not added to AI route neighbors: %#v", neighbors)
	}
}

func TestAILandEntryMoveCostIncludesTerrainAndPassageCost(t *testing.T) {
	from := &world.Region{ID: "from"}
	target := &world.Region{ID: "target"}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			from.ID:   from,
			target.ID: target,
		},
		LandPassages: []world.LandPassage{{From: from.ID, To: target.ID, MoveCost: 3}},
	}

	cost, allowed := aiLandEntryMoveCost(gs, from.ID, target)
	if !allowed || cost != 3 {
		t.Fatalf("unexpected passage movement cost: allowed=%v cost=%d", allowed, cost)
	}
}

func TestAITerrainMovePenaltyIncludesAttrition(t *testing.T) {
	from := &world.Region{ID: "from"}
	target := &world.Region{ID: "terrain", IsTerrainArea: true, TerrainAreaID: "area", IsLocked: false}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			from.ID:   from,
			target.ID: target,
		},
		TerrainAreas: []world.TerrainArea{{ID: "area", MoveCost: -1, AttritionCost: 10}},
	}

	penalty := aiTerrainMovePenalty(gs, from.ID, target)
	if penalty != 30 {
		t.Fatalf("unexpected terrain risk penalty: %d", penalty)
	}
}

func TestAIRouteUsesMoveCostInsteadOfTerrainLabelPassability(t *testing.T) {
	from := &world.Region{ID: "from", OwnerID: "f", Neighbors: []world.RegionID{"lake"}}
	target := &world.Region{ID: "lake", OwnerID: "f", Terrain: world.TerrainLake}
	armyRef := &army.Army{ID: "army", OwnerID: "f", RegionID: from.ID}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			from.ID:   from,
			target.ID: target,
		},
		Armies: map[army.ArmyID]*army.Army{armyRef.ID: armyRef},
	}

	routes := aiWeightedLandRoutes(gs, armyRef, from.ID, aiRouteFriendly, 0, nil)
	if _, reachable := routes.distance(target.ID); !reachable {
		t.Fatal("a non-zero move-cost terrain label should not block the AI route")
	}
}

func TestAIRouteRejectsBlockedTerrainArea(t *testing.T) {
	from := &world.Region{ID: "from", OwnerID: "f", Neighbors: []world.RegionID{"blocked"}}
	target := &world.Region{ID: "blocked", OwnerID: "", IsTerrainArea: true, TerrainAreaID: "area"}
	armyRef := &army.Army{ID: "army", OwnerID: "f", RegionID: from.ID}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			from.ID:   from,
			target.ID: target,
		},
		Armies:       map[army.ArmyID]*army.Army{armyRef.ID: armyRef},
		TerrainAreas: []world.TerrainArea{{ID: "area", MoveCost: 0}},
	}

	routes := aiWeightedLandRoutes(gs, armyRef, from.ID, aiRouteGeneral, 0, nil)
	if _, reachable := routes.distance(target.ID); reachable {
		t.Fatal("a zero move-cost terrain area must remain blocked for the AI route")
	}
}

func TestAITransitAgreementEnablesRouteThroughForeignLand(t *testing.T) {
	const (
		actor  = faction.FactionID("actor")
		bridge = faction.FactionID("bridge")
		enemy  = faction.FactionID("enemy")
	)
	start := &world.Region{ID: "start", OwnerID: string(actor), Neighbors: []world.RegionID{"bridge_land"}}
	bridgeLand := &world.Region{ID: "bridge_land", OwnerID: string(bridge), Neighbors: []world.RegionID{"start", "enemy_land"}}
	enemyLand := &world.Region{ID: "enemy_land", OwnerID: string(enemy), Neighbors: []world.RegionID{"bridge_land"}}
	armyRef := &army.Army{ID: "army", OwnerID: string(actor), RegionID: start.ID, MovePoints: 2}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			start.ID:      start,
			bridgeLand.ID: bridgeLand,
			enemyLand.ID:  enemyLand,
		},
		Armies: map[army.ArmyID]*army.Army{armyRef.ID: armyRef},
		Factions: map[faction.FactionID]*faction.Faction{
			actor:  {ID: actor},
			bridge: {ID: bridge},
			enemy:  {ID: enemy},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(actor, enemy): {
				FactionA: actor,
				FactionB: enemy,
				Stance:   faction.StanceWar,
			},
		},
	}

	if routes := aiWeightedLandRoutes(gs, armyRef, start.ID, aiRouteGeneral, 0, nil); routes.nextStep(enemyLand.ID) != "" {
		t.Fatal("AI route unexpectedly crossed foreign land without a transit agreement")
	}

	gs.AddTransitAgreement(actor, bridge)
	routes := aiWeightedLandRoutes(gs, armyRef, start.ID, aiRouteGeneral, 0, nil)
	if got := routes.nextStep(enemyLand.ID); got != bridgeLand.ID {
		t.Fatalf("AI route next step = %q, want transit territory %q", got, bridgeLand.ID)
	}
	if score := scoreMove(gs, armyRef, bridgeLand); score <= 0 {
		t.Fatalf("AI movement scorer rejected transit territory: %d", score)
	}

	gs.Relations[faction.RelationKey(actor, bridge)] = &faction.Relation{
		FactionA: actor,
		FactionB: bridge,
		Stance:   faction.StanceWar,
	}
	if routes = aiWeightedLandRoutes(gs, armyRef, start.ID, aiRouteGeneral, 0, nil); routes.nextStep(enemyLand.ID) != "" {
		t.Fatal("AI route continued through transit land after war began")
	}
}

func TestAIMovementThroughTransitAgreementDoesNotConquerRegion(t *testing.T) {
	const (
		actor  = faction.FactionID("actor")
		bridge = faction.FactionID("bridge")
	)
	start := &world.Region{ID: "start", OwnerID: string(actor), Neighbors: []world.RegionID{"bridge_land"}}
	bridgeLand := &world.Region{ID: "bridge_land", OwnerID: string(bridge), Neighbors: []world.RegionID{"start"}}
	armyRef := &army.Army{
		ID: "army", OwnerID: string(actor), RegionID: start.ID, MovePoints: 2,
		Units: []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}},
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			start.ID:      start,
			bridgeLand.ID: bridgeLand,
		},
		Armies: map[army.ArmyID]*army.Army{armyRef.ID: armyRef},
		Factions: map[faction.FactionID]*faction.Faction{
			actor:  {ID: actor},
			bridge: {ID: bridge},
		},
	}
	gs.AddTransitAgreement(actor, bridge)

	outcome := executeMove(gs, armyRef, bridgeLand.ID, actor)
	if !outcome.survived || armyRef.RegionID != bridgeLand.ID {
		t.Fatalf("AI ordusu transit bölgesine giremedi: outcome=%+v region=%q", outcome, armyRef.RegionID)
	}
	if bridgeLand.OwnerID != string(bridge) {
		t.Fatalf("AI transit geçişini fethe çevirdi: bölge sahibi=%q, beklenen=%q", bridgeLand.OwnerID, bridge)
	}
}

func TestAIMovementRejectsLockedBlockedTerrainArea(t *testing.T) {
	from := &world.Region{ID: "from", Neighbors: []world.RegionID{"blocked"}}
	target := &world.Region{ID: "blocked", IsTerrainArea: true, TerrainAreaID: "area", IsLocked: true}
	armyRef := &army.Army{
		ID: "army", OwnerID: "east_rome", RegionID: from.ID, MovePoints: 1,
		Units: []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}},
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			from.ID:   from,
			target.ID: target,
		},
		Armies:       map[army.ArmyID]*army.Army{armyRef.ID: armyRef},
		TerrainAreas: []world.TerrainArea{{ID: "area", MoveCost: 0}},
	}

	executeMove(gs, armyRef, target.ID, faction.FactionID(armyRef.OwnerID))
	if armyRef.RegionID != from.ID {
		t.Fatalf("AI entered blocked terrain area: region=%s", armyRef.RegionID)
	}
}
