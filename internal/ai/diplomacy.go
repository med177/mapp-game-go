package ai

import (
	"sort"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	aiRelationshipRepairChancePercent      = 60
	aiRelationshipRepairCooldownTurns      = 4
	aiNegotiationMinimumResourceReserve    = 10
	aiNegotiationMaxTransferSharePercent   = 25
	aiNegotiationMinimumMarketValuePercent = 90
	aiMercenaryPurchaseCostPercent         = 80
	// Hediye, ilişki bakımının pahalı ve seyrek kullanılan biçimidir. AI'nin
	// küçük/orta hazinesinin tek seferde büyük bölümünü tüketmesini engeller.
	aiGiftMaxTreasuryPercent = 25
)

// aiHandleDiplomacyWithSteps resolves AI peace, alliance and trade decisions.
// The public wrapper remains in ai.go; this file owns the diplomacy decision loop.
// Direct callers retain the legacy all-in-one behaviour; the real turn prelude
// disables relation spending and runs it after the economic priorities.
func aiHandleDiplomacyWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	aiHandleDiplomacyWithStepsMode(gs, fid, steps, true)
}

func aiHandleDiplomacyForTurn(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	aiHandleDiplomacyWithStepsMode(gs, fid, steps, false)
}

func aiHandleDiplomacyWithStepsMode(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep, allowRelationSpending bool) {
	gs.SyncWarLedgers()
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return
	}
	if diplomacy.DirectOverlord(gs, fid) != "" {
		return
	}
	aiOfferMinorPrivilegeWithSteps(gs, fid, steps)
	// Tarihsel genişleme hedefi karşısında tek başına yeterli olmayan devlet,
	// genel diplomasi taramasından önce aynı hedefe baskı yapabilecek müttefik
	// arar. Kabul edilen AI-AI ittifakı bu turdaki savaş koalisyonuna da girer.
	aiPursueHistoricalWarAlliance(gs, fid, steps)
	relationRepairUsed := false

	for _, otherID := range aiSortedFactionIDs(gs) {
		other := gs.Factions[otherID]
		if otherID == fid || other == nil || other.IsEliminated {
			continue
		}
		if overlord := diplomacy.DirectOverlord(gs, otherID); overlord != "" && overlord != fid {
			continue
		}

		rel := diplomacy.EnsureRelation(gs, fid, otherID)
		switch rel.Stance {
		case faction.StanceWar:
			if gs.DiplomacyOfferQuotaRemaining(fid) <= 0 {
				break
			}
			shouldProposePeace := false
			// Tüm senaryolar aynı savaş yorgunluğu, claim/core ve acil durum
			// değerlendirmesini kullanır. Eski güç/region karşılaştırması savaşı
			// ilk turda bitirebildiği için burada özellikle kaldırılmıştır.
			shouldProposePeace = diplomacy.AssessPeaceDesire(gs, fid, otherID).ShouldPropose()
			if shouldProposePeace {
				if !aiDiplomacyOfferRetryAllowed(gs, fid, otherID, diplomacy.ActionProposePeace) {
					continue
				}
				if otherID == gs.PlayerFactionID {
					priority, reason := aiDiplomacyOfferPriorityDetails(gs, fid, otherID, diplomacy.ActionProposePeace)
					if diplomacy.QueueOfferWithMeta(gs, fid, otherID, diplomacy.ActionProposePeace, priority, reason) {
						gs.MarkPeaceOffer(fid, otherID)
						addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: otherID, Message: turnFactionName(gs, fid) + " sana barış teklif ediyor."})
					}
				} else {
					result := diplomacy.ExecuteAIPeace(gs, fid, otherID)
					if !result.Applied {
						gs.MarkPeaceOffer(fid, otherID)
					}
					if result.Applied || result.Accepted {
						addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: otherID, Message: turnFactionName(gs, fid) + ": " + result.Message})
					}
				}
			}
		case faction.StancePeace:
			if gs.DiplomacyOfferQuotaRemaining(fid) <= 0 {
				break
			}
			if allowRelationSpending && !relationRepairUsed && aiHandleRelationshipRepairWithSteps(gs, fid, otherID, rel, steps) {
				relationRepairUsed = true
				continue
			}
			var allianceAssessment diplomacy.AllianceProposalAssessment
			if diplomacy.RelationScore(gs, fid, otherID) >= diplomacy.AllianceRelationThreshold(gs) {
				allianceAssessment = diplomacy.AssessAllianceProposal(gs, rel, fid, otherID)
			}
			if aiShouldAttemptAllianceOffer(gs, fid, otherID, allianceAssessment) && aiDiplomacyOfferRetryAllowed(gs, fid, otherID, diplomacy.ActionProposeAlliance) {
				if otherID == gs.PlayerFactionID {
					priority, reason := aiDiplomacyOfferPriorityDetails(gs, fid, otherID, diplomacy.ActionProposeAlliance)
					if diplomacy.QueueOfferWithMeta(gs, fid, otherID, diplomacy.ActionProposeAlliance, priority, reason) {
						addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: otherID, Message: turnFactionName(gs, fid) + " sana ittifak teklif ediyor."})
					}
				} else {
					result := diplomacy.Execute(gs, fid, otherID, diplomacy.ActionProposeAlliance)
					if result.Applied || result.Accepted {
						addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: otherID, Message: turnFactionName(gs, fid) + ": " + result.Message})
					}
				}
				continue
			}
			if diplomacy.RelationScore(gs, fid, otherID) >= 15 && diplomacy.Relation(gs, fid, otherID).Stance == faction.StancePeace && aiTradePartnerCount(gs, fid) < 3 && aiTradePartnerCount(gs, otherID) < 3 && !diplomacy.HasDirectThreat(gs, fid, otherID) && aiDiplomacyOfferRetryAllowed(gs, fid, otherID, diplomacy.ActionProposeTrade) {
				if otherID == gs.PlayerFactionID {
					assessment := diplomacy.AssessTradeProposal(gs, diplomacy.Relation(gs, fid, otherID), fid, otherID)
					if assessment.BlockReason == "" {
						priority, reason := aiDiplomacyOfferPriorityDetails(gs, fid, otherID, diplomacy.ActionProposeTrade)
						if diplomacy.QueueOfferWithMeta(gs, fid, otherID, diplomacy.ActionProposeTrade, priority, reason) {
							addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: otherID, Message: turnFactionName(gs, fid) + " sana ticaret teklif ediyor."})
						}
					}
				} else {
					result := diplomacy.Execute(gs, fid, otherID, diplomacy.ActionProposeTrade)
					if result.Applied || result.Accepted {
						addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: otherID, Message: turnFactionName(gs, fid) + ": " + result.Message})
					}
				}
			}
		case faction.StanceTrade:
			if allowRelationSpending && aiHandleRelationshipRepairWithSteps(gs, fid, otherID, rel, steps) {
				continue
			}
		case faction.StanceAllied:
			if allowRelationSpending && aiHandleRelationshipRepairWithSteps(gs, fid, otherID, rel, steps) {
				continue
			}
			if aiShouldCancelAlliance(gs, fid, otherID) {
				activeObjectiveConflict := false
				activeObjectiveConflict = diplomacy.AssessStrategicAlliance(gs, fid, otherID).ActiveObjectiveConflict
				result := diplomacy.Execute(gs, fid, otherID, diplomacy.ActionCancelAlliance)
				if result.Applied && activeObjectiveConflict {
					if current := diplomacy.Relation(gs, fid, otherID); current != nil && current.Stance == faction.StanceTrade {
						tradeResult := diplomacy.Execute(gs, fid, otherID, diplomacy.ActionCancelTrade)
						if tradeResult.Applied {
							result.Message += " Aktif stratejik hedef nedeniyle ticaret anlaşması da sona erdirildi."
						}
					}
				}
				if result.Applied || result.Accepted {
					addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: otherID, Message: turnFactionName(gs, fid) + ": " + result.Message})
				}
			}
		}
	}

	// Aynı savaşta oyuncuya hem barış hem de kuşatma teslimiyeti koşulu
	// oluşabiliyorsa önce barış teklifi kuyruğa alınır. Oyuncu barışı kabul
	// ederse diplomasi katmanı artık geçersiz kalan teslimiyet teklifini siler;
	// reddederse teslimiyet teklifi sonraki bekleyen teklif olur.
	aiHandleSiegeSurrenderOffersWithSteps(gs, fid, steps)

	aiEvaluateWarOpportunitiesWithSteps(gs, fid, steps)
}

