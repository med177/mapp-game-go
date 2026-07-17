package render

import (
	"fmt"
	"image/color"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func siegeBreachLabelTR(level int) string {
	switch level {
	case 2:
		return "Büyük gedik"
	case 1:
		return "Küçük gedik"
	default:
		return "Gedik yok"
	}
}

func (r *Renderer) openSiegeDecision(attacker *army.Army, target *world.Region) {
	if r == nil || r.gs == nil || attacker == nil || target == nil {
		return
	}
	fortLevel := target.FortificationLevel()
	bestTier := attacker.HighestSiegeTier(r.gs.UnitTypes)
	hasSiegeUnits := attacker.HasSiegeUnits(r.gs.UnitTypes)
	commanderSummary := commanderSiegeSummary(attacker.Commander)
	if active := r.gs.SiegeAt(target.ID); active != nil && active.AttackerArmyID == attacker.ID {
		msg := fmt.Sprintf("%s kuşatması sürüyor. Tahkimat: %d | İlerleme: %d | Durum: %s | Gedik kapasitesi: T%d/T%d. %s", target.NameTR, active.FortLevel, active.BreachProgress, siegeBreachLabelTR(active.BreachLevel), bestTier, active.FortLevel, commanderSummary)
		r.confirmDialog = confirmDialogState{
			show:          true,
			title:         "Kuşatma Kararı",
			message:       msg,
			messageLines:  wrapTextLines(msg, FaceSmall, float64(confirmDialogW)-40),
			acceptLabel:   "Genel Hücum",
			thirdLabel:    "Kuşatmayı Kaldır",
			declineLabel:  "İptal",
			pendingAction: InputAction{Kind: ActionAssaultSiege, ArmyID: attacker.ID, TargetRegion: target.ID, BattleStance: combat.BattleStanceBalanced},
			thirdAction:   InputAction{Kind: ActionLiftSiege, ArmyID: attacker.ID, TargetRegion: target.ID},
		}
		return
	}
	msg := fmt.Sprintf("%s tahkimli. Tahkimat seviyesi: %d | Kuşatma gücü: %d | Gedik kapasitesi: T%d/T%d. %s", target.NameTR, fortLevel, attacker.SiegeUnitScore(r.gs.UnitTypes), bestTier, fortLevel, commanderSummary)
	thirdLabel := "Genel Hücum"
	if !hasSiegeUnits {
		msg += " Bu ordu kuşatma birimi taşımıyor; yalnız kuşatma kurup teslim bekleyebilirsin."
		thirdLabel = ""
	}
	r.confirmDialog = confirmDialogState{
		show:          true,
		title:         "Kuşatma Kararı",
		message:       msg,
		messageLines:  wrapTextLines(msg, FaceSmall, float64(confirmDialogW)-40),
		acceptLabel:   "Kuşatma Başlat",
		thirdLabel:    thirdLabel,
		declineLabel:  "İptal",
		pendingAction: InputAction{Kind: ActionStartSiege, ArmyID: attacker.ID, TargetRegion: target.ID},
		thirdAction:   InputAction{Kind: ActionAssaultSiege, ArmyID: attacker.ID, TargetRegion: target.ID, BattleStance: combat.BattleStanceBalanced},
	}
}

func (r *Renderer) selectedSiegePanelState() (*army.Army, *state.SiegeState, *world.Region, bool) {
	if r == nil || r.gs == nil || r.gs.Phase != state.PhasePlayerTurn || r.SelectedArmy == "" {
		return nil, nil, nil, false
	}
	if r.confirmDialog.show || r.warConfirm.show || r.warSummary.show || r.battlePlan.show || r.showHistoricalEvent ||
		r.battleReport.show || r.eventDetail != "" || r.showVictoryDetail || r.showEventCodex || r.showDiplomacy ||
		r.showTech || r.showTrade {
		return nil, nil, nil, false
	}
	attacker := r.gs.Armies[r.SelectedArmy]
	if attacker == nil || attacker.OwnerID != string(r.gs.PlayerFactionID) {
		return nil, nil, nil, false
	}
	siege := r.gs.SiegeByArmy(attacker.ID)
	if siege == nil {
		return nil, nil, nil, false
	}
	target := r.gs.Regions[siege.RegionID]
	if target == nil {
		return nil, nil, nil, false
	}
	return attacker, siege, target, true
}

func buildSelectedSiegePanel() gameui.Panel {
	x := (ScreenWidth - selectedSiegePanelW) / 2
	// Ordu detay paneli ekranın altına sabitlendiği için kuşatma panelini
	// onun üstüne yerleştir; iki panel aynı dikey alanı paylaşmasın.
	y := float64(armyPanelGeometry().panelY) - selectedSiegePanelH - 16
	if y < 8 {
		y = 8
	}
	return gameui.NewPanel(x, y, selectedSiegePanelW, selectedSiegePanelH)
}

func buildSelectedSiegeButtons() (gameui.Button, gameui.Button) {
	panel := buildSelectedSiegePanel()
	btnY := panel.Rect.Y + panel.Rect.H - selectedSiegeButtonH - 14
	gap := 14.0
	totalW := selectedSiegeButtonW*2 + gap
	startX := panel.Rect.X + (panel.Rect.W-totalW)/2
	assaultBtn := gameui.NewButton(startX, btnY, selectedSiegeButtonW, selectedSiegeButtonH, "Genel Hücum").WithIcon(gameui.IconSword)
	liftBtn := gameui.NewButton(startX+selectedSiegeButtonW+gap, btnY, selectedSiegeButtonW, selectedSiegeButtonH, "Kuşatmayı Kaldır").WithIcon(gameui.IconExit)
	return assaultBtn, liftBtn
}

func (r *Renderer) selectedSiegePanelHit(fx, fy float64) bool {
	if _, _, _, ok := r.selectedSiegePanelState(); !ok {
		return false
	}
	return buildSelectedSiegePanel().HitTest(fx, fy)
}

func (r *Renderer) selectedSiegePanelHovering(fx, fy float64) bool {
	if _, _, _, ok := r.selectedSiegePanelState(); !ok {
		return false
	}
	assaultBtn, liftBtn := buildSelectedSiegeButtons()
	attacker := r.gs.Armies[r.SelectedArmy]
	return (attacker != nil && attacker.HasSiegeUnits(r.gs.UnitTypes) && assaultBtn.HitTest(fx, fy)) || liftBtn.HitTest(fx, fy)
}

func (r *Renderer) drawSelectedSiegePanel(screen *ebiten.Image) {
	attacker, siege, target, ok := r.selectedSiegePanelState()
	if !ok {
		return
	}
	panel := buildSelectedSiegePanel()
	gameui.DrawPanel(screen, panel, gameui.PanelStyle{
		BG:          color.RGBA{14, 10, 7, 246},
		Border:      color.RGBA{184, 140, 62, 255},
		BorderWidth: 2,
	})
	vector.FillRect(screen, float32(panel.Rect.X), float32(panel.Rect.Y), float32(panel.Rect.W), 4, color.RGBA{205, 160, 66, 255}, false)
	drawUILabel(screen, gameui.Rect{X: panel.Rect.X + 18, Y: panel.Rect.Y + 11}, "Kuşatma Emri", ColorYellow, gameui.TextLarge, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: panel.Rect.X + panel.Rect.W - 18 - 120, Y: panel.Rect.Y + 15, W: 120}, "AKTİF KUŞATMA", color.RGBA{220, 166, 72, 220}, gameui.TextSmall, gameui.TextAlignEnd)

	drawUILabel(screen, gameui.Rect{X: panel.Rect.X + 18, Y: panel.Rect.Y + 42}, "Hedef", color.RGBA{164, 132, 76, 255}, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: panel.Rect.X + 74, Y: panel.Rect.Y + 40, W: panel.Rect.W - 92}, trimTextToWidth(target.NameTR, FaceMed, panel.Rect.W-92), color.RGBA{244, 238, 222, 255}, gameui.TextMedium, gameui.TextAlignStart)

	bestTier := attacker.HighestSiegeTier(r.gs.UnitTypes)
	status := siegeBreachLabelTR(siege.BreachLevel)
	statusColor := color.RGBA{204, 184, 126, 255}
	if siege.BreachLevel > 0 {
		statusColor = color.RGBA{232, 170, 76, 255}
	}
	canAssault := attacker.HasSiegeUnits(r.gs.UnitTypes)
	statusRect := gameui.Rect{X: panel.Rect.X + 16, Y: panel.Rect.Y + 64, W: panel.Rect.W - 32, H: 36}
	drawUICardRect(screen, statusRect, color.RGBA{38, 28, 15, 238}, color.RGBA{128, 96, 42, 220}, 1)
	drawUILabel(screen, gameui.Rect{X: statusRect.X + 12, Y: statusRect.Y + 9}, "DURUM", color.RGBA{226, 185, 92, 255}, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: statusRect.X + 76, Y: statusRect.Y + 8, W: statusRect.W - 88}, status, statusColor, gameui.TextMedium, gameui.TextAlignStart)
	remainingText := fmt.Sprintf("Teslim için yaklaşık %d tur", siege.TurnsUntilSurrender())
	if siege.BreachLevel >= 2 && siege.DefenderArmyID == "" {
		remainingText = "Büyük gedik: teslim olabilir"
	}
	drawUILabel(screen, gameui.Rect{X: statusRect.X + statusRect.W - 220, Y: statusRect.Y + 10, W: 208}, remainingText, color.RGBA{190, 208, 170, 245}, gameui.TextSmall, gameui.TextAlignEnd)

	metricY := panel.Rect.Y + 112
	metricW := (panel.Rect.W - 44) / 2
	drawSelectedSiegeMetric(screen, gameui.Rect{X: panel.Rect.X + 16, Y: metricY, W: metricW, H: 46}, "TAHKİMAT", fmt.Sprintf("T%d", siege.FortLevel), ColorWhite)
	drawSelectedSiegeMetric(screen, gameui.Rect{X: panel.Rect.X + 28 + metricW, Y: metricY, W: metricW, H: 46}, "İLERLEME", itoa(siege.BreachProgress), color.RGBA{238, 210, 138, 255})
	drawSelectedSiegeMetric(screen, gameui.Rect{X: panel.Rect.X + 16, Y: metricY + 54, W: metricW, H: 46}, "GEDİK", status, statusColor)
	assaultValue := fmt.Sprintf("T%d / T%d", bestTier, siege.FortLevel)
	assaultColor := color.RGBA{202, 222, 190, 255}
	if !canAssault {
		assaultValue = "KULLANILAMAZ"
		assaultColor = color.RGBA{214, 130, 112, 255}
	}
	drawSelectedSiegeMetric(screen, gameui.Rect{X: panel.Rect.X + 28 + metricW, Y: metricY + 54, W: metricW, H: 46}, "GENEL HÜCUM", assaultValue, assaultColor)

	commanderInfo := trimTextToWidth(commanderSiegeSummary(attacker.Commander), FaceSmall, panel.Rect.W-36)
	drawUILabel(screen, gameui.Rect{X: panel.Rect.X + 18, Y: panel.Rect.Y + 216, W: panel.Rect.W - 36}, commanderInfo, color.RGBA{154, 190, 220, 255}, gameui.TextSmall, gameui.TextAlignStart)
	drawUISeparator(screen, float32(panel.Rect.X+16), float32(panel.Rect.Y+242), float32(panel.Rect.X+panel.Rect.W-16), 1, color.RGBA{96, 72, 38, 190})
	assaultBtn, liftBtn := buildSelectedSiegeButtons()
	if canAssault {
		drawUIButtonWidget(screen, assaultBtn, solidButtonStyle(color.RGBA{70, 140, 70, 240}, color.RGBA{120, 180, 120, 255}, ColorWhite, 10))
	} else {
		drawUIButtonWidget(screen, assaultBtn, solidButtonStyle(color.RGBA{72, 70, 64, 220}, color.RGBA{104, 102, 96, 255}, color.RGBA{184, 180, 172, 255}, 10))
	}
	drawUIButtonWidget(screen, liftBtn, solidButtonStyle(color.RGBA{145, 95, 45, 235}, color.RGBA{190, 135, 75, 255}, ColorWhite, 10))
}

