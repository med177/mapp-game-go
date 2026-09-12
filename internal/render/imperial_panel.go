package render

import (
	"image/color"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	imperialPanelMaxW          = 980.0
	imperialPanelMaxH          = 650.0
	imperialPanelMinW          = 620.0
	imperialPanelMinH          = 500.0
	imperialPanelHeaderH       = 54.0
	imperialPanelSummaryH      = 88.0
	imperialPanelMemberRowH    = 56.0
	imperialPanelMemberFooterH = 28.0
	imperialDecisionOptionRowH = 68.0
	imperialDecisionDescY      = 46.0
)

func imperialPanelRect() gameui.Rect {
	w := imperialPanelMaxW
	h := imperialPanelMaxH
	if ScreenWidth-32 < w {
		w = ScreenWidth - 32
	}
	if ScreenHeight-32 < h {
		h = ScreenHeight - 32
	}
	if w < imperialPanelMinW {
		w = ScreenWidth - 20
	}
	if h < imperialPanelMinH {
		h = ScreenHeight - 20
	}
	if w < 280 {
		w = ScreenWidth
	}
	if h < 240 {
		h = ScreenHeight
	}
	return gameui.Rect{
		X: (ScreenWidth - w) / 2,
		Y: (ScreenHeight - h) / 2,
		W: w,
		H: h,
	}
}

func imperialPanelCloseButton() gameui.Button {
	r := imperialPanelRect()
	btn := gameui.NewButton(r.X+r.W-42, r.Y+12, 28, 28, "").WithIcon(gameui.IconClose)
	btn.IconSize = 14
	return btn
}

func imperialPanelAvailableCandidates(gs *state.GameState) []faction.FactionID {
	if gs == nil || gs.Imperial == nil {
		return nil
	}
	ids := []faction.FactionID{gs.Imperial.EmpireID}
	for id, member := range gs.Imperial.Members {
		if member == nil || member.ElectorWeight <= 0 || gs.Factions[id] == nil {
			continue
		}
		ids = append(ids, id)
	}
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			if ids[j] < ids[i] {
				ids[i], ids[j] = ids[j], ids[i]
			}
		}
	}
	return ids
}

func imperialMemberStatusLabel(status state.ImperialMemberStatus) string {
	switch status {
	case state.ImperialMemberElector:
		return "Elektör"
	case state.ImperialMemberPrince:
		return "Prenslik"
	case state.ImperialMemberFreeCity:
		return "Özgür şehir"
	case state.ImperialMemberOrder:
		return "Tarikat"
	case state.ImperialMemberVassal:
		return "Vassal"
	default:
		return "Üye"
	}
}

func imperialDecisionTitle(kind state.ImperialDecisionKind) string {
	switch kind {
	case state.ImperialDecisionDiet:
		return "İmparatorluk Diyeti"
	case state.ImperialDecisionElection:
		return "İmparatorluk Seçimi"
	default:
		return "İmparatorluk Kararı"
	}
}

func imperialDecisionButtonRects(r gameui.Rect) [3]gameui.Rect {
	buttonW := r.W - 42
	if buttonW > 330 {
		buttonW = 330
	}
	x := r.X + r.W - buttonW - 20
	startY := r.Y + 126
	return [3]gameui.Rect{
		{X: x, Y: startY, W: buttonW, H: 42},
		{X: x, Y: startY + imperialDecisionOptionRowH, W: buttonW, H: 42},
		{X: x, Y: startY + imperialDecisionOptionRowH*2, W: buttonW, H: 42},
	}
}

func imperialElectionButtonRects(r gameui.Rect, count int) []gameui.Rect {
	if count <= 0 {
		return nil
	}
	buttonW := r.W - 42
	if buttonW > 330 {
		buttonW = 330
	}
	x := r.X + r.W - buttonW - 20
	buttons := make([]gameui.Rect, count)
	for i := range buttons {
		buttons[i] = gameui.Rect{X: x, Y: r.Y + 130 + float64(i*50), W: buttonW, H: 42}
	}
	return buttons
}

