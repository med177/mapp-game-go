package diplomacy

import (
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

type Action string

const (
	ActionDeclareWar         Action = "declare_war"
	ActionJoinWarCall        Action = "join_war_call"
	ActionProposePeace       Action = "propose_peace"
	ActionProposeAlliance    Action = "propose_alliance"
	ActionProposeTrade       Action = "propose_trade"
	ActionCancelAlliance     Action = "cancel_alliance"
	ActionCancelTrade        Action = "cancel_trade"
	ActionImproveRelations   Action = "improve_relations"
	ActionSendGift           Action = "send_gift"
	ActionOfferVassalization Action = "offer_vassalization"
	ActionReleaseVassal      Action = "release_vassal"
	ActionAnnexVassal        Action = "annex_vassal"
)

type Result struct {
	Accepted bool
	Applied  bool
	Message  string
}

const tradeAcceptanceThreshold = 45
const tradeRelationThreshold = 15

type TradeProposalAssessment struct {
	Chance      int
	BlockReason string
}

func (a TradeProposalAssessment) Accepted() bool {
	return a.BlockReason == "" && a.Chance >= tradeAcceptanceThreshold
}

const allianceAcceptanceThreshold = 45
const allianceRelationThreshold = 25

type AllianceProposalAssessment struct {
	Chance          int
	BlockReason     string
	ActorStrategic  StrategicAllianceAssessment
	TargetStrategic StrategicAllianceAssessment
}

func (a AllianceProposalAssessment) Accepted() bool {
	return a.BlockReason == "" && a.Chance >= allianceAcceptanceThreshold
}

func Execute(gs *state.GameState, actor, target faction.FactionID, action Action) Result {
	if reason := ActionBlockReason(gs, actor, target, action); reason != "" {
		return Result{Message: reason}
	}
	if actionUsesDiplomacyOfferQuota(action) && !spendDiplomacyOfferQuota(gs, actor) {
		return Result{Message: diplomacyOfferQuotaBlockReasonTR}
	}

	rel := EnsureRelation(gs, actor, target)
	switch action {
	case ActionDeclareWar:
		return ExecuteWarDeclaration(gs, actor, target, nil).Result

	case ActionProposePeace:
		if !acceptPeace(gs, rel, actor, target) {
			return Result{Message: factionLabel(gs, target) + " barışı reddetti."}
		}
		setPeaceBetweenCoalitions(gs, actor, target)
		return Result{Accepted: true, Applied: true, Message: factionLabel(gs, target) + " barışı kabul etti."}

	case ActionProposeTrade:
		if rel.Stance == faction.StanceTrade {
			if HasTradeRouteBetween(gs, actor, target) {
				return Result{Message: "Zaten aktif bir ticaret anlaşması var."}
			}
			ensureTradeRoutesBetween(gs, actor, target)
			return Result{Accepted: true, Applied: true, Message: factionLabel(gs, target) + " ile eksik ticaret rotaları yeniden kuruldu."}
		}
		if rel.Stance == faction.StanceAllied && HasTradeRouteBetween(gs, actor, target) {
			return Result{Message: "Bu müttefik ile ticaret zaten aktif."}
		}
		if !acceptTrade(gs, rel, actor, target) {
			return Result{Message: factionLabel(gs, target) + " ticaret teklifini reddetti."}
		}
		prevStance := rel.Stance
		if prevStance != faction.StanceAllied {
			rel.Stance = faction.StanceTrade
		}
		rel.Score = clamp(rel.Score+15, -100, 100)
		ensureTradeRoutesBetween(gs, actor, target)
		if prevStance == faction.StanceAllied {
			return Result{Accepted: true, Applied: true, Message: factionLabel(gs, target) + " ile müttefiklik korunarak ticaret anlaşması açıldı."}
		}
		return Result{Accepted: true, Applied: true, Message: factionLabel(gs, target) + " ile ticaret anlaşması imzalandı."}

	case ActionProposeAlliance:
		if !acceptAlliance(gs, rel, actor, target) {
			return Result{Message: factionLabel(gs, target) + " ittifak teklifini reddetti."}
		}
		rel.Stance = faction.StanceAllied
		rel.Score = clamp(rel.Score+20, -100, 100)
		return Result{Accepted: true, Applied: true, Message: factionLabel(gs, target) + " ile ittifak kuruldu."}

	case ActionCancelAlliance:
		hasTrade := HasTradeRouteBetween(gs, actor, target)
		if hasTrade {
			rel.Stance = faction.StanceTrade
		} else {
			rel.Stance = faction.StancePeace
		}
		rel.Score = clamp(rel.Score-15, -100, 100)
		return Result{Accepted: true, Applied: true, Message: factionLabel(gs, target) + " ile ittifak sona erdirildi."}

	case ActionCancelTrade:
		removeTradeRoutesBetween(gs, actor, target)
		if rel.Stance == faction.StanceTrade {
			rel.Stance = faction.StancePeace
		}
		rel.Score = clamp(rel.Score-5, -100, 100)
		return Result{Accepted: true, Applied: true, Message: factionLabel(gs, target) + " ile ticaret anlaşması sona erdirildi."}

	case ActionImproveRelations:
		return applyRelationImprovement(gs, actor, target, relationImprovementCost, relationImprovementBonus, 0, "diplomatik heyet")

	case ActionSendGift:
		return applyRelationImprovement(gs, actor, target, giftCost, giftRelationBonus, giftReceiverGold, "hediye")

	case ActionOfferVassalization:
		if !AssessVassalizationProposal(gs, rel, actor, target).Accepted() {
			return Result{Message: factionLabel(gs, target) + " vassallık teklifini reddetti."}
		}
		return applyVassalization(gs, actor, target)

	case ActionReleaseVassal:
		return releaseVassalage(gs, actor, target)
	}

	return Result{Message: "Bilinmeyen diplomasi aksiyonu."}
}

func EnsureRelation(gs *state.GameState, a, b faction.FactionID) *faction.Relation {
	if gs.Relations == nil {
		gs.Relations = make(map[string]*faction.Relation)
	}
	key := faction.RelationKey(a, b)
	if rel := gs.Relations[key]; rel != nil {
		return rel
	}
	rel := &faction.Relation{
		FactionA: a,
		FactionB: b,
		Score:    0,
		Stance:   faction.StancePeace,
	}
	gs.Relations[key] = rel
	return rel
}

func Relation(gs *state.GameState, a, b faction.FactionID) *faction.Relation {
	if gs == nil {
		return nil
	}
	return gs.Relations[faction.RelationKey(a, b)]
}

func IsWar(gs *state.GameState, a, b faction.FactionID) bool {
	rel := Relation(gs, a, b)
	return rel != nil && rel.Stance == faction.StanceWar
}

func ForceRelation(gs *state.GameState, a, b faction.FactionID, stance faction.DiplomaticStance, scoreDelta int) {
	if gs == nil || a == "" || b == "" || a == b {
		return
	}
	rel := EnsureRelation(gs, a, b)
	prevStance := rel.Stance
	rel.Score = clamp(rel.Score+scoreDelta, -100, 100)
	if stance != "" {
		rel.Stance = stance
	}
	switch rel.Stance {
	case faction.StanceWar, faction.StancePeace:
		removeTradeRoutesBetween(gs, a, b)
	case faction.StanceTrade:
		ensureTradeRoutesBetween(gs, a, b)
	}
	if prevStance == faction.StanceTrade && (rel.Stance == faction.StanceWar || rel.Stance == faction.StancePeace) {
		removeTradeRoutesBetween(gs, a, b)
	}
	if prevStance != faction.StanceWar && rel.Stance == faction.StanceWar {
		gs.BeginWarLedger(a, b)
	} else if prevStance == faction.StanceWar && rel.Stance != faction.StanceWar {
		gs.EndWarLedger(a, b)
	}
}

func ApplyRelationDecay(gs *state.GameState) {
	for _, rel := range gs.Relations {
		if rel == nil {
			continue
		}
		switch rel.Stance {
		case faction.StanceWar:
			rel.Score = clamp(rel.Score-1, -100, 100)
		case faction.StancePeace:
			if rel.Score < 0 {
				rel.Score++
			}
		case faction.StanceTrade:
			if rel.Score < 30 {
				rel.Score++
			}
		case faction.StanceAllied:
			if SameRealm(gs, rel.FactionA, rel.FactionB) {
				if rel.Score < 50 {
					rel.Score++
				}
				continue
			}
			if HasDirectThreat(gs, rel.FactionA, rel.FactionB) && !HasCommonEnemy(gs, rel.FactionA, rel.FactionB) && !HasSharedMajorThreat(gs, rel.FactionA, rel.FactionB) {
				rel.Score = clamp(rel.Score-2, -100, 100)
				continue
			}
			if allianceHasStrategicBasis(gs, rel.FactionA, rel.FactionB) {
				if rel.Score < 50 {
					rel.Score++
				}
				continue
			}
			if rel.Score > 20 {
				rel.Score--
			} else if rel.Score < 20 {
				rel.Score++
			}
		}
	}
}

func EnsureTradeRoutesForActiveRelations(gs *state.GameState) {
	if gs == nil || len(gs.Relations) == 0 {
		return
	}
	SanitizeTradeRoutes(gs)
	for _, rel := range gs.Relations {
		if rel == nil {
			continue
		}
		if rel.Stance != faction.StanceTrade {
			continue
		}
		if !SameRealm(gs, rel.FactionA, rel.FactionB) && !CanEstablishTradeRoute(gs, rel.FactionA, rel.FactionB) {
			continue
		}
		ensureTradeRoutesBetween(gs, rel.FactionA, rel.FactionB)
	}
}

func SanitizeTradeRoutes(gs *state.GameState) {
	if gs == nil || len(gs.TradeRoutes) == 0 {
		return
	}
	filtered := gs.TradeRoutes[:0]
	seen := make(map[string]struct{}, len(gs.TradeRoutes))
	for _, route := range gs.TradeRoutes {
		if route == nil || route.FromFactionID == "" || route.ToFactionID == "" || route.FromFactionID == route.ToFactionID {
			continue
		}
		fromID := faction.FactionID(route.FromFactionID)
		toID := faction.FactionID(route.ToFactionID)
		fromFaction := gs.Factions[fromID]
		toFaction := gs.Factions[toID]
		if fromFaction == nil || toFaction == nil || fromFaction.IsEliminated || toFaction.IsEliminated {
			continue
		}
		if !relationAllowsTrade(Relation(gs, fromID, toID)) {
			continue
		}
		if !SameRealm(gs, fromID, toID) && !CanEstablishTradeRoute(gs, fromID, toID) {
			continue
		}
		key := route.FromFactionID + "->" + route.ToFactionID
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		filtered = append(filtered, route)
	}
	gs.TradeRoutes = filtered
}

func MilitaryPower(gs *state.GameState, fid faction.FactionID) int {
	total := 0
	for _, a := range gs.Armies {
		if a == nil || a.OwnerID != string(fid) {
			continue
		}
		if gs.UnitTypes != nil {
			total += a.TotalStrength(gs.UnitTypes)
			continue
		}
		total += len(a.Units) * 10
	}
	return total
}

func HasCommonEnemy(gs *state.GameState, a, b faction.FactionID) bool {
	for otherID := range gs.Factions {
		if otherID == a || otherID == b {
			continue
		}
		if IsWar(gs, a, otherID) && IsWar(gs, b, otherID) {
			return true
		}
	}
	return false
}

func baseRelationScore(gs *state.GameState, a, b faction.FactionID) int {
	if gs == nil {
		return 0
	}
	af := gs.Factions[a]
	bf := gs.Factions[b]
	if af == nil || bf == nil {
		return 0
	}
	return religion.Relation(af.Religion, bf.Religion)
}

func HasDiplomaticContact(gs *state.GameState, a, b faction.FactionID) bool {
	if gs == nil || a == "" || b == "" || a == b {
		return false
	}
	if rel := Relation(gs, a, b); rel != nil {
		if rel.Stance != faction.StancePeace {
			return true
		}
		if rel.Score != baseRelationScore(gs, a, b) {
			return true
		}
	}
	if sharesBorder(gs, a, b) {
		return true
	}
	if HasCommonEnemy(gs, a, b) || HasSharedMajorThreat(gs, a, b) {
		return true
	}
	if HasTradeRouteBetween(gs, a, b) || CanEstablishTradeRoute(gs, a, b) {
		return true
	}
	return false
}

func CanEstablishTradeRoute(gs *state.GameState, a, b faction.FactionID) bool {
	if gs == nil || a == "" || b == "" || a == b {
		return false
	}
	if SameRealm(gs, a, b) {
		return true
	}
	return canEstablishLandTradeRoute(gs, a, b) || canEstablishSeaTradeRoute(gs, a, b)
}

func allianceHasStrategicBasis(gs *state.GameState, a, b faction.FactionID) bool {
	if gs == nil || a == "" || b == "" || a == b {
		return false
	}
	if sharesBorder(gs, a, b) {
		return true
	}
	if HasTradeRouteBetween(gs, a, b) || CanEstablishTradeRoute(gs, a, b) {
		return true
	}
	if HasCommonEnemy(gs, a, b) {
		return true
	}
	if HasSharedMajorThreat(gs, a, b) {
		return true
	}
	return false
}

func canEstablishLandTradeRoute(gs *state.GameState, a, b faction.FactionID) bool {
	if gs == nil {
		return false
	}
	targets := make(map[world.RegionID]struct{})
	queue := make([]world.RegionID, 0, len(gs.Regions))
	seen := make(map[world.RegionID]struct{}, len(gs.Regions))
	for rid, region := range gs.Regions {
		if !tradeLandRegionAnchor(region) {
			continue
		}
		owner := faction.FactionID(region.OwnerID)
		switch {
		case SameRealm(gs, owner, a):
			queue = append(queue, rid)
			seen[rid] = struct{}{}
		case SameRealm(gs, owner, b):
			targets[rid] = struct{}{}
		}
	}
	if len(queue) == 0 || len(targets) == 0 {
		return false
	}
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]
		if _, ok := targets[currentID]; ok {
			return true
		}
		current := gs.Regions[currentID]
		if current == nil {
			continue
		}
		for _, neighborID := range current.Neighbors {
			if _, ok := seen[neighborID]; ok {
				continue
			}
			neighbor := gs.Regions[neighborID]
			if !tradeLandRegionPassable(gs, neighbor, a, b) {
				continue
			}
			seen[neighborID] = struct{}{}
			queue = append(queue, neighborID)
		}
	}
	return false
}

