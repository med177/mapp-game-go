package state

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

// LandRegionMoveCost returns the cost of entering a land region at its anchor.
// Normal regions are always passable; painted terrain is represented by
// separate runtime child nodes and only those nodes carry terrain movement
// costs or blocking.
func (s *GameState) LandRegionMoveCost(region *world.Region) (int, bool) {
	if s == nil || region == nil || region.IsSea {
		return 0, true
	}
	if region.IsTerrainArea {
		for _, area := range s.TerrainAreas {
			if area.ID == region.TerrainAreaID {
				extra, blocked := terrainAreaCostForID(area)
				if blocked {
					return 0, true
				}
				return 1 - extra, false
			}
		}
	}
	return 1, false
}

// LandRegionEntryCost returns the complete cost of entering a land region.
// Both player and AI movement must use this boundary: region locks reject
// entry, while painted terrain passability is determined by MoveCost.
func (s *GameState) LandRegionEntryCost(from world.RegionID, target *world.Region) (int, bool) {
	if target == nil || target.IsSea || !target.CanLandEnter() {
		return 0, false
	}
	cost, blocked := s.LandRegionMoveCost(target)
	if blocked {
		return 0, false
	}
	if passage := world.LandPassageBetween(s.LandPassages, from, target.ID); passage != nil && passage.MoveCost > cost {
		cost = passage.MoveCost
	}
	if cost < 1 {
		cost = 1
	}
	return cost, true
}

func terrainAreaCostForID(area world.TerrainArea) (int, bool) {
	if area.MoveCost == 0 {
		return 0, true
	}
	return area.MoveCost, false
}

// RepairArmiesInBlockedTerrain, senaryo veya eski kayıt yüklenirken geçilmez
// arazi alanının içinde kalmış kara ordularını bir kez geçerli bir kara
// bölgesine taşır.
func (s *GameState) RepairArmiesInBlockedTerrain() int {
	if s == nil || len(s.Armies) == 0 {
		return 0
	}
	repaired := 0
	for _, currentArmy := range s.Armies {
		if currentArmy == nil || currentArmy.IsNaval {
			continue
		}
		region := s.Regions[currentArmy.RegionID]
		if region == nil || region.IsSea {
			continue
		}
		if _, blocked := s.LandRegionMoveCost(region); !blocked {
			continue
		}
		destination := s.nearestRepairLandRegion(region.ID, currentArmy.OwnerID)
		if destination == "" {
			continue
		}
		currentArmy.PreviousRegionID = currentArmy.RegionID
		currentArmy.RegionID = destination
		currentArmy.DockedRegionID = ""
		currentArmy.DockedSettlementID = ""
		repaired++
	}
	return repaired
}

func (s *GameState) nearestRepairLandRegion(start world.RegionID, ownerID string) world.RegionID {
	if s == nil || start == "" {
		return ""
	}
	type queueItem struct{ id world.RegionID }
	queue := []queueItem{{id: start}}
	visited := map[world.RegionID]bool{start: true}
	var fallback world.RegionID
	for len(queue) > 0 {
		currentID := queue[0].id
		queue = queue[1:]
		region := s.Regions[currentID]
		if region != nil && currentID != start && !region.IsSea {
			if _, allowed := s.LandRegionEntryCost(start, region); allowed {
				if fallback == "" {
					fallback = currentID
				}
				if ownerID != "" && region.OwnerID == ownerID {
					return currentID
				}
			}
		}
		neighbors := append([]world.RegionID(nil), region.Neighbors...)
		sort.Slice(neighbors, func(i, j int) bool { return neighbors[i] < neighbors[j] })
		for _, neighborID := range neighbors {
			if neighborID == "" || visited[neighborID] || s.Regions[neighborID] == nil {
				continue
			}
			visited[neighborID] = true
			queue = append(queue, queueItem{id: neighborID})
		}
	}
	return fallback
}

