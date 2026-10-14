package ai

import (
	"math"
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
)

const (
	aiCouncilExactMatchingMaxRoles       = 8
	aiCouncilWarValueMultiplier          = 3
	aiCouncilPeaceValueMultiplier        = 1
	aiCharacterAdoptionAmortizationTurns = 12
)

type aiCouncilAssignment struct {
	score  int
	byRole []*army.Commander
}

type aiCouncilValueContext struct {
	taxIncome             float64
	production            state.RegionProductionSummary
	currentResourceMod    float64
	grainUtilityPercent   float64
	satisfactionWeight    float64
	armyStrength          float64
	warValueMultiplier    float64
	relationTargetCount   float64
	researchValuePerSpeed float64
}

// aiManageNeutralCharacters rebalances a faction's council to maximize its
// estimated current return, and adopts one neutral character only if doing so
// improves that return.
func aiManageNeutralCharacters(gs *state.GameState, fid faction.FactionID) {
	aiManageNeutralCharactersWithContext(gs, fid, nil)
}

func aiManageNeutralCharactersWithContext(gs *state.GameState, fid faction.FactionID, planningContext *StrategicContext) {
	if gs == nil || fid == "" || (fid == gs.PlayerFactionID && !gs.AIControlsPlayerFaction) {
		return
	}
	owner := gs.Factions[fid]
	if owner == nil || owner.IsEliminated {
		return
	}

	roles := gs.CouncilConfig.WithDefaults().Roles
	if len(roles) == 0 {
		return
	}
	characters := gs.CharacterRoster(string(fid))
	sort.Slice(characters, func(i, j int) bool { return characters[i].ID < characters[j].ID })
	valueContext := aiNewCouncilValueContext(gs, fid, planningContext)
	best := aiBestCouncilAssignment(gs, fid, roles, characters, valueContext)

	var bestCandidate *army.Commander
	bestWithCandidate := best
	for _, candidate := range gs.NeutralCharacterCandidates() {
		if candidate.AdoptionCostGold < 0 || candidate.GoldUpkeep < 0 ||
			owner.Gold < candidate.AdoptionCostGold {
			continue
		}
		trial := append(append([]*army.Commander(nil), characters...), candidate)
		sort.Slice(trial, func(i, j int) bool { return trial[i].ID < trial[j].ID })
		assignment := aiBestCouncilAssignment(gs, fid, roles, trial, valueContext)
		adoptionCostPerTurn := (candidate.AdoptionCostGold + aiCharacterAdoptionAmortizationTurns - 1) /
			aiCharacterAdoptionAmortizationTurns
		upkeep := candidate.GoldUpkeep
		for roleIndex, assigned := range assignment.byRole {
			if assigned == candidate && roles[roleIndex].NoCost {
				upkeep = 0
				break
			}
		}
		if assignment.score-best.score <= upkeep+adoptionCostPerTurn ||
			assignment.score <= bestWithCandidate.score || !aiAssignmentContains(assignment, candidate) {
			continue
		}
		bestCandidate = candidate
		bestWithCandidate = assignment
	}

	if bestCandidate != nil && gs.AdoptNeutralCharacterForFaction(fid, bestCandidate.ID) {
		characters = append(characters, bestCandidate)
		best = bestWithCandidate
	}
	for _, character := range characters {
		if character != nil {
			character.AssignedCouncilRole = ""
		}
	}
	for roleIndex, character := range best.byRole {
		if character != nil {
			gs.AssignCharacterToCouncilRole(fid, character.ID, roles[roleIndex].ID)
		}
	}
}

func aiAssignmentContains(assignment aiCouncilAssignment, character *army.Commander) bool {
	if character == nil {
		return false
	}
	for _, assigned := range assignment.byRole {
		if assigned == character {
			return true
		}
	}
	return false
}

func aiNewCouncilValueContext(gs *state.GameState, fid faction.FactionID, planningContext *StrategicContext) aiCouncilValueContext {
	context := aiCouncilValueContext{
		taxIncome:          float64(gs.TaxIncomeForFaction(fid)),
		production:         gs.FactionProductionSummary(fid),
		currentResourceMod: gs.CouncilResourceProductionMod(fid),
		grainUtilityPercent: float64(aiGrainUtilityPercent(
			aiBuildEconomySnapshot(gs, fid),
		)),
		warValueMultiplier: aiCouncilPeaceValueMultiplier,
	}
	for _, region := range gs.LandRegionsOwnedBy(fid) {
		switch {
		case region.Satisfaction < 40:
			context.satisfactionWeight += 4
		case region.Satisfaction < 60:
			context.satisfactionWeight += 3
		case region.Satisfaction < 80:
			context.satisfactionWeight += 2
		case region.Satisfaction < 100:
			context.satisfactionWeight++
		}
	}
	for _, relation := range gs.Relations {
		if relation != nil && relation.Stance == faction.StanceWar &&
			(relation.FactionA == fid || relation.FactionB == fid) {
			context.warValueMultiplier = aiCouncilWarValueMultiplier
			break
		}
	}
	for _, currentArmy := range aiSortedArmies(gs) {
		if currentArmy.OwnerID != string(fid) {
			continue
		}
		strength := currentArmy.TotalStrength(gs.UnitTypes)
		if strength == 0 && gs.UnitTypes == nil {
			strength = len(currentArmy.Units) * 10
		}
		context.armyStrength += float64(strength)
	}
	context.relationTargetCount = aiCouncilRelationTargetCount(gs, fid)
	context.researchValuePerSpeed = aiCouncilResearchValuePerSpeed(gs, fid, planningContext)
	if context.currentResourceMod <= 0 {
		context.currentResourceMod = 1
	}
	return context
}

