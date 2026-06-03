package render

import (
	"image/color"
	"sort"

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
	TradeTabNew                    // yeni rota oluştur
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
	tradePanelW   = float32(600)
	tradePanelH   = float32(520)
	tradeStartY   = float32(80)
	tradeTabH     = float32(32)
	tradeRowH     = float32(40)
	tradeGoodBtnW = float32(80)
	tradeGoodBtnH = float32(28)
	tradeActBtnW  = float32(76)
	tradeActBtnH  = float32(24)
)

const tradeBottomReserved = float32(120)

// tradePanelRect ticaret panelinin ortalanmış dikdörtgenini döner.
func tradePanelRect() (x, y, w, h float32) {
	w = tradePanelW
	h = tradePanelH
	x = float32(ScreenWidth)/2 - w/2
	y = float32(ScreenHeight)/2 - h/2
	return x, y, w, h
}

// tradeCloseRect kapatma butonu.
func tradeCloseRect() (x, y, w, h float32) {
	px, py, pw, _ := tradePanelRect()
	w = 30
	h = 26
	x = px + pw - w - 10
	y = py + 10
	return x, y, w, h
}

// tradeCloseHit tıklama kontrolü.
func tradeCloseHit(mx, my float64) bool {
	x, y, w, h := tradeCloseRect()
	return mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h)
}

// DrawTradePanel ticaret panelini çizer.
// Tab 0: mevcut rotalar, Tab 1: yeni rota oluştur, Tab 2: piyasa fiyatları
func DrawTradePanel(screen *ebiten.Image, gs *state.GameState, tab TradeTab, focusFaction int, focusGood int, scroll int, amount int, listFilter TradeListFilter, listSort TradeListSort) {
	px, py, pw, ph := tradePanelRect()

	// Arka plan overlay
	overlay := ebiten.NewImage(int(ScreenWidth), int(ScreenHeight))
	overlay.Fill(color.RGBA{8, 6, 4, 200})
	screen.DrawImage(overlay, nil)

	// Panel çerçevesi
	vector.FillRect(screen, px, py, pw, ph, panelBg, false)
	drawPanelBorder(screen, px, py, pw, ph)
	vector.FillRect(screen, px, py, pw, 3, panelBorder, false)

	// Başlık
	DrawTextCentered(screen, "── Ticaret ──", float64(px)+float64(pw)/2, float64(py)+16, FaceLarge, ColorYellow)

	// Kapatma butonu
	closeBtn := buildTradeCloseButton()
	drawTradeButton(screen, closeBtn, false)

	// Sekmeler
	for _, btn := range buildTradeTabButtons() {
		drawTradeButton(screen, btn.Button, btn.Tab != tab)
	}

	contentY := py + 40 + tradeTabH + 8
	contentH := ph - (contentY - py) - 10

	switch tab {
	case TradeTabRoutes:
		drawTradeRoutesTab(screen, gs, px, contentY, pw, contentH, scroll)
	case TradeTabNew:
		drawTradeNewTab(screen, gs, px, contentY, pw, contentH, focusFaction, focusGood, scroll, amount, listFilter, listSort)
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
	colX := []float32{px + 10, px + w*0.35, px + w*0.55, px + w*0.75}
	headers := []string{"Mal", "Gönderen", "Alan", "Miktar/Tur"}
	colW := []float32{w * 0.25, w * 0.20, w * 0.20, w * 0.15}
	for i, hdr := range headers {
		DrawText(screen, hdr, float64(colX[i]), float64(y)+4, FaceSmall, ColorGold)
		_ = colW[i]
	}

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

		goodName := economy.GoodNameTR(tr.Good)
		DrawText(screen, goodName, float64(colX[0]), float64(ry)+10, FaceSmall, ColorWhite)
		DrawText(screen, factionDisplayName(gs, tr.FromFactionID), float64(colX[1]), float64(ry)+10, FaceSmall, ColorGray)
		DrawText(screen, factionDisplayName(gs, tr.ToFactionID), float64(colX[2]), float64(ry)+10, FaceSmall, ColorGray)
		DrawText(screen, itoa(tr.AmountPerTurn)+" @"+itoa(tr.GoldPerUnit)+" altın", float64(colX[3]), float64(ry)+10, FaceSmall, ColorGold)
	}

	// Scroll bilgisi
	if len(routes) > visibleRows {
		info := "Rotalar: " + itoa(start+1) + "-" + itoa(end) + "/" + itoa(len(routes))
		DrawText(screen, info, float64(px)+10, float64(y+h-16), FaceSmall, ColorGray)
	}
}

