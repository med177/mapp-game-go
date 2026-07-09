package render

import (
	"image/color"
	"sort"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	diplomRowH            = 58.0
	diplomHistoryPanelW   = 286.0
	diplomHistoryPanelGap = 12.0
	diplomOfferMainW      = 430.0
	diplomHistoryPanelH   = 324.0
)

type diplomacyHistoryDirectionFilter int

const (
	diplomacyHistoryDirectionAll diplomacyHistoryDirectionFilter = iota
	diplomacyHistoryDirectionIncoming
	diplomacyHistoryDirectionOutgoing
)

type diplomacyHistoryFilterButton struct {
	Button    gameui.Button
	Direction diplomacyHistoryDirectionFilter
	Action    ActionKind
	IsAction  bool
}

type diplomacyHistoryActionMeta struct {
	Action ActionKind
	Label  string
	Icon   gameui.IconID
	Color  color.RGBA
}

type diplomAction struct {
	label  string
	color  color.RGBA
	action ActionKind
}

var diplomActions = []diplomAction{
	{diplomacy.ActionLabelTR(diplomacy.ActionDeclareWar), color.RGBA{180, 50, 50, 220}, ActionDeclareWar},
	{diplomacy.ActionLabelTR(diplomacy.ActionProposePeace), color.RGBA{50, 120, 180, 220}, ActionProposePeace},
	{diplomacy.ActionLabelTR(diplomacy.ActionProposeAlliance), color.RGBA{50, 160, 80, 220}, ActionProposeAlliance},
	{diplomacy.ActionLabelTR(diplomacy.ActionProposeTrade), color.RGBA{160, 130, 50, 220}, ActionProposeTrade},
}

var diplomacyHistoryActions = [4]diplomacyHistoryActionMeta{
	{Action: ActionProposePeace, Label: "Barış", Icon: gameui.IconCheck, Color: color.RGBA{54, 118, 176, 220}},
	{Action: ActionProposeTrade, Label: "Ticaret", Icon: gameui.IconSend, Color: color.RGBA{164, 128, 44, 220}},
	{Action: ActionProposeAlliance, Label: "İttifak", Icon: gameui.IconBook, Color: color.RGBA{52, 146, 74, 220}},
	{Action: ActionDeclareWar, Label: "Savaş", Icon: gameui.IconSword, Color: color.RGBA{170, 58, 58, 220}},
}

func actionKindForDiplomacyAction(action diplomacy.Action) ActionKind {
	switch action {
	case diplomacy.ActionDeclareWar:
		return ActionDeclareWar
	case diplomacy.ActionProposePeace:
		return ActionProposePeace
	case diplomacy.ActionProposeAlliance:
		return ActionProposeAlliance
	case diplomacy.ActionProposeTrade:
		return ActionProposeTrade
	default:
		return ActionNone
	}
}

func diplomacyActionDisabledReason(gs *state.GameState, target faction.FactionID, action ActionKind) string {
	if gs == nil || target == "" {
		return ""
	}
	rel := diplomacy.Relation(gs, gs.PlayerFactionID, target)
	stance := faction.StancePeace
	if rel != nil {
		stance = rel.Stance
	}
	switch action {
	case ActionDeclareWar:
		if stance == faction.StanceWar {
			return "Zaten savaş halindesin."
		}
	case ActionProposePeace:
		if stance != faction.StanceWar {
			return "Barış teklifi sadece savaşta yapılır."
		}
	case ActionProposeAlliance:
		if stance == faction.StanceWar {
			return "Savaş halindeyken ittifak teklif edilemez."
		}
		if stance == faction.StanceAllied {
			return "Zaten müttefiksin."
		}
	case ActionProposeTrade:
		if stance == faction.StanceWar {
			return "Savaş halindeyken ticaret teklif edilemez."
		}
		if stance == faction.StanceTrade && diplomacy.HasTradeRouteBetween(gs, gs.PlayerFactionID, target) {
			return "Zaten ticaret anlaşması aktif."
		}
		if stance == faction.StanceAllied && diplomacy.HasTradeRouteBetween(gs, gs.PlayerFactionID, target) {
			return "Bu müttefik ile ticaret zaten aktif."
		}
		if assessment := diplomacy.AssessTradeProposal(gs, rel, gs.PlayerFactionID, target); assessment.BlockReason != "" {
			return assessment.BlockReason
		}
	}
	return ""
}

func minF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

type rectF struct {
	x float64
	y float64
	w float64
	h float64
}

type diplomacyListLayout struct {
	panelRect   gameui.Rect
	titleRect   gameui.Rect
	listRect    gameui.Rect
	historyRect gameui.Rect
	footerRect  gameui.Rect
}

type diplomacyOfferLayout struct {
	panelRect    gameui.Rect
	titleRect    gameui.Rect
	targetRect   gameui.Rect
	statusRect   gameui.Rect
	actionsRect  gameui.Rect
	selectedRect gameui.Rect
	historyRect  gameui.Rect
	backRect     gameui.Rect
	sendRect     gameui.Rect
}

func listPageRect() rectF {
	w := minF(ScreenWidth-80, 1100)
	h := ScreenHeight - 190
	if h < 240 {
		h = 240
	}
	x := (ScreenWidth - w) / 2
	y := (ScreenHeight - h) / 2
	return rectF{x: x, y: y, w: w, h: h}
}

func diplomacyListLayoutForScreen() diplomacyListLayout {
	r := listPageRect()
	panel := gameui.Rect{X: r.x, Y: r.y, W: r.w, H: r.h}
	box := gameui.BoxFromRect(panel).InsetXY(14, 14)
	titleRect, box := box.CutTop(24, 14)
	footerRect, box := box.CutBottom(20, 0)
	listRect := box.Rect
	historyRect := gameui.Rect{}
	const listHistoryThreshold = 1058.0
	if listRect.W >= listHistoryThreshold {
		listRect.W = 760
		historyRect = gameui.Rect{
			X: listRect.X + listRect.W + diplomHistoryPanelGap,
			Y: listRect.Y,
			W: diplomHistoryPanelW,
			H: minF(listRect.H, diplomHistoryPanelH),
		}
	}
	return diplomacyListLayout{
		panelRect:   panel,
		titleRect:   titleRect,
		listRect:    listRect,
		historyRect: historyRect,
		footerRect:  footerRect,
	}
}

