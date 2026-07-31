package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	aiNavalPatrolEnemyWeight    = 1000
	aiNavalPatrolThreatWeight   = 700
	aiNavalPatrolTradeWeight    = 400
	aiNavalPatrolPortWeight     = 100
	aiNavalPatrolPressureWeight = 5
)

// aiNavalPatrolMove, aktif bir çıkarma görevi bulunmayan savaş gemisini
// ticaret hattı, tehdit altındaki liman veya düşman filosuna doğru yönlendirir.
// Hedefe girişteki gerçek savaş çözümü executeMove içinde kalır; bu fonksiyon
// yalnızca bir sonraki deniz adımını seçer.
func aiNavalPatrolMove(gs *state.GameState, fleet *army.Army, ctx *StrategicContext) (world.RegionID, bool) {
	if gs == nil || fleet == nil || !isWarshipFleet(fleet, gs.UnitTypes) {
		return "", false
	}

	weights := make(map[world.RegionID]int)
	addTarget := func(regionID world.RegionID, weight int) {
		region := gs.Regions[regionID]
		if region == nil || !region.IsSea || region.IsLocked || regionID == fleet.RegionID {
			return
		}
		if weight > weights[regionID] {
			weights[regionID] = weight
		}
	}

	hostileBySea, _ := aiNavalPowerMaps(gs, faction.FactionID(fleet.OwnerID))
	for seaID, power := range hostileBySea {
		addTarget(seaID, aiNavalPatrolEnemyWeight+power)
	}

	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.OwnerID != fleet.OwnerID || !region.HasPort() {
			continue
		}
		threat := maxPortApproachThreatPower(gs, faction.FactionID(fleet.OwnerID), region)
		weight := aiNavalPatrolPortWeight
		if threat > 0 {
			weight = aiNavalPatrolThreatWeight + threat
		}
		for _, neighborID := range aiSortedNeighborIDs(region) {
			if neighbor := gs.Regions[neighborID]; neighbor != nil && neighbor.IsSea {
				addTarget(neighborID, weight)
			}
		}
	}

	if ctx != nil {
		for _, portID := range ctx.ThreatenedPortIDs {
			port := gs.Regions[portID]
			for _, neighborID := range aiSortedNeighborIDs(port) {
				if neighbor := gs.Regions[neighborID]; neighbor != nil && neighbor.IsSea {
					addTarget(neighborID, aiNavalPatrolThreatWeight)
				}
			}
		}
	}

	for _, route := range gs.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 || route.FromFactionID != fleet.OwnerID && route.ToFactionID != fleet.OwnerID {
			continue
		}
		for _, seaID := range gs.MerchantTradeRouteSeaRegions(route) {
			addTarget(seaID, aiNavalPatrolTradeWeight)
		}
	}

	if len(weights) == 0 {
		return "", true
	}

	owner := faction.FactionID(fleet.OwnerID)
	type patrolCandidate struct {
		next   world.RegionID
		target world.RegionID
		score  int
		route  aiSeaRouteResult
	}
	var best *patrolCandidate
	ids := make([]world.RegionID, 0, len(weights))
	for regionID := range weights {
		ids = append(ids, regionID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, targetID := range ids {
		var route aiSeaRouteResult
		if ctx != nil && ctx.FactionID == owner {
			route = ctx.navalSeaRoute(fleet.RegionID, targetID)
		} else {
			route = aiThreatAwareSeaRoute(gs, owner, fleet.RegionID, targetID)
		}
		if !route.Reachable || route.FirstStep == "" || !aiNavalFleetMeetsSafetyGate(gs, fleet, route.FirstStep) {
			continue
		}
		score := weights[targetID]*100 + aiSeaPressure(gs, fleet.OwnerID, targetID)*aiNavalPatrolPressureWeight - route.Hops
		candidate := patrolCandidate{next: route.FirstStep, target: targetID, score: score, route: route}
		if best == nil || candidate.score > best.score || candidate.score == best.score && (candidate.target < best.target || candidate.target == best.target && candidate.next < best.next) {
			copy := candidate
			best = &copy
		}
	}
	if best == nil {
		return "", true
	}
	return best.next, true
}

func maxPortApproachThreatPower(gs *state.GameState, owner faction.FactionID, region *world.Region) int {
	if gs == nil || region == nil {
		return 0
	}
	maxThreat := 0
	for _, neighborID := range aiSortedNeighborIDs(region) {
		neighbor := gs.Regions[neighborID]
		if neighbor == nil || !neighbor.IsSea {
			continue
		}
		if threat := aiPortApproachThreatPower(gs, owner, neighborID); threat > maxThreat {
			maxThreat = threat
		}
	}
	return maxThreat
}
