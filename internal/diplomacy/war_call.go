package diplomacy

import (
	"sort"
	"strings"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

const (
	warCallAcceptanceThreshold = 50
	warCallRefusalScorePenalty = 10
)

type WarCallAssessment struct {
	Chance      int
	BlockReason string
	AutoJoin    bool
	StatusTR    string
}

func (a WarCallAssessment) Accepted() bool {
	return a.BlockReason == "" && (a.AutoJoin || a.Chance >= warCallAcceptanceThreshold)
}

type WarParticipantPreview struct {
	FactionID      faction.FactionID
	NameTR         string
	RoleTR         string
	StatusTR       string
	NoteTR         string
	JoinChance     int
	AutoJoin       bool
	Selectable     bool
	VassalCount    int
	ImperialMember bool
}

type WarSidePreview struct {
	PrimaryFactionID faction.FactionID
	PrimaryNameTR    string
	AutoParticipants []WarParticipantPreview
	CallableAllies   []WarParticipantPreview
}

type WarDeclarationPreview struct {
	Attacker WarSidePreview
	Defender WarSidePreview
}

type WarCallOutcome struct {
	FactionID       faction.FactionID
	NameTR          string
	Joined          bool
	AutoJoin        bool
	PendingDecision bool
	AllianceBroken  bool
	RelationPenalty int
	ImperialMember  bool
	LimitedSupport  bool
	SupportGold     int
	SupportGrain    int
	StatusTR        string
}

type WarDeclarationResult struct {
	Result
	PlayerCalls []WarCallOutcome
	EnemyCalls  []WarCallOutcome
}

func WarCoalitionMembers(gs *state.GameState, fid faction.FactionID) []faction.FactionID {
	return warCoalition(gs, fid)
}

func BuildWarDeclarationPreview(gs *state.GameState, actor, target faction.FactionID) WarDeclarationPreview {
	actorRoot := realmRoot(gs, actor)
	if actorRoot == "" {
		actorRoot = actor
	}
	targetRoot := realmRoot(gs, target)
	if targetRoot == "" {
		targetRoot = target
	}
	return WarDeclarationPreview{
		Attacker: buildWarSidePreview(gs, actorRoot, targetRoot, true),
		Defender: buildWarSidePreview(gs, targetRoot, actorRoot, false),
	}
}

func ExecuteWarDeclaration(gs *state.GameState, actor, target faction.FactionID, calledAllies []faction.FactionID) WarDeclarationResult {
	if reason := ActionBlockReason(gs, actor, target, ActionDeclareWar); reason != "" {
		return WarDeclarationResult{
			Result: Result{Message: reason},
		}
	}

	actorRoot := realmRoot(gs, actor)
	if actorRoot == "" {
		actorRoot = actor
	}
	targetRoot := realmRoot(gs, target)
	if targetRoot == "" {
		targetRoot = target
	}

	alliancePenalties := applyAllianceWarRelationPenalties(gs, actorRoot, targetRoot)
	attackerCalls := resolveAttackerWarCalls(gs, actorRoot, targetRoot, calledAllies)
	defenderCalls := resolveAutoWarCalls(gs, targetRoot, actorRoot, actorRoot)

	setWarBetweenCoalitions(gs, actorRoot, targetRoot)
	for _, outcome := range attackerCalls {
		if outcome.Joined {
			setWarBetweenCoalitions(gs, outcome.FactionID, targetRoot)
		}
	}
	for _, outcome := range defenderCalls {
		if outcome.Joined {
			// Savaş kaydında ilk ilan eden taraf solda kalır; savunanın
			// çağırdığı müttefik bu yüzden hedef taraf olarak eklenir.
			setWarBetweenCoalitions(gs, actorRoot, outcome.FactionID)
		}
	}

	message := buildWarDeclarationMessage(gs, actorRoot, targetRoot, attackerCalls, defenderCalls, alliancePenalties)
	return WarDeclarationResult{
		Result: Result{
			Accepted: true,
			Applied:  true,
			Message:  message,
		},
		PlayerCalls: attackerCalls,
		EnemyCalls:  defenderCalls,
	}
}

func AssessWarCall(gs *state.GameState, caller, ally, enemy faction.FactionID) WarCallAssessment {
	assessment := WarCallAssessment{}
	if gs == nil || caller == "" || ally == "" || enemy == "" {
		assessment.BlockReason = "Geçersiz savaş çağrısı"
		return assessment
	}
	if caller == ally {
		assessment.AutoJoin = true
		assessment.Chance = 100
		assessment.StatusTR = "Zaten lider taraf"
		return assessment
	}
	if sameRealm(gs, caller, ally) {
		assessment.AutoJoin = true
		assessment.Chance = 100
		assessment.StatusTR = "Realm içi otomatik katılım"
		return assessment
	}

	callerRoot := realmRoot(gs, caller)
	if callerRoot == "" {
		callerRoot = caller
	}
	allyRoot := realmRoot(gs, ally)
	if allyRoot == "" {
		allyRoot = ally
	}
	enemyRoot := realmRoot(gs, enemy)
	if enemyRoot == "" {
		enemyRoot = enemy
	}
	if sameRealm(gs, allyRoot, enemyRoot) {
		assessment.BlockReason = "Hedefle aynı realm içinde"
		return assessment
	}

	rel := Relation(gs, callerRoot, allyRoot)
	if rel == nil || rel.Stance != faction.StanceAllied {
		assessment.BlockReason = "Aktif ittifak yok"
		return assessment
	}
	if enemyRel := Relation(gs, allyRoot, enemyRoot); enemyRel != nil && enemyRel.Stance == faction.StanceAllied {
		assessment.BlockReason = "Hedefle de müttefik"
		return assessment
	}
	if IsWar(gs, allyRoot, enemyRoot) {
		assessment.AutoJoin = true
		assessment.Chance = 100
		assessment.StatusTR = "Zaten savaşta"
		return assessment
	}

	chance := 20 + rel.Score
	chance += clamp(-relationScoreAgainst(gs, allyRoot, enemyRoot)/3, -10, 24)
	if HasTradeRouteBetween(gs, allyRoot, enemyRoot) {
		chance -= 14
	}
	if HasCommonEnemy(gs, callerRoot, allyRoot) {
		chance += 8
	}
	if HasSharedMajorThreat(gs, callerRoot, allyRoot) {
		chance += 8
	}
	if HasDirectThreat(gs, allyRoot, enemyRoot) {
		chance += 16
	}
	if HasDirectThreat(gs, callerRoot, allyRoot) {
		chance -= 10
	}

	callerPower := MilitaryPower(gs, callerRoot)
	allyPower := MilitaryPower(gs, allyRoot)
	enemyPower := MilitaryPower(gs, enemyRoot)
	switch {
	case allyPower > enemyPower:
		chance += min(10, (allyPower-enemyPower)/15)
	case enemyPower > allyPower:
		chance -= min(16, (enemyPower-allyPower)/12)
	}
	switch {
	case callerPower > enemyPower:
		chance += min(8, (callerPower-enemyPower)/20)
	case enemyPower > callerPower:
		chance -= min(8, (enemyPower-callerPower)/25)
	}

	chance += clamp(landRegionCount(gs, callerRoot)-landRegionCount(gs, enemyRoot), -6, 10)
	chance -= economicStress(gs, allyRoot) / 2
	if hasExternalWar(gs, allyRoot, callerRoot) {
		chance -= 8
	}

	assessment.Chance = clamp(chance, 0, 100)
	switch {
	case assessment.Chance >= 75:
		assessment.StatusTR = "Yüksek ihtimal"
	case assessment.Chance >= warCallAcceptanceThreshold:
		assessment.StatusTR = "Kararsız ama mümkün"
	default:
		assessment.StatusTR = "Düşük ihtimal"
	}
	return assessment
}

func directExternalAlliesOf(gs *state.GameState, fid faction.FactionID) []faction.FactionID {
	if gs == nil || fid == "" {
		return nil
	}
	allies := make([]faction.FactionID, 0, 6)
	for otherID, other := range gs.Factions {
		if other == nil || other.IsEliminated || otherID == fid || sameRealm(gs, fid, otherID) {
			continue
		}
		if rel := Relation(gs, fid, otherID); rel != nil && rel.Stance == faction.StanceAllied {
			allies = append(allies, otherID)
		}
	}
	sort.Slice(allies, func(i, j int) bool {
		return factionLabel(gs, allies[i]) < factionLabel(gs, allies[j])
	})
	return allies
}

func buildWarSidePreview(gs *state.GameState, leader, enemy faction.FactionID, selectableAllies bool) WarSidePreview {
	side := WarSidePreview{
		PrimaryFactionID: leader,
		PrimaryNameTR:    factionLabel(gs, leader),
	}

	for _, vassalID := range VassalsOf(gs, leader) {
		side.AutoParticipants = append(side.AutoParticipants, WarParticipantPreview{
			FactionID:   vassalID,
			NameTR:      factionLabel(gs, vassalID),
			RoleTR:      "Vassal",
			StatusTR:    "Kesin katılır",
			NoteTR:      "Overlord savaşa girince otomatik dahil olur.",
			JoinChance:  100,
			AutoJoin:    true,
			Selectable:  false,
			VassalCount: 0,
		})
	}

	for _, allyID := range directExternalAlliesOf(gs, leader) {
		assessment := AssessWarCall(gs, leader, allyID, enemy)
		entry := WarParticipantPreview{
			FactionID:   allyID,
			NameTR:      factionLabel(gs, allyID),
			RoleTR:      "Müttefik",
			StatusTR:    assessment.StatusTR,
			JoinChance:  assessment.Chance,
			AutoJoin:    assessment.AutoJoin,
			Selectable:  selectableAllies && !assessment.AutoJoin && assessment.BlockReason == "",
			VassalCount: len(VassalsOf(gs, allyID)),
		}
		if entry.VassalCount > 0 {
			entry.NoteTR = "Katılırsa " + itoa(entry.VassalCount) + " vassalı da gelir."
		}
		if assessment.BlockReason != "" {
			entry.StatusTR = assessment.BlockReason
			entry.NoteTR = ""
		}
		if assessment.AutoJoin {
			if entry.NoteTR == "" {
				entry.NoteTR = "Bu cephede zaten aktif."
			}
			side.AutoParticipants = append(side.AutoParticipants, entry)
			continue
		}
		side.CallableAllies = append(side.CallableAllies, entry)
	}
	imperialAuto, imperialCallable := imperialPreviewMembers(gs, leader, enemy, selectableAllies)
	side.AutoParticipants = append(side.AutoParticipants, imperialAuto...)
	side.CallableAllies = append(side.CallableAllies, imperialCallable...)

	sortWarParticipants(side.AutoParticipants)
	sortWarParticipants(side.CallableAllies)
	return side
}

func sortWarParticipants(entries []WarParticipantPreview) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].AutoJoin != entries[j].AutoJoin {
			return entries[i].AutoJoin
		}
		if entries[i].JoinChance != entries[j].JoinChance {
			return entries[i].JoinChance > entries[j].JoinChance
		}
		return entries[i].NameTR < entries[j].NameTR
	})
}

