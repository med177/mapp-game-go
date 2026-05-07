package render

import (
	"image/color"
	"math"
	"sort"

	"mapp-game-go/internal/army"
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
	CurrentSettings Settings

	// Duraklama menüsü
	pauseCursor int

	// Kayıt/yükleme slot seçim ekranı
	slotCursor        int
	saveSelectMode    bool   // true=kaydetme, false=yükleme
	pendingDeleteSlot string // onay bekleyen slot adı ("" = onay yok)

	// Olay logu (sağ üst panel)
	eventLog []string

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

	// Savaş ilan onay diyalogu
	warConfirm warConfirmState
}

type warConfirmState struct {
	show        bool
	factionName string
	factionID   string
	pendingArmy army.ArmyID
	pendingDest world.RegionID
}

// New başlangıç kamera pozisyonuyla yeni bir Renderer döner.
func New(gs *state.GameState) *Renderer {
	r := &Renderer{
		gs:        gs,
		worldMap:  NewWorldMap(gs),
		prevKeys:  make(map[ebiten.Key]bool),
		prevMouse: make(map[ebiten.MouseButton]bool),
	}
	r.resetCamera()
	return r
}

// resetCamera kamerayı mevcut ScreenWidth/ScreenHeight'e göre dünyayı tam dolduracak şekilde ayarlar.
func (r *Renderer) resetCamera() {
	scaleX := ScreenWidth / float64(WorldW)
	scaleY := ScreenHeight / float64(WorldH)
	r.camScale = math.Min(scaleX, scaleY)
	r.camX = float64(WorldW) / 2
	// Haritanın üst kenarını ekranın üstüne hizala
	r.camY = ScreenHeight / (2 * r.camScale)
}

// SetCursor menü veya ekran imlecini sıfırlar.
func (r *Renderer) SetCursor(n int) { r.factionCursor = n }

// MarkMapDirty sahiplik değiştiğinde çağrılır.
func (r *Renderer) MarkMapDirty() { r.worldMap.MarkDirty() }

// ReloadGameState yükleme sonrası yeni state ve yeni worldmap ile günceller.
func (r *Renderer) ReloadGameState(gs *state.GameState) {
	r.gs = gs
	r.worldMap = NewWorldMap(gs)
	r.SelectedRegion = ""
	r.SelectedArmy = ""
}

// AddEvent olay loguna yeni bir giriş ekler (maksimum 8).
func (r *Renderer) AddEvent(msg string) {
	r.eventLog = append([]string{msg}, r.eventLog...)
	if len(r.eventLog) > 8 {
		r.eventLog = r.eventLog[:8]
	}
}

// ShowCombatResult savaş sonuç mesajını ekranda ~3 saniye gösterir.
func (r *Renderer) ShowCombatResult(msg string) {
	r.combatLog = msg
	r.combatLogTimer = 180
	r.AddEvent(msg)
}

