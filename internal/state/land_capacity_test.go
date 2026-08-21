package state

import (
	"testing"

	"mapp-game-go/internal/world"
)

func TestManpowerCapTracksMaxLandArmies(t *testing.T) {
	regions := make(map[world.RegionID]*world.Region, 19)
	for i := 0; i < 19; i++ {
		id := world.RegionID("r" + string(rune('a'+i)))
		regions[id] = &world.Region{ID: id, OwnerID: "p1"}
	}
	regions["ra"].Buildings = []string{"barracks", "barracks"}
	gs := &GameState{Regions: regions}

	if got := gs.MaxLandArmies("p1"); got != 10 {
		t.Fatalf("19 kara bölgesi 10 ordu sınırı vermeli, got=%d", got)
	}
	if got := gs.ManpowerCap("p1"); got != 200 {
		t.Fatalf("10 ordu 200 savaşçı sınırı vermeli, got=%d", got)
	}
}
