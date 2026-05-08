package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type uiRect [4]float64

func backButtonRect() uiRect {
	return uiRect{24, 18, 92, 30}
}

func drawBackButton(screen *ebiten.Image) {
	drawMenuButton(screen, backButtonRect(), "Geri", color.RGBA{32, 28, 18, 220})
}

func drawMenuButton(screen *ebiten.Image, r uiRect, label string, bg color.RGBA) {
	vector.FillRect(screen, float32(r[0]), float32(r[1]), float32(r[2]), float32(r[3]), bg, false)
	vector.StrokeRect(screen, float32(r[0]), float32(r[1]), float32(r[2]), float32(r[3]), 1, panelBorder, false)
	tw := MeasureText(label, FaceSmall)
	DrawText(screen, label, r[0]+r[2]/2-tw/2, r[1]+8, FaceSmall, ColorGold)
}

func uiRectHit(mx, my float64, r uiRect) bool {
	return mx >= r[0] && mx <= r[0]+r[2] && my >= r[1] && my <= r[1]+r[3]
}