func drawSelectedSiegeMetric(screen *ebiten.Image, rect gameui.Rect, label, value string, valueColor color.RGBA) {
	drawUICardRect(screen, rect, color.RGBA{23, 20, 15, 238}, color.RGBA{86, 70, 42, 220}, 1)
	drawUILabel(screen, gameui.Rect{X: rect.X + 10, Y: rect.Y + 7}, label, color.RGBA{157, 133, 86, 255}, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: rect.X + 10, Y: rect.Y + 23, W: rect.W - 20}, value, valueColor, gameui.TextMedium, gameui.TextAlignStart)
}

func battlePlanInstructionTR(context combat.BattleContext) string {
	switch combat.NormalizeBattleContext(context) {
	case combat.BattleContextNaval:
		return "Bir duruş seçin. Alt satırdaki tahminler mevcut filo gücü, modlar ve zar aralığına göre hesaplanır."
	case combat.BattleContextAmphibious:
		return "Bir duruş seçin. Alt satırdaki tahminler çıkarma gücü, kıyı savunması ve zar aralığına göre hesaplanır."
	default:
		return "Bir duruş seçin. Alt satırdaki tahminler mevcut güç, arazi ve zar aralığına göre hesaplanır."
	}
}

func (r *Renderer) openBattlePlan(attacker *army.Army, target *world.Region, defender *army.Army, actionKind ActionKind, battleContext combat.BattleContext) {
	if target == nil {
		return
	}
	r.openBattlePlanWithDestination(attacker, target, defender, actionKind, battleContext, target.ID)
}

// openBattlePlanWithDestination savaşın gerçekleştiği bölge ile zaferden
// sonra ordunun ilerleyeceği bölge farklı olduğunda kullanılır. Huruçta savaş
// kuşatılan bölgede gerçekleşir, ancak başarılı sonuçta seçilen hedefe çıkılır.
func (r *Renderer) openBattlePlanWithDestination(attacker *army.Army, target *world.Region, defender *army.Army, actionKind ActionKind, battleContext combat.BattleContext, destination world.RegionID) {
	if r == nil || r.gs == nil || attacker == nil || target == nil || defender == nil {
		return
	}
	previewAttacker := attacker
	previewCommander := attacker.Commander
	if combat.NormalizeBattleContext(battleContext) == combat.BattleContextAmphibious {
		previewCommander = r.gs.AmphibiousCommander(attacker.ID)
		previewAttacker = &army.Army{
			OwnerID:   attacker.OwnerID,
			Units:     attacker.EmbarkedUnits,
			Commander: previewCommander,
		}
	}
	atkMods := combat.TechModsFor(r.gs, attacker.OwnerID)
	defMods := combat.TechModsFor(r.gs, defender.OwnerID)
	state := battlePlanState{
		show:            true,
		actionKind:      actionKind,
		battleContext:   combat.NormalizeBattleContext(battleContext),
		pendingArmy:     attacker.ID,
		pendingEnemy:    defender.ID,
		pendingDest:     destination,
		regionName:      target.NameTR,
		attackerSummary: commanderBattlePlanSummary("Saldıran komutan", previewCommander),
		defenderSummary: commanderBattlePlanSummary("Savunan komutan", defender.Commander),
		focus:           1,
	}
	for i, stance := range battlePlanStances {
		state.previews[i] = combat.PreviewBattleWithContextMods(previewAttacker, defender, target.Terrain, r.gs.UnitTypes, atkMods, defMods, state.battleContext, stance)
	}
	if factionInfo := r.gs.Factions[faction.FactionID(defender.OwnerID)]; factionInfo != nil {
		state.defenderFaction = factionInfo.NameTR
	} else {
		state.defenderFaction = defender.OwnerID
	}
	state.defenderName = "Savunan: " + state.defenderFaction
	if target.NameTR != "" {
		state.regionName = target.NameTR
	}
	r.battlePlan = state
}

func battlePlanLossText(expected, minLoss, maxLoss int) string {
	if minLoss == maxLoss {
		return itoa(expected)
	}
	return itoa(expected) + " (" + itoa(minLoss) + "-" + itoa(maxLoss) + ")"
}

func battlePlanHPText(expected, minLoss, maxLoss int) string {
	if minLoss == maxLoss {
		return "~" + itoa(expected) + " HP"
	}
	return "~" + itoa(expected) + " HP (" + itoa(minLoss) + "-" + itoa(maxLoss) + ")"
}

// --- Savaş ilan onay diyalogu ---

func warConfirmSideRects(modal gameui.Modal) (gameui.Rect, gameui.Rect) {
	panel := modal.Panel.Rect
	const (
		pad = 22.0
		gap = 18.0
	)
	contentY := panel.Y + 118
	contentH := panel.H - 188
	sideW := (panel.W - pad*2 - gap) / 2
	left := gameui.Rect{X: panel.X + pad, Y: contentY, W: sideW, H: contentH}
	right := gameui.Rect{X: left.X + sideW + gap, Y: contentY, W: sideW, H: contentH}
	return left, right
}

func warConfirmContentRect(modal gameui.Modal) gameui.Rect {
	panel := modal.Panel.Rect
	return gameui.Rect{
		X: panel.X + 22,
		Y: panel.Y + 118,
		W: panel.W - 44,
		H: panel.H - 188,
	}
}

func warConfirmScrollHit(modal gameui.Modal, mx, my float64) bool {
	if !modal.Panel.Rect.Hit(mx, my) {
		return false
	}
	acceptBtn, declineBtn := buildWarConfirmButtons()
	return !acceptBtn.HitTest(mx, my) && !declineBtn.HitTest(mx, my)
}

func warConfirmCallHeaderY(sideRect gameui.Rect, autoCount int) float64 {
	base := sideRect.Y + 182
	y := sideRect.Y + 90 + float64(autoCount)*60 + 10
	if y > base {
		return y
	}
	return base
}

func warConfirmCallViewport(sideRect gameui.Rect, autoCount int) gameui.Rect {
	headerY := warConfirmCallHeaderY(sideRect, autoCount)
	top := headerY + 26
	bottom := sideRect.Y + sideRect.H - 10
	if bottom < top {
		bottom = top
	}
	return gameui.Rect{
		X: sideRect.X + 10,
		Y: top,
		W: sideRect.W - 20,
		H: bottom - top,
	}
}

