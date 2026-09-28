package state

import (
	"sort"

	"mapp-game-go/internal/world"
)

// TransferRevokedMinorOwnership, diplomacy.RevokeMinorPrivilege sonrasında
// imtiyaz kaldırmanın ortak sahiplik/state temizliğini uygular. İlişki cezası
// diplomasi katmanında kalır; bu yöntem oyuncu ve AI akışlarının aynı bölge
// devri, üretim kuyruğu ve kuvvet tahliyesi sonucunu paylaşmasını sağlar.
func (s *GameState) TransferRevokedMinorOwnership(regionID world.RegionID, sovereignOwnerID string) string {
	if s == nil || regionID == "" || sovereignOwnerID == "" {
		return ""
	}
	region := s.Regions[regionID]
	if region == nil || region.IsSea {
		return ""
	}
	formerOwnerID := region.OwnerID
	region.OwnerID = sovereignOwnerID
	s.ClearProductionOrdersForRegion(regionID)
	delete(s.Sieges, regionID)
	s.evictForcesFromRevokedMinorRegion(regionID, sovereignOwnerID)
	return formerOwnerID
}

func (s *GameState) evictForcesFromRevokedMinorRegion(regionID world.RegionID, protectedOwnerID string) {
	if s == nil || regionID == "" {
		return
	}
	reference := s.Regions[regionID]
	if reference == nil {
		return
	}
	for _, currentArmy := range s.Armies {
		if currentArmy == nil || currentArmy.OwnerID == "" || currentArmy.OwnerID == protectedOwnerID {
			continue
		}
		if currentArmy.IsNaval {
			if currentArmy.RegionID != regionID && currentArmy.DockedRegionID != regionID {
				continue
			}
			if nearestSea := s.nearestSeaRegion(regionID); nearestSea != "" {
				currentArmy.RegionID = nearestSea
			}
			currentArmy.DockedRegionID = ""
			currentArmy.DockedSettlementID = ""
			continue
		}
		if currentArmy.RegionID != regionID {
			continue
		}
		if retreatRegion := s.nearestSovereignLandRegion(currentArmy.OwnerID, reference); retreatRegion != "" {
			currentArmy.RegionID = retreatRegion
			currentArmy.DockedRegionID = ""
			currentArmy.DockedSettlementID = ""
		}
	}
}

func (s *GameState) nearestSeaRegion(regionID world.RegionID) world.RegionID {
	region := s.Regions[regionID]
	if region == nil {
		return ""
	}
	for _, neighborID := range region.Neighbors {
		if neighbor := s.Regions[neighborID]; neighbor != nil && neighbor.IsSea {
			return neighbor.ID
		}
	}
	return ""
}

func (s *GameState) nearestSovereignLandRegion(ownerID string, reference *world.Region) world.RegionID {
	if s == nil || ownerID == "" || reference == nil {
		return ""
	}
	type candidate struct {
		id   world.RegionID
		dist int
	}
	candidates := make([]candidate, 0)
	for _, region := range s.Regions {
		if region == nil || region.IsSea || region.IsTerrainArea || region.ID == reference.ID || s.SovereignOwnerID(region) != ownerID {
			continue
		}
		dx := region.WorldX - reference.WorldX
		dy := region.WorldY - reference.WorldY
		candidates = append(candidates, candidate{id: region.ID, dist: dx*dx + dy*dy})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].dist != candidates[j].dist {
			return candidates[i].dist < candidates[j].dist
		}
		return candidates[i].id < candidates[j].id
	})
	if len(candidates) == 0 {
		return ""
	}
	return candidates[0].id
}
