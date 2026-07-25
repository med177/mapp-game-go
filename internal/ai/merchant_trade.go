package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const merchantShipTypeID = "merchant_ship"

type aiMerchantRoute struct {
	route *economy.TradeRoute
	key   string
	seas  []world.RegionID
}

// aiExecuteMerchantTradeStrategy 1300 senaryosunda ticaret cumhuriyetlerinin
// merchant filolarını aktif deniz rotalarına bağlar ve eksik kapasiteyi üretir.
func aiExecuteMerchantTradeStrategy(gs *state.GameState, fid faction.FactionID, budget *aiBudget, _ *StrategicContext, steps *[]TurnStep) {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || !aiMerchantTradeFaction(fid) {
		return
	}
	routes := aiEligibleMerchantRoutes(gs, fid)
	aiAssignMerchantTradeFleets(gs, fid, routes)
	if len(routes) == 0 {
		return
	}
	if aiProduceTradeEscortIfNeeded(gs, fid, routes, budget, steps) {
		return
	}
	aiProduceMerchantShipIfNeeded(gs, fid, routes, budget, steps)
}

func aiMerchantTradeFaction(fid faction.FactionID) bool {
	return fid == "venice" || fid == "genoa"
}

func aiMerchantTradeResourceReserve(gs *state.GameState, fid faction.FactionID) economy.ResourceCost {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || !aiMerchantTradeFaction(fid) {
		return economy.ResourceCost{}
	}
	routes := aiEligibleMerchantRoutes(gs, fid)
	if len(routes) == 0 {
		return economy.ResourceCost{}
	}
	desired := len(routes) * economy.MaxMerchantAmountBonusPerRoute
	existing := 0
	for _, fleet := range aiSortedArmies(gs) {
		if fleet.OwnerID == string(fid) && fleet.IsNaval {
			existing += aiMerchantShipCount(gs, fleet)
		}
	}
	for _, order := range gs.ProductionQueue {
		if order.Kind == aiProductionKindUnit && order.FactionID == string(fid) && order.TypeID == merchantShipTypeID {
			existing++
		}
	}
	if existing >= desired {
		return economy.ResourceCost{}
	}
	merchantType := gs.UnitTypes[merchantShipTypeID]
	portType := gs.BuildingTypes["port"]
	if merchantType == nil || portType == nil {
		return economy.ResourceCost{}
	}
	target := aiLeastCoveredMerchantRoute(gs, routes)
	port, _ := aiOwnedTradeCenterPort(gs, fid, target.seas)
	if port == nil {
		return economy.ResourceCost{}
	}
	if aiBuildingLevel(port, "port")+aiQueuedBuildingCount(gs, port.ID, "port", fid) < maxInt(1, merchantType.RequiredBldgLevel) {
		return economy.ResourceCost{
			Gold: portType.GoldCost + merchantType.GoldCost, Grain: portType.GrainCost + merchantType.GrainCost,
			Iron: portType.IronCost + merchantType.IronCost, Timber: portType.TimberCost + merchantType.TimberCost,
			Stone: portType.StoneCost + merchantType.StoneCost,
		}
	}
	return economy.ResourceCost{
		Gold: merchantType.GoldCost, Grain: merchantType.GrainCost, Iron: merchantType.IronCost,
		Timber: merchantType.TimberCost, Stone: merchantType.StoneCost,
		Spice: merchantType.SpiceCost, Cloth: merchantType.ClothCost,
	}
}

func aiEligibleMerchantRoutes(gs *state.GameState, fid faction.FactionID) []aiMerchantRoute {
	if gs == nil || fid == "" {
		return nil
	}
	fidString := string(fid)
	result := make([]aiMerchantRoute, 0)
	for _, route := range gs.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 || route.FromFactionID != fidString && route.ToFactionID != fidString {
			continue
		}
		seas := gs.MerchantTradeRouteSeaRegions(route)
		if len(seas) == 0 {
			continue
		}
		result = append(result, aiMerchantRoute{route: route, key: route.AssignmentKey(), seas: seas})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].route.GoldPerUnit != result[j].route.GoldPerUnit {
			return result[i].route.GoldPerUnit > result[j].route.GoldPerUnit
		}
		return result[i].key < result[j].key
	})
	return result
}