func warConfirmVisibleRows(viewport gameui.Rect) int {
	rows := int(viewport.H / 52)
	if rows < 1 {
		return 1
	}
	return rows
}

func warConfirmSharedMaxScroll(attackerEntries, defenderEntries []diplomacy.WarParticipantPreview, attackerViewport, defenderViewport gameui.Rect) int {
	attackerMax := warConfirmMaxScroll(len(attackerEntries), attackerViewport)
	defenderMax := warConfirmMaxScroll(len(defenderEntries), defenderViewport)
	return maxScreenInt(attackerMax, defenderMax)
}

func warConfirmSharedVisibleRows(attackerViewport, defenderViewport gameui.Rect) int {
	return minScreenInt(warConfirmVisibleRows(attackerViewport), warConfirmVisibleRows(defenderViewport))
}

func warConfirmMaxScroll(entryCount int, viewport gameui.Rect) int {
	maxScroll := entryCount - warConfirmVisibleRows(viewport)
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func clampWarConfirmScroll(entryCount int, viewport gameui.Rect, scroll int) int {
	if scroll < 0 {
		return 0
	}
	maxScroll := warConfirmMaxScroll(entryCount, viewport)
	if scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func warConfirmEntryRowRect(viewport gameui.Rect, visibleIndex int) gameui.Rect {
	return gameui.Rect{
		X: viewport.X,
		Y: viewport.Y + float64(visibleIndex*52),
		W: viewport.W,
		H: 44,
	}
}

func warConfirmCheckboxes(viewport gameui.Rect, entries []diplomacy.WarParticipantPreview, selected map[faction.FactionID]bool, scroll int) []gameui.Checkbox {
	if len(entries) == 0 {
		return nil
	}
	if scroll < 0 {
		scroll = 0
	}
	if scroll >= len(entries) {
		return nil
	}
	visibleRows := warConfirmVisibleRows(viewport)
	end := scroll + visibleRows
	if end > len(entries) {
		end = len(entries)
	}
	checkboxes := make([]gameui.Checkbox, 0, end-scroll)
	for i := scroll; i < end; i++ {
		rowRect := warConfirmEntryRowRect(viewport, i-scroll)
		box := gameui.NewCheckbox(rowRect.X+2, rowRect.Y, rowRect.W-4, rowRect.H, "")
		box.Checked = selected[entries[i].FactionID]
		checkboxes = append(checkboxes, box)
	}
	return checkboxes
}

func warConfirmEntryIndexAt(viewport gameui.Rect, entries []diplomacy.WarParticipantPreview, scroll int, mx, my float64) int {
	if len(entries) == 0 || !viewport.Hit(mx, my) {
		return -1
	}
	if scroll < 0 {
		scroll = 0
	}
	for visibleIdx, checkbox := range warConfirmCheckboxes(viewport, entries, nil, scroll) {
		if checkbox.HitTest(mx, my) {
			return scroll + visibleIdx
		}
	}
	return -1
}

func drawWarConfirmScrollbar(screen *ebiten.Image, track gameui.Rect, visibleRows, maxScroll, scroll int) {
	if maxScroll <= 0 {
		return
	}
	drawUICardRect(screen, track, color.RGBA{22, 20, 16, 210}, color.RGBA{72, 62, 42, 180}, 1)
	thumbH := track.H * float64(visibleRows) / float64(visibleRows+maxScroll)
	if thumbH < 24 {
		thumbH = 24
	}
	thumbY := track.Y
	if track.H > thumbH {
		thumbY += (track.H - thumbH) * float64(scroll) / float64(maxScroll)
	}
	drawUICardRect(screen, gameui.Rect{X: track.X, Y: thumbY, W: track.W, H: thumbH}, color.RGBA{176, 144, 78, 230}, color.RGBA{214, 190, 120, 210}, 1)
}

func warConfirmStatusText(entry diplomacy.WarParticipantPreview) string {
	if entry.AutoJoin {
		return "Kesin"
	}
	if entry.JoinChance <= 0 {
		return entry.StatusTR
	}
	return entry.StatusTR + " · %" + itoa(entry.JoinChance)
}

func drawWarConfirmParticipantRow(screen *ebiten.Image, rowRect gameui.Rect, entry diplomacy.WarParticipantPreview, selected bool, selectable bool) {
	fill := color.RGBA{24, 20, 14, 228}
	border := color.RGBA{84, 68, 42, 220}
	if selectable && selected {
		fill = color.RGBA{46, 54, 28, 236}
		border = color.RGBA{156, 182, 102, 245}
	}
	drawUICardRect(screen, rowRect, fill, border, 1)

	textX := rowRect.X + 14
	if selectable {
		checkbox := gameui.NewCheckbox(rowRect.X+12, rowRect.Y+10, rowRect.W-24, 24, "")
		checkbox.Checked = selected
		gameui.DrawCheckbox(screen, checkbox, gameui.CheckboxStyle{
			BoxBG:        color.RGBA{12, 14, 10, 255},
			BoxBorder:    color.RGBA{140, 152, 108, 255},
			CheckColor:   color.RGBA{176, 212, 116, 255},
			TextColor:    ColorWhite,
			DisabledText: ColorGray,
			BoxSize:      16,
			TextOffsetY:  0,
			TextVariant:  gameui.TextSmall,
			BorderWidth:  1,
		}, renderText)
		textX += 28
	}

	drawUILabel(screen, gameui.Rect{X: textX, Y: rowRect.Y + 8, W: rowRect.W - (textX - rowRect.X) - 110}, entry.NameTR, ColorWhite, gameui.TextMedium, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: rowRect.X + rowRect.W - 108, Y: rowRect.Y + 8, W: 96}, entry.RoleTR, color.RGBA{190, 170, 108, 255}, gameui.TextSmall, gameui.TextAlignEnd)
	drawUILabel(screen, gameui.Rect{X: textX, Y: rowRect.Y + 24, W: rowRect.W - (textX - rowRect.X) - 10}, warConfirmStatusText(entry), color.RGBA{176, 200, 164, 255}, gameui.TextSmall, gameui.TextAlignStart)
	if entry.NoteTR != "" {
		drawUILabel(screen, gameui.Rect{X: textX, Y: rowRect.Y + 38, W: rowRect.W - (textX - rowRect.X) - 10}, entry.NoteTR, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	}
}

func drawWarConfirmSide(screen *ebiten.Image, sideRect gameui.Rect, title, leaderName, emptyAutoText, emptyCallText string, autoEntries, callableEntries []diplomacy.WarParticipantPreview, selected map[faction.FactionID]bool, selectable bool, scroll int) {
	drawUICardRect(screen, sideRect, color.RGBA{18, 14, 10, 232}, color.RGBA{94, 74, 42, 210}, 1)
	drawUILabel(screen, gameui.Rect{X: sideRect.X + 14, Y: sideRect.Y + 12, W: sideRect.W - 28}, title, color.RGBA{255, 220, 100, 255}, gameui.TextMedium, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: sideRect.X + 14, Y: sideRect.Y + 34, W: sideRect.W - 28}, "Lider: "+leaderName, ColorGray, gameui.TextSmall, gameui.TextAlignStart)

	drawUILabel(screen, gameui.Rect{X: sideRect.X + 14, Y: sideRect.Y + 66, W: sideRect.W - 28}, "Kesin Katılanlar", color.RGBA{194, 184, 136, 255}, gameui.TextSmall, gameui.TextAlignStart)
	y := sideRect.Y + 90
	if len(autoEntries) == 0 {
		drawUILabel(screen, gameui.Rect{X: sideRect.X + 14, Y: y + 8, W: sideRect.W - 28}, emptyAutoText, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	} else {
		for _, entry := range autoEntries {
			rowRect := gameui.Rect{X: sideRect.X + 10, Y: y, W: sideRect.W - 20, H: 52}
			drawWarConfirmParticipantRow(screen, rowRect, entry, false, false)
			y += 60
		}
	}

	callHeaderY := warConfirmCallHeaderY(sideRect, len(autoEntries))
	drawUILabel(screen, gameui.Rect{X: sideRect.X + 14, Y: callHeaderY, W: sideRect.W - 28}, "Çağrılabilir Müttefikler", color.RGBA{194, 184, 136, 255}, gameui.TextSmall, gameui.TextAlignStart)
	if selectable {
		drawUILabel(screen, gameui.Rect{X: sideRect.X + 14, Y: callHeaderY + 16, W: sideRect.W - 28}, "Seçip de gelmeyen müttefik ittifakı bozar ve ilişki düşürür.", color.RGBA{174, 146, 118, 255}, gameui.TextSmall, gameui.TextAlignStart)
	}
	viewport := warConfirmCallViewport(sideRect, len(autoEntries))
	if len(callableEntries) == 0 {
		drawUILabel(screen, gameui.Rect{X: sideRect.X + 14, Y: viewport.Y + 8, W: sideRect.W - 28}, emptyCallText, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		return
	}
	visibleRows := warConfirmVisibleRows(viewport)
	end := scroll + visibleRows
	if end > len(callableEntries) {
		end = len(callableEntries)
	}
	for i := scroll; i < end; i++ {
		rowRect := warConfirmEntryRowRect(viewport, i-scroll)
		drawWarConfirmParticipantRow(screen, rowRect, callableEntries[i], selected[callableEntries[i].FactionID], selectable)
	}
}

func selectedWarAlliesFromState(wc warConfirmState) []faction.FactionID {
	if len(wc.preview.Attacker.CallableAllies) == 0 || len(wc.selectedAllies) == 0 {
		return nil
	}
	selected := make([]faction.FactionID, 0, len(wc.selectedAllies))
	for _, entry := range wc.preview.Attacker.CallableAllies {
		if wc.selectedAllies[entry.FactionID] {
			selected = append(selected, entry.FactionID)
		}
	}
	return selected
}

func (r *Renderer) finalizeWarConfirm(wc warConfirmState) InputAction {
	action := InputAction{
		Kind:          ActionDeclareWar,
		TargetFaction: faction.FactionID(wc.factionID),
		WarAllies:     selectedWarAlliesFromState(wc),
	}
	if wc.pendingArmy == "" || wc.pendingDest == "" {
		return action
	}

	attacker := r.gs.Armies[wc.pendingArmy]
	target := r.gs.Regions[wc.pendingDest]
	isAlliedRegion := armyRegionIsFriendly(r.gs, attacker, target)
	supportingSiege := r.canJoinActiveSiege(attacker, wc.pendingDest)
	canEnterSiege := r.canEnterActiveSiegedRegion(attacker, wc.pendingDest)
	activeSiege := r.gs.SiegeAt(wc.pendingDest)
	if attacker != nil && r.gs.IsArmyDefendingSiegedRegion(attacker) {
		if siege := r.gs.SiegeAt(attacker.RegionID); siege != nil {
			if siegeArmy := r.gs.Armies[siege.AttackerArmyID]; siegeArmy != nil {
				source := r.gs.Regions[attacker.RegionID]
				if source != nil {
					r.openBattlePlanWithDestination(attacker, source, siegeArmy, ActionMoveArmy, combat.BattleContextLand, wc.pendingDest)
				}
			}
		}
		return action
	}
	if renderTargetRequiresSiegeDecision(r.gs, attacker, target) && !supportingSiege {
		if (activeSiege == nil || activeSiege.AttackerArmyID == attacker.ID) && !isAlliedRegion {
			r.openSiegeDecision(attacker, target)
			return action
		}
	}
	if activeSiege != nil && !supportingSiege && !canEnterSiege {
		r.ShowCombatResult("Bu bölge zaten başka bir ordu tarafından kuşatılıyor.")
		return InputAction{}
	}
	if supportingSiege {
		r.SelectedArmy = ""
		action.Kind = ActionDeclareWarAndMove
		action.ArmyID = wc.pendingArmy
		action.TargetRegion = wc.pendingDest
		return action
	}
	if wc.opensBattlePlan {
		if attacker != nil && target != nil {
			if defender := r.gs.Armies[wc.pendingEnemy]; defender != nil {
				r.openBattlePlan(attacker, target, defender, wc.battleAction, wc.battleContext)
			}
		}
		return action
	}
	r.SelectedArmy = ""
	action.Kind = ActionDeclareWarAndMove
	action.ArmyID = wc.pendingArmy
	action.TargetRegion = wc.pendingDest
	return action
}

func (r *Renderer) drawWarConfirmDialog(screen *ebiten.Image) {
	modal := buildWarConfirmModal()
	gameui.DrawModal(screen, modal, standardModalStyle, nil, nil)
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 20, W: modal.Panel.Rect.W - 48}, r.warConfirm.factionName+" devletine savaş ilanı", color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 50, W: modal.Panel.Rect.W - 48}, "Bu ilanla hangi devletlerin cepheye çekilebileceğini aşağıda görüyorsun.", color.RGBA{220, 220, 220, 255}, gameui.TextSmall, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 72, W: modal.Panel.Rect.W - 48}, "Sol tarafta kendi çağıracağın müttefikleri seçebilirsin.", color.RGBA{186, 170, 132, 255}, gameui.TextSmall, gameui.TextAlignCenter)

	leftRect, rightRect := warConfirmSideRects(modal)
	scroll := r.warConfirm.scroll
	leftViewport := warConfirmCallViewport(leftRect, len(r.warConfirm.preview.Attacker.AutoParticipants))
	rightViewport := warConfirmCallViewport(rightRect, len(r.warConfirm.preview.Defender.AutoParticipants))
	maxScroll := warConfirmSharedMaxScroll(r.warConfirm.preview.Attacker.CallableAllies, r.warConfirm.preview.Defender.CallableAllies, leftViewport, rightViewport)
	if scroll > maxScroll {
		scroll = maxScroll
	}
	drawWarConfirmSide(screen, leftRect, "Senin Cephen", r.warConfirm.preview.Attacker.PrimaryNameTR, "Ek otomatik katılım yok.", "Çağrılabilir müttefik yok.", r.warConfirm.preview.Attacker.AutoParticipants, r.warConfirm.preview.Attacker.CallableAllies, r.warConfirm.selectedAllies, true, scroll)
	drawWarConfirmSide(screen, rightRect, "Karşı Cephe", r.warConfirm.preview.Defender.PrimaryNameTR, "Ek otomatik katılım yok.", "Çağrılabilir müttefik yok.", r.warConfirm.preview.Defender.AutoParticipants, r.warConfirm.preview.Defender.CallableAllies, nil, false, scroll)
	if maxScroll > 0 {
		contentRect := warConfirmContentRect(modal)
		track := gameui.Rect{X: modal.Panel.Rect.X + modal.Panel.Rect.W - 12, Y: contentRect.Y, W: 4, H: contentRect.H}
		visibleRows := warConfirmSharedVisibleRows(leftViewport, rightViewport)
		drawWarConfirmScrollbar(screen, track, visibleRows, maxScroll, scroll)
	}

	acceptBtn, declineBtn := buildWarConfirmButtons()
	drawUIButtonWidget(screen, acceptBtn,
		solidButtonStyle(color.RGBA{160, 40, 40, 230}, color.RGBA{205, 90, 90, 255}, color.RGBA{255, 220, 220, 255}, 10))
	drawUIButtonWidget(screen, declineBtn,
		solidButtonStyle(color.RGBA{50, 50, 50, 230}, color.RGBA{120, 120, 120, 255}, color.RGBA{200, 200, 200, 255}, 10))
}

