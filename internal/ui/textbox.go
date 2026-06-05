package ui

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type TextBoxStyle struct {
	BG          color.RGBA
	Border      color.RGBA
	Focused     color.RGBA
	Text        color.RGBA
	Placeholder color.RGBA
	BorderWidth float32
	TextOffsetX float64
	TextOffsetY float64
	TextVariant TextVariant
}

type TextBox struct {
	Rect        Rect
	Value       string
	Placeholder string
	Focused     bool
	Enabled     bool
	Visible     bool
	MaxLen      int
}

func NewTextBox(x, y, w, h float64, placeholder string) TextBox {
	return TextBox{
		Rect:        Rect{X: x, Y: y, W: w, H: h},
		Placeholder: placeholder,
		Enabled:     true,
		Visible:     true,
	}
}

func (t TextBox) HitTest(mx, my float64) bool {
	return t.Visible && t.Rect.Hit(mx, my)
}

func (t *TextBox) HandleInput(input InputState) bool {
	if t == nil || !t.Visible || !t.Enabled {
		return false
	}
	if input.LeftJustPressed {
		t.Focused = t.HitTest(input.MouseX, input.MouseY)
		return t.Focused
	}
	if !t.Focused {
		return false
	}
	if input.BackspaceJustPressed && t.Value != "" {
		t.Value = dropLastRune(t.Value)
		return true
	}
	if input.TextInput != "" {
		next := t.Value + strings.ReplaceAll(input.TextInput, "\n", "")
		if t.MaxLen > 0 && len([]rune(next)) > t.MaxLen {
			runes := []rune(next)
			next = string(runes[:t.MaxLen])
		}
		if next != t.Value {
			t.Value = next
			return true
		}
	}
	return input.EnterJustPressed || input.TabJustPressed
}

func (t TextBox) Draw(_ *ebiten.Image, _ TextRenderer) {}

func DrawTextBox(screen *ebiten.Image, t TextBox, style TextBoxStyle, text TextRenderer) {
	if !t.Visible {
		return
	}
	border := style.Border
	if t.Focused {
		border = style.Focused
	}
	vector.FillRect(screen, float32(t.Rect.X), float32(t.Rect.Y), float32(t.Rect.W), float32(t.Rect.H), style.BG, false)
	vector.StrokeRect(screen, float32(t.Rect.X), float32(t.Rect.Y), float32(t.Rect.W), float32(t.Rect.H), style.BorderWidth, border, false)

	label := t.Value
	col := style.Text
	if label == "" {
		label = t.Placeholder
		col = style.Placeholder
	}
	text.Draw(screen, label, t.Rect.X+style.TextOffsetX, t.Rect.Y+style.TextOffsetY, col, style.TextVariant)
}

func dropLastRune(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	return string(runes[:len(runes)-1])
}
