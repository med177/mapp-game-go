package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/world"
)

func TestLandRegionMoveCostConvertsNormalRegionShapeCoordinates(t *testing.T) {
	offsetX, offsetY := -530.0, -180.0
	scaleX, scaleY := 2.025, 2.025
	gs := &GameState{MapConfig: scenario.MapConfig{
		ShapeOffsetX: &offsetX,
		ShapeOffsetY: &offsetY,
		ShapeScaleX:  &scaleX,
		ShapeScaleY:  &scaleY,
	}, TerrainAreas: []world.TerrainArea{{
		ID:       "alps",
		MoveCost: 0,
		Polygons: [][][2]int{{
			{1598, 792}, {1608, 792}, {1608, 802}, {1598, 802},
		}},
	}}}

	blocked, isBlocked := gs.LandRegionMoveCost(&world.Region{
		ID: "bithynia", WorldX: 1053, WorldY: 482,
	})
	if !isBlocked || blocked != 0 {
		t.Fatalf("expected transformed region center to be blocked, got cost=%d blocked=%v", blocked, isBlocked)
	}

	_, isBlocked = gs.LandRegionMoveCost(&world.Region{
		ID: "far_region", WorldX: 787, WorldY: 327,
	})
	if isBlocked {
		t.Fatal("a region outside the terrain polygon was incorrectly blocked")
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
