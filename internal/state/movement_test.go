package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

func TestLandRegionMoveCostKeepsNormalRegionPassableInsideTerrainPolygon(t *testing.T) {
	gs := &GameState{TerrainAreas: []world.TerrainArea{{
		ID:       "alps",
		MoveCost: 0,
		Polygons: [][][2]int{{
			{0, 0}, {20, 0}, {20, 20}, {0, 20},
		}},
	}}}

	cost, isBlocked := gs.LandRegionMoveCost(&world.Region{
		ID: "germiyan", WorldX: 10, WorldY: 10,
	})
	if isBlocked || cost != 1 {
		t.Fatalf("normal region should remain passable inside terrain polygon, got cost=%d blocked=%v", cost, isBlocked)
	}
}

func TestLandRegionEntryCostRejectsLockedOrBlockedTerrain(t *testing.T) {
	from := world.RegionID("from")
	gs := &GameState{
		LandPassages: nil,
		TerrainAreas: []world.TerrainArea{{ID: "mountain", MoveCost: 0}},
	}

	tests := []struct {
		name   string
		target *world.Region
	}{
		{
			name: "locked terrain",
			target: &world.Region{
				ID: "area::mountain", IsTerrainArea: true, TerrainAreaID: "mountain", IsLocked: true,
			},
		},
		{
			name: "stale unlocked terrain",
			target: &world.Region{
				ID: "area::mountain", IsTerrainArea: true, TerrainAreaID: "mountain", IsLocked: false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if cost, allowed := gs.LandRegionEntryCost(from, test.target); allowed || cost != 0 {
				t.Fatalf("blocked terrain entry unexpectedly allowed: cost=%d allowed=%v", cost, allowed)
			}
		})
	}
}

func TestRepairArmiesInBlockedTerrainMovesArmyToOwnedPassableNeighbor(t *testing.T) {
	blocked := &world.Region{ID: "blocked", OwnerID: "f", IsTerrainArea: true, TerrainAreaID: "area", Neighbors: []world.RegionID{"safe"}}
	safe := &world.Region{ID: "safe", OwnerID: "f"}
	current := &army.Army{ID: "army", OwnerID: "f", RegionID: blocked.ID}
	gs := &GameState{
		Regions:      map[world.RegionID]*world.Region{blocked.ID: blocked, safe.ID: safe},
		TerrainAreas: []world.TerrainArea{{ID: "area", MoveCost: 0}},
		Armies:       map[army.ArmyID]*army.Army{current.ID: current},
	}

	if repaired := gs.RepairArmiesInBlockedTerrain(); repaired != 1 {
		t.Fatalf("expected one repaired army, got %d", repaired)
	}
	if current.RegionID != safe.ID || current.PreviousRegionID != blocked.ID {
		t.Fatalf("unexpected repaired army location: region=%s previous=%s", current.RegionID, current.PreviousRegionID)
	}
	if repaired := gs.RepairArmiesInBlockedTerrain(); repaired != 0 {
		t.Fatalf("repair should be idempotent after the first migration, got %d", repaired)
	}
}
