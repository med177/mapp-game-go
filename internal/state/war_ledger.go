package state

import (
	"math"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

// PostPeaceTruceTurns kabul edilen barıştan sonra tarafların yeniden savaş
// ilan edemeyeceği ateşkes süresidir.
const PostPeaceTruceTurns = 5

// RecordTruce barıştan sonra aynı tarafların hemen yeniden savaşa girmesini
// önleyen save-backed ateşkes bitişini kaydeder.
func (s *GameState) RecordTruce(a, b faction.FactionID) {
	if s == nil || a == "" || b == "" || a == b {
		return
	}
	if s.RecentTruces == nil {
		s.RecentTruces = make(map[string]int)
	}
	s.RecentTruces[faction.RelationKey(a, b)] = s.Turn + PostPeaceTruceTurns
}

// TruceRemaining kalan ateşkes turunu döner; sıfır, savaş ilanının serbest
// olduğunu belirtir. Süresi dolan kayıtlar okunurken etkisiz kabul edilir.
func (s *GameState) TruceRemaining(a, b faction.FactionID) int {
	if s == nil || s.RecentTruces == nil {
		return 0
	}
	return max(0, s.RecentTruces[faction.RelationKey(a, b)]-s.Turn)
}

// RecordRegionAcquisition kısa vadeli genişleme takibine bir bölge kazanımı
// ekler. Savaş dışı barış devri de bu kaydı kullanabilir; böylece koalisyon
// mantığı yalnızca savaş ledger'larına bağlı kalmaz.
func (s *GameState) RecordRegionAcquisition(conqueror, previousOwner faction.FactionID) {
	s.RecordRegionAcquisitions(conqueror, previousOwner, 1)
}

// RecordRegionAcquisitions, tek bir siyasi kazanımın birden fazla kara
// bölgesine karşılık geldiği durumlarda genişleme geçmişini bölge sayısıyla
// günceller. Vassallıkta bölge sahibi değişmediği için bu kayıt, vassalın
// topraklarını doğrudan fetih gibi ledger'a eklemek için kullanılır.
func (s *GameState) RecordRegionAcquisitions(conqueror, previousOwner faction.FactionID, count int) {
	if s == nil || conqueror == "" || conqueror == previousOwner || count <= 0 {
		return
	}
	if s.RecentFactionExpansion == nil {
		s.RecentFactionExpansion = make(map[faction.FactionID]FactionExpansionRecord)
	}
	turn := s.Turn
	if turn <= 0 {
		turn = 1
	}
	lastTurns := s.AggressiveExpansionWindowTurns()
	record := s.RecentFactionExpansion[conqueror]
	gains := expansionTurnGains(record, turn)
	for gainTurn := range gains {
		if turn-gainTurn >= lastTurns {
			delete(gains, gainTurn)
		}
	}
	gains[turn] += count
	record.TurnGains = gains
	record.RegionsGained = 0
	record.WindowStartTurn = turn
	for gainTurn, count := range gains {
		record.RegionsGained += count
		if gainTurn < record.WindowStartTurn {
			record.WindowStartTurn = gainTurn
		}
	}
	s.RecentFactionExpansion[conqueror] = record
}

// RecentRegionGain kısa genişleme penceresindeki kademeli baskıyı yuvarlanmış
// bölge karşılığı olarak döner. Eski veya süresi dolmuş kazanımlar etkisizdir.
func (s *GameState) RecentRegionGain(fid faction.FactionID) int {
	if s == nil || fid == "" || s.RecentFactionExpansion == nil {
		return 0
	}
	record := s.RecentFactionExpansion[fid]
	if record.RegionsGained <= 0 {
		return 0
	}
	return int(math.Round(s.recentRegionGainValue(record)))
}

// AggressiveExpansionWindowTurns aktif senaryonun aşırı genişleme baskısının
// kaç tur sürdüğünü döner. Doğrudan oluşturulan eski test/state nesneleri için
// sabit varsayılan korunur; yüklenmiş senaryolar bu değeri doldurur.
func (s *GameState) AggressiveExpansionWindowTurns() int {
	if s != nil && s.AggressiveExpansionLastTurns > 0 {
		return s.AggressiveExpansionLastTurns
	}
	return RecentFactionExpansionWindowTurns
}

// expansionTurnGains, eski save formatını yeni tur bazlı geçmiş biçimine
// dönüştürür. Map'in sahibi çağıran olduğu için burada yerinde güncellenir.
func expansionTurnGains(record FactionExpansionRecord, fallbackTurn int) map[int]int {
	if record.TurnGains != nil {
		return record.TurnGains
	}
	if record.RegionsGained <= 0 {
		return make(map[int]int)
	}
	turn := record.WindowStartTurn
	if turn <= 0 {
		turn = fallbackTurn
	}
	return map[int]int{turn: record.RegionsGained}
}

// recentRegionGainValue, her kazanımın etkisini kalan süre oranının karesiyle
// azaltır. Böylece baskı tur tur düşer ve sürenin sonuna doğru daha hızlı söner.
func (s *GameState) recentRegionGainValue(record FactionExpansionRecord) float64 {
	if s == nil || record.RegionsGained <= 0 {
		return 0
	}
	lastTurns := s.AggressiveExpansionWindowTurns()
	if lastTurns <= 0 {
		return 0
	}
	gains := expansionTurnGains(record, s.Turn)
	value := 0.0
	for gainTurn, count := range gains {
		if count <= 0 {
			continue
		}
		age := s.Turn - gainTurn
		if age < 0 || age >= lastTurns {
			continue
		}
		remainingRatio := float64(lastTurns-age) / float64(lastTurns)
		value += float64(count) * remainingRatio * remainingRatio
	}
	return value
}

// OverextensionScore son kısa genişleme penceresini oyuncuya ve AI'ye ortak
// bir 0-100 risk değeri olarak sunar. Mutlak kazanım ve mevcut devlete göre
// büyüme oranından yüksek olanı kullanır; böylece dört bölge kazanan büyük bir
// devlet ile iki bölge kazanarak iki katına çıkan küçük devlet aynı baskı
// sinyalini paylaşabilir. Skor türetilmiştir; geçmiş kazanımlar save'de tutulur,
// skor için ayrıca bir alan gerekmez.
func (s *GameState) OverextensionScore(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	record := s.RecentFactionExpansion[fid]
	gained := s.recentRegionGainValue(record)
	if gained <= 0 {
		return 0
	}
	owned := len(s.LandRegionsOwnedBy(fid))
	if owned <= 0 {
		return 0
	}
	absScore := int(math.Ceil(gained * 20))
	relativeScore := int(math.Ceil(gained * 100 / float64(owned) * 2))
	score := absScore
	if relativeScore > score {
		score = relativeScore
	}
	if score > 100 {
		return 100
	}
	return score
}

// WarLedger aktif bir savaşın barış değerlendirmesinde kullanılan kalıcı
// başlangıç durumunu ve iki taraflı sonuçlarını tutar. FactionA/FactionB,
// RelationKey ile aynı alfabetik sıradadır.
type WarLedger struct {
	FactionA faction.FactionID `json:"faction_a"`
	FactionB faction.FactionID `json:"faction_b"`
	// DeclarerFactionID ve DefenderFactionID savaş ilanının yönünü korur.
	// FactionA/FactionB ise kayıp ve bölge sayaçları için RelationKey ile
	// uyumlu alfabetik sıralamayı sürdürür.
	DeclarerFactionID  faction.FactionID `json:"declarer_faction_id,omitempty"`
	DefenderFactionID  faction.FactionID `json:"defender_faction_id,omitempty"`
	StartedTurn        int               `json:"started_turn"`
	InitialRegionsA    int               `json:"initial_regions_a"`
	InitialRegionsB    int               `json:"initial_regions_b"`
	CasualtiesA        int               `json:"casualties_a,omitempty"`
	CasualtiesB        int               `json:"casualties_b,omitempty"`
	CasualtiesArmyA    int               `json:"casualties_army_a,omitempty"`
	CasualtiesArmyB    int               `json:"casualties_army_b,omitempty"`
	CasualtiesFleetA   int               `json:"casualties_fleet_a,omitempty"`
	CasualtiesFleetB   int               `json:"casualties_fleet_b,omitempty"`
	RegionsCapturedA   int               `json:"regions_captured_a,omitempty"`
	RegionsCapturedB   int               `json:"regions_captured_b,omitempty"`
	LastBattleTurn     int               `json:"last_battle_turn,omitempty"`
	LastPeaceOfferTurn int               `json:"last_peace_offer_turn,omitempty"`
	TargetRegionID     world.RegionID    `json:"target_region_id,omitempty"`
	TargetLockedTurn   int               `json:"target_locked_turn,omitempty"`
	// RecklessDeclaration, AI'nin nadir riskli savaş zarını kullandığını taşır.
	// Bu savaşlarda erken barış baskısı bilinçli olarak biraz daha yüksektir.
	RecklessDeclaration bool `json:"reckless_declaration,omitempty"`
}

// BeginWarLedger savaş başlangıcını yalnızca ilk geçişte kaydeder.
func (s *GameState) BeginWarLedger(a, b faction.FactionID) *WarLedger {
	if s == nil || a == "" || b == "" || a == b {
		return nil
	}
	if s.WarLedgers == nil {
		s.WarLedgers = make(map[string]*WarLedger)
	}
	key := faction.RelationKey(a, b)
	if existing := s.WarLedgers[key]; existing != nil {
		return existing
	}
	left, right := a, b
	if right < left {
		left, right = right, left
	}
	ledger := &WarLedger{
		FactionA:          left,
		FactionB:          right,
		DeclarerFactionID: a,
		DefenderFactionID: b,
		StartedTurn:       s.Turn,
		InitialRegionsA:   len(s.LandRegionsOwnedBy(left)),
		InitialRegionsB:   len(s.LandRegionsOwnedBy(right)),
	}
	s.WarLedgers[key] = ledger
	return ledger
}

// EndWarLedger artık aktif olmayan savaşın sayacını kaldırır ve ilgili AI
// planını bir sonraki turda yeniden değerlendirmeye zorlar.
func (s *GameState) EndWarLedger(a, b faction.FactionID) {
	if s == nil || a == "" || b == "" || a == b {
		return
	}
	delete(s.WarLedgers, faction.RelationKey(a, b))
	for _, pair := range [][2]faction.FactionID{{a, b}, {b, a}} {
		plan := s.AIPlans[pair[0]]
		if plan == nil || plan.TargetFactionID != pair[1] {
			continue
		}
		plan.ReassessTurn = s.Turn
		plan.RallyRegionID = ""
		plan.RallyDeadlineTurn = 0
	}
}

// SyncWarLedgers kayıt göçü ve doğrudan stance düzenleyen eski kod yolları için
// aktif savaşlarla ledger haritasını uzlaştırır.
func (s *GameState) SyncWarLedgers() {
	if s == nil {
		return
	}
	if s.WarLedgers == nil {
		s.WarLedgers = make(map[string]*WarLedger)
	}
	active := make(map[string]struct{})
	for _, rel := range s.Relations {
		if rel == nil || rel.Stance != faction.StanceWar || rel.FactionA == "" || rel.FactionB == "" {
			continue
		}
		active[faction.RelationKey(rel.FactionA, rel.FactionB)] = struct{}{}
		s.BeginWarLedger(rel.FactionA, rel.FactionB)
	}
	for key := range s.WarLedgers {
		if _, ok := active[key]; !ok {
			delete(s.WarLedgers, key)
		}
	}
}

func (s *GameState) WarLedgerFor(a, b faction.FactionID) *WarLedger {
	if s == nil {
		return nil
	}
	return s.WarLedgers[faction.RelationKey(a, b)]
}

// RecordWarCasualties muharebedeki tamamen kaybedilen birlik sayılarını iki
// savaşan tarafa yazar. Aktif savaş ilişkisi yoksa kayıt üretmez.
func (s *GameState) RecordWarCasualties(attacker, defender faction.FactionID, attackerLost, defenderLost int) {
	s.recordWarCasualties(attacker, defender, attackerLost, defenderLost, false, false, true)
}

// RecordWarCasualtiesByType muharebe kayıplarını toplam sayaçların yanında
// ordunun kara kuvveti mi yoksa filo mu olduğuna göre ayrı sayaçlara yazar.
func (s *GameState) RecordWarCasualtiesByType(attacker, defender faction.FactionID, attackerLost, defenderLost int, attackerNaval, defenderNaval bool) {
	s.recordWarCasualties(attacker, defender, attackerLost, defenderLost, attackerNaval, defenderNaval, true)
}

// RecordWarAttritionCasualties kuşatma baskısı gibi muharebe dışı kayıpları
// LastBattleTurn değerini değiştirmeden yazar.
func (s *GameState) RecordWarAttritionCasualties(attacker, defender faction.FactionID, attackerLost, defenderLost int) {
	s.recordWarCasualties(attacker, defender, attackerLost, defenderLost, false, false, false)
}

func (s *GameState) recordWarCasualties(attacker, defender faction.FactionID, attackerLost, defenderLost int, attackerNaval, defenderNaval, markBattle bool) {
	if s == nil || attacker == "" || defender == "" || attacker == defender {
		return
	}
	rel := s.Relations[faction.RelationKey(attacker, defender)]
	if rel == nil || rel.Stance != faction.StanceWar {
		return
	}
	ledger := s.BeginWarLedger(attacker, defender)
	if ledger == nil {
		return
	}
	if attackerLost < 0 {
		attackerLost = 0
	}
	if defenderLost < 0 {
		defenderLost = 0
	}
	if attacker == ledger.FactionA {
		ledger.CasualtiesA += attackerLost
		ledger.CasualtiesB += defenderLost
		if attackerNaval {
			ledger.CasualtiesFleetA += attackerLost
		} else {
			ledger.CasualtiesArmyA += attackerLost
		}
		if defenderNaval {
			ledger.CasualtiesFleetB += defenderLost
		} else {
			ledger.CasualtiesArmyB += defenderLost
		}
	} else {
		ledger.CasualtiesB += attackerLost
		ledger.CasualtiesA += defenderLost
		if attackerNaval {
			ledger.CasualtiesFleetB += attackerLost
		} else {
			ledger.CasualtiesArmyB += attackerLost
		}
		if defenderNaval {
			ledger.CasualtiesFleetA += defenderLost
		} else {
			ledger.CasualtiesArmyA += defenderLost
		}
	}
	if markBattle {
		ledger.LastBattleTurn = s.Turn
	}
}

// RecordWarRegionCapture yalnızca iki aktif savaş tarafı arasındaki sahiplik
// değişimini sayar.
func (s *GameState) RecordWarRegionCapture(conqueror, previousOwner faction.FactionID) {
	if s == nil || conqueror == "" || previousOwner == "" || conqueror == previousOwner {
		return
	}
	s.RecordRegionAcquisition(conqueror, previousOwner)
	rel := s.Relations[faction.RelationKey(conqueror, previousOwner)]
	if rel == nil || rel.Stance != faction.StanceWar {
		return
	}
	ledger := s.BeginWarLedger(conqueror, previousOwner)
	if ledger == nil {
		return
	}
	if conqueror == ledger.FactionA {
		ledger.RegionsCapturedA++
	} else {
		ledger.RegionsCapturedB++
	}
}

// MarkPeaceOffer savaş başına kısa teklif tekrar aralığını kalıcılaştırır.
func (s *GameState) MarkPeaceOffer(a, b faction.FactionID) {
	if ledger := s.WarLedgerFor(a, b); ledger != nil {
		ledger.LastPeaceOfferTurn = s.Turn
	}
}
