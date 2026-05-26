package render

import (
	"image"
	"image/color"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/audio"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/victory"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Layout sabitleri ────────────────────────────────────────────────

const (
	bottomBarH   = float32(80)
	topStatusW   = float32(1000)
	topStatusH   = float32(80)
	topDateHudW  = float32(255)
	topDateHudH  = float32(80)
	actionHudPad = float32(8)
	actionHudGap = float32(5)

	minimapW = float32(240)
	minimapH = float32(165)

	evLogW             = float32(255)
	evLogH             = float32(520)
	evLogMinH          = float32(36)
	eventCardH         = float32(52)
	eventCardGap       = float32(7)
	maxEventLogEntries = 16

	infoPanelW = float32(305)
	infoPanelH = float32(600)

	btnW = float32(90)
	btnH = float32(52)

	panelPad = float64(12)
)

func bottomBarTop() float32 { return float32(ScreenHeight) - bottomBarH }
func minimapX() float32     { return float32(ScreenWidth) - minimapW - 5 }
func minimapY() float32     { return float32(ScreenHeight) - minimapH }
func evLogX() float32       { return float32(ScreenWidth) - evLogW }
func evLogY() float32       { return topDateHudH + 8 }
func infoPanelX() float32   { return 0 }
func infoPanelY() float32   { return float32(ScreenHeight) - infoPanelH }

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

	// buildingSheet bina sprite sheet'i (assets/sprites/buildings.png)
	// 3×2 grid: [barracks, market, temple] / [walls, farm, port]
	buildingSheet       *ebiten.Image
	buildingSheetLoaded bool
)

// buildingDisplayOrder bina slotlarının sırasını belirler.
var buildingDisplayOrder = []string{"market", "farm", "barracks", "walls", "temple", "port"}

func ensureBuildingSheet() {
	if buildingSheetLoaded {
		return
	}
	buildingSheetLoaded = true
	buildingSheet = tryLoadImage(ActiveScenarioPath + "/sprites/buildings.png")
}

// buildingSpriteRect sprite sheet'in gerçek boyutlarına göre bina hücresini döner.
// Görüntü 3 sütun × 2 satır eşit hücrelerden oluşur.
func buildingSpriteRect(id string, sheet *ebiten.Image) image.Rectangle {
	idx := map[string]int{
		"barracks": 0, "market": 1, "temple": 2,
		"walls": 3, "farm": 4, "port": 5,
	}[id]
	w := sheet.Bounds().Dx()
	h := sheet.Bounds().Dy()
	cellW := w / 3
	cellH := h / 2
	col := idx % 3
	row := idx / 3
	x0 := col * cellW
	y0 := row * cellH
	return image.Rect(x0, y0, x0+cellW, y0+cellH)
}

func ensureMiniMapBg() {
	if miniMapLoaded {
		return
	}
	miniMapLoaded = true
	miniMapBg = tryLoadImage(ActiveScenarioPath + "/maps/mini-map.png")
}

func bottomActionHudRect() (x, y, w, h float32) {
	w = btnW*4 + actionHudGap*3 + actionHudPad*2
	h = btnH + actionHudPad*2
	x = float32(ScreenWidth)/2 - w/2
	y = float32(ScreenHeight) - h
	if x < 0 {
		x = 0
	}
	return x, y, w, h
}

func mapModeHudRect() (x, y, w, h float32) {
	ax, ay, aw, _ := bottomActionHudRect()
	w = 230
	h = 30
	x = ax + aw/2 - w/2
	y = ay - h - 6
	return x, y, w, h
}

// mapModeButtonRects [0]=Normal [1]=Ticaret
func mapModeButtonRects() [2][4]float32 {
	x, y, w, h := mapModeHudRect()
	half := (w - 6) / 2
	return [2][4]float32{
		{x + 2, y + 2, half, h - 4},
		{x + 4 + half, y + 2, half, h - 4},
	}
}

// BottomButtonRects alt-orta aksiyon HUD'undaki buton dikdörtgenlerini döner.
// [0]=Ordu [1]=Diplomasi [2]=Teknoloji [3]=Tur Bitir
func BottomButtonRects() [4][4]float32 {
	hudX, hudY, _, _ := bottomActionHudRect()
	by := hudY + actionHudPad
	armyX := hudX + actionHudPad
	diplX := armyX + btnW + actionHudGap
	techX := diplX + btnW + actionHudGap
	endX := techX + btnW + actionHudGap
	return [4][4]float32{
		{armyX, by, btnW, btnH},
		{diplX, by, btnW, btnH},
		{techX, by, btnW, btnH},
		{endX, by, btnW, btnH},
	}
}

func bottomActionHudHit(fx, fy float64) bool {
	x, y, w, h := bottomActionHudRect()
	if fx >= float64(x) && fx <= float64(x+w) && fy >= float64(y) && fy <= float64(y+h) {
		return true
	}
	mx, my, mw, mh := mapModeHudRect()
	return fx >= float64(mx) && fx <= float64(mx+mw) && fy >= float64(my) && fy <= float64(my+mh)
}

func bottomActionButtonHit(fx, fy float64) bool {
	for _, r := range BottomButtonRects() {
		if rectF32Hit(fx, fy, r) {
			return true
		}
	}
	for _, r := range mapModeButtonRects() {
		if rectF32Hit(fx, fy, r) {
			return true
		}
	}
	return false
}

func topStatusPanelHit(fx, fy float64) bool {
	w := float64(topStatusW)
	if w > ScreenWidth {
		w = ScreenWidth
	}
	return fx >= 0 && fx <= w && fy >= 0 && fy <= float64(topStatusH)
}

func topDateHudRect() (x, y, w, h float32) {
	w = topDateHudW
	h = topDateHudH
	x = float32(ScreenWidth) - w
	if x < 0 {
		x = 0
	}
	return x, 0, w, h
}

func topDateHudHit(fx, fy float64) bool {
	x, y, w, h := topDateHudRect()
	return fx >= float64(x) && fx <= float64(x+w) && fy >= float64(y) && fy <= float64(y+h)
}

func topDateHudMenuButtonRect() (x, y, w, h float32) {
	hudX, hudY, hudW, _ := topDateHudRect()
	w = 72
	h = 34
	x = hudX + hudW - w - 10
	y = hudY + 23
	return x, y, w, h
}

func topDateHudMenuButtonHit(fx, fy float64) bool {
	x, y, w, h := topDateHudMenuButtonRect()
	return fx >= float64(x) && fx <= float64(x+w) && fy >= float64(y) && fy <= float64(y+h)
}

func musicHudRect() (x, y, w, h float32) {
	w = 430
	h = 36
	x = topStatusW
	y = 0
	if x+w > float32(ScreenWidth) {
		x = float32(ScreenWidth) - w
		if x < 0 {
			x = 0
		}
	}
	return x, y, w, h
}

func musicHudToggleRect() [4]float32 {
	x, y, _, _ := musicHudRect()
	return [4]float32{x + 310, y + 7, 46, 22}
}

func musicHudNextRect() [4]float32 {
	x, y, _, _ := musicHudRect()
	return [4]float32{x + 362, y + 7, 54, 22}
}

func musicHudInteractiveHit(fx, fy float64) bool {
	status := audio.MusicStatusNow()
	if !status.HasPlaylist {
		return false
	}
	return rectF32Hit(fx, fy, musicHudToggleRect()) || rectF32Hit(fx, fy, musicHudNextRect())
}

func musicHudHit(fx, fy float64) bool {
	status := audio.MusicStatusNow()
	if !status.HasPlaylist {
		return false
	}
	x, y, w, h := musicHudRect()
	return fx >= float64(x) && fx <= float64(x+w) && fy >= float64(y) && fy <= float64(y+h)
}

// ── Ana alt bar ──────────────────────────────────────────────────────

