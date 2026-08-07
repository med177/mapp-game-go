package ai

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

// aiRecruitAndBuild orchestrates barracks construction and land-unit queueing;
// unit selection and region scoring remain in their dedicated strategy helpers.
func aiRecruitAndBuild(gs *state.GameState, fid faction.FactionID) {
	aiRecruitAndBuildWithSteps(gs, fid, nil)
}

func aiRecruitAndBuildWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	aiRecruitAndBuildWithBudgetAndSteps(gs, fid, nil, steps)
}

func aiRecruitAndBuildWithBudgetAndSteps(gs *state.GameState, fid faction.FactionID, budget *aiBudget, steps *[]TurnStep) {
	aiRecruitAndBuildWithStrategicContextAndSteps(gs, fid, budget, nil, steps)
}

func aiRecruitAndBuildWithStrategicContextAndSteps(gs *state.GameState, fid faction.FactionID, budget *aiBudget, strategicContext *StrategicContext, steps *[]TurnStep) {
	f, ok := gs.Factions[fid]
	if !ok || f.IsEliminated {
		return
	}
	reserveShortfall := aiLandReserveShortfall(gs, fid, strategicContext)
	deployed := gs.DeployedLandUnits(fid) + aiPendingLandUnitCount(gs, fid)
	barracksCost := aiBarracksResourceCost(gs)
	if aiNeedsBarracksForMilitaryProduction(gs, fid, strategicContext, gs.ManpowerCap(fid)-deployed) && aiCanAffordForBudget(f, barracksCost, budget, aiBudgetArmy) {
		aiBuildBarracksWithBudgetAndSteps(gs, fid, barracksCost, budget, steps)
	}
	for reserveShortfall > 0 {
		if !aiCanAffordForBudget(f, economy.ResourceCost{Gold: aiMilitiaCost}, budget, aiBudgetArmy) {
			break
		}
		if !aiRecruitOneWithStrategicContextAndSteps(gs, fid, budget, strategicContext, steps) {
			break
		}
		reserveShortfall = aiLandReserveShortfall(gs, fid, strategicContext)
	}
}

func aiBarracksResourceCost(gs *state.GameState) economy.ResourceCost {
	if gs == nil || gs.BuildingTypes == nil {
		return economy.ResourceCost{Gold: 150}
	}
	if btype := gs.BuildingTypes["barracks"]; btype != nil {
		return aiBuildingResourceCost(btype)
	}
	return economy.ResourceCost{Gold: 150}
}

func aiHasBuildableBarracksRegion(gs *state.GameState, fid faction.FactionID) bool {
	if gs == nil || gs.BuildingTypes == nil || gs.BuildingTypes["barracks"] == nil {
		return false
	}
	btype := gs.BuildingTypes["barracks"]
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.OwnerID != string(fid) || gs.SiegeAt(region.ID) != nil {
			continue
		}
		queued := aiQueuedBuildingCount(gs, region.ID, "barracks", fid)
		if aiBuildingLevel(region, "barracks")+queued >= btype.MaxPerRegion || !aiBuildingAllowed(gs, region, "barracks", btype.RequiredTerrain) {
			continue
		}
		return true
	}
	return false
}

// aiPotentialRecruitmentRegionAfterBarracks, mevcut kışlası olmayan ancak
// kışla tamamlandığında bir askerî üretim sırasına ev sahipliği yapabilecek
// bölgeyi bulur. Bu yalnızca tedarik/önkoşul kararı içindir; gerçek sıra
// seçimi yine aiFindStrategicRecruitRegion tarafından yapılır.
func aiPotentialRecruitmentRegionAfterBarracks(gs *state.GameState, fid faction.FactionID, unitType *army.UnitType, ctx *StrategicContext) bool {
	if gs == nil || unitType == nil || ctx == nil {
		return false
	}
	requiredBuilding := unitType.RequiredBldg
	if requiredBuilding != "" && requiredBuilding != "barracks" {
		return false
	}
	requiredLevel := maxInt(1, unitType.RequiredBldgLevel)
	btype := gs.BuildingTypes["barracks"]
	if btype == nil {
		return false
	}
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.IsLocked || region.OwnerID != string(fid) || gs.SiegeAt(region.ID) != nil {
			continue
		}
		queuedBuildings := aiQueuedBuildingCount(gs, region.ID, "barracks", fid)
		level := aiBuildingLevel(region, "barracks")
		if level+queuedBuildings >= btype.MaxPerRegion || level+1 < requiredLevel || !aiBuildingAllowed(gs, region, "barracks", btype.RequiredTerrain) {
			continue
		}
		if aiPendingUnitCountByRegion(gs, region.ID, fid) >= aiMaxRegionQueue || !aiCanQueueLandUnit(gs, fid, region.ID, unitType) {
			continue
		}
		if !aiRecruitmentRegionSecure(gs, fid, ctx, region) {
			continue
		}
		demand, capacity, _ := aiProjectedRecruitRegionLogistics(gs, fid, region, unitType)
		if demand > capacity {
			continue
		}
		return true
	}
	return false
}

