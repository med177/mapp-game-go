package state

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
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
	case army.NavalMissionPatrol, army.NavalMissionBlockade:
		if !fleetHasWarship(s, fleet) {
			return false, "Bu görev için savaş gemisi gerekir."
		}
		if !s.validSeaMissionTarget(mission.TargetRegionID) {
			return false, "Hedef deniz bölgesi geçerli değil."
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
