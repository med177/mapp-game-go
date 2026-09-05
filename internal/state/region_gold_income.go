package state

import (
	"fmt"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

// RegionGoldIncomeContribution, bölge panelindeki altın üretiminin tek bir
// kaynağını gösterir. Value değerleri toplandığında Total ile aynı sonucu
// verir.
type RegionGoldIncomeContribution struct {
	Label string
	Value int
}

// RegionGoldIncomeBreakdown bölgenin üretim özetindeki Gold kalemini, aynı
// hesap sırasını koruyarak kaynaklarına ayırır.
func (s *GameState) RegionGoldIncomeBreakdown(region *world.Region) ([]RegionGoldIncomeContribution, int) {
	return s.regionGoldIncomeBreakdown(region, true)
}

func (s *GameState) regionGoldIncomeBreakdown(region *world.Region, applyBlockade bool) (lines []RegionGoldIncomeContribution, total int) {
	if s == nil || region == nil || region.IsSea || region.IsTerrainArea || region.OwnerID == "" {
		return nil, 0
	}

	appendLine := func(label string, value int) {
		if value != 0 {
			lines = append(lines, RegionGoldIncomeContribution{Label: label, Value: value})
		}
	}

	baseTax := region.GoldIncome()
	local := baseTax
	appendLine("Vergi (oran + memnuniyet)", local)
	goldMultiplier := 1.0
	buildingCounts := make(map[string]int)
	buildingOrder := make([]string, 0, len(region.Buildings))
	for _, bid := range region.Buildings {
		building := s.BuildingTypes[bid]
		if building == nil {
			continue
		}
		if buildingCounts[bid] == 0 {
			buildingOrder = append(buildingOrder, bid)
		}
		buildingCounts[bid]++
	}
	for _, bid := range buildingOrder {
		building := s.BuildingTypes[bid]
		if building == nil || building.GoldMod == 1 {
			continue
		}
		for level := 0; level < buildingCounts[bid]; level++ {
			goldMultiplier *= building.GoldMod
		}
		next := int(float64(baseTax) * goldMultiplier)
		appendLine(fmt.Sprintf("%s etkisi", building.NameTR), next-local)
		local = next
	}

	seasonValue := local * s.CurrentSeason().HarvestMod() / 100
	appendLine("Mevsim etkisi", seasonValue-local)
	local = seasonValue

	retention := 100
	if applyBlockade {
		retention = s.RegionBlockadeOutputRetentionPercent(region)
	}
	retained := scaleBlockadeOutput(local, retention)
	appendLine("Abluka etkisi", retained-local)
	local = retained

	tradeBase := s.BaseRegionTradeIncome(region) - s.RegionTradeCenterIncome(region)
	trade := s.BaseRegionTradeIncome(region) * s.CurrentSeason().TradeMod() / 100
	trade = scaleBlockadeOutput(trade, retention)
	tradeBase = tradeBase * s.CurrentSeason().TradeMod() / 100
	tradeBase = scaleBlockadeOutput(tradeBase, retention)
	if owner, ok := s.Factions[faction.FactionID(region.OwnerID)]; ok && owner != nil && s.TechTypes != nil {
		marketGoldMod := tech.ComputeEffects(owner.Research.Completed, s.TechTypes).MarketGoldMod
		techTrade := int(float64(trade) * (1.0 + marketGoldMod))
		techBase := int(float64(tradeBase) * (1.0 + marketGoldMod))
		appendLine("Pasif ticaret", techBase)
		appendLine("Ticaret merkezi", techTrade-techBase)
		appendLine("Teknoloji (ticaret)", techTrade-trade)
		trade = techTrade
	} else {
		appendLine("Pasif ticaret", tradeBase)
		appendLine("Ticaret merkezi", trade-tradeBase)
	}

	appendLine("Teknoloji (bölge)", s.regionTechnologyGoldBonus(region))
	appendLine("Başkent bonusu", s.CapitalRegionBonus(region).Gold)
	for _, line := range lines {
		total += line.Value
	}
	return lines, total
}

func (s *GameState) regionTechnologyGoldBonus(region *world.Region) int {
	if s == nil || region == nil || s.TechTypes == nil {
		return 0
	}
	owner, ok := s.Factions[faction.FactionID(region.OwnerID)]
	if !ok || owner == nil {
		return 0
	}
	return tech.ComputeEffects(owner.Research.Completed, s.TechTypes).GoldPerRegion
}
