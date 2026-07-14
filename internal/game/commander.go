package game

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/render"
)

// recordCommanderBattle savaş sonucunu gerçek orduların komutanlarına aktarır.
// DefenderIDs birleşik savunma kullanıldığında sanal ordunun yerine kaynak
// orduları gösterir; böylece XP gerçek komutanlarda kalır.
func (g *Game) recordCommanderBattle(attacker *army.Army, defender *army.Army, defenderIDs []army.ArmyID, attackerWon bool) {
	if g != nil {
		g.lastCommanderProgress = nil
	}
	if attacker != nil {
		g.recordCommanderProgress("Saldıran", attacker, attackerWon)
	}
	if len(defenderIDs) > 0 && g != nil && g.gs != nil {
		for _, defenderID := range defenderIDs {
			if source := g.gs.Armies[defenderID]; source != nil {
				g.recordCommanderProgress("Savunan", source, !attackerWon)
			}
		}
		return
	}
	if defender != nil {
		g.recordCommanderProgress("Savunan", defender, !attackerWon)
	}
}

func (g *Game) recordCommanderProgress(side string, currentArmy *army.Army, won bool) {
	if g == nil || currentArmy == nil || currentArmy.Commander == nil {
		return
	}
	commander := currentArmy.Commander
	progress := currentArmy.RecordBattle(won)
	entry := render.BattleReportCommanderProgress{
		SideLabel:     side,
		Name:          commander.Name,
		XPGained:      progress.XPGained,
		PreviousLevel: progress.PreviousLevel,
		CurrentLevel:  progress.CurrentLevel,
		NewTraits:     make([]string, 0, len(progress.NewTraits)),
	}
	for _, trait := range progress.NewTraits {
		entry.NewTraits = append(entry.NewTraits, army.TraitLabelTR(trait))
	}
	g.lastCommanderProgress = append(g.lastCommanderProgress, entry)
}
