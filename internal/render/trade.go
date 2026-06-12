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
	TradeTabMarket                 // aktif ağ içinde manuel pazar işlemi
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

const (
	tradePanelMinW = float32(920)
	tradePanelMaxW = float32(1560)
	tradePanelMinH = float32(620)
	tradePanelMaxH = float32(960)
	tradeTabH      = float32(44)
	tradeRowH      = float32(40)
	tradeActBtnH   = float32(34)
)

const tradeBottomReserved = float32(170)

type tradeLayout struct {
	panelRect       gameui.Rect
	titleRect       gameui.Rect
	closeRect       gameui.Rect
	tabRects        []gameui.Rect
	filterLabelRect gameui.Rect
	filterBtnRects  []gameui.Rect
	sortLabelRect   gameui.Rect
	sortBtnRects    []gameui.Rect
	leftTitleRect   gameui.Rect
	leftListRect    gameui.Rect
	rightTitleRect  gameui.Rect
	rightListRect   gameui.Rect
	actionCardRect  gameui.Rect
	listH           float32
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
	bodyCols := box.SplitColumns(28, 0.40, 0.60)
	leftBox := gameui.BoxFromRect(bodyCols[0])
	rightBox := gameui.BoxFromRect(bodyCols[1])
	leftTitleRect, leftBox := leftBox.CutTop(24, 8)
	actionCardRect, rightBox := rightBox.CutBottom(138, 18)
	rightTitleRect, rightBox := rightBox.CutTop(24, 8)
	listH := float32(leftBox.Rect.H)
	if listH < 120 {
		listH = 120
	}
	controlCols := gameui.BoxFromRect(controlRow).SplitColumns(24, 1, 1)
	filterLabelRect, filterBtnBox := gameui.BoxFromRect(controlCols[0]).CutLeft(76, 12)
	sortLabelRect, sortBtnBox := gameui.BoxFromRect(controlCols[1]).CutLeft(52, 12)

	return tradeLayout{
		panelRect:       panelRect,
		titleRect:       titleBox.Rect,
		closeRect:       closeRect,
		tabRects:        gameui.BoxFromRect(tabsRow).SplitColumns(12, 1, 1, 1, 1),
		filterLabelRect: filterLabelRect,
		filterBtnRects:  filterBtnBox.SplitColumns(12, 1, 1, 1),
		sortLabelRect:   sortLabelRect,
		sortBtnRects:    sortBtnBox.SplitColumns(12, 1, 1, 1),
		leftTitleRect:   leftTitleRect,
		leftListRect:    leftBox.Rect,
		rightTitleRect:  rightTitleRect,
		rightListRect:   rightBox.Rect,
		actionCardRect:  actionCardRect,
		listH:           listH,
	}
}

func tradeGoodsListHeight(rows int) float32 {
	if rows < 1 {
		rows = 1
	}
	return float32(rows * 28)
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
func DrawTradePanel(screen *ebiten.Image, gs *state.GameState, tab TradeTab, focusFaction int, focusGood int, scroll int, amount int, listFilter TradeListFilter, listSort TradeListSort) {
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
		drawTradeRoutesTab(screen, gs, px, contentY, pw, contentH, scroll)
	case TradeTabNew:
		drawTradeNewTab(screen, gs, layout, focusFaction, scroll)
	case TradeTabMarket:
		drawTradeMarketTab(screen, gs, layout, focusFaction, focusGood, scroll, amount, listFilter, listSort)
	case TradeTabPrices:
		drawTradePricesTab(screen, gs, px, contentY, pw, contentH)
	}
}