func tradeLandRegionAnchor(region *world.Region) bool {
	return region != nil && !region.IsSea && !region.IsLocked && region.OwnerID != "" && region.TradeCapacity > 0
}

func tradeLandRegionPassable(gs *state.GameState, region *world.Region, a, b faction.FactionID) bool {
	if region == nil || region.IsSea || region.IsLocked || region.OwnerID == "" {
		return false
	}
	owner := faction.FactionID(region.OwnerID)
	return SameRealm(gs, owner, a) || SameRealm(gs, owner, b)
}

func canEstablishSeaTradeRoute(gs *state.GameState, a, b faction.FactionID) bool {
	if gs == nil {
		return false
	}
	queue := make([]world.RegionID, 0, len(gs.Regions))
	seen := make(map[world.RegionID]struct{}, len(gs.Regions))
	targets := make(map[world.RegionID]struct{})
	for rid, region := range gs.Regions {
		if region == nil || region.IsSea || region.IsLocked || !region.HasPort() {
			continue
		}
		owner := faction.FactionID(region.OwnerID)
		seaNeighbors := adjacentSeaRegions(gs, rid)
		switch {
		case SameRealm(gs, owner, a):
			for _, seaID := range seaNeighbors {
				if _, ok := seen[seaID]; ok {
					continue
				}
				seen[seaID] = struct{}{}
				queue = append(queue, seaID)
			}
		case SameRealm(gs, owner, b):
			for _, seaID := range seaNeighbors {
				targets[seaID] = struct{}{}
			}
		}
	}
	if len(queue) == 0 || len(targets) == 0 {
		return false
	}
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]
		if _, ok := targets[currentID]; ok {
			return true
		}
		current := gs.Regions[currentID]
		if current == nil || !current.IsSea {
			continue
		}
		for _, neighborID := range current.Neighbors {
			if _, ok := seen[neighborID]; ok {
				continue
			}
			neighbor := gs.Regions[neighborID]
			if neighbor == nil || !neighbor.IsSea || neighbor.IsLocked {
				continue
			}
			seen[neighborID] = struct{}{}
			queue = append(queue, neighborID)
		}
	}
	return false
}

