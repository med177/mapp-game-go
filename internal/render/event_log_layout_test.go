package render

import "testing"

func TestEventLogLayoutKeepsHeaderAndCardsSeparated(t *testing.T) {
	layout := buildEventLogLayout(false)
	if !layout.panel.Rect.Hit(layout.title.X, layout.title.Y) {
		t.Fatalf("başlık panelin dışında: panel=%+v title=%+v", layout.panel.Rect, layout.title)
	}
	if layout.title.X+layout.title.W > layout.codex.X {
		t.Fatalf("başlık alanı Kodex düğmesiyle çakışıyor: title=%+v codex=%+v", layout.title, layout.codex)
	}
	if layout.codex.X+layout.codex.W > layout.toggle.X {
		t.Fatalf("Kodex ve daraltma düğmeleri çakışıyor: codex=%+v toggle=%+v", layout.codex, layout.toggle)
	}

	visible := eventLogVisibleCount()
	if visible < 1 {
		t.Fatalf("olay listesi için görünür satır ayrılmamış: content=%+v", layout.content)
	}
	for i := 0; i < visible; i++ {
		x, y, w, h := eventLogCardRect(i)
		card := struct {
			x, y, w, h float32
		}{x, y, w, h}
		if float64(card.x) < layout.content.X || float64(card.y) < layout.content.Y ||
			float64(card.x+card.w) > layout.content.X+layout.content.W ||
			float64(card.y+card.h) > layout.content.Y+layout.content.H {
			t.Fatalf("kart içerik alanı dışına taşıyor: index=%d card=%+v content=%+v", i, card, layout.content)
		}
		close := buildEventLogCloseButton(i)
		if !layout.panel.Rect.Hit(close.X, close.Y) || !layout.panel.Rect.Hit(close.X+close.W, close.Y+close.H) {
			t.Fatalf("kapatma düğmesi panel dışına taşıyor: index=%d button=%+v", i, close)
		}
	}
}
