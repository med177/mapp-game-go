package game

import (
	"math/rand"

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
		captorID := ""
		if !attackerWon && defender != nil {
			captorID = defender.OwnerID
		}
		g.recordCommanderProgress("Saldıran", attacker, attackerWon, captorID)
	}
	if len(defenderIDs) > 0 && g != nil && g.gs != nil {
		captorID := ""
		if attacker != nil {
			captorID = attacker.OwnerID
		}
		for _, defenderID := range defenderIDs {
			if source := g.gs.Armies[defenderID]; source != nil {
				g.recordCommanderProgress("Savunan", source, !attackerWon, captorID)
			}
		}
		return
	}
	if defender != nil {
		captorID := ""
		if attackerWon && attacker != nil {
			captorID = attacker.OwnerID
		}
		g.recordCommanderProgress("Savunan", defender, !attackerWon, captorID)
	}
}

func (g *Game) recordCommanderProgress(side string, currentArmy *army.Army, won bool, captorID string) {
	if g == nil || currentArmy == nil || currentArmy.Commander == nil {
		return
	}
	commander := currentArmy.Commander
	progress := currentArmy.RecordBattle(won)
	if !won && g.gs != nil {
		captureChance := g.gs.CommanderCaptureChanceOnDefeat
		if captureChance > 0 && rand.Intn(100) < captureChance && g.gs.CaptureCommanderAtRegion(commander.ID, currentArmy.RegionID) {
			if captorID == string(g.gs.PlayerFactionID) {
				g.queueCapturedCommanderDecision(commander.ID, commander.CaptiveAtSettlementID, captorID, true)
			}
		} else {
			g.gs.InjureCommander(commander.ID, g.gs.CommanderInjuryDuration(commander), "Yaralı")
		}
	}
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