func adjacentSeaRegions(gs *state.GameState, regionID world.RegionID) []world.RegionID {
	region := gs.Regions[regionID]
	if region == nil {
		return nil
	}
	out := make([]world.RegionID, 0, len(region.Neighbors))
	for _, neighborID := range region.Neighbors {
		neighbor := gs.Regions[neighborID]
		if neighbor != nil && neighbor.IsSea && !neighbor.IsLocked {
			out = append(out, neighborID)
		}
	}
	return out
}

func HasSharedMajorThreat(gs *state.GameState, a, b faction.FactionID) bool {
	if gs == nil || a == "" || b == "" || a == b {
		return false
	}

	// Build the ownership adjacency and land-count snapshot once. The previous
	// implementation called sharesBorder and landRegionCount again for every
	// candidate threat, which made alliance scans quadratic in the region count.
	landCounts := make(map[faction.FactionID]int, len(gs.Factions))
	borders := make(map[faction.FactionID]map[faction.FactionID]struct{}, len(gs.Factions))
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID == "" {
			continue
		}
		owner := faction.FactionID(region.OwnerID)
		landCounts[owner]++
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil || neighbor.IsSea || neighbor.OwnerID == "" {
				continue
			}
			if borders[owner] == nil {
				borders[owner] = make(map[faction.FactionID]struct{})
			}
			borders[owner][faction.FactionID(neighbor.OwnerID)] = struct{}{}
		}
	}

	powers := make(map[faction.FactionID]int, len(gs.Factions))
	powerOf := func(fid faction.FactionID) int {
		if power, ok := powers[fid]; ok {
			return power
		}
		power := MilitaryPower(gs, fid)
		powers[fid] = power
		return power
	}
	sharesBorderSnapshot := func(left, right faction.FactionID) bool {
		_, ok := borders[left][right]
		return ok
	}
	isMajorThreatSnapshot := func(threat, target faction.FactionID) bool {
		if threat == "" || target == "" || threat == target {
			return false
		}
		threatFaction := gs.Factions[threat]
		targetFaction := gs.Factions[target]
		if threatFaction == nil || targetFaction == nil || threatFaction.IsEliminated || targetFaction.IsEliminated {
			return false
		}
		if !sharesBorderSnapshot(threat, target) && !IsWar(gs, threat, target) {
			return false
		}

		threatPower := powerOf(threat)
		targetPower := powerOf(target)
		powerThreat := false
		switch {
		case threatPower > 0 && targetPower == 0:
			powerThreat = true
		case targetPower > 0 && threatPower > max(targetPower*13/10, targetPower+15):
			powerThreat = true
		}
		return powerThreat || landCounts[threat] > landCounts[target]+2
	}

	for otherID, other := range gs.Factions {
		if otherID == a || otherID == b || other == nil || other.IsEliminated {
			continue
		}
		if isMajorThreatSnapshot(otherID, a) && isMajorThreatSnapshot(otherID, b) {
			return true
		}
	}
	return false
}

