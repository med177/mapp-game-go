package state

import (
	"math"
	"sort"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

const generatedTradeNetworkLinkCount = 2

// EnsureTradeNetworkCoverage her aktif, kara toprağı olan devletin tarihsel
// merkez ağına bir giriş düğümü olmasını sağlar. Tarihsel merkezi olmayan veya
// onu kaybetmiş devletler için en uygun kendi bölgesinde NetworkOnly geçidi
// üretilir. Bu geçit görünür ağ bağlantısı ve merchant rota fallback'i sağlar;
// tarihsel merkez kapasitesi ya da gümrük geliri vermez.
func (s *GameState) EnsureTradeNetworkCoverage() {
	if s == nil || len(s.Factions) == 0 || len(s.Regions) == 0 {
		return
	}

	staticCenters := make([]world.TradeCenterDef, 0, len(s.TradeCenters.Centers))
	for _, center := range s.TradeCenters.Centers {
		if !center.NetworkOnly {
			staticCenters = append(staticCenters, center)
		}
	}
	if len(staticCenters) == 0 {
		s.TradeCenters.Centers = staticCenters
		return
	}

	factionIDs := make([]faction.FactionID, 0, len(s.Factions))
	for factionID := range s.Factions {
		factionIDs = append(factionIDs, factionID)
	}
	sort.Slice(factionIDs, func(i, j int) bool { return factionIDs[i] < factionIDs[j] })

	generated := make([]world.TradeCenterDef, 0, len(factionIDs))
	for _, factionID := range factionIDs {
		f := s.Factions[factionID]
		if f == nil || f.IsEliminated || s.factionHasActiveHistoricalTradeCenter(factionID, staticCenters) {
			continue
		}
		gateway := s.bestTradeNetworkGatewayRegion(factionID)
		if gateway == nil {
			continue
		}
		links := s.nearestHistoricalTradeCenters(gateway, staticCenters)
		if len(links) == 0 {
			continue
		}
		generated = append(generated, world.TradeCenterDef{
			ID:          gateway.ID,
			Tier:        world.TradeCenterSecondary,
			Links:       links,
			NetworkOnly: true,
		})
	}

	s.TradeCenters.Centers = append(staticCenters, generated...)
}

func (s *GameState) factionHasActiveHistoricalTradeCenter(factionID faction.FactionID, centers []world.TradeCenterDef) bool {
	for _, center := range centers {
		if !center.ActiveInYear(s.Year) || center.OffMap {
			continue
		}
		region := s.Regions[center.ID]
		if region != nil && !region.IsSea && region.OwnerID == string(factionID) {
			return true
		}
	}
	return false
}

func (s *GameState) bestTradeNetworkGatewayRegion(factionID faction.FactionID) *world.Region {
	var best *world.Region
	bestScore := -1
	for _, region := range s.Regions {
		if region == nil || region.IsSea || region.IsLocked || region.OwnerID != string(factionID) {
			continue
		}
		score := region.TradeCapacity*1000 + region.BaseGoldIncome*10
		if region.IsCoastal(s.Regions) {
			score += 100000
		}
		if region.HasPort() {
			score += 10000
		}
		if best == nil || score > bestScore || score == bestScore && region.ID < best.ID {
			best = region
			bestScore = score
		}
	}
	return best
}

func (s *GameState) nearestHistoricalTradeCenters(region *world.Region, centers []world.TradeCenterDef) []world.RegionID {
	if region == nil {
		return nil
	}
	type candidate struct {
		id       world.RegionID
		distance float64
	}
	candidates := make([]candidate, 0, len(centers))
	for _, center := range centers {
		if !center.ActiveInYear(s.Year) || center.OffMap || center.ID == region.ID {
			continue
		}
		target := s.Regions[center.ID]
		if target == nil || target.IsSea {
			continue
		}
		candidates = append(candidates, candidate{
			id:       center.ID,
			distance: math.Hypot(float64(region.WorldX-target.WorldX), float64(region.WorldY-target.WorldY)),
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].distance != candidates[j].distance {
			return candidates[i].distance < candidates[j].distance
		}
		return candidates[i].id < candidates[j].id
	})
	limit := generatedTradeNetworkLinkCount
	if len(candidates) < limit {
		limit = len(candidates)
	}
	links := make([]world.RegionID, 0, limit)
	for i := 0; i < limit; i++ {
		links = append(links, candidates[i].id)
	}
	return links
}
