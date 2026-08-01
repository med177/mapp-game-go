package render

import (
	"image/color"
	"sort"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	activeWarsPanelW        = 570.0
	activeWarsPanelMaxH     = 600.0
	activeWarsPanelPad      = 12.0
	activeWarsPanelHeaderH  = 42.0
	activeWarRowH           = 82.0
	activeWarRowGap         = 7.0
	activeWarFlagSize       = 50.0
	activeWarFlagSidePad    = 10.0
	activeWarFlagTextGap    = 10.0
	activeWarsHudButtonGap  = 20.0
	activeWarsHudButtonSize = 36.0
	activeWarsHudButtonStep = 46.0
	activeWarsHudButtonTop  = 5.0
	activeWarsPanelGap      = 12.0
)

// ActiveWarSummary, panelin harita state'inden bağımsız çizilebilir snapshot'ıdır.
// Güç değerleri mevcut ordulardan, süre ve kayıplar WarLedger'dan okunur.
type ActiveWarSummary struct {
	FactionANameTR string
	FactionBNameTR string
	FactionA       faction.FactionID
	FactionB       faction.FactionID
	Turns          int
	PowerA         int
	PowerB         int
	ArmiesA        int
	ArmiesB        int
	UnitsA         int
	UnitsB         int
	CasualtiesA    int
	CasualtiesB    int
}

// topHudUtilityButtonRect, müzik kartının sağındaki ortak küçük düğme
// şeridindeki slotu döndürür. Yeni HUD durum düğmeleri aynı şeritte slot
// index'i artırılarak yan yana eklenebilir.
func topHudUtilityButtonRect(slot int) [4]float32 {
	if slot < 0 {
		slot = 0
	}
	x, y, w, _ := musicHudRect()
	return [4]float32{x + w + activeWarsHudButtonGap + float32(slot)*activeWarsHudButtonStep, y + activeWarsHudButtonTop, activeWarsHudButtonSize, activeWarsHudButtonSize}
}

func activeWarsHudButtonRect() [4]float32 {
	return topHudUtilityButtonRect(0)
}

func buildActiveWarsHUDButton() gameui.Button {
	r := activeWarsHudButtonRect()
	btn := buttonFromRectF32(r, "").WithIcon(gameui.IconSword)
	btn.IconSize = 21
	return btn
}

func activeWarsHudButtonHit(mx, my float64) bool {
	return buildActiveWarsHUDButton().HitTest(mx, my)
}

func activeWarsPanelRect() gameui.Rect {
	_, musicY, _, musicH := musicHudRect()
	y := float64(musicY + musicH + 6)
	h := activeWarsPanelMaxH
	if remaining := ScreenHeight - y - 16; remaining < h {
		h = remaining
	}
	if h < 150 {
		h = 150
	}
	// Olaylar paneli ekranın sağ kenarında kalır; aktif savaşlar onun
	// solundaki ayrılmış kolona oturur ve iki panel yatayda çakışmaz.
	right := float64(evLogX()) - activeWarsPanelGap
	if right < 10 {
		right = 10
	}
	panelW := activeWarsPanelW
	if available := right - 10; available < panelW {
		panelW = available
	}
	if panelW < 0 {
		panelW = 0
	}
	return gameui.Rect{
		X: right - panelW,
		Y: y,
		W: panelW,
		H: h,
	}
}

func activeWarsPanelCloseButton() gameui.Button {
	panel := activeWarsPanelRect()
	btn := gameui.NewButton(panel.X+panel.W-38, panel.Y+9, 26, 24, "").WithIcon(gameui.IconClose)
	btn.IconSize = 13
	return btn
}

func activeWarsPanelViewport() gameui.Rect {
	panel := activeWarsPanelRect()
	return gameui.Rect{
		X: panel.X + activeWarsPanelPad,
		Y: panel.Y + activeWarsPanelHeaderH,
		W: panel.W - activeWarsPanelPad*2,
		H: panel.H - activeWarsPanelHeaderH - activeWarsPanelPad,
	}
}

