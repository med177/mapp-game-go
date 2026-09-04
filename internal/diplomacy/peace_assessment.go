package diplomacy

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	peaceDesireThreshold   = 42
	peaceMinimumWarTurns   = 4
	peaceOfferCooldown     = 1
	peaceEmergencyDiscount = 10
)

// PeaceProposalAssessment, teklif ekranı ile gerçek barış kararının aynı
// değerlendirmeyi kullanmasını sağlar. Chance burada rastgele zar sonucu
// değildir; eşik tabanlı diplomasi modelindeki kabul kesinliğidir. Bu nedenle
// Accepted false iken Chance hiçbir zaman 100 olamaz.
type PeaceProposalAssessment struct {
	Chance      int
	Score       int
	Threshold   int
	Accepted    bool
	BlockReason string
}

// AssessPeaceProposal barış teklifini actor perspektifinden değerlendirir.
// Hedefin kabul kararını hesapladığı için renderer'ın yaklaşık bir formülle
// farklı sonuç göstermesini engeller.
func AssessPeaceProposal(gs *state.GameState, actor, target faction.FactionID) PeaceProposalAssessment {
	assessment := PeaceProposalAssessment{Threshold: peaceDesireThreshold}
	if gs == nil || actor == "" || target == "" || actor == target {
		assessment.BlockReason = "Geçersiz diplomasi hedefi"
		return assessment
	}
	rel := Relation(gs, actor, target)
	if rel == nil || rel.Stance != faction.StanceWar {
		assessment.BlockReason = "Barış teklifi sadece savaşta yapılır."
		return assessment
	}

	// Kabul kararı her senaryoda hedefin perspektifinden verilir. Teklif
	// sahibinin barış teknolojisi, hedef için yapılan anlaşma teşviki olarak
	// ayrıca değerlendirmeye eklenir.
	desire := assessPeaceAcceptance(gs, actor, target)
	assessment.Score = desire.Score
	assessment.Threshold = desire.Threshold
	assessment.Accepted = desire.ShouldPropose()
	return finalizePeaceProposalAssessment(assessment)
}

func assessPeaceAcceptance(gs *state.GameState, proposer, responder faction.FactionID) PeaceAssessment {
	assessment := AssessPeaceDesire(gs, responder, proposer)
	assessment.Score += peaceTechBonus(gs, proposer)
	return assessment
}

func finalizePeaceProposalAssessment(assessment PeaceProposalAssessment) PeaceProposalAssessment {
	if assessment.Accepted {
		assessment.Chance = 100
		return assessment
	}
	if assessment.Threshold <= 0 || assessment.Score <= 0 {
		return assessment
	}
	assessment.Chance = assessment.Score * 100 / assessment.Threshold
	if assessment.Chance >= 100 {
		assessment.Chance = 99
	}
	return assessment
}

// PeaceAssessment bir savaş tarafının barış isteğini açıklar. Score yükseldikçe
// savaşı bitirme baskısı artar.
type PeaceAssessment struct {
	Score                int
	WarScore             int
	WarExhaustion        int
	GoldPressure         int
	GrainPressure        int
	SatisfactionPressure int
	RelationshipPressure int
	FutureLossPressure   int
	Threshold            int
	WarTurns             int
	Eligible             bool
	Emergency            bool
	CapitalThreat        bool
	MilitaryCollapse     bool
	ObjectiveDone        bool
	Stalemate            bool
	OwnLosses            int
	EnemyLosses          int
	OwnRegionsLost       int
	RegionsGained        int
	ObjectiveHeld        int
	ObjectiveTotal       int
	UnresolvedCoreClaims int
	UnresolvedClaimValue int
	UnresolvedClaimCount int
}

func (a PeaceAssessment) ShouldPropose() bool {
	if !a.Eligible {
		return false
	}
	// Durgunluk, çözümsüz bir çekirdek/toprak talebini tek başına ortadan
	// kaldırmaz. Aksi halde AI dört tur hareketsiz kaldığında hedef bölgesini
	// almadan otomatik beyaz barışa döner.
	return a.Score >= a.Threshold || (a.Stalemate && a.UnresolvedClaimValue == 0)
}

