package state

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

const (
	tradePowerPerCapacity = 10
	tradePowerPerRoute    = 5
	primaryCenterPower    = 50
	secondaryCenterPower  = 25
)

// TradePowerForFaction, devletin mevcut yerel kapasitesi, aktif rota hacmi ve
// sahip olduğu ticaret merkezlerinden türeyen EU4-benzeri ticaret gücünü verir.
// Değer save'e yazılmaz; rota, bina veya mülkiyet değişince kendiliğinden yenilenir.
func (s *GameState) TradePowerForFaction(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	power := s.EffectiveFactionTradeCapacity(fid) * tradePowerPerCapacity
	for _, route := range s.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 {
			continue
		}
		if route.FromFactionID == string(fid) || route.ToFactionID == string(fid) {
			power += route.EffectiveAmountPerTurn() * tradePowerPerRoute
		}
	}
	for _, center := range s.TradeCenters.Centers {
		region := s.Regions[center.ID]
		if region == nil || region.OwnerID != string(fid) || !center.ActiveInYear(s.Year) {
			continue
		}
		if center.Tier == world.TradeCenterPrimary {
			power += primaryCenterPower
		} else {
			power += secondaryCenterPower
		}
	}
	if power < 0 {
		return 0
	}
	return power
}

// TradePowerSharePercent, tüm aktif devletler arasındaki toplam ticaret gücü
// payını tam yüzde olarak döndürür.
func (s *GameState) TradePowerSharePercent(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	total := 0
	for id, f := range s.Factions {
		if f == nil || f.IsEliminated {
			continue
		}
		total += s.TradePowerForFaction(id)
	}
	if total <= 0 {
		return 0
	}
	return s.TradePowerForFaction(fid) * 100 / total
}

func (s *GameState) tradeCenterPower(center world.TradeCenterDef) map[faction.FactionID]int {
	powers := make(map[faction.FactionID]int)
	region := s.Regions[center.ID]
	if region == nil || !center.ActiveInYear(s.Year) {
		return powers
	}
	owner := faction.FactionID(region.OwnerID)
	powers[owner] = s.EffectiveRegionTradeCapacity(region) * tradePowerPerCapacity
	if center.Tier == world.TradeCenterPrimary {
		powers[owner] += primaryCenterPower
	} else {
		powers[owner] += secondaryCenterPower
	}
	for _, route := range s.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 || route.EffectiveAmountPerTurn() <= 0 {
			continue
		}
		if route.FromFactionID == region.OwnerID || route.ToFactionID == region.OwnerID {
			powers[faction.FactionID(route.FromFactionID)] += route.EffectiveAmountPerTurn() * tradePowerPerRoute
			powers[faction.FactionID(route.ToFactionID)] += route.EffectiveAmountPerTurn() * tradePowerPerRoute
		}
	}
	return powers
}

// TradeCenterPowerSharePercent bir merkezin gelir havuzundaki payı döndürür.
func (s *GameState) TradeCenterPowerSharePercent(centerID world.RegionID, fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	var center world.TradeCenterDef
	found := false
	for _, candidate := range s.TradeCenters.Centers {
		if candidate.ID == centerID {
			center, found = candidate, true
			break
		}
	}
	if !found {
		return 0
	}
	powers := s.tradeCenterPower(center)
	total := 0
	for _, power := range powers {
		total += power
	}
	if total <= 0 {
		return 0
	}
	return powers[fid] * 100 / total
}

// TradePowerCommerceIncome, merkez havuzlarının ticaret gücü payına düşen
// sınırlı ek altın geliridir. Rota hacmi havuzu büyüttüğü için rota açmak,
// yalnızca transfer değil merkez ticareti açısından da değer üretir.
func (s *GameState) TradePowerCommerceIncome(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	income := 0
	for _, center := range s.TradeCenters.Centers {
		region := s.Regions[center.ID]
		if region == nil || !center.ActiveInYear(s.Year) {
			continue
		}
		pool := 4 + s.TradeCenterVolume(region)/10
		income += pool * s.TradeCenterPowerSharePercent(center.ID, fid) / 100
	}
	return income
}
