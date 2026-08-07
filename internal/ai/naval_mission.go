package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

type aiNavalMissionKind string

const (
	aiNavalMissionAssault aiNavalMissionKind = "overseas_assault"
	aiNavalMissionRelief  aiNavalMissionKind = "overseas_relief"
)

// aiNavalMission save'e yazılmaz. Aktif stratejik plan, kara orduları, filolar
// ve üretim kuyruğundan her AI turunda deterministik olarak yeniden türetilir.
type aiNavalMission struct {
	Kind               aiNavalMissionKind
	TargetRegionID     world.RegionID
	TargetFactionID    faction.FactionID
	EmbarkArmyID       army.ArmyID
	FleetID            army.ArmyID
	EmbarkRegionID     world.RegionID
	EmbarkSeaRegionID  world.RegionID
	LandingSeaRegionID world.RegionID
	RequiredCapacity   int
	AvailableCapacity  int
	MissingCapacity    int
	RouteThreatPower   int
	Score              int
}

type aiNavalMissionTarget struct {
	region *world.Region
	kind   aiNavalMissionKind
	order  int
}

type aiNavalMissionCandidate struct {
	mission      aiNavalMission
	targetOrder  int
	landDistance int
	seaDistance  int
}

func buildAINavalMission(ctx *StrategicContext) *aiNavalMission {
	if ctx == nil || ctx.gs == nil || ctx.gs.ScenarioID != "1300_ottoman_rise" || ctx.FactionID == "" {
		return nil
	}
	targets := aiNavalMissionTargets(ctx)
	if len(targets) == 0 {
		return nil
	}

	// Taşınmış bir kuvvet, plan yeniden türetildiğinde görevini kaybetmemeli.
	// Bu adaylar kara ordusu adaylarından önce değerlendirilir.
	var bestLoaded *aiNavalMissionCandidate
	for _, target := range targets {
		for _, fleet := range aiSortedArmies(ctx.gs) {
			if fleet.OwnerID != string(ctx.FactionID) || !fleet.IsNaval || len(fleet.EmbarkedUnits) == 0 {
				continue
			}
			landingSea, seaDistance := aiBestLandingSeaForContext(ctx, fleet.RegionID, target.region)
			if landingSea == "" {
				continue
			}
			route := ctx.navalSeaRoute(fleet.RegionID, landingSea)
			candidate := aiNavalMissionCandidate{
				mission: aiNavalMission{
					Kind:               target.kind,
					TargetRegionID:     target.region.ID,
					TargetFactionID:    aiNavalMissionTargetFaction(ctx, target),
					FleetID:            fleet.ID,
					LandingSeaRegionID: landingSea,
					RequiredCapacity:   len(fleet.EmbarkedUnits),
					AvailableCapacity:  len(fleet.EmbarkedUnits),
					RouteThreatPower:   route.MaxThreat,
					Score:              10000 + len(fleet.EmbarkedUnits)*20 - target.order*200 - seaDistance*5,
				},
				targetOrder: target.order,
				seaDistance: seaDistance,
			}
			if aiNavalMissionCandidateBetter(candidate, bestLoaded) {
				copy := candidate
				bestLoaded = &copy
			}
		}
	}
	if bestLoaded != nil {
		return &bestLoaded.mission
	}

	var best *aiNavalMissionCandidate
	for _, target := range targets {
		if aiAnyEligibleArmyCanReachTarget(ctx, target.region, target.kind) {
			continue
		}
		for _, landArmy := range aiSortedArmies(ctx.gs) {
			if !aiNavalMissionArmyEligible(ctx, landArmy, target.kind) {
				continue
			}
			if distance := aiNavalTargetLandDistance(ctx, landArmy, target.region, target.kind); distance >= 0 {
				continue
			}
			embarkRegion, embarkSea, landingSea, landDistance, seaDistance := aiBestEmbarkRoute(ctx, landArmy, target.region)
			if embarkRegion == "" || embarkSea == "" || landingSea == "" {
				continue
			}
			available := aiReachableTransportCapacity(ctx.gs, ctx.FactionID, embarkSea)
			required := len(landArmy.Units)
			missing := maxInt(0, required-available)
			route := ctx.navalSeaRoute(embarkSea, landingSea)
			score := len(landArmy.Units)*20 - target.order*200 - landDistance*2 - seaDistance*5 - route.MaxThreat*8
			if target.kind == aiNavalMissionRelief {
				score += 80
			}
			if aiRegionHasPortBuilding(ctx.gs.Regions[embarkRegion]) {
				score += 30
			}
			candidate := aiNavalMissionCandidate{
				mission: aiNavalMission{
					Kind:               target.kind,
					TargetRegionID:     target.region.ID,
					TargetFactionID:    aiNavalMissionTargetFaction(ctx, target),
					EmbarkArmyID:       landArmy.ID,
					EmbarkRegionID:     embarkRegion,
					EmbarkSeaRegionID:  embarkSea,
					LandingSeaRegionID: landingSea,
					RequiredCapacity:   required,
					AvailableCapacity:  available,
					MissingCapacity:    missing,
					RouteThreatPower:   route.MaxThreat,
					Score:              score,
				},
				targetOrder:  target.order,
				landDistance: landDistance,
				seaDistance:  seaDistance,
			}
			if aiNavalMissionCandidateBetter(candidate, best) {
				copy := candidate
				best = &copy
			}
		}
	}
	if best == nil {
		return nil
	}
	return &best.mission
}

