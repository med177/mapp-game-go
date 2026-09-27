package render

import (
	"testing"

	"mapp-game-go/internal/state"
)

func TestOverextensionHUDTextShowsPlayerValue(t *testing.T) {
	gs := &state.GameState{PlayerFactionID: "player"}
	text, _ := overextensionHUDText(gs)
	if text != "Aşırı Genişleme: %0" {
		t.Fatalf("başlangıç aşırı genişleme HUD metni = %q", text)
	}
}
