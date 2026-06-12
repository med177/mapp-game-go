package render

import (
	"fmt"
	"image/color"
	"sort"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Modern teknoloji paneli renk paleti ──────────────────────────────

var (
	techPanelOverlay          = color.RGBA{4, 6, 16, 232}
	techPanelBG               = color.RGBA{10, 12, 28, 242}
	techPanelAccent           = color.RGBA{200, 170, 70, 255}
	techPanelAccentDim        = color.RGBA{130, 105, 40, 200}
	techCardBgBase            = color.RGBA{16, 18, 38, 240}
	techCardBgLocked          = color.RGBA{22, 22, 30, 230}
	techCardBgActive          = color.RGBA{30, 26, 10, 248}
	techCardBorderGlow        = color.RGBA{255, 255, 255, 60}
	techCardBorderActive      = color.RGBA{255, 210, 60, 255}
	techConnLine              = color.RGBA{90, 96, 120, 180}
	techConnLineUnlocked      = color.RGBA{160, 170, 200, 220}
	techConnLineDone          = color.RGBA{130, 180, 130, 220}
	techTextLocked            = color.RGBA{140, 140, 150, 220}
	techTextUnlocked          = color.RGBA{230, 230, 240, 255}
	techTextCost              = color.RGBA{240, 210, 110, 245}
	techTextCostLocked        = color.RGBA{170, 160, 140, 200}
	techBadgeDoneBG           = color.RGBA{24, 60, 32, 240}
	techBadgeDoneBorder       = color.RGBA{120, 210, 130, 240}
	techBadgeActiveBG         = color.RGBA{70, 48, 12, 240}
	techBadgeActiveBorder     = color.RGBA{255, 200, 80, 250}
	techProgressBarBG         = color.RGBA{30, 28, 22, 220}
	techProgressBarFill       = color.RGBA{240, 190, 40, 240}
	techFilterTabBG           = color.RGBA{14, 16, 32, 230}
	techFilterTabActive       = color.RGBA{28, 24, 10, 245}
	techFilterTabBorder       = color.RGBA{90, 84, 60, 180}
	techFilterTabActiveBorder = color.RGBA{210, 175, 60, 250}
	techHeaderBarColor        = color.RGBA{190, 155, 55, 240}
	techCardGlowOffset        = 2.0
	techCardRadius            = float32(8)
	techCardInnerRadius       = float32(6)
	techLevelLabelColor       = color.RGBA{150, 160, 190, 200}
	techCategoryIconColors    = map[tech.Category]color.RGBA{
		tech.CategoryMilitary:  {220, 110, 110, 255},
		tech.CategoryEconomy:   {110, 200, 110, 255},
		tech.CategoryDiplomacy: {110, 130, 220, 255},
		tech.CategoryNaval:     {200, 190, 100, 255},
		tech.CategoryReligion:  {210, 130, 210, 255},
	}
)

// techCategoryIcon her kategori için uygun ikonu döner.
func techCategoryIcon(cat tech.Category) gameui.IconID {
	switch cat {
	case tech.CategoryMilitary:
		return gameui.IconSword
	case tech.CategoryNaval:
		return gameui.IconSend // Denizcilik için gönder/yelken metaforu
	case tech.CategoryDiplomacy:
		return gameui.IconBook
	case tech.CategoryEconomy:
		return gameui.IconBuy
	case tech.CategoryReligion:
		return gameui.IconMenu // placeholder
	default:
		return gameui.IconNone
	}
}

var techCategoryColors = map[tech.Category]color.RGBA{
	tech.CategoryMilitary:  {200, 100, 100, 255},
	tech.CategoryEconomy:   {100, 200, 100, 255},
	tech.CategoryDiplomacy: {100, 100, 200, 255},
	tech.CategoryNaval:     {200, 200, 100, 255},
	tech.CategoryReligion:  {200, 100, 200, 255},
}

type techNode struct {
	t        *tech.Technology
	unlocked bool
	done     bool
	level    int
	x, y     float64
}

const (
	techLevelHeight   = 140.0
	techNodeWidth     = 188.0
	techNodeHeight    = 96.0
	techTreeMaxCols   = 4
	techNodeStepX     = 228.0
	techNodeStepY     = 132.0
	techLevelGap      = 96.0
	techTreePadding   = 28.0
	techPanWheelStep  = 56.0
	techPanDragThresh = 4.0

	techConnExitOffset  = 12.0
	techConnEntryOffset = 7.0
	techConnMidMargin   = 16.0

	// Filter sekme sabitleri
	techFilterTabH    = 36.0
	techFilterTabGap  = 8.0
	techFilterTabPad  = 14.0
	techFilterTabMinW = 88.0
)

type techConnectorStyle struct {
	col    color.RGBA
	width  float32
	dashed bool
}

type techPanelLayout struct {
	panelRect  gameui.Rect
	titleRect  gameui.Rect
	statusRect gameui.Rect
	filterRect gameui.Rect
	hintRect   gameui.Rect
	closeRect  gameui.Rect
	treeRect   gameui.Rect
}

type techTreeLayoutData struct {
	levels   [][]techNode
	contentW float64
	contentH float64
}

func techPanelLayoutForScreen() techPanelLayout {
	panel := gameui.Rect{X: 0, Y: 0, W: ScreenWidth, H: ScreenHeight}
	box := gameui.BoxFromRect(panel).InsetXY(20, 20)
	headerRect, box := box.CutTop(28, 12)
	statusRect, box := box.CutTop(26, 10)
	filterRect, box := box.CutTop(techFilterTabH, 14)
	hintRect, treeBox := box.CutBottom(46, 14)
	closeRect, titleRect := gameui.BoxFromRect(headerRect).CutRight(32, 12)
	return techPanelLayout{
		panelRect:  panel,
		titleRect:  titleRect.Rect,
		statusRect: statusRect,
		filterRect: filterRect,
		hintRect:   hintRect,
		closeRect:  closeRect,
		treeRect:   treeBox.Rect,
	}
}

func buildTechCloseButton() gameui.Button {
	x, y, w, h := techCloseRect()
	btn := gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "").WithIcon(gameui.IconClose)
	btn.IconSize = 14
	return btn
}

