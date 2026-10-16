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

type TextVerticalAlign uint8

const (
	TextVerticalAlignTop TextVerticalAlign = iota
	TextVerticalAlignCenter
	TextVerticalAlignBottom
)

type Label struct {
	Rect          Rect
	Text          string
	Color         color.Color
	Variant       TextVariant
	Align         TextAlign
	VerticalAlign TextVerticalAlign
	Visible       bool
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
	y := alignedTextY(text, l.Rect, l.Variant, l.VerticalAlign)
	text.Draw(screen, l.Text, x, y, l.Color, l.Variant)
}

func alignedTextY(text TextRenderer, rect Rect, variant TextVariant, align TextVerticalAlign) float64 {
	y := rect.Y
	if rect.H <= 0 {
		return y
	}
	height := text.MeasureHeight(variant)
	switch align {
	case TextVerticalAlignCenter:
		return rect.Y + (rect.H-height)/2
	case TextVerticalAlignBottom:
		return rect.Y + rect.H - height
	default:
		return y
	}
}