type aiResourceNegotiationCandidate struct {
	target    faction.FactionID
	requested []state.DiplomaticTransfer
	offered   []state.DiplomaticTransfer
	chance    int
	netValue  int
}

// aiHandleResourceNegotiationWithSteps trades only to address a strategic
// resource shortfall. The AI retains its planned costs and safety reserve,
// while limiting each transfer to a quarter of the available surplus.
func aiHandleResourceNegotiationWithSteps(gs *state.GameState, fid faction.FactionID, demand economy.ResourceCost, steps *[]TurnStep) bool {
	if gs == nil || fid == "" || gs.DiplomacyOfferQuotaRemaining(fid) <= 0 {
		return false
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated || diplomacy.DirectOverlord(gs, fid) != "" {
		return false
	}
	if candidate, ok := aiBestStrategicRegionPurchase(gs, fid); ok {
		return aiSubmitStrategicTransfer(gs, fid, candidate, steps)
	}
	if candidate, ok := aiBestMercenaryPurchase(gs, fid); ok {
		return aiSubmitStrategicTransfer(gs, fid, candidate, steps)
	}
	if candidate, ok := aiBestMercenaryOffer(gs, fid); ok {
		return aiSubmitStrategicTransfer(gs, fid, candidate, steps)
	}

	best := aiResourceNegotiationCandidate{}
	for _, targetID := range aiSortedFactionIDs(gs) {
		if targetID == fid || diplomacy.SameRealm(gs, fid, targetID) {
			continue
		}
		target := gs.Factions[targetID]
		if target == nil || target.IsEliminated {
			continue
		}
		if overlord := diplomacy.DirectOverlord(gs, targetID); overlord != "" && overlord != fid {
			continue
		}
		if rel := diplomacy.Relation(gs, fid, targetID); rel != nil && rel.Stance == faction.StanceWar {
			continue
		}
		if !aiDiplomacyOfferRetryAllowed(gs, fid, targetID, diplomacy.ActionProposeTransfer) || aiTransferOfferPending(gs, fid, targetID) {
			continue
		}
		candidate, ok := aiBestResourceNegotiationForTarget(gs, fid, targetID, demand)
		if !ok {
			continue
		}
		if best.target == "" || candidate.netValue > best.netValue ||
			(candidate.netValue == best.netValue && candidate.chance > best.chance) {
			best = candidate
		}
	}
	if best.target == "" {
		return false
	}

	if best.target == gs.PlayerFactionID {
		if !diplomacy.QueueTransferOffer(gs, fid, best.target, best.requested, best.offered, 50, "stratejik kaynak açığını kapatma") {
			return false
		}
		addTurnStep(steps, TurnStep{
			FactionID:     fid,
			Kind:          TurnStepDiplomacy,
			TargetFaction: best.target,
			Message:       turnFactionName(gs, fid) + " kaynak açığını kapatmak için sana diplomatik pazarlık önerdi.",
		})
		return true
	}

	result := diplomacy.ExecuteTransferOffer(gs, fid, best.target, best.requested, best.offered)
	if result.Applied || result.Accepted {
		addTurnStep(steps, TurnStep{
			FactionID:     fid,
			Kind:          TurnStepDiplomacy,
			TargetFaction: best.target,
			Message:       turnFactionName(gs, fid) + " ile " + turnFactionName(gs, best.target) + " kaynak pazarlığı yaptı: " + result.Message,
		})
		return true
	}
	return false
}

type aiStrategicTransferCandidate struct {
	target    faction.FactionID
	regionID  world.RegionID
	requested []state.DiplomaticTransfer
	offered   []state.DiplomaticTransfer
	priority  int
	reason    string
}

func aiBestStrategicRegionPurchase(gs *state.GameState, fid faction.FactionID) (aiStrategicTransferCandidate, bool) {
	if gs == nil || fid == "" {
		return aiStrategicTransferCandidate{}, false
	}
	self := gs.Factions[fid]
	if self == nil {
		return aiStrategicTransferCandidate{}, false
	}
	var best aiStrategicTransferCandidate
	bestScore := -1
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.IsTerrainArea || region.OwnerID == "" || region.OwnerID == string(fid) {
			continue
		}
		claimValue := aiStrategicRegionClaimValue(gs, fid, region.ID)
		if claimValue <= 0 {
			continue
		}
		targetID := faction.FactionID(region.OwnerID)
		if !aiStrategicTradeTargetAllowed(gs, fid, targetID) {
			continue
		}
		rel := diplomacy.Relation(gs, fid, targetID)
		if rel != nil && rel.Stance == faction.StanceWar {
			continue
		}
		if !aiDiplomacyOfferRetryAllowed(gs, fid, targetID, diplomacy.ActionProposeTransfer) || aiTransferOfferPending(gs, fid, targetID) {
			continue
		}
		requested := []state.DiplomaticTransfer{{Kind: "region", ID: string(region.ID), Amount: 1}}
		regionValue := diplomacy.RegionTenTurnEconomicValue(gs, region.ID)
		if regionValue <= 0 {
			continue
		}
		if payment, ok := aiBestPaymentForTransferAtLeastPercent(gs, fid, targetID, fid, regionValue, requested, nil, 100); ok {
			score := claimValue*10000 + regionValue
			candidate := aiStrategicTransferCandidate{
				target:    targetID,
				regionID:  region.ID,
				requested: requested,
				offered:   payment,
				priority:  70 + claimValue,
				reason:    "stratejik bölge hedefi karşılığında adil ödeme",
			}
			if score > bestScore {
				best, bestScore = candidate, score
			}
		}
		for _, paymentRegion := range aiTradableRegions(gs, fid) {
			offered := []state.DiplomaticTransfer{{Kind: "region", ID: string(paymentRegion.ID), Amount: 1}}
			value := diplomacy.TransferListValue(gs, fid, offered)
			if !aiAssetValueAtLeastInNegotiationRange(value, regionValue) {
				continue
			}
			if accepted, _ := diplomacy.AssessTransferOffer(gs, fid, targetID, requested, offered); !accepted {
				continue
			}
			score := claimValue*10000 + regionValue - value
			candidate := aiStrategicTransferCandidate{
				target:    targetID,
				regionID:  region.ID,
				requested: requested,
				offered:   offered,
				priority:  70 + claimValue,
				reason:    "stratejik bölge için bölge takası",
			}
			if score > bestScore {
				best, bestScore = candidate, score
			}
		}
	}
	return best, best.target != ""
}

