package game

import (
	"fmt"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/world"
)

type pendingConquestDecision struct {
	RegionID          world.RegionID
	AttackerFactionID faction.FactionID
	DefenderFactionID faction.FactionID
}

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

func (g *Game) showPendingConquestDecision(showAfterBattleReport bool) {
	if g == nil || g.renderer == nil || len(g.pendingConquestDecisions) == 0 {
		return
	}
	decision := g.pendingConquestDecisions[0]
	region := g.gs.Regions[decision.RegionID]
	if region == nil {
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
