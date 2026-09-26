package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIConsolidatesOverLimitArmiesDespiteRegionalSupplyPressure(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{},
		Armies:  map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Attack: 10, HP: 100, GrainUpkeep: 10},
		},
	}
	for i := 0; i < 8; i++ {
		id := world.RegionID("region_" + string(rune('a'+i)))
		gs.Regions[id] = &world.Region{ID: id, OwnerID: "ai", Neighbors: nil}
	}
	for i := 0; i < 6; i++ {
		id := army.ArmyID("army_" + string(rune('a'+i)))
		gs.Armies[id] = &army.Army{
			ID:       id,
			OwnerID:  "ai",
			RegionID: "region_a",
			Units:    []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
		}
	}

	if got := gs.MaxLandArmies("ai"); got != 5 {
		t.Fatalf("AI ordu limiti = %d, want 5", got)
	}
	if got := gs.ArmyOrganizationPenaltyPercent("ai"); got != 10 {
		t.Fatalf("AI organizasyon cezası = %d, want 10", got)
	}
	_, _, overload := aiRegionLogistics(gs, gs.Regions["region_a"], "ai")
	if overload <= 0 {
		t.Fatalf("fixture bölgesel ikmal baskısı üretmedi: overload=%d", overload)
	}
	if !aiShouldConsolidateInRegion(gs, gs.Regions["region_a"], "ai", false) {
		t.Fatal("ordu slotu aşılmışken AI birleşme kararını reddetti")
	}
	if got, want := gs.EffectiveArmyStrength(gs.Armies["army_a"]), 9; got != want {
		t.Fatalf("AI efektif güç = %d, want %d", got, want)
	}

	aiConsolidateArmies(gs, "ai")
	if got := gs.CurrentLandArmies("ai"); got != 5 {
		t.Fatalf("konsolidasyon sonrası AI ordu sayısı = %d, want 5", got)
	}
}