func aiBestMercenaryOffer(gs *state.GameState, fid faction.FactionID) (aiStrategicTransferCandidate, bool) {
	if gs == nil || fid == "" || aiFactionAtWar(gs, string(fid)) {
		return aiStrategicTransferCandidate{}, false
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return aiStrategicTransferCandidate{}, false
	}
	selfContext := prepareStrategicContext(gs, fid)
	if selfContext.CriticalThreat {
		return aiStrategicTransferCandidate{}, false
	}
	if plan := gs.AIPlans[fid]; plan != nil && plan.Kind == state.AIObjectiveExpand {
		return aiStrategicTransferCandidate{}, false
	}
	selfRequirement := aiForceRequirements(gs, fid, selfContext)
	selfReserve := maxInt(1, selfRequirement.LandPresent*75/100)
	surplus := selfRequirement.LandPresent - selfReserve
	if surplus <= 0 {
		return aiStrategicTransferCandidate{}, false
	}
	maxUnitsToSell := minInt(surplus, maxInt(1, selfRequirement.LandPresent/4))
	type unitCandidate struct {
		id    string
		value int
		count int
	}
	unitTypes := make([]string, 0, len(gs.UnitTypes))
	for id, unitType := range gs.UnitTypes {
		if unitType != nil && aiLandUnitCategory(unitType.Category) {
			unitTypes = append(unitTypes, id)
		}
	}
	sort.Strings(unitTypes)
	units := make([]unitCandidate, 0, len(unitTypes))
	for _, id := range unitTypes {
		count := aiTransferableLandUnitCount(gs, fid, id)
		if count <= 0 {
			continue
		}
		units = append(units, unitCandidate{id: id, count: count, value: diplomacy.TransferListValue(gs, fid, []state.DiplomaticTransfer{{Kind: "army", ID: id, Amount: 1}})})
	}
	if len(units) == 0 {
		return aiStrategicTransferCandidate{}, false
	}

	var best aiStrategicTransferCandidate
	bestScore := -1
	for _, targetID := range aiSortedFactionIDs(gs) {
		if !aiStrategicTradeTargetAllowed(gs, fid, targetID) || targetID == fid {
			continue
		}
		target := gs.Factions[targetID]
		if target == nil || target.IsEliminated || aiFactionAtWar(gs, string(targetID)) {
			continue
		}
		rel := diplomacy.Relation(gs, fid, targetID)
		if rel != nil && rel.Stance == faction.StanceWar {
			continue
		}
		if !aiDiplomacyOfferRetryAllowed(gs, fid, targetID, diplomacy.ActionProposeTransfer) || aiTransferOfferPending(gs, fid, targetID) {
			continue
		}
		targetContext := prepareStrategicContext(gs, targetID)
		targetRequirement := aiForceRequirements(gs, targetID, targetContext)
		need := targetRequirement.LandTarget - targetRequirement.LandPresent - targetRequirement.LandPending
		if need <= 0 {
			continue
		}
		for _, unit := range units {
			amount := minInt(unit.count, minInt(need, maxUnitsToSell))
			if amount <= 0 {
				continue
			}
			offered := []state.DiplomaticTransfer{{Kind: "army", ID: unit.id, Amount: amount}}
			value := diplomacy.TransferListValue(gs, fid, offered)
			if value <= 0 {
				continue
			}
			if payment, ok := aiBestPaymentForTransfer(gs, fid, targetID, targetID, value, nil, offered); ok {
				score := minInt(need, amount)*100 + unit.value
				candidate := aiStrategicTransferCandidate{
					target:    targetID,
					requested: payment,
					offered:   offered,
					priority:  60 + minInt(need, 20),
					reason:    "askerî kuvvet açığı için paralı asker teklifi",
				}
				if score > bestScore {
					best, bestScore = candidate, score
				}
			}
		}
	}
	return best, best.target != ""
}

func aiBestMercenaryPurchase(gs *state.GameState, fid faction.FactionID) (aiStrategicTransferCandidate, bool) {
	if gs == nil || fid == "" {
		return aiStrategicTransferCandidate{}, false
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return aiStrategicTransferCandidate{}, false
	}
	selfContext := prepareStrategicContext(gs, fid)
	selfRequirement := aiForceRequirements(gs, fid, selfContext)
	need := selfRequirement.LandTarget - selfRequirement.LandPresent - selfRequirement.LandPending
	if need <= 0 {
		return aiStrategicTransferCandidate{}, false
	}
	type unitCandidate struct {
		id    string
		value int
		count int
	}
	unitTypes := make([]string, 0, len(gs.UnitTypes))
	for id, unitType := range gs.UnitTypes {
		if unitType != nil && aiLandUnitCategory(unitType.Category) {
			unitTypes = append(unitTypes, id)
		}
	}
	sort.Strings(unitTypes)
	var best aiStrategicTransferCandidate
	bestScore := -1
	for _, targetID := range aiSortedFactionIDs(gs) {
		if targetID == fid || !aiStrategicTradeTargetAllowed(gs, fid, targetID) {
			continue
		}
		target := gs.Factions[targetID]
		if target == nil || aiFactionAtWar(gs, string(targetID)) {
			continue
		}
		rel := diplomacy.Relation(gs, fid, targetID)
		if rel != nil && rel.Stance == faction.StanceWar {
			continue
		}
		if !aiDiplomacyOfferRetryAllowed(gs, fid, targetID, diplomacy.ActionProposeTransfer) || aiTransferOfferPending(gs, fid, targetID) {
			continue
		}
		targetContext := prepareStrategicContext(gs, targetID)
		if targetContext.CriticalThreat {
			continue
		}
		if plan := gs.AIPlans[targetID]; plan != nil && plan.Kind == state.AIObjectiveExpand {
			continue
		}
		targetRequirement := aiForceRequirements(gs, targetID, targetContext)
		targetReserve := maxInt(1, targetRequirement.LandPresent*75/100)
		targetSurplus := targetRequirement.LandPresent - targetReserve
		maxUnitsToBuy := minInt(targetSurplus, maxInt(1, targetRequirement.LandPresent/4))
		if maxUnitsToBuy <= 0 {
			continue
		}
		for _, unitTypeID := range unitTypes {
			count := aiTransferableLandUnitCount(gs, targetID, unitTypeID)
			amount := minInt(need, minInt(count, maxUnitsToBuy))
			if amount <= 0 {
				continue
			}
			unitType := gs.UnitTypes[unitTypeID]
			unitProductionCost := aiGoldEquivalentCost(gs, aiUnitResourceCost(unitType))
			if unitProductionCost <= 0 {
				continue
			}
			requested := []state.DiplomaticTransfer{{Kind: "army", ID: unitTypeID, Amount: amount}}
			value := diplomacy.TransferListValue(gs, targetID, requested)
			if value <= 0 {
				continue
			}
			maximumPayment := unitProductionCost * amount * aiMercenaryPurchaseCostPercent / 100
			if payment, ok := aiBestMercenaryPurchasePayment(gs, fid, targetID, requested, maximumPayment); ok {
				score := amount*100 + diplomacy.RelationScore(gs, fid, targetID)
				candidate := aiStrategicTransferCandidate{
					target:    targetID,
					requested: requested,
					offered:   payment,
					priority:  65 + minInt(need, 20),
					reason:    "askerî kuvvet açığını kapatmak için paralı asker satın alma",
				}
				if score > bestScore {
					best, bestScore = candidate, score
				}
			}
		}
	}
	return best, best.target != ""
}

