package render

import (
	"image"
	"image/color"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Layout sabitleri ────────────────────────────────────────────────

const (
	bottomBarH = float32(80)

	minimapW = float32(240)
	minimapH = float32(165)

	evLogW = float32(255)
	evLogH = float32(190)

	infoPanelW = float32(265)
	infoPanelH = float32(355)

	btnW = float32(90)
	btnH = float32(52)

	panelPad = float64(12)
)

func bottomBarTop() float32 { return float32(ScreenHeight) - bottomBarH }
func minimapX() float32     { return float32(ScreenWidth) - minimapW - 5 }
func minimapY() float32     { return bottomBarTop() - minimapH - 5 }
func evLogX() float32       { return float32(ScreenWidth) - evLogW - 5 }
func infoPanelX() float32   { return 5 }
func infoPanelY() float32   { return bottomBarTop() - infoPanelH - 5 }

var (
	panelBg     = color.RGBA{12, 10, 8, 230}
	panelBorder = color.RGBA{110, 90, 50, 255}
	panelBg2    = color.RGBA{18, 15, 10, 215}

	// whiteImage DrawTriangles için renk kaynağı olarak kullanılır.
	whiteImage = func() *ebiten.Image {
		img := ebiten.NewImage(1, 1)
		img.Fill(color.White)
		return img
	}()

	// miniMapBg minimap arka plan görseli (assets/maps/mini-map.png)
	miniMapBg     *ebiten.Image
	miniMapLoaded bool

	// buildingSheet bina sprite sheet'i (assets/sprites/buildings.jpg)
	// 3×2 grid: market/farm/barracks üst sıra, port/walls/temple alt sıra
	buildingSheet       *ebiten.Image
	buildingSheetLoaded bool
)

// buildingDisplayOrder bina slotlarının sırasını belirler.
var buildingDisplayOrder = []string{"market", "farm", "barracks", "port", "walls", "temple"}

func ensureBuildingSheet() {
	if buildingSheetLoaded {
		return
	}
	buildingSheetLoaded = true
	buildingSheet = tryLoadImage("assets/sprites/buildings.jpg")
}

// buildingSpriteRect sprite sheet'in gerçek boyutlarına göre bina hücresini döner.
// Görüntü 3 sütun × 2 satır eşit hücrelerden oluşur; alt %10 label alanı kırpılır.
func buildingSpriteRect(id string, sheet *ebiten.Image) image.Rectangle {
	idx := map[string]int{
		"market": 0, "farm": 1, "barracks": 2,
		"port": 3, "walls": 4, "temple": 5,
	}[id]
	w := sheet.Bounds().Dx()
	h := sheet.Bounds().Dy()
	cellW := w / 3
	cellH := h / 2
	col := idx % 3
	row := idx / 3
	x0 := col * cellW
	y0 := row * cellH
	// Alt %10 kırpılır — label metni hariç tutulur
	spriteH := cellH * 9 / 10
	return image.Rect(x0, y0, x0+cellW, y0+spriteH)
}

func ensureMiniMapBg() {
	if miniMapLoaded {
		return
	}
	miniMapLoaded = true
	miniMapBg = tryLoadImage("assets/maps/mini-map.png")
}

// BottomButtonRects Total War stilinde sağ alt aksiyon butonlarının piksel dikdörtgenlerini döner.
// [0]=Diplomasi  [1]=Teknoloji  [2]=Tur Bitir
func BottomButtonRects() [3][4]float32 {
	by := bottomBarTop() + (bottomBarH-btnH)/2
	endX := float32(ScreenWidth) - btnW - 5
	techX := endX - btnW - 5
	diplX := techX - btnW - 5
	return [3][4]float32{
		{diplX, by, btnW, btnH},
		{techX, by, btnW, btnH},
		{endX, by, btnW, btnH},
	}
}

// ── Ana alt bar ──────────────────────────────────────────────────────

// DrawBottomPanel Total War stilinde tam genişlik alt kaynak/aksiyon barını çizer.
func DrawBottomPanel(screen *ebiten.Image, gs *state.GameState, showDiplomacy, showTech bool) {
	by := bottomBarTop()
	bw := float32(ScreenWidth)

	vector.FillRect(screen, 0, by, bw, bottomBarH, panelBg, false)
	vector.FillRect(screen, 0, by, bw, 3, panelBorder, false)
	vector.StrokeLine(screen, 0, by, bw, by, 1.5, panelBorder, false)

	f, hasPlayer := gs.Factions[gs.PlayerFactionID]

	// Sol blok: fraksiyon amblemi + isim (satır 1) + tarih (satır 2) + sezon/tur (satır 3)
	if hasPlayer {
		fc := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		cx := float32(34)
		cy := by + bottomBarH/2
		vector.FillCircle(screen, cx, cy, 22, fc, true)
		vector.StrokeCircle(screen, cx, cy, 22, 2, panelBorder, true)
		initial := string([]rune(f.NameTR)[:1])
		DrawTextCentered(screen, initial, float64(cx), float64(cy)-8, FaceLarge, color.RGBA{255, 255, 255, 240})

		DrawText(screen, f.NameTR, 64, float64(by)+8, FaceLarge, fc)

		months := [...]string{"", "Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran",
			"Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık"}
		sea := gs.CurrentSeason()
		dateStr := months[gs.Month] + " " + itoa(gs.Year)
		DrawText(screen, dateStr, 64, float64(by)+30, FaceMed, ColorGold)
		DrawText(screen, sea.DisplayName()+"  •  Tur "+itoa(gs.Turn), 64, float64(by)+52, FaceSmall,
			color.RGBA{160, 200, 100, 220})
	}

	// Kaynaklar: 3 sütun — sol kenarı sol bloktan yeterince uzakta
	if hasPlayer {
		// Sütun 1: Altın / Tahıl
		rx1 := float64(310)
		// Sütun 2: Demir / Kereste
		rx2 := float64(450)
		// Sütun 3: Gelir / Faz
		rx3 := float64(590)
		ry := float64(by) + 12

		drawResRow(screen, rx1, ry, "Altin", itoa(f.Gold), ColorGold)
		drawResRow(screen, rx1, ry+26, "Tahil", itoa(f.Grain), ColorWhite)

		drawResRow(screen, rx2, ry, "Demir", itoa(f.Iron), color.RGBA{180, 180, 220, 255})
		drawResRow(screen, rx2, ry+26, "Kereste", itoa(f.Timber), color.RGBA{180, 140, 80, 255})

		income := calcPlayerIncome(gs)
		incCol := ColorGold
		if income < 0 {
			incCol = ColorRed
		}
		sign := "+"
		if income < 0 {
			sign = ""
		}
		drawResRow(screen, rx3, ry, "Gelir", sign+itoa(income)+"/tur", incCol)
		DrawText(screen, phaseLabel(gs.Phase), rx3, ry+26, FaceSmall, ColorGray)
	}

	// Zafer göstergesi — kaynak sütunundan sonra başlar
	if hasPlayer {
		drawVictoryProgress(screen, gs)
	}

	// Sağ: aksiyon butonları
	rects := BottomButtonRects()
	labels := [3]string{"Diplomasi", "Teknoloji", "Tur Bitir ►"}
	active := [3]bool{showDiplomacy, showTech, false}
	bgNorm := [3]color.RGBA{
		{40, 65, 110, 215},
		{60, 40, 95, 215},
		{40, 90, 40, 230},
	}
	bgAct := [3]color.RGBA{
		{80, 130, 200, 240},
		{110, 70, 170, 240},
		{70, 150, 70, 255},
	}
	for i, r := range rects {
		bg := bgNorm[i]
		if active[i] {
			bg = bgAct[i]
		}
		vector.FillRect(screen, r[0], r[1], r[2], r[3], bg, false)
		vector.StrokeRect(screen, r[0], r[1], r[2], r[3], 1.5, panelBorder, false)
		tw := MeasureText(labels[i], FaceMed)
		DrawText(screen, labels[i], float64(r[0])+float64(r[2])/2-tw/2, float64(r[1])+15, FaceMed, ColorWhite)
	}
}

// ── Olay Logu (sağ üst) ──────────────────────────────────────────────

// DrawEventLog sağ üst köşede son olayları listeler.
func DrawEventLog(screen *ebiten.Image, events []string) {
	if len(events) == 0 {
		return
	}
	ex := evLogX()
	ey := float32(10)

	vector.FillRect(screen, ex, ey, evLogW, evLogH, panelBg, false)
	drawPanelBorder(screen, ex, ey, evLogW, evLogH)
	vector.FillRect(screen, ex, ey, evLogW, 3, panelBorder, false)

	titleW := MeasureText("Olay Mesajları", FaceMed)
	DrawText(screen, "Olay Mesajları", float64(ex)+float64(evLogW)/2-titleW/2, float64(ey)+8, FaceMed,
		color.RGBA{220, 190, 100, 255})

	lx := float64(ex) + 10
	ly := float64(ey) + 30
	for _, ev := range events {
		if ly > float64(ey)+float64(evLogH)-18 {
			break
		}
		DrawText(screen, "• "+ev, lx, ly, FaceSmall, color.RGBA{210, 200, 180, 230})
		ly += 18
	}
}

// ── Minimap (sağ alt, alt barın üstünde) ─────────────────────────────

// DrawMinimap küçük ölçekli dünya haritasını, fraksiyon sahipliğini ve
// kamera viewport dikdörtgenini çizer.
func DrawMinimap(screen *ebiten.Image, gs *state.GameState, camX, camY, camScale float64) {
	ensureMiniMapBg()

	mx := minimapX()
	my := minimapY()

	const borderThick = float32(3)
	const cornerSize = float32(8)

	// Dış gölge efekti
	vector.FillRect(screen, mx-4, my-4, minimapW+8, minimapH+8, color.RGBA{0, 0, 0, 100}, false)

	// Dış çerçeve — altın rengi
	vector.FillRect(screen, mx-borderThick, my-borderThick, minimapW+borderThick*2, minimapH+borderThick*2,
		color.RGBA{140, 110, 50, 255}, false)

	// Minimap içi — görsel varsa onu kullan, yoksa koyu arka plan
	if miniMapBg != nil {
		bw, bh := miniMapBg.Bounds().Dx(), miniMapBg.Bounds().Dy()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(minimapW)/float64(bw), float64(minimapH)/float64(bh))
		op.GeoM.Translate(float64(mx), float64(my))
		// Hafif karartma — sahiplik renkleri daha net görünsün
		op.ColorScale.Scale(0.72, 0.72, 0.72, 1.0)
		screen.DrawImage(miniMapBg, op)
	} else {
		vector.FillRect(screen, mx, my, minimapW, minimapH, color.RGBA{15, 22, 35, 255}, false)
		drawMinimapPolygons(screen, gs, mx, my)
	}

	// Sahiplik renk katmanı — yarı saydam doldurulmuş daireler
	scaleX := float64(minimapW) / float64(WorldW)
	scaleY := float64(minimapH) / float64(WorldH)
	drawMinimapOwnershipOverlay(screen, gs, float32(scaleX), float32(scaleY), mx, my)

	// İç kenara ince koyu çizgi
	vector.StrokeRect(screen, mx, my, minimapW, minimapH, 1, color.RGBA{30, 25, 15, 200}, false)

	// Köşe süslemeleri
	drawMinimapCorner(screen, mx, my, cornerSize, cornerSize)
	drawMinimapCorner(screen, mx+minimapW, my, -cornerSize, cornerSize)
	drawMinimapCorner(screen, mx, my+minimapH, cornerSize, -cornerSize)
	drawMinimapCorner(screen, mx+minimapW, my+minimapH, -cornerSize, -cornerSize)

	// Başlık etiketi
	titleW := float32(MeasureText("MİNİ HARİTA", FaceSmall))
	DrawText(screen, "MİNİ HARİTA",
		float64(mx)+float64(minimapW)/2-float64(titleW)/2,
		float64(my)-14,
		FaceSmall, color.RGBA{200, 170, 80, 200})

	// Viewport dikdörtgeni
	vpW := float32((ScreenWidth / camScale) * scaleX)
	vpH := float32((ScreenHeight / camScale) * scaleY)
	vpX := mx + float32((camX-ScreenWidth/(2*camScale))*scaleX)
	vpY := my + float32((camY-ScreenHeight/(2*camScale))*scaleY)

	if vpX < mx {
		vpW -= mx - vpX
		vpX = mx
	}
	if vpY < my {
		vpH -= my - vpY
		vpY = my
	}
	if vpX+vpW > mx+minimapW {
		vpW = mx + minimapW - vpX
	}
	if vpY+vpH > my+minimapH {
		vpH = my + minimapH - vpY
	}
	if vpW > 1 && vpH > 1 {
		// Viewport kenarlığı — parlak sarı, iç kısmı tamamen şeffaf
		vector.StrokeRect(screen, vpX, vpY, vpW, vpH, 2, color.RGBA{255, 225, 55, 240}, false)
		// Köşe vurguları
		cLen := float32(5)
		vgold := color.RGBA{255, 245, 130, 255}
		vector.StrokeLine(screen, vpX, vpY, vpX+cLen, vpY, 2, vgold, false)
		vector.StrokeLine(screen, vpX, vpY, vpX, vpY+cLen, 2, vgold, false)
		vector.StrokeLine(screen, vpX+vpW, vpY, vpX+vpW-cLen, vpY, 2, vgold, false)
		vector.StrokeLine(screen, vpX+vpW, vpY, vpX+vpW, vpY+cLen, 2, vgold, false)
		vector.StrokeLine(screen, vpX, vpY+vpH, vpX+cLen, vpY+vpH, 2, vgold, false)
		vector.StrokeLine(screen, vpX, vpY+vpH, vpX, vpY+vpH-cLen, 2, vgold, false)
		vector.StrokeLine(screen, vpX+vpW, vpY+vpH, vpX+vpW-cLen, vpY+vpH, 2, vgold, false)
		vector.StrokeLine(screen, vpX+vpW, vpY+vpH, vpX+vpW, vpY+vpH-cLen, 2, vgold, false)
	}
}

