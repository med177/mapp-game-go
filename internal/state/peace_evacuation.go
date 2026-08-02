package state

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

// EvacuateNavalLandingSiegesAfterPeace, barışla kapanan koalisyon savaşı
// sırasında denizden başlatılmış aktif kuşatmaları güvenli biçimde sonlandırır.
// Önce hedefe en yakın dost nakliye filolarının toplam boş kapasitesi denenir;
// yeterli kapasite yoksa kara ordusu en yakın kendi kara bölgesine çekilir.
func (s *GameState) EvacuateNavalLandingSiegesAfterPeace(left, right []faction.FactionID) int {
	if s == nil || len(s.Sieges) == 0 || len(left) == 0 || len(right) == 0 {
		return 0
	}

	leftSet := factionSet(left)
	rightSet := factionSet(right)
	evacuated := 0
	regionIDs := make([]world.RegionID, 0, len(s.Sieges))
	for regionID := range s.Sieges {
		regionIDs = append(regionIDs, regionID)
	}
	sort.Slice(regionIDs, func(i, j int) bool { return regionIDs[i] < regionIDs[j] })
	for _, regionID := range regionIDs {
		siege := s.Sieges[regionID]
		if siege == nil || !siege.NavalLanding {
			continue
		}
		attacker := s.Armies[siege.AttackerArmyID]
		target := s.Regions[regionID]
		if attacker == nil || attacker.IsNaval || target == nil || !opposingPeaceCoalitions(attacker.OwnerID, target.OwnerID, leftSet, rightSet) {
			continue
		}

		if fleets := s.nearestEvacuationFleets(attacker.OwnerID, target, len(attacker.Units)); len(fleets) > 0 {
			if commander := attacker.Commander; commander != nil {
				for _, fleet := range fleets {
					if fleet.EmbarkedCommander == nil {
						s.MoveCommanderIntoFleet(attacker.ID, fleet.ID)
						break
					}
				}
			}
			remaining := attacker.Units
			for _, fleet := range fleets {
				available := fleet.AvailableTransportCapacity(s.UnitTypes)
				if available > len(remaining) {
					available = len(remaining)
				}
				fleet.EmbarkedUnits = append(fleet.EmbarkedUnits, remaining[:available]...)
				remaining = remaining[available:]
				if len(remaining) == 0 {
					break
				}
			}
			if len(remaining) == 0 {
				s.RemoveArmy(attacker.ID)
				delete(s.Sieges, regionID)
				evacuated++
				continue
			}
		}

		if retreatRegion := s.nearestOwnedLandRegion(attacker.OwnerID, target); retreatRegion != "" {
			attacker.RegionID = retreatRegion
			attacker.DockedRegionID = ""
			attacker.DockedSettlementID = ""
			delete(s.Sieges, regionID)
			evacuated++
			continue
		}

		// Elenmiş veya bozuk legacy state'lerinde geçersiz kuşatma kaydı
		// ordunun state'te kalmasına izin verilmeden temizlenir.
		delete(s.Sieges, regionID)
	}
	return evacuated
}

type evacuationFleetCandidate struct {
	fleet *army.Army
	dist  int64
}

func (s *GameState) nearestEvacuationFleets(ownerID string, target *world.Region, unitCount int) []*army.Army {
	if s == nil || ownerID == "" || target == nil || unitCount <= 0 {
		return nil
	}
	candidates := make([]evacuationFleetCandidate, 0)
	for _, fleet := range s.Armies {
		if fleet == nil || fleet.OwnerID != ownerID || !fleet.IsNaval || fleet.AvailableTransportCapacity(s.UnitTypes) <= 0 {
			continue
		}
		location := s.Regions[fleet.RegionID]
		if fleet.DockedRegionID != "" {
			location = s.Regions[fleet.DockedRegionID]
		}
		if location == nil {
			continue
		}
		dx := int64(location.WorldX - target.WorldX)
		dy := int64(location.WorldY - target.WorldY)
		candidates = append(candidates, evacuationFleetCandidate{fleet: fleet, dist: dx*dx + dy*dy})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].dist != candidates[j].dist {
			return candidates[i].dist < candidates[j].dist
		}
		return candidates[i].fleet.ID < candidates[j].fleet.ID
	})

	selected := make([]*army.Army, 0, len(candidates))
	capacity := 0
	for _, candidate := range candidates {
		selected = append(selected, candidate.fleet)
		capacity += candidate.fleet.AvailableTransportCapacity(s.UnitTypes)
		if capacity >= unitCount {
			return selected
		}
	}
	return nil
}

func (s *GameState) nearestOwnedLandRegion(ownerID string, reference *world.Region) world.RegionID {
	if s == nil || ownerID == "" || reference == nil {
		return ""
	}
	var best *world.Region
	var bestDist int64
	for _, region := range s.Regions {
		if region == nil || region.IsSea || region.OwnerID != ownerID || region.ID == reference.ID {
			continue
		}
		dx := int64(region.WorldX - reference.WorldX)
		dy := int64(region.WorldY - reference.WorldY)
		dist := dx*dx + dy*dy
		if best == nil || dist < bestDist || dist == bestDist && region.ID < best.ID {
			best = region
			bestDist = dist
		}
	}
	if best == nil {
		return ""
	}
	return best.ID
}

func factionSet(ids []faction.FactionID) map[string]struct{} {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id != "" {
			set[string(id)] = struct{}{}
		}
	}
	return set
}

func opposingPeaceCoalitions(attackerOwner, defenderOwner string, left, right map[string]struct{}) bool {
	_, attackerLeft := left[attackerOwner]
	_, attackerRight := right[attackerOwner]
	_, defenderLeft := left[defenderOwner]
	_, defenderRight := right[defenderOwner]
	return attackerLeft && defenderRight || attackerRight && defenderLeft
}
