package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestArmyOrganizationPenaltyPercent(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{},
		Armies:  map[army.ArmyID]*army.Army{},
	}
	for i := 0; i < 8; i++ {
		id := world.RegionID("region_" + string(rune('a'+i)))
		gs.Regions[id] = &world.Region{ID: id, OwnerID: "player"}
	}
	for i := 0; i < 5; i++ {
		id := army.ArmyID("army_" + string(rune('a'+i)))
		gs.Armies[id] = &army.Army{ID: id, OwnerID: "player", Units: []army.Unit{{TypeID: "infantry"}}}
	}

	fid := faction.FactionID("player")
	if got := gs.MaxLandArmies(fid); got != 5 {
		t.Fatalf("ordu limiti = %d, want 5", got)
	}
	if got := gs.ArmyOrganizationPenaltyPercent(fid); got != 0 {
		t.Fatalf("limit içindeki organizasyon cezası = %d, want 0", got)
	}

	gs.Armies["army_f"] = &army.Army{ID: "army_f", OwnerID: "player", Units: []army.Unit{{TypeID: "infantry"}}}
	if got := gs.ArmyOrganizationPenaltyPercent(fid); got != 10 {
		t.Fatalf("6/5 organizasyon cezası = %d, want 10", got)
	}
	if got := gs.ArmyOrganizationMultiplier(fid); got != 0.9 {
		t.Fatalf("6/5 organizasyon çarpanı = %v, want 0.9", got)
	}
}
