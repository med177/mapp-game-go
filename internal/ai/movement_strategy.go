package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

type aiLogisticsSnapshot struct {
	demand   int
	capacity int
	overload int
}

// moveScoreContext caches the expensive state scans used while evaluating one
// army's movement targets. It is discarded after the movement is applied.
type moveScoreContext struct {
	ownerID        string
	atCapacity     bool
	armiesByRegion map[world.RegionID][]*army.Army
	logistics      map[world.RegionID]aiLogisticsSnapshot
}

func newMoveScoreContext(gs *state.GameState, a *army.Army) *moveScoreContext {
	ctx := &moveScoreContext{
		armiesByRegion: make(map[world.RegionID][]*army.Army),
		logistics:      make(map[world.RegionID]aiLogisticsSnapshot),
	}
	if gs == nil || a == nil {
		return ctx
	}

	ctx.ownerID = a.OwnerID
	allArmies := make([]*army.Army, 0, len(gs.Armies))
	deployed := 0
	for _, candidate := range gs.Armies {
		if candidate == nil {
			continue
		}
		allArmies = append(allArmies, candidate)
		if candidate.OwnerID == a.OwnerID && !candidate.IsNaval {
			deployed += len(candidate.Units)
		}
	}
	sort.Slice(allArmies, func(i, j int) bool { return allArmies[i].ID < allArmies[j].ID })
	for _, candidate := range allArmies {
		ctx.armiesByRegion[candidate.RegionID] = append(ctx.armiesByRegion[candidate.RegionID], candidate)
	}
	ctx.atCapacity = deployed >= gs.ManpowerCap(faction.FactionID(a.OwnerID))
	return ctx
}

func (ctx *moveScoreContext) findEmbarkFleet(gs *state.GameState, seaRegionID world.RegionID, unitCount int) *army.Army {
	if ctx == nil || gs == nil {
		return nil
	}
	for _, candidate := range ctx.armiesByRegion[seaRegionID] {
		if candidate.OwnerID == ctx.ownerID && candidate.IsNaval && candidate.CanEmbarkUnits(gs.UnitTypes, unitCount) {
			return candidate
		}
	}
	return nil
}

func (ctx *moveScoreContext) regionLogistics(gs *state.GameState, region *world.Region) (demand, capacity, overload int) {
	if ctx == nil || gs == nil || region == nil || region.IsSea || ctx.ownerID == "" {
		return 0, 0, 0
	}
	if cached, ok := ctx.logistics[region.ID]; ok {
		return cached.demand, cached.capacity, cached.overload
	}

	militaryProduction := gs.RegionMilitaryGrainProduction(region)
	if region.OwnerID != ctx.ownerID {
		militaryProduction = 0
	}
	settlementBuffer := aiRegionSettlementBuffer(gs, region)
	blockadePercent := gs.RegionBlockadePercent(region, ctx.ownerID)
	settlementBuffer = settlementBuffer * (100 - blockadePercent) / 100
	reserveSupport := aiRegionReserveSupport(gs, ctx.ownerID, militaryProduction, settlementBuffer)
	capacity = militaryProduction + settlementBuffer + reserveSupport
	if capacity < 4 {
		capacity = 4
	}
	for _, candidate := range ctx.armiesByRegion[region.ID] {
		if candidate.OwnerID == ctx.ownerID && !candidate.IsNaval {
			demand += gs.EffectiveArmyGrainUpkeep(candidate)
		}
	}
	overload = maxInt(0, demand-capacity)
	ctx.logistics[region.ID] = aiLogisticsSnapshot{demand: demand, capacity: capacity, overload: overload}
	return demand, capacity, overload
}

func aiEmbarkScore(gs *state.GameState, a *army.Army, seaRegion *world.Region) int {
	return aiEmbarkScoreWithContext(gs, a, seaRegion, newMoveScoreContext(gs, a))
}

func aiEmbarkScoreWithContext(gs *state.GameState, a *army.Army, seaRegion *world.Region, ctx *moveScoreContext) int {
	if gs == nil || a == nil || seaRegion == nil || !seaRegion.IsSea {
		return 0
	}
	if !aiCanEmbarkArmy(gs, a) || ctx.findEmbarkFleet(gs, seaRegion.ID, len(a.Units)) == nil {
		return 0
	}
	best := 10 + aiSeaPressure(gs, a.OwnerID, seaRegion.ID)/2
	for _, nid := range seaRegion.Neighbors {
		land, ok := gs.Regions[nid]
		if !ok || land.IsSea {
			continue
		}
		score := scoreMoveWithContext(gs, a, land, ctx)
		if score > best {
			best = score
		}
	}
	return best
}

