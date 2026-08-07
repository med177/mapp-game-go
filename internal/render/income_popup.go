package render

import (
	"image/color"

	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/victory"

	"github.com/hajimehoshi/ebiten/v2"
)

// incomeHUDValueRect, çizim ve cursor hit-test'inin aynı gelir rakamı
// geometri sözleşmesini kullanmasını sağlar.
func incomeHUDValueRect() gameui.Rect {
	_, _, rightCol, _, rightColW := topResourceHUDColumns()
	return gameui.Rect{
		X: rightCol + 42,
		Y: 12 + 22 - 5,
		W: rightColW - 42,
		H: 20,
	}
}

func playerGoldEconomyStatus(gs *state.GameState) state.GoldEconomyStatus {
	if gs == nil {
		return state.GoldEconomyStatus{}
	}
	if status, ok := gs.GoldEconomy[gs.PlayerFactionID]; ok {
		return status
	}
	// İlk ekonomi çözümlemesi gerçekleşmeden önce de HUD net değeri gösterebilir.
	// Ayrıntılı kalemler ilk tur çözümünde aynı snapshot'a dolar.
	gross := victory.CurrentGoldIncome(gs)
	upkeep := gs.FactionGoldUpkeep(gs.PlayerFactionID)
	return state.GoldEconomyStatus{
		FactionID: gs.PlayerFactionID,
		Income:    gross,
		Upkeep:    upkeep,
		NetChange: gross - upkeep,
	}
}

type goldIncomePopupLine struct {
	label string
	value int
	color color.RGBA
}

func goldIncomePopupLines(status state.GoldEconomyStatus) [12]goldIncomePopupLine {
	return [12]goldIncomePopupLine{
		{label: "Vergi", value: status.TaxIncome, color: ColorGold},
		{label: "Pasif ticaret", value: status.TradeIncome, color: color.RGBA{145, 220, 155, 255}},
		{label: "Ticaret rotası geliri", value: status.TradeRouteIncome, color: color.RGBA{145, 220, 155, 255}},
		{label: "Ticaret rotası ödemesi", value: -status.TradeRouteExpense, color: ColorRed},
		{label: "Haraç geliri", value: status.TributeIncome, color: ColorGold},
		{label: "Ödenen haraç", value: -status.TributePaid, color: ColorRed},
		{label: "Teknoloji", value: status.TechnologyIncome, color: color.RGBA{180, 180, 255, 255}},
		{label: "Başkent", value: status.CapitalIncome, color: ColorGold},
		{label: "Abluka ganimeti", value: status.BlockadeIncome, color: ColorGold},
		{label: "Yağma", value: status.RaidIncome, color: ColorGold},
		{label: "Hediyeler", value: status.GiftIncome - status.GiftExpense, color: color.RGBA{210, 170, 120, 255}},
		{label: "Ordu masrafı", value: -status.Upkeep, color: ColorRed},
	}
}

func goldIncomePopupRect() gameui.Rect {
	_, _, rightCol, _, rightColW := topResourceHUDColumns()
	const popupW = 330.0
	const popupH = 340.0
	x := rightCol + rightColW - popupW
	if x < 8 {
		x = 8
	}
	y := float64(topStatusH) + 8
	if y+popupH > ScreenHeight-8 {
		y = float64(topStatusH) - popupH - 8
		if y < 8 {
			y = 8
		}
	}
	return gameui.Rect{X: x, Y: y, W: popupW, H: popupH}
}

func (r *Renderer) goldIncomePopupHovering(fx, fy float64) bool {
	if r == nil || r.gs == nil || r.gs.PlayerFactionID == "" {
		return false
	}
	return incomeHUDValueRect().Hit(fx, fy)
}

func (r *Renderer) drawGoldIncomePopup(screen *ebiten.Image) {
	if r == nil || r.gs == nil {
		return
	}
	mx, my := ebiten.CursorPosition()
	if !r.goldIncomePopupHovering(float64(mx), float64(my)) {
		return
	}

	status := playerGoldEconomyStatus(r.gs)
	popup := goldIncomePopupRect()
	drawUIPanelFrame(screen, popup, panelBg, panelBorder, 1.5, 5)
	drawUILabel(screen, gameui.Rect{X: popup.X + 12, Y: popup.Y + 10}, "Gelir hesabı / tur", ColorGold, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: popup.X + popup.W - 12, Y: popup.Y + 10, W: 0}, formatSignedAmount(status.NetChange), ColorGold, gameui.TextSmall, gameui.TextAlignEnd)
	drawUISeparator(screen, float32(popup.X+10), float32(popup.Y+32), float32(popup.W-20), 1, panelBorder)

	lines := goldIncomePopupLines(status)
	y := popup.Y + 42
	for _, line := range lines {
		if line.value == 0 && line.label != "Ordu masrafı" {
			continue
		}
		drawUIKeyValueRowWithGap(screen, popup.X+12, y, popup.W-24, line.label, formatSignedAmount(line.value), ColorGray, line.color, 8)
		y += 21
	}
	drawUISeparator(screen, float32(popup.X+10), float32(popup.Y+popup.H-30), float32(popup.W-20), 1, panelBorder)
	drawUILabel(screen, gameui.Rect{X: popup.X + 12, Y: popup.Y + popup.H - 22}, "Net değişim", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: popup.X + popup.W - 12, Y: popup.Y + popup.H - 22, W: 0}, formatSignedAmount(status.NetChange), ColorGold, gameui.TextSmall, gameui.TextAlignEnd)
}