func aiBestMercenaryPurchasePayment(gs *state.GameState, buyer, seller faction.FactionID, requested []state.DiplomaticTransfer, maximumValue int) ([]state.DiplomaticTransfer, bool) {
	if gs == nil || buyer == "" || seller == "" || maximumValue <= 0 || len(requested) == 0 {
		return nil, false
	}
	var best []state.DiplomaticTransfer
	bestValue := 0
	for _, kind := range economy.CostResourceKinds() {
		price := aiResourceMarketValue(gs, kind)
		if price <= 0 {
			continue
		}
		available := economy.FactionResourceAmount(gs.Factions[buyer], kind) - aiNegotiationPaymentReserve(gs, buyer, kind)
		maxAmount := minInt(available*aiNegotiationMaxTransferSharePercent/100, maximumValue/price)
		if maxAmount <= 0 {
			continue
		}
		payment := []state.DiplomaticTransfer{{Kind: "resource", ID: string(kind), Amount: maxAmount}}
		accepted, _ := diplomacy.AssessTransferOffer(gs, buyer, seller, requested, payment)
		paymentValue := diplomacy.TransferListValue(gs, buyer, payment)
		if !accepted || paymentValue > maximumValue || paymentValue <= bestValue {
			continue
		}
		best = payment
		bestValue = paymentValue
	}
	return best, best != nil
}

func aiSubmitStrategicTransfer(gs *state.GameState, fid faction.FactionID, candidate aiStrategicTransferCandidate, steps *[]TurnStep) bool {
	if candidate.target == "" {
		return false
	}
	message := turnFactionName(gs, fid) + " " + candidate.reason + " teklif etti."
	if candidate.regionID != "" {
		message = turnFactionName(gs, fid) + " " + turnRegionName(gs, candidate.regionID) + " bölgesini stratejik hedefi olarak istiyor ve karşılığında takas teklif etti."
	}
	if candidate.target == gs.PlayerFactionID {
		if !diplomacy.QueueTransferOffer(gs, fid, candidate.target, candidate.requested, candidate.offered, candidate.priority, candidate.reason) {
			return false
		}
		addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: candidate.target, TargetRegion: candidate.regionID, FocusRegion: candidate.regionID, Message: message})
		return true
	}
	result := diplomacy.ExecuteTransferOffer(gs, fid, candidate.target, candidate.requested, candidate.offered)
	if result.Message != "" {
		message += " " + result.Message
	}
	addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: candidate.target, TargetRegion: candidate.regionID, FocusRegion: candidate.regionID, Message: message})
	return true
}

func aiBestPaymentForTransfer(gs *state.GameState, from, to, payer faction.FactionID, assetValue int, requestedAsset, offeredAsset []state.DiplomaticTransfer) ([]state.DiplomaticTransfer, bool) {
	return aiBestPaymentForTransferAtLeastPercent(gs, from, to, payer, assetValue, requestedAsset, offeredAsset, aiNegotiationMinimumMarketValuePercent)
}

func aiBestPaymentForTransferAtLeastPercent(gs *state.GameState, from, to, payer faction.FactionID, assetValue int, requestedAsset, offeredAsset []state.DiplomaticTransfer, minimumValuePercent int) ([]state.DiplomaticTransfer, bool) {
	if assetValue <= 0 || (requestedAsset == nil && offeredAsset == nil) {
		return nil, false
	}
	var best []state.DiplomaticTransfer
	bestValue := int(^uint(0) >> 1)
	for _, kind := range economy.CostResourceKinds() {
		price := aiResourceMarketValue(gs, kind)
		if price <= 0 {
			continue
		}
		available := economy.FactionResourceAmount(gs.Factions[payer], kind) - aiNegotiationPaymentReserve(gs, payer, kind)
		maxAmount := available * aiNegotiationMaxTransferSharePercent / 100
		minAmount := (assetValue*minimumValuePercent + 100*price - 1) / (100 * price)
		maxValue := assetValue * 125 / 100
		maxMarketAmount := maxValue / price
		if maxAmount > maxMarketAmount {
			maxAmount = maxMarketAmount
		}
		if minAmount <= 0 || minAmount > maxAmount {
			continue
		}
		low, high := minAmount, maxAmount
		for low < high {
			mid := low + (high-low)/2
			payment := []state.DiplomaticTransfer{{Kind: "resource", ID: string(kind), Amount: mid}}
			requested, offered := payment, offeredAsset
			if payer == from {
				requested, offered = requestedAsset, payment
			}
			accepted, _ := diplomacy.AssessTransferOffer(gs, from, to, requested, offered)
			if accepted {
				high = mid
			} else {
				low = mid + 1
			}
		}
		payment := []state.DiplomaticTransfer{{Kind: "resource", ID: string(kind), Amount: low}}
		requested, offered := payment, offeredAsset
		if payer == from {
			requested, offered = requestedAsset, payment
		}
		accepted, _ := diplomacy.AssessTransferOffer(gs, from, to, requested, offered)
		value := diplomacy.TransferListValue(gs, payer, payment)
		if !accepted || value*100 < assetValue*minimumValuePercent || value*100 > assetValue*125 || value >= bestValue {
			continue
		}
		best = payment
		bestValue = value
	}
	return best, best != nil
}

func aiNegotiationPaymentReserve(gs *state.GameState, payer faction.FactionID, kind economy.ResourceKind) int {
	if kind == economy.ResourceGold {
		return aiMinGoldReserve
	}
	if kind == economy.ResourceGrain {
		return maxInt(100, aiFactionGrainDemand(gs, payer)*aiGrainReserveMonths)
	}
	return aiNegotiationMinimumResourceReserve
}

func aiAssetValueInNegotiationRange(paymentValue, assetValue int) bool {
	return paymentValue*100 >= assetValue*aiNegotiationMinimumMarketValuePercent && paymentValue*100 <= assetValue*125
}

func aiAssetValueAtLeastInNegotiationRange(paymentValue, assetValue int) bool {
	return paymentValue >= assetValue && paymentValue*100 <= assetValue*125
}

func aiStrategicTradeTargetAllowed(gs *state.GameState, fid, targetID faction.FactionID) bool {
	if gs == nil || fid == "" || targetID == "" || fid == targetID || diplomacy.SameRealm(gs, fid, targetID) {
		return false
	}
	target := gs.Factions[targetID]
	if target == nil || target.IsEliminated {
		return false
	}
	if overlord := diplomacy.DirectOverlord(gs, targetID); overlord != "" && overlord != fid {
		return false
	}
	return diplomacy.DirectOverlord(gs, fid) == ""
}

