package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

func TestLandContactRetreatRegionRejectsBlockedTerrainArea(t *testing.T) {
	current := &world.Region{
		ID:        "izmir",
		Neighbors: []world.RegionID{"area::ege_mountains"},
	}
	blocked := &world.Region{
		ID:            "area::ege_mountains",
		IsTerrainArea: true,
		TerrainAreaID: "ege_mountains",
	}
	armyRef := &army.Army{ID: "defender", OwnerID: "byzantium", RegionID: current.ID}
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			current.ID: current,
			blocked.ID: blocked,
		},
		TerrainAreas: []world.TerrainArea{{ID: "ege_mountains", MoveCost: 0}},
	}

	if retreat := gs.LandContactRetreatRegion(armyRef); retreat != "" {
		t.Fatalf("blocked terrain area was selected as retreat target: %s", retreat)
	}
}
