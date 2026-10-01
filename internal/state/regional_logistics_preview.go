package state

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

// RegionSettlementLogisticsBuffer, yerleşimlerin çözümleme sırasında orduya
// sağladığı yerel tamponu hesaplar. Bu değer hem gerçek ikmal çözümlemesinde
// hem de çözümleme snapshot'ı henüz oluşmamış panel önizlemesinde aynıdır.
func (s *GameState) RegionSettlementLogisticsBuffer(region *world.Region) int {
	if region == nil {
		return 0
	}
	buffer := 0
	for _, settlement := range region.Settlements {
		switch settlement.Type {
		case world.SettlementCity:
			buffer += 8
		case world.SettlementTown:
			buffer += 5
		case world.SettlementFortress:
			buffer += 6
		case world.SettlementPort:
			buffer += 6
		default:
			buffer += 4
		}
		if settlement.IsCenter {
			buffer += 4
		}
	}
	if s != nil && s.IsCapitalRegion(region) {
		buffer += CapitalRegionLogisticsBonus
	}
	if tradeCapacity := region.TradeCapacity / 2; tradeCapacity > 0 {
		if tradeCapacity > 6 {
			tradeCapacity = 6
		}
		buffer += tradeCapacity
	}
	return buffer
}

// RegionReserveSupport, yerel üretim ve yerleşim tamponundan türeyen sınırlı
// merkez rezervi desteğini hesaplar.
func RegionReserveSupport(availableGrain, production, settlementBuffer int) int {
	if availableGrain <= 0 {
		return 0
	}
	cap := production/2 + settlementBuffer/2 + 4
	if cap < 4 {
		cap = 4
	}
	reserve := availableGrain / 10
	if reserve > cap {
		reserve = cap
	}
	return reserve
}

// RegionGranaryStorageCapacity, bölgedeki ambar binalarının ikmal için
// kullanılabilecek kapasitesini toplar.
func (s *GameState) RegionGranaryStorageCapacity(region *world.Region) int {
	if s == nil || region == nil {
		return 0
	}
	capacity := 0
	for _, buildingID := range region.Buildings {
		building := s.BuildingTypes[buildingID]
		if building != nil && building.StorageCapacity > 0 {
			capacity += building.StorageCapacity
		}
	}
	return capacity
}

// PreviewRegionalLogisticsStatuses, henüz ekonomi çözümlemesi snapshot
// üretmemişse panelin kullanacağı tüm bölgelerin yan etkisiz ikmal
// önizlemesini tek geçişte üretir.
//
// Merkez rezervi desteği sınırlı ve bölge sırasına bağlı olduğu için tek bir
// bölgeyi hesaplamak bile önceki bölgelerin sırasını gerektirir. Sonuç kümesini
// bir defada üretmek, panelin her çiziminde aynı global taramayı tekrarlamayı
// önler. Filo yükü burada tüketilmez; yalnızca gösterilecek mevcut katkı
// hesaplanır.
func (s *GameState) PreviewRegionalLogisticsStatuses() map[world.RegionID]RegionLogisticsStatus {
	if s == nil {
		return nil
	}
	if s.RegionalLogisticsPreviewCacheValid {
		return s.RegionalLogisticsPreviewCache
	}
	statuses := make(map[world.RegionID]RegionLogisticsStatus)

	armiesByRegion := make(map[world.RegionID][]*army.Army)
	for _, currentArmy := range s.Armies {
		if currentArmy == nil || currentArmy.IsNaval || len(currentArmy.Units) == 0 {
			continue
		}
		region := s.Regions[currentArmy.RegionID]
		if region == nil || region.IsSea {
			continue
		}
		armiesByRegion[currentArmy.RegionID] = append(armiesByRegion[currentArmy.RegionID], currentArmy)
	}
	for _, armies := range armiesByRegion {
		sort.Slice(armies, func(i, j int) bool {
			return armies[i].ID < armies[j].ID
		})
	}

	regionIDs := make([]world.RegionID, 0, len(armiesByRegion))
	for currentRegionID := range armiesByRegion {
		regionIDs = append(regionIDs, currentRegionID)
	}
	sort.Slice(regionIDs, func(i, j int) bool {
		left := s.Regions[regionIDs[i]]
		right := s.Regions[regionIDs[j]]
		leftCapital := s.IsCapitalRegion(left)
		rightCapital := s.IsCapitalRegion(right)
		if leftCapital != rightCapital {
			return leftCapital
		}
		return regionIDs[i] < regionIDs[j]
	})

	availableReserveByFaction := make(map[string]int, len(s.Factions))
	for factionID, faction := range s.Factions {
		if faction != nil && faction.Grain > 0 {
			availableReserveByFaction[string(factionID)] = faction.Grain
		}
	}

	for _, currentRegionID := range regionIDs {
		armies := armiesByRegion[currentRegionID]
		region := s.Regions[currentRegionID]
		if region == nil || len(armies) == 0 {
			continue
		}

		ownerID := armies[0].OwnerID
		demand := 0
		units := 0
		for _, currentArmy := range armies {
			demand += s.EffectiveArmyGrainUpkeep(currentArmy)
			units += len(currentArmy.Units)
		}
		if units <= 0 {
			continue
		}

		status := RegionLogisticsStatus{
			RegionID:  currentRegionID,
			OwnerID:   ownerID,
			Demand:    demand,
			ArmyCount: len(armies),
		}
		if demand > 0 {
			status.LocalProduction = s.RegionMilitaryGrainProduction(region)
			if region.OwnerID != ownerID {
				status.LocalProduction = 0
			}
			status.SettlementBuffer = s.RegionSettlementLogisticsBuffer(region)
			status.BlockadePercent = s.RegionBlockadePercent(region, ownerID)
			status.SettlementBuffer = status.SettlementBuffer * (100 - status.BlockadePercent) / 100

			availableReserve := availableReserveByFaction[ownerID]
			status.GranarySupport = minRegionalLogistics(availableReserve, s.RegionGranaryStorageCapacity(region))
			availableReserve -= status.GranarySupport
			status.ReserveSupport = RegionReserveSupport(availableReserve, status.LocalProduction, status.SettlementBuffer)
			availableReserveByFaction[ownerID] = availableReserve - status.ReserveSupport
			status.Capacity = status.LocalProduction + status.SettlementBuffer + status.GranarySupport + status.ReserveSupport

			if shortage := demand - status.Capacity; shortage > 0 {
				navalSupport := previewNavalSupplyForRegion(s, armies, shortage)
				status.NavalSupplyGrainSpent = navalSupport
				status.Capacity += navalSupport
			}
			if status.Capacity < 4 {
				status.Capacity = 4
			}
			status.Overload = demand - status.Capacity
		}

		if demand > 0 {
			statuses[currentRegionID] = status
		}
	}
	s.RegionalLogisticsPreviewCache = statuses
	s.RegionalLogisticsPreviewCacheValid = true
	return statuses
}