// drawTradeNewTab yeni ticaret rotası oluşturma arayüzü.
func drawTradeNewTab(screen *ebiten.Image, gs *state.GameState, px float32, y float32, w float32, h float32, focusFaction int, focusGood int, scroll int, amount int, listFilter TradeListFilter, listSort TradeListSort) {
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

	// Sol sütun: hedef fraksiyon listesi
	leftW := w * 0.40
	factions := sortedFactionsForTrade(gs, focusGood, listFilter, listSort)
	if len(factions) == 0 {
		DrawTextCentered(screen, "Uygun ticaret partneri yok.", float64(px)+float64(w)/2, float64(y)+40, FaceMed, ColorGray)
		DrawTextCentered(screen, "Barış/ticaret anlaşması ve ticaret ağı bağlantısı gerekir.", float64(px)+float64(w)/2, float64(y)+62, FaceSmall, ColorGray)
		return
	}
	drawTradeListControls(screen, px, y, listFilter, listSort)

	// Fraksiyon/mal başlıkları kontrol satırının altına alınır.
	DrawText(screen, "Hedef Fraksiyon:", float64(px)+8, float64(y)+42, FaceSmall, ColorGold)

	visibleRows := int((h - tradeBottomReserved - 62) / 28)
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
	factionList := buildTradeFactionList(px, y, leftW, visibleRows, factionItems, scroll, focusFaction)
	gameui.DrawListView(screen, factionList, tradeListViewStyle(), renderSmallText)

	// Sağ sütun: mal seçimi
	rightX := px + leftW + 8
	rightW := w - leftW - 16
	DrawText(screen, "Mal Seçimi:", float64(rightX), float64(y)+42, FaceSmall, ColorGold)

	goods := tradeSelectableGoods()

	// Hedef fraksiyon seçiliyse mal listesini göster
	if focusFaction >= 0 && focusFaction < len(factions) {
		targetFid := factions[focusFaction]
		targetF := gs.Factions[targetFid]
		if targetF != nil {
			goodItems := make([]string, 0, len(goods))
			for gi, good := range goods {
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
				if gi == focusGood {
					_ = gi
				}
				goodItems = append(goodItems, trimTextToWidth(line, FaceSmall, float64(rightW)-12))
			}
			goodsList := buildTradeGoodsList(px, y, leftW, rightW, visibleRows, goodItems, focusGood)
			gameui.DrawListView(screen, goodsList, tradeListViewStyle(), renderSmallText)
		}
	} else {
		DrawText(screen, "Önce sol listeden bir hedef fraksiyon seçin.", float64(rightX)+6, float64(y)+72, FaceSmall, ColorGray)
	}

	// Scroll info (alt aksiyon alanının üstüne sabitlenir)
	if len(factions) > visibleRows {
		info := "Partnerler: " + itoa(start+1) + "-" + itoa(end) + "/" + itoa(len(factions))
		DrawText(screen, info, float64(px)+10, float64(y+h-tradeBottomReserved+6), FaceSmall, ColorGray)
	}

	// Manuel al/sat butonları (tek seferlik pazar işlemi)
	if focusFaction >= 0 && focusFaction < len(factions) && focusGood >= 0 && focusGood < len(goods) {
		btnY := y + h - tradeActBtnH - 12
		qtyButtons, buyBtn, sellBtn := buildTradeActionButtons(px, y, w, h)
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

		line := "Seçili: " + economy.GoodNameTR(good) + " | Hedef: " + target.NameTR
		DrawText(screen, line, float64(rightX), float64(btnY)-34, FaceSmall, color.RGBA{200, 190, 170, 220})
		line2 := "Miktar: " + itoa(amount) + " | Tutar: " + itoa(totalGold) + " altın"
		DrawText(screen, line2, float64(rightX), float64(btnY)-18, FaceSmall, color.RGBA{230, 210, 155, 230})
		line3 := "Al max: " + itoa(maxBuy) + " | Sat max: " + itoa(maxSell)
		DrawText(screen, line3, float64(rightX), float64(btnY)-2, FaceSmall, color.RGBA{160, 190, 210, 220})
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
	return gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "X")
}

