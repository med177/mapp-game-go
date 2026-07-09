package ai

import (
	"fmt"
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
	runTurnPrelude(gs, fid, nil)

	// Ordu listesinin anlık kopyasını al — iterasyon sırasında map değişebilir
	var ownArmies []*army.Army
	for _, a := range gs.Armies {
		if a.OwnerID == string(fid) {
			ownArmies = append(ownArmies, a)
		}
	}

	for _, a := range ownArmies {
		// Ordu hâlâ haritada mı?
		if _, alive := gs.Armies[a.ID]; !alive {
			continue
		}
		moveArmyWithSteps(gs, a, fid, nil)
	}
}

func runTurnPrelude(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	if gs == nil {
		return
	}
	// Difficulty 3: koalisyon mantığını çalıştır
	if gs.Difficulty >= 3 {
		formCoalitionAgainstPlayer(gs, fid, steps)
	}

	aiHandleDiplomacyWithSteps(gs, fid, steps)

	// Teknoloji araştırma (önce yap, altın biterse diğerlerini etkilemesin)
	aiResearchWithSteps(gs, fid, steps)

	// Ekonomi optimizasyonu (pazar, çiftlik)
	aiEconomyBuildWithSteps(gs, fid, steps)

	// Deniz stratejisi (liman + gemi)
	aiNavalStrategyWithSteps(gs, fid, steps)

	// Birim alımı ve kışla inşası (elite birimler dahil)
	aiRecruitAndBuildWithSteps(gs, fid, steps)

	// Aynı bölgede olan orduları konsolide et (önceki turlardan veya yeni alımlardan kalan)
	aiConsolidateArmies(gs, fid)
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

func aiHandleDiplomacyWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return
	}

	for otherID, other := range gs.Factions {
		if otherID == fid || other == nil || other.IsEliminated {
			continue
		}

		rel := diplomacy.EnsureRelation(gs, fid, otherID)
		switch rel.Stance {
		case faction.StanceWar:
			selfPower := diplomacy.MilitaryPower(gs, fid)
			otherPower := diplomacy.MilitaryPower(gs, otherID)
			if rel.Score <= -90 || selfPower < otherPower || len(gs.RegionsOwnedBy(fid)) < len(gs.RegionsOwnedBy(otherID)) {
				if otherID == gs.PlayerFactionID {
					priority, reason := aiDiplomacyOfferPriorityDetails(gs, fid, otherID, diplomacy.ActionProposePeace)
					diplomacy.QueueOfferWithMeta(gs, fid, otherID, diplomacy.ActionProposePeace, priority, reason)
					addTurnStep(steps, TurnStep{
						FactionID:     fid,
						Kind:          TurnStepDiplomacy,
						TargetFaction: otherID,
						Message:       turnFactionName(gs, fid) + " sana barış teklif ediyor.",
					})
				} else {
					result := diplomacy.Execute(gs, fid, otherID, diplomacy.ActionProposePeace)
					if result.Applied || result.Accepted {
						addTurnStep(steps, TurnStep{
							FactionID:     fid,
							Kind:          TurnStepDiplomacy,
							TargetFaction: otherID,
							Message:       turnFactionName(gs, fid) + ": " + result.Message,
						})
					}
				}
			}
		case faction.StancePeace:
			if rel.Score >= 20 && diplomacy.HasCommonEnemy(gs, fid, otherID) && !diplomacy.HasDirectThreat(gs, fid, otherID) {
				if otherID == gs.PlayerFactionID {
					priority, reason := aiDiplomacyOfferPriorityDetails(gs, fid, otherID, diplomacy.ActionProposeAlliance)
					diplomacy.QueueOfferWithMeta(gs, fid, otherID, diplomacy.ActionProposeAlliance, priority, reason)
					addTurnStep(steps, TurnStep{
						FactionID:     fid,
						Kind:          TurnStepDiplomacy,
						TargetFaction: otherID,
						Message:       turnFactionName(gs, fid) + " sana ittifak teklif ediyor.",
					})
				} else {
					result := diplomacy.Execute(gs, fid, otherID, diplomacy.ActionProposeAlliance)
					if result.Applied || result.Accepted {
						addTurnStep(steps, TurnStep{
							FactionID:     fid,
							Kind:          TurnStepDiplomacy,
							TargetFaction: otherID,
							Message:       turnFactionName(gs, fid) + ": " + result.Message,
						})
					}
				}
				continue
			}
			if rel.Score >= 15 &&
				diplomacy.Relation(gs, fid, otherID).Stance == faction.StancePeace &&
				aiTradePartnerCount(gs, fid) < 3 &&
				aiTradePartnerCount(gs, otherID) < 3 &&
				!diplomacy.HasDirectThreat(gs, fid, otherID) {
				if otherID == gs.PlayerFactionID {
					assessment := diplomacy.AssessTradeProposal(gs, diplomacy.Relation(gs, fid, otherID), fid, otherID)
					if assessment.BlockReason == "" {
						priority, reason := aiDiplomacyOfferPriorityDetails(gs, fid, otherID, diplomacy.ActionProposeTrade)
						diplomacy.QueueOfferWithMeta(gs, fid, otherID, diplomacy.ActionProposeTrade, priority, reason)
						addTurnStep(steps, TurnStep{
							FactionID:     fid,
							Kind:          TurnStepDiplomacy,
							TargetFaction: otherID,
							Message:       turnFactionName(gs, fid) + " sana ticaret teklif ediyor.",
						})
					}
				} else {
					result := diplomacy.Execute(gs, fid, otherID, diplomacy.ActionProposeTrade)
					if result.Applied || result.Accepted {
						addTurnStep(steps, TurnStep{
							FactionID:     fid,
							Kind:          TurnStepDiplomacy,
							TargetFaction: otherID,
							Message:       turnFactionName(gs, fid) + ": " + result.Message,
						})
					}
				}
			}
		}
	}

	aiEvaluateWarOpportunitiesWithSteps(gs, fid, steps)
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
		if diplomacy.HasCommonEnemy(gs, from, to) {
			score += 16
			reasons = append(reasons, "ortak düşman")
		}
		if !diplomacy.HasDirectThreat(gs, from, to) {
			score += 8
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

func aiEvaluateWarOpportunitiesWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	if gs == nil || gs.Difficulty <= 1 {
		return
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return
	}
	if aiActiveWarCount(gs, fid) >= aiMaxConcurrentWars(gs, fid) {
		return
	}
	if !aiWarCadenceAllows(gs, fid) {
		return
	}

	bestScore := aiWarThreshold
	bestTarget := faction.FactionID("")
	if gs.Difficulty >= 3 {
		bestScore -= 10
	}

	for otherID, other := range gs.Factions {
		if otherID == fid || other == nil || other.IsEliminated {
			continue
		}
		rel := diplomacy.Relation(gs, fid, otherID)
		if rel == nil || rel.Stance != faction.StancePeace {
			continue
		}
		score := aiWarOpportunityScore(gs, fid, otherID, rel)
		if score > bestScore {
			bestScore = score
			bestTarget = otherID
		}
	}

	if bestTarget != "" {
		result := diplomacy.Execute(gs, fid, bestTarget, diplomacy.ActionDeclareWar)
		if result.Applied || result.Accepted {
			addTurnStep(steps, TurnStep{
				FactionID:     fid,
				Kind:          TurnStepDiplomacy,
				TargetFaction: bestTarget,
				Message:       turnFactionName(gs, fid) + ": " + result.Message,
			})
		}
	}
}