func (r *Renderer) handleWarConfirmInput() InputAction {
	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	acceptBtn, declineBtn := buildWarConfirmButtons()
	modal := buildWarConfirmModal()
	leftRect, rightRect := warConfirmSideRects(modal)
	leftViewport := warConfirmCallViewport(leftRect, len(r.warConfirm.preview.Attacker.AutoParticipants))
	rightViewport := warConfirmCallViewport(rightRect, len(r.warConfirm.preview.Defender.AutoParticipants))
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		step := 1
		if wheelY > 0 {
			step = -1
		}
		if warConfirmScrollHit(modal, mx, my) {
			maxScroll := warConfirmSharedMaxScroll(r.warConfirm.preview.Attacker.CallableAllies, r.warConfirm.preview.Defender.CallableAllies, leftViewport, rightViewport)
			r.warConfirm.scroll += step
			if r.warConfirm.scroll < 0 {
				r.warConfirm.scroll = 0
			}
			if r.warConfirm.scroll > maxScroll {
				r.warConfirm.scroll = maxScroll
			}
			return InputAction{}
		}
	}

	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if idx := warConfirmEntryIndexAt(leftViewport, r.warConfirm.preview.Attacker.CallableAllies, r.warConfirm.scroll, mx, my); idx >= 0 {
			entry := r.warConfirm.preview.Attacker.CallableAllies[idx]
			r.warConfirm.selectedAllies[entry.FactionID] = !r.warConfirm.selectedAllies[entry.FactionID]
			return InputAction{}
		}
		if acceptBtn.HitTest(mx, my) {
			wc := r.warConfirm
			r.warConfirm = warConfirmState{}
			return r.finalizeWarConfirm(wc)
		}
		if declineBtn.HitTest(mx, my) {
			r.warConfirm = warConfirmState{}
			return InputAction{}
		}
	}
	if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyN) {
		r.warConfirm = warConfirmState{}
	}
	if r.keyJustPressed(ebiten.KeyY) || r.keyJustPressed(ebiten.KeyEnter) {
		wc := r.warConfirm
		r.warConfirm = warConfirmState{}
		return r.finalizeWarConfirm(wc)
	}
	return InputAction{}
}

