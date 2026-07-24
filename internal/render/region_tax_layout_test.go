package render

import (
	"testing"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestRegionTaxButtonsUseCompactAlignedRects(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"bursa": {ID: "bursa", OwnerID: "osm"},
		},
	}

	dec, inc := regionTaxButtonRects(gs, "bursa")
	contentRight := infoPanelX() + infoPanelW - float32(panelPad)

	if dec[2] != regionPanelTaxButtonW || inc[2] != regionPanelTaxButtonW {
		t.Fatalf("vergi butonları kompakt genişlikte olmalı: dec=%v inc=%v", dec[2], inc[2])
	}
	if dec[3] != regionPanelTaxButtonH || inc[3] != regionPanelTaxButtonH {
		t.Fatalf("vergi butonları kompakt yükseklikte olmalı: dec=%v inc=%v", dec[3], inc[3])
	}
	if inc[0]+inc[2] != contentRight {
		t.Fatalf("artı butonu içerik sağ kenarına hizalanmalı: got=%v want=%v", inc[0]+inc[2], contentRight)
	}
	if dec[0]+dec[2]+regionPanelTaxButtonGap != inc[0] {
		t.Fatalf("vergi butonları arasındaki boşluk yanlış: got=%v want=%v", inc[0]-(dec[0]+dec[2]), regionPanelTaxButtonGap)
	}
	if dec[1] != inc[1] {
		t.Fatalf("vergi butonları aynı düşey hizada olmalı: dec=%v inc=%v", dec[1], inc[1])
	}
}

func TestRegionTaxInteractiveBarStopsBeforeDecreaseButton(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"bursa": {ID: "bursa", OwnerID: "osm"},
		},
	}
	dec, _ := regionTaxButtonRects(gs, "bursa")
	barX, barW := regionPanelTaxInteractiveBarLayout(infoPanelX()+float32(panelPad), infoPanelW-float32(panelPad*2), dec[0])

	if barW <= 0 {
		t.Fatalf("etkileşimli vergi barı görünür genişlikte olmalı: got=%v", barW)
	}
	baseBarX, _ := regionPanelTaxBarLayout(infoPanelX()+float32(panelPad), infoPanelW-float32(panelPad*2))
	if barX-baseBarX != regionPanelTaxButtonPad {
		t.Fatalf("vergi barı başlangıçta buton genişliği kadar içeri alınmalı: got=%v want=%v", barX-baseBarX, regionPanelTaxButtonPad)
	}
	if barX+barW+regionPanelTaxButtonPad != dec[0] {
		t.Fatalf("vergi barı bitişte de buton genişliği kadar mesafede olmalı: got=%v want=%v", barX+barW+regionPanelTaxButtonPad, dec[0])
	}
}