func aiWarOpportunityScore(gs *state.GameState, actor, target faction.FactionID, rel *faction.Relation) int {
	self := gs.Factions[actor]
	other := gs.Factions[target]
	if self == nil || other == nil || rel == nil {
		return -1
	}
	isExpansionTarget := aiHasExpansionTarget(self, target)
	maxPeaceScore := -20
	if isExpansionTarget {
		maxPeaceScore = 10
	} else if self.AIAggressiveness >= 70 {
		maxPeaceScore = -10
	}
	if rel.Score > maxPeaceScore || !aiSharesLandBorder(gs, actor, target) {
		return -1
	}

	selfPower := diplomacy.MilitaryPower(gs, actor)
	targetPower := diplomacy.MilitaryPower(gs, target)
	if selfPower <= 0 {
		return -1
	}
	if targetPower > 0 && selfPower*100 < targetPower*115 {
		return -1
	}

	frontierPower := aiFrontierPower(gs, actor, target)
	if frontierPower <= 0 {
		return -1
	}
	targetFrontierPower := aiFrontierPower(gs, target, actor)

	score := 20
	if targetPower == 0 {
		score += 30
	} else {
		powerEdge := (selfPower - targetPower) / 12
		score += minInt(30, maxInt(0, powerEdge))
	}

	if targetFrontierPower == 0 {
		score += 16
	} else if frontierPower > targetFrontierPower {
		score += minInt(22, (frontierPower-targetFrontierPower)/10+8)
	} else {
		score -= 18
	}

	score += minInt(18, maxInt(0, -rel.Score/2))
	if rel.Score > 0 {
		score -= rel.Score
	}

	selfRegions := len(gs.LandRegionsOwnedBy(actor))
	targetRegions := len(gs.LandRegionsOwnedBy(target))
	if targetRegions <= 2 {
		score += 12
	}
	if selfRegions >= targetRegions {
		score += 8
	}
	if gs.DeployedLandUnits(actor) >= gs.ManpowerCap(actor) {
		score += 8
	}

	score += minInt(15, aiBestBorderTargetValue(gs, actor, target)/15)
	if self.Religion != other.Religion {
		score += 6
	} else {
		score -= 6
	}
	score += (self.AIAggressiveness - 45) / 2
	if isExpansionTarget {
		score += 18
		if rel.Score <= 0 {
			score += 6
		}
		if self.AIAggressiveness >= 60 {
			score += 4
		}
	}

	if target == gs.PlayerFactionID {
		score -= 18
		if gs.Difficulty >= 3 {
			score += 8
		}
	}
	return score
}

