package render

import (
	"fmt"
	"math"
	"sort"
	"time"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/audio"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

const mapRegionDoubleClickWindow = 400 * time.Millisecond

func (r *Renderer) HandleInput() InputAction {
	r.updateCursorShape()
	r.updateEditDropdownPositions()

	// Tarihsel olay popup'ı çizimde en üstte olduğundan inputta da ilk öncelik olmalı.
	if r.showHistoricalEvent {
		return r.handleHistoricalEventInput()
	}
	if r.showCommanderPanel {
		return r.handleCommanderPanelInput()
	}
	if r.showMerchantRoutePanel {
		return r.handleMerchantRoutePanelInput()
	}

	// Onay diyaloğu açıkken normal input engellenir
	if r.confirmDialog.show {
		return r.handleConfirmDialogInput()
	}
	if r.warConfirm.show {
		return r.handleWarConfirmInput()
	}
	if r.warSummary.show {
		return r.handleWarSummaryInput()
	}
	if r.battlePlan.show {
		return r.handleBattlePlanInput()
	}
	if offerIdx, ok := r.playerDiplomacyOfferIndex(); ok {
		return r.handleDiplomacyOfferInput(offerIdx)
	}
	if r.battleReport.show {
		mx, my := ebiten.CursorPosition()
		if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyEnter) ||
			r.keyJustPressed(ebiten.KeySpace) || (r.mouseJustPressed(ebiten.MouseButtonLeft) &&
			(battleReportCloseHit(float64(mx), float64(my)) || battleReportContinueHit(float64(mx), float64(my)) ||
				!battleReportPopupHit(float64(mx), float64(my)))) {
			r.HideBattleReport()
		}
		return InputAction{}
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
	if r.showAIDiagnostic {
		return r.handleAIDiagnosticInput()
	}
	if r.gs.DevelopmentMode && r.keyJustPressed(ebiten.KeyF3) {
		r.toggleAIDiagnostic()
		return InputAction{}
	}
	if r.worldInputLockedByPhase() {
		if r.keyJustPressed(ebiten.KeyF11) {
			ebiten.SetFullscreen(!ebiten.IsFullscreen())
		}
		return InputAction{}
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

	// Tech panel aktifken girişi kamera ve harita işlemlerinden önce yönlendir.
	// Böylece panel üzerindeki tekerlek/drag ve panel dışındaki tıklamalar
	// arka plandaki harita zoom, pan veya seçim akışına sızmaz.
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
			r.clearArmySplitSelection()
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
		r.diplomacyListSort = diplomacyListSortAlphabetical
		r.diplomacyActionFocus = 0
		r.diplomacyTargetFaction = ""
		r.diplomacyHistoryVisible = false
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
		r.clearArmySplitSelection()
		return InputAction{}
	}
	if r.selectedFactionPanel != "" && factionPanelCloseHit(fx, fy) {
		r.closeFactionPanel()
		return InputAction{}
	}
	if r.settlementPanelCloseHit(fx, fy) {
		r.clearSelectedSettlement()
		return InputAction{}
	}
	if r.SelectedRegion != "" && regionPanelCloseHit(fx, fy) {
		r.SelectedRegion = ""
		r.closeFactionPanel()
		r.devNeighborListExpanded = false
		r.regionPanelTab = regionPanelTabBuildings
		r.regionPanelScroll = 0
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
	if turnTechHudTechHit(fx, fy) {
		r.showTech = true
		r.showRecruitPanel = false
		r.showDiplomacy = false
		r.techCursor = 0
		return InputAction{}
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
		r.diplomacyListSort = diplomacyListSortAlphabetical
		r.diplomacyActionFocus = 0
		r.diplomacyTargetFaction = ""
		r.diplomacyHistoryVisible = false
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
		if tab, ok := regionPanelTabHit(fx, fy, r.gs, r.SelectedRegion); ok {
			if tab == regionPanelTabEvents && r.regionPanelTab != regionPanelTabEvents {
				// Olaylar sekmesi mevcut davranışta komşuları açık gösterir;
				// kullanıcı isterse başlıktaki düğmeyle daraltabilir.
				r.devNeighborListExpanded = true
			}
			r.regionPanelTab = tab
			r.regionPanelScroll = 0
			return InputAction{}
		}
		if fid, ok := regionOwnerNameHit(fx, fy, r.gs, r.SelectedRegion); ok {
			r.openFactionPanel(fid)
			r.clearSelectedSettlement()
			r.showRecruitPanel = false
			r.resetRecruitSelection()
			return InputAction{}
		}
		if delta := regionTaxButtonHit(fx, fy, r.gs, r.SelectedRegion); delta != 0 {
			return InputAction{Kind: ActionAdjustTax, TargetRegion: r.SelectedRegion, Delta: delta}
		}
		if regionGrainAidButtonHitForTab(fx, fy, r.gs, r.SelectedRegion, r.regionPanelTab) {
			if r.gs.CanApplyGrainAid(r.SelectedRegion) {
				return InputAction{Kind: ActionGrainAid, TargetRegion: r.SelectedRegion}
			}
			return InputAction{}
		}
		if regionDiplomacyButtonHitForTab(fx, fy, r.gs, r.SelectedRegion, r.regionPanelTab) {
			region := r.gs.Regions[r.SelectedRegion]
			if region != nil && region.OwnerID != "" && region.OwnerID != string(r.gs.PlayerFactionID) {
				r.openDiplomacyTarget(faction.FactionID(region.OwnerID), 0)
				return InputAction{}
			}
		}
		if r.regionPanelTab == regionPanelTabEvents {
			if idx, ok := regionActiveEventPanelHit(fx, fy, r.gs, r.gs.Regions[r.SelectedRegion], r.regionPanelScroll); ok {
				r.eventDetail = r.activeRegionEventDetailAt(idx)
				return InputAction{}
			}
			if regionActivityNeighborToggleHit(fx, fy, r.gs, r.gs.Regions[r.SelectedRegion], r.regionPanelScroll) {
				r.devNeighborListExpanded = !r.devNeighborListExpanded
				r.regionPanelScroll = clampRegionPanelScrollForNeighbors(r.gs, r.SelectedRegion, r.regionPanelScroll, r.devNeighborListExpanded)
				return InputAction{}
			}
		}
		if regionNeighborToggleHit(fx, fy, r.gs, r.SelectedRegion) {
			r.devNeighborListExpanded = !r.devNeighborListExpanded
			return InputAction{}
		}
		if r.regionPanelTab == regionPanelTabBuildings {
			if bid := BuildingGridHitTest(fx, fy, r.gs, r.SelectedRegion, r.devNeighborListExpanded); bid != "" {
				return InputAction{Kind: ActionBuild, TargetRegion: r.SelectedRegion, BuildingID: bid}
			}
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
		assaultBtn, liftBtn, surrenderBtn := buildSelectedSiegeButtons()
		attacker := r.gs.Armies[r.SelectedArmy]
		if attacker != nil && assaultBtn.HitTest(fx, fy) {
			return InputAction{Kind: ActionAssaultSiege, ArmyID: r.SelectedArmy, TargetRegion: siege.RegionID, BattleStance: combat.BattleStanceBalanced}
		}
		if liftBtn.HitTest(fx, fy) {
			return InputAction{Kind: ActionLiftSiege, ArmyID: r.SelectedArmy, TargetRegion: siege.RegionID}
		}
		if attacker != nil {
			_, canSend := r.attackerSiegeSurrenderState(attacker, r.gs.Regions[siege.RegionID])
			if canSend && surrenderBtn.HitTest(fx, fy) {
				return InputAction{Kind: ActionProposeSiegeSurrender, ArmyID: r.SelectedArmy, TargetRegion: siege.RegionID}
			}
		}
		if r.selectedSiegePanelHit(fx, fy) {
			return InputAction{}
		}
	}
	if defender, _, siege, _, surrenderOffered, ok := r.selectedDefensiveSiegePanelState(); ok {
		sortieBtn, surrenderBtn := buildDefensiveSiegeButtons()
		if defender != nil && sortieBtn.HitTest(fx, fy) {
			return InputAction{Kind: ActionSortieSiege, ArmyID: defender.ID, TargetRegion: siege.RegionID, BattleStance: combat.BattleStanceBalanced}
		}
		if surrenderOffered && surrenderBtn.HitTest(fx, fy) {
			armyID := army.ArmyID("")
			if defender != nil {
				armyID = defender.ID
			}
			return InputAction{Kind: ActionSurrenderSiege, ArmyID: armyID, TargetRegion: siege.RegionID}
		}
		if r.selectedSiegePanelHit(fx, fy) {
			return InputAction{}
		}
	}
	if r.SelectedArmy != "" && ArmyPanelBoundsHit(fx, fy, r.gs, r.SelectedArmy) {
		if r.selectedArmyIsPlayerOwned() && merchantRouteButtonHit(fx, fy, r.gs, r.SelectedArmy) {
			r.openMerchantRoutePanel()
			return InputAction{}
		}
		if r.selectedArmyIsPlayerOwned() && CommanderPortraitHitTest(fx, fy, r.gs, r.SelectedArmy) {
			r.OpenCommanderPanel(r.SelectedArmy)
			return InputAction{}
		}
		if r.selectedArmyIsPlayerOwned() {
			if unitIndex, ok := ArmyPanelUnitHit(fx, fy, r.gs, r.SelectedArmy); ok {
				r.toggleArmySplitUnit(unitIndex)
				return InputAction{}
			}
		}
		if r.selectedArmyIsPlayerOwned() && SplitButtonHitTest(fx, fy, r.gs, r.SelectedArmy, r.splitSelectedUnits) {
			indices := r.selectedArmySplitIndices()
			r.clearArmySplitSelection()
			return InputAction{Kind: ActionSplitArmy, ArmyID: r.SelectedArmy, UnitIndices: indices}
		}
		if r.selectedArmyIsPlayerOwned() && MergeButtonHitTest(fx, fy, r.gs, r.SelectedArmy) {
			return InputAction{Kind: ActionMergeArmies, ArmyID: r.SelectedArmy}
		}
		return InputAction{}
	}
	if aid, ok := r.armyHitAt(fx, fy); ok {
		if r.SelectedArmy == aid {
			r.SelectedArmy = ""
			r.clearArmySplitSelection()
			return InputAction{}
		}
		r.clearArmySplitSelection()
		r.SelectedArmy = aid
		r.SelectedRegion = ""
		r.closeFactionPanel()
		r.clearSelectedSettlement()
		r.showRecruitPanel = false
		r.resetRecruitSelection()
		return InputAction{Kind: ActionSelectArmy, ArmyID: aid}
	}
	if rid, idx, ok := r.settlementHitAt(fx, fy); ok {
		if r.selectMapRegionFromMapClick(rid) {
			return InputAction{}
		}
		r.selectSettlement(rid, idx)
		if !RecruitPanelVisible(r.gs, rid) {
			r.showRecruitPanel = false
		}
		r.resetRecruitSelection()
		return InputAction{}
	}
	if region, settlement, ok := r.selectedSettlement(); ok && region != nil && region.ID == r.SelectedRegion {
		if btn, active := settlementCapitalActionButton(r.gs, region, settlement); active && btn.HitTest(fx, fy) {
			name := settlement.NameTR
			if name == "" {
				name = settlement.Name
			}
			if name == "" {
				name = region.NameTR
			}
			msg := fmt.Sprintf("%s yerleşimini başkent yapma süreci başlasın mı? Taşıma %d tur sürer.", name, state.DefaultCapitalMoveTurns)
			r.ShowConfirmDialog("Başkent Taşı", msg, "Başlat", "İptal", InputAction{
				Kind:         ActionScheduleCapitalMove,
				TargetRegion: region.ID,
				BuildingID:   settlement.ID,
			}, nil)
			return InputAction{}
		}
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

	// Bölge / deniz bölgesi seçimi
	wx, wy := r.screenToWorld(fx, fy)
	rid := r.worldMap.RegionAt(int(wx), int(wy))
	// Deniz bölgesi sol tıkta sadece seçilir; hareket sağ tıkla verilir.
	// Kara bölgesine çift tıklanırsa seçimden sonra bölge sahibinin diplomasi
	// teklif paneli açılır.
	r.selectMapRegionFromMapClick(rid)
	return InputAction{}
}

func (r *Renderer) selectMapRegionFromMapClick(rid world.RegionID) bool {
	doubleClick := r.mapRegionDoubleClicked(rid)
	r.selectMapRegion(rid)
	if !doubleClick || r.gs == nil {
		return false
	}
	region := r.gs.Regions[rid]
	if region == nil || region.IsSea || region.OwnerID == "" || region.OwnerID == string(r.gs.PlayerFactionID) {
		return false
	}
	r.openDiplomacyTarget(faction.FactionID(region.OwnerID), 0)
	r.resetMapRegionDoubleClick()
	return true
}

func (r *Renderer) mapRegionDoubleClicked(rid world.RegionID) bool {
	if rid == "" {
		r.resetMapRegionDoubleClick()
		return false
	}
	now := time.Now()
	doubleClick := r.lastMapRegionClickID == rid &&
		!r.lastMapRegionClickAt.IsZero() &&
		now.Sub(r.lastMapRegionClickAt) <= mapRegionDoubleClickWindow
	r.lastMapRegionClickID = rid
	r.lastMapRegionClickAt = now
	return doubleClick
}

func (r *Renderer) resetMapRegionDoubleClick() {
	r.lastMapRegionClickID = ""
	r.lastMapRegionClickAt = time.Time{}
}

func (r *Renderer) selectMapRegion(rid world.RegionID) {
	r.SelectedArmy = ""
	r.clearArmySplitSelection()
	if r.SelectedRegion != rid {
		r.devNeighborListExpanded = true
		r.regionPanelTab = regionPanelTabBuildings
		r.regionPanelScroll = 0
	}
	r.SelectedRegion = rid
	r.syncFactionPanelToSelectedRegion()
	r.clearSelectedSettlement()
	// Birim oluşturma paneli yalnızca alt paneldeki Ordu butonuyla açılır.
	r.showRecruitPanel = false
	r.resetRecruitSelection()
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

func (r *Renderer) openFactionPanel(fid faction.FactionID) {
	r.selectedFactionPanel = fid
	r.factionPanelScroll = 0
}

func (r *Renderer) syncFactionPanelToSelectedRegion() {
	if r.selectedFactionPanel == "" {
		return
	}
	if r.gs == nil {
		r.closeFactionPanel()
		return
	}
	region := r.gs.Regions[r.SelectedRegion]
	if region == nil || region.OwnerID == "" {
		r.closeFactionPanel()
		return
	}
	ownerID := faction.FactionID(region.OwnerID)
	if ownerID == r.selectedFactionPanel {
		return
	}
	r.openFactionPanel(ownerID)
}

func (r *Renderer) closeFactionPanel() {
	r.selectedFactionPanel = ""
	r.factionPanelScroll = 0
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

func (r *Renderer) clearArmySplitSelection() {
	r.splitSelectedUnits = nil
}

func (r *Renderer) toggleArmySplitUnit(unitIndex int) {
	if unitIndex < 0 {
		return
	}
	if r.splitSelectedUnits == nil {
		r.splitSelectedUnits = make(map[int]bool)
	}
	if r.splitSelectedUnits[unitIndex] {
		delete(r.splitSelectedUnits, unitIndex)
		return
	}
	r.splitSelectedUnits[unitIndex] = true
}

func (r *Renderer) selectedArmySplitIndices() []int {
	if len(r.splitSelectedUnits) == 0 {
		return nil
	}
	a := r.gs.Armies[r.SelectedArmy]
	if a == nil {
		return nil
	}
	indices := make([]int, 0, len(r.splitSelectedUnits))
	for index, selected := range r.splitSelectedUnits {
		if selected && index >= 0 && index < len(a.Units) {
			indices = append(indices, index)
		}
	}
	sort.Ints(indices)
	return indices
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
		r.clearArmySplitSelection()
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
		isAlliedRegion := false
		if target.OwnerID != "" && target.OwnerID != a.OwnerID {
			key := faction.RelationKey(faction.FactionID(a.OwnerID), faction.FactionID(target.OwnerID))
			if diplomacy.SameRealm(r.gs, faction.FactionID(a.OwnerID), faction.FactionID(target.OwnerID)) {
				isAlliedRegion = true
			} else if rel, exists := r.gs.Relations[key]; exists && rel.Stance == faction.StanceAllied {
				isAlliedRegion = true
			}
		}
		allySieging := false
		if !a.IsNaval && target.CanLandEnter() && target.OwnerID != "" && target.OwnerID != a.OwnerID && target.IsFortified() {
			if r.canJoinActiveSiege(a, rid) {
				allySieging = true
			} else if siege := r.gs.SiegeAt(rid); siege != nil && siege.AttackerArmyID != a.ID {
				if !r.canEnterActiveSiegedRegion(a, rid) {
					r.ShowCombatResult("Bu bölge zaten başka bir ordu tarafından kuşatılıyor.")
					return InputAction{}
				}
			}
		}
		enemyArmy := r.gs.SelectBattleDefender(a, rid, a.IsNaval && target.CanNavalEnter())
		targetSiege := r.gs.SiegeAt(rid)
		battleAction, battleContext, opensBattlePlan := r.battlePlanIntent(a, target, enemyArmy)
		amphibiousSiegeLanding := renderTargetRequiresAmphibiousSiegeLanding(r.gs, a, target)
		// Düşman kara bölgesi ama savaş yok → onay diyalogu aç.
		// Donanma-deniz hareketinde savaş ilanı zorunlu değil.
		// Müttefik bölgesine çıkarma için "Karaya In" göster.
		if !(a.IsNaval && target.CanNavalEnter()) && !navalCanDockAtRegion(r.gs, a, target) && target.OwnerID != "" && target.OwnerID != a.OwnerID {
			if armyRegionIsFriendly(r.gs, a, target) && len(a.EmbarkedUnits) > 0 {
				// Müttefik kıyısına çıkarma — savaş popup'ı değil, karaya inme onayı
				r.ShowConfirmDialog(
					"Karaya In",
					"Gemideki birlikler dost bölgeye insin mi?",
					"Karaya In",
					"İptal",
					InputAction{Kind: ActionDisembarkArmy, ArmyID: r.SelectedArmy, TargetRegion: rid},
					nil,
				)
				return InputAction{}
			}
			if shouldPromptWarConfirmForMove(r.gs, a, target) {
				name := target.OwnerID
				if f, ok := r.gs.Factions[faction.FactionID(target.OwnerID)]; ok {
					name = f.NameTR
				}
				pendingEnemy := army.ArmyID("")
				if enemyArmy != nil {
					pendingEnemy = enemyArmy.ID
				}
				r.openWarConfirm(faction.FactionID(target.OwnerID), name, r.SelectedArmy, rid, pendingEnemy, opensBattlePlan && !amphibiousSiegeLanding, battleAction, battleContext)
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
		// Kuşatılan bölgedeki sahip veya müttefik ordu önce kuşatanla huruç
		// savaşı yapmalıdır. Savaş kaynak bölgede, başarılı hareket ise seçilen
		// komşu hedefte çözülür.
		if !a.IsNaval && r.gs.IsArmyDefendingSiegedRegion(a) {
			siege := r.gs.SiegeAt(a.RegionID)
			var sortieDefender *army.Army
			if siege != nil {
				sortieDefender = r.gs.Armies[siege.AttackerArmyID]
			}
			if sortieDefender != nil {
				r.openBattlePlanWithDestination(a, src, sortieDefender, ActionMoveArmy, combat.BattleContextLand, rid)
				return InputAction{}
			}
		}
		if renderTargetRequiresSiegeDecision(r.gs, a, target) && !allySieging && !isAlliedRegion && (targetSiege == nil || targetSiege.AttackerArmyID == a.ID) {
			r.openSiegeDecision(a, target)
			return InputAction{}
		}
		if opensBattlePlan && !allySieging && !amphibiousSiegeLanding {
			r.openBattlePlan(a, target, enemyArmy, battleAction, battleContext)
			return InputAction{}
		}
		act := InputAction{Kind: ActionMoveArmy, ArmyID: r.SelectedArmy, TargetRegion: rid}
		r.SelectedArmy = ""
		r.clearArmySplitSelection()
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
		mx, my := ebiten.CursorPosition()
		if r.SelectedRegion != "" && r.regionPanelTab == regionPanelTabEvents && regionPanelActivityHit(float64(mx), float64(my), r.gs, r.SelectedRegion) {
			r.regionPanelScroll = clampRegionPanelScrollForNeighbors(r.gs, r.SelectedRegion, r.regionPanelScroll-dy*regionPanelScrollStep, r.devNeighborListExpanded)
			return
		}
		if r.selectedFactionPanel != "" && factionPanelHit(float64(mx), float64(my)) {
			r.factionPanelScroll -= dy * factionPanelScrollStep
			if r.factionPanelScroll < 0 {
				r.factionPanelScroll = 0
			}
			return
		}
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
