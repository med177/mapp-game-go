package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	aiRallyMaxWaitTurns      = 1
	aiRallyForceSharePercent = 60
)

// ensurePlanRallyState güvenli rally bölgesini kalıcı plan state'inde tutar.
// Mevcut rally hâlâ aynı aktif hedefe komşu ve güvenliyse deadline sıfırlanmaz.
func ensurePlanRallyState(ctx *StrategicContext) {
	if ctx == nil || ctx.gs == nil {
		return
	}
	plan := ctx.gs.AIPlans[ctx.FactionID]
	if plan == nil || plan.Kind != state.AIObjectiveExpand || plan.TargetFactionID == "" {
		clearPlanRally(plan)
		return
	}
	_, offensiveTarget := aiOffensiveAnchor(ctx, plan)
	if offensiveTarget == "" {
		clearPlanRally(plan)
		return
	}

	if plan.RallyRegionID != "" && aiRallyRegionSafeForTarget(ctx, plan.RallyRegionID, offensiveTarget) {
		ctx.RallyRegionID = plan.RallyRegionID
		ctx.RallyDeadlineTurn = plan.RallyDeadlineTurn
		return
	}

	rallyRegionID := selectSafeRallyRegion(ctx, offensiveTarget)
	if rallyRegionID == "" {
		clearPlanRally(plan)
		return
	}
	plan.RallyRegionID = rallyRegionID
	plan.RallyDeadlineTurn = ctx.gs.Turn + aiRallyMaxWaitTurns
	ctx.RallyRegionID = plan.RallyRegionID
	ctx.RallyDeadlineTurn = plan.RallyDeadlineTurn
}

func clearPlanRally(plan *state.AIPlanState) {
	if plan == nil {
		return
	}
	plan.RallyRegionID = ""
	plan.RallyDeadlineTurn = 0
}

func selectSafeRallyRegion(ctx *StrategicContext, target faction.FactionID) world.RegionID {
	if ctx == nil || ctx.gs == nil || target == "" {
		return ""
	}
	bestID := world.RegionID("")
	bestScore := -int(^uint(0)>>1) - 1
	for _, front := range ctx.Fronts {
		if front.EnemyFactionID != target {
			continue
		}
		for _, regionID := range front.FriendlyRegions {
			if !aiRallyRegionSafeForTarget(ctx, regionID, target) {
				continue
			}
			score := aiRallyRegionSafetyScore(ctx, regionID, target)
			if score > bestScore || (score == bestScore && (bestID == "" || regionID < bestID)) {
				bestID = regionID
				bestScore = score
			}
		}
	}
	return bestID
}

func aiRallyRegionSafeForTarget(ctx *StrategicContext, regionID world.RegionID, target faction.FactionID) bool {
	if ctx == nil || ctx.gs == nil || regionID == "" || target == "" {
		return false
	}
	region := ctx.gs.Regions[regionID]
	if region == nil || region.IsSea || region.OwnerID != string(ctx.FactionID) || ctx.gs.SiegeAt(regionID) != nil {
		return false
	}
	sharesTargetBorder := false
	for _, neighborID := range region.Neighbors {
		neighbor := ctx.gs.Regions[neighborID]
		if neighbor != nil && !neighbor.IsSea && neighbor.OwnerID == string(target) {
			sharesTargetBorder = true
			break
		}
	}
	if !sharesTargetBorder {
		return false
	}
	for _, armyRef := range aiSortedArmies(ctx.gs) {
		if armyRef.IsNaval || armyRef.RegionID != regionID || armyRef.OwnerID == string(ctx.FactionID) {
			continue
		}
		return false
	}
	return true
}

func aiRallyRegionSafetyScore(ctx *StrategicContext, regionID world.RegionID, target faction.FactionID) int {
	region := ctx.gs.Regions[regionID]
	if region == nil {
		return -1 << 30
	}
	demand, capacity, overload := aiRegionLogistics(ctx.gs, region, string(ctx.FactionID))
	friendlyPower := 0
	enemyPower := 0
	for _, armyRef := range aiSortedArmies(ctx.gs) {
		if armyRef.IsNaval {
			continue
		}
		if armyRef.OwnerID == string(ctx.FactionID) && armyRef.RegionID == regionID {
			friendlyPower += armyRef.TotalStrength(ctx.gs.UnitTypes)
			continue
		}
		if armyRef.OwnerID != string(target) {
			continue
		}
		for _, neighborID := range region.Neighbors {
			if armyRef.RegionID == neighborID {
				enemyPower += armyRef.TotalStrength(ctx.gs.UnitTypes)
				break
			}
		}
	}
	score := friendlyPower - enemyPower + maxInt(0, capacity-demand)*5 - overload*12
	score += ctx.strategicRegionValue(region) / 5
	if region.IsFortified() {
		score += 30
	}
	if ctx.gs.IsCapitalRegion(region) {
		score += 20
	}
	return score
}

