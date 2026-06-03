package faction

import "testing"

func TestDiplomaticStanceMetadata(t *testing.T) {
	if got := DiplomaticStanceLabelTR(StanceTrade); got != "Ticaret" {
		t.Fatalf("stance label mismatch: got=%q", got)
	}
	if got := DiplomaticStanceBadgeTR(StanceWar); got != "WAR Savaş" {
		t.Fatalf("stance badge mismatch: got=%q", got)
	}
}

func TestNextDiplomaticStance(t *testing.T) {
	if got := NextDiplomaticStance(StanceAllied); got != StanceTrade {
		t.Fatalf("next stance mismatch: got=%q want=%q", got, StanceTrade)
	}
}