func diplomacyOfferLayoutForScreen() diplomacyOfferLayout {
	r := offerPageRect()
	panel := gameui.Rect{X: r.x, Y: r.y, W: r.w, H: r.h}
	contentRect := gameui.Rect{X: panel.X + 16, Y: panel.Y + 16, W: diplomOfferMainW, H: panel.H - 32}
	box := gameui.BoxFromRect(contentRect)
	headerRect, box := box.CutTop(28, 14)
	statusRect, box := box.CutTop(78, 22)
	actionsRect, box := box.CutTop(float64(len(diplomActions))*42+float64(len(diplomActions)-1)*12, 22)
	selectedRect, box := box.CutTop(28, 18)
	footerRect, _ := box.CutBottom(40, 0)
	footerCols := gameui.BoxFromRect(footerRect).SplitColumns(12, 1, 1)
	return diplomacyOfferLayout{
		panelRect:    panel,
		titleRect:    headerRect,
		targetRect:   gameui.Rect{X: statusRect.X, Y: statusRect.Y, W: statusRect.W, H: 28},
		statusRect:   gameui.Rect{X: statusRect.X, Y: statusRect.Y + 32, W: statusRect.W, H: statusRect.H - 32},
		actionsRect:  actionsRect,
		selectedRect: selectedRect,
		historyRect: gameui.Rect{
			X: contentRect.X + contentRect.W + diplomHistoryPanelGap,
			Y: contentRect.Y,
			W: diplomHistoryPanelW,
			H: minF(contentRect.H, diplomHistoryPanelH),
		},
		backRect: footerCols[0],
		sendRect: footerCols[1],
	}
}

func diplomacyOfferHistoryRelevant(entry state.DiplomaticOfferHistoryEntry, playerID faction.FactionID) bool {
	return entry.FromFactionID == playerID || entry.ToFactionID == playerID
}

func diplomacyOfferHistoryDirectionTR(entry state.DiplomaticOfferHistoryEntry, playerID faction.FactionID) string {
	if entry.FromFactionID == playerID {
		return "Giden"
	}
	if entry.ToFactionID == playerID {
		return "Gelen"
	}
	return "İlgili"
}

func diplomacyHistoryDirectionLabelTR(dir diplomacyHistoryDirectionFilter) string {
	switch dir {
	case diplomacyHistoryDirectionIncoming:
		return "Gelen"
	case diplomacyHistoryDirectionOutgoing:
		return "Giden"
	default:
		return "Tümü"
	}
}

func diplomacyHistoryActionLabelTR(action ActionKind) string {
	switch action {
	case ActionProposePeace:
		return "Barış"
	case ActionProposeTrade:
		return "Ticaret"
	case ActionProposeAlliance:
		return "İttifak"
	case ActionDeclareWar:
		return "Savaş"
	default:
		return "Tümü"
	}
}

func diplomacyHistoryBrowseLabelTR() string {
	return "Geçmişten seçildi"
}

func diplomacyHistoryActionIcon(action ActionKind) gameui.IconID {
	switch action {
	case ActionProposePeace:
		return gameui.IconCheck
	case ActionProposeTrade:
		return gameui.IconSend
	case ActionProposeAlliance:
		return gameui.IconBook
	case ActionDeclareWar:
		return gameui.IconSword
	default:
		return gameui.IconNone
	}
}

func diplomacyHistoryActionColor(action ActionKind) color.RGBA {
	switch action {
	case ActionProposePeace:
		return color.RGBA{54, 118, 176, 220}
	case ActionProposeTrade:
		return color.RGBA{164, 128, 44, 220}
	case ActionProposeAlliance:
		return color.RGBA{52, 146, 74, 220}
	case ActionDeclareWar:
		return color.RGBA{170, 58, 58, 220}
	default:
		return color.RGBA{96, 88, 68, 220}
	}
}

func diplomacyHistoryOutcomeBadgeTR(entry state.DiplomaticOfferHistoryEntry) string {
	if !entry.Accepted {
		return "RET"
	}
	if entry.Applied {
		return "UYG"
	}
	return "KAB"
}

func diplomacyHistoryDirectionMatches(entry state.DiplomaticOfferHistoryEntry, playerID faction.FactionID, dir diplomacyHistoryDirectionFilter) bool {
	switch dir {
	case diplomacyHistoryDirectionIncoming:
		return entry.ToFactionID == playerID
	case diplomacyHistoryDirectionOutgoing:
		return entry.FromFactionID == playerID
	default:
		return true
	}
}

func diplomacyHistoryActionForEntry(entry state.DiplomaticOfferHistoryEntry) ActionKind {
	if kind := actionKindForDiplomacyAction(diplomacy.Action(entry.Action)); kind != ActionNone {
		return kind
	}
	return ActionNone
}

func diplomacyHistoryActionMatches(entry state.DiplomaticOfferHistoryEntry, actionFilter ActionKind) bool {
	if actionFilter == ActionNone {
		return true
	}
	return diplomacyHistoryActionForEntry(entry) == actionFilter
}

func diplomacyOfferHistoryMatches(entry state.DiplomaticOfferHistoryEntry, playerID faction.FactionID, dirFilter diplomacyHistoryDirectionFilter, actionFilter ActionKind) bool {
	if !diplomacyOfferHistoryRelevant(entry, playerID) {
		return false
	}
	if !diplomacyHistoryDirectionMatches(entry, playerID, dirFilter) {
		return false
	}
	if !diplomacyHistoryActionMatches(entry, actionFilter) {
		return false
	}
	return true
}

func diplomacyOfferHistoryFilterLabelTR(dirFilter diplomacyHistoryDirectionFilter, actionFilter ActionKind) string {
	return "Filtre: " + diplomacyHistoryDirectionLabelTR(dirFilter) + " / " + diplomacyHistoryActionLabelTR(actionFilter)
}

