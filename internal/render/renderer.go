package render

import (
	"image/color"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/audio"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
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
	confirmDialogW             = float32(460)
	confirmDialogH             = float32(166)
	confirmDialogBtnW          = float32(120)
	confirmDialogBtnH          = float32(36)
	selectedSiegePanelW        = 520.0
	selectedSiegePanelH        = 298.0
	selectedSiegeButtonW       = 224.0
	selectedSiegeButtonH       = 38.0
	initialCameraZoomFactor    = 2.50
	maxCameraZoomScale         = 10
	activeEventIconSize        = float32(22)
	activeEventIconSpacingY    = float32(24)
	activeEventIconLiftY       = float32(48)
	settlementMarkerSpriteSize = float32(26)
	capitalLabelIconSmallSize  = float32(18)
	capitalLabelIconMediumSize = float32(20)
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
	factionPanelScroll       float64
	selectedSettlementRegion world.RegionID
	selectedSettlementIndex  int
	devNeighborListExpanded  bool
	regionPanelScroll        float64
	showRecruitPanel         bool
	recruitUnitID            string
	recruitQty               int

	// Senaryo seçim ekranı
	scenarioCursor int

	// Fraksiyon seçim ekranı
	factionCursor int

	// Diplomasi paneli
	showDiplomacy                   bool
	diplomacyFocus                  int
	diplomacyScroll                 int
	diplomacyActionFocus            int
	diplomacyTargetFaction          faction.FactionID
	diplomacyOfferHistoryBrowse     faction.FactionID
	diplomacyHistoryVisible         bool
	diplomacyHistoryDirectionFilter diplomacyHistoryDirectionFilter
	diplomacyHistoryActionFilter    ActionKind

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
	combatLog       string
	combatLogTimer  int
	aiTurnFactionID faction.FactionID
	aiTurnActor     string
	aiTurnInitial   string
	aiTurnDetail    string

	// Tarihsel olay tam ekran bildirimi
	historicalEventTitle   string
	historicalEventDesc    string
	historicalEventPrompt  string
	historicalEventChoices []HistoricalEventChoice
	historicalEventFocus   int
	showHistoricalEvent    bool
	battleReport           battleReportState
	queuedBattleReport     battleReportState
	warSummary             warSummaryState
	showCommanderPanel     bool
	commanderPanelArmy     army.ArmyID
	commanderPanelFocus    int

	// İlk frame kamera başlatma
	firstDraw bool

	// Input state (just-pressed takibi)
	prevKeys  map[ebiten.Key]bool
	prevMouse map[ebiten.MouseButton]bool

	// Genel onay diyaloğu
	warConfirm          warConfirmState
	battlePlan          battlePlanState
	confirmDialog       confirmDialogState
	queuedConfirmDialog confirmDialogState
	offerCursor         int

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
	declineAction InputAction
	declineActs   bool
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
	preview         diplomacy.WarDeclarationPreview
	selectedAllies  map[faction.FactionID]bool
	scroll          int
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

func (r *Renderer) canJoinActiveSiege(attacker *army.Army, regionID world.RegionID) bool {
	return r != nil && r.gs != nil && r.gs.CanJoinActiveSiege(attacker, regionID)
}

func (r *Renderer) canEnterActiveSiegedRegion(attacker *army.Army, regionID world.RegionID) bool {
	return r != nil && r.gs != nil && r.gs.CanEnterActiveSiegedRegion(attacker, regionID)
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
	attackerSummary string
	defenderSummary string
	focus           int
	previews        [3]combat.Preview
}

func (r *Renderer) openWarConfirm(targetID faction.FactionID, targetName string, pendingArmy army.ArmyID, pendingDest world.RegionID, pendingEnemy army.ArmyID, opensBattlePlan bool, battleAction ActionKind, battleContext combat.BattleContext) {
	if r == nil || r.gs == nil || targetID == "" {
		return
	}
	r.warConfirm = warConfirmState{
		show:            true,
		factionName:     targetName,
		factionID:       string(targetID),
		pendingArmy:     pendingArmy,
		pendingDest:     pendingDest,
		pendingEnemy:    pendingEnemy,
		opensBattlePlan: opensBattlePlan,
		battleAction:    battleAction,
		battleContext:   battleContext,
		preview:         diplomacy.BuildWarDeclarationPreview(r.gs, r.gs.PlayerFactionID, targetID),
		selectedAllies:  make(map[faction.FactionID]bool),
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
	// Haritanın üst kenarını ekranın üstüne hizala.
	r.camY = ScreenHeight / (2 * r.camScale * mapPitchY)
	if focusX, focusY, ok := r.initialCameraFocusPoint(); ok {
		r.camX = focusX
		r.camY = focusY
	}
	r.camX, r.camY = clampCameraCenter(r.camX, r.camY, r.camScale)
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

func (r *Renderer) initialCameraFocusPoint() (float64, float64, bool) {
	if r == nil || r.gs == nil || r.gs.PlayerFactionID == "" {
		return 0, 0, false
	}
	region, settlement, _, ok := r.gs.FactionCapital(r.gs.PlayerFactionID)
	if !ok || region == nil {
		return 0, 0, false
	}
	if settlement != nil {
		return wcX(settlement.X), wcY(settlement.Y), true
	}
	return wcX(region.WorldX), wcY(region.WorldY), true
}

func clampCameraCenter(camX, camY, scale float64) (float64, float64) {
	if scale <= 0 {
		return camX, camY
	}
	halfW := ScreenWidth / (2 * scale)
	halfH := ScreenHeight / (2 * scale * mapPitchY)
	minX, maxX := halfW, float64(WorldW)-halfW
	minY, maxY := halfH, float64(WorldH)-halfH
	if minX > maxX {
		camX = float64(WorldW) / 2
	} else {
		camX = math.Max(minX, math.Min(maxX, camX))
	}
	if minY > maxY {
		camY = float64(WorldH) / 2
	} else {
		camY = math.Max(minY, math.Min(maxY, camY))
	}
	return camX, camY
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

func (r *Renderer) SetAITurnStatus(fid faction.FactionID, actor, detail string) {
	r.aiTurnFactionID = fid
	r.aiTurnActor = actor
	r.aiTurnInitial = ""
	if _, size := utf8.DecodeRuneInString(actor); size > 0 {
		r.aiTurnInitial = actor[:size]
	}
	r.aiTurnDetail = detail
}

func (r *Renderer) ClearAITurnStatus() {
	r.aiTurnFactionID = ""
	r.aiTurnActor = ""
	r.aiTurnInitial = ""
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
		resetBuildingSpriteCache()
		miniMapLoaded = false
		armySpritesLoaded = false
		unitSprites = nil
		legacyUnitSprites = nil
		legacyArmySheet = nil
		settlementImageCache = map[string]*ebiten.Image{}
		settlementImageLoaded = map[string]bool{}
		resetFactionFlagCache()
		resetCommanderPortraitCache()
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
	r.closeFactionPanel()
	r.CloseCommanderPanel()
	r.clearSelectedSettlement()
	r.ClearAITurnStatus()
	r.confirmDialog = confirmDialogState{}
	r.queuedConfirmDialog = confirmDialogState{}
	r.battleReport = battleReportState{}
	r.queuedBattleReport = battleReportState{}
	r.warSummary = warSummaryState{}
	r.eventLogScroll = 0
	r.editRegionPaintOverrides = make(map[int]world.RegionID)
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

// PrepareForTurnAdvance oyuncu turundan çıkarken açık panelleri kapatır
// ve haritayı nötr görünüme döndürür.
func (r *Renderer) PrepareForTurnAdvance() {
	if r == nil {
		return
	}
	r.SelectedRegion = ""
	r.SelectedArmy = ""
	r.closeFactionPanel()
	r.CloseCommanderPanel()
	r.devNeighborListExpanded = false
	r.clearSelectedSettlement()
	r.showRecruitPanel = false
	r.resetRecruitSelection()
	r.showDiplomacy = false
	r.diplomacyFocus = 0
	r.diplomacyScroll = 0
	r.diplomacyActionFocus = 0
	r.diplomacyTargetFaction = ""
	r.diplomacyOfferHistoryBrowse = ""
	r.diplomacyHistoryVisible = false
	r.showTech = false
	r.techCursor = 0
	r.techDragging = false
	r.showTrade = false
	r.tradeScroll = 0
	r.tradeFactionFocus = 0
	r.tradeGoodFocus = 0
	r.tradeHoverIdx = -1
	r.tradeCenterIdx = -1
	r.mapMode = MapModeNormal
	r.CloseEventCodex()
	r.eventDetail = ""
	r.showVictoryDetail = false
	r.victoryDetailScroll = 0
	r.warSummary = warSummaryState{}
	r.queuedBattleReport = battleReportState{}
	r.combatLog = ""
	r.combatLogTimer = 0
}

func (r *Renderer) worldInputLockedByPhase() bool {
	if r == nil || r.gs == nil {
		return false
	}
	return r.gs.Phase == state.PhaseAITurn || r.gs.Phase == state.PhaseTurnResolution
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
		recruitReason := ""
		if !recruitEnabled {
			if r.isSettlementPanelOpen() && RecruitPanelButtonEnabled(r.gs, r.SelectedRegion) {
				recruitReason = "Yerleşim Paneli Açık"
			} else {
				recruitReason = recruitPanelDisabledReason(r.gs, r.SelectedRegion)
			}
		}
		DrawBottomPanel(screen, r.gs, r.showRecruitPanel, recruitEnabled, recruitReason, r.showDiplomacy, r.showTech, r.mapMode)
		DrawRegionPanelExpandedScrolled(screen, r.gs, r.SelectedRegion, r.devNeighborListExpanded, r.regionPanelScroll)
		if region, settlement, ok := r.selectedSettlement(); ok && region.ID == r.SelectedRegion {
			DrawSettlementPanel(screen, r.gs, region, settlement)
		}
		DrawFactionDetailPanel(screen, r.gs, r.selectedFactionPanel, r.factionPanelScroll)
		if r.mapMode != MapModeTrade && r.showRecruitPanel {
			DrawRecruitPanel(screen, r.gs, r.SelectedRegion, r.recruitUnitID, r.recruitQty)
		}
		DrawArmyDetailPanel(screen, r.gs, r.SelectedArmy)
		DrawMinimap(screen, r.gs, r.camX, r.camY, r.camScale)
		r.drawSelectedSiegePanel(screen)
	}
	if r.gs.Phase != state.PhaseEditMode {
		DrawEventLog(screen, r.eventLog, r.eventLogCollapsed, r.eventLogScroll, r.HasEventCodex())
		DrawHoverTooltip(screen, r.gs, r.SelectedRegion, r.SelectedArmy, r.showRecruitPanel)
	} else {
		r.drawEditModeHud(screen)
		r.drawEditInspector(screen)
		r.drawEditFactionForm(screen)
	}

	// 7. Diplomasi paneli (üst katman)
	if r.showDiplomacy {
		DrawDiplomacyPanel(screen, r.gs, r.diplomacyFocus, r.diplomacyScroll, r.diplomacyActionFocus, r.diplomacyTargetFaction, r.diplomacyOfferHistoryBrowse, r.diplomacyHistoryVisible, r.diplomacyHistoryDirectionFilter, r.diplomacyHistoryActionFilter)
	}

	// 8. Teknoloji paneli (üst katman)
	if r.showTech {
		r.DrawTechPanel(screen)
	}

	if r.gs.Phase == state.PhaseAITurn && r.aiTurnActor != "" {
		r.drawAITurnOverlay(screen)
	}

	// 9. Onay diyalogu (diğer popupların altında kalmaması için üst katman)
	if r.confirmDialog.show {
		r.drawConfirmDialog(screen)
	} else if r.warConfirm.show {
		r.drawWarConfirmDialog(screen)
	} else if r.warSummary.show {
		drawWarSummaryDialog(screen, r.warSummary)
	} else if r.battlePlan.show {
		r.drawBattlePlanDialog(screen)
	} else if offerIdx, ok := r.playerDiplomacyOfferIndex(); ok {
		r.drawDiplomacyOfferDialog(screen, offerIdx)
	}

	if r.battleReport.show {
		drawBattleReportPopup(screen, r.battleReport.data)
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

	// 14. Bildirim mesajı (panellerin üstünde görünmeli)
	if r.combatLogTimer > 0 {
		alpha := uint8(255)
		if r.combatLogTimer < 60 {
			alpha = uint8(r.combatLogTimer * 255 / 60)
		}
		popupH := infoPopupHeight(r.combatLog)
		popupY := float32(ScreenHeight)*0.26 - popupH/2
		if r.gs.Phase == state.PhaseAITurn && r.aiTurnActor != "" {
			const aiOverlayGap = float32(40)
			const aiOverlayH = float32(180)
			_, turnHudY, _, turnHudH := turnTechHudRect()
			popupY = turnHudY + turnHudH + aiOverlayGap + aiOverlayH + 16
		}
		drawInfoPopupAt(screen, r.combatLog, alpha, popupY)
		r.combatLogTimer--
	}

	// 15. Tarihsel olay popup'ı gerçek üst modal olmalı.
	if r.showHistoricalEvent {
		drawHistoricalEventPopup(screen, r.historicalEventTitle, r.historicalEventDesc, r.historicalEventPrompt, r.historicalEventChoices, r.historicalEventFocus)
	}
	if r.showCommanderPanel {
		r.DrawCommanderPanel(screen)
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
	if r.confirmDialog.show || r.warConfirm.show || r.warSummary.show || r.battlePlan.show || r.battleReport.show || r.eventDetail != "" {
		return false
	}
	if _, ok := r.playerDiplomacyOfferIndex(); ok {
		return false
	}
	return true
}

func (r *Renderer) BattleReportVisible() bool {
	return r != nil && r.battleReport.show
}

func (r *Renderer) WarSummaryVisible() bool {
	return r != nil && r.warSummary.show
}

func (r *Renderer) drawAITurnOverlay(screen *ebiten.Image) {
	if r.aiTurnActor == "" {
		return
	}
	const panelW, panelH = float32(620), float32(180)
	const flagSize = float64(128)
	x := float32(ScreenWidth)/2 - panelW/2
	_, turnHudY, _, turnHudH := turnTechHudRect()
	y := turnHudY + turnHudH + 40
	drawRoundedRect(screen, x, y, panelW, panelH, 8, color.RGBA{16, 14, 10, 228})
	drawPanelBorder(screen, x, y, panelW, panelH)
	vector.FillRect(screen, x, y, panelW, 3, color.RGBA{205, 168, 72, 255}, false)
	DrawText(screen, "HAMLELER", float64(x)+16, float64(y)+10, FaceSmall, ColorGray)

	flagBG := color.RGBA{70, 58, 32, 255}
	if r.gs != nil {
		if f := r.gs.Factions[r.aiTurnFactionID]; f != nil {
			flagBG = color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		}
	}
	flagX := float64(x) + 14
	flagY := float64(y) + 36
	drawFactionFlagBadge(screen, r.aiTurnFactionID, r.aiTurnInitial, flagX, flagY, flagSize, flagBG, nil)

	contentX := float64(x) + 160
	DrawText(screen, r.aiTurnActor, contentX, float64(y)+48, FaceMed, ColorGold)
	if r.aiTurnDetail != "" {
		drawUIWrappedLabel(screen, gameui.Rect{X: contentX, Y: float64(y) + 78, W: float64(panelW) - 176}, r.aiTurnDetail, color.RGBA{230, 222, 204, 255}, gameui.TextSmall, 16, 2)
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
			if diplomacy.SameRealm(gs, faction.FactionID(a.OwnerID), faction.FactionID(target.OwnerID)) {
				return true
			}
			key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(target.OwnerID))
			rel, ok := gs.Relations[key]
			return ok && (rel.Stance == faction.StanceWar || rel.Stance == faction.StanceAllied)
		}
		return target.CanNavalEnter()
	}
	if target.CanNavalEnter() {
		return armyCanEmbark(gs, a) && findFriendlyEmbarkFleetFromRegion(gs, a.OwnerID, a.RegionID, target.ID, len(a.Units)) != nil
	}
	return target.CanLandEnter()
}

func armyRegionIsFriendly(gs *state.GameState, attacker *army.Army, target *world.Region) bool {
	if gs == nil || attacker == nil || target == nil || target.OwnerID == "" || target.OwnerID == attacker.OwnerID {
		return false
	}
	attackerFID := faction.FactionID(attacker.OwnerID)
	targetFID := faction.FactionID(target.OwnerID)
	if diplomacy.SameRealm(gs, attackerFID, targetFID) {
		return true
	}
	rel := diplomacy.Relation(gs, attackerFID, targetFID)
	return rel != nil && rel.Stance == faction.StanceAllied
}

func shouldPromptWarConfirmForMove(gs *state.GameState, attacker *army.Army, target *world.Region) bool {
	if gs == nil || attacker == nil || target == nil || target.OwnerID == "" || target.OwnerID == attacker.OwnerID {
		return false
	}
	if armyRegionIsFriendly(gs, attacker, target) {
		return false
	}
	rel := diplomacy.Relation(gs, faction.FactionID(attacker.OwnerID), faction.FactionID(target.OwnerID))
	return rel == nil || rel.Stance != faction.StanceWar
}

func navalShowsFriendlyDisembark(gs *state.GameState, fleet *army.Army, target *world.Region) bool {
	if gs == nil || fleet == nil || target == nil || !fleet.IsNaval || len(fleet.EmbarkedUnits) == 0 || !target.CanLandEnter() {
		return false
	}
	if target.OwnerID == "" || target.OwnerID == fleet.OwnerID {
		return true
	}
	return diplomacy.SameRealm(gs, faction.FactionID(fleet.OwnerID), faction.FactionID(target.OwnerID))
}

func navalCanDockAtRegion(gs *state.GameState, fleet *army.Army, target *world.Region) bool {
	if gs == nil || fleet == nil || target == nil || !fleet.IsNaval || target.IsSea || target.OwnerID == "" {
		return false
	}
	if fleet.OwnerID != target.OwnerID {
		if diplomacy.SameRealm(gs, faction.FactionID(fleet.OwnerID), faction.FactionID(target.OwnerID)) {
			return target.HasPortBuilding()
		}
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
	Region      *world.Region
	Index       int
	Text        string
	TextX       float64
	X, Y        float64
	W, H        float64
	SX, SY      float64
	DrawLabel   bool
	CapitalIcon bool
	Priority    int
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
		sort.Slice(aids, func(i, j int) bool {
			ai := r.gs.Armies[aids[i]]
			aj := r.gs.Armies[aids[j]]
			aiSieging := ai != nil && r.gs.SiegeByArmy(ai.ID) != nil
			ajSieging := aj != nil && r.gs.SiegeByArmy(aj.ID) != nil
			// Kuşatma ordusu solda, ayrılan ve hareket edebilen parça sağda
			// kalsın; armyHitAt sağdan taradığı için seçim deterministik olur.
			if aiSieging != ajSieging {
				return aiSieging
			}
			return aids[i] < aids[j]
		})

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
				if ax, ay, ok := r.siegeLandArmyAnchor(region); ok {
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

func (r *Renderer) siegeLandArmyAnchor(region *world.Region) (int, int, bool) {
	if region == nil || region.IsSea {
		return 0, 0, false
	}
	// Kuşatma orduları için kale (fortress) tipi yerleşimi öncele
	for i, settlement := range region.Settlements {
		if settlement.Type == world.SettlementFortress {
			if ax, ay, ok := r.worldMap.SettlementAnchor(region.ID, i); ok {
				return ax, ay, true
			}
		}
	}
	// Kale bulunamazsa standart kara ordusu konumlandırmasına dön
	return r.landArmyAnchor(region)
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
		r.drawArmyIcon(screen, a.ID, a.OwnerID, pos.X, pos.Y, fc, unitCount, isSelected, a.IsNaval)
		if embarkableFleetForSelectedArmy(r.gs, selectedArmy, a) {
			vector.StrokeCircle(screen, pos.X, pos.Y, 17, 3, color.RGBA{120, 230, 240, 220}, true)
			DrawTextCentered(screen, "BIN", float64(pos.X), float64(pos.Y)+15, FaceSmall, color.RGBA{210, 248, 255, 230})
		}
	}
}

func armyIconBorderColor(gs *state.GameState, ownerID string, selected bool) color.RGBA {
	if selected {
		return color.RGBA{255, 215, 0, 255}
	}
	if gs == nil || ownerID == "" {
		return color.RGBA{200, 200, 200, 220}
	}

	owner := faction.FactionID(ownerID)
	if gs.PlayerFactionID != "" && diplomacy.SameRealm(gs, gs.PlayerFactionID, owner) {
		return borderColorPlayerRealm
	}
	if rel := diplomacy.Relation(gs, gs.PlayerFactionID, owner); rel != nil {
		switch rel.Stance {
		case faction.StanceAllied:
			return borderColorAlly
		case faction.StanceWar:
			return borderColorEnemy
		}
	}
	return color.RGBA{200, 200, 200, 220}
}

// drawArmyIcon tek bir ordu ikonunu çizer.
// Kara ordusu → kare, deniz donanması → daire.
func (r *Renderer) drawArmyIcon(screen *ebiten.Image, aid army.ArmyID, ownerID string, cx, cy float32, col color.RGBA, unitCount int, selected bool, isNaval bool) {
	borderCol := armyIconBorderColor(r.gs, ownerID, selected)

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
		r.drawSettlementMarkerSprite(screen, armySiegeBadgeImage(), badgeX, badgeY, badgeSize-2)
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
		labelColor := labelCol
		shadowColor := shadowCol
		if selectedOK && item.Region != nil && item.Region.ID == selectedRID && item.Index == selectedIdx {
			labelColor = ColorGold
			shadowColor = color.RGBA{34, 22, 8, 210}
		}
		if !item.DrawLabel && !forceLabel {
			isPrimary := settlement.IsCapital || item.Index == 0
			r.drawSettlementMarker(screen, item.Region, settlement, float32(item.SX), float32(item.SY), isPrimary)
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
			if item.CapitalIcon {
				r.drawCapitalLabelIcon(screen, float32(item.X), float32(item.Y), variant)
			}
			outlined := gameui.NewOutlinedLabel(gameui.Rect{X: item.TextX, Y: item.Y}, item.Text, labelColor, shadowColor, variant, gameui.TextAlignStart)
			outlined.Offsets = [][2]float64{{1, 1}}
			outlined.Draw(screen, renderText)
			r.labelRectBuf = append(r.labelRectBuf, rect)
		}

		isPrimary := settlement.IsCapital || item.Index == 0
		r.drawSettlementMarker(screen, item.Region, settlement, float32(item.SX), float32(item.SY), isPrimary)
		r.drawSettlementSelectionOverlay(screen, settlement, item.Region, float32(item.SX), float32(item.SY))
	}
}

func (r *Renderer) appendSettlementDraws(region *world.Region) {
	if len(region.Settlements) == 0 {
		return
	}

	for i, settlement := range region.Settlements {
		isPrimary := settlement.IsCapital || i == 0
		isFactionCapital := r.isCapitalSettlement(region, settlement)

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
		drawLabel := r.shouldDrawSettlementLabel(settlement, isPrimary, isFactionCapital)
		r.appendSettlementDraw(region, i, name, sx, sy, drawLabel, settlementLabelPriority(settlement, isPrimary, isFactionCapital), isFactionCapital)
	}
}

func (r *Renderer) appendSettlementDraw(region *world.Region, index int, text string, sx, sy float64, drawLabel bool, priority int, capitalIcon bool) {
	if sx < -50 || sx > ScreenWidth+50 || sy < -20 || sy > ScreenHeight+20 {
		return
	}

	face := FaceSmall
	if r.camScale >= 1.0 {
		face = FaceMed
	}

	isMediumFace := face == FaceMed
	iconAdvance := 0.0
	if capitalIcon {
		iconAdvance = capitalLabelIconAdvance(isMediumFace)
	}
	textW := MeasureText(text, face)
	totalW := textW + iconAdvance
	lx := sx - totalW/2
	h := float64(16)
	if face == FaceMed {
		h = 20
	}
	r.regionLabelBuf = append(r.regionLabelBuf, settlementDraw{
		Region: region,
		Index:  index,
		Text:   text,
		TextX:  lx + iconAdvance,
		X:      lx,
		// Etiket noktaların altına çizilir; okunabilirlik artar.
		Y:           sy + 16,
		W:           totalW,
		H:           h,
		SX:          sx,
		SY:          sy,
		DrawLabel:   drawLabel,
		CapitalIcon: capitalIcon,
		Priority:    priority,
	})
}

func settlementLabelPriority(settlement world.Settlement, isPrimary bool, isFactionCapital bool) int {
	if isFactionCapital {
		return 110
	}
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

func (r *Renderer) shouldDrawSettlementLabel(settlement world.Settlement, isPrimary bool, isFactionCapital bool) bool {
	// Zoom düşükken sadece başkent/şehir etiketleri.
	if r.camScale < 0.8 {
		return isFactionCapital || settlement.IsCapital || settlement.Type == world.SettlementCity
	}
	// Orta zoomda liman ve kaleler de açılır.
	if r.camScale < 1.05 {
		return isFactionCapital || settlement.IsCapital || settlement.Type == world.SettlementCity ||
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

func (r *Renderer) drawSettlementMarker(screen *ebiten.Image, region *world.Region, settlement world.Settlement, sx, sy float32, isPrimary bool) {
	if region == nil {
		return
	}
	drawn := false
	if img := r.settlementMarkerSprite(region, settlement, isPrimary); img != nil {
		if r.drawSettlementMarkerSprite(screen, img, sx, sy, settlementMarkerSpriteSize) {
			drawn = true
		}
	}
	if !drawn {
		switch settlement.Type {
		case world.SettlementFortress:
			if !r.drawFortressMarkerSprite(screen, sx, sy) {
				r.drawFortressMarker(screen, region, sx, sy)
			}
		case world.SettlementPort:
			if !r.drawPortMarkerSprite(screen, sx, sy) {
				r.drawPortMarker(screen, region, sx, sy)
			}
		default:
			r.drawCityDot(screen, region, sx, sy)
		}
	}
	if isPrimary {
		r.drawVassalSettlementBadge(screen, region, sx, sy)
	}
}

func (r *Renderer) settlementMarkerSprite(region *world.Region, settlement world.Settlement, isPrimary bool) *ebiten.Image {
	if r == nil || region == nil {
		return nil
	}
	if isPrimary && r.isSettlementUnderSiege(region) {
		return settlementMarkerSiegeImage()
	}
	switch settlement.Type {
	case world.SettlementFortress:
		return settlementMarkerCastleImage()
	case world.SettlementPort:
		return settlementMarkerHarbourImage()
	case world.SettlementCity:
		return settlementMarkerCityImage()
	case world.SettlementTown:
		return settlementMarkerDistrictImage()
	default:
		return nil
	}
}

func armySiegeBadgeImage() *ebiten.Image {
	return settlementMarkerSwordImage()
}

func (r *Renderer) isSettlementUnderSiege(region *world.Region) bool {
	if r == nil || r.gs == nil || region == nil {
		return false
	}
	siege := r.gs.SiegeAt(region.ID)
	return siege != nil
}

func (r *Renderer) drawFortressMarkerSprite(screen *ebiten.Image, sx, sy float32) bool {
	return r.drawSettlementMarkerSprite(screen, settlementMarkerCastleImage(), sx, sy, settlementMarkerSpriteSize)
}

func (r *Renderer) drawPortMarkerSprite(screen *ebiten.Image, sx, sy float32) bool {
	return r.drawSettlementMarkerSprite(screen, settlementMarkerHarbourImage(), sx, sy, settlementMarkerSpriteSize)
}

func (r *Renderer) isCapitalSettlement(region *world.Region, settlement world.Settlement) bool {
	if r == nil || r.gs == nil || region == nil || region.OwnerID == "" || settlement.ID == "" {
		return false
	}
	return r.gs.IsFactionCapitalSettlement(faction.FactionID(region.OwnerID), settlement.ID)
}

func (r *Renderer) drawCapitalLabelIcon(screen *ebiten.Image, x, y float32, variant gameui.TextVariant) {
	size := capitalLabelIconSmallSize
	if variant == gameui.TextMedium {
		size = capitalLabelIconMediumSize
	}
	img := settlementMarkerStarImage()
	if img != nil {
		r.drawSettlementLabelSprite(screen, img, x, y-2, size)
		return
	}
	vector.FillCircle(screen, x+size/2, y+size/2+2, size/2-1, color.RGBA{255, 220, 110, 245}, true)
}

func (r *Renderer) drawVassalSettlementBadge(screen *ebiten.Image, region *world.Region, sx, sy float32) {
	if screen == nil || !r.isVassalRegion(region) {
		return
	}
	col := r.vassalOverlordColor(region)
	cx := sx + 7
	cy := sy - 6
	vector.FillCircle(screen, cx, cy, 4.1, color.RGBA{214, 185, 88, 240}, true)
	vector.FillCircle(screen, cx, cy, 2.5, col, true)
	vector.StrokeCircle(screen, cx, cy, 4.1, 1.1, color.RGBA{38, 28, 14, 225}, true)
}

func (r *Renderer) isVassalRegion(region *world.Region) bool {
	if r == nil || r.gs == nil || region == nil || region.OwnerID == "" {
		return false
	}
	return diplomacy.DirectOverlord(r.gs, faction.FactionID(region.OwnerID)) != ""
}

func (r *Renderer) vassalOverlordColor(region *world.Region) color.RGBA {
	if r == nil || r.gs == nil || region == nil || region.OwnerID == "" {
		return color.RGBA{130, 130, 130, 235}
	}
	overlord := diplomacy.DirectOverlord(r.gs, faction.FactionID(region.OwnerID))
	if overlord == "" {
		return color.RGBA{130, 130, 130, 235}
	}
	return factionColor(r.gs, string(overlord))
}

func (r *Renderer) drawSettlementMarkerSprite(screen *ebiten.Image, img *ebiten.Image, sx, sy, size float32) bool {
	if screen == nil || img == nil || size <= 0 {
		return false
	}
	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		return false
	}

	// Koyu renk simgelerin harita üzerinde belirgin olması için beyaz daire arka plan
	circleRadius := float32(size/2) + 1
	vector.DrawFilledCircle(screen, sx, sy, circleRadius, color.RGBA{255, 255, 255, 255}, true)

	op := &ebiten.DrawImageOptions{}
	scaleX := float64(size) / float64(bounds.Dx())
	scaleY := float64(size) / float64(bounds.Dy())
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(float64(sx-size/2), float64(sy-size/2))
	screen.DrawImage(img, op)
	return true
}

func (r *Renderer) drawSettlementLabelSprite(screen *ebiten.Image, img *ebiten.Image, x, y, size float32) bool {
	if screen == nil || img == nil || size <= 0 {
		return false
	}
	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		return false
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(size)/float64(bounds.Dx()), float64(size)/float64(bounds.Dy()))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
	return true
}

var (
	settlementMarkerCastleSprite   *ebiten.Image
	settlementMarkerHarbourSprite  *ebiten.Image
	settlementMarkerSiegeSprite    *ebiten.Image
	settlementMarkerSwordSprite    *ebiten.Image
	settlementMarkerStarSprite     *ebiten.Image
	settlementMarkerCitySprite     *ebiten.Image
	settlementMarkerDistrictSprite *ebiten.Image
	settlementMarkerSpritesLoaded  bool

	eventIconPlague   *ebiten.Image
	eventIconFamine   *ebiten.Image
	eventIconRevolt   *ebiten.Image
	eventIconBlessing *ebiten.Image
	eventIconNotify   *ebiten.Image
	eventIconsReady   bool
)

func ensureSettlementMarkerSprites() {
	if settlementMarkerSpritesLoaded {
		return
	}
	settlementMarkerSpritesLoaded = true
	base := resolveMarkerAssetDir()
	if base == "" {
		return
	}
	settlementMarkerCastleSprite = tryLoadImage(filepath.Join(base, "castle.png"))
	settlementMarkerHarbourSprite = tryLoadImage(filepath.Join(base, "harbour.png"))
	settlementMarkerSiegeSprite = tryLoadImage(filepath.Join(base, "siege.png"))
	settlementMarkerSwordSprite = tryLoadImage(filepath.Join(base, "sword.png"))
	settlementMarkerStarSprite = tryLoadImage(filepath.Join(base, "star.png"))
	settlementMarkerCitySprite = tryLoadImage(filepath.Join(base, "city.png"))
	settlementMarkerDistrictSprite = tryLoadImage(filepath.Join(base, "district.png"))
}

func settlementMarkerCastleImage() *ebiten.Image {
	ensureSettlementMarkerSprites()
	return settlementMarkerCastleSprite
}

func settlementMarkerHarbourImage() *ebiten.Image {
	ensureSettlementMarkerSprites()
	return settlementMarkerHarbourSprite
}

func settlementMarkerSiegeImage() *ebiten.Image {
	ensureSettlementMarkerSprites()
	return settlementMarkerSiegeSprite
}

func settlementMarkerSwordImage() *ebiten.Image {
	ensureSettlementMarkerSprites()
	return settlementMarkerSwordSprite
}

func settlementMarkerStarImage() *ebiten.Image {
	ensureSettlementMarkerSprites()
	return settlementMarkerStarSprite
}

func settlementMarkerCityImage() *ebiten.Image {
	ensureSettlementMarkerSprites()
	return settlementMarkerCitySprite
}

func settlementMarkerDistrictImage() *ebiten.Image {
	ensureSettlementMarkerSprites()
	return settlementMarkerDistrictSprite
}

func resolveMarkerAssetDir() string {
	candidates := []string{filepath.Join("assets", "ui", "markers")}
	prefix := ""
	for i := 0; i < 5; i++ {
		prefix = filepath.Join(prefix, "..")
		candidates = append(candidates, filepath.Join(prefix, "assets", "ui", "markers"))
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}

func capitalLabelIconAdvance(isMediumFace bool) float64 {
	if isMediumFace {
		return 24
	}
	return 22
}

// ensureEventIcons event ikon sprite'larını yükler.
func ensureEventIcons() {
	if eventIconsReady {
		return
	}
	eventIconsReady = true
	// Runtime'da renkli daire ikonlar oluştur
	eventIconPlague = createEventIconSprite(color.RGBA{180, 60, 60, 255}, color.RGBA{80, 10, 10, 255})
	eventIconFamine = createEventIconSprite(color.RGBA{200, 140, 40, 255}, color.RGBA{80, 50, 10, 255})
	eventIconRevolt = createEventIconSprite(color.RGBA{210, 50, 50, 255}, color.RGBA{60, 5, 5, 255})
	eventIconBlessing = createEventIconSprite(color.RGBA{60, 180, 60, 255}, color.RGBA{10, 80, 10, 255})
	eventIconNotify = createEventIconSprite(color.RGBA{180, 180, 60, 255}, color.RGBA{80, 80, 10, 255})
}

// createEventIconSprite belirtilen renklerle küçük bir daire ikon oluşturur.
func createEventIconSprite(fill, border color.RGBA) *ebiten.Image {
	size := 20
	img := ebiten.NewImage(size, size)
	// Dış halka (border)
	vector.FillCircle(img, float32(size/2), float32(size/2), float32(size/2), border, true)
	// İç daire (fill)
	vector.FillCircle(img, float32(size/2), float32(size/2), float32(size/2)-2, fill, true)
	return img
}

// eventIconImage tip adına göre uygun event ikonunu döner.
func eventIconImage(eventType string) *ebiten.Image {
	ensureEventIcons()
	switch eventType {
	case "plague":
		return eventIconPlague
	case "famine":
		return eventIconFamine
	case "revolt":
		return eventIconRevolt
	case "blessing":
		return eventIconBlessing
	default:
		return eventIconNotify
	}
}

// drawActiveEventIcons haritada aktif bölge event ikonlarını çizer.
// Aynı bölgede birden fazla event varsa alt alta listelenir.
func (r *Renderer) drawActiveEventIcons(screen *ebiten.Image) {
	if r.gs == nil || len(r.gs.ActiveRegionEvents) == 0 {
		return
	}
	if r.camScale < 0.6 {
		return
	}

	// Aynı bölgedeki event'leri grupla ve sırala
	type regionGroup struct {
		regionID world.RegionID
		events   []state.RegionEventStatus
	}
	groupMap := make(map[world.RegionID]*regionGroup, len(r.gs.ActiveRegionEvents))
	groupOrder := make([]*regionGroup, 0, len(r.gs.ActiveRegionEvents))
	for i := range r.gs.ActiveRegionEvents {
		evt := &r.gs.ActiveRegionEvents[i]
		if !activeRegionEventVisible(r.gs, *evt) {
			continue
		}
		g, exists := groupMap[evt.RegionID]
		if !exists {
			g = &regionGroup{regionID: evt.RegionID}
			groupMap[evt.RegionID] = g
			groupOrder = append(groupOrder, g)
		}
		g.events = append(g.events, *evt)
	}

	for _, g := range groupOrder {
		region := r.gs.Regions[g.regionID]
		if region == nil || region.IsSea {
			continue
		}

		baseSX, baseSY, ok := r.eventIconScreenAnchor(g.regionID)
		if !ok {
			continue
		}

		// Ekran dışı kontrolü (base noktasına göre)
		if baseSX < -50 || baseSX > ScreenWidth+50 || baseSY < -50 || baseSY > ScreenHeight+50 {
			continue
		}

		// Event marker'ları settlement anchor'ın üstüne taşır; çoklu event daha da yukarı stack edilir.
		eventCount := len(g.events)
		for idx, evt := range g.events {
			sx, sy := activeRegionEventMarkerPoint(baseSX, baseSY, eventCount, idx)

			sx, sy = r.clampActiveRegionEventScreenPoint(g.regionID, sx, sy)

			// Ekran dışı kontrolü
			if sx < -30 || sx > ScreenWidth+30 || sy < -50 || sy > ScreenHeight+50 {
				continue
			}

			img := eventIconImage(evt.Type)
			if img == nil {
				continue
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(activeEventIconSize)/float64(img.Bounds().Dx()), float64(activeEventIconSize)/float64(img.Bounds().Dy()))
			op.GeoM.Translate(sx-float64(activeEventIconSize)/2, sy-float64(activeEventIconSize)/2)

			// Süre azaldıkça alpha azalt
			alpha := 1.0
			if evt.TurnsLeft <= 2 {
				alpha = 0.5 + float64(evt.TurnsLeft)*0.25
			}
			op.ColorScale.SetA(float32(alpha))
			screen.DrawImage(img, op)

			// Yüksek zoomda kısa etiket
			if r.camScale >= 1.05 && evt.LabelTR != "" {
				labelText := evt.LabelTR
				if len(labelText) > 18 {
					labelText = labelText[:18] + "..."
				}
				tw := MeasureText(labelText, FaceSmall)
				labelCol := color.RGBA{255, 255, 255, uint8(220 * alpha)}
				shadowCol := color.RGBA{0, 0, 0, uint8(140 * alpha)}
				outlined := gameui.NewOutlinedLabel(gameui.Rect{
					X: sx - tw/2,
					Y: sy - float64(activeEventIconSize)/2 - 14,
				}, labelText, labelCol, shadowCol, gameui.TextSmall, gameui.TextAlignStart)
				outlined.Offsets = [][2]float64{{1, 1}}
				outlined.Draw(screen, renderText)
			}
		}
	}
}

func (r *Renderer) activeRegionEventHitAt(mx, my float64) (int, bool) {
	if r == nil || r.gs == nil || len(r.gs.ActiveRegionEvents) == 0 || r.camScale < 0.6 {
		return -1, false
	}

	type eventView struct {
		idx int
		evt state.RegionEventStatus
	}
	type regionGroup struct {
		regionID world.RegionID
		events   []eventView
	}

	groupMap := make(map[world.RegionID]*regionGroup, len(r.gs.ActiveRegionEvents))
	groupOrder := make([]*regionGroup, 0, len(r.gs.ActiveRegionEvents))
	for i := range r.gs.ActiveRegionEvents {
		evt := r.gs.ActiveRegionEvents[i]
		if !activeRegionEventVisible(r.gs, evt) {
			continue
		}
		g, exists := groupMap[evt.RegionID]
		if !exists {
			g = &regionGroup{regionID: evt.RegionID}
			groupMap[evt.RegionID] = g
			groupOrder = append(groupOrder, g)
		}
		g.events = append(g.events, eventView{idx: i, evt: evt})
	}

	for _, g := range groupOrder {
		region := r.gs.Regions[g.regionID]
		if region == nil || region.IsSea {
			continue
		}

		baseSX, baseSY, ok := r.eventIconScreenAnchor(g.regionID)
		if !ok {
			continue
		}
		if baseSX < -50 || baseSX > ScreenWidth+50 || baseSY < -50 || baseSY > ScreenHeight+50 {
			continue
		}

		eventCount := len(g.events)
		for idx, entry := range g.events {
			sx, sy := activeRegionEventMarkerPoint(baseSX, baseSY, eventCount, idx)

			sx, sy = r.clampActiveRegionEventScreenPoint(g.regionID, sx, sy)
			if sx < -30 || sx > ScreenWidth+30 || sy < -30 || sy > ScreenHeight+30 {
				continue
			}
			if math.Abs(mx-sx) <= float64(activeEventIconSize)/2 && math.Abs(my-sy) <= float64(activeEventIconSize)/2 {
				return entry.idx, true
			}
		}
	}

	return -1, false
}

func activeRegionEventMarkerPoint(baseSX, baseSY float64, eventCount, idx int) (float64, float64) {
	if eventCount <= 0 {
		return baseSX, baseSY
	}

	sx := baseSX
	sy := baseSY - float64(activeEventIconLiftY)
	if eventCount > 1 {
		sy -= float64(activeEventIconSpacingY) * float64(eventCount-1-idx)
	}
	return sx, sy
}

func activeRegionEventVisible(gs *state.GameState, evt state.RegionEventStatus) bool {
	if gs == nil || evt.RegionID == "" {
		return false
	}
	region := gs.Regions[evt.RegionID]
	return region != nil && !region.IsSea
}

func (r *Renderer) clampActiveRegionEventScreenPoint(rid world.RegionID, sx, sy float64) (float64, float64) {
	if r == nil || r.gs == nil || r.worldMap == nil || rid == "" {
		return sx, sy
	}
	region := r.gs.Regions[rid]
	if region == nil || region.IsSea {
		return sx, sy
	}

	wx, wy := r.screenToWorld(sx, sy)
	pureX := int(math.Round((wx - shapeOffX) / shapeScaleX))
	pureY := int(math.Round((wy - shapeOffY) / shapeScaleY))
	if regionContainsPoint(region, float64(pureX), float64(pureY)) {
		return sx, sy
	}

	if px, py, ok := nearestPointInRegionShape(region, pureX, pureY); ok {
		return r.worldToScreen(wcX(px), wcY(py))
	}
	ix, iy := int(math.Round(wx)), int(math.Round(wy))
	if px, py, ok := r.worldMap.nearestRegionPixel(rid, ix, iy); ok {
		return r.worldToScreen(float64(px), float64(py))
	}
	return sx, sy
}

func (r *Renderer) activeRegionEventHovering(fx, fy float64) bool {
	_, ok := r.activeRegionEventHitAt(fx, fy)
	return ok
}

func (r *Renderer) activeRegionEventDetailAt(idx int) string {
	if r == nil || r.gs == nil || idx < 0 || idx >= len(r.gs.ActiveRegionEvents) {
		return ""
	}
	evt := r.gs.ActiveRegionEvents[idx]
	title := evt.LabelTR
	if title == "" {
		title = evt.EventID
	}
	regionName := string(evt.RegionID)
	if region := r.gs.Regions[evt.RegionID]; region != nil && region.NameTR != "" {
		regionName = region.NameTR
	}
	lines := []string{
		title,
		"",
		"Kaynak: Harita izi",
		"",
		"Bölge: " + regionName,
		"Tip: " + activeRegionEventTypeLabel(evt.Type),
		"Kalan tur: " + strconv.Itoa(evt.TurnsLeft),
	}
	if evt.EventID != "" {
		lines = append(lines, "Event ID: "+evt.EventID)
	}
	return strings.Join(lines, "\n")
}

func activeRegionEventTypeLabel(eventType string) string {
	switch eventType {
	case "plague":
		return "Veba"
	case "famine":
		return "Kıtlık"
	case "revolt":
		return "İsyan"
	case "blessing":
		return "Bereket"
	default:
		return "Bildirim"
	}
}

func (r *Renderer) eventIconAnchor(rid world.RegionID) (int, int, bool) {
	if r == nil {
		return 0, 0, false
	}
	if r.worldMap != nil {
		if ax, ay, ok := r.worldMap.PrimarySettlementAnchor(rid); ok {
			return ax, ay, true
		}
		if ax, ay, ok := r.worldMap.RegionAnchor(rid); ok {
			return ax, ay, true
		}
	}
	if r.gs != nil {
		if region := r.gs.Regions[rid]; region != nil {
			return int(math.Round(wcX(region.WorldX))), int(math.Round(wcY(region.WorldY))), true
		}
	}
	return 0, 0, false
}

func (r *Renderer) eventIconScreenAnchor(rid world.RegionID) (float64, float64, bool) {
	ax, ay, ok := r.eventIconAnchor(rid)
	if !ok {
		return 0, 0, false
	}
	sx, sy := r.worldToScreen(float64(ax), float64(ay))
	return sx, sy, true
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
