package render

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strconv"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/audio"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	gameui "mapp-game-go/internal/ui"
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
	confirmDialogW          = float32(460)
	confirmDialogH          = float32(166)
	confirmDialogBtnW       = float32(120)
	confirmDialogBtnH       = float32(36)
	selectedSiegePanelW     = 420.0
	selectedSiegePanelH     = 146.0
	selectedSiegeButtonW    = 170.0
	selectedSiegeButtonH    = 36.0
	regionDoubleClickFrames = 18
	initialCameraZoomFactor = 1.40
	maxCameraZoomScale      = 4.5
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
	SelectedRegion           world.RegionID
	SelectedArmy             army.ArmyID
	selectedFactionPanel     faction.FactionID
	selectedSettlementRegion world.RegionID
	selectedSettlementIndex  int
	devNeighborListExpanded  bool
	showRecruitPanel         bool
	recruitUnitID            string
	recruitQty               int
	lastRegionClickID        world.RegionID
	lastRegionClickTick      int

	// Senaryo seçim ekranı
	scenarioCursor int

	// Fraksiyon seçim ekranı
	factionCursor int

	// Diplomasi paneli
	showDiplomacy          bool
	diplomacyFocus         int
	diplomacyScroll        int
	diplomacyActionFocus   int
	diplomacyTargetFaction faction.FactionID

	// Teknoloji paneli
	showTech           bool
	techCursor         int
	techPanX           float64
	techPanY           float64
	techDragging       bool
	techDragLastMX     float64
	techDragLastMY     float64
	techFilterCategory tech.Category

	// Ticaret paneli
	showTrade         bool
	tradeTab          TradeTab
	tradeScroll       int
	tradeFactionFocus int
	tradeGoodFocus    int
	tradeAmount       int
	tradeListFilter   TradeListFilter
	tradeListSort     TradeListSort
	mapMode           MapMode
	animationTick     int

	// Ana menü
	menuTick        int
	HasSave         bool
	HasAutoSave     bool
	CurrentSettings Settings
	LoadingMessage  string
	LoadingProgress int

	// Duraklama menüsü
	pauseCursor int

	// Kayıt/yükleme slot seçim ekranı
	slotCursor        int
	saveSelectMode    bool   // true=kaydetme, false=yükleme
	pendingDeleteSlot string // onay bekleyen slot adı ("" = onay yok)

	// Olay logu (sağ üst panel)
	eventLog            []string
	eventLogDetails     []string
	eventLogCollapsed   bool
	eventCodexEntries   [4][]EventCodexEntry
	showEventCodex      bool
	eventCodexFilter    EventCodexFilter
	eventCodexFocus     int
	eventCodexScroll    int
	eventDetail         string
	showVictoryDetail   bool
	victoryDetailScroll float64
	eventLogScroll      int

	// Savaş / bildirim mesajı (kısa süreli)
	combatLog      string
	combatLogTimer int
	aiTurnActor    string
	aiTurnDetail   string

	// Tarihsel olay tam ekran bildirimi
	historicalEventTitle   string
	historicalEventDesc    string
	historicalEventPrompt  string
	historicalEventChoices []HistoricalEventChoice
	historicalEventFocus   int
	showHistoricalEvent    bool

	// İlk frame kamera başlatma
	firstDraw bool

	// Input state (just-pressed takibi)
	prevKeys  map[ebiten.Key]bool
	prevMouse map[ebiten.MouseButton]bool

	// Genel onay diyaloğu
	warConfirm    warConfirmState
	battlePlan    battlePlanState
	confirmDialog confirmDialogState
	offerCursor   int

	armyIconBuf    []armyIconPos
	regionLabelBuf []settlementDraw
	labelRectBuf   []screenRect
	tradeCorridors []tradeCorridorInfo
	tradeHoverIdx  int
	tradeCenters   []tradeCenterVisual
	tradeCenterIdx int

	editSelectedRegion               world.RegionID
	editSelectedSettlement           int
	editSelectedFaction              faction.FactionID
	editDraggingSettlement           bool
	editDraggingRegion               bool
	editRenaming                     bool
	editTextTarget                   editTextTarget
	editTextRunes                    []rune
	editInspectorTab                 editInspectorTab
	editDirty                        bool
	editVoronoiDebug                 bool
	editOwnerDropdown                *gameui.Dropdown
	editTerrainDropdown              *gameui.Dropdown
	editSettlementTypeDropdown       *gameui.Dropdown
	editUnitTypeDropdown             *gameui.Dropdown
	editSelectedUnitType             string
	armyNeighborBuf                  []world.RegionID
	editVisualNeighborBuf            []world.RegionID
	editBoundaryPixelBuf             []int
	editShapeSession                 *shapeEditSession
	editShapePainting                bool
	editShapeTool                    editShapeTool
	editShapeBrushMode               editShapeBrushMode
	editShapeBrushRadius             int
	editShapeStrokeBefore            *editWorldSnapshot
	editShapeStrokeLastX             int
	editShapeStrokeLastY             int
	editShapeStrokeHasLast           bool
	editShapeStrokeDirty             bool
	editShapeStrokeAffectsLandShapes bool
	editRegionPaintOverrides         map[int]world.RegionID
	editRegionPaintBaseline          []uint16
	editRegionPaintStrokeStart       map[int]uint16
	editRegionPaintStrokeList        []int
	editUndoStack                    []editCommand
	editRedoStack                    []editCommand
	editRegionDragStart              *editRegionCenterSnapshot
	editSettlementDragStart          []editRegionSettlementsSnapshot
	editFactionForm                  editFactionFormState
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

type CameraState struct {
	X     float64
	Y     float64
	Scale float64
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

type HistoricalEventChoice struct {
	Label      string
	Desc       string
	Effect     string
	FollowUp   string
	Conditions string
}

type EventCodexFilter int

const (
	EventCodexAll EventCodexFilter = iota
	EventCodexReady
	EventCodexCalendar
	EventCodexLocked
)

type EventCodexEntry struct {
	Title       string
	Status      string
	DateLabel   string
	Summary     string
	Detail      string
	MonthsUntil int
}

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
	show            bool
	factionName     string
	factionID       string
	pendingArmy     army.ArmyID
	pendingDest     world.RegionID
	pendingEnemy    army.ArmyID
	opensBattlePlan bool
	battleAction    ActionKind
	battleContext   combat.BattleContext
}

func renderTargetRequiresSiegeDecision(gs *state.GameState, attacker *army.Army, target *world.Region) bool {
	return gs != nil &&
		attacker != nil &&
		target != nil &&
		!attacker.IsNaval &&
		target.CanLandEnter() &&
		target.OwnerID != "" &&
		target.OwnerID != attacker.OwnerID &&
		target.IsFortified()
}

type battlePlanState struct {
	show            bool
	actionKind      ActionKind
	battleContext   combat.BattleContext
	pendingArmy     army.ArmyID
	pendingEnemy    army.ArmyID
	pendingDest     world.RegionID
	regionName      string
	defenderName    string
	defenderFaction string
	focus           int
	previews        [3]combat.Preview
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
		editRegionPaintStrokeStart: make(map[int]uint16),
		editRegionPaintStrokeList:  make([]int, 0, 2048),
		editOwnerDropdown:          gameui.NewDropdown(float64(dropX), float64(dropY), float64(dropW), float64(dropH), "Sahip Sec", float64(editOwnerDropdownHeaderH), float64(editOwnerDropdownRowH), editOwnerDropdownVisibleRows),
		editTerrainDropdown:        gameui.NewDropdown(float64(dropX), float64(dropY), float64(dropW), float64(dropH), "Arazi Tipi", float64(editOwnerDropdownHeaderH), float64(editOwnerDropdownRowH), editOwnerDropdownVisibleRows),
		editSettlementTypeDropdown: gameui.NewDropdown(float64(dropX), float64(dropY), float64(dropW), float64(dropH), "Yerlesim Tipi", float64(editOwnerDropdownHeaderH), float64(editOwnerDropdownRowH), editOwnerDropdownVisibleRows),
		editUnitTypeDropdown:       gameui.NewDropdown(float64(dropX), float64(dropY), float64(dropW), float64(dropH), "Birim Tipi", float64(editOwnerDropdownHeaderH), float64(editOwnerDropdownRowH), editOwnerDropdownVisibleRows),
		tradeCorridors:             make([]tradeCorridorInfo, 0, 96),
		tradeHoverIdx:              -1,
		tradeCenters:               make([]tradeCenterVisual, 0, 12),
		tradeCenterIdx:             -1,
		tradeAmount:                5,
		selectedSettlementIndex:    -1,
	}
	r.resetCamera()
	return r
}

func phaseNeedsWorldMap(phase state.Phase) bool {
	switch phase {
	case state.PhasePlayerTurn, state.PhaseAITurn, state.PhaseTurnResolution, state.PhasePauseMenu, state.PhaseEditMode:
		return true
	default:
		return false
	}
}

func (r *Renderer) ensureWorldMap() {
	if r == nil || r.worldMap != nil || !phaseNeedsWorldMap(r.gs.Phase) {
		return
	}
	r.worldMap = NewWorldMap(r.gs)
	r.resetCamera()
}

// resetCamera kamerayı mevcut ScreenWidth/ScreenHeight'e göre dünyayı tam dolduracak şekilde ayarlar.
func (r *Renderer) resetCamera() {
	r.camScale = initialCameraScale()
	r.camX = float64(WorldW) / 2
	// Haritanın üst kenarını ekranın üstüne hizala
	r.camY = ScreenHeight / (2 * r.camScale)
}

func minCameraScale() float64 {
	scaleX := ScreenWidth / float64(WorldW)
	scaleY := ScreenHeight / float64(WorldH)
	return math.Min(scaleX, scaleY)
}

func initialCameraScale() float64 {
	scale := minCameraScale() * initialCameraZoomFactor
	if scale > maxCameraZoomScale {
		return maxCameraZoomScale
	}
	return scale
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

func (r *Renderer) RebuildSettlementAnchors() {
	if r == nil || r.worldMap == nil || r.gs == nil {
		return
	}
	r.worldMap.RebuildSettlementAnchors(r.gs)
}

func (r *Renderer) MarkEditSaved() { r.editDirty = false }

func (r *Renderer) CameraSnapshot() CameraState {
	if r == nil {
		return CameraState{}
	}
	return CameraState{X: r.camX, Y: r.camY, Scale: r.camScale}
}

func (r *Renderer) RestoreCamera(state CameraState) {
	if r == nil {
		return
	}
	minScale := minCameraScale()
	scale := state.Scale
	if scale < minScale {
		scale = minScale
	}
	if scale > maxCameraZoomScale {
		scale = maxCameraZoomScale
	}
	r.camX = state.X
	r.camY = state.Y
	r.camScale = scale
}

func (r *Renderer) CenterCameraOnRegion(rid world.RegionID) bool {
	if r == nil || r.gs == nil || rid == "" {
		return false
	}
	region := r.gs.Regions[rid]
	if region == nil {
		return false
	}
	r.camX = wcX(region.WorldX)
	r.camY = wcY(region.WorldY)
	return true
}

func (r *Renderer) SetAITurnStatus(actor, detail string) {
	r.aiTurnActor = actor
	r.aiTurnDetail = detail
}

func (r *Renderer) ClearAITurnStatus() {
	r.aiTurnActor = ""
	r.aiTurnDetail = ""
}

func (r *Renderer) SetLoadingMessage(message string) {
	r.LoadingMessage = message
}

func (r *Renderer) SetLoadingProgress(progress int) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	r.LoadingProgress = progress
}

// ReloadGameState yükleme sonrası yeni state ve yeni worldmap ile günceller.
// ActiveScenarioPath aktif senaryonun klasör yolu; asset yükleyiciler buradan türetir.
var ActiveScenarioPath string

func (r *Renderer) ReloadGameState(gs *state.GameState) {
	r.ReloadGameStateWithPreparedMap(gs, nil)
}

func (r *Renderer) ReloadGameStateWithPreparedMap(gs *state.GameState, prepared *WorldMap) {
	r.gs = gs
	if gs.ScenarioPath != "" {
		ActiveScenarioPath = gs.ScenarioPath
		// Senaryo değişince asset cache'lerini sıfırla
		buildingSheetLoaded = false
		miniMapLoaded = false
		armySheetLoaded = false
		settlementImageCache = map[string]*ebiten.Image{}
		settlementImageLoaded = map[string]bool{}
	}
	if prepared != nil {
		r.worldMap = FinalizePreparedWorldMap(prepared)
	} else if phaseNeedsWorldMap(gs.Phase) {
		r.worldMap = NewWorldMap(gs)
	} else {
		r.worldMap = nil
	}
	r.invalidateShapeEditSession()
	r.resetCamera()
	r.SelectedRegion = ""
	r.SelectedArmy = ""
	r.selectedFactionPanel = ""
	r.clearSelectedSettlement()
	r.ClearAITurnStatus()
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
	r.AddEventDetail(msg, msg)
}

// AddEventDetail olay loguna başlık ve detay metniyle yeni bir giriş ekler.
func (r *Renderer) AddEventDetail(msg, detail string) {
	r.eventLog = append([]string{msg}, r.eventLog...)
	r.eventLogDetails = append([]string{detail}, r.eventLogDetails...)
	if len(r.eventLog) > maxEventLogEntries {
		r.eventLog = r.eventLog[:maxEventLogEntries]
	}
	if len(r.eventLogDetails) > maxEventLogEntries {
		r.eventLogDetails = r.eventLogDetails[:maxEventLogEntries]
	}
	r.eventLogScroll = 0
}

func (r *Renderer) RemoveEventAt(idx int) {
	if idx < 0 || idx >= len(r.eventLog) {
		return
	}
	r.eventLog = append(r.eventLog[:idx], r.eventLog[idx+1:]...)
	if idx < len(r.eventLogDetails) {
		r.eventLogDetails = append(r.eventLogDetails[:idx], r.eventLogDetails[idx+1:]...)
	}
	r.clampEventLogScroll()
}

func (r *Renderer) EventDetailAt(idx int) string {
	if idx >= 0 && idx < len(r.eventLogDetails) && r.eventLogDetails[idx] != "" {
		return r.eventLogDetails[idx]
	}
	if idx >= 0 && idx < len(r.eventLog) {
		return r.eventLog[idx]
	}
	return ""
}

func (r *Renderer) SetEventCodexEntries(entries [4][]EventCodexEntry) {
	r.eventCodexEntries = entries
	if !r.HasEventCodex() {
		r.showEventCodex = false
		r.eventCodexFilter = EventCodexAll
		r.eventCodexFocus = 0
		r.eventCodexScroll = 0
	}
}

func (r *Renderer) HasEventCodex() bool {
	for _, page := range r.eventCodexEntries {
		if len(page) > 0 {
			return true
		}
	}
	return false
}

func (r *Renderer) OpenEventCodex() {
	if !r.HasEventCodex() {
		return
	}
	r.showEventCodex = true
	r.eventCodexFilter = EventCodexAll
	r.eventCodexFocus = 0
	r.eventCodexScroll = 0
}

func (r *Renderer) CloseEventCodex() {
	r.showEventCodex = false
}

func (r *Renderer) currentEventCodexEntries() []EventCodexEntry {
	return r.eventCodexEntries[int(r.eventCodexFilter)]
}

func (r *Renderer) currentEventCodexEntry() *EventCodexEntry {
	entries := r.currentEventCodexEntries()
	if len(entries) == 0 {
		return nil
	}
	if r.eventCodexFocus < 0 {
		r.eventCodexFocus = 0
	}
	if r.eventCodexFocus >= len(entries) {
		r.eventCodexFocus = len(entries) - 1
	}
	return &entries[r.eventCodexFocus]
}

func (r *Renderer) cycleEventCodexFilter(delta int) {
	if !r.HasEventCodex() {
		return
	}
	count := len(r.eventCodexEntries)
	next := (int(r.eventCodexFilter) + delta + count) % count
	r.eventCodexFilter = EventCodexFilter(next)
	r.eventCodexFocus = 0
	r.eventCodexScroll = 0
}

func (r *Renderer) cycleEventCodexFocus(delta int) {
	entries := r.currentEventCodexEntries()
	if len(entries) == 0 {
		r.eventCodexFocus = 0
		r.eventCodexScroll = 0
		return
	}
	r.eventCodexFocus = (r.eventCodexFocus + delta + len(entries)) % len(entries)
	r.ensureEventCodexFocusVisible()
}

func (r *Renderer) clampEventCodexScroll() {
	maxScroll := eventCodexMaxScroll(len(r.currentEventCodexEntries()))
	if r.eventCodexScroll < 0 {
		r.eventCodexScroll = 0
	}
	if r.eventCodexScroll > maxScroll {
		r.eventCodexScroll = maxScroll
	}
}

func (r *Renderer) ensureEventCodexFocusVisible() {
	visibleCount := eventCodexVisibleCount()
	if visibleCount <= 0 {
		r.eventCodexScroll = 0
		return
	}
	r.clampEventCodexScroll()
	if r.eventCodexFocus < r.eventCodexScroll {
		r.eventCodexScroll = r.eventCodexFocus
	}
	bottom := r.eventCodexScroll + visibleCount - 1
	if r.eventCodexFocus > bottom {
		r.eventCodexScroll = r.eventCodexFocus - visibleCount + 1
	}
	r.clampEventCodexScroll()
}

func (r *Renderer) scrollEventCodex(delta int) {
	r.eventCodexScroll += delta
	r.clampEventCodexScroll()
}

// ShowCombatResult oyun içi kısa uyarı/bilgi mesajını ekranda ~3 saniye gösterir.
func (r *Renderer) ShowCombatResult(msg string) {
	r.combatLog = msg
	r.combatLogTimer = 180
	audio.PlaySound("combat")
}

// ShowHistoricalEvent büyük tarihsel olayı tam ekran popup olarak gösterir.
func (r *Renderer) ShowHistoricalEvent(title, desc, prompt string, choices []HistoricalEventChoice) {
	r.historicalEventTitle = title
	r.historicalEventDesc = desc
	r.historicalEventPrompt = prompt
	r.historicalEventChoices = append(r.historicalEventChoices[:0], choices...)
	r.historicalEventFocus = 0
	r.showHistoricalEvent = true
}

