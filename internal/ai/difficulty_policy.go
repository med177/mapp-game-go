package ai

import (
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
)

func aiDifficultyLevel(gs *state.GameState) (scenario.AIDifficultyLevel, bool) {
	if gs == nil {
		return scenario.AIDifficultyLevel{}, false
	}
	return gs.AIDifficultyPolicy.Level(gs.Difficulty)
}

func aiPlanHorizonTurns(gs *state.GameState) int {
	if level, ok := aiDifficultyLevel(gs); ok && level.PlanHorizonTurns > 0 {
		return level.PlanHorizonTurns
	}
	return aiStrategicPlanReassessTurns
}

func aiPlanTargetRegionLimit(gs *state.GameState) int {
	if level, ok := aiDifficultyLevel(gs); ok && level.PlanTargetRegionLimit > 0 {
		return level.PlanTargetRegionLimit
	}
	return aiStrategicPlanRegionLimit
}

func aiPathSearchDepth(gs *state.GameState) int {
	if level, ok := aiDifficultyLevel(gs); ok && level.PathSearchDepth > 0 {
		return level.PathSearchDepth
	}
	return 8
}

func aiPlanMoveBonusPercent(gs *state.GameState) int {
	if level, ok := aiDifficultyLevel(gs); ok && level.PlanMoveBonusPercent > 0 {
		return level.PlanMoveBonusPercent
	}
	return 100
}

func aiProactiveWarEnabled(gs *state.GameState) bool {
	if level, ok := aiDifficultyLevel(gs); ok {
		return level.ProactiveWar
	}
	return gs != nil && gs.Difficulty > 1
}

func aiWarThresholdForDifficulty(gs *state.GameState) int {
	if level, ok := aiDifficultyLevel(gs); ok && level.WarThreshold > 0 {
		return level.WarThreshold
	}
	threshold := aiWarThreshold
	if gs != nil && gs.Difficulty >= 3 {
		threshold -= 10
	}
	return threshold
}

func aiMinAttackPowerPercent(gs *state.GameState) int {
	if level, ok := aiDifficultyLevel(gs); ok && level.MinAttackPowerPercent > 0 {
		return level.MinAttackPowerPercent
	}
	return 115
}

func aiWarCadenceBase(gs *state.GameState) int {
	if level, ok := aiDifficultyLevel(gs); ok && level.WarCadenceTurns > 0 {
		return level.WarCadenceTurns
	}
	if gs != nil && gs.Difficulty >= 3 {
		return 7
	}
	return 10
}

func aiConfiguredWarCapacity(gs *state.GameState) (int, bool) {
	if level, ok := aiDifficultyLevel(gs); ok && level.MaxConcurrentWars > 0 {
		return level.MaxConcurrentWars, true
	}
	return 0, false
}

func aiPlayerTargetScoreBonus(gs *state.GameState) int {
	if level, ok := aiDifficultyLevel(gs); ok {
		return level.PlayerTargetScoreBonus
	}
	if gs != nil && gs.Difficulty >= 3 {
		return 8
	}
	return 0
}
