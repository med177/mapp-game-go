package game

import (
	"fmt"
	"sort"
	"strings"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

const capitalLootPercent = 50

func (g *Game) queueCapitalMove(fid faction.FactionID, settlementID string, turns int, source string) bool {
	if g == nil || g.gs == nil || fid == "" || settlementID == "" {
		return false
	}
	if turns <= 0 {
		turns = state.DefaultCapitalMoveTurns
	}
	if !g.gs.StartCapitalMove(fid, settlementID, turns) {
		return false
	}
	region, settlement, _, _ := g.gs.FindSettlementByID(settlementID)
	name := capitalSettlementName(region, settlement, settlementID)
	factionName := factionDisplayName(g.gs, fid)
	msg := fmt.Sprintf("%s için başkent taşıması başladı: %s (%d tur).", factionName, name, turns)
	if source != "" {
		msg = fmt.Sprintf("%s için başkent taşıması başladı: %s (%s, %d tur).", factionName, name, source, turns)
	}
	if g.renderer != nil {
		if fid == g.gs.PlayerFactionID {
			g.renderer.ShowCombatResult(msg)
		}
		g.renderer.AddEvent("[BASKENT] " + msg)
	}
	return true
}

func (g *Game) handleCapitalMoveProgress(updates []state.CapitalMoveProgress) {
	if g == nil || g.gs == nil || g.renderer == nil || len(updates) == 0 {
		return
	}
	for _, update := range updates {
		factionName := factionDisplayName(g.gs, update.FactionID)
		switch {
		case update.Completed:
			msg := fmt.Sprintf("%s yeni başkentini %s olarak kurdu.", factionName, update.SettlementName)
			if update.FactionID == g.gs.PlayerFactionID {
				g.renderer.ShowCombatResult(msg)
			}
			g.renderer.AddEvent("[BASKENT] " + msg)
		case update.Cancelled:
			msg := fmt.Sprintf("%s için %s yönündeki başkent taşıması iptal oldu.", factionName, update.SettlementName)
			if update.FactionID == g.gs.PlayerFactionID {
				g.renderer.ShowCombatResult(msg)
			}
			g.renderer.AddEvent("[BASKENT] " + msg)
		}
	}
}

func (g *Game) handleCapitalCapture(prevFactionID faction.FactionID, newOwnerID string, targetRegion *world.Region) {
	if g == nil || g.gs == nil || prevFactionID == "" || newOwnerID == "" || targetRegion == nil {
		return
	}
	defender := g.gs.Factions[prevFactionID]
	attacker := g.gs.Factions[faction.FactionID(newOwnerID)]
	if defender == nil || attacker == nil {
		return
	}
	capitalSettlementID := defender.CapitalSettlementID
	if capitalSettlementID == "" || !regionHasSettlement(targetRegion, capitalSettlementID) {
		return
	}

	loot := transferCapitalStockpile(defender, attacker)
	techs := captureCapitalTechnologies(attacker, defender, g.gs.TechTypes)

	defender.CapitalSettlementID = ""
	defender.PendingCapitalSettlementID = ""
	defender.PendingCapitalTurns = 0

	newCapitalName := ""
	if nextCapitalID, ok := g.gs.BestCapitalSettlementForFaction(prevFactionID); ok {
		g.gs.SetFactionCapital(prevFactionID, nextCapitalID)
		region, settlement, _, _ := g.gs.FindSettlementByID(nextCapitalID)
		newCapitalName = capitalSettlementName(region, settlement, nextCapitalID)
	}

	g.announceCapitalCapture(prevFactionID, faction.FactionID(newOwnerID), targetRegion, loot, techs, newCapitalName)
}

func (g *Game) announceCapitalCapture(defenderID, attackerID faction.FactionID, region *world.Region, loot map[economy.ResourceKind]int, techs []string, newCapitalName string) {
	if g == nil || g.renderer == nil || region == nil {
		return
	}
	defenderName := factionDisplayName(g.gs, defenderID)
	attackerName := factionDisplayName(g.gs, attackerID)
	msg := fmt.Sprintf("%s, %s başkentini ele geçirdi.", attackerName, defenderName)
	if attackerID == g.gs.PlayerFactionID || defenderID == g.gs.PlayerFactionID {
		g.renderer.ShowCombatResult(msg)
	}

	detailParts := []string{msg, "Yer: " + region.NameTR}
	if lootSummary := formatCapitalLoot(loot); lootSummary != "" {
		detailParts = append(detailParts, "Ele geçirilen hazine: "+lootSummary)
	}
	if len(techs) > 0 {
		detailParts = append(detailParts, "Kopyalanan teknolojiler: "+strings.Join(techLabels(g.gs, techs), ", "))
	}
	if newCapitalName != "" {
		detailParts = append(detailParts, defenderName+" için yeni başkent: "+newCapitalName)
	}
	g.renderer.AddEventDetail("[BASKENT] "+msg, strings.Join(detailParts, "\n"))
}

func transferCapitalStockpile(defender, attacker *faction.Faction) map[economy.ResourceKind]int {
	loot := make(map[economy.ResourceKind]int, 7)
	for _, kind := range economy.AllResourceKinds() {
		amount := economy.FactionResourceAmount(defender, kind) * capitalLootPercent / 100
		if amount <= 0 {
			continue
		}
		economy.AddFactionResource(defender, kind, -amount)
		economy.AddFactionResource(attacker, kind, amount)
		loot[kind] = amount
	}
	return loot
}

func captureCapitalTechnologies(attacker, defender *faction.Faction, techTypes map[string]*tech.Technology) []string {
	if attacker == nil || defender == nil {
		return nil
	}
	if attacker.Research.Completed == nil {
		attacker.Research.Completed = make(map[string]bool)
	}
	missing := make([]string, 0)
	for techID := range defender.Research.Completed {
		if attacker.Research.Completed[techID] {
			continue
		}
		if techTypes != nil {
			if _, ok := techTypes[techID]; !ok {
				continue
			}
		}
		missing = append(missing, techID)
	}
	if len(missing) == 0 {
		return nil
	}
	sort.SliceStable(missing, func(i, j int) bool {
		left := techTypes[missing[i]]
		right := techTypes[missing[j]]
		if left == nil || right == nil {
			return missing[i] < missing[j]
		}
		if left.TurnsRequired != right.TurnsRequired {
			return left.TurnsRequired > right.TurnsRequired
		}
		if left.GoldCost != right.GoldCost {
			return left.GoldCost > right.GoldCost
		}
		if left.Category != right.Category {
			return tech.CategoryOrder(left.Category) < tech.CategoryOrder(right.Category)
		}
		return left.ID < right.ID
	})
	limit := (len(missing) + 1) / 2
	gained := make([]string, 0, limit)
	for _, techID := range missing[:limit] {
		attacker.Research.Completed[techID] = true
		if attacker.Research.ActiveID == techID {
			attacker.Research.ActiveID = ""
			attacker.Research.TurnsLeft = 0
		}
		delete(attacker.Research.PausedTurns, techID)
		gained = append(gained, techID)
	}
	return gained
}

func regionHasSettlement(region *world.Region, settlementID string) bool {
	if region == nil || settlementID == "" {
		return false
	}
	for _, settlement := range region.Settlements {
		if settlement.ID == settlementID {
			return true
		}
	}
	return false
}

func capitalSettlementName(region *world.Region, settlement *world.Settlement, fallback string) string {
	if settlement != nil {
		if settlement.NameTR != "" {
			return settlement.NameTR
		}
		if settlement.Name != "" {
			return settlement.Name
		}
	}
	if region != nil && region.NameTR != "" {
		return region.NameTR
	}
	return fallback
}

func factionDisplayName(gs *state.GameState, fid faction.FactionID) string {
	if gs == nil || fid == "" {
		return string(fid)
	}
	if f := gs.Factions[fid]; f != nil && f.NameTR != "" {
		return f.NameTR
	}
	return string(fid)
}

func formatCapitalLoot(loot map[economy.ResourceKind]int) string {
	if len(loot) == 0 {
		return ""
	}
	parts := make([]string, 0, len(loot))
	for _, kind := range economy.AllResourceKinds() {
		if amount := loot[kind]; amount > 0 {
			parts = append(parts, fmt.Sprintf("%s +%d", economy.ResourceNameTR(kind), amount))
		}
	}
	return strings.Join(parts, ", ")
}
