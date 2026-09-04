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
	AIArmyRoleAssault      AIArmyRole = "assault"
	AIArmyRoleSiege        AIArmyRole = "siege"
	AIArmyRoleDefense      AIArmyRole = "defense"
	AIArmyRoleReserve      AIArmyRole = "reserve"
	AIArmyRoleRelief       AIArmyRole = "relief"
	AIArmyRoleRetreat      AIArmyRole = "retreat"
	AIArmyRoleSecurity     AIArmyRole = "security"
	AIArmyRoleTransport    AIArmyRole = "transport"
	AIArmyRoleEscort       AIArmyRole = "escort"
	AIArmyRolePatrol       AIArmyRole = "patrol"
	aiFrontTargetLockTurns            = 4
)

type AIFront struct {
	EnemyFactionID   faction.FactionID
	FriendlyRegions  []world.RegionID
	EnemyRegions     []world.RegionID
	TargetRegionID   world.RegionID
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
	if !aiStrategicPlanningEnabled(gs) {
		return ctx
	}
	ensureStrategicPlan(gs, fid, ctx)
	buildAIFronts(ctx)
	ensurePlanRallyState(ctx)
	assignAIArmyRoles(ctx)
	applySecurityAssignments(ctx)
	applyRetreatAssignments(ctx)
	applyRallyAssignments(ctx)
	buildAINavalThreatSnapshot(ctx)
	ctx.navalMission = buildAINavalMission(ctx)
	assignAINavalRoles(ctx)
	return ctx
}

// assignAINavalRoles projects the active naval mission onto the same runtime
// ArmyAssignments map used by land forces. Transport and escort fleets keep
// their operational role visible to movement and diagnostics without entering
// the land reserve/front allocation.
func assignAINavalRoles(ctx *StrategicContext) {
	if ctx == nil || ctx.gs == nil || ctx.FactionID == "" {
		return
	}
	mission := ctx.navalMission
	for _, fleet := range aiSortedArmies(ctx.gs) {
		if fleet == nil || !fleet.IsNaval || fleet.OwnerID != string(ctx.FactionID) {
			continue
		}
		if fleet.TransportCapacity(ctx.gs.UnitTypes) > 0 {
			anchor := fleet.RegionID
			reason := "nakliye filosu"
			if mission != nil {
				if fleet.ID == mission.FleetID && mission.LandingSeaRegionID != "" {
					anchor = mission.LandingSeaRegionID
					reason = "aktif nakliye görevi"
				} else if mission.EmbarkSeaRegionID != "" {
					anchor = mission.EmbarkSeaRegionID
					reason = "nakliye çıkışına yaklaş"
				}
			}
			ctx.ArmyAssignments[fleet.ID] = AIArmyAssignment{Role: AIArmyRoleTransport, AnchorRegionID: anchor, Reason: reason}
			continue
		}
		if aiFleetHasWarship(ctx.gs, fleet) {
			anchor := fleet.RegionID
			reason := "ticaret ve liman devriyesi"
			role := AIArmyRolePatrol
			if mission != nil && mission.EmbarkSeaRegionID != "" {
				anchor = mission.EmbarkSeaRegionID
				reason = "nakliye hattı escortu"
				role = AIArmyRoleEscort
			}
			ctx.ArmyAssignments[fleet.ID] = AIArmyAssignment{Role: role, AnchorRegionID: anchor, Reason: reason}
		}
	}
}