func resolvePlayerWarCalls(gs *state.GameState, actorRoot, targetRoot faction.FactionID, calledAllies []faction.FactionID) []WarCallOutcome {
	if len(calledAllies) == 0 {
		return nil
	}
	seen := make(map[faction.FactionID]struct{}, len(calledAllies))
	out := make([]WarCallOutcome, 0, len(calledAllies))
	for _, allyID := range calledAllies {
		allyRoot := realmRoot(gs, allyID)
		if allyRoot == "" {
			allyRoot = allyID
		}
		if allyRoot == actorRoot || allyRoot == targetRoot {
			continue
		}
		if _, exists := seen[allyRoot]; exists {
			continue
		}
		seen[allyRoot] = struct{}{}
		if imperialMemberFor(gs, actorRoot, allyRoot) != nil {
			out = append(out, imperialWarCallOutcome(gs, actorRoot, allyRoot, targetRoot, actorRoot))
			continue
		}
		out = append(out, resolveWarCall(gs, actorRoot, allyRoot, targetRoot, true))
	}
	return out
}

func resolveAttackerWarCalls(gs *state.GameState, actorRoot, targetRoot faction.FactionID, calledAllies []faction.FactionID) []WarCallOutcome {
	if actorRoot == playerRealmRoot(gs) {
		return resolvePlayerWarCalls(gs, actorRoot, targetRoot, calledAllies)
	}
	return resolveAutoWarCalls(gs, actorRoot, targetRoot, actorRoot)
}

