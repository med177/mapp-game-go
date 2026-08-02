package state

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

const (
	navalEscortDefenseBonusPerFleet = 0.15
	navalEscortDefenseBonusCap      = 0.30
)

// CanAssignNavalMission merkezi oyuncu filo görevi doğrulamasıdır. Renderer
// yalnız adayları gösterir; gerçek state değişikliği bu kapıdan geçer.
func (s *GameState) CanAssignNavalMission(fleetID army.ArmyID, mission army.NavalMission) (bool, string) {
	if s == nil || fleetID == "" {
		return false, "Geçersiz filo."
	}
	fleet := s.Armies[fleetID]
	if fleet == nil || !fleet.IsNaval {
		return false, "Seçilen ordu bir donanma değil."
	}
	if mission.Kind == "" {
		return false, "Filo görevi seçilmedi."
	}

	switch mission.Kind {
	case army.NavalMissionPatrol:
		if !fleetHasWarship(s, fleet) {
			return false, "Bu görev için savaş gemisi gerekir."
		}
		if !fleet.IsAtSea() || mission.TargetRegionID != fleet.RegionID {
			return false, "Devriye görevi yalnızca filonun bulunduğu açık denizde atanabilir."
		}
		if !s.validSeaMissionTarget(mission.TargetRegionID) {
			return false, "Hedef deniz bölgesi geçerli değil."
		}
	case army.NavalMissionBlockade:
		if !fleetHasWarship(s, fleet) {
			return false, "Bu görev için savaş gemisi gerekir."
		}
		if !fleet.IsAtSea() || mission.TargetRegionID != fleet.RegionID {
			return false, "Abluka görevi yalnızca filonun bulunduğu açık denizde atanabilir."
		}
		if !s.IsValidNavalBlockadeTarget(fleet, mission.TargetRegionID) {
			return false, "Abluka yalnızca düşman kıyısına komşu deniz bölgesine atanabilir."
		}
	case army.NavalMissionEscort:
		if !fleetHasWarship(s, fleet) {
			return false, "Escort görevi için savaş gemisi gerekir."
		}
		target := s.Armies[mission.TargetFleetID]
		if target == nil || target.ID == fleet.ID || target.OwnerID != fleet.OwnerID || !target.IsNaval || target.TransportCapacity(s.UnitTypes) <= 0 {
			return false, "Escort hedefi aynı devlete ait nakliye filosu olmalı."
		}
	case army.NavalMissionTransport:
		if fleet.TransportCapacity(s.UnitTypes) <= 0 {
			return false, "Bu görev için nakliye kapasitesi gerekir."
		}
		if len(fleet.EmbarkedUnits) == 0 {
			return false, "Nakliye görevi için filoda taşınan kara ordusu yok."
		}
		target := s.Regions[mission.TargetRegionID]
		if target == nil || target.IsSea || !target.CanLandEnter() || !target.IsCoastal(s.Regions) {
			return false, "Nakliye hedefi kıyı kara bölgesi olmalı."
		}
	default:
		return false, "Bilinmeyen filo görevi."
	}
	return true, ""
}

func (s *GameState) validSeaMissionTarget(regionID world.RegionID) bool {
	if s == nil || regionID == "" {
		return false
	}
	region := s.Regions[regionID]
	return region != nil && region.IsSea && !region.IsLocked
}

// IsValidNavalBlockadeTarget, ablukanın yalnızca savaş halindeki düşmanın
// kıyı kara bölgelerine komşu denizlerde kurulabilmesini sağlar.
func (s *GameState) IsValidNavalBlockadeTarget(fleet *army.Army, seaID world.RegionID) bool {
	if s == nil || fleet == nil || !fleet.IsNaval || !s.validSeaMissionTarget(seaID) {
		return false
	}
	for _, land := range s.Regions {
		if land == nil || land.IsSea || land.OwnerID == "" || land.OwnerID == fleet.OwnerID || !s.atWar(fleet.OwnerID, land.OwnerID) {
			continue
		}
		for _, neighborID := range land.Neighbors {
			if neighborID == seaID {
				return true
			}
		}
	}
	return false
}

