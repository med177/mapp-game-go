package ui

import "testing"

func TestTextBoxFocusInputAndBackspace(t *testing.T) {
	tb := NewTextBox(10, 10, 100, 24, "name")
	if !tb.HandleInput(InputState{MouseX: 20, MouseY: 20, LeftJustPressed: true}) {
		t.Fatalf("textbox click should focus and consume")
	}
	if !tb.Focused {
		t.Fatalf("textbox should be focused")
	}
	if !tb.HandleInput(InputState{TextInput: "abc"}) {
		t.Fatalf("textbox text input should be consumed")
	}
	if tb.Value != "abc" {
		t.Fatalf("expected abc, got %q", tb.Value)
	}
	if !tb.HandleInput(InputState{BackspaceJustPressed: true}) {
		t.Fatalf("textbox backspace should be consumed")
	}
	if tb.Value != "ab" {
		t.Fatalf("expected ab, got %q", tb.Value)
	}
}

func TestTextBoxMaxLen(t *testing.T) {
	tb := NewTextBox(0, 0, 100, 24, "")
	tb.Focused = true
	tb.MaxLen = 2
	tb.HandleInput(InputState{TextInput: "abcd"})
	if tb.Value != "ab" {
		t.Fatalf("expected max-len clipped value, got %q", tb.Value)
	}
}

func TestAnchorRect(t *testing.T) {
	parent := Rect{X: 10, Y: 20, W: 200, H: 100}
	got := AnchorRect(parent, 50, 20, AnchorRight, AnchorBottom, 8, 6)
	if got.X != 152 || got.Y != 94 || got.W != 50 || got.H != 20 {
		t.Fatalf("unexpected anchored rect: %+v", got)
	}
}

func TestTooltipCopiesLines(t *testing.T) {
	lines := []string{"a", "b"}
	tt := NewTooltip(0, 0, 120, lines, 16, 4)
	lines[0] = "changed"
	if tt.Lines[0] != "a" {
		t.Fatalf("tooltip should copy line slice")
	}
	if tt.Rect.H != 40 {
		t.Fatalf("expected tooltip height 40, got %v", tt.Rect.H)
	}
}

func TestImageHitTest(t *testing.T) {
	img := NewImage(10, 10, 20, 20, nil)
	if !img.HitTest(15, 15) {
		t.Fatalf("image hit-test should use rect")
	}
	if img.HitTest(50, 50) {
		t.Fatalf("image should ignore outside point")
	}
}
