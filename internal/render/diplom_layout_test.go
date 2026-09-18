package render

import "testing"

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
		if float64(y+h) > footerY {
			t.Fatalf("aksiyon %d footer alanına taşıyor: bottom=%v footerY=%v", i, y+h, footerY)
		}
	}
}
