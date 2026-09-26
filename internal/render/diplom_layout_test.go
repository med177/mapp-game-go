package render

import (
	"testing"

	gameui "mapp-game-go/internal/ui"
)

func TestDiplomacyOfferLayoutUsesExpandedActionRows(t *testing.T) {
	originalHeight := ScreenHeight
	defer func() { ScreenHeight = originalHeight }()
	ScreenHeight = 720

	layout := diplomacyOfferLayoutForScreen()
	if diplomActionButtonH <= 42 {
		t.Fatalf("diplomasi teklif butonu yüksekliği artmamış: %v", diplomActionButtonH)
	}
	if layout.panelRect.Y < 0 || layout.panelRect.Y+layout.panelRect.H > ScreenHeight {
		t.Fatalf("teklif paneli ekran dışına taşıyor: y=%v h=%v ekran=%v", layout.panelRect.Y, layout.panelRect.H, ScreenHeight)
	}

	footerY := layout.backRect.Y
	for i := range diplomActions {
		_, y, _, h := diplomActionRect(i)
		if float64(h) != diplomActionButtonH {
			t.Fatalf("aksiyon %d yüksekliği = %v, want %v", i, h, diplomActionButtonH)
		}
		if got := float64(y+h) - diplomActionDetailY(y, h); got != diplomActionDetailBottomPadding {
			t.Fatalf("aksiyon %d alt padding = %v, want %v", i, got, diplomActionDetailBottomPadding)
		}
		if float64(y+h) > footerY {
			t.Fatalf("aksiyon %d footer alanına taşıyor: bottom=%v footerY=%v", i, y+h, footerY)
		}
	}
}

func TestDiplomacyListMetricsGiveTreasurySpaceAndEmbedMilitaryRank(t *testing.T) {
	row := gameui.Rect{X: 100, Y: 100, W: 728, H: diplomRowH - 10}
	_, relation, power, treasury := diplomacyListMetricColumnRects(row)

	if treasury.W <= power.W {
		t.Fatalf("hazine kolonu askeri güç kolonundan geniş değil: hazine=%v güç=%v", treasury.W, power.W)
	}
	if relation.W <= 0 || power.W <= 0 || treasury.W <= 0 {
		t.Fatalf("diplomasi metrik kolonlarından biri geçersiz: ilişki=%v güç=%v hazine=%v", relation.W, power.W, treasury.W)
	}
	if relation.X+relation.W > power.X {
		t.Fatalf("ilişki ve askeri güç kolonları çakışıyor: ilişki=%+v güç=%+v", relation, power)
	}
	if power.X+power.W > treasury.X {
		t.Fatalf("askeri güç ve hazine kolonları çakışıyor: güç=%+v hazine=%+v", power, treasury)
	}
	if got := diplomacyMilitaryPowerLabel(253, 35, 1, 10); got != "1. 253/35" {
		t.Fatalf("askeri güç etiketi = %q, want %q", got, "1. 253/35")
	}
}

func TestDiplomacyHistoryResultsStartBelowFilterButtons(t *testing.T) {
	panel := gameui.Rect{X: 100, Y: 200, W: 430, H: diplomHistoryPanelH}
	buttons := buildDiplomacyHistoryFilterButtons(panel, diplomacyHistoryDirectionAll, ActionNone)
	resultsY := diplomacyHistoryResultsY(panel)
	for _, filter := range buttons {
		if resultsY < filter.Button.Y+filter.Button.H {
			t.Fatalf("geçmiş içeriği filtre düğmesine taşıyor: sonuçY=%v düğme=%+v", resultsY, filter.Button)
		}
	}
	for i := 0; i < 3; i++ {
		if buttons[i+3].Button.Y < buttons[i].Button.Y+buttons[i].Button.H {
			t.Fatalf("filtre satırları çakışıyor: üst=%+v alt=%+v", buttons[i].Button, buttons[i+3].Button)
		}
	}
	card := diplomacyOfferHistoryCardRect(panel, 0)
	if card.Y != resultsY {
		t.Fatalf("ilk geçmiş kartı filtre grubunun altından başlamıyor: kartY=%v sonuçY=%v", card.Y, resultsY)
	}
}