func HasDirectThreat(gs *state.GameState, a, b faction.FactionID) bool {
	if !sharesBorder(gs, a, b) {
		return false
	}
	powerA := MilitaryPower(gs, a)
	powerB := MilitaryPower(gs, b)
	if powerA == 0 || powerB == 0 {
		return powerA != powerB
	}
	if powerA > powerB*12/10 || powerB > powerA*12/10 {
		return true
	}
	return frontierArmyCount(gs, a, b) > frontierArmyCount(gs, b, a)+1 ||
		frontierArmyCount(gs, b, a) > frontierArmyCount(gs, a, b)+1
}

func acceptPeace(gs *state.GameState, rel *faction.Relation, actor, target faction.FactionID) bool {
	if gs != nil && gs.ScenarioID == "1300_ottoman_rise" {
		return AssessPeaceDesire(gs, target, actor).ShouldPropose()
	}
	warPressure := 0
	if rel.Score < -80 {
		warPressure = -rel.Score - 80
	}

	actorPower := MilitaryPower(gs, actor)
	targetPower := MilitaryPower(gs, target)
	strengthPressure := 0
	if actorPower > targetPower {
		strengthPressure += min(25, (actorPower-targetPower)/8)
	} else if targetPower > actorPower {
		strengthPressure -= min(10, (targetPower-actorPower)/12)
	}

	actorRegions := len(gs.RegionsOwnedBy(actor))
	targetRegions := len(gs.RegionsOwnedBy(target))
	if actorRegions > targetRegions {
		strengthPressure += min(20, (actorRegions-targetRegions)*4)
	}

	return warPressure+strengthPressure+economicStress(gs, target)+peaceTechBonus(gs, actor) >= 18
}

