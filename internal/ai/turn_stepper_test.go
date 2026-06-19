package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestTurnStepperMovesArmyOneStepPerCall(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Catholic},
		},
		Regions: map[world.RegionID]*world.Region{
			"a": {ID: "a", OwnerID: "ai_1", Neighbors: []world.RegionID{"b"}},
			"b": {ID: "b", Neighbors: []world.RegionID{"a", "c"}},
			"c": {ID: "c", Neighbors: []world.RegionID{"b"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"stack": {
				ID:            "stack",
				OwnerID:       "ai_1",
				RegionID:      "a",
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 10, Defense: 8, Morale: 50},
		},
	}

	stepper := NewTurnStepper(gs, "ai_1")

	step, done := stepper.Step()
	if done {
		t.Fatal("ilk adımda tur bitmemeliydi")
	}
	if step.Kind != TurnStepConquest || step.TargetRegion != "b" {
		t.Fatalf("ilk adım b bölgesine fetih olmalıydı, got=%+v", step)
	}
	if got := gs.Armies["stack"].RegionID; got != "b" {
		t.Fatalf("ordu ilk adım sonunda b'de olmalıydı, got=%s", got)
	}
	if got := gs.Armies["stack"].MovePoints; got != 1 {
		t.Fatalf("ilk adım sonrası 1 hareket puanı kalmalıydı, got=%d", got)
	}

	step, done = stepper.Step()
	if done {
		t.Fatal("ikinci adımda da tur bitmemeliydi")
	}
	if step.Kind != TurnStepConquest || step.TargetRegion != "c" {
		t.Fatalf("ikinci adım c bölgesine fetih olmalıydı, got=%+v", step)
	}
	if got := gs.Armies["stack"].RegionID; got != "c" {
		t.Fatalf("ordu ikinci adım sonunda c'de olmalıydı, got=%s", got)
	}
	if got := gs.Armies["stack"].MovePoints; got != 0 {
		t.Fatalf("ikinci adım sonrası hareket puanı bitmeliydi, got=%d", got)
	}

	step, done = stepper.Step()
	if !done {
		t.Fatal("hareket puanı bitince stepper tamamlanmalıydı")
	}
	if step.Kind != TurnStepComplete {
		t.Fatalf("tamamlanma adımı bekleniyordu, got=%+v", step)
	}
}
