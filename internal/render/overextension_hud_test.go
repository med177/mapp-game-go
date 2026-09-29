package render

import (
	"testing"

	"mapp-game-go/internal/state"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestTopAlertHUDRectSitsBelowTopStatusAndStaysThin(t *testing.T) {
	alert := topAlertHudRect()
	wantY := float64(topStatusH + topAlertHUDGap)
	if alert.Y != wantY {
		t.Fatalf("uyarı paneli Y = %.1f, want %.1f", alert.Y, wantY)
	}
	if alert.H != float64(topAlertHUDH) {
		t.Fatalf("uyarı paneli yüksekliği = %.1f, want %.1f", alert.H, topAlertHUDH)
	}
	if alert.W <= 0 || alert.X+alert.W > ScreenWidth {
		t.Fatalf("uyarı paneli ekran dışına taşıyor: %+v, ekran genişliği %.1f", alert, ScreenWidth)
	}
	if !topStatusPanelHit(alert.X+alert.W/2, alert.Y+alert.H/2) {
		t.Fatal("uyarı paneli üst HUD input alanına dahil değil")
	}
	if topStatusPanelHit(alert.X+alert.W/2, float64(topStatusH)+float64(topAlertHUDGap)/2) {
		t.Fatal("üst HUD ile uyarı paneli arasındaki boşluk input alanına dahil edildi")
	}
	if topStatusPanelHit(alert.X+alert.W/2, alert.Y+alert.H+1) {
		t.Fatal("uyarı panelinin altındaki harita alanı üst HUD kabul edildi")
	}
}

func TestOverextensionHUDTextShowsPlayerValue(t *testing.T) {
	gs := &state.GameState{PlayerFactionID: "player"}
	text, _ := overextensionHUDText(gs)
	if text != "Aşırı Genişleme: %0" {
		t.Fatalf("başlangıç aşırı genişleme HUD metni = %q", text)
	}
}

func TestOverextensionHUDUsesPointerAndPopupThresholdLines(t *testing.T) {
	r := &Renderer{gs: &state.GameState{Phase: state.PhasePlayerTurn, PlayerFactionID: "player"}}
	r.rebuildUILayers()
	alert := topAlertHudRect()
	mx := alert.X + alert.W/2
	my := alert.Y + alert.H/2
	if !r.overextensionHUDHovering(mx, my) {
		t.Fatal("Aşırı Genişleme alanı hover olarak algılanmadı")
	}
	if got := r.cursorShapeAt(mx, my); got != ebiten.CursorShapePointer {
		t.Fatalf("Aşırı Genişleme cursor şekli = %v, want pointer", got)
	}
	if got := len(overextensionPopupRangeLines()); got != 4 {
		t.Fatalf("Aşırı Genişleme popup eşik satırı = %d, want 4", got)
	}
}

func TestNearestEventHUDUsesSmallestRemainingTurn(t *testing.T) {
	r := &Renderer{eventCodexEntries: [6][]EventCodexEntry{
		{
			{EventID: "far", Title: "Uzak Event", Status: "Takvim", TurnsUntil: 8},
			{EventID: "near", Title: "Yakın Event", Status: "Kilitli", TurnsUntil: 2},
		},
	}}

	entry, ok := r.nearestEventCodexEntry()
	if !ok || entry.EventID != "near" {
		t.Fatalf("en yakın event = %+v, ok=%v; near event bekleniyordu", entry, ok)
	}
}

func TestNearestEventClickOpensCodexWithMatchingEntryFocused(t *testing.T) {
	r := &Renderer{eventCodexEntries: [6][]EventCodexEntry{
		{
			{EventID: "far", Title: "Aynı Başlık", Status: "Takvim", TurnsUntil: 6},
			{EventID: "near", Title: "Aynı Başlık", Status: "Takvim", TurnsUntil: 1},
		},
	}}

	if !r.openNearestEventCodex() {
		t.Fatal("en yakın event için Event Kodex açılmadı")
	}
	if r.eventCodexFilter != EventCodexAll || r.eventCodexFocus != 1 {
		t.Fatalf("Kodex seçimi = filter=%v focus=%d, all/1 bekleniyordu", r.eventCodexFilter, r.eventCodexFocus)
	}
}

func TestNearestEventHUDKeepsRemainingTurnsInParentheses(t *testing.T) {
	entry := EventCodexEntry{Title: "Bursa'nın Fethi", TurnsUntil: 3}
	text := nearestEventHUDLabel(entry.Title, entry.TurnsUntil)
	if text != "En yakın event: Bursa'nın Fethi (3 tur)" {
		t.Fatalf("event HUD metni = %q", text)
	}
}
