package render

import (
	"fmt"
	"image/color"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	merchantRoutePanelW           = float32(660)
	merchantRoutePanelHeaderH     = float32(66)
	merchantRoutePanelRowH        = float32(58)
	merchantRoutePanelFooterH     = float32(42)
	merchantRoutePanelVisibleRows = 7
	merchantRouteFooterButtonH    = float32(20)
	merchantRouteFooterButtonGap  = float32(4)
)

type merchantRoutePanelLayout struct {
	panelX, panelY, panelW, panelH float32
	rowX, rowY, rowW               float32
	close                          gameui.Button
}

func merchantRoutePanelLayoutFor(rowCount int) merchantRoutePanelLayout {
	visibleRows := rowCount
	if visibleRows < 1 {
		visibleRows = 1
	}
	if visibleRows > merchantRoutePanelVisibleRows {
		visibleRows = merchantRoutePanelVisibleRows
	}
	panelH := merchantRoutePanelHeaderH + float32(visibleRows)*merchantRoutePanelRowH + merchantRoutePanelFooterH
	panelX := float32(ScreenWidth)/2 - merchantRoutePanelW/2
	panelY := float32(ScreenHeight)/2 - panelH/2
	close := gameui.NewButton(float64(panelX+merchantRoutePanelW-42), float64(panelY+10), 28, 28, "").WithIcon(gameui.IconClose)
	close.IconSize = 13
	return merchantRoutePanelLayout{
		panelX: panelX, panelY: panelY, panelW: merchantRoutePanelW, panelH: panelH,
		rowX: panelX + 18, rowY: panelY + merchantRoutePanelHeaderH,
		rowW: merchantRoutePanelW - 36, close: close,
	}
}

func merchantRoutePanelRowCount(r *Renderer) int {
	if r == nil {
		return 1
	}
	return len(r.merchantRouteOptions) + 1 // ilk satır görevden çıkarma
}

func merchantRoutePanelRect(r *Renderer) gameui.Rect {
	layout := merchantRoutePanelLayoutFor(merchantRoutePanelRowCount(r))
	return gameui.Rect{X: float64(layout.panelX), Y: float64(layout.panelY), W: float64(layout.panelW), H: float64(layout.panelH)}
}

func merchantRoutePanelRowRect(layout merchantRoutePanelLayout, visibleRow int) gameui.Rect {
	y := layout.rowY + float32(visibleRow)*merchantRoutePanelRowH
	return gameui.Rect{
		X: float64(layout.rowX), Y: float64(y + 5),
		W: float64(layout.rowW), H: float64(merchantRoutePanelRowH - 10),
	}
}

func merchantRoutePanelRowButton(rect gameui.Rect) gameui.Button {
	return gameui.NewButton(rect.X, rect.Y, rect.W, rect.H, "")
}

// merchantRoutePanelInteractiveHit, cursor ve input tarafında aynı görünür
// satır/kapatma düğmesi geometry'sini kullanır. Satır araları ve panel içindeki
// salt bilgi alanları etkileşimli kabul edilmez.
func merchantRoutePanelInteractiveHit(r *Renderer, fx, fy float64) bool {
	if r == nil || !r.showMerchantRoutePanel {
		return false
	}
	layout := merchantRoutePanelLayoutFor(merchantRoutePanelRowCount(r))
	if layout.close.HitTest(fx, fy) {
		return true
	}
	if !merchantRoutePanelRect(r).Hit(fx, fy) {
		return false
	}

	rowCount := merchantRoutePanelRowCount(r)
	maxScroll := rowCount - merchantRoutePanelVisibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	start := r.merchantRouteScroll
	if start < 0 {
		start = 0
	}
	if start > maxScroll {
		start = maxScroll
	}
	visibleRow := int((fy - float64(layout.rowY+5)) / float64(merchantRoutePanelRowH))
	if visibleRow < 0 || visibleRow >= merchantRoutePanelVisibleRows {
		return false
	}
	row := visibleRow + start
	return row >= 0 && row < rowCount && merchantRoutePanelRowRect(layout, visibleRow).Hit(fx, fy)
}

func merchantRouteAssignmentButtonRect(layout armyPanelLayout) gameui.Rect {
	footerY := layout.panelY + layout.panelH - siegeFooterH
	return gameui.Rect{
		X: float64(layout.gridX + 8), Y: float64(footerY + merchantRouteFooterButtonGap), W: 132, H: float64(merchantRouteFooterButtonH),
	}
}