func (r *Renderer) drawBattlePlanDialog(screen *ebiten.Image) {
	modal := buildBattlePlanModal()
	gameui.DrawModal(screen, modal, standardModalStyle, nil, nil)

	title := combat.BattleContextLabelTR(r.battlePlan.battleContext) + " Planı"
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 24}, title, color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignStart)
	subtitle := r.battlePlan.regionName
	if subtitle == "" {
		subtitle = string(r.battlePlan.pendingDest)
	}
	if r.battlePlan.defenderFaction != "" {
		subtitle += " | " + r.battlePlan.defenderFaction
	}
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 52, W: modal.Panel.Rect.W - 48}, subtitle, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 72, W: modal.Panel.Rect.W - 48}, battlePlanInstructionTR(r.battlePlan.battleContext), color.RGBA{220, 220, 220, 255}, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 90, W: modal.Panel.Rect.W - 48}, trimTextToWidth(r.battlePlan.attackerSummary, FaceSmall, modal.Panel.Rect.W-48), color.RGBA{172, 224, 164, 255}, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 106, W: modal.Panel.Rect.W - 48}, trimTextToWidth(r.battlePlan.defenderSummary, FaceSmall, modal.Panel.Rect.W-48), color.RGBA{168, 196, 226, 255}, gameui.TextSmall, gameui.TextAlignStart)

	cardRects := battlePlanCardRects()
	buttons, cancelBtn := buildBattlePlanButtons()
	for i := range r.battlePlan.previews {
		preview := r.battlePlan.previews[i]
		card := cardRects[i]
		fill := color.RGBA{34, 28, 20, 236}
		border := color.RGBA{108, 86, 54, 255}
		if i == r.battlePlan.focus {
			fill = color.RGBA{54, 40, 24, 244}
			border = color.RGBA{220, 170, 82, 255}
		}
		vector.FillRect(screen, float32(card.X), float32(card.Y), float32(card.W), float32(card.H), fill, false)
		vector.StrokeRect(screen, float32(card.X), float32(card.Y), float32(card.W), float32(card.H), 1.5, border, false)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 14, W: card.W - 28}, preview.StanceLabelTR, color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignStart)
		drawUIWrappedLabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 42, W: card.W - 28}, preview.StanceSummaryTR, color.RGBA{212, 212, 212, 255}, gameui.TextSmall, 17, 2)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 88, W: card.W - 28}, "Zafer Şansı: %"+itoa(preview.WinChance), color.RGBA{178, 228, 150, 255}, gameui.TextMedium, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 114, W: card.W - 28}, "Muhtemel Sonuç: "+preview.LikelyOutcome, color.RGBA{226, 226, 226, 255}, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 138, W: card.W - 28}, "Güç: "+itoa(preview.AttackStrength)+" / "+itoa(preview.DefenseStrength), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 162, W: card.W - 28}, "Sizin HP: "+battlePlanHPText(preview.AttackerHPExpected, preview.AttackerHPDamageMin, preview.AttackerHPDamageMax), color.RGBA{255, 198, 148, 255}, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 184, W: card.W - 28}, "Sizin Birim: "+battlePlanLossText(preview.AttackerLossExpected, preview.AttackerLossMin, preview.AttackerLossMax), color.RGBA{232, 182, 132, 255}, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 208, W: card.W - 28}, "Düşman HP: "+battlePlanHPText(preview.DefenderHPExpected, preview.DefenderHPDamageMin, preview.DefenderHPDamageMax), color.RGBA{168, 220, 168, 255}, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 230, W: card.W - 28}, "Düşman Birim: "+battlePlanLossText(preview.DefenderLossExpected, preview.DefenderLossMin, preview.DefenderLossMax), color.RGBA{140, 206, 140, 255}, gameui.TextSmall, gameui.TextAlignStart)

		btn := buttons[i]
		btn.Label = preview.StanceLabelTR
		btnStyle := solidButtonStyle(color.RGBA{84, 68, 44, 240}, color.RGBA{146, 112, 62, 255}, ColorWhite, 10)
		if i == r.battlePlan.focus {
			btnStyle = solidButtonStyle(color.RGBA{140, 94, 38, 245}, color.RGBA{206, 150, 70, 255}, ColorWhite, 10)
		}
		drawUIButtonWidget(screen, btn, btnStyle)
	}

	cancelBtn.Label = "İptal"
	drawUIButtonWidget(screen, cancelBtn,
		solidButtonStyle(color.RGBA{72, 72, 72, 220}, color.RGBA{118, 118, 118, 255}, ColorWhite, 10))
}

func (r *Renderer) handleBattlePlanInput() InputAction {
	if !r.battlePlan.show {
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyLeft) || r.keyJustPressed(ebiten.KeyUp) {
		r.battlePlan.focus = (r.battlePlan.focus + len(battlePlanStances) - 1) % len(battlePlanStances)
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyRight) || r.keyJustPressed(ebiten.KeyDown) || r.keyJustPressed(ebiten.KeyTab) {
		r.battlePlan.focus = (r.battlePlan.focus + 1) % len(battlePlanStances)
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.Key1) {
		r.battlePlan.focus = 0
	}
	if r.keyJustPressed(ebiten.Key2) {
		r.battlePlan.focus = 1
	}
	if r.keyJustPressed(ebiten.Key3) {
		r.battlePlan.focus = 2
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.battlePlan = battlePlanState{}
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace) {
		bp := r.battlePlan
		r.battlePlan = battlePlanState{}
		r.SelectedArmy = ""
		return InputAction{
			Kind:         bp.actionKind,
			ArmyID:       bp.pendingArmy,
			TargetRegion: bp.pendingDest,
			BattleStance: battlePlanStances[bp.focus],
		}
	}

	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	buttons, cancelBtn := buildBattlePlanButtons()
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if cancelBtn.HitTest(mx, my) {
			r.battlePlan = battlePlanState{}
			return InputAction{}
		}
		for i, btn := range buttons {
			if !btn.HitTest(mx, my) {
				continue
			}
			bp := r.battlePlan
			r.battlePlan = battlePlanState{}
			r.SelectedArmy = ""
			return InputAction{
				Kind:         bp.actionKind,
				ArmyID:       bp.pendingArmy,
				TargetRegion: bp.pendingDest,
				BattleStance: battlePlanStances[i],
			}
		}
	}
	return InputAction{}
}

func (r *Renderer) playerDiplomacyOfferIndex() (int, bool) {
	if r.gs == nil || len(r.gs.DiplomaticOffers) == 0 || r.diplomacyOfferHistoryBrowse != "" {
		return 0, false
	}
	if offerIdx, ok := diplomacy.BestOfferIndex(r.gs, r.gs.PlayerFactionID); ok {
		return offerIdx, true
	}
	return 0, false
}

func diplomacyOfferActionLabelTR(action string) string {
	switch action {
	case "join_war_call":
		return "savaşa katılım"
	case "propose_peace":
		return "barış"
	case "propose_alliance":
		return "ittifak"
	case "propose_trade":
		return "ticaret"
	default:
		return "teklif"
	}
}

func diplomacyOfferTitleTR(offer state.DiplomaticOffer) string {
	if offer.Action == string(diplomacy.ActionJoinWarCall) {
		return "Savaşa Katılım Çağrısı"
	}
	return "Anlaşma Teklifi"
}

func diplomacyOfferMessageTR(gs *state.GameState, offer state.DiplomaticOffer) string {
	fromName := factionDisplayName(gs, string(offer.FromFactionID))
	if fromName == "" {
		fromName = string(offer.FromFactionID)
	}
	if offer.Action != string(diplomacy.ActionJoinWarCall) {
		return fromName + " devleti size " + diplomacyOfferActionLabelTR(offer.Action) + " teklif etti."
	}

	declarerName := factionDisplayName(gs, string(offer.WarDeclarerFactionID))
	if declarerName == "" {
		declarerName = string(offer.WarDeclarerFactionID)
	}
	enemyName := factionDisplayName(gs, string(offer.WarEnemyFactionID))
	if enemyName == "" {
		enemyName = string(offer.WarEnemyFactionID)
	}
	if offer.WarDeclarerFactionID == offer.FromFactionID {
		return declarerName + " devleti " + enemyName + " devletine savaş ilan etti. Müttefikinizin tarafında yer alacak mısınız?"
	}
	return declarerName + " devleti " + fromName + " devletine savaş ilan etti. Müttefikiniz sizi kendi safında savaşa çağırıyor."
}

func diplomacyOfferReasonTextTR(offer state.DiplomaticOffer) string {
	if offer.PriorityReason != "" {
		return "Sebep: " + offer.PriorityReason
	}
	if offer.Action == string(diplomacy.ActionJoinWarCall) {
		return "Sebep: aktif ittifak nedeniyle savaşa çağrıldınız"
	}
	return "Sebep: standart diplomasi akışı"
}

func diplomacyOfferOutcomeLabelTR(entry state.DiplomaticOfferHistoryEntry) string {
	if !entry.Accepted {
		return "Reddedildi"
	}
	if entry.Applied {
		return "Kabul edildi"
	}
	return "Kabul edildi ama uygulanamadı"
}

func diplomacyOfferTurnLabelTR(entry state.DiplomaticOfferHistoryEntry) string {
	if entry.CreatedTurn > 0 && entry.ResolvedTurn > 0 && entry.CreatedTurn != entry.ResolvedTurn {
		return fmt.Sprintf("Tur %d->%d", entry.CreatedTurn, entry.ResolvedTurn)
	}
	if entry.ResolvedTurn > 0 {
		return fmt.Sprintf("Tur %d", entry.ResolvedTurn)
	}
	if entry.CreatedTurn > 0 {
		return fmt.Sprintf("Tur %d", entry.CreatedTurn)
	}
	return "Tur ?"
}

