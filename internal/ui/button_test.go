package ui

import (
	"math"
	"testing"
)

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

func TestButtonFocusableReflectsEnabledState(t *testing.T) {
	btn := NewButton(0, 0, 10, 10, "Test")
	if !btn.IsFocusable() {
		t.Fatalf("enabled button should be focusable")
	}
	btn.SetFocused(true)
	if !btn.Focused {
		t.Fatalf("button focus flag should be set")
	}
	btn.Enabled = false
	if btn.IsFocusable() {
		t.Fatalf("disabled button should not be focusable")
	}
}

func TestButtonWithIconPreservesButtonAndSetsIcon(t *testing.T) {
	btn := NewButton(4, 5, 60, 20, "Geri").WithIcon(IconBack)
	if btn.Icon != IconBack {
		t.Fatalf("expected icon to be set")
	}
	if btn.Label != "Geri" || btn.X != 4 || btn.Y != 5 {
		t.Fatalf("button data should be preserved after WithIcon")
	}
}

func TestButtonIconSizeUsesDifferentScalesByHeight(t *testing.T) {
	smallH := 22.0
	mediumH := 28.0
	largeH := 40.0
	small := buttonIconSize(NewButton(0, 0, 60, smallH, "").WithIcon(IconBack))
	medium := buttonIconSize(NewButton(0, 0, 80, mediumH, "").WithIcon(IconBack))
	large := buttonIconSize(NewButton(0, 0, 100, largeH, "").WithIcon(IconBack))
	if !(small/smallH > medium/mediumH && medium/mediumH > large/largeH) {
		t.Fatalf("expected icon ratio to shrink as buttons get taller: small=%v medium=%v large=%v", small/smallH, medium/mediumH, large/largeH)
	}
	if math.Abs(small-18.0) > 0.001 {
		t.Fatalf("unexpected small icon size: got %v want 18.0", small)
	}
	if math.Abs(medium-21.28) > 0.001 {
		t.Fatalf("unexpected medium icon size: got %v want 21.28", medium)
	}
	if math.Abs(large-27.2) > 0.001 {
		t.Fatalf("unexpected large icon size: got %v want 27.2", large)
	}
}

func TestButtonTextYUsesExplicitOffsetWhenProvided(t *testing.T) {
	btn := NewButton(20, 40, 120, 36, "Gönder")
	y := buttonTextY(btn, ButtonStyle{TextVariant: TextMedium, TextOffsetY: 9})
	if y != 49 {
		t.Fatalf("expected explicit text offset to be used, got %v want 49", y)
	}
}
