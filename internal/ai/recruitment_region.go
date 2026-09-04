package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	aiRecruitThroughputScore     = 250
	aiRecruitBuildingLevelScore  = 40
	aiRecruitLogisticsSpareScore = 8
	aiRecruitLogisticsSpareLimit = 40
	aiRecruitRouteCostPenalty    = 2
	aiRecruitLaneQueuePenalty    = 120
	aiRecruitOtherQueuePenalty   = 20
	aiRecruitAdjacentThreatBase  = 80
	aiRecruitAdjacentThreatScale = 3
	aiRecruitAdjacentThreatLimit = 400
)

type aiRecruitRegionCandidate struct {
	RegionID          world.RegionID
	Score             int
	RemainingCapacity int
	BuildingLevel     int
	LogisticsSpare    int
	RouteCost         int
	LaneQueue         int
	TotalQueue        int
	ThreatPenalty     int
}

func aiFindRecruitRegionForStrategicContext(gs *state.GameState, fid faction.FactionID, unitType *army.UnitType, ctx *StrategicContext) world.RegionID {
	return aiFindStrategicRecruitRegion(gs, fid, unitType, ctx)
}

func aiFindStrategicRecruitRegion(gs *state.GameState, fid faction.FactionID, unitType *army.UnitType, ctx *StrategicContext) world.RegionID {
	if gs == nil || unitType == nil || fid == "" {
		return ""
	}
	if ctx == nil {
		ctx = prepareStrategicContext(gs, fid)
	}
	anchor := aiRecruitmentRegionAnchor(gs, fid, unitType, ctx)

	var best aiRecruitRegionCandidate
	found := false
	for _, region := range aiSortedRegions(gs) {
		candidate, ok := aiScoreRecruitRegionCandidate(gs, fid, unitType, ctx, anchor, region)
		if !ok {
			continue
		}
		if !found || aiRecruitRegionCandidateBetter(candidate, best) {
			best = candidate
			found = true
		}
	}
	if !found {
		return ""
	}
	return best.RegionID
}

func aiScoreRecruitRegionCandidate(gs *state.GameState, fid faction.FactionID, unitType *army.UnitType, ctx *StrategicContext, anchor world.RegionID, region *world.Region) (aiRecruitRegionCandidate, bool) {
	if region == nil || region.IsSea || region.IsLocked || region.OwnerID != string(fid) {
		return aiRecruitRegionCandidate{}, false
	}
	requiredBuilding := unitType.RequiredBldg
	if requiredBuilding == "" {
		requiredBuilding = "barracks"
	}
	requiredLevel := maxInt(1, unitType.RequiredBldgLevel)
	buildingLevel := aiBuildingLevel(region, requiredBuilding)
	if buildingLevel < requiredLevel || !aiCanQueueLandUnit(gs, fid, region.ID, unitType) {
		return aiRecruitRegionCandidate{}, false
	}
	remainingCapacity := aiLaneRemainingCapacity(gs, region.ID, fid, unitType)
	if remainingCapacity <= 0 || aiPendingUnitCountByRegion(gs, region.ID, fid) >= aiMaxRegionQueue {
		return aiRecruitRegionCandidate{}, false
	}
	if !aiRecruitmentRegionSecure(gs, fid, ctx, region) {
		return aiRecruitRegionCandidate{}, false
	}

	demand, capacity, _ := aiProjectedRecruitRegionLogistics(gs, fid, region, unitType)
	if demand > capacity {
		return aiRecruitRegionCandidate{}, false
	}
	logisticsSpare := capacity - demand
	routeCost := 0
	if anchor != "" && anchor != region.ID {
		probe := &army.Army{
			ID:       army.ArmyID("ai_recruit_" + string(region.ID) + "_" + unitType.ID),
			OwnerID:  string(fid),
			RegionID: region.ID,
			Units:    []army.Unit{{TypeID: unitType.ID, CurrentHP: army.MaxUnitHP}},
		}
		routes := ctx.routesFor(probe, region.ID, aiRouteFriendly, 0)
		var reachable bool
		routeCost, reachable = routes.distance(anchor)
		if !reachable {
			return aiRecruitRegionCandidate{}, false
		}
	}

	laneQueue := aiPendingUnitCountByRegionInLane(gs, region.ID, fid, aiProductionLane(unitType))
	totalQueue := aiPendingUnitCountByRegion(gs, region.ID, fid)
	otherQueue := maxInt(0, totalQueue-laneQueue)
	threatPenalty := aiRecruitmentAdjacentThreatPenalty(gs, fid, region)
	score := remainingCapacity*aiRecruitThroughputScore + buildingLevel*aiRecruitBuildingLevelScore
	score += minInt(aiRecruitLogisticsSpareLimit, logisticsSpare) * aiRecruitLogisticsSpareScore
	score -= routeCost * aiRecruitRouteCostPenalty
	score -= laneQueue*aiRecruitLaneQueuePenalty + otherQueue*aiRecruitOtherQueuePenalty
	score -= threatPenalty

	return aiRecruitRegionCandidate{
		RegionID:          region.ID,
		Score:             score,
		RemainingCapacity: remainingCapacity,
		BuildingLevel:     buildingLevel,
		LogisticsSpare:    logisticsSpare,
		RouteCost:         routeCost,
		LaneQueue:         laneQueue,
		TotalQueue:        totalQueue,
		ThreatPenalty:     threatPenalty,
	}, true
}

