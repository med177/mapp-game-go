package state

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

// PostPeaceTruceTurns kabul edilen barıştan sonra tarafların yeniden savaş
// ilan edemeyeceği ateşkes süresidir.
const PostPeaceTruceTurns = 6

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
	RegionsCapturedA   int               `json:"regions_captured_a,omitempty"`
	RegionsCapturedB   int               `json:"regions_captured_b,omitempty"`
	LastBattleTurn     int               `json:"last_battle_turn,omitempty"`
	LastPeaceOfferTurn int               `json:"last_peace_offer_turn,omitempty"`
	TargetRegionID     world.RegionID    `json:"target_region_id,omitempty"`
	TargetLockedTurn   int               `json:"target_locked_turn,omitempty"`
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
	s.recordWarCasualties(attacker, defender, attackerLost, defenderLost, true)
}

// RecordWarAttritionCasualties kuşatma baskısı gibi muharebe dışı kayıpları
// LastBattleTurn değerini değiştirmeden yazar.
func (s *GameState) RecordWarAttritionCasualties(attacker, defender faction.FactionID, attackerLost, defenderLost int) {
	s.recordWarCasualties(attacker, defender, attackerLost, defenderLost, false)
}

func (s *GameState) recordWarCasualties(attacker, defender faction.FactionID, attackerLost, defenderLost int, markBattle bool) {
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
	} else {
		ledger.CasualtiesB += attackerLost
		ledger.CasualtiesA += defenderLost
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
