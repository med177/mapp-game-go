package ai

import (
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
	cap := gs.ManpowerCap(fid)
	deployed := gs.DeployedLandUnits(fid) + aiPendingLandUnitCount(gs, fid)
	barracksCost := economy.ResourceCost{Gold: 150}
	if b, ok2 := gs.BuildingTypes["barracks"]; ok2 {
		barracksCost = economy.ResourceCost{Gold: b.GoldCost, Grain: b.GrainCost, Iron: b.IronCost, Timber: b.TimberCost, Stone: b.StoneCost}
	}
	if cap-deployed <= state.ManpowerPerRegion && aiCanAffordForBudget(f, barracksCost, budget, aiBudgetArmy) {
		aiBuildBarracksWithBudgetAndSteps(gs, fid, barracksCost, budget, steps)
	}
	for {
		if gs.DeployedLandUnits(fid)+aiPendingLandUnitCount(gs, fid) >= gs.ManpowerCap(fid) || !aiCanAffordForBudget(f, economy.ResourceCost{Gold: aiMilitiaCost}, budget, aiBudgetArmy) {
			break
		}
		if !aiRecruitOneWithStrategicContextAndSteps(gs, fid, budget, strategicContext, steps) {
			break
		}
	}
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
	unitTypeID := aiSelectBestUnitForStrategicContext(gs, f, budget, strategicContext)
	if unitTypeID == "" {
		return false
	}
	utype, ok := gs.UnitTypes[unitTypeID]
	if !ok {
		return false
	}
	unitCost := economy.ResourceCost{Gold: utype.GoldCost, Grain: utype.GrainCost, Iron: utype.IronCost, Timber: utype.TimberCost, Stone: utype.StoneCost}
	if !aiCanAffordForBudget(f, unitCost, budget, aiBudgetArmy) {
		return false
	}
	recruitRegion := aiFindRecruitRegionForStrategicContext(gs, fid, utype, strategicContext)
	if recruitRegion == "" || aiPendingUnitCountByRegion(gs, recruitRegion, fid) >= aiMaxRegionQueue || !aiCanQueueLandUnit(gs, fid, recruitRegion, utype) {
		return false
	}
	if !aiApplyBudgetedCost(f, unitCost, budget, aiBudgetArmy) {
		return false
	}
	aiEnqueueProduction(gs, fid, aiProductionKindUnit, recruitRegion, unitTypeID, utype.TurnsRequired)
	return true
}