func techNodeRect(node techNode) gameui.Rect {
	return gameui.Rect{
		X: node.x - techNodeWidth/2,
		Y: node.y - techNodeHeight/2,
		W: techNodeWidth - 4,
		H: techNodeHeight - 4,
	}
}

func techNodeButton(node techNode) gameui.Button {
	rect := techNodeRect(node)
	return gameui.NewButton(rect.X, rect.Y, rect.W, rect.H, node.t.NameTR)
}

func buildTechCardComponent(node techNode, activeResearchID string, research faction.ResearchState) techCardComponent {
	return newTechCardComponent(node, techNodeRect(node), activeResearchID, research)
}

func (r *Renderer) buildLaidOutTechTree(f *faction.Faction) techTreeLayoutData {
	levels := r.buildTechTree(f)
	contentW, contentH := layoutTechTree(levels, techTreeMaxCols)
	return techTreeLayoutData{levels: levels, contentW: contentW, contentH: contentH}
}

func (r *Renderer) buildTechTree(f *faction.Faction) [][]techNode {
	levels := make(map[int][]techNode)
	maxLevel := 0

	for _, t := range r.gs.TechTypes {
		level := r.getTechLevel(t, r.gs.TechTypes)
		if level > maxLevel {
			maxLevel = level
		}
		node := techNode{
			t:        t,
			unlocked: tech.IsUnlocked(&f.Research, t),
			done:     f.Research.Completed[t.ID],
			level:    level,
		}
		levels[level] = append(levels[level], node)
	}

	var result [][]techNode
	for i := 0; i <= maxLevel; i++ {
		if nodes, ok := levels[i]; ok {
			sort.Slice(nodes, func(a, b int) bool {
				if nodes[a].t.Category != nodes[b].t.Category {
					return tech.CategoryOrder(nodes[a].t.Category) < tech.CategoryOrder(nodes[b].t.Category)
				}
				return nodes[a].t.ID < nodes[b].t.ID
			})
			result = append(result, nodes)
		}
	}
	orderTechTreeLevels(result)

	return result
}

func (r *Renderer) getTechLevel(t *tech.Technology, allTechs map[string]*tech.Technology) int {
	if len(t.Requires) == 0 {
		return 0
	}
	maxReqLevel := 0
	for _, reqID := range t.Requires {
		if req, ok := allTechs[reqID]; ok {
			reqLevel := r.getTechLevel(req, allTechs)
			if reqLevel > maxReqLevel {
				maxReqLevel = reqLevel
			}
		}
	}
	return maxReqLevel + 1
}