// DrawBottomPanel üst sol durum panelini, sağ üst tarih HUD'unu ve alt-orta aksiyon HUD'unu çizer.
func DrawBottomPanel(screen *ebiten.Image, gs *state.GameState, showRecruit, recruitEnabled, showDiplomacy, showTech bool, mapMode MapMode) {
	by := float32(0)
	bw := topStatusW
	if bw > float32(ScreenWidth) {
		bw = float32(ScreenWidth)
	}

	vector.FillRect(screen, 0, by, bw, topStatusH, panelBg, false)
	vector.FillRect(screen, 0, by, bw, 3, panelBorder, false)
	vector.StrokeLine(screen, 0, by+topStatusH, bw, by+topStatusH, 1.5, panelBorder, false)
	vector.StrokeLine(screen, bw, by+4, bw, by+topStatusH, 1, color.RGBA{80, 65, 35, 120}, false)

	f, hasPlayer := gs.Factions[gs.PlayerFactionID]

	// Sol blok: fraksiyon amblemi + isim
	if hasPlayer {
		fc := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		cx := float32(34)
		cy := by + bottomBarH/2
		vector.FillCircle(screen, cx, cy, 22, fc, true)
		vector.StrokeCircle(screen, cx, cy, 22, 2, panelBorder, true)
		initial := string([]rune(f.NameTR)[:1])
		DrawTextCentered(screen, initial, float64(cx), float64(cy)-8, FaceLarge, color.RGBA{255, 255, 255, 240})

		DrawText(screen, f.NameTR, 64, float64(by)+25, FaceLarge, fc)
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

		// Teknoloji bilgisi
		if f.Research.ActiveID != "" {
			if tech, ok := gs.TechTypes[f.Research.ActiveID]; ok {
				techStr := tech.NameTR + " (" + itoa(f.Research.TurnsLeft) + " tur)"
				const victoryCardX = 718.0
				maxTechW := victoryCardX - rx3 - 12
				if maxTechW < 40 {
					maxTechW = 40
				}
				techStr = trimTextToWidth(techStr, FaceSmall, maxTechW)
				DrawText(screen, techStr, rx3, ry+40, FaceSmall, color.RGBA{100, 220, 100, 255})
			}
		} else {
			DrawText(screen, "Teknoloji yok", rx3, ry+40, FaceSmall, ColorGray)
		}
	}

	// Askeri kapasite göstergesi
	if hasPlayer {
		drawManpowerDisplay(screen, gs, float64(by))
	}

	// Zafer göstergesi — kaynak sütunundan sonra başlar
	if hasPlayer {
		drawVictoryProgress(screen, gs, float64(by))
		drawVictoryAchievedBanner(screen, gs)
	}

	// Alt-orta: aksiyon HUD'u
	hudX, hudY, hudW, hudH := bottomActionHudRect()
	vector.FillRect(screen, hudX, hudY, hudW, hudH, panelBg, false)
	drawPanelBorder(screen, hudX, hudY, hudW, hudH)
	vector.FillRect(screen, hudX, hudY, hudW, 3, panelBorder, false)

	rects := BottomButtonRects()
	labels := [4]string{"Ordu", "Diplomasi", "Teknoloji", "Tur Bitir ►"}
	active := [4]bool{showRecruit, showDiplomacy, showTech, false}
	enabled := [4]bool{recruitEnabled, true, true, true}
	bgNorm := [4]color.RGBA{
		{88, 62, 30, 220},
		{40, 65, 110, 215},
		{60, 40, 95, 215},
		{40, 90, 40, 230},
	}
	bgAct := [4]color.RGBA{
		{150, 106, 48, 245},
		{80, 130, 200, 240},
		{110, 70, 170, 240},
		{70, 150, 70, 255},
	}
	for i, r := range rects {
		bg := bgNorm[i]
		txtCol := ColorWhite
		if !enabled[i] {
			bg = color.RGBA{34, 30, 24, 180}
			txtCol = color.RGBA{120, 112, 96, 210}
		}
		if active[i] {
			bg = bgAct[i]
		}
		vector.FillRect(screen, r[0], r[1], r[2], r[3], bg, false)
		vector.StrokeRect(screen, r[0], r[1], r[2], r[3], 1.5, panelBorder, false)
		tw := MeasureText(labels[i], FaceMed)
		DrawText(screen, labels[i], float64(r[0])+float64(r[2])/2-tw/2, float64(r[1])+15, FaceMed, txtCol)
	}
	drawMapModeHud(screen, mapMode)

	drawDateMenuHud(screen, gs, mapMode)
	drawMusicHud(screen)
}

func drawMapModeHud(screen *ebiten.Image, mapMode MapMode) {
	x, y, w, h := mapModeHudRect()
	vector.FillRect(screen, x, y, w, h, color.RGBA{14, 14, 18, 220}, false)
	vector.StrokeRect(screen, x, y, w, h, 1.2, panelBorder, false)
	buttons := mapModeButtonRects()
	labels := [2]string{"Normal", "Ticaret"}
	for i, b := range buttons {
		active := (i == 0 && mapMode == MapModeNormal) || (i == 1 && mapMode == MapModeTrade)
		fill := color.RGBA{44, 48, 56, 220}
		txt := color.RGBA{184, 194, 204, 220}
		if active {
			fill = color.RGBA{66, 90, 122, 240}
			txt = color.RGBA{235, 245, 255, 240}
		}
		vector.FillRect(screen, b[0], b[1], b[2], b[3], fill, false)
		vector.StrokeRect(screen, b[0], b[1], b[2], b[3], 1, color.RGBA{120, 96, 54, 210}, false)
		tw := MeasureText(labels[i], FaceSmall)
		DrawText(screen, labels[i], float64(b[0])+float64(b[2])/2-tw/2, float64(b[1])+6, FaceSmall, txt)
	}
}

func drawMusicHud(screen *ebiten.Image) {
	status := audio.MusicStatusNow()
	if !status.HasPlaylist {
		return
	}
	x, y, w, h := musicHudRect()
	vector.FillRect(screen, x, y, w, h, color.RGBA{14, 12, 9, 220}, false)
	vector.StrokeRect(screen, x, y, w, h, 1, panelBorder, false)

	track := status.Track
	if track == "" {
		track = "Playlist hazir"
	}
	track = strings.TrimSuffix(track, ".ogg")
	track = strings.TrimSuffix(track, ".mp3")
	track = strings.TrimSuffix(track, ".wav")
	label := trimTextToWidth("Muzik: "+track, FaceSmall, 292)
	DrawText(screen, label, float64(x)+10, float64(y)+11, FaceSmall, ColorGray)

	toggle := "Dur"
	if !status.Playing {
		toggle = "Cal"
	}
	tr := musicHudToggleRect()
	nr := musicHudNextRect()
	drawTinyPanelButton(screen, tr[0], tr[1], tr[2], tr[3], toggle, true)
	drawTinyPanelButton(screen, nr[0], nr[1], nr[2], nr[3], "Sonr", true)
}

func drawDateMenuHud(screen *ebiten.Image, gs *state.GameState, mapMode MapMode) {
	x, y, w, h := topDateHudRect()
	vector.FillRect(screen, x, y, w, h, panelBg, false)
	drawPanelBorder(screen, x, y, w, h)
	vector.FillRect(screen, x, y, w, 3, panelBorder, false)

	months := [...]string{"", "Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran",
		"Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık"}
	month := ""
	if gs.Month >= 1 && gs.Month <= 12 {
		month = months[gs.Month]
	}
	dateStr := month + " " + itoa(gs.Year)
	DrawText(screen, dateStr, float64(x)+12, float64(y)+13, FaceMed, ColorGold)
	DrawText(screen, gs.CurrentSeason().DisplayName()+"  •  Tur "+itoa(gs.Turn),
		float64(x)+12, float64(y)+42, FaceSmall, color.RGBA{160, 200, 100, 220})

	_ = mapMode

	bx, by, bw, bh := topDateHudMenuButtonRect()
	vector.FillRect(screen, bx, by, bw, bh, color.RGBA{45, 38, 28, 230}, false)
	vector.StrokeRect(screen, bx, by, bw, bh, 1.5, panelBorder, false)
	label := "Menü"
	tw := MeasureText(label, FaceMed)
	DrawText(screen, label, float64(bx)+float64(bw)/2-tw/2, float64(by)+8, FaceMed, ColorWhite)
}

// ── Olay Logu (sağ üst) ──────────────────────────────────────────────

