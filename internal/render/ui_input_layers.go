package render

import (
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	uiLayerMainMenu       = "main-menu"
	uiLayerScreenMenu     = "screen-menu"
	uiLayerPauseMenu      = "pause-menu"
	uiLayerTopStatus      = "hud-top-status"
	uiLayerTopDate        = "hud-top-date"
	uiLayerTurnTech       = "hud-turn-tech"
	uiLayerMusic          = "hud-music"
	uiLayerBottom         = "hud-bottom"
	uiLayerEventLog       = "event-log"
	uiLayerMinimap        = "minimap"
	uiLayerRegion         = "region-panel"
	uiLayerSettlement     = "settlement-panel"
	uiLayerFaction        = "faction-panel"
	uiLayerRecruit        = "recruit-panel"
	uiLayerArmy           = "army-panel"
	uiLayerSiege          = "siege-panel"
	uiLayerActiveWars     = "active-wars-panel"
	uiLayerMerchantRoute  = "merchant-route-panel"
	uiLayerNavalMission   = "naval-mission-panel"
	uiLayerDiplomacy      = "diplomacy-panel"
	uiLayerTech           = "tech-panel"
	uiLayerTrade          = "trade-panel"
	uiLayerImperial       = "imperial-panel"
	uiLayerAIDiagnostic   = "ai-diagnostic"
	uiLayerCommander      = "commander-panel"
	uiLayerShortcuts      = "shortcuts"
	uiLayerPopup          = "popup"
	uiLayerModal          = "modal"
	uiLayerHistorical     = "historical-event"
	uiLayerEventCodex     = "event-codex"
	uiLayerEventDetail    = "event-detail"
	uiLayerVictoryDetail  = "victory-detail"
	uiLayerBattleReport   = "battle-report"
	uiLayerWarSummary     = "war-summary"
	uiLayerBattlePlan     = "battle-plan"
	uiLayerConfirm        = "confirm-dialog"
	uiLayerRegionTask     = "region-task-dialog"
	uiLayerDiplomacyOffer = "diplomacy-offer"
	uiLayerEditInspector  = "edit-inspector"
	uiLayerEditForm       = "edit-form"
	uiLayerEditDropdown   = "edit-dropdown"
)

func uiLayerRect(x, y, w, h float64) gameui.Rect {
	return gameui.Rect{X: x, Y: y, W: w, H: h}
}

func floatRect(x, y, w, h float32) gameui.Rect {
	return uiLayerRect(float64(x), float64(y), float64(w), float64(h))
}

func (r *Renderer) addUILayer(id string, rect gameui.Rect) {
	r.uiLayers.AddRect(id, rect)
}

func (r *Renderer) addUIScreenLayer(id string) {
	r.uiLayers.AddScreen(id, ScreenWidth, ScreenHeight)
}

// uiLayerAllowsAt, mevcut koordinattaki en üst katman verilen UI yüzeyi
// değilse alt yüzey handler'ının çalışmasını engeller.
func (r *Renderer) uiLayerAllowsAt(mx, my float64, ids ...string) bool {
	layer, ok := r.uiLayers.TopAt(mx, my)
	if !ok {
		return true
	}
	for _, id := range ids {
		if layer.ID == id {
			return true
		}
	}
	return false
}

func (r *Renderer) uiLayerAllowsCurrentCursor(ids ...string) bool {
	mx, my := ebiten.CursorPosition()
	return r.uiLayerAllowsAt(float64(mx), float64(my), ids...)
}

// uiLayerAllowsCurrentInput, klavye input'unu cursor panelin dışındayken de
// açık panele ulaştırır; fare/tekerlek input'unda ise yalnızca üst katman
// handler'ının çalışmasına izin verir.
func (r *Renderer) uiLayerAllowsCurrentInput(ids ...string) bool {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) ||
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		return r.uiLayerAllowsCurrentCursor(ids...)
	}
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		return r.uiLayerAllowsCurrentCursor(ids...)
	}
	return true
}

