package ai

import (
	"fmt"
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	aiStrategicPlanReassessTurns = 6
	aiStrategicPlanRegionLimit   = 4
)

// StrategicContext bir AI turu boyunca kullanılan, save'e yazılmayan türetilmiş
// stratejik veridir. Pahalı güç/değer hesapları yalnız ihtiyaç halinde cache'lenir.
type StrategicContext struct {
	FactionID            faction.FactionID
	Turn                 int
	ManpowerCap          int
	DeployedLandUnits    int
	OwnedLandRegionIDs   []world.RegionID
	BorderRegionIDs      []world.RegionID
	WarEnemies           []faction.FactionID
	Fronts               []AIFront
	ArmyAssignments      map[army.ArmyID]AIArmyAssignment
	TotalMobilePower     int
	ReservePercent       int
	ReserveTargetPower   int
	ReserveAssignedPower int
	CriticalThreat       bool
	RallyRegionID        world.RegionID
	RallyDeadlineTurn    int
	RallyRequiredPower   int
	RallyGatheredPower   int
	RallyActive          bool
	RallyReady           bool
	NavalThreats         []AINavalThreat
	ThreatenedPortIDs    []world.RegionID
	navalMission         *aiNavalMission

	gs               *state.GameState
	factionPower     map[faction.FactionID]int
	frontierPower    map[faction.FactionID]int
	regionValue      map[world.RegionID]int
	routeCache       map[aiRouteCacheKey]*aiRouteMap
	navalThreatPower map[world.RegionID]int
	budget           *aiBudget
}

type scenarioObjectiveCandidate struct {
	objective scenario.AIObjectiveDef
	targetID  faction.FactionID
	regions   []world.RegionID
	score     int
	order     int
	kind      state.AIObjectiveKind
}

func buildStrategicContext(gs *state.GameState, fid faction.FactionID) *StrategicContext {
	ctx := &StrategicContext{
		FactionID:     fid,
		factionPower:  make(map[faction.FactionID]int),
		frontierPower: make(map[faction.FactionID]int),
		regionValue:   make(map[world.RegionID]int),
		gs:            gs,
	}
	if gs == nil {
		return ctx
	}
	ctx.Turn = gs.Turn
	ctx.ManpowerCap = gs.ManpowerCap(fid)
	ctx.DeployedLandUnits = gs.DeployedLandUnits(fid)

	borderSet := make(map[world.RegionID]struct{})
	for _, region := range aiSortedRegions(gs) {
		if region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		ctx.OwnedLandRegionIDs = append(ctx.OwnedLandRegionIDs, region.ID)
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil || neighbor.IsSea || neighbor.OwnerID == "" || neighbor.OwnerID == string(fid) {
				continue
			}
			borderSet[region.ID] = struct{}{}
			break
		}
	}
	for regionID := range borderSet {
		ctx.BorderRegionIDs = append(ctx.BorderRegionIDs, regionID)
	}
	sort.Slice(ctx.BorderRegionIDs, func(i, j int) bool { return ctx.BorderRegionIDs[i] < ctx.BorderRegionIDs[j] })

	for _, otherID := range aiSortedFactionIDs(gs) {
		if otherID == fid {
			continue
		}
		if rel := diplomacy.Relation(gs, fid, otherID); rel != nil && rel.Stance == faction.StanceWar {
			ctx.WarEnemies = append(ctx.WarEnemies, otherID)
		}
	}
	return ctx
}

func (ctx *StrategicContext) militaryPower(fid faction.FactionID) int {
	if ctx == nil || ctx.gs == nil || fid == "" {
		return 0
	}
	if value, ok := ctx.factionPower[fid]; ok {
		return value
	}
	value := diplomacy.MilitaryPower(ctx.gs, fid)
	ctx.factionPower[fid] = value
	return value
}

func (ctx *StrategicContext) powerAtFrontier(target faction.FactionID) int {
	if ctx == nil || ctx.gs == nil || target == "" {
		return 0
	}
	if value, ok := ctx.frontierPower[target]; ok {
		return value
	}
	value := aiFrontierPower(ctx.gs, ctx.FactionID, target)
	ctx.frontierPower[target] = value
	return value
}

