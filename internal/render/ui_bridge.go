package render

import (
	"image/color"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

type smallTextRenderer struct{}

func (smallTextRenderer) Measure(text string) float64 {
	return MeasureText(text, FaceSmall)
}

func (smallTextRenderer) Draw(screen *ebiten.Image, text string, x, y float64, col color.Color) {
	DrawText(screen, text, x, y, FaceSmall, col)
}

var renderSmallText smallTextRenderer

func drawUIButton(screen *ebiten.Image, x, y, w, h float64, label string, enabled bool, style gameui.ButtonStyle) {
	btn := gameui.NewButton(x, y, w, h, label)
	btn.Enabled = enabled
	gameui.DrawButton(screen, btn, style, renderSmallText)
}

func drawUIDropdown(screen *ebiten.Image, d *gameui.Dropdown) {
	gameui.DrawDropdown(screen, d, dropdownStyle, renderSmallText)
}
