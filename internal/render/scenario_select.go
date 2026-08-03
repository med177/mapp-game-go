package render

import (
	"image/color"
	"strings"

	"mapp-game-go/internal/scenario"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

// ScenarioList senaryo seçim ekranında gösterilecek senaryolar.
// game.go tarafından doldurulur, render paketince okunur.
var ScenarioList []*scenario.Scenario

func buildScenarioCardButtons(scenarios []*scenario.Scenario) []gameui.Button {
	cardW := 560.0
	cardH := 130.0
	padY := 16.0
	stack := centeredStackRect(len(scenarios), cardW, cardH, padY, 20)
	buttons := make([]gameui.Button, 0, len(scenarios))
	for i, sc := range scenarios {
		r := stackItemRect(stack, cardH, padY, i)
		buttons = append(buttons, gameui.NewButton(r.X, r.Y, r.W, r.H, sc.Name))
	}
	return buttons
}

// DrawScenarioSelect senaryo seçim ekranını çizer.
func DrawScenarioSelect(screen *ebiten.Image, scenarios []*scenario.Scenario, cursor int) {
	if background := mainMenuBackgroundImage(); background != nil {
		drawUIImageCover(screen, background)
		// Görsel kartların ve üst başlığın arkasında kalır; kartların kendi
		// koyu yüzeyi korunurken ekranın tamamı hafifçe karartılır.
		drawUIOverlay(screen, color.RGBA{0, 0, 0, 100})
		drawUIScreenChromeOverlay(screen, "MAPP — Senaryo Seç", "Senaryo kartını seçmek için tıkla")
	} else {
		drawUIScreenChrome(screen, color.RGBA{6, 8, 14, 255}, "MAPP — Senaryo Seç", "Senaryo kartını seçmek için tıkla")
	}
	drawBackButton(screen)

	if len(scenarios) == 0 {
		drawUILabel(screen, gameui.Rect{X: 0, Y: ScreenHeight / 2, W: ScreenWidth}, "Senaryo bulunamadı!", ColorRed, gameui.TextLarge, gameui.TextAlignCenter)
		return
	}

	cardW := 560.0
	cardH := 130.0
	padY := 16.0
	stack := centeredStackRect(len(scenarios), cardW, cardH, padY, 20)

	for i, sc := range scenarios {
		rect := stackItemRect(stack, cardH, padY, i)
		x := float32(rect.X)
		y := float32(rect.Y)
		isSelected := i == cursor

		bgCol := color.RGBA{22, 18, 12, 220}
		borderCol := color.RGBA{80, 65, 40, 200}
		if isSelected {
			bgCol = color.RGBA{50, 40, 18, 240}
			borderCol = color.RGBA{220, 180, 60, 255}
		}

		drawUICardRect(screen, rect, bgCol, borderCol, 1.5)

		// Seçim oku
		prefix := "  "
		if isSelected {
			prefix = "► "
		}
		nameCol := color.RGBA{200, 185, 140, 220}
		if isSelected {
			nameCol = ColorYellow
		}
		drawUILabel(screen, gameui.Rect{X: float64(x) + 18, Y: float64(y) + 18}, prefix+sc.Name, nameCol, gameui.TextLarge, gameui.TextAlignStart)

		// Yıl etiketi
		yearStr := itoa(sc.Year) + " — " + monthName(sc.Month)
		drawUILabel(screen, gameui.Rect{X: float64(x) + 18, Y: float64(y) + 46}, yearStr, color.RGBA{160, 140, 90, 200}, gameui.TextSmall, gameui.TextAlignStart)

		// Açıklama (uzunsa kes)
		desc := sc.Description
		if len(desc) > 90 {
			desc = desc[:87] + "..."
		}
		// Açıklamayı satırlara sar
		lines := splitLines(desc, 72)
		for j, line := range lines {
			drawUILabel(screen, gameui.Rect{X: float64(x) + 18, Y: float64(y) + 68 + float64(j)*18}, line, color.RGBA{140, 125, 90, 180}, gameui.TextSmall, gameui.TextAlignStart)
		}
	}
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
func (r *Renderer) handleScenarioSelectInput(input gameui.InputState) InputAction {
	scenarios := ScenarioList
	if len(scenarios) == 0 {
		return InputAction{}
	}
	n := len(scenarios)
	buttons := buildScenarioCardButtons(scenarios)

	for i, btn := range buttons {
		if btn.HitTest(input.MouseX, input.MouseY) {
			r.scenarioCursor = i
			break
		}
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) {
		r.scenarioCursor = (r.scenarioCursor + 1) % n
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) {
		r.scenarioCursor = (r.scenarioCursor - 1 + n) % n
	}
	if r.keyJustPressed(ebiten.KeyTab) {
		next := focusButtonIndex(buttons, r.scenarioCursor, ebiten.IsKeyPressed(ebiten.KeyShift))
		if next >= 0 && next < n {
			r.scenarioCursor = next
		}
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
	if input.LeftJustPressed {
		if buildBackButton().HandleInput(input) {
			return InputAction{Kind: ActionBack}
		}
		for i, btn := range buttons {
			if btn.HandleInput(input) {
				return InputAction{
					Kind:       ActionSelectScenario,
					BuildingID: scenarios[i].Path,
				}
			}
		}
	}
	return InputAction{}
}

func (r *Renderer) scenarioHoverIndex(mx, my float64) int {
	for i, btn := range buildScenarioCardButtons(ScenarioList) {
		if btn.HitTest(mx, my) {
			return i
		}
	}
	return -1
}