// drawTradeRoutesTab mevcut aktif ticaret rotalarını listeler.
func drawTradeRoutesTab(screen *ebiten.Image, gs *state.GameState, px float32, y float32, w float32, h float32, scroll int) {
	routes := gs.TradeRoutes
	if len(routes) == 0 {
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

	// Sil butonu başlığı
	DrawText(screen, "İptal", float64(px+w-50), float64(y)+4, FaceSmall, ColorRed)

	rowY := y + 24
	visibleRows := int((h - 30) / tradeRowH)
	if visibleRows < 1 {
		visibleRows = 1
	}
	start := scroll
	end := start + visibleRows
	if end > len(routes) {
		end = len(routes)
	}
	for i := start; i < end; i++ {
		tr := routes[i]
		ry := rowY + float32(i-start)*tradeRowH

		bg := color.RGBA{20, 18, 12, 200}
		if i%2 == 0 {
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
	}

	// Scroll bilgisi
	if len(routes) > visibleRows {
		info := "Rotalar: " + itoa(start+1) + "-" + itoa(end) + "/" + itoa(len(routes))
		DrawText(screen, info, float64(px)+10, float64(y+h-16), FaceSmall, ColorGray)
	}
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
	start := scroll
	end := start + visibleRows
	if end > len(candidates) {
		end = len(candidates)
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
	gameui.DrawListView(screen, factionList, tradeListViewStyle(), renderText)
	if len(candidates) > visibleRows {
		info := "Adaylar: " + itoa(start+1) + "-" + itoa(end) + "/" + itoa(len(candidates))
		drawUIMutedText(screen, layout.leftListRect.X, layout.leftListRect.Y+layout.leftListRect.H+10, info)
	}

	drawUISectionLabel(screen, layout.rightTitleRect.X, layout.rightTitleRect.Y+2, "Anlaşma Özeti:")
	if focusFaction < 0 || focusFaction >= len(candidates) {
		drawUIMutedText(screen, layout.rightListRect.X+6, layout.rightListRect.Y+10, "Önce sol listeden bir hedef fraksiyon seçin.")
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
		"Ticaret kapasitesi: Sen " + itoa(selected.PlayerTradeCap) + " / Onlar " + itoa(selected.TargetTradeCap),
		"Aktif partner: Sen " + itoa(selected.PlayerPartners) + "/4 | Onlar " + itoa(selected.TargetPartners) + "/4",
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

	px, y, w := float32(layout.panelRect.X), float32(layout.leftTitleRect.Y), float32(layout.panelRect.W)
	factions := sortedFactionsForMarket(gs, focusGood, listFilter, listSort)
	drawTradeListControls(screen, layout, listFilter, listSort)
	if len(factions) == 0 {
		DrawTextCentered(screen, "Aktif ticaret ağına bağlı pazar partneri yok.", float64(px)+float64(w)/2, float64(y)+40, FaceMed, ColorGray)
		DrawTextCentered(screen, "Filtreyi değiştirin ya da Yeni Rota sekmesinden ticaret anlaşması açın.", float64(px)+float64(w)/2, float64(y)+62, FaceSmall, ColorGray)
		return
	}

	drawUISectionLabel(screen, layout.leftTitleRect.X, layout.leftTitleRect.Y+2, "Hedef Fraksiyon:")

	visibleRows := int(layout.listH / 28)
	if visibleRows < 1 {
		visibleRows = 1
	}
	start := scroll
	end := start + visibleRows
	if end > len(factions) {
		end = len(factions)
	}
	factionItems := make([]string, 0, len(factions))
	for _, fid := range factions {
		if f := gs.Factions[fid]; f != nil {
			factionItems = append(factionItems, f.NameTR)
		}
	}
	factionList := buildTradeFactionList(layout, visibleRows, factionItems, scroll, focusFaction)
	gameui.DrawListView(screen, factionList, tradeListViewStyle(), renderText)

	rightX := float32(layout.rightTitleRect.X)
	rightW := float32(layout.rightListRect.W)
	drawUISectionLabel(screen, layout.rightTitleRect.X, layout.rightTitleRect.Y+2, "Mal Seçimi:")

	goods := tradeSelectableGoods()
	if focusFaction >= 0 && focusFaction < len(factions) {
		targetFid := factions[focusFaction]
		targetF := gs.Factions[targetFid]
		if targetF != nil {
			goodItems := make([]string, 0, len(goods))
			for _, good := range goods {
				goodName := economy.GoodNameTR(good)
				srcAmount := getFactionGoodAmount(playerF, good)
				dstAmount := getFactionGoodAmount(targetF, good)
				price := "?"
				if gs.MarketPrices != nil {
					if p, ok := gs.MarketPrices[good]; ok {
						price = itoa(p) + " altın"
					}
				}
				line := goodName + " | Sende: " + itoa(srcAmount) + " | " + targetF.NameTR + ": " + itoa(dstAmount) + " | Fiyat: " + price
				goodItems = append(goodItems, trimTextToWidth(line, FaceSmall, float64(rightW)-12))
			}
			goodsList := buildTradeGoodsList(layout, visibleRows, goodItems, focusGood)
			gameui.DrawListView(screen, goodsList, tradeListViewStyle(), renderText)
		}
	} else {
		drawUIMutedText(screen, float64(rightX)+6, layout.rightListRect.Y+10, "Önce sol listeden bir hedef fraksiyon seçin.")
	}

	if len(factions) > visibleRows {
		info := "Partnerler: " + itoa(start+1) + "-" + itoa(end) + "/" + itoa(len(factions))
		drawUIMutedText(screen, layout.leftListRect.X, layout.leftListRect.Y+layout.leftListRect.H+10, info)
	}

	if focusFaction >= 0 && focusFaction < len(factions) && focusGood >= 0 && focusGood < len(goods) {
		goodsRows := minTradeInt(visibleRows, len(goods))
		cardX, cardY, cardW, cardH := tradeActionCardRect(layout, goodsRows)
		vector.FillRect(screen, cardX, cardY, cardW, cardH, color.RGBA{16, 14, 10, 220}, false)
		vector.StrokeRect(screen, cardX, cardY, cardW, cardH, 1, panelBorder, false)
		qtyButtons, buyBtn, sellBtn := buildTradeActionButtons(layout, goodsRows)
		for _, btn := range qtyButtons {
			drawTradeButton(screen, btn, false)
		}
		drawTradeActionButton(screen, buyBtn, color.RGBA{62, 112, 62, 230}, color.RGBA{150, 200, 150, 230})
		drawTradeActionButton(screen, sellBtn, color.RGBA{112, 76, 52, 230}, color.RGBA{210, 170, 130, 230})

		target := gs.Factions[factions[focusFaction]]
		good := goods[focusGood]
		price := gs.MarketPrices[good]
		maxBuy := tradeMaxBuyAmount(playerF, target, good, price)
		maxSell := tradeMaxSellAmount(playerF, target, good, price)
		totalGold := amount * price

		drawUIInfoBlock(screen, float64(cardX)+12, float64(cardY)+16, []string{
			"Seçili: " + economy.GoodNameTR(good) + " | Hedef: " + target.NameTR,
			"Miktar: " + itoa(amount) + " | Tutar: " + itoa(totalGold) + " altın",
			"Al max: " + itoa(maxBuy) + " | Sat max: " + itoa(maxSell),
		}, []color.Color{
			color.RGBA{200, 190, 170, 220},
			color.RGBA{230, 210, 155, 230},
			color.RGBA{160, 190, 210, 220},
		})
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
	tabLabels := []string{"Mevcut Rotalar", "Yeni Rota", "Pazar", "Piyasa Fiyatları"}
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

func buildTradeSortButtons(layout tradeLayout) []tradeChoiceButton {
	labels := []string{"Yakın", "Fiyat", "Stok"}
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

func buildTradeFactionList(layout tradeLayout, visibleRows int, items []string, scroll int, selected int) gameui.ListView {
	h := layout.leftListRect.H
	list := gameui.NewListView(layout.leftListRect.X, layout.leftListRect.Y, layout.leftListRect.W, h, 28, visibleRows, items)
	list.Scroll = scroll
	list.Selected = selected
	return list
}

func buildTradeGoodsList(layout tradeLayout, visibleRows int, items []string, selected int) gameui.ListView {
	rows := minTradeInt(visibleRows, len(items))
	h := float64(tradeGoodsListHeight(rows))
	list := gameui.NewListView(layout.rightListRect.X, layout.rightListRect.Y, layout.rightListRect.W, h, 28, rows, items)
	list.Selected = selected
	return list
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

func buildTradeActionButtons(layout tradeLayout, goodsRows int) ([]gameui.Button, gameui.Button, gameui.Button) {
	cardX, cardY, cardW, cardH := tradeActionCardRect(layout, goodsRows)
	rightX := cardX + 12
	btnY := cardY + cardH - tradeActBtnH - 18
	qty := []gameui.Button{
		gameui.NewButton(float64(rightX), float64(btnY), 54, float64(tradeActBtnH), "-10"),
		gameui.NewButton(float64(rightX+64), float64(btnY), 54, float64(tradeActBtnH), "-1"),
		gameui.NewButton(float64(rightX+128), float64(btnY), 54, float64(tradeActBtnH), "+1"),
		gameui.NewButton(float64(rightX+192), float64(btnY), 54, float64(tradeActBtnH), "+10"),
	}
	qty[0] = qty[0].WithIcon(gameui.IconMinus)
	qty[1] = qty[1].WithIcon(gameui.IconMinus)
	qty[2] = qty[2].WithIcon(gameui.IconPlus)
	qty[3] = qty[3].WithIcon(gameui.IconPlus)
	buyX := cardX + cardW - 244
	buyBtn := gameui.NewButton(float64(buyX), float64(btnY), 110, float64(tradeActBtnH), "AL").WithIcon(gameui.IconBuy)
	sellBtn := gameui.NewButton(float64(buyX+124), float64(btnY), 110, float64(tradeActBtnH), "SAT").WithIcon(gameui.IconSell)
	return qty, buyBtn, sellBtn
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
	style.TextOffsetY = 7
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
	style.TextOffsetY = 10
	drawUIButtonWidget(screen, btn, style)
}

func drawTradeChoiceButton(screen *ebiten.Image, btn gameui.Button, active bool, activeBG color.RGBA) {
	style := tradeButtonStyle(true)
	style.TextOffsetY = 9
	style.BG = color.RGBA{26, 24, 21, 220}
	if active {
		style.BG = activeBG
	}
	drawUIButtonWidget(screen, btn, style)
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
	headers := []string{"Mal", "Base Fiyat", "Güncel Fiyat", "Değişim", "Toplam Arz"}
	DrawText(screen, "Dinamik Piyasa Fiyatları (Arz-Talep):", float64(px)+10, float64(y)+4, FaceSmall, ColorGold)
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

		// Toplam arz (tüm fraksiyonların stokları)
		totalSupply := totalGoodSupply(gs, good)
		DrawText(screen, itoa(totalSupply), float64(colX[4]), float64(ry)+3, FaceSmall, color.RGBA{180, 180, 220, 255})

		ry += 24
	}

	// Alt bilgi
	DrawText(screen, "Not: Fiyatlar her tur sonu güncellenir.", float64(px)+10, float64(ry)+10, FaceSmall, ColorGray)
}

func tradeSelectableGoods() []economy.GoodType {
	return economy.TradeGoods()
}

func isTradeSellerForPlayer(targetStock int) bool {
	return targetStock > 0
}

func isTradeBuyerForPlayer(playerStock, targetStock, price, targetGold int) bool {
	if playerStock <= 0 || price <= 0 || targetGold < price {
		return false
	}
	return targetStock < playerStock
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
	ID             faction.FactionID
	Stance         faction.DiplomaticStance
	Score          int
	PlayerTradeCap int
	TargetTradeCap int
	PlayerPartners int
	TargetPartners int
	Eligible       bool
	BlockReason    string
}

func sortedFactionsForTradeAgreements(gs *state.GameState) []tradeAgreementCandidate {
	if gs == nil {
		return nil
	}
	playerCap := tradeCapacityForFaction(gs, gs.PlayerFactionID)
	playerPartners := activeTradePartnerCount(gs, gs.PlayerFactionID)
	list := make([]tradeAgreementCandidate, 0, len(gs.Factions))
	for fid, f := range gs.Factions {
		if fid == gs.PlayerFactionID || f == nil || f.IsEliminated {
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
		targetPartners := activeTradePartnerCount(gs, fid)
		reason := tradeAgreementBlockReason(gs, rel, gs.PlayerFactionID, fid, playerCap, targetCap, playerPartners, targetPartners)
		list = append(list, tradeAgreementCandidate{
			ID:             fid,
			Stance:         rel.Stance,
			Score:          rel.Score,
			PlayerTradeCap: playerCap,
			TargetTradeCap: targetCap,
			PlayerPartners: playerPartners,
			TargetPartners: targetPartners,
			Eligible:       reason == "",
			BlockReason:    reason,
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

// sortedFactionsForMarket aktif ticaret ağı üzerinden pazar yapılabilecek partnerleri filtreler/sıralar.
func sortedFactionsForMarket(gs *state.GameState, focusGood int, listFilter TradeListFilter, listSort TradeListSort) []faction.FactionID {
	goods := tradeSelectableGoods()
	if focusGood < 0 || focusGood >= len(goods) {
		focusGood = 0
	}
	selectedGood := goods[focusGood]
	player := gs.Factions[gs.PlayerFactionID]
	playerStock := getFactionGoodAmount(player, selectedGood)
	distances := tradeNetworkDistances(gs, gs.PlayerFactionID)

	type candidate struct {
		id    faction.FactionID
		dist  int
		stock int
		price int
	}
	list := make([]candidate, 0, len(gs.Factions))
	for fid, f := range gs.Factions {
		if fid == gs.PlayerFactionID || f == nil || f.IsEliminated {
			continue
		}
		rel := relationForTrade(gs, gs.PlayerFactionID, fid)
		if rel == nil || !(rel.Stance == faction.StancePeace || rel.Stance == faction.StanceTrade || rel.Stance == faction.StanceAllied) {
			continue
		}
		dist, linked := distances[fid]
		if !linked {
			continue
		}
		stock := getFactionGoodAmount(f, selectedGood)
		price := 0
		if gs.MarketPrices != nil {
			price = gs.MarketPrices[selectedGood]
		}
		if listFilter == TradeListSellers && !isTradeSellerForPlayer(stock) {
			continue
		}
		if listFilter == TradeListBuyers {
			if !isTradeBuyerForPlayer(playerStock, stock, price, f.Gold) {
				continue
			}
		}
		list = append(list, candidate{id: fid, dist: dist, stock: stock, price: price})
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
		default:
			if list[i].dist != list[j].dist {
				return list[i].dist < list[j].dist
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

func tradeAgreementBlockReason(gs *state.GameState, rel *faction.Relation, actor, target faction.FactionID, actorCap, targetCap, actorPartners, targetPartners int) string {
	if rel == nil {
		return "İlişki verisi yok"
	}
	if rel.Score < 10 {
		return "İlişki puanı 10 altı"
	}
	if landRegionCountForFaction(gs, actor) == 0 {
		return "Sende kara bölgesi yok"
	}
	if landRegionCountForFaction(gs, target) == 0 {
		return "Hedefin kara bölgesi yok"
	}
	if actorCap < 4 {
		return "Senin ticaret kapasiten 4 altı"
	}
	if targetCap < 4 {
		return "Hedefin ticaret kapasitesi 4 altı"
	}
	if diplomacy.HasDirectThreat(gs, actor, target) {
		return "Doğrudan sınır tehdidi var"
	}
	if actorPartners >= 4 {
		return "Senin aktif partner sınırın dolu"
	}
	if targetPartners >= 4 {
		return "Hedefin aktif partner sınırı dolu"
	}
	return ""
}

func tradeCapacityForFaction(gs *state.GameState, fid faction.FactionID) int {
	total := 0
	if gs == nil {
		return total
	}
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		total += region.TradeCapacity
	}
	return total
}

func landRegionCountForFaction(gs *state.GameState, fid faction.FactionID) int {
	count := 0
	if gs == nil {
		return count
	}
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		count++
	}
	return count
}

func activeTradePartnerCount(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || len(gs.TradeRoutes) == 0 || fid == "" {
		return 0
	}
	partners := make(map[string]struct{})
	self := string(fid)
	for _, route := range gs.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 {
			continue
		}
		switch {
		case route.FromFactionID == self && route.ToFactionID != "":
			partners[route.ToFactionID] = struct{}{}
		case route.ToFactionID == self && route.FromFactionID != "":
			partners[route.FromFactionID] = struct{}{}
		}
	}
	return len(partners)
}

func tradeNetworkDistances(gs *state.GameState, src faction.FactionID) map[faction.FactionID]int {
	dist := map[faction.FactionID]int{src: 0}
	if gs == nil || len(gs.TradeRoutes) == 0 {
		return dist
	}
	adj := make(map[faction.FactionID][]faction.FactionID)
	for _, tr := range gs.TradeRoutes {
		if tr == nil || tr.SuspendedTurns > 0 {
			continue
		}
		a := faction.FactionID(tr.FromFactionID)
		b := faction.FactionID(tr.ToFactionID)
		adj[a] = append(adj[a], b)
		adj[b] = append(adj[b], a)
	}
	q := []faction.FactionID{src}
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		for _, nxt := range adj[cur] {
			if _, seen := dist[nxt]; seen {
				continue
			}
			dist[nxt] = dist[cur] + 1
			q = append(q, nxt)
		}
	}
	return dist
}

func drawTradeListControls(screen *ebiten.Image, layout tradeLayout, listFilter TradeListFilter, listSort TradeListSort) {
	drawUISectionLabel(screen, layout.filterLabelRect.X, layout.filterLabelRect.Y+12, "Filtre:")
	for _, btn := range buildTradeFilterButtons(layout) {
		drawTradeChoiceButton(screen, btn.Button, int(listFilter) == btn.Value, color.RGBA{70, 62, 36, 235})
	}
	drawUISectionLabel(screen, layout.sortLabelRect.X, layout.sortLabelRect.Y+12, "Sıra:")
	for _, btn := range buildTradeSortButtons(layout) {
		drawTradeChoiceButton(screen, btn.Button, int(listSort) == btn.Value, color.RGBA{52, 70, 82, 235})
	}
}

// getFactionGoodAmount bir fraksiyonun belirli bir maldan kaç adet olduğunu döner.
func getFactionGoodAmount(f *faction.Faction, good economy.GoodType) int {
	if kind, ok := economy.GoodToResourceKind(good); ok {
		return economy.FactionResourceAmount(f, kind)
	}
	return 0
}

// totalGoodSupply tüm aktif fraksiyonların belirli bir maldan toplam stokunu döner.
func totalGoodSupply(gs *state.GameState, good economy.GoodType) int {
	total := 0
	for _, f := range gs.Factions {
		if f == nil || f.IsEliminated {
			continue
		}
		total += getFactionGoodAmount(f, good)
	}
	return total
}

func drawTradeQtyButtons(screen *ebiten.Image, x, y float32) {
	labels := []string{"-10", "-1", "+1", "+10"}
	for i, label := range labels {
		bx := x + float32(i)*44
		vector.FillRect(screen, bx, y, 36, tradeActBtnH, color.RGBA{38, 32, 24, 230}, false)
		vector.StrokeRect(screen, bx, y, 36, tradeActBtnH, 1, panelBorder, false)
		DrawTextCentered(screen, label, float64(bx)+18, float64(y)+6, FaceSmall, ColorWhite)
	}
}

func tradeMaxBuyAmount(player, target *faction.Faction, good economy.GoodType, price int) int {
	if player == nil || target == nil || price <= 0 {
		return 0
	}
	maxByGold := player.Gold / price
	targetStock := getFactionGoodAmount(target, good)
	if maxByGold < targetStock {
		return maxByGold
	}
	return targetStock
}

func tradeMaxSellAmount(player, target *faction.Faction, good economy.GoodType, price int) int {
	if player == nil || target == nil || price <= 0 {
		return 0
	}
	maxByBuyerGold := target.Gold / price
	playerStock := getFactionGoodAmount(player, good)
	if maxByBuyerGold < playerStock {
		return maxByBuyerGold
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
func tradePanelPointerHit(mx, my float64, gs *state.GameState, tab TradeTab, focusFaction int, focusGood int, scroll int, listFilter TradeListFilter, listSort TradeListSort) bool {
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
	factions := sortedFactionsForMarket(gs, focusGood, listFilter, listSort)
	visibleRows := int(layout.listH / 28)
	if visibleRows < 1 {
		visibleRows = 1
	}
	factionItems := make([]string, 0, len(factions))
	for _, fid := range factions {
		if f := gs.Factions[fid]; f != nil {
			factionItems = append(factionItems, f.NameTR)
		}
	}
	if buildTradeFactionList(layout, visibleRows, factionItems, scroll, focusFaction).HitTest(mx, my) {
		return true
	}
	if focusFaction >= 0 && focusFaction < len(factions) {
		target := gs.Factions[factions[focusFaction]]
		if target != nil {
			goods := tradeSelectableGoods()
			items := make([]string, 0, len(goods))
			for _, good := range goods {
				price := "?"
				if gs.MarketPrices != nil {
					if p, ok := gs.MarketPrices[good]; ok {
						price = itoa(p) + " altın"
					}
				}
				line := economy.GoodNameTR(good) + " | Sende: " + itoa(getFactionGoodAmount(gs.Factions[gs.PlayerFactionID], good)) + " | " + target.NameTR + ": " + itoa(getFactionGoodAmount(target, good)) + " | Fiyat: " + price
				items = append(items, trimTextToWidth(line, FaceSmall, layout.rightListRect.W-12))
			}
			if buildTradeGoodsList(layout, visibleRows, items, focusGood).HitTest(mx, my) {
				return true
			}
		}
	}
	qtyButtons, buyBtn, sellBtn := buildTradeActionButtons(layout, minTradeInt(visibleRows, len(tradeSelectableGoods())))
	for _, btn := range qtyButtons {
		if btn.HitTest(mx, my) {
			return true
		}
	}
	return buyBtn.HitTest(mx, my) || sellBtn.HitTest(mx, my)
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
	visibleRows := int(layout.listH / 28)
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
	factions := sortedFactionsForMarket(r.gs, r.tradeGoodFocus, r.tradeListFilter, r.tradeListSort)
	factionItems := make([]string, 0, len(factions))
	for _, fid := range factions {
		if f := r.gs.Factions[fid]; f != nil {
			factionItems = append(factionItems, f.NameTR)
		}
	}
	factionList := buildTradeFactionList(layout, visibleRows, factionItems, r.tradeScroll, r.tradeFactionFocus)
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
		return InputAction{}
	}
	if r.tradeFactionFocus < 0 {
		r.tradeFactionFocus = 0
	}
	if r.tradeFactionFocus >= len(factions) {
		r.tradeFactionFocus = len(factions) - 1
	}
	target := r.gs.Factions[factions[r.tradeFactionFocus]]
	goods := tradeSelectableGoods()
	goodItems := make([]string, 0, len(goods))
	for _, good := range goods {
		price := "?"
		if r.gs.MarketPrices != nil {
			if p, ok := r.gs.MarketPrices[good]; ok {
				price = itoa(p) + " altın"
			}
		}
		line := economy.GoodNameTR(good) + " | Sende: " + itoa(getFactionGoodAmount(r.gs.Factions[r.gs.PlayerFactionID], good)) + " | " + target.NameTR + ": " + itoa(getFactionGoodAmount(target, good)) + " | Fiyat: " + price
		goodItems = append(goodItems, trimTextToWidth(line, FaceSmall, layout.rightListRect.W-12))
	}
	goodsList := buildTradeGoodsList(layout, visibleRows, goodItems, r.tradeGoodFocus)
	if idx, ok := tradeListClickedIndex(goodsList, input); ok {
		r.tradeGoodFocus = idx
		return InputAction{}
	}
	if goodsList.HandleInput(input) {
		if goodsList.Selected >= 0 {
			r.tradeGoodFocus = goodsList.Selected
		}
		return InputAction{}
	}
	qtyButtons, buyBtn, sellBtn := buildTradeActionButtons(layout, minTradeInt(visibleRows, len(goods)))
	deltas := []int{-10, -1, 1, 10}
	for i, btn := range qtyButtons {
		if btn.HandleInput(input) {
			r.tradeAmount += deltas[i]
			if r.tradeAmount < 1 {
				r.tradeAmount = 1
			}
			if r.tradeAmount > 999 {
				r.tradeAmount = 999
			}
			return InputAction{}
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
