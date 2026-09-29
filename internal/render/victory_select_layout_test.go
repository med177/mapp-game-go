package render

import "testing"

func TestHistoricalVictoryCardsUseFullRow(t *testing.T) {
	const (
		total      = 8
		historical = 5
		cardW      = 780.0
		cardH      = 126.0
		gap        = 12.0
		headerH    = 80.0
	)

	for i := 0; i < historical; i++ {
		rect := victoryCardRect(i, total, historical, cardW, cardH, gap, headerH)
		if rect.W != cardW {
			t.Fatalf("tarihsel kart %d tam satır genişliğinde değil: genişlik %.1f, beklenen %.1f", i, rect.W, cardW)
		}
	}

	firstGeneral := victoryCardRect(historical, total, historical, cardW, cardH, gap, headerH)
	secondGeneral := victoryCardRect(historical+1, total, historical, cardW, cardH, gap, headerH)
	if firstGeneral.W >= cardW || secondGeneral.W >= cardW {
		t.Fatalf("genel hedeflerin iki sütunlu düzeni korunmadı: %.1f, %.1f", firstGeneral.W, secondGeneral.W)
	}
}

func TestVictorySelectScrollKeepsViewportFixed(t *testing.T) {
	const (
		total      = 12
		historical = 8
		cardW      = 780.0
		cardH      = 126.0
		gap        = 12.0
		headerH    = 80.0
	)

	atTop := victoryLayoutScrolled(total, historical, cardW, cardH, gap, headerH, 0)
	maxScroll := victorySelectMaxScroll(total, historical, cardW, cardH, gap, headerH)
	atBottom := victoryLayoutScrolled(total, historical, cardW, cardH, gap, headerH, maxScroll)
	if maxScroll <= 0 {
		t.Fatal("uzun zafer listesi için pozitif scroll alanı oluşturulmadı")
	}
	if atTop.viewport != atBottom.viewport {
		t.Fatalf("scroll viewport'i hareket ettirdi: üst=%+v alt=%+v", atTop.viewport, atBottom.viewport)
	}
	if atBottom.historicalStack.Y >= atTop.historicalStack.Y {
		t.Fatalf("scroll kart listesini aşağı taşımadı: üst=%.1f alt=%.1f", atTop.historicalStack.Y, atBottom.historicalStack.Y)
	}
}