// ShowHistoricalEvent büyük tarihsel olayı tam ekran popup olarak gösterir.
func (r *Renderer) ShowHistoricalEvent(title, desc string) {
	r.historicalEventTitle = title
	r.historicalEventDesc = desc
	r.showHistoricalEvent = true
	r.AddEvent(title)
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
		DrawMainMenu(screen, r.factionCursor, r.HasSave, r.menuTick)
		return
	}

	// Ayarlar ekranı
	if r.gs.Phase == state.PhaseSettings {
		DrawSettingsScreen(screen, r.CurrentSettings, r.factionCursor)
		return
	}

	// Fraksiyon seçim ekranı
	if r.gs.Phase == "faction_select" {
		DrawFactionSelect(screen, r.gs, r.factionCursor)
		return
	}

	// Zafer koşulu seçim ekranı
	if r.gs.Phase == "victory_select" {
		DrawVictorySelect(screen, r.factionCursor)
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
		r.applyMapGeoM(mapOp, WorldW, WorldH)
		screen.DrawImage(r.worldMap.Image(), mapOp)
		r.menuTick++
		DrawPauseMenu(screen, r.pauseCursor, r.HasSave, r.menuTick)
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
	r.applyMapGeoM(mapOp, WorldW, WorldH)
	screen.DrawImage(r.worldMap.Image(), mapOp)

	// 2. Seçim vurgusu (bölge)
	if r.SelectedRegion != "" {
		r.drawSelectionHighlight(screen)
	}

	// 3. Ordu hareket hedefleri
	if r.SelectedArmy != "" {
		r.drawMoveTargets(screen)
	}

	// 4. Bölge etiketleri
	r.drawRegionLabels(screen)

	// 5. Ordu ikonları
	r.drawArmies(screen)

	// 6. UI panelleri
	DrawBottomPanel(screen, r.gs, r.showDiplomacy, r.showTech)
	DrawRegionPanel(screen, r.gs, r.SelectedRegion)
	DrawRecruitPanel(screen, r.gs, r.SelectedRegion)
	DrawArmyDetailPanel(screen, r.gs, r.SelectedArmy)
	DrawMinimap(screen, r.gs, r.camX, r.camY, r.camScale)
	DrawEventLog(screen, r.eventLog)

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
		DrawText(screen, r.combatLog, ScreenWidth/2-150, ScreenHeight/2-40, FaceLarge,
			color.RGBA{255, 220, 80, alpha})
		r.combatLogTimer--
	}

	// 10. Savaş ilan onay diyalogu (diğer popupların altında kalmaması için üst katman)
	if r.warConfirm.show {
		r.drawWarConfirmDialog(screen)
	}

	// 11. Tarihsel olay tam ekran popup
	if r.showHistoricalEvent {
		drawHistoricalEventPopup(screen, r.historicalEventTitle, r.historicalEventDesc)
	}
}

// drawSelectionHighlight seçili bölgenin üstüne vurgu çizer.
func (r *Renderer) drawSelectionHighlight(screen *ebiten.Image) {
	region, ok := r.gs.Regions[r.SelectedRegion]
	if !ok {
		return
	}

	sx, sy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))

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