func aiNavalMissionTargetFaction(ctx *StrategicContext, target aiNavalMissionTarget) faction.FactionID {
	if target.kind == aiNavalMissionRelief {
		if plan := ctx.gs.AIPlans[ctx.FactionID]; plan != nil {
			return plan.TargetFactionID
		}
	}
	return faction.FactionID(target.region.OwnerID)
}

func aiNavalMissionTargets(ctx *StrategicContext) []aiNavalMissionTarget {
	plan := ctx.gs.AIPlans[ctx.FactionID]
	if plan == nil || plan.Kind == state.AIObjectiveConsolidate {
		return nil
	}
	kind := aiNavalMissionAssault
	wantedOwner := string(plan.TargetFactionID)
	if plan.Kind == state.AIObjectiveDefend {
		kind = aiNavalMissionRelief
		wantedOwner = string(ctx.FactionID)
	}
	if wantedOwner == "" {
		return nil
	}

	seen := make(map[world.RegionID]struct{})
	result := make([]aiNavalMissionTarget, 0, len(plan.TargetRegionIDs))
	appendTarget := func(regionID world.RegionID) {
		region := ctx.gs.Regions[regionID]
		if region == nil || region.IsSea || region.OwnerID != wantedOwner || !region.IsCoastal(ctx.gs.Regions) {
			return
		}
		if _, exists := seen[regionID]; exists {
			return
		}
		seen[regionID] = struct{}{}
		result = append(result, aiNavalMissionTarget{region: region, kind: kind, order: len(result)})
	}
	for _, regionID := range plan.TargetRegionIDs {
		appendTarget(regionID)
	}
	// Genişleme objective'i kara hedefleriyle başlamış olsa bile aynı hedef
	// devletin ulaşılabilir kıyısı meşru bir çıkarma kapısıdır.
	if kind == aiNavalMissionAssault {
		for _, region := range aiSortedRegions(ctx.gs) {
			if region.OwnerID == wantedOwner {
				appendTarget(region.ID)
			}
		}
	}
	return result
}

func aiNavalMissionArmyEligible(ctx *StrategicContext, candidate *army.Army, kind aiNavalMissionKind) bool {
	if ctx == nil || candidate == nil || candidate.OwnerID != string(ctx.FactionID) || candidate.IsNaval || candidate.IsGarrison || len(candidate.Units) == 0 || !candidate.CanEmbark(ctx.gs.UnitTypes) {
		return false
	}
	if ctx.gs.SiegeByArmy(candidate.ID) != nil || ctx.gs.IsArmyDefendingSiegedRegion(candidate) {
		return false
	}
	assignment, assigned := strategicContextAssignment(ctx, candidate.ID)
	if !assigned {
		return true
	}
	if assignment.Role == AIArmyRoleRetreat || assignment.Role == AIArmyRoleSecurity || assignment.Role == AIArmyRoleReserve {
		return false
	}
	if kind == aiNavalMissionAssault && (assignment.Role == AIArmyRoleDefense || assignment.Role == AIArmyRoleRelief) {
		return false
	}
	return true
}

func aiAnyEligibleArmyCanReachTarget(ctx *StrategicContext, target *world.Region, kind aiNavalMissionKind) bool {
	for _, candidate := range aiSortedArmies(ctx.gs) {
		if !aiNavalMissionArmyEligible(ctx, candidate, kind) {
			continue
		}
		if aiNavalTargetLandDistance(ctx, candidate, target, kind) >= 0 {
			return true
		}
	}
	return false
}

