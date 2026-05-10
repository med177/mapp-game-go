package render

import (
	"image/color"
	"math"
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/audio"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	ScreenWidth  float64 = 1280
	ScreenHeight float64 = 720
	mapPitchY            = 1.0 // Düz 2D harita ölçeği
	mapShearX            = 0.0 // Harita bükme/yatıklık şimdilik kapalı
)

const (
	confirmDialogW    = float32(460)
	confirmDialogH    = float32(166)
	confirmDialogBtnW = float32(120)
	confirmDialogBtnH = float32(36)
)

// Renderer kamerayı ve dünya haritasını yönetir.
type Renderer struct {
	gs       *state.GameState
	worldMap *WorldMap

	// Kamera: dünya uzayında merkez noktası ve zoom
	camX, camY float64
	camScale   float64

	// Sürükleme takibi
	lastMX, lastMY int
	isDragging     bool

	// Seçim
	SelectedRegion world.RegionID
	SelectedArmy   army.ArmyID

	// Senaryo seçim ekranı
	scenarioCursor int

	// Fraksiyon seçim ekranı
	factionCursor int

	// Diplomasi paneli
	showDiplomacy  bool
	diplomacyFocus int

	// Teknoloji paneli
	showTech   bool
	techCursor int

	// Ana menü
	menuTick        int
	HasSave         bool
	HasAutoSave     bool
	CurrentSettings Settings
	LoadingMessage  string

	// Duraklama menüsü
	pauseCursor int

	// Kayıt/yükleme slot seçim ekranı
	slotCursor        int
	saveSelectMode    bool   // true=kaydetme, false=yükleme
	pendingDeleteSlot string // onay bekleyen slot adı ("" = onay yok)

	// Olay logu (sağ üst panel)
	eventLog          []string
	eventLogCollapsed bool
	eventDetail       string
	eventLogScroll    int

	// Savaş / bildirim mesajı (kısa süreli)
	combatLog      string
	combatLogTimer int

	// Tarihsel olay tam ekran bildirimi
	historicalEventTitle string
	historicalEventDesc  string
	showHistoricalEvent  bool

	// İlk frame kamera başlatma
	firstDraw bool

	// Input state (just-pressed takibi)
	prevKeys  map[ebiten.Key]bool
	prevMouse map[ebiten.MouseButton]bool

	// Genel onay diyaloğu
	warConfirm    warConfirmState
	confirmDialog confirmDialogState

	armyIconBuf    []armyIconPos
	regionLabelBuf []settlementDraw
	labelRectBuf   []screenRect

	editSelectedRegion         world.RegionID
	editSelectedSettlement     int
	editDraggingSettlement     bool
	editDraggingRegion         bool
	editRenaming               bool
	editNameRunes              []rune
	editDirty                  bool
	editVoronoiDebug           bool
	editOwnerDropdown          *Dropdown
	editTerrainDropdown        *Dropdown
	editSettlementTypeDropdown *Dropdown
	editVisualNeighborBuf      []world.RegionID
	editBoundaryPixelBuf       []int
}

type confirmDialogState struct {
	show          bool
	title         string
	message       string
	messageLines  []string
	acceptLabel   string
	declineLabel  string
	thirdLabel    string
	pendingAction InputAction
	thirdAction   InputAction
	declineHook   func()
}

type warConfirmState struct {
	show        bool
	factionName string
	factionID   string
	pendingArmy army.ArmyID
	pendingDest world.RegionID
}

// Dropdown component for reusable dropdown UI
type Dropdown struct {
	open       bool
	scroll     int
	options    []string
	selected   string
	x, y, w, h float32
	title      string
}

// NewDropdown creates a new dropdown with given position and size
func NewDropdown(x, y, w, h float32, title string) *Dropdown {
	return &Dropdown{
		x: x, y: y, w: w, h: h,
		title: title,
	}
}

// SetPosition sets the dropdown position
func (d *Dropdown) SetPosition(x, y float32) {
	d.x, d.y = x, y
}

// SetOptions sets the dropdown options and resets selection
func (d *Dropdown) SetOptions(options []string, selected string) {
	d.options = make([]string, len(options))
	copy(d.options, options)
	d.selected = selected
	d.scroll = 0
}

// Toggle opens/closes the dropdown
func (d *Dropdown) Toggle() {
	d.open = !d.open
	if d.open {
		d.scroll = 0
	}
}

// Close closes the dropdown
func (d *Dropdown) Close() {
	d.open = false
	d.scroll = 0
}

// IsOpen returns whether dropdown is open
func (d *Dropdown) IsOpen() bool {
	return d.open
}

// HitTest checks if point is inside dropdown
func (d *Dropdown) HitTest(mx, my float64) bool {
	return mx >= float64(d.x) && mx <= float64(d.x+d.w) && my >= float64(d.y) && my <= float64(d.y+d.h)
}

// Scroll adjusts scroll position
func (d *Dropdown) Scroll(dy float64) {
	if dy > 0 {
		d.scroll--
	} else if dy < 0 {
		d.scroll++
	}
	maxScroll := len(d.options) - editOwnerDropdownVisibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if d.scroll < 0 {
		d.scroll = 0
	}
	if d.scroll > maxScroll {
		d.scroll = maxScroll
	}
}

// GetSelectedOption returns the selected option index and whether valid
func (d *Dropdown) GetSelectedOption(mx, my float64) (int, bool) {
	if mx < float64(d.x+8) || mx > float64(d.x+d.w-8) {
		return 0, false
	}
	startY := float64(d.y + editOwnerDropdownHeaderH)
	if my < startY {
		return 0, false
	}
	row := int((my - startY) / float64(editOwnerDropdownRowH))
	if row < 0 || row >= editOwnerDropdownVisibleRows {
		return 0, false
	}
	idx := d.scroll + row
	if idx < 0 || idx >= len(d.options) {
		return 0, false
	}
	return idx, true
}

// Draw renders the dropdown
func (d *Dropdown) Draw(screen *ebiten.Image) {
	if !d.open {
		return
	}
	drawRoundedRect(screen, d.x, d.y, d.w, d.h, 6, color.RGBA{16, 20, 24, 242})
	drawPanelBorder(screen, d.x, d.y, d.w, d.h)
	DrawText(screen, d.title, float64(d.x)+10, float64(d.y)+8, FaceSmall, ColorGold)

	rowX := float64(d.x) + 8
	rowY := float64(d.y) + float64(editOwnerDropdownHeaderH)
	rowW := float64(d.w) - 16
	for i := 0; i < editOwnerDropdownVisibleRows; i++ {
		optionIndex := d.scroll + i
		if optionIndex >= len(d.options) {
			break
		}
		option := d.options[optionIndex]
		oy := rowY + float64(i)*float64(editOwnerDropdownRowH)
		bg := color.RGBA{28, 24, 18, 220}
		txt := ColorWhite
		if option == d.selected {
			bg = color.RGBA{86, 64, 24, 238}
			txt = ColorGold
		}
		vector.FillRect(screen, float32(rowX), float32(oy), float32(rowW), editOwnerDropdownRowH-2, bg, false)
		DrawText(screen, option, rowX+8, oy+5, FaceSmall, txt)
	}
	if len(d.options) > editOwnerDropdownVisibleRows {
		DrawText(screen, itoa(d.scroll+1)+"-"+itoa(editMinInt(d.scroll+editOwnerDropdownVisibleRows, len(d.options)))+"/"+itoa(len(d.options)),
			float64(d.x)+float64(d.w)-68, float64(d.y)+8, FaceSmall, ColorGray)
	}
}

// New başlangıç kamera pozisyonuyla yeni bir Renderer döner.
func New(gs *state.GameState) *Renderer {
	x, y, w, _ := editInspectorRect()
	dropW := float32(292)
	dropH := editOwnerDropdownHeaderH + editOwnerDropdownRowH*editOwnerDropdownVisibleRows + 10
	dropX := x + w + 8
	dropY := y

	r := &Renderer{
		gs:                         gs,
		worldMap:                   NewWorldMap(gs),
		prevKeys:                   make(map[ebiten.Key]bool),
		prevMouse:                  make(map[ebiten.MouseButton]bool),
		editVoronoiDebug:           true,
		editVisualNeighborBuf:      make([]world.RegionID, 0, 16),
		editBoundaryPixelBuf:       make([]int, 0, 4096),
		editOwnerDropdown:          NewDropdown(dropX, dropY, dropW, dropH, "Sahip Sec"),
		editTerrainDropdown:        NewDropdown(dropX, dropY, dropW, dropH, "Arazi Tipi"),
		editSettlementTypeDropdown: NewDropdown(dropX, dropY, dropW, dropH, "Yerlesim Tipi"),
	}
	r.resetCamera()
	return r
}

// resetCamera kamerayı mevcut ScreenWidth/ScreenHeight'e göre dünyayı tam dolduracak şekilde ayarlar.
func (r *Renderer) resetCamera() {
	r.camScale = minCameraScale()
	r.camX = float64(WorldW) / 2
	// Haritanın üst kenarını ekranın üstüne hizala
	r.camY = ScreenHeight / (2 * r.camScale)
}

func minCameraScale() float64 {
	scaleX := ScreenWidth / float64(WorldW)
	scaleY := ScreenHeight / float64(WorldH)
	return math.Min(scaleX, scaleY)
}

// SetCursor menü veya ekran imlecini sıfırlar.
func (r *Renderer) SetCursor(n int) { r.factionCursor = n }

// MarkMapDirty sahiplik değiştiğinde çağrılır.
func (r *Renderer) MarkMapDirty() { r.worldMap.MarkDirty() }

func (r *Renderer) MarkEditSaved() { r.editDirty = false }

func (r *Renderer) SetLoadingMessage(message string) {
	r.LoadingMessage = message
}

// ReloadGameState yükleme sonrası yeni state ve yeni worldmap ile günceller.
// ActiveScenarioPath aktif senaryonun klasör yolu; asset yükleyiciler buradan türetir.
var ActiveScenarioPath string

func (r *Renderer) ReloadGameState(gs *state.GameState) {
	r.gs = gs
	if gs.ScenarioPath != "" {
		ActiveScenarioPath = gs.ScenarioPath
		// Senaryo değişince asset cache'lerini sıfırla
		buildingSheetLoaded = false
		miniMapLoaded = false
		armySheetLoaded = false
	}
	r.worldMap = NewWorldMap(gs)
	r.resetCamera()
	r.SelectedRegion = ""
	r.SelectedArmy = ""
	r.eventLogScroll = 0
}

// AddEvent olay loguna yeni bir giriş ekler.
func (r *Renderer) AddEvent(msg string) {
	r.eventLog = append([]string{msg}, r.eventLog...)
	if len(r.eventLog) > maxEventLogEntries {
		r.eventLog = r.eventLog[:maxEventLogEntries]
	}
	r.eventLogScroll = 0
}

// ShowCombatResult oyun içi kısa uyarı/bilgi mesajını ekranda ~3 saniye gösterir.
func (r *Renderer) ShowCombatResult(msg string) {
	r.combatLog = msg
	r.combatLogTimer = 180
	audio.PlaySound("combat")
}

// ShowHistoricalEvent büyük tarihsel olayı tam ekran popup olarak gösterir.
func (r *Renderer) ShowHistoricalEvent(title, desc string) {
	r.historicalEventTitle = title
	r.historicalEventDesc = desc
	r.showHistoricalEvent = true
	r.AddEvent(title)
}

// ShowTechPanel teknoloji panelini açar.
func (r *Renderer) ShowTechPanel() {
	r.showTech = true
	r.techCursor = 0
}

// --- Kamera dönüşümleri ---

func (r *Renderer) worldToScreen(wx, wy float64) (float64, float64) {
	dx := wx - r.camX
	dy := wy - r.camY
	sx := (dx+dy*mapShearX)*r.camScale + ScreenWidth/2
	sy := dy*r.camScale*mapPitchY + ScreenHeight/2
	return sx, sy
}

func (r *Renderer) screenToWorld(sx, sy float64) (float64, float64) {
	dy := (sy - ScreenHeight/2) / (r.camScale * mapPitchY)
	dx := (sx-ScreenWidth/2)/r.camScale - dy*mapShearX
	wx := r.camX + dx
	wy := r.camY + dy
	return wx, wy
}

