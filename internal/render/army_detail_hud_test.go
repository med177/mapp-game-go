package render

import "testing"

func TestArmyDetailHUDButtonUsesSelectionSpecificLabelAndPosition(t *testing.T) {
	armyButton := BottomButtonRects()[0]
	land := buildArmyDetailHUDButton(false)
	naval := buildArmyDetailHUDButton(true)

	if land.Label != "Ordu Detay" {
		t.Fatalf("kara etiketi = %q, want %q", land.Label, "Ordu Detay")
	}
	if naval.Label != "Donanma Detay" {
		t.Fatalf("donanma etiketi = %q, want %q", naval.Label, "Donanma Detay")
	}
	if land.X+land.W != float64(armyButton[0]-8) {
		t.Fatalf("detay düğmesi Ordu düğmesinden 8 px ayrılmadı: detay sağ=%v ordu sol=%v", land.X+land.W, armyButton[0])
	}
	if naval.X != land.X || naval.Y != land.Y || naval.W != land.W || naval.H != land.H {
		t.Fatal("kara ve donanma detay düğmeleri aynı geometry kaynağını kullanmıyor")
	}
}