func resolveAutoWarCalls(gs *state.GameState, callerRoot, enemyRoot, warDeclarerRoot faction.FactionID) []WarCallOutcome {
	allies := directExternalAlliesOf(gs, callerRoot)
	playerRoot := playerRealmRoot(gs)
	out := make([]WarCallOutcome, 0, len(allies)+4)
	for _, allyID := range allies {
		allyRoot := realmRoot(gs, allyID)
		if allyRoot == "" {
			allyRoot = allyID
		}
		assessment := AssessWarCall(gs, callerRoot, allyRoot, enemyRoot)
		if assessment.BlockReason != "" {
			continue
		}
		if allyRoot == playerRoot {
			if queuePlayerWarJoinOffer(gs, callerRoot, enemyRoot, warDeclarerRoot) {
				out = append(out, WarCallOutcome{
					FactionID:       allyRoot,
					NameTR:          factionLabel(gs, allyRoot),
					PendingDecision: true,
				})
			}
			continue
		}
		out = append(out, resolveWarCall(gs, callerRoot, allyRoot, enemyRoot, false))
	}
	if gs != nil && gs.Imperial != nil && gs.Imperial.EmpireID == callerRoot {
		out = append(out, imperialWarCallMembers(gs, callerRoot, enemyRoot, warDeclarerRoot)...)
	}
	return out
}

