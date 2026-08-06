package diplomacy

import (
	"sort"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// Açılış ekonomi/ordu bandı ilk 42 turda korunur; daha sonra AI-AI barışları
// somut tazminat/toprak/vassallık sonucu uygulayabilir.
const peaceSettlementActivationTurn = 42

// PeaceOutcome savaşın barışla kapanırken ortaya çıkan siyasi sonucu belirtir.
// Normal oyuncu barış aksiyonu güvenli biçimde beyaz barıştır; AI-AI akışı
// AssessPeaceSettlement sonucunu otomatik uygulayabilir.
type PeaceOutcome string

const (
	PeaceOutcomeWhitePeace  PeaceOutcome = "white_peace"
	PeaceOutcomeCedeRegion  PeaceOutcome = "cede_region"
	PeaceOutcomeReparations PeaceOutcome = "reparations"
	PeaceOutcomeVassalage   PeaceOutcome = "vassalage"
)

// PeaceSettlement barış sonucu için mutasyondan önce üretilen deterministik
// karardır. RegionID yalnız bölge bırakma, Gold ise tazminat sonuçlarında doludur.
type PeaceSettlement struct {
	Outcome  PeaceOutcome
	Winner   faction.FactionID
	Loser    faction.FactionID
	RegionID world.RegionID
	Gold     int
}

// AssessPeaceSettlement savaşın mevcut ledger ve sahiplik durumuna göre barış
// sonucunu seçer. Bu fonksiyon state'i değiştirmez.
func AssessPeaceSettlement(gs *state.GameState, proposer, responder faction.FactionID) PeaceSettlement {
	settlement := PeaceSettlement{Outcome: PeaceOutcomeWhitePeace}
	if gs == nil || proposer == "" || responder == "" || proposer == responder || !IsWar(gs, proposer, responder) {
		return settlement
	}
	proposerAssessment := AssessPeaceDesire(gs, proposer, responder)
	responderAssessment := AssessPeaceDesire(gs, responder, proposer)
	winner := faction.FactionID("")
	if proposerAssessment.WarScore >= 40 && proposerAssessment.WarScore > responderAssessment.WarScore {
		winner = proposer
	} else if responderAssessment.WarScore >= 40 {
		winner = responder
	}
	if winner == "" {
		return settlement
	}
	loser := responder
	if winner == responder {
		loser = proposer
	}
	settlement.Winner = winner
	settlement.Loser = loser

	winnerAssessment := proposerAssessment
	if winner == responder {
		winnerAssessment = responderAssessment
	}
	if canImposeVassalage(gs, winner, loser, winnerAssessment) {
		settlement.Outcome = PeaceOutcomeVassalage
		return settlement
	}
	if regionID := peaceCessionRegion(gs, winner, loser); regionID != "" {
		settlement.Outcome = PeaceOutcomeCedeRegion
		settlement.RegionID = regionID
		return settlement
	}
	if gold := peaceReparationAmount(gs, winner, loser); gold > 0 {
		settlement.Outcome = PeaceOutcomeReparations
		settlement.Gold = gold
	}
	return settlement
}

// ExecuteAIPeace AI-AI barışında sınıflandırılmış siyasi sonucu uygular.
// Oyuncu aksiyonlarında Execute kullanılır; böylece oyuncu açıkça seçmeden
// toprak veya altın kaybetmez.
func ExecuteAIPeace(gs *state.GameState, proposer, responder faction.FactionID) Result {
	if gs == nil || proposer == "" || responder == "" || proposer == responder {
		return Result{Message: "Geçersiz barış hedefi."}
	}
	rel := Relation(gs, proposer, responder)
	if rel == nil || rel.Stance != faction.StanceWar {
		return Result{Message: "Barış teklifi sadece aktif savaşta uygulanır."}
	}
	if !spendDiplomacyOfferQuota(gs, proposer) {
		return Result{Message: diplomacyOfferQuotaBlockReasonTR}
	}
	if !assessPeaceAcceptance(gs, proposer, responder).ShouldPropose() {
		return Result{Message: factionLabel(gs, responder) + " barışı reddetti."}
	}
	settlement := AssessPeaceSettlement(gs, proposer, responder)
	setPeaceBetweenCoalitions(gs, proposer, responder)
	result := Result{Accepted: true, Applied: true, Settlement: &settlement, Message: factionLabel(gs, responder) + " barışı kabul etti."}
	if settlement.Outcome != PeaceOutcomeWhitePeace &&
		(gs.ScenarioID != "1300_ottoman_rise" || gs.Turn > peaceSettlementActivationTurn) {
		if applied := applyPeaceSettlement(gs, settlement); applied {
			result.Message += " " + peaceSettlementMessageTR(gs, settlement)
		}
	}
	return result
}

func canImposeVassalage(gs *state.GameState, winner, loser faction.FactionID, winnerAssessment PeaceAssessment) bool {
	if gs == nil || winner == "" || loser == "" || DirectOverlord(gs, winner) != "" || DirectOverlord(gs, loser) != "" {
		return false
	}
	if len(gs.LandRegionsOwnedBy(loser)) > 4 {
		return false
	}
	winnerPower := MilitaryPower(gs, winner)
	loserPower := MilitaryPower(gs, loser)
	if loserPower > 0 && winnerPower < loserPower*2 {
		return false
	}
	return winnerAssessment.WarScore >= 70
}

func peaceCessionRegion(gs *state.GameState, winner, loser faction.FactionID) world.RegionID {
	if gs == nil || winner == "" || loser == "" {
		return ""
	}
	candidates := make([]*world.Region, 0, len(gs.Regions))
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(loser) || isCapitalRegion(gs, loser, region.ID) {
			continue
		}
		candidates = append(candidates, region)
	}
	sort.Slice(candidates, func(i, j int) bool {
		iBorder := regionBordersFaction(gs, candidates[i], winner)
		jBorder := regionBordersFaction(gs, candidates[j], winner)
		if iBorder != jBorder {
			return iBorder
		}
		if candidates[i].FortificationLevel() != candidates[j].FortificationLevel() {
			return candidates[i].FortificationLevel() < candidates[j].FortificationLevel()
		}
		return candidates[i].ID < candidates[j].ID
	})
	if len(candidates) == 0 {
		return ""
	}
	return candidates[0].ID
}