func activeWarsPanelHit(mx, my float64) bool {
	return activeWarsPanelRect().Hit(mx, my)
}

// activeWarsPanelInteractiveHit yalnızca gerçekten tıklanabilen savaş satırları
// ve ortak kapatma düğmesi için pointer cursor üretir. Panelin geri kalanı
// bilgi/scroll yüzeyidir; dışı ise harita input'una bırakılır.
func activeWarsPanelInteractiveHit(r *Renderer, mx, my float64) bool {
	if r == nil || !r.showActiveWars {
		return false
	}
	if activeWarsPanelCloseButton().HitTest(mx, my) {
		return true
	}
	return activeWarRowAt(mx, my, r.activeWarsBuf, r.activeWarsScroll) >= 0
}

func activeWarVisibleRows(viewport gameui.Rect) int {
	rows := int((viewport.H + activeWarRowGap) / (activeWarRowH + activeWarRowGap))
	if rows < 1 {
		return 1
	}
	return rows
}

func activeWarMaxScroll(entryCount int, viewport gameui.Rect) int {
	maxScroll := entryCount - activeWarVisibleRows(viewport)
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func clampActiveWarScroll(entryCount int, viewport gameui.Rect, scroll int) int {
	if scroll < 0 {
		return 0
	}
	if maxScroll := activeWarMaxScroll(entryCount, viewport); scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func activeWarRowRect(viewport gameui.Rect, visibleIndex int) gameui.Rect {
	return gameui.Rect{
		X: viewport.X,
		Y: viewport.Y + float64(visibleIndex)*(activeWarRowH+activeWarRowGap),
		W: viewport.W,
		H: activeWarRowH,
	}
}

func activeWarRowAt(mx, my float64, wars []ActiveWarSummary, scroll int) int {
	viewport := activeWarsPanelViewport()
	if !viewport.Hit(mx, my) {
		return -1
	}
	visibleRows := activeWarVisibleRows(viewport)
	end := scroll + visibleRows
	if end > len(wars) {
		end = len(wars)
	}
	for index := scroll; index < end; index++ {
		if activeWarRowRect(viewport, index-scroll).Hit(mx, my) {
			return index
		}
	}
	return -1
}

func activeWarFactionName(gs *state.GameState, id faction.FactionID) string {
	if gs != nil {
		if f := gs.Factions[id]; f != nil {
			if f.NameTR != "" {
				return f.NameTR
			}
			if f.Name != "" {
				return f.Name
			}
		}
	}
	return string(id)
}

func activeWarArmyStats(gs *state.GameState, owner faction.FactionID) (armies, units int) {
	if gs == nil {
		return 0, 0
	}
	for _, a := range gs.Armies {
		if a == nil || a.OwnerID != string(owner) {
			continue
		}
		armies++
		units += len(a.Units) + len(a.EmbarkedUnits)
	}
	return armies, units
}

// collectActiveWarSummaries yalnızca savaş stance'ındaki ilişkileri okur.
// dst Renderer tarafından yeniden kullanıldığı için normal Draw akışında yeni
// özet slice'ı oluşturmaz.
func collectActiveWarSummaries(gs *state.GameState, dst []ActiveWarSummary) []ActiveWarSummary {
	dst = dst[:0]
	if gs == nil {
		return dst
	}
	for _, rel := range gs.Relations {
		if rel == nil || rel.Stance != faction.StanceWar || rel.FactionA == "" || rel.FactionB == "" || rel.FactionA == rel.FactionB {
			continue
		}
		a, b := rel.FactionA, rel.FactionB
		if b < a {
			a, b = b, a
		}
		ledger := gs.WarLedgerFor(a, b)
		startedTurn := gs.Turn
		casualtiesA, casualtiesB := 0, 0
		if ledger != nil {
			startedTurn = ledger.StartedTurn
			casualtiesA = ledger.CasualtiesA
			casualtiesB = ledger.CasualtiesB
			if a != ledger.FactionA {
				casualtiesA, casualtiesB = casualtiesB, casualtiesA
			}
		}
		turns := gs.Turn - startedTurn
		if turns < 0 {
			turns = 0
		}
		armiesA, unitsA := activeWarArmyStats(gs, a)
		armiesB, unitsB := activeWarArmyStats(gs, b)
		dst = append(dst, ActiveWarSummary{
			FactionA:       a,
			FactionB:       b,
			FactionANameTR: activeWarFactionName(gs, a),
			FactionBNameTR: activeWarFactionName(gs, b),
			Turns:          turns,
			PowerA:         diplomacy.MilitaryPower(gs, a),
			PowerB:         diplomacy.MilitaryPower(gs, b),
			ArmiesA:        armiesA,
			ArmiesB:        armiesB,
			UnitsA:         unitsA,
			UnitsB:         unitsB,
			CasualtiesA:    casualtiesA,
			CasualtiesB:    casualtiesB,
		})
	}
	sort.SliceStable(dst, func(i, j int) bool {
		if dst[i].FactionA != dst[j].FactionA {
			return dst[i].FactionA < dst[j].FactionA
		}
		return dst[i].FactionB < dst[j].FactionB
	})
	return dst
}

func countActiveWars(gs *state.GameState) int {
	count := 0
	if gs == nil {
		return count
	}
	for _, rel := range gs.Relations {
		if rel != nil && rel.Stance == faction.StanceWar && rel.FactionA != "" && rel.FactionB != "" && rel.FactionA != rel.FactionB {
			count++
		}
	}
	return count
}

func activeWarRepresentativeRegion(gs *state.GameState, fid faction.FactionID) *world.Region {
	if gs == nil || fid == "" {
		return nil
	}
	if capital, _, _, ok := gs.FactionCapital(fid); ok && capital != nil {
		return capital
	}
	var representative *world.Region
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		if representative == nil || region.ID < representative.ID {
			representative = region
		}
	}
	return representative
}

func (r *Renderer) focusActiveWar(war ActiveWarSummary) bool {
	if r == nil || r.gs == nil {
		return false
	}
	regionA := activeWarRepresentativeRegion(r.gs, war.FactionA)
	regionB := activeWarRepresentativeRegion(r.gs, war.FactionB)
	if regionA == nil && regionB == nil {
		return false
	}

	focusX, focusY := 0.0, 0.0
	switch {
	case regionA != nil && regionB != nil:
		focusX = (wcX(regionA.WorldX) + wcX(regionB.WorldX)) / 2
		focusY = (wcY(regionA.WorldY) + wcY(regionB.WorldY)) / 2
	case regionA != nil:
		focusX = wcX(regionA.WorldX)
		focusY = wcY(regionA.WorldY)
	default:
		focusX = wcX(regionB.WorldX)
		focusY = wcY(regionB.WorldY)
	}
	r.camX, r.camY = clampCameraCenter(focusX, focusY, r.camScale)
	return true
}

func drawActiveWarsHUDButton(screen *ebiten.Image, gs *state.GameState, open bool) {
	btn := buildActiveWarsHUDButton()
	centerX := float32(btn.X + btn.W/2)
	centerY := float32(btn.Y + btn.H/2)
	fill := color.RGBA{58, 38, 24, 235}
	border := color.RGBA{166, 124, 60, 255}
	if open {
		fill = color.RGBA{144, 76, 36, 250}
		border = color.RGBA{255, 208, 104, 255}
	} else if countActiveWars(gs) == 0 {
		fill = color.RGBA{38, 34, 28, 220}
		border = color.RGBA{110, 100, 82, 220}
	}
	vector.FillCircle(screen, centerX, centerY, float32(btn.W/2), fill, true)
	vector.StrokeCircle(screen, centerX, centerY, float32(btn.W/2-1), 1.5, border, true)
	gameui.DrawIcon(screen, btn.Icon, btn.X+(btn.W-btn.IconSize)/2, btn.Y+(btn.H-btn.IconSize)/2, btn.IconSize, ColorWhite)

	count := countActiveWars(gs)
	if count > 0 {
		badgeX, badgeY := centerX+8, centerY-12
		vector.FillCircle(screen, badgeX, badgeY, 8, color.RGBA{170, 42, 32, 255}, true)
		drawUILabel(screen, gameui.Rect{X: float64(badgeX - 8), Y: float64(badgeY - 7), W: 16, H: 14}, itoa(count), ColorWhite, gameui.TextSmall, gameui.TextAlignCenter)
	}
}

func drawActiveWarsScrollbar(screen *ebiten.Image, viewport gameui.Rect, entryCount, scroll int) {
	maxScroll := activeWarMaxScroll(entryCount, viewport)
	if maxScroll <= 0 {
		return
	}
	track := gameui.Rect{X: viewport.X + viewport.W - 5, Y: viewport.Y, W: 3, H: viewport.H}
	drawUICardRect(screen, track, color.RGBA{42, 35, 25, 210}, color.RGBA{90, 72, 44, 180}, 1)
	thumbH := track.H * float64(activeWarVisibleRows(viewport)) / float64(entryCount)
	if thumbH < 24 {
		thumbH = 24
	}
	scroll = clampActiveWarScroll(entryCount, viewport, scroll)
	thumbY := track.Y
	if track.H > thumbH {
		thumbY += (track.H - thumbH) * float64(scroll) / float64(maxScroll)
	}
	drawUICardRect(screen, gameui.Rect{X: track.X, Y: thumbY, W: track.W, H: thumbH}, color.RGBA{190, 148, 74, 235}, color.RGBA{238, 206, 130, 220}, 1)
}

func activeWarFactionFlagColor(gs *state.GameState, id faction.FactionID) color.RGBA {
	if gs != nil {
		if f := gs.Factions[id]; f != nil {
			return color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		}
	}
	return color.RGBA{62, 52, 38, 255}
}

func activeWarRowContentRects(row gameui.Rect) (leftFlag, center, rightFlag gameui.Rect) {
	leftFlag = gameui.Rect{
		X: row.X + activeWarFlagSidePad,
		Y: row.Y + (row.H-activeWarFlagSize)/2,
		W: activeWarFlagSize,
		H: activeWarFlagSize,
	}
	rightFlag = gameui.Rect{
		X: row.X + row.W - activeWarFlagSidePad - activeWarFlagSize,
		Y: leftFlag.Y,
		W: activeWarFlagSize,
		H: activeWarFlagSize,
	}
	center = gameui.Rect{
		X: leftFlag.X + leftFlag.W + activeWarFlagTextGap,
		Y: row.Y + 4,
		W: rightFlag.X - (leftFlag.X + leftFlag.W) - activeWarFlagTextGap*2,
		H: row.H - 8,
	}
	return leftFlag, center, rightFlag
}

func drawActiveWarRow(screen *ebiten.Image, gs *state.GameState, row gameui.Rect, war ActiveWarSummary) {
	drawUICardRect(screen, row, color.RGBA{28, 22, 16, 235}, color.RGBA{92, 68, 38, 215}, 1)
	leftFlag, center, rightFlag := activeWarRowContentRects(row)
	drawFactionFlagBadge(screen, war.FactionA, factionInitial(war.FactionANameTR), leftFlag.X, leftFlag.Y, leftFlag.W, activeWarFactionFlagColor(gs, war.FactionA), panelBorder)
	drawFactionFlagBadge(screen, war.FactionB, factionInitial(war.FactionBNameTR), rightFlag.X, rightFlag.Y, rightFlag.W, activeWarFactionFlagColor(gs, war.FactionB), panelBorder)

	drawUILabel(screen, gameui.Rect{X: center.X, Y: row.Y + 7, W: center.W, H: 18}, trimTextToWidth(war.FactionANameTR+"  ↔  "+war.FactionBNameTR, FaceMed, center.W), color.RGBA{255, 220, 118, 255}, gameui.TextMedium, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: center.X, Y: row.Y + 28, W: center.W, H: 16}, trimTextToWidth(itoa(war.Turns)+" turdur savaşta • Kayıp "+itoa(war.CasualtiesA)+" / "+itoa(war.CasualtiesB), FaceSmall, center.W), ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: center.X, Y: row.Y + 47, W: center.W, H: 16}, trimTextToWidth("Güç "+itoa(war.PowerA)+"  ↔  "+itoa(war.PowerB), FaceSmall, center.W), color.RGBA{220, 202, 164, 255}, gameui.TextSmall, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: center.X, Y: row.Y + 64, W: center.W, H: 16}, trimTextToWidth("Ordu "+itoa(war.ArmiesA)+" ("+itoa(war.UnitsA)+" birim)  ↔  "+itoa(war.ArmiesB)+" ("+itoa(war.UnitsB)+" birim)", FaceSmall, center.W), color.RGBA{176, 168, 148, 245}, gameui.TextSmall, gameui.TextAlignCenter)
}

