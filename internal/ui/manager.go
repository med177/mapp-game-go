package ui

import "github.com/hajimehoshi/ebiten/v2"

// Manager widget kayıt ve event consume sırasını merkezileştirir.
type Manager struct {
	widgets []Widget
	focus   int
}

func NewManager() *Manager {
	return &Manager{widgets: make([]Widget, 0, 16), focus: -1}
}

func (m *Manager) Reset() {
	m.clearFocus()
	m.widgets = m.widgets[:0]
}

func (m *Manager) Add(w Widget) {
	if w == nil {
		return
	}
	m.widgets = append(m.widgets, w)
}

type Focusable interface {
	Widget
	IsFocusable() bool
	SetFocused(bool)
}

func (m *Manager) FocusIndex() int {
	return m.focus
}

func (m *Manager) SetFocusIndex(idx int) bool {
	if idx < 0 || idx >= len(m.widgets) {
		m.clearFocus()
		return false
	}
	if w, ok := m.widgets[idx].(Focusable); ok && w.IsFocusable() {
		m.setFocus(idx)
		return true
	}
	return false
}

func (m *Manager) FocusNext() int {
	return m.moveFocus(1)
}

func (m *Manager) FocusPrevious() int {
	return m.moveFocus(-1)
}

func (m *Manager) moveFocus(delta int) int {
	if len(m.widgets) == 0 {
		m.focus = -1
		return m.focus
	}
	start := m.focus
	if start < 0 || start >= len(m.widgets) {
		if delta >= 0 {
			start = -1
		} else {
			start = 0
		}
	}
	for step := 0; step < len(m.widgets); step++ {
		next := (start + delta + len(m.widgets)) % len(m.widgets)
		if w, ok := m.widgets[next].(Focusable); ok && w.IsFocusable() {
			m.setFocus(next)
			return m.focus
		}
		start = next
	}
	m.clearFocus()
	return m.focus
}

func (m *Manager) setFocus(idx int) {
	if m.focus == idx {
		return
	}
	if m.focus >= 0 && m.focus < len(m.widgets) {
		if w, ok := m.widgets[m.focus].(Focusable); ok {
			w.SetFocused(false)
		}
	}
	m.focus = idx
	if w, ok := m.widgets[m.focus].(Focusable); ok {
		w.SetFocused(true)
	}
}

func (m *Manager) clearFocus() {
	if m.focus >= 0 && m.focus < len(m.widgets) {
		if w, ok := m.widgets[m.focus].(Focusable); ok {
			w.SetFocused(false)
		}
	}
	m.focus = -1
}

// Dispatch son eklenen widget'tan başlayarak event consume zinciri uygular.
func (m *Manager) Dispatch(input InputState) bool {
	for i := len(m.widgets) - 1; i >= 0; i-- {
		if m.widgets[i].HandleInput(input) {
			return true
		}
	}
	return false
}

func (m *Manager) Draw(screen *ebiten.Image, text TextRenderer) {
	for _, w := range m.widgets {
		w.Draw(screen, text)
	}
}
