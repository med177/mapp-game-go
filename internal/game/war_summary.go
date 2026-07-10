package game

import (
	"fmt"
	"strings"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
)

func (g *Game) buildWarSummary(targetID faction.FactionID, result diplomacy.WarDeclarationResult) render.WarSummaryReport {
	attackerRoot := diplomacy.RealmRoot(g.gs, g.gs.PlayerFactionID)
	if attackerRoot == "" {
		attackerRoot = g.gs.PlayerFactionID
	}
	defenderRoot := diplomacy.RealmRoot(g.gs, targetID)
	if defenderRoot == "" {
		defenderRoot = targetID
	}

	attacker := g.buildWarSummarySide("Saldıran Cephe", attackerRoot, result.PlayerCalls)
	defender := g.buildWarSummarySide("Savunan Cephe", defenderRoot, result.EnemyCalls)
	return render.WarSummaryReport{
		Title:        g.factionNameTR(string(defenderRoot)) + " Savaşı Özeti",
		BalanceLabel: warBalanceLabelTR(attacker.TotalStrength, defender.TotalStrength),
		PowerText:    fmt.Sprintf("%s %d güç • %s %d güç", attacker.LeaderNameTR, attacker.TotalStrength, defender.LeaderNameTR, defender.TotalStrength),
		Attacker:     attacker,
		Defender:     defender,
	}
}

func (g *Game) buildWarSummarySide(label string, primaryRoot faction.FactionID, calls []diplomacy.WarCallOutcome) render.WarSummarySide {
	side := render.WarSummarySide{
		Label:         label,
		LeaderNameTR:  g.factionNameTR(string(primaryRoot)),
		Participants:  make([]render.WarSummaryParticipant, 0, 8),
		Refused:       make([]string, 0, 4),
		TotalStrength: 0,
	}
	seen := map[faction.FactionID]struct{}{}
	addRealm := func(root faction.FactionID, rootRole, vassalRole string) {
		for _, memberID := range diplomacy.WarCoalitionMembers(g.gs, root) {
			if _, exists := seen[memberID]; exists {
				continue
			}
			seen[memberID] = struct{}{}
			role := vassalRole
			if memberID == root {
				role = rootRole
			}
			strength := diplomacy.MilitaryPower(g.gs, memberID)
			side.TotalStrength += strength
			side.Participants = append(side.Participants, render.WarSummaryParticipant{
				NameTR:   g.factionNameTR(string(memberID)),
				RoleTR:   role,
				Strength: strength,
			})
		}
	}

	addRealm(primaryRoot, "Lider", "Vassal")
	for _, outcome := range calls {
		if outcome.Joined {
			addRealm(outcome.FactionID, "Müttefik", "Müttefik Vassalı")
			continue
		}
		side.Refused = append(side.Refused, outcome.NameTR)
	}
	return side
}

func warBalanceLabelTR(attackerStrength, defenderStrength int) string {
	switch {
	case attackerStrength == defenderStrength:
		return "Güç dengesi: başa baş"
	case attackerStrength == 0 && defenderStrength > 0:
		return "Güç dengesi: savunan ezici üstün"
	case defenderStrength == 0 && attackerStrength > 0:
		return "Güç dengesi: saldıran ezici üstün"
	case attackerStrength > defenderStrength*12/10:
		return "Güç dengesi: saldıran üstün"
	case defenderStrength > attackerStrength*12/10:
		return "Güç dengesi: savunan üstün"
	default:
		return "Güç dengesi: yakın"
	}
}

func warSummaryDetailText(report render.WarSummaryReport) string {
	var b strings.Builder
	b.WriteString(report.Title)
	b.WriteString("\n")
	b.WriteString(report.BalanceLabel)
	b.WriteString("\n")
	b.WriteString(report.PowerText)
	appendSide := func(side render.WarSummarySide) {
		b.WriteString("\n\n")
		b.WriteString(side.Label)
		b.WriteString(" - ")
		b.WriteString(side.LeaderNameTR)
		b.WriteString(fmt.Sprintf("\nToplam Güç: %d", side.TotalStrength))
		for _, participant := range side.Participants {
			b.WriteString("\n- ")
			b.WriteString(participant.NameTR)
			if participant.RoleTR != "" {
				b.WriteString(" (")
				b.WriteString(participant.RoleTR)
				b.WriteString(")")
			}
			b.WriteString(": ")
			b.WriteString(fmt.Sprintf("%d", participant.Strength))
		}
		if len(side.Refused) > 0 {
			b.WriteString("\nKatılmayan: ")
			b.WriteString(strings.Join(side.Refused, ", "))
		}
	}
	appendSide(report.Attacker)
	appendSide(report.Defender)
	return b.String()
}
