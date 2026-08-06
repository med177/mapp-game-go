package ai

import (
	"fmt"

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
		reason := ""
		switch {
		case projectedSatisfaction < aiTaxEmergencySatisfactionThreshold:
			delta = aiTaxEmergencyStep
			reason = "isyan riskini azaltmak için"
		case projectedSatisfaction < aiTaxReliefSatisfactionThreshold:
			delta = aiTaxReliefStep
			reason = "memnuniyet açığını kapatmak için"
		case projectedSatisfaction >= aiTaxIncreaseSatisfactionThreshold:
			delta = aiTaxIncreaseStep
			reason = "yüksek memnuniyetten gelir sağlamak için"
		default:
			continue
		}

		oldTax := region.TaxRate
		region.TaxRate = clampAIInt(region.TaxRate+delta, 0, 100)
		if region.TaxRate == oldTax {
			continue
		}
		addTurnStep(steps, TurnStep{
			FactionID:    fid,
			Kind:         TurnStepInfo,
			TargetRegion: region.ID,
			FocusRegion:  region.ID,
			Message: fmt.Sprintf("%s %s bölgesinde vergiyi %%%d'den %%%d'e çekti (%s).",
				turnFactionName(gs, fid), turnRegionName(gs, region.ID), oldTax, region.TaxRate, reason),
		})
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
