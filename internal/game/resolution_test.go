package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestCheckRegionUnlocksUnlocksTimedRegionAtTurn(t *testing.T) {
	gs := &state.GameState{
		Turn: 5,
		Regions: map[world.RegionID]*world.Region{
			"locked": {ID: "locked", IsLocked: true, UnlockTurn: 5},
		},
		Armies: map[army.ArmyID]*army.Army{},
	}

	unlocked := checkRegionUnlocks(gs)

	if gs.Regions["locked"].IsLocked {
		t.Fatal("timed region açılmadı")
	}
	if len(unlocked) != 1 || unlocked[0] != "locked" {
		t.Fatalf("beklenen unlock listesi [locked], got=%v", unlocked)
	}
}

func TestCheckRegionUnlocksDoesNotUnlockTimedRegionEarlyByAdjacency(t *testing.T) {
	gs := &state.GameState{
		Turn: 4,
		Regions: map[world.RegionID]*world.Region{
			"src":    {ID: "src", Neighbors: []world.RegionID{"locked"}},
			"locked": {ID: "locked", IsLocked: true, UnlockTurn: 5},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", RegionID: "src"},
		},
	}

	unlocked := checkRegionUnlocks(gs)

	if !gs.Regions["locked"].IsLocked {
		t.Fatal("timed region erken açıldı")
	}
	if len(unlocked) != 0 {
		t.Fatalf("erken unlock listesi boş olmalı, got=%v", unlocked)
	}
}

func TestCheckRegionUnlocksUnlocksDiscoveryRegionByAdjacency(t *testing.T) {
	gs := &state.GameState{
		Turn: 4,
		Regions: map[world.RegionID]*world.Region{
			"src":    {ID: "src", Neighbors: []world.RegionID{"locked"}},
			"locked": {ID: "locked", IsLocked: true, UnlockTurn: 0},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", RegionID: "src"},
		},
	}

	unlocked := checkRegionUnlocks(gs)

	if gs.Regions["locked"].IsLocked {
		t.Fatal("discovery region komşulukla açılmadı")
	}
	if len(unlocked) != 1 || unlocked[0] != "locked" {
		t.Fatalf("beklenen unlock listesi [locked], got=%v", unlocked)
	}
}

func TestApplyEconomyTickAddsTradeIncome(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", Religion: religion.Catholic, Gold: 10, Grain: 0, Spice: 10},
			"b": {ID: "b", Religion: religion.Catholic, Gold: 30, Grain: 0, Spice: 0},
		},
		Regions: map[world.RegionID]*world.Region{
			"a1": {ID: "a1", OwnerID: "a", TaxRate: 50, Satisfaction: 50},
			"b1": {ID: "b1", OwnerID: "b", TaxRate: 50, Satisfaction: 50},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "a", ToFactionID: "b", Good: economy.GoodSpice, AmountPerTurn: 2, GoldPerUnit: 12},
		},
		Armies:    map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{},
	}

	applyEconomyTick(gs)

	// a: 10 (başlangıç) + 24 (2 spice * 12 gold, b'den ödeme) = 34, spice: 10 - 2 = 8
	if gs.Factions["a"].Gold != 34 {
		t.Fatalf("ticaret geliri altına eklenmedi, got=%d", gs.Factions["a"].Gold)
	}
	if gs.Factions["a"].Spice != 8 {
		t.Fatalf("ticaret rotası malı kaynaktan çıkarmalıydı, got=%d", gs.Factions["a"].Spice)
	}
	// b: 30 (başlangıç) - 24 (ticaret ödemesi) = 6, spice: 0 + 2 = 2
	if gs.Factions["b"].Gold != 6 {
		t.Fatalf("ticaret rotası hedeften altın çıkarmalıydı, got=%d", gs.Factions["b"].Gold)
	}
	if gs.Factions["b"].Spice != 2 {
		t.Fatalf("ticaret rotası malı hedefe eklemeliydi, got=%d", gs.Factions["b"].Spice)
	}
}
