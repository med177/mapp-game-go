package render

import (
	"fmt"
	"image"
	"image/color"
	"sort"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var techCategoryColors = map[tech.Category]color.RGBA{
	tech.CategoryMilitary:  {200, 100, 100, 255}, // Kırmızımsı
	tech.CategoryEconomy:   {100, 200, 100, 255}, // Yeşil
	tech.CategoryDiplomacy: {100, 100, 200, 255}, // Mavi
	tech.CategoryNaval:     {200, 200, 100, 255}, // Sarı
	tech.CategoryReligion:  {200, 100, 200, 255}, // Magenta
}

type techNode struct {
	t        *tech.Technology
	unlocked bool
	done     bool
	level    int
	x, y     float64 // Ağaç pozisyonu
}

const (
	techLevelHeight = 120.0
	techNodeWidth   = 180.0
	techNodeHeight  = 84.0

	techTreeMaxCols   = 4
	techNodeStepX     = 190.0
	techNodeStepY     = 106.0
	techLevelGap      = 52.0
	techTreePadding   = 24.0
	techPanWheelStep  = 54.0
	techPanDragThresh = 4.0

	techConnectorStem     = 12.0
	techConnectorLaneStep = 8.0
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
	statusRect, box := box.CutTop(24, 12)
	hintRect, treeBox := box.CutBottom(42, 0)
	closeRect, titleRect := gameui.BoxFromRect(headerRect).CutRight(30, 12)
	return techPanelLayout{
		panelRect:  panel,
		titleRect:  titleRect.Rect,
		statusRect: statusRect,
		hintRect:   hintRect,
		closeRect:  closeRect,
		treeRect:   treeBox.Rect,
	}
}

func buildTechCloseButton() gameui.Button {
	x, y, w, h := techCloseRect()
	btn := gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "").WithIcon(gameui.IconClose)
	btn.IconSize = 13
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

func (r *Renderer) buildLaidOutTechTree(f *faction.Faction) techTreeLayoutData {
	levels := r.buildTechTree(f)
	contentW, contentH := layoutTechTree(levels, techTreeMaxCols)
	return techTreeLayoutData{levels: levels, contentW: contentW, contentH: contentH}
}

func (r *Renderer) buildTechTree(f *faction.Faction) [][]techNode {
	// Teknolojiyi seviyelere göre gruplandır
	levels := make(map[int][]techNode)
	maxLevel := 0

	// Önce tüm teknolojileri işle
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

	// Seviyeleri sırala
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
	col := color.RGBA{132, 132, 140, 210}
	if child.unlocked {
		col = color.RGBA{188, 188, 196, 232}
	}
	if child.done && parent.done {
		col = color.RGBA{208, 220, 208, 236}
	}
	return techConnectorStyle{col: col, width: 2}
}

func techConnectorMidY(parent, child techNode, reqIdx, reqCount int) float64 {
	startY := parent.y + techNodeHeight/2
	endY := child.y - techNodeHeight/2
	minY := startY + techConnectorStem
	maxY := endY - techConnectorStem
	if maxY <= minY {
		return (startY + endY) / 2
	}

	baseY := minY + (maxY-minY)*0.45
	if reqCount <= 1 {
		return baseY
	}

	spread := float64(reqCount-1) * techConnectorLaneStep
	offset := float64(reqIdx)*techConnectorLaneStep - spread/2
	midY := baseY + offset
	if midY < minY {
		return minY
	}
	if midY > maxY {
		return maxY
	}
	return midY
}

func drawTechLine(screen *ebiten.Image, x1, y1, x2, y2 float64, style techConnectorStyle) {
	if style.dashed {
		drawDashedTechLine(screen, x1, y1, x2, y2, style)
		return
	}
	vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), style.width, style.col, false)
}

func drawDashedTechLine(screen *ebiten.Image, x1, y1, x2, y2 float64, style techConnectorStyle) {
	const dashLen = 8.0
	const gapLen = 5.0
	if x1 == x2 {
		if y2 < y1 {
			y1, y2 = y2, y1
		}
		for segStart := y1; segStart < y2; segStart += dashLen + gapLen {
			segEnd := segStart + dashLen
			if segEnd > y2 {
				segEnd = y2
			}
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
			vector.StrokeLine(screen, float32(segStart), float32(y1), float32(segEnd), float32(y2), style.width, style.col, false)
		}
	}
}