// drawMinimapCorner köşe L şeklinde süsleme çizer. dx/dy negatifse ters yöne çizer.
func drawMinimapCorner(screen *ebiten.Image, x, y, dx, dy float32) {
	col := color.RGBA{200, 165, 60, 255}
	absX := dx
	if absX < 0 {
		absX = -absX
	}
	absY := dy
	if absY < 0 {
		absY = -absY
	}
	vector.StrokeLine(screen, x, y, x+absX*(dx/absX), y, 2, col, false)
	vector.StrokeLine(screen, x, y, x, y+absY*(dy/absY), 2, col, false)
}

// drawMinimapPolygons ülke sınırlarını poligon olarak çizer.
func drawMinimapPolygons(screen *ebiten.Image, gs *state.GameState, offsetX, offsetY float32) {
	if gs.ShapeData.Bounds.MaxX == 0 { // Veri yüklenmemişse atla
		return
	}
	bounds := gs.ShapeData.Bounds
	mapW := bounds.MaxX - bounds.MinX
	mapH := bounds.MaxY - bounds.MinY
	scaleX := minimapW / mapW
	scaleY := minimapH / mapH

	borderCol := color.RGBA{70, 60, 50, 255}

	for _, region := range gs.Regions {
		if region.IsSea {
			continue
		}
		for _, polygon := range region.Shape {
			if len(polygon) < 3 {
				continue
			}
			for i := 0; i < len(polygon); i++ {
				p1 := polygon[i]
				p2 := polygon[(i+1)%len(polygon)]
				x1 := offsetX + (p1[0]-bounds.MinX)*scaleX
				y1 := offsetY + (p1[1]-bounds.MinY)*scaleY
				x2 := offsetX + (p2[0]-bounds.MinX)*scaleX
				y2 := offsetY + (p2[1]-bounds.MinY)*scaleY
				vector.StrokeLine(screen, x1, y1, x2, y2, 1, borderCol, true)
			}
		}
	}
}

