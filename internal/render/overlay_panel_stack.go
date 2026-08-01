package render

import "github.com/hajimehoshi/ebiten/v2"

type overlayPanelID uint8

const (
	overlayPanelMerchantRoute overlayPanelID = iota
	overlayPanelNavalMission
	overlayPanelActiveWars
)

const overlayPanelCount = int(overlayPanelActiveWars + 1)

func defaultOverlayPanelOrder() [overlayPanelCount]overlayPanelID {
	return [overlayPanelCount]overlayPanelID{
		overlayPanelMerchantRoute,
		overlayPanelNavalMission,
		overlayPanelActiveWars,
	}
}

func (r *Renderer) ensureOverlayPanelOrder() {
	if r == nil || r.overlayPanelOrderLen == overlayPanelCount {
		return
	}
	r.overlayPanelOrder = defaultOverlayPanelOrder()
	r.overlayPanelOrderLen = overlayPanelCount
}

// bringOverlayPanelToFront, panelin tekrar açılması durumunda da onu stack'in
// en üstüne taşır. Kapalı paneller stack'te kalabilir; görünürlük filtrelenir.
func (r *Renderer) bringOverlayPanelToFront(panel overlayPanelID) {
	if r == nil {
		return
	}
	r.ensureOverlayPanelOrder()
	index := -1
	for i := 0; i < r.overlayPanelOrderLen; i++ {
		if r.overlayPanelOrder[i] == panel {
			index = i
			break
		}
	}
	if index < 0 || index == r.overlayPanelOrderLen-1 {
		return
	}
	for i := index; i < r.overlayPanelOrderLen-1; i++ {
		r.overlayPanelOrder[i] = r.overlayPanelOrder[i+1]
	}
	r.overlayPanelOrder[r.overlayPanelOrderLen-1] = panel
}

func (r *Renderer) overlayPanelVisible(panel overlayPanelID) bool {
	if r == nil {
		return false
	}
	switch panel {
	case overlayPanelMerchantRoute:
		return r.showMerchantRoutePanel
	case overlayPanelNavalMission:
		return r.showNavalMissionPanel
	case overlayPanelActiveWars:
		return r.showActiveWars
	default:
		return false
	}
}

// handleOverlayPanelInput input önceliğini en üstteki görünür panelden aşağıya
// doğru uygular. Aktif savaşlar paneli dışındaki tıklamayı haritaya bırakabildiği
// için handled=false dönebilir; alttaki panel varsa o da denenir.
func (r *Renderer) handleOverlayPanelInput() (InputAction, bool) {
	if r == nil {
		return InputAction{}, false
	}
	r.ensureOverlayPanelOrder()
	for i := r.overlayPanelOrderLen - 1; i >= 0; i-- {
		panel := r.overlayPanelOrder[i]
		if !r.overlayPanelVisible(panel) {
			continue
		}
		switch panel {
		case overlayPanelMerchantRoute:
			return r.handleMerchantRoutePanelInput(), true
		case overlayPanelNavalMission:
			return r.handleNavalMissionPanelInput(), true
		case overlayPanelActiveWars:
			if r.handleActiveWarsOverlayInput() {
				return InputAction{}, true
			}
		}
	}
	return InputAction{}, false
}

func (r *Renderer) drawOverlayPanels(screen *ebiten.Image) {
	if r == nil {
		return
	}
	r.ensureOverlayPanelOrder()
	for i := 0; i < r.overlayPanelOrderLen; i++ {
		switch panel := r.overlayPanelOrder[i]; panel {
		case overlayPanelMerchantRoute:
			if r.showMerchantRoutePanel {
				r.drawMerchantRoutePanel(screen)
			}
		case overlayPanelNavalMission:
			if r.showNavalMissionPanel {
				r.drawNavalMissionPanel(screen)
			}
		case overlayPanelActiveWars:
			if r.showActiveWars {
				drawActiveWarsPanel(screen, r.gs, r.activeWarsBuf, r.activeWarsScroll)
			}
		}
	}
}
