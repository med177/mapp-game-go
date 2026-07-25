package ai

import (
	"fmt"
	"hash/fnv"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

const (
	aiMilitiaID      = "militia"
	aiMilitiaCost    = 60  // units.json'daki milis maliyeti
	aiMinGoldReserve = 80  // AI bu miktarın altına düşmemeli
	aiTechReserve    = 100 // Teknoloji için ayırılacak minimum altın
	aiReliefMoveBase = 35
	aiWarThreshold   = 70
	// Yeni bir cephe açılmadan önce en az iki aylık operasyonel tahıl stoğu
	// korunur. Üç aylık stratejik rezerv ekonomi/ticaret sistemi içindir;
	// savaş hazırlığı bu rezervin tamamını kilitlemez.
	aiWarMinimumGrainReserveMonths = 2
	aiWarMinimumGrainReserve       = 100
	aiWarLogisticsActivationTurn   = 24
)

const (
	aiProductionKindBuilding = "building"
	aiProductionKindUnit     = "unit"
	aiMaxRegionQueue         = 20
)

// coalitionThreshold oyuncunun bu kadar bölgeyi geçmesi koalisyon tetikler.
const coalitionThreshold = 8

// TakeTurn belirtilen fraksiyon için tüm AI kararlarını verir ve uygular.
func aiTechMods(gs *state.GameState, ownerID string) combat.TechMods {
	f, ok := gs.Factions[faction.FactionID(ownerID)]
	if !ok || gs.TechTypes == nil {
		return combat.TechMods{}
	}
	fx := tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
	return combat.TechMods{
		AttackMod:       fx.InfantryAttackMod + fx.CavalryAttackMod + fx.SiegeAttackMod,
		DefenseMod:      fx.LandDefenseMod,
		NavalAttackMod:  fx.NavalAttackMod,
		NavalDefenseMod: fx.NavalDefenseMod,
	}
}

// relationScore iki fraksiyon arasındaki ilişki puanını döner; yoksa 0.
func relationScore(gs *state.GameState, a, b string) (int, faction.DiplomaticStance) {
	if rel := diplomacy.Relation(gs, faction.FactionID(a), faction.FactionID(b)); rel != nil {
		return rel.Score, rel.Stance
	}
	return 0, faction.StancePeace
}

// TakeTurn belirtilen fraksiyon için tüm AI kararlarını verir ve uygular.
func TakeTurn(gs *state.GameState, fid faction.FactionID) {
	strategicContext := runTurnPrelude(gs, fid, nil)

	// Ordu listesinin anlık kopyasını al — iterasyon sırasında map değişebilir
	var ownArmies []*army.Army
	for _, a := range aiSortedArmies(gs) {
		if a.OwnerID == string(fid) {
			ownArmies = append(ownArmies, a)
		}
	}

	for _, a := range ownArmies {
		// Ordu hâlâ haritada mı?
		if _, alive := gs.Armies[a.ID]; !alive {
			continue
		}
		moveArmyWithStrategicContext(gs, a, fid, nil, strategicContext)
	}
}

func runTurnPrelude(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) *StrategicContext {
	if gs == nil {
		return nil
	}
	var planningContext *StrategicContext
	if gs.ScenarioID == "1300_ottoman_rise" {
		planningContext = prepareStrategicContext(gs, fid)
	}
	// Difficulty 3: koalisyon mantığını çalıştır
	if gs.Difficulty >= 3 {
		formCoalitionAgainstPlayer(gs, fid, steps)
	}

	aiHandleDiplomacyWithSteps(gs, fid, steps)

	budget := prepareAIBudget(gs, fid, planningContext)
	if budget == nil {
		// Diğer senaryoların mevcut harcama sırasını ve sabit rezerv davranışını koru.
		aiResearchWithSteps(gs, fid, steps)
		aiEconomyBuildWithSteps(gs, fid, steps)
		aiNavalStrategyWithSteps(gs, fid, steps)
		aiRecruitAndBuildWithSteps(gs, fid, steps)
	} else {
		for _, category := range budget.Order {
			switch category {
			case aiBudgetArmy:
				aiRecruitAndBuildWithStrategicContextAndSteps(gs, fid, budget, planningContext, steps)
			case aiBudgetEconomy:
				aiEconomyBuildWithStrategicContextAndSteps(gs, fid, budget, planningContext, steps)
			case aiBudgetResearch:
				aiResearchWithStrategicContextAndSteps(gs, fid, budget, planningContext, steps)
			case aiBudgetNaval:
				aiNavalStrategyWithStrategicContextAndSteps(gs, fid, budget, planningContext, steps)
			}
			budget.release(category)
		}
	}

	// Aynı bölgede olan orduları konsolide et (önceki turlardan veya yeni alımlardan kalan)
	aiConsolidateArmies(gs, fid)

	// Yeni üretilen veya daha önce komutansız kalan AI ordularını kariyer havuzuna bağla.
	gs.EnsureFactionCommanders(string(fid))

	if gs.ScenarioID == "1300_ottoman_rise" {
		result := prepareStrategicContext(gs, fid)
		result.budget = budget
		return result
	}
	return nil
}

func addTurnStep(steps *[]TurnStep, step TurnStep) {
	if steps == nil || step.Message == "" {
		return
	}
	*steps = append(*steps, step)
}

func turnFactionName(gs *state.GameState, fid faction.FactionID) string {
	if gs == nil {
		return string(fid)
	}
	if f := gs.Factions[fid]; f != nil && f.NameTR != "" {
		return f.NameTR
	}
	return string(fid)
}

func turnRegionName(gs *state.GameState, rid world.RegionID) string {
	if gs == nil {
		return string(rid)
	}
	if region := gs.Regions[rid]; region != nil && region.NameTR != "" {
		return region.NameTR
	}
	return string(rid)
}

func aiHandleDiplomacy(gs *state.GameState, fid faction.FactionID) {
	aiHandleDiplomacyWithSteps(gs, fid, nil)
}

func aiPeaceOfferPriority(gs *state.GameState, from, to faction.FactionID) int {
	priority, _ := aiDiplomacyOfferPriorityDetails(gs, from, to, diplomacy.ActionProposePeace)
	return priority
}

func aiDiplomacyOfferPriority(gs *state.GameState, from, to faction.FactionID, action diplomacy.Action) int {
	priority, _ := aiDiplomacyOfferPriorityDetails(gs, from, to, action)
	return priority
}

func aiDiplomacyOfferPriorityDetails(gs *state.GameState, from, to faction.FactionID, action diplomacy.Action) (int, string) {
	if gs == nil || from == "" || to == "" {
		return 0, ""
	}
	fromFaction := gs.Factions[from]
	toFaction := gs.Factions[to]
	if fromFaction == nil || toFaction == nil {
		return 0, ""
	}

	score := 0
	reasons := make([]string, 0, 5)
	if rel := diplomacy.Relation(gs, from, to); rel != nil {
		relScore := minInt(20, maxInt(0, -rel.Score/4))
		score += relScore
		if relScore > 0 {
			reasons = append(reasons, "ilişki baskısı")
		}
	}

	fromTech := len(fromFaction.Research.Completed)
	toTech := len(toFaction.Research.Completed)
	if toTech > fromTech {
		techScore := minInt(20, (toTech-fromTech)*3)
		score += techScore
		if techScore > 0 {
			reasons = append(reasons, "teknoloji farkı")
		}
	} else {
		score += minInt(6, (fromTech-toTech)*2)
	}

	fromPower := diplomacy.MilitaryPower(gs, from)
	toPower := diplomacy.MilitaryPower(gs, to)
	if toPower > fromPower {
		powerScore := minInt(24, (toPower-fromPower)/10)
		score += powerScore
		if powerScore > 0 {
			reasons = append(reasons, "askeri baskı")
		}
	} else {
		score += minInt(8, (fromPower-toPower)/20)
	}

	if len(gs.LandRegionsOwnedBy(to)) > len(gs.LandRegionsOwnedBy(from)) {
		score += minInt(10, (len(gs.LandRegionsOwnedBy(to))-len(gs.LandRegionsOwnedBy(from)))*2)
	}

	switch action {
	case diplomacy.ActionProposePeace:
		if diplomacy.HasDirectThreat(gs, from, to) {
			score += 12
			reasons = append(reasons, "doğrudan tehdit")
		}
		score += minInt(8, len(gs.RegionsOwnedBy(from))/8)
	case diplomacy.ActionProposeAlliance:
		assessment := diplomacy.AssessAllianceProposal(gs, diplomacy.EnsureRelation(gs, from, to), from, to)
		score += assessment.Chance
		if gs.ScenarioID == "1300_ottoman_rise" {
			strategic := assessment.ActorStrategic
			score += minInt(30, strategic.Score)
			if strategic.BufferValue > 0 {
				reasons = append(reasons, "tampon devlet")
			}
			if strategic.FrontSupportValue > 0 {
				reasons = append(reasons, "cephe desteği")
			}
			if strategic.TradeValue >= 8 {
				reasons = append(reasons, "ticaret değeri")
			}
		}
		if diplomacy.HasCommonEnemy(gs, from, to) {
			reasons = append(reasons, "ortak düşman")
		}
		if diplomacy.HasSharedMajorThreat(gs, from, to) {
			reasons = append(reasons, "ortak büyük tehdit")
		}
		if diplomacy.HasDirectThreat(gs, from, to) {
			reasons = append(reasons, "sınır gerilimi")
		} else {
			reasons = append(reasons, "güvenli diplomasi")
		}
		score += minInt(10, len(gs.RegionsOwnedBy(from))/10)
	case diplomacy.ActionProposeTrade:
		if assessment := diplomacy.AssessTradeProposal(gs, diplomacy.Relation(gs, from, to), from, to); assessment.BlockReason == "" {
			score += assessment.Chance
			reasons = append(reasons, "ticaret fırsatı")
		} else {
			score += 8
		}
		if diplomacy.HasCommonEnemy(gs, from, to) {
			score += 4
		}
		if diplomacy.HasDirectThreat(gs, from, to) {
			score -= 10
		}
		score += minInt(8, aiTradePartnerCount(gs, from)*2)
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "genel öncelik")
	}
	if len(reasons) > 3 {
		reasons = reasons[:3]
	}
	return score, strings.Join(reasons, ", ")
}

func aiEvaluateWarOpportunities(gs *state.GameState, fid faction.FactionID) {
	aiEvaluateWarOpportunitiesWithSteps(gs, fid, nil)
}

func aiShouldAttemptAllianceOffer(gs *state.GameState, from, to faction.FactionID, assessment diplomacy.AllianceProposalAssessment) bool {
	if gs == nil || assessment.BlockReason != "" {
		return false
	}
	if gs.ScenarioID == "1300_ottoman_rise" {
		strategic := assessment.ActorStrategic
		if strategic.ActiveObjectiveConflict || strategic.Score < 18 || assessment.Chance < 45 {
			return false
		}
		urgent := strategic.ThreatValue >= 18 || strategic.BufferValue >= 10
		if aiAllianceSoftCap(gs, from) <= aiActiveAllianceCount(gs, from) && !urgent {
			return false
		}
		if urgent && assessment.Chance >= 60 {
			return true
		}
		threshold := assessment.Chance + minInt(18, maxInt(0, strategic.Score-18)/2)
		threshold = maxInt(25, minInt(92, threshold))
		return aiDiplomacyOfferRoll(gs, from, to, diplomacy.ActionProposeAlliance) < threshold
	}
	commonEnemy := diplomacy.HasCommonEnemy(gs, from, to)
	sharedThreat := diplomacy.HasSharedMajorThreat(gs, from, to)
	if !aiAllianceHasMeaningfulBenefit(gs, from, to) {
		return false
	}
	if aiAllianceSoftCap(gs, from) <= aiActiveAllianceCount(gs, from) && !commonEnemy && !sharedThreat {
		return false
	}
	if aiAllianceExpansionTension(gs, from, to) && !commonEnemy && !sharedThreat {
		return false
	}
	if assessment.Chance >= 72 {
		return true
	}
	if assessment.Chance >= 60 && (commonEnemy || sharedThreat) {
		return true
	}
	if !commonEnemy && !sharedThreat && assessment.Chance < 78 {
		return false
	}

	threshold := assessment.Chance
	if f := gs.Factions[from]; f != nil {
		threshold += (f.AIAggressiveness - 45) / 3
	}
	if to == gs.PlayerFactionID {
		threshold += 8
	}
	if commonEnemy {
		threshold += 6
	}
	if sharedThreat {
		threshold += 8
	}
	threshold = maxInt(22, minInt(92, threshold))
	return aiDiplomacyOfferRoll(gs, from, to, diplomacy.ActionProposeAlliance) < threshold
}

