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
	TextVariant  TextVariant
	BorderWidth  float32
	CornerRadius float32
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
	boxX := float32(c.Rect.X)
	boxSize := float32(style.BoxSize)
	if style.CornerRadius > 0 {
		drawCheckboxRoundedRect(screen, boxX, float32(boxY), boxSize, boxSize, style.CornerRadius, style.BoxBorder)
		inset := style.BorderWidth
		innerRadius := style.CornerRadius - inset
		if innerRadius < 0 {
			innerRadius = 0
		}
		drawCheckboxRoundedRect(screen, boxX+inset, float32(boxY)+inset, boxSize-inset*2, boxSize-inset*2, innerRadius, style.BoxBG)
	} else {
		vector.FillRect(screen, boxX, float32(boxY), boxSize, boxSize, style.BoxBG, false)
		vector.StrokeRect(screen, boxX, float32(boxY), boxSize, boxSize, style.BorderWidth, style.BoxBorder, false)
	}
	if c.Checked {
		checkX := boxX + 4
		checkY := float32(boxY + 4)
		checkSize := boxSize - 8
		if style.CornerRadius > 0 {
			drawCheckboxRoundedRect(screen, checkX, checkY, checkSize, checkSize, style.CornerRadius/2, style.CheckColor)
		} else {
			vector.FillRect(screen, checkX, checkY, checkSize, checkSize, style.CheckColor, false)
		}
	}
	col := style.TextColor
	if !c.Enabled {
		col = style.DisabledText
	}
	text.Draw(screen, c.Label, c.Rect.X+style.BoxSize+8, c.Rect.Y+style.TextOffsetY, col, style.TextVariant)
}

func drawCheckboxRoundedRect(screen *ebiten.Image, x, y, w, h, radius float32, col color.RGBA) {
	if radius <= 0 {
		vector.FillRect(screen, x, y, w, h, col, false)
		return
	}
	if radius*2 > w {
		radius = w / 2
	}
	if radius*2 > h {
		radius = h / 2
	}
	vector.FillRect(screen, x+radius, y, w-radius*2, h, col, false)
	vector.FillRect(screen, x, y+radius, w, h-radius*2, col, false)
	vector.FillCircle(screen, x+radius, y+radius, radius, col, false)
	vector.FillCircle(screen, x+w-radius, y+radius, radius, col, false)
	vector.FillCircle(screen, x+radius, y+h-radius, radius, col, false)
	vector.FillCircle(screen, x+w-radius, y+h-radius, radius, col, false)
}