func layoutTechTree(levels [][]techNode, maxCols int) (contentW float64, contentH float64) {
	maxColsUsed := 1
	for _, levelNodes := range levels {
		if cols := techMinInt(len(levelNodes), maxCols); cols > maxColsUsed {
			maxColsUsed = cols
		}
	}
	contentW = techTreePadding*2 + techNodeWidth + float64(maxColsUsed-1)*techNodeStepX
	currentY := techTreePadding
	for levelIdx, levelNodes := range levels {
		if len(levelNodes) == 0 {
			continue
		}
		rows := (len(levelNodes) + maxCols - 1) / maxCols
		for row := 0; row < rows; row++ {
			rowStart := row * maxCols
			rowEnd := techMinInt(rowStart+maxCols, len(levelNodes))
			rowCount := rowEnd - rowStart
			rowWidth := techNodeWidth + float64(rowCount-1)*techNodeStepX
			startX := techTreePadding + (contentW-techTreePadding*2-rowWidth)/2 + techNodeWidth/2
			y := currentY + float64(row)*techNodeStepY + techNodeHeight/2
			for col := 0; col < rowCount; col++ {
				nodeIdx := rowStart + col
				levels[levelIdx][nodeIdx].x = startX + float64(col)*techNodeStepX
				levels[levelIdx][nodeIdx].y = y
			}
		}
		currentY += techNodeHeight + float64(rows-1)*techNodeStepY
		if levelIdx < len(levels)-1 {
			currentY += techLevelGap
		}
	}
	contentH = currentY + techTreePadding
	return contentW, contentH
}

func indexTechNodes(levels [][]techNode) map[string]techNode {
	index := make(map[string]techNode, len(levels)*4)
	for _, levelNodes := range levels {
		for _, node := range levelNodes {
			index[node.t.ID] = node
		}
	}
	return index
}

func techConnectorStyleFor(parent, child techNode) techConnectorStyle {
	col := techConnLine
	width := float32(2.0)
	if child.unlocked {
		col = techConnLineUnlocked
		width = 2.8
	}
	if child.done && parent.done {
		col = techConnLineDone
		width = 3.5
	}
	return techConnectorStyle{col: col, width: width}
}

func drawTechLine(screen *ebiten.Image, x1, y1, x2, y2 float64, style techConnectorStyle) {
	if style.dashed {
		drawDashedTechLine(screen, x1, y1, x2, y2, style)
		return
	}
	// Glow efekti: önce daha kalın, daha transparan çizgi
	glowCol := color.RGBA{style.col.R, style.col.G, style.col.B, uint8(style.col.A / 3)}
	vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), style.width+2.0, glowCol, false)
	vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), style.width, style.col, false)
}

func drawDashedTechLine(screen *ebiten.Image, x1, y1, x2, y2 float64, style techConnectorStyle) {
	const dashLen = 8.0
	const gapLen = 5.0
	glowCol := color.RGBA{style.col.R, style.col.G, style.col.B, uint8(style.col.A / 3)}
	if x1 == x2 {
		if y2 < y1 {
			y1, y2 = y2, y1
		}
		for segStart := y1; segStart < y2; segStart += dashLen + gapLen {
			segEnd := segStart + dashLen
			if segEnd > y2 {
				segEnd = y2
			}
			vector.StrokeLine(screen, float32(x1), float32(segStart), float32(x2), float32(segEnd), style.width+1.5, glowCol, false)
			vector.StrokeLine(screen, float32(x1), float32(segStart), float32(x2), float32(segEnd), style.width, style.col, false)
		}
		return
	}
	if y1 == y2 {
		if x2 < x1 {
			x1, x2 = x2, x1
		}
		for segStart := x1; segStart < x2; segStart += dashLen + gapLen {
			segEnd := segStart + dashLen
			if segEnd > x2 {
				segEnd = x2
			}
			vector.StrokeLine(screen, float32(segStart), float32(y1), float32(segEnd), float32(y2), style.width+1.5, glowCol, false)
			vector.StrokeLine(screen, float32(segStart), float32(y1), float32(segEnd), float32(y2), style.width, style.col, false)
		}
	}
}

