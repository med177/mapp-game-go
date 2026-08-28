package diplomacy

import (
	"sort"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

type Action string

const (
	ActionDeclareWar                Action = "declare_war"
	ActionJoinWarCall               Action = "join_war_call"
	ActionProposePeace              Action = "propose_peace"
	ActionProposeAlliance           Action = "propose_alliance"
	ActionProposeTrade              Action = "propose_trade"
	ActionProposeSurrender          Action = "propose_surrender"
	ActionProposeSiegeVassalization Action = "propose_siege_vassalization"
	ActionCancelAlliance            Action = "cancel_alliance"
	ActionCancelTrade               Action = "cancel_trade"
	ActionImproveRelations          Action = "improve_relations"
	ActionSendGift                  Action = "send_gift"
	ActionOfferVassalization        Action = "offer_vassalization"
	ActionReleaseVassal             Action = "release_vassal"
	ActionAnnexVassal               Action = "annex_vassal"
)

type Result struct {
	Accepted   bool
	Applied    bool
	Message    string
	Settlement *PeaceSettlement
}

const tradeAcceptanceThreshold = 45
const tradeRelationThreshold = 15

// MaxTradePartners dış ticaret anlaşmalarında devlet başına verilen temel aktif
// partner sayısıdır. Tam geliştirilmiş liman+pazar bölgeleri bu tabana eklenir.
// Aynı realm içindeki overlord-vassal rotaları bu kota ve rota kapasitesi
// rezervinden muaftır.
const MaxTradePartners = 4

// MaxTradeRouteAmountPerTurn tam geliştirme bonusları öncesindeki temel rota
// hacmi tavanıdır.
const MaxTradeRouteAmountPerTurn = economy.MaxTradeRouteAmountPerTurn

const maxTradeRouteMarketBonus = 2

// TradePartnerLimit devletin sahip olduğu, hem limanı hem pazarı bina tanımındaki
// maksimum seviyeye ulaşmış her bölge için bir ek dış ticaret partneri verir.
func TradePartnerLimit(gs *state.GameState, fid faction.FactionID) int {
	limit := MaxTradePartners
	if gs == nil || fid == "" {
		return limit
	}
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.IsLocked || region.OwnerID != string(fid) {
			continue
		}
		if buildingAtMaxLevel(gs, region, "port") && buildingAtMaxLevel(gs, region, "market") {
			limit++
		}
	}
	return limit
}

// TradeRouteAmountLimit devletin her maksimum seviyedeki pazar bölgesi için
// anlaşma başına temel rota tavanına iki hacim ekler.
func TradeRouteAmountLimit(gs *state.GameState, fid faction.FactionID) int {
	limit := MaxTradeRouteAmountPerTurn
	if gs == nil || fid == "" {
		return limit
	}
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.IsLocked || region.OwnerID != string(fid) {
			continue
		}
		if buildingAtMaxLevel(gs, region, "market") {
			limit += maxTradeRouteMarketBonus
		}
	}
	return limit
}

// TradeAgreementAmountLimit iki tarafın da karşılayabildiği anlaşma hacmi
// tavanını döner. Hacim, taraflardan zayıf olanın tavanını aşamaz.
func TradeAgreementAmountLimit(gs *state.GameState, a, b faction.FactionID) int {
	limitA := TradeRouteAmountLimit(gs, a)
	limitB := TradeRouteAmountLimit(gs, b)
	if limitA < limitB {
		return limitA
	}
	return limitB
}

func buildingAtMaxLevel(gs *state.GameState, region *world.Region, buildingID string) bool {
	if gs == nil || region == nil || buildingID == "" {
		return false
	}
	building := gs.BuildingTypes[buildingID]
	return building != nil && building.MaxPerRegion > 0 && region.BuildingLevel(buildingID) >= building.MaxPerRegion
}

// rejectedOfferRelationPenalty her reddedilen normal diplomasi teklifinin
// ilişkiye uyguladığı küçük cezadır. Savaş çağrısı kendi özel sonucunu kullanır.
const rejectedOfferRelationPenalty = 3