func aiFleetHasWarship(gs *state.GameState, fleet *army.Army) bool {
	if gs == nil || fleet == nil || gs.UnitTypes == nil {
		return false
	}
	for _, unit := range fleet.Units {
		if unitType := gs.UnitTypes[unit.TypeID]; unitType != nil && unitType.Category == army.CategoryNavalWar {
			return true
		}
	}
	return false
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
				continue
			}
			if aiCoordinatedWarParticipant(gs, ctx.FactionID, faction.FactionID(armyRef.OwnerID), enemyID) {
				if _, ok := builder.friendlyRegion[armyRef.RegionID]; ok {
					builder.front.FriendlyPower += armyRef.TotalStrength(gs.UnitTypes)
				}
			}
		}

		if sharedTarget := sharedAIWarTarget(ctx, builder.front); sharedTarget != "" {
			if ledger := gs.WarLedgerFor(ctx.FactionID, enemyID); ledger != nil {
				ledger.TargetRegionID = sharedTarget
				ledger.TargetLockedTurn = gs.Turn
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
		builder.front.TargetRegionID = selectAIFrontTarget(ctx, builder.front)
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

func selectAIFrontTarget(ctx *StrategicContext, front AIFront) world.RegionID {
	if ctx == nil || ctx.gs == nil {
		return ""
	}
	if front.AtWar {
		if ledger := ctx.gs.WarLedgerFor(ctx.FactionID, front.EnemyFactionID); ledger != nil && ledger.TargetRegionID != "" {
			if target := ctx.gs.Regions[ledger.TargetRegionID]; target != nil && target.OwnerID == string(front.EnemyFactionID) && containsRegionID(front.EnemyRegions, target.ID) {
				if siege := ctx.gs.SiegeAt(target.ID); siege != nil || ctx.gs.Turn-ledger.TargetLockedTurn < aiFrontTargetLockTurns {
					return target.ID
				}

			}
		}
		if target := sharedAIWarTarget(ctx, front); target != "" {
			return target
		}
	}
	plan := ctx.gs.AIPlans[ctx.FactionID]
	capitalRegionID := world.RegionID("")
	if capital, _, _, ok := ctx.gs.FactionCapital(front.EnemyFactionID); ok && capital != nil {
		capitalRegionID = capital.ID
	}

	bestRegion := world.RegionID("")
	bestScore := -1
	for _, regionID := range front.EnemyRegions {
		region := ctx.gs.Regions[regionID]
		if region == nil || region.IsSea || region.OwnerID != string(front.EnemyFactionID) {
			continue
		}
		score := ctx.strategicRegionValue(region)
		score += len(region.Settlements) * 12
		score += region.FortificationLevel() * 10
		if region.ID == capitalRegionID {
			score += 350
		}
		if siege := ctx.gs.SiegeAt(region.ID); siege != nil {
			attacker := ctx.gs.Armies[siege.AttackerArmyID]
			if attacker != nil && diplomacy.SameRealm(ctx.gs, ctx.FactionID, faction.FactionID(attacker.OwnerID)) {
				score += 220
			}
		}
		for index, targetID := range planTargetRegions(plan) {
			if targetID == region.ID {
				score += 180 - index*30
				break
			}
		}

		defenderPower := 0
		friendlyAccess := 0
		for _, neighborID := range region.Neighbors {
			neighbor := ctx.gs.Regions[neighborID]
			if neighbor == nil || neighbor.IsSea {
				continue
			}
			if neighbor.OwnerID == string(ctx.FactionID) {
				friendlyAccess++
			}
		}
		for _, armyRef := range aiSortedArmies(ctx.gs) {
			if armyRef != nil && !armyRef.IsNaval && armyRef.OwnerID == string(front.EnemyFactionID) && armyRef.RegionID == region.ID {
				defenderPower += armyRef.TotalStrength(ctx.gs.UnitTypes)
			}
		}
		score += friendlyAccess * 15
		score -= minInt(240, defenderPower/4)
		if bestRegion == "" || score > bestScore || (score == bestScore && region.ID < bestRegion) {
			bestRegion = region.ID
			bestScore = score
		}
	}
	if front.AtWar {
		if ledger := ctx.gs.WarLedgerFor(ctx.FactionID, front.EnemyFactionID); ledger != nil && bestRegion != "" {
			ledger.TargetRegionID = bestRegion
			ledger.TargetLockedTurn = ctx.gs.Turn
		}
	}
	return bestRegion
}

func aiCoordinatedWarParticipant(gs *state.GameState, commander, candidate, enemy faction.FactionID) bool {
	if gs == nil || commander == "" || candidate == "" || enemy == "" || candidate == enemy {
		return false
	}
	if !diplomacy.SameRealm(gs, commander, candidate) {
		rel := diplomacy.Relation(gs, commander, candidate)
		if rel == nil || rel.Stance != faction.StanceAllied {
			return false
		}
	}
	war := diplomacy.Relation(gs, candidate, enemy)
	return war != nil && war.Stance == faction.StanceWar
}

func sharedAIWarTarget(ctx *StrategicContext, front AIFront) world.RegionID {
	if ctx == nil || ctx.gs == nil || !front.AtWar {
		return ""
	}
	for _, candidate := range aiSortedFactionIDs(ctx.gs) {
		if candidate == ctx.FactionID || !aiCoordinatedWarParticipant(ctx.gs, ctx.FactionID, candidate, front.EnemyFactionID) {
			continue
		}
		ledger := ctx.gs.WarLedgerFor(candidate, front.EnemyFactionID)
		if ledger == nil || ledger.TargetRegionID == "" {
			continue
		}
		target := ctx.gs.Regions[ledger.TargetRegionID]
		if target != nil && target.OwnerID == string(front.EnemyFactionID) && containsRegionID(front.EnemyRegions, target.ID) {
			return target.ID
		}
	}
	return ""
}

func planTargetRegions(plan *state.AIPlanState) []world.RegionID {
	if plan == nil {
		return nil
	}
	return plan.TargetRegionIDs
}

func containsRegionID(regionIDs []world.RegionID, target world.RegionID) bool {
	for _, regionID := range regionIDs {
		if regionID == target {
			return true
		}
	}
	return false
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
	warSupplyCrisis := aiWarLogisticsPolicyActive(gs) && aiWarSupplyCrisis(gs, ctx.FactionID)
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
		candidate := nearestReliefArmy(ctx, mobile, regionID)
		if candidate != nil {
			ctx.ArmyAssignments[candidate.ID] = AIArmyAssignment{Role: AIArmyRoleRelief, AnchorRegionID: regionID, Reason: "dost kuşatmasını kaldır"}
			continue
		}
		group := selectReliefRallyGroup(ctx, mobile, regionID)
		if len(group) == 0 {
			continue
		}
		rallyRegionID := group[0].RegionID
		for _, groupArmy := range group {
			ctx.ArmyAssignments[groupArmy.ID] = AIArmyAssignment{
				Role:           AIArmyRoleRelief,
				AnchorRegionID: rallyRegionID,
				Rallying:       true,
				Reason:         "kuşatmayı kaldırmak için yardım kuvveti toplanıyor",
			}
		}
	}

	reserveAnchor := aiReserveAnchor(ctx)
	ctx.ReservePercent = aiReservePercentForFrontRisk(ctx)
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
	primaryOffensiveEnemy := primaryOffensiveFrontEnemy(ctx, plan)
	for _, front := range ctx.Fronts {
		if !front.AtWar || front.AnchorRegionID == "" {
			continue
		}
		// Genişleme objective'i zaten aşağıdaki genel objective hücumuyla
		// yürütülür. Buradaki özel rol yalnız savunma/konsolidasyon planının
		// aktif savaşı tamamen kilitlemesini önlemek içindir.
		if plan != nil && plan.Kind == state.AIObjectiveExpand {
			if front.ThreatScore <= 0 || ((front.ObjectiveRelated || front.EnemyFactionID == offensiveTarget) && !front.CriticalThreat) {
				continue
			}
			candidate := nearestUnassignedArmy(ctx, mobile, front.AnchorRegionID, strongest, true)
			if candidate == nil {
				continue
			}
			ctx.ArmyAssignments[candidate.ID] = AIArmyAssignment{Role: AIArmyRoleDefense, AnchorRegionID: front.AnchorRegionID, FrontFactionID: front.EnemyFactionID, Reason: "tehdit altındaki cephe"}
			continue
		}
		candidate := nearestUnassignedArmy(ctx, mobile, front.AnchorRegionID, strongest, true)
		if candidate == nil {
			continue
		}
		if !aiActiveWarMatureForOffense(gs, ctx.FactionID, front.EnemyFactionID) {
			ctx.ArmyAssignments[candidate.ID] = AIArmyAssignment{Role: AIArmyRoleDefense, AnchorRegionID: front.AnchorRegionID, FrontFactionID: front.EnemyFactionID, Reason: "aktif savaşta seferberlik"}
			continue
		}
		if warSupplyCrisis {
			ctx.ArmyAssignments[candidate.ID] = AIArmyAssignment{Role: AIArmyRoleDefense, AnchorRegionID: front.AnchorRegionID, FrontFactionID: front.EnemyFactionID, Reason: "lojistik rezerv toparlanması"}
			continue
		}
		if front.CriticalThreat || front.CapitalThreat {
			ctx.ArmyAssignments[candidate.ID] = AIArmyAssignment{Role: AIArmyRoleDefense, AnchorRegionID: front.AnchorRegionID, FrontFactionID: front.EnemyFactionID, Reason: "tehdit altındaki cephe"}
			continue
		}
		attackAnchor := world.RegionID("")
		if front.TargetRegionID != "" {
			attackAnchor = front.TargetRegionID
		}
		if attackAnchor == "" {
			attackAnchor = offensiveAnchor
		}
		if attackAnchor == "" {
			continue
		}
		if primaryOffensiveEnemy != front.EnemyFactionID {
			ctx.ArmyAssignments[candidate.ID] = AIArmyAssignment{Role: AIArmyRoleDefense, AnchorRegionID: front.AnchorRegionID, FrontFactionID: front.EnemyFactionID, Reason: "ikincil savaş cephesi savunması"}
			continue
		}
		role := AIArmyRoleAssault
		reason := "aktif savaş cephesi hücumu"
		if candidate.HasSiegeUnits(gs.UnitTypes) {
			role = AIArmyRoleSiege
			reason = "aktif savaş cephesi kuşatması"
		}
		ctx.ArmyAssignments[candidate.ID] = AIArmyAssignment{Role: role, AnchorRegionID: attackAnchor, FrontFactionID: front.EnemyFactionID, Reason: reason}
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
		if warSupplyCrisis && (assignment.Role == AIArmyRoleAssault || assignment.Role == AIArmyRoleSiege) {
			assignment.Role = AIArmyRoleDefense
			assignment.AnchorRegionID = aiDefenseAnchor(ctx)
			assignment.FrontFactionID = ""
			assignment.Reason = "lojistik rezerv toparlanması"
		}
		if assignment.AnchorRegionID == "" {
			assignment.AnchorRegionID = armyRef.RegionID
		}
		ctx.ArmyAssignments[armyRef.ID] = assignment
	}
}

func aiReservePercentForFrontRisk(ctx *StrategicContext) int {
	if ctx == nil {
		return 15
	}
	if ctx.CriticalThreat {
		return 30
	}
	activeWars := 0
	threatenedFront := false
	for _, front := range ctx.Fronts {
		if !front.AtWar {
			continue
		}
		activeWars++
		if front.ThreatScore > 0 {
			threatenedFront = true
		}
	}
	if activeWars >= 2 || threatenedFront {
		return 25
	}
	return 15
}

func primaryOffensiveFrontEnemy(ctx *StrategicContext, plan *state.AIPlanState) faction.FactionID {
	if ctx == nil || ctx.gs == nil {
		return ""
	}
	bestEnemy := faction.FactionID("")
	bestScore := -1
	for _, front := range ctx.Fronts {
		if !front.AtWar || front.AnchorRegionID == "" || front.CriticalThreat || front.CapitalThreat || front.TargetRegionID == "" {
			continue
		}
		if !aiActiveWarMatureForOffense(ctx.gs, ctx.FactionID, front.EnemyFactionID) {
			continue
		}
		score := 0
		if plan != nil && plan.TargetFactionID == front.EnemyFactionID {
			score += 300
		}
		if target := ctx.gs.Regions[front.TargetRegionID]; target != nil {
			score += ctx.strategicRegionValue(target)
		}
		score += maxInt(0, front.FriendlyPower-front.EnemyPower) / 4
		if bestEnemy == "" || score > bestScore || (score == bestScore && front.EnemyFactionID < bestEnemy) {
			bestEnemy = front.EnemyFactionID
			bestScore = score
		}
	}
	return bestEnemy
}

func aiWarLogisticsPolicyActive(gs *state.GameState) bool {
	return gs != nil && gs.Turn > aiWarLogisticsActivationTurn
}

func aiActiveWarMatureForOffense(gs *state.GameState, actor, opponent faction.FactionID) bool {
	if gs == nil || actor == "" || opponent == "" {
		return false
	}
	ledger := gs.WarLedgerFor(actor, opponent)
	return ledger != nil && gs.Turn-ledger.StartedTurn >= 12
}

func aiOffensiveAnchor(ctx *StrategicContext, plan *state.AIPlanState) (world.RegionID, faction.FactionID) {
	if ctx == nil || ctx.gs == nil || plan == nil {
		return "", ""
	}
	for _, front := range ctx.Fronts {
		if front.AtWar && front.EnemyFactionID == plan.TargetFactionID {
			// Mevcut expand objective'inin öncelik sırası açılış temposunun
			// parçasıdır; yeni cephe hedef skoru savunma/konsolidasyon fallback'i
			// için kullanılmalıdır.
			return firstOwnedRegion(ctx.gs, plan.TargetRegionIDs, plan.TargetFactionID), plan.TargetFactionID
		}
	}
	// Kalıcı objective başka bir devleti gösterse bile ilan edilmiş savaşı sahipsiz
	// bırakma. Plan yeniden değerlendirildiğinde kalıcı hedef de bu cepheyle hizalanır.
	for _, front := range ctx.Fronts {
		if front.AtWar && front.TargetRegionID != "" {
			return front.TargetRegionID, front.EnemyFactionID
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
		if siege == nil || region == nil || region.IsSea {
			continue
		}
		targetOwner := faction.FactionID(region.OwnerID)
		if targetOwner == "" || (targetOwner != ctx.FactionID && !aiCoordinatedWarParticipant(ctx.gs, ctx.FactionID, targetOwner, faction.FactionID(siege.AttackerFactionID))) {
			continue
		}
		siegeArmy := ctx.gs.Armies[siege.AttackerArmyID]
		if siegeArmy == nil || diplomacy.SameRealm(ctx.gs, ctx.FactionID, faction.FactionID(siegeArmy.OwnerID)) {
			continue
		}
		if siege.AttackerFactionID == "" || faction.FactionID(siegeArmy.OwnerID) != faction.FactionID(siege.AttackerFactionID) {
			continue
		}
		if rel := diplomacy.Relation(ctx.gs, targetOwner, faction.FactionID(siege.AttackerFactionID)); rel == nil || rel.Stance != faction.StanceWar {
			continue
		}
		targets = append(targets, regionID)
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i] < targets[j] })
	return targets
}

