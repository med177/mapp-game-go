package render

import (
	"fmt"
	"image"
	"image/color"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type tooltipLine struct {
	text string
	col  color.RGBA
}

func DrawHoverTooltip(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, recruitPanelOpen bool) {
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)

	if idx := regionDiplomacyButtonHit(fx, fy, gs, rid); idx >= 0 {
		region := gs.Regions[rid]
		if region != nil {
			if reason := regionDiplomacyButtonDisabledReason(gs, region.OwnerID, idx); reason != "" {
				drawSmallHoverHint(screen, reason, fx, fy)
				return
			}
		}
	}

	if bid := BuildingGridHoverID(fx, fy, gs, rid); bid != "" {
		drawBuildingTooltip(screen, gs, rid, bid, fx, fy)
		return
	}
	if recruitPanelOpen {
		if uid := RecruitPanelHitTest(fx, fy, gs, rid); uid != "" {
			drawUnitTooltip(screen, gs, rid, uid, fx, fy)
		}
	}
}

func BuildingGridHoverID(mx, my float64, gs *state.GameState, rid world.RegionID) string {
	if rid == "" {
		return ""
	}
	region, ok := gs.Regions[rid]
	if !ok || region.IsSea {
		return ""
	}

	px := infoPanelX()
	pw := infoPanelW
	startY := buildingGridStartY(gs, region)

	const cols = 3
	pad := float32(panelPad)
	availW := pw - pad*2
	slotW := availW / float32(cols)
	spriteH := float32(76)
	nameH := float32(18)
	rowH := spriteH + nameH + 7

	display := visibleBuildingIDs(gs, region)
	for i, bid := range display {
		col := i % cols
		row := i / cols
		sx := px + pad + float32(col)*slotW
		sy := startY + float32(row)*rowH
		innerW := slotW - 3
		if mx >= float64(sx) && mx <= float64(sx+innerW) && my >= float64(sy) && my <= float64(sy+spriteH+nameH) {
			return bid
		}
	}
	return ""
}

func drawBuildingTooltip(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, bid string, mx, my float64) {
	b := gs.BuildingTypes[bid]
	region := gs.Regions[rid]
	if b == nil || region == nil {
		return
	}

	ensureBuildingSheet()
	costLines := buildingCostRequirementLines(gs, b)
	reqLines, reqMissing := buildingRequirementLines(region, b)
	effectLines := buildingEffectLines(b)
	status, statusCol := buildingAvailabilityStatus(gs, region, b, reqMissing)
	tooltipH := 146.0 + float64(len(costLines))*14 + float64(len(reqLines))*14 + float64(len(effectLines))*16
	x, y, w, h := tooltipRect(mx, my, 308, tooltipH)
	drawTooltipBox(screen, x, y, w, h)

	iconX, iconY := x+10.0, y+14.0
	iconW, iconH := 70.0, 58.0
	textX := iconX + iconW + 12.0

	DrawText(screen, b.NameTR, textX, y+12, FaceMed, ColorGold)
	DrawText(screen, "Durum:", textX, y+34, FaceSmall, ColorGray)
	DrawText(screen, status, textX+46, y+34, FaceSmall, statusCol)

	DrawText(screen, "Maliyet:", textX, y+50, FaceSmall, ColorGray)
	for i, line := range costLines {
		DrawText(screen, line.text, textX, y+64+float64(i)*14, FaceSmall, line.col)
	}

	reqY := y + 64 + float64(len(costLines))*14 + 2
	DrawText(screen, "Gereksinim:", textX, reqY, FaceSmall, ColorGray)
	for i, line := range reqLines {
		DrawText(screen, line.text, textX, reqY+14+float64(i)*14, FaceSmall, line.col)
	}

	if buildingSheet != nil {
		vector.FillRect(screen, float32(iconX), float32(iconY), float32(iconW), float32(iconH), color.RGBA{252, 252, 252, 242}, false)
		vector.StrokeRect(screen, float32(iconX), float32(iconY), float32(iconW), float32(iconH), 1, color.RGBA{160, 160, 160, 225}, false)
		r := buildingSpriteRect(bid, buildingSheet)
		sub := buildingSheet.SubImage(r).(*ebiten.Image)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(iconW/float64(r.Dx()), iconH/float64(r.Dy()))
		op.GeoM.Translate(iconX, iconY)
		screen.DrawImage(sub, op)
	}

	effectY := reqY + 14 + float64(len(reqLines))*14 + 8
	DrawText(screen, "Etkiler:", textX, effectY, FaceSmall, ColorGray)
	effectY += 14
	for i, line := range effectLines {
		DrawText(screen, line, textX, effectY+float64(i)*16, FaceSmall, ColorGray)
	}
}

