package ai

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	aiRetreatStrengthPercent   = 45
	aiRetreatEnemyPowerPercent = 135
	aiSiegeReliefPowerPercent  = 150
)

// applyRetreatAssignments açık arazide ağır yıpranmış veya yerel olarak ezilen
// orduları güvenli ikmal bölgelerine çeker. Aktif kuşatmalar yalnız ikmal aşımı
// ve ezici yaklaşan yardım gücü birlikte varsa bu kurala girer.
func applyRetreatAssignments(ctx *StrategicContext) {
	if ctx == nil || ctx.gs == nil || ctx.FactionID == "" {
		return
	}
	for _, armyRef := range aiSortedArmies(ctx.gs) {
		if armyRef.OwnerID != string(ctx.FactionID) || armyRef.IsNaval || armyRef.IsGarrison || len(armyRef.Units) == 0 {
			continue
		}

		activeSiege := ctx.gs.SiegeByArmy(armyRef.ID)
		reason := ""
		if activeSiege != nil {
			if !aiShouldWithdrawFromSiege(ctx, armyRef, activeSiege) {
				continue
			}
			reason = "ikmal aşımı ve ezici düşman yardım gücü"
		} else {
			weak := aiArmyStrengthBelowPercent(armyRef, ctx.gs.UnitTypes, aiRetreatStrengthPercent)
			enemyPower := aiLocalHostilePower(ctx.gs, armyRef)
			ownPower := armyRef.TotalStrength(ctx.gs.UnitTypes)
			outmatched := enemyPower*100 >= ownPower*aiRetreatEnemyPowerPercent
			if !weak && !outmatched {
				continue
			}
			if weak {
				reason = "ordu gücü yüzde 45 altına düştü"
			} else {
				reason = "yerel düşman gücü yüzde 135 eşiğini aştı"
			}
		}

		anchor := selectSafeRecoveryRegion(ctx, armyRef)
		if anchor == "" {
			continue
		}
		ctx.ArmyAssignments[armyRef.ID] = AIArmyAssignment{
			Role:           AIArmyRoleRetreat,
			AnchorRegionID: anchor,
			Reason:         reason,
		}
	}
}

// Karşılaştırma yuvarlama kaybı olmadan, birimlerin deneyim dahil tam-can
// saldırı gücünü CurrentHP ile ağırlıklandırır. Tam yüzde 45 geri çekilmez;
// eşik yalnız bunun altıdır.
func aiArmyStrengthBelowPercent(armyRef *army.Army, unitTypes map[string]*army.UnitType, percent int) bool {
	if armyRef == nil || len(armyRef.Units) == 0 || percent <= 0 {
		return false
	}
	currentWeighted := 0
	maximumWeighted := 0
	for index := range armyRef.Units {
		unit := &armyRef.Units[index]
		unitType := unitTypes[unit.TypeID]
		if unitType == nil {
			continue
		}
		basePower := unit.EffectiveAttack(unitTypes) + unitType.Morale/10
		if basePower <= 0 {
			continue
		}
		hp := unit.CurrentHP
		if hp < 0 {
			hp = 0
		}
		if hp > army.MaxUnitHP {
			hp = army.MaxUnitHP
		}
		currentWeighted += basePower * hp
		maximumWeighted += basePower * army.MaxUnitHP
	}
	return maximumWeighted > 0 && currentWeighted*100 < maximumWeighted*percent
}

// aiLocalHostilePower aynı ve komşu bölgelerde savaş halinde olunan kara
// ordularının gücünü toplar.
func aiLocalHostilePower(gs *state.GameState, armyRef *army.Army) int {
	if gs == nil || armyRef == nil {
		return 0
	}
	region := gs.Regions[armyRef.RegionID]
	if region == nil {
		return 0
	}
	localRegions := make(map[world.RegionID]struct{}, len(region.Neighbors)+1)
	localRegions[region.ID] = struct{}{}
	for _, neighborID := range region.Neighbors {
		localRegions[neighborID] = struct{}{}
	}
	return aiHostilePowerInRegions(gs, faction.FactionID(armyRef.OwnerID), localRegions)
}

