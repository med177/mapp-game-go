package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/world"
)

func TestGrainStorageCapacityForFactionUsesPopulationArmyAndGranary(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID:         "home",
				OwnerID:    "player",
				Population: 200,
				Buildings:  []string{"granary"},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"field": {
				ID:      "field",
				OwnerID: "player",
				Units:   []army.Unit{{TypeID: "inf"}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 4},
		},
		BuildingTypes: map[string]*city.Building{
			"granary": {ID: "granary", StorageCapacity: 100},
		},
	}

	if got := gs.GrainStorageCapacityForFaction("player"); got != 184 {
		t.Fatalf("HUD kapasitesi nüfus + ordu + ambar bonusunu kullanmalı, got=%d", got)
	}
}

func TestGrainStorageCapacityKeepsMinimumReserve(t *testing.T) {
	if got := GrainStorageCapacity(1, 0, 0); got != 100 {
		t.Fatalf("küçük talep profili minimum 100 kapasite korumalı, got=%d", got)
	}
	if got := GrainStorageCapacity(20, 10, 100); got != 250 {
		t.Fatalf("talep ve ambar bonusu birlikte hesaplanmalı, got=%d", got)
	}
}
