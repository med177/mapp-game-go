package game

import (
	"fmt"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
)

type battleArmySnapshot struct {
	Units                       int
	HP                          int
	Strength                    int
	CommanderName               string
	CommanderPortraitAsset      string
	CommanderBattleEffects      string
	CommanderOperationalEffects string
}

func snapshotBattleArmy(a *army.Army, types map[string]*army.UnitType) battleArmySnapshot {
	if a == nil {
		return battleArmySnapshot{}
	}
	snap := battleArmySnapshot{
		Units: len(a.Units),
		HP:    totalBattleUnitsHP(a.Units),
	}
	if types != nil {
		snap.Strength = a.TotalStrength(types)
	}
	if a.Commander != nil {
		snap.CommanderName = a.Commander.Name
		snap.CommanderPortraitAsset = a.Commander.PortraitAsset
		snap.CommanderBattleEffects = battleReportCommanderBattleEffects(a.Commander)
		snap.CommanderOperationalEffects = battleReportCommanderOperationalEffects(a.Commander)
	}
	return snap
}

func battleReportCommanderBattleEffects(commander *army.Commander) string {
	if commander == nil {
		return ""
	}
	parts := make([]string, 0, 3)
	if attack := commander.AttackModifier(); attack > 0 {
		parts = append(parts, fmt.Sprintf("Saldırı +%.0f%%", attack*100))
	}
	if defense := commander.DefenseModifier(); defense > 0 {
		parts = append(parts, fmt.Sprintf("Savunma +%.0f%%", defense*100))
	}
	if morale := commander.MoraleModifier(); morale > 0 {
		parts = append(parts, fmt.Sprintf("Moral +%.0f%%", morale*100))
	}
	return strings.Join(parts, "  |  ")
}

func battleReportCommanderOperationalEffects(commander *army.Commander) string {
	if commander == nil {
		return ""
	}
	parts := make([]string, 0, 2)
	if move := commander.MoveBonus(); move > 0 {
		parts = append(parts, fmt.Sprintf("Hareket +%d", move))
	}
	progress, breach := commander.SiegeBonuses()
	if progress > 0 || breach > 0 {
		parts = append(parts, fmt.Sprintf("Kuşatma +%d/+%d", progress, breach))
	}
	return strings.Join(parts, "  |  ")
}

func totalBattleUnitsHP(units []army.Unit) int {
	total := 0
	for i := range units {
		hp := units[i].CurrentHP
		if hp < 1 {
			continue
		}
		if hp > army.MaxUnitHP {
			hp = army.MaxUnitHP
		}
		total += hp
	}
	return total
}

func (g *Game) factionNameTR(ownerID string) string {
	if ownerID == "" || g == nil || g.gs == nil {
		return ownerID
	}
	if f := g.gs.Factions[faction.FactionID(ownerID)]; f != nil && f.NameTR != "" {
		return f.NameTR
	}
	return ownerID
}

func buildBattleReportSide(label, factionName string, before, after battleArmySnapshot) render.BattleReportSide {
	lostUnits := before.Units - after.Units
	if lostUnits < 0 {
		lostUnits = 0
	}
	hpDamage := before.HP - after.HP
	if hpDamage < 0 {
		hpDamage = 0
	}
	return render.BattleReportSide{
		Label:                       label,
		Faction:                     factionName,
		CommanderName:               before.CommanderName,
		CommanderPortraitAsset:      before.CommanderPortraitAsset,
		CommanderBattleEffects:      before.CommanderBattleEffects,
		CommanderOperationalEffects: before.CommanderOperationalEffects,
		StrengthBefore:              before.Strength,
		StrengthAfter:               after.Strength,
		UnitsBefore:                 before.Units,
		UnitsAfter:                  after.Units,
		UnitsLost:                   lostUnits,
		HPBefore:                    before.HP,
		HPAfter:                     after.HP,
		HPDamage:                    hpDamage,
	}
}

func battleReportSceneTitle(scene render.BattleScene) string {
	switch scene {
	case render.BattleSceneNaval:
		return "Deniz Muharebesi Raporu"
	case render.BattleSceneAmphibious:
		return "Çıkarma Muharebesi Raporu"
	case render.BattleSceneSiege:
		return "Kuşatma Hücumu Raporu"
	default:
		return "Kara Muharebesi Raporu"
	}
}

func battleReportEventPrefix(scene render.BattleScene) string {
	switch scene {
	case render.BattleSceneNaval:
		return "[DENIZ]"
	case render.BattleSceneAmphibious:
		return "[CIKARMA]"
	case render.BattleSceneSiege:
		return "[KUSATMA]"
	default:
		return "[MUHAREBE]"
	}
}

