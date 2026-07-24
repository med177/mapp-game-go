package ai

import (
	"testing"

	"mapp-game-go/internal/faction"
)

func TestRecordAIDiagnosticRoundCapturesSortedAIHistory(t *testing.T) {
	gs := aiFrontTestState()
	gs.PlayerFactionID = "other"
	gs.DevelopmentMode = true
	gs.Turn = 14
	gs.AIDiagnosticCaptureTurnsRemain = 1
	gs.Relations[faction.RelationKey("ai", "enemy")].Stance = faction.StanceWar
	gs.BeginWarLedger("ai", "enemy")

	if !RecordAIDiagnosticRound(gs) {
		t.Fatal("geliştirme modunda AI teşhis turu kaydedilmeliydi")
	}
	if gs.AIDiagnosticCaptureTurnsRemain != 0 {
		t.Fatalf("tek kayıt sonrası kalan tur sıfırlanmalıydı: %d", gs.AIDiagnosticCaptureTurnsRemain)
	}
	if len(gs.AIDiagnosticHistory) != 2 {
		t.Fatalf("oyuncu dışındaki iki AI fraksiyonu kaydedilmeliydi: %+v", gs.AIDiagnosticHistory)
	}
	if gs.AIDiagnosticHistory[0].FactionID != "ai" || gs.AIDiagnosticHistory[1].FactionID != "enemy" {
		t.Fatalf("teşhis geçmişi deterministik ID sırasına sahip olmalıydı: %+v", gs.AIDiagnosticHistory)
	}
	if got := gs.AIDiagnosticHistory[0]; got.ActiveWarCount != 1 || got.TargetRegionID == "" || got.FrontCount == 0 {
		t.Fatalf("aktif savaş cephe özeti eksik: %+v", got)
	}
}
