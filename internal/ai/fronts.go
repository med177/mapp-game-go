package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// AIArmyRole bir AI ordusunun bu turdaki türetilmiş stratejik görevini belirtir.
// Roller save'e yazılmaz; her AI turunda güncel cephe ve plan state'inden üretilir.
type AIArmyRole string

const (
	AIArmyRoleAssault  AIArmyRole = "assault"
	AIArmyRoleSiege    AIArmyRole = "siege"
	AIArmyRoleDefense  AIArmyRole = "defense"
	AIArmyRoleReserve  AIArmyRole = "reserve"
	AIArmyRoleRelief   AIArmyRole = "relief"
	AIArmyRoleRetreat  AIArmyRole = "retreat"
	AIArmyRoleSecurity AIArmyRole = "security"
)

type AIFront struct {
	EnemyFactionID   faction.FactionID
	FriendlyRegions  []world.RegionID
	EnemyRegions     []world.RegionID
	AnchorRegionID   world.RegionID
	FriendlyPower    int
	EnemyPower       int
	ThreatScore      int
	AtWar            bool
	ObjectiveRelated bool
	CapitalThreat    bool
	CriticalThreat   bool
}

type AIArmyAssignment struct {
	Role           AIArmyRole
	AnchorRegionID world.RegionID
	FrontFactionID faction.FactionID
	Rallying       bool
	Reason         string
}

type aiFrontBuilder struct {
	front          AIFront
	friendlyRegion map[world.RegionID]struct{}
	enemyRegion    map[world.RegionID]struct{}
}

// prepareStrategicContext planı doğrular ve yalnız bu AI turunda kullanılacak
// cephe, rezerv ve ordu rolü snapshot'ını üretir.
func prepareStrategicContext(gs *state.GameState, fid faction.FactionID) *StrategicContext {
	ctx := buildStrategicContext(gs, fid)
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" {
		return ctx
	}
	ensureStrategicPlan(gs, fid, ctx)
	buildAIFronts(ctx)
	ensurePlanRallyState(ctx)
	assignAIArmyRoles(ctx)
	applySecurityAssignments(ctx)
	applyRetreatAssignments(ctx)
	applyRallyAssignments(ctx)
	return ctx
}