func diplomacyHistoryFilterButtonStyle(active bool, accent color.RGBA) gameui.ButtonStyle {
	style := solidButtonStyle(color.RGBA{30, 24, 17, 220}, color.RGBA{88, 72, 40, 170}, color.RGBA{230, 224, 214, 255}, 5)
	style.TextVariant = gameui.TextSmall
	style.BorderWidth = 1
	if active {
		style.BG = accent
		style.Border = color.RGBA{220, 200, 140, 255}
		style.Text = ColorWhite
	} else {
		style.BG = color.RGBA{24, 18, 12, 218}
		style.Border = color.RGBA{78, 62, 34, 160}
		style.Text = color.RGBA{210, 202, 186, 255}
	}
	return style
}

func buildDiplomacyHistoryFilterButtons(panelRect gameui.Rect, dirFilter diplomacyHistoryDirectionFilter, actionFilter ActionKind) [7]diplomacyHistoryFilterButton {
	var buttons [7]diplomacyHistoryFilterButton
	if panelRect.W <= 0 || panelRect.H <= 0 {
		return buttons
	}
	const (
		padX  = 10.0
		gap   = 6.0
		rowH  = 22.0
		row1Y = 80.0
		row2Y = 106.0
	)
	dirBtnW := (panelRect.W - padX*2 - gap*2) / 3
	actionBtnW := (panelRect.W - padX*2 - gap*3) / 4
	dirLabels := [3]struct {
		filter diplomacyHistoryDirectionFilter
		label  string
	}{
		{diplomacyHistoryDirectionAll, "Tümü"},
		{diplomacyHistoryDirectionIncoming, "Gelen"},
		{diplomacyHistoryDirectionOutgoing, "Giden"},
	}
	for i, item := range dirLabels {
		x := panelRect.X + padX + float64(i)*(dirBtnW+gap)
		btn := gameui.NewButton(x, panelRect.Y+row1Y, dirBtnW, rowH, item.label)
		buttons[i] = diplomacyHistoryFilterButton{
			Button:    btn,
			Direction: item.filter,
		}
	}
	for i, meta := range diplomacyHistoryActions {
		x := panelRect.X + padX + float64(i)*(actionBtnW+gap)
		btn := gameui.NewButton(x, panelRect.Y+row2Y, actionBtnW, rowH, meta.Label).WithIcon(meta.Icon)
		btn.IconSize = 11
		btn.IconGap = 4
		buttons[3+i] = diplomacyHistoryFilterButton{
			Button:   btn,
			Action:   meta.Action,
			IsAction: true,
		}
	}
	return buttons
}

func diplomacyHistoryFilterHit(panelRect gameui.Rect, dirFilter diplomacyHistoryDirectionFilter, actionFilter ActionKind, mx, my float64) (diplomacyHistoryDirectionFilter, ActionKind, bool) {
	buttons := buildDiplomacyHistoryFilterButtons(panelRect, dirFilter, actionFilter)
	for _, btn := range buttons {
		if !btn.Button.HitTest(mx, my) {
			continue
		}
		if btn.IsAction {
			return dirFilter, btn.Action, true
		}
		return btn.Direction, actionFilter, true
	}
	return diplomacyHistoryDirectionAll, ActionNone, false
}

func (r *Renderer) applyDiplomacyHistoryFilterHit(panelRect gameui.Rect, mx, my float64) bool {
	dir, action, ok := diplomacyHistoryFilterHit(panelRect, r.diplomacyHistoryDirectionFilter, r.diplomacyHistoryActionFilter, mx, my)
	if !ok {
		return false
	}
	r.diplomacyHistoryDirectionFilter = dir
	r.diplomacyHistoryActionFilter = action
	return true
}

func diplomacyOfferHistoryOtherFaction(entry state.DiplomaticOfferHistoryEntry, playerID faction.FactionID) (faction.FactionID, bool) {
	if entry.FromFactionID == playerID {
		return entry.ToFactionID, true
	}
	if entry.ToFactionID == playerID {
		return entry.FromFactionID, true
	}
	return "", false
}

func diplomacyOfferHistoryCardRect(panelRect gameui.Rect, drawn int) gameui.Rect {
	return gameui.Rect{
		X: panelRect.X + 10,
		Y: panelRect.Y + 134 + float64(drawn)*44,
		W: panelRect.W - 20,
		H: 38,
	}
}

func diplomacyOfferHistorySelection(gs *state.GameState, panelRect gameui.Rect, mx, my float64, maxEntries int, dirFilter diplomacyHistoryDirectionFilter, actionFilter ActionKind) (faction.FactionID, int, bool) {
	if gs == nil || maxEntries <= 0 {
		return "", 0, false
	}
	drawn := 0
	for i := len(gs.DiplomaticOfferHistory) - 1; i >= 0 && drawn < maxEntries; i-- {
		entry := gs.DiplomaticOfferHistory[i]
		if !diplomacyOfferHistoryMatches(entry, gs.PlayerFactionID, dirFilter, actionFilter) {
			continue
		}
		rect := diplomacyOfferHistoryCardRect(panelRect, drawn)
		if rect.Hit(mx, my) {
			target, ok := diplomacyOfferHistoryOtherFaction(entry, gs.PlayerFactionID)
			if !ok {
				return "", 0, false
			}
			if f := gs.Factions[target]; f == nil || f.IsEliminated {
				return "", 0, false
			}
			actionKind := actionKindForDiplomacyAction(diplomacy.Action(entry.Action))
			if actionKind == ActionNone {
				actionKind = ActionProposePeace
			}
			actionFocus := 0
			for j, da := range diplomActions {
				if da.action == actionKind {
					actionFocus = j
					break
				}
			}
			return target, actionFocus, true
		}
		drawn++
	}
	return "", 0, false
}

func diplomacyOfferHistoryHit(gs *state.GameState, panelRect gameui.Rect, mx, my float64, maxEntries int, dirFilter diplomacyHistoryDirectionFilter, actionFilter ActionKind) bool {
	_, _, ok := diplomacyOfferHistorySelection(gs, panelRect, mx, my, maxEntries, dirFilter, actionFilter)
	return ok
}

