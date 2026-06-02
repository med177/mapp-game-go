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

func pauseMenuPanelRect(itemCount int) (bx, by, bw, bh float64) {
	bw = 390
	bh = float64(itemCount)*64 + 110
	bx = ScreenWidth/2 - bw/2
	by = ScreenHeight/2 - bh/2
	return bx, by, bw, bh
}

func buildPauseMenuButtons(hasSave bool, settings Settings) []gameui.Button {
	items := buildPauseItems(hasSave)
	bx, by, bw, _ := pauseMenuPanelRect(len(items))
	startY := by + 68
	itemH := 64.0
	buttons := make([]gameui.Button, 0, len(items))
	for i, item := range items {
		y := startY + float64(i)*itemH
		label := item.label
		switch item.action {
		case ActionToggleMusic:
			label = "Müzik: " + boolLabel(settings.MusicOn)
		case ActionAdjustMusic:
			label = "Müzik Seviyesi: ◄ " + itoa(settings.MusicVolume) + "% ►"
		}
		btn := gameui.NewButton(bx+16, y-6, bw-32, itemH-12, label)
		btn.Enabled = !item.disabled
		buttons = append(buttons, btn)
	}
	return buttons
}

func buildPauseItems(hasSave bool) []pauseMenuItem {
	return []pauseMenuItem{
		{"Devam Et", ActionResume, false, 0},
		{"Müzik", ActionToggleMusic, false, 0},
		{"Müzik Seviyesi", ActionAdjustMusic, false, 10},
		{"Kaydet", ActionOpenSaveSelect, false, 0},
		{"Yükle", ActionLoadFromPause, !hasSave, 0},
		{"Ana Menü", ActionGoMainMenu, false, 0},
		{"Oyundan Çık", ActionQuit, false, 0},
	}
}

// DrawPauseMenu oyun içi duraklama menüsünü yarı saydam overlay üzerine çizer.
func DrawPauseMenu(screen *ebiten.Image, cursor int, hasSave bool, tick int, settings Settings) {
	// Karartma katmanı
	overlay := ebiten.NewImage(int(ScreenWidth), int(ScreenHeight))
	overlay.Fill(color.RGBA{0, 0, 0, 170})
	screen.DrawImage(overlay, nil)

	items := buildPauseItems(hasSave)

	bx, by, bw, bh := pauseMenuPanelRect(len(items))
	fbx, fby, fbw, fbh := float32(bx), float32(by), float32(bw), float32(bh)

	// Panel arka planı — animasyonlu hafif titreşen kenarlık
	vector.FillRect(screen, fbx, fby, fbw, fbh, color.RGBA{10, 8, 5, 240}, false)
	phase := float64(tick) / 90.0
	glow := uint8(140 + uint8(20*math.Abs(math.Sin(phase))))
	vector.StrokeRect(screen, fbx, fby, fbw, fbh, 2, color.RGBA{glow, glow - 30, 30, 255}, false)
	vector.StrokeRect(screen, fbx+4, fby+4, fbw-8, fbh-8, 1, color.RGBA{80, 65, 30, 180}, false)

	// Üst şerit
	vector.FillRect(screen, fbx, fby, fbw, 4, color.RGBA{200, 160, 50, 255}, false)

	// Başlık
	titleW := MeasureText("DURAKLANDI", FaceLarge)
	DrawText(screen, "DURAKLANDI",
		bx+bw/2-titleW/2,
		by+18,
		FaceLarge, color.RGBA{220, 190, 80, 255})

	sepY := fby + 52
	vector.StrokeLine(screen, fbx+20, sepY, fbx+fbw-20, sepY, 1, color.RGBA{100, 80, 35, 180}, false)

	// Menü maddeleri
	startY := by + 68
	itemH := 64.0
	buttons := buildPauseMenuButtons(hasSave, settings)

	for i, item := range items {
		y := startY + float64(i)*itemH
		isSelected := i == cursor

		if isSelected && !item.disabled {
			vector.FillRect(screen, fbx+16, float32(y)-6, fbw-32, float32(itemH)-12,
				color.RGBA{45, 35, 12, 200}, false)
			vector.StrokeRect(screen, fbx+16, float32(y)-6, fbw-32, float32(itemH)-12,
				1, color.RGBA{180, 145, 50, 200}, false)
		}

		col := menuItemColor(isSelected, item.disabled)
		prefix := "  "
		if isSelected && !item.disabled {
			prefix = "► "
		}
		label := buttons[i].Label
		tw := MeasureText(prefix+label, FaceLarge)
		DrawText(screen, prefix+label,
			bx+bw/2-tw/2,
			y+8, FaceLarge, col)
	}

	DrawTextCentered(screen, "Menü seçeneğini tıklayarak devam et",
		ScreenWidth/2, by+bh-22, FaceSmall, color.RGBA{80, 80, 80, 200})
}

// handlePauseMenuInput duraklama menüsü girişini işler.
func (r *Renderer) handlePauseMenuInput(input gameui.InputState) InputAction {
	items := buildPauseItems(r.HasSave)
	n := len(items)
	buttons := buildPauseMenuButtons(r.HasSave, r.CurrentSettings)
	for i, btn := range buttons {
		if btn.HitTest(input.MouseX, input.MouseY) {
			r.pauseCursor = i
			break
		}
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) {
		r.pauseCursor = (r.pauseCursor + 1) % n
		for items[r.pauseCursor].disabled {
			r.pauseCursor = (r.pauseCursor + 1) % n
		}
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) {
		r.pauseCursor = (r.pauseCursor - 1 + n) % n
		for items[r.pauseCursor].disabled {
			r.pauseCursor = (r.pauseCursor - 1 + n) % n
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
		for i, btn := range buttons {
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
