package render

import (
	"image/color"
	"math"
	"sort"
	"strconv"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/audio"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
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
	showDiplomacy   bool
	diplomacyFocus  int
	diplomacyScroll int

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
	editSelectedFaction        faction.FactionID
	editDraggingSettlement     bool
	editDraggingRegion         bool
	editRenaming               bool
	editTextTarget             editTextTarget
	editTextRunes              []rune
	editInspectorTab           editInspectorTab
	editDirty                  bool
	editVoronoiDebug           bool
	editOwnerDropdown          *Dropdown
	editTerrainDropdown        *Dropdown
	editSettlementTypeDropdown *Dropdown
	editUnitTypeDropdown       *Dropdown
	editSelectedUnitType       string
	armyNeighborBuf            []world.RegionID
	editVisualNeighborBuf      []world.RegionID
	editBoundaryPixelBuf       []int
	editShapeSession           *shapeEditSession
	editShapePainting          bool
	editShapeTool              editShapeTool
	editShapeBrushMode         editShapeBrushMode
	editShapeBrushRadius       int
	editShapeStrokeBefore      *editWorldSnapshot
	editShapeStrokeLastX       int
	editShapeStrokeLastY       int
	editShapeStrokeHasLast     bool
	editShapeStrokeDirty       bool
	editRegionPaintOverrides   map[int]world.RegionID
	editRegionPaintBaseline    []uint16
	editUndoStack              []editCommand
	editRedoStack              []editCommand
	editRegionDragStart        *editRegionCenterSnapshot
	editSettlementDragStart    []editRegionSettlementsSnapshot
	editFactionForm            editFactionFormState
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

type editCommand struct {
	undo func(*Renderer)
	redo func(*Renderer)
}

type editRegionCenterSnapshot struct {
	Region world.RegionID
	X      int
	Y      int
}

type editShapeTool int

const (
	editShapeToolShape editShapeTool = iota
	editShapeToolRegion
)

type editRegionSettlementsSnapshot struct {
	Region      world.RegionID
	Settlements []world.Settlement
}

type editWorldSnapshot struct {
	Regions              map[world.RegionID]*world.Region
	RegionOrder          []world.RegionID
	Factions             map[faction.FactionID]*faction.Faction
	Armies               map[army.ArmyID]*army.Army
	Relations            map[string]*faction.Relation
	ShapeData            world.CountryShapeJSON
	RegionPaintOverrides map[int]world.RegionID
	Selected             world.RegionID
	Settlement           int
	Faction              faction.FactionID
	Army                 army.ArmyID
	Player               faction.FactionID
}

type editFactionFormState struct {
	show           bool
	create         bool
	active         editFactionFormField
	originalID     faction.FactionID
	id             string
	name           string
	nameTR         string
	religion       religion.Type
	color          [3]uint8
	playable       bool
	gold           string
	grain          string
	iron           string
	timber         string
	spice          string
	cloth          string
	ai             string
	relationTarget faction.FactionID
	relationScore  string
	relationStance faction.DiplomaticStance
	errorText      string
}

type editFactionFormField int

const (
	editFactionFieldNone editFactionFormField = iota
	editFactionFieldID
	editFactionFieldName
	editFactionFieldNameTR
	editFactionFieldGold
	editFactionFieldGrain
	editFactionFieldIron
	editFactionFieldTimber
	editFactionFieldSpice
	editFactionFieldCloth
	editFactionFieldAI
)

type editTextTarget int

const (
	editTextNone editTextTarget = iota
	editTextSettlementNameTR
	editTextRegionNameTR
	editTextRegionName
)

type editInspectorTab int

const (
	editInspectorMap editInspectorTab = iota
	editInspectorShape
	editInspectorData
)

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
		armyNeighborBuf:            make([]world.RegionID, 0, 16),
		editVisualNeighborBuf:      make([]world.RegionID, 0, 16),
		editBoundaryPixelBuf:       make([]int, 0, 4096),
		editShapeBrushMode:         editShapeBrushPaint,
		editShapeBrushRadius:       6,
		editUndoStack:              make([]editCommand, 0, 64),
		editRedoStack:              make([]editCommand, 0, 64),
		editRegionPaintOverrides:   make(map[int]world.RegionID),
		editOwnerDropdown:          NewDropdown(dropX, dropY, dropW, dropH, "Sahip Sec"),
		editTerrainDropdown:        NewDropdown(dropX, dropY, dropW, dropH, "Arazi Tipi"),
		editSettlementTypeDropdown: NewDropdown(dropX, dropY, dropW, dropH, "Yerlesim Tipi"),
		editUnitTypeDropdown:       NewDropdown(dropX, dropY, dropW, dropH, "Birim Tipi"),
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
func (r *Renderer) MarkMapDirty() {
	if r == nil || r.worldMap == nil {
		return
	}
	r.worldMap.MarkDirty()
}

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
	r.invalidateShapeEditSession()
	r.resetCamera()
	r.SelectedRegion = ""
	r.SelectedArmy = ""
	r.eventLogScroll = 0
	// Oyun durumundan region paint overrides'ı geri yükle
	if gs.RegionPaintOverrides != nil {
		r.editRegionPaintOverrides = make(map[int]world.RegionID, len(gs.RegionPaintOverrides))
		for k, v := range gs.RegionPaintOverrides {
			r.editRegionPaintOverrides[k] = v
		}
		// Overrides'ı visual haritaya uygula
		r.rebuildEditWorldMap()
	}
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
		r.drawEditShapeOverlay(screen)
	}

	// 5. Ordu ikonları
	r.drawArmies(screen, armyPositions)

	// 6. UI panelleri
	if r.gs.Phase != state.PhaseEditMode {
		DrawBottomPanel(screen, r.gs, r.showDiplomacy, r.showTech)
		DrawRegionPanel(screen, r.gs, r.SelectedRegion)
		DrawRecruitPanel(screen, r.gs, r.SelectedRegion)
		DrawArmyDetailPanel(screen, r.gs, r.SelectedArmy)
		DrawMinimap(screen, r.gs, r.camX, r.camY, r.camScale)
	}
	if r.gs.Phase != state.PhaseEditMode {
		DrawEventLog(screen, r.eventLog, r.eventLogCollapsed, r.eventLogScroll)
		DrawHoverTooltip(screen, r.gs, r.SelectedRegion)
	} else {
		r.drawEditModeHud(screen)
		r.drawEditInspector(screen)
		r.drawEditFactionForm(screen)
	}

	// 7. Diplomasi paneli (üst katman)
	if r.showDiplomacy {
		DrawDiplomacyPanel(screen, r.gs, r.diplomacyFocus, r.diplomacyScroll)
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

func armyCanEmbark(gs *state.GameState, a *army.Army) bool {
	if gs == nil || a == nil || a.IsNaval || len(a.Units) == 0 {
		return false
	}
	for _, u := range a.Units {
		ut, ok := gs.UnitTypes[u.TypeID]
		if !ok || !ut.Embarkable {
			return false
		}
	}
	return true
}

func hasFriendlyEmbarkFleet(gs *state.GameState, ownerID string, seaRegionID world.RegionID) bool {
	if gs == nil {
		return false
	}
	for _, candidate := range gs.Armies {
		if candidate == nil || candidate.OwnerID != ownerID || !candidate.IsNaval || candidate.RegionID != seaRegionID {
			continue
		}
		if len(candidate.EmbarkedUnits) > 0 {
			continue
		}
		for _, u := range candidate.Units {
			ut, ok := gs.UnitTypes[u.TypeID]
			if ok && ut.Category == army.CategoryNavalTrans {
				return true
			}
		}
	}
	return false
}

func armyCanEnterRegion(gs *state.GameState, a *army.Army, target *world.Region) bool {
	if target == nil || target.IsLocked || a == nil {
		return false
	}
	if a.IsNaval {
		if target.CanLandEnter() {
			if len(a.EmbarkedUnits) == 0 {
				return false
			}
			if target.OwnerID == "" || target.OwnerID == a.OwnerID {
				return true
			}
			key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(target.OwnerID))
			rel, ok := gs.Relations[key]
			return ok && rel.Stance == faction.StanceWar
		}
		return target.CanNavalEnter()
	}
	if target.CanNavalEnter() {
		return armyCanEmbark(gs, a) && hasFriendlyEmbarkFleet(gs, a.OwnerID, target.ID)
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
			if ok && armyCanEnterRegion(gs, playerArmy, targetRegion) {
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
		if !armyCanEnterRegion(r.gs, a, nRegion) {
			continue
		}

		sx, sy := r.regionScreenPos(nRegion)

		var col color.RGBA
		if a.IsNaval {
			if nRegion.CanLandEnter() {
				col = color.RGBA{255, 215, 110, 220}
			} else {
				// Deniz bölgeleri için sabit açık mavi — tarafsız su
				col = color.RGBA{100, 200, 255, 220}
			}
		} else {
			if nRegion.IsSea {
				col = color.RGBA{120, 230, 240, 220}
				vector.StrokeCircle(screen, float32(sx), float32(sy), 18, 3, col, true)
				DrawTextCentered(screen, "⛴", sx, sy-8, FaceSmall, color.RGBA{200, 240, 255, 220})
				continue
			}
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

type armyDisplayGroupKey struct {
	RegionID world.RegionID
	AnchorX  int
	AnchorY  int
	Anchored bool
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
// Kara orduları region/yerleşim anchor'ında, sadece demirli donanmalar bağlı
// liman yerleşimi anchor'ında, diğer donanmalar ise deniz bölgesi anchor'ında çizilir.
func (r *Renderer) armyIconPositions() []armyIconPos {
	const iconStep = float32(26) // ikon genişliği 20 + 6px boşluk

	byGroup := map[armyDisplayGroupKey][]army.ArmyID{}
	groupBase := map[armyDisplayGroupKey][2]float32{}
	for aid, a := range r.gs.Armies {
		key, sx, sy, ok := r.armyDisplayGroup(a)
		if !ok {
			continue
		}
		byGroup[key] = append(byGroup[key], aid)
		if _, exists := groupBase[key]; !exists {
			groupBase[key] = [2]float32{sx, sy - 22}
		}
	}

	r.armyIconBuf = r.armyIconBuf[:0]
	for key, aids := range byGroup {
		base := groupBase[key]
		sort.Slice(aids, func(i, j int) bool { return aids[i] < aids[j] })

		n := float32(len(aids))
		startX := base[0] - (n-1)*iconStep/2
		for i, aid := range aids {
			r.armyIconBuf = append(r.armyIconBuf, armyIconPos{
				ArmyID: aid,
				X:      startX + float32(i)*iconStep,
				Y:      base[1],
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

func (r *Renderer) armyDisplayGroup(a *army.Army) (armyDisplayGroupKey, float32, float32, bool) {
	if a == nil {
		return armyDisplayGroupKey{}, 0, 0, false
	}
	region := r.gs.Regions[a.RegionID]
	if region == nil {
		return armyDisplayGroupKey{}, 0, 0, false
	}
	if a.IsNaval && region.IsSea {
		if ax, ay, ok := r.dockedFleetAnchor(a, region); ok {
			sx, sy := r.worldToScreen(float64(ax), float64(ay))
			return armyDisplayGroupKey{AnchorX: ax, AnchorY: ay, Anchored: true}, float32(sx), float32(sy), true
		}
	}
	sx, sy := r.regionScreenPos(region)
	return armyDisplayGroupKey{RegionID: region.ID}, float32(sx), float32(sy), true
}

func (r *Renderer) dockedFleetAnchor(a *army.Army, seaRegion *world.Region) (int, int, bool) {
	if a == nil || seaRegion == nil || !a.IsNaval || !seaRegion.IsSea || a.DockedRegionID == "" {
		return 0, 0, false
	}
	dockedRegion := r.gs.Regions[a.DockedRegionID]
	if dockedRegion == nil || dockedRegion.IsSea {
		return 0, 0, false
	}
	return r.dockedSettlementAnchor(dockedRegion, a.DockedSettlementID)
}

func (r *Renderer) dockedSettlementAnchor(region *world.Region, settlementID string) (int, int, bool) {
	if region == nil {
		return 0, 0, false
	}
	if settlementID != "" {
		for i, settlement := range region.Settlements {
			if settlement.ID != settlementID {
				continue
			}
			if ax, ay, ok := r.worldMap.SettlementAnchor(region.ID, i); ok {
				return ax, ay, true
			}
			break
		}
	}
	for i, settlement := range region.Settlements {
		if settlement.Type != world.SettlementPort {
			continue
		}
		if ax, ay, ok := r.worldMap.SettlementAnchor(region.ID, i); ok {
			return ax, ay, true
		}
	}
	if ax, ay, ok := r.worldMap.PrimarySettlementAnchor(region.ID); ok {
		return ax, ay, true
	}
	return 0, 0, false
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
		if r.gs.Phase != state.PhaseEditMode && a.OwnerID != string(r.gs.PlayerFactionID) && !enemyArmyInPlayerMoveRange(r.gs, a) {
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
	tx := float64(cx) - tw/2
	ty := float64(cy) - 5
	textCol, shadowCol := armyIconCountColors(col)
	DrawText(screen, countStr, tx-1, ty, FaceSmall, shadowCol)
	DrawText(screen, countStr, tx+1, ty, FaceSmall, shadowCol)
	DrawText(screen, countStr, tx, ty-1, FaceSmall, shadowCol)
	DrawText(screen, countStr, tx, ty+1, FaceSmall, shadowCol)
	DrawText(screen, countStr, tx, ty, FaceSmall, textCol)
}

func armyIconCountColors(bg color.RGBA) (color.RGBA, color.RGBA) {
	luminance := 0.299*float64(bg.R) + 0.587*float64(bg.G) + 0.114*float64(bg.B)
	if luminance >= 160 {
		return color.RGBA{20, 16, 12, 255}, color.RGBA{245, 240, 230, 210}
	}
	return color.RGBA{255, 255, 255, 255}, color.RGBA{12, 10, 8, 220}
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
	DrawText(screen, "Sol: sec/tasi   Alt+sol: yerlesim   Ctrl+Alt+sol: bolge   Shift+sol: merkez   Ctrl+Z/Y: geri/ileri",
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
	historyState := "Geri/Ileri: " + itoa(len(r.editUndoStack)) + "/" + itoa(len(r.editRedoStack))
	if r.editRenaming {
		DrawText(screen, r.editTextLabel()+": "+string(r.editTextRunes), float64(x)+14, float64(y)+80, FaceSmall, ColorGold)
	} else {
		DrawText(screen, debugState+"   "+historyState+"   V: debug   Esc: ana menu", float64(x)+14, float64(y)+80, FaceSmall, ColorGray)
	}
}

func (r *Renderer) drawEditInspector(screen *ebiten.Image) {
	x, y, w, h := editInspectorRect()
	drawRoundedRect(screen, x, y, w, h, 8, color.RGBA{16, 20, 24, 226})
	drawPanelBorder(screen, x, y, w, h)

	DrawText(screen, "EDITOR", float64(x)+14, float64(y)+10, FaceMed, ColorGold)
	drawEditInspectorTab(screen, editInspectorMap, "Harita")
	drawEditInspectorTab(screen, editInspectorShape, "Shape")
	drawEditInspectorTab(screen, editInspectorData, "Veri")
	ly := float64(y) + 58

	if r.editInspectorTab == editInspectorShape {
		r.drawEditShapeInspector(screen, ly)
		return
	}

	if r.editInspectorTab == editInspectorData {
		r.drawEditDataInspector(screen, ly)
		return
	}

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
	regionKind := "Kara Bolgesi"
	ownerLabel := region.OwnerID
	settlementLabel := itoa(len(region.Settlements))
	if region.IsSea {
		regionKind = "Deniz Bolgesi"
		if ownerLabel == "" {
			ownerLabel = "-"
		}
		settlementLabel = "yok"
	}
	DrawText(screen, name, float64(x)+14, ly, FaceSmall, ColorWhite)
	ly += 18
	DrawText(screen, "ID: "+string(region.ID), float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 18
	DrawText(screen, "Tur: "+regionKind+"   Sahip: "+ownerLabel+"   Arazi: "+string(region.Terrain), float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 18
	DrawText(screen, "Merkez: "+itoa(region.WorldX)+","+itoa(region.WorldY)+"   Yerlesim: "+settlementLabel, float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 22
	DrawText(screen, "Kilit: "+editBoolLabel(region.IsLocked)+"   Acilis: "+itoa(region.UnlockTurn)+"   Komsu: "+itoa(len(region.Neighbors)), float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 20

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
	} else if region.IsSea {
		DrawText(screen, "Deniz bolgesinde yerlesim yok.", float64(x)+14, ly, FaceSmall, ColorGray)
	} else {
		DrawText(screen, "Yerlesim secili degil.", float64(x)+14, ly, FaceSmall, ColorGray)
	}

	r.drawEditInspectorButtons(screen, region)
	r.editOwnerDropdown.Draw(screen)
	r.editTerrainDropdown.Draw(screen)
	r.editSettlementTypeDropdown.Draw(screen)
	r.editUnitTypeDropdown.Draw(screen)
}

func (r *Renderer) drawEditInspectorButtons(screen *ebiten.Image, region *world.Region) {
	canAdd := region != nil && !region.IsSea
	canRegion := region != nil
	canSettlement := r.hasEditSelection()
	addSettlementLabel := "Yerlesim Ekle"
	settlementTypeLabel := "Tip"
	renameSettlementLabel := "Isim"
	deleteSettlementLabel := "Yerlesim Sil"
	if region != nil && region.IsSea {
		addSettlementLabel = "Denizde Yok"
		settlementTypeLabel = "Tip Yok"
		renameSettlementLabel = "Isim Yok"
		deleteSettlementLabel = "Silinmez"
	} else if !canSettlement {
		settlementTypeLabel = "Tip Sec"
		renameSettlementLabel = "Isim Sec"
		deleteSettlementLabel = "Sil Sec"
	}
	drawEditInspectorButton(screen, editButtonAddSettlement, addSettlementLabel, canAdd)
	drawEditInspectorButton(screen, editButtonSettlementType, settlementTypeLabel, canSettlement)
	drawEditInspectorButton(screen, editButtonSetCapitalSettlement, "Ana Yap", canSettlement)
	drawEditInspectorButton(screen, editButtonRenameSettlement, renameSettlementLabel, canSettlement)
	drawEditInspectorButton(screen, editButtonRegionTerrain, "Arazi", canRegion)
	drawEditInspectorButton(screen, editButtonRegionOwner, "Sahip", canRegion)
	drawEditInspectorButton(screen, editButtonRegionNameTR, "Ad TR", canRegion)
	drawEditInspectorButton(screen, editButtonRegionName, "Ad EN", canRegion)
	drawEditInspectorButton(screen, editButtonRegionLock, "Kilit", canRegion)
	drawEditInspectorButton(screen, editButtonUnlockMinus, "-10 Tur", canRegion)
	drawEditInspectorButton(screen, editButtonUnlockPlus, "+10 Tur", canRegion)
	drawEditInspectorButton(screen, editButtonSyncNeighbors, "Komsu Sync", canRegion)
	drawEditInspectorButton(screen, editButtonAddRegion, "Bolge Ekle", canRegion)
	drawEditInspectorButton(screen, editButtonDeleteRegion, "Bolge Sil", canRegion)
	drawEditInspectorButton(screen, editButtonDeleteSettlement, deleteSettlementLabel, canSettlement)
	drawEditInspectorButton(screen, editButtonSaveScenario, "Kaydet", true)
}

func drawEditInspectorTab(screen *ebiten.Image, tab editInspectorTab, label string) {
	rect := editInspectorTabRect(tab)
	drawTinyPanelButton(screen, float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3]), label, true)
}

func (r *Renderer) drawEditDataInspector(screen *ebiten.Image, ly float64) {
	x, _, _, _ := editInspectorRect()
	region := r.gs.Regions[r.editSelectedRegion]
	f := r.selectedEditFaction()

	DrawText(screen, "GENIS VERI EDITORU", float64(x)+14, ly, FaceSmall, ColorGold)
	ly += 22
	if f == nil {
		DrawText(screen, "Sahipli bolge veya ordu sec.", float64(x)+14, ly, FaceSmall, ColorGray)
		ly += 20
	} else {
		name := f.NameTR
		if name == "" {
			name = f.Name
		}
		DrawText(screen, "Fraksiyon: "+name+" ["+string(f.ID)+"]", float64(x)+14, ly, FaceSmall, ColorWhite)
		ly += 18
		DrawText(screen, "Altin "+itoa(f.Gold)+"  Tahil "+itoa(f.Grain)+"  Demir "+itoa(f.Iron), float64(x)+14, ly, FaceSmall, ColorGray)
		ly += 18
		DrawText(screen, "Kereste "+itoa(f.Timber)+"  Baharat "+itoa(f.Spice)+"  Kumas "+itoa(f.Cloth), float64(x)+14, ly, FaceSmall, ColorGray)
		ly += 18
		DrawText(screen, "Playable: "+editBoolLabel(f.IsPlayable)+"  AI: "+itoa(f.AIAggressiveness), float64(x)+14, ly, FaceSmall, ColorGray)
	}
	ly += 24

	if r.SelectedArmy != "" {
		if a := r.gs.Armies[r.SelectedArmy]; a != nil {
			r.ensureEditSelectedUnitType(a)
			DrawText(screen, "Ordu: "+string(a.ID), float64(x)+14, ly, FaceSmall, ColorGold)
			ly += 18
			kind := "Kara"
			if a.IsNaval {
				kind = "Donanma"
			}
			DrawText(screen, "Tip: "+kind+"  Sahip: "+a.OwnerID+"  Bolge: "+string(a.RegionID), float64(x)+14, ly, FaceSmall, ColorGray)
			ly += 18
			DrawText(screen, "Birim: "+itoa(len(a.Units))+" / "+itoa(army.MaxArmySize)+"  Secili: "+r.editSelectedUnitType, float64(x)+14, ly, FaceSmall, ColorGray)
			ly += 18
			r.drawEditArmyUnitCounts(screen, a, float64(x)+14, ly)
		}
	} else {
		DrawText(screen, "Ordu secili degil.", float64(x)+14, ly, FaceSmall, ColorGray)
	}

	drawEditInspectorButton(screen, editButtonAddFaction, "Faction Ekle", true)
	drawEditInspectorButton(screen, editButtonEditFaction, "Faction Duzenle", f != nil)
	drawEditInspectorButton(screen, editButtonDeleteFaction, "Faction Sil", f != nil)
	drawEditInspectorButton(screen, editButtonAddArmy, "Ordu Ekle", r.canAddEditLandArmy(region))
	drawEditInspectorButton(screen, editButtonAddFleet, "Donanma Ekle", r.canAddEditFleet(region))
	drawEditInspectorButton(screen, editButtonDeleteArmy, "Ordu Sil", r.SelectedArmy != "")
	unitTypeLabel := "Birim Tipi"
	if r.editSelectedUnitType != "" {
		unitTypeLabel = "Birim Tipi: " + r.editSelectedUnitType
	}
	drawEditInspectorButton(screen, editButtonArmyUnitType, unitTypeLabel, r.SelectedArmy != "")
	drawEditInspectorButton(screen, editButtonArmyUnitMinus, "Birim -", r.canRemoveSelectedArmyUnit())
	drawEditInspectorButton(screen, editButtonArmyUnitPlus, "Birim +", r.canAddSelectedArmyUnit())
	drawEditInspectorButton(screen, editButtonArmyOwnerFromRegion, "Sahibi Al", r.SelectedArmy != "" && region != nil && region.OwnerID != "")
	drawEditInspectorButton(screen, editButtonSaveScenario, "Kaydet", true)
	r.editUnitTypeDropdown.Draw(screen)
}

func (r *Renderer) drawEditArmyUnitCounts(screen *ebiten.Image, a *army.Army, x, y float64) {
	if len(a.Units) == 0 {
		DrawText(screen, "Birim yok.", x, y, FaceSmall, ColorGray)
		return
	}
	var types [army.MaxArmySize]string
	var counts [army.MaxArmySize]int
	typeCount := 0
	for _, unit := range a.Units {
		found := -1
		for i := 0; i < typeCount; i++ {
			if types[i] == unit.TypeID {
				found = i
				break
			}
		}
		if found >= 0 {
			counts[found]++
			continue
		}
		if typeCount < len(types) {
			types[typeCount] = unit.TypeID
			counts[typeCount] = 1
			typeCount++
		}
	}
	drawn := 0
	for i := 0; i < typeCount; i++ {
		typeID := types[i]
		name := typeID
		if utype := r.gs.UnitTypes[typeID]; utype != nil {
			name = utype.NameTR
			if name == "" {
				name = utype.Name
			}
		}
		DrawText(screen, name+": "+itoa(counts[i]), x, y+float64(drawn*16), FaceSmall, ColorGray)
		drawn++
		if drawn >= 4 {
			if typeCount > drawn {
				DrawText(screen, "...", x, y+float64(drawn*16), FaceSmall, ColorGray)
			}
			return
		}
	}
}

func (r *Renderer) drawEditFactionForm(screen *ebiten.Image) {
	if !r.editFactionForm.show {
		return
	}
	x, y, w, h := editFactionFormRect()
	drawRoundedRect(screen, x, y, w, h, 8, color.RGBA{14, 18, 22, 244})
	drawPanelBorder(screen, x, y, w, h)
	title := "FACTION EKLE"
	if !r.editFactionForm.create {
		title = "FACTION DUZENLE"
	}
	DrawText(screen, title, float64(x)+18, float64(y)+14, FaceLarge, ColorGold)

	r.drawFactionFormField(screen, editFactionFieldID, "ID", r.editFactionForm.id)
	r.drawFactionFormField(screen, editFactionFieldNameTR, "Ad TR", r.editFactionForm.nameTR)
	r.drawFactionFormField(screen, editFactionFieldName, "Ad EN", r.editFactionForm.name)
	r.drawFactionFormField(screen, editFactionFieldGold, "Altin", r.editFactionForm.gold)
	r.drawFactionFormField(screen, editFactionFieldGrain, "Tahil", r.editFactionForm.grain)
	r.drawFactionFormField(screen, editFactionFieldIron, "Demir", r.editFactionForm.iron)
	r.drawFactionFormField(screen, editFactionFieldTimber, "Kereste", r.editFactionForm.timber)
	r.drawFactionFormField(screen, editFactionFieldSpice, "Baharat", r.editFactionForm.spice)
	r.drawFactionFormField(screen, editFactionFieldCloth, "Kumas", r.editFactionForm.cloth)
	r.drawFactionFormField(screen, editFactionFieldAI, "AI", r.editFactionForm.ai)

	drawEditFactionFormButton(screen, editFactionFormReligion, "Din: "+string(r.editFactionForm.religion))
	drawEditFactionFormButton(screen, editFactionFormPlayable, "Playable: "+editBoolLabel(r.editFactionForm.playable))
	relationTitle := "Iliski: yok"
	if r.editFactionForm.relationTarget != "" {
		relationTitle = "Iliski: " + string(r.editFactionForm.relationTarget)
	}
	drawEditFactionFormButton(screen, editFactionFormRelationTarget, relationTitle)
	drawEditFactionFormButton(screen, editFactionFormRelationStance, "Durum: "+string(r.editFactionForm.relationStance))
	drawEditFactionFormButton(screen, editFactionFormRelationScoreMinus, "Skor -10")
	drawEditFactionFormButton(screen, editFactionFormRelationScorePlus, "Skor +10")
	DrawText(screen, "Skor: "+r.editFactionForm.relationScore, float64(x)+18, float64(y)+304, FaceSmall, ColorGray)

	col := r.editFactionForm.color
	preview := editFactionFormColorPreviewRect()
	vector.FillRect(screen, float32(preview[0]), float32(preview[1]), float32(preview[2]), float32(preview[3]), color.RGBA{col[0], col[1], col[2], 255}, false)
	vector.StrokeRect(screen, float32(preview[0]), float32(preview[1]), float32(preview[2]), float32(preview[3]), 1, ColorGold, false)
	DrawText(screen, "Renk "+itoa(int(col[0]))+","+itoa(int(col[1]))+","+itoa(int(col[2])), float64(x)+338, float64(y)+332, FaceSmall, ColorGray)
	drawEditFactionFormButton(screen, editFactionFormRedMinus, "R-")
	drawEditFactionFormButton(screen, editFactionFormRedPlus, "R+")
	drawEditFactionFormButton(screen, editFactionFormGreenMinus, "G-")
	drawEditFactionFormButton(screen, editFactionFormGreenPlus, "G+")
	drawEditFactionFormButton(screen, editFactionFormBlueMinus, "B-")
	drawEditFactionFormButton(screen, editFactionFormBluePlus, "B+")

	if r.editFactionForm.errorText != "" {
		DrawText(screen, r.editFactionForm.errorText, float64(x)+18, float64(y)+float64(h)-74, FaceSmall, ColorRed)
	}
	drawEditFactionFormButton(screen, editFactionFormSave, "Kaydet")
	drawEditFactionFormButton(screen, editFactionFormCancel, "Iptal")
}

func drawEditFactionFormButton(screen *ebiten.Image, kind editFactionFormButton, label string) {
	x, y, w, h := rectXYWH(editFactionFormButtonRect(kind))
	drawTinyPanelButton(screen, x, y, w, h, label, true)
}

func (r *Renderer) drawFactionFormField(screen *ebiten.Image, field editFactionFormField, label, value string) {
	rect := editFactionFieldRect(field)
	col := color.RGBA{28, 32, 38, 235}
	if r.editFactionForm.active == field {
		col = color.RGBA{44, 48, 54, 245}
	}
	vector.FillRect(screen, float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3]), col, false)
	vector.StrokeRect(screen, float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3]), 1, color.RGBA{120, 105, 60, 210}, false)
	DrawText(screen, label, rect[0], rect[1]-16, FaceSmall, ColorGray)
	DrawText(screen, value, rect[0]+8, rect[1]+7, FaceSmall, ColorWhite)
}

func rectXYWH(rect uiRect) (float32, float32, float32, float32) {
	return float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3])
}

func editFactionFormRect() (float32, float32, float32, float32) {
	const w, h = float32(640), float32(520)
	return float32(ScreenWidth)/2 - w/2, float32(ScreenHeight)/2 - h/2, w, h
}

func editFactionFormHit(mx, my float64) bool {
	x, y, w, h := editFactionFormRect()
	return mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h)
}

func editFactionFieldRect(field editFactionFormField) uiRect {
	x, y, _, _ := editFactionFormRect()
	left := float64(x) + 18
	right := float64(x) + 338
	top := float64(y) + 78
	const fw, fh, gap = float64(284), float64(30), float64(24)
	row := func(n int) float64 { return top + float64(n)*(fh+gap) }
	switch field {
	case editFactionFieldID:
		return uiRect{left, row(0), fw, fh}
	case editFactionFieldNameTR:
		return uiRect{right, row(0), fw, fh}
	case editFactionFieldName:
		return uiRect{left, row(1), fw, fh}
	case editFactionFieldGold:
		return uiRect{right, row(1), fw/2 - 6, fh}
	case editFactionFieldGrain:
		return uiRect{right + fw/2 + 6, row(1), fw/2 - 6, fh}
	case editFactionFieldIron:
		return uiRect{left, row(2), fw/2 - 6, fh}
	case editFactionFieldTimber:
		return uiRect{left + fw/2 + 6, row(2), fw/2 - 6, fh}
	case editFactionFieldSpice:
		return uiRect{right, row(2), fw/2 - 6, fh}
	case editFactionFieldCloth:
		return uiRect{right + fw/2 + 6, row(2), fw/2 - 6, fh}
	case editFactionFieldAI:
		return uiRect{left, row(3), fw/2 - 6, fh}
	default:
		return uiRect{}
	}
}

type editFactionFormButton int

const (
	editFactionFormSave editFactionFormButton = iota
	editFactionFormCancel
	editFactionFormReligion
	editFactionFormPlayable
	editFactionFormRelationTarget
	editFactionFormRelationStance
	editFactionFormRelationScoreMinus
	editFactionFormRelationScorePlus
	editFactionFormRedMinus
	editFactionFormRedPlus
	editFactionFormGreenMinus
	editFactionFormGreenPlus
	editFactionFormBlueMinus
	editFactionFormBluePlus
)

func editFactionFormButtonRect(kind editFactionFormButton) uiRect {
	x, y, w, h := editFactionFormRect()
	right := float64(x) + 338
	switch kind {
	case editFactionFormReligion:
		return uiRect{right, float64(y) + 240, 136, 28}
	case editFactionFormPlayable:
		return uiRect{right + 148, float64(y) + 240, 136, 28}
	case editFactionFormRelationTarget:
		return uiRect{float64(x) + 18, float64(y) + 240, 284, 28}
	case editFactionFormRelationStance:
		return uiRect{float64(x) + 18, float64(y) + 272, 136, 28}
	case editFactionFormRelationScoreMinus:
		return uiRect{float64(x) + 166, float64(y) + 272, 64, 28}
	case editFactionFormRelationScorePlus:
		return uiRect{float64(x) + 238, float64(y) + 272, 64, 28}
	case editFactionFormRedMinus:
		return uiRect{right, float64(y) + 382, 42, 26}
	case editFactionFormRedPlus:
		return uiRect{right + 48, float64(y) + 382, 42, 26}
	case editFactionFormGreenMinus:
		return uiRect{right + 100, float64(y) + 382, 42, 26}
	case editFactionFormGreenPlus:
		return uiRect{right + 148, float64(y) + 382, 42, 26}
	case editFactionFormBlueMinus:
		return uiRect{right + 200, float64(y) + 382, 42, 26}
	case editFactionFormBluePlus:
		return uiRect{right + 248, float64(y) + 382, 42, 26}
	case editFactionFormSave:
		return uiRect{float64(x) + float64(w) - 264, float64(y) + float64(h) - 52, 116, 32}
	case editFactionFormCancel:
		return uiRect{float64(x) + float64(w) - 136, float64(y) + float64(h) - 52, 116, 32}
	default:
		return uiRect{}
	}
}

func editFactionFormColorPreviewRect() uiRect {
	x, y, _, _ := editFactionFormRect()
	return uiRect{float64(x) + 338, float64(y) + 352, 284, 22}
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
	editButtonRegionNameTR
	editButtonRegionName
	editButtonRegionLock
	editButtonUnlockMinus
	editButtonUnlockPlus
	editButtonSyncNeighbors
	editButtonAddRegion
	editButtonDeleteRegion
	editButtonDeleteSettlement
	editButtonSaveScenario
	editButtonShapePaint
	editButtonShapeErase
	editButtonShapeRegionPaint
	editButtonShapeRegionErase
	editButtonShapeBrushMinus
	editButtonShapeBrushPlus
	editButtonAddFaction
	editButtonEditFaction
	editButtonDeleteFaction
	editButtonAddArmy
	editButtonAddFleet
	editButtonDeleteArmy
	editButtonArmyUnitType
	editButtonArmyUnitMinus
	editButtonArmyUnitPlus
	editButtonArmyOwnerFromRegion
)

func editInspectorRect() (float32, float32, float32, float32) {
	const w, h = float32(360), float32(580)
	return 18, float32(ScreenHeight) - h - 18, w, h
}

func editInspectorHit(mx, my float64) bool {
	x, y, w, h := editInspectorRect()
	return mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h)
}

func editInspectorButtonRect(kind editInspectorButton) uiRect {
	x, y, _, h := editInspectorRect()
	const bw, bh, gap = float64(158), float64(24), float64(8)
	left := float64(x) + 14
	right := left + bw + gap
	row1 := float64(y) + float64(h) - 264
	row2 := row1 + bh + gap
	row3 := row2 + bh + gap
	row4 := row3 + bh + gap
	row5 := row4 + bh + gap
	row6 := row5 + bh + gap
	row7 := row6 + bh + gap
	row8 := row7 + bh + gap
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
	case editButtonRegionNameTR:
		return uiRect{left, row4, bw, bh}
	case editButtonRegionName:
		return uiRect{right, row4, bw, bh}
	case editButtonRegionLock:
		return uiRect{left, row5, bw, bh}
	case editButtonUnlockMinus:
		return uiRect{right, row5, (bw - gap) / 2, bh}
	case editButtonUnlockPlus:
		return uiRect{right + (bw+gap)/2, row5, (bw - gap) / 2, bh}
	case editButtonSyncNeighbors:
		return uiRect{left, row6, bw, bh}
	case editButtonAddRegion:
		return uiRect{right, row6, bw, bh}
	case editButtonDeleteRegion:
		return uiRect{left, row7, bw, bh}
	case editButtonDeleteSettlement:
		return uiRect{right, row7, bw, bh}
	case editButtonSaveScenario:
		return uiRect{left, row8, bw*2 + gap, bh}
	case editButtonShapePaint:
		return uiRect{left, row1, bw, bh}
	case editButtonShapeErase:
		return uiRect{right, row1, bw, bh}
	case editButtonShapeRegionPaint:
		return uiRect{left, row2, bw, bh}
	case editButtonShapeRegionErase:
		return uiRect{right, row2, bw, bh}
	case editButtonShapeBrushMinus:
		return uiRect{left, row3, bw, bh}
	case editButtonShapeBrushPlus:
		return uiRect{right, row3, bw, bh}
	case editButtonAddFaction:
		return uiRect{left, row1, bw, bh}
	case editButtonEditFaction:
		return uiRect{right, row1, bw, bh}
	case editButtonDeleteFaction:
		return uiRect{left, row2, bw, bh}
	case editButtonAddArmy:
		return uiRect{right, row2, bw, bh}
	case editButtonAddFleet:
		return uiRect{left, row3, bw, bh}
	case editButtonDeleteArmy:
		return uiRect{right, row3, bw, bh}
	case editButtonArmyUnitType:
		return uiRect{left, row4, bw, bh}
	case editButtonArmyUnitMinus:
		return uiRect{right, row4, (bw - gap) / 2, bh}
	case editButtonArmyUnitPlus:
		return uiRect{right + (bw+gap)/2, row4, (bw - gap) / 2, bh}
	case editButtonArmyOwnerFromRegion:
		return uiRect{right, row5, bw, bh}
	default:
		return uiRect{}
	}
}

func editInspectorTabRect(tab editInspectorTab) uiRect {
	x, y, _, _ := editInspectorRect()
	const tw, th, gap = float64(68), float64(24), float64(8)
	left := float64(x) + 82 + float64(tab)*(tw+gap)
	return uiRect{left, float64(y) + 9, tw, th}
}

func editInspectorButtonAt(mx, my float64) editInspectorButton {
	if kind := editMapInspectorButtonAt(mx, my); kind != editButtonNone {
		return kind
	}
	return editDataInspectorButtonAt(mx, my)
}

func editMapInspectorButtonAt(mx, my float64) editInspectorButton {
	for kind := editButtonAddSettlement; kind <= editButtonSaveScenario; kind++ {
		if uiRectHit(mx, my, editInspectorButtonRect(kind)) {
			return kind
		}
	}
	return editButtonNone
}

func editShapeInspectorButtonAt(mx, my float64) editInspectorButton {
	for kind := editButtonShapePaint; kind <= editButtonShapeBrushPlus; kind++ {
		if uiRectHit(mx, my, editInspectorButtonRect(kind)) {
			return kind
		}
	}
	if uiRectHit(mx, my, editInspectorButtonRect(editButtonSaveScenario)) {
		return editButtonSaveScenario
	}
	return editButtonNone
}

func editDataInspectorButtonAt(mx, my float64) editInspectorButton {
	for kind := editButtonAddFaction; kind <= editButtonArmyOwnerFromRegion; kind++ {
		if uiRectHit(mx, my, editInspectorButtonRect(kind)) {
			return kind
		}
	}
	if uiRectHit(mx, my, editInspectorButtonRect(editButtonSaveScenario)) {
		return editButtonSaveScenario
	}
	return editButtonNone
}

func (r *Renderer) editInspectorActiveButtonAt(mx, my float64) editInspectorButton {
	if uiRectHit(mx, my, editInspectorTabRect(editInspectorMap)) ||
		uiRectHit(mx, my, editInspectorTabRect(editInspectorShape)) ||
		uiRectHit(mx, my, editInspectorTabRect(editInspectorData)) {
		return editButtonSaveScenario
	}
	if r.editInspectorTab == editInspectorShape {
		return editShapeInspectorButtonAt(mx, my)
	}
	if r.editInspectorTab == editInspectorData {
		return editDataInspectorButtonAt(mx, my)
	}
	return editMapInspectorButtonAt(mx, my)
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
	r.editUnitTypeDropdown.SetPosition(dx, dy)
}

func editMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (r *Renderer) drawEditRegionCenters(screen *ebiten.Image) {
	for _, region := range r.gs.Regions {
		if region == nil || region.IsLocked {
			continue
		}
		sx, sy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
		col := color.RGBA{80, 220, 255, 190}
		if region.IsSea {
			col = color.RGBA{120, 210, 255, 210}
		}
		if region.ID == r.editSelectedRegion && r.editSelectedSettlement < 0 {
			if region.IsSea {
				col = color.RGBA{70, 235, 255, 245}
			} else {
				col = color.RGBA{255, 190, 45, 240}
			}
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
	if region == nil {
		r.drawEditVoronoiLegend(screen, "", nil)
		return
	}

	r.editVisualNeighborBuf = r.worldMap.VisualNeighbors(rid, r.editVisualNeighborBuf)
	r.editBoundaryPixelBuf = r.worldMap.BoundaryPixels(rid, r.editBoundaryPixelBuf)
	r.drawEditVoronoiBoundary(screen, r.editBoundaryPixelBuf)

	cx, cy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
	for _, nrid := range r.editVisualNeighborBuf {
		neighbor := r.gs.Regions[nrid]
		if neighbor == nil {
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
		if neighbor == nil {
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

func (r *Renderer) pushEditCommand(cmd editCommand) {
	if cmd.undo == nil || cmd.redo == nil {
		return
	}
	r.editUndoStack = append(r.editUndoStack, cmd)
	r.editRedoStack = r.editRedoStack[:0]
	r.editDirty = true
}

func (r *Renderer) undoEditCommand() {
	if len(r.editUndoStack) == 0 {
		return
	}
	last := len(r.editUndoStack) - 1
	cmd := r.editUndoStack[last]
	r.editUndoStack = r.editUndoStack[:last]
	cmd.undo(r)
	r.editRedoStack = append(r.editRedoStack, cmd)
	r.editDirty = true
}

func (r *Renderer) redoEditCommand() {
	if len(r.editRedoStack) == 0 {
		return
	}
	last := len(r.editRedoStack) - 1
	cmd := r.editRedoStack[last]
	r.editRedoStack = r.editRedoStack[:last]
	cmd.redo(r)
	r.editUndoStack = append(r.editUndoStack, cmd)
	r.editDirty = true
}

func cloneSettlements(settlements []world.Settlement) []world.Settlement {
	if settlements == nil {
		return nil
	}
	clone := make([]world.Settlement, len(settlements))
	copy(clone, settlements)
	return clone
}

func (r *Renderer) settlementSnapshot(rid world.RegionID) editRegionSettlementsSnapshot {
	region := r.gs.Regions[rid]
	if region == nil {
		return editRegionSettlementsSnapshot{Region: rid}
	}
	return editRegionSettlementsSnapshot{
		Region:      rid,
		Settlements: cloneSettlements(region.Settlements),
	}
}

func uniqueSettlementSnapshots(snaps []editRegionSettlementsSnapshot) []editRegionSettlementsSnapshot {
	out := snaps[:0]
	for _, snap := range snaps {
		seen := false
		for _, existing := range out {
			if existing.Region == snap.Region {
				seen = true
				break
			}
		}
		if !seen {
			out = append(out, snap)
		}
	}
	return out
}

func (r *Renderer) restoreSettlementSnapshots(snaps []editRegionSettlementsSnapshot) {
	for _, snap := range snaps {
		region := r.gs.Regions[snap.Region]
		if region == nil {
			continue
		}
		region.Settlements = cloneSettlements(snap.Settlements)
	}
	r.editDraggingSettlement = false
	r.editDraggingRegion = false
	r.editRenaming = false
	r.worldMap.RebuildSettlementAnchors(r.gs)
}

func (r *Renderer) pushSettlementSnapshots(before, after []editRegionSettlementsSnapshot, selectedRegion world.RegionID, selectedSettlement int) {
	before = uniqueSettlementSnapshots(before)
	after = uniqueSettlementSnapshots(after)
	if len(before) == 0 || len(after) == 0 || settlementSnapshotsEqual(before, after) {
		return
	}
	beforeCopy := cloneSettlementSnapshots(before)
	afterCopy := cloneSettlementSnapshots(after)
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.restoreSettlementSnapshots(beforeCopy)
			rr.editSelectedRegion = selectedRegion
			rr.editSelectedSettlement = -1
		},
		redo: func(rr *Renderer) {
			rr.restoreSettlementSnapshots(afterCopy)
			rr.editSelectedRegion = selectedRegion
			rr.editSelectedSettlement = selectedSettlement
		},
	})
}

func cloneSettlementSnapshots(snaps []editRegionSettlementsSnapshot) []editRegionSettlementsSnapshot {
	out := make([]editRegionSettlementsSnapshot, len(snaps))
	for i, snap := range snaps {
		out[i] = editRegionSettlementsSnapshot{
			Region:      snap.Region,
			Settlements: cloneSettlements(snap.Settlements),
		}
	}
	return out
}

func settlementSnapshotsEqual(a, b []editRegionSettlementsSnapshot) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Region != b[i].Region || !settlementsEqual(a[i].Settlements, b[i].Settlements) {
			return false
		}
	}
	return true
}

func settlementsEqual(a, b []world.Settlement) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func editUndoPressed() bool {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyControlRight)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftRight)
	return ctrl && !shift
}

func editRedoPressed() bool {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyControlRight)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftRight)
	return ctrl && shift
}

func (r *Renderer) handleEditModeInput() InputAction {
	if r.editRenaming {
		return r.handleEditRenameInput()
	}
	if r.editFactionForm.show {
		return r.handleEditFactionFormInput()
	}

	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	leftJustPressed := r.mouseJustPressed(ebiten.MouseButtonLeft)
	rightPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	rightJustPressed := r.mouseJustPressed(ebiten.MouseButtonRight)

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

	if r.editUnitTypeDropdown.IsOpen() {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 && r.editUnitTypeDropdown.HitTest(fx, fy) {
			r.editUnitTypeDropdown.Scroll(wheelY)
			return InputAction{}
		}
	}

	if !r.editOwnerDropdown.IsOpen() && !r.editTerrainDropdown.IsOpen() && !r.editSettlementTypeDropdown.IsOpen() && !r.editUnitTypeDropdown.IsOpen() {
		r.handleCamera()
	}

	if r.editShapePainting && !rightPressed {
		r.finishShapePaintStroke()
		return InputAction{}
	}

	if r.keyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if r.keyJustPressed(ebiten.KeyV) {
		r.editVoronoiDebug = !r.editVoronoiDebug
	}
	if r.keyJustPressed(ebiten.KeyZ) {
		if editRedoPressed() {
			r.redoEditCommand()
			return InputAction{}
		}
		if editUndoPressed() {
			r.undoEditCommand()
			return InputAction{}
		}
	}
	if r.keyJustPressed(ebiten.KeyY) && editUndoPressed() {
		r.redoEditCommand()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.editOwnerDropdown.Close()
		r.editTerrainDropdown.Close()
		r.editSettlementTypeDropdown.Close()
		r.editUnitTypeDropdown.Close()
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
	if r.keyJustPressed(ebiten.KeyDelete) && !r.hasEditSelection() && r.editSelectedRegion != "" {
		r.deleteSelectedRegion()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyDelete) && r.hasEditSelection() {
		r.deleteSelectedSettlement()
		return InputAction{}
	}
	if (r.keyJustPressed(ebiten.KeyF2) || r.keyJustPressed(ebiten.KeyEnter)) && r.hasEditSelection() {
		r.beginEditRename(editTextSettlementNameTR)
		return InputAction{}
	}

	if leftJustPressed {
		if action, ok := r.handleEditInspectorClick(fx, fy); ok {
			return action
		}
	}

	if r.editInspectorTab == editInspectorShape {
		if rightJustPressed && r.beginShapePaintStroke(fx, fy) {
			return InputAction{}
		}
		if r.editShapePainting {
			r.continueShapePaintStroke(fx, fy)
			return InputAction{}
		}
	}

	if r.editDraggingRegion && !leftPressed {
		r.finishRegionCenterDrag()
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
				r.editUnitTypeDropdown.Close()
				r.editSelectedRegion = rid
				r.setEditFactionFromRegion(rid)
				r.editSelectedSettlement = -1
				r.editDraggingRegion = true
				r.editDraggingSettlement = false
				r.editRenaming = false
				r.beginRegionCenterDrag(rid)
				r.moveSelectedRegionCenterTo(fx, fy)
				return InputAction{}
			}
		}
		if editAddModifierPressed() {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.editUnitTypeDropdown.Close()
			if editCreateRegionModifierPressed() {
				r.addRegionAt(fx, fy)
				return InputAction{}
			}
			r.addSettlementAt(fx, fy)
			return InputAction{}
		}

		if aid, ok := r.editArmyAt(fx, fy); ok {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.editUnitTypeDropdown.Close()
			r.SelectedArmy = aid
			if a := r.gs.Armies[aid]; a != nil {
				r.editSelectedRegion = a.RegionID
				r.setEditFactionFromArmy(a)
				r.ensureEditSelectedUnitType(a)
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
			r.editUnitTypeDropdown.Close()
			r.SelectedArmy = ""
			r.editSelectedRegion = rid
			r.setEditFactionFromRegion(rid)
			r.editSelectedSettlement = idx
			r.editDraggingSettlement = true
			r.editDraggingRegion = false
			r.beginSettlementDrag(rid)
			return InputAction{}
		}
		if rid := r.editRegionAt(fx, fy); rid != "" {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.editUnitTypeDropdown.Close()
			r.SelectedArmy = ""
			r.editSelectedRegion = rid
			r.setEditFactionFromRegion(rid)
			r.editSelectedSettlement = -1
			r.editRenaming = false
			r.editDraggingRegion = false
			r.editDraggingSettlement = false
			return InputAction{}
		}
		r.editOwnerDropdown.Close()
		r.editTerrainDropdown.Close()
		r.editSettlementTypeDropdown.Close()
		r.editUnitTypeDropdown.Close()
		r.SelectedArmy = ""
		r.editSelectedRegion = ""
		r.editSelectedFaction = ""
		r.editSelectedSettlement = -1
		r.editRenaming = false
		r.editDraggingRegion = false
	}

	if !leftPressed {
		if r.editDraggingSettlement {
			r.finishSettlementDrag()
		}
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
	if r.editUnitTypeDropdown.IsOpen() {
		if idx, ok := r.editUnitTypeDropdown.GetSelectedOption(fx, fy); ok {
			r.editSelectedUnitType = r.editUnitTypeDropdown.options[idx]
			r.editUnitTypeDropdown.Close()
			return InputAction{}, true
		}
		if r.editUnitTypeDropdown.HitTest(fx, fy) {
			return InputAction{}, true
		}
		if !editInspectorHit(fx, fy) {
			r.editUnitTypeDropdown.Close()
			return InputAction{}, false
		}
	}
	if !editInspectorHit(fx, fy) {
		return InputAction{}, false
	}
	if uiRectHit(fx, fy, editInspectorTabRect(editInspectorMap)) {
		r.editInspectorTab = editInspectorMap
		return InputAction{}, true
	}
	if uiRectHit(fx, fy, editInspectorTabRect(editInspectorShape)) {
		r.editInspectorTab = editInspectorShape
		return InputAction{}, true
	}
	if uiRectHit(fx, fy, editInspectorTabRect(editInspectorData)) {
		r.editInspectorTab = editInspectorData
		return InputAction{}, true
	}
	if r.editInspectorTab == editInspectorShape {
		return r.handleEditShapeInspectorClick(fx, fy)
	}
	if r.editInspectorTab == editInspectorData {
		return r.handleEditDataInspectorClick(fx, fy)
	}
	switch editMapInspectorButtonAt(fx, fy) {
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
			r.beginEditRename(editTextSettlementNameTR)
		}
	case editButtonRegionTerrain:
		r.toggleEditTerrainDropdown()
	case editButtonRegionOwner:
		r.toggleEditOwnerDropdown()
	case editButtonRegionNameTR:
		r.beginEditRename(editTextRegionNameTR)
	case editButtonRegionName:
		r.beginEditRename(editTextRegionName)
	case editButtonRegionLock:
		r.toggleSelectedRegionLock()
	case editButtonUnlockMinus:
		r.adjustSelectedRegionUnlockTurn(-10)
	case editButtonUnlockPlus:
		r.adjustSelectedRegionUnlockTurn(10)
	case editButtonSyncNeighbors:
		r.syncSelectedRegionNeighborsFromVisual()
	case editButtonAddRegion:
		r.addRegionNearSelected()
	case editButtonDeleteRegion:
		r.deleteSelectedRegion()
	case editButtonDeleteSettlement:
		if r.hasEditSelection() {
			r.deleteSelectedSettlement()
		}
	case editButtonSaveScenario:
		return InputAction{Kind: ActionSaveScenario}, true
	}
	return InputAction{}, true
}

func (r *Renderer) handleEditDataInspectorClick(fx, fy float64) (InputAction, bool) {
	if r.editUnitTypeDropdown.IsOpen() {
		if idx, ok := r.editUnitTypeDropdown.GetSelectedOption(fx, fy); ok {
			r.editSelectedUnitType = r.editUnitTypeDropdown.options[idx]
			r.editUnitTypeDropdown.Close()
			return InputAction{}, true
		}
		if r.editUnitTypeDropdown.HitTest(fx, fy) {
			return InputAction{}, true
		}
		if !editInspectorHit(fx, fy) {
			r.editUnitTypeDropdown.Close()
			return InputAction{}, false
		}
	}
	switch editDataInspectorButtonAt(fx, fy) {
	case editButtonAddFaction:
		r.openFactionCreateForm()
	case editButtonEditFaction:
		r.openFactionEditForm()
	case editButtonDeleteFaction:
		r.deleteSelectedFaction()
	case editButtonAddArmy:
		r.addEditLandArmy()
	case editButtonAddFleet:
		r.addEditFleet()
	case editButtonDeleteArmy:
		r.deleteSelectedArmy()
	case editButtonArmyUnitType:
		r.toggleEditUnitTypeDropdown()
	case editButtonArmyUnitMinus:
		r.removeSelectedArmyUnit()
	case editButtonArmyUnitPlus:
		r.addSelectedArmyUnit()
	case editButtonArmyOwnerFromRegion:
		r.setSelectedArmyOwnerFromRegion()
	case editButtonSaveScenario:
		return InputAction{Kind: ActionSaveScenario}, true
	}
	return InputAction{}, true
}

func (r *Renderer) toggleEditOwnerDropdown() {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil {
		r.editOwnerDropdown.Close()
		return
	}

	dx, dy, _, _ := editOwnerDropdownRect()
	r.editOwnerDropdown.SetPosition(dx, dy)
	r.editOwnerDropdown.SetOptions(editOwnerOptions(r.gs.Factions), region.OwnerID)
	r.editOwnerDropdown.Toggle()
}

func (r *Renderer) toggleEditTerrainDropdown() {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil {
		r.editTerrainDropdown.Close()
		return
	}

	dx, dy, _, _ := editTerrainDropdownRect()
	r.editTerrainDropdown.SetPosition(dx, dy)
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

	dx, dy, _, _ := editSettlementTypeDropdownRect()
	r.editSettlementTypeDropdown.SetPosition(dx, dy)
	region := r.gs.Regions[r.editSelectedRegion]
	settlement := region.Settlements[r.editSelectedSettlement]
	r.editSettlementTypeDropdown.SetOptions(world.AllSettlementTypes(), string(settlement.Type))
	r.editSettlementTypeDropdown.Toggle()
}

func (r *Renderer) hasEditSelection() bool {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	return ok && region != nil && r.editSelectedSettlement >= 0 &&
		r.editSelectedSettlement < len(region.Settlements)
}

func (r *Renderer) beginEditRename(target editTextTarget) {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	switch target {
	case editTextSettlementNameTR:
		if !r.hasEditSelection() {
			return
		}
	case editTextRegionNameTR:
	case editTextRegionName:
	default:
		return
	}
	r.editTextTarget = target
	r.editTextRunes = r.editTextRunes[:0]
	r.editRenaming = true
	r.editDraggingSettlement = false
}

func (r *Renderer) handleEditRenameInput() InputAction {
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.editRenaming = false
		r.editTextTarget = editTextNone
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		r.commitEditRename()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyBackspace) && len(r.editTextRunes) > 0 {
		r.editTextRunes = r.editTextRunes[:len(r.editTextRunes)-1]
	}
	r.editTextRunes = ebiten.AppendInputChars(r.editTextRunes)
	if len(r.editTextRunes) > 64 {
		r.editTextRunes = r.editTextRunes[:64]
	}
	return InputAction{}
}