func drawActiveWarsPanel(screen *ebiten.Image, gs *state.GameState, wars []ActiveWarSummary, scroll int) {
	panel := activeWarsPanelRect()
	drawUIPanelFrame(screen, panel, color.RGBA{12, 10, 8, 244}, color.RGBA{154, 112, 54, 255}, 1.5, 5)
	drawUILabel(screen, gameui.Rect{X: panel.X + activeWarsPanelPad, Y: panel.Y + 9, W: panel.W - 60}, "Aktif Savaşlar ("+itoa(len(wars))+")", color.RGBA{255, 220, 118, 255}, gameui.TextMedium, gameui.TextAlignStart)
	drawUIButtonWidget(screen, activeWarsPanelCloseButton(), eventLogButtonStyle(ColorGray))

	viewport := activeWarsPanelViewport()
	if len(wars) == 0 {
		drawUILabel(screen, viewport, "Aktif savaş bulunmuyor.", ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
		return
	}
	scroll = clampActiveWarScroll(len(wars), viewport, scroll)
	visibleRows := activeWarVisibleRows(viewport)
	end := scroll + visibleRows
	if end > len(wars) {
		end = len(wars)
	}
	for i := scroll; i < end; i++ {
		drawActiveWarRow(screen, gs, activeWarRowRect(viewport, i-scroll), wars[i])
	}
	drawActiveWarsScrollbar(screen, viewport, len(wars), scroll)
}

// handleActiveWarsOverlayInput yalnız panelin kendi yüzeyindeki input'u tüketir.
// Panel dışındaki sol/sağ tıklama ve orta tuş sürüklemesi ana harita akışına kalır.
func (r *Renderer) handleActiveWarsOverlayInput() bool {
	if r == nil || !r.showActiveWars {
		return false
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.showActiveWars = false
		r.activeWarsScroll = 0
		return true
	}
	mx, my := ebiten.CursorPosition()
	if !activeWarsPanelHit(float64(mx), float64(my)) {
		return false
	}
	if _, wheelY := ebiten.Wheel(); wheelY != 0 {
		viewport := activeWarsPanelViewport()
		r.activeWarsScroll = clampActiveWarScroll(len(r.activeWarsBuf), viewport, r.activeWarsScroll-int(wheelY))
		return true
	}
	if activeWarsPanelCloseButton().HitTest(float64(mx), float64(my)) && r.mouseJustPressed(ebiten.MouseButtonLeft) {
		r.showActiveWars = false
		r.activeWarsScroll = 0
		return true
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if index := activeWarRowAt(float64(mx), float64(my), r.activeWarsBuf, r.activeWarsScroll); index >= 0 {
			r.focusActiveWar(r.activeWarsBuf[index])
			return true
		}
	}
	leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	leftWasPressed := r.prevMouse[ebiten.MouseButtonLeft]
	if leftPressed || leftWasPressed {
		r.prevMouse[ebiten.MouseButtonLeft] = leftPressed
		return true
	}
	rightPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	rightWasPressed := r.prevMouse[ebiten.MouseButtonRight]
	if rightPressed || rightWasPressed {
		r.prevMouse[ebiten.MouseButtonRight] = rightPressed
		return true
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		r.isDragging = false
		return true
	}
	return false
}
