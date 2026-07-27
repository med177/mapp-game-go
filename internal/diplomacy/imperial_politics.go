package diplomacy

import (
	"sort"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

// ImperialDietReport turn çözümlemesinde otomatik gerçekleşen siyasi sonucu
// oyun katmanının olay günlüğüne aktarır.
type ImperialDietReport struct {
	Held            bool
	ElectionHeld    bool
	Pending         bool
	WinnerID        faction.FactionID
	AuthorityBefore int
	AuthorityAfter  int
	Message         string
}

type ImperialElectionResult struct {
	WinnerID faction.FactionID
	Votes    map[faction.FactionID]faction.FactionID
	Totals   map[faction.FactionID]int
}

// AdvanceImperialPolitics Diyet takvimini ve seçim tarihini save-backed state
// üzerinden ilerletir. Oyuncu kararı gerektiren teklif/oy UI'sı gelene kadar
// AI-only kampanyalarda deterministik bir varsayılan karar uygulanır.
func AdvanceImperialPolitics(gs *state.GameState) ImperialDietReport {
	report := ImperialDietReport{}
	if gs == nil || gs.Imperial == nil {
		return report
	}
	imperial := gs.Imperial
	imperial.Clamp()
	report.AuthorityBefore = imperial.Authority
	if imperial.PendingDecision != nil {
		report.Pending = true
		report.AuthorityAfter = imperial.Authority
		return report
	}

	if imperial.NextDietTurn > 0 && gs.Turn >= imperial.NextDietTurn {
		imperial.NextDietTurn = gs.Turn + 12
		if gs.PlayerFactionID == imperial.EmpireID {
			imperial.PendingDecision = &state.ImperialPendingDecision{
				Kind:        state.ImperialDecisionDiet,
				CreatedTurn: gs.Turn,
			}
			report.Pending = true
		} else {
			report.Held = true
			holdImperialDiet(gs)
		}
	}
	if !report.Pending && imperial.ElectionDueTurn > 0 && gs.Turn >= imperial.ElectionDueTurn {
		imperial.ElectionDueTurn = 0
		if gs.PlayerFactionID == imperial.EmpireID {
			imperial.PendingDecision = &state.ImperialPendingDecision{
				Kind:        state.ImperialDecisionElection,
				CreatedTurn: gs.Turn,
			}
			report.Pending = true
		} else {
			election := HoldImperialElection(gs)
			if election.WinnerID != "" {
				report.ElectionHeld = true
				report.WinnerID = election.WinnerID
			}
		}
	}

	report.AuthorityAfter = imperial.Authority
	if report.Held || report.ElectionHeld {
		report.Message = imperialReportMessage(gs, report)
	}
	return report
}

// ResolveImperialDiet oyuncu HRE'sinin bekleyen Diyet kararını uygular.
// choice: 0 merkezileşme, 1 imtiyazlar, 2 askerî katkı.
func ResolveImperialDiet(gs *state.GameState, choice int) (bool, string) {
	if gs == nil || gs.Imperial == nil || gs.PlayerFactionID != gs.Imperial.EmpireID ||
		gs.Imperial.PendingDecision == nil || gs.Imperial.PendingDecision.Kind != state.ImperialDecisionDiet {
		return false, "Bekleyen Diyet kararı yok."
	}
	if choice < 0 || choice > 2 {
		return false, "Geçersiz Diyet kararı."
	}

	message := ""
	switch choice {
	case 0:
		gs.Imperial.Authority = clamp(gs.Imperial.Authority+5, 0, 100)
		for _, member := range gs.Imperial.Members {
			if member != nil {
				member.Loyalty = clamp(member.Loyalty-2, 0, 100)
			}
		}
		message = "Diyet merkezi otoriteyi güçlendirdi."
	case 1:
		gs.Imperial.Authority = clamp(gs.Imperial.Authority+1, 0, 100)
		for _, member := range gs.Imperial.Members {
			if member != nil {
				member.Loyalty = clamp(member.Loyalty+3, 0, 100)
				member.Autonomy = clamp(member.Autonomy+2, 0, 100)
			}
		}
		message = "Diyet prensliklerin imtiyazlarını korudu."
	case 2:
		gs.Imperial.Authority = clamp(gs.Imperial.Authority+2, 0, 100)
		for _, member := range gs.Imperial.Members {
			if member != nil {
				member.MilitaryCommitment = clamp(member.MilitaryCommitment+6, 0, 100)
				member.Loyalty = clamp(member.Loyalty-1, 0, 100)
			}
		}
		message = "Diyet imparatorluk üyelerinden askerî katkı talep etti."
	}
	gs.Imperial.PendingDecision = nil
	return true, message
}

// ResolveImperialElection oyuncunun seçtiği geçerli adayı imparator yapar.
// Üye oylarının ayrıntısı HoldImperialElection ile hesaplanır; oyuncu HRE'si
// kendi siyasi etkisini kullanarak geçerli adaylardan birini tercih edebilir.
func ResolveImperialElection(gs *state.GameState, candidate faction.FactionID) (bool, string) {
	if gs == nil || gs.Imperial == nil || gs.PlayerFactionID != gs.Imperial.EmpireID ||
		gs.Imperial.PendingDecision == nil || gs.Imperial.PendingDecision.Kind != state.ImperialDecisionElection {
		return false, "Bekleyen imparatorluk seçimi yok."
	}
	valid := false
	for _, available := range imperialElectionCandidates(gs) {
		if available == candidate {
			valid = true
			break
		}
	}
	if !valid {
		return false, "Bu aday imparatorluk seçiminde geçerli değil."
	}
	gs.Imperial.EmperorID = candidate
	gs.Imperial.Authority = clamp(gs.Imperial.Authority+8, 0, 100)
	gs.Imperial.PendingDecision = nil
	return true, "Yeni imparator: " + factionLabel(gs, candidate) + "."
}

func holdImperialDiet(gs *state.GameState) {
	if gs == nil || gs.Imperial == nil {
		return
	}
	total, count := 0, 0
	for _, member := range gs.Imperial.Members {
		if member == nil {
			continue
		}
		total += member.Loyalty
		count++
	}
	if count == 0 {
		return
	}
	average := total / count
	switch {
	case average >= 70:
		gs.Imperial.Authority += 2
	case average < 45:
		gs.Imperial.Authority -= 3
	default:
		gs.Imperial.Authority -= 1
	}
	gs.Imperial.Clamp()
}

// HoldImperialElection her üyenin ağırlıklı tek oy kullandığı deterministik
// seçimdir. 1300 senaryosundaki henüz sabitlenmemiş elektör yapısı için üyelik
// verisindeki ElectorWeight kullanılır; ağırlık yoksa prens bir oy verir.
func HoldImperialElection(gs *state.GameState) ImperialElectionResult {
	result := ImperialElectionResult{
		Votes:  make(map[faction.FactionID]faction.FactionID),
		Totals: make(map[faction.FactionID]int),
	}
	if gs == nil || gs.Imperial == nil {
		return result
	}
	candidates := imperialElectionCandidates(gs)
	if len(candidates) == 0 {
		return result
	}
	electors := ImperialMembersOf(gs, gs.Imperial.EmpireID)
	for _, electorID := range electors {
		member := gs.Imperial.Members[electorID]
		if member == nil {
			continue
		}
		candidate := bestImperialCandidate(gs, electorID, candidates)
		if candidate == "" {
			continue
		}
		weight := member.ElectorWeight
		if weight <= 0 {
			weight = 1
		}
		result.Votes[electorID] = candidate
		result.Totals[candidate] += weight
	}

	for _, candidate := range candidates {
		if result.Totals[candidate] > result.Totals[result.WinnerID] ||
			(result.Totals[candidate] == result.Totals[result.WinnerID] && candidate < result.WinnerID) {
			result.WinnerID = candidate
		}
	}
	if result.WinnerID == "" {
		return result
	}
	gs.Imperial.EmperorID = result.WinnerID
	gs.Imperial.Authority = clamp(gs.Imperial.Authority+8, 0, 100)
	return result
}

func imperialElectionCandidates(gs *state.GameState) []faction.FactionID {
	if gs == nil || gs.Imperial == nil {
		return nil
	}
	set := map[faction.FactionID]struct{}{}
	if empire := gs.Imperial.EmpireID; empire != "" && gs.Factions[empire] != nil && !gs.Factions[empire].IsEliminated {
		set[empire] = struct{}{}
	}
	for _, memberID := range ImperialMembersOf(gs, gs.Imperial.EmpireID) {
		member := gs.Imperial.Members[memberID]
		if member == nil || member.ElectorWeight <= 0 {
			continue
		}
		set[memberID] = struct{}{}
	}
	candidates := make([]faction.FactionID, 0, len(set))
	for id := range set {
		candidates = append(candidates, id)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i] < candidates[j] })
	return candidates
}

func bestImperialCandidate(gs *state.GameState, elector faction.FactionID, candidates []faction.FactionID) faction.FactionID {
	best := faction.FactionID("")
	bestScore := -1 << 30
	for _, candidate := range candidates {
		score := 0
		if candidate == gs.Imperial.EmperorID {
			score += 20
		}
		if rel := Relation(gs, elector, candidate); rel != nil {
			score += rel.Score
		}
		score += MilitaryPower(gs, candidate) / 10
		score += landRegionCount(gs, candidate) * 2
		if HasDirectThreat(gs, candidate, elector) {
			score -= 12
		}
		if score > bestScore || (score == bestScore && candidate < best) {
			best = candidate
			bestScore = score
		}
	}
	return best
}

func imperialReportMessage(gs *state.GameState, report ImperialDietReport) string {
	if gs == nil || gs.Imperial == nil {
		return ""
	}
	message := "İmparatorluk Diyeti toplandı. Otorite " + itoa(report.AuthorityBefore) + " → " + itoa(report.AuthorityAfter) + "."
	if report.ElectionHeld {
		message += " Yeni imparator: " + factionLabel(gs, report.WinnerID) + "."
	}
	return message
}
