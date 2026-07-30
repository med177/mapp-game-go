package diplomacy

import (
	"strconv"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

const (
	relationImprovementCost     = 40
	relationImprovementBonus    = 8
	giftCost                    = 120
	giftReceiverGold            = 80
	giftRelationBonus           = 15
	vassalizationMinScore       = 55
	vassalAcceptanceThreshold   = 65
	vassalTributeRatePercent    = 20
	vassalInternalRelationFloor = 40
	// VassalAnnexationMinimumTurns, yeni vassal yapılan devletin ilhak
	// edilebilmesi için realm içinde kalması gereken asgari tur sayısıdır.
	VassalAnnexationMinimumTurns = 12
)

type VassalProposalAssessment struct {
	Chance      int
	BlockReason string
}

func (a VassalProposalAssessment) Accepted() bool {
	return a.BlockReason == "" && a.Chance >= vassalAcceptanceThreshold
}

func DirectOverlord(gs *state.GameState, fid faction.FactionID) faction.FactionID {
	if gs == nil || fid == "" {
		return ""
	}
	f := gs.Factions[fid]
	if f == nil || f.IsEliminated || f.OverlordID == "" || f.OverlordID == fid {
		return ""
	}
	overlord := gs.Factions[f.OverlordID]
	if overlord == nil || overlord.IsEliminated {
		return ""
	}
	return f.OverlordID
}

func RootOverlord(gs *state.GameState, fid faction.FactionID) faction.FactionID {
	if gs == nil || fid == "" {
		return ""
	}
	current := fid
	seen := map[faction.FactionID]struct{}{}
	for {
		overlord := DirectOverlord(gs, current)
		if overlord == "" {
			if current == fid {
				return ""
			}
			return current
		}
		if _, exists := seen[overlord]; exists {
			return ""
		}
		seen[overlord] = struct{}{}
		current = overlord
	}
}

func realmRoot(gs *state.GameState, fid faction.FactionID) faction.FactionID {
	if fid == "" {
		return ""
	}
	if root := RootOverlord(gs, fid); root != "" {
		return root
	}
	return fid
}

func sameRealm(gs *state.GameState, a, b faction.FactionID) bool {
	return a != "" && b != "" && realmRoot(gs, a) == realmRoot(gs, b)
}

func RealmRoot(gs *state.GameState, fid faction.FactionID) faction.FactionID {
	return realmRoot(gs, fid)
}

func SameRealm(gs *state.GameState, a, b faction.FactionID) bool {
	return sameRealm(gs, a, b)
}

func VassalsOf(gs *state.GameState, overlord faction.FactionID) []faction.FactionID {
	if gs == nil || overlord == "" {
		return nil
	}
	root := realmRoot(gs, overlord)
	if root == "" {
		root = overlord
	}
	out := make([]faction.FactionID, 0, len(gs.Factions))
	for fid, f := range gs.Factions {
		if f == nil || f.IsEliminated || fid == root {
			continue
		}
		if realmRoot(gs, fid) == root {
			out = append(out, fid)
		}
	}
	return out
}

func warCoalition(gs *state.GameState, fid faction.FactionID) []faction.FactionID {
	if gs == nil || fid == "" {
		return nil
	}
	root := realmRoot(gs, fid)
	if root == "" {
		root = fid
	}
	members := []faction.FactionID{root}
	for _, vassalID := range VassalsOf(gs, root) {
		members = append(members, vassalID)
	}
	return members
}

func relationSnapshot(gs *state.GameState, a, b faction.FactionID) (int, faction.DiplomaticStance) {
	rel := Relation(gs, a, b)
	if rel == nil {
		return 0, faction.StancePeace
	}
	return rel.Score, rel.Stance
}

func AssessVassalizationProposal(gs *state.GameState, rel *faction.Relation, actor, target faction.FactionID) VassalProposalAssessment {
	assessment := VassalProposalAssessment{}
	if gs == nil || actor == "" || target == "" || actor == target {
		assessment.BlockReason = "Geçersiz diplomasi hedefi"
		return assessment
	}
	if DirectOverlord(gs, actor) != "" {
		assessment.BlockReason = "Bağlı devletler yeni vassal edinemaz"
		return assessment
	}
	if overlord := DirectOverlord(gs, target); overlord != "" {
		if overlord == actor {
			assessment.BlockReason = "Bu devlet zaten sana bağlı"
		} else {
			assessment.BlockReason = "Hedef başka bir devlete bağlı"
		}
		return assessment
	}
	score := 0
	if rel != nil {
		score = rel.Score
	}
	if score < vassalizationMinScore {
		assessment.BlockReason = "İlişki puanı 55 altı"
		return assessment
	}
	if hasExternalWar(gs, target, actor) {
		assessment.BlockReason = "Hedef başka bir savaşın içinde"
		return assessment
	}

	actorPower := MilitaryPower(gs, actor)
	targetPower := MilitaryPower(gs, target)
	actorRegions := len(gs.LandRegionsOwnedBy(actor))
	targetRegions := len(gs.LandRegionsOwnedBy(target))
	if actorPower < targetPower*3/2 && actorRegions <= targetRegions+1 {
		assessment.BlockReason = "Askeri ve bölgesel üstünlük yetersiz"
		return assessment
	}

	chance := 15 + score
	chance += clamp((actorPower-targetPower)/5, -10, 35)
	chance += clamp((actorRegions-targetRegions)*4, -8, 24)
	chance += economicStress(gs, target)
	if HasCommonEnemy(gs, actor, target) {
		chance += 6
	}
	if HasDirectThreat(gs, actor, target) {
		chance += 6
	}
	assessment.Chance = clamp(chance, 0, 100)
	return assessment
}

func ActionBlockReason(gs *state.GameState, actor, target faction.FactionID, action Action) string {
	return actionBlockReason(gs, actor, target, action, true)
}

func actionBlockReason(gs *state.GameState, actor, target faction.FactionID, action Action, checkQuota bool) string {
	if gs == nil {
		return "Diplomasi durumu yok."
	}
	if actor == "" || target == "" || actor == target {
		return "Geçersiz diplomasi hedefi."
	}
	actorFaction := gs.Factions[actor]
	targetFaction := gs.Factions[target]
	if actorFaction == nil || targetFaction == nil {
		return "Fraksiyon bulunamadı."
	}
	if actorFaction.IsEliminated || targetFaction.IsEliminated {
		return "Elenmiş fraksiyonlarla diplomasi kurulamaz."
	}
	if checkQuota && actionUsesDiplomacyOfferQuota(action) {
		if reason := diplomacyOfferQuotaBlockReason(gs, actor); reason != "" {
			return reason
		}
	}

	actorOverlord := DirectOverlord(gs, actor)
	targetOverlord := DirectOverlord(gs, target)
	switch {
	case actorOverlord != "" && target != actorOverlord:
		return factionLabel(gs, actor) + " yalnız " + factionLabel(gs, actorOverlord) + " adına diplomasi yürütebilir."
	case targetOverlord != "" && actor != targetOverlord:
		return factionLabel(gs, target) + " doğrudan diplomasi yürütemez; " + factionLabel(gs, targetOverlord) + " ile görüş."
	case actorOverlord == target || targetOverlord == actor:
		switch action {
		case ActionImproveRelations, ActionSendGift:
		case ActionReleaseVassal, ActionAnnexVassal:
			if targetOverlord != actor {
				return "Yalnız doğrudan bağlı devlet yönetilebilir."
			}
		case ActionOfferVassalization:
			if targetOverlord == actor {
				return "Bu devlet zaten sana bağlı."
			}
			return "Bağlı olduğun devlete vassallık teklif edemezsin."
		default:
			return "Bağlı devlet ilişkilerinde yalnız sadakat artırıcı adımlar kullanılabilir."
		}
	}

	score, stance := relationSnapshot(gs, actor, target)
	switch action {
	case ActionDeclareWar:
		if sameRealm(gs, actor, target) {
			return "Aynı vassal zincirindeki devletlere savaş açılamaz."
		}
		if stance == faction.StanceWar {
			return factionLabel(gs, target) + " ile zaten savaş halindesiniz."
		}
		if remaining := gs.TruceRemaining(actor, target); remaining > 0 {
			return factionLabel(gs, target) + " ile ateşkes sürüyor (" + itoa(remaining) + " tur)."
		}
	case ActionProposePeace:
		if stance != faction.StanceWar {
			return "Barış teklifi sadece savaşta yapılır."
		}
	case ActionProposeAlliance:
		if stance == faction.StanceWar {
			return "Savaş halindeyken ittifak kurulamaz."
		}
		if stance == faction.StanceAllied {
			return "Zaten müttefiksiniz."
		}
		if score < allianceRelationThresholdFor(gs) {
			return allianceRelationBlockReason(gs) + "."
		}
		if assessment := AssessAllianceProposal(gs, Relation(gs, actor, target), actor, target); assessment.BlockReason != "" {
			return assessment.BlockReason
		}
	case ActionCancelAlliance:
		if sameRealm(gs, actor, target) {
			return "Vassal bağı ittifak iptaliyle sona erdirilemez."
		}
		if stance != faction.StanceAllied {
			return "Bu devletle aktif bir ittifak yok."
		}
	case ActionProposeTrade:
		if stance == faction.StanceWar {
			return "Savaş halindeyken ticaret yapılamaz."
		}
		if stance == faction.StanceTrade && HasTradeRouteBetween(gs, actor, target) {
			return "Zaten aktif bir ticaret anlaşması var."
		}
		if stance == faction.StanceAllied && HasTradeRouteBetween(gs, actor, target) {
			return "Bu müttefik ile ticaret zaten aktif."
		}
		if assessment := AssessTradeProposal(gs, Relation(gs, actor, target), actor, target); assessment.BlockReason != "" {
			return assessment.BlockReason
		}
	case ActionCancelTrade:
		if sameRealm(gs, actor, target) {
			return "Vassal ticareti vassallık sürdüğü sürece iptal edilemez."
		}
		if !HasTradeRouteBetween(gs, actor, target) {
			return "Bu devletle aktif bir ticaret anlaşması yok."
		}
	case ActionImproveRelations:
		if stance == faction.StanceWar {
			return "Savaş halindeyken heyet gönderilemez."
		}
		if actorFaction.Gold < relationImprovementCost {
			return "Heyet göndermek için 40 altın gerekiyor."
		}
		if score >= 95 {
			return "İlişki zaten çok yüksek."
		}
	case ActionSendGift:
		if stance == faction.StanceWar {
			return "Savaş halindeyken hediye gönderilemez."
		}
		if actorFaction.Gold < giftCost {
			return "Hediye göndermek için 120 altın gerekiyor."
		}
		if score >= 98 {
			return "İlişki zaten çok yüksek."
		}
	case ActionOfferVassalization:
		if sameRealm(gs, actor, target) {
			return "Aynı vassal zincirindeki devletlere vassallık teklif edilemez."
		}
		if stance == faction.StanceWar {
			return "Savaş halindeyken vassallık teklif edilemez."
		}
		if assessment := AssessVassalizationProposal(gs, Relation(gs, actor, target), actor, target); assessment.BlockReason != "" {
			return assessment.BlockReason
		}
	case ActionReleaseVassal, ActionAnnexVassal:
		if DirectOverlord(gs, target) != actor {
			return "Hedef doğrudan sana bağlı bir devlet değil."
		}
		if action == ActionAnnexVassal && targetFaction.VassalizedTurn > 0 {
			elapsed := gs.Turn - targetFaction.VassalizedTurn
			if elapsed < VassalAnnexationMinimumTurns {
				remaining := VassalAnnexationMinimumTurns - elapsed
				return factionLabel(gs, target) + " en az 12 tur vassal kaldıktan sonra ilhak edilebilir (" + itoa(remaining) + " tur kaldı)."
			}
		}
	}
	return ""
}

func hasExternalWar(gs *state.GameState, fid, ignored faction.FactionID) bool {
	if gs == nil || fid == "" {
		return false
	}
	for _, rel := range gs.Relations {
		if rel == nil || rel.Stance != faction.StanceWar {
			continue
		}
		var other faction.FactionID
		switch fid {
		case rel.FactionA:
			other = rel.FactionB
		case rel.FactionB:
			other = rel.FactionA
		default:
			continue
		}
		if other == ignored || sameRealm(gs, fid, other) {
			continue
		}
		return true
	}
	return false
}

func normalizeVassalRealmRelations(gs *state.GameState, root faction.FactionID) {
	if gs == nil || root == "" {
		return
	}
	members := warCoalition(gs, root)
	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			rel := EnsureRelation(gs, members[i], members[j])
			rel.Stance = faction.StanceAllied
			if rel.Score < vassalInternalRelationFloor {
				rel.Score = vassalInternalRelationFloor
			}
		}
	}
	gs.SyncWarLedgers()
}

