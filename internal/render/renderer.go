package render

import (
	"image/color"
	"math"

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

	r.worldMap.Refresh(r.gs, r.SelectedRegion)

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
	DrawArmyPanel(screen, r.gs, r.SelectedArmy)
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

	// 10. Tarihsel olay tam ekran popup
	if r.showHistoricalEvent {
		drawHistoricalEventPopup(screen, r.historicalEventTitle, r.historicalEventDesc)
	}
}

// drawSelectionHighlight seçili bölgenin kenar piksellerini vurgular.
func (r *Renderer) drawSelectionHighlight(screen *ebiten.Image) {
	region, ok := r.gs.Regions[r.SelectedRegion]
	if !ok || region.IsSea {
		return
	}

	sx, sy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
	vector.StrokeCircle(screen, float32(sx), float32(sy+4), 16, 3, color.RGBA{255, 220, 70, 230}, true)
	vector.StrokeCircle(screen, float32(sx), float32(sy+4), 22, 1.5, color.RGBA{30, 20, 5, 180}, true)
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
		if !ok || nRegion.IsSea || nRegion.IsLocked {
			continue
		}
		sx, sy := r.worldToScreen(wcX(nRegion.WorldX), wcY(nRegion.WorldY))

		var col color.RGBA
		switch {
		case nRegion.OwnerID != "" && nRegion.OwnerID != a.OwnerID:
			key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(nRegion.OwnerID))
			rel, exists := r.gs.Relations[key]
			if exists && rel.Stance == faction.StanceWar {
				col = color.RGBA{220, 60, 60, 200} // düşman, savaş halinde — saldır
			} else {
				col = color.RGBA{140, 140, 140, 130} // düşman ama barış — girilmez
			}
		case nRegion.OwnerID == "":
			col = color.RGBA{60, 220, 60, 200} // sahipsiz — fetih
		default:
			col = color.RGBA{80, 160, 255, 160} // kendi bölgesi — salt hareket
		}

		vector.StrokeCircle(screen, float32(sx), float32(sy), 18, 3, col, true)
	}
}

// drawArmies tüm orduları harita üzerinde çizer.
func (r *Renderer) drawArmies(screen *ebiten.Image) {
	for aid, a := range r.gs.Armies {
		region, ok := r.gs.Regions[a.RegionID]
		if !ok {
			continue
		}
		sx, sy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
		// İkon şehir noktasının üstünde
		iconX := float32(sx)
		iconY := float32(sy) - 22

		// Fraksiyon rengi
		fc := factionColor(r.gs, a.OwnerID)

		isSelected := aid == r.SelectedArmy
		r.drawArmyIcon(screen, iconX, iconY, fc, len(a.Units), isSelected)
	}
}

// drawArmyIcon tek bir ordu ikonunu çizer.
func (r *Renderer) drawArmyIcon(screen *ebiten.Image, cx, cy float32, col color.RGBA, unitCount int, selected bool) {
	// Dış kare arka plan
	half := float32(10)
	borderCol := color.RGBA{200, 200, 200, 220}
	if selected {
		borderCol = color.RGBA{255, 215, 0, 255}
	}
	vector.FillRect(screen, cx-half-2, cy-half-2, half*2+4, half*2+4, borderCol, false)
	vector.FillRect(screen, cx-half, cy-half, half*2, half*2, col, false)

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
		r.SelectedArmy = ""
		r.SelectedRegion = ""
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

	// Ordu ikonu tıklaması → seç / seçimi kaldır
	for aid, a := range r.gs.Armies {
		region, ok := r.gs.Regions[a.RegionID]
		if !ok {
			continue
		}
		sx, sy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
		dx := fx - sx
		dy := fy - (sy - 22)
		if math.Sqrt(dx*dx+dy*dy) < 14 {
			if r.SelectedArmy == aid {
				r.SelectedArmy = ""
				return InputAction{}
			}
			r.SelectedArmy = aid
			r.SelectedRegion = ""
			return InputAction{Kind: ActionSelectArmy, ArmyID: aid}
		}
	}

	// Bölge seçimi
	wx, wy := r.screenToWorld(fx, fy)
	rid := r.worldMap.RegionAt(int(wx), int(wy))
	if rid != "" {
		if region, ok := r.gs.Regions[rid]; ok && region.IsSea {
			rid = ""
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
		if n == rid {
			act := InputAction{Kind: ActionMoveArmy, ArmyID: r.SelectedArmy, TargetRegion: rid}
			r.SelectedArmy = ""
			return act
		}
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
