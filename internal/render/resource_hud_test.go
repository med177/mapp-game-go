package render

import "testing"

func TestFormatResourceHUDValue(t *testing.T) {
	tests := []struct {
		name    string
		current int
		change  int
		want    string
	}{
		{name: "positive", current: 405, change: 55, want: "+55 / 405"},
		{name: "negative", current: 405, change: -12, want: "-12 / 405"},
		{name: "zero", current: 405, change: 0, want: "0 / 405"},
		{name: "thousands", current: 10672, change: 253, want: "+253 / 10.672"},
		{name: "negative thousands", current: 12925, change: -195, want: "-195 / 12.925"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatResourceHUDValue(tt.current, tt.change); got != tt.want {
				t.Fatalf("HUD kaynak değeri yanlış: got=%q want=%q", got, tt.want)
			}
		})
	}
}

func TestFormatNumberTR(t *testing.T) {
	tests := map[int]string{
		0:       "0",
		999:     "999",
		1000:    "1.000",
		10000:   "10.000",
		-12925:  "-12.925",
		1000000: "1.000.000",
	}
	for input, want := range tests {
		if got := formatNumberTR(input); got != want {
			t.Fatalf("Türkçe sayı biçimi yanlış: input=%d got=%q want=%q", input, got, want)
		}
	}
}

func TestTopResourceHUDColumnsKeepIncomeSeparate(t *testing.T) {
	leftCol1, leftCol2, rightCol, leftColW, rightColW := topResourceHUDColumns()
	if leftCol1+leftColW > leftCol2 {
		t.Fatalf("ilk kaynak sütunu sonraki sütuna taşıyor: end=%.1f next=%.1f", leftCol1+leftColW, leftCol2)
	}
	if leftCol2+leftColW > rightCol {
		t.Fatalf("ikinci kaynak sütunu Gelir sütununa taşıyor: end=%.1f income=%.1f", leftCol2+leftColW, rightCol)
	}
	if rightCol+rightColW > victoryProgressPanelRect().X-12 {
		t.Fatalf("Gelir/Altın sütunu zafer kartına taşıyor: end=%.1f card=%.1f", rightCol+rightColW, victoryProgressPanelRect().X-12)
	}
	if rightCol <= 578 {
		t.Fatalf("Gelir/Altın sütunu sağa alınmadı: x=%.1f", rightCol)
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
