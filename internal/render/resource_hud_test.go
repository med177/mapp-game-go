package render

import "testing"

func TestFormatResourceHUDValue(t *testing.T) {
	tests := []struct {
		name    string
		current int
		change  int
		want    string
	}{
		{name: "positive", current: 405, change: 55, want: "+55/405"},
		{name: "negative", current: 405, change: -12, want: "-12/405"},
		{name: "zero", current: 405, change: 0, want: "0/405"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatResourceHUDValue(tt.current, tt.change); got != tt.want {
				t.Fatalf("HUD kaynak değeri yanlış: got=%q want=%q", got, tt.want)
			}
		})
	}
}

func TestTurnTechHudTechHitUsesResearchRow(t *testing.T) {
	oldScreenW, oldScreenH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldScreenW, oldScreenH
	}()

	x, y, w, h := turnTechHudTechRect()
	if !turnTechHudTechHit(float64(x+w/2), float64(y+h/2)) {
		t.Fatal("aktif teknoloji satırının ortası tıklanabilir olmalı")
	}
	_, hudY, _, _ := turnTechHudRect()
	if turnTechHudTechHit(float64(x+w/2), float64(hudY+8)) {
		t.Fatal("teknoloji HUD'ının üst durum satırı teknoloji açma hit-test'ine girmemeli")
	}
	if turnTechHudTechHit(float64(x-1), float64(y+h/2)) {
		t.Fatal("teknoloji HUD'ının dışı hit-test'e girmemeli")
	}
}