func (r *Renderer) applyMapGeoM(op *ebiten.DrawImageOptions, sourceW, sourceH float64) {
	scaleX := float64(WorldW) / sourceW
	scaleY := float64(WorldH) / sourceH

	op.GeoM.SetElement(0, 0, r.camScale*scaleX)
	op.GeoM.SetElement(0, 1, r.camScale*mapShearX*scaleY)
	op.GeoM.SetElement(1, 0, 0)
	op.GeoM.SetElement(1, 1, r.camScale*mapPitchY*scaleY)
	op.GeoM.SetElement(0, 2, ScreenWidth/2-r.camScale*r.camX-r.camScale*mapShearX*r.camY)
	op.GeoM.SetElement(1, 2, ScreenHeight/2-r.camScale*mapPitchY*r.camY)
}

// --- Draw ---

// Draw her frame çağrılır.
func (r *Renderer) Draw(screen *ebiten.Image) {
	// İlk frame'de Layout() zaten gerçek pencere boyutunu güncellemiştir;
	// kamerayı bu boyuta göre yeniden ayarla.
	if !r.firstDraw {
		r.firstDraw = true
		r.resetCamera()
	}

	// Ana menü
	if r.gs.Phase == state.PhaseMainMenu {
		r.menuTick++
		DrawMainMenu(screen, r.factionCursor, r.HasSave, r.HasAutoSave, r.menuTick)
		return
	}

	if r.gs.Phase == state.PhaseLoading {
		r.menuTick++
		DrawLoadingScreen(screen, r.LoadingMessage, r.menuTick)
		return
	}

	// Ayarlar ekranı
	if r.gs.Phase == state.PhaseSettings {
		DrawSettingsScreen(screen, r.CurrentSettings, r.factionCursor)
		return
	}

	// Senaryo seçim ekranı
	if r.gs.Phase == state.PhaseScenarioSelect {
		DrawScenarioSelect(screen, ScenarioList, r.scenarioCursor)
		return
	}

	// Fraksiyon seçim ekranı
	if r.gs.Phase == "faction_select" {
		DrawFactionSelect(screen, r.gs, r.factionCursor)
		return
	}

	// Zafer koşulu seçim ekranı
	if r.gs.Phase == "victory_select" {
		DrawVictorySelect(screen, r.gs, r.factionCursor)
		return
	}

	// Oyun sonu ekranı
	if r.gs.Phase == "game_over" {
		drawGameOver(screen, r.gs)
		return
	}

	// Kayıt/yükleme slot seçim ekranı
	if r.gs.Phase == state.PhaseLoadSelect {
		DrawSlotSelectScreen(screen, r.slotCursor, false, r.pendingDeleteSlot)
		return
	}
	if r.gs.Phase == state.PhaseSaveSelect {
		DrawSlotSelectScreen(screen, r.slotCursor, true, r.pendingDeleteSlot)
		return
	}

	// Duraklama menüsü — haritayı altta çiz, üstüne overlay
	if r.gs.Phase == state.PhasePauseMenu {
		r.worldMap.Refresh(r.gs, r.SelectedRegion)
		mapOp := &ebiten.DrawImageOptions{}
		r.applyMapGeoM(mapOp, float64(WorldW), float64(WorldH))
		screen.DrawImage(r.worldMap.Image(), mapOp)
		r.menuTick++
		DrawPauseMenu(screen, r.pauseCursor, r.HasSave, r.menuTick, r.CurrentSettings)
		return
	}

	// Seçili bölge veya donanmanın deniz bölgesini vurgula
	highlightRegion := world.RegionID(r.SelectedRegion)
	if r.SelectedArmy != "" {
		if a, ok := r.gs.Armies[r.SelectedArmy]; ok {
			if reg, ok2 := r.gs.Regions[a.RegionID]; ok2 && reg.IsSea {
				highlightRegion = a.RegionID
			}
		}
	}
	r.worldMap.Refresh(r.gs, highlightRegion)

	// 1. Üretilen dünya haritası
	mapOp := &ebiten.DrawImageOptions{}
	r.applyMapGeoM(mapOp, float64(WorldW), float64(WorldH))
	screen.DrawImage(r.worldMap.Image(), mapOp)

	// 2. Seçim vurgusu (bölge) kaldırıldı

	// 3. Ordu hareket hedefleri
	if r.selectedArmyIsPlayerOwned() {
		r.drawMoveTargets(screen)
	}

	armyPositions := r.armyIconPositions()

	// 4. Bölge etiketleri
	r.drawRegionLabels(screen, armyPositions)
	if r.gs.Phase == state.PhaseEditMode {
		r.drawEditRegionCenters(screen)
		r.drawEditVoronoiDebug(screen)
	}

	// 5. Ordu ikonları
	r.drawArmies(screen, armyPositions)

	// 6. UI panelleri
	if r.gs.Phase != state.PhaseEditMode {
		DrawBottomPanel(screen, r.gs, r.showDiplomacy, r.showTech)
		DrawRegionPanel(screen, r.gs, r.SelectedRegion)
		DrawRecruitPanel(screen, r.gs, r.SelectedRegion)
		DrawArmyDetailPanel(screen, r.gs, r.SelectedArmy)
	}
	DrawMinimap(screen, r.gs, r.camX, r.camY, r.camScale)
	if r.gs.Phase != state.PhaseEditMode {
		DrawEventLog(screen, r.eventLog, r.eventLogCollapsed, r.eventLogScroll)
		DrawHoverTooltip(screen, r.gs, r.SelectedRegion)
	} else {
		r.drawEditModeHud(screen)
		r.drawEditInspector(screen)
	}

	// 7. Diplomasi paneli (üst katman)
	if r.showDiplomacy {
		DrawDiplomacyPanel(screen, r.gs, r.diplomacyFocus)
	}

	// 8. Teknoloji paneli (üst katman)
	if r.showTech {
		r.DrawTechPanel(screen)
	}

	// 9. Bildirim mesajı
	if r.combatLogTimer > 0 {
		alpha := uint8(255)
		if r.combatLogTimer < 60 {
			alpha = uint8(r.combatLogTimer * 255 / 60)
		}
		drawInfoPopup(screen, r.combatLog, alpha)
		r.combatLogTimer--
	}

	// 10. Onay diyalogu (diğer popupların altında kalmaması için üst katman)
	if r.confirmDialog.show {
		r.drawConfirmDialog(screen)
	} else if r.warConfirm.show {
		r.drawWarConfirmDialog(screen)
	}

	// 11. Tarihsel olay tam ekran popup
	if r.showHistoricalEvent {
		drawHistoricalEventPopup(screen, r.historicalEventTitle, r.historicalEventDesc)
	}

	if r.eventDetail != "" {
		drawEventDetailPopup(screen, r.eventDetail)
	}
}

// drawSelectionHighlight seçili bölgenin üstüne vurgu çizer.
func (r *Renderer) drawSelectionHighlight(screen *ebiten.Image) {
	region, ok := r.gs.Regions[r.SelectedRegion]
	if !ok {
		return
	}

	sx, sy := r.regionScreenPos(region)

	if region.IsSea {
		// Deniz bölgesi seçimi: büyük beyaz daire halkası
		vector.StrokeCircle(screen, float32(sx), float32(sy), 28, 2.5, color.RGBA{180, 230, 255, 200}, true)
		vector.StrokeCircle(screen, float32(sx), float32(sy), 20, 1.5, color.RGBA{100, 200, 255, 160}, true)
	} else {
		// Kara bölgesi seçimi
		vector.StrokeCircle(screen, float32(sx), float32(sy+4), 16, 3, color.RGBA{255, 220, 70, 230}, true)
		vector.StrokeCircle(screen, float32(sx), float32(sy+4), 22, 1.5, color.RGBA{30, 20, 5, 180}, true)
	}
}

func (r *Renderer) selectedArmyIsPlayerOwned() bool {
	a, ok := r.gs.Armies[r.SelectedArmy]
	return ok && a.OwnerID == string(r.gs.PlayerFactionID)
}

func armyCanEnterRegion(a *army.Army, target *world.Region) bool {
	if target == nil || target.IsLocked {
		return false
	}
	if a.IsNaval {
		return target.CanNavalEnter()
	}
	return target.CanLandEnter()
}

func enemyArmyInPlayerMoveRange(gs *state.GameState, targetArmy *army.Army) bool {
	if targetArmy == nil || targetArmy.OwnerID == string(gs.PlayerFactionID) {
		return false
	}
	for _, playerArmy := range gs.Armies {
		if playerArmy.OwnerID != string(gs.PlayerFactionID) || playerArmy.MovePoints <= 0 {
			continue
		}
		src, ok := gs.Regions[playerArmy.RegionID]
		if !ok {
			continue
		}
		for _, nid := range src.Neighbors {
			if nid != targetArmy.RegionID {
				continue
			}
			targetRegion, ok := gs.Regions[nid]
			if ok && armyCanEnterRegion(playerArmy, targetRegion) {
				return true
			}
		}
	}
	return false
}

// drawMoveTargets seçili ordunun gidebileceği komşu bölgeleri vurgular.
func (r *Renderer) drawMoveTargets(screen *ebiten.Image) {
	a, ok := r.gs.Armies[r.SelectedArmy]
	if !ok || a.OwnerID != string(r.gs.PlayerFactionID) || a.MovePoints <= 0 {
		return
	}
	src, ok := r.gs.Regions[a.RegionID]
	if !ok {
		return
	}

	for _, nid := range src.Neighbors {
		nRegion, ok := r.gs.Regions[nid]
		if !ok || nRegion.IsLocked {
			continue
		}
		// Naval: sadece deniz bölgelerine gidebilir; kara: sadece kara bölgelerine
		if a.IsNaval && !nRegion.IsSea {
			continue
		}
		if !a.IsNaval && nRegion.IsSea {
			continue
		}

		sx, sy := r.regionScreenPos(nRegion)

		var col color.RGBA
		if a.IsNaval {
			// Deniz bölgeleri için sabit açık mavi — tarafsız su
			col = color.RGBA{100, 200, 255, 220}
		} else {
			switch {
			case nRegion.OwnerID != "" && nRegion.OwnerID != a.OwnerID:
				key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(nRegion.OwnerID))
				rel, exists := r.gs.Relations[key]
				if exists && rel.Stance == faction.StanceWar {
					col = color.RGBA{220, 60, 60, 200}
				} else {
					col = color.RGBA{220, 140, 30, 210}
				}
			case nRegion.OwnerID == "":
				col = color.RGBA{60, 220, 60, 200}
			default:
				col = color.RGBA{80, 160, 255, 160}
			}
			// Barış halindeki düşman bölgeye kılıç ikonu
			if nRegion.OwnerID != "" && nRegion.OwnerID != a.OwnerID {
				key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(nRegion.OwnerID))
				rel, exists := r.gs.Relations[key]
				if !exists || rel.Stance != faction.StanceWar {
					DrawTextCentered(screen, "⚔", sx, sy-8, FaceSmall, color.RGBA{255, 200, 80, 230})
				}
			}
		}

		vector.StrokeCircle(screen, float32(sx), float32(sy), 18, 3, col, true)
	}
}

// armyIconPos bir ordunun ekrandaki ikon koordinatlarını tutar.
type armyIconPos struct {
	ArmyID army.ArmyID
	X, Y   float32
}

type settlementDraw struct {
	Region    *world.Region
	Index     int
	Text      string
	X, Y      float64
	W, H      float64
	SX, SY    float64
	DrawLabel bool
}

type screenRect struct {
	X, Y, W, H float64
}

func (r *Renderer) regionScreenPos(region *world.Region) (float64, float64) {
	wx, wy := r.regionWorldPos(region)
	return r.worldToScreen(wx, wy)
}

func (r *Renderer) regionWorldPos(region *world.Region) (float64, float64) {
	if region != nil && region.IsSea {
		if ax, ay, ok := r.worldMap.RegionAnchor(region.ID); ok {
			return float64(ax), float64(ay)
		}
	}
	if region != nil {
		if ax, ay, ok := r.worldMap.PrimarySettlementAnchor(region.ID); ok {
			return float64(ax), float64(ay)
		}
	}
	return wcX(region.WorldX), wcY(region.WorldY)
}