func buildingAvailabilityStatus(gs *state.GameState, region *world.Region, b *city.Building, reqMissing bool) (string, color.RGBA) {
	if b == nil || region == nil {
		return "Bilinmiyor", ColorGray
	}
	level := 0
	for _, builtID := range region.Buildings {
		if builtID == b.ID {
			level++
		}
	}
	maxLevel := 1
	if b.MaxPerRegion > 0 {
		maxLevel = b.MaxPerRegion
	}
	if level >= maxLevel {
		return fmt.Sprintf("Maksimum seviye (Lv%d)", level), color.RGBA{190, 170, 110, 230}
	}
	if level > 0 {
		return fmt.Sprintf("Seviye: Lv%d/%d", level, maxLevel), ColorGold
	}
	if reqMissing {
		return "Gereksinim eksik", ColorRed
	}
	if !buildingCost(gs, b).CanAfford(gs.Factions[gs.PlayerFactionID]) {
		return "Kaynak yetersiz", ColorRed
	}
	return "İnşa edilebilir", color.RGBA{120, 210, 120, 230}
}

func buildingCost(gs *state.GameState, b *city.Building) economy.ResourceCost {
	_ = gs
	if b == nil {
		return economy.ResourceCost{}
	}
	return economy.ResourceCost{
		Gold:   b.GoldCost,
		Grain:  b.GrainCost,
		Iron:   b.IronCost,
		Timber: b.TimberCost,
		Stone:  b.StoneCost,
	}
}

func buildingCostRequirementLines(gs *state.GameState, b *city.Building) []tooltipLine {
	return resourceTooltipLines(gs, buildingCost(gs, b))
}

func buildingRequirementLines(region *world.Region, b *city.Building) ([]tooltipLine, bool) {
	if b == nil || region == nil {
		return []tooltipLine{{text: "Gereksinim bilgisi yok", col: ColorGray}}, false
	}
	if b.RequiredTerrain == "" {
		return []tooltipLine{{text: "Ek koşul yok", col: color.RGBA{170, 145, 90, 230}}}, false
	}
	want := terrainLabel(world.TerrainType(b.RequiredTerrain))
	have := terrainLabel(region.Terrain)
	missing := string(region.Terrain) != b.RequiredTerrain
	col := color.RGBA{170, 145, 90, 230}
	if missing {
		col = ColorRed
	}
	return []tooltipLine{{
		text: fmt.Sprintf("Arazi: %s (mevcut: %s)", want, have),
		col:  col,
	}}, missing
}

func buildingEffectLines(b *city.Building) []string {
	lines := []string{}
	if b.GoldMod != 1 {
		lines = append(lines, fmt.Sprintf("Altın geliri: x%.1f", b.GoldMod))
	}
	if b.GrainMod != 1 {
		lines = append(lines, fmt.Sprintf("Tahıl üretimi: x%.1f", b.GrainMod))
	}
	if b.SatBonus != 0 {
		lines = append(lines, fmt.Sprintf("Memnuniyet: %+d", b.SatBonus))
	}
	if b.DefBonus != 0 {
		lines = append(lines, fmt.Sprintf("Savunma: %+d", b.DefBonus))
	}
	if len(lines) == 0 {
		lines = append(lines, "Yerel gelişim binası")
	}
	return lines
}

