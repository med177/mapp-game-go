package render

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestUnitTooltipLayoutPlacesRequirementsBelowImage(t *testing.T) {
	const iconH = 201.0

	layout := newUnitTooltipLayout(iconH, 6, 12)

	if layout.requirementY < 14+iconH+10 {
		t.Fatalf("gereksinim başlığı ikonun altına taşmıyor: got %v", layout.requirementY)
	}
	if layout.requirementLinesY != layout.requirementY+14 {
		t.Fatalf("gereksinim satır başlangıcı = %v, want %v", layout.requirementLinesY, layout.requirementY+14)
	}
	if layout.upkeepY >= layout.requirementY {
		t.Fatalf("bakım üst bölüme taşınmamış: bakım=%v gereksinim=%v", layout.upkeepY, layout.requirementY)
	}
	if layout.attributesY >= layout.requirementY {
		t.Fatalf("nitelikler üst bölüme taşınmamış: nitelik=%v gereksinim=%v", layout.attributesY, layout.requirementY)
	}
	if layout.height <= layout.requirementLinesY+12*14 {
		t.Fatalf("tooltip yüksekliği gereksinim bloğunu kapsamıyor: height=%v", layout.height)
	}
}

func TestUnitTooltipImageMetricsPreservesSpriteAspect(t *testing.T) {
	sprite := ebiten.NewImage(300, 500)
	width, height := unitTooltipImageMetrics(sprite)

	if got, want := width/height, 300.0/500.0; math.Abs(got-want) > 1e-9 {
		t.Fatalf("tooltip görsel oranı = %v, want %v", got, want)
	}
}