// uiLayerPointerAt, haritayı engelleyen UI alanı ile gerçekten tıklanabilir
// kontrolü ayırır. Panelin bilgi/boş alanı pointer cursor üretmez.
func (r *Renderer) uiLayerPointerAt(mx, my float64) bool {
	layer, ok := r.uiLayers.TopAt(mx, my)
	if !ok {
		return false
	}
	switch layer.ID {
	case uiLayerTopStatus:
		modeButtons := buildMapModeButtons()
		return activeWarsHudButtonHit(mx, my) || (topStatusPanelHit(mx, my) && (r.grainEconomyPopupHovering(mx, my) || r.goldIncomePopupHovering(mx, my) ||
			r.armyOrganizationPopupHovering(mx, my) ||
			victoryProgressHit(mx, my) || modeButtons[0].HitTest(mx, my) || modeButtons[1].HitTest(mx, my) ||
			(imperialPanelAvailable(r.gs) && imperialHUDButtonHit(mx, my))))
	case uiLayerTopDate:
		return topDateHudMenuButtonHit(mx, my)
	case uiLayerTurnTech:
		return turnTechHudTechHit(mx, my) || turnTechHudTradeRouteHit(r.gs, mx, my) || turnTechHudWarFatigueHit(r.gs, mx, my)
	case uiLayerMusic:
		return musicHudInteractiveHit(mx, my)
	case uiLayerBottom:
		return bottomActionButtonHit(mx, my)
	case uiLayerEventLog:
		return eventLogInteractiveHit(mx, my, len(r.eventLog), r.eventLogCollapsed, r.eventLogScroll, r.HasEventCodex())
	case uiLayerMinimap:
		return false
	case uiLayerRegion:
		if r.SelectedRegion == "" {
			return false
		}
		if close := regionPanelCloseHit(mx, my); close {
			return true
		}
		if logistics, ok := regionPanelLogisticsRect(r.gs, r.SelectedRegion); ok && logistics.Hit(mx, my) {
			return true
		}
		return regionPanelInteractiveHitForTab(mx, my, r.gs, r.SelectedRegion, r.regionPanelTab, r.regionPanelScroll)
	case uiLayerSettlement:
		if r.settlementPanelCloseHit(mx, my) {
			return true
		}
		region, settlement, ok := r.selectedSettlement()
		if !ok || region == nil || settlement == nil {
			return false
		}
		button, active := settlementCapitalActionButton(r.gs, region, settlement)
		return active && button.HitTest(mx, my)
	case uiLayerFaction:
		return r.selectedFactionPanel != "" && factionPanelCloseHit(mx, my)
	case uiLayerRecruit:
		return RecruitPanelInteractiveHit(mx, my, r.gs, r.SelectedRegion)
	case uiLayerArmy:
		return r.showArmyDetailPanel && r.SelectedArmy != "" && ArmyPanelInteractiveHit(mx, my, r.gs, r.SelectedArmy, r.splitSelectedUnits)
	case uiLayerSiege:
		return r.selectedSiegePanelHovering(mx, my)
	case uiLayerImperial:
		return r.imperialPanelPointerHit(mx, my)
	case uiLayerPopup:
		return r.combatLogTimer > 0 && r.infoPopupRect().Hit(mx, my)
	case uiLayerAIDiagnostic:
		return false
	case uiLayerCommander:
		return r.commanderPanelHovering(mx, my)
	default:
		return false
	}
}

