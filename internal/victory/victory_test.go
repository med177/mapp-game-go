package victory

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

func TestConquerCityVictoryTriggersWhenTargetOwned(t *testing.T) {
	gs := &state.GameState{
		Turn:            1,
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: faction.FactionID("ottoman"),
		Victory: state.VictoryCondition{
			Type:            state.VictoryConquerCity,
			RequiredRegions: []world.RegionID{"constantinople"},
		},
		Regions: map[world.RegionID]*world.Region{
			"bithynia": {
				ID:      "bithynia",
				OwnerID: "ottoman",
			},
			"constantinople": {
				ID:      "constantinople",
				OwnerID: "ottoman",
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman"},
		},
	}

	Check(gs)

	if gs.Phase != state.PhaseGameOver {
		t.Fatalf("expected game over, got %s", gs.Phase)
	}
	if gs.WinnerID != gs.PlayerFactionID {
		t.Fatalf("expected winner %s, got %s", gs.PlayerFactionID, gs.WinnerID)
	}
}

func TestConquerCityVictoryWaitsForTargetOwnership(t *testing.T) {
	gs := &state.GameState{
		Turn:            1,
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: faction.FactionID("ottoman"),
		Victory: state.VictoryCondition{
			Type:            state.VictoryConquerCity,
			RequiredRegions: []world.RegionID{"constantinople"},
		},
		Regions: map[world.RegionID]*world.Region{
			"bithynia": {
				ID:      "bithynia",
				OwnerID: "ottoman",
			},
			"constantinople": {
				ID:      "constantinople",
				OwnerID: "byzantine",
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman"},
		},
	}

	Check(gs)

	if gs.Phase == state.PhaseGameOver {
		t.Fatal("expected game to continue before target ownership")
	}
}

func TestCurrentGoldIncomeIncludesRegionsTradeAndTech(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "ottoman",
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {
				ID: "ottoman",
				Research: faction.ResearchState{
					Completed: map[string]bool{"tax_office": true},
				},
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"a": {ID: "a", OwnerID: "ottoman", BaseGoldIncome: 100, TaxRate: 50, Satisfaction: 50, Buildings: []string{"market"}},
			"b": {ID: "b", OwnerID: "ottoman", BaseGoldIncome: 80, TaxRate: 50, Satisfaction: 50},
		},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", GoldMod: 2},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "ottoman", ToFactionID: "venice", Good: economy.GoodSpice, AmountPerTurn: 10, GoldPerUnit: 10},
		},
		TechTypes: map[string]*tech.Technology{
			"tax_office": {ID: "tax_office", Effects: tech.Effects{GoldPerRegion: 5}},
		},
	}

	got := CurrentGoldIncome(gs)

	if got != 250 {
		t.Fatalf("beklenen gelir 250, got=%d", got)
	}
}

func TestEconomicVictoryUsesIncomeThreshold(t *testing.T) {
	gs := &state.GameState{
		Turn:            2,
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: "ottoman",
		Victory: state.VictoryCondition{
			Type:             state.VictoryEconomic,
			TargetGoldIncome: 120,
			GoldHoldTurns:    2,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman", Gold: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"a": {ID: "a", OwnerID: "ottoman", BaseGoldIncome: 120, TaxRate: 100, Satisfaction: 50},
		},
	}

	Check(gs)
	if gs.EconomicVictoryTurns != 1 {
		t.Fatalf("ilk tur sayaci 1 olmali, got=%d", gs.EconomicVictoryTurns)
	}
	if gs.Phase == state.PhaseGameOver {
		t.Fatal("tek turda ekonomik zafer olmamali")
	}

	Check(gs)
	if gs.Phase != state.PhaseGameOver {
		t.Fatalf("ikinci tur sonunda oyun bitmeli, got=%s", gs.Phase)
	}
	if gs.WinnerID != "ottoman" {
		t.Fatalf("kazanan ottoman olmali, got=%s", gs.WinnerID)
	}
}
