package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type RadioGroupStyle struct {
	OuterColor   color.RGBA
	InnerColor   color.RGBA
	TextColor    color.RGBA
	DisabledText color.RGBA
	CircleSize   float64
	RowHeight    float64
	TextOffsetY  float64
	BorderWidth  float32
}

type RadioGroup struct {
	Rect     Rect
	Options  []string
	Selected int
	Enabled  bool
}

func NewRadioGroup(x, y, w float64, options []string, rowHeight float64) RadioGroup {
	return RadioGroup{
		Rect:     Rect{X: x, Y: y, W: w, H: rowHeight * float64(len(options))},
		Options:  append([]string(nil), options...),
		Selected: -1,
		Enabled:  true,
	}
}

func (r RadioGroup) HitTest(mx, my float64) bool {
	return r.Enabled && r.Rect.Hit(mx, my)
}

func (r *RadioGroup) HandleInput(input InputState) bool {
	if r == nil || !r.HitTest(input.MouseX, input.MouseY) || !input.LeftJustPressed || len(r.Options) == 0 {
		return false
	}
	rowH := r.Rect.H / float64(len(r.Options))
	idx := int((input.MouseY - r.Rect.Y) / rowH)
	if idx < 0 || idx >= len(r.Options) {
		return false
	}
	r.Selected = idx
	return true
}

func (r RadioGroup) Draw(_ *ebiten.Image, _ TextRenderer) {}

func DrawRadioGroup(screen *ebiten.Image, r RadioGroup, style RadioGroupStyle, text TextRenderer) {
	if len(r.Options) == 0 {
		return
	}
	rowH := style.RowHeight
	if rowH <= 0 {
		rowH = r.Rect.H / float64(len(r.Options))
	}
	for i, option := range r.Options {
		cy := r.Rect.Y + float64(i)*rowH + (rowH-style.CircleSize)/2
		vector.StrokeRect(screen, float32(r.Rect.X), float32(cy), float32(style.CircleSize), float32(style.CircleSize), style.BorderWidth, style.OuterColor, false)
		if i == r.Selected {
			vector.FillRect(screen, float32(r.Rect.X+4), float32(cy+4), float32(style.CircleSize-8), float32(style.CircleSize-8), style.InnerColor, false)
		}
		col := style.TextColor
		if !r.Enabled {
			col = style.DisabledText
		}
		text.Draw(screen, option, r.Rect.X+style.CircleSize+8, r.Rect.Y+float64(i)*rowH+style.TextOffsetY, col)
	}
}
