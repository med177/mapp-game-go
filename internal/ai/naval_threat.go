package ai

import (
	"container/heap"
	"math"
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const aiNavalMissionSafetyPercent = 110

// AINavalThreat bir AI turunda türetilen, save'e yazılmayan deniz güç özetidir.
// HostilePower ilgili denizde savaş halindeki filoların efektif savunma gücüdür.
type AINavalThreat struct {
	SeaRegionID   world.RegionID
	HostilePower  int
	FriendlyPower int
}

type aiSeaRouteResult struct {
	Reachable   bool
	FirstStep   world.RegionID
	Hops        int
	MaxThreat   int
	TotalThreat int
}

type aiSeaRouteLabel struct {
	regionID    world.RegionID
	firstStep   world.RegionID
	hops        int
	maxThreat   int
	totalThreat int
}

type aiSeaRouteQueue []aiSeaRouteLabel

func (queue aiSeaRouteQueue) Len() int { return len(queue) }
func (queue aiSeaRouteQueue) Less(i, j int) bool {
	return aiSeaRouteLabelBetter(queue[i], queue[j])
}
func (queue aiSeaRouteQueue) Swap(i, j int)   { queue[i], queue[j] = queue[j], queue[i] }
func (queue *aiSeaRouteQueue) Push(value any) { *queue = append(*queue, value.(aiSeaRouteLabel)) }
func (queue *aiSeaRouteQueue) Pop() any {
	old := *queue
	last := len(old) - 1
	item := old[last]
	*queue = old[:last]
	return item
}

func buildAINavalThreatSnapshot(ctx *StrategicContext) {
	if ctx == nil || ctx.gs == nil || ctx.FactionID == "" {
		return
	}
	ctx.NavalThreats = ctx.NavalThreats[:0]
	ctx.ThreatenedPortIDs = ctx.ThreatenedPortIDs[:0]
	hostileBySea, friendlyBySea := aiNavalPowerMaps(ctx.gs, ctx.FactionID)
	ctx.navalThreatPower = hostileBySea
	for _, region := range aiSortedRegions(ctx.gs) {
		if region == nil || !region.IsSea {
			continue
		}
		hostile := hostileBySea[region.ID]
		friendly := friendlyBySea[region.ID]
		if hostile > 0 {
			ctx.NavalThreats = append(ctx.NavalThreats, AINavalThreat{
				SeaRegionID: region.ID, HostilePower: hostile, FriendlyPower: friendly,
			})
		}
	}

	seenPorts := make(map[world.RegionID]struct{})
	for _, region := range aiSortedRegions(ctx.gs) {
		if region == nil || region.IsSea || region.OwnerID != string(ctx.FactionID) || !aiRegionHasPortBuilding(region) {
			continue
		}
		for _, neighborID := range region.Neighbors {
			sea := ctx.gs.Regions[neighborID]
			if sea == nil || !sea.IsSea || aiPortApproachThreatPowerFromMap(ctx.gs, sea.ID, hostileBySea) <= 0 {
				continue
			}
			seenPorts[region.ID] = struct{}{}
			break
		}
	}
	for regionID := range seenPorts {
		ctx.ThreatenedPortIDs = append(ctx.ThreatenedPortIDs, regionID)
	}
	sort.Slice(ctx.ThreatenedPortIDs, func(i, j int) bool { return ctx.ThreatenedPortIDs[i] < ctx.ThreatenedPortIDs[j] })
}

func aiEffectiveNavalPower(gs *state.GameState, fleet *army.Army, attacking bool) int {
	if gs == nil || fleet == nil || !fleet.IsNaval || len(fleet.Units) == 0 {
		return 0
	}
	mods := aiTechMods(gs, fleet.OwnerID)
	technologyMod := mods.NavalDefenseMod
	commanderAttack, commanderDefense := fleet.CommanderModifiers()
	commanderMod := commanderDefense
	if attacking {
		technologyMod = mods.NavalAttackMod
		commanderMod = commanderAttack
	}
	multiplier := 1.0 + technologyMod + commanderMod + fleet.CommanderMoraleModifier()
	if multiplier < 0.1 {
		multiplier = 0.1
	}
	return maxInt(1, int(math.Round(float64(fleet.TotalStrength(gs.UnitTypes))*multiplier)))
}

func aiHostileNavalPowerAtSea(gs *state.GameState, owner faction.FactionID, seaRegionID world.RegionID) int {
	if gs == nil || owner == "" || seaRegionID == "" {
		return 0
	}
	power := 0
	for _, fleet := range aiSortedArmies(gs) {
		if fleet == nil || !fleet.IsAtSea() || fleet.RegionID != seaRegionID || fleet.OwnerID == string(owner) {
			continue
		}
		relation := diplomacy.Relation(gs, owner, faction.FactionID(fleet.OwnerID))
		if relation == nil || relation.Stance != faction.StanceWar {
			continue
		}
		power += aiEffectiveNavalPower(gs, fleet, false)
	}
	return power
}

func aiNavalPowerMaps(gs *state.GameState, owner faction.FactionID) (map[world.RegionID]int, map[world.RegionID]int) {
	hostile := make(map[world.RegionID]int)
	friendly := make(map[world.RegionID]int)
	if gs == nil || owner == "" {
		return hostile, friendly
	}
	for _, fleet := range aiSortedArmies(gs) {
		if fleet == nil || !fleet.IsAtSea() {
			continue
		}
		if fleet.OwnerID == string(owner) {
			friendly[fleet.RegionID] += aiEffectiveNavalPower(gs, fleet, true)
			continue
		}
		relation := diplomacy.Relation(gs, owner, faction.FactionID(fleet.OwnerID))
		if relation != nil && relation.Stance == faction.StanceWar {
			hostile[fleet.RegionID] += aiEffectiveNavalPower(gs, fleet, false)
		}
	}
	return hostile, friendly
}

func aiPortApproachThreatPower(gs *state.GameState, owner faction.FactionID, portSea world.RegionID) int {
	if gs == nil || portSea == "" {
		return 0
	}
	hostileBySea, _ := aiNavalPowerMaps(gs, owner)
	return aiPortApproachThreatPowerFromMap(gs, portSea, hostileBySea)
}

func aiPortApproachThreatPowerFromMap(gs *state.GameState, portSea world.RegionID, hostileBySea map[world.RegionID]int) int {
	if gs == nil || portSea == "" {
		return 0
	}
	threat := hostileBySea[portSea]
	sea := gs.Regions[portSea]
	for _, neighborID := range aiSortedNeighborIDs(sea) {
		neighbor := gs.Regions[neighborID]
		if neighbor == nil || !neighbor.IsSea {
			continue
		}
		threat += hostileBySea[neighborID]
	}
	return threat
}

func aiThreatAwareSeaRoute(gs *state.GameState, owner faction.FactionID, start, target world.RegionID) aiSeaRouteResult {
	hostileBySea, _ := aiNavalPowerMaps(gs, owner)
	return aiThreatAwareSeaRouteWithThreats(gs, start, target, hostileBySea)
}

func aiThreatAwareSeaRouteWithThreats(gs *state.GameState, start, target world.RegionID, hostileBySea map[world.RegionID]int) aiSeaRouteResult {
	if gs == nil || start == "" || target == "" {
		return aiSeaRouteResult{}
	}
	startRegion := gs.Regions[start]
	targetRegion := gs.Regions[target]
	if startRegion == nil || targetRegion == nil || !startRegion.IsSea || !targetRegion.IsSea {
		return aiSeaRouteResult{}
	}
	if start == target {
		return aiSeaRouteResult{Reachable: true}
	}

	best := map[world.RegionID]aiSeaRouteLabel{start: {regionID: start}}
	queue := &aiSeaRouteQueue{{regionID: start}}
	heap.Init(queue)
	for queue.Len() > 0 {
		current := heap.Pop(queue).(aiSeaRouteLabel)
		known, exists := best[current.regionID]
		if !exists || !aiSeaRouteLabelEqual(current, known) {
			continue
		}
		if current.regionID == target {
			return aiSeaRouteResult{
				Reachable: true, FirstStep: current.firstStep, Hops: current.hops,
				MaxThreat: current.maxThreat, TotalThreat: current.totalThreat,
			}
		}
		for _, neighborID := range aiSortedNeighborIDs(gs.Regions[current.regionID]) {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil || !neighbor.IsSea {
				continue
			}
			threat := hostileBySea[neighborID]
			candidate := aiSeaRouteLabel{
				regionID: neighborID, firstStep: current.firstStep, hops: current.hops + 1,
				maxThreat: maxInt(current.maxThreat, threat), totalThreat: current.totalThreat + threat,
			}
			if current.regionID == start {
				candidate.firstStep = neighborID
			}
			old, seen := best[neighborID]
			if seen && !aiSeaRouteLabelBetter(candidate, old) {
				continue
			}
			best[neighborID] = candidate
			heap.Push(queue, candidate)
		}
	}
	return aiSeaRouteResult{}
}

func (ctx *StrategicContext) navalSeaRoute(start, target world.RegionID) aiSeaRouteResult {
	if ctx == nil || ctx.gs == nil {
		return aiSeaRouteResult{}
	}
	return aiThreatAwareSeaRouteWithThreats(ctx.gs, start, target, ctx.navalThreatPower)
}

func aiSeaRouteLabelBetter(candidate, old aiSeaRouteLabel) bool {
	if candidate.maxThreat != old.maxThreat {
		return candidate.maxThreat < old.maxThreat
	}
	if candidate.totalThreat != old.totalThreat {
		return candidate.totalThreat < old.totalThreat
	}
	if candidate.hops != old.hops {
		return candidate.hops < old.hops
	}
	if candidate.firstStep != old.firstStep {
		return candidate.firstStep < old.firstStep
	}
	return candidate.regionID < old.regionID
}

func aiSeaRouteLabelEqual(left, right aiSeaRouteLabel) bool {
	return left.regionID == right.regionID && left.firstStep == right.firstStep && left.hops == right.hops && left.maxThreat == right.maxThreat && left.totalThreat == right.totalThreat
}

func aiNavalFleetMeetsSafetyGate(gs *state.GameState, fleet *army.Army, target world.RegionID) bool {
	if gs == nil || fleet == nil || target == "" {
		return false
	}
	threat := aiHostileNavalPowerAtSea(gs, faction.FactionID(fleet.OwnerID), target)
	if threat <= 0 {
		return true
	}
	return aiEffectiveNavalPower(gs, fleet, true)*100 >= threat*aiNavalMissionSafetyPercent
}

func aiThreatAwareSeaNextStep(gs *state.GameState, owner faction.FactionID, start, target world.RegionID) world.RegionID {
	route := aiThreatAwareSeaRoute(gs, owner, start, target)
	if !route.Reachable {
		return ""
	}
	return route.FirstStep
}
