package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

type aiFishingSeaCandidate struct {
	regionID world.RegionID
	route    aiSeaRouteResult
}

func aiNavalFishingMissionMove(gs *state.GameState, fleet *army.Army, ctx *StrategicContext) (world.RegionID, bool) {
	if gs == nil || fleet == nil || !fleet.IsNaval || fleet.OwnerID == "" ||
		fleet.TradeRouteKey != "" || len(fleet.EmbarkedUnits) > 0 ||
		gs.NavalFishingGrainIncomeForFleet(fleet) <= 0 {
		return "", false
	}
	if fleet.NavalMission != nil && fleet.NavalMission.Kind != army.NavalMissionFishing {
		return "", false
	}

	fid := faction.FactionID(fleet.OwnerID)
	owner := gs.Factions[fid]
	if owner == nil || owner.IsEliminated || ctx == nil || ctx.FactionID != fid || ctx.navalMission != nil {
		return "", false
	}
	preview := gs.FactionGrainEconomyPreview(fid)
	needsFishing := preview.NetChange < 0 || owner.Grain < gs.GrainStorageCapacityForFaction(fid)

	if fleet.NavalMission != nil {
		if needsFishing && fleet.IsAtSea() && fleet.NavalMission.TargetRegionID == fleet.RegionID &&
			gs.IsValidNavalFishingTarget(fleet.OwnerID, fleet.RegionID) {
			if aiNavalFleetMeetsSafetyGate(gs, fleet, fleet.RegionID) {
				return "", true
			}
		}
		gs.ClearNavalMission(fleet.ID)
		if !needsFishing {
			return "", false
		}
	}

	if !needsFishing {
		return "", false
	}

	candidates := make([]aiFishingSeaCandidate, 0)
	for _, region := range aiSortedRegions(gs) {
		if region == nil || !region.IsSea || !gs.IsValidNavalFishingTarget(fleet.OwnerID, region.ID) {
			continue
		}
		route := aiThreatAwareSeaRoute(gs, fid, fleet.RegionID, region.ID)
		if !route.Reachable {
			continue
		}
		if route.FirstStep != "" && !aiNavalFleetMeetsSafetyGate(gs, fleet, route.FirstStep) {
			continue
		}
		if !aiNavalFleetMeetsSafetyGate(gs, fleet, region.ID) {
			continue
		}
		candidates = append(candidates, aiFishingSeaCandidate{regionID: region.ID, route: route})
	}
	sort.Slice(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		if left.route.MaxThreat != right.route.MaxThreat {
			return left.route.MaxThreat < right.route.MaxThreat
		}
		if left.route.TotalThreat != right.route.TotalThreat {
			return left.route.TotalThreat < right.route.TotalThreat
		}
		if left.route.Hops != right.route.Hops {
			return left.route.Hops < right.route.Hops
		}
		return left.regionID < right.regionID
	})
	if len(candidates) == 0 {
		return "", false
	}

	target := candidates[0]
	if fleet.IsDocked() && fleet.RegionID == target.regionID {
		return fleet.RegionID, true
	}
	if fleet.RegionID == target.regionID {
		mission := army.NavalMission{Kind: army.NavalMissionFishing, TargetRegionID: target.regionID}
		if assigned, _ := gs.AssignNavalMission(fleet.ID, mission); assigned {
			return "", true
		}
		return "", false
	}
	if target.route.FirstStep == "" {
		return "", false
	}
	return target.route.FirstStep, true
}