func diplomacyOfferHistorySummary(gs *state.GameState, dirFilter diplomacyHistoryDirectionFilter, actionFilter ActionKind) (total, accepted, rejected, applied int) {
	if gs == nil {
		return 0, 0, 0, 0
	}
	for i := range gs.DiplomaticOfferHistory {
		entry := gs.DiplomaticOfferHistory[i]
		if !diplomacyOfferHistoryMatches(entry, gs.PlayerFactionID, dirFilter, actionFilter) {
			continue
		}
		total++
		if entry.Accepted {
			accepted++
		} else {
			rejected++
		}
		if entry.Applied {
			applied++
		}
	}
	return total, accepted, rejected, applied
}

func offerPageRect() rectF {
	w := minF(ScreenWidth-120, 760)
	h := minF(ScreenHeight-180, 600)
	if h < 360 {
		h = 360
	}
	x := (ScreenWidth - w) / 2
	y := (ScreenHeight - h) / 2
	return rectF{x: x, y: y, w: w, h: h}
}

func listRowStartY() float64 {
	return diplomacyListLayoutForScreen().listRect.Y
}

func diplomVisibleRows() int {
	layout := diplomacyListLayoutForScreen()
	usable := layout.listRect.H
	rows := int(usable / diplomRowH)
	if rows < 1 {
		return 1
	}
	return rows
}

func diplomMaxScroll(total int) int {
	max := total - diplomVisibleRows()
	if max < 0 {
		return 0
	}
	return max
}

func clampDiplomScroll(total, scroll int) int {
	if scroll < 0 {
		return 0
	}
	max := diplomMaxScroll(total)
	if scroll > max {
		return max
	}
	return scroll
}

func clampDiplomFocus(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func ensureDiplomFocusVisible(total, focus, scroll int) int {
	scroll = clampDiplomScroll(total, scroll)
	visible := diplomVisibleRows()
	if focus < scroll {
		return focus
	}
	if focus >= scroll+visible {
		return focus - visible + 1
	}
	return scroll
}

func diplomActionRect(i int) (x, y, w, h float32) {
	layout := diplomacyOfferLayoutForScreen()
	btnW := float32(layout.actionsRect.W)
	btnH := float32(42)
	gap := float32(12)
	x = float32(layout.actionsRect.X)
	y = float32(layout.actionsRect.Y + float64(i)*(float64(btnH)+float64(gap)))
	return x, y, btnW, btnH
}

func diplomSendRect() (x, y, w, h float32) {
	r := diplomacyOfferLayoutForScreen().sendRect
	return float32(r.X), float32(r.Y), float32(r.W), float32(r.H)
}

func diplomBackRect() (x, y, w, h float32) {
	r := diplomacyOfferLayoutForScreen().backRect
	return float32(r.X), float32(r.Y), float32(r.W), float32(r.H)
}

type diplomacyActionButton struct {
	Index  int
	Button gameui.Button
}

func buildDiplomacyCloseButton() gameui.Button {
	x, y, w, h := diplomacyCloseRect()
	btn := gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "").WithIcon(gameui.IconClose)
	btn.IconSize = 13
	return btn
}

func buildDiplomacyListView(gs *state.GameState, focusIdx, scroll int) gameui.ListView {
	factions := sortedFactions(gs)
	items := make([]string, 0, len(factions))
	for _, fid := range factions {
		if f := gs.Factions[fid]; f != nil {
			items = append(items, f.NameTR)
		}
	}
	layout := diplomacyListLayoutForScreen()
	list := gameui.NewListView(layout.listRect.X, layout.listRect.Y, layout.listRect.W, layout.listRect.H, diplomRowH, diplomVisibleRows(), items)
	list.Scroll = clampDiplomScroll(len(items), scroll)
	list.Selected = clampDiplomFocus(focusIdx, 0, len(items)-1)
	return list
}

func buildDiplomacyBackButton() gameui.Button {
	x, y, w, h := diplomBackRect()
	return gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "Geri").WithIcon(gameui.IconBack)
}

func diplomacyListClickedIndex(list gameui.ListView, input gameui.InputState) (int, bool) {
	if !input.LeftJustPressed || !list.HitTest(input.MouseX, input.MouseY) || list.RowHeight <= 0 {
		return -1, false
	}
	row := int((input.MouseY - list.Rect.Y) / list.RowHeight)
	idx := list.Scroll + row
	if row < 0 || row >= list.VisibleRows || idx < 0 || idx >= len(list.Items) {
		return -1, false
	}
	return idx, true
}

func buildDiplomacySendButton() gameui.Button {
	x, y, w, h := diplomSendRect()
	return gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "Teklif Gönder").WithIcon(gameui.IconSend)
}

func buildDiplomacyActionButtons() []diplomacyActionButton {
	out := make([]diplomacyActionButton, 0, len(diplomActions))
	for i, da := range diplomActions {
		x, y, w, h := diplomActionRect(i)
		out = append(out, diplomacyActionButton{
			Index:  i,
			Button: gameui.NewButton(float64(x), float64(y), float64(w), float64(h), da.label),
		})
	}
	return out
}

// DrawDiplomacyPanel diplomasi panelini çizer.
func DrawDiplomacyPanel(screen *ebiten.Image, gs *state.GameState, focusIdx, scroll, actionFocus int, target faction.FactionID, browseTarget faction.FactionID, historyDirFilter diplomacyHistoryDirectionFilter, historyActionFilter ActionKind) {
	drawUIOverlay(screen, color.RGBA{8, 6, 4, 220})

	drawUIPanelTitle(screen, gameui.Rect{X: 0, Y: 24, W: ScreenWidth, H: 24}, "── Diplomasi ──")
	drawDiplomacyCloseButton(screen)

	factions := sortedFactions(gs)
	scroll = clampDiplomScroll(len(factions), scroll)
	focusIdx = clampDiplomFocus(focusIdx, 0, len(factions)-1)
	start := scroll
	end := start + diplomVisibleRows()
	if end > len(factions) {
		end = len(factions)
	}

	if target == "" {
		drawDiplomacyListPage(screen, gs, factions, focusIdx, start, end, browseTarget, historyDirFilter, historyActionFilter)
	} else {
		drawDiplomacyOfferPanel(screen, gs, target, actionFocus, browseTarget, historyDirFilter, historyActionFilter)
	}

	if target == "" && len(factions) > end-start {
		info := "Liste: " + itoa(start+1) + "-" + itoa(end) + "/" + itoa(len(factions))
		layout := diplomacyListLayoutForScreen()
		drawUIMutedText(screen, layout.footerRect.X, layout.footerRect.Y, info)
	}
}

