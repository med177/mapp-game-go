package render

import (
	"image/color"
	"strings"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

type WarSummaryParticipant struct {
	NameTR   string
	RoleTR   string
	Strength int
}

type WarSummarySide struct {
	Label         string
	LeaderNameTR  string
	TotalStrength int
	Participants  []WarSummaryParticipant
	Refused       []string
}

type WarSummaryReport struct {
	Title        string
	BalanceLabel string
	PowerText    string
	Attacker     WarSummarySide
	Defender     WarSummarySide
}

type warSummaryState struct {
	show           bool
	data           WarSummaryReport
	attackerScroll int
	defenderScroll int
}

type warSummaryLayout struct {
	panelRect        gameui.Rect
	titleRect        gameui.Rect
	balanceRect      gameui.Rect
	attackerRect     gameui.Rect
	defenderRect     gameui.Rect
	footerRect       gameui.Rect
	attackerListRect gameui.Rect
	defenderListRect gameui.Rect
}

func (r *Renderer) ShowWarSummary(report WarSummaryReport) {
	if r == nil {
		return
	}
	if report.Title == "" {
		report.Title = "Savaş Özeti"
	}
	r.warSummary = warSummaryState{
		show: true,
		data: report,
	}
	r.combatLogTimer = 0
}

func (r *Renderer) HideWarSummary() {
	if r == nil {
		return
	}
	r.warSummary = warSummaryState{}
	if r.queuedBattleReport.show {
		report := r.queuedBattleReport.data
		r.queuedBattleReport = battleReportState{}
		r.ShowBattleReport(report)
	}
}

func buildWarSummaryModal() gameui.Modal {
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, 980, 560, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildWarSummaryLayout() warSummaryLayout {
	modal := buildWarSummaryModal()
	panelRect := modal.Panel.Rect
	box := gameui.BoxFromRect(panelRect).Inset(20)
	titleRect, box := box.CutTop(34, 14)
	balanceRect, box := box.CutTop(72, 18)
	sidesRow, footerBox := box.CutTop(box.Rect.H-46, 16)
	cols := gameui.BoxFromRect(sidesRow).SplitColumns(18, 0.5, 0.5)
	layout := warSummaryLayout{
		panelRect:   panelRect,
		titleRect:   titleRect,
		balanceRect: balanceRect,
		footerRect:  footerBox.Rect,
	}
	if len(cols) == 2 {
		layout.attackerRect = cols[0]
		layout.defenderRect = cols[1]
		layout.attackerListRect = warSummaryListRect(cols[0])
		layout.defenderListRect = warSummaryListRect(cols[1])
	}
	return layout
}

func warSummaryListRect(sideRect gameui.Rect) gameui.Rect {
	top := sideRect.Y + 96
	bottom := sideRect.Y + sideRect.H - 16
	if bottom < top {
		bottom = top
	}
	return gameui.Rect{
		X: sideRect.X + 14,
		Y: top,
		W: sideRect.W - 28,
		H: bottom - top,
	}
}

func buildWarSummaryCloseButton() gameui.Button {
	const (
		btnW = 184.0
		btnH = 34.0
	)
	modal := buildWarSummaryModal()
	x := modal.Panel.Rect.X + (modal.Panel.Rect.W-btnW)/2
	y := modal.Panel.Rect.Y + modal.Panel.Rect.H - btnH - 18
	return gameui.NewButton(x, y, btnW, btnH, "Devam Et").WithIcon(gameui.IconCheck)
}

func warSummaryPopupHit(fx, fy float64) bool {
	return buildWarSummaryModal().Panel.Rect.Hit(fx, fy)
}

func warSummaryCloseHit(fx, fy float64) bool {
	return buildWarSummaryCloseButton().HitTest(fx, fy)
}

func warSummaryVisibleRows(viewport gameui.Rect) int {
	rows := int(viewport.H / 52)
	if rows < 1 {
		return 1
	}
	return rows
}

func warSummaryMaxScroll(entryCount int, viewport gameui.Rect) int {
	maxScroll := entryCount - warSummaryVisibleRows(viewport)
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func clampWarSummaryScroll(entryCount int, viewport gameui.Rect, scroll int) int {
	if scroll < 0 {
		return 0
	}
	maxScroll := warSummaryMaxScroll(entryCount, viewport)
	if scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func warSummaryRowRect(viewport gameui.Rect, visibleIndex int) gameui.Rect {
	return gameui.Rect{
		X: viewport.X,
		Y: viewport.Y + float64(visibleIndex*52),
		W: viewport.W,
		H: 44,
	}
}

func drawWarSummaryScrollbar(screen *ebiten.Image, viewport gameui.Rect, entryCount, scroll int) {
	maxScroll := warSummaryMaxScroll(entryCount, viewport)
	if maxScroll <= 0 {
		return
	}
	scroll = clampWarSummaryScroll(entryCount, viewport, scroll)
	track := gameui.Rect{
		X: viewport.X + viewport.W - 6,
		Y: viewport.Y,
		W: 4,
		H: viewport.H,
	}
	drawUICardRect(screen, track, color.RGBA{22, 20, 16, 210}, color.RGBA{72, 62, 42, 180}, 1)
	thumbH := track.H * float64(warSummaryVisibleRows(viewport)) / float64(entryCount)
	if thumbH < 24 {
		thumbH = 24
	}
	thumbY := track.Y
	if track.H > thumbH {
		thumbY += (track.H - thumbH) * float64(scroll) / float64(maxScroll)
	}
	drawUICardRect(screen, gameui.Rect{X: track.X, Y: thumbY, W: track.W, H: thumbH}, color.RGBA{176, 144, 78, 230}, color.RGBA{214, 190, 120, 210}, 1)
}

func drawWarSummarySide(screen *ebiten.Image, sideRect, listRect gameui.Rect, side WarSummarySide, scroll int) {
	drawUICardRect(screen, sideRect, color.RGBA{20, 14, 10, 234}, color.RGBA{94, 74, 42, 210}, 1)
	drawUILabel(screen, gameui.Rect{X: sideRect.X + 14, Y: sideRect.Y + 14, W: sideRect.W - 28}, side.Label, color.RGBA{255, 220, 100, 255}, gameui.TextMedium, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: sideRect.X + 14, Y: sideRect.Y + 38, W: sideRect.W - 28}, "Lider: "+side.LeaderNameTR, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: sideRect.X + 14, Y: sideRect.Y + 58, W: sideRect.W - 28}, "Toplam Güç: "+itoa(side.TotalStrength), color.RGBA{212, 202, 176, 255}, gameui.TextSmall, gameui.TextAlignStart)
	refusedText := "Katılmayan yok."
	if len(side.Refused) > 0 {
		refusedText = "Katılmayan: " + strings.Join(side.Refused, ", ")
	}
	drawUIWrappedLabel(screen, gameui.Rect{X: sideRect.X + 14, Y: sideRect.Y + 76, W: sideRect.W - 28}, refusedText, color.RGBA{174, 146, 118, 255}, gameui.TextSmall, 17, 2)

	if len(side.Participants) == 0 {
		drawUILabel(screen, gameui.Rect{X: listRect.X, Y: listRect.Y + 8, W: listRect.W}, "Aktif katılımcı yok.", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		return
	}
	scroll = clampWarSummaryScroll(len(side.Participants), listRect, scroll)
	visibleRows := warSummaryVisibleRows(listRect)
	end := scroll + visibleRows
	if end > len(side.Participants) {
		end = len(side.Participants)
	}
	for i := scroll; i < end; i++ {
		entry := side.Participants[i]
		rowRect := warSummaryRowRect(listRect, i-scroll)
		drawUICardRect(screen, rowRect, color.RGBA{28, 22, 16, 224}, color.RGBA{78, 64, 40, 190}, 1)
		drawUILabel(screen, gameui.Rect{X: rowRect.X + 12, Y: rowRect.Y + 8, W: rowRect.W - 120}, entry.NameTR, ColorWhite, gameui.TextMedium, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: rowRect.X + rowRect.W - 108, Y: rowRect.Y + 8, W: 96}, "Güç "+itoa(entry.Strength), color.RGBA{204, 190, 146, 255}, gameui.TextSmall, gameui.TextAlignEnd)
		drawUILabel(screen, gameui.Rect{X: rowRect.X + 12, Y: rowRect.Y + 24, W: rowRect.W - 24}, entry.RoleTR, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	}
	drawWarSummaryScrollbar(screen, listRect, len(side.Participants), scroll)
}

func drawWarSummaryDialog(screen *ebiten.Image, state warSummaryState) {
	modal := buildWarSummaryModal()
	layout := buildWarSummaryLayout()
	gameui.DrawModal(screen, modal, standardModalStyle, nil, nil)
	drawUILabel(screen, layout.titleRect, state.data.Title, color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignCenter)
	drawUICardRect(screen, layout.balanceRect, color.RGBA{24, 18, 12, 228}, color.RGBA{102, 78, 42, 210}, 1)
	drawUILabel(screen, gameui.Rect{X: layout.balanceRect.X + 18, Y: layout.balanceRect.Y + 14, W: layout.balanceRect.W - 36}, state.data.BalanceLabel, ColorWhite, gameui.TextMedium, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: layout.balanceRect.X + 18, Y: layout.balanceRect.Y + 40, W: layout.balanceRect.W - 36}, state.data.PowerText, color.RGBA{208, 200, 182, 255}, gameui.TextSmall, gameui.TextAlignCenter)
	drawWarSummarySide(screen, layout.attackerRect, layout.attackerListRect, state.data.Attacker, state.attackerScroll)
	drawWarSummarySide(screen, layout.defenderRect, layout.defenderListRect, state.data.Defender, state.defenderScroll)
	drawUIButtonWidget(screen, buildWarSummaryCloseButton(), solidButtonStyle(color.RGBA{70, 98, 62, 235}, color.RGBA{122, 160, 112, 255}, ColorWhite, 10))
}

func (r *Renderer) handleWarSummaryInput() InputAction {
	if r == nil {
		return InputAction{}
	}
	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	layout := buildWarSummaryLayout()
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		step := 1
		if wheelY > 0 {
			step = -1
		}
		switch {
		case layout.attackerListRect.Hit(mx, my):
			r.warSummary.attackerScroll = clampWarSummaryScroll(len(r.warSummary.data.Attacker.Participants), layout.attackerListRect, r.warSummary.attackerScroll+step)
			return InputAction{}
		case layout.defenderListRect.Hit(mx, my):
			r.warSummary.defenderScroll = clampWarSummaryScroll(len(r.warSummary.data.Defender.Participants), layout.defenderListRect, r.warSummary.defenderScroll+step)
			return InputAction{}
		}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) && (warSummaryCloseHit(mx, my) || !warSummaryPopupHit(mx, my)) {
		r.HideWarSummary()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace) {
		r.HideWarSummary()
	}
	return InputAction{}
}