func (ctx *StrategicContext) strategicRegionValue(region *world.Region) int {
	if ctx == nil || ctx.gs == nil || region == nil {
		return 0
	}
	if value, ok := ctx.regionValue[region.ID]; ok {
		return value
	}
	value := aiRegionStrategicValue(ctx.gs, region)
	ctx.regionValue[region.ID] = value
	return value
}

func ensureStrategicPlan(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext) *state.AIPlanState {
	if !aiStrategicPlanningEnabled(gs) {
		return nil
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		if gs.AIPlans != nil {
			delete(gs.AIPlans, fid)
		}
		return nil
	}
	if ctx == nil {
		ctx = buildStrategicContext(gs, fid)
	}
	if existing := gs.AIPlans[fid]; existing != nil {
		recordCompletedScenarioObjective(gs, fid, existing.ObjectiveID)
		if aiStrategicPlanValid(gs, fid, existing) {
			return existing
		}
	}

	plan := chooseStrategicPlan(gs, fid, ctx)
	if plan == nil {
		return nil
	}
	if gs.AIPlans == nil {
		gs.AIPlans = make(map[faction.FactionID]*state.AIPlanState)
	}
	gs.AIPlans[fid] = plan
	return plan
}

// aiStrategicPlanningEnabled tüm senaryolarda kalıcı AI planlarını açar.
// Tarihsel hedefler ve strateji profilleri varsa plan seçimini zenginleştirir;
// yoksa genel konsolidasyon/genişleme davranışı kullanılır.
func aiStrategicPlanningEnabled(gs *state.GameState) bool {
	return gs != nil
}

func aiStrategicPlanValid(gs *state.GameState, fid faction.FactionID, plan *state.AIPlanState) bool {
	if gs == nil || plan == nil || plan.ObjectiveID == "" || plan.Kind == "" {
		return false
	}
	if plan.ReassessTurn <= gs.Turn {
		return false
	}
	if !aiPlanHardGateActive(gs, fid, plan.ObjectiveID) {
		return false
	}
	if plan.Kind == state.AIObjectiveConsolidate {
		return true
	}
	if plan.TargetFactionID == "" || plan.TargetFactionID == fid {
		return false
	}
	target := gs.Factions[plan.TargetFactionID]
	if target == nil || target.IsEliminated || diplomacy.SameRealm(gs, fid, plan.TargetFactionID) || len(gs.LandRegionsOwnedBy(plan.TargetFactionID)) == 0 {
		return false
	}
	if strategy, ok := gs.AIStrategies[string(fid)]; ok {
		for _, objective := range strategy.Objectives {
			if objective.ID == plan.ObjectiveID {
				if scenarioObjectiveCompleted(gs, fid, objective) {
					return false
				}
				objectiveForPlan := objective
				if objective.Kind == string(state.AIObjectiveConsolidate) && plan.Kind == state.AIObjectiveExpand {
					// Bir consolidate claim'i kaybedildiyse plan recovery
					// amacıyla expand'e çevrilir; doğrulama da hedef devletin
					// tuttuğu kayıp claim'i aramalıdır.
					objectiveForPlan.Kind = string(state.AIObjectiveExpand)
				}
				return len(objectiveRelevantRegions(gs, fid, plan.TargetFactionID, objectiveForPlan)) > 0
			}
		}
	}
	return true
}

func chooseStrategicPlan(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext) *state.AIPlanState {
	self := gs.Factions[fid]
	if self == nil {
		return nil
	}
	if plan := chooseScenarioObjectivePlan(gs, self, ctx); plan != nil {
		return plan
	}
	if plan := chooseVictoryStrategicPlan(gs, self, ctx); plan != nil {
		return plan
	}
	if targetID := aiBestExpansionPlanTarget(gs, self, ctx); targetID != "" {
		return newStrategicPlan(gs, self, ctx, state.AIObjectiveExpand, targetID, "senaryo genişleme hedefi")
	}
	if len(ctx.WarEnemies) > 0 {
		targetID := ctx.WarEnemies[0]
		return newStrategicPlan(gs, self, ctx, state.AIObjectiveDefend, targetID, "aktif savaş cephesi")
	}

	return newConsolidationPlan(gs, self, ctx, "yakın genişleme hedefi yok; iç konsolidasyon")
}

