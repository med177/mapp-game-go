package render

import "testing"

func TestOverlayPanelOrderBringsLastOpenedPanelToFront(t *testing.T) {
	r := &Renderer{}
	r.ensureOverlayPanelOrder()
	if got := r.overlayPanelOrder[r.overlayPanelOrderLen-1]; got != overlayPanelActiveWars {
		t.Fatalf("varsayılan en üst panel aktif savaşlar olmalı: got=%d", got)
	}

	r.bringOverlayPanelToFront(overlayPanelMerchantRoute)
	if got := r.overlayPanelOrder[r.overlayPanelOrderLen-1]; got != overlayPanelMerchantRoute {
		t.Fatalf("merchant paneli açılınca en üste taşınmalı: got=%d", got)
	}
	r.bringOverlayPanelToFront(overlayPanelNavalMission)
	if got := r.overlayPanelOrder[r.overlayPanelOrderLen-1]; got != overlayPanelNavalMission {
		t.Fatalf("donanma paneli son açılınca en üste taşınmalı: got=%d", got)
	}
	r.bringOverlayPanelToFront(overlayPanelActiveWars)
	if got := r.overlayPanelOrder[r.overlayPanelOrderLen-1]; got != overlayPanelActiveWars {
		t.Fatalf("aktif savaşlar paneli tekrar açılınca en üste taşınmalı: got=%d", got)
	}
}

func TestOverlayPanelCursorUsesFrontPanelSurface(t *testing.T) {
	r := &Renderer{
		showMerchantRoutePanel: true,
		showNavalMissionPanel:  true,
	}
	r.bringOverlayPanelToFront(overlayPanelMerchantRoute)

	navalLayout := navalMissionPanelLayoutFor(1)
	merchantLayout := merchantRoutePanelLayoutFor(1)
	if merchantLayout.panelX+merchantLayout.panelW <= navalLayout.panelX || navalLayout.panelX+navalLayout.panelW <= merchantLayout.panelX ||
		merchantLayout.panelY+merchantLayout.panelH <= navalLayout.panelY || navalLayout.panelY+navalLayout.panelH <= merchantLayout.panelY {
		t.Fatal("test panelleri en az kısmen üst üste gelmeli")
	}

	// Merchant son açıldığı için aynı koordinatta donanma satırının hover'ı
	// görünür kalmamalı; merchant satırı pointer üretmeli.
	x := float64(merchantLayout.rowX + 20)
	y := float64(merchantLayout.rowY + 5 + 20)
	pointer, handled := r.overlayPanelCursorHit(x, y)
	if !handled || !pointer {
		t.Fatalf("öndeki merchant satırı pointer üretmeli: pointer=%t handled=%t", pointer, handled)
	}
}
