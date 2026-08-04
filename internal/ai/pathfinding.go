package ai

import (
	"container/heap"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	aiRouteTerrainCostScale    = 10
	aiRouteAlliedAccessCost    = 5
	aiRouteTerminalAccessCost  = 10
	aiRouteSiegeCost           = 20
	aiRouteThreatCostAtParity  = 40
	aiRouteThreatCostLimit     = 80
	aiRouteLogisticsCostPerGap = 6
	aiRouteLogisticsCostLimit  = 60
)

type aiRouteMode uint8

const (
	aiRouteGeneral aiRouteMode = iota
	aiRouteFriendly
)

type aiRouteCacheKey struct {
	armyID  army.ArmyID
	start   world.RegionID
	mode    aiRouteMode
	maxHops int
}

type aiRouteMap struct {
	cost      map[world.RegionID]int
	hops      map[world.RegionID]int
	firstStep map[world.RegionID]world.RegionID
}

func (routes *aiRouteMap) distance(regionID world.RegionID) (int, bool) {
	if routes == nil {
		return 0, false
	}
	cost, ok := routes.cost[regionID]
	return cost, ok
}

func (routes *aiRouteMap) hopCount(regionID world.RegionID) (int, bool) {
	if routes == nil {
		return 0, false
	}
	hops, ok := routes.hops[regionID]
	return hops, ok
}

func (routes *aiRouteMap) nextStep(regionID world.RegionID) world.RegionID {
	if routes == nil {
		return ""
	}
	return routes.firstStep[regionID]
}

type aiRouteSnapshot struct {
	gs                 *state.GameState
	armyRef            *army.Army
	ownerID            faction.FactionID
	mode               aiRouteMode
	moveScore          *moveScoreContext
	hostilePower       map[world.RegionID]int
	hostileArmyPresent map[world.RegionID]bool
	foreignArmyPresent map[world.RegionID]bool
}

func newAIRouteSnapshot(gs *state.GameState, armyRef *army.Army, mode aiRouteMode, moveScore *moveScoreContext) *aiRouteSnapshot {
	snapshot := &aiRouteSnapshot{
		gs:                 gs,
		armyRef:            armyRef,
		mode:               mode,
		moveScore:          moveScore,
		hostilePower:       make(map[world.RegionID]int),
		hostileArmyPresent: make(map[world.RegionID]bool),
		foreignArmyPresent: make(map[world.RegionID]bool),
	}
	if gs == nil || armyRef == nil {
		return snapshot
	}
	snapshot.ownerID = faction.FactionID(armyRef.OwnerID)
	if snapshot.moveScore == nil {
		snapshot.moveScore = newMoveScoreContext(gs, armyRef)
	}
	for _, candidate := range aiSortedArmies(gs) {
		if candidate.IsNaval || len(candidate.Units) == 0 || candidate.OwnerID == armyRef.OwnerID {
			continue
		}
		snapshot.foreignArmyPresent[candidate.RegionID] = true
		if diplomacy.SameRealm(gs, snapshot.ownerID, faction.FactionID(candidate.OwnerID)) {
			continue
		}
		relation := diplomacy.Relation(gs, snapshot.ownerID, faction.FactionID(candidate.OwnerID))
		if relation == nil || relation.Stance != faction.StanceWar {
			continue
		}
		snapshot.hostileArmyPresent[candidate.RegionID] = true
		power := candidate.TotalStrength(gs.UnitTypes)
		snapshot.hostilePower[candidate.RegionID] += power
		region := gs.Regions[candidate.RegionID]
		if region == nil {
			continue
		}
		for _, neighborID := range region.Neighbors {
			snapshot.hostilePower[neighborID] += power
		}
	}
	return snapshot
}

