package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type CheckboxStyle struct {
	BoxBG        color.RGBA
	BoxBorder    color.RGBA
	CheckColor   color.RGBA
	TextColor    color.RGBA
	DisabledText color.RGBA
	BoxSize      float64
	TextOffsetY  float64
	BorderWidth  float32
}

type Checkbox struct {
	Rect    Rect
	Label   string
	Checked bool
	Enabled bool
}

func NewCheckbox(x, y, w, h float64, label string) Checkbox {
	return Checkbox{
		Rect:    Rect{X: x, Y: y, W: w, H: h},
		Label:   label,
		Enabled: true,
	}
}

func (c Checkbox) HitTest(mx, my float64) bool {
	return c.Enabled && c.Rect.Hit(mx, my)
}

func (c *Checkbox) HandleInput(input InputState) bool {
	if c == nil || !c.HitTest(input.MouseX, input.MouseY) || !input.LeftJustPressed {
		return false
	}
	c.Checked = !c.Checked
	return true
}

func (c Checkbox) Draw(_ *ebiten.Image, _ TextRenderer) {}

func DrawCheckbox(screen *ebiten.Image, c Checkbox, style CheckboxStyle, text TextRenderer) {
	boxY := c.Rect.Y + (c.Rect.H-style.BoxSize)/2
	vector.FillRect(screen, float32(c.Rect.X), float32(boxY), float32(style.BoxSize), float32(style.BoxSize), style.BoxBG, false)
	vector.StrokeRect(screen, float32(c.Rect.X), float32(boxY), float32(style.BoxSize), float32(style.BoxSize), style.BorderWidth, style.BoxBorder, false)
	if c.Checked {
		vector.FillRect(screen, float32(c.Rect.X+4), float32(boxY+4), float32(style.BoxSize-8), float32(style.BoxSize-8), style.CheckColor, false)
	}
	col := style.TextColor
	if !c.Enabled {
		col = style.DisabledText
	}
	text.Draw(screen, c.Label, c.Rect.X+style.BoxSize+8, c.Rect.Y+style.TextOffsetY, col)
}
