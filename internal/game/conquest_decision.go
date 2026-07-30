package game

import (
	"fmt"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/world"
)

type pendingConquestDecision struct {
	RegionID           world.RegionID
	AttackerFactionID  faction.FactionID
	DefenderFactionID  faction.FactionID
	SuccessorFactionID faction.FactionID
}

type successorDecisionOutcome uint8

const (
	successorDecisionAnnex successorDecisionOutcome = iota
	successorDecisionRelease
	successorDecisionVassalize
)

func (g *Game) shouldOfferPostWarVassalization(attackerID, defenderID faction.FactionID, targetRegion *world.Region) bool {
	if g == nil || g.gs == nil || targetRegion == nil || attackerID == "" || defenderID == "" || attackerID == defenderID {
		return false
	}
	if g.gs.PlayerFactionID != attackerID {
		return false
	}
	if targetRegion.OwnerID != string(defenderID) {
		return false
	}
	if diplomacy.DirectOverlord(g.gs, attackerID) != "" || diplomacy.DirectOverlord(g.gs, defenderID) != "" {
		return false
	}
	return len(g.gs.LandRegionsOwnedBy(defenderID)) == 1
}

func (g *Game) queueConquestDecision(attackerID faction.FactionID, targetRegion *world.Region, showAfterBattleReport bool) bool {
	if g == nil || g.gs == nil || g.renderer == nil || targetRegion == nil {
		return false
	}
	defenderID := faction.FactionID(targetRegion.OwnerID)
	successorID := faction.FactionID(targetRegion.SuccessorFactionID)
	if g.shouldOfferSuccessorDecision(attackerID, defenderID, successorID, targetRegion) {
		g.pendingConquestDecisions = append(g.pendingConquestDecisions, pendingConquestDecision{
			RegionID:           targetRegion.ID,
			AttackerFactionID:  attackerID,
			DefenderFactionID:  defenderID,
			SuccessorFactionID: successorID,
		})
		if len(g.pendingConquestDecisions) == 1 {
			g.showPendingConquestDecision(showAfterBattleReport)
		}
		return true
	}
	if !g.shouldOfferPostWarVassalization(attackerID, defenderID, targetRegion) {
		return false
	}
	g.pendingConquestDecisions = append(g.pendingConquestDecisions, pendingConquestDecision{
		RegionID:          targetRegion.ID,
		AttackerFactionID: attackerID,
		DefenderFactionID: defenderID,
	})
	if len(g.pendingConquestDecisions) == 1 {
		g.showPendingConquestDecision(showAfterBattleReport)
	}
	return true
}

func (g *Game) shouldOfferSuccessorDecision(attackerID, defenderID, successorID faction.FactionID, targetRegion *world.Region) bool {
	if g == nil || g.gs == nil || targetRegion == nil || attackerID == "" || defenderID == "" || successorID == "" {
		return false
	}
	if g.gs.PlayerFactionID != attackerID || targetRegion.IsSea || targetRegion.OwnerID != string(defenderID) || attackerID == successorID {
		return false
	}
	if g.gs.Factions[successorID] == nil || g.gs.Factions[successorID].IsEliminated && len(g.gs.LandRegionsOwnedBy(successorID)) > 0 {
		return false
	}
	return true
}

func (g *Game) showPendingConquestDecision(showAfterBattleReport bool) {
	if g == nil || g.renderer == nil || len(g.pendingConquestDecisions) == 0 {
		return
	}
	decision := g.pendingConquestDecisions[0]
	region := g.gs.Regions[decision.RegionID]
	if region == nil {
		return
	}
	if decision.SuccessorFactionID != "" {
		successorName := g.factionNameTR(string(decision.SuccessorFactionID))
		message := fmt.Sprintf("%s bölgesinin ardıl devleti %s. Bölgenin kaderini seç.", region.NameTR, successorName)
		acceptAction := render.InputAction{Kind: render.ActionAnnexSuccessor}
		releaseAction := render.InputAction{Kind: render.ActionReleaseSuccessor}
		vassalAction := render.InputAction{Kind: render.ActionVassalizeSuccessor}
		if showAfterBattleReport {
			g.renderer.QueueThreeChoiceDialogAfterBattleReport("Ardıl Devlet Kararı", message, "İlhak Et", "Serbest Bırak", "Vassal Yap", acceptAction, releaseAction, vassalAction)
		} else {
			g.renderer.ShowThreeChoiceDialog(
				"Ardıl Devlet Kararı",
				message,
				"İlhak Et",
				"Serbest Bırak",
				"Vassal Yap",
				acceptAction,
				releaseAction,
				vassalAction,
			)
		}
		return
	}
	defenderName := g.factionNameTR(string(decision.DefenderFactionID))
	message := fmt.Sprintf("%s devleti son toprağında teslim oldu. %s bölgesini ilhak edebilir ya da devleti haraç veren bir vassal olarak bırakabilirsin.", defenderName, region.NameTR)
	annexAction := render.InputAction{Kind: render.ActionAnnexDefeatedFaction}
	vassalAction := render.InputAction{Kind: render.ActionVassalizeDefeatedFaction}
	if showAfterBattleReport {
		g.renderer.QueueChoiceDialogAfterBattleReport("Savaş Sonrası Düzen", message, "İlhak Et", "Vassal Yap", annexAction, vassalAction)
		return
	}
	g.renderer.ShowChoiceDialog("Savaş Sonrası Düzen", message, "İlhak Et", "Vassal Yap", annexAction, vassalAction)
}