// AssessPeaceDesire aktif savaşı actor perspektifinden değerlendirir. Tüm
// senaryolarda ortak savaş yorgunluğu, claim/core ve acil durum modelidir.
func AssessPeaceDesire(gs *state.GameState, actor, opponent faction.FactionID) PeaceAssessment {
	assessment := PeaceAssessment{Threshold: peaceDesireThreshold}
	if gs == nil || actor == "" || opponent == "" || actor == opponent || !IsWar(gs, actor, opponent) {
		return assessment
	}
	gs.SyncWarLedgers()
	ledger := gs.WarLedgerFor(actor, opponent)
	if ledger == nil {
		return assessment
	}
	assessment.WarTurns = max(0, gs.Turn-ledger.StartedTurn)
	assessment.OwnLosses, assessment.EnemyLosses = warCasualtiesFor(ledger, actor)
	actorCaptured, opponentCaptured := warCapturesFor(ledger, actor)
	assessment.RegionsGained = actorCaptured
	assessment.OwnRegionsLost = opponentCaptured

	initialActor, _ := warInitialRegionsFor(ledger, actor)
	currentActor := landRegionCount(gs, actor)
	assessment.OwnRegionsLost = max(assessment.OwnRegionsLost, max(0, initialActor-currentActor))
	assessment.WarScore, assessment.ObjectiveHeld, assessment.ObjectiveTotal = warScoreFor(gs, actor, opponent, ledger)
	assessment.WarExhaustion = warExhaustionFor(gs, actor, opponent, ledger, assessment.WarTurns, assessment.OwnLosses, assessment.OwnRegionsLost)
	assessment.GoldPressure = goldPressureFor(gs, actor)
	assessment.GrainPressure = grainPressureFor(gs, actor)
	assessment.SatisfactionPressure = satisfactionPressureFor(gs, actor)
	assessment.RelationshipPressure = relationshipPressureFor(gs, actor, opponent)

	assessment.Score += min(24, max(0, assessment.WarTurns-peaceMinimumWarTurns)*3)
	assessment.Score += assessment.OwnRegionsLost * 9
	assessment.Score -= assessment.RegionsGained * 7
	if assessment.OwnLosses > assessment.EnemyLosses {
		assessment.Score += min(20, (assessment.OwnLosses-assessment.EnemyLosses)*3)
	} else if assessment.EnemyLosses > assessment.OwnLosses {
		assessment.Score -= min(14, (assessment.EnemyLosses-assessment.OwnLosses)*2)
	}
	actorPower := MilitaryPower(gs, actor)
	opponentPower := MilitaryPower(gs, opponent)
	switch {
	case actorPower == 0 && opponentPower > 0:
		assessment.Score += 45
		assessment.MilitaryCollapse = true
	case opponentPower > 0 && actorPower*100 < opponentPower*55:
		assessment.Score += 38
		assessment.MilitaryCollapse = true
	case opponentPower > 0 && actorPower*100 < opponentPower*80:
		assessment.Score += 22
	case opponentPower > 0 && actorPower < opponentPower:
		assessment.Score += 10
	case actorPower > 0 && actorPower*100 > opponentPower*150:
		assessment.Score -= 18
	case actorPower > 0 && actorPower*100 > opponentPower*120:
		assessment.Score -= 10
	}

	assessment.Score += economicStress(gs, actor)
	assessment.Score += assessment.RelationshipPressure
	assessment.FutureLossPressure = futureLossPressureFor(gs, actor, opponent)
	assessment.Score += assessment.FutureLossPressure
	if extraWars := activeWarCount(gs, actor) - 1; extraWars > 0 {
		assessment.Score += min(24, extraWars*12)
	}

	assessment.CapitalThreat = capitalUnderWarThreat(gs, actor, opponent)
	if assessment.CapitalThreat {
		assessment.Score += 40
	}
	if capitalUnderWarThreat(gs, opponent, actor) {
		assessment.Score -= 15
	}

	assessment.UnresolvedCoreClaims, assessment.UnresolvedClaimValue, assessment.UnresolvedClaimCount = territorialClaimPressure(gs, actor, opponent, ledger)
	assessment.Score -= min(40, assessment.UnresolvedClaimValue/2)
	assessment.Threshold += min(20, assessment.UnresolvedClaimValue/5)
	if assessment.UnresolvedCoreClaims > 0 {
		assessment.Score -= 25
	}

	assessment.ObjectiveDone = warObjectiveCompleted(gs, actor, opponent)
	if assessment.ObjectiveDone {
		assessment.Score += 28
	} else if plan := gs.AIPlans[actor]; plan != nil && plan.TargetFactionID == opponent {
		commitmentPenalty := 8 + plan.Commitment/10
		assessment.Score -= min(20, commitmentPenalty)
	}

	if assessment.WarTurns >= 6 && (ledger.LastBattleTurn == 0 || gs.Turn-ledger.LastBattleTurn >= 4) {
		assessment.Score += 10
	}

	assessment.Emergency = assessment.CapitalThreat || assessment.MilitaryCollapse
	if assessment.Emergency {
		assessment.Threshold -= peaceEmergencyDiscount
	}
	assessment.Eligible = assessment.WarTurns >= peaceMinimumWarTurns || assessment.Emergency
	// Başkent tehdidi veya askerî çöküş gerçek acil durumdur; bunun dışındaki
	// durumda düşmanın elindeki core bölgesi için barış kapısı kapalı kalır.
	if assessment.UnresolvedCoreClaims > 0 && !assessment.Emergency {
		assessment.Eligible = false
	}
	if ledger.LastPeaceOfferTurn > 0 && gs.Turn-ledger.LastPeaceOfferTurn < peaceOfferCooldown {
		assessment.Eligible = false
	}
	assessment.Stalemate = isWarStalemate(gs, actor, opponent, ledger)
	return assessment
}

