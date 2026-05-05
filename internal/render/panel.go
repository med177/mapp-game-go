package render

import (
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

	minimapW = float32(205)
	minimapH = float32(155)

	evLogW = float32(255)
	evLogH = float32(190)

	infoPanelW = float32(265)
	infoPanelH = float32(200)

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
)

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

	// Sol: fraksiyon amblemi + isim + tarih
	if hasPlayer {
		fc := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		cx := float32(34)
		cy := by + bottomBarH/2
		vector.FillCircle(screen, cx, cy, 22, fc, true)
		vector.StrokeCircle(screen, cx, cy, 22, 2, panelBorder, true)
		// İlk harf
		initial := string([]rune(f.NameTR)[:1])
		DrawTextCentered(screen, initial, float64(cx), float64(cy)-8, FaceLarge, color.RGBA{255, 255, 255, 240})

		DrawText(screen, f.NameTR, 64, float64(by)+10, FaceLarge, fc)

		months := [...]string{"", "Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran",
			"Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık"}
		sea := gs.CurrentSeason()
		dateStr := months[gs.Month] + " " + itoa(gs.Year)
		DrawText(screen, dateStr, 64, float64(by)+34, FaceMed, ColorGold)
		DrawText(screen, sea.DisplayName()+"  Tur "+itoa(gs.Turn), 64+130, float64(by)+34, FaceSmall, color.RGBA{160, 200, 100, 220})
	}

	// Orta: kaynaklar
	if hasPlayer {
		rx := float64(260)
		ry := float64(by) + 12
		drawResRow(screen, rx, ry, "✦ Altın", itoa(f.Gold), ColorGold)
		drawResRow(screen, rx, ry+26, "◈ Tahıl", itoa(f.Grain), ColorWhite)

		drawResRow(screen, rx+160, ry, "⚙ Demir", itoa(f.Iron), color.RGBA{180, 180, 220, 255})
		drawResRow(screen, rx+160, ry+26, "🪵 Kereste", itoa(f.Timber), color.RGBA{180, 140, 80, 255})

		// Gelir tahmini
		income := calcPlayerIncome(gs)
		incCol := ColorGold
		if income < 0 {
			incCol = ColorRed
		}
		sign := "+"
		if income < 0 {
			sign = ""
		}
		drawResRow(screen, rx+320, ry, "↑ Gelir", sign+itoa(income)+"/tur", incCol)
		DrawText(screen, phaseLabel(gs.Phase), rx+320, ry+26, FaceSmall, ColorGray)
	}

	// Orta: zafer göstergesi
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

// DrawMinimap küçük ölçekli dünya haritasını ve fraksiyon sahipliğini çizer.
func DrawMinimap(screen *ebiten.Image, gs *state.GameState) {
	mx := minimapX()
	my := minimapY()

	// Çerçeve + arka plan
	vector.FillRect(screen, mx-2, my-2, minimapW+4, minimapH+4, panelBorder, false)
	vector.FillRect(screen, mx, my, minimapW, minimapH, color.RGBA{8, 12, 18, 255}, false)

	// Harita poligonlarını çiz
	drawMinimapPolygons(screen, gs, mx, my)

	// Sahiplik noktalarını çiz
	scaleX := float64(minimapW) / float64(WorldW)
	scaleY := float64(minimapH) / float64(WorldH)
	drawMinimapOwnership(screen, gs, float32(scaleX), float32(scaleY), mx, my)
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

	landCol := color.RGBA{45, 40, 30, 255}
	borderCol := color.RGBA{70, 60, 50, 255}
	seaCol := color.RGBA{20, 35, 50, 255}

	var path vector.Path
	for _, region := range gs.Regions {
		isOwned := region.OwnerID != ""
		col := landCol
		if isOwned {
			c := factionColor(gs, region.OwnerID)
			c.A = 200
			col = c
		}
		if region.IsSea {
			col = seaCol
		}

		for _, polygon := range region.Shape {
			if len(polygon) < 3 {
				continue
			}
			path.Reset()
			for i, p := range polygon {
				px := offsetX + (p[0]-bounds.MinX)*scaleX
				py := offsetY + (p[1]-bounds.MinY)*scaleY
				if i == 0 {
					path.MoveTo(px, py)
				} else {
					path.LineTo(px, py)
				}
			}
			path.Close()

			vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
			for i := range vs {
				vs[i].ColorR, vs[i].ColorG, vs[i].ColorB, vs[i].ColorA = colorToScale(col)
			}
			screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{
				FillRule: ebiten.EvenOdd,
			})

			// Sınır çizgileri
			if !region.IsSea {
				// Ebiten v2.2+ path.Stroke metodu var
				// path.Stroke(screen, 1.0, borderCol, true)
				// Şimdilik manuel çizim:
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
		px := offsetX + float32(region.WorldX*MapScale)*scaleX
		py := offsetY + float32(region.WorldY*MapScale)*scaleY
		col := factionColor(gs, region.OwnerID)
		col.A = 200
		vector.FillCircle(screen, px, py, 3, col, true)
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

	if len(region.Buildings) > 0 {
		var built string
		for i, bid := range region.Buildings {
			if b, ok2 := gs.BuildingTypes[bid]; ok2 {
				if i > 0 {
					built += ", "
				}
				built += b.NameTR
			}
		}
		DrawText(screen, built, lx, ly, FaceSmall, ColorGold)
		ly += 16
	}

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

	if region.OwnerID == string(gs.PlayerFactionID) {
		DrawText(screen, "[R] Milis  [1-6] Bina  [,.] Vergi", lx, ly, FaceSmall,
			color.RGBA{100, 200, 100, 200})
		ly += 14
	}

	if region.IsRebellionRisk() {
		DrawText(screen, "⚠  İSYAN RİSKİ!", lx, ly, FaceMed, ColorRed)
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

	vx := float64(580)
	vy := float64(bottomBarTop()) + 8
	barW := float32(160)

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