func battleReportDetailText(report render.BattleReport) string {
	var b strings.Builder
	title := report.Title
	if title == "" {
		title = battleReportSceneTitle(report.Scene)
	}
	b.WriteString(title)
	if report.RegionName != "" {
		b.WriteString("\nBölge: ")
		b.WriteString(report.RegionName)
	}
	if report.StanceLabel != "" {
		b.WriteString("\nDuruş: ")
		b.WriteString(report.StanceLabel)
	}
	if report.Outcome != "" {
		b.WriteString("\nSonuç: ")
		b.WriteString(report.Outcome)
	}
	if report.OutcomeDetail != "" {
		b.WriteString("\n")
		b.WriteString(report.OutcomeDetail)
	}
	appendSide := func(side render.BattleReportSide) {
		b.WriteString("\n\n")
		b.WriteString(side.Label)
		if side.Faction != "" {
			b.WriteString(" - ")
			b.WriteString(side.Faction)
		}
		b.WriteString(fmt.Sprintf("\nGüç: %d -> %d", side.StrengthBefore, side.StrengthAfter))
		b.WriteString(fmt.Sprintf("\nBirim: %d -> %d (kayıp %d)", side.UnitsBefore, side.UnitsAfter, side.UnitsLost))
		b.WriteString(fmt.Sprintf("\nHP: %d -> %d (hasar %d)", side.HPBefore, side.HPAfter, side.HPDamage))
	}
	appendSide(report.Attacker)
	appendSide(report.Defender)
	if len(report.CommanderProgress) > 0 {
		b.WriteString("\n\nKomutan gelişimi")
		for _, progress := range report.CommanderProgress {
			b.WriteString("\n")
			b.WriteString(renderBattleCommanderProgressText(progress))
		}
	}
	return b.String()
}

func renderBattleCommanderProgressText(progress render.BattleReportCommanderProgress) string {
	level := fmt.Sprintf("Lv%d", progress.CurrentLevel)
	if progress.PreviousLevel != progress.CurrentLevel {
		level = fmt.Sprintf("Lv%d -> Lv%d", progress.PreviousLevel, progress.CurrentLevel)
	}
	text := fmt.Sprintf("%s (%s): +%d XP | %s", progress.Name, progress.SideLabel, progress.XPGained, level)
	if len(progress.NewTraits) > 0 {
		text += " | Yeni: " + strings.Join(progress.NewTraits, ", ")
	}
	return text
}

func (g *Game) presentBattleReport(report render.BattleReport) {
	if g == nil || g.renderer == nil {
		return
	}
	if report.Title == "" {
		report.Title = battleReportSceneTitle(report.Scene)
	}
	g.renderer.ShowBattleReport(report)
	title := report.RegionName
	if title == "" {
		title = report.Title
	}
	g.renderer.AddEventDetail(battleReportEventPrefix(report.Scene)+" "+title+": "+report.Outcome, battleReportDetailText(report))
}

func (g *Game) makeBattleReport(scene render.BattleScene, regionName string, stance combat.BattleStance, outcome, detail string, attackerLabel, defenderLabel, attackerFaction, defenderFaction string, attackerBefore battleArmySnapshot, attackerAfter *army.Army, defenderBefore battleArmySnapshot, defenderAfter *army.Army) render.BattleReport {
	stanceLabel := ""
	if stance != "" {
		stanceLabel = combat.BattleStanceLabelTR(stance)
	}
	return g.makeBattleReportFromSnapshots(scene, regionName, stanceLabel, outcome, detail, attackerLabel, defenderLabel, attackerFaction, defenderFaction, attackerBefore, snapshotBattleArmy(attackerAfter, g.gs.UnitTypes), defenderBefore, snapshotBattleArmy(defenderAfter, g.gs.UnitTypes))
}

func (g *Game) makeBattleReportFromSnapshots(scene render.BattleScene, regionName, stanceLabel, outcome, detail, attackerLabel, defenderLabel, attackerFaction, defenderFaction string, attackerBefore, attackerAfter, defenderBefore, defenderAfter battleArmySnapshot) render.BattleReport {
	report := render.BattleReport{
		Scene:         scene,
		RegionName:    regionName,
		Title:         battleReportSceneTitle(scene),
		Outcome:       outcome,
		OutcomeDetail: detail,
		StanceLabel:   stanceLabel,
		Attacker:      buildBattleReportSide(attackerLabel, attackerFaction, attackerBefore, attackerAfter),
		Defender:      buildBattleReportSide(defenderLabel, defenderFaction, defenderBefore, defenderAfter),
	}
	if g != nil {
		report.CommanderProgress = append([]render.BattleReportCommanderProgress(nil), g.lastCommanderProgress...)
		g.lastCommanderProgress = nil
	}
	return report
}