// DrawEventLog sağ üst köşede son olayları kartlar halinde listeler.
func DrawEventLog(screen *ebiten.Image, events []string, collapsed bool, scroll int) {
	ex := evLogX()
	ey := evLogY()
	eh := eventLogPanelH(collapsed)

	vector.FillRect(screen, ex, ey, evLogW, eh, panelBg, false)
	drawPanelBorder(screen, ex, ey, evLogW, eh)
	vector.FillRect(screen, ex, ey, evLogW, 3, panelBorder, false)

	titleW := MeasureText("Olay Mesajları", FaceMed)
	DrawText(screen, "Olay Mesajları", float64(ex)+12, float64(ey)+8, FaceMed,
		color.RGBA{220, 190, 100, 255})
	if len(events) > 0 {
		count := "(" + itoa(len(events)) + ")"
		DrawText(screen, count, float64(ex)+18+titleW, float64(ey)+9, FaceSmall, ColorGray)
	}

	tx, ty, tw, th := eventLogToggleRect()
	vector.FillRect(screen, tx, ty, tw, th, color.RGBA{42, 34, 24, 220}, false)
	vector.StrokeRect(screen, tx, ty, tw, th, 1, panelBorder, false)
	toggleLabel := "−"
	if collapsed {
		toggleLabel = "+"
	}
	DrawTextCentered(screen, toggleLabel, float64(tx)+float64(tw)/2, float64(ty)+2, FaceMed, ColorGold)

	if collapsed {
		return
	}

	if len(events) == 0 {
		DrawTextCentered(screen, "Henüz olay yok", float64(ex)+float64(evLogW)/2, float64(ey)+58, FaceSmall,
			color.RGBA{150, 140, 120, 190})
		DrawTextCentered(screen, "Oyun olayları burada listelenir", float64(ex)+float64(evLogW)/2, float64(ey)+76, FaceSmall,
			color.RGBA{110, 105, 95, 170})
		return
	}

	visibleCount := eventLogVisibleCount()
	if scroll < 0 {
		scroll = 0
	}
	maxScroll := eventLogMaxScroll(len(events), collapsed)
	if scroll > maxScroll {
		scroll = maxScroll
	}
	for visibleIndex := 0; visibleIndex < visibleCount; visibleIndex++ {
		eventIndex := scroll + visibleIndex
		if eventIndex >= len(events) {
			break
		}
		ev := events[eventIndex]
		cardX, cardY, cardW, cardH := eventLogCardRect(visibleIndex)
		drawRoundedRect(screen, cardX, cardY, cardW, cardH, 6, color.RGBA{24, 20, 14, 225})
		vector.StrokeRect(screen, cardX, cardY, cardW, cardH, 1, color.RGBA{90, 72, 38, 210}, false)

		x, y, w, _ := eventLogCloseRect(visibleIndex)
		DrawTextCentered(screen, "X", float64(x)+float64(w)/2, float64(y)+2, FaceSmall, ColorGray)

		lines := wrapTextLines(ev, FaceSmall, float64(cardW-34))
		if len(lines) > 2 {
			lines = lines[:2]
			lines[1] = trimTextToWidth(lines[1]+"...", FaceSmall, float64(cardW-34))
		}
		for li, line := range lines {
			DrawText(screen, line, float64(cardX)+10, float64(cardY)+8+float64(li)*15, FaceSmall,
				color.RGBA{220, 210, 185, 235})
		}
	}
	drawEventLogScrollbar(screen, len(events), scroll)
}

func eventLogPanelH(collapsed bool) float32 {
	if collapsed {
		return evLogMinH
	}
	maxH := minimapY() - evLogY() - 8
	if maxH < evLogMinH {
		return evLogMinH
	}
	if evLogH > maxH {
		return maxH
	}
	return evLogH
}

func eventLogPanelHit(mx, my float64, collapsed bool) bool {
	x, y := evLogX(), evLogY()
	h := eventLogPanelH(collapsed)
	return mx >= float64(x) && mx <= float64(x+evLogW) && my >= float64(y) && my <= float64(y+h)
}

func eventLogToggleRect() (x, y, w, h float32) {
	w, h = 24, 22
	x = evLogX() + evLogW - w - 8
	y = evLogY() + 7
	return x, y, w, h
}

func eventLogToggleHit(mx, my float64, collapsed bool) bool {
	if collapsed {

	}
	x, y, w, h := eventLogToggleRect()
	return mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h)
}

func eventLogCardRect(index int) (x, y, w, h float32) {
	x = evLogX() + 8
	y = evLogY() + 31 + float32(index)*(eventCardH+eventCardGap)
	w = evLogW - 16
	h = eventCardH
	return x, y, w, h
}

func eventLogCloseRect(index int) (x, y, w, h float32) {
	cardX, cardY, cardW, _ := eventLogCardRect(index)
	w, h = 18, 18
	x = cardX + cardW - w - 5
	y = cardY + 5
	return x, y, w, h
}

func eventLogCardHit(mx, my float64, eventCount int, collapsed bool, scroll int) int {
	if collapsed {
		return -1
	}
	visibleCount := eventLogVisibleCount()
	for i := 0; i < visibleCount; i++ {
		eventIndex := scroll + i
		if eventIndex >= eventCount {
			break
		}
		x, y, w, h := eventLogCardRect(i)
		if mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h) {
			return eventIndex
		}
	}
	return -1
}

func eventLogCloseHit(mx, my float64, eventCount int, collapsed bool, scroll int) int {
	if collapsed {
		return -1
	}
	visibleCount := eventLogVisibleCount()
	for i := 0; i < visibleCount; i++ {
		eventIndex := scroll + i
		if eventIndex >= eventCount {
			break
		}
		x, y, w, h := eventLogCloseRect(i)
		if mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h) {
			return eventIndex
		}
	}
	return -1
}

func eventLogInteractiveHit(mx, my float64, eventCount int, collapsed bool, scroll int) bool {
	if eventLogToggleHit(mx, my, collapsed) {
		return true
	}
	if eventLogCloseHit(mx, my, eventCount, collapsed, scroll) >= 0 {
		return true
	}
	return eventLogCardHit(mx, my, eventCount, collapsed, scroll) >= 0
}

func eventLogVisibleCount() int {
	available := eventLogPanelH(false) - 31 - 8
	if available <= 0 {
		return 0
	}
	return int((available + eventCardGap) / (eventCardH + eventCardGap))
}

func eventLogMaxScroll(eventCount int, collapsed bool) int {
	if collapsed {
		return 0
	}
	maxScroll := eventCount - eventLogVisibleCount()
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func drawEventLogScrollbar(screen *ebiten.Image, eventCount int, scroll int) {
	visibleCount := eventLogVisibleCount()
	if eventCount <= visibleCount || visibleCount <= 0 {
		return
	}
	trackX := evLogX() + evLogW - 5
	trackY := evLogY() + 34
	trackH := eventLogPanelH(false) - 44
	vector.FillRect(screen, trackX, trackY, 2, trackH, color.RGBA{70, 58, 38, 160}, false)

	thumbH := trackH * float32(visibleCount) / float32(eventCount)
	if thumbH < 24 {
		thumbH = 24
	}
	maxScroll := eventLogMaxScroll(eventCount, false)
	thumbY := trackY
	if maxScroll > 0 {
		thumbY += (trackH - thumbH) * float32(scroll) / float32(maxScroll)
	}
	vector.FillRect(screen, trackX-1, thumbY, 4, thumbH, color.RGBA{180, 145, 70, 210}, false)
}

func eventDetailPopupRect() (x, y, w, h float32) {
	w = 620
	h = 300
	x = float32(ScreenWidth)/2 - w/2
	y = float32(ScreenHeight)/2 - h/2
	return x, y, w, h
}

func eventDetailPopupHit(mx, my float64) bool {
	x, y, w, h := eventDetailPopupRect()
	return mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h)
}

func eventDetailCloseRect() (x, y, w, h float32) {
	px, py, pw, _ := eventDetailPopupRect()
	w, h = 30, 26
	x = px + pw - w - 12
	y = py + 10
	return x, y, w, h
}

func eventDetailCloseHit(mx, my float64) bool {
	x, y, w, h := eventDetailCloseRect()
	return mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h)
}

func minimapHit(mx, my float64) bool {
	x, y := minimapX(), minimapY()
	return mx >= float64(x) && mx <= float64(x+minimapW) && my >= float64(y) && my <= float64(y+minimapH)
}

