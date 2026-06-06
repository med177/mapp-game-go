package render

import (
	"image/color"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

type uiRect [4]float64

func backButtonRect() uiRect {
	return uiRect{24, 18, 92, 30}
}

func drawBackButton(screen *ebiten.Image) {
	drawMenuButton(screen, buildBackButton(), color.RGBA{32, 28, 18, 220})
}

func buildBackButton() gameui.Button {
	r := backButtonRect()
	return gameui.NewButton(r[0], r[1], r[2], r[3], "Geri").WithIcon(gameui.IconBack)
}

func drawMenuButton(screen *ebiten.Image, btn gameui.Button, bg color.RGBA) {
	style := menuButtonStyle
	style.BG = bg
	drawUIButtonWidget(screen, btn, style)
}

func focusButtonIndex(buttons []gameui.Button, current int, backward bool) int {
	mgr := gameui.NewManager()
	for i := range buttons {
		mgr.Add(&buttons[i])
	}
	mgr.SetFocusIndex(current)
	if backward {
		return mgr.FocusPrevious()
	}
	return mgr.FocusNext()
}