// strategicContextAssignment returns the planned role assignment for an army
// when strategic planning has already produced one. Keeping this lookup next
// to route selection makes movement decisions independent from the turn loop.
func strategicContextAssignment(ctx *StrategicContext, armyID army.ArmyID) (AIArmyAssignment, bool) {
	if ctx == nil || ctx.ArmyAssignments == nil {
		return AIArmyAssignment{}, false
	}
	assignment, ok := ctx.ArmyAssignments[armyID]
	return assignment, ok
}

// chooseBestMove selects the highest-scoring legal neighboring step and falls
// back to long-range route planning when no adjacent target is useful.
func chooseBestMove(gs *state.GameState, a *army.Army) world.RegionID {
	return chooseBestMoveWithStrategicContext(gs, a, nil)
}

func chooseBestMoveWithStrategicContext(gs *state.GameState, a *army.Army, strategicContext *StrategicContext) world.RegionID {
	src, ok := gs.Regions[a.RegionID]
	if !ok {
		return ""
	}
	if assignment, assigned := strategicContextAssignment(strategicContext, a.ID); assigned {
		switch assignment.Role {
		case AIArmyRoleRetreat:
			return aiRetreatNextStep(strategicContext, a)
		case AIArmyRoleSecurity:
			return aiSecurityNextStep(strategicContext, a)
		}
	}
	if activeSiege := gs.SiegeByArmy(a.ID); activeSiege != nil {
		target := gs.Regions[activeSiege.RegionID]
		if !aiCanStartSiege(gs, a, target) {
			return ""
		}
		if activeSiege.BreachLevel >= 2 {
			return activeSiege.RegionID
		}
		return ""
	}
	if a.IsNaval {
		if target, handled := aiNavalMissionMove(gs, a, strategicContext); handled {
			return target
		}
	} else if target, handled := aiNavalEmbarkArmyMove(gs, a, strategicContext); handled {
		return target
	}

	bestScore := 0
	var bestTarget world.RegionID
	if a.IsNaval {
		for _, nid := range src.Neighbors {
			n, ok := gs.Regions[nid]
			if !ok {
				continue
			}
			if n.IsSea {
				score := 15 + aiSeaPressure(gs, a.OwnerID, n.ID)
				if len(a.EmbarkedUnits) > 0 {
					for _, landID := range n.Neighbors {
						land, ok := gs.Regions[landID]
						if !ok || land.IsSea {
							continue
						}
						if land.OwnerID != "" && land.OwnerID != a.OwnerID {
							score += 20
						}
					}
				}
				if score > bestScore {
					bestScore = score
					bestTarget = nid
				}
				continue
			}
			if !aiCanDisembarkToLand(gs, a, n) {
				continue
			}
			score := 40
			enemyArmy := aiEnemyArmyInRegion(gs, a.OwnerID, n.ID)
			if enemyArmy != nil {
				landingStr := aiLandingStrength(gs, a)
				defStr := enemyArmy.TotalStrength(gs.UnitTypes)
				if landingStr <= defStr {
					continue
				}
				score = 75
			} else if n.OwnerID != "" && n.OwnerID != a.OwnerID {
				score = 60
			}
			if score > bestScore {
				bestScore = score
				bestTarget = nid
			}
		}
		return bestTarget
	}
	ctx := newMoveScoreContext(gs, a)
	for _, nid := range src.Neighbors {
		n, ok := gs.Regions[nid]
		if !ok {
			continue
		}
		if n.IsSea {
			score := aiEmbarkScoreWithContext(gs, a, n, ctx)
			if score > bestScore {
				bestScore = score
				bestTarget = nid
			}
			continue
		}
		score := scoreMoveWithContext(gs, a, n, ctx)
		score = aiRoleAdjustedMoveScore(strategicContext, a, n, score)
		if score > bestScore {
			bestScore = score
			bestTarget = nid
		}
	}
	if bestScore == 0 {
		bestTarget = findLongRangeMoveWithStrategicContext(gs, a, src, ctx, strategicContext)
	}
	return bestTarget
}

// scoreMove rates a legal movement target for logistics, diplomacy, combat and
// conquest pressure while leaving the actual state mutation to executeMove.
func scoreMove(gs *state.GameState, a *army.Army, target *world.Region) int {
	return scoreMoveWithContext(gs, a, target, newMoveScoreContext(gs, a))
}