// aiNeedsBarracksForMilitaryProduction tek bir karar sözleşmesi olarak hem
// tedarik öncesi kışla maliyetini hem de gerçek askerî üretim adımını besler.
// Böylece kışla için gereken demir/kereste satın alınmadan kışla üretimi bloke
// olmaz; mevcut ordu limiti doluysa gereksiz kışla stoğu da kurulmaz.
func aiNeedsBarracksForMilitaryProduction(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext, spareManpower int) bool {
	if gs == nil || gs.BuildingTypes == nil || gs.BuildingTypes["barracks"] == nil || !aiHasBuildableBarracksRegion(gs, fid) {
		return false
	}
	if aiFactionBarracksCount(gs, fid) == 0 || spareManpower <= state.ManpowerPerRegion {
		return true
	}
	if gs.CurrentLandArmies(fid) >= gs.MaxLandArmies(fid) {
		return false
	}
	self := gs.Factions[fid]
	if self == nil {
		return false
	}
	for _, unitType := range gs.UnitTypes {
		if unitType == nil || !aiLandUnitCategory(unitType.Category) || !unitType.HasAllRequiredTechs(self.Research.Completed) || self.Gold-unitType.GoldCost < aiMinGoldReserve+unitType.GoldUpkeep*aiGoldUpkeepReserveTurns {
			continue
		}
		if aiPotentialRecruitmentRegionAfterBarracks(gs, fid, unitType, ctx) {
			return true
		}
	}
	return false
}

func aiFactionBarracksCount(gs *state.GameState, fid faction.FactionID) int {
	count := 0
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		count += aiBuildingLevel(region, "barracks")
	}
	for _, order := range gs.ProductionQueue {
		if order.FactionID == string(fid) && order.Kind == aiProductionKindBuilding && order.TypeID == "barracks" {
			count++
		}
	}
	return count
}

func aiBuildBarracks(gs *state.GameState, fid faction.FactionID, cost economy.ResourceCost) {
	aiBuildBarracksWithSteps(gs, fid, cost, nil)
}

func aiBuildBarracksWithSteps(gs *state.GameState, fid faction.FactionID, cost economy.ResourceCost, steps *[]TurnStep) {
	aiBuildBarracksWithBudgetAndSteps(gs, fid, cost, nil, steps)
}

func aiBuildBarracksWithBudgetAndSteps(gs *state.GameState, fid faction.FactionID, cost economy.ResourceCost, budget *aiBudget, steps *[]TurnStep) {
	f := gs.Factions[fid]
	btype := gs.BuildingTypes["barracks"]
	if btype == nil {
		return
	}
	for _, r := range aiSortedRegions(gs) {
		if r.OwnerID != string(fid) || r.IsSea {
			continue
		}
		queued := aiQueuedBuildingCount(gs, r.ID, "barracks", fid)
		if aiBuildingLevel(r, "barracks")+queued >= btype.MaxPerRegion || !aiBuildingAllowed(gs, r, "barracks", btype.RequiredTerrain) {
			continue
		}
		if budget == nil {
			cost.Apply(f)
		} else if !aiApplyBudgetedCost(f, cost, budget, aiBudgetArmy) {
			return
		}
		turns := aiBuildingTurnsRequired(r, "barracks", btype.TurnsRequired, queued)
		aiEnqueueProduction(gs, fid, aiProductionKindBuilding, r.ID, "barracks", turns)
		addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepBuild, TargetRegion: r.ID, FocusRegion: r.ID, Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, r.ID) + " bölgesinde kışla kuruyor."})
		return
	}
}

func aiRecruitOne(gs *state.GameState, fid faction.FactionID) bool {
	return aiRecruitOneWithSteps(gs, fid, nil)
}

func aiRecruitOneWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) bool {
	return aiRecruitOneWithBudgetAndSteps(gs, fid, nil, steps)
}

func aiRecruitOneWithBudgetAndSteps(gs *state.GameState, fid faction.FactionID, budget *aiBudget, steps *[]TurnStep) bool {
	return aiRecruitOneWithStrategicContextAndSteps(gs, fid, budget, nil, steps)
}

func aiRecruitOneWithStrategicContextAndSteps(gs *state.GameState, fid faction.FactionID, budget *aiBudget, strategicContext *StrategicContext, _ *[]TurnStep) bool {
	f := gs.Factions[fid]
	if gs.UnitTypes == nil {
		return false
	}
	reserveShortfall := aiLandReserveShortfall(gs, fid, strategicContext)
	unitTypeID := aiSelectBestUnitForStrategicContext(gs, f, budget, strategicContext)
	if reserveShortfall > 0 {
		unitTypeID = aiSelectReserveLandUnit(gs, f, budget, strategicContext)
	}
	if unitTypeID == "" {
		return false
	}
	utype, ok := gs.UnitTypes[unitTypeID]
	if !ok {
		return false
	}
	unitCost := economy.ResourceCost{Gold: utype.GoldCost, Grain: utype.GrainCost, Iron: utype.IronCost, Timber: utype.TimberCost, Stone: utype.StoneCost, Spice: utype.SpiceCost, Cloth: utype.ClothCost}
	if !aiCanAffordUnitForBudget(f, utype, budget, aiBudgetArmy) {
		return false
	}
	recruitRegion := aiFindRecruitRegionForStrategicContext(gs, fid, utype, strategicContext)
	if recruitRegion == "" && reserveShortfall > 0 {
		recruitRegion = aiFindReserveRecruitRegion(gs, fid, utype, strategicContext)
	}
	if recruitRegion == "" || aiPendingUnitCountByRegion(gs, recruitRegion, fid) >= aiMaxRegionQueue || !aiCanQueueLandUnit(gs, fid, recruitRegion, utype) {
		return false
	}
	if !aiApplyBudgetedCost(f, unitCost, budget, aiBudgetArmy) {
		return false
	}
	aiEnqueueProduction(gs, fid, aiProductionKindUnit, recruitRegion, unitTypeID, utype.TurnsRequired)
	return true
}
