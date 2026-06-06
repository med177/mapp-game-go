package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestApplyConquestWithNavalEvictionUndocksPreviousOwnerFleet(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"old_owner": {ID: "old_owner"},
			"new_owner": {ID: "new_owner"},
		},
		Regions: map[world.RegionID]*world.Region{
			"land_a": {ID: "land_a", OwnerID: "old_owner", Neighbors: []world.RegionID{"sea_near"}},
			"land_b": {ID: "land_b", OwnerID: "old_owner"},
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

	result := g.applyConquestWithNavalEviction(gs.Regions["land_a"], "new_owner")

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
	if result.FactionID != "" {
		t.Fatalf("fraksiyon tamamen yikilmamaliydi, got=%+v", result)
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

func TestSanitizeDockedFleetsKeepsAlliedPortDocking(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"ally_land": {
				ID:          "ally_land",
				OwnerID:     "ally",
				Neighbors:   []world.RegionID{"sea_near"},
				Settlements: []world.Settlement{{ID: "ally_port", Type: world.SettlementPort, NameTR: "Müttefik Limanı"}},
			},
			"sea_near": {ID: "sea_near", IsSea: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_player": {
				ID:                 "fleet_player",
				OwnerID:            "player",
				RegionID:           "sea_near",
				DockedRegionID:     "ally_land",
				DockedSettlementID: "ally_port",
				IsNaval:            true,
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ally", "player"): {FactionA: "ally", FactionB: "player", Stance: faction.StanceAllied},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.sanitizeDockedFleets()

	fleet := gs.Armies["fleet_player"]
	if fleet.DockedRegionID != "ally_land" || fleet.DockedSettlementID != "ally_port" {
		t.Fatalf("müttefik liman bağı korunmalıydı, got docked_region=%q docked_settlement=%q", fleet.DockedRegionID, fleet.DockedSettlementID)
	}
}

func TestApplyConquestWithNavalEvictionTransfersDefeatedFactionForces(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"old_owner": {ID: "old_owner"},
			"new_owner": {ID: "new_owner", NameTR: "Yeni Sahip"},
		},
		Regions: map[world.RegionID]*world.Region{
			"land_a": {
				ID:          "land_a",
				OwnerID:     "old_owner",
				Neighbors:   []world.RegionID{"sea_near"},
				Settlements: []world.Settlement{{ID: "port_a", Type: world.SettlementPort, NameTR: "Liman"}},
			},
			"sea_near": {ID: "sea_near", IsSea: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_old": {
				ID:       "army_old",
				OwnerID:  "old_owner",
				RegionID: "land_a",
			},
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
	g := &Game{gs: gs, renderer: render.New(gs)}

	result := g.applyConquestWithNavalEviction(gs.Regions["land_a"], "new_owner")

	if !gs.Factions["old_owner"].IsEliminated {
		t.Fatal("son kara bolgesi dusen fraksiyon elenmeliydi")
	}
	if result.FactionID != "old_owner" || result.SuccessorID != "new_owner" {
		t.Fatalf("beklenen yikilis ozeti donmedi, got=%+v", result)
	}
	if result.TransferredArmies != 1 || result.TransferredFleets != 1 {
		t.Fatalf("devralinan kuvvet sayilari hatali, got=%+v", result)
	}
	if gs.Armies["army_old"].OwnerID != "new_owner" {
		t.Fatalf("kara ordusu galibe gecmeliydi, got=%s", gs.Armies["army_old"].OwnerID)
	}
	if gs.Armies["fleet_old"].OwnerID != "new_owner" {
		t.Fatalf("donanma galibe gecmeliydi, got=%s", gs.Armies["fleet_old"].OwnerID)
	}
	if gs.Armies["fleet_old"].DockedRegionID != "land_a" {
		t.Fatalf("devralinan filo gecerli liman bagini korumaliydi, got=%s", gs.Armies["fleet_old"].DockedRegionID)
	}
}
