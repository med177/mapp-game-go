package render

import (
	"image/color"

	"mapp-game-go/internal/economy"
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

// grainHUDValueRect, çizim ve cursor hit-test'inin aynı Tahıl rakamı
// geometri sözleşmesini kullanmasını sağlar.
func grainHUDValueRect() gameui.Rect {
	leftCol1, _, _, leftColW, _ := topResourceHUDColumns()
	return gameui.Rect{
		X: leftCol1 + 42,
		Y: 12 - 5,
		W: leftColW - 42,
		H: 20,
	}
}

func playerGrainEconomyStatus(gs *state.GameState) state.GrainEconomyStatus {
	if gs == nil {
		return state.GrainEconomyStatus{}
	}
	if status, ok := gs.GrainEconomy[gs.PlayerFactionID]; ok {
		return status
	}

	// İlk ekonomi çözümlemesi gerçekleşmeden önce de popup, HUD'daki tahıl
	// değişimini oluşturan aynı üretim/tüketim helper'larından doldurulur.
	fid := gs.PlayerFactionID
	production := gs.FactionProductionSummary(fid).Grain
	civilianDemand := 0
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) || gs.SiegeAt(region.ID) != nil {
			continue
		}
		civilianDemand += gs.CivilianGrainDemandForRegion(region)
	}
	armyUpkeep := 0
	for _, currentArmy := range gs.Armies {
		if currentArmy != nil && currentArmy.OwnerID == string(fid) {
			armyUpkeep += gs.EffectiveArmyGrainUpkeep(currentArmy)
		}
	}
	totalDemand := civilianDemand + armyUpkeep
	return state.GrainEconomyStatus{
		FactionID:      fid,
		Production:     production,
		CivilianDemand: civilianDemand,
		ArmyUpkeep:     armyUpkeep,
		TotalDemand:    totalDemand,
		NetChange:      production - totalDemand,
	}
}

type grainEconomyPopupLine struct {
	label       string
	value       int
	color       color.RGBA
	signedValue bool
}

func grainEconomyPopupLines(gs *state.GameState, status state.GrainEconomyStatus) [6]grainEconomyPopupLine {
	marketOffer := 0
	if gs != nil {
		marketOffer = gs.MarketSellOffer(gs.PlayerFactionID, economy.GoodGrain)
	}
	return [6]grainEconomyPopupLine{
		{label: "Üretim", value: status.Production, color: color.RGBA{145, 220, 155, 255}, signedValue: true},
		{label: "Halk tüketimi", value: -status.CivilianDemand, color: ColorRed, signedValue: true},
		{label: "Ordu tüketimi", value: -status.ArmyUpkeep, color: ColorRed, signedValue: true},
		{label: "Toplam tüketim", value: -status.TotalDemand, color: ColorRed, signedValue: true},
		{label: "Pazar arzı", value: marketOffer, color: color.RGBA{145, 220, 155, 255}},
		{label: "Otomatik ihracat", value: -status.AutoExportSold, color: ColorRed, signedValue: true},
	}
}

func grainEconomyPopupRect() gameui.Rect {
	leftCol1, _, _, leftColW, _ := topResourceHUDColumns()
	const popupW = 330.0
	const popupH = 220.0
	x := leftCol1 + leftColW - popupW
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

func (r *Renderer) grainEconomyPopupHovering(fx, fy float64) bool {
	if r == nil || r.gs == nil || r.gs.PlayerFactionID == "" {
		return false
	}
	return grainHUDValueRect().Hit(fx, fy)
}

func (r *Renderer) drawGrainEconomyPopup(screen *ebiten.Image) {
	if r == nil || r.gs == nil {
		return
	}
	mx, my := ebiten.CursorPosition()
	if !r.grainEconomyPopupHovering(float64(mx), float64(my)) {
		return
	}

	status := playerGrainEconomyStatus(r.gs)
	popup := grainEconomyPopupRect()
	drawUIPanelFrame(screen, popup, panelBg, panelBorder, 1.5, 5)
	drawUILabel(screen, gameui.Rect{X: popup.X + 12, Y: popup.Y + 10}, "Tahıl hesabı / tur", ColorGold, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: popup.X + popup.W - 12, Y: popup.Y + 10, W: 0}, formatSignedAmount(status.NetChange), grainPopupNetColor(status.NetChange), gameui.TextSmall, gameui.TextAlignEnd)
	drawUISeparator(screen, float32(popup.X+10), float32(popup.Y+32), float32(popup.W-20), 1, panelBorder)

	lines := grainEconomyPopupLines(r.gs, status)
	y := popup.Y + 42
	for _, line := range lines {
		if line.value == 0 && line.label != "Toplam tüketim" {
			continue
		}
		value := formatNumberTR(line.value)
		if line.signedValue {
			value = formatSignedAmount(line.value)
		}
		drawUIKeyValueRowWithGap(screen, popup.X+12, y, popup.W-24, line.label, value, ColorGray, line.color, 8)
		y += 21
	}
	drawUISeparator(screen, float32(popup.X+10), float32(popup.Y+popup.H-30), float32(popup.W-20), 1, panelBorder)
	drawUILabel(screen, gameui.Rect{X: popup.X + 12, Y: popup.Y + popup.H - 22}, "Net değişim", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: popup.X + popup.W - 12, Y: popup.Y + popup.H - 22, W: 0}, formatSignedAmount(status.NetChange), grainPopupNetColor(status.NetChange), gameui.TextSmall, gameui.TextAlignEnd)
}

func grainPopupNetColor(value int) color.RGBA {
	if value < 0 {
		return ColorRed
	}
	return color.RGBA{145, 220, 155, 255}
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