func buildTradeTabButtons() []tradeTabButton {
	px, py, pw, _ := tradePanelRect()
	tabLabels := []string{"Mevcut Rotalar", "Yeni Rota", "Piyasa Fiyatları"}
	tabW := (pw - 16) / float32(len(tabLabels))
	tabY := py + 40
	out := make([]tradeTabButton, 0, len(tabLabels))
	for i, label := range tabLabels {
		tx := px + 8 + float32(i)*tabW
		out = append(out, tradeTabButton{
			Tab:    TradeTab(i),
			Button: gameui.NewButton(float64(tx), float64(tabY), float64(tabW-4), float64(tradeTabH), label),
		})
	}
	return out
}

func buildTradeFilterButtons(px, y float32) []tradeChoiceButton {
	labels := []string{"Hepsi", "Satanlar", "Alanlar"}
	out := make([]tradeChoiceButton, 0, len(labels))
	for i, label := range labels {
		bx := px + 56 + float32(i)*72
		out = append(out, tradeChoiceButton{
			Value:  i,
			Button: gameui.NewButton(float64(bx), float64(y+2), 66, 22, label),
		})
	}
	return out
}

func buildTradeSortButtons(px, y float32) []tradeChoiceButton {
	labels := []string{"Yakın", "Fiyat", "Stok"}
	out := make([]tradeChoiceButton, 0, len(labels))
	for i, label := range labels {
		bx := px + 325 + float32(i)*62
		out = append(out, tradeChoiceButton{
			Value:  i,
			Button: gameui.NewButton(float64(bx), float64(y+2), 56, 22, label),
		})
	}
	return out
}

func buildTradeFactionList(px, y, leftW float32, visibleRows int, items []string, scroll int, selected int) gameui.ListView {
	h := float64(visibleRows) * 28
	list := gameui.NewListView(float64(px+8), float64(y+62), float64(leftW-12), h, 28, visibleRows, items)
	list.Scroll = scroll
	list.Selected = selected
	return list
}

func buildTradeGoodsList(px, y, leftW, rightW float32, visibleRows int, items []string, selected int) gameui.ListView {
	h := float64(minTradeInt(visibleRows, len(items))) * 28
	list := gameui.NewListView(float64(px+leftW+8), float64(y+62), float64(rightW), h, 28, len(items), items)
	list.Selected = selected
	return list
}

func buildTradeActionButtons(px, y, w, h float32) ([]gameui.Button, gameui.Button, gameui.Button) {
	leftW := w * 0.40
	rightX := px + leftW + 8
	btnY := y + h - tradeActBtnH - 12
	qty := []gameui.Button{
		gameui.NewButton(float64(rightX), float64(btnY), 36, float64(tradeActBtnH), "-10"),
		gameui.NewButton(float64(rightX+44), float64(btnY), 36, float64(tradeActBtnH), "-1"),
		gameui.NewButton(float64(rightX+88), float64(btnY), 36, float64(tradeActBtnH), "+1"),
		gameui.NewButton(float64(rightX+132), float64(btnY), 36, float64(tradeActBtnH), "+10"),
	}
	buyX := rightX + 170
	buyBtn := gameui.NewButton(float64(buyX), float64(btnY), float64(tradeActBtnW), float64(tradeActBtnH), "AL")
	sellBtn := gameui.NewButton(float64(buyX+tradeActBtnW+8), float64(btnY), float64(tradeActBtnW), float64(tradeActBtnH), "SAT")
	return qty, buyBtn, sellBtn
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
	}
}

