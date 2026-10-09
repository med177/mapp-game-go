package render

import (
	"fmt"
	"image/color"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/diplomacy"
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
	unitTooltipWidth         = 400.0
	armySummaryTileWidth     = 60.0
	armySummaryTileHeight    = 120.0
	armySummaryTileGap       = 6.0
	armySummaryTooltipPad    = 10.0
	armySummaryColumns       = 4
)

type unitTooltipLayout struct {
	upkeepY           float64
	movementY         float64
	costY             float64
	costLinesY        float64
	attributesY       float64
	requirementY      float64
	requirementLinesY float64
	height            float64
}

func newUnitTooltipLayout(iconH float64, costLineCount, requirementLineCount int) unitTooltipLayout {
	const (
		iconY         = 14.0
		lineH         = 14.0
		contentBottom = 12.0
	)

	layout := unitTooltipLayout{
		upkeepY:     50,
		movementY:   68,
		costY:       88,
		costLinesY:  102,
		attributesY: 88,
	}
	// Gereksinimler görselin sağ kolonu yerine görselin altında popup'ın
	// tamamını kullanır. Maliyet bloğu beklenmedik şekilde büyürse de
	// gereksinim başlığıyla çakışmaması için alt sınırı koru.
	layout.requirementY = iconY + iconH + 10
	costBottom := layout.costLinesY + float64(costLineCount)*lineH
	if costBottom+8 > layout.requirementY {
		layout.requirementY = costBottom + 8
	}
	layout.requirementLinesY = layout.requirementY + lineH
	layout.height = layout.requirementLinesY + float64(requirementLineCount)*lineH + contentBottom
	return layout
}