func drawEventDetailPopup(screen *ebiten.Image, message string) {
	overlay := ebiten.NewImage(int(ScreenWidth), int(ScreenHeight))
	overlay.Fill(color.RGBA{0, 0, 0, 120})
	screen.DrawImage(overlay, nil)

	px, py, pw, ph := eventDetailPopupRect()
	drawRoundedRect(screen, px, py, pw, ph, 8, panelBg)
	drawPanelBorder(screen, px, py, pw, ph)
	vector.FillRect(screen, px, py, pw, 3, panelBorder, false)

	DrawText(screen, "Olay Detayı", float64(px)+18, float64(py)+16, FaceLarge, ColorGold)

	cx, cy, cw, ch := eventDetailCloseRect()
	vector.FillRect(screen, cx, cy, cw, ch, color.RGBA{44, 34, 24, 230}, false)
	vector.StrokeRect(screen, cx, cy, cw, ch, 1, panelBorder, false)
	DrawTextCentered(screen, "X", float64(cx)+float64(cw)/2, float64(cy)+5, FaceMed, ColorGray)

	lines := wrapTextLines(message, FaceMed, float64(pw-40))
	maxLines := int((ph - 78) / 19)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines[len(lines)-1] = trimTextToWidth(lines[len(lines)-1]+"...", FaceMed, float64(pw-40))
	}
	for i, line := range lines {
		DrawText(screen, line, float64(px)+20, float64(py)+60+float64(i)*19, FaceMed,
			color.RGBA{230, 224, 205, 240})
	}
}

func drawInfoPopup(screen *ebiten.Image, message string, alpha uint8) {
	pw := float32(430)
	px := float32(ScreenWidth)/2 - pw/2
	lines := wrapTextLines(message, FaceMed, float64(pw-40))
	if len(lines) > 3 {
		lines = lines[:3]
		lines[len(lines)-1] = trimTextToWidth(lines[len(lines)-1]+"...", FaceMed, float64(pw-40))
	}
	ph := float32(48 + len(lines)*20)
	py := float32(ScreenHeight)*0.22 - ph/2

	bgAlpha := alpha
	if bgAlpha > 235 {
		bgAlpha = 235
	}
	drawRoundedRect(screen, px, py, pw, ph, 8, color.RGBA{18, 14, 10, bgAlpha})
	vector.StrokeRect(screen, px, py, pw, ph, 1.5, color.RGBA{130, 105, 55, alpha}, false)
	vector.FillRect(screen, px, py, pw, 3, color.RGBA{210, 170, 65, alpha}, false)

	DrawText(screen, "Bilgi", float64(px)+16, float64(py)+12, FaceSmall, color.RGBA{220, 190, 100, alpha})
	for i, line := range lines {
		DrawText(screen, line, float64(px)+20, float64(py)+34+float64(i)*20, FaceMed,
			color.RGBA{240, 230, 205, alpha})
	}
}

// ── Minimap (sağ alt, alt kenara yapışık) ────────────────────────────

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

	// Ordu konumları katmanı
	scaleX := float64(minimapW) / float64(WorldW)
	scaleY := float64(minimapH) / float64(WorldH)
	drawMinimapArmies(screen, gs, float32(scaleX), float32(scaleY), mx, my)

	// İç kenara ince koyu çizgi
	vector.StrokeRect(screen, mx, my, minimapW, minimapH, 1, color.RGBA{30, 25, 15, 200}, false)

	// Köşe süslemeleri
	drawMinimapCorner(screen, mx, my, cornerSize, cornerSize)
	drawMinimapCorner(screen, mx+minimapW, my, -cornerSize, cornerSize)
	drawMinimapCorner(screen, mx, my+minimapH, cornerSize, -cornerSize)
	drawMinimapCorner(screen, mx+minimapW, my+minimapH, -cornerSize, -cornerSize)

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

// drawMinimapArmies ordu konumlarını minimap üzerinde fraksiyon rengiyle gösterir.
// Kara ordusu → kare, deniz donanması → daire.
func drawMinimapArmies(screen *ebiten.Image, gs *state.GameState, scaleX, scaleY, offsetX, offsetY float32) {
	playerID := gs.PlayerFactionID
	for _, a := range gs.Armies {
		region, ok := gs.Regions[a.RegionID]
		if !ok {
			continue
		}
		px := offsetX + float32(wcX(region.WorldX))*scaleX
		py := offsetY + float32(wcY(region.WorldY))*scaleY

		if faction.FactionID(a.OwnerID) == playerID {
			// Oyuncu orduları/donanmaları minimap'te yeşil nokta olarak gösterilir.
			vector.FillCircle(screen, px+1, py+1, 3.5, color.RGBA{0, 0, 0, 90}, true)
			vector.FillCircle(screen, px, py, 2.5, color.RGBA{80, 220, 120, 240}, true)
			vector.StrokeCircle(screen, px, py, 2.5, 1, color.RGBA{20, 60, 30, 220}, true)
			continue
		}
		rel := gs.Relations[faction.RelationKey(playerID, faction.FactionID(a.OwnerID))]
		if rel == nil || (rel.Stance != faction.StanceWar && rel.Stance != faction.StanceAllied) {
			continue
		}

		col := factionColor(gs, a.OwnerID)
		col.A = 220

		borderCol := color.RGBA{0, 0, 0, 100}

		if a.IsNaval {
			r := float32(4)
			vector.FillCircle(screen, px+1, py+1, r+1, color.RGBA{0, 0, 0, 80}, true)
			vector.FillCircle(screen, px, py, r, col, true)
			vector.StrokeCircle(screen, px, py, r, 1.2, borderCol, true)
		} else {
			h := float32(3.5)
			vector.FillRect(screen, px-h-1, py-h-1, h*2+2, h*2+2, color.RGBA{0, 0, 0, 80}, false)
			vector.FillRect(screen, px-h, py-h, h*2, h*2, col, false)
			vector.StrokeRect(screen, px-h-0.5, py-h-0.5, h*2+1, h*2+1, 1.2, borderCol, false)
		}
	}
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
	if !ok {
		return
	}

	if region.IsSea {
		DrawSeaRegionPanel(screen, gs, region)
		return
	}

	px := infoPanelX()
	py := infoPanelY()
	pw := infoPanelW
	ph := infoPanelH

	vector.FillRect(screen, px, py, pw, ph, panelBg, false)
	drawPanelBorder(screen, px, py, pw, ph)
	vector.FillRect(screen, px, py, pw, 3, panelBorder, false)
	drawPanelCloseButton(screen, px, py, pw)

	lx := float64(px) + panelPad
	ly := float64(py) + 10

	DrawText(screen, region.NameTR, lx, ly, FaceLarge, ColorYellow)
	ly += 24

	// Development mode bilgileri
	if gs.DevelopmentMode {
		DrawText(screen, "ID: "+string(region.ID), lx, ly, FaceSmall, ColorGray)
		ly += 16
		DrawText(screen, "Koordinat: "+itoa(region.WorldX)+","+itoa(region.WorldY), lx, ly, FaceSmall, ColorGray)
		ly += 18
	}

	ownerName, ownerCol := ownerDisplay(gs, region.OwnerID)
	ownerLine := "Sahip: " + ownerName
	if region.IsLocked {
		if region.UnlockTurn > 0 {
			ownerLine += "  LOCK Tur " + itoa(region.UnlockTurn)
		} else {
			ownerLine += "  LOCK Kilitli"
		}
	} else if region.UnlockTurn > 0 {
		ownerLine += "  ACIK"
	}
	DrawText(screen, ownerLine, lx, ly, FaceSmall, ownerCol)
	ly += 18

	var stypeStr string
	if len(region.Settlements) > 0 {
		capital := region.Settlements[0]
		for _, s := range region.Settlements {
			if s.IsCapital {
				capital = s
				break
			}
		}
		stypeStr = "  |  " + settlementTypeLabel(capital.Type)
	}

	DrawText(screen, terrainLabel(region.Terrain)+"  |  "+religionLabel(string(region.Religion))+stypeStr, lx, ly, FaceSmall, ColorGray)
	ly += 16

	sepW := pw - float32(panelPad*2)
	vector.StrokeLine(screen, float32(lx), float32(ly), float32(lx)+sepW, float32(ly), 1, panelBorder, false)
	ly += 8

	// Kaynaklar — iki sütun
	DrawText(screen, "G "+itoa(region.GoldIncome())+" Altin", lx, ly, FaceSmall, ColorGold)
	DrawText(screen, "T "+itoa(region.BaseGrainOutput)+" Tahil", lx+120, ly, FaceSmall, ColorWhite)
	ly += 18

	statBarX := float32(lx) + 122
	DrawText(screen, "Memnuniyet: "+itoa(region.Satisfaction)+"%", lx, ly, FaceSmall, ColorGray)
	drawBar(screen, statBarX, float32(ly)+1, sepW-(statBarX-float32(lx)), 9, float64(region.Satisfaction)/100,
		satisfactionColor(region.Satisfaction))
	ly += 18

	DrawText(screen, "Vergi: %"+itoa(region.TaxRate), lx, ly, FaceSmall, ColorGray)
	taxBarW := sepW - (statBarX - float32(lx))
	if region.OwnerID == string(gs.PlayerFactionID) && !region.IsLocked {
		dec, inc := regionTaxButtonRects(gs)
		taxBarW = dec[0] - statBarX - 8
		drawBar(screen, statBarX, float32(ly)+1, taxBarW, 9, float64(region.TaxRate)/100,
			color.RGBA{200, 140, 40, 255})
		drawTinyPanelButton(screen, dec[0], dec[1], dec[2], dec[3], "-", true)
		drawTinyPanelButton(screen, inc[0], inc[1], inc[2], inc[3], "+", true)
	} else {
		drawBar(screen, statBarX, float32(ly)+1, taxBarW, 9, float64(region.TaxRate)/100,
			color.RGBA{200, 140, 40, 255})
	}
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

	if gs.DevelopmentMode {
		// Komşu bölgeler (deniz paneliyle aynı görünüm)
		neighborTitle := "Komşu Bölgeler:"
		if len(region.Neighbors) == 0 {
			neighborTitle = "Komşu: Yok"
		} else if len(region.Neighbors) <= 5 {
			neighborTitle = "Komşu (" + itoa(len(region.Neighbors)) + "):"
		} else {
			neighborTitle = "Komşu (" + itoa(len(region.Neighbors)) + ") [gösterilen: 4]"
		}
		DrawText(screen, neighborTitle, lx, ly, FaceSmall, color.RGBA{200, 170, 90, 220})
		ly += 18

		displayCount := len(region.Neighbors)
		if displayCount > 4 {
			displayCount = 4
		}
		for i := 0; i < displayCount; i++ {
			neighborID := region.Neighbors[i]
			neighborRegion, ok := gs.Regions[neighborID]
			if !ok {
				continue
			}
			col := color.RGBA{180, 180, 180, 200}
			if neighborRegion.IsSea {
				col = color.RGBA{100, 160, 220, 200}
			}
			DrawText(screen, "• "+neighborRegion.NameTR, lx+15, ly, FaceSmall, col)
			ly += 16
		}
	}

	// ── Binalar bölümü ────────────────────────────────────────────────
	ly += 4
	vector.StrokeLine(screen, float32(lx), float32(ly), float32(lx)+sepW, float32(ly), 1, panelBorder, false)
	ly += 6

	bldTitleW := MeasureText("BİNALAR", FaceSmall)
	DrawText(screen, "BİNALAR", float64(px)+float64(pw)/2-bldTitleW/2, ly, FaceSmall, color.RGBA{200, 170, 90, 220})
	ly += 17

	drawBuildingGrid(screen, gs, region, px, float32(ly), pw)

	if region.OwnerID != "" && region.OwnerID != string(gs.PlayerFactionID) {
		drawRegionDiplomacyButtons(screen, gs, region.OwnerID, px, py, pw, ph)
	}
}

