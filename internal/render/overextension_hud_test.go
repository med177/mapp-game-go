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

func TestOverextensionHUDUsesPointerAndEventPopup(t *testing.T) {
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
	r.eventCodexEntries = [6][]EventCodexEntry{{
		{EventID: "event-1", Title: "Bursa'nın Fethi", Status: "Takvim", DateLabel: "1326/04", TurnsUntil: 3, ConditionSummary: "Takvim bekleniyor", EffectSummary: "Etki: +100 altın"},
	}}
	if !r.nearestEventHUDHovering(mx, my) {
		t.Fatal("yaklaşan event alanı hover olarak algılanmadı")
	}
}

func TestEventAlertPopupUsesDateConditionsAndEffects(t *testing.T) {
	event := EventCodexEntry{
		Title:            "Bursa'nın Fethi",
		DateLabel:        "1326/04",
		TurnsUntil:       3,
		ConditionSummary: "Bursa sahibi olmalı",
		EffectSummary:    "Etki: +100 altın",
	}
	lines := eventAlertPopupLines(event)
	if lines[1] != "Tarih: 1326/04 • Kalan: 3 tur" {
		t.Fatalf("event popup tarih satırı = %q", lines[1])
	}
	if lines[2] != "Şartlar: Bursa sahibi olmalı" {
		t.Fatalf("event popup şart satırı = %q", lines[2])
	}
	if lines[3] != "Getiriler: +100 altın" {
		t.Fatalf("event popup getiri satırı = %q", lines[3])
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
