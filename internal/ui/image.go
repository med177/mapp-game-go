package ui

import "github.com/hajimehoshi/ebiten/v2"

type Image struct {
	Rect    Rect
	Src     *ebiten.Image
	Visible bool
}

func NewImage(x, y, w, h float64, src *ebiten.Image) Image {
	return Image{Rect: Rect{X: x, Y: y, W: w, H: h}, Src: src, Visible: true}
}

func (i Image) HitTest(mx, my float64) bool {
	return i.Visible && i.Rect.Hit(mx, my)
}

func (i Image) HandleInput(_ InputState) bool {
	return false
}

func (i Image) Draw(screen *ebiten.Image, _ TextRenderer) {
	if !i.Visible || i.Src == nil {
		return
	}
	sw, sh := i.Src.Bounds().Dx(), i.Src.Bounds().Dy()
	if sw == 0 || sh == 0 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(i.Rect.W/float64(sw), i.Rect.H/float64(sh))
	op.GeoM.Translate(i.Rect.X, i.Rect.Y)
	screen.DrawImage(i.Src, op)
}

type Icon = Image