func drawUnitTooltip(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, uid string, mx, my float64) {
	utype := gs.UnitTypes[uid]
	if utype == nil {
		return
	}

	ensureArmySheet()
	costLines := unitCostRequirementLines(gs, utype)
	reqLines, reqMissing := unitRequirementLines(gs, rid, utype)
	status, statusCol := unitAvailabilityStatus(gs, rid, utype, reqMissing)
	tooltipH := 190.0 + float64(len(costLines))*14 + float64(len(reqLines))*14
	x, y, w, h := tooltipRect(mx, my, 328, tooltipH)
	drawTooltipBox(screen, x, y, w, h)

	iconX, iconY := x+10.0, y+14.0
	iconW, iconH := float64(recruitCardW), 76.0
	textX := iconX + iconW + 12

	vector.FillRect(screen, float32(iconX), float32(iconY), float32(iconW), float32(iconH), color.RGBA{252, 252, 252, 242}, false)
	vector.StrokeRect(screen, float32(iconX), float32(iconY), float32(iconW), float32(iconH), 1, color.RGBA{160, 160, 160, 225}, false)

	DrawText(screen, utype.NameTR, textX, y+12, FaceMed, ColorGold)
	DrawText(screen, "Durum:", textX, y+34, FaceSmall, ColorGray)
	DrawText(screen, status, textX+46, y+34, FaceSmall, statusCol)

	DrawText(screen, "Maliyet:", textX, y+50, FaceSmall, ColorGray)
	for i, line := range costLines {
		DrawText(screen, line.text, textX, y+64+float64(i)*14, FaceSmall, line.col)
	}

	reqY := y + 64 + float64(len(costLines))*14 + 2
	DrawText(screen, "Gereksinim:", textX, reqY, FaceSmall, ColorGray)
	for i, line := range reqLines {
		DrawText(screen, line.text, textX, reqY+14+float64(i)*14, FaceSmall, line.col)
	}

	upkeepY := reqY + 14 + float64(len(reqLines))*14 + 6
	DrawText(screen, fmt.Sprintf("Bakım: %d tahıl/tur", utype.GrainUpkeep), textX, upkeepY, FaceSmall, ColorGray)

	if armySheet != nil {
		r := unitSpriteRect(uid, armySheet)
		if !r.Empty() {
			sub := armySheet.SubImage(r).(*ebiten.Image)
			op := &ebiten.DrawImageOptions{}
			fitW := iconW + 50
			fitH := iconH + 40
			scale := fitW / float64(r.Dx())
			if hScale := fitH / float64(r.Dy()); hScale < scale {
				scale = hScale
			}
			drawW := float64(r.Dx()) * scale
			drawH := float64(r.Dy()) * scale
			if recruitClipBuf != nil {
				clipW := int(iconW - 2)
				clipH := int(iconH - 2)
				if clipW > 0 && clipH > 0 && clipW <= 160 && clipH <= 120 {
					recruitClipBuf.Clear()
					op.GeoM.Scale(scale, scale)
					op.GeoM.Translate(float64(clipW)/2-drawW/2, float64(clipH)/2-drawH/2)
					recruitClipBuf.DrawImage(sub, op)
					cropped := recruitClipBuf.SubImage(image.Rect(0, 0, clipW, clipH)).(*ebiten.Image)
					dst := &ebiten.DrawImageOptions{}
					dst.GeoM.Translate(iconX+1, iconY+1)
					screen.DrawImage(cropped, dst)
				}
			}
		}
	}

	statY := upkeepY + 18
	DrawText(screen, fmt.Sprintf("Saldırı: %d", utype.Attack), textX, statY, FaceSmall, ColorGray)
	statY += 16
	DrawText(screen, fmt.Sprintf("Savunma: %d", utype.Defense), textX, statY, FaceSmall, ColorGray)
	statY += 16
	DrawText(screen, fmt.Sprintf("Moral: %d", utype.Morale), textX, statY, FaceSmall, ColorGray)
	statY += 16
	DrawText(screen, fmt.Sprintf("Can: %d", utype.HP), textX, statY, FaceSmall, ColorGray)
}

func unitAvailabilityStatus(gs *state.GameState, rid world.RegionID, utype *army.UnitType, reqMissing bool) (string, color.RGBA) {
	if utype == nil {
		return "Bilinmiyor", ColorGray
	}
	ff := gs.Factions[gs.PlayerFactionID]
	if reqMissing {
		return "Gereksinim eksik", ColorRed
	}
	if !unitCost(utype).CanAfford(ff) {
		return "Kaynak yetersiz", ColorRed
	}
	return "Yetiştirilebilir", color.RGBA{120, 210, 120, 230}
}

