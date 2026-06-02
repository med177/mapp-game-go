package ui

import "testing"

func TestButtonHandleInputConsumesInsideClick(t *testing.T) {
	btn := NewButton(10, 20, 100, 30, "Test")
	input := InputState{MouseX: 50, MouseY: 35, LeftJustPressed: true}
	if !btn.HandleInput(input) {
		t.Fatalf("button click should be consumed")
	}
}

func TestButtonHandleInputIgnoresOutsideClick(t *testing.T) {
	btn := NewButton(10, 20, 100, 30, "Test")
	input := InputState{MouseX: 500, MouseY: 350, LeftJustPressed: true}
	if btn.HandleInput(input) {
		t.Fatalf("outside click should not be consumed")
	}
}

func TestButtonHandleInputIgnoresDisabledButton(t *testing.T) {
	btn := NewButton(10, 20, 100, 30, "Test")
	btn.Enabled = false
	input := InputState{MouseX: 50, MouseY: 35, LeftJustPressed: true}
	if btn.HandleInput(input) {
		t.Fatalf("disabled button should not consume click")
	}
}