func (r *Renderer) HideHistoricalEvent() {
	r.showHistoricalEvent = false
	r.historicalEventTitle = ""
	r.historicalEventDesc = ""
	r.historicalEventPrompt = ""
	r.historicalEventChoices = r.historicalEventChoices[:0]
	r.historicalEventFocus = 0
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
		DrawLoadingScreen(screen, r.LoadingMessage, r.LoadingProgress, r.menuTick)
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
		r.ensureWorldMap()
		r.worldMap.Refresh(r.gs, r.SelectedRegion, r.mapMode)
		mapOp := &ebiten.DrawImageOptions{}
		r.applyMapGeoM(mapOp, float64(WorldW), float64(WorldH))
		screen.DrawImage(r.worldMap.Image(), mapOp)
		r.menuTick++
		DrawPauseMenu(screen, r.pauseCursor, r.HasSave, r.menuTick, r.CurrentSettings)
		return
	}

	r.ensureWorldMap()

	// Seçili bölge veya donanmanın deniz bölgesini vurgula
	highlightRegion := world.RegionID(r.SelectedRegion)
	if r.SelectedArmy != "" {
		if a, ok := r.gs.Armies[r.SelectedArmy]; ok {
			if reg, ok2 := r.gs.Regions[a.RegionID]; ok2 && reg.IsSea {
				highlightRegion = a.RegionID
			}
		}
	}
	r.worldMap.Refresh(r.gs, highlightRegion, r.mapMode)

	// 1. Üretilen dünya haritası
	mapOp := &ebiten.DrawImageOptions{}
	r.applyMapGeoM(mapOp, float64(WorldW), float64(WorldH))
	screen.DrawImage(r.worldMap.Image(), mapOp)

	// 2. Seçim vurgusu (bölge) kaldırıldı

	tradeOverlayVisible := r.tradeOverlayVisible()

	// 3. Ticaret mod backdrop (rotalar/tabelalar daha sonra en üste çizilecek)
	if tradeOverlayVisible {
		r.drawTradeModeBackdrop(screen)
	}

	var armyPositions []armyIconPos
	if r.mapMode != MapModeTrade {
		armyPositions = r.armyIconPositions()
	}

	// 4. Ordu hareket hedefleri (ticaret modunda gizlenir)
	if r.mapMode != MapModeTrade && r.selectedArmyIsPlayerOwned() {
		r.drawMoveTargets(screen)
	}

	// 5. Bölge etiketleri
	r.drawRegionLabels(screen, armyPositions)
	if r.gs.Phase == state.PhaseEditMode {
		r.drawEditRegionCenters(screen)
		r.drawEditVoronoiDebug(screen)
		r.drawEditShapeOverlay(screen)
	}

	// 6. Ordu ikonları (ticaret modunda gizlenir)
	if r.mapMode != MapModeTrade {
		r.drawArmies(screen, armyPositions)
	}

	// 7. Ticaret rotaları + ticaret merkezi tabelaları (en üst harita katmanı)
	if tradeOverlayVisible {
		r.drawTradeRoutes(screen)
	} else {
		r.clearTradeOverlayHover()
	}

	// 8. UI panelleri
	if r.gs.Phase != state.PhaseEditMode {
		recruitEnabled := RecruitPanelButtonEnabled(r.gs, r.SelectedRegion) && !r.isSettlementPanelOpen()
		DrawBottomPanel(screen, r.gs, r.showRecruitPanel, recruitEnabled, r.showDiplomacy, r.showTech, r.mapMode)
		DrawRegionPanelExpanded(screen, r.gs, r.SelectedRegion, r.devNeighborListExpanded)
		if region, settlement, ok := r.selectedSettlement(); ok && region.ID == r.SelectedRegion {
			DrawSettlementPanel(screen, r.gs, region, settlement)
		}
		DrawFactionDetailPanel(screen, r.gs, r.selectedFactionPanel)
		if r.mapMode != MapModeTrade && r.showRecruitPanel {
			DrawRecruitPanel(screen, r.gs, r.SelectedRegion, r.recruitUnitID, r.recruitQty)
		}
		DrawArmyDetailPanel(screen, r.gs, r.SelectedArmy)
		DrawMinimap(screen, r.gs, r.camX, r.camY, r.camScale)
		r.drawSelectedSiegePanel(screen)
	}
	if r.gs.Phase != state.PhaseEditMode {
		DrawEventLog(screen, r.eventLog, r.eventLogCollapsed, r.eventLogScroll, r.HasEventCodex())
		DrawHoverTooltip(screen, r.gs, r.SelectedRegion, r.showRecruitPanel)
	} else {
		r.drawEditModeHud(screen)
		r.drawEditInspector(screen)
		r.drawEditFactionForm(screen)
	}

	// 7. Diplomasi paneli (üst katman)
	if r.showDiplomacy {
		DrawDiplomacyPanel(screen, r.gs, r.diplomacyFocus, r.diplomacyScroll, r.diplomacyActionFocus, r.diplomacyTargetFaction)
	}

	// 8. Teknoloji paneli (üst katman)
	if r.showTech {
		r.DrawTechPanel(screen)
	}

	if r.gs.Phase == state.PhaseAITurn && r.aiTurnActor != "" {
		r.drawAITurnOverlay(screen)
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
	} else if r.battlePlan.show {
		r.drawBattlePlanDialog(screen)
	} else if offerIdx, ok := r.playerDiplomacyOfferIndex(); ok {
		r.drawDiplomacyOfferDialog(screen, offerIdx)
	}

	if r.showEventCodex {
		drawEventCodexPopup(screen, r.eventCodexFilter, r.currentEventCodexEntries(), r.eventCodexFocus, r.eventCodexScroll)
	}

	if r.eventDetail != "" {
		drawEventDetailPopup(screen, r.eventDetail)
	}

	if r.showVictoryDetail {
		drawVictoryDetailPopup(screen, r.gs, r.victoryDetailScroll)
	}

	// 12. Ticaret koridor tooltip'i (en üst katman, trade panel hariç)
	if tradeOverlayVisible && !r.showTrade {
		r.drawTradeHoverTooltip(screen)
	}

	// 13. Ticaret paneli (üst katman)
	if r.showTrade {
		DrawTradePanel(screen, r.gs, r.tradeTab, r.tradeFactionFocus, r.tradeGoodFocus, r.tradeScroll, r.tradeAmount, r.tradeListFilter, r.tradeListSort)
	}

	// 14. Tarihsel olay popup'ı gerçek üst modal olmalı.
	if r.showHistoricalEvent {
		drawHistoricalEventPopup(screen, r.historicalEventTitle, r.historicalEventDesc, r.historicalEventPrompt, r.historicalEventChoices, r.historicalEventFocus)
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

func (r *Renderer) tradeOverlayVisible() bool {
	if r.mapMode != MapModeTrade {
		return false
	}
	if r.showTech || r.showDiplomacy || r.showTrade || r.showEventCodex || r.showVictoryDetail || r.showHistoricalEvent {
		return false
	}
	if r.confirmDialog.show || r.warConfirm.show || r.battlePlan.show || r.eventDetail != "" {
		return false
	}
	if _, ok := r.playerDiplomacyOfferIndex(); ok {
		return false
	}
	return true
}

func (r *Renderer) drawAITurnOverlay(screen *ebiten.Image) {
	if r.aiTurnActor == "" {
		return
	}
	const panelW, panelH = float32(430), float32(84)
	x := float32(ScreenWidth)/2 - panelW/2
	_, turnHudY, _, turnHudH := turnTechHudRect()
	y := turnHudY + turnHudH + 40
	drawRoundedRect(screen, x, y, panelW, panelH, 8, color.RGBA{16, 14, 10, 228})
	drawPanelBorder(screen, x, y, panelW, panelH)
	vector.FillRect(screen, x, y, panelW, 3, color.RGBA{205, 168, 72, 255}, false)
	DrawText(screen, "AI HAMLESİ", float64(x)+16, float64(y)+10, FaceSmall, ColorGray)
	DrawText(screen, r.aiTurnActor, float64(x)+16, float64(y)+30, FaceMed, ColorGold)
	if r.aiTurnDetail != "" {
		drawUIWrappedLabel(screen, gameui.Rect{X: float64(x) + 16, Y: float64(y) + 48, W: float64(panelW - 32)}, r.aiTurnDetail, color.RGBA{230, 222, 204, 255}, gameui.TextSmall, 16, 2)
	}
}

func (r *Renderer) clearTradeOverlayHover() {
	r.tradeCorridors = r.tradeCorridors[:0]
	r.tradeHoverIdx = -1
	r.tradeCenters = r.tradeCenters[:0]
	r.tradeCenterIdx = -1
}

func (r *Renderer) tradeOverlayOccludesPoint(x, y float64) bool {
	if topStatusPanelHit(x, y) || topDateHudHit(x, y) || musicHudHit(x, y) || bottomActionHudHit(x, y) {
		return true
	}
	tx, ty, tw, th := turnTechHudRect()
	if x >= float64(tx) && x <= float64(tx+tw) && y >= float64(ty) && y <= float64(ty+th) {
		return true
	}
	if eventLogPanelHit(x, y, r.eventLogCollapsed) || minimapHit(x, y) {
		return true
	}
	if r.SelectedRegion != "" && regionPanelHit(x, y) {
		return true
	}
	if r.isSettlementPanelOpen() && settlementPanelHit(x, y) {
		return true
	}
	if r.selectedFactionPanel != "" && factionPanelHit(x, y) {
		return true
	}
	if rect, ok := armyDetailPanelRect(r.gs, r.SelectedArmy); ok && rect.Hit(x, y) {
		return true
	}
	for i := range r.tradeCenters {
		c := r.tradeCenters[i]
		if c.labelW <= 0 || c.labelH <= 0 {
			continue
		}
		if x >= c.labelX && x <= c.labelX+c.labelW && y >= c.labelY && y <= c.labelY+c.labelH {
			return true
		}
	}
	if rect, ok := r.tradeHoverTooltipRect(); ok && rect.Hit(x, y) {
		return true
	}
	return false
}

func (r *Renderer) tradeOverlayOccludesSegment(x1, y1, x2, y2 float64) bool {
	for _, t := range [...]float64{0, 0.25, 0.5, 0.75, 1} {
		x := x1 + (x2-x1)*t
		y := y1 + (y2-y1)*t
		if r.tradeOverlayOccludesPoint(x, y) {
			return true
		}
	}
	return false
}

func (r *Renderer) tradeHoverTooltipRect() (gameui.Rect, bool) {
	if r.tradeHoverIdx < 0 || r.tradeHoverIdx >= len(r.tradeCorridors) {
		return gameui.Rect{}, false
	}
	mx, my := ebiten.CursorPosition()
	x := float64(mx + 14)
	y := float64(my + 16)
	w := 292.0
	h := 90.0
	if x+w > ScreenWidth-6 {
		x = float64(mx) - w - 14
	}
	if y+h > ScreenHeight-6 {
		y = float64(my) - h - 12
	}
	return gameui.Rect{X: x, Y: y, W: w, H: h}, true
}

func armyCanEmbark(gs *state.GameState, a *army.Army) bool {
	if gs == nil || a == nil {
		return false
	}
	return a.CanEmbark(gs.UnitTypes)
}

func findFriendlyEmbarkFleet(gs *state.GameState, ownerID string, seaRegionID world.RegionID, unitCount int) *army.Army {
	return findFriendlyEmbarkFleetFromRegion(gs, ownerID, "", seaRegionID, unitCount)
}

func fleetCanEmbarkFromRegion(gs *state.GameState, fleet *army.Army, sourceRegionID world.RegionID) bool {
	if gs == nil || fleet == nil || !fleet.IsNaval || sourceRegionID == "" {
		return false
	}
	if fleet.DockedRegionID == sourceRegionID {
		return true
	}
	src := gs.Regions[sourceRegionID]
	if src == nil {
		return false
	}
	for _, nid := range src.Neighbors {
		if nid == fleet.RegionID {
			return true
		}
	}
	return false
}

func findFriendlyEmbarkFleetFromRegion(gs *state.GameState, ownerID string, sourceRegionID, seaRegionID world.RegionID, unitCount int) *army.Army {
	if gs == nil {
		return nil
	}
	var fallback *army.Army
	for _, candidate := range gs.Armies {
		if candidate == nil || candidate.OwnerID != ownerID || !candidate.IsNaval || candidate.RegionID != seaRegionID {
			continue
		}
		if !candidate.CanEmbarkUnits(gs.UnitTypes, unitCount) {
			continue
		}
		if sourceRegionID != "" && candidate.DockedRegionID == sourceRegionID {
			return candidate
		}
		if fallback == nil {
			fallback = candidate
		}
	}
	return fallback
}

func embarkableFleetForSelectedArmy(gs *state.GameState, selected *army.Army, fleet *army.Army) bool {
	if gs == nil || selected == nil || fleet == nil || selected.IsNaval || fleet.OwnerID != selected.OwnerID || !fleet.IsNaval {
		return false
	}
	return fleetCanEmbarkFromRegion(gs, fleet, selected.RegionID) &&
		selected.CanEmbark(gs.UnitTypes) &&
		fleet.CanEmbarkUnits(gs.UnitTypes, len(selected.Units))
}

func embarkBlockedMessage(gs *state.GameState, a *army.Army) string {
	if gs == nil || a == nil {
		return "Bu ordudaki bazı birimler denizden taşınamaz."
	}
	blockers := a.EmbarkBlockerNames(gs.UnitTypes)
	if len(blockers) == 0 {
		return "Bu ordudaki bazı birimler denizden taşınamaz."
	}
	return "Bu ordu denizden taşınamaz. Uygun olmayan birlikler: " + strings.Join(blockers, ", ") + "."
}

func armyCanEnterRegion(gs *state.GameState, a *army.Army, target *world.Region) bool {
	if target == nil || target.IsLocked || a == nil {
		return false
	}
	if a.IsNaval {
		if navalCanDockAtRegion(gs, a, target) {
			return true
		}
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
		return armyCanEmbark(gs, a) && findFriendlyEmbarkFleetFromRegion(gs, a.OwnerID, a.RegionID, target.ID, len(a.Units)) != nil
	}
	return target.CanLandEnter()
}

func navalShowsFriendlyDisembark(gs *state.GameState, fleet *army.Army, target *world.Region) bool {
	if gs == nil || fleet == nil || target == nil || !fleet.IsNaval || len(fleet.EmbarkedUnits) == 0 || !target.CanLandEnter() {
		return false
	}
	if target.OwnerID == "" || target.OwnerID == fleet.OwnerID {
		return true
	}
	return false
}

func navalCanDockAtRegion(gs *state.GameState, fleet *army.Army, target *world.Region) bool {
	if gs == nil || fleet == nil || target == nil || !fleet.IsNaval || target.IsSea || target.OwnerID == "" {
		return false
	}
	if fleet.OwnerID != target.OwnerID {
		key := faction.RelationKey(faction.FactionID(fleet.OwnerID), faction.FactionID(target.OwnerID))
		rel, ok := gs.Relations[key]
		if !ok || rel.Stance != faction.StanceAllied {
			return false
		}
	}
	return target.HasPortBuilding()
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

type tradeRouteVisual struct {
	factionA string
	factionB string
	goodName string
	amount   int
	bestFlow int
}

type tradeCenterVisual struct {
	id       world.RegionID
	regionID world.RegionID
	nameTR   string
	tier     world.TradeCenterTier
	worldX   float64
	worldY   float64
	x        float64
	y        float64
	labelX   float64
	labelY   float64
	labelW   float64
	labelH   float64
	offMap   bool
}

type tradeCorridorInfo struct {
	fromName string
	toName   string
	amount   int
	factions int
	goods    string
	sx       float64
	sy       float64
	cx       float64
	cy       float64
	dx       float64
	dy       float64
	hitWidth float64
}

func tradeRoutePairKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "|" + b
}

func routeCurveOffset(key string, dist float64) float64 {
	if dist <= 0 {
		return 0
	}
	h := 0
	for i := 0; i < len(key); i++ {
		h = (h*31 + int(key[i])) & 0x7fffffff
	}
	sign := 1.0
	if h%2 == 0 {
		sign = -1.0
	}
	mag := dist * 0.11
	if mag < 18 {
		mag = 18
	}
	if mag > 96 {
		mag = 96
	}
	return sign * mag
}

func quadBezierPoint(x0, y0, cx, cy, x1, y1, t float64) (float64, float64) {
	u := 1 - t
	x := u*u*x0 + 2*u*t*cx + t*t*x1
	y := u*u*y0 + 2*u*t*cy + t*t*y1
	return x, y
}

func (r *Renderer) drawTradeModeBackdrop(screen *ebiten.Image) {
	w := float32(ScreenWidth)
	h := float32(ScreenHeight)
	// Trade modunda haritayı tamamen kapatmak yerine hafif tint uygula.
	vector.FillRect(screen, 0, 0, w, h, color.RGBA{18, 26, 34, 72}, false)

}

func (r *Renderer) buildTradeCenters(maxCenters int) []tradeCenterVisual {
	if maxCenters <= 0 || len(r.gs.TradeCenters.Centers) == 0 {
		return nil
	}
	centers := make([]tradeCenterVisual, 0, maxCenters)
	for _, def := range r.gs.TradeCenters.Centers {
		if len(centers) >= maxCenters {
			break
		}
		if !def.ActiveInYear(r.gs.Year) {
			continue
		}
		if def.OffMap {
			sx, sy := r.worldToScreen(float64(def.WorldX), float64(def.WorldY))
			centers = append(centers, tradeCenterVisual{
				id:     def.ID,
				nameTR: def.NameTR,
				tier:   def.Tier,
				worldX: float64(def.WorldX),
				worldY: float64(def.WorldY),
				x:      sx,
				y:      sy,
				offMap: true,
			})
			continue
		}
		reg := r.gs.Regions[def.ID]
		if reg == nil || reg.IsSea || reg.TradeCapacity <= 0 {
			continue
		}
		sx, sy := r.regionScreenPos(reg)
		centers = append(centers, tradeCenterVisual{
			id:       reg.ID,
			regionID: reg.ID,
			nameTR:   chooseRegionLabel(reg),
			tier:     def.Tier,
			worldX:   float64(reg.WorldX),
			worldY:   float64(reg.WorldY),
			x:        sx,
			y:        sy,
		})
	}
	return centers
}

func chooseRegionLabel(region *world.Region) string {
	if region == nil {
		return ""
	}
	if region.NameTR != "" {
		return region.NameTR
	}
	if region.Name != "" {
		return region.Name
	}
	return string(region.ID)
}

func sqDistPointSegment(px, py, ax, ay, bx, by float64) float64 {
	abx := bx - ax
	aby := by - ay
	den := abx*abx + aby*aby
	if den <= 1e-6 {
		dx := px - ax
		dy := py - ay
		return dx*dx + dy*dy
	}
	t := ((px-ax)*abx + (py-ay)*aby) / den
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	cx := ax + abx*t
	cy := ay + aby*t
	dx := px - cx
	dy := py - cy
	return dx*dx + dy*dy
}

func (r *Renderer) tradeCorridorAt(fx, fy float64) int {
	if r.tradeOverlayOccludesPoint(fx, fy) {
		return -1
	}
	bestIdx := -1
	bestD2 := math.MaxFloat64
	for i := range r.tradeCorridors {
		c := r.tradeCorridors[i]
		segments := 24
		threshold := c.hitWidth * c.hitWidth
		prevX, prevY := quadBezierPoint(c.sx, c.sy, c.cx, c.cy, c.dx, c.dy, 0)
		for s := 1; s <= segments; s++ {
			t := float64(s) / float64(segments)
			x, y := quadBezierPoint(c.sx, c.sy, c.cx, c.cy, c.dx, c.dy, t)
			d2 := sqDistPointSegment(fx, fy, prevX, prevY, x, y)
			if d2 <= threshold && d2 < bestD2 {
				bestD2 = d2
				bestIdx = i
			}
			prevX, prevY = x, y
		}
	}
	return bestIdx
}

func (r *Renderer) tradeCenterAt(fx, fy float64) int {
	if r.tradeOverlayOccludesPoint(fx, fy) {
		return -1
	}
	bestIdx := -1
	best := math.MaxFloat64
	for i := range r.tradeCenters {
		c := r.tradeCenters[i]
		d := math.Hypot(fx-c.x, fy-c.y)
		if d <= 12 && d < best {
			best = d
			bestIdx = i
		}
	}
	return bestIdx
}

func (r *Renderer) updateTradeHover() {
	r.tradeHoverIdx = -1
	r.tradeCenterIdx = -1
	if r.showTrade {
		return
	}
	if r.mapMode != MapModeTrade || (len(r.tradeCorridors) == 0 && len(r.tradeCenters) == 0) {
		return
	}
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	r.tradeHoverIdx = r.tradeCorridorAt(fx, fy)
	if r.tradeHoverIdx < 0 {
		r.tradeCenterIdx = r.tradeCenterAt(fx, fy)
	}
}

func (r *Renderer) drawTradeHoverTooltip(screen *ebiten.Image) {
	if r.tradeHoverIdx < 0 || r.tradeHoverIdx >= len(r.tradeCorridors) {
		return
	}
	c := r.tradeCorridors[r.tradeHoverIdx]
	segments := 28
	for i := 0; i < segments; i++ {
		t1 := float64(i) / float64(segments)
		t2 := float64(i+1) / float64(segments)
		x1, y1 := quadBezierPoint(c.sx, c.sy, c.cx, c.cy, c.dx, c.dy, t1)
		x2, y2 := quadBezierPoint(c.sx, c.sy, c.cx, c.cy, c.dx, c.dy, t2)
		if r.tradeOverlayOccludesSegment(x1, y1, x2, y2) {
			continue
		}
		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 9.0, color.RGBA{255, 228, 144, 56}, false)
		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 3.0, color.RGBA{255, 241, 192, 230}, false)
	}
	if !r.tradeOverlayOccludesPoint(c.sx, c.sy) {
		vector.FillCircle(screen, float32(c.sx), float32(c.sy), 6, color.RGBA{255, 236, 180, 230}, true)
	}
	if !r.tradeOverlayOccludesPoint(c.dx, c.dy) {
		vector.FillCircle(screen, float32(c.dx), float32(c.dy), 6, color.RGBA{255, 236, 180, 230}, true)
	}

	rect, ok := r.tradeHoverTooltipRect()
	if !ok {
		return
	}
	x := float32(rect.X)
	y := float32(rect.Y)
	w := float32(rect.W)
	h := float32(rect.H)
	vector.FillRect(screen, x, y, w, h, color.RGBA{10, 14, 20, 230}, false)
	vector.StrokeRect(screen, x, y, w, h, 1.2, color.RGBA{145, 120, 74, 230}, false)
	DrawText(screen, "Ticaret Koridoru", float64(x)+10, float64(y)+8, FaceSmall, color.RGBA{242, 226, 174, 255})
	DrawText(screen, c.fromName+" ↔ "+c.toName, float64(x)+10, float64(y)+28, FaceSmall, color.RGBA{215, 225, 236, 235})
	DrawText(screen, "Hacim: "+itoa(c.amount)+"/tur   Fraksiyon: "+itoa(c.factions), float64(x)+10, float64(y)+46, FaceSmall, color.RGBA{187, 203, 222, 230})
	DrawText(screen, "Emtia: "+c.goods, float64(x)+10, float64(y)+64, FaceSmall, color.RGBA{197, 190, 168, 230})
}

