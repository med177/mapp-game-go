package ai

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/satisfaction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	aiTaxEmergencySatisfactionThreshold = 35
	aiTaxReliefSatisfactionThreshold    = 50
	aiTaxIncreaseSatisfactionThreshold  = 75
	aiTaxEmergencyStep                  = -20
	aiTaxReliefStep                     = -10
	aiTaxIncreaseStep                   = 10
)

// aiAdjustTaxesWithSteps vergi politikasını bölge bazında uygular. Ortak
// ekonomi-tick memnuniyet projeksiyonu kullanılır; böylece mevcut değeri
// yüksek görünen ama bir sonraki tick'te düşecek bölgelerde vergi azaltılır.
func aiAdjustTaxesWithSteps(gs *state.GameState, fid faction.FactionID, _ *[]TurnStep) {
	if gs == nil || fid == "" {
		return
	}
	satisfactionCalculator := satisfaction.NewCalculator(gs)
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.IsLocked || region.OwnerID != string(fid) {
			continue
		}

		// Vergi kararı, bir sonraki ekonomi tick'inde uygulanacak ortak
		// memnuniyet deltalarını hesaba katmalı. Aksi halde AI, örneğin
		// mevcut memnuniyeti 55 olan ancak vergi/tahıl/kuşatma nedeniyle
		// bir sonraki tick'te 50'nin altına inecek bölgede indirimi bir tur
		// geciktirir.
		projectedSatisfaction := region.Satisfaction + satisfactionCalculator.ForRegion(region).Total
		delta := 0
		switch {
		case projectedSatisfaction < aiTaxEmergencySatisfactionThreshold:
			delta = aiTaxEmergencyStep
		case projectedSatisfaction < aiTaxReliefSatisfactionThreshold:
			delta = aiTaxReliefStep
		case projectedSatisfaction >= aiTaxIncreaseSatisfactionThreshold:
			delta = aiTaxIncreaseStep
		default:
			continue
		}

		oldTax := region.TaxRate
		region.TaxRate = world.ClampTaxRate(region.TaxRate + delta)
		if region.TaxRate == oldTax {
			continue
		}
	}
}

func clampAIInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
