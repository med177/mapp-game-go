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
	techNodeHeight  = 60.0

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

func techPanelLayoutForScreen() techPanelLayout {
	panel := gameui.Rect{X: 0, Y: 0, W: ScreenWidth, H: ScreenHeight}
	box := gameui.BoxFromRect(panel).InsetXY(20, 20)
	headerRect, box := box.CutTop(28, 12)
	statusRect, box := box.CutTop(24, 12)
	hintRect, treeBox := box.CutBottom(20, 0)
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

func (r *Renderer) buildLaidOutTechTree(f *faction.Faction) [][]techNode {
	levels := r.buildTechTree(f)
	layout := techPanelLayoutForScreen()
	treeStartY := techTreeStartY(layout.treeRect, float64(len(levels)), techLevelHeight)
	layoutTechTree(levels, layout.treeRect, techNodeWidth, techNodeHeight, treeStartY, techLevelHeight)
	return levels
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
			// Her seviyedeki teknolojileri kategoriye göre sırala
			sort.Slice(nodes, func(a, b int) bool {
				if nodes[a].t.Category != nodes[b].t.Category {
					return tech.CategoryOrder(nodes[a].t.Category) < tech.CategoryOrder(nodes[b].t.Category)
				}
				return nodes[a].t.ID < nodes[b].t.ID
			})
			result = append(result, nodes)
		}
	}

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

func layoutTechTree(levels [][]techNode, bounds gameui.Rect, nodeWidth, nodeHeight, treeStartY, levelHeight float64) {
	for levelIdx, levelNodes := range levels {
		levelY := treeStartY + float64(levelIdx)*levelHeight
		levelWidth := float64(len(levelNodes)) * nodeWidth
		startX := bounds.X + (bounds.W-levelWidth)/2
		for nodeIdx := range levelNodes {
			levels[levelIdx][nodeIdx].x = startX + float64(nodeIdx)*nodeWidth + nodeWidth/2
			levels[levelIdx][nodeIdx].y = levelY + nodeHeight/2
		}
	}
}

