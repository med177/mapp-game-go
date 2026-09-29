package render

import "testing"

func TestFactionHUDNameTextStopsBeforeResourceColumn(t *testing.T) {
	originalWidth := ScreenWidth
	defer func() { ScreenWidth = originalWidth }()
	ScreenWidth = 1050

	textX := 5 + factionHUDFlagSize + 13
	name := factionHUDNameText("Fransa Krallığı (Capet/Valois)", textX)
	leftCol1, _, _, _, _ := topResourceHUDColumns()
	maxWidth := leftCol1 - textX - 12

	if name == "Fransa Krallığı (Capet/Valois)" {
		t.Fatal("uzun devlet adı üst HUD kaynak sütunundan önce kırpılmadı")
	}
	if MeasureText(name, FaceLarge) > maxWidth {
		t.Fatalf("kırpılmış devlet adı genişliği = %.1f, üst sınır %.1f", MeasureText(name, FaceLarge), maxWidth)
	}
}