func drawDiplomacyOfferHistoryPanelRect(screen *ebiten.Image, gs *state.GameState, panelRect gameui.Rect, maxEntries int, dirFilter diplomacyHistoryDirectionFilter, actionFilter ActionKind) {
	if gs == nil || panelRect.W <= 0 || panelRect.H <= 0 {
		return
	}
	if maxEntries <= 0 {
		maxEntries = 3
	}
	drawUIPanelFrame(screen, panelRect, color.RGBA{18, 14, 10, 228}, color.RGBA{88, 72, 40, 180}, 1, 3)

	drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: panelRect.Y + 10, W: panelRect.W - 28}, "Geçmiş", color.RGBA{255, 220, 100, 255}, gameui.TextMedium, gameui.TextAlignStart)
	drawUIMutedText(screen, panelRect.X+14, panelRect.Y+30, "Sana ilgili son çözümler")
	total, accepted, rejected, applied := diplomacyOfferHistorySummary(gs, dirFilter, actionFilter)
	summaryLine := fmt.Sprintf("Toplam %d | Kabul %d | Ret %d | Uygulanan %d", total, accepted, rejected, applied)
	drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: panelRect.Y + 42, W: panelRect.W - 28}, summaryLine, color.RGBA{210, 198, 172, 235}, gameui.TextSmall, gameui.TextAlignStart)
	drawUIMutedText(screen, panelRect.X+14, panelRect.Y+58, diplomacyOfferHistoryFilterLabelTR(dirFilter, actionFilter))

	buttons := buildDiplomacyHistoryFilterButtons(panelRect, dirFilter, actionFilter)
	for i := range buttons {
		btn := buttons[i]
		if btn.IsAction {
			style := diplomacyHistoryFilterButtonStyle(btn.Action == actionFilter, diplomacyHistoryActionColor(btn.Action))
			drawUIButtonWidget(screen, btn.Button, style)
			continue
		}
		style := diplomacyHistoryFilterButtonStyle(btn.Direction == dirFilter, color.RGBA{104, 82, 42, 228})
		drawUIButtonWidget(screen, btn.Button, style)
	}

	drawn := 0
	for i := len(gs.DiplomaticOfferHistory) - 1; i >= 0 && drawn < maxEntries; i-- {
		entry := gs.DiplomaticOfferHistory[i]
		if !diplomacyOfferHistoryMatches(entry, gs.PlayerFactionID, dirFilter, actionFilter) {
			continue
		}
		fromName := factionDisplayName(gs, string(entry.FromFactionID))
		if fromName == "" {
			fromName = string(entry.FromFactionID)
		}
		toName := factionDisplayName(gs, string(entry.ToFactionID))
		if toName == "" {
			toName = string(entry.ToFactionID)
		}

		itemRect := diplomacyOfferHistoryCardRect(panelRect, drawn)
		bg := color.RGBA{26, 20, 14, 220}
		border := color.RGBA{82, 66, 36, 150}
		textCol := color.RGBA{235, 228, 214, 255}
		switch {
		case entry.Accepted && entry.Applied:
			bg = color.RGBA{22, 44, 24, 225}
			border = color.RGBA{92, 148, 92, 210}
			textCol = color.RGBA{224, 245, 224, 255}
		case entry.Accepted && !entry.Applied:
			bg = color.RGBA{54, 42, 18, 225}
			border = color.RGBA{178, 136, 58, 210}
			textCol = color.RGBA{248, 234, 198, 255}
		case !entry.Accepted:
			bg = color.RGBA{52, 20, 20, 225}
			border = color.RGBA{168, 78, 78, 210}
			textCol = color.RGBA{246, 214, 214, 255}
		}
		drawUICardRect(screen, itemRect, bg, border, 1)
		actionKind := diplomacyHistoryActionForEntry(entry)
		iconRect := gameui.Rect{X: itemRect.X + 9, Y: itemRect.Y + 10, W: 18, H: 18}
		drawUICardRect(screen, iconRect, diplomacyHistoryActionColor(actionKind), color.RGBA{20, 16, 12, 160}, 1)
		if icon := diplomacyHistoryActionIcon(actionKind); icon != gameui.IconNone {
			gameui.DrawIcon(screen, icon, iconRect.X+3, iconRect.Y+3, 12, ColorWhite)
		}

		actionLabel := diplomacyHistoryActionLabelTR(actionKind)
		if actionKind == ActionNone {
			actionLabel = diplomacyOfferActionLabelTR(entry.Action)
		}
		line1 := diplomacyOfferHistoryDirectionTR(entry, gs.PlayerFactionID) + " | " + fromName + " -> " + toName + " | " + actionLabel + " | " + diplomacyOfferOutcomeLabelTR(entry)
		contentW := itemRect.W - 116
		drawUILabel(screen, gameui.Rect{X: itemRect.X + 34, Y: itemRect.Y + 4, W: contentW}, trimTextToWidth(line1, FaceSmall, contentW), textCol, gameui.TextSmall, gameui.TextAlignStart)

		line2 := diplomacyOfferTurnLabelTR(entry)
		if entry.Priority != 0 {
			line2 += " | Öncelik " + itoa(entry.Priority)
		}
		if entry.ResultMessage != "" {
			line2 += " | " + entry.ResultMessage
		}
		drawUILabel(screen, gameui.Rect{X: itemRect.X + 34, Y: itemRect.Y + 20, W: contentW}, trimTextToWidth(line2, FaceSmall, contentW), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		statusText := diplomacyHistoryOutcomeBadgeTR(entry)
		statusRect := gameui.Rect{X: itemRect.X + itemRect.W - 70, Y: itemRect.Y + 4, W: 60, H: 16}
		statusBg := color.RGBA{74, 64, 44, 220}
		statusBorder := color.RGBA{112, 96, 64, 220}
		switch statusText {
		case "UYG":
			statusBg = color.RGBA{34, 88, 42, 220}
			statusBorder = color.RGBA{84, 150, 88, 220}
		case "KAB":
			statusBg = color.RGBA{120, 92, 34, 220}
			statusBorder = color.RGBA{182, 140, 60, 220}
		case "RET":
			statusBg = color.RGBA{112, 42, 42, 220}
			statusBorder = color.RGBA{182, 86, 86, 220}
		}
		drawUICardRect(screen, statusRect, statusBg, statusBorder, 1)
		drawUILabel(screen, statusRect, statusText, ColorWhite, gameui.TextSmall, gameui.TextAlignCenter)
		drawn++
	}

	if drawn == 0 {
		drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: panelRect.Y + 126, W: panelRect.W - 28}, "Bu filtreyle eşleşen çözüm yok.", ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
	}
}

func diplomacyFactionOwnedRegionCount(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" {
		return 0
	}
	count := 0
	for _, region := range gs.Regions {
		if region == nil || region.OwnerID != string(fid) {
			continue
		}
		count++
	}
	return count
}

func diplomacyFactionLandRegionCount(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" {
		return 0
	}
	count := 0
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		count++
	}
	return count
}

type diplomacyOfferRelationGroup int

const (
	diplomacyOfferRelationAllied diplomacyOfferRelationGroup = iota
	diplomacyOfferRelationTrade
	diplomacyOfferRelationWar
)

func diplomacyOfferRelationGroupMatches(gs *state.GameState, subject, other faction.FactionID, group diplomacyOfferRelationGroup) bool {
	if gs == nil || subject == "" || other == "" || subject == other {
		return false
	}
	switch group {
	case diplomacyOfferRelationAllied:
		if rel := diplomacy.Relation(gs, subject, other); rel != nil && rel.Stance == faction.StanceAllied {
			return true
		}
	case diplomacyOfferRelationTrade:
		return diplomacy.HasTradeRouteBetween(gs, subject, other)
	case diplomacyOfferRelationWar:
		return diplomacy.IsWar(gs, subject, other)
	}
	return false
}

func diplomacyOfferRelationLine(gs *state.GameState, subject faction.FactionID, group diplomacyOfferRelationGroup) string {
	if gs == nil || subject == "" {
		return "Yok"
	}
	factions := sortedFactions(gs)
	if gs.PlayerFactionID != "" && gs.PlayerFactionID != subject {
		factions = append(factions, gs.PlayerFactionID)
	}

	var names [3]string
	count := 0
	for _, otherID := range factions {
		other := gs.Factions[otherID]
		if other == nil || other.IsEliminated || !diplomacyOfferRelationGroupMatches(gs, subject, otherID, group) {
			continue
		}
		name := factionDisplayName(gs, string(otherID))
		if name == "" {
			name = string(otherID)
		}
		if count < len(names) {
			names[count] = name
		}
		count++
		if count == len(names) {
			break
		}
	}
	if count == 0 {
		return "Yok"
	}
	var b strings.Builder
	for i := 0; i < count && i < len(names); i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(names[i])
	}
	if count > len(names) {
		b.WriteString(", ...")
	}
	return b.String()
}