func (r *Renderer) nearestTradeCenterIndex(region *world.Region, centers []tradeCenterVisual) int {
	if region == nil || len(centers) == 0 {
		return -1
	}
	rx := float64(region.WorldX)
	ry := float64(region.WorldY)
	bestIdx := -1
	bestDist := math.MaxFloat64
	for i, c := range centers {
		if c.offMap {
			continue
		}
		d := math.Hypot(rx-c.worldX, ry-c.worldY)
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	return bestIdx
}

func (r *Renderer) buildTradeCenterAdjacency(centers []tradeCenterVisual) map[int][]int {
	adj := make(map[int][]int, len(centers))
	if len(centers) == 0 {
		return adj
	}
	indexByID := make(map[world.RegionID]int, len(centers))
	for i := range centers {
		indexByID[centers[i].id] = i
	}

	// explicit links from scenario data
	for _, def := range r.gs.TradeCenters.Centers {
		from, ok := indexByID[def.ID]
		if !ok {
			continue
		}
		for _, lid := range def.Links {
			to, ok := indexByID[lid]
			if !ok || to == from {
				continue
			}
			adj[from] = append(adj[from], to)
			adj[to] = append(adj[to], from)
		}
	}

	// dedup + sort
	for i := range centers {
		neighbors := adj[i]
		if len(neighbors) == 0 {
			continue
		}
		sort.Ints(neighbors)
		uniq := neighbors[:0]
		prev := -1
		for _, n := range neighbors {
			if n == prev {
				continue
			}
			prev = n
			uniq = append(uniq, n)
		}
		adj[i] = uniq
	}
	return adj
}

func shortestCenterPath(adj map[int][]int, from, to int) []int {
	if from < 0 || to < 0 {
		return nil
	}
	if from == to {
		return []int{from}
	}
	queue := []int{from}
	prev := map[int]int{from: -1}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, nxt := range adj[cur] {
			if _, seen := prev[nxt]; seen {
				continue
			}
			prev[nxt] = cur
			if nxt == to {
				path := []int{to}
				for p := cur; p >= 0; p = prev[p] {
					path = append(path, p)
				}
				// reverse
				for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
					path[i], path[j] = path[j], path[i]
				}
				return path
			}
			queue = append(queue, nxt)
		}
	}
	return nil
}

// drawTradeRoutes tüm aktif ticaret rotalarını harita üzerinde sade koridorlar olarak çizer.
// Çift yönlü rotalar (A->B ve B->A) tek bir görsel hatta birleştirilir.
// Uzak zoom'da yalnızca oyuncuyla ilgili rotalar gösterilerek çizgi karmaşası azaltılır.
func (r *Renderer) drawTradeRoutes(screen *ebiten.Image) {
	r.animationTick += 12
	if r.camScale < 0.6 {
		return
	}
	playerID := string(r.gs.PlayerFactionID)
	onlyPlayerRoutes := r.camScale < 0.85
	showLabels := r.camScale >= 1.05

	merged := make(map[string]tradeRouteVisual, len(r.gs.TradeRoutes))
	for _, tr := range r.gs.TradeRoutes {
		if tr == nil || tr.FromFactionID == "" || tr.ToFactionID == "" || tr.FromFactionID == tr.ToFactionID {
			continue
		}
		if onlyPlayerRoutes && tr.FromFactionID != playerID && tr.ToFactionID != playerID {
			continue
		}
		key := tradeRoutePairKey(tr.FromFactionID, tr.ToFactionID)
		route := merged[key]
		if route.factionA == "" {
			if tr.FromFactionID < tr.ToFactionID {
				route.factionA = tr.FromFactionID
				route.factionB = tr.ToFactionID
			} else {
				route.factionA = tr.ToFactionID
				route.factionB = tr.FromFactionID
			}
		}
		route.amount += tr.AmountPerTurn
		candidateGood := economy.GoodNameTR(tr.Good)
		if route.goodName == "" || tr.AmountPerTurn > route.bestFlow {
			route.goodName = candidateGood
			route.bestFlow = tr.AmountPerTurn
		}
		merged[key] = route
	}
	centers := r.buildTradeCenters(len(r.gs.TradeCenters.Centers))
	if len(centers) == 0 {
		r.tradeCorridors = r.tradeCorridors[:0]
		r.tradeHoverIdx = -1
		r.tradeCenters = r.tradeCenters[:0]
		r.tradeCenterIdx = -1
		return
	}
	r.tradeCorridors = r.tradeCorridors[:0]
	r.tradeCenters = append(r.tradeCenters[:0], centers...)
	mx, my := ebiten.CursorPosition()
	preFocusCenter := -1
	bestD := 13.0
	for i := range centers {
		d := math.Hypot(float64(mx)-centers[i].x, float64(my)-centers[i].y)
		if d < bestD {
			bestD = d
			preFocusCenter = i
		}
	}
	adj := r.buildTradeCenterAdjacency(centers)
	factionHub := make(map[string]*world.Region, len(merged)*2)
	factionCenter := make(map[string]int, len(merged)*2)
	type linkAgg struct {
		flow     int
		factions map[string]struct{}
		goods    map[string]int
	}
	centerLinkFlow := map[string]*linkAgg{}
	centerSpokeFlow := map[string]int{}
	mergedKeys := make([]string, 0, len(merged))
	for key := range merged {
		mergedKeys = append(mergedKeys, key)
	}
	sort.Strings(mergedKeys)
	for _, key := range mergedKeys {
		route := merged[key]
		if factionHub[route.factionA] == nil {
			factionHub[route.factionA] = r.factionPrimaryRegion(route.factionA)
		}
		if factionHub[route.factionB] == nil {
			factionHub[route.factionB] = r.factionPrimaryRegion(route.factionB)
		}
		ca, ok := factionCenter[route.factionA]
		if !ok {
			ca = r.nearestTradeCenterIndex(factionHub[route.factionA], centers)
			factionCenter[route.factionA] = ca
		}
		cb, ok := factionCenter[route.factionB]
		if !ok {
			cb = r.nearestTradeCenterIndex(factionHub[route.factionB], centers)
			factionCenter[route.factionB] = cb
		}
		if ca >= 0 {
			centerSpokeFlow[route.factionA] += route.amount
		}
		if cb >= 0 {
			centerSpokeFlow[route.factionB] += route.amount
		}
		if ca < 0 || cb < 0 || ca == cb {
			continue
		}
		path := shortestCenterPath(adj, ca, cb)
		if len(path) < 2 {
			continue
		}
		for pi := 0; pi < len(path)-1; pi++ {
			ka, kb := path[pi], path[pi+1]
			if ka > kb {
				ka, kb = kb, ka
			}
			key := itoa(ka) + "|" + itoa(kb)
			agg := centerLinkFlow[key]
			if agg == nil {
				agg = &linkAgg{
					factions: make(map[string]struct{}, 4),
					goods:    make(map[string]int, 4),
				}
				centerLinkFlow[key] = agg
			}
			agg.flow += route.amount
			agg.factions[route.factionA] = struct{}{}
			agg.factions[route.factionB] = struct{}{}
			if route.goodName != "" {
				agg.goods[route.goodName] += route.amount
			}
		}
	}

	// Faction -> trade center spokes (çok hafif)
	if r.camScale >= 0.95 {
		factionIDs := make([]string, 0, len(factionHub))
		for fid := range factionHub {
			factionIDs = append(factionIDs, fid)
		}
		sort.Strings(factionIDs)
		for _, fid := range factionIDs {
			hub := factionHub[fid]
			if hub == nil {
				continue
			}
			centerIdx := factionCenter[fid]
			if centerIdx < 0 || centerIdx >= len(centers) {
				continue
			}
			flow := centerSpokeFlow[fid]
			if flow <= 0 {
				continue
			}
			hx, hy := r.regionScreenPos(hub)
			c := centers[centerIdx]
			w := float32(0.8)
			if flow >= 12 {
				w = 1.2
			}
			col := color.RGBA{180, 195, 220, 62}
			if preFocusCenter >= 0 && centerIdx != preFocusCenter {
				col = color.RGBA{120, 135, 160, 18}
			}
			if !r.tradeOverlayOccludesSegment(hx, hy, c.x, c.y) {
				vector.StrokeLine(screen, float32(hx), float32(hy), float32(c.x), float32(c.y), w, col, false)
			}
		}
	}

	// Trade center <-> trade center corridors (ana ağ)
	linkKeySet := make(map[string]struct{}, len(centerLinkFlow))
	for key := range centerLinkFlow {
		linkKeySet[key] = struct{}{}
	}
	for fromIdx, list := range adj {
		for _, toIdx := range list {
			a, b := fromIdx, toIdx
			if a > b {
				a, b = b, a
			}
			linkKeySet[itoa(a)+"|"+itoa(b)] = struct{}{}
		}
	}
	linkKeys := make([]string, 0, len(linkKeySet))
	for key := range linkKeySet {
		linkKeys = append(linkKeys, key)
	}
	sort.Strings(linkKeys)
	for _, key := range linkKeys {
		agg := centerLinkFlow[key]
		parts := strings.Split(key, "|")
		if len(parts) != 2 {
			continue
		}
		i, errI := strconv.Atoi(parts[0])
		j, errJ := strconv.Atoi(parts[1])
		if errI != nil || errJ != nil {
			continue
		}
		if i < 0 || j < 0 || i >= len(centers) || j >= len(centers) || i == j {
			continue
		}
		amount := 0
		if agg != nil {
			amount = agg.flow
		}
		sx, sy := centers[i].x, centers[i].y
		dx, dy := centers[j].x, centers[j].y
		mx := (sx + dx) / 2
		my := (sy + dy) / 2
		vx := dx - sx
		vy := dy - sy
		dist := math.Hypot(vx, vy)
		if dist < 1 {
			continue
		}
		px := -vy / dist
		py := vx / dist
		curve := routeCurveOffset(key, dist)
		cx := mx + px*curve
		cy := my + py*curve

		glow := color.RGBA{120, 108, 86, 18}
		core := color.RGBA{165, 150, 118, 42}
		coreW := float32(1.0)
		glowW := float32(2.8)
		if amount > 0 {
			alphaScale := min(uint8(80+(amount*12)), 255)
			glow = color.RGBA{255, 224, 138, alphaScale}
			core = color.RGBA{247, 232, 176, alphaScale}
			coreW = 1.5
			glowW = 5.0
			if amount >= 14 {
				coreW = 2.1
				glowW = 7.0
			} else if amount >= 8 {
				coreW = 1.8
				glowW = 6.0
			}
		}

		if preFocusCenter >= 0 && i != preFocusCenter && j != preFocusCenter {
			if amount > 0 {
				glow = color.RGBA{180, 170, 130, 10}
				core = color.RGBA{170, 165, 140, 34}
				coreW = 1.1
				glowW = 3.4
			} else {
				glow = color.RGBA{100, 94, 82, 8}
				core = color.RGBA{120, 116, 104, 22}
			}
		}
		segments := 22
		for i := 0; i < segments; i++ {
			t1 := float64(i) / float64(segments)
			t2 := float64(i+1) / float64(segments)
			x1, y1 := quadBezierPoint(sx, sy, cx, cy, dx, dy, t1)
			x2, y2 := quadBezierPoint(sx, sy, cx, cy, dx, dy, t2)
			if r.tradeOverlayOccludesSegment(x1, y1, x2, y2) {
				continue
			}
			vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), glowW, glow, false)
			vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), coreW, core, false)
		}
		if amount > 0 && showLabels {
			lx, ly := quadBezierPoint(sx, sy, cx, cy, dx, dy, 0.5)
			if !r.tradeOverlayOccludesPoint(lx, ly) {
				qtyStr := itoa(amount) + "/tur"
				tw2 := MeasureText(qtyStr, FaceSmall)
				DrawText(screen, qtyStr, lx-tw2/2, ly-8, FaceSmall, color.RGBA{225, 204, 144, 230})
			}
		}
		goodsList := make([]struct {
			name string
			flow int
		}, 0)
		if agg != nil {
			goodsList = make([]struct {
				name string
				flow int
			}, 0, len(agg.goods))
			for name, flow := range agg.goods {
				goodsList = append(goodsList, struct {
					name string
					flow int
				}{name: name, flow: flow})
			}
		}
		sort.Slice(goodsList, func(a, b int) bool {
			if goodsList[a].flow != goodsList[b].flow {
				return goodsList[a].flow > goodsList[b].flow
			}
			return goodsList[a].name < goodsList[b].name
		})
		goodsSummary := "-"
		if len(goodsList) > 0 {
			goodsSummary = goodsList[0].name
			if len(goodsList) > 1 {
				goodsSummary += ", " + goodsList[1].name
			}
		}
		factionCount := 0
		if agg != nil {
			factionCount = len(agg.factions)
		}
		r.tradeCorridors = append(r.tradeCorridors, tradeCorridorInfo{
			fromName: centers[i].nameTR,
			toName:   centers[j].nameTR,
			amount:   amount,
			factions: factionCount,
			goods:    goodsSummary,
			sx:       sx,
			sy:       sy,
			cx:       cx,
			cy:       cy,
			dx:       dx,
			dy:       dy,
			hitWidth: float64(glowW) + 4,
		})
	}
	r.updateTradeHover()

	focusCenter := r.tradeCenterIdx
	if r.tradeHoverIdx >= 0 && r.tradeHoverIdx < len(r.tradeCorridors) {
		c := r.tradeCorridors[r.tradeHoverIdx]
		for i := range centers {
			if centers[i].nameTR == c.fromName || centers[i].nameTR == c.toName {
				focusCenter = i
				break
			}
		}
	}

	// Compute active volume for each trade center (local capacity + trade route activity)
	centerVolume := make([]int, len(centers))
	for idx, c := range centers {
		vol := 0
		reg := r.gs.Regions[c.regionID]
		if reg != nil && !c.offMap {
			vol += reg.TradeCapacity
		}
		for _, tr := range r.gs.TradeRoutes {
			if tr.ToFactionID != "" && tr.FromFactionID != "" {
				fromHub := r.factionPrimaryRegion(tr.FromFactionID)
				toHub := r.factionPrimaryRegion(tr.ToFactionID)
				if fromHub != nil && toHub != nil {
					ca := r.nearestTradeCenterIndex(fromHub, centers)
					cb := r.nearestTradeCenterIndex(toHub, centers)
					if ca == idx || cb == idx {
						vol += tr.AmountPerTurn
					}
				}
			}
		}
		centerVolume[idx] = vol
	}

	for i := range centers {
		alphaBg := uint8(235)
		alphaBorder := uint8(235)
		alphaText := uint8(255)
		if focusCenter >= 0 {
			isFocus := false
			if i == focusCenter {
				isFocus = true
			} else {
				for _, c := range r.tradeCorridors {
					if c.fromName == centers[i].nameTR || c.toName == centers[i].nameTR {
						if c.fromName == centers[focusCenter].nameTR || c.toName == centers[focusCenter].nameTR {
							isFocus = true
							break
						}
					}
				}
			}
			if !isFocus {
				alphaBg = 100
				alphaBorder = 100
				alphaText = 140
			}
		}

		nameW := float32(MeasureText(centers[i].nameTR, FaceSmall))
		contentW := nameW
		volStr := ""
		if !centers[i].offMap {
			volStr = "Hacim: " + itoa(centerVolume[i])
			volW := float32(MeasureText(volStr, FaceSmall))
			if volW > contentW {
				contentW = volW
			}
		}
		w := contentW + 40 // yatay padding + ikon/kenar payı
		if w < 116 {
			w = 116
		}
		h := float32(38)
		if centers[i].offMap {
			h = 22
		}
		x := float32(centers[i].x) - w/2
		y := float32(centers[i].y) - h/2
		labelRect := gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}
		centers[i].labelX = labelRect.X
		centers[i].labelY = labelRect.Y
		centers[i].labelW = labelRect.W
		centers[i].labelH = labelRect.H
		r.tradeCenters[i].labelX = labelRect.X
		r.tradeCenters[i].labelY = labelRect.Y
		r.tradeCenters[i].labelW = labelRect.W
		r.tradeCenters[i].labelH = labelRect.H
		if topStatusPanelHit(labelRect.X+labelRect.W/2, labelRect.Y+labelRect.H/2) ||
			topDateHudHit(labelRect.X+labelRect.W/2, labelRect.Y+labelRect.H/2) ||
			musicHudHit(labelRect.X+labelRect.W/2, labelRect.Y+labelRect.H/2) ||
			bottomActionHudHit(labelRect.X+labelRect.W/2, labelRect.Y+labelRect.H/2) ||
			eventLogPanelHit(labelRect.X+labelRect.W/2, labelRect.Y+labelRect.H/2, r.eventLogCollapsed) ||
			minimapHit(labelRect.X+labelRect.W/2, labelRect.Y+labelRect.H/2) {
			continue
		}

		// semi-transparent dark wood background
		bgColor := color.RGBA{18, 14, 10, alphaBg}
		if centers[i].offMap {
			bgColor = color.RGBA{14, 18, 22, alphaBg}
		}
		vector.FillRect(screen, x, y, w, h, bgColor, false)

		// Border: gold for primary, bronze for secondary
		borderColor := color.RGBA{197, 160, 89, alphaBorder}
		if centers[i].tier == world.TradeCenterPrimary {
			vector.StrokeRect(screen, x-1, y-1, w+2, h+2, 1.2, color.RGBA{235, 200, 110, alphaBorder}, false)
			vector.StrokeRect(screen, x+1, y+1, w-2, h-2, 0.8, color.RGBA{150, 110, 50, alphaBorder}, false)
		} else {
			borderColor = color.RGBA{160, 130, 90, alphaBorder}
			if centers[i].offMap {
				borderColor = color.RGBA{118, 156, 188, alphaBorder}
			}
		}
		vector.StrokeRect(screen, x, y, w, h, 1.0, borderColor, false)

		// Center Name
		nameCol := color.RGBA{242, 226, 174, alphaText}
		if centers[i].tier == world.TradeCenterPrimary {
			nameCol = color.RGBA{255, 235, 170, alphaText}
		}
		if centers[i].offMap {
			nameCol = color.RGBA{210, 228, 245, alphaText}
		}
		DrawText(screen, centers[i].nameTR, float64(x)+20, float64(y)+4, FaceSmall, nameCol)

		// Volume indicator
		if !centers[i].offMap {
			DrawText(screen, volStr, float64(x)+20, float64(y)+20, FaceSmall, color.RGBA{180, 180, 170, alphaText})
		}
	}
}