func drawDiplomacyListPage(screen *ebiten.Image, gs *state.GameState, factions []faction.FactionID, focusIdx, start, end int, browseTarget faction.FactionID, historyDirFilter diplomacyHistoryDirectionFilter, historyActionFilter ActionKind) {
	layout := diplomacyListLayoutForScreen()
	drawUIPanelFrame(screen, layout.panelRect, color.RGBA{15, 12, 9, 235}, panelBorder, 1.2, 3)
	DrawText(screen, "Diplomatik Hedef", layout.titleRect.X, layout.titleRect.Y, FaceLarge, ColorGold)
	hintText := "Fare tekeri veya ok tuşları ile listeyi kaydırın"
	if browseTarget != "" {
		hintText += " | " + diplomacyHistoryBrowseLabelTR()
	}
	drawUIMutedText(screen, layout.titleRect.X, layout.titleRect.Y+22, hintText)
	drawUICardRect(screen, layout.listRect, color.RGBA{11, 9, 7, 225}, color.RGBA{92, 74, 38, 190}, 1)

	list := buildDiplomacyListView(gs, focusIdx, start)
	for row, i := 0, list.Scroll; i < end; i, row = i+1, row+1 {
		fid := factions[i]
		f := gs.Factions[fid]
		rel := gs.Relations[faction.RelationKey(gs.PlayerFactionID, fid)]

		rowRect := gameui.Rect{
			X: list.Rect.X + 10,
			Y: list.Rect.Y + 8 + float64(row)*diplomRowH,
			W: list.Rect.W - 32,
			H: diplomRowH - 10,
		}
		rowCol := color.RGBA{24, 18, 12, 210}
		borderCol := color.RGBA{78, 62, 34, 150}
		browseActive := browseTarget != "" && fid == browseTarget
		switch {
		case i == focusIdx && browseActive:
			rowCol = color.RGBA{70, 56, 18, 238}
			borderCol = color.RGBA{224, 188, 92, 235}
		case browseActive:
			rowCol = color.RGBA{54, 44, 14, 232}
			borderCol = color.RGBA{210, 172, 72, 228}
		case i == focusIdx:
			rowCol = color.RGBA{64, 50, 22, 235}
			borderCol = color.RGBA{186, 148, 74, 230}
		}
		drawUICardRect(screen, rowRect, rowCol, borderCol, 1)
		if browseActive {
			drawUICardAccent(screen, rowRect, 7, color.RGBA{218, 176, 78, 235})
			badgeRect := gameui.Rect{X: rowRect.X + rowRect.W - 74, Y: rowRect.Y + 6, W: 62, H: 18}
			drawUICardRect(screen, badgeRect, color.RGBA{84, 64, 22, 235}, color.RGBA{236, 206, 120, 235}, 1)
			drawUILabel(screen, badgeRect, "GEÇMİŞ", color.RGBA{255, 238, 194, 255}, gameui.TextSmall, gameui.TextAlignCenter)
		}

		fc := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		drawUICardAccent(screen, rowRect, 6, fc)

		regionCount := len(gs.RegionsOwnedBy(fid))
		leftRow := gameui.NewTableRow(gameui.Rect{X: rowRect.X + 18, Y: rowRect.Y + 7, W: rowRect.W - 248}, []gameui.TableCell{
			{Text: trimTextToWidth(f.NameTR, FaceMed, rowRect.W-248), Color: ColorWhite, Variant: gameui.TextMedium, Align: gameui.TextAlignStart, Weight: 1},
		}, 0)
		drawUITableRow(screen, leftRow)
		subRow := gameui.NewTableRow(gameui.Rect{X: rowRect.X + 18, Y: rowRect.Y + 29, W: rowRect.W - 248}, []gameui.TableCell{
			{Text: itoa(regionCount) + " bölge", Color: ColorGray, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 1},
		}, 0)
		drawUITableRow(screen, subRow)

		statusX := rowRect.X + rowRect.W - 220
		if rel != nil {
			stanceCol, stanceTR := stanceDisplay(rel.Stance)
			scoreCol := scoreColor(rel.Score)
			rightRow := gameui.NewTableRow(gameui.Rect{X: statusX, Y: rowRect.Y + 7, W: 206}, []gameui.TableCell{
				{Text: stanceTR, Color: stanceCol, Variant: gameui.TextMedium, Align: gameui.TextAlignStart, Weight: 1},
			}, 0)
			drawUITableRow(screen, rightRow)
			scoreRow := gameui.NewTableRow(gameui.Rect{X: statusX, Y: rowRect.Y + 29, W: 206}, []gameui.TableCell{
				{Text: "İlişki: " + itoa(rel.Score), Color: scoreCol, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 1},
			}, 0)
			drawUITableRow(screen, scoreRow)
		} else {
			neutralRow := gameui.NewTableRow(gameui.Rect{X: statusX, Y: rowRect.Y + 7, W: 206}, []gameui.TableCell{
				{Text: "Tarafsız", Color: ColorGray, Variant: gameui.TextMedium, Align: gameui.TextAlignStart, Weight: 1},
			}, 0)
			drawUITableRow(screen, neutralRow)
		}
	}
	drawDiplomacyListScrollbar(screen, len(factions), list.Scroll)
	if layout.historyRect.W > 0 {
		drawDiplomacyOfferHistoryPanelRect(screen, gs, layout.historyRect, 4, historyDirFilter, historyActionFilter)
	}
}

