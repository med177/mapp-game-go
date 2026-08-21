package render

import (
	"fmt"
	"image/color"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
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

const (
	buildingTooltipImageSize = 200.0
	buildingTooltipWidth     = 450.0
	unitTooltipImageExtraH   = 50.0
	unitTooltipWidth         = 358.0
)

func unitTooltipImageMetrics() (width, height float64) {
	height = float64(unitSpriteHeight(recruitCardW)) + unitTooltipImageExtraH
	width = height / float64(unitSpriteAspectH)
	return width, height
}

func tooltipRichLines(lines []tooltipLine) []gameui.RichTextLine {
	out := make([]gameui.RichTextLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, gameui.RichTextLine{
			Text:    line.text,
			Color:   line.col,
			Variant: gameui.TextSmall,
			Align:   gameui.TextAlignStart,
		})
	}
	return out
}

func plainRichLines(lines []string, col color.RGBA) []gameui.RichTextLine {
	out := make([]gameui.RichTextLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, gameui.RichTextLine{
			Text:    line,
			Color:   col,
			Variant: gameui.TextSmall,
			Align:   gameui.TextAlignStart,
		})
	}
	return out
}

func drawTooltipStatusRow(screen *ebiten.Image, x, y float64, value string, valueCol color.Color) {
	row := gameui.NewKeyValueRow(gameui.Rect{X: x, Y: y, W: 220}, "Durum:", value)
	row.LabelColor = ColorGray
	row.ValueColor = valueCol
	row.LabelVariant = gameui.TextSmall
	row.ValueVariant = gameui.TextSmall
	row.Gap = 14
	row.ValueAlign = gameui.TextAlignStart
	drawUIKeyValueWidget(screen, row)
}

func DrawHoverTooltip(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, aid army.ArmyID, recruitPanelOpen bool) {
	DrawHoverTooltipWithTab(screen, gs, rid, aid, recruitPanelOpen, regionPanelTabBuildings)
}

func DrawHoverTooltipWithTab(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, aid army.ArmyID, recruitPanelOpen bool, activeTab regionPanelTab) {
	drawHoverTooltipWithTab(screen, gs, rid, aid, recruitPanelOpen, activeTab, true)
}

func drawHoverTooltipWithTab(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, aid army.ArmyID, recruitPanelOpen bool, activeTab regionPanelTab, armyPanelActive bool) {
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)

	// Ordu paneli recruit ve bölge panellerinin üstünde çizilir; örtüşme
	// durumunda hover da aynı görsel katman sırasını izlemelidir.
	if armyPanelActive && aid != "" {
		if a := gs.Armies[aid]; a != nil {
			if targetID, ok := MergeButtonTargetAt(fx, fy, gs, aid); ok {
				if target := gs.Armies[targetID]; target != nil {
					drawArmyMergePreviewTooltip(screen, gs, a, target, fx, fy)
					return
				}
			}
			if unit, unitCount, ok := armyPanelUnitHover(fx, fy, gs, aid); ok {
				drawArmyUnitTooltip(screen, gs, a, unit, unitCount, fx, fy)
				return
			}
		}
	}
	if deltaRect, ok := regionSatisfactionDeltaRect(gs, rid); ok && deltaRect.Hit(fx, fy) {
		drawSatisfactionTooltip(screen, gs, rid, fx, fy)
		return
	}

	if regionDiplomacyButtonHitForTab(fx, fy, gs, rid, activeTab) {
		drawSmallHoverHint(screen, "Diplomasi ekranını aç", fx, fy)
		return
	}
	if regionLiberateButtonHitForTab(fx, fy, gs, rid, activeTab) {
		successorID, _ := regionLiberationSuccessor(gs, gs.Regions[rid])
		successorName := factionDisplayName(gs, string(successorID))
		if successorName == "" {
			successorName = string(successorID)
		}
		drawSmallHoverHint(screen, "Ardıl devlet: "+successorName, fx, fy)
		return
	}
	if regionGrainAidButtonHitForTab(fx, fy, gs, rid, activeTab) {
		if reason := gs.GrainAidBlockReason(rid); reason != "" {
			drawSmallHoverHint(screen, reason, fx, fy)
		} else {
			drawSmallHoverHint(screen, "12 tahıl harca, memnuniyeti +10 artır", fx, fy)
		}
		return
	}

	if bid := BuildingGridHoverIDForTab(fx, fy, gs, rid, activeTab); bid != "" {
		drawBuildingTooltip(screen, gs, rid, bid, fx, fy)
		return
	}
	if recruitPanelOpen {
		if uid := RecruitPanelHitTest(fx, fy, gs, rid); uid != "" {
			drawUnitTooltip(screen, gs, rid, uid, fx, fy)
		}
	}
}

