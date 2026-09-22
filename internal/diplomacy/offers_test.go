package diplomacy

import "testing"

func TestIsRelationshipNotificationOnlyMatchesNonBlockingOffers(t *testing.T) {
	for _, action := range []Action{ActionImproveRelations, ActionSendGift} {
		if !IsRelationshipNotification(action) {
			t.Fatalf("%q ilişki bildirimi olarak işaretlenmedi", action)
		}
	}
	for _, action := range []Action{ActionProposePeace, ActionProposeTrade, ActionProposeAlliance, ActionOfferVassalization} {
		if IsRelationshipNotification(action) {
			t.Fatalf("%q oyuncu kararı isteyen teklif olmasına rağmen bildirim sayıldı", action)
		}
	}
}
