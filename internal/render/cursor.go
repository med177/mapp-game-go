package render

import (
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/state"

	"github.com/hajimehoshi/ebiten/v2"
)

// updateCursorShape her frame fare konumuna göre OS imlecini günceller.
func (r *Renderer) updateCursorShape() {
	mx, my := ebiten.CursorPosition()
	ebiten.SetCursorShape(r.cursorShapeAt(float64(mx), float64(my)))
}

// cursorShapeAt bütün ekran ve oyun fazları için tek cursor karar yoludur.
// UI katmanı önce değerlendirilir; harita üzerindeki faza özel hedefler ancak
// cursor bir UI yüzeyinin üzerinde değilse kontrol edilir.
func (r *Renderer) cursorShapeAt(fx, fy float64) ebiten.CursorShapeType {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) && r.uiLayers.BlocksAt(fx, fy) {
		return ebiten.CursorShapeDefault
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		return ebiten.CursorShapeMove
	}
	if r.navalMissionTargeting {
		if r.navalMissionTargetHovering(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}

	// Açık paneller öncelikli kontrol
	if r.showHistoricalEvent {
		if r.historicalEventHovering(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.battleReport.show {
		if r.battleReportHovering(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.combatSummary.show {
		if r.combatSummaryHovering(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.showVictoryDetail {
		if victoryDetailCloseHit(fx, fy) || !victoryDetailPopupHit(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.regionTaskDialog.show {
		if r.regionTaskDialogHovering(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.confirmDialog.show {
		if r.confirmDialogHovering(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.warConfirm.show {
		if r.warConfirmHovering(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.warSummary.show {
		if r.warSummaryHovering(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.battlePlan.show {
		if r.battlePlanHovering(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.negotiation.show {
		if r.diplomacyNegotiationPointerHit(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if _, ok := r.playerDiplomacyOfferIndex(); ok {
		if r.diplomacyOfferHovering(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.showEventCodex {
		if eventCodexCloseHit(fx, fy) || !eventCodexPopupHit(fx, fy) {
			return ebiten.CursorShapePointer
		}
		for _, btn := range buildEventCodexFilterButtons() {
			if btn.HitTest(fx, fy) {
				return ebiten.CursorShapePointer
			}
		}
		if eventCodexEntryHit(fx, fy, len(r.currentEventCodexEntries()), r.eventCodexScroll) >= 0 {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.eventDetail != "" {
		if eventDetailCloseHit(fx, fy) || !eventDetailPopupHit(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.showImperialPanel {
		if r.imperialPanelPointerHit(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if pointer, handled := r.overlayPanelCursorHit(fx, fy); handled {
		if pointer {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.grainEconomyPopupHovering(fx, fy) {
		return ebiten.CursorShapePointer
	}
	if r.goldIncomePopupHovering(fx, fy) {
		return ebiten.CursorShapePointer
	}
	if r.overextensionHUDHovering(fx, fy) {
		return ebiten.CursorShapePointer
	}
	if r.nearestEventHUDHovering(fx, fy) {
		return ebiten.CursorShapePointer
	}
	if r.showDiplomacy {
		if r.diplomacyPanelPointerHit(fx, fy, r.diplomacyFocus, r.diplomacyScroll, r.diplomacyTargetFaction, r.diplomacyHistoryDirectionFilter, r.diplomacyHistoryActionFilter) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.showTech {
		if r.techPanelPointerHit(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.showTrade {
		if tradePanelPointerHit(fx, fy, r.gs, r.tradeTab, r.tradeFactionFocus, r.tradeGoodFocus, r.tradeScroll, r.tradeAmount, r.tradeListFilter, r.tradeListSort) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.showCommanderPanel {
		if r.commanderPanelHovering(fx, fy) {
			return ebiten.CursorShapePointer
		}
		return ebiten.CursorShapeDefault
	}
	if r.selectedSiegePanelHovering(fx, fy) {
		return ebiten.CursorShapePointer
	}

	switch r.gs.Phase {
	case state.PhaseMainMenu:
		if r.mainMenuHoverIndex(fx, fy) >= 0 {
			return ebiten.CursorShapePointer
		}
	case state.PhaseScenarioSelect:
		if buildBackButton().HitTest(fx, fy) || r.scenarioHoverIndex(fx, fy) >= 0 {
			return ebiten.CursorShapePointer
		}
	case state.PhaseFactionSelect:
		if buildBackButton().HitTest(fx, fy) || r.factionCardHoverIndex(fx, fy) >= 0 {
			return ebiten.CursorShapePointer
		}
	case state.PhaseVictorySelect:
		if buildBackButton().HitTest(fx, fy) || r.victoryCardHoverIndex(fx, fy) >= 0 {
			return ebiten.CursorShapePointer
		}
	case state.PhasePlayerTurn, state.PhaseEditMode:
		if r.gs.Phase == state.PhaseEditMode && r.editBuildingsPanel {
			if r.editBuildingPanelInteractiveHit(fx, fy) {
				return ebiten.CursorShapePointer
			}
			return ebiten.CursorShapeDefault
		}
		if r.gs.Phase == state.PhaseEditMode && r.editNewShapeModal.show {
			if editNewShapeModalHit(fx, fy) {
				return ebiten.CursorShapePointer
			}
			return ebiten.CursorShapeDefault
		}
		if r.gs.Phase == state.PhaseEditMode && r.editLandPassageForm.show {
			if landPassageFormRect().Hit(fx, fy) {
				return ebiten.CursorShapePointer
			}
			return ebiten.CursorShapeDefault
		}
		if r.gs.Phase == state.PhaseEditMode && r.editRegionForm.show {
			if r.editRegionFormInteractiveHit(fx, fy) {
				return ebiten.CursorShapePointer
			}
			return ebiten.CursorShapeDefault
		}
		if r.gs.Phase == state.PhaseEditMode && r.editFactionForm.show {
			if editFactionFormHit(fx, fy) {
				return ebiten.CursorShapePointer
			}
			return ebiten.CursorShapeDefault
		}
		if r.landPassageHoverAt(fx, fy) >= 0 {
			return ebiten.CursorShapePointer
		}
		if r.uiLayers.BlocksAt(fx, fy) {
			if r.uiLayerPointerAt(fx, fy) {
				return ebiten.CursorShapePointer
			}
			return ebiten.CursorShapeDefault
		}
		if r.gs.Phase == state.PhasePlayerTurn {
			if _, ok := r.navalSupplyCargoHitAt(fx, fy); ok ||
				r.navalMovementTargetHovering(fx, fy) ||
				r.embarkFleetTargetHovering(fx, fy) ||
				r.armyMovementTargetHovering(fx, fy) ||
				r.currentRegionArmyTaskHovering(fx, fy) ||
				r.inGameHovering(fx, fy) {
				return ebiten.CursorShapePointer
			}
			return ebiten.CursorShapeDefault
		}
		if r.editShapeHelpPanelHit(fx, fy) {
			return ebiten.CursorShapeDefault
		}
		if _, ok := r.editRegionCenterAt(fx, fy); ok {
			return ebiten.CursorShapePointer
		}
		if editModifierPressed() && r.editRegionAt(fx, fy) != "" {
			return ebiten.CursorShapePointer
		}
		if editAddModifierPressed() && r.editRegionAt(fx, fy) != "" {
			return ebiten.CursorShapePointer
		}
		if _, _, ok := r.editSettlementAt(fx, fy); ok {
			return ebiten.CursorShapePointer
		}
		if _, ok := r.editArmyAt(fx, fy); ok {
			return ebiten.CursorShapePointer
		}
	case state.PhasePauseMenu:
		if r.pauseMenuHoverIndex(fx, fy) >= 0 {
			return ebiten.CursorShapePointer
		}
	case state.PhaseLoadSelect:
		if r.slotSelectHovering(fx, fy, false) {
			return ebiten.CursorShapePointer
		}
	case state.PhaseSaveSelect:
		if r.slotSelectHovering(fx, fy, true) {
			return ebiten.CursorShapePointer
		}
	case state.PhaseSettings:
		if r.settingsHoverIndex(fx, fy) >= 0 {
			return ebiten.CursorShapePointer
		}
	}

	return ebiten.CursorShapeDefault
}

func (r *Renderer) currentRegionArmyTaskHovering(fx, fy float64) bool {
	if r == nil || r.gs == nil || r.worldMap == nil || r.SelectedArmy == "" {
		return false
	}
	attacker := r.gs.Armies[r.SelectedArmy]
	if attacker == nil || attacker.OwnerID != string(r.gs.PlayerFactionID) || attacker.IsNaval {
		return false
	}
	wx, wy := r.screenToWorld(fx, fy)
	if r.worldMap.RegionAt(int(wx), int(wy)) != attacker.RegionID {
		return false
	}
	return r.currentRegionArmyTaskIndicatorVisible(attacker, r.gs.Regions[attacker.RegionID])
}

// overlayPanelCursorHit cursor önceliğini input ve draw stack'iyle aynı sırada
// uygular. Üstteki panel kendi yüzeyindeyse alttaki panelin hover'ı görünmez.
func (r *Renderer) overlayPanelCursorHit(fx, fy float64) (pointer, handled bool) {
	if r == nil {
		return false, false
	}
	r.ensureOverlayPanelOrder()
	top, hasTop := r.uiLayers.TopAt(fx, fy)
	for i := r.overlayPanelOrderLen - 1; i >= 0; i-- {
		panel := r.overlayPanelOrder[i]
		if !r.overlayPanelVisible(panel) {
			continue
		}
		if hasTop && top.ID != overlayPanelLayerID(panel) {
			continue
		}
		switch panel {
		case overlayPanelMerchantRoute:
			return merchantRoutePanelInteractiveHit(r, fx, fy), true
		case overlayPanelNavalMission:
			return r.navalMissionPanelInteractiveHit(fx, fy), true
		case overlayPanelActiveWars:
			if activeWarsPanelHit(fx, fy) {
				return activeWarsPanelInteractiveHit(r, fx, fy), true
			}
		}
	}
	return false, false
}

func (r *Renderer) historicalEventHovering(fx, fy float64) bool {
	if !r.showHistoricalEvent {
		return false
	}
	if len(r.historicalEventChoices) == 0 {
		if len(r.commanderArrivals) > 0 {
			return buildCommanderArrivalModal().Panel.Rect.Hit(fx, fy)
		}
		return historicalEventPopupHit(fx, fy, r.historicalEventTitle, r.historicalEventDesc, r.historicalEventPrompt, r.historicalEventChoices)
	}
	for _, btn := range buildHistoricalEventChoiceButtons(r.historicalEventTitle, r.historicalEventDesc, r.historicalEventPrompt, r.historicalEventChoices) {
		if btn.HitTest(fx, fy) {
			return true
		}
	}
	return false
}

func historicalEventPopupHit(fx, fy float64, title, desc, prompt string, choices []HistoricalEventChoice) bool {
	modal := buildHistoricalEventModal(title, desc, prompt, choices)
	return modal.Panel.Rect.Hit(fx, fy)
}

// --- Hit-test yardımcıları ---

func (r *Renderer) mainMenuHoverIndex(fx, fy float64) int {
	items := buildMenuItems(r.HasSave, r.HasAutoSave, r.EditModeEnabled)
	for i, btn := range buildMainMenuButtons(r.HasSave, r.HasAutoSave, r.EditModeEnabled) {
		if items[i].disabled {
			continue
		}
		if btn.HitTest(fx, fy) {
			return i
		}
	}
	return -1
}

func (r *Renderer) factionCardHoverIndex(fx, fy float64) int {
	factions, historicalCount := selectableFactions(r.gs)
	viewport := factionGroupLayoutScrolled(len(factions), historicalCount, 3, 350, 138, 30, 12, 70, r.factionSelectScroll).viewport
	if !viewport.Hit(fx, fy) {
		return -1
	}
	for i, btn := range buildFactionCardButtons(r.gs, r.factionSelectScroll) {
		if btn.HitTest(fx, fy) {
			return i
		}
	}
	return -1
}

func (r *Renderer) victoryCardHoverIndex(fx, fy float64) int {
	opts, historicalCount := orderedVictoryOptions(r.gs)
	cardW, cardH := victoryCardDimensions()
	viewport := victoryLayoutScrolled(len(opts), historicalCount, cardW, cardH, 12, 80, r.victorySelectScroll).viewport
	if !viewport.Hit(fx, fy) {
		return -1
	}
	for i, btn := range buildVictoryCardButtons(r.gs, r.victorySelectScroll) {
		if btn.HitTest(fx, fy) {
			return i
		}
	}
	return -1
}

func (r *Renderer) confirmDialogHovering(fx, fy float64) bool {
	acceptBtn, thirdBtn, declineBtn, hasThird := buildConfirmDialogButtons(r.confirmDialog)
	if acceptBtn.HitTest(fx, fy) {
		return true
	}
	if hasThird && thirdBtn.Enabled && thirdBtn.HitTest(fx, fy) {
		return true
	}
	return declineBtn.Enabled && declineBtn.HitTest(fx, fy)
}

func (r *Renderer) warConfirmHovering(fx, fy float64) bool {
	acceptBtn, declineBtn := buildWarConfirmButtons()
	if acceptBtn.HitTest(fx, fy) || declineBtn.HitTest(fx, fy) {
		return true
	}
	leftRect, rightRect := warConfirmSideRects(buildWarConfirmModal())
	leftAutoViewport := warConfirmAutoViewport(leftRect)
	leftCallViewport := warConfirmCallViewport(leftRect)
	rightAutoViewport := warConfirmAutoViewport(rightRect)
	rightCallViewport := warConfirmCallViewport(rightRect)
	if leftAutoViewport.Hit(fx, fy) || leftCallViewport.Hit(fx, fy) || rightAutoViewport.Hit(fx, fy) || rightCallViewport.Hit(fx, fy) {
		return true
	}
	for _, checkbox := range warConfirmCheckboxes(leftCallViewport, r.warConfirm.preview.Attacker.CallableAllies, r.warConfirm.selectedAllies, r.warConfirm.attackerCallScroll) {
		if checkbox.HitTest(fx, fy) {
			return true
		}
	}
	return false
}

func (r *Renderer) warSummaryHovering(fx, fy float64) bool {
	if warSummaryCloseHit(fx, fy) {
		return true
	}
	layout := buildWarSummaryLayout()
	return layout.attackerListRect.Hit(fx, fy) || layout.defenderListRect.Hit(fx, fy)
}

func (r *Renderer) battlePlanHovering(fx, fy float64) bool {
	buttons, cancelBtn := buildBattlePlanButtons()
	if cancelBtn.HitTest(fx, fy) {
		return true
	}
	for _, btn := range buttons {
		if btn.HitTest(fx, fy) {
			return true
		}
	}
	return false
}

func (r *Renderer) battleReportHovering(fx, fy float64) bool {
	return battleReportCloseHit(fx, fy) || battleReportContinueHit(fx, fy) || !battleReportPopupHit(fx, fy)
}

func (r *Renderer) diplomacyOfferHovering(fx, fy float64) bool {
	offerIdx, ok := r.playerDiplomacyOfferIndex()
	if !ok {
		return false
	}
	offer := r.gs.DiplomaticOffers[offerIdx]
	if diplomacyOfferIsNotification(offer) {
		return buildDiplomacyOfferNoticeButton().HitTest(fx, fy)
	}
	acceptBtn, rejectBtn := buildDiplomacyOfferButtons()
	if acceptBtn.HitTest(fx, fy) || rejectBtn.HitTest(fx, fy) {
		return true
	}
	return offer.Action == string(diplomacy.ActionProposeTransfer) && buildDiplomacyOfferCounterButton().HitTest(fx, fy)
}

func (r *Renderer) inGameHovering(fx, fy float64) bool {
	if topDateHudMenuButtonHit(fx, fy) || bottomActionButtonHit(fx, fy) || (imperialPanelAvailable(r.gs) && imperialHUDButtonHit(fx, fy)) || musicHudInteractiveHit(fx, fy) || activeWarsHudButtonHit(fx, fy) || turnTechHudTechHit(fx, fy) || turnTechHudTradeRouteHit(r.gs, fx, fy) || turnTechHudWarFatigueHit(r.gs, fx, fy) || r.armyOrganizationPopupHovering(fx, fy) {
		return true
	}
	if victoryProgressHit(fx, fy) {
		return true
	}
	if eventLogInteractiveHit(fx, fy, len(r.eventLog), r.eventLogCollapsed, r.eventLogScroll, r.HasEventCodex()) {
		return true
	}
	if r.SelectedRegion != "" {
		if logisticsRect, ok := regionPanelLogisticsRect(r.gs, r.SelectedRegion); ok && logisticsRect.Hit(fx, fy) {
			return true
		}
		if regionPanelInteractiveHitForTab(fx, fy, r.gs, r.SelectedRegion, r.regionPanelTab, r.regionPanelScroll) ||
			r.settlementPanelHit(fx, fy) || r.settlementPanelCloseHit(fx, fy) ||
			(r.mapMode != MapModeTrade && r.showRecruitPanel && RecruitPanelInteractiveHit(fx, fy, r.gs, r.SelectedRegion)) {
			return true
		}
	}
	if r.showArmyDetailPanel && r.SelectedArmy != "" && ArmyPanelBoundsHit(fx, fy, r.gs, r.SelectedArmy) {
		if r.SelectedEmbarkedArmyFleet == r.SelectedArmy {
			return buildArmyPanelCloseButton().HitTest(fx, fy)
		}
		return ArmyPanelInteractiveHit(fx, fy, r.gs, r.SelectedArmy, r.splitSelectedUnits)
	}
	if r.selectedSiegePanelHit(fx, fy) {
		return true
	}
	if r.mapMode == MapModeTrade && (r.tradeCorridorAt(fx, fy) >= 0 || r.tradeCenterAt(fx, fy) >= 0) {
		return true
	}
	// Ordu/donanma etiketi üzerinde mi?
	if _, ok := r.navalMissionPendingHitAt(fx, fy); ok {
		return true
	}
	if _, ok := r.navalMissionBonusHitAt(fx, fy); ok {
		return true
	}
	if _, ok := r.merchantTradeBonusHitAt(fx, fy); ok {
		return true
	}
	if _, ok := r.embarkedArmyHitAt(fx, fy); ok {
		return true
	}
	if _, ok := r.armyHitAt(fx, fy); ok {
		return true
	}
	// Yerleşim noktası üzerinde mi?
	if _, _, ok := r.settlementHitAt(fx, fy); ok {
		return true
	}
	return false
}
