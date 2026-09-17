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
