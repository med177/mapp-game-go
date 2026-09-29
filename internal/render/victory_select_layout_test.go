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