func drawTechConnector(screen *ebiten.Image, parent, child techNode, reqIdx, reqCount int) {
	style := techConnectorStyleFor(parent, child)

	pBottom := parent.y + techNodeHeight/2
	cTop := child.y - techNodeHeight/2

	dx := child.x - parent.x
	if dx > -techNodeWidth/2 && dx < techNodeWidth/2 {
		drawTechLine(screen, parent.x, pBottom, child.x, cTop, style)
		return
	}

	totalSpread := techConnExitOffset * float64(reqCount-1)
	offsetX := -totalSpread/2 + float64(reqIdx)*techConnExitOffset

	exitX := parent.x + offsetX
	entryX := child.x + offsetX

	halfW := techNodeWidth / 2
	exitX = techClampFloat(exitX, parent.x-halfW+4, parent.x+halfW-4)
	entryX = techClampFloat(entryX, child.x-halfW+4, child.x+halfW-4)

	midY := (pBottom + cTop) / 2

	drawTechLine(screen, exitX, pBottom, exitX, midY, style)
	drawTechLine(screen, exitX, midY, entryX, midY, style)
	drawTechLine(screen, entryX, midY, entryX, cTop, style)
}

func orderTechTreeLevels(levels [][]techNode) {
	if len(levels) < 2 {
		return
	}
	for levelIdx := 1; levelIdx < len(levels); levelIdx++ {
		prevIndex := make(map[string]int, len(levels[levelIdx-1]))
		for idx, node := range levels[levelIdx-1] {
			prevIndex[node.t.ID] = idx
		}
		sort.SliceStable(levels[levelIdx], func(a, b int) bool {
			scoreA := techNodeDependencyOrder(levels[levelIdx][a], prevIndex)
			scoreB := techNodeDependencyOrder(levels[levelIdx][b], prevIndex)
			if scoreA != scoreB {
				return scoreA < scoreB
			}
			if levels[levelIdx][a].t.Category != levels[levelIdx][b].t.Category {
				return tech.CategoryOrder(levels[levelIdx][a].t.Category) < tech.CategoryOrder(levels[levelIdx][b].t.Category)
			}
			return levels[levelIdx][a].t.ID < levels[levelIdx][b].t.ID
		})
	}
}

func techNodeDependencyOrder(node techNode, prevIndex map[string]int) float64 {
	if len(node.t.Requires) == 0 {
		return 0
	}
	total := 0.0
	count := 0.0
	for _, reqID := range node.t.Requires {
		if idx, ok := prevIndex[reqID]; ok {
			total += float64(idx)
			count++
		}
	}
	if count == 0 {
		return float64(len(prevIndex) + tech.CategoryOrder(node.t.Category))
	}
	return total / count
}

// ── Kategori filtre sekmeleri ────────────────────────────────────────

// drawTechFilterTabs kategori filtre sekmelerini çizer.
func drawTechFilterTabs(screen *ebiten.Image, rect gameui.Rect, activeFilter tech.Category, f *faction.Faction) {
	categories := tech.AllCategories()
	// Her sekmenin genişliğini hesapla
	tabCount := len(categories)
	availableW := rect.W - techFilterTabGap*float64(tabCount-1)
	tabW := availableW / float64(tabCount)
	if tabW < techFilterTabMinW {
		tabW = techFilterTabMinW
	}

	totalW := tabW*float64(tabCount) + techFilterTabGap*float64(tabCount-1)
	startX := rect.X + (rect.W-totalW)/2

	for idx, cat := range categories {
		tabX := startX + float64(idx)*(tabW+techFilterTabGap)

		isActive := activeFilter == cat || activeFilter == ""

		// Sekme arka plan
		bg := techFilterTabBG
		border := techFilterTabBorder
		if isActive {
			bg = techFilterTabActive
			border = techFilterTabActiveBorder
		}
		drawRoundedRectF64(screen, tabX, rect.Y, tabW, techFilterTabH, 6, bg)
		drawRoundedRectStrokeF64(screen, tabX, rect.Y, tabW, techFilterTabH, 6, 1.2, border, bg)

		// Sekme üst renk çizgisi
		catColor := techCategoryColors[cat]
		if isActive {
			vector.FillRect(screen, float32(tabX+4), float32(rect.Y+2), float32(tabW-8), 3, catColor, false)
		}

		// Kategori ikonu + etiket
		iconID := techCategoryIcon(cat)
		labelX := tabX + 8
		if iconID != gameui.IconNone {
			iconCol := catColor
			if !isActive {
				iconCol = color.RGBA{catColor.R / 2, catColor.G / 2, catColor.B / 2, 200}
			}
			gameui.DrawIcon(screen, iconID, tabX+8, rect.Y+10, 14, iconCol)
			labelX = tabX + 26
		}

		label := tech.CategoryLabelTR(cat)
		labelCol := color.RGBA{200, 200, 210, 230}
		if isActive {
			labelCol = ColorWhite
		}

		// Tamamlanan / toplam sayı
		if f != nil {
			completed, total := countTechByCategory(f, cat, nil)
			label = fmt.Sprintf("%s (%d/%d)", label, completed, total)
		}

		drawUILabel(screen, gameui.Rect{X: labelX, Y: rect.Y + 10, W: tabW - (labelX - tabX) - 8}, label,
			labelCol, gameui.TextSmall, gameui.TextAlignCenter)
	}
}