func aiNavalTargetLandDistance(ctx *StrategicContext, candidate *army.Army, target *world.Region, kind aiNavalMissionKind) int {
	if ctx == nil || candidate == nil || target == nil {
		return -1
	}
	mode := aiRouteGeneral
	if kind == aiNavalMissionRelief {
		mode = aiRouteFriendly
	}
	return ctx.routeDistance(candidate, target.ID, mode)
}

func aiBestEmbarkRoute(ctx *StrategicContext, landArmy *army.Army, target *world.Region) (world.RegionID, world.RegionID, world.RegionID, int, int) {
	bestRegion := world.RegionID("")
	bestSea := world.RegionID("")
	bestLanding := world.RegionID("")
	bestLandDistance := 0
	bestSeaDistance := 0
	bestScore := int(^uint(0) >> 1)
	for _, coast := range aiSortedRegions(ctx.gs) {
		if coast.IsSea || coast.OwnerID != string(ctx.FactionID) || !coast.IsCoastal(ctx.gs.Regions) || !aiNavalEmbarkPortViable(ctx.gs, ctx.FactionID, coast) {
			continue
		}
		landDistance := 0
		if coast.ID != landArmy.RegionID {
			var reachable bool
			landDistance, reachable = ctx.routesFor(landArmy, landArmy.RegionID, aiRouteFriendly, 0).distance(coast.ID)
			if !reachable {
				continue
			}
		}
		embarkSea := aiSeaNeighbor(ctx.gs, coast)
		if embarkSea == "" {
			continue
		}
		landingSea, seaDistance := aiBestLandingSeaForContext(ctx, embarkSea, target)
		if landingSea == "" {
			continue
		}
		route := ctx.navalSeaRoute(embarkSea, landingSea)
		score := landDistance*2 + seaDistance*5 + route.MaxThreat*8 + route.TotalThreat*2
		if aiRegionHasPortBuilding(coast) {
			score -= 30
		} else if aiQueuedBuildingCount(ctx.gs, coast.ID, "port", ctx.FactionID) > 0 {
			score -= 15
		}
		if score < bestScore || (score == bestScore && (bestRegion == "" || coast.ID < bestRegion)) {
			bestScore = score
			bestRegion = coast.ID
			bestSea = embarkSea
			bestLanding = landingSea
			bestLandDistance = landDistance
			bestSeaDistance = seaDistance
		}
	}
	return bestRegion, bestSea, bestLanding, bestLandDistance, bestSeaDistance
}

func aiNavalEmbarkPortViable(gs *state.GameState, fid faction.FactionID, region *world.Region) bool {
	if gs == nil || region == nil || aiSeaNeighbor(gs, region) == "" {
		return false
	}
	if aiRegionHasPortBuilding(region) || aiQueuedBuildingCount(gs, region.ID, "port", fid) > 0 {
		return true
	}
	portType := gs.BuildingTypes["port"]
	if portType == nil || !aiBuildingAllowed(gs, region, "port", portType.RequiredTerrain) {
		return false
	}
	return aiBuildingLevel(region, "port") < portType.MaxPerRegion
}

func aiBestLandingSeaForContext(ctx *StrategicContext, start world.RegionID, target *world.Region) (world.RegionID, int) {
	if ctx == nil {
		return "", -1
	}
	return aiBestLandingSeaWithThreats(ctx.gs, start, target, ctx.navalThreatPower)
}

func aiBestLandingSeaWithThreats(gs *state.GameState, start world.RegionID, target *world.Region, hostileBySea map[world.RegionID]int) (world.RegionID, int) {
	if gs == nil || target == nil || start == "" {
		return "", -1
	}
	best := world.RegionID("")
	bestRoute := aiSeaRouteResult{}
	for _, neighborID := range aiSortedNeighborIDs(target) {
		neighbor := gs.Regions[neighborID]
		if neighbor == nil || !neighbor.IsSea {
			continue
		}
		route := aiThreatAwareSeaRouteWithThreats(gs, start, neighborID, hostileBySea)
		if !route.Reachable {
			continue
		}
		if best == "" || aiSeaRouteResultBetter(route, bestRoute) || (aiSeaRouteResultEqual(route, bestRoute) && neighborID < best) {
			best = neighborID
			bestRoute = route
		}
	}
	if best == "" {
		return "", -1
	}
	return best, bestRoute.Hops
}

