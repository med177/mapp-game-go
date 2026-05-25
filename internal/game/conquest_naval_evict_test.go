package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestApplyConquestWithNavalEvictionUndocksPreviousOwnerFleet(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"land_a": {ID: "land_a", OwnerID: "old_owner", Neighbors: []world.RegionID{"sea_near"}},
			"sea_near": {
				ID:    "sea_near",
				IsSea: true,
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_old": {
				ID:                 "fleet_old",
				OwnerID:            "old_owner",
				IsNaval:            true,
				RegionID:           "sea_near",
				DockedRegionID:     "land_a",
				DockedSettlementID: "port_a",
			},
		},
	}
	g := &Game{gs: gs}

	g.applyConquestWithNavalEviction(gs.Regions["land_a"], "new_owner")

	fleet := gs.Armies["fleet_old"]
	if fleet == nil {
		t.Fatal("fleet_old bulunamadı")
	}
	if fleet.RegionID != "sea_near" {
		t.Fatalf("filo en yakin denizde kalmaliydi, got=%s", fleet.RegionID)
	}
	if fleet.DockedRegionID != "" || fleet.DockedSettlementID != "" {
		t.Fatalf("filo limandan ayrilmis olmaliydi, docked_region=%q docked_settlement=%q", fleet.DockedRegionID, fleet.DockedSettlementID)
	}
	if gs.Regions["land_a"].OwnerID != "new_owner" {
		t.Fatalf("bolge sahipligi degismeliydi, got=%s", gs.Regions["land_a"].OwnerID)
	}
}

func TestSanitizeDockedFleetsUndocksForeignFleetFromOwnedPort(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"land_my": {ID: "land_my", OwnerID: "player", Neighbors: []world.RegionID{"sea_near"}},
			"sea_near": {
				ID:    "sea_near",
				IsSea: true,
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_foreign": {
				ID:                 "fleet_foreign",
				OwnerID:            "other",
				IsNaval:            true,
				RegionID:           "sea_near",
				DockedRegionID:     "land_my",
				DockedSettlementID: "port_my",
			},
		},
	}
	g := &Game{gs: gs}

	g.sanitizeDockedFleets()

	fleet := gs.Armies["fleet_foreign"]
	if fleet == nil {
		t.Fatal("fleet_foreign bulunamadı")
	}
	if fleet.RegionID != "sea_near" {
		t.Fatalf("filo en yakin deniz bolgesine cikmaliydi, got=%s", fleet.RegionID)
	}
	if fleet.DockedRegionID != "" || fleet.DockedSettlementID != "" {
		t.Fatalf("yabanci liman bagi temizlenmeliydi, docked_region=%q docked_settlement=%q", fleet.DockedRegionID, fleet.DockedSettlementID)
	}
}