type TradeProposalAssessment struct {
	Chance      int
	BlockReason string
}

func (a TradeProposalAssessment) Accepted() bool {
	return a.BlockReason == "" && a.Chance >= tradeAcceptanceThreshold
}

const allianceAcceptanceThreshold = 45
const allianceRelationThreshold = 25
const historicalAllianceRelationThreshold = 40

func allianceRelationThresholdFor(gs *state.GameState) int {
	if gs != nil && gs.ScenarioID == "1300_ottoman_rise" {
		return historicalAllianceRelationThreshold
	}
	return allianceRelationThreshold
}

func allianceRelationBlockReason(gs *state.GameState) string {
	if allianceRelationThresholdFor(gs) == historicalAllianceRelationThreshold {
		return "İttifak için ilişki puanı 40 altı"
	}
	return "İttifak için ilişki puanı 25 altı"
}

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
	return execute(gs, actor, target, action, true)
}

// execute bir diplomasi aksiyonunu uygular. Kuyruktaki teklif çözülürken
// consumeQuota false olur; gönderenin kotası teklif kuyruğa alınırken harcanmıştır.
func execute(gs *state.GameState, actor, target faction.FactionID, action Action, consumeQuota bool) Result {
	if reason := actionBlockReason(gs, actor, target, action, consumeQuota); reason != "" {
		return Result{Message: reason}
	}
	if consumeQuota && actionUsesDiplomacyOfferQuota(action) && !spendDiplomacyOfferQuota(gs, actor) {
		return Result{Message: diplomacyOfferQuotaBlockReasonTR}
	}

	rel := EnsureRelation(gs, actor, target)
	switch action {
	case ActionDeclareWar:
		return ExecuteWarDeclaration(gs, actor, target, nil).Result

	case ActionProposePeace:
		if !acceptPeace(gs, actor, target) {
			markRejectedDiplomaticOffer(gs, actor, target, action)
			return Result{Message: factionLabel(gs, target) + " barışı reddetti."}
		}
		settlement := AssessPeaceSettlement(gs, actor, target)
		setPeaceBetweenCoalitions(gs, actor, target)
		return Result{Accepted: true, Applied: true, Settlement: &settlement, Message: factionLabel(gs, target) + " barışı kabul etti."}

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
			markRejectedDiplomaticOffer(gs, actor, target, action)
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
			markRejectedDiplomaticOffer(gs, actor, target, action)
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
		return applyRelationImprovement(gs, actor, target, RelationImprovementGoldCost, relationImprovementBonus, 0, "diplomatik heyet")

	case ActionSendGift:
		return applyRelationImprovement(gs, actor, target, GiftGoldCost, giftRelationBonus, giftReceiverGold, "hediye")

	case ActionOfferVassalization:
		if !AssessVassalizationProposal(gs, rel, actor, target).Accepted() {
			markRejectedDiplomaticOffer(gs, actor, target, action)
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

// markRejectedDiplomaticOffer ret bilgisini kaydeder ve ilişkiyi küçük bir
// miktar azaltır. Retry kaydı AI'nin aynı teklifi her tur otomatik yinelemesini
// engeller; save/load ile birlikte kalıcıdır.
func markRejectedDiplomaticOffer(gs *state.GameState, actor, target faction.FactionID, action Action) {
	if gs == nil || actor == "" || target == "" || actor == target {
		return
	}
	// Bu retry/ilişki sonucu oyuncunun cevapladığı diplomasi akışına aittir.
	// AI-AI otomatik değerlendirmeleri senaryo tempo kalibrasyonunu etkilemez.
	if gs.PlayerFactionID == "" || (actor != gs.PlayerFactionID && target != gs.PlayerFactionID) {
		return
	}
	rel := EnsureRelation(gs, actor, target)
	rel.Score = clamp(rel.Score-rejectedOfferRelationPenalty, -100, 100)
	gs.MarkDiplomaticOfferRejected(string(actor), string(target), string(action))
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
	relationKeys := make([]string, 0, len(gs.Relations))
	for key := range gs.Relations {
		relationKeys = append(relationKeys, key)
	}
	sort.Strings(relationKeys)
	for _, key := range relationKeys {
		rel := gs.Relations[key]
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
	RebalanceTradeRouteCapacities(gs)
}

func SanitizeTradeRoutes(gs *state.GameState) {
	if gs == nil || len(gs.TradeRoutes) == 0 {
		return
	}
	validByAgreement := make(map[string][]*economy.TradeRoute)
	seenDirections := make(map[string]struct{}, len(gs.TradeRoutes))
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
		directionKey := route.AssignmentKey()
		if _, exists := seenDirections[directionKey]; exists {
			continue
		}
		seenDirections[directionKey] = struct{}{}
		key, _, _ := tradeAgreementKey(fromID, toID)
		validByAgreement[key] = append(validByAgreement[key], route)
	}

	agreementKeys := make([]string, 0, len(validByAgreement))
	for key := range validByAgreement {
		agreementKeys = append(agreementKeys, key)
	}
	sort.Strings(agreementKeys)
	partnerCount := make(map[faction.FactionID]int)
	filtered := make([]*economy.TradeRoute, 0, len(gs.TradeRoutes))
	for _, key := range agreementKeys {
		routes := validByAgreement[key]
		if len(routes) == 0 {
			continue
		}
		fromID := faction.FactionID(routes[0].FromFactionID)
		toID := faction.FactionID(routes[0].ToFactionID)
		_, left, right := tradeAgreementKey(fromID, toID)
		if !SameRealm(gs, left, right) {
			if partnerCount[left] >= TradePartnerLimit(gs, left) || partnerCount[right] >= TradePartnerLimit(gs, right) {
				continue
			}
			partnerCount[left]++
			partnerCount[right]++
		}
		filtered = append(filtered, routes...)
	}
	sortTradeRoutes(filtered)
	gs.TradeRoutes = filtered
	RebalanceTradeRouteCapacities(gs)
}

func sortTradeRoutes(routes []*economy.TradeRoute) {
	sort.SliceStable(routes, func(i, j int) bool {
		return routes[i].AssignmentKey() < routes[j].AssignmentKey()
	})
}

// MilitaryPowerBreakdown kara ve deniz ordularının etkin güçlerini ayrı döner.
// Birimin etkin saldırısı, can yüzdesi ve ordu morali Army.TotalStrength içinde
// hesaba katılır. Donanma gücü de devletin toplam askerî gücünün parçasıdır.
func MilitaryPowerBreakdown(gs *state.GameState, fid faction.FactionID) (land, naval int) {
	if gs == nil || fid == "" {
		return 0, 0
	}
	for _, a := range gs.Armies {
		if a == nil || a.OwnerID != string(fid) {
			continue
		}
		power := 0
		if gs.UnitTypes != nil {
			power = a.TotalStrength(gs.UnitTypes)
		} else {
			power = len(a.Units) * 10
		}
		if a.IsNaval {
			naval += power
		} else {
			land += power
		}
	}
	return land, naval
}

// MilitaryPower devletin kara ve deniz etkin güçlerinin toplamını döner.
func MilitaryPower(gs *state.GameState, fid faction.FactionID) int {
	land, naval := MilitaryPowerBreakdown(gs, fid)
	return land + naval
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

func acceptPeace(gs *state.GameState, actor, target faction.FactionID) bool {
	if assessment := AssessPeaceProposal(gs, actor, target); assessment.BlockReason == "" {
		return assessment.Accepted
	}
	return false
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
	if rel.Score < allianceRelationThresholdFor(gs) {
		assessment.BlockReason = allianceRelationBlockReason(gs)
		return assessment
	}
	actorFaction := gs.Factions[actor]
	targetFaction := gs.Factions[target]
	if actorFaction == nil || targetFaction == nil {
		assessment.BlockReason = "Fraksiyon bulunamadı"
		return assessment
	}
	if allyID, conflict := allianceWarConflictBetween(gs, actor, target); conflict {
		assessment.BlockReason = factionLabel(gs, allyID) + " ile savaş halinde olan devlete ittifak teklif edilemez"
		return assessment
	}
	if activeAllianceObjectiveConflict(gs, actor, target) {
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

	strategic := assessment.TargetStrategic
	chance += strategic.BufferValue/2 + strategic.FrontSupportValue/2
	chance += strategic.TradeValue/3 + strategic.PartnerSupportValue/3
	chance -= strategic.ExpansionTensionPenalty

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
	actorPartners := ActiveTradePartnerCount(gs, actor)
	if actorPartners >= TradePartnerLimit(gs, actor) {
		assessment.BlockReason = "Senin aktif partner sınırın dolu"
		return assessment
	}
	targetPartners := ActiveTradePartnerCount(gs, target)
	if targetPartners >= TradePartnerLimit(gs, target) {
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
	if gs == nil || a == "" || b == "" || a == b || !canMaintainOrAddTradePartner(gs, a, b) {
		return
	}
	removeTradeRoutesBetween(gs, a, b)
	routeAB := buildTradeRoute(gs, a, b)
	routeBA := buildTradeRoute(gs, b, a)
	gs.TradeRoutes = append(gs.TradeRoutes, routeAB, routeBA)
	sortTradeRoutes(gs.TradeRoutes)
	RebalanceTradeRouteCapacities(gs)
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
	sortTradeRoutes(filtered)
	gs.TradeRoutes = filtered
	RebalanceTradeRouteCapacities(gs)
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
	amountLimit := TradeAgreementAmountLimit(gs, a, b)
	if capacity <= 0 {
		return 1
	}
	if capacity > amountLimit {
		return amountLimit
	}
	return capacity
}

func totalTradeCapacity(gs *state.GameState, fid faction.FactionID) int {
	return gs.EffectiveFactionTradeCapacity(fid)
}

// ActiveTradePartnerCount aynı realm dışındaki, askıya alınmamış rota
// anlaşmalarından türeyen benzersiz partner sayısını döner.
func ActiveTradePartnerCount(gs *state.GameState, fid faction.FactionID) int {
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
			partnerID := faction.FactionID(route.ToFactionID)
			if !SameRealm(gs, fid, partnerID) {
				partners[route.ToFactionID] = struct{}{}
			}
		case route.ToFactionID == self && route.FromFactionID != "":
			partnerID := faction.FactionID(route.FromFactionID)
			if !SameRealm(gs, fid, partnerID) {
				partners[route.FromFactionID] = struct{}{}
			}
		}
	}
	return len(partners)
}

func canMaintainOrAddTradePartner(gs *state.GameState, a, b faction.FactionID) bool {
	if gs == nil || a == "" || b == "" || a == b || SameRealm(gs, a, b) {
		return gs != nil && a != "" && b != "" && a != b
	}
	if hasTradeRouteRecordBetween(gs, a, b) {
		return true
	}
	return ActiveTradePartnerCount(gs, a) < TradePartnerLimit(gs, a) && ActiveTradePartnerCount(gs, b) < TradePartnerLimit(gs, b)
}

func hasTradeRouteRecordBetween(gs *state.GameState, a, b faction.FactionID) bool {
	if gs == nil || a == "" || b == "" || a == b {
		return false
	}
	aStr := string(a)
	bStr := string(b)
	for _, route := range gs.TradeRoutes {
		if route == nil {
			continue
		}
		if (route.FromFactionID == aStr && route.ToFactionID == bStr) ||
			(route.FromFactionID == bStr && route.ToFactionID == aStr) {
			return true
		}
	}
	return false
}

func tradeAgreementKey(a, b faction.FactionID) (string, faction.FactionID, faction.FactionID) {
	if a < b {
		return string(a) + "|" + string(b), a, b
	}
	return string(b) + "|" + string(a), b, a
}

// RebalanceTradeRouteCapacities dış ticaret anlaşmalarına bir devletin efektif
// kapasitesini dengeli olarak paylaştırır. Her partner önce eşit kapasite
// payını alır, kalan birimler ID sırasındaki ilk partnerlere dağıtılır ve tek
// anlaşma hacmi tarafların dinamik tavanıyla sınırlanır. İki yönlü rota aynı
// temel hacmi paylaşır.
func RebalanceTradeRouteCapacities(gs *state.GameState) {
	if gs == nil || len(gs.TradeRoutes) == 0 {
		return
	}
	type agreement struct {
		key                            string
		left, right                    faction.FactionID
		routes                         []*economy.TradeRoute
		hasLeftToRight, hasRightToLeft bool
		active                         bool
	}
	agreementsByKey := make(map[string]*agreement)
	for _, route := range gs.TradeRoutes {
		if route == nil || route.FromFactionID == "" || route.ToFactionID == "" || route.FromFactionID == route.ToFactionID {
			continue
		}
		fromID := faction.FactionID(route.FromFactionID)
		toID := faction.FactionID(route.ToFactionID)
		key, left, right := tradeAgreementKey(fromID, toID)
		item := agreementsByKey[key]
		if item == nil {
			item = &agreement{key: key, left: left, right: right}
			agreementsByKey[key] = item
		}
		item.routes = append(item.routes, route)
		if fromID == left && toID == right {
			item.hasLeftToRight = true
		} else {
			item.hasRightToLeft = true
		}
		if route.SuspendedTurns <= 0 {
			item.active = true
		}
	}

	partnersByFaction := make(map[faction.FactionID][]string)
	for key, item := range agreementsByKey {
		if !item.active || !item.hasLeftToRight || !item.hasRightToLeft || SameRealm(gs, item.left, item.right) {
			continue
		}
		partnersByFaction[item.left] = append(partnersByFaction[item.left], key)
		partnersByFaction[item.right] = append(partnersByFaction[item.right], key)
	}
	sharesByFaction := make(map[faction.FactionID]map[string]int, len(partnersByFaction))
	for fid, keys := range partnersByFaction {
		sort.Strings(keys)
		capacity := totalTradeCapacity(gs, fid)
		baseShare := 0
		remainder := 0
		if len(keys) > 0 && capacity > 0 {
			baseShare = capacity / len(keys)
			remainder = capacity % len(keys)
		}
		shares := make(map[string]int, len(keys))
		for i, key := range keys {
			share := baseShare
			if i < remainder {
				share++
			}
			amountLimit := TradeRouteAmountLimit(gs, fid)
			if share > amountLimit {
				share = amountLimit
			}
			shares[key] = share
		}
		sharesByFaction[fid] = shares
	}

	for _, item := range agreementsByKey {
		if !item.active || !item.hasLeftToRight || !item.hasRightToLeft || SameRealm(gs, item.left, item.right) {
			continue
		}
		amount := sharesByFaction[item.left][item.key]
		if other := sharesByFaction[item.right][item.key]; other < amount {
			amount = other
		}
		if agreementLimit := TradeAgreementAmountLimit(gs, item.left, item.right); amount > agreementLimit {
			amount = agreementLimit
		}
		for _, route := range item.routes {
			route.AmountPerTurn = amount
		}
	}
}

// TradeRouteCapacityUsage bir devletin dış ticaret anlaşmalarına ayırdığı
// temel rota kapasitesini döner. Merchant bonusları kapasite rezervine girmez.
func TradeRouteCapacityUsage(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" || len(gs.TradeRoutes) == 0 {
		return 0
	}
	amountByAgreement := make(map[string]int)
	for _, route := range gs.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 {
			continue
		}
		fromID := faction.FactionID(route.FromFactionID)
		toID := faction.FactionID(route.ToFactionID)
		if (fromID != fid && toID != fid) || SameRealm(gs, fromID, toID) {
			continue
		}
		key, _, _ := tradeAgreementKey(fromID, toID)
		if route.AmountPerTurn > amountByAgreement[key] {
			amountByAgreement[key] = route.AmountPerTurn
		}
	}
	total := 0
	for _, amount := range amountByAgreement {
		total += amount
	}
	return total
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
