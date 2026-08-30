package render

import (
	"image/color"
	"sort"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// TradeTab ticaret panelindeki sekmeler.
type TradeTab int

const (
	TradeTabRoutes TradeTab = iota // mevcut rotalar
	TradeTabNew                    // yeni ticaret anlaşması oluştur
	TradeTabMarket                 // açık pazarda manuel pazar işlemi
	TradeTabPrices                 // piyasa fiyatları
)

type TradeListFilter int

const (
	TradeListAll TradeListFilter = iota
	TradeListSellers
	TradeListBuyers
)

type TradeListSort int

const (
	TradeSortDistance TradeListSort = iota
	TradeSortPrice
	TradeSortStock
)

// TradeRouteListFilter, mevcut rotalar sekmesinin görünüm filtresidir.
// Sıfır değerinin oyuncuya ait rotalar olması, panel ilk açıldığında
// oyuncunun kendi ticaret akışını göstermesini sağlar.
type TradeRouteListFilter int

const (
	TradeRouteFilterOwned TradeRouteListFilter = iota
	TradeRouteFilterAll
)

const (
	tradePanelMinW     = float32(920)
	tradePanelMaxW     = float32(1560)
	tradePanelMinH     = float32(620)
	tradePanelMaxH     = float32(960)
	tradeTabH          = float32(44)
	tradeRowH          = float32(40)
	tradeActBtnH       = float32(34)
	tradeRouteSummaryH = float32(60)
)

const tradeBottomReserved = float32(170)

const tradeListFooterH = float64(20)

type tradeLayout struct {
	panelRect           gameui.Rect
	titleRect           gameui.Rect
	closeRect           gameui.Rect
	tabRects            []gameui.Rect
	routeFilterBtnRects []gameui.Rect
	filterLabelRect     gameui.Rect
	filterBtnRects      []gameui.Rect
	sortLabelRect       gameui.Rect
	sortBtnRects        []gameui.Rect
	goodFilterLabelRect gameui.Rect
	goodFilterBtnRects  []gameui.Rect
	leftTitleRect       gameui.Rect
	leftListRect        gameui.Rect
	rightTitleRect      gameui.Rect
	rightListRect       gameui.Rect
	actionCardRect      gameui.Rect
	marketTitleRect     gameui.Rect
	marketListRect      gameui.Rect
	marketActionRect    gameui.Rect
	autoExportRect      gameui.Rect
	listH               float32
	marketListH         float32
}

func clampTradeFloat(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func tradePanelLayout() tradeLayout {
	sw := float32(ScreenWidth)
	sh := float32(ScreenHeight)

	panelW := clampTradeFloat(sw*0.72, tradePanelMinW, tradePanelMaxW)
	if maxW := sw - 80; panelW > maxW {
		panelW = maxW
	}
	panelH := clampTradeFloat(sh*0.78, tradePanelMinH, tradePanelMaxH)
	if maxH := sh - 64; panelH > maxH {
		panelH = maxH
	}

	panelRect := gameui.Rect{
		X: float64(sw/2 - panelW/2),
		Y: float64(sh/2 - panelH/2),
		W: float64(panelW),
		H: float64(panelH),
	}
	box := gameui.BoxFromRect(panelRect).InsetXY(18, 12)
	headerRect, box := box.CutTop(36, 18)
	closeRect, titleBox := gameui.BoxFromRect(headerRect).CutRight(40, 18)
	tabsRow, box := box.CutTop(float64(tradeTabH), 20)
	controlRow, box := box.CutTop(32, 14)
	goodFilterRow, box := box.CutTop(32, 10)
	bodyCols := box.SplitColumns(28, 0.40, 0.60)
	leftBox := gameui.BoxFromRect(bodyCols[0])
	rightBox := gameui.BoxFromRect(bodyCols[1])
	leftTitleRect, leftBox := leftBox.CutTop(24, 8)
	_, leftListBox := leftBox.CutBottom(tradeListFooterH, 0)
	actionCardRect, rightBox := rightBox.CutBottom(174, 18)
	rightTitleRect, rightBox := rightBox.CutTop(24, 8)
	listH := float32(leftBox.Rect.H)
	if listH < 120 {
		listH = 120
	}
	controlCols := gameui.BoxFromRect(controlRow).SplitColumns(18, 0.34, 0.33, 0.33)
	filterLabelRect, filterBtnBox := gameui.BoxFromRect(controlCols[0]).CutLeft(76, 12)
	sortLabelRect, sortBtnBox := gameui.BoxFromRect(controlCols[1]).CutLeft(52, 12)
	_, routeFilterBtnBox := gameui.BoxFromRect(controlCols[0]).CutLeft(76, 12)
	goodFilterLabelRect, goodFilterBtnBox := gameui.BoxFromRect(goodFilterRow).CutLeft(52, 10)
	marketActionRect, marketBody := box.CutBottom(174, 18)
	marketTitleRect, marketListBox := marketBody.CutTop(24, 8)
	_, marketListBody := marketListBox.CutBottom(tradeListFooterH, 0)

	return tradeLayout{
		panelRect:           panelRect,
		titleRect:           titleBox.Rect,
		closeRect:           closeRect,
		tabRects:            gameui.BoxFromRect(tabsRow).SplitColumns(12, 1, 1, 1, 1),
		routeFilterBtnRects: routeFilterBtnBox.SplitColumns(12, 1, 1),
		filterLabelRect:     filterLabelRect,
		filterBtnRects:      filterBtnBox.SplitColumns(12, 1, 1, 1),
		sortLabelRect:       sortLabelRect,
		sortBtnRects:        sortBtnBox.SplitColumns(12, 1, 1, 1),
		goodFilterLabelRect: goodFilterLabelRect,
		goodFilterBtnRects:  goodFilterBtnBox.SplitColumns(8, 1, 1, 1, 1, 1, 1),
		leftTitleRect:       leftTitleRect,
		leftListRect:        leftListBox.Rect,
		rightTitleRect:      rightTitleRect,
		rightListRect:       rightBox.Rect,
		actionCardRect:      actionCardRect,
		marketTitleRect:     marketTitleRect,
		marketListRect:      marketListBody.Rect,
		marketActionRect:    marketActionRect,
		autoExportRect:      controlCols[2],
		listH:               float32(leftListBox.Rect.H),
		marketListH:         float32(marketListBody.Rect.H),
	}
}

func tradeActionCardRect(layout tradeLayout, _ int) (x, y, w, h float32) {
	r := layout.actionCardRect
	return float32(r.X), float32(r.Y), float32(r.W), float32(r.H)
}

// tradePanelRect ticaret panelinin ortalanmış dikdörtgenini döner.
func tradePanelRect() (x, y, w, h float32) {
	layout := tradePanelLayout()
	return float32(layout.panelRect.X), float32(layout.panelRect.Y), float32(layout.panelRect.W), float32(layout.panelRect.H)
}

// tradeCloseRect kapatma butonu.
func tradeCloseRect() (x, y, w, h float32) {
	layout := tradePanelLayout()
	return float32(layout.closeRect.X), float32(layout.closeRect.Y), float32(layout.closeRect.W), float32(layout.closeRect.H)
}

// tradeCloseHit tıklama kontrolü.
func tradeCloseHit(mx, my float64) bool {
	x, y, w, h := tradeCloseRect()
	return mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h)
}

// DrawTradePanel ticaret panelini çizer.
// Tab 0: mevcut rotalar, Tab 1: yeni rota, Tab 2: pazar, Tab 3: piyasa fiyatları
func DrawTradePanel(screen *ebiten.Image, gs *state.GameState, tab TradeTab, focusFaction int, focusGood int, scroll int, amount int, routeFilter TradeRouteListFilter, listFilter TradeListFilter, listSort TradeListSort) {
	layout := tradePanelLayout()
	px, py, pw, ph := float32(layout.panelRect.X), float32(layout.panelRect.Y), float32(layout.panelRect.W), float32(layout.panelRect.H)

	// Arka plan overlay
	drawUIOverlay(screen, color.RGBA{8, 6, 4, 200})

	// Panel çerçevesi
	drawUIPanelRect(screen, layout.panelRect, panelBg, panelBorder, 1)
	vector.FillRect(screen, px, py, pw, 3, panelBorder, false)

	// Başlık
	drawUIPanelTitle(screen, gameui.Rect{X: layout.titleRect.X, Y: layout.titleRect.Y + 2, W: layout.titleRect.W, H: layout.titleRect.H}, "── Ticaret ──")

	// Kapatma butonu
	closeBtn := buildTradeCloseButton()
	drawTradeButton(screen, closeBtn, false)

	// Sekmeler
	for _, btn := range buildTradeTabButtons() {
		drawTradeButton(screen, btn.Button, btn.Tab != tab)
	}

	contentY := float32(layout.leftTitleRect.Y)
	contentH := ph - (contentY - py) - 18

	switch tab {
	case TradeTabRoutes:
		drawTradeRoutesTab(screen, gs, layout, contentY, pw, contentH, scroll, routeFilter)
	case TradeTabNew:
		drawTradeNewTab(screen, gs, layout, focusFaction, scroll)
	case TradeTabMarket:
		drawTradeMarketTab(screen, gs, layout, focusFaction, focusGood, scroll, amount, listFilter, listSort)
	case TradeTabPrices:
		drawTradePricesTab(screen, gs, px, contentY, pw, contentH)
	}
}

// drawTradeRoutesTab mevcut aktif ticaret rotalarını listeler.
func drawTradeRoutesTab(screen *ebiten.Image, gs *state.GameState, layout tradeLayout, y float32, w float32, h float32, scroll int, routeFilter TradeRouteListFilter) {
	px := float32(layout.panelRect.X)
	drawTradeRouteListControls(screen, layout, routeFilter)
	listReservedH := float32(0)
	if routeFilter == TradeRouteFilterOwned {
		listReservedH = tradeRouteSummaryH + 8
		drawTradeRouteSummary(screen, layout, gs, y, w, h)
	}
	routeCount := filteredTradeRouteCount(gs, routeFilter)
	if routeCount == 0 {
		DrawTextCentered(screen, "Aktif ticaret rotası yok.", float64(px)+float64(w)/2, float64(y)+40, FaceMed, ColorGray)
		DrawTextCentered(screen, "Diplomasi → Ticaret anlaşması yaparak rota oluşturun.", float64(px)+float64(w)/2, float64(y)+62, FaceSmall, ColorGray)
		return
	}

	// Başlıklar
	headerRow := gameui.NewTableRow(gameui.Rect{X: float64(px) + 10, Y: float64(y) + 4, W: float64(w) - 70}, []gameui.TableCell{
		{Text: "Mal", Color: ColorGold, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 0.35},
		{Text: "Gönderen", Color: ColorGold, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 0.20},
		{Text: "Alan", Color: ColorGold, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 0.20},
		{Text: "Miktar/Tur", Color: ColorGold, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 0.25},
	}, 0)
	drawUITableRow(screen, headerRow)

	rowY := y + 24
	listH := h - listReservedH
	visibleRows := int((listH - 30) / tradeRowH)
	if visibleRows < 1 {
		visibleRows = 1
	}
	start := scroll
	end := start + visibleRows
	if end > routeCount {
		end = routeCount
	}
	filteredIndex := 0
	shownRows := 0
	for _, tr := range gs.TradeRoutes {
		if !tradeRouteMatchesFilter(gs, tr, routeFilter) {
			continue
		}
		if filteredIndex < start {
			filteredIndex++
			continue
		}
		if filteredIndex >= end {
			break
		}
		ry := rowY + float32(shownRows)*tradeRowH

		bg := color.RGBA{20, 18, 12, 200}
		if shownRows%2 == 0 {
			bg = color.RGBA{28, 22, 16, 210}
		}
		vector.FillRect(screen, px+4, ry, w-8, tradeRowH-4, bg, false)

		row := gameui.NewTableRow(gameui.Rect{X: float64(px) + 10, Y: float64(ry) + 10, W: float64(w) - 70}, []gameui.TableCell{
			{Text: economy.GoodNameTR(tr.Good), Color: ColorWhite, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 0.35},
			{Text: factionDisplayName(gs, tr.FromFactionID), Color: ColorGray, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 0.20},
			{Text: factionDisplayName(gs, tr.ToFactionID), Color: ColorGray, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 0.20},
			{Text: itoa(tr.AmountPerTurn) + " @" + itoa(tr.GoldPerUnit) + " altın", Color: ColorGold, Variant: gameui.TextSmall, Align: gameui.TextAlignStart, Weight: 0.25},
		}, 0)
		drawUITableRow(screen, row)
		filteredIndex++
		shownRows++
	}

	// Scroll bilgisi
	if routeCount > visibleRows {
		info := "Rotalar: " + itoa(start+1) + "-" + itoa(end) + "/" + itoa(routeCount)
		DrawText(screen, info, float64(px)+10, float64(y+listH-16), FaceSmall, ColorGray)
	}
}

func drawTradeRouteSummary(screen *ebiten.Image, layout tradeLayout, gs *state.GameState, y, w, h float32) {
	income, expense, customs := tradeRouteFinancialTotals(gs)
	rect := tradeRouteSummaryRect(layout, y, w, h)
	drawUIPanelRect(screen, rect, color.RGBA{16, 14, 10, 220}, panelBorder, 1)
	playerID := gs.PlayerFactionID
	power := gs.TradePowerForFaction(playerID)
	share := gs.TradePowerSharePercent(playerID)
	drawUISectionLabel(screen, rect.X+10, rect.Y+5, "Toplam Rota Özeti (tur başına):")
	drawUILabel(screen, gameui.Rect{X: rect.X + rect.W - 10, Y: rect.Y + 5, W: 0}, "Ticaret gücü: "+itoa(power)+" ("+itoa(share)+"%)", ColorGray, gameui.TextSmall, gameui.TextAlignEnd)

	const gap = 12.0
	columnW := (rect.W - 20 - gap*3) / 4
	rowY := rect.Y + 27
	drawUIKeyValueRowLeading(screen, rect.X+10, rowY, columnW, "Gelir", "+"+itoa(income)+" altın", ColorGray, color.RGBA{145, 220, 155, 245}, 12)
	drawUIKeyValueRowLeading(screen, rect.X+10+columnW+gap, rowY, columnW, "Gider", "-"+itoa(expense)+" altın", ColorGray, color.RGBA{230, 170, 135, 240}, 12)
	drawUIKeyValueRowLeading(screen, rect.X+10+(columnW+gap)*2, rowY, columnW, "Gümrük (%"+itoa(tradeRouteCustomsRate(gs))+")", "+"+itoa(customs)+" altın", ColorGray, color.RGBA{205, 180, 110, 245}, 12)
	netColor := color.RGBA{145, 220, 155, 245}
	net := income - expense + customs
	if net < 0 {
		netColor = color.RGBA{230, 170, 135, 240}
	}
	drawUIKeyValueRowLeading(screen, rect.X+10+(columnW+gap)*3, rowY, columnW, "Net Fark", formatSignedAmount(net)+" altın", ColorGray, netColor, 12)
}

func tradeRouteSummaryRect(layout tradeLayout, y, w, h float32) gameui.Rect {
	return gameui.Rect{
		X: layout.panelRect.X + 4,
		Y: float64(y + h - tradeRouteSummaryH),
		W: float64(w - 8),
		H: float64(tradeRouteSummaryH - 4),
	}
}

func tradeRouteFinancialTotals(gs *state.GameState) (income, expense, customs int) {
	if gs == nil {
		return 0, 0, 0
	}
	playerID := string(gs.PlayerFactionID)
	for _, route := range gs.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 {
			continue
		}
		gold := route.GoldEarned()
		if route.FromFactionID == playerID {
			income += gold
		}
		if route.ToFactionID == playerID {
			expense += gold
			customs += gold * tradeRouteCustomsRate(gs) / 100
		}
	}
	return income, expense, customs
}