func aiSeaRouteResultBetter(candidate, old aiSeaRouteResult) bool {
	if candidate.MaxThreat != old.MaxThreat {
		return candidate.MaxThreat < old.MaxThreat
	}
	if candidate.TotalThreat != old.TotalThreat {
		return candidate.TotalThreat < old.TotalThreat
	}
	if candidate.Hops != old.Hops {
		return candidate.Hops < old.Hops
	}
	return candidate.FirstStep < old.FirstStep
}

func aiSeaRouteResultEqual(left, right aiSeaRouteResult) bool {
	return left.Reachable == right.Reachable && left.FirstStep == right.FirstStep && left.Hops == right.Hops && left.MaxThreat == right.MaxThreat && left.TotalThreat == right.TotalThreat
}

func aiReachableTransportCapacity(gs *state.GameState, fid faction.FactionID, embarkSea world.RegionID) int {
	if gs == nil || embarkSea == "" {
		return 0
	}
	capacity := 0
	for _, fleet := range aiSortedArmies(gs) {
		if fleet.OwnerID != string(fid) || !fleet.IsNaval || aiSeaRouteDistance(gs, fleet.RegionID, embarkSea) < 0 {
			continue
		}
		capacity += fleet.AvailableTransportCapacity(gs.UnitTypes)
	}
	for _, order := range gs.ProductionQueue {
		if order.Kind != aiProductionKindUnit || order.FactionID != string(fid) {
			continue
		}
		unitType := gs.UnitTypes[order.TypeID]
		if unitType == nil || unitType.Category != army.CategoryNavalTrans || unitType.CarryCapacity <= 0 {
			continue
		}
		port := gs.Regions[order.RegionID]
		sea := aiSeaNeighbor(gs, port)
		if sea == "" || aiSeaRouteDistance(gs, sea, embarkSea) < 0 {
			continue
		}
		capacity += unitType.CarryCapacity
	}
	return minInt(army.MaxArmySize, capacity)
}

func aiNavalMissionCandidateBetter(candidate aiNavalMissionCandidate, best *aiNavalMissionCandidate) bool {
	if best == nil || candidate.mission.Score != best.mission.Score {
		return best == nil || candidate.mission.Score > best.mission.Score
	}
	if candidate.targetOrder != best.targetOrder {
		return candidate.targetOrder < best.targetOrder
	}
	if candidate.landDistance != best.landDistance {
		return candidate.landDistance < best.landDistance
	}
	if candidate.seaDistance != best.seaDistance {
		return candidate.seaDistance < best.seaDistance
	}
	if candidate.mission.TargetRegionID != best.mission.TargetRegionID {
		return candidate.mission.TargetRegionID < best.mission.TargetRegionID
	}
	if candidate.mission.EmbarkArmyID != best.mission.EmbarkArmyID {
		return candidate.mission.EmbarkArmyID < best.mission.EmbarkArmyID
	}
	return candidate.mission.FleetID < best.mission.FleetID
}

func aiSeaRouteDistance(gs *state.GameState, start, target world.RegionID) int {
	if start == "" || target == "" || gs == nil || gs.Regions[start] == nil || gs.Regions[target] == nil || !gs.Regions[start].IsSea || !gs.Regions[target].IsSea {
		return -1
	}
	if start == target {
		return 0
	}
	distance := map[world.RegionID]int{start: 0}
	queue := []world.RegionID{start}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, neighborID := range aiSortedNeighborIDs(gs.Regions[current]) {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil || !neighbor.IsSea {
				continue
			}
			if _, seen := distance[neighborID]; seen {
				continue
			}
			distance[neighborID] = distance[current] + 1
			if neighborID == target {
				return distance[neighborID]
			}
			queue = append(queue, neighborID)
		}
	}
	return -1
}

