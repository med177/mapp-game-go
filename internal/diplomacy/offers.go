package diplomacy

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
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

// QueueSurrenderOffer kuşatma ile ilişkili teslimiyet teklifini kuyruğa ekler.
// Bölge kimliği teklifin hangi aktif kuşatmaya ait olduğunu ayırt eder.
func QueueSurrenderOffer(gs *state.GameState, from, to faction.FactionID, regionID world.RegionID, priority int, reason string) bool {
	if gs == nil || from == "" || to == "" || from == to || regionID == "" {
		return false
	}
	fromFaction := gs.Factions[from]
	toFaction := gs.Factions[to]
	if fromFaction == nil || toFaction == nil || fromFaction.IsEliminated || toFaction.IsEliminated {
		return false
	}
	target := gs.Regions[regionID]
	siege := gs.SiegeAt(regionID)
	if target == nil || target.IsSea || siege == nil {
		return false
	}
	attacker := gs.Armies[siege.AttackerArmyID]
	if attacker == nil || attacker.IsNaval || attacker.OwnerID != siege.AttackerFactionID || attacker.OwnerID == target.OwnerID || !IsWar(gs, faction.FactionID(attacker.OwnerID), faction.FactionID(target.OwnerID)) {
		return false
	}
	validDirection := (from == faction.FactionID(attacker.OwnerID) && to == faction.FactionID(target.OwnerID)) ||
		(from == faction.FactionID(target.OwnerID) && to == faction.FactionID(attacker.OwnerID))
	if !validDirection {
		return false
	}
	for _, offer := range gs.DiplomaticOffers {
		if offer.Action == string(ActionProposeSurrender) && offer.FromFactionID == from && offer.ToFactionID == to && offer.RegionID == regionID {
			return false
		}
	}
	if gs.DiplomaticOfferRegionRetryBlocked(string(from), string(to), string(ActionProposeSurrender), regionID, 1) {
		return false
	}
	gs.DiplomaticOffers = append(gs.DiplomaticOffers, state.DiplomaticOffer{
		FromFactionID:  from,
		ToFactionID:    to,
		Action:         string(ActionProposeSurrender),
		RegionID:       regionID,
		CreatedTurn:    gs.Turn,
		Priority:       priority,
		PriorityReason: reason,
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
		if !found || pendingOfferPrecedes(offer, gs.DiplomaticOffers[bestIdx], offer.Priority, bestPriority, offer.CreatedTurn, bestTurn, i, bestIdx) {
			bestIdx = i
			bestPriority = offer.Priority
			bestTurn = offer.CreatedTurn
			found = true
		}
	}
	return bestIdx, found
}

// pendingOfferPrecedes, bekleyen tekliflerdeki karar aşamasını uygular:
// barış, kuşatma teslimiyetinden; teslimiyet de diğer normal tekliflerden önce
// gösterilir. Aynı aşamadaki teklifler mevcut Priority sıralamasını korur.
func pendingOfferPrecedes(candidate, current state.DiplomaticOffer, candidatePriority, currentPriority, candidateTurn, currentTurn, candidateIndex, currentIndex int) bool {
	candidateStage := pendingOfferStage(candidate.Action)
	currentStage := pendingOfferStage(current.Action)
	if candidateStage != currentStage {
		return candidateStage > currentStage
	}
	return candidatePriority > currentPriority || (candidatePriority == currentPriority && (candidateTurn < currentTurn || (candidateTurn == currentTurn && candidateIndex < currentIndex)))
}

func pendingOfferStage(action string) int {
	switch Action(action) {
	case ActionProposePeace:
		return 2
	case ActionProposeSurrender:
		return 1
	default:
		return 0
	}
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
			Message:  factionLabel(gs, offer.FromFactionID) + " den gelen teklif reddedildi.",
		}
	}
	if action == ActionProposePeace {
		rel := EnsureRelation(gs, offer.FromFactionID, offer.ToFactionID)
		if rel.Stance != faction.StanceWar {
			return Result{Message: "Barış teklifi artık geçerli değil."}
		}
		settlement := AssessPeaceSettlement(gs, offer.FromFactionID, offer.ToFactionID)
		setPeaceBetweenCoalitions(gs, offer.FromFactionID, offer.ToFactionID)
		return Result{
			Accepted:   true,
			Applied:    true,
			Settlement: &settlement,
			Message:    factionLabel(gs, offer.ToFactionID) + " barışı kabul etti.",
		}
	}
	if action == ActionProposeAlliance {
		return resolveAcceptedAllianceOffer(gs, offer)
	}
	// Gönderen diplomasi kotasını teklif kuyruğa alınırken zaten harcadı.
	// Teklifin güncel koşullarını yeniden doğrula, ancak kabul sırasında aynı
	// teklif için kotayı ikinci kez tüketme.
	result := execute(gs, offer.FromFactionID, offer.ToFactionID, action, false)
	if accepted && !result.Applied {
		return Result{
			Accepted: false,
			Applied:  false,
			Message:  "Teklif koşulları değiştiği için uygulanamadı.",
		}
	}
	return result
}

// resolveAcceptedAllianceOffer, AI'nin daha önce oluşturduğu ittifak teklifini
// kabul eder. Teklif kuyruğa alındıktan sonra AI hazırlık akışındaki ortak tehdit
// veya stratejik durum değişiklikleri teklifin kabulündeki kararı yeniden zar
// hesabına sokmamalıdır. Sadece ilişkinin artık savaşta veya aynı realm'de olup
// olmadığı gibi teklifin temel geçerlilik koşulları yeniden kontrol edilir.
func resolveAcceptedAllianceOffer(gs *state.GameState, offer state.DiplomaticOffer) Result {
	if gs == nil || offer.FromFactionID == "" || offer.ToFactionID == "" || offer.FromFactionID == offer.ToFactionID {
		return Result{Message: "İttifak teklifi artık geçerli değil."}
	}
	actor := gs.Factions[offer.FromFactionID]
	target := gs.Factions[offer.ToFactionID]
	if actor == nil || target == nil || actor.IsEliminated || target.IsEliminated {
		return Result{Message: "İttifak teklifi artık geçerli değil."}
	}
	if sameRealm(gs, offer.FromFactionID, offer.ToFactionID) {
		return Result{Message: "İttifak teklifi artık geçerli değil."}
	}
	rel := EnsureRelation(gs, offer.FromFactionID, offer.ToFactionID)
	if rel.Stance == faction.StanceWar {
		return Result{Message: "İttifak teklifi artık geçerli değil."}
	}
	if rel.Stance == faction.StanceAllied {
		return Result{Accepted: true, Applied: true, Message: "Zaten müttefiksiniz."}
	}
	rel.Stance = faction.StanceAllied
	rel.Score = clamp(rel.Score+20, -100, 100)
	return Result{
		Accepted: true,
		Applied:  true,
		Message:  factionLabel(gs, offer.ToFactionID) + " ile ittifak kuruldu.",
	}
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