func regionDiplomacyButtonRect(i int, px, py, pw, ph float32) (x, y, w, h float32) {
	btnW := float32(70)
	btnH := float32(20)
	gap := float32(6)
	totalW := btnW*4 + gap*3
	x = px + pw - totalW - 5 + float32(i)*(btnW+gap)
	y = py + ph - btnH - 8
	return x, y, btnW, btnH
}

func drawRegionDiplomacyButtons(screen *ebiten.Image, gs *state.GameState, ownerID string, px, py, pw, ph float32) {
	labels := []string{"Savaş", "Barış", "İttifak", "Ticaret"}
	colors := []color.RGBA{
		{180, 50, 50, 220},
		{50, 120, 180, 220},
		{50, 160, 80, 220},
		{160, 130, 50, 220},
	}
	for i := 0; i < 4; i++ {
		x, y, w, h := regionDiplomacyButtonRect(i, px, py, pw, ph)
		active := regionDiplomacyButtonDisabledReason(gs, ownerID, i) == ""
		btnCol := colors[i]
		txtCol := ColorWhite
		if !active {
			btnCol.A = 110
			txtCol = ColorGray
		}
		vector.FillRect(screen, x, y, w, h, btnCol, false)
		vector.StrokeRect(screen, x, y, w, h, 1, panelBorder, false)
		tw := MeasureText(labels[i], FaceSmall)
		DrawText(screen, labels[i], float64(x)+float64(w)/2-tw/2, float64(y)+4, FaceSmall, txtCol)
	}
}

func regionDiplomacyButtonDisabledReason(gs *state.GameState, ownerID string, idx int) string {
	if gs == nil || ownerID == "" || idx < 0 || idx > 3 {
		return ""
	}
	rel := gs.Relations[faction.RelationKey(gs.PlayerFactionID, faction.FactionID(ownerID))]
	if rel == nil {
		return ""
	}
	switch idx {
	case 0:
		if rel.Stance == faction.StanceWar {
			return "Zaten savaş halindesin."
		}
	case 1:
		if rel.Stance != faction.StanceWar {
			return "Barış teklifi sadece savaşta yapılır."
		}
	case 2:
		if rel.Stance == faction.StanceWar {
			return "Savaş halindeyken ittifak teklif edilemez."
		}
		if rel.Stance == faction.StanceAllied {
			return "Zaten müttefiksin."
		}
	case 3:
		if rel.Stance == faction.StanceWar {
			return "Savaş halindeyken ticaret teklif edilemez."
		}
		if rel.Stance == faction.StanceTrade {
			return "Zaten ticaret anlaşması aktif."
		}
	}
	return ""
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
	drawPanelCloseButton(screen, px, py, pw)

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
	stats := "Tur: " + itoa(gs.Turn) + "  |  Yil: " + itoa(gs.Year) +
		"  |  Bolge: " + itoa(regionCount) + "  |  Ordu: " + itoa(armyCount)
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
	DrawTextCentered(screen, "- TARIHSEL OLAY -", ScreenWidth/2, cy, FaceSmall, color.RGBA{180, 140, 50, 200})
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
	case state.VictoryConquerCity:
		return "Fetih Zaferi"
	}
	return "Zafer"
}

// ── Zafer İlerleme Göstergesi ─────────────────────────────────────────

// drawManpowerDisplay savaşçı kapasitesini ve ordu sayısını gösterir.
func drawManpowerDisplay(screen *ebiten.Image, gs *state.GameState, panelY float64) {
	pid := gs.PlayerFactionID
	deployed := gs.DeployedLandUnits(pid)
	cap := gs.ManpowerCap(pid)
	armies := gs.CurrentLandArmies(pid)
	maxArmies := gs.MaxLandArmies(pid)

	cardX := float32(878)
	cardY := float32(panelY) + 7
	cardW := float32(112)
	cardH := topStatusH - 14
	drawTopStatusCard(screen, cardX, cardY, cardW, cardH)

	mx := float64(cardX) + 12
	my := panelY + 16

	DrawText(screen, "Savaşçı", mx, my, FaceSmall, ColorGray)
	unitStr := itoa(deployed) + "/" + itoa(cap)
	unitCol := ColorGold
	if cap > 0 && deployed >= cap {
		unitCol = ColorRed
	}
	unitW := MeasureText(unitStr, FaceMed)
	DrawText(screen, unitStr, float64(cardX+cardW)-12-unitW, my, FaceMed, unitCol)

	DrawText(screen, "Ordu", mx, my+28, FaceSmall, ColorGray)
	armyStr := itoa(armies) + "/" + itoa(maxArmies)
	armyCol := ColorGold
	if armies >= maxArmies {
		armyCol = ColorRed
	}
	armyW := MeasureText(armyStr, FaceMed)
	DrawText(screen, armyStr, float64(cardX+cardW)-12-armyW, my+28, FaceMed, armyCol)
}

