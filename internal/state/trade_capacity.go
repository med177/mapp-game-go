package state

import (
	"math"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

const (
	tradeCenterGrowthStartVolume   = 50
	primaryTradeCenterGrowthStep   = 25
	secondaryTradeCenterGrowthStep = 50
)

// EffectiveRegionTradeCapacity bölgenin ticaret üretimi, diplomasi kapasitesi
// ve rota hacminde ortak kullanılan gerçek kapasitesini döner. Bina çarpanları
// önce ham kapasiteye uygulanır; merkez bonusu ise hacme göre dinamik büyür.
func (s *GameState) EffectiveRegionTradeCapacity(region *world.Region) int {
	if s == nil || region == nil || region.IsSea || region.TradeCapacity <= 0 {
		return 0
	}

	effective := s.baseRegionTradeCapacity(region)
	capacityBonus, _ := s.TradeCenterBenefits(region)
	effective += capacityBonus
	if effective < 0 {
		return 0
	}
	return effective
}

func (s *GameState) baseRegionTradeCapacity(region *world.Region) int {
	if s == nil || region == nil || region.IsSea || region.TradeCapacity <= 0 {
		return 0
	}
	capacity := float64(region.TradeCapacity)
	for _, buildingID := range region.Buildings {
		if building := s.BuildingTypes[buildingID]; building != nil && building.TradeCapacityMod > 0 {
			capacity *= building.TradeCapacityMod
		}
	}
	return int(math.Max(0, math.Round(capacity)))
}

// TradeCenterVolume, merkez tabelasında görünen hacmin state karşılığıdır.
// Merkezin bina sonrası yerel kapasitesine, merkez sahibinin aktif rota
// hacimleri eklenir. Merkez bonusu bu değerin üzerine uygulandığı için hesap
// dairesel değildir.
func (s *GameState) TradeCenterVolume(region *world.Region) int {
	if s == nil || region == nil || region.IsSea {
		return 0
	}
	volume := s.baseRegionTradeCapacity(region)
	for _, route := range s.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 || route.AmountPerTurn <= 0 {
			continue
		}
		if route.FromFactionID != region.OwnerID && route.ToFactionID != region.OwnerID {
			continue
		}
		volume += route.EffectiveAmountPerTurn()
	}
	return volume
}

// TradeCenterBenefits aktif ticaret merkezi olan bir bölgenin kapasite ve
// gümrük geliri bonuslarını döner. Bonus, bölgeyi elinde tutan devlete üretim
// çözümünde otomatik geçtiği için fetihte ek bir state aktarımı gerekmez.
func (s *GameState) TradeCenterBenefits(region *world.Region) (capacityBonus, incomeBonus int) {
	if s == nil || region == nil || region.IsSea {
		return 0, 0
	}
	for _, center := range s.TradeCenters.Centers {
		if center.ID != region.ID || !center.ActiveInYear(s.Year) {
			continue
		}
		return s.tradeCenterCapacityBonus(center), s.tradeCenterIncomeBonus(center)
	}
	return 0, 0
}

func (s *GameState) tradeCenterCapacityBonus(center world.TradeCenterDef) int {
	base := 0
	if center.TradeCapacityBonus != 0 {
		base = center.TradeCapacityBonus
	} else {
		switch center.Tier {
		case world.TradeCenterPrimary:
			base = s.TradeCenters.PrimaryTradeCapacityBonus
		case world.TradeCenterSecondary:
			base = s.TradeCenters.SecondaryTradeCapacityBonus
		}
	}
	return base + s.tradeCenterVolumeGrowth(center)
}

func (s *GameState) tradeCenterIncomeBonus(center world.TradeCenterDef) int {
	base := 0
	if center.TradeIncomeBonus != 0 {
		base = center.TradeIncomeBonus
	} else {
		switch center.Tier {
		case world.TradeCenterPrimary:
			base = s.TradeCenters.PrimaryTradeIncomeBonus
		case world.TradeCenterSecondary:
			base = s.TradeCenters.SecondaryTradeIncomeBonus
		}
	}
	return base + s.tradeCenterVolumeGrowth(center)*2
}

func (s *GameState) tradeCenterVolumeGrowth(center world.TradeCenterDef) int {
	region := s.Regions[center.ID]
	if region == nil || region.IsSea {
		return 0
	}
	volume := s.TradeCenterVolume(region)
	if volume <= tradeCenterGrowthStartVolume {
		return 0
	}
	step := primaryTradeCenterGrowthStep
	if center.Tier == world.TradeCenterSecondary {
		step = secondaryTradeCenterGrowthStep
	}
	return (volume - tradeCenterGrowthStartVolume + step - 1) / step
}

// BaseRegionTradeIncome mevsim, abluka ve teknoloji çarpanları öncesindeki
// pasif ticaret gelirini döner. Kapasite gelirinin yanında, merkez sahibinin
// topladığı küçük gümrük geliri de aynı ticaret akışına dahil edilir.
func (s *GameState) BaseRegionTradeIncome(region *world.Region) int {
	if s == nil || region == nil || region.IsSea {
		return 0
	}
	_, incomeBonus := s.TradeCenterBenefits(region)
	income := economy.RegionTradeIncome(s.EffectiveRegionTradeCapacity(region)) + incomeBonus
	if income < 0 {
		return 0
	}
	return income
}

// EffectiveFactionTradeCapacity bir devletin sahip olduğu kara bölgelerinin
// ortak efektif ticaret kapasitesini döner.
func (s *GameState) EffectiveFactionTradeCapacity(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	total := 0
	for _, region := range s.Regions {
		if region == nil || region.OwnerID != string(fid) {
			continue
		}
		total += s.EffectiveRegionTradeCapacity(region)
	}
	return total
}