func aiHostilePowerInRegions(gs *state.GameState, owner faction.FactionID, regionIDs map[world.RegionID]struct{}) int {
	if gs == nil || owner == "" || len(regionIDs) == 0 {
		return 0
	}
	total := 0
	for _, candidate := range aiSortedArmies(gs) {
		if candidate.IsNaval || candidate.OwnerID == string(owner) {
			continue
		}
		if _, local := regionIDs[candidate.RegionID]; !local {
			continue
		}
		if diplomacy.SameRealm(gs, owner, faction.FactionID(candidate.OwnerID)) {
			continue
		}
		relation := diplomacy.Relation(gs, owner, faction.FactionID(candidate.OwnerID))
		if relation == nil || relation.Stance != faction.StanceWar {
			continue
		}
		total += candidate.TotalStrength(gs.UnitTypes)
	}
	return total
}

func aiShouldWithdrawFromSiege(ctx *StrategicContext, armyRef *army.Army, siege *state.SiegeState) bool {
	if ctx == nil || ctx.gs == nil || armyRef == nil || siege == nil {
		return false
	}
	supplyRegion := ctx.gs.Regions[armyRef.RegionID]
	_, _, overload := aiRegionLogistics(ctx.gs, supplyRegion, armyRef.OwnerID)
	if overload <= 0 && armyRef.OverCapacityTurns <= 0 {
		return false
	}
	target := ctx.gs.Regions[siege.RegionID]
	if target == nil {
		return false
	}
	adjacent := make(map[world.RegionID]struct{}, len(target.Neighbors))
	for _, neighborID := range target.Neighbors {
		adjacent[neighborID] = struct{}{}
	}
	reliefPower := aiHostilePowerInRegions(ctx.gs, faction.FactionID(armyRef.OwnerID), adjacent)
	return reliefPower*100 > armyRef.TotalStrength(ctx.gs.UnitTypes)*aiSiegeReliefPowerPercent
}

func selectSafeRecoveryRegion(ctx *StrategicContext, armyRef *army.Army) world.RegionID {
	if ctx == nil || ctx.gs == nil || armyRef == nil {
		return ""
	}
	routes := ctx.routesFor(armyRef, armyRef.RegionID, aiRouteFriendly, 0)
	bestID := world.RegionID("")
	bestDistance := int(^uint(0) >> 1)
	bestScore := -int(^uint(0)>>1) - 1
	for _, region := range aiSortedRegions(ctx.gs) {
		if !aiSafeRecoveryRegion(ctx.gs, ctx.FactionID, armyRef, region) {
			continue
		}
		distance, reachable := routes.distance(region.ID)
		if !reachable {
			continue
		}
		score := aiRecoveryRegionScore(ctx, armyRef, region)
		if distance < bestDistance || (distance == bestDistance && (score > bestScore || (score == bestScore && (bestID == "" || region.ID < bestID)))) {
			bestID = region.ID
			bestDistance = distance
			bestScore = score
		}
	}
	return bestID
}

func aiSafeRecoveryRegion(gs *state.GameState, fid faction.FactionID, armyRef *army.Army, region *world.Region) bool {
	if gs == nil || armyRef == nil || region == nil || region.IsSea || region.OwnerID != string(fid) {
		return false
	}
	if gs.SiegeAt(region.ID) != nil {
		return false
	}
	for _, candidate := range aiSortedArmies(gs) {
		if candidate.IsNaval || candidate.RegionID != region.ID || candidate.OwnerID == string(fid) {
			continue
		}
		return false
	}
	if aiRegionHasAdjacentHostileArmy(gs, fid, region) {
		return false
	}
	demand, capacity, overload := aiRegionLogistics(gs, region, string(fid))
	if region.ID != armyRef.RegionID {
		demand += gs.RegionalArmyGrainDemand(armyRef)
	}
	return overload == 0 && demand <= capacity
}