func aiAllianceHasMeaningfulBenefit(gs *state.GameState, actor, target faction.FactionID) bool {
	if gs == nil || actor == "" || target == "" || actor == target {
		return false
	}
	score := aiAllianceBenefitScore(gs, actor, target)
	targetRegions := len(gs.LandRegionsOwnedBy(target))
	commonEnemy := diplomacy.HasCommonEnemy(gs, actor, target)
	sharedThreat := diplomacy.HasSharedMajorThreat(gs, actor, target)
	threshold := 12
	if targetRegions <= 1 {
		threshold = 18
	}
	if commonEnemy || sharedThreat {
		threshold -= 10
	}
	if threshold < 6 {
		threshold = 6
	}
	return score >= threshold
}

func aiAllianceBenefitScore(gs *state.GameState, actor, target faction.FactionID) int {
	if gs == nil {
		return 0
	}
	actorPower := diplomacy.MilitaryPower(gs, actor)
	targetPower := diplomacy.MilitaryPower(gs, target)
	actorRegions := len(gs.LandRegionsOwnedBy(actor))
	targetRegions := len(gs.LandRegionsOwnedBy(target))
	score := 0

	if targetPower > 0 {
		score += minInt(18, targetPower/8)
	}
	if targetRegions > 0 {
		score += minInt(12, targetRegions*3)
	}
	if aiSharesLandBorder(gs, actor, target) {
		score += 4
		frontierSupport := aiFrontierPower(gs, target, actor)
		if frontierSupport > 0 {
			score += minInt(12, frontierSupport/10+4)
		}
	}
	if diplomacy.CanEstablishTradeRoute(gs, actor, target) {
		score += 4
	}
	if diplomacy.HasCommonEnemy(gs, actor, target) {
		score += 10
	}
	if diplomacy.HasSharedMajorThreat(gs, actor, target) {
		score += 12
	}

	if targetRegions <= 1 {
		score -= 8
		if actorRegions >= 4 {
			score -= 6
		}
	}
	if actorRegions > maxInt(3, targetRegions*3) {
		score -= 6
	}
	if targetPower == 0 {
		score -= 8
	} else if actorPower > maxInt(targetPower*3, targetPower+80) {
		score -= 10
	}

	return score
}

func aiShouldCancelAlliance(gs *state.GameState, from, to faction.FactionID) bool {
	if gs == nil || from == "" || to == "" || diplomacy.SameRealm(gs, from, to) {
		return false
	}
	rel := diplomacy.Relation(gs, from, to)
	if rel == nil || rel.Stance != faction.StanceAllied {
		return false
	}
	if gs.ScenarioID == "1300_ottoman_rise" {
		strategic := diplomacy.AssessStrategicAlliance(gs, from, to)
		if strategic.ActiveObjectiveConflict {
			return true
		}
		if strategic.ExpansionTensionPenalty > 0 && strategic.ThreatValue == 0 && strategic.Score < 22 {
			return true
		}
		if strategic.Score < strategicAllianceRetentionFloor(gs, from) && rel.Score <= 50 {
			return true
		}
		if aiActiveAllianceCount(gs, from) > aiAllianceSoftCap(gs, from) && strategic.ThreatValue == 0 && strategic.Score < 24 {
			return true
		}
		return false
	}
	commonEnemy := diplomacy.HasCommonEnemy(gs, from, to)
	sharedThreat := diplomacy.HasSharedMajorThreat(gs, from, to)
	if commonEnemy || sharedThreat {
		return false
	}
	if !aiAllianceHasMeaningfulBenefit(gs, from, to) && rel.Score <= 45 {
		return true
	}
	hasTrade := diplomacy.HasTradeRouteBetween(gs, from, to)
	hasBorder := aiSharesLandBorder(gs, from, to)
	expansionTension := aiAllianceExpansionTension(gs, from, to)
	directThreat := diplomacy.HasDirectThreat(gs, from, to)
	if !hasBorder && !hasTrade && rel.Score <= 35 {
		return true
	}
	if expansionTension && rel.Score <= 40 {
		return true
	}
	if directThreat && rel.Score <= 45 {
		return true
	}
	if aiActiveAllianceCount(gs, from) > aiAllianceSoftCap(gs, from) && rel.Score <= 40 {
		return true
	}
	return false
}

func strategicAllianceRetentionFloor(gs *state.GameState, fid faction.FactionID) int {
	floor := 10
	if gs == nil {
		return floor
	}
	if len(gs.LandRegionsOwnedBy(fid)) >= 8 {
		floor = 14
	}
	return floor
}

func aiDiplomacyOfferRoll(gs *state.GameState, from, to faction.FactionID, action diplomacy.Action) int {
	hasher := fnv.New32a()
	_, _ = fmt.Fprintf(hasher, "%d|%s|%s|%s", gs.Turn, from, to, action)
	return int(hasher.Sum32() % 100)
}

// aiDiplomacyOfferRetryAllowed reddedilmiş bir teklifin tekrarını üç tur
// bekletir. Bekleme bitince her tur deterministik bir zar atılır; böylece aynı
// teklif koşullar aynı kalsa bile otomatik olarak her tur gönderilmez.
func aiDiplomacyOfferRetryAllowed(gs *state.GameState, from, to faction.FactionID, action diplomacy.Action) bool {
	if gs == nil {
		return false
	}
	key := state.DiplomaticOfferRejectionKey(string(from), string(to), string(action))
	lastRejected, rejectedBefore := gs.OfferRejectionTurns[key]
	if !rejectedBefore {
		return true
	}
	if gs.Turn-lastRejected < state.DiplomaticOfferRetryCooldownTurns {
		return false
	}
	const retryChance = 35
	return aiDiplomacyOfferRoll(gs, from, to, action) < retryChance
}

func aiDiplomacyOfferRetryAllowedForRegion(gs *state.GameState, from, to faction.FactionID, action diplomacy.Action, regionID world.RegionID) bool {
	if gs == nil || regionID == "" {
		return false
	}
	if gs.DiplomaticOfferRegionRetryBlocked(string(from), string(to), string(action), regionID, state.DiplomaticOfferRetryCooldownTurns) {
		return false
	}
	key := state.DiplomaticOfferRegionRejectionKey(string(from), string(to), string(action), regionID)
	_, rejectedBefore := gs.OfferRejectionTurns[key]
	if !rejectedBefore {
		return true
	}
	const retryChance = 35
	return aiDiplomacyOfferRoll(gs, from, to, action) < retryChance
}

func aiWarCadenceAllows(gs *state.GameState, fid faction.FactionID) bool {
	if gs == nil || gs.Turn == 0 {
		return true
	}
	interval := aiWarCadenceBase(gs)
	if f := gs.Factions[fid]; f != nil {
		if len(f.AIExpansionTargets) > 0 {
			interval -= 2
		}
		if f.AIAggressiveness >= 65 {
			interval -= 2
		}
	}
	if interval < 4 {
		interval = 4
	}
	offset := 0
	for _, ch := range string(fid) {
		offset += int(ch)
	}
	return (gs.Turn+offset)%interval == 0
}

func aiHasExpansionTarget(self *faction.Faction, target faction.FactionID) bool {
	if self == nil {
		return false
	}
	for _, targetID := range self.AIExpansionTargets {
		if targetID == target {
			return true
		}
	}
	return false
}

func aiAllianceExpansionTension(gs *state.GameState, a, b faction.FactionID) bool {
	if gs == nil {
		return false
	}
	return aiHasExpansionTarget(gs.Factions[a], b) || aiHasExpansionTarget(gs.Factions[b], a)
}

func aiAllianceSoftCap(gs *state.GameState, fid faction.FactionID) int {
	cap := 1
	if gs == nil {
		return cap
	}
	landRegions := len(gs.LandRegionsOwnedBy(fid))
	switch {
	case landRegions >= 10:
		cap = 3
	case landRegions >= 4:
		cap = 2
	}
	if f := gs.Factions[fid]; f != nil && f.AIAggressiveness >= 70 && cap < 3 {
		cap++
	}
	return cap
}

func aiActiveAllianceCount(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" {
		return 0
	}
	count := 0
	for _, rel := range gs.Relations {
		if rel == nil || rel.Stance != faction.StanceAllied {
			continue
		}
		if diplomacy.SameRealm(gs, rel.FactionA, rel.FactionB) {
			continue
		}
		switch fid {
		case rel.FactionA:
			count++
		case rel.FactionB:
			count++
		}
	}
	return count
}
func aiMaxConcurrentWars(gs *state.GameState, fid faction.FactionID) int {
	if configured, ok := aiConfiguredWarCapacity(gs); ok {
		return configured
	}
	limit := 1
	if gs != nil && gs.Difficulty >= 3 {
		limit = 2
	}
	if gs != nil {
		if f := gs.Factions[fid]; f != nil && f.AIAggressiveness >= 65 {
			limit++
		}
	}
	return limit
}

func aiActiveWarCount(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil {
		return 0
	}
	seen := make(map[faction.FactionID]struct{})
	root := fid
	if realmRoot := diplomacy.RealmRoot(gs, fid); realmRoot != "" {
		root = realmRoot
	}
	for _, rel := range gs.Relations {
		if rel == nil || rel.Stance != faction.StanceWar {
			continue
		}
		if !diplomacy.SameRealm(gs, root, rel.FactionA) && !diplomacy.SameRealm(gs, root, rel.FactionB) {
			continue
		}
		other := rel.FactionB
		if diplomacy.SameRealm(gs, root, rel.FactionB) {
			other = rel.FactionA
		}
		otherRoot := diplomacy.RealmRoot(gs, other)
		if otherRoot == "" {
			otherRoot = other
		}
		seen[otherRoot] = struct{}{}
	}
	return len(seen)
}

func aiSharesLandBorder(gs *state.GameState, a, b faction.FactionID) bool {
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(a) {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor != nil && !neighbor.IsSea && neighbor.OwnerID == string(b) {
				return true
			}
		}
	}
	return false
}

func aiFrontierPower(gs *state.GameState, owner, against faction.FactionID) int {
	total := 0
	for _, armyRef := range gs.Armies {
		if armyRef == nil || armyRef.IsNaval || armyRef.OwnerID != string(owner) {
			continue
		}
		region := gs.Regions[armyRef.RegionID]
		if region == nil || region.IsSea {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor != nil && !neighbor.IsSea && neighbor.OwnerID == string(against) {
				total += armyRef.TotalStrength(gs.UnitTypes)
				break
			}
		}
	}
	return total
}

func aiBestBorderTargetValue(gs *state.GameState, actor, target faction.FactionID) int {
	best := 0
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(actor) {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil || neighbor.IsSea || neighbor.OwnerID != string(target) {
				continue
			}
			value := aiRegionStrategicValue(gs, neighbor)
			if value > best {
				best = value
			}
		}
	}
	return best
}