func aiCouncilRelationTargetCount(gs *state.GameState, fid faction.FactionID) float64 {
	count := 0
	for _, otherID := range aiSortedFactionIDs(gs) {
		if otherID == fid {
			continue
		}
		other := gs.Factions[otherID]
		if other == nil || other.IsEliminated {
			continue
		}
		relation := diplomacy.Relation(gs, fid, otherID)
		if relation == nil {
			relation = &faction.Relation{FactionA: fid, FactionB: otherID, Stance: faction.StancePeace}
		}
		action, _, ok := aiRelationshipRepairActionWithThreat(gs, fid, otherID, relation, false, false)
		if ok && action == diplomacy.ActionImproveRelations {
			count++
		}
	}
	if count > 2 {
		count = 2
	}
	return float64(count)
}

func aiCouncilResearchValuePerSpeed(gs *state.GameState, fid faction.FactionID, planningContext *StrategicContext) float64 {
	owner := gs.Factions[fid]
	if owner == nil || gs.TechTypes == nil {
		return 0
	}
	var research *tech.Technology
	if owner.Research.ActiveID != "" {
		research = gs.TechTypes[owner.Research.ActiveID]
	}
	if research == nil {
		if planningContext != nil {
			research = aiSelectResearchTechnology(gs, owner, nil, planningContext)
		} else {
			research = aiSelectLegacyResearchTechnology(gs, owner, nil, aiOwnedRegionSet(gs, fid))
		}
	}
	if research == nil {
		return 0
	}
	researchEffects := tech.ComputeEffects(owner.Research.Completed, gs.TechTypes)
	currentSpeed := researchEffects.ResearchSpeedMod + gs.CouncilResearchSpeedMod(fid)
	turns := tech.EffectiveTurnsRequired(research, currentSpeed)
	if turns <= 0 {
		return 0
	}
	goldCost := tech.EffectiveResearchGoldCost(research, researchEffects.ResearchCostMod)
	return float64(goldCost) / float64(turns)
}

func aiBestCouncilAssignment(
	gs *state.GameState,
	fid faction.FactionID,
	roles []scenario.CouncilRoleDefinition,
	characters []*army.Commander,
	context aiCouncilValueContext,
) aiCouncilAssignment {
	scores := make([][]int, len(roles))
	for roleIndex, role := range roles {
		scores[roleIndex] = make([]int, len(characters))
		for characterIndex, character := range characters {
			if !aiCouncilCharacterEligible(gs, fid, role.ID, character) {
				continue
			}
			effects := gs.CouncilRoleEffectsForCharacter(character, role.ID)
			scores[roleIndex][characterIndex] = aiCouncilRoleValue(gs, fid, role.ID, character, effects, context)
		}
	}
	if len(roles) > aiCouncilExactMatchingMaxRoles {
		return aiGreedyCouncilAssignment(roles, characters, scores)
	}

	type previousChoice struct {
		roleIndex int
		valid     bool
	}
	stateCount := 1 << len(roles)
	unreachable := -int(^uint(0)>>1) / 4
	previous := make([]int, stateCount)
	for mask := range previous {
		previous[mask] = unreachable
	}
	previous[0] = 0
	choices := make([][]previousChoice, len(characters)+1)

	for characterIndex := range characters {
		next := make([]int, stateCount)
		for mask := range next {
			next[mask] = unreachable
		}
		choices[characterIndex+1] = make([]previousChoice, stateCount)
		for mask, currentScore := range previous {
			if currentScore == unreachable {
				continue
			}
			if currentScore > next[mask] {
				next[mask] = currentScore
				choices[characterIndex+1][mask] = previousChoice{roleIndex: -1, valid: true}
			}
			for roleIndex, roleScore := range scores {
				if mask&(1<<roleIndex) != 0 || roleScore[characterIndex] <= 0 {
					continue
				}
				nextMask := mask | 1<<roleIndex
				score := currentScore + roleScore[characterIndex]
				if score > next[nextMask] {
					next[nextMask] = score
					choices[characterIndex+1][nextMask] = previousChoice{roleIndex: roleIndex, valid: true}
				}
			}
		}
		previous = next
	}

	bestMask := 0
	bestScore := 0
	for mask, score := range previous {
		if score > bestScore {
			bestMask, bestScore = mask, score
		}
	}
	result := aiCouncilAssignment{score: bestScore, byRole: make([]*army.Commander, len(roles))}
	for characterIndex := len(characters); characterIndex > 0; characterIndex-- {
		choice := choices[characterIndex][bestMask]
		if !choice.valid {
			continue
		}
		if choice.roleIndex >= 0 {
			result.byRole[choice.roleIndex] = characters[characterIndex-1]
			bestMask &^= 1 << choice.roleIndex
		}
	}
	return result
}