func aiStrategicRegionClaimValue(gs *state.GameState, fid faction.FactionID, regionID world.RegionID) int {
	self := gs.Factions[fid]
	if self == nil || regionID == "" {
		return 0
	}
	value := 0
	for _, claim := range self.TerritorialClaims {
		if claim.RegionID == string(regionID) && claim.Value > value {
			value = claim.Value
		}
	}
	if plan := gs.AIPlans[fid]; plan != nil && plan.Kind == state.AIObjectiveExpand {
		for index, targetRegionID := range plan.TargetRegionIDs {
			if targetRegionID == regionID {
				value = maxInt(value, maxInt(30, 100-index*5))
				break
			}
		}
	}
	return value
}

func aiTradableRegions(gs *state.GameState, owner faction.FactionID) []*world.Region {
	regions := make([]*world.Region, 0)
	self := gs.Factions[owner]
	if self == nil || len(gs.LandRegionsOwnedBy(owner)) <= 1 {
		return regions
	}
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.IsTerrainArea || region.OwnerID != string(owner) || aiRegionIsCapital(gs, self, region) {
			continue
		}
		claimed := false
		for _, claim := range self.TerritorialClaims {
			if claim.RegionID == string(region.ID) {
				claimed = true
				break
			}
		}
		if !claimed {
			regions = append(regions, region)
		}
	}
	return regions
}

func aiRegionIsCapital(gs *state.GameState, owner *faction.Faction, region *world.Region) bool {
	if gs == nil || owner == nil || region == nil || owner.CapitalSettlementID == "" {
		return false
	}
	capitalRegion, _, _, ok := gs.FindSettlementByID(owner.CapitalSettlementID)
	return ok && capitalRegion != nil && capitalRegion.ID == region.ID
}

func aiTransferableLandUnitCount(gs *state.GameState, owner faction.FactionID, unitTypeID string) int {
	count := 0
	for _, current := range aiSortedArmies(gs) {
		if current == nil || current.OwnerID != string(owner) || current.IsNaval || current.IsGarrison || current.Commander != nil || current.EmbarkedCommander != nil {
			continue
		}
		for _, unit := range current.Units {
			if unit.TypeID == unitTypeID {
				count++
			}
		}
	}
	return count
}

func aiBestResourceNegotiationForTarget(gs *state.GameState, fid, targetID faction.FactionID, demand economy.ResourceCost) (aiResourceNegotiationCandidate, bool) {
	self := gs.Factions[fid]
	target := gs.Factions[targetID]
	if self == nil || target == nil {
		return aiResourceNegotiationCandidate{}, false
	}

	best := aiResourceNegotiationCandidate{}
	for _, requestedKind := range economy.CostResourceKinds() {
		required := demand.Amount(requestedKind)
		if requestedKind == economy.ResourceGrain {
			required = maxInt(required, maxInt(100, aiFactionGrainDemand(gs, fid)*aiGrainReserveMonths))
		}
		if requestedKind == economy.ResourceGold {
			required = maxInt(required, aiMinGoldReserve)
		}
		shortfall := required - economy.FactionResourceAmount(self, requestedKind)
		if shortfall <= 0 {
			continue
		}

		targetAmount := economy.FactionResourceAmount(target, requestedKind)
		requestLimit := minInt(shortfall, targetAmount*aiNegotiationMaxTransferSharePercent/100)
		if requestLimit <= 0 {
			continue
		}
		requested := []state.DiplomaticTransfer{{Kind: "resource", ID: string(requestedKind), Amount: requestLimit}}
		requestedValue := diplomacy.TransferListValue(gs, targetID, requested)
		if requestedValue <= 0 {
			continue
		}

		for _, offeredKind := range economy.CostResourceKinds() {
			if offeredKind == requestedKind {
				continue
			}
			targetNeed := aiNegotiationTargetResourceNeed(gs, targetID, offeredKind)
			if targetNeed <= 0 {
				continue
			}
			reserve := maxInt(demand.Amount(offeredKind), aiNegotiationMinimumResourceReserve)
			if offeredKind == economy.ResourceGold {
				reserve = maxInt(reserve, aiMinGoldReserve)
			}
			if offeredKind == economy.ResourceGrain {
				reserve = maxInt(reserve, maxInt(100, aiFactionGrainDemand(gs, fid)*aiGrainReserveMonths))
			}
			surplus := economy.FactionResourceAmount(self, offeredKind) - reserve
			maxOffer := minInt(surplus*aiNegotiationMaxTransferSharePercent/100, targetNeed)
			if maxOffer <= 0 {
				continue
			}

			offeredUnitValue := aiResourceMarketValue(gs, offeredKind)
			if offeredUnitValue <= 0 {
				continue
			}
			minimumOfferValue := (requestedValue*aiNegotiationMinimumMarketValuePercent + 99) / 100
			minimumOfferAmount := (minimumOfferValue + offeredUnitValue - 1) / offeredUnitValue
			maximumMarketOfferAmount := requestedValue / offeredUnitValue
			if minimumOfferAmount > maximumMarketOfferAmount {
				continue
			}
			if maximumMarketOfferAmount < maxOffer {
				maxOffer = maximumMarketOfferAmount
			}
			if maxOffer < minimumOfferAmount {
				continue
			}

			offered := []state.DiplomaticTransfer{{Kind: "resource", ID: string(offeredKind)}}
			low, high := minimumOfferAmount, maxOffer
			for low < high {
				mid := low + (high-low)/2
				offered[0].Amount = mid
				accepted, _ := diplomacy.AssessTransferOffer(gs, fid, targetID, requested, offered)
				if accepted {
					high = mid
				} else {
					low = mid + 1
				}
			}
			offered[0].Amount = low
			accepted, chance := diplomacy.AssessTransferOffer(gs, fid, targetID, requested, offered)
			if !accepted {
				continue
			}
			offeredValue := diplomacy.TransferListValue(gs, fid, offered)
			if offeredValue*100 < requestedValue*aiNegotiationMinimumMarketValuePercent || offeredValue > requestedValue {
				continue
			}

			candidate := aiResourceNegotiationCandidate{
				target:    targetID,
				requested: append([]state.DiplomaticTransfer(nil), requested...),
				offered:   append([]state.DiplomaticTransfer(nil), offered...),
				chance:    chance,
				netValue:  requestedValue - offeredValue,
			}
			if best.target == "" || candidate.netValue > best.netValue ||
				(candidate.netValue == best.netValue && candidate.chance > best.chance) {
				best = candidate
			}
		}
	}
	return best, best.target != ""
}

func aiNegotiationTargetResourceNeed(gs *state.GameState, targetID faction.FactionID, kind economy.ResourceKind) int {
	target := gs.Factions[targetID]
	if target == nil {
		return 0
	}
	stock := economy.FactionResourceAmount(target, kind)
	switch kind {
	case economy.ResourceGold:
		return maxInt(0, aiMinGoldReserve-stock)
	case economy.ResourceGrain:
		reserve := maxInt(100, aiFactionGrainDemand(gs, targetID)*aiGrainReserveMonths)
		return maxInt(0, reserve-stock)
	default:
		good := economy.ResourceDefByKind(kind).TradeGood
		return gs.MarketBuyOrder(targetID, good, aiResourcePrice(gs, good))
	}
}

