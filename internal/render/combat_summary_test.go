package render

import (
	"testing"

	"mapp-game-go/internal/faction"
)

func TestCollectCombatSummaryTotalsSeparatesArmyFleetAndDestroyed(t *testing.T) {
	totals := collectCombatSummaryTotals([]CombatSummaryEntry{
		{AttackerFactionID: "a", DefenderFactionID: "b", AttackerLost: 3, DefenderLost: 1, AttackerNaval: false, DefenderNaval: true, DefenderDestroyed: true},
		{AttackerFactionID: "b", DefenderFactionID: "a", AttackerLost: 2, DefenderLost: 4, AttackerNaval: true, DefenderNaval: false, AttackerDestroyed: true},
	})
	if len(totals) != 2 {
		t.Fatalf("fraksiyon toplamı sayısı = %d, want 2", len(totals))
	}
	if totals[0].FactionID != faction.FactionID("a") || totals[0].ArmyLost != 7 || totals[0].FleetLost != 0 || totals[0].ArmyDestroyed != 0 {
		t.Fatalf("A toplamı beklenmeyen değer taşıyor: %+v", totals[0])
	}
	if totals[1].FactionID != faction.FactionID("b") || totals[1].ArmyLost != 0 || totals[1].FleetLost != 3 || totals[1].ArmyDestroyed != 0 || totals[1].FleetDestroyed != 2 {
		t.Fatalf("B toplamı beklenmeyen değer taşıyor: %+v", totals[1])
	}
}
