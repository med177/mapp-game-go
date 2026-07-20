package render

import (
	"image/color"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestDiplomacyOfferQuotaHUDText(t *testing.T) {
	gs := &state.GameState{PlayerFactionID: "player"}

	text, col := diplomacyOfferQuotaHUDText(gs)
	if text != "Elçi 3/3" {
		t.Fatalf("başlangıçta tam hak görünmeliydi, got=%q", text)
	}
	if col != (color.RGBA{232, 190, 100, 255}) {
		t.Fatalf("tam hak rengi farklıydı, got=%v", col)
	}

	gs.DiplomacyOfferCounts = map[faction.FactionID]int{"player": 2}
	text, col = diplomacyOfferQuotaHUDText(gs)
	if text != "Elçi 1/3" {
		t.Fatalf("iki teklif sonrası kalan hak 1/3 görünmeliydi, got=%q", text)
	}
	if col != (color.RGBA{220, 130, 60, 255}) {
		t.Fatalf("tek hak rengi farklıydı, got=%v", col)
	}

	gs.DiplomacyOfferCounts["player"] = 3
	text, col = diplomacyOfferQuotaHUDText(gs)
	if text != "Elçi 0/3" {
		t.Fatalf("hak bitince 0/3 görünmeliydi, got=%q", text)
	}
	if col != (color.RGBA{220, 90, 90, 255}) {
		t.Fatalf("hak bitti rengi farklıydı, got=%v", col)
	}
}

func TestPlayerMilitaryPowerStandingRanksActiveFactions(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"strong": {ID: "strong"},
			"weak":   {ID: "weak"},
			"dead":   {ID: "dead", IsEliminated: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player-army": {ID: "player-army", OwnerID: "player", Units: []army.Unit{{}, {}}},
			"strong-army": {ID: "strong-army", OwnerID: "strong", Units: []army.Unit{{}, {}, {}}},
			"dead-army":   {ID: "dead-army", OwnerID: "dead", Units: []army.Unit{{}, {}, {}, {}}},
		},
	}

	power, rank, count := playerMilitaryPowerStanding(gs)
	if power != 20 || rank != 2 || count != 3 {
		t.Fatalf("oyuncu askeri standing yanlis: power=%d rank=%d count=%d", power, rank, count)
	}
}

func TestPlayerMilitaryPowerStandingUsesFactionIDForTies(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "z-player",
		Factions: map[faction.FactionID]*faction.Faction{
			"a-state":  {ID: "a-state"},
			"z-player": {ID: "z-player"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player-army": {ID: "player-army", OwnerID: "z-player", Units: []army.Unit{{}}},
			"other-army":  {ID: "other-army", OwnerID: "a-state", Units: []army.Unit{{}}},
		},
	}

	_, rank, count := playerMilitaryPowerStanding(gs)
	if rank != 2 || count != 2 {
		t.Fatalf("esit guc tie-break sirasi yanlis: rank=%d count=%d", rank, count)
	}
}
