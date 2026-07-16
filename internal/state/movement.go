package state

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
)

// ArmyMaxMovePoints bu tur için bir ordunun toplam hareket havuzunu hesaplar.
// Mevsim etkisi önce en yavaş birimin tabanına uygulanır; komutan, teknoloji ve
// zorluk bonusları bu iklimlendirilmiş değerin üzerine eklenir.
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
	if s.Difficulty >= 3 && a.OwnerID != string(s.PlayerFactionID) {
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
