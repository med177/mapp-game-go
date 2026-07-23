package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestUnitCostTooltipLinesShowRequiredAmountAndShortageOnly(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1", Gold: 88, Grain: 405, Iron: 461},
		},
	}

	lines := unitCostTooltipLines(gs, economy.ResourceCost{Gold: 220, Grain: 24, Iron: 12})

	if got, want := lines[0].text, "Altın: 220 eksik"; got != want {
		t.Fatalf("altın maliyeti hatalı: got=%q want=%q", got, want)
	}
	if got, want := lines[1].text, "Tahıl: 24"; got != want {
		t.Fatalf("tahıl maliyeti hatalı: got=%q want=%q", got, want)
	}
	if got, want := lines[2].text, "Demir: 12"; got != want {
		t.Fatalf("demir maliyeti hatalı: got=%q want=%q", got, want)
	}

	for _, line := range lines {
		if line.text == "Altın: 88/220 eksik" || line.text == "Tahıl: 405/24" || line.text == "Demir: 461/12" {
			t.Fatalf("mevcut miktar popup maliyetinde gösterilmemeli: %q", line.text)
		}
	}
}

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

func TestRecruitPanelButtonRemainsEnabledWhenGoldIsInsufficient(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"bursa": {
				ID:        "bursa",
				OwnerID:   "p1",
				Buildings: []string{"barracks"},
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1", Gold: 0, Grain: 100, Iron: 100, Timber: 100, Stone: 100},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {
				ID:                "infantry",
				GoldCost:          120,
				RequiredBldg:      "barracks",
				RequiredBldgLevel: 1,
			},
		},
	}

	if !RecruitPanelButtonEnabled(gs, "bursa") {
		t.Fatal("kışlalı bölgede altın yetersiz olsa da Ordu butonu etkin kalmalıydı")
	}
}