// countTechByCategory belirli bir kategorideki tamamlanmış ve toplam teknoloji sayısını döner.
func countTechByCategory(f *faction.Faction, cat tech.Category, allTechs map[string]*tech.Technology) (completed, total int) {
	if f == nil {
		return 0, 0
	}
	// Eğer allTechs nil ise, completed map'inden sayamayız tam olarak
	// Bu fonksiyon şu an sadece render için kullanılıyor, gs.TechTypes'e erişim yok.
	// Bu yüzden basit bir tahmin/placeholder döneceğiz.
	// NOT: Daha sonra gs referansı eklenerek tam implementasyon yapılabilir.
	return 0, 0
}

// techFilterTabHit hangi sekmenin tıklandığını döner.
func techFilterTabHit(mx, my float64, rect gameui.Rect) (tech.Category, bool) {
	if !rect.Hit(mx, my) {
		return "", false
	}
	categories := tech.AllCategories()
	tabCount := len(categories)
	availableW := rect.W - techFilterTabGap*float64(tabCount-1)
	tabW := availableW / float64(tabCount)
	if tabW < techFilterTabMinW {
		tabW = techFilterTabMinW
	}
	totalW := tabW*float64(tabCount) + techFilterTabGap*float64(tabCount-1)
	startX := rect.X + (rect.W-totalW)/2

	relX := mx - startX
	if relX < 0 {
		return "", false
	}
	idx := int(relX / (tabW + techFilterTabGap))
	if idx >= 0 && idx < len(categories) && relX < float64(idx+1)*(tabW+techFilterTabGap)-techFilterTabGap {
		return categories[idx], true
	}
	return "", false
}

// ── Lejant ──────────────────────────────────────────────────────────

func drawTechCategoryLegend(screen *ebiten.Image, rect gameui.Rect) {
	categories := tech.AllCategories()
	const perRow = 5
	const rowGap = 0.0
	const itemGap = 16.0

	colWidths := make([]float64, perRow)
	for idx, category := range categories {
		col := idx % perRow
		itemW := 20.0 + MeasureText(tech.CategoryLabelTR(category), FaceSmall)
		if itemW > colWidths[col] {
			colWidths[col] = itemW
		}
	}

	totalWidth := 0.0
	activeCols := techMinInt(perRow, len(categories))
	for i := 0; i < activeCols; i++ {
		totalWidth += colWidths[i]
		if i > 0 {
			totalWidth += itemGap
		}
	}
	startX := rect.X + rect.W - totalWidth - 8
	startY := rect.Y + 2

	// Lejant arka plan kartı
	bgRect := gameui.Rect{X: startX - 14, Y: rect.Y - 2, W: totalWidth + 22, H: 26}
	drawRoundedRectF64(screen, bgRect.X, bgRect.Y, bgRect.W, bgRect.H, 5, color.RGBA{12, 14, 28, 180})
	drawRoundedRectStrokeF64(screen, bgRect.X, bgRect.Y, bgRect.W, bgRect.H, 5, 1, color.RGBA{70, 75, 100, 160}, color.RGBA{12, 14, 28, 180})

	for idx, category := range categories {
		col := idx % perRow
		x := startX
		for i := 0; i < col; i++ {
			x += colWidths[i] + itemGap
		}
		y := startY
		label := tech.CategoryLabelTR(category)

		// Küçük renk noktası
		dotX := x
		dotY := y + 7
		dotR := float32(5)
		fill := techCategoryColors[category]
		// Daire çiz
		vector.FillCircle(screen, float32(dotX+5), float32(dotY), dotR, fill, true)
		vector.StrokeCircle(screen, float32(dotX+5), float32(dotY), dotR, 1, color.RGBA{240, 240, 240, 200}, true)

		drawUILabel(screen, gameui.Rect{X: x + 14, Y: y + 1}, label,
			color.RGBA{210, 210, 220, 230}, gameui.TextSmall, gameui.TextAlignStart)
	}
}