func setWarBetweenCoalitions(gs *state.GameState, a, b faction.FactionID) {
	if gs == nil || a == "" || b == "" || sameRealm(gs, a, b) {
		return
	}
	left := warCoalition(gs, a)
	right := warCoalition(gs, b)
	for _, lhs := range left {
		for _, rhs := range right {
			if lhs == rhs {
				continue
			}
			rel := EnsureRelation(gs, lhs, rhs)
			wasWar := rel.Stance == faction.StanceWar
			rel.Stance = faction.StanceWar
			rel.Score = -80
			removeTradeRoutesBetween(gs, lhs, rhs)
			if !wasWar {
				gs.BeginWarLedger(lhs, rhs)
			}
		}
	}
	gs.SyncWarLedgers()
}

func setPeaceBetweenCoalitions(gs *state.GameState, a, b faction.FactionID) {
	if gs == nil || a == "" || b == "" {
		return
	}
	left := warCoalition(gs, a)
	right := warCoalition(gs, b)
	for _, lhs := range left {
		for _, rhs := range right {
			if lhs == rhs {
				continue
			}
			rel := EnsureRelation(gs, lhs, rhs)
			wasWar := rel.Stance == faction.StanceWar
			rel.Stance = faction.StancePeace
			rel.Score = -20
			removeTradeRoutesBetween(gs, lhs, rhs)
			if wasWar {
				gs.EndWarLedger(lhs, rhs)
				gs.RecordTruce(lhs, rhs)
			}
		}
	}
	removePendingSurrenderOffersBetween(gs, left, right)
}