func unitRequirementLines(gs *state.GameState, rid world.RegionID, utype *army.UnitType) ([]tooltipLine, bool) {
	if utype == nil {
		return []tooltipLine{{text: "Gereksinim bilgisi yok", col: ColorGray}}, false
	}
	region := gs.Regions[rid]
	ff := gs.Factions[gs.PlayerFactionID]
	lines := make([]tooltipLine, 0, 2)
	missing := false

	if utype.RequiredBldg != "" {
		requiredLevel := utype.RequiredBldgLevel
		if requiredLevel <= 0 {
			requiredLevel = 1
		}
		currentLevel := 0
		name := utype.RequiredBldg
		if b := gs.BuildingTypes[utype.RequiredBldg]; b != nil {
			name = b.NameTR
		}
		if region != nil {
			for _, bid := range region.Buildings {
				if bid == utype.RequiredBldg {
					currentLevel++
				}
			}
		}
		col := color.RGBA{170, 145, 90, 230}
		if currentLevel < requiredLevel {
			col = ColorRed
			missing = true
		}
		lines = append(lines, tooltipLine{
			text: fmt.Sprintf("%s Lv%d gerekli (mevcut: Lv%d)", name, requiredLevel, currentLevel),
			col:  col,
		})
	}

	if utype.RequiredTech != "" {
		name := utype.RequiredTech
		if t := gs.TechTypes[utype.RequiredTech]; t != nil {
			name = t.NameTR
		}
		done := ff != nil && ff.Research.Completed[utype.RequiredTech]
		col := color.RGBA{170, 145, 90, 230}
		if !done {
			col = ColorRed
			missing = true
		}
		stateText := "hazır"
		if !done {
			stateText = "eksik"
		}
		lines = append(lines, tooltipLine{
			text: fmt.Sprintf("Teknoloji: %s (%s)", name, stateText),
			col:  col,
		})
	}

	if len(lines) == 0 {
		return []tooltipLine{{text: "Ek koşul yok", col: color.RGBA{170, 145, 90, 230}}}, false
	}
	return lines, missing
}

func unitCostRequirementLines(gs *state.GameState, utype *army.UnitType) []tooltipLine {
	if utype == nil {
		return []tooltipLine{{text: "-", col: ColorGray}}
	}
	return resourceTooltipLines(gs, unitCost(utype))
}

func resourceTooltipLines(gs *state.GameState, cost economy.ResourceCost) []tooltipLine {
	f := gs.Factions[gs.PlayerFactionID]
	lines := make([]tooltipLine, 0, 5)

	appendLine := func(name string, need int, have int) {
		if need <= 0 {
			return
		}
		col := ColorWhite
		text := fmt.Sprintf("%s: %d/%d", name, have, need)
		if have < need {
			col = ColorRed
			text += " eksik"
		}
		lines = append(lines, tooltipLine{text: text, col: col})
	}

	if f == nil {
		appendLine("Altın", cost.Gold, 0)
		appendLine("Tahıl", cost.Grain, 0)
		appendLine("Demir", cost.Iron, 0)
		appendLine("Kereste", cost.Timber, 0)
		appendLine("Taş", cost.Stone, 0)
	} else {
		appendLine("Altın", cost.Gold, f.Gold)
		appendLine("Tahıl", cost.Grain, f.Grain)
		appendLine("Demir", cost.Iron, f.Iron)
		appendLine("Kereste", cost.Timber, f.Timber)
		appendLine("Taş", cost.Stone, f.Stone)
	}

	if len(lines) == 0 {
		return []tooltipLine{{text: "Bedava", col: ColorWhite}}
	}
	return lines
}

func tooltipRect(mx, my float64, w, h float64) (float64, float64, float64, float64) {
	x := mx + 18
	y := my + 18
	if x+w > ScreenWidth-8 {
		x = mx - w - 18
	}
	if y+h > ScreenHeight-8 {
		y = my - h - 18
	}
	if x < 8 {
		x = 8
	}
	if y < 8 {
		y = 8
	}
	return x, y, w, h
}

func drawTooltipBox(screen *ebiten.Image, x, y, w, h float64) {
	tooltip := gameui.Tooltip{
		Rect:    gameui.Rect{X: x, Y: y, W: w, H: h},
		Visible: true,
	}
	gameui.DrawTooltip(screen, tooltip, hoverTooltipStyle, renderSmallText)
	vector.FillRect(screen, float32(x), float32(y), float32(w), 3, panelBorder, false)
}

func drawSmallHoverHint(screen *ebiten.Image, message string, mx, my float64) {
	w := MeasureText(message, FaceSmall) + 20
	if w < 220 {
		w = 220
	}
	x, y, ww, hh := tooltipRect(mx, my, w, 40)
	drawTooltipBox(screen, x, y, ww, hh)
	DrawText(screen, message, x+10, y+12, FaceSmall, ColorGray)
}