// drawVictoryProgress seçilen zafer tipine göre ilerlemeyi gösterir.
func drawVictoryProgress(screen *ebiten.Image, gs *state.GameState, panelY float64) {
	if gs.PlayerFactionID == "" {
		return
	}

	cardX := float32(718)
	cardY := float32(panelY) + 7
	cardW := float32(150)
	cardH := topStatusH - 14
	drawTopStatusCard(screen, cardX, cardY, cardW, cardH)

	vx := float64(cardX) + 12
	vy := panelY + 14
	barW := cardW - 24
	barX := cardX + 12

	titleCol := color.RGBA{220, 190, 100, 220}
	DrawText(screen, "Zafer Hedefi", vx, vy, FaceSmall, titleCol)
	vy += 17

	switch gs.Victory.Type {
	case state.VictoryDomination, "":
		target := gs.Victory.TargetRegionCount
		if target == 0 {
			target = 20
		}
		current := len(gs.RegionsOwnedBy(gs.PlayerFactionID))
		DrawText(screen, "Hedef: "+itoa(current)+"/"+itoa(target), vx, vy, FaceMed, ColorWhite)
		vy += 18
		drawTopProgressBar(screen, barX, float32(vy), barW, 7, clampF(float64(current)/float64(target)), ColorGold)

	case state.VictoryEconomic:
		threshold := gs.Victory.TargetGoldIncome
		if threshold == 0 {
			threshold = 500
		}
		holdTurns := gs.Victory.GoldHoldTurns
		if holdTurns == 0 {
			holdTurns = 5
		}
		goldIncome := victory.CurrentGoldIncome(gs)
		DrawText(screen, "Gelir: "+itoa(goldIncome)+"/"+itoa(threshold), vx, vy, FaceMed, ColorGold)
		vy += 18
		drawTopProgressBar(screen, barX, float32(vy), barW, 7, clampF(float64(goldIncome)/float64(threshold)), ColorGold)
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
		DrawText(screen, "Güç: "+itoa(totalStr)+"/"+itoa(targetStr), vx, vy, FaceMed, ColorWhite)
		vy += 18
		drawTopProgressBar(screen, barX, float32(vy), barW, 7, clampF(float64(totalStr)/float64(targetStr)), color.RGBA{200, 80, 80, 255})
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
		DrawText(screen, "Kutsal: "+itoa(held)+"/"+itoa(total), vx, vy, FaceMed, color.RGBA{200, 160, 255, 255})
		vy += 18
		drawTopProgressBar(screen, barX, float32(vy), barW, 7, clampF(float64(held)/float64(total+1)), color.RGBA{160, 120, 255, 255})
		vy += 12
		DrawText(screen, itoa(gs.ReligiousVictoryTurns)+"/12 tur", vx, vy, FaceSmall, ColorGray)

	case state.VictoryConquerCity:
		held := 0
		total := len(gs.Victory.RequiredRegions)
		if total == 0 {
			return
		}
		for _, rid := range gs.Victory.RequiredRegions {
			if r, ok := gs.Regions[rid]; ok && r.OwnerID == string(gs.PlayerFactionID) {
				held++
			}
		}
		DrawText(screen, "Hedef: "+itoa(held)+"/"+itoa(total), vx, vy, FaceMed, ColorWhite)
		vy += 18
		drawTopProgressBar(screen, barX, float32(vy), barW, 7, clampF(float64(held)/float64(total)), ColorGold)
	}
}

func drawTopStatusCard(screen *ebiten.Image, x, y, w, h float32) {
	vector.FillRect(screen, x, y, w, h, color.RGBA{18, 16, 12, 150}, false)
	vector.FillRect(screen, x, y, w, 1, color.RGBA{170, 135, 60, 80}, false)
	vector.StrokeRect(screen, x, y, w, h, 1, color.RGBA{95, 78, 42, 115}, false)
}

func drawTopProgressBar(screen *ebiten.Image, x, y, w, h float32, fill float64, col color.Color) {
	fill = clampF(fill)
	vector.FillRect(screen, x, y, w, h, color.RGBA{42, 42, 40, 210}, false)
	if fill > 0 {
		vector.FillRect(screen, x, y, float32(float64(w)*fill), h, col, false)
	}
	vector.StrokeRect(screen, x, y, w, h, 1, color.RGBA{120, 100, 55, 150}, false)
}

func drawVictoryAchievedBanner(screen *ebiten.Image, gs *state.GameState) {
	if gs == nil || !gs.VictoryAchieved || gs.WinnerID != gs.PlayerFactionID {
		return
	}
	msg := "Kalıcı Olay: " + victoryTypeLabel(gs.Victory.Type) + " gerçekleşti (Tur " + itoa(gs.VictoryAchievedTurn) + ")"
	maxW := ScreenWidth - 320
	if maxW < 260 {
		maxW = 260
	}
	msg = trimTextToWidth(msg, FaceSmall, maxW)
	w := float32(MeasureText(msg, FaceSmall) + 24)
	if w > float32(maxW)+24 {
		w = float32(maxW) + 24
	}
	h := float32(24)
	x := float32(ScreenWidth)/2 - w/2
	y := topStatusH + 6
	vector.FillRect(screen, x, y, w, h, color.RGBA{42, 34, 16, 220}, false)
	vector.StrokeRect(screen, x, y, w, h, 1, color.RGBA{190, 150, 70, 230}, false)
	DrawText(screen, msg, float64(x)+12, float64(y)+6, FaceSmall, color.RGBA{245, 215, 140, 255})
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

	builtCount := make(map[string]int, len(region.Buildings))
	for _, bid := range region.Buildings {
		builtCount[bid]++
	}
	queuedSet := make(map[string]int)
	queuedTurnsMin := make(map[string]int)
	for _, order := range gs.ProductionQueue {
		if order.Kind == "building" && order.RegionID == region.ID {
			queuedSet[order.TypeID]++
			if queuedTurnsMin[order.TypeID] == 0 || order.TurnsLeft < queuedTurnsMin[order.TypeID] {
				queuedTurnsMin[order.TypeID] = order.TurnsLeft
			}
		}
	}

	const cols = 3
	pad := float32(panelPad)
	availW := panelW - pad*2
	slotW := availW / float32(cols)
	spriteH := float32(76)
	nameH := float32(18)
	rowH := spriteH + nameH + 7

	display := visibleBuildingIDs(gs, region)
	for i, bid := range display {
		col := i % cols
		row := i / cols

		sx := panelX + pad + float32(col)*slotW
		sy := startY + float32(row)*rowH
		innerW := slotW - 3

		b, hasDef := gs.BuildingTypes[bid]
		level := builtCount[bid]
		queuedCount := queuedSet[bid]
		turnsLeft := queuedTurnsMin[bid]
		isQueued := queuedCount > 0
		canAfford := false
		if f := gs.Factions[gs.PlayerFactionID]; f != nil && hasDef {
			canAfford = f.Gold >= b.GoldCost
		}
		isBuilt := level > 0
		maxLevel := 1
		if hasDef && b.MaxPerRegion > 0 {
			maxLevel = b.MaxPerRegion
		}
		isMaxLevel := level >= maxLevel

		// Arka plan ve çerçeve
		slotBg := color.RGBA{250, 250, 250, 240}
		borderCol := color.RGBA{160, 160, 160, 220}
		switch {
		case isBuilt:
			slotBg = color.RGBA{255, 255, 255, 245}
			borderCol = color.RGBA{150, 130, 85, 230}
		case isQueued:
			slotBg = color.RGBA{248, 248, 248, 240}
			borderCol = color.RGBA{145, 145, 145, 220}
		case canAfford:
			slotBg = color.RGBA{252, 252, 252, 242}
			borderCol = color.RGBA{165, 165, 165, 220}
		}
		vector.FillRect(screen, sx, sy, innerW, spriteH, slotBg, false)
		vector.StrokeRect(screen, sx, sy, innerW, spriteH, 1, borderCol, false)

		if buildingSheet != nil {
			r := buildingSpriteRect(bid, buildingSheet)
			sub := buildingSheet.SubImage(r).(*ebiten.Image)
			op := &ebiten.DrawImageOptions{}
			fitW := float64(innerW - 2)
			fitH := float64(spriteH - 2)
			scale := fitW / float64(r.Dx())
			if hScale := fitH / float64(r.Dy()); hScale < scale {
				scale = hScale
			}
			drawW := float64(r.Dx()) * scale
			drawH := float64(r.Dy()) * scale
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(
				float64(sx)+float64(innerW)/2-drawW/2,
				float64(sy)+float64(spriteH)/2-drawH/2,
			)
			if !isBuilt {
				if canAfford {
					op.ColorScale.Scale(0.65, 0.65, 0.65, 0.9)
				} else {
					op.ColorScale.Scale(0.35, 0.35, 0.35, 0.85)
				}
			}
			screen.DrawImage(sub, op)

			if isBuilt {
				vector.StrokeRect(screen, sx+1, sy+1, innerW-2, spriteH-2, 1, color.RGBA{160, 130, 50, 120}, false)
				lvText := "Lv" + itoa(level)
				lvX := float64(sx) + 6
				lvY := float64(sy) + 4
				lvW := float32(MeasureText(lvText, FaceSmall) + 8)
				lvH := float32(14)
				vector.FillRect(screen, float32(lvX)-3, float32(lvY)-2, lvW, lvH, color.RGBA{18, 14, 8, 225}, false)
				vector.StrokeRect(screen, float32(lvX)-3, float32(lvY)-2, lvW, lvH, 1, color.RGBA{170, 140, 75, 220}, false)
				DrawText(screen, lvText, lvX, lvY, FaceSmall, color.RGBA{255, 245, 220, 250})
			}
			if isQueued {
				qLabel := itoa(turnsLeft) + " Tur"
				if queuedCount > 1 {
					qLabel = "x" + itoa(queuedCount) + " " + qLabel
				}
				DrawTextCentered(screen, qLabel,
					float64(sx)+float64(innerW)/2, float64(sy)+float64(spriteH)/2-7,
					FaceSmall, color.RGBA{235, 210, 125, 235})
			}
		}

		// Bina adı
		bname := bid
		if b, ok := gs.BuildingTypes[bid]; ok {
			bname = b.NameTR
		}
		nameCol := color.RGBA{75, 65, 50, 200}
		switch {
		case isBuilt:
			nameCol = ColorGold
		case isQueued:
			nameCol = color.RGBA{210, 190, 120, 230}
		case canAfford:
			nameCol = color.RGBA{170, 145, 85, 220}
		}
		DrawTextCentered(screen, bname, float64(sx)+float64(innerW)/2, float64(sy+spriteH)+3, FaceSmall, nameCol)
		if isMaxLevel {
			DrawTextCentered(screen, "Maks", float64(sx)+float64(innerW)/2, float64(sy+spriteH)+16, FaceSmall, color.RGBA{170, 155, 95, 210})
		}
	}
}