func aiRegionStrategicValue(gs *state.GameState, region *world.Region) int {
	if gs == nil || region == nil {
		return 0
	}
	prod := gs.RegionProductionSummary(region)
	return prod.Gold + prod.Grain + prod.Iron + prod.Timber + prod.Stone + prod.Spice*2 + prod.Cloth*2
}

func aiEnqueueProduction(gs *state.GameState, fid faction.FactionID, kind string, rid world.RegionID, typeID string, turns int) state.ProductionOrder {
	if turns < 1 {
		turns = 1
	}
	gs.NextProductionSeq++
	order := state.ProductionOrder{
		ID:        fmt.Sprintf("prod_%d", gs.NextProductionSeq),
		Kind:      kind,
		FactionID: string(fid),
		RegionID:  rid,
		TypeID:    typeID,
		TurnsLeft: turns,
	}
	gs.ProductionQueue = append(gs.ProductionQueue, order)
	return order
}

func aiQueuedBuildingCount(gs *state.GameState, rid world.RegionID, buildingID string, fid faction.FactionID) int {
	count := 0
	for _, order := range gs.ProductionQueue {
		if order.Kind == aiProductionKindBuilding && order.RegionID == rid && order.TypeID == buildingID && order.FactionID == string(fid) {
			count++
		}
	}
	return count
}

func aiPendingLandUnitCount(gs *state.GameState, fid faction.FactionID) int {
	count := 0
	for _, order := range gs.ProductionQueue {
		if order.Kind != aiProductionKindUnit || order.FactionID != string(fid) {
			continue
		}
		if utype, ok := gs.UnitTypes[order.TypeID]; ok && utype.RequiredBldg != "port" {
			count++
		}
	}
	return count
}

func aiPendingUnitCountByRegion(gs *state.GameState, rid world.RegionID, fid faction.FactionID) int {
	count := 0
	for _, order := range gs.ProductionQueue {
		if order.Kind == aiProductionKindUnit && order.RegionID == rid && order.FactionID == string(fid) {
			count++
		}
	}
	return count
}

func aiProductionLane(unitType *army.UnitType) string {
	if unitType == nil {
		return "barracks"
	}
	if unitType.RequiredBldg == "port" {
		return "port"
	}
	switch unitType.Category {
	case army.CategoryNavalWar, army.CategoryNavalTrans, army.CategoryNavalTrade:
		return "port"
	}
	return "barracks"
}

func aiPendingUnitCountByRegionInLane(gs *state.GameState, rid world.RegionID, fid faction.FactionID, lane string) int {
	count := 0
	for _, order := range gs.ProductionQueue {
		if order.Kind != aiProductionKindUnit || order.RegionID != rid || order.FactionID != string(fid) {
			continue
		}
		if aiProductionLane(gs.UnitTypes[order.TypeID]) != lane {
			continue
		}
		count++
	}
	return count
}

func aiLaneRemainingCapacity(gs *state.GameState, rid world.RegionID, fid faction.FactionID, unitType *army.UnitType) int {
	if gs == nil {
		return 0
	}
	region := gs.Regions[rid]
	capacity := state.LandUnitProductionLimit(region)
	if aiProductionLane(unitType) == "port" {
		capacity = state.NavalUnitProductionLimit(region)
	}
	if capacity <= 0 {
		return 0
	}
	pending := aiPendingUnitCountByRegionInLane(gs, rid, fid, aiProductionLane(unitType))
	remaining := capacity - pending
	if remaining < 0 {
		return 0
	}
	return remaining
}

func aiPendingNavalUnitCount(gs *state.GameState, seaRegion world.RegionID, fid faction.FactionID) int {
	count := 0
	for _, order := range gs.ProductionQueue {
		if order.Kind != aiProductionKindUnit || order.FactionID != string(fid) {
			continue
		}
		utype, ok := gs.UnitTypes[order.TypeID]
		if !ok || utype.RequiredBldg != "port" {
			continue
		}
		region := gs.Regions[order.RegionID]
		if region == nil {
			continue
		}
		for _, nid := range region.Neighbors {
			if nid == seaRegion {
				count++
				break
			}
		}
	}
	return count
}

func aiPendingNavalFleetCount(gs *state.GameState, fid faction.FactionID) int {
	seenSeaRegions := make(map[world.RegionID]struct{})
	for _, order := range gs.ProductionQueue {
		if order.Kind != aiProductionKindUnit || order.FactionID != string(fid) {
			continue
		}
		utype, ok := gs.UnitTypes[order.TypeID]
		if ok && aiProductionLane(utype) == "port" {
			region := gs.Regions[order.RegionID]
			if region == nil {
				continue
			}
			seaRegion := aiSeaNeighbor(gs, region)
			if seaRegion == "" {
				continue
			}
			hasExistingFleet := false
			for _, a := range gs.Armies {
				if a.OwnerID == string(fid) && a.IsDocked() && a.DockedRegionID == region.ID {
					hasExistingFleet = true
					break
				}
			}
			if hasExistingFleet {
				continue
			}
			seenSeaRegions[seaRegion] = struct{}{}
		}
	}
	return len(seenSeaRegions)
}

func aiPendingTransportOrderCount(gs *state.GameState, fid faction.FactionID) int {
	count := 0
	for _, order := range gs.ProductionQueue {
		if order.Kind != aiProductionKindUnit || order.FactionID != string(fid) {
			continue
		}
		utype, ok := gs.UnitTypes[order.TypeID]
		if !ok || utype.Category != army.CategoryNavalTrans {
			continue
		}
		count++
	}
	return count
}

type aiEscortFrontCandidate struct {
	region        *world.Region
	seaID         world.RegionID
	pressure      int
	score         int
	hasEnemyFleet bool
}

func aiEscortFrontCandidates(gs *state.GameState, fid faction.FactionID, coastalRegions []*world.Region, warshipType *army.UnitType) []aiEscortFrontCandidate {
	if gs == nil || warshipType == nil {
		return nil
	}
	candidates := make(map[world.RegionID]aiEscortFrontCandidate, len(coastalRegions))
	for _, r := range coastalRegions {
		if r == nil || !aiRegionHasPortBuilding(r) {
			continue
		}
		if aiBuildingLevel(r, "port") < warshipType.RequiredBldgLevel {
			continue
		}
		seaID := aiSeaNeighbor(gs, r)
		if seaID == "" {
			continue
		}
		pressure := aiSeaPressure(gs, string(fid), seaID)
		enemyFleet := aiEnemyNavalInRegion(gs, string(fid), seaID) != nil
		score := pressure
		if enemyFleet {
			score += 20
		}
		if pending := aiPendingNavalUnitCount(gs, seaID, fid); pending > 0 {
			score += pending * 2
		}
		current, exists := candidates[seaID]
		if exists {
			if current.score > score {
				continue
			}
			if current.score == score {
				if current.pressure > pressure || (current.pressure == pressure && current.region != nil && string(current.region.ID) <= string(r.ID)) {
					continue
				}
			}
		}
		candidates[seaID] = aiEscortFrontCandidate{
			region:        r,
			seaID:         seaID,
			pressure:      pressure,
			score:         score,
			hasEnemyFleet: enemyFleet,
		}
	}
	out := make([]aiEscortFrontCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, candidate)
	}
	return out
}

func aiEscortThreatFrontCount(candidates []aiEscortFrontCandidate) int {
	count := 0
	for _, candidate := range candidates {
		if candidate.pressure >= 25 || candidate.hasEnemyFleet {
			count++
		}
	}
	return count
}

func aiEscortLimit(atWar bool, candidates []aiEscortFrontCandidate) int {
	limit := 1
	if atWar {
		limit++
	}
	if aiEscortThreatFrontCount(candidates) >= 2 {
		limit++
	}
	maxPressure := 0
	for _, candidate := range candidates {
		if candidate.pressure > maxPressure {
			maxPressure = candidate.pressure
		}
	}
	if maxPressure >= 60 {
		limit++
	}
	if limit > 3 {
		limit = 3
	}
	return limit
}

func aiBuildingLevel(region *world.Region, buildingID string) int {
	if region == nil || buildingID == "" {
		return 0
	}
	level := 0
	for _, bid := range region.Buildings {
		if bid == buildingID {
			level++
		}
	}
	return level
}

func aiBuildingTurnsRequired(region *world.Region, buildingID string, baseTurns, queued int) int {
	turns := baseTurns + aiBuildingLevel(region, buildingID) + queued
	if turns < 1 {
		return 1
	}
	return turns
}

func aiBuildingAllowed(gs *state.GameState, region *world.Region, buildingID, requiredTerrain string) bool {
	if gs == nil || region == nil || region.IsSea || region.IsLocked {
		return false
	}
	if buildingID == "port" {
		return region.IsCoastal(gs.Regions)
	}
	return requiredTerrain == "" || string(region.Terrain) == requiredTerrain
}

func aiTradePartnerCount(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" {
		return 0
	}
	partners := make(map[string]struct{})
	self := string(fid)
	for _, route := range gs.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 {
			continue
		}
		switch {
		case route.FromFactionID == self && route.ToFactionID != "":
			partners[route.ToFactionID] = struct{}{}
		case route.ToFactionID == self && route.FromFactionID != "":
			partners[route.FromFactionID] = struct{}{}
		}
	}
	return len(partners)
}

func aiFindRecruitRegion(gs *state.GameState, fid faction.FactionID, utype *army.UnitType) world.RegionID {
	if gs == nil || utype == nil {
		return ""
	}
	requiredBuilding := utype.RequiredBldg
	if requiredBuilding == "" {
		requiredBuilding = "barracks"
	}
	requiredLevel := utype.RequiredBldgLevel
	if requiredLevel <= 0 {
		requiredLevel = 1
	}
	bestRemaining := -1
	bestLevel := -1
	bestRegionID := world.RegionID("")
	for _, r := range aiSortedRegions(gs) {
		if r == nil || r.OwnerID != string(fid) || r.IsSea || r.IsLocked {
			continue
		}
		if aiBuildingLevel(r, requiredBuilding) < requiredLevel {
			continue
		}
		if aiPendingUnitCountByRegion(gs, r.ID, fid) >= aiMaxRegionQueue {
			continue
		}
		if !aiCanQueueLandUnit(gs, fid, r.ID, utype) {
			continue
		}
		remaining := aiLaneRemainingCapacity(gs, r.ID, fid, utype)
		if remaining <= 0 {
			continue
		}
		level := aiBuildingLevel(r, requiredBuilding)
		if remaining > bestRemaining || (remaining == bestRemaining && (level > bestLevel || (level == bestLevel && (bestRegionID == "" || string(r.ID) < string(bestRegionID))))) {
			bestRemaining = remaining
			bestLevel = level
			bestRegionID = r.ID
		}
	}
	return bestRegionID
}

func aiCanQueueLandUnit(gs *state.GameState, fid faction.FactionID, rid world.RegionID, unitType *army.UnitType) bool {
	pendingInRegion := aiPendingUnitCountByRegionInLane(gs, rid, fid, aiProductionLane(unitType))
	for _, a := range aiSortedArmies(gs) {
		if a == nil || a.RegionID != rid || a.OwnerID != string(fid) || a.IsNaval || a.IsGarrison {
			continue
		}
		return len(a.Units)+pendingInRegion < army.MaxArmySize
	}
	return gs.CurrentLandArmies(fid) < gs.MaxLandArmies(fid)
}

// aiSelectBestUnit altın ve teknoloji durumuna göre en uygun birim tipini seçer.
// Öncelik: piyade > süvari > milis. Topçu sadece zengin AI'ler için.
func aiSelectBestUnit(gs *state.GameState, f *faction.Faction) string {
	return aiSelectBestUnitForBudget(gs, f, nil)
}

func aiSelectBestUnitForBudget(gs *state.GameState, f *faction.Faction, budget *aiBudget) string {
	return aiSelectBestUnitForStrategicContext(gs, f, budget, nil)
}

