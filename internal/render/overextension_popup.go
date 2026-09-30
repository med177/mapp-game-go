package render

import (
	"image/color"
	"strings"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	overextensionPopupW = 410.0
	overextensionPopupH = 184.0
	eventAlertPopupW    = 500.0
	eventAlertPopupH    = 118.0
)

// overextensionHUDHovering, üst uyarı katmanındaki ilk bilgi öğesinin ortak
// hover kontrolüdür. Yeni uyarılar aynı katmana eklendiğinde bu kontrol,
// öğeye özel rect helper'larına ayrılabilir.
func (r *Renderer) overextensionHUDHovering(fx, fy float64) bool {
	return r != nil && r.gs != nil && r.gs.PlayerFactionID != "" && topAlertOverextensionRect(r.gs).Hit(fx, fy)
}

func (r *Renderer) nearestEventHUDHovering(fx, fy float64) bool {
	if r == nil || r.gs == nil || r.gs.PlayerFactionID == "" {
		return false
	}
	_, ok := r.nearestEventCodexEntry()
	return ok && topAlertEventRect(r.gs).Hit(fx, fy)
}

func overextensionPopupRect(mx, my float64) gameui.Rect {
	x, y, w, h := tooltipRect(mx, my, overextensionPopupW, overextensionPopupH)
	return gameui.Rect{X: x, Y: y, W: w, H: h}
}

func (r *Renderer) overextensionPopupHoveringAtCursor() bool {
	if r == nil {
		return false
	}
	mx, my := ebiten.CursorPosition()
	return r.overextensionHUDHovering(float64(mx), float64(my))
}

func (r *Renderer) drawOverextensionPopup(screen *ebiten.Image) {
	if r == nil || r.gs == nil || !r.overextensionPopupHoveringAtCursor() {
		return
	}

	mx, my := ebiten.CursorPosition()
	popup := overextensionPopupRect(float64(mx), float64(my))
	drawTooltipBox(screen, popup.X, popup.Y, popup.W, popup.H)

	label, labelColor := overextensionHUDText(r.gs)
	DrawText(screen, "Aşırı Genişleme", popup.X+10, popup.Y+10, FaceMed, ColorGold)
	DrawText(screen, label, popup.X+10, popup.Y+32, FaceSmall, labelColor)
	DrawText(screen, "Son "+itoa(r.gs.AggressiveExpansionWindowTurns())+" turdaki hızlı kara kazanımından", popup.X+10, popup.Y+52, FaceSmall, ColorGray)
	DrawText(screen, "türetilen 0–500 arası risk göstergesidir.", popup.X+10, popup.Y+68, FaceSmall, ColorGray)

	for index, line := range overextensionPopupRangeLines() {
		DrawText(screen, line.text, popup.X+10, popup.Y+92+float64(index)*16, FaceSmall, line.color)
	}
	DrawText(screen, "Gelir veya doğrudan savaş cezası değildir; diplomatik baskıdır.", popup.X+10, popup.Y+164, FaceTiny, ColorGray)
}

type overextensionPopupRangeLine struct {
	text  string
	color color.RGBA
}

func overextensionPopupRangeLines() [4]overextensionPopupRangeLine {
	return [4]overextensionPopupRangeLine{
		{text: "%0–24   Düşük: Genişleme baskısı sınırlı.", color: ColorGreen},
		{text: "%25–99 Orta: Yeni fetihler daha dikkatli değerlendirilir.", color: ColorYellow},
		{text: "%100–199 Yüksek: Yeni savaşlar daha riskli hale gelir.", color: color.RGBA{235, 155, 60, 255}},
		{text: "%200–500 Kritik: Diplomatik baskı ve savaş riski artar.", color: ColorRed},
	}
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
