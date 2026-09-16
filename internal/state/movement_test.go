package state

import (
	"testing"

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
