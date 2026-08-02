package game

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestDemolishBuildingRemovesOneCompletedLevelWithoutRefund(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1", Gold: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "p1", Buildings: []string{"market", "market"}},
		},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", NameTR: "Pazar", MaxPerRegion: 3},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.demolishBuilding("r1", "market")

	if got := gs.Regions["r1"].Buildings; len(got) != 1 || got[0] != "market" {
		t.Fatalf("tek bina seviyesi kaldırılmalıydı, got=%v", got)
	}
	if got := gs.Factions["p1"].Gold; got != 100 {
		t.Fatalf("yıkım kaynak iadesi yapmamalıydı, gold=%d", got)
	}
}

func TestDemolishBuildingRemovesGeneratedPortSettlement(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Factions:        map[faction.FactionID]*faction.Faction{"p1": {ID: "p1"}},
		Regions: map[world.RegionID]*world.Region{
			"coast": {
				ID: "coast", OwnerID: "p1", Buildings: []string{"port"},
				Settlements: []world.Settlement{{ID: "coast_port", Name: "Port", NameTR: "Liman", Type: world.SettlementPort}},
			},
		},
		BuildingTypes: map[string]*city.Building{
			"port": {ID: "port", NameTR: "Liman", MaxPerRegion: 1},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.demolishBuilding("coast", "port")

	if got := gs.Regions["coast"].Settlements; len(got) != 0 {
		t.Fatalf("otomatik liman yerleşimi de kaldırılmalıydı, got=%v", got)
	}
}

func TestDemolishBuildingBlockedWhileUpgradeQueued(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Factions:        map[faction.FactionID]*faction.Faction{"p1": {ID: "p1"}},
		Regions:         map[world.RegionID]*world.Region{"r1": {ID: "r1", OwnerID: "p1", Buildings: []string{"market"}}},
		BuildingTypes:   map[string]*city.Building{"market": {ID: "market", NameTR: "Pazar", MaxPerRegion: 2}},
		ProductionQueue: []state.ProductionOrder{{Kind: productionKindBuilding, FactionID: "p1", RegionID: "r1", TypeID: "market", TurnsLeft: 2}},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.demolishBuilding("r1", "market")

	if got := len(gs.Regions["r1"].Buildings); got != 1 {
		t.Fatalf("bekleyen yükseltme varken bina yıkılmamalıydı, level=%d", got)
	}
}