// applyRallyAssignments saldırı ve kuşatma rollerini rally noktasında toplar.
// İki uygun ordu ve gerekli güç oluştuğunda deadline mevcut tura çekilerek hazırlık
// kalıcı biçimde tamamlanır. Üç turluk deadline dolduğunda eksik kuvvetle de bekleme biter.
func applyRallyAssignments(ctx *StrategicContext) {
	if ctx == nil || ctx.gs == nil {
		return
	}
	plan := ctx.gs.AIPlans[ctx.FactionID]
	if plan == nil || plan.RallyRegionID == "" {
		return
	}

	type offensiveArmy struct {
		armyRef    *army.Army
		assignment AIArmyAssignment
	}
	var offensive []offensiveArmy
	for armyID, assignment := range ctx.ArmyAssignments {
		if assignment.Role != AIArmyRoleAssault && assignment.Role != AIArmyRoleSiege {
			continue
		}
		armyRef := ctx.gs.Armies[armyID]
		if armyRef == nil || armyRef.IsNaval || armyRef.IsGarrison || ctx.gs.SiegeByArmy(armyID) != nil {
			continue
		}
		offensive = append(offensive, offensiveArmy{armyRef: armyRef, assignment: assignment})
	}
	sort.Slice(offensive, func(i, j int) bool { return offensive[i].armyRef.ID < offensive[j].armyRef.ID })
	if len(offensive) < 2 {
		clearPlanRally(plan)
		ctx.RallyRegionID = ""
		ctx.RallyDeadlineTurn = 0
		ctx.RallyReady = true
		return
	}

	totalPower := 0
	gatheredPower := 0
	gatheredArmies := 0
	for _, candidate := range offensive {
		power := candidate.armyRef.TotalStrength(ctx.gs.UnitTypes)
		totalPower += power
		if candidate.armyRef.RegionID == plan.RallyRegionID {
			gatheredPower += power
			gatheredArmies++
		}
	}
	rallyTarget := plan.TargetFactionID
	for _, candidate := range offensive {
		if candidate.assignment.FrontFactionID != "" {
			rallyTarget = candidate.assignment.FrontFactionID
			break
		}
	}
	targetFrontierPower := aiFrontierPower(ctx.gs, rallyTarget, ctx.FactionID)
	requiredPower := maxInt(
		percentageCeil(totalPower, aiRallyForceSharePercent),
		percentageCeil(targetFrontierPower, aiMinAttackPowerPercent(ctx.gs)),
	)
	ctx.RallyRegionID = plan.RallyRegionID
	ctx.RallyDeadlineTurn = plan.RallyDeadlineTurn
	ctx.RallyRequiredPower = requiredPower
	ctx.RallyGatheredPower = gatheredPower

	ready := gatheredArmies >= 2 && gatheredPower >= requiredPower
	deadlineReached := plan.RallyDeadlineTurn <= ctx.gs.Turn
	if ready || deadlineReached {
		if ready && !deadlineReached {
			plan.RallyDeadlineTurn = ctx.gs.Turn
			ctx.RallyDeadlineTurn = plan.RallyDeadlineTurn
		}
		ctx.RallyReady = true
		return
	}

	ctx.RallyActive = true
	for _, candidate := range offensive {
		assignment := candidate.assignment
		assignment.AnchorRegionID = plan.RallyRegionID
		assignment.Rallying = true
		assignment.Reason = "koordineli hücum için rally"
		ctx.ArmyAssignments[candidate.armyRef.ID] = assignment
	}
}

func percentageCeil(value, percent int) int {
	if value <= 0 || percent <= 0 {
		return 0
	}
	return (value*percent + 99) / 100
}
