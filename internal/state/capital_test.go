package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestArmyReplenishmentHPDoublesAtFactionCapital(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", CapitalSettlementID: "capital_city"},
			"vassal": {ID: "vassal", CapitalSettlementID: "vassal_city", OverlordID: "player"},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {
				ID:      "capital",
				OwnerID: "player",
				Buildings: []string{
					"farm", "farm", "granary",
				},
				Settlements: []world.Settlement{{ID: "capital_city", IsCenter: true}},
			},
			"field": {
				ID:      "field",
				OwnerID: "player",
				Buildings: []string{
					"farm", "farm", "granary",
				},
			},
			"vassal_capital": {
				ID:      "vassal_capital",
				OwnerID: "vassal",
				Buildings: []string{
					"farm", "farm", "granary",
				},
				Settlements: []world.Settlement{{ID: "vassal_city", IsCenter: true}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"capital_army":        {ID: "capital_army", OwnerID: "player", RegionID: "capital"},
			"field_army":          {ID: "field_army", OwnerID: "player", RegionID: "field"},
			"player_vassal_army":  {ID: "player_vassal_army", OwnerID: "player", RegionID: "vassal_capital"},
			"vassal_capital_army": {ID: "vassal_capital_army", OwnerID: "vassal", RegionID: "vassal_capital"},
		},
	}

	if got, want := gs.ArmyReplenishmentHP(gs.Armies["capital_army"]), 16; got != want {
		t.Fatalf("başkent ordusu normal toparlanmanın iki katını almalıydı, got=%d want=%d", got, want)
	}
	if got, want := gs.ArmyReplenishmentHP(gs.Armies["field_army"]), 8; got != want {
		t.Fatalf("normal bölgenin toparlanma değeri değişmemeli, got=%d want=%d", got, want)
	}
	if got, want := gs.ArmyReplenishmentHP(gs.Armies["player_vassal_army"]), 8; got != want {
		t.Fatalf("oyuncu ordusu vassal başkentinde başkent bonusu almamalı, got=%d want=%d", got, want)
	}
	if got, want := gs.ArmyReplenishmentHP(gs.Armies["vassal_capital_army"]), 16; got != want {
		t.Fatalf("vassalın kendi ordusu kendi başkentinde bonus almalı, got=%d want=%d", got, want)
	}
}

func TestBestCapitalSettlementPrefersMostDevelopedRegionOverHigherIncome(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{"ai": {ID: "ai"}},
		Regions: map[world.RegionID]*world.Region{
			"rich": {
				ID: "rich", OwnerID: "ai", BaseGoldIncome: 500,
				Settlements: []world.Settlement{{ID: "rich_city", Type: world.SettlementCity, IsCenter: true}},
			},
			"developed": {
				ID: "developed", OwnerID: "ai", BaseGoldIncome: 50,
				Buildings:   []string{"market", "market", "walls"},
				Settlements: []world.Settlement{{ID: "developed_city", Type: world.SettlementCity, IsCenter: true}},
			},
		},
	}

	if got, ok := gs.BestCapitalSettlementForFaction("ai"); !ok || got != "developed_city" {
		t.Fatalf("en gelişmiş bölgenin başkenti seçilmeliydi: ok=%v got=%q", ok, got)
	}
}

func TestNormalizeFactionCapitalsChoosesBestOwnedSettlement(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ankara": {
				ID:             "ankara",
				OwnerID:        "ottoman",
				BaseGoldIncome: 30,
				TradeCapacity:  2,
				Settlements: []world.Settlement{
					{ID: "ankara_city", NameTR: "Ankara", IsCenter: true},
				},
			},
			"bursa": {
				ID:              "bursa",
				OwnerID:         "ottoman",
				BaseGoldIncome:  90,
				BaseGrainOutput: 6,
				TradeCapacity:   8,
				Buildings:       []string{"market"},
				Settlements: []world.Settlement{
					{ID: "bursa_city", NameTR: "Bursa", IsCenter: true},
				},
			},
		},
	}

	gs.NormalizeFactionCapitals()

	if got := gs.Factions["ottoman"].CapitalSettlementID; got != "bursa_city" {
		t.Fatalf("beklenen başkent bursa_city, got=%s", got)
	}
}

func TestAdvanceCapitalMovesCompletesAndCancels(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman", CapitalSettlementID: "ankara_city"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ankara": {
				ID:      "ankara",
				OwnerID: "ottoman",
				Settlements: []world.Settlement{
					{ID: "ankara_city", NameTR: "Ankara", IsCenter: true},
				},
			},
			"bursa": {
				ID:      "bursa",
				OwnerID: "ottoman",
				Settlements: []world.Settlement{
					{ID: "bursa_city", NameTR: "Bursa", IsCenter: true},
				},
			},
		},
	}

	if !gs.StartCapitalMove("ottoman", "bursa_city", 2) {
		t.Fatal("başkent taşıma başlamalıydı")
	}
	updates := gs.AdvanceCapitalMoves()
	if len(updates) != 1 || updates[0].RemainingTurns != 1 {
		t.Fatalf("ilk tur sonrası kalan süre 1 olmalıydı, got=%+v", updates)
	}
	updates = gs.AdvanceCapitalMoves()
	if len(updates) != 1 || !updates[0].Completed {
		t.Fatalf("ikinci tur sonunda taşıma tamamlanmalıydı, got=%+v", updates)
	}
	if got := gs.Factions["ottoman"].CapitalSettlementID; got != "bursa_city" {
		t.Fatalf("başkent bursa_city olmalıydı, got=%s", got)
	}

	gs.Factions["ottoman"].PendingCapitalSettlementID = "ghost_city"
	gs.Factions["ottoman"].PendingCapitalTurns = 2
	updates = gs.AdvanceCapitalMoves()
	if len(updates) != 1 || !updates[0].Cancelled {
		t.Fatalf("geçersiz hedef iptal edilmeliydi, got=%+v", updates)
	}
}