// factionPrimaryRegion bir fraksiyonun görsel temsili için ana bölgesini döner.
// Önce başkent settlement'ı olan bölgeyi, yoksa ilk bulunan bölgeyi döner.
func (r *Renderer) factionPrimaryRegion(factionID string) *world.Region {
	candidates := make([]*world.Region, 0, 16)
	if len(r.gs.RegionOrder) > 0 {
		for _, rid := range r.gs.RegionOrder {
			region := r.gs.Regions[rid]
			if region == nil || region.OwnerID != factionID || region.IsSea {
				continue
			}
			candidates = append(candidates, region)
		}
	} else {
		ids := make([]string, 0, len(r.gs.Regions))
		for rid := range r.gs.Regions {
			ids = append(ids, string(rid))
		}
		sort.Strings(ids)
		for _, id := range ids {
			region := r.gs.Regions[world.RegionID(id)]
			if region == nil || region.OwnerID != factionID || region.IsSea {
				continue
			}
			candidates = append(candidates, region)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	best := candidates[0]
	bestCapital := false
	bestScore := -1
	for _, region := range candidates {
		capital := false
		for _, settlement := range region.Settlements {
			if settlement.IsCapital {
				capital = true
				break
			}
		}
		if capital && !bestCapital {
			best = region
			bestCapital = true
			bestScore = region.TradeCapacity
			continue
		}
		if capital == bestCapital && region.TradeCapacity > bestScore {
			best = region
			bestScore = region.TradeCapacity
		}
	}
	return best
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
		canPreviewWarLanding := a.IsNaval &&
			len(a.EmbarkedUnits) > 0 &&
			nRegion.CanLandEnter() &&
			nRegion.OwnerID != "" &&
			nRegion.OwnerID != a.OwnerID
		if !armyCanEnterRegion(r.gs, a, nRegion) && !canPreviewWarLanding {
			continue
		}

		sx, sy := r.regionScreenPos(nRegion)

		var col color.RGBA
		if a.IsNaval {
			if nRegion.CanLandEnter() {
				switch {
				case nRegion.OwnerID != "" && nRegion.OwnerID != a.OwnerID:
					key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(nRegion.OwnerID))
					rel, exists := r.gs.Relations[key]
					if exists && rel.Stance == faction.StanceWar {
						col = color.RGBA{220, 60, 60, 200}
					} else {
						col = color.RGBA{220, 140, 30, 210}
					}
					DrawTextCentered(screen, "WAR", sx, sy-8, FaceSmall, color.RGBA{255, 200, 80, 230})
				case nRegion.OwnerID == "":
					col = color.RGBA{60, 220, 60, 200}
				default:
					col = color.RGBA{80, 160, 255, 160}
				}
				if navalShowsFriendlyDisembark(r.gs, a, nRegion) {
					DrawTextCentered(screen, "IN", sx, sy-8, FaceSmall, color.RGBA{210, 248, 255, 230})
				}
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
			// Baris halindeki dusman bolgeye savas isareti
			if nRegion.OwnerID != "" && nRegion.OwnerID != a.OwnerID {
				key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(nRegion.OwnerID))
				rel, exists := r.gs.Relations[key]
				if !exists || rel.Stance != faction.StanceWar {
					DrawTextCentered(screen, "WAR", sx, sy-8, FaceSmall, color.RGBA{255, 200, 80, 230})
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

type armyIconCoordKey struct {
	X int
	Y int
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
	Priority  int
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

	// Aynı ekran merkezine düşen farklı grupları (örn. kara + donanma)
	// yatayda ayırarak ikon çakışmasını engeller.
	byCoord := map[armyIconCoordKey][]int{}
	for i := range r.armyIconBuf {
		key := armyIconCoordKey{
			X: int(r.armyIconBuf[i].X * 100),
			Y: int(r.armyIconBuf[i].Y * 100),
		}
		byCoord[key] = append(byCoord[key], i)
	}
	for _, idxs := range byCoord {
		if len(idxs) <= 1 {
			continue
		}
		sort.Slice(idxs, func(i, j int) bool {
			ai := r.gs.Armies[r.armyIconBuf[idxs[i]].ArmyID]
			aj := r.gs.Armies[r.armyIconBuf[idxs[j]].ArmyID]
			// Kara solda, donanma sağda dursun.
			if ai != nil && aj != nil && ai.IsNaval != aj.IsNaval {
				return !ai.IsNaval
			}
			return r.armyIconBuf[idxs[i]].ArmyID < r.armyIconBuf[idxs[j]].ArmyID
		})
		baseX := r.armyIconBuf[idxs[0]].X
		n := float32(len(idxs))
		startX := baseX - (n-1)*iconStep/2
		for j, idx := range idxs {
			r.armyIconBuf[idx].X = startX + float32(j)*iconStep
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
	if !a.IsNaval {
		if siege := r.gs.SiegeByArmy(a.ID); siege != nil {
			if region := r.gs.Regions[siege.RegionID]; region != nil {
				if ax, ay, ok := r.landArmyAnchor(region); ok {
					sx, sy := r.worldToScreen(float64(ax), float64(ay))
					return armyDisplayGroupKey{AnchorX: ax, AnchorY: ay, Anchored: true}, float32(sx), float32(sy), true
				}
				sx, sy := r.regionScreenPos(region)
				return armyDisplayGroupKey{RegionID: region.ID}, float32(sx), float32(sy), true
			}
		}
	}
	region := r.gs.Regions[a.RegionID]
	if region == nil {
		return armyDisplayGroupKey{}, 0, 0, false
	}
	if !a.IsNaval && !region.IsSea {
		if ax, ay, ok := r.landArmyAnchor(region); ok {
			sx, sy := r.worldToScreen(float64(ax), float64(ay))
			return armyDisplayGroupKey{AnchorX: ax, AnchorY: ay, Anchored: true}, float32(sx), float32(sy), true
		}
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

func (r *Renderer) landArmyAnchor(region *world.Region) (int, int, bool) {
	if region == nil || region.IsSea {
		return 0, 0, false
	}
	preferredIdx := -1
	fallbackIdx := -1
	for i, settlement := range region.Settlements {
		if settlement.Type == world.SettlementPort {
			continue
		}
		if fallbackIdx < 0 {
			fallbackIdx = i
		}
		if settlement.IsCapital {
			preferredIdx = i
			break
		}
	}
	if preferredIdx < 0 {
		preferredIdx = fallbackIdx
	}
	if preferredIdx >= 0 {
		if ax, ay, ok := r.worldMap.SettlementAnchor(region.ID, preferredIdx); ok {
			return ax, ay, true
		}
	}
	if ax, ay, ok := r.worldMap.PrimarySettlementAnchor(region.ID); ok {
		return ax, ay, true
	}
	return 0, 0, false
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
	selectedArmy := r.gs.Armies[r.SelectedArmy]
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
		r.drawArmyIcon(screen, a.ID, pos.X, pos.Y, fc, unitCount, isSelected, a.IsNaval)
		if embarkableFleetForSelectedArmy(r.gs, selectedArmy, a) {
			vector.StrokeCircle(screen, pos.X, pos.Y, 17, 3, color.RGBA{120, 230, 240, 220}, true)
			DrawTextCentered(screen, "BIN", float64(pos.X), float64(pos.Y)+15, FaceSmall, color.RGBA{210, 248, 255, 230})
		}
	}
}

// drawArmyIcon tek bir ordu ikonunu çizer.
// Kara ordusu → kare, deniz donanması → daire.
func (r *Renderer) drawArmyIcon(screen *ebiten.Image, aid army.ArmyID, cx, cy float32, col color.RGBA, unitCount int, selected bool, isNaval bool) {
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
	drawUIOutlinedLabel(screen, gameui.Rect{X: tx, Y: ty}, countStr, textCol, shadowCol, gameui.TextSmall, gameui.TextAlignStart)
	if isNaval {
		if a := r.gs.Armies[aid]; a != nil && len(a.EmbarkedUnits) > 0 {
			badgeW := float32(14)
			badgeX := cx - badgeW/2
			badgeY := cy - 28
			vector.FillRect(screen, badgeX, badgeY, badgeW, badgeW, color.RGBA{24, 34, 48, 240}, false)
			vector.StrokeRect(screen, badgeX, badgeY, badgeW, badgeW, 1.5, color.RGBA{214, 226, 242, 230}, false)
			embarkedStr := itoa(len(a.EmbarkedUnits))
			if len(a.EmbarkedUnits) > 99 {
				embarkedStr = "99"
			}
			DrawTextCentered(screen, embarkedStr, float64(badgeX+badgeW/2), float64(badgeY+badgeW/2)-5, FaceSmall, color.RGBA{245, 248, 252, 255})
		}
	}
	if siege := r.gs.SiegeByArmy(aid); siege != nil {
		badgeSize := float32(15)
		badgeX := cx + 8
		badgeY := cy - 27
		vector.FillRect(screen, badgeX-badgeSize/2, badgeY-badgeSize/2, badgeSize, badgeSize, color.RGBA{58, 26, 22, 240}, false)
		vector.StrokeRect(screen, badgeX-badgeSize/2, badgeY-badgeSize/2, badgeSize, badgeSize, 1.5, color.RGBA{224, 182, 96, 245}, false)
		gameui.DrawIcon(screen, gameui.IconSword, float64(badgeX-badgeSize/2+1), float64(badgeY-badgeSize/2+1), float64(badgeSize-2), color.RGBA{255, 229, 176, 255})
	}
	if status, ok := r.gs.ArmyLogistics[aid]; ok && status.TotalHPDamage > 0 {
		badgeX := cx + 8
		badgeY := cy - 12
		vector.FillCircle(screen, badgeX, badgeY, 5, color.RGBA{175, 48, 48, 240}, false)
		DrawTextCentered(screen, "!", float64(badgeX), float64(badgeY)-4, FaceSmall, color.RGBA{255, 244, 232, 255})
	}
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
	tradeCenterRegion := map[world.RegionID]struct{}{}
	if r.mapMode == MapModeTrade {
		for _, def := range r.gs.TradeCenters.Centers {
			if !def.ActiveInYear(r.gs.Year) || def.OffMap {
				continue
			}
			tradeCenterRegion[def.ID] = struct{}{}
		}
	}
	for _, region := range r.gs.Regions {
		if region.IsSea || region.IsLocked {
			continue
		}
		if r.mapMode == MapModeTrade {
			if _, isTradeCenter := tradeCenterRegion[region.ID]; isTradeCenter {
				continue
			}
		}
		r.appendSettlementDraws(region)
	}

	sort.SliceStable(r.regionLabelBuf, func(i, j int) bool {
		if r.regionLabelBuf[i].Priority != r.regionLabelBuf[j].Priority {
			return r.regionLabelBuf[i].Priority > r.regionLabelBuf[j].Priority
		}
		if r.regionLabelBuf[i].SY != r.regionLabelBuf[j].SY {
			return r.regionLabelBuf[i].SY < r.regionLabelBuf[j].SY
		}
		if r.regionLabelBuf[i].SX != r.regionLabelBuf[j].SX {
			return r.regionLabelBuf[i].SX < r.regionLabelBuf[j].SX
		}
		return r.regionLabelBuf[i].Region.ID < r.regionLabelBuf[j].Region.ID
	})

	hoverRID, hoverIdx := r.settlementHoverCandidate()
	selectedRID, selectedIdx, selectedOK := r.selectedSettlementIdentity()

	r.labelRectBuf = r.labelRectBuf[:0]
	for _, item := range r.regionLabelBuf {
		settlement := world.Settlement{}
		if item.Region != nil && item.Index >= 0 && item.Index < len(item.Region.Settlements) {
			settlement = item.Region.Settlements[item.Index]
		}
		forceLabel := item.Region != nil && ((selectedOK && item.Region.ID == selectedRID && item.Index == selectedIdx) ||
			(item.Region.ID == hoverRID && item.Index == hoverIdx))
		if !item.DrawLabel && !forceLabel {
			r.drawSettlementMarker(screen, item.Region, settlement, float32(item.SX), float32(item.SY))
			r.drawSettlementSelectionOverlay(screen, settlement, item.Region, float32(item.SX), float32(item.SY))
			continue
		}

		rect := screenRect{X: item.X, Y: item.Y, W: item.W, H: item.H}
		drawText := true
		if !forceLabel {
			for _, used := range r.labelRectBuf {
				if rectIntersects(expandRect(rect, 4), expandRect(used, 4)) {
					drawText = false
					break
				}
			}
		}
		if drawText && !forceLabel {
			for _, pos := range armyPositions {
				armyRect := screenRect{X: float64(pos.X) - 15, Y: float64(pos.Y) - 15, W: 30, H: 30}
				if rectIntersects(expandRect(rect, 3), armyRect) {
					drawText = false
					break
				}
			}
		}

		if drawText {
			variant := gameui.TextSmall
			if r.camScale >= 1.0 {
				variant = gameui.TextMedium
			}
			outlined := gameui.NewOutlinedLabel(gameui.Rect{X: item.X, Y: item.Y}, item.Text, labelCol, shadowCol, variant, gameui.TextAlignStart)
			outlined.Offsets = [][2]float64{{1, 1}}
			outlined.Draw(screen, renderText)
			r.labelRectBuf = append(r.labelRectBuf, rect)
		}

		r.drawSettlementMarker(screen, item.Region, settlement, float32(item.SX), float32(item.SY))
		r.drawSettlementSelectionOverlay(screen, settlement, item.Region, float32(item.SX), float32(item.SY))
	}
}

func (r *Renderer) appendSettlementDraws(region *world.Region) {
	if len(region.Settlements) == 0 {
		sx, sy := r.regionScreenPos(region)
		r.appendSettlementDraw(region, -1, region.NameTR, sx, sy, true, 10)
		return
	}

	for i, settlement := range region.Settlements {
		isPrimary := settlement.IsCapital || i == 0

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
		drawLabel := r.shouldDrawSettlementLabel(settlement, isPrimary)
		r.appendSettlementDraw(region, i, name, sx, sy, drawLabel, settlementLabelPriority(settlement, isPrimary))
	}
}

func (r *Renderer) appendSettlementDraw(region *world.Region, index int, text string, sx, sy float64, drawLabel bool, priority int) {
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
		Region: region,
		Index:  index,
		Text:   text,
		X:      lx,
		// Etiket noktaların altına çizilir; okunabilirlik artar.
		Y:         sy + 16,
		W:         w,
		H:         h,
		SX:        sx,
		SY:        sy,
		DrawLabel: drawLabel,
		Priority:  priority,
	})
}

func settlementLabelPriority(settlement world.Settlement, isPrimary bool) int {
	if settlement.IsCapital {
		return 100
	}
	switch settlement.Type {
	case world.SettlementCity:
		return 90
	case world.SettlementPort:
		return 70
	case world.SettlementFortress:
		return 60
	case world.SettlementTown:
		return 50
	default:
		if isPrimary {
			return 80
		}
		return 40
	}
}

func (r *Renderer) shouldDrawSettlementLabel(settlement world.Settlement, isPrimary bool) bool {
	// Zoom düşükken sadece başkent/şehir etiketleri.
	if r.camScale < 0.8 {
		return settlement.IsCapital || settlement.Type == world.SettlementCity
	}
	// Orta zoomda liman ve kaleler de açılır.
	if r.camScale < 1.05 {
		return settlement.IsCapital || settlement.Type == world.SettlementCity ||
			settlement.Type == world.SettlementPort || settlement.Type == world.SettlementFortress
	}
	// Yüksek zoomda tüm yerleşim etiketleri açılır.
	if r.camScale >= 1.05 {
		return true
	}
	return isPrimary
}

func (r *Renderer) selectedSettlementIdentity() (world.RegionID, int, bool) {
	if region, _, ok := r.selectedSettlement(); ok && region != nil {
		return region.ID, r.selectedSettlementIndex, true
	}
	if r.gs.Phase == state.PhaseEditMode && r.editSelectedRegion != "" && r.editSelectedSettlement >= 0 {
		return r.editSelectedRegion, r.editSelectedSettlement, true
	}
	return "", -1, false
}

func (r *Renderer) settlementHoverCandidate() (world.RegionID, int) {
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	bestRID := world.RegionID("")
	bestIdx := -1
	bestDist := 14.0 * 14.0
	for _, item := range r.regionLabelBuf {
		if item.Region == nil || item.Index < 0 {
			continue
		}
		dx := fx - item.SX
		dy := fy - (item.SY + 4)
		dist := dx*dx + dy*dy
		if dist <= bestDist {
			bestDist = dist
			bestRID = item.Region.ID
			bestIdx = item.Index
		}
	}
	return bestRID, bestIdx
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
	drawUIDropdown(screen, r.editOwnerDropdown)
	drawUIDropdown(screen, r.editTerrainDropdown)
	drawUIDropdown(screen, r.editSettlementTypeDropdown)
	drawUIDropdown(screen, r.editUnitTypeDropdown)
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
		DrawText(screen, economy.FormatResourceAmountTR(economy.ResourceGold, f.Gold)+"  "+economy.FormatResourceAmountTR(economy.ResourceGrain, f.Grain)+"  "+economy.FormatResourceAmountTR(economy.ResourceIron, f.Iron), float64(x)+14, ly, FaceSmall, ColorGray)
		ly += 18
		DrawText(screen, economy.FormatResourceAmountTR(economy.ResourceTimber, f.Timber)+"  "+economy.FormatResourceAmountTR(economy.ResourceSpice, f.Spice)+"  "+economy.FormatResourceAmountTR(economy.ResourceCloth, f.Cloth), float64(x)+14, ly, FaceSmall, ColorGray)
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
	drawUIDropdown(screen, r.editUnitTypeDropdown)
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
	r.drawFactionFormField(screen, editFactionFieldGold, economy.ResourceNameTR(economy.ResourceGold), r.editFactionForm.gold)
	r.drawFactionFormField(screen, editFactionFieldGrain, economy.ResourceNameTR(economy.ResourceGrain), r.editFactionForm.grain)
	r.drawFactionFormField(screen, editFactionFieldIron, economy.ResourceNameTR(economy.ResourceIron), r.editFactionForm.iron)
	r.drawFactionFormField(screen, editFactionFieldTimber, economy.ResourceNameTR(economy.ResourceTimber), r.editFactionForm.timber)
	r.drawFactionFormField(screen, editFactionFieldSpice, economy.ResourceNameTR(economy.ResourceSpice), r.editFactionForm.spice)
	r.drawFactionFormField(screen, editFactionFieldCloth, economy.ResourceNameTR(economy.ResourceCloth), r.editFactionForm.cloth)
	r.drawFactionFormField(screen, editFactionFieldAI, "AI", r.editFactionForm.ai)

	drawEditFactionFormButton(screen, editFactionFormReligion, "Din: "+religion.DisplayNameTR(r.editFactionForm.religion))
	drawEditFactionFormButton(screen, editFactionFormPlayable, "Playable: "+editBoolLabel(r.editFactionForm.playable))
	relationTitle := "Iliski: yok"
	if r.editFactionForm.relationTarget != "" {
		relationTitle = "Iliski: " + string(r.editFactionForm.relationTarget)
	}
	drawEditFactionFormButton(screen, editFactionFormRelationTarget, relationTitle)
	drawEditFactionFormButton(screen, editFactionFormRelationStance, "Durum: "+faction.DiplomaticStanceLabelTR(r.editFactionForm.relationStance))
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

func editRectButton(r uiRect, label string) gameui.Button {
	return gameui.NewButton(r[0], r[1], r[2], r[3], label)
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
	return (gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}).Hit(mx, my)
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

func buildEditFactionFieldButton(field editFactionFormField, label string) gameui.Button {
	return editRectButton(editFactionFieldRect(field), label)
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

func buildEditFactionFormButton(kind editFactionFormButton, label string) gameui.Button {
	return editRectButton(editFactionFormButtonRect(kind), label)
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
	return (gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}).Hit(mx, my)
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

func buildEditInspectorTabButton(tab editInspectorTab, label string) gameui.Button {
	return editRectButton(editInspectorTabRect(tab), label)
}

func buildEditInspectorActionButton(kind editInspectorButton, label string) gameui.Button {
	return editRectButton(editInspectorButtonRect(kind), label)
}

func editInspectorButtonAt(mx, my float64) editInspectorButton {
	if kind := editMapInspectorButtonAt(mx, my); kind != editButtonNone {
		return kind
	}
	return editDataInspectorButtonAt(mx, my)
}

func editMapInspectorButtonAt(mx, my float64) editInspectorButton {
	for kind := editButtonAddSettlement; kind <= editButtonSaveScenario; kind++ {
		if buildEditInspectorActionButton(kind, "").HitTest(mx, my) {
			return kind
		}
	}
	return editButtonNone
}

func editShapeInspectorButtonAt(mx, my float64) editInspectorButton {
	for kind := editButtonShapePaint; kind <= editButtonShapeBrushPlus; kind++ {
		if buildEditInspectorActionButton(kind, "").HitTest(mx, my) {
			return kind
		}
	}
	if buildEditInspectorActionButton(editButtonSaveScenario, "").HitTest(mx, my) {
		return editButtonSaveScenario
	}
	return editButtonNone
}

func editDataInspectorButtonAt(mx, my float64) editInspectorButton {
	for kind := editButtonAddFaction; kind <= editButtonArmyOwnerFromRegion; kind++ {
		if buildEditInspectorActionButton(kind, "").HitTest(mx, my) {
			return kind
		}
	}
	if buildEditInspectorActionButton(editButtonSaveScenario, "").HitTest(mx, my) {
		return editButtonSaveScenario
	}
	return editButtonNone
}

func (r *Renderer) editInspectorActiveButtonAt(mx, my float64) editInspectorButton {
	if buildEditInspectorTabButton(editInspectorMap, "").HitTest(mx, my) ||
		buildEditInspectorTabButton(editInspectorShape, "").HitTest(mx, my) ||
		buildEditInspectorTabButton(editInspectorData, "").HitTest(mx, my) {
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
	r.editOwnerDropdown.SetPosition(float64(dx), float64(dy))
	r.editTerrainDropdown.SetPosition(float64(dx), float64(dy))
	r.editSettlementTypeDropdown.SetPosition(float64(dx), float64(dy))
	r.editUnitTypeDropdown.SetPosition(float64(dx), float64(dy))
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

	if r.editInspectorTab == editInspectorShape && leftJustPressed && r.editShapeHelpPanelHit(fx, fy) {
		return InputAction{}
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
			r.setSelectedRegionOwner(r.editOwnerDropdown.OptionAt(idx))
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
			r.setSelectedRegionTerrain(world.TerrainType(r.editTerrainDropdown.OptionAt(idx)))
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
			r.setSelectedSettlementType(r.editSettlementTypeDropdown.OptionAt(idx))
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
			r.editSelectedUnitType = r.editUnitTypeDropdown.OptionAt(idx)
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
	if buildEditInspectorTabButton(editInspectorMap, "Harita").HitTest(fx, fy) {
		r.editInspectorTab = editInspectorMap
		return InputAction{}, true
	}
	if buildEditInspectorTabButton(editInspectorShape, "Shape").HitTest(fx, fy) {
		r.editInspectorTab = editInspectorShape
		return InputAction{}, true
	}
	if buildEditInspectorTabButton(editInspectorData, "Veri").HitTest(fx, fy) {
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
			r.editSelectedUnitType = r.editUnitTypeDropdown.OptionAt(idx)
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
	r.editOwnerDropdown.SetPosition(float64(dx), float64(dy))
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
	r.editTerrainDropdown.SetPosition(float64(dx), float64(dy))
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
	r.editSettlementTypeDropdown.SetPosition(float64(dx), float64(dy))
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
		form.errorText = economy.ResourceInvalidCountMessageTR(economy.ResourceGold)
		return false
	}
	grain, ok := parseEditInt(form.grain, 0, 999999)
	if !ok {
		form.errorText = economy.ResourceInvalidCountMessageTR(economy.ResourceGrain)
		return false
	}
	iron, ok := parseEditInt(form.iron, 0, 999999)
	if !ok {
		form.errorText = economy.ResourceInvalidCountMessageTR(economy.ResourceIron)
		return false
	}
	timber, ok := parseEditInt(form.timber, 0, 999999)
	if !ok {
		form.errorText = economy.ResourceInvalidCountMessageTR(economy.ResourceTimber)
		return false
	}
	spice, ok := parseEditInt(form.spice, 0, 999999)
	if !ok {
		form.errorText = economy.ResourceInvalidCountMessageTR(economy.ResourceSpice)
		return false
	}
	cloth, ok := parseEditInt(form.cloth, 0, 999999)
	if !ok {
		form.errorText = economy.ResourceInvalidCountMessageTR(economy.ResourceCloth)
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
		if buildEditFactionFieldButton(field, "").HitTest(fx, fy) {
			r.editFactionForm.active = field
			r.editFactionForm.errorText = ""
			return false
		}
	}
	switch {
	case buildEditFactionFormButton(editFactionFormSave, "Kaydet").HitTest(fx, fy):
		return r.saveFactionForm()
	case buildEditFactionFormButton(editFactionFormCancel, "Iptal").HitTest(fx, fy):
		r.editFactionForm = editFactionFormState{}
	case buildEditFactionFormButton(editFactionFormReligion, "").HitTest(fx, fy):
		r.editFactionForm.religion = nextEditReligion(r.editFactionForm.religion)
	case buildEditFactionFormButton(editFactionFormPlayable, "").HitTest(fx, fy):
		r.editFactionForm.playable = !r.editFactionForm.playable
	case buildEditFactionFormButton(editFactionFormRelationTarget, "").HitTest(fx, fy):
		r.cycleFactionFormRelationTarget()
	case buildEditFactionFormButton(editFactionFormRelationStance, "").HitTest(fx, fy):
		r.editFactionForm.relationStance = nextEditStance(r.editFactionForm.relationStance)
	case buildEditFactionFormButton(editFactionFormRelationScoreMinus, "").HitTest(fx, fy):
		r.adjustFactionFormRelationScore(-10)
	case buildEditFactionFormButton(editFactionFormRelationScorePlus, "").HitTest(fx, fy):
		r.adjustFactionFormRelationScore(10)
	case buildEditFactionFormButton(editFactionFormRedMinus, "").HitTest(fx, fy):
		r.adjustFactionFormColor(0, -10)
	case buildEditFactionFormButton(editFactionFormRedPlus, "").HitTest(fx, fy):
		r.adjustFactionFormColor(0, 10)
	case buildEditFactionFormButton(editFactionFormGreenMinus, "").HitTest(fx, fy):
		r.adjustFactionFormColor(1, -10)
	case buildEditFactionFormButton(editFactionFormGreenPlus, "").HitTest(fx, fy):
		r.adjustFactionFormColor(1, 10)
	case buildEditFactionFormButton(editFactionFormBlueMinus, "").HitTest(fx, fy):
		r.adjustFactionFormColor(2, -10)
	case buildEditFactionFormButton(editFactionFormBluePlus, "").HitTest(fx, fy):
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
	return religion.Next(current)
}

func nextEditStance(current faction.DiplomaticStance) faction.DiplomaticStance {
	return faction.NextDiplomaticStance(current)
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
	r.editUnitTypeDropdown.SetPosition(float64(dx), float64(dy))
	r.editUnitTypeDropdown.SetOptions(r.editUnitTypeOptions(a.IsNaval), r.editSelectedUnitType)
	r.editUnitTypeDropdown.Toggle()
}

func (r *Renderer) canAddEditLandArmy(region *world.Region) bool {
	return region != nil && !region.IsSea && !region.IsLocked && r.editOwnerForRegion(region) != "" && r.defaultEditUnitType(false) != ""
}

func (r *Renderer) canAddEditFleet(region *world.Region) bool {
	return region != nil && !region.IsSea && r.editOwnerForRegion(region) != "" &&
		region.HasPortBuilding() && r.selectedRegionHasPortSettlement(region) &&
		r.editFleetSeaRegion(region) != "" && r.defaultEditUnitType(true) != ""
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
	if !regionPaintOverridesEqual(r.editRegionPaintOverrides, r.gs.RegionPaintOverrides) {
		r.applyRegionPaintOverrides()
	}
}

func (r *Renderer) buildRegionPaintBaseline() {
	if r.worldMap == nil {
		r.editRegionPaintBaseline = nil
		return
	}
	if len(r.worldMap.baseRegionAt) == len(r.worldMap.regionAt) {
		r.editRegionPaintBaseline = make([]uint16, len(r.worldMap.baseRegionAt))
		copy(r.editRegionPaintBaseline, r.worldMap.baseRegionAt)
		return
	}
	if len(r.editRegionPaintOverrides) == 0 && len(r.gs.RegionPaintOverrides) == 0 {
		r.editRegionPaintBaseline = make([]uint16, len(r.worldMap.regionAt))
		copy(r.editRegionPaintBaseline, r.worldMap.regionAt)
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

func regionPaintOverridesEqual(a, b map[int]world.RegionID) bool {
	if len(a) != len(b) {
		return false
	}
	for pIdx, rid := range a {
		if b[pIdx] != rid {
			return false
		}
	}
	return true
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
	outerR := float32(5.5)
	innerR := float32(3.5)

	outerCol := r.settlementOwnerColor(region)

	vector.FillCircle(screen, sx, sy+4, outerR, outerCol, true)
	vector.FillCircle(screen, sx, sy+4, innerR, color.RGBA{240, 230, 200, 255}, true)
}

func (r *Renderer) settlementOwnerColor(region *world.Region) color.RGBA {
	outerCol := color.RGBA{220, 220, 220, 200}
	if r.gs == nil || region == nil || region.OwnerID == "" {
		return outerCol
	}
	for fid, f := range r.gs.Factions {
		if string(fid) == region.OwnerID {
			return color.RGBA{f.Color[0], f.Color[1], f.Color[2], 230}
		}
	}
	return outerCol
}

func (r *Renderer) drawSettlementMarker(screen *ebiten.Image, region *world.Region, settlement world.Settlement, sx, sy float32) {
	if region == nil {
		return
	}
	if settlement.Type == world.SettlementFortress {
		r.drawFortressMarker(screen, region, sx, sy)
		return
	}
	if settlement.Type == world.SettlementPort {
		r.drawPortMarker(screen, region, sx, sy)
		return
	}
	r.drawCityDot(screen, region, sx, sy)
}

func (r *Renderer) drawFortressMarker(screen *ebiten.Image, region *world.Region, sx, sy float32) {
	outerCol := r.settlementOwnerColor(region)
	borderCol := color.RGBA{28, 22, 16, 235}
	stoneCol := color.RGBA{240, 230, 200, 255}
	topY := sy - 4
	baseY := sy + 1
	vector.FillRect(screen, sx-7, baseY-1, 14, 8, borderCol, true)
	vector.FillRect(screen, sx-6, baseY, 12, 6, outerCol, true)
	vector.FillRect(screen, sx-5, baseY+1, 10, 4, stoneCol, true)
	vector.FillRect(screen, sx-6, topY, 3, 4, borderCol, true)
	vector.FillRect(screen, sx-2, topY-1, 4, 5, borderCol, true)
	vector.FillRect(screen, sx+3, topY, 3, 4, borderCol, true)
	vector.FillRect(screen, sx-1, baseY+2, 2, 4, borderCol, true)
	vector.FillRect(screen, sx-3, baseY+3, 6, 1, stoneCol, true)
}

func (r *Renderer) drawPortMarker(screen *ebiten.Image, region *world.Region, sx, sy float32) {
	outerCol := r.settlementOwnerColor(region)
	borderCol := color.RGBA{28, 22, 16, 235}
	stoneCol := color.RGBA{240, 230, 200, 255}
	waterCol := color.RGBA{80, 120, 165, 230}
	topY := sy - 5
	baseY := sy + 1
	waterY := sy + 6

	vector.FillRect(screen, sx-8, baseY+2, 16, 2, borderCol, true)
	vector.FillRect(screen, sx-7, baseY+1, 14, 2, outerCol, true)
	vector.FillRect(screen, sx-6, baseY+2, 12, 1, stoneCol, true)

	vector.FillRect(screen, sx-2, topY, 4, 1, borderCol, true)
	vector.FillRect(screen, sx-1, topY-3, 2, 4, borderCol, true)
	vector.FillRect(screen, sx, topY-2, 1, 2, stoneCol, true)
	vector.FillRect(screen, sx+1, topY-2, 3, 2, outerCol, true)

	vector.FillRect(screen, sx-8, baseY+3, 4, 1, borderCol, true)
	vector.FillRect(screen, sx+4, baseY+3, 4, 1, borderCol, true)

	vector.StrokeLine(screen, sx-7, waterY, sx-3, waterY, 1.2, waterCol, true)
	vector.StrokeLine(screen, sx+1, waterY, sx+5, waterY, 1.2, waterCol, true)
	vector.StrokeLine(screen, sx-5, waterY+2, sx-1, waterY+2, 1.2, waterCol, true)
	vector.StrokeLine(screen, sx+2, waterY+2, sx+6, waterY+2, 1.2, waterCol, true)
}

func (r *Renderer) drawSettlementSelectionOverlay(screen *ebiten.Image, settlement world.Settlement, region *world.Region, sx, sy float32) {
	if r.gs == nil || r.gs.Phase != state.PhaseEditMode || region == nil {
		return
	}
	if region.ID != r.editSelectedRegion || r.editSelectedSettlement < 0 || r.editSelectedSettlement >= len(region.Settlements) {
		return
	}
	if settlement.Type == world.SettlementFortress || settlement.Type == world.SettlementPort {
		vector.StrokeRect(screen, sx-9, sy-8, 18, 16, 2, color.RGBA{255, 190, 45, 230}, true)
		return
	}
	vector.StrokeCircle(screen, sx, sy+4, 10, 2, color.RGBA{255, 190, 45, 230}, true)
}

// --- Input ---

// HandleInput kamera ve oyun girişlerini işler, InputAction döner.
func (r *Renderer) HandleInput() InputAction {
	r.updateCursorShape()
	r.updateEditDropdownPositions()

	// Tarihsel olay popup'ı çizimde en üstte olduğundan inputta da ilk öncelik olmalı.
	if r.showHistoricalEvent {
		return r.handleHistoricalEventInput()
	}

	// Onay diyaloğu açıkken normal input engellenir
	if r.confirmDialog.show {
		return r.handleConfirmDialogInput()
	}
	if r.warConfirm.show {
		return r.handleWarConfirmInput()
	}
	if r.battlePlan.show {
		return r.handleBattlePlanInput()
	}
	if offerIdx, ok := r.playerDiplomacyOfferIndex(); ok {
		return r.handleDiplomacyOfferInput(offerIdx)
	}

	// Oyun sonu ekranı inputu
	if r.gs.Phase == state.PhaseGameOver {
		if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyEnter) ||
			r.mouseJustPressed(ebiten.MouseButtonLeft) {
			return InputAction{Kind: ActionBack}
		}
		return InputAction{}
	}

	if r.showEventCodex {
		return r.handleEventCodexInput()
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

	if r.showVictoryDetail {
		mx, my := ebiten.CursorPosition()
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 && victoryDetailScrollHit(float64(mx), float64(my)) {
			r.victoryDetailScroll = clampVictoryDetailScroll(r.gs, r.victoryDetailScroll-wheelY*28)
		}
		if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyEnter) ||
			r.keyJustPressed(ebiten.KeySpace) || (r.mouseJustPressed(ebiten.MouseButtonLeft) &&
			(victoryDetailCloseHit(float64(mx), float64(my)) || !victoryDetailPopupHit(float64(mx), float64(my)))) {
			r.showVictoryDetail = false
			r.victoryDetailScroll = 0
		}
		return InputAction{}
	}

	// Ana menü inputu
	if r.gs.Phase == state.PhaseMainMenu {
		mx, my := ebiten.CursorPosition()
		input := gameui.InputState{
			MouseX:          float64(mx),
			MouseY:          float64(my),
			LeftJustPressed: r.mouseJustPressed(ebiten.MouseButtonLeft),
		}
		return r.handleMainMenuInput(r.HasSave, r.HasAutoSave, input)
	}

	// Ayarlar ekranı inputu
	if r.gs.Phase == state.PhaseSettings {
		return r.handleSettingsInput(&r.CurrentSettings)
	}

	// Senaryo seçim ekranı inputu
	if r.gs.Phase == state.PhaseScenarioSelect {
		mx, my := ebiten.CursorPosition()
		input := gameui.InputState{
			MouseX:          float64(mx),
			MouseY:          float64(my),
			LeftJustPressed: r.mouseJustPressed(ebiten.MouseButtonLeft),
		}
		return r.handleScenarioSelectInput(input)
	}

	// Fraksiyon seçim ekranı inputu
	if r.gs.Phase == "faction_select" {
		mx, my := ebiten.CursorPosition()
		input := gameui.InputState{
			MouseX:          float64(mx),
			MouseY:          float64(my),
			LeftJustPressed: r.mouseJustPressed(ebiten.MouseButtonLeft),
		}
		return r.handleFactionSelectInput(input)
	}

	// Zafer koşulu seçim ekranı inputu
	if r.gs.Phase == "victory_select" {
		mx, my := ebiten.CursorPosition()
		input := gameui.InputState{
			MouseX:          float64(mx),
			MouseY:          float64(my),
			LeftJustPressed: r.mouseJustPressed(ebiten.MouseButtonLeft),
		}
		return r.handleVictorySelectInput(input)
	}

	// Duraklama menüsü inputu
	if r.gs.Phase == state.PhasePauseMenu {
		mx, my := ebiten.CursorPosition()
		input := gameui.InputState{
			MouseX:          float64(mx),
			MouseY:          float64(my),
			LeftJustPressed: r.mouseJustPressed(ebiten.MouseButtonLeft),
		}
		return r.handlePauseMenuInput(input)
	}

	// Kayıt seçim ekranları inputu
	if r.gs.Phase == state.PhaseLoadSelect {
		mx, my := ebiten.CursorPosition()
		input := gameui.InputState{
			MouseX:          float64(mx),
			MouseY:          float64(my),
			LeftJustPressed: r.mouseJustPressed(ebiten.MouseButtonLeft),
		}
		return r.handleSlotSelectInput(false, input)
	}
	if r.gs.Phase == state.PhaseSaveSelect {
		mx, my := ebiten.CursorPosition()
		input := gameui.InputState{
			MouseX:          float64(mx),
			MouseY:          float64(my),
			LeftJustPressed: r.mouseJustPressed(ebiten.MouseButtonLeft),
		}
		return r.handleSlotSelectInput(true, input)
	}
	if r.gs.Phase == state.PhaseEditMode {
		r.ensureWorldMap()
		return r.handleEditModeInput()
	}

	r.ensureWorldMap()

	// Diplomasi paneli açıkken ayrı input
	if r.showDiplomacy {
		if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyTab) {
			if r.diplomacyTargetFaction != "" {
				r.diplomacyTargetFaction = ""
			} else {
				r.showDiplomacy = false
			}
			return InputAction{}
		}
		mx, my := ebiten.CursorPosition()
		_, wheelY := ebiten.Wheel()
		leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
		leftWasPressed := r.prevMouse[ebiten.MouseButtonLeft]
		r.prevMouse[ebiten.MouseButtonLeft] = leftPressed
		input := gameui.InputState{
			MouseX:           float64(mx),
			MouseY:           float64(my),
			LeftPressed:      leftPressed,
			LeftJustPressed:  leftPressed && !leftWasPressed,
			LeftJustReleased: !leftPressed && leftWasPressed,
			WheelY:           wheelY,
		}
		return r.handleDiplomacyInput(input)
	}

	// Ticaret paneli açıkken: ESC veya tıklama kapatır
	if r.showTrade {
		if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyC) {
			r.showTrade = false
			return InputAction{}
		}
		mx, my := ebiten.CursorPosition()
		_, wheelY := ebiten.Wheel()
		leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
		leftWasPressed := r.prevMouse[ebiten.MouseButtonLeft]
		r.prevMouse[ebiten.MouseButtonLeft] = leftPressed
		input := gameui.InputState{
			MouseX:           float64(mx),
			MouseY:           float64(my),
			LeftPressed:      leftPressed,
			LeftJustPressed:  leftPressed && !leftWasPressed,
			LeftJustReleased: !leftPressed && leftWasPressed,
			WheelY:           wheelY,
		}
		return handleTradePanelInput(r, input)
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
			r.clearSelectedSettlement()
			r.showRecruitPanel = false
			r.resetRecruitSelection()
			r.showDiplomacy = false
			r.diplomacyTargetFaction = ""
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
		r.diplomacyActionFocus = 0
		r.diplomacyTargetFaction = ""
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyM) {
		r.mapMode = r.mapMode.Next()
		return InputAction{}
	}
	// T: teknoloji paneli (ticaret paneli açıkken T paneli kapatır)
	if r.keyJustPressed(ebiten.KeyT) {
		if r.showTrade {
			r.showTrade = false
			return InputAction{}
		}
		r.showTech = !r.showTech
		r.techCursor = 0
		r.techDragging = false
		if r.showTech {
			r.techPanX = 0
			r.techPanY = 0
		}
		return InputAction{}
	}
	// C: ticaret paneli (tech paneli açıkken ticareti açar)
	if r.keyJustPressed(ebiten.KeyC) {
		if r.showTech {
			r.showTech = false
		}
		r.showTrade = !r.showTrade
		r.tradeTab = TradeTabRoutes
		r.tradeScroll = 0
		r.tradeFactionFocus = 0
		r.tradeGoodFocus = 0
		r.tradeAmount = 5
		r.tradeListFilter = TradeListAll
		r.tradeListSort = TradeSortDistance
		return InputAction{}
	}
	// Tech panel aktifken girişi yönlendir
	if r.showTech {
		if f := r.gs.Factions[r.gs.PlayerFactionID]; f != nil {
			if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyT) {
				r.showTech = false
				r.techDragging = false
				return InputAction{}
			}
			mx, my := ebiten.CursorPosition()
			_, wheelY := ebiten.Wheel()
			leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
			leftWasPressed := r.prevMouse[ebiten.MouseButtonLeft]
			r.prevMouse[ebiten.MouseButtonLeft] = leftPressed
			input := gameui.InputState{
				MouseX:           float64(mx),
				MouseY:           float64(my),
				LeftPressed:      leftPressed,
				LeftJustPressed:  leftPressed && !leftWasPressed,
				LeftJustReleased: !leftPressed && leftWasPressed,
				WheelY:           wheelY,
			}
			return r.handleTechInput(f, input)
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
func (r *Renderer) handleFactionSelectInput(input gameui.InputState) InputAction {
	factions, _ := selectableFactions(r.gs)
	n := len(factions)
	if n == 0 {
		if r.keyJustPressed(ebiten.KeyEscape) {
			r.factionCursor = 0
			return InputAction{Kind: ActionBack}
		}
		return InputAction{}
	}
	buttons := buildFactionCardButtons(r.gs)

	// Hover ile kart vurgusunu güncelle
	for i, btn := range buttons {
		if btn.HitTest(input.MouseX, input.MouseY) {
			r.factionCursor = i
			break
		}
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) || r.keyJustPressed(ebiten.KeyArrowRight) {
		r.factionCursor = (r.factionCursor + 1) % n
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) || r.keyJustPressed(ebiten.KeyArrowLeft) {
		r.factionCursor = (r.factionCursor - 1 + n) % n
	}
	if r.keyJustPressed(ebiten.KeyTab) {
		next := focusButtonIndex(buttons, r.factionCursor, ebiten.IsKeyPressed(ebiten.KeyShift))
		if next >= 0 && next < n {
			r.factionCursor = next
		}
	}
	if r.keyJustPressed(ebiten.KeyEnter) && r.factionCursor < len(factions) {
		return InputAction{Kind: ActionSelectFaction, TargetFaction: factions[r.factionCursor]}
	}
	if input.LeftJustPressed {
		if buildBackButton().HandleInput(input) {
			r.factionCursor = 0
			return InputAction{Kind: ActionBack}
		}
		for i, btn := range buttons {
			if btn.HandleInput(input) {
				return InputAction{Kind: ActionSelectFaction, TargetFaction: factions[i]}
			}
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
	if r.selectedFactionPanel != "" && factionPanelCloseHit(fx, fy) {
		r.selectedFactionPanel = ""
		return InputAction{}
	}
	if r.settlementPanelCloseHit(fx, fy) {
		r.clearSelectedSettlement()
		return InputAction{}
	}
	if r.SelectedRegion != "" && regionPanelCloseHit(fx, fy) {
		r.SelectedRegion = ""
		r.selectedFactionPanel = ""
		r.devNeighborListExpanded = false
		r.clearSelectedSettlement()
		r.showRecruitPanel = false
		r.resetRecruitSelection()
		return InputAction{}
	}

	if eventLogToggleHit(fx, fy, r.eventLogCollapsed) {
		r.eventLogCollapsed = !r.eventLogCollapsed
		return InputAction{}
	}
	if r.HasEventCodex() && eventLogCodexHit(fx, fy) {
		r.OpenEventCodex()
		return InputAction{Kind: ActionOpenEventCodex}
	}
	if idx := eventLogCloseHit(fx, fy, len(r.eventLog), r.eventLogCollapsed, r.eventLogScroll); idx >= 0 {
		r.RemoveEventAt(idx)
		return InputAction{}
	}
	if idx := eventLogCardHit(fx, fy, len(r.eventLog), r.eventLogCollapsed, r.eventLogScroll); idx >= 0 {
		r.eventDetail = r.EventDetailAt(idx)
		return InputAction{}
	}

	if topDateHudMenuButtonHit(fx, fy) {
		r.pauseCursor = 0
		return InputAction{Kind: ActionOpenPauseMenu}
	}
	if victoryProgressHit(fx, fy) {
		r.showVictoryDetail = true
		r.victoryDetailScroll = 0
		return InputAction{}
	}
	modeButtons := buildMapModeButtons()
	if modeButtons[0].HitTest(fx, fy) {
		r.mapMode = MapModeNormal
		return InputAction{}
	}
	if modeButtons[1].HitTest(fx, fy) {
		r.mapMode = MapModeTrade
		return InputAction{}
	}
	if r.mapMode == MapModeTrade && tradeToggleButtonHit(fx, fy) {
		r.showTech = false
		r.showTrade = !r.showTrade
		r.tradeTab = TradeTabNew
		r.tradeScroll = 0
		r.tradeFactionFocus = 0
		r.tradeGoodFocus = 0
		r.tradeAmount = 5
		r.tradeListFilter = TradeListAll
		r.tradeListSort = TradeSortDistance
		return InputAction{}
	}
	if r.mapMode == MapModeTrade {
		if idx := r.tradeCorridorAt(fx, fy); idx >= 0 && idx < len(r.tradeCorridors) {
			c := r.tradeCorridors[idx]
			r.ShowCombatResult("Koridor: " + c.fromName + " ↔ " + c.toName + " | " + itoa(c.amount) + "/tur | " + itoa(c.factions) + " fraksiyon")
			return InputAction{}
		}
		if cidx := r.tradeCenterAt(fx, fy); cidx >= 0 && cidx < len(r.tradeCenters) {
			centerName := r.tradeCenters[cidx].nameTR
			connected := 0
			total := 0
			for _, c := range r.tradeCorridors {
				if c.fromName == centerName || c.toName == centerName {
					connected++
					total += c.amount
				}
			}
			r.ShowCombatResult("Merkez: " + centerName + " | " + itoa(connected) + " koridor | " + itoa(total) + "/tur")
			return InputAction{}
		}
	}
	if musicHudInteractiveHit(fx, fy) {
		toggleBtn, nextBtn := buildMusicHudButtons(audio.MusicStatusNow().Playing)
		if toggleBtn.HitTest(fx, fy) {
			return InputAction{Kind: ActionToggleMusic}
		}
		if nextBtn.HitTest(fx, fy) {
			return InputAction{Kind: ActionNextMusic}
		}
	}

	// --- Alt panel butonları ---
	bottomButtons := buildBottomActionButtons(RecruitPanelButtonEnabled(r.gs, r.SelectedRegion))
	if bottomButtons[0].HitTest(fx, fy) {
		if RecruitPanelButtonEnabled(r.gs, r.SelectedRegion) && !r.isSettlementPanelOpen() {
			r.showRecruitPanel = !r.showRecruitPanel
			if r.showRecruitPanel {
				r.clearSelectedSettlement()
			}
			r.showDiplomacy = false
			r.showTech = false
		}
		return InputAction{}
	}
	if bottomButtons[1].HitTest(fx, fy) {
		r.showDiplomacy = !r.showDiplomacy
		r.showRecruitPanel = false
		r.showTech = false
		r.diplomacyFocus = 0
		r.diplomacyScroll = 0
		r.diplomacyActionFocus = 0
		r.diplomacyTargetFaction = ""
		return InputAction{}
	}
	if bottomButtons[2].HitTest(fx, fy) {
		r.showTech = !r.showTech
		r.showRecruitPanel = false
		r.showDiplomacy = false
		r.techCursor = 0
		return InputAction{}
	}
	if bottomButtons[3].HitTest(fx, fy) {
		return InputAction{Kind: ActionEndTurn}
	}

	// UI alanlarında tıklama işleme
	if topStatusPanelHit(fx, fy) || topDateHudHit(fx, fy) || bottomActionHudHit(fx, fy) || musicHudHit(fx, fy) ||
		eventLogPanelHit(fx, fy, r.eventLogCollapsed) || minimapHit(fx, fy) {
		return InputAction{}
	}

	if r.SelectedRegion != "" {
		if fid, ok := regionOwnerNameHit(fx, fy, r.gs, r.SelectedRegion); ok {
			r.selectedFactionPanel = fid
			r.clearSelectedSettlement()
			r.showRecruitPanel = false
			r.resetRecruitSelection()
			return InputAction{}
		}
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
				action, ok := regionDiplomacyActionAt(idx)
				if !ok {
					return InputAction{}
				}
				switch action {
				case diplomacy.ActionDeclareWar:
					r.showDiplomacy = false
					return InputAction{Kind: ActionDeclareWar, TargetFaction: target}
				case diplomacy.ActionProposePeace:
					r.showDiplomacy = false
					return InputAction{Kind: ActionProposePeace, TargetFaction: target}
				case diplomacy.ActionProposeAlliance:
					r.showDiplomacy = false
					return InputAction{Kind: ActionProposeAlliance, TargetFaction: target}
				case diplomacy.ActionProposeTrade:
					r.showDiplomacy = false
					return InputAction{Kind: ActionProposeTrade, TargetFaction: target}
				}
			}
		}
		if regionNeighborToggleHit(fx, fy, r.gs, r.SelectedRegion) {
			r.devNeighborListExpanded = !r.devNeighborListExpanded
			return InputAction{}
		}
		if bid := BuildingGridHitTest(fx, fy, r.gs, r.SelectedRegion, r.devNeighborListExpanded); bid != "" {
			return InputAction{Kind: ActionBuild, TargetRegion: r.SelectedRegion, BuildingID: bid}
		}
	}

	if r.mapMode != MapModeTrade && r.showRecruitPanel {
		// Birim oluştur paneli tıklaması — bölge seçiminden önce kontrol edilmeli
		if act := RecruitPanelActionHitTest(fx, fy, r.gs, r.SelectedRegion); act.Kind != RecruitPanelActionNone {
			switch act.Kind {
			case RecruitPanelActionIncrease:
				r.ensureRecruitSelection(act.UnitID)
				if r.recruitQty < 9 {
					r.recruitQty++
				}
				return InputAction{}
			case RecruitPanelActionDecrease:
				r.ensureRecruitSelection(act.UnitID)
				if r.recruitQty > 1 {
					r.recruitQty--
				}
				return InputAction{}
			case RecruitPanelActionRecruit:
				r.ensureRecruitSelection(act.UnitID)
				return InputAction{Kind: ActionRecruitSpecific, TargetRegion: r.SelectedRegion, BuildingID: act.UnitID, Quantity: r.recruitQty}
			case RecruitPanelActionCancel:
				return InputAction{Kind: ActionCancelRecruitOrder, TargetRegion: r.SelectedRegion, BuildingID: act.OrderID}
			case RecruitPanelActionClose:
				r.showRecruitPanel = false
				return InputAction{}
			}
		}
		if RecruitPanelBoundsHit(fx, fy, r.gs, r.SelectedRegion) {
			return InputAction{}
		}
	}
	if _, siege, _, ok := r.selectedSiegePanelState(); ok {
		assaultBtn, liftBtn := buildSelectedSiegeButtons()
		if assaultBtn.HitTest(fx, fy) {
			return InputAction{Kind: ActionAssaultSiege, ArmyID: r.SelectedArmy, TargetRegion: siege.RegionID, BattleStance: combat.BattleStanceBalanced}
		}
		if liftBtn.HitTest(fx, fy) {
			return InputAction{Kind: ActionLiftSiege, ArmyID: r.SelectedArmy, TargetRegion: siege.RegionID}
		}
		if r.selectedSiegePanelHit(fx, fy) {
			return InputAction{}
		}
	}
	if aid, ok := r.armyHitAt(fx, fy); ok {
		if r.SelectedArmy == aid {
			r.SelectedArmy = ""
			return InputAction{}
		}
		r.SelectedArmy = aid
		r.SelectedRegion = ""
		r.selectedFactionPanel = ""
		r.clearSelectedSettlement()
		r.showRecruitPanel = false
		r.resetRecruitSelection()
		return InputAction{Kind: ActionSelectArmy, ArmyID: aid}
	}
	if rid, idx, ok := r.settlementHitAt(fx, fy); ok {
		r.SelectedArmy = ""
		if r.SelectedRegion != rid {
			r.devNeighborListExpanded = false
		}
		r.SelectedRegion = rid
		r.selectedFactionPanel = ""
		r.selectSettlement(rid, idx)
		if !RecruitPanelVisible(r.gs, rid) {
			r.showRecruitPanel = false
		}
		r.resetRecruitSelection()
		return InputAction{}
	}
	if r.settlementPanelHit(fx, fy) {
		return InputAction{}
	}
	if r.selectedFactionPanel != "" && factionPanelHit(fx, fy) {
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

	// Bölge / deniz bölgesi seçimi
	wx, wy := r.screenToWorld(fx, fy)
	rid := r.worldMap.RegionAt(int(wx), int(wy))
	if rid != "" {
		if region, ok := r.gs.Regions[rid]; ok && region.IsSea {
			// Deniz bölgesi sol tıkta sadece seçilir; hareket sağ tıkla verilir.
			r.SelectedArmy = ""
			if r.SelectedRegion != rid {
				r.devNeighborListExpanded = false
			}
			r.SelectedRegion = rid
			r.selectedFactionPanel = ""
			r.clearSelectedSettlement()
			r.showRecruitPanel = false
			r.resetRecruitSelection()
			return InputAction{}
		}
	}
	r.SelectedArmy = ""
	if r.SelectedRegion != rid {
		r.devNeighborListExpanded = false
	}
	r.SelectedRegion = rid
	r.selectedFactionPanel = ""
	r.clearSelectedSettlement()
	// Tek tıkta panel daima kapanır; yalnızca çift tık açar.
	r.showRecruitPanel = false
	isDoubleClick := rid != "" &&
		rid == r.lastRegionClickID &&
		r.animationTick-r.lastRegionClickTick <= regionDoubleClickFrames
	r.lastRegionClickID = rid
	r.lastRegionClickTick = r.animationTick
	if isDoubleClick && RecruitPanelButtonEnabled(r.gs, rid) {
		r.showRecruitPanel = true
		r.clearSelectedSettlement()
		r.showDiplomacy = false
		r.showTech = false
	}
	r.resetRecruitSelection()
	return InputAction{}
}

func (r *Renderer) ensureRecruitSelection(unitID string) {
	if unitID == "" {
		return
	}
	if r.recruitUnitID != unitID {
		r.recruitUnitID = unitID
		r.recruitQty = 1
		return
	}
	if r.recruitQty < 1 {
		r.recruitQty = 1
	}
}

func (r *Renderer) clearSelectedSettlement() {
	r.selectedSettlementRegion = ""
	r.selectedSettlementIndex = -1
}

func (r *Renderer) selectSettlement(rid world.RegionID, idx int) {
	r.selectedSettlementRegion = rid
	r.selectedSettlementIndex = idx
}

func (r *Renderer) selectedSettlement() (*world.Region, *world.Settlement, bool) {
	if r.selectedSettlementRegion == "" || r.selectedSettlementIndex < 0 {
		return nil, nil, false
	}
	region := r.gs.Regions[r.selectedSettlementRegion]
	if region == nil {
		return nil, nil, false
	}
	if r.selectedSettlementIndex >= len(region.Settlements) {
		return nil, nil, false
	}
	return region, &region.Settlements[r.selectedSettlementIndex], true
}

func (r *Renderer) isSettlementPanelOpen() bool {
	region, _, ok := r.selectedSettlement()
	return ok && region != nil && region.ID == r.SelectedRegion
}

func (r *Renderer) settlementPanelHit(mx, my float64) bool {
	return r.isSettlementPanelOpen() && settlementPanelHit(mx, my)
}

func (r *Renderer) settlementPanelCloseHit(mx, my float64) bool {
	return r.isSettlementPanelOpen() && settlementPanelCloseHit(mx, my)
}

func (r *Renderer) armyHitAt(mx, my float64) (army.ArmyID, bool) {
	armyPositions := r.armyIconPositions()
	for i := len(armyPositions) - 1; i >= 0; i-- {
		pos := armyPositions[i]
		dx := mx - float64(pos.X)
		dy := my - float64(pos.Y)
		if math.Sqrt(dx*dx+dy*dy) < 14 {
			return pos.ArmyID, true
		}
	}
	return "", false
}

func (r *Renderer) settlementHitAt(mx, my float64) (world.RegionID, int, bool) {
	bestRID := world.RegionID("")
	bestIdx := -1
	bestDist := math.MaxFloat64

	for i := len(r.regionLabelBuf) - 1; i >= 0; i-- {
		item := r.regionLabelBuf[i]
		if item.Region == nil || item.Index < 0 || item.Index >= len(item.Region.Settlements) {
			continue
		}
		dx := mx - item.SX
		dy := my - (item.SY + 4)
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > 13 {
			continue
		}
		// Aynı pikselde birden çok aday varsa en yakınını seç.
		// Eşitlikte seçili bölgeye öncelik ver.
		if dist < bestDist || (dist == bestDist && item.Region.ID == r.SelectedRegion) {
			bestDist = dist
			bestRID = item.Region.ID
			bestIdx = item.Index
		}
	}
	if bestIdx >= 0 {
		return bestRID, bestIdx, true
	}
	return "", -1, false
}

func (r *Renderer) resetRecruitSelection() {
	r.recruitUnitID = ""
	r.recruitQty = 1
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

	src, srcOK := r.gs.Regions[a.RegionID]
	if !srcOK {
		return InputAction{}
	}
	wx, wy := r.screenToWorld(float64(mx), float64(my))
	rid := r.worldMap.RegionAt(int(wx), int(wy))
	if clickedID, hit := r.armyHitAt(fx, fy); hit {
		if fleet := r.gs.Armies[clickedID]; fleet != nil && !a.IsNaval && fleet.OwnerID == a.OwnerID && fleet.IsNaval {
			if fleetCanEmbarkFromRegion(r.gs, fleet, a.RegionID) {
				if !armyCanEmbark(r.gs, a) {
					r.ShowCombatResult(embarkBlockedMessage(r.gs, a))
					return InputAction{}
				}
				if !fleet.CanEmbarkUnits(r.gs.UnitTypes, len(a.Units)) {
					r.ShowCombatResult("Seçilen filoda yeterli nakliye kapasitesi yok.")
					return InputAction{}
				}
				r.ShowConfirmDialog(
					"Gemiye Bin",
					"Seçili ordu bu nakliye filosuna binsin mi?",
					"Gemiye Bin",
					"Iptal",
					InputAction{Kind: ActionEmbarkArmy, ArmyID: r.SelectedArmy, TargetArmyID: fleet.ID},
					nil,
				)
				return InputAction{}
			}
		}
	}
	if rid == "" {
		return InputAction{}
	}
	// Limana bağlı donanma aynı deniz bölgesine sağ tıklarsa limandan ayrılıp
	// bölgenin deniz merkezine geçiş (undock) emri verebilir.
	if a.IsNaval && a.DockedRegionID != "" && rid == a.RegionID {
		r.SelectedArmy = ""
		return InputAction{Kind: ActionMoveArmy, ArmyID: a.ID, TargetRegion: rid}
	}
	for _, n := range src.Neighbors {
		if n != rid {
			continue
		}
		target, ok := r.gs.Regions[rid]
		if !ok {
			break
		}
		if !a.IsNaval && target.CanLandEnter() && target.OwnerID != "" && target.OwnerID != a.OwnerID && target.IsFortified() {
			if siege := r.gs.SiegeAt(rid); siege != nil && siege.AttackerArmyID != a.ID {
				r.ShowCombatResult("Bu bölge zaten başka bir ordu tarafından kuşatılıyor.")
				return InputAction{}
			}
			if !a.HasSiegeUnits(r.gs.UnitTypes) {
				r.ShowCombatResult("Bu tahkimatı zorlamak için orduda en az bir kuşatma birimi olmalı.")
				return InputAction{}
			}
		}
		enemyArmy := r.gs.SelectBattleDefender(a, rid, a.IsNaval && target.CanNavalEnter())
		battleAction, battleContext, opensBattlePlan := r.battlePlanIntent(a, target, enemyArmy)
		// Düşman kara bölgesi ama savaş yok → onay diyalogu aç.
		// Donanma-deniz hareketinde savaş ilanı zorunlu değil.
		if !(a.IsNaval && target.CanNavalEnter()) && !navalCanDockAtRegion(r.gs, a, target) && target.OwnerID != "" && target.OwnerID != a.OwnerID {
			key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(target.OwnerID))
			rel, exists := r.gs.Relations[key]
			if !exists || rel.Stance != faction.StanceWar {
				name := target.OwnerID
				if f, ok := r.gs.Factions[faction.FactionID(target.OwnerID)]; ok {
					name = f.NameTR
				}
				r.warConfirm = warConfirmState{
					show:            true,
					factionName:     name,
					factionID:       target.OwnerID,
					pendingArmy:     r.SelectedArmy,
					pendingDest:     rid,
					opensBattlePlan: opensBattlePlan,
					battleAction:    battleAction,
					battleContext:   battleContext,
				}
				if enemyArmy != nil {
					r.warConfirm.pendingEnemy = enemyArmy.ID
				}
				return InputAction{}
			}
		}
		if a.IsNaval && navalShowsFriendlyDisembark(r.gs, a, target) {
			r.ShowConfirmDialog(
				"Karaya In",
				"Gemideki birlikler bu bölgeye insin mi?",
				"Karaya In",
				"Iptal",
				InputAction{Kind: ActionDisembarkArmy, ArmyID: r.SelectedArmy, TargetRegion: rid},
				nil,
			)
			return InputAction{}
		}
		if renderTargetRequiresSiegeDecision(r.gs, a, target) {
			r.openSiegeDecision(a, target)
			return InputAction{}
		}
		if opensBattlePlan {
			r.openBattlePlan(a, target, enemyArmy, battleAction, battleContext)
			return InputAction{}
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
		if dy > 0 && r.camScale < maxCameraZoomScale {
			r.camScale *= 1.12
			if r.camScale > maxCameraZoomScale {
				r.camScale = maxCameraZoomScale
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

var battlePlanStances = [3]combat.BattleStance{
	combat.BattleStanceAggressive,
	combat.BattleStanceBalanced,
	combat.BattleStanceDefensive,
}

func (r *Renderer) battlePlanIntent(attacker *army.Army, target *world.Region, defender *army.Army) (ActionKind, combat.BattleContext, bool) {
	if attacker == nil || target == nil || defender == nil {
		return ActionNone, combat.BattleContextLand, false
	}
	if attacker.IsNaval {
		if target.CanNavalEnter() {
			return ActionMoveArmy, combat.BattleContextNaval, true
		}
		if target.CanLandEnter() && len(attacker.EmbarkedUnits) > 0 {
			return ActionDisembarkArmy, combat.BattleContextAmphibious, true
		}
		return ActionNone, combat.BattleContextLand, false
	}
	if target.CanLandEnter() {
		return ActionMoveArmy, combat.BattleContextLand, true
	}
	return ActionNone, combat.BattleContextLand, false
}

func siegeBreachLabelTR(level int) string {
	switch level {
	case 2:
		return "Büyük gedik"
	case 1:
		return "Küçük gedik"
	default:
		return "Gedik yok"
	}
}

func (r *Renderer) openSiegeDecision(attacker *army.Army, target *world.Region) {
	if r == nil || r.gs == nil || attacker == nil || target == nil {
		return
	}
	fortLevel := target.FortificationLevel()
	bestTier := attacker.HighestSiegeTier(r.gs.UnitTypes)
	if active := r.gs.SiegeAt(target.ID); active != nil && active.AttackerArmyID == attacker.ID {
		msg := fmt.Sprintf("%s kuşatması sürüyor. Tahkimat: %d | İlerleme: %d | Durum: %s | Gedik kapasitesi: T%d/T%d.", target.NameTR, active.FortLevel, active.BreachProgress, siegeBreachLabelTR(active.BreachLevel), bestTier, active.FortLevel)
		r.confirmDialog = confirmDialogState{
			show:          true,
			title:         "Kuşatma Kararı",
			message:       msg,
			messageLines:  wrapTextLines(msg, FaceSmall, float64(confirmDialogW)-40),
			acceptLabel:   "Genel Hücum",
			thirdLabel:    "Kuşatmayı Kaldır",
			declineLabel:  "İptal",
			pendingAction: InputAction{Kind: ActionAssaultSiege, ArmyID: attacker.ID, TargetRegion: target.ID, BattleStance: combat.BattleStanceBalanced},
			thirdAction:   InputAction{Kind: ActionLiftSiege, ArmyID: attacker.ID, TargetRegion: target.ID},
		}
		return
	}
	msg := fmt.Sprintf("%s tahkimli. Tahkimat seviyesi: %d | Kuşatma gücü: %d | Gedik kapasitesi: T%d/T%d. İstersen kuşatma kur, istersen doğrudan genel hücum dene.", target.NameTR, fortLevel, attacker.SiegeUnitScore(r.gs.UnitTypes), bestTier, fortLevel)
	r.confirmDialog = confirmDialogState{
		show:          true,
		title:         "Kuşatma Kararı",
		message:       msg,
		messageLines:  wrapTextLines(msg, FaceSmall, float64(confirmDialogW)-40),
		acceptLabel:   "Kuşatma Başlat",
		thirdLabel:    "Genel Hücum",
		declineLabel:  "İptal",
		pendingAction: InputAction{Kind: ActionStartSiege, ArmyID: attacker.ID, TargetRegion: target.ID},
		thirdAction:   InputAction{Kind: ActionAssaultSiege, ArmyID: attacker.ID, TargetRegion: target.ID, BattleStance: combat.BattleStanceBalanced},
	}
}

func (r *Renderer) selectedSiegePanelState() (*army.Army, *state.SiegeState, *world.Region, bool) {
	if r == nil || r.gs == nil || r.gs.Phase != state.PhasePlayerTurn || r.SelectedArmy == "" {
		return nil, nil, nil, false
	}
	if r.confirmDialog.show || r.warConfirm.show || r.battlePlan.show || r.showHistoricalEvent ||
		r.eventDetail != "" || r.showVictoryDetail || r.showEventCodex || r.showDiplomacy ||
		r.showTech || r.showTrade {
		return nil, nil, nil, false
	}
	attacker := r.gs.Armies[r.SelectedArmy]
	if attacker == nil || attacker.OwnerID != string(r.gs.PlayerFactionID) {
		return nil, nil, nil, false
	}
	siege := r.gs.SiegeByArmy(attacker.ID)
	if siege == nil {
		return nil, nil, nil, false
	}
	target := r.gs.Regions[siege.RegionID]
	if target == nil {
		return nil, nil, nil, false
	}
	return attacker, siege, target, true
}

func buildSelectedSiegePanel() gameui.Panel {
	x := (ScreenWidth - selectedSiegePanelW) / 2
	y := (ScreenHeight - selectedSiegePanelH) / 2
	return gameui.NewPanel(x, y, selectedSiegePanelW, selectedSiegePanelH)
}

func buildSelectedSiegeButtons() (gameui.Button, gameui.Button) {
	panel := buildSelectedSiegePanel()
	btnY := panel.Rect.Y + panel.Rect.H - selectedSiegeButtonH - 14
	gap := 16.0
	totalW := selectedSiegeButtonW*2 + gap
	startX := panel.Rect.X + (panel.Rect.W-totalW)/2
	assaultBtn := gameui.NewButton(startX, btnY, selectedSiegeButtonW, selectedSiegeButtonH, "Genel Hücum").WithIcon(gameui.IconSword)
	liftBtn := gameui.NewButton(startX+selectedSiegeButtonW+gap, btnY, selectedSiegeButtonW, selectedSiegeButtonH, "Kuşatmayı Kaldır").WithIcon(gameui.IconExit)
	return assaultBtn, liftBtn
}

func (r *Renderer) selectedSiegePanelHit(fx, fy float64) bool {
	if _, _, _, ok := r.selectedSiegePanelState(); !ok {
		return false
	}
	return buildSelectedSiegePanel().HitTest(fx, fy)
}

func (r *Renderer) selectedSiegePanelHovering(fx, fy float64) bool {
	if _, _, _, ok := r.selectedSiegePanelState(); !ok {
		return false
	}
	assaultBtn, liftBtn := buildSelectedSiegeButtons()
	return assaultBtn.HitTest(fx, fy) || liftBtn.HitTest(fx, fy)
}

func (r *Renderer) drawSelectedSiegePanel(screen *ebiten.Image) {
	attacker, siege, target, ok := r.selectedSiegePanelState()
	if !ok {
		return
	}
	panel := buildSelectedSiegePanel()
	gameui.DrawPanel(screen, panel, gameui.PanelStyle{
		BG:          color.RGBA{18, 12, 8, 234},
		Border:      color.RGBA{154, 112, 48, 255},
		BorderWidth: 2,
	})
	drawUILabel(screen, gameui.Rect{X: panel.Rect.X + 18, Y: panel.Rect.Y + 12}, "Kuşatma Emri", ColorYellow, gameui.TextLarge, gameui.TextAlignStart)
	bestTier := attacker.HighestSiegeTier(r.gs.UnitTypes)
	status := "Gedik yok"
	if level := siegeBreachLabelTR(siege.BreachLevel); level != "" {
		status = level
	}
	info := fmt.Sprintf("%s kuşatması sürüyor. Tahkimat: %d | İlerleme: %d | Durum: %s | Gedik: T%d/T%d", target.NameTR, siege.FortLevel, siege.BreachProgress, status, bestTier, siege.FortLevel)
	hint := "Başka komşu bölgeye hareket emri verirsen kuşatma otomatik kaldırılır."
	drawUIWrappedLabel(screen, gameui.Rect{X: panel.Rect.X + 18, Y: panel.Rect.Y + 42, W: panel.Rect.W - 36}, info, color.RGBA{228, 224, 214, 255}, gameui.TextSmall, 17, 2)
	drawUIWrappedLabel(screen, gameui.Rect{X: panel.Rect.X + 18, Y: panel.Rect.Y + 76, W: panel.Rect.W - 36}, hint, color.RGBA{170, 196, 152, 255}, gameui.TextSmall, 17, 2)
	assaultBtn, liftBtn := buildSelectedSiegeButtons()
	drawUIButtonWidget(screen, assaultBtn, solidButtonStyle(color.RGBA{70, 140, 70, 240}, color.RGBA{120, 180, 120, 255}, ColorWhite, 10))
	drawUIButtonWidget(screen, liftBtn, solidButtonStyle(color.RGBA{145, 95, 45, 235}, color.RGBA{190, 135, 75, 255}, ColorWhite, 10))
}

func battlePlanInstructionTR(context combat.BattleContext) string {
	switch combat.NormalizeBattleContext(context) {
	case combat.BattleContextNaval:
		return "Bir duruş seçin. Alt satırdaki tahminler mevcut filo gücü, modlar ve zar aralığına göre hesaplanır."
	case combat.BattleContextAmphibious:
		return "Bir duruş seçin. Alt satırdaki tahminler çıkarma gücü, kıyı savunması ve zar aralığına göre hesaplanır."
	default:
		return "Bir duruş seçin. Alt satırdaki tahminler mevcut güç, arazi ve zar aralığına göre hesaplanır."
	}
}

func (r *Renderer) openBattlePlan(attacker *army.Army, target *world.Region, defender *army.Army, actionKind ActionKind, battleContext combat.BattleContext) {
	if r == nil || attacker == nil || target == nil || defender == nil {
		return
	}
	previewAttacker := attacker
	if combat.NormalizeBattleContext(battleContext) == combat.BattleContextAmphibious {
		previewAttacker = &army.Army{
			OwnerID: attacker.OwnerID,
			Units:   attacker.EmbarkedUnits,
		}
	}
	atkMods := combat.TechModsFor(r.gs, attacker.OwnerID)
	defMods := combat.TechModsFor(r.gs, defender.OwnerID)
	state := battlePlanState{
		show:          true,
		actionKind:    actionKind,
		battleContext: combat.NormalizeBattleContext(battleContext),
		pendingArmy:   attacker.ID,
		pendingEnemy:  defender.ID,
		pendingDest:   target.ID,
		regionName:    target.NameTR,
		focus:         1,
	}
	for i, stance := range battlePlanStances {
		state.previews[i] = combat.PreviewBattleWithContextMods(previewAttacker, defender, target.Terrain, r.gs.UnitTypes, atkMods, defMods, state.battleContext, stance)
	}
	if factionInfo := r.gs.Factions[faction.FactionID(defender.OwnerID)]; factionInfo != nil {
		state.defenderFaction = factionInfo.NameTR
	} else {
		state.defenderFaction = defender.OwnerID
	}
	state.defenderName = "Savunan: " + state.defenderFaction
	if target.NameTR != "" {
		state.regionName = target.NameTR
	}
	r.battlePlan = state
}

func battlePlanLossText(expected, minLoss, maxLoss int) string {
	if minLoss == maxLoss {
		return itoa(expected)
	}
	return itoa(expected) + " (" + itoa(minLoss) + "-" + itoa(maxLoss) + ")"
}

func battlePlanHPText(expected, minLoss, maxLoss int) string {
	if minLoss == maxLoss {
		return "~" + itoa(expected) + " HP"
	}
	return "~" + itoa(expected) + " HP (" + itoa(minLoss) + "-" + itoa(maxLoss) + ")"
}

// --- Savaş ilan onay diyalogu ---

func (r *Renderer) drawWarConfirmDialog(screen *ebiten.Image) {
	modal := buildWarConfirmModal()
	gameui.DrawModal(screen, modal, standardModalStyle, nil, nil)

	// Mesaj
	msg := r.warConfirm.factionName + " ile savaş ilan edilsin mi?"
	tw := MeasureText(msg, FaceMed)
	DrawText(screen, msg, modal.Panel.Rect.X+(modal.Panel.Rect.W-tw)/2, modal.Panel.Rect.Y+18, FaceMed, color.RGBA{255, 220, 100, 255})

	acceptBtn, declineBtn := buildWarConfirmButtons()
	acceptBtn.Label = "Savaş İlan Et"
	declineBtn.Label = "Hayır"
	drawUIButtonWidget(screen, acceptBtn,
		solidButtonStyle(color.RGBA{160, 40, 40, 230}, color.RGBA{205, 90, 90, 255}, color.RGBA{255, 220, 220, 255}, 10))
	drawUIButtonWidget(screen, declineBtn,
		solidButtonStyle(color.RGBA{50, 50, 50, 230}, color.RGBA{120, 120, 120, 255}, color.RGBA{200, 200, 200, 255}, 10))
}

func (r *Renderer) handleWarConfirmInput() InputAction {
	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	acceptBtn, declineBtn := buildWarConfirmButtons()

	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if acceptBtn.HitTest(mx, my) {
			wc := r.warConfirm
			r.warConfirm = warConfirmState{}
			attacker := r.gs.Armies[wc.pendingArmy]
			target := r.gs.Regions[wc.pendingDest]
			if renderTargetRequiresSiegeDecision(r.gs, attacker, target) {
				if attacker != nil && attacker.HasSiegeUnits(r.gs.UnitTypes) {
					r.openSiegeDecision(attacker, target)
				} else {
					r.ShowCombatResult("Bu tahkimatı zorlamak için orduda en az bir kuşatma birimi olmalı.")
				}
				return InputAction{
					Kind:          ActionDeclareWar,
					TargetFaction: faction.FactionID(wc.factionID),
				}
			}
			if wc.opensBattlePlan {
				if attacker != nil {
					if target != nil {
						if defender := r.gs.Armies[wc.pendingEnemy]; defender != nil {
							r.openBattlePlan(attacker, target, defender, wc.battleAction, wc.battleContext)
						}
					}
				}
				return InputAction{
					Kind:          ActionDeclareWar,
					TargetFaction: faction.FactionID(wc.factionID),
				}
			}
			r.SelectedArmy = ""
			return InputAction{
				Kind:          ActionDeclareWarAndMove,
				ArmyID:        wc.pendingArmy,
				TargetRegion:  wc.pendingDest,
				TargetFaction: faction.FactionID(wc.factionID),
			}
		}
		if declineBtn.HitTest(mx, my) {
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
		attacker := r.gs.Armies[wc.pendingArmy]
		target := r.gs.Regions[wc.pendingDest]
		if renderTargetRequiresSiegeDecision(r.gs, attacker, target) {
			if attacker != nil && attacker.HasSiegeUnits(r.gs.UnitTypes) {
				r.openSiegeDecision(attacker, target)
			} else {
				r.ShowCombatResult("Bu tahkimatı zorlamak için orduda en az bir kuşatma birimi olmalı.")
			}
			return InputAction{
				Kind:          ActionDeclareWar,
				TargetFaction: faction.FactionID(wc.factionID),
			}
		}
		if wc.opensBattlePlan {
			if attacker != nil {
				if target != nil {
					if defender := r.gs.Armies[wc.pendingEnemy]; defender != nil {
						r.openBattlePlan(attacker, target, defender, wc.battleAction, wc.battleContext)
					}
				}
			}
			return InputAction{
				Kind:          ActionDeclareWar,
				TargetFaction: faction.FactionID(wc.factionID),
			}
		}
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

func (r *Renderer) drawBattlePlanDialog(screen *ebiten.Image) {
	modal := buildBattlePlanModal()
	gameui.DrawModal(screen, modal, standardModalStyle, nil, nil)

	title := combat.BattleContextLabelTR(r.battlePlan.battleContext) + " Planı"
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 24}, title, color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignStart)
	subtitle := r.battlePlan.regionName
	if subtitle == "" {
		subtitle = string(r.battlePlan.pendingDest)
	}
	if r.battlePlan.defenderFaction != "" {
		subtitle += " | " + r.battlePlan.defenderFaction
	}
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 52, W: modal.Panel.Rect.W - 48}, subtitle, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 72, W: modal.Panel.Rect.W - 48}, battlePlanInstructionTR(r.battlePlan.battleContext), color.RGBA{220, 220, 220, 255}, gameui.TextSmall, gameui.TextAlignStart)

	cardRects := battlePlanCardRects()
	buttons, cancelBtn := buildBattlePlanButtons()
	for i := range r.battlePlan.previews {
		preview := r.battlePlan.previews[i]
		card := cardRects[i]
		fill := color.RGBA{34, 28, 20, 236}
		border := color.RGBA{108, 86, 54, 255}
		if i == r.battlePlan.focus {
			fill = color.RGBA{54, 40, 24, 244}
			border = color.RGBA{220, 170, 82, 255}
		}
		vector.FillRect(screen, float32(card.X), float32(card.Y), float32(card.W), float32(card.H), fill, false)
		vector.StrokeRect(screen, float32(card.X), float32(card.Y), float32(card.W), float32(card.H), 1.5, border, false)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 14, W: card.W - 28}, preview.StanceLabelTR, color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignStart)
		drawUIWrappedLabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 42, W: card.W - 28}, preview.StanceSummaryTR, color.RGBA{212, 212, 212, 255}, gameui.TextSmall, 17, 2)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 88, W: card.W - 28}, "Zafer Şansı: %"+itoa(preview.WinChance), color.RGBA{178, 228, 150, 255}, gameui.TextMedium, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 114, W: card.W - 28}, "Muhtemel Sonuç: "+preview.LikelyOutcome, color.RGBA{226, 226, 226, 255}, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 138, W: card.W - 28}, "Güç: "+itoa(preview.AttackStrength)+" / "+itoa(preview.DefenseStrength), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 162, W: card.W - 28}, "Sizin HP: "+battlePlanHPText(preview.AttackerHPExpected, preview.AttackerHPDamageMin, preview.AttackerHPDamageMax), color.RGBA{255, 198, 148, 255}, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 184, W: card.W - 28}, "Sizin Birim: "+battlePlanLossText(preview.AttackerLossExpected, preview.AttackerLossMin, preview.AttackerLossMax), color.RGBA{232, 182, 132, 255}, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 208, W: card.W - 28}, "Düşman HP: "+battlePlanHPText(preview.DefenderHPExpected, preview.DefenderHPDamageMin, preview.DefenderHPDamageMax), color.RGBA{168, 220, 168, 255}, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: card.X + 14, Y: card.Y + 230, W: card.W - 28}, "Düşman Birim: "+battlePlanLossText(preview.DefenderLossExpected, preview.DefenderLossMin, preview.DefenderLossMax), color.RGBA{140, 206, 140, 255}, gameui.TextSmall, gameui.TextAlignStart)

		btn := buttons[i]
		btn.Label = preview.StanceLabelTR
		btnStyle := solidButtonStyle(color.RGBA{84, 68, 44, 240}, color.RGBA{146, 112, 62, 255}, ColorWhite, 10)
		if i == r.battlePlan.focus {
			btnStyle = solidButtonStyle(color.RGBA{140, 94, 38, 245}, color.RGBA{206, 150, 70, 255}, ColorWhite, 10)
		}
		drawUIButtonWidget(screen, btn, btnStyle)
	}

	cancelBtn.Label = "İptal"
	drawUIButtonWidget(screen, cancelBtn,
		solidButtonStyle(color.RGBA{72, 72, 72, 220}, color.RGBA{118, 118, 118, 255}, ColorWhite, 10))
}

func (r *Renderer) handleBattlePlanInput() InputAction {
	if !r.battlePlan.show {
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyLeft) || r.keyJustPressed(ebiten.KeyUp) {
		r.battlePlan.focus = (r.battlePlan.focus + len(battlePlanStances) - 1) % len(battlePlanStances)
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyRight) || r.keyJustPressed(ebiten.KeyDown) || r.keyJustPressed(ebiten.KeyTab) {
		r.battlePlan.focus = (r.battlePlan.focus + 1) % len(battlePlanStances)
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.Key1) {
		r.battlePlan.focus = 0
	}
	if r.keyJustPressed(ebiten.Key2) {
		r.battlePlan.focus = 1
	}
	if r.keyJustPressed(ebiten.Key3) {
		r.battlePlan.focus = 2
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.battlePlan = battlePlanState{}
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace) {
		bp := r.battlePlan
		r.battlePlan = battlePlanState{}
		r.SelectedArmy = ""
		return InputAction{
			Kind:         bp.actionKind,
			ArmyID:       bp.pendingArmy,
			TargetRegion: bp.pendingDest,
			BattleStance: battlePlanStances[bp.focus],
		}
	}

	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	buttons, cancelBtn := buildBattlePlanButtons()
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if cancelBtn.HitTest(mx, my) {
			r.battlePlan = battlePlanState{}
			return InputAction{}
		}
		for i, btn := range buttons {
			if !btn.HitTest(mx, my) {
				continue
			}
			bp := r.battlePlan
			r.battlePlan = battlePlanState{}
			r.SelectedArmy = ""
			return InputAction{
				Kind:         bp.actionKind,
				ArmyID:       bp.pendingArmy,
				TargetRegion: bp.pendingDest,
				BattleStance: battlePlanStances[i],
			}
		}
	}
	return InputAction{}
}

func (r *Renderer) playerDiplomacyOfferIndex() (int, bool) {
	if r.gs == nil || len(r.gs.DiplomaticOffers) == 0 {
		return 0, false
	}
	for i, offer := range r.gs.DiplomaticOffers {
		from := r.gs.Factions[offer.FromFactionID]
		to := r.gs.Factions[offer.ToFactionID]
		if from == nil || to == nil || from.IsEliminated || to.IsEliminated {
			continue
		}
		if offer.ToFactionID == r.gs.PlayerFactionID {
			return i, true
		}
	}
	return 0, false
}

func (r *Renderer) drawDiplomacyOfferDialog(screen *ebiten.Image, offerIdx int) {
	offer := r.gs.DiplomaticOffers[offerIdx]
	fromName := string(offer.FromFactionID)
	if f := r.gs.Factions[offer.FromFactionID]; f != nil && f.NameTR != "" {
		fromName = f.NameTR
	}
	actionLabel := "teklif"
	switch offer.Action {
	case "propose_peace":
		actionLabel = "barış"
	case "propose_alliance":
		actionLabel = "ittifak"
	case "propose_trade":
		actionLabel = "ticaret"
	}

	modal := buildDiplomacyOfferModal()
	gameui.DrawModal(screen, modal, standardModalStyle, nil, nil)

	title := "Anlaşma Teklifi"
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 20, Y: modal.Panel.Rect.Y + 30}, title, color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignStart)
	message := fromName + " devleti size " + actionLabel + " teklif etti."
	drawUIWrappedLabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 20, Y: modal.Panel.Rect.Y + 64, W: modal.Panel.Rect.W - 40}, message, color.RGBA{220, 220, 220, 255}, gameui.TextMedium, 20, 3)
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 20, Y: modal.Panel.Rect.Y + 124}, "Kabul etmek için Enter/Y, reddetmek için Esc/N kullanabilirsiniz.", ColorGray, gameui.TextSmall, gameui.TextAlignStart)

	acceptBtn, rejectBtn := buildDiplomacyOfferButtons()
	drawUIButtonWidget(screen, acceptBtn,
		solidButtonStyle(color.RGBA{70, 140, 70, 240}, color.RGBA{120, 180, 120, 255}, ColorWhite, 10))
	drawUIButtonWidget(screen, rejectBtn,
		solidButtonStyle(color.RGBA{140, 70, 70, 240}, color.RGBA{190, 110, 110, 255}, ColorWhite, 10))
}