// imperialMemberListLayout çizim ve input tarafının aynı liste/alt bilgi
// geometrisini kullanmasını sağlar. Alt bilgi alanı satırların üzerine
// binmemesi için viewport'tan özellikle ayrılır.
func imperialMemberListLayout(panel gameui.Rect) (left, viewport, footer gameui.Rect, visible int) {
	left = gameui.Rect{X: panel.X + 18, Y: panel.Y + 160, W: panel.W * 0.57, H: panel.H - 178}
	footer = gameui.Rect{
		X: left.X,
		Y: left.Y + left.H - imperialPanelMemberFooterH,
		W: left.W,
		H: imperialPanelMemberFooterH,
	}
	viewport = gameui.Rect{
		X: left.X,
		Y: left.Y + 30,
		W: left.W,
		H: footer.Y - (left.Y + 30) - 4,
	}
	if viewport.H < imperialPanelMemberRowH {
		viewport.H = imperialPanelMemberRowH
	}
	visible = int(viewport.H / imperialPanelMemberRowH)
	if visible < 1 {
		visible = 1
	}
	return left, viewport, footer, visible
}

func imperialDecisionBoxRect(panel gameui.Rect) gameui.Rect {
	return gameui.Rect{X: panel.X + panel.W*0.60, Y: panel.Y + 160, W: panel.W*0.38 - 18, H: panel.H - 178}
}

func imperialMemberRowRect(viewport gameui.Rect, visibleIndex int) gameui.Rect {
	return gameui.Rect{
		X: viewport.X,
		Y: viewport.Y + float64(visibleIndex)*imperialPanelMemberRowH,
		W: viewport.W - 10,
		H: imperialPanelMemberRowH - 4,
	}
}

