package diplomacy

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

const diplomacyOfferQuotaBlockReasonTR = "Bu tur diplomasi elçisi hakkın doldu."

func actionUsesDiplomacyOfferQuota(action Action) bool {
	switch action {
	case ActionProposePeace, ActionProposeAlliance, ActionProposeTrade, ActionImproveRelations, ActionSendGift, ActionOfferVassalization, ActionJoinWarCall:
		return true
	default:
		return false
	}
}

func diplomacyOfferQuotaBlockReason(gs *state.GameState, actor faction.FactionID) string {
	if gs == nil || actor == "" {
		return ""
	}
	if gs.DiplomacyOfferQuotaRemaining(actor) > 0 {
		return ""
	}
	return diplomacyOfferQuotaBlockReasonTR
}

func spendDiplomacyOfferQuota(gs *state.GameState, actor faction.FactionID) bool {
	if gs == nil || actor == "" {
		return false
	}
	return gs.SpendDiplomacyOfferQuota(actor)
}