// removePendingSurrenderOffersBetween, barış kabul edildiğinde aynı savaşın
// artık geçersiz olan kuşatma teslimiyeti tekliflerini kuyruktan düşürür.
func removePendingSurrenderOffersBetween(gs *state.GameState, left, right []faction.FactionID) {
	if gs == nil || len(gs.DiplomaticOffers) == 0 {
		return
	}
	leftSet := make(map[faction.FactionID]struct{}, len(left))
	rightSet := make(map[faction.FactionID]struct{}, len(right))
	for _, fid := range left {
		leftSet[fid] = struct{}{}
	}
	for _, fid := range right {
		rightSet[fid] = struct{}{}
	}
	offers := gs.DiplomaticOffers[:0]
	for _, offer := range gs.DiplomaticOffers {
		if offer.Action == string(ActionProposeSurrender) &&
			((containsFaction(leftSet, offer.FromFactionID) && containsFaction(rightSet, offer.ToFactionID)) ||
				(containsFaction(rightSet, offer.FromFactionID) && containsFaction(leftSet, offer.ToFactionID))) {
			continue
		}
		offers = append(offers, offer)
	}
	gs.DiplomaticOffers = offers
}

func containsFaction(set map[faction.FactionID]struct{}, fid faction.FactionID) bool {
	_, ok := set[fid]
	return ok
}

