package ui

import "testing"

func TestCheckboxToggle(t *testing.T) {
	c := NewCheckbox(10, 10, 120, 24, "test")
	consumed := c.HandleInput(InputState{MouseX: 20, MouseY: 20, LeftJustPressed: true})
	if !consumed {
		t.Fatalf("checkbox click should be consumed")
	}
	if !c.Checked {
		t.Fatalf("checkbox should toggle on")
	}
}

func TestRadioGroupSelect(t *testing.T) {
	r := NewRadioGroup(0, 0, 200, []string{"a", "b", "c"}, 24)
	consumed := r.HandleInput(InputState{MouseX: 5, MouseY: 30, LeftJustPressed: true})
	if !consumed {
		t.Fatalf("radio click should be consumed")
	}
	if r.Selected != 1 {
		t.Fatalf("expected selected index 1, got %d", r.Selected)
	}
}

func TestListViewSelectionAndScroll(t *testing.T) {
	l := NewListView(0, 0, 200, 100, 20, 4, []string{"a", "b", "c", "d", "e", "f"})
	if !l.HandleInput(InputState{MouseX: 10, MouseY: 10, LeftJustPressed: true}) {
		t.Fatalf("list click should be consumed")
	}
	if l.Selected != 0 {
		t.Fatalf("expected first row selected, got %d", l.Selected)
	}
	if !l.HandleInput(InputState{MouseX: 10, MouseY: 10, WheelY: -1}) {
		t.Fatalf("list wheel should be consumed")
	}
	if l.Scroll != 1 {
		t.Fatalf("expected scroll 1, got %d", l.Scroll)
	}
}

func TestVBoxLayout(t *testing.T) {
	rects := VBox(10, 20, 100, 24, 6, 3)
	if len(rects) != 3 {
		t.Fatalf("expected 3 rects, got %d", len(rects))
	}
	if rects[1].Y != 50 {
		t.Fatalf("expected second rect y=50, got %v", rects[1].Y)
	}
}

func TestModalDismissOnOutsideTap(t *testing.T) {
	m := NewModal(1280, 720, NewPanel(100, 100, 200, 120))
	m.DismissOnOutsideTap = true
	consumed := m.HandleInput(InputState{MouseX: 20, MouseY: 20, LeftJustPressed: true})
	if !consumed {
		t.Fatalf("outside tap should be consumed")
	}
	if m.Visible {
		t.Fatalf("modal should dismiss on outside tap")
	}
}

func TestModalDoesNotDismissOnInsideTap(t *testing.T) {
	m := NewModal(1280, 720, NewPanel(100, 100, 200, 120))
	m.DismissOnOutsideTap = true
	consumed := m.HandleInput(InputState{MouseX: 150, MouseY: 150, LeftJustPressed: true})
	if !consumed {
		t.Fatalf("inside tap should be consumed")
	}
	if !m.Visible {
		t.Fatalf("modal should stay visible on inside tap")
	}
}
