package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Settings oyun ayarlarını tutar — renderer aracılığıyla game'e iletilir.
type Settings struct {
	Difficulty int // 1=Kolay 2=Normal 3=Zor
	MusicOn    bool
	SoundOn    bool
}

func DefaultSettings() Settings {
	return Settings{Difficulty: 2, MusicOn: true, SoundOn: true}
}

var difficultyLabels = []string{"", "Kolay", "Normal", "Zor"}

// DrawSettingsScreen ayarlar ekranını çizer.
func DrawSettingsScreen(screen *ebiten.Image, s Settings, cursor int) {
	screen.Fill(color.RGBA{8, 10, 18, 255})
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)

	DrawTextCentered(screen, "[ AYARLAR ]", ScreenWidth/2, 60, FaceLarge, ColorYellow)

	type row struct {
		label string
		value string
	}
	rows := []row{
		{"Zorluk", difficultyLabels[s.Difficulty]},
		{"Müzik", boolLabel(s.MusicOn)},
		{"Ses Efektleri", boolLabel(s.SoundOn)},
		{"← Geri Dön", ""},
	}

	rowH := 60.0
	startY := ScreenHeight/2 - float64(len(rows))*rowH/2

	for i, r := range rows {
		y := startY + float64(i)*rowH
		isSelected := i == cursor

		if isSelected {
			bw := float32(500)
			bx := float32(ScreenWidth/2) - bw/2
			vector.FillRect(screen, bx, float32(y)-8, bw, float32(rowH)-4, color.RGBA{50, 40, 15, 200}, false)
			vector.StrokeRect(screen, bx, float32(y)-8, bw, float32(rowH)-4, 1, color.RGBA{200, 160, 60, 200}, false)
		}

		col := ColorGray
		if isSelected {
			col = ColorYellow
		}
		DrawText(screen, r.label, ScreenWidth/2-220, y+6, FaceLarge, col)
		if r.value != "" {
			DrawText(screen, "◄  "+r.value+"  ►", ScreenWidth/2+60, y+6, FaceLarge, ColorGold)
		}
	}

	DrawTextCentered(screen, "Fareyle seç / değiştir", ScreenWidth/2, ScreenHeight-30, FaceSmall, color.RGBA{80, 80, 80, 200})
}

func boolLabel(b bool) string {
	if b {
		return "Açık"
	}
	return "Kapalı"
}

// handleSettingsInput ayarlar ekranı girişini işler.
func (r *Renderer) handleSettingsInput(s *Settings) InputAction {
	rowCount := 4 // zorluk, müzik, ses, geri dön
	mx, my := ebiten.CursorPosition()
	if i := r.settingsHoverIndex(float64(mx), float64(my)); i >= 0 {
		r.factionCursor = i
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) {
		r.factionCursor = (r.factionCursor + 1) % rowCount
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) {
		r.factionCursor = (r.factionCursor - 1 + rowCount) % rowCount
	}

	switch r.factionCursor {
	case 0: // Zorluk
		if r.keyJustPressed(ebiten.KeyArrowRight) && s.Difficulty < 3 {
			s.Difficulty++
		}
		if r.keyJustPressed(ebiten.KeyArrowLeft) && s.Difficulty > 1 {
			s.Difficulty--
		}
	case 1: // Müzik
		if r.keyJustPressed(ebiten.KeyArrowLeft) || r.keyJustPressed(ebiten.KeyArrowRight) || r.keyJustPressed(ebiten.KeyEnter) {
			s.MusicOn = !s.MusicOn
		}
	case 2: // Ses efektleri
		if r.keyJustPressed(ebiten.KeyArrowLeft) || r.keyJustPressed(ebiten.KeyArrowRight) || r.keyJustPressed(ebiten.KeyEnter) {
			s.SoundOn = !s.SoundOn
		}
	case 3: // Geri dön
		if r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeyEscape) {
			r.factionCursor = 0
			return InputAction{Kind: ActionSaveSettings}
		}
	}

	if r.keyJustPressed(ebiten.KeyEscape) {
		r.factionCursor = 0
		return InputAction{Kind: ActionSaveSettings}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		switch r.factionCursor {
		case 0:
			s.Difficulty++
			if s.Difficulty > 3 {
				s.Difficulty = 1
			}
		case 1:
			s.MusicOn = !s.MusicOn
		case 2:
			s.SoundOn = !s.SoundOn
		case 3:
			r.factionCursor = 0
			return InputAction{Kind: ActionSaveSettings}
		}
	}
	return InputAction{}
}

func (r *Renderer) settingsHoverIndex(fx, fy float64) int {
	rowH := 60.0
	rowCount := 4
	startY := ScreenHeight/2 - float64(rowCount)*rowH/2
	bw := 500.0
	bx := ScreenWidth/2 - bw/2
	for i := 0; i < rowCount; i++ {
		y := startY + float64(i)*rowH
		if fx >= bx && fx <= bx+bw && fy >= y-8 && fy <= y+rowH-4 {
			return i
		}
	}
	return -1
}
