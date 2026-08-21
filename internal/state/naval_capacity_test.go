package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestNavalCapScalesWithRegionsPopulationAndPortLevels(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"a":     {ID: "a", OwnerID: "p1", Population: 1500, Buildings: []string{"port", "port"}},
			"b":     {ID: "b", OwnerID: "p1", Population: 1000},
			"sea":   {ID: "sea", IsSea: true, OwnerID: "p1", Population: 9999, Buildings: []string{"port"}},
			"enemy": {ID: "enemy", OwnerID: "p2", Population: 9000, Buildings: []string{"port", "port"}},
		},
	}

	// 2 kara bölgesi + 2.500 nüfus / 1.000 + 2 liman seviyesi * 2 = 8.
	if got := gs.NavalCap("p1"); got != 8 {
		t.Fatalf("donanma kapasitesi bölge+nüfus+liman seviyesine göre 8 olmalıydı, got=%d", got)
	}
}

func TestNavalUnitsIncludingQueueCountsAllShipCategories(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "p1", Population: 1000, Buildings: []string{"port"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {ID: "fleet", OwnerID: "p1", IsNaval: true, Units: []army.Unit{
				{TypeID: "warship"}, {TypeID: "transport"},
			}},
			"land": {ID: "land", OwnerID: "p1", Units: []army.Unit{{TypeID: "infantry"}}},
		},
		Factions: map[faction.FactionID]*faction.Faction{"p1": {ID: "p1"}},
		UnitTypes: map[string]*army.UnitType{
			"warship":       {ID: "warship", Category: army.CategoryNavalWar, RequiredBldg: "port"},
			"transport":     {ID: "transport", Category: army.CategoryNavalTrans, RequiredBldg: "port"},
			"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade, RequiredBldg: "port"},
		},
		ProductionQueue: []ProductionOrder{
			{Kind: "unit", FactionID: "p1", TypeID: "merchant_ship"},
			{Kind: "unit", FactionID: "p1", TypeID: "infantry"},
			{Kind: "unit", FactionID: "p2", TypeID: "merchant_ship"},
		},
	}

	if got := gs.DeployedNavalUnits("p1"); got != 2 {
		t.Fatalf("aktif donanma yalnız gemileri saymalıydı, got=%d", got)
	}
	if got := gs.PendingNavalUnits("p1"); got != 1 {
		t.Fatalf("kuyrukta yalnız p1 deniz emri sayılmalıydı, got=%d", got)
	}
	if got := gs.NavalUnitsIncludingQueue("p1"); got != 3 {
		t.Fatalf("aktif+kuyruktaki donanma 3 olmalıydı, got=%d", got)
	}
}