func aiAssignMerchantTradeFleets(gs *state.GameState, fid faction.FactionID, routes []aiMerchantRoute) {
	if gs == nil || fid == "" {
		return
	}
	byKey := make(map[string]aiMerchantRoute, len(routes))
	coverage := make(map[string]int, len(routes))
	for _, candidate := range routes {
		byKey[candidate.key] = candidate
	}

	var unassigned []*army.Army
	for _, fleet := range aiSortedArmies(gs) {
		if fleet == nil || fleet.OwnerID != string(fid) || !fleet.IsNaval {
			continue
		}
		merchantCount := aiMerchantShipCount(gs, fleet)
		if merchantCount == 0 {
			if fleet.TradeRouteKey != "" {
				fleet.TradeRouteKey = ""
			}
			continue
		}
		if _, valid := byKey[fleet.TradeRouteKey]; !valid {
			fleet.TradeRouteKey = ""
			unassigned = append(unassigned, fleet)
			continue
		}
		coverage[fleet.TradeRouteKey] = minInt(economy.MaxMerchantAmountBonusPerRoute, coverage[fleet.TradeRouteKey]+merchantCount)
	}

	for _, fleet := range unassigned {
		bestKey := ""
		bestCoverage := economy.MaxMerchantAmountBonusPerRoute + 1
		for _, candidate := range routes {
			current := coverage[candidate.key]
			if current >= economy.MaxMerchantAmountBonusPerRoute {
				continue
			}
			if current < bestCoverage {
				bestKey = candidate.key
				bestCoverage = current
			}
		}
		if bestKey == "" {
			continue
		}
		fleet.TradeRouteKey = bestKey
		coverage[bestKey] = minInt(economy.MaxMerchantAmountBonusPerRoute, coverage[bestKey]+aiMerchantShipCount(gs, fleet))
	}
}

func aiMerchantTradeFleetMove(gs *state.GameState, fleet *army.Army, ctx *StrategicContext) (world.RegionID, bool) {
	if gs == nil || fleet == nil || !fleet.IsNaval || aiMerchantShipCount(gs, fleet) == 0 {
		return "", false
	}
	if fleet.TradeRouteKey == "" {
		return "", true
	}
	var assigned *economy.TradeRoute
	for _, route := range gs.TradeRoutes {
		if route != nil && route.SuspendedTurns <= 0 && route.AssignmentKey() == fleet.TradeRouteKey {
			assigned = route
			break
		}
	}
	if assigned == nil || fleet.OwnerID != assigned.FromFactionID && fleet.OwnerID != assigned.ToFactionID {
		return "", true
	}
	seas := gs.MerchantTradeRouteSeaRegions(assigned)
	for _, seaID := range seas {
		if seaID == fleet.RegionID {
			return "", true
		}
	}

	best := aiSeaRouteResult{}
	bestTarget := world.RegionID("")
	for _, target := range seas {
		var candidate aiSeaRouteResult
		if ctx != nil && ctx.FactionID == faction.FactionID(fleet.OwnerID) {
			candidate = ctx.navalSeaRoute(fleet.RegionID, target)
		} else {
			candidate = aiThreatAwareSeaRoute(gs, faction.FactionID(fleet.OwnerID), fleet.RegionID, target)
		}
		if !candidate.Reachable {
			continue
		}
		if bestTarget == "" || aiMerchantSeaRouteBetter(candidate, target, best, bestTarget) {
			best = candidate
			bestTarget = target
		}
	}
	if bestTarget == "" || best.FirstStep == "" || !aiNavalFleetMeetsSafetyGate(gs, fleet, best.FirstStep) {
		return "", true
	}
	return best.FirstStep, true
}