// territorialClaimPressure actor'un opponent'tan almak istediği, hâlen
// opponent elinde bulunan bölgeleri toplar. Bölgesel hak iddiasının tek statik
// kaynağı faction üzerinde materialize edilen TerritorialClaims listesidir;
// hedef devlet listeleri claim yerine kullanılmaz.
func territorialClaimPressure(gs *state.GameState, actor, opponent faction.FactionID, ledger *state.WarLedger) (coreCount, value, count int) {
	if gs == nil || actor == "" || opponent == "" {
		return 0, 0, 0
	}
	claims := make(map[string]faction.TerritorialClaim, 8)
	add := func(regionID string, claimValue int, core bool) {
		if regionID == "" {
			return
		}
		region := gs.Regions[world.RegionID(regionID)]
		if region == nil || region.IsSea || region.OwnerID != string(opponent) {
			return
		}
		if claimValue < 1 {
			claimValue = 1
		}
		current, exists := claims[regionID]
		if !exists || claimValue > current.Value || (core && !current.Core) {
			if claimValue < current.Value {
				claimValue = current.Value
			}
			claims[regionID] = faction.TerritorialClaim{
				RegionID: regionID,
				Value:    claimValue,
				Core:     core || current.Core,
			}
		}
	}

	if actorFaction := gs.Factions[actor]; actorFaction != nil {
		if actorFaction.CapitalSettlementID != "" {
			if capitalRegion, _, _, ok := gs.FindSettlementByID(actorFaction.CapitalSettlementID); ok && capitalRegion != nil {
				add(string(capitalRegion.ID), 100, true)
			}
		}
		for _, claim := range actorFaction.TerritorialClaims {
			add(claim.RegionID, claim.Value, claim.Core)
		}
	}
	if plan := gs.AIPlans[actor]; plan != nil && plan.Kind == state.AIObjectiveExpand && plan.TargetFactionID == opponent {
		for index, regionID := range plan.TargetRegionIDs {
			add(string(regionID), max(30, 50-index*5), false)
		}
	}
	if ledger != nil {
		add(string(ledger.TargetRegionID), 55, false)
	}

	for _, claim := range claims {
		count++
		value += claim.Value
		if claim.Core {
			coreCount++
		}
	}
	return coreCount, min(100, value), count
}

func warExhaustionFor(gs *state.GameState, actor, opponent faction.FactionID, ledger *state.WarLedger, warTurns, ownLosses, ownRegionsLost int) int {
	if gs == nil || ledger == nil {
		return 0
	}
	exhaustion := min(35, max(0, warTurns-2)*2)
	exhaustion += min(25, max(0, ownLosses)*2)
	exhaustion += min(20, max(0, ownRegionsLost)*5)
	if activeWarCount(gs, actor) > 1 {
		exhaustion += min(12, (activeWarCount(gs, actor)-1)*6)
	}
	if capitalUnderWarThreat(gs, actor, opponent) {
		exhaustion += 8
	}
	return min(100, exhaustion)
}

func goldPressureFor(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || gs.Factions[fid] == nil || gs.Factions[fid].Gold >= 80 {
		return 0
	}
	return min(20, 8+(80-gs.Factions[fid].Gold)/10)
}

