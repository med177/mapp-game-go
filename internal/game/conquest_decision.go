package game

import (
	"fmt"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

type pendingConquestDecision struct {
	RegionID           world.RegionID
	AttackerFactionID  faction.FactionID
	DefenderFactionID  faction.FactionID
	SuccessorFactionID faction.FactionID
}

type captiveDecisionOutcome uint8

const (
	captiveDecisionRelease captiveDecisionOutcome = iota
	captiveDecisionKeep
	captiveDecisionExecute
	captiveDecisionRansom
)

func (g *Game) syncPendingCaptiveDecisionsToState() {
	if g == nil || g.gs == nil {
		return
	}
	g.gs.PendingCaptiveDecisions = append([]state.PendingCaptiveDecision(nil), g.pendingCaptiveDecisions...)
}

func (g *Game) restorePendingCaptiveDecisions() {
	if g == nil || g.gs == nil {
		return
	}
	g.pendingCaptiveDecisions = append([]state.PendingCaptiveDecision(nil), g.gs.PendingCaptiveDecisions...)
}

func (g *Game) queueCaptiveDecisionsForRegion(region *world.Region, newOwnerID string) {
	if g == nil || g.gs == nil || region == nil || newOwnerID == "" || g.gs.PlayerFactionID != faction.FactionID(newOwnerID) {
		return
	}
	pending := make(map[string]bool, len(g.pendingCaptiveDecisions))
	for _, decision := range g.pendingCaptiveDecisions {
		pending[decision.CommanderID] = true
	}
	settlements := make(map[string]bool, len(region.Settlements))
	for _, settlement := range region.Settlements {
		settlements[settlement.ID] = true
	}
	for _, commander := range g.gs.Commanders {
		if commander == nil || commander.CaptiveAtSettlementID == "" || commander.OwnerID == newOwnerID || pending[commander.ID] || !settlements[commander.CaptiveAtSettlementID] {
			continue
		}
		g.pendingCaptiveDecisions = append(g.pendingCaptiveDecisions, state.PendingCaptiveDecision{
			CommanderID:     commander.ID,
			SettlementID:    commander.CaptiveAtSettlementID,
			CaptorFactionID: faction.FactionID(newOwnerID),
		})
	}
	g.syncPendingCaptiveDecisionsToState()
	if len(g.pendingCaptiveDecisions) == 1 {
		g.showPendingCaptiveDecision(false)
	}
}

func (g *Game) queueCapturedCommanderDecision(commanderID, settlementID, captorID string, afterBattleReport bool) {
	if g == nil || g.gs == nil || commanderID == "" || settlementID == "" || captorID == "" || g.gs.PlayerFactionID != faction.FactionID(captorID) {
		return
	}
	for _, decision := range g.pendingCaptiveDecisions {
		if decision.CommanderID == commanderID {
			return
		}
	}
	g.pendingCaptiveDecisions = append(g.pendingCaptiveDecisions, state.PendingCaptiveDecision{
		CommanderID:     commanderID,
		SettlementID:    settlementID,
		CaptorFactionID: faction.FactionID(captorID),
	})
	g.syncPendingCaptiveDecisionsToState()
	if len(g.pendingCaptiveDecisions) == 1 {
		g.showPendingCaptiveDecision(afterBattleReport)
	}
}

func (g *Game) showPendingCaptiveDecision(afterBattleReport bool) {
	if g == nil || g.gs == nil || g.renderer == nil || len(g.pendingCaptiveDecisions) == 0 {
		return
	}
	decision := g.pendingCaptiveDecisions[0]
	commander := g.gs.Commanders[decision.CommanderID]
	if commander == nil {
		g.pendingCaptiveDecisions = g.pendingCaptiveDecisions[1:]
		g.syncPendingCaptiveDecisionsToState()
		g.showPendingCaptiveDecision(afterBattleReport)
		return
	}
	name := commander.Name
	if name == "" {
		name = commander.ID
	}
	regionID := regionForSettlement(g.gs, decision.SettlementID)
	regionName := string(regionID)
	if region := g.gs.Regions[regionID]; region != nil && region.NameTR != "" {
		regionName = region.NameTR
	}
	message := fmt.Sprintf("%s bölgesinde %s adlı %s komutanı tutsak edildi. Karar ver.", regionName, name, g.factionNameTR(commander.OwnerID))
	amount := g.gs.CommanderRansomGold
	if amount <= 0 {
		amount = state.DefaultCommanderRansomGold
	}
	message += fmt.Sprintf(" Fidye bedeli: %d altın.", amount)
	ransomAction := render.InputAction{Kind: render.ActionRansomCaptiveCommander}
	acceptAction := render.InputAction{Kind: render.ActionReleaseCaptiveCommander}
	keepAction := render.InputAction{Kind: render.ActionKeepCaptiveCommander}
	executeAction := render.InputAction{Kind: render.ActionExecuteCaptiveCommander}
	if afterBattleReport {
		g.renderer.QueueFourChoiceDialogAfterBattleReport("Tutsak Kararı", message, "Fidye iste", "Serbest bırak", "Tutsak tut", "İnfaz et", ransomAction, acceptAction, keepAction, executeAction)
		return
	}
	g.renderer.ShowFourChoiceDialog("Tutsak Kararı", message, "Fidye iste", "Serbest bırak", "Tutsak tut", "İnfaz et", ransomAction, acceptAction, keepAction, executeAction)
}

func (g *Game) resolvePendingCaptiveDecision(outcome captiveDecisionOutcome) {
	if g == nil || g.gs == nil || len(g.pendingCaptiveDecisions) == 0 {
		return
	}
	decision := g.pendingCaptiveDecisions[0]
	g.pendingCaptiveDecisions = g.pendingCaptiveDecisions[1:]
	g.syncPendingCaptiveDecisionsToState()
	commander := g.gs.Commanders[decision.CommanderID]
	regionID := regionForSettlement(g.gs, decision.SettlementID)
	region := g.gs.Regions[regionID]
	valid := commander != nil && commander.CaptiveAtSettlementID == decision.SettlementID && region != nil && region.OwnerID == string(decision.CaptorFactionID)
	if valid {
		switch outcome {
		case captiveDecisionRansom:
			if !g.gs.PayCommanderRansom(commander.ID, decision.CaptorFactionID) {
				g.pendingCaptiveDecisions = append([]state.PendingCaptiveDecision{decision}, g.pendingCaptiveDecisions...)
				g.syncPendingCaptiveDecisionsToState()
				if g.renderer != nil {
					g.renderer.ShowCombatResult("Fidye ödenemedi: komutanın sahibinin yeterli altını yok.")
					g.showPendingCaptiveDecision(false)
				}
				return
			}
			diplomacy.AddRelationScoreBoth(g.gs, decision.CaptorFactionID, faction.FactionID(commander.OwnerID), state.CommanderRansomRelationDelta)
		case captiveDecisionRelease:
			g.gs.ReleaseCommanderCaptivity(commander.ID)
			diplomacy.AddRelationScoreBoth(g.gs, decision.CaptorFactionID, faction.FactionID(commander.OwnerID), state.CommanderReleaseRelationDelta)
		case captiveDecisionExecute:
			g.gs.DismissCommander(commander.ID)
			diplomacy.AddRelationScoreBoth(g.gs, decision.CaptorFactionID, faction.FactionID(commander.OwnerID), state.CommanderExecuteRelationDelta)
		case captiveDecisionKeep:
			if !commander.CaptivityKeepPenaltyApplied {
				diplomacy.AddRelationScoreBoth(g.gs, decision.CaptorFactionID, faction.FactionID(commander.OwnerID), state.CommanderKeepRelationDelta)
				commander.CaptivityKeepPenaltyApplied = true
			}
		}
	}
	if g.renderer != nil && valid {
		g.renderer.ShowCombatResult("Tutsak kararı uygulandı.")
	}
	if len(g.pendingCaptiveDecisions) > 0 {
		g.showPendingCaptiveDecision(false)
	}
}

func (g *Game) resolveHeldCaptiveDecision(commanderID string, outcome captiveDecisionOutcome) {
	if g == nil || g.gs == nil || commanderID == "" {
		return
	}
	commander := g.gs.Commanders[commanderID]
	if commander == nil || commander.CaptiveAtSettlementID == "" {
		return
	}
	regionID := regionForSettlement(g.gs, commander.CaptiveAtSettlementID)
	region := g.gs.Regions[regionID]
	if region == nil || region.OwnerID != string(g.gs.PlayerFactionID) {
		return
	}
	captiveOwner := faction.FactionID(commander.OwnerID)
	switch outcome {
	case captiveDecisionRansom:
		if !g.gs.PayCommanderRansom(commanderID, g.gs.PlayerFactionID) {
			if g.renderer != nil {
				g.renderer.ShowCombatResult("Fidye ödenemedi: yeterli altın yok.")
			}
			return
		}
		diplomacy.AddRelationScoreBoth(g.gs, g.gs.PlayerFactionID, captiveOwner, state.CommanderRansomRelationDelta)
	case captiveDecisionRelease:
		g.gs.ReleaseCommanderCaptivity(commanderID)
		diplomacy.AddRelationScoreBoth(g.gs, g.gs.PlayerFactionID, captiveOwner, state.CommanderReleaseRelationDelta)
	case captiveDecisionKeep:
		if !commander.CaptivityKeepPenaltyApplied {
			diplomacy.AddRelationScoreBoth(g.gs, g.gs.PlayerFactionID, captiveOwner, state.CommanderKeepRelationDelta)
			commander.CaptivityKeepPenaltyApplied = true
		}
	case captiveDecisionExecute:
		g.gs.DismissCommander(commanderID)
		diplomacy.AddRelationScoreBoth(g.gs, g.gs.PlayerFactionID, captiveOwner, state.CommanderExecuteRelationDelta)
	}
	if g.renderer != nil {
		g.renderer.ShowCombatResult("Tutsak kararı uygulandı.")
	}
}

func (g *Game) offerCommanderRansom(commanderID string) {
	if g == nil || g.gs == nil || commanderID == "" {
		return
	}
	commander := g.gs.Commanders[commanderID]
	if commander == nil || commander.OwnerID != string(g.gs.PlayerFactionID) || commander.CaptiveAtSettlementID == "" {
		return
	}
	regionID := regionForSettlement(g.gs, commander.CaptiveAtSettlementID)
	region := g.gs.Regions[regionID]
	if region == nil || region.OwnerID == string(g.gs.PlayerFactionID) {
		return
	}
	captorID := faction.FactionID(region.OwnerID)
	accepted := !diplomacy.IsWar(g.gs, g.gs.PlayerFactionID, captorID) && diplomacy.RelationScore(g.gs, g.gs.PlayerFactionID, captorID) >= -25
	if !accepted {
		if g.renderer != nil {
			g.renderer.ShowCombatResult("Fidye teklifi reddedildi.")
		}
		return
	}
	if !g.gs.PayCommanderRansom(commanderID, captorID) {
		if g.renderer != nil {
			g.renderer.ShowCombatResult("Fidye ödenemedi: yeterli altın yok.")
		}
		return
	}
	diplomacy.AddRelationScoreBoth(g.gs, g.gs.PlayerFactionID, captorID, state.CommanderRansomRelationDelta)
	if g.renderer != nil {
		g.renderer.ShowCombatResult("Fidye kabul edildi. Komutan serbest bırakıldı.")
	}
}

func regionForSettlement(gs *state.GameState, settlementID string) world.RegionID {
	if gs == nil || settlementID == "" {
		return ""
	}
	for regionID, region := range gs.Regions {
		if region == nil {
			continue
		}
		for _, settlement := range region.Settlements {
			if settlement.ID == settlementID {
				return regionID
			}
		}
	}
	return ""
}

func (g *Game) syncPendingConquestDecisionsToState() {
	if g == nil || g.gs == nil {
		return
	}
	decisions := make([]state.PendingConquestDecision, 0, len(g.pendingConquestDecisions))
	for _, decision := range g.pendingConquestDecisions {
		decisions = append(decisions, state.PendingConquestDecision{
			RegionID:           decision.RegionID,
			AttackerFactionID:  decision.AttackerFactionID,
			DefenderFactionID:  decision.DefenderFactionID,
			SuccessorFactionID: decision.SuccessorFactionID,
		})
	}
	g.gs.PendingConquestDecisions = decisions
}

func (g *Game) restorePendingConquestDecisions() {
	if g == nil || g.gs == nil {
		return
	}
	if len(g.gs.PendingConquestDecisions) == 0 {
		g.recoverPendingConquestDecisionFromHistory()
	}
	g.pendingConquestDecisions = make([]pendingConquestDecision, 0, len(g.gs.PendingConquestDecisions))
	for _, decision := range g.gs.PendingConquestDecisions {
		g.pendingConquestDecisions = append(g.pendingConquestDecisions, pendingConquestDecision{
			RegionID:           decision.RegionID,
			AttackerFactionID:  decision.AttackerFactionID,
			DefenderFactionID:  decision.DefenderFactionID,
			SuccessorFactionID: decision.SuccessorFactionID,
		})
	}
}

// recoverPendingConquestDecisionFromHistory, karar alanı eklenmeden önce
// kaydedilmiş teslimiyet save'lerinde kaybolan ardıl kararını geri kurar.
func (g *Game) recoverPendingConquestDecisionFromHistory() {
	if g == nil || g.gs == nil || len(g.gs.DiplomaticOfferHistory) == 0 {
		return
	}
	for i := len(g.gs.DiplomaticOfferHistory) - 1; i >= 0; i-- {
		history := g.gs.DiplomaticOfferHistory[i]
		if history.Action != string(diplomacy.ActionProposeSurrender) || !history.Accepted || !history.Applied || history.RegionID == "" {
			continue
		}
		region := g.gs.Regions[history.RegionID]
		if region == nil || region.OwnerID != string(history.FromFactionID) || region.SuccessorFactionID == "" {
			continue
		}
		successorID := faction.FactionID(region.SuccessorFactionID)
		successor := g.gs.Factions[successorID]
		if successor == nil || !successor.IsEliminated || len(g.gs.LandRegionsOwnedBy(successorID)) != 0 {
			continue
		}
		attacker := g.gs.Factions[history.ToFactionID]
		if attacker == nil || attacker.IsEliminated || history.ToFactionID == history.FromFactionID {
			continue
		}
		g.gs.PendingConquestDecisions = []state.PendingConquestDecision{{
			RegionID:           region.ID,
			AttackerFactionID:  history.ToFactionID,
			DefenderFactionID:  history.FromFactionID,
			SuccessorFactionID: successorID,
		}}
		if g.gs.Sieges != nil {
			delete(g.gs.Sieges, region.ID)
		}
		return
	}
}

func (g *Game) enqueuePendingConquestDecision(decision pendingConquestDecision) {
	if g == nil {
		return
	}
	g.pendingConquestDecisions = append(g.pendingConquestDecisions, decision)
	g.syncPendingConquestDecisionsToState()
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
	// Deniz savaşı bir kara bölgesinin fethi değildir. Deniz bölgelerinde
	// OwnerID kalıntısı bulunsa bile savaş sonrası ilhak/vassallık paneli
	// açılmamalıdır.
	if targetRegion.IsSea {
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
	if targetRegion.IsSea {
		return false
	}
	defenderID := faction.FactionID(targetRegion.OwnerID)
	// Kendi kuşatılmış bölgemizi kurtarmak bir fetih değildir. Özellikle
	// bölgede elenmiş bir ardıl devlet tanımı kalmışsa, aşağıdaki ardıl karar
	// dalı aksi halde yanlışlıkla ilhak/vassallık paneli açabilir.
	if defenderID == "" || defenderID == attackerID {
		return false
	}
	successorID := faction.FactionID(targetRegion.SuccessorFactionID)
	if successorID != "" && !g.gs.CanRestoreSuccessorAtRegion(targetRegion) {
		// Ardıl devlet hâlâ oyundaysa veya geçersiz bir state taşıyorsa,
		// genel son-toprak vassallık paneline düşmeden bölge ilhak edilir.
		return false
	}
	if g.shouldOfferSuccessorDecision(attackerID, defenderID, successorID, targetRegion) {
		g.enqueuePendingConquestDecision(pendingConquestDecision{
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
	g.enqueuePendingConquestDecision(pendingConquestDecision{
		RegionID:          targetRegion.ID,
		AttackerFactionID: attackerID,
		DefenderFactionID: defenderID,
	})
	if len(g.pendingConquestDecisions) == 1 {
		g.showPendingConquestDecision(showAfterBattleReport)
	}
	return true
}

// captureUnfortifiedRegion, düşman bölgesinde temas sonrası pozisyonunu koruyan
// ordunun, hedefte düşman ordusu yoksa, tahkimatsız hedefi doğrudan ele
// geçirmesini sağlar. Tahkimatlı veya düşman ordusu bulunan hedef bu aksiyona
// hiç giremez; sırasıyla kuşatma veya muharebe akışından devam edilmelidir.
func (g *Game) captureUnfortifiedRegion(aid army.ArmyID, targetID world.RegionID) {
	if g == nil || g.gs == nil {
		return
	}
	attacker := g.gs.Armies[aid]
	target := g.gs.Regions[targetID]
	if attacker == nil || target == nil || attacker.OwnerID != string(g.gs.PlayerFactionID) || attacker.IsNaval ||
		target.IsSea || target.OwnerID == "" || target.OwnerID == attacker.OwnerID || target.IsFortified() ||
		attacker.RegionID != target.ID || !gameFactionsAtWar(g.gs, attacker.OwnerID, target.OwnerID) ||
		g.gs.SelectBattleDefender(attacker, target.ID, false) != nil {
		return
	}

	g.gs.RecordFactionRegionAttackAgainst(faction.FactionID(attacker.OwnerID), faction.FactionID(target.OwnerID))
	collapse := g.applyConquestWithNavalEviction(target, attacker.OwnerID)
	if g.renderer != nil {
		g.renderer.MarkMapDirty()
		message := fmt.Sprintf("%s ele geçirildi.", target.NameTR)
		g.renderer.ShowCombatResult(message)
		g.renderer.AddEventDetail("[FETİH] "+message, fmt.Sprintf("%s tahkimli olmadığı için bölge doğrudan %s yönetimine katıldı.", target.NameTR, g.factionNameTR(attacker.OwnerID)))
		g.showPendingCaptiveReleaseNotification()
	}
	g.announceElimination(collapse)
}

func (g *Game) shouldOfferSuccessorDecision(attackerID, defenderID, successorID faction.FactionID, targetRegion *world.Region) bool {
	if g == nil || g.gs == nil || targetRegion == nil || attackerID == "" || defenderID == "" || successorID == "" {
		return false
	}
	if g.gs.PlayerFactionID != attackerID || targetRegion.IsSea || targetRegion.OwnerID != string(defenderID) || attackerID == successorID {
		return false
	}
	if !g.gs.CanRestoreSuccessorAtRegion(targetRegion) {
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
	g.syncPendingConquestDecisionsToState()
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
	if !g.gs.CanRestoreSuccessorAtRegion(region) {
		collapse := g.applyConquestWithNavalEviction(region, string(decision.AttackerFactionID))
		g.finishSuccessorDecision(region, "İlhak edildi.", collapse)
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
	g.syncPendingConquestDecisionsToState()

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