func chooseScenarioObjectivePlan(gs *state.GameState, self *faction.Faction, ctx *StrategicContext) *state.AIPlanState {
	if gs == nil || self == nil || ctx == nil || len(gs.AIStrategies) == 0 {
		return nil
	}
	strategy, ok := gs.AIStrategies[string(self.ID)]
	if !ok {
		return nil
	}

	var best *scenarioObjectiveCandidate
	for objectiveIndex, objective := range strategy.Objectives {
		if scenarioObjectiveCompleted(gs, self.ID, objective) {
			markScenarioObjectiveCompleted(gs, self.ID, objective.ID)
			continue
		}
		if !scenarioObjectiveHardGateActive(gs, objective) {
			continue
		}
		// İttifak objective'i diplomasi metadata'sıdır. Bu planlayıcı yalnız
		// askeri expand/defend/consolidate objective'leri için AIPlan üretir;
		// ally kaydını saldırı hedefi gibi yorumlamak yanlış olur.
		if objective.Kind == "ally" {
			continue
		}
		kind := state.AIObjectiveKind(objective.Kind)
		if kind == state.AIObjectiveConsolidate {
			recoveryObjective := objective
			recoveryObjective.Kind = string(state.AIObjectiveExpand)
			recoveryTargets := objectiveRecoveryTargetFactionIDs(gs, self.ID, objective)
			if !scenarioObjectiveWasCompleted(gs, self.ID, objective.ID) {
				recoveryTargets = nil
			}
			if len(recoveryTargets) > 0 {
				// Bir claim dışarıda kaldıysa mevcut consolidate amacı artık
				// yalnızca iç hazırlık değildir: kaybedilen claim'i geri alma
				// planı olarak yeniden açılır.
				for targetIndex, targetID := range recoveryTargets {
					regions := objectiveRelevantRegions(gs, self.ID, targetID, recoveryObjective)
					if len(regions) == 0 {
						continue
					}
					score := objective.Priority*2 + 100 - objectiveIndex - targetIndex
					current := scenarioObjectiveCandidate{
						objective: recoveryObjective,
						targetID:  targetID,
						regions:   regions,
						score:     score,
						order:     objectiveIndex,
						kind:      state.AIObjectiveExpand,
					}
					if betterObjectiveCandidate(&current, best) {
						copy := current
						best = &copy
					}
				}
				continue
			}
			regions := objectiveRelevantRegions(gs, self.ID, "", objective)
			if len(regions) == 0 {
				continue
			}
			current := scenarioObjectiveCandidate{objective: objective, regions: regions, score: objective.Priority * 100, order: objectiveIndex, kind: kind}
			if betterObjectiveCandidate(&current, best) {
				copy := current
				best = &copy
			}
			continue
		}
		for targetIndex, targetID := range objectiveTargetFactionIDs(gs, self.ID, objective) {
			target := gs.Factions[targetID]
			if targetID == "" || targetID == self.ID || target == nil || target.IsEliminated || diplomacy.SameRealm(gs, self.ID, targetID) {
				continue
			}
			regions := objectiveRelevantRegions(gs, self.ID, targetID, objective)
			if len(regions) == 0 {
				continue
			}
			score := objective.Priority*2 - objectiveIndex - targetIndex
			// Tarihsel olarak kilitli bir objective açıldığında, artık geçerli
			// olmayan uzun vadeli konsolidasyon planının önüne geçmelidir. Bu
			// bonus yalnız MinYear/RequiredEventFlags taşıyan objective'lere
			// verilir; normal profil öncelikleri aynı kalır.
			if objective.MinYear > 0 || len(objective.RequiredEventFlags) > 0 {
				score += 10000
			}
			score += objectiveDeadlineUrgency(gs, objective)
			// Savunma objective'leri tarihsel profillerde genellikle daha yüksek
			// önceliklidir. Tehdit yokken bu ham öncelik, genişleme objective'lerini
			// sürekli gölgede bırakıp AI'yi kendi sınırlarında kilitliyordu.
			// Genişleme hedefi hâlâ cephe/güç/erişilebilirlik kontrollerinden geçer;
			// bonus yalnızca barış döneminde gerçek bir saldırı niyeti üretir.
			if kind == state.AIObjectiveExpand && gs.Turn >= aiWarLogisticsActivationTurn && !ctx.CriticalThreat && len(ctx.WarEnemies) == 0 {
				score += 55
			}
			if rel := diplomacy.Relation(gs, self.ID, targetID); rel != nil && rel.Stance == faction.StanceWar {
				score += 50
			}
			if aiSharesLandBorder(gs, self.ID, targetID) {
				score += 20
			}
			selfPower := ctx.militaryPower(self.ID)
			targetPower := ctx.militaryPower(targetID)
			if targetPower == 0 {
				score += 20
			} else {
				score += maxInt(-40, minInt(40, (selfPower-targetPower)/12))
			}
			for _, regionID := range objective.ReadinessRegions {
				if region := gs.Regions[world.RegionID(regionID)]; region != nil && region.OwnerID == string(self.ID) {
					score += 4
				} else {
					score -= 10
				}
			}
			current := scenarioObjectiveCandidate{objective: objective, targetID: targetID, regions: regions, score: score, order: objectiveIndex, kind: kind}
			if betterObjectiveCandidate(&current, best) {
				copy := current
				best = &copy
			}
		}
	}
	if best == nil {
		return newConsolidationPlan(gs, self, ctx, "senaryo objective'i tamamlandı veya zaman/event bekliyor")
	}

	commitment := best.objective.Commitment
	if commitment <= 0 {
		commitment = aiPlanCommitment(self, state.AIObjectiveKind(best.objective.Kind))
	}
	if commitment < 25 {
		commitment = 25
	}
	if commitment > 90 {
		commitment = 90
	}
	return &state.AIPlanState{
		ObjectiveID:        best.objective.ID,
		Kind:               best.kind,
		TargetFactionID:    best.targetID,
		TargetRegionIDs:    best.regions,
		StartedTurn:        gs.Turn,
		ReassessTurn:       gs.Turn + aiPlanHorizonTurns(gs),
		Commitment:         commitment,
		AllowVassalization: best.objective.AllowVassalization,
		Reason:             fmt.Sprintf("%s profili: %s", strategy.Profile, best.objective.ID),
	}
}