// colorToScale rengi DrawTriangles için float32 ölçeğine dönüştürür.
func colorToScale(clr color.Color) (float32, float32, float32, float32) {
	r, g, b, a := clr.RGBA()
	rf := float32(r) / 0xffff
	gf := float32(g) / 0xffff
	bf := float32(b) / 0xffff
	af := float32(a) / 0xffff
	return rf, gf, bf, af
}

// drawMinimapOwnership fraksiyon sahipliğini minimap üzerinde küçük daireler olarak gösterir.
func drawMinimapOwnership(screen *ebiten.Image, gs *state.GameState, scaleX, scaleY, offsetX, offsetY float32) {
	for _, region := range gs.Regions {
		if region.IsSea || region.OwnerID == "" {
			continue
		}
		px := offsetX + float32(wcX(region.WorldX))*scaleX
		py := offsetY + float32(wcY(region.WorldY))*scaleY
		col := factionColor(gs, region.OwnerID)
		col.A = 200
		vector.FillCircle(screen, px, py, 3, col, true)
	}
}

// drawMinimapOwnershipOverlay fraksiyon sahipliğini mini-map.png üstüne yarı saydam
// renkli daireler olarak katmanlar; oyuncu bölgeleri biraz daha büyük gösterilir.
func drawMinimapOwnershipOverlay(screen *ebiten.Image, gs *state.GameState, scaleX, scaleY, offsetX, offsetY float32) {
	for _, region := range gs.Regions {
		if region.IsSea || region.OwnerID == "" {
			continue
		}
		px := offsetX + float32(wcX(region.WorldX))*scaleX
		py := offsetY + float32(wcY(region.WorldY))*scaleY

		col := factionColor(gs, region.OwnerID)

		isPlayer := region.OwnerID == string(gs.PlayerFactionID)
		radius := float32(4)
		if isPlayer {
			radius = 5.5
		}

		// Hafif gölge
		shadow := color.RGBA{0, 0, 0, 80}
		vector.FillCircle(screen, px+1, py+1, radius+1, shadow, true)

		// Dolu daire — yarı saydam fraksiyon rengi
		col.A = 180
		vector.FillCircle(screen, px, py, radius, col, true)

		// Oyuncu bölgesi ise parlak kenarlık
		if isPlayer {
			vector.StrokeCircle(screen, px, py, radius, 1.5, color.RGBA{255, 240, 120, 230}, true)
		}
	}
}