func aiSelectBestUnitForStrategicContext(gs *state.GameState, f *faction.Faction, budget *aiBudget, strategicContext *StrategicContext) string {
	if gs != nil && gs.ScenarioID == "1300_ottoman_rise" {
		return aiSelectStrategicLandUnit(gs, f, budget, strategicContext)
	}
	return aiSelectLegacyLandUnit(gs, f, budget)
}

func aiSelectLegacyLandUnit(gs *state.GameState, f *faction.Faction, budget *aiBudget) string {
	// Askeri güç istatistiği
	armyCount := 0
	cavalryCount := 0
	for _, a := range gs.Armies {
		if a.OwnerID == string(f.ID) && !a.IsNaval {
			armyCount++
			for _, u := range a.Units {
				if ut, ok := gs.UnitTypes[u.TypeID]; ok && ut.Category == "cavalry" {
					cavalryCount++
				}
			}
		}
	}

	// Tier 3 elite piyade (seçkin piyade) - çok zenginse ve teknolojisi varsa
	if aiUnitBudgetThresholdMet(f, budget, 350) {
		if ut, ok := gs.UnitTypes["elite_infantry"]; ok {
			if aiUnitAvailableForBudget(gs, f, ut, budget) {
				return "elite_infantry"
			}
		}
	}

	// Ağır süvari - zengin ve teknolojisi varsa
	if aiUnitBudgetThresholdMet(f, budget, 450) && cavalryCount < armyCount*2 {
		if ut, ok := gs.UnitTypes["heavy_cavalry"]; ok {
			if aiUnitAvailableForBudget(gs, f, ut, budget) {
				return "heavy_cavalry"
			}
		}
	}

	// Tier 2 piyade (normal piyade) - orta düzey altın ve teknoloji
	if aiUnitBudgetThresholdMet(f, budget, 180) {
		if ut, ok := gs.UnitTypes["infantry"]; ok {
			if aiUnitAvailableForBudget(gs, f, ut, budget) {
				return "infantry"
			}
		}
	}

	// Süvari - teknolojisi varsa ve altın yeterliyse
	if aiUnitBudgetThresholdMet(f, budget, 300) && cavalryCount < armyCount*3 {
		if ut, ok := gs.UnitTypes["cavalry"]; ok {
			if aiUnitAvailableForBudget(gs, f, ut, budget) {
				return "cavalry"
			}
		}
	}

	// Hafif süvari - her zaman uygun
	if aiUnitBudgetThresholdMet(f, budget, 200) && cavalryCount < armyCount*4 {
		if ut, ok := gs.UnitTypes["light_cavalry"]; ok && aiUnitAvailableForBudget(gs, f, ut, budget) {
			return "light_cavalry"
		}
	}

	// Topçu - çok zenginse ve savaşta ise
	if aiUnitBudgetThresholdMet(f, budget, 650) {
		// Savaş halinde mi kontrol et
		atWar := false
		for _, rel := range gs.Relations {
			if (rel.FactionA == f.ID || rel.FactionB == f.ID) && rel.Stance == faction.StanceWar {
				atWar = true
				break
			}
		}
		if atWar {
			if ut, ok := gs.UnitTypes["cannon"]; ok {
				if aiUnitAvailableForBudget(gs, f, ut, budget) {
					return "cannon"
				}
			}
			if ut, ok := gs.UnitTypes["bombard"]; ok {
				if aiUnitAvailableForBudget(gs, f, ut, budget) {
					return "bombard"
				}
			}
		}
	}

	// Varsayılan: milis
	if ut, ok := gs.UnitTypes["militia"]; ok && aiUnitAvailableForBudget(gs, f, ut, budget) {
		return "militia"
	}
	return ""
}

func aiUnitAvailableForRecruitment(gs *state.GameState, f *faction.Faction, utype *army.UnitType) bool {
	return aiUnitAvailableForBudget(gs, f, utype, nil)
}

func aiUnitAvailableForBudget(gs *state.GameState, f *faction.Faction, utype *army.UnitType, budget *aiBudget) bool {
	if gs == nil || f == nil || utype == nil {
		return false
	}
	if !utype.HasAllRequiredTechs(f.Research.Completed) {
		return false
	}
	cost := economy.ResourceCost{
		Gold:   utype.GoldCost,
		Grain:  utype.GrainCost,
		Iron:   utype.IronCost,
		Timber: utype.TimberCost,
		Stone:  utype.StoneCost,
		Spice:  utype.SpiceCost,
		Cloth:  utype.ClothCost,
	}
	if !aiCanAffordForBudget(f, cost, budget, aiBudgetArmy) {
		return false
	}
	return aiFindRecruitRegion(gs, f.ID, utype) != ""
}

func aiUnitBudgetThresholdMet(f *faction.Faction, budget *aiBudget, gold int) bool {
	if budget == nil {
		return f != nil && f.Gold >= gold+aiMinGoldReserve
	}
	return aiCanAffordForBudget(f, economy.ResourceCost{Gold: gold}, budget, aiBudgetArmy)
}

// FormCoalitionAgainstPlayer oyuncu tehdit eşiğini geçmişse diğer AI fraksiyonlarla ittifak kurar.
func FormCoalitionAgainstPlayer(gs *state.GameState, fid faction.FactionID) {
	formCoalitionAgainstPlayer(gs, fid, nil)
}

func formCoalitionAgainstPlayer(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	playerRegions := len(gs.RegionsOwnedBy(gs.PlayerFactionID))
	if playerRegions < coalitionThreshold {
		return
	}

	result := diplomacy.Execute(gs, fid, gs.PlayerFactionID, diplomacy.ActionDeclareWar)
	if result.Applied || result.Accepted {
		addTurnStep(steps, TurnStep{
			FactionID:     fid,
			Kind:          TurnStepDiplomacy,
			TargetFaction: gs.PlayerFactionID,
			Message:       turnFactionName(gs, fid) + ": " + result.Message,
		})
	}

	// Diğer AI fraksiyonlarla ittifak kur (düşman değillerse)
	for _, otherFID := range aiSortedFactionIDs(gs) {
		if otherFID == fid || otherFID == gs.PlayerFactionID {
			continue
		}
		if gs.Factions[otherFID].IsEliminated {
			continue
		}
		rel := diplomacy.EnsureRelation(gs, fid, otherFID)
		if rel.Stance == faction.StanceWar {
			continue
		}
		if rel.Score < 20 {
			rel.Score = 20
		}
		result := diplomacy.Execute(gs, fid, otherFID, diplomacy.ActionProposeAlliance)
		if result.Applied || result.Accepted {
			addTurnStep(steps, TurnStep{
				FactionID:     fid,
				Kind:          TurnStepDiplomacy,
				TargetFaction: otherFID,
				Message:       turnFactionName(gs, fid) + ": " + result.Message,
			})
		}
	}
}

// moveArmy tek bir orduyu hareket puanı tükenene kadar hareket ettirir.
func moveArmy(gs *state.GameState, a *army.Army) {
	moveArmyWithSteps(gs, a, faction.FactionID(a.OwnerID), nil)
}

func moveArmyWithSteps(gs *state.GameState, a *army.Army, fid faction.FactionID, steps *[]TurnStep) {
	moveArmyWithStrategicContext(gs, a, fid, steps, nil)
}

func moveArmyWithStrategicContext(gs *state.GameState, a *army.Army, fid faction.FactionID, steps *[]TurnStep, strategicContext *StrategicContext) {
	if step, withdrew := executeStrategicSiegeWithdrawal(gs, a, fid, strategicContext); withdrew {
		addTurnStep(steps, step)
	}
	for a.MovePoints > 0 {
		target := chooseBestMoveWithStrategicContext(gs, a, strategicContext)
		if target == "" {
			break
		}
		movePointsBefore := a.MovePoints

		// Escort mantığı: transport filosu hareket edecekse, önce aynı bölgedeki escort savaş gemisi gitsin
		if a.IsNaval && a.TransportCapacity(gs.UnitTypes) > 0 {
			aiEscortMoveFirst(gs, a, target, fid, steps)
		}

		outcome := executeMove(gs, a, target, fid)
		if outcome.step.Message != "" {
			addTurnStep(steps, outcome.step)
		}
		if !outcome.survived {
			break
		}
		// Seçim ve uygulama kuralları geçici olarak ayrışsa bile aynı hedefin
		// hareket puanı harcanmadan sonsuza kadar yeniden seçilmesine izin verme.
		if a.MovePoints >= movePointsBefore {
			a.MovePoints = 0
			break
		}
	}
}

// aiEscortMoveFirst transport filosu hareket etmeden önce aynı bölgedeki
// escort savaş gemisini hedef deniz bölgesine gönderir.
func aiEscortMoveFirst(gs *state.GameState, transport *army.Army, target world.RegionID, fid faction.FactionID, steps *[]TurnStep) {
	if gs == nil || transport == nil || target == "" {
		return
	}

	// Sadece denizden denize hareketlerde escort mantığı çalışır
	targetRegion := gs.Regions[target]
	if targetRegion == nil || !targetRegion.IsSea {
		return
	}

	// Aynı bölgede escort savaş gemisi bul
	var escort *army.Army
	for _, a := range aiSortedArmies(gs) {
		if a.ID == transport.ID || !a.IsAtSea() || a.OwnerID != transport.OwnerID || a.RegionID != transport.RegionID {
			continue
		}
		if isWarshipFleet(a, gs.UnitTypes) && a.MovePoints > 0 {
			escort = a
			break
		}
	}
	if escort == nil {
		return
	}

	// Hedef deniz bölgesinde düşman filosu var mı?
	hasEnemy := false
	for _, ea := range gs.Armies {
		if ea.RegionID == target && ea.OwnerID != transport.OwnerID && ea.IsAtSea() {
			_, stance := relationScore(gs, transport.OwnerID, ea.OwnerID)
			if stance == faction.StanceWar {
				hasEnemy = true
				break
			}
		}
	}

	// Hedef deniz bölgesi savaş baskısı altındaysa escort'u gönder
	if !hasEnemy {
		pressure := aiSeaPressure(gs, string(fid), target)
		if pressure < 25 {
			return
		}
	}

	// Escort'u önden gönder
	actorName := turnFactionName(gs, fid)
	sourceName := turnRegionName(gs, escort.RegionID)
	targetName := turnRegionName(gs, target)

	escort.RegionID = target
	escort.DockedRegionID = ""
	escort.DockedSettlementID = ""
	escort.MovePoints--

	// Hedefte düşman filosu varsa çatış
	enemyInTarget := aiEnemyNavalInRegion(gs, transport.OwnerID, target)
	if enemyInTarget != nil {
		_, stance := relationScore(gs, transport.OwnerID, enemyInTarget.OwnerID)
		if stance == faction.StanceWar {
			atkMods := aiTechMods(gs, escort.OwnerID)
			defMods := aiTechMods(gs, enemyInTarget.OwnerID)
			seaTerrain := string(world.TerrainSea)
			result := combat.ResolveBattleWithContextPlan(escort, enemyInTarget, world.TerrainType(seaTerrain), gs.UnitTypes, atkMods, defMods, combat.BattleContextNaval, combat.BattleStanceAggressive)
			gs.RecordWarCasualties(faction.FactionID(escort.OwnerID), faction.FactionID(enemyInTarget.OwnerID), result.AttackerLost, result.DefenderLost)
			recordCommanderBattle(gs, escort, enemyInTarget, nil, result.AttackerWins)

			if result.AttackerWins {
				if len(enemyInTarget.Units) == 0 {
					gs.RemoveArmy(enemyInTarget.ID)
					addTurnStep(steps, TurnStep{
						FactionID:    fid,
						Kind:         TurnStepBattle,
						ArmyID:       escort.ID,
						FromRegion:   escort.RegionID,
						TargetRegion: target,
						FocusRegion:  target,
						Message:      actorName + " escort savaş gemisi " + targetName + " bölgesindeki düşman filosunu yok etti.",
					})
				}
			} else {
				if len(escort.Units) == 0 {
					gs.RemoveArmy(escort.ID)
				}
				addTurnStep(steps, TurnStep{
					FactionID:    fid,
					Kind:         TurnStepBattle,
					ArmyID:       escort.ID,
					FromRegion:   escort.RegionID,
					TargetRegion: target,
					FocusRegion:  target,
					Message:      actorName + " escort savaş gemisi " + targetName + " bölgesinde geri püskürtüldü.",
				})
			}
			return
		}
	}

	addTurnStep(steps, TurnStep{
		FactionID:    fid,
		Kind:         TurnStepMove,
		ArmyID:       escort.ID,
		FromRegion:   escort.RegionID,
		TargetRegion: target,
		FocusRegion:  target,
		Message:      actorName + " escort savaş gemisi " + sourceName + " bölgesinden " + targetName + " bölgesine keşfe çıktı.",
	})
}

