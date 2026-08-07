package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	// Bir kara birimi, senaryo nüfus ölçeğinde yaklaşık bu kadar nüfustan
	// düzenli askerî rezerv çıkarılmasını temsil eder. ManpowerCap yine mutlak
	// üst sınırdır; bu sayı yalnız AI'nin ulaşmaya çalışacağı tabanı belirler.
	aiReservePopulationPerLandUnit    = 200
	aiReserveCoastalRegionsPerWarship = 2
	// Tarihsel fetih hedefi için AI, hedefin yalnız eşiti değil belirgin
	// üstünlüğü olacak kadar kara gücü kurar.
	aiExpansionPowerAdvantagePercent = 135
)

// aiForceRequirement, kapasiteyi değil devletin ekonomik ve coğrafi ölçeğine
// uygun asgarî kuvvet hedefini taşır. Böylece küçük mevcut ordu üzerinden
// hesaplanan dinamik hareket rezervi, askerî üretimi sıfıra kilitlemez.
type aiForceRequirement struct {
	LandTarget               int
	LandPresent              int
	LandPending              int
	ObjectiveEnemyLandPower  int
	ObjectiveLandPowerTarget int
	CoastalRegions           int
	WarshipTarget            int
	WarshipsPresent          int
	WarshipsPending          int
}

func aiForceRequirements(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext) aiForceRequirement {
	requirement := aiForceRequirement{}
	if gs == nil || fid == "" {
		return requirement
	}

	landRegions := 0
	population := 0
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		landRegions++
		population += maxInt(0, region.Population)
		if region.IsCoastal(gs.Regions) {
			requirement.CoastalRegions++
		}
	}
	if landRegions == 0 {
		return requirement
	}

	requirement.LandTarget = (population + aiReservePopulationPerLandUnit - 1) / aiReservePopulationPerLandUnit
	// Çok düşük nüfuslu veya yeni açılmış bölgelerde de devlet tamamen ordusuz
	// kalmasın; bu taban nüfus oranının yerini almaz, yalnız güvenli alt sınırdır.
	requirement.LandTarget = maxInt(requirement.LandTarget, maxInt(1, (landRegions+2)/3))
	planKind := state.AIObjectiveConsolidate
	if plan := gs.AIPlans[fid]; plan != nil && plan.Kind != "" {
		planKind = plan.Kind
	}
	if aiFactionAtWar(gs, string(fid)) || (ctx != nil && ctx.CriticalThreat) {
		requirement.LandTarget = requirement.LandTarget * 3 / 2
	} else if planKind == state.AIObjectiveExpand {
		requirement.LandTarget = (requirement.LandTarget*5 + 3) / 4
	} else if planKind == state.AIObjectiveDefend {
		requirement.LandTarget = (requirement.LandTarget*6 + 4) / 5
	}
	requirement.LandTarget = minInt(requirement.LandTarget, gs.ManpowerCap(fid))
	requirement.applyExpansionPowerTarget(gs, fid, planKind)

	if requirement.CoastalRegions > 0 {
		requirement.WarshipTarget = (requirement.CoastalRegions + aiReserveCoastalRegionsPerWarship - 1) / aiReserveCoastalRegionsPerWarship
		if aiHasNavalFocus(gs, fid) {
			// Denizci cumhuriyetler ve okyanus güçleri, yalnız kıyı savunması
			// yapmaz: her kıyı için iki savaş gemisi ve en az altı gemilik
			// ana filo hedefler. Bu yönelim senaryo JSON'undan gelir.
			requirement.WarshipTarget = maxInt(6, requirement.CoastalRegions*2)
		}
		if aiFactionAtWar(gs, string(fid)) || (ctx != nil && (ctx.CriticalThreat || len(ctx.NavalThreats) > 0)) {
			requirement.WarshipTarget *= 2
		} else if planKind == state.AIObjectiveExpand {
			requirement.WarshipTarget++
		}
	}

	for _, armyRef := range aiSortedArmies(gs) {
		if armyRef == nil || armyRef.OwnerID != string(fid) {
			continue
		}
		if !armyRef.IsNaval {
			requirement.LandPresent += len(armyRef.Units)
			continue
		}
		for _, unit := range armyRef.Units {
			if unitType := gs.UnitTypes[unit.TypeID]; unitType != nil && unitType.Category == army.CategoryNavalWar {
				requirement.WarshipsPresent++
			}
		}
	}
	for _, order := range gs.ProductionQueue {
		if order.Kind != aiProductionKindUnit || order.FactionID != string(fid) {
			continue
		}
		unitType := gs.UnitTypes[order.TypeID]
		if unitType == nil {
			continue
		}
		if aiLandUnitCategory(unitType.Category) {
			requirement.LandPending++
		} else if unitType.Category == army.CategoryNavalWar {
			requirement.WarshipsPending++
		}
	}
	return requirement
}