// drawMoveTargets seçili ordunun gidebileceği komşu bölgeleri vurgular.
func (r *Renderer) drawMoveTargets(screen *ebiten.Image) {
	a, ok := r.gs.Armies[r.SelectedArmy]
	if !ok || a.MovePoints <= 0 {
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

		sx, sy := r.worldToScreen(wcX(nRegion.WorldX), wcY(nRegion.WorldY))

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

// armyIconPositions tüm orduların ekran koordinatlarını hesaplar.
// Aynı bölgedeki birden fazla ordu yan yana offset'lenir.
func (r *Renderer) armyIconPositions() []armyIconPos {
	const iconStep = float32(26) // ikon genişliği 20 + 6px boşluk

	byRegion := map[world.RegionID][]army.ArmyID{}
	for aid, a := range r.gs.Armies {
		byRegion[a.RegionID] = append(byRegion[a.RegionID], aid)
	}

	var result []armyIconPos
	for rid, aids := range byRegion {
		region, ok := r.gs.Regions[rid]
		if !ok {
			continue
		}
		sx, sy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
		baseY := float32(sy) - 22

		sort.Slice(aids, func(i, j int) bool { return aids[i] < aids[j] })

		n := float32(len(aids))
		startX := float32(sx) - (n-1)*iconStep/2
		for i, aid := range aids {
			result = append(result, armyIconPos{
				ArmyID: aid,
				X:      startX + float32(i)*iconStep,
				Y:      baseY,
			})
		}
	}
	return result
}

// drawArmies tüm orduları harita üzerinde çizer.
func (r *Renderer) drawArmies(screen *ebiten.Image) {
	for _, pos := range r.armyIconPositions() {
		a, ok := r.gs.Armies[pos.ArmyID]
		if !ok {
			continue
		}
		fc := factionColor(r.gs, a.OwnerID)
		isSelected := pos.ArmyID == r.SelectedArmy
		r.drawArmyIcon(screen, pos.X, pos.Y, fc, len(a.Units), isSelected, a.IsNaval)
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
	countStr := itoa(unitCount)
	tw := MeasureText(countStr, FaceSmall)
	DrawText(screen, countStr, float64(cx)-tw/2, float64(cy)-5, FaceSmall, color.RGBA{255, 255, 255, 255})
}

// drawRegionLabels zoom yeterliyse bölge isimlerini merkeze yazar.
func (r *Renderer) drawRegionLabels(screen *ebiten.Image) {
	if r.camScale < 0.5 {
		return
	}

	labelCol := color.RGBA{255, 255, 255, 220}
	shadowCol := color.RGBA{0, 0, 0, 160}

	for _, region := range r.gs.Regions {
		if region.IsSea || region.IsLocked {
			continue
		}
		sx, sy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))

		if sx < -50 || sx > ScreenWidth+50 || sy < -20 || sy > ScreenHeight+20 {
			continue
		}

		face := FaceSmall
		if r.camScale >= 1.0 {
			face = FaceMed
		}

		w := MeasureText(region.NameTR, face)
		lx := sx - w/2

		DrawText(screen, region.NameTR, lx+1, sy-6, face, shadowCol)
		DrawText(screen, region.NameTR, lx, sy-7, face, labelCol)

		r.drawCityDot(screen, region, float32(sx), float32(sy))
	}
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

	// Savaş ilan onay diyalogu açıkken normal input engellenir
	if r.warConfirm.show {
		return r.handleWarConfirmInput()
	}

	// Oyun sonu ekranı inputu
	if r.gs.Phase == state.PhaseGameOver {
		if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyEnter) {
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

	// Ana menü inputu
	if r.gs.Phase == state.PhaseMainMenu {
		return r.handleMainMenuInput(r.HasSave)
	}

	// Ayarlar ekranı inputu
	if r.gs.Phase == state.PhaseSettings {
		return r.handleSettingsInput(&r.CurrentSettings)
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

	// Alt bar ve sağ taraf (minimap/eventlog) alanlarında tıklama işleme
	if fy > float64(bottomBarTop()) {
		return InputAction{}
	}
	if fx > float64(evLogX()) {
		return InputAction{}
	}

	// BÖLDÜR butonu tıklaması
	if r.SelectedArmy != "" && SplitButtonHitTest(fx, fy, r.gs, r.SelectedArmy) {
		return InputAction{Kind: ActionSplitArmy, ArmyID: r.SelectedArmy}
	}
	// BİRLEŞTİR butonu tıklaması
	if r.SelectedArmy != "" && MergeButtonHitTest(fx, fy, r.gs, r.SelectedArmy) {
		return InputAction{Kind: ActionMergeArmies, ArmyID: r.SelectedArmy}
	}

	// Asker al paneli tıklaması — bölge seçiminden önce kontrol edilmeli
	if uid := RecruitPanelHitTest(fx, fy, r.gs, r.SelectedRegion); uid != "" {
		return InputAction{Kind: ActionRecruitSpecific, TargetRegion: r.SelectedRegion, BuildingID: uid}
	}

	// Ordu ikonu tıklaması → seç / seçimi kaldır
	for _, pos := range r.armyIconPositions() {
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
			if r.SelectedArmy != "" {
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
	if float64(my) > float64(bottomBarTop()) || float64(mx) > float64(evLogX()) {
		return InputAction{}
	}

	a, ok := r.gs.Armies[r.SelectedArmy]
	if !ok || a.MovePoints <= 0 {
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
		mouseWX, mouseWY := r.screenToWorld(float64(mx), float64(my))
		if dy > 0 && r.camScale < 3.0 {
			r.camScale *= 1.12
		} else if dy < 0 && r.camScale > 0.25 {
			r.camScale /= 1.12
		}
		afterWX, afterWY := r.screenToWorld(float64(mx), float64(my))
		r.camX += mouseWX - afterWX
		r.camY += mouseWY - afterWY
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
	tw2 := MeasureText("Evet - Savaş İlan Et", FaceSmall)
	DrawText(screen, "Evet - Savaş İlan Et", float64(yesX)+(float64(btnDW)-tw2)/2, float64(btnY)+10, FaceSmall, color.RGBA{255, 220, 220, 255})

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
