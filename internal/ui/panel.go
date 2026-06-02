package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type PanelStyle struct {
	BG          color.RGBA
	Border      color.RGBA
	BorderWidth float32
}

type Panel struct {
	Rect    Rect
	Visible bool
}

func NewPanel(x, y, w, h float64) Panel {
	return Panel{Rect: Rect{X: x, Y: y, W: w, H: h}, Visible: true}
}

func (p Panel) HitTest(mx, my float64) bool {
	return p.Visible && p.Rect.Hit(mx, my)
}

func (p Panel) HandleInput(_ InputState) bool {
	return false
}

func (p Panel) Draw(_ *ebiten.Image, _ TextRenderer) {}

func DrawPanel(screen *ebiten.Image, p Panel, style PanelStyle) {
	if !p.Visible {
		return
	}
	vector.FillRect(screen, float32(p.Rect.X), float32(p.Rect.Y), float32(p.Rect.W), float32(p.Rect.H), style.BG, false)
	vector.StrokeRect(screen, float32(p.Rect.X), float32(p.Rect.Y), float32(p.Rect.W), float32(p.Rect.H), style.BorderWidth, style.Border, false)
}