func drawDiplomacyOfferPanel(screen *ebiten.Image, gs *state.GameState, target faction.FactionID, actionFocus int, browseTarget faction.FactionID, historyDirFilter diplomacyHistoryDirectionFilter, historyActionFilter ActionKind) {
	f := gs.Factions[target]
	if f == nil {
		return
	}
	layout := diplomacyOfferLayoutForScreen()
	drawUIPanelFrame(screen, layout.panelRect, color.RGBA{14, 11, 8, 235}, panelBorder, 1.2, 3)

	browseActive := browseTarget != "" && browseTarget == target
	drawUILabel(screen, gameui.Rect{X: layout.titleRect.X, Y: layout.titleRect.Y}, "Teklif Paneli", ColorGold, gameui.TextLarge, gameui.TextAlignStart)
	if browseActive {
		drawUIMutedText(screen, layout.titleRect.X, layout.titleRect.Y+22, diplomacyHistoryBrowseLabelTR())
	}
	targetCardRect := gameui.Rect{X: layout.targetRect.X, Y: layout.targetRect.Y - 2, W: layout.targetRect.W, H: layout.targetRect.H + 6}
	targetFill := color.RGBA{22, 18, 12, 220}
	targetBorder := color.RGBA{90, 72, 40, 170}
	if browseActive {
		targetFill = color.RGBA{36, 28, 12, 232}
		targetBorder = color.RGBA{214, 178, 82, 240}
	}
	drawUICardRect(screen, targetCardRect, targetFill, targetBorder, 1)
	if browseActive {
		drawUICardAccent(screen, targetCardRect, 7, color.RGBA{218, 176, 78, 235})
		badgeRect := gameui.Rect{X: targetCardRect.X + targetCardRect.W - 74, Y: targetCardRect.Y + 6, W: 62, H: 18}
		drawUICardRect(screen, badgeRect, color.RGBA{84, 64, 22, 235}, color.RGBA{236, 206, 120, 235}, 1)
		drawUILabel(screen, badgeRect, "GEÇMİŞ", color.RGBA{255, 238, 194, 255}, gameui.TextSmall, gameui.TextAlignCenter)
	}
	targetRow := gameui.NewKeyValueRow(gameui.Rect{X: layout.targetRect.X + 12, Y: layout.targetRect.Y + 5, W: layout.targetRect.W - 24}, "Hedef:", trimTextToWidth(f.NameTR, FaceMed, layout.targetRect.W-88))
	targetRow.LabelColor = ColorGray
	targetRow.ValueColor = ColorWhite
	targetRow.LabelVariant = gameui.TextMedium
	targetRow.ValueVariant = gameui.TextMedium
	targetRow.Gap = 12
	targetRow.ValueAlign = gameui.TextAlignStart
	drawUIKeyValueWidget(screen, targetRow)

	rel := gs.Relations[faction.RelationKey(gs.PlayerFactionID, target)]
	relScore := 0
	relStance := faction.StancePeace
	if rel != nil {
		relScore = rel.Score
		relStance = rel.Stance
	}
	drawUICardRect(screen, layout.statusRect, color.RGBA{19, 16, 12, 220}, color.RGBA{92, 74, 38, 170}, 1)
	drawUIInfoBlock(screen, layout.statusRect.X+12, layout.statusRect.Y+8, []string{
		"Durum: " + stanceDisplayText(relStance),
		"İlişki Skoru: " + itoa(relScore),
	}, []color.Color{
		ColorGray,
		scoreColor(relScore),
	})

	drawUIMutedText(screen, layout.actionsRect.X, layout.actionsRect.Y-18, "Teklif Türü")
	for _, btn := range buildDiplomacyActionButtons() {
		i := btn.Index
		da := diplomActions[i]
		chance, status := estimateDiplomacyChance(gs, target, da.action)
		disabledReason := diplomacyActionDisabledReason(gs, target, da.action)
		bg := da.color
		textCol := ColorWhite
		if i != actionFocus {
			bg.A = 170
		}
		if disabledReason != "" {
			bg.A = 110
			textCol = ColorGray
			status = disabledReason
		}
		drawDiplomacyButton(screen, btn.Button, bg, panelBorder, FaceMed, 7)
		bx, by, bw, _ := diplomActionRect(i)
		chanceText := "%" + itoa(chance)
		drawUILabel(screen, gameui.Rect{X: float64(bx), Y: float64(by) + 7, W: float64(bw - 14)}, chanceText, textCol, gameui.TextMedium, gameui.TextAlignEnd)
		drawUILabel(screen, gameui.Rect{X: float64(bx) + 14, Y: float64(by) + 25, W: float64(bw - 28)}, status, color.RGBA{235, 230, 210, 230}, gameui.TextSmall, gameui.TextAlignStart)
	}

	selected := "Seçili teklif: " + diplomActions[actionFocus].label
	drawUICardRect(screen, layout.selectedRect, color.RGBA{18, 14, 10, 215}, color.RGBA{78, 62, 34, 150}, 1)
	slw := MeasureText(selected, FaceSmall)
	selectedY := layout.selectedRect.Y + 6
	drawUIMutedText(screen, layout.selectedRect.X+layout.selectedRect.W/2-slw/2, selectedY, selected)

	drawDiplomacyOfferHistoryPanelRect(screen, gs, layout.historyRect, 3, historyDirFilter, historyActionFilter)

	drawDiplomacyButton(screen, buildDiplomacyBackButton(), color.RGBA{70, 70, 70, 230}, panelBorder, FaceMed, 10)
	drawDiplomacyButton(screen, buildDiplomacySendButton(), color.RGBA{48, 130, 72, 235}, panelBorder, FaceMed, 10)
}

func diplomacyCloseRect() (x, y, w, h float32) {
	return float32(ScreenWidth) - 58, 20, 30, 26
}

func drawDiplomacyCloseButton(screen *ebiten.Image) {
	drawDiplomacyButton(screen, buildDiplomacyCloseButton(), color.RGBA{45, 34, 25, 230}, panelBorder, FaceSmall, 6)
}

func drawDiplomacyButton(screen *ebiten.Image, btn gameui.Button, bg color.RGBA, border color.Color, face *text.GoTextFace, textOffsetY float64) {
	style := menuButtonStyle
	style.BG = bg
	style.Border = color.RGBAModel.Convert(border).(color.RGBA)
	style.Text = ColorWhite
	style.TextOffsetY = textOffsetY
	if face == FaceMed {
		style.TextVariant = gameui.TextMedium
	} else {
		style.TextVariant = gameui.TextSmall
	}
	drawUIButtonWidget(screen, btn, style)
}

