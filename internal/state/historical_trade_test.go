package state

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func historicalTradeLinks(ids ...world.RegionID) []world.TradeCenterLink {
	links := make([]world.TradeCenterLink, 0, len(ids))
	for _, id := range ids {
		links = append(links, world.TradeCenterLink{RegionID: id})
	}
	return links
}

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
				{ID: "alexandria", Links: historicalTradeLinks("spice_route")},
				{ID: "basra", Links: historicalTradeLinks("spice_route")},
				{ID: "spice_route", OffMap: true, NameTR: "Baharat Yolu", Links: historicalTradeLinks("alexandria", "basra")},
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

func TestDerivedHistoricalFlowsPropagateMultipleGoodsThroughConnectedCenters(t *testing.T) {
	gs := &GameState{
		Year:     1310,
		Factions: map[faction.FactionID]*faction.Faction{"owner": {ID: "owner"}},
		Regions: map[world.RegionID]*world.Region{
			"azerbaijan": {ID: "azerbaijan", OwnerID: "owner", BaseGrainOutput: 20},
			"trebizond":  {ID: "trebizond", OwnerID: "owner", BaseIronOutput: 20},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{
			{ID: "silk_road", OffMap: true, NameTR: "İpek Yolu", Links: historicalTradeLinks("azerbaijan"), SourceGoods: []world.HistoricalTradeGood{{Good: economy.GoodCloth, AmountPerTurn: 6}}},
			{ID: "azerbaijan", Links: historicalTradeLinks("trebizond")},
			{ID: "trebizond", Links: historicalTradeLinks("azerbaijan")},
		}},
	}

	seen := map[string]bool{}
	for _, flow := range gs.ActiveHistoricalTradeFlows() {
		seen[string(flow.FromRegionID)+"->"+string(flow.ToRegionID)+":"+string(flow.Good)] = true
	}
	for _, want := range []string{
		"silk_road->azerbaijan:cloth",
		"azerbaijan->trebizond:cloth",
		"azerbaijan->trebizond:grain",
		"trebizond->azerbaijan:iron",
	} {
		if !seen[want] {
			t.Fatalf("derived flow %q missing; flows = %v", want, seen)
		}
	}
	if seen["azerbaijan->silk_road:grain"] {
		t.Fatal("natural production must not flow back into the Silk Road source center")
	}
	if seen["azerbaijan->silk_road:cloth"] {
		t.Fatal("Silk Road source goods must not flow back into the source center")
	}
}

func TestDerivedHistoricalFlowsCarryAmericaGoldThroughAtlanticToRealOwner(t *testing.T) {
	gs := &GameState{
		Year:     1500,
		Factions: map[faction.FactionID]*faction.Faction{"castile": {ID: "castile"}},
		Regions: map[world.RegionID]*world.Region{
			"portugal": {ID: "portugal", OwnerID: "castile"},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{
			{ID: "north_america_route", OffMap: true, NameTR: "Kuzey Amerika Yolu", Links: historicalTradeLinks("atlantic_route"), SourceGoods: []world.HistoricalTradeGood{{Good: economy.GoodGold, AmountPerTurn: 10, GoldIncomePerTurn: 25}}},
			{ID: "atlantic_route", OffMap: true, NameTR: "Atlantik Yolu", Links: historicalTradeLinks("portugal")},
			{ID: "portugal"},
		}},
	}

	seenGoldToPortugal := false
	for _, flow := range gs.ActiveHistoricalTradeFlows() {
		if flow.Good == economy.GoodGold && flow.FromRegionID == "atlantic_route" && flow.ToRegionID == "portugal" {
			seenGoldToPortugal = true
			break
		}
	}
	if !seenGoldToPortugal {
		t.Fatal("America gold must pass through the Atlantic route to Portugal")
	}
	if got := gs.HistoricalTradeIncomeForFaction("castile"); got <= 0 {
		t.Fatalf("the real owner of Portugal must receive America gold income, got=%d", got)
	}
}

func TestDerivedHistoricalFlowsDoNotReturnGoodsIntoSourceRoutes(t *testing.T) {
	gs := &GameState{
		Year:     1310,
		Factions: map[faction.FactionID]*faction.Faction{"owner": {ID: "owner"}},
		Regions: map[world.RegionID]*world.Region{
			"portugal": {ID: "portugal", OwnerID: "owner", BaseGrainOutput: 20},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{
			{ID: "atlantic_route", OffMap: true, NameTR: "Atlantik Yolu", Links: historicalTradeLinks("portugal")},
			// This reverse link is intentionally invalid for a source route.
			{ID: "portugal", Links: historicalTradeLinks("atlantic_route")},
		}},
	}

	for _, flow := range gs.ActiveHistoricalTradeFlows() {
		if flow.FromRegionID == "portugal" && flow.ToRegionID == "atlantic_route" {
			t.Fatalf("goods must not flow back into a source route: %+v", flow)
		}
	}
}

func TestDerivedHistoricalFlowsTreatOneCenterLinkAsBidirectionalTrade(t *testing.T) {
	gs := &GameState{
		Year:     1310,
		Factions: map[faction.FactionID]*faction.Faction{"owner": {ID: "owner"}},
		Regions: map[world.RegionID]*world.Region{
			"aleppo":     {ID: "aleppo", OwnerID: "owner", BaseGrainOutput: 20},
			"alexandria": {ID: "alexandria", OwnerID: "owner", BaseIronOutput: 20},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{
			{ID: "source_route", OffMap: true, NameTR: "Kaynak Yolu", Links: historicalTradeLinks("aleppo"), SourceGoods: []world.HistoricalTradeGood{{Good: economy.GoodCloth, AmountPerTurn: 1}}},
			{ID: "aleppo", Links: historicalTradeLinks("alexandria")},
			{ID: "alexandria"},
		}},
	}

	seen := map[string]bool{}
	for _, flow := range gs.ActiveHistoricalTradeFlows() {
		seen[string(flow.FromRegionID)+"->"+string(flow.ToRegionID)+":"+string(flow.Good)] = true
	}
	for _, want := range []string{
		"aleppo->alexandria:grain",
		"alexandria->aleppo:iron",
	} {
		if !seen[want] {
			t.Fatalf("one center link should support both trade directions; missing %q in %v", want, seen)
		}
	}
}