func unitTooltipImageMetrics(sprite *ebiten.Image) (width, height float64) {
	height = float64(unitSpriteHeight(recruitCardW)) + unitTooltipImageExtraH
	if sprite != nil {
		bounds := sprite.Bounds()
		if bounds.Dx() > 0 && bounds.Dy() > 0 {
			width = height * float64(bounds.Dx()) / float64(bounds.Dy())
			return width, height
		}
	}
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
				drawArmyUnitTooltip(screen, gs, a, unit, unitCount, playerCanSeeArmyDetails(gs, a), fx, fy)
				return
			}
		}
	}
	if goldRect, ok := regionGoldProductionRect(gs, rid); ok && goldRect.Hit(fx, fy) {
		drawRegionGoldTooltip(screen, gs, rid, fx, fy)
		return
	}
	if deltaRect, ok := regionSatisfactionDeltaRect(gs, rid); ok && deltaRect.Hit(fx, fy) {
		drawSatisfactionTooltip(screen, gs, rid, fx, fy)
		return
	}
	if logisticsRect, ok := regionPanelLogisticsRect(gs, rid); ok && logisticsRect.Hit(fx, fy) {
		drawRegionLogisticsTooltip(screen, gs, rid, fx, fy)
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
	if regionVassalizeSuccessorButtonHitForTab(fx, fy, gs, rid, activeTab) {
		successorID, _ := regionLiberationSuccessor(gs, gs.Regions[rid])
		successorName := factionDisplayName(gs, string(successorID))
		if successorName == "" {
			successorName = string(successorID)
		}
		drawSmallHoverHint(screen, successorName+" devletini vassallaştır", fx, fy)
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
	if regionOfferPrivilegeButtonHitForTab(fx, fy, gs, rid, activeTab) {
		drawSmallHoverHint(screen, "Bir devlete imtiyaz teklif et", fx, fy)
		return
	}
	if regionRevokePrivilegeButtonHitForTab(fx, fy, gs, rid, activeTab) {
		drawSmallHoverHint(screen, fmt.Sprintf("İmtiyazı kaldır, ilişki -%d", diplomacy.PrivilegeRevocationRelationPenalty), fx, fy)
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

// drawArmyMarkerHoverTooltip, detay paneli kapalıyken harita marker'ının
// birim bileşimini gösterir. Görünürlük kontrolü, yabancı orduların gizli
// istihbaratını popup üzerinden açığa çıkarmamasını sağlar.
func (r *Renderer) drawArmyMarkerHoverTooltip(screen *ebiten.Image) {
	if r == nil || r.gs == nil || r.mapMode == MapModeTrade || r.worldInputLockedByPhase() {
		return
	}
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	if _, ok := r.uiLayers.TopAt(fx, fy); ok {
		return
	}
	aid, ok := r.armyHitAt(fx, fy)
	if !ok {
		return
	}
	a := r.gs.Armies[aid]
	if a == nil {
		return
	}
	if !playerCanSeeArmyDetails(r.gs, a) && !enemyArmyInPlayerMoveRange(r.gs, a) &&
		!enemyUnderPlayerSiege(r.gs, a) && !playerHasRevealEnemyStrength(r.gs) {
		drawHiddenArmyMarkerTooltip(screen, a, fx, fy)
		return
	}
	drawArmyMarkerUnitSummaryTooltip(screen, r.gs, a, fx, fy)
}

func drawHiddenArmyMarkerTooltip(screen *ebiten.Image, a *army.Army, mx, my float64) {
	if a == nil {
		return
	}
	const tooltipWidth = 260.0
	const tooltipHeight = 54.0
	x, y, w, h := tooltipRect(mx, my, tooltipWidth, tooltipHeight)
	drawTooltipBox(screen, x, y, w, h)
	label := "Ordu"
	if a.IsNaval {
		label = "Filo"
	}
	DrawText(screen, label+": birim detayları gizli", x+10, y+12, FaceSmall, ColorGold)
	DrawText(screen, "İstihbarat yetersiz", x+10, y+31, FaceSmall, ColorGray)
}

func armyMarkerSummaryRows(gs *state.GameState, a *army.Army) []armyMergePreviewRow {
	if gs == nil || a == nil {
		return nil
	}
	units := a.Units
	if !playerCanSeeArmyDetails(gs, a) {
		fullIntel := playerHasRevealEnemyStrength(gs)
		revealRatio := 0.50
		if enemyUnderPlayerSiege(gs, a) {
			revealRatio = 0.75
		}
		revealed := scoutedEnemyRevealCount(len(a.Units), fullIntel, revealRatio)
		units = make([]army.Unit, 0, revealed)
		for displayIndex := 0; displayIndex < revealed; displayIndex++ {
			unitIndex := armyPanelUnitIndex(a.Units, gs.UnitTypes, displayIndex)
			if unitIndex < 0 {
				break
			}
			units = append(units, a.Units[unitIndex])
		}
	}
	rows := make([]armyMergePreviewRow, 0, len(units))
	for _, unit := range units {
		rowIndex := -1
		for i := range rows {
			if rows[i].typeID == unit.TypeID {
				rowIndex = i
				break
			}
		}
		if rowIndex < 0 {
			rows = append(rows, armyMergePreviewRow{typeID: unit.TypeID, count: 1})
		} else {
			rows[rowIndex].count++
		}
	}
	return rows
}

func drawArmyMarkerUnitSummaryTooltip(screen *ebiten.Image, gs *state.GameState, a *army.Army, mx, my float64) {
	if gs == nil || a == nil {
		return
	}
	rows := armyMarkerSummaryRows(gs, a)
	rowLines := armySummaryRowCount(len(rows))
	tooltipHeight := 40.0 + armySummaryGridHeight(rowLines) + armySummaryTooltipPad
	if len(rows) == 0 {
		tooltipHeight = 54
	}
	x, y, w, h := tooltipRect(mx, my, armySummaryTooltipWidth(), tooltipHeight)
	drawTooltipBox(screen, x, y, w, h)
	label := "Ordu"
	if a.IsNaval {
		label = "Filo"
	}
	header := label + ": " + itoa(len(a.Units)) + " birim"
	if !playerCanSeeArmyDetails(gs, a) && !playerHasRevealEnemyStrength(gs) {
		header = label + ": kısmi istihbarat"
	}
	DrawText(screen, header, x+10, y+10, FaceSmall, ColorGold)

	drawArmyUnitSummaryTiles(screen, gs, a.OwnerID, rows, x+armySummaryTooltipPad, y+32)
}

func armySummaryTooltipWidth() float64 {
	return armySummaryTooltipPad*2 + armySummaryColumns*armySummaryTileWidth + (armySummaryColumns-1)*armySummaryTileGap
}

func armySummaryRowCount(itemCount int) int {
	if itemCount <= 0 {
		return 0
	}
	return (itemCount + armySummaryColumns - 1) / armySummaryColumns
}

func armySummaryGridHeight(rowCount int) float64 {
	if rowCount <= 0 {
		return 0
	}
	return float64(rowCount)*armySummaryTileHeight + float64(rowCount-1)*armySummaryTileGap
}

// drawArmyUnitSummaryTiles, marker ve birleştirme popup'larının ortak kart
// geometrisini ve görsel katmanlarını üretir.
func drawArmyUnitSummaryTiles(screen *ebiten.Image, gs *state.GameState, ownerID string, rows []armyMergePreviewRow, startX, startY float64) {
	for index, row := range rows {
		column := index % armySummaryColumns
		line := index / armySummaryColumns
		tileX := startX + float64(column)*(armySummaryTileWidth+armySummaryTileGap)
		tileY := startY + float64(line)*(armySummaryTileHeight+armySummaryTileGap)
		vector.FillRect(screen, float32(tileX), float32(tileY), float32(armySummaryTileWidth), float32(armySummaryTileHeight), color.RGBA{248, 246, 238, 235}, false)
		if sprite := unitSpriteForFaction(gs, ownerID, row.typeID); sprite != nil {
			drawArmyMarkerUnitImage(screen, sprite, float32(tileX), float32(tileY), float32(armySummaryTileWidth), float32(armySummaryTileHeight))
		}
		countY := tileY + armySummaryTileHeight - 26
		vector.FillRect(screen, float32(tileX), float32(countY), float32(armySummaryTileWidth), 26, color.RGBA{25, 20, 15, 175}, false)
		vector.StrokeRect(screen, float32(tileX), float32(countY), float32(armySummaryTileWidth), 26, 1, color.RGBA{190, 160, 90, 190}, false)
		drawUIOutlinedLabel(screen,
			gameui.Rect{X: tileX, Y: countY + 5, W: armySummaryTileWidth, H: 18},
			"x"+itoa(row.count), color.White, color.Black, gameui.TextMedium, gameui.TextAlignCenter)
		vector.StrokeRect(screen, float32(tileX), float32(tileY), float32(armySummaryTileWidth), float32(armySummaryTileHeight), 1, color.RGBA{150, 125, 72, 220}, false)
	}
}

// drawArmyMarkerUnitImage görseli kartın tamamına yayar; marker popup'ındaki
// kartlar kaynak görselin dikey oranına göre boşluk bırakmaz.
func drawArmyMarkerUnitImage(screen *ebiten.Image, sprite *ebiten.Image, x, y, width, height float32) bool {
	if screen == nil || sprite == nil || width <= 0 || height <= 0 {
		return false
	}
	bounds := sprite.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return false
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(width)/float64(bounds.Dx()), float64(height)/float64(bounds.Dy()))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(sprite, op)
	return true
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

	summaryRows := rows[:rowCount]
	rowLines := armySummaryRowCount(rowCount)
	previewHeight := 58.0 + armySummaryGridHeight(rowLines) + armySummaryTooltipPad
	x, y, w, h := tooltipRect(mx, my, armySummaryTooltipWidth(), previewHeight)
	drawTooltipBox(screen, x, y, w, h)

	DrawText(screen, "Hedef ordu: "+itoa(len(target.Units))+" birim", x+10, y+10, FaceSmall, ColorGold)
	DrawText(screen, "Birleşince: "+itoa(mergeResultUnitCount(source, target)), x+10, y+27, FaceSmall, ColorWhite)
	drawArmyUnitSummaryTiles(screen, gs, target.OwnerID, summaryRows, x+armySummaryTooltipPad, y+46)
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

	costLines := buildingCostRequirementLines(gs, region, b)
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
	if configuredMax := gs.BuildingLevelCap(region, b.ID); configuredMax > 0 {
		maxLevel = configuredMax
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
	if !buildingCost(gs, region, b).CanAfford(gs.Factions[gs.PlayerFactionID]) {
		return "Kaynak yetersiz", ColorRed
	}
	return "İnşa edilebilir", color.RGBA{120, 210, 120, 230}
}

func buildingCost(gs *state.GameState, region *world.Region, b *city.Building) economy.ResourceCost {
	if b == nil {
		return economy.ResourceCost{}
	}
	targetLevel := 1
	if region != nil {
		targetLevel = region.BuildingLevel(b.ID) + 1
		for _, order := range gs.ProductionQueue {
			if order.Kind == "building" && order.RegionID == region.ID && order.TypeID == b.ID {
				targetLevel++
			}
		}
	}
	return economy.BuildingCostAtLevel(b, targetLevel)
}

func buildingCostRequirementLines(gs *state.GameState, region *world.Region, b *city.Building) []tooltipLine {
	return resourceTooltipLines(gs, buildingCost(gs, region, b))
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
	if b.GoldMaintenance > 0 {
		lines = append(lines, fmt.Sprintf("Bina bakımı: %d altın/tur", b.GoldMaintenance))
	}
	if b.GoldMod != 1 {
		lines = append(lines, fmt.Sprintf("Altın geliri: x%.2f", b.GoldMod))
	}
	if b.GrainMod != 1 {
		lines = append(lines, fmt.Sprintf("Tahıl üretimi: x%.2f", b.GrainMod))
	}
	if b.GrainBonus != 0 {
		lines = append(lines, fmt.Sprintf("Tahıl üretimi: %+d", b.GrainBonus))
	}
	if b.IronMod != 1 {
		lines = append(lines, fmt.Sprintf("Demir üretimi: x%.2f", b.IronMod))
	}
	if b.IronBonus != 0 {
		lines = append(lines, fmt.Sprintf("Demir üretimi: %+d", b.IronBonus))
	}
	if b.TimberMod != 1 {
		lines = append(lines, fmt.Sprintf("Kereste üretimi: x%.2f", b.TimberMod))
	}
	if b.TimberBonus != 0 {
		lines = append(lines, fmt.Sprintf("Kereste üretimi: %+d", b.TimberBonus))
	}
	if b.StoneMod != 1 {
		lines = append(lines, fmt.Sprintf("Taş üretimi: x%.2f", b.StoneMod))
	}
	if b.StoneBonus != 0 {
		lines = append(lines, fmt.Sprintf("Taş üretimi: %+d", b.StoneBonus))
	}
	if b.SpiceMod != 1 {
		lines = append(lines, fmt.Sprintf("Baharat üretimi: x%.2f", b.SpiceMod))
	}
	if b.SpiceBonus != 0 {
		lines = append(lines, fmt.Sprintf("Baharat üretimi: %+d", b.SpiceBonus))
	}
	if b.ClothMod != 1 {
		lines = append(lines, fmt.Sprintf("Kumaş üretimi: x%.2f", b.ClothMod))
	}
	if b.ClothBonus != 0 {
		lines = append(lines, fmt.Sprintf("Kumaş üretimi: %+d", b.ClothBonus))
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
	maxLevel := gs.BuildingLevelCap(region, b.ID)
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
		fmt.Sprintf("Donanma sınırı: %d -> %d", currentCap, nextCap),
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
	maxLevel := gs.BuildingLevelCap(region, b.ID)
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
				fmt.Sprintf("Savaşçı sınırı: %d", landCap),
				fmt.Sprintf("Temel ordu: %d × %d", baseArmyCap, army.MaxArmySize),
				fmt.Sprintf("Kışla üretim limiti: %d -> %d birim/tur", productionLimit, nextProductionLimit),
			}
		}
	}
	return []string{
		fmt.Sprintf("Savaşçı sınırı: %d", landCap),
		fmt.Sprintf("Temel ordu: %d × %d; +1 slot ayrı", baseArmyCap, army.MaxArmySize),
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
	iconW, iconH := unitTooltipImageMetrics(sprite)
	layout := newUnitTooltipLayout(iconH, len(costLines), len(reqLines))
	tooltipH := layout.height
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

	DrawText(screen, fmt.Sprintf("Bakım: %d tahıl + %d altın/tur", utype.GrainUpkeep, utype.GoldUpkeep), textX, y+layout.upkeepY, FaceSmall, ColorGray)
	DrawText(screen, fmt.Sprintf("Hareket: %d PU", utype.BaseMovementPoints()), textX, y+layout.movementY, FaceSmall, ColorGray)

	DrawText(screen, "Maliyet:", textX, y+layout.costY, FaceSmall, ColorGray)
	drawUIRichTextBlock(screen, gameui.Rect{X: textX, Y: y + layout.costLinesY, W: w - (textX - x) - 10}, tooltipRichLines(costLines), 14)

	attributeX := x + w - 145
	DrawText(screen, "Nitelik:", attributeX, y+layout.attributesY, FaceSmall, ColorGray)
	DrawText(screen, fmt.Sprintf("Saldırı: %d", utype.Attack), attributeX, y+layout.attributesY+14, FaceSmall, ColorGray)
	DrawText(screen, fmt.Sprintf("Savunma: %d", utype.Defense), attributeX, y+layout.attributesY+30, FaceSmall, ColorGray)
	DrawText(screen, fmt.Sprintf("Moral: %d", utype.Morale), attributeX, y+layout.attributesY+46, FaceSmall, ColorGray)
	DrawText(screen, fmt.Sprintf("Can: %d", utype.HP), attributeX, y+layout.attributesY+62, FaceSmall, ColorGray)

	// Gereksinimler görselin sağ kolonu yerine görselin altında popup'ın
	// tamamını kullanır; uzun teknoloji adları artık sağ kenardan taşmaz.
	requirementX := x + 10
	DrawText(screen, "Gereksinim:", requirementX, y+layout.requirementY, FaceSmall, ColorGray)
	drawUIRichTextBlock(screen, gameui.Rect{X: requirementX, Y: y + layout.requirementLinesY, W: w - 20}, tooltipRichLines(reqLines), 14)

	if sprite != nil {
		drawUnitSpriteCard(screen, sprite, float32(iconX), float32(iconY), float32(iconW), [3]float32{1, 1, 1})
	}
}

// drawArmyUnitTooltip, recruit tooltip'ından ayrı olarak seçili ordudaki
// gerçek birim örneğinin bilgilerini gösterir. Üretim maliyetleri ve
// gereksinimler bu bağlamda anlamlı olmadığı için yalnız adet, tur başı bakım,
// savaş değerleri ve anlık can çizilir.
func drawArmyUnitTooltip(screen *ebiten.Image, gs *state.GameState, a *army.Army, unit army.Unit, unitCount int, showCurrentHP bool, mx, my float64) {
	if gs == nil || a == nil {
		return
	}
	utype := gs.UnitTypes[unit.TypeID]
	if utype == nil {
		return
	}

	ensureArmySprites()
	sprite := unitSpriteForFaction(gs, a.OwnerID, unit.TypeID)
	tooltipH := 190.0
	iconW, iconH := unitTooltipImageMetrics(sprite)
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
	DrawText(screen, fmt.Sprintf("Hareket: %d PU", utype.BaseMovementPoints()), textX, y+74, FaceSmall, ColorGray)

	statY := y + 98
	DrawText(screen, fmt.Sprintf("Saldırı: %d", utype.Attack), textX, statY, FaceSmall, ColorGray)
	statY += 16
	DrawText(screen, fmt.Sprintf("Savunma: %d", utype.Defense), textX, statY, FaceSmall, ColorGray)
	statY += 16
	DrawText(screen, fmt.Sprintf("Moral: %d", utype.Morale), textX, statY, FaceSmall, ColorGray)
	statY += 16
	if showCurrentHP {
		currentHP := unit.CurrentHP
		if currentHP < 0 {
			currentHP = 0
		}
		if currentHP > army.MaxUnitHP {
			currentHP = army.MaxUnitHP
		}
		DrawText(screen, fmt.Sprintf("Can: %d", currentHP), textX, statY, FaceSmall, ColorGray)
	} else {
		DrawText(screen, "Can: ~", textX, statY, FaceSmall, ColorGray)
	}

	if sprite != nil {
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

	buildingLevels := region.BuildingLevels()
	for _, requirement := range utype.BuildingRequirements() {
		currentLevel := buildingLevels[requirement.ID]
		name := requirement.ID
		if b := gs.BuildingTypes[requirement.ID]; b != nil {
			name = b.NameTR
		}
		col := color.RGBA{170, 145, 90, 230}
		if currentLevel < requirement.Level {
			col = ColorRed
			missing = true
		}
		lines = append(lines, tooltipLine{
			text: fmt.Sprintf("%s Lv%d gerekli (mevcut: Lv%d)", name, requirement.Level, currentLevel),
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
	lines := make([]tooltipLine, 0, 5)

	appendLine := func(kind economy.ResourceKind, need int) {
		if need <= 0 {
			return
		}
		col := ColorWhite
		text := fmt.Sprintf("%s: %d", economy.ResourceNameTR(kind), need)
		if gs != nil {
			f := gs.Factions[gs.PlayerFactionID]
			have := 0
			if f != nil {
				have = economy.FactionResourceAmount(f, kind)
			}
			if have < need {
				col = ColorRed
				text += " eksik"
			}
		}
		lines = append(lines, tooltipLine{text: text, col: col})
	}

	for _, kind := range economy.CostResourceKinds() {
		appendLine(kind, cost.Amount(kind))
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

func drawRegionGoldTooltip(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, mx, my float64) {
	region := gs.Regions[rid]
	if region == nil {
		return
	}
	lines, total := gs.RegionGoldIncomeBreakdown(region)
	const tooltipWidth = 330.0
	// Son katkı satırı ile toplam ayırıcısı arasında yeterli görsel boşluk bırak.
	tooltipHeight := 60.0 + float64(len(lines))*20
	x, y, w, h := tooltipRect(mx, my, tooltipWidth, tooltipHeight)
	drawTooltipBox(screen, x, y, w, h)
	DrawText(screen, "Altın üretimi / tur", x+10, y+10, FaceSmall, ColorGold)
	DrawText(screen, formatSignedAmount(total), x+w-10-MeasureText(formatSignedAmount(total), FaceSmall), y+10, FaceSmall, ColorGold)
	drawUISeparator(screen, float32(x+10), float32(y+30), float32(x+w-10), 1, panelBorder)

	rowY := y + 38
	for _, line := range lines {
		lineColor := color.RGBA{145, 220, 155, 255}
		if line.Value < 0 {
			lineColor = ColorRed
		}
		drawUIKeyValueRowWithGap(screen, x+10, rowY, w-20, line.Label, formatSignedAmount(line.Value), ColorGray, lineColor, 8)
		rowY += 20
	}
	drawUISeparator(screen, float32(x+10), float32(y+h-25), float32(x+w-10), 1, panelBorder)
	drawUIKeyValueRowWithGap(screen, x+10, y+h-19, w-20, "Toplam", formatSignedAmount(total), ColorGray, ColorGold, 8)
}

func drawRegionLogisticsTooltip(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, mx, my float64) {
	region := gs.Regions[rid]
	if region == nil {
		return
	}
	status, ok := regionPanelLogisticsStatus(gs, region)
	if !ok {
		return
	}
	type logisticsLine struct {
		label string
		value string
		col   color.RGBA
	}
	lines := []logisticsLine{
		{"Yerel üretim sonrası", "+" + itoa(status.LocalProduction), color.RGBA{145, 220, 155, 255}},
		{"Yerleşim/ticaret tamponu", "+" + itoa(status.SettlementBuffer), ColorGray},
		{"Ambar desteği", "+" + itoa(status.GranarySupport), ColorGray},
		{"Merkez rezerv desteği", "+" + itoa(status.ReserveSupport), ColorGray},
		{"Filo ikmali", "+" + itoa(status.NavalSupplyGrainSpent), color.RGBA{125, 190, 230, 255}},
		{"Toplam kapasite", itoa(status.Capacity), ColorGold},
		{"Ordu talebi", itoa(status.Demand), ColorWhite},
	}
	shortageLabel := "İkmal açığı"
	shortageValue := itoa(status.Overload)
	shortageColor := ColorRed
	if status.Overload <= 0 {
		shortageLabel = "Fazla kapasite"
		shortageValue = "+" + itoa(status.Capacity-status.Demand)
		shortageColor = color.RGBA{145, 220, 155, 255}
	}
	lines = append(lines, logisticsLine{shortageLabel, shortageValue, shortageColor})

	const tooltipWidth = 360.0
	tooltipHeight := 60.0 + float64(len(lines))*20
	x, y, w, h := tooltipRect(mx, my, tooltipWidth, tooltipHeight)
	drawTooltipBox(screen, x, y, w, h)
	DrawText(screen, "Bölgesel ikmal hesabı", x+10, y+10, FaceSmall, ColorGold)
	drawUISeparator(screen, float32(x+10), float32(y+30), float32(x+w-10), 1, panelBorder)
	rowY := y + 38
	for _, line := range lines {
		drawUIKeyValueRowWithGap(screen, x+10, rowY, w-20, line.label, line.value, ColorGray, line.col, 8)
		rowY += 20
	}
	drawUISeparator(screen, float32(x+10), float32(y+h-25), float32(x+w-10), 1, panelBorder)
	drawUIKeyValueRowWithGap(screen, x+10, y+h-19, w-20, "Son çözüm zayiatı", itoa(status.TotalHPDamage)+" HP", ColorGray, color.RGBA{225, 135, 100, 255}, 8)
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
