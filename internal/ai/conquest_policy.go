package ai

import (
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const aiVassalResistanceAggressiveness = 65

// TryResolvePostWarVassalization 1300 senaryosunda bir AI devletinin son
// toprağında yenilen hedefi vassal bırakıp bırakmayacağını değerlendirir ve
// onaylanan kararı standart diplomasi executor'ıyla uygular.
func TryResolvePostWarVassalization(gs *state.GameState, attackerID faction.FactionID, targetRegion *world.Region) diplomacy.Result {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || targetRegion == nil || attackerID == "" || attackerID == gs.PlayerFactionID {
		return diplomacy.Result{}
	}
	defenderID := faction.FactionID(targetRegion.OwnerID)
	if defenderID == "" || defenderID == attackerID || len(gs.LandRegionsOwnedBy(defenderID)) != 1 {
		return diplomacy.Result{}
	}
	plan := gs.AIPlans[attackerID]
	if plan == nil || !plan.AllowVassalization || plan.TargetFactionID != defenderID {
		return diplomacy.Result{}
	}
	for _, annexRegionID := range plan.AnnexRegionIDs {
		if annexRegionID == targetRegion.ID {
			return diplomacy.Result{}
		}
	}
	attacker := gs.Factions[attackerID]
	defender := gs.Factions[defenderID]
	if attacker == nil || defender == nil || attacker.IsEliminated || defender.IsEliminated || defender.AIAggressiveness >= aiVassalResistanceAggressiveness {
		return diplomacy.Result{}
	}
	if diplomacy.DirectOverlord(gs, attackerID) != "" || diplomacy.DirectOverlord(gs, defenderID) != "" {
		return diplomacy.Result{}
	}
	if rel := diplomacy.Relation(gs, attackerID, defenderID); rel == nil || rel.Stance != faction.StanceWar {
		return diplomacy.Result{}
	}
	if aiHasExternalAlly(gs, defenderID, attackerID) {
		return diplomacy.Result{}
	}

	attackerPower := diplomacy.MilitaryPower(gs, attackerID)
	defenderPower := diplomacy.MilitaryPower(gs, defenderID)
	if attackerPower <= 0 || (defenderPower > 0 && attackerPower*100 < defenderPower*160) {
		return diplomacy.Result{}
	}
	return diplomacy.ForceVassalizeAfterWar(gs, attackerID, defenderID)
}

func aiHasExternalAlly(gs *state.GameState, defenderID, attackerID faction.FactionID) bool {
	for _, relation := range gs.Relations {
		if relation == nil || relation.Stance != faction.StanceAllied {
			continue
		}
		otherID := faction.FactionID("")
		switch defenderID {
		case relation.FactionA:
			otherID = relation.FactionB
		case relation.FactionB:
			otherID = relation.FactionA
		default:
			continue
		}
		if otherID == attackerID || diplomacy.SameRealm(gs, defenderID, otherID) {
			continue
		}
		if other := gs.Factions[otherID]; other != nil && !other.IsEliminated {
			return true
		}
	}
	return false
}