func (requirement *aiForceRequirement) applyExpansionPowerTarget(gs *state.GameState, fid faction.FactionID, planKind state.AIObjectiveKind) {
	if requirement == nil || gs == nil || planKind != state.AIObjectiveExpand {
		return
	}
	plan := gs.AIPlans[fid]
	if plan == nil || plan.TargetFactionID == "" || plan.TargetFactionID == fid {
		return
	}
	requirement.ObjectiveEnemyLandPower = aiFactionProjectedLandPower(gs, plan.TargetFactionID)
	if requirement.ObjectiveEnemyLandPower <= 0 {
		return
	}
	requirement.ObjectiveLandPowerTarget = (requirement.ObjectiveEnemyLandPower*aiExpansionPowerAdvantagePercent + 99) / 100
	selfPower := aiFactionProjectedLandPower(gs, fid)
	if selfPower >= requirement.ObjectiveLandPowerTarget {
		return
	}
	unitPower := aiExpectedLandReserveUnitPower(gs, fid)
	additionalUnits := (requirement.ObjectiveLandPowerTarget - selfPower + unitPower - 1) / unitPower
	requirement.LandTarget = minInt(gs.ManpowerCap(fid), maxInt(requirement.LandTarget, requirement.LandPresent+requirement.LandPending+additionalUnits))
}

// aiFactionProjectedLandPower counts both active and queued land units, so a
// state does not misjudge an opponent whose barracks are about to complete.
func aiFactionProjectedLandPower(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" {
		return 0
	}
	power := 0
	for _, armyRef := range aiSortedArmies(gs) {
		if armyRef != nil && armyRef.OwnerID == string(fid) && !armyRef.IsNaval {
			power += armyRef.TotalStrength(gs.UnitTypes)
		}
	}
	for _, order := range gs.ProductionQueue {
		if order.Kind == aiProductionKindUnit && order.FactionID == string(fid) {
			if unitType := gs.UnitTypes[order.TypeID]; unitType != nil && aiLandUnitCategory(unitType.Category) {
				power += aiUnitCombatPower(unitType)
			}
		}
	}
	return power
}

func aiExpectedLandReserveUnitPower(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" {
		return 10
	}
	self := gs.Factions[fid]
	if self != nil {
		if unitID := aiSelectReserveLandUnitForProcurement(gs, self, nil); unitID != "" {
			if unitType := gs.UnitTypes[unitID]; unitType != nil {
				return maxInt(1, aiUnitCombatPower(unitType))
			}
		}
	}
	// Kışla daha kurulmamışsa bile nüfus rezervini hedefe çevirebilmek için
	// teknolojiye açık temel piyade referansını kullan.
	if infantry := gs.UnitTypes["infantry"]; infantry != nil && (self == nil || infantry.HasAllRequiredTechs(self.Research.Completed)) {
		return maxInt(1, aiUnitCombatPower(infantry))
	}
	if militia := gs.UnitTypes["militia"]; militia != nil {
		return maxInt(1, aiUnitCombatPower(militia))
	}
	return 10
}

func aiUnitCombatPower(unitType *army.UnitType) int {
	if unitType == nil {
		return 0
	}
	return maxInt(1, unitType.Attack+unitType.Morale/10)
}

func aiHasNavalFocus(gs *state.GameState, fid faction.FactionID) bool {
	if gs == nil || fid == "" {
		return false
	}
	return gs.AIStrategies[string(fid)].NavalFocus
}

func aiLandReserveShortfall(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext) int {
	requirement := aiForceRequirements(gs, fid, ctx)
	return maxInt(0, requirement.LandTarget-requirement.LandPresent-requirement.LandPending)
}

func aiWarshipReserveShortfall(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext) int {
	requirement := aiForceRequirements(gs, fid, ctx)
	return maxInt(0, requirement.WarshipTarget-requirement.WarshipsPresent-requirement.WarshipsPending)
}