func (r *Renderer) commitEditRename() {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		r.editRenaming = false
		r.editTextTarget = editTextNone
		return
	}
	newName := strings.TrimSpace(string(r.editTextRunes))
	rid := region.ID
	switch r.editTextTarget {
	case editTextSettlementNameTR:
		if !r.hasEditSelection() {
			break
		}
		idx := r.editSelectedSettlement
		oldName := region.Settlements[idx].NameTR
		if newName != "" && oldName != newName {
			region.Settlements[idx].NameTR = newName
			r.pushEditCommand(editCommand{
				undo: func(rr *Renderer) {
					rr.setSettlementNameTR(rid, idx, oldName)
				},
				redo: func(rr *Renderer) {
					rr.setSettlementNameTR(rid, idx, newName)
				},
			})
			r.editDirty = true
		}
	case editTextRegionNameTR:
		oldName := region.NameTR
		if newName != "" && oldName != newName {
			region.NameTR = newName
			r.pushEditCommand(editCommand{
				undo: func(rr *Renderer) { rr.setRegionNameTR(rid, oldName) },
				redo: func(rr *Renderer) { rr.setRegionNameTR(rid, newName) },
			})
			r.editDirty = true
		}
	case editTextRegionName:
		oldName := region.Name
		if newName != "" && oldName != newName {
			region.Name = newName
			r.pushEditCommand(editCommand{
				undo: func(rr *Renderer) { rr.setRegionName(rid, oldName) },
				redo: func(rr *Renderer) { rr.setRegionName(rid, newName) },
			})
			r.editDirty = true
		}
	}
	r.editRenaming = false
	r.editTextTarget = editTextNone
}

