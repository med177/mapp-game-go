package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAILogisticsUsesSharedGrainProductionDemandAndUpkeepRules(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 0},
		},
		Regions: map[world.RegionID]*world.Region{
			"farm": {
				ID:              "farm",
				OwnerID:         "player",
				Population:      200,
				BaseGrainOutput: 100,
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"field": {ID: "field", OwnerID: "player", RegionID: "farm", Units: []army.Unit{{TypeID: "inf"}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 2},
		},
		ActiveRegionEvents: []state.RegionEventStatus{{
			RegionID:               "farm",
			TurnsLeft:              2,
			GrainProductionPercent: -50,
			GrainDemandPercent:     100,
		}},
	}

	demand, eventCapacity, _ := aiRegionLogistics(gs, gs.Regions["farm"], "player")
	if demand != gs.EffectiveArmyGrainUpkeep(gs.Armies["field"]) {
		t.Fatalf("AI ordu talebi ortak efektif bakım kuralını kullanmalıydı: got=%d", demand)
	}
	if got := gs.RegionMilitaryGrainProduction(gs.Regions["farm"]); got != 26 {
		t.Fatalf("AI lojistik için ortak askeri üretim seam'i yanlış: got=%d", got)
	}

	gs.ActiveRegionEvents = nil
	_, normalCapacity, _ := aiRegionLogistics(gs, gs.Regions["farm"], "player")
	if eventCapacity >= normalCapacity {
		t.Fatalf("aktif tahıl olayı AI ikmal kapasitesini azaltmalıydı: event=%d normal=%d", eventCapacity, normalCapacity)
	}
}
