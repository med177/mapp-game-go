package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestNewMarketOrderPreparationIncludesPlayerForEconomyOnlyMode(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID:         "player",
		AIControlsPlayerEconomy: true,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"north":  {ID: "north"},
		},
	}

	preparation := NewMarketOrderPreparation(gs, nil)
	if len(preparation.factionIDs) != 2 {
		t.Fatalf("ekonomi-only modunda pazar hazırlığı oyuncuyu içermiyor: got=%v", preparation.factionIDs)
	}
}

func TestPlayerEconomyOnlyTurnSkipsArmySteps(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID:         "player",
		AIControlsPlayerEconomy: true,
		Armies: map[army.ArmyID]*army.Army{
			"army": {ID: "army", OwnerID: "player", MovePoints: 2},
		},
	}
	stepper := &TurnStepper{gs: gs, fid: "player", preludeDone: true}

	step, done := stepper.Step()
	if !done || step.Kind != TurnStepComplete {
		t.Fatalf("ekonomi-only modunda askerî adım atlandıktan sonra tur tamamlanmalı: step=%+v done=%v", step, done)
	}
	if got := gs.Armies["army"].MovePoints; got != 2 {
		t.Fatalf("ekonomi-only modunda oyuncu ordusunun hareket puanı değişti: got=%d", got)
	}
}