func scoreMoveWithContext(gs *state.GameState, a *army.Army, target *world.Region, ctx *moveScoreContext) int {
	source := gs.Regions[a.RegionID]
	planBonus := aiPlanMoveScoreBonus(gs, faction.FactionID(a.OwnerID), target)
	armyDemand := gs.EffectiveArmyGrainUpkeep(a)
	if target.OwnerID == a.OwnerID {
		score := 0
		srcDemand, srcCap, srcOverload := ctx.regionLogistics(gs, source)
		tgtDemand, tgtCap, tgtOverload := ctx.regionLogistics(gs, target)
		srcAfter := maxInt(0, srcDemand-armyDemand-srcCap)
		tgtAfter := maxInt(0, tgtDemand+armyDemand-tgtCap)

		if srcOverload > 0 || a.OverCapacityTurns > 0 {
			relief := srcOverload - srcAfter
			if relief > 0 {
				score += aiReliefMoveBase + relief*2
			}
			if tgtAfter == 0 {
				score += 18
			}
			if tgtOverload == 0 {
				spare := tgtCap - tgtDemand
				if spare > 0 {
					score += minInt(18, spare)
				}
			}
			if tgtAfter > srcAfter {
				score -= 45
			}
		}

		for _, ea := range ctx.armiesByRegion[target.ID] {
			if ea.RegionID == target.ID && ea.OwnerID == a.OwnerID && ea.ID != a.ID && ea.IsNaval == a.IsNaval {
				if len(a.Units)+len(ea.Units) <= army.MaxArmySize && aiShouldConsolidateInRegion(gs, target, a.OwnerID, a.IsNaval) && score < 60 {
					score = 60
				}
			}
		}
		return score + planBonus
	}

	if target.OwnerID != "" {
		_, stance := relationScore(gs, a.OwnerID, target.OwnerID)
		if !a.IsNaval && target.CanLandEnter() && target.IsFortified() {
			if siege := gs.SiegeAt(target.ID); siege != nil && siege.AttackerArmyID != a.ID {
				if gs.CanJoinActiveSiege(a, target.ID) {
					_, _, srcOverload := ctx.regionLogistics(gs, source)
					if srcOverload > 0 || a.OverCapacityTurns > 0 {
						tgtDemand, tgtCap, tgtOverload := ctx.regionLogistics(gs, target)
						if tgtDemand+armyDemand <= tgtCap && tgtOverload == 0 {
							return aiReliefMoveBase
						}
					}
					return 5
				}
				if !gs.CanEnterActiveSiegedRegion(a, target.ID) {
					return -1
				}
				enemyArmy := aiEnemyArmyInRegion(gs, a.OwnerID, target.ID)
				if enemyArmy != nil {
					_, enemyStance := relationScore(gs, a.OwnerID, enemyArmy.OwnerID)
					if enemyStance == faction.StanceWar && a.TotalStrength(gs.UnitTypes) > enemyArmy.TotalStrength(gs.UnitTypes) {
						if stance == faction.StanceAllied {
							return 65
						}
						return 95 + planBonus
					}
				}
				return -1
			}
		}
		if stance == faction.StanceAllied {
			_, _, srcOverload := ctx.regionLogistics(gs, source)
			if srcOverload > 0 || a.OverCapacityTurns > 0 {
				tgtDemand, tgtCap, tgtOverload := ctx.regionLogistics(gs, target)
				if tgtDemand+armyDemand <= tgtCap && tgtOverload == 0 {
					return aiReliefMoveBase
				}
			}
			return 5
		}
		if stance != faction.StanceWar {
			return -1
		}
	}
	if !a.IsNaval && target.CanLandEnter() && target.OwnerID != "" && target.OwnerID != a.OwnerID && target.IsFortified() {
		if siege := gs.SiegeAt(target.ID); siege != nil && siege.AttackerArmyID != a.ID {
			return -1
		}
		if ctx.atCapacity {
			return 100 + planBonus
		}
		return 92 + planBonus
	}

	atCapacity := ctx.atCapacity
	for _, ea := range ctx.armiesByRegion[target.ID] {
		if ea.RegionID != target.ID || ea.OwnerID == a.OwnerID {
			continue
		}
		if a.TotalStrength(gs.UnitTypes) > ea.TotalStrength(gs.UnitTypes) {
			_, stance := relationScore(gs, a.OwnerID, target.OwnerID)
			if stance == faction.StanceWar {
				return 95 + planBonus
			}
			return 75
		}
		return -1
	}
	if target.OwnerID == "" {
		if atCapacity {
			return 70
		}
		return 50
	}
	_, stance := relationScore(gs, a.OwnerID, target.OwnerID)
	if stance == faction.StanceWar {
		if atCapacity {
			return 100 + planBonus
		}
		return 90 + planBonus
	}
	if stance == faction.StanceAllied {
		return 10
	}
	return -1
}