func grainPressureFor(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || gs.Factions[fid] == nil {
		return 0
	}
	if status, ok := gs.GrainEconomy[fid]; ok {
		switch status.SupplyLevel {
		case state.GrainSupplyFamine:
			return 20
		case state.GrainSupplyCritical:
			return 14
		case state.GrainSupplyWarning:
			return 6
		}
		return 0
	}
	if gs.Factions[fid].Grain < 40 {
		return 8
	}
	return 0
}

func satisfactionPressureFor(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil {
		return 0
	}
	totalDeficit := 0
	regions := 0
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) || region.Satisfaction <= 0 {
			continue
		}
		regions++
		totalDeficit += max(0, 50-region.Satisfaction)
	}
	if regions == 0 {
		return 0
	}
	return min(20, totalDeficit/regions)
}

func relationshipPressureFor(gs *state.GameState, actor, opponent faction.FactionID) int {
	rel := Relation(gs, actor, opponent)
	if rel == nil || rel.Score >= 0 {
		return 0
	}
	return min(20, -rel.Score/5)
}

// futureLossPressureFor, barış reddedilirse actor'un kısa vadede asker ve
// bölge kaybetme riskini ölçer. İlişki skoru geçmiş duyguyu, bu değer ise
// mevcut savaş gücü ve cephe durumundan çıkan ileriye dönük maliyeti temsil
// eder. Kabul kararı hedefin perspektifinden hesaplandığı için actor burada
// barış teklifini değerlendiren devlettir.
func futureLossPressureFor(gs *state.GameState, actor, opponent faction.FactionID) int {
	if gs == nil || actor == "" || opponent == "" || actor == opponent {
		return 0
	}
	actorPower := MilitaryPower(gs, actor)
	opponentPower := MilitaryPower(gs, opponent)
	if opponentPower <= 0 {
		return 0
	}

	pressure := 0
	switch {
	case actorPower == 0:
		pressure += 28
	case opponentPower*100 >= actorPower*300:
		pressure += 24
	case opponentPower*100 >= actorPower*200:
		pressure += 18
	case opponentPower*100 >= actorPower*150:
		pressure += 12
	case opponentPower*100 >= actorPower*120:
		pressure += 6
	}

	actorFrontier := frontierArmyCount(gs, actor, opponent)
	opponentFrontier := frontierArmyCount(gs, opponent, actor)
	if opponentFrontier > actorFrontier {
		pressure += min(10, (opponentFrontier-actorFrontier)*4)
	}

	actorRegions := landRegionCount(gs, actor)
	if actorRegions > 0 && actorRegions <= 4 && opponentPower > actorPower {
		pressure += 8
	}
	return min(40, pressure)
}

func isWarStalemate(gs *state.GameState, actor, opponent faction.FactionID, ledger *state.WarLedger) bool {
	if gs == nil || ledger == nil {
		return false
	}
	warTurns := gs.Turn - ledger.StartedTurn
	if warTurns < 4 {
		return false
	}
	lastActionTurn := ledger.LastBattleTurn
	if lastActionTurn == 0 {
		lastActionTurn = ledger.StartedTurn
	}
	if gs.Turn-lastActionTurn < 3 {
		return false
	}
	for regionID, siege := range gs.Sieges {
		if siege == nil {
			continue
		}
		attacker := gs.Armies[siege.AttackerArmyID]
		target := gs.Regions[regionID]
		if attacker == nil || target == nil {
			continue
		}
		if (attacker.OwnerID == string(actor) && target.OwnerID == string(opponent)) ||
			(attacker.OwnerID == string(opponent) && target.OwnerID == string(actor)) {
			return false
		}
	}
	return true
}

func warCasualtiesFor(ledger *state.WarLedger, actor faction.FactionID) (int, int) {
	if ledger == nil {
		return 0, 0
	}
	if actor == ledger.FactionA {
		return ledger.CasualtiesA, ledger.CasualtiesB
	}
	return ledger.CasualtiesB, ledger.CasualtiesA
}

func warCapturesFor(ledger *state.WarLedger, actor faction.FactionID) (int, int) {
	if ledger == nil {
		return 0, 0
	}
	if actor == ledger.FactionA {
		return ledger.RegionsCapturedA, ledger.RegionsCapturedB
	}
	return ledger.RegionsCapturedB, ledger.RegionsCapturedA
}