func (r *Renderer) handleDiplomacyOfferInput(offerIdx int) InputAction {
	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	acceptBtn, rejectBtn := buildDiplomacyOfferButtons()
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if acceptBtn.HitTest(mx, my) {
			return InputAction{Kind: ActionRespondDiplomacyOffer, OfferIndex: offerIdx, OfferAccepted: true}
		}
		if rejectBtn.HitTest(mx, my) {
			return InputAction{Kind: ActionRespondDiplomacyOffer, OfferIndex: offerIdx, OfferAccepted: false}
		}
	}
	if r.keyJustPressed(ebiten.KeyY) || r.keyJustPressed(ebiten.KeyEnter) {
		return InputAction{Kind: ActionRespondDiplomacyOffer, OfferIndex: offerIdx, OfferAccepted: true}
	}
	if r.keyJustPressed(ebiten.KeyN) || r.keyJustPressed(ebiten.KeyEscape) {
		return InputAction{Kind: ActionRespondDiplomacyOffer, OfferIndex: offerIdx, OfferAccepted: false}
	}
	return InputAction{}
}

func (r *Renderer) handleHistoricalEventInput() InputAction {
	if len(r.historicalEventChoices) == 0 {
		if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyEnter) ||
			r.keyJustPressed(ebiten.KeySpace) || r.mouseJustPressed(ebiten.MouseButtonLeft) {
			r.showHistoricalEvent = false
			r.historicalEventPrompt = ""
			r.historicalEventChoices = r.historicalEventChoices[:0]
		}
		return InputAction{}
	}

	buttons := buildHistoricalEventChoiceButtons(len(r.historicalEventChoices))
	if r.keyJustPressed(ebiten.KeyLeft) || r.keyJustPressed(ebiten.KeyUp) {
		r.historicalEventFocus = (r.historicalEventFocus + len(r.historicalEventChoices) - 1) % len(r.historicalEventChoices)
	}
	if r.keyJustPressed(ebiten.KeyRight) || r.keyJustPressed(ebiten.KeyDown) || r.keyJustPressed(ebiten.KeyTab) {
		r.historicalEventFocus = (r.historicalEventFocus + 1) % len(r.historicalEventChoices)
	}
	if r.keyJustPressed(ebiten.Key1) {
		return InputAction{Kind: ActionChooseHistoricalEvent, ChoiceIndex: 0}
	}
	if len(r.historicalEventChoices) > 1 && r.keyJustPressed(ebiten.Key2) {
		return InputAction{Kind: ActionChooseHistoricalEvent, ChoiceIndex: 1}
	}
	if r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace) {
		return InputAction{Kind: ActionChooseHistoricalEvent, ChoiceIndex: r.historicalEventFocus}
	}

	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		mxi, myi := ebiten.CursorPosition()
		mx, my := float64(mxi), float64(myi)
		for i, btn := range buttons {
			if btn.HitTest(mx, my) {
				r.historicalEventFocus = i
				return InputAction{Kind: ActionChooseHistoricalEvent, ChoiceIndex: i}
			}
		}
	}
	return InputAction{}
}