// ── Bölge Bilgi Paneli (sol alt) ──────────────────────────────────────

// DrawRegionPanel seçili bölge bilgisini sol altta gösterir.
func DrawRegionPanel(screen *ebiten.Image, gs *state.GameState, rid world.RegionID) {
	if rid == "" {
		return
	}
	region, ok := gs.Regions[rid]
	if !ok || region.IsSea {
		return
	}

	px := infoPanelX()
	py := infoPanelY()
	pw := infoPanelW
	ph := infoPanelH

	vector.FillRect(screen, px, py, pw, ph, panelBg, false)
	drawPanelBorder(screen, px, py, pw, ph)
	vector.FillRect(screen, px, py, pw, 3, panelBorder, false)

	lx := float64(px) + panelPad
	ly := float64(py) + 10

	DrawText(screen, region.NameTR, lx, ly, FaceLarge, ColorYellow)
	ly += 24

	ownerName, ownerCol := ownerDisplay(gs, region.OwnerID)
	DrawText(screen, "Sahip: "+ownerName, lx, ly, FaceSmall, ownerCol)
	ly += 18

	DrawText(screen, terrainLabel(region.Terrain)+"  │  "+religionLabel(region.Religion), lx, ly, FaceSmall, ColorGray)
	ly += 16

	sepW := pw - float32(panelPad*2)
	vector.StrokeLine(screen, float32(lx), float32(ly), float32(lx)+sepW, float32(ly), 1, panelBorder, false)
	ly += 8

	// Kaynaklar — iki sütun
	DrawText(screen, "✦ "+itoa(region.GoldIncome())+" Altın", lx, ly, FaceSmall, ColorGold)
	DrawText(screen, "◈ "+itoa(region.BaseGrainOutput)+" Tahıl", lx+120, ly, FaceSmall, ColorWhite)
	ly += 18

	DrawText(screen, "Memnuniyet: "+itoa(region.Satisfaction)+"%", lx, ly, FaceSmall, ColorGray)
	drawBar(screen, float32(lx+100), float32(ly)+1, sepW-100, 9, float64(region.Satisfaction)/100,
		satisfactionColor(region.Satisfaction))
	ly += 18

	DrawText(screen, "Vergi: %"+itoa(region.TaxRate), lx, ly, FaceSmall, ColorGray)
	drawBar(screen, float32(lx+100), float32(ly)+1, sepW-100, 9, float64(region.TaxRate)/100,
		color.RGBA{200, 140, 40, 255})
	ly += 18

	// Din dönüşüm ilerlemesi
	if region.ConversionTurns > 0 {
		ownerRel := ""
		if f, ok2 := gs.Factions[gs.PlayerFactionID]; ok2 && region.OwnerID == string(gs.PlayerFactionID) {
			ownerRel = string(f.Religion)
		} else {
			for fid, f := range gs.Factions {
				if string(fid) == region.OwnerID {
					ownerRel = string(f.Religion)
					break
				}
			}
		}
		if ownerRel != "" && ownerRel != region.Religion {
			convPct := float64(region.ConversionTurns) / 24.0
			DrawText(screen, "☩ Dönüşüm: "+religionLabel(ownerRel), lx, ly, FaceSmall, color.RGBA{180, 140, 240, 200})
			ly += 14
			drawBar(screen, float32(lx), float32(ly), sepW, 7, convPct, color.RGBA{150, 100, 220, 220})
			ly += 12
		}
	}

	if region.IsRebellionRisk() {
		DrawText(screen, "⚠  İSYAN RİSKİ!", lx, ly, FaceMed, ColorRed)
		ly += 18
	}

	// ── Binalar bölümü ────────────────────────────────────────────────
	ly += 4
	vector.StrokeLine(screen, float32(lx), float32(ly), float32(lx)+sepW, float32(ly), 1, panelBorder, false)
	ly += 6

	bldTitleW := MeasureText("BİNALAR", FaceSmall)
	DrawText(screen, "BİNALAR", float64(px)+float64(pw)/2-bldTitleW/2, ly, FaceSmall, color.RGBA{200, 170, 90, 220})
	ly += 17

	drawBuildingGrid(screen, gs, region, px, float32(ly), pw)

	// Oyuncu ipucu — panelin en altına sabitlendi
	if region.OwnerID == string(gs.PlayerFactionID) {
		DrawText(screen, "[R] Milis  [1-6] Bina  [,.] Vergi", lx, float64(py)+float64(ph)-14, FaceSmall,
			color.RGBA{100, 200, 100, 170})
	}
}

