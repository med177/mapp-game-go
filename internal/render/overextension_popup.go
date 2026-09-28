package render

import (
	"image/color"

	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	overextensionPopupW = 410.0
	overextensionPopupH = 184.0
)

// overextensionHUDHovering, üst uyarı katmanındaki ilk bilgi öğesinin ortak
// hover kontrolüdür. Yeni uyarılar aynı katmana eklendiğinde bu kontrol,
// öğeye özel rect helper'larına ayrılabilir.
func (r *Renderer) overextensionHUDHovering(fx, fy float64) bool {
	return r != nil && r.gs != nil && r.gs.PlayerFactionID != "" && topAlertHudRect().Hit(fx, fy)
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
	DrawText(screen, "Son "+itoa(state.RecentFactionExpansionWindowTurns)+" turdaki hızlı kara kazanımından", popup.X+10, popup.Y+52, FaceSmall, ColorGray)
	DrawText(screen, "türetilen 0–100 arası risk göstergesidir.", popup.X+10, popup.Y+68, FaceSmall, ColorGray)

	for index, line := range overextensionPopupRangeLines() {
		DrawText(screen, line.text, popup.X+10, popup.Y+92+float64(index)*16, FaceSmall, line.color)
	}
	DrawText(screen, "Doğrudan gelir, isyan veya savaş cezası değildir.", popup.X+10, popup.Y+164, FaceTiny, ColorGray)
}

type overextensionPopupRangeLine struct {
	text  string
	color color.RGBA
}

func overextensionPopupRangeLines() [4]overextensionPopupRangeLine {
	return [4]overextensionPopupRangeLine{
		{text: "%0–24   Düşük: Genişleme baskısı sınırlı.", color: ColorGreen},
		{text: "%25–49 Orta: Yeni fetihler daha dikkatli değerlendirilir.", color: ColorYellow},
		{text: "%50–74 Yüksek: Yeni savaşlar daha riskli hale gelir.", color: color.RGBA{235, 155, 60, 255}},
		{text: "%75–100 Kritik: Temkinli AI genişlemeyi güçlü biçimde yavaşlatır.", color: ColorRed},
	}
}
