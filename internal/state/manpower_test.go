package state

import (
	"testing"

	"mapp-game-go/internal/army"
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

func TestCanQueueLandUnitUsesSharedCapacityAndPendingOrders(t *testing.T) {
	const fid = faction.FactionID("owner")
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: string(fid)},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia": {ID: "militia", Category: army.CategoryInfantry},
		},
		ProductionQueue: []ProductionOrder{{
			Kind: "unit", FactionID: string(fid), TypeID: "militia",
		}},
	}

	if got, want := gs.LandUnitCapacityRemaining(fid), 19; got != want {
		t.Fatalf("bekleyen emir kapasiteye dahil edilmedi: got=%d want=%d", got, want)
	}
	if !gs.CanQueueLandUnit(fid) {
		t.Fatal("boş kapasite varken kara birimi emri reddedildi")
	}

	gs.ProductionQueue = append(gs.ProductionQueue, ProductionOrder{
		Kind: "unit", FactionID: string(fid), TypeID: "militia",
	})
	for gs.LandUnitCapacityRemaining(fid) > 0 {
		gs.ProductionQueue = append(gs.ProductionQueue, ProductionOrder{
			Kind: "unit", FactionID: string(fid), TypeID: "militia",
		})
	}
	if gs.CanQueueLandUnit(fid) {
		t.Fatal("savaşçı kapasitesi doluyken kara birimi emri kabul edildi")
	}
}
