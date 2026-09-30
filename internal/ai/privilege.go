package ai

import (
	"strconv"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	aiPrivilegeKeepRelationScore = 40
	aiPrivilegeMinimumGoldGain   = 5
)

// ShouldAcceptMinorPrivilegeOffer, AI hedefinin ekonomik avantajı ilişki
// yükünü aştığında kampanya karar zarını kullanır. Zar aynı teklif ve tur için
// seed tabanlıdır; böylece kayıt yükleme ve headless simülasyon deterministik
// kalırken kabul olasılığı gerçek anlamda %60 olur.
func ShouldAcceptMinorPrivilegeOffer(gs *state.GameState, from, to faction.FactionID, rid world.RegionID, assessment diplomacy.MinorPrivilegeOfferAssessment) bool {
	if gs == nil || assessment.BlockReason != "" {
		return false
	}
	roll := aiDecisionRoll(gs, from, to, string(diplomacy.ActionOfferMinorPrivilege)+"|"+string(rid))
	return roll < assessment.Chance
}

// aiFactionGoldFromProduction, imtiyazlı minor bölgelerde AI'nin gerçekten
// alacağı altın payını hesaplar. Kaynak üretimi OwnerID'de kalırken parasal
// gelir sovereign/operator arasında state katmanında paylaşılır.
func aiFactionGoldFromProduction(gs *state.GameState, fid faction.FactionID, region *world.Region, production state.RegionProductionSummary) int {
	if gs == nil || fid == "" || region == nil {
		return 0
	}
	sovereignGold, operatorGold := gs.RegionIncomeShares(region, production.Gold)
	total := 0
	if gs.SovereignOwnerID(region) == string(fid) {
		total += sovereignGold
	}
	if region.OwnerID == string(fid) {
		total += operatorGold
	}
	return total
}

func aiFactionGoldProduction(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" {
		return 0
	}
	total := 0
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || gs.SiegeAt(region.ID) != nil {
			continue
		}
		total += aiFactionGoldFromProduction(gs, fid, region, gs.RegionProductionSummary(region))
	}
	return total + gs.ExpectedPrivilegedTradeIncomeForFaction(fid)
}

func aiManageMinorPrivilegesWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) bool {
	if gs == nil || fid == "" {
		return false
	}
	changed := false
	for _, region := range aiSortedRegions(gs) {
		if !aiShouldRevokeMinorPrivilege(gs, fid, region) {
			continue
		}
		operator := faction.FactionID(region.OwnerID)
		if !aiApplyMinorPrivilegeRevocation(gs, fid, region) {
			continue
		}
		changed = true
		message := turnFactionName(gs, fid) + " " + turnRegionName(gs, region.ID) + " bölgesindeki imtiyazı kaldırdı; bölge doğrudan egemen yönetimine geçti."
		if operator != "" {
			message += " Kullanım sahibiyle ilişki -" + strconv.Itoa(diplomacy.PrivilegeRevocationRelationPenalty) + "."
			if operatorFaction := gs.Factions[operator]; operatorFaction != nil && operatorFaction.IsEliminated {
				message += " Kullanım sahibi devlet topraksız kaldığı için elendi."
			}
		}
		addTurnStep(steps, TurnStep{
			FactionID:     fid,
			Kind:          TurnStepDiplomacy,
			TargetRegion:  region.ID,
			FocusRegion:   region.ID,
			TargetFaction: operator,
			Message:       message,
		})
	}
	return changed
}

