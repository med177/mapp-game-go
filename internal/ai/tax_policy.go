package ai

import (
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

const (
	aiTaxEmergencySatisfactionThreshold = 35
	aiTaxReliefSatisfactionThreshold    = 50
	aiTaxIncreaseSatisfactionThreshold  = 75
	aiTaxEmergencyStep                  = -20
	aiTaxReliefStep                     = -10
	aiTaxIncreaseStep                   = 10
)

// aiAdjustTaxesWithSteps vergi politikasını bölge bazında uygular. Savaş
// yorgunluğu da projeksiyona katılır; böylece savaşta yüksek görünen ama
// ekonomi tick'inden sonra isyan eşiğine yaklaşacak bölgelerde vergi azaltılır.
func aiAdjustTaxesWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	if gs == nil || fid == "" {
		return
	}
	warPenalty := diplomacy.IndependentWarSatisfactionPenalty(gs, fid)
	for _, region := range aiSortedRegions(gs) {
		if region == nil || region.IsSea || region.IsLocked || region.OwnerID != string(fid) {
			continue
		}

		projectedSatisfaction := region.Satisfaction - warPenalty
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
		region.TaxRate = clampAIInt(region.TaxRate+delta, 0, 100)
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