func queuePlayerWarJoinOffer(gs *state.GameState, callerRoot, enemyRoot, warDeclarerRoot faction.FactionID) bool {
	if gs == nil || gs.PlayerFactionID == "" || callerRoot == "" || enemyRoot == "" || warDeclarerRoot == "" {
		return false
	}
	return QueueWarJoinOffer(gs, callerRoot, gs.PlayerFactionID, warDeclarerRoot, enemyRoot, "Aktif ittifak savaş çağrısı")
}

// alreadyAtWar, savaş çağrısının alıcısının hedef düşmanla zaten aynı cephede
// olup olmadığını realm kökleri üzerinden kontrol eder. Teklif kuyruğu ve modal
// seçimi aynı geçerlilik kuralını kullanmalıdır.
func alreadyAtWar(gs *state.GameState, first, second faction.FactionID) bool {
	if gs == nil || first == "" || second == "" || first == second {
		return false
	}
	firstRoot := realmRoot(gs, first)
	if firstRoot == "" {
		firstRoot = first
	}
	secondRoot := realmRoot(gs, second)
	if secondRoot == "" {
		secondRoot = second
	}
	return IsWar(gs, firstRoot, secondRoot)
}

func resolveWarCall(gs *state.GameState, callerRoot, allyRoot, enemyRoot faction.FactionID, breakOnRefusal bool) WarCallOutcome {
	assessment := AssessWarCall(gs, callerRoot, allyRoot, enemyRoot)
	outcome := WarCallOutcome{
		FactionID: allyRoot,
		NameTR:    factionLabel(gs, allyRoot),
		AutoJoin:  assessment.AutoJoin,
		Joined:    assessment.Accepted(),
	}
	if outcome.Joined || !breakOnRefusal {
		if !outcome.Joined && !breakOnRefusal {
			breakAllianceForWarRefusal(gs, callerRoot, allyRoot)
			outcome.AllianceBroken = true
			outcome.RelationPenalty = warCallRefusalScorePenalty
		}
		return outcome
	}
	breakAllianceForWarRefusal(gs, callerRoot, allyRoot)
	outcome.AllianceBroken = true
	outcome.RelationPenalty = warCallRefusalScorePenalty
	return outcome
}

func breakAllianceForWarRefusal(gs *state.GameState, caller, ally faction.FactionID) {
	if gs == nil || caller == "" || ally == "" {
		return
	}
	rel := EnsureRelation(gs, caller, ally)
	if HasTradeRouteBetween(gs, caller, ally) {
		rel.Stance = faction.StanceTrade
	} else {
		rel.Stance = faction.StancePeace
	}
	rel.Score = clamp(rel.Score-warCallRefusalScorePenalty, -100, 100)
}

