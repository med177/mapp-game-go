package ai

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// aiProduceNavalReserve closes the peacetime fleet gap derived from owned
// coast. It deliberately counts only warships: merchant and transport fleets
// retain their separate economic/operation policies and cannot satisfy naval
// deterrence by themselves.
func aiProduceNavalReserve(gs *state.GameState, fid faction.FactionID, budget *aiBudget, ctx *StrategicContext, steps *[]TurnStep) {
	if gs == nil || fid == "" || aiWarshipReserveShortfall(gs, fid, ctx) <= 0 || gs.UnitTypes == nil || gs.BuildingTypes == nil {
		return
	}
	self := gs.Factions[fid]
	warshipType := gs.UnitTypes["warship"]
	portType := gs.BuildingTypes["port"]
	if self == nil || self.IsEliminated || warshipType == nil || portType == nil {
		return
	}

	if !warshipType.HasAllRequiredTechs(self.Research.Completed) {
		return // research strategy receives the same reserve shortfall signal.
	}

	requiredPortLevel := maxInt(1, warshipType.RequiredBldgLevel)
	if productionRegion, seaRegion := aiFindWarshipReserveProductionPort(gs, fid, warshipType); productionRegion != nil {
		for aiWarshipReserveShortfall(gs, fid, ctx) > 0 {
			if aiPendingUnitCountByRegion(gs, productionRegion.ID, fid) >= aiMaxRegionQueue || aiLaneRemainingCapacity(gs, productionRegion.ID, fid, warshipType) <= 0 || !aiApplyUnitCostForBudget(self, warshipType, budget, aiBudgetNaval) {
				break
			}
			aiEnqueueProduction(gs, fid, aiProductionKindUnit, productionRegion.ID, warshipType.ID, warshipType.TurnsRequired)
			addTurnStep(steps, TurnStep{
				FactionID: fid, Kind: TurnStepRecruit, TargetRegion: productionRegion.ID, FocusRegion: seaRegion,
				Message: turnFactionName(gs, fid) + " kıyı rezervini tamamlamak için " + turnRegionName(gs, productionRegion.ID) + " limanında savaş gemisi hazırlıyor.",
			})
		}
		return
	}

	// Üretim limanı yoksa ilk uygulanabilir kıyıda gerekli liman seviyesini
	// kurar. Aynı turda yalnız bir seviye açılır; kuyruktaki seviye tamamlanana
	// kadar tekrar maliyet yazılmaz.
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.OwnerID != string(fid) || !region.IsCoastal(gs.Regions) || !aiBuildingAllowed(gs, region, "port", portType.RequiredTerrain) {
			continue
		}
		current := aiBuildingLevel(region, "port")
		queued := aiQueuedBuildingCount(gs, region.ID, "port", fid)
		if current+queued >= requiredPortLevel || current+queued >= portType.MaxPerRegion {
			continue
		}
		cost := aiBuildingResourceCost(portType)
		if !aiApplyBudgetedCost(self, cost, budget, aiBudgetNaval) {
			return
		}
		turns := aiBuildingTurnsRequired(region, "port", portType.TurnsRequired, queued)
		aiEnqueueProduction(gs, fid, aiProductionKindBuilding, region.ID, "port", turns)
		addTurnStep(steps, TurnStep{
			FactionID: fid, Kind: TurnStepBuild, TargetRegion: region.ID,
			Message: turnFactionName(gs, fid) + " kıyı savaş gemisi rezervi için " + turnRegionName(gs, region.ID) + " limanını geliştiriyor.",
		})
		return
	}
}

func aiFindWarshipReserveProductionPort(gs *state.GameState, fid faction.FactionID, warshipType *army.UnitType) (*world.Region, world.RegionID) {
	if gs == nil || warshipType == nil {
		return nil, ""
	}
	requiredPortLevel := maxInt(1, warshipType.RequiredBldgLevel)
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.OwnerID != string(fid) || !region.IsCoastal(gs.Regions) || aiBuildingLevel(region, "port") < requiredPortLevel || aiPendingUnitCountByRegion(gs, region.ID, fid) >= aiMaxRegionQueue || aiLaneRemainingCapacity(gs, region.ID, fid, warshipType) <= 0 {
			continue
		}
		for _, neighborID := range aiSortedNeighborIDs(region) {
			if sea := gs.Regions[neighborID]; sea != nil && sea.IsSea {
				return region, neighborID
			}
		}
	}
	return nil, ""
}

// aiNavalReserveProcurementCost exposes the immediate next legal prerequisite
// to the common trade-network procurement chain: first a port level, then a
// warship recipe. It never fabricates resources without a real production path.
func aiNavalReserveProcurementCost(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext) economy.ResourceCost {
	if gs == nil || aiWarshipReserveShortfall(gs, fid, ctx) <= 0 || gs.UnitTypes == nil || gs.BuildingTypes == nil {
		return economy.ResourceCost{}
	}
	warshipType := gs.UnitTypes["warship"]
	portType := gs.BuildingTypes["port"]
	self := gs.Factions[fid]
	if warshipType == nil || portType == nil || self == nil {
		return economy.ResourceCost{}
	}
	if !warshipType.HasAllRequiredTechs(self.Research.Completed) {
		return economy.ResourceCost{}
	}
	if region, _ := aiFindWarshipReserveProductionPort(gs, fid, warshipType); region != nil {
		return aiUnitResourceCost(warshipType)
	}
	requiredPortLevel := maxInt(1, warshipType.RequiredBldgLevel)
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.OwnerID != string(fid) || !region.IsCoastal(gs.Regions) {
			continue
		}
		if aiBuildingLevel(region, "port")+aiQueuedBuildingCount(gs, region.ID, "port", fid) < requiredPortLevel {
			return aiBuildingResourceCost(portType)
		}
	}
	return economy.ResourceCost{}
}