// armyIconPositions tüm orduların ekran koordinatlarını hesaplar.
// Aynı bölgedeki birden fazla ordu yan yana offset'lenir.
func (r *Renderer) armyIconPositions() []armyIconPos {
	const iconStep = float32(26) // ikon genişliği 20 + 6px boşluk

	byRegion := map[world.RegionID][]army.ArmyID{}
	for aid, a := range r.gs.Armies {
		byRegion[a.RegionID] = append(byRegion[a.RegionID], aid)
	}

	r.armyIconBuf = r.armyIconBuf[:0]
	for rid, aids := range byRegion {
		region, ok := r.gs.Regions[rid]
		if !ok {
			continue
		}
		sx, sy := r.regionScreenPos(region)
		baseY := float32(sy) - 22

		sort.Slice(aids, func(i, j int) bool { return aids[i] < aids[j] })

		n := float32(len(aids))
		startX := float32(sx) - (n-1)*iconStep/2
		for i, aid := range aids {
			r.armyIconBuf = append(r.armyIconBuf, armyIconPos{
				ArmyID: aid,
				X:      startX + float32(i)*iconStep,
				Y:      baseY,
			})
		}
	}
	sort.SliceStable(r.armyIconBuf, func(i, j int) bool {
		if r.armyIconBuf[i].Y != r.armyIconBuf[j].Y {
			return r.armyIconBuf[i].Y < r.armyIconBuf[j].Y
		}
		if r.armyIconBuf[i].X != r.armyIconBuf[j].X {
			return r.armyIconBuf[i].X < r.armyIconBuf[j].X
		}
		return r.armyIconBuf[i].ArmyID < r.armyIconBuf[j].ArmyID
	})
	return r.armyIconBuf
}

// drawArmies tüm orduları harita üzerinde çizer.
func (r *Renderer) drawArmies(screen *ebiten.Image, positions []armyIconPos) {
	for _, pos := range positions {
		a, ok := r.gs.Armies[pos.ArmyID]
		if !ok {
			continue
		}
		fc := factionColor(r.gs, a.OwnerID)
		isSelected := pos.ArmyID == r.SelectedArmy
		unitCount := len(a.Units)
		if a.OwnerID != string(r.gs.PlayerFactionID) && !enemyArmyInPlayerMoveRange(r.gs, a) {
			unitCount = -1
		}
		r.drawArmyIcon(screen, pos.X, pos.Y, fc, unitCount, isSelected, a.IsNaval)
	}
}

// drawArmyIcon tek bir ordu ikonunu çizer.
// Kara ordusu → kare, deniz donanması → daire.
func (r *Renderer) drawArmyIcon(screen *ebiten.Image, cx, cy float32, col color.RGBA, unitCount int, selected bool, isNaval bool) {
	borderCol := color.RGBA{200, 200, 200, 220}
	if selected {
		borderCol = color.RGBA{255, 215, 0, 255}
	}

	if isNaval {
		// Dış daire (border) + iç daire (fraksiyon rengi)
		vector.FillCircle(screen, cx, cy, 13, borderCol, false)
		vector.FillCircle(screen, cx, cy, 11, col, false)
	} else {
		// Dış kare (border) + iç kare (fraksiyon rengi)
		half := float32(10)
		vector.FillRect(screen, cx-half-2, cy-half-2, half*2+4, half*2+4, borderCol, false)
		vector.FillRect(screen, cx-half, cy-half, half*2, half*2, col, false)
	}

	// Birim sayısı
	countStr := "?"
	if unitCount >= 0 {
		countStr = itoa(unitCount)
	}
	tw := MeasureText(countStr, FaceSmall)
	DrawText(screen, countStr, float64(cx)-tw/2, float64(cy)-5, FaceSmall, color.RGBA{255, 255, 255, 255})
}

// drawRegionLabels zoom yeterliyse bölgedeki yerleşim noktalarını ve adlarını yazar.
func (r *Renderer) drawRegionLabels(screen *ebiten.Image, armyPositions []armyIconPos) {
	if r.camScale < 0.5 {
		return
	}

	labelCol := color.RGBA{255, 255, 255, 220}
	shadowCol := color.RGBA{0, 0, 0, 160}

	r.regionLabelBuf = r.regionLabelBuf[:0]
	for _, region := range r.gs.Regions {
		if region.IsSea || region.IsLocked {
			continue
		}
		r.appendSettlementDraws(region)
	}

	sort.SliceStable(r.regionLabelBuf, func(i, j int) bool {
		if r.regionLabelBuf[i].SY != r.regionLabelBuf[j].SY {
			return r.regionLabelBuf[i].SY < r.regionLabelBuf[j].SY
		}
		if r.regionLabelBuf[i].SX != r.regionLabelBuf[j].SX {
			return r.regionLabelBuf[i].SX < r.regionLabelBuf[j].SX
		}
		return r.regionLabelBuf[i].Region.ID < r.regionLabelBuf[j].Region.ID
	})

	r.labelRectBuf = r.labelRectBuf[:0]
	for _, item := range r.regionLabelBuf {
		if !item.DrawLabel {
			r.drawCityDot(screen, item.Region, float32(item.SX), float32(item.SY))
			if r.gs.Phase == state.PhaseEditMode && item.Region != nil &&
				item.Region.ID == r.editSelectedRegion && item.Index == r.editSelectedSettlement {
				vector.StrokeCircle(screen, float32(item.SX), float32(item.SY)+4, 10, 2, color.RGBA{255, 190, 45, 230}, true)
			}
			continue
		}

		rect := screenRect{X: item.X, Y: item.Y, W: item.W, H: item.H}
		drawText := true
		for _, used := range r.labelRectBuf {
			if rectIntersects(expandRect(rect, 4), expandRect(used, 4)) {
				drawText = false
				break
			}
		}
		if drawText {
			for _, pos := range armyPositions {
				armyRect := screenRect{X: float64(pos.X) - 15, Y: float64(pos.Y) - 15, W: 30, H: 30}
				if rectIntersects(expandRect(rect, 3), armyRect) {
					drawText = false
					break
				}
			}
		}

		if drawText {
			face := FaceSmall
			if r.camScale >= 1.0 {
				face = FaceMed
			}
			DrawText(screen, item.Text, item.X+1, item.Y+1, face, shadowCol)
			DrawText(screen, item.Text, item.X, item.Y, face, labelCol)
			r.labelRectBuf = append(r.labelRectBuf, rect)
		}

		r.drawCityDot(screen, item.Region, float32(item.SX), float32(item.SY))
		if r.gs.Phase == state.PhaseEditMode && item.Region != nil &&
			item.Region.ID == r.editSelectedRegion && item.Index == r.editSelectedSettlement {
			vector.StrokeCircle(screen, float32(item.SX), float32(item.SY)+4, 10, 2, color.RGBA{255, 190, 45, 230}, true)
		}
	}
}

func (r *Renderer) appendSettlementDraws(region *world.Region) {
	if len(region.Settlements) == 0 {
		sx, sy := r.regionScreenPos(region)
		r.appendSettlementDraw(region, -1, region.NameTR, sx, sy, true)
		return
	}

	for i, settlement := range region.Settlements {
		isPrimary := settlement.IsCapital || i == 0
		if !isPrimary && r.camScale < 0.85 {
			continue
		}

		ax, ay, ok := r.worldMap.SettlementAnchor(region.ID, i)
		if !ok {
			continue
		}
		sx, sy := r.worldToScreen(float64(ax), float64(ay))
		name := settlement.NameTR
		if name == "" {
			name = settlement.Name
		}
		if name == "" {
			name = region.NameTR
		}
		drawLabel := isPrimary || r.camScale >= 1.25
		r.appendSettlementDraw(region, i, name, sx, sy, drawLabel)
	}
}

func (r *Renderer) appendSettlementDraw(region *world.Region, index int, text string, sx, sy float64, drawLabel bool) {
	if sx < -50 || sx > ScreenWidth+50 || sy < -20 || sy > ScreenHeight+20 {
		return
	}

	face := FaceSmall
	if r.camScale >= 1.0 {
		face = FaceMed
	}

	w := MeasureText(text, face)
	lx := sx - w/2
	h := float64(16)
	if face == FaceMed {
		h = 20
	}
	r.regionLabelBuf = append(r.regionLabelBuf, settlementDraw{
		Region:    region,
		Index:     index,
		Text:      text,
		X:         lx,
		Y:         sy - 7,
		W:         w,
		H:         h,
		SX:        sx,
		SY:        sy,
		DrawLabel: drawLabel,
	})
}

func expandRect(r screenRect, pad float64) screenRect {
	return screenRect{X: r.X - pad, Y: r.Y - pad, W: r.W + pad*2, H: r.H + pad*2}
}

func rectIntersects(a, b screenRect) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X && a.Y < b.Y+b.H && a.Y+a.H > b.Y
}

func (r *Renderer) drawEditModeHud(screen *ebiten.Image) {
	const panelW, panelH = float32(620), float32(112)
	x, y := float32(18), float32(18)
	drawRoundedRect(screen, x, y, panelW, panelH, 8, color.RGBA{16, 20, 24, 220})
	drawPanelBorder(screen, x, y, panelW, panelH)

	title := "EDIT MODE"
	if r.editDirty {
		title += " *"
	}
	DrawText(screen, title, float64(x)+14, float64(y)+10, FaceMed, ColorGold)
	DrawText(screen, "Sol: sec/tasi   Alt+sol: ekle   Delete: sil   Shift+sol: bolge merkezi   V: Voronoi   Ctrl+S: kaydet",
		float64(x)+14, float64(y)+36, FaceSmall, ColorWhite)

	info := "Secili: yok"
	if region, ok := r.gs.Regions[r.editSelectedRegion]; ok &&
		r.editSelectedSettlement >= 0 && r.editSelectedSettlement < len(region.Settlements) {
		s := region.Settlements[r.editSelectedSettlement]
		info = region.NameTR + " / " + s.NameTR + "  (" + itoa(s.X) + "," + itoa(s.Y) + ")"
	} else if region, ok := r.gs.Regions[r.editSelectedRegion]; ok && region != nil {
		info = "Merkez: " + region.NameTR + "  (" + itoa(region.WorldX) + "," + itoa(region.WorldY) + ")"
	}
	DrawText(screen, info, float64(x)+14, float64(y)+58, FaceSmall, ColorGray)
	debugState := "Voronoi debug: kapali"
	if r.editVoronoiDebug {
		debugState = "Voronoi debug: acik"
	}
	if r.editRenaming {
		DrawText(screen, "Isim: "+string(r.editNameRunes), float64(x)+14, float64(y)+80, FaceSmall, ColorGold)
	} else {
		DrawText(screen, debugState+"   Esc: ana menu", float64(x)+14, float64(y)+80, FaceSmall, ColorGray)
	}
}