func (r *Renderer) handleEventCodexInput() InputAction {
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.CloseEventCodex()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyLeft) {
		r.cycleEventCodexFilter(-1)
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyRight) || r.keyJustPressed(ebiten.KeyTab) {
		r.cycleEventCodexFilter(1)
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyUp) {
		r.cycleEventCodexFocus(-1)
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyDown) {
		r.cycleEventCodexFocus(1)
		return InputAction{}
	}
	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 && eventCodexListHit(mx, my) {
		if wheelY > 0 {
			r.scrollEventCodex(-1)
		} else if wheelY < 0 {
			r.scrollEventCodex(1)
		}
		return InputAction{}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if eventCodexCloseHit(mx, my) || !eventCodexPopupHit(mx, my) {
			r.CloseEventCodex()
			return InputAction{}
		}
		buttons := buildEventCodexFilterButtons()
		for i, btn := range buttons {
			if btn.HitTest(mx, my) {
				r.eventCodexFilter = EventCodexFilter(i)
				r.eventCodexFocus = 0
				r.eventCodexScroll = 0
				return InputAction{}
			}
		}
		if idx := eventCodexEntryHit(mx, my, len(r.currentEventCodexEntries()), r.eventCodexScroll); idx >= 0 {
			r.eventCodexFocus = idx
			r.ensureEventCodexFocusVisible()
			return InputAction{}
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
	modal := buildConfirmDialogModal()
	gameui.DrawModal(screen, modal, standardModalStyle, nil, nil)

	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 20, Y: modal.Panel.Rect.Y + 28}, r.confirmDialog.title, color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignStart)
	drawUIWrappedLabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 20, Y: modal.Panel.Rect.Y + 58, W: modal.Panel.Rect.W - 40}, r.confirmDialog.message, color.RGBA{220, 220, 220, 255}, gameui.TextSmall, 17, 3)
	r.drawConfirmDialogButtons(screen)
}

