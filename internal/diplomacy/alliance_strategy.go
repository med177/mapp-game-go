package diplomacy

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

// AllianceRelationThreshold, AI'nin teklif üretirken kullandığı senaryo bazlı
// ilişki eşiğini dışarıya açar. Teklif üretimi ile actionBlockReason aynı kapıyı
// kullanmalıdır; aksi halde AI kota harcayıp kesin bloke olacak teklif üretir.
func AllianceRelationThreshold(gs *state.GameState) int {
	return allianceRelationThresholdFor(gs)
}

// allianceWarConflictBetween, iki devlet arasındaki olası ittifakın mevcut
// müttefiklerden birini doğrudan karşı cepheye düşürüp düşürmediğini kontrol eder.
// Kontrol iki yönlüdür: teklif sahibi kendi müttefikinin düşmanına yaklaşamaz,
// hedef de kendi müttefikinin savaşta olduğu devleti ittifakla yanına alamaz.
func allianceWarConflictBetween(gs *state.GameState, a, b faction.FactionID) (faction.FactionID, bool) {
	if gs == nil || a == "" || b == "" || a == b {
		return "", false
	}
	aRoot := realmRoot(gs, a)
	if aRoot == "" {
		aRoot = a
	}
	bRoot := realmRoot(gs, b)
	if bRoot == "" {
		bRoot = b
	}
	if aRoot == bRoot {
		return "", false
	}

	// İttifak ve savaş kayıtları realm kökleri üzerinden tutulabildiği için
	// kontrolü yalnız ham faction ID çiftlerini tarayarak yapma. Bu yol hem
	// oyuncu aksiyonunu hem de bekleyen AI teklifinin kabulünü aynı kurala bağlar.
	for _, allyID := range directExternalAlliesOf(gs, aRoot) {
		allyRoot := realmRoot(gs, allyID)
		if allyRoot == "" {
			allyRoot = allyID
		}
		if allyRoot != bRoot && IsWar(gs, allyRoot, bRoot) {
			return allyRoot, true
		}
	}
	for _, allyID := range directExternalAlliesOf(gs, bRoot) {
		allyRoot := realmRoot(gs, allyID)
		if allyRoot == "" {
			allyRoot = allyID
		}
		if allyRoot != aRoot && IsWar(gs, allyRoot, aRoot) {
			return allyRoot, true
		}
	}
	return "", false
}

const (
	// attackedAllyRelationPenalty tarafsız kalan doğrudan müttefike yönelik
	// saldırının diplomatik itibar maliyetidir.
	attackedAllyRelationPenalty = 25

	// attackedExistingAllyPenalty saldıranın kendi müttefikinin müttefikine
	// savaş açmasıyla oluşan doğrudan ittifak ihlalinin daha ağır maliyetidir.
	attackedExistingAllyPenalty = 35
)

type AllianceWarRelationPenalty struct {
	FactionID       faction.FactionID
	RelationPenalty int
	AllianceBroken  bool
}

