package render

import (
	"testing"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestCoreUIGeometryFitsCommonViewports(t *testing.T) {
	cases := []struct {
		w float64
		h float64
	}{
		{1280, 720},
		{1600, 900},
		{1920, 1080},
	}
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()

	for _, tc := range cases {
		ScreenWidth = tc.w
		ScreenHeight = tc.h
		for _, btn := range buildBottomActionButtons(true) {
			assertButtonInside(t, tc.w, tc.h, btn)
		}
		for _, btn := range buildMapModeButtons() {
			assertButtonInside(t, tc.w, tc.h, btn)
		}
		assertButtonInside(t, tc.w, tc.h, buildTradeToggleButton())
		assertTradePanelInside(t, tc.w, tc.h)
		assertButtonInside(t, tc.w, tc.h, buildTopDateHudMenuButton())
		assertButtonInside(t, tc.w, tc.h, buildEventDetailCloseButton())
		assertModalInside(t, tc.w, tc.h, buildConfirmDialogModal())
		assertModalInside(t, tc.w, tc.h, buildWarConfirmModal())
		assertModalInside(t, tc.w, tc.h, buildDiplomacyOfferModal())
	}
}

func assertTradePanelInside(t *testing.T, screenW, screenH float64) {
	t.Helper()
	layout := tradePanelLayout()
	const eps = 0.5
	if layout.panelRect.X < 0 || layout.panelRect.Y < 0 || layout.panelRect.X+layout.panelRect.W > screenW || layout.panelRect.Y+layout.panelRect.H > screenH {
		t.Fatalf("trade panel outside %.0fx%.0f viewport: %+v", screenW, screenH, layout)
	}
	if layout.rightListRect.X+layout.rightListRect.W > layout.panelRect.X+layout.panelRect.W-8+eps {
		t.Fatalf("trade panel right column overflow in %.0fx%.0f viewport: %+v", screenW, screenH, layout)
	}
	cardX, cardY, cardW, cardH := tradeActionCardRect(layout, len(tradeSelectableGoods()))
	if float64(cardX) < layout.rightListRect.X-eps || float64(cardX+cardW) > layout.panelRect.X+layout.panelRect.W-8+eps {
		t.Fatalf("trade action card overflow in %.0fx%.0f viewport: layout=%+v card=(%.1f,%.1f,%.1f,%.1f)", screenW, screenH, layout, cardX, cardY, cardW, cardH)
	}
	if float64(cardY) < layout.rightListRect.Y+float64(tradeGoodsListHeight(len(tradeSelectableGoods())))-eps {
		t.Fatalf("trade action card overlaps goods list in %.0fx%.0f viewport: layout=%+v card=(%.1f,%.1f,%.1f,%.1f)", screenW, screenH, layout, cardX, cardY, cardW, cardH)
	}
	for _, btn := range buildTradeTabButtons() {
		assertButtonInside(t, screenW, screenH, btn.Button)
	}
	for _, btn := range buildTradeFilterButtons(layout) {
		assertButtonInside(t, screenW, screenH, btn.Button)
	}
	for _, btn := range buildTradeSortButtons(layout) {
		assertButtonInside(t, screenW, screenH, btn.Button)
	}
	qtyButtons, buyBtn, sellBtn := buildTradeActionButtons(layout, len(tradeSelectableGoods()))
	for _, btn := range qtyButtons {
		assertButtonInside(t, screenW, screenH, btn)
	}
	assertButtonInside(t, screenW, screenH, buyBtn)
	assertButtonInside(t, screenW, screenH, sellBtn)
}

func TestTradeFilterPredicates(t *testing.T) {
	if !isTradeSellerForPlayer(1) {
		t.Fatal("stocku olan fraksiyon seller sayilmali")
	}
	if isTradeSellerForPlayer(0) {
		t.Fatal("stogu olmayan fraksiyon seller sayilmamali")
	}
	if !isTradeBuyerForPlayer(20, 5, 2, 50) {
		t.Fatal("oyuncudan daha az stogu olan ve odeyebilen fraksiyon buyer sayilmali")
	}
	if isTradeBuyerForPlayer(20, 25, 2, 50) {
		t.Fatal("oyuncudan fazla stogu olan fraksiyon buyer sayilmamali")
	}
	if isTradeBuyerForPlayer(0, 0, 2, 50) {
		t.Fatal("oyuncuda mal yoksa buyer olusmamali")
	}
}

func TestMainMenuRenderSmokeCommonViewports(t *testing.T) {
	cases := []struct {
		w int
		h int
	}{
		{1280, 720},
		{1600, 900},
		{1920, 1080},
	}
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()

	for _, tc := range cases {
		ScreenWidth = float64(tc.w)
		ScreenHeight = float64(tc.h)
		screen := ebiten.NewImage(tc.w, tc.h)
		DrawMainMenu(screen, 0, true, true, 0)
	}
}

func assertButtonInside(t *testing.T, screenW, screenH float64, btn gameui.Button) {
	t.Helper()
	if btn.X < 0 || btn.Y < 0 || btn.X+btn.W > screenW || btn.Y+btn.H > screenH {
		t.Fatalf("button %q outside %.0fx%.0f viewport: %+v", btn.Label, screenW, screenH, btn)
	}
}

func assertModalInside(t *testing.T, screenW, screenH float64, modal gameui.Modal) {
	t.Helper()
	p := modal.Panel.Rect
	if p.X < 0 || p.Y < 0 || p.X+p.W > screenW || p.Y+p.H > screenH {
		t.Fatalf("modal panel outside %.0fx%.0f viewport: %+v", screenW, screenH, p)
	}
}
