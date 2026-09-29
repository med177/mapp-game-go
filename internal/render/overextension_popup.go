package render

import (
	"strings"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	eventAlertPopupW = 500.0
	eventAlertPopupH = 118.0
)

// overextensionHUDHovering, üst uyarı katmanındaki ilk bilgi öğesinin ortak
// hover kontrolüdür. Yeni uyarılar aynı katmana eklendiğinde bu kontrol,
// öğeye özel rect helper'larına ayrılabilir.
func (r *Renderer) overextensionHUDHovering(fx, fy float64) bool {
	return r != nil && r.gs != nil && r.gs.PlayerFactionID != "" && topAlertHudRect().Hit(fx, fy)
}

func (r *Renderer) nearestEventHUDHovering(fx, fy float64) bool {
	if r == nil || r.gs == nil || r.gs.PlayerFactionID == "" {
		return false
	}
	_, ok := r.nearestEventCodexEntry()
	return ok && topAlertHudRect().Hit(fx, fy)
}

func eventAlertPopupRect(mx, my float64) gameui.Rect {
	x, y, w, h := tooltipRect(mx, my, eventAlertPopupW, eventAlertPopupH)
	return gameui.Rect{X: x, Y: y, W: w, H: h}
}

func (r *Renderer) eventAlertPopupHoveringAtCursor() bool {
	if r == nil {
		return false
	}
	mx, my := ebiten.CursorPosition()
	return r.nearestEventHUDHovering(float64(mx), float64(my))
}

func (r *Renderer) drawNearestEventPopup(screen *ebiten.Image) {
	if r == nil || !r.eventAlertPopupHoveringAtCursor() {
		return
	}

	event, ok := r.nearestEventCodexEntry()
	if !ok {
		return
	}
	mx, my := ebiten.CursorPosition()
	popup := eventAlertPopupRect(float64(mx), float64(my))
	drawTooltipBox(screen, popup.X, popup.Y, popup.W, popup.H)

	lines := eventAlertPopupLines(event)
	DrawText(screen, trimTextToWidth(lines[0], FaceMed, eventAlertPopupW-20), popup.X+10, popup.Y+10, FaceMed, ColorGold)
	DrawText(screen, trimTextToWidth(lines[1], FaceSmall, eventAlertPopupW-20), popup.X+10, popup.Y+34, FaceSmall, ColorGray)
	DrawText(screen, trimTextToWidth(lines[2], FaceSmall, eventAlertPopupW-20), popup.X+10, popup.Y+54, FaceSmall, ColorWhite)
	DrawText(screen, trimTextToWidth(lines[3], FaceSmall, eventAlertPopupW-20), popup.X+10, popup.Y+74, FaceSmall, ColorWhite)
	if event.Summary != "" {
		DrawText(screen, trimTextToWidth(event.Summary, FaceTiny, eventAlertPopupW-20), popup.X+10, popup.Y+96, FaceTiny, ColorGray)
	}
}

func eventAlertPopupLines(event EventCodexEntry) [4]string {
	date := event.DateLabel
	if date == "" {
		date = "Tarih belirtilmedi"
	}
	condition := event.ConditionSummary
	if condition == "" {
		condition = event.Status
	}
	if condition == "" {
		condition = "Belirtilmedi"
	}
	effect := event.EffectSummary
	if effect == "" {
		effect = "Belirtilmedi"
	}
	effect = strings.TrimPrefix(effect, "Etki: ")
	return [4]string{
		event.Title,
		"Tarih: " + date + " • Kalan: " + itoa(event.TurnsUntil) + " tur",
		"Şartlar: " + condition,
		"Getiriler: " + effect,
	}
}