func aiRecruitmentRegionSecure(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext, region *world.Region) bool {
	if gs == nil || region == nil || gs.SiegeAt(region.ID) != nil || region.IsRebellionRisk() {
		return false
	}
	for _, armyRef := range aiSortedArmies(gs) {
		if armyRef.IsNaval || armyRef.RegionID != region.ID || armyRef.OwnerID == string(fid) {
			continue
		}
		return false
	}
	if ctx == nil {
		return true
	}
	for _, front := range ctx.Fronts {
		if !front.CriticalThreat && !front.CapitalThreat {
			continue
		}
		if regionIDInList(region.ID, front.FriendlyRegions) {
			return false
		}
	}
	return true
}

func aiProjectedRecruitRegionLogistics(gs *state.GameState, fid faction.FactionID, region *world.Region, unitType *army.UnitType) (demand, capacity, overload int) {
	demand, capacity, _ = aiRegionLogistics(gs, region, string(fid))
	for _, order := range gs.ProductionQueue {
		if order.Kind != aiProductionKindUnit || order.FactionID != string(fid) || order.RegionID != region.ID {
			continue
		}
		pendingType := gs.UnitTypes[order.TypeID]
		if pendingType == nil || !aiLandUnitCategory(pendingType.Category) {
			continue
		}
		demand += pendingType.GrainUpkeep
	}
	if unitType != nil {
		demand += unitType.GrainUpkeep
	}
	return demand, capacity, maxInt(0, demand-capacity)
}

func aiRecruitmentAdjacentThreatPenalty(gs *state.GameState, fid faction.FactionID, region *world.Region) int {
	if gs == nil || region == nil {
		return 0
	}
	nearby := make(map[world.RegionID]struct{}, len(region.Neighbors))
	for _, neighborID := range region.Neighbors {
		nearby[neighborID] = struct{}{}
	}
	hostilePower := aiHostilePowerInRegions(gs, fid, nearby)
	if hostilePower <= 0 {
		return 0
	}
	return aiRecruitAdjacentThreatBase + minInt(aiRecruitAdjacentThreatLimit, hostilePower*aiRecruitAdjacentThreatScale)
}

