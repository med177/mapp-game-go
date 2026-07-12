package render

import (
	"image/color"
	"testing"

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