func aiRegionHasAdjacentHostileArmy(gs *state.GameState, fid faction.FactionID, region *world.Region) bool {
	if gs == nil || region == nil {
		return false
	}
	adjacent := make(map[world.RegionID]struct{}, len(region.Neighbors))
	for _, neighborID := range region.Neighbors {
		adjacent[neighborID] = struct{}{}
	}
	return aiHostilePowerInRegions(gs, fid, adjacent) > 0
}

func aiRecoveryRegionScore(ctx *StrategicContext, armyRef *army.Army, region *world.Region) int {
	demand, capacity, _ := aiRegionLogistics(ctx.gs, region, armyRef.OwnerID)
	if region.ID != armyRef.RegionID {
		demand += ctx.gs.RegionalArmyGrainDemand(armyRef)
	}
	score := maxInt(0, capacity-demand) * 8
	for _, candidate := range aiSortedArmies(ctx.gs) {
		if candidate.ID != armyRef.ID && !candidate.IsNaval && candidate.OwnerID == armyRef.OwnerID && candidate.RegionID == region.ID {
			score += candidate.TotalStrength(ctx.gs.UnitTypes)
		}
	}
	if region.IsFortified() {
		score += 30
	}
	if ctx.gs.IsCapitalRegion(region) {
		score += 20
	}
	return score
}

func aiRetreatNextStep(ctx *StrategicContext, armyRef *army.Army) world.RegionID {
	if ctx == nil || ctx.gs == nil || armyRef == nil {
		return ""
	}
	assignment, ok := ctx.ArmyAssignments[armyRef.ID]
	if !ok || assignment.Role != AIArmyRoleRetreat || assignment.AnchorRegionID == "" || armyRef.RegionID == assignment.AnchorRegionID {
		return ""
	}
	return ctx.routeNextStep(armyRef, assignment.AnchorRegionID, aiRouteFriendly)
}

// executeStrategicSiegeWithdrawal kuşatan ordu retreat rolü aldığında hareket
// seçilmeden önce kuşatmayı aynı fraksiyondan kalan orduya devreder veya kaldırır.
func executeStrategicSiegeWithdrawal(gs *state.GameState, armyRef *army.Army, fid faction.FactionID, ctx *StrategicContext) (TurnStep, bool) {
	if gs == nil || armyRef == nil || ctx == nil {
		return TurnStep{}, false
	}
	assignment, ok := ctx.ArmyAssignments[armyRef.ID]
	if !ok || assignment.Role != AIArmyRoleRetreat {
		return TurnStep{}, false
	}
	siege := gs.SiegeByArmy(armyRef.ID)
	if siege == nil {
		return TurnStep{}, false
	}
	siegeRegionID := siege.RegionID
	transferred := aiTransferOrClearSiegeForWithdrawal(gs, siege, armyRef.ID)
	action := "kuşatmayı kaldırdı"
	if transferred {
		action = "kuşatma görevini devretti"
	}
	return TurnStep{
		FactionID:    fid,
		Kind:         TurnStepMove,
		ArmyID:       armyRef.ID,
		FromRegion:   armyRef.RegionID,
		TargetRegion: assignment.AnchorRegionID,
		FocusRegion:  siegeRegionID,
		Message:      turnFactionName(gs, fid) + " " + turnRegionName(gs, siegeRegionID) + " " + action + " ve takviye için geri çekiliyor.",
	}, true
}

func aiTransferOrClearSiegeForWithdrawal(gs *state.GameState, siege *state.SiegeState, leavingArmyID army.ArmyID) bool {
	if gs == nil || siege == nil {
		return false
	}
	attackerFactionID := siege.AttackerFactionID
	if attackerFactionID == "" {
		if leaving := gs.Armies[leavingArmyID]; leaving != nil {
			attackerFactionID = leaving.OwnerID
		}
	}
	for _, candidate := range aiSortedArmies(gs) {
		if candidate.ID == leavingArmyID || candidate.IsNaval || len(candidate.Units) == 0 ||
			candidate.OwnerID != attackerFactionID || candidate.RegionID != siege.RegionID {
			continue
		}
		siege.AttackerArmyID = candidate.ID
		return true
	}
	delete(gs.Sieges, siege.RegionID)
	return false
}