func aiMerchantSeaRouteBetter(candidate aiSeaRouteResult, candidateTarget world.RegionID, current aiSeaRouteResult, currentTarget world.RegionID) bool {
	if candidate.MaxThreat != current.MaxThreat {
		return candidate.MaxThreat < current.MaxThreat
	}
	if candidate.TotalThreat != current.TotalThreat {
		return candidate.TotalThreat < current.TotalThreat
	}
	if candidate.Hops != current.Hops {
		return candidate.Hops < current.Hops
	}
	return candidateTarget < currentTarget
}

func aiProduceMerchantShipIfNeeded(gs *state.GameState, fid faction.FactionID, routes []aiMerchantRoute, budget *aiBudget, steps *[]TurnStep) bool {
	self := gs.Factions[fid]
	merchantType := gs.UnitTypes[merchantShipTypeID]
	portType := gs.BuildingTypes["port"]
	if self == nil || merchantType == nil || portType == nil || !merchantType.HasAllRequiredTechs(self.Research.Completed) {
		return false
	}
	desired := len(routes) * economy.MaxMerchantAmountBonusPerRoute
	existing := 0
	for _, fleet := range aiSortedArmies(gs) {
		if fleet.OwnerID == string(fid) && fleet.IsNaval {
			existing += aiMerchantShipCount(gs, fleet)
		}
	}
	for _, order := range gs.ProductionQueue {
		if order.Kind == aiProductionKindUnit && order.FactionID == string(fid) && order.TypeID == merchantShipTypeID {
			existing++
		}
	}
	if existing >= desired {
		return false
	}

	target := aiLeastCoveredMerchantRoute(gs, routes)
	port, sea := aiOwnedTradeCenterPort(gs, fid, target.seas)
	if port == nil || sea == "" {
		return false
	}
	requiredPortLevel := maxInt(1, merchantType.RequiredBldgLevel)
	currentPortLevel := aiBuildingLevel(port, "port")
	queuedPortLevels := aiQueuedBuildingCount(gs, port.ID, "port", fid)
	if currentPortLevel < requiredPortLevel {
		if queuedPortLevels > 0 || currentPortLevel+queuedPortLevels >= portType.MaxPerRegion || !aiBuildingAllowed(gs, port, "port", portType.RequiredTerrain) {
			return false
		}
		cost := economy.ResourceCost{Gold: portType.GoldCost, Grain: portType.GrainCost, Iron: portType.IronCost, Timber: portType.TimberCost, Stone: portType.StoneCost, Spice: portType.SpiceCost, Cloth: portType.ClothCost}
		if !aiCanAffordForBudget(self, cost, budget, aiBudgetNaval) || !aiApplyBudgetedCost(self, cost, budget, aiBudgetNaval) {
			return false
		}
		turns := aiBuildingTurnsRequired(port, "port", portType.TurnsRequired, queuedPortLevels)
		aiEnqueueProduction(gs, fid, aiProductionKindBuilding, port.ID, "port", turns)
		addTurnStep(steps, TurnStep{
			FactionID: fid, Kind: TurnStepBuild, TargetRegion: port.ID, FocusRegion: sea,
			Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, port.ID) + " ticaret merkezinin merchant kapasitesini büyütüyor.",
		})
		return true
	}
	if aiPendingUnitCountByRegion(gs, port.ID, fid) >= aiMaxRegionQueue || aiLaneRemainingCapacity(gs, port.ID, fid, merchantType) <= 0 {
		return false
	}
	cost := economy.ResourceCost{Gold: merchantType.GoldCost, Grain: merchantType.GrainCost, Iron: merchantType.IronCost, Timber: merchantType.TimberCost, Stone: merchantType.StoneCost, Spice: merchantType.SpiceCost, Cloth: merchantType.ClothCost}
	if !aiCanAffordForBudget(self, cost, budget, aiBudgetNaval) || !aiApplyBudgetedCost(self, cost, budget, aiBudgetNaval) {
		return false
	}
	aiEnqueueProduction(gs, fid, aiProductionKindUnit, port.ID, merchantShipTypeID, merchantType.TurnsRequired)
	addTurnStep(steps, TurnStep{
		FactionID: fid, Kind: TurnStepRecruit, TargetRegion: port.ID, FocusRegion: sea,
		Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, port.ID) + " limanında aktif ticaret rotası için merchant gemisi hazırlıyor.",
	})
	return true
}

