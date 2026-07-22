package state

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestFactionProductionSummaryExcludesSiegedRegions(t *testing.T) {
	gs := &GameState{
		Month: 6,
		Regions: map[world.RegionID]*world.Region{
			"safe": {
				ID:              "safe",
				OwnerID:         "player",
				Terrain:         world.TerrainPlain,
				BaseGrainOutput: 10,
				BaseIronOutput:  4,
			},
			"sieged": {
				ID:              "sieged",
				OwnerID:         "player",
				Terrain:         world.TerrainPlain,
				BaseGrainOutput: 20,
				BaseIronOutput:  8,
			},
		},
		Sieges: map[world.RegionID]*SiegeState{
			"sieged": {RegionID: "sieged"},
		},
	}

	got := gs.FactionProductionSummary(faction.FactionID("player"))
	if got.Grain != 12 || got.Iron != 4 {
		t.Fatalf("kuşatılmış bölge üretime katılmamalı: got grain=%d iron=%d", got.Grain, got.Iron)
	}
}

func TestFactionGrainNetChangeSubtractsCivilianDemand(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"region": {
				ID:              "region",
				OwnerID:         "player",
				Terrain:         world.TerrainCoast,
				BaseGrainOutput: 10,
				Population:      41,
			},
		},
	}

	if got := gs.FactionGrainNetChange("player"); got != 7 {
		t.Fatalf("tahıl net değişimi sivil talebini düşmeli: got=%d want=7", got)
	}
}
