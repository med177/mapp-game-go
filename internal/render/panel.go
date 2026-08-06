package render

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"sort"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/audio"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/victory"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Layout sabitleri ────────────────────────────────────────────────

const (
	bottomBarH        = float32(80)
	topStatusW        = float32(1050)
	topStatusH        = float32(82)
	topDateHudW       = float32(255)
	topDateHudH       = float32(80)
	actionHudPad      = float32(8)
	actionHudGap      = float32(5)
	actionHudGroupGap = float32(18)

	minimapW = float32(240)
	minimapH = float32(165)

	evLogW             = float32(255)
	evLogH             = float32(520)
	evLogMinH          = float32(36)
	eventCardH         = float32(52)
	eventCardGap       = float32(7)
	maxEventLogEntries = 16

	infoPanelW                  = float32(305)
	infoPanelH                  = float32(780)
	factionPanelHeaderH         = 70.0
	factionPanelBodyPadBottom   = 12.0
	factionPanelScrollStep      = 28.0
	factionPanelScrollBarW      = 4.0
	factionPanelScrollBarGap    = 6.0
	factionPanelRowH            = 18.0
	factionPanelSectionH        = 16.0
	factionPanelTechSummaryH    = 48.0
	factionHUDFlagSize          = 58.0
	factionPanelFlagSize        = 44.0
	regionPanelFlagBadgeSize    = 48.0
	regionPanelImperialFlagSize = 30.0
	regionPanelFlagGap          = 6.0

	btnW = float32(90)
	btnH = float32(52)

	panelPad = float64(12)

	regionPanelStatRowGap      = 22.0
	regionPanelBarYOffset      = 4.0
	regionPanelBarH            = float32(8)
	regionPanelTaxButtonW      = float32(18)
	regionPanelTaxButtonH      = float32(16)
	regionPanelTaxButtonGap    = float32(4)
	regionPanelMeterValueW     = float32(54)
	regionPanelMeterGap        = float32(10)
	regionOwnerNameH           = float32(20)
	regionVassalInfoH          = 14.0
	regionPanelScrollStep      = 28.0
	regionPanelActivityMinH    = 32.0
	regionPanelActionBarHeight = 30.0
	regionPanelTabH            = 24.0
	regionPanelTabGap          = 4.0
	regionPanelTabMinContentH  = 120.0
	buildingGridNameH          = float32(18)
	buildingGridRowGap         = float32(7)
)

var factionNameColor = color.RGBA{0, 152, 204, 255}

// regionPanelTab seçili bölge panelindeki ortak içerik alanının görünümünü
// belirler. Sıfır değer bilinçli olarak Binalar görünümüdür.
type regionPanelTab int

const (
	regionPanelTabBuildings regionPanelTab = iota
	regionPanelTabEvents
)

func bottomBarTop() float32     { return float32(ScreenHeight) - bottomBarH }
func minimapX() float32         { return float32(ScreenWidth) - minimapW - 5 }
func minimapY() float32         { return float32(ScreenHeight) - minimapH }
func evLogX() float32           { return float32(ScreenWidth) - evLogW }
func evLogY() float32           { return topDateHudH + 8 }
func infoPanelX() float32       { return 0 }
func infoPanelY() float32       { return float32(ScreenHeight) - infoPanelH }
func settlementPanelX() float32 { return infoPanelX() + infoPanelW + 8 }
func settlementPanelY() float32 { return infoPanelY() }
func factionPanelX() float32    { return settlementPanelX() }
func factionPanelY() float32    { return settlementPanelY() }

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

	factionPanelBodyBuffer = func() *ebiten.Image {
		w, h := factionPanelBodyCanvasSize()
		return ebiten.NewImage(w, h)
	}()

	// buildingSpriteCache aktif senaryodaki bina ID'lerini ayrı PNG asset'lerine
	// eşler. Eksik asset'ler de nil olarak cache'lenir; böylece her frame diske
	// yeniden erişilmez.
	buildingSpriteCache = map[string]*ebiten.Image{}

	settlementImageCache  = map[string]*ebiten.Image{}
	settlementImageLoaded = map[string]bool{}

	// factionFlagCache senaryo bayraklarını faction ID'siyle eşleştirir.
	// nil değerler de cache'lenir; böylece eksik asset her frame diskten aranmaz.
	factionFlagCache = map[string]*ebiten.Image{}
)

// buildingDisplayOrder bina slotlarının sırasını belirler.
var buildingDisplayOrder = []string{"market", "farm", "barracks", "walls", "temple", "port", "granary"}

func buildingSpritePath(id string) string {
	if ActiveScenarioPath == "" || id == "" {
		return ""
	}
	return filepath.Join(ActiveScenarioPath, "sprites", "buildings", id+".png")
}

// buildingSpriteImage bir bina ID'sinin aynı ID'li PNG asset'ini döner.
func buildingSpriteImage(id string) *ebiten.Image {
	path := buildingSpritePath(id)
	if path == "" {
		return nil
	}
	if img, loaded := buildingSpriteCache[path]; loaded {
		return img
	}

	img := tryLoadImage(path)
	buildingSpriteCache[path] = img
	return img
}

func resetBuildingSpriteCache() {
	buildingSpriteCache = map[string]*ebiten.Image{}
}

func ensureMiniMapBg() {
	if miniMapLoaded {
		return
	}
	miniMapLoaded = true
	miniMapBg = tryLoadImage(ActiveScenarioPath + "/maps/mini-map.png")
}

// factionFlagImage faction ID'sine karşılık gelen senaryo bayrağını döner.
// Bayrak bulunamazsa nil döner ve HUD mevcut baş harf fallback'ini kullanır.
func factionFlagImage(fid faction.FactionID) *ebiten.Image {
	if fid == "" || ActiveScenarioPath == "" {
		return nil
	}

	path := filepath.Join(ActiveScenarioPath, "sprites", "flags", string(fid)+".png")
	if img, loaded := factionFlagCache[path]; loaded {
		return img
	}

	img := tryLoadImage(path)
	factionFlagCache[path] = img
	return img
}

func resetFactionFlagCache() {
	factionFlagCache = map[string]*ebiten.Image{}
}

// drawFactionFlagBadge kare faction rozetini çizer; asset yoksa baş harfi gösterir.
func drawFactionFlagBadge(screen *ebiten.Image, fid faction.FactionID, initial string, x, y, size float64, bg, border color.Color) {
	vector.FillRect(screen, float32(x), float32(y), float32(size), float32(size), bg, false)
	if flag := factionFlagImage(fid); flag != nil {
		bounds := flag.Bounds()
		flagW := float64(bounds.Dx())
		flagH := float64(bounds.Dy())
		scale := size / flagW
		if flagH > flagW {
			scale = size / flagH
		}
		drawW := flagW * scale
		drawH := flagH * scale
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(x+(size-drawW)/2, y+(size-drawH)/2)
		screen.DrawImage(flag, op)
	} else if initial != "" {
		DrawTextCentered(screen, initial, x+size/2, y+size/2-8, FaceLarge, color.RGBA{255, 255, 255, 240})
	}
	if border != nil {
		vector.StrokeRect(screen, float32(x), float32(y), float32(size), float32(size), 2, border, false)
	}
}

func factionInitial(name string) string {
	for _, r := range name {
		return string(r)
	}
	return "?"
}

func regionPanelOwnerFlagRect(lx, py float64) gameui.Rect {
	return gameui.Rect{
		X: lx,
		Y: py - regionPanelFlagBadgeSize,
		W: regionPanelFlagBadgeSize,
		H: regionPanelFlagBadgeSize,
	}
}

func regionPanelImperialFlagRect(lx, py float64) gameui.Rect {
	ownerRect := regionPanelOwnerFlagRect(lx, py)
	return gameui.Rect{
		X: ownerRect.X + ownerRect.W + regionPanelFlagGap,
		Y: ownerRect.Y + (ownerRect.H-regionPanelImperialFlagSize)/2,
		W: regionPanelImperialFlagSize,
		H: regionPanelImperialFlagSize,
	}
}

// regionImperialEmpireID bölge sahibinin bağımsız bir imparatorluk üyesi
// olup olmadığını belirler. İmparatorluğun doğrudan sahip olduğu bölgelerde
// ikinci bayrak gösterilmez; zaten ilk bayrak kurumun kendisidir.
func regionImperialEmpireID(gs *state.GameState, ownerID string) faction.FactionID {
	if gs == nil || gs.Imperial == nil || ownerID == "" || ownerID == string(gs.Imperial.EmpireID) {
		return ""
	}
	member, ok := gs.Imperial.Members[faction.FactionID(ownerID)]
	if !ok || member == nil {
		return ""
	}
	return gs.Imperial.EmpireID
}

func bottomActionHudRect() (x, y, w, h float32) {
	// Tur Bitir, aksiyon grubundan görsel olarak ayrıldığı için aradaki
	// normal buton boşluğuna ek bir grup aralığı kullanır.
	w = btnW*5 + actionHudGap*3 + actionHudGroupGap + actionHudPad*2
	h = btnH + actionHudPad*2
	x = float32(ScreenWidth)/2 - w/2
	y = float32(ScreenHeight) - h
	if x < 0 {
		x = 0
	}
	return x, y, w, h
}

func mapModeHudRect() (x, y, w, h float32) {
	w = 230
	h = 30
	// Harita modu düğmeleri minimap'in hemen üstünde, onunla aynı yatay
	// eksende durur.
	x = minimapX() + minimapW/2 - w/2
	y = minimapY() - h - 6
	if x < 0 {
		x = 0
	}
	if x+w > float32(ScreenWidth) {
		x = float32(ScreenWidth) - w
	}
	if y < 0 {
		y = 0
	}
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

func buttonFromRectF32(r [4]float32, label string) gameui.Button {
	return gameui.NewButton(float64(r[0]), float64(r[1]), float64(r[2]), float64(r[3]), label)
}

func buildBottomActionButtons(recruitEnabled bool) [5]gameui.Button {
	rects := BottomButtonRects()
	labels := [5]string{"Ordu", "Pazar", "Diplomasi", "Teknoloji", "Tur Bitir ►"}
	var buttons [5]gameui.Button
	for i, rect := range rects {
		btn := buttonFromRectF32(rect, labels[i])
		if i == 0 {
			btn.Enabled = recruitEnabled
		}
		buttons[i] = btn
	}
	return buttons
}

func buildMapModeButtons() [2]gameui.Button {
	rects := mapModeButtonRects()
	return [2]gameui.Button{
		buttonFromRectF32(rects[0], "Normal"),
		buttonFromRectF32(rects[1], "Ticaret"),
	}
}

func imperialPanelAvailable(gs *state.GameState) bool {
	return gs != nil && gs.Imperial != nil && gs.Imperial.EmpireID != "" && gs.PlayerFactionID == gs.Imperial.EmpireID
}

func imperialHUDButtonRect() [4]float32 {
	ax, ay, aw, _ := bottomActionHudRect()
	const buttonW = float32(116)
	x := ax + aw + actionHudGap
	y := ay + actionHudPad
	if x+buttonW > float32(ScreenWidth)-8 {
		x = float32(ScreenWidth) - buttonW - 8
	}
	if x < 8 {
		x = 8
	}
	return [4]float32{x, y, buttonW, btnH}
}

func buildImperialHUDButton() gameui.Button {
	return buttonFromRectF32(imperialHUDButtonRect(), "İmparatorluk")
}

func imperialHUDStatusText(gs *state.GameState) string {
	if gs == nil || gs.Imperial == nil {
		return ""
	}
	if gs.Imperial.PendingDecision != nil {
		return "Karar bekliyor"
	}
	return "Otorite " + itoa(gs.Imperial.Authority) + " • Diyet " + itoa(gs.Imperial.NextDietTurn)
}

func imperialHUDButtonHit(fx, fy float64) bool {
	return buildImperialHUDButton().HitTest(fx, fy)
}

// BottomButtonRects alt-orta aksiyon HUD'undaki buton dikdörtgenlerini döner.
// [0]=Ordu [1]=Pazar [2]=Diplomasi [3]=Teknoloji [4]=Tur Bitir
func BottomButtonRects() [5][4]float32 {
	hudX, hudY, _, _ := bottomActionHudRect()
	by := hudY + actionHudPad
	armyX := hudX + actionHudPad
	marketX := armyX + btnW + actionHudGap
	diplX := marketX + btnW + actionHudGap
	techX := diplX + btnW + actionHudGap
	endX := techX + btnW + actionHudGroupGap
	return [5][4]float32{
		{armyX, by, btnW, btnH},
		{marketX, by, btnW, btnH},
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
	if fx >= float64(mx) && fx <= float64(mx+mw) && fy >= float64(my) && fy <= float64(my+mh) {
		return true
	}
	return false
}

func bottomActionButtonHit(fx, fy float64) bool {
	for _, btn := range buildBottomActionButtons(true) {
		if btn.HitTest(fx, fy) {
			return true
		}
	}
	for _, btn := range buildMapModeButtons() {
		if btn.HitTest(fx, fy) {
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
	return buildTopDateHudMenuButton().HitTest(fx, fy)
}

func buildTopDateHudMenuButton() gameui.Button {
	x, y, w, h := topDateHudMenuButtonRect()
	return gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "Menü").WithIcon(gameui.IconMenu)
}

func musicHudRect() (x, y, w, h float32) {
	w = 470
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
	return [4]float32{x + 310, y + 7, 58, 22}
}

func musicHudNextRect() [4]float32 {
	x, y, _, _ := musicHudRect()
	return [4]float32{x + 374, y + 7, 84, 22}
}

func musicHudInteractiveHit(fx, fy float64) bool {
	status := audio.MusicStatusNow()
	if !status.HasPlaylist {
		return false
	}
	toggle, next := buildMusicHudButtons(status.Playing)
	return toggle.HitTest(fx, fy) || next.HitTest(fx, fy)
}

func buildMusicHudButtons(playing bool) (gameui.Button, gameui.Button) {
	toggle := "Dur"
	toggleIcon := gameui.IconPause
	if !playing {
		toggle = "Çal"
		toggleIcon = gameui.IconPlay
	}
	return buttonFromRectF32(musicHudToggleRect(), toggle).WithIcon(toggleIcon),
		buttonFromRectF32(musicHudNextRect(), "Sonraki").WithIcon(gameui.IconNext)
}

func musicHudHit(fx, fy float64) bool {
	status := audio.MusicStatusNow()
	if !status.HasPlaylist {
		return false
	}
	x, y, w, h := musicHudRect()
	return fx >= float64(x) && fx <= float64(x+w) && fy >= float64(y) && fy <= float64(y+h)
}

func turnTechHudRect() (x, y, w, h float32) {
	mx, my, mw, mh := musicHudRect()
	x = mx
	y = my + mh
	w = mw
	h = 46
	return x, y, w, h
}

func turnTechHudTechRect() (x, y, w, h float32) {
	x, y, w, h = turnTechHudRect()
	y += 18
	h -= 18
	return x, y, w, h
}

func turnTechHudTechHit(fx, fy float64) bool {
	x, y, w, h := turnTechHudTechRect()
	return fx >= float64(x) && fx <= float64(x+w) && fy >= float64(y) && fy <= float64(y+h)
}

// ── Ana alt bar ──────────────────────────────────────────────────────

// DrawBottomPanel üst sol durum panelini, sağ üst tarih HUD'unu ve alt-orta aksiyon HUD'unu çizer.
func DrawBottomPanel(screen *ebiten.Image, gs *state.GameState, showRecruit, recruitEnabled bool, recruitReason string, showTrade, showDiplomacy, showTech, showActiveWars bool, mapMode MapMode) {
	by := float32(0)
	bw := topStatusW
	if bw > float32(ScreenWidth) {
		bw = float32(ScreenWidth)
	}
	statusRect := gameui.Rect{X: 0, Y: float64(by), W: float64(bw), H: float64(topStatusH)}

	drawUIPanelFrame(screen, statusRect, panelBg, panelBorder, 1.5, 3)
	drawUISeparator(screen, 0, by+topStatusH, bw, 1.5, panelBorder)
	vector.StrokeLine(screen, bw, by+4, bw, by+topStatusH, 1, color.RGBA{80, 65, 35, 120}, false)

	f, hasPlayer := gs.Factions[gs.PlayerFactionID]

	// Sol blok: fraksiyon amblemi + isim + askeri güç sırası
	if hasPlayer {
		fc := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		initial := string([]rune(f.NameTR)[:1])
		flagX := float64(5)
		flagY := float64(by) + (float64(topStatusH)-factionHUDFlagSize)/2
		drawFactionFlagBadge(screen, f.ID, initial, flagX, flagY, factionHUDFlagSize, fc, panelBorder)

		textX := flagX + factionHUDFlagSize + 13
		DrawText(screen, f.NameTR, textX, float64(by)+10, FaceLarge, fc)
		militaryPower, militaryRank, factionCount := playerMilitaryPowerStanding(gs)
		DrawText(screen, "Askeri güç: "+formatNumberTR(militaryPower), textX, float64(by)+34, FaceSmall, ColorGray)
		DrawText(screen, "Güç sırası: "+formatNumberTR(militaryRank)+"/"+formatNumberTR(factionCount), textX, float64(by)+53, FaceSmall, ColorGray)
	}

	// Kaynaklar: solda 2x2 mal ızgarası, sağda Gelir/Altın
	if hasPlayer {
		leftCol1, leftCol2, rightCol, leftColW, rightColW := topResourceHUDColumns()
		ry := float64(by) + 12
		rowGap := 22.0

		// 2x2 mallar
		production := gs.FactionProductionSummary(f.ID)
		grainChange := gs.FactionGrainNetChange(f.ID)
		grainValue := formatResourceHUDValue(f.Grain, grainChange)
		grainCapacity := gs.GrainStorageCapacityForFaction(f.ID)
		grainColor := ColorWhite
		if grainStatus, ok := gs.GrainEconomy[f.ID]; ok {
			if grainStatus.StorageCapacity > 0 {
				grainCapacity = grainStatus.StorageCapacity
			}
			switch grainStatus.SupplyLevel {
			case state.GrainSupplyWarning:
				grainColor = ColorYellow
			case state.GrainSupplyCritical, state.GrainSupplyFamine:
				grainColor = ColorRed
			}
		}
		if grainChange < 0 {
			grainColor = ColorRed
		}
		drawResRow(screen, leftCol1, ry, leftColW, economy.ResourceNameTR(economy.ResourceGrain), grainValue, grainColor)
		drawResRow(screen, leftCol1, ry+rowGap, leftColW, economy.ResourceNameTR(economy.ResourceSpice), formatResourceHUDValue(f.Spice, production.Spice), resourceHUDChangeColor(production.Spice, color.RGBA{170, 170, 170, 255}))
		drawResRow(screen, leftCol1, ry+rowGap*2, leftColW, economy.ResourceNameTR(economy.ResourceCloth), formatResourceHUDValue(f.Cloth, production.Cloth), resourceHUDChangeColor(production.Cloth, color.RGBA{170, 170, 170, 255}))

		drawResRow(screen, leftCol2, ry, leftColW, economy.ResourceNameTR(economy.ResourceTimber), formatResourceHUDValue(f.Timber, production.Timber), resourceHUDChangeColor(production.Timber, color.RGBA{180, 140, 80, 255}))
		drawResRow(screen, leftCol2, ry+rowGap, leftColW, economy.ResourceNameTR(economy.ResourceStone), formatResourceHUDValue(f.Stone, production.Stone), resourceHUDChangeColor(production.Stone, color.RGBA{170, 170, 170, 255}))
		drawResRow(screen, leftCol2, ry+rowGap*2, leftColW, economy.ResourceNameTR(economy.ResourceIron), formatResourceHUDValue(f.Iron, production.Iron), resourceHUDChangeColor(production.Iron, color.RGBA{180, 180, 220, 255}))

		income := calcPlayerIncome(gs)
		incCol := ColorGold
		if income < 0 {
			incCol = ColorRed
		}
		sign := "+"
		if income < 0 {
			sign = ""
		}
		if grainCapacity > 0 {
			drawResRow(screen, rightCol, ry, rightColW, "Ambar", formatNumberTR(grainCapacity), ColorGray)
		}
		drawResRow(screen, rightCol, ry+rowGap, rightColW, "Gelir", sign+formatNumberTR(income), incCol)
		drawResRow(screen, rightCol, ry+rowGap*2, rightColW, "Hazine", formatNumberTR(f.Gold), ColorGold)
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
	drawUIPanelFrame(screen, gameui.Rect{X: float64(hudX), Y: float64(hudY), W: float64(hudW), H: float64(hudH)}, panelBg, panelBorder, 1.5, 3)

	rects := BottomButtonRects()
	active := [5]bool{showRecruit, showTrade, showDiplomacy, showTech, false}
	enabled := [5]bool{recruitEnabled, true, true, true, true}
	buttons := buildBottomActionButtons(recruitEnabled)
	bgNorm := [5]color.RGBA{
		{88, 62, 30, 220},
		{64, 82, 46, 235},
		{40, 65, 110, 215},
		{60, 40, 95, 215},
		{105, 28, 28, 230},
	}
	bgAct := [5]color.RGBA{
		{150, 106, 48, 245},
		{92, 128, 62, 245},
		{80, 130, 200, 240},
		{110, 70, 170, 240},
		{165, 48, 48, 255},
	}
	for i, r := range rects {
		style := solidButtonStyle(bgNorm[i], panelBorder, ColorWhite, 15)
		if active[i] {
			style.BG = bgAct[i]
		}
		if !enabled[i] {
			style.DisabledBG = color.RGBA{34, 30, 24, 180}
			style.DisabledBorder = panelBorder
			style.DisabledText = color.RGBA{120, 112, 96, 210}
		}
		style.BorderWidth = 1.5
		style.TextVariant = gameui.TextMedium
		drawUIButtonWidget(screen, buttons[i], style)
		if i == 0 && !enabled[i] && recruitReason != "" {
			reason := trimTextToWidth(recruitReason, FaceTiny, float64(r[2])-10)
			reasonW := MeasureText(reason, FaceTiny)
			DrawText(screen, reason, float64(r[0])+float64(r[2])/2-reasonW/2, float64(r[1])+31, FaceTiny, color.RGBA{120, 112, 96, 210})
		}
	}
	drawMapModeHud(screen, mapMode)
	if imperialPanelAvailable(gs) {
		btn := buildImperialHUDButton()
		style := solidButtonStyle(color.RGBA{48, 38, 23, 235}, color.RGBA{184, 148, 70, 255}, ColorWhite, 6)
		drawUIButtonWidget(screen, btn, style)
		status := imperialHUDStatusText(gs)
		status = trimTextToWidth(status, FaceSmall, float64(btn.W)-10)
		drawUILabel(screen, gameui.Rect{X: btn.X + 5, Y: btn.Y + 32, W: btn.W - 10}, status, color.RGBA{232, 210, 162, 235}, gameui.TextSmall, gameui.TextAlignCenter)
	}

	drawDateMenuHud(screen, gs, mapMode)
	drawMusicHud(screen)
	drawActiveWarsHUDButton(screen, gs, showActiveWars)
	drawTurnTechHud(screen, gs)
}

// factionMilitaryPowerStanding seçili devletin askeri gücünü ve aktif devletler
// arasındaki deterministik sırasını döner. Eşit güçte faction ID'si küçük olan
// devlet üstte kabul edilir; böylece map iterasyon sırası sonucu değiştirmez.
func factionMilitaryPowerStanding(gs *state.GameState, fid faction.FactionID) (power, rank, factionCount int) {
	if gs == nil || fid == "" {
		return 0, 0, 0
	}

	selected, ok := gs.Factions[fid]
	if !ok || selected == nil || selected.IsEliminated {
		return 0, 0, 0
	}

	power = diplomacy.MilitaryPower(gs, fid)
	rank = 1
	for candidateID, candidate := range gs.Factions {
		if candidate == nil || candidate.IsEliminated {
			continue
		}
		factionCount++
		if candidateID == fid {
			continue
		}

		candidatePower := diplomacy.MilitaryPower(gs, candidateID)
		if candidatePower > power || (candidatePower == power && candidateID < fid) {
			rank++
		}
	}
	return power, rank, factionCount
}

// playerMilitaryPowerStanding oyuncunun askeri standing'ini döner.
func playerMilitaryPowerStanding(gs *state.GameState) (power, rank, factionCount int) {
	if gs == nil {
		return 0, 0, 0
	}
	return factionMilitaryPowerStanding(gs, gs.PlayerFactionID)
}

func drawMapModeHud(screen *ebiten.Image, mapMode MapMode) {
	x, y, w, h := mapModeHudRect()
	drawUICardRect(screen, gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}, color.RGBA{14, 14, 18, 220}, panelBorder, 1.2)
	buttons := buildMapModeButtons()
	for i, btn := range buttons {
		active := (i == 0 && mapMode == MapModeNormal) || (i == 1 && mapMode == MapModeTrade)
		drawUIButton(screen, btn.X, btn.Y, btn.W, btn.H, btn.Label, true, mapModeButtonStyle(active))
	}
}

func drawMusicHud(screen *ebiten.Image) {
	status := audio.MusicStatusNow()
	if !status.HasPlaylist {
		return
	}
	x, y, w, h := musicHudRect()
	drawUICardRect(screen, gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}, color.RGBA{14, 12, 9, 220}, panelBorder, 1)

	track := status.Track
	if track == "" {
		track = "Çalma Listesi Hazır"
	}
	track = strings.TrimSuffix(track, ".ogg")
	track = strings.TrimSuffix(track, ".mp3")
	track = strings.TrimSuffix(track, ".wav")
	label := trimTextToWidth("Müzik: "+track, FaceSmall, 250)
	DrawText(screen, label, float64(x)+10, float64(y)+11, FaceSmall, ColorGray)

	toggleBtn, nextBtn := buildMusicHudButtons(status.Playing)
	drawTinyPanelButtonWidget(screen, toggleBtn, true)
	drawTinyPanelButtonWidget(screen, nextBtn, true)
}

func drawTurnTechHud(screen *ebiten.Image, gs *state.GameState) {
	if gs == nil {
		return
	}
	f, ok := gs.Factions[gs.PlayerFactionID]
	if !ok || f == nil {
		return
	}

	x, y, w, h := turnTechHudRect()
	drawUICardRect(screen, gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}, color.RGBA{14, 12, 9, 220}, panelBorder, 1)

	phaseStr := trimTextToWidth(phaseLabel(gs.Phase), FaceSmall, float64(w)-20)
	DrawText(screen, phaseStr, float64(x)+10, float64(y)+8, FaceSmall, ColorGray)
	if quotaText, quotaColor := diplomacyOfferQuotaHUDText(gs); quotaText != "" {
		quotaW := MeasureText(quotaText, FaceSmall)
		DrawText(screen, quotaText, float64(x+w)-10-quotaW, float64(y)+8, FaceSmall, quotaColor)
	}

	techStr := "Teknoloji yok"
	techCol := ColorGray
	if f.Research.ActiveID != "" {
		if tech, ok := gs.TechTypes[f.Research.ActiveID]; ok {
			techStr = tech.NameTR + " (" + itoa(f.Research.TurnsLeft) + " tur)"
			techCol = color.RGBA{100, 220, 100, 255}
		}
	}
	techStr = trimTextToWidth(techStr, FaceSmall, float64(w)-20)
	DrawText(screen, techStr, float64(x)+10, float64(y)+26, FaceSmall, techCol)
}

func diplomacyOfferQuotaHUDText(gs *state.GameState) (string, color.RGBA) {
	if gs == nil || gs.PlayerFactionID == "" {
		return "", color.RGBA{}
	}
	remaining := gs.DiplomacyOfferQuotaRemaining(gs.PlayerFactionID)
	label := "Elçi " + itoa(remaining) + "/" + itoa(state.MaxDiplomacyOffersPerTurn)
	switch remaining {
	case 3:
		return label, color.RGBA{232, 190, 100, 255}
	case 2:
		return label, color.RGBA{210, 168, 70, 255}
	case 1:
		return label, color.RGBA{220, 130, 60, 255}
	default:
		return label, color.RGBA{220, 90, 90, 255}
	}
}

func drawDateMenuHud(screen *ebiten.Image, gs *state.GameState, mapMode MapMode) {
	x, y, w, h := topDateHudRect()
	drawUIPanelFrame(screen, gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}, panelBg, panelBorder, 1.5, 3)

	dateStr := strategicTurnDateTR(gs)
	DrawText(screen, dateStr, float64(x)+12, float64(y)+13, FaceMed, ColorGold)
	DrawText(screen, gs.CurrentSeason().DisplayName()+"  •  Tur "+itoa(gs.Turn),
		float64(x)+12, float64(y)+42, FaceSmall, color.RGBA{160, 200, 100, 220})
	DrawText(screen, "Zorluk: "+difficultyLabelTR(gs.Difficulty),
		float64(x)+12, float64(y)+61, FaceSmall, ColorGray)

	_ = mapMode

	btn := buildTopDateHudMenuButton()
	drawUIButtonWidget(screen, btn, dateMenuButtonStyle)
}

