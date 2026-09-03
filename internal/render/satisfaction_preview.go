package render

import (
	"fmt"
	"image/color"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/satisfaction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func regionSatisfactionBreakdown(gs *state.GameState, rid world.RegionID) satisfaction.Breakdown {
	if gs == nil || rid == "" {
		return satisfaction.Breakdown{}
	}
	return satisfaction.Calculate(gs, gs.Regions[rid])
}

func regionSatisfactionDeltaRect(gs *state.GameState, rid world.RegionID) (gameui.Rect, bool) {
	if gs == nil || rid == "" {
		return gameui.Rect{}, false
	}
	region := gs.Regions[rid]
	if region == nil || region.IsSea || region.OwnerID == "" {
		return gameui.Rect{}, false
	}
	valueX := float64(infoPanelX()) + panelPad + 96
	x := valueX + MeasureText("%"+itoa(region.Satisfaction), FaceMed) + 4
	barX, _ := regionPanelTaxBarLayout(float32(infoPanelX()+float32(panelPad)), infoPanelW-float32(panelPad*2))
	if x+22 > float64(barX)-2 {
		x = float64(barX) - 24
	}
	y := regionPanelSatisfactionRowY(gs, region)
	return gameui.Rect{X: x, Y: y, W: 22, H: regionPanelStatRowGap}, true
}

func regionPanelSatisfactionRowY(gs *state.GameState, region *world.Region) float64 {
	if gs == nil || region == nil {
		return 0
	}
	y := regionPanelStatRowsStartY(gs, region.OwnerID)
	if gs.RegionBlockadeEconomicEffect(region).BlockadePercent > 0 {
		y += 16
	}
	return y
}

func satisfactionDeltaText(delta int) string {
	return fmt.Sprintf("%+d", delta)
}

func satisfactionDeltaColor(delta int) color.RGBA {
	switch {
	case delta > 0:
		return color.RGBA{100, 220, 100, 255}
	case delta < 0:
		return color.RGBA{220, 90, 90, 255}
	default:
		return ColorGray
	}
}

func drawSatisfactionDelta(screen *ebiten.Image, rect gameui.Rect, delta int) {
	drawUILabel(screen, rect, satisfactionDeltaText(delta), satisfactionDeltaColor(delta), gameui.TextSmall, gameui.TextAlignStart)
}

func satisfactionBreakdownLines(gs *state.GameState, region *world.Region, breakdown satisfaction.Breakdown) []tooltipLine {
	taxRate := 0
	if region != nil {
		taxRate = region.TaxRate
	}
	lines := []tooltipLine{
		{text: fmt.Sprintf("Vergi (%%%d): %+d", taxRate, breakdown.Tax), col: satisfactionDeltaColor(breakdown.Tax)},
		{text: fmt.Sprintf("Tahıl arzı: %+d", breakdown.Grain), col: satisfactionDeltaColor(breakdown.Grain)},
		{text: fmt.Sprintf("Teknoloji: %+d", breakdown.Technology), col: satisfactionDeltaColor(breakdown.Technology)},
		{text: fmt.Sprintf("Savaş yorgunluğu: %+d", breakdown.WarFatigue), col: satisfactionDeltaColor(breakdown.WarFatigue)},
		{text: fmt.Sprintf("Genişleme: %+d", breakdown.Overextension), col: satisfactionDeltaColor(breakdown.Overextension)},
		{text: fmt.Sprintf("Yerel ordu: %+d", breakdown.Army), col: satisfactionDeltaColor(breakdown.Army)},
		{text: fmt.Sprintf("Yıllık yıpranma: %+d", breakdown.Annual), col: satisfactionDeltaColor(breakdown.Annual)},
		{text: fmt.Sprintf("Kuşatma: %+d", breakdown.Siege), col: satisfactionDeltaColor(breakdown.Siege)},
		{text: fmt.Sprintf("Toplam: %+d", breakdown.Total), col: satisfactionDeltaColor(breakdown.Total)},
	}
	if region != nil && gs != nil {
		if owner := gs.Factions[faction.FactionID(region.OwnerID)]; owner != nil && owner.OverlordID != "" {
			rate := owner.TributeRate
			if !owner.TributeRateConfigured {
				rate = diplomacy.VassalTributeRatePercent()
			}
			lines = append([]tooltipLine{{text: fmt.Sprintf("Haraç (%%%d): %+d", rate, breakdown.Tribute), col: satisfactionDeltaColor(breakdown.Tribute)}}, lines...)
		}
	}
	buildingLines := satisfactionBuildingLines(gs, region)
	if len(buildingLines) == 0 {
		buildingLines = append(buildingLines, tooltipLine{text: "Binalar +0", col: satisfactionDeltaColor(0)})
	}
	otherLines := lines[1:]
	lines = append([]tooltipLine{lines[0]}, buildingLines...)
	return append(lines, otherLines...)
}

func satisfactionBuildingLines(gs *state.GameState, region *world.Region) []tooltipLine {
	if gs == nil || region == nil || len(region.Buildings) == 0 {
		return nil
	}
	amountByID := make(map[string]int, len(region.Buildings))
	order := make([]string, 0, len(region.Buildings))
	for _, buildingID := range region.Buildings {
		building := gs.BuildingTypes[buildingID]
		if building == nil || building.SatBonus == 0 {
			continue
		}
		if _, exists := amountByID[buildingID]; !exists {
			order = append(order, buildingID)
		}
		amountByID[buildingID] += building.SatBonus
	}
	lines := make([]tooltipLine, 0, len(order))
	for _, buildingID := range order {
		building := gs.BuildingTypes[buildingID]
		name := buildingID
		if building != nil {
			name = building.NameTR
			if name == "" {
				name = building.Name
			}
			if name == "" {
				name = buildingID
			}
		}
		amount := amountByID[buildingID]
		lines = append(lines, tooltipLine{text: fmt.Sprintf("%s %+d", name, amount), col: satisfactionDeltaColor(amount)})
	}
	return lines
}

func drawSatisfactionTooltip(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, mx, my float64) {
	region := gs.Regions[rid]
	if region == nil {
		return
	}
	breakdown := regionSatisfactionBreakdown(gs, rid)
	lines := satisfactionBreakdownLines(gs, region, breakdown)
	w := 292.0
	h := 54.0 + float64(len(lines))*16
	x, y, ww, hh := tooltipRect(mx, my, w, h)
	drawTooltipBox(screen, x, y, ww, hh)
	DrawText(screen, "Memnuniyet Hesabı", x+12, y+12, FaceMed, ColorGold)

	drawUIRichTextBlock(screen, gameui.Rect{X: x + 12, Y: y + 48, W: w - 24, H: h - 48}, tooltipRichLines(lines), 16)
}