func acceptTrade(gs *state.GameState, rel *faction.Relation, actor, target faction.FactionID) bool {
	assessment := AssessTradeProposal(gs, rel, actor, target)
	return assessment.Accepted()
}

func AssessAllianceProposal(gs *state.GameState, rel *faction.Relation, actor, target faction.FactionID) AllianceProposalAssessment {
	assessment := AllianceProposalAssessment{}
	if gs == nil || rel == nil || actor == "" || target == "" || actor == target {
		assessment.BlockReason = "Geçersiz diplomasi hedefi"
		return assessment
	}
	if rel.Score < allianceRelationThreshold {
		assessment.BlockReason = "İttifak için ilişki puanı 25 altı"
		return assessment
	}
	actorFaction := gs.Factions[actor]
	targetFaction := gs.Factions[target]
	if actorFaction == nil || targetFaction == nil {
		assessment.BlockReason = "Fraksiyon bulunamadı"
		return assessment
	}
	if gs.ScenarioID == "1300_ottoman_rise" && activeAllianceObjectiveConflict(gs, actor, target) {
		assessment.ActorStrategic.ActiveObjectiveConflict = true
		assessment.TargetStrategic.ActiveObjectiveConflict = true
		assessment.BlockReason = "Aktif stratejik hedefler ittifakla çakışıyor"
		return assessment
	}
	sharesLandBorder := sharesBorder(gs, actor, target)
	hasTradeRoute := HasTradeRouteBetween(gs, actor, target)
	hasTradeAccess := hasTradeRoute || CanEstablishTradeRoute(gs, actor, target)
	commonEnemy := HasCommonEnemy(gs, actor, target)
	sharedMajorThreat := HasSharedMajorThreat(gs, actor, target)
	if !sharesLandBorder && !hasTradeAccess && !commonEnemy && !sharedMajorThreat {
		assessment.BlockReason = "İttifak için coğrafi veya stratejik yakınlık yok"
		return assessment
	}
	if gs.ScenarioID == "1300_ottoman_rise" {
		// Trade reachability is a pair-level fact. Reuse the result for both
		// strategic perspectives instead of running the land/sea BFS twice.
		assessment.ActorStrategic = assessStrategicAllianceWithTrade(gs, actor, target, commonEnemy, sharedMajorThreat, hasTradeAccess)
		assessment.TargetStrategic = assessStrategicAllianceWithTrade(gs, target, actor, commonEnemy, sharedMajorThreat, hasTradeAccess)
		if assessment.ActorStrategic.ActiveObjectiveConflict || assessment.TargetStrategic.ActiveObjectiveConflict {
			assessment.BlockReason = "Aktif stratejik hedefler ittifakla çakışıyor"
			return assessment
		}
		if target != gs.PlayerFactionID && assessment.TargetStrategic.Score < strategicAllianceAcceptanceFloor {
			assessment.BlockReason = "İttifak hedef devlet için yeterli stratejik değer üretmiyor"
			return assessment
		}
	}
	actorPower := MilitaryPower(gs, actor)
	targetPower := MilitaryPower(gs, target)
	actorRegions := landRegionCount(gs, actor)
	targetRegions := landRegionCount(gs, target)

	chance := 20 + rel.Score
	if rel.Stance == faction.StanceTrade {
		chance += 8
	}
	chance += allianceReligionAffinityBonus(actorFaction.Religion, targetFaction.Religion)
	if sharesLandBorder {
		chance += 6
	}
	if commonEnemy {
		chance += 12
	}
	if sharedMajorThreat {
		chance += 15
	}
	if !sharesLandBorder && !hasTradeRoute {
		chance -= 10
	}
	if HasDirectThreat(gs, actor, target) {
		chance -= 15
	}
	if gs.ScenarioID == "1300_ottoman_rise" {
		strategic := assessment.TargetStrategic
		chance += strategic.BufferValue/2 + strategic.FrontSupportValue/2
		chance += strategic.TradeValue/3 + strategic.PartnerSupportValue/3
		chance -= strategic.ExpansionTensionPenalty
	}
	if actorPower > targetPower {
		chance += min(10, (actorPower-targetPower)/15)
	}
	chance += clamp((actorRegions-targetRegions)*2, -6, 10)

	assessment.Chance = clamp(chance, 0, 100)
	return assessment
}

