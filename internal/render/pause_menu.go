package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type pauseMenuItem struct {
	label    string
	action   ActionKind
	disabled bool
}

func buildPauseItems(hasSave bool) []pauseMenuItem {
	return []pauseMenuItem{
		{"Devam Et", ActionResume, false},
		{"Kaydet", ActionOpenSaveSelect, false},
		{"Yükle", ActionLoadFromPause, !hasSave},
		{"Ana Menü", ActionGoMainMenu, false},
		{"Oyundan Çık", ActionQuit, false},
	}
}

// DrawPauseMenu oyun içi duraklama menüsünü yarı saydam overlay üzerine çizer.
func DrawPauseMenu(screen *ebiten.Image, cursor int, hasSave bool, tick int) {
	// Karartma katmanı
	overlay := ebiten.NewImage(int(ScreenWidth), int(ScreenHeight))
	overlay.Fill(color.RGBA{0, 0, 0, 170})
	screen.DrawImage(overlay, nil)

	items := buildPauseItems(hasSave)

	bw := float32(360)
	bh := float32(float64(len(items))*64 + 110)
	bx := float32(ScreenWidth/2) - bw/2
	by := float32(ScreenHeight/2) - bh/2

	// Panel arka planı — animasyonlu hafif titreşen kenarlık
	vector.FillRect(screen, bx, by, bw, bh, color.RGBA{10, 8, 5, 240}, false)
	phase := float64(tick) / 90.0
	glow := uint8(140 + uint8(20*math.Abs(math.Sin(phase))))
	vector.StrokeRect(screen, bx, by, bw, bh, 2, color.RGBA{glow, glow - 30, 30, 255}, false)
	vector.StrokeRect(screen, bx+4, by+4, bw-8, bh-8, 1, color.RGBA{80, 65, 30, 180}, false)

	// Üst şerit
	vector.FillRect(screen, bx, by, bw, 4, color.RGBA{200, 160, 50, 255}, false)

	// Başlık
	titleW := MeasureText("DURAKLANDI", FaceLarge)
	DrawText(screen, "DURAKLANDI",
		float64(bx)+float64(bw)/2-titleW/2,
		float64(by)+18,
		FaceLarge, color.RGBA{220, 190, 80, 255})

	sepY := by + 52
	vector.StrokeLine(screen, bx+20, sepY, bx+bw-20, sepY, 1, color.RGBA{100, 80, 35, 180}, false)

	// Menü maddeleri
	startY := float64(by) + 68
	itemH := 64.0

	for i, item := range items {
		y := startY + float64(i)*itemH
		isSelected := i == cursor

		if isSelected && !item.disabled {
			vector.FillRect(screen, bx+16, float32(y)-6, bw-32, float32(itemH)-12,
				color.RGBA{45, 35, 12, 200}, false)
			vector.StrokeRect(screen, bx+16, float32(y)-6, bw-32, float32(itemH)-12,
				1, color.RGBA{180, 145, 50, 200}, false)
		}

		col := menuItemColor(isSelected, item.disabled)
		prefix := "  "
		if isSelected && !item.disabled {
			prefix = "► "
		}
		tw := MeasureText(prefix+item.label, FaceLarge)
		DrawText(screen, prefix+item.label,
			float64(bx)+float64(bw)/2-tw/2,
			y+8, FaceLarge, col)
	}

	DrawTextCentered(screen, "[↑↓] Seç   [Enter] Onayla   [Esc] Devam Et",
		ScreenWidth/2, float64(by)+float64(bh)-22, FaceSmall, color.RGBA{80, 80, 80, 200})
}

// handlePauseMenuInput duraklama menüsü girişini işler.
func (r *Renderer) handlePauseMenuInput() InputAction {
	items := buildPauseItems(r.HasSave)
	n := len(items)

	mx, my := ebiten.CursorPosition()
	if i := r.pauseMenuHoverIndex(float64(mx), float64(my)); i >= 0 {
		r.pauseCursor = i
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
	if r.keyJustPressed(ebiten.KeyEscape) {
		return InputAction{Kind: ActionResume}
	}
	if r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace) {
		item := items[r.pauseCursor]
		if !item.disabled {
			return InputAction{Kind: item.action}
		}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if i := r.pauseMenuHoverIndex(float64(mx), float64(my)); i >= 0 && !items[i].disabled {
			return InputAction{Kind: items[i].action}
		}
	}
	return InputAction{}
}

// pauseMenuHoverIndex fareye göre hangi menü maddesinin üzerinde olduğunu döner.
func (r *Renderer) pauseMenuHoverIndex(mx, my float64) int {
	items := buildPauseItems(r.HasSave)
	bw := float64(360)
	bh := float64(len(items))*64 + 110
	bx := ScreenWidth/2 - bw/2
	by := ScreenHeight/2 - bh/2
	startY := by + 68
	itemH := 64.0

	for i := range items {
		y := startY + float64(i)*itemH
		if mx >= bx+16 && mx <= bx+bw-16 && my >= y-6 && my <= y+itemH-18 {
			return i
		}
	}
	return -1
}
