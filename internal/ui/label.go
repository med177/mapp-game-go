package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type Label struct {
	X       float64
	Y       float64
	Text    string
	Color   color.Color
	Visible bool
}

func NewLabel(x, y float64, text string, col color.Color) Label {
	return Label{X: x, Y: y, Text: text, Color: col, Visible: true}
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
	text.Draw(screen, l.Text, l.X, l.Y, l.Color)
}