func (r *Renderer) editTextLabel() string {
	switch r.editTextTarget {
	case editTextRegionNameTR:
		return "Bolge Ad TR"
	case editTextRegionName:
		return "Bolge Ad EN"
	default:
		return "Isim"
	}
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
	if region, ok := r.gs.Regions[rid]; ok && region != nil {
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

func (r *Renderer) beginRegionCenterDrag(rid world.RegionID) {
	region := r.gs.Regions[rid]
	if region == nil {
		r.editRegionDragStart = nil
		return
	}
	r.editRegionDragStart = &editRegionCenterSnapshot{
		Region: rid,
		X:      region.WorldX,
		Y:      region.WorldY,
	}
}

func (r *Renderer) finishRegionCenterDrag() {
	start := r.editRegionDragStart
	r.editRegionDragStart = nil
	if start == nil {
		return
	}
	region := r.gs.Regions[start.Region]
	if region == nil || (region.WorldX == start.X && region.WorldY == start.Y) {
		return
	}
	begin := *start
	end := editRegionCenterSnapshot{Region: start.Region, X: region.WorldX, Y: region.WorldY}
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.restoreRegionCenter(begin)
		},
		redo: func(rr *Renderer) {
			rr.restoreRegionCenter(end)
		},
	})
}

func (r *Renderer) restoreRegionCenter(snapshot editRegionCenterSnapshot) {
	region := r.gs.Regions[snapshot.Region]
	if region == nil {
		return
	}
	region.WorldX = snapshot.X
	region.WorldY = snapshot.Y
	r.editSelectedRegion = snapshot.Region
	r.editSelectedSettlement = -1
	r.editDraggingRegion = false
	r.editDraggingSettlement = false
	r.rebuildEditWorldMap()
}