// aiFindReserveRecruitRegion is the safe interior fallback used only while a
// state is below its population-based force floor. It keeps actual production,
// throughput and local-logistics rules, but does not require a route to an
// offensive rally point: a new state must first be able to raise its reserve.
func aiFindReserveRecruitRegion(gs *state.GameState, fid faction.FactionID, unitType *army.UnitType, ctx *StrategicContext) world.RegionID {
	if gs == nil || unitType == nil || fid == "" {
		return ""
	}
	requiredBuilding := unitType.RequiredBldg
	if requiredBuilding == "" {
		requiredBuilding = "barracks"
	}
	requiredLevel := maxInt(1, unitType.RequiredBldgLevel)

	type candidate struct {
		region    *world.Region
		remaining int
		level     int
	}
	var candidates []candidate
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.IsLocked || region.OwnerID != string(fid) || aiBuildingLevel(region, requiredBuilding) < requiredLevel {
			continue
		}
		if !aiCanQueueLandUnit(gs, fid, region.ID, unitType) || aiPendingUnitCountByRegion(gs, region.ID, fid) >= aiMaxRegionQueue {
			continue
		}
		remaining := aiLaneRemainingCapacity(gs, region.ID, fid, unitType)
		if remaining <= 0 || !aiRecruitmentRegionSecure(gs, fid, ctx, region) {
			continue
		}
		demand, capacity, _ := aiProjectedRecruitRegionLogistics(gs, fid, region, unitType)
		if demand > capacity {
			continue
		}
		candidates = append(candidates, candidate{region: region, remaining: remaining, level: aiBuildingLevel(region, requiredBuilding)})
	}
	if len(candidates) == 0 {
		return ""
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].remaining != candidates[j].remaining {
			return candidates[i].remaining > candidates[j].remaining
		}
		if candidates[i].level != candidates[j].level {
			return candidates[i].level > candidates[j].level
		}
		return candidates[i].region.ID < candidates[j].region.ID
	})
	return candidates[0].region.ID
}

// aiSelectReserveLandUnit preserves the strategic composition selection when
// possible. If that selection has no route-valid line, it chooses the most
// sustainable legal unit that the reserve fallback can actually produce.
func aiSelectReserveLandUnit(gs *state.GameState, self *faction.Faction, budget *aiBudget, ctx *StrategicContext) string {
	if selected := aiSelectBestUnitForStrategicContext(gs, self, budget, ctx); selected != "" {
		return selected
	}
	selected := aiSelectReserveLandUnitForProcurement(gs, self, ctx)
	if selected == "" || !aiUnitAvailableForBudget(gs, self, gs.UnitTypes[selected], budget) {
		return ""
	}
	return selected
}

// aiSelectReserveLandUnitForProcurement uses the same composition score as
// normal 1300 recruitment but deliberately ignores the current material stock.
// The selected recipe is therefore able to create a grain/iron/cloth demand
// instead of falling back to a cheaper unit merely because a resource is empty.
func aiSelectReserveLandUnitForProcurement(gs *state.GameState, self *faction.Faction, ctx *StrategicContext) string {
	if gs == nil || self == nil {
		return ""
	}
	if selected := aiSelectStrategicLandUnitForProcurement(gs, self, ctx); selected != "" {
		return selected
	}
	if ctx == nil {
		ctx = prepareStrategicContext(gs, self.ID)
	}
	plan := gs.AIPlans[self.ID]
	target := aiCompositionTargetForStrategicContext(gs, self.ID, plan, ctx)
	composition := aiFactionLandComposition(gs, self.ID)
	needs := aiBuildRecruitmentBattleNeeds(gs, self.ID, ctx)
	economySnapshot := aiBuildEconomySnapshot(gs, self.ID)
	unitIDs := make([]string, 0, len(gs.UnitTypes))
	for unitID := range gs.UnitTypes {
		unitIDs = append(unitIDs, unitID)
	}
	sort.Strings(unitIDs)
	var best aiUnitCandidate
	found := false
	for _, unitID := range unitIDs {
		unitType := gs.UnitTypes[unitID]
		if unitType == nil || !aiLandUnitCategory(unitType.Category) || !unitType.HasAllRequiredTechs(self.Research.Completed) || self.Gold-unitType.GoldCost < aiMinGoldReserve+unitType.GoldUpkeep*aiGoldUpkeepReserveTurns {
			continue
		}
		if aiFindReserveRecruitRegion(gs, self.ID, unitType, ctx) == "" {
			continue
		}
		candidate := aiScoreLandUnitCandidate(gs, self, unitType, target, composition, needs, economySnapshot, false)
		if !found || aiUnitCandidateBetter(candidate, best) {
			best, found = candidate, true
		}
	}
	if !found {
		return ""
	}
	return best.TypeID
}
