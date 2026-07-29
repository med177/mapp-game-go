package game

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

func TestApplyConquestWithCapitalCaptureTransfersLootTechAndReassigns(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"attacker": {
				ID:    "attacker",
				Gold:  100,
				Grain: 40,
				Iron:  5,
				Research: faction.ResearchState{
					Completed: map[string]bool{"basic": true},
				},
			},
			"defender": {
				ID:                  "defender",
				Gold:                200,
				Grain:               80,
				Iron:                20,
				Timber:              10,
				Stone:               30,
				Spice:               6,
				Cloth:               12,
				CapitalSettlementID: "capital_city",
				Research: faction.ResearchState{
					Completed: map[string]bool{
						"basic":    true,
						"advanced": true,
						"elite":    true,
					},
				},
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital_region": {
				ID:      "capital_region",
				OwnerID: "defender",
				Settlements: []world.Settlement{
					{ID: "capital_city", NameTR: "Payitaht", IsCenter: true},
				},
			},
			"fallback": {
				ID:             "fallback",
				OwnerID:        "defender",
				BaseGoldIncome: 90,
				TradeCapacity:  6,
				Settlements: []world.Settlement{
					{ID: "fallback_city", NameTR: "Bursa", IsCenter: true},
				},
			},
		},
		TechTypes: map[string]*tech.Technology{
			"basic":    {ID: "basic", TurnsRequired: 1, GoldCost: 50},
			"advanced": {ID: "advanced", TurnsRequired: 5, GoldCost: 200},
			"elite":    {ID: "elite", TurnsRequired: 3, GoldCost: 120},
		},
	}
	g := &Game{gs: gs}

	g.applyConquestWithNavalEviction(gs.Regions["capital_region"], "attacker")

	if gs.Factions["attacker"].Gold != 200 {
		t.Fatalf("altın loot transferi hatalı, got=%d", gs.Factions["attacker"].Gold)
	}
	if gs.Factions["defender"].Gold != 100 {
		t.Fatalf("savunan altını yarıya düşmeliydi, got=%d", gs.Factions["defender"].Gold)
	}
	if gs.Factions["attacker"].Grain != 80 || gs.Factions["attacker"].Iron != 15 {
		t.Fatalf("hammadde transferi eksik, attacker=%+v", gs.Factions["attacker"])
	}
	if !gs.Factions["attacker"].Research.Completed["advanced"] || gs.Factions["attacker"].Research.Completed["elite"] {
		t.Fatalf("yarım teknoloji transferi bekleniyordu, attacker tech=%+v", gs.Factions["attacker"].Research.Completed)
	}
	if got := gs.Factions["defender"].CapitalSettlementID; got != "fallback_city" {
		t.Fatalf("otomatik yeni başkent fallback_city olmalıydı, got=%s", got)
	}
}
