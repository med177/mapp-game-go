package diplomacy

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

const warJoinOfferPriority = 100

// QueueOffer geçerli ve tekrar etmeyen diplomatik teklifi kuyruğa ekler.
func QueueOffer(gs *state.GameState, from, to faction.FactionID, action Action) bool {
	return QueueOfferWithPriority(gs, from, to, action, 0)
}

// QueueOfferWithPriority geçerli ve tekrar etmeyen diplomatik teklifi önceliğiyle kuyruğa ekler.
func QueueOfferWithPriority(gs *state.GameState, from, to faction.FactionID, action Action, priority int) bool {
	return QueueOfferWithMeta(gs, from, to, action, priority, "")
}

// QueueOfferWithMeta geçerli ve tekrar etmeyen diplomatik teklifi öncelik ve sebep bilgisiyle kuyruğa ekler.
func QueueOfferWithMeta(gs *state.GameState, from, to faction.FactionID, action Action, priority int, reason string) bool {
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
	if !spendDiplomacyOfferQuota(gs, from) {
		return false
	}
	gs.DiplomaticOffers = append(gs.DiplomaticOffers, state.DiplomaticOffer{
		FromFactionID:  from,
		ToFactionID:    to,
		Action:         string(action),
		CreatedTurn:    gs.Turn,
		Priority:       priority,
		PriorityReason: reason,
	})
	return true
}

func QueueWarJoinOffer(gs *state.GameState, caller, player, declarer, enemy faction.FactionID, reason string) bool {
	if gs == nil || caller == "" || player == "" || caller == player || declarer == "" || enemy == "" {
		return false
	}
	callerFaction := gs.Factions[caller]
	playerFaction := gs.Factions[player]
	if callerFaction == nil || playerFaction == nil || callerFaction.IsEliminated || playerFaction.IsEliminated {
		return false
	}
	for _, offer := range gs.DiplomaticOffers {
		if offer.Action != string(ActionJoinWarCall) {
			continue
		}
		if offer.FromFactionID == caller && offer.ToFactionID == player &&
			offer.WarDeclarerFactionID == declarer && offer.WarEnemyFactionID == enemy {
			return false
		}
	}
	if reason == "" {
		reason = "Aktif ittifak nedeniyle savaş çağrısı"
	}
	if !spendDiplomacyOfferQuota(gs, caller) {
		return false
	}
	gs.DiplomaticOffers = append(gs.DiplomaticOffers, state.DiplomaticOffer{
		FromFactionID:        caller,
		ToFactionID:          player,
		Action:               string(ActionJoinWarCall),
		CreatedTurn:          gs.Turn,
		Priority:             warJoinOfferPriority,
		PriorityReason:       reason,
		WarDeclarerFactionID: declarer,
		WarEnemyFactionID:    enemy,
	})
	return true
}

// BestOfferIndex belirtilen hedef için en yüksek öncelikli diplomatik teklifin indeksini döner.
func BestOfferIndex(gs *state.GameState, target faction.FactionID) (int, bool) {
	if gs == nil || target == "" || len(gs.DiplomaticOffers) == 0 {
		return 0, false
	}

	bestIdx := -1
	bestPriority := 0
	bestTurn := 0
	found := false
	for i, offer := range gs.DiplomaticOffers {
		if offer.ToFactionID != target {
			continue
		}
		fromFaction := gs.Factions[offer.FromFactionID]
		toFaction := gs.Factions[offer.ToFactionID]
		if fromFaction == nil || toFaction == nil || fromFaction.IsEliminated || toFaction.IsEliminated {
			continue
		}
		if !found || offer.Priority > bestPriority || (offer.Priority == bestPriority && (offer.CreatedTurn < bestTurn || (offer.CreatedTurn == bestTurn && i < bestIdx))) {
			bestIdx = i
			bestPriority = offer.Priority
			bestTurn = offer.CreatedTurn
			found = true
		}
	}
	return bestIdx, found
}