func synchronizeVassalWars(gs *state.GameState, root faction.FactionID) {
	if gs == nil || root == "" {
		return
	}
	for otherID, other := range gs.Factions {
		if other == nil || other.IsEliminated || sameRealm(gs, root, otherID) {
			continue
		}
		otherRoot := realmRoot(gs, otherID)
		if otherRoot == "" {
			otherRoot = otherID
		}
		if rel := Relation(gs, root, otherRoot); rel != nil && rel.Stance == faction.StanceWar {
			setWarBetweenCoalitions(gs, root, otherID)
		}
	}
}

func applyRelationImprovement(gs *state.GameState, actor, target faction.FactionID, cost, delta int, receiverGold int, sourceLabel string) Result {
	actorFaction := gs.Factions[actor]
	targetFaction := gs.Factions[target]
	if actorFaction == nil || targetFaction == nil {
		return Result{Message: "Fraksiyon bulunamadı."}
	}
	actorFaction.Gold -= cost
	if actorFaction.Gold < 0 {
		actorFaction.Gold = 0
	}
	if receiverGold > 0 {
		targetFaction.Gold += receiverGold
	}
	rel := EnsureRelation(gs, actor, target)
	rel.Score = clamp(rel.Score+delta, -100, 100)
	return Result{
		Accepted: true,
		Applied:  true,
		Message:  factionLabel(gs, target) + " için " + sourceLabel + " gönderildi. İlişki +" + strconv.Itoa(delta) + ".",
	}
}

