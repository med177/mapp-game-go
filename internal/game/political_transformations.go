package game

import (
	"fmt"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/events"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/world"
)

// resolvePoliticalTransformation event seçiminin tanımladığı siyasi
// dönüşümü state'e uygular. Dönüşüm kaydı bulunamazsa event'in geri kalan
// etkileri bozulmadan korunur.
func (g *Game) resolvePoliticalTransformation(effect events.Effect) {
	if g == nil || g.gs == nil || effect.PoliticalTransformationID == "" {
		return
	}
	var transformation *scenario.PoliticalTransformation
	for i := range g.gs.PoliticalTransformations {
		candidate := &g.gs.PoliticalTransformations[i]
		if candidate.ID == effect.PoliticalTransformationID {
			transformation = candidate
			break
		}
	}
	if transformation == nil || len(transformation.Members) < 2 {
		if g.renderer != nil {
			g.renderer.AddEvent("[SİYASİ DÖNÜŞÜM] Tanım bulunamadı: " + effect.PoliticalTransformationID)
		}
		return
	}

	resultID := faction.FactionID(transformation.ResultFaction)
	memberIDs := make([]faction.FactionID, 0, len(transformation.Members))
	for _, memberID := range transformation.Members {
		memberIDs = append(memberIDs, faction.FactionID(memberID))
	}
	if g.gs.Factions[resultID] == nil {
		return
	}
	switch transformation.Type {
	case "dynastic_civil_war":
		g.startPoliticalCivilWar(transformation, resultID, memberIDs)
	case "reunification", "dynastic_union":
		g.resolvePoliticalReunification(transformation, effect, resultID, memberIDs)
	case "dissolution":
		g.resolvePoliticalDissolution(transformation, resultID, memberIDs)
	}
}

func (g *Game) startPoliticalCivilWar(transformation *scenario.PoliticalTransformation, resultID faction.FactionID, memberIDs []faction.FactionID) {
	// İç savaşa katılan hanedanlar artık aynı vassallık düzeninin parçası
	// değildir. Özellikle Bolton ihaneti sonrasında eski overlord bağı ve
	// haraç ayarları state'te korunmamalıdır.
	for _, memberID := range memberIDs {
		member := g.gs.Factions[memberID]
		if member == nil || member.OverlordID == "" {
			continue
		}
		member.OverlordID = ""
		member.VassalizedTurn = 0
		member.TributeRate = 0
		member.TributeRateConfigured = false
	}
	g.gs.ActivatePoliticalMembers(resultID, memberIDs)
	regionTargets := make(map[world.RegionID]faction.FactionID, len(transformation.RegionAssignments))
	for regionID, targetID := range transformation.RegionAssignments {
		regionTargets[world.RegionID(regionID)] = faction.FactionID(targetID)
	}
	for regionID, targetID := range regionTargets {
		region := g.gs.Regions[regionID]
		if region == nil || g.gs.Factions[targetID] == nil {
			continue
		}
		region.OwnerID = string(targetID)
	}
	report := g.gs.SplitFactionForces([]faction.FactionID{resultID}, memberIDs, regionTargets)
	for _, memberID := range memberIDs {
		diplomacy.ForceRelation(g.gs, memberID, resultID, faction.StanceWar, -100)
	}
	if len(memberIDs) >= 2 {
		diplomacy.ForceRelation(g.gs, memberIDs[0], memberIDs[1], faction.StanceWar, -100)
	}
	g.gs.NormalizeFactionCapitals()
	if g.renderer != nil {
		g.renderer.AddEventDetail("[SİYASİ DÖNÜŞÜM] "+transformation.DescriptionTR,
			fmt.Sprintf("%s ordusu: %d, donanması: %d; dönüşüme katılan hanedan: %d.",
				g.factionNameTR(string(resultID)), report.TransferredArmies, report.TransferredFleets, len(memberIDs)))
	}
}

func (g *Game) resolvePoliticalReunification(transformation *scenario.PoliticalTransformation, effect events.Effect, resultID faction.FactionID, memberIDs []faction.FactionID) {
	winnerID := faction.FactionID(effect.PoliticalWinnerFactionID)
	if effect.PoliticalWinnerFromPlayer && containsFactionID(memberIDs, g.gs.PlayerFactionID) {
		winnerID = g.gs.PlayerFactionID
	}
	if winnerID != "" && !containsFactionID(memberIDs, winnerID) {
		return
	}
	playerWasMember := containsFactionID(memberIDs, g.gs.PlayerFactionID)
	if transformation.Transfer.Relations == "merge" {
		g.gs.MergePoliticalRelations(memberIDs, resultID)
	}
	report := g.gs.UnitePoliticalFactions(memberIDs, resultID)
	if playerWasMember && g.gs.Factions[resultID] != nil && !g.gs.Factions[resultID].IsEliminated {
		g.gs.PlayerFactionID = resultID
	}
	if g.renderer == nil {
		return
	}
	winnerLabel := "hanedanların birleşmesi"
	if winnerID != "" {
		winnerLabel = g.factionNameTR(string(winnerID)) + " zaferi"
	}
	g.renderer.AddEventDetail("[SİYASİ BİRLEŞME] "+winnerLabel,
		fmt.Sprintf("Sonuç faction: %s; aktarılan kara ordusu: %d, donanma: %d.",
			g.factionNameTR(string(resultID)), report.TransferredArmies, report.TransferredFleets))
}

func (g *Game) resolvePoliticalDissolution(transformation *scenario.PoliticalTransformation, resultID faction.FactionID, memberIDs []faction.FactionID) {
	regionTargets := make(map[world.RegionID]faction.FactionID, len(transformation.RegionAssignments))
	for regionID, targetID := range transformation.RegionAssignments {
		regionTargets[world.RegionID(regionID)] = faction.FactionID(targetID)
	}
	report := g.gs.DissolvePoliticalFaction(resultID, memberIDs, regionTargets)
	if g.renderer == nil {
		return
	}
	g.renderer.AddEventDetail("[SİYASİ DAĞILMA] "+transformation.DescriptionTR,
		fmt.Sprintf("Yeniden etkinleşen faction: %d; aktarılan kara ordusu: %d, donanma: %d.",
			len(memberIDs), report.TransferredArmies, report.TransferredFleets))
}

func containsFactionID(ids []faction.FactionID, target faction.FactionID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
