package state

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

// NavalContactDecision, temas eden filonun çatışma tutumunu belirtir.
type NavalContactDecision string

const (
	NavalContactUndecided NavalContactDecision = ""
	NavalContactClash     NavalContactDecision = "clash"
	NavalContactWithdraw  NavalContactDecision = "withdraw"
	NavalContactHold      NavalContactDecision = "hold"
)

// NavalContactWithdrawMovementCost, deniz temasından geri çekilmenin bu tur
// harcadığı hareket puanıdır. Filo daha az puana sahipse kalan puan sıfırlanır.
const NavalContactWithdrawMovementCost = 2

// NavalContactTrigger, temasın aynı denize hareketle mi yoksa savaş açılmasıyla
// mı doğduğunu ayırır. Bu ayrım görev varsayılanlarını korur.
type NavalContactTrigger string

const (
	NavalContactMovement          NavalContactTrigger = "movement"
	NavalContactDirectAttack      NavalContactTrigger = "direct_attack"
	NavalContactWarOpening        NavalContactTrigger = "war_opening"
	NavalContactMissionAssignment NavalContactTrigger = "mission_assignment"
)

// NavalContact, iki düşman filonun aynı açık denizde karşılaşma kararını taşır.
// Geçici tur state'idir; temas çözülünce temizlenir ve save'e yazılmaz.
type NavalContact struct {
	AttackerArmyID       army.ArmyID
	DefenderArmyID       army.ArmyID
	SeaRegionID          world.RegionID
	AttackerFromRegionID world.RegionID
	Trigger              NavalContactTrigger
	AttackerDecision     NavalContactDecision
	DefenderDecision     NavalContactDecision
	PlayerArmyID         army.ArmyID
	Prompted             bool
	MovementConsumed     bool
}

// BeginNavalContact yeni deniz temasını başlatır. Oyuncu tarafı kararsız
// bırakılır; AI tarafı görev ve temas türüne göre varsayılan karar alır.
func (s *GameState) BeginNavalContact(attacker, defender *army.Army, seaID, attackerFrom world.RegionID, trigger NavalContactTrigger) *NavalContact {
	if s == nil || attacker == nil || defender == nil || !attacker.IsAtSea() || !defender.IsAtSea() || seaID == "" || attacker.ID == defender.ID || attacker.OwnerID == defender.OwnerID {
		return nil
	}
	if s.PendingNavalContact != nil {
		return s.PendingNavalContact
	}
	contact := &NavalContact{
		AttackerArmyID:       attacker.ID,
		DefenderArmyID:       defender.ID,
		SeaRegionID:          seaID,
		AttackerFromRegionID: attackerFrom,
		Trigger:              trigger,
		AttackerDecision:     navalContactDefaultDecision(attacker, trigger),
		DefenderDecision:     navalContactDefaultDecision(defender, trigger),
	}
	if attacker.OwnerID == string(s.PlayerFactionID) {
		contact.PlayerArmyID = attacker.ID
		contact.AttackerDecision = NavalContactUndecided
	} else if defender.OwnerID == string(s.PlayerFactionID) {
		contact.PlayerArmyID = defender.ID
		contact.DefenderDecision = NavalContactUndecided
	}
	s.PendingNavalContact = contact
	return contact
}

// NavalContactDecisionForPlayer, temas kaydındaki oyuncu tarafının kararını
// döndürür.
func (s *GameState) NavalContactDecisionForPlayer(contact *NavalContact, decision NavalContactDecision) bool {
	if s == nil || contact == nil || contact.PlayerArmyID == "" || decision == NavalContactUndecided {
		return false
	}
	if decision == NavalContactWithdraw {
		fleet := s.Armies[contact.PlayerArmyID]
		if fleet == nil || fleet.MovePoints <= 0 || s.NavalContactRetreatRegion(fleet, contact.AttackerFromRegionID) == "" {
			return false
		}
	}
	if contact.PlayerArmyID == contact.AttackerArmyID {
		contact.AttackerDecision = decision
		return true
	}
	if contact.PlayerArmyID == contact.DefenderArmyID {
		contact.DefenderDecision = decision
		return true
	}
	return false
}

func (s *GameState) ClearNavalContact() {
	if s != nil {
		s.PendingNavalContact = nil
	}
}