func aiGreedyCouncilAssignment(
	roles []scenario.CouncilRoleDefinition,
	characters []*army.Commander,
	scores [][]int,
) aiCouncilAssignment {
	result := aiCouncilAssignment{byRole: make([]*army.Commander, len(roles))}
	used := make([]bool, len(characters))
	for roleIndex := range roles {
		bestIndex, bestScore := -1, 0
		for characterIndex, score := range scores[roleIndex] {
			if !used[characterIndex] && score > bestScore {
				bestIndex, bestScore = characterIndex, score
			}
		}
		if bestIndex >= 0 {
			result.byRole[roleIndex] = characters[bestIndex]
			used[bestIndex] = true
			result.score += bestScore
		}
	}
	return result
}

func aiCouncilCharacterEligible(gs *state.GameState, fid faction.FactionID, roleID string, character *army.Commander) bool {
	if gs == nil || character == nil ||
		(character.Type != army.CharacterTypeAdministrator && character.Type != army.CharacterTypeCharacter) ||
		!character.ActiveInYear(gs.Year) || (character.AssignedArmyID != "" && roleID != state.CouncilRoleMilitaryResponsible) ||
		gs.DeadCommanderIDs[character.ID] || gs.DismissedCommanderIDs[character.ID] ||
		gs.ExiledCommanderIDs[character.ID] || character.CaptiveAtSettlementID != "" ||
		gs.CommanderUnavailable(character) {
		return false
	}
	return character.OwnerID == string(fid) || character.OwnerID == state.NeutralCharacterOwnerID
}

func aiCouncilRoleValue(
	gs *state.GameState,
	fid faction.FactionID,
	roleID string,
	character *army.Commander,
	effects scenario.CouncilEffects,
	context aiCouncilValueContext,
) int {
	value := (normalizedCouncilMod(effects.IncomeMod) - 1) * context.taxIncome
	resourceDelta := normalizedCouncilMod(effects.ResourceProductionMod) - 1
	if resourceDelta != 0 {
		base := context.production
		base.Grain = int(float64(base.Grain) / context.currentResourceMod)
		base.Iron = int(float64(base.Iron) / context.currentResourceMod)
		base.Timber = int(float64(base.Timber) / context.currentResourceMod)
		base.Stone = int(float64(base.Stone) / context.currentResourceMod)
		base.Spice = int(float64(base.Spice) / context.currentResourceMod)
		base.Cloth = int(float64(base.Cloth) / context.currentResourceMod)
		value += resourceDelta * float64(maxInt(0, base.Grain)) *
			float64(aiResourcePrice(gs, economy.GoodGrain)) * context.grainUtilityPercent / 100
		value += resourceDelta * float64(maxInt(0, base.Iron)) * float64(aiResourcePrice(gs, economy.GoodIron))
		value += resourceDelta * float64(maxInt(0, base.Timber)) * float64(aiResourcePrice(gs, economy.GoodTimber))
		value += resourceDelta * float64(maxInt(0, base.Stone)) * float64(aiResourcePrice(gs, economy.GoodStone))
		value += resourceDelta * float64(maxInt(0, base.Spice)) * float64(aiResourcePrice(gs, economy.GoodSpice))
		value += resourceDelta * float64(maxInt(0, base.Cloth)) * float64(aiResourcePrice(gs, economy.GoodCloth))
	}
	value += float64(effects.SatisfactionPoints) * context.satisfactionWeight * 2
	militaryDelta := normalizedCouncilMod(effects.MilitaryPowerMod) - 1
	value += militaryDelta * context.armyStrength * context.warValueMultiplier * 0.5
	if roleID == state.CouncilRoleMilitaryResponsible && character.AssignedArmyID != "" {
		if currentArmy := gs.Armies[army.ArmyID(character.AssignedArmyID)]; currentArmy != nil &&
			currentArmy.Commander != nil && currentArmy.Commander.ID == character.ID {
			strength := currentArmy.TotalStrength(gs.UnitTypes)
			value += (normalizedCouncilMod(effects.CommanderPowerMod) - 1) *
				float64(strength) * context.warValueMultiplier * 0.5
		}
	}
	value += float64(effects.RelationBonusPoints) * context.relationTargetCount * 2
	value += (normalizedCouncilMod(effects.ResearchSpeedMod) - 1) * context.researchValuePerSpeed
	if math.IsNaN(value) {
		return 0
	}
	return int(math.Round(value))
}

func normalizedCouncilMod(mod float64) float64 {
	if mod == 0 {
		return 1
	}
	return mod
}