func merchantRouteButtonHit(fx, fy float64, gs *state.GameState, aid army.ArmyID) bool {
	if gs == nil || aid == "" {
		return false
	}
	a := gs.Armies[aid]
	if a == nil || a.OwnerID != string(gs.PlayerFactionID) || !a.IsNaval || !armyHasMerchantShip(gs, a) {
		return false
	}
	layout := armyPanelGeometry()
	return merchantRouteAssignmentButtonRect(layout).Hit(fx, fy)
}

func armyHasMerchantShip(gs *state.GameState, a *army.Army) bool {
	return merchantShipCount(gs, a) > 0
}

func merchantShipCount(gs *state.GameState, a *army.Army) int {
	if gs == nil || a == nil || !a.IsNaval {
		return 0
	}
	count := 0
	for _, unit := range a.Units {
		unitType := gs.UnitTypes[unit.TypeID]
		if unit.TypeID == "merchant_ship" || unitType != nil && unitType.Category == army.CategoryNavalTrade {
			count++
		}
	}
	return count
}

func merchantRouteDisplayName(gs *state.GameState, route *economy.TradeRoute) string {
	if route == nil {
		return "Rota atanmadı"
	}
	return factionDisplayName(gs, route.FromFactionID) + " → " + factionDisplayName(gs, route.ToFactionID)
}

func merchantRouteSeaDisplayName(gs *state.GameState, route *economy.TradeRoute) string {
	if gs == nil || route == nil {
		return "Bilinmiyor"
	}
	seaID, ok := gs.MerchantTradeRouteTargetSeaRegion(route)
	if !ok {
		return "Bilinmiyor"
	}
	return regionDisplayName(gs, string(seaID))
}

func merchantRouteForKey(gs *state.GameState, key string) *economy.TradeRoute {
	if gs == nil || key == "" {
		return nil
	}
	for _, route := range gs.TradeRoutes {
		if route != nil && route.AssignmentKey() == key {
			return route
		}
	}
	return nil
}

func drawMerchantRouteFooter(screen *ebiten.Image, gs *state.GameState, a *army.Army, layout armyPanelLayout) {
	if !armyHasMerchantShip(gs, a) || a.OwnerID != string(gs.PlayerFactionID) {
		return
	}
	rect := merchantRouteAssignmentButtonRect(layout)
	button := gameui.NewButton(rect.X, rect.Y, rect.W, rect.H, "ROTA ATA")
	gameui.DrawButton(screen, button, gameui.ButtonStyle{
		BG: color.RGBA{50, 35, 12, 220}, Border: color.RGBA{160, 120, 40, 220},
		Text: ColorGold, BorderWidth: 1, TextVariant: gameui.TextSmall,
	}, sharedTextRenderer{})

	footer := armyPanelFooterLayoutFor(gs, a, layout)
	if footer.routeStatus.W > 0 {
		label, _, _ := merchantRouteFooterInfo(gs, a)
		drawArmyPanelFooterCenteredText(screen, footer.routeStatus, label, color.RGBA{205, 185, 140, 230})
	}
}

func merchantRouteFooterInfo(gs *state.GameState, a *army.Army) (label, bonusText string, bonusColor color.Color) {
	label = "Rota yok"
	bonusColor = color.RGBA{180, 180, 180, 220}
	if gs == nil || a == nil {
		return label, "", bonusColor
	}
	route := merchantRouteForKey(gs, a.TradeRouteKey)
	if route != nil {
		label = "Aktif: " + merchantRouteDisplayName(gs, route)
	}
	shipCount := merchantShipCount(gs, a)
	bonusText = "Gemi bonusu: +0 hacim/tur · " + itoa(shipCount) + " gemi · rota yok"
	if route == nil {
		return label, bonusText, bonusColor
	}
	bonus := gs.MerchantFleetTradeRouteBonus(a, route)
	bonusText = "Gemi bonusu: +" + itoa(bonus) + " hacim/tur · " + itoa(shipCount) + " gemi"
	if gs.MerchantFleetSupportsTradeRoute(a, route) {
		bonusText += " · konum uygun"
		bonusColor = color.RGBA{145, 220, 155, 240}
	} else {
		bonusText += " · hedef denize git"
		bonusColor = color.RGBA{230, 180, 105, 240}
	}
	return label, bonusText, bonusColor
}

