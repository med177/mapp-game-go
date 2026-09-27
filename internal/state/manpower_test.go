package state

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestManpowerCapGrowsWithBarracksLevels(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: "owner", Buildings: []string{"barracks", "barracks"}},
		},
	}
	base := gs.ManpowerCap(faction.FactionID("owner"))
	if want := 20 + 2*landArmyCapacityPerBarracksLevel; base != want {
		t.Fatalf("kışla kapasitesi yanlış: got=%d want=%d", base, want)
	}
	gs.Regions["capital"].Buildings = nil
	if got, want := gs.ManpowerCap("owner"), 20; got != want {
		t.Fatalf("temel kapasite değişti: got=%d want=%d", got, want)
	}
}