func (r *Renderer) drawEditInspector(screen *ebiten.Image) {
	x, y, w, h := editInspectorRect()
	drawRoundedRect(screen, x, y, w, h, 8, color.RGBA{16, 20, 24, 226})
	drawPanelBorder(screen, x, y, w, h)

	DrawText(screen, "BILGI", float64(x)+14, float64(y)+10, FaceMed, ColorGold)
	ly := float64(y) + 38

	region := r.gs.Regions[r.editSelectedRegion]
	if r.SelectedArmy != "" {
		if a, ok := r.gs.Armies[r.SelectedArmy]; ok && a != nil {
			DrawText(screen, "Ordu: "+string(a.ID), float64(x)+14, ly, FaceSmall, ColorWhite)
			ly += 18
			DrawText(screen, "Bolge: "+string(a.RegionID), float64(x)+14, ly, FaceSmall, ColorGray)
			ly += 18
			DrawText(screen, "Birim: "+itoa(len(a.Units))+" / 20", float64(x)+14, ly, FaceSmall, ColorGray)
			r.drawEditInspectorButtons(screen, nil)
			return
		}
	}

	if region == nil {
		DrawText(screen, "Haritadan bir bolge veya yerlesim sec.", float64(x)+14, ly, FaceSmall, ColorGray)
		r.drawEditInspectorButtons(screen, nil)
		return
	}

	name := region.NameTR
	if name == "" {
		name = region.Name
	}
	DrawText(screen, name, float64(x)+14, ly, FaceSmall, ColorWhite)
	ly += 18
	DrawText(screen, "ID: "+string(region.ID), float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 18
	DrawText(screen, "Sahip: "+region.OwnerID+"   Arazi: "+string(region.Terrain), float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 18
	DrawText(screen, "Merkez: "+itoa(region.WorldX)+","+itoa(region.WorldY)+"   Yerlesim: "+itoa(len(region.Settlements)), float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 22

	if r.hasEditSelection() {
		settlement := region.Settlements[r.editSelectedSettlement]
		sName := settlement.NameTR
		if sName == "" {
			sName = settlement.Name
		}
		DrawText(screen, "Secili yerlesim: "+sName, float64(x)+14, ly, FaceSmall, ColorGold)
		ly += 18
		DrawText(screen, settlement.ID+"  "+string(settlement.Type)+"  "+itoa(settlement.X)+","+itoa(settlement.Y),
			float64(x)+14, ly, FaceSmall, ColorGray)
		if settlement.IsCapital {
			ly += 18
			DrawText(screen, "Ana yerlesim", float64(x)+14, ly, FaceSmall, ColorGray)
		}
	} else {
		DrawText(screen, "Yerlesim secili degil.", float64(x)+14, ly, FaceSmall, ColorGray)
	}

	r.drawEditInspectorButtons(screen, region)
	r.editOwnerDropdown.Draw(screen)
	r.editTerrainDropdown.Draw(screen)
	r.editSettlementTypeDropdown.Draw(screen)
}

func (r *Renderer) drawEditInspectorButtons(screen *ebiten.Image, region *world.Region) {
	canAdd := region != nil && !region.IsSea
	canRegion := region != nil && !region.IsSea
	canSettlement := r.hasEditSelection()
	drawEditInspectorButton(screen, editButtonAddSettlement, "Yerlesim Ekle", canAdd)
	drawEditInspectorButton(screen, editButtonSettlementType, "Tip", canSettlement)
	drawEditInspectorButton(screen, editButtonSetCapitalSettlement, "Ana Yap", canSettlement)
	drawEditInspectorButton(screen, editButtonRenameSettlement, "Isim", canSettlement)
	drawEditInspectorButton(screen, editButtonRegionTerrain, "Arazi", canRegion)
	drawEditInspectorButton(screen, editButtonRegionOwner, "Sahip", canRegion)
	drawEditInspectorButton(screen, editButtonDeleteSettlement, "Sil", canSettlement)
	drawEditInspectorButton(screen, editButtonSaveScenario, "Kaydet", true)
}

func drawEditInspectorButton(screen *ebiten.Image, kind editInspectorButton, label string, active bool) {
	rect := editInspectorButtonRect(kind)
	drawTinyPanelButton(screen, float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3]), label, active)
}

type editInspectorButton int

const (
	editButtonNone editInspectorButton = iota
	editButtonAddSettlement
	editButtonSettlementType
	editButtonSetCapitalSettlement
	editButtonRenameSettlement
	editButtonRegionTerrain
	editButtonRegionOwner
	editButtonDeleteSettlement
	editButtonSaveScenario
)

func editInspectorRect() (float32, float32, float32, float32) {
	const w, h = float32(340), float32(282)
	return 18, float32(ScreenHeight) - h - 18, w, h
}

func editInspectorHit(mx, my float64) bool {
	x, y, w, h := editInspectorRect()
	return mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h)
}

func editInspectorButtonRect(kind editInspectorButton) uiRect {
	x, y, _, h := editInspectorRect()
	const bw, bh, gap = float64(145), float64(26), float64(8)
	left := float64(x) + 14
	right := left + bw + gap
	row1 := float64(y) + float64(h) - 136
	row2 := row1 + bh + gap
	row3 := row2 + bh + gap
	row4 := row3 + bh + gap
	switch kind {
	case editButtonAddSettlement:
		return uiRect{left, row1, bw, bh}
	case editButtonSettlementType:
		return uiRect{right, row1, bw, bh}
	case editButtonSetCapitalSettlement:
		return uiRect{left, row2, bw, bh}
	case editButtonRenameSettlement:
		return uiRect{right, row2, bw, bh}
	case editButtonRegionTerrain:
		return uiRect{left, row3, bw, bh}
	case editButtonRegionOwner:
		return uiRect{right, row3, bw, bh}
	case editButtonDeleteSettlement:
		return uiRect{left, row4, bw, bh}
	case editButtonSaveScenario:
		return uiRect{right, row4, bw, bh}
	default:
		return uiRect{}
	}
}

func editInspectorButtonAt(mx, my float64) editInspectorButton {
	for kind := editButtonAddSettlement; kind <= editButtonSaveScenario; kind++ {
		if uiRectHit(mx, my, editInspectorButtonRect(kind)) {
			return kind
		}
	}
	return editButtonNone
}

const (
	editOwnerDropdownVisibleRows = 10
	editOwnerDropdownRowH        = float32(24)
	editOwnerDropdownHeaderH     = float32(30)
)

func editOwnerDropdownRect() (float32, float32, float32, float32) {
	x, y, w, _ := editInspectorRect()
	dropW := float32(292)
	dropH := editOwnerDropdownHeaderH + editOwnerDropdownRowH*editOwnerDropdownVisibleRows + 10
	return x + w + 8, y, dropW, dropH
}

func editTerrainDropdownRect() (float32, float32, float32, float32) {
	x, y, w, _ := editInspectorRect()
	dropW := float32(292)
	dropH := editOwnerDropdownHeaderH + editOwnerDropdownRowH*editOwnerDropdownVisibleRows + 10
	return x + w + 8, y, dropW, dropH
}

func editSettlementTypeDropdownRect() (float32, float32, float32, float32) {
	x, y, w, _ := editInspectorRect()
	dropW := float32(292)
	dropH := editOwnerDropdownHeaderH + editOwnerDropdownRowH*editOwnerDropdownVisibleRows + 10
	return x + w + 8, y, dropW, dropH
}

func (r *Renderer) updateEditDropdownPositions() {
	dx, dy, _, _ := editOwnerDropdownRect()
	r.editOwnerDropdown.SetPosition(dx, dy)
	r.editTerrainDropdown.SetPosition(dx, dy)
	r.editSettlementTypeDropdown.SetPosition(dx, dy)
}

func editMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (r *Renderer) drawEditRegionCenters(screen *ebiten.Image) {
	for _, region := range r.gs.Regions {
		if region == nil || region.IsSea || region.IsLocked {
			continue
		}
		sx, sy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
		col := color.RGBA{80, 220, 255, 190}
		if region.ID == r.editSelectedRegion && r.editSelectedSettlement < 0 {
			col = color.RGBA{255, 190, 45, 240}
		}
		x, y := float32(sx), float32(sy)
		vector.StrokeCircle(screen, x, y, 6, 1.5, col, true)
		vector.StrokeLine(screen, x-8, y, x+8, y, 1.5, col, true)
		vector.StrokeLine(screen, x, y-8, x, y+8, 1.5, col, true)
	}
}

func (r *Renderer) drawEditVoronoiDebug(screen *ebiten.Image) {
	if !r.editVoronoiDebug {
		return
	}
	rid := r.editSelectedRegion
	if rid == "" {
		mx, my := ebiten.CursorPosition()
		rid = r.editRegionAt(float64(mx), float64(my))
	}
	region := r.gs.Regions[rid]
	if region == nil || region.IsSea {
		r.drawEditVoronoiLegend(screen, "", nil)
		return
	}

	r.editVisualNeighborBuf = r.worldMap.VisualNeighbors(rid, r.editVisualNeighborBuf)
	r.editBoundaryPixelBuf = r.worldMap.BoundaryPixels(rid, r.editBoundaryPixelBuf)
	r.drawEditVoronoiBoundary(screen, r.editBoundaryPixelBuf)

	cx, cy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
	for _, nrid := range r.editVisualNeighborBuf {
		neighbor := r.gs.Regions[nrid]
		if neighbor == nil || neighbor.IsSea {
			continue
		}
		nx, ny := r.worldToScreen(wcX(neighbor.WorldX), wcY(neighbor.WorldY))
		col := color.RGBA{90, 220, 125, 205}
		if !regionHasNeighbor(region, nrid) {
			col = color.RGBA{235, 80, 80, 220}
		}
		vector.StrokeLine(screen, float32(cx), float32(cy), float32(nx), float32(ny), 1.5, col, true)
		mx, my := (cx+nx)/2, (cy+ny)/2
		vector.FillRect(screen, float32(mx)-3, float32(my)-3, 6, 6, col, true)
	}

	for _, nrid := range region.Neighbors {
		if visualNeighborContains(r.editVisualNeighborBuf, nrid) {
			continue
		}
		neighbor := r.gs.Regions[nrid]
		if neighbor == nil || neighbor.IsSea {
			continue
		}
		nx, ny := r.worldToScreen(wcX(neighbor.WorldX), wcY(neighbor.WorldY))
		col := color.RGBA{180, 180, 180, 150}
		vector.StrokeLine(screen, float32(cx), float32(cy), float32(nx), float32(ny), 1, col, true)
	}

	vector.StrokeCircle(screen, float32(cx), float32(cy), 12, 2.5, color.RGBA{255, 220, 70, 245}, true)
	r.drawEditVoronoiLegend(screen, rid, r.editVisualNeighborBuf)
}

func (r *Renderer) drawEditVoronoiBoundary(screen *ebiten.Image, pixels []int) {
	step := 1
	if r.camScale < 0.8 {
		step = 2
	}
	if r.camScale < 0.45 {
		step = 4
	}
	size := float32(2)
	if r.camScale >= 1.25 {
		size = 3
	}
	col := color.RGBA{80, 210, 255, 215}
	for i := 0; i < len(pixels); i += step {
		pIdx := pixels[i]
		wx := float64(pIdx % WorldW)
		wy := float64(pIdx / WorldW)
		sx, sy := r.worldToScreen(wx, wy)
		if sx < -4 || sx > ScreenWidth+4 || sy < -4 || sy > ScreenHeight+4 {
			continue
		}
		vector.FillRect(screen, float32(sx)-size/2, float32(sy)-size/2, size, size, col, true)
	}
}

func (r *Renderer) drawEditVoronoiLegend(screen *ebiten.Image, rid world.RegionID, visual []world.RegionID) {
	const panelW, panelH = float32(360), float32(104)
	x := float32(ScreenWidth) - panelW - 18
	y := float32(18)
	drawRoundedRect(screen, x, y, panelW, panelH, 8, color.RGBA{16, 20, 24, 218})
	drawPanelBorder(screen, x, y, panelW, panelH)
	DrawText(screen, "VORONOI DEBUG", float64(x)+12, float64(y)+10, FaceSmall, ColorGold)
	DrawText(screen, "camgobegi: raster sinir", float64(x)+12, float64(y)+31, FaceSmall, ColorGray)
	DrawText(screen, "yesil: gorunen+JSON   kirmizi: sadece gorunen", float64(x)+12, float64(y)+48, FaceSmall, ColorGray)

	mx, my := ebiten.CursorPosition()
	wx, wy := r.screenToWorld(float64(mx), float64(my))
	hover := r.worldMap.RegionAt(int(wx), int(wy))
	sx, sy := scenarioCoordsFromWorld(wx, wy)
	DrawText(screen, "Hover: "+string(hover)+"  "+itoa(sx)+","+itoa(sy), float64(x)+12, float64(y)+68, FaceSmall, ColorWhite)
	if rid != "" {
		region := r.gs.Regions[rid]
		jsonCount := 0
		if region != nil {
			jsonCount = len(region.Neighbors)
		}
		DrawText(screen, "Secili: "+string(rid)+"  visual/json: "+itoa(len(visual))+"/"+itoa(jsonCount),
			float64(x)+12, float64(y)+85, FaceSmall, ColorWhite)
	}
}

func regionHasNeighbor(region *world.Region, rid world.RegionID) bool {
	for _, nrid := range region.Neighbors {
		if nrid == rid {
			return true
		}
	}
	return false
}

func visualNeighborContains(neighbors []world.RegionID, rid world.RegionID) bool {
	for _, nrid := range neighbors {
		if nrid == rid {
			return true
		}
	}
	return false
}

