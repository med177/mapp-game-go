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

var techCategoryOrder = []tech.Category{
	tech.CategoryMilitary,
	tech.CategoryEconomy,
	tech.CategoryDiplomacy,
	tech.CategoryNaval,
	tech.CategoryReligion,
}

type techEntry struct {
	t        *tech.Technology
	unlocked bool
	done     bool
}

func (r *Renderer) buildTechEntries(f *faction.Faction) []techEntry {
	var entries []techEntry
	for _, cat := range techCategoryOrder {
		var catTechs []*tech.Technology
		for _, t := range r.gs.TechTypes {
			if t.Category == cat {
				catTechs = append(catTechs, t)
			}
		}
		sort.Slice(catTechs, func(i, j int) bool { return catTechs[i].ID < catTechs[j].ID })
		for _, t := range catTechs {
			entries = append(entries, techEntry{
				t:        t,
				unlocked: tech.IsUnlocked(&f.Research, t),
				done:     f.Research.Completed[t.ID],
			})
		}
	}
	return entries
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

	entries := r.buildTechEntries(f)

	rowH := 36
	listY := int(py) + 80
	visibleRows := (int(ph) - 110) / rowH
	colW := int(pw) / 2

	if r.techCursor < 0 {
		r.techCursor = 0
	}
	if len(entries) > 0 && r.techCursor >= len(entries) {
		r.techCursor = len(entries) - 1
	}

	offset := 0
	if r.techCursor >= visibleRows {
		offset = r.techCursor - visibleRows + 1
	}

	for i := offset; i < len(entries) && i < offset+visibleRows; i++ {
		e := entries[i]
		rowY := float32(listY + (i-offset)*rowH)

		if i == r.techCursor {
			vector.FillRect(screen, px+10, rowY-2, pw-20, float32(rowH-2), color.RGBA{60, 60, 100, 180}, false)
		}

		var nameCol color.RGBA
		var statusStr string
		switch {
		case e.done:
			nameCol = color.RGBA{100, 220, 100, 255}
			statusStr = "Tamamlandi"
		case !e.unlocked:
			nameCol = color.RGBA{120, 120, 120, 255}
			statusStr = "Kilitli"
		case f.Research.ActiveID == e.t.ID:
			nameCol = color.RGBA{255, 220, 80, 255}
			statusStr = fmt.Sprintf("Arastiriliyor (%d tur)", f.Research.TurnsLeft)
		default:
			nameCol = color.RGBA{220, 200, 160, 255}
			statusStr = fmt.Sprintf("%dg / %d tur", e.t.GoldCost, e.t.TurnsRequired)
		}

		catLabel := techCategoryLabels[e.t.Category]
		DrawText(screen, fmt.Sprintf("[%s] %s", catLabel, e.t.NameTR), float64(px)+18, float64(rowY)+4, FaceMed, nameCol)
		DrawText(screen, e.t.DescriptionTR, float64(px)+18, float64(rowY)+18, FaceSmall, ColorGray)
		DrawText(screen, statusStr, float64(px)+float64(colW)+10, float64(rowY)+4, FaceSmall, ColorGold)

		if len(e.t.Requires) > 0 && !e.done {
			reqStr := "Gerekir: "
			for j, req := range e.t.Requires {
				if j > 0 {
					reqStr += ", "
				}
				if rt, ok := r.gs.TechTypes[req]; ok {
					reqStr += rt.NameTR
				} else {
					reqStr += req
				}
			}
			DrawText(screen, reqStr, float64(px)+float64(colW)+10, float64(rowY)+18, FaceSmall, color.RGBA{140, 120, 80, 255})
		}
	}

	hintY := float64(py) + float64(ph) - 18
	DrawText(screen, "Bir teknolojiyi tıklayarak araştır   Altin: "+fmt.Sprintf("%d", f.Gold),
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
	entries := r.buildTechEntries(f)

	// Fare hover → satır seç
	mx, my := ebiten.CursorPosition()
	px, py := float32(0), float32(0)
	pw := float32(ScreenWidth)
	rowH := 36
	listY := int(py) + 80
	visibleRows := (int(float32(ScreenHeight)) - 110) / rowH
	offset := 0
	if r.techCursor >= visibleRows {
		offset = r.techCursor - visibleRows + 1
	}
	for i := offset; i < len(entries) && i < offset+visibleRows; i++ {
		rowY := float32(listY + (i-offset)*rowH)
		if float64(my) >= float64(rowY-2) && float64(my) <= float64(rowY+float32(rowH)-2) &&
			float64(mx) >= float64(px+10) && float64(mx) <= float64(px+pw-20) {
			r.techCursor = i
			break
		}
	}

	// Sol tık → araştırmayı başlat (uygunsa)
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if techCloseHit(float64(mx), float64(my)) {
			r.showTech = false
			return InputAction{}
		}
		if r.techCursor < len(entries) {
			e := entries[r.techCursor]
			if e.unlocked && !e.done && f.Research.ActiveID == "" {
				return InputAction{Kind: ActionResearch, BuildingID: e.t.ID}
			}
		}
	}

	if r.keyJustPressed(ebiten.KeyArrowUp) && r.techCursor > 0 {
		r.techCursor--
	}
	if r.keyJustPressed(ebiten.KeyArrowDown) && r.techCursor < len(entries)-1 {
		r.techCursor++
	}
	if r.keyJustPressed(ebiten.KeyEnter) && r.techCursor < len(entries) {
		e := entries[r.techCursor]
		if e.unlocked && !e.done && f.Research.ActiveID == "" {
			return InputAction{Kind: ActionResearch, BuildingID: e.t.ID}
		}
	}
	return InputAction{}
}
