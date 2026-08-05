package render

import (
	"image/color"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func textFace(variant gameui.TextVariant) *text.GoTextFace {
	switch variant {
	case gameui.TextEmphasized:
		return FaceLargeBold
	case gameui.TextLarge:
		return FaceLarge
	case gameui.TextMedium:
		return FaceMed
	default:
		return FaceSmall
	}
}

type sharedTextRenderer struct{}

func (sharedTextRenderer) Measure(text string, variant gameui.TextVariant) float64 {
	return MeasureText(text, textFace(variant))
}

func (sharedTextRenderer) Draw(screen *ebiten.Image, text string, x, y float64, col color.Color, variant gameui.TextVariant) {
	DrawText(screen, text, x, y, textFace(variant), col)
}

var renderText sharedTextRenderer

func drawUIButton(screen *ebiten.Image, x, y, w, h float64, label string, enabled bool, style gameui.ButtonStyle) {
	btn := gameui.NewButton(x, y, w, h, label)
	btn.Enabled = enabled
	gameui.DrawButton(screen, btn, style, renderText)
}

func drawUIButtonWidget(screen *ebiten.Image, btn gameui.Button, style gameui.ButtonStyle) {
	gameui.DrawButton(screen, btn, style, renderText)
}

func drawUIDropdown(screen *ebiten.Image, d *gameui.Dropdown) {
	gameui.DrawDropdown(screen, d, dropdownStyle, renderText)
}
