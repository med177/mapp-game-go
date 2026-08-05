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

type aiCompositionTarget struct {
	Infantry int
	Cavalry  int
	Siege    int
}

type aiLandComposition struct {
	Infantry int
	Cavalry  int
	Siege    int
	Total    int
}

type aiRecruitmentBattleNeeds struct {
	AttackWeight         int
	DefenseWeight        int
	MoraleWeight         int
	TargetTerrain        world.TerrainType
	FortifiedTarget      bool
	FortifiedTargetCount int
	RequiredSiegeUnits   int
	SiegeShortfall       int
}

type aiUnitCandidate struct {
	TypeID          string
	Category        army.UnitCategory
	Score           int
	CompositionNeed int
	QualityScore    int
	EffectiveCost   int
	SustainedCombat int
}

func aiCompositionTargetForPlan(plan *state.AIPlanState) aiCompositionTarget {
	if plan == nil {
		return aiCompositionTarget{Infantry: 65, Cavalry: 25, Siege: 10}
	}
	switch plan.Kind {
	case state.AIObjectiveExpand:
		return aiCompositionTarget{Infantry: 55, Cavalry: 25, Siege: 20}
	case state.AIObjectiveDefend:
		return aiCompositionTarget{Infantry: 75, Cavalry: 15, Siege: 10}
	default:
		return aiCompositionTarget{Infantry: 65, Cavalry: 25, Siege: 10}
	}
}

func aiSelectStrategicLandUnit(gs *state.GameState, self *faction.Faction, budget *aiBudget, ctx *StrategicContext) string {
	return aiSelectStrategicLandUnitWithResourceCheck(gs, self, budget, ctx, true)
}

// aiSelectStrategicLandUnitForProcurement aynı askerî seçim puanını kullanır,
// ancak mevcut malzeme stoğu eksik olduğu için adayı elmez. Böylece tedarik
// aşaması hangi üretimin bloke olduğunu önceden görebilir.
func aiSelectStrategicLandUnitForProcurement(gs *state.GameState, self *faction.Faction, ctx *StrategicContext) string {
	return aiSelectStrategicLandUnitWithResourceCheck(gs, self, nil, ctx, false)
}

