package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
)

const aiResearchProjectionTurns = 12

type aiResearchSignals struct {
	Plan                  *state.AIPlanState
	Context               *StrategicContext
	Composition           aiLandComposition
	CompositionTarget     aiCompositionTarget
	BattleNeeds           aiRecruitmentBattleNeeds
	Economy               aiEconomySnapshot
	OwnedRegions          int
	CoastalRegions        int
	TradeIncome           int
	LowSatisfaction       int
	SatisfactionDeficit   int
	ReligionMismatch      int
	ActiveWars            int
	IronProduction        int
	TimberProduction      int
	StoneProduction       int
	NavalReserveShortfall int
}

type aiResearchCandidate struct {
	Technology    *tech.Technology
	Score         int
	CategoryScore int
	EffectScore   int
	UnlockScore   int
	FutureScore   int
	DurationCost  int
	GoldCost      int
}

func aiSelectResearchTechnology(gs *state.GameState, self *faction.Faction, budget *aiBudget, ctx *StrategicContext) *tech.Technology {
	if gs == nil || self == nil || self.IsEliminated || gs.TechTypes == nil {
		return nil
	}
	if gs.ScenarioID != "1300_ottoman_rise" {
		return aiSelectLegacyResearchTechnology(gs, self, budget, aiOwnedRegionSet(gs, self.ID))
	}
	if ctx == nil {
		ctx = prepareStrategicContext(gs, self.ID)
	}
	signals := aiBuildResearchSignals(gs, self, ctx)
	techIDs := aiSortedTechnologyIDs(gs)
	ownedRegions := aiOwnedRegionSet(gs, self.ID)

	var best aiResearchCandidate
	found := false
	for _, techID := range techIDs {
		technology := gs.TechTypes[techID]
		if !aiResearchCandidateAvailable(gs, self, technology, budget, ownedRegions) {
			continue
		}
		candidate := aiScoreResearchCandidate(gs, self, technology, signals)
		if !found || aiResearchCandidateBetter(candidate, best) {
			best = candidate
			found = true
		}
	}
	if !found {
		return nil
	}
	return best.Technology
}

func aiSelectLegacyResearchTechnology(gs *state.GameState, self *faction.Faction, budget *aiBudget, ownedRegions map[string]bool) *tech.Technology {
	var best *tech.Technology
	bestScore := -int(^uint(0)>>1) - 1
	for _, techID := range aiSortedTechnologyIDs(gs) {
		technology := gs.TechTypes[techID]
		if !aiResearchCandidateAvailable(gs, self, technology, budget, ownedRegions) {
			continue
		}
		score := 0
		switch technology.Category {
		case tech.CategoryMilitary:
			score = 100
			if technology.Effects.InfantryAttackMod > 0 || technology.Effects.CavalryAttackMod > 0 {
				score += 20
			}
		case tech.CategoryEconomy:
			score = 70
			if technology.Effects.GoldPerRegion > 0 {
				score += 15
			}
		case tech.CategoryNaval:
			score = 50
		case tech.CategoryDiplomacy:
			score = 40
		case tech.CategoryReligion:
			score = 30
		}
		score -= technology.TurnsRequired / 2
		if score > bestScore {
			best = technology
			bestScore = score
		}
	}
	return best
}

func aiResearchCandidateAvailable(gs *state.GameState, self *faction.Faction, technology *tech.Technology, budget *aiBudget, ownedRegions map[string]bool) bool {
	if gs == nil || self == nil || technology == nil || self.Research.Completed[technology.ID] || !tech.IsUnlockedForContext(&self.Research, technology, gs.Year, ownedRegions) {
		return false
	}
	return aiCanAffordForBudget(self, economy.ResourceCost{Gold: technology.GoldCost}, budget, aiBudgetResearch)
}

func aiOwnedRegionSet(gs *state.GameState, fid faction.FactionID) map[string]bool {
	owned := make(map[string]bool)
	if gs == nil {
		return owned
	}
	for _, region := range gs.LandRegionsOwnedBy(fid) {
		owned[string(region.ID)] = true
	}
	return owned
}

