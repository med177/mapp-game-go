package ai

import (
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

const aiRelationshipRepairChancePercent = 60

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
	// Tarihsel genişleme hedefi karşısında tek başına yeterli olmayan devlet,
	// genel diplomasi taramasından önce aynı hedefe baskı yapabilecek müttefik
	// arar. Kabul edilen AI-AI ittifakı bu turdaki savaş koalisyonuna da girer.
	aiPursueHistoricalWarAlliance(gs, fid, steps)

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
			if allowRelationSpending && aiHandleRelationshipRepairWithSteps(gs, fid, otherID, rel, steps) {
				continue
			}
			var allianceAssessment diplomacy.AllianceProposalAssessment
			if rel.Score >= diplomacy.AllianceRelationThreshold(gs) {
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
			if rel.Score >= 15 && diplomacy.Relation(gs, fid, otherID).Stance == faction.StancePeace && aiTradePartnerCount(gs, fid) < 3 && aiTradePartnerCount(gs, otherID) < 3 && !diplomacy.HasDirectThreat(gs, fid, otherID) && aiDiplomacyOfferRetryAllowed(gs, fid, otherID, diplomacy.ActionProposeTrade) {
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
				if gs.ScenarioID == "1300_ottoman_rise" {
					activeObjectiveConflict = diplomacy.AssessStrategicAlliance(gs, fid, otherID).ActiveObjectiveConflict
				}
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
		if rel == nil || rel.Stance != faction.StancePeace || rel.Score < diplomacy.AllianceRelationThreshold(gs) || !aiDiplomacyOfferRetryAllowed(gs, fid, candidateID, diplomacy.ActionProposeAlliance) {
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
		power := aiWarWeightedFactionPower(gs, candidateID, battlefield)
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
	action, reason, ok := aiRelationshipRepairAction(gs, fid, otherID, rel)
	if !ok || gs.DiplomacyOfferQuotaRemaining(fid) <= 0 {
		return false
	}
	self := gs.Factions[fid]
	if self == nil {
		return false
	}
	cost := aiRelationshipActionCost(action)
	if !aiCanAffordForBudget(self, cost, budget, aiBudgetEconomy) {
		if action != diplomacy.ActionSendGift || !aiCanAffordForBudget(self, economy.ResourceCost{Gold: diplomacy.RelationImprovementGoldCost}, budget, aiBudgetEconomy) {
			return false
		}
		action = diplomacy.ActionImproveRelations
		cost = aiRelationshipActionCost(action)
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
		if budget != nil {
			budget.consume(aiBudgetEconomy, cost.Gold)
		}
		return true
	}

	result := diplomacy.Execute(gs, fid, otherID, action)
	if !result.Applied {
		return false
	}
	if budget != nil {
		budget.consume(aiBudgetEconomy, cost.Gold)
	}
	return true
}

// aiHandleRelationshipRepairsAfterBudget keeps gifts and envoys as the last
// treasury-funded AI action. The 1300 budget exposes only the leftover
// FlexibleGold after every priority category has been released.
func aiHandleRelationshipRepairsAfterBudget(gs *state.GameState, fid faction.FactionID, budget *aiBudget) {
	if gs == nil || fid == "" || gs.DiplomacyOfferQuotaRemaining(fid) <= 0 {
		return
	}
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
		aiHandleRelationshipRepairWithBudget(gs, fid, otherID, rel, budget)
	}
}

func aiRelationshipActionCost(action diplomacy.Action) economy.ResourceCost {
	if action == diplomacy.ActionSendGift {
		return economy.ResourceCost{Gold: diplomacy.GiftGoldCost}
	}
	return economy.ResourceCost{Gold: diplomacy.RelationImprovementGoldCost}
}

// aiRelationshipRepairAction, ilişki aksiyonunu yalnızca somut ticari veya
// güvenlik çıkarı varsa seçer. Önce ucuz heyetle ticaret eşiğine ulaşır; daha
// yüksek ilişki hedefi gerekiyorsa ve altın rezervi uygunsa hediye kullanır.
func aiRelationshipRepairAction(gs *state.GameState, fid, otherID faction.FactionID, rel *faction.Relation) (diplomacy.Action, string, bool) {
	if gs == nil || rel == nil || fid == "" || otherID == "" || fid == otherID || rel.Stance == faction.StanceWar || diplomacy.SameRealm(gs, fid, otherID) {
		return "", "", false
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return "", "", false
	}

	strategicTarget := aiIsStrategicDiplomacyTarget(gs, fid, otherID)
	hasActiveTrade := diplomacy.HasTradeRouteBetween(gs, fid, otherID)
	hasTradeInterest := hasActiveTrade || diplomacy.CanEstablishTradeRoute(gs, fid, otherID)
	hasAllianceInterest := aiAllianceHasMeaningfulBenefit(gs, fid, otherID)
	commonEnemy := diplomacy.HasCommonEnemy(gs, fid, otherID)
	sharedThreat := diplomacy.HasSharedMajorThreat(gs, fid, otherID)
	directThreat := diplomacy.HasDirectThreat(gs, fid, otherID)
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
	if rel.Score >= desiredScore {
		return "", "", false
	}

	reason := "ticaret hattını güvenceye alma"
	if hasAllianceInterest || commonEnemy || sharedThreat {
		reason = "stratejik ortaklık hazırlığı"
	} else if directThreat {
		reason = "sınır gerilimini azaltma"
	}

	if rel.Score < 15 {
		if self.Gold < diplomacy.RelationImprovementGoldCost+aiMinGoldReserve {
			return "", "", false
		}
		return diplomacy.ActionImproveRelations, reason, true
	}
	if strategicTarget {
		if self.Gold < diplomacy.RelationImprovementGoldCost+aiMinGoldReserve {
			return "", "", false
		}
		return diplomacy.ActionImproveRelations, "stratejik hedef karşısında sınır baskısını azaltma", true
	}
	if self.Gold >= diplomacy.GiftGoldCost+aiMinGoldReserve {
		return diplomacy.ActionSendGift, reason, true
	}
	if self.Gold >= diplomacy.RelationImprovementGoldCost+aiMinGoldReserve {
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
				defenderPower = defender.TotalStrength(gs.UnitTypes)
			}
			shouldOffer = siege.TurnsElapsed >= 2 && (siege.BreachLevel >= 1 || defenderPower == 0 || attacker.TotalStrength(gs.UnitTypes) >= defenderPower*125/100)
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
				defenderPower = defender.TotalStrength(gs.UnitTypes)
			}
			attackerPower := attacker.TotalStrength(gs.UnitTypes)
			shouldOffer = siege.TurnsElapsed >= 3 && (siege.BreachLevel >= 1 || defenderPower*100 < attackerPower*80)
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