func aiProduceTradeEscortIfNeeded(gs *state.GameState, fid faction.FactionID, routes []aiMerchantRoute, budget *aiBudget, steps *[]TurnStep) bool {
	self := gs.Factions[fid]
	warshipType := gs.UnitTypes["warship"]
	portType := gs.BuildingTypes["port"]
	if self == nil || warshipType == nil || portType == nil || !warshipType.HasAllRequiredTechs(self.Research.Completed) {
		return false
	}
	var threatenedSea world.RegionID
	var threatenedPort *world.Region
	threatPower := 0
	for _, candidate := range routes {
		port, sea := aiOwnedTradeCenterPort(gs, fid, candidate.seas)
		if port == nil || sea == "" {
			continue
		}
		threat := aiPortApproachThreatPower(gs, fid, sea)
		if threat > threatPower || threat == threatPower && threat > 0 && sea < threatenedSea {
			threatPower = threat
			threatenedSea = sea
			threatenedPort = port
		}
	}
	if threatPower <= 0 || threatenedPort == nil {
		return false
	}
	requiredPower := (threatPower*aiNavalMissionSafetyPercent + 99) / 100
	projectedPower := aiProjectedTradeCenterPower(gs, fid, threatenedPort.ID, threatenedSea)
	if projectedPower >= requiredPower {
		return false
	}

	requiredPortLevel := maxInt(1, warshipType.RequiredBldgLevel)
	currentPortLevel := aiBuildingLevel(threatenedPort, "port")
	queuedPortLevels := aiQueuedBuildingCount(gs, threatenedPort.ID, "port", fid)
	if currentPortLevel < requiredPortLevel {
		if queuedPortLevels > 0 || currentPortLevel+queuedPortLevels >= portType.MaxPerRegion || !aiBuildingAllowed(gs, threatenedPort, "port", portType.RequiredTerrain) {
			return false
		}
		cost := economy.ResourceCost{Gold: portType.GoldCost, Grain: portType.GrainCost, Iron: portType.IronCost, Timber: portType.TimberCost, Stone: portType.StoneCost, Spice: portType.SpiceCost, Cloth: portType.ClothCost}
		if !aiCanAffordForBudget(self, cost, budget, aiBudgetNaval) || !aiApplyBudgetedCost(self, cost, budget, aiBudgetNaval) {
			return false
		}
		turns := aiBuildingTurnsRequired(threatenedPort, "port", portType.TurnsRequired, queuedPortLevels)
		aiEnqueueProduction(gs, fid, aiProductionKindBuilding, threatenedPort.ID, "port", turns)
		addTurnStep(steps, TurnStep{
			FactionID: fid, Kind: TurnStepBuild, TargetRegion: threatenedPort.ID, FocusRegion: threatenedSea,
			Message: turnFactionName(gs, fid) + " tehdit altındaki " + turnRegionName(gs, threatenedPort.ID) + " ticaret limanını escort üretimi için güçlendiriyor.",
		})
		return true
	}

	unitPower := aiEffectiveNavalPower(gs, &army.Army{OwnerID: string(fid), IsNaval: true, Units: []army.Unit{{TypeID: warshipType.ID, CurrentHP: army.MaxUnitHP}}}, true)
	if unitPower <= 0 {
		return false
	}
	cost := economy.ResourceCost{Gold: warshipType.GoldCost, Grain: warshipType.GrainCost, Iron: warshipType.IronCost, Timber: warshipType.TimberCost, Stone: warshipType.StoneCost, Spice: warshipType.SpiceCost, Cloth: warshipType.ClothCost}
	queued := false
	for projectedPower < requiredPower {
		if aiPendingUnitCountByRegion(gs, threatenedPort.ID, fid) >= aiMaxRegionQueue || aiLaneRemainingCapacity(gs, threatenedPort.ID, fid, warshipType) <= 0 {
			break
		}
		if !aiCanAffordForBudget(self, cost, budget, aiBudgetNaval) || !aiApplyBudgetedCost(self, cost, budget, aiBudgetNaval) {
			break
		}
		aiEnqueueProduction(gs, fid, aiProductionKindUnit, threatenedPort.ID, warshipType.ID, warshipType.TurnsRequired)
		addTurnStep(steps, TurnStep{
			FactionID: fid, Kind: TurnStepRecruit, TargetRegion: threatenedPort.ID, FocusRegion: threatenedSea,
			Message: turnFactionName(gs, fid) + " ticaret merkezini yüzde 110 güvenlik eşiğine taşımak için escort savaş gemisi hazırlıyor.",
		})
		projectedPower += unitPower
		queued = true
	}
	return queued
}