func aiSortedNeighborIDs(region *world.Region) []world.RegionID {
	if region == nil || len(region.Neighbors) == 0 {
		return nil
	}
	result := append([]world.RegionID(nil), region.Neighbors...)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func aiNavalMissionMove(gs *state.GameState, fleet *army.Army, ctx *StrategicContext) (world.RegionID, bool) {
	if gs == nil || fleet == nil || !fleet.IsNaval || gs.ScenarioID != "1300_ottoman_rise" {
		return "", false
	}
	// Docked filoların RegionID'si zaten komşu deniz ankrajıdır. Bu yüzden
	// merchant rota kontrolü bu ankrajı "hedefe ulaştı" sanıp hiçbir hareket
	// üretmeden önce filonun gerçek liman bağını bırakmasını sağlamalıdır.
	if fleet.IsDocked() {
		sea := gs.Regions[fleet.RegionID]
		if sea == nil || !sea.IsSea {
			return "", true
		}
		return fleet.RegionID, true
	}
	if target, handled := aiMerchantTradeFleetMove(gs, fleet, ctx); handled {
		return target, true
	}
	if ctx == nil || ctx.FactionID != faction.FactionID(fleet.OwnerID) {
		if isWarshipFleet(fleet, gs.UnitTypes) {
			return aiNavalPatrolMove(gs, fleet, nil)
		}
		return "", true
	}
	mission := ctx.navalMission
	if mission == nil {
		// Görevsiz filo rastgele yabancı kıyı veya uzak deniz aramaz. Taşınmış
		// birlik varsa yalnız güvenli bir dost kıyıya tahliye edilir.
		if len(fleet.EmbarkedUnits) > 0 {
			for _, neighborID := range aiSortedNeighborIDs(gs.Regions[fleet.RegionID]) {
				target := gs.Regions[neighborID]
				if target == nil || target.IsSea || !aiCanDisembarkToLand(gs, fleet, target) {
					continue
				}
				if target.OwnerID == fleet.OwnerID || diplomacy.SameRealm(gs, faction.FactionID(fleet.OwnerID), faction.FactionID(target.OwnerID)) {
					return target.ID, true
				}
			}
		}
		if isWarshipFleet(fleet, gs.UnitTypes) {
			return aiNavalPatrolMove(gs, fleet, ctx)
		}
		return "", true
	}

	if len(fleet.EmbarkedUnits) > 0 {
		if mission.FleetID != "" && mission.FleetID != fleet.ID {
			return "", true
		}
		// Plan hedefi savaş state'i değiştikten sonra barışta kalmış olabilir.
		// Böyle bir hedefe doğru yüklü filoyu göndermek, hedef kıyısında
		// aiCanDisembarkToLand tarafından reddedildiğinde filoyu sonsuza kadar
		// kilitler. Önce mevcut savaş düşmanları arasından çıkarılabilir bir kıyıya
		// retarget et; geçerli hedef yoksa mevcut bekleme davranışını koru.
		target := gs.Regions[mission.TargetRegionID]
		if !aiCanDisembarkToLand(gs, fleet, target) {
			if !aiRetargetLoadedNavalMissionToWarCoast(gs, fleet, ctx, mission) {
				return "", true
			}
			target = gs.Regions[mission.TargetRegionID]
		}
		if fleet.RegionID == mission.LandingSeaRegionID {
			if !aiCanDisembarkToLand(gs, fleet, target) {
				return "", true
			}
			if defender := aiEnemyArmyInRegion(gs, fleet.OwnerID, target.ID); defender != nil && aiLandingStrength(gs, fleet) <= defender.TotalStrength(gs.UnitTypes) {
				return "", true
			}
			return target.ID, true
		}
		next := aiThreatAwareSeaNextStep(gs, ctx.FactionID, fleet.RegionID, mission.LandingSeaRegionID)
		if next != "" && !aiNavalFleetMeetsSafetyGate(gs, fleet, next) {
			return "", true
		}
		return next, true
	}

	if fleet.TransportCapacity(gs.UnitTypes) > 0 {
		if mission.EmbarkArmyID == "" {
			return "", true
		}
		next := aiThreatAwareSeaNextStep(gs, ctx.FactionID, fleet.RegionID, mission.EmbarkSeaRegionID)
		if next != "" && !aiNavalFleetMeetsSafetyGate(gs, fleet, next) {
			return "", true
		}
		return next, true
	}

	// Savaş gemileri aktif taşıma hattının çıkış noktasında toplanır; yüklenmiş
	// görev filosu varsa önce onun bulunduğu denize yaklaşır.
	escortTarget := mission.EmbarkSeaRegionID
	if mission.FleetID != "" {
		if missionFleet := gs.Armies[mission.FleetID]; missionFleet != nil {
			escortTarget = missionFleet.RegionID
			if escortTarget == fleet.RegionID {
				escortTarget = mission.LandingSeaRegionID
			}
		}
	}
	next := aiThreatAwareSeaNextStep(gs, ctx.FactionID, fleet.RegionID, escortTarget)
	if next != "" && !aiNavalFleetMeetsSafetyGate(gs, fleet, next) {
		return "", true
	}
	return next, true
}

type aiNavalWarCoastCandidate struct {
	target      *world.Region
	landingSea  world.RegionID
	seaDistance int
	route       aiSeaRouteResult
}

// aiRetargetLoadedNavalMissionToWarCoast, artık çıkarma yapılamayan mevcut
// plan hedefini mevcut savaşlardan en yakın ve kazanılabilir kıyı hedefiyle
// değiştirir. Bu yalnız yüklü filonun hareket görevini düzeltir; barışta olan
// devletlere yeni savaş ilan etmez ve stratejik planı save'e yazmaz.
func aiRetargetLoadedNavalMissionToWarCoast(gs *state.GameState, fleet *army.Army, ctx *StrategicContext, mission *aiNavalMission) bool {
	if gs == nil || fleet == nil || ctx == nil || mission == nil || !fleet.IsNaval || len(fleet.EmbarkedUnits) == 0 {
		return false
	}

	var best *aiNavalWarCoastCandidate
	landingStrength := aiLandingStrength(gs, fleet)
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.OwnerID == "" || region.OwnerID == fleet.OwnerID || !region.IsCoastal(gs.Regions) {
			continue
		}
		relation := diplomacy.Relation(gs, faction.FactionID(fleet.OwnerID), faction.FactionID(region.OwnerID))
		if relation == nil || relation.Stance != faction.StanceWar {
			continue
		}
		defender := aiEnemyArmyInRegion(gs, fleet.OwnerID, region.ID)
		if defender != nil && landingStrength <= defender.TotalStrength(gs.UnitTypes) {
			continue
		}
		landingSea, seaDistance := aiBestLandingSeaForContext(ctx, fleet.RegionID, region)
		if landingSea == "" || seaDistance < 0 {
			continue
		}
		route := ctx.navalSeaRoute(fleet.RegionID, landingSea)
		if !route.Reachable {
			continue
		}
		candidate := aiNavalWarCoastCandidate{
			target:      region,
			landingSea:  landingSea,
			seaDistance: seaDistance,
			route:       route,
		}
		if aiNavalWarCoastCandidateBetter(candidate, best) {
			copy := candidate
			best = &copy
		}
	}
	if best == nil {
		return false
	}

	mission.Kind = aiNavalMissionAssault
	mission.TargetRegionID = best.target.ID
	mission.TargetFactionID = faction.FactionID(best.target.OwnerID)
	mission.LandingSeaRegionID = best.landingSea
	mission.RouteThreatPower = best.route.MaxThreat
	mission.Score = 10000 + len(fleet.EmbarkedUnits)*20 - best.seaDistance*5
	return true
}