func (s *GameState) NavalContactBothClash(contact *NavalContact) bool {
	return contact != nil && contact.AttackerDecision == NavalContactClash && contact.DefenderDecision == NavalContactClash
}

// NavalContactWillClash, geri çekilmeyen taraflardan en az biri çatışmayı
// kabul ettiğinde muharebenin başlayacağını belirtir. Pozisyonu koru,
// çatışmadan kaçış değil savunma hazırlığıdır.
func (s *GameState) NavalContactWillClash(contact *NavalContact) bool {
	if contact == nil || contact.AttackerDecision == NavalContactWithdraw || contact.DefenderDecision == NavalContactWithdraw {
		return false
	}
	return contact.AttackerDecision == NavalContactClash || contact.DefenderDecision == NavalContactClash
}

// NavalContactRetreatRegion, geri çekilen filoya düşman filosu olmayan deniz
// komşusunu seçer. excludedRegions, düşmanın geldiği kaynak deniz gibi geri
// dönülmemesi gereken bölgeleri belirtir. Güvenli hedef yoksa boş döner.
func (s *GameState) NavalContactRetreatRegion(fleet *army.Army, excludedRegions ...world.RegionID) world.RegionID {
	if fleet == nil {
		return ""
	}
	return s.NavalContactRetreatRegionAt(fleet, fleet.RegionID, excludedRegions...)
}

// NavalContactRetreatRegionAt, henüz hedef denize girmemiş bir filo için
// temasın oluşacağı deniz bölgesinden kaçış rotasını hesaplar.
func (s *GameState) NavalContactRetreatRegionAt(fleet *army.Army, regionID world.RegionID, excludedRegions ...world.RegionID) world.RegionID {
	if s == nil || fleet == nil {
		return ""
	}
	region := s.Regions[regionID]
	if region == nil {
		return ""
	}
	neighbors := append([]world.RegionID(nil), region.Neighbors...)
	sort.Slice(neighbors, func(i, j int) bool { return neighbors[i] < neighbors[j] })
	for _, neighborID := range neighbors {
		neighbor := s.Regions[neighborID]
		if neighbor == nil || !neighbor.IsSea || neighbor.ID == fleet.RegionID {
			continue
		}
		excluded := false
		for _, excludedID := range excludedRegions {
			if neighbor.ID == excludedID {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}
		if s.SelectBattleDefender(fleet, neighbor.ID, true) == nil {
			return neighbor.ID
		}
	}
	return ""
}

// QueueNavalContactForWar, savaş ilişkisi kurulurken zaten aynı açık denizde
// bulunan ilk düşman filo çiftini temasa alır.
func (s *GameState) QueueNavalContactForWar(first, second faction.FactionID) *NavalContact {
	if s == nil {
		return nil
	}
	if first == "" || second == "" || first == second || s.PendingNavalContact != nil {
		return s.PendingNavalContact
	}
	ids := make([]army.ArmyID, 0, len(s.Armies))
	for id, fleet := range s.Armies {
		if fleet != nil && fleet.IsAtSea() && (fleet.OwnerID == string(first) || fleet.OwnerID == string(second)) {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for i, leftID := range ids {
		left := s.Armies[leftID]
		for _, rightID := range ids[i+1:] {
			right := s.Armies[rightID]
			if right == nil || left.RegionID != right.RegionID || left.OwnerID == right.OwnerID {
				continue
			}
			if !((left.OwnerID == string(first) && right.OwnerID == string(second)) || (left.OwnerID == string(second) && right.OwnerID == string(first))) {
				continue
			}
			return s.BeginNavalContact(left, right, left.RegionID, "", NavalContactWarOpening)
		}
	}
	return nil
}

func navalContactDefaultDecision(fleet *army.Army, trigger NavalContactTrigger) NavalContactDecision {
	if fleet != nil && fleet.NavalMission != nil {
		switch fleet.NavalMission.Kind {
		case army.NavalMissionPatrol:
			return NavalContactClash
		case army.NavalMissionBlockade:
			return NavalContactHold
		}
	}
	if trigger == NavalContactDirectAttack || trigger == NavalContactWarOpening || trigger == NavalContactMovement || trigger == NavalContactMissionAssignment {
		return NavalContactClash
	}
	return NavalContactHold
}
