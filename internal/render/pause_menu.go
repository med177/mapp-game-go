package render

import (
	"image/color"
	"math"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type pauseMenuItem struct {
	label    string
	action   ActionKind
	disabled bool
	delta    int
}

type pauseMenuLayout struct {
	panelRect  gameui.Rect
	titleRect  gameui.Rect
	itemsRect  gameui.Rect
	footerRect gameui.Rect
}

func pauseMenuPanelRect(itemCount int) (bx, by, bw, bh float64) {
	r := pauseMenuLayoutForScreen(itemCount).panelRect
	return r.X, r.Y, r.W, r.H
}

func pauseMenuLayoutForScreen(itemCount int) pauseMenuLayout {
	panel := gameui.Rect{
		X: ScreenWidth/2 - 390/2,
		Y: ScreenHeight/2 - (float64(itemCount)*64+110)/2,
		W: 390,
		H: float64(itemCount)*64 + 110,
	}
	box := gameui.BoxFromRect(panel).InsetXY(16, 14)
	titleRect, box := box.CutTop(38, 16)
	footerRect, itemsBox := box.CutBottom(18, 0)
	return pauseMenuLayout{
		panelRect:  panel,
		titleRect:  titleRect,
		itemsRect:  itemsBox.Rect,
		footerRect: footerRect,
	}
}

func buildPauseMenuButtons(hasSave bool, settings Settings) []gameui.Button {
	items := buildPauseItems(hasSave)
	layout := pauseMenuLayoutForScreen(len(items))
	itemH := layout.itemsRect.H / float64(len(items))
	buttons := make([]gameui.Button, 0, len(items))
	for i, item := range items {
		y := layout.itemsRect.Y + float64(i)*itemH
		label := item.label
		switch item.action {
		case ActionToggleMusic:
			label = "Müzik: " + boolLabel(settings.MusicOn)
		case ActionAdjustMusic:
			label = "Müzik Seviyesi: " + itoa(settings.MusicVolume) + "%"
		}
		btn := gameui.NewButton(layout.itemsRect.X, y, layout.itemsRect.W, itemH-10, label)
		btn.Enabled = !item.disabled
		buttons = append(buttons, btn)
	}
	return buttons
}

func pauseMusicVolumeButtons(row gameui.Button, volume int) (gameui.Button, gameui.Button) {
	const buttonW = 34.0
	const buttonH = 32.0
	minus := gameui.NewButton(row.X+198, row.Y+(row.H-buttonH)/2, buttonW, buttonH, "-")
	plus := gameui.NewButton(row.X+row.W-44, row.Y+(row.H-buttonH)/2, buttonW, buttonH, "+")
	minus.Enabled = volume > 0
	plus.Enabled = volume < 100
	return minus, plus
}

func buildPauseItems(hasSave bool) []pauseMenuItem {
	return []pauseMenuItem{
		{"Devam Et", ActionResume, false, 0},
		{"Müzik", ActionToggleMusic, false, 0},
		{"Müzik Seviyesi", ActionAdjustMusic, false, 5},
		{"Kaydet", ActionOpenSaveSelect, false, 0},
		{"Yükle", ActionLoadFromPause, !hasSave, 0},
		{"Ana Menü", ActionGoMainMenu, false, 0},
		{"Oyundan Çık", ActionQuit, false, 0},
	}
}

// DrawPauseMenu oyun içi duraklama menüsünü yarı saydam overlay üzerine çizer.
func DrawPauseMenu(screen *ebiten.Image, cursor int, hasSave bool, tick int, settings Settings) {
	// Karartma katmanı
	drawUIOverlay(screen, color.RGBA{0, 0, 0, 170})

	items := buildPauseItems(hasSave)
	layout := pauseMenuLayoutForScreen(len(items))

	bx, by, bw, bh := layout.panelRect.X, layout.panelRect.Y, layout.panelRect.W, layout.panelRect.H
	fbx, fby, fbw, fbh := float32(bx), float32(by), float32(bw), float32(bh)

	// Panel arka planı — animasyonlu hafif titreşen kenarlık
	drawUIPanelRect(screen, layout.panelRect, color.RGBA{10, 8, 5, 240}, color.RGBA{10, 8, 5, 240}, 0)
	phase := float64(tick) / 90.0
	glow := uint8(140 + uint8(20*math.Abs(math.Sin(phase))))
	vector.StrokeRect(screen, fbx, fby, fbw, fbh, 2, color.RGBA{glow, glow - 30, 30, 255}, false)
	vector.StrokeRect(screen, fbx+4, fby+4, fbw-8, fbh-8, 1, color.RGBA{80, 65, 30, 180}, false)

	// Üst şerit
	vector.FillRect(screen, fbx, fby, fbw, 4, color.RGBA{200, 160, 50, 255}, false)

	// Başlık
	drawUILabel(screen, layout.titleRect, "DURAKLANDI", color.RGBA{220, 190, 80, 255}, gameui.TextLarge, gameui.TextAlignCenter)

	sepY := fby + 52
	vector.StrokeLine(screen, fbx+20, sepY, fbx+fbw-20, sepY, 1, color.RGBA{100, 80, 35, 180}, false)

	// Menü maddeleri
	itemH := layout.itemsRect.H / float64(len(items))
	buttons := buildPauseMenuButtons(hasSave, settings)

	for i, item := range items {
		y := layout.itemsRect.Y + float64(i)*itemH
		isSelected := i == cursor

		if isSelected && !item.disabled {
			vector.FillRect(screen, float32(layout.itemsRect.X), float32(y), float32(layout.itemsRect.W), float32(itemH)-10,
				color.RGBA{45, 35, 12, 200}, false)
			vector.StrokeRect(screen, float32(layout.itemsRect.X), float32(y), float32(layout.itemsRect.W), float32(itemH)-10,
				1, color.RGBA{180, 145, 50, 200}, false)
		}

		col := menuItemColor(isSelected, item.disabled)
		prefix := "  "
		if isSelected && !item.disabled {
			prefix = "► "
		}
		if item.action == ActionAdjustMusic {
			minus, plus := pauseMusicVolumeButtons(buttons[i], settings.MusicVolume)
			drawUILabel(screen, gameui.Rect{X: buttons[i].X, Y: y + 8, W: 190}, prefix+"Müzik Seviyesi:", col, gameui.TextLarge, gameui.TextAlignCenter)
			drawUILabel(screen, gameui.Rect{X: buttons[i].X + 238, Y: y + 8, W: 70}, itoa(settings.MusicVolume)+"%", ColorGold, gameui.TextLarge, gameui.TextAlignCenter)
			drawUIButtonWidget(screen, minus, tinyButtonStyle)
			drawUIButtonWidget(screen, plus, tinyButtonStyle)
		} else {
			label := buttons[i].Label
			drawUILabel(screen, gameui.Rect{X: layout.itemsRect.X, Y: y + 8, W: layout.itemsRect.W}, prefix+label, col, gameui.TextLarge, gameui.TextAlignCenter)
		}
	}

	drawUILabel(screen, layout.footerRect, "Menü seçeneğini tıklayarak devam et", ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
}

// handlePauseMenuInput duraklama menüsü girişini işler.
func (r *Renderer) handlePauseMenuInput(input gameui.InputState) InputAction {
	items := buildPauseItems(r.HasSave)
	n := len(items)
	buttons := buildPauseMenuButtons(r.HasSave, r.CurrentSettings)
	for i, btn := range buttons {
		if btn.HitTest(input.MouseX, input.MouseY) {
			if !items[i].disabled {
				r.setMenuCursor(&r.pauseCursor, i)
			} else {
				r.pauseCursor = i
			}
			break
		}
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) {
		next := (r.pauseCursor + 1) % n
		for items[next].disabled {
			next = (next + 1) % n
		}
		r.setMenuCursor(&r.pauseCursor, next)
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) {
		next := (r.pauseCursor - 1 + n) % n
		for items[next].disabled {
			next = (next - 1 + n) % n
		}
		r.setMenuCursor(&r.pauseCursor, next)
	}
	if r.keyJustPressed(ebiten.KeyTab) {
		next := focusButtonIndex(buttons, r.pauseCursor, ebiten.IsKeyPressed(ebiten.KeyShift))
		if next >= 0 && next < n {
			r.setMenuCursor(&r.pauseCursor, next)
		}
	}
	if items[r.pauseCursor].action == ActionAdjustMusic {
		if r.keyJustPressed(ebiten.KeyArrowLeft) {
			return InputAction{Kind: ActionAdjustMusic, Delta: -5}
		}
		if r.keyJustPressed(ebiten.KeyArrowRight) {
			return InputAction{Kind: ActionAdjustMusic, Delta: 5}
		}
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		return InputAction{Kind: ActionResume}
	}
	if r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace) {
		item := items[r.pauseCursor]
		if !item.disabled {
			return InputAction{Kind: item.action, Delta: item.delta}
		}
	}
	if input.LeftJustPressed {
		musicRow := buttons[2]
		minus, plus := pauseMusicVolumeButtons(musicRow, r.CurrentSettings.MusicVolume)
		if minus.HitTest(input.MouseX, input.MouseY) && minus.Enabled {
			r.setMenuCursor(&r.pauseCursor, 2)
			return InputAction{Kind: ActionAdjustMusic, Delta: -5}
		}
		if plus.HitTest(input.MouseX, input.MouseY) && plus.Enabled {
			r.setMenuCursor(&r.pauseCursor, 2)
			return InputAction{Kind: ActionAdjustMusic, Delta: 5}
		}
		for i, btn := range buttons {
			if items[i].action == ActionAdjustMusic {
				continue
			}
			if btn.HandleInput(input) && !items[i].disabled {
				return InputAction{Kind: items[i].action, Delta: items[i].delta}
			}
		}
	}
	return InputAction{}
}

// pauseMenuHoverIndex fareye göre hangi menü maddesinin üzerinde olduğunu döner.
func (r *Renderer) pauseMenuHoverIndex(mx, my float64) int {
	for i, btn := range buildPauseMenuButtons(r.HasSave, r.CurrentSettings) {
		if btn.HitTest(mx, my) {
			return i
		}
	}
	return -1
}
