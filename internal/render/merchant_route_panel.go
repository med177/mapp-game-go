package render

import (
	"fmt"
	"image/color"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	merchantRoutePanelW           = float32(660)
	merchantRoutePanelHeaderH     = float32(66)
	merchantRoutePanelRowH        = float32(48)
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
	close := gameui.NewButton(float64(panelX+merchantRoutePanelW-42), float64(panelY+10), 28, 28, "×")
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

	label := "Rota yok"
	if route := merchantRouteForKey(gs, a.TradeRouteKey); route != nil {
		label = "Aktif: " + merchantRouteDisplayName(gs, route)
	}
	footerX := float32(rect.X + rect.W + 12)
	footerY := layout.panelY + layout.panelH - siegeFooterH
	powerText := "Güç: " + itoa(a.TotalStrength(gs.UnitTypes)) + " / " + itoa(a.TotalDefense(gs.UnitTypes))
	powerX := layout.panelX + layout.panelW - armyPanelPadX*2 - float32(MeasureText(powerText, FaceSmall))

	route := merchantRouteForKey(gs, a.TradeRouteKey)
	shipCount := merchantShipCount(gs, a)
	bonusText := "Gemi bonusu: +0 hacim/tur · " + itoa(shipCount) + " gemi · rota yok"
	bonusColor := color.RGBA{180, 180, 180, 220}
	if route != nil {
		bonus := gs.MerchantFleetTradeRouteBonus(a, route)
		bonusText = "Gemi bonusu: +" + itoa(bonus) + " hacim/tur · " + itoa(shipCount) + " gemi"
		if gs.MerchantFleetSupportsTradeRoute(a, route) {
			bonusText += " · konum uygun"
			bonusColor = color.RGBA{145, 220, 155, 240}
		} else {
			bonusText += " · rota denizine git"
			bonusColor = color.RGBA{230, 180, 105, 240}
		}
	}
	bonusRight := powerX - 14
	bonusMaxW := bonusRight - footerX - 12
	if bonusMaxW < 0 {
		bonusMaxW = 0
	}
	bonusText = trimTextToWidth(bonusText, FaceSmall, float64(bonusMaxW))
	bonusW := float32(MeasureText(bonusText, FaceSmall))
	bonusX := bonusRight - bonusW
	routeW := bonusX - footerX - 12
	if routeW < 0 {
		routeW = 0
	}
	DrawText(screen, trimTextToWidth(label, FaceSmall, float64(routeW)), float64(footerX), float64(footerY+7), FaceSmall, color.RGBA{205, 185, 140, 230})
	DrawText(screen, bonusText, float64(bonusX), float64(footerY+7), FaceSmall, bonusColor)
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
	r.merchantRouteScroll = 0
	r.showMerchantRoutePanel = true
}

func (r *Renderer) closeMerchantRoutePanel() {
	if r == nil {
		return
	}
	r.showMerchantRoutePanel = false
	r.merchantRouteArmy = ""
	r.merchantRouteOptions = r.merchantRouteOptions[:0]
	r.merchantRouteScroll = 0
}

func (r *Renderer) drawMerchantRoutePanel(screen *ebiten.Image) {
	if r == nil || !r.showMerchantRoutePanel {
		return
	}
	layout := merchantRoutePanelLayoutFor(merchantRoutePanelRowCount(r))
	// Modal açıkken alttaki ordu panelinin butonları ve metinleri okunabilir
	// kalmamalı; input zaten bu katmanda tüketiliyor.
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), float32(ScreenHeight), color.RGBA{0, 0, 0, 205}, false)
	vector.FillRect(screen, layout.panelX, layout.panelY, layout.panelW, layout.panelH, panelBg, false)
	drawPanelBorder(screen, layout.panelX, layout.panelY, layout.panelW, layout.panelH)
	vector.FillRect(screen, layout.panelX, layout.panelY, layout.panelW, 3, panelBorder, false)

	DrawText(screen, "Merchant Ticaret Rotası", float64(layout.panelX+20), float64(layout.panelY+22), FaceLarge, ColorGold)
	DrawText(screen, "Seçili filonun görev rotasını belirleyin.", float64(layout.panelX+20), float64(layout.panelY+47), FaceSmall, ColorGray)
	gameui.DrawButton(screen, layout.close, gameui.ButtonStyle{
		BG: panelBg, Border: panelBorder, Text: ColorWhite, BorderWidth: 1,
	}, sharedTextRenderer{})

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
		y := layout.rowY + float32(row-start)*merchantRoutePanelRowH
		rowRect := gameui.Rect{X: float64(layout.rowX), Y: float64(y + 5), W: float64(layout.rowW), H: float64(merchantRoutePanelRowH - 10)}
		bg := color.RGBA{28, 22, 14, 220}
		if row%2 == 0 {
			bg = color.RGBA{38, 29, 16, 230}
		}
		vector.FillRect(screen, float32(rowRect.X), float32(rowRect.Y), float32(rowRect.W), float32(rowRect.H), bg, false)
		vector.StrokeRect(screen, float32(rowRect.X), float32(rowRect.Y), float32(rowRect.W), float32(rowRect.H), 1, color.RGBA{92, 70, 34, 180}, false)

		if row == 0 {
			DrawText(screen, "Görev yok", rowRect.X+14, rowRect.Y+8, FaceMed, color.RGBA{215, 180, 120, 255})
			DrawText(screen, "Bu filonun merchant rotasını kaldır", rowRect.X+180, rowRect.Y+11, FaceSmall, ColorGray)
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
		DrawText(screen, merchantRouteDisplayName(r.gs, route), rowRect.X+14, rowRect.Y+7, FaceMed, textColor)
		DrawText(screen, fmt.Sprintf("%s · %d/tur · %d altın/birim", economy.GoodNameTR(route.Good), amount, route.GoldPerUnit), rowRect.X+14, rowRect.Y+25, FaceSmall, ColorGray)
		if active {
			DrawText(screen, "AKTİF", rowRect.X+rowRect.W-62, rowRect.Y+12, FaceSmall, color.RGBA{170, 235, 170, 255})
		}
	}

	footerY := layout.panelY + layout.panelH - merchantRoutePanelFooterH
	if rowCount > merchantRoutePanelVisibleRows {
		DrawText(screen, fmt.Sprintf("Rotalar: %d-%d/%d", start+1, end, rowCount), float64(layout.panelX+18), float64(footerY+15), FaceSmall, ColorGray)
	}
	DrawText(screen, "ESC: kapat", float64(layout.panelX+layout.panelW-90), float64(footerY+15), FaceSmall, ColorGray)
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
	row := int((fy - float64(layout.rowY+4)) / float64(merchantRoutePanelRowH))
	if fx < float64(layout.rowX) || fx > float64(layout.rowX+layout.rowW) || row < 0 || row >= merchantRoutePanelVisibleRows {
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
	r.closeMerchantRoutePanel()
	return InputAction{Kind: ActionAssignMerchantRoute, ArmyID: aid, BuildingID: route.AssignmentKey()}
}
