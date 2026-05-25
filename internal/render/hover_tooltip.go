package render

import (
	"fmt"
	"image/color"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func DrawHoverTooltip(screen *ebiten.Image, gs *state.GameState, rid world.RegionID) {
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
	if uid := RecruitPanelHitTest(fx, fy, gs, rid); uid != "" {
		drawUnitTooltip(screen, gs, uid, fx, fy)
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
	spriteH := float32(54)
	nameH := float32(16)
	rowH := spriteH + nameH + 5

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
	x, y, w, h := tooltipRect(mx, my, 300, 154)
	drawTooltipBox(screen, x, y, w, h)

	DrawText(screen, b.NameTR, x+84, y+12, FaceMed, ColorGold)
	DrawText(screen, fmt.Sprintf("Maliyet: %d altın", b.GoldCost), x+84, y+34, FaceSmall, ColorWhite)

	status := "İnşa edilebilir"
	statusCol := color.RGBA{120, 210, 120, 230}
	level := 0
	for _, builtID := range region.Buildings {
		if builtID == bid {
			level++
		}
	}
	maxLevel := 1
	if b.MaxPerRegion > 0 {
		maxLevel = b.MaxPerRegion
	}
	if level > 0 {
		status = fmt.Sprintf("Seviye: Lv%d/%d", level, maxLevel)
		statusCol = ColorGold
	}
	if level >= maxLevel {
		status = fmt.Sprintf("Maksimum seviye (Lv%d)", level)
		statusCol = color.RGBA{190, 170, 110, 230}
	}
	DrawText(screen, status, x+84, y+52, FaceSmall, statusCol)

	if buildingSheet != nil {
		r := buildingSpriteRect(bid, buildingSheet)
		sub := buildingSheet.SubImage(r).(*ebiten.Image)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(70/float64(r.Dx()), 58/float64(r.Dy()))
		op.GeoM.Translate(x+10, y+14)
		screen.DrawImage(sub, op)
	}

	lines := buildingEffectLines(b)
	for i, line := range lines {
		DrawText(screen, line, x+12, y+82+float64(i)*16, FaceSmall, ColorGray)
	}
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
	if b.RequiredTerrain != "" {
		lines = append(lines, "Arazi: "+b.RequiredTerrain)
	}
	if len(lines) == 0 {
		lines = append(lines, "Yerel gelişim binası")
	}
	return lines
}

func drawUnitTooltip(screen *ebiten.Image, gs *state.GameState, uid string, mx, my float64) {
	utype := gs.UnitTypes[uid]
	if utype == nil {
		return
	}
	ensureArmySheet()
	x, y, w, h := tooltipRect(mx, my, 300, 166)
	drawTooltipBox(screen, x, y, w, h)

	DrawText(screen, utype.NameTR, x+84, y+12, FaceMed, ColorGold)
	DrawText(screen, fmt.Sprintf("Maliyet: %d altın", utype.GoldCost), x+84, y+34, FaceSmall, ColorWhite)
	DrawText(screen, fmt.Sprintf("Bakım: %d tahıl/tur", utype.GrainUpkeep), x+84, y+52, FaceSmall, ColorGray)

	if armySheet != nil {
		r := unitSpriteRect(uid, armySheet)
		if !r.Empty() {
			sub := armySheet.SubImage(r).(*ebiten.Image)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(70/float64(r.Dx()), 58/float64(r.Dy()))
			op.GeoM.Translate(x+10, y+14)
			screen.DrawImage(sub, op)
		}
	}

	DrawText(screen, fmt.Sprintf("Saldırı: %d   Savunma: %d", utype.Attack, utype.Defense), x+12, y+84, FaceSmall, ColorGray)
	DrawText(screen, fmt.Sprintf("Moral: %d   Can: %d", utype.Morale, utype.HP), x+12, y+102, FaceSmall, ColorGray)
	DrawText(screen, unitRequirementText(gs, utype.RequiredBldg, utype.RequiredTech), x+12, y+124, FaceSmall, color.RGBA{170, 145, 90, 230})
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
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), color.RGBA{10, 8, 6, 245}, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1.5, panelBorder, false)
	vector.FillRect(screen, float32(x), float32(y), float32(w), 3, panelBorder, false)
}

func unitRequirementText(gs *state.GameState, buildingID, techID string) string {
	req := "Gereksinim: "
	if buildingID == "" && techID == "" {
		return req + "Yok"
	}
	first := true
	if buildingID != "" {
		name := buildingID
		if b := gs.BuildingTypes[buildingID]; b != nil {
			name = b.NameTR
		}
		req += name
		first = false
	}
	if techID != "" {
		if !first {
			req += ", "
		}
		if t := gs.TechTypes[techID]; t != nil {
			req += t.NameTR
		} else {
			req += techID
		}
	}
	return req
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