func betterObjectiveCandidate(current, best *scenarioObjectiveCandidate) bool {
	if current == nil {
		return false
	}
	if best == nil || current.score != best.score {
		return best == nil || current.score > best.score
	}
	if current.order != best.order {
		return current.order < best.order
	}
	return current.targetID < best.targetID
}

// scenarioObjectiveCompleted tamamlanabilir bölgesel objective'leri, bütün
// claim bölgeleri devletin eline geçtiğinde bitmiş sayar. Savunma objective'i
// bilerek bu kurala dahil değildir: kendi topraklarının elde olması savunma
// niyetinin bitmesi değil, korunacak çekirdeğin mevcut olmasıdır.
func scenarioObjectiveCompleted(gs *state.GameState, ownerID faction.FactionID, objective scenario.AIObjectiveDef) bool {
	if gs == nil || ownerID == "" || (objective.Kind != string(state.AIObjectiveExpand) && objective.Kind != string(state.AIObjectiveConsolidate)) {
		return false
	}
	regionIDs := objectiveClaimRegionIDs(objective)
	if len(regionIDs) == 0 {
		return false
	}
	for _, rawRegionID := range regionIDs {
		region := gs.Regions[world.RegionID(rawRegionID)]
		if region == nil || region.IsSea || region.OwnerID != string(ownerID) {
			return false
		}
	}
	return true
}

