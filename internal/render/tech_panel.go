package render

import (
	"fmt"
	"image/color"
	"sort"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var techCategoryLabels = map[tech.Category]string{
	tech.CategoryMilitary:  "Askeri",
	tech.CategoryEconomy:   "Ekonomi",
	tech.CategoryDiplomacy: "Diplomasi",
	tech.CategoryNaval:     "Denizcilik",
	tech.CategoryReligion:  "Din",
}

var techCategoryColors = map[tech.Category]color.RGBA{
	tech.CategoryMilitary:  {200, 100, 100, 255}, // Kırmızımsı
	tech.CategoryEconomy:   {100, 200, 100, 255}, // Yeşil
	tech.CategoryDiplomacy: {100, 100, 200, 255}, // Mavi
	tech.CategoryNaval:     {200, 200, 100, 255}, // Sarı
	tech.CategoryReligion:  {200, 100, 200, 255}, // Magenta
}

var techCategoryOrder = []tech.Category{
	tech.CategoryMilitary,
	tech.CategoryEconomy,
	tech.CategoryDiplomacy,
	tech.CategoryNaval,
	tech.CategoryReligion,
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
)

func buildTechCloseButton() gameui.Button {
	x, y, w, h := techCloseRect()
	return gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "X")
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
	treeStartY := techTreeStartY(float64(len(levels)), techLevelHeight)
	layoutTechTree(levels, float64(ScreenWidth), techNodeWidth, techNodeHeight, treeStartY, techLevelHeight)
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
					return nodes[a].t.Category < nodes[b].t.Category
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

func layoutTechTree(levels [][]techNode, screenWidth, nodeWidth, nodeHeight, treeStartY, levelHeight float64) {
	for levelIdx, levelNodes := range levels {
		levelY := treeStartY + float64(levelIdx)*levelHeight
		levelWidth := float64(len(levelNodes)) * nodeWidth
		startX := (screenWidth - levelWidth) / 2
		for nodeIdx := range levelNodes {
			levels[levelIdx][nodeIdx].x = startX + float64(nodeIdx)*nodeWidth + nodeWidth/2
			levels[levelIdx][nodeIdx].y = levelY + nodeHeight/2
		}
	}
}

func techTreeStartY(levelCount, levelHeight float64) float64 {
	totalHeight := levelCount * levelHeight
	centeredY := (float64(ScreenHeight) - totalHeight) / 2
	if centeredY < 80.0 {
		return 80.0
	}
	return centeredY
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

	overlay := ebiten.NewImage(int(ScreenWidth), int(ScreenHeight))
	overlay.Fill(color.RGBA{8, 6, 4, 220})
	screen.DrawImage(overlay, nil)

	px, py := float32(0), float32(0)
	pw, ph := float32(ScreenWidth), float32(ScreenHeight)

	vector.FillRect(screen, px, py, pw, ph, color.RGBA{20, 20, 40, 230}, false)
	vector.FillRect(screen, px, py, pw, 2, color.RGBA{180, 150, 60, 255}, false)
	drawTechCloseButton(screen)

	DrawTextCentered(screen, "── Teknoloji Ağacı ──", ScreenWidth/2, 24, FaceLarge, ColorYellow)

	activeY := float64(py) + 50
	if f.Research.ActiveID != "" {
		if t, ok := r.gs.TechTypes[f.Research.ActiveID]; ok {
			msg := fmt.Sprintf("Araştırılıyor: %s  (%d tur kaldı)", t.NameTR, f.Research.TurnsLeft)
			DrawText(screen, msg, float64(px)+20, activeY, FaceMed, color.RGBA{100, 220, 100, 255})
		}
	} else {
		DrawText(screen, "Aktif araştırma yok", float64(px)+20, activeY, FaceSmall, ColorGray)
	}

	levels := r.buildLaidOutTechTree(f)

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

			// Teknoloji adı
			nameY := nodeRect.Y + 8
			textColor := ColorWhite
			if node.unlocked && !node.done {
				textColor = color.RGBA{uint8(nodeColor.R / 3), uint8(nodeColor.G / 3), uint8(nodeColor.B / 3), 255}
			}
			DrawTextCentered(screen, node.t.NameTR, node.x, nameY, FaceMed, textColor)

			// Kategori etiketi
			catLabel := techCategoryLabels[node.t.Category]
			catY := node.y - 8
			catColor := techCategoryColors[node.t.Category]
			catColor.A = 200
			DrawTextCentered(screen, fmt.Sprintf("[%s]", catLabel), node.x, catY, FaceSmall, catColor)

			// Maliyet bilgisi (kilitli değilse)
			if node.unlocked && !node.done {
				costY := nodeRect.Y + nodeRect.H - 20
				costStr := fmt.Sprintf("%dg/%dt", node.t.GoldCost, node.t.TurnsRequired)
				DrawTextCentered(screen, costStr, node.x, costY, FaceSmall, ColorGold)
			}

			// Tamamlandı tik rozeti
			if node.done {
				badgeW := 24.0
				badgeH := 18.0
				badgeX := nodeRect.X + nodeRect.W - badgeW - 8
				badgeY := nodeRect.Y + 8
				vector.FillRect(screen, float32(badgeX), float32(badgeY), float32(badgeW), float32(badgeH), color.RGBA{35, 35, 35, 220}, false)
				vector.StrokeRect(screen, float32(badgeX), float32(badgeY), float32(badgeW), float32(badgeH), 1, color.RGBA{220, 220, 220, 255}, false)
				tw := MeasureText("✓", FaceSmall)
				DrawText(screen, "✓", badgeX+badgeW/2-tw/2, badgeY+2, FaceSmall, ColorWhite)
			}

			// Bağlantı çizgileri (gereksinimlere)
			if len(node.t.Requires) > 0 {
				for _, reqID := range node.t.Requires {
					if reqTech, ok := r.gs.TechTypes[reqID]; ok {
						// Gereksinim teknolojisinin pozisyonunu bul
						reqLevel := r.getTechLevel(reqTech, r.gs.TechTypes)
						if reqLevel < len(levels) {
							for _, reqNode := range levels[reqLevel] {
								if reqNode.t.ID == reqID {
									// Çizgi çiz
									vector.StrokeLine(screen,
										float32(reqNode.x), float32(reqNode.y+techNodeHeight/2),
										float32(node.x), float32(node.y-techNodeHeight/2),
										2, color.RGBA{150, 150, 150, 255}, false)
									break
								}
							}
						}
					}
				}
			}
		}
	}

	hintY := float64(ph) - 18
	DrawText(screen, "Teknoloji düğümlerine tıklayarak araştır   Altin: "+fmt.Sprintf("%d", f.Gold),
		float64(px)+20, hintY, FaceSmall, color.RGBA{160, 160, 100, 255})
}

func techCloseRect() (x, y, w, h float32) {
	return float32(ScreenWidth) - 58, 20, 30, 26
}

func drawTechCloseButton(screen *ebiten.Image) {
	drawTechButton(screen, buildTechCloseButton(), color.RGBA{45, 34, 25, 230}, ColorGold, 6)
}

func drawTechButton(screen *ebiten.Image, btn gameui.Button, bg color.RGBA, textColor color.Color, textOffsetY float64) {
	vector.FillRect(screen, float32(btn.X), float32(btn.Y), float32(btn.W), float32(btn.H), bg, false)
	vector.StrokeRect(screen, float32(btn.X), float32(btn.Y), float32(btn.W), float32(btn.H), 1, panelBorder, false)
	tw := MeasureText(btn.Label, FaceSmall)
	DrawText(screen, btn.Label, btn.X+btn.W/2-tw/2, btn.Y+textOffsetY, FaceSmall, textColor)
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