func (r *Renderer) beginSettlementDrag(rid world.RegionID) {
	r.editSettlementDragStart = r.editSettlementDragStart[:0]
	r.editSettlementDragStart = append(r.editSettlementDragStart, r.settlementSnapshot(rid))
}

func (r *Renderer) ensureSettlementDragSnapshot(rid world.RegionID) {
	for _, snap := range r.editSettlementDragStart {
		if snap.Region == rid {
			return
		}
	}
	r.editSettlementDragStart = append(r.editSettlementDragStart, r.settlementSnapshot(rid))
}

func (r *Renderer) finishSettlementDrag() {
	if len(r.editSettlementDragStart) == 0 {
		return
	}
	before := cloneSettlementSnapshots(r.editSettlementDragStart)
	after := make([]editRegionSettlementsSnapshot, 0, len(before)+1)
	for _, snap := range before {
		after = append(after, r.settlementSnapshot(snap.Region))
	}
	if r.editSelectedRegion != "" {
		after = append(after, r.settlementSnapshot(r.editSelectedRegion))
	}
	r.pushSettlementSnapshots(before, after, r.editSelectedRegion, r.editSelectedSettlement)
	r.editSettlementDragStart = r.editSettlementDragStart[:0]
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
	if !ok || region == nil {
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
	before := []editRegionSettlementsSnapshot{r.settlementSnapshot(rid)}

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
	after := []editRegionSettlementsSnapshot{r.settlementSnapshot(rid)}
	r.pushSettlementSnapshots(before, after, rid, r.editSelectedSettlement)
}

func (r *Renderer) deleteSelectedSettlement() {
	if !r.hasEditSelection() {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
	rid := region.ID
	before := []editRegionSettlementsSnapshot{r.settlementSnapshot(rid)}
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
	after := []editRegionSettlementsSnapshot{r.settlementSnapshot(rid)}
	r.pushSettlementSnapshots(before, after, rid, -1)
}

func (r *Renderer) addRegionAt(fx, fy float64) {
	sourceID := r.editRegionAt(fx, fy)
	if sourceID == "" {
		return
	}
	wx, wy := r.screenToWorld(fx, fy)
	x, y := scenarioCoordsFromWorld(wx, wy)
	r.addRegionFromSource(sourceID, x, y)
}

func (r *Renderer) addRegionNearSelected() {
	source := r.gs.Regions[r.editSelectedRegion]
	if source == nil {
		return
	}
	r.addRegionFromSource(source.ID, source.WorldX+12, source.WorldY+12)
}

func (r *Renderer) addRegionFromSource(sourceID world.RegionID, x, y int) {
	source := r.gs.Regions[sourceID]
	if source == nil {
		return
	}
	before := r.worldSnapshot()
	rid := nextRegionID(r.gs)
	nameNo := itoa(len(r.gs.Regions) + 1)
	region := &world.Region{
		ID:               rid,
		Name:             "New Region " + nameNo,
		NameTR:           "Yeni Bolge " + nameNo,
		Terrain:          source.Terrain,
		OwnerID:          source.OwnerID,
		WorldX:           x,
		WorldY:           y,
		ShapeID:          source.ShapeID,
		IsSea:            source.IsSea,
		IsLocked:         source.IsLocked,
		UnlockTurn:       source.UnlockTurn,
		BaseGoldIncome:   source.BaseGoldIncome,
		BaseGrainOutput:  source.BaseGrainOutput,
		BaseIronOutput:   source.BaseIronOutput,
		BaseTimberOutput: source.BaseTimberOutput,
		BaseSpiceOutput:  source.BaseSpiceOutput,
		BaseClothOutput:  source.BaseClothOutput,
		TradeCapacity:    source.TradeCapacity,
		Satisfaction:     source.Satisfaction,
		TaxRate:          source.TaxRate,
		Population:       source.Population,
		Religion:         source.Religion,
		ActiveEventID:    source.ActiveEventID,
		Buildings:        cloneStringSlice(source.Buildings),
	}
	if region.Terrain == "" {
		if region.IsSea {
			region.Terrain = world.TerrainSea
		} else {
			region.Terrain = world.TerrainPlain
		}
	}
	if region.Satisfaction == 0 {
		region.Satisfaction = 70
	}
	if region.TaxRate == 0 {
		region.TaxRate = 45
	}
	r.gs.Regions[rid] = region
	r.insertRegionOrderAfter(sourceID, rid)
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
	r.SelectedArmy = ""
	r.rebuildEditWorldMap()
	visual := r.worldMap.VisualNeighbors(rid, r.editVisualNeighborBuf[:0])
	r.applyVisualNeighbors(rid, visual)
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) deleteSelectedRegion() {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	before := r.worldSnapshot()
	rid := region.ID
	for _, other := range r.gs.Regions {
		removeNeighborID(other, rid)
	}
	delete(r.gs.Regions, rid)
	r.removeRegionFromOrder(rid)
	for aid, a := range r.gs.Armies {
		if a != nil && a.RegionID == rid {
			delete(r.gs.Armies, aid)
		}
	}
	r.editSelectedRegion = ""
	r.editSelectedSettlement = -1
	r.SelectedArmy = ""
	r.rebuildEditWorldMap()
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) setSelectedSettlementCapital() {
	if !r.hasEditSelection() {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
	before := []editRegionSettlementsSnapshot{r.settlementSnapshot(region.ID)}
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
		after := []editRegionSettlementsSnapshot{r.settlementSnapshot(region.ID)}
		r.pushSettlementSnapshots(before, after, region.ID, r.editSelectedSettlement)
	}
}

func (r *Renderer) setSelectedRegionTerrain(terrain world.TerrainType) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil {
		return
	}
	if region.Terrain == terrain {
		return
	}
	rid := region.ID
	old := region.Terrain
	region.Terrain = terrain
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.setRegionTerrainValue(rid, old)
		},
		redo: func(rr *Renderer) {
			rr.setRegionTerrainValue(rid, terrain)
		},
	})
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
	rid := region.ID
	idx := r.editSelectedSettlement
	old := settlement.Type
	settlement.Type = st
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.setSettlementTypeValue(rid, idx, old)
		},
		redo: func(rr *Renderer) {
			rr.setSettlementTypeValue(rid, idx, st)
		},
	})
	r.editDirty = true
}