func scenarioObjectiveWasCompleted(gs *state.GameState, ownerID faction.FactionID, objectiveID string) bool {
	if gs == nil || ownerID == "" || objectiveID == "" || gs.AICompletedObjectives == nil {
		return false
	}
	return gs.AICompletedObjectives[ownerID][objectiveID]
}

func markScenarioObjectiveCompleted(gs *state.GameState, ownerID faction.FactionID, objectiveID string) {
	if gs == nil || ownerID == "" || objectiveID == "" {
		return
	}
	if gs.AICompletedObjectives == nil {
		gs.AICompletedObjectives = make(map[faction.FactionID]map[string]bool)
	}
	if gs.AICompletedObjectives[ownerID] == nil {
		gs.AICompletedObjectives[ownerID] = make(map[string]bool)
	}
	gs.AICompletedObjectives[ownerID][objectiveID] = true
}

func recordCompletedScenarioObjective(gs *state.GameState, ownerID faction.FactionID, objectiveID string) {
	if gs == nil || ownerID == "" || objectiveID == "" || gs.AIStrategies == nil {
		return
	}
	strategy, ok := gs.AIStrategies[string(ownerID)]
	if !ok {
		return
	}
	for _, objective := range strategy.Objectives {
		if objective.ID == objectiveID && scenarioObjectiveCompleted(gs, ownerID, objective) {
			markScenarioObjectiveCompleted(gs, ownerID, objectiveID)
			return
		}
	}
}

func objectiveRecoveryTargetFactionIDs(gs *state.GameState, ownerID faction.FactionID, objective scenario.AIObjectiveDef) []faction.FactionID {
	if objective.Kind != string(state.AIObjectiveConsolidate) {
		return nil
	}
	recoveryObjective := objective
	recoveryObjective.Kind = string(state.AIObjectiveExpand)
	return objectiveTargetFactionIDs(gs, ownerID, recoveryObjective)
}

func scenarioObjectiveHardGateActive(gs *state.GameState, objective scenario.AIObjectiveDef) bool {
	if gs == nil || (objective.MinYear > 0 && gs.Year < objective.MinYear) || (objective.MaxYear > 0 && gs.Year > objective.MaxYear) {
		return false
	}
	for _, flag := range objective.RequiredEventFlags {
		if flag == "" {
			continue
		}
		if !gs.FiredEventIDs["flag:"+flag] {
			return false
		}
	}
	return true
}

// objectiveDeadlineUrgency son geçerli yıl yaklaşırken objective'i öne alır.
// MaxYear yılı içinde objective hâlâ geçerlidir; bir sonraki yılda hard gate kapanır.
func objectiveDeadlineUrgency(gs *state.GameState, objective scenario.AIObjectiveDef) int {
	if gs == nil || objective.MaxYear <= 0 {
		return 0
	}
	yearsLeft := objective.MaxYear - gs.Year
	switch {
	case yearsLeft <= 0:
		return 180
	case yearsLeft == 1:
		return 120
	case yearsLeft <= 3:
		return 70
	case yearsLeft <= 5:
		return 35
	default:
		return 0
	}
}

func aiPlanHardGateActive(gs *state.GameState, fid faction.FactionID, objectiveID string) bool {
	if gs == nil || objectiveID == "" {
		return false
	}
	strategy, ok := gs.AIStrategies[string(fid)]
	if !ok {
		return true
	}
	for _, objective := range strategy.Objectives {
		if objective.ID == objectiveID {
			return scenarioObjectiveHardGateActive(gs, objective)
		}
	}
	return true
}

