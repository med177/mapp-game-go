package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

const (
	aiSecuritySatisfactionThreshold    = 35
	aiSecurityForeignReligionThreshold = 45
)

type aiSecurityTarget struct {
	region           *world.Region
	immediateRisk    bool
	religionMismatch bool
}

// applySecurityAssignments düşük memnuniyetli, surla veya sabit garnizonla
// korunmayan bölgeler için en küçük uygun mobil orduyu security rolüne alır.
// Bu state kalıcı değildir; her AI turunda güncel memnuniyetten yeniden türetilir.
func applySecurityAssignments(ctx *StrategicContext) {
	if ctx == nil || ctx.gs == nil || ctx.FactionID == "" {
		return
	}
	targets := aiSecurityTargets(ctx)
	if len(targets) == 0 {
		return
	}
	mobileCount := aiSecurityMobileArmyCount(ctx)
	for _, target := range targets {
		if aiSecurityRegionAlreadyCovered(ctx, target.region.ID) {
			continue
		}
		if mobileCount <= 1 && !target.immediateRisk {
			continue
		}
		candidate := selectSecurityArmy(ctx, target)
		if candidate == nil {
			continue
		}
		reason := "düşük memnuniyetli bölge güvenliği"
		if target.religionMismatch {
			reason = "din farkı bulunan bölge güvenliği"
		}
		if target.immediateRisk {
			reason = "acil isyan riski"
		}
		previous := ctx.ArmyAssignments[candidate.ID]
		if previous.Role == AIArmyRoleReserve {
			ctx.ReserveAssignedPower = maxInt(0, ctx.ReserveAssignedPower-candidate.TotalStrength(ctx.gs.UnitTypes))
		}
		ctx.ArmyAssignments[candidate.ID] = AIArmyAssignment{
			Role:           AIArmyRoleSecurity,
			AnchorRegionID: target.region.ID,
			Reason:         reason,
		}
	}
}

func aiSecurityTargets(ctx *StrategicContext) []aiSecurityTarget {
	owner := ctx.gs.Factions[ctx.FactionID]
	ownerReligion := ""
	if owner != nil {
		ownerReligion = string(owner.Religion)
	}
	targets := make([]aiSecurityTarget, 0)
	for _, region := range aiSortedRegions(ctx.gs) {
		if region.IsSea || region.OwnerID != string(ctx.FactionID) || aiRegionHasWalls(region) || ctx.gs.SiegeAt(region.ID) != nil {
			continue
		}
		mismatch := ownerReligion != "" && region.Religion != "" && region.Religion != ownerReligion
		threshold := aiSecuritySatisfactionThreshold
		if mismatch {
			threshold = aiSecurityForeignReligionThreshold
		}
		if region.Satisfaction >= threshold {
			continue
		}
		targets = append(targets, aiSecurityTarget{
			region:           region,
			immediateRisk:    region.IsRebellionRisk(),
			religionMismatch: mismatch,
		})
	}
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].immediateRisk != targets[j].immediateRisk {
			return targets[i].immediateRisk
		}
		if targets[i].region.Satisfaction != targets[j].region.Satisfaction {
			return targets[i].region.Satisfaction < targets[j].region.Satisfaction
		}
		if targets[i].religionMismatch != targets[j].religionMismatch {
			return targets[i].religionMismatch
		}
		iv := ctx.strategicRegionValue(targets[i].region)
		jv := ctx.strategicRegionValue(targets[j].region)
		if iv != jv {
			return iv > jv
		}
		return targets[i].region.ID < targets[j].region.ID
	})
	return targets
}

func aiRegionHasWalls(region *world.Region) bool {
	return region != nil && region.BuildingLevel("walls") > 0
}

