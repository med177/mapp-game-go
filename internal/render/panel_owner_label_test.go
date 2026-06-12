package render

import (
	"image/color"
	"testing"
)

func TestOwnerLabelOutlineColorUsesOppositeContrast(t *testing.T) {
	darkOutline := ownerLabelOutlineColor(color.RGBA{210, 210, 210, 255})
	if darkOutline != (color.RGBA{18, 16, 12, 220}) {
		t.Fatalf("acik renk icin koyu outline bekleniyordu, alinan=%v", darkOutline)
	}

	lightOutline := ownerLabelOutlineColor(color.RGBA{30, 70, 145, 255})
	if lightOutline != (color.RGBA{245, 240, 230, 210}) {
		t.Fatalf("koyu renk icin acik outline bekleniyordu, alinan=%v", lightOutline)
	}
}
