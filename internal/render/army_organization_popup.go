package render

import (
	"fmt"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

func armyOrganizationPopupRect() gameui.Rect {
	const (
		popupW = 270.0
		popupH = 76.0
	)
	x := 908.0 + 130.0 - popupW
	if x < 8 {
		x = 8
	}
	y := float64(topStatusH) + 8
	if y+popupH > ScreenHeight-8 {
		y = float64(topStatusH) - popupH - 8
	}
	return gameui.Rect{X: x, Y: y, W: popupW, H: popupH}
}

func (r *Renderer) armyOrganizationPopupHovering(fx, fy float64) bool {
	return r != nil && r.gs != nil && armyOrganizationHUDValueRect(r.gs).Hit(fx, fy)
}

func (r *Renderer) armyOrganizationPopupHoveringAtCursor() bool {
	if r == nil {
		return false
	}
	mx, my := ebiten.CursorPosition()
	return r.armyOrganizationPopupHovering(float64(mx), float64(my))
}

func (r *Renderer) drawArmyOrganizationPopup(screen *ebiten.Image) {
	if r == nil || r.gs == nil || !r.armyOrganizationPopupHoveringAtCursor() {
		return
	}

	current := r.gs.CurrentLandArmies(r.gs.PlayerFactionID)
	maximum := r.gs.MaxLandArmies(r.gs.PlayerFactionID)
	penalty := r.gs.ArmyOrganizationPenaltyPercent(r.gs.PlayerFactionID)
	popup := armyOrganizationPopupRect()
	drawUIPanelFrame(screen, popup, panelBg, panelBorder, 1.5, 5)
	drawUILabel(screen, gameui.Rect{X: popup.X + 12, Y: popup.Y + 10, W: popup.W - 24}, "Ordu organizasyonu", ColorGold, gameui.TextSmall, gameui.TextAlignStart)
	drawUIKeyValueRowWithGap(screen, popup.X+12, popup.Y+36, popup.W-24,
		"Ordu sınırı", fmt.Sprintf("%d/%d", current, maximum), ColorGray, ColorWhite, 8)
	if penalty > 0 {
		drawUILabel(screen, gameui.Rect{X: popup.X + 12, Y: popup.Y + 56, W: popup.W - 24}, fmt.Sprintf("Organizasyon: -%d%%", penalty), ColorRed, gameui.TextSmall, gameui.TextAlignStart)
	} else {
		drawUILabel(screen, gameui.Rect{X: popup.X + 12, Y: popup.Y + 56, W: popup.W - 24}, "Organizasyon Cezası yok", ColorGreen, gameui.TextSmall, gameui.TextAlignStart)
	}
}