func strategicTurnDateTR(gs *state.GameState) string {
	if gs == nil {
		return ""
	}
	startMonth := monthName(gs.Month)
	endYear, endMonth := gs.CurrentTurnEndDate()
	endMonthName := monthName(endMonth)
	if startMonth == "" || endMonthName == "" {
		return itoa(gs.Year)
	}
	if gs.Year == endYear && gs.Month == endMonth {
		return startMonth + " " + itoa(gs.Year)
	}
	if gs.Year == endYear {
		return startMonth + "–" + endMonthName + " " + itoa(gs.Year)
	}
	return startMonth + " " + itoa(gs.Year) + "–" + endMonthName + " " + itoa(endYear)
}

// ── Olay Logu (sağ üst) ──────────────────────────────────────────────

// DrawEventLog sağ üst köşede son olayları kartlar halinde listeler.
func DrawEventLog(screen *ebiten.Image, events []string, collapsed bool, scroll int, hasCodex bool) {
	ex := evLogX()
	ey := evLogY()
	eh := eventLogPanelH(collapsed)

	drawUIPanelFrame(screen, gameui.Rect{X: float64(ex), Y: float64(ey), W: float64(evLogW), H: float64(eh)}, panelBg, panelBorder, 1.5, 3)

	titleW := MeasureText("Olay Mesajları", FaceMed)
	DrawText(screen, "Olay Mesajları", float64(ex)+12, float64(ey)+8, FaceMed,
		color.RGBA{220, 190, 100, 255})
	if len(events) > 0 {
		count := "(" + itoa(len(events)) + ")"
		DrawText(screen, count, float64(ex)+18+titleW, float64(ey)+9, FaceSmall, ColorGray)
	}

	toggleBtn := buildEventLogToggleButton(collapsed)
	toggleBtn.Label = ""
	if collapsed {
		toggleBtn.Icon = gameui.IconPlus
	} else {
		toggleBtn.Icon = gameui.IconMinus
	}
	toggleBtn.IconSize = 14
	drawUIButtonWidget(screen, toggleBtn, eventLogButtonStyle(ColorGold))
	if hasCodex {
		codexBtn := buildEventLogCodexButton()
		drawUIButtonWidget(screen, codexBtn, eventLogButtonStyle(ColorGold))
	}

	if collapsed {
		return
	}

	if len(events) == 0 {
		drawUILabel(screen, gameui.Rect{X: float64(ex), Y: float64(ey) + 58, W: float64(evLogW)}, "Henüz olay yok", color.RGBA{150, 140, 120, 190}, gameui.TextSmall, gameui.TextAlignCenter)
		drawUILabel(screen, gameui.Rect{X: float64(ex), Y: float64(ey) + 76, W: float64(evLogW)}, "Oyun olayları burada listelenir", color.RGBA{110, 105, 95, 170}, gameui.TextSmall, gameui.TextAlignCenter)
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

		closeBtn := buildEventLogCloseButton(visibleIndex)
		drawUIButtonWidget(screen, closeBtn, eventLogButtonStyle(ColorGray))

		drawUIWrappedLabel(screen, gameui.Rect{X: float64(cardX) + 10, Y: float64(cardY) + 8, W: float64(cardW - 34)}, ev, color.RGBA{220, 210, 185, 235}, gameui.TextSmall, 15, 2)
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

func eventLogCodexRect() (x, y, w, h float32) {
	w, h = 54, 22
	tx, y, tw, _ := eventLogToggleRect()
	x = tx - w - 6
	if tw == 0 {
		x = evLogX() + evLogW - w - 38
	}
	return x, y, w, h
}

func eventLogToggleHit(mx, my float64, collapsed bool) bool {
	return buildEventLogToggleButton(collapsed).HitTest(mx, my)
}

func buildEventLogToggleButton(_ bool) gameui.Button {
	x, y, w, h := eventLogToggleRect()
	btn := gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "")
	btn.IconSize = 14
	return btn
}

func eventLogCodexHit(mx, my float64) bool {
	return buildEventLogCodexButton().HitTest(mx, my)
}

func buildEventLogCodexButton() gameui.Button {
	x, y, w, h := eventLogCodexRect()
	btn := gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "Kodex").WithIcon(gameui.IconBook)
	btn.IconSize = 13
	return btn
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
		if buildEventLogCloseButton(i).HitTest(mx, my) {
			return eventIndex
		}
	}
	return -1
}

func buildEventLogCloseButton(index int) gameui.Button {
	x, y, w, h := eventLogCloseRect(index)
	btn := gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "").WithIcon(gameui.IconClose)
	btn.IconSize = 12
	return btn
}

func eventLogInteractiveHit(mx, my float64, eventCount int, collapsed bool, scroll int, hasCodex bool) bool {
	if eventLogToggleHit(mx, my, collapsed) {
		return true
	}
	if hasCodex && eventLogCodexHit(mx, my) {
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
	layout := buildEventDetailLayout()
	return float32(layout.panelRect.X), float32(layout.panelRect.Y), float32(layout.panelRect.W), float32(layout.panelRect.H)
}

func eventDetailPopupHit(mx, my float64) bool {
	return buildEventDetailModal().Panel.HitTest(mx, my)
}

func eventDetailCloseRect() (x, y, w, h float32) {
	btn := buildEventDetailCloseButton()
	return float32(btn.X), float32(btn.Y), float32(btn.W), float32(btn.H)
}

func eventDetailCloseHit(mx, my float64) bool {
	return buildEventDetailCloseButton().HitTest(mx, my)
}

func eventCodexPopupRect() (x, y, w, h float32) {
	layout := buildEventCodexLayout()
	return float32(layout.panelRect.X), float32(layout.panelRect.Y), float32(layout.panelRect.W), float32(layout.panelRect.H)
}

func eventCodexPopupHit(mx, my float64) bool {
	return buildEventCodexModal().Panel.HitTest(mx, my)
}

func eventCodexCloseRect() (x, y, w, h float32) {
	btn := buildEventCodexCloseButton()
	return float32(btn.X), float32(btn.Y), float32(btn.W), float32(btn.H)
}

func eventCodexCloseHit(mx, my float64) bool {
	return buildEventCodexCloseButton().HitTest(mx, my)
}

func victoryDetailPopupHit(mx, my float64) bool {
	return buildVictoryDetailModal().Panel.HitTest(mx, my)
}

func victoryDetailCloseHit(mx, my float64) bool {
	return buildVictoryDetailCloseButton().HitTest(mx, my)
}

type victoryDetailContentLine struct {
	text     string
	color    color.Color
	variant  gameui.TextVariant
	lineStep float64
}

func victoryDetailScrollHit(mx, my float64) bool {
	return buildVictoryDetailLayout().scrollRect.Hit(mx, my)
}

func victoryDetailContentHeight(gs *state.GameState) float64 {
	return buildVictoryDetailContentLines(gs).contentHeight
}

func clampVictoryDetailScroll(gs *state.GameState, scroll float64) float64 {
	max := victoryDetailMaxScroll(gs)
	if scroll < 0 {
		return 0
	}
	if scroll > max {
		return max
	}
	return scroll
}

func victoryDetailMaxScroll(gs *state.GameState) float64 {
	layout := buildVictoryDetailLayout()
	max := victoryDetailContentHeight(gs) - layout.scrollRect.H
	if max < 0 {
		return 0
	}
	return max
}

type victoryDetailContent struct {
	lines         []victoryDetailContentLine
	contentHeight float64
}

func buildVictoryDetailContentLines(gs *state.GameState) victoryDetailContent {
	bodyWidth := buildVictoryDetailLayout().bodyRect.W
	opt, hasOpt := currentVictoryOption(gs)
	desc := ""
	targetSummary := activeVictoryTargetSummary(gs)
	longSummary := targetSummary
	if hasOpt {
		desc = opt.Description
		if sum := victoryTargetSummary(gs, opt); sum != "" {
			longSummary = sum
		}
	}

	lines := make([]victoryDetailContentLine, 0, 48)
	contentH := 0.0
	appendSectionLabel := func(label string) {
		lines = append(lines, victoryDetailContentLine{
			text:     label,
			color:    ColorGold,
			variant:  gameui.TextSmall,
			lineStep: 18,
		})
		contentH += 18
	}
	appendWrapped := func(text string, face *text.GoTextFace, col color.Color, variant gameui.TextVariant, lineStep float64) {
		if text == "" {
			return
		}
		for _, line := range wrapTextLines(text, face, bodyWidth) {
			lines = append(lines, victoryDetailContentLine{
				text:     line,
				color:    col,
				variant:  variant,
				lineStep: lineStep,
			})
			contentH += lineStep
		}
	}
	appendInfoLines := func(items []string, colors []color.Color) {
		for i, line := range items {
			col := color.Color(ColorGray)
			if i < len(colors) && colors[i] != nil {
				col = colors[i]
			}
			lines = append(lines, victoryDetailContentLine{
				text:     line,
				color:    col,
				variant:  gameui.TextSmall,
				lineStep: 24,
			})
			contentH += 24
		}
	}
	addGap := func(h float64) {
		contentH += h
	}

	appendSectionLabel("Zafer Detayı")
	addGap(18)
	appendWrapped(desc, FaceMed, color.RGBA{226, 220, 204, 240}, gameui.TextMedium, 20)

	if targetSummary != "" {
		if desc != "" {
			addGap(14)
		}
		appendSectionLabel("Aktif Hedef")
		addGap(18)
		appendWrapped(targetSummary, FaceSmall, color.RGBA{210, 200, 176, 235}, gameui.TextSmall, 17)
	}

	if longSummary != "" && longSummary != targetSummary {
		addGap(14)
		appendSectionLabel("Kapsam")
		addGap(18)
		appendWrapped(longSummary, FaceSmall, color.RGBA{180, 170, 146, 230}, gameui.TextSmall, 17)
	}

	addGap(14)
	appendSectionLabel("İlerleme")
	addGap(18)
	appendInfoLines(victoryDetailProgressLines(gs), nil)

	checkLines, checkColors := victoryChecklistEntries(gs)
	if len(checkLines) > 0 {
		addGap(14)
		appendSectionLabel("Kontrol Listesi")
		addGap(18)
		appendInfoLines(checkLines, checkColors)
	}

	if hasOpt && opt.Detail != "" {
		addGap(14)
		appendSectionLabel("Not")
		addGap(18)
		appendWrapped(opt.Detail, FaceSmall, color.RGBA{168, 154, 126, 220}, gameui.TextSmall, 17)
	}

	return victoryDetailContent{
		lines:         lines,
		contentHeight: contentH,
	}
}

func drawVictoryDetailScrollbar(screen *ebiten.Image, gs *state.GameState, scroll float64) {
	layout := buildVictoryDetailLayout()
	maxScroll := victoryDetailMaxScroll(gs)
	if maxScroll <= 0 {
		return
	}
	track := layout.scrollbar
	vector.FillRect(screen, float32(track.X), float32(track.Y), float32(track.W), float32(track.H), color.RGBA{34, 28, 20, 210}, false)
	thumbH := track.H * (layout.scrollRect.H / victoryDetailContentHeight(gs))
	if thumbH < 44 {
		thumbH = 44
	}
	if thumbH > track.H {
		thumbH = track.H
	}
	thumbY := track.Y
	if maxScroll > 0 && track.H > thumbH {
		thumbY += (track.H - thumbH) * (scroll / maxScroll)
	}
	vector.FillRect(screen, float32(track.X), float32(thumbY), float32(track.W), float32(thumbH), color.RGBA{176, 138, 62, 220}, false)
}

func victoryProgressPanelRect() gameui.Rect {
	return gameui.Rect{X: 718, Y: 7, W: 180, H: float64(topStatusH - 14)}
}

func victoryProgressHit(mx, my float64) bool {
	return victoryProgressPanelRect().Hit(mx, my)
}

func minimapHit(mx, my float64) bool {
	x, y := minimapX(), minimapY()
	return mx >= float64(x) && mx <= float64(x+minimapW) && my >= float64(y) && my <= float64(y+minimapH)
}

func drawEventDetailPopup(screen *ebiten.Image, message string) {
	modal := buildEventDetailModal()
	gameui.DrawModal(screen, modal, eventDetailModalStyle, nil, nil)

	layout := buildEventDetailLayout()
	drawUIPanelTopBar(screen, layout.panelRect, 3, panelBorder)

	lines := eventDetailLines(message, layout.bodyRect.W)
	title := "Olay Detayı"
	bodyLines := lines
	if len(lines) > 0 {
		if trimmed := strings.TrimSpace(lines[0]); trimmed != "" {
			title = trimmed
		}
		if len(lines) > 1 {
			bodyStart := 1
			for bodyStart < len(lines) && strings.TrimSpace(lines[bodyStart]) == "" {
				bodyStart++
			}
			if bodyStart < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[bodyStart]), "Kaynak:") {
				bodyStart++
			}
			for bodyStart < len(lines) && strings.TrimSpace(lines[bodyStart]) == "" {
				bodyStart++
			}
			if bodyStart < len(lines) {
				bodyLines = lines[bodyStart:]
			} else {
				bodyLines = nil
			}
		} else {
			bodyLines = nil
		}
	}

	DrawText(screen, title, layout.titleRect.X, layout.titleRect.Y+6, FaceLarge, ColorGold)
	if source := eventDetailSourceLabel(title, bodyLines); source != "" {
		drawUILabel(screen, layout.filtersRect, source, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	}

	closeBtn := buildEventDetailCloseButton()
	drawUIButtonWidget(screen, closeBtn, tinyButtonStyle)

	if len(bodyLines) == 0 {
		return
	}
	linesForDraw := make([]gameui.RichTextLine, 0, len(bodyLines))
	for _, line := range bodyLines {
		col := color.RGBA{230, 224, 205, 240}
		variant := gameui.TextMedium
		switch {
		case line == "":
			col = color.RGBA{230, 224, 205, 0}
		case strings.HasPrefix(line, "Kaynak:"):
			col = color.RGBA{200, 184, 142, 240}
		case strings.HasPrefix(line, "Seçim:") || strings.HasPrefix(line, "Etki:") || strings.HasPrefix(line, "Zincir:") || strings.HasPrefix(line, "Kosul:") ||
			strings.HasPrefix(line, "Bölge:") || strings.HasPrefix(line, "Tip:") || strings.HasPrefix(line, "Kalan tur:") || strings.HasPrefix(line, "Event ID:"):
			col = color.RGBA{226, 182, 92, 240}
		case strings.HasPrefix(line, "- "):
			col = color.RGBA{240, 236, 224, 240}
		case strings.HasPrefix(line, "  "):
			col = color.RGBA{208, 200, 182, 240}
		}
		linesForDraw = append(linesForDraw, gameui.RichTextLine{
			Text:    line,
			Color:   col,
			Variant: variant,
			Align:   gameui.TextAlignStart,
		})
	}
	drawUIRichTextBlock(screen, gameui.Rect{X: layout.bodyRect.X, Y: layout.bodyRect.Y, W: layout.bodyRect.W}, linesForDraw, 19)
}

func eventDetailSourceLabel(title string, bodyLines []string) string {
	switch {
	case strings.HasPrefix(title, "[KARAR] "):
		return "Kaynak: Karar kaydı"
	case strings.HasPrefix(title, "[OLAY] "):
		return "Kaynak: Olay kaydı"
	case strings.HasPrefix(title, "Karar:"):
		return "Kaynak: Karar izi"
	case strings.HasPrefix(title, "Olay:"):
		return "Kaynak: Olay izi"
	}
	for _, line := range bodyLines {
		switch {
		case strings.HasPrefix(line, "Kaynak: Harita izi"):
			return "Kaynak: Harita izi"
		case strings.HasPrefix(line, "Kaynak: Olay kaydı"):
			return "Kaynak: Olay kaydı"
		case strings.HasPrefix(line, "Kaynak: Karar kaydı"):
			return "Kaynak: Karar kaydı"
		}
	}
	return ""
}