// InvalidateRegionalLogisticsPreview, ordu/filo/rezerv veya bölge üretimi
// değiştiğinde panel önizleme cache'ini bir sonraki isteğe bırakır.
func (s *GameState) InvalidateRegionalLogisticsPreview() {
	if s == nil {
		return
	}
	s.RegionalLogisticsPreviewCacheValid = false
}

// PreviewRegionalLogisticsStatus, tek bölge isteyen eski çağrı noktaları için
// tüm önizleme kümesinden ilgili sonucu seçer. GameState cache'i aynı state
// invalid edilene kadar bu global taramayı bir kez tutar.
func (s *GameState) PreviewRegionalLogisticsStatus(regionID world.RegionID) (RegionLogisticsStatus, bool) {
	if s == nil || regionID == "" {
		return RegionLogisticsStatus{}, false
	}
	status, ok := s.PreviewRegionalLogisticsStatuses()[regionID]
	return status, ok
}

// ArmyLogisticsDamageVisible, bölge bilgi panelindeki güncel ikmal aşımına göre
// kara ordusunun uyarı rozetini belirler. Deniz orduları bölgesel önizlemeye
// dahil olmadığı için onlar için çözümlemede kaydedilmiş deniz zayiatı kullanılır.
func (s *GameState) ArmyLogisticsDamageVisible(armyID army.ArmyID) bool {
	return s.armyLogisticsDamageVisible(armyID, s.PreviewRegionalLogisticsStatuses())
}

// ArmyLogisticsDamageVisibleFromPreview aynı global önizleme kümesini birden
// fazla ordu için kullanır. Marker cache'i topluca yenilenirken her ordu için
// lojistik ağacını baştan kurmamak için renderer tarafından kullanılır.
func (s *GameState) ArmyLogisticsDamageVisibleFromPreview(armyID army.ArmyID, previews map[world.RegionID]RegionLogisticsStatus) bool {
	return s.armyLogisticsDamageVisible(armyID, previews)
}

func (s *GameState) armyLogisticsDamageVisible(armyID army.ArmyID, previews map[world.RegionID]RegionLogisticsStatus) bool {
	if s == nil || armyID == "" {
		return false
	}
	target := s.Armies[armyID]
	if target == nil {
		return false
	}
	if target.IsNaval {
		status, ok := s.ArmyLogistics[armyID]
		return ok && status.TotalHPDamage > 0
	}
	preview, ok := previews[target.RegionID]
	return ok && preview.Overload > 0
}

func previewNavalSupplyForRegion(s *GameState, armies []*army.Army, shortage int) int {
	if s == nil || shortage <= 0 {
		return 0
	}
	supplied := 0
	for _, target := range armies {
		if shortage <= 0 || target == nil || target.IsNaval || len(target.Units) == 0 {
			continue
		}
		demand := s.EffectiveArmyGrainUpkeep(target)
		if demand <= 0 {
			continue
		}
		for _, fleet := range s.Armies {
			if fleet == nil || !fleet.IsNaval || fleet.NavalMission == nil || fleet.NavalMission.Kind != army.NavalMissionSupplyArmy || fleet.NavalMission.TargetArmyID != target.ID || fleet.SupplyCargo.Grain <= 0 || !fleet.IsAtSea() || fleet.OwnerID != target.OwnerID {
				continue
			}
			land := s.Regions[target.RegionID]
			if land == nil || land.IsSea || !land.CanLandEnter() || !land.IsCoastal(s.Regions) || !regionsShareSeaNeighbor(s, land, fleet.RegionID) {
				continue
			}
			amount := fleet.SupplyCargo.Grain
			if amount > shortage {
				amount = shortage
			}
			if amount > demand {
				amount = demand
			}
			supplied += amount
			shortage -= amount
			break
		}
	}
	return supplied
}

func minRegionalLogistics(left, right int) int {
	if left < right {
		return left
	}
	return right
}