func (r *Renderer) drawConfirmDialogButtons(screen *ebiten.Image) {
	acceptBtn, thirdBtn, declineBtn, hasThird := buildConfirmDialogButtons(r.confirmDialog)
	acceptBtn = decorateConfirmDialogButton(acceptBtn, r.confirmDialog.acceptLabel, "accept")
	drawUIButtonWidget(screen, acceptBtn,
		solidButtonStyle(color.RGBA{70, 140, 70, 240}, color.RGBA{120, 180, 120, 255}, ColorWhite, 10))
	if hasThird {
		thirdBtn = decorateConfirmDialogButton(thirdBtn, r.confirmDialog.thirdLabel, "third")
		declineBtn = decorateConfirmDialogButton(declineBtn, r.confirmDialog.declineLabel, "decline")
		drawUIButtonWidget(screen, thirdBtn,
			solidButtonStyle(color.RGBA{145, 95, 45, 235}, color.RGBA{190, 135, 75, 255}, ColorWhite, 10))
		drawUIButtonWidget(screen, declineBtn,
			solidButtonStyle(color.RGBA{70, 70, 70, 220}, color.RGBA{120, 120, 120, 255}, ColorWhite, 10))
		return
	}
	declineBtn = decorateConfirmDialogButton(declineBtn, r.confirmDialog.declineLabel, "decline")
	drawUIButtonWidget(screen, declineBtn,
		solidButtonStyle(color.RGBA{70, 70, 70, 220}, color.RGBA{120, 120, 120, 255}, ColorWhite, 10))
}

func decorateConfirmDialogButton(btn gameui.Button, label string, role string) gameui.Button {
	btn.Label = label
	switch role {
	case "accept":
		if label == "Kaydet" {
			return btn.WithIcon(gameui.IconSave)
		}
		return btn.WithIcon(gameui.IconCheck)
	case "third":
		if label == "Kaydetmeden Cik" {
			return btn.WithIcon(gameui.IconExit)
		}
		return btn.WithIcon(gameui.IconSave)
	default:
		return btn.WithIcon(gameui.IconClose)
	}
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
	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	acceptBtn, thirdBtn, declineBtn, hasThird := buildConfirmDialogButtons(r.confirmDialog)

	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if acceptBtn.HitTest(mx, my) {
			action := r.confirmDialog.pendingAction
			r.confirmDialog = confirmDialogState{}
			return action
		}
		if hasThird && thirdBtn.HitTest(mx, my) {
			action := r.confirmDialog.thirdAction
			r.confirmDialog = confirmDialogState{}
			return action
		}
		if declineBtn.HitTest(mx, my) {
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