func aiSecurityMobileArmyCount(ctx *StrategicContext) int {
	count := 0
	for _, armyRef := range aiSortedArmies(ctx.gs) {
		if armyRef.OwnerID == string(ctx.FactionID) && !armyRef.IsNaval && !armyRef.IsGarrison && len(armyRef.Units) > 0 {
			count++
		}
	}
	return count
}

func aiSecurityRegionAlreadyCovered(ctx *StrategicContext, regionID world.RegionID) bool {
	for _, armyRef := range aiSortedArmies(ctx.gs) {
		if armyRef.OwnerID != string(ctx.FactionID) || armyRef.IsNaval || armyRef.RegionID != regionID || len(armyRef.Units) == 0 {
			continue
		}
		if armyRef.IsGarrison || ctx.gs.SiegeByArmy(armyRef.ID) != nil {
			return true
		}
		assignment, assigned := ctx.ArmyAssignments[armyRef.ID]
		if !assigned || assignment.AnchorRegionID != regionID {
			continue
		}
		switch assignment.Role {
		case AIArmyRoleReserve, AIArmyRoleSecurity:
			return true
		}
	}
	return false
}

func selectSecurityArmy(ctx *StrategicContext, target aiSecurityTarget) *army.Army {
	type candidateScore struct {
		armyRef  *army.Army
		power    int
		distance int
	}
	var candidates []candidateScore
	for _, armyRef := range aiSortedArmies(ctx.gs) {
		if armyRef.OwnerID != string(ctx.FactionID) || armyRef.IsNaval || armyRef.IsGarrison || len(armyRef.Units) == 0 || ctx.gs.SiegeByArmy(armyRef.ID) != nil {
			continue
		}
		assignment := ctx.ArmyAssignments[armyRef.ID]
		if !aiSecurityCanOverrideAssignment(assignment) {
			continue
		}
		routes := ctx.routesFor(armyRef, armyRef.RegionID, aiRouteFriendly, 0)
		distance, reachable := routes.distance(target.region.ID)
		if !reachable {
			continue
		}
		hops, _ := routes.hopCount(target.region.ID)
		if target.immediateRisk && hops > maxInt(0, armyRef.MovePoints) {
			continue
		}
		candidates = append(candidates, candidateScore{
			armyRef:  armyRef,
			power:    armyRef.TotalStrength(ctx.gs.UnitTypes),
			distance: distance,
		})
	}
	if len(candidates) == 0 {
		return nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].power != candidates[j].power {
			return candidates[i].power < candidates[j].power
		}
		if candidates[i].distance != candidates[j].distance {
			return candidates[i].distance < candidates[j].distance
		}
		return candidates[i].armyRef.ID < candidates[j].armyRef.ID
	})
	return candidates[0].armyRef
}

func aiSecurityCanOverrideAssignment(assignment AIArmyAssignment) bool {
	switch assignment.Role {
	case AIArmyRoleSiege, AIArmyRoleRelief, AIArmyRoleRetreat, AIArmyRoleSecurity:
		return false
	case AIArmyRoleDefense:
		// Gerçek bir tehdit cephesine atanmış savunma ordusunu iç güvenlik için çekme.
		return assignment.FrontFactionID == ""
	case AIArmyRoleReserve:
		// Stratejik rezerv iç güvenlikte kullanılabilir; atanan güç sayacı çağıran
		// tarafından düşürülür ve yeni savaş hazırlığı eksik rezervi görür.
		return true
	default:
		return true
	}
}

func aiSecurityNextStep(ctx *StrategicContext, armyRef *army.Army) world.RegionID {
	if ctx == nil || ctx.gs == nil || armyRef == nil {
		return ""
	}
	assignment, ok := ctx.ArmyAssignments[armyRef.ID]
	if !ok || assignment.Role != AIArmyRoleSecurity || assignment.AnchorRegionID == "" || armyRef.RegionID == assignment.AnchorRegionID {
		return ""
	}
	return ctx.routeNextStep(armyRef, assignment.AnchorRegionID, aiRouteFriendly)
}