func drawDiplomacyOfferSummaryPanel(screen *ebiten.Image, gs *state.GameState, panelRect gameui.Rect, offer state.DiplomaticOffer) {
	if gs == nil || panelRect.W <= 0 || panelRect.H <= 0 {
		return
	}
	drawUIPanelFrame(screen, panelRect, color.RGBA{18, 14, 10, 228}, color.RGBA{88, 72, 40, 180}, 1, 3)

	subject := offer.FromFactionID
	subjectName := factionDisplayName(gs, string(subject))
	if subjectName == "" {
		subjectName = string(subject)
	}
	drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: panelRect.Y + 10, W: panelRect.W - 28}, "Teklif Eden Devlet", color.RGBA{255, 220, 100, 255}, gameui.TextMedium, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: panelRect.Y + 32, W: panelRect.W - 28}, trimTextToWidth(subjectName, FaceSmall, panelRect.W-28), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
	drawUIMutedText(screen, panelRect.X+14, panelRect.Y+48, "Karar vermek için anlık durum özeti")

	rel := diplomacy.Relation(gs, gs.PlayerFactionID, subject)
	relationColor, relationLabel := diplomacyStatusDisplay(gs, gs.PlayerFactionID, subject, rel)
	y := panelRect.Y + 72
	drawUIKeyValueRow(screen, panelRect.X+14, y, panelRect.W-28, "Sizinle durum", relationLabel, ColorGray, relationColor)
	y += 22

	governanceLabel := "Bağımsız"
	governanceColor := ColorGray
	if overlord := diplomacy.DirectOverlord(gs, subject); overlord != "" {
		governanceLabel = factionDisplayName(gs, string(overlord))
		if governanceLabel == "" {
			governanceLabel = string(overlord)
		}
		governanceLabel += " himayesinde"
		governanceColor = ColorGold
	}
	drawUIKeyValueRow(screen, panelRect.X+14, y, panelRect.W-28, "Yönetim", governanceLabel, ColorGray, governanceColor)
	y += 22

	drawUIKeyValueRow(screen, panelRect.X+14, y, panelRect.W-28, "Askeri güç", itoa(diplomacy.MilitaryPower(gs, subject)), ColorGray, ColorWhite)
	y += 22
	drawUIKeyValueRow(screen, panelRect.X+14, y, panelRect.W-28, "Kara bölge", itoa(diplomacyFactionLandRegionCount(gs, subject)), ColorGray, ColorWhite)
	y += 22
	drawUIKeyValueRow(screen, panelRect.X+14, y, panelRect.W-28, "Toplam bölge", itoa(diplomacyFactionOwnedRegionCount(gs, subject)), ColorGray, ColorWhite)
	y += 22
	drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: y + 4, W: panelRect.W - 28}, "İttifaklar", color.RGBA{90, 190, 110, 255}, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: y + 20, W: panelRect.W - 28}, trimTextToWidth(diplomacyOfferRelationLine(gs, subject, diplomacyOfferRelationAllied), FaceSmall, panelRect.W-28), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
	y += 42

	drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: y + 4, W: panelRect.W - 28}, "Ticaret", color.RGBA{205, 168, 78, 255}, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: y + 20, W: panelRect.W - 28}, trimTextToWidth(diplomacyOfferRelationLine(gs, subject, diplomacyOfferRelationTrade), FaceSmall, panelRect.W-28), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
	y += 42

	drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: y + 4, W: panelRect.W - 28}, "Savaş", color.RGBA{214, 92, 92, 255}, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: y + 20, W: panelRect.W - 28}, trimTextToWidth(diplomacyOfferRelationLine(gs, subject, diplomacyOfferRelationWar), FaceSmall, panelRect.W-28), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
}

func (r *Renderer) drawDiplomacyOfferSummaryPanel(screen *ebiten.Image, modal gameui.Modal, offer state.DiplomaticOffer) {
	if r == nil || r.gs == nil {
		return
	}
	panelRect := gameui.Rect{
		X: modal.Panel.Rect.X + 458,
		Y: modal.Panel.Rect.Y + 16,
		W: 286,
		H: 304,
	}
	drawDiplomacyOfferSummaryPanel(screen, r.gs, panelRect, offer)
}

func (r *Renderer) drawDiplomacyOfferDialog(screen *ebiten.Image, offerIdx int) {
	offer := r.gs.DiplomaticOffers[offerIdx]
	actionLabel := diplomacyOfferActionLabelTR(offer.Action)

	modal := buildDiplomacyOfferModal()
	gameui.DrawModal(screen, modal, standardModalStyle, nil, nil)

	leftRect := gameui.Rect{X: modal.Panel.Rect.X + 16, Y: modal.Panel.Rect.Y + 16, W: 430, H: 188}
	drawUIPanelFrame(screen, leftRect, color.RGBA{18, 14, 10, 228}, color.RGBA{88, 72, 40, 180}, 1, 3)

	title := diplomacyOfferTitleTR(offer)
	drawUILabel(screen, gameui.Rect{X: leftRect.X + 14, Y: leftRect.Y + 10, W: leftRect.W - 28}, title, color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignStart)
	message := diplomacyOfferMessageTR(r.gs, offer)
	drawUIWrappedLabel(screen, gameui.Rect{X: leftRect.X + 14, Y: leftRect.Y + 38, W: leftRect.W - 28}, message, color.RGBA{220, 220, 220, 255}, gameui.TextMedium, 20, 2)
	priorityLine := fmt.Sprintf("Tür: %s | Öncelik: %d", actionLabel, offer.Priority)
	drawUILabel(screen, gameui.Rect{X: leftRect.X + 14, Y: leftRect.Y + 82, W: leftRect.W - 28}, priorityLine, color.RGBA{255, 205, 120, 255}, gameui.TextSmall, gameui.TextAlignStart)
	reasonText := diplomacyOfferReasonTextTR(offer)
	drawUIWrappedLabel(screen, gameui.Rect{X: leftRect.X + 14, Y: leftRect.Y + 104, W: leftRect.W - 28}, reasonText, color.RGBA{210, 210, 210, 255}, gameui.TextSmall, 18, 2)
	drawUILabel(screen, gameui.Rect{X: leftRect.X + 14, Y: leftRect.Y + 158, W: leftRect.W - 28}, "Kabul etmek için Enter/Y, reddetmek için Esc/N kullanabilirsiniz.", ColorGray, gameui.TextSmall, gameui.TextAlignStart)

	r.drawDiplomacyOfferSummaryPanel(screen, modal, offer)

	acceptBtn, rejectBtn := buildDiplomacyOfferButtons()
	drawUIButtonWidget(screen, acceptBtn,
		solidButtonStyle(color.RGBA{70, 140, 70, 240}, color.RGBA{120, 180, 120, 255}, ColorWhite, 10))
	drawUIButtonWidget(screen, rejectBtn,
		solidButtonStyle(color.RGBA{140, 70, 70, 240}, color.RGBA{190, 110, 110, 255}, ColorWhite, 10))
}

func (r *Renderer) handleDiplomacyOfferInput(offerIdx int) InputAction {
	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	input := gameui.InputState{
		MouseX:           mx,
		MouseY:           my,
		LeftJustPressed:  r.mouseJustPressed(ebiten.MouseButtonLeft),
		LeftJustReleased: false,
	}
	return r.handleDiplomacyOfferInputState(offerIdx, input)
}

func (r *Renderer) handleDiplomacyOfferInputState(offerIdx int, input gameui.InputState) InputAction {
	acceptBtn, rejectBtn := buildDiplomacyOfferButtons()
	if input.LeftJustPressed {
		if acceptBtn.HitTest(input.MouseX, input.MouseY) {
			return InputAction{Kind: ActionRespondDiplomacyOffer, OfferIndex: offerIdx, OfferAccepted: true}
		}
		if rejectBtn.HitTest(input.MouseX, input.MouseY) {
			return InputAction{Kind: ActionRespondDiplomacyOffer, OfferIndex: offerIdx, OfferAccepted: false}
		}
	}
	if r.keyJustPressed(ebiten.KeyY) || r.keyJustPressed(ebiten.KeyEnter) {
		return InputAction{Kind: ActionRespondDiplomacyOffer, OfferIndex: offerIdx, OfferAccepted: true}
	}
	if r.keyJustPressed(ebiten.KeyN) || r.keyJustPressed(ebiten.KeyEscape) {
		return InputAction{Kind: ActionRespondDiplomacyOffer, OfferIndex: offerIdx, OfferAccepted: false}
	}
	return InputAction{}
}

func (r *Renderer) handleHistoricalEventInput() InputAction {
	if len(r.historicalEventChoices) == 0 {
		if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyEnter) ||
			r.keyJustPressed(ebiten.KeySpace) || r.mouseJustPressed(ebiten.MouseButtonLeft) {
			r.showHistoricalEvent = false
			r.historicalEventPrompt = ""
			r.historicalEventChoices = r.historicalEventChoices[:0]
		}
		return InputAction{}
	}

	buttons := buildHistoricalEventChoiceButtons(len(r.historicalEventChoices))
	if r.keyJustPressed(ebiten.KeyLeft) || r.keyJustPressed(ebiten.KeyUp) {
		r.historicalEventFocus = (r.historicalEventFocus + len(r.historicalEventChoices) - 1) % len(r.historicalEventChoices)
	}
	if r.keyJustPressed(ebiten.KeyRight) || r.keyJustPressed(ebiten.KeyDown) || r.keyJustPressed(ebiten.KeyTab) {
		r.historicalEventFocus = (r.historicalEventFocus + 1) % len(r.historicalEventChoices)
	}
	if r.keyJustPressed(ebiten.Key1) {
		return InputAction{Kind: ActionChooseHistoricalEvent, ChoiceIndex: 0}
	}
	if len(r.historicalEventChoices) > 1 && r.keyJustPressed(ebiten.Key2) {
		return InputAction{Kind: ActionChooseHistoricalEvent, ChoiceIndex: 1}
	}
	if r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace) {
		return InputAction{Kind: ActionChooseHistoricalEvent, ChoiceIndex: r.historicalEventFocus}
	}

	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		mxi, myi := ebiten.CursorPosition()
		mx, my := float64(mxi), float64(myi)
		for i, btn := range buttons {
			if btn.HitTest(mx, my) {
				r.historicalEventFocus = i
				return InputAction{Kind: ActionChooseHistoricalEvent, ChoiceIndex: i}
			}
		}
	}
	return InputAction{}
}

