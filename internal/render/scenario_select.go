package render

import (
	"image/color"
	"strings"

	"mapp-game-go/internal/scenario"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ScenarioList senaryo seçim ekranında gösterilecek senaryolar.
// game.go tarafından doldurulur, render paketince okunur.
var ScenarioList []*scenario.Scenario

// DrawScenarioSelect senaryo seçim ekranını çizer.
func DrawScenarioSelect(screen *ebiten.Image, scenarios []*scenario.Scenario, cursor int) {
	screen.Fill(color.RGBA{6, 8, 14, 255})

	// Üst dekoratif çizgi
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)

	titleY := 40.0
	DrawTextCentered(screen, "MAPP — Senaryo Seç", ScreenWidth/2, titleY, FaceLarge, ColorYellow)
	DrawTextCentered(screen, "[↑↓] Seç   [Enter] Onayla   [Esc] Geri", ScreenWidth/2, titleY+30, FaceSmall, ColorGray)

	if len(scenarios) == 0 {
		DrawTextCentered(screen, "Senaryo bulunamadı!", ScreenWidth/2, ScreenHeight/2, FaceLarge, ColorRed)
		return
	}

	cardW := float32(560)
	cardH := float32(130)
	padY := float32(16)
	totalH := float32(len(scenarios))*(cardH+padY) - padY
	startY := float32(ScreenHeight)/2 - totalH/2 + 10
	startX := float32(ScreenWidth)/2 - cardW/2

	for i, sc := range scenarios {
		x := startX
		y := startY + float32(i)*(cardH+padY)
		isSelected := i == cursor

		bgCol := color.RGBA{22, 18, 12, 220}
		borderCol := color.RGBA{80, 65, 40, 200}
		if isSelected {
			bgCol = color.RGBA{50, 40, 18, 240}
			borderCol = color.RGBA{220, 180, 60, 255}
		}

		vector.FillRect(screen, x, y, cardW, cardH, bgCol, false)
		vector.StrokeRect(screen, x, y, cardW, cardH, 1.5, borderCol, false)

		// Seçim oku
		prefix := "  "
		if isSelected {
			prefix = "► "
		}
		nameCol := color.RGBA{200, 185, 140, 220}
		if isSelected {
			nameCol = ColorYellow
		}
		DrawText(screen, prefix+sc.Name, float64(x)+18, float64(y)+18, FaceLarge, nameCol)

		// Yıl etiketi
		yearStr := itoa(sc.Year) + " — " + monthName(sc.Month)
		DrawText(screen, yearStr, float64(x)+18, float64(y)+46, FaceSmall, color.RGBA{160, 140, 90, 200})

		// Açıklama (uzunsa kes)
		desc := sc.Description
		if len(desc) > 90 {
			desc = desc[:87] + "..."
		}
		// Açıklamayı satırlara sar
		lines := splitLines(desc, 72)
		for j, line := range lines {
			DrawText(screen, line, float64(x)+18, float64(y)+68+float64(j)*18, FaceSmall, color.RGBA{140, 125, 90, 180})
		}
	}

	vector.FillRect(screen, 0, float32(ScreenHeight)-3, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)
}

func monthName(m int) string {
	names := []string{"", "Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran",
		"Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık"}
	if m < 1 || m > 12 {
		return ""
	}
	return names[m]
}

// splitLines metni maxChars genişliğinde kelime bazlı satırlara böler.
func splitLines(text string, maxChars int) []string {
	words := strings.Fields(text)
	var lines []string
	current := ""
	for _, w := range words {
		if len(current)+len(w)+1 > maxChars {
			if current != "" {
				lines = append(lines, current)
			}
			current = w
		} else {
			if current == "" {
				current = w
			} else {
				current += " " + w
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// handleScenarioSelectInput senaryo seçim ekranı klavye ve fare girişini işler.
func (r *Renderer) handleScenarioSelectInput() InputAction {
	scenarios := ScenarioList
	if len(scenarios) == 0 {
		return InputAction{}
	}
	n := len(scenarios)

	mx, my := ebiten.CursorPosition()
	if i := r.scenarioHoverIndex(float64(mx), float64(my)); i >= 0 {
		r.scenarioCursor = i
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) {
		r.scenarioCursor = (r.scenarioCursor + 1) % n
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) {
		r.scenarioCursor = (r.scenarioCursor - 1 + n) % n
	}
	if r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace) {
		return InputAction{
			Kind:       ActionSelectScenario,
			BuildingID: scenarios[r.scenarioCursor].Path,
		}
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		return InputAction{Kind: ActionBack}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if i := r.scenarioHoverIndex(float64(mx), float64(my)); i >= 0 {
			return InputAction{
				Kind:       ActionSelectScenario,
				BuildingID: scenarios[i].Path,
			}
		}
	}
	return InputAction{}
}

func (r *Renderer) scenarioHoverIndex(mx, my float64) int {
	scenarios := ScenarioList
	if len(scenarios) == 0 {
		return -1
	}
	cardW := float64(560)
	cardH := float64(130)
	padY := float64(16)
	totalH := float64(len(scenarios))*(cardH+padY) - padY
	startY := ScreenHeight/2 - totalH/2 + 10
	startX := ScreenWidth/2 - cardW/2

	for i := range scenarios {
		x := startX
		y := startY + float64(i)*(cardH+padY)
		if mx >= x && mx <= x+cardW && my >= y && my <= y+cardH {
			return i
		}
	}
	return -1
}
