package render

import (
	"fmt"
	"image/color"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

// CombatSummaryEntry, tur çözümlemesindeki tek bir çatışmanın kayıp kaydını
// renderer'a taşır. Kalıcı savaş ledger'ından ayrı, tek gösterimlik veridir.
type CombatSummaryEntry struct {
	AttackerFactionID faction.FactionID
	DefenderFactionID faction.FactionID
	AttackerLost      int
	DefenderLost      int
	AttackerNaval     bool
	DefenderNaval     bool
	AttackerDestroyed bool
	DefenderDestroyed bool
	Turn              int
}

type CombatSummaryReport struct {
	Turn    int
	Entries []CombatSummaryEntry
}

type combatSummaryState struct {
	show   bool
	data   CombatSummaryReport
	scroll int
}

type combatSummaryFactionTotal struct {
	FactionID      faction.FactionID
	Battles        int
	ArmyLost       int
	FleetLost      int
	ArmyDestroyed  int
	FleetDestroyed int
}

type combatSummaryLayout struct {
	panelRect  gameui.Rect
	titleRect  gameui.Rect
	introRect  gameui.Rect
	listRect   gameui.Rect
	footerRect gameui.Rect
}

func (r *Renderer) ShowCombatSummary(report CombatSummaryReport) {
	if r == nil || len(report.Entries) == 0 {
		return
	}
	report.Entries = filterPlayerRelatedCombatSummaryEntries(r.gs, report.Entries)
	if len(report.Entries) == 0 {
		return
	}
	if report.Turn <= 0 {
		for _, entry := range report.Entries {
			if entry.Turn > 0 {
				report.Turn = entry.Turn
				break
			}
		}
	}
	summary := combatSummaryState{show: true, data: report}
	if r.combatSummary.show || r.combatSummaryBlockedByEventWindow() {
		r.queuedCombatSummary = summary
		return
	}
	r.combatSummary = summary
	r.combatLogTimer = 0
}

func (r *Renderer) HideCombatSummary() {
	if r == nil {
		return
	}
	r.combatSummary = combatSummaryState{}
	r.promoteQueuedCombatSummary()
}

func filterPlayerRelatedCombatSummaryEntries(gs *state.GameState, entries []CombatSummaryEntry) []CombatSummaryEntry {
	if gs == nil || gs.PlayerFactionID == "" || len(entries) == 0 {
		return nil
	}
	filtered := make([]CombatSummaryEntry, 0, len(entries))
	for _, entry := range entries {
		if combatSummaryFactionRelatedToPlayer(gs, entry.AttackerFactionID) ||
			combatSummaryFactionRelatedToPlayer(gs, entry.DefenderFactionID) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func combatSummaryFactionRelatedToPlayer(gs *state.GameState, fid faction.FactionID) bool {
	if gs == nil || fid == "" || gs.PlayerFactionID == "" {
		return false
	}
	if fid == gs.PlayerFactionID || diplomacy.SameRealm(gs, gs.PlayerFactionID, fid) {
		return true
	}
	rel := diplomacy.Relation(gs, gs.PlayerFactionID, fid)
	return rel != nil && (rel.Stance == faction.StanceAllied || rel.Stance == faction.StanceWar)
}

func (r *Renderer) combatSummaryBlockedByEventWindow() bool {
	if r == nil {
		return false
	}
	if r.showHistoricalEvent || r.showEventCodex || r.eventDetail != "" {
		return true
	}
	_, hasOffer := r.playerDiplomacyOfferIndex()
	return hasOffer
}

func (r *Renderer) promoteQueuedCombatSummary() {
	if r == nil || !r.queuedCombatSummary.show || r.combatSummary.show || r.combatSummaryBlockedByEventWindow() {
		return
	}
	r.combatSummary = r.queuedCombatSummary
	r.queuedCombatSummary = combatSummaryState{}
	r.combatLogTimer = 0
}

func buildCombatSummaryModal() gameui.Modal {
	panelW := ScreenWidth - 40
	if panelW <= 0 {
		panelW = 920
	}
	if panelW > 1040 {
		panelW = 1040
	}
	if panelW < 640 {
		panelW = 640
	}
	panelH := ScreenHeight - 40
	if panelH <= 0 {
		panelH = 580
	}
	if panelH > 680 {
		panelH = 680
	}
	if panelH < 420 {
		panelH = 420
	}
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, panelW, panelH, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	return gameui.NewModal(ScreenWidth, ScreenHeight, gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H))
}

func buildCombatSummaryLayout() combatSummaryLayout {
	modal := buildCombatSummaryModal()
	box := gameui.BoxFromRect(modal.Panel.Rect).Inset(22)
	titleRect, box := box.CutTop(38, 12)
	introRect, box := box.CutTop(28, 12)
	listRect, footerBox := box.CutTop(box.Rect.H-52, 14)
	return combatSummaryLayout{
		panelRect:  modal.Panel.Rect,
		titleRect:  titleRect,
		introRect:  introRect,
		listRect:   listRect,
		footerRect: footerBox.Rect,
	}
}

func buildCombatSummaryButton() gameui.Button {
	const (
		buttonW = 184.0
		buttonH = 34.0
	)
	modal := buildCombatSummaryModal()
	x := modal.Panel.Rect.X + (modal.Panel.Rect.W-buttonW)/2
	y := modal.Panel.Rect.Y + modal.Panel.Rect.H - buttonH - 18
	return gameui.NewButton(x, y, buttonW, buttonH, "Tamam").WithIcon(gameui.IconCheck)
}

func combatSummaryPopupHit(fx, fy float64) bool {
	return buildCombatSummaryModal().Panel.Rect.Hit(fx, fy)
}

func combatSummaryButtonHit(fx, fy float64) bool {
	return buildCombatSummaryButton().HitTest(fx, fy)
}

func combatSummaryVisibleRows(viewport gameui.Rect) int {
	rows := int(viewport.H / 58)
	if rows < 1 {
		return 1
	}
	return rows
}

func combatSummaryMaxScroll(entryCount int, viewport gameui.Rect) int {
	maxScroll := entryCount - combatSummaryVisibleRows(viewport)
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func clampCombatSummaryScroll(entryCount int, viewport gameui.Rect, scroll int) int {
	if scroll < 0 {
		return 0
	}
	maxScroll := combatSummaryMaxScroll(entryCount, viewport)
	if scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func combatSummaryRowRect(viewport gameui.Rect, visibleIndex int) gameui.Rect {
	return gameui.Rect{
		X: viewport.X,
		Y: viewport.Y + float64(visibleIndex*58),
		W: viewport.W,
		H: 50,
	}
}

func collectCombatSummaryTotals(entries []CombatSummaryEntry) []combatSummaryFactionTotal {
	if len(entries) == 0 {
		return nil
	}
	totals := make([]combatSummaryFactionTotal, 0, len(entries)*2)
	indices := make(map[faction.FactionID]int, len(entries)*2)
	add := func(id faction.FactionID, lost int, naval, destroyed bool) {
		if id == "" || (lost <= 0 && !destroyed) {
			return
		}
		index, ok := indices[id]
		if !ok {
			index = len(totals)
			indices[id] = index
			totals = append(totals, combatSummaryFactionTotal{FactionID: id})
		}
		total := &totals[index]
		total.Battles++
		if naval {
			total.FleetLost += max(0, lost)
			if destroyed {
				total.FleetDestroyed++
			}
			return
		}
		total.ArmyLost += max(0, lost)
		if destroyed {
			total.ArmyDestroyed++
		}
	}
	for _, entry := range entries {
		add(entry.AttackerFactionID, entry.AttackerLost, entry.AttackerNaval, entry.AttackerDestroyed)
		add(entry.DefenderFactionID, entry.DefenderLost, entry.DefenderNaval, entry.DefenderDestroyed)
	}
	return totals
}

func combatSummaryFactionName(gs *state.GameState, id faction.FactionID) string {
	if gs != nil && gs.Factions != nil {
		if f := gs.Factions[id]; f != nil && f.NameTR != "" {
			return f.NameTR
		}
	}
	return string(id)
}

func drawCombatSummaryScrollbar(screen *ebiten.Image, viewport gameui.Rect, entryCount, scroll int) {
	maxScroll := combatSummaryMaxScroll(entryCount, viewport)
	if maxScroll <= 0 {
		return
	}
	scroll = clampCombatSummaryScroll(entryCount, viewport, scroll)
	track := gameui.Rect{X: viewport.X + viewport.W - 6, Y: viewport.Y, W: 4, H: viewport.H}
	drawUICardRect(screen, track, color.RGBA{22, 20, 16, 210}, color.RGBA{72, 62, 42, 180}, 1)
	thumbH := track.H * float64(combatSummaryVisibleRows(viewport)) / float64(entryCount)
	if thumbH < 24 {
		thumbH = 24
	}
	thumbY := track.Y
	if track.H > thumbH {
		thumbY += (track.H - thumbH) * float64(scroll) / float64(maxScroll)
	}
	drawUICardRect(screen, gameui.Rect{X: track.X, Y: thumbY, W: track.W, H: thumbH}, color.RGBA{176, 144, 78, 230}, color.RGBA{214, 190, 120, 210}, 1)
}

func drawCombatSummaryDialog(screen *ebiten.Image, gs *state.GameState, summary combatSummaryState) {
	modal := buildCombatSummaryModal()
	layout := buildCombatSummaryLayout()
	totals := collectCombatSummaryTotals(summary.data.Entries)
	gameui.DrawModal(screen, modal, eventDetailModalStyle, nil, nil)
	drawUIPanelTopBar(screen, layout.panelRect, 3, panelBorder)
	drawUILabel(screen, layout.titleRect, "Çatışma Özeti", ColorGold, gameui.TextLarge, gameui.TextAlignCenter)
	intro := fmt.Sprintf("Tur %d: %d çatışmada ordu veya donanma kaybı yaşandı.", summary.data.Turn, len(summary.data.Entries))
	drawUILabel(screen, layout.introRect, intro, color.RGBA{220, 214, 202, 255}, gameui.TextMedium, gameui.TextAlignCenter)

	if len(totals) == 0 {
		drawUILabel(screen, layout.listRect, "Kayıtlı çatışma kaybı yok.", ColorGray, gameui.TextMedium, gameui.TextAlignCenter)
	} else {
		scroll := clampCombatSummaryScroll(len(totals), layout.listRect, summary.scroll)
		visibleRows := combatSummaryVisibleRows(layout.listRect)
		end := min(len(totals), scroll+visibleRows)
		for i := scroll; i < end; i++ {
			total := totals[i]
			row := combatSummaryRowRect(layout.listRect, i-scroll)
			accent := color.RGBA{172, 132, 82, 255}
			if gs != nil && total.FactionID == gs.PlayerFactionID {
				accent = color.RGBA{130, 190, 132, 255}
			}
			drawUICardRect(screen, row, color.RGBA{28, 22, 16, 224}, color.RGBA{78, 64, 40, 190}, 1)
			drawUILabel(screen, gameui.Rect{X: row.X + 14, Y: row.Y + 8, W: row.W - 28}, combatSummaryFactionName(gs, total.FactionID), accent, gameui.TextMedium, gameui.TextAlignStart)
			lossText := fmt.Sprintf("Kara ordusu kaybı: %d  |  Donanma kaybı: %d  |  Çatışma: %d", total.ArmyLost, total.FleetLost, total.Battles)
			drawUILabel(screen, gameui.Rect{X: row.X + 14, Y: row.Y + 29, W: row.W - 28}, trimTextToWidth(lossText, FaceSmall, row.W-28), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
			destroyedText := "Tamamen yok edilen: yok"
			if total.ArmyDestroyed > 0 || total.FleetDestroyed > 0 {
				destroyedText = fmt.Sprintf("Tamamen yok edilen: %d ordu, %d donanma", total.ArmyDestroyed, total.FleetDestroyed)
			}
			drawUILabel(screen, gameui.Rect{X: row.X + row.W*0.56, Y: row.Y + 8, W: row.W * 0.40}, trimTextToWidth(destroyedText, FaceSmall, row.W*0.40), color.RGBA{238, 190, 128, 255}, gameui.TextSmall, gameui.TextAlignEnd)
		}
		drawCombatSummaryScrollbar(screen, layout.listRect, len(totals), scroll)
	}
	drawUIButtonWidget(screen, buildCombatSummaryButton(), solidButtonStyle(color.RGBA{70, 98, 62, 235}, color.RGBA{122, 160, 112, 255}, ColorWhite, 10))
}

func (r *Renderer) handleCombatSummaryInput() InputAction {
	if r == nil {
		return InputAction{}
	}
	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	layout := buildCombatSummaryLayout()
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 && layout.listRect.Hit(mx, my) {
		step := 1
		if wheelY > 0 {
			step = -1
		}
		totals := collectCombatSummaryTotals(r.combatSummary.data.Entries)
		r.combatSummary.scroll = clampCombatSummaryScroll(len(totals), layout.listRect, r.combatSummary.scroll+step)
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace) ||
		(r.mouseJustPressed(ebiten.MouseButtonLeft) && combatSummaryButtonHit(mx, my)) {
		r.HideCombatSummary()
	}
	return InputAction{}
}

func (r *Renderer) combatSummaryHovering(fx, fy float64) bool {
	return combatSummaryButtonHit(fx, fy)
}