func drawEventCodexPopup(screen *ebiten.Image, filter EventCodexFilter, entries []EventCodexEntry, focus int, scroll int) {
	modal := buildEventCodexModal()
	gameui.DrawModal(screen, modal, eventDetailModalStyle, nil, nil)

	layout := buildEventCodexLayout()
	drawUIPanelTopBar(screen, layout.panelRect, 3, panelBorder)

	DrawText(screen, "Event Kodex", layout.titleRect.X, layout.titleRect.Y+6, FaceLarge, ColorGold)

	closeBtn := buildEventCodexCloseButton()
	drawUIButtonWidget(screen, closeBtn, tinyButtonStyle)

	filterButtons := buildEventCodexFilterButtons()
	for i, btn := range filterButtons {
		active := int(filter) == i
		style := tinyButtonStyle
		if active {
			style = solidButtonStyle(color.RGBA{118, 84, 40, 245}, color.RGBA{226, 182, 92, 255}, ColorWhite, 8)
		}
		drawUIButton(screen, btn.X, btn.Y, btn.W, btn.H, btn.Label, true, style)
	}
	drawUICardRect(screen, layout.listRect, color.RGBA{20, 16, 12, 220}, color.RGBA{90, 72, 38, 210}, 1)
	drawUICardRect(screen, layout.detailRect, color.RGBA{20, 16, 12, 220}, color.RGBA{90, 72, 38, 210}, 1)

	if len(entries) == 0 {
		drawUILabel(screen, gameui.Rect{X: layout.detailRect.X + 18, Y: layout.detailRect.Y + 18}, "Bu filtre için event yok.", ColorGray, gameui.TextMedium, gameui.TextAlignStart)
		return
	}
	if focus < 0 {
		focus = 0
	}
	if focus >= len(entries) {
		focus = len(entries) - 1
	}
	if scroll < 0 {
		scroll = 0
	}
	if maxScroll := eventCodexMaxScroll(len(entries)); scroll > maxScroll {
		scroll = maxScroll
	}
	visibleCount := eventCodexVisibleCount()
	start := scroll
	end := min(len(entries), start+visibleCount)
	for i := start; i < end; i++ {
		entry := entries[i]
		cardX, cardY, cardW, cardH := eventCodexEntryRect(i - start)
		bg := color.RGBA{28, 22, 16, 225}
		border := color.RGBA{90, 72, 38, 210}
		accent := color.RGBA{140, 120, 72, 220}
		if i == focus {
			bg = color.RGBA{74, 55, 28, 240}
			border = color.RGBA{226, 182, 92, 255}
			accent = border
		}
		drawRoundedRect(screen, cardX, cardY, cardW, cardH, 6, bg)
		vector.StrokeRect(screen, cardX, cardY, cardW, cardH, 1, border, false)
		vector.FillRect(screen, cardX, cardY, 4, cardH, accent, false)
		title := trimTextToWidth(codexStatusIcon(entry.Status)+" "+entry.Title, FaceSmall, float64(cardW)-24)
		DrawText(screen, title, float64(cardX)+12, float64(cardY)+8, FaceSmall, eventCodexLineColor(codexStatusIcon(entry.Status)))
		meta := entry.DateLabel + " • " + entry.Status
		if entry.MonthsUntil > 0 {
			meta += fmt.Sprintf(" • %d ay", entry.MonthsUntil)
		}
		DrawText(screen, trimTextToWidth(meta, FaceSmall, float64(cardW)-24), float64(cardX)+12, float64(cardY)+24, FaceSmall, ColorGray)
		summary := trimTextToWidth(entry.Summary, FaceSmall, float64(cardW)-24)
		DrawText(screen, summary, float64(cardX)+12, float64(cardY)+42, FaceSmall, color.RGBA{196, 184, 160, 230})
	}
	drawEventCodexScrollbar(screen, len(entries), visibleCount, scroll)

	selected := entries[focus]
	DrawText(screen, selected.Title, layout.detailRect.X+16, layout.detailRect.Y+16, FaceMed, eventCodexLineColor(codexStatusIcon(selected.Status)))
	meta := selected.DateLabel + " • " + selected.Status
	if selected.MonthsUntil > 0 {
		meta += fmt.Sprintf(" • %d ay", selected.MonthsUntil)
	}
	DrawText(screen, meta, layout.detailRect.X+16, layout.detailRect.Y+40, FaceSmall, ColorGray)

	drawUIWrappedLabel(screen, gameui.Rect{X: layout.detailRect.X + 16, Y: layout.detailRect.Y + 68, W: layout.detailRect.W - 32}, selected.Detail, eventCodexLineColor(selected.Detail), gameui.TextMedium, 19, int((layout.detailRect.H-76)/19))
}

func drawVictoryDetailPopup(screen *ebiten.Image, gs *state.GameState, scroll float64) {
	modal := buildVictoryDetailModal()
	gameui.DrawModal(screen, modal, eventDetailModalStyle, nil, nil)

	layout := buildVictoryDetailLayout()
	drawUIPanelTopBar(screen, layout.panelRect, 3, panelBorder)

	opt, hasOpt := currentVictoryOption(gs)
	title := victoryTypeLabel(gs.Victory.Type)
	if hasOpt {
		if opt.Title != "" {
			title = opt.Title
		}
	}

	DrawText(screen, title, layout.titleRect.X, layout.titleRect.Y+6, FaceLarge, ColorGold)

	closeBtn := buildVictoryDetailCloseButton()
	drawUIButtonWidget(screen, closeBtn, tinyButtonStyle)

	content := buildVictoryDetailContentLines(gs)
	scroll = clampVictoryDetailScroll(gs, scroll)
	bodyX := layout.bodyRect.X
	bodyY := layout.bodyRect.Y
	viewTop := layout.scrollRect.Y
	viewBottom := layout.scrollRect.Y + layout.scrollRect.H
	contentY := bodyY - scroll
	for _, line := range content.lines {
		lineBottom := contentY + line.lineStep
		if lineBottom > viewTop && contentY < viewBottom {
			drawUILabel(screen, gameui.Rect{X: bodyX, Y: contentY, W: layout.bodyRect.W}, line.text, line.color, line.variant, gameui.TextAlignStart)
		}
		contentY += line.lineStep
	}
	drawVictoryDetailScrollbar(screen, gs, scroll)
}

func eventCodexEntryRect(index int) (x, y, w, h float32) {
	layout := buildEventCodexLayout()
	const (
		cardH = 60.0
		gap   = 8.0
		pad   = 10.0
	)
	x = float32(layout.listRect.X + pad)
	y = float32(layout.listRect.Y + pad + float64(index)*(cardH+gap))
	w = float32(layout.listRect.W - pad*2)
	h = float32(cardH)
	return x, y, w, h
}

func eventCodexEntryHit(mx, my float64, count int, scroll int) int {
	visibleCount := eventCodexVisibleCount()
	for i := 0; i < visibleCount; i++ {
		entryIndex := scroll + i
		if entryIndex >= count {
			break
		}
		x, y, w, h := eventCodexEntryRect(i)
		if mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h) {
			return entryIndex
		}
	}
	return -1
}

func eventCodexVisibleCount() int {
	layout := buildEventCodexLayout()
	const (
		cardH = 60.0
		gap   = 8.0
		pad   = 10.0
	)
	usable := layout.listRect.H - pad*2
	if usable < cardH {
		return 1
	}
	count := int((usable + gap) / (cardH + gap))
	if count < 1 {
		return 1
	}
	return count
}

func eventCodexMaxScroll(count int) int {
	maxScroll := count - eventCodexVisibleCount()
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func eventCodexListHit(mx, my float64) bool {
	layout := buildEventCodexLayout()
	return mx >= layout.listRect.X && mx <= layout.listRect.X+layout.listRect.W &&
		my >= layout.listRect.Y && my <= layout.listRect.Y+layout.listRect.H
}

func drawEventCodexScrollbar(screen *ebiten.Image, count int, visibleCount int, scroll int) {
	if count <= visibleCount || visibleCount <= 0 {
		return
	}
	layout := buildEventCodexLayout()
	const pad = 10.0
	trackX := float32(layout.listRect.X + layout.listRect.W - pad - 3)
	trackY := float32(layout.listRect.Y + pad)
	trackH := float32(layout.listRect.H - pad*2)
	vector.FillRect(screen, trackX, trackY, 3, trackH, color.RGBA{56, 46, 28, 210}, false)
	thumbH := trackH * float32(visibleCount) / float32(count)
	if thumbH < 28 {
		thumbH = 28
	}
	maxScroll := eventCodexMaxScroll(count)
	thumbY := trackY
	if maxScroll > 0 {
		thumbY += (trackH - thumbH) * float32(scroll) / float32(maxScroll)
	}
	vector.FillRect(screen, trackX-1, thumbY, 5, thumbH, color.RGBA{180, 145, 70, 220}, false)
}

func eventCodexLineColor(line string) color.RGBA {
	switch {
	case strings.HasPrefix(line, "[+]"):
		return color.RGBA{132, 214, 132, 240}
	case strings.HasPrefix(line, "[~]"):
		return color.RGBA{226, 196, 104, 240}
	case strings.HasPrefix(line, "[!]"):
		return color.RGBA{218, 120, 120, 240}
	case strings.HasPrefix(line, "Kritik eksik:") || strings.HasPrefix(line, "Neden:") || strings.HasPrefix(line, "Kritik eksik"):
		return color.RGBA{212, 154, 154, 235}
	case strings.HasPrefix(line, "Kalan süre:"):
		return color.RGBA{214, 190, 120, 235}
	case strings.HasPrefix(line, "Koşullar sağlanıyor."):
		return color.RGBA{150, 208, 150, 235}
	case strings.HasPrefix(line, "Bekleyen tarihsel zincirler"):
		return color.RGBA{200, 184, 142, 235}
	default:
		return color.RGBA{230, 224, 205, 240}
	}
}

func codexStatusIcon(status string) string {
	switch status {
	case "Hazir":
		return "[+]"
	case "Takvim":
		return "[~]"
	case "Kilitli":
		return "[!]"
	default:
		return "[-]"
	}
}

func eventDetailLines(message string, maxWidth float64) []string {
	parts := strings.Split(message, "\n")
	lines := make([]string, 0, len(parts)+4)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			lines = append(lines, "")
			continue
		}
		wrapped := wrapTextLines(part, FaceMed, maxWidth)
		lines = append(lines, wrapped...)
	}
	return lines
}

func infoPopupHeight(message string) float32 {
	const pw = float32(430)

	lines := wrapTextLines(message, FaceMed, float64(pw-40))
	lineCount := len(lines)
	if lineCount > 3 {
		lineCount = 3
	}
	return float32(48 + lineCount*20)
}