func (r *Renderer) handleEditModeInput() InputAction {
	if r.editRenaming {
		return r.handleEditRenameInput()
	}

	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	leftJustPressed := r.mouseJustPressed(ebiten.MouseButtonLeft)

	if r.editOwnerDropdown.IsOpen() {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 && r.editOwnerDropdown.HitTest(fx, fy) {
			r.editOwnerDropdown.Scroll(wheelY)
			return InputAction{}
		}
	}

	if r.editTerrainDropdown.IsOpen() {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 && r.editTerrainDropdown.HitTest(fx, fy) {
			r.editTerrainDropdown.Scroll(wheelY)
			return InputAction{}
		}
	}

	if r.editSettlementTypeDropdown.IsOpen() {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 && r.editSettlementTypeDropdown.HitTest(fx, fy) {
			r.editSettlementTypeDropdown.Scroll(wheelY)
			return InputAction{}
		}
	}

	if !r.editOwnerDropdown.IsOpen() && !r.editTerrainDropdown.IsOpen() && !r.editSettlementTypeDropdown.IsOpen() {
		r.handleCamera()
	}

	if r.keyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if r.keyJustPressed(ebiten.KeyV) {
		r.editVoronoiDebug = !r.editVoronoiDebug
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.editOwnerDropdown.Close()
		r.editTerrainDropdown.Close()
		r.editSettlementTypeDropdown.Close()
		if r.editDirty {
			r.showEditExitConfirm()
			return InputAction{}
		}
		return InputAction{Kind: ActionGoMainMenu}
	}
	if r.keyJustPressed(ebiten.KeyS) && (ebiten.IsKeyPressed(ebiten.KeyControl) ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)) {
		return InputAction{Kind: ActionSaveScenario}
	}
	if r.keyJustPressed(ebiten.KeyDelete) && r.hasEditSelection() {
		r.deleteSelectedSettlement()
		return InputAction{}
	}
	if (r.keyJustPressed(ebiten.KeyF2) || r.keyJustPressed(ebiten.KeyEnter)) && r.hasEditSelection() {
		r.beginEditRename()
		return InputAction{}
	}

	if leftJustPressed {
		if action, ok := r.handleEditInspectorClick(fx, fy); ok {
			return action
		}
	}

	if r.editDraggingRegion && !leftPressed {
		r.editDraggingRegion = false
		r.rebuildEditWorldMap()
	}

	if leftJustPressed {
		if editModifierPressed() {
			rid := r.editRegionAt(fx, fy)
			if rid != "" {
				r.editOwnerDropdown.Close()
				r.editTerrainDropdown.Close()
				r.editSettlementTypeDropdown.Close()
				r.editSelectedRegion = rid
				r.editSelectedSettlement = -1
				r.editDraggingRegion = true
				r.editDraggingSettlement = false
				r.editRenaming = false
				r.moveSelectedRegionCenterTo(fx, fy)
				return InputAction{}
			}
		}
		if editAddModifierPressed() {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.addSettlementAt(fx, fy)
			return InputAction{}
		}

		if aid, ok := r.editArmyAt(fx, fy); ok {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.SelectedArmy = aid
			if a := r.gs.Armies[aid]; a != nil {
				r.editSelectedRegion = a.RegionID
			}
			r.editSelectedSettlement = -1
			r.editDraggingSettlement = false
			r.editDraggingRegion = false
			r.editRenaming = false
			return InputAction{}
		}

		rid, idx, ok := r.editSettlementAt(fx, fy)
		if ok {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.SelectedArmy = ""
			r.editSelectedRegion = rid
			r.editSelectedSettlement = idx
			r.editDraggingSettlement = true
			r.editDraggingRegion = false
			return InputAction{}
		}
		if rid := r.editRegionAt(fx, fy); rid != "" {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.SelectedArmy = ""
			r.editSelectedRegion = rid
			r.editSelectedSettlement = -1
			r.editRenaming = false
			r.editDraggingRegion = false
			r.editDraggingSettlement = false
			return InputAction{}
		}
		r.editOwnerDropdown.Close()
		r.editTerrainDropdown.Close()
		r.editSettlementTypeDropdown.Close()
		r.SelectedArmy = ""
		r.editSelectedRegion = ""
		r.editSelectedSettlement = -1
		r.editRenaming = false
		r.editDraggingRegion = false
	}

	if !leftPressed {
		r.editDraggingSettlement = false
	}

	if r.editDraggingRegion {
		r.moveSelectedRegionCenterTo(fx, fy)
		return InputAction{}
	}

	if r.editDraggingSettlement {
		r.moveSelectedSettlementTo(fx, fy)
	}

	return InputAction{}
}

func (r *Renderer) handleEditInspectorClick(fx, fy float64) (InputAction, bool) {
	if r.editOwnerDropdown.IsOpen() {
		if idx, ok := r.editOwnerDropdown.GetSelectedOption(fx, fy); ok {
			r.setSelectedRegionOwner(r.editOwnerDropdown.options[idx])
			r.editOwnerDropdown.Close()
			return InputAction{}, true
		}
		if r.editOwnerDropdown.HitTest(fx, fy) {
			return InputAction{}, true
		}
		if !editInspectorHit(fx, fy) {
			r.editOwnerDropdown.Close()
			return InputAction{}, false
		}
	}
	if r.editTerrainDropdown.IsOpen() {
		if idx, ok := r.editTerrainDropdown.GetSelectedOption(fx, fy); ok {
			r.setSelectedRegionTerrain(world.TerrainType(r.editTerrainDropdown.options[idx]))
			r.editTerrainDropdown.Close()
			return InputAction{}, true
		}
		if r.editTerrainDropdown.HitTest(fx, fy) {
			return InputAction{}, true
		}
		if !editInspectorHit(fx, fy) {
			r.editTerrainDropdown.Close()
			return InputAction{}, false
		}
	}
	if r.editSettlementTypeDropdown.IsOpen() {
		if idx, ok := r.editSettlementTypeDropdown.GetSelectedOption(fx, fy); ok {
			r.setSelectedSettlementType(r.editSettlementTypeDropdown.options[idx])
			r.editSettlementTypeDropdown.Close()
			return InputAction{}, true
		}
		if r.editSettlementTypeDropdown.HitTest(fx, fy) {
			return InputAction{}, true
		}
		if !editInspectorHit(fx, fy) {
			r.editSettlementTypeDropdown.Close()
			return InputAction{}, false
		}
	}
	if !editInspectorHit(fx, fy) {
		return InputAction{}, false
	}
	switch editInspectorButtonAt(fx, fy) {
	case editButtonAddSettlement:
		r.addSettlementToSelectedRegion()
	case editButtonSettlementType:
		if r.hasEditSelection() {
			r.toggleEditSettlementTypeDropdown()
		}
	case editButtonSetCapitalSettlement:
		if r.hasEditSelection() {
			r.setSelectedSettlementCapital()
		}
	case editButtonRenameSettlement:
		if r.hasEditSelection() {
			r.beginEditRename()
		}
	case editButtonRegionTerrain:
		r.toggleEditTerrainDropdown()
	case editButtonRegionOwner:
		r.toggleEditOwnerDropdown()
	case editButtonDeleteSettlement:
		if r.hasEditSelection() {
			r.deleteSelectedSettlement()
		}
	case editButtonSaveScenario:
		return InputAction{Kind: ActionSaveScenario}, true
	}
	return InputAction{}, true
}

func (r *Renderer) toggleEditOwnerDropdown() {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || region.IsSea {
		r.editOwnerDropdown.Close()
		return
	}

	// Position dropdown below the owner button
	buttonRect := editInspectorButtonRect(editButtonRegionOwner)
	dropX := float32(buttonRect[0])
	dropY := float32(buttonRect[1] + buttonRect[3] + 4) // Below the button with small gap

	r.editOwnerDropdown.SetPosition(dropX, dropY)
	r.editOwnerDropdown.SetOptions(editOwnerOptions(r.gs.Factions), region.OwnerID)
	r.editOwnerDropdown.Toggle()
}

func (r *Renderer) toggleEditTerrainDropdown() {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || region.IsSea {
		r.editTerrainDropdown.Close()
		return
	}

	// Position dropdown below the terrain button
	buttonRect := editInspectorButtonRect(editButtonRegionTerrain)
	dropX := float32(buttonRect[0])
	dropY := float32(buttonRect[1] + buttonRect[3] + 4) // Below the button with small gap

	r.editTerrainDropdown.SetPosition(dropX, dropY)
	terrainOptions := editTerrainOptions()
	stringOptions := make([]string, len(terrainOptions))
	for i, t := range terrainOptions {
		stringOptions[i] = string(t)
	}
	r.editTerrainDropdown.SetOptions(stringOptions, string(region.Terrain))
	r.editTerrainDropdown.Toggle()
}

func (r *Renderer) toggleEditSettlementTypeDropdown() {
	if !r.hasEditSelection() {
		r.editSettlementTypeDropdown.Close()
		return
	}

	// Position dropdown below the settlement type button
	buttonRect := editInspectorButtonRect(editButtonSettlementType)
	dropX := float32(buttonRect[0])
	dropY := float32(buttonRect[1] + buttonRect[3] + 4) // Below the button with small gap

	region := r.gs.Regions[r.editSelectedRegion]
	settlement := region.Settlements[r.editSelectedSettlement]
	r.editSettlementTypeDropdown.SetPosition(dropX, dropY)
	r.editSettlementTypeDropdown.SetOptions(world.AllSettlementTypes(), string(settlement.Type))
	r.editSettlementTypeDropdown.Toggle()
}

func (r *Renderer) hasEditSelection() bool {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	return ok && region != nil && r.editSelectedSettlement >= 0 &&
		r.editSelectedSettlement < len(region.Settlements)
}

func (r *Renderer) beginEditRename() {
	region := r.gs.Regions[r.editSelectedRegion]
	name := region.Settlements[r.editSelectedSettlement].NameTR
	if name == "" {
		name = region.Settlements[r.editSelectedSettlement].Name
	}
	r.editNameRunes = append(r.editNameRunes[:0], []rune(name)...)
	r.editRenaming = true
	r.editDraggingSettlement = false
}

func (r *Renderer) handleEditRenameInput() InputAction {
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.editRenaming = false
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		r.commitEditRename()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyBackspace) && len(r.editNameRunes) > 0 {
		r.editNameRunes = r.editNameRunes[:len(r.editNameRunes)-1]
	}
	r.editNameRunes = ebiten.AppendInputChars(r.editNameRunes)
	if len(r.editNameRunes) > 48 {
		r.editNameRunes = r.editNameRunes[:48]
	}
	return InputAction{}
}

func (r *Renderer) commitEditRename() {
	if !r.hasEditSelection() {
		r.editRenaming = false
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
	newName := string(r.editNameRunes)
	if newName != "" && region.Settlements[r.editSelectedSettlement].NameTR != newName {
		region.Settlements[r.editSelectedSettlement].NameTR = newName
		r.editDirty = true
	}
	r.editRenaming = false
}

func (r *Renderer) editSettlementAt(fx, fy float64) (world.RegionID, int, bool) {
	bestDist := float64(14 * 14)
	var bestRegion world.RegionID
	bestIndex := -1
	for rid, region := range r.gs.Regions {
		if region == nil || region.IsSea {
			continue
		}
		for i := range region.Settlements {
			ax, ay, ok := r.worldMap.SettlementAnchor(rid, i)
			if !ok {
				continue
			}
			sx, sy := r.worldToScreen(float64(ax), float64(ay))
			dx, dy := fx-sx, fy-sy
			dist := dx*dx + dy*dy
			if dist <= bestDist {
				bestDist = dist
				bestRegion = rid
				bestIndex = i
			}
		}
	}
	return bestRegion, bestIndex, bestIndex >= 0
}

func (r *Renderer) editRegionAt(fx, fy float64) world.RegionID {
	wx, wy := r.screenToWorld(fx, fy)
	rid := r.worldMap.RegionAt(int(wx), int(wy))
	if region, ok := r.gs.Regions[rid]; ok && region != nil && !region.IsSea {
		return rid
	}
	return ""
}

func (r *Renderer) editArmyAt(fx, fy float64) (army.ArmyID, bool) {
	armyPositions := r.armyIconPositions()
	for i := len(armyPositions) - 1; i >= 0; i-- {
		pos := armyPositions[i]
		dx := fx - float64(pos.X)
		dy := fy - float64(pos.Y)
		if dx*dx+dy*dy < 14*14 {
			return pos.ArmyID, true
		}
	}
	return "", false
}

func (r *Renderer) moveSelectedSettlementTo(fx, fy float64) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || r.editSelectedSettlement < 0 ||
		r.editSelectedSettlement >= len(region.Settlements) {
		return
	}
	wx, wy := r.screenToWorld(fx, fy)
	newX, newY := scenarioCoordsFromWorld(wx, wy)
	targetRegionID := r.worldMap.RegionAt(int(wx), int(wy))
	if targetRegion, ok := r.gs.Regions[targetRegionID]; ok && targetRegion != nil &&
		!targetRegion.IsSea && targetRegion.ID != region.ID {
		r.transferSelectedSettlement(targetRegion.ID, newX, newY)
		return
	}
	region.Settlements[r.editSelectedSettlement].X = newX
	region.Settlements[r.editSelectedSettlement].Y = newY
	r.worldMap.UpdateSettlementAnchor(r.gs, r.editSelectedRegion, r.editSelectedSettlement)
	r.editDirty = true
}