func (snapshot *aiRouteSnapshot) regionAccess(region *world.Region) (allowed, transit bool, penalty int) {
	if snapshot == nil || snapshot.gs == nil || snapshot.armyRef == nil || region == nil || !region.CanLandEnter() {
		return false, false, 0
	}
	if props, ok := aiRouteTerrainProps(region); !ok || !props.Passable {
		return false, false, 0
	}
	if snapshot.mode == aiRouteFriendly {
		if region.OwnerID != snapshot.armyRef.OwnerID || snapshot.gs.SiegeAt(region.ID) != nil || snapshot.foreignArmyPresent[region.ID] {
			return false, false, 0
		}
		return true, true, 0
	}

	terminal := false
	if snapshot.hostileArmyPresent[region.ID] {
		terminal = true
	}
	if siege := snapshot.gs.SiegeAt(region.ID); siege != nil && siege.AttackerArmyID != snapshot.armyRef.ID {
		if !snapshot.gs.CanJoinActiveSiege(snapshot.armyRef, region.ID) && !snapshot.gs.CanEnterActiveSiegedRegion(snapshot.armyRef, region.ID) {
			return false, false, 0
		}
		penalty += aiRouteSiegeCost
		terminal = true
	}

	switch {
	case region.OwnerID == snapshot.armyRef.OwnerID:
		if terminal {
			return true, false, penalty
		}
		return true, true, penalty
	case region.OwnerID == "":
		return true, false, penalty + aiRouteTerminalAccessCost
	case diplomacy.SameRealm(snapshot.gs, snapshot.ownerID, faction.FactionID(region.OwnerID)):
		if terminal {
			return true, false, penalty + aiRouteAlliedAccessCost
		}
		return true, true, penalty + aiRouteAlliedAccessCost
	}

	relation := diplomacy.Relation(snapshot.gs, snapshot.ownerID, faction.FactionID(region.OwnerID))
	if relation == nil {
		return false, false, 0
	}
	switch relation.Stance {
	case faction.StanceAllied:
		if terminal {
			return true, false, penalty + aiRouteAlliedAccessCost
		}
		return true, true, penalty + aiRouteAlliedAccessCost
	case faction.StanceWar:
		return true, false, penalty + aiRouteTerminalAccessCost
	default:
		return false, false, 0
	}
}

func aiRouteTerrainProps(region *world.Region) (world.TerrainProps, bool) {
	if region == nil {
		return world.TerrainProps{}, false
	}
	terrain := region.Terrain
	if terrain == "" {
		terrain = world.TerrainPlain
	}
	props, ok := world.TerrainData[terrain]
	return props, ok
}

func (snapshot *aiRouteSnapshot) entryCost(region *world.Region) (int, bool) {
	allowed, _, accessCost := snapshot.regionAccess(region)
	if !allowed {
		return 0, false
	}
	props, _ := aiRouteTerrainProps(region)
	cost := maxInt(1, props.MoveCost) * aiRouteTerrainCostScale
	cost += accessCost

	armyPower := maxInt(1, snapshot.armyRef.TotalStrength(snapshot.gs.UnitTypes))
	threatCost := (snapshot.hostilePower[region.ID]*aiRouteThreatCostAtParity + armyPower - 1) / armyPower
	cost += minInt(aiRouteThreatCostLimit, threatCost)

	if region.OwnerID == snapshot.armyRef.OwnerID && snapshot.moveScore != nil {
		demand, capacity, _ := snapshot.moveScore.regionLogistics(snapshot.gs, region)
		if region.ID != snapshot.armyRef.RegionID {
			demand += snapshot.gs.RegionalArmyGrainDemand(snapshot.armyRef)
		}
		gap := maxInt(0, demand-capacity)
		cost += minInt(aiRouteLogisticsCostLimit, gap*aiRouteLogisticsCostPerGap)
	}
	return cost, true
}

type aiRouteQueueItem struct {
	regionID world.RegionID
	cost     int
	hops     int
	first    world.RegionID
}

type aiRouteQueue []aiRouteQueueItem

func (queue aiRouteQueue) Len() int { return len(queue) }
func (queue aiRouteQueue) Less(i, j int) bool {
	if queue[i].cost != queue[j].cost {
		return queue[i].cost < queue[j].cost
	}
	if queue[i].hops != queue[j].hops {
		return queue[i].hops < queue[j].hops
	}
	if queue[i].regionID != queue[j].regionID {
		return queue[i].regionID < queue[j].regionID
	}
	return queue[i].first < queue[j].first
}
func (queue aiRouteQueue) Swap(i, j int)   { queue[i], queue[j] = queue[j], queue[i] }
func (queue *aiRouteQueue) Push(value any) { *queue = append(*queue, value.(aiRouteQueueItem)) }
func (queue *aiRouteQueue) Pop() any {
	old := *queue
	last := len(old) - 1
	item := old[last]
	*queue = old[:last]
	return item
}