const aiReliefPowerMarginPercent = 110

// nearestReliefArmy, kuşatmayı kaldırma ihtimali olan en yakın orduyu seçer.
func nearestReliefArmy(ctx *StrategicContext, armies []*army.Army, target world.RegionID) *army.Army {
	if ctx == nil || ctx.gs == nil {
		return nil
	}
	var best *army.Army
	bestDistance := int(^uint(0) >> 1)
	for _, candidate := range armies {
		if candidate == nil {
			continue
		}
		if _, assigned := ctx.ArmyAssignments[candidate.ID]; assigned || !aiCanDefeatSiege(ctx, candidate, target) {
			continue
		}
		distance := ctx.routeDistance(candidate, target, aiRouteGeneral)
		if distance < 0 {
			continue
		}
		if best == nil || distance < bestDistance || (distance == bestDistance && candidate.ID < best.ID) {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}

func aiCanDefeatSiege(ctx *StrategicContext, candidate *army.Army, target world.RegionID) bool {
	if ctx == nil || ctx.gs == nil || candidate == nil || target == "" {
		return false
	}
	siege := ctx.gs.SiegeAt(target)
	if siege == nil || siege.AttackerArmyID == candidate.ID {
		return false
	}
	siegeArmy := ctx.gs.Armies[siege.AttackerArmyID]
	if siegeArmy == nil || len(candidate.Units) == 0 || len(siegeArmy.Units) == 0 {
		return false
	}
	return candidate.TotalStrength(ctx.gs.UnitTypes)*100 >= siegeArmy.TotalStrength(ctx.gs.UnitTypes)*aiReliefPowerMarginPercent
}

// selectReliefRallyGroup, tek ordunun yetmediği durumda kuşatıcıyı yenebilecek
// toplam güce ulaşana kadar hedefe en yakın erişilebilir orduları seçer.
// Seçilen ordular ilk adayın bölgesinde toplanır; mevcut tur sonu birleşme
// akışı onları kapasite elverdiği ölçüde tek orduya dönüştürür.
func selectReliefRallyGroup(ctx *StrategicContext, armies []*army.Army, target world.RegionID) []*army.Army {
	if ctx == nil || ctx.gs == nil {
		return nil
	}
	siege := ctx.gs.SiegeAt(target)
	if siege == nil {
		return nil
	}
	siegeArmy := ctx.gs.Armies[siege.AttackerArmyID]
	if siegeArmy == nil || len(siegeArmy.Units) == 0 {
		return nil
	}
	candidates := make([]*army.Army, 0, len(armies))
	distances := make(map[army.ArmyID]int, len(armies))
	for _, candidate := range armies {
		if candidate == nil || candidate.IsNaval || candidate.IsGarrison || len(candidate.Units) == 0 {
			continue
		}
		if _, assigned := ctx.ArmyAssignments[candidate.ID]; assigned {
			continue
		}
		distance := ctx.routeDistance(candidate, target, aiRouteGeneral)
		if distance < 0 {
			continue
		}
		candidates = append(candidates, candidate)
		distances[candidate.ID] = distance
	}
	sort.Slice(candidates, func(i, j int) bool {
		if distances[candidates[i].ID] != distances[candidates[j].ID] {
			return distances[candidates[i].ID] < distances[candidates[j].ID]
		}
		return candidates[i].ID < candidates[j].ID
	})

	group := make([]*army.Army, 0, len(candidates))
	groupPower := 0
	for _, candidate := range candidates {
		group = append(group, candidate)
		groupPower += candidate.TotalStrength(ctx.gs.UnitTypes)
		if groupPower*100 >= siegeArmy.TotalStrength(ctx.gs.UnitTypes)*aiReliefPowerMarginPercent {
			return group
		}
	}
	return nil
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
	if ctx == nil || ctx.gs == nil {
		return true
	}
	// Warning seviyesindeki rezerv eksikliği hazırlık ve saldırı temposunu
	// düşürür, fakat yeni ve zayıf bir hedefe karşı tüm fırsat savaşlarını
	// kilitlemez. Gerçek tahıl krizi ise saldırıyı hâlâ tamamen durdurur.
	if aiWarLogisticsPolicyActive(ctx.gs) && !aiWarLogisticsReady(ctx.gs, ctx.FactionID) && aiWarSupplyCrisis(ctx.gs, ctx.FactionID) {
		return false
	}
	if ctx.CriticalThreat || (ctx.ReserveTargetPower > 0 && ctx.ReserveAssignedPower*100 < ctx.ReserveTargetPower*50) {
		return false
	}
	if ctx.RallyActive {
		return false
	}
	navalMissionReady := ctx.navalMission != nil && ctx.navalMission.Kind == aiNavalMissionAssault && ctx.navalMission.TargetFactionID == target && ctx.navalMission.EmbarkArmyID != ""
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
	if attackPower <= 0 && !navalMissionReady {
		// Savunma/konsolidasyon planı düşük tehditli bir cephede saldırı rolü
		// atamamış olabilir. Sınırda yeterli kuvvet varsa bu ordu fırsat savaşı
		// için kullanılabilir; kritik tehditler yukarıdaki kapıda elenir.
		if target == "" || ctx.CriticalThreat {
			return false
		}
		attackPower = aiFrontierPower(ctx.gs, ctx.FactionID, target)
		if attackPower <= 0 {
			return false
		}
	}
	if attackPower <= 0 && navalMissionReady {
		if embarkArmy := ctx.gs.Armies[ctx.navalMission.EmbarkArmyID]; embarkArmy != nil {
			attackPower = embarkArmy.TotalStrength(ctx.gs.UnitTypes)
		}
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

// aiWarLogisticsReady yeni bir cephe açmadan önce mevcut ekonomik durumun
// saldırıyı taşıyıp taşıyamadığını kontrol eder. Runtime tahıl snapshot'ı
// varsa ekonomi tick'inin sonucu, yoksa aynı talep hesaplarının state fallback'i
// kullanılır. Böylece save yükleme veya ilk tur planlaması farklı davranmaz.
func aiWarLogisticsReady(gs *state.GameState, fid faction.FactionID) bool {
	if gs == nil || fid == "" {
		return true
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return false
	}

	goldReserve := aiMinGoldReserve
	if budget := prepareAIBudget(gs, fid, nil); budget != nil && budget.EmergencyGold > goldReserve {
		goldReserve = budget.EmergencyGold
	}
	if self.Gold < goldReserve {
		return false
	}

	if status, ok := gs.GrainEconomy[fid]; ok && status.TotalDemand > 0 {
		if status.SupplyLevel >= state.GrainSupplyCritical {
			return false
		}
		return status.MonthsOfSupply < 0 || status.MonthsOfSupply >= aiWarMinimumGrainReserveMonths
	}

	totalDemand := 0
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		totalDemand += gs.CivilianGrainDemandForRegion(region)
	}
	for _, currentArmy := range gs.Armies {
		if currentArmy != nil && currentArmy.OwnerID == string(fid) {
			totalDemand += gs.EffectiveArmyGrainUpkeep(currentArmy)
		}
	}
	if totalDemand <= 0 {
		return true
	}
	requiredGrain := maxInt(aiWarMinimumGrainReserve, totalDemand*aiWarMinimumGrainReserveMonths)
	return self.Grain >= requiredGrain
}

// aiWarSupplyCrisis aktif savaşın hücumunu tamamen durduracak kadar ağır
// tahıl kıtlığını bildirir. Warning seviyesi savaşın temposunu düşürür, fakat
// cepheyi gereksiz yere savunmaya kilitlemez; kritik/famine seviyesi ise
// saldırı ordularını savunma ve ikmal görevine çevirir.
func aiWarSupplyCrisis(gs *state.GameState, fid faction.FactionID) bool {
	if gs == nil || fid == "" {
		return false
	}
	if status, ok := gs.GrainEconomy[fid]; ok && status.TotalDemand > 0 {
		return status.SupplyLevel >= state.GrainSupplyCritical
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return true
	}
	totalDemand := 0
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		totalDemand += gs.CivilianGrainDemandForRegion(region)
	}
	for _, currentArmy := range gs.Armies {
		if currentArmy != nil && currentArmy.OwnerID == string(fid) {
			totalDemand += gs.EffectiveArmyGrainUpkeep(currentArmy)
		}
	}
	return totalDemand > 0 && self.Grain < totalDemand
}