// aiEnemyNavalInRegion hedef deniz bölgesindeki düşman donanmasını bulur.
func aiEnemyNavalInRegion(gs *state.GameState, ownerID string, seaRegionID world.RegionID) *army.Army {
	if gs == nil {
		return nil
	}
	for _, a := range aiSortedArmies(gs) {
		if a == nil || !a.IsAtSea() || a.OwnerID == ownerID || a.RegionID != seaRegionID {
			continue
		}
		return a
	}
	return nil
}

type moveOutcome struct {
	survived bool
	step     TurnStep
}

func aiCanEmbarkArmy(gs *state.GameState, a *army.Army) bool {
	if gs == nil || a == nil {
		return false
	}
	return a.CanEmbark(gs.UnitTypes)
}

func aiFindEmbarkFleet(gs *state.GameState, ownerID string, seaRegionID world.RegionID, unitCount int) *army.Army {
	for _, candidate := range aiSortedArmies(gs) {
		if candidate.OwnerID != ownerID || !candidate.IsNaval || candidate.RegionID != seaRegionID {
			continue
		}
		if candidate.CanEmbarkUnits(gs.UnitTypes, unitCount) {
			return candidate
		}
	}
	return nil
}

func aiFactionAtWar(gs *state.GameState, ownerID string) bool {
	if gs == nil {
		return false
	}
	for _, rel := range gs.Relations {
		if rel == nil || rel.Stance != faction.StanceWar {
			continue
		}
		if string(rel.FactionA) == ownerID || string(rel.FactionB) == ownerID {
			return true
		}
	}
	return false
}

func aiSeaPressure(gs *state.GameState, ownerID string, seaRegionID world.RegionID) int {
	if gs == nil {
		return 0
	}
	sea, ok := gs.Regions[seaRegionID]
	if !ok || sea == nil || !sea.IsSea {
		return 0
	}
	score := 0
	hostileCoasts := 0
	for _, nid := range sea.Neighbors {
		land, ok := gs.Regions[nid]
		if !ok || land.IsSea {
			continue
		}
		switch {
		case land.OwnerID == "":
			score += 8
		case land.OwnerID == ownerID:
			score += 2
		default:
			_, stance := relationScore(gs, ownerID, land.OwnerID)
			if stance == faction.StanceWar {
				score += 28
				hostileCoasts++
				if enemyArmy := aiEnemyArmyInRegion(gs, ownerID, land.ID); enemyArmy == nil {
					score += 8
				}
			} else {
				score -= 2
			}
		}
	}

	friendlyFleets := 0
	for _, a := range gs.Armies {
		if a == nil || a.OwnerID != ownerID || !a.IsAtSea() || a.RegionID != seaRegionID {
			continue
		}
		friendlyFleets++
		if len(a.EmbarkedUnits) > 0 {
			score += 6
		}
	}
	if hostileCoasts > 0 && friendlyFleets == 0 {
		score += 12
	}
	if friendlyFleets > 1 {
		score -= (friendlyFleets - 1) * 10
	}
	return score
}

func aiCanDisembarkToLand(gs *state.GameState, fleet *army.Army, target *world.Region) bool {
	if gs == nil || fleet == nil || target == nil || !fleet.IsNaval || len(fleet.EmbarkedUnits) == 0 {
		return false
	}
	if target.OwnerID == "" || target.OwnerID == fleet.OwnerID {
		return true
	}
	_, stance := relationScore(gs, fleet.OwnerID, target.OwnerID)
	return stance == faction.StanceWar || stance == faction.StanceAllied
}

func aiLandingStrength(gs *state.GameState, fleet *army.Army) int {
	if gs == nil || fleet == nil || len(fleet.EmbarkedUnits) == 0 {
		return 0
	}
	tmp := &army.Army{OwnerID: fleet.OwnerID, Units: fleet.EmbarkedUnits}
	return tmp.TotalStrength(gs.UnitTypes)
}

func aiEnemyArmyInRegion(gs *state.GameState, ownerID string, rid world.RegionID) *army.Army {
	for _, ea := range aiSortedArmies(gs) {
		if ea.RegionID == rid && ea.OwnerID != ownerID {
			return ea
		}
	}
	return nil
}

func aiSpawnDisembarkedArmy(gs *state.GameState, ownerID string, target world.RegionID, units []army.Unit) *army.Army {
	if gs == nil || len(units) == 0 {
		return nil
	}
	gs.NextArmySeq++
	newID := army.ArmyID(fmt.Sprintf("army_%s_%d", ownerID, gs.NextArmySeq))
	landed := &army.Army{
		ID:            newID,
		OwnerID:       ownerID,
		RegionID:      target,
		Units:         units,
		MovePoints:    0,
		MaxMovePoints: 2,
		IsNaval:       false,
	}
	gs.Armies[newID] = landed
	return landed
}

func aiOwnerReligion(gs *state.GameState, ownerID string) string {
	if gs == nil {
		return ""
	}
	f, ok := gs.Factions[faction.FactionID(ownerID)]
	if !ok {
		return ""
	}
	return string(f.Religion)
}