// applyAllianceWarRelationPenalties savaş ilan edilmeden önce hedefin doğrudan
// dış müttefiklerini snapshot olarak değerlendirir. Sadece savaş ilan eden
// devlet cezalandırılır; savaş çağrısını kabul eden müttefikler zaten
// setWarBetweenCoalitions tarafından doğrudan savaş ilişkisine geçirilir.
//
// Bu fonksiyon setWarBetweenCoalitions içine konulmaz: koalisyonun her üyesi
// için çağrıldığında aynı ilişki cezası birden fazla kez uygulanabilir.
func applyAllianceWarRelationPenalties(gs *state.GameState, attacker, target faction.FactionID) []AllianceWarRelationPenalty {
	if gs == nil || attacker == "" || target == "" || attacker == target {
		return nil
	}
	attackerRoot := realmRoot(gs, attacker)
	if attackerRoot == "" {
		attackerRoot = attacker
	}
	targetRoot := realmRoot(gs, target)
	if targetRoot == "" {
		targetRoot = target
	}

	allies := directExternalAlliesOf(gs, targetRoot)
	penalties := make([]AllianceWarRelationPenalty, 0, len(allies))
	for _, allyID := range allies {
		allyRoot := realmRoot(gs, allyID)
		if allyRoot == "" {
			allyRoot = allyID
		}
		if allyRoot == attackerRoot || allyRoot == targetRoot || sameRealm(gs, attackerRoot, allyRoot) {
			continue
		}

		rel := EnsureRelation(gs, attackerRoot, allyRoot)
		if rel.Stance == faction.StanceWar {
			// Mevcut savaş -80 ile zaten temsil ediliyor; ikinci bir olay cezası
			// ilişkiyi aynı savaş başlangıcında tekrar değiştirmemeli.
			continue
		}

		penalty := attackedAllyRelationPenalty
		allianceBroken := rel.Stance == faction.StanceAllied
		if allianceBroken {
			penalty = attackedExistingAllyPenalty
			// İttifak bozulur, mevcut ticaret anlaşması varsa korunur. Bu,
			// savaş çağrısını reddeden müttefiklerin mevcut davranışıyla aynıdır.
			if HasTradeRouteBetween(gs, attackerRoot, allyRoot) {
				rel.Stance = faction.StanceTrade
			} else {
				rel.Stance = faction.StancePeace
			}
		}
		rel.Score = clamp(rel.Score-penalty, -100, 100)
		penalties = append(penalties, AllianceWarRelationPenalty{
			FactionID:       allyRoot,
			RelationPenalty: penalty,
			AllianceBroken:  allianceBroken,
		})
	}
	return penalties
}

const strategicAllianceAcceptanceFloor = 12

// StrategicAllianceAssessment 1300 senaryosunda actor'un target ile ittifaktan
// elde edeceği stratejik değeri bileşenlerine ayırır.
type StrategicAllianceAssessment struct {
	Score                   int
	ThreatValue             int
	BufferValue             int
	FrontSupportValue       int
	TradeValue              int
	PartnerSupportValue     int
	ExpansionTensionPenalty int
	ActiveObjectiveConflict bool
	CommonEnemy             bool
	SharedMajorThreat       bool
}

// AssessStrategicAlliance değerlendirmeyi actor perspektifinden yapar. Teklif
// sahibinin girişim kararı ve hedef AI'nin kabul kararı aynı bileşenleri ters
// perspektiflerden okuyabilir.
func AssessStrategicAlliance(gs *state.GameState, actor, target faction.FactionID) StrategicAllianceAssessment {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || actor == "" || target == "" || actor == target {
		return StrategicAllianceAssessment{}
	}
	return assessStrategicAlliance(gs, actor, target, HasCommonEnemy(gs, actor, target), HasSharedMajorThreat(gs, actor, target))
}

func assessStrategicAlliance(gs *state.GameState, actor, target faction.FactionID, commonEnemy, sharedMajorThreat bool) StrategicAllianceAssessment {
	tradeAccess := HasTradeRouteBetween(gs, actor, target) || CanEstablishTradeRoute(gs, actor, target)
	return assessStrategicAllianceWithTrade(gs, actor, target, commonEnemy, sharedMajorThreat, tradeAccess)
}

func assessStrategicAllianceWithTrade(gs *state.GameState, actor, target faction.FactionID, commonEnemy, sharedMajorThreat bool, tradeAccess bool) StrategicAllianceAssessment {
	assessment := StrategicAllianceAssessment{}
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || actor == "" || target == "" || actor == target {
		return assessment
	}
	assessment.ActiveObjectiveConflict = activeAllianceObjectiveConflict(gs, actor, target)
	assessment.CommonEnemy = commonEnemy
	assessment.SharedMajorThreat = sharedMajorThreat
	if assessment.CommonEnemy {
		assessment.ThreatValue += 20
	}
	if assessment.SharedMajorThreat {
		assessment.ThreatValue += 18
	}

	threats := allianceThreatsAgainst(gs, actor, target)
	assessment.BufferValue = allianceBufferValue(gs, target, threats)
	assessment.FrontSupportValue = allianceFrontSupportValue(gs, target, threats)
	assessment.TradeValue = allianceTradeValue(gs, actor, target, tradeAccess)
	assessment.PartnerSupportValue = alliancePartnerSupportValue(gs, target)
	if staticAllianceExpansionTension(gs, actor, target) {
		assessment.ExpansionTensionPenalty = 18
	}
	assessment.Score = assessment.ThreatValue +
		assessment.BufferValue +
		assessment.FrontSupportValue +
		assessment.TradeValue +
		assessment.PartnerSupportValue -
		assessment.ExpansionTensionPenalty
	return assessment
}

