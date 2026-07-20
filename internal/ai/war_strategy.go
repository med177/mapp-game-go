package ai

import (
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

// aiEvaluateWarOpportunitiesWithSteps selects at most one opportunistic war
// target after diplomacy has resolved peace/alliance/trade actions.
func aiEvaluateWarOpportunitiesWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	if gs == nil || !aiProactiveWarEnabled(gs) {
		return
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return
	}
	if diplomacy.DirectOverlord(gs, fid) != "" {
		return
	}
	if aiActiveWarCount(gs, fid) >= aiMaxConcurrentWars(gs, fid) || !aiWarCadenceAllows(gs, fid) {
		return
	}

	strategicContext := prepareStrategicContext(gs, fid)
	bestScore := aiWarThresholdForDifficulty(gs)
	bestTarget := faction.FactionID("")
	for _, otherID := range aiSortedFactionIDs(gs) {
		other := gs.Factions[otherID]
		if otherID == fid || other == nil || other.IsEliminated {
			continue
		}
		if overlord := diplomacy.DirectOverlord(gs, otherID); overlord != "" && overlord != fid {
			continue
		}
		rel := diplomacy.Relation(gs, fid, otherID)
		if rel == nil || rel.Stance != faction.StancePeace {
			continue
		}
		score := aiWarOpportunityScoreWithContext(gs, fid, otherID, rel, strategicContext)
		if score > bestScore {
			bestScore = score
			bestTarget = otherID
		}
	}

	if bestTarget == "" {
		return
	}
	result := diplomacy.Execute(gs, fid, bestTarget, diplomacy.ActionDeclareWar)
	if result.Applied || result.Accepted {
		addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: bestTarget, Message: turnFactionName(gs, fid) + ": " + result.Message})
	}
}

func aiWarOpportunityScore(gs *state.GameState, actor, target faction.FactionID, rel *faction.Relation) int {
	return aiWarOpportunityScoreWithContext(gs, actor, target, rel, prepareStrategicContext(gs, actor))
}

func aiWarOpportunityScoreWithContext(gs *state.GameState, actor, target faction.FactionID, rel *faction.Relation, strategicContext *StrategicContext) int {
	self := gs.Factions[actor]
	other := gs.Factions[target]
	if self == nil || other == nil || rel == nil {
		return -1
	}
	isExpansionTarget := aiHasExpansionTarget(self, target)
	isPlanTarget := aiPlanTargetsFaction(gs, actor, target)
	maxPeaceScore := -20
	if isPlanTarget {
		maxPeaceScore = 20
	} else if isExpansionTarget {
		maxPeaceScore = 10
	} else if self.AIAggressiveness >= 70 {
		maxPeaceScore = -10
	}
	if rel.Score > maxPeaceScore || !aiSharesLandBorder(gs, actor, target) {
		return -1
	}

	selfPower := diplomacy.MilitaryPower(gs, actor)
	targetPower := diplomacy.MilitaryPower(gs, target)
	if selfPower <= 0 || (targetPower > 0 && selfPower*100 < targetPower*aiMinAttackPowerPercent(gs)) || !aiStrategicWarReady(strategicContext, target) {
		return -1
	}
	frontierPower := aiFrontierPower(gs, actor, target)
	if frontierPower <= 0 {
		return -1
	}
	targetFrontierPower := aiFrontierPower(gs, target, actor)

	score := 20
	if targetPower == 0 {
		score += 30
	} else {
		score += minInt(30, maxInt(0, (selfPower-targetPower)/12))
	}
	if targetFrontierPower == 0 {
		score += 16
	} else if frontierPower > targetFrontierPower {
		score += minInt(22, (frontierPower-targetFrontierPower)/10+8)
	} else {
		score -= 18
	}
	score += minInt(18, maxInt(0, -rel.Score/2))
	if rel.Score > 0 {
		score -= rel.Score
	}
	selfRegions := len(gs.LandRegionsOwnedBy(actor))
	targetRegions := len(gs.LandRegionsOwnedBy(target))
	if targetRegions <= 2 {
		score += 12
	}
	if selfRegions >= targetRegions {
		score += 8
	}
	if gs.DeployedLandUnits(actor) >= gs.ManpowerCap(actor) {
		score += 8
	}
	score += minInt(15, aiBestBorderTargetValue(gs, actor, target)/15)
	if self.Religion != other.Religion {
		score += 6
	} else {
		score -= 6
	}
	score += (self.AIAggressiveness - 45) / 2
	if isExpansionTarget {
		score += 18
		if rel.Score <= 0 {
			score += 6
		}
		if self.AIAggressiveness >= 60 {
			score += 4
		}
	}
	if isPlanTarget {
		commitment := 50
		if plan := gs.AIPlans[actor]; plan != nil {
			commitment = plan.Commitment
		}
		score += minInt(36, 12+commitment/3)
	}
	if target == gs.PlayerFactionID {
		score -= 18
		score += aiPlayerTargetScoreBonus(gs)
	}
	return score
}