type armyMergePreviewRow struct {
	typeID string
	count  int
}

// drawArmyMergePreviewTooltip hedef ordunun birim kompozisyonunu küçük kartlar
// halinde gösterir. Butonun etiketiyle aynı target state'i kullanır; böylece
// hover önizlemesi tıklanacak ordudan farklı bir orduyu anlatamaz.
func drawArmyMergePreviewTooltip(screen *ebiten.Image, gs *state.GameState, source, target *army.Army, mx, my float64) {
	if gs == nil || source == nil || target == nil {
		return
	}

	var rows [army.MaxArmySize]armyMergePreviewRow
	rowCount := 0
	for _, unit := range target.Units {
		rowIndex := -1
		for i := 0; i < rowCount; i++ {
			if rows[i].typeID == unit.TypeID {
				rowIndex = i
				break
			}
		}
		if rowIndex < 0 && rowCount < len(rows) {
			rowIndex = rowCount
			rows[rowIndex].typeID = unit.TypeID
			rowCount++
		}
		if rowIndex >= 0 {
			rows[rowIndex].count++
		}
	}

	const (
		previewWidth   = 326.0
		tileWidth      = 76.0
		tileHeight     = 72.0
		tileGap        = 4.0
		previewPadding = 8.0
	)
	columns := 4
	rowLines := (rowCount + columns - 1) / columns
	previewHeight := 52.0 + float64(rowLines)*tileHeight + float64(maxInt(rowLines-1, 0))*tileGap + previewPadding
	x, y, w, h := tooltipRect(mx, my, previewWidth, previewHeight)
	drawTooltipBox(screen, x, y, w, h)

	DrawText(screen, "Hedef ordu: "+itoa(len(target.Units))+" birim", x+10, y+10, FaceSmall, ColorGold)
	DrawText(screen, "Birleşince: "+itoa(mergeResultUnitCount(source, target)), x+10, y+27, FaceSmall, ColorWhite)

	for index := 0; index < rowCount; index++ {
		column := index % columns
		line := index / columns
		tileX := x + previewPadding + float64(column)*(tileWidth+tileGap)
		tileY := y + 46 + float64(line)*(tileHeight+tileGap)
		vector.FillRect(screen, float32(tileX), float32(tileY), float32(tileWidth), float32(tileHeight), color.RGBA{248, 246, 238, 235}, false)
		vector.StrokeRect(screen, float32(tileX), float32(tileY), float32(tileWidth), float32(tileHeight), 1, color.RGBA{150, 125, 72, 220}, false)

		if sprite := unitSpriteForFaction(gs, target.OwnerID, rows[index].typeID); sprite != nil {
			drawUnitSpriteCard(screen, sprite, float32(tileX+(tileWidth-30)/2), float32(tileY+3), 30, [3]float32{1, 1, 1})
		}
		DrawTextCentered(screen, "x"+itoa(rows[index].count), tileX+tileWidth/2, tileY+57, FaceMed, color.RGBA{115, 80, 20, 255})
	}
}

func BuildingGridHoverIDForTab(mx, my float64, gs *state.GameState, rid world.RegionID, activeTab regionPanelTab) string {
	if activeTab != regionPanelTabBuildings {
		return ""
	}
	return BuildingGridHoverID(mx, my, gs, rid)
}