func warInitialRegionsFor(ledger *state.WarLedger, actor faction.FactionID) (int, int) {
	if ledger == nil {
		return 0, 0
	}
	if actor == ledger.FactionA {
		return ledger.InitialRegionsA, ledger.InitialRegionsB
	}
	return ledger.InitialRegionsB, ledger.InitialRegionsA
}

// warScoreFor savaş sonucunu actor perspektifinden türetir. Pozitif değer actor'un
// avantajını, negatif değer opponent baskısını gösterir. Bu değer save'e yazılmaz;
// ledger, mevcut sahiplik ve plan hedeflerinden her değerlendirmede yeniden hesaplanır.
func warScoreFor(gs *state.GameState, actor, opponent faction.FactionID, ledger *state.WarLedger) (int, int, int) {
	if gs == nil || ledger == nil || actor == "" || opponent == "" {
		return 0, 0, 0
	}
	actorLosses, opponentLosses := warCasualtiesFor(ledger, actor)
	actorCaptured, opponentCaptured := warCapturesFor(ledger, actor)
	score := clamp((actorCaptured-opponentCaptured)*20, -40, 40)
	score += clamp((opponentLosses-actorLosses)*2, -30, 30)

	initialActor, initialOpponent := warInitialRegionsFor(ledger, actor)
	currentActor := landRegionCount(gs, actor)
	currentOpponent := landRegionCount(gs, opponent)
	score += clamp((initialActor-currentActor)*15, -30, 30)
	score += clamp((initialOpponent-currentOpponent)*15, -30, 30)

	if capitalUnderWarThreat(gs, actor, opponent) {
		score -= 25
	}
	if capitalUnderWarThreat(gs, opponent, actor) {
		score += 25
	}

	held, total := warObjectiveProgress(gs, actor, opponent)
	if total > 0 {
		score += held * 12
	}
	return clamp(score, -100, 100), held, total
}

func warObjectiveProgress(gs *state.GameState, actor, opponent faction.FactionID) (held, total int) {
	if gs == nil {
		return 0, 0
	}
	plan := gs.AIPlans[actor]
	if plan == nil || plan.Kind != state.AIObjectiveExpand || plan.TargetFactionID != opponent {
		return 0, 0
	}
	for _, regionID := range plan.TargetRegionIDs {
		region := gs.Regions[regionID]
		if region == nil {
			continue
		}
		total++
		if region.OwnerID == string(actor) {
			held++
		}
	}
	return held, total
}

func activeWarCount(gs *state.GameState, actor faction.FactionID) int {
	count := 0
	for _, rel := range gs.Relations {
		if rel == nil || rel.Stance != faction.StanceWar {
			continue
		}
		if rel.FactionA == actor || rel.FactionB == actor {
			count++
		}
	}
	return count
}

func capitalUnderWarThreat(gs *state.GameState, actor, opponent faction.FactionID) bool {
	if gs == nil || actor == "" || opponent == "" {
		return false
	}
	f := gs.Factions[actor]
	if f == nil || f.CapitalSettlementID == "" {
		return landRegionCount(gs, actor) <= 1 && MilitaryPower(gs, actor) == 0
	}
	capitalRegion, _, _, ok := gs.FindSettlementByID(f.CapitalSettlementID)
	if !ok || capitalRegion == nil || capitalRegion.OwnerID != string(actor) {
		return true
	}
	if siege := gs.SiegeAt(capitalRegion.ID); siege != nil && siege.AttackerFactionID == string(opponent) {
		return true
	}
	for _, armyRef := range gs.Armies {
		if armyRef == nil || armyRef.IsNaval || armyRef.OwnerID != string(opponent) {
			continue
		}
		if armyRef.RegionID == capitalRegion.ID {
			return true
		}
		for _, neighborID := range capitalRegion.Neighbors {
			if armyRef.RegionID == neighborID {
				return true
			}
		}
	}
	return false
}

func warObjectiveCompleted(gs *state.GameState, actor, opponent faction.FactionID) bool {
	if gs == nil {
		return false
	}
	plan := gs.AIPlans[actor]
	if plan == nil || plan.Kind != state.AIObjectiveExpand || plan.TargetFactionID != opponent || len(plan.TargetRegionIDs) == 0 {
		return false
	}
	for _, regionID := range plan.TargetRegionIDs {
		region := gs.Regions[regionID]
		if region != nil && region.OwnerID == string(opponent) {
			return false
		}
	}
	return true
}
