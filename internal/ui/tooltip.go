package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type TooltipStyle struct {
	Panel   PanelStyle
	Text    color.RGBA
	Padding float64
	LineH   float64
}

type Tooltip struct {
	Rect    Rect
	Lines   []string
	Visible bool
}

func NewTooltip(x, y, w float64, lines []string, lineH, padding float64) Tooltip {
	h := padding*2 + float64(len(lines))*lineH
	return Tooltip{
		Rect:    Rect{X: x, Y: y, W: w, H: h},
		Lines:   append([]string(nil), lines...),
		Visible: true,
	}
}

func (t Tooltip) HitTest(mx, my float64) bool {
	return t.Visible && t.Rect.Hit(mx, my)
}

func (t Tooltip) HandleInput(_ InputState) bool {
	return false
}

func (t Tooltip) Draw(_ *ebiten.Image, _ TextRenderer) {}

func DrawTooltip(screen *ebiten.Image, t Tooltip, style TooltipStyle, text TextRenderer) {
	if !t.Visible {
		return
	}
	DrawPanel(screen, NewPanel(t.Rect.X, t.Rect.Y, t.Rect.W, t.Rect.H), style.Panel)
	for i, line := range t.Lines {
		text.Draw(screen, line, t.Rect.X+style.Padding, t.Rect.Y+style.Padding+float64(i)*style.LineH, style.Text)
	}
}