func (r *Renderer) openMerchantRoutePanel() {
	if r == nil || r.gs == nil || r.SelectedArmy == "" {
		return
	}
	fleet := r.gs.Armies[r.SelectedArmy]
	if fleet == nil || fleet.OwnerID != string(r.gs.PlayerFactionID) || !armyHasMerchantShip(r.gs, fleet) {
		return
	}
	r.merchantRouteArmy = fleet.ID
	r.merchantRouteOptions = r.gs.MerchantTradeRoutesForFleet(fleet)
	r.merchantRouteSeaLabels = make([]string, len(r.merchantRouteOptions))
	for i, route := range r.merchantRouteOptions {
		r.merchantRouteSeaLabels[i] = merchantRouteSeaDisplayName(r.gs, route)
	}
	r.merchantRouteScroll = 0
	r.showMerchantRoutePanel = true
	r.bringOverlayPanelToFront(overlayPanelMerchantRoute)
}

func (r *Renderer) closeMerchantRoutePanel() {
	if r == nil {
		return
	}
	r.showMerchantRoutePanel = false
	r.merchantRouteArmy = ""
	r.merchantRouteOptions = r.merchantRouteOptions[:0]
	r.merchantRouteSeaLabels = r.merchantRouteSeaLabels[:0]
	r.merchantRouteScroll = 0
}

func (r *Renderer) focusMerchantRouteTarget(route *economy.TradeRoute) {
	if r == nil || r.gs == nil || route == nil {
		return
	}
	seaID, ok := r.gs.MerchantTradeRouteTargetSeaRegion(route)
	if !ok {
		return
	}
	r.merchantRouteHighlight = seaID
	r.CenterCameraOnRegion(seaID)
	r.SelectedArmy = ""
	r.clearArmySplitSelection()
}

