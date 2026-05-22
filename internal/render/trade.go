package render

import (
	"image/color"
	"sort"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"

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

const (
	tradePanelW   = float32(600)
	tradePanelH   = float32(480)
	tradeStartY   = float32(80)
	tradeTabH     = float32(32)
	tradeRowH     = float32(40)
	tradeGoodBtnW = float32(80)
	tradeGoodBtnH = float32(28)
)

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
func DrawTradePanel(screen *ebiten.Image, gs *state.GameState, tab TradeTab, focusFaction int, focusGood int, scroll int) {
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
	cx, cy, cw, ch := tradeCloseRect()
	vector.FillRect(screen, cx, cy, cw, ch, color.RGBA{45, 34, 25, 230}, false)
	vector.StrokeRect(screen, cx, cy, cw, ch, 1, panelBorder, false)
	DrawTextCentered(screen, "X", float64(cx)+float64(cw)/2, float64(cy)+5, FaceSmall, ColorGold)

	// Sekmeler
	tabLabels := []string{"Mevcut Rotalar", "Yeni Rota", "Piyasa Fiyatları"}
	tabW := (pw - 16) / float32(len(tabLabels))
	tabY := py + 40
	for i, label := range tabLabels {
		tx := px + 8 + float32(i)*tabW
		bg := color.RGBA{25, 20, 14, 200}
		if i == int(tab) {
			bg = color.RGBA{55, 45, 25, 230}
		}
		vector.FillRect(screen, tx, tabY, tabW-4, tradeTabH, bg, false)
		tw := MeasureText(label, FaceSmall)
		DrawText(screen, label, float64(tx)+float64(tabW-4)/2-float64(tw)/2, float64(tabY)+8, FaceSmall, ColorWhite)
	}

	contentY := tabY + tradeTabH + 8
	contentH := ph - (contentY - py) - 10

	switch tab {
	case TradeTabRoutes:
		drawTradeRoutesTab(screen, gs, px, contentY, pw, contentH, scroll)
	case TradeTabNew:
		drawTradeNewTab(screen, gs, px, contentY, pw, contentH, focusFaction, focusGood, scroll)
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

		goodName := goodDisplayName(tr.Good)
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
func drawTradeNewTab(screen *ebiten.Image, gs *state.GameState, px float32, y float32, w float32, h float32, focusFaction int, focusGood int, scroll int) {
	playerF := gs.Factions[gs.PlayerFactionID]
	if playerF == nil {
		return
	}

	// Sol sütun: hedef fraksiyon listesi
	leftW := w * 0.40
	factions := sortedFactionsForTrade(gs)
	if len(factions) == 0 {
		DrawTextCentered(screen, "Ticaret yapılacak fraksiyon yok.", float64(px)+float64(w)/2, float64(y)+40, FaceMed, ColorGray)
		return
	}

	// Fraksiyon listesinin başlığı
	DrawText(screen, "Hedef Fraksiyon:", float64(px)+8, float64(y)+4, FaceSmall, ColorGold)

	rowY := y + 20
	visibleRows := int((h - 30) / 28)
	if visibleRows < 1 {
		visibleRows = 1
	}
	start := scroll
	end := start + visibleRows
	if end > len(factions) {
		end = len(factions)
	}
	for i := start; i < end; i++ {
		fid := factions[i]
		f := gs.Factions[fid]
		if f == nil {
			continue
		}
		ry := rowY + float32(i-start)*28
		bg := color.RGBA{20, 18, 12, 200}
		if i == focusFaction {
			bg = color.RGBA{55, 45, 25, 230}
		}
		vector.FillRect(screen, px+8, ry, leftW-12, 24, bg, false)
		DrawText(screen, f.NameTR, float64(px)+14, float64(ry)+4, FaceSmall, ColorWhite)
	}

	// Sağ sütun: mal seçimi
	rightX := px + leftW + 8
	rightW := w - leftW - 16
	DrawText(screen, "Mal Seçimi:", float64(rightX), float64(y)+4, FaceSmall, ColorGold)

	goods := []economy.GoodType{
		economy.GoodGrain,
		economy.GoodIron,
		economy.GoodTimber,
		economy.GoodSpice,
		economy.GoodCloth,
	}

	// Hedef fraksiyon seçiliyse mal listesini göster
	if focusFaction >= 0 && focusFaction < len(factions) {
		targetFid := factions[focusFaction]
		targetF := gs.Factions[targetFid]
		if targetF != nil {
			gy := y + 20
			for gi, good := range goods {
				goodName := goodDisplayName(good)
				srcAmount := getFactionGoodAmount(playerF, good)
				dstAmount := getFactionGoodAmount(targetF, good)

				bg := color.RGBA{20, 18, 12, 200}
				if gi == focusGood {
					bg = color.RGBA{55, 45, 25, 230}
				}
				vector.FillRect(screen, rightX, gy, rightW, 24, bg, false)

				// Fiyat bilgisi
				price := "?"
				if gs.MarketPrices != nil {
					if p, ok := gs.MarketPrices[good]; ok {
						price = itoa(p) + " altın"
					}
				}
				line := goodName + " | Sende: " + itoa(srcAmount) + " | " + targetF.NameTR + ": " + itoa(dstAmount) + " | Fiyat: " + price
				DrawText(screen, line, float64(rightX)+6, float64(gy)+4, FaceSmall, color.RGBA{220, 210, 185, 240})
				gy += 28
			}
		}
	} else {
		DrawText(screen, "Önce sol listeden bir hedef fraksiyon seçin.", float64(rightX)+6, float64(y)+30, FaceSmall, ColorGray)
	}

	// Scroll info
	if len(factions) > visibleRows {
		info := "Fraksiyonlar: " + itoa(start+1) + "-" + itoa(end) + "/" + itoa(len(factions))
		DrawText(screen, info, float64(px)+10, float64(y+h-16), FaceSmall, ColorGray)
	}
}

// drawTradePricesTab piyasa fiyatlarını gösterir.
func drawTradePricesTab(screen *ebiten.Image, gs *state.GameState, px float32, y float32, w float32, _ float32) {
	if gs.MarketPrices == nil {
		DrawTextCentered(screen, "Piyasa fiyatları henüz oluşturulmadı.", float64(px)+float64(w)/2, float64(y)+40, FaceMed, ColorGray)
		return
	}

	goods := []economy.GoodType{
		economy.GoodGrain,
		economy.GoodIron,
		economy.GoodTimber,
		economy.GoodSpice,
		economy.GoodCloth,
	}

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
		goodName := goodDisplayName(good)

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

// goodDisplayName mal adını Türkçe döner.
func goodDisplayName(good economy.GoodType) string {
	switch good {
	case economy.GoodGrain:
		return "Tahıl"
	case economy.GoodIron:
		return "Demir"
	case economy.GoodTimber:
		return "Kereste"
	case economy.GoodSpice:
		return "Baharat"
	case economy.GoodCloth:
		return "Kumaş"
	default:
		return string(good)
	}
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

// sortedFactionsForTrade ticaret yapılabilecek fraksiyonları sıralar.
// Oyuncu ve elenmiş fraksiyonlar hariç.
func sortedFactionsForTrade(gs *state.GameState) []faction.FactionID {
	var fids []faction.FactionID
	for fid := range gs.Factions {
		if fid == gs.PlayerFactionID {
			continue
		}
		f := gs.Factions[fid]
		if f == nil || f.IsEliminated {
			continue
		}
		// StanceTrade veya StancePeace olanlarla ticaret mümkün
		key := faction.RelationKey(gs.PlayerFactionID, fid)
		if rel, ok := gs.Relations[key]; ok {
			if rel.Stance == faction.StanceWar {
				continue
			}
		}
		fids = append(fids, fid)
	}
	sort.Slice(fids, func(i, j int) bool { return fids[i] < fids[j] })
	return fids
}

// getFactionGoodAmount bir fraksiyonun belirli bir maldan kaç adet olduğunu döner.
func getFactionGoodAmount(f *faction.Faction, good economy.GoodType) int {
	if f == nil {
		return 0
	}
	switch good {
	case economy.GoodGrain:
		return f.Grain
	case economy.GoodIron:
		return f.Iron
	case economy.GoodTimber:
		return f.Timber
	case economy.GoodSpice:
		return f.Spice
	case economy.GoodCloth:
		return f.Cloth
	default:
		return 0
	}
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