// executeMove hareketi ve varsa savaşı uygular.
// Ordu hayatta kaldıysa true, yok edildiyse false döner.
func executeMove(gs *state.GameState, a *army.Army, target world.RegionID, fid faction.FactionID) moveOutcome {
	targetRegion, ok := gs.Regions[target]
	if !ok {
		return moveOutcome{survived: true}
	}
	fromRegion := a.RegionID
	actorName := turnFactionName(gs, fid)
	targetName := turnRegionName(gs, target)
	sourceName := turnRegionName(gs, fromRegion)
	if a.IsNaval && a.IsDocked() && targetRegion.IsSea {
		// Docked filonun RegionID'si rota hesapları için deniz ankrajıdır;
		// gerçek konum ancak denize çıkış emriyle tekrar deniz bölgesi olur.
		a.DockedRegionID = ""
		a.DockedSettlementID = ""
	}

	if !a.IsNaval {
		if siege := gs.SiegeAt(target); siege != nil && siege.AttackerArmyID != a.ID {
			if !gs.CanJoinActiveSiege(a, target) && !gs.CanEnterActiveSiegedRegion(a, target) {
				return moveOutcome{survived: true}
			}
		}
	}

	// Kuşatma altındaki ordu hareket edemez; önce huruç savaşı yapmalı.
	// Eğer kuşatan oyuncu ise sortie step'i döner (battle plan UI için).
	if !a.IsNaval {
		if siege := gs.SiegeAt(fromRegion); siege != nil && siege.AttackerArmyID != a.ID && gs.IsArmyDefendingSiegedRegion(a) {
			siegeArmy := gs.Armies[siege.AttackerArmyID]
			sourceRegion := gs.Regions[fromRegion]
			if siegeArmy != nil && sourceRegion != nil {
				// AI vs AI sortie (veya oyuncu kuşatıyorsa): hemen çöz
				atkMods := aiTechMods(gs, a.OwnerID)
				defMods := aiTechMods(gs, siegeArmy.OwnerID)
				defMods.DefenseMod += 0.10
				result := combat.ResolveBattleWithMods(a, siegeArmy, sourceRegion.Terrain, gs.UnitTypes, atkMods, defMods)
				gs.RecordWarCasualties(faction.FactionID(a.OwnerID), faction.FactionID(siegeArmy.OwnerID), result.AttackerLost, result.DefenderLost)
				recordCommanderBattle(gs, a, siegeArmy, nil, result.AttackerWins)
				if result.AttackerWins {
					if len(siegeArmy.Units) == 0 {
						gs.RemoveArmy(siegeArmy.ID)
					}
					delete(gs.Sieges, fromRegion)
					canExitToTarget := targetRegion.OwnerID == "" || targetRegion.OwnerID == a.OwnerID || diplomacy.SameRealm(gs, faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID))
					if !canExitToTarget && targetRegion.OwnerID != "" {
						_, stance := relationScore(gs, a.OwnerID, targetRegion.OwnerID)
						canExitToTarget = stance == faction.StanceAllied
					}
					if canExitToTarget && len(a.Units) > 0 {
						a.RegionID = target
						a.MovePoints = maxInt(0, a.MovePoints-1)
					}
					msg := actorName + " " + sourceName + " kuşatmasını yardı ve çıktı."
					return moveOutcome{survived: len(a.Units) > 0, step: TurnStep{FactionID: fid, Kind: TurnStepBattle, ArmyID: a.ID, FromRegion: fromRegion, TargetRegion: target, FocusRegion: fromRegion, Message: msg}}
				}
				a.MovePoints = 0
				if len(a.Units) == 0 {
					gs.RemoveArmy(a.ID)
				}
				if len(siegeArmy.Units) == 0 {
					gs.RemoveArmy(siegeArmy.ID)
					delete(gs.Sieges, fromRegion)
				}
				msg := actorName + " " + sourceName + " kuşatmasını yaramadı."
				return moveOutcome{survived: len(a.Units) > 0, step: TurnStep{FactionID: fid, Kind: TurnStepBattle, ArmyID: a.ID, FromRegion: fromRegion, TargetRegion: target, FocusRegion: fromRegion, Message: msg}}
			}
		}
	}

	if a.IsNaval && targetRegion.CanLandEnter() {
		if !aiCanDisembarkToLand(gs, a, targetRegion) {
			return moveOutcome{survived: true}
		}
		enemyArmy := aiEnemyArmyInRegion(gs, a.OwnerID, target)
		isAlliedDisembark := false
		if targetRegion.OwnerID != "" && targetRegion.OwnerID != a.OwnerID {
			if diplomacy.SameRealm(gs, faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID)) {
				isAlliedDisembark = true
			} else if rel := diplomacy.Relation(gs, faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID)); rel != nil && rel.Stance == faction.StanceAllied {
				isAlliedDisembark = true
			}
		}
		if !isAlliedDisembark && targetRegion.OwnerID != "" && targetRegion.OwnerID != a.OwnerID && targetRegion.IsFortified() {
			units := make([]army.Unit, len(a.EmbarkedUnits))
			copy(units, a.EmbarkedUnits)
			a.EmbarkedUnits = a.EmbarkedUnits[:0]
			landed := aiSpawnDisembarkedArmy(gs, a.OwnerID, target, units)
			if landed != nil {
				gs.MoveEmbarkedCommanderToArmy(a.ID, landed.ID)
			}
			a.MovePoints--
			if landed != nil {
				if active := gs.SiegeAt(target); active != nil {
					if gs.CanJoinActiveSiege(landed, target) {
						landed.MovePoints = 0
						return moveOutcome{survived: true, step: TurnStep{FactionID: fid, Kind: TurnStepDisembark, ArmyID: a.ID, FromRegion: fromRegion, TargetRegion: target, FocusRegion: target, Message: actorName + " " + targetName + " kuşatmasına denizden destek verdi."}}
					}
					return moveOutcome{survived: true, step: TurnStep{FactionID: fid, Kind: TurnStepDisembark, ArmyID: a.ID, FromRegion: fromRegion, TargetRegion: target, FocusRegion: target, Message: actorName + " " + targetName + " kıyısına çıktı; kale zaten kuşatma altında."}}
				}
				if aiCanStartSiege(gs, landed, targetRegion) {
					aiStartSiege(gs, landed, targetRegion, enemyArmy)
					return moveOutcome{survived: true, step: TurnStep{FactionID: fid, Kind: TurnStepDisembark, ArmyID: a.ID, FromRegion: fromRegion, TargetRegion: target, FocusRegion: target, Message: actorName + " " + targetName + " kıyısına çıktı ve kuşatma başlattı."}}
				}
			}
			return moveOutcome{survived: true, step: TurnStep{FactionID: fid, Kind: TurnStepDisembark, ArmyID: a.ID, FromRegion: fromRegion, TargetRegion: target, FocusRegion: target, Message: actorName + " " + targetName + " kıyısına çıktı."}}
		}
		if enemyArmy != nil {
			landing := &army.Army{
				OwnerID:   a.OwnerID,
				Units:     append([]army.Unit(nil), a.EmbarkedUnits...),
				Commander: gs.AmphibiousCommander(a.ID),
			}
			atkMods := aiTechMods(gs, a.OwnerID)
			defMods := aiTechMods(gs, enemyArmy.OwnerID)
			result := combat.ResolveBattleWithMods(landing, enemyArmy, targetRegion.Terrain, gs.UnitTypes, atkMods, defMods)
			gs.RecordWarCasualties(faction.FactionID(landing.OwnerID), faction.FactionID(enemyArmy.OwnerID), result.AttackerLost, result.DefenderLost)
			recordCommanderBattle(gs, landing, enemyArmy, nil, result.AttackerWins)
			a.EmbarkedUnits = a.EmbarkedUnits[:0]
			a.MovePoints--
			if result.AttackerWins {
				if len(enemyArmy.Units) == 0 {
					gs.RemoveArmy(enemyArmy.ID)
				}
				landed := aiSpawnDisembarkedArmy(gs, a.OwnerID, target, landing.Units)
				if landed != nil {
					gs.MoveEmbarkedCommanderToArmy(a.ID, landed.ID)
				}
				vassalized := TryResolvePostWarVassalization(gs, faction.FactionID(a.OwnerID), targetRegion).Applied
				if !vassalized {
					aiApplyConquest(gs, targetRegion, a.OwnerID)
				}
				message := actorName + " " + targetName + " kıyısına çıkarma yapıp bölgeyi aldı."
				if vassalized {
					message = actorName + " " + targetName + " kıyısındaki zaferden sonra devleti vassal bıraktı."
				}
				return moveOutcome{
					survived: true,
					step: TurnStep{
						FactionID:    fid,
						Kind:         TurnStepBattle,
						ArmyID:       a.ID,
						FromRegion:   fromRegion,
						TargetRegion: target,
						FocusRegion:  target,
						Message:      message,
					},
				}
			}
			gs.ReleaseEmbarkedCommander(a.ID)
			return moveOutcome{
				survived: true,
				step: TurnStep{
					FactionID:    fid,
					Kind:         TurnStepBattle,
					ArmyID:       a.ID,
					FromRegion:   fromRegion,
					TargetRegion: target,
					FocusRegion:  target,
					Message:      actorName + " " + targetName + " kıyısındaki çıkarmada geri püskürtüldü.",
				},
			}
		}
		units := make([]army.Unit, len(a.EmbarkedUnits))
		copy(units, a.EmbarkedUnits)
		a.EmbarkedUnits = a.EmbarkedUnits[:0]
		landed := aiSpawnDisembarkedArmy(gs, a.OwnerID, target, units)
		if landed != nil {
			gs.MoveEmbarkedCommanderToArmy(a.ID, landed.ID)
		}
		stepKind := TurnStepDisembark
		msg := actorName + " " + targetName + " kıyısına çıkarma yaptı."
		isAlliedDisembark = false
		if targetRegion.OwnerID != a.OwnerID && targetRegion.OwnerID != "" {
			key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID))
			if rel, exists := gs.Relations[key]; exists && rel.Stance == faction.StanceAllied {
				isAlliedDisembark = true
			}
		}
		if targetRegion.OwnerID != a.OwnerID && !isAlliedDisembark {
			vassalized := TryResolvePostWarVassalization(gs, faction.FactionID(a.OwnerID), targetRegion).Applied
			if !vassalized {
				aiApplyConquest(gs, targetRegion, a.OwnerID)
			}
			stepKind = TurnStepConquest
			if vassalized {
				msg = actorName + " " + targetName + " kıyısına çıktı ve teslim olan devleti vassal bıraktı."
			} else {
				msg = actorName + " " + targetName + " kıyısına çıktı ve bölgeyi ele geçirdi."
			}
		}
		a.MovePoints--
		return moveOutcome{
			survived: true,
			step: TurnStep{
				FactionID:    fid,
				Kind:         stepKind,
				ArmyID:       a.ID,
				FromRegion:   fromRegion,
				TargetRegion: target,
				FocusRegion:  target,
				Message:      msg,
			},
		}
	}
	if !a.IsNaval && targetRegion.IsSea {
		if !aiCanEmbarkArmy(gs, a) {
			return moveOutcome{survived: true}
		}
		fleet := aiFindEmbarkFleet(gs, a.OwnerID, target, len(a.Units))
		if fleet == nil {
			return moveOutcome{survived: true}
		}
		fleet.EmbarkedUnits = append(fleet.EmbarkedUnits, a.Units...)
		if fleet.MovePoints > 0 {
			fleet.MovePoints--
		}
		gs.MoveCommanderIntoFleet(a.ID, fleet.ID)
		gs.RemoveArmy(a.ID)
		return moveOutcome{
			survived: false,
			step: TurnStep{
				FactionID:    fid,
				Kind:         TurnStepEmbark,
				ArmyID:       a.ID,
				FromRegion:   fromRegion,
				TargetRegion: target,
				FocusRegion:  fromRegion,
				Message:      actorName + " " + sourceName + " bölgesinden filoya bindi.",
			},
		}
	}

	if aiCanStartSiege(gs, a, targetRegion) {
		activeSiege := gs.SiegeAt(target)
		if activeSiege == nil {
			defender := aiEnemyArmyInRegion(gs, a.OwnerID, target)
			aiStartSiege(gs, a, targetRegion, defender)
			return moveOutcome{
				survived: true,
				step: TurnStep{
					FactionID:    fid,
					Kind:         TurnStepBattle,
					ArmyID:       a.ID,
					FromRegion:   fromRegion,
					TargetRegion: target,
					FocusRegion:  target,
					Message:      actorName + " " + targetName + " tahkimatını kuşatmaya aldı.",
				},
			}
		}
		if activeSiege.AttackerArmyID == a.ID {
			defender := aiEnemyArmyInRegion(gs, a.OwnerID, target)
			virtualDefense := false
			if defender == nil {
				defender = aiVirtualSiegeGarrison(gs, targetRegion)
				virtualDefense = true
			}
			atkMods := aiTechMods(gs, a.OwnerID)
			defMods := aiTechMods(gs, targetRegion.OwnerID)
			defMods.DefenseMod += aiSiegeDefenseBonus(activeSiege.FortLevel, activeSiege.BreachLevel)
			result := combat.ResolveBattleWithContextPlan(a, defender, targetRegion.Terrain, gs.UnitTypes, atkMods, defMods, combat.BattleContextLand, combat.BattleStanceBalanced)
			gs.RecordWarCasualties(faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID), result.AttackerLost, result.DefenderLost)
			recordCommanderBattle(gs, a, defender, nil, result.AttackerWins)
			if result.AttackerWins {
				if !virtualDefense && len(defender.Units) == 0 {
					gs.RemoveArmy(defender.ID)
				}
				if len(a.Units) > 0 {
					a.RegionID = target
					a.DockedRegionID = ""
					a.DockedSettlementID = ""
					a.MovePoints--
					vassalized := TryResolvePostWarVassalization(gs, faction.FactionID(a.OwnerID), targetRegion).Applied
					if !vassalized {
						aiApplyConquest(gs, targetRegion, a.OwnerID)
					}
					aiClearSiege(gs, target)
					message := actorName + " " + targetName + " tahkimatına genel hücumla girdi ve kazandı."
					if vassalized {
						message = actorName + " " + targetName + " tahkimatını düşürdü ve devleti vassal bıraktı."
					}
					return moveOutcome{
						survived: true,
						step: TurnStep{
							FactionID:    fid,
							Kind:         TurnStepBattle,
							ArmyID:       a.ID,
							FromRegion:   fromRegion,
							TargetRegion: target,
							FocusRegion:  target,
							Message:      message,
						},
					}
				}
				gs.RemoveArmy(a.ID)
				aiClearSiegesByArmy(gs, a.ID)
				return moveOutcome{
					survived: false,
					step: TurnStep{
						FactionID:    fid,
						Kind:         TurnStepBattle,
						ArmyID:       a.ID,
						FromRegion:   fromRegion,
						TargetRegion: target,
						FocusRegion:  target,
						Message:      actorName + " " + targetName + " kuşatmasını kazansa da ordusu dağıldı.",
					},
				}
			}
			if len(a.Units) == 0 {
				gs.RemoveArmy(a.ID)
				aiClearSiegesByArmy(gs, a.ID)
			}
			a.MovePoints = 0
			return moveOutcome{
				survived: len(a.Units) > 0,
				step: TurnStep{
					FactionID:    fid,
					Kind:         TurnStepBattle,
					ArmyID:       a.ID,
					FromRegion:   fromRegion,
					TargetRegion: target,
					FocusRegion:  target,
					Message:      actorName + " " + targetName + " tahkimatına yaptığı genel hücumda geri püskürtüldü.",
				},
			}
		}
	}

	// Hedefte düşman ordusu var mı? (müttefikler dahil birleşik savunma)
	combinedDef, defSourceIDs := gs.CollectDefenders(a, target, a.IsNaval && targetRegion.IsSea)
	var enemyArmy *army.Army
	if combinedDef == nil {
		for _, ea := range aiSortedArmies(gs) {
			if ea.RegionID == target && ea.OwnerID != a.OwnerID && (!a.IsNaval || !targetRegion.IsSea || ea.IsAtSea()) {
				enemyArmy = ea
				break
			}
		}
	} else {
		// Birleşik ordudan refakat için ilk orduyu bul
		for _, ea := range aiSortedArmies(gs) {
			if ea.RegionID == target && ea.OwnerID != a.OwnerID && (!a.IsNaval || !targetRegion.IsSea || ea.IsAtSea()) {
				enemyArmy = ea
				break
			}
		}
	}

	if combinedDef != nil || enemyArmy != nil {
		var defForBattle *army.Army
		if combinedDef != nil {
			defForBattle = combinedDef
		} else {
			defForBattle = enemyArmy
		}
		atkMods := aiTechMods(gs, a.OwnerID)
		defOwnerID := a.OwnerID
		if enemyArmy != nil {
			defOwnerID = enemyArmy.OwnerID
		}
		defMods := aiTechMods(gs, defOwnerID)
		result := combat.ResolveBattleWithMods(a, defForBattle, targetRegion.Terrain, gs.UnitTypes, atkMods, defMods)
		gs.RecordWarCasualties(faction.FactionID(a.OwnerID), faction.FactionID(defOwnerID), result.AttackerLost, result.DefenderLost)
		recordCommanderBattle(gs, a, defForBattle, defSourceIDs, result.AttackerWins)
		if result.AttackerWins {
			if len(defSourceIDs) > 0 {
				gs.DistributeDefenderLosses(defSourceIDs, result.DefenderLost)
			} else if enemyArmy != nil && len(enemyArmy.Units) == 0 {
				gs.RemoveArmy(enemyArmy.ID)
			}
			battleLiftsSiege := false
			if targetSiege := gs.SiegeAt(target); targetSiege != nil && targetSiege.AttackerArmyID != a.ID {
				for _, sid := range defSourceIDs {
					if sid == targetSiege.AttackerArmyID {
						battleLiftsSiege = true
						break
					}
				}
				if !battleLiftsSiege && enemyArmy != nil && enemyArmy.ID == targetSiege.AttackerArmyID {
					battleLiftsSiege = true
				}
			}
			isAlliedTarget := false
			if targetRegion.OwnerID != "" && targetRegion.OwnerID != a.OwnerID {
				if diplomacy.SameRealm(gs, faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID)) {
					isAlliedTarget = true
				}
				key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID))
				if rel, exists := gs.Relations[key]; exists && rel.Stance == faction.StanceAllied {
					isAlliedTarget = true
				}
			}
			if battleLiftsSiege {
				aiClearSiege(gs, target)
			}
			if len(a.Units) > 0 {
				a.RegionID = target
				a.DockedRegionID = ""
				a.DockedSettlementID = ""
				vassalized := false
				if !isAlliedTarget {
					vassalized = TryResolvePostWarVassalization(gs, faction.FactionID(a.OwnerID), targetRegion).Applied
					if !vassalized {
						aiApplyConquest(gs, targetRegion, a.OwnerID)
					}
				}
				a.MovePoints--
				message := actorName + " " + targetName + " bölgesindeki savaşı kazandı."
				if isAlliedTarget && battleLiftsSiege {
					message = actorName + " " + targetName + " bölgesindeki savaşı kazandı ve kuşatmayı kaldırdı."
				} else if vassalized {
					message = actorName + " " + targetName + " bölgesindeki savaşı kazandı ve devleti vassal bıraktı."
				}
				return moveOutcome{
					survived: true,
					step: TurnStep{
						FactionID:    fid,
						Kind:         TurnStepBattle,
						ArmyID:       a.ID,
						FromRegion:   fromRegion,
						TargetRegion: target,
						FocusRegion:  target,
						Message:      message,
					},
				}
			}
			gs.RemoveArmy(a.ID)
			return moveOutcome{
				survived: false,
				step: TurnStep{
					FactionID:    fid,
					Kind:         TurnStepBattle,
					ArmyID:       a.ID,
					FromRegion:   fromRegion,
					TargetRegion: target,
					FocusRegion:  target,
					Message:      actorName + " " + targetName + " savaşını kazansa da ordusu dağıldı.",
				},
			}
		}
		// Saldıran yenildi
		if len(a.Units) == 0 {
			gs.RemoveArmy(a.ID)
		}
		return moveOutcome{
			survived: false,
			step: TurnStep{
				FactionID:    fid,
				Kind:         TurnStepBattle,
				ArmyID:       a.ID,
				FromRegion:   fromRegion,
				TargetRegion: target,
				FocusRegion:  target,
				Message:      actorName + " " + targetName + " saldırısında yenildi.",
			},
		}
	}

	// Savaşsız hareket
	a.RegionID = target
	a.DockedRegionID = ""
	a.DockedSettlementID = ""
	a.MovePoints--
	stepKind := TurnStepMove
	msg := actorName + " " + sourceName + " bölgesinden " + targetName + " bölgesine ilerledi."
	isAlliedTarget := false
	if targetRegion.OwnerID != a.OwnerID {
		key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID))
		if rel, exists := gs.Relations[key]; exists && rel.Stance == faction.StanceAllied {
			isAlliedTarget = true
		}
	}
	if targetRegion.OwnerID != a.OwnerID && !isAlliedTarget {
		vassalized := TryResolvePostWarVassalization(gs, faction.FactionID(a.OwnerID), targetRegion).Applied
		if !vassalized {
			aiApplyConquest(gs, targetRegion, a.OwnerID)
		}
		stepKind = TurnStepConquest
		if vassalized {
			msg = actorName + " " + targetName + " bölgesinde teslim olan devleti vassal bıraktı."
		} else {
			msg = actorName + " " + targetName + " bölgesini savaşsız ele geçirdi."
		}
	}

	// Konsolidasyon (Dost orduyla birleşme)
	if tryMergeAIArmies(gs, a) {
		return moveOutcome{
			survived: false,
			step: TurnStep{
				FactionID:    fid,
				Kind:         stepKind,
				ArmyID:       a.ID,
				FromRegion:   fromRegion,
				TargetRegion: target,
				FocusRegion:  target,
				Message:      msg,
			},
		}
	}

	return moveOutcome{
		survived: true,
		step: TurnStep{
			FactionID:    fid,
			Kind:         stepKind,
			ArmyID:       a.ID,
			FromRegion:   fromRegion,
			TargetRegion: target,
			FocusRegion:  target,
			Message:      msg,
		},
	}
}

