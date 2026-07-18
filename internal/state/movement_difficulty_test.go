package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
)

func TestFairMovementPolicyRemovesHardAIMovementCheat(t *testing.T) {
	gs := &GameState{
		Difficulty:      3,
		PlayerFactionID: "player",
		Month:           6,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ai":     {ID: "ai"},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, MovementPoints: 2},
		},
		AIDifficultyPolicy: scenario.AIDifficultyPolicy{FairMovement: true},
	}
	aiArmy := &army.Army{ID: "ai_army", OwnerID: "ai", Units: []army.Unit{{TypeID: "inf"}}}
	playerArmy := &army.Army{ID: "player_army", OwnerID: "player", Units: []army.Unit{{TypeID: "inf"}}}

	if got := gs.ArmyMaxMovePoints(aiArmy); got != 2 {
		t.Fatalf("fair movement politikasında zor AI oyuncuyla aynı hızda olmalıydı: %d", got)
	}
	if got := gs.ArmyMaxMovePoints(playerArmy); got != 2 {
		t.Fatalf("oyuncu hareketi değişmemeliydi: %d", got)
	}

	gs.AIDifficultyPolicy = scenario.AIDifficultyPolicy{}
	if got := gs.ArmyMaxMovePoints(aiArmy); got != 3 {
		t.Fatalf("politikasız legacy senaryoda eski zor AI +1 hareketi korunmalıydı: %d", got)
	}
}
