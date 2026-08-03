package render

import (
	"testing"

	"mapp-game-go/internal/state"
)

func TestStrategicTurnDateTRShowsQuarterlyRange(t *testing.T) {
	if got := strategicTurnDateTR(&state.GameState{Year: 1300, Month: 3, MonthsPerTurn: 3}); got != "Mart–Mayıs 1300" {
		t.Fatalf("üç aylık tur aralığı görünmeliydi: %q", got)
	}
	if got := strategicTurnDateTR(&state.GameState{Year: 1300, Month: 12, MonthsPerTurn: 3}); got != "Aralık 1300–Şubat 1301" {
		t.Fatalf("yıl aşan tur aralığı görünmeliydi: %q", got)
	}
	if got := strategicTurnDateTR(&state.GameState{Year: 1300, Month: 3}); got != "Mart 1300" {
		t.Fatalf("eski aylık gösterim korunmalıydı: %q", got)
	}
}