func allianceReligionAffinityBonus(a, b religion.Type) int {
	if a == "" || b == "" {
		return 0
	}
	switch {
	case a == b:
		return 8
	case (a == religion.Catholic && b == religion.Orthodox) || (a == religion.Orthodox && b == religion.Catholic):
		return 2
	case (a == religion.Sunni && b == religion.Shia) || (a == religion.Shia && b == religion.Sunni):
		return -8
	default:
		return -4
	}
}

func AssessTradeProposal(gs *state.GameState, rel *faction.Relation, actor, target faction.FactionID) TradeProposalAssessment {
	assessment := TradeProposalAssessment{}
	if gs == nil || rel == nil || actor == "" || target == "" || actor == target {
		assessment.BlockReason = "Geçersiz diplomasi hedefi"
		return assessment
	}
	if rel.Score < tradeRelationThreshold {
		assessment.BlockReason = "Ticaret için ilişki puanı 15 altı"
		return assessment
	}
	actorLand := landRegionCount(gs, actor)
	if actorLand == 0 {
		assessment.BlockReason = "Sende kara bölgesi yok"
		return assessment
	}
	targetLand := landRegionCount(gs, target)
	if targetLand == 0 {
		assessment.BlockReason = "Hedefin kara bölgesi yok"
		return assessment
	}
	actorCap := totalTradeCapacity(gs, actor)
	if actorCap < 4 {
		assessment.BlockReason = "Senin ticaret kapasiten 4 altı"
		return assessment
	}
	targetCap := totalTradeCapacity(gs, target)
	if targetCap < 4 {
		assessment.BlockReason = "Hedefin ticaret kapasitesi 4 altı"
		return assessment
	}
	actorPartners := activeTradePartners(gs, actor)
	if actorPartners >= 4 {
		assessment.BlockReason = "Senin aktif partner sınırın dolu"
		return assessment
	}
	targetPartners := activeTradePartners(gs, target)
	if targetPartners >= 4 {
		assessment.BlockReason = "Hedefin aktif partner sınırı dolu"
		return assessment
	}
	if !CanEstablishTradeRoute(gs, actor, target) {
		assessment.BlockReason = "Ticaret için bağlanabilir kara veya deniz hattı yok"
		return assessment
	}

	regionDelta := actorLand - targetLand
	chance := 40 + rel.Score + clamp(regionDelta, -10, 20)
	if rel.Stance == faction.StanceAllied {
		chance += 8
	}
	if HasCommonEnemy(gs, actor, target) {
		chance += 6
	}
	if HasDirectThreat(gs, actor, target) {
		chance -= 25
	}
	if actorPartners == 3 {
		chance -= 6
	}
	if targetPartners == 3 {
		chance -= 6
	}
	assessment.Chance = clamp(chance, 0, 100)
	return assessment
}