func clampImperialMemberScroll(total, visible, scroll int) int {
	maxScroll := total - visible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll < 0 {
		return 0
	}
	if scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func drawImperialMemberScrollbar(screen *ebiten.Image, viewport gameui.Rect, total, visible, scroll int) {
	maxScroll := total - visible
	if maxScroll <= 0 {
		return
	}
	track := gameui.Rect{X: viewport.X + viewport.W - 6, Y: viewport.Y + 2, W: 4, H: viewport.H - 4}
	drawRoundedRect(screen, float32(track.X), float32(track.Y), float32(track.W), float32(track.H), 2, color.RGBA{64, 52, 34, 180})
	thumbH := track.H * float64(visible) / float64(total)
	if thumbH < 22 {
		thumbH = 22
	}
	scroll = clampImperialMemberScroll(total, visible, scroll)
	thumbY := track.Y + (track.H-thumbH)*float64(scroll)/float64(maxScroll)
	drawRoundedRect(screen, float32(track.X), float32(thumbY), float32(track.W), float32(thumbH), 2, color.RGBA{204, 164, 76, 235})
}

func (r *Renderer) DrawImperialPanel(screen *ebiten.Image) {
	if r == nil || r.gs == nil || !imperialPanelAvailable(r.gs) {
		return
	}
	panel := imperialPanelRect()
	drawUIOverlay(screen, color.RGBA{8, 6, 4, 210})
	drawUIPanelFrame(screen, panel, color.RGBA{18, 14, 9, 248}, color.RGBA{170, 132, 58, 255}, 2, 6)
	vector.FillRect(screen, float32(panel.X), float32(panel.Y), float32(panel.W), 4, color.RGBA{208, 169, 74, 255}, false)
	drawUILabel(screen, gameui.Rect{X: panel.X + 20, Y: panel.Y + 12, W: panel.W - 80}, "Kutsal Roma İmparatorluğu", color.RGBA{255, 220, 116, 255}, gameui.TextLarge, gameui.TextAlignStart)
	drawUIButtonWidget(screen, imperialPanelCloseButton(), tinyButtonStyle)

	imperial := r.gs.Imperial
	emperorName := factionLabelForRender(r.gs, imperial.EmperorID)
	emperorName = trimTextToWidth(emperorName, FaceMed, 235)
	lastCall := "Son imparatorluk savaşı yok"
	if imperial.LastWarCall != nil {
		lastCall = "Son çağrı: " + factionLabelForRender(r.gs, imperial.LastWarCall.EnemyID) + " (Tur " + itoa(imperial.LastWarCall.StartedTurn) + ")"
	}
	summary := gameui.Rect{X: panel.X + 18, Y: panel.Y + imperialPanelHeaderH, W: panel.W - 36, H: imperialPanelSummaryH}
	drawUIPanelFrame(screen, summary, color.RGBA{30, 24, 14, 235}, color.RGBA{100, 78, 40, 230}, 1, 3)
	columnW := (summary.W - 28) / 3
	drawUILabel(screen, gameui.Rect{X: summary.X + 14, Y: summary.Y + 10, W: columnW}, "İmparator", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: summary.X + 14, Y: summary.Y + 30, W: columnW}, emperorName, ColorWhite, gameui.TextMedium, gameui.TextAlignStart)
	authorityX := summary.X + 14 + columnW + 8
	drawUILabel(screen, gameui.Rect{X: authorityX, Y: summary.Y + 10, W: columnW}, "Otorite", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: authorityX, Y: summary.Y + 30, W: columnW}, itoa(imperial.Authority)+" / 100", imperialAuthorityColor(imperial.Authority), gameui.TextMedium, gameui.TextAlignStart)
	drawUIProgressBar(screen, float32(authorityX), float32(summary.Y+53), float32(columnW-8), 7, float64(imperial.Authority)/100, color.RGBA{20, 18, 13, 230}, color.RGBA{92, 70, 36, 220}, imperialAuthorityColor(imperial.Authority), 1)
	calendarX := authorityX + columnW + 8
	drawUILabel(screen, gameui.Rect{X: calendarX, Y: summary.Y + 10, W: columnW}, "Siyasi takvim", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	calendar := trimTextToWidth(imperialCalendarText(r.gs), FaceSmall, columnW)
	drawUILabel(screen, gameui.Rect{X: calendarX, Y: summary.Y + 30, W: columnW}, calendar, ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
	lastCall = trimTextToWidth(lastCall, FaceSmall, summary.W-28)
	drawUILabel(screen, gameui.Rect{X: summary.X + 14, Y: summary.Y + 58, W: summary.W - 28}, lastCall, color.RGBA{190, 174, 136, 235}, gameui.TextSmall, gameui.TextAlignStart)

	members := diplomacy.ImperialMembersOf(r.gs, imperial.EmpireID)
	left, viewport, footer, visible := imperialMemberListLayout(panel)
	drawUILabel(screen, gameui.Rect{X: left.X, Y: left.Y, W: left.W}, "İmparatorluk Üyeleri", ColorGold, gameui.TextMedium, gameui.TextAlignStart)
	maxScroll := len(members) - visible
	if maxScroll < 0 {
		maxScroll = 0
	}
	r.imperialScroll = clampImperialMemberScroll(len(members), visible, r.imperialScroll)
	for row := 0; row < visible; row++ {
		idx := r.imperialScroll + row
		if idx >= len(members) {
			break
		}
		memberID := members[idx]
		member := imperial.Members[memberID]
		if member == nil {
			continue
		}
		rowRect := imperialMemberRowRect(viewport, row)
		fill := color.RGBA{28, 23, 15, 225}
		border := color.RGBA{76, 61, 34, 220}
		drawUICardRect(screen, rowRect, fill, border, 1)
		name := factionLabelForRender(r.gs, memberID)
		nameW := rowRect.W - 78
		drawUILabel(screen, gameui.Rect{X: rowRect.X + 10, Y: rowRect.Y + 5, W: nameW}, trimTextToWidth(name, FaceSmall, nameW), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: rowRect.X + rowRect.W - 68, Y: rowRect.Y + 5, W: 58}, "Oy "+itoa(member.ElectorWeight), color.RGBA{218, 188, 110, 255}, gameui.TextSmall, gameui.TextAlignEnd)
		drawUILabel(screen, gameui.Rect{X: rowRect.X + 10, Y: rowRect.Y + 27, W: 78}, imperialMemberStatusLabel(member.Status), color.RGBA{198, 170, 108, 255}, gameui.TextSmall, gameui.TextAlignStart)
		drawImperialMemberMetric(screen, rowRect.X+94, rowRect.Y+25, 102, "Sadakat", member.Loyalty, color.RGBA{104, 184, 126, 255})
		drawImperialMemberMetric(screen, rowRect.X+204, rowRect.Y+25, 102, "Özerklik", member.Autonomy, color.RGBA{112, 158, 208, 255})
		drawImperialMemberMetric(screen, rowRect.X+314, rowRect.Y+25, rowRect.W-324, "Askerî", member.MilitaryCommitment, color.RGBA{204, 136, 88, 255})
	}
	drawImperialMemberScrollbar(screen, viewport, len(members), visible, r.imperialScroll)
	footerText := "Satıra tıklayarak diplomasi aç"
	if len(members) > visible {
		footerText = "Tekerlek: listeyi kaydır • Satır: diplomasi"
	}
	if len(members) > 0 {
		drawUILabel(screen, gameui.Rect{X: footer.X, Y: footer.Y + 5, W: footer.W}, footerText, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	}

	r.drawImperialDecisionArea(screen, panel)
}

func drawImperialMemberMetric(screen *ebiten.Image, x, y, w float64, label string, value int, fill color.Color) {
	if w < 24 {
		return
	}
	labelW := w
	if labelW > 72 {
		labelW = 72
	}
	drawUILabel(screen, gameui.Rect{X: x, Y: y, W: labelW}, label, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	barX := x + labelW
	barW := w - labelW
	if barW >= 28 {
		drawUIProgressBar(screen, float32(barX), float32(y+3), float32(barW-22), 6, float64(value)/100, color.RGBA{20, 18, 13, 220}, color.RGBA{74, 58, 36, 180}, fill, 1)
		drawUILabel(screen, gameui.Rect{X: x + w - 20, Y: y - 1, W: 20}, itoa(value), ColorWhite, gameui.TextSmall, gameui.TextAlignEnd)
	}
}

func factionLabelForRender(gs *state.GameState, id faction.FactionID) string {
	if gs != nil && gs.Factions[id] != nil {
		if gs.Factions[id].NameTR != "" {
			return gs.Factions[id].NameTR
		}
		return gs.Factions[id].Name
	}
	return string(id)
}

func imperialAuthorityColor(authority int) color.Color {
	switch {
	case authority >= 70:
		return color.RGBA{124, 220, 126, 255}
	case authority >= 40:
		return color.RGBA{232, 196, 96, 255}
	default:
		return color.RGBA{230, 116, 96, 255}
	}
}

func imperialCalendarText(gs *state.GameState) string {
	if gs == nil || gs.Imperial == nil {
		return "-"
	}
	if pending := gs.Imperial.PendingDecision; pending != nil {
		return imperialDecisionTitle(pending.Kind) + " bekliyor"
	}
	parts := "Diyet Tur " + itoa(gs.Imperial.NextDietTurn)
	if gs.Imperial.ElectionDueTurn > 0 {
		parts += " • Seçim Tur " + itoa(gs.Imperial.ElectionDueTurn)
	}
	return parts
}

func (r *Renderer) drawImperialDecisionArea(screen *ebiten.Image, panel gameui.Rect) {
	box := imperialDecisionBoxRect(panel)
	drawUICardRect(screen, box, color.RGBA{26, 21, 14, 235}, color.RGBA{90, 70, 38, 230}, 1)
	pending := r.gs.Imperial.PendingDecision
	if pending == nil {
		drawUILabel(screen, gameui.Rect{X: box.X + 14, Y: box.Y + 16, W: box.W - 28}, "İmparatorluk yönetimi", ColorGold, gameui.TextMedium, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: box.X + 14, Y: box.Y + 54, W: box.W - 28}, "Diyet ve seçim zamanı geldiğinde", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: box.X + 14, Y: box.Y + 74, W: box.W - 28}, "bu alanda oyuncu kararı açılır.", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		return
	}

	drawUILabel(screen, gameui.Rect{X: box.X + 14, Y: box.Y + 14, W: box.W - 28}, imperialDecisionTitle(pending.Kind), ColorGold, gameui.TextMedium, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: box.X + 14, Y: box.Y + 48, W: box.W - 28}, "Bu karar verilmeden tur bitirilemez.", color.RGBA{236, 190, 126, 255}, gameui.TextSmall, gameui.TextAlignStart)
	if pending.Kind == state.ImperialDecisionDiet {
		labels := []string{"Merkezî otoriteyi güçlendir", "Prenslik imtiyazlarını koru", "Askerî katkı talep et"}
		desc := []string{"+5 otorite, üyelerde sadakat maliyeti", "+1 otorite, sadakat ve özerklik artışı", "+2 otorite, askerî bağlılık artışı"}
		buttons := imperialDecisionButtonRects(box)
		for i, rect := range buttons {
			btn := gameui.NewButton(rect.X, rect.Y, rect.W, rect.H, labels[i])
			drawUIButtonWidget(screen, btn, solidButtonStyle(color.RGBA{86, 63, 28, 240}, color.RGBA{190, 148, 70, 255}, ColorWhite, 9))
			drawUILabel(screen, gameui.Rect{X: rect.X, Y: rect.Y + imperialDecisionDescY, W: rect.W}, desc[i], ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		}
		return
	}

	drawUILabel(screen, gameui.Rect{X: box.X + 14, Y: box.Y + 78, W: box.W - 28}, "Bir aday seçerek siyasi desteğinizi belirleyin.", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	candidates := imperialPanelAvailableCandidates(r.gs)
	buttons := imperialElectionButtonRects(box, len(candidates))
	for i, candidate := range candidates {
		if i >= len(buttons) {
			break
		}
		label := factionLabelForRender(r.gs, candidate)
		btn := gameui.NewButton(buttons[i].X, buttons[i].Y, buttons[i].W, buttons[i].H, label)
		drawUIButtonWidget(screen, btn, solidButtonStyle(color.RGBA{58, 68, 92, 240}, color.RGBA{128, 154, 194, 255}, ColorWhite, 9))
	}
}

func (r *Renderer) handleImperialPanelInput() InputAction {
	if r == nil || r.gs == nil {
		return InputAction{}
	}
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	if r.keyJustPressed(ebiten.KeyEscape) {
		if r.gs.Imperial == nil || r.gs.Imperial.PendingDecision == nil {
			r.CloseImperialPanel()
		}
		return InputAction{}
	}
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		if wheelY > 0 {
			r.imperialScroll--
		} else {
			r.imperialScroll++
		}
	}
	if !r.mouseJustPressed(ebiten.MouseButtonLeft) {
		return InputAction{}
	}
	if imperialPanelCloseButton().HitTest(fx, fy) {
		if r.gs.Imperial == nil || r.gs.Imperial.PendingDecision == nil {
			r.CloseImperialPanel()
		}
		return InputAction{}
	}
	panel := imperialPanelRect()
	if !panel.Hit(fx, fy) {
		if r.gs.Imperial == nil || r.gs.Imperial.PendingDecision == nil {
			r.CloseImperialPanel()
		}
		return InputAction{}
	}
	if pending := r.gs.Imperial.PendingDecision; pending != nil {
		box := imperialDecisionBoxRect(panel)
		if pending.Kind == state.ImperialDecisionDiet {
			for i, rect := range imperialDecisionButtonRects(box) {
				if gameui.NewButton(rect.X, rect.Y, rect.W, rect.H, "").HitTest(fx, fy) {
					return InputAction{Kind: ActionImperialDietChoice, ChoiceIndex: i}
				}
			}
		} else {
			candidates := imperialPanelAvailableCandidates(r.gs)
			for i, rect := range imperialElectionButtonRects(box, len(candidates)) {
				if i < len(candidates) && gameui.NewButton(rect.X, rect.Y, rect.W, rect.H, "").HitTest(fx, fy) {
					return InputAction{Kind: ActionImperialElectionVote, TargetFaction: candidates[i]}
				}
			}
		}
		return InputAction{}
	}

	members := diplomacy.ImperialMembersOf(r.gs, r.gs.Imperial.EmpireID)
	_, viewport, _, visible := imperialMemberListLayout(panel)
	for row := 0; row < visible; row++ {
		idx := r.imperialScroll + row
		if idx < 0 || idx >= len(members) {
			continue
		}
		rowRect := imperialMemberRowRect(viewport, row)
		if rowRect.Hit(fx, fy) {
			target := members[idx]
			r.CloseImperialPanel()
			r.openDiplomacyTarget(target, 0)
			return InputAction{}
		}
	}
	return InputAction{}
}
