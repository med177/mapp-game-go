package diplomacy

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

// QueueOffer geçerli ve tekrar etmeyen diplomatik teklifi kuyruğa ekler.
func QueueOffer(gs *state.GameState, from, to faction.FactionID, action Action) bool {
	if gs == nil || from == "" || to == "" || from == to {
		return false
	}
	if action != ActionProposePeace && action != ActionProposeAlliance && action != ActionProposeTrade {
		return false
	}
	fromFaction := gs.Factions[from]
	toFaction := gs.Factions[to]
	if fromFaction == nil || toFaction == nil || fromFaction.IsEliminated || toFaction.IsEliminated {
		return false
	}
	for _, offer := range gs.DiplomaticOffers {
		if offer.FromFactionID == from && offer.ToFactionID == to && offer.Action == string(action) {
			return false
		}
	}
	gs.DiplomaticOffers = append(gs.DiplomaticOffers, state.DiplomaticOffer{
		FromFactionID: from,
		ToFactionID:   to,
		Action:        string(action),
		CreatedTurn:   gs.Turn,
	})
	return true
}

// ResolveOffer teklifi kabul/red ile sonuçlandırır ve kuyruktan düşürür.
func ResolveOffer(gs *state.GameState, index int, accepted bool) Result {
	if gs == nil || index < 0 || index >= len(gs.DiplomaticOffers) {
		return Result{Message: "Geçersiz diplomasi teklifi."}
	}
	offer := gs.DiplomaticOffers[index]
	gs.DiplomaticOffers = append(gs.DiplomaticOffers[:index], gs.DiplomaticOffers[index+1:]...)

	action := Action(offer.Action)
	if !accepted {
		return Result{
			Accepted: false,
			Applied:  false,
			Message:  factionLabel(gs, offer.FromFactionID) + " teklifiniz reddedildi.",
		}
	}
	return Execute(gs, offer.FromFactionID, offer.ToFactionID, action)
}
