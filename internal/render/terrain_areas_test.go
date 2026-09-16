package render

import (
	"image/color"
	"testing"
)

func TestTerrainAreaTypeColorKeepsAlpha(t *testing.T) {
	base := color.RGBA{12, 34, 56, 85}
	got := terrainAreaTypeColor(base, color.RGBA{98, 65, 32, 255})
	want := color.RGBA{98, 65, 32, 85}
	if got != want {
		t.Fatalf("terrain area type color = %#v, want %#v", got, want)
	}
}
