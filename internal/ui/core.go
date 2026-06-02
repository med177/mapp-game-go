package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// InputState frame başına tek kez toplanan UI girdisini taşır.
type InputState struct {
	MouseX          float64
	MouseY          float64
	LeftJustPressed bool
	WheelY          float64
}

// Widget ortak UI sözleşmesinin çekirdek yüzeyidir.
type Widget interface {
	HitTest(mx, my float64) bool
	HandleInput(input InputState) bool
	Draw(screen *ebiten.Image, text TextRenderer)
}

// TextRenderer UI paketinin render katmanından font bağımsız kalmasını sağlar.
type TextRenderer interface {
	Measure(text string) float64
	Draw(screen *ebiten.Image, text string, x, y float64, col color.Color)
}