func aiScoreResearchCandidate(gs *state.GameState, self *faction.Faction, technology *tech.Technology, signals aiResearchSignals) aiResearchCandidate {
	categoryScore := aiResearchCategoryScore(technology.Category, signals)
	effectScore := aiResearchEffectScore(gs, self, technology, signals)
	unlockScore := aiResearchUnitUnlockScore(gs, self, technology, signals)
	futureScore := aiResearchFollowOnScore(gs, self, technology)
	durationCost := maxInt(1, technology.TurnsRequired)*5 + technology.GoldCost/4
	return aiResearchCandidate{
		Technology:    technology,
		Score:         categoryScore + effectScore + unlockScore + futureScore - durationCost,
		CategoryScore: categoryScore,
		EffectScore:   effectScore,
		UnlockScore:   unlockScore,
		FutureScore:   futureScore,
		DurationCost:  durationCost,
		GoldCost:      technology.GoldCost,
	}
}

func aiResearchCategoryScore(category tech.Category, signals aiResearchSignals) int {
	kind := state.AIObjectiveConsolidate
	if signals.Plan != nil {
		kind = signals.Plan.Kind
	}
	switch kind {
	case state.AIObjectiveExpand:
		switch category {
		case tech.CategoryMilitary:
			return 100
		case tech.CategoryEconomy:
			return 60
		case tech.CategoryNaval:
			return 45
		case tech.CategoryDiplomacy, tech.CategoryReligion:
			return 35
		}
	case state.AIObjectiveDefend:
		switch category {
		case tech.CategoryMilitary:
			return 90
		case tech.CategoryEconomy:
			return 70
		case tech.CategoryReligion:
			return 65
		case tech.CategoryDiplomacy:
			return 55
		case tech.CategoryNaval:
			return 35
		}
	default:
		switch category {
		case tech.CategoryEconomy:
			return 100
		case tech.CategoryReligion:
			return 75
		case tech.CategoryDiplomacy:
			return 65
		case tech.CategoryMilitary:
			return 55
		case tech.CategoryNaval:
			return 40
		}
	}
	return 0
}

func aiResearchEffectScore(gs *state.GameState, self *faction.Faction, technology *tech.Technology, signals aiResearchSignals) int {
	effects := technology.Effects
	kind := state.AIObjectiveConsolidate
	if signals.Plan != nil {
		kind = signals.Plan.Kind
	}
	attackPercent := int((effects.InfantryAttackMod + effects.CavalryAttackMod + effects.SiegeAttackMod) * 100)
	defensePercent := int(effects.LandDefenseMod * 100)
	attackWeight, defenseWeight := 3, 3
	switch kind {
	case state.AIObjectiveExpand:
		attackWeight, defenseWeight = 5, 2
	case state.AIObjectiveDefend:
		attackWeight, defenseWeight = 2, 5
	}
	score := attackPercent*attackWeight + defensePercent*defenseWeight
	if signals.BattleNeeds.FortifiedTarget && effects.SiegeAttackMod > 0 {
		score += int(effects.SiegeAttackMod*100) * 6
	}
	moveWeight := 45
	if kind == state.AIObjectiveExpand {
		moveWeight = 70
	}
	score += effects.MoveBonus * moveWeight

	projectedValue := effects.GoldPerRegion * signals.OwnedRegions * aiResearchProjectionTurns
	projectedValue += int(float64(signals.TradeIncome)*effects.MarketGoldMod) * aiResearchProjectionTurns
	grainGain := int(float64(signals.Economy.GrainProduction) * effects.GrainMod)
	projectedValue += grainGain * aiResourcePrice(gs, economy.GoodGrain) * aiResearchProjectionTurns * aiGrainUtilityPercent(signals.Economy) / 100
	projectedValue += aiResearchResourceProjectedValue(gs, self.Iron, signals.IronProduction, effects.IronMod, economy.GoodIron)
	projectedValue += aiResearchResourceProjectedValue(gs, self.Timber, signals.TimberProduction, effects.TimberMod, economy.GoodTimber)
	projectedValue += aiResearchResourceProjectedValue(gs, self.Stone, signals.StoneProduction, effects.StoneMod, economy.GoodStone)
	if projectedValue > 0 {
		score += minInt(350, projectedValue*20/maxInt(1, technology.GoldCost))
	}

	stabilityValue := effects.SatisfactionBonus * (signals.LowSatisfaction*12 + signals.SatisfactionDeficit/10)
	if kind == state.AIObjectiveDefend || kind == state.AIObjectiveConsolidate {
		stabilityValue = stabilityValue * 3 / 2
	}
	score += minInt(300, stabilityValue)
	score += int(effects.ConversionSpeedMod) * signals.ReligionMismatch * 55

	peaceWeight := 1
	switch kind {
	case state.AIObjectiveDefend:
		peaceWeight = 3
	case state.AIObjectiveConsolidate:
		peaceWeight = 4
	}
	score += effects.PeaceRelationBonus * peaceWeight * maxInt(1, signals.ActiveWars)

	if signals.CoastalRegions > 0 {
		navalPercent := int((effects.NavalAttackMod + effects.NavalDefenseMod) * 100)
		score += navalPercent * 3
		score += effects.NavalMoveBonus * 50
		if signals.Context != nil && (len(signals.Context.NavalThreats) > 0 || len(signals.Context.ThreatenedPortIDs) > 0) && technology.Category == tech.CategoryNaval {
			score += 220
		}
	} else if technology.Category == tech.CategoryNaval {
		score -= 140
	}
	return score
}

