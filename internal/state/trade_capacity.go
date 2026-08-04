package state

import (
	"math"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

// EffectiveRegionTradeCapacity bölgenin ticaret üretimi, diplomasi kapasitesi
// ve rota hacminde ortak kullanılan gerçek kapasitesini döner. Bina çarpanları
// önce ham kapasiteye uygulanır; aktif ticaret merkezi bonusu sonradan sabit
// kapasite olarak eklenir. Böylece merkez fethedildiğinde avantaj doğrudan yeni
// sahibinin ekonomi ve diplomasi hesaplarına geçer.
func (s *GameState) EffectiveRegionTradeCapacity(region *world.Region) int {
	if s == nil || region == nil || region.IsSea || region.TradeCapacity <= 0 {
		return 0
	}

	capacity := float64(region.TradeCapacity)
	for _, buildingID := range region.Buildings {
		if building := s.BuildingTypes[buildingID]; building != nil && building.TradeCapacityMod > 0 {
			capacity *= building.TradeCapacityMod
		}
	}
	if capacity < 0 {
		capacity = 0
	}

	effective := int(math.Round(capacity))
	capacityBonus, _ := s.TradeCenterBenefits(region)
	effective += capacityBonus
	if effective < 0 {
		return 0
	}
	return effective
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
	if center.TradeCapacityBonus != 0 {
		return center.TradeCapacityBonus
	}
	switch center.Tier {
	case world.TradeCenterPrimary:
		return s.TradeCenters.PrimaryTradeCapacityBonus
	case world.TradeCenterSecondary:
		return s.TradeCenters.SecondaryTradeCapacityBonus
	default:
		return 0
	}
}

func (s *GameState) tradeCenterIncomeBonus(center world.TradeCenterDef) int {
	if center.TradeIncomeBonus != 0 {
		return center.TradeIncomeBonus
	}
	switch center.Tier {
	case world.TradeCenterPrimary:
		return s.TradeCenters.PrimaryTradeIncomeBonus
	case world.TradeCenterSecondary:
		return s.TradeCenters.SecondaryTradeIncomeBonus
	default:
		return 0
	}
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
