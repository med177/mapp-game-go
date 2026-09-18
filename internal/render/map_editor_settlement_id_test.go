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
