package game

import (
	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func (g *Game) beginNavalContact(attacker, defender *army.Army, seaID, fromRegion world.RegionID, trigger state.NavalContactTrigger, moveBeforePrompt bool) bool {
	if g == nil || g.gs == nil || attacker == nil || defender == nil {
		return false
	}
	contact := g.gs.BeginNavalContact(attacker, defender, seaID, fromRegion, trigger)
	if contact == nil {
		return false
	}
	ai.ResolveNavalContactDecision(g.gs, contact)
	if moveBeforePrompt && trigger == state.NavalContactMovement && contact.AttackerArmyID == attacker.ID && contact.AttackerFromRegionID == fromRegion && attacker.RegionID == fromRegion {
		attacker.RegionID = seaID
		attacker.DockedRegionID = ""
		attacker.DockedSettlementID = ""
		if attacker.MovePoints > 0 {
			attacker.MovePoints--
		}
		contact.MovementConsumed = true
		g.renderer.MarkMapDirty()
	}
	if contact.PlayerArmyID != "" {
		g.presentPendingNavalContact()
	}
	return true
}

func (g *Game) presentPendingNavalContact() {
	if g == nil || g.gs == nil || g.renderer == nil {
		return
	}
	contact := g.gs.PendingNavalContact
	if contact == nil || contact.PlayerArmyID == "" || contact.Prompted {
		return
	}
	playerFleet := g.gs.Armies[contact.PlayerArmyID]
	if playerFleet == nil {
		g.gs.ClearNavalContact()
		return
	}
	opponentID := contact.AttackerArmyID
	if opponentID == playerFleet.ID {
		opponentID = contact.DefenderArmyID
	}
	opponent := g.gs.Armies[opponentID]
	seaName := string(contact.SeaRegionID)
	if sea := g.gs.Regions[contact.SeaRegionID]; sea != nil && sea.NameTR != "" {
		seaName = sea.NameTR
	}
	opponentName := "Düşman filosu"
	if opponent != nil {
		opponentName = g.factionNameTR(opponent.OwnerID) + " filosu"
	}
	defaultDecision := contact.DefenderDecision
	if contact.PlayerArmyID == contact.DefenderArmyID {
		defaultDecision = contact.AttackerDecision
	}
	message := opponentName + " " + seaName + " denizinde tespit edildi. Savaş yalnız iki taraf da Çatış seçerse başlayacak. Karşı tarafın varsayılan tutumu: " + navalContactDecisionLabel(defaultDecision) + "."
	contact.Prompted = true
	g.renderer.ShowThreeChoiceDialogWithThirdEnabled(
		"Düşman Filo Tespit Edildi",
		message,
		"Çatış",
		"Geri Çekil",
		"Pozisyonu Koru",
		render.InputAction{Kind: render.ActionResolveNavalContact, ChoiceIndex: 0},
		render.InputAction{Kind: render.ActionResolveNavalContact, ChoiceIndex: 1},
		render.InputAction{Kind: render.ActionResolveNavalContact, ChoiceIndex: 2},
		playerFleet.MovePoints > 0 && g.gs.NavalContactRetreatRegion(playerFleet, contact.AttackerFromRegionID) != "",
	)
}

func navalContactDecisionLabel(decision state.NavalContactDecision) string {
	switch decision {
	case state.NavalContactClash:
		return "Çatış"
	case state.NavalContactWithdraw:
		return "Geri çekil"
	default:
		return "Pozisyonu koru"
	}
}

func navalContactDecisionFromChoice(choice int) state.NavalContactDecision {
	switch choice {
	case 0:
		return state.NavalContactClash
	case 1:
		return state.NavalContactWithdraw
	default:
		return state.NavalContactHold
	}
}

func (g *Game) resolveNavalContactChoice(choice int) {
	if g == nil || g.gs == nil {
		return
	}
	contact := g.gs.PendingNavalContact
	if contact == nil {
		return
	}
	if !g.gs.NavalContactDecisionForPlayer(contact, navalContactDecisionFromChoice(choice)) {
		return
	}
	if contact.AttackerDecision == state.NavalContactUndecided || contact.DefenderDecision == state.NavalContactUndecided {
		return
	}
	attacker := g.gs.Armies[contact.AttackerArmyID]
	defender := g.gs.Armies[contact.DefenderArmyID]
	if attacker == nil || defender == nil {
		g.gs.ClearNavalContact()
		return
	}
	if g.gs.NavalContactBothClash(contact) {
		playerIsAttacker := attacker.OwnerID == string(g.gs.PlayerFactionID)
		movementConsumed := contact.MovementConsumed
		g.gs.ClearNavalContact()
		if playerIsAttacker {
			if g.renderer.ShowNavalContactBattlePlan(attacker.ID, defender.ID, contact.SeaRegionID, movementConsumed) {
				return
			}
			g.moveArmyToSettlementWithStanceAndContactResolved(attacker.ID, contact.SeaRegionID, "", combat.BattleStanceBalanced, true, true, movementConsumed)
			return
		}
		step := ai.ResolveNavalContactBattle(g.gs, attacker.ID, contact.SeaRegionID)
		if step.Message != "" {
			g.renderer.ShowCombatResult(step.Message)
			g.renderer.AddEvent("[DENİZ TEMASI] " + step.Message)
		}
		return
	}

	g.resolveNavalContactWithoutBattle(contact, attacker, defender)
	g.gs.ClearNavalContact()
	g.renderer.MarkMapDirty()
	g.renderer.ShowCombatResult("Deniz teması çatışmaya dönüşmeden sona erdi.")
}

func (g *Game) resolveNavalContactWithoutBattle(contact *state.NavalContact, attacker, defender *army.Army) {
	if g == nil || g.gs == nil || contact == nil {
		return
	}
	if contact.AttackerDecision == state.NavalContactWithdraw {
		previousLocation := attacker.LocationID()
		retreat := navalContactRetreatRegion(g.gs, attacker, contact.AttackerFromRegionID)
		if retreat != "" {
			attacker.RegionID = retreat
			attacker.DockedRegionID = ""
			attacker.DockedSettlementID = ""
			attacker.MovePoints = max(0, attacker.MovePoints-state.NavalContactWithdrawMovementCost)
			g.gs.ClearNavalMissionAfterRelocation(attacker, previousLocation)
		}
	}
	if contact.DefenderDecision == state.NavalContactWithdraw {
		previousLocation := defender.LocationID()
		if retreat := navalContactRetreatRegion(g.gs, defender, contact.AttackerFromRegionID); retreat != "" {
			defender.RegionID = retreat
			defender.DockedRegionID = ""
			defender.DockedSettlementID = ""
			defender.MovePoints = max(0, defender.MovePoints-state.NavalContactWithdrawMovementCost)
			g.gs.ClearNavalMissionAfterRelocation(defender, previousLocation)
		}
	}
}

func navalContactRetreatRegion(gs *state.GameState, fleet *army.Army, excludedRegions ...world.RegionID) world.RegionID {
	if gs == nil {
		return ""
	}
	return gs.NavalContactRetreatRegion(fleet, excludedRegions...)
}