func activeAllianceObjectiveConflict(gs *state.GameState, a, b faction.FactionID) bool {
	if gs == nil {
		return false
	}
	if plan := gs.AIPlans[a]; plan != nil && plan.TargetFactionID == b {
		return true
	}
	if plan := gs.AIPlans[b]; plan != nil && plan.TargetFactionID == a {
		return true
	}
	return false
}

func staticAllianceExpansionTension(gs *state.GameState, a, b faction.FactionID) bool {
	return factionTargets(gs, a, b) || factionTargets(gs, b, a)
}

func factionTargets(gs *state.GameState, actor, target faction.FactionID) bool {
	if gs == nil {
		return false
	}
	f := gs.Factions[actor]
	if f == nil {
		return false
	}
	for _, targetID := range f.AIExpansionTargets {
		if targetID == target {
			return true
		}
	}
	return false
}

func allianceThreatsAgainst(gs *state.GameState, actor, candidate faction.FactionID) map[faction.FactionID]struct{} {
	threats := make(map[faction.FactionID]struct{})
	potentialThreats := make(map[faction.FactionID]struct{})
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(candidate) {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil || neighbor.IsSea || neighbor.OwnerID == "" {
				continue
			}
			potentialThreats[faction.FactionID(neighbor.OwnerID)] = struct{}{}
		}
	}
	for otherID := range potentialThreats {
		other := gs.Factions[otherID]
		if otherID == actor || otherID == candidate || other == nil || other.IsEliminated {
			continue
		}
		if IsWar(gs, actor, otherID) || isMajorThreatTo(gs, otherID, actor) {
			threats[otherID] = struct{}{}
		}
	}
	return threats
}

func allianceBufferValue(gs *state.GameState, candidate faction.FactionID, threats map[faction.FactionID]struct{}) int {
	if gs == nil || len(threats) == 0 {
		return 0
	}
	borderedThreats := 0
	for threatID := range threats {
		if sharesBorder(gs, candidate, threatID) {
			borderedThreats++
		}
	}
	if borderedThreats == 0 {
		return 0
	}
	return min(16, 7+borderedThreats*3+landRegionCount(gs, candidate)/2)
}

func allianceFrontSupportValue(gs *state.GameState, candidate faction.FactionID, threats map[faction.FactionID]struct{}) int {
	if gs == nil || len(threats) == 0 {
		return 0
	}
	frontPower := 0
	for _, armyRef := range gs.Armies {
		if armyRef == nil || armyRef.IsNaval || armyRef.OwnerID != string(candidate) {
			continue
		}
		region := gs.Regions[armyRef.RegionID]
		if region == nil {
			continue
		}
		atThreatFront := false
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil {
				continue
			}
			if _, ok := threats[faction.FactionID(neighbor.OwnerID)]; ok {
				atThreatFront = true
				break
			}
		}
		if !atThreatFront {
			continue
		}
		if gs.UnitTypes != nil {
			frontPower += armyRef.TotalStrength(gs.UnitTypes)
		} else {
			frontPower += len(armyRef.Units) * 10
		}
	}
	return min(16, frontPower/15)
}

func allianceTradeValue(gs *state.GameState, actor, target faction.FactionID, tradeAccess bool) int {
	value := 0
	if HasTradeRouteBetween(gs, actor, target) {
		value += 10
	} else if tradeAccess {
		value += 5
	}
	value += min(6, totalTradeCapacity(gs, target)/4)
	return value
}

func alliancePartnerSupportValue(gs *state.GameState, target faction.FactionID) int {
	if gs == nil {
		return 0
	}
	return min(14, MilitaryPower(gs, target)/20) + min(8, landRegionCount(gs, target)*2)
}
