package ui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

type consumeWidget struct{ hit bool }

func (w consumeWidget) HitTest(mx, my float64) bool { return w.hit }
func (w consumeWidget) HandleInput(input InputState) bool {
	return w.hit && input.LeftJustPressed
}
func (w consumeWidget) Draw(_ *ebiten.Image, _ TextRenderer) {}

func TestDropdownGetSelectedOptionUsesScrollAndRows(t *testing.T) {
	d := NewDropdown(100, 50, 200, 300, "Test", 30, 24, 3)
	d.SetOptions([]string{"a", "b", "c", "d", "e"}, "a")
	d.Toggle()
	d.Scroll(-1)
	d.Scroll(-1)

	idx, ok := d.GetSelectedOption(120, 50+30+24+2)
	if !ok {
		t.Fatalf("expected row hit")
	}
	if idx != 3 {
		t.Fatalf("expected option index 3, got %d", idx)
	}
}

func TestManagerDispatchConsumesTopMostWidget(t *testing.T) {
	input := InputState{MouseX: 20, MouseY: 20, LeftJustPressed: true}
	m := NewManager()
	lower := consumeWidget{hit: true}
	upper := consumeWidget{hit: true}
	m.Add(lower)
	m.Add(upper)
	if !m.Dispatch(input) {
		t.Fatalf("expected click to be consumed")
	}
}