func objectiveRelevantRegions(gs *state.GameState, ownerID, targetID faction.FactionID, objective scenario.AIObjectiveDef) []world.RegionID {
	regionLimit := aiPlanTargetRegionLimit(gs)
	regions := make([]world.RegionID, 0, regionLimit)
	wantedOwner := string(targetID)
	if objective.Kind == string(state.AIObjectiveDefend) || objective.Kind == string(state.AIObjectiveConsolidate) {
		wantedOwner = string(ownerID)
	}
	targetRegionIDs := objectiveClaimRegionIDs(objective)
	if len(targetRegionIDs) == 0 && objective.Kind == string(state.AIObjectiveExpand) {
		if strategy, ok := gs.AIStrategies[string(ownerID)]; ok {
			for _, claim := range strategy.TerritorialClaims {
				targetRegionIDs = append(targetRegionIDs, claim.RegionID)
			}
		}
	}
	if len(targetRegionIDs) > 0 {
		for _, rawRegionID := range targetRegionIDs {
			regionID := world.RegionID(rawRegionID)
			region := gs.Regions[regionID]
			if region == nil || region.IsSea || region.OwnerID != wantedOwner {
				continue
			}
			regions = append(regions, regionID)
			if len(regions) == regionLimit {
				break
			}
		}
		return regions
	}
	for _, region := range aiSortedRegions(gs) {
		if region.IsSea || region.OwnerID != wantedOwner {
			continue
		}
		regions = append(regions, region.ID)
		if len(regions) == regionLimit {
			break
		}
	}
	return regions
}

// objectiveTargetFactionIDs hedef devlet metadata'sına değil, claim edilen
// bölgelerin güncel sahiplerine göre hesaplanır. Böylece bir bölge fethedilince
// AI aynı stratejik hedefi yeni sahibine karşı sürdürür.
func objectiveTargetFactionIDs(gs *state.GameState, ownerID faction.FactionID, objective scenario.AIObjectiveDef) []faction.FactionID {
	if gs == nil || ownerID == "" {
		return nil
	}
	if objective.Kind == string(state.AIObjectiveDefend) {
		regionIDs := objectiveClaimRegionIDs(objective)
		if len(regionIDs) == 0 {
			regionIDs = []string{stringRegionIDForFaction(gs, ownerID)}
		}
		seen := make(map[faction.FactionID]struct{})
		owners := make([]faction.FactionID, 0, len(regionIDs))
		for _, rawRegionID := range regionIDs {
			region := gs.Regions[world.RegionID(rawRegionID)]
			if region == nil || region.IsSea || region.OwnerID != string(ownerID) {
				continue
			}
			for _, neighborID := range region.Neighbors {
				neighbor := gs.Regions[neighborID]
				if neighbor == nil || neighbor.IsSea || neighbor.OwnerID == "" || neighbor.OwnerID == string(ownerID) {
					continue
				}
				targetID := faction.FactionID(neighbor.OwnerID)
				if gs.Factions[targetID] == nil {
					continue
				}
				if _, duplicate := seen[targetID]; duplicate {
					continue
				}
				seen[targetID] = struct{}{}
				owners = append(owners, targetID)
			}
		}
		sort.Slice(owners, func(i, j int) bool { return owners[i] < owners[j] })
		return owners
	}
	if objective.Kind != string(state.AIObjectiveExpand) {
		return nil
	}
	regionIDs := objectiveClaimRegionIDs(objective)
	if len(regionIDs) == 0 {
		if strategy, ok := gs.AIStrategies[string(ownerID)]; ok {
			for _, claim := range strategy.TerritorialClaims {
				regionIDs = append(regionIDs, claim.RegionID)
			}
		}
	}
	seen := make(map[faction.FactionID]struct{}, len(regionIDs))
	owners := make([]faction.FactionID, 0, len(regionIDs))
	for _, rawRegionID := range regionIDs {
		region := gs.Regions[world.RegionID(rawRegionID)]
		if region == nil || region.IsSea || region.OwnerID == "" || region.OwnerID == string(ownerID) {
			continue
		}
		targetID := faction.FactionID(region.OwnerID)
		if gs.Factions[targetID] == nil {
			continue
		}
		if _, duplicate := seen[targetID]; duplicate {
			continue
		}
		seen[targetID] = struct{}{}
		owners = append(owners, targetID)
	}
	sort.Slice(owners, func(i, j int) bool { return owners[i] < owners[j] })
	return owners
}