func drawInfoPopupAt(screen *ebiten.Image, message string, alpha uint8, py float32) {
	const pw = float32(430)

	px := float32(ScreenWidth)/2 - pw/2
	ph := infoPopupHeight(message)

	bgAlpha := alpha
	if bgAlpha > 235 {
		bgAlpha = 235
	}
	drawRoundedRect(screen, px, py, pw, ph, 8, color.RGBA{18, 14, 10, bgAlpha})
	vector.StrokeRect(screen, px, py, pw, ph, 1.5, color.RGBA{130, 105, 55, alpha}, false)
	vector.FillRect(screen, px, py, pw, 3, color.RGBA{210, 170, 65, alpha}, false)

	drawUILabel(screen, gameui.Rect{X: float64(px) + 16, Y: float64(py) + 12}, "Bilgi", color.RGBA{220, 190, 100, alpha}, gameui.TextSmall, gameui.TextAlignStart)
	drawUIWrappedLabel(screen, gameui.Rect{X: float64(px) + 20, Y: float64(py) + 34, W: float64(pw - 40)}, message, color.RGBA{240, 230, 205, alpha}, gameui.TextMedium, 20, 3)
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

	// Dünya->minimap ölçeği viewport dikdörtgeni için kullanılır.
	scaleX := float64(minimapW) / float64(WorldW)
	scaleY := float64(minimapH) / float64(WorldH)

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

func drawMinimapEventMarkers(screen *ebiten.Image, gs *state.GameState, scaleX, scaleY, offsetX, offsetY float32) {
	if gs == nil || len(gs.ActiveRegionEvents) == 0 {
		return
	}

	const markerRadius = float32(3.25)
	const markerStackGap = float32(5.5)

	for i := range gs.ActiveRegionEvents {
		evt := gs.ActiveRegionEvents[i]
		if !activeRegionEventVisible(gs, evt) {
			continue
		}
		region, ok := gs.Regions[evt.RegionID]
		if !ok || region == nil || region.IsSea {
			continue
		}

		px, py := minimapEventMarkerPosition(region, scaleX, scaleY, offsetX, offsetY)

		stack := minimapEventMarkerStackOffset(gs.ActiveRegionEvents, i, evt.RegionID)
		if stack > 0 {
			py -= markerStackGap * float32(stack)
		}

		fill, border := minimapEventMarkerPalette(evt.Type)
		vector.FillCircle(screen, px+1, py+1, markerRadius+1, color.RGBA{0, 0, 0, 90}, true)
		vector.FillCircle(screen, px, py, markerRadius, fill, true)
		vector.StrokeCircle(screen, px, py, markerRadius, 1.1, border, true)
	}
}

func minimapEventMarkerPosition(region *world.Region, scaleX, scaleY, offsetX, offsetY float32) (float32, float32) {
	if region == nil {
		return offsetX, offsetY
	}
	return offsetX + float32(wcX(region.WorldX))*scaleX, offsetY + float32(wcY(region.WorldY))*scaleY
}

func minimapEventMarkerStackOffset(events []state.RegionEventStatus, idx int, rid world.RegionID) int {
	stack := 0
	for j := 0; j < idx; j++ {
		if events[j].RegionID == rid {
			stack++
		}
	}
	return stack
}

func minimapEventMarkerPalette(eventType string) (color.RGBA, color.RGBA) {
	switch eventType {
	case "plague":
		return color.RGBA{180, 58, 58, 235}, color.RGBA{72, 14, 14, 235}
	case "famine":
		return color.RGBA{204, 152, 54, 235}, color.RGBA{82, 54, 12, 235}
	case "revolt":
		return color.RGBA{206, 58, 74, 235}, color.RGBA{70, 8, 18, 235}
	case "blessing":
		return color.RGBA{84, 184, 84, 235}, color.RGBA{18, 72, 18, 235}
	default:
		return color.RGBA{198, 174, 82, 235}, color.RGBA{78, 56, 12, 235}
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
	DrawRegionPanelExpandedScrolled(screen, gs, rid, true, 0)
}

func DrawRegionPanelExpanded(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, neighborExpanded bool) {
	DrawRegionPanelExpandedScrolled(screen, gs, rid, neighborExpanded, 0)
}

func DrawRegionPanelExpandedScrolled(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, neighborExpanded bool, scroll float64) {
	DrawRegionPanelExpandedScrolledWithTab(screen, gs, rid, neighborExpanded, regionPanelTabBuildings, scroll)
}

func DrawRegionPanelExpandedScrolledWithTab(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, neighborExpanded bool, activeTab regionPanelTab, scroll float64) {
	if rid == "" {
		return
	}
	region, ok := gs.Regions[rid]
	if !ok {
		return
	}

	if region.IsSea {
		DrawSeaRegionPanel(screen, gs, region, neighborExpanded)
		return
	}

	px := infoPanelX()
	py := infoPanelY()
	pw := infoPanelW
	ph := infoPanelH

	drawUIPanelFrame(screen, gameui.Rect{X: float64(px), Y: float64(py), W: float64(pw), H: float64(ph)}, panelBg, panelBorder, 1.5, 3)
	drawPanelCloseButton(screen, px, py, pw)

	lx := float64(px) + panelPad
	ly := float64(py) + 10
	sepW := pw - float32(panelPad*2)
	production := gs.RegionProductionSummary(region)
	ownerName, ownerCol := ownerDisplay(gs, region.OwnerID)

	if region.OwnerID != "" {
		badgeBG := color.Color(ColorGray)
		if ownerFaction := gs.Factions[faction.FactionID(region.OwnerID)]; ownerFaction != nil {
			badgeBG = color.RGBA{ownerFaction.Color[0], ownerFaction.Color[1], ownerFaction.Color[2], 255}
		}
		ownerFlag := regionPanelOwnerFlagRect(lx, float64(py))
		drawFactionFlagBadge(screen, faction.FactionID(region.OwnerID), factionInitial(ownerName), ownerFlag.X, ownerFlag.Y, ownerFlag.W, badgeBG, panelBorder)

		if empireID := regionImperialEmpireID(gs, region.OwnerID); empireID != "" {
			empireName, _ := ownerDisplay(gs, string(empireID))
			empireFaction := gs.Factions[empireID]
			empireBG := color.Color(ColorGray)
			if empireFaction != nil {
				empireBG = color.RGBA{empireFaction.Color[0], empireFaction.Color[1], empireFaction.Color[2], 255}
			}
			empireFlag := regionPanelImperialFlagRect(lx, float64(py))
			drawFactionFlagBadge(screen, empireID, factionInitial(empireName), empireFlag.X, empireFlag.Y, empireFlag.W, empireBG, color.RGBA{212, 172, 76, 255})
		}
	}

	DrawText(screen, region.NameTR, lx, ly, FaceLarge, ColorYellow)
	ly += 24

	drawUIOutlinedLabel(screen, gameui.Rect{X: lx, Y: ly, W: float64(sepW)}, ownerName, ownerCol, ownerLabelOutlineColor(ownerCol), gameui.TextLarge, gameui.TextAlignStart)
	if ownerRect, _, ok := regionOwnerNameRect(gs, rid); ok {
		underlineY := ownerRect.Y + ownerRect.H - 1
		vector.StrokeLine(screen, float32(ownerRect.X), float32(underlineY), float32(ownerRect.X+ownerRect.W), float32(underlineY), 1, color.RGBA{215, 215, 215, 120}, false)
	}
	ly += float64(regionOwnerNameH)
	if overlordLabel, overlordCol, ok := vassalOverlordDisplay(gs, region.OwnerID); ok {
		drawUILabel(screen, gameui.Rect{X: lx, Y: ly, W: float64(sepW)}, "Bağlı: "+overlordLabel, overlordCol, gameui.TextSmall, gameui.TextAlignStart)
		ly += regionVassalInfoH
	}
	if tributeLabel, tributeCol, ok := vassalTributeDisplay(gs, region.OwnerID); ok {
		drawUILabel(screen, gameui.Rect{X: lx, Y: ly, W: float64(sepW)}, tributeLabel, tributeCol, gameui.TextSmall, gameui.TextAlignStart)
		ly += regionVassalInfoH
	}

	// Development mode bilgileri
	if gs.DevelopmentMode {
		drawUIRichTextBlock(screen, gameui.Rect{X: lx, Y: ly}, []gameui.RichTextLine{
			{Text: "ID: " + string(region.ID), Color: ColorGray, Variant: gameui.TextSmall, Align: gameui.TextAlignStart},
			{Text: "Koordinat: " + itoa(region.WorldX) + "," + itoa(region.WorldY), Color: ColorGray, Variant: gameui.TextSmall, Align: gameui.TextAlignStart},
		}, 16)
		ly += 34
	}

	var stypeStr string
	if len(region.Settlements) > 0 {
		capital := region.Settlements[0]
		for _, s := range region.Settlements {
			if s.IsCenter {
				capital = s
				break
			}
		}
		stypeStr = "  |  " + capital.Type.LabelTR()
	}

	drawUILabel(screen, gameui.Rect{X: lx, Y: ly, W: float64(sepW)}, region.Terrain.LabelTR()+"  |  "+religion.DisplayNameTR(religion.Type(region.Religion))+stypeStr, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	ly += 16
	drawUILabel(screen, gameui.Rect{X: lx, Y: ly, W: float64(sepW)}, regionPopulationDisplayText(region), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
	ly += 16

	drawUISeparator(screen, float32(lx), float32(ly), float32(lx)+sepW, 1, panelBorder)
	ly += 8

	// Üretim — iki sütun
	drawRegionProductionGrid(screen, gs, region, lx, ly, sepW, production)
	ly += regionPanelStatRowGap * 4
	blockadeEffect := gs.RegionBlockadeEconomicEffect(region)
	if blockadeEffect.BlockadePercent > 0 {
		lossPercent := 100 - blockadeEffect.OutputRetentionPercent
		drawUILabel(
			screen,
			gameui.Rect{X: lx, Y: ly, W: float64(sepW)},
			fmt.Sprintf("Abluka %%%d • sonraki tur yerel -%%%d", blockadeEffect.BlockadePercent, lossPercent),
			color.RGBA{225, 135, 45, 255},
			gameui.TextSmall,
			gameui.TextAlignStart,
		)
		ly += 16
	}

	drawRegionMeterRow(screen, lx, ly, sepW, "Memnuniyet", "%"+itoa(region.Satisfaction), float64(region.Satisfaction)/100, satisfactionColor(region.Satisfaction))
	ly += regionPanelStatRowGap

	taxBarX, taxBarW := regionPanelTaxBarLayout(float32(lx), sepW)
	if region.OwnerID == string(gs.PlayerFactionID) && !region.IsLocked {
		drawRegionMeterLabels(screen, lx, ly, "Vergi", "%"+itoa(region.TaxRate))
		dec, inc := regionTaxButtonRects(gs, rid)
		taxBarX, taxBarW = regionPanelTaxInteractiveBarLayout(float32(lx), sepW, dec[0])
		drawBar(screen, taxBarX, float32(ly)+regionPanelBarYOffset, taxBarW, regionPanelBarH, float64(region.TaxRate)/100, color.RGBA{200, 140, 40, 255})
		drawTinyPanelButton(screen, dec[0], dec[1], dec[2], dec[3], "-", true)
		drawTinyPanelButton(screen, inc[0], inc[1], inc[2], inc[3], "+", true)
	} else {
		drawRegionMeterRow(screen, lx, ly, sepW, "Vergi", "%"+itoa(region.TaxRate), float64(region.TaxRate)/100, color.RGBA{200, 140, 40, 255})
	}
	ly += regionPanelStatRowGap

	if logistics, ok := gs.RegionLogistics[rid]; ok && logistics.Demand > 0 {
		meter := float64(logistics.Demand) / float64(maxInt(1, logistics.Capacity))
		if meter > 1 {
			meter = 1
		}
		drawRegionMeterRow(screen, lx, ly, sepW, "İkmal", fmt.Sprintf("%d / %d", logistics.Capacity, logistics.Demand), meter, logisticsPressureColor(logistics))
		ly += regionPanelStatRowGap
		if logistics.FriendlySupplyGrainSpent > 0 {
			drawUILabel(
				screen,
				gameui.Rect{X: lx, Y: ly, W: float64(sepW)},
				fmt.Sprintf("Dost ikmali: %d ordu  •  -%d tahıl/tur", logistics.FriendlySupplyArmies, logistics.FriendlySupplyGrainSpent),
				color.RGBA{137, 199, 141, 245},
				gameui.TextSmall,
				gameui.TextAlignStart,
			)
			ly += 14
		}
		if logistics.Overload > 0 {
			drawUILabel(
				screen,
				gameui.Rect{X: lx, Y: ly, W: float64(sepW)},
				fmt.Sprintf("Aşım %d  •  zayiat %d HP", logistics.Overload, logistics.TotalHPDamage),
				ColorRed,
				gameui.TextSmall,
				gameui.TextAlignStart,
			)
			ly += 14
		}
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
			drawUILabel(screen, gameui.Rect{X: lx, Y: ly, W: float64(sepW)}, "☩ Dönüşüm: "+religion.DisplayNameTR(religion.Type(ownerRel)), color.RGBA{180, 140, 240, 200}, gameui.TextSmall, gameui.TextAlignStart)
			ly += 14
			drawBar(screen, float32(lx), float32(ly), sepW, 7, convPct, color.RGBA{150, 100, 220, 220})
			ly += 12
		}
	}

	if region.IsRebellionRisk() {
		DrawText(screen, "⚠  İSYAN RİSKİ!", lx, ly, FaceMed, ColorRed)
		ly += 18
	}

	// ── Binalar / olaylar sekmeli ortak içerik alanı ──────────────────
	ly += 4
	drawUISeparator(screen, float32(lx), float32(ly), float32(lx)+sepW, 1, panelBorder)
	ly += 6

	buildingStartY := float32(ly) + regionPanelTabH + 6
	buildingEndY := buildingGridEndY(gs, region, buildingStartY)
	drawRegionPanelTabs(screen, gs, region, activeTab)

	contentEndY := buildingEndY
	if activeTab == regionPanelTabEvents {
		contentEndY = regionPanelTabContentEndY(gs, region, buildingStartY, buildingEndY)
		activityViewport := gameui.Rect{
			X: lx,
			Y: float64(buildingStartY),
			W: float64(sepW),
			H: float64(contentEndY - buildingStartY),
		}
		if activityViewport.H > regionPanelActivityMinH {
			drawRegionActivityNeighborSection(screen, gs, region, activityViewport, scroll, neighborExpanded)
		}
	} else {
		drawBuildingGrid(screen, gs, region, px, buildingStartY, pw)
	}

	// Sekme içeriğinin hemen altındaki aksiyon bandı korunur.
	actionBarY := float64(contentEndY) + 5
	drawRegionActionBar(screen, gs, region, px, float32(actionBarY), pw)
}

func regionPanelTabRects(gs *state.GameState, region *world.Region) (gameui.Rect, gameui.Rect) {
	startY := buildingGridStartY(gs, region, false)
	y := float64(startY) - regionPanelTabH - 6
	x := float64(infoPanelX()) + panelPad
	w := float64(infoPanelW) - panelPad*2
	tabW := (w - regionPanelTabGap) / 2
	return gameui.Rect{X: x, Y: y, W: tabW, H: regionPanelTabH},
		gameui.Rect{X: x + tabW + regionPanelTabGap, Y: y, W: tabW, H: regionPanelTabH}
}

func drawRegionPanelTabs(screen *ebiten.Image, gs *state.GameState, region *world.Region, activeTab regionPanelTab) {
	buildingTab, eventTab := regionPanelTabRects(gs, region)
	for tab, rect := range []gameui.Rect{buildingTab, eventTab} {
		selected := regionPanelTab(tab) == activeTab
		fill := color.RGBA{24, 22, 18, 230}
		border := color.RGBA{90, 78, 48, 255}
		textCol := ColorGray
		if selected {
			fill = color.RGBA{72, 56, 20, 245}
			border = color.RGBA{220, 176, 55, 255}
			textCol = ColorGold
		}
		drawUICardRect(screen, rect, fill, border, 1)
		label := "BİNALAR"
		if tab == int(regionPanelTabEvents) {
			label = "OLAYLAR"
		}
		drawUILabel(screen, gameui.Rect{X: rect.X, Y: rect.Y + 4, W: rect.W, H: rect.H - 4}, label, textCol, gameui.TextSmall, gameui.TextAlignCenter)
	}
}

func regionPanelTabHit(mx, my float64, gs *state.GameState, rid world.RegionID) (regionPanelTab, bool) {
	if gs == nil || rid == "" {
		return regionPanelTabBuildings, false
	}
	region := gs.Regions[rid]
	if region == nil {
		return regionPanelTabBuildings, false
	}
	buildingTab, eventTab := regionPanelTabRects(gs, region)
	if buildingTab.Hit(mx, my) {
		return regionPanelTabBuildings, true
	}
	if eventTab.Hit(mx, my) {
		return regionPanelTabEvents, true
	}
	return regionPanelTabBuildings, false
}

func regionPanelTabContentEndY(gs *state.GameState, region *world.Region, startY, buildingEndY float32) float32 {
	endY := buildingEndY
	minEndY := startY + regionPanelTabMinContentH
	if endY < minEndY {
		endY = minEndY
	}
	panelBottom := float32(infoPanelY()+infoPanelH) - 10 - regionPanelActionBarHeight - 5
	if endY > panelBottom {
		endY = panelBottom
	}
	return endY
}

func regionActiveEventCount(gs *state.GameState, rid world.RegionID) int {
	if gs == nil || rid == "" {
		return 0
	}
	count := 0
	for i := range gs.ActiveRegionEvents {
		evt := gs.ActiveRegionEvents[i]
		if evt.RegionID == rid && evt.TurnsLeft > 0 {
			count++
		}
	}
	return count
}

func regionEventStatusColor(eventType string) color.RGBA {
	switch eventType {
	case "plague", "revolt":
		return color.RGBA{235, 105, 85, 255}
	case "famine":
		return color.RGBA{230, 180, 75, 255}
	case "blessing":
		return color.RGBA{105, 205, 115, 255}
	default:
		return ColorGold
	}
}

func regionActivityNeighborVisible(gs *state.GameState, region *world.Region) bool {
	return gs != nil && region != nil && (regionActiveEventCount(gs, region.ID) > 0 || len(region.Neighbors) > 0 || gs.DevelopmentMode)
}

func buildingGridEndY(gs *state.GameState, region *world.Region, startY float32) float32 {
	count := len(visibleBuildingIDs(gs, region))
	if count == 0 {
		return startY
	}
	rows := (count + 2) / 3
	cardH := buildingGridSpriteHeight(infoPanelW) + buildingGridNameH
	return startY + float32(rows)*cardH + float32(rows-1)*buildingGridRowGap
}

func drawRegionActionBar(screen *ebiten.Image, gs *state.GameState, region *world.Region, px, y, pw float32) {
	bar := gameui.Rect{X: float64(px) + panelPad, Y: float64(y), W: float64(pw) - panelPad*2, H: regionPanelActionBarHeight}
	drawUICardRect(screen, bar, color.RGBA{20, 19, 16, 225}, panelBorder, 1)
	if region == nil || region.OwnerID == "" {
		return
	}
	if region.OwnerID == string(gs.PlayerFactionID) {
		if _, ok := regionLiberationSuccessor(gs, region); ok {
			liberateBtn := buildRegionLiberateButton(float32(bar.X), float32(bar.Y), float32(bar.W), float32(bar.H))
			drawUIButtonWidget(screen, liberateBtn, solidButtonStyle(color.RGBA{76, 112, 66, 235}, color.RGBA{145, 190, 108, 255}, ColorWhite, 0))
		}
		btn := buildRegionGrainAidButton(gs, region.ID, float32(bar.X), float32(bar.Y), float32(bar.W), float32(bar.H))
		drawUIButton(screen, btn.X, btn.Y, btn.W, btn.H, btn.Label, gs.CanApplyGrainAid(region.ID), solidButtonStyle(color.RGBA{112, 82, 36, 225}, color.RGBA{184, 142, 70, 255}, ColorWhite, 0))
		return
	}
	btn := buildRegionDiplomacyButtons(gs, region.OwnerID, float32(bar.X), float32(bar.Y), float32(bar.W), float32(bar.H))
	drawUIButtonWidget(screen, btn, solidButtonStyle(color.RGBA{55, 92, 142, 225}, panelBorder, ColorWhite, 0))
}

func regionActivityNeighborContentHeight(gs *state.GameState, region *world.Region) float64 {
	return regionActivityNeighborContentHeightForExpanded(gs, region, true)
}

func regionActivityNeighborContentHeightForExpanded(gs *state.GameState, region *world.Region, expanded bool) float64 {
	height := 8.0
	if eventCount := regionActiveEventCount(gs, region.ID); eventCount > 0 {
		height += 17 + float64(eventCount)*28 + 6
	}
	_, _, _, rows := neighborBlockLayout(gs, region, expanded)
	height += devNeighborTitleHeight + float64(rows)*devNeighborLineHeight
	return height
}

func clampRegionPanelValue(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func clampRegionPanelScroll(gs *state.GameState, rid world.RegionID, scroll float64) float64 {
	return clampRegionPanelScrollForNeighbors(gs, rid, scroll, true)
}

func clampRegionPanelScrollForNeighbors(gs *state.GameState, rid world.RegionID, scroll float64, neighborExpanded bool) float64 {
	if gs == nil || rid == "" {
		return 0
	}
	region := gs.Regions[rid]
	if region == nil {
		return 0
	}
	viewportHeight := regionPanelActivityViewport(gs, region).H
	if viewportHeight < regionPanelActivityMinH {
		return 0
	}
	maxScroll := regionActivityNeighborContentHeightForExpanded(gs, region, neighborExpanded) - viewportHeight
	return clampRegionPanelValue(scroll, 0, maxFloat64Value(maxScroll))
}

func maxFloat64Value(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

func regionPanelActivityViewport(gs *state.GameState, region *world.Region) gameui.Rect {
	if gs == nil || region == nil {
		return gameui.Rect{}
	}
	start := buildingGridStartY(gs, region, false)
	end := regionPanelTabContentEndY(gs, region, start, buildingGridEndY(gs, region, start))
	return gameui.Rect{
		X: float64(infoPanelX()) + panelPad,
		Y: float64(start),
		W: float64(infoPanelW) - panelPad*2,
		H: float64(end - start),
	}
}

func regionPanelActivityHit(mx, my float64, gs *state.GameState, rid world.RegionID) bool {
	if gs == nil {
		return false
	}
	region := gs.Regions[rid]
	viewport := regionPanelActivityViewport(gs, region)
	return viewport.H > regionPanelActivityMinH && viewport.Hit(mx, my)
}

func regionActiveEventPanelHit(mx, my float64, gs *state.GameState, region *world.Region, scroll float64) (int, bool) {
	if gs == nil || region == nil {
		return -1, false
	}
	viewport := regionPanelActivityViewport(gs, region)
	if viewport.H <= regionPanelActivityMinH || !viewport.Hit(mx, my) || regionActiveEventCount(gs, region.ID) == 0 {
		return -1, false
	}
	_, y := regionActivityContentOrigin(viewport, scroll)
	y += 17
	for i := range gs.ActiveRegionEvents {
		evt := gs.ActiveRegionEvents[i]
		if evt.RegionID != region.ID || evt.TurnsLeft <= 0 {
			continue
		}
		row := gameui.Rect{X: viewport.X, Y: y - 2, W: viewport.W, H: 28}
		if row.Hit(mx, my) && row.Y+row.H >= viewport.Y && row.Y <= viewport.Y+viewport.H {
			return i, true
		}
		y += 28
	}
	return -1, false
}

func regionActivityNeighborToggleHit(mx, my float64, gs *state.GameState, region *world.Region, scroll float64) bool {
	if gs == nil || region == nil || len(region.Neighbors) <= devNeighborCollapsedCount {
		return false
	}
	viewport := regionPanelActivityViewport(gs, region)
	if viewport.H <= regionPanelActivityMinH || !viewport.Hit(mx, my) {
		return false
	}
	_, y := regionActivityContentOrigin(viewport, scroll)
	if regionActiveEventCount(gs, region.ID) > 0 {
		y += 17 + float64(regionActiveEventCount(gs, region.ID))*28 + 6
	}
	toggle := gameui.Rect{X: viewport.X, Y: y, W: viewport.W, H: devNeighborTitleHeight}
	return toggle.Hit(mx, my)
}

func regionActivityContentOrigin(viewport gameui.Rect, scroll float64) (float64, float64) {
	return viewport.X + panelPad, viewport.Y + 8.0 - scroll
}

func regionActivityContentCenterX(viewport gameui.Rect) float64 {
	return viewport.X + viewport.W/2
}

func drawRegionActivityNeighborSection(screen *ebiten.Image, gs *state.GameState, region *world.Region, viewport gameui.Rect, scroll float64, neighborExpanded bool) {
	contentHeight := regionActivityNeighborContentHeightForExpanded(gs, region, neighborExpanded)
	scroll = clampRegionPanelValue(scroll, 0, maxFloat64Value(contentHeight-viewport.H))
	drawUICardRect(screen, viewport, color.RGBA{16, 15, 13, 220}, panelBorder, 1)

	left := int(viewport.X)
	top := int(viewport.Y)
	right := int(viewport.X + viewport.W)
	bottom := int(viewport.Y + viewport.H)
	if right <= left || bottom <= top {
		return
	}
	body := screen.SubImage(image.Rect(left, top, right, bottom)).(*ebiten.Image)
	// Ebiten SubImage hedefinin Bounds değeri ana ekran koordinatlarını korur.
	// Bu nedenle içerik koordinatlarını yerel (12, 8) yerine viewport'un
	// ekran koordinatında üretmek gerekir; aksi halde metinler ekranın üstüne
	// çizilip viewport clipping'i tarafından görünmez olur.
	x, y := regionActivityContentOrigin(viewport, scroll)
	width := float32(viewport.W)

	if eventCount := regionActiveEventCount(gs, region.ID); eventCount > 0 {
		drawUICenteredSectionLabel(body, regionActivityContentCenterX(viewport), y, "AKTİF OLAYLAR")
		y += 17
		for i := range gs.ActiveRegionEvents {
			evt := gs.ActiveRegionEvents[i]
			if evt.RegionID != region.ID || evt.TurnsLeft <= 0 {
				continue
			}
			label := evt.LabelTR
			if label == "" {
				label = evt.EventID
			}
			drawUILabel(body, gameui.Rect{X: x, Y: y, W: float64(width)}, "• "+label, regionEventStatusColor(evt.Type), gameui.TextSmall, gameui.TextAlignStart)
			y += 14
			drawUILabel(body, gameui.Rect{X: x + 12, Y: y, W: float64(width) - 12}, activeRegionEventTypeLabel(evt.Type)+"  •  Kalan tur: "+itoa(evt.TurnsLeft), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
			y += 14
		}
		y += 6
	}

	drawNeighborBlock(body, gs, region, x, y, width, neighborExpanded, color.RGBA{200, 170, 90, 220})
	drawRegionPanelScrollbar(screen, viewport, contentHeight, scroll)
}

func drawRegionPanelScrollbar(screen *ebiten.Image, viewport gameui.Rect, contentHeight, scroll float64) {
	maxScroll := contentHeight - viewport.H
	if maxScroll <= 0 {
		return
	}
	track := gameui.Rect{X: viewport.X + viewport.W - 7, Y: viewport.Y + 4, W: 4, H: viewport.H - 8}
	thumbH := track.H * viewport.H / contentHeight
	if thumbH < 18 {
		thumbH = 18
	}
	thumbY := track.Y + (track.H-thumbH)*(scroll/maxScroll)
	drawRoundedRect(screen, float32(track.X), float32(track.Y), float32(track.W), float32(track.H), 2, color.RGBA{70, 65, 55, 180})
	drawRoundedRect(screen, float32(track.X), float32(thumbY), float32(track.W), float32(thumbH), 2, color.RGBA{210, 175, 85, 230})
}

func regionDiplomacyButtonRect(_ int, px, py, pw, ph float32) (x, y, w, h float32) {
	btnW := float32(92)
	btnH := float32(24)
	x = px + pw - btnW - 5
	y = py + (ph-btnH)/2
	return x, y, btnW, btnH
}

func drawRegionDiplomacyButtons(screen *ebiten.Image, gs *state.GameState, ownerID string, px, py, pw, ph float32) {
	btnCol := color.RGBA{55, 92, 142, 225}
	btn := buildRegionDiplomacyButtons(gs, ownerID, px, py, pw, ph)
	drawUIButtonWidget(screen, btn, solidButtonStyle(btnCol, panelBorder, ColorWhite, 0))
}

func logisticsPressureColor(status state.RegionLogisticsStatus) color.RGBA {
	switch {
	case status.Overload <= 0:
		return color.RGBA{90, 170, 95, 255}
	case status.Overload*100 <= maxInt(1, status.Capacity)*25:
		return color.RGBA{210, 165, 50, 255}
	default:
		return color.RGBA{190, 70, 60, 255}
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
	py := infoPanelY() + infoPanelH - 148
	pw := infoPanelW
	ph := float32(148)

	drawUIPanelFrame(screen, gameui.Rect{X: float64(px), Y: float64(py), W: float64(pw), H: float64(ph)}, panelBg, panelBorder, 1.5, 3)
	drawPanelCloseButton(screen, px, py, pw)

	lx := float64(px) + panelPad
	ly := float64(py) + 10

	DrawText(screen, "Seçili Ordu", lx, ly, FaceLarge, ColorYellow)
	ly += 22

	if region, ok2 := gs.Regions[a.RegionID]; ok2 {
		drawUIKeyValueRow(screen, lx, ly, float64(pw)-panelPad*2, "Konum", region.NameTR, ColorGray, ColorWhite)
	}
	ly += 18

	drawUIKeyValueRow(screen, lx, ly, float64(pw)-panelPad*2, "Birim", itoa(len(a.Units))+"/"+itoa(army.MaxArmySize), ColorGray, ColorWhite)
	ly += 18

	mpCol := ColorGold
	if a.MovePoints == 0 {
		mpCol = ColorRed
	}
	drawUIKeyValueRow(screen, lx, ly, float64(pw)-panelPad*2, "Hareket", itoa(a.MovePoints)+"/"+itoa(a.MaxMovePoints), ColorGray, mpCol)
	ly += 18
	if grainNeed := gs.EffectiveArmyGrainUpkeep(a); grainNeed > 0 {
		drawUIKeyValueRow(screen, lx, ly, float64(pw)-panelPad*2, "Tahıl", itoa(grainNeed)+" / tur", ColorGray, color.RGBA{205, 185, 120, 235})
		ly += 18
	}
	if logistics, ok := gs.ArmyLogistics[aid]; ok {
		if logistics.FriendlySupplyGrainSpent > 0 {
			drawUIKeyValueRow(screen, lx, ly, float64(pw)-panelPad*2, "Dost ikmali", "-"+itoa(logistics.FriendlySupplyGrainSpent)+" tahıl/tur", ColorGray, color.RGBA{137, 199, 141, 245})
			ly += 18
		}
		if logistics.TotalHPDamage > 0 {
			drawUIKeyValueRow(screen, lx, ly, float64(pw)-panelPad*2, "Lojistik", "-"+itoa(logistics.DamagePerUnit)+" HP / birim", ColorGray, ColorRed)
			ly += 18
		}
	}

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
func drawHistoricalEventPopup(screen *ebiten.Image, title, desc, prompt string, choices []HistoricalEventChoice, focus int) {
	modal := buildHistoricalEventModal()
	gameui.DrawModal(screen, modal, historicalEventModalStyle, nil, nil)

	bx, by, bw, bh := float32(modal.Panel.Rect.X), float32(modal.Panel.Rect.Y), float32(modal.Panel.Rect.W), float32(modal.Panel.Rect.H)
	vector.StrokeRect(screen, bx+4, by+4, bw-8, bh-8, 1, color.RGBA{120, 90, 30, 200}, false)

	// Üst şerit
	vector.FillRect(screen, bx, by, bw, 4, color.RGBA{220, 170, 50, 255}, false)

	cy := float64(by) + 28
	drawUILabel(screen, gameui.Rect{X: 0, Y: cy, W: ScreenWidth}, "- TARIHSEL OLAY -", color.RGBA{180, 140, 50, 200}, gameui.TextSmall, gameui.TextAlignCenter)
	cy += 26
	drawUILabel(screen, gameui.Rect{X: 0, Y: cy, W: ScreenWidth}, title, color.RGBA{255, 220, 80, 255}, gameui.TextLarge, gameui.TextAlignCenter)
	cy += 30

	drawUIWrappedLabel(screen, gameui.Rect{X: float64(bx) + 30, Y: cy, W: float64(bw - 60)}, desc, color.RGBA{210, 200, 180, 230}, gameui.TextMedium, 22, 0)

	if len(choices) == 0 {
		cy = float64(by) + float64(bh) - 28
		drawUILabel(screen, gameui.Rect{X: 0, Y: cy, W: ScreenWidth}, "[Enter / Boşluk / Tıkla] Devam Et", color.RGBA{140, 130, 100, 200}, gameui.TextSmall, gameui.TextAlignCenter)
		return
	}

	promptY := float64(by) + 210
	if prompt != "" {
		drawUILabel(screen, gameui.Rect{X: 0, Y: promptY, W: ScreenWidth}, prompt, color.RGBA{230, 214, 175, 240}, gameui.TextMedium, gameui.TextAlignCenter)
	}

	buttons := buildHistoricalEventChoiceButtons(len(choices))
	for i, choice := range choices {
		if i >= len(buttons) {
			break
		}
		btn := buttons[i]
		active := i == focus
		bg := color.RGBA{78, 62, 36, 235}
		border := color.RGBA{150, 120, 68, 255}
		textCol := color.RGBA{245, 236, 214, 255}
		if active {
			bg = color.RGBA{118, 84, 40, 245}
			border = color.RGBA{226, 182, 92, 255}
		}
		drawUIButton(screen, btn.X, btn.Y, btn.W, btn.H, choice.Label, true, solidButtonStyle(bg, border, textCol, 10))
		drawHistoricalChoiceInfo(screen, btn, choice)
	}
}

const (
	commanderArrivalCardW   = 214.0
	commanderArrivalCardH   = 110.0
	commanderArrivalCardGap = 12.0
)

func commanderArrivalListViewport() gameui.Rect {
	modal := buildHistoricalEventModal()
	panel := modal.Panel.Rect
	top := panel.Y + 126
	bottom := panel.Y + panel.H - 52
	if bottom < top {
		bottom = top
	}
	return gameui.Rect{
		X: panel.X + 24,
		Y: top,
		W: panel.W - 48,
		H: bottom - top,
	}
}

func commanderArrivalVisibleRows(viewport gameui.Rect) int {
	rows := int((viewport.H + commanderArrivalCardGap) / (commanderArrivalCardH + commanderArrivalCardGap))
	if rows < 1 {
		return 1
	}
	return rows
}

func commanderArrivalMaxScroll(count int, viewport gameui.Rect) int {
	rowCount := (count + 2) / 3
	maxScroll := rowCount - commanderArrivalVisibleRows(viewport)
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func clampCommanderArrivalScroll(scroll, count int, viewport gameui.Rect) int {
	if scroll < 0 {
		return 0
	}
	if maxScroll := commanderArrivalMaxScroll(count, viewport); scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func commanderArrivalCardRect(index, scroll int) gameui.Rect {
	modal := buildHistoricalEventModal()
	panel := modal.Panel.Rect
	totalW := commanderArrivalCardW*3 + commanderArrivalCardGap*2
	startX := panel.X + (panel.W-totalW)/2
	return gameui.Rect{
		X: startX + float64(index%3)*(commanderArrivalCardW+commanderArrivalCardGap),
		Y: commanderArrivalListViewport().Y + float64(index/3-scroll)*(commanderArrivalCardH+commanderArrivalCardGap),
		W: commanderArrivalCardW,
		H: commanderArrivalCardH,
	}
}

func drawCommanderArrivalScrollbar(screen *ebiten.Image, viewport gameui.Rect, count, scroll int) {
	maxScroll := commanderArrivalMaxScroll(count, viewport)
	if maxScroll <= 0 {
		return
	}
	scroll = clampCommanderArrivalScroll(scroll, count, viewport)
	track := gameui.Rect{X: viewport.X + viewport.W + 8, Y: viewport.Y + 4, W: 5, H: viewport.H - 8}
	if track.H <= 0 {
		return
	}
	rowCount := (count + 2) / 3
	thumbH := track.H * float64(commanderArrivalVisibleRows(viewport)) / float64(rowCount)
	if thumbH < 18 {
		thumbH = 18
	}
	if thumbH > track.H {
		thumbH = track.H
	}
	thumbY := track.Y
	if track.H > thumbH {
		thumbY += (track.H - thumbH) * float64(scroll) / float64(maxScroll)
	}
	drawRoundedRect(screen, float32(track.X), float32(track.Y), float32(track.W), float32(track.H), 2, color.RGBA{70, 65, 55, 180})
	drawRoundedRect(screen, float32(track.X), float32(thumbY), float32(track.W), float32(thumbH), 2, color.RGBA{210, 175, 85, 230})
}

func drawCommanderArrivalPopup(screen *ebiten.Image, title, desc string, arrivals []*army.Commander, scroll int) {
	modal := buildHistoricalEventModal()
	gameui.DrawModal(screen, modal, historicalEventModalStyle, nil, nil)
	bx, by, bw, bh := float32(modal.Panel.Rect.X), float32(modal.Panel.Rect.Y), float32(modal.Panel.Rect.W), float32(modal.Panel.Rect.H)
	vector.StrokeRect(screen, bx+4, by+4, bw-8, bh-8, 1, color.RGBA{120, 90, 30, 200}, false)
	vector.FillRect(screen, bx, by, bw, 4, color.RGBA{220, 170, 50, 255}, false)
	drawUILabel(screen, gameui.Rect{X: 0, Y: float64(by) + 28, W: ScreenWidth}, "- KOMUTAN GELİŞİ -", color.RGBA{180, 140, 50, 200}, gameui.TextSmall, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: 0, Y: float64(by) + 54, W: ScreenWidth}, title, color.RGBA{255, 220, 80, 255}, gameui.TextLarge, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: 0, Y: float64(by) + 88, W: ScreenWidth}, desc, color.RGBA{210, 200, 180, 230}, gameui.TextMedium, gameui.TextAlignCenter)

	viewport := commanderArrivalListViewport()
	scroll = clampCommanderArrivalScroll(scroll, len(arrivals), viewport)
	visibleRows := commanderArrivalVisibleRows(viewport)
	start := scroll * 3
	end := (scroll + visibleRows) * 3
	if start > len(arrivals) {
		start = len(arrivals)
	}
	if end > len(arrivals) {
		end = len(arrivals)
	}
	left, top := int(viewport.X), int(viewport.Y)
	right, bottom := int(viewport.X+viewport.W), int(viewport.Y+viewport.H)
	if right > left && bottom > top {
		body := screen.SubImage(image.Rect(left, top, right, bottom)).(*ebiten.Image)
		for i := start; i < end; i++ {
			commander := arrivals[i]
			if commander == nil {
				continue
			}
			row := commanderArrivalCardRect(i, scroll)
			x, y := row.X, row.Y
			vector.FillRect(body, float32(x), float32(y), commanderArrivalCardW, commanderArrivalCardH, color.RGBA{35, 26, 14, 235}, false)
			vector.StrokeRect(body, float32(x), float32(y), commanderArrivalCardW, commanderArrivalCardH, 1, color.RGBA{150, 120, 68, 255}, false)
			drawCommanderPortrait(body, commander, x+8, y+8, 54, 62)
			drawUILabel(body, gameui.Rect{X: x + 70, Y: y + 12, W: commanderArrivalCardW - 78}, commander.Name, ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
			drawUILabel(body, gameui.Rect{X: x + 70, Y: y + 38, W: commanderArrivalCardW - 78}, fmt.Sprintf("Seviye %d", commander.Level), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
			drawUILabel(body, gameui.Rect{X: x + 8, Y: y + 76, W: commanderArrivalCardW - 16}, commanderArrivalTraits(commander), ColorGold, gameui.TextSmall, gameui.TextAlignStart)
			date := fmt.Sprintf("Görev: %d", commander.StartYear)
			if commander.EndYear != 0 {
				date += fmt.Sprintf("–%d", commander.EndYear)
			}
			drawUILabel(body, gameui.Rect{X: x + 70, Y: y + 62, W: commanderArrivalCardW - 78}, date, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		}
	}
	drawCommanderArrivalScrollbar(screen, viewport, len(arrivals), scroll)
	drawUILabel(screen, gameui.Rect{X: 0, Y: float64(by+bh) - 28, W: ScreenWidth}, "[Enter / Boşluk / Tıkla] Devam Et", color.RGBA{140, 130, 100, 200}, gameui.TextSmall, gameui.TextAlignCenter)
}

func commanderArrivalTraits(commander *army.Commander) string {
	if commander == nil || len(commander.Traits) == 0 {
		return "Özellik: Yok"
	}
	labels := make([]string, 0, len(commander.Traits))
	for _, trait := range commander.Traits {
		labels = append(labels, army.TraitLabelTR(trait))
	}
	return "Özellik: " + strings.Join(labels, ", ")
}

func drawHistoricalChoiceInfo(screen *ebiten.Image, btn gameui.Button, choice HistoricalEventChoice) {
	infoX := btn.X - 6
	infoW := btn.W + 12
	startY := btn.Y - 122
	if choice.Desc != "" {
		drawUIWrappedLabelAligned(screen, gameui.Rect{X: infoX, Y: startY, W: infoW}, choice.Desc, color.RGBA{162, 150, 120, 210}, gameui.TextSmall, 16, 2, gameui.TextAlignCenter)
	}
	if choice.Effect != "" {
		drawUIWrappedLabelAligned(screen, gameui.Rect{X: infoX, Y: btn.Y - 74, W: infoW}, choice.Effect, color.RGBA{190, 176, 142, 220}, gameui.TextSmall, 16, 2, gameui.TextAlignCenter)
	}
	if choice.FollowUp != "" {
		drawUIWrappedLabelAligned(screen, gameui.Rect{X: infoX, Y: btn.Y - 42, W: infoW}, choice.FollowUp, color.RGBA{232, 196, 112, 230}, gameui.TextSmall, 16, 2, gameui.TextAlignCenter)
	}
	if choice.Conditions != "" {
		drawUIWrappedLabelAligned(screen, gameui.Rect{X: infoX, Y: btn.Y - 10, W: infoW}, choice.Conditions, color.RGBA{144, 138, 126, 220}, gameui.TextSmall, 16, 2, gameui.TextAlignCenter)
	}
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
	case state.VictorySurviveTurns:
		return "Hayatta Kalma Zaferi"
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

	cardX := float32(908)
	cardY := float32(panelY) + 7
	cardW := float32(130)
	cardH := topStatusH - 14
	drawTopStatusCard(screen, cardX, cardY, cardW, cardH)

	mx := float64(cardX) + 12
	my := panelY + 16

	unitStr := itoa(deployed) + "/" + itoa(cap)
	unitCol := ColorGold
	if cap > 0 && deployed >= cap {
		unitCol = ColorRed
	}
	drawUIKeyValueRow(screen, mx, my, float64(cardW)-24, "Savaşçı", unitStr, ColorGray, unitCol)

	armyStr := itoa(armies) + "/" + itoa(maxArmies)
	armyCol := ColorGold
	if armies >= maxArmies {
		armyCol = ColorRed
	}
	drawUIKeyValueRow(screen, mx, my+28, float64(cardW)-24, "Ordu", armyStr, ColorGray, armyCol)
}

// drawVictoryProgress seçilen zafer tipine göre ilerlemeyi gösterir.
func drawVictoryProgress(screen *ebiten.Image, gs *state.GameState, panelY float64) {
	if gs.PlayerFactionID == "" {
		return
	}

	cardX := float32(718)
	cardY := float32(panelY) + 7
	cardW := float32(180)
	cardH := topStatusH - 14
	drawTopStatusCard(screen, cardX, cardY, cardW, cardH)

	vx := float64(cardX) + 12
	vy := panelY + 14
	barW := cardW - 24
	barX := cardX + 12

	titleCol := color.RGBA{220, 190, 100, 220}
	drawUISectionLabel(screen, vx, vy, "Zafer Hedefi")
	vy += 17
	if deadline := trimTextToWidth(formatVictoryDeadline(gs.Victory.DeadlineYear, gs.Victory.DeadlineMonth), FaceSmall, float64(barW)); deadline != "" {
		DrawText(screen, deadline, vx, vy, FaceSmall, color.RGBA{205, 176, 108, 220})
		vy += 14
	}

	switch gs.Victory.Type {
	case state.VictoryDomination, "":
		target := gs.Victory.TargetRegionCount
		if target == 0 {
			target = 20
		}
		current := len(gs.RegionsOwnedBy(gs.PlayerFactionID))
		drawUIKeyValueRow(screen, vx, vy, float64(barW), "Hedef", itoa(current)+"/"+itoa(target), titleCol, ColorWhite)
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
		drawUIKeyValueRow(screen, vx, vy, float64(barW), "Gelir", itoa(goldIncome)+"/"+itoa(threshold), titleCol, ColorGold)
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
		drawUIKeyValueRow(screen, vx, vy, float64(barW), "Güç", itoa(totalStr)+"/"+itoa(targetStr), titleCol, ColorWhite)
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
		drawUIKeyValueRow(screen, vx, vy, float64(barW), "Kutsal", itoa(held)+"/"+itoa(total), titleCol, color.RGBA{200, 160, 255, 255})
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
		drawUIKeyValueRow(screen, vx, vy, float64(barW), "Hedef", itoa(held)+"/"+itoa(total), titleCol, ColorWhite)
		vy += 18
		drawTopProgressBar(screen, barX, float32(vy), barW, 7, clampF(float64(held)/float64(total)), ColorGold)

	case state.VictorySurviveTurns:
		target := gs.Victory.TargetTurns
		if target == 0 {
			target = 60
		}
		current := gs.Turn
		if current > target {
			current = target
		}
		drawUIKeyValueRow(screen, vx, vy, float64(barW), "Tur", itoa(current)+"/"+itoa(target), titleCol, ColorWhite)
		vy += 18
		drawTopProgressBar(screen, barX, float32(vy), barW, 7, clampF(float64(current)/float64(target)), color.RGBA{120, 180, 255, 255})
	}
}

func victoryDetailProgressLines(gs *state.GameState) []string {
	if gs == nil {
		return nil
	}
	lines := make([]string, 0, 3)
	if deadline := formatVictoryDeadline(gs.Victory.DeadlineYear, gs.Victory.DeadlineMonth); deadline != "" {
		lines = append(lines, deadline)
	}
	switch gs.Victory.Type {
	case state.VictoryDomination, "":
		target := gs.Victory.TargetRegionCount
		if target == 0 {
			target = 20
		}
		return append(lines,
			"Kontrol edilen bölge: "+itoa(len(gs.RegionsOwnedBy(gs.PlayerFactionID)))+" / "+itoa(target),
		)
	case state.VictoryEconomic:
		threshold := gs.Victory.TargetGoldIncome
		if threshold == 0 {
			threshold = 500
		}
		holdTurns := gs.Victory.GoldHoldTurns
		if holdTurns == 0 {
			holdTurns = 5
		}
		return append(lines,
			"Mevcut gelir: "+itoa(victory.CurrentGoldIncome(gs))+" / "+itoa(threshold)+" altın",
			"Korunan süre: "+itoa(gs.EconomicVictoryTurns)+" / "+itoa(holdTurns)+" tur",
		)
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
		eliminated := 0
		for _, a := range gs.Armies {
			if a.OwnerID == string(gs.PlayerFactionID) {
				totalStr += a.TotalStrength(gs.UnitTypes)
			}
		}
		for fid, f := range gs.Factions {
			if fid != gs.PlayerFactionID && f.IsEliminated {
				eliminated++
			}
		}
		return append(lines,
			"Toplam ordu gücü: "+itoa(totalStr)+" / "+itoa(targetStr),
			"Elenen rakip: "+itoa(eliminated)+" / "+itoa(targetDef),
		)
	case state.VictoryReligious:
		held := 0
		for _, rid := range gs.Victory.RequiredRegions {
			if r, ok := gs.Regions[rid]; ok && r.OwnerID == string(gs.PlayerFactionID) {
				held++
			}
		}
		return append(lines,
			"Tutulan kutsal bölge: "+itoa(held)+" / "+itoa(len(gs.Victory.RequiredRegions)),
			"Koruma süresi: "+itoa(gs.ReligiousVictoryTurns)+" / 12 tur",
		)
	case state.VictoryConquerCity:
		held := 0
		for _, rid := range gs.Victory.RequiredRegions {
			if r, ok := gs.Regions[rid]; ok && r.OwnerID == string(gs.PlayerFactionID) {
				held++
			}
		}
		return append(lines,
			"Ele geçirilen hedef: "+itoa(held)+" / "+itoa(len(gs.Victory.RequiredRegions)),
		)
	case state.VictorySurviveTurns:
		target := gs.Victory.TargetTurns
		if target == 0 {
			target = 60
		}
		current := gs.Turn
		if current > target {
			current = target
		}
		return append(lines,
			"Dayanılan süre: "+itoa(current)+" / "+itoa(target)+" tur",
		)
	}
	return lines
}

func victoryChecklistEntries(gs *state.GameState) ([]string, []color.Color) {
	if gs == nil {
		return nil, nil
	}

	lines := make([]string, 0, 8)
	colors := make([]color.Color, 0, 8)
	ownerSuffix := func(rid world.RegionID) string {
		if gs == nil {
			return ""
		}
		region, ok := gs.Regions[rid]
		if !ok || region == nil || region.OwnerID == "" || region.OwnerID == string(gs.PlayerFactionID) {
			return ""
		}
		if f := gs.Factions[faction.FactionID(region.OwnerID)]; f != nil {
			if f.NameTR != "" {
				return " (" + f.NameTR + ")"
			}
			if f.Name != "" {
				return " (" + f.Name + ")"
			}
		}
		return " (" + region.OwnerID + ")"
	}
	addItem := func(ok bool, text string) {
		prefix := "✗ "
		col := color.Color(color.RGBA{214, 112, 92, 235})
		if ok {
			prefix = "✓ "
			col = color.RGBA{132, 198, 124, 235}
		}
		lines = append(lines, prefix+text)
		colors = append(colors, col)
	}

	switch gs.Victory.Type {
	case state.VictoryConquerCity, state.VictoryReligious:
		for _, rid := range gs.Victory.RequiredRegions {
			name := regionDisplayName(gs, string(rid))
			owned := false
			if region, ok := gs.Regions[rid]; ok && region.OwnerID == string(gs.PlayerFactionID) {
				owned = true
			}
			if !owned {
				name += ownerSuffix(rid)
			}
			addItem(owned, name)
		}
	case state.VictoryDomination:
		target := gs.Victory.TargetRegionCount
		if target == 0 {
			target = 20
		}
		current := len(gs.RegionsOwnedBy(gs.PlayerFactionID))
		addItem(current >= target, "Bölge sayısı: "+itoa(current)+"/"+itoa(target))
		for _, rid := range gs.Victory.RequiredRegions {
			name := regionDisplayName(gs, string(rid))
			owned := false
			if region, ok := gs.Regions[rid]; ok && region.OwnerID == string(gs.PlayerFactionID) {
				owned = true
			}
			if !owned {
				name += ownerSuffix(rid)
			}
			addItem(owned, name)
		}
	case state.VictoryEconomic:
		threshold := gs.Victory.TargetGoldIncome
		if threshold == 0 {
			threshold = 500
		}
		holdTurns := gs.Victory.GoldHoldTurns
		if holdTurns == 0 {
			holdTurns = 5
		}
		currentIncome := victory.CurrentGoldIncome(gs)
		addItem(currentIncome >= threshold, "Gelir eşiği: "+itoa(currentIncome)+"/"+itoa(threshold)+" altın")
		addItem(gs.EconomicVictoryTurns >= holdTurns, "Koruma süresi: "+itoa(gs.EconomicVictoryTurns)+"/"+itoa(holdTurns)+" tur")
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
		eliminated := 0
		for _, a := range gs.Armies {
			if a.OwnerID == string(gs.PlayerFactionID) {
				totalStr += a.TotalStrength(gs.UnitTypes)
			}
		}
		for fid, f := range gs.Factions {
			if fid != gs.PlayerFactionID && f.IsEliminated {
				eliminated++
			}
		}
		addItem(totalStr >= targetStr, "Ordu gücü: "+itoa(totalStr)+"/"+itoa(targetStr))
		addItem(eliminated >= targetDef, "Elenen rakip: "+itoa(eliminated)+"/"+itoa(targetDef))
	case state.VictorySurviveTurns:
		target := gs.Victory.TargetTurns
		if target == 0 {
			target = 60
		}
		addItem(gs.Turn >= target, "Hayatta kalma süresi: "+itoa(min(gs.Turn, target))+"/"+itoa(target)+" tur")
	}

	return lines, colors
}

func drawTopStatusCard(screen *ebiten.Image, x, y, w, h float32) {
	drawUIPanelFrame(screen, gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}, color.RGBA{18, 16, 12, 150}, color.RGBA{95, 78, 42, 115}, 1, 1)
}

func drawTopProgressBar(screen *ebiten.Image, x, y, w, h float32, fill float64, col color.Color) {
	drawUIProgressBar(screen, x, y, w, h, fill, color.RGBA{42, 42, 40, 210}, color.RGBA{120, 100, 55, 150}, col, 1)
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
	drawUICardRect(screen, gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}, color.RGBA{42, 34, 16, 220}, color.RGBA{190, 150, 70, 230}, 1)
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

func drawResRow(screen *ebiten.Image, x, y, w float64, label, value string, col color.RGBA) {
	drawUIKeyValueRowWithGap(screen, x, y, w, label, value, ColorGray, col, 8)
}

func topResourceHUDColumns() (leftCol1, leftCol2, rightCol, leftColW, rightColW float64) {
	const resStartX = 300.0
	const colGap = 10.0
	rightColW = 118.0
	resEndX := victoryProgressPanelRect().X - 12
	// Gelir/Altın sütunu sağa yaslanır. Dar kaynak sütunlarının sonuna
	// sıkıştırılmadığı için formatlı değerler Kereste/Taş değerlerine
	// taşmaz; zafer kartından önce de sabit bir boşluk korunur.
	leftColW = (resEndX - resStartX - rightColW - colGap*2) / 2
	if leftColW < 88 {
		leftColW = 88
	}
	leftCol1 = resStartX
	leftCol2 = leftCol1 + leftColW + colGap
	rightCol = resEndX - rightColW
	return
}

func formatResourceHUDValue(current, change int) string {
	return formatSignedAmount(change) + " / " + formatNumberTR(current)
}

func formatSignedAmount(amount int) string {
	if amount > 0 {
		return "+" + formatNumberTR(amount)
	}
	return formatNumberTR(amount)
}

// formatNumberTR sayıları Türkçe binlik ayıracıyla gösterir (10.000 gibi).
// HUD değerleri için kendi tamponunu kullandığından fmt.Sprintf'e göre daha
// az geçici veri üretir.
func formatNumberTR(n int) string {
	value := int64(n)
	negative := value < 0
	var magnitude uint64
	if negative {
		magnitude = uint64(-(value + 1)) + 1
	} else {
		magnitude = uint64(value)
	}
	if magnitude == 0 {
		return "0"
	}

	var buf [32]byte
	pos := len(buf)
	digits := 0
	for magnitude > 0 {
		if digits == 3 {
			pos--
			buf[pos] = '.'
			digits = 0
		}
		pos--
		buf[pos] = byte('0' + magnitude%10)
		magnitude /= 10
		digits++
	}
	if negative {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

func resourceHUDChangeColor(change int, normal color.RGBA) color.RGBA {
	if change < 0 {
		return ColorRed
	}
	return normal
}

func calcPlayerIncome(gs *state.GameState) int {
	return victory.CurrentGoldIncome(gs)
}

// drawBuildingGrid bölgedeki binaları kare sprite thumbnail'leri olarak 3 sütunlu ızgarada çizer.
// İnşa edilmiş binalar renkli sprite ile, boş slotlar soluk çerçeve ile gösterilir.
func drawBuildingGrid(screen *ebiten.Image, gs *state.GameState, region *world.Region, panelX, startY, panelW float32) {
	cards := buildBuildingCardComponents(gs, region, panelX, startY, panelW)
	cacheBuildingGridComponents(region.ID, cards)
	for _, card := range cards {
		card.Draw(screen)
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
	btn := buildPanelCloseButton(px, py, pw)
	drawTinyPanelButtonWidget(screen, btn, true)
}

func panelCloseRect(px, py, pw float32) (x, y, w, h float32) {
	return px + pw - 24, py + 6, 18, 18
}

func drawTinyPanelButton(screen *ebiten.Image, x, y, w, h float32, label string, active bool) {
	drawUIButton(screen, float64(x), float64(y), float64(w), float64(h), label, active, tinyButtonStyle)
}

func drawTinyPanelButtonWidget(screen *ebiten.Image, btn gameui.Button, active bool) {
	btn.Enabled = active
	drawUIButtonWidget(screen, btn, tinyButtonStyle)
}

func panelCloseHit(mx, my float64, px, py, pw float32) bool {
	return buildPanelCloseButton(px, py, pw).HitTest(mx, my)
}

func buildPanelCloseButton(px, py, pw float32) gameui.Button {
	x, y, w, h := panelCloseRect(px, py, pw)
	btn := gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "").WithIcon(gameui.IconClose)
	btn.IconSize = 12
	return btn
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

func settlementPanelHit(mx, my float64) bool {
	px := settlementPanelX()
	py := settlementPanelY()
	return mx >= float64(px) && mx <= float64(px+infoPanelW) && my >= float64(py) && my <= float64(py+infoPanelH)
}

func settlementPanelCloseHit(mx, my float64) bool {
	return panelCloseHit(mx, my, settlementPanelX(), settlementPanelY(), infoPanelW)
}

func factionPanelHit(mx, my float64) bool {
	px := factionPanelX()
	py := factionPanelY()
	return mx >= float64(px) && mx <= float64(px+infoPanelW) && my >= float64(py) && my <= float64(py+infoPanelH)
}

func factionPanelCloseHit(mx, my float64) bool {
	return panelCloseHit(mx, my, factionPanelX(), factionPanelY(), infoPanelW)
}

func regionOwnerNameHit(mx, my float64, gs *state.GameState, rid world.RegionID) (faction.FactionID, bool) {
	rect, fid, ok := regionOwnerNameRect(gs, rid)
	if !ok {
		return "", false
	}
	return fid, mx >= rect.X && mx <= rect.X+rect.W && my >= rect.Y && my <= rect.Y+rect.H
}

func regionOwnerNameRect(gs *state.GameState, rid world.RegionID) (gameui.Rect, faction.FactionID, bool) {
	if gs == nil || rid == "" {
		return gameui.Rect{}, "", false
	}
	region, ok := gs.Regions[rid]
	if !ok || region == nil || region.IsSea || region.OwnerID == "" {
		return gameui.Rect{}, "", false
	}
	ownerName, _ := ownerDisplay(gs, region.OwnerID)
	width := MeasureText(ownerName, FaceLarge)
	maxWidth := float64(infoPanelW - float32(panelPad*2))
	if width > maxWidth {
		width = maxWidth
	}
	return gameui.Rect{
		X: float64(infoPanelX()) + panelPad,
		Y: float64(infoPanelY()) + 34,
		W: width,
		H: float64(regionOwnerNameH),
	}, faction.FactionID(region.OwnerID), true
}

func regionPanelInteractiveHit(mx, my float64, gs *state.GameState, rid world.RegionID) bool {
	return regionPanelInteractiveHitForTab(mx, my, gs, rid, regionPanelTabBuildings, 0)
}

func regionPanelInteractiveHitForTab(mx, my float64, gs *state.GameState, rid world.RegionID, activeTab regionPanelTab, scroll float64) bool {
	if rid == "" {
		return false
	}
	if _, ok := regionPanelTabHit(mx, my, gs, rid); ok {
		return true
	}
	if regionPanelCloseHit(mx, my) {
		return true
	}
	if _, ok := regionOwnerNameHit(mx, my, gs, rid); ok {
		return true
	}
	if delta := regionTaxButtonHit(mx, my, gs, rid); delta != 0 {
		return true
	}
	if regionGrainAidButtonHitForTab(mx, my, gs, rid, activeTab) {
		return true
	}
	if regionDiplomacyButtonHitForTab(mx, my, gs, rid, activeTab) {
		return true
	}
	if regionNeighborToggleHit(mx, my, gs, rid) {
		return true
	}
	if activeTab == regionPanelTabEvents {
		region := gs.Regions[rid]
		if regionActivityNeighborToggleHit(mx, my, gs, region, scroll) {
			return true
		}
		_, ok := regionActiveEventPanelHit(mx, my, gs, region, scroll)
		return ok
	}
	if _, ok := BuildingGridDemolishHitTest(mx, my, gs, rid, false); ok {
		return true
	}
	return buildingGridHitTest(mx, my, gs, rid, false) != ""
}

func DrawSettlementPanel(screen *ebiten.Image, gs *state.GameState, region *world.Region, settlement *world.Settlement) {
	if gs == nil || region == nil || settlement == nil {
		return
	}

	px := settlementPanelX()
	py := settlementPanelY()
	pw := infoPanelW
	ph := infoPanelH

	drawUIPanelFrame(screen, gameui.Rect{X: float64(px), Y: float64(py), W: float64(pw), H: float64(ph)}, panelBg, panelBorder, 1.5, 3)
	drawPanelCloseButton(screen, px, py, pw)

	lx := float64(px) + panelPad
	ly := float64(py) + 10

	name := settlement.NameTR
	if name == "" {
		name = settlement.Name
	}
	if name == "" {
		name = "Yerleşim"
	}
	DrawText(screen, name, lx, ly, FaceLarge, ColorYellow)
	ly += 24

	if gs.DevelopmentMode {
		drawUIKeyValueRow(screen, lx, ly, float64(pw)-panelPad*2, "ID", string(settlement.ID), ColorGray, ColorWhite)
		ly += 18
	}

	drawUIKeyValueRow(screen, lx, ly, float64(pw)-panelPad*2, "Bölge", region.NameTR, ColorGray, ColorWhite)
	ly += 18
	drawUIKeyValueRow(screen, lx, ly, float64(pw)-panelPad*2, "Tip", settlement.Type.LabelTR(), ColorGray, ColorWhite)
	ly += 18
	drawUIKeyValueRow(screen, lx, ly, float64(pw)-panelPad*2, "Nüfus", itoa(settlement.Population), ColorGray, ColorWhite)
	ly += 18
	drawUIKeyValueRow(screen, lx, ly, float64(pw)-panelPad*2, "Koordinat", itoa(settlement.X)+","+itoa(settlement.Y), ColorGray, ColorWhite)
	ly += 20

	// Üst görsel alanı (asset varsa göster, yoksa placeholder).
	imgX := float32(lx)
	imgY := float32(ly)
	imgW := pw - float32(panelPad*2)
	imgH := float32(170)
	drawUICardRect(screen, gameui.Rect{X: float64(imgX), Y: float64(imgY), W: float64(imgW), H: float64(imgH)}, panelBg2, panelBorder, 1)
	if sImg := loadSettlementImage(region, settlement); sImg != nil {
		b := sImg.Bounds()
		sw := float64(b.Dx())
		sh := float64(b.Dy())
		if sw > 0 && sh > 0 {
			scale := minFloat64(float64(imgW)/sw, float64(imgH)/sh)
			dw := sw * scale
			dh := sh * scale
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(imgX)+(float64(imgW)-dw)/2, float64(imgY)+(float64(imgH)-dh)/2)
			screen.DrawImage(sImg, op)
		}
	} else {
		phText := "Yerleşim Görseli"
		tw := MeasureText(phText, FaceMed)
		DrawText(screen, phText, float64(imgX)+float64(imgW)/2-tw/2, float64(imgY)+72, FaceMed, ColorGray)
	}

	ly += float64(imgH) + 16
	drawUISeparator(screen, float32(lx), float32(ly), float32(lx)+imgW, 1, panelBorder)
	ly += 10
	drawUISectionLabel(screen, lx, ly, "Tarihçe")
	ly += 18
	DrawText(screen, "Bu alan daha sonra metinsel içerikle doldurulacak.", lx, ly, FaceSmall, ColorGray)
	ly += 32

	if region.OwnerID == string(gs.PlayerFactionID) {
		drawUISeparator(screen, float32(lx), float32(ly), float32(lx)+imgW, 1, panelBorder)
		ly += 10
		drawUISectionLabel(screen, lx, ly, "Başkent")
		ly += 18
		status := settlementCapitalStatusText(gs, region, settlement)
		drawUIWrappedLabel(screen, gameui.Rect{X: lx, Y: ly, W: float64(imgW)}, status, ColorGray, gameui.TextSmall, 16, 3)
		ly += 52
		if btn, ok := settlementCapitalActionButton(gs, region, settlement); ok {
			drawUIButtonWidget(screen, btn, solidButtonStyle(color.RGBA{126, 94, 28, 235}, color.RGBA{214, 176, 82, 255}, ColorWhite, 10))
		}
	}
}

func settlementCapitalActionButton(gs *state.GameState, region *world.Region, settlement *world.Settlement) (gameui.Button, bool) {
	if gs == nil || region == nil || settlement == nil || region.OwnerID != string(gs.PlayerFactionID) {
		return gameui.Button{}, false
	}
	player := gs.Factions[gs.PlayerFactionID]
	if player == nil {
		return gameui.Button{}, false
	}
	if gs.IsFactionCapitalSettlement(gs.PlayerFactionID, settlement.ID) {
		return gameui.Button{}, false
	}
	if player.PendingCapitalSettlementID == settlement.ID && player.PendingCapitalTurns > 0 {
		return gameui.Button{}, false
	}
	label := "Başkent Yap"
	if player.PendingCapitalSettlementID != "" && player.PendingCapitalTurns > 0 {
		label = "Taşımayı Buraya Çevir"
	}
	return gameui.NewButton(
		float64(settlementPanelX())+panelPad,
		float64(settlementPanelY()+infoPanelH)-56,
		float64(infoPanelW)-panelPad*2,
		36,
		label,
	).WithIcon(gameui.IconCheck), true
}

func settlementCapitalStatusText(gs *state.GameState, region *world.Region, settlement *world.Settlement) string {
	if gs == nil || region == nil || settlement == nil || region.OwnerID == "" {
		return "Sahipsiz yerleşimlerde başkent statüsü yok."
	}
	fid := faction.FactionID(region.OwnerID)
	f := gs.Factions[fid]
	if f == nil {
		return "Başkent bilgisi okunamadı."
	}
	bonusSummary := fmt.Sprintf(
		"Bölge bonusu: +%d altın, +%d tahıl, +%d demir, +%d kereste, +%d taş, +%d baharat, +%d kumaş, +%d lojistik.",
		state.CapitalRegionGoldBonus,
		state.CapitalRegionGrainBonus,
		state.CapitalRegionIronBonus,
		state.CapitalRegionTimberBonus,
		state.CapitalRegionStoneBonus,
		state.CapitalRegionSpiceBonus,
		state.CapitalRegionClothBonus,
		state.CapitalRegionLogisticsBonus,
	)
	captureRisk := " Başkent düşerse depoların yarısı ve savunanın bilmediğin teknolojilerinin yaklaşık yarısı fetheden devlete geçer."
	if gs.IsFactionCapitalSettlement(fid, settlement.ID) {
		return "Bu yerleşim devletin mevcut başkentidir. " + bonusSummary + captureRisk
	}
	if f.PendingCapitalSettlementID == settlement.ID && f.PendingCapitalTurns > 0 {
		return "Bu yerleşim başkent taşımasının aktif hedefidir. Kalan süre: " + itoa(f.PendingCapitalTurns) + " tur. " + bonusSummary
	}
	if f.PendingCapitalSettlementID != "" && f.PendingCapitalTurns > 0 {
		targetName := f.PendingCapitalSettlementID
		if targetRegion, targetSettlement, _, ok := gs.FindSettlementByID(f.PendingCapitalSettlementID); ok {
			if targetSettlement.NameTR != "" {
				targetName = targetSettlement.NameTR
			} else if targetRegion != nil && targetRegion.NameTR != "" {
				targetName = targetRegion.NameTR
			}
		}
		return "Aktif başkent taşıması sürüyor: " + targetName + " (" + itoa(f.PendingCapitalTurns) + " tur). " + bonusSummary
	}
	return "Bu yerleşim başkent değil. İstersen 5 tur süren bir taşıma süreci başlatabilirsin. " + bonusSummary
}

type factionTradeOverview struct {
	RouteCount     int
	SuspendedCount int
	PartnerCount   int
	ExportGold     int
	ImportGold     int
	NetGold        int
}

func DrawFactionDetailPanel(screen *ebiten.Image, gs *state.GameState, fid faction.FactionID, scroll float64) {
	if gs == nil || fid == "" {
		return
	}
	f := gs.Factions[fid]
	if f == nil {
		return
	}

	px := factionPanelX()
	py := factionPanelY()
	pw := infoPanelW
	ph := infoPanelH
	drawUIPanelFrame(screen, gameui.Rect{X: float64(px), Y: float64(py), W: float64(pw), H: float64(ph)}, panelBg, panelBorder, 1.5, 3)
	drawPanelCloseButton(screen, px, py, pw)

	lx := float64(px) + panelPad
	ly := float64(py) + 10
	sepW := float64(pw - float32(panelPad*2))

	name := f.NameTR
	if strings.TrimSpace(name) == "" {
		name = f.Name
	}
	if strings.TrimSpace(name) == "" {
		name = string(fid)
	}
	factionColor := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
	drawFactionFlagBadge(screen, fid, string([]rune(name)[:1]), lx, ly-2, factionPanelFlagSize, factionColor, panelBorder)
	nameX := lx + factionPanelFlagSize + 10
	DrawText(screen, name, nameX, ly, FaceLarge, factionNameColor)
	ly += 24

	drawUIWrappedLabel(screen, gameui.Rect{X: nameX, Y: ly, W: sepW - (nameX - lx)}, factionPanelSubtitle(gs, fid, f), ColorGray, gameui.TextSmall, 16, 2)
	ly += 28

	drawUISeparator(screen, float32(lx), float32(ly), float32(lx)+float32(sepW), 1, panelBorder)
	ly += 8

	summary := buildFactionDiplomacySummary(gs, fid)
	bodyY := ly
	bodyH := float64(ph) - (bodyY - float64(py)) - panelPad
	if bodyH < 1 {
		bodyH = 1
	}
	bodyCanvas := factionPanelBodyCanvas()
	bodyW := float64(bodyCanvas.Bounds().Dx())
	contentHeight := factionPanelContentHeight(gs, fid, f, summary, bodyW)
	scroll = clampFactionPanelScroll(contentHeight, bodyH, scroll)

	bodyCanvas.Fill(color.RGBA{0, 0, 0, 0})
	drawFactionDetailBody(bodyCanvas, gs, fid, f, summary, bodyW, scroll)
	bodyOpts := &ebiten.DrawImageOptions{}
	bodyOpts.GeoM.Translate(lx, bodyY)
	screen.DrawImage(bodyCanvas, bodyOpts)
	drawFactionPanelScrollbar(screen, float64(px)+float64(pw)-8, bodyY, bodyH, contentHeight, scroll)
}

func factionPanelSubtitle(gs *state.GameState, fid faction.FactionID, f *faction.Faction) string {
	parts := []string{religion.DisplayNameTR(f.Religion)}
	if fid == gs.PlayerFactionID {
		parts = append(parts, "Siz")
	} else if f.IsEliminated {
		parts = append(parts, "Yıkıldı")
	} else {
		parts = append(parts, "Rakip")
	}
	if rel := factionRelationToPlayer(gs, fid); rel != nil && fid != gs.PlayerFactionID {
		parts = append(parts, "İlişki: "+faction.DiplomaticStanceLabelTR(rel.Stance)+" ("+itoa(rel.Score)+")")
	}
	return strings.Join(parts, "  •  ")
}

func factionRelationToPlayer(gs *state.GameState, fid faction.FactionID) *faction.Relation {
	if gs == nil || fid == "" || gs.PlayerFactionID == "" || fid == gs.PlayerFactionID {
		return nil
	}
	return gs.Relations[faction.RelationKey(gs.PlayerFactionID, fid)]
}

func factionNavalArmyCount(gs *state.GameState, fid faction.FactionID) int {
	count := 0
	for _, a := range gs.Armies {
		if a != nil && a.OwnerID == string(fid) && a.IsNaval {
			count++
		}
	}
	return count
}

func factionMilitaryPowerBreakdownLabel(gs *state.GameState, fid faction.FactionID) string {
	land, naval := diplomacy.MilitaryPowerBreakdown(gs, fid)
	return itoa(land) + " / " + itoa(naval)
}

func factionActiveResearchLabel(gs *state.GameState, f *faction.Faction) string {
	if f == nil || f.Research.ActiveID == "" {
		return "Aktif araştırma yok"
	}
	if gs != nil && gs.TechTypes != nil {
		if t := gs.TechTypes[f.Research.ActiveID]; t != nil {
			return t.NameTR + " • " + itoa(f.Research.TurnsLeft) + " tur"
		}
	}
	return f.Research.ActiveID + " • " + itoa(f.Research.TurnsLeft) + " tur"
}

func drawFactionResourceGrid(screen *ebiten.Image, gs *state.GameState, fid faction.FactionID, f *faction.Faction, x, y float64, width float32) {
	type resourceItem struct {
		label string
		value string
		col   color.Color
	}

	items := [...]resourceItem{
		{label: "Altın", value: itoa(f.Gold), col: ColorGold},
		{label: "Tahıl", value: itoa(f.Grain), col: ColorWhite},
		{label: "Demir", value: itoa(f.Iron), col: color.RGBA{200, 205, 215, 255}},
		{label: "Kereste", value: itoa(f.Timber), col: color.RGBA{145, 205, 145, 255}},
		{label: "Taş", value: itoa(f.Stone), col: color.RGBA{185, 185, 185, 255}},
		{label: "Baharat", value: itoa(f.Spice), col: color.RGBA{230, 165, 90, 255}},
		{label: "Kumaş", value: itoa(f.Cloth), col: color.RGBA{175, 150, 220, 255}},
		{label: "Gelir", value: "+" + itoa(victory.GoldIncomeForFaction(gs, fid)) + "/tur", col: ColorGold},
	}

	colGap := 14.0
	colW := (float64(width) - colGap) / 2
	for i, item := range items {
		col := float64(i % 2)
		row := float64(i / 2)
		drawUIKeyValueRow(screen, x+col*(colW+colGap), y+row*regionPanelStatRowGap, colW, item.label, item.value, ColorGray, item.col)
	}
}

func factionTradeStats(gs *state.GameState, fid faction.FactionID) factionTradeOverview {
	if gs == nil || fid == "" {
		return factionTradeOverview{}
	}
	self := string(fid)
	stats := factionTradeOverview{}
	partners := make(map[string]struct{})
	for _, route := range gs.TradeRoutes {
		if route == nil {
			continue
		}
		if route.FromFactionID != self && route.ToFactionID != self {
			continue
		}
		if route.SuspendedTurns > 0 {
			stats.SuspendedCount++
			continue
		}
		stats.RouteCount++
		if route.FromFactionID == self {
			stats.ExportGold += route.GoldEarned()
			if route.ToFactionID != "" {
				partners[route.ToFactionID] = struct{}{}
			}
		}
		if route.ToFactionID == self {
			stats.ImportGold += route.GoldEarned()
		}
		if route.ToFactionID == self && route.FromFactionID != "" {
			partners[route.FromFactionID] = struct{}{}
		}
	}
	stats.PartnerCount = len(partners)
	stats.NetGold = stats.ExportGold - stats.ImportGold
	return stats
}

func drawFactionCompletedTechList(screen *ebiten.Image, gs *state.GameState, f *faction.Faction, x, y, width float64) {
	names, hidden := factionCompletedTechPreview(gs, f, 12)
	if len(names) == 0 {
		drawUILabel(screen, gameui.Rect{X: x, Y: y, W: width}, "Henüz tamamlanmış teknoloji yok", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		return
	}

	colGap := 12.0
	colW := (width - colGap) / 2
	for i, name := range names {
		col := float64(i % 2)
		row := float64(i / 2)
		drawUILabel(screen, gameui.Rect{X: x + col*(colW+colGap), Y: y + row*16, W: colW}, "• "+name, ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
	}
	if hidden > 0 {
		rows := (len(names) + 1) / 2
		drawUILabel(screen, gameui.Rect{X: x, Y: y + float64(rows)*16 + 4, W: width}, "+"+itoa(hidden)+" teknoloji daha", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	}
}

func drawFactionPanelScrollbar(screen *ebiten.Image, x, y, viewportHeight, contentHeight, scroll float64) {
	if viewportHeight <= 0 || contentHeight <= viewportHeight {
		return
	}
	maxScroll := contentHeight - viewportHeight
	if maxScroll <= 0 {
		return
	}
	track := gameui.Rect{X: x, Y: y, W: factionPanelScrollBarW, H: viewportHeight}
	drawUICardRect(screen, track, color.RGBA{22, 20, 16, 210}, color.RGBA{72, 62, 42, 180}, 1)
	thumbH := track.H * (viewportHeight / contentHeight)
	if thumbH < 24 {
		thumbH = 24
	}
	if thumbH > track.H {
		thumbH = track.H
	}
	thumbY := track.Y
	if track.H > thumbH {
		thumbY += (track.H - thumbH) * (scroll / maxScroll)
	}
	drawUICardRect(screen, gameui.Rect{X: track.X, Y: thumbY, W: track.W, H: thumbH}, color.RGBA{176, 144, 78, 230}, color.RGBA{214, 190, 120, 210}, 1)
}

func factionCompletedTechPreview(gs *state.GameState, f *faction.Faction, maxItems int) ([]string, int) {
	if f == nil || len(f.Research.Completed) == 0 || maxItems <= 0 {
		return nil, 0
	}
	type techLabel struct {
		name string
		cat  int
	}
	items := make([]techLabel, 0, len(f.Research.Completed))
	for id := range f.Research.Completed {
		name := id
		cat := 99
		if gs != nil && gs.TechTypes != nil {
			if t := gs.TechTypes[id]; t != nil {
				name = t.NameTR
				cat = tech.CategoryOrder(t.Category)
			}
		}
		items = append(items, techLabel{name: name, cat: cat})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].cat != items[j].cat {
			return items[i].cat < items[j].cat
		}
		return items[i].name < items[j].name
	})
	names := make([]string, 0, minFactionInt(maxItems, len(items)))
	for idx, item := range items {
		if idx >= maxItems {
			break
		}
		names = append(names, item.name)
	}
	return names, len(items) - len(names)
}

func minFactionInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type factionDiplomacyEntry struct {
	ID    faction.FactionID
	Name  string
	Score int
}

type factionDiplomacySummary struct {
	Overlord      factionDiplomacyEntry
	HasOverlord   bool
	Vassals       []factionDiplomacyEntry
	Allies        []factionDiplomacyEntry
	Trade         []factionDiplomacyEntry
	Enemies       []factionDiplomacyEntry
	VassalCount   int
	AllianceCount int
	TradeCount    int
	EnemyCount    int
}

func factionPanelBodyCanvasSize() (int, int) {
	w := int(float64(infoPanelW) - panelPad*2 - 12)
	h := int(float64(infoPanelH) - factionPanelHeaderH - factionPanelBodyPadBottom)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}

func factionPanelBodyWidth() float64 {
	w, _ := factionPanelBodyCanvasSize()
	return float64(w)
}

func factionPanelBodyCanvas() *ebiten.Image {
	return factionPanelBodyBuffer
}

func factionPanelBodyHeight() float64 {
	_, h := factionPanelBodyCanvasSize()
	return float64(h)
}

func clampFactionPanelScroll(contentHeight, viewportHeight, scroll float64) float64 {
	if scroll < 0 {
		return 0
	}
	maxScroll := contentHeight - viewportHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func factionDiplomacyEntryColor(kind faction.DiplomaticStance) color.RGBA {
	switch kind {
	case faction.StanceWar:
		return color.RGBA{196, 72, 72, 255}
	case faction.StanceAllied:
		return color.RGBA{86, 164, 94, 255}
	case faction.StanceTrade:
		return color.RGBA{200, 158, 74, 255}
	default:
		return color.RGBA{182, 172, 154, 255}
	}
}

func factionDiplomacyEntryLabel(entry factionDiplomacyEntry, suffix string, showScore bool) string {
	label := entry.Name
	if suffix != "" {
		label += " • " + suffix
	}
	if showScore {
		label += " (" + fmt.Sprintf("%+d", entry.Score) + ")"
	}
	return label
}

func sortFactionDiplomacyEntries(entries []factionDiplomacyEntry, descending bool) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Score != entries[j].Score {
			if descending {
				return entries[i].Score > entries[j].Score
			}
			return entries[i].Score < entries[j].Score
		}
		return entries[i].Name < entries[j].Name
	})
}

func buildFactionDiplomacySummary(gs *state.GameState, fid faction.FactionID) factionDiplomacySummary {
	summary := factionDiplomacySummary{}
	if gs == nil || fid == "" {
		return summary
	}

	if overlordID := diplomacy.DirectOverlord(gs, fid); overlordID != "" {
		if f := gs.Factions[overlordID]; f != nil {
			summary.HasOverlord = true
			summary.Overlord = factionDiplomacyEntry{
				ID:    overlordID,
				Name:  factionDisplayName(gs, string(overlordID)),
				Score: 0,
			}
		}
	}

	for otherID, other := range gs.Factions {
		if other == nil || other.IsEliminated || otherID == fid {
			continue
		}
		if other.OverlordID == fid {
			summary.Vassals = append(summary.Vassals, factionDiplomacyEntry{
				ID:    otherID,
				Name:  factionDisplayName(gs, string(otherID)),
				Score: 0,
			})
			continue
		}
		if diplomacy.SameRealm(gs, fid, otherID) {
			continue
		}
		rel := diplomacy.Relation(gs, fid, otherID)
		if rel == nil {
			continue
		}
		entry := factionDiplomacyEntry{ID: otherID, Name: factionDisplayName(gs, string(otherID)), Score: rel.Score}
		switch rel.Stance {
		case faction.StanceWar:
			summary.Enemies = append(summary.Enemies, entry)
		case faction.StanceAllied:
			summary.Allies = append(summary.Allies, entry)
		case faction.StanceTrade:
			summary.Trade = append(summary.Trade, entry)
		}
	}

	sortFactionDiplomacyEntries(summary.Vassals, false)
	sortFactionDiplomacyEntries(summary.Allies, true)
	sortFactionDiplomacyEntries(summary.Trade, true)
	sortFactionDiplomacyEntries(summary.Enemies, false)
	summary.VassalCount = len(summary.Vassals)
	summary.AllianceCount = len(summary.Allies)
	summary.TradeCount = len(summary.Trade)
	summary.EnemyCount = len(summary.Enemies)
	return summary
}

func factionPanelContentHeight(gs *state.GameState, fid faction.FactionID, f *faction.Faction, summary factionDiplomacySummary, width float64) float64 {
	y := 0.0
	y += factionPanelSectionH
	y += 24
	y += factionPanelRowH * 4
	y += 22

	y += factionPanelSectionH
	y += factionPanelRowH * 2
	if paused := len(f.Research.PausedTurns); paused > 0 {
		y += factionPanelRowH
	}
	y += factionPanelTechSummaryH

	y += factionPanelSectionH
	y += regionPanelStatRowGap * 4

	y += factionPanelSectionH
	y += factionPanelRowH * 3
	if trade := factionTradeStats(gs, fid); trade.SuspendedCount > 0 {
		y += factionPanelRowH
	}
	if rel := factionRelationToPlayer(gs, fid); rel != nil && fid != gs.PlayerFactionID {
		y += 22
	} else {
		y += 4
	}

	y += factionPanelSectionH
	y += factionPanelRowH * 5

	if summary.VassalCount > 0 {
		y += factionPanelSectionH
		y += float64(summary.VassalCount) * factionPanelRowH
	}
	if summary.AllianceCount > 0 {
		y += factionPanelSectionH
		y += float64(summary.AllianceCount) * factionPanelRowH
	}
	if summary.TradeCount > 0 {
		y += factionPanelSectionH
		y += float64(summary.TradeCount) * factionPanelRowH
	}
	if summary.EnemyCount > 0 {
		y += factionPanelSectionH
		y += float64(summary.EnemyCount) * factionPanelRowH
	}

	y += factionPanelSectionH
	y += factionCompletedTechContentHeight(gs, f)
	y += factionAIDebugContentHeight(gs, fid, width)
	return y
}

func drawFactionDetailBody(screen *ebiten.Image, gs *state.GameState, fid faction.FactionID, f *faction.Faction, summary factionDiplomacySummary, width float64, scroll float64) {
	y := -scroll

	drawUISectionLabel(screen, 0, y, "Durum")
	y += factionPanelSectionH
	_, militaryRank, _ := factionMilitaryPowerStanding(gs, fid)
	rankValue := "Yok"
	if militaryRank > 0 {
		rankValue = itoa(militaryRank)
	}
	rankRow := gameui.NewKeyValueRow(gameui.Rect{X: 0, Y: y, W: width}, "Güç Sırası", rankValue)
	rankRow.LabelColor = ColorGold
	rankRow.ValueColor = ColorGold
	rankRow.LabelVariant = gameui.TextMedium
	rankRow.ValueVariant = gameui.TextEmphasized
	drawUIKeyValueWidget(screen, rankRow)
	y += 24
	drawUIKeyValueRow(screen, 0, y, width, "Kara / Deniz Gücü", factionMilitaryPowerBreakdownLabel(gs, fid), ColorGray, ColorGold)
	y += factionPanelRowH
	drawUIKeyValueRow(screen, 0, y, width, "Bölgeler", itoa(len(gs.LandRegionsOwnedBy(fid))), ColorGray, ColorWhite)
	y += factionPanelRowH
	drawUIKeyValueRow(screen, 0, y, width, "Ordu", itoa(gs.CurrentLandArmies(fid))+" / "+itoa(gs.DeployedLandUnits(fid))+" birim", ColorGray, ColorWhite)
	y += factionPanelRowH
	drawUIKeyValueRow(screen, 0, y, width, "Donanma", itoa(factionNavalArmyCount(gs, fid))+" Filo", ColorGray, ColorWhite)
	y += 22

	drawUISectionLabel(screen, 0, y, "Araştırma")
	y += factionPanelSectionH
	drawUIKeyValueRow(screen, 0, y, width, "Aktif", factionActiveResearchLabel(gs, f), ColorGray, ColorWhite)
	y += factionPanelRowH
	drawUIKeyValueRow(screen, 0, y, width, "Tamamlanan", itoa(len(f.Research.Completed))+" teknoloji", ColorGray, ColorWhite)
	y += factionPanelRowH
	if paused := len(f.Research.PausedTurns); paused > 0 {
		drawUIKeyValueRow(screen, 0, y, width, "Beklemede", itoa(paused)+" araştırma", ColorGray, ColorWhite)
		y += factionPanelRowH
	}
	buffSummary := techEffectsSummary(tech.ComputeEffects(f.Research.Completed, gs.TechTypes), "Belirgin bonus yok")
	drawUIWrappedLabel(screen, gameui.Rect{X: 0, Y: y, W: width}, buffSummary, color.RGBA{225, 220, 204, 235}, gameui.TextSmall, 15, 3)
	y += factionPanelTechSummaryH

	drawUISectionLabel(screen, 0, y, "Kaynaklar")
	y += factionPanelSectionH
	drawFactionResourceGrid(screen, gs, fid, f, 0, y, float32(width))
	y += regionPanelStatRowGap * 4

	drawUISectionLabel(screen, 0, y, "Ticaret")
	y += factionPanelSectionH
	trade := factionTradeStats(gs, fid)
	drawUIKeyValueRow(screen, 0, y, width, "Ortak", itoa(trade.PartnerCount)+" devlet", ColorGray, ColorWhite)
	y += factionPanelRowH
	drawUIKeyValueRow(screen, 0, y, width, "Hat", itoa(trade.RouteCount)+" aktif", ColorGray, ColorWhite)
	y += factionPanelRowH
	drawUIKeyValueRow(screen, 0, y, width, "İhracat", itoa(trade.ExportGold)+" altın / tur", ColorGray, ColorGold)
	y += factionPanelRowH
	if trade.SuspendedCount > 0 {
		drawUIKeyValueRow(screen, 0, y, width, "Askıda", itoa(trade.SuspendedCount)+" rota", ColorGray, color.RGBA{210, 160, 90, 255})
		y += factionPanelRowH
	}
	if rel := factionRelationToPlayer(gs, fid); rel != nil && fid != gs.PlayerFactionID {
		drawUIKeyValueRow(screen, 0, y, width, "Oyuncuya Durum", faction.DiplomaticStanceLabelTR(rel.Stance)+" ("+itoa(rel.Score)+")", ColorGray, ColorWhite)
		y += 22
	} else {
		y += 4
	}

	drawUISectionLabel(screen, 0, y, "Diplomasi Özeti")
	y += factionPanelSectionH
	overlordValue := "Yok"
	if summary.HasOverlord {
		overlordValue = summary.Overlord.Name
	}
	drawUIKeyValueRow(screen, 0, y, width, "Üst Devlet", overlordValue, ColorGray, factionDiplomacyEntryColor(faction.StanceAllied))
	y += factionPanelRowH
	drawUIKeyValueRow(screen, 0, y, width, "Vassal", itoa(summary.VassalCount)+" devlet", ColorGray, color.RGBA{126, 170, 220, 255})
	y += factionPanelRowH
	drawUIKeyValueRow(screen, 0, y, width, "İttifak", itoa(summary.AllianceCount)+" devlet", ColorGray, factionDiplomacyEntryColor(faction.StanceAllied))
	y += factionPanelRowH
	drawUIKeyValueRow(screen, 0, y, width, "Ticaret", itoa(summary.TradeCount)+" devlet", ColorGray, factionDiplomacyEntryColor(faction.StanceTrade))
	y += factionPanelRowH
	drawUIKeyValueRow(screen, 0, y, width, "Düşman", itoa(summary.EnemyCount)+" devlet", ColorGray, factionDiplomacyEntryColor(faction.StanceWar))
	y += factionPanelRowH

	y = drawFactionDiplomacyGroup(screen, 0, y, width, "Vassallar", summary.Vassals, "Bağlı", false, factionDiplomacyEntryColor(faction.StanceAllied))
	y = drawFactionDiplomacyGroup(screen, 0, y, width, "İttifaklar", summary.Allies, faction.DiplomaticStanceLabelTR(faction.StanceAllied), true, factionDiplomacyEntryColor(faction.StanceAllied))
	y = drawFactionDiplomacyGroup(screen, 0, y, width, "Ticaret Anlaşmaları", summary.Trade, faction.DiplomaticStanceLabelTR(faction.StanceTrade), true, factionDiplomacyEntryColor(faction.StanceTrade))
	y = drawFactionDiplomacyGroup(screen, 0, y, width, "Düşmanlar", summary.Enemies, faction.DiplomaticStanceLabelTR(faction.StanceWar), true, factionDiplomacyEntryColor(faction.StanceWar))

	drawUISectionLabel(screen, 0, y, "Tamamlanan Teknolojiler")
	y += factionPanelSectionH
	drawFactionCompletedTechList(screen, gs, f, 0, y, width)
	y += factionCompletedTechContentHeight(gs, f)
	drawFactionAIDebugSection(screen, gs, fid, 0, y, width)
}

func factionCompletedTechContentHeight(gs *state.GameState, f *faction.Faction) float64 {
	names, hidden := factionCompletedTechPreview(gs, f, 12)
	if len(names) == 0 {
		return 18
	}
	height := float64((len(names)+1)/2) * 16
	if hidden > 0 {
		height += 20
	}
	return height
}

func factionAIDebugVisible(gs *state.GameState, fid faction.FactionID) bool {
	if gs == nil || !gs.DevelopmentMode || fid == "" || fid == gs.PlayerFactionID || gs.AIStrategies == nil {
		return false
	}
	_, ok := gs.AIStrategies[string(fid)]
	return ok
}

func factionAIDebugStrategy(gs *state.GameState, fid faction.FactionID) (scenario.AIFactionStrategy, bool) {
	if !factionAIDebugVisible(gs, fid) {
		return scenario.AIFactionStrategy{}, false
	}
	strategy, ok := gs.AIStrategies[string(fid)]
	return strategy, ok
}

func factionAIDebugActiveObjective(gs *state.GameState, fid faction.FactionID, strategy scenario.AIFactionStrategy) (scenario.AIObjectiveDef, bool) {
	if gs == nil || gs.AIPlans == nil {
		return scenario.AIObjectiveDef{}, false
	}
	plan := gs.AIPlans[fid]
	if plan == nil || plan.ObjectiveID == "" {
		return scenario.AIObjectiveDef{}, false
	}
	for _, objective := range strategy.Objectives {
		if objective.ID == plan.ObjectiveID {
			return objective, true
		}
	}
	return scenario.AIObjectiveDef{}, false
}

func factionAIDebugClaims(strategy scenario.AIFactionStrategy, objective scenario.AIObjectiveDef, hasObjective bool) []scenario.AITerritorialClaimDef {
	if hasObjective && len(objective.TerritorialClaims) > 0 {
		return objective.TerritorialClaims
	}
	return strategy.TerritorialClaims
}

func factionAIDebugYearRange(objective scenario.AIObjectiveDef) string {
	minYear := "hemen"
	if objective.MinYear > 0 {
		minYear = itoa(objective.MinYear)
	}
	maxYear := "süresiz"
	if objective.MaxYear > 0 {
		maxYear = itoa(objective.MaxYear)
	}
	return minYear + " - " + maxYear
}

func factionAIDebugKindLabel(kind string) string {
	switch kind {
	case "expand":
		return "genişleme"
	case "defend":
		return "savunma"
	case "consolidate":
		return "konsolidasyon"
	case "ally":
		return "ittifak"
	default:
		return kind
	}
}

func factionAIDebugObjectiveSummary(objective scenario.AIObjectiveDef) string {
	return objective.ID + " • " + factionAIDebugKindLabel(objective.Kind) + " • " + factionAIDebugYearRange(objective) + " • " + itoa(len(objective.TerritorialClaims)) + " claim"
}

func factionAIDebugClaimLabel(gs *state.GameState, claim scenario.AITerritorialClaimDef) string {
	regionName := claim.RegionID
	ownerName := "sahipsiz"
	if gs != nil {
		if region := gs.Regions[world.RegionID(claim.RegionID)]; region != nil {
			regionName = region.NameTR
			if strings.TrimSpace(regionName) == "" {
				regionName = region.Name
			}
			if strings.TrimSpace(regionName) == "" {
				regionName = claim.RegionID
			}
			if region.OwnerID != "" {
				ownerName = factionDisplayName(gs, region.OwnerID)
			}
		}
	}
	return "• " + regionName + " → " + ownerName + " (" + itoa(claim.Value) + ")"
}

func factionAIDebugContentHeight(gs *state.GameState, fid faction.FactionID, _ float64) float64 {
	strategy, ok := factionAIDebugStrategy(gs, fid)
	if !ok {
		return 0
	}

	height := factionPanelSectionH + factionPanelRowH*2
	activeObjective, hasActiveObjective := factionAIDebugActiveObjective(gs, fid, strategy)
	claims := factionAIDebugClaims(strategy, activeObjective, hasActiveObjective)
	if hasActiveObjective {
		height += factionPanelRowH + factionPanelSectionH
		if len(claims) == 0 {
			height += factionPanelRowH
		} else {
			height += float64(len(claims)) * factionPanelRowH
		}
	} else if len(claims) > 0 {
		height += factionPanelSectionH + float64(len(claims))*factionPanelRowH
	}

	otherCount := 0
	for _, objective := range strategy.Objectives {
		if !hasActiveObjective || objective.ID != activeObjective.ID {
			otherCount++
		}
	}
	if otherCount > 0 {
		height += factionPanelSectionH + float64(otherCount)*factionPanelRowH
	}
	return height + 4
}

func drawFactionAIDebugSection(screen *ebiten.Image, gs *state.GameState, fid faction.FactionID, x, y, width float64) {
	strategy, ok := factionAIDebugStrategy(gs, fid)
	if !ok {
		return
	}

	drawUISectionLabel(screen, x, y, "AI Stratejisi (DEV)")
	y += factionPanelSectionH
	profile := strategy.Profile
	if strings.TrimSpace(profile) == "" {
		profile = "tanımsız"
	}
	drawUIKeyValueRow(screen, x, y, width, "Profil", profile, ColorGray, ColorWhite)
	y += factionPanelRowH

	plan := (*state.AIPlanState)(nil)
	if gs.AIPlans != nil {
		plan = gs.AIPlans[fid]
	}
	focus := "plan yok"
	if plan != nil && plan.ObjectiveID != "" {
		focus = plan.ObjectiveID + " [" + factionAIDebugKindLabel(string(plan.Kind)) + "]"
	}
	drawUIKeyValueRow(screen, x, y, width, "Odak", focus, ColorGray, ColorGold)
	y += factionPanelRowH

	activeObjective, hasActiveObjective := factionAIDebugActiveObjective(gs, fid, strategy)
	claims := factionAIDebugClaims(strategy, activeObjective, hasActiveObjective)
	if hasActiveObjective {
		drawUIKeyValueRow(screen, x, y, width, "Tarih", factionAIDebugYearRange(activeObjective), ColorGray, ColorWhite)
		y += factionPanelRowH
		drawUISectionLabel(screen, x, y, "Claim hedefleri")
		y += factionPanelSectionH
		if len(claims) == 0 {
			drawUILabel(screen, gameui.Rect{X: x, Y: y, W: width}, "• Tanımlı claim yok", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
			y += factionPanelRowH
		} else {
			for _, claim := range claims {
				label := trimTextToWidth(factionAIDebugClaimLabel(gs, claim), FaceSmall, width)
				drawUILabel(screen, gameui.Rect{X: x, Y: y, W: width}, label, ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
				y += factionPanelRowH
			}
		}
	} else if len(claims) > 0 {
		drawUISectionLabel(screen, x, y, "Genel claim hedefleri")
		y += factionPanelSectionH
		for _, claim := range claims {
			label := trimTextToWidth(factionAIDebugClaimLabel(gs, claim), FaceSmall, width)
			drawUILabel(screen, gameui.Rect{X: x, Y: y, W: width}, label, ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
			y += factionPanelRowH
		}
	}

	otherObjectives := 0
	for _, objective := range strategy.Objectives {
		if hasActiveObjective && objective.ID == activeObjective.ID {
			continue
		}
		if otherObjectives == 0 {
			drawUISectionLabel(screen, x, y, "Diğer objective'ler")
			y += factionPanelSectionH
		}
		label := trimTextToWidth("• "+factionAIDebugObjectiveSummary(objective), FaceSmall, width)
		drawUILabel(screen, gameui.Rect{X: x, Y: y, W: width}, label, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		y += factionPanelRowH
		otherObjectives++
	}
}

func drawFactionDiplomacyGroup(screen *ebiten.Image, x, y, width float64, title string, entries []factionDiplomacyEntry, suffix string, showScore bool, accent color.RGBA) float64 {
	if len(entries) == 0 {
		return y
	}
	drawUISectionLabel(screen, x, y, title)
	y += factionPanelSectionH
	for _, entry := range entries {
		label := factionDiplomacyEntryLabel(entry, suffix, showScore)
		drawUILabel(screen, gameui.Rect{X: x, Y: y, W: width}, trimTextToWidth("• "+label, FaceSmall, width), accent, gameui.TextSmall, gameui.TextAlignStart)
		y += factionPanelRowH
	}
	return y
}

func loadSettlementImage(region *world.Region, settlement *world.Settlement) *ebiten.Image {
	if region == nil || settlement == nil || ActiveScenarioPath == "" {
		return nil
	}
	cacheKey := string(region.ID) + "::" + settlement.ID
	if settlementImageLoaded[cacheKey] {
		return settlementImageCache[cacheKey]
	}

	candidates := settlementImageCandidates(region, settlement)
	for _, p := range candidates {
		if img := tryLoadImage(p); img != nil {
			settlementImageLoaded[cacheKey] = true
			settlementImageCache[cacheKey] = img
			return img
		}
	}

	settlementImageLoaded[cacheKey] = true
	settlementImageCache[cacheKey] = nil
	return nil
}

func settlementImageCandidates(region *world.Region, settlement *world.Settlement) []string {
	base := filepath.Join(ActiveScenarioPath, "images", "settlements")
	id := strings.TrimSpace(settlement.ID)
	rid := strings.TrimSpace(string(region.ID))
	out := make([]string, 0, 8)
	if id != "" {
		out = append(out,
			filepath.Join(base, id+".png"),
			filepath.Join(base, id+".jpg"),
			filepath.Join(base, id+".jpeg"),
		)
	}
	if rid != "" {
		out = append(out,
			filepath.Join(base, rid+".png"),
			filepath.Join(base, rid+".jpg"),
			filepath.Join(base, rid+".jpeg"),
		)
	}
	return out
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// regionDiplomacyButtonHit oyuncuya ait olmayan bölge panelindeki Diplomasi butonunu döner.
func regionDiplomacyButtonHit(mx, my float64, gs *state.GameState, rid world.RegionID) bool {
	return regionDiplomacyButtonHitForTab(mx, my, gs, rid, regionPanelTabBuildings)
}

func regionPanelActionBarY(gs *state.GameState, region *world.Region, activeTab regionPanelTab) float64 {
	if gs == nil || region == nil {
		return 0
	}
	start := buildingGridStartY(gs, region, false)
	end := buildingGridEndY(gs, region, start)
	if activeTab == regionPanelTabEvents {
		end = regionPanelTabContentEndY(gs, region, start, end)
	}
	return float64(end) + 5
}

func regionDiplomacyButtonHitForTab(mx, my float64, gs *state.GameState, rid world.RegionID, activeTab regionPanelTab) bool {
	if rid == "" || gs == nil {
		return false
	}
	region, ok := gs.Regions[rid]
	if !ok || region.IsSea || region.OwnerID == "" || region.OwnerID == string(gs.PlayerFactionID) {
		return false
	}
	barY := float32(regionPanelActionBarY(gs, region, activeTab))
	bar := gameui.Rect{X: float64(infoPanelX()) + panelPad, Y: float64(barY), W: float64(infoPanelW) - panelPad*2, H: regionPanelActionBarHeight}
	return buildRegionDiplomacyButtons(gs, region.OwnerID, float32(bar.X), float32(bar.Y), float32(bar.W), float32(bar.H)).HitTest(mx, my)
}

func regionGrainAidButtonHit(mx, my float64, gs *state.GameState, rid world.RegionID) bool {
	return regionGrainAidButtonHitForTab(mx, my, gs, rid, regionPanelTabBuildings)
}

func regionGrainAidButtonHitForTab(mx, my float64, gs *state.GameState, rid world.RegionID, activeTab regionPanelTab) bool {
	if gs == nil || rid == "" {
		return false
	}
	region := gs.Regions[rid]
	if region == nil || region.IsSea || region.OwnerID != string(gs.PlayerFactionID) {
		return false
	}
	barY := float32(regionPanelActionBarY(gs, region, activeTab))
	bar := gameui.Rect{X: float64(infoPanelX()) + panelPad, Y: float64(barY), W: float64(infoPanelW) - panelPad*2, H: regionPanelActionBarHeight}
	return buildRegionGrainAidButton(gs, rid, float32(bar.X), float32(bar.Y), float32(bar.W), float32(bar.H)).HitTest(mx, my)
}

func regionLiberationSuccessor(gs *state.GameState, region *world.Region) (faction.FactionID, bool) {
	if gs == nil || region == nil || region.IsSea || region.OwnerID != string(gs.PlayerFactionID) || region.SuccessorFactionID == "" {
		return "", false
	}
	successorID := faction.FactionID(region.SuccessorFactionID)
	successor := gs.Factions[successorID]
	if successor == nil || !successor.IsEliminated || len(gs.LandRegionsOwnedBy(successorID)) != 0 || successorID == gs.PlayerFactionID {
		return "", false
	}
	return successorID, true
}

func buildRegionLiberateButton(px, py, pw, ph float32) gameui.Button {
	const btnW = float32(112)
	const btnH = float32(24)
	x := px + pw - btnW - 5
	y := py + (ph-btnH)/2
	return gameui.NewButton(float64(x), float64(y), float64(btnW), float64(btnH), "Özgürleştir")
}

func regionLiberateButtonHit(mx, my float64, gs *state.GameState, rid world.RegionID) bool {
	return regionLiberateButtonHitForTab(mx, my, gs, rid, regionPanelTabBuildings)
}

func regionLiberateButtonHitForTab(mx, my float64, gs *state.GameState, rid world.RegionID, activeTab regionPanelTab) bool {
	if gs == nil || rid == "" {
		return false
	}
	region := gs.Regions[rid]
	if _, ok := regionLiberationSuccessor(gs, region); !ok {
		return false
	}
	barY := float32(regionPanelActionBarY(gs, region, activeTab))
	bar := gameui.Rect{X: float64(infoPanelX()) + panelPad, Y: float64(barY), W: float64(infoPanelW) - panelPad*2, H: regionPanelActionBarHeight}
	return buildRegionLiberateButton(float32(bar.X), float32(bar.Y), float32(bar.W), float32(bar.H)).HitTest(mx, my)
}

func armyPanelCloseHit(mx, my float64) bool {
	return buildArmyPanelCloseButton().HitTest(mx, my)
}

func regionTaxButtonHit(mx, my float64, gs *state.GameState, rid world.RegionID) int {
	region, ok := gs.Regions[rid]
	if !ok || region.IsSea || region.IsLocked || region.OwnerID != string(gs.PlayerFactionID) {
		return 0
	}
	dec, inc := buildRegionTaxButtons(gs, rid)
	if dec.HitTest(mx, my) {
		return -5
	}
	if inc.HitTest(mx, my) {
		return 5
	}
	return 0
}

func buildRegionTaxButtons(gs *state.GameState, rid world.RegionID) (gameui.Button, gameui.Button) {
	dec, inc := regionTaxButtonRects(gs, rid)
	return buttonFromRectF32(dec, "-"), buttonFromRectF32(inc, "+")
}

func buildRegionDiplomacyButtons(_ *state.GameState, _ string, px, py, pw, ph float32) gameui.Button {
	x, y, w, h := regionDiplomacyButtonRect(0, px, py, pw, ph)
	return gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "Diplomasi")
}

func buildRegionGrainAidButton(_ *state.GameState, _ world.RegionID, px, py, _, ph float32) gameui.Button {
	const btnW = float32(112)
	const btnH = float32(24)
	x := px + 5
	y := py + (ph-btnH)/2
	return gameui.NewButton(float64(x), float64(y), float64(btnW), float64(btnH), "Tahıl Yardımı")
}

func regionTaxButtonRects(gs *state.GameState, rid world.RegionID) ([4]float32, [4]float32) {
	px := infoPanelX()
	pw := infoPanelW
	ownerID := ""
	if gs != nil && rid != "" {
		if region := gs.Regions[rid]; region != nil {
			ownerID = region.OwnerID
		}
	}
	ly := regionPanelStatRowsStartY(gs, ownerID)
	y := float32(ly + regionPanelStatRowGap + (regionPanelStatRowGap-float64(regionPanelTaxButtonH))/2 - 1)
	contentRight := px + pw - float32(panelPad)
	incX := contentRight - regionPanelTaxButtonW
	decX := incX - regionPanelTaxButtonGap - regionPanelTaxButtonW
	return [4]float32{decX, y, regionPanelTaxButtonW, regionPanelTaxButtonH}, [4]float32{incX, y, regionPanelTaxButtonW, regionPanelTaxButtonH}
}

func regionPanelStatRowsStartY(gs *state.GameState, ownerID string) float64 {
	ly := float64(infoPanelY()) + 10
	ly += 24
	ly += regionOwnerBlockHeight(gs, ownerID)
	if gs.DevelopmentMode {
		ly += 34
	}
	ly += 16 // arazi/din satırı
	ly += 16 // nüfus satırı
	ly += 8
	ly += regionPanelStatRowGap * 4
	return ly
}

func regionPanelTaxBarLayout(x float32, width float32) (float32, float32) {
	barX := x + 96 + regionPanelMeterValueW + regionPanelMeterGap
	barW := width - (barX - x)
	if barW < 0 {
		barW = 0
	}
	return barX, barW
}

func regionPanelTaxInteractiveBarLayout(x, width, decreaseButtonX float32) (float32, float32) {
	barX, _ := regionPanelTaxBarLayout(x, width)
	barW := decreaseButtonX - barX - regionPanelTaxButtonGap
	if barW < 0 {
		barW = 0
	}
	return barX, barW
}

func regionPopulationDisplayText(region *world.Region) string {
	if region == nil {
		return "Nüfus: -"
	}
	return fmt.Sprintf("Nüfus: %d  (Kırsal: %d / Yerleşim: %d)", region.Population, region.RuralPopulation, region.SettlementPopulation())
}

func drawRegionMeterRow(screen *ebiten.Image, x, y float64, width float32, label, value string, fill float64, barColor color.Color) {
	drawRegionMeterLabels(screen, x, y, label, value)
	barX, barW := regionPanelTaxBarLayout(float32(x), width)
	drawBar(screen, barX, float32(y)+regionPanelBarYOffset, barW, regionPanelBarH, fill, barColor)
}

func drawRegionMeterLabels(screen *ebiten.Image, x, y float64, label, value string) {
	labelW := 96.0
	valueX := x + labelW
	drawUILabel(screen, gameui.Rect{X: x, Y: y, W: labelW}, label, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: valueX, Y: y, W: float64(regionPanelMeterValueW)}, value, ColorWhite, gameui.TextMedium, gameui.TextAlignStart)
}

func drawRegionProductionGrid(screen *ebiten.Image, gs *state.GameState, region *world.Region, x, y float64, width float32, production state.RegionProductionSummary) {
	type productionItem struct {
		kind  economy.ResourceKind
		value string
		col   color.Color
	}

	grainValue := regionGrainProductionDisplayValue(gs, region, production)
	items := [...]productionItem{
		{kind: economy.ResourceGold, value: itoa(production.Gold), col: ColorGold},
		{kind: economy.ResourceGrain, value: grainValue, col: ColorWhite},
		{kind: economy.ResourceIron, value: itoa(production.Iron), col: color.RGBA{200, 205, 215, 255}},
		{kind: economy.ResourceTimber, value: itoa(production.Timber), col: color.RGBA{145, 205, 145, 255}},
		{kind: economy.ResourceStone, value: itoa(production.Stone), col: color.RGBA{185, 185, 185, 255}},
		{kind: economy.ResourceSpice, value: itoa(production.Spice), col: color.RGBA{230, 165, 90, 255}},
		{kind: economy.ResourceCloth, value: itoa(production.Cloth), col: color.RGBA{175, 150, 220, 255}},
	}

	colGap := 14.0
	colW := (float64(width) - colGap) / 2
	for i, item := range items {
		col := float64(i % 2)
		row := float64(i / 2)
		drawUIKeyValueRow(
			screen,
			x+col*(colW+colGap),
			y+row*regionPanelStatRowGap,
			colW,
			economy.ResourceNameTR(item.kind),
			item.value,
			ColorGray,
			item.col,
		)
	}
}

func regionGrainProductionDisplayValue(gs *state.GameState, region *world.Region, production state.RegionProductionSummary) string {
	if gs == nil || region == nil {
		return itoa(production.Grain)
	}
	return fmt.Sprintf("+%d/%d", gs.RegionMilitaryGrainProduction(region), production.Grain)
}

func BuildingGridHitTest(mx, my float64, gs *state.GameState, rid world.RegionID, neighborExpanded bool) string {
	return buildingGridHitTest(mx, my, gs, rid, neighborExpanded)
}

func BuildingGridDemolishHitTest(mx, my float64, gs *state.GameState, rid world.RegionID, neighborExpanded bool) (string, bool) {
	if gs == nil || rid == "" {
		return "", false
	}
	region, ok := gs.Regions[rid]
	if !ok || !regionBuildingDemolitionAvailable(gs, region) {
		return "", false
	}
	if bid, ok := lastDrawnBuildingGridDemolishHit(mx, my, rid); ok {
		return bid, true
	}
	px := infoPanelX()
	startY := buildingGridStartY(gs, region, neighborExpanded)
	for _, card := range buildBuildingCardComponents(gs, region, px, startY, infoPanelW) {
		if card.CanDemolish && card.DemolishBtn.HitTest(mx, my) {
			return card.ID, true
		}
	}
	return "", false
}

func buildingGridHitTest(mx, my float64, gs *state.GameState, rid world.RegionID, neighborExpanded bool) string {
	if gs == nil || rid == "" {
		return ""
	}
	region, ok := gs.Regions[rid]
	if !ok || !regionBuildingActionsAvailable(gs, region) {
		return ""
	}
	if bid, ok := lastDrawnBuildingGridHit(mx, my, rid); ok {
		return bid
	}
	px := infoPanelX()
	pw := infoPanelW
	startY := buildingGridStartY(gs, region, neighborExpanded)

	for _, card := range buildBuildingCardComponents(gs, region, px, startY, pw) {
		if card.HitTest(mx, my) {
			return card.ID
		}
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
	if bid == "port" {
		return true
	}
	return b.RequiredTerrain == "" || string(region.Terrain) == b.RequiredTerrain
}

func buildingGridStartY(gs *state.GameState, region *world.Region, _ bool) float32 {
	if gs == nil || region == nil {
		return 0
	}
	py := infoPanelY()
	ly := float64(py) + 10
	ly += 24
	ly += regionOwnerBlockHeight(gs, region.OwnerID)
	if gs.DevelopmentMode {
		ly += 34
	}
	ly += 16 + 16 + 8 // arazi/din ve nüfus satırları
	ly += regionPanelStatRowGap * 4
	ly += regionPanelStatRowGap * 2
	if logistics, ok := gs.RegionLogistics[region.ID]; ok && logistics.Demand > 0 {
		ly += regionPanelStatRowGap
		if logistics.FriendlySupplyGrainSpent > 0 {
			ly += 14
		}
		if logistics.Overload > 0 {
			ly += 14
		}
	}
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
	ly += 4 + 6 + float64(regionPanelTabH) + 6
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
	drawUIProgressBar(screen, x, y, w, h, fill, color.RGBA{40, 40, 40, 180}, color.RGBA{}, col, 0)
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
			if string(fid) == string(gs.PlayerFactionID) {
				return f.NameTR + " (Siz)", ColorWhite
			}
			return f.NameTR, ColorWhite
		}
	}
	return ownerID, ColorGray
}

func vassalOverlordDisplay(gs *state.GameState, ownerID string) (string, color.Color, bool) {
	if gs == nil || ownerID == "" {
		return "", nil, false
	}
	overlord := diplomacy.DirectOverlord(gs, faction.FactionID(ownerID))
	if overlord == "" {
		return "", nil, false
	}
	name, _ := ownerDisplay(gs, string(overlord))
	if f := gs.Factions[overlord]; f != nil {
		return name, color.RGBA{f.Color[0], f.Color[1], f.Color[2], 235}, true
	}
	return name, ColorGold, true
}

func vassalTributeDisplay(gs *state.GameState, ownerID string) (string, color.Color, bool) {
	if gs == nil || ownerID == "" || gs.PlayerFactionID == "" {
		return "", nil, false
	}
	f := gs.Factions[faction.FactionID(ownerID)]
	if f == nil || f.IsEliminated || f.OverlordID != gs.PlayerFactionID {
		return "", nil, false
	}
	tribute := projectedFactionTributeToOverlord(gs, faction.FactionID(ownerID))
	return "Haraç: +" + itoa(tribute) + " altın/tur", ColorGold, true
}

func projectedFactionTributeToOverlord(gs *state.GameState, fid faction.FactionID) int {
	income := projectedFactionTributeIncome(gs, fid)
	if income <= 0 {
		return 0
	}
	return income * diplomacy.VassalTributeRatePercent() / 100
}

func projectedFactionTributeIncome(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" {
		return 0
	}
	f := gs.Factions[fid]
	if f == nil || f.IsEliminated {
		return 0
	}

	season := gs.CurrentSeason()
	harvestMod := season.HarvestMod()
	tradeMod := season.TradeMod()

	var fx tech.Effects
	if gs.TechTypes != nil {
		fx = tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
	}

	income := 0
	ownedCount := 0
	for _, region := range gs.Regions {
		if region == nil || region.OwnerID != string(fid) {
			continue
		}
		ownedCount++
		if region.IsSea || gs.SiegeAt(region.ID) != nil {
			continue
		}

		goldMod := 1.0
		for _, bid := range region.Buildings {
			if building, ok := gs.BuildingTypes[bid]; ok && building != nil {
				goldMod *= building.GoldMod
			}
		}

		retention := gs.RegionBlockadeOutputRetentionPercent(region)
		regionIncome := state.ScaleBlockadeOutputForEconomy(int(float64(region.GoldIncome())*goldMod*float64(harvestMod)/100), retention)
		tradeIncome := gs.BaseRegionTradeIncome(region)
		tradeIncome = tradeIncome * tradeMod / 100
		tradeIncome = state.ScaleBlockadeOutputForEconomy(tradeIncome, retention)
		if fx.MarketGoldMod != 0 {
			tradeIncome = int(float64(tradeIncome) * (1.0 + fx.MarketGoldMod))
		}

		income += regionIncome + tradeIncome
		if bonus := gs.CapitalRegionBonus(region); bonus != (state.RegionProductionSummary{}) {
			income += bonus.Gold
		}
	}

	income += gs.BlockadeLootForFaction(fid).Gold
	income += fx.GoldPerRegion * ownedCount
	if income < 0 {
		return 0
	}
	return income
}

func regionOwnerBlockHeight(gs *state.GameState, ownerID string) float64 {
	height := float64(regionOwnerNameH)
	if _, _, ok := vassalOverlordDisplay(gs, ownerID); ok {
		height += regionVassalInfoH
	}
	if _, _, ok := vassalTributeDisplay(gs, ownerID); ok {
		height += regionVassalInfoH
	}
	return height
}

func ownerLabelOutlineColor(fill color.Color) color.RGBA {
	r, g, b, _ := fill.RGBA()
	luminance := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
	if luminance >= 160 {
		return color.RGBA{18, 16, 12, 220}
	}
	return color.RGBA{245, 240, 230, 210}
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
func DrawSeaRegionPanel(screen *ebiten.Image, gs *state.GameState, region *world.Region, neighborExpanded bool) {
	px := infoPanelX()
	py := infoPanelY()
	pw := infoPanelW
	ph := infoPanelH

	// Panel arka plan
	drawUIPanelFrame(screen, gameui.Rect{X: float64(px), Y: float64(py), W: float64(pw), H: float64(ph)}, panelBg, panelBorder, 1.5, 3)
	drawPanelCloseButton(screen, px, py, pw)

	lx := float64(px) + panelPad
	ly := float64(py) + 10
	sepW := pw - float32(panelPad*2)

	// Başlık
	DrawText(screen, region.NameTR, lx, ly, FaceLarge, color.RGBA{100, 180, 255, 255})
	ly += 24

	// Development mode bilgileri
	if gs.DevelopmentMode {
		drawUIRichTextBlock(screen, gameui.Rect{X: lx, Y: ly}, []gameui.RichTextLine{
			{Text: "ID: " + string(region.ID), Color: ColorGray, Variant: gameui.TextSmall, Align: gameui.TextAlignStart},
			{Text: "Koordinat: " + itoa(region.WorldX) + "," + itoa(region.WorldY), Color: ColorGray, Variant: gameui.TextSmall, Align: gameui.TextAlignStart},
		}, 16)
		ly += 34
	}

	// Deniz bölgesi (italik vurgu)
	drawUIKeyValueRow(screen, lx, ly, float64(sepW), "Tip", "Deniz Bölgesi", ColorGray, color.RGBA{120, 160, 200, 200})
	ly += 18
	if region.IsLocked {
		lockValue := "Kilitli"
		if region.UnlockTurn > 0 {
			lockValue += "  |  Açılış Turu: " + itoa(region.UnlockTurn)
		}
		drawUIKeyValueRow(screen, lx, ly, float64(sepW), "Durum", lockValue, ColorGray, color.RGBA{220, 150, 90, 220})
		ly += 18
	}

	drawUISeparator(screen, float32(lx), float32(ly), float32(lx)+sepW, 1, panelBorder)
	ly += 8

	drawNeighborBlock(screen, gs, region, lx, ly, sepW, neighborExpanded, ColorGold)
}

const (
	devNeighborCollapsedCount = 4
	devNeighborLineHeight     = 16.0
	devNeighborTitleHeight    = 18.0
	devNeighborColumnGap      = 12.0
)

func drawNeighborBlock(screen *ebiten.Image, gs *state.GameState, region *world.Region, x, y float64, width float32, expanded bool, titleColor color.Color) float64 {
	title, items, cols, rows := neighborBlockLayout(gs, region, expanded)
	drawUILabel(screen, gameui.Rect{X: x, Y: y, W: float64(width)}, title, titleColor, gameui.TextSmall, gameui.TextAlignStart)
	y += devNeighborTitleHeight
	if len(items) == 0 {
		return devNeighborTitleHeight
	}

	colW := (float64(width) - 15 - devNeighborColumnGap*float64(cols-1)) / float64(cols)
	for idx, item := range items {
		col := idx / rows
		row := idx % rows
		lineX := x + 15 + float64(col)*(colW+devNeighborColumnGap)
		lineY := y + float64(row)*devNeighborLineHeight
		drawUILabel(screen, gameui.Rect{X: lineX, Y: lineY, W: colW}, item.Text, item.Color, item.Variant, gameui.TextAlignStart)
	}
	return devNeighborTitleHeight + float64(rows)*devNeighborLineHeight
}

func neighborBlockLayout(gs *state.GameState, region *world.Region, expanded bool) (string, []gameui.RichTextLine, int, int) {
	total := len(region.Neighbors)
	title := "Komşu Bölgeler:"
	if total == 0 {
		title = "Komşu: Yok"
	} else {
		title = "Komşu (" + itoa(total) + ")"
		if total > devNeighborCollapsedCount {
			if expanded {
				title += "  [Daralt]"
			} else {
				title += "  [Tümünü Göster]"
			}
		} else {
			title += ":"
		}
	}

	displayCount := total
	if !expanded && displayCount > devNeighborCollapsedCount {
		displayCount = devNeighborCollapsedCount
	}

	items := make([]gameui.RichTextLine, 0, displayCount)
	for i := 0; i < displayCount; i++ {
		neighborID := region.Neighbors[i]
		neighborRegion, ok := gs.Regions[neighborID]
		if !ok || neighborRegion == nil {
			continue
		}
		col := color.RGBA{180, 180, 180, 200}
		if neighborRegion.IsSea {
			col = color.RGBA{100, 160, 220, 200}
		}
		items = append(items, gameui.RichTextLine{
			Text:    "• " + neighborRegion.NameTR,
			Color:   col,
			Variant: gameui.TextSmall,
			Align:   gameui.TextAlignStart,
		})
	}

	cols := 1
	if expanded && len(items) > 8 {
		cols = 2
	}
	rows := 0
	if len(items) > 0 {
		rows = (len(items) + cols - 1) / cols
	}
	return title, items, cols, rows
}

func regionNeighborToggleHit(mx, my float64, gs *state.GameState, rid world.RegionID) bool {
	if rid == "" || gs == nil {
		return false
	}
	region, ok := gs.Regions[rid]
	if !ok || region == nil || len(region.Neighbors) <= devNeighborCollapsedCount {
		return false
	}
	// Kara bölge komşuları artık ayrı scroll viewport'unda tam liste olarak
	// gösteriliyor; eski genişlet/daralt hit-test'i bina/aksiyon alanına
	// sarkmamalı. Deniz paneli mevcut toggle davranışını korur.
	if !region.IsSea {
		return false
	}
	x, y, w := neighborToggleRect(gs, region)
	return mx >= x && mx <= x+w && my >= y && my <= y+devNeighborTitleHeight
}

func neighborToggleRect(gs *state.GameState, region *world.Region) (float64, float64, float64) {
	x := float64(infoPanelX()) + panelPad
	y := neighborBlockStartY(gs, region)
	w := float64(infoPanelW - float32(panelPad*2))
	return x, y, w
}

func neighborBlockStartY(gs *state.GameState, region *world.Region) float64 {
	ly := float64(infoPanelY()) + 10
	ly += 24
	if !region.IsSea {
		ly += 18
	}
	if gs.DevelopmentMode {
		ly += 34
	}
	if region.IsSea {
		ly += 18
		if region.IsLocked {
			ly += 18
		}
		ly += 8
		return ly
	}

	ly += 16
	ly += 8
	ly += regionPanelStatRowGap * 4
	ly += regionPanelStatRowGap * 2
	if logistics, ok := gs.RegionLogistics[region.ID]; ok && logistics.Demand > 0 {
		ly += regionPanelStatRowGap
		if logistics.FriendlySupplyGrainSpent > 0 {
			ly += 14
		}
		if logistics.Overload > 0 {
			ly += 14
		}
	}
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
	return ly
}