func tradeRouteCustomsRate(gs *state.GameState) int {
	if gs == nil {
		return economy.TradeRouteCustomsRatePercent
	}
	rate := economy.TradeRouteCustomsRatePercent + gs.TradePowerSharePercent(gs.PlayerFactionID)/10
	if rate > 20 {
		return 20
	}
	return rate
}

func tradeRouteMatchesFilter(gs *state.GameState, route *economy.TradeRoute, filter TradeRouteListFilter) bool {
	if gs == nil || route == nil {
		return false
	}
	return filter != TradeRouteFilterOwned || route.FromFactionID == string(gs.PlayerFactionID) || route.ToFactionID == string(gs.PlayerFactionID)
}

func filteredTradeRouteCount(gs *state.GameState, filter TradeRouteListFilter) int {
	count := 0
	if gs == nil {
		return count
	}
	for _, route := range gs.TradeRoutes {
		if tradeRouteMatchesFilter(gs, route, filter) {
			count++
		}
	}
	return count
}

func filteredTradeRoutes(gs *state.GameState, filter TradeRouteListFilter) []*economy.TradeRoute {
	if gs == nil {
		return nil
	}
	routes := make([]*economy.TradeRoute, 0, len(gs.TradeRoutes))
	for _, route := range gs.TradeRoutes {
		if !tradeRouteMatchesFilter(gs, route, filter) {
			continue
		}
		routes = append(routes, route)
	}
	return routes
}

