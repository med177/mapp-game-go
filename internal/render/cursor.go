package render

import (
	"mapp-game-go/internal/state"

	"github.com/hajimehoshi/ebiten/v2"
)

// updateCursorShape her frame fare konumuna göre OS imlecini günceller.
func (r *Renderer) updateCursorShape() {
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		ebiten.SetCursorShape(ebiten.CursorShapeMove)
		return
	}

	// Açık paneller öncelikli kontrol
	if r.showHistoricalEvent {
		if r.historicalEventHovering(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}
	if r.showVictoryDetail {
		if victoryDetailCloseHit(fx, fy) || !victoryDetailPopupHit(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}
	if r.confirmDialog.show {
		if r.confirmDialogHovering(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}
	if r.warConfirm.show {
		if r.warConfirmHovering(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}
	if r.battlePlan.show {
		if r.battlePlanHovering(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}
	if _, ok := r.playerDiplomacyOfferIndex(); ok {
		if r.diplomacyOfferHovering(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}
	if r.showEventCodex {
		if eventCodexCloseHit(fx, fy) || !eventCodexPopupHit(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		for _, btn := range buildEventCodexFilterButtons() {
			if btn.HitTest(fx, fy) {
				ebiten.SetCursorShape(ebiten.CursorShapePointer)
				return
			}
		}
		if eventCodexEntryHit(fx, fy, len(r.currentEventCodexEntries()), r.eventCodexScroll) >= 0 {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}
	if r.eventDetail != "" {
		if eventDetailCloseHit(fx, fy) || !eventDetailPopupHit(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}
	if r.showDiplomacy {
		if diplomacyPanelPointerHit(fx, fy, r.gs, r.diplomacyFocus, r.diplomacyScroll, r.diplomacyTargetFaction) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}
	if r.showTech {
		if r.techPanelPointerHit(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}
	if r.showTrade {
		if tradePanelPointerHit(fx, fy, r.gs, r.tradeTab, r.tradeFactionFocus, r.tradeGoodFocus, r.tradeScroll, r.tradeListFilter, r.tradeListSort) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}
	if r.selectedSiegePanelHovering(fx, fy) {
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
		return
	}

	switch r.gs.Phase {
	case state.PhaseMainMenu:
		if r.mainMenuHoverIndex(fx, fy) >= 0 {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhaseScenarioSelect:
		if buildBackButton().HitTest(fx, fy) || r.scenarioHoverIndex(fx, fy) >= 0 {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhaseFactionSelect:
		if buildBackButton().HitTest(fx, fy) || r.factionCardHoverIndex(fx, fy) >= 0 {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhaseVictorySelect:
		if buildBackButton().HitTest(fx, fy) || r.victoryCardHoverIndex(fx, fy) >= 0 {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhasePlayerTurn:
		if r.inGameHovering(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhaseEditMode:
		if r.editFactionForm.show {
			if editFactionFormHit(fx, fy) {
				ebiten.SetCursorShape(ebiten.CursorShapePointer)
				return
			}
			ebiten.SetCursorShape(ebiten.CursorShapeDefault)
			return
		}
		if r.editOwnerDropdown.IsOpen() && r.editOwnerDropdown.HitTest(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		if r.editTerrainDropdown.IsOpen() && r.editTerrainDropdown.HitTest(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		if r.editSettlementTypeDropdown.IsOpen() && r.editSettlementTypeDropdown.HitTest(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		if r.editInspectorActiveButtonAt(fx, fy) != editButtonNone {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		if r.editShapeHelpPanelHit(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapeDefault)
			return
		}
		if editModifierPressed() && r.editRegionAt(fx, fy) != "" {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		if editAddModifierPressed() && r.editRegionAt(fx, fy) != "" {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		if _, _, ok := r.editSettlementAt(fx, fy); ok {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		if _, ok := r.editArmyAt(fx, fy); ok {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhasePauseMenu:
		if r.pauseMenuHoverIndex(fx, fy) >= 0 {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhaseLoadSelect:
		if r.slotSelectHovering(fx, fy, false) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhaseSaveSelect:
		if r.slotSelectHovering(fx, fy, true) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhaseSettings:
		if r.settingsHoverIndex(fx, fy) >= 0 {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	}

	ebiten.SetCursorShape(ebiten.CursorShapeDefault)
}

func (r *Renderer) historicalEventHovering(fx, fy float64) bool {
	if !r.showHistoricalEvent {
		return false
	}
	if len(r.historicalEventChoices) == 0 {
		return historicalEventPopupHit(fx, fy)
	}
	for _, btn := range buildHistoricalEventChoiceButtons(len(r.historicalEventChoices)) {
		if btn.HitTest(fx, fy) {
			return true
		}
	}
	return false
}

func historicalEventPopupHit(fx, fy float64) bool {
	modal := buildHistoricalEventModal()
	return modal.Panel.Rect.Hit(fx, fy)
}

// --- Hit-test yardımcıları ---

func (r *Renderer) mainMenuHoverIndex(fx, fy float64) int {
	items := buildMenuItems(r.HasSave, r.HasAutoSave)
	for i, btn := range buildMainMenuButtons(r.HasSave, r.HasAutoSave) {
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
	for i, btn := range buildFactionCardButtons(r.gs) {
		if btn.HitTest(fx, fy) {
			return i
		}
	}
	return -1
}

func (r *Renderer) victoryCardHoverIndex(fx, fy float64) int {
	for i, btn := range buildVictoryCardButtons(r.gs) {
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
	if hasThird && thirdBtn.HitTest(fx, fy) {
		return true
	}
	return declineBtn.HitTest(fx, fy)
}

func (r *Renderer) warConfirmHovering(fx, fy float64) bool {
	acceptBtn, declineBtn := buildWarConfirmButtons()
	return acceptBtn.HitTest(fx, fy) || declineBtn.HitTest(fx, fy)
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

func (r *Renderer) diplomacyOfferHovering(fx, fy float64) bool {
	acceptBtn, rejectBtn := buildDiplomacyOfferButtons()
	return acceptBtn.HitTest(fx, fy) || rejectBtn.HitTest(fx, fy)
}

func (r *Renderer) inGameHovering(fx, fy float64) bool {
	if topDateHudMenuButtonHit(fx, fy) || bottomActionButtonHit(fx, fy) || musicHudInteractiveHit(fx, fy) {
		return true
	}
	if victoryProgressHit(fx, fy) {
		return true
	}
	if eventLogInteractiveHit(fx, fy, len(r.eventLog), r.eventLogCollapsed, r.eventLogScroll, r.HasEventCodex()) {
		return true
	}
	if r.SelectedRegion != "" {
		if regionPanelInteractiveHit(fx, fy, r.gs, r.SelectedRegion) ||
			r.settlementPanelHit(fx, fy) || r.settlementPanelCloseHit(fx, fy) ||
			RecruitPanelInteractiveHit(fx, fy, r.gs, r.SelectedRegion) {
			return true
		}
	}
	if r.SelectedArmy != "" && ArmyPanelInteractiveHit(fx, fy, r.gs, r.SelectedArmy) {
		return true
	}
	if r.selectedSiegePanelHit(fx, fy) {
		return true
	}
	if r.mapMode == MapModeTrade && (r.tradeCorridorAt(fx, fy) >= 0 || r.tradeCenterAt(fx, fy) >= 0) {
		return true
	}
	// Ordu/donanma etiketi üzerinde mi?
	if _, ok := r.armyHitAt(fx, fy); ok {
		return true
	}
	// Yerleşim noktası üzerinde mi?
	if _, _, ok := r.settlementHitAt(fx, fy); ok {
		return true
	}
	return false
}
