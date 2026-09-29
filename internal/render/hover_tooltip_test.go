package render

import "testing"

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
