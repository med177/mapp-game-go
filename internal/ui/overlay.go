package ui

import "github.com/hajimehoshi/ebiten/v2"

type Overlay struct {
	Rect         Rect
	Visible      bool
	ConsumeInput bool
	DrawFunc     func(screen *ebiten.Image)
}

func NewOverlay(x, y, w, h float64) Overlay {
	return Overlay{
		Rect:    Rect{X: x, Y: y, W: w, H: h},
		Visible: true,
	}
}

func (o Overlay) HitTest(mx, my float64) bool {
	return o.Visible && o.Rect.Hit(mx, my)
}

func (o Overlay) HandleInput(input InputState) bool {
	return o.Visible && o.ConsumeInput && input.LeftJustPressed && o.HitTest(input.MouseX, input.MouseY)
}

func (o Overlay) Draw(screen *ebiten.Image, _ TextRenderer) {
	if !o.Visible || o.DrawFunc == nil {
		return
	}
	o.DrawFunc(screen)
}
