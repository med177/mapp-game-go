package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type ModalStyle struct {
	Overlay color.RGBA
	Panel   PanelStyle
}

type Modal struct {
	OverlayRect         Rect
	Panel               Panel
	Visible             bool
	DismissOnOutsideTap bool
	Children            []Widget
}

func NewModal(screenW, screenH float64, panel Panel) Modal {
	return Modal{
		OverlayRect: Rect{X: 0, Y: 0, W: screenW, H: screenH},
		Panel:       panel,
		Visible:     true,
	}
}

func (m Modal) HitTest(mx, my float64) bool {
	return m.Visible && m.OverlayRect.Hit(mx, my)
}

func (m *Modal) HandleInput(input InputState) bool {
	if m == nil || !m.Visible || !m.HitTest(input.MouseX, input.MouseY) {
		return false
	}
	for i := len(m.Children) - 1; i >= 0; i-- {
		if m.Children[i].HandleInput(input) {
			return true
		}
	}
	if m.DismissOnOutsideTap && input.LeftJustPressed && !m.Panel.HitTest(input.MouseX, input.MouseY) {
		m.Visible = false
		return true
	}
	return input.LeftJustPressed
}

func (m Modal) Draw(_ *ebiten.Image, _ TextRenderer) {}

func DrawModal(screen *ebiten.Image, m Modal, style ModalStyle, text TextRenderer, drawChildren func()) {
	if !m.Visible {
		return
	}
	vector.FillRect(screen, float32(m.OverlayRect.X), float32(m.OverlayRect.Y), float32(m.OverlayRect.W), float32(m.OverlayRect.H), style.Overlay, false)
	DrawPanel(screen, m.Panel, style.Panel)
	if drawChildren != nil {
		drawChildren()
		return
	}
	for _, child := range m.Children {
		child.Draw(screen, text)
	}
}
