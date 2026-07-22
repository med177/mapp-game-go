package ai

import (
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	aiBuildingProjectionTurns    = 12
	aiMinBuildingInvestmentScore = 80
)

var aiEconomyBuildingIDs = []string{"farm", "market", "temple", "walls"}

type aiBuildingCandidate struct {
	RegionID        world.RegionID
	BuildingID      string
	Cost            economy.ResourceCost
	Turns           int
	Score           int
	ROIScore        int
	BottleneckScore int
	ThreatScore     int
	ObjectiveScore  int
	StabilityScore  int
	QueuePenalty    int
}

type aiEconomySnapshot struct {
	GoldProduction  int
	GrainProduction int
	GrainUpkeep     int
	GrainStock      int
}

type aiRegionInvestmentSignals struct {
	Border           bool
	AtWar            bool
	Critical         bool
	Capital          bool
	ObjectiveOwned   bool
	ObjectiveStaging bool
	Rally            bool
	Threat           int
}

func aiEconomyBuildWithStrategicContextAndSteps(gs *state.GameState, fid faction.FactionID, budget *aiBudget, ctx *StrategicContext, steps *[]TurnStep) {
	if gs == nil {
		return
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated || gs.BuildingTypes == nil {
		return
	}
	if gs.ScenarioID != "1300_ottoman_rise" {
		aiLegacyEconomyBuildWithSteps(gs, fid, budget, steps)
		return
	}
	if ctx == nil {
		ctx = prepareStrategicContext(gs, fid)
	}

	candidate, ok := aiBestBuildingInvestment(gs, fid, budget, ctx)
	if !ok || !aiApplyBudgetedCost(self, candidate.Cost, budget, aiBudgetEconomy) {
		return
	}
	btype := gs.BuildingTypes[candidate.BuildingID]
	aiEnqueueProduction(gs, fid, aiProductionKindBuilding, candidate.RegionID, candidate.BuildingID, candidate.Turns)
	name := candidate.BuildingID
	if btype != nil && btype.NameTR != "" {
		name = btype.NameTR
	}
	addTurnStep(steps, TurnStep{
		FactionID:    fid,
		Kind:         TurnStepBuild,
		TargetRegion: candidate.RegionID,
		FocusRegion:  candidate.RegionID,
		Message:      turnFactionName(gs, fid) + " " + turnRegionName(gs, candidate.RegionID) + " bölgesinde " + name + " inşasını başlattı.",
	})
}

func aiBestBuildingInvestment(gs *state.GameState, fid faction.FactionID, budget *aiBudget, ctx *StrategicContext) (aiBuildingCandidate, bool) {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || gs.Factions[fid] == nil {
		return aiBuildingCandidate{}, false
	}
	self := gs.Factions[fid]
	snapshot := aiBuildEconomySnapshot(gs, fid)
	var best aiBuildingCandidate
	found := false
	for _, region := range aiSortedRegions(gs) {
		if region.IsSea || region.OwnerID != string(fid) || gs.SiegeAt(region.ID) != nil {
			continue
		}
		signals := aiInvestmentSignals(gs, fid, region, ctx)
		for _, buildingID := range aiEconomyBuildingIDs {
			btype := gs.BuildingTypes[buildingID]
			if btype == nil || !aiBuildingAllowed(gs, region, buildingID, btype.RequiredTerrain) {
				continue
			}
			queued := aiQueuedBuildingCount(gs, region.ID, buildingID, fid)
			level := aiBuildingLevel(region, buildingID)
			if btype.MaxPerRegion <= 0 || level+queued >= btype.MaxPerRegion {
				continue
			}
			if buildingID == "walls" && !signals.Border && !signals.Capital && !signals.ObjectiveOwned && !signals.Rally {
				continue
			}
			if buildingID == "temple" && region.Satisfaction >= 70 {
				continue
			}

			cost := aiBuildingResourceCost(btype)
			if !aiCanAffordForBudget(self, cost, budget, aiBudgetEconomy) {
				continue
			}
			turns := aiBuildingTurnsRequired(region, buildingID, btype.TurnsRequired, queued)
			candidate := aiScoreBuildingInvestment(gs, self, region, btype, cost, turns, level, queued, snapshot, signals)
			// Zayıf yatırım adayında ekonomi payını sırf harcamış olmak için tüketme;
			// bütçe serbest bırakılarak sonraki donanma/ordu kategorisine aktarılır.
			if candidate.Score < aiMinBuildingInvestmentScore {
				continue
			}
			if !found || aiBuildingCandidateBetter(candidate, best) {
				best = candidate
				found = true
			}
		}
	}
	return best, found
}

func aiScoreBuildingInvestment(gs *state.GameState, self *faction.Faction, region *world.Region, btype *city.Building, cost economy.ResourceCost, turns, level, queued int, snapshot aiEconomySnapshot, signals aiRegionInvestmentSignals) aiBuildingCandidate {
	before, after := aiBuildingMarginalProduction(gs, region, btype.ID)
	goldGain := maxInt(0, after.Gold-before.Gold)
	grainGain := maxInt(0, after.Grain-before.Grain)
	grainUtility := aiGrainUtilityPercent(snapshot)
	projectedValue := goldGain * aiBuildingProjectionTurns
	projectedValue += grainGain * aiResourcePrice(gs, economy.GoodGrain) * aiBuildingProjectionTurns * grainUtility / 100
	effectiveCost := aiGoldEquivalentCost(gs, cost)
	roiScore := projectedValue * 100 / maxInt(1, effectiveCost)
	if roiScore > 400 {
		roiScore = 400
	}

	bottleneckScore := aiBuildingBottleneckScore(self, btype.ID, cost, snapshot)
	threatScore := aiBuildingThreatScore(btype.ID, signals)
	objectiveScore := aiBuildingObjectiveScore(btype.ID, signals, gs.AIPlans[self.ID])
	stabilityNeed := maxInt(0, 70-region.Satisfaction)
	stabilityScore := btype.SatBonus * stabilityNeed / 2
	if btype.ID == "temple" && region.Satisfaction < 30 {
		stabilityScore += 180
	}
	queuePenalty := aiQueuedBuildingCountForRegion(gs, region.ID, self.ID)*30 + queued*25
	durationPenalty := maxInt(0, turns-1) * 8
	levelPenalty := level * 12

	return aiBuildingCandidate{
		RegionID:        region.ID,
		BuildingID:      btype.ID,
		Cost:            cost,
		Turns:           turns,
		Score:           roiScore + bottleneckScore + threatScore + objectiveScore + stabilityScore - queuePenalty - durationPenalty - levelPenalty,
		ROIScore:        roiScore,
		BottleneckScore: bottleneckScore,
		ThreatScore:     threatScore,
		ObjectiveScore:  objectiveScore,
		StabilityScore:  stabilityScore,
		QueuePenalty:    queuePenalty,
	}
}

func aiBuildEconomySnapshot(gs *state.GameState, fid faction.FactionID) aiEconomySnapshot {
	snapshot := aiEconomySnapshot{}
	if gs == nil {
		return snapshot
	}
	if self := gs.Factions[fid]; self != nil {
		snapshot.GrainStock = self.Grain
	}
	for _, region := range aiSortedRegions(gs) {
		if region.IsSea || region.OwnerID != string(fid) || gs.SiegeAt(region.ID) != nil {
			continue
		}
		production := gs.RegionProductionSummary(region)
		snapshot.GoldProduction += production.Gold
		snapshot.GrainProduction += production.Grain
	}
	for _, armyRef := range aiSortedArmies(gs) {
		if armyRef.OwnerID == string(fid) {
			snapshot.GrainUpkeep += gs.EffectiveArmyGrainUpkeep(armyRef)
		}
	}
	return snapshot
}

func aiGrainUtilityPercent(snapshot aiEconomySnapshot) int {
	target := maxInt(120, snapshot.GrainUpkeep*6)
	switch {
	case snapshot.GrainProduction <= snapshot.GrainUpkeep || snapshot.GrainStock < snapshot.GrainUpkeep*2:
		return 150
	case snapshot.GrainStock < target:
		return 100
	case snapshot.GrainStock < target*2:
		return 60
	default:
		return 25
	}
}

func aiBuildingBottleneckScore(self *faction.Faction, buildingID string, cost economy.ResourceCost, snapshot aiEconomySnapshot) int {
	score := 0
	if buildingID == "farm" {
		target := maxInt(120, snapshot.GrainUpkeep*6)
		switch {
		case snapshot.GrainProduction <= snapshot.GrainUpkeep || snapshot.GrainStock < snapshot.GrainUpkeep*2:
			score += 140
		case snapshot.GrainStock < target:
			score += 90
		case snapshot.GrainStock < target*2:
			score += 35
		}
	}
	if self == nil {
		return score
	}
	for _, resource := range []struct {
		stock int
		cost  int
	}{
		{self.Grain, cost.Grain},
		{self.Iron, cost.Iron},
		{self.Timber, cost.Timber},
		{self.Stone, cost.Stone},
	} {
		if resource.cost <= 0 {
			continue
		}
		if resource.stock < resource.cost*2 {
			score -= 30
		} else if resource.stock < resource.cost*4 {
			score -= 12
		}
	}
	return score
}

func aiBuildingThreatScore(buildingID string, signals aiRegionInvestmentSignals) int {
	if buildingID != "walls" {
		score := 0
		if signals.AtWar {
			score -= 20
		}
		if signals.Critical {
			score -= 40
		}
		return score
	}
	score := 0
	if signals.Border {
		score += 20
	}
	if signals.AtWar {
		score += 100
	}
	if signals.Threat > 0 {
		score += minInt(100, signals.Threat/2)
	}
	if signals.Critical {
		score += 140
	}
	if signals.Capital {
		score += 50
	}
	return score
}

func aiBuildingObjectiveScore(buildingID string, signals aiRegionInvestmentSignals, plan *state.AIPlanState) int {
	if plan == nil {
		return 0
	}
	relevant := signals.ObjectiveOwned || signals.ObjectiveStaging || signals.Rally
	switch plan.Kind {
	case state.AIObjectiveExpand:
		if !relevant {
			return 0
		}
		switch buildingID {
		case "farm":
			return 60
		case "market":
			return 35
		case "walls":
			return 30
		case "temple":
			return 20
		}
	case state.AIObjectiveDefend:
		if !relevant && !signals.Critical {
			return 0
		}
		switch buildingID {
		case "walls":
			return 150
		case "farm":
			return 30
		case "temple":
			return 40
		case "market":
			return 10
		}
	case state.AIObjectiveConsolidate:
		switch buildingID {
		case "market":
			return 70
		case "farm":
			return 40
		case "temple":
			return 50
		}
	}
	return 0
}

func aiInvestmentSignals(gs *state.GameState, fid faction.FactionID, region *world.Region, ctx *StrategicContext) aiRegionInvestmentSignals {
	signals := aiRegionInvestmentSignals{}
	if gs == nil || region == nil {
		return signals
	}
	if capital, _, _, ok := gs.FactionCapital(fid); ok && capital != nil && capital.ID == region.ID {
		signals.Capital = true
	}
	plan := gs.AIPlans[fid]
	if plan != nil {
		for _, targetID := range plan.TargetRegionIDs {
			if targetID == region.ID {
				signals.ObjectiveOwned = true
			}
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil || neighbor.IsSea {
				continue
			}
			if neighbor.OwnerID == string(plan.TargetFactionID) || regionIDInList(neighbor.ID, plan.TargetRegionIDs) {
				signals.ObjectiveStaging = true
			}
		}
		if plan.RallyRegionID == region.ID {
			signals.Rally = true
		}
	}
	for _, neighborID := range region.Neighbors {
		neighbor := gs.Regions[neighborID]
		if neighbor != nil && !neighbor.IsSea && neighbor.OwnerID != "" && neighbor.OwnerID != string(fid) && !diplomacy.SameRealm(gs, fid, faction.FactionID(neighbor.OwnerID)) {
			signals.Border = true
		}
	}
	if ctx == nil {
		return signals
	}
	for _, front := range ctx.Fronts {
		if !regionIDInList(region.ID, front.FriendlyRegions) {
			continue
		}
		signals.Border = true
		signals.AtWar = signals.AtWar || front.AtWar
		signals.Critical = signals.Critical || front.CriticalThreat || front.CapitalThreat
		signals.ObjectiveStaging = signals.ObjectiveStaging || front.ObjectiveRelated
		if front.ThreatScore > signals.Threat {
			signals.Threat = front.ThreatScore
		}
	}
	return signals
}

func regionIDInList(regionID world.RegionID, values []world.RegionID) bool {
	for _, value := range values {
		if value == regionID {
			return true
		}
	}
	return false
}

func aiBuildingMarginalProduction(gs *state.GameState, region *world.Region, buildingID string) (state.RegionProductionSummary, state.RegionProductionSummary) {
	before := gs.RegionProductionSummary(region)
	clone := *region
	clone.Buildings = append(append([]string(nil), region.Buildings...), buildingID)
	return before, gs.RegionProductionSummary(&clone)
}

func aiGoldEquivalentCost(gs *state.GameState, cost economy.ResourceCost) int {
	return cost.Gold +
		cost.Grain*aiResourcePrice(gs, economy.GoodGrain) +
		cost.Iron*aiResourcePrice(gs, economy.GoodIron) +
		cost.Timber*aiResourcePrice(gs, economy.GoodTimber) +
		cost.Stone*aiResourcePrice(gs, economy.GoodStone)
}

func aiResourcePrice(gs *state.GameState, good economy.GoodType) int {
	if gs != nil && gs.MarketPrices != nil && gs.MarketPrices[good] > 0 {
		return gs.MarketPrices[good]
	}
	return maxInt(1, economy.BaseGoldValue[good])
}

func aiBuildingResourceCost(building *city.Building) economy.ResourceCost {
	if building == nil {
		return economy.ResourceCost{}
	}
	return economy.ResourceCost{
		Gold: building.GoldCost, Grain: building.GrainCost, Iron: building.IronCost,
		Timber: building.TimberCost, Stone: building.StoneCost,
	}
}

func aiQueuedBuildingCountForRegion(gs *state.GameState, regionID world.RegionID, fid faction.FactionID) int {
	count := 0
	for _, order := range gs.ProductionQueue {
		if order.Kind == aiProductionKindBuilding && order.RegionID == regionID && order.FactionID == string(fid) {
			count++
		}
	}
	return count
}

func aiBuildingCandidateBetter(candidate, best aiBuildingCandidate) bool {
	if candidate.Score != best.Score {
		return candidate.Score > best.Score
	}
	if candidate.ROIScore != best.ROIScore {
		return candidate.ROIScore > best.ROIScore
	}
	if candidate.ObjectiveScore != best.ObjectiveScore {
		return candidate.ObjectiveScore > best.ObjectiveScore
	}
	if candidate.ThreatScore != best.ThreatScore {
		return candidate.ThreatScore > best.ThreatScore
	}
	if candidate.Turns != best.Turns {
		return candidate.Turns < best.Turns
	}
	if candidate.RegionID != best.RegionID {
		return candidate.RegionID < best.RegionID
	}
	return candidate.BuildingID < best.BuildingID
}

func aiLegacyEconomyBuildWithSteps(gs *state.GameState, fid faction.FactionID, budget *aiBudget, steps *[]TurnStep) {
	self := gs.Factions[fid]
	if self == nil || (budget == nil && self.Gold < 200+aiMinGoldReserve) {
		return
	}
	for _, buildingID := range []string{"farm", "market", "walls"} {
		btype := gs.BuildingTypes[buildingID]
		if btype == nil {
			continue
		}
		cost := aiBuildingResourceCost(btype)
		if !aiCanAffordForBudget(self, cost, budget, aiBudgetEconomy) {
			continue
		}
		for _, region := range aiSortedRegions(gs) {
			if region.IsSea || region.OwnerID != string(fid) || !aiBuildingAllowed(gs, region, buildingID, btype.RequiredTerrain) {
				continue
			}
			queued := aiQueuedBuildingCount(gs, region.ID, buildingID, fid)
			if btype.MaxPerRegion <= 0 || aiBuildingLevel(region, buildingID)+queued >= btype.MaxPerRegion || !aiLegacyBuildingNeeded(gs, fid, region, buildingID) {
				continue
			}
			if !aiApplyBudgetedCost(self, cost, budget, aiBudgetEconomy) {
				continue
			}
			turns := aiBuildingTurnsRequired(region, buildingID, btype.TurnsRequired, queued)
			aiEnqueueProduction(gs, fid, aiProductionKindBuilding, region.ID, buildingID, turns)
			addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepBuild, TargetRegion: region.ID, FocusRegion: region.ID,
				Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, region.ID) + " bölgesinde " + btype.NameTR + " inşasını başlattı."})
			return
		}
	}
}

func aiLegacyBuildingNeeded(gs *state.GameState, fid faction.FactionID, region *world.Region, buildingID string) bool {
	switch buildingID {
	case "farm":
		return region.BaseGrainOutput < 20
	case "market":
		return true
	case "walls":
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor != nil && neighbor.OwnerID != "" && neighbor.OwnerID != string(fid) {
				return true
			}
		}
	}
	return false
}
