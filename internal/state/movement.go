package state

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

// LandRegionMoveCost returns the cost of entering a land region at its anchor.
// It keeps the current node-based map model compatible while allowing painted
// subregions to block or tax movement.
func (s *GameState) LandRegionMoveCost(region *world.Region) (int, bool) {
	if s == nil || region == nil || region.IsSea {
		return 0, true
	}
	parentID := region.ID
	x, y := region.WorldX, region.WorldY
	if region.IsTerrainArea {
		parentID = region.ParentRegionID
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
	extra, blocked := world.TerrainAreaMovementCost(s.TerrainAreas, parentID, x, y)
	if blocked {
		return 0, true
	}
	return 1 - extra, false
}

func terrainAreaCostForID(area world.TerrainArea) (int, bool) {
	if area.MoveCost == 0 {
		return 0, true
	}
	return area.MoveCost, false
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