func aiRecruitmentRegionAnchor(gs *state.GameState, fid faction.FactionID, unitType *army.UnitType, ctx *StrategicContext) world.RegionID {
	if gs == nil || ctx == nil {
		return ""
	}
	if ctx.RallyRegionID != "" && aiOwnedRecruitmentAnchor(gs, fid, ctx.RallyRegionID) {
		return ctx.RallyRegionID
	}
	plan := gs.AIPlans[fid]
	if unitType != nil && unitType.Category == army.CategorySiege {
		if anchor := aiMissingSiegeSupportAnchor(gs, fid, ctx); anchor != "" {
			return anchor
		}
	}
	if plan != nil && plan.Kind == state.AIObjectiveDefend {
		for _, regionID := range plan.TargetRegionIDs {
			if aiOwnedRecruitmentAnchor(gs, fid, regionID) {
				return regionID
			}
		}
	}
	if plan != nil && plan.TargetFactionID != "" {
		for _, front := range ctx.Fronts {
			if front.EnemyFactionID == plan.TargetFactionID && aiOwnedRecruitmentAnchor(gs, fid, front.AnchorRegionID) {
				return front.AnchorRegionID
			}
		}
	}
	for _, front := range ctx.Fronts {
		if (front.AtWar || front.ObjectiveRelated) && aiOwnedRecruitmentAnchor(gs, fid, front.AnchorRegionID) {
			return front.AnchorRegionID
		}
	}
	if capital, _, _, ok := gs.FactionCapital(fid); ok && capital != nil && aiOwnedRecruitmentAnchor(gs, fid, capital.ID) {
		return capital.ID
	}
	if len(ctx.OwnedLandRegionIDs) > 0 {
		return ctx.OwnedLandRegionIDs[0]
	}
	return ""
}

func aiMissingSiegeSupportAnchor(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext) world.RegionID {
	if gs == nil || ctx == nil {
		return ""
	}
	armyIDs := make([]army.ArmyID, 0, len(ctx.ArmyAssignments))
	for armyID := range ctx.ArmyAssignments {
		armyIDs = append(armyIDs, armyID)
	}
	sort.Slice(armyIDs, func(i, j int) bool { return armyIDs[i] < armyIDs[j] })
	for _, armyID := range armyIDs {
		assignment := ctx.ArmyAssignments[armyID]
		if assignment.Role != AIArmyRoleAssault && assignment.Role != AIArmyRoleSiege {
			continue
		}
		armyRef := gs.Armies[armyID]
		if armyRef == nil || armyRef.IsNaval || armyRef.IsGarrison || armyRef.HasSiegeUnits(gs.UnitTypes) {
			continue
		}
		if assignment.FrontFactionID != "" {
			for _, front := range ctx.Fronts {
				if front.EnemyFactionID == assignment.FrontFactionID && aiOwnedRecruitmentAnchor(gs, fid, front.AnchorRegionID) {
					return front.AnchorRegionID
				}
			}
		}
		if aiOwnedRecruitmentAnchor(gs, fid, armyRef.RegionID) {
			return armyRef.RegionID
		}
	}
	return ""
}

func aiOwnedRecruitmentAnchor(gs *state.GameState, fid faction.FactionID, regionID world.RegionID) bool {
	region := gs.Regions[regionID]
	return region != nil && !region.IsSea && region.OwnerID == string(fid)
}

func aiRecruitRegionCandidateBetter(candidate, current aiRecruitRegionCandidate) bool {
	if candidate.Score != current.Score {
		return candidate.Score > current.Score
	}
	if candidate.RouteCost != current.RouteCost {
		return candidate.RouteCost < current.RouteCost
	}
	if candidate.LogisticsSpare != current.LogisticsSpare {
		return candidate.LogisticsSpare > current.LogisticsSpare
	}
	if candidate.RemainingCapacity != current.RemainingCapacity {
		return candidate.RemainingCapacity > current.RemainingCapacity
	}
	if candidate.BuildingLevel != current.BuildingLevel {
		return candidate.BuildingLevel > current.BuildingLevel
	}
	if candidate.LaneQueue != current.LaneQueue {
		return candidate.LaneQueue < current.LaneQueue
	}
	return candidate.RegionID < current.RegionID
}