func (g *Game) resolvePendingSuccessorDecision(outcome successorDecisionOutcome) {
	if g == nil || g.gs == nil || len(g.pendingConquestDecisions) == 0 {
		return
	}
	decision := g.pendingConquestDecisions[0]
	if decision.SuccessorFactionID == "" {
		return
	}
	g.pendingConquestDecisions = g.pendingConquestDecisions[1:]
	region := g.gs.Regions[decision.RegionID]
	if region == nil || region.OwnerID != string(decision.DefenderFactionID) {
		if g.renderer != nil {
			g.renderer.ShowCombatResult("Ardıl devlet kararı artık geçerli değil.")
		}
		g.showPendingConquestDecision(false)
		return
	}

	successorID := decision.SuccessorFactionID
	successor := g.gs.Factions[successorID]
	if successor == nil || decision.AttackerFactionID == successorID {
		if g.renderer != nil {
			g.renderer.ShowCombatResult("Ardıl devlet kararı uygulanamadı.")
		}
		g.showPendingConquestDecision(false)
		return
	}

	if outcome == successorDecisionAnnex {
		collapse := g.applyConquestWithNavalEviction(region, string(decision.AttackerFactionID))
		g.finishSuccessorDecision(region, "İlhak edildi.", collapse)
		g.showPendingConquestDecision(false)
		return
	}

	if !g.reviveSuccessorAtRegion(region.ID, successorID) {
		if g.renderer != nil {
			g.renderer.ShowCombatResult("Ardıl devlet yeniden kurulamadı.")
		}
		g.showPendingConquestDecision(false)
		return
	}

	var result diplomacy.Result
	if outcome == successorDecisionVassalize {
		result = diplomacy.ForceVassalizeAfterWar(g.gs, decision.AttackerFactionID, successorID)
	} else {
		result = diplomacy.ForceReleaseAfterWar(g.gs, decision.AttackerFactionID, successorID)
	}
	if !result.Applied {
		if g.renderer != nil {
			g.renderer.ShowCombatResult(result.Message)
		}
		g.showPendingConquestDecision(false)
		return
	}

	collapse := eliminationResult{}
	if region.OwnerID != string(successorID) {
		collapse = g.applyConquestWithNavalEviction(region, string(successorID))
	} else {
		g.retreatArmiesFromCapturedRegion(region.ID, string(successorID))
	}
	label := "Serbest bırakıldı."
	if outcome == successorDecisionVassalize {
		label = "vassal olarak bırakıldı."
	}
	g.finishSuccessorDecision(region, label, collapse)
	g.showPendingConquestDecision(false)
}

func (g *Game) finishSuccessorDecision(region *world.Region, suffix string, collapse eliminationResult) {
	if g == nil || g.renderer == nil || region == nil {
		return
	}
	g.renderer.MarkMapDirty()
	msg := fmt.Sprintf("%s %s", region.NameTR, suffix)
	g.renderer.ShowCombatResult(msg)
	g.renderer.AddEventDetail("[ARDIL DEVLET] "+msg, "Bölgenin savaş sonrası siyasi düzeni uygulandı.")
	g.announceElimination(collapse)
}

func (g *Game) resolvePendingConquestDecision(vassalize bool) {
	if g == nil || g.gs == nil || len(g.pendingConquestDecisions) == 0 {
		return
	}
	decision := g.pendingConquestDecisions[0]
	g.pendingConquestDecisions = g.pendingConquestDecisions[1:]

	region := g.gs.Regions[decision.RegionID]
	if region == nil {
		g.showPendingConquestDecision(false)
		return
	}
	if region.OwnerID != string(decision.DefenderFactionID) {
		if g.renderer != nil {
			g.renderer.ShowCombatResult("Savaş sonrası karar artık geçerli değil.")
		}
		g.showPendingConquestDecision(false)
		return
	}

	if vassalize {
		result := diplomacy.ForceVassalizeAfterWar(g.gs, decision.AttackerFactionID, decision.DefenderFactionID)
		if g.renderer != nil && result.Message != "" {
			g.renderer.ShowCombatResult(result.Message)
			g.renderer.AddEventDetail("[DIPLOMASI] "+result.Message, fmt.Sprintf("%s devleti savaş sonrası vassal statüsüne geçirildi. %s bölgesi yerel yönetimde bırakıldı.", g.factionNameTR(string(decision.DefenderFactionID)), region.NameTR))
		}
		g.showPendingConquestDecision(false)
		return
	}

	collapse := g.applyConquestWithNavalEviction(region, string(decision.AttackerFactionID))
	if g.renderer != nil {
		g.renderer.MarkMapDirty()
		msg := fmt.Sprintf("%s ilhak edildi.", region.NameTR)
		g.renderer.ShowCombatResult(msg)
		g.renderer.AddEventDetail("[FETİH] "+msg, fmt.Sprintf("%s bölgesi doğrudan %s yönetimine geçti.", region.NameTR, g.factionNameTR(string(decision.AttackerFactionID))))
	}
	g.announceElimination(collapse)
	g.showPendingConquestDecision(false)
}
