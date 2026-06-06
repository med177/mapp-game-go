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
	if !l.HandleInput(InputState{MouseX: 10, MouseY: 10, LeftJustPressed: true, LeftPressed: true}) {
		t.Fatalf("list click should be consumed")
	}
	if l.Selected != -1 {
		t.Fatalf("selection should wait for release, got %d", l.Selected)
	}
	if !l.HandleInput(InputState{MouseX: 10, MouseY: 10, LeftJustReleased: true}) {
		t.Fatalf("list release should be consumed")
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

func TestListViewDragScrollDoesNotSelectItem(t *testing.T) {
	l := NewListView(0, 0, 200, 80, 20, 4, []string{"a", "b", "c", "d", "e", "f", "g"})
	if !l.HandleInput(InputState{MouseX: 20, MouseY: 30, LeftJustPressed: true, LeftPressed: true}) {
		t.Fatalf("drag start should be consumed")
	}
	if !l.HandleInput(InputState{MouseX: 20, MouseY: 5, LeftPressed: true}) {
		t.Fatalf("drag move should be consumed")
	}
	if l.Scroll != 1 {
		t.Fatalf("expected drag to scroll list by one row, got %d", l.Scroll)
	}
	if !l.HandleInput(InputState{MouseX: 20, MouseY: 5, LeftJustReleased: true}) {
		t.Fatalf("drag release should be consumed")
	}
	if l.Selected != -1 {
		t.Fatalf("dragging should not select an item, got %d", l.Selected)
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

func TestBoxCutsAndSplits(t *testing.T) {
	root := BoxFromRect(Rect{X: 10, Y: 20, W: 300, H: 200}).Inset(10)
	top, rest := root.CutTop(40, 8)
	if top.X != 20 || top.Y != 30 || top.W != 280 || top.H != 40 {
		t.Fatalf("unexpected top rect: %+v", top)
	}
	if rest.Rect.Y != 78 || rest.Rect.H != 132 {
		t.Fatalf("unexpected rest box: %+v", rest.Rect)
	}

	cols := rest.SplitColumns(12, 2, 3)
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	if cols[0].X != 20 || cols[1].X <= cols[0].X+cols[0].W {
		t.Fatalf("expected gapped columns, got %+v", cols)
	}
	if cols[0].H != rest.Rect.H || cols[1].H != rest.Rect.H {
		t.Fatalf("expected full-height columns, got %+v", cols)
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