// drawTradeNewTab yeni ticaret anlaşması arayüzü.
func drawTradeNewTab(screen *ebiten.Image, gs *state.GameState, layout tradeLayout, focusFaction int, scroll int) {
	px, y, w := float32(layout.panelRect.X), float32(layout.leftTitleRect.Y), float32(layout.panelRect.W)
	candidates := sortedFactionsForTradeAgreements(gs)
	if len(candidates) == 0 {
		DrawTextCentered(screen, "Şu anda yeni ticaret anlaşması açılabilir hedef yok.", float64(px)+float64(w)/2, float64(y)+40, FaceMed, ColorGray)
		DrawTextCentered(screen, "Savaşta olmayan, kapasitesi yeterli ve rota sınırına takılmayan devletler burada görünür.", float64(px)+float64(w)/2, float64(y)+62, FaceSmall, ColorGray)
		return
	}

	drawUISectionLabel(screen, layout.leftTitleRect.X, layout.leftTitleRect.Y+2, "Anlaşma Adayları:")
	visibleRows := int(layout.listH / 28)
	if visibleRows < 1 {
		visibleRows = 1
	}
	factionItems := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		f := gs.Factions[candidate.ID]
		name := string(candidate.ID)
		if f != nil && f.NameTR != "" {
			name = f.NameTR
		}
		label := name + " | " + faction.DiplomaticStanceLabelTR(candidate.Stance) + " | Skor: " + itoa(candidate.Score)
		if !candidate.Eligible && candidate.BlockReason != "" {
			label += " | " + candidate.BlockReason
		}
		factionItems = append(factionItems, trimTextToWidth(label, FaceSmall, layout.leftListRect.W-12))
	}
	factionList := buildTradeFactionList(layout, visibleRows, factionItems, scroll, focusFaction)
	style := tradeListViewStyle()
	style.PaginationPrefix = "Adaylar:"
	gameui.DrawListView(screen, factionList, style, renderText)

	drawUISectionLabel(screen, layout.rightTitleRect.X, layout.rightTitleRect.Y+2, "Anlaşma Özeti:")
	if focusFaction < 0 || focusFaction >= len(candidates) {
		drawUIMutedText(screen, layout.rightListRect.X+6, layout.rightListRect.Y+10, "Önce sol listeden bir hedef devlet seçin.")
		return
	}
	selected := candidates[focusFaction]
	target := gs.Factions[selected.ID]
	cardX, cardY, cardW, cardH := tradeActionCardRect(layout, visibleRows)
	vector.FillRect(screen, cardX, cardY, cardW, cardH, color.RGBA{16, 14, 10, 220}, false)
	vector.StrokeRect(screen, cardX, cardY, cardW, cardH, 1, panelBorder, false)

	name := string(selected.ID)
	if target != nil && target.NameTR != "" {
		name = target.NameTR
	}
	detailLines := []string{
		"Hedef: " + name,
		"Duruş: " + faction.DiplomaticStanceLabelTR(selected.Stance) + " | İlişki: " + itoa(selected.Score),
		"Rota kapasitesi: Sen " + itoa(selected.PlayerRouteCapacityUsed) + "/" + itoa(selected.PlayerTradeCap) + " | Onlar " + itoa(selected.TargetRouteCapacityUsed) + "/" + itoa(selected.TargetTradeCap),
		"Aktif partner: Sen " + itoa(selected.PlayerPartners) + "/" + itoa(diplomacy.TradePartnerLimit(gs, gs.PlayerFactionID)) + " | Onlar " + itoa(selected.TargetPartners) + "/" + itoa(diplomacy.TradePartnerLimit(gs, selected.ID)),
	}
	detailColors := []color.Color{
		color.RGBA{220, 205, 170, 230},
		color.RGBA{190, 190, 215, 230},
		color.RGBA{170, 205, 175, 230},
		color.RGBA{170, 190, 210, 230},
	}
	if selected.BlockReason == "" {
		detailLines = append(detailLines, "Durum: Anlaşma açılabilir")
		detailColors = append(detailColors, color.RGBA{150, 220, 150, 230})
	} else {
		detailLines = append(detailLines, "Engel: "+selected.BlockReason)
		detailColors = append(detailColors, color.RGBA{230, 170, 120, 230})
	}
	drawUIInfoBlock(screen, float64(layout.rightListRect.X)+6, layout.rightListRect.Y+10, detailLines, detailColors)

	actionBtn := buildTradeAgreementButton(layout, selected.Eligible)
	styleBG := color.RGBA{72, 104, 54, 230}
	styleBorder := color.RGBA{160, 205, 140, 230}
	if !selected.Eligible {
		styleBG = color.RGBA{52, 52, 52, 210}
		styleBorder = color.RGBA{110, 110, 110, 210}
	}
	drawTradeActionButton(screen, actionBtn, styleBG, styleBorder)
}