func aiNavalWarCoastCandidateBetter(candidate aiNavalWarCoastCandidate, best *aiNavalWarCoastCandidate) bool {
	if best == nil || candidate.seaDistance != best.seaDistance {
		return best == nil || candidate.seaDistance < best.seaDistance
	}
	if candidate.route.MaxThreat != best.route.MaxThreat {
		return candidate.route.MaxThreat < best.route.MaxThreat
	}
	if candidate.route.TotalThreat != best.route.TotalThreat {
		return candidate.route.TotalThreat < best.route.TotalThreat
	}
	return candidate.target.ID < best.target.ID
}

func aiNavalEmbarkArmyMove(gs *state.GameState, landArmy *army.Army, ctx *StrategicContext) (world.RegionID, bool) {
	if gs == nil || landArmy == nil || landArmy.IsNaval || gs.ScenarioID != "1300_ottoman_rise" || ctx == nil || ctx.navalMission == nil || ctx.navalMission.EmbarkArmyID != landArmy.ID {
		return "", false
	}
	mission := ctx.navalMission
	if landArmy.RegionID != mission.EmbarkRegionID {
		return ctx.routeNextStep(landArmy, mission.EmbarkRegionID, aiRouteFriendly), true
	}
	if aiFindEmbarkFleet(gs, landArmy.OwnerID, mission.EmbarkSeaRegionID, len(landArmy.Units)) == nil {
		return "", true
	}
	return mission.EmbarkSeaRegionID, true
}