func (r *Renderer) setSelectedRegionOwner(ownerID string) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil {
		return
	}
	if region.OwnerID == ownerID {
		return
	}
	rid := region.ID
	old := region.OwnerID
	region.OwnerID = ownerID
	r.editSelectedFaction = faction.FactionID(ownerID)
	r.worldMap.MarkDirty()
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.setRegionOwnerValue(rid, old)
		},
		redo: func(rr *Renderer) {
			rr.setRegionOwnerValue(rid, ownerID)
		},
	})
	r.editDirty = true
}

func (r *Renderer) setSettlementNameTR(rid world.RegionID, index int, name string) {
	region := r.gs.Regions[rid]
	if region == nil || index < 0 || index >= len(region.Settlements) {
		return
	}
	region.Settlements[index].NameTR = name
	r.editSelectedRegion = rid
	r.editSelectedSettlement = index
}

func (r *Renderer) setRegionNameTR(rid world.RegionID, name string) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	region.NameTR = name
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
}

func (r *Renderer) setRegionName(rid world.RegionID, name string) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	region.Name = name
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
}

func (r *Renderer) toggleSelectedRegionLock() {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	rid := region.ID
	old := region.IsLocked
	region.IsLocked = !region.IsLocked
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) { rr.setRegionLockValue(rid, old) },
		redo: func(rr *Renderer) { rr.setRegionLockValue(rid, !old) },
	})
	r.editDirty = true
}

func (r *Renderer) setRegionLockValue(rid world.RegionID, locked bool) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	region.IsLocked = locked
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
}

func (r *Renderer) adjustSelectedRegionUnlockTurn(delta int) {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	old := region.UnlockTurn
	next := old + delta
	if next < 0 {
		next = 0
	}
	if old == next {
		return
	}
	rid := region.ID
	region.UnlockTurn = next
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) { rr.setRegionUnlockTurn(rid, old) },
		redo: func(rr *Renderer) { rr.setRegionUnlockTurn(rid, next) },
	})
	r.editDirty = true
}

func (r *Renderer) setRegionUnlockTurn(rid world.RegionID, turn int) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	region.UnlockTurn = turn
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
}

func (r *Renderer) setSettlementTypeValue(rid world.RegionID, index int, typ world.SettlementType) {
	region := r.gs.Regions[rid]
	if region == nil || index < 0 || index >= len(region.Settlements) {
		return
	}
	region.Settlements[index].Type = typ
	r.editSelectedRegion = rid
	r.editSelectedSettlement = index
}

func (r *Renderer) setRegionTerrainValue(rid world.RegionID, terrain world.TerrainType) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	region.Terrain = terrain
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
}

func (r *Renderer) setRegionOwnerValue(rid world.RegionID, ownerID string) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	region.OwnerID = ownerID
	r.editSelectedRegion = rid
	r.editSelectedFaction = faction.FactionID(ownerID)
	r.editSelectedSettlement = -1
	r.worldMap.MarkDirty()
}

func (r *Renderer) syncSelectedRegionNeighborsFromVisual() {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	visual := r.worldMap.VisualNeighbors(region.ID, r.editVisualNeighborBuf[:0])
	before := r.neighborSnapshot(region.ID, visual)
	r.applyVisualNeighbors(region.ID, visual)
	after := r.neighborSnapshot(region.ID, visual)
	if neighborSnapshotsEqual(before, after) {
		return
	}
	rid := region.ID
	beforeCopy := cloneNeighborSnapshots(before)
	afterCopy := cloneNeighborSnapshots(after)
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.restoreNeighborSnapshots(beforeCopy)
			rr.editSelectedRegion = rid
			rr.editSelectedSettlement = -1
		},
		redo: func(rr *Renderer) {
			rr.restoreNeighborSnapshots(afterCopy)
			rr.editSelectedRegion = rid
			rr.editSelectedSettlement = -1
		},
	})
	r.editDirty = true
}

func (r *Renderer) worldSnapshot() editWorldSnapshot {
	return editWorldSnapshot{
		Regions:              cloneRegionMap(r.gs.Regions),
		RegionOrder:          cloneRegionIDSlice(r.gs.RegionOrder),
		Factions:             cloneFactionMap(r.gs.Factions),
		Armies:               cloneArmyMap(r.gs.Armies),
		Relations:            cloneRelationMap(r.gs.Relations),
		ShapeData:            cloneCountryShapeJSON(r.gs.ShapeData),
		RegionPaintOverrides: cloneRegionPaintOverrides(r.editRegionPaintOverrides),
		Selected:             r.editSelectedRegion,
		Settlement:           r.editSelectedSettlement,
		Faction:              r.editSelectedFaction,
		Army:                 r.SelectedArmy,
		Player:               r.gs.PlayerFactionID,
	}
}