func (r *Renderer) moveSelectedRegionCenterTo(fx, fy float64) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || region.IsSea {
		return
	}
	wx, wy := r.screenToWorld(fx, fy)
	newX, newY := scenarioCoordsFromWorld(wx, wy)
	if region.WorldX == newX && region.WorldY == newY {
		return
	}
	region.WorldX = newX
	region.WorldY = newY
	r.editDirty = true
}

func (r *Renderer) addSettlementAt(fx, fy float64) {
	wx, wy := r.screenToWorld(fx, fy)
	rid := r.worldMap.RegionAt(int(wx), int(wy))
	x, y := scenarioCoordsFromWorld(wx, wy)
	r.addSettlement(rid, x, y)
}

func (r *Renderer) addSettlementToSelectedRegion() {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || region.IsSea {
		return
	}
	r.addSettlement(region.ID, region.WorldX, region.WorldY)
}

func (r *Renderer) addSettlement(rid world.RegionID, x, y int) {
	region, ok := r.gs.Regions[rid]
	if !ok || region == nil || region.IsSea {
		return
	}

	name := region.NameTR
	if name == "" {
		name = region.Name
	}
	if len(region.Settlements) > 0 {
		name += " " + itoa(len(region.Settlements)+1)
	}
	settlement := world.Settlement{
		ID:        nextSettlementID(region),
		NameTR:    name,
		X:         x,
		Y:         y,
		Type:      "city",
		IsCapital: len(region.Settlements) == 0,
	}
	region.Settlements = append(region.Settlements, settlement)
	r.editSelectedRegion = rid
	r.editSelectedSettlement = len(region.Settlements) - 1
	r.editDraggingSettlement = false
	r.editDraggingRegion = false
	r.worldMap.UpdateSettlementAnchor(r.gs, rid, r.editSelectedSettlement)
	r.editDirty = true
}

func (r *Renderer) deleteSelectedSettlement() {
	if !r.hasEditSelection() {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
	removedCapital := region.Settlements[r.editSelectedSettlement].IsCapital
	region.Settlements = append(region.Settlements[:r.editSelectedSettlement], region.Settlements[r.editSelectedSettlement+1:]...)
	if removedCapital {
		ensurePrimarySettlement(region)
	}
	r.editSelectedSettlement = -1
	r.editDraggingSettlement = false
	r.editDraggingRegion = false
	r.worldMap.RebuildSettlementAnchors(r.gs)
	r.editDirty = true
}

func (r *Renderer) setSelectedSettlementCapital() {
	if !r.hasEditSelection() {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
	changed := false
	for i := range region.Settlements {
		isCapital := i == r.editSelectedSettlement
		if region.Settlements[i].IsCapital != isCapital {
			region.Settlements[i].IsCapital = isCapital
			changed = true
		}
	}
	if changed {
		r.worldMap.RebuildSettlementAnchors(r.gs)
		r.editDirty = true
	}
}

func (r *Renderer) setSelectedRegionTerrain(terrain world.TerrainType) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || region.IsSea {
		return
	}
	if region.Terrain == terrain {
		return
	}
	region.Terrain = terrain
	r.editDirty = true
}

func (r *Renderer) setSelectedSettlementType(typ string) {
	if !r.hasEditSelection() {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
	settlement := &region.Settlements[r.editSelectedSettlement]
	st := world.SettlementType(typ)
	if settlement.Type == st {
		return
	}
	settlement.Type = st
	r.editDirty = true
}

func (r *Renderer) setSelectedRegionOwner(ownerID string) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || region.IsSea {
		return
	}
	if region.OwnerID == ownerID {
		return
	}
	region.OwnerID = ownerID
	r.worldMap.MarkDirty()
	r.editDirty = true
}

func (r *Renderer) rebuildEditWorldMap() {
	r.worldMap = NewWorldMap(r.gs)
}

func scenarioCoordsFromWorld(wx, wy float64) (int, int) {
	return int((wx-shapeOffX)/shapeScaleX + 0.5), int((wy-shapeOffY)/shapeScaleY + 0.5)
}

func editModifierPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyShift) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftRight)
}

func editAddModifierPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyAlt) ||
		ebiten.IsKeyPressed(ebiten.KeyAltLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyAltRight)
}

func editOwnerOptions(factions map[faction.FactionID]*faction.Faction) []string {
	ids := make([]string, 0, len(factions)+1)
	ids = append(ids, "")
	for fid := range factions {
		ids = append(ids, string(fid))
	}
	sort.Strings(ids[1:])
	return ids
}

func editTerrainOptions() []world.TerrainType {
	return []world.TerrainType{
		world.TerrainPlain,
		world.TerrainForest,
		world.TerrainMountain,
		world.TerrainPass,
		world.TerrainCoast,
	}
}

func (r *Renderer) editOwnerLabel(ownerID string) string {
	if ownerID == "" {
		return "(sahipsiz)"
	}
	if f, ok := r.gs.Factions[faction.FactionID(ownerID)]; ok && f != nil {
		name := f.NameTR
		if name == "" {
			name = f.Name
		}
		if name != "" {
			return name + "  [" + ownerID + "]"
		}
	}
	return ownerID
}

func nextSettlementID(region *world.Region) string {
	base := string(region.ID) + "_settlement_"
	for n := len(region.Settlements) + 1; ; n++ {
		id := base + itoa(n)
		used := false
		for _, settlement := range region.Settlements {
			if settlement.ID == id {
				used = true
				break
			}
		}
		if !used {
			return id
		}
	}
}

func (r *Renderer) transferSelectedSettlement(targetID world.RegionID, x, y int) {
	source := r.gs.Regions[r.editSelectedRegion]
	target := r.gs.Regions[targetID]
	if source == nil || target == nil || r.editSelectedSettlement < 0 ||
		r.editSelectedSettlement >= len(source.Settlements) {
		return
	}

	settlement := source.Settlements[r.editSelectedSettlement]
	settlement.X = x
	settlement.Y = y
	source.Settlements = append(source.Settlements[:r.editSelectedSettlement], source.Settlements[r.editSelectedSettlement+1:]...)

	if settlement.IsCapital {
		settlement.IsCapital = false
		ensurePrimarySettlement(source)
	}
	if !hasCapitalSettlement(target) {
		settlement.IsCapital = true
	}

	target.Settlements = append(target.Settlements, settlement)
	r.editSelectedRegion = targetID
	r.editSelectedSettlement = len(target.Settlements) - 1
	r.worldMap.RebuildSettlementAnchors(r.gs)
	r.editDirty = true
}

func hasCapitalSettlement(region *world.Region) bool {
	for _, settlement := range region.Settlements {
		if settlement.IsCapital {
			return true
		}
	}
	return false
}

func ensurePrimarySettlement(region *world.Region) {
	if region == nil || len(region.Settlements) == 0 || hasCapitalSettlement(region) {
		return
	}
	region.Settlements[0].IsCapital = true
}

// drawCityDot bölge merkezine küçük iyon çizer.
func (r *Renderer) drawCityDot(screen *ebiten.Image, region *world.Region, sx, sy float32) {
	outerR := float32(4)
	innerR := float32(2.5)

	outerCol := color.RGBA{220, 220, 220, 200}
	if region.OwnerID != "" {
		for fid, f := range r.gs.Factions {
			if string(fid) == region.OwnerID {
				outerCol = color.RGBA{f.Color[0], f.Color[1], f.Color[2], 230}
				break
			}
		}
	}

	vector.FillCircle(screen, sx, sy+4, outerR, outerCol, true)
	vector.FillCircle(screen, sx, sy+4, innerR, color.RGBA{240, 230, 200, 255}, true)
}

// --- Input ---

// HandleInput kamera ve oyun girişlerini işler, InputAction döner.
func (r *Renderer) HandleInput() InputAction {
	r.updateCursorShape()
	r.updateEditDropdownPositions()

	// Onay diyaloğu açıkken normal input engellenir
	if r.confirmDialog.show {
		return r.handleConfirmDialogInput()
	}
	if r.warConfirm.show {
		return r.handleWarConfirmInput()
	}

	// Oyun sonu ekranı inputu
	if r.gs.Phase == state.PhaseGameOver {
		if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyEnter) ||
			r.mouseJustPressed(ebiten.MouseButtonLeft) {
			return InputAction{Kind: ActionBack}
		}
		return InputAction{}
	}

	// Tarihsel olay popup açıkken her tuş/tık kapatır
	if r.showHistoricalEvent {
		if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyEnter) ||
			r.keyJustPressed(ebiten.KeySpace) || r.mouseJustPressed(ebiten.MouseButtonLeft) {
			r.showHistoricalEvent = false
		}
		return InputAction{}
	}

	if r.eventDetail != "" {
		mx, my := ebiten.CursorPosition()
		if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyEnter) ||
			r.keyJustPressed(ebiten.KeySpace) || (r.mouseJustPressed(ebiten.MouseButtonLeft) &&
			(eventDetailCloseHit(float64(mx), float64(my)) || !eventDetailPopupHit(float64(mx), float64(my)))) {
			r.eventDetail = ""
		}
		return InputAction{}
	}

	// Ana menü inputu
	if r.gs.Phase == state.PhaseMainMenu {
		return r.handleMainMenuInput(r.HasSave, r.HasAutoSave)
	}

	// Ayarlar ekranı inputu
	if r.gs.Phase == state.PhaseSettings {
		return r.handleSettingsInput(&r.CurrentSettings)
	}

	// Senaryo seçim ekranı inputu
	if r.gs.Phase == state.PhaseScenarioSelect {
		return r.handleScenarioSelectInput()
	}

	// Fraksiyon seçim ekranı inputu
	if r.gs.Phase == "faction_select" {
		return r.handleFactionSelectInput()
	}

	// Zafer koşulu seçim ekranı inputu
	if r.gs.Phase == "victory_select" {
		return r.handleVictorySelectInput()
	}

	// Duraklama menüsü inputu
	if r.gs.Phase == state.PhasePauseMenu {
		return r.handlePauseMenuInput()
	}

	// Kayıt seçim ekranları inputu
	if r.gs.Phase == state.PhaseLoadSelect {
		return r.handleSlotSelectInput(false)
	}
	if r.gs.Phase == state.PhaseSaveSelect {
		return r.handleSlotSelectInput(true)
	}
	if r.gs.Phase == state.PhaseEditMode {
		return r.handleEditModeInput()
	}

	// Diplomasi paneli açıkken ayrı input
	if r.showDiplomacy {
		return r.handleDiplomacyInput()
	}

	r.handleCamera()

	if r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace) {
		return InputAction{Kind: ActionEndTurn}
	}
	if r.keyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		if r.SelectedArmy != "" || r.SelectedRegion != "" || r.showDiplomacy || r.showTech {
			r.SelectedArmy = ""
			r.SelectedRegion = ""
			r.showDiplomacy = false
			r.showTech = false
		} else {
			r.pauseCursor = 0
			return InputAction{Kind: ActionOpenPauseMenu}
		}
	}
	if r.keyJustPressed(ebiten.KeyTab) {
		r.showDiplomacy = true
		r.diplomacyFocus = 0
		return InputAction{}
	}
	// T: teknoloji paneli
	if r.keyJustPressed(ebiten.KeyT) {
		r.showTech = !r.showTech
		r.techCursor = 0
		return InputAction{}
	}
	// Tech panel aktifken girişi yönlendir
	if r.showTech {
		if f := r.gs.Factions[r.gs.PlayerFactionID]; f != nil {
			return r.handleTechInput(f)
		}
		return InputAction{}
	}
	// R: birlik al, N: gemi inşa et
	if r.keyJustPressed(ebiten.KeyR) && r.SelectedRegion != "" {
		return InputAction{Kind: ActionRecruitUnit, TargetRegion: r.SelectedRegion}
	}
	if r.keyJustPressed(ebiten.KeyN) && r.SelectedRegion != "" {
		return InputAction{Kind: ActionRecruitNaval, TargetRegion: r.SelectedRegion}
	}
	// B: bina inşa et (1–6 tuşları ile seçim)
	if r.SelectedRegion != "" {
		if act := r.handleBuildKey(); act.Kind != ActionNone {
			return act
		}
	}
	// S: kaydet, L: yükle
	if r.keyJustPressed(ebiten.KeyS) {
		return InputAction{Kind: ActionSave}
	}
	if r.keyJustPressed(ebiten.KeyL) {
		return InputAction{Kind: ActionLoad}
	}
	// Vergi ayarlama: seçili kendi bölgesinde . ve , tuşları
	if r.SelectedRegion != "" {
		if r.keyJustPressed(ebiten.KeyPeriod) {
			return InputAction{Kind: ActionAdjustTax, TargetRegion: r.SelectedRegion, Delta: 5}
		}
		if r.keyJustPressed(ebiten.KeyComma) {
			return InputAction{Kind: ActionAdjustTax, TargetRegion: r.SelectedRegion, Delta: -5}
		}
	}

	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		return r.handleLeftClick()
	}
	if r.mouseJustPressed(ebiten.MouseButtonRight) {
		return r.handleRightClick()
	}
	return InputAction{}
}