// handleDiplomacyInput diplomasi paneli klavye ve fare girişini işler.
func (r *Renderer) handleDiplomacyInput(input gameui.InputState) InputAction {
	factions := sortedFactions(r.gs)
	n := len(factions)
	if n == 0 {
		return InputAction{}
	}
	r.diplomacyScroll = clampDiplomScroll(n, r.diplomacyScroll)
	r.diplomacyFocus = clampDiplomFocus(r.diplomacyFocus, 0, n-1)
	r.diplomacyActionFocus = clampDiplomFocus(r.diplomacyActionFocus, 0, len(diplomActions)-1)
	if input.LeftJustPressed && !diplomacyPanelPointerHit(input.MouseX, input.MouseY, r.gs, r.diplomacyFocus, r.diplomacyScroll, r.diplomacyTargetFaction, r.diplomacyHistoryDirectionFilter, r.diplomacyHistoryActionFilter) {
		r.showDiplomacy = false
		r.diplomacyTargetFaction = ""
		r.diplomacyOfferHistoryBrowse = ""
		return InputAction{}
	}
	if buildDiplomacyCloseButton().HandleInput(input) {
		r.showDiplomacy = false
		r.diplomacyTargetFaction = ""
		r.diplomacyOfferHistoryBrowse = ""
		return InputAction{}
	}
	if r.diplomacyTargetFaction == "" {
		if input.LeftJustPressed && r.applyDiplomacyHistoryFilterHit(diplomacyListLayoutForScreen().historyRect, input.MouseX, input.MouseY) {
			return InputAction{}
		}
		if input.WheelY != 0 && diplomacyListLayoutForScreen().panelRect.Hit(input.MouseX, input.MouseY) {
			r.diplomacyScroll = clampDiplomScroll(n, r.diplomacyScroll-wheelToDiplomStep(input.WheelY))
			return InputAction{}
		}
		if input.LeftJustPressed {
			if target, actionFocus, ok := diplomacyOfferHistorySelection(r.gs, diplomacyListLayoutForScreen().historyRect, input.MouseX, input.MouseY, 4, r.diplomacyHistoryDirectionFilter, r.diplomacyHistoryActionFilter); ok {
				r.diplomacyTargetFaction = target
				r.diplomacyActionFocus = actionFocus
				for i, fid := range factions {
					if fid == target {
						r.diplomacyFocus = i
						r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
						break
					}
				}
				return InputAction{}
			}
		}
		list := buildDiplomacyListView(r.gs, r.diplomacyFocus, r.diplomacyScroll)
		if idx, ok := diplomacyListClickedIndex(list, input); ok {
			if idx == r.diplomacyFocus && idx < len(factions) {
				r.diplomacyTargetFaction = factions[idx]
				r.diplomacyActionFocus = 0
				return InputAction{}
			}
			r.diplomacyFocus = idx
			r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
			return InputAction{}
		}
		if list.HandleInput(input) {
			r.diplomacyScroll = list.Scroll
			if list.Selected >= 0 {
				r.diplomacyFocus = list.Selected
				r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
			}
			return InputAction{}
		}
	} else {
		if input.LeftJustPressed && r.applyDiplomacyHistoryFilterHit(diplomacyOfferLayoutForScreen().historyRect, input.MouseX, input.MouseY) {
			return InputAction{}
		}
		if buildDiplomacyBackButton().HandleInput(input) {
			r.diplomacyTargetFaction = ""
			return InputAction{}
		}
		if input.LeftJustPressed {
			if target, actionFocus, ok := diplomacyOfferHistorySelection(r.gs, diplomacyOfferLayoutForScreen().historyRect, input.MouseX, input.MouseY, 3, r.diplomacyHistoryDirectionFilter, r.diplomacyHistoryActionFilter); ok {
				r.diplomacyTargetFaction = target
				r.diplomacyActionFocus = actionFocus
				for i, fid := range factions {
					if fid == target {
						r.diplomacyFocus = i
						r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
						break
					}
				}
				return InputAction{}
			}
		}
		for _, btn := range buildDiplomacyActionButtons() {
			if btn.Button.HandleInput(input) {
				r.diplomacyActionFocus = btn.Index
				return InputAction{}
			}
		}
		if buildDiplomacySendButton().HandleInput(input) {
			target := r.diplomacyTargetFaction
			if reason := diplomacyActionDisabledReason(r.gs, target, diplomActions[r.diplomacyActionFocus].action); reason != "" {
				r.ShowCombatResult(reason)
				return InputAction{}
			}
			r.showDiplomacy = false
			r.diplomacyTargetFaction = ""
			return InputAction{Kind: diplomActions[r.diplomacyActionFocus].action, TargetFaction: target}
		}
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) && r.diplomacyFocus < n-1 {
		r.diplomacyFocus++
		r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) && r.diplomacyFocus > 0 {
		r.diplomacyFocus--
		r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
	}
	if r.diplomacyTargetFaction != "" {
		if r.keyJustPressed(ebiten.KeyArrowRight) && r.diplomacyActionFocus < len(diplomActions)-1 {
			r.diplomacyActionFocus++
		}
		if r.keyJustPressed(ebiten.KeyArrowLeft) && r.diplomacyActionFocus > 0 {
			r.diplomacyActionFocus--
		}
	}
	if r.keyJustPressed(ebiten.KeyTab) || r.keyJustPressed(ebiten.KeyEscape) {
		if r.diplomacyTargetFaction != "" {
			r.diplomacyTargetFaction = ""
		} else {
			r.showDiplomacy = false
		}
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		if r.diplomacyTargetFaction == "" {
			if r.diplomacyFocus < len(factions) {
				r.diplomacyTargetFaction = factions[r.diplomacyFocus]
				r.diplomacyActionFocus = 0
				return InputAction{}
			}
		} else {
			target := r.diplomacyTargetFaction
			if reason := diplomacyActionDisabledReason(r.gs, target, diplomActions[r.diplomacyActionFocus].action); reason != "" {
				r.ShowCombatResult(reason)
				return InputAction{}
			}
			r.showDiplomacy = false
			r.diplomacyTargetFaction = ""
			return InputAction{Kind: diplomActions[r.diplomacyActionFocus].action, TargetFaction: target}
		}
	}
	return InputAction{}
}

