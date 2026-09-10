package state

import (
	"math"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
)

// SiegeBreachGainPreview, aktif kuşatmanın mevcut canlı state ile bir turda
// üreteceği gedik ilerlemesini hesaplar. Çözümleme ile aynı birim uygunluğu,
// destek ordusu, HP, komutan, teknoloji ve savunucu etkilerini kullanır; yan
// etkisi yoktur ve yalnızca UI önizlemesi için state okur.
func (s *GameState) SiegeBreachGainPreview(siege *SiegeState, attacker *army.Army) float64 {
	if s == nil || siege == nil || attacker == nil || attacker.IsNaval {
		return 0
	}
	target := s.Regions[siege.RegionID]
	if target == nil {
		return 0
	}
	fortLevel := target.FortificationLevel()
	breachPower, breachBonus := siegeBreachPowerPreview(s, attacker, fortLevel)
	for candidateID, candidate := range s.Armies {
		if candidateID == attacker.ID || candidate == nil || candidate.IsNaval ||
			candidate.RegionID != siege.RegionID || len(candidate.Units) == 0 ||
			!s.CanJoinActiveSiege(candidate, siege.RegionID) {
			continue
		}
		candidatePower, candidateBonus := siegeBreachPowerPreview(s, candidate, fortLevel)
		breachPower += candidatePower
		breachBonus += candidateBonus
	}
	if breachPower <= 0 {
		return 0
	}

	owner := s.Factions[faction.FactionID(attacker.OwnerID)]
	techBonus := 0.0
	if owner != nil && s.TechTypes != nil {
		techBonus = tech.ComputeEffects(owner.Research.Completed, s.TechTypes).SiegeAttackMod
	}
	gain := breachPower + float64(int(techBonus*8+0.5)+breachBonus-fortLevel/2)
	defender := s.SelectBattleDefender(attacker, siege.RegionID, false)
	if defender != nil {
		gain -= float64(defender.TotalDefense(s.UnitTypes) / 120)
	}
	if gain < 0 {
		return 0
	}
	return gain
}

func siegeBreachPowerPreview(s *GameState, attacker *army.Army, fortLevel int) (power float64, bonus int) {
	if s == nil || attacker == nil {
		return 0, 0
	}
	for _, unit := range attacker.Units {
		unitType := s.UnitTypes[unit.TypeID]
		if unitType == nil || unitType.Category != army.CategorySiege {
			continue
		}
		maxFortLevel := int(unitType.Tier) + 2
		if unitType.SiegeBreachMaxFortLevel > maxFortLevel {
			maxFortLevel = unitType.SiegeBreachMaxFortLevel
		}
		if fortLevel > maxFortLevel {
			continue
		}
		hp := unit.CurrentHP
		if hp < 0 {
			hp = 0
		}
		if hp > army.MaxUnitHP {
			hp = army.MaxUnitHP
		}
		multiplier := unitType.SiegeBreachMultiplier
		if multiplier <= 0 {
			multiplier = 1
		}
		power += float64(2+int(unitType.Tier)) * float64(hp) / float64(army.MaxUnitHP) * multiplier
	}
	_, bonus = attacker.CommanderSiegeBonuses()
	return power, bonus
}

// SiegeTurnsUntilMinorBreach, mevcut kesirli ilerleme dahil olmak üzere ilk
// küçük gediğin kaç çözümleme turu sonra oluşacağını döner.
func SiegeTurnsUntilMinorBreach(siege *SiegeState, gain float64) (int, bool) {
	if siege == nil {
		return 0, false
	}
	minor, _ := SiegeBreachThresholds(siege.FortLevel)
	remaining := float64(minor-siege.BreachProgress) - siege.BreachProgressRemainder
	if remaining <= 0 {
		return 0, true
	}
	if gain <= 0 || math.IsNaN(gain) || math.IsInf(gain, 0) {
		return 0, false
	}
	return int(math.Ceil(remaining / gain)), true
}
