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
	drawMenuButton(screen, backButtonRect(), "Geri", color.RGBA{32, 28, 18, 220})
}

func buildBackButton() gameui.Button {
	r := backButtonRect()
	return gameui.NewButton(r[0], r[1], r[2], r[3], "Geri")
}

func drawMenuButton(screen *ebiten.Image, r uiRect, label string, bg color.RGBA) {
	style := menuButtonStyle
	style.BG = bg
	drawUIButton(screen, r[0], r[1], r[2], r[3], label, true, style)
}

func uiRectHit(mx, my float64, r uiRect) bool {
	return mx >= r[0] && mx <= r[0]+r[2] && my >= r[1] && my <= r[1]+r[3]
}