func aiNavalStrategyWithStrategicContextAndSteps(gs *state.GameState, fid faction.FactionID, budget *aiBudget, strategicContext *StrategicContext, steps *[]TurnStep) {
	if gs != nil && gs.ScenarioID == "1300_ottoman_rise" {
		aiExecuteNavalMissionProduction(gs, fid, budget, strategicContext, steps)
		aiExecuteMerchantTradeStrategy(gs, fid, budget, strategicContext, steps)
		return
	}
	f := gs.Factions[fid]
	if f.IsEliminated || gs.BuildingTypes == nil || gs.UnitTypes == nil {
		return
	}

	// Kıyı bölgesi var mı?
	var coastalRegions []*world.Region
	for _, r := range aiSortedRegions(gs) {
		if r.OwnerID == string(fid) && !r.IsSea && r.IsCoastal(gs.Regions) {
			coastalRegions = append(coastalRegions, r)
		}
	}
	if len(coastalRegions) == 0 {
		return
	}

	// Liman tipi var mı?
	portType, hasPort := gs.BuildingTypes["port"]
	if !hasPort {
		return
	}
	transportType, hasTransport := gs.UnitTypes["transport"]
	if !hasTransport {
		return
	}

	// Liman inşası (en az bir liman olsun)
	for _, r := range coastalRegions {
		queued := aiQueuedBuildingCount(gs, r.ID, "port", fid)
		portCost := economy.ResourceCost{
			Gold:   portType.GoldCost,
			Grain:  portType.GrainCost,
			Iron:   portType.IronCost,
			Timber: portType.TimberCost,
			Stone:  portType.StoneCost,
			Spice:  portType.SpiceCost,
			Cloth:  portType.ClothCost,
		}
		if aiBuildingLevel(r, "port")+queued < portType.MaxPerRegion &&
			aiBuildingAllowed(gs, r, "port", portType.RequiredTerrain) &&
			aiCanAffordForBudget(f, portCost, budget, aiBudgetNaval) {
			if !aiApplyBudgetedCost(f, portCost, budget, aiBudgetNaval) {
				continue
			}
			turns := aiBuildingTurnsRequired(r, "port", portType.TurnsRequired, queued)
			aiEnqueueProduction(gs, fid, aiProductionKindBuilding, r.ID, "port", turns)
			addTurnStep(steps, TurnStep{
				FactionID:    fid,
				Kind:         TurnStepBuild,
				TargetRegion: r.ID,
				FocusRegion:  r.ID,
				Message:      turnFactionName(gs, fid) + " " + turnRegionName(gs, r.ID) + " kıyısında liman kuruyor.",
			})
			break // Bir liman yeter bu tur
		}
	}

	// Gemi alımı (liman olan bölgelerden)
	fleetLimit := 1
	if len(coastalRegions) >= 3 {
		fleetLimit++
	}
	if aiFactionAtWar(gs, string(fid)) {
		fleetLimit++
	}
	if fleetLimit > 3 {
		fleetLimit = 3
	}
	fleetCount := 0
	for _, a := range gs.Armies {
		if a.OwnerID == string(fid) && a.IsNaval {
			fleetCount++
		}
	}
	fleetCount += aiPendingNavalFleetCount(gs, fid)
	if fleetCount >= fleetLimit {
		return
	}

	bestScore := -1
	var bestRegion *world.Region
	var bestSeaRegion world.RegionID
	for _, r := range coastalRegions {
		// Liman var mı?
		hasPortBldg := false
		for _, bid := range r.Buildings {
			if bid == "port" {
				hasPortBldg = true
				break
			}
		}
		if !hasPortBldg {
			continue
		}

		// Komşu deniz bölgesi bul
		var seaRegion world.RegionID
		for _, nid := range r.Neighbors {
			if n, ok := gs.Regions[nid]; ok && n.IsSea {
				seaRegion = nid
				break
			}
		}
		if seaRegion == "" {
			continue
		}
		if aiPendingUnitCountByRegion(gs, r.ID, fid) >= aiMaxRegionQueue {
			continue
		}
		if aiLaneRemainingCapacity(gs, r.ID, fid, transportType) <= 0 {
			continue
		}
		currentUnits := 0
		for _, a := range aiSortedArmies(gs) {
			if a.RegionID == seaRegion && a.OwnerID == string(fid) && a.IsNaval {
				currentUnits = len(a.Units)
				break
			}
		}
		if currentUnits+aiPendingNavalUnitCount(gs, seaRegion, fid) >= army.MaxArmySize {
			continue
		}
		score := aiSeaPressure(gs, string(fid), seaRegion)
		if score > bestScore {
			bestScore = score
			bestRegion = r
			bestSeaRegion = seaRegion
		}
	}
	if bestRegion == nil || bestSeaRegion == "" {
		return
	}

	// Altın kontrolü
	shipCost := economy.ResourceCost{
		Gold:   transportType.GoldCost,
		Grain:  transportType.GrainCost,
		Iron:   transportType.IronCost,
		Timber: transportType.TimberCost,
		Stone:  transportType.StoneCost,
		Spice:  transportType.SpiceCost,
		Cloth:  transportType.ClothCost,
	}
	if !aiCanAffordForBudget(f, shipCost, budget, aiBudgetNaval) {
		return
	}

	if !aiApplyBudgetedCost(f, shipCost, budget, aiBudgetNaval) {
		return
	}
	aiEnqueueProduction(gs, fid, aiProductionKindUnit, bestRegion.ID, "transport", transportType.TurnsRequired)
	addTurnStep(steps, TurnStep{
		FactionID:    fid,
		Kind:         TurnStepRecruit,
		TargetRegion: bestRegion.ID,
		FocusRegion:  bestSeaRegion,
		Message:      turnFactionName(gs, fid) + " " + turnRegionName(gs, bestRegion.ID) + " limanında nakliye gemisi hazırlıyor.",
	})

	// Escort savaş gemisi üretimi — transport varsa ve savaş halinde veya deniz baskısı yüksekse
	aiProduceEscortIfNeeded(gs, fid, coastalRegions, budget, steps)
}