func buildWarDeclarationMessage(gs *state.GameState, actorRoot, targetRoot faction.FactionID, attackerCalls, defenderCalls []WarCallOutcome, alliancePenalties []AllianceWarRelationPenalty) string {
	parts := []string{factionLabel(gs, actorRoot) + " ile " + factionLabel(gs, targetRoot) + " arasında savaş başladı."}
	if penaltyText := allianceWarPenaltyMessage(gs, alliancePenalties); penaltyText != "" {
		parts = append(parts, penaltyText)
	}
	if joined := joinedNames(attackerCalls); joined != "" {
		parts = append(parts, factionLabel(gs, actorRoot)+" tarafına katılanlar: "+joined+".")
	}
	if pending := pendingNames(attackerCalls); pending != "" {
		parts = append(parts, factionLabel(gs, actorRoot)+" tarafında cevap beklenenler: "+pending+".")
	}
	if refused := refusedNames(attackerCalls); refused != "" {
		parts = append(parts, factionLabel(gs, actorRoot)+" tarafında çağrıyı reddedenler: "+refused+" (ittifak bozuldu).")
	}
	if limited := limitedSupportNames(attackerCalls); limited != "" {
		parts = append(parts, factionLabel(gs, actorRoot)+" tarafına sınırlı yardım gönderenler: "+limited+".")
	}
	if joined := joinedNames(defenderCalls); joined != "" {
		parts = append(parts, factionLabel(gs, targetRoot)+" tarafına katılanlar: "+joined+".")
	}
	if pending := pendingNames(defenderCalls); pending != "" {
		parts = append(parts, factionLabel(gs, targetRoot)+" tarafında cevap beklenenler: "+pending+".")
	}
	if refused := refusedNames(defenderCalls); refused != "" {
		parts = append(parts, factionLabel(gs, targetRoot)+" tarafında çağrıyı reddedenler: "+refused+".")
	}
	if limited := limitedSupportNames(defenderCalls); limited != "" {
		parts = append(parts, factionLabel(gs, targetRoot)+" tarafına sınırlı yardım gönderenler: "+limited+".")
	}
	return strings.Join(parts, " ")
}

func allianceWarPenaltyMessage(gs *state.GameState, penalties []AllianceWarRelationPenalty) string {
	if gs == nil || len(penalties) == 0 {
		return ""
	}
	parts := make([]string, 0, len(penalties))
	for _, penalty := range penalties {
		name := factionLabel(gs, penalty.FactionID)
		if penalty.AllianceBroken {
			parts = append(parts, name+" ile ittifak bozuldu (ilişki -"+itoa(penalty.RelationPenalty)+")")
			continue
		}
		parts = append(parts, name+" ile ilişki -"+itoa(penalty.RelationPenalty))
	}
	return "Müttefik saldırısı nedeniyle: " + strings.Join(parts, ", ") + "."
}

func joinedNames(outcomes []WarCallOutcome) string {
	names := make([]string, 0, len(outcomes))
	for _, outcome := range outcomes {
		if outcome.Joined {
			names = append(names, outcome.NameTR)
		}
	}
	return strings.Join(names, ", ")
}

func refusedNames(outcomes []WarCallOutcome) string {
	names := make([]string, 0, len(outcomes))
	for _, outcome := range outcomes {
		if !outcome.Joined && !outcome.PendingDecision && !outcome.LimitedSupport {
			names = append(names, outcome.NameTR)
		}
	}
	return strings.Join(names, ", ")
}

func limitedSupportNames(outcomes []WarCallOutcome) string {
	names := make([]string, 0, len(outcomes))
	for _, outcome := range outcomes {
		if outcome.LimitedSupport {
			names = append(names, outcome.NameTR)
		}
	}
	return strings.Join(names, ", ")
}

func pendingNames(outcomes []WarCallOutcome) string {
	names := make([]string, 0, len(outcomes))
	for _, outcome := range outcomes {
		if outcome.PendingDecision {
			names = append(names, outcome.NameTR)
		}
	}
	return strings.Join(names, ", ")
}

func playerRealmRoot(gs *state.GameState) faction.FactionID {
	if gs == nil || gs.PlayerFactionID == "" {
		return ""
	}
	playerRoot := realmRoot(gs, gs.PlayerFactionID)
	if playerRoot == "" {
		return gs.PlayerFactionID
	}
	return playerRoot
}

func relationScoreAgainst(gs *state.GameState, a, b faction.FactionID) int {
	rel := Relation(gs, a, b)
	if rel == nil {
		return 0
	}
	return rel.Score
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	var buf [12]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return sign + string(buf[i:])
}