func aiResearchResourceProjectedValue(gs *state.GameState, stock, production int, modifier float64, good economy.GoodType) int {
	if modifier <= 0 || production <= 0 {
		return 0
	}
	gain := int(float64(production) * modifier)
	if gain <= 0 {
		return 0
	}
	utility := aiResearchResourceUtility(stock, production)
	return gain * aiResourcePrice(gs, good) * aiResearchProjectionTurns * utility / 100
}

func aiResearchResourceUtility(stock, production int) int {
	switch {
	case production <= 0 || stock < maxInt(40, production*2):
		return 150
	case stock < maxInt(100, production*4):
		return 100
	case stock < maxInt(200, production*8):
		return 60
	default:
		return 25
	}
}

func aiResearchUnitUnlockScore(gs *state.GameState, self *faction.Faction, technology *tech.Technology, signals aiResearchSignals) int {
	score := 0
	unitIDs := make([]string, 0, len(gs.UnitTypes))
	for unitID := range gs.UnitTypes {
		unitIDs = append(unitIDs, unitID)
	}
	sort.Strings(unitIDs)
	for _, unitID := range unitIDs {
		unitType := gs.UnitTypes[unitID]
		if unitType == nil || !unitType.RequiresTech(technology.ID) {
			continue
		}
		unlockValue := 60 + (unitType.Attack+unitType.Defense+unitType.Morale/5)/2
		if aiLandUnitCategory(unitType.Category) {
			need := aiCategoryCompositionNeed(signals.CompositionTarget, signals.Composition, unitType.Category)
			needPercent := need / maxInt(1, signals.Composition.Total+1)
			unlockValue += minInt(180, needPercent*4)
			if unitType.Category == army.CategorySiege && signals.BattleNeeds.FortifiedTarget && signals.BattleNeeds.SiegeShortfall > 0 {
				unlockValue += 220
			}
			if !aiFactionHasUnitBuildingLevel(gs, self.ID, unitType) {
				unlockValue /= 2
			}
		} else {
			if signals.CoastalRegions == 0 {
				unlockValue /= 4
			} else if unitType.Category == army.CategoryNavalTrans && signals.ActiveWars > 0 {
				unlockValue += 80
			} else if unitType.Category == army.CategoryNavalWar && signals.ActiveWars > 0 {
				unlockValue += 180
			}
			if unitType.Category == army.CategoryNavalWar && signals.NavalReserveShortfall > 0 {
				unlockValue += 160 + minInt(240, signals.NavalReserveShortfall*40)
			}
		}
		score += unlockValue
	}
	return score
}

func aiFactionHasUnitBuildingLevel(gs *state.GameState, fid faction.FactionID, unitType *army.UnitType) bool {
	if gs == nil || unitType == nil {
		return false
	}
	requiredBuilding := unitType.RequiredBldg
	if requiredBuilding == "" {
		requiredBuilding = "barracks"
	}
	requiredLevel := maxInt(1, unitType.RequiredBldgLevel)
	for _, region := range aiSortedRegions(gs) {
		if region.OwnerID == string(fid) && !region.IsSea && !region.IsLocked && aiBuildingLevel(region, requiredBuilding) >= requiredLevel {
			return true
		}
	}
	return false
}

