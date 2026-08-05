package victory

import (
	"strconv"
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

	if gs.Phase != state.PhasePlayerTurn {
		t.Fatalf("expected game to continue, got %s", gs.Phase)
	}
	if gs.WinnerID != gs.PlayerFactionID {
		t.Fatalf("expected winner %s, got %s", gs.PlayerFactionID, gs.WinnerID)
	}
	if !gs.VictoryAchieved {
		t.Fatal("expected victory achieved flag")
	}
}

func TestConquerCityVictoryRequiresAllTargets(t *testing.T) {
	gs := &state.GameState{
		Turn:            1,
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: faction.FactionID("ottoman"),
		Victory: state.VictoryCondition{
			Type:            state.VictoryConquerCity,
			RequiredRegions: []world.RegionID{"constantinople", "ankara"},
		},
		Regions: map[world.RegionID]*world.Region{
			"constantinople": {
				ID:      "constantinople",
				OwnerID: "ottoman",
			},
			"ankara": {
				ID:      "ankara",
				OwnerID: "karaman",
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman"},
		},
	}

	Check(gs)

	if gs.VictoryAchieved {
		t.Fatal("tum hedefler alinmadan conquer_city zaferi olmamali")
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
				OwnerID: "east_rome",
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman"},
		},
	}

	Check(gs)

	if gs.VictoryAchieved {
		t.Fatal("expected game to continue before target ownership")
	}
}

func TestCurrentGoldIncomeIncludesRegionsTradeAndTech(t *testing.T) {
	gs := &state.GameState{
		Month:           7,
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
			"a": {ID: "a", OwnerID: "ottoman", BaseGoldIncome: 100, TaxRate: 50, Satisfaction: 50, TradeCapacity: 10, Buildings: []string{"market"}},
			"b": {ID: "b", OwnerID: "ottoman", BaseGoldIncome: 80, TaxRate: 50, Satisfaction: 50},
		},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", GoldMod: 2, TradeCapacityMod: 1.5},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "ottoman", ToFactionID: "venice", Good: economy.GoodSpice, AmountPerTurn: 4, GoldPerUnit: 10},
		},
		TechTypes: map[string]*tech.Technology{
			"tax_office": {ID: "tax_office", Effects: tech.Effects{GoldPerRegion: 5}},
		},
	}

	got := CurrentGoldIncome(gs)

	if got != 220 {
		t.Fatalf("beklenen gelir 220, got=%d", got)
	}
	if got := GoldIncomeForFaction(gs, "ottoman"); got != 220 {
		t.Fatalf("seçilen devlet geliri mevcut gelirle aynı olmalı: got=%d", got)
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
	if gs.VictoryAchieved {
		t.Fatal("tek turda ekonomik zafer olmamali")
	}

	Check(gs)
	if gs.Phase != state.PhasePlayerTurn {
		t.Fatalf("zaferden sonra oyun devam etmeli, got=%s", gs.Phase)
	}
	if gs.WinnerID != "ottoman" {
		t.Fatalf("kazanan ottoman olmali, got=%s", gs.WinnerID)
	}
	if !gs.VictoryAchieved {
		t.Fatal("victory achieved bekleniyordu")
	}
}

func TestSurviveTurnsVictoryTriggersAtTargetTurn(t *testing.T) {
	gs := &state.GameState{
		Turn:            80,
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: "genoa",
		Victory: state.VictoryCondition{
			Type:        state.VictorySurviveTurns,
			TargetTurns: 80,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"genoa": {ID: "genoa"},
		},
		Regions: map[world.RegionID]*world.Region{
			"holland": {ID: "holland", OwnerID: "genoa"},
		},
	}

	Check(gs)

	if !gs.VictoryAchieved {
		t.Fatal("survive_turns zaferi bekleniyordu")
	}
	if gs.WinnerID != "genoa" {
		t.Fatalf("kazanan genoa olmali, got=%s", gs.WinnerID)
	}
}

