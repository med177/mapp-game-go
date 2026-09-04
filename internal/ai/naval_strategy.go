package ai

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

// aiNavalStrategy is the stable entry point for the common naval policy.
// Mission production and merchant routing remain in their dedicated modules.
func aiNavalStrategy(gs *state.GameState, fid faction.FactionID) {
	aiNavalStrategyWithSteps(gs, fid, nil)
}

func aiNavalStrategyWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	aiNavalStrategyWithBudgetAndSteps(gs, fid, nil, steps)
}

func aiNavalStrategyWithBudgetAndSteps(gs *state.GameState, fid faction.FactionID, budget *aiBudget, steps *[]TurnStep) {
	var strategicContext *StrategicContext
	if gs != nil {
		strategicContext = prepareStrategicContext(gs, fid)
	}
	aiNavalStrategyWithStrategicContextAndSteps(gs, fid, budget, strategicContext, steps)
}
