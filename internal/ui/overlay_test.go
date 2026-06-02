package ui

import "testing"

func TestOverlayConsumesInsideClickWhenEnabled(t *testing.T) {
	o := NewOverlay(10, 20, 100, 40)
	o.ConsumeInput = true
	if !o.HandleInput(InputState{MouseX: 20, MouseY: 30, LeftJustPressed: true}) {
		t.Fatalf("overlay should consume inside click when enabled")
	}
}

func TestOverlayIgnoresOutsideClick(t *testing.T) {
	o := NewOverlay(10, 20, 100, 40)
	o.ConsumeInput = true
	if o.HandleInput(InputState{MouseX: 500, MouseY: 500, LeftJustPressed: true}) {
		t.Fatalf("overlay should ignore outside click")
	}
}