func aiExecuteNavalMissionProduction(gs *state.GameState, fid faction.FactionID, budget *aiBudget, ctx *StrategicContext, steps *[]TurnStep) {
	if gs == nil || fid == "" {
		return
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated || gs.BuildingTypes == nil || gs.UnitTypes == nil {
		return
	}
	if ctx == nil || ctx.FactionID != fid {
		ctx = prepareStrategicContext(gs, fid)
	}
	mission := ctx.navalMission
	if mission == nil || mission.EmbarkRegionID == "" {
		return
	}
	embarkRegion := gs.Regions[mission.EmbarkRegionID]
	transportType := gs.UnitTypes["transport"]
	portType := gs.BuildingTypes["port"]
	if embarkRegion == nil || transportType == nil || portType == nil {
		return
	}

	requiredPortLevel := maxInt(1, transportType.RequiredBldgLevel)
	currentPortLevel := aiBuildingLevel(embarkRegion, "port")
	queuedPortLevels := aiQueuedBuildingCount(gs, embarkRegion.ID, "port", fid)
	if currentPortLevel < requiredPortLevel {
		if queuedPortLevels > 0 || currentPortLevel+queuedPortLevels >= portType.MaxPerRegion || !aiBuildingAllowed(gs, embarkRegion, "port", portType.RequiredTerrain) {
			return
		}
		portCost := economy.ResourceCost{
			Gold: portType.GoldCost, Grain: portType.GrainCost, Iron: portType.IronCost,
			Timber: portType.TimberCost, Stone: portType.StoneCost,
			Spice: portType.SpiceCost, Cloth: portType.ClothCost,
		}
		if !aiCanAffordForBudget(self, portCost, budget, aiBudgetNaval) || !aiApplyBudgetedCost(self, portCost, budget, aiBudgetNaval) {
			return
		}
		turns := aiBuildingTurnsRequired(embarkRegion, "port", portType.TurnsRequired, queuedPortLevels)
		aiEnqueueProduction(gs, fid, aiProductionKindBuilding, embarkRegion.ID, "port", turns)
		addTurnStep(steps, TurnStep{
			FactionID: fid, Kind: TurnStepBuild, TargetRegion: embarkRegion.ID, FocusRegion: embarkRegion.ID,
			Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, embarkRegion.ID) + " kıyısında denizaşırı görev için liman kuruyor.",
		})
		return
	}
	if !transportType.HasAllRequiredTechs(self.Research.Completed) {
		return
	}
	if transportType.CarryCapacity <= 0 {
		return
	}

	missingCapacity := maxInt(0, mission.RequiredCapacity-aiReachableTransportCapacity(gs, fid, mission.EmbarkSeaRegionID))
	for missingCapacity > 0 {
		if aiPendingUnitCountByRegion(gs, embarkRegion.ID, fid) >= aiMaxRegionQueue || aiLaneRemainingCapacity(gs, embarkRegion.ID, fid, transportType) <= 0 {
			break
		}
		currentUnits := 0
		for _, fleet := range aiSortedArmies(gs) {
			if fleet.OwnerID == string(fid) && fleet.IsNaval && fleet.RegionID == mission.EmbarkSeaRegionID {
				currentUnits += len(fleet.Units)
			}
		}
		if currentUnits+aiPendingNavalUnitCount(gs, mission.EmbarkSeaRegionID, fid) >= army.MaxArmySize {
			break
		}
		if !aiApplyUnitCostForBudget(self, transportType, budget, aiBudgetNaval) {
			break
		}
		aiEnqueueProduction(gs, fid, aiProductionKindUnit, embarkRegion.ID, "transport", transportType.TurnsRequired)
		addTurnStep(steps, TurnStep{
			FactionID: fid, Kind: TurnStepRecruit, TargetRegion: embarkRegion.ID, FocusRegion: mission.EmbarkSeaRegionID,
			Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, embarkRegion.ID) + " limanında denizaşırı görev için nakliye gemisi hazırlıyor.",
		})
		missingCapacity -= transportType.CarryCapacity
	}

	// Escort da yalnız somut taşıma görevinin tehdit ve çıkış hattında değerlendirilir.
	aiProduceMissionEscortIfNeeded(gs, fid, mission, embarkRegion, budget, steps)
}