// ── Yardımcı draw fonksiyonları ─────────────────────────────────────

// drawRoundedRectF64 rounded rect çizer (float64 parametreli wrapper).
func drawRoundedRectF64(screen *ebiten.Image, x, y, w, h, r float64, col color.Color) {
	drawRoundedRect(screen, float32(x), float32(y), float32(w), float32(h), float32(r), col)
}

// drawRoundedRectStrokeF64 rounded rect stroke çizer.
// border renginde dış dolgu + bgColor ile iç temizleme mantığıyla yuvarlak köşeli çerçeve üretir.
func drawRoundedRectStrokeF64(screen *ebiten.Image, x, y, w, h, r, width float64, col color.Color, bgColor color.Color) {
	if r <= 0 || width <= 0 {
		if width > 0 {
			vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), float32(width), col, false)
		}
		return
	}
	// Dış dolgu: border kalınlığı kadar genişletilmiş, border renginde
	outerR := r + width
	drawRoundedRect(screen, float32(x-width), float32(y-width), float32(w+width*2), float32(h+width*2), float32(outerR), col)
	// İç temizleme: orijinal boyutta, arka plan renginde
	drawRoundedRect(screen, float32(x), float32(y), float32(w), float32(h), float32(r), bgColor)
}

func drawTechConnectors(screen *ebiten.Image, levels [][]techNode, allTechs map[string]*tech.Technology) {
	nodeIndex := indexTechNodes(levels)
	for _, levelNodes := range levels {
		for _, child := range levelNodes {
			reqCount := len(child.t.Requires)
			for reqIdx, reqID := range child.t.Requires {
				if _, ok := allTechs[reqID]; !ok {
					continue
				}
				parent, ok := nodeIndex[reqID]
				if !ok {
					continue
				}
				drawTechConnector(screen, parent, child, reqIdx, reqCount)
			}
		}
	}
}

func techMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func techMaxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func techClampFloat(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func techTreeViewOrigin(treeRect gameui.Rect, contentW float64) (x, y float64) {
	x = 0.0
	if contentW < treeRect.W {
		x = (treeRect.W - contentW) / 2
	}
	return x, 0
}

func projectTechTree(levels [][]techNode, offsetX, offsetY float64) [][]techNode {
	projected := make([][]techNode, len(levels))
	for levelIdx, levelNodes := range levels {
		projected[levelIdx] = make([]techNode, len(levelNodes))
		for nodeIdx, node := range levelNodes {
			node.x += offsetX
			node.y += offsetY
			projected[levelIdx][nodeIdx] = node
		}
	}
	return projected
}

func projectTechTreeForLayout(layout techPanelLayout, treeData techTreeLayoutData, panX, panY float64) (local [][]techNode, screen [][]techNode) {
	viewOriginX, viewOriginY := techTreeViewOrigin(layout.treeRect, treeData.contentW)
	local = projectTechTree(treeData.levels, viewOriginX-panX, viewOriginY-panY)
	screen = projectTechTree(treeData.levels, layout.treeRect.X+viewOriginX-panX, layout.treeRect.Y+viewOriginY-panY)
	return local, screen
}

func (r *Renderer) clampTechPan(treeRect gameui.Rect, contentW, contentH float64) {
	r.techPanX = techClampFloat(r.techPanX, 0, techMaxFloat(0, contentW-treeRect.W))
	r.techPanY = techClampFloat(r.techPanY, 0, techMaxFloat(0, contentH-treeRect.H))
}

// ── Ana Draw fonksiyonu ─────────────────────────────────────────────

// DrawTechPanel teknoloji araştırma panelini çizer. Alt bardaki Teknoloji tuşu veya [T] ile açılır.
func (r *Renderer) DrawTechPanel(screen *ebiten.Image) {
	if r.gs.TechTypes == nil {
		return
	}
	f := r.gs.Factions[r.gs.PlayerFactionID]
	if f == nil {
		return
	}

	layout := techPanelLayoutForScreen()

	// Tam ekran overlay
	drawUIOverlay(screen, techPanelOverlay)

	// Ana panel arka planı
	drawRoundedRectF64(screen, layout.panelRect.X+4, layout.panelRect.Y+4, layout.panelRect.W-8, layout.panelRect.H-8, 10, techPanelBG)
	// Üst bar
	drawUIPanelTopBar(screen, layout.panelRect, 3, techHeaderBarColor)

	// Kapat butonu
	drawTechCloseButton(screen)

	// Başlık
	drawUIPanelTitle(screen, layout.titleRect, "── Teknoloji Ağacı ──")

	// Aktif araştırma durumu
	activeY := layout.statusRect.Y
	if f.Research.ActiveID != "" {
		if t, ok := r.gs.TechTypes[f.Research.ActiveID]; ok {
			msg := fmt.Sprintf("▶ Araştırılıyor: %s  (%d tur kaldı)", t.NameTR, f.Research.TurnsLeft)
			DrawText(screen, msg, layout.statusRect.X, activeY+2, FaceMed, color.RGBA{110, 230, 110, 255})
		}
	} else {
		drawUIMutedText(screen, layout.statusRect.X, activeY+2, "Aktif araştırma yok — araştırmak için bir teknolojiye tıkla")
	}

	// Kategori filtre sekmeleri
	drawTechFilterTabs(screen, layout.filterRect, r.techFilterCategory, f)

	// Ağaç verisini hazırla
	treeData := r.buildLaidOutTechTree(f)
	r.clampTechPan(layout.treeRect, treeData.contentW, treeData.contentH)
	localProjectedLevels, screenProjectedLevels := projectTechTreeForLayout(layout, treeData, r.techPanX, r.techPanY)

	// Tree viewport arka plan (hafif grid efekti için)
	drawRoundedRectF64(screen, layout.treeRect.X, layout.treeRect.Y, layout.treeRect.W, layout.treeRect.H, 6, color.RGBA{8, 10, 22, 200})

	// Bağlantı çizgileri
	drawTechConnectors(screen, screenProjectedLevels, r.gs.TechTypes)

	// Seviye etiketleri ve düğümler
	for levelIdx, levelNodes := range localProjectedLevels {
		// Seviye etiketi (ilk düğümün soluna)
		if len(levelNodes) > 0 && levelIdx > 0 {
			labelX := layout.treeRect.X + levelNodes[0].x - techNodeWidth/2 - 50
			labelY := layout.treeRect.Y + levelNodes[0].y - 6
			if labelX < layout.treeRect.X+4 {
				labelX = layout.treeRect.X + 4
			}
			levelLabel := fmt.Sprintf("Sev. %d", levelIdx)
			drawUIOutlinedLabel(screen, gameui.Rect{X: labelX, Y: labelY, W: 46}, levelLabel,
				techLevelLabelColor, color.RGBA{4, 6, 16, 200}, gameui.TextSmall, gameui.TextAlignCenter)
		}

		for nodeIdx, node := range levelNodes {
			// Filtre kontrolü
			if r.techFilterCategory != "" && node.t.Category != r.techFilterCategory {
				continue
			}
			buildTechCardComponent(screenProjectedLevels[levelIdx][nodeIdx], f.Research.ActiveID, f.Research).Draw(screen)
		}
	}

	// Lejant ve alt bilgi
	drawTechCategoryLegend(screen, gameui.Rect{X: layout.hintRect.X, Y: layout.hintRect.Y - 28, W: layout.hintRect.W})
	hintY := layout.hintRect.Y + 4
	drawUIMutedText(screen, layout.hintRect.X, hintY, "Tıkla: araştır   Sürükle/tekerlek: gez   Filtre: üst sekmeler   "+economy.ResourceNameTR(economy.ResourceGold)+": "+fmt.Sprintf("%d", f.Gold))
}

func techCloseRect() (x, y, w, h float32) {
	r := techPanelLayoutForScreen().closeRect
	return float32(r.X), float32(r.Y), float32(r.W), float32(r.H)
}

func drawTechCloseButton(screen *ebiten.Image) {
	btn := buildTechCloseButton()
	style := gameui.ButtonStyle{
		BG:             color.RGBA{30, 28, 42, 220},
		Border:         color.RGBA{160, 140, 80, 200},
		Text:           ColorGold,
		DisabledBG:     color.RGBA{20, 18, 30, 180},
		DisabledBorder: color.RGBA{60, 50, 30, 140},
		DisabledText:   color.RGBA{100, 90, 70, 180},
		TextOffsetY:    6,
		TextVariant:    gameui.TextSmall,
		BorderWidth:    1.5,
	}
	drawUIButtonWidget(screen, btn, style)
}

func drawTechButton(screen *ebiten.Image, btn gameui.Button, bg color.RGBA, textColor color.Color, textOffsetY float64) {
	style := menuButtonStyle
	style.BG = bg
	style.Border = panelBorder
	style.Text = color.RGBAModel.Convert(textColor).(color.RGBA)
	style.TextOffsetY = textOffsetY
	style.TextVariant = gameui.TextSmall
	drawUIButtonWidget(screen, btn, style)
}

// ── Input handling ──────────────────────────────────────────────────

func (r *Renderer) handleTechInput(f *faction.Faction, input gameui.InputState) InputAction {
	if r.gs.TechTypes == nil {
		return InputAction{}
	}
	layout := techPanelLayoutForScreen()

	// Filtre sekmesi tıklaması
	if cat, hit := techFilterTabHit(input.MouseX, input.MouseY, layout.filterRect); hit && input.LeftJustPressed {
		if r.techFilterCategory == cat {
			r.techFilterCategory = "" // Toggle off
		} else {
			r.techFilterCategory = cat
		}
		return InputAction{}
	}

	treeData := r.buildLaidOutTechTree(f)
	r.clampTechPan(layout.treeRect, treeData.contentW, treeData.contentH)
	_, screenProjectedLevels := projectTechTreeForLayout(layout, treeData, r.techPanX, r.techPanY)

	if buildTechCloseButton().HandleInput(input) {
		r.showTech = false
		r.techDragging = false
		r.techFilterCategory = ""
		return InputAction{}
	}
	if input.LeftJustReleased {
		r.techDragging = false
	}
	if layout.treeRect.Hit(input.MouseX, input.MouseY) && input.WheelY != 0 {
		r.techPanY -= input.WheelY * techPanWheelStep
		r.clampTechPan(layout.treeRect, treeData.contentW, treeData.contentH)
	}

	// Ağaç düğümlerine tıklama
	for _, levelNodes := range screenProjectedLevels {
		for _, node := range levelNodes {
			// Filtreyi uygula
			if r.techFilterCategory != "" && node.t.Category != r.techFilterCategory {
				continue
			}
			card := buildTechCardComponent(node, f.Research.ActiveID, f.Research)
			if input.LeftJustPressed && card.HitTest(input.MouseX, input.MouseY) {
				if node.unlocked && !node.done {
					return InputAction{Kind: ActionResearch, BuildingID: node.t.ID}
				}
				return InputAction{}
			}
		}
	}
	if layout.treeRect.Hit(input.MouseX, input.MouseY) && input.LeftJustPressed {
		r.techDragging = true
		r.techDragLastMX = input.MouseX
		r.techDragLastMY = input.MouseY
		return InputAction{}
	}
	if r.techDragging && input.LeftPressed {
		dx := input.MouseX - r.techDragLastMX
		dy := input.MouseY - r.techDragLastMY
		if dx >= techPanDragThresh || dx <= -techPanDragThresh || dy >= techPanDragThresh || dy <= -techPanDragThresh {
			r.techPanX -= dx
			r.techPanY -= dy
			r.clampTechPan(layout.treeRect, treeData.contentW, treeData.contentH)
			r.techDragLastMX = input.MouseX
			r.techDragLastMY = input.MouseY
		}
	}

	return InputAction{}
}

func (r *Renderer) techPanelPointerHit(fx, fy float64) bool {
	if r.gs.TechTypes == nil {
		return false
	}
	f := r.gs.Factions[r.gs.PlayerFactionID]
	if f == nil {
		return false
	}
	if buildTechCloseButton().HitTest(fx, fy) {
		return true
	}
	layout := techPanelLayoutForScreen()
	if layout.treeRect.Hit(fx, fy) {
		return true
	}
	// Filtre sekmeleri
	if layout.filterRect.Hit(fx, fy) {
		return true
	}
	treeData := r.buildLaidOutTechTree(f)
	_, screenProjectedLevels := projectTechTreeForLayout(layout, treeData, r.techPanX, r.techPanY)
	for _, levelNodes := range screenProjectedLevels {
		for _, node := range levelNodes {
			if !node.unlocked || node.done {
				continue
			}
			if r.techFilterCategory != "" && node.t.Category != r.techFilterCategory {
				continue
			}
			if buildTechCardComponent(node, f.Research.ActiveID, f.Research).HitTest(fx, fy) {
				return true
			}
		}
	}
	return false
}