func objectiveClaimRegionIDs(objective scenario.AIObjectiveDef) []string {
	if len(objective.TerritorialClaims) > 0 {
		ids := make([]string, 0, len(objective.TerritorialClaims))
		for _, claim := range objective.TerritorialClaims {
			if claim.RegionID != "" {
				ids = append(ids, claim.RegionID)
			}
		}
		return ids
	}
	return objective.TargetRegions
}

func stringRegionIDForFaction(gs *state.GameState, ownerID faction.FactionID) string {
	if gs == nil || ownerID == "" {
		return ""
	}
	for _, region := range aiSortedRegions(gs) {
		if region != nil && !region.IsSea && region.OwnerID == string(ownerID) {
			return string(region.ID)
		}
	}
	return ""
}

func aiPlanTargetsFaction(gs *state.GameState, actor, target faction.FactionID) bool {
	if gs == nil || actor == "" || target == "" {
		return false
	}
	plan := gs.AIPlans[actor]
	return plan != nil && plan.Kind == state.AIObjectiveExpand && plan.TargetFactionID == target
}

// aiIsStrategicDiplomacyTarget ilişki onarımının savaş hazırlığıyla çelişeceği
// devletleri tek bir yerde tanımlar. Açık genişleme hedefi veya aktif expand
// planı yanında, bir claim bölgesini elinde tutan güncel sahibi de hedef sayılır.
// Claim sahibi savaş sonrasında değişebildiği için burada faction metadata'sı
// değil bölgenin anlık OwnerID değeri kullanılır.
func aiIsStrategicDiplomacyTarget(gs *state.GameState, actor, target faction.FactionID) bool {
	if gs == nil || actor == "" || target == "" || actor == target {
		return false
	}
	self := gs.Factions[actor]
	if self == nil {
		return false
	}
	if aiHasExpansionTarget(self, target) || aiPlanTargetsFaction(gs, actor, target) {
		return true
	}
	for _, claim := range self.TerritorialClaims {
		region := gs.Regions[world.RegionID(claim.RegionID)]
		if region != nil && !region.IsSea && region.OwnerID == string(target) {
			return true
		}
	}
	return false
}

func aiPlanMoveScoreBonus(gs *state.GameState, actor faction.FactionID, target *world.Region) int {
	if gs == nil || target == nil {
		return 0
	}
	plan := gs.AIPlans[actor]
	if plan == nil {
		return 0
	}
	for index, regionID := range plan.TargetRegionIDs {
		if regionID != target.ID {
			continue
		}
		if plan.Kind == state.AIObjectiveExpand && target.OwnerID != string(plan.TargetFactionID) {
			return 0
		}
		if plan.Kind == state.AIObjectiveDefend && target.OwnerID != string(actor) {
			return 0
		}
		bonus := 36 - index*4
		if plan.Kind == state.AIObjectiveDefend && target.OwnerID == string(actor) {
			bonus += 8
		}
		return maxInt(12, maxInt(16, bonus)*aiPlanMoveBonusPercent(gs)/100)
	}
	if plan.Kind == state.AIObjectiveExpand && target.OwnerID == string(plan.TargetFactionID) {
		return maxInt(8, 12*aiPlanMoveBonusPercent(gs)/100)
	}
	return 0
}

func aiBestExpansionPlanTarget(gs *state.GameState, self *faction.Faction, ctx *StrategicContext) faction.FactionID {
	if gs == nil || self == nil || ctx == nil {
		return ""
	}
	bestID := faction.FactionID("")
	bestScore := -1 << 30
	seen := make(map[faction.FactionID]struct{}, len(self.AIExpansionTargets))
	for index, targetID := range self.AIExpansionTargets {
		if targetID == "" || targetID == self.ID {
			continue
		}
		if _, duplicate := seen[targetID]; duplicate {
			continue
		}
		seen[targetID] = struct{}{}
		target := gs.Factions[targetID]
		if target == nil || target.IsEliminated || len(gs.LandRegionsOwnedBy(targetID)) == 0 {
			continue
		}
		rel := diplomacy.Relation(gs, self.ID, targetID)
		if rel != nil && rel.Stance == faction.StanceAllied {
			continue
		}

		score := 100 - index*8
		if aiSharesLandBorder(gs, self.ID, targetID) {
			score += 35
		}
		if rel != nil && rel.Stance == faction.StanceWar {
			score += 50
		}
		powerAdvantage := ctx.militaryPower(self.ID) - ctx.militaryPower(targetID)
		if powerAdvantage > 0 {
			score += minInt(30, powerAdvantage/10)
		} else {
			score += maxInt(-40, powerAdvantage/10)
		}
		score += minInt(20, ctx.powerAtFrontier(targetID)/20)
		if score > bestScore || (score == bestScore && (bestID == "" || targetID < bestID)) {
			bestScore = score
			bestID = targetID
		}
	}
	return bestID
}

