package state

import "mapp-game-go/internal/world"

const (
	minTradeNetworkModifierPercent = -100
	maxTradeNetworkModifierPercent = 200
)

func clampTradeNetworkModifierPercent(value int) int {
	if value < minTradeNetworkModifierPercent {
		return minTradeNetworkModifierPercent
	}
	if value > maxTradeNetworkModifierPercent {
		return maxTradeNetworkModifierPercent
	}
	return value
}

func containsTradeNetworkRegion(ids []world.RegionID, id world.RegionID) bool {
	for _, candidate := range ids {
		if candidate == id {
			return true
		}
	}
	return false
}

// TradeNetworkIncomeModifier, bir ticaret merkezi üzerindeki event kaynaklı
// gelir değişimini toplar. Modifier'lar save'e yazıldığı için yeni senaryolar
// aynı mekanizmayı farklı merkezlerle kullanabilir.
func (s *GameState) TradeNetworkIncomeModifier(centerID world.RegionID) int {
	if s == nil || centerID == "" {
		return 0
	}
	total := 0
	for _, modifier := range s.TradeNetworkModifiers {
		if containsTradeNetworkRegion(modifier.CenterIDs, centerID) {
			total += modifier.TradeIncomePercent
		}
	}
	return clampTradeNetworkModifierPercent(total)
}

// RegionSpiceProductionModifier, event kaynaklı baharat üretim değişimini
// bölge bazında toplar ve ekonomi çözümlemesinin ortak giriş noktası olarak
// kullanılır.
func (s *GameState) RegionSpiceProductionModifier(regionID world.RegionID) int {
	if s == nil || regionID == "" {
		return 0
	}
	total := 0
	for _, modifier := range s.TradeNetworkModifiers {
		if containsTradeNetworkRegion(modifier.RegionIDs, regionID) {
			total += modifier.SpicePercent
		}
	}
	return clampTradeNetworkModifierPercent(total)
}