func aiResourceMarketValue(gs *state.GameState, kind economy.ResourceKind) int {
	if kind == economy.ResourceGold {
		return 1
	}
	good := economy.ResourceDefByKind(kind).TradeGood
	return aiResourcePrice(gs, good)
}

func aiTransferOfferPending(gs *state.GameState, from, to faction.FactionID) bool {
	if gs == nil {
		return false
	}
	for _, offer := range gs.DiplomaticOffers {
		if offer.Action == string(diplomacy.ActionProposeTransfer) && offer.FromFactionID == from && offer.ToFactionID == to {
			return true
		}
	}
	return false
}

// aiOfferMinorPrivilegeWithSteps, egemen AI'nin henüz imtiyaz verilmemiş bir
// minor bölgeyi dış bir devlete işletme hakkı olarak önermesini sağlar. Hedef
// oyuncuysa teklif kuyruğa bırakılır; AI hedefi aynı turda ekonomik faydayı
// değerlendirerek kabul veya reddeder.
func aiOfferMinorPrivilegeWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) bool {
	if gs == nil || fid == "" || gs.DiplomacyOfferQuotaRemaining(fid) <= 0 {
		return false
	}
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || !region.IsMinorRegion || region.IsPrivileged || region.OwnerID != string(fid) || gs.SovereignOwnerID(region) != string(fid) {
			continue
		}
		for _, targetID := range aiSortedFactionIDs(gs) {
			if targetID == fid {
				continue
			}
			target := gs.Factions[targetID]
			if target == nil || target.IsEliminated {
				continue
			}
			assessment := diplomacy.AssessMinorPrivilegeOffer(gs, fid, targetID, region.ID)
			if assessment.BlockReason != "" || assessment.Chance < 45 {
				continue
			}
			if !diplomacy.QueueMinorPrivilegeOffer(gs, fid, targetID, region.ID, 30+assessment.Chance, "ekonomik imtiyaz geliri ve diplomatik yakınlaşma") {
				continue
			}
			message := turnFactionName(gs, fid) + " " + turnRegionName(gs, region.ID) + " için " + turnFactionName(gs, targetID) + " devletine imtiyaz teklif etti."
			if targetID == gs.PlayerFactionID {
				addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: targetID, TargetRegion: region.ID, FocusRegion: region.ID, Message: message})
				return true
			}
			for index, offer := range gs.DiplomaticOffers {
				if offer.Action != string(diplomacy.ActionOfferMinorPrivilege) || offer.FromFactionID != fid || offer.ToFactionID != targetID || offer.RegionID != region.ID {
					continue
				}
				result := diplomacy.ResolveOffer(gs, index, ShouldAcceptMinorPrivilegeOffer(gs, fid, targetID, region.ID, assessment))
				if result.Applied || result.Accepted {
					message += " " + result.Message
				}
				addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: targetID, TargetRegion: region.ID, FocusRegion: region.ID, Message: message})
				return result.Applied
			}
		}
	}
	return false
}

func aiPursueHistoricalWarAlliance(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	if gs == nil || fid == "" || gs.DiplomacyOfferQuotaRemaining(fid) <= 0 {
		return
	}
	plan := gs.AIPlans[fid]
	if plan == nil || plan.Kind != state.AIObjectiveExpand || plan.TargetFactionID == "" || plan.TargetFactionID == fid {
		return
	}
	target := plan.TargetFactionID
	targetRelation := diplomacy.Relation(gs, fid, target)
	if targetRelation == nil || targetRelation.Stance != faction.StancePeace {
		return
	}
	coalition := aiWarCoalitionAssessment(gs, fid, target)
	if coalition.DefenderPower <= 0 || coalition.AttackerPower*100 >= coalition.DefenderPower*aiMinAttackPowerPercent(gs) {
		return
	}
	battlefield := aiWarBattlefieldRegions(gs, diplomacy.RealmRoot(gs, target))
	if len(battlefield) == 0 {
		battlefield = aiWarBattlefieldRegions(gs, target)
	}

	best := faction.FactionID("")
	bestPower := 0
	for _, candidateID := range aiSortedFactionIDs(gs) {
		if candidateID == fid || candidateID == target || diplomacy.SameRealm(gs, fid, candidateID) || diplomacy.SameRealm(gs, target, candidateID) {
			continue
		}
		candidate := gs.Factions[candidateID]
		if candidate == nil || candidate.IsEliminated || diplomacy.DirectOverlord(gs, candidateID) != "" {
			continue
		}
		rel := diplomacy.Relation(gs, fid, candidateID)
		if rel == nil || rel.Stance != faction.StancePeace || diplomacy.RelationScore(gs, fid, candidateID) < diplomacy.AllianceRelationThreshold(gs) || !aiDiplomacyOfferRetryAllowed(gs, fid, candidateID, diplomacy.ActionProposeAlliance) {
			continue
		}
		// Hedefin mevcut müttefikiyle hedefe karşı ittifak aranmaz. Adayın
		// hedefle sınırı veya aynı genişleme planı olmalı ki yardım gerçek olsun.
		if targetRel := diplomacy.Relation(gs, candidateID, target); targetRel != nil && targetRel.Stance == faction.StanceAllied {
			continue
		}
		if !diplomacy.HasDirectThreat(gs, candidateID, target) && !aiPlanTargetsFaction(gs, candidateID, target) {
			continue
		}
		assessment := diplomacy.AssessAllianceProposal(gs, rel, fid, candidateID)
		if assessment.BlockReason != "" || assessment.Chance < 45 {
			continue
		}
		power := aiWarWeightedFactionPowerAsSeenBy(gs, fid, candidateID, battlefield)
		if power > bestPower || (power == bestPower && power > 0 && (best == "" || candidateID < best)) {
			best, bestPower = candidateID, power
		}
	}
	if best == "" {
		return
	}
	if best == gs.PlayerFactionID {
		priority, reason := aiDiplomacyOfferPriorityDetails(gs, fid, best, diplomacy.ActionProposeAlliance)
		if diplomacy.QueueOfferWithMeta(gs, fid, best, diplomacy.ActionProposeAlliance, priority+30, "tarihsel hedef için ortak savaş hazırlığı: "+reason) {
			addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: best, Message: turnFactionName(gs, fid) + " " + turnFactionName(gs, best) + " ile tarihsel hedef için ittifak arıyor."})
		}
		return
	}
	result := diplomacy.Execute(gs, fid, best, diplomacy.ActionProposeAlliance)
	if result.Applied || result.Accepted {
		addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: best, Message: turnFactionName(gs, fid) + " " + turnFactionName(gs, best) + " ile " + turnFactionName(gs, target) + " hedefi için ittifak kurdu."})
	}
}