// handleBuildKey 1–6 rakam tuşlarıyla bina inşaatı başlatır.
func (r *Renderer) handleBuildKey() InputAction {
	buildingSlots := []string{"market", "farm", "barracks", "port", "walls", "temple"}
	keys := []ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5, ebiten.Key6}
	for i, k := range keys {
		if r.keyJustPressed(k) && i < len(buildingSlots) {
			return InputAction{Kind: ActionBuild, TargetRegion: r.SelectedRegion, BuildingID: buildingSlots[i]}
		}
	}
	return InputAction{}
}

// handleFactionSelectInput fraksiyon seçim ekranındaki tuş ve fare girişlerini işler.
func (r *Renderer) handleFactionSelectInput() InputAction {
	factions := selectableFactions(r.gs)
	n := len(factions)

	// Hover ile kart vurgusunu güncelle
	mx, my := ebiten.CursorPosition()
	if i := r.factionCardHoverIndex(float64(mx), float64(my)); i >= 0 {
		r.factionCursor = i
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) || r.keyJustPressed(ebiten.KeyArrowRight) {
		r.factionCursor = (r.factionCursor + 1) % n
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) || r.keyJustPressed(ebiten.KeyArrowLeft) {
		r.factionCursor = (r.factionCursor - 1 + n) % n
	}
	if r.keyJustPressed(ebiten.KeyEnter) && r.factionCursor < len(factions) {
		return InputAction{Kind: ActionSelectFaction, TargetFaction: factions[r.factionCursor]}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if uiRectHit(float64(mx), float64(my), backButtonRect()) {
			r.factionCursor = 0
			return InputAction{Kind: ActionBack}
		}
		if i := r.factionCardHoverIndex(float64(mx), float64(my)); i >= 0 {
			return InputAction{Kind: ActionSelectFaction, TargetFaction: factions[i]}
		}
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.factionCursor = 0
		return InputAction{Kind: ActionBack}
	}
	return InputAction{}
}

// handleLeftClick sol tıklamayı yorumlar: UI tuşları, ordu seçimi, bölge seçimi.
func (r *Renderer) handleLeftClick() InputAction {
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)

	if r.SelectedArmy != "" && armyPanelCloseHit(fx, fy) {
		r.SelectedArmy = ""
		return InputAction{}
	}
	if r.SelectedRegion != "" && regionPanelCloseHit(fx, fy) {
		r.SelectedRegion = ""
		return InputAction{}
	}

	if eventLogToggleHit(fx, fy, r.eventLogCollapsed) {
		r.eventLogCollapsed = !r.eventLogCollapsed
		return InputAction{}
	}
	if idx := eventLogCloseHit(fx, fy, len(r.eventLog), r.eventLogCollapsed, r.eventLogScroll); idx >= 0 {
		r.eventLog = append(r.eventLog[:idx], r.eventLog[idx+1:]...)
		r.clampEventLogScroll()
		return InputAction{}
	}
	if idx := eventLogCardHit(fx, fy, len(r.eventLog), r.eventLogCollapsed, r.eventLogScroll); idx >= 0 {
		r.eventDetail = r.eventLog[idx]
		return InputAction{}
	}

	if topDateHudMenuButtonHit(fx, fy) {
		r.pauseCursor = 0
		return InputAction{Kind: ActionOpenPauseMenu}
	}
	if musicHudInteractiveHit(fx, fy) {
		if rectF32Hit(fx, fy, musicHudToggleRect()) {
			return InputAction{Kind: ActionToggleMusic}
		}
		if rectF32Hit(fx, fy, musicHudNextRect()) {
			return InputAction{Kind: ActionNextMusic}
		}
	}

	// --- Alt panel butonları ---
	rects := BottomButtonRects()
	if fx >= float64(rects[0][0]) && fx <= float64(rects[0][0]+rects[0][2]) &&
		fy >= float64(rects[0][1]) && fy <= float64(rects[0][1]+rects[0][3]) {
		r.showDiplomacy = !r.showDiplomacy
		r.showTech = false
		r.diplomacyFocus = 0
		return InputAction{}
	}
	if fx >= float64(rects[1][0]) && fx <= float64(rects[1][0]+rects[1][2]) &&
		fy >= float64(rects[1][1]) && fy <= float64(rects[1][1]+rects[1][3]) {
		r.showTech = !r.showTech
		r.showDiplomacy = false
		r.techCursor = 0
		return InputAction{}
	}
	if fx >= float64(rects[2][0]) && fx <= float64(rects[2][0]+rects[2][2]) &&
		fy >= float64(rects[2][1]) && fy <= float64(rects[2][1]+rects[2][3]) {
		return InputAction{Kind: ActionEndTurn}
	}

	// UI alanlarında tıklama işleme
	if topStatusPanelHit(fx, fy) || topDateHudHit(fx, fy) || bottomActionHudHit(fx, fy) || musicHudHit(fx, fy) ||
		eventLogPanelHit(fx, fy, r.eventLogCollapsed) || minimapHit(fx, fy) {
		return InputAction{}
	}

	if r.SelectedRegion != "" {
		if delta := regionTaxButtonHit(fx, fy, r.gs, r.SelectedRegion); delta != 0 {
			return InputAction{Kind: ActionAdjustTax, TargetRegion: r.SelectedRegion, Delta: delta}
		}
		if bid := BuildingGridHitTest(fx, fy, r.gs, r.SelectedRegion); bid != "" {
			return InputAction{Kind: ActionBuild, TargetRegion: r.SelectedRegion, BuildingID: bid}
		}
	}

	// Birim oluştur paneli tıklaması — bölge seçiminden önce kontrol edilmeli
	if uid := RecruitPanelHitTest(fx, fy, r.gs, r.SelectedRegion); uid != "" {
		return InputAction{Kind: ActionRecruitSpecific, TargetRegion: r.SelectedRegion, BuildingID: uid}
	}
	if RecruitPanelBoundsHit(fx, fy, r.gs, r.SelectedRegion) {
		return InputAction{}
	}
	if r.SelectedRegion != "" && regionPanelHit(fx, fy) {
		return InputAction{}
	}

	// BÖL butonu tıklaması
	if r.selectedArmyIsPlayerOwned() && SplitButtonHitTest(fx, fy, r.gs, r.SelectedArmy) {
		return InputAction{Kind: ActionSplitArmy, ArmyID: r.SelectedArmy}
	}
	// BİRLEŞTİR butonu tıklaması
	if r.selectedArmyIsPlayerOwned() && MergeButtonHitTest(fx, fy, r.gs, r.SelectedArmy) {
		return InputAction{Kind: ActionMergeArmies, ArmyID: r.SelectedArmy}
	}

	// Ordu ikonu tıklaması → seç / seçimi kaldır
	armyPositions := r.armyIconPositions()
	for i := len(armyPositions) - 1; i >= 0; i-- {
		pos := armyPositions[i]
		dx := fx - float64(pos.X)
		dy := fy - float64(pos.Y)
		if math.Sqrt(dx*dx+dy*dy) < 14 {
			if r.SelectedArmy == pos.ArmyID {
				r.SelectedArmy = ""
				return InputAction{}
			}
			r.SelectedArmy = pos.ArmyID
			r.SelectedRegion = ""
			return InputAction{Kind: ActionSelectArmy, ArmyID: pos.ArmyID}
		}
	}

	// Bölge / deniz bölgesi seçimi
	wx, wy := r.screenToWorld(fx, fy)
	rid := r.worldMap.RegionAt(int(wx), int(wy))
	if rid != "" {
		if region, ok := r.gs.Regions[rid]; ok && region.IsSea {
			// Naval donanma seçiliyse deniz bölgesine tıklama = hareket komutu
			if r.selectedArmyIsPlayerOwned() {
				if a, ok2 := r.gs.Armies[r.SelectedArmy]; ok2 && a.IsNaval {
					return InputAction{Kind: ActionMoveArmy, ArmyID: r.SelectedArmy, TargetRegion: rid}
				}
			}
			// Seçili donanma yoksa deniz bölgesini seç (highlight için)
			r.SelectedArmy = ""
			r.SelectedRegion = rid
			return InputAction{}
		}
	}
	r.SelectedArmy = ""
	r.SelectedRegion = rid
	return InputAction{}
}

// handleRightClick sağ tıklamayı yorumlar: seçili ordunun hareket/saldırı emri.
func (r *Renderer) handleRightClick() InputAction {
	if r.SelectedArmy == "" {
		return InputAction{}
	}

	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	if topStatusPanelHit(fx, fy) || topDateHudHit(fx, fy) || bottomActionHudHit(fx, fy) || musicHudHit(fx, fy) ||
		eventLogPanelHit(fx, fy, r.eventLogCollapsed) || minimapHit(fx, fy) {
		return InputAction{}
	}

	a, ok := r.gs.Armies[r.SelectedArmy]
	if !ok || a.OwnerID != string(r.gs.PlayerFactionID) || a.MovePoints <= 0 {
		return InputAction{}
	}

	wx, wy := r.screenToWorld(float64(mx), float64(my))
	rid := r.worldMap.RegionAt(int(wx), int(wy))
	if rid == "" {
		return InputAction{}
	}

	src, srcOK := r.gs.Regions[a.RegionID]
	if !srcOK {
		return InputAction{}
	}
	for _, n := range src.Neighbors {
		if n != rid {
			continue
		}
		target, ok := r.gs.Regions[rid]
		if !ok {
			break
		}
		// Düşman bölge ama savaş yok → onay diyalogu aç
		if target.OwnerID != "" && target.OwnerID != a.OwnerID {
			key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(target.OwnerID))
			rel, exists := r.gs.Relations[key]
			if !exists || rel.Stance != faction.StanceWar {
				name := target.OwnerID
				if f, ok := r.gs.Factions[faction.FactionID(target.OwnerID)]; ok {
					name = f.NameTR
				}
				r.warConfirm = warConfirmState{
					show:        true,
					factionName: name,
					factionID:   target.OwnerID,
					pendingArmy: r.SelectedArmy,
					pendingDest: rid,
				}
				return InputAction{}
			}
		}
		act := InputAction{Kind: ActionMoveArmy, ArmyID: r.SelectedArmy, TargetRegion: rid}
		r.SelectedArmy = ""
		return act
	}
	return InputAction{}
}

func (r *Renderer) handleCamera() {
	speed := 6.0 / r.camScale

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		r.camX -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		r.camX += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		r.camY -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		r.camY += speed
	}

	mx, my := ebiten.CursorPosition()
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		if r.isDragging {
			prevWX, prevWY := r.screenToWorld(float64(r.lastMX), float64(r.lastMY))
			curWX, curWY := r.screenToWorld(float64(mx), float64(my))
			r.camX += prevWX - curWX
			r.camY += prevWY - curWY
		}
		r.isDragging = true
	} else {
		r.isDragging = false
	}
	r.lastMX, r.lastMY = mx, my

	_, dy := ebiten.Wheel()
	if dy != 0 {
		if eventLogPanelHit(float64(mx), float64(my), r.eventLogCollapsed) && !r.eventLogCollapsed {
			r.scrollEventLog(dy)
			return
		}
		mouseWX, mouseWY := r.screenToWorld(float64(mx), float64(my))
		minScale := minCameraScale()
		if dy > 0 && r.camScale < 3.0 {
			r.camScale *= 1.12
			if r.camScale > 3.0 {
				r.camScale = 3.0
			}
		} else if dy < 0 && r.camScale > minScale {
			r.camScale /= 1.12
			if r.camScale < minScale {
				r.camScale = minScale
			}
		}
		afterWX, afterWY := r.screenToWorld(float64(mx), float64(my))
		r.camX += mouseWX - afterWX
		r.camY += mouseWY - afterWY
	}
}