func visibleBuildingIDs(gs *state.GameState, region *world.Region) []string {
	builtCount := make(map[string]int, len(region.Buildings))
	for _, bid := range region.Buildings {
		builtCount[bid]++
	}
	ids := make([]string, 0, len(buildingDisplayOrder))
	for _, bid := range buildingDisplayOrder {
		b, ok := gs.BuildingTypes[bid]
		if !ok {
			continue
		}
		if builtCount[bid] > 0 || buildingVisibleByRegionRules(gs, region, bid, b) {
			ids = append(ids, bid)
		}
	}
	return ids
}

func drawPanelCloseButton(screen *ebiten.Image, px, py, pw float32) {
	x, y, w, h := panelCloseRect(px, py, pw)
	drawTinyPanelButton(screen, x, y, w, h, "X", true)
}

func panelCloseRect(px, py, pw float32) (x, y, w, h float32) {
	return px + pw - 24, py + 6, 18, 18
}

func drawTinyPanelButton(screen *ebiten.Image, x, y, w, h float32, label string, active bool) {
	bg := color.RGBA{34, 26, 15, 230}
	border := panelBorder
	txt := ColorGold
	if !active {
		bg = color.RGBA{18, 16, 12, 180}
		border = color.RGBA{45, 38, 25, 160}
		txt = color.RGBA{85, 78, 62, 190}
	}
	vector.FillRect(screen, x, y, w, h, bg, false)
	vector.StrokeRect(screen, x, y, w, h, 1, border, false)
	tw := MeasureText(label, FaceSmall)
	DrawText(screen, label, float64(x)+float64(w)/2-tw/2, float64(y)+2, FaceSmall, txt)
}

func panelCloseHit(mx, my float64, px, py, pw float32) bool {
	x, y, w, h := panelCloseRect(px, py, pw)
	return mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h)
}

func regionPanelHit(mx, my float64) bool {
	px := infoPanelX()
	py := infoPanelY()
	return mx >= float64(px) && mx <= float64(px+infoPanelW) && my >= float64(py) && my <= float64(py+infoPanelH)
}

func regionPanelCloseHit(mx, my float64) bool {
	px := infoPanelX()
	return panelCloseHit(mx, my, px, infoPanelY(), infoPanelW)
}

func regionPanelInteractiveHit(mx, my float64, gs *state.GameState, rid world.RegionID) bool {
	if rid == "" {
		return false
	}
	if regionPanelCloseHit(mx, my) {
		return true
	}
	if delta := regionTaxButtonHit(mx, my, gs, rid); delta != 0 {
		return true
	}
	if idx := regionDiplomacyButtonHit(mx, my, gs, rid); idx >= 0 {
		return true
	}
	return buildingGridHitTest(mx, my, gs, rid) != ""
}

// regionDiplomacyButtonHit oyuncuya ait olmayan bölge panelindeki hızlı diplomasi butonunu döner.
// 0=Savaş, 1=Barış, 2=İttifak, 3=Ticaret, -1=hiçbiri.
func regionDiplomacyButtonHit(mx, my float64, gs *state.GameState, rid world.RegionID) int {
	if rid == "" || gs == nil {
		return -1
	}
	region, ok := gs.Regions[rid]
	if !ok || region.IsSea || region.OwnerID == "" || region.OwnerID == string(gs.PlayerFactionID) {
		return -1
	}
	px, py, pw, ph := infoPanelX(), infoPanelY(), infoPanelW, infoPanelH
	for i := 0; i < 4; i++ {
		x, y, w, h := regionDiplomacyButtonRect(i, px, py, pw, ph)
		if mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h) {
			return i
		}
	}
	return -1
}

func armyPanelCloseHit(mx, my float64) bool {
	px := infoPanelX()
	py := infoPanelY() + infoPanelH - 130
	return panelCloseHit(mx, my, px, py, infoPanelW)
}

func regionTaxButtonHit(mx, my float64, gs *state.GameState, rid world.RegionID) int {
	region, ok := gs.Regions[rid]
	if !ok || region.IsSea || region.IsLocked || region.OwnerID != string(gs.PlayerFactionID) {
		return 0
	}
	dec, inc := regionTaxButtonRects(gs)
	if rectF32Hit(mx, my, dec) {
		return -5
	}
	if rectF32Hit(mx, my, inc) {
		return 5
	}
	return 0
}

func regionTaxButtonRects(gs *state.GameState) ([4]float32, [4]float32) {
	px := infoPanelX()
	py := infoPanelY()
	pw := infoPanelW
	ly := float64(py) + 10
	ly += 24
	if gs.DevelopmentMode {
		ly += 16 + 16 + 18
	}
	ly += 18 + 16 + 8 + 18 + 18
	y := float32(ly - 17)
	return [4]float32{px + pw - 70, y, 26, 18}, [4]float32{px + pw - 38, y, 26, 18}
}

func rectF32Hit(mx, my float64, r [4]float32) bool {
	return mx >= float64(r[0]) && mx <= float64(r[0]+r[2]) && my >= float64(r[1]) && my <= float64(r[1]+r[3])
}

func BuildingGridHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) string {
	return buildingGridHitTest(mx, my, gs, rid)
}

func buildingGridHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) string {
	if rid == "" {
		return ""
	}
	region, ok := gs.Regions[rid]
	if !ok || region.IsSea || region.IsLocked || region.OwnerID != string(gs.PlayerFactionID) {
		return ""
	}
	px := infoPanelX()
	pw := infoPanelW
	startY := buildingGridStartY(gs, region)

	const cols = 3
	pad := float32(panelPad)
	availW := pw - pad*2
	slotW := availW / float32(cols)
	spriteH := float32(76)
	nameH := float32(18)
	rowH := spriteH + nameH + 7

	displayIdx := 0
	for _, bid := range buildingDisplayOrder {
		if !buildingVisibleInRegion(gs, region, bid) {
			continue
		}
		if regionHasBuilding(region, bid) {
			displayIdx++
			continue
		}
		col := displayIdx % cols
		row := displayIdx / cols
		sx := px + pad + float32(col)*slotW
		sy := startY + float32(row)*rowH
		innerW := slotW - 3
		if mx >= float64(sx) && mx <= float64(sx+innerW) && my >= float64(sy) && my <= float64(sy+spriteH+nameH) {
			return bid
		}
		displayIdx++
	}
	return ""
}

