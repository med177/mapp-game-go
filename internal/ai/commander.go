package ai

import (
	"math/rand"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

// recordCommanderBattle AI savaşlarında gerçek orduların komutanlarını ilerletir.
func recordCommanderBattle(gs *state.GameState, attacker, defender *army.Army, defenderIDs []army.ArmyID, attackerWon bool) {
	if attacker != nil {
		captorID := ""
		if !attackerWon && defender != nil {
			captorID = defender.OwnerID
		}
		recordCommanderProgress(gs, attacker, attackerWon, captorID)
	}
	if len(defenderIDs) > 0 && gs != nil {
		captorID := ""
		if attacker != nil {
			captorID = attacker.OwnerID
		}
		for _, defenderID := range defenderIDs {
			if source := gs.Armies[defenderID]; source != nil {
				recordCommanderProgress(gs, source, !attackerWon, captorID)
			}
		}
		return
	}
	if defender != nil {
		captorID := ""
		if attackerWon && attacker != nil {
			captorID = attacker.OwnerID
		}
		recordCommanderProgress(gs, defender, !attackerWon, captorID)
	}
}

func recordCommanderProgress(gs *state.GameState, currentArmy *army.Army, won bool, captorID string) {
	if currentArmy == nil || currentArmy.Commander == nil {
		return
	}
	commander := currentArmy.Commander
	currentArmy.RecordBattle(won)
	if !won && gs != nil {
		captureChance := gs.CommanderCaptureChanceOnDefeat
		if captureChance > 0 && rand.Intn(100) < captureChance && gs.CaptureCommanderAtRegion(commander.ID, currentArmy.RegionID) {
			resolveAICaptiveDecision(gs, commander.ID, captorID)
		} else {
			gs.InjureCommander(commander.ID, gs.CommanderInjuryDuration(commander), "Yaralı")
		}
	}
}

func resolveAICaptiveDecision(gs *state.GameState, commanderID, captorID string) {
	if gs == nil || commanderID == "" || captorID == "" {
		return
	}
	commander := gs.Commanders[commanderID]
	if commander == nil {
		return
	}
	captor := faction.FactionID(captorID)
	captiveOwner := faction.FactionID(commander.OwnerID)
	score := diplomacy.RelationScore(gs, captor, captiveOwner)
	outcome := 1
	ransomAmount := gs.CommanderRansomGold
	if ransomAmount <= 0 {
		ransomAmount = state.DefaultCommanderRansomGold
	}
	if payer := gs.Factions[captiveOwner]; payer != nil && payer.Gold >= ransomAmount && score >= 0 && rand.Intn(100) < 30 {
		outcome = 3
	} else if score >= 25 {
		outcome = 0
	} else if diplomacy.IsWar(gs, captor, captiveOwner) && rand.Intn(100) < 35 {
		outcome = 2
	}
	switch outcome {
	case 3:
		if gs.PayCommanderRansom(commanderID, captor) {
			diplomacy.AddRelationScoreBoth(gs, captor, captiveOwner, state.CommanderRansomRelationDelta)
			break
		}
		fallthrough
	case 0:
		gs.ReleaseCommanderCaptivity(commanderID)
		diplomacy.AddRelationScoreBoth(gs, captor, captiveOwner, state.CommanderReleaseRelationDelta)
	case 1:
		if !commander.CaptivityKeepPenaltyApplied {
			diplomacy.AddRelationScoreBoth(gs, captor, captiveOwner, state.CommanderKeepRelationDelta)
			commander.CaptivityKeepPenaltyApplied = true
		}
	case 2:
		gs.DismissCommander(commanderID)
		diplomacy.AddRelationScoreBoth(gs, captor, captiveOwner, state.CommanderExecuteRelationDelta)
	}
}