func (r *Renderer) scrollEventLog(dy float64) {
	if dy > 0 {
		r.eventLogScroll--
	} else if dy < 0 {
		r.eventLogScroll++
	}
	r.clampEventLogScroll()
}

func (r *Renderer) clampEventLogScroll() {
	maxScroll := eventLogMaxScroll(len(r.eventLog), r.eventLogCollapsed)
	if r.eventLogScroll < 0 {
		r.eventLogScroll = 0
	}
	if r.eventLogScroll > maxScroll {
		r.eventLogScroll = maxScroll
	}
}

// --- Input yardımcıları ---

func (r *Renderer) keyJustPressed(key ebiten.Key) bool {
	pressed := ebiten.IsKeyPressed(key)
	was := r.prevKeys[key]
	r.prevKeys[key] = pressed
	return pressed && !was
}

func (r *Renderer) mouseJustPressed(btn ebiten.MouseButton) bool {
	pressed := ebiten.IsMouseButtonPressed(btn)
	was := r.prevMouse[btn]
	r.prevMouse[btn] = pressed
	return pressed && !was
}

// --- Savaş ilan onay diyalogu ---

func (r *Renderer) drawWarConfirmDialog(screen *ebiten.Image) {
	const (
		dlgW  = float32(380)
		dlgH  = float32(130)
		btnDW = float32(110)
		btnDH = float32(36)
	)
	cx := float32(ScreenWidth)/2 - dlgW/2
	cy := float32(ScreenHeight)/2 - dlgH/2

	// Arka plan
	vector.FillRect(screen, cx-2, cy-2, dlgW+4, dlgH+4, color.RGBA{110, 90, 50, 255}, false)
	vector.FillRect(screen, cx, cy, dlgW, dlgH, color.RGBA{12, 10, 8, 245}, false)

	// Mesaj
	msg := r.warConfirm.factionName + " ile savaş ilan edilsin mi?"
	tw := MeasureText(msg, FaceMed)
	DrawText(screen, msg, float64(cx)+(float64(dlgW)-tw)/2, float64(cy)+18, FaceMed, color.RGBA{255, 220, 100, 255})

	// Evet butonu
	yesX := cx + dlgW/2 - btnDW - 10
	btnY := cy + dlgH - btnDH - 16
	vector.FillRect(screen, yesX, btnY, btnDW, btnDH, color.RGBA{160, 40, 40, 230}, false)
	tw2 := MeasureText("Savaş İlan Et", FaceSmall)
	DrawText(screen, "Savaş İlan Et", float64(yesX)+(float64(btnDW)-tw2)/2, float64(btnY)+10, FaceSmall, color.RGBA{255, 220, 220, 255})

	// Hayır butonu
	noX := cx + dlgW/2 + 10
	vector.FillRect(screen, noX, btnY, btnDW, btnDH, color.RGBA{50, 50, 50, 230}, false)
	tw3 := MeasureText("Hayır", FaceSmall)
	DrawText(screen, "Hayır", float64(noX)+(float64(btnDW)-tw3)/2, float64(btnY)+10, FaceSmall, color.RGBA{200, 200, 200, 255})
}

func (r *Renderer) handleWarConfirmInput() InputAction {
	const (
		dlgW  = float32(380)
		dlgH  = float32(130)
		btnDW = float32(110)
		btnDH = float32(36)
	)
	cx := float32(ScreenWidth)/2 - dlgW/2
	cy := float32(ScreenHeight)/2 - dlgH/2
	btnY := cy + dlgH - btnDH - 16
	yesX := cx + dlgW/2 - btnDW - 10
	noX := cx + dlgW/2 + 10

	mxi, myi := ebiten.CursorPosition()
	mx, my := float32(mxi), float32(myi)

	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if mx >= yesX && mx <= yesX+btnDW && my >= btnY && my <= btnY+btnDH {
			// Evet: savaş ilan et ve orduyu taşı
			wc := r.warConfirm
			r.warConfirm = warConfirmState{}
			r.SelectedArmy = ""
			return InputAction{
				Kind:          ActionDeclareWarAndMove,
				ArmyID:        wc.pendingArmy,
				TargetRegion:  wc.pendingDest,
				TargetFaction: faction.FactionID(wc.factionID),
			}
		}
		if mx >= noX && mx <= noX+btnDW && my >= btnY && my <= btnY+btnDH {
			r.warConfirm = warConfirmState{}
			return InputAction{}
		}
	}
	if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyN) {
		r.warConfirm = warConfirmState{}
	}
	if r.keyJustPressed(ebiten.KeyY) || r.keyJustPressed(ebiten.KeyEnter) {
		wc := r.warConfirm
		r.warConfirm = warConfirmState{}
		r.SelectedArmy = ""
		return InputAction{
			Kind:          ActionDeclareWarAndMove,
			ArmyID:        wc.pendingArmy,
			TargetRegion:  wc.pendingDest,
			TargetFaction: faction.FactionID(wc.factionID),
		}
	}
	return InputAction{}
}

func (r *Renderer) ShowConfirmDialog(title, message, acceptLabel, declineLabel string, action InputAction, declineHook func()) {
	r.confirmDialog = confirmDialogState{
		show:          true,
		title:         title,
		message:       message,
		messageLines:  wrapTextLines(message, FaceSmall, float64(confirmDialogW)-40),
		acceptLabel:   acceptLabel,
		declineLabel:  declineLabel,
		pendingAction: action,
		declineHook:   declineHook,
	}
}

func (r *Renderer) showEditExitConfirm() {
	r.confirmDialog = confirmDialogState{
		show:          true,
		title:         "Kaydedilmemis Degisiklik",
		message:       "Edit mode degisiklikleri kaydedilmedi. Cikmadan once ne yapmak istiyorsunuz?",
		messageLines:  wrapTextLines("Edit mode degisiklikleri kaydedilmedi. Cikmadan once ne yapmak istiyorsunuz?", FaceSmall, float64(confirmDialogW)-40),
		acceptLabel:   "Kaydet",
		thirdLabel:    "Kaydetmeden Cik",
		declineLabel:  "Iptal",
		pendingAction: InputAction{Kind: ActionSaveScenarioAndGoMainMenu},
		thirdAction:   InputAction{Kind: ActionGoMainMenu},
	}
}

func (r *Renderer) drawConfirmDialog(screen *ebiten.Image) {
	cx := float32(ScreenWidth)/2 - confirmDialogW/2
	cy := float32(ScreenHeight)/2 - confirmDialogH/2

	vector.FillRect(screen, cx-2, cy-2, confirmDialogW+4, confirmDialogH+4, color.RGBA{110, 90, 50, 255}, false)
	vector.FillRect(screen, cx, cy, confirmDialogW, confirmDialogH, color.RGBA{12, 10, 8, 245}, false)

	DrawText(screen, r.confirmDialog.title, float64(cx)+20, float64(cy)+28, FaceLarge, color.RGBA{255, 220, 100, 255})
	lines := r.confirmDialog.messageLines
	for i, line := range lines {
		if i >= 3 {
			break
		}
		DrawText(screen, line, float64(cx)+20, float64(cy)+58+float64(i)*17, FaceSmall, color.RGBA{220, 220, 220, 255})

	}
	r.drawConfirmDialogButtons(screen, cx, cy)
}

func (r *Renderer) drawConfirmDialogButtons(screen *ebiten.Image, cx, cy float32) {
	btnY := cy + confirmDialogH - confirmDialogBtnH - 16
	if r.confirmDialog.thirdLabel != "" {
		saveX, discardX, cancelX := confirmDialogThreeButtonXs(cx)
		drawConfirmDialogButton(screen, saveX, btnY, r.confirmDialog.acceptLabel, color.RGBA{70, 140, 70, 240})
		drawConfirmDialogButton(screen, discardX, btnY, r.confirmDialog.thirdLabel, color.RGBA{145, 95, 45, 235})
		drawConfirmDialogButton(screen, cancelX, btnY, r.confirmDialog.declineLabel, color.RGBA{70, 70, 70, 220})
		return
	}
	yesX := cx + confirmDialogW/2 - confirmDialogBtnW - 10
	noX := cx + confirmDialogW/2 + 10
	drawConfirmDialogButton(screen, yesX, btnY, r.confirmDialog.acceptLabel, color.RGBA{70, 140, 70, 240})
	drawConfirmDialogButton(screen, noX, btnY, r.confirmDialog.declineLabel, color.RGBA{70, 70, 70, 220})
}

func drawConfirmDialogButton(screen *ebiten.Image, x, y float32, label string, bg color.RGBA) {
	vector.FillRect(screen, x, y, confirmDialogBtnW, confirmDialogBtnH, bg, false)
	w := MeasureText(label, FaceSmall)
	DrawText(screen, label, float64(x)+(float64(confirmDialogBtnW)-w)/2, float64(y)+10, FaceSmall, ColorWhite)
}

func confirmDialogThreeButtonXs(cx float32) (float32, float32, float32) {
	gap := float32(14)
	totalW := confirmDialogBtnW*3 + gap*2
	saveX := cx + (confirmDialogW-totalW)/2
	discardX := saveX + confirmDialogBtnW + gap
	cancelX := discardX + confirmDialogBtnW + gap
	return saveX, discardX, cancelX
}

func (r *Renderer) handleConfirmDialogInput() InputAction {
	cx := float32(ScreenWidth)/2 - confirmDialogW/2
	cy := float32(ScreenHeight)/2 - confirmDialogH/2
	btnY := cy + confirmDialogH - confirmDialogBtnH - 16
	yesX := cx + confirmDialogW/2 - confirmDialogBtnW - 10
	noX := cx + confirmDialogW/2 + 10
	var thirdX float32
	if r.confirmDialog.thirdLabel != "" {
		yesX, thirdX, noX = confirmDialogThreeButtonXs(cx)
	}

	mxi, myi := ebiten.CursorPosition()
	mx, my := float32(mxi), float32(myi)

	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if mx >= yesX && mx <= yesX+confirmDialogBtnW && my >= btnY && my <= btnY+confirmDialogBtnH {
			action := r.confirmDialog.pendingAction
			r.confirmDialog = confirmDialogState{}
			return action
		}
		if r.confirmDialog.thirdLabel != "" &&
			mx >= thirdX && mx <= thirdX+confirmDialogBtnW && my >= btnY && my <= btnY+confirmDialogBtnH {
			action := r.confirmDialog.thirdAction
			r.confirmDialog = confirmDialogState{}
			return action
		}
		if mx >= noX && mx <= noX+confirmDialogBtnW && my >= btnY && my <= btnY+confirmDialogBtnH {
			if r.confirmDialog.declineHook != nil {
				r.confirmDialog.declineHook()
			}
			r.confirmDialog = confirmDialogState{}
			return InputAction{}
		}
	}
	if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyN) {
		if r.confirmDialog.declineHook != nil {
			r.confirmDialog.declineHook()
		}
		r.confirmDialog = confirmDialogState{}
	}
	if r.keyJustPressed(ebiten.KeyY) || r.keyJustPressed(ebiten.KeyEnter) {
		action := r.confirmDialog.pendingAction
		r.confirmDialog = confirmDialogState{}
		return action
	}
	return InputAction{}
}

// --- Alt çizim yardımcıları ---

func drawControls(screen *ebiten.Image) {
	col := color.RGBA{120, 120, 120, 120}
	DrawText(screen, "[WASD] Kamera  [Orta Tuş] Sürükle  [Tekerlek] Zoom  [F11] Tam Ekran",
		10, float64(ScreenHeight-20), FaceSmall, col)
}

// factionColor fraksiyon rengini döner; bulunamazsa gri.
func factionColor(gs *state.GameState, ownerID string) color.RGBA {
	for fid, f := range gs.Factions {
		if string(fid) == ownerID {
			return color.RGBA{f.Color[0], f.Color[1], f.Color[2], 220}
		}
	}
	return color.RGBA{120, 120, 120, 200}
}
