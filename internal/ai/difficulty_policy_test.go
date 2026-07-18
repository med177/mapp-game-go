package ai

import (
	"testing"

	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
)

func difficultyPolicyTestState(difficulty int) *state.GameState {
	return &state.GameState{
		Difficulty: difficulty,
		AIDifficultyPolicy: scenario.AIDifficultyPolicy{Levels: map[string]scenario.AIDifficultyLevel{
			"1": {PlanHorizonTurns: 4, PlanTargetRegionLimit: 3, PathSearchDepth: 5, PlanMoveBonusPercent: 70, WarThreshold: 82, MinAttackPowerPercent: 130, WarCadenceTurns: 12, MaxConcurrentWars: 1},
			"2": {PlanHorizonTurns: 6, PlanTargetRegionLimit: 4, PathSearchDepth: 8, PlanMoveBonusPercent: 100, ProactiveWar: true, WarThreshold: 70, MinAttackPowerPercent: 115, WarCadenceTurns: 10, MaxConcurrentWars: 1},
			"3": {PlanHorizonTurns: 9, PlanTargetRegionLimit: 5, PathSearchDepth: 12, PlanMoveBonusPercent: 125, ProactiveWar: true, WarThreshold: 65, MinAttackPowerPercent: 100, WarCadenceTurns: 7, MaxConcurrentWars: 2, PlayerTargetScoreBonus: 4},
		}},
	}
}

func TestDifficultyPolicyScalesPlanningQualityWithoutRuleBonus(t *testing.T) {
	easy := difficultyPolicyTestState(1)
	normal := difficultyPolicyTestState(2)
	hard := difficultyPolicyTestState(3)

	if aiProactiveWarEnabled(easy) {
		t.Fatal("kolay AI proaktif savaş açmamalı")
	}
	if aiPlanHorizonTurns(easy) != 4 || aiPlanHorizonTurns(normal) != 6 || aiPlanHorizonTurns(hard) != 9 {
		t.Fatalf("plan ufku zorlukla ölçeklenmedi: easy=%d normal=%d hard=%d", aiPlanHorizonTurns(easy), aiPlanHorizonTurns(normal), aiPlanHorizonTurns(hard))
	}
	if aiPathSearchDepth(easy) != 5 || aiPathSearchDepth(normal) != 8 || aiPathSearchDepth(hard) != 12 {
		t.Fatalf("rota arama derinliği zorlukla ölçeklenmedi: easy=%d normal=%d hard=%d", aiPathSearchDepth(easy), aiPathSearchDepth(normal), aiPathSearchDepth(hard))
	}
	if aiPlanMoveBonusPercent(easy) != 70 || aiPlanMoveBonusPercent(normal) != 100 || aiPlanMoveBonusPercent(hard) != 125 {
		t.Fatalf("objective koordinasyon ağırlığı zorlukla ölçeklenmedi: easy=%d normal=%d hard=%d", aiPlanMoveBonusPercent(easy), aiPlanMoveBonusPercent(normal), aiPlanMoveBonusPercent(hard))
	}
	if aiWarThresholdForDifficulty(hard) != 65 || aiMinAttackPowerPercent(hard) != 100 || aiPlayerTargetScoreBonus(hard) != 4 {
		t.Fatalf("zor risk politikası yanlış: threshold=%d ratio=%d player_bonus=%d", aiWarThresholdForDifficulty(hard), aiMinAttackPowerPercent(hard), aiPlayerTargetScoreBonus(hard))
	}
	if capacity, ok := aiConfiguredWarCapacity(hard); !ok || capacity != 2 {
		t.Fatalf("zor AI kontrollü ikinci cephe kapasitesi almalıydı: capacity=%d ok=%v", capacity, ok)
	}
}

func TestHardDifficultyExtendsDurablePlanHorizon(t *testing.T) {
	gs := strategicPlanTestState()
	gs.Difficulty = 3
	gs.AIDifficultyPolicy = difficultyPolicyTestState(3).AIDifficultyPolicy

	plan := ensureStrategicPlan(gs, "ottoman", buildStrategicContext(gs, "ottoman"))
	if plan == nil || plan.StartedTurn != 4 || plan.ReassessTurn != 13 {
		t.Fatalf("zor AI planı dokuz tur korunmalıydı: %+v", plan)
	}
}
