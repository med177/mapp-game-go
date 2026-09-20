package state

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestHistoricalTradeIncomeFollowsDateAndCurrentCenterOwner(t *testing.T) {
	gs := &GameState{
		Year: 1310,
		Factions: map[faction.FactionID]*faction.Faction{
			"alexandria_owner": {ID: "alexandria_owner"},
			"basra_owner":      {ID: "basra_owner"},
		},
		Regions: map[world.RegionID]*world.Region{
			"alexandria": {ID: "alexandria", OwnerID: "alexandria_owner"},
			"basra":      {ID: "basra", OwnerID: "basra_owner"},
		},
		TradeCenters: world.TradeCenterConfig{
			Centers: []world.TradeCenterDef{
				{ID: "alexandria", Links: []world.RegionID{"spice_route"}},
				{ID: "basra", Links: []world.RegionID{"spice_route"}},
				{ID: "spice_route", OffMap: true, NameTR: "Baharat Yolu", Links: []world.RegionID{"alexandria", "basra"}},
				{ID: "cape_route", OffMap: true, NameTR: "Ümit Burnu", UnlockYear: 1498, CompetitionImpacts: []world.TradeCompetitionImpact{{CenterID: "spice_route", IncomePercent: -35, AmountPercent: -35}}},
			},
			HistoricalFlows: []world.HistoricalTradeFlow{
				{FromRegionID: "spice_route", ToRegionID: "alexandria", Good: economy.GoodSpice, AmountPerTurn: 8, GoldIncomePerTurn: 18, StartYear: 1300},
				{FromRegionID: "spice_route", ToRegionID: "basra", Good: economy.GoodSpice, AmountPerTurn: 6, GoldIncomePerTurn: 14, StartYear: 1300},
				{FromRegionID: "spice_route", ToRegionID: "alexandria", Good: economy.GoodSpice, AmountPerTurn: 4, GoldIncomePerTurn: 10, StartYear: 1400},
			},
		},
	}

	if got := gs.HistoricalTradeIncomeForFaction("alexandria_owner"); got != 18 {
		t.Fatalf("alexandria owner income = %d, want 18", got)
	}
	if got := gs.HistoricalTradeIncomeForFaction("basra_owner"); got != 14 {
		t.Fatalf("basra owner income = %d, want 14", got)
	}
	gs.Factions["alexandria_owner"].IsEliminated = true
	if got := gs.HistoricalTradeIncomeForFaction("alexandria_owner"); got != 0 {
		t.Fatalf("eliminated center owner income = %d, want 0", got)
	}
	gs.Factions["alexandria_owner"].IsEliminated = false
	gs.Regions["alexandria"].OwnerID = "basra_owner"
	if got := gs.HistoricalTradeIncomeForFaction("basra_owner"); got != 32 {
		t.Fatalf("new owner income = %d, want 32", got)
	}
	gs.Year = 1498
	if got, want := gs.HistoricalTradeFlowAmount(gs.TradeCenters.HistoricalFlows[0]), 5; got != want {
		t.Fatalf("cape competition amount = %d, want %d", got, want)
	}
	if got, want := gs.HistoricalTradeFlowIncome(gs.TradeCenters.HistoricalFlows[0]), 11; got != want {
		t.Fatalf("cape competition income = %d, want %d", got, want)
	}
}
