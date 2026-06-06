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
	TextVariant    TextVariant
	BorderWidth    float32
}

type Button struct {
	X       float64
	Y       float64
	W       float64
	H       float64
	Label   string
	Icon    IconID
	IconGap float64
	IconSize float64
	Enabled bool
	Focused bool
}

func NewButton(x, y, w, h float64, label string) Button {
	return Button{X: x, Y: y, W: w, H: h, Label: label, Enabled: true}
}

func (b Button) WithIcon(icon IconID) Button {
	b.Icon = icon
	return b
}

func (b Button) HitTest(mx, my float64) bool {
	return mx >= b.X && mx <= b.X+b.W && my >= b.Y && my <= b.Y+b.H
}

func (b Button) HandleInput(input InputState) bool {
	return b.Enabled && input.LeftJustPressed && b.HitTest(input.MouseX, input.MouseY)
}

func (b Button) Draw(_ *ebiten.Image, _ TextRenderer) {}

func (b *Button) IsFocusable() bool {
	return b != nil && b.Enabled
}

func (b *Button) SetFocused(v bool) {
	if b == nil {
		return
	}
	b.Focused = v
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
	if b.Focused && b.Enabled {
		vector.StrokeRect(screen, float32(b.X+2), float32(b.Y+2), float32(b.W-4), float32(b.H-4), 1, txt, false)
	}
	tw := text.Measure(b.Label, style.TextVariant)
	iconSize := buttonIconSize(b)
	iconGap := buttonIconGap(b)
	hasIcon := b.Icon != IconNone && iconImage(b.Icon) != nil
	if b.Label == "" {
		iconGap = 0
	}
	contentW := tw
	if hasIcon {
		contentW += iconSize + iconGap
	}
	contentX := b.X + (b.W-contentW)/2
	if hasIcon {
		drawButtonIcon(screen, b.Icon, contentX, b.Y+(b.H-iconSize)/2, iconSize, txt)
		contentX += iconSize + iconGap
	}
	if b.Label != "" {
		text.Draw(screen, b.Label, contentX, buttonTextY(b, style), txt, style.TextVariant)
	}
}

func buttonIconGap(b Button) float64 {
	if b.IconGap > 0 {
		return b.IconGap
	}
	return 6
}

func buttonIconSize(b Button) float64 {
	if b.IconSize > 0 {
		return b.IconSize
	}
	scale := 0.68
	switch {
	case b.H <= 24:
		scale = 0.82
	case b.H <= 32:
		scale = 0.76
	}
	size := b.H * scale
	maxSize := b.H - 4
	if size > maxSize {
		size = maxSize
	}
	if size < 10 {
		size = 10
	}
	return size
}

func buttonTextY(b Button, style ButtonStyle) float64 {
	textH := buttonTextHeight(style.TextVariant)
	return b.Y + (b.H-textH)/2
}

func buttonTextHeight(variant TextVariant) float64 {
	switch variant {
	case TextLarge:
		return 18
	case TextMedium:
		return 14
	default:
		return 12
	}
}

func drawButtonIcon(screen *ebiten.Image, icon IconID, x, y, size float64, tint color.Color) {
	src := iconImage(icon)
	if src == nil {
		return
	}
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	if sw == 0 || sh == 0 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(size/float64(sw), size/float64(sh))
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(tint)
	screen.DrawImage(src, op)
}
