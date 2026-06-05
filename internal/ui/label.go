package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type TextAlign uint8

const (
	TextAlignStart TextAlign = iota
	TextAlignCenter
	TextAlignEnd
)

type Label struct {
	Rect    Rect
	Text    string
	Color   color.Color
	Variant TextVariant
	Align   TextAlign
	Visible bool
}

func NewLabel(x, y float64, text string, col color.Color) Label {
	return Label{
		Rect:    Rect{X: x, Y: y},
		Text:    text,
		Color:   col,
		Variant: TextSmall,
		Align:   TextAlignStart,
		Visible: true,
	}
}

func NewTextLabel(rect Rect, text string, col color.Color, variant TextVariant, align TextAlign) Label {
	return Label{
		Rect:    rect,
		Text:    text,
		Color:   col,
		Variant: variant,
		Align:   align,
		Visible: true,
	}
}

func (l Label) HitTest(_, _ float64) bool {
	return false
}

func (l Label) HandleInput(_ InputState) bool {
	return false
}

func (l Label) Draw(screen *ebiten.Image, text TextRenderer) {
	if !l.Visible {
		return
	}
	x := alignedTextX(text, l.Rect, l.Text, l.Variant, l.Align)
	text.Draw(screen, l.Text, x, l.Rect.Y, l.Color, l.Variant)
}