func drawTechConnector(screen *ebiten.Image, parent, child techNode, reqIdx, reqCount int) {
	style := techConnectorStyleFor(parent, child)
	startY := parent.y + techNodeHeight/2
	endY := child.y - techNodeHeight/2
	midY := techConnectorMidY(parent, child, reqIdx, reqCount)

	drawTechLine(screen, parent.x, startY, parent.x, midY, style)
	drawTechLine(screen, parent.x, midY, child.x, midY, style)
	drawTechLine(screen, child.x, midY, child.x, endY, style)
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

func techNodeNameColor(node techNode) color.Color {
	if node.unlocked || node.done {
		return ColorWhite
	}
	return color.RGBA{214, 214, 214, 255}
}

func drawTechNodeText(screen *ebiten.Image, node techNode, nodeRect gameui.Rect) {
	nameRect := gameui.Rect{X: nodeRect.X + 10, Y: nodeRect.Y + 8, W: nodeRect.W - 20}
	drawUIOutlinedLabel(screen, nameRect, node.t.NameTR, techNodeNameColor(node), color.RGBA{8, 8, 12, 255}, gameui.TextMedium, gameui.TextAlignCenter)

	buffRect := gameui.Rect{X: nodeRect.X + 10, Y: nodeRect.Y + 31, W: nodeRect.W - 20}
	drawUIWrappedLabelAligned(screen, buffRect, techEffectSummary(node.t), color.RGBA{245, 245, 245, 245}, gameui.TextSmall, 14, 2, gameui.TextAlignCenter)

	costStr := fmt.Sprintf("%d altın  •  %d tur", node.t.GoldCost, node.t.TurnsRequired)
	costColor := color.RGBA{255, 228, 138, 255}
	if !node.unlocked && !node.done {
		costColor = color.RGBA{210, 210, 210, 240}
	}
	costRect := gameui.Rect{X: nodeRect.X + 10, Y: nodeRect.Y + nodeRect.H - 18, W: nodeRect.W - 20}
	drawUIOutlinedLabel(screen, costRect, costStr, costColor, color.RGBA{16, 12, 8, 255}, gameui.TextSmall, gameui.TextAlignCenter)
}

func drawTechCategoryLegend(screen *ebiten.Image, rect gameui.Rect) {
	categories := tech.AllCategories()
	const perRow = 3
	const rowGap = 18.0
	const itemGap = 18.0
	colWidths := make([]float64, perRow)
	for idx, category := range categories {
		col := idx % perRow
		itemW := 16.0 + MeasureText(tech.CategoryLabelTR(category), FaceSmall)
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
	startX := rect.X + rect.W - totalWidth
	startY := rect.Y + 1

	bgRect := gameui.Rect{X: startX - 12, Y: rect.Y - 1, W: totalWidth + 18, H: 36}
	vector.FillRect(screen, float32(bgRect.X), float32(bgRect.Y), float32(bgRect.W), float32(bgRect.H), color.RGBA{10, 12, 18, 145}, false)
	vector.StrokeRect(screen, float32(bgRect.X), float32(bgRect.Y), float32(bgRect.W), float32(bgRect.H), 1, color.RGBA{75, 80, 98, 160}, false)

	for idx, category := range categories {
		row := idx / perRow
		col := idx % perRow
		x := startX
		for i := 0; i < col; i++ {
			x += colWidths[i] + itemGap
		}
		y := startY + float64(row)*rowGap
		label := tech.CategoryLabelTR(category)
		swatch := gameui.Rect{X: x, Y: y + 4, W: 10, H: 10}
		fill := techCategoryColors[category]
		vector.FillRect(screen, float32(swatch.X), float32(swatch.Y), float32(swatch.W), float32(swatch.H), fill, false)
		vector.StrokeRect(screen, float32(swatch.X), float32(swatch.Y), float32(swatch.W), float32(swatch.H), 1, color.RGBA{240, 240, 240, 220}, false)
		drawUIOutlinedLabel(screen, gameui.Rect{X: x + 16, Y: y}, label, ColorWhite, color.RGBA{8, 8, 12, 255}, gameui.TextSmall, gameui.TextAlignStart)
	}
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

func (r *Renderer) clampTechPan(treeRect gameui.Rect, contentW, contentH float64) {
	r.techPanX = techClampFloat(r.techPanX, 0, techMaxFloat(0, contentW-treeRect.W))
	r.techPanY = techClampFloat(r.techPanY, 0, techMaxFloat(0, contentH-treeRect.H))
}

func techTreeViewport(screen *ebiten.Image, treeRect gameui.Rect) *ebiten.Image {
	rect := image.Rect(
		int(treeRect.X),
		int(treeRect.Y),
		int(treeRect.X+treeRect.W),
		int(treeRect.Y+treeRect.H),
	)
	return screen.SubImage(rect).(*ebiten.Image)
}

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

	drawUIOverlay(screen, color.RGBA{8, 6, 4, 220})

	drawUIPanelRect(screen, layout.panelRect, color.RGBA{20, 20, 40, 230}, color.RGBA{20, 20, 40, 230}, 0)
	drawUIPanelTopBar(screen, layout.panelRect, 2, color.RGBA{180, 150, 60, 255})
	drawTechCloseButton(screen)

	drawUIPanelTitle(screen, layout.titleRect, "── Teknoloji Ağacı ──")

	activeY := layout.statusRect.Y
	if f.Research.ActiveID != "" {
		if t, ok := r.gs.TechTypes[f.Research.ActiveID]; ok {
			msg := fmt.Sprintf("Araştırılıyor: %s  (%d tur kaldı)", t.NameTR, f.Research.TurnsLeft)
			DrawText(screen, msg, layout.statusRect.X, activeY, FaceMed, color.RGBA{100, 220, 100, 255})
		}
	} else {
		drawUIMutedText(screen, layout.statusRect.X, activeY, "Aktif araştırma yok")
	}

	treeData := r.buildLaidOutTechTree(f)
	r.clampTechPan(layout.treeRect, treeData.contentW, treeData.contentH)
	viewOriginX, viewOriginY := techTreeViewOrigin(layout.treeRect, treeData.contentW)
	projectedLevels := projectTechTree(treeData.levels, viewOriginX-r.techPanX, viewOriginY-r.techPanY)
	treeScreen := techTreeViewport(screen, layout.treeRect)
	drawTechConnectors(treeScreen, projectedLevels, r.gs.TechTypes)

	// Her seviye için düğümleri çiz
	for _, levelNodes := range projectedLevels {
		for _, node := range levelNodes {
			// Düğüm rengi
			var nodeColor color.RGBA
			baseCategoryColor := techCategoryColors[node.t.Category]
			if !node.unlocked {
				nodeColor = color.RGBA{80, 80, 80, 255} // Kilitli - gri
			} else if f.Research.ActiveID == node.t.ID {
				nodeColor = color.RGBA{255, 220, 80, 255} // Araştırılıyor - sarı
			} else if node.done {
				nodeColor = color.RGBA{baseCategoryColor.R, baseCategoryColor.G, baseCategoryColor.B, 200}
			} else {
				nodeColor = baseCategoryColor // Kategori rengi
			}

			nodeRect := techNodeRect(node)

			// Düğüm arka planı
			vector.FillRect(treeScreen, float32(nodeRect.X), float32(nodeRect.Y),
				float32(nodeRect.W), float32(nodeRect.H), nodeColor, false)

			// Düğüm çerçevesi
			vector.StrokeRect(treeScreen, float32(nodeRect.X), float32(nodeRect.Y),
				float32(nodeRect.W), float32(nodeRect.H), 2, color.RGBA{255, 255, 255, 255}, false)
			drawTechNodeText(treeScreen, node, nodeRect)

			// Tamamlandı tik rozeti
			if node.done {
				badgeW := 24.0
				badgeH := 18.0
				badgeX := nodeRect.X + nodeRect.W - badgeW - 8
				badgeY := nodeRect.Y + 8
				vector.FillRect(treeScreen, float32(badgeX), float32(badgeY), float32(badgeW), float32(badgeH), color.RGBA{28, 56, 34, 232}, false)
				vector.StrokeRect(treeScreen, float32(badgeX), float32(badgeY), float32(badgeW), float32(badgeH), 1, color.RGBA{170, 225, 170, 255}, false)
				gameui.DrawIcon(treeScreen, gameui.IconCheck, badgeX+4, badgeY+1, 14, color.RGBA{245, 255, 245, 255})
			}

			// Araştırılıyor rozeti
			if f.Research.ActiveID == node.t.ID {
				badgeW := 24.0
				badgeH := 18.0
				badgeX := nodeRect.X + 8
				badgeY := nodeRect.Y + 8
				vector.FillRect(treeScreen, float32(badgeX), float32(badgeY), float32(badgeW), float32(badgeH), color.RGBA{86, 60, 18, 236}, false)
				vector.StrokeRect(treeScreen, float32(badgeX), float32(badgeY), float32(badgeW), float32(badgeH), 1, color.RGBA{245, 210, 110, 255}, false)
				gameui.DrawIcon(treeScreen, gameui.IconPlay, badgeX+5, badgeY+1, 14, color.RGBA{255, 244, 210, 255})
			}

		}
	}

	drawTechCategoryLegend(screen, gameui.Rect{X: layout.hintRect.X, Y: layout.hintRect.Y, W: layout.hintRect.W})
	hintY := layout.hintRect.Y + 20
	drawUIMutedText(screen, layout.hintRect.X, hintY, "Tıkla: araştır   Sürükle/tekerlek: gez   "+economy.ResourceNameTR(economy.ResourceGold)+": "+fmt.Sprintf("%d", f.Gold))
}

func techCloseRect() (x, y, w, h float32) {
	r := techPanelLayoutForScreen().closeRect
	return float32(r.X), float32(r.Y), float32(r.W), float32(r.H)
}

func drawTechCloseButton(screen *ebiten.Image) {
	drawTechButton(screen, buildTechCloseButton(), color.RGBA{45, 34, 25, 230}, ColorGold, 6)
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

// handleTechInput teknoloji paneli klavye ve fare girişlerini işler.
func (r *Renderer) handleTechInput(f *faction.Faction, input gameui.InputState) InputAction {
	if r.gs.TechTypes == nil {
		return InputAction{}
	}
	layout := techPanelLayoutForScreen()
	treeData := r.buildLaidOutTechTree(f)
	r.clampTechPan(layout.treeRect, treeData.contentW, treeData.contentH)
	viewOriginX, viewOriginY := techTreeViewOrigin(layout.treeRect, treeData.contentW)
	contentMouseX := input.MouseX - layout.treeRect.X - viewOriginX + r.techPanX
	contentMouseY := input.MouseY - layout.treeRect.Y - viewOriginY + r.techPanY
	contentInput := input
	contentInput.MouseX = contentMouseX
	contentInput.MouseY = contentMouseY

	if buildTechCloseButton().HandleInput(input) {
		r.showTech = false
		r.techDragging = false
		return InputAction{}
	}
	if input.LeftJustReleased {
		r.techDragging = false
	}
	if layout.treeRect.Hit(input.MouseX, input.MouseY) && input.WheelY != 0 {
		r.techPanY -= input.WheelY * techPanWheelStep
		r.clampTechPan(layout.treeRect, treeData.contentW, treeData.contentH)
	}

	projectedLevels := treeData.levels

	// Ağaç düğümlerine tıklama
	for _, levelNodes := range projectedLevels {
		for _, node := range levelNodes {
			if contentInput.LeftJustPressed && techNodeRect(node).Hit(contentInput.MouseX, contentInput.MouseY) {
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
	treeData := r.buildLaidOutTechTree(f)
	viewOriginX, viewOriginY := techTreeViewOrigin(layout.treeRect, treeData.contentW)
	contentFX := fx - layout.treeRect.X - viewOriginX + r.techPanX
	contentFY := fy - layout.treeRect.Y - viewOriginY + r.techPanY
	for _, levelNodes := range treeData.levels {
		for _, node := range levelNodes {
			if !node.unlocked || node.done {
				continue
			}
			if techNodeButton(node).HitTest(contentFX, contentFY) {
				return true
			}
		}
	}
	return false
}
