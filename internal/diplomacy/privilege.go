package diplomacy

import (
	"fmt"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// PrivilegeRevocationRelationPenalty, egemen devletin imtiyazı tek taraflı
// kaldırmasının kullanım sahibiyle ilişkiye verdiği puan cezasıdır.
const PrivilegeRevocationRelationPenalty = 10

// PrivilegeOfferRelationBonus, kabul edilen ilk imtiyazın iki devlet arasında
// oluşturduğu karşılıklı ilişki artışıdır.
const PrivilegeOfferRelationBonus = 15

// privilegeEconomicBenefitRelationWeight, altın/tur cinsinden ekonomik
// faydanın ilişki puanı yüküyle karşılaştırıldığı denge katsayısıdır. İlişki
// skoru bir altın gibi bire bir sayılmaz; aksi halde düşük gelirli minorlar
// -100 ilişki durumunda gereksiz yere hiç teklif alamaz.
const privilegeEconomicBenefitRelationWeight = 20

// MinorPrivilegeOfferAssessment, imtiyaz teklifinin hedef devlet açısından
// ekonomik ve diplomatik kabul edilebilirliğini taşır.
type MinorPrivilegeOfferAssessment struct {
	Chance            int
	EconomicBenefit   int
	EconomicAdvantage bool
	BlockReason       string
}

func (a MinorPrivilegeOfferAssessment) Accepted() bool {
	return a.BlockReason == "" && a.Chance >= 45
}

// MinorPrivilegeOfferBlockReason, teklif düğmesi ve teklif kuyruğu için ortak
// ilk imtiyaz uygunluk kontrolüdür.
func MinorPrivilegeOfferBlockReason(gs *state.GameState, actor, target faction.FactionID, rid world.RegionID) string {
	if gs == nil || actor == "" || target == "" || actor == target || rid == "" {
		return "Geçersiz imtiyaz teklifi."
	}
	from := gs.Factions[actor]
	to := gs.Factions[target]
	if from == nil || to == nil || from.IsEliminated || to.IsEliminated {
		return "Elenmiş devletlere imtiyaz teklif edilemez."
	}
	if from.IsVirtual || to.IsVirtual {
		return "Sanal devletlere imtiyaz teklif edilemez."
	}
	region := gs.Regions[rid]
	if region == nil || region.IsSea || !region.IsMinorRegion {
		return "İmtiyaz yalnız minor bölgelere verilebilir."
	}
	if region.IsPrivileged {
		return "Bu bölgenin zaten aktif bir imtiyazı var."
	}
	if gs.SovereignOwnerID(region) != string(actor) || region.OwnerID != string(actor) {
		return "İmtiyazı yalnız bölgenin doğrudan egemen sahibi verebilir."
	}
	if IsWar(gs, actor, target) {
		return "Savaş halindeki devlete imtiyaz teklif edilemez."
	}
	if sameRealm(gs, actor, target) {
		return "Aynı realm içindeki devlete imtiyaz teklif edilemez."
	}
	if !gs.CanSpendDiplomacyOfferQuota(actor) {
		return "Bu tur diplomasi elçisi hakkın doldu."
	}
	for _, offer := range gs.DiplomaticOffers {
		if offer.Action == string(ActionOfferMinorPrivilege) && offer.FromFactionID == actor && offer.ToFactionID == target && offer.RegionID == rid {
			return "Bu imtiyaz teklifi zaten bekliyor."
		}
	}
	if gs.DiplomaticOfferRegionRetryBlocked(string(actor), string(target), string(ActionOfferMinorPrivilege), rid, 1) {
		return "Bu devlet bu bölge imtiyazını yakın zamanda reddetti."
	}
	return ""
}

// AssessMinorPrivilegeOffer, teklif alacak devletin imtiyazdan elde edeceği
// yerel payı ve kurulacak rotanın tahmini brüt ticaret değerini ilişki yüküyle
// karşılaştırır. Ekonomik fayda ilişki yükünü ölçeklendirilmiş olarak aşarsa
// ilişki skoru kararı ezemez; AI bu durumda %90 kabul şansına zar atar.
func AssessMinorPrivilegeOffer(gs *state.GameState, actor, target faction.FactionID, rid world.RegionID) MinorPrivilegeOfferAssessment {
	if reason := MinorPrivilegeOfferBlockReason(gs, actor, target, rid); reason != "" {
		return MinorPrivilegeOfferAssessment{BlockReason: reason}
	}
	region := gs.Regions[rid]
	production := gs.RegionProductionSummary(region)
	localShare := production.Gold * state.PrivilegeIncomeSharePercent / 100
	tradeRoute := buildTradeRoute(gs, actor, target)
	tradeRoute.IsPrivilegedMinor = true
	tradeValue := gs.PrivilegedTradeRouteValue(tradeRoute, tradeRoute.AmountPerTurn)
	economicBenefit := localShare + tradeValue
	relationScore := RelationScore(gs, actor, target)
	relationBurden := maxInt(0, -relationScore)
	minimumEconomicBenefit := (relationBurden + privilegeEconomicBenefitRelationWeight - 1) / privilegeEconomicBenefitRelationWeight
	if minimumEconomicBenefit < 1 {
		minimumEconomicBenefit = 1
	}
	economicAdvantage := economicBenefit >= minimumEconomicBenefit
	chance := 45 + relationScore/4
	if economicAdvantage {
		chance = 90
	}
	if chance < 0 {
		chance = 0
	}
	if chance > 95 {
		chance = 95
	}
	return MinorPrivilegeOfferAssessment{
		Chance:            chance,
		EconomicBenefit:   economicBenefit,
		EconomicAdvantage: economicAdvantage,
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RevokeMinorPrivilege, egemen devletin kendi imtiyazlı minor bölgesindeki
// kullanım hakkını kaldırır. Kullanım sahibiyle savaş başlatmaz; yalnızca
// ilişki puanını düşürür.
func RevokeMinorPrivilege(gs *state.GameState, actor faction.FactionID, rid world.RegionID) Result {
	if gs == nil || actor == "" || rid == "" {
		return Result{Message: "İmtiyaz kaldırma isteği geçersiz."}
	}
	region := gs.Regions[rid]
	if region == nil || region.IsSea || !region.IsMinorRegion || !region.IsPrivileged {
		return Result{Message: "Bu bölgenin aktif imtiyazı yok."}
	}
	if gs.SovereignOwnerID(region) != string(actor) {
		return Result{Message: "İmtiyazı yalnızca bölgenin egemen sahibi kaldırabilir."}
	}
	operator := faction.FactionID(region.OwnerID)
	if operator == "" || operator == actor {
		return Result{Message: "Bu bölgenin ayrı bir imtiyaz sahibi yok."}
	}

	region.IsPrivileged = false
	EnsurePrivilegedMinorTradeRoutes(gs)
	ForceRelation(gs, actor, operator, "", -PrivilegeRevocationRelationPenalty)
	return Result{
		Accepted: true,
		Applied:  true,
		Message:  fmt.Sprintf("%s bölgesindeki imtiyaz kaldırıldı. %s ile ilişki -%d.", region.NameTR, factionLabel(gs, operator), PrivilegeRevocationRelationPenalty),
	}
}

// OfferMinorPrivilege, kabul edilmiş ilk imtiyaz kararını uygular. Teklifin
// kuyruğa alınması offers.go'daki kota ve tekrar kontrollerinden geçer.
func OfferMinorPrivilege(gs *state.GameState, actor, target faction.FactionID, rid world.RegionID) Result {
	if reason := MinorPrivilegeOfferBlockReasonWithoutQueueChecks(gs, actor, target, rid); reason != "" {
		return Result{Message: reason}
	}
	region := gs.Regions[rid]
	region.IsPrivileged = true
	region.OwnerID = string(target)
	region.PrivilegeGrantedTurn = gs.Turn
	EnsurePrivilegedMinorTradeRoutes(gs)
	AddRelationScoreBoth(gs, actor, target, PrivilegeOfferRelationBonus)
	return Result{
		Accepted: true,
		Applied:  true,
		Message:  fmt.Sprintf("%s bölgesinde %s devletine imtiyaz verildi; rota kuruldu ve ilişki +%d arttı.", region.NameTR, factionLabel(gs, target), PrivilegeOfferRelationBonus),
	}
}

func MinorPrivilegeOfferBlockReasonWithoutQueueChecks(gs *state.GameState, actor, target faction.FactionID, rid world.RegionID) string {
	if gs == nil || actor == "" || target == "" || actor == target || rid == "" {
		return "Geçersiz imtiyaz teklifi."
	}
	from := gs.Factions[actor]
	to := gs.Factions[target]
	if from == nil || to == nil || from.IsEliminated || to.IsEliminated || from.IsVirtual || to.IsVirtual {
		return "İmtiyaz teklifi taraflarından biri artık geçerli değil."
	}
	region := gs.Regions[rid]
	if region == nil || region.IsSea || !region.IsMinorRegion || region.IsPrivileged {
		return "Bu minor bölge için imtiyaz artık verilemez."
	}
	if gs.SovereignOwnerID(region) != string(actor) || region.OwnerID != string(actor) || IsWar(gs, actor, target) || sameRealm(gs, actor, target) {
		return "İmtiyaz teklifi artık geçerli değil."
	}
	return ""
}