// findLongRangeMove uses weighted routes for the 1300 scenario and preserves
// the legacy BFS route selection for other scenarios.
func findLongRangeMove(gs *state.GameState, a *army.Army, start *world.Region) world.RegionID {
	return findLongRangeMoveWithContext(gs, a, start, newMoveScoreContext(gs, a))
}

func findLongRangeMoveWithContext(gs *state.GameState, a *army.Army, start *world.Region, ctx *moveScoreContext) world.RegionID {
	return findLongRangeMoveWithStrategicContext(gs, a, start, ctx, nil)
}

func findLongRangeMoveWithStrategicContext(gs *state.GameState, a *army.Army, start *world.Region, ctx *moveScoreContext, strategicContext *StrategicContext) world.RegionID {
	if gs != nil && gs.ScenarioID == "1300_ottoman_rise" {
		return findWeightedLongRangeMove(gs, a, start, ctx, strategicContext)
	}
	return findLongRangeMoveBFS(gs, a, start, ctx, strategicContext)
}

func findWeightedLongRangeMove(gs *state.GameState, a *army.Army, start *world.Region, ctx *moveScoreContext, strategicContext *StrategicContext) world.RegionID {
	if gs == nil || a == nil || start == nil {
		return ""
	}
	maxHops := aiPathSearchDepth(gs)
	var routes *aiRouteMap
	if strategicContext != nil {
		routes = strategicContext.routesFor(a, start.ID, aiRouteGeneral, maxHops)
	} else {
		routes = aiWeightedLandRoutes(gs, a, start.ID, aiRouteGeneral, maxHops, ctx)
	}

	bestID := world.RegionID("")
	bestCost := int(^uint(0) >> 1)
	bestScore := 0
	for _, region := range aiSortedRegions(gs) {
		if region.ID == start.ID {
			continue
		}
		cost, reachable := routes.distance(region.ID)
		if !reachable || routes.nextStep(region.ID) == "" {
			continue
		}
		score := scoreMoveWithContext(gs, a, region, ctx)
		score = aiRoleAdjustedMoveScore(strategicContext, a, region, score)
		if score <= 0 {
			continue
		}
		if cost < bestCost || (cost == bestCost && (score > bestScore || (score == bestScore && (bestID == "" || region.ID < bestID)))) {
			bestID = region.ID
			bestCost = cost
			bestScore = score
		}
	}
	return routes.nextStep(bestID)
}

func findLongRangeMoveBFS(gs *state.GameState, a *army.Army, start *world.Region, ctx *moveScoreContext, strategicContext *StrategicContext) world.RegionID {
	type queueItem struct {
		id   world.RegionID
		path []world.RegionID
	}

	visited := make(map[world.RegionID]bool)
	queue := []queueItem{{id: start.ID, path: nil}}
	visited[start.ID] = true

	maxDepth := aiPathSearchDepth(gs)

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if len(curr.path) > maxDepth {
			continue
		}

		r, ok := gs.Regions[curr.id]
		if !ok {
			continue
		}

		// Kendi bölgesi değilse ve score > 0 ise hedef bulduk demektir.
		if curr.id != start.ID {
			score := scoreMoveWithContext(gs, a, r, ctx)
			score = aiRoleAdjustedMoveScore(strategicContext, a, r, score)
			if score > 0 {
				return curr.path[0]
			}
			// Düşman toprağıysa daha ileri gitme.
			if r.OwnerID != a.OwnerID && r.OwnerID != "" {
				continue
			}
		}

		for _, nid := range r.Neighbors {
			n, ok := gs.Regions[nid]
			if !ok || n.IsSea || visited[nid] {
				continue
			}
			visited[nid] = true

			newPath := make([]world.RegionID, len(curr.path))
			copy(newPath, curr.path)
			newPath = append(newPath, nid)

			queue = append(queue, queueItem{id: nid, path: newPath})
		}
	}
	return ""
}