func drawTradeMarketTab(screen *ebiten.Image, gs *state.GameState, layout tradeLayout, focusFaction int, focusGood int, scroll int, amount int, listFilter TradeListFilter, listSort TradeListSort) {
	playerF := gs.Factions[gs.PlayerFactionID]
	if playerF == nil {
		return
	}
	if amount < 1 {
		amount = 1
	}
	if focusGood < 0 {
		focusGood = 0
	}
	goods := tradeSelectableGoods()
	if focusGood >= len(goods) {
		focusGood = len(goods) - 1
	}
	selectedGood := goods[focusGood]
	price := tradeMarketPrice(gs, selectedGood)
	amount = clampTradeAmountToGold(amount, playerF, price)

	px, y, w := float32(layout.panelRect.X), float32(layout.marketTitleRect.Y), float32(layout.panelRect.W)
	factions := sortedFactionsForMarket(gs, focusGood, listFilter, listSort)
	drawTradeListControls(screen, layout, gs, listFilter, listSort)
	drawTradeGoodFilterControls(screen, layout, focusGood)
	if len(factions) == 0 {
		DrawTextCentered(screen, "Açık pazarda işlem yapılabilecek devlet yok.", float64(px)+float64(w)/2, float64(y)+40, FaceMed, ColorGray)
		DrawTextCentered(screen, "Savaşta olmadığınız bir devlet veya uygun stok bulunmuyor.", float64(px)+float64(w)/2, float64(y)+62, FaceSmall, ColorGray)
		cardX, cardY, cardW, cardH := tradeMarketActionCardRect(layout)
		vector.FillRect(screen, cardX, cardY, cardW, cardH, color.RGBA{16, 14, 10, 220}, false)
		vector.StrokeRect(screen, cardX, cardY, cardW, cardH, 1, panelBorder, false)
		drawUIInfoBlock(screen, float64(cardX)+12, float64(cardY)+16, []string{
			"Acil tahıl satışı",
			"Yalnızca depo kapasitesini aşan tahıl satılabilir.",
			"Birim fiyat: " + itoa(gs.EmergencyGrainSaleUnitPrice()) + " altın (-%30)",
		}, []color.Color{
			color.RGBA{220, 205, 170, 230},
			color.RGBA{190, 190, 215, 230},
			color.RGBA{230, 180, 120, 230},
		})
		emergencyBtn := buildTradeEmergencyGrainSaleButtonAt(tradeMarketActionCardRect, layout, gs.EmergencyGrainSaleLimit() > 0)
		drawTradeEmergencyButton(screen, emergencyBtn)
		return
	}

	drawUISectionLabel(screen, layout.marketTitleRect.X, layout.marketTitleRect.Y+2, "Hedef Fraksiyon | Mal: "+economy.GoodNameTR(selectedGood))

	visibleRows := int(layout.marketListH / 28)
	if visibleRows < 1 {
		visibleRows = 1
	}
	factionItems := make([]string, 0, len(factions))
	for _, fid := range factions {
		if f := gs.Factions[fid]; f != nil {
			price := tradeMarketPrice(gs, selectedGood)
			supply := gs.MarketSellOffer(fid, selectedGood)
			demand := gs.MarketBuyOrder(fid, selectedGood, price)
			line := f.NameTR + " | Stok: " + itoa(getFactionGoodAmount(f, selectedGood)) + " | Satış arzı: " + itoa(supply) + " | Alım talebi: " + itoa(demand) + " | Fiyat: " + itoa(price) + " altın"
			factionItems = append(factionItems, trimTextToWidth(line, FaceSmall, float64(layout.marketListRect.W)-12))
		}
	}
	factionList := buildTradeMarketFactionList(layout, visibleRows, factionItems, scroll, focusFaction)
	style := tradeListViewStyle()
	style.PaginationPrefix = "Partnerler:"
	gameui.DrawListView(screen, factionList, style, renderText)

	if focusFaction >= 0 && focusFaction < len(factions) {
		cardX, cardY, cardW, cardH := tradeMarketActionCardRect(layout)
		vector.FillRect(screen, cardX, cardY, cardW, cardH, color.RGBA{16, 14, 10, 220}, false)
		vector.StrokeRect(screen, cardX, cardY, cardW, cardH, 1, panelBorder, false)
		qtyButtons, buyBtn, sellBtn := buildTradeMarketActionButtons(layout, amount, playerF, price)
		for _, btn := range qtyButtons {
			drawTradeButton(screen, btn, false)
		}
		drawTradeActionButton(screen, buyBtn, color.RGBA{62, 112, 62, 230}, color.RGBA{150, 200, 150, 230})
		drawTradeActionButton(screen, sellBtn, color.RGBA{112, 76, 52, 230}, color.RGBA{210, 170, 130, 230})

		target := gs.Factions[factions[focusFaction]]
		good := selectedGood
		price := tradeMarketPrice(gs, good)
		maxBuy := tradeMaxBuyAmount(gs, playerF, target, good, price)
		maxSell := tradeMaxSellAmount(gs, playerF, target, good, price)
		totalGold := amount * price
		if good == economy.GoodGrain {
			emergencyBtn := buildTradeEmergencyGrainSaleButtonAt(tradeMarketActionCardRect, layout, gs.EmergencyGrainSaleLimit() > 0)
			drawTradeEmergencyButton(screen, emergencyBtn)
		}

		detailX, detailY := float64(cardX)+12, float64(cardY)+16
		drawUILabel(screen, gameui.Rect{X: detailX, Y: detailY}, "Seçili: "+economy.GoodNameTR(good)+" | Hedef: "+target.NameTR, color.RGBA{200, 190, 170, 220}, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: detailX, Y: detailY + 24}, "Miktar: "+itoa(amount)+" | Tutar: "+itoa(totalGold)+" altın", color.RGBA{230, 210, 155, 230}, gameui.TextMedium, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: detailX, Y: detailY + 48}, "Al max: "+itoa(maxBuy)+" | Sat max: "+itoa(maxSell), color.RGBA{160, 190, 210, 220}, gameui.TextSmall, gameui.TextAlignStart)
	}
}

type tradeTabButton struct {
	Tab    TradeTab
	Button gameui.Button
}

type tradeChoiceButton struct {
	Value  int
	Button gameui.Button
}

func buildTradeCloseButton() gameui.Button {
	x, y, w, h := tradeCloseRect()
	btn := gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "").WithIcon(gameui.IconClose)
	btn.IconSize = 13
	return btn
}

func buildTradeTabButtons() []tradeTabButton {
	layout := tradePanelLayout()
	tabLabels := []string{"Mevcut Rotalar", "Yeni Rota", "Açık Pazar", "Piyasa Fiyatları"}
	out := make([]tradeTabButton, 0, len(tabLabels))
	for i, label := range tabLabels {
		r := layout.tabRects[i]
		out = append(out, tradeTabButton{
			Tab:    TradeTab(i),
			Button: gameui.NewButton(r.X, r.Y, r.W, r.H, label),
		})
	}
	return out
}

func buildTradeFilterButtons(layout tradeLayout) []tradeChoiceButton {
	labels := []string{"Hepsi", "Satanlar", "Alanlar"}
	out := make([]tradeChoiceButton, 0, len(labels))
	for i, label := range labels {
		r := layout.filterBtnRects[i]
		out = append(out, tradeChoiceButton{
			Value:  i,
			Button: gameui.NewButton(r.X, r.Y, r.W, r.H, label),
		})
	}
	return out
}

func buildTradeRouteFilterButtons(layout tradeLayout) []tradeChoiceButton {
	labels := []string{"Tüm Rotalar", "Sadece Bize Ait"}
	values := []TradeRouteListFilter{TradeRouteFilterAll, TradeRouteFilterOwned}
	out := make([]tradeChoiceButton, 0, len(labels))
	for i, label := range labels {
		r := layout.routeFilterBtnRects[i]
		out = append(out, tradeChoiceButton{
			Value:  int(values[i]),
			Button: gameui.NewButton(r.X, r.Y, r.W, r.H, label),
		})
	}
	return out
}

func buildTradeSortButtons(layout tradeLayout) []tradeChoiceButton {
	labels := []string{"İsim", "Fiyat", "Stok"}
	out := make([]tradeChoiceButton, 0, len(labels))
	for i, label := range labels {
		r := layout.sortBtnRects[i]
		out = append(out, tradeChoiceButton{
			Value:  i,
			Button: gameui.NewButton(r.X, r.Y, r.W, r.H, label),
		})
	}
	return out
}

func buildTradeGoodFilterButtons(layout tradeLayout) []tradeChoiceButton {
	labels := []string{"Tahıl", "Demir", "Kereste", "Taş", "Baharat", "Kumaş"}
	goods := tradeSelectableGoods()
	out := make([]tradeChoiceButton, 0, len(labels))
	for i, label := range labels {
		if i >= len(layout.goodFilterBtnRects) || i >= len(goods) {
			break
		}
		r := layout.goodFilterBtnRects[i]
		out = append(out, tradeChoiceButton{
			Value:  i,
			Button: gameui.NewButton(r.X, r.Y, r.W, r.H, label),
		})
	}
	return out
}

func buildTradeFactionList(layout tradeLayout, visibleRows int, items []string, scroll int, selected int) gameui.ListView {
	return buildTradeFactionListAt(layout.leftListRect, visibleRows, items, scroll, selected)
}

func buildTradeMarketFactionList(layout tradeLayout, visibleRows int, items []string, scroll int, selected int) gameui.ListView {
	return buildTradeFactionListAt(layout.marketListRect, visibleRows, items, scroll, selected)
}

