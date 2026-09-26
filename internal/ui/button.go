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
	TextVariant    TextVariant
	BorderWidth    float32
	CornerRadius   float32
}

type Button struct {
	X        float64
	Y        float64
	W        float64
	H        float64
	Label    string
	Icon     IconID
	IconGap  float64
	IconSize float64
	Enabled  bool
	Focused  bool
}

func NewButton(x, y, w, h float64, label string) Button {
	return Button{X: x, Y: y, W: w, H: h, Label: label, Enabled: true}
}

// NewCloseButton panel ve modal başlıklarında kullanılan ortak sağ üst kapatma
// kontrolünü oluşturur. Kutu ölçüsü layout'a göre değişebilir; ikonun oranı ve
// etkileşim davranışı ortak kalır.
func NewCloseButton(x, y, w, h float64) Button {
	button := NewButton(x, y, w, h, "").WithIcon(IconClose)
	size := w
	if h < size {
		size = h
	}
	button.IconSize = size * 0.5
	return button
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
	drawButtonBackground(screen, b, style.CornerRadius, bg, border, style.BorderWidth)
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

func drawButtonBackground(screen *ebiten.Image, b Button, radius float32, bg, border color.RGBA, borderWidth float32) {
	if radius <= 0 {
		vector.FillRect(screen, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), bg, false)
		drawButtonBorder(screen, b.X, b.Y, b.W, b.H, 0, float64(borderWidth), border, bg)
		return
	}
	drawButtonRoundedRect(screen, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), radius, bg)
	drawButtonBorder(screen, b.X, b.Y, b.W, b.H, float64(radius), float64(borderWidth), border, bg)
}

func drawButtonBorder(screen *ebiten.Image, x, y, w, h, radius, width float64, border, fill color.RGBA) {
	if radius <= 0 {
		vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), float32(width), border, false)
		return
	}
	drawButtonRoundedRect(screen, float32(x), float32(y), float32(w), float32(h), float32(radius), border)
	inset := width
	innerRadius := radius - inset
	if innerRadius < 0 {
		innerRadius = 0
	}
	drawButtonRoundedRect(screen, float32(x+inset), float32(y+inset), float32(w-inset*2), float32(h-inset*2), float32(innerRadius), fill)
}

func drawButtonRoundedRect(screen *ebiten.Image, x, y, w, h, radius float32, col color.RGBA) {
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
	// Button etiketleri, ikonlar gibi her zaman buton kutusunun dikey merkezine
	// göre çizilir.
	textH := buttonTextHeight(style.TextVariant)
	return b.Y + (b.H-textH)/2
}

func buttonTextHeight(variant TextVariant) float64 {
	switch variant {
	case TextEmphasized:
		return 18
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