func BuildingGridHoverID(mx, my float64, gs *state.GameState, rid world.RegionID) string {
	if gs == nil || rid == "" {
		return ""
	}
	region, ok := gs.Regions[rid]
	if !ok || !regionBuildingActionsAvailable(gs, region) {
		return ""
	}
	if bid, ok := lastDrawnBuildingGridHit(mx, my, rid); ok {
		return bid
	}

	px := infoPanelX()
	pw := infoPanelW
	startY := buildingGridStartY(gs, region, false)

	for _, card := range buildBuildingCardComponents(gs, region, px, startY, pw) {
		if card.HitTest(mx, my) {
			return card.ID
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

	costLines := buildingCostRequirementLines(gs, b)
	reqLines, reqMissing := buildingRequirementLines(region, b)
	effectLines := buildingEffectLines(b)
	effectLines = append(effectLines, buildingLandCapacityEffectLines(gs, region, b)...)
	effectLines = append(effectLines, buildingNavalCapacityEffectLines(gs, region, b)...)
	iconW, iconH := buildingTooltipImageSize, buildingTooltipImageSize
	status, statusCol := buildingAvailabilityStatus(gs, region, b, reqMissing)
	tooltipH := 146.0 + float64(len(costLines))*14 + float64(len(reqLines))*14 + float64(len(effectLines))*16
	if tooltipH < iconH+28 {
		tooltipH = iconH + 28
	}
	x, y, w, h := tooltipRect(mx, my, buildingTooltipWidth, tooltipH)
	drawTooltipBox(screen, x, y, w, h)

	iconX, iconY := x+10.0, y+14.0
	textX := iconX + iconW + 12.0

	DrawText(screen, b.NameTR, textX, y+12, FaceMed, ColorGold)
	drawTooltipStatusRow(screen, textX, y+34, status, statusCol)

	DrawText(screen, "Maliyet:", textX, y+50, FaceSmall, ColorGray)
	drawUIRichTextBlock(screen, gameui.Rect{X: textX, Y: y + 64}, tooltipRichLines(costLines), 14)

	reqY := y + 64 + float64(len(costLines))*14 + 2
	DrawText(screen, "Gereksinim:", textX, reqY, FaceSmall, ColorGray)
	drawUIRichTextBlock(screen, gameui.Rect{X: textX, Y: reqY + 14}, tooltipRichLines(reqLines), 14)

	if sprite := buildingSpriteImage(bid); sprite != nil {
		vector.FillRect(screen, float32(iconX), float32(iconY), float32(iconW), float32(iconH), color.RGBA{252, 252, 252, 242}, false)
		vector.StrokeRect(screen, float32(iconX), float32(iconY), float32(iconW), float32(iconH), 1, color.RGBA{160, 160, 160, 225}, false)
		spriteRect := buildingSpriteDrawRect(sprite, gameui.Rect{X: iconX, Y: iconY, W: iconW, H: iconH})
		op := &ebiten.DrawImageOptions{}
		bounds := sprite.Bounds()
		op.GeoM.Scale(spriteRect.W/float64(bounds.Dx()), spriteRect.H/float64(bounds.Dy()))
		op.GeoM.Translate(spriteRect.X, spriteRect.Y)
		screen.DrawImage(sprite, op)
	}

	effectY := reqY + 14 + float64(len(reqLines))*14 + 8
	DrawText(screen, "Etkiler:", textX, effectY, FaceSmall, ColorGray)
	drawUIRichTextBlock(screen, gameui.Rect{X: textX, Y: effectY + 14}, plainRichLines(effectLines, ColorGray), 16)
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
		Spice:  b.SpiceCost,
		Cloth:  b.ClothCost,
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
	want := world.TerrainType(b.RequiredTerrain).LabelTR()
	have := region.Terrain.LabelTR()
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
	if b.TradeCapacityMod != 1 {
		lines = append(lines, fmt.Sprintf("Ticaret kapasitesi: x%.2f", b.TradeCapacityMod))
	}
	if b.SatBonus != 0 {
		lines = append(lines, fmt.Sprintf("Memnuniyet: %+d", b.SatBonus))
	}
	if b.DefBonus != 0 {
		lines = append(lines, fmt.Sprintf("Savunma: %+d", b.DefBonus))
	}
	if b.StorageCapacity != 0 {
		lines = append(lines, fmt.Sprintf("Tahıl depolama: +%d", b.StorageCapacity))
	}
	if len(lines) == 0 {
		lines = append(lines, "Yerel gelişim binası")
	}
	return lines
}

// buildingNavalCapacityEffectLines liman tooltip'inde mevcut kapasiteyi ve
// bir sonraki liman seviyesinin devlet sınırına etkisini gösterir.
func buildingNavalCapacityEffectLines(gs *state.GameState, region *world.Region, b *city.Building) []string {
	if gs == nil || region == nil || b == nil || b.ID != "port" || region.OwnerID == "" {
		return nil
	}
	level := 0
	for _, buildingID := range region.Buildings {
		if buildingID == b.ID {
			level++
		}
	}
	maxLevel := b.MaxPerRegion
	if maxLevel <= 0 {
		maxLevel = 1
	}
	ownerID := faction.FactionID(region.OwnerID)
	currentCap := gs.NavalCap(ownerID)
	if level >= maxLevel {
		return []string{fmt.Sprintf("Donanma sınırı: %d (maksimum)", currentCap)}
	}
	nextCap := currentCap + state.NavalCapacityPerPortLevel
	return []string{
		fmt.Sprintf("Donanma kapasitesi: +%d gemi", state.NavalCapacityPerPortLevel),
		fmt.Sprintf("Donanma sınırı: %d → %d", currentCap, nextCap),
	}
}

// buildingLandCapacityEffectLines kışla tooltip'inde savaşçı sınırının ordu
// sayısına bağlı olduğunu ve kışlanın üretim hattı etkisini gösterir.
func buildingLandCapacityEffectLines(gs *state.GameState, region *world.Region, b *city.Building) []string {
	if gs == nil || region == nil || b == nil || b.ID != "barracks" || region.OwnerID == "" {
		return nil
	}
	level := 0
	for _, buildingID := range region.Buildings {
		if buildingID == b.ID {
			level++
		}
	}
	maxLevel := b.MaxPerRegion
	if maxLevel <= 0 {
		maxLevel = 1
	}
	ownerID := faction.FactionID(region.OwnerID)
	landCap := gs.ManpowerCap(ownerID)
	baseArmyCap := landCap / army.MaxArmySize
	productionLimit := state.LandUnitProductionLimit(region)
	if level < maxLevel {
		nextProductionLimit := level + 1
		if nextProductionLimit < 1 {
			nextProductionLimit = 1
		}
		if nextProductionLimit > productionLimit {
			return []string{
				fmt.Sprintf("Savaşçı sınırı: %d (%d temel ordu × %d)", landCap, baseArmyCap, army.MaxArmySize),
				fmt.Sprintf("Kışla üretim limiti: %d → %d birim/tur", productionLimit, nextProductionLimit),
			}
		}
	}
	return []string{
		fmt.Sprintf("Savaşçı sınırı: %d (%d temel ordu × %d; +1 slot ayrı)", landCap, baseArmyCap, army.MaxArmySize),
		fmt.Sprintf("Kışla üretim limiti: %d birim/tur", productionLimit),
	}
}

func drawUnitTooltip(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, uid string, mx, my float64) {
	utype := gs.UnitTypes[uid]
	if utype == nil {
		return
	}

	ensureArmySprites()
	sprite := unitSpriteForFaction(gs, string(gs.PlayerFactionID), uid)
	costLines := unitCostRequirementLines(gs, utype)
	reqLines, reqMissing := unitRequirementLines(gs, rid, utype)
	status, statusCol := unitAvailabilityStatus(gs, utype, reqMissing)
	tooltipH := 190.0 + float64(len(costLines))*14 + float64(len(reqLines))*14
	iconW, iconH := unitTooltipImageMetrics()
	if tooltipH < iconH+28 {
		tooltipH = iconH + 28
	}
	x, y, w, h := tooltipRect(mx, my, unitTooltipWidth, tooltipH)
	drawTooltipBox(screen, x, y, w, h)

	iconX, iconY := x+10.0, y+14.0
	textX := iconX + iconW + 12

	vector.FillRect(screen, float32(iconX), float32(iconY), float32(iconW), float32(iconH), color.RGBA{252, 252, 252, 242}, false)
	vector.StrokeRect(screen, float32(iconX), float32(iconY), float32(iconW), float32(iconH), 1, color.RGBA{160, 160, 160, 225}, false)

	DrawText(screen, utype.NameTR, textX, y+12, FaceMed, ColorGold)
	drawTooltipStatusRow(screen, textX, y+34, status, statusCol)

	DrawText(screen, "Maliyet:", textX, y+50, FaceSmall, ColorGray)
	drawUIRichTextBlock(screen, gameui.Rect{X: textX, Y: y + 64}, tooltipRichLines(costLines), 14)

	reqY := y + 64 + float64(len(costLines))*14 + 2
	DrawText(screen, "Gereksinim:", textX, reqY, FaceSmall, ColorGray)
	drawUIRichTextBlock(screen, gameui.Rect{X: textX, Y: reqY + 14}, tooltipRichLines(reqLines), 14)

	upkeepY := reqY + 14 + float64(len(reqLines))*14 + 6
	DrawText(screen, fmt.Sprintf("Bakım: %d tahıl + %d altın/tur", utype.GrainUpkeep, utype.GoldUpkeep), textX, upkeepY, FaceSmall, ColorGray)

	if sprite != nil {
		drawUnitSpriteCard(screen, sprite, float32(iconX), float32(iconY), float32(iconW), [3]float32{1, 1, 1})
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

// drawArmyUnitTooltip, recruit tooltip'ından ayrı olarak seçili ordudaki
// gerçek birim örneğinin bilgilerini gösterir. Üretim maliyetleri ve
// gereksinimler bu bağlamda anlamlı olmadığı için yalnız adet, tur başı bakım,
// savaş değerleri ve anlık can çizilir.
func drawArmyUnitTooltip(screen *ebiten.Image, gs *state.GameState, a *army.Army, unit army.Unit, unitCount int, mx, my float64) {
	if gs == nil || a == nil {
		return
	}
	utype := gs.UnitTypes[unit.TypeID]
	if utype == nil {
		return
	}

	ensureArmySprites()
	tooltipH := 190.0
	iconW, iconH := unitTooltipImageMetrics()
	if tooltipH < iconH+28 {
		tooltipH = iconH + 28
	}
	x, y, w, h := tooltipRect(mx, my, unitTooltipWidth, tooltipH)
	drawTooltipBox(screen, x, y, w, h)

	iconX, iconY := x+10.0, y+14.0
	textX := iconX + iconW + 12

	vector.FillRect(screen, float32(iconX), float32(iconY), float32(iconW), float32(iconH), color.RGBA{252, 252, 252, 242}, false)
	vector.StrokeRect(screen, float32(iconX), float32(iconY), float32(iconW), float32(iconH), 1, color.RGBA{160, 160, 160, 225}, false)

	DrawText(screen, utype.NameTR, textX, y+12, FaceMed, ColorGold)
	DrawText(screen, fmt.Sprintf("Birlik adedi: %d", unitCount), textX, y+38, FaceSmall, ColorWhite)
	DrawText(screen, fmt.Sprintf("Bakım: %d tahıl + %d altın/tur", utype.GrainUpkeep, utype.GoldUpkeep), textX, y+56, FaceSmall, ColorGray)

	statY := y + 80
	DrawText(screen, fmt.Sprintf("Saldırı: %d", utype.Attack), textX, statY, FaceSmall, ColorGray)
	statY += 16
	DrawText(screen, fmt.Sprintf("Savunma: %d", utype.Defense), textX, statY, FaceSmall, ColorGray)
	statY += 16
	DrawText(screen, fmt.Sprintf("Moral: %d", utype.Morale), textX, statY, FaceSmall, ColorGray)
	statY += 16
	currentHP := unit.CurrentHP
	if currentHP < 0 {
		currentHP = 0
	}
	if currentHP > army.MaxUnitHP {
		currentHP = army.MaxUnitHP
	}
	DrawText(screen, fmt.Sprintf("Can: %d", currentHP), textX, statY, FaceSmall, ColorGray)

	if sprite := unitSpriteForFaction(gs, a.OwnerID, unit.TypeID); sprite != nil {
		drawUnitSpriteCard(screen, sprite, float32(iconX), float32(iconY), float32(iconW), [3]float32{1, 1, 1})
	}
}

func unitAvailabilityStatus(gs *state.GameState, utype *army.UnitType, reqMissing bool) (string, color.RGBA) {
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

	for _, requiredTech := range utype.RequiredTech {
		name := requiredTech
		if t := gs.TechTypes[requiredTech]; t != nil {
			name = t.NameTR
		}
		done := ff != nil && ff.Research.Completed[requiredTech]
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
	return unitCostTooltipLines(gs, unitCost(utype))
}

func unitCostTooltipLines(gs *state.GameState, cost economy.ResourceCost) []tooltipLine {
	f := gs.Factions[gs.PlayerFactionID]
	lines := make([]tooltipLine, 0, 5)

	for _, kind := range economy.CostResourceKinds() {
		need := cost.Amount(kind)
		if need <= 0 {
			continue
		}

		col := ColorWhite
		text := fmt.Sprintf("%s: %d", economy.ResourceNameTR(kind), need)
		have := economy.FactionResourceAmount(f, kind)
		if have < need {
			col = ColorRed
			text += " eksik"
		}
		lines = append(lines, tooltipLine{text: text, col: col})
	}

	if len(lines) == 0 {
		return []tooltipLine{{text: "Bedava", col: ColorWhite}}
	}
	return lines
}

func resourceTooltipLines(gs *state.GameState, cost economy.ResourceCost) []tooltipLine {
	f := gs.Factions[gs.PlayerFactionID]
	lines := make([]tooltipLine, 0, 5)

	appendLine := func(kind economy.ResourceKind, need int, have int) {
		if need <= 0 {
			return
		}
		col := ColorWhite
		text := fmt.Sprintf("%s: %d/%d", economy.ResourceNameTR(kind), have, need)
		if have < need {
			col = ColorRed
			text += " eksik"
		}
		lines = append(lines, tooltipLine{text: text, col: col})
	}

	for _, kind := range economy.CostResourceKinds() {
		have := 0
		if f != nil {
			have = economy.FactionResourceAmount(f, kind)
		}
		appendLine(kind, cost.Amount(kind), have)
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
	gameui.DrawTooltip(screen, tooltip, hoverTooltipStyle, renderText)
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
