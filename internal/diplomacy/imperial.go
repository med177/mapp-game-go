package diplomacy

import (
	"sort"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

const (
	imperialCallAcceptThreshold = 55
	imperialCallAutoThreshold   = 80
	imperialLimitedThreshold    = 35
	imperialSupportGoldLimit    = 30
	imperialSupportGrainLimit   = 45
)

// ImperialWarCallAssessment bir imparatorluk üyesinin çağrıya vereceği
// cevabı, ordinary alliance ilişkisine ihtiyaç duymadan hesaplar.
type ImperialWarCallAssessment struct {
	Chance         int
	BlockReason    string
	AutoJoin       bool
	LimitedSupport bool
	StatusTR       string
}

func (a ImperialWarCallAssessment) Accepted() bool {
	return a.BlockReason == "" && (a.AutoJoin || a.Chance >= imperialCallAcceptThreshold)
}

// ImperialMembersOf dışarıdan görülebilen, fakat gerçek vassal realm'lerini
// değiştirmeyen imparatorluk üyelerini deterministik sırada döndürür.
func ImperialMembersOf(gs *state.GameState, empire faction.FactionID) []faction.FactionID {
	if gs == nil || gs.Imperial == nil || empire == "" || gs.Imperial.EmpireID != empire {
		return nil
	}
	members := make([]faction.FactionID, 0, len(gs.Imperial.Members))
	for id, member := range gs.Imperial.Members {
		if member == nil || id == empire {
			continue
		}
		if f := gs.Factions[id]; f == nil || f.IsEliminated {
			continue
		}
		members = append(members, id)
	}
	sort.Slice(members, func(i, j int) bool {
		return factionLabel(gs, members[i]) < factionLabel(gs, members[j])
	})
	return members
}

func imperialMemberFor(gs *state.GameState, empire, memberID faction.FactionID) *state.ImperialMember {
	if gs == nil || gs.Imperial == nil || gs.Imperial.EmpireID != empire {
		return nil
	}
	member := gs.Imperial.Members[memberID]
	if member == nil || gs.Factions[memberID] == nil || gs.Factions[memberID].IsEliminated {
		return nil
	}
	return member
}

// AssessImperialWarCall HRE üyesinin doğrudan savaşa girme veya sınırlı
// yardım gönderme ihtimalini hesaplar. SameRealm üyeleri burada tekrar
// değerlendirilmez; onlar mevcut vassal koalisyon kuralıyla otomatik gelir.
func AssessImperialWarCall(gs *state.GameState, empire, memberID, enemy faction.FactionID) ImperialWarCallAssessment {
	assessment := ImperialWarCallAssessment{}
	if gs == nil || empire == "" || memberID == "" || enemy == "" || empire == memberID || empire == enemy || memberID == enemy {
		assessment.BlockReason = "Geçersiz imparatorluk çağrısı"
		return assessment
	}
	member := imperialMemberFor(gs, empire, memberID)
	if member == nil {
		assessment.BlockReason = "İmparatorluk üyesi değil"
		return assessment
	}
	if sameRealm(gs, empire, memberID) {
		assessment.BlockReason = "Realm içi otomatik katılım"
		assessment.AutoJoin = true
		assessment.Chance = 100
		assessment.StatusTR = "Realm içi otomatik katılım"
		return assessment
	}
	if IsWar(gs, memberID, enemy) {
		assessment.AutoJoin = true
		assessment.Chance = 100
		assessment.StatusTR = "Zaten savaşta"
		return assessment
	}
	if sameRealm(gs, empire, enemy) {
		assessment.BlockReason = "Hedef imparatorluk realm'inde"
		return assessment
	}

	authority := gs.Imperial.Authority
	chance := 14 + member.Loyalty/3 + member.MilitaryCommitment/4 + authority/5
	chance -= member.Autonomy / 10

	if rel := Relation(gs, empire, memberID); rel != nil {
		chance += clamp(rel.Score/4, -15, 18)
	}
	if sharesBorder(gs, memberID, enemy) {
		chance += 22
	}
	if sharesBorder(gs, empire, enemy) {
		chance += 8
	}
	if HasDirectThreat(gs, enemy, memberID) {
		chance += 12
	}
	if HasCommonEnemy(gs, empire, memberID) {
		chance += 8
	}
	if hasExternalWar(gs, memberID, empire) {
		chance -= 18
	}
	if MilitaryPower(gs, memberID) == 0 {
		chance -= 12
	}
	if authority < 35 {
		chance -= 15
	}

	assessment.Chance = clamp(chance, 0, 100)
	assessment.LimitedSupport = assessment.Chance >= imperialLimitedThreshold && assessment.Chance < imperialCallAcceptThreshold
	switch {
	case assessment.Chance >= imperialCallAutoThreshold:
		assessment.AutoJoin = true
		assessment.StatusTR = "İmparatorluk çağrısına kesin katılır"
	case assessment.Chance >= imperialCallAcceptThreshold:
		assessment.StatusTR = "İmparatorluk çağrısına katılır"
	case assessment.LimitedSupport:
		assessment.StatusTR = "Sınırlı yardım gönderir"
	default:
		assessment.StatusTR = "Tarafsız kalır"
	}
	return assessment
}

func imperialWarCallOutcome(gs *state.GameState, empire, memberID, enemy faction.FactionID, warDeclarer faction.FactionID) WarCallOutcome {
	assessment := AssessImperialWarCall(gs, empire, memberID, enemy)
	outcome := WarCallOutcome{
		FactionID:      memberID,
		NameTR:         factionLabel(gs, memberID),
		AutoJoin:       assessment.AutoJoin,
		ImperialMember: true,
		StatusTR:       assessment.StatusTR,
	}
	if assessment.BlockReason != "" {
		outcome.StatusTR = assessment.BlockReason
		return outcome
	}
	if assessment.Accepted() {
		outcome.Joined = true
		if gs.Imperial != nil {
			if member := gs.Imperial.Members[memberID]; member != nil {
				member.Loyalty = clamp(member.Loyalty+1, 0, 100)
			}
		}
		return outcome
	}
	if assessment.LimitedSupport {
		gold := imperialSupportGold(gs, memberID)
		grain := imperialSupportGrain(gs, memberID)
		outcome.LimitedSupport = applyImperialLimitedSupport(gs, empire, memberID)
		if outcome.LimitedSupport {
			outcome.SupportGold = gold
			outcome.SupportGrain = grain
			return outcome
		}
	}
	if gs.Imperial != nil {
		if member := gs.Imperial.Members[memberID]; member != nil {
			member.Loyalty = clamp(member.Loyalty-1, 0, 100)
		}
	}
	return outcome
}

func imperialSupportGold(gs *state.GameState, memberID faction.FactionID) int {
	if gs == nil || gs.Factions[memberID] == nil {
		return 0
	}
	return min(imperialSupportGoldLimit, gs.Factions[memberID].Gold)
}

func imperialSupportGrain(gs *state.GameState, memberID faction.FactionID) int {
	if gs == nil || gs.Factions[memberID] == nil {
		return 0
	}
	return min(imperialSupportGrainLimit, gs.Factions[memberID].Grain)
}

func applyImperialLimitedSupport(gs *state.GameState, empire, memberID faction.FactionID) bool {
	if gs == nil || gs.Factions[empire] == nil || gs.Factions[memberID] == nil {
		return false
	}
	gold := imperialSupportGold(gs, memberID)
	grain := imperialSupportGrain(gs, memberID)
	if gold == 0 && grain == 0 {
		return false
	}
	member := gs.Factions[memberID]
	member.Gold -= gold
	member.Grain -= grain
	gs.Factions[empire].Gold += gold
	gs.Factions[empire].Grain += grain
	return true
}

func imperialWarCallMembers(gs *state.GameState, empire, enemy, warDeclarer faction.FactionID) []WarCallOutcome {
	if gs == nil || gs.Imperial == nil || gs.Imperial.EmpireID != empire {
		return nil
	}
	if gs.Imperial.LastWarCall == nil || gs.Imperial.LastWarCall.CallerID != empire || gs.Imperial.LastWarCall.EnemyID != enemy || gs.Imperial.LastWarCall.StartedTurn != gs.Turn {
		gs.Imperial.LastWarCall = &state.ImperialWarCall{
			CallerID: empire, EnemyID: enemy, Reason: "İmparatorluk savaşı", StartedTurn: gs.Turn,
		}
	}
	playerRoot := playerRealmRoot(gs)
	out := make([]WarCallOutcome, 0, len(gs.Imperial.Members))
	for _, memberID := range ImperialMembersOf(gs, empire) {
		if sameRealm(gs, empire, memberID) {
			continue
		}
		if memberID == playerRoot {
			if queuePlayerWarJoinOffer(gs, empire, enemy, warDeclarer) {
				out = append(out, WarCallOutcome{
					FactionID: memberID, NameTR: factionLabel(gs, memberID), PendingDecision: true, ImperialMember: true,
				})
			}
			continue
		}
		out = append(out, imperialWarCallOutcome(gs, empire, memberID, enemy, warDeclarer))
	}
	return out
}

func imperialPreviewMembers(gs *state.GameState, empire, enemy faction.FactionID, selectable bool) (auto, callable []WarParticipantPreview) {
	for _, memberID := range ImperialMembersOf(gs, empire) {
		if sameRealm(gs, empire, memberID) {
			continue
		}
		assessment := AssessImperialWarCall(gs, empire, memberID, enemy)
		entry := WarParticipantPreview{
			FactionID:      memberID,
			NameTR:         factionLabel(gs, memberID),
			RoleTR:         "İmparatorluk üyesi",
			StatusTR:       assessment.StatusTR,
			JoinChance:     assessment.Chance,
			AutoJoin:       assessment.AutoJoin,
			Selectable:     selectable && !assessment.AutoJoin && assessment.BlockReason == "",
			ImperialMember: true,
		}
		if assessment.BlockReason != "" {
			entry.StatusTR = assessment.BlockReason
		}
		if assessment.AutoJoin {
			auto = append(auto, entry)
		} else {
			callable = append(callable, entry)
		}
	}
	return auto, callable
}
