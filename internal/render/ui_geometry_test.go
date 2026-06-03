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
		assertButtonInside(t, tc.w, tc.h, buildTopDateHudMenuButton())
		assertButtonInside(t, tc.w, tc.h, buildEventDetailCloseButton())
		assertModalInside(t, tc.w, tc.h, buildConfirmDialogModal())
		assertModalInside(t, tc.w, tc.h, buildWarConfirmModal())
		assertModalInside(t, tc.w, tc.h, buildDiplomacyOfferModal())
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
