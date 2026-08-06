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
		},
		Armies: map[army.ArmyID]*army.Army{
			"capital_army": {ID: "capital_army", OwnerID: "player", RegionID: "capital"},
			"field_army":   {ID: "field_army", OwnerID: "player", RegionID: "field"},
		},
	}

	if got, want := gs.ArmyReplenishmentHP(gs.Armies["capital_army"]), 16; got != want {
		t.Fatalf("başkent ordusu normal toparlanmanın iki katını almalıydı, got=%d want=%d", got, want)
	}
	if got, want := gs.ArmyReplenishmentHP(gs.Armies["field_army"]), 8; got != want {
		t.Fatalf("normal bölgenin toparlanma değeri değişmemeli, got=%d want=%d", got, want)
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