func applyVassalization(gs *state.GameState, actor, target faction.FactionID) Result {
	targetFaction := gs.Factions[target]
	if targetFaction == nil {
		return Result{Message: "Fraksiyon bulunamadı."}
	}
	targetFaction.OverlordID = actor
	targetFaction.VassalizedTurn = gs.Turn
	if targetFaction.VassalizedTurn <= 0 {
		targetFaction.VassalizedTurn = 1
	}
	root := realmRoot(gs, actor)
	if root == "" {
		root = actor
	}
	normalizeVassalRealmRelations(gs, root)
	sanitizeVassalExternalDiplomacy(gs, target, root)
	ensureTradeRoutesBetween(gs, actor, target)
	synchronizeVassalWars(gs, root)
	return Result{
		Accepted: true,
		Applied:  true,
		Message:  factionLabel(gs, target) + " artık " + factionLabel(gs, actor) + " vassalı oldu.",
	}
}

func releaseVassalage(gs *state.GameState, actor, target faction.FactionID) Result {
	targetFaction := gs.Factions[target]
	if targetFaction == nil {
		return Result{Message: "Fraksiyon bulunamadı."}
	}
	targetFaction.OverlordID = ""
	targetFaction.VassalizedTurn = 0
	rel := EnsureRelation(gs, actor, target)
	rel.Stance = faction.StanceTrade
	if rel.Score < 25 {
		rel.Score = 25
	}
	ensureTradeRoutesBetween(gs, actor, target)
	normalizeVassalRealmRelations(gs, actor)
	normalizeVassalRealmRelations(gs, target)
	return Result{
		Accepted: true,
		Applied:  true,
		Message:  factionLabel(gs, target) + " artık bağımsız bir devlet; ticaret anlaşması devam ediyor.",
	}
}

func ForceVassalizeAfterWar(gs *state.GameState, actor, target faction.FactionID) Result {
	if gs == nil || actor == "" || target == "" || actor == target {
		return Result{Message: "Geçersiz diplomasi hedefi."}
	}
	actorFaction := gs.Factions[actor]
	targetFaction := gs.Factions[target]
	if actorFaction == nil || targetFaction == nil {
		return Result{Message: "Fraksiyon bulunamadı."}
	}
	if actorFaction.IsEliminated || targetFaction.IsEliminated {
		return Result{Message: "Elenmiş fraksiyonlarla bu karar uygulanamaz."}
	}
	if DirectOverlord(gs, actor) != "" {
		return Result{Message: "Bağlı devletler savaş sonrası vassal edinemaz."}
	}
	if overlord := DirectOverlord(gs, target); overlord != "" && overlord != actor {
		return Result{Message: "Hedef başka bir devlete bağlı."}
	}
	setPeaceBetweenCoalitions(gs, actor, target)
	return applyVassalization(gs, actor, target)
}

