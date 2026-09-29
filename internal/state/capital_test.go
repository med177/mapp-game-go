package state

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestCapitalBonusDoesNotFollowConqueredRegion(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"old_owner": {ID: "old_owner", CapitalSettlementID: "capital_main"},
			"new_owner": {ID: "new_owner", CapitalSettlementID: "new_capital"},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {
				ID:      "capital",
				OwnerID: "old_owner",
				Settlements: []world.Settlement{
					{ID: "capital_main", Type: world.SettlementCity},
				},
			},
			"new_capital": {
				ID:      "new_capital",
				OwnerID: "new_owner",
				Settlements: []world.Settlement{
					{ID: "new_capital", Type: world.SettlementCity},
				},
			},
		},
	}

	if got := gs.CapitalRegionBonus(gs.Regions["capital"]).Gold; got != CapitalRegionGoldBonus {
		t.Fatalf("sahip başkent bonusu = %d, %d bekleniyordu", got, CapitalRegionGoldBonus)
	}

	gs.Regions["capital"].OwnerID = "new_owner"
	if got := gs.CapitalRegionBonus(gs.Regions["capital"]).Gold; got != 0 {
		t.Fatalf("ele geçirilen eski başkent bonusu taşımaya devam ediyor: %d", got)
	}
}

func TestCapitalMinimumInfrastructureIsFreeOnlyForActiveCapital(t *testing.T) {
	buildingTypes := make(map[string]*city.Building)
	for _, buildingID := range []string{"walls", "port", "barracks", "granary", "temple", "market", "farm"} {
		buildingTypes[buildingID] = &city.Building{ID: buildingID, GoldMaintenance: 1}
	}
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"owner": {ID: "owner", CapitalSettlementID: "capital_main"},
			"other": {ID: "other", CapitalSettlementID: "other_main"},
		},
		BuildingTypes: buildingTypes,
		Regions: map[world.RegionID]*world.Region{
			"capital": {
				ID:      "capital",
				OwnerID: "owner",
				Settlements: []world.Settlement{
					{ID: "capital_main", Type: world.SettlementFortress},
					{ID: "capital_port", Type: world.SettlementPort},
				},
				Buildings: []string{
					"walls", "walls", "port", "port", "barracks", "granary", "temple", "market", "farm",
				},
			},
			"other": {
				ID:        "other",
				OwnerID:   "owner",
				Buildings: []string{"walls", "port", "barracks", "granary", "temple", "market", "farm"},
			},
		},
	}

	// Başkentteki ilk zorunlu seviye ücretsiz, ikinci seviyeler ve farm ücretli:
	// 2 walls + 2 port + 4 capital buildings + farm = 9 toplamdan 6 ücretsiz.
	if got := gs.FactionBuildingGoldUpkeep("owner"); got != 10 {
		t.Fatalf("aktif başkent bakım hesabı = %d, 10 bekleniyordu", got)
	}

	gs.Regions["capital"].OwnerID = "other"
	if got := gs.FactionBuildingGoldUpkeep("owner"); got != 7 {
		t.Fatalf("el değiştiren başkent bakım hesabı = %d, 7 bekleniyordu", got)
	}
	if got := gs.FactionBuildingGoldUpkeep("other"); got != 9 {
		t.Fatalf("yeni sahip başkent altyapısını yanlışlıkla ücretsiz aldı: %d", got)
	}
}

func TestCapitalMoveTransfersBonusFromOldToNewRegion(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"owner": {ID: "owner", CapitalSettlementID: "old_capital"},
		},
		Regions: map[world.RegionID]*world.Region{
			"old_region": {
				ID:      "old_region",
				OwnerID: "owner",
				Settlements: []world.Settlement{
					{ID: "old_capital", Type: world.SettlementCity},
				},
			},
			"new_region": {
				ID:      "new_region",
				OwnerID: "owner",
				Settlements: []world.Settlement{
					{ID: "new_capital", Type: world.SettlementCity},
				},
			},
		},
	}

	if !gs.StartCapitalMove("owner", "new_capital", 1) {
		t.Fatal("başkent taşıma kuyruğu başlatılamadı")
	}
	if got := gs.CapitalRegionBonus(gs.Regions["old_region"]).Gold; got != CapitalRegionGoldBonus {
		t.Fatalf("taşıma tamamlanmadan eski başkent bonusu = %d, %d bekleniyordu", got, CapitalRegionGoldBonus)
	}
	if got := gs.CapitalRegionBonus(gs.Regions["new_region"]).Gold; got != 0 {
		t.Fatalf("taşıma tamamlanmadan yeni bölge bonusu = %d, 0 bekleniyordu", got)
	}

	updates := gs.AdvanceCapitalMoves()
	if len(updates) != 1 || !updates[0].Completed {
		t.Fatalf("başkent taşıması tamamlanma bildirimi = %+v", updates)
	}
	if got := gs.CapitalRegionBonus(gs.Regions["old_region"]).Gold; got != 0 {
		t.Fatalf("taşıma sonrası eski bölge bonusu hâlâ %d", got)
	}
	if got := gs.CapitalRegionBonus(gs.Regions["new_region"]).Gold; got != CapitalRegionGoldBonus {
		t.Fatalf("taşıma sonrası yeni bölge bonusu = %d, %d bekleniyordu", got, CapitalRegionGoldBonus)
	}
}