func (r *Renderer) pushWorldSnapshotCommand(before, after editWorldSnapshot) {
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) { rr.restoreWorldSnapshot(before) },
		redo: func(rr *Renderer) { rr.restoreWorldSnapshot(after) },
	})
}

func (r *Renderer) restoreWorldSnapshot(snapshot editWorldSnapshot) {
	r.gs.Regions = cloneRegionMap(snapshot.Regions)
	r.gs.RegionOrder = cloneRegionIDSlice(snapshot.RegionOrder)
	r.gs.Factions = cloneFactionMap(snapshot.Factions)
	r.gs.Armies = cloneArmyMap(snapshot.Armies)
	r.gs.Relations = cloneRelationMap(snapshot.Relations)
	r.gs.ShapeData = cloneCountryShapeJSON(snapshot.ShapeData)
	r.editRegionPaintOverrides = cloneRegionPaintOverrides(snapshot.RegionPaintOverrides)
	// Region paint overrides'ı oyun durumuna da senkronize et
	if len(r.editRegionPaintOverrides) > 0 {
		if r.gs.RegionPaintOverrides == nil {
			r.gs.RegionPaintOverrides = make(map[int]world.RegionID)
		}
		for k, v := range r.editRegionPaintOverrides {
			r.gs.RegionPaintOverrides[k] = v
		}
	} else {
		r.gs.RegionPaintOverrides = nil
	}
	r.editSelectedRegion = snapshot.Selected
	r.editSelectedSettlement = snapshot.Settlement
	r.editSelectedFaction = snapshot.Faction
	r.SelectedArmy = snapshot.Army
	r.gs.PlayerFactionID = snapshot.Player
	r.editDraggingSettlement = false
	r.editDraggingRegion = false
	r.editShapePainting = false
	r.editShapeStrokeBefore = nil
	r.editRenaming = false
	r.rebuildEditWorldMap()
}

func cloneRegionMap(src map[world.RegionID]*world.Region) map[world.RegionID]*world.Region {
	dst := make(map[world.RegionID]*world.Region, len(src))
	for rid, region := range src {
		if region == nil {
			continue
		}
		copyRegion := *region
		copyRegion.Neighbors = cloneRegionIDSlice(region.Neighbors)
		copyRegion.Settlements = cloneSettlements(region.Settlements)
		copyRegion.Buildings = cloneStringSlice(region.Buildings)
		if len(region.Shape) > 0 {
			copyRegion.Shape = make([][][2]float32, len(region.Shape))
			for i := range region.Shape {
				copyRegion.Shape[i] = make([][2]float32, len(region.Shape[i]))
				copy(copyRegion.Shape[i], region.Shape[i])
			}
		}
		dst[rid] = &copyRegion
	}
	return dst
}

func cloneArmyMap(src map[army.ArmyID]*army.Army) map[army.ArmyID]*army.Army {
	dst := make(map[army.ArmyID]*army.Army, len(src))
	for aid, a := range src {
		if a == nil {
			continue
		}
		copyArmy := *a
		copyArmy.Units = make([]army.Unit, len(a.Units))
		copy(copyArmy.Units, a.Units)
		dst[aid] = &copyArmy
	}
	return dst
}

func cloneFactionMap(src map[faction.FactionID]*faction.Faction) map[faction.FactionID]*faction.Faction {
	dst := make(map[faction.FactionID]*faction.Faction, len(src))
	for fid, f := range src {
		if f == nil {
			continue
		}
		copyFaction := *f
		if f.Research.Completed != nil {
			copyFaction.Research.Completed = make(map[string]bool, len(f.Research.Completed))
			for id, done := range f.Research.Completed {
				copyFaction.Research.Completed[id] = done
			}
		}
		dst[fid] = &copyFaction
	}
	return dst
}

func cloneRelationMap(src map[string]*faction.Relation) map[string]*faction.Relation {
	dst := make(map[string]*faction.Relation, len(src))
	for key, rel := range src {
		if rel == nil {
			continue
		}
		copyRel := *rel
		dst[key] = &copyRel
	}
	return dst
}

func cloneRegionIDSlice(src []world.RegionID) []world.RegionID {
	if src == nil {
		return nil
	}
	dst := make([]world.RegionID, len(src))
	copy(dst, src)
	return dst
}

func cloneStringSlice(src []string) []string {
	if src == nil {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func (r *Renderer) insertRegionOrderAfter(after, rid world.RegionID) {
	r.removeRegionFromOrder(rid)
	if len(r.gs.RegionOrder) == 0 {
		r.gs.RegionOrder = append(r.gs.RegionOrder, rid)
		return
	}
	for i, existing := range r.gs.RegionOrder {
		if existing == after {
			r.gs.RegionOrder = append(r.gs.RegionOrder, "")
			copy(r.gs.RegionOrder[i+2:], r.gs.RegionOrder[i+1:])
			r.gs.RegionOrder[i+1] = rid
			return
		}
	}
	r.gs.RegionOrder = append(r.gs.RegionOrder, rid)
}

func (r *Renderer) removeRegionFromOrder(rid world.RegionID) {
	out := r.gs.RegionOrder[:0]
	for _, existing := range r.gs.RegionOrder {
		if existing != rid {
			out = append(out, existing)
		}
	}
	r.gs.RegionOrder = out
}

type editRegionNeighborsSnapshot struct {
	Region    world.RegionID
	Neighbors []world.RegionID
}

func (r *Renderer) neighborSnapshot(rid world.RegionID, affected []world.RegionID) []editRegionNeighborsSnapshot {
	snaps := make([]editRegionNeighborsSnapshot, 0, len(affected)+1)
	snaps = append(snaps, r.singleNeighborSnapshot(rid))
	for _, nrid := range affected {
		if nrid != rid {
			snaps = append(snaps, r.singleNeighborSnapshot(nrid))
		}
	}
	return uniqueNeighborSnapshots(snaps)
}

func (r *Renderer) singleNeighborSnapshot(rid world.RegionID) editRegionNeighborsSnapshot {
	region := r.gs.Regions[rid]
	if region == nil {
		return editRegionNeighborsSnapshot{Region: rid}
	}
	neighbors := make([]world.RegionID, len(region.Neighbors))
	copy(neighbors, region.Neighbors)
	return editRegionNeighborsSnapshot{Region: rid, Neighbors: neighbors}
}

func (r *Renderer) applyVisualNeighbors(rid world.RegionID, visual []world.RegionID) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	oldNeighbors := region.Neighbors
	region.Neighbors = sortedRegionIDs(visual)
	for _, oldID := range oldNeighbors {
		if !regionIDContains(visual, oldID) {
			removeNeighborID(r.gs.Regions[oldID], rid)
		}
	}
	for _, nrid := range visual {
		addNeighborID(r.gs.Regions[nrid], rid)
	}
}

func (r *Renderer) restoreNeighborSnapshots(snaps []editRegionNeighborsSnapshot) {
	for _, snap := range snaps {
		region := r.gs.Regions[snap.Region]
		if region == nil {
			continue
		}
		region.Neighbors = make([]world.RegionID, len(snap.Neighbors))
		copy(region.Neighbors, snap.Neighbors)
	}
	r.editDraggingSettlement = false
	r.editDraggingRegion = false
	r.editRenaming = false
}

func uniqueNeighborSnapshots(snaps []editRegionNeighborsSnapshot) []editRegionNeighborsSnapshot {
	out := snaps[:0]
	for _, snap := range snaps {
		seen := false
		for _, existing := range out {
			if existing.Region == snap.Region {
				seen = true
				break
			}
		}
		if !seen {
			out = append(out, snap)
		}
	}
	return out
}

func cloneNeighborSnapshots(snaps []editRegionNeighborsSnapshot) []editRegionNeighborsSnapshot {
	out := make([]editRegionNeighborsSnapshot, len(snaps))
	for i, snap := range snaps {
		out[i].Region = snap.Region
		out[i].Neighbors = make([]world.RegionID, len(snap.Neighbors))
		copy(out[i].Neighbors, snap.Neighbors)
	}
	return out
}

func neighborSnapshotsEqual(a, b []editRegionNeighborsSnapshot) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Region != b[i].Region || !regionIDSlicesEqual(a[i].Neighbors, b[i].Neighbors) {
			return false
		}
	}
	return true
}

func sortedRegionIDs(ids []world.RegionID) []world.RegionID {
	out := make([]world.RegionID, 0, len(ids))
	for _, rid := range ids {
		if rid != "" && !regionIDContains(out, rid) {
			out = append(out, rid)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func addNeighborID(region *world.Region, rid world.RegionID) {
	if region == nil || rid == "" || regionHasNeighbor(region, rid) {
		return
	}
	region.Neighbors = append(region.Neighbors, rid)
	sort.Slice(region.Neighbors, func(i, j int) bool { return region.Neighbors[i] < region.Neighbors[j] })
}

func removeNeighborID(region *world.Region, rid world.RegionID) {
	if region == nil {
		return
	}
	out := region.Neighbors[:0]
	for _, nrid := range region.Neighbors {
		if nrid != rid {
			out = append(out, nrid)
		}
	}
	region.Neighbors = out
}

func regionIDContains(ids []world.RegionID, rid world.RegionID) bool {
	for _, id := range ids {
		if id == rid {
			return true
		}
	}
	return false
}

func regionIDSlicesEqual(a, b []world.RegionID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (r *Renderer) selectedEditFaction() *faction.Faction {
	if r.editSelectedFaction != "" {
		if f := r.gs.Factions[r.editSelectedFaction]; f != nil {
			return f
		}
	}
	if region := r.gs.Regions[r.editSelectedRegion]; region != nil && region.OwnerID != "" {
		return r.gs.Factions[faction.FactionID(region.OwnerID)]
	}
	if r.SelectedArmy != "" {
		if a := r.gs.Armies[r.SelectedArmy]; a != nil && a.OwnerID != "" {
			return r.gs.Factions[faction.FactionID(a.OwnerID)]
		}
	}
	return nil
}

func (r *Renderer) setEditFactionFromRegion(rid world.RegionID) {
	region := r.gs.Regions[rid]
	if region == nil {
		r.editSelectedFaction = ""
		return
	}
	r.editSelectedFaction = faction.FactionID(region.OwnerID)
}

func (r *Renderer) setEditFactionFromArmy(a *army.Army) {
	if a == nil {
		r.editSelectedFaction = ""
		return
	}
	r.editSelectedFaction = faction.FactionID(a.OwnerID)
}

func (r *Renderer) deleteSelectedFaction() {
	f := r.selectedEditFaction()
	if f == nil {
		return
	}
	before := r.worldSnapshot()
	fid := f.ID
	delete(r.gs.Factions, fid)
	for _, region := range r.gs.Regions {
		if region != nil && region.OwnerID == string(fid) {
			region.OwnerID = ""
		}
	}
	for aid, a := range r.gs.Armies {
		if a != nil && a.OwnerID == string(fid) {
			delete(r.gs.Armies, aid)
		}
	}
	if r.gs.PlayerFactionID == fid {
		r.gs.PlayerFactionID = ""
	}
	for key, rel := range r.gs.Relations {
		if rel != nil && (rel.FactionA == fid || rel.FactionB == fid) {
			delete(r.gs.Relations, key)
		}
	}
	r.editSelectedFaction = ""
	r.SelectedArmy = ""
	r.worldMap.MarkDirty()
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) openFactionCreateForm() {
	fid := nextFactionID(r.gs)
	form := editFactionFormState{
		show:     true,
		create:   true,
		active:   editFactionFieldNameTR,
		id:       string(fid),
		name:     "New Faction",
		nameTR:   "",
		religion: religion.Catholic,
		color:    editFactionColor(len(r.gs.Factions) + 1),
		playable: true,
		gold:     "500",
		grain:    "100",
		iron:     "50",
		timber:   "50",
		spice:    "0",
		cloth:    "0",
		ai:       "50",
	}
	if f := r.selectedEditFaction(); f != nil {
		form.religion = f.Religion
	}
	r.editFactionForm = form
	r.setFactionFormRelationTarget(firstRelationTarget(r.gs, fid))
}

func (r *Renderer) openFactionEditForm() {
	f := r.selectedEditFaction()
	if f == nil {
		return
	}
	r.editFactionForm = editFactionFormState{
		show:       true,
		create:     false,
		active:     editFactionFieldNameTR,
		originalID: f.ID,
		id:         string(f.ID),
		name:       f.Name,
		nameTR:     f.NameTR,
		religion:   f.Religion,
		color:      f.Color,
		playable:   f.IsPlayable,
		gold:       itoa(f.Gold),
		grain:      itoa(f.Grain),
		iron:       itoa(f.Iron),
		timber:     itoa(f.Timber),
		spice:      itoa(f.Spice),
		cloth:      itoa(f.Cloth),
		ai:         itoa(f.AIAggressiveness),
	}
	r.setFactionFormRelationTarget(firstRelationTarget(r.gs, f.ID))
}

func (r *Renderer) saveFactionForm() bool {
	form := &r.editFactionForm
	fid := faction.FactionID(strings.TrimSpace(form.id))
	if fid == "" {
		form.errorText = "ID bos olamaz."
		return false
	}
	if existing := r.gs.Factions[fid]; existing != nil && (form.create || fid != form.originalID) {
		form.errorText = "Bu faction ID zaten var."
		return false
	}
	name := strings.TrimSpace(form.name)
	nameTR := strings.TrimSpace(form.nameTR)
	if name == "" && nameTR == "" {
		form.errorText = "En az bir isim gir."
		return false
	}
	gold, ok := parseEditInt(form.gold, 0, 999999)
	if !ok {
		form.errorText = "Altin sayisi gecersiz."
		return false
	}
	grain, ok := parseEditInt(form.grain, 0, 999999)
	if !ok {
		form.errorText = "Tahil sayisi gecersiz."
		return false
	}
	iron, ok := parseEditInt(form.iron, 0, 999999)
	if !ok {
		form.errorText = "Demir sayisi gecersiz."
		return false
	}
	timber, ok := parseEditInt(form.timber, 0, 999999)
	if !ok {
		form.errorText = "Kereste sayisi gecersiz."
		return false
	}
	spice, ok := parseEditInt(form.spice, 0, 999999)
	if !ok {
		form.errorText = "Baharat sayisi gecersiz."
		return false
	}
	cloth, ok := parseEditInt(form.cloth, 0, 999999)
	if !ok {
		form.errorText = "Kumas sayisi gecersiz."
		return false
	}
	aiValue, ok := parseEditInt(form.ai, 0, 100)
	if !ok {
		form.errorText = "AI 0-100 araliginda olmali."
		return false
	}
	relationScore, ok := parseEditInt(form.relationScore, -100, 100)
	if !ok && form.relationTarget != "" {
		form.errorText = "Iliski skoru -100 ile 100 arasinda olmali."
		return false
	}

	before := r.worldSnapshot()
	var existingFaction *faction.Faction
	if !form.create && form.originalID != "" {
		existingFaction = r.gs.Factions[form.originalID]
	}
	if !form.create && form.originalID != "" && form.originalID != fid {
		delete(r.gs.Factions, form.originalID)
		r.renameFactionRelations(form.originalID, fid)
		for _, region := range r.gs.Regions {
			if region != nil && region.OwnerID == string(form.originalID) {
				region.OwnerID = string(fid)
			}
		}
		for _, a := range r.gs.Armies {
			if a != nil && a.OwnerID == string(form.originalID) {
				a.OwnerID = string(fid)
			}
		}
		if r.gs.PlayerFactionID == form.originalID {
			r.gs.PlayerFactionID = fid
		}
	}
	next := &faction.Faction{
		ID:               fid,
		Name:             name,
		NameTR:           nameTR,
		Religion:         form.religion,
		Color:            form.color,
		IsPlayable:       form.playable,
		Gold:             gold,
		Grain:            grain,
		Iron:             iron,
		Timber:           timber,
		Spice:            spice,
		Cloth:            cloth,
		AIAggressiveness: aiValue,
	}
	if existingFaction != nil {
		next.IsEliminated = existingFaction.IsEliminated
		next.Research = existingFaction.Research
	}
	r.gs.Factions[fid] = next
	r.ensureRelationsForFaction(fid)
	if form.relationTarget != "" && r.gs.Factions[form.relationTarget] != nil && form.relationTarget != fid {
		r.setRelationValue(fid, form.relationTarget, relationScore, form.relationStance)
	}
	r.editSelectedFaction = fid
	r.worldMap.MarkDirty()
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editFactionForm = editFactionFormState{}
	r.editDirty = true
	return true
}

func (r *Renderer) handleEditFactionFormInput() InputAction {
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.editFactionForm = editFactionFormState{}
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		if r.saveFactionForm() {
			return InputAction{Kind: ActionSaveScenario}
		}
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyBackspace) {
		r.editFactionFormBackspace()
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if r.handleFactionFormClick(fx, fy) {
			return InputAction{Kind: ActionSaveScenario}
		}
	}
	if r.editFactionForm.active != editFactionFieldNone {
		for _, ch := range ebiten.AppendInputChars(nil) {
			r.appendFactionFormRune(ch)
		}
	}
	return InputAction{}
}

func (r *Renderer) handleFactionFormClick(fx, fy float64) bool {
	if !editFactionFormHit(fx, fy) {
		return false
	}
	for field := editFactionFieldID; field <= editFactionFieldAI; field++ {
		if uiRectHit(fx, fy, editFactionFieldRect(field)) {
			r.editFactionForm.active = field
			r.editFactionForm.errorText = ""
			return false
		}
	}
	switch {
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormSave)):
		return r.saveFactionForm()
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormCancel)):
		r.editFactionForm = editFactionFormState{}
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormReligion)):
		r.editFactionForm.religion = nextEditReligion(r.editFactionForm.religion)
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormPlayable)):
		r.editFactionForm.playable = !r.editFactionForm.playable
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormRelationTarget)):
		r.cycleFactionFormRelationTarget()
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormRelationStance)):
		r.editFactionForm.relationStance = nextEditStance(r.editFactionForm.relationStance)
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormRelationScoreMinus)):
		r.adjustFactionFormRelationScore(-10)
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormRelationScorePlus)):
		r.adjustFactionFormRelationScore(10)
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormRedMinus)):
		r.adjustFactionFormColor(0, -10)
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormRedPlus)):
		r.adjustFactionFormColor(0, 10)
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormGreenMinus)):
		r.adjustFactionFormColor(1, -10)
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormGreenPlus)):
		r.adjustFactionFormColor(1, 10)
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormBlueMinus)):
		r.adjustFactionFormColor(2, -10)
	case uiRectHit(fx, fy, editFactionFormButtonRect(editFactionFormBluePlus)):
		r.adjustFactionFormColor(2, 10)
	}
	return false
}