func aiShouldRevokeMinorPrivilege(gs *state.GameState, fid faction.FactionID, region *world.Region) bool {
	if gs == nil || region == nil || region.IsSea || !region.IsMinorRegion || !region.IsPrivileged || region.OwnerID == "" || region.OwnerID == string(fid) {
		return false
	}
	if gs.SovereignOwnerID(region) != string(fid) {
		return false
	}
	operator := faction.FactionID(region.OwnerID)
	operatorFaction := gs.Factions[operator]
	if operatorFaction == nil || operatorFaction.IsEliminated {
		return false
	}
	// Egemen devletle savaş, yeni imtiyaz korumasını bozan bilinçli askerî
	// istisnadır; savaş yoksa ilk 30 turda AI ekonomik gerekçeyle kaldıramaz.
	if diplomacy.IsWar(gs, fid, operator) {
		return true
	}
	if gs.MinorPrivilegeProtectionRemaining(region) > 0 {
		return false
	}
	relation := diplomacy.Relation(gs, fid, operator)
	if relation != nil && (relation.Stance == faction.StanceAllied || diplomacy.RelationScore(gs, fid, operator) >= aiPrivilegeKeepRelationScore) {
		return false
	}
	production := gs.RegionProductionSummary(region)
	_, operatorGold := gs.RegionIncomeShares(region, production.Gold)
	// İmtiyazlı rota telifi, kullanım sahibine bırakılacak yerel payı
	// karşılıyorsa egemen AI ekonomik avantajı korur. Aksi durumda eski
	// davranış gibi bölgenin tamamını geri almayı değerlendirebilir.
	if gs.ExpectedPrivilegeRoyaltyForRegion(region) >= operatorGold {
		return false
	}
	return operatorGold >= aiPrivilegeMinimumGoldGain
}

// aiApplyMinorPrivilegeRevocation, oyuncu akışındaki imtiyaz kaldırmanın AI
// tarafındaki state etkilerini uygular. Diplomasi işlemi ortak relation
// cezasını verir; bu yardımcı ise bölge devri ve askerî/üretim temizliğini
// tamamlar.
func aiApplyMinorPrivilegeRevocation(gs *state.GameState, actor faction.FactionID, region *world.Region) bool {
	if gs == nil || region == nil {
		return false
	}
	sovereign := faction.FactionID(gs.SovereignOwnerID(region))
	formerOwner := faction.FactionID(region.OwnerID)
	if sovereign == "" || formerOwner == "" || sovereign == formerOwner {
		return false
	}
	result := diplomacy.RevokeMinorPrivilege(gs, actor, region.ID)
	if !result.Applied {
		return false
	}

	formerOwner = faction.FactionID(gs.TransferRevokedMinorOwnership(region.ID, string(sovereign)))

	if formerOwner != sovereign && len(gs.LandRegionsOwnedBy(formerOwner)) == 0 {
		aiEliminateFactionAfterPrivilegeRevoke(gs, formerOwner)
	}
	gs.NormalizeFactionCapitals()
	gs.RefreshArmyMovePoints(false)
	return true
}
func aiEliminateFactionAfterPrivilegeRevoke(gs *state.GameState, fid faction.FactionID) {
	if gs == nil || fid == "" {
		return
	}
	f := gs.Factions[fid]
	if f == nil || f.IsEliminated {
		return
	}
	f.IsEliminated = true
	f.OverlordID = ""
	f.CapitalSettlementID = ""
	f.PendingCapitalSettlementID = ""
	f.PendingCapitalTurns = 0
	if gs.AIPlans != nil {
		delete(gs.AIPlans, fid)
	}
	for _, other := range gs.Factions {
		if other != nil && other.OverlordID == fid {
			other.OverlordID = ""
		}
	}
	for armyID, currentArmy := range gs.Armies {
		if currentArmy != nil && currentArmy.OwnerID == string(fid) {
			gs.RemoveArmy(armyID)
		}
	}
	for key, relation := range gs.Relations {
		if relation != nil && (relation.FactionA == fid || relation.FactionB == fid) {
			delete(gs.Relations, key)
		}
	}
	if len(gs.DiplomaticOffers) > 0 {
		offers := gs.DiplomaticOffers[:0]
		for _, offer := range gs.DiplomaticOffers {
			if offer.FromFactionID == fid || offer.ToFactionID == fid {
				continue
			}
			offers = append(offers, offer)
		}
		gs.DiplomaticOffers = offers
	}
	if len(gs.TradeRoutes) > 0 {
		routes := gs.TradeRoutes[:0]
		for _, route := range gs.TradeRoutes {
			if route == nil || route.FromFactionID == string(fid) || route.ToFactionID == string(fid) {
				continue
			}
			routes = append(routes, route)
		}
		gs.TradeRoutes = routes
	}
}
