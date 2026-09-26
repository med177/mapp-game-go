package ui

import (
	"image"
	"image/color"
	"testing"
)

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

func TestIconFitDimensionsPreservesAspectRatio(t *testing.T) {
	if gotW, gotH := iconFitDimensions(128, 64, 24); gotW != 24 || gotH != 12 {
		t.Fatalf("iconFitDimensions() = (%v, %v), want (24, 12)", gotW, gotH)
	}
	if gotW, gotH := iconFitDimensions(64, 128, 24); gotW != 12 || gotH != 24 {
		t.Fatalf("iconFitDimensions() = (%v, %v), want (12, 24)", gotW, gotH)
	}
}

func TestCropTransparentBorderRemovesOnlyTransparentPadding(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 8, 6))
	src.SetNRGBA(2, 1, color.NRGBA{R: 255, A: 255})
	src.SetNRGBA(6, 4, color.NRGBA{B: 255, A: 128})

	cropped := cropTransparentBorder(src)
	if got := cropped.Bounds().Size(); got.X != 5 || got.Y != 4 {
		t.Fatalf("cropTransparentBorder() size = %v, want (5, 4)", got)
	}
}
