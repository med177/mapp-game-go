package render

import (
	"image"
	"image/color"
	"sort"

	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/victory"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	diplomRowH            = 58.0
	diplomNameColumnW     = 340.0
	diplomFactionFlagSize = 40.0
	diplomFactionFlagGap  = 10.0
	diplomColumnGap       = 24.0
	diplomHistoryPanelW   = 286.0
	diplomHistoryPanelGap = 12.0
	diplomOfferMainW      = 430.0
	diplomHistoryPanelH   = 324.0
	diplomActionButtonH   = 42.0
	diplomActionGap       = 8.0
	diplomRelationRowH    = 17.0
	diplomRelationHeaderH = 20.0
	diplomRelationGapH    = 12.0
)

type diplomacyListSort int

const (
	diplomacyListSortAlphabetical diplomacyListSort = iota
	diplomacyListSortRelation
	diplomacyListSortPowerRanking
	diplomacyListSortEconomicRanking
)

type diplomacyListSortButton struct {
	Sort   diplomacyListSort
	Button gameui.Button
}

type diplomacyHistoryDirectionFilter int

const (
	diplomacyHistoryDirectionAll diplomacyHistoryDirectionFilter = iota
	diplomacyHistoryDirectionIncoming
	diplomacyHistoryDirectionOutgoing
)

type diplomacyRelationCategory int

const (
	diplomacyRelationWar diplomacyRelationCategory = iota
	diplomacyRelationAlliance
	diplomacyRelationTrade
)

type diplomacyRelationCategoryMeta struct {
	Category diplomacyRelationCategory
	Label    string
	Color    color.RGBA
}

var diplomacyRelationCategories = [3]diplomacyRelationCategoryMeta{
	{Category: diplomacyRelationWar, Label: "Savaşta", Color: color.RGBA{180, 58, 58, 235}},
	{Category: diplomacyRelationAlliance, Label: "İttifaklar", Color: color.RGBA{58, 164, 82, 235}},
	{Category: diplomacyRelationTrade, Label: "Ticaret Anlaşmaları", Color: color.RGBA{184, 142, 54, 235}},
}

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
	{diplomacy.ActionLabelTR(diplomacy.ActionImproveRelations), color.RGBA{72, 124, 174, 220}, ActionImproveRelations},
	{diplomacy.ActionLabelTR(diplomacy.ActionSendGift), color.RGBA{182, 120, 58, 220}, ActionSendGift},
	{diplomacy.ActionLabelTR(diplomacy.ActionOfferVassalization), color.RGBA{86, 132, 68, 220}, ActionOfferVassalization},
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
	case diplomacy.ActionJoinWarCall:
		return ActionDeclareWar
	case diplomacy.ActionProposePeace:
		return ActionProposePeace
	case diplomacy.ActionProposeAlliance:
		return ActionProposeAlliance
	case diplomacy.ActionProposeTrade:
		return ActionProposeTrade
	case diplomacy.ActionImproveRelations:
		return ActionImproveRelations
	case diplomacy.ActionSendGift:
		return ActionSendGift
	case diplomacy.ActionOfferVassalization:
		return ActionOfferVassalization
	default:
		return ActionNone
	}
}

func diplomacyActionDisabledReason(gs *state.GameState, target faction.FactionID, action ActionKind) string {
	if gs == nil || target == "" {
		return ""
	}
	var actionValue diplomacy.Action
	switch action {
	case ActionDeclareWar:
		actionValue = diplomacy.ActionDeclareWar
	case ActionProposePeace:
		actionValue = diplomacy.ActionProposePeace
	case ActionProposeAlliance:
		actionValue = diplomacy.ActionProposeAlliance
	case ActionProposeTrade:
		actionValue = diplomacy.ActionProposeTrade
	case ActionCancelAlliance:
		actionValue = diplomacy.ActionCancelAlliance
	case ActionCancelTrade:
		actionValue = diplomacy.ActionCancelTrade
	case ActionImproveRelations:
		actionValue = diplomacy.ActionImproveRelations
	case ActionSendGift:
		actionValue = diplomacy.ActionSendGift
	case ActionOfferVassalization:
		actionValue = diplomacy.ActionOfferVassalization
	default:
		return ""
	}
	return diplomacy.ActionBlockReason(gs, gs.PlayerFactionID, target, actionValue)
}

