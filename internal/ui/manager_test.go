package ui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

type focusProbe struct {
	Button
	focusable bool
	focused   bool
}

func newFocusProbe(label string, focusable bool) *focusProbe {
	return &focusProbe{Button: NewButton(0, 0, 10, 10, label), focusable: focusable}
}

func (p *focusProbe) IsFocusable() bool {
	return p.focusable
}

func (p *focusProbe) SetFocused(v bool) {
	p.focused = v
}

func (p *focusProbe) Draw(_ *ebiten.Image, _ TextRenderer) {}

func TestManagerFocusNextSkipsNonFocusableWidgets(t *testing.T) {
	m := NewManager()
	a := newFocusProbe("a", false)
	b := newFocusProbe("b", true)
	c := newFocusProbe("c", true)
	m.Add(a)
	m.Add(b)
	m.Add(c)

	if got := m.FocusNext(); got != 1 {
		t.Fatalf("expected focus index 1, got %d", got)
	}
	if !b.focused || c.focused {
		t.Fatalf("expected only second widget focused")
	}

	if got := m.FocusNext(); got != 2 {
		t.Fatalf("expected focus index 2, got %d", got)
	}
	if b.focused || !c.focused {
		t.Fatalf("expected focus to move from second to third widget")
	}
}

func TestManagerFocusPreviousWraps(t *testing.T) {
	m := NewManager()
	a := newFocusProbe("a", true)
	b := newFocusProbe("b", false)
	c := newFocusProbe("c", true)
	m.Add(a)
	m.Add(b)
	m.Add(c)

	if got := m.FocusPrevious(); got != 2 {
		t.Fatalf("expected focus index 2 after reverse wrap, got %d", got)
	}
	if a.focused || !c.focused {
		t.Fatalf("expected focus to wrap to last focusable widget")
	}
}

func TestManagerResetClearsFocus(t *testing.T) {
	m := NewManager()
	a := newFocusProbe("a", true)
	m.Add(a)
	m.FocusNext()
	m.Reset()

	if got := m.FocusIndex(); got != -1 {
		t.Fatalf("expected no focus after reset, got %d", got)
	}
	if a.focused {
		t.Fatalf("expected reset to clear focused widget")
	}
}

func TestManagerSetFocusIndex(t *testing.T) {
	m := NewManager()
	a := newFocusProbe("a", true)
	b := newFocusProbe("b", true)
	m.Add(a)
	m.Add(b)

	if !m.SetFocusIndex(1) {
		t.Fatalf("expected focus assignment to succeed")
	}
	if got := m.FocusIndex(); got != 1 {
		t.Fatalf("expected focus index 1, got %d", got)
	}
	if a.focused || !b.focused {
		t.Fatalf("expected only second widget focused")
	}
}