func aiWeightedLandRoutes(gs *state.GameState, armyRef *army.Army, start world.RegionID, mode aiRouteMode, maxHops int, moveScore *moveScoreContext) *aiRouteMap {
	routes := &aiRouteMap{
		cost:      make(map[world.RegionID]int),
		hops:      make(map[world.RegionID]int),
		firstStep: make(map[world.RegionID]world.RegionID),
	}
	if gs == nil || armyRef == nil || start == "" || gs.Regions[start] == nil {
		return routes
	}
	snapshot := newAIRouteSnapshot(gs, armyRef, mode, moveScore)
	routes.cost[start] = 0
	routes.hops[start] = 0
	queue := &aiRouteQueue{{regionID: start}}
	heap.Init(queue)

	for queue.Len() > 0 {
		current := heap.Pop(queue).(aiRouteQueueItem)
		if current.cost != routes.cost[current.regionID] || current.hops != routes.hops[current.regionID] || current.first != routes.firstStep[current.regionID] {
			continue
		}
		if maxHops > 0 && current.hops >= maxHops {
			continue
		}
		currentRegion := gs.Regions[current.regionID]
		if currentRegion == nil {
			continue
		}
		if current.regionID != start {
			_, transit, _ := snapshot.regionAccess(currentRegion)
			if !transit {
				continue
			}
		}

		for _, neighborID := range currentRegion.Neighbors {
			neighbor := gs.Regions[neighborID]
			entryCost, allowed := snapshot.entryCost(neighbor)
			if !allowed {
				continue
			}
			candidate := aiRouteQueueItem{
				regionID: neighborID,
				cost:     current.cost + entryCost,
				hops:     current.hops + 1,
				first:    current.first,
			}
			if current.regionID == start {
				candidate.first = neighborID
			}
			oldCost, seen := routes.cost[neighborID]
			if seen && !aiRouteLabelBetter(candidate, oldCost, routes.hops[neighborID], routes.firstStep[neighborID]) {
				continue
			}
			routes.cost[neighborID] = candidate.cost
			routes.hops[neighborID] = candidate.hops
			routes.firstStep[neighborID] = candidate.first
			heap.Push(queue, candidate)
		}
	}
	return routes
}

func aiRouteLabelBetter(candidate aiRouteQueueItem, oldCost, oldHops int, oldFirst world.RegionID) bool {
	if candidate.cost != oldCost {
		return candidate.cost < oldCost
	}
	if candidate.hops != oldHops {
		return candidate.hops < oldHops
	}
	return oldFirst == "" || candidate.first < oldFirst
}

func (ctx *StrategicContext) routesFor(armyRef *army.Army, start world.RegionID, mode aiRouteMode, maxHops int) *aiRouteMap {
	if ctx == nil || ctx.gs == nil || armyRef == nil || start == "" {
		return &aiRouteMap{}
	}
	if ctx.routeCache == nil {
		ctx.routeCache = make(map[aiRouteCacheKey]*aiRouteMap)
	}
	key := aiRouteCacheKey{armyID: armyRef.ID, start: start, mode: mode, maxHops: maxHops}
	if cached := ctx.routeCache[key]; cached != nil {
		return cached
	}
	routes := aiWeightedLandRoutes(ctx.gs, armyRef, start, mode, maxHops, nil)
	ctx.routeCache[key] = routes
	return routes
}

func (ctx *StrategicContext) routeDistance(armyRef *army.Army, target world.RegionID, mode aiRouteMode) int {
	if ctx == nil || armyRef == nil || target == "" {
		return -1
	}
	distance, reachable := ctx.routesFor(armyRef, armyRef.RegionID, mode, 0).distance(target)
	if !reachable {
		return -1
	}
	return distance
}

func (ctx *StrategicContext) routeNextStep(armyRef *army.Army, target world.RegionID, mode aiRouteMode) world.RegionID {
	if ctx == nil || armyRef == nil || target == "" || armyRef.RegionID == target {
		return ""
	}
	return ctx.routesFor(armyRef, armyRef.RegionID, mode, 0).nextStep(target)
}