func aiProduceMissionEscortIfNeeded(gs *state.GameState, fid faction.FactionID, mission *aiNavalMission, embarkRegion *world.Region, budget *aiBudget, steps *[]TurnStep) {
	if gs == nil || mission == nil || embarkRegion == nil {
		return
	}
	self := gs.Factions[fid]
	warshipType := gs.UnitTypes["warship"]
	portType := gs.BuildingTypes["port"]
	if self == nil || warshipType == nil || portType == nil || !warshipType.HasAllRequiredTechs(self.Research.Completed) {
		return
	}
	threatPower := maxInt(mission.RouteThreatPower, aiPortApproachThreatPower(gs, fid, mission.EmbarkSeaRegionID))
	if threatPower <= 0 {
		return
	}
	requiredPower := (threatPower*aiNavalMissionSafetyPercent + 99) / 100
	projectedPower := aiProjectedMissionFleetPower(gs, fid, mission.EmbarkSeaRegionID)
	if projectedPower >= requiredPower {
		return
	}

	requiredPortLevel := maxInt(1, warshipType.RequiredBldgLevel)
	currentPortLevel := aiBuildingLevel(embarkRegion, "port")
	queuedPortLevels := aiQueuedBuildingCount(gs, embarkRegion.ID, "port", fid)
	if currentPortLevel < requiredPortLevel {
		if queuedPortLevels > 0 || currentPortLevel+queuedPortLevels >= portType.MaxPerRegion || !aiBuildingAllowed(gs, embarkRegion, "port", portType.RequiredTerrain) {
			return
		}
		portCost := economy.ResourceCost{
			Gold: portType.GoldCost, Grain: portType.GrainCost, Iron: portType.IronCost,
			Timber: portType.TimberCost, Stone: portType.StoneCost,
			Spice: portType.SpiceCost, Cloth: portType.ClothCost,
		}
		if !aiCanAffordForBudget(self, portCost, budget, aiBudgetNaval) || !aiApplyBudgetedCost(self, portCost, budget, aiBudgetNaval) {
			return
		}
		turns := aiBuildingTurnsRequired(embarkRegion, "port", portType.TurnsRequired, queuedPortLevels)
		aiEnqueueProduction(gs, fid, aiProductionKindBuilding, embarkRegion.ID, "port", turns)
		addTurnStep(steps, TurnStep{
			FactionID: fid, Kind: TurnStepBuild, TargetRegion: embarkRegion.ID, FocusRegion: mission.EmbarkSeaRegionID,
			Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, embarkRegion.ID) + " görev limanını escort üretimi için güçlendiriyor.",
		})
		return
	}

	unitPower := aiEffectiveNavalPower(gs, &army.Army{
		OwnerID: string(fid), IsNaval: true, Units: []army.Unit{{TypeID: warshipType.ID, CurrentHP: army.MaxUnitHP}},
	}, true)
	if unitPower <= 0 {
		return
	}
	for projectedPower < requiredPower {
		if aiPendingUnitCountByRegion(gs, embarkRegion.ID, fid) >= aiMaxRegionQueue || aiLaneRemainingCapacity(gs, embarkRegion.ID, fid, warshipType) <= 0 {
			break
		}
		currentUnits := 0
		for _, fleet := range aiSortedArmies(gs) {
			if fleet.OwnerID == string(fid) && fleet.IsNaval && fleet.RegionID == mission.EmbarkSeaRegionID {
				currentUnits += len(fleet.Units)
			}
		}
		if currentUnits+aiPendingNavalUnitCount(gs, mission.EmbarkSeaRegionID, fid) >= army.MaxArmySize {
			break
		}
		if !aiApplyUnitCostForBudget(self, warshipType, budget, aiBudgetNaval) {
			break
		}
		aiEnqueueProduction(gs, fid, aiProductionKindUnit, embarkRegion.ID, warshipType.ID, warshipType.TurnsRequired)
		addTurnStep(steps, TurnStep{
			FactionID: fid, Kind: TurnStepRecruit, TargetRegion: embarkRegion.ID, FocusRegion: mission.EmbarkSeaRegionID,
			Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, embarkRegion.ID) + " görev hattı için escort savaş gemisi hazırlıyor.",
		})
		projectedPower += unitPower
	}
}

func aiProjectedMissionFleetPower(gs *state.GameState, fid faction.FactionID, embarkSea world.RegionID) int {
	if gs == nil || embarkSea == "" {
		return 0
	}
	power := 0
	for _, fleet := range aiSortedArmies(gs) {
		if fleet.OwnerID != string(fid) || !fleet.IsNaval || aiSeaRouteDistance(gs, fleet.RegionID, embarkSea) < 0 {
			continue
		}
		power += aiEffectiveNavalPower(gs, fleet, true)
	}
	for _, order := range gs.ProductionQueue {
		if order.Kind != aiProductionKindUnit || order.FactionID != string(fid) {
			continue
		}
		unitType := gs.UnitTypes[order.TypeID]
		if unitType == nil || unitType.RequiredBldg != "port" {
			continue
		}
		sea := aiSeaNeighbor(gs, gs.Regions[order.RegionID])
		if sea == "" || aiSeaRouteDistance(gs, sea, embarkSea) < 0 {
			continue
		}
		power += aiEffectiveNavalPower(gs, &army.Army{
			OwnerID: string(fid), IsNaval: true, Units: []army.Unit{{TypeID: unitType.ID, CurrentHP: army.MaxUnitHP}},
		}, true)
	}
	return power
}
