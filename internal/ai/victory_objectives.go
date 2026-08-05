package ai

import (
	"fmt"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const aiVictoryObjectivePrefix = "victory:"

// chooseVictoryStrategicPlan senaryonun victory_conditions verisini AI'nin
// kalıcı planına çevirir. Fraksiyona özel tarihsel hedefler önce gelir; böyle
// bir hedef yoksa herkes için açık genel hedefler kullanılır. Güç, lojistik ve
// diplomasi güvenlik kapıları plan seçiminden sonra normal akışta yine uygulanır.
func chooseVictoryStrategicPlan(gs *state.GameState, self *faction.Faction, ctx *StrategicContext) *state.AIPlanState {
	if plan := chooseHistoricalVictoryStrategicPlan(gs, self, ctx); plan != nil {
		return plan
	}
	return chooseGeneralVictoryStrategicPlan(gs, self, ctx)
}

func chooseHistoricalVictoryStrategicPlan(gs *state.GameState, self *faction.Faction, ctx *StrategicContext) *state.AIPlanState {
	return chooseVictoryPlanByAudience(gs, self, ctx, true)
}

func chooseGeneralVictoryStrategicPlan(gs *state.GameState, self *faction.Faction, ctx *StrategicContext) *state.AIPlanState {
	return chooseVictoryPlanByAudience(gs, self, ctx, false)
}

func chooseVictoryPlanByAudience(gs *state.GameState, self *faction.Faction, ctx *StrategicContext, historical bool) *state.AIPlanState {
	if gs == nil || self == nil || ctx == nil || len(gs.ScenarioVictories) == 0 {
		return nil
	}

	options := make([]scenario.VictoryOptionDef, 0, len(gs.ScenarioVictories))
	for _, option := range gs.ScenarioVictories {
		if !option.VisibleForFaction(string(self.ID)) {
			continue
		}
		isHistorical := len(option.AllowedFactions) > 0
		if isHistorical == historical {
			options = append(options, option)
		}
	}
	return chooseVictoryPlanFromOptions(gs, self, ctx, options, historical)
}

func chooseVictoryPlanFromOptions(gs *state.GameState, self *faction.Faction, ctx *StrategicContext, options []scenario.VictoryOptionDef, historical bool) *state.AIPlanState {
	var best *state.AIPlanState
	bestScore := -1 << 30
	for optionIndex, option := range options {
		if plan, score := aiVictoryRegionalPlan(gs, self, ctx, option, historical); plan != nil && score > bestScore {
			best, bestScore = plan, score
		}
		if option.Type == "military" {
			if target := aiBestGeneralVictoryTarget(gs, self, ctx); target != "" {
				score := 160 - optionIndex
				if historical {
					score += 1000
				}
				if score > bestScore {
					plan := newStrategicPlan(gs, self, ctx, state.AIObjectiveExpand, target, victoryPlanReason(option, historical))
					plan.ObjectiveID = aiVictoryObjectivePrefix + option.ID
					plan.Commitment = aiVictoryPlanCommitment(self, historical)
					best, bestScore = plan, score
				}
			}
		}
		// Bölgesel hedefi olmayan ekonomik/beka koşulları, ilgili fraksiyonu
		// rastgele uzak bir savaşa sürüklemek yerine ekonomisini ve savunmasını
		// büyütmeye yönlendirir. Genel hedeflerde bunlar bölgesel/militer
		// fırsatlardan sonra gelir.
		if option.Type == "economic" || option.Type == "survive_turns" {
			score := 80 - optionIndex
			if historical {
				score += 1000
			}
			if score > bestScore {
				best = newVictoryConsolidationPlan(gs, self, ctx, option, historical)
				bestScore = score
			}
		}
	}
	if best != nil {
		return best
	}
	return nil
}

func aiVictoryRegionalPlan(gs *state.GameState, self *faction.Faction, ctx *StrategicContext, option scenario.VictoryOptionDef, historical bool) (*state.AIPlanState, int) {
	if len(option.RegionTargets()) == 0 {
		return nil, 0
	}
	type targetCandidate struct {
		id           faction.FactionID
		regions      []world.RegionID
		directBorder bool
	}
	targets := make(map[faction.FactionID]*targetCandidate)
	for _, rawRegionID := range option.RegionTargets() {
		region := gs.Regions[world.RegionID(rawRegionID)]
		if region == nil || region.IsSea || region.OwnerID == "" || region.OwnerID == string(self.ID) {
			continue
		}
		targetID := faction.FactionID(region.OwnerID)
		target := gs.Factions[targetID]
		if target == nil || target.IsEliminated || diplomacy.SameRealm(gs, self.ID, targetID) {
			continue
		}
		candidate := targets[targetID]
		if candidate == nil {
			candidate = &targetCandidate{id: targetID}
			targets[targetID] = candidate
		}
		candidate.regions = append(candidate.regions, region.ID)
		for _, neighborID := range region.Neighbors {
			if neighbor := gs.Regions[neighborID]; neighbor != nil && !neighbor.IsSea && neighbor.OwnerID == string(self.ID) {
				candidate.directBorder = true
				break
			}
		}
	}

	var best *state.AIPlanState
	bestScore := -1 << 30
	for _, candidate := range targets {
		if !candidate.directBorder || !aiSharesLandBorder(gs, self.ID, candidate.id) {
			continue
		}
		rel := diplomacy.Relation(gs, self.ID, candidate.id)
		if rel != nil && rel.Stance == faction.StanceAllied {
			continue
		}
		score := 200 + len(candidate.regions)*80
		if historical {
			score += 1000
		}
		if rel != nil && rel.Stance == faction.StanceWar {
			score += 80
		}
		powerDelta := ctx.militaryPower(self.ID) - ctx.militaryPower(candidate.id)
		score += maxInt(-50, minInt(50, powerDelta/10))
		score += minInt(40, ctx.powerAtFrontier(candidate.id)/15)
		if score <= bestScore {
			continue
		}
		plan := newStrategicPlan(gs, self, ctx, state.AIObjectiveExpand, candidate.id, victoryPlanReason(option, historical))
		plan.ObjectiveID = aiVictoryObjectivePrefix + option.ID
		plan.TargetRegionIDs = aiVictoryTargetRegions(gs, self.ID, candidate.id, candidate.regions, ctx)
		plan.Commitment = aiVictoryPlanCommitment(self, historical)
		best, bestScore = plan, score
	}
	return best, bestScore
}

func newVictoryConsolidationPlan(gs *state.GameState, self *faction.Faction, ctx *StrategicContext, option scenario.VictoryOptionDef, historical bool) *state.AIPlanState {
	targets := append([]world.RegionID(nil), ctx.OwnedLandRegionIDs...)
	if len(targets) > aiPlanTargetRegionLimit(gs) {
		targets = targets[:aiPlanTargetRegionLimit(gs)]
	}
	return &state.AIPlanState{
		ObjectiveID:     aiVictoryObjectivePrefix + option.ID,
		Kind:            state.AIObjectiveConsolidate,
		TargetRegionIDs: targets,
		StartedTurn:     gs.Turn,
		ReassessTurn:    gs.Turn + aiPlanHorizonTurns(gs),
		Commitment:      aiVictoryPlanCommitment(self, historical),
		Reason:          victoryPlanReason(option, historical),
	}
}

func aiVictoryTargetRegions(gs *state.GameState, owner, target faction.FactionID, preferred []world.RegionID, ctx *StrategicContext) []world.RegionID {
	limit := aiPlanTargetRegionLimit(gs)
	regions := make([]world.RegionID, 0, limit)
	for _, regionID := range preferred {
		if region := gs.Regions[regionID]; region != nil && !region.IsSea && region.OwnerID == string(target) {
			regions = append(regions, regionID)
			if len(regions) == limit {
				return regions
			}
		}
	}
	for _, regionID := range aiPlanTargetRegions(gs, owner, target, ctx) {
		if containsRegionID(regions, regionID) {
			continue
		}
		regions = append(regions, regionID)
		if len(regions) == limit {
			break
		}
	}
	return regions
}

func aiBestGeneralVictoryTarget(gs *state.GameState, self *faction.Faction, ctx *StrategicContext) faction.FactionID {
	best := faction.FactionID("")
	bestScore := -1 << 30
	for _, targetID := range aiSortedFactionIDs(gs) {
		if targetID == self.ID || !aiSharesLandBorder(gs, self.ID, targetID) {
			continue
		}
		target := gs.Factions[targetID]
		if target == nil || target.IsEliminated || diplomacy.SameRealm(gs, self.ID, targetID) {
			continue
		}
		rel := diplomacy.Relation(gs, self.ID, targetID)
		if rel != nil && rel.Stance == faction.StanceAllied {
			continue
		}
		score := ctx.powerAtFrontier(targetID) - ctx.militaryPower(targetID)/12
		if rel != nil && rel.Stance == faction.StanceWar {
			score += 50
		}
		if score > bestScore || (score == bestScore && (best == "" || targetID < best)) {
			best, bestScore = targetID, score
		}
	}
	return best
}

func aiVictoryPlanCommitment(self *faction.Faction, historical bool) int {
	commitment := aiPlanCommitment(self, state.AIObjectiveExpand)
	if historical {
		return maxInt(70, commitment)
	}
	return maxInt(55, commitment)
}

func victoryPlanReason(option scenario.VictoryOptionDef, historical bool) string {
	if historical {
		return fmt.Sprintf("tarihsel zafer hedefi: %s", option.ID)
	}
	return fmt.Sprintf("genel zafer hedefi: %s", option.ID)
}

func aiPlanIsVictoryObjective(plan *state.AIPlanState) bool {
	return plan != nil && len(plan.ObjectiveID) > len(aiVictoryObjectivePrefix) && plan.ObjectiveID[:len(aiVictoryObjectivePrefix)] == aiVictoryObjectivePrefix
}
