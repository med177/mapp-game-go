package ai

import (
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

// aiHandleDiplomacyWithSteps resolves AI peace, alliance and trade decisions.
// The public wrapper remains in ai.go; this file owns the diplomacy decision loop.
func aiHandleDiplomacyWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	gs.SyncWarLedgers()
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return
	}
	if diplomacy.DirectOverlord(gs, fid) != "" {
		return
	}
	aiHandleSiegeSurrenderOffersWithSteps(gs, fid, steps)

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
			if gs.ScenarioID == "1300_ottoman_rise" {
				shouldProposePeace = diplomacy.AssessPeaceDesire(gs, fid, otherID).ShouldPropose()
			} else {
				selfPower := diplomacy.MilitaryPower(gs, fid)
				otherPower := diplomacy.MilitaryPower(gs, otherID)
				shouldProposePeace = rel.Score <= -90 || selfPower < otherPower || len(gs.RegionsOwnedBy(fid)) < len(gs.RegionsOwnedBy(otherID))
			}
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
					result := diplomacy.Execute(gs, fid, otherID, diplomacy.ActionProposePeace)
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
			var allianceAssessment diplomacy.AllianceProposalAssessment
			if rel.Score >= 25 {
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
		case faction.StanceAllied:
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

	aiEvaluateWarOpportunitiesWithSteps(gs, fid, steps)
}

// aiHandleSiegeSurrenderOffersWithSteps, oyuncu ile doğrudan ilişkili aktif
// kuşatmalarda iki yönlü teslimiyet teklifini üretir. AI-AI kuşatmaları modal
// gerektirmediği için burada teklif kuyruğuna alınmaz; otomatik savaş akışı
// onları kendi kararlarıyla çözer.
func aiHandleSiegeSurrenderOffersWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	if gs == nil || fid == "" || gs.PlayerFactionID == "" || fid == gs.PlayerFactionID || gs.DiplomacyOfferQuotaRemaining(fid) <= 0 {
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
			lastRegion := len(gs.LandRegionsOwnedBy(fid)) == 1
			shouldOffer = lastRegion || siege.TurnsElapsed >= 3 && (siege.BreachLevel >= 1 || defenderPower*100 < attackerPower*80)
			if shouldOffer && aiDiplomacyOfferRoll(gs, fid, gs.PlayerFactionID, diplomacy.ActionProposeSurrender) >= 60 {
				shouldOffer = false
			}
			from, to = fid, gs.PlayerFactionID
			priority = 175
			reason = "Kuşatma baskısı ve savunma çöküşü"
		}
		if !shouldOffer || from == "" || to == "" || !aiDiplomacyOfferRetryAllowed(gs, from, to, diplomacy.ActionProposeSurrender) {
			continue
		}
		if diplomacy.QueueSurrenderOffer(gs, from, to, target.ID, priority, reason) {
			addTurnStep(steps, TurnStep{
				FactionID:     fid,
				Kind:          TurnStepDiplomacy,
				TargetFaction: to,
				TargetRegion:  target.ID,
				Message:       turnFactionName(gs, fid) + " " + turnRegionName(gs, target.ID) + " için teslimiyet teklifi gönderdi.",
			})
			if gs.DiplomacyOfferQuotaRemaining(fid) <= 0 {
				return
			}
		}
	}
}
