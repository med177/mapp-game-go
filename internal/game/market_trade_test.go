package game

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestCanPlayerOneTimeTradeWithUsesOpenMarketButExcludesEnemies(t *testing.T) {
	playerID := faction.FactionID("player")
	openMarketID := faction.FactionID("open_market")
	enemyID := faction.FactionID("enemy")
	gs := &state.GameState{
		PlayerFactionID: playerID,
		Factions: map[faction.FactionID]*faction.Faction{
			playerID:     {ID: playerID},
			openMarketID: {ID: openMarketID},
			enemyID:      {ID: enemyID},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(playerID, enemyID): {FactionA: playerID, FactionB: enemyID, Stance: faction.StanceWar},
		},
	}

	if !canPlayerOneTimeTradeWith(gs, openMarketID) {
		t.Fatal("aktif rota veya ilişki kaydı olmadan barıştaki açık pazar satıcısıyla işlem yapılmalı")
	}
	if canPlayerOneTimeTradeWith(gs, enemyID) {
		t.Fatal("savaştaki devlet açık pazarda işlem ortağı olmamalı")
	}
}