func wheelToDiplomStep(wheelY float64) int {
	if wheelY > 0 {
		return 1
	}
	if wheelY < 0 {
		return -1
	}
	return 0
}

func drawDiplomacyListScrollbar(screen *ebiten.Image, total, scroll int) {
	visible := diplomVisibleRows()
	if total <= visible {
		return
	}
	layout := diplomacyListLayoutForScreen()
	trackRect := gameui.Rect{
		X: layout.listRect.X + layout.listRect.W - 14,
		Y: layout.listRect.Y + 10,
		W: 6,
		H: layout.listRect.H - 20,
	}
	drawUICardRect(screen, trackRect, color.RGBA{24, 20, 15, 220}, color.RGBA{64, 50, 28, 120}, 1)
	maxScroll := diplomMaxScroll(total)
	thumbH := trackRect.H * float64(visible) / float64(total)
	if thumbH < 28 {
		thumbH = 28
	}
	thumbY := trackRect.Y
	if maxScroll > 0 {
		thumbY += (trackRect.H - thumbH) * float64(scroll) / float64(maxScroll)
	}
	drawUICardRect(screen, gameui.Rect{X: trackRect.X, Y: thumbY, W: trackRect.W, H: thumbH}, color.RGBA{176, 144, 78, 230}, color.RGBA{214, 190, 120, 210}, 1)
}

func diplomacyPanelPointerHit(mx, my float64, gs *state.GameState, focusIdx, scroll int, target faction.FactionID, dirFilter diplomacyHistoryDirectionFilter, actionFilter ActionKind) bool {
	if buildDiplomacyCloseButton().HitTest(mx, my) {
		return true
	}
	if target != "" {
		if buildDiplomacyBackButton().HitTest(mx, my) || buildDiplomacySendButton().HitTest(mx, my) {
			return true
		}
		for _, btn := range buildDiplomacyActionButtons() {
			if btn.Button.HitTest(mx, my) {
				return true
			}
		}
		if _, _, ok := diplomacyHistoryFilterHit(diplomacyOfferLayoutForScreen().historyRect, dirFilter, actionFilter, mx, my); ok {
			return true
		}
		p := diplomacyOfferLayoutForScreen().panelRect
		return p.Hit(mx, my)
	}
	if _, _, ok := diplomacyHistoryFilterHit(diplomacyListLayoutForScreen().historyRect, dirFilter, actionFilter, mx, my); ok {
		return true
	}
	list := buildDiplomacyListView(gs, focusIdx, scroll)
	if list.HitTest(mx, my) {
		return true
	}
	return diplomacyListLayoutForScreen().panelRect.Hit(mx, my)
}

func sortedFactions(gs *state.GameState) []faction.FactionID {
	var fids []faction.FactionID
	for fid := range gs.Factions {
		if fid == gs.PlayerFactionID {
			continue
		}
		if f := gs.Factions[fid]; f == nil || f.IsEliminated {
			continue
		}
		fids = append(fids, fid)
	}
	sort.Slice(fids, func(i, j int) bool { return fids[i] < fids[j] })
	return fids
}

func stanceDisplay(s faction.DiplomaticStance) (color.Color, string) {
	switch s {
	case faction.StanceWar:
		return ColorRed, faction.DiplomaticStanceBadgeTR(s)
	case faction.StanceAllied:
		return color.RGBA{60, 220, 60, 255}, faction.DiplomaticStanceBadgeTR(s)
	case faction.StanceTrade:
		return ColorGold, faction.DiplomaticStanceBadgeTR(s)
	default:
		return ColorGray, faction.DiplomaticStanceLabelTR(faction.NormalizeStance(s))
	}
}

func stanceDisplayText(s faction.DiplomaticStance) string {
	_, label := stanceDisplay(s)
	return label
}

func scoreColor(score int) color.Color {
	if score >= 50 {
		return color.RGBA{60, 220, 60, 255}
	}
	if score >= 0 {
		return ColorGray
	}
	if score >= -50 {
		return color.RGBA{220, 160, 60, 255}
	}
	return ColorRed
}

func estimateDiplomacyChance(gs *state.GameState, target faction.FactionID, action ActionKind) (int, string) {
	rel := gs.Relations[faction.RelationKey(gs.PlayerFactionID, target)]
	score := 0
	stance := faction.StancePeace
	if rel != nil {
		score = rel.Score
		stance = rel.Stance
	}
	playerRegions := len(gs.RegionsOwnedBy(gs.PlayerFactionID))
	targetRegions := len(gs.RegionsOwnedBy(target))
	regionDelta := playerRegions - targetRegions

	chance := 50 + score/2
	switch action {
	case ActionDeclareWar:
		if stance == faction.StanceWar {
			chance = 0
		} else {
			chance = 100
		}
	case ActionProposePeace:
		if stance != faction.StanceWar {
			chance = 0
		} else {
			chance = 35 + (-score / 2) + regionDelta*4
		}
	case ActionProposeAlliance:
		if stance == faction.StanceWar {
			chance = 0
		} else {
			chance = 15 + score + regionDelta*2
		}
	case ActionProposeTrade:
		if stance == faction.StanceWar {
			chance = 0
		} else {
			assessment := diplomacy.AssessTradeProposal(gs, rel, gs.PlayerFactionID, target)
			if assessment.BlockReason != "" {
				return 0, assessment.BlockReason
			}
			chance = assessment.Chance
		}
	}
	if chance < 0 {
		chance = 0
	}
	if chance > 100 {
		chance = 100
	}
	switch {
	case chance == 0:
		return chance, "Geçersiz / Mümkün değil"
	case chance >= 75:
		return chance, "Yüksek kabul olasılığı"
	case chance >= 45:
		return chance, "Orta kabul olasılığı"
	default:
		return chance, "Düşük kabul olasılığı"
	}
}