func (r *Renderer) handleEventCodexInput() InputAction {
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.CloseEventCodex()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyLeft) {
		r.cycleEventCodexFilter(-1)
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyRight) || r.keyJustPressed(ebiten.KeyTab) {
		r.cycleEventCodexFilter(1)
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyUp) {
		r.cycleEventCodexFocus(-1)
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyDown) {
		r.cycleEventCodexFocus(1)
		return InputAction{}
	}
	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 && eventCodexListHit(mx, my) {
		if wheelY > 0 {
			r.scrollEventCodex(-1)
		} else if wheelY < 0 {
			r.scrollEventCodex(1)
		}
		return InputAction{}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if eventCodexCloseHit(mx, my) || !eventCodexPopupHit(mx, my) {
			r.CloseEventCodex()
			return InputAction{}
		}
		buttons := buildEventCodexFilterButtons()
		for i, btn := range buttons {
			if btn.HitTest(mx, my) {
				r.eventCodexFilter = EventCodexFilter(i)
				r.eventCodexFocus = 0
				r.eventCodexScroll = 0
				return InputAction{}
			}
		}
		if idx := eventCodexEntryHit(mx, my, len(r.currentEventCodexEntries()), r.eventCodexScroll); idx >= 0 {
			r.eventCodexFocus = idx
			r.ensureEventCodexFocusVisible()
			return InputAction{}
		}
	}
	return InputAction{}
}

func (r *Renderer) ShowConfirmDialog(title, message, acceptLabel, declineLabel string, action InputAction, declineHook func()) {
	r.confirmDialog = confirmDialogState{
		show:          true,
		title:         title,
		message:       message,
		messageLines:  wrapTextLines(message, FaceSmall, float64(confirmDialogW)-40),
		acceptLabel:   acceptLabel,
		declineLabel:  declineLabel,
		pendingAction: action,
		declineHook:   declineHook,
	}
}

func (r *Renderer) ShowChoiceDialog(title, message, acceptLabel, declineLabel string, acceptAction, declineAction InputAction) {
	r.confirmDialog = confirmDialogState{
		show:          true,
		title:         title,
		message:       message,
		messageLines:  wrapTextLines(message, FaceSmall, float64(confirmDialogW)-40),
		acceptLabel:   acceptLabel,
		declineLabel:  declineLabel,
		pendingAction: acceptAction,
		declineAction: declineAction,
		declineActs:   true,
	}
}

func (r *Renderer) QueueChoiceDialogAfterBattleReport(title, message, acceptLabel, declineLabel string, acceptAction, declineAction InputAction) {
	r.queuedConfirmDialog = confirmDialogState{
		show:          true,
		title:         title,
		message:       message,
		messageLines:  wrapTextLines(message, FaceSmall, float64(confirmDialogW)-40),
		acceptLabel:   acceptLabel,
		declineLabel:  declineLabel,
		pendingAction: acceptAction,
		declineAction: declineAction,
		declineActs:   true,
	}
}

func (r *Renderer) showEditExitConfirm() {
	r.confirmDialog = confirmDialogState{
		show:          true,
		title:         "Kaydedilmemis Degisiklik",
		message:       "Edit mode degisiklikleri kaydedilmedi. Cikmadan once ne yapmak istiyorsunuz?",
		messageLines:  wrapTextLines("Edit mode degisiklikleri kaydedilmedi. Cikmadan once ne yapmak istiyorsunuz?", FaceSmall, float64(confirmDialogW)-40),
		acceptLabel:   "Kaydet",
		thirdLabel:    "Kaydetmeden Cik",
		declineLabel:  "Iptal",
		pendingAction: InputAction{Kind: ActionSaveScenarioAndGoMainMenu},
		thirdAction:   InputAction{Kind: ActionGoMainMenu},
	}
}

func (r *Renderer) drawConfirmDialog(screen *ebiten.Image) {
	modal := buildConfirmDialogModal()
	gameui.DrawModal(screen, modal, standardModalStyle, nil, nil)

	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 20, Y: modal.Panel.Rect.Y + 28}, r.confirmDialog.title, color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignStart)
	drawUIWrappedLabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 20, Y: modal.Panel.Rect.Y + 58, W: modal.Panel.Rect.W - 40}, r.confirmDialog.message, color.RGBA{220, 220, 220, 255}, gameui.TextSmall, 17, 3)
	r.drawConfirmDialogButtons(screen)
}

func (r *Renderer) drawConfirmDialogButtons(screen *ebiten.Image) {
	acceptBtn, thirdBtn, declineBtn, hasThird := buildConfirmDialogButtons(r.confirmDialog)
	acceptBtn = decorateConfirmDialogButton(acceptBtn, r.confirmDialog.acceptLabel, "accept")
	drawUIButtonWidget(screen, acceptBtn,
		solidButtonStyle(color.RGBA{70, 140, 70, 240}, color.RGBA{120, 180, 120, 255}, ColorWhite, 10))
	if hasThird {
		thirdBtn = decorateConfirmDialogButton(thirdBtn, r.confirmDialog.thirdLabel, "third")
		declineBtn = decorateConfirmDialogButton(declineBtn, r.confirmDialog.declineLabel, "decline")
		drawUIButtonWidget(screen, thirdBtn,
			solidButtonStyle(color.RGBA{145, 95, 45, 235}, color.RGBA{190, 135, 75, 255}, ColorWhite, 10))
		drawUIButtonWidget(screen, declineBtn,
			solidButtonStyle(color.RGBA{70, 70, 70, 220}, color.RGBA{120, 120, 120, 255}, ColorWhite, 10))
		return
	}
	declineBtn = decorateConfirmDialogButton(declineBtn, r.confirmDialog.declineLabel, "decline")
	if r.confirmDialog.declineActs {
		drawUIButtonWidget(screen, declineBtn,
			solidButtonStyle(color.RGBA{145, 95, 45, 235}, color.RGBA{190, 135, 75, 255}, ColorWhite, 10))
		return
	}
	drawUIButtonWidget(screen, declineBtn,
		solidButtonStyle(color.RGBA{70, 70, 70, 220}, color.RGBA{120, 120, 120, 255}, ColorWhite, 10))
}

func decorateConfirmDialogButton(btn gameui.Button, label string, role string) gameui.Button {
	btn.Label = label
	switch role {
	case "accept":
		if strings.Contains(label, "İlhak") {
			return btn.WithIcon(gameui.IconSword)
		}
		if label == "Kaydet" {
			return btn.WithIcon(gameui.IconSave)
		}
		return btn.WithIcon(gameui.IconCheck)
	case "third":
		if label == "Kaydetmeden Cik" {
			return btn.WithIcon(gameui.IconExit)
		}
		return btn.WithIcon(gameui.IconSave)
	default:
		if label == "Vassal Yap" {
			return btn.WithIcon(gameui.IconCheck)
		}
		return btn.WithIcon(gameui.IconClose)
	}
}

func confirmDialogThreeButtonXs(cx float32) (float32, float32, float32) {
	gap := float32(14)
	totalW := confirmDialogBtnW*3 + gap*2
	saveX := cx + (confirmDialogW-totalW)/2
	discardX := saveX + confirmDialogBtnW + gap
	cancelX := discardX + confirmDialogBtnW + gap
	return saveX, discardX, cancelX
}

func (r *Renderer) handleConfirmDialogInput() InputAction {
	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	acceptBtn, thirdBtn, declineBtn, hasThird := buildConfirmDialogButtons(r.confirmDialog)

	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if acceptBtn.HitTest(mx, my) {
			action := r.confirmDialog.pendingAction
			r.confirmDialog = confirmDialogState{}
			return action
		}
		if hasThird && thirdBtn.HitTest(mx, my) {
			action := r.confirmDialog.thirdAction
			r.confirmDialog = confirmDialogState{}
			return action
		}
		if declineBtn.HitTest(mx, my) {
			if r.confirmDialog.declineActs {
				action := r.confirmDialog.declineAction
				r.confirmDialog = confirmDialogState{}
				return action
			}
			if r.confirmDialog.declineHook != nil {
				r.confirmDialog.declineHook()
			}
			r.confirmDialog = confirmDialogState{}
			return InputAction{}
		}
	}
	if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyN) {
		if r.confirmDialog.declineActs {
			return InputAction{}
		}
		if r.confirmDialog.declineHook != nil {
			r.confirmDialog.declineHook()
		}
		r.confirmDialog = confirmDialogState{}
	}
	if r.keyJustPressed(ebiten.KeyY) || r.keyJustPressed(ebiten.KeyEnter) {
		action := r.confirmDialog.pendingAction
		r.confirmDialog = confirmDialogState{}
		return action
	}
	return InputAction{}
}

// --- Alt çizim yardımcıları ---
