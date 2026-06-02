package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type ButtonStyle struct {
	BG             color.RGBA
	Border         color.RGBA
	Text           color.RGBA
	DisabledBG     color.RGBA
	DisabledBorder color.RGBA
	DisabledText   color.RGBA
	TextOffsetY    float64
	BorderWidth    float32
}

type Button struct {
	X       float64
	Y       float64
	W       float64
	H       float64
	Label   string
	Enabled bool
}

func NewButton(x, y, w, h float64, label string) Button {
	return Button{X: x, Y: y, W: w, H: h, Label: label, Enabled: true}
}

func (b Button) HitTest(mx, my float64) bool {
	return mx >= b.X && mx <= b.X+b.W && my >= b.Y && my <= b.Y+b.H
}

func (b Button) HandleInput(input InputState) bool {
	return b.Enabled && input.LeftJustPressed && b.HitTest(input.MouseX, input.MouseY)
}

func DrawButton(screen *ebiten.Image, b Button, style ButtonStyle, text TextRenderer) {
	bg := style.BG
	border := style.Border
	txt := style.Text
	if !b.Enabled {
		bg = style.DisabledBG
		border = style.DisabledBorder
		txt = style.DisabledText
	}
	vector.FillRect(screen, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), bg, false)
	vector.StrokeRect(screen, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), style.BorderWidth, border, false)
	tw := text.Measure(b.Label)
	text.Draw(screen, b.Label, b.X+b.W/2-tw/2, b.Y+style.TextOffsetY, txt)
}
