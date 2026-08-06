package ai

import (
	"fmt"
	"hash/fnv"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const aiVassalResistanceAggressiveness = 65
const aiVassalizationChancePercent = 50

func aiVassalizationRoll(gs *state.GameState, attackerID, defenderID faction.FactionID, regionID world.RegionID) int {
	hasher := fnv.New32a()
	_, _ = fmt.Fprintf(hasher, "%d|%s|%s|%s|vassalization", gs.Turn, attackerID, defenderID, regionID)
	return int(hasher.Sum32() % 100)
}

func aiApplyConquest(gs *state.GameState, region *world.Region, newOwnerID string) {
	if gs == nil || region == nil || newOwnerID == "" {
		return
	}
	gs.RecordWarRegionCapture(faction.FactionID(newOwnerID), faction.FactionID(region.OwnerID))
	region.ApplyConquest(newOwnerID, aiOwnerReligion(gs, newOwnerID))
}

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
	if targetRegion.SuccessorFactionID != "" && !gs.CanRestoreSuccessorAtRegion(targetRegion) {
		// Aktif ardıl devlet metadata'sı olan bölge vassal bırakılmaz; çağıran
		// conquest akışı bu sonucu doğrudan ilhak olarak çözer.
		return diplomacy.Result{}
	}
	plan := gs.AIPlans[attackerID]
	if plan == nil || !plan.AllowVassalization || plan.TargetFactionID != defenderID {
		return diplomacy.Result{}
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
	// Vassallık, bütün stratejik uygunluk kontrolleri geçildikten sonra
	// bölgenin ele geçirildiği anda kararlaştırılır. Zar deterministik tutulur:
	// aynı save, tur, bölge ve taraflar aynı sonucu üretir.
	if aiVassalizationRoll(gs, attackerID, defenderID, targetRegion.ID) >= aiVassalizationChancePercent {
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