// DrawArmyPanel seçili ordu bilgisini sol altta gösterir.
func DrawArmyPanel(screen *ebiten.Image, gs *state.GameState, aid army.ArmyID) {
	if aid == "" {
		return
	}
	a, ok := gs.Armies[aid]
	if !ok {
		return
	}

	px := infoPanelX()
	py := infoPanelY() + infoPanelH - 130
	pw := infoPanelW
	ph := float32(130)

	vector.FillRect(screen, px, py, pw, ph, panelBg, false)
	drawPanelBorder(screen, px, py, pw, ph)
	vector.FillRect(screen, px, py, pw, 3, panelBorder, false)

	lx := float64(px) + panelPad
	ly := float64(py) + 10

	DrawText(screen, "Seçili Ordu", lx, ly, FaceLarge, ColorYellow)
	ly += 22

	if region, ok2 := gs.Regions[a.RegionID]; ok2 {
		DrawText(screen, "Konum: "+region.NameTR, lx, ly, FaceSmall, ColorGray)
	}
	ly += 18

	DrawText(screen, "Birim: "+itoa(len(a.Units))+"/"+itoa(army.MaxArmySize), lx, ly, FaceSmall, ColorWhite)
	ly += 18

	mpCol := ColorGold
	if a.MovePoints == 0 {
		mpCol = ColorRed
	}
	DrawText(screen, "Hareket: "+itoa(a.MovePoints)+"/"+itoa(a.MaxMovePoints), lx, ly, FaceSmall, mpCol)
	ly += 18

	hint := "Sağ tık → hareket / saldırı"
	hintCol := color.RGBA{120, 200, 120, 200}
	if a.MovePoints == 0 {
		hint = "Bu tur hareket puanı tükendi"
		hintCol = color.RGBA{180, 100, 100, 200}
	}
	DrawText(screen, hint, lx, ly, FaceSmall, hintCol)
}

