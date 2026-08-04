package ai

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const aiNavalContactRetreatThresholdPercent = 125

// ResolveNavalContactDecision, temas oluştuğu anda AI taraflarının tutumunu
// belirler. Güç farkı belirginse ve filo hâlâ hareket edebiliyorsa AI geri
// çekilir; aksi halde görevinin normal temas tutumu korunur.
func ResolveNavalContactDecision(gs *state.GameState, contact *state.NavalContact) {
	if gs == nil || contact == nil {
		return
	}
	setDecision := func(fleetID army.ArmyID, opponentID army.ArmyID, decision *state.NavalContactDecision, excludedRegions ...world.RegionID) {
		fleet := gs.Armies[fleetID]
		opponent := gs.Armies[opponentID]
		if fleet == nil || opponent == nil || decision == nil || fleet.OwnerID == string(gs.PlayerFactionID) {
			return
		}
		chosen := aiNavalContactBaseDecision(fleet)
		if fleet.MovePoints > 0 && gs.NavalContactRetreatRegionAt(fleet, contact.SeaRegionID, excludedRegions...) != "" && aiNavalContactShouldRetreat(gs, fleet, opponent) {
			chosen = state.NavalContactWithdraw
		}
		*decision = chosen
	}

	setDecision(contact.AttackerArmyID, contact.DefenderArmyID, &contact.AttackerDecision, contact.AttackerFromRegionID)
	setDecision(contact.DefenderArmyID, contact.AttackerArmyID, &contact.DefenderDecision, contact.AttackerFromRegionID)
}

// ResolveAIOnlyNavalContact immediately resolves a contact that contains no
// player fleet. Such contacts must never remain in PendingNavalContact: the
// game loop reserves that field for the player's decision modal.
func ResolveAIOnlyNavalContact(gs *state.GameState, contact *state.NavalContact) TurnStep {
	if gs == nil || contact == nil || contact.PlayerArmyID != "" {
		return TurnStep{}
	}
	attacker := gs.Armies[contact.AttackerArmyID]
	defender := gs.Armies[contact.DefenderArmyID]
	if attacker == nil || defender == nil {
		gs.ClearNavalContact()
		return TurnStep{}
	}
	if !gs.NavalContactBothClash(contact) {
		resolveAINavalContactWithoutBattle(gs, contact, attacker, defender)
		gs.ClearNavalContact()
		return TurnStep{
			FactionID:    faction.FactionID(attacker.OwnerID),
			Kind:         TurnStepMove,
			ArmyID:       attacker.ID,
			FromRegion:   contact.AttackerFromRegionID,
			TargetRegion: contact.SeaRegionID,
			FocusRegion:  contact.SeaRegionID,
			Message:      "Deniz teması çatışmaya dönüşmeden sona erdi.",
		}
	}

	// ResolveNavalContactBattle expects the contact to have already been
	// accepted and only uses the fleets' current locations. Clear the transient
	// prompt before resolving so a battle cannot leave a stale modal state.
	gs.ClearNavalContact()
	return ResolveNavalContactBattle(gs, attacker.ID, contact.SeaRegionID)
}

func aiNavalContactBaseDecision(fleet *army.Army) state.NavalContactDecision {
	if fleet != nil && fleet.NavalMission != nil {
		switch fleet.NavalMission.Kind {
		case army.NavalMissionPatrol:
			return state.NavalContactClash
		case army.NavalMissionBlockade, army.NavalMissionEscort, army.NavalMissionTransport:
			return state.NavalContactHold
		}
	}
	return state.NavalContactClash
}

func aiNavalContactShouldRetreat(gs *state.GameState, fleet, opponent *army.Army) bool {
	if gs == nil || fleet == nil || opponent == nil {
		return false
	}
	fleetPower := aiNavalContactPower(gs, fleet)
	opponentPower := aiNavalContactPower(gs, opponent)
	if fleetPower <= 0 || opponentPower <= 0 {
		return false
	}
	return opponentPower*100 > fleetPower*aiNavalContactRetreatThresholdPercent
}

func aiNavalContactPower(gs *state.GameState, fleet *army.Army) int {
	if gs == nil || fleet == nil {
		return 0
	}
	power := maxInt(aiEffectiveNavalPower(gs, fleet, true), aiEffectiveNavalPower(gs, fleet, false))
	if power <= 0 {
		power = len(fleet.Units)
	}
	return power
}

// resolveAINavalContactWithoutBattle, iki AI filosu çatışmayı kabul etmediğinde
// geri çekilmeyi gerçek bir deniz hareketine çevirir.
func resolveAINavalContactWithoutBattle(gs *state.GameState, contact *state.NavalContact, attacker, defender *army.Army) {
	if gs == nil || contact == nil {
		return
	}
	if contact.AttackerDecision == state.NavalContactWithdraw {
		retreat := aiNavalContactRetreatRegion(gs, attacker, contact.AttackerFromRegionID)
		if retreat != "" && attacker != nil {
			attacker.RegionID = retreat
			attacker.DockedRegionID = ""
			attacker.DockedSettlementID = ""
			attacker.MovePoints = maxInt(0, attacker.MovePoints-state.NavalContactWithdrawMovementCost)
		}
	}
	if contact.DefenderDecision == state.NavalContactWithdraw && defender != nil {
		if retreat := aiNavalContactRetreatRegion(gs, defender, contact.AttackerFromRegionID); retreat != "" {
			defender.RegionID = retreat
			defender.DockedRegionID = ""
			defender.DockedSettlementID = ""
			defender.MovePoints = maxInt(0, defender.MovePoints-state.NavalContactWithdrawMovementCost)
		}
	}
}

func aiNavalContactRetreatRegion(gs *state.GameState, fleet *army.Army, excludedRegions ...world.RegionID) world.RegionID {
	if gs == nil {
		return ""
	}
	return gs.NavalContactRetreatRegion(fleet, excludedRegions...)
}