func aiSelectStrategicLandUnitWithResourceCheck(gs *state.GameState, self *faction.Faction, budget *aiBudget, ctx *StrategicContext, requireResources bool) string {
	if gs == nil || self == nil || self.IsEliminated || gs.UnitTypes == nil {
		return ""
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
	for unitTypeID := range gs.UnitTypes {
		unitIDs = append(unitIDs, unitTypeID)
	}
	sort.Strings(unitIDs)

	var best aiUnitCandidate
	found := false
	for _, unitTypeID := range unitIDs {
		unitType := gs.UnitTypes[unitTypeID]
		if unitType == nil || !aiLandUnitCategory(unitType.Category) || !aiUnitCandidateAvailableForSelection(gs, self, unitType, budget, ctx, requireResources) {
			continue
		}
		if aiFindRecruitRegionForStrategicContext(gs, self.ID, unitType, ctx) == "" {
			continue
		}
		// Tedarik öncesi aday ararken eksik kaynak cezası uygulanmaz. Aksi halde
		// düşük demir, demir isteyen piyadeyi sistematik olarak cezalandırıp
		// maliyeti demirsiz olan milisi seçtirir; böylece demir talebi hiç
		// oluşmaz ve AI aynı darboğanda kalır. Gerçek üretim seçiminde mevcut
		// stok baskısı korunur.
		candidate := aiScoreLandUnitCandidate(gs, self, unitType, target, composition, needs, economySnapshot, requireResources)
		if !found || aiUnitCandidateBetter(candidate, best) {
			best = candidate
			found = true
		}
	}
	if !found {
		return ""
	}
	return best.TypeID
}

func aiUnitCandidateAvailableForSelection(gs *state.GameState, self *faction.Faction, unitType *army.UnitType, budget *aiBudget, ctx *StrategicContext, requireResources bool) bool {
	if gs == nil || self == nil || unitType == nil || !unitType.HasAllRequiredTechs(self.Research.Completed) {
		return false
	}
	if !requireResources {
		return self.Gold-unitType.GoldCost >= aiMinGoldReserve && aiFindRecruitRegionForStrategicContext(gs, self.ID, unitType, ctx) != ""
	}
	return aiUnitAvailableForBudget(gs, self, unitType, budget)
}

// aiCompositionTargetForStrategicContext planın genel oranını aktif ana
// cephenin ihtiyacına göre daraltır. Böylece aynı devlet savunma objective'inde
// olsa bile uzun savaştaki gerçek ana saldırı için kuşatma ağırlığı üretebilir;
// ikincil cepheler ise gereksiz hücum birimi yığmaz.
func aiCompositionTargetForStrategicContext(gs *state.GameState, fid faction.FactionID, plan *state.AIPlanState, ctx *StrategicContext) aiCompositionTarget {
	target := aiCompositionTargetForPlan(plan)
	if gs == nil || ctx == nil {
		return target
	}
	if ctx.CriticalThreat {
		return aiCompositionTarget{Infantry: 75, Cavalry: 15, Siege: 10}
	}
	primaryEnemy := primaryOffensiveFrontEnemy(ctx, plan)
	if primaryEnemy == "" {
		return target
	}
	for _, front := range ctx.Fronts {
		if !front.AtWar || front.EnemyFactionID != primaryEnemy || front.TargetRegionID == "" {
			continue
		}
		region := gs.Regions[front.TargetRegionID]
		if region != nil && region.FortificationLevel() > 0 {
			return aiCompositionTarget{Infantry: 55, Cavalry: 25, Siege: 20}
		}
		return aiCompositionTarget{Infantry: 60, Cavalry: 25, Siege: 15}
	}
	return target
}

func aiScoreLandUnitCandidate(gs *state.GameState, self *faction.Faction, unitType *army.UnitType, target aiCompositionTarget, composition aiLandComposition, needs aiRecruitmentBattleNeeds, economySnapshot aiEconomySnapshot, includeResourcePressurePenalty bool) aiUnitCandidate {
	compositionNeed := aiCategoryCompositionNeed(target, composition, unitType.Category)
	combatValue := unitType.Attack*needs.AttackWeight + unitType.Defense*needs.DefenseWeight + unitType.Morale*needs.MoraleWeight
	cost := aiUnitResourceCost(unitType)
	effectiveCost := aiGoldEquivalentCost(gs, cost)
	efficiency := combatValue * 100 / maxInt(1, effectiveCost)
	resourcePenalty := 0
	if includeResourcePressurePenalty {
		resourcePenalty = aiUnitResourcePressurePenalty(self, cost)
	}
	grainPressure := aiGrainUtilityPercent(economySnapshot)
	upkeepMultiplier := 3
	sustainedCombatWeight := 1
	if grainPressure >= 100 {
		upkeepMultiplier = 10
		sustainedCombatWeight = 2
	}
	// Birliklerin mutlak tahıl bakımı tek başına doğru karar verdirmez:
	// milis ile elit piyade arasındaki bakım farkı sınırlıyken savaş değeri
	// çok büyüktür. Güç/tahıl verimi, altın maliyeti ve stok baskısıyla birlikte
	// hesaba katılır; böylece AI kaynakları uygunsa elit kaliteye geçer.
	sustainedCombat := combatValue * 100 / maxInt(1, unitType.GrainUpkeep)
	turns := maxInt(1, unitType.TurnsRequired)
	efficiencyWeight := 3
	if !includeResourcePressurePenalty {
		// Tedarik planı, kaynak açığını kapatacağı için mevcut altın-verim
		// baskısını üretim kararındaki kadar ağır taşımamalı; aksi halde ucuz
		// milis, daha güçlü ama demir isteyen piyadeyi her zaman bastırır.
		efficiencyWeight = 1
	}
	qualityScore := combatValue + efficiency*efficiencyWeight + sustainedCombat*sustainedCombatWeight/8 - resourcePenalty - unitType.GrainUpkeep*upkeepMultiplier - turns*8
	score := compositionNeed*4 + qualityScore
	if unitType.Category == army.CategorySiege && needs.FortifiedTarget {
		if needs.SiegeShortfall > 0 {
			score += 1200
		} else {
			score += 120
		}
	}
	return aiUnitCandidate{
		TypeID:          unitType.ID,
		Category:        unitType.Category,
		Score:           score,
		CompositionNeed: compositionNeed,
		QualityScore:    qualityScore,
		EffectiveCost:   effectiveCost,
		SustainedCombat: sustainedCombat,
	}
}

func aiCategoryCompositionNeed(target aiCompositionTarget, composition aiLandComposition, category army.UnitCategory) int {
	targetPercent := 0
	current := 0
	switch category {
	case army.CategoryInfantry:
		targetPercent = target.Infantry
		current = composition.Infantry
	case army.CategoryCavalry:
		targetPercent = target.Cavalry
		current = composition.Cavalry
	case army.CategorySiege:
		targetPercent = target.Siege
		current = composition.Siege
	default:
		return 0
	}
	return maxInt(0, targetPercent*(composition.Total+1)-current*100)
}

func aiFactionLandComposition(gs *state.GameState, fid faction.FactionID) aiLandComposition {
	composition := aiLandComposition{}
	if gs == nil {
		return composition
	}
	addUnit := func(unitTypeID string) {
		unitType := gs.UnitTypes[unitTypeID]
		if unitType == nil {
			return
		}
		switch unitType.Category {
		case army.CategoryInfantry:
			composition.Infantry++
		case army.CategoryCavalry:
			composition.Cavalry++
		case army.CategorySiege:
			composition.Siege++
		default:
			return
		}
		composition.Total++
	}
	for _, armyRef := range aiSortedArmies(gs) {
		if armyRef.OwnerID != string(fid) {
			continue
		}
		if armyRef.IsNaval {
			for _, unit := range armyRef.EmbarkedUnits {
				addUnit(unit.TypeID)
			}
			continue
		}
		for _, unit := range armyRef.Units {
			addUnit(unit.TypeID)
		}
	}
	for _, order := range gs.ProductionQueue {
		if order.Kind == aiProductionKindUnit && order.FactionID == string(fid) {
			addUnit(order.TypeID)
		}
	}
	return composition
}

func aiBuildRecruitmentBattleNeeds(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext) aiRecruitmentBattleNeeds {
	needs := aiRecruitmentBattleNeeds{AttackWeight: 2, DefenseWeight: 2, MoraleWeight: 1, TargetTerrain: world.TerrainPlain}
	plan := gs.AIPlans[fid]
	if plan != nil {
		switch plan.Kind {
		case state.AIObjectiveExpand:
			needs.AttackWeight = 3
		case state.AIObjectiveDefend:
			needs.DefenseWeight = 3
		}
	}
	if target := aiRecruitmentTargetRegion(gs, plan, ctx); target != nil {
		needs.TargetTerrain = target.Terrain
		targetOwner := faction.FactionID(target.OwnerID)
		hostileTarget := targetOwner != "" && !diplomacy.SameRealm(gs, fid, targetOwner)
		needs.FortifiedTarget = hostileTarget && target.FortificationLevel() > 0
		terrainPressure := 0
		switch target.Terrain {
		case world.TerrainMountain, world.TerrainPass:
			terrainPressure = 2
			needs.MoraleWeight++
		case world.TerrainForest:
			terrainPressure = 1
			needs.MoraleWeight++
		case world.TerrainCoast:
			terrainPressure = 1
		}
		if hostileTarget {
			needs.AttackWeight += terrainPressure
		} else {
			needs.DefenseWeight += terrainPressure
		}
	}
	needs.FortifiedTargetCount = aiFortifiedOffensiveTargetCount(gs, fid, plan, ctx)
	needs.FortifiedTarget = needs.FortifiedTarget || needs.FortifiedTargetCount > 0
	enemyAttack, enemyDefense := aiRecruitmentEnemyProfile(gs, fid, plan, ctx)
	if enemyAttack > enemyDefense {
		needs.DefenseWeight++
	} else if enemyDefense > enemyAttack {
		needs.AttackWeight++
	}
	if needs.FortifiedTarget {
		needs.RequiredSiegeUnits = minInt(3, maxInt(1, needs.FortifiedTargetCount))
		needs.SiegeShortfall = aiOffensiveSiegeShortfall(gs, fid, ctx, needs.RequiredSiegeUnits)
	}
	return needs
}

// aiFortifiedOffensiveTargetCount, aktif plan ve savaş cephelerindeki farklı
// tahkimli düşman bölgelerini sayar. Bu sayı birden fazla kaleye karşı tek bir
// mancınıkla yetinmek yerine küçük bir kuşatma kolu kurmak için kullanılır.
func aiFortifiedOffensiveTargetCount(gs *state.GameState, fid faction.FactionID, plan *state.AIPlanState, ctx *StrategicContext) int {
	if gs == nil || fid == "" {
		return 0
	}
	seen := make(map[world.RegionID]struct{})
	add := func(regionID world.RegionID) {
		region := gs.Regions[regionID]
		if region == nil || region.IsSea || region.OwnerID == "" || region.FortificationLevel() <= 0 || diplomacy.SameRealm(gs, fid, faction.FactionID(region.OwnerID)) {
			return
		}
		seen[regionID] = struct{}{}
	}
	if plan != nil {
		for _, regionID := range plan.TargetRegionIDs {
			add(regionID)
		}
	}
	if ctx != nil {
		for _, front := range ctx.Fronts {
			if !front.AtWar {
				continue
			}
			add(front.TargetRegionID)
			for _, regionID := range front.EnemyRegions {
				add(regionID)
			}
		}
	}
	return len(seen)
}

func aiOffensiveSiegeShortfall(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext, requiredSiegeUnits int) int {
	missing := 0
	assignedOffense := 0
	activeSiegeUnits := 0
	if requiredSiegeUnits <= 0 {
		requiredSiegeUnits = 1
	}
	for _, armyRef := range aiSortedArmies(gs) {
		if armyRef == nil || armyRef.OwnerID != string(fid) || armyRef.IsNaval || armyRef.IsGarrison {
			continue
		}
		for _, unit := range armyRef.Units {
			if unitType := gs.UnitTypes[unit.TypeID]; unitType != nil && unitType.Category == army.CategorySiege {
				activeSiegeUnits++
			}
		}
	}
	if ctx != nil {
		for armyID, assignment := range ctx.ArmyAssignments {
			if assignment.Role != AIArmyRoleAssault && assignment.Role != AIArmyRoleSiege {
				continue
			}
			armyRef := gs.Armies[armyID]
			if armyRef == nil || armyRef.OwnerID != string(fid) || armyRef.IsNaval {
				continue
			}
			assignedOffense++
			if !armyRef.HasSiegeUnits(gs.UnitTypes) {
				missing++
			}
		}
	}
	if assignedOffense == 0 {
		missing = requiredSiegeUnits
	}
	if capacityShortfall := requiredSiegeUnits - activeSiegeUnits; capacityShortfall > missing {
		missing = capacityShortfall
	}
	for _, order := range gs.ProductionQueue {
		if missing <= 0 {
			break
		}
		unitType := gs.UnitTypes[order.TypeID]
		if order.Kind == aiProductionKindUnit && order.FactionID == string(fid) && unitType != nil && unitType.Category == army.CategorySiege {
			missing--
		}
	}
	return maxInt(0, missing)
}

func aiRecruitmentTargetRegion(gs *state.GameState, plan *state.AIPlanState, ctx *StrategicContext) *world.Region {
	if gs == nil {
		return nil
	}
	if plan != nil {
		for _, regionID := range plan.TargetRegionIDs {
			if region := gs.Regions[regionID]; region != nil && !region.IsSea && region.OwnerID != "" {
				return region
			}
		}
	}
	if ctx != nil {
		for _, front := range ctx.Fronts {
			if !front.AtWar {
				continue
			}
			if front.TargetRegionID != "" {
				if region := gs.Regions[front.TargetRegionID]; region != nil && !region.IsSea {
					return region
				}
			}
			for _, regionID := range front.EnemyRegions {
				if region := gs.Regions[regionID]; region != nil && !region.IsSea {
					return region
				}
			}
		}
	}
	return nil
}

func aiRecruitmentEnemyProfile(gs *state.GameState, fid faction.FactionID, plan *state.AIPlanState, ctx *StrategicContext) (int, int) {
	enemyIDs := make(map[faction.FactionID]struct{})
	if plan != nil && plan.TargetFactionID != "" {
		enemyIDs[plan.TargetFactionID] = struct{}{}
	}
	if ctx != nil {
		for _, enemyID := range ctx.WarEnemies {
			enemyIDs[enemyID] = struct{}{}
		}
	}
	if len(enemyIDs) == 0 {
		for _, relation := range gs.Relations {
			if relation == nil || relation.Stance != faction.StanceWar {
				continue
			}
			switch fid {
			case relation.FactionA:
				enemyIDs[relation.FactionB] = struct{}{}
			case relation.FactionB:
				enemyIDs[relation.FactionA] = struct{}{}
			}
		}
	}
	attack := 0
	defense := 0
	for _, armyRef := range aiSortedArmies(gs) {
		if armyRef.IsNaval {
			continue
		}
		if _, enemy := enemyIDs[faction.FactionID(armyRef.OwnerID)]; !enemy {
			continue
		}
		for _, unit := range armyRef.Units {
			unitType := gs.UnitTypes[unit.TypeID]
			if unitType == nil {
				continue
			}
			attack += unitType.Attack
			defense += unitType.Defense
		}
	}
	return attack, defense
}

func aiUnitResourcePressurePenalty(self *faction.Faction, cost economy.ResourceCost) int {
	if self == nil {
		return 0
	}
	penalty := 0
	for _, resource := range []struct {
		stock int
		cost  int
	}{
		{self.Grain, cost.Grain},
		{self.Iron, cost.Iron},
		{self.Timber, cost.Timber},
		{self.Stone, cost.Stone},
	} {
		if resource.cost <= 0 {
			continue
		}
		if resource.stock < resource.cost*2 {
			penalty += 40
		} else if resource.stock < resource.cost*4 {
			penalty += 15
		}
	}
	return penalty
}

func aiUnitResourceCost(unitType *army.UnitType) economy.ResourceCost {
	if unitType == nil {
		return economy.ResourceCost{}
	}
	return economy.ResourceCost{
		Gold: unitType.GoldCost, Grain: unitType.GrainCost, Iron: unitType.IronCost,
		Timber: unitType.TimberCost, Stone: unitType.StoneCost,
		Spice: unitType.SpiceCost, Cloth: unitType.ClothCost,
	}
}

func aiLandUnitCategory(category army.UnitCategory) bool {
	return category == army.CategoryInfantry || category == army.CategoryCavalry || category == army.CategorySiege
}

func aiUnitCandidateBetter(candidate, best aiUnitCandidate) bool {
	if candidate.Score != best.Score {
		return candidate.Score > best.Score
	}
	if candidate.CompositionNeed != best.CompositionNeed {
		return candidate.CompositionNeed > best.CompositionNeed
	}
	if candidate.QualityScore != best.QualityScore {
		return candidate.QualityScore > best.QualityScore
	}
	if candidate.EffectiveCost != best.EffectiveCost {
		return candidate.EffectiveCost < best.EffectiveCost
	}
	return candidate.TypeID < best.TypeID
}