func diplomacyActionPaymentNote(action ActionKind) string {
	switch action {
	case ActionImproveRelations:
		return "Karşı devlete ödeme gitmez"
	case ActionSendGift:
		return "Karşı devletin hazinesine 80 altın gider"
	default:
		return ""
	}
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
	sortRect    gameui.Rect
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

type diplomacyVassalManagementLayout struct {
	panelRect     gameui.Rect
	releaseButton gameui.Button
	annexButton   gameui.Button
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
	sortRect, box := box.CutTop(28, 8)
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
		sortRect:    sortRect,
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
	footerRect, box := box.CutBottom(40, 0)
	headerRect, box := box.CutTop(28, 14)
	// İki satırlı durum metni için alt çerçeveye değmeyecek sabit iç boşluk.
	statusRect, box := box.CutTop(86, 16)
	actionsRect, box := box.CutTop(float64(len(diplomActions))*diplomActionButtonH+float64(len(diplomActions)-1)*diplomActionGap, 16)
	selectedRect, _ := box.CutTop(28, 0)
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

func buildDiplomacyVassalManagementLayout() diplomacyVassalManagementLayout {
	side := diplomacyOfferLayoutForScreen().historyRect
	panel := gameui.Rect{X: side.X, Y: side.Y + side.H + 12, W: side.W, H: 122}
	return diplomacyVassalManagementLayout{
		panelRect:     panel,
		releaseButton: gameui.NewButton(panel.X+12, panel.Y+36, panel.W-24, 32, diplomacy.ActionLabelTR(diplomacy.ActionReleaseVassal)),
		annexButton:   gameui.NewButton(panel.X+12, panel.Y+76, panel.W-24, 32, diplomacy.ActionLabelTR(diplomacy.ActionAnnexVassal)),
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

func buildDiplomacyHistoryFilterButtons(panelRect gameui.Rect, _ diplomacyHistoryDirectionFilter, _ ActionKind) [7]diplomacyHistoryFilterButton {
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
	h := minF(ScreenHeight-80, 680)
	if h < 420 {
		h = 420
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
	btnH := float32(diplomActionButtonH)
	gap := float32(diplomActionGap)
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

func buildDiplomacySideViewButton(panelRect gameui.Rect, historyVisible bool) gameui.Button {
	label := "Geçmiş"
	icon := gameui.IconBook
	if historyVisible {
		label = "İlişkiler"
		icon = gameui.IconBack
	}
	btn := gameui.NewButton(panelRect.X+panelRect.W-100, panelRect.Y+8, 90, 24, label).WithIcon(icon)
	btn.IconSize = 11
	btn.IconGap = 5
	return btn
}

func diplomacyListSortLabelTR(sortMode diplomacyListSort) string {
	switch sortMode {
	case diplomacyListSortRelation:
		return "İlişki"
	case diplomacyListSortPowerRanking:
		return "Güç Sıralaması"
	case diplomacyListSortEconomicRanking:
		return "Ekonomik Sıralama"
	default:
		return "Alfabetik"
	}
}

func buildDiplomacyListSortButtons(layout diplomacyListLayout) [4]diplomacyListSortButton {
	var buttons [4]diplomacyListSortButton
	if layout.sortRect.W <= 0 || layout.sortRect.H <= 0 {
		return buttons
	}
	const gap = 8.0
	buttonW := (layout.sortRect.W - gap*3) / 4
	for i, sortMode := range [...]diplomacyListSort{
		diplomacyListSortAlphabetical,
		diplomacyListSortRelation,
		diplomacyListSortPowerRanking,
		diplomacyListSortEconomicRanking,
	} {
		buttons[i] = diplomacyListSortButton{
			Sort: sortMode,
			Button: gameui.NewButton(
				layout.sortRect.X+float64(i)*(buttonW+gap),
				layout.sortRect.Y,
				buttonW,
				layout.sortRect.H,
				diplomacyListSortLabelTR(sortMode),
			),
		}
	}
	return buttons
}

func diplomacyListSortHit(layout diplomacyListLayout, mx, my float64) (diplomacyListSort, bool) {
	for _, btn := range buildDiplomacyListSortButtons(layout) {
		if btn.Button.HitTest(mx, my) {
			return btn.Sort, true
		}
	}
	return diplomacyListSortAlphabetical, false
}

func drawDiplomacyListSortButton(screen *ebiten.Image, btn diplomacyListSortButton, active bool) {
	style := solidButtonStyle(
		color.RGBA{28, 22, 15, 225},
		color.RGBA{88, 72, 40, 180},
		color.RGBA{214, 206, 190, 255},
		0,
	)
	style.TextVariant = gameui.TextSmall
	if active {
		style.BG = color.RGBA{112, 82, 32, 235}
		style.Border = color.RGBA{232, 194, 104, 255}
		style.Text = ColorWhite
	}
	drawUIButtonWidget(screen, btn.Button, style)
}

func buildDiplomacyListView(gs *state.GameState, focusIdx, scroll int) gameui.ListView {
	return buildDiplomacyListViewForSort(gs, focusIdx, scroll, diplomacyListSortAlphabetical)
}

func buildDiplomacyListViewForSort(gs *state.GameState, focusIdx, scroll int, sortMode diplomacyListSort) gameui.ListView {
	factions := sortedDiplomacyFactions(gs, sortMode)
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

func diplomacyListColumnRects(rowRect gameui.Rect) (gameui.Rect, gameui.Rect) {
	content := gameui.Rect{
		X: rowRect.X + 18,
		Y: rowRect.Y + 7,
		W: rowRect.W - 36,
		H: 22,
	}
	nameW := diplomNameColumnW
	// Yeni hazine kolonu için minimum metrik alanını da ayır. Dar listelerde
	// devlet adı kısalabilir; metrik kolonları birbirinin üzerine binmemeli.
	maxNameW := content.W - diplomColumnGap - 380
	if nameW > maxNameW {
		nameW = maxNameW
	}
	if nameW < 0 {
		nameW = 0
	}
	nameRect := content
	nameRect.X += diplomFactionFlagSize + diplomFactionFlagGap
	nameRect.W = nameW - diplomFactionFlagSize - diplomFactionFlagGap
	if nameRect.W < 0 {
		nameRect.W = 0
	}
	relationRect := gameui.Rect{
		X: content.X + nameW + diplomColumnGap,
		Y: content.Y,
		W: content.W - nameW - diplomColumnGap,
		H: content.H,
	}
	return nameRect, relationRect
}

func diplomacyListMetricColumnRects(rowRect gameui.Rect) (nameRect, relationRect, powerRect, rankRect, treasuryRect gameui.Rect) {
	var metricsRect gameui.Rect
	nameRect, metricsRect = diplomacyListColumnRects(rowRect)
	const metricGap = 10.0
	powerW := metricsRect.W * 0.20
	if powerW > 110 {
		powerW = 110
	}
	if powerW < 78 {
		powerW = 78
	}
	rankW := metricsRect.W * 0.20
	if rankW > 105 {
		rankW = 105
	}
	if rankW < 78 {
		rankW = 78
	}
	treasuryW := metricsRect.W * 0.24
	if treasuryW > 125 {
		treasuryW = 125
	}
	if treasuryW < 94 {
		treasuryW = 94
	}
	relationW := metricsRect.W - powerW - rankW - treasuryW - metricGap*3
	if relationW < 0 {
		relationW = 0
	}
	relationRect = metricsRect
	relationRect.W = relationW
	powerRect = metricsRect
	powerRect.X = relationRect.X + relationRect.W + metricGap
	powerRect.W = powerW
	rankRect = powerRect
	rankRect.X = powerRect.X + powerRect.W + metricGap
	rankRect.W = rankW
	treasuryRect = rankRect
	treasuryRect.X = rankRect.X + rankRect.W + metricGap
	treasuryRect.W = treasuryW
	return nameRect, relationRect, powerRect, rankRect, treasuryRect
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

func diplomacyDoubleClickTarget(gs *state.GameState, target faction.FactionID) faction.FactionID {
	if gs == nil || target == "" {
		return target
	}
	if overlord := diplomacy.DirectOverlord(gs, target); overlord != "" && overlord != gs.PlayerFactionID {
		return overlord
	}
	return target
}

func buildDiplomacySendButton() gameui.Button {
	return buildDiplomacySendButtonForAction(ActionNone)
}

func buildDiplomacySendButtonForAction(action ActionKind) gameui.Button {
	x, y, w, h := diplomSendRect()
	label := "Teklif Gönder"
	if action == ActionCancelAlliance || action == ActionCancelTrade {
		label = "Anlaşmayı Bitir"
	}
	return gameui.NewButton(float64(x), float64(y), float64(w), float64(h), label).WithIcon(gameui.IconSend)
}

func diplomacyActionForTarget(gs *state.GameState, target faction.FactionID, index int) ActionKind {
	if index < 0 || index >= len(diplomActions) {
		return ActionNone
	}
	base := diplomActions[index].action
	if gs == nil || target == "" || diplomacy.SameRealm(gs, gs.PlayerFactionID, target) {
		return base
	}
	switch base {
	case ActionProposeAlliance:
		if rel := diplomacy.Relation(gs, gs.PlayerFactionID, target); rel != nil && rel.Stance == faction.StanceAllied {
			return ActionCancelAlliance
		}
	case ActionProposeTrade:
		if diplomacy.HasTradeRouteBetween(gs, gs.PlayerFactionID, target) {
			return ActionCancelTrade
		}
	}
	return base
}

func diplomacyActionLabel(gs *state.GameState, target faction.FactionID, index int) string {
	switch diplomacyActionForTarget(gs, target, index) {
	case ActionCancelAlliance:
		return diplomacy.ActionLabelTR(diplomacy.ActionCancelAlliance)
	case ActionCancelTrade:
		return diplomacy.ActionLabelTR(diplomacy.ActionCancelTrade)
	default:
		if index >= 0 && index < len(diplomActions) {
			return diplomActions[index].label
		}
		return ""
	}
}

func buildDiplomacyActionButtons(gs *state.GameState, target faction.FactionID) []diplomacyActionButton {
	out := make([]diplomacyActionButton, 0, len(diplomActions))
	for i := range diplomActions {
		x, y, w, h := diplomActionRect(i)
		out = append(out, diplomacyActionButton{
			Index:  i,
			Button: gameui.NewButton(float64(x), float64(y), float64(w), float64(h), diplomacyActionLabel(gs, target, i)),
		})
	}
	return out
}

func diplomacyActionEnabled(gs *state.GameState, target faction.FactionID, index int) bool {
	action := diplomacyActionForTarget(gs, target, index)
	return action != ActionNone && diplomacyActionDisabledReason(gs, target, action) == ""
}

func enabledDiplomacyActionFocus(gs *state.GameState, target faction.FactionID, preferred int) int {
	if diplomacyActionEnabled(gs, target, preferred) {
		return preferred
	}
	for i := range diplomActions {
		if diplomacyActionEnabled(gs, target, i) {
			return i
		}
	}
	return clampDiplomFocus(preferred, 0, len(diplomActions)-1)
}

func nextEnabledDiplomacyAction(gs *state.GameState, target faction.FactionID, current, direction int) int {
	for i := current + direction; i >= 0 && i < len(diplomActions); i += direction {
		if diplomacyActionEnabled(gs, target, i) {
			return i
		}
	}
	return current
}

// DrawDiplomacyPanel diplomasi panelini çizer.
func DrawDiplomacyPanel(screen *ebiten.Image, gs *state.GameState, focusIdx, scroll, actionFocus int, target faction.FactionID, browseTarget faction.FactionID, historyVisible bool, historyDirFilter diplomacyHistoryDirectionFilter, historyActionFilter ActionKind) {
	drawDiplomacyPanelWithSortAndRelationScroll(screen, gs, focusIdx, scroll, actionFocus, target, browseTarget, historyVisible, historyDirFilter, historyActionFilter, diplomacyListSortAlphabetical, 0)
}

func DrawDiplomacyPanelWithSort(screen *ebiten.Image, gs *state.GameState, focusIdx, scroll, actionFocus int, target faction.FactionID, browseTarget faction.FactionID, historyVisible bool, historyDirFilter diplomacyHistoryDirectionFilter, historyActionFilter ActionKind, sortMode diplomacyListSort) {
	drawDiplomacyPanelWithSortAndRelationScroll(screen, gs, focusIdx, scroll, actionFocus, target, browseTarget, historyVisible, historyDirFilter, historyActionFilter, sortMode, 0)
}

func drawDiplomacyPanelWithSortAndRelationScroll(screen *ebiten.Image, gs *state.GameState, focusIdx, scroll, actionFocus int, target faction.FactionID, browseTarget faction.FactionID, historyVisible bool, historyDirFilter diplomacyHistoryDirectionFilter, historyActionFilter ActionKind, sortMode diplomacyListSort, relationScroll int) {
	drawUIOverlay(screen, color.RGBA{8, 6, 4, 220})

	drawUIPanelTitle(screen, gameui.Rect{X: 0, Y: 24, W: ScreenWidth, H: 24}, "── Diplomasi ──")
	drawDiplomacyCloseButton(screen)

	factions := sortedDiplomacyFactions(gs, sortMode)
	scroll = clampDiplomScroll(len(factions), scroll)
	focusIdx = clampDiplomFocus(focusIdx, 0, len(factions)-1)
	start := scroll
	end := start + diplomVisibleRows()
	if end > len(factions) {
		end = len(factions)
	}

	if target == "" {
		drawDiplomacyListPage(screen, gs, factions, sortMode, focusIdx, start, end, browseTarget, historyVisible, historyDirFilter, historyActionFilter, relationScroll)
	} else {
		drawDiplomacyOfferPanel(screen, gs, factions, target, actionFocus, browseTarget, historyVisible, historyDirFilter, historyActionFilter, relationScroll)
	}

	if target == "" && len(factions) > end-start {
		info := "Liste: " + itoa(start+1) + "-" + itoa(end) + "/" + itoa(len(factions))
		layout := diplomacyListLayoutForScreen()
		drawUIMutedText(screen, layout.footerRect.X, layout.footerRect.Y, info)
	}
}

func drawDiplomacyListPage(screen *ebiten.Image, gs *state.GameState, factions []faction.FactionID, sortMode diplomacyListSort, focusIdx, start, end int, browseTarget faction.FactionID, historyVisible bool, historyDirFilter diplomacyHistoryDirectionFilter, historyActionFilter ActionKind, relationScroll int) {
	layout := diplomacyListLayoutForScreen()
	drawUIPanelFrame(screen, layout.panelRect, color.RGBA{15, 12, 9, 235}, panelBorder, 1.2, 3)
	DrawText(screen, "Diplomatik Hedef", layout.titleRect.X, layout.titleRect.Y, FaceLarge, ColorGold)
	hintText := "Fare tekeri veya ok tuşları ile listeyi kaydırın"
	if browseTarget != "" {
		hintText += " | " + diplomacyHistoryBrowseLabelTR()
	}
	drawUIMutedText(screen, layout.titleRect.X, layout.titleRect.Y+22, hintText)
	for _, btn := range buildDiplomacyListSortButtons(layout) {
		drawDiplomacyListSortButton(screen, btn, btn.Sort == sortMode)
	}
	drawUICardRect(screen, layout.listRect, color.RGBA{11, 9, 7, 225}, color.RGBA{92, 74, 38, 190}, 1)

	list := buildDiplomacyListViewForSort(gs, focusIdx, start, sortMode)
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
		nameInitial := ""
		if f.NameTR != "" {
			nameInitial = string([]rune(f.NameTR)[:1])
		}
		drawFactionFlagBadge(screen, fid, nameInitial, rowRect.X+18, rowRect.Y+4, diplomFactionFlagSize, fc, panelBorder)

		regionCount := len(gs.LandRegionsOwnedBy(fid))
		nameRect, relationRect, powerRect, rankRect, treasuryRect := diplomacyListMetricColumnRects(rowRect)
		leftRow := gameui.NewTableRow(nameRect, []gameui.TableCell{
			{Text: trimTextToWidth(f.NameTR, FaceMed, nameRect.W), Color: ColorWhite, Variant: gameui.TextMedium, Align: gameui.TextAlignStart, Weight: 1},
		}, 0)
		drawUITableRow(screen, leftRow)
		subRect := nameRect
		subRect.Y = rowRect.Y + 29
		subRow := gameui.NewTableRow(subRect, []gameui.TableCell{
			{Text: itoa(regionCount) + " bölge", Color: ColorGray, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 1},
		}, 0)
		drawUITableRow(screen, subRow)

		if fid == gs.PlayerFactionID {
			selfRow := gameui.NewTableRow(relationRect, []gameui.TableCell{
				{Text: "Oyuncu", Color: ColorGold, Variant: gameui.TextMedium, Align: gameui.TextAlignStart, Weight: 1},
			}, 0)
			drawUITableRow(screen, selfRow)
			selfScoreRect := relationRect
			selfScoreRect.Y = rowRect.Y + 29
			selfScoreRow := gameui.NewTableRow(selfScoreRect, []gameui.TableCell{
				{Text: "Kendi devletin", Color: ColorGray, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 1},
			}, 0)
			drawUITableRow(screen, selfScoreRow)
		} else if rel != nil || diplomacy.DirectOverlord(gs, fid) != "" || diplomacy.DirectOverlord(gs, gs.PlayerFactionID) == fid {
			stanceCol, stanceTR := diplomacyStatusDisplay(gs, gs.PlayerFactionID, fid, rel)
			scoreValue := 0
			if rel != nil {
				scoreValue = rel.Score
			}
			scoreCol := scoreColor(scoreValue)
			rightRow := gameui.NewTableRow(relationRect, []gameui.TableCell{
				{Text: trimTextToWidth(stanceTR, FaceMed, relationRect.W), Color: stanceCol, Variant: gameui.TextMedium, Align: gameui.TextAlignStart, Weight: 1},
			}, 0)
			drawUITableRow(screen, rightRow)
			scoreRect := relationRect
			scoreRect.Y = rowRect.Y + 29
			scoreRow := gameui.NewTableRow(scoreRect, []gameui.TableCell{
				{Text: "İlişki: " + itoa(scoreValue), Color: scoreCol, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 1},
			}, 0)
			drawUITableRow(screen, scoreRow)
		} else {
			neutralRow := gameui.NewTableRow(relationRect, []gameui.TableCell{
				{Text: "Tarafsız", Color: ColorGray, Variant: gameui.TextMedium, Align: gameui.TextAlignStart, Weight: 1},
			}, 0)
			drawUITableRow(screen, neutralRow)
		}

		_, militaryRank, factionCount := factionMilitaryPowerStanding(gs, fid)
		drawUILabel(screen, powerRect, "Askeri güç", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		powerValueRect := powerRect
		powerValueRect.Y = rowRect.Y + 27
		drawUILabel(screen, powerValueRect, factionMilitaryPowerBreakdownLabel(gs, fid), ColorWhite, gameui.TextMedium, gameui.TextAlignStart)
		drawUILabel(screen, rankRect, "Güç sırası", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		rankValueRect := rankRect
		rankValueRect.Y = rowRect.Y + 27
		rankText := "-"
		if factionCount > 0 {
			rankText = itoa(militaryRank)
		}
		drawUILabel(screen, rankValueRect, rankText, ColorWhite, gameui.TextMedium, gameui.TextAlignStart)
		drawUILabel(screen, treasuryRect, "Hazine", ColorGold, gameui.TextSmall, gameui.TextAlignStart)
		treasuryValueRect := treasuryRect
		treasuryValueRect.Y = rowRect.Y + 27
		drawUILabel(screen, treasuryValueRect, factionTreasuryLabel(gs, fid), ColorWhite, gameui.TextMedium, gameui.TextAlignStart)
	}
	drawDiplomacyListScrollbar(screen, len(factions), list.Scroll)
	if layout.historyRect.W > 0 {
		selected := faction.FactionID("")
		if focusIdx >= 0 && focusIdx < len(factions) {
			selected = factions[focusIdx]
		}
		drawDiplomacySidePanel(screen, gs, layout.historyRect, selected, factions, historyVisible, 4, historyDirFilter, historyActionFilter, relationScroll)
	}
}

func drawDiplomacyOfferPanel(screen *ebiten.Image, gs *state.GameState, factions []faction.FactionID, target faction.FactionID, actionFocus int, browseTarget faction.FactionID, historyVisible bool, historyDirFilter diplomacyHistoryDirectionFilter, historyActionFilter ActionKind, relationScroll int) {
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
	if rel != nil {
		relScore = rel.Score
	}
	statusColor, statusLabel := diplomacyStatusDisplay(gs, gs.PlayerFactionID, target, rel)
	drawUICardRect(screen, layout.statusRect, color.RGBA{19, 16, 12, 220}, color.RGBA{92, 74, 38, 170}, 1)
	drawUIInfoBlock(screen, layout.statusRect.X+12, layout.statusRect.Y+8, []string{
		"Durum: " + statusLabel,
		"İlişki Skoru: " + itoa(relScore),
	}, []color.Color{
		statusColor,
		scoreColor(relScore),
	})

	selectedDisabledReason := ""
	for _, btn := range buildDiplomacyActionButtons(gs, target) {
		i := btn.Index
		da := diplomActions[i]
		action := diplomacyActionForTarget(gs, target, i)
		chance, status := estimateDiplomacyChance(gs, target, action)
		disabledReason := diplomacyActionDisabledReason(gs, target, action)
		bg := da.color
		if action == ActionCancelAlliance || action == ActionCancelTrade {
			bg = color.RGBA{156, 72, 48, 225}
		}
		textCol := ColorWhite
		if i != actionFocus {
			bg.A = 170
		}
		if disabledReason != "" {
			bg.A = 110
			textCol = ColorGray
			status = disabledReason
			if i == actionFocus {
				selectedDisabledReason = disabledReason
			}
		}
		drawDiplomacyActionButton(screen, btn.Button, bg, disabledReason != "", i == actionFocus)
		bx, by, bw, _ := diplomActionRect(i)
		chanceText := "%" + itoa(chance)
		if action == ActionCancelAlliance || action == ActionCancelTrade {
			chanceText = "AKTİF"
		}
		if disabledReason != "" {
			chanceText = "PASİF"
		}
		drawUILabel(screen, gameui.Rect{X: float64(bx), Y: float64(by) + 7, W: float64(bw - 14)}, chanceText, textCol, gameui.TextMedium, gameui.TextAlignEnd)
		// Aktif tekliflerde kabul olasılığı, pasif tekliflerde engel nedeni gösterilir.
		// Ayrıntı satırını alt kenardan en az 7 px içeride tut.
		detailX := float64(bx) + 14
		detailW := float64(bw - 28)
		paymentNote := diplomacyActionPaymentNote(action)
		if paymentNote != "" {
			paymentW := MeasureText(paymentNote, FaceSmall)
			statusW := detailW - paymentW - 12
			if statusW < 0 {
				statusW = 0
			}
			drawUILabel(screen, gameui.Rect{X: detailX, Y: float64(by) + 23, W: statusW}, trimTextToWidth(status, FaceSmall, statusW), color.RGBA{235, 230, 210, 230}, gameui.TextSmall, gameui.TextAlignStart)
			drawUILabel(screen, gameui.Rect{X: detailX, Y: float64(by) + 23, W: detailW}, paymentNote, color.RGBA{210, 205, 190, 220}, gameui.TextSmall, gameui.TextAlignEnd)
		} else {
			drawUILabel(screen, gameui.Rect{X: detailX, Y: float64(by) + 23, W: detailW}, status, color.RGBA{235, 230, 210, 230}, gameui.TextSmall, gameui.TextAlignStart)
		}
	}

	selectedAction := diplomacyActionForTarget(gs, target, actionFocus)
	drawUICardRect(screen, layout.selectedRect, color.RGBA{18, 14, 10, 215}, color.RGBA{78, 62, 34, 150}, 1)
	drawDiplomacySidePanel(screen, gs, layout.historyRect, target, factions, historyVisible, 3, historyDirFilter, historyActionFilter, relationScroll)
	if diplomacy.DirectOverlord(gs, target) == gs.PlayerFactionID {
		drawDiplomacyVassalManagementPanel(screen, gs, target)
	}

	drawDiplomacyButton(screen, buildDiplomacyBackButton(), color.RGBA{70, 70, 70, 230}, panelBorder, FaceMed, 10)
	sendColor := color.RGBA{48, 130, 72, 235}
	if selectedDisabledReason != "" {
		sendColor = color.RGBA{58, 58, 54, 150}
	}
	drawDiplomacyButton(screen, buildDiplomacySendButtonForAction(selectedAction), sendColor, panelBorder, FaceMed, 10)
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

func diplomacyActionButtonStyle(bg color.RGBA, disabled, selected bool) gameui.ButtonStyle {
	style := menuButtonStyle
	style.BG = bg
	style.Border = color.RGBA{108, 88, 52, 185}
	style.Text = ColorWhite
	style.TextOffsetY = 7
	style.TextVariant = gameui.TextMedium
	if selected && !disabled {
		style.Border = color.RGBA{242, 198, 82, 255}
		style.BorderWidth = 2
	}
	if disabled {
		style.Border = color.RGBA{72, 68, 60, 145}
		style.Text = color.RGBA{142, 138, 130, 230}
	}
	return style
}

func drawDiplomacyActionButton(screen *ebiten.Image, btn gameui.Button, bg color.RGBA, disabled, selected bool) {
	style := diplomacyActionButtonStyle(bg, disabled, selected)
	drawUIButtonWidget(screen, btn, style)
}

func drawDiplomacyVassalManagementPanel(screen *ebiten.Image, gs *state.GameState, target faction.FactionID) {
	layout := buildDiplomacyVassalManagementLayout()
	drawUIPanelFrame(screen, layout.panelRect, color.RGBA{18, 14, 10, 228}, color.RGBA{104, 82, 42, 190}, 1, 3)
	drawUILabel(screen, gameui.Rect{X: layout.panelRect.X + 12, Y: layout.panelRect.Y + 10, W: layout.panelRect.W - 24}, "Vassal Yönetimi", ColorGold, gameui.TextMedium, gameui.TextAlignStart)
	drawDiplomacyButton(screen, layout.releaseButton, color.RGBA{62, 104, 142, 225}, panelBorder, FaceSmall, 6)
	annexDisabled := diplomacy.ActionBlockReason(gs, gs.PlayerFactionID, target, diplomacy.ActionAnnexVassal) != ""
	drawDiplomacyActionButton(screen, layout.annexButton, color.RGBA{154, 54, 48, 230}, annexDisabled, false)
}

func diplomacyRelationCategoryMatches(gs *state.GameState, subject, other faction.FactionID, category diplomacyRelationCategory) bool {
	if gs == nil || subject == "" || other == "" || subject == other {
		return false
	}
	switch category {
	case diplomacyRelationTrade:
		return diplomacy.HasTradeRouteBetween(gs, subject, other)
	case diplomacyRelationWar:
		rel := diplomacy.Relation(gs, subject, other)
		return rel != nil && rel.Stance == faction.StanceWar
	case diplomacyRelationAlliance:
		if diplomacy.SameRealm(gs, subject, other) {
			return false
		}
		rel := diplomacy.Relation(gs, subject, other)
		return rel != nil && rel.Stance == faction.StanceAllied
	default:
		return false
	}
}

func diplomacyRelationCategoryCount(gs *state.GameState, subject faction.FactionID, factions []faction.FactionID, category diplomacyRelationCategory) int {
	if gs == nil {
		return 0
	}
	count := 0
	playerIncluded := false
	for _, other := range factions {
		if other == gs.PlayerFactionID {
			playerIncluded = true
		}
		if diplomacyRelationCategoryMatches(gs, subject, other, category) {
			count++
		}
	}
	if subject != gs.PlayerFactionID && !playerIncluded && diplomacyRelationCategoryMatches(gs, subject, gs.PlayerFactionID, category) {
		count++
	}
	return count
}

func diplomacyRelationsPanelViewport(panelRect gameui.Rect) gameui.Rect {
	return gameui.Rect{
		X: panelRect.X + 10,
		Y: panelRect.Y + 68,
		W: panelRect.W - 20,
		H: panelRect.H - 78,
	}
}

func diplomacyRelationSectionRows(count int) int {
	if count < 1 {
		count = 1 // "Yok" satırı
	}
	sectionH := diplomRelationHeaderH + float64(count)*diplomRelationRowH + diplomRelationGapH
	return int((sectionH + diplomRelationRowH - 1) / diplomRelationRowH)
}

func diplomacyRelationsContentRows(gs *state.GameState, subject faction.FactionID, factions []faction.FactionID) int {
	rows := 0
	for _, meta := range diplomacyRelationCategories {
		rows += diplomacyRelationSectionRows(diplomacyRelationCategoryCount(gs, subject, factions, meta.Category))
	}
	return rows
}

func diplomacyRelationsVisibleRows(viewport gameui.Rect) int {
	rows := int(viewport.H / diplomRelationRowH)
	if rows < 1 {
		return 1
	}
	return rows
}

func diplomacyRelationsMaxScroll(totalRows int, viewport gameui.Rect) int {
	max := totalRows - diplomacyRelationsVisibleRows(viewport)
	if max < 0 {
		return 0
	}
	return max
}

func clampDiplomacyRelationsScroll(totalRows int, viewport gameui.Rect, scroll int) int {
	if scroll < 0 {
		return 0
	}
	max := diplomacyRelationsMaxScroll(totalRows, viewport)
	if scroll > max {
		return max
	}
	return scroll
}

func drawDiplomacySidePanel(screen *ebiten.Image, gs *state.GameState, panelRect gameui.Rect, subject faction.FactionID, factions []faction.FactionID, historyVisible bool, maxHistoryEntries int, dirFilter diplomacyHistoryDirectionFilter, actionFilter ActionKind, relationScroll int) {
	if historyVisible {
		drawDiplomacyOfferHistoryPanelRect(screen, gs, panelRect, maxHistoryEntries, dirFilter, actionFilter)
	} else {
		drawDiplomacyRelationsPanelRect(screen, gs, panelRect, subject, factions, relationScroll)
	}
	btn := buildDiplomacySideViewButton(panelRect, historyVisible)
	drawDiplomacyButton(screen, btn, color.RGBA{74, 58, 28, 235}, color.RGBA{176, 142, 72, 220}, FaceSmall, 5)
}

func drawDiplomacyRelationsPanelRect(screen *ebiten.Image, gs *state.GameState, panelRect gameui.Rect, subject faction.FactionID, factions []faction.FactionID, relationScroll int) {
	if gs == nil || panelRect.W <= 0 || panelRect.H <= 0 {
		return
	}
	drawUIPanelFrame(screen, panelRect, color.RGBA{18, 14, 10, 228}, color.RGBA{88, 72, 40, 180}, 1, 3)
	drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: panelRect.Y + 10, W: panelRect.W - 124}, "Aktif İlişkiler", color.RGBA{255, 220, 100, 255}, gameui.TextMedium, gameui.TextAlignStart)
	f := gs.Factions[subject]
	if f == nil {
		drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: panelRect.Y + 50, W: panelRect.W - 28}, "Bir devlet seçin.", ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
		return
	}
	drawUILabel(screen, gameui.Rect{X: panelRect.X + 14, Y: panelRect.Y + 32, W: panelRect.W - 28}, trimTextToWidth(f.NameTR, FaceSmall, panelRect.W-28), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
	hierarchy := "Bağımsız devlet"
	if overlord := diplomacy.DirectOverlord(gs, subject); overlord != "" {
		hierarchy = factionDisplayName(gs, string(overlord)) + " himayesinde"
	} else if vassalCount := directVassalCount(gs, subject); vassalCount > 0 {
		hierarchy = itoa(vassalCount) + " bağlı devlet"
	}
	drawUIMutedText(screen, panelRect.X+14, panelRect.Y+48, hierarchy)

	viewport := diplomacyRelationsPanelViewport(panelRect)
	totalRows := diplomacyRelationsContentRows(gs, subject, factions)
	relationScroll = clampDiplomacyRelationsScroll(totalRows, viewport, relationScroll)
	left, top := int(viewport.X), int(viewport.Y)
	right, bottom := int(viewport.X+viewport.W), int(viewport.Y+viewport.H)
	if right <= left || bottom <= top {
		return
	}
	body := screen.SubImage(image.Rect(left, top, right, bottom)).(*ebiten.Image)
	contentRow := -relationScroll
	for _, meta := range diplomacyRelationCategories {
		count := diplomacyRelationCategoryCount(gs, subject, factions, meta.Category)
		sectionY := viewport.Y + float64(contentRow)*diplomRelationRowH
		drawUILabel(body, gameui.Rect{X: viewport.X + 4, Y: sectionY, W: viewport.W - 12}, meta.Label+" ("+itoa(count)+")", meta.Color, gameui.TextSmall, gameui.TextAlignStart)
		contentRow++
		if count == 0 {
			drawUIMutedText(body, viewport.X+14, viewport.Y+float64(contentRow)*diplomRelationRowH, "Yok")
			contentRow += 2
			continue
		}
		playerIncluded := false
		for _, other := range factions {
			if other == gs.PlayerFactionID {
				playerIncluded = true
			}
			if !diplomacyRelationCategoryMatches(gs, subject, other, meta.Category) {
				continue
			}
			drawDiplomacyRelationNameAt(body, gs, viewport, viewport.Y+float64(contentRow)*diplomRelationRowH, other, meta.Color)
			contentRow++
		}
		if subject != gs.PlayerFactionID && !playerIncluded && diplomacyRelationCategoryMatches(gs, subject, gs.PlayerFactionID, meta.Category) {
			drawDiplomacyRelationNameAt(body, gs, viewport, viewport.Y+float64(contentRow)*diplomRelationRowH, gs.PlayerFactionID, meta.Color)
			contentRow++
		}
		contentRow++ // bölüm aralığı
	}
	drawDiplomacyRelationsScrollbar(screen, viewport, totalRows, relationScroll)
}

func directVassalCount(gs *state.GameState, overlord faction.FactionID) int {
	if gs == nil || overlord == "" {
		return 0
	}
	count := 0
	for fid, f := range gs.Factions {
		if f != nil && !f.IsEliminated && fid != overlord && f.OverlordID == overlord {
			count++
		}
	}
	return count
}

func drawDiplomacyRelationName(screen *ebiten.Image, gs *state.GameState, panelRect gameui.Rect, sectionY float64, row int, fid faction.FactionID, accent color.RGBA) {
	drawDiplomacyRelationNameAt(screen, gs, gameui.Rect{X: panelRect.X + 10, Y: panelRect.Y, W: panelRect.W - 20, H: panelRect.H}, sectionY+20+float64(row)*diplomRelationRowH, fid, accent)
}

func drawDiplomacyRelationNameAt(screen *ebiten.Image, gs *state.GameState, viewport gameui.Rect, y float64, fid faction.FactionID, accent color.RGBA) {
	name := factionDisplayName(gs, string(fid))
	if name == "" {
		name = string(fid)
	}
	drawUICardRect(screen, gameui.Rect{X: viewport.X + 8, Y: y + 4, W: 5, H: 5}, accent, accent, 1)
	textW := viewport.W - 34
	drawUILabel(screen, gameui.Rect{X: viewport.X + 20, Y: y, W: textW}, trimTextToWidth(name, FaceSmall, textW), color.RGBA{226, 220, 208, 255}, gameui.TextSmall, gameui.TextAlignStart)
}

func drawDiplomacyRelationsScrollbar(screen *ebiten.Image, viewport gameui.Rect, totalRows, scroll int) {
	maxScroll := diplomacyRelationsMaxScroll(totalRows, viewport)
	if maxScroll <= 0 {
		return
	}
	track := gameui.Rect{X: viewport.X + viewport.W - 6, Y: viewport.Y + 2, W: 4, H: viewport.H - 4}
	thumbH := track.H * float64(diplomacyRelationsVisibleRows(viewport)) / float64(totalRows)
	if thumbH < 18 {
		thumbH = 18
	}
	scroll = clampDiplomacyRelationsScroll(totalRows, viewport, scroll)
	thumbY := track.Y + (track.H-thumbH)*float64(scroll)/float64(maxScroll)
	drawRoundedRect(screen, float32(track.X), float32(track.Y), float32(track.W), float32(track.H), 2, color.RGBA{70, 65, 55, 180})
	drawRoundedRect(screen, float32(track.X), float32(thumbY), float32(track.W), float32(thumbH), 2, color.RGBA{210, 175, 85, 230})
}

// handleDiplomacyInput diplomasi paneli klavye ve fare girişini işler.
func (r *Renderer) handleDiplomacyInput(input gameui.InputState) InputAction {
	factions := sortedDiplomacyFactions(r.gs, r.diplomacyListSort)
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
		r.diplomacyHistoryVisible = false
		return InputAction{}
	}
	if buildDiplomacyCloseButton().HandleInput(input) {
		r.showDiplomacy = false
		r.diplomacyTargetFaction = ""
		r.diplomacyOfferHistoryBrowse = ""
		r.diplomacyHistoryVisible = false
		return InputAction{}
	}
	if r.diplomacyTargetFaction == "" {
		layout := diplomacyListLayoutForScreen()
		if input.LeftJustPressed {
			if sortMode, ok := diplomacyListSortHit(layout, input.MouseX, input.MouseY); ok {
				r.diplomacyListSort = sortMode
				r.diplomacyFocus = 0
				r.diplomacyScroll = 0
				r.diplomacyRelationScroll = 0
				return InputAction{}
			}
		}
		sideRect := diplomacyListLayoutForScreen().historyRect
		if sideRect.W > 0 && buildDiplomacySideViewButton(sideRect, r.diplomacyHistoryVisible).HandleInput(input) {
			r.diplomacyHistoryVisible = !r.diplomacyHistoryVisible
			return InputAction{}
		}
		if r.diplomacyHistoryVisible && input.LeftJustPressed && r.applyDiplomacyHistoryFilterHit(sideRect, input.MouseX, input.MouseY) {
			return InputAction{}
		}
		if input.WheelY != 0 && !r.diplomacyHistoryVisible {
			viewport := diplomacyRelationsPanelViewport(sideRect)
			if viewport.Hit(input.MouseX, input.MouseY) {
				relationSubject := faction.FactionID("")
				if r.diplomacyFocus >= 0 && r.diplomacyFocus < len(factions) {
					relationSubject = factions[r.diplomacyFocus]
				}
				r.diplomacyRelationScroll = clampDiplomacyRelationsScroll(
					diplomacyRelationsContentRows(r.gs, relationSubject, factions),
					viewport,
					r.diplomacyRelationScroll-wheelToDiplomStep(input.WheelY),
				)
				return InputAction{}
			}
		}
		if input.WheelY != 0 && diplomacyListLayoutForScreen().panelRect.Hit(input.MouseX, input.MouseY) {
			r.diplomacyScroll = clampDiplomScroll(n, r.diplomacyScroll-wheelToDiplomStep(input.WheelY))
			return InputAction{}
		}
		if r.diplomacyHistoryVisible && input.LeftJustPressed {
			if target, actionFocus, ok := diplomacyOfferHistorySelection(r.gs, sideRect, input.MouseX, input.MouseY, 4, r.diplomacyHistoryDirectionFilter, r.diplomacyHistoryActionFilter); ok {
				r.diplomacyTargetFaction = target
				r.diplomacyRelationScroll = 0
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
		list := buildDiplomacyListViewForSort(r.gs, r.diplomacyFocus, r.diplomacyScroll, r.diplomacyListSort)
		if idx, ok := diplomacyListClickedIndex(list, input); ok {
			if idx == r.diplomacyFocus && idx < len(factions) {
				// Oyuncunun kendi satırı listede görünür; ancak teklif hedefi
				// olamayacağı için çift tıklama teklif sayfası açmaz.
				if factions[idx] != r.gs.PlayerFactionID {
					r.openDiplomacyTarget(diplomacyDoubleClickTarget(r.gs, factions[idx]), 0)
				}
				return InputAction{}
			}
			r.diplomacyFocus = idx
			r.diplomacyRelationScroll = 0
			r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
			return InputAction{}
		}
		if list.HandleInput(input) {
			r.diplomacyScroll = list.Scroll
			if list.Selected >= 0 {
				r.diplomacyFocus = list.Selected
				r.diplomacyRelationScroll = 0
				r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
			}
			return InputAction{}
		}
	} else {
		sideRect := diplomacyOfferLayoutForScreen().historyRect
		if sideRect.W > 0 && buildDiplomacySideViewButton(sideRect, r.diplomacyHistoryVisible).HandleInput(input) {
			r.diplomacyHistoryVisible = !r.diplomacyHistoryVisible
			return InputAction{}
		}
		if r.diplomacyHistoryVisible && input.LeftJustPressed && r.applyDiplomacyHistoryFilterHit(sideRect, input.MouseX, input.MouseY) {
			return InputAction{}
		}
		if input.WheelY != 0 && !r.diplomacyHistoryVisible {
			viewport := diplomacyRelationsPanelViewport(sideRect)
			if viewport.Hit(input.MouseX, input.MouseY) {
				r.diplomacyRelationScroll = clampDiplomacyRelationsScroll(
					diplomacyRelationsContentRows(r.gs, r.diplomacyTargetFaction, factions),
					viewport,
					r.diplomacyRelationScroll-wheelToDiplomStep(input.WheelY),
				)
				return InputAction{}
			}
		}
		if buildDiplomacyBackButton().HandleInput(input) {
			r.diplomacyTargetFaction = ""
			r.diplomacyRelationScroll = 0
			return InputAction{}
		}
		if r.diplomacyHistoryVisible && input.LeftJustPressed {
			if target, actionFocus, ok := diplomacyOfferHistorySelection(r.gs, sideRect, input.MouseX, input.MouseY, 3, r.diplomacyHistoryDirectionFilter, r.diplomacyHistoryActionFilter); ok {
				r.diplomacyTargetFaction = target
				r.diplomacyRelationScroll = 0
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
		if diplomacy.DirectOverlord(r.gs, r.diplomacyTargetFaction) == r.gs.PlayerFactionID {
			management := buildDiplomacyVassalManagementLayout()
			target := r.diplomacyTargetFaction
			name := factionDisplayName(r.gs, string(target))
			if management.releaseButton.HandleInput(input) {
				r.ShowConfirmDialog(
					"Vasallığı Bitir",
					name+" bağımsız bir devlet olacak. Mevcut ticaret anlaşması devam edecek.",
					"Serbest Bırak",
					"İptal",
					InputAction{Kind: ActionReleaseVassal, TargetFaction: target},
					nil,
				)
				return InputAction{}
			}
			if management.annexButton.HandleInput(input) {
				if reason := diplomacy.ActionBlockReason(r.gs, r.gs.PlayerFactionID, target, diplomacy.ActionAnnexVassal); reason != "" {
					r.ShowCombatResult(reason)
					return InputAction{}
				}
				r.ShowConfirmDialog(
					"Vassalı İlhak Et",
					name+" devleti tamamen ilhak edilecek; bölgeleri, kuvvetleri ve kaynakları doğrudan yönetimine geçecek.",
					"İlhak Et",
					"İptal",
					InputAction{Kind: ActionAnnexVassal, TargetFaction: target},
					nil,
				)
				return InputAction{}
			}
		}
		for _, btn := range buildDiplomacyActionButtons(r.gs, r.diplomacyTargetFaction) {
			if btn.Button.HandleInput(input) {
				if !diplomacyActionEnabled(r.gs, r.diplomacyTargetFaction, btn.Index) {
					return InputAction{}
				}
				r.diplomacyActionFocus = btn.Index
				return InputAction{}
			}
		}
		if buildDiplomacySendButton().HandleInput(input) {
			target := r.diplomacyTargetFaction
			action := diplomacyActionForTarget(r.gs, target, r.diplomacyActionFocus)
			if reason := diplomacyActionDisabledReason(r.gs, target, action); reason != "" {
				r.ShowCombatResult(reason)
				return InputAction{}
			}
			r.showDiplomacy = false
			r.diplomacyTargetFaction = ""
			if action == ActionDeclareWar {
				r.openWarConfirm(target, factionDisplayName(r.gs, string(target)), "", "", "", false, ActionNone, combat.BattleContextLand)
				return InputAction{}
			}
			return InputAction{Kind: action, TargetFaction: target}
		}
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) && r.diplomacyFocus < n-1 {
		r.diplomacyFocus++
		r.diplomacyRelationScroll = 0
		r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) && r.diplomacyFocus > 0 {
		r.diplomacyFocus--
		r.diplomacyRelationScroll = 0
		r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
	}
	if r.diplomacyTargetFaction != "" {
		if r.keyJustPressed(ebiten.KeyArrowRight) {
			r.diplomacyActionFocus = nextEnabledDiplomacyAction(r.gs, r.diplomacyTargetFaction, r.diplomacyActionFocus, 1)
		}
		if r.keyJustPressed(ebiten.KeyArrowLeft) {
			r.diplomacyActionFocus = nextEnabledDiplomacyAction(r.gs, r.diplomacyTargetFaction, r.diplomacyActionFocus, -1)
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
				r.diplomacyActionFocus = enabledDiplomacyActionFocus(r.gs, factions[r.diplomacyFocus], 0)
				r.diplomacyHistoryVisible = false
				return InputAction{}
			}
		} else {
			target := r.diplomacyTargetFaction
			action := diplomacyActionForTarget(r.gs, target, r.diplomacyActionFocus)
			if reason := diplomacyActionDisabledReason(r.gs, target, action); reason != "" {
				r.ShowCombatResult(reason)
				return InputAction{}
			}
			r.showDiplomacy = false
			r.diplomacyTargetFaction = ""
			if action == ActionDeclareWar {
				r.openWarConfirm(target, factionDisplayName(r.gs, string(target)), "", "", "", false, ActionNone, combat.BattleContextLand)
				return InputAction{}
			}
			return InputAction{Kind: action, TargetFaction: target}
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
		for _, btn := range buildDiplomacyActionButtons(gs, target) {
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

func factionTreasuryLabel(gs *state.GameState, fid faction.FactionID) string {
	if gs == nil || fid == "" {
		return "0 / 0"
	}
	income := victory.GoldIncomeForFaction(gs, fid)
	gold := 0
	if f := gs.Factions[fid]; f != nil {
		gold = f.Gold
	}
	return itoa(income) + " / " + itoa(gold)
}

func sortedFactions(gs *state.GameState) []faction.FactionID {
	return sortedDiplomacyFactions(gs, diplomacyListSortAlphabetical)
}

func sortedDiplomacyFactions(gs *state.GameState, sortMode diplomacyListSort) []faction.FactionID {
	if gs == nil {
		return nil
	}
	var fids []faction.FactionID
	for fid := range gs.Factions {
		if f := gs.Factions[fid]; f == nil || f.IsEliminated {
			continue
		}
		fids = append(fids, fid)
	}
	relationScores := make(map[faction.FactionID]int)
	adjacentToPlayer := make(map[faction.FactionID]bool)
	economicIncome := make(map[faction.FactionID]int)
	economicGold := make(map[faction.FactionID]int)
	if sortMode == diplomacyListSortRelation {
		for _, fid := range fids {
			if rel := diplomacy.Relation(gs, gs.PlayerFactionID, fid); rel != nil {
				relationScores[fid] = rel.Score
			}
			adjacentToPlayer[fid] = factionsShareLandBorder(gs, gs.PlayerFactionID, fid)
		}
	}
	if sortMode == diplomacyListSortEconomicRanking {
		for _, fid := range fids {
			economicIncome[fid] = victory.GoldIncomeForFaction(gs, fid)
			if f := gs.Factions[fid]; f != nil {
				economicGold[fid] = f.Gold
			}
		}
	}
	sort.Slice(fids, func(i, j int) bool {
		leftID, rightID := fids[i], fids[j]
		switch sortMode {
		case diplomacyListSortRelation:
			leftRelation := relationScores[leftID]
			rightRelation := relationScores[rightID]
			if leftRelation != rightRelation {
				return leftRelation > rightRelation
			}
			if adjacentToPlayer[leftID] != adjacentToPlayer[rightID] {
				return adjacentToPlayer[leftID]
			}
		case diplomacyListSortPowerRanking:
			_, leftRank, _ := factionMilitaryPowerStanding(gs, leftID)
			_, rightRank, _ := factionMilitaryPowerStanding(gs, rightID)
			if leftRank != rightRank {
				return leftRank < rightRank
			}
		case diplomacyListSortEconomicRanking:
			if economicIncome[leftID] != economicIncome[rightID] {
				return economicIncome[leftID] > economicIncome[rightID]
			}
			if economicGold[leftID] != economicGold[rightID] {
				return economicGold[leftID] > economicGold[rightID]
			}
		}

		return leftID < rightID
	})
	return fids
}

func factionsShareLandBorder(gs *state.GameState, left, right faction.FactionID) bool {
	if gs == nil || left == "" || right == "" || left == right {
		return false
	}
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(left) {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor != nil && !neighbor.IsSea && neighbor.OwnerID == string(right) {
				return true
			}
		}
	}
	return false
}

func (r *Renderer) openDiplomacyTarget(target faction.FactionID, actionFocus int) {
	if r == nil || r.gs == nil || target == "" || target == r.gs.PlayerFactionID {
		return
	}
	r.showDiplomacy = true
	r.diplomacyTargetFaction = target
	r.diplomacyRelationScroll = 0
	r.diplomacyActionFocus = enabledDiplomacyActionFocus(r.gs, target, actionFocus)
	r.diplomacyHistoryVisible = false
	factions := sortedDiplomacyFactions(r.gs, r.diplomacyListSort)
	for i, fid := range factions {
		if fid != target {
			continue
		}
		r.diplomacyFocus = i
		r.diplomacyScroll = ensureDiplomFocusVisible(len(factions), r.diplomacyFocus, r.diplomacyScroll)
		return
	}
}

func (r *Renderer) CloseDiplomacyPanel() {
	if r == nil {
		return
	}
	r.showDiplomacy = false
	r.diplomacyTargetFaction = ""
	r.diplomacyRelationScroll = 0
	r.diplomacyOfferHistoryBrowse = ""
	r.diplomacyHistoryVisible = false
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

func diplomacyStatusDisplay(gs *state.GameState, actor, target faction.FactionID, rel *faction.Relation) (color.Color, string) {
	if gs != nil {
		if diplomacy.DirectOverlord(gs, actor) == target {
			return color.RGBA{210, 188, 92, 255}, "LORD Bağlı Olduğun Devlet"
		}
		if diplomacy.DirectOverlord(gs, target) == actor {
			return color.RGBA{86, 188, 94, 255}, "VASSAL Bağlı Devlet"
		}
		if overlord := diplomacy.DirectOverlord(gs, target); overlord != "" {
			return color.RGBA{168, 154, 104, 255}, "VASSAL"
		}
	}
	stance := faction.StancePeace
	if rel != nil {
		stance = rel.Stance
	}
	return stanceDisplay(stance)
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
	var actionValue diplomacy.Action
	switch action {
	case ActionDeclareWar:
		actionValue = diplomacy.ActionDeclareWar
	case ActionProposePeace:
		actionValue = diplomacy.ActionProposePeace
	case ActionProposeAlliance:
		actionValue = diplomacy.ActionProposeAlliance
	case ActionProposeTrade:
		actionValue = diplomacy.ActionProposeTrade
	case ActionCancelAlliance:
		actionValue = diplomacy.ActionCancelAlliance
	case ActionCancelTrade:
		actionValue = diplomacy.ActionCancelTrade
	case ActionImproveRelations:
		actionValue = diplomacy.ActionImproveRelations
	case ActionSendGift:
		actionValue = diplomacy.ActionSendGift
	case ActionOfferVassalization:
		actionValue = diplomacy.ActionOfferVassalization
	}
	if reason := diplomacy.ActionBlockReason(gs, gs.PlayerFactionID, target, actionValue); reason != "" {
		return 0, reason
	}
	rel := gs.Relations[faction.RelationKey(gs.PlayerFactionID, target)]
	score := 0
	stance := faction.StancePeace
	if rel != nil {
		score = rel.Score
		stance = rel.Stance
	}
	chance := 50 + score/2
	switch action {
	case ActionDeclareWar:
		if stance == faction.StanceWar {
			chance = 0
		} else {
			chance = 100
		}
	case ActionProposePeace:
		assessment := diplomacy.AssessPeaceProposal(gs, gs.PlayerFactionID, target)
		chance = assessment.Chance
	case ActionProposeAlliance:
		assessment := diplomacy.AssessAllianceProposal(gs, rel, gs.PlayerFactionID, target)
		chance = assessment.Chance
	case ActionProposeTrade:
		assessment := diplomacy.AssessTradeProposal(gs, rel, gs.PlayerFactionID, target)
		chance = assessment.Chance
	case ActionImproveRelations:
		return 100, "İlişki +8 / 40 altın"
	case ActionSendGift:
		return 100, "İlişki +15 / 120 altın"
	case ActionOfferVassalization:
		chance = diplomacy.AssessVassalizationProposal(gs, rel, gs.PlayerFactionID, target).Chance
	case ActionCancelAlliance:
		return 100, "İttifak sona erecek / Ticaret korunur"
	case ActionCancelTrade:
		return 100, "Ticaret sona erecek / İttifak korunur"
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
	case chance == 100:
		return chance, "Kesin kabul"
	case chance >= 75:
		return chance, "Yüksek kabul olasılığı"
	case chance >= 45:
		return chance, "Orta kabul olasılığı"
	default:
		return chance, "Düşük kabul olasılığı"
	}
}