func aiWarCadenceAllows(gs *state.GameState, fid faction.FactionID) bool {
	if gs == nil || gs.Turn == 0 {
		return true
	}
	interval := 10
	if gs.Difficulty >= 3 {
		interval = 7
	}
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

func aiMaxConcurrentWars(gs *state.GameState, fid faction.FactionID) int {
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
	count := 0
	for _, rel := range gs.Relations {
		if rel == nil || rel.Stance != faction.StanceWar {
			continue
		}
		if rel.FactionA == fid || rel.FactionB == fid {
			count++
		}
	}
	return count
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
				if a.OwnerID == string(fid) && a.IsNaval && a.RegionID == seaRegion {
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

// aiRecruitAndBuild AI fraksiyonu için kışla inşa eder ve manpower sınırına kadar birim alır.
func aiRecruitAndBuild(gs *state.GameState, fid faction.FactionID) {
	aiRecruitAndBuildWithSteps(gs, fid, nil)
}

func aiRecruitAndBuildWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	f, ok := gs.Factions[fid]
	if !ok || f.IsEliminated {
		return
	}

	// Manpower dar ve altın yeterliyse kışla inşa et
	cap := gs.ManpowerCap(fid)
	deployed := gs.DeployedLandUnits(fid) + aiPendingLandUnitCount(gs, fid)
	barracksCost := economy.ResourceCost{Gold: 150}
	if b, ok2 := gs.BuildingTypes["barracks"]; ok2 {
		barracksCost = economy.ResourceCost{
			Gold:   b.GoldCost,
			Grain:  b.GrainCost,
			Iron:   b.IronCost,
			Timber: b.TimberCost,
			Stone:  b.StoneCost,
		}
	}
	if cap-deployed <= state.ManpowerPerRegion && aiCanAffordWithReserve(f, barracksCost) {
		aiBuildBarracksWithSteps(gs, fid, barracksCost, steps)
	}

	// Kapasite dolana veya altın bitene kadar birim al
	for {
		if gs.DeployedLandUnits(fid)+aiPendingLandUnitCount(gs, fid) >= gs.ManpowerCap(fid) {
			break
		}
		if f.Gold < aiMilitiaCost+aiMinGoldReserve {
			break
		}
		if !aiRecruitOneWithSteps(gs, fid, steps) {
			break
		}
	}
}

// aiBuildBarracks kışlası olmayan ilk uygun bölgeye kışla inşa eder.
func aiBuildBarracks(gs *state.GameState, fid faction.FactionID, cost economy.ResourceCost) {
	aiBuildBarracksWithSteps(gs, fid, cost, nil)
}

func aiBuildBarracksWithSteps(gs *state.GameState, fid faction.FactionID, cost economy.ResourceCost, steps *[]TurnStep) {
	f := gs.Factions[fid]
	btype := gs.BuildingTypes["barracks"]
	if btype == nil {
		return
	}
	for _, r := range gs.Regions {
		if r.OwnerID != string(fid) || r.IsSea {
			continue
		}
		queued := aiQueuedBuildingCount(gs, r.ID, "barracks", fid)
		if aiBuildingLevel(r, "barracks")+queued >= btype.MaxPerRegion {
			continue
		}
		if !aiBuildingAllowed(gs, r, "barracks", btype.RequiredTerrain) {
			continue
		}
		cost.Apply(f)
		turns := aiBuildingTurnsRequired(r, "barracks", btype.TurnsRequired, queued)
		aiEnqueueProduction(gs, fid, aiProductionKindBuilding, r.ID, "barracks", turns)
		addTurnStep(steps, TurnStep{
			FactionID:    fid,
			Kind:         TurnStepBuild,
			TargetRegion: r.ID,
			FocusRegion:  r.ID,
			Message:      turnFactionName(gs, fid) + " " + turnRegionName(gs, r.ID) + " bölgesinde kışla kuruyor.",
		})
		return
	}
}

// aiRecruitOne kışlası olan bir bölgede en iyi uygun birimi alır.
// Askeri teknoloji ve altın durumuna göre milis, piyade, süvari veya topçu seçer.
// Başarılıysa true, koşul sağlanamadıysa false döner.
func aiRecruitOne(gs *state.GameState, fid faction.FactionID) bool {
	return aiRecruitOneWithSteps(gs, fid, nil)
}

func aiRecruitOneWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) bool {
	f := gs.Factions[fid]
	if gs.UnitTypes == nil {
		return false
	}

	// En iyi birimi seç (stratejik karar)
	unitTypeID := aiSelectBestUnit(gs, f)
	if unitTypeID == "" {
		return false
	}

	utype, ok := gs.UnitTypes[unitTypeID]
	if !ok {
		return false
	}

	unitCost := economy.ResourceCost{
		Gold:   utype.GoldCost,
		Grain:  utype.GrainCost,
		Iron:   utype.IronCost,
		Timber: utype.TimberCost,
		Stone:  utype.StoneCost,
	}
	if !aiCanAffordWithReserve(f, unitCost) {
		return false
	}

	recruitRegion := aiFindRecruitRegion(gs, fid, utype)
	if recruitRegion == "" {
		return false
	}
	if aiPendingUnitCountByRegion(gs, recruitRegion, fid) >= aiMaxRegionQueue {
		return false
	}
	if !aiCanQueueLandUnit(gs, fid, recruitRegion, utype) {
		return false
	}

	unitCost.Apply(f)
	aiEnqueueProduction(gs, fid, aiProductionKindUnit, recruitRegion, unitTypeID, utype.TurnsRequired)
	return true
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
	for _, r := range gs.Regions {
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
	for _, a := range gs.Armies {
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
	if f.Gold >= 350+aiMinGoldReserve {
		if ut, ok := gs.UnitTypes["elite_infantry"]; ok {
			if aiUnitAvailableForRecruitment(gs, f, ut) {
				return "elite_infantry"
			}
		}
	}

	// Ağır süvari - zengin ve teknolojisi varsa
	if f.Gold >= 450+aiMinGoldReserve && cavalryCount < armyCount*2 {
		if ut, ok := gs.UnitTypes["heavy_cavalry"]; ok {
			if aiUnitAvailableForRecruitment(gs, f, ut) {
				return "heavy_cavalry"
			}
		}
	}

	// Tier 2 piyade (normal piyade) - orta düzey altın ve teknoloji
	if f.Gold >= 180+aiMinGoldReserve {
		if ut, ok := gs.UnitTypes["infantry"]; ok {
			if aiUnitAvailableForRecruitment(gs, f, ut) {
				return "infantry"
			}
		}
	}

	// Süvari - teknolojisi varsa ve altın yeterliyse
	if f.Gold >= 300+aiMinGoldReserve && cavalryCount < armyCount*3 {
		if ut, ok := gs.UnitTypes["cavalry"]; ok {
			if aiUnitAvailableForRecruitment(gs, f, ut) {
				return "cavalry"
			}
		}
	}

	// Hafif süvari - her zaman uygun
	if f.Gold >= 200+aiMinGoldReserve && cavalryCount < armyCount*4 {
		if ut, ok := gs.UnitTypes["light_cavalry"]; ok && aiUnitAvailableForRecruitment(gs, f, ut) {
			return "light_cavalry"
		}
	}

	// Topçu - çok zenginse ve savaşta ise
	if f.Gold >= 650+aiMinGoldReserve {
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
				if aiUnitAvailableForRecruitment(gs, f, ut) {
					return "cannon"
				}
			}
			if ut, ok := gs.UnitTypes["bombard"]; ok {
				if aiUnitAvailableForRecruitment(gs, f, ut) {
					return "bombard"
				}
			}
		}
	}

	// Varsayılan: milis
	if ut, ok := gs.UnitTypes["militia"]; ok && aiUnitAvailableForRecruitment(gs, f, ut) {
		return "militia"
	}
	return ""
}

func aiUnitAvailableForRecruitment(gs *state.GameState, f *faction.Faction, utype *army.UnitType) bool {
	if gs == nil || f == nil || utype == nil {
		return false
	}
	if utype.RequiredTech != "" && !f.Research.Completed[utype.RequiredTech] {
		return false
	}
	cost := economy.ResourceCost{
		Gold:   utype.GoldCost,
		Grain:  utype.GrainCost,
		Iron:   utype.IronCost,
		Timber: utype.TimberCost,
		Stone:  utype.StoneCost,
	}
	if !aiCanAffordWithReserve(f, cost) {
		return false
	}
	return aiFindRecruitRegion(gs, f.ID, utype) != ""
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
	for otherFID := range gs.Factions {
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
	for a.MovePoints > 0 {
		target := chooseBestMove(gs, a)
		if target == "" {
			break
		}

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
	for _, a := range gs.Armies {
		if a.ID == transport.ID || !a.IsNaval || a.OwnerID != transport.OwnerID || a.RegionID != transport.RegionID {
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
		if ea.RegionID == target && ea.OwnerID != transport.OwnerID && ea.IsNaval {
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

			if result.AttackerWins {
				if len(enemyInTarget.Units) == 0 {
					delete(gs.Armies, enemyInTarget.ID)
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
					delete(gs.Armies, escort.ID)
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
	for _, a := range gs.Armies {
		if a == nil || !a.IsNaval || a.OwnerID == ownerID || a.RegionID != seaRegionID {
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
	for _, candidate := range gs.Armies {
		if candidate.OwnerID != ownerID || !candidate.IsNaval || candidate.RegionID != seaRegionID {
			continue
		}
		if candidate.CanEmbarkUnits(gs.UnitTypes, unitCount) {
			return candidate
		}
	}
	return nil
}

func aiEmbarkScore(gs *state.GameState, a *army.Army, seaRegion *world.Region) int {
	if gs == nil || a == nil || seaRegion == nil || !seaRegion.IsSea {
		return 0
	}
	if !aiCanEmbarkArmy(gs, a) || aiFindEmbarkFleet(gs, a.OwnerID, seaRegion.ID, len(a.Units)) == nil {
		return 0
	}
	best := 10 + aiSeaPressure(gs, a.OwnerID, seaRegion.ID)/2
	for _, nid := range seaRegion.Neighbors {
		land, ok := gs.Regions[nid]
		if !ok || land.IsSea {
			continue
		}
		score := scoreMove(gs, a, land)
		if score > best {
			best = score
		}
	}
	return best
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
		if a == nil || a.OwnerID != ownerID || !a.IsNaval || a.RegionID != seaRegionID {
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
	for _, ea := range gs.Armies {
		if ea.RegionID == rid && ea.OwnerID != ownerID {
			return ea
		}
	}
	return nil
}

func aiSpawnDisembarkedArmy(gs *state.GameState, ownerID string, target world.RegionID, units []army.Unit) {
	if gs == nil || len(units) == 0 {
		return
	}
	gs.NextArmySeq++
	newID := army.ArmyID(fmt.Sprintf("army_%s_%d", ownerID, gs.NextArmySeq))
	gs.Armies[newID] = &army.Army{
		ID:            newID,
		OwnerID:       ownerID,
		RegionID:      target,
		Units:         units,
		MovePoints:    0,
		MaxMovePoints: 2,
		IsNaval:       false,
	}
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

func aiEnsureSiegeMap(gs *state.GameState) {
	if gs != nil && gs.Sieges == nil {
		gs.Sieges = make(map[world.RegionID]*state.SiegeState)
	}
}

func aiClearSiege(gs *state.GameState, regionID world.RegionID) {
	if gs == nil || gs.Sieges == nil || regionID == "" {
		return
	}
	delete(gs.Sieges, regionID)
}

func aiClearSiegesByArmy(gs *state.GameState, armyID army.ArmyID) {
	if gs == nil || gs.Sieges == nil || armyID == "" {
		return
	}
	for rid, siege := range gs.Sieges {
		if siege != nil && siege.AttackerArmyID == armyID {
			delete(gs.Sieges, rid)
		}
	}
}

func aiSiegeDefenseBonus(fortLevel, breachLevel int) float64 {
	if fortLevel <= 0 {
		return 0
	}
	base := float64(fortLevel) * 0.14
	switch breachLevel {
	case 2:
		return base * 0.25
	case 1:
		return base * 0.55
	default:
		return base + 0.18
	}
}

func aiVirtualSiegeGarrison(gs *state.GameState, target *world.Region) *army.Army {
	if gs == nil || target == nil {
		return nil
	}
	unitTypeID := aiMilitiaID
	if _, ok := gs.UnitTypes[unitTypeID]; !ok {
		for id, ut := range gs.UnitTypes {
			if ut != nil && ut.Category == army.CategoryInfantry {
				unitTypeID = id
				break
			}
		}
	}
	fortLevel := target.FortificationLevel()
	unitCount := 1 + fortLevel
	if unitCount > 6 {
		unitCount = 6
	}
	units := make([]army.Unit, 0, unitCount)
	for i := 0; i < unitCount; i++ {
		units = append(units, army.Unit{TypeID: unitTypeID, CurrentHP: army.MaxUnitHP})
	}
	return &army.Army{
		OwnerID:    target.OwnerID,
		RegionID:   target.ID,
		Units:      units,
		MovePoints: 0,
	}
}

func aiCanStartSiege(gs *state.GameState, a *army.Army, target *world.Region) bool {
	if gs == nil || a == nil || target == nil || a.IsNaval || !target.CanLandEnter() {
		return false
	}
	if target.OwnerID == "" || target.OwnerID == a.OwnerID || !target.IsFortified() {
		return false
	}
	if !a.HasSiegeUnits(gs.UnitTypes) {
		return false
	}
	_, stance := relationScore(gs, a.OwnerID, target.OwnerID)
	if stance != faction.StanceWar {
		return false
	}
	if active := gs.SiegeByArmy(a.ID); active != nil && active.RegionID != target.ID {
		return false
	}
	if siege := gs.SiegeAt(target.ID); siege != nil && siege.AttackerArmyID != a.ID {
		return false
	}
	return true
}

func aiStartSiege(gs *state.GameState, a *army.Army, target *world.Region, defender *army.Army) {
	if gs == nil || a == nil || target == nil {
		return
	}
	aiEnsureSiegeMap(gs)
	siege := &state.SiegeState{
		RegionID:          target.ID,
		AttackerArmyID:    a.ID,
		AttackerFactionID: a.OwnerID,
		StartedTurn:       gs.Turn,
		FortLevel:         target.FortificationLevel(),
	}
	if defender != nil {
		siege.DefenderArmyID = defender.ID
	}
	gs.Sieges[target.ID] = siege
	a.MovePoints = 0
}

// chooseBestMove ordunun komşuları arasında en iyi hedefi seçer.
// Negatif skor dönen hedefler atlanır; hiç geçerli hedef yoksa "" döner.
func chooseBestMove(gs *state.GameState, a *army.Army) world.RegionID {
	src, ok := gs.Regions[a.RegionID]
	if !ok {
		return ""
	}
	if activeSiege := gs.SiegeByArmy(a.ID); activeSiege != nil {
		target := gs.Regions[activeSiege.RegionID]
		if !aiCanStartSiege(gs, a, target) {
			return ""
		}
		if activeSiege.BreachLevel >= 2 {
			return activeSiege.RegionID
		}
		return ""
	}

	bestScore := 0
	var bestTarget world.RegionID

	if a.IsNaval {
		for _, nid := range src.Neighbors {
			n, ok := gs.Regions[nid]
			if !ok {
				continue
			}
			if n.IsSea {
				score := 15 + aiSeaPressure(gs, a.OwnerID, n.ID)
				if len(a.EmbarkedUnits) > 0 {
					for _, landID := range n.Neighbors {
						land, ok := gs.Regions[landID]
						if !ok || land.IsSea {
							continue
						}
						if land.OwnerID != "" && land.OwnerID != a.OwnerID {
							score += 20
						}
					}
				}
				if score > bestScore {
					bestScore = score
					bestTarget = nid
				}
				continue
			}
			if !aiCanDisembarkToLand(gs, a, n) {
				continue
			}
			score := 40
			enemyArmy := aiEnemyArmyInRegion(gs, a.OwnerID, n.ID)
			if enemyArmy != nil {
				landingStr := aiLandingStrength(gs, a)
				defStr := enemyArmy.TotalStrength(gs.UnitTypes)
				if landingStr <= defStr {
					continue
				}
				score = 75
			} else if n.OwnerID != "" && n.OwnerID != a.OwnerID {
				score = 60
			}
			if score > bestScore {
				bestScore = score
				bestTarget = nid
			}
		}
		return bestTarget
	}

	for _, nid := range src.Neighbors {
		n, ok := gs.Regions[nid]
		if !ok {
			continue
		}
		if n.IsSea {
			score := aiEmbarkScore(gs, a, n)
			if score > bestScore {
				bestScore = score
				bestTarget = nid
			}
			continue
		}
		score := scoreMove(gs, a, n)
		if score > bestScore {
			bestScore = score
			bestTarget = nid
		}
	}

	// Eğer komşularda mantıklı bir hedef yoksa, uzun menzilli planlama yap (BFS)
	if bestScore == 0 {
		bestTarget = findLongRangeMove(gs, a, src)
	}

	return bestTarget
}

// findLongRangeMove BFS kullanarak en yakın değerli (score > 0) bölgeye giden ilk adımı bulur.
func findLongRangeMove(gs *state.GameState, a *army.Army, start *world.Region) world.RegionID {
	type queueItem struct {
		id   world.RegionID
		path []world.RegionID
	}

	visited := make(map[world.RegionID]bool)
	queue := []queueItem{{id: start.ID, path: nil}}
	visited[start.ID] = true

	maxDepth := 8 // En fazla 8 bölge uzağa bak

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if len(curr.path) > maxDepth {
			continue
		}

		r, ok := gs.Regions[curr.id]
		if !ok {
			continue
		}

		// Kendi bölgesi değilse ve score > 0 ise hedef bulduk demektir
		if curr.id != start.ID {
			score := scoreMove(gs, a, r)
			if score > 0 {
				return curr.path[0] // Hedefe giden ilk adımı dön
			}
			// Düşman toprağıysa daha ileri gitme
			if r.OwnerID != a.OwnerID && r.OwnerID != "" {
				continue
			}
		}

		for _, nid := range r.Neighbors {
			n, ok := gs.Regions[nid]
			if !ok || n.IsSea || visited[nid] {
				continue
			}
			visited[nid] = true

			newPath := make([]world.RegionID, len(curr.path))
			copy(newPath, curr.path)
			newPath = append(newPath, nid)

			queue = append(queue, queueItem{id: nid, path: newPath})
		}
	}
	return ""
}

// scoreMove bir hedefe yapılacak hareketin değerini puanlar.
func scoreMove(gs *state.GameState, a *army.Army, target *world.Region) int {
	fid := faction.FactionID(a.OwnerID)
	source := gs.Regions[a.RegionID]
	armyDemand := a.TotalGrainUpkeep(gs.UnitTypes)
	if target.OwnerID == a.OwnerID {
		score := 0
		srcDemand, srcCap, srcOverload := aiRegionLogistics(gs, source, a.OwnerID)
		tgtDemand, tgtCap, tgtOverload := aiRegionLogistics(gs, target, a.OwnerID)
		srcAfter := maxInt(0, srcDemand-armyDemand-srcCap)
		tgtAfter := maxInt(0, tgtDemand+armyDemand-tgtCap)

		if srcOverload > 0 || a.OverCapacityTurns > 0 {
			relief := srcOverload - srcAfter
			if relief > 0 {
				score += aiReliefMoveBase + relief*2
			}
			if tgtAfter == 0 {
				score += 18
			}
			if tgtOverload == 0 {
				spare := tgtCap - tgtDemand
				if spare > 0 {
					score += minInt(18, spare)
				}
			}
			if tgtAfter > srcAfter {
				score -= 45
			}
		}

		// Dost bölgede birleşebileceğimiz ordu var mı? (Konsolidasyon)
		for _, ea := range gs.Armies {
			if ea.RegionID == target.ID && ea.OwnerID == a.OwnerID && ea.ID != a.ID && ea.IsNaval == a.IsNaval {
				if len(a.Units)+len(ea.Units) <= army.MaxArmySize && aiShouldConsolidateInRegion(gs, target, a.OwnerID, a.IsNaval) {
					if score < 60 {
						score = 60
					}
				}
			}
		}
		return score
	}

	// Müttefik veya savaş halindeki fraksiyona göre hareket et.
	if target.OwnerID != "" {
		_, stance := relationScore(gs, a.OwnerID, target.OwnerID)
		if stance == faction.StanceAllied {
			// Müttefik bölgesine savaşsız geçiş: lojistik rahatlatma veya yol amaçlı
			_, _, srcOverload := aiRegionLogistics(gs, source, a.OwnerID)
			if srcOverload > 0 || a.OverCapacityTurns > 0 {
				tgtDemand, tgtCap, tgtOverload := aiRegionLogistics(gs, target, a.OwnerID)
				if tgtDemand+armyDemand <= tgtCap && tgtOverload == 0 {
					return aiReliefMoveBase
				}
			}
			// Uzun menzilli hareket için müttefik topraklarından geçişe düşük skor
			return 5
		}
		if stance != faction.StanceWar {
			return -1
		}
	}
	if !a.IsNaval && target.CanLandEnter() && target.OwnerID != "" && target.OwnerID != a.OwnerID && target.IsFortified() {
		if !a.HasSiegeUnits(gs.UnitTypes) {
			return -1
		}
		if siege := gs.SiegeAt(target.ID); siege != nil && siege.AttackerArmyID != a.ID {
			return -1
		}
		if atCapacity := gs.DeployedLandUnits(fid) >= gs.ManpowerCap(fid); atCapacity {
			return 100
		}
		return 92
	}

	// Kapasite doluysa fetih yaparak manpower artırmak öncelikli
	atCapacity := gs.DeployedLandUnits(fid) >= gs.ManpowerCap(fid)

	// Düşman ordusu var mı?
	for _, ea := range gs.Armies {
		if ea.RegionID != target.ID || ea.OwnerID == a.OwnerID {
			continue
		}
		atkStr := a.TotalStrength(gs.UnitTypes)
		defStr := ea.TotalStrength(gs.UnitTypes)
		if atkStr > defStr {
			// Savaş halindeyse öncelikli hedef
			_, stance := relationScore(gs, a.OwnerID, target.OwnerID)
			if stance == faction.StanceWar {
				return 95
			}
			return 75
		}
		return -1
	}

	// Kapasite doluysa sahipsiz bölge almak çok değerli (manpower genişler)
	if target.OwnerID == "" {
		if atCapacity {
			return 70
		}
		return 50
	}
	// Düşman bölgesi, ordu yok — savaş halindeyse puanla
	_, stance := relationScore(gs, a.OwnerID, target.OwnerID)
	if stance == faction.StanceWar {
		if atCapacity {
			return 100
		}
		return 90
	}
	// Müttefik bölgesine savaşsız geçiş (düşük öncelik)
	if stance == faction.StanceAllied {
		return 10
	}
	return -1
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

	// Kuşatma altındaki ordu hareket edemez; önce huruç savaşı yapmalı.
	// Eğer kuşatan oyuncu ise sortie step'i döner (battle plan UI için).
	if !a.IsNaval {
		if siege := gs.SiegeAt(fromRegion); siege != nil && siege.AttackerArmyID != a.ID {
			siegeArmy := gs.Armies[siege.AttackerArmyID]
			if siegeArmy != nil && siegeArmy.OwnerID != a.OwnerID {
				// AI vs AI sortie (veya oyuncu kuşatıyorsa): hemen çöz
				atkMods := aiTechMods(gs, a.OwnerID)
				defMods := aiTechMods(gs, siegeArmy.OwnerID)
				defMods.DefenseMod += 0.10
				result := combat.ResolveBattleWithMods(a, siegeArmy, targetRegion.Terrain, gs.UnitTypes, atkMods, defMods)
				if result.AttackerWins {
					if len(siegeArmy.Units) == 0 {
						delete(gs.Armies, siegeArmy.ID)
					}
					if siege.AttackerHomeRegionID != "" {
						siegeArmy.RegionID = siege.AttackerHomeRegionID
					}
					delete(gs.Sieges, fromRegion)
					msg := actorName + " " + sourceName + " kuşatmasını yardı ve çıktı."
					return moveOutcome{survived: len(a.Units) > 0, step: TurnStep{FactionID: fid, Kind: TurnStepBattle, ArmyID: a.ID, FromRegion: fromRegion, TargetRegion: target, FocusRegion: fromRegion, Message: msg}}
				}
				a.MovePoints = 0
				if len(a.Units) == 0 {
					delete(gs.Armies, a.ID)
				}
				if len(siegeArmy.Units) == 0 {
					delete(gs.Armies, siegeArmy.ID)
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
		if enemyArmy != nil {
			landing := &army.Army{
				OwnerID: a.OwnerID,
				Units:   append([]army.Unit(nil), a.EmbarkedUnits...),
			}
			atkMods := aiTechMods(gs, a.OwnerID)
			defMods := aiTechMods(gs, enemyArmy.OwnerID)
			result := combat.ResolveBattleWithMods(landing, enemyArmy, targetRegion.Terrain, gs.UnitTypes, atkMods, defMods)
			a.EmbarkedUnits = a.EmbarkedUnits[:0]
			a.MovePoints--
			if result.AttackerWins {
				if len(enemyArmy.Units) == 0 {
					delete(gs.Armies, enemyArmy.ID)
				}
				aiSpawnDisembarkedArmy(gs, a.OwnerID, target, landing.Units)
				targetRegion.ApplyConquest(a.OwnerID, aiOwnerReligion(gs, a.OwnerID))
				return moveOutcome{
					survived: true,
					step: TurnStep{
						FactionID:    fid,
						Kind:         TurnStepBattle,
						ArmyID:       a.ID,
						FromRegion:   fromRegion,
						TargetRegion: target,
						FocusRegion:  target,
						Message:      actorName + " " + targetName + " kıyısına çıkarma yapıp bölgeyi aldı.",
					},
				}
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
					Message:      actorName + " " + targetName + " kıyısındaki çıkarmada geri püskürtüldü.",
				},
			}
		}
		units := make([]army.Unit, len(a.EmbarkedUnits))
		copy(units, a.EmbarkedUnits)
		a.EmbarkedUnits = a.EmbarkedUnits[:0]
		aiSpawnDisembarkedArmy(gs, a.OwnerID, target, units)
		stepKind := TurnStepDisembark
		msg := actorName + " " + targetName + " kıyısına çıkarma yaptı."
		isAlliedDisembark := false
		if targetRegion.OwnerID != a.OwnerID && targetRegion.OwnerID != "" {
			key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(targetRegion.OwnerID))
			if rel, exists := gs.Relations[key]; exists && rel.Stance == faction.StanceAllied {
				isAlliedDisembark = true
			}
		}
		if targetRegion.OwnerID != a.OwnerID && !isAlliedDisembark {
			targetRegion.ApplyConquest(a.OwnerID, aiOwnerReligion(gs, a.OwnerID))
			stepKind = TurnStepConquest
			msg = actorName + " " + targetName + " kıyısına çıktı ve bölgeyi ele geçirdi."
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
		delete(gs.Armies, a.ID)
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
			if result.AttackerWins {
				if !virtualDefense && len(defender.Units) == 0 {
					delete(gs.Armies, defender.ID)
				}
				if len(a.Units) > 0 {
					a.RegionID = target
					a.DockedRegionID = ""
					a.DockedSettlementID = ""
					a.MovePoints--
					targetRegion.ApplyConquest(a.OwnerID, aiOwnerReligion(gs, a.OwnerID))
					aiClearSiege(gs, target)
					return moveOutcome{
						survived: true,
						step: TurnStep{
							FactionID:    fid,
							Kind:         TurnStepBattle,
							ArmyID:       a.ID,
							FromRegion:   fromRegion,
							TargetRegion: target,
							FocusRegion:  target,
							Message:      actorName + " " + targetName + " tahkimatına genel hücumla girdi ve kazandı.",
						},
					}
				}
				delete(gs.Armies, a.ID)
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
				delete(gs.Armies, a.ID)
				aiClearSiegesByArmy(gs, a.ID)
			}
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
	combinedDef, defSourceIDs := gs.CollectDefenders(a, target, false)
	var enemyArmy *army.Army
	if combinedDef == nil {
		for _, ea := range gs.Armies {
			if ea.RegionID == target && ea.OwnerID != a.OwnerID {
				enemyArmy = ea
				break
			}
		}
	} else {
		// Birleşik ordudan refakat için ilk orduyu bul
		for _, ea := range gs.Armies {
			if ea.RegionID == target && ea.OwnerID != a.OwnerID {
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
		if result.AttackerWins {
			if len(defSourceIDs) > 0 {
				gs.DistributeDefenderLosses(defSourceIDs, result.DefenderLost)
			} else if enemyArmy != nil && len(enemyArmy.Units) == 0 {
				delete(gs.Armies, enemyArmy.ID)
			}
			if len(a.Units) > 0 {
				a.RegionID = target
				a.DockedRegionID = ""
				a.DockedSettlementID = ""
				targetRegion.OwnerID = a.OwnerID
				a.MovePoints--
				return moveOutcome{
					survived: true,
					step: TurnStep{
						FactionID:    fid,
						Kind:         TurnStepBattle,
						ArmyID:       a.ID,
						FromRegion:   fromRegion,
						TargetRegion: target,
						FocusRegion:  target,
						Message:      actorName + " " + targetName + " bölgesindeki savaşı kazandı.",
					},
				}
			}
			delete(gs.Armies, a.ID)
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
			delete(gs.Armies, a.ID)
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
		targetRegion.OwnerID = a.OwnerID
		stepKind = TurnStepConquest
		msg = actorName + " " + targetName + " bölgesini savaşsız ele geçirdi."
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

// aiResearch aktif araştırma yoksa stratejik teknoloji seçer ve başlatır.
// Öncelik: askeri > ekonomi > diplomasi > diğer
func aiResearch(gs *state.GameState, fid faction.FactionID) {
	aiResearchWithSteps(gs, fid, nil)
}

func aiResearchWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	f := gs.Factions[fid]
	if f.IsEliminated || gs.TechTypes == nil {
		return
	}
	// Zaten araştırma var mı?
	if f.Research.ActiveID != "" {
		return
	}
	// Yeterli altın var mı?
	if f.Gold < aiTechReserve {
		return
	}

	// Uygun teknolojileri puanla
	type scoredTech struct {
		t     *tech.Technology
		score int
	}
	var candidates []scoredTech

	for _, t := range gs.TechTypes {
		// Zaten tamamlandı mı?
		if f.Research.Completed[t.ID] {
			continue
		}
		// Gereksinimler sağlanıyor mu?
		if !tech.IsUnlocked(&f.Research, t) {
			continue
		}
		// Yeterli altın var mı?
		if f.Gold < t.GoldCost+aiMinGoldReserve {
			continue
		}

		score := 0
		switch t.Category {
		case tech.CategoryMilitary:
			score = 100 // Askeri teknolojiler en yüksek öncelik
			if t.Effects.InfantryAttackMod > 0 || t.Effects.CavalryAttackMod > 0 {
				score += 20
			}
		case tech.CategoryEconomy:
			score = 70 // Ekonomi ikinci öncelik
			if t.Effects.GoldPerRegion > 0 {
				score += 15
			}
		case tech.CategoryNaval:
			score = 50 // Deniz teknolojisi (kıyı fraksiyonları için daha yüksek olabilir)
		case tech.CategoryDiplomacy:
			score = 40
		case tech.CategoryReligion:
			score = 30
		}

		// Daha kısa süren teknolojilere hafif bonus
		score -= t.TurnsRequired / 2

		candidates = append(candidates, scoredTech{t, score})
	}

	if len(candidates) == 0 {
		return
	}

	// En yüksek puanlı teknolojiyi seç
	var best *tech.Technology
	bestScore := -1
	for _, c := range candidates {
		if c.score > bestScore {
			bestScore = c.score
			best = c.t
		}
	}

	if best != nil {
		if tech.StartResearch(&f.Research, best, &f.Gold) {
			addTurnStep(steps, TurnStep{
				FactionID: fid,
				Kind:      TurnStepResearch,
				Message:   turnFactionName(gs, fid) + " " + best.NameTR + " araştırmasını başlattı.",
			})
		}
	}
}

// aiEconomyBuild ekonomik binalar inşa eder (pazar, çiftlik).
// Her tur sadece bir bina inşa eder (limitli).
func aiEconomyBuild(gs *state.GameState, fid faction.FactionID) {
	aiEconomyBuildWithSteps(gs, fid, nil)
}

func aiEconomyBuildWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	f := gs.Factions[fid]
	if f.IsEliminated || gs.BuildingTypes == nil {
		return
	}
	if f.Gold < 200+aiMinGoldReserve {
		return
	}

	// Bina öncelikleri ve maliyet kontrolü
	type buildingPlan struct {
		id     string
		needFn func(*world.Region) bool
		prio   int
	}

	plans := []buildingPlan{
		{"farm", func(r *world.Region) bool {
			// Tahıl üretimi düşük bölgelere çiftlik
			return r.BaseGrainOutput < 20
		}, 60},
		{"market", func(r *world.Region) bool {
			// Geliri artırmak için pazar
			return true
		}, 80},
		{"walls", func(r *world.Region) bool {
			// Sınır bölgelerine sur
			for _, nid := range r.Neighbors {
				if n, ok := gs.Regions[nid]; ok && n.OwnerID != "" && n.OwnerID != string(fid) {
					return true
				}
			}
			return false
		}, 50},
	}

	for _, plan := range plans {
		btype, ok := gs.BuildingTypes[plan.id]
		if !ok {
			continue
		}
		buildCost := economy.ResourceCost{
			Gold:   btype.GoldCost,
			Grain:  btype.GrainCost,
			Iron:   btype.IronCost,
			Timber: btype.TimberCost,
			Stone:  btype.StoneCost,
		}
		if !aiCanAffordWithReserve(f, buildCost) {
			continue
		}

		// Uygun bölge bul
		for _, r := range gs.Regions {
			if r.OwnerID != string(fid) || r.IsSea {
				continue
			}
			if !aiBuildingAllowed(gs, r, plan.id, btype.RequiredTerrain) {
				continue
			}
			queued := aiQueuedBuildingCount(gs, r.ID, plan.id, fid)
			if aiBuildingLevel(r, plan.id)+queued >= btype.MaxPerRegion {
				continue
			}
			// İhtiyaç var mı?
			if plan.needFn(r) {
				buildCost.Apply(f)
				turns := aiBuildingTurnsRequired(r, plan.id, btype.TurnsRequired, queued)
				aiEnqueueProduction(gs, fid, aiProductionKindBuilding, r.ID, plan.id, turns)
				addTurnStep(steps, TurnStep{
					FactionID:    fid,
					Kind:         TurnStepBuild,
					TargetRegion: r.ID,
					FocusRegion:  r.ID,
					Message:      turnFactionName(gs, fid) + " " + turnRegionName(gs, r.ID) + " bölgesinde " + btype.NameTR + " inşasını başlattı.",
				})
				return // Bir bina inşa ettik, turu bitir
			}
		}
	}
}

// aiNavalStrategy kıyı fraksiyonları için liman ve gemi inşası yapar.
func aiNavalStrategy(gs *state.GameState, fid faction.FactionID) {
	aiNavalStrategyWithSteps(gs, fid, nil)
}

func aiNavalStrategyWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	f := gs.Factions[fid]
	if f.IsEliminated || gs.BuildingTypes == nil || gs.UnitTypes == nil {
		return
	}

	// Kıyı bölgesi var mı?
	var coastalRegions []*world.Region
	for _, r := range gs.Regions {
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
		}
		if aiBuildingLevel(r, "port")+queued < portType.MaxPerRegion &&
			aiBuildingAllowed(gs, r, "port", portType.RequiredTerrain) &&
			aiCanAffordWithReserve(f, portCost) {
			portCost.Apply(f)
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
		for _, a := range gs.Armies {
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
	}
	if !aiCanAffordWithReserve(f, shipCost) {
		return
	}

	shipCost.Apply(f)
	aiEnqueueProduction(gs, fid, aiProductionKindUnit, bestRegion.ID, "transport", transportType.TurnsRequired)
	addTurnStep(steps, TurnStep{
		FactionID:    fid,
		Kind:         TurnStepRecruit,
		TargetRegion: bestRegion.ID,
		FocusRegion:  bestSeaRegion,
		Message:      turnFactionName(gs, fid) + " " + turnRegionName(gs, bestRegion.ID) + " limanında nakliye gemisi hazırlıyor.",
	})

	// Escort savaş gemisi üretimi — transport varsa ve savaş halinde veya deniz baskısı yüksekse
	aiProduceEscortIfNeeded(gs, fid, coastalRegions, steps)
}

// aiProduceEscortIfNeeded transport hattı olan AI için escort savaş gemisi üretir.
func aiProduceEscortIfNeeded(gs *state.GameState, fid faction.FactionID, coastalRegions []*world.Region, steps *[]TurnStep) {
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
	}
	if !aiCanAffordWithReserve(f, warshipCost) {
		return
	}

	// Teknoloji kontrolü
	if warshipType.RequiredTech != "" && !f.Research.Completed[warshipType.RequiredTech] {
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
		for _, a := range gs.Armies {
			if a.RegionID == candidate.seaID && a.OwnerID == string(fid) && a.IsNaval && isWarshipFleet(a, gs.UnitTypes) {
				currentWarshipUnits = len(a.Units)
				break
			}
		}
		if currentWarshipUnits+aiPendingNavalUnitCount(gs, candidate.seaID, fid) >= army.MaxArmySize {
			continue
		}
		if !aiCanAffordWithReserve(f, warshipCost) {
			break
		}
		warshipCost.Apply(f)
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
	for _, a := range gs.Armies {
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
			if a1.RegionID == a2.RegionID && a1.IsNaval == a2.IsNaval {
				region := gs.Regions[a1.RegionID]
				if !aiShouldConsolidateInRegion(gs, region, a1.OwnerID, a1.IsNaval) {
					continue
				}
				if len(a1.Units)+len(a2.Units) <= army.MaxArmySize {
					a1.Units = append(a1.Units, a2.Units...)
					delete(gs.Armies, a2.ID)
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
	for otherID, other := range gs.Armies {
		if otherID == a.ID || other.RegionID != a.RegionID || other.OwnerID != a.OwnerID || other.IsNaval != a.IsNaval {
			continue
		}
		if len(a.Units)+len(other.Units) <= army.MaxArmySize {
			other.Units = append(other.Units, a.Units...)
			delete(gs.Armies, a.ID)
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

	production := gs.RegionProductionSummary(region).Grain
	settlementBuffer := aiRegionSettlementBuffer(gs, region)
	reserveSupport := aiRegionReserveSupport(gs, ownerID, production, settlementBuffer)
	capacity = production + settlementBuffer + reserveSupport
	if capacity < 4 {
		capacity = 4
	}

	for _, candidate := range gs.Armies {
		if candidate == nil || candidate.IsNaval || candidate.OwnerID != ownerID || candidate.RegionID != region.ID {
			continue
		}
		demand += candidate.TotalGrainUpkeep(gs.UnitTypes)
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
