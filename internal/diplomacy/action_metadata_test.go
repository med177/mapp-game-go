package diplomacy

import "testing"

func TestActionLabelTR(t *testing.T) {
	if got := ActionLabelTR(ActionProposeAlliance); got != "İttifak" {
		t.Fatalf("action label mismatch: got=%q", got)
	}
	if got := ActionLabelTR(ActionOfferVassalization); got != "Vassallık" {
		t.Fatalf("new action label mismatch: got=%q", got)
	}
}

func TestVisibleActionsOrder(t *testing.T) {
	actions := VisibleActions()
	if len(actions) != 7 {
		t.Fatalf("unexpected action count: got=%d", len(actions))
	}
	if actions[0] != ActionDeclareWar || actions[3] != ActionProposeTrade || actions[6] != ActionOfferVassalization {
		t.Fatalf("unexpected action order: got=%v", actions)
	}
}

func TestQuickActionsRemainCoreSet(t *testing.T) {
	actions := QuickActions()
	if len(actions) != 4 {
		t.Fatalf("unexpected quick action count: got=%d", len(actions))
	}
	if actions[0] != ActionDeclareWar || actions[3] != ActionProposeTrade {
		t.Fatalf("unexpected quick action order: got=%v", actions)
	}
}
