package diplomacy

import "testing"

func TestActionLabelTR(t *testing.T) {
	if got := ActionLabelTR(ActionProposeAlliance); got != "İttifak" {
		t.Fatalf("action label mismatch: got=%q", got)
	}
}

func TestVisibleActionsOrder(t *testing.T) {
	actions := VisibleActions()
	if len(actions) != 4 {
		t.Fatalf("unexpected action count: got=%d", len(actions))
	}
	if actions[0] != ActionDeclareWar || actions[3] != ActionProposeTrade {
		t.Fatalf("unexpected action order: got=%v", actions)
	}
}
