package game

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestShouldNotifyPlayerTechnologyCompletion(t *testing.T) {
	gs := &state.GameState{PlayerFactionID: faction.FactionID("player")}

	if !shouldNotifyPlayerTechnologyCompletion(gs, "player") {
		t.Fatal("oyuncunun teknoloji tamamlanması bildirilmeli")
	}
	if shouldNotifyPlayerTechnologyCompletion(gs, "ai") {
		t.Fatal("AI devletinin teknoloji tamamlanması oyuncuya bildirilmemeli")
	}
}
