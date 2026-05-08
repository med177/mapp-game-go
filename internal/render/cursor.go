package render

import (
	"math"

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
	if r.showDiplomacy {
		if r.diplomaPanelHovering(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}
	if r.showTech {
		if r.techPanelHovering(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}

	switch r.gs.Phase {
	case state.PhaseMainMenu:
		if r.mainMenuHoverIndex(fx, fy) >= 0 {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhaseScenarioSelect:
		if uiRectHit(fx, fy, backButtonRect()) || r.scenarioHoverIndex(fx, fy) >= 0 {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhaseFactionSelect:
		if uiRectHit(fx, fy, backButtonRect()) || r.factionCardHoverIndex(fx, fy) >= 0 {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhaseVictorySelect:
		if uiRectHit(fx, fy, backButtonRect()) || r.victoryCardHoverIndex(fx, fy) >= 0 {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhasePlayerTurn:
		if r.inGameHovering(fx, fy) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhasePauseMenu:
		if r.pauseMenuHoverIndex(fx, fy) >= 0 {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			return
		}
	case state.PhaseLoadSelect, state.PhaseSaveSelect:
		if uiRectHit(fx, fy, backButtonRect()) || r.slotHoverIndex(fx, fy) >= 0 {
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

// --- Hit-test yardımcıları ---

func (r *Renderer) mainMenuHoverIndex(fx, fy float64) int {
	items := buildMenuItems(r.HasSave)
	itemH := 52.0
	startY := ScreenHeight/2 - float64(len(items))*itemH/2 + 20
	barW := 280.0
	barX := ScreenWidth/2 - barW/2
	for i, item := range items {
		if item.disabled {
			continue
		}
		y := startY + float64(i)*itemH
		if fx >= barX && fx <= barX+barW && fy >= y-6 && fy <= y+itemH-8 {
			return i
		}
	}
	return -1
}

func (r *Renderer) factionCardHoverIndex(fx, fy float64) int {
	factions := selectableFactions(r.gs)
	cols := 3
	rows := (len(factions) + cols - 1) / cols
	cardW, cardH := 350.0, 110.0
	padX, padY := 30.0, 12.0
	gridW := cardW*float64(cols) + padX*float64(cols-1)
	gridH := cardH*float64(rows) + padY*float64(rows-1)
	headerH := 70.0
	startX := ScreenWidth/2 - gridW/2
	startY := ScreenHeight/2 - (gridH+headerH)/2 + headerH

	for i := range factions {
		col := i % cols
		row := i / cols
		x := startX + float64(col)*(cardW+padX)
		y := startY + float64(row)*(cardH+padY)
		if fx >= x && fx <= x+cardW && fy >= y && fy <= y+cardH {
			return i
		}
	}
	return -1
}

func (r *Renderer) victoryCardHoverIndex(fx, fy float64) int {
	opts := r.gs.AvailableVictories
	cardW, cardH := 520.0, 100.0
	gap := 12.0
	n := float64(len(opts))
	totalH := n*cardH + (n-1)*gap
	headerH := 80.0
	startY := (ScreenHeight-(totalH+headerH))/2 + headerH
	cx := ScreenWidth/2 - cardW/2
	for i := range opts {
		y := startY + float64(i)*(cardH+gap)
		if fx >= cx && fx <= cx+cardW && fy >= y && fy <= y+cardH {
			return i
		}
	}
	return -1
}

func (r *Renderer) diplomaPanelHovering(fx, fy float64) bool {
	if diplomacyCloseHit(fx, fy) {
		return true
	}
	factions := sortedFactions(r.gs)
	for i := range factions {
		y := diplomStartY + float64(i)*diplomRowH
		if fy >= y && fy <= y+diplomRowH-4 && fx >= 28 && fx <= ScreenWidth-56 {
			return true
		}
		if i == r.diplomacyFocus {
			for j := range diplomActions {
				bx, by, bw, bh := diplomActionRect(y, j)
				if fx >= float64(bx) && fx <= float64(bx+bw) && fy >= float64(by) && fy <= float64(by+bh) {
					return true
				}
			}
		}
	}
	return false
}

func (r *Renderer) techPanelHovering(fx, fy float64) bool {
	px, py := float32(60), float32(40)
	pw := float32(ScreenWidth - 120)
	ph := float32(ScreenHeight - 80)
	return fx >= float64(px) && fx <= float64(px+pw) && fy >= float64(py) && fy <= float64(py+ph)
}

func (r *Renderer) inGameHovering(fx, fy float64) bool {
	for _, rect := range BottomButtonRects() {
		if fx >= float64(rect[0]) && fx <= float64(rect[0]+rect[2]) &&
			fy >= float64(rect[1]) && fy <= float64(rect[1]+rect[3]) {
			return true
		}
	}
	if r.SelectedRegion != "" {
		if regionPanelHit(fx, fy) || RecruitPanelBoundsHit(fx, fy, r.gs, r.SelectedRegion) {
			return true
		}
	}
	if r.SelectedArmy != "" && armyPanelCloseHit(fx, fy) {
		return true
	}
	// Ordu ikonu üzerinde mi?
	for _, pos := range r.armyIconPositions() {
		dx := fx - float64(pos.X)
		dy := fy - float64(pos.Y)
		if math.Sqrt(dx*dx+dy*dy) < 14 {
			return true
		}
	}
	// BÖLDÜR / BİRLEŞTİR butonları
	if r.selectedArmyIsPlayerOwned() && SplitButtonHitTest(fx, fy, r.gs, r.SelectedArmy) {
		return true
	}
	if r.selectedArmyIsPlayerOwned() && MergeButtonHitTest(fx, fy, r.gs, r.SelectedArmy) {
		return true
	}
	// Sağ minimap / event log alanı üzerinde mi?
	if fx > float64(evLogX()) {
		return true
	}
	return false
}