func TestSurviveTurnsVictoryWaitsBeforeTargetTurn(t *testing.T) {
	gs := &state.GameState{
		Turn:            79,
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: "genoa",
		Victory: state.VictoryCondition{
			Type:        state.VictorySurviveTurns,
			TargetTurns: 80,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"genoa": {ID: "genoa"},
		},
		Regions: map[world.RegionID]*world.Region{
			"holland": {ID: "holland", OwnerID: "genoa"},
		},
	}

	Check(gs)

	if gs.VictoryAchieved {
		t.Fatal("hedef turdan once survive_turns zaferi olmamali")
	}
}

func TestDeadlineMonthIsInclusive(t *testing.T) {
	gs := &state.GameState{
		Turn:            10,
		Year:            1561,
		Month:           1,
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: "ottoman",
		Victory: state.VictoryCondition{
			Type:            state.VictoryConquerCity,
			RequiredRegions: []world.RegionID{"constantinople"},
			DeadlineYear:    1561,
			DeadlineMonth:   1,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman"},
		},
		Regions: map[world.RegionID]*world.Region{
			"constantinople": {ID: "constantinople", OwnerID: "east_rome"},
			"bithynia":       {ID: "bithynia", OwnerID: "ottoman"},
		},
	}

	Check(gs)

	if gs.Phase == state.PhaseGameOver {
		t.Fatal("deadline ayi icinde oyun bitmemeli")
	}
	if gs.VictoryAchieved {
		t.Fatal("hedef saglanmadan zafer olmamali")
	}
}

func TestDeadlineFailureTriggersAfterDeadlineMonth(t *testing.T) {
	gs := &state.GameState{
		Turn:            11,
		Year:            1561,
		Month:           2,
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: "ottoman",
		Victory: state.VictoryCondition{
			Type:            state.VictoryConquerCity,
			RequiredRegions: []world.RegionID{"constantinople"},
			DeadlineYear:    1561,
			DeadlineMonth:   1,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman"},
		},
		Regions: map[world.RegionID]*world.Region{
			"constantinople": {ID: "constantinople", OwnerID: "east_rome"},
			"bithynia":       {ID: "bithynia", OwnerID: "ottoman"},
		},
	}

	Check(gs)

	if gs.Phase != state.PhaseGameOver {
		t.Fatal("deadline gecince oyun bitmeli")
	}
	if gs.WinnerID != "" {
		t.Fatalf("deadline kaybinda AI kazanan olmamali, got=%s", gs.WinnerID)
	}
	if gs.VictoryAchieved {
		t.Fatal("deadline kaybi zafer olarak isaretlenmemeli")
	}
}

func TestAIGrowthDoesNotAutoWinBeforePlayerDeadline(t *testing.T) {
	gs := &state.GameState{
		Turn:            20,
		Year:            1500,
		Month:           6,
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: "ottoman",
		Victory: state.VictoryCondition{
			Type:            state.VictoryConquerCity,
			RequiredRegions: []world.RegionID{"constantinople"},
			DeadlineYear:    1561,
			DeadlineMonth:   1,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman":   {ID: "ottoman"},
			"east_rome": {ID: "east_rome"},
			"france":    {ID: "france"},
		},
		Regions: map[world.RegionID]*world.Region{
			"constantinople": {ID: "constantinople", OwnerID: "east_rome"},
			"bithynia":       {ID: "bithynia", OwnerID: "ottoman"},
		},
	}

	for i := 0; i < 35; i++ {
		id := world.RegionID("ai_region_" + strconv.Itoa(i))
		gs.Regions[id] = &world.Region{ID: id, OwnerID: "france"}
	}

	Check(gs)

	if gs.Phase == state.PhaseGameOver {
		t.Fatal("AI buyudugu icin otomatik kazanmamali")
	}
	if gs.WinnerID != "" {
		t.Fatalf("AI auto-win olmamali, got=%s", gs.WinnerID)
	}
}
