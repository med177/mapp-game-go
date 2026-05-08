package render

import (
	"fmt"
	"image/color"
	"sort"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"

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

	levels := r.buildTechTree(f)

	// Ağaç çizimi için koordinatlar
	treeStartY := 80.0
	levelHeight := 120.0
	nodeWidth := 180.0
	nodeHeight := 60.0

	layoutTechTree(levels, float64(ScreenWidth), nodeWidth, nodeHeight, treeStartY, levelHeight)

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

			// Düğüm arka planı
			vector.FillRect(screen, float32(node.x-nodeWidth/2), float32(node.y-nodeHeight/2),
				float32(nodeWidth-4), float32(nodeHeight-4), nodeColor, false)

			// Düğüm çerçevesi
			vector.StrokeRect(screen, float32(node.x-nodeWidth/2), float32(node.y-nodeHeight/2),
				float32(nodeWidth-4), float32(nodeHeight-4), 2, color.RGBA{255, 255, 255, 255}, false)

			// Teknoloji adı
			nameY := node.y - nodeHeight/2 + 8
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
				costY := node.y + nodeHeight/2 - 20
				costStr := fmt.Sprintf("%dg/%dt", node.t.GoldCost, node.t.TurnsRequired)
				DrawTextCentered(screen, costStr, node.x, costY, FaceSmall, ColorGold)
			}

			// Tamamlandı tik rozeti
			if node.done {
				badgeW := 24.0
				badgeH := 18.0
				badgeX := node.x + nodeWidth/2 - badgeW - 8
				badgeY := node.y - nodeHeight/2 + 8
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
										float32(reqNode.x), float32(reqNode.y+nodeHeight/2),
										float32(node.x), float32(node.y-nodeHeight/2),
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
	x, y, w, h := techCloseRect()
	vector.FillRect(screen, x, y, w, h, color.RGBA{45, 34, 25, 230}, false)
	vector.StrokeRect(screen, x, y, w, h, 1, panelBorder, false)
	tw := MeasureText("X", FaceSmall)
	DrawText(screen, "X", float64(x)+float64(w)/2-tw/2, float64(y)+6, FaceSmall, ColorGold)
}

func techCloseHit(mx, my float64) bool {
	x, y, w, h := techCloseRect()
	return mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h)
}

// handleTechInput teknoloji paneli klavye ve fare girişlerini işler.
func (r *Renderer) handleTechInput(f *faction.Faction) InputAction {
	if r.gs.TechTypes == nil {
		return InputAction{}
	}

	levels := r.buildTechTree(f)

	// Ağaç pozisyonlarını yeniden hesapla, böylece tıklama doğru çalışır
	treeStartY := 80.0
	levelHeight := 120.0
	nodeWidth := 180.0
	nodeHeight := 60.0
	layoutTechTree(levels, float64(ScreenWidth), nodeWidth, nodeHeight, treeStartY, levelHeight)

	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)

	// Close button kontrolü
	if techCloseHit(fx, fy) && r.mouseJustPressed(ebiten.MouseButtonLeft) {
		r.showTech = false
		return InputAction{}
	}

	// Ağaç düğümlerine tıklama
	for _, levelNodes := range levels {
		for _, node := range levelNodes {
			nodeLeft := node.x - nodeWidth/2
			nodeRight := node.x + nodeWidth/2
			nodeTop := node.y - nodeHeight/2
			nodeBottom := node.y + nodeHeight/2

			if fx >= nodeLeft && fx <= nodeRight && fy >= nodeTop && fy <= nodeBottom {
				if r.mouseJustPressed(ebiten.MouseButtonLeft) {
					if node.unlocked && !node.done {
						if f.Research.ActiveID == node.t.ID {
							return InputAction{Kind: ActionCancelResearch}
						} else {
							return InputAction{Kind: ActionResearch, BuildingID: node.t.ID}
						}
					}
				}
				break
			}
		}
	}

	return InputAction{}
}