func buildAIFronts(ctx *StrategicContext) {
	if ctx == nil || ctx.gs == nil || ctx.FactionID == "" {
		return
	}
	gs := ctx.gs
	plan := gs.AIPlans[ctx.FactionID]
	capitalRegionID := world.RegionID("")
	if capitalRegion, _, _, ok := gs.FactionCapital(ctx.FactionID); ok && capitalRegion != nil {
		capitalRegionID = capitalRegion.ID
	}

	criticalRegions := make(map[world.RegionID]struct{})
	if capitalRegionID != "" {
		criticalRegions[capitalRegionID] = struct{}{}
	}
	if plan != nil && (plan.Kind == state.AIObjectiveDefend || plan.Kind == state.AIObjectiveConsolidate) {
		for _, regionID := range plan.TargetRegionIDs {
			if region := gs.Regions[regionID]; region != nil && region.OwnerID == string(ctx.FactionID) {
				criticalRegions[regionID] = struct{}{}
			}
		}
	}

	builders := make(map[faction.FactionID]*aiFrontBuilder)
	for _, region := range aiSortedRegions(gs) {
		if region.IsSea || region.OwnerID != string(ctx.FactionID) {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil || neighbor.IsSea || neighbor.OwnerID == "" || neighbor.OwnerID == string(ctx.FactionID) {
				continue
			}
			enemyID := faction.FactionID(neighbor.OwnerID)
			if diplomacy.SameRealm(gs, ctx.FactionID, enemyID) {
				continue
			}
			rel := diplomacy.Relation(gs, ctx.FactionID, enemyID)
			if rel != nil && rel.Stance == faction.StanceAllied {
				continue
			}
			builder := builders[enemyID]
			if builder == nil {
				builder = &aiFrontBuilder{
					front:          AIFront{EnemyFactionID: enemyID},
					friendlyRegion: make(map[world.RegionID]struct{}),
					enemyRegion:    make(map[world.RegionID]struct{}),
				}
				builders[enemyID] = builder
			}
			builder.friendlyRegion[region.ID] = struct{}{}
			builder.enemyRegion[neighbor.ID] = struct{}{}
		}
	}

	for enemyID, builder := range builders {
		rel := diplomacy.Relation(gs, ctx.FactionID, enemyID)
		builder.front.AtWar = rel != nil && rel.Stance == faction.StanceWar
		builder.front.ObjectiveRelated = plan != nil && plan.TargetFactionID == enemyID
		builder.front.FriendlyRegions = sortedRegionSet(builder.friendlyRegion)
		builder.front.EnemyRegions = sortedRegionSet(builder.enemyRegion)

		for _, armyRef := range aiSortedArmies(gs) {
			if armyRef.IsNaval {
				continue
			}
			if armyRef.OwnerID == string(ctx.FactionID) {
				if _, ok := builder.friendlyRegion[armyRef.RegionID]; ok {
					builder.front.FriendlyPower += armyRef.TotalStrength(gs.UnitTypes)
				}
				continue
			}
			if armyRef.OwnerID == string(enemyID) {
				if _, ok := builder.enemyRegion[armyRef.RegionID]; ok {
					builder.front.EnemyPower += armyRef.TotalStrength(gs.UnitTypes)
				}
			}
		}

		bestAnchorValue := -1
		for _, regionID := range builder.front.FriendlyRegions {
			region := gs.Regions[regionID]
			value := ctx.strategicRegionValue(region)
			if _, critical := criticalRegions[regionID]; critical {
				value += 10000
			}
			if value > bestAnchorValue || (value == bestAnchorValue && (builder.front.AnchorRegionID == "" || regionID < builder.front.AnchorRegionID)) {
				bestAnchorValue = value
				builder.front.AnchorRegionID = regionID
			}
			if regionID == capitalRegionID && builder.front.AtWar {
				builder.front.CapitalThreat = true
			}
			if _, critical := criticalRegions[regionID]; critical && builder.front.AtWar {
				builder.front.CriticalThreat = true
			}
			if siege := gs.SiegeAt(regionID); siege != nil {
				if siegeArmy := gs.Armies[siege.AttackerArmyID]; siegeArmy != nil && siegeArmy.OwnerID != string(ctx.FactionID) {
					builder.front.CriticalThreat = true
					if regionID == capitalRegionID {
						builder.front.CapitalThreat = true
					}
				}
			}
		}

		builder.front.ThreatScore = builder.front.EnemyPower - builder.front.FriendlyPower
		if builder.front.AtWar {
			builder.front.ThreatScore += 40
		}
		if builder.front.ObjectiveRelated {
			builder.front.ThreatScore += 20
		}
		if builder.front.CriticalThreat {
			builder.front.ThreatScore += 80
		}
		if builder.front.CapitalThreat {
			builder.front.ThreatScore += 80
		}
		ctx.Fronts = append(ctx.Fronts, builder.front)
	}

	sort.Slice(ctx.Fronts, func(i, j int) bool {
		if ctx.Fronts[i].ThreatScore != ctx.Fronts[j].ThreatScore {
			return ctx.Fronts[i].ThreatScore > ctx.Fronts[j].ThreatScore
		}
		return ctx.Fronts[i].EnemyFactionID < ctx.Fronts[j].EnemyFactionID
	})
	for _, front := range ctx.Fronts {
		if front.CriticalThreat || front.CapitalThreat {
			ctx.CriticalThreat = true
		}
	}
}