func (r *Renderer) editFactionFormBackspace() {
	switch r.editFactionForm.active {
	case editFactionFieldID:
		r.editFactionForm.id = trimLastRune(r.editFactionForm.id)
	case editFactionFieldName:
		r.editFactionForm.name = trimLastRune(r.editFactionForm.name)
	case editFactionFieldNameTR:
		r.editFactionForm.nameTR = trimLastRune(r.editFactionForm.nameTR)
	case editFactionFieldGold:
		r.editFactionForm.gold = trimLastRune(r.editFactionForm.gold)
	case editFactionFieldGrain:
		r.editFactionForm.grain = trimLastRune(r.editFactionForm.grain)
	case editFactionFieldIron:
		r.editFactionForm.iron = trimLastRune(r.editFactionForm.iron)
	case editFactionFieldTimber:
		r.editFactionForm.timber = trimLastRune(r.editFactionForm.timber)
	case editFactionFieldSpice:
		r.editFactionForm.spice = trimLastRune(r.editFactionForm.spice)
	case editFactionFieldCloth:
		r.editFactionForm.cloth = trimLastRune(r.editFactionForm.cloth)
	case editFactionFieldAI:
		r.editFactionForm.ai = trimLastRune(r.editFactionForm.ai)
	}
}

func (r *Renderer) appendFactionFormRune(ch rune) {
	if r.editFactionForm.active >= editFactionFieldGold && r.editFactionForm.active <= editFactionFieldAI {
		if ch < '0' || ch > '9' {
			return
		}
	}
	switch r.editFactionForm.active {
	case editFactionFieldID:
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' {
			r.editFactionForm.id = limitStringRunes(r.editFactionForm.id+string(ch), 40)
		}
	case editFactionFieldName:
		r.editFactionForm.name = limitStringRunes(r.editFactionForm.name+string(ch), 64)
	case editFactionFieldNameTR:
		r.editFactionForm.nameTR = limitStringRunes(r.editFactionForm.nameTR+string(ch), 64)
	case editFactionFieldGold:
		r.editFactionForm.gold = limitStringRunes(r.editFactionForm.gold+string(ch), 8)
	case editFactionFieldGrain:
		r.editFactionForm.grain = limitStringRunes(r.editFactionForm.grain+string(ch), 8)
	case editFactionFieldIron:
		r.editFactionForm.iron = limitStringRunes(r.editFactionForm.iron+string(ch), 8)
	case editFactionFieldTimber:
		r.editFactionForm.timber = limitStringRunes(r.editFactionForm.timber+string(ch), 8)
	case editFactionFieldSpice:
		r.editFactionForm.spice = limitStringRunes(r.editFactionForm.spice+string(ch), 8)
	case editFactionFieldCloth:
		r.editFactionForm.cloth = limitStringRunes(r.editFactionForm.cloth+string(ch), 8)
	case editFactionFieldAI:
		r.editFactionForm.ai = limitStringRunes(r.editFactionForm.ai+string(ch), 3)
	}
}

func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}
	return string(runes[:len(runes)-1])
}

func limitStringRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func (r *Renderer) adjustFactionFormColor(index int, delta int) {
	value := int(r.editFactionForm.color[index]) + delta
	if value < 0 {
		value = 0
	}
	if value > 255 {
		value = 255
	}
	r.editFactionForm.color[index] = uint8(value)
}

func nextEditReligion(current religion.Type) religion.Type {
	options := []religion.Type{religion.Catholic, religion.Orthodox, religion.Sunni, religion.Shia}
	for i, option := range options {
		if option == current {
			return options[(i+1)%len(options)]
		}
	}
	return religion.Catholic
}

func nextEditStance(current faction.DiplomaticStance) faction.DiplomaticStance {
	options := []faction.DiplomaticStance{
		faction.StancePeace,
		faction.StanceWar,
		faction.StanceAllied,
		faction.StanceTrade,
	}
	for i, option := range options {
		if option == current {
			return options[(i+1)%len(options)]
		}
	}
	return faction.StancePeace
}

func nextFactionID(gs *state.GameState) faction.FactionID {
	for n := len(gs.Factions) + 1; ; n++ {
		fid := faction.FactionID("new_faction_" + itoa(n))
		if _, used := gs.Factions[fid]; !used {
			return fid
		}
	}
}

func editFactionColor(seed int) [3]uint8 {
	return [3]uint8{
		uint8(70 + (seed*53)%160),
		uint8(70 + (seed*97)%150),
		uint8(70 + (seed*139)%150),
	}
}

func parseEditInt(value string, minValue, maxValue int) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n < minValue || n > maxValue {
		return 0, false
	}
	return n, true
}

func firstRelationTarget(gs *state.GameState, self faction.FactionID) faction.FactionID {
	for _, fid := range sortedFactionIDs(gs.Factions) {
		if fid != self {
			return fid
		}
	}
	return ""
}

func sortedFactionIDs(factions map[faction.FactionID]*faction.Faction) []faction.FactionID {
	ids := make([]faction.FactionID, 0, len(factions))
	for fid := range factions {
		ids = append(ids, fid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (r *Renderer) setFactionFormRelationTarget(target faction.FactionID) {
	r.editFactionForm.relationTarget = target
	if target == "" {
		r.editFactionForm.relationScore = "0"
		r.editFactionForm.relationStance = faction.StancePeace
		return
	}
	self := faction.FactionID(strings.TrimSpace(r.editFactionForm.id))
	if self == "" {
		self = r.editFactionForm.originalID
	}
	if rel := r.relationForForm(self, target); rel != nil {
		r.editFactionForm.relationScore = itoa(rel.Score)
		r.editFactionForm.relationStance = rel.Stance
		return
	}
	targetFaction := r.gs.Factions[target]
	score := 0
	stance := faction.StancePeace
	if targetFaction != nil {
		score = religion.Relation(r.editFactionForm.religion, targetFaction.Religion)
		if (r.editFactionForm.religion == religion.Sunni && targetFaction.Religion == religion.Shia) ||
			(r.editFactionForm.religion == religion.Shia && targetFaction.Religion == religion.Sunni) {
			stance = faction.StanceWar
		}
	}
	r.editFactionForm.relationScore = itoa(score)
	r.editFactionForm.relationStance = stance
}

func (r *Renderer) relationForForm(self, target faction.FactionID) *faction.Relation {
	if self == "" || target == "" || self == target {
		return nil
	}
	return r.gs.Relations[faction.RelationKey(self, target)]
}

func (r *Renderer) cycleFactionFormRelationTarget() {
	self := faction.FactionID(strings.TrimSpace(r.editFactionForm.id))
	ids := sortedFactionIDs(r.gs.Factions)
	if len(ids) == 0 {
		r.setFactionFormRelationTarget("")
		return
	}
	current := r.editFactionForm.relationTarget
	start := 0
	for i, fid := range ids {
		if fid == current {
			start = i + 1
			break
		}
	}
	for offset := 0; offset < len(ids); offset++ {
		fid := ids[(start+offset)%len(ids)]
		if fid != self {
			r.setFactionFormRelationTarget(fid)
			return
		}
	}
	r.setFactionFormRelationTarget("")
}

func (r *Renderer) adjustFactionFormRelationScore(delta int) {
	score, ok := parseEditInt(r.editFactionForm.relationScore, -100, 100)
	if !ok {
		score = 0
	}
	score += delta
	if score < -100 {
		score = -100
	}
	if score > 100 {
		score = 100
	}
	r.editFactionForm.relationScore = itoa(score)
}

func (r *Renderer) ensureRelationsForFaction(fid faction.FactionID) {
	if r.gs.Relations == nil {
		r.gs.Relations = make(map[string]*faction.Relation)
	}
	self := r.gs.Factions[fid]
	for otherID, other := range r.gs.Factions {
		if otherID == fid || other == nil {
			continue
		}
		key := faction.RelationKey(fid, otherID)
		if r.gs.Relations[key] != nil {
			continue
		}
		score := 0
		stance := faction.StancePeace
		if self != nil {
			score = religion.Relation(self.Religion, other.Religion)
			if (self.Religion == religion.Sunni && other.Religion == religion.Shia) ||
				(self.Religion == religion.Shia && other.Religion == religion.Sunni) {
				stance = faction.StanceWar
			}
		}
		r.gs.Relations[key] = &faction.Relation{FactionA: fid, FactionB: otherID, Score: score, Stance: stance}
	}
}

func (r *Renderer) setRelationValue(a, b faction.FactionID, score int, stance faction.DiplomaticStance) {
	key := faction.RelationKey(a, b)
	r.gs.Relations[key] = &faction.Relation{FactionA: a, FactionB: b, Score: score, Stance: stance}
}

func (r *Renderer) renameFactionRelations(oldID, newID faction.FactionID) {
	next := make(map[string]*faction.Relation, len(r.gs.Relations))
	for _, rel := range r.gs.Relations {
		if rel == nil {
			continue
		}
		copyRel := *rel
		if copyRel.FactionA == oldID {
			copyRel.FactionA = newID
		}
		if copyRel.FactionB == oldID {
			copyRel.FactionB = newID
		}
		if copyRel.FactionA == copyRel.FactionB {
			continue
		}
		next[faction.RelationKey(copyRel.FactionA, copyRel.FactionB)] = &copyRel
	}
	r.gs.Relations = next
}

func (r *Renderer) moveSelectedArmyToEditRegion() {
	a := r.gs.Armies[r.SelectedArmy]
	region := r.gs.Regions[r.editSelectedRegion]
	if a == nil || region == nil || a.RegionID == region.ID {
		return
	}
	if (a.IsNaval && !region.IsSea) || (!a.IsNaval && region.IsSea) {
		return
	}
	aid := a.ID
	old := a.RegionID
	oldDockedRegion := a.DockedRegionID
	oldDockedSettlement := a.DockedSettlementID
	next := region.ID
	a.RegionID = next
	a.DockedRegionID = ""
	a.DockedSettlementID = ""
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) { rr.setArmyLocation(aid, old, oldDockedRegion, oldDockedSettlement) },
		redo: func(rr *Renderer) { rr.setArmyLocation(aid, next, "", "") },
	})
	r.editDirty = true
}