// drawGameOver oyun sonu ekranını çizer.
func drawGameOver(screen *ebiten.Image, gs *state.GameState) {
	overlay := ebiten.NewImage(int(ScreenWidth), int(ScreenHeight))
	overlay.Fill(color.RGBA{5, 3, 2, 230})
	screen.DrawImage(overlay, nil)

	cy := ScreenHeight/2 - 80

	switch gs.WinnerID {
	case gs.PlayerFactionID:
		DrawTextCentered(screen, "ZAFERİN!", ScreenWidth/2, cy, FaceLarge, ColorGold)
		cy += 34
		vtitle := victoryTypeLabel(gs.Victory.Type)
		DrawTextCentered(screen, vtitle, ScreenWidth/2, cy, FaceMed, color.RGBA{255, 200, 80, 230})
		cy += 26
		if f, ok := gs.Factions[gs.PlayerFactionID]; ok {
			DrawTextCentered(screen, f.NameTR+" tarihe geçti.", ScreenWidth/2, cy, FaceMed, ColorWhite)
		}
	case "":
		DrawTextCentered(screen, "YENİLDİN", ScreenWidth/2, cy, FaceLarge, ColorRed)
		cy += 34
		DrawTextCentered(screen, "Tüm bölgelerini kaybettin.", ScreenWidth/2, cy, FaceMed, ColorGray)
	default:
		DrawTextCentered(screen, "YENİLDİN", ScreenWidth/2, cy, FaceLarge, ColorRed)
		cy += 34
		if f, ok := gs.Factions[gs.WinnerID]; ok {
			DrawTextCentered(screen, f.NameTR+" galip geldi.", ScreenWidth/2, cy, FaceMed, ColorGray)
		}
	}

	cy += 40
	// İstatistik satırı
	regionCount := len(gs.RegionsOwnedBy(gs.PlayerFactionID))
	armyCount := 0
	for _, a := range gs.Armies {
		if a.OwnerID == string(gs.PlayerFactionID) {
			armyCount++
		}
	}
	stats := "Tur: " + itoa(gs.Turn) + "  │  Yıl: " + itoa(gs.Year) +
		"  │  Bölge: " + itoa(regionCount) + "  │  Ordu: " + itoa(armyCount)
	DrawTextCentered(screen, stats, ScreenWidth/2, cy, FaceSmall, ColorGray)
	cy += 30
	DrawTextCentered(screen, "[Esc] Ana Menü", ScreenWidth/2, cy, FaceSmall, color.RGBA{160, 160, 160, 200})
}

// drawHistoricalEventPopup büyük tarihsel olayları dramatik bir tam ekran katmanıyla gösterir.
func drawHistoricalEventPopup(screen *ebiten.Image, title, desc string) {
	overlay := ebiten.NewImage(int(ScreenWidth), int(ScreenHeight))
	overlay.Fill(color.RGBA{2, 1, 5, 210})
	screen.DrawImage(overlay, nil)

	// Dekoratif çerçeve
	bx, by := float32(ScreenWidth/2-340), float32(ScreenHeight/2-120)
	bw, bh := float32(680), float32(240)
	vector.FillRect(screen, bx, by, bw, bh, color.RGBA{15, 10, 25, 245}, false)
	vector.StrokeRect(screen, bx, by, bw, bh, 2.5, color.RGBA{180, 140, 50, 255}, false)
	vector.StrokeRect(screen, bx+4, by+4, bw-8, bh-8, 1, color.RGBA{120, 90, 30, 200}, false)

	// Üst şerit
	vector.FillRect(screen, bx, by, bw, 4, color.RGBA{220, 170, 50, 255}, false)

	cy := float64(by) + 28
	DrawTextCentered(screen, "— TARİHSEL OLAY —", ScreenWidth/2, cy, FaceSmall, color.RGBA{180, 140, 50, 200})
	cy += 26
	DrawTextCentered(screen, title, ScreenWidth/2, cy, FaceLarge, color.RGBA{255, 220, 80, 255})
	cy += 30

	// Açıklama — uzun metni satırlara böl
	wrapText(screen, desc, float64(bx)+30, cy, float64(bw-60), FaceMed, color.RGBA{210, 200, 180, 230})

	cy = float64(by) + float64(bh) - 28
	DrawTextCentered(screen, "[Enter / Boşluk / Tıkla] Devam Et", ScreenWidth/2, cy, FaceSmall, color.RGBA{140, 130, 100, 200})
}

// wrapText metni belirtilen genişlikte kelime bazlı satırlara bölerek çizer.
func wrapText(screen *ebiten.Image, text string, x, y, maxW float64, face *text.GoTextFace, col color.Color) {
	words := splitWords(text)
	line := ""
	ly := y
	for _, w := range words {
		test := line
		if test != "" {
			test += " "
		}
		test += w
		if MeasureText(test, face) > maxW && line != "" {
			DrawText(screen, line, x, ly, face, col)
			ly += 22
			line = w
		} else {
			line = test
		}
	}
	if line != "" {
		DrawText(screen, line, x, ly, face, col)
	}
}

// splitWords metni boşluklara göre böler.
func splitWords(s string) []string {
	var words []string
	cur := ""
	for _, r := range s {
		if r == ' ' {
			if cur != "" {
				words = append(words, cur)
				cur = ""
			}
		} else {
			cur += string(r)
		}
	}
	if cur != "" {
		words = append(words, cur)
	}
	return words
}