// AssignNavalMission doğrulanmış görevi filoya kopyalar.
func (s *GameState) AssignNavalMission(fleetID army.ArmyID, mission army.NavalMission) (bool, string) {
	if ok, reason := s.CanAssignNavalMission(fleetID, mission); !ok {
		return false, reason
	}
	fleet := s.Armies[fleetID]
	missionCopy := mission
	fleet.NavalMission = &missionCopy
	return true, ""
}

func (s *GameState) ClearNavalMission(fleetID army.ArmyID) bool {
	if s == nil {
		return false
	}
	fleet := s.Armies[fleetID]
	if fleet == nil || !fleet.IsNaval {
		return false
	}
	fleet.NavalMission = nil
	return true
}

// NavalFleetsAutoEngage, otomatik deniz savaşının tek görev kombinasyonunu
// tanımlar: aynı denizdeki Devriye filosu açıkça hedeflenmiş Abluka filosunu
// yakalar. Görevsiz filo, escort veya nakliye bu kapıdan otomatik savaş açmaz.
func (s *GameState) NavalFleetsAutoEngage(attacker, defender *army.Army) bool {
	if s == nil || attacker == nil || defender == nil || attacker.ID == defender.ID || attacker.OwnerID == defender.OwnerID || !attacker.IsAtSea() || !defender.IsAtSea() || attacker.RegionID != defender.RegionID {
		return false
	}
	return s.NavalFleetsAutoEngageAtSea(attacker, defender, attacker.RegionID)
}

// NavalFleetsAutoEngageAtSea, hareket eden filo henüz hedef denize yazılmadan
// savunucu seçilirken görev çiftini hedef deniz üzerinde değerlendirir.
func (s *GameState) NavalFleetsAutoEngageAtSea(attacker, defender *army.Army, seaID world.RegionID) bool {
	if s == nil || attacker == nil || defender == nil || seaID == "" || attacker.ID == defender.ID || attacker.OwnerID == defender.OwnerID || !attacker.IsAtSea() || !defender.IsAtSea() {
		return false
	}
	return navalMissionTargetsSea(attacker, army.NavalMissionPatrol, seaID) && navalMissionTargetsSea(defender, army.NavalMissionBlockade, seaID) ||
		navalMissionTargetsSea(attacker, army.NavalMissionBlockade, seaID) && navalMissionTargetsSea(defender, army.NavalMissionPatrol, seaID)
}

func navalMissionTargetsSea(fleet *army.Army, kind army.NavalMissionKind, seaID world.RegionID) bool {
	return fleet != nil && fleet.NavalMission != nil && fleet.NavalMission.Kind == kind && fleet.NavalMission.TargetRegionID == seaID
}

// NavalEscortDefenseBonus, aynı deniz bölgesindeki nakliye filosuna atanmış
// escort savaş gemilerinin deniz savunmasına katkısını döndürür. Bonus yalnız
// gerçek escort hedefi de aynı savaşa savunmacı olarak katılıyorsa uygulanır.
// Birden fazla escort toplamda yüzde 30 ile sınırlıdır.
func (s *GameState) NavalEscortDefenseBonus(sourceIDs []army.ArmyID, targetRegionID world.RegionID) float64 {
	if s == nil || targetRegionID == "" || len(sourceIDs) == 0 {
		return 0
	}
	escortCount := 0
	for _, sourceID := range sourceIDs {
		escort := s.Armies[sourceID]
		if escort == nil || escort.RegionID != targetRegionID || escort.NavalMission == nil || escort.NavalMission.Kind != army.NavalMissionEscort {
			continue
		}
		transport := s.Armies[escort.NavalMission.TargetFleetID]
		if transport == nil || transport.ID == escort.ID || transport.RegionID != targetRegionID || !transport.IsNaval || transport.TransportCapacity(s.UnitTypes) <= 0 {
			continue
		}
		escortCount++
	}
	bonus := float64(escortCount) * navalEscortDefenseBonusPerFleet
	if bonus > navalEscortDefenseBonusCap {
		return navalEscortDefenseBonusCap
	}
	return bonus
}

func fleetHasWarship(s *GameState, fleet *army.Army) bool {
	if s == nil || fleet == nil {
		return false
	}
	for _, unit := range fleet.Units {
		if unitType := s.UnitTypes[unit.TypeID]; unitType != nil && unitType.Category == army.CategoryNavalWar {
			return true
		}
	}
	return false
}
