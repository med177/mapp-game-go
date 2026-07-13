package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestRecruitQueueItemsMarksOnlyCurrentTurnCapacityAsActive(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"r1": {
				ID:        "r1",
				OwnerID:   "p1",
				Buildings: []string{"barracks", "barracks", "port"},
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia":   {ID: "militia", RequiredBldg: "barracks"},
			"transport": {ID: "transport", RequiredBldg: "port"},
		},
		ProductionQueue: []state.ProductionOrder{
			{ID: "prod_1", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "militia", TurnsLeft: 1},
			{ID: "prod_2", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "militia", TurnsLeft: 1},
			{ID: "prod_3", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "militia", TurnsLeft: 1},
			{ID: "prod_4", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "transport", TurnsLeft: 1},
			{ID: "prod_5", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "transport", TurnsLeft: 1},
			{ID: "prod_6", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "militia", TurnsLeft: 2},
		},
	}

	items := recruitQueueItems(gs, "r1")

	if len(items) != 6 {
		t.Fatalf("beklenmeyen queue item sayısı: got=%d", len(items))
	}
	want := []bool{true, true, false, true, false, false}
	for i, active := range want {
		if items[i].progressesThisTurn != active {
			t.Fatalf("item %d active durumu hatalı: got=%v want=%v items=%+v", i, items[i].progressesThisTurn, active, items)
		}
	}
}

func TestRecruitQueueItemsUsesSeparateBarracksAndPortLanes(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"r1": {
				ID:        "r1",
				OwnerID:   "p1",
				Buildings: []string{"barracks", "port", "port"},
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia":   {ID: "militia", RequiredBldg: "barracks"},
			"transport": {ID: "transport", RequiredBldg: "port"},
		},
		ProductionQueue: []state.ProductionOrder{
			{ID: "prod_1", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "militia", TurnsLeft: 1},
			{ID: "prod_2", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "transport", TurnsLeft: 1},
			{ID: "prod_3", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "transport", TurnsLeft: 1},
			{ID: "prod_4", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "transport", TurnsLeft: 1},
		},
	}

	items := recruitQueueItems(gs, "r1")

	want := []bool{true, true, true, false}
	for i, active := range want {
		if items[i].progressesThisTurn != active {
			t.Fatalf("item %d active durumu hatalı: got=%v want=%v items=%+v", i, items[i].progressesThisTurn, active, items)
		}
	}
}

func TestRecruitQueueItemsMarksOnlyFirstCapacityAsProgressing(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"r1": {
				ID:        "r1",
				OwnerID:   "p1",
				Buildings: []string{"barracks", "barracks", "barracks"},
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", RequiredBldg: "barracks"},
		},
		ProductionQueue: []state.ProductionOrder{
			{ID: "prod_1", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "infantry", TurnsLeft: 2},
			{ID: "prod_2", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "infantry", TurnsLeft: 2},
			{ID: "prod_3", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "infantry", TurnsLeft: 2},
			{ID: "prod_4", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "infantry", TurnsLeft: 2},
			{ID: "prod_5", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "infantry", TurnsLeft: 2},
			{ID: "prod_6", Kind: "unit", FactionID: "p1", RegionID: "r1", TypeID: "infantry", TurnsLeft: 2},
		},
	}

	items := recruitQueueItems(gs, "r1")

	want := []bool{true, true, true, false, false, false}
	for i, active := range want {
		if items[i].progressesThisTurn != active {
			t.Fatalf("item %d progress durumu hatalı: got=%v want=%v items=%+v", i, items[i].progressesThisTurn, active, items)
		}
	}
}

func TestRecruitPanelDisabledReasonUsesResourceShortage(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"bursa": {
				ID:      "bursa",
				OwnerID: "p1",
				Buildings: []string{
					"barracks",
				},
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {
				ID:     "p1",
				Gold:   23,
				Grain:  100,
				Iron:   100,
				Timber: 100,
				Stone:  100,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia": {
				ID:                "militia",
				GoldCost:          60,
				GrainCost:         12,
				RequiredBldg:      "",
				RequiredBldgLevel: 0,
			},
		},
	}

	if got := recruitPanelDisabledReason(gs, "bursa"); got != "Yetersiz Altın" {
		t.Fatalf("beklenen yetersiz altın nedeni, got=%q", got)
	}
}