func acceptAlliance(gs *state.GameState, rel *faction.Relation, actor, target faction.FactionID) bool {
	assessment := AssessAllianceProposal(gs, rel, actor, target)
	return assessment.Accepted()
}

func peaceTechBonus(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || gs.TechTypes == nil || fid == "" {
		return 0
	}
	f := gs.Factions[fid]
	if f == nil {
		return 0
	}
	return tech.ComputeEffects(f.Research.Completed, gs.TechTypes).PeaceRelationBonus
}

func ensureTradeRoutesBetween(gs *state.GameState, a, b faction.FactionID) {
	removeTradeRoutesBetween(gs, a, b)
	routeAB := buildTradeRoute(gs, a, b)
	routeBA := buildTradeRoute(gs, b, a)
	gs.TradeRoutes = append(gs.TradeRoutes, routeAB, routeBA)
}

func removeTradeRoutesBetween(gs *state.GameState, a, b faction.FactionID) {
	if len(gs.TradeRoutes) == 0 {
		return
	}
	filtered := gs.TradeRoutes[:0]
	aStr := string(a)
	bStr := string(b)
	for _, route := range gs.TradeRoutes {
		if route == nil {
			continue
		}
		if (route.FromFactionID == aStr && route.ToFactionID == bStr) ||
			(route.FromFactionID == bStr && route.ToFactionID == aStr) {
			continue
		}
		filtered = append(filtered, route)
	}
	gs.TradeRoutes = filtered
}

func HasTradeRouteBetween(gs *state.GameState, a, b faction.FactionID) bool {
	if gs == nil || len(gs.TradeRoutes) == 0 || a == "" || b == "" || a == b {
		return false
	}
	aStr := string(a)
	bStr := string(b)
	for _, route := range gs.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 {
			continue
		}
		if (route.FromFactionID == aStr && route.ToFactionID == bStr) ||
			(route.FromFactionID == bStr && route.ToFactionID == aStr) {
			return true
		}
	}
	return false
}

func relationAllowsTrade(rel *faction.Relation) bool {
	return rel != nil && (rel.Stance == faction.StanceTrade || rel.Stance == faction.StanceAllied)
}

func buildTradeRoute(gs *state.GameState, from, to faction.FactionID) *economy.TradeRoute {
	good := chooseTradeRouteGood(gs, from, to)
	return &economy.TradeRoute{
		FromFactionID: string(from),
		ToFactionID:   string(to),
		Good:          good,
		AmountPerTurn: tradeAmount(gs, from, to),
		GoldPerUnit:   economy.BaseGoldValue[good],
	}
}