// aiProduceEscortIfNeeded transport hattı olan AI için escort savaş gemisi üretir.
func aiProduceEscortIfNeeded(gs *state.GameState, fid faction.FactionID, coastalRegions []*world.Region, budget *aiBudget, steps *[]TurnStep) {
	f := gs.Factions[fid]
	if f == nil || f.IsEliminated || gs.UnitTypes == nil {
		return
	}

	// warship birimi var mı?
	warshipType, hasWarship := gs.UnitTypes["warship"]
	if !hasWarship {
		return
	}

	// AI'ın transport filosu var mı?
	hasTransport := false
	atWar := aiFactionAtWar(gs, string(fid))
	highSeaPressure := false
	for _, a := range gs.Armies {
		if a.OwnerID == string(fid) && a.IsNaval && a.TransportCapacity(gs.UnitTypes) > 0 {
			hasTransport = true
			seaRegion := gs.Regions[a.RegionID]
			if seaRegion != nil && seaRegion.IsSea && aiSeaPressure(gs, string(fid), a.RegionID) >= 30 {
				highSeaPressure = true
			}
		}
	}
	if !hasTransport && aiPendingTransportOrderCount(gs, fid) > 0 {
		hasTransport = true
	}
	if !hasTransport {
		return
	}

	candidates := aiEscortFrontCandidates(gs, fid, coastalRegions, warshipType)
	if len(candidates) == 0 {
		return
	}

	// Escort gerekiyor mu?
	needEscort := atWar || highSeaPressure || aiEscortThreatFrontCount(candidates) >= 2
	if !needEscort {
		return
	}

	// Escort filosu sayısı kontrolü
	escortLimit := aiEscortLimit(atWar, candidates)
	escortFleetCount := 0
	for _, a := range gs.Armies {
		if a.OwnerID == string(fid) && a.IsNaval && isWarshipFleet(a, gs.UnitTypes) {
			escortFleetCount++
		}
	}
	escortFleetCount += aiPendingWarshipOrderCount(gs, fid)
	if escortFleetCount >= escortLimit {
		return
	}

	// Altın ve kaynak kontrolü
	warshipCost := economy.ResourceCost{
		Gold:   warshipType.GoldCost,
		Grain:  warshipType.GrainCost,
		Iron:   warshipType.IronCost,
		Timber: warshipType.TimberCost,
		Stone:  warshipType.StoneCost,
		Spice:  warshipType.SpiceCost,
		Cloth:  warshipType.ClothCost,
	}
	if !aiCanAffordForBudget(f, warshipCost, budget, aiBudgetNaval) {
		return
	}

	// Teknoloji kontrolü
	if !warshipType.HasAllRequiredTechs(f.Research.Completed) {
		return
	}

	usedSeas := make(map[world.RegionID]bool, len(candidates))
	for escortFleetCount < escortLimit {
		candidate, ok := aiBestEscortFrontCandidate(candidates, usedSeas)
		if !ok {
			break
		}
		usedSeas[candidate.seaID] = true
		if candidate.region == nil {
			continue
		}
		if aiPendingUnitCountByRegion(gs, candidate.region.ID, fid) >= aiMaxRegionQueue {
			continue
		}
		if aiLaneRemainingCapacity(gs, candidate.region.ID, fid, warshipType) <= 0 {
			continue
		}
		currentWarshipUnits := 0
		for _, a := range aiSortedArmies(gs) {
			if a.RegionID == candidate.seaID && a.OwnerID == string(fid) && a.IsNaval && isWarshipFleet(a, gs.UnitTypes) {
				currentWarshipUnits = len(a.Units)
				break
			}
		}
		if currentWarshipUnits+aiPendingNavalUnitCount(gs, candidate.seaID, fid) >= army.MaxArmySize {
			continue
		}
		if !aiCanAffordForBudget(f, warshipCost, budget, aiBudgetNaval) {
			break
		}
		if !aiApplyBudgetedCost(f, warshipCost, budget, aiBudgetNaval) {
			break
		}
		aiEnqueueProduction(gs, fid, aiProductionKindUnit, candidate.region.ID, "warship", warshipType.TurnsRequired)
		addTurnStep(steps, TurnStep{
			FactionID:    fid,
			Kind:         TurnStepRecruit,
			TargetRegion: candidate.region.ID,
			FocusRegion:  candidate.seaID,
			Message:      turnFactionName(gs, fid) + " " + turnRegionName(gs, candidate.region.ID) + " limanında savaş gemisi inşa ediyor.",
		})
		escortFleetCount++
	}
}

func aiBestEscortFrontCandidate(candidates []aiEscortFrontCandidate, used map[world.RegionID]bool) (aiEscortFrontCandidate, bool) {
	var (
		best   aiEscortFrontCandidate
		found  bool
		bestID world.RegionID
	)
	for _, candidate := range candidates {
		if used != nil && used[candidate.seaID] {
			continue
		}
		if !found || candidate.score > best.score || (candidate.score == best.score && (candidate.pressure > best.pressure || (candidate.pressure == best.pressure && string(candidate.seaID) < string(bestID)))) {
			best = candidate
			bestID = candidate.seaID
			found = true
		}
	}
	return best, found
}

// isWarshipFleet bir filonun savaş gemisi filosu olup olmadığını kontrol eder.
func isWarshipFleet(a *army.Army, unitTypes map[string]*army.UnitType) bool {
	if a == nil || !a.IsNaval || len(a.Units) == 0 {
		return false
	}
	for _, u := range a.Units {
		ut, ok := unitTypes[u.TypeID]
		if !ok {
			continue
		}
		if ut.Category == army.CategoryNavalWar {
			return true
		}
	}
	return false
}

// aiPendingWarshipOrderCount kuyruktaki savaş gemisi siparişlerini sayar.
func aiPendingWarshipOrderCount(gs *state.GameState, fid faction.FactionID) int {
	count := 0
	for _, o := range gs.ProductionQueue {
		if o.Kind == aiProductionKindUnit && o.FactionID == string(fid) && o.TypeID == "warship" {
			count++
		}
	}
	return count
}

// aiRegionHasPortBuilding bir bölgede liman binası olup olmadığını kontrol eder.
func aiRegionHasPortBuilding(r *world.Region) bool {
	if r == nil {
		return false
	}
	for _, bid := range r.Buildings {
		if bid == "port" {
			return true
		}
	}
	return false
}

// aiSeaNeighbor bir kara bölgesinin komşu deniz bölgesini döner.
func aiSeaNeighbor(gs *state.GameState, r *world.Region) world.RegionID {
	if r == nil {
		return ""
	}
	for _, nid := range r.Neighbors {
		if n, ok := gs.Regions[nid]; ok && n.IsSea {
			return nid
		}
	}
	return ""
}

func aiCanAffordWithReserve(f *faction.Faction, cost economy.ResourceCost) bool {
	if f == nil {
		return false
	}
	if f.Gold-cost.Gold < aiMinGoldReserve {
		return false
	}
	if f.Grain < cost.Grain || f.Iron < cost.Iron || f.Timber < cost.Timber || f.Stone < cost.Stone {
		return false
	}
	return true
}

// aiConsolidateArmies aynı bölgedeki aynı tipteki (kara/deniz) kendi ordularını birleştirir.
func aiConsolidateArmies(gs *state.GameState, fid faction.FactionID) {
	var armies []*army.Army
	for _, a := range aiSortedArmies(gs) {
		if a.OwnerID == string(fid) {
			armies = append(armies, a)
		}
	}

	for i := 0; i < len(armies); i++ {
		a1 := armies[i]
		if _, ok := gs.Armies[a1.ID]; !ok {
			continue
		}
		for j := i + 1; j < len(armies); j++ {
			a2 := armies[j]
			if _, ok := gs.Armies[a2.ID]; !ok {
				continue
			}
			if a1.LocationID() == a2.LocationID() && a1.IsNaval == a2.IsNaval {
				if a1.IsNaval && (a1.TradeRouteKey != "" || a2.TradeRouteKey != "") {
					continue
				}
				region := gs.Regions[a1.RegionID]
				if !aiShouldConsolidateInRegion(gs, region, a1.OwnerID, a1.IsNaval) {
					continue
				}
				if len(a1.Units)+len(a2.Units) <= army.MaxArmySize {
					a1.Units = append(a1.Units, a2.Units...)
					gs.RemoveArmy(a2.ID)
				} else {
					transfer := army.MaxArmySize - len(a1.Units)
					if transfer > 0 {
						a1.Units = append(a1.Units, a2.Units[:transfer]...)
						a2.Units = a2.Units[transfer:]
					}
				}
			}
		}
	}
}

// tryMergeAIArmies hareket sonrası dost bölgede başka dost ordu varsa kapasite dahilinde birleşir.
// Birleşme sonucu ordu tamamen silinirse true döner.
func tryMergeAIArmies(gs *state.GameState, a *army.Army) bool {
	region := gs.Regions[a.RegionID]
	if !aiShouldConsolidateInRegion(gs, region, a.OwnerID, a.IsNaval) {
		return false
	}
	for _, other := range aiSortedArmies(gs) {
		otherID := other.ID
		if otherID == a.ID || other.LocationID() != a.LocationID() || other.OwnerID != a.OwnerID || other.IsNaval != a.IsNaval {
			continue
		}
		if a.IsNaval && (a.TradeRouteKey != "" || other.TradeRouteKey != "") {
			continue
		}
		if len(a.Units)+len(other.Units) <= army.MaxArmySize {
			other.Units = append(other.Units, a.Units...)
			gs.RemoveArmy(a.ID)
			return true
		} else {
			// Kapasite kadar aktar
			transfer := army.MaxArmySize - len(other.Units)
			if transfer > 0 {
				other.Units = append(other.Units, a.Units[:transfer]...)
				a.Units = a.Units[transfer:]
			}
		}
	}
	return false
}

func aiShouldConsolidateInRegion(gs *state.GameState, region *world.Region, ownerID string, isNaval bool) bool {
	if isNaval || region == nil {
		return true
	}
	_, _, overload := aiRegionLogistics(gs, region, ownerID)
	return overload <= 0
}

func aiRegionLogistics(gs *state.GameState, region *world.Region, ownerID string) (demand, capacity, overload int) {
	if gs == nil || region == nil || region.IsSea || ownerID == "" {
		return 0, 0, 0
	}

	militaryProduction := gs.RegionMilitaryGrainProduction(region)
	if region.OwnerID != ownerID {
		militaryProduction = 0
	}
	settlementBuffer := aiRegionSettlementBuffer(gs, region)
	blockadePercent := gs.RegionBlockadePercent(region, ownerID)
	settlementBuffer = settlementBuffer * (100 - blockadePercent) / 100
	reserveSupport := aiRegionReserveSupport(gs, ownerID, militaryProduction, settlementBuffer)
	capacity = militaryProduction + settlementBuffer + reserveSupport
	if capacity < 4 {
		capacity = 4
	}

	for _, candidate := range gs.Armies {
		if candidate == nil || candidate.IsNaval || candidate.OwnerID != ownerID || candidate.RegionID != region.ID {
			continue
		}
		demand += gs.EffectiveArmyGrainUpkeep(candidate)
	}
	overload = maxInt(0, demand-capacity)
	return demand, capacity, overload
}

func aiRegionSettlementBuffer(gs *state.GameState, region *world.Region) int {
	if region == nil {
		return 0
	}
	buffer := 0
	for _, settlement := range region.Settlements {
		switch settlement.Type {
		case world.SettlementCity:
			buffer += 8
		case world.SettlementTown:
			buffer += 5
		case world.SettlementFortress:
			buffer += 6
		case world.SettlementPort:
			buffer += 6
		default:
			buffer += 4
		}
		if settlement.IsCapital {
			buffer += 4
		}
	}
	if gs != nil && gs.IsCapitalRegion(region) {
		buffer += state.CapitalRegionLogisticsBonus
	}
	if tc := region.TradeCapacity / 2; tc > 0 {
		if tc > 6 {
			tc = 6
		}
		buffer += tc
	}
	return buffer
}

func aiRegionReserveSupport(gs *state.GameState, ownerID string, production, settlementBuffer int) int {
	if gs == nil || ownerID == "" {
		return 0
	}
	f := gs.Factions[faction.FactionID(ownerID)]
	if f == nil || f.Grain <= 0 {
		return 0
	}
	cap := production/2 + settlementBuffer/2 + 4
	if cap < 4 {
		cap = 4
	}
	reserve := f.Grain / 10
	if reserve > cap {
		reserve = cap
	}
	return reserve
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func aiPreferredDockSettlementID(region *world.Region) string {
	if region == nil {
		return ""
	}
	for _, settlement := range region.Settlements {
		if settlement.Type == world.SettlementPort {
			return settlement.ID
		}
	}
	if len(region.Settlements) > 0 {
		return region.Settlements[0].ID
	}
	return ""
}