func buildTradeFactionListAt(rect gameui.Rect, visibleRows int, items []string, scroll int, selected int) gameui.ListView {
	list := gameui.NewListView(rect.X, rect.Y, rect.W, rect.H, 28, visibleRows, items)
	list.Scroll = scroll
	list.Selected = selected
	return list
}

func tradeMarketActionCardRect(layout tradeLayout) (x, y, w, h float32) {
	r := layout.marketActionRect
	return float32(r.X), float32(r.Y), float32(r.W), float32(r.H)
}

func tradeListClickedIndex(list gameui.ListView, input gameui.InputState) (int, bool) {
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

func buildTradeActionButtons(layout tradeLayout, _ int) ([]gameui.Button, gameui.Button, gameui.Button) {
	return buildTradeActionButtonsAt(defaultTradeActionCardRect, layout)
}

func defaultTradeActionCardRect(layout tradeLayout) (x, y, w, h float32) {
	return tradeActionCardRect(layout, 1)
}

func tradeActionControlsX(cardX, cardW float32) float32 {
	return cardX + clampTradeFloat(cardW*0.20, 220, 360)
}

func buildTradeActionButtonsAt(cardRect func(tradeLayout) (float32, float32, float32, float32), layout tradeLayout) ([]gameui.Button, gameui.Button, gameui.Button) {
	return buildTradeActionButtonsAtWithPlusEnabled(cardRect, layout, true)
}

func buildTradeActionButtonsAtWithPlusEnabled(cardRect func(tradeLayout) (float32, float32, float32, float32), layout tradeLayout, plusEnabled bool) ([]gameui.Button, gameui.Button, gameui.Button) {
	cardX, cardY, cardW, cardH := cardRect(layout)
	controlsX := tradeActionControlsX(cardX, cardW)
	btnY := cardY + cardH - tradeActBtnH - 18
	qty := []gameui.Button{
		gameui.NewButton(float64(controlsX), float64(btnY), 54, float64(tradeActBtnH), "-10"),
		gameui.NewButton(float64(controlsX+64), float64(btnY), 54, float64(tradeActBtnH), "+10"),
	}
	qty[1].Enabled = plusEnabled
	buyX := controlsX + 136
	buyBtn := gameui.NewButton(float64(buyX), float64(btnY), 110, float64(tradeActBtnH), "AL").WithIcon(gameui.IconBuy)
	sellBtn := gameui.NewButton(float64(buyX+124), float64(btnY), 110, float64(tradeActBtnH), "SAT").WithIcon(gameui.IconSell)
	return qty, buyBtn, sellBtn
}

func buildTradeMarketActionButtons(layout tradeLayout, amount int, player *faction.Faction, price int) ([]gameui.Button, gameui.Button, gameui.Button) {
	maxAffordable := tradeMaxAffordableAmount(player, price)
	plusEnabled := amount+10 <= maxAffordable
	return buildTradeActionButtonsAtWithPlusEnabled(tradeMarketActionCardRect, layout, plusEnabled)
}

func buildTradeEmergencyGrainSaleButton(layout tradeLayout, _ int, enabled bool) gameui.Button {
	return buildTradeEmergencyGrainSaleButtonAt(defaultTradeActionCardRect, layout, enabled)
}

func buildTradeEmergencyGrainSaleButtonAt(cardRect func(tradeLayout) (float32, float32, float32, float32), layout tradeLayout, enabled bool) gameui.Button {
	cardX, cardY, cardW, cardH := cardRect(layout)
	btn := gameui.NewButton(float64(tradeActionControlsX(cardX, cardW)+136), float64(cardY+cardH-2*tradeActBtnH-28), 234, float64(tradeActBtnH), "ACİL TAHIL SAT").WithIcon(gameui.IconSell)
	btn.Enabled = enabled
	return btn
}

func buildTradeAgreementButton(layout tradeLayout, enabled bool) gameui.Button {
	cardX, cardY, cardW, cardH := tradeActionCardRect(layout, 1)
	btn := gameui.NewButton(float64(cardX+cardW-250), float64(cardY+cardH-tradeActBtnH-18), 220, float64(tradeActBtnH), "Ticaret Anlaşması Aç").WithIcon(gameui.IconSend)
	btn.Enabled = enabled
	return btn
}

func tradeButtonStyle(active bool) gameui.ButtonStyle {
	style := menuButtonStyle
	style.Text = ColorWhite
	style.TextOffsetY = 0
	if active {
		style.BG = color.RGBA{55, 45, 25, 230}
	} else {
		style.BG = color.RGBA{25, 20, 14, 200}
	}
	return style
}

func tradeListViewStyle() gameui.ListViewStyle {
	return gameui.ListViewStyle{
		RowBG:          color.RGBA{20, 18, 12, 200},
		SelectedRowBG:  color.RGBA{55, 45, 25, 230},
		TextColor:      ColorWhite,
		SelectedText:   ColorWhite,
		MutedText:      ColorGray,
		RowTextOffsetY: 4,
		TextVariant:    gameui.TextSmall,
	}
}

func drawTradeButton(screen *ebiten.Image, btn gameui.Button, inactive bool) {
	drawUIButtonWidget(screen, btn, tradeButtonStyle(!inactive))
}

func drawTradeActionButton(screen *ebiten.Image, btn gameui.Button, bg, border color.RGBA) {
	style := menuButtonStyle
	style.BG = bg
	style.Border = border
	style.Text = ColorWhite
	style.TextOffsetY = 0
	drawUIButtonWidget(screen, btn, style)
}

func drawTradeEmergencyButton(screen *ebiten.Image, btn gameui.Button) {
	bg := color.RGBA{112, 76, 52, 230}
	border := color.RGBA{210, 170, 130, 230}
	if !btn.Enabled {
		bg = color.RGBA{52, 52, 52, 210}
		border = color.RGBA{110, 110, 110, 210}
	}
	drawTradeActionButton(screen, btn, bg, border)
}

func drawTradeChoiceButton(screen *ebiten.Image, btn gameui.Button, active bool, activeBG color.RGBA) {
	style := tradeButtonStyle(true)
	style.TextOffsetY = 0
	style.BG = color.RGBA{26, 24, 21, 220}
	if active {
		style.BG = activeBG
	}
	drawUIButtonWidget(screen, btn, style)
}

func drawTradeRouteListControls(screen *ebiten.Image, layout tradeLayout, filter TradeRouteListFilter) {
	drawUISectionLabel(screen, layout.filterLabelRect.X, layout.filterLabelRect.Y+12, "Filtre:")
	for _, btn := range buildTradeRouteFilterButtons(layout) {
		drawTradeChoiceButton(screen, btn.Button, TradeRouteListFilter(btn.Value) == filter, color.RGBA{70, 62, 36, 235})
	}
}

// drawTradePricesTab piyasa fiyatlarını gösterir.
func drawTradePricesTab(screen *ebiten.Image, gs *state.GameState, px float32, y float32, w float32, _ float32) {
	if gs.MarketPrices == nil {
		DrawTextCentered(screen, "Piyasa fiyatları henüz oluşturulmadı.", float64(px)+float64(w)/2, float64(y)+40, FaceMed, ColorGray)
		return
	}

	goods := economy.TradeGoods()

	// Başlıklar
	colX := []float32{px + 10, px + w*0.30, px + w*0.55, px + w*0.75, px + w*0.90}
	headers := []string{"Mal", "Base Fiyat", "Güncel Fiyat", "Değişim", "Pazar Arzı"}
	DrawText(screen, "Dinamik Piyasa Fiyatları (Açık Pazar Arz-Talep):", float64(px)+10, float64(y)+4, FaceSmall, ColorGold)
	for i, hdr := range headers {
		DrawText(screen, hdr, float64(colX[i]), float64(y)+20, FaceSmall, ColorGray)
	}

	ry := y + 38
	for _, good := range goods {
		basePrice := economy.BaseGoldValue[good]
		currentPrice := gs.MarketPrices[good]
		goodName := economy.GoodNameTR(good)

		// Değişim yüzdesi
		changePct := ((currentPrice - basePrice) * 100) / basePrice
		changeStr := "+" + itoa(changePct) + "%"
		changeCol := color.RGBA{60, 220, 60, 255}
		if changePct < 0 {
			changeStr = itoa(changePct) + "%"
			changeCol = color.RGBA{220, 60, 60, 255}
		}
		if changePct == 0 {
			changeStr = "0%"
			changeCol = ColorGray
		}

		bg := color.RGBA{20, 18, 12, 200}
		vector.FillRect(screen, px+4, ry, w-8, 22, bg, false)

		DrawText(screen, goodName, float64(colX[0]), float64(ry)+3, FaceSmall, ColorWhite)
		DrawText(screen, itoa(basePrice), float64(colX[1]), float64(ry)+3, FaceSmall, ColorGray)
		DrawText(screen, itoa(currentPrice), float64(colX[2]), float64(ry)+3, FaceSmall, ColorYellow)
		DrawText(screen, changeStr, float64(colX[3]), float64(ry)+3, FaceSmall, changeCol)

		// Pazar arzı: rezervde tutulan toplam stok değil, açık satış emirleri.
		DrawText(screen, itoa(gs.OpenMarketSupply(good)), float64(colX[4]), float64(ry)+3, FaceSmall, color.RGBA{180, 180, 220, 255})

		ry += 24
	}

	// Alt bilgi
	DrawText(screen, "Not: Fiyatlar her tur sonu güncellenir.", float64(px)+10, float64(ry)+10, FaceSmall, ColorGray)
}

func tradeSelectableGoods() []economy.GoodType {
	return economy.TradeGoods()
}

func isTradeSellerForPlayer(supply int) bool {
	return supply > 0
}

func isTradeBuyerForPlayer(playerStock, targetStock, price, targetGold int) bool {
	if playerStock <= 0 || price <= 0 || targetGold < price {
		return false
	}
	return targetStock < playerStock
}

func isTradeBuyerWithDemand(playerStock, demand, price, targetGold int) bool {
	if playerStock <= 0 || demand <= 0 || price <= 0 || targetGold < price {
		return false
	}
	return true
}

// factionDisplayName bir fraksiyon ID'sinin görünen adını döner.
func factionDisplayName(gs *state.GameState, fid string) string {
	f := gs.Factions[faction.FactionID(fid)]
	if f == nil {
		return fid
	}
	if f.NameTR != "" {
		return f.NameTR
	}
	return f.Name
}

type tradeAgreementCandidate struct {
	ID                      faction.FactionID
	Stance                  faction.DiplomaticStance
	Score                   int
	PlayerTradeCap          int
	TargetTradeCap          int
	PlayerRouteCapacityUsed int
	TargetRouteCapacityUsed int
	PlayerPartners          int
	TargetPartners          int
	Eligible                bool
	BlockReason             string
}

func sortedFactionsForTradeAgreements(gs *state.GameState) []tradeAgreementCandidate {
	if gs == nil {
		return nil
	}
	playerCap := tradeCapacityForFaction(gs, gs.PlayerFactionID)
	playerPartners := diplomacy.ActiveTradePartnerCount(gs, gs.PlayerFactionID)
	playerRouteCapacityUsed := diplomacy.TradeRouteCapacityUsage(gs, gs.PlayerFactionID)
	list := make([]tradeAgreementCandidate, 0, len(gs.Factions))
	for fid, f := range gs.Factions {
		if fid == gs.PlayerFactionID || f == nil || f.IsEliminated || f.IsVirtual {
			continue
		}
		rel := relationForTrade(gs, gs.PlayerFactionID, fid)
		if rel == nil || rel.Stance == faction.StanceWar {
			continue
		}
		if diplomacy.HasTradeRouteBetween(gs, gs.PlayerFactionID, fid) {
			continue
		}
		targetCap := tradeCapacityForFaction(gs, fid)
		targetPartners := diplomacy.ActiveTradePartnerCount(gs, fid)
		targetRouteCapacityUsed := diplomacy.TradeRouteCapacityUsage(gs, fid)
		reason := tradeAgreementBlockReason(gs, rel, gs.PlayerFactionID, fid)
		list = append(list, tradeAgreementCandidate{
			ID:                      fid,
			Stance:                  rel.Stance,
			Score:                   rel.Score,
			PlayerTradeCap:          playerCap,
			TargetTradeCap:          targetCap,
			PlayerRouteCapacityUsed: playerRouteCapacityUsed,
			TargetRouteCapacityUsed: targetRouteCapacityUsed,
			PlayerPartners:          playerPartners,
			TargetPartners:          targetPartners,
			Eligible:                reason == "",
			BlockReason:             reason,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Eligible != list[j].Eligible {
			return list[i].Eligible
		}
		if list[i].Score != list[j].Score {
			return list[i].Score > list[j].Score
		}
		return list[i].ID < list[j].ID
	})
	return list
}

// sortedFactionsForMarket açık pazardaki savaş dışı partnerleri filtreler/sıralar.
func sortedFactionsForMarket(gs *state.GameState, focusGood int, listFilter TradeListFilter, listSort TradeListSort) []faction.FactionID {
	goods := tradeSelectableGoods()
	if focusGood < 0 || focusGood >= len(goods) {
		focusGood = 0
	}
	selectedGood := goods[focusGood]
	player := gs.Factions[gs.PlayerFactionID]
	playerStock := getFactionGoodAmount(player, selectedGood)

	type candidate struct {
		id     faction.FactionID
		stock  int
		supply int
		demand int
		price  int
	}
	list := make([]candidate, 0, len(gs.Factions))
	for fid, f := range gs.Factions {
		if fid == gs.PlayerFactionID || f == nil || f.IsEliminated || f.IsVirtual {
			continue
		}
		if diplomacy.IsWar(gs, gs.PlayerFactionID, fid) {
			continue
		}
		stock := getFactionGoodAmount(f, selectedGood)
		price := tradeMarketPrice(gs, selectedGood)
		supply := gs.MarketSellOffer(fid, selectedGood)
		demand := gs.MarketBuyOrder(fid, selectedGood, price)
		if listFilter == TradeListSellers && !isTradeSellerForPlayer(supply) {
			continue
		}
		if listFilter == TradeListBuyers {
			if !isTradeBuyerWithDemand(playerStock, demand, price, f.Gold) {
				continue
			}
		}
		list = append(list, candidate{id: fid, stock: stock, supply: supply, demand: demand, price: price})
	}
	sort.Slice(list, func(i, j int) bool {
		switch listSort {
		case TradeSortPrice:
			if list[i].price != list[j].price {
				return list[i].price < list[j].price
			}
		case TradeSortStock:
			if list[i].stock != list[j].stock {
				return list[i].stock > list[j].stock
			}
		}
		return list[i].id < list[j].id
	})
	fids := make([]faction.FactionID, 0, len(list))
	for _, c := range list {
		fids = append(fids, c.id)
	}
	return fids
}

func relationForTrade(gs *state.GameState, a, b faction.FactionID) *faction.Relation {
	if gs == nil {
		return nil
	}
	if rel, ok := gs.Relations[faction.RelationKey(a, b)]; ok {
		return rel
	}
	return nil
}

func tradeAgreementBlockReason(gs *state.GameState, rel *faction.Relation, actor, target faction.FactionID) string {
	return diplomacy.AssessTradeProposal(gs, rel, actor, target).BlockReason
}

func tradeCapacityForFaction(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil {
		return 0
	}
	return gs.EffectiveFactionTradeCapacity(fid)
}

func buildTradeAutoExportButton(layout tradeLayout, enabled bool) gameui.Button {
	r := layout.autoExportRect
	label := "Oto. İhracat: KAPALI"
	if enabled {
		label = "Oto. İhracat: AÇIK"
	}
	return gameui.NewButton(r.X, r.Y, r.W, r.H, label)
}

func drawTradeListControls(screen *ebiten.Image, layout tradeLayout, gs *state.GameState, listFilter TradeListFilter, listSort TradeListSort) {
	drawUISectionLabel(screen, layout.filterLabelRect.X, layout.filterLabelRect.Y+12, "Filtre:")
	for _, btn := range buildTradeFilterButtons(layout) {
		drawTradeChoiceButton(screen, btn.Button, int(listFilter) == btn.Value, color.RGBA{70, 62, 36, 235})
	}
	drawUISectionLabel(screen, layout.sortLabelRect.X, layout.sortLabelRect.Y+12, "Sıra:")
	for _, btn := range buildTradeSortButtons(layout) {
		drawTradeChoiceButton(screen, btn.Button, int(listSort) == btn.Value, color.RGBA{52, 70, 82, 235})
	}
	autoBtn := buildTradeAutoExportButton(layout, gs != nil && gs.AutoGrainExport)
	drawTradeChoiceButton(screen, autoBtn, gs != nil && gs.AutoGrainExport, color.RGBA{92, 70, 34, 235})
}

func drawTradeGoodFilterControls(screen *ebiten.Image, layout tradeLayout, focusGood int) {
	drawUISectionLabel(screen, layout.goodFilterLabelRect.X, layout.goodFilterLabelRect.Y+12, "Mal:")
	for _, btn := range buildTradeGoodFilterButtons(layout) {
		drawTradeChoiceButton(screen, btn.Button, focusGood == btn.Value, color.RGBA{70, 62, 36, 235})
	}
}

// getFactionGoodAmount bir fraksiyonun belirli bir maldan kaç adet olduğunu döner.
func getFactionGoodAmount(f *faction.Faction, good economy.GoodType) int {
	if kind, ok := economy.GoodToResourceKind(good); ok {
		return economy.FactionResourceAmount(f, kind)
	}
	return 0
}

func tradeMarketPrice(gs *state.GameState, good economy.GoodType) int {
	if gs != nil && gs.MarketPrices != nil && gs.MarketPrices[good] > 0 {
		return gs.MarketPrices[good]
	}
	return economy.BaseGoldValue[good]
}

func tradeMaxAffordableAmount(player *faction.Faction, price int) int {
	if player == nil || price <= 0 || player.Gold <= 0 {
		return 0
	}
	return player.Gold / price
}

func clampTradeAmountToGold(amount int, player *faction.Faction, price int) int {
	if amount < 1 {
		amount = 1
	}
	if maxAffordable := tradeMaxAffordableAmount(player, price); maxAffordable > 0 && amount > maxAffordable {
		return maxAffordable
	}
	return amount
}

func tradeMaxBuyAmount(gs *state.GameState, player, target *faction.Faction, good economy.GoodType, price int) int {
	if gs == nil || player == nil || target == nil || price <= 0 {
		return 0
	}
	maxByGold := tradeMaxAffordableAmount(player, price)
	targetOffer := gs.MarketSellOffer(target.ID, good)
	if maxByGold < targetOffer {
		return maxByGold
	}
	return targetOffer
}

func tradeMaxSellAmount(gs *state.GameState, player, target *faction.Faction, good economy.GoodType, price int) int {
	if gs == nil || player == nil || target == nil || price <= 0 {
		return 0
	}
	playerStock := getFactionGoodAmount(player, good)
	maxByBuyerDemand := gs.MarketBuyOrder(target.ID, good, price)
	if maxByBuyerDemand < playerStock {
		playerStock = maxByBuyerDemand
	}
	if good == economy.GoodGrain {
		maxBySaleBudget := gs.GrainSaleGoldBudget(gs.PlayerFactionID) / price
		if maxBySaleBudget < playerStock {
			playerStock = maxBySaleBudget
		}
	}
	return playerStock
}

func minTradeInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// tradePanelPointerHit panel açıkken hangi bölgelerde pointer imleci gösterileceğini belirler.
func tradePanelPointerHit(mx, my float64, gs *state.GameState, tab TradeTab, focusFaction int, focusGood int, scroll int, amount int, listFilter TradeListFilter, listSort TradeListSort) bool {
	if buildTradeCloseButton().HitTest(mx, my) {
		return true
	}
	layout := tradePanelLayout()
	px, py, pw, ph := float32(layout.panelRect.X), float32(layout.panelRect.Y), float32(layout.panelRect.W), float32(layout.panelRect.H)
	if mx < float64(px) || mx > float64(px+pw) || my < float64(py) || my > float64(py+ph) {
		return false
	}
	for _, btn := range buildTradeTabButtons() {
		if btn.Button.HitTest(mx, my) {
			return true
		}
	}
	if tab == TradeTabRoutes {
		for _, btn := range buildTradeRouteFilterButtons(layout) {
			if btn.Button.HitTest(mx, my) {
				return true
			}
		}
		return false
	}
	if tab == TradeTabNew {
		candidates := sortedFactionsForTradeAgreements(gs)
		visibleRows := int(layout.listH / 28)
		if visibleRows < 1 {
			visibleRows = 1
		}
		items := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			label := string(candidate.ID)
			if f := gs.Factions[candidate.ID]; f != nil && f.NameTR != "" {
				label = f.NameTR
			}
			items = append(items, label)
		}
		if buildTradeFactionList(layout, visibleRows, items, scroll, focusFaction).HitTest(mx, my) {
			return true
		}
		if focusFaction >= 0 && focusFaction < len(candidates) && buildTradeAgreementButton(layout, candidates[focusFaction].Eligible).HitTest(mx, my) {
			return true
		}
		return false
	}
	if tab != TradeTabMarket {
		return false
	}
	for _, btn := range buildTradeFilterButtons(layout) {
		if btn.Button.HitTest(mx, my) {
			return true
		}
	}
	for _, btn := range buildTradeSortButtons(layout) {
		if btn.Button.HitTest(mx, my) {
			return true
		}
	}
	for _, btn := range buildTradeGoodFilterButtons(layout) {
		if btn.Button.HitTest(mx, my) {
			return true
		}
	}
	if buildTradeAutoExportButton(layout, gs != nil && gs.AutoGrainExport).HitTest(mx, my) {
		return true
	}
	factions := sortedFactionsForMarket(gs, focusGood, listFilter, listSort)
	visibleRows := int(layout.marketListH / 28)
	if visibleRows < 1 {
		visibleRows = 1
	}
	factionItems := make([]string, 0, len(factions))
	goods := tradeSelectableGoods()
	selectedGood := goods[0]
	if focusGood >= 0 && focusGood < len(goods) {
		selectedGood = goods[focusGood]
	}
	price := tradeMarketPrice(gs, selectedGood)
	for _, fid := range factions {
		if f := gs.Factions[fid]; f != nil {
			line := f.NameTR + " | Stok: " + itoa(getFactionGoodAmount(f, selectedGood)) + " | Satış arzı: " + itoa(gs.MarketSellOffer(fid, selectedGood)) + " | Alım talebi: " + itoa(gs.MarketBuyOrder(fid, selectedGood, price)) + " | Fiyat: " + itoa(price) + " altın"
			factionItems = append(factionItems, trimTextToWidth(line, FaceSmall, float64(layout.marketListRect.W)-12))
		}
	}
	if buildTradeMarketFactionList(layout, visibleRows, factionItems, scroll, focusFaction).HitTest(mx, my) {
		return true
	}
	player := gs.Factions[gs.PlayerFactionID]
	amount = clampTradeAmountToGold(amount, player, price)
	qtyButtons, buyBtn, sellBtn := buildTradeMarketActionButtons(layout, amount, player, price)
	for _, btn := range qtyButtons {
		if btn.Enabled && btn.HitTest(mx, my) {
			return true
		}
	}
	emergencyBtn := buildTradeEmergencyGrainSaleButtonAt(tradeMarketActionCardRect, layout, focusGood == 0 && gs.EmergencyGrainSaleLimit() > 0)
	return buyBtn.HitTest(mx, my) || sellBtn.HitTest(mx, my) || emergencyBtn.HitTest(mx, my)
}

func handleTradePanelInput(r *Renderer, input gameui.InputState) InputAction {
	if input.LeftJustPressed {
		px, py, pw, ph := tradePanelRect()
		if !(gameui.Rect{X: float64(px), Y: float64(py), W: float64(pw), H: float64(ph)}).Hit(input.MouseX, input.MouseY) {
			r.showTrade = false
			return InputAction{}
		}
	}
	if buildTradeCloseButton().HandleInput(input) {
		r.showTrade = false
		return InputAction{}
	}
	for _, btn := range buildTradeTabButtons() {
		if btn.Button.HandleInput(input) {
			r.tradeTab = btn.Tab
			r.tradeScroll = 0
			r.tradeFactionFocus = 0
			return InputAction{}
		}
	}
	if r.tradeTab == TradeTabRoutes {
		layout := tradePanelLayout()
		for _, btn := range buildTradeRouteFilterButtons(layout) {
			if btn.Button.HandleInput(input) {
				r.tradeRouteFilter = TradeRouteListFilter(btn.Value)
				r.tradeScroll = 0
				return InputAction{}
			}
		}
		if input.WheelY != 0 {
			if input.WheelY > 0 {
				r.tradeScroll--
			} else {
				r.tradeScroll++
			}
			if r.tradeScroll < 0 {
				r.tradeScroll = 0
			}
		}
		return InputAction{}
	}
	if r.tradeTab == TradeTabNew {
		layout := tradePanelLayout()
		visibleRows := int(layout.listH / 28)
		if visibleRows < 1 {
			visibleRows = 1
		}
		candidates := sortedFactionsForTradeAgreements(r.gs)
		items := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			if f := r.gs.Factions[candidate.ID]; f != nil {
				items = append(items, f.NameTR)
			}
		}
		factionList := buildTradeFactionList(layout, visibleRows, items, r.tradeScroll, r.tradeFactionFocus)
		if idx, ok := tradeListClickedIndex(factionList, input); ok {
			r.tradeFactionFocus = idx
			return InputAction{}
		}
		if factionList.HandleInput(input) {
			r.tradeScroll = factionList.Scroll
			if factionList.Selected >= 0 {
				r.tradeFactionFocus = factionList.Selected
			}
			return InputAction{}
		}
		if r.tradeFactionFocus < 0 {
			r.tradeFactionFocus = 0
		}
		if len(candidates) == 0 || r.tradeFactionFocus >= len(candidates) {
			return InputAction{}
		}
		selected := candidates[r.tradeFactionFocus]
		if buildTradeAgreementButton(layout, selected.Eligible).HandleInput(input) && selected.Eligible {
			return InputAction{Kind: ActionCreateTradeRoute, TargetFaction: selected.ID}
		}
		return InputAction{}
	}
	if r.tradeTab != TradeTabMarket {
		return InputAction{}
	}
	layout := tradePanelLayout()
	visibleRows := int(layout.marketListH / 28)
	if visibleRows < 1 {
		visibleRows = 1
	}
	for _, btn := range buildTradeFilterButtons(layout) {
		if btn.Button.HandleInput(input) {
			r.tradeListFilter = TradeListFilter(btn.Value)
			r.tradeFactionFocus = 0
			r.tradeScroll = 0
			return InputAction{}
		}
	}
	for _, btn := range buildTradeSortButtons(layout) {
		if btn.Button.HandleInput(input) {
			r.tradeListSort = TradeListSort(btn.Value)
			r.tradeFactionFocus = 0
			return InputAction{}
		}
	}
	for _, btn := range buildTradeGoodFilterButtons(layout) {
		if btn.Button.HandleInput(input) {
			r.tradeGoodFocus = btn.Value
			r.tradeFactionFocus = 0
			r.tradeScroll = 0
			return InputAction{}
		}
	}
	if buildTradeAutoExportButton(layout, r.gs != nil && r.gs.AutoGrainExport).HandleInput(input) {
		return InputAction{Kind: ActionToggleAutoGrainExport}
	}
	factions := sortedFactionsForMarket(r.gs, r.tradeGoodFocus, r.tradeListFilter, r.tradeListSort)
	factionItems := make([]string, 0, len(factions))
	goods := tradeSelectableGoods()
	selectedGood := goods[0]
	if r.tradeGoodFocus >= 0 && r.tradeGoodFocus < len(goods) {
		selectedGood = goods[r.tradeGoodFocus]
	}
	price := tradeMarketPrice(r.gs, selectedGood)
	player := r.gs.Factions[r.gs.PlayerFactionID]
	r.tradeAmount = clampTradeAmountToGold(r.tradeAmount, player, price)
	for _, fid := range factions {
		if f := r.gs.Factions[fid]; f != nil {
			line := f.NameTR + " | Stok: " + itoa(getFactionGoodAmount(f, selectedGood)) + " | Satış arzı: " + itoa(r.gs.MarketSellOffer(fid, selectedGood)) + " | Alım talebi: " + itoa(r.gs.MarketBuyOrder(fid, selectedGood, price)) + " | Fiyat: " + itoa(price) + " altın"
			factionItems = append(factionItems, trimTextToWidth(line, FaceSmall, float64(layout.marketListRect.W)-12))
		}
	}
	factionList := buildTradeMarketFactionList(layout, visibleRows, factionItems, r.tradeScroll, r.tradeFactionFocus)
	if idx, ok := tradeListClickedIndex(factionList, input); ok {
		r.tradeFactionFocus = idx
		return InputAction{}
	}
	if factionList.HandleInput(input) {
		r.tradeScroll = factionList.Scroll
		if factionList.Selected != r.tradeFactionFocus && factionList.Selected >= 0 {
			r.tradeFactionFocus = factionList.Selected
		}
		return InputAction{}
	}
	if len(factions) == 0 {
		emergencyBtn := buildTradeEmergencyGrainSaleButtonAt(tradeMarketActionCardRect, layout, r.gs.EmergencyGrainSaleLimit() > 0)
		if emergencyBtn.HandleInput(input) && emergencyBtn.Enabled {
			return InputAction{Kind: ActionEmergencyGrainSale, Delta: r.tradeAmount}
		}
		return InputAction{}
	}
	if r.tradeFactionFocus < 0 {
		r.tradeFactionFocus = 0
	}
	if r.tradeFactionFocus >= len(factions) {
		r.tradeFactionFocus = len(factions) - 1
	}
	qtyButtons, buyBtn, sellBtn := buildTradeMarketActionButtons(layout, r.tradeAmount, player, price)
	deltas := []int{-10, 10}
	for i, btn := range qtyButtons {
		if btn.HandleInput(input) {
			r.tradeAmount += deltas[i]
			if r.tradeAmount < 1 {
				r.tradeAmount = 1
			}
			r.tradeAmount = clampTradeAmountToGold(r.tradeAmount, player, price)
			if r.tradeAmount > 999 {
				r.tradeAmount = 999
			}
			return InputAction{}
		}
	}
	if r.tradeGoodFocus >= 0 && r.tradeGoodFocus < len(goods) && goods[r.tradeGoodFocus] == economy.GoodGrain {
		emergencyBtn := buildTradeEmergencyGrainSaleButtonAt(tradeMarketActionCardRect, layout, r.gs.EmergencyGrainSaleLimit() > 0)
		if emergencyBtn.HandleInput(input) && emergencyBtn.Enabled {
			return InputAction{Kind: ActionEmergencyGrainSale, Delta: r.tradeAmount}
		}
	}
	buyClicked := buyBtn.HandleInput(input)
	sellClicked := sellBtn.HandleInput(input)
	if buyClicked || sellClicked {
		delta := r.tradeAmount
		if sellClicked {
			delta = -delta
		}
		if r.tradeGoodFocus >= 0 && r.tradeGoodFocus < len(goods) {
			return InputAction{
				Kind:          ActionOneTimeTrade,
				TargetFaction: factions[r.tradeFactionFocus],
				BuildingID:    string(goods[r.tradeGoodFocus]),
				Delta:         delta,
			}
		}
	}
	return InputAction{}
}