func aiResearchFollowOnScore(gs *state.GameState, self *faction.Faction, technology *tech.Technology) int {
	if gs == nil || self == nil || technology == nil {
		return 0
	}
	score := 0
	for _, candidateID := range aiSortedTechnologyIDs(gs) {
		candidate := gs.TechTypes[candidateID]
		if candidate == nil || !stringInList(technology.ID, candidate.Requires) || !aiOtherTechRequirementsCompleted(self, candidate, technology.ID) {
			continue
		}
		score += 15
		for _, unitType := range gs.UnitTypes {
			if unitType != nil && unitType.RequiresTech(candidate.ID) {
				score += 30
				break
			}
		}
	}
	return score
}

func aiOtherTechRequirementsCompleted(self *faction.Faction, technology *tech.Technology, ignored string) bool {
	for _, requirement := range technology.Requires {
		if requirement != ignored && !self.Research.Completed[requirement] {
			return false
		}
	}
	return true
}

func aiBuildResearchSignals(gs *state.GameState, self *faction.Faction, ctx *StrategicContext) aiResearchSignals {
	signals := aiResearchSignals{
		Plan:                  gs.AIPlans[self.ID],
		Context:               ctx,
		Composition:           aiFactionLandComposition(gs, self.ID),
		Economy:               aiBuildEconomySnapshot(gs, self.ID),
		OwnedRegions:          len(gs.LandRegionsOwnedBy(self.ID)),
		TradeIncome:           aiResearchTradeIncome(gs, self.ID),
		NavalReserveShortfall: aiWarshipReserveShortfall(gs, self.ID, ctx),
		ActiveWars:            len(ctx.WarEnemies),
	}
	signals.CompositionTarget = aiCompositionTargetForPlan(signals.Plan)
	signals.BattleNeeds = aiBuildRecruitmentBattleNeeds(gs, self.ID, ctx)
	for _, region := range aiSortedRegions(gs) {
		if region.IsSea || region.OwnerID != string(self.ID) || gs.SiegeAt(region.ID) != nil {
			continue
		}
		production := gs.RegionProductionSummary(region)
		signals.IronProduction += production.Iron
		signals.TimberProduction += production.Timber
		signals.StoneProduction += production.Stone
		if region.Satisfaction < 60 {
			signals.LowSatisfaction++
			signals.SatisfactionDeficit += 60 - region.Satisfaction
		}
		if self.Religion != "" && region.Religion != "" && region.Religion != string(self.Religion) {
			signals.ReligionMismatch++
		}
		for _, neighborID := range region.Neighbors {
			if neighbor := gs.Regions[neighborID]; neighbor != nil && neighbor.IsSea {
				signals.CoastalRegions++
				break
			}
		}
	}
	return signals
}

func aiResearchTradeIncome(gs *state.GameState, fid faction.FactionID) int {
	total := 0
	for _, region := range aiSortedRegions(gs) {
		if region.IsSea || region.OwnerID != string(fid) || gs.SiegeAt(region.ID) != nil {
			continue
		}
		income := gs.BaseRegionTradeIncome(region)
		total += income * gs.CurrentSeason().TradeMod() / 100
	}
	// Ticaret gücü geliri, rota kurulmadan da merkez havuzundaki paydan gelir;
	// araştırma seçimi bu yeni ticaret mekanizmasını da gelecekteki getirinin
	// parçası olarak değerlendirmelidir.
	return total + gs.TradePowerCommerceIncome(fid)
}

func aiSortedTechnologyIDs(gs *state.GameState) []string {
	if gs == nil {
		return nil
	}
	ids := make([]string, 0, len(gs.TechTypes))
	for id := range gs.TechTypes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func aiResearchCandidateBetter(candidate, current aiResearchCandidate) bool {
	if candidate.Score != current.Score {
		return candidate.Score > current.Score
	}
	if candidate.UnlockScore != current.UnlockScore {
		return candidate.UnlockScore > current.UnlockScore
	}
	if candidate.EffectScore != current.EffectScore {
		return candidate.EffectScore > current.EffectScore
	}
	if candidate.FutureScore != current.FutureScore {
		return candidate.FutureScore > current.FutureScore
	}
	if candidate.DurationCost != current.DurationCost {
		return candidate.DurationCost < current.DurationCost
	}
	if candidate.GoldCost != current.GoldCost {
		return candidate.GoldCost < current.GoldCost
	}
	return candidate.Technology.ID < current.Technology.ID
}

func stringInList(value string, values []string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
