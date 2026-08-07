package render

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestIncomeHUDValueRectIsInteractive(t *testing.T) {
	rect := incomeHUDValueRect()
	if !rect.Hit(rect.X+rect.W/2, rect.Y+rect.H/2) {
		t.Fatalf("gelir rakamı rect'i kendi merkezini kapsamalı: %+v", rect)
	}
	if rect.Hit(rect.X-1, rect.Y+rect.H/2) {
		t.Fatalf("gelir rakamı rect'i sol dış noktayı kapsamamalı: %+v", rect)
	}
}

func TestPlayerGoldEconomyStatusUsesNetChange(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions:        map[faction.FactionID]*faction.Faction{"player": {ID: "player"}},
		GoldEconomy: map[faction.FactionID]state.GoldEconomyStatus{
			"player": {FactionID: "player", Income: 3580, Upkeep: 52, NetChange: 3528},
		},
	}
	if got := calcPlayerIncome(gs); got != 3528 {
		t.Fatalf("HUD brüt geliri değil net değişimi göstermeli: got=%d", got)
	}
}