func (r *Renderer) addEditLandArmy() {
	region := r.gs.Regions[r.editSelectedRegion]
	if !r.canAddEditLandArmy(region) {
		return
	}
	ownerID := r.editOwnerForRegion(region)
	unitTypeID := r.defaultEditUnitType(false)
	if ownerID == "" || unitTypeID == "" {
		return
	}
	before := r.worldSnapshot()
	aid := nextEditArmyID(r.gs)
	r.gs.Armies[aid] = &army.Army{
		ID:            aid,
		OwnerID:       ownerID,
		RegionID:      region.ID,
		Units:         army.MakeUnits(unitTypeID, 1),
		MovePoints:    2,
		MaxMovePoints: 2,
		IsNaval:       false,
	}
	r.SelectedArmy = aid
	r.editSelectedFaction = faction.FactionID(ownerID)
	r.editSelectedUnitType = unitTypeID
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) addEditFleet() {
	region := r.gs.Regions[r.editSelectedRegion]
	if !r.canAddEditFleet(region) {
		return
	}
	ownerID := r.editOwnerForRegion(region)
	seaID := r.editFleetSeaRegion(region)
	unitTypeID := r.defaultEditUnitType(true)
	if ownerID == "" || seaID == "" || unitTypeID == "" {
		return
	}
	before := r.worldSnapshot()
	aid := nextEditArmyID(r.gs)
	r.gs.Armies[aid] = &army.Army{
		ID:                 aid,
		OwnerID:            ownerID,
		RegionID:           seaID,
		DockedRegionID:     region.ID,
		DockedSettlementID: r.editPreferredDockSettlementID(region),
		Units:              army.MakeUnits(unitTypeID, 1),
		MovePoints:         2,
		MaxMovePoints:      2,
		IsNaval:            true,
	}
	r.SelectedArmy = aid
	r.editSelectedFaction = faction.FactionID(ownerID)
	r.editSelectedUnitType = unitTypeID
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) deleteSelectedArmy() {
	a := r.gs.Armies[r.SelectedArmy]
	if a == nil {
		return
	}
	before := r.worldSnapshot()
	delete(r.gs.Armies, a.ID)
	r.SelectedArmy = ""
	r.editSelectedUnitType = ""
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) addSelectedArmyUnit() {
	a := r.gs.Armies[r.SelectedArmy]
	if !r.canAddSelectedArmyUnit() || a == nil {
		return
	}
	before := r.worldSnapshot()
	a.Units = append(a.Units, army.Unit{TypeID: r.editSelectedUnitType, CurrentHP: 100})
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) removeSelectedArmyUnit() {
	a := r.gs.Armies[r.SelectedArmy]
	if !r.canRemoveSelectedArmyUnit() || a == nil {
		return
	}
	before := r.worldSnapshot()
	for i := len(a.Units) - 1; i >= 0; i-- {
		if a.Units[i].TypeID == r.editSelectedUnitType {
			a.Units = append(a.Units[:i], a.Units[i+1:]...)
			break
		}
	}
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) toggleEditUnitTypeDropdown() {
	a := r.gs.Armies[r.SelectedArmy]
	if a == nil {
		r.editUnitTypeDropdown.Close()
		return
	}
	r.ensureEditSelectedUnitType(a)
	dx, dy, _, _ := editOwnerDropdownRect()
	r.editUnitTypeDropdown.SetPosition(dx, dy)
	r.editUnitTypeDropdown.SetOptions(r.editUnitTypeOptions(a.IsNaval), r.editSelectedUnitType)
	r.editUnitTypeDropdown.Toggle()
}

func (r *Renderer) canAddEditLandArmy(region *world.Region) bool {
	return region != nil && !region.IsSea && !region.IsLocked && r.editOwnerForRegion(region) != "" && r.defaultEditUnitType(false) != ""
}

func (r *Renderer) canAddEditFleet(region *world.Region) bool {
	return region != nil && !region.IsSea && r.editOwnerForRegion(region) != "" &&
		r.selectedRegionHasPortSettlement(region) && r.editFleetSeaRegion(region) != "" && r.defaultEditUnitType(true) != ""
}

func (r *Renderer) canAddSelectedArmyUnit() bool {
	a := r.gs.Armies[r.SelectedArmy]
	if a == nil || len(a.Units) >= army.MaxArmySize {
		return false
	}
	r.ensureEditSelectedUnitType(a)
	return r.editSelectedUnitType != "" && r.unitTypeMatchesArmy(a, r.editSelectedUnitType)
}

func (r *Renderer) canRemoveSelectedArmyUnit() bool {
	a := r.gs.Armies[r.SelectedArmy]
	if a == nil || len(a.Units) == 0 {
		return false
	}
	r.ensureEditSelectedUnitType(a)
	for _, u := range a.Units {
		if u.TypeID == r.editSelectedUnitType {
			return true
		}
	}
	return false
}

func (r *Renderer) ensureEditSelectedUnitType(a *army.Army) {
	if a == nil {
		r.editSelectedUnitType = ""
		return
	}
	if r.editSelectedUnitType != "" && r.unitTypeMatchesArmy(a, r.editSelectedUnitType) {
		return
	}
	if len(a.Units) > 0 && r.unitTypeMatchesArmy(a, a.Units[0].TypeID) {
		r.editSelectedUnitType = a.Units[0].TypeID
		return
	}
	r.editSelectedUnitType = r.defaultEditUnitType(a.IsNaval)
}

func (r *Renderer) editUnitTypeOptions(isNaval bool) []string {
	options := make([]string, 0, len(r.gs.UnitTypes))
	for typeID := range r.gs.UnitTypes {
		if r.unitTypeIsNaval(typeID) == isNaval {
			options = append(options, typeID)
		}
	}
	sort.Strings(options)
	return options
}

func (r *Renderer) unitTypeMatchesArmy(a *army.Army, typeID string) bool {
	if a == nil || r.gs.UnitTypes[typeID] == nil {
		return false
	}
	return r.unitTypeIsNaval(typeID) == a.IsNaval
}

func (r *Renderer) unitTypeIsNaval(typeID string) bool {
	utype := r.gs.UnitTypes[typeID]
	return utype != nil && utype.RequiredBldg == "port"
}

func (r *Renderer) defaultEditUnitType(isNaval bool) string {
	preferred := "militia"
	if isNaval {
		preferred = "transport"
	}
	if r.gs.UnitTypes[preferred] != nil && r.unitTypeIsNaval(preferred) == isNaval {
		return preferred
	}
	options := r.editUnitTypeOptions(isNaval)
	if len(options) == 0 {
		return ""
	}
	return options[0]
}

func (r *Renderer) selectedRegionHasPortSettlement(region *world.Region) bool {
	if region == nil {
		return false
	}
	if r.editSelectedSettlement >= 0 && r.editSelectedSettlement < len(region.Settlements) {
		return region.Settlements[r.editSelectedSettlement].Type == world.SettlementPort
	}
	for _, settlement := range region.Settlements {
		if settlement.Type == world.SettlementPort {
			return true
		}
	}
	return false
}

func (r *Renderer) editFleetSeaRegion(region *world.Region) world.RegionID {
	if region == nil {
		return ""
	}
	for _, nid := range region.Neighbors {
		if n := r.gs.Regions[nid]; n != nil && n.IsSea {
			return n.ID
		}
	}
	visual := r.worldMap.VisualNeighbors(region.ID, r.editVisualNeighborBuf[:0])
	for _, nid := range visual {
		if n := r.gs.Regions[nid]; n != nil && n.IsSea {
			return n.ID
		}
	}
	return ""
}

func (r *Renderer) editOwnerForRegion(region *world.Region) string {
	if region != nil && region.OwnerID != "" {
		return region.OwnerID
	}
	if r.editSelectedFaction != "" {
		return string(r.editSelectedFaction)
	}
	return ""
}

func nextEditArmyID(gs *state.GameState) army.ArmyID {
	for i := len(gs.Armies) + 1; ; i++ {
		id := army.ArmyID("army_edit_" + itoa(i))
		if gs.Armies[id] == nil {
			return id
		}
	}
}

func (r *Renderer) setSelectedArmyOwnerFromRegion() {
	a := r.gs.Armies[r.SelectedArmy]
	region := r.gs.Regions[r.editSelectedRegion]
	if a == nil || region == nil || region.OwnerID == "" || a.OwnerID == region.OwnerID {
		return
	}
	aid := a.ID
	old := a.OwnerID
	next := region.OwnerID
	a.OwnerID = next
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) { rr.setArmyOwner(aid, old) },
		redo: func(rr *Renderer) { rr.setArmyOwner(aid, next) },
	})
	r.editDirty = true
}

func (r *Renderer) setArmyLocation(aid army.ArmyID, rid, dockedRegionID world.RegionID, dockedSettlementID string) {
	if a := r.gs.Armies[aid]; a != nil {
		a.RegionID = rid
		a.DockedRegionID = dockedRegionID
		a.DockedSettlementID = dockedSettlementID
		r.SelectedArmy = aid
		r.editSelectedRegion = rid
		r.editSelectedSettlement = -1
	}
}

func (r *Renderer) editPreferredDockSettlementID(region *world.Region) string {
	if region == nil {
		return ""
	}
	if r.editSelectedSettlement >= 0 && r.editSelectedSettlement < len(region.Settlements) {
		settlement := region.Settlements[r.editSelectedSettlement]
		if settlement.Type == world.SettlementPort {
			return settlement.ID
		}
	}
	for _, settlement := range region.Settlements {
		if settlement.Type == world.SettlementPort {
			return settlement.ID
		}
	}
	if len(region.Settlements) > 0 {
		return region.Settlements[0].ID
	}
	return ""
}

func (r *Renderer) setArmyOwner(aid army.ArmyID, ownerID string) {
	if a := r.gs.Armies[aid]; a != nil {
		a.OwnerID = ownerID
		r.SelectedArmy = aid
		r.editSelectedFaction = faction.FactionID(ownerID)
	}
}

func editBoolLabel(value bool) string {
	if value {
		return "evet"
	}
	return "hayir"
}

func (r *Renderer) rebuildEditWorldMap() {
	r.invalidateShapeEditSession()
	r.worldMap = NewWorldMap(r.gs)
	r.buildRegionPaintBaseline()
	r.applyRegionPaintOverrides()
}

func (r *Renderer) buildRegionPaintBaseline() {
	if r.worldMap == nil {
		r.editRegionPaintBaseline = nil
		return
	}
	r.editRegionPaintBaseline = make([]uint16, len(r.worldMap.regionAt))
	copy(r.editRegionPaintBaseline, r.worldMap.regionAt)
}

func (r *Renderer) applyRegionPaintOverrides() {
	if r.worldMap == nil || len(r.editRegionPaintOverrides) == 0 {
		return
	}
	for pIdx, rid := range r.editRegionPaintOverrides {
		r.applyRegionOverride(pIdx, rid)
	}
}

func cloneRegionPaintOverrides(src map[int]world.RegionID) map[int]world.RegionID {
	if src == nil {
		return nil
	}
	dst := make(map[int]world.RegionID, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (r *Renderer) applyRegionOverride(pIdx int, rid world.RegionID) {
	if r.worldMap == nil || pIdx < 0 || pIdx >= len(r.worldMap.regionAt) {
		return
	}
	if rid == "" {
		return
	}
	newIdx, ok := r.worldMap.regionIdx[rid]
	if !ok {
		newIdx = uint16(len(r.worldMap.regionIDs))
		r.worldMap.regionIDs = append(r.worldMap.regionIDs, rid)
		r.worldMap.regionIdx[rid] = newIdx
	}
	oldIdx := r.worldMap.regionAt[pIdx]
	if oldIdx == newIdx {
		return
	}
	if oldIdx != 0 {
		oldID := r.worldMap.regionIDs[oldIdx]
		r.worldMap.regionPx[oldID] = removePixelIndex(r.worldMap.regionPx[oldID], pIdx)
	}
	r.worldMap.regionAt[pIdx] = newIdx
	r.worldMap.regionPx[rid] = append(r.worldMap.regionPx[rid], pIdx)
}

func removePixelIndex(slice []int, value int) []int {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
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

func editCreateRegionModifierPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyControl) ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyControlRight)
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

func nextRegionID(gs *state.GameState) world.RegionID {
	for n := len(gs.Regions) + 1; ; n++ {
		rid := world.RegionID("new_region_" + itoa(n))
		if _, used := gs.Regions[rid]; !used {
			return rid
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
	r.ensureSettlementDragSnapshot(targetID)

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
		r.diplomacyScroll = 0
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
		r.diplomacyScroll = 0
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
		if idx := regionDiplomacyButtonHit(fx, fy, r.gs, r.SelectedRegion); idx >= 0 {
			region := r.gs.Regions[r.SelectedRegion]
			if region != nil && region.OwnerID != "" && region.OwnerID != string(r.gs.PlayerFactionID) {
				if reason := regionDiplomacyButtonDisabledReason(r.gs, region.OwnerID, idx); reason != "" {
					r.ShowCombatResult(reason)
					return InputAction{}
				}
				target := faction.FactionID(region.OwnerID)
				switch idx {
				case 0:
					r.showDiplomacy = false
					return InputAction{Kind: ActionDeclareWar, TargetFaction: target}
				case 1:
					r.showDiplomacy = false
					return InputAction{Kind: ActionProposePeace, TargetFaction: target}
				case 2:
					r.showDiplomacy = false
					return InputAction{Kind: ActionProposeAlliance, TargetFaction: target}
				case 3:
					r.showDiplomacy = false
					return InputAction{Kind: ActionProposeTrade, TargetFaction: target}
				}
			}
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
			// Deniz bölgesi sol tıkta sadece seçilir; hareket sağ tıkla verilir.
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
