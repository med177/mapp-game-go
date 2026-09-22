package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestRenameSettlementIDUpdatesRuntimeReferences(t *testing.T) {
	rid := world.RegionID("region_a")
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			rid: {
				ID:          rid,
				Settlements: []world.Settlement{{ID: "old_settlement"}},
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"faction_a": {
				CapitalSettlementID:        "old_settlement",
				PendingCapitalSettlementID: "old_settlement",
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_a": {DockedSettlementID: "old_settlement"},
		},
	}
	r := &Renderer{gs: gs}

	r.renameSettlementID(rid, 0, "old_settlement", "new_settlement")

	if got := gs.Regions[rid].Settlements[0].ID; got != "new_settlement" {
		t.Fatalf("settlement ID = %q, want new_settlement", got)
	}
	f := gs.Factions["faction_a"]
	if f.CapitalSettlementID != "new_settlement" || f.PendingCapitalSettlementID != "new_settlement" {
		t.Fatalf("faction settlement references were not renamed: %+v", f)
	}
	if got := gs.Armies["fleet_a"].DockedSettlementID; got != "new_settlement" {
		t.Fatalf("docked settlement ID = %q, want new_settlement", got)
	}
}

func TestRenameRegionIDUpdatesRuntimeAndMapReferences(t *testing.T) {
	oldID := world.RegionID("old_region")
	newID := world.RegionID("new_region")
	wm := &WorldMap{
		regionIDs:         []world.RegionID{"", oldID},
		regionIdx:         map[world.RegionID]uint16{oldID: 1},
		regionPx:          map[world.RegionID][]int{oldID: {0}},
		regionAnchor:      map[world.RegionID][2]int{oldID: {10, 20}},
		settlementAnchor:  map[settlementAnchorKey][2]int{{Region: oldID, Index: 0}: {11, 21}},
		primarySettlement: map[world.RegionID][2]int{oldID: {11, 21}},
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			oldID:      {ID: oldID, Neighbors: []world.RegionID{"neighbor"}},
			"neighbor": {ID: "neighbor", Neighbors: []world.RegionID{oldID}},
		},
		RegionOrder:  []world.RegionID{oldID, "neighbor"},
		LandPassages: []world.LandPassage{{From: oldID, To: "neighbor"}},
	}
	r := &Renderer{
		gs:                       gs,
		worldMap:                 wm,
		editSelectedRegion:       oldID,
		SelectedRegion:           oldID,
		selectedSettlementRegion: oldID,
		lastMapRegionClickID:     oldID,
		editLandPassageFrom:      oldID,
		editNeighborAddFrom:      oldID,
	}

	r.renameRegionID(oldID, newID)

	if gs.Regions[oldID] != nil || gs.Regions[newID] == nil {
		t.Fatalf("region map key was not renamed: %v", gs.Regions)
	}
	if gs.RegionOrder[0] != newID || gs.LandPassages[0].From != newID {
		t.Fatalf("region references were not renamed: order=%v passages=%v", gs.RegionOrder, gs.LandPassages)
	}
	if gs.Regions["neighbor"].Neighbors[0] != newID {
		t.Fatalf("neighbor reference was not renamed: %v", gs.Regions["neighbor"].Neighbors)
	}
	if r.editSelectedRegion != newID || r.SelectedRegion != newID || r.selectedSettlementRegion != newID {
		t.Fatalf("selection references were not renamed: selected=%q map=%q settlement=%q", r.editSelectedRegion, r.SelectedRegion, r.selectedSettlementRegion)
	}
	if _, ok := wm.regionIdx[newID]; !ok {
		t.Fatal("world map region index was not renamed")
	}
}

func TestSettlementIDInUseIgnoresEditedSettlement(t *testing.T) {
	first := world.RegionID("region_a")
	second := world.RegionID("region_b")
	r := &Renderer{gs: &state.GameState{Regions: map[world.RegionID]*world.Region{
		first:  {ID: first, Settlements: []world.Settlement{{ID: "shared_name"}}},
		second: {ID: second, Settlements: []world.Settlement{{ID: "other_name"}}},
	}}}

	if r.settlementIDInUse("shared_name", first, 0) {
		t.Fatal("the edited settlement was reported as a duplicate")
	}
	if !r.settlementIDInUse("shared_name", second, 0) {
		t.Fatal("an ID used by another settlement was not reported")
	}
}