func regionHasBuilding(region *world.Region, bid string) bool {
	for _, builtID := range region.Buildings {
		if builtID == bid {
			return true
		}
	}
	return false
}

func buildingVisibleInRegion(gs *state.GameState, region *world.Region, bid string) bool {
	b, ok := gs.BuildingTypes[bid]
	if !ok {
		return false
	}
	return regionHasBuilding(region, bid) || buildingVisibleByRegionRules(gs, region, bid, b)
}

func buildingVisibleByRegionRules(gs *state.GameState, region *world.Region, bid string, b *city.Building) bool {
	if bid == "port" && !region.IsCoastal(gs.Regions) {
		return false
	}
	return b.RequiredTerrain == "" || string(region.Terrain) == b.RequiredTerrain
}

func buildingGridStartY(gs *state.GameState, region *world.Region) float32 {
	py := infoPanelY()
	ly := float64(py) + 10
	ly += 24
	if gs.DevelopmentMode {
		ly += 16 + 16 + 18
	}
	ly += 18 + 16 + 8 + 18 + 18 + 18
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
			ly += 14 + 12
		}
	}
	if region.IsRebellionRisk() {
		ly += 18
	}
	if gs.DevelopmentMode {
		ly += 18 // Komşu başlığı
		displayCount := len(region.Neighbors)
		if displayCount > 4 {
			displayCount = 4
		}
		ly += float64(displayCount) * 16
	}
	ly += 4 + 6 + 17
	return float32(ly)
}

func drawPanelBorder(screen *ebiten.Image, x, y, w, h float32) {
	vector.StrokeLine(screen, x, y, x+w, y, 1.5, panelBorder, false)
	vector.StrokeLine(screen, x, y+h, x+w, y+h, 1.5, panelBorder, false)
	vector.StrokeLine(screen, x, y, x, y+h, 1.5, panelBorder, false)
	vector.StrokeLine(screen, x+w, y, x+w, y+h, 1.5, panelBorder, false)
}

func drawRoundedRect(screen *ebiten.Image, x, y, w, h, r float32, col color.Color) {
	if r <= 0 {
		vector.FillRect(screen, x, y, w, h, col, false)
		return
	}
	if r*2 > w {
		r = w / 2
	}
	if r*2 > h {
		r = h / 2
	}
	vector.FillRect(screen, x+r, y, w-r*2, h, col, false)
	vector.FillRect(screen, x, y+r, w, h-r*2, col, false)
	vector.FillCircle(screen, x+r, y+r, r, col, false)
	vector.FillCircle(screen, x+w-r, y+r, r, col, false)
	vector.FillCircle(screen, x+r, y+h-r, r, col, false)
	vector.FillCircle(screen, x+w-r, y+h-r, r, col, false)
}

func wrapTextLines(s string, face *text.GoTextFace, maxWidth float64) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	lines := make([]string, 0, 3)
	line := words[0]
	for _, word := range words[1:] {
		candidate := line + " " + word
		if MeasureText(candidate, face) <= maxWidth {
			line = candidate
			continue
		}
		lines = append(lines, line)
		line = word
	}
	lines = append(lines, line)

	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		if MeasureText(ln, face) <= maxWidth {
			out = append(out, ln)
			continue
		}
		out = append(out, splitLongWord(ln, face, maxWidth)...)
	}
	return out
}

func splitLongWord(s string, face *text.GoTextFace, maxWidth float64) []string {
	runes := []rune(s)
	lines := []string{}
	start := 0
	for start < len(runes) {
		end := start + 1
		for end <= len(runes) && MeasureText(string(runes[start:end]), face) <= maxWidth {
			end++
		}
		if end == start+1 {
			lines = append(lines, string(runes[start:end]))
			start = end
			continue
		}
		lines = append(lines, string(runes[start:end-1]))
		start = end - 1
	}
	return lines
}

func trimTextToWidth(s string, face *text.GoTextFace, maxWidth float64) string {
	if MeasureText(s, face) <= maxWidth {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && MeasureText(string(runes), face) > maxWidth {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
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

func settlementTypeLabel(t world.SettlementType) string {
	switch t {
	case world.SettlementCity:
		return "Şehir"
	case world.SettlementTown:
		return "Kasaba"
	case world.SettlementFortress:
		return "Kale"
	case world.SettlementPort:
		return "Liman"
	default:
		return string(t)
	}
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

// DrawSeaRegionPanel deniz bölgesi bilgisini sol altta gösterir.
func DrawSeaRegionPanel(screen *ebiten.Image, gs *state.GameState, region *world.Region) {
	px := infoPanelX()
	py := infoPanelY()
	pw := infoPanelW
	ph := infoPanelH

	// Panel arka plan
	vector.FillRect(screen, px, py, pw, ph, panelBg, false)
	drawPanelBorder(screen, px, py, pw, ph)
	vector.FillRect(screen, px, py, pw, 3, panelBorder, false)
	drawPanelCloseButton(screen, px, py, pw)

	lx := float64(px) + panelPad
	ly := float64(py) + 10

	// Başlık
	DrawText(screen, region.NameTR, lx, ly, FaceLarge, color.RGBA{100, 180, 255, 255})
	ly += 24

	// Development mode bilgileri
	if gs.DevelopmentMode {
		DrawText(screen, "ID: "+string(region.ID), lx, ly, FaceSmall, ColorGray)
		ly += 16
		DrawText(screen, "Koordinat: "+itoa(region.WorldX)+","+itoa(region.WorldY), lx, ly, FaceSmall, ColorGray)
		ly += 18
	}

	// Deniz bölgesi (italik vurgu)
	DrawText(screen, "Deniz Bölgesi", lx, ly, FaceSmall, color.RGBA{120, 160, 200, 200})
	ly += 18
	if region.IsLocked {
		lockLine := "Durum: Kilitli"
		if region.UnlockTurn > 0 {
			lockLine += "  |  Acilis Turu: " + itoa(region.UnlockTurn)
		}
		DrawText(screen, lockLine, lx, ly, FaceSmall, color.RGBA{220, 150, 90, 220})
		ly += 18
	}

	sepW := pw - float32(panelPad*2)
	vector.StrokeLine(screen, float32(lx), float32(ly), float32(lx)+sepW, float32(ly), 1, panelBorder, false)
	ly += 8

	// Komşu bölgeler
	neighborTitle := "Komşu Bölgeler:"
	if len(region.Neighbors) == 0 {
		neighborTitle = "Komşu: Yok"
	} else if len(region.Neighbors) <= 5 {
		neighborTitle = "Komşu (" + itoa(len(region.Neighbors)) + "):"
	} else {
		neighborTitle = "Komşu (" + itoa(len(region.Neighbors)) + ") [gösterilen: 4]"
	}
	DrawText(screen, neighborTitle, lx, ly, FaceSmall, color.RGBA{200, 170, 90, 220})
	ly += 18

	// İlk 4 komşuyu listele
	displayCount := len(region.Neighbors)
	if displayCount > 4 {
		displayCount = 4
	}
	for i := 0; i < displayCount; i++ {
		neighborID := region.Neighbors[i]
		neighborRegion, ok := gs.Regions[neighborID]
		if ok {
			col := color.RGBA{180, 180, 180, 200}
			if neighborRegion.IsSea {
				col = color.RGBA{100, 160, 220, 200}
			}
			DrawText(screen, "• "+neighborRegion.NameTR, lx+15, ly, FaceSmall, col)
			ly += 16
		}
	}

	ly += 10
	vector.StrokeLine(screen, float32(lx), float32(ly), float32(lx)+sepW, float32(ly), 1, panelBorder, false)
	ly += 8

	// Bilgi
	DrawText(screen, "Tıkla: Özel fırsat yok", lx, ly, FaceSmall, color.RGBA{100, 200, 100, 150})
}