// rebuildUILayers, Draw sırasındaki UI kompozisyonunu input/cursor sırasına
// taşır. Her panelin haritaya izin verip vermediğini ayrı ayrı kontrol etmek
// yerine bütün yüzeyler tek bir z-order stack'inde tutulur.
func (r *Renderer) rebuildUILayers() {
	r.uiLayers.Reset()
	if r == nil || r.gs == nil {
		return
	}

	switch r.gs.Phase {
	case state.PhaseMainMenu:
		r.addUIScreenLayer(uiLayerMainMenu)
		if r.showShortcuts {
			r.addUIScreenLayer(uiLayerShortcuts)
		}
		return
	case state.PhaseSettings, state.PhaseScenarioSelect, state.PhaseFactionSelect,
		state.PhaseVictorySelect, state.PhaseLoadSelect, state.PhaseSaveSelect,
		state.PhaseGameOver:
		r.addUIScreenLayer(uiLayerScreenMenu)
		return
	case state.PhasePauseMenu:
		r.addUIScreenLayer(uiLayerPauseMenu)
		return
	}

	if r.gs.Phase == state.PhaseEditMode {
		r.addUILayer(uiLayerEditInspector, editInspectorLayerRect())
		if r.editFactionForm.show {
			x, y, w, h := editFactionFormRect()
			r.addUILayer(uiLayerEditForm, floatRect(x, y, w, h))
		}
		if r.editRegionForm.show {
			x, y, w, h := editRegionFormRect()
			r.addUILayer(uiLayerEditForm, floatRect(x, y, w, h))
		}
		if r.editOwnerDropdown != nil && r.editOwnerDropdown.IsOpen() {
			r.addUILayer(uiLayerEditDropdown, r.editOwnerDropdown.Bounds())
		}
		if r.editSuccessorDropdown != nil && r.editSuccessorDropdown.IsOpen() {
			r.addUILayer(uiLayerEditDropdown, r.editSuccessorDropdown.Bounds())
		}
		if r.editTerrainDropdown != nil && r.editTerrainDropdown.IsOpen() {
			r.addUILayer(uiLayerEditDropdown, r.editTerrainDropdown.Bounds())
		}
		if r.editSettlementTypeDropdown != nil && r.editSettlementTypeDropdown.IsOpen() {
			r.addUILayer(uiLayerEditDropdown, r.editSettlementTypeDropdown.Bounds())
		}
		if r.editNewShapeModal.show || r.editRenaming {
			r.addUIScreenLayer(uiLayerModal)
		}
		return
	}

	// Harita üzerindeki HUD ve paneller, Draw içindeki aynı üstten çizim
	// sırasıyla eklenir. Böylece örtüşen iki yüzeyde yalnızca üstteki katman
	// koordinatı sahiplenir.
	r.addUILayer(uiLayerTopStatus, uiLayerRect(0, 0, ScreenWidth, float64(topStatusH)))
	x, y, w, h := topDateHudRect()
	r.addUILayer(uiLayerTopDate, floatRect(x, y, w, h))
	x, y, w, h = turnTechHudRect()
	r.addUILayer(uiLayerTurnTech, floatRect(x, y, w, h))
	x, y, w, h = musicHudRect()
	r.addUILayer(uiLayerMusic, floatRect(x, y, w, h))
	x, y, w, h = bottomActionHudRect()
	r.addUILayer(uiLayerBottom, floatRect(x, y, w, h))
	// Aktif savaş düğmesi müzik HUD'ının sağındaki ortak yardımcı düğme
	// şeridinde çizilir; tarih HUD'ı ile yatayda örtüşebildiği için kendi
	// çizim/hit-test geometrisiyle tüm üst HUD katmanlarından sonra eklenir.
	// Böylece düğme üzerindeyken tarih katmanı cursor kararını gölgelemez.
	activeWarsButton := buildActiveWarsHUDButton()
	r.addUILayer(uiLayerTopStatus, gameui.Rect{
		X: activeWarsButton.X,
		Y: activeWarsButton.Y,
		W: activeWarsButton.W,
		H: activeWarsButton.H,
	})
	r.addUILayer(uiLayerEventLog, uiLayerRect(float64(evLogX()), float64(evLogY()), float64(evLogW), float64(eventLogPanelH(r.eventLogCollapsed))))
	r.addUILayer(uiLayerMinimap, uiLayerRect(float64(minimapX()), float64(minimapY()), float64(minimapW), float64(minimapH)))
	if r.grainEconomyPopupHoveringAtCursor() {
		r.addUILayer(uiLayerPopup, grainEconomyPopupRect())
	}
	if r.goldIncomePopupHoveringAtCursor() {
		r.addUILayer(uiLayerPopup, goldIncomePopupRect())
	}
	if r.armyOrganizationPopupHoveringAtCursor() {
		r.addUILayer(uiLayerPopup, armyOrganizationPopupRect())
	}

	if r.SelectedRegion != "" {
		r.addUILayer(uiLayerRegion, uiLayerRect(float64(infoPanelX()), float64(infoPanelY()), float64(infoPanelW), float64(infoPanelH)))
	}
	if r.isSettlementPanelOpen() {
		r.addUILayer(uiLayerSettlement, uiLayerRect(float64(settlementPanelX()), float64(settlementPanelY()), float64(infoPanelW), float64(infoPanelH)))
	}
	if r.selectedFactionPanel != "" {
		r.addUILayer(uiLayerFaction, uiLayerRect(float64(factionPanelX()), float64(factionPanelY()), float64(infoPanelW), float64(infoPanelH)))
	}
	if r.mapMode != MapModeTrade && r.showRecruitPanel {
		if rect := recruitPanelLayerRect(r.gs, r.SelectedRegion); rect.W > 0 && rect.H > 0 {
			r.addUILayer(uiLayerRecruit, rect)
		}
	}
	if r.showArmyDetailPanel {
		if rect, ok := armyDetailPanelRect(r.gs, r.SelectedArmy); ok {
			r.addUILayer(uiLayerArmy, rect)
		}
	}
	if _, _, _, ok := r.selectedSiegePanelState(); ok {
		r.addUILayer(uiLayerSiege, buildSelectedSiegePanel().Rect)
	}
	if _, _, _, _, _, ok := r.selectedDefensiveSiegePanelState(); ok {
		r.addUILayer(uiLayerSiege, buildSelectedSiegePanel().Rect)
	}

	if r.showDiplomacy {
		if r.diplomacyTargetFaction == "" {
			layout := diplomacyListLayoutForScreen()
			r.addUILayer(uiLayerDiplomacy, layout.panelRect)
			r.addUILayer(uiLayerDiplomacy, layout.historyRect)
		} else {
			layout := diplomacyOfferLayoutForScreen()
			r.addUILayer(uiLayerDiplomacy, layout.panelRect)
			r.addUILayer(uiLayerDiplomacy, layout.historyRect)
			if target := r.gs.Factions[r.diplomacyTargetFaction]; target != nil && target.OverlordID == r.gs.PlayerFactionID {
				r.addUILayer(uiLayerDiplomacy, buildDiplomacyVassalManagementLayout(r.gs, r.diplomacyTargetFaction).panelRect)
			}
		}
	}
	if r.showTech {
		r.addUILayer(uiLayerTech, techPanelLayoutForScreen().panelRect)
	}
	if r.showTrade {
		x, y, w, h := tradePanelRect()
		r.addUILayer(uiLayerTrade, floatRect(x, y, w, h))
	}
	if r.showImperialPanel {
		r.addUILayer(uiLayerImperial, imperialPanelRect())
	}
	r.ensureOverlayPanelOrder()
	for i := 0; i < r.overlayPanelOrderLen; i++ {
		switch panel := r.overlayPanelOrder[i]; panel {
		case overlayPanelMerchantRoute:
			if r.showMerchantRoutePanel {
				r.addUILayer(uiLayerMerchantRoute, merchantRoutePanelRect(r))
			}
		case overlayPanelNavalMission:
			if r.showNavalMissionPanel {
				r.addUILayer(uiLayerNavalMission, r.navalMissionPanelRect())
			}
		case overlayPanelActiveWars:
			if r.showActiveWars {
				r.addUILayer(uiLayerActiveWars, activeWarsPanelRect())
			}
		}
	}
	// Modal yüzeyleri tam ekran eklenir; modal panelinin dışındaki harita da
	// karar penceresi açıkken input alamaz.
	if r.regionTaskDialog.show {
		r.addUIScreenLayer(uiLayerRegionTask)
	}
	if r.confirmDialog.show {
		r.addUIScreenLayer(uiLayerConfirm)
	}
	if r.warConfirm.show {
		r.addUIScreenLayer(uiLayerConfirm)
	}
	if r.warSummary.show {
		r.addUIScreenLayer(uiLayerWarSummary)
	}
	if r.battlePlan.show {
		r.addUIScreenLayer(uiLayerBattlePlan)
	}
	if r.battleReport.show {
		r.addUIScreenLayer(uiLayerBattleReport)
	}
	if _, ok := r.playerDiplomacyOfferIndex(); ok {
		r.addUIScreenLayer(uiLayerDiplomacyOffer)
	}
	if r.showEventCodex {
		r.addUIScreenLayer(uiLayerEventCodex)
	}
	if r.eventDetail != "" {
		r.addUIScreenLayer(uiLayerEventDetail)
	}
	if r.showVictoryDetail {
		r.addUIScreenLayer(uiLayerVictoryDetail)
	}
	if r.combatLogTimer > 0 {
		r.addUILayer(uiLayerPopup, r.infoPopupRect())
	}
	if r.showAIDiagnostic {
		r.addUILayer(uiLayerAIDiagnostic, aiDiagnosticPanelRect())
	}
	if r.showHistoricalEvent {
		r.addUIScreenLayer(uiLayerHistorical)
	}
	if r.showCommanderPanel {
		r.addUILayer(uiLayerCommander, commanderPanelRect())
	}
	if r.showShortcuts {
		r.addUIScreenLayer(uiLayerShortcuts)
	}
}

func editInspectorLayerRect() gameui.Rect {
	x, y, w, h := editInspectorRect()
	return floatRect(x, y, w, h)
}

func recruitPanelLayerRect(gs *state.GameState, rid world.RegionID) gameui.Rect {
	if !RecruitPanelVisible(gs, rid) {
		return gameui.Rect{}
	}
	slots := recruitPanelSlots()
	px := float64(recruitPanelX(slots))
	metrics := recruitPanelMetricsFor(gs, rid)
	py := float64(recruitPanelYForMetrics(metrics))
	return gameui.Rect{X: px, Y: py, W: float64(recruitPanelW(slots)), H: float64(metrics.panelH)}
}

func (r *Renderer) grainEconomyPopupHoveringAtCursor() bool {
	if r == nil {
		return false
	}
	mx, my := ebiten.CursorPosition()
	return r.grainEconomyPopupHovering(float64(mx), float64(my))
}

func (r *Renderer) goldIncomePopupHoveringAtCursor() bool {
	if r == nil {
		return false
	}
	mx, my := ebiten.CursorPosition()
	return r.goldIncomePopupHovering(float64(mx), float64(my))
}