// chooseTradeRouteGood, hedefin gerçek tahıl talebi varsa ve kaynakta rezerv
// üstü stok bulunuyorsa rotayı tahıla yönlendirir. Talep yoksa mevcut yüksek
// değerli ihracat seçimi korunur.
func chooseTradeRouteGood(gs *state.GameState, from, to faction.FactionID) economy.GoodType {
	if gs != nil && gs.StrategicGrainDemand(to) > 0 && gs.StrategicGrainSurplus(from) > 0 {
		return economy.GoodGrain
	}
	return chooseExportGood(gs, from)
}

func chooseExportGood(gs *state.GameState, fid faction.FactionID) economy.GoodType {
	f := gs.Factions[fid]
	if f == nil {
		return economy.GoodGrain
	}
	type goodStock struct {
		good  economy.GoodType
		stock int
	}
	options := []goodStock{
		{economy.GoodSpice, f.Spice},
		{economy.GoodCloth, f.Cloth},
		{economy.GoodIron, f.Iron},
		{economy.GoodTimber, f.Timber},
		{economy.GoodGrain, f.Grain},
	}
	best := options[len(options)-1].good
	bestScore := -1
	for _, option := range options {
		score := option.stock * economy.BaseGoldValue[option.good]
		if score > bestScore {
			bestScore = score
			best = option.good
		}
	}
	return best
}

func tradeAmount(gs *state.GameState, a, b faction.FactionID) int {
	capA := totalTradeCapacity(gs, a)
	capB := totalTradeCapacity(gs, b)
	capacity := min(capA, capB)
	switch {
	case capacity <= 0:
		return 1
	case capacity >= 8:
		return 4
	case capacity >= 5:
		return 3
	case capacity >= 2:
		return 2
	default:
		return 1
	}
}

func totalTradeCapacity(gs *state.GameState, fid faction.FactionID) int {
	total := 0
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		total += region.TradeCapacity
	}
	return total
}

func activeTradePartners(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || len(gs.TradeRoutes) == 0 || fid == "" {
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

func landRegionCount(gs *state.GameState, fid faction.FactionID) int {
	count := 0
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		count++
	}
	return count
}

func sharesBorder(gs *state.GameState, a, b faction.FactionID) bool {
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

func frontierArmyCount(gs *state.GameState, owner, against faction.FactionID) int {
	count := 0
	for _, armyRef := range gs.Armies {
		if armyRef == nil || armyRef.OwnerID != string(owner) || armyRef.IsNaval {
			continue
		}
		region := gs.Regions[armyRef.RegionID]
		if region == nil {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor != nil && neighbor.OwnerID == string(against) {
				count++
				break
			}
		}
	}
	return count
}

func isMajorThreatTo(gs *state.GameState, threat, target faction.FactionID) bool {
	if gs == nil || threat == "" || target == "" || threat == target {
		return false
	}
	threatFaction := gs.Factions[threat]
	targetFaction := gs.Factions[target]
	if threatFaction == nil || targetFaction == nil || threatFaction.IsEliminated || targetFaction.IsEliminated {
		return false
	}
	if !sharesBorder(gs, threat, target) && !IsWar(gs, threat, target) {
		return false
	}

	threatPower := MilitaryPower(gs, threat)
	targetPower := MilitaryPower(gs, target)
	threatRegions := landRegionCount(gs, threat)
	targetRegions := landRegionCount(gs, target)

	powerThreat := false
	switch {
	case threatPower > 0 && targetPower == 0:
		powerThreat = true
	case targetPower > 0 && threatPower > max(targetPower*13/10, targetPower+15):
		powerThreat = true
	}
	landThreat := threatRegions > targetRegions+2
	return powerThreat || landThreat
}

func economicStress(gs *state.GameState, fid faction.FactionID) int {
	f := gs.Factions[fid]
	if f == nil {
		return 0
	}
	stress := 0
	if f.Gold < 80 {
		stress += 8
	}
	if f.Grain < 40 {
		stress += 8
	}
	if landRegionCount(gs, fid) <= 2 {
		stress += 6
	}
	return stress
}

func factionLabel(gs *state.GameState, fid faction.FactionID) string {
	if f := gs.Factions[fid]; f != nil {
		if f.NameTR != "" {
			return f.NameTR
		}
		if f.Name != "" {
			return f.Name
		}
	}
	return string(fid)
}

func clamp(v, minValue, maxValue int) int {
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