func sortedRegionSet(values map[world.RegionID]struct{}) []world.RegionID {
	result := make([]world.RegionID, 0, len(values))
	for regionID := range values {
		result = append(result, regionID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func assignAIArmyRoles(ctx *StrategicContext) {
	if ctx == nil || ctx.gs == nil || ctx.FactionID == "" {
		return
	}
	gs := ctx.gs
	ctx.ArmyAssignments = make(map[army.ArmyID]AIArmyAssignment)

	var mobile []*army.Army
	var strongest *army.Army
	for _, armyRef := range aiSortedArmies(gs) {
		if armyRef.OwnerID != string(ctx.FactionID) || armyRef.IsNaval {
			continue
		}
		if armyRef.IsGarrison {
			ctx.ArmyAssignments[armyRef.ID] = AIArmyAssignment{Role: AIArmyRoleDefense, AnchorRegionID: armyRef.RegionID, Reason: "sabit garnizon"}
			continue
		}
		power := armyRef.TotalStrength(gs.UnitTypes)
		ctx.TotalMobilePower += power
		mobile = append(mobile, armyRef)
		if strongest == nil || power > strongest.TotalStrength(gs.UnitTypes) || (power == strongest.TotalStrength(gs.UnitTypes) && armyRef.ID < strongest.ID) {
			strongest = armyRef
		}
	}
	if len(mobile) == 0 {
		return
	}

	for _, armyRef := range mobile {
		if siege := gs.SiegeByArmy(armyRef.ID); siege != nil {
			ctx.ArmyAssignments[armyRef.ID] = AIArmyAssignment{Role: AIArmyRoleSiege, AnchorRegionID: siege.RegionID, Reason: "aktif kuşatma"}
		}
	}

	for _, regionID := range aiFriendlyReliefTargets(ctx) {
		candidate := nearestUnassignedArmy(ctx, mobile, regionID, strongest, false)
		if candidate == nil {
			continue
		}
		ctx.ArmyAssignments[candidate.ID] = AIArmyAssignment{Role: AIArmyRoleRelief, AnchorRegionID: regionID, Reason: "dost kuşatmasını kaldır"}
	}

	reserveAnchor := aiReserveAnchor(ctx)
	ctx.ReservePercent = 15
	if ctx.CriticalThreat {
		ctx.ReservePercent = 30
	}
	ctx.ReserveTargetPower = (ctx.TotalMobilePower*ctx.ReservePercent + 99) / 100
	if len(mobile) <= 1 {
		ctx.ReserveTargetPower = 0
	} else if strongest != nil {
		maxReservable := ctx.TotalMobilePower - strongest.TotalStrength(gs.UnitTypes)
		ctx.ReserveTargetPower = minInt(ctx.ReserveTargetPower, maxInt(0, maxReservable))
	}
	if reserveAnchor != "" && ctx.ReserveTargetPower > 0 {
		candidates := append([]*army.Army(nil), mobile...)
		sort.Slice(candidates, func(i, j int) bool {
			di := ctx.routeDistance(candidates[i], reserveAnchor, aiRouteGeneral)
			dj := ctx.routeDistance(candidates[j], reserveAnchor, aiRouteGeneral)
			if di != dj {
				return normalizedDistance(di) < normalizedDistance(dj)
			}
			pi := candidates[i].TotalStrength(gs.UnitTypes)
			pj := candidates[j].TotalStrength(gs.UnitTypes)
			if pi != pj {
				return pi < pj
			}
			return candidates[i].ID < candidates[j].ID
		})
		for _, candidate := range candidates {
			if ctx.ReserveAssignedPower >= ctx.ReserveTargetPower {
				break
			}
			if _, assigned := ctx.ArmyAssignments[candidate.ID]; assigned || (strongest != nil && candidate.ID == strongest.ID) {
				continue
			}
			ctx.ArmyAssignments[candidate.ID] = AIArmyAssignment{Role: AIArmyRoleReserve, AnchorRegionID: reserveAnchor, Reason: "dinamik stratejik rezerv"}
			ctx.ReserveAssignedPower += candidate.TotalStrength(gs.UnitTypes)
		}
	}

	plan := gs.AIPlans[ctx.FactionID]
	offensiveAnchor, offensiveTarget := aiOffensiveAnchor(ctx, plan)
	for _, front := range ctx.Fronts {
		if !front.AtWar || front.ThreatScore <= 0 || front.AnchorRegionID == "" {
			continue
		}
		if (front.ObjectiveRelated || front.EnemyFactionID == offensiveTarget) && !front.CriticalThreat {
			continue
		}
		candidate := nearestUnassignedArmy(ctx, mobile, front.AnchorRegionID, strongest, true)
		if candidate == nil {
			continue
		}
		ctx.ArmyAssignments[candidate.ID] = AIArmyAssignment{Role: AIArmyRoleDefense, AnchorRegionID: front.AnchorRegionID, FrontFactionID: front.EnemyFactionID, Reason: "tehdit altındaki cephe"}
	}

	for _, armyRef := range mobile {
		if _, assigned := ctx.ArmyAssignments[armyRef.ID]; assigned {
			continue
		}
		assignment := AIArmyAssignment{Role: AIArmyRoleDefense, AnchorRegionID: aiDefenseAnchor(ctx), Reason: "bölgesel savunma"}
		if plan != nil && plan.Kind == state.AIObjectiveExpand {
			assignment.Role = AIArmyRoleAssault
			assignment.AnchorRegionID = offensiveAnchor
			assignment.FrontFactionID = offensiveTarget
			assignment.Reason = "aktif objective hücumu"
			if offensiveTarget != "" && offensiveTarget != plan.TargetFactionID {
				assignment.Reason = "aktif savaş cephesini sonuçlandır"
			}
			if armyRef.HasSiegeUnits(gs.UnitTypes) {
				assignment.Role = AIArmyRoleSiege
				assignment.Reason = "objective kuşatma gücü"
			}
		}
		if assignment.AnchorRegionID == "" {
			assignment.AnchorRegionID = armyRef.RegionID
		}
		ctx.ArmyAssignments[armyRef.ID] = assignment
	}
}

func aiOffensiveAnchor(ctx *StrategicContext, plan *state.AIPlanState) (world.RegionID, faction.FactionID) {
	if ctx == nil || ctx.gs == nil || plan == nil {
		return "", ""
	}
	for _, front := range ctx.Fronts {
		if front.AtWar && front.EnemyFactionID == plan.TargetFactionID {
			return firstOwnedRegion(ctx.gs, plan.TargetRegionIDs, plan.TargetFactionID), plan.TargetFactionID
		}
	}
	// Kalıcı objective başka bir devleti gösterse bile ilan edilmiş savaşı sahipsiz
	// bırakma. Plan yeniden değerlendirildiğinde kalıcı hedef de bu cepheyle hizalanır.
	for _, front := range ctx.Fronts {
		if front.AtWar && len(front.EnemyRegions) > 0 {
			return front.EnemyRegions[0], front.EnemyFactionID
		}
	}
	return firstOwnedRegion(ctx.gs, plan.TargetRegionIDs, plan.TargetFactionID), plan.TargetFactionID
}

func aiFriendlyReliefTargets(ctx *StrategicContext) []world.RegionID {
	if ctx == nil || ctx.gs == nil {
		return nil
	}
	var targets []world.RegionID
	for regionID, siege := range ctx.gs.Sieges {
		region := ctx.gs.Regions[regionID]
		if siege == nil || region == nil || region.IsSea || !diplomacy.SameRealm(ctx.gs, ctx.FactionID, faction.FactionID(region.OwnerID)) {
			continue
		}
		siegeArmy := ctx.gs.Armies[siege.AttackerArmyID]
		if siegeArmy == nil || diplomacy.SameRealm(ctx.gs, ctx.FactionID, faction.FactionID(siegeArmy.OwnerID)) {
			continue
		}
		targets = append(targets, regionID)
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i] < targets[j] })
	return targets
}

func aiReserveAnchor(ctx *StrategicContext) world.RegionID {
	if ctx == nil || ctx.gs == nil {
		return ""
	}
	for _, front := range ctx.Fronts {
		if front.CapitalThreat && front.AnchorRegionID != "" {
			return front.AnchorRegionID
		}
	}
	for _, front := range ctx.Fronts {
		if front.CriticalThreat && front.AnchorRegionID != "" {
			return front.AnchorRegionID
		}
	}
	if capital, _, _, ok := ctx.gs.FactionCapital(ctx.FactionID); ok && capital != nil {
		return capital.ID
	}
	return aiDefenseAnchor(ctx)
}

func aiDefenseAnchor(ctx *StrategicContext) world.RegionID {
	if ctx == nil {
		return ""
	}
	for _, front := range ctx.Fronts {
		if front.AtWar && front.AnchorRegionID != "" {
			return front.AnchorRegionID
		}
	}
	if len(ctx.BorderRegionIDs) > 0 {
		return ctx.BorderRegionIDs[0]
	}
	if len(ctx.OwnedLandRegionIDs) > 0 {
		return ctx.OwnedLandRegionIDs[0]
	}
	return ""
}

func firstOwnedRegion(gs *state.GameState, regionIDs []world.RegionID, ownerID faction.FactionID) world.RegionID {
	for _, regionID := range regionIDs {
		if gs != nil && gs.Regions[regionID] != nil && gs.Regions[regionID].OwnerID == string(ownerID) {
			return regionID
		}
	}
	return ""
}

func nearestUnassignedArmy(ctx *StrategicContext, armies []*army.Army, target world.RegionID, strongest *army.Army, allowStrongest bool) *army.Army {
	var best *army.Army
	bestDistance := int(^uint(0) >> 1)
	for _, candidate := range armies {
		if candidate == nil {
			continue
		}
		if _, assigned := ctx.ArmyAssignments[candidate.ID]; assigned {
			continue
		}
		if !allowStrongest && strongest != nil && candidate.ID == strongest.ID {
			continue
		}
		distance := normalizedDistance(ctx.routeDistance(candidate, target, aiRouteGeneral))
		if best == nil || distance < bestDistance || (distance == bestDistance && candidate.ID < best.ID) {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}

func normalizedDistance(distance int) int {
	if distance < 0 {
		return int(^uint(0) >> 2)
	}
	return distance
}

func aiRoleAdjustedMoveScore(ctx *StrategicContext, armyRef *army.Army, target *world.Region, baseScore int) int {
	if ctx == nil || armyRef == nil || target == nil || ctx.gs == nil {
		return baseScore
	}
	assignment, ok := ctx.ArmyAssignments[armyRef.ID]
	if !ok || assignment.AnchorRegionID == "" {
		return baseScore
	}
	// Rol bonusu standart geçiş/savaş kurallarının reddettiği bir hedefi
	// yeniden geçerli kılamaz; yalnız zaten yasal olan hamleleri sıralar.
	if baseScore < 0 {
		return baseScore
	}
	reducesDistance := ctx.routeNextStep(armyRef, assignment.AnchorRegionID, aiRouteGeneral) == target.ID
	if assignment.Rallying {
		if armyRef.RegionID == assignment.AnchorRegionID || target.OwnerID != armyRef.OwnerID || !reducesDistance {
			return -1
		}
		return maxInt(1, baseScore) + 160
	}

	switch assignment.Role {
	case AIArmyRoleRetreat:
		// Retreat rotası chooseBestMoveWithStrategicContext içinde yalnız güvenli
		// dost transit bölgelerinden kurulur. Bu kol fallback puanlamasının orduyu
		// yeniden düşman toprağına yöneltmesini engeller.
		if armyRef.RegionID == assignment.AnchorRegionID || target.OwnerID != armyRef.OwnerID || !reducesDistance {
			return -1
		}
		return maxInt(1, baseScore) + 200
	case AIArmyRoleSecurity:
		if armyRef.RegionID == assignment.AnchorRegionID || target.OwnerID != armyRef.OwnerID || !reducesDistance {
			return -1
		}
		return maxInt(1, baseScore) + 170
	case AIArmyRoleReserve:
		if target.OwnerID != armyRef.OwnerID || !reducesDistance {
			return -1
		}
		return maxInt(1, baseScore) + 120
	case AIArmyRoleDefense:
		if target.OwnerID != "" && target.OwnerID != armyRef.OwnerID && !diplomacy.SameRealm(ctx.gs, ctx.FactionID, faction.FactionID(target.OwnerID)) {
			return -1
		}
		if reducesDistance {
			return maxInt(1, baseScore) + 90
		}
	case AIArmyRoleRelief:
		if reducesDistance || target.ID == assignment.AnchorRegionID {
			return maxInt(1, baseScore) + 140
		}
	case AIArmyRoleSiege:
		if reducesDistance || target.ID == assignment.AnchorRegionID {
			return maxInt(1, baseScore) + 60
		}
	case AIArmyRoleAssault:
		if reducesDistance || target.ID == assignment.AnchorRegionID {
			return maxInt(1, baseScore) + 40
		}
	}
	return baseScore
}

func aiStrategicWarReady(ctx *StrategicContext, target faction.FactionID) bool {
	if ctx == nil || ctx.gs == nil || ctx.gs.ScenarioID != "1300_ottoman_rise" {
		return true
	}
	if ctx.CriticalThreat || ctx.ReserveAssignedPower < ctx.ReserveTargetPower {
		return false
	}
	if ctx.RallyActive {
		return false
	}
	attackPower := 0
	for armyID, assignment := range ctx.ArmyAssignments {
		if assignment.Role != AIArmyRoleAssault && assignment.Role != AIArmyRoleSiege {
			continue
		}
		armyRef := ctx.gs.Armies[armyID]
		if armyRef == nil || armyRef.IsNaval || armyRef.IsGarrison {
			continue
		}
		attackPower += armyRef.TotalStrength(ctx.gs.UnitTypes)
	}
	if attackPower <= 0 {
		return false
	}
	availablePower := maxInt(1, ctx.TotalMobilePower-ctx.ReserveAssignedPower)
	if attackPower*100 < availablePower*60 {
		return false
	}
	if target != "" {
		targetFrontierPower := aiFrontierPower(ctx.gs, target, ctx.FactionID)
		if targetFrontierPower > 0 && attackPower*100 < targetFrontierPower*aiMinAttackPowerPercent(ctx.gs) {
			return false
		}
	}
	return true
}
