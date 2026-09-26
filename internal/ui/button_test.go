package ui

import "testing"

func TestButtonTextYIsVerticallyCentered(t *testing.T) {
	button := Button{Y: 100, H: 32}
	style := ButtonStyle{TextVariant: TextSmall}

	want := 110.0
	if got := buttonTextY(button, style); got != want {
		t.Fatalf("buttonTextY() = %v, want %v", got, want)
	}
}

func TestNewCloseButtonUsesSharedIconAndProportionalSize(t *testing.T) {
	button := NewCloseButton(10, 20, 30, 24)

	if button.Icon != IconClose {
		t.Fatalf("NewCloseButton() icon = %q, want %q", button.Icon, IconClose)
	}
	if button.IconSize != 12 {
		t.Fatalf("NewCloseButton() icon size = %v, want 12", button.IconSize)
	}
	if button.Label != "" || !button.Enabled {
		t.Fatalf("NewCloseButton() = %+v, want enabled icon-only button", button)
	}
}