func drawTradeButton(screen *ebiten.Image, btn gameui.Button, inactive bool) {
	drawUIButton(screen, btn.X, btn.Y, btn.W, btn.H, btn.Label, btn.Enabled, tradeButtonStyle(!inactive))
}

func drawTradeActionButton(screen *ebiten.Image, btn gameui.Button, bg, border color.RGBA) {
	style := menuButtonStyle
	style.BG = bg
	style.Border = border
	style.Text = ColorWhite
	style.TextOffsetY = 6
	drawUIButton(screen, btn.X, btn.Y, btn.W, btn.H, btn.Label, btn.Enabled, style)
}

func drawTradeChoiceButton(screen *ebiten.Image, btn gameui.Button, active bool, activeBG color.RGBA) {
	style := tradeButtonStyle(true)
	style.TextOffsetY = 4
	style.BG = color.RGBA{26, 24, 21, 220}
	if active {
		style.BG = activeBG
	}
	drawUIButton(screen, btn.X, btn.Y, btn.W, btn.H, btn.Label, btn.Enabled, style)
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

// sortedFactionsForTrade ticaret yapılabilecek partnerleri filtreler/sıralar.
func sortedFactionsForTrade(gs *state.GameState, focusGood int, listFilter TradeListFilter, listSort TradeListSort) []faction.FactionID {
	goods := tradeSelectableGoods()
	if focusGood < 0 || focusGood >= len(goods) {
		focusGood = 0
	}
	selectedGood := goods[focusGood]
	player := gs.Factions[gs.PlayerFactionID]
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
		if listFilter == TradeListSellers && stock <= 0 {
			continue
		}
		if listFilter == TradeListBuyers {
			playerStock := getFactionGoodAmount(player, selectedGood)
			if price <= 0 || f.Gold/price <= 0 || playerStock <= 0 {
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

func drawTradeListControls(screen *ebiten.Image, px, y float32, listFilter TradeListFilter, listSort TradeListSort) {
	DrawText(screen, "Filtre:", float64(px)+8, float64(y)+14, FaceSmall, ColorGold)
	for _, btn := range buildTradeFilterButtons(px, y) {
		drawTradeChoiceButton(screen, btn.Button, int(listFilter) == btn.Value, color.RGBA{70, 62, 36, 235})
	}
	DrawText(screen, "Sıra:", float64(px)+286, float64(y)+14, FaceSmall, ColorGold)
	for _, btn := range buildTradeSortButtons(px, y) {
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
	px, py, pw, ph := tradePanelRect()
	if mx < float64(px) || mx > float64(px+pw) || my < float64(py) || my > float64(py+ph) {
		return false
	}
	for _, btn := range buildTradeTabButtons() {
		if btn.Button.HitTest(mx, my) {
			return true
		}
	}
	if tab != TradeTabNew {
		return false
	}
	for _, btn := range buildTradeFilterButtons(px, py+tradeTabH+48) {
		if btn.Button.HitTest(mx, my) {
			return true
		}
	}
	for _, btn := range buildTradeSortButtons(px, py+tradeTabH+48) {
		if btn.Button.HitTest(mx, my) {
			return true
		}
	}
	factions := sortedFactionsForTrade(gs, focusGood, listFilter, listSort)
	visibleRows := int((ph - (py + tradeTabH + 48 - py) - tradeBottomReserved - 62) / 28)
	if visibleRows < 1 {
		visibleRows = 1
	}
	factionItems := make([]string, 0, len(factions))
	for _, fid := range factions {
		if f := gs.Factions[fid]; f != nil {
			factionItems = append(factionItems, f.NameTR)
		}
	}
	if buildTradeFactionList(px, py+tradeTabH+48, pw*0.40, visibleRows, factionItems, scroll, focusFaction).HitTest(mx, my) {
		return true
	}
	if focusFaction >= 0 && focusFaction < len(factions) {
		target := gs.Factions[factions[focusFaction]]
		if target != nil {
			goods := tradeSelectableGoods()
			items := make([]string, 0, len(goods))
			rightW := pw - pw*0.40 - 16
			for _, good := range goods {
				price := "?"
				if gs.MarketPrices != nil {
					if p, ok := gs.MarketPrices[good]; ok {
						price = itoa(p) + " altın"
					}
				}
				line := economy.GoodNameTR(good) + " | Sende: " + itoa(getFactionGoodAmount(gs.Factions[gs.PlayerFactionID], good)) + " | " + target.NameTR + ": " + itoa(getFactionGoodAmount(target, good)) + " | Fiyat: " + price
				items = append(items, trimTextToWidth(line, FaceSmall, float64(rightW)-12))
			}
			if buildTradeGoodsList(px, py+tradeTabH+48, pw*0.40, rightW, visibleRows, items, focusGood).HitTest(mx, my) {
				return true
			}
		}
	}
	qtyButtons, buyBtn, sellBtn := buildTradeActionButtons(px, py+tradeTabH+48, pw, ph-(tradeTabH+58))
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
	if r.tradeTab != TradeTabNew {
		return InputAction{}
	}
	px, py, pw, ph := tradePanelRect()
	contentY := py + tradeTabH + 48
	visibleRows := int((ph - (contentY - py) - tradeBottomReserved - 62) / 28)
	if visibleRows < 1 {
		visibleRows = 1
	}
	for _, btn := range buildTradeFilterButtons(px, contentY) {
		if btn.Button.HandleInput(input) {
			r.tradeListFilter = TradeListFilter(btn.Value)
			r.tradeFactionFocus = 0
			r.tradeScroll = 0
			return InputAction{}
		}
	}
	for _, btn := range buildTradeSortButtons(px, contentY) {
		if btn.Button.HandleInput(input) {
			r.tradeListSort = TradeListSort(btn.Value)
			r.tradeFactionFocus = 0
			return InputAction{}
		}
	}
	factions := sortedFactionsForTrade(r.gs, r.tradeGoodFocus, r.tradeListFilter, r.tradeListSort)
	factionItems := make([]string, 0, len(factions))
	for _, fid := range factions {
		if f := r.gs.Factions[fid]; f != nil {
			factionItems = append(factionItems, f.NameTR)
		}
	}
	factionList := buildTradeFactionList(px, contentY, pw*0.40, visibleRows, factionItems, r.tradeScroll, r.tradeFactionFocus)
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
	rightW := pw - pw*0.40 - 16
	for _, good := range goods {
		price := "?"
		if r.gs.MarketPrices != nil {
			if p, ok := r.gs.MarketPrices[good]; ok {
				price = itoa(p) + " altın"
			}
		}
		line := economy.GoodNameTR(good) + " | Sende: " + itoa(getFactionGoodAmount(r.gs.Factions[r.gs.PlayerFactionID], good)) + " | " + target.NameTR + ": " + itoa(getFactionGoodAmount(target, good)) + " | Fiyat: " + price
		goodItems = append(goodItems, trimTextToWidth(line, FaceSmall, float64(rightW)-12))
	}
	goodsList := buildTradeGoodsList(px, contentY, pw*0.40, rightW, visibleRows, goodItems, r.tradeGoodFocus)
	if goodsList.HandleInput(input) {
		if goodsList.Selected >= 0 {
			r.tradeGoodFocus = goodsList.Selected
		}
		return InputAction{}
	}
	qtyButtons, buyBtn, sellBtn := buildTradeActionButtons(px, contentY, pw, ph-(tradeTabH+58))
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
