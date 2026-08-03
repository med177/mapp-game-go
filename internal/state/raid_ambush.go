package state

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

const (
	RaidTaxLossPercent        = 80
	RaidProductionLossPercent = 50
)

// RaidState, bu tur ekonomi tick'inde uygulanacak tek bölge yağmasını taşır.
type RaidState struct {
	RegionID        world.RegionID    `json:"region_id"`
	RaiderFactionID faction.FactionID `json:"raider_faction_id"`
	RaiderArmyID    army.ArmyID       `json:"raider_army_id,omitempty"`
	Turn            int               `json:"turn"`
}

// RaidLootPreview, ekonomi tick'inde bu bölgeden yağmalanacak değerleri
// hesaplar. UI'daki kazanç tooltip'i ile gerçek ekonomi çözümlemesi aynı
// vergi/mevsim/arazi/abluka ve üretim katsayılarını kullanır.
func (s *GameState) RaidLootPreview(region *world.Region) RegionProductionSummary {
	if s == nil || region == nil || region.IsSea || region.OwnerID == "" {
		return RegionProductionSummary{}
	}

	goldMod := 1.0
	grainMod := 1.0
	for _, bid := range region.Buildings {
		if building := s.BuildingTypes[bid]; building != nil {
			goldMod *= building.GoldMod
			grainMod *= building.GrainMod
		}
	}
	retention := s.RegionBlockadeOutputRetentionPercent(region)
	income := ScaleBlockadeOutputForEconomy(
		int(float64(region.GoldIncome())*goldMod*float64(s.CurrentSeason().HarvestMod())/100),
		retention,
	)
	grain := int(float64(region.BaseGrainOutput) * grainMod)
	iron := region.BaseIronOutput
	timber := region.BaseTimberOutput
	stone := region.BaseStoneOutput
	spice := region.BaseSpiceOutput
	cloth := region.BaseClothOutput
	grain, iron, timber, stone, spice, cloth = applyRegionTerrainSpecialization(region.Terrain, grain, iron, timber, stone, spice, cloth)
	grain = ScaleBlockadeOutputForEconomy(grain, retention)
	iron = ScaleBlockadeOutputForEconomy(iron, retention)
	timber = ScaleBlockadeOutputForEconomy(timber, retention)
	stone = ScaleBlockadeOutputForEconomy(stone, retention)
	spice = ScaleBlockadeOutputForEconomy(spice, retention)
	cloth = ScaleBlockadeOutputForEconomy(cloth, retention)
	if bonus := s.CapitalRegionBonus(region); bonus != (RegionProductionSummary{}) {
		grain += bonus.Grain
	}
	productionPercent := 100 + s.RegionGrainProductionModifier(region.ID)
	if productionPercent < 0 {
		productionPercent = 0
	}
	grain = grain * productionPercent / 100

	return RegionProductionSummary{
		Gold:   income * RaidTaxLossPercent / 100,
		Grain:  grain * RaidProductionLossPercent / 100,
		Iron:   iron * RaidProductionLossPercent / 100,
		Timber: timber * RaidProductionLossPercent / 100,
		Stone:  stone * RaidProductionLossPercent / 100,
		Spice:  spice * RaidProductionLossPercent / 100,
		Cloth:  cloth * RaidProductionLossPercent / 100,
	}
}

func (s *GameState) RaidBlockReason(armyRef *army.Army, region *world.Region) string {
	if s == nil || armyRef == nil || region == nil {
		return "Yağmalama için geçerli bir ordu ve bölge gerekli."
	}
	if armyRef.IsNaval || armyRef.RegionID != region.ID || region.IsSea || region.OwnerID == "" || region.OwnerID == armyRef.OwnerID {
		return "Yağmalama yalnız düşman toprağında bekleyen kara ordusuyla yapılabilir."
	}
	if !s.isWarBetween(armyRef.OwnerID, region.OwnerID) {
		return "Yağmalama için iki devlet savaş halinde olmalı."
	}
	if s.SiegeByArmy(armyRef.ID) != nil {
		return "Kuşatma yapan ordu yağmalama görevi alamaz."
	}
	if s.Raids != nil {
		if raid := s.Raids[region.ID]; raid != nil && raid.Turn == s.Turn {
			return "Bu bölge bu tur zaten yağmalandı."
		}
	}
	return ""
}

func (s *GameState) CanRaid(armyRef *army.Army, region *world.Region) bool {
	return s.RaidBlockReason(armyRef, region) == ""
}

// ApplyRaid görevi kaydeder; kaynak transferi ekonomi tick'inde yapılır.
func (s *GameState) ApplyRaid(armyRef *army.Army, region *world.Region) bool {
	if !s.CanRaid(armyRef, region) {
		return false
	}
	if s.Raids == nil {
		s.Raids = make(map[world.RegionID]*RaidState)
	}
	s.Raids[region.ID] = &RaidState{
		RegionID:        region.ID,
		RaiderFactionID: faction.FactionID(armyRef.OwnerID),
		RaiderArmyID:    armyRef.ID,
		Turn:            s.Turn,
	}
	armyRef.MovePoints = 0
	armyRef.InAmbush = false
	return true
}

func (s *GameState) AmbushBlockReason(armyRef *army.Army, region *world.Region) string {
	if s == nil || armyRef == nil || region == nil {
		return "Pusu için geçerli bir ordu ve bölge gerekli."
	}
	if armyRef.IsNaval || armyRef.RegionID != region.ID || region.IsSea || region.OwnerID == "" || region.OwnerID == armyRef.OwnerID {
		return "Pusu yalnız düşman toprağında bekleyen kara ordusuyla kurulabilir."
	}
	if !s.isWarBetween(armyRef.OwnerID, region.OwnerID) {
		return "Pusu için iki devlet savaş halinde olmalı."
	}
	if s.SiegeByArmy(armyRef.ID) != nil {
		return "Kuşatma yapan ordu pusu kuramaz."
	}
	for _, candidate := range s.Armies {
		if candidate != nil && candidate.ID != armyRef.ID && !candidate.IsNaval && candidate.RegionID == region.ID && candidate.OwnerID != armyRef.OwnerID {
			return "Bölgede görünür bir düşman ordusu varken pusu kurulamaz."
		}
	}
	return ""
}

func (s *GameState) CanSetAmbush(armyRef *army.Army, region *world.Region) bool {
	return s.AmbushBlockReason(armyRef, region) == ""
}

func (s *GameState) SetAmbush(armyRef *army.Army, region *world.Region) bool {
	if !s.CanSetAmbush(armyRef, region) {
		return false
	}
	armyRef.InAmbush = true
	armyRef.MovePoints = 0
	return true
}

func (s *GameState) isWarBetween(a, b string) bool {
	if s == nil || a == "" || b == "" || a == b {
		return false
	}
	rel := s.Relations[faction.RelationKey(faction.FactionID(a), faction.FactionID(b))]
	return rel != nil && rel.Stance == faction.StanceWar
}
