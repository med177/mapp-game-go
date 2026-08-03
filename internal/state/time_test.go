package state

import "testing"

func TestQuarterlyTurnCoversCalendarRangeAndAdvancesThreeMonths(t *testing.T) {
	gs := &GameState{Year: 1300, Month: 3, MonthsPerTurn: 3}
	endYear, endMonth := gs.CurrentTurnEndDate()
	if endYear != 1300 || endMonth != 5 {
		t.Fatalf("Mart başlangıçlı mevsim turu Mayıs'ta bitmeli: %d/%d", endYear, endMonth)
	}
	if !gs.HistoricalDateOccursThisTurn(1300, 5) || gs.HistoricalDateOccursThisTurn(1300, 6) {
		t.Fatalf("tarihsel olay takvim aralığı yanlış hesaplandı")
	}
	if gs.CurrentTurnIncludesMonth(12) {
		t.Fatal("Mart-Mayıs turu yıl sınırını geçmemeli")
	}

	gs.AdvanceTurn()
	if gs.Turn != 1 || gs.Year != 1300 || gs.Month != 6 {
		t.Fatalf("üç aylık tur Haziran 1300'e ilerlemeli: %+v", gs)
	}

	gs.Month = 12
	if !gs.CurrentTurnIncludesMonth(12) {
		t.Fatal("Aralık başlangıçlı tur yıl sınırını geçmeli")
	}
	gs.AdvanceTurn()
	if gs.Year != 1301 || gs.Month != 3 {
		t.Fatalf("Aralık başlangıçlı tur Mart 1301'e ilerlemeli: %d/%d", gs.Year, gs.Month)
	}
}
