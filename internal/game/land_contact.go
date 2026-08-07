package game

import (
	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func (g *Game) beginLandContact(attacker, defender *army.Army, landID, fromRegion world.RegionID, trigger state.LandContactTrigger, moveBeforePrompt bool) bool {
	if g == nil || g.gs == nil || attacker == nil || defender == nil {
		return false
	}
	contact := g.gs.BeginLandContact(attacker, defender, landID, fromRegion, trigger)
	if contact == nil {
		return false
	}
	ai.ResolveLandContactDecision(g.gs, contact)
	if moveBeforePrompt && contact.AttackerArmyID == attacker.ID && contact.AttackerFromRegionID == fromRegion && attacker.RegionID == fromRegion {
		attacker.RegionID = landID
		attacker.DockedRegionID = ""
		attacker.DockedSettlementID = ""
		if attacker.MovePoints > 0 {
			attacker.MovePoints--
		}
		contact.MovementConsumed = true
		g.renderer.MarkMapDirty()
	}
	if contact.PlayerArmyID != "" {
		g.presentPendingLandContact()
	}
	return true
}

func (g *Game) presentPendingLandContact() {
	if g == nil || g.gs == nil || g.renderer == nil {
		return
	}
	contact := g.gs.PendingLandContact
	if contact == nil || contact.PlayerArmyID == "" || contact.Prompted {
		return
	}
	playerArmy := g.gs.Armies[contact.PlayerArmyID]
	if playerArmy == nil {
		g.gs.ClearLandContact()
		return
	}
	opponentID := contact.AttackerArmyID
	if opponentID == playerArmy.ID {
		opponentID = contact.DefenderArmyID
	}
	opponent := g.gs.Armies[opponentID]
	landName := string(contact.LandRegionID)
	if region := g.gs.Regions[contact.LandRegionID]; region != nil && region.NameTR != "" {
		landName = region.NameTR
	}
	opponentName := "Düşman ordusu"
	if opponent != nil {
		opponentName = g.factionNameTR(opponent.OwnerID) + " ordusu"
	}
	defaultDecision := contact.DefenderDecision
	if contact.PlayerArmyID == contact.DefenderArmyID {
		defaultDecision = contact.AttackerDecision
	}
	message := opponentName + " " + landName + " bölgesinde temas etti. Taraflardan biri Çatış seçip diğeri geri çekilmezse savaş başlayacak. Karşı tarafın varsayılan tutumu: " + landContactDecisionLabel(defaultDecision) + "."
	thirdEnabled := g.gs.LandContactHasSafeWithdrawal(contact)
	if contact.AmbushArmyID != "" {
		message = opponentName + " " + landName + " bölgesine girdi ve pusuya düştü. Düşman ordusu geri çekilemez; pusu tarafı Çatış veya Geri Çekil seçebilir."
		thirdEnabled = contact.PlayerArmyID == contact.DefenderArmyID && g.gs.LandContactHasSafeWithdrawal(contact)
	}
	contact.Prompted = true
	g.renderer.ShowThreeChoiceDialogWithThirdEnabled(
		"Düşman Ordusu Tespit Edildi",
		message,
		"Çatış",
		"Geri Çekil",
		"Pozisyonu Koru",
		render.InputAction{Kind: render.ActionResolveLandContact, ChoiceIndex: 0},
		render.InputAction{Kind: render.ActionResolveLandContact, ChoiceIndex: 1},
		render.InputAction{Kind: render.ActionResolveLandContact, ChoiceIndex: 2},
		thirdEnabled,
	)
	if contact.AmbushArmyID != "" {
		g.renderer.SetPendingContactHoldDisabled()
	}
}

func landContactDecisionLabel(decision state.LandContactDecision) string {
	switch decision {
	case state.LandContactClash:
		return "Çatış"
	case state.LandContactWithdraw:
		return "Geri çekil"
	default:
		return "Pozisyonu koru"
	}
}

func landContactDecisionFromChoice(choice int) state.LandContactDecision {
	switch choice {
	case 0:
		return state.LandContactClash
	case 1:
		return state.LandContactWithdraw
	default:
		return state.LandContactHold
	}
}

func (g *Game) resolveLandContactChoice(choice int) {
	if g == nil || g.gs == nil {
		return
	}
	contact := g.gs.PendingLandContact
	if contact == nil {
		return
	}
	if !g.gs.LandContactDecisionForPlayer(contact, landContactDecisionFromChoice(choice)) {
		return
	}
	if contact.AttackerDecision == state.LandContactUndecided || contact.DefenderDecision == state.LandContactUndecided {
		return
	}
	attacker := g.gs.Armies[contact.AttackerArmyID]
	defender := g.gs.Armies[contact.DefenderArmyID]
	if attacker == nil || defender == nil {
		g.gs.ClearLandContact()
		return
	}
	if g.gs.LandContactWillClash(contact) {
		playerIsAttacker := attacker.OwnerID == string(g.gs.PlayerFactionID)
		movementConsumed := contact.MovementConsumed
		attackerHolding := contact.AttackerDecision == state.LandContactHold
		defenderHolding := contact.DefenderDecision == state.LandContactHold
		g.gs.ClearLandContact()
		if playerIsAttacker {
			if contact.AmbushArmyID == "" {
				if target := g.gs.Regions[contact.LandRegionID]; target != nil && target.IsFortified() {
					if g.renderer.ShowLandContactSiegeDecision(attacker.ID, contact.LandRegionID) {
						return
					}
				}
			}
			if g.renderer.ShowLandContactBattlePlan(attacker.ID, defender.ID, contact.LandRegionID, attackerHolding, defenderHolding) {
				return
			}
			g.moveArmyToSettlementWithStanceAndContactResolved(attacker.ID, contact.LandRegionID, "", combat.BattleStanceBalanced, false, true, movementConsumed, attackerHolding, defenderHolding)
			return
		}
		step := ai.ResolveLandContactBattle(g.gs, attacker.ID, contact.LandRegionID, movementConsumed, attackerHolding, defenderHolding)
		if step.Message != "" {
			g.renderer.ShowCombatResult(step.Message)
			g.renderer.AddEvent("[KARA TEMASI] " + step.Message)
		}
		return
	}

	ai.ResolveLandContactWithoutBattle(g.gs, contact, defender)
	if ambusher := g.gs.Armies[contact.AmbushArmyID]; ambusher != nil {
		ambusher.InAmbush = false
	}
	g.gs.ClearLandContact()
	g.renderer.MarkMapDirty()
	g.renderer.ShowCombatResult("Kara teması çatışmaya dönüşmeden sona erdi.")
}
