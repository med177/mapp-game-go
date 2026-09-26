package render

import (
	"image"
	"image/color"
	"math"
	"sync"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type menuItem struct {
	label    string
	action   ActionKind
	disabled bool
}

var (
	mainMenuBackgroundOnce sync.Once
	mainMenuBackground     *ebiten.Image
	mainMenuGradientImage  *ebiten.Image
	mainMenuGradientWidth  int
)

func buildMainMenuButtons(hasSave bool, hasAutoSave bool, editModeEnabled bool) []gameui.Button {
	items := buildMenuItems(hasSave, hasAutoSave, editModeEnabled)
	startY := mainMenuItemStartY(mainMenuVisualRowCount(items))
	barW := 280.0
	barX := ScreenWidth/2 - barW/2
	buttons := make([]gameui.Button, 0, len(items))
	for i, item := range items {
		y := mainMenuItemY(startY, items, i)
		btn := gameui.NewButton(barX, y-6, barW, mainMenuButtonHeight-8, item.label)
		btn.Enabled = !item.disabled
		buttons = append(buttons, btn)
	}
	return buttons
}

// DrawMainMenu ana menü ekranını çizer.
func DrawMainMenu(screen *ebiten.Image, cursor int, hasSave bool, hasAutoSave bool, editModeEnabled bool, tick int) {
	if background := mainMenuBackgroundImage(); background != nil {
		drawMainMenuBackground(screen, background)
		// Yalnızca menü ekseninde yumuşak bir siyah geçiş kullanılır; görselin
		// geri kalanındaki renkler genel bir karartmayla soldurulmaz.
		drawMainMenuFocusGradient(screen)
	} else {
		screen.Fill(color.RGBA{8, 10, 18, 255})
		// Asset yoksa eski animasyonlu koyu arka plan korunur.
		for i := 0; i < 6; i++ {
			phase := float64(tick)/180.0 + float64(i)*0.4
			alpha := uint8(18 + 10*math.Sin(phase))
			y := float32(float64(i) * ScreenHeight / 6)
			vector.FillRect(screen, 0, y, float32(ScreenWidth), float32(ScreenHeight/6), color.RGBA{20, 30, 60, alpha}, false)
		}
	}

	// Üst dekoratif çizgi
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)
	vector.FillRect(screen, 0, float32(ScreenHeight)-3, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)

	// Başlık
	titleY := ScreenHeight/2 - 200
	drawUILabel(screen, gameui.Rect{X: 0, Y: titleY, W: ScreenWidth}, "MAPP GAME", ColorYellow, gameui.TextLarge, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: 0, Y: titleY + 34, W: ScreenWidth}, "Harita Strateji", color.RGBA{180, 160, 100, 200}, gameui.TextSmall, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: 0, Y: titleY + 52, W: ScreenWidth}, "Oyunu", color.RGBA{140, 120, 80, 180}, gameui.TextSmall, gameui.TextAlignCenter)

	// Ayraç
	sepY := float32(titleY + 80)
	vector.FillRect(screen, float32(ScreenWidth/2)-120, sepY, 240, 1, color.RGBA{120, 100, 50, 180}, false)

	// Menü maddeleri
	items := buildMenuItems(hasSave, hasAutoSave, editModeEnabled)
	startY := mainMenuItemStartY(mainMenuVisualRowCount(items))

	for i, item := range items {
		y := mainMenuItemY(startY, items, i)
		isSelected := i == cursor

		// Seçili satır vurgusu
		if isSelected && !item.disabled {
			barW := float32(280)
			barX := float32(ScreenWidth/2) - barW/2
			vector.FillRect(screen, barX, float32(y)-6, barW, float32(mainMenuButtonHeight-8), color.RGBA{50, 40, 15, 180}, false)
			vector.StrokeRect(screen, barX, float32(y)-6, barW, float32(mainMenuButtonHeight-8), 1, color.RGBA{200, 160, 60, 180}, false)
		}

		col := menuItemColor(isSelected, item.disabled)
		prefix := "  "
		if isSelected && !item.disabled {
			prefix = "► "
		}
		drawUILabel(screen, gameui.Rect{X: 0, Y: y + 8, W: ScreenWidth}, prefix+item.label, col, gameui.TextLarge, gameui.TextAlignCenter)
	}

	// Alt bilgi
	drawUILabel(screen, gameui.Rect{X: 0, Y: ScreenHeight - 30, W: ScreenWidth}, "Versiyon 1.0.0	Alfa", color.RGBA{80, 80, 80, 200}, gameui.TextSmall, gameui.TextAlignCenter)
}