func regionBordersFaction(gs *state.GameState, region *world.Region, fid faction.FactionID) bool {
	if gs == nil || region == nil {
		return false
	}
	for _, neighborID := range region.Neighbors {
		neighbor := gs.Regions[neighborID]
		if neighbor != nil && neighbor.OwnerID == string(fid) {
			return true
		}
	}
	return false
}

func isCapitalRegion(gs *state.GameState, fid faction.FactionID, regionID world.RegionID) bool {
	if gs == nil || gs.Factions[fid] == nil || gs.Factions[fid].CapitalSettlementID == "" {
		return false
	}
	region, _, _, ok := gs.FindSettlementByID(gs.Factions[fid].CapitalSettlementID)
	return ok && region != nil && region.ID == regionID
}

func peaceReparationAmount(gs *state.GameState, winner, loser faction.FactionID) int {
	if gs == nil || gs.Factions[loser] == nil || gs.Factions[loser].Gold <= 0 {
		return 0
	}
	winnerPower := MilitaryPower(gs, winner)
	amount := winnerPower / 5
	if amount < 25 {
		amount = 25
	}
	if amount > 200 {
		amount = 200
	}
	if amount > gs.Factions[loser].Gold/2 {
		amount = gs.Factions[loser].Gold / 2
	}
	return amount
}

func applyPeaceSettlement(gs *state.GameState, settlement PeaceSettlement) bool {
	if gs == nil || settlement.Winner == "" || settlement.Loser == "" {
		return false
	}
	winner := gs.Factions[settlement.Winner]
	loser := gs.Factions[settlement.Loser]
	if winner == nil || loser == nil {
		return false
	}
	switch settlement.Outcome {
	case PeaceOutcomeCedeRegion:
		region := gs.Regions[settlement.RegionID]
		if region == nil || region.IsSea || region.OwnerID != string(settlement.Loser) || isCapitalRegion(gs, settlement.Loser, region.ID) {
			return false
		}
		region.ApplyConquest(string(settlement.Winner), string(winner.Religion))
		gs.ClearProductionOrdersForRegion(region.ID)
		return true
	case PeaceOutcomeReparations:
		amount := settlement.Gold
		if amount <= 0 || amount > loser.Gold {
			return false
		}
		loser.Gold -= amount
		winner.Gold += amount
		return true
	case PeaceOutcomeVassalage:
		return applyVassalization(gs, settlement.Winner, settlement.Loser).Applied
	default:
		return false
	}
}

func peaceSettlementMessageTR(gs *state.GameState, settlement PeaceSettlement) string {
	switch settlement.Outcome {
	case PeaceOutcomeCedeRegion:
		return factionLabel(gs, settlement.Loser) + " " + regionLabel(gs, settlement.RegionID) + " bölgesini bıraktı."
	case PeaceOutcomeReparations:
		return factionLabel(gs, settlement.Loser) + " " + itoa(settlement.Gold) + " altın tazminat ödedi."
	case PeaceOutcomeVassalage:
		return factionLabel(gs, settlement.Loser) + " vassal oldu."
	default:
		return ""
	}
}

func regionLabel(gs *state.GameState, rid world.RegionID) string {
	if gs != nil && gs.Regions[rid] != nil && gs.Regions[rid].NameTR != "" {
		return gs.Regions[rid].NameTR
	}
	return string(rid)
}
