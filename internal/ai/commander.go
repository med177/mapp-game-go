package ai

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
)

// recordCommanderBattle AI savaşlarında gerçek orduların komutanlarını ilerletir.
func recordCommanderBattle(gs *state.GameState, attacker, defender *army.Army, defenderIDs []army.ArmyID, attackerWon bool) {
	if attacker != nil {
		attacker.RecordBattle(attackerWon)
	}
	if len(defenderIDs) > 0 && gs != nil {
		for _, defenderID := range defenderIDs {
			if source := gs.Armies[defenderID]; source != nil {
				source.RecordBattle(!attackerWon)
			}
		}
		return
	}
	if defender != nil {
		defender.RecordBattle(!attackerWon)
	}
}