func victoryTypeLabel(vtype state.VictoryType) string {
	switch vtype {
	case state.VictoryDomination:
		return "Toprak Hakimiyeti Zaferi"
	case state.VictoryEconomic:
		return "Ekonomik Üstünlük Zaferi"
	case state.VictoryMilitary:
		return "Askeri Üstünlük Zaferi"
	case state.VictoryReligious:
		return "Dinî Zafer"
	}
	return "Zafer"
}

// ── Zafer İlerleme Göstergesi ─────────────────────────────────────────

// drawVictoryProgress alt barda seçilen zafer tipine göre ilerlemeyi gösterir.
func drawVictoryProgress(screen *ebiten.Image, gs *state.GameState) {
	if gs.PlayerFactionID == "" {
		return
	}

	vx := float64(730)
	vy := float64(bottomBarTop()) + 8
	barW := float32(150)

	titleCol := color.RGBA{220, 190, 100, 220}
	DrawText(screen, "Zafer Hedefi", vx, vy, FaceSmall, titleCol)
	vy += 18

	switch gs.Victory.Type {
	case state.VictoryDomination, "":
		target := gs.Victory.TargetRegionCount
		if target == 0 {
			target = 20
		}
		current := len(gs.RegionsOwnedBy(gs.PlayerFactionID))
		DrawText(screen, "⚑ Bölge: "+itoa(current)+"/"+itoa(target), vx, vy, FaceMed, ColorWhite)
		vy += 18
		drawBar(screen, float32(vx), float32(vy), barW, 8, clampF(float64(current)/float64(target)), ColorGold)

	case state.VictoryEconomic:
		threshold := gs.Victory.TargetGoldIncome
		if threshold == 0 {
			threshold = 5000
		}
		holdTurns := gs.Victory.GoldHoldTurns
		if holdTurns == 0 {
			holdTurns = 5
		}
		gold := 0
		if f, ok := gs.Factions[gs.PlayerFactionID]; ok {
			gold = f.Gold
		}
		DrawText(screen, "✦ Altın: "+itoa(gold)+"/"+itoa(threshold), vx, vy, FaceMed, ColorGold)
		vy += 18
		drawBar(screen, float32(vx), float32(vy), barW, 8, clampF(float64(gold)/float64(threshold)), ColorGold)
		vy += 12
		turnsStr := itoa(gs.EconomicVictoryTurns) + "/" + itoa(holdTurns) + " tur korundu"
		DrawText(screen, turnsStr, vx, vy, FaceSmall, ColorGray)

	case state.VictoryMilitary:
		targetStr := gs.Victory.TargetArmyStrength
		if targetStr == 0 {
			targetStr = 200
		}
		targetDef := gs.Victory.TargetDefeated
		if targetDef == 0 {
			targetDef = 3
		}
		totalStr := 0
		for _, a := range gs.Armies {
			if a.OwnerID == string(gs.PlayerFactionID) {
				totalStr += a.TotalStrength(gs.UnitTypes)
			}
		}
		eliminated := 0
		for fid, f := range gs.Factions {
			if fid != gs.PlayerFactionID && f.IsEliminated {
				eliminated++
			}
		}
		DrawText(screen, "⚔ Güç: "+itoa(totalStr)+"/"+itoa(targetStr), vx, vy, FaceMed, ColorWhite)
		vy += 18
		drawBar(screen, float32(vx), float32(vy), barW, 8, clampF(float64(totalStr)/float64(targetStr)), color.RGBA{200, 80, 80, 255})
		vy += 12
		DrawText(screen, "Yenilgi: "+itoa(eliminated)+"/"+itoa(targetDef), vx, vy, FaceSmall, ColorGray)

	case state.VictoryReligious:
		held := 0
		total := len(gs.Victory.RequiredRegions)
		for _, rid := range gs.Victory.RequiredRegions {
			if r, ok := gs.Regions[rid]; ok && r.OwnerID == string(gs.PlayerFactionID) {
				held++
			}
		}
		DrawText(screen, "✝ Kutsal Şehir: "+itoa(held)+"/"+itoa(total), vx, vy, FaceMed, color.RGBA{200, 160, 255, 255})
		vy += 18
		drawBar(screen, float32(vx), float32(vy), barW, 8, clampF(float64(held)/float64(total+1)), color.RGBA{160, 120, 255, 255})
		vy += 12
		DrawText(screen, itoa(gs.ReligiousVictoryTurns)+"/12 tur", vx, vy, FaceSmall, ColorGray)
	}
}