func aiLeastCoveredMerchantRoute(gs *state.GameState, routes []aiMerchantRoute) aiMerchantRoute {
	coverage := make(map[string]int, len(routes))
	for _, fleet := range aiSortedArmies(gs) {
		if fleet.TradeRouteKey != "" {
			coverage[fleet.TradeRouteKey] += aiMerchantShipCount(gs, fleet)
		}
	}
	best := routes[0]
	for _, candidate := range routes[1:] {
		if coverage[candidate.key] < coverage[best.key] {
			best = candidate
		}
	}
	return best
}

func aiOwnedTradeCenterPort(gs *state.GameState, fid faction.FactionID, allowedSeas []world.RegionID) (*world.Region, world.RegionID) {
	if gs == nil || fid == "" || len(allowedSeas) == 0 {
		return nil, ""
	}
	allowed := make(map[world.RegionID]bool, len(allowedSeas))
	for _, seaID := range allowedSeas {
		allowed[seaID] = true
	}
	var bestPort *world.Region
	var bestSea world.RegionID
	for _, def := range gs.TradeCenters.Centers {
		if !def.ActiveInYear(gs.Year) {
			continue
		}
		port := gs.Regions[def.ID]
		if port == nil || port.OwnerID != string(fid) || port.IsSea {
			continue
		}
		for _, neighborID := range aiSortedNeighborIDs(port) {
			if !allowed[neighborID] || gs.Regions[neighborID] == nil || !gs.Regions[neighborID].IsSea {
				continue
			}
			if bestPort == nil || aiBuildingLevel(port, "port") > aiBuildingLevel(bestPort, "port") || aiBuildingLevel(port, "port") == aiBuildingLevel(bestPort, "port") && port.ID < bestPort.ID {
				bestPort = port
				bestSea = neighborID
			}
		}
	}
	return bestPort, bestSea
}

func aiProjectedTradeCenterPower(gs *state.GameState, fid faction.FactionID, portID, seaID world.RegionID) int {
	power := 0
	for _, fleet := range aiSortedArmies(gs) {
		if fleet.OwnerID == string(fid) && fleet.IsAtSea() && fleet.RegionID == seaID {
			power += aiEffectiveNavalPower(gs, fleet, true)
		}
	}
	for _, order := range gs.ProductionQueue {
		if order.Kind != aiProductionKindUnit || order.FactionID != string(fid) || order.RegionID != portID {
			continue
		}
		unitType := gs.UnitTypes[order.TypeID]
		if unitType == nil || unitType.Category != army.CategoryNavalWar {
			continue
		}
		power += aiEffectiveNavalPower(gs, &army.Army{OwnerID: string(fid), IsNaval: true, Units: []army.Unit{{TypeID: unitType.ID, CurrentHP: army.MaxUnitHP}}}, true)
	}
	return power
}

func aiMerchantShipCount(gs *state.GameState, fleet *army.Army) int {
	if gs == nil || fleet == nil {
		return 0
	}
	count := 0
	for _, unit := range fleet.Units {
		unitType := gs.UnitTypes[unit.TypeID]
		if unit.TypeID == merchantShipTypeID || unitType != nil && unitType.Category == army.CategoryNavalTrade {
			count++
		}
	}
	return count
}