func newStrategicPlan(gs *state.GameState, self *faction.Faction, ctx *StrategicContext, kind state.AIObjectiveKind, targetID faction.FactionID, reason string) *state.AIPlanState {
	return &state.AIPlanState{
		ObjectiveID:     fmt.Sprintf("%s:%s", kind, targetID),
		Kind:            kind,
		TargetFactionID: targetID,
		TargetRegionIDs: aiPlanTargetRegions(gs, self.ID, targetID, ctx),
		StartedTurn:     gs.Turn,
		ReassessTurn:    gs.Turn + aiPlanHorizonTurns(gs),
		Commitment:      aiPlanCommitment(self, kind),
		Reason:          reason,
	}
}

func newConsolidationPlan(gs *state.GameState, self *faction.Faction, ctx *StrategicContext, reason string) *state.AIPlanState {
	targetRegions := append([]world.RegionID(nil), ctx.BorderRegionIDs...)
	if len(targetRegions) == 0 && len(ctx.OwnedLandRegionIDs) > 0 {
		targetRegions = append(targetRegions, ctx.OwnedLandRegionIDs[0])
	}
	return &state.AIPlanState{
		ObjectiveID:     "consolidate:" + string(self.ID),
		Kind:            state.AIObjectiveConsolidate,
		TargetRegionIDs: targetRegions,
		StartedTurn:     gs.Turn,
		ReassessTurn:    gs.Turn + aiPlanHorizonTurns(gs),
		Commitment:      aiPlanCommitment(self, state.AIObjectiveConsolidate),
		Reason:          reason,
	}
}

func aiPlanTargetRegions(gs *state.GameState, ownerID, targetID faction.FactionID, ctx *StrategicContext) []world.RegionID {
	type scoredRegion struct {
		id     world.RegionID
		value  int
		border bool
	}
	regions := make([]scoredRegion, 0)
	for _, region := range aiSortedRegions(gs) {
		if region.IsSea || region.OwnerID != string(targetID) {
			continue
		}
		border := false
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor != nil && !neighbor.IsSea && neighbor.OwnerID == string(ownerID) {
				border = true
				break
			}
		}
		regions = append(regions, scoredRegion{id: region.ID, value: ctx.strategicRegionValue(region), border: border})
	}
	sort.Slice(regions, func(i, j int) bool {
		if regions[i].border != regions[j].border {
			return regions[i].border
		}
		if regions[i].value != regions[j].value {
			return regions[i].value > regions[j].value
		}
		return regions[i].id < regions[j].id
	})
	regionLimit := aiPlanTargetRegionLimit(gs)
	if len(regions) > regionLimit {
		regions = regions[:regionLimit]
	}
	out := make([]world.RegionID, 0, len(regions))
	for _, region := range regions {
		out = append(out, region.id)
	}
	return out
}

func aiPlanCommitment(self *faction.Faction, kind state.AIObjectiveKind) int {
	commitment := 40
	if self != nil {
		commitment = self.AIAggressiveness
	}
	if kind == state.AIObjectiveDefend && commitment < 60 {
		commitment = 60
	}
	if kind == state.AIObjectiveConsolidate {
		commitment = minInt(commitment, 50)
	}
	if commitment < 25 {
		return 25
	}
	if commitment > 90 {
		return 90
	}
	return commitment
}
