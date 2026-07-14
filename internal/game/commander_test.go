package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
)

func TestRecordCommanderBattleUsesRealCombinedDefenders(t *testing.T) {
	attackerCommander := army.NewCommander("atk_cmd", "Saldıran")
	defenderOneCommander := army.NewCommander("def_1_cmd", "Savunan Bir")
	defenderTwoCommander := army.NewCommander("def_2_cmd", "Savunan İki")
	gs := &state.GameState{Armies: map[army.ArmyID]*army.Army{
		"atk":  {ID: "atk", Commander: attackerCommander},
		"def1": {ID: "def1", Commander: defenderOneCommander},
		"def2": {ID: "def2", Commander: defenderTwoCommander},
	}}
	g := &Game{gs: gs}
	combined := &army.Army{Units: []army.Unit{{TypeID: "inf"}}}

	g.recordCommanderBattle(gs.Armies["atk"], combined, []army.ArmyID{"def1", "def2"}, true)

	if attackerCommander.Victories != 1 || attackerCommander.Experience != army.CommanderWinXP {
		t.Fatalf("saldıran komutan ilerlemesi yanlış: %+v", attackerCommander)
	}
	for _, id := range []army.ArmyID{"def1", "def2"} {
		commander := gs.Armies[id].Commander
		if commander.Battles != 1 || commander.Victories != 0 || commander.Experience != army.CommanderLossXP {
			t.Fatalf("birleşik savunucu %s ilerlemesi yanlış: %+v", id, commander)
		}
	}
	if len(g.lastCommanderProgress) != 3 {
		t.Fatalf("savaş raporu için üç komutan ilerlemesi bekleniyordu: %+v", g.lastCommanderProgress)
	}
	if g.lastCommanderProgress[0].Name != "Saldıran" || g.lastCommanderProgress[0].XPGained != army.CommanderWinXP {
		t.Fatalf("saldıran komutan rapor ilerlemesi yanlış: %+v", g.lastCommanderProgress[0])
	}
}