// ResolveOffer teklifi kabul/red ile sonuçlandırır ve kuyruktan düşürür.
func ResolveOffer(gs *state.GameState, index int, accepted bool) Result {
	if gs == nil || index < 0 || index >= len(gs.DiplomaticOffers) {
		return Result{Message: "Geçersiz diplomasi teklifi."}
	}
	offer := gs.DiplomaticOffers[index]
	gs.DiplomaticOffers = append(gs.DiplomaticOffers[:index], gs.DiplomaticOffers[index+1:]...)

	action := Action(offer.Action)
	if action == ActionJoinWarCall {
		if accepted {
			return resolveAcceptedWarJoinOffer(gs, offer)
		}
		return resolveRejectedWarJoinOffer(gs, offer)
	}
	if !accepted {
		markRejectedDiplomaticOffer(gs, offer.FromFactionID, offer.ToFactionID, action)
		return Result{
			Accepted: false,
			Applied:  false,
			Message:  factionLabel(gs, offer.FromFactionID) + " teklif reddedildi.",
		}
	}
	if action == ActionProposePeace {
		rel := EnsureRelation(gs, offer.FromFactionID, offer.ToFactionID)
		if rel.Stance != faction.StanceWar {
			return Result{Message: "Barış teklifi artık geçerli değil."}
		}
		setPeaceBetweenCoalitions(gs, offer.FromFactionID, offer.ToFactionID)
		return Result{
			Accepted: true,
			Applied:  true,
			Message:  factionLabel(gs, offer.ToFactionID) + " barışı kabul etti.",
		}
	}
	result := Execute(gs, offer.FromFactionID, offer.ToFactionID, action)
	if accepted && !result.Applied {
		return Result{
			Accepted: false,
			Applied:  false,
			Message:  "Teklif koşulları değiştiği için uygulanamadı.",
		}
	}
	return result
}

func resolveAcceptedWarJoinOffer(gs *state.GameState, offer state.DiplomaticOffer) Result {
	callerRoot := realmRoot(gs, offer.FromFactionID)
	if callerRoot == "" {
		callerRoot = offer.FromFactionID
	}
	playerRoot := realmRoot(gs, offer.ToFactionID)
	if playerRoot == "" {
		playerRoot = offer.ToFactionID
	}
	enemyRoot := realmRoot(gs, offer.WarEnemyFactionID)
	if enemyRoot == "" {
		enemyRoot = offer.WarEnemyFactionID
	}
	if callerRoot == "" || playerRoot == "" || enemyRoot == "" {
		return Result{Message: "Savaş çağrısı artık geçerli değil."}
	}
	if IsWar(gs, playerRoot, enemyRoot) {
		return Result{
			Accepted: true,
			Applied:  true,
			Message:  factionLabel(gs, enemyRoot) + " savaşına zaten dahilsiniz.",
		}
	}
	if assessment := AssessWarCall(gs, callerRoot, playerRoot, enemyRoot); assessment.BlockReason != "" {
		return Result{Message: "Savaş çağrısı artık geçerli değil."}
	}
	setWarBetweenCoalitions(gs, playerRoot, enemyRoot)
	return Result{
		Accepted: true,
		Applied:  true,
		Message:  factionLabel(gs, callerRoot) + " tarafında " + factionLabel(gs, enemyRoot) + " savaşına katıldınız.",
	}
}

func resolveRejectedWarJoinOffer(gs *state.GameState, offer state.DiplomaticOffer) Result {
	callerRoot := realmRoot(gs, offer.FromFactionID)
	if callerRoot == "" {
		callerRoot = offer.FromFactionID
	}
	playerRoot := realmRoot(gs, offer.ToFactionID)
	if playerRoot == "" {
		playerRoot = offer.ToFactionID
	}
	if callerRoot != "" && playerRoot != "" {
		breakAllianceForWarRefusal(gs, callerRoot, playerRoot)
	}
	return Result{
		Accepted: false,
		Applied:  true,
		Message:  factionLabel(gs, callerRoot) + " tarafındaki savaş çağrısını reddettiniz. İttifak bozuldu.",
	}
}
