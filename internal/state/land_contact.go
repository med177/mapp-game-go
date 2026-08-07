package state

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

// LandContactDecision, temas eden kara ordusunun çatışma tutumunu belirtir.
type LandContactDecision string

const (
	LandContactUndecided LandContactDecision = ""
	LandContactClash     LandContactDecision = "clash"
	LandContactWithdraw  LandContactDecision = "withdraw"
	LandContactHold      LandContactDecision = "hold"
)

// LandContactTrigger, kara temasının hareketle mi yoksa savaş açılışıyla mı
// doğduğunu ayırır. Şimdilik tüm kara temasları hareket üzerinden oluşur;
// trigger alanı ileride doğrudan saldırı akışlarını ayırmak için korunur.
type LandContactTrigger string

const (
	LandContactMovement   LandContactTrigger = "movement"
	LandContactWarOpening LandContactTrigger = "war_opening"
)

// LandContact, iki düşman kara ordusunun aynı hedef bölge için karşılaşma
// kararını taşır. Temas çözülünce temizlenir ve save'e yazılmaz.
type LandContact struct {
	AttackerArmyID       army.ArmyID
	DefenderArmyID       army.ArmyID
	LandRegionID         world.RegionID
	AttackerFromRegionID world.RegionID
	Trigger              LandContactTrigger
	AttackerDecision     LandContactDecision
	DefenderDecision     LandContactDecision
	PlayerArmyID         army.ArmyID
	Prompted             bool
	MovementConsumed     bool
	AmbushArmyID         army.ArmyID
}

// BeginLandContact yeni kara temasını başlatır. Temas popup'ı açıldığında
// hareket eden saldıran ordu hedef bölgededir; kararlar tamamlanınca savaş,
// kuşatma veya geri çekilme bu konum üzerinden devam eder.
func (s *GameState) BeginLandContact(attacker, defender *army.Army, landID, attackerFrom world.RegionID, trigger LandContactTrigger) *LandContact {
	if s == nil || attacker == nil || defender == nil || attacker.IsNaval || defender.IsNaval || landID == "" || attacker.ID == defender.ID || attacker.OwnerID == defender.OwnerID {
		return nil
	}
	if s.PendingLandContact != nil {
		return s.PendingLandContact
	}
	contact := &LandContact{
		AttackerArmyID:       attacker.ID,
		DefenderArmyID:       defender.ID,
		LandRegionID:         landID,
		AttackerFromRegionID: attackerFrom,
		Trigger:              trigger,
		AttackerDecision:     LandContactClash,
		DefenderDecision:     LandContactClash,
	}
	if defender.InAmbush {
		contact.AmbushArmyID = defender.ID
	}
	if attacker.OwnerID == string(s.PlayerFactionID) {
		contact.PlayerArmyID = attacker.ID
		contact.AttackerDecision = LandContactUndecided
	} else if defender.OwnerID == string(s.PlayerFactionID) {
		contact.PlayerArmyID = defender.ID
		contact.DefenderDecision = LandContactUndecided
	}
	s.PendingLandContact = contact
	return contact
}

// LandContactDecisionForPlayer, temas kaydındaki oyuncu tarafının kararını
// uygular ve geri çekilme seçeneğinin gerçekten kullanılabilir olduğunu
// doğrular.
func (s *GameState) LandContactDecisionForPlayer(contact *LandContact, decision LandContactDecision) bool {
	if s == nil || contact == nil || contact.PlayerArmyID == "" || decision == LandContactUndecided {
		return false
	}
	if decision == LandContactWithdraw {
		if contact.AmbushArmyID != "" && contact.PlayerArmyID == contact.AttackerArmyID {
			return false
		}
		playerArmy := s.Armies[contact.PlayerArmyID]
		if playerArmy == nil {
			return false
		}
		if contact.PlayerArmyID == contact.AttackerArmyID {
			if playerArmy.RegionID != contact.LandRegionID || contact.AttackerFromRegionID == "" {
				return false
			}
		} else if (playerArmy.MovePoints <= 0 && contact.AmbushArmyID != playerArmy.ID) || s.LandContactRetreatRegion(playerArmy, contact.AttackerFromRegionID) == "" {
			return false
		}
	}
	if contact.AmbushArmyID != "" && decision == LandContactHold {
		return false
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

func (s *GameState) ClearLandContact() {
	if s != nil {
		s.PendingLandContact = nil
	}
}

func (s *GameState) LandContactBothClash(contact *LandContact) bool {
	return contact != nil && contact.AttackerDecision == LandContactClash && contact.DefenderDecision == LandContactClash
}

// LandContactWillClash, geri çekilmeyen taraflardan en az biri çatışmayı
// kabul ettiğinde muharebenin başlayacağını belirtir. Pozisyonu koru,
// çatışmadan kaçış değil savunma hazırlığıdır.
func (s *GameState) LandContactWillClash(contact *LandContact) bool {
	if contact == nil || contact.AttackerDecision == LandContactWithdraw || contact.DefenderDecision == LandContactWithdraw {
		return false
	}
	return contact.AttackerDecision == LandContactClash || contact.DefenderDecision == LandContactClash
}

// LandContactRetreatRegion, savunan kara ordusunun düşman olmayan komşu kara
// bölgesine çekilebileceği ilk deterministik hedefi döndürür.
func (s *GameState) LandContactRetreatRegion(armyRef *army.Army, excludedRegions ...world.RegionID) world.RegionID {
	if s == nil || armyRef == nil {
		return ""
	}
	region := s.Regions[armyRef.RegionID]
	if region == nil {
		return ""
	}
	neighbors := append([]world.RegionID(nil), region.Neighbors...)
	sort.Slice(neighbors, func(i, j int) bool { return neighbors[i] < neighbors[j] })
	for _, neighborID := range neighbors {
		neighbor := s.Regions[neighborID]
		if neighbor == nil || neighbor.IsSea || neighbor.ID == armyRef.RegionID || !landContactRegionExcluded(neighbor.ID, excludedRegions...) {
			continue
		}
		if neighbor.OwnerID != "" && neighbor.OwnerID != armyRef.OwnerID {
			continue
		}
		occupiedByEnemy := false
		for _, candidate := range s.Armies {
			if candidate != nil && !candidate.IsNaval && candidate.RegionID == neighbor.ID && candidate.OwnerID != armyRef.OwnerID {
				occupiedByEnemy = true
				break
			}
		}
		if !occupiedByEnemy {
			return neighbor.ID
		}
	}
	return ""
}

func landContactRegionExcluded(regionID world.RegionID, excluded ...world.RegionID) bool {
	for _, excludedID := range excluded {
		if regionID == excludedID {
			return false
		}
	}
	return true
}

// LandContactHasSafeWithdrawal reports whether a player-side withdrawal is
// available for the current contact and is used by the popup button state.
func (s *GameState) LandContactHasSafeWithdrawal(contact *LandContact) bool {
	if s == nil || contact == nil {
		return false
	}
	playerArmy := s.Armies[contact.PlayerArmyID]
	if playerArmy == nil {
		return false
	}
	if contact.AmbushArmyID != "" && contact.PlayerArmyID == contact.AttackerArmyID {
		return false
	}
	if contact.PlayerArmyID == contact.AttackerArmyID {
		return playerArmy.RegionID == contact.LandRegionID && contact.AttackerFromRegionID != ""
	}
	return (playerArmy.MovePoints > 0 || contact.AmbushArmyID == playerArmy.ID) && s.LandContactRetreatRegion(playerArmy, contact.AttackerFromRegionID) != ""
}
