package state

import "testing"

func TestLandContactHoldRequiresBattleResolution(t *testing.T) {
	contact := &LandContact{
		AttackerDecision: LandContactHold,
		DefenderDecision: LandContactHold,
	}
	if !(&GameState{}).LandContactWillClash(contact) {
		t.Fatal("iki taraf pozisyonunu koruduğunda temas savaş planına gitmeli")
	}
	contact.DefenderDecision = LandContactWithdraw
	if (&GameState{}).LandContactWillClash(contact) {
		t.Fatal("taraflardan biri geri çekilirse temas savaşa dönüşmemeli")
	}
}