func mainMenuBackgroundImage() *ebiten.Image {
	mainMenuBackgroundOnce.Do(func() {
		mainMenuBackground = tryLoadImage("assets/images/main_menu_bg.png")
	})
	return mainMenuBackground
}

func drawMainMenuBackground(screen, background *ebiten.Image) {
	if background == nil || background.Bounds().Dx() <= 0 || background.Bounds().Dy() <= 0 {
		screen.Fill(color.RGBA{8, 10, 18, 255})
		return
	}

	// Görsel ekranı tamamen kaplar; farklı pencere oranlarında kenarlardan
	// kırpılır ve menü arkasında siyah bar oluşmaz.
	drawUIImageCover(screen, background)
}

func drawMainMenuFocusGradient(screen *ebiten.Image) {
	gradientWidth := int(math.Round(ScreenWidth))
	if gradientWidth <= 0 {
		return
	}
	if mainMenuGradientImage == nil || mainMenuGradientWidth != gradientWidth {
		const maxGradientAlpha = 170

		// Tek piksellik yatay alfa dokusu, dikdörtgen şeritlerin oluşturduğu
		// çizgileri ortadan kaldırır. Görsel ekran yüksekliğine ölçeklenir.
		gradient := image.NewRGBA(image.Rect(0, 0, gradientWidth, 1))
		halfWidth := math.Min(240, math.Max(160, ScreenWidth*0.17))
		for x := 0; x < gradientWidth; x++ {
			distance := math.Abs((float64(x) + 0.5 - ScreenWidth/2) / halfWidth)
			if distance >= 1 {
				continue
			}
			fade := 1 - distance*distance
			alpha := uint8(float64(maxGradientAlpha) * fade)
			gradient.SetRGBA(x, 0, color.RGBA{0, 0, 0, alpha})
		}
		mainMenuGradientImage = ebiten.NewImageFromImage(gradient)
		mainMenuGradientWidth = gradientWidth
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(ScreenWidth/float64(gradientWidth), ScreenHeight)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(mainMenuGradientImage, op)
}

const (
	mainMenuButtonHeight = 52.0
	mainMenuButtonGap    = 52.0
)

func mainMenuItemStartY(itemCount int) float64 {
	centeredY := ScreenHeight/2 - float64(itemCount)*mainMenuButtonHeight/2 + 20
	separatorY := ScreenHeight/2 - 200 + 80
	minimumY := separatorY + 24
	if centeredY < minimumY {
		return minimumY
	}
	return centeredY
}

func mainMenuVisualRowCount(items []menuItem) int {
	rows := len(items)
	for _, item := range items {
		if item.action == ActionQuit {
			rows++
		}
	}
	return rows
}

func mainMenuItemY(startY float64, items []menuItem, index int) float64 {
	y := startY + float64(index)*mainMenuButtonHeight
	for i := 0; i < index; i++ {
		if items[i].action == ActionQuit {
			y += mainMenuButtonGap
		}
	}
	return y
}

func buildMenuItems(hasSave bool, hasAutoSave bool, editModeEnabled bool) []menuItem {
	items := make([]menuItem, 0, 6)
	items = append(items,
		menuItem{"Yeni Oyun", ActionNewGame, false},
		menuItem{"Devam et", ActionContinue, !hasAutoSave},
		menuItem{"Kayıttan Yükle", ActionOpenLoadSelect, !hasSave},
		menuItem{"Ayarlar", ActionOpenSettings, false},
		menuItem{"Çıkış", ActionQuit, false},
	)
	if editModeEnabled {
		items = append(items, menuItem{"EDIT MODE", ActionEditMode, false})
	}
	return items
}

// InitialMainMenuCursor yeni açılan ana menüdeki ilk seçimi belirler.
func InitialMainMenuCursor(hasAutoSave bool, editModeEnabled bool) int {
	if hasAutoSave {
		return 1
	}
	return 0
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
	items := buildMenuItems(hasSave, hasAutoSave, r.EditModeEnabled)
	n := len(items)
	buttons := buildMainMenuButtons(hasSave, hasAutoSave, r.EditModeEnabled)

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
		r.toggleFullscreen()
	}
	return InputAction{}
}