// aiHandleRelationshipRepairWithSteps, AI'nin yalnız kendi stratejisine katkı
// sağlayan barışçıl ilişkilerde kullandığı tek taraflı ilişki aksiyonlarını
// uygular. AI-AI işlemleri hemen çözülür; oyuncuya giden işlemler ise oyuncunun
// barış tekliflerinde gördüğü aynı modal kuyruğuna girer.
func aiHandleRelationshipRepairWithSteps(gs *state.GameState, fid, otherID faction.FactionID, rel *faction.Relation, _ *[]TurnStep) bool {
	return aiHandleRelationshipRepairWithBudget(gs, fid, otherID, rel, nil)
}

func aiHandleRelationshipRepairWithBudget(gs *state.GameState, fid, otherID faction.FactionID, rel *faction.Relation, budget *aiBudget) bool {
	return aiHandleRelationshipRepairWithBudgetAndThreat(gs, fid, otherID, rel, budget, false, false)
}

func aiHandleRelationshipRepairWithBudgetAndThreat(gs *state.GameState, fid, otherID faction.FactionID, rel *faction.Relation, budget *aiBudget, sharedThreat bool, sharedThreatReady bool) bool {
	if gs == nil || rel == nil || gs.Turn < rel.NextAIRelationRepairTurn {
		return false
	}
	action, reason, ok := aiRelationshipRepairActionWithThreat(gs, fid, otherID, rel, sharedThreat, sharedThreatReady)
	if !ok || gs.DiplomacyOfferQuotaRemaining(fid) <= 0 {
		return false
	}
	self := gs.Factions[fid]
	if self == nil {
		return false
	}
	cost := aiRelationshipActionCost(gs, action)
	if action == diplomacy.ActionSendGift && !aiGiftTreasurySafe(self, cost.Gold, budget) {
		// Hediye güvenli değilse, aynı ilişki hedefi için daha ucuz heyeti dene.
		action = diplomacy.ActionImproveRelations
		cost = aiRelationshipActionCost(gs, action)
	}
	if !aiCanAffordForBudget(self, cost, budget, aiBudgetEconomy) {
		return false
	}
	// İlişki onarımı uygun ve karşılanabilir olsa bile her tur otomatikleşmesin;
	// aynı deterministik tur/faction/hedef zarı save ve replay akışını korur.
	if aiDiplomacyOfferRoll(gs, fid, otherID, action) >= aiRelationshipRepairChancePercent {
		return false
	}

	if otherID == gs.PlayerFactionID {
		priority, _ := aiDiplomacyOfferPriorityDetails(gs, fid, otherID, action)
		if !diplomacy.QueueOfferWithMeta(gs, fid, otherID, action, priority+20, reason) {
			return false
		}
		rel.NextAIRelationRepairTurn = gs.Turn + aiRelationshipRepairCooldownTurns
		if budget != nil {
			budget.consume(aiBudgetEconomy, cost.Gold)
		}
		return true
	}

	result := diplomacy.Execute(gs, fid, otherID, action)
	if !result.Applied {
		return false
	}
	rel.NextAIRelationRepairTurn = gs.Turn + aiRelationshipRepairCooldownTurns
	if budget != nil {
		budget.consume(aiBudgetEconomy, cost.Gold)
	}
	return true
}

// aiGiftTreasurySafe, hediyenin yalnızca mevcut bütçe rezervi içinde değil,
// hazine ölçeğine göre de makul kalmasını sağlar. Böylece başlangıçta 5.000
// altını olan bir devlet 2.500 altınlık hediyeyi, yalnızca bu tutarı teknik
// olarak karşılayabildiği için göndermez.
func aiGiftTreasurySafe(self *faction.Faction, giftCost int, budget *aiBudget) bool {
	if self == nil || giftCost <= 0 || self.Gold <= 0 {
		return false
	}
	if giftCost > self.Gold*aiGiftMaxTreasuryPercent/100 {
		return false
	}

	reserve := aiMinGoldReserve
	if budget != nil && budget.EmergencyGold > reserve {
		reserve = budget.EmergencyGold
	}
	// Hediye sonrasında bir sonraki aynı büyüklükteki diplomatik harcamayı
	// ve acil rezervi karşılayacak alan kalsın.
	return self.Gold-giftCost >= reserve+giftCost
}

// aiHandleRelationshipRepairsAfterBudget keeps gifts and envoys as the last
// treasury-funded AI action. The 1300 budget exposes only the leftover
// FlexibleGold after every priority category has been released.
func aiHandleRelationshipRepairsAfterBudget(gs *state.GameState, fid faction.FactionID, budget *aiBudget) {
	if gs == nil || fid == "" || gs.DiplomacyOfferQuotaRemaining(fid) <= 0 {
		return
	}
	sharedThreats := diplomacy.SharedMajorThreats(gs, fid)
	for _, otherID := range aiSortedFactionIDs(gs) {
		if gs.DiplomacyOfferQuotaRemaining(fid) <= 0 {
			return
		}
		other := gs.Factions[otherID]
		if otherID == fid || other == nil || other.IsEliminated {
			continue
		}
		if overlord := diplomacy.DirectOverlord(gs, otherID); overlord != "" && overlord != fid {
			continue
		}
		rel := diplomacy.EnsureRelation(gs, fid, otherID)
		if aiHandleRelationshipRepairWithBudgetAndThreat(gs, fid, otherID, rel, budget, sharedThreats[otherID], true) {
			return
		}
	}
}

func aiRelationshipActionCost(gs *state.GameState, action diplomacy.Action) economy.ResourceCost {
	if action == diplomacy.ActionSendGift {
		return economy.ResourceCost{Gold: diplomacy.GiftGoldCostFor(gs)}
	}
	return economy.ResourceCost{Gold: diplomacy.RelationImprovementGoldCostFor(gs)}
}

// aiRelationshipRepairAction, ilişki aksiyonunu yalnızca somut ticari veya
// güvenlik çıkarı varsa seçer. Önce ucuz heyetle ticaret eşiğine ulaşır; daha
// yüksek ilişki hedefi gerekiyorsa ve altın rezervi uygunsa hediye kullanır.
func aiRelationshipRepairAction(gs *state.GameState, fid, otherID faction.FactionID, rel *faction.Relation) (diplomacy.Action, string, bool) {
	return aiRelationshipRepairActionWithThreat(gs, fid, otherID, rel, false, false)
}