func (r *Renderer) drawMerchantRoutePanel(screen *ebiten.Image) {
	if r == nil || !r.showMerchantRoutePanel {
		return
	}
	layout := merchantRoutePanelLayoutFor(merchantRoutePanelRowCount(r))
	// Modal açıkken alttaki ordu panelinin butonları ve metinleri okunabilir
	// kalmamalı; input zaten bu katmanda tüketiliyor.
	drawUIOverlay(screen, color.RGBA{0, 0, 0, 205})
	drawUIPanelFrame(screen, gameui.Rect{
		X: float64(layout.panelX), Y: float64(layout.panelY),
		W: float64(layout.panelW), H: float64(layout.panelH),
	}, panelBg, panelBorder, 1.5, 3)

	drawUILabel(screen, gameui.Rect{X: float64(layout.panelX + 20), Y: float64(layout.panelY + 22), W: float64(layout.panelW - 80)}, "Merchant Ticaret Rotası", ColorGold, gameui.TextLarge, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: float64(layout.panelX + 20), Y: float64(layout.panelY + 47), W: float64(layout.panelW - 80)}, "Seçili filonun görev rotasını belirleyin.", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawUIButtonWidget(screen, layout.close, tinyButtonStyle)

	rowCount := merchantRoutePanelRowCount(r)
	maxScroll := rowCount - merchantRoutePanelVisibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	start := r.merchantRouteScroll
	if start > maxScroll {
		start = maxScroll
	}
	end := start + merchantRoutePanelVisibleRows
	if end > rowCount {
		end = rowCount
	}
	for row := start; row < end; row++ {
		rowRect := merchantRoutePanelRowRect(layout, row-start)
		bg := color.RGBA{28, 22, 14, 220}
		if row%2 == 0 {
			bg = color.RGBA{38, 29, 16, 230}
		}
		rowButton := merchantRoutePanelRowButton(rowRect)
		gameui.DrawButton(screen, rowButton, gameui.ButtonStyle{
			BG: bg, Border: color.RGBA{92, 70, 34, 180}, Text: ColorWhite, BorderWidth: 1,
		}, sharedTextRenderer{})

		if row == 0 {
			drawUILabel(screen, gameui.Rect{X: rowRect.X + 14, Y: rowRect.Y + 8, W: 170}, "Görev yok", color.RGBA{215, 180, 120, 255}, gameui.TextMedium, gameui.TextAlignStart)
			drawUILabel(screen, gameui.Rect{X: rowRect.X + 180, Y: rowRect.Y + 11, W: rowRect.W - 194}, "Bu filonun merchant rotasını kaldır", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
			continue
		}
		route := r.merchantRouteOptions[row-1]
		if route == nil {
			continue
		}
		amount := route.AmountPerTurn
		current := (*economy.TradeRoute)(nil)
		if fleet := r.gs.Armies[r.merchantRouteArmy]; fleet != nil {
			current = merchantRouteForKey(r.gs, fleet.TradeRouteKey)
		}
		active := current != nil && current.AssignmentKey() == route.AssignmentKey()
		textColor := ColorWhite
		if active {
			textColor = color.RGBA{170, 235, 170, 255}
		}
		seaX := rowRect.X + rowRect.W*0.56
		seaRight := rowRect.X + rowRect.W - 14
		if active {
			seaRight -= 70
		}
		seaW := seaRight - seaX
		nameW := seaX - rowRect.X - 28
		name := trimTextToWidth(merchantRouteDisplayName(r.gs, route), FaceMed, nameW)
		detail := trimTextToWidth(fmt.Sprintf("%s · %d/tur · %d altın/birim", economy.GoodNameTR(route.Good), amount, route.GoldPerUnit), FaceSmall, rowRect.W-28)
		drawUILabel(screen, gameui.Rect{X: rowRect.X + 14, Y: rowRect.Y + 7, W: nameW}, name, textColor, gameui.TextMedium, gameui.TextAlignStart)
		seaLabel := "Bilinmiyor"
		if routeIndex := row - 1; routeIndex >= 0 && routeIndex < len(r.merchantRouteSeaLabels) {
			seaLabel = r.merchantRouteSeaLabels[routeIndex]
		}
		seaLabel = trimTextToWidth(seaLabel, FaceSmall, seaW)
		drawUILabel(screen, gameui.Rect{X: seaX, Y: rowRect.Y + 10, W: seaW}, seaLabel, color.RGBA{100, 185, 255, 255}, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: rowRect.X + 14, Y: rowRect.Y + 27, W: rowRect.W - 28}, detail, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		if active {
			drawUILabel(screen, gameui.Rect{X: rowRect.X + rowRect.W - 62, Y: rowRect.Y + 12, W: 48}, "AKTİF", color.RGBA{170, 235, 170, 255}, gameui.TextSmall, gameui.TextAlignEnd)
		}
	}

	footerY := layout.panelY + layout.panelH - merchantRoutePanelFooterH
	if rowCount > merchantRoutePanelVisibleRows {
		drawUILabel(screen, gameui.Rect{X: float64(layout.panelX + 18), Y: float64(footerY + 15), W: 180}, fmt.Sprintf("Rotalar: %d-%d/%d", start+1, end, rowCount), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	}
	drawUILabel(screen, gameui.Rect{X: float64(layout.panelX + layout.panelW - 120), Y: float64(footerY + 15), W: 102}, "ESC: kapat", ColorGray, gameui.TextSmall, gameui.TextAlignEnd)
}

func (r *Renderer) handleMerchantRoutePanelInput() InputAction {
	if r == nil || !r.showMerchantRoutePanel {
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.closeMerchantRoutePanel()
		return InputAction{}
	}
	layout := merchantRoutePanelLayoutFor(merchantRoutePanelRowCount(r))
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		r.merchantRouteScroll -= int(wheelY)
		maxScroll := merchantRoutePanelRowCount(r) - merchantRoutePanelVisibleRows
		if maxScroll < 0 {
			maxScroll = 0
		}
		if r.merchantRouteScroll < 0 {
			r.merchantRouteScroll = 0
		}
		if r.merchantRouteScroll > maxScroll {
			r.merchantRouteScroll = maxScroll
		}
	}
	if !r.mouseJustPressed(ebiten.MouseButtonLeft) {
		return InputAction{}
	}
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	if layout.close.HitTest(fx, fy) || !merchantRoutePanelRect(r).Hit(fx, fy) {
		r.closeMerchantRoutePanel()
		return InputAction{}
	}
	row := int((fy - float64(layout.rowY+5)) / float64(merchantRoutePanelRowH))
	if row < 0 || row >= merchantRoutePanelVisibleRows || !merchantRoutePanelRowRect(layout, row).Hit(fx, fy) {
		return InputAction{}
	}
	row += r.merchantRouteScroll
	if row < 0 || row >= merchantRoutePanelRowCount(r) {
		return InputAction{}
	}
	aid := r.merchantRouteArmy
	if row == 0 {
		r.closeMerchantRoutePanel()
		return InputAction{Kind: ActionClearMerchantRoute, ArmyID: aid}
	}
	route := r.merchantRouteOptions[row-1]
	if route == nil {
		return InputAction{}
	}
	r.focusMerchantRouteTarget(route)
	r.closeMerchantRoutePanel()
	return InputAction{Kind: ActionAssignMerchantRoute, ArmyID: aid, BuildingID: route.AssignmentKey()}
}