// ForceReleaseAfterWar, savaş sonrası ardıl devlet kararında hedefi
// bağımsız bırakır. Normal diplomasi ekranındaki Vasallığı Bitir aksiyonundan
// farklı olarak burada ilişki savaşın hemen ardından müttefikliğe çekilir.
func ForceReleaseAfterWar(gs *state.GameState, actor, target faction.FactionID) Result {
	if gs == nil || actor == "" || target == "" || actor == target {
		return Result{Message: "Geçersiz diplomasi hedefi."}
	}
	actorFaction := gs.Factions[actor]
	targetFaction := gs.Factions[target]
	if actorFaction == nil || targetFaction == nil || actorFaction.IsEliminated || targetFaction.IsEliminated {
		return Result{Message: "Elenmiş fraksiyonlarla bu karar uygulanamaz."}
	}
	setPeaceBetweenCoalitions(gs, actor, target)
	targetFaction.OverlordID = ""
	targetFaction.VassalizedTurn = 0
	ForceRelation(gs, actor, target, faction.StanceAllied, 50)
	normalizeVassalRealmRelations(gs, actor)
	normalizeVassalRealmRelations(gs, target)
	return Result{
		Accepted: true,
		Applied:  true,
		Message:  factionLabel(gs, target) + " serbest bırakıldı ve bağımsız müttefik oldu.",
	}
}

func sanitizeVassalExternalDiplomacy(gs *state.GameState, vassal, root faction.FactionID) {
	if gs == nil || vassal == "" {
		return
	}
	for _, rel := range gs.Relations {
		if rel == nil {
			continue
		}
		var other faction.FactionID
		switch vassal {
		case rel.FactionA:
			other = rel.FactionB
		case rel.FactionB:
			other = rel.FactionA
		default:
			continue
		}
		if sameRealm(gs, vassal, other) {
			continue
		}
		rootRel := Relation(gs, root, realmRoot(gs, other))
		if rootRel != nil && rootRel.Stance == faction.StanceWar {
			rel.Stance = faction.StanceWar
			rel.Score = -80
			continue
		}
		rel.Stance = faction.StancePeace
		if rel.Score < -20 {
			rel.Score = -20
		}
		removeTradeRoutesBetween(gs, vassal, other)
	}
	if len(gs.DiplomaticOffers) > 0 {
		offers := gs.DiplomaticOffers[:0]
		for _, offer := range gs.DiplomaticOffers {
			keep := true
			if offer.FromFactionID == vassal && !sameRealm(gs, offer.ToFactionID, root) {
				keep = false
			}
			if offer.ToFactionID == vassal && !sameRealm(gs, offer.FromFactionID, root) {
				keep = false
			}
			if keep {
				offers = append(offers, offer)
			}
		}
		gs.DiplomaticOffers = offers
	}
}

func NormalizeVassalage(gs *state.GameState) {
	if gs == nil || len(gs.Factions) == 0 {
		return
	}
	for fid, f := range gs.Factions {
		if f == nil {
			continue
		}
		if f.OverlordID == "" {
			f.VassalizedTurn = 0
			continue
		}
		if f.OverlordID == fid {
			f.OverlordID = ""
			f.VassalizedTurn = 0
			continue
		}
		overlord := gs.Factions[f.OverlordID]
		if overlord == nil || overlord.IsEliminated {
			f.OverlordID = ""
			f.VassalizedTurn = 0
			continue
		}
		if f.VassalizedTurn <= 0 {
			f.VassalizedTurn = gs.Turn
			if f.VassalizedTurn <= 0 {
				f.VassalizedTurn = 1
			}
		}
	}

	for fid, f := range gs.Factions {
		if f == nil || f.IsEliminated {
			continue
		}
		if root := realmRoot(gs, fid); root != fid {
			normalizeVassalRealmRelations(gs, root)
			sanitizeVassalExternalDiplomacy(gs, fid, root)
			ensureTradeRoutesBetween(gs, f.OverlordID, fid)
		}
	}
	gs.SyncWarLedgers()
}

func VassalTributeRatePercent() int {
	return vassalTributeRatePercent
}