func aiRelationshipRepairActionWithThreat(gs *state.GameState, fid, otherID faction.FactionID, rel *faction.Relation, sharedThreat, sharedThreatReady bool) (diplomacy.Action, string, bool) {
	if gs == nil || rel == nil || fid == "" || otherID == "" || fid == otherID || rel.Stance == faction.StanceWar || diplomacy.SameRealm(gs, fid, otherID) {
		return "", "", false
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return "", "", false
	}

	strategicTarget := aiIsStrategicDiplomacyTarget(gs, fid, otherID)
	hasActiveTrade := diplomacy.HasTradeRouteBetween(gs, fid, otherID)
	hasLandBorder := diplomacy.SharesLandBorder(gs, fid, otherID)
	commonEnemy := diplomacy.HasCommonEnemy(gs, fid, otherID)
	if !sharedThreatReady {
		sharedThreat = diplomacy.HasSharedMajorThreat(gs, fid, otherID)
	}
	directThreat := false
	if hasLandBorder {
		directThreat = diplomacy.HasDirectThreat(gs, fid, otherID)
	}
	// İlişki onarımı öncelikle AI'nin yakın çevresine yönelir. Sadece uzak bir
	// deniz ticareti ihtimali veya genel ittifak puanı, tek başına heyet/hediye
	// göndermek için yeterli değildir; mevcut ticaret ya da gerçek güvenlik
	// bağlantıları uzak hedefleri yine meşru kılar.
	if !hasLandBorder && !hasActiveTrade && !commonEnemy && !sharedThreat && !directThreat {
		return "", "", false
	}
	hasAllianceInterest := aiAllianceHasMeaningfulBenefitWithThreats(gs, fid, otherID, commonEnemy, sharedThreat)
	hasTradeInterest := hasActiveTrade || diplomacy.CanEstablishTradeRoute(gs, fid, otherID)
	// AI stratejik hedefinden vazgeçmiş değildir; ancak hedef sınırında askeri
	// olarak müşkül durumdaysa zaman kazanmak için ucuz heyet kullanabilir.
	// Hediye daha sonra seçilmesin diye bu istisna aşağıda doğrudan heyete
	// yönlendirilir.
	if strategicTarget && !directThreat {
		return "", "", false
	}
	if !hasTradeInterest && !hasAllianceInterest && !commonEnemy && !sharedThreat && !directThreat {
		return "", "", false
	}

	desiredScore := 15
	if hasActiveTrade {
		desiredScore = 30
	}
	if hasAllianceInterest || commonEnemy || sharedThreat || directThreat {
		desiredScore = maxInt(desiredScore, diplomacy.AllianceRelationThreshold(gs))
	}
	if diplomacy.RelationScore(gs, self.ID, otherID) >= desiredScore {
		return "", "", false
	}

	reason := "ticaret hattını güvenceye alma"
	if hasAllianceInterest || commonEnemy || sharedThreat {
		reason = "stratejik ortaklık hazırlığı"
	} else if directThreat {
		reason = "sınır gerilimini azaltma"
	}

	if diplomacy.RelationScore(gs, self.ID, otherID) < 15 {
		if self.Gold < diplomacy.RelationImprovementGoldCostFor(gs)+aiMinGoldReserve {
			return "", "", false
		}
		return diplomacy.ActionImproveRelations, reason, true
	}
	if strategicTarget {
		if self.Gold < diplomacy.RelationImprovementGoldCostFor(gs)+aiMinGoldReserve {
			return "", "", false
		}
		return diplomacy.ActionImproveRelations, "stratejik hedef karşısında sınır baskısını azaltma", true
	}
	if self.Gold >= diplomacy.GiftGoldCostFor(gs)+aiMinGoldReserve {
		return diplomacy.ActionSendGift, reason, true
	}
	if self.Gold >= diplomacy.RelationImprovementGoldCostFor(gs)+aiMinGoldReserve {
		return diplomacy.ActionImproveRelations, reason, true
	}
	return "", "", false
}

// aiHandleSiegeSurrenderOffersWithSteps, oyuncu ile doğrudan ilişkili aktif
// kuşatmalarda iki yönlü teslimiyet teklifini üretir. AI-AI kuşatmaları modal
// gerektirmediği için burada teklif kuyruğuna alınmaz; otomatik savaş akışı
// onları kendi kararlarıyla çözer.
func aiHandleSiegeSurrenderOffersWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	if gs == nil || fid == "" || gs.PlayerFactionID == "" || fid == gs.PlayerFactionID {
		return
	}
	for _, target := range aiSortedRegions(gs) {
		if target == nil || target.IsSea {
			continue
		}
		siege := gs.SiegeAt(target.ID)
		if siege == nil {
			continue
		}
		attacker := gs.Armies[siege.AttackerArmyID]
		if attacker == nil || attacker.IsNaval || attacker.OwnerID == target.OwnerID || !diplomacy.IsWar(gs, faction.FactionID(attacker.OwnerID), faction.FactionID(target.OwnerID)) {
			continue
		}

		var from, to faction.FactionID
		priority := 0
		reason := ""
		shouldOffer := false
		offerAction := diplomacy.ActionProposeSurrender
		switch {
		case string(fid) == attacker.OwnerID && target.OwnerID == string(gs.PlayerFactionID):
			// Kuşatan AI, savunma hattı çözüldüğünde oyuncudan teslim olmasını ister.
			defender := gs.SelectBattleDefender(attacker, target.ID, false)
			defenderPower := 0
			if defender != nil {
				defenderPower = aiArmyStrength(gs, defender)
			}
			shouldOffer = siege.TurnsElapsed >= 2 && (siege.BreachLevel >= 1 || defenderPower == 0 || aiArmyStrength(gs, attacker) >= defenderPower*125/100)
			if shouldOffer && aiDiplomacyOfferRoll(gs, fid, gs.PlayerFactionID, diplomacy.ActionProposeSurrender) >= 70 {
				shouldOffer = false
			}
			from, to = fid, gs.PlayerFactionID
			priority = 155
			reason = "Kuşatma hattı çöktü; teslimiyet talebi"
		case string(fid) == target.OwnerID && attacker.OwnerID == string(gs.PlayerFactionID):
			// Kuşatılan AI, ağır baskı altında oyuncuya teslim olmayı teklif eder.
			defender := gs.SelectBattleDefender(attacker, target.ID, false)
			defenderPower := 0
			if defender != nil {
				defenderPower = aiArmyStrength(gs, defender)
			}
			attackerPower := aiArmyStrength(gs, attacker)
			shouldOffer = defender != nil && !gs.HasCapableSiegeReliefArmy(attacker, target) && siege.TurnsElapsed >= 3 && (siege.BreachLevel >= 1 || defenderPower*100 < attackerPower*80)
			if shouldOffer && aiDiplomacyOfferRoll(gs, fid, gs.PlayerFactionID, diplomacy.ActionProposeSurrender) >= 60 {
				shouldOffer = false
			}
			from, to = fid, gs.PlayerFactionID
			priority = 175
			reason = "Kuşatma baskısı ve savunma çöküşü"
		}
		if len(gs.LandRegionsOwnedBy(faction.FactionID(target.OwnerID))) == 1 {
			offerAction = diplomacy.ActionProposeSiegeVassalization
			reason = "Son toprak için kuşatma vassallığı"
		}
		if !shouldOffer || from == "" || to == "" || !aiDiplomacyOfferRetryAllowedForRegion(gs, from, to, offerAction, target.ID) {
			continue
		}
		queued := false
		if offerAction == diplomacy.ActionProposeSiegeVassalization {
			queued = diplomacy.QueueSiegeVassalizationOffer(gs, from, to, target.ID, priority, reason)
		} else {
			queued = diplomacy.QueueSurrenderOffer(gs, from, to, target.ID, priority, reason)
		}
		if queued {
			offerLabel := "teslimiyet"
			if offerAction == diplomacy.ActionProposeSiegeVassalization {
				offerLabel = "vassallık"
			}
			addTurnStep(steps, TurnStep{
				FactionID:     fid,
				Kind:          TurnStepDiplomacy,
				TargetFaction: to,
				TargetRegion:  target.ID,
				FocusRegion:   target.ID,
				Message:       turnFactionName(gs, fid) + " " + turnRegionName(gs, target.ID) + " için " + offerLabel + " teklifi gönderdi.",
			})
		}
	}
}