// clampF 0.0–1.0 aralığına sıkıştırır.
func clampF(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func drawResRow(screen *ebiten.Image, x, y float64, label, value string, col color.RGBA) {
	DrawText(screen, label, x, y, FaceSmall, ColorGray)
	tw := MeasureText(value, FaceMed)
	DrawText(screen, value, x+90-tw, y, FaceMed, col)
}

func calcPlayerIncome(gs *state.GameState) int {
	total := 0
	for _, r := range gs.Regions {
		if r.OwnerID == string(gs.PlayerFactionID) {
			total += r.GoldIncome()
		}
	}
	return total
}

// drawBuildingGrid bölgedeki binaları sprite thumbnail'leri olarak 3×2 ızgarada çizer.
// İnşa edilmiş binalar renkli sprite ile, boş slotlar soluk çerçeve ile gösterilir.
func drawBuildingGrid(screen *ebiten.Image, gs *state.GameState, region *world.Region, panelX, startY, panelW float32) {
	ensureBuildingSheet()

	builtSet := make(map[string]bool, len(region.Buildings))
	for _, bid := range region.Buildings {
		builtSet[bid] = true
	}

	const cols = 3
	pad := float32(panelPad)
	availW := panelW - pad*2
	slotW := availW / float32(cols)
	spriteH := float32(54)
	nameH := float32(16)
	rowH := spriteH + nameH + 5

	for i, bid := range buildingDisplayOrder {
		col := i % cols
		row := i / cols

		sx := panelX + pad + float32(col)*slotW
		sy := startY + float32(row)*rowH
		innerW := slotW - 3

		isBuilt := builtSet[bid]

		// Arka plan ve çerçeve
		slotBg := color.RGBA{20, 16, 12, 200}
		borderCol := color.RGBA{55, 45, 30, 200}
		if isBuilt {
			slotBg = color.RGBA{42, 34, 18, 245}
			borderCol = panelBorder
		}
		vector.FillRect(screen, sx, sy, innerW, spriteH, slotBg, false)
		vector.StrokeRect(screen, sx, sy, innerW, spriteH, 1, borderCol, false)

		if isBuilt && buildingSheet != nil {
			r := buildingSpriteRect(bid, buildingSheet)
			sub := buildingSheet.SubImage(r).(*ebiten.Image)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(
				float64(innerW-2)/float64(r.Dx()),
				float64(spriteH-2)/float64(r.Dy()),
			)
			op.GeoM.Translate(float64(sx+1), float64(sy+1))
			screen.DrawImage(sub, op)

			// İnce altın vurgu çerçevesi
			vector.StrokeRect(screen, sx+1, sy+1, innerW-2, spriteH-2, 1, color.RGBA{160, 130, 50, 120}, false)
		} else {
			// Boş slot — kilitli görünüm
			DrawTextCentered(screen, "—", float64(sx)+float64(innerW)/2, float64(sy)+float64(spriteH)/2-8, FaceLarge, color.RGBA{55, 45, 35, 180})
		}

		// Bina adı
		bname := bid
		if b, ok := gs.BuildingTypes[bid]; ok {
			bname = b.NameTR
		}
		nameCol := color.RGBA{75, 65, 50, 200}
		if isBuilt {
			nameCol = ColorGold
		}
		DrawTextCentered(screen, bname, float64(sx)+float64(innerW)/2, float64(sy+spriteH)+3, FaceSmall, nameCol)
	}
}

func drawPanelBorder(screen *ebiten.Image, x, y, w, h float32) {
	vector.StrokeLine(screen, x, y, x+w, y, 1.5, panelBorder, false)
	vector.StrokeLine(screen, x, y+h, x+w, y+h, 1.5, panelBorder, false)
	vector.StrokeLine(screen, x, y, x, y+h, 1.5, panelBorder, false)
	vector.StrokeLine(screen, x+w, y, x+w, y+h, 1.5, panelBorder, false)
}

func drawBar(screen *ebiten.Image, x, y, w, h float32, fill float64, col color.Color) {
	vector.FillRect(screen, x, y, w, h, color.RGBA{40, 40, 40, 180}, false)
	if fill > 0 {
		vector.FillRect(screen, x, y, float32(float64(w)*fill), h, col, false)
	}
}

func satisfactionColor(v int) color.Color {
	if v >= 70 {
		return color.RGBA{60, 200, 60, 255}
	} else if v >= 40 {
		return color.RGBA{220, 180, 40, 255}
	}
	return color.RGBA{220, 60, 60, 255}
}

func ownerDisplay(gs *state.GameState, ownerID string) (string, color.Color) {
	if ownerID == "" {
		return "Sahipsiz", ColorGray
	}
	for fid, f := range gs.Factions {
		if string(fid) == ownerID {
			col := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
			if string(fid) == string(gs.PlayerFactionID) {
				return f.NameTR + " (Siz)", col
			}
			return f.NameTR, col
		}
	}
	return ownerID, ColorGray
}

func terrainLabel(t world.TerrainType) string {
	switch t {
	case world.TerrainPlain:
		return "Ova"
	case world.TerrainForest:
		return "Orman"
	case world.TerrainMountain:
		return "Dağ"
	case world.TerrainPass:
		return "Geçit"
	case world.TerrainCoast:
		return "Kıyı"
	case world.TerrainSea:
		return "Deniz"
	}
	return string(t)
}

func religionLabel(r string) string {
	switch r {
	case "catholic":
		return "Katolik"
	case "orthodox":
		return "Ortodoks"
	case "sunni":
		return "Sünni İslam"
	case "shia":
		return "Şii İslam"
	}
	return r
}

func phaseLabel(p state.Phase) string {
	switch p {
	case state.PhasePlayerTurn:
		return "Sizin Turunuz"
	case state.PhaseAITurn:
		return "AI Turu"
	case state.PhaseTurnResolution:
		return "Tur Sonu"
	default:
		return string(p)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte(n%10) + '0'
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