// LandRegionAttritionPercent, bir arazi alanına giren ordunun kaybedeceği HP
// yüzdesini döner (çöl sıcağı, dağ yorgunluğu vb.). Normal bölgeler ve deniz
// bölgeleri için her zaman 0 döner; yıpranma yalnızca boyanmış arazi alanı
// çocuk düğümlerine özgüdür.
func (s *GameState) LandRegionAttritionPercent(region *world.Region) int {
	if s == nil || region == nil || !region.IsTerrainArea {
		return 0
	}
	for _, area := range s.TerrainAreas {
		if area.ID == region.TerrainAreaID {
			if area.MoveCost == 0 {
				return 0
			}
			return area.AttritionCost
		}
	}
	return 0
}

// ApplyLandRegionEntryAttrition, bir kara ordusu bir arazi alanına girdiğinde
// o alanın yıpranma yüzdesini orduya uygular. Deniz orduları ve normal
// bölgeler etkilenmez; kayıp birim sayısını döner.
func (s *GameState) ApplyLandRegionEntryAttrition(a *army.Army) int {
	if s == nil || a == nil || a.IsNaval {
		return 0
	}
	region := s.Regions[a.RegionID]
	percent := s.LandRegionAttritionPercent(region)
	if percent <= 0 {
		return 0
	}
	return a.ApplyAttritionPercent(percent)
}

// ArmyMaxMovePoints bu tur için bir ordunun toplam hareket havuzunu hesaplar.
// Mevsim etkisi önce en yavaş birimin tabanına uygulanır; komutan, teknoloji ve
// runtime bonusları bu iklimlendirilmiş değerin üzerine eklenir.
func (s *GameState) ArmyMaxMovePoints(a *army.Army) int {
	if s == nil || a == nil {
		return 1
	}

	movePoints := a.BaseMovePoints(s.UnitTypes)
	movePoints = movePoints * s.CurrentSeason().MovementMod() / 100
	if movePoints < 1 {
		movePoints = 1
	}

	if owner, ok := s.Factions[faction.FactionID(a.OwnerID)]; ok && s.TechTypes != nil {
		effects := tech.ComputeEffects(owner.Research.Completed, s.TechTypes)
		movePoints += effects.MoveBonus
		if a.IsNaval {
			movePoints += effects.NavalMoveBonus
		}
	}
	// Legacy senaryolardaki +1 zor AI hareketi korunur. 1300'ün veri güdümlü
	// fair_movement politikası oyuncu ve AI için aynı hareket kurallarını kullanır.
	if !s.AIDifficultyPolicy.FairMovement && s.Difficulty >= 3 && a.OwnerID != string(s.PlayerFactionID) {
		movePoints++
	}
	movePoints += a.CommanderMoveBonus()
	if movePoints < 1 {
		return 1
	}
	return movePoints
}

// RefreshArmyMovePoints hareket havuzlarının birim kompozisyonuyla
// senkronize olmasını sağlar. reset=true yeni senaryo başlangıcında tüm
// orduları yeni tur havuzuna alır; false mevcut turda harcanmış puanları korur.
func (s *GameState) RefreshArmyMovePoints(reset bool) {
	if s == nil {
		return
	}
	for _, currentArmy := range s.Armies {
		if currentArmy == nil {
			continue
		}
		maxPoints := s.ArmyMaxMovePoints(currentArmy)
		currentArmy.MaxMovePoints = maxPoints
		if reset || currentArmy.MovePoints > maxPoints {
			currentArmy.MovePoints = maxPoints
		}
		if currentArmy.MovePoints < 0 {
			currentArmy.MovePoints = 0
		}
	}
}

// RefreshArmyMovePointsAfterCompositionChange kompozisyonu değişen bir ordunun
// bu turdaki hareket havuzunu yeniden hesaplar. movementUsed false ise yeni
// havuz tamamen kullanılabilir; true ise kalan puan geri verilmez.
func (s *GameState) RefreshArmyMovePointsAfterCompositionChange(a *army.Army, movementUsed bool) {
	if s == nil || a == nil {
		return
	}

	a.MaxMovePoints = s.ArmyMaxMovePoints(a)
	if !movementUsed {
		a.MovePoints = a.MaxMovePoints
		return
	}
	if a.MovePoints > a.MaxMovePoints {
		a.MovePoints = a.MaxMovePoints
	}
	if a.MovePoints < 0 {
		a.MovePoints = 0
	}
}