func techTreeStartY(bounds gameui.Rect, levelCount, levelHeight float64) float64 {
	totalHeight := levelCount * levelHeight
	centeredY := bounds.Y + (bounds.H-totalHeight)/2
	if centeredY < bounds.Y {
		return bounds.Y
	}
	return centeredY
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

func techNodeNameColor(node techNode) color.Color {
	if node.unlocked || node.done {
		return ColorWhite
	}
	return color.RGBA{214, 214, 214, 255}
}

func drawTechNodeText(screen *ebiten.Image, node techNode, nodeRect gameui.Rect) {
	titleStrip := gameui.Rect{X: nodeRect.X + 6, Y: nodeRect.Y + 6, W: nodeRect.W - 12, H: 18}
	footerStrip := gameui.Rect{X: nodeRect.X + 8, Y: nodeRect.Y + nodeRect.H - 20, W: nodeRect.W - 16, H: 14}
	vector.FillRect(screen, float32(titleStrip.X), float32(titleStrip.Y), float32(titleStrip.W), float32(titleStrip.H), color.RGBA{10, 12, 18, 150}, false)
	vector.FillRect(screen, float32(footerStrip.X), float32(footerStrip.Y), float32(footerStrip.W), float32(footerStrip.H), color.RGBA{8, 10, 16, 135}, false)

	nameRect := gameui.Rect{X: nodeRect.X + 10, Y: nodeRect.Y + 8, W: nodeRect.W - 20}
	drawUIOutlinedLabel(screen, nameRect, node.t.NameTR, techNodeNameColor(node), color.RGBA{8, 8, 12, 255}, gameui.TextMedium, gameui.TextAlignCenter)

	catLabel := fmt.Sprintf("[%s]", tech.CategoryLabelTR(node.t.Category))
	catRect := gameui.Rect{X: nodeRect.X + 10, Y: nodeRect.Y + 29, W: nodeRect.W - 20}
	catColor := techCategoryColors[node.t.Category]
	catColor.A = 248
	drawUIOutlinedLabel(screen, catRect, catLabel, catColor, color.RGBA{10, 10, 14, 255}, gameui.TextSmall, gameui.TextAlignCenter)

	if node.unlocked && !node.done {
		costStr := fmt.Sprintf("%dg/%dt", node.t.GoldCost, node.t.TurnsRequired)
		costRect := gameui.Rect{X: nodeRect.X + 10, Y: nodeRect.Y + nodeRect.H - 19, W: nodeRect.W - 20}
		drawUIOutlinedLabel(screen, costRect, costStr, color.RGBA{255, 228, 138, 255}, color.RGBA{16, 12, 8, 255}, gameui.TextSmall, gameui.TextAlignCenter)
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

	levels := r.buildLaidOutTechTree(f)
	drawTechConnectors(screen, levels, r.gs.TechTypes)

	// Her seviye için düğümleri çiz
	for _, levelNodes := range levels {
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
			vector.FillRect(screen, float32(nodeRect.X), float32(nodeRect.Y),
				float32(nodeRect.W), float32(nodeRect.H), nodeColor, false)

			// Düğüm çerçevesi
			vector.StrokeRect(screen, float32(nodeRect.X), float32(nodeRect.Y),
				float32(nodeRect.W), float32(nodeRect.H), 2, color.RGBA{255, 255, 255, 255}, false)
			drawTechNodeText(screen, node, nodeRect)

			// Tamamlandı tik rozeti
			if node.done {
				badgeW := 24.0
				badgeH := 18.0
				badgeX := nodeRect.X + nodeRect.W - badgeW - 8
				badgeY := nodeRect.Y + 8
				vector.FillRect(screen, float32(badgeX), float32(badgeY), float32(badgeW), float32(badgeH), color.RGBA{28, 56, 34, 232}, false)
				vector.StrokeRect(screen, float32(badgeX), float32(badgeY), float32(badgeW), float32(badgeH), 1, color.RGBA{170, 225, 170, 255}, false)
				gameui.DrawIcon(screen, gameui.IconCheck, badgeX+4, badgeY+1, 14, color.RGBA{245, 255, 245, 255})
			}

			// Araştırılıyor rozeti
			if f.Research.ActiveID == node.t.ID {
				badgeW := 24.0
				badgeH := 18.0
				badgeX := nodeRect.X + 8
				badgeY := nodeRect.Y + 8
				vector.FillRect(screen, float32(badgeX), float32(badgeY), float32(badgeW), float32(badgeH), color.RGBA{86, 60, 18, 236}, false)
				vector.StrokeRect(screen, float32(badgeX), float32(badgeY), float32(badgeW), float32(badgeH), 1, color.RGBA{245, 210, 110, 255}, false)
				gameui.DrawIcon(screen, gameui.IconPlay, badgeX+5, badgeY+1, 14, color.RGBA{255, 244, 210, 255})
			}

		}
	}

	hintY := layout.hintRect.Y
	drawUIMutedText(screen, layout.hintRect.X, hintY, "Teknoloji düğümlerine tıklayarak araştır   "+economy.ResourceNameTR(economy.ResourceGold)+": "+fmt.Sprintf("%d", f.Gold))
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

	levels := r.buildLaidOutTechTree(f)

	if buildTechCloseButton().HandleInput(input) {
		r.showTech = false
		return InputAction{}
	}

	// Ağaç düğümlerine tıklama
	for _, levelNodes := range levels {
		for _, node := range levelNodes {
			if techNodeButton(node).HandleInput(input) {
				if node.unlocked && !node.done {
					if f.Research.ActiveID == node.t.ID {
						return InputAction{Kind: ActionCancelResearch}
					} else {
						return InputAction{Kind: ActionResearch, BuildingID: node.t.ID}
					}
				}
				return InputAction{}
			}
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
	levels := r.buildLaidOutTechTree(f)
	for _, levelNodes := range levels {
		for _, node := range levelNodes {
			if !node.unlocked || node.done {
				continue
			}
			if techNodeButton(node).HitTest(fx, fy) {
				return true
			}
		}
	}
	return false
}
