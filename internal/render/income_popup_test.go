package render

import (
	"testing"

	"mapp-game-go/internal/economy"
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

func TestGrainHUDValueRectIsInteractive(t *testing.T) {
	rect := grainHUDValueRect()
	if !rect.Hit(rect.X+rect.W/2, rect.Y+rect.H/2) {
		t.Fatalf("tahıl rakamı rect'i kendi merkezini kapsamalı: %+v", rect)
	}
	if rect.Hit(rect.X-1, rect.Y+rect.H/2) {
		t.Fatalf("tahıl rakamı rect'i sol dış noktayı kapsamamalı: %+v", rect)
	}
	spiceRowY := 12.0 + 22.0
	if rect.Hit(rect.X+rect.W/2, spiceRowY+rect.H/2) {
		t.Fatalf("tahıl rakamı rect'i Baharat satırına taşmamalı: %+v", rect)
	}
}

func TestGrainEconomyPopupUsesCurrentSnapshot(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 100},
		},
		GrainEconomy: map[faction.FactionID]state.GrainEconomyStatus{
			"player": {
				FactionID:      "player",
				Production:     113,
				CivilianDemand: 42,
				ArmyUpkeep:     19,
				TotalDemand:    61,
				NetChange:      52,
				AutoExportSold: 7,
			},
		},
	}
	gs.SetMarketSellOffer("player", economy.GoodGrain, 23)
	lines := grainEconomyPopupLines(gs, playerGrainEconomyStatus(gs))
	if lines[0].value != 113 || lines[1].value != -42 || lines[2].value != -19 || lines[3].value != -61 || lines[4].value != 23 || lines[5].value != -7 {
		t.Fatalf("tahıl popup satırları snapshot ile eşleşmiyor: %+v", lines)
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
