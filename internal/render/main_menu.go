package render

import (
	"image/color"
	"math"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type menuItem struct {
	label    string
	action   ActionKind
	disabled bool
}

func buildMainMenuButtons(hasSave bool, hasAutoSave bool) []gameui.Button {
	items := buildMenuItems(hasSave, hasAutoSave)
	itemH := 52.0
	startY := ScreenHeight/2 - float64(len(items))*itemH/2 + 20
	barW := 280.0
	barX := ScreenWidth/2 - barW/2
	buttons := make([]gameui.Button, 0, len(items))
	for i, item := range items {
		y := startY + float64(i)*itemH
		btn := gameui.NewButton(barX, y-6, barW, itemH-8, item.label)
		btn.Enabled = !item.disabled
		buttons = append(buttons, btn)
	}
	return buttons
}

// DrawMainMenu ana menü ekranını çizer.
func DrawMainMenu(screen *ebiten.Image, cursor int, hasSave bool, hasAutoSave bool, tick int) {
	screen.Fill(color.RGBA{8, 10, 18, 255})

	// Animasyonlu arka plan — yavaş titreşen renk şeritleri
	for i := 0; i < 6; i++ {
		phase := float64(tick)/180.0 + float64(i)*0.4
		alpha := uint8(18 + 10*math.Sin(phase))
		y := float32(float64(i) * ScreenHeight / 6)
		vector.FillRect(screen, 0, y, float32(ScreenWidth), float32(ScreenHeight/6), color.RGBA{20, 30, 60, alpha}, false)
	}

	// Üst dekoratif çizgi
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)
	vector.FillRect(screen, 0, float32(ScreenHeight)-3, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)

	// Başlık
	titleY := ScreenHeight/2 - 200
	drawUILabel(screen, gameui.Rect{X: 0, Y: titleY, W: ScreenWidth}, "MAPP", ColorYellow, gameui.TextLarge, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: 0, Y: titleY + 34, W: ScreenWidth}, "Orta Çağ Strateji", color.RGBA{180, 160, 100, 200}, gameui.TextSmall, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: 0, Y: titleY + 52, W: ScreenWidth}, "1300 – 1600", color.RGBA{140, 120, 80, 180}, gameui.TextSmall, gameui.TextAlignCenter)

	// Ayraç
	sepY := float32(titleY + 80)
	vector.FillRect(screen, float32(ScreenWidth/2)-120, sepY, 240, 1, color.RGBA{120, 100, 50, 180}, false)

	// Menü maddeleri
	items := buildMenuItems(hasSave, hasAutoSave)
	itemH := 52.0
	startY := ScreenHeight/2 - float64(len(items))*itemH/2 + 20

	for i, item := range items {
		y := startY + float64(i)*itemH
		isSelected := i == cursor

		// Seçili satır vurgusu
		if isSelected && !item.disabled {
			barW := float32(280)
			barX := float32(ScreenWidth/2) - barW/2
			vector.FillRect(screen, barX, float32(y)-6, barW, float32(itemH)-8, color.RGBA{50, 40, 15, 180}, false)
			vector.StrokeRect(screen, barX, float32(y)-6, barW, float32(itemH)-8, 1, color.RGBA{200, 160, 60, 180}, false)
		}

		col := menuItemColor(isSelected, item.disabled)
		prefix := "  "
		if isSelected && !item.disabled {
			prefix = "► "
		}
		drawUILabel(screen, gameui.Rect{X: 0, Y: y + 8, W: ScreenWidth}, prefix+item.label, col, gameui.TextLarge, gameui.TextAlignCenter)
	}

	// Alt bilgi
	drawUILabel(screen, gameui.Rect{X: 0, Y: ScreenHeight - 30, W: ScreenWidth}, "Menü seçeneğini tıklayarak devam et", color.RGBA{80, 80, 80, 200}, gameui.TextSmall, gameui.TextAlignCenter)
}

func buildMenuItems(hasSave bool, hasAutoSave bool) []menuItem {
	return []menuItem{
		{"Yeni Oyun", ActionNewGame, false},
		{"Devam et", ActionContinue, !hasAutoSave},
		{"Kayıttan Yükle", ActionOpenLoadSelect, !hasSave},
		{"Ayarlar", ActionOpenSettings, false},
		{"Çıkış", ActionQuit, false},
	}
}

func menuItemColor(selected, disabled bool) color.RGBA {
	if disabled {
		return color.RGBA{60, 60, 60, 180}
	}
	if selected {
		return color.RGBA{255, 220, 80, 255}
	}
	return color.RGBA{200, 185, 140, 220}
}

// handleMainMenuInput ana menü klavye ve fare girişini işler.
func (r *Renderer) handleMainMenuInput(hasSave bool, hasAutoSave bool, input gameui.InputState) InputAction {
	items := buildMenuItems(hasSave, hasAutoSave)
	n := len(items)
	buttons := buildMainMenuButtons(hasSave, hasAutoSave)

	// Hover ile satır vurgusunu güncelle
	for i, btn := range buttons {
		if btn.HitTest(input.MouseX, input.MouseY) {
			r.factionCursor = i
			break
		}
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) {
		r.factionCursor = (r.factionCursor + 1) % n
		for items[r.factionCursor].disabled {
			r.factionCursor = (r.factionCursor + 1) % n
		}
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) {
		r.factionCursor = (r.factionCursor - 1 + n) % n
		for items[r.factionCursor].disabled {
			r.factionCursor = (r.factionCursor - 1 + n) % n
		}
	}
	if r.keyJustPressed(ebiten.KeyTab) {
		next := focusButtonIndex(buttons, r.factionCursor, ebiten.IsKeyPressed(ebiten.KeyShift))
		if next >= 0 && next < n {
			r.factionCursor = next
		}
	}
	if r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace) {
		item := items[r.factionCursor]
		if !item.disabled {
			return InputAction{Kind: item.action}
		}
	}
	if input.LeftJustPressed {
		for i, btn := range buttons {
			if btn.HandleInput(input) && !items[i].disabled {
				return InputAction{Kind: items[i].action}
			}
		}
	}
	if r.keyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	return InputAction{}
}
