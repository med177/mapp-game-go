package ai

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const aiLandContactRetreatThresholdPercent = 135

// ResolveLandContactDecision, temas oluştuğu anda AI tarafının tutumunu
// belirler. Belirgin biçimde zayıf ve güvenli geri çekilme yolu olan AI,
// savaşa girmek yerine geri çekilmeyi seçer; diğer durumlarda çatışır.
func ResolveLandContactDecision(gs *state.GameState, contact *state.LandContact) {
	if gs == nil || contact == nil {
		return
	}
	setDecision := func(armyID army.ArmyID, opponentID army.ArmyID, decision *state.LandContactDecision) {
		current := gs.Armies[armyID]
		opponent := gs.Armies[opponentID]
		if current == nil || opponent == nil || decision == nil || current.OwnerID == string(gs.PlayerFactionID) {
			return
		}
		chosen := state.LandContactClash
		// Pusuya giren hareketli tarafın geri çekilme hakkı yoktur; temas
		// otomatik olarak çatışma kararına gider.
		if contact.AmbushArmyID != "" && armyID == contact.AttackerArmyID {
			*decision = state.LandContactClash
			return
		}
		canRetreat := current.MovePoints > 0 || contact.AmbushArmyID == current.ID
		if canRetreat && gs.LandContactRetreatRegion(current, contact.AttackerFromRegionID) != "" && aiLandContactShouldRetreat(gs, current, opponent) {
			chosen = state.LandContactWithdraw
		}
		*decision = chosen
	}

	setDecision(contact.AttackerArmyID, contact.DefenderArmyID, &contact.AttackerDecision)
	setDecision(contact.DefenderArmyID, contact.AttackerArmyID, &contact.DefenderDecision)
}

func aiLandContactShouldRetreat(gs *state.GameState, current, opponent *army.Army) bool {
	if gs == nil || current == nil || opponent == nil {
		return false
	}
	currentPower := current.TotalStrength(gs.UnitTypes)
	opponentPower := opponent.TotalStrength(gs.UnitTypes)
	if currentPower <= 0 || opponentPower <= 0 {
		return false
	}
	return opponentPower*100 >= currentPower*aiLandContactRetreatThresholdPercent
}

// ResolveLandContactBattle, oyuncu kara temasında Çatış seçtiğinde AI
// saldırısının iki tarafın karar verdiği normal hareket/savaş hattına
// devam etmesini sağlar.
func ResolveLandContactBattle(gs *state.GameState, attackerID army.ArmyID, target world.RegionID) TurnStep {
	if gs == nil {
		return TurnStep{}
	}
	attacker := gs.Armies[attackerID]
	if attacker == nil {
		return TurnStep{}
	}
	outcome := executeMoveWithNavalPatrolAndContact(gs, attacker, target, faction.FactionID(attacker.OwnerID), false, true)
	return outcome.step
}

// ResolveLandContactWithoutBattle, iki AI ordusu aynı hedefte çatışmayı
// seçmediğinde temasın hareket karşılığını uygular. AI saldıranı henüz hedefe
// girmediği için geri çekilmesi kaynak konumunda kalır; savunmacı ise güvenli
// komşu kara bölgesine çekilebilir.
func ResolveLandContactWithoutBattle(gs *state.GameState, contact *state.LandContact, defender *army.Army) {
	if gs == nil || contact == nil {
		return
	}
	if contact.AttackerDecision == state.LandContactWithdraw {
		if attacker := gs.Armies[contact.AttackerArmyID]; attacker != nil && attacker.RegionID == contact.LandRegionID {
			attacker.RegionID = contact.AttackerFromRegionID
			attacker.DockedRegionID = ""
			attacker.DockedSettlementID = ""
		}
	}
	if contact.DefenderDecision == state.LandContactWithdraw && defender != nil {
		if retreat := gs.LandContactRetreatRegion(defender, contact.AttackerFromRegionID); retreat != "" {
			defender.RegionID = retreat
			defender.DockedRegionID = ""
			defender.DockedSettlementID = ""
			defender.MovePoints = maxInt(0, defender.MovePoints-1)
		}
	}
	if ambusher := gs.Armies[contact.AmbushArmyID]; ambusher != nil {
		ambusher.InAmbush = false
	}
}
