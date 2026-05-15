package game

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestApplyAIDifficultyStartBonusSkipsPlayerAndBuffsOnlyAI(t *testing.T) {
	g := &Game{
		gs: &state.GameState{
			Difficulty:      3,
			PlayerFactionID: "player",
			Factions: map[faction.FactionID]*faction.Faction{
				"player": {ID: "player", Gold: 100, Grain: 50},
				"ai_1":   {ID: "ai_1", Gold: 40, Grain: 20},
				"ai_2":   {ID: "ai_2", Gold: 10, Grain: 5},
			},
		},
	}

	g.applyAIDifficultyStartBonus()

	if g.gs.Factions["player"].Gold != 100 || g.gs.Factions["player"].Grain != 50 {
		t.Fatalf("oyuncu bonus almamali: %+v", g.gs.Factions["player"])
	}
	if g.gs.Factions["ai_1"].Gold != 340 || g.gs.Factions["ai_1"].Grain != 120 {
		t.Fatalf("ai_1 bonusu hatali: %+v", g.gs.Factions["ai_1"])
	}
	if g.gs.Factions["ai_2"].Gold != 310 || g.gs.Factions["ai_2"].Grain != 105 {
		t.Fatalf("ai_2 bonusu hatali: %+v", g.gs.Factions["ai_2"])
	}
}

func TestApplyAIDifficultyStartBonusDoesNothingWithoutPlayerSelection(t *testing.T) {
	g := &Game{
		gs: &state.GameState{
			Difficulty: 3,
			Factions: map[faction.FactionID]*faction.Faction{
				"ai_1": {ID: "ai_1", Gold: 40, Grain: 20},
			},
		},
	}

	g.applyAIDifficultyStartBonus()

	if g.gs.Factions["ai_1"].Gold != 40 || g.gs.Factions["ai_1"].Grain != 20 {
		t.Fatalf("oyuncu secimi yokken bonus uygulanmamali: %+v", g.gs.Factions["ai_1"])
	}
}
