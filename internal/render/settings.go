package render

import (
	"encoding/json"
	"image/color"
	"os"

	"mapp-game-go/internal/audio"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

// Settings oyun ayarlarını tutar — renderer aracılığıyla game'e iletilir.
type Settings struct {
	Difficulty  int // 1=Kolay 2=Normal 3=Zor
	Fullscreen  bool
	FastAITurns bool
	MusicOn     bool
	MusicVolume int // 0-100
	SoundOn     bool
	SoundVolume int // 0-100
}

func DefaultSettings() Settings {
	return Settings{Difficulty: 2, Fullscreen: false, FastAITurns: false, MusicOn: true, MusicVolume: 45, SoundOn: true, SoundVolume: 35}
}

var difficultyLabels = []string{"", "Kolay", "Normal", "Zor"}

func difficultyLabelTR(difficulty int) string {
	if difficulty >= 1 && difficulty < len(difficultyLabels) {
		return difficultyLabels[difficulty]
	}
	return difficultyLabels[2]
}

const settingsPath = "saves/settings.json"

const settingsRowCount = 8

func settingsRowsRect(rowCount int) gameui.Rect {
	return centeredStackRect(rowCount, 500, 56, 4, 40)
}

func settingsRowRect(i int) gameui.Rect {
	return stackItemRect(settingsRowsRect(settingsRowCount), 56, 4, i)
}

// LoadSettings ayarları dosyadan yükler, yoksa varsayılanı döner.
func LoadSettings() Settings {
	s := DefaultSettings()
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		json.Unmarshal(data, &s)
	}
	return s
}

// SaveSettings ayarları dosyaya kaydeder.
func SaveSettingsToFile(s Settings) {
	os.MkdirAll("saves", 0755)
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(settingsPath, data, 0644)
}

// DrawSettingsScreen ayarlar ekranını çizer.
func DrawSettingsScreen(screen *ebiten.Image, s Settings, cursor int) {
	drawUIScreenChrome(screen, color.RGBA{8, 10, 18, 255}, "[ AYARLAR ]", "Fareyle seç / değiştir")

	type row struct {
		label string
		value string
	}
	rows := []row{
		{"Zorluk", difficultyLabelTR(s.Difficulty)},
		{"Ekran Modu", displayModeLabelTR(s.Fullscreen)},
		{"Hızlı AI Hamleleri", boolLabel(s.FastAITurns)},
		{"Müzik", boolLabel(s.MusicOn)},
		{"Müzik Seviyesi", itoa(s.MusicVolume) + "%"},
		{"Ses Efektleri", boolLabel(s.SoundOn)},
		{"Ses Seviyesi", itoa(s.SoundVolume) + "%"},
		{"← Geri Dön", ""},
	}

	for i, r := range rows {
		rect := settingsRowRect(i)
		y := rect.Y
		isSelected := i == cursor

		if isSelected {
			highlightRect := gameui.Rect{X: rect.X, Y: y - 8, W: rect.W, H: 56}
			drawUICardRect(screen, highlightRect, color.RGBA{50, 40, 15, 200}, color.RGBA{200, 160, 60, 200}, 1)
		}

		col := ColorGray
		if isSelected {
			col = ColorYellow
		}
		drawUILabel(screen, gameui.Rect{X: rect.X + 30, Y: y + 6}, r.label, col, gameui.TextLarge, gameui.TextAlignStart)
		if r.value != "" {
			drawUILabel(screen, gameui.Rect{X: rect.X + 310, Y: y + 6}, "◄  "+r.value+"  ►", ColorGold, gameui.TextLarge, gameui.TextAlignStart)
		}
	}
	drawUILabel(screen, gameui.Rect{X: 0, Y: ScreenHeight - 30, W: ScreenWidth}, "Sol tık: değiştir  •  ESC: kaydet ve çık", ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
}

func boolLabel(b bool) string {
	if b {
		return "Açık"
	}
	return "Kapalı"
}

func displayModeLabelTR(fullscreen bool) string {
	if fullscreen {
		return "Tam Ekran"
	}
	return "Pencereli"
}

// ApplyDisplaySettings oyun penceresinin görünümünü ayarlara uygular.
func ApplyDisplaySettings(s Settings) {
	ebiten.SetFullscreen(s.Fullscreen)
}

func (r *Renderer) toggleFullscreen() {
	r.CurrentSettings.Fullscreen = !r.CurrentSettings.Fullscreen
	ApplyDisplaySettings(r.CurrentSettings)
	SaveSettingsToFile(r.CurrentSettings)
}

// handleSettingsInput ayarlar ekranı girişini işler.
func (r *Renderer) handleSettingsInput(s *Settings) InputAction {
	rowCount := settingsRowCount // zorluk, ekran modu, AI hamleleri, müzik, müzik seviyesi, ses, ses seviyesi, geri dön
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
	case 1: // Ekran modu
		if r.keyJustPressed(ebiten.KeyArrowLeft) || r.keyJustPressed(ebiten.KeyArrowRight) || r.keyJustPressed(ebiten.KeyEnter) {
			s.Fullscreen = !s.Fullscreen
			ApplyDisplaySettings(*s)
		}
	case 2: // AI hamleleri
		if r.keyJustPressed(ebiten.KeyArrowLeft) || r.keyJustPressed(ebiten.KeyArrowRight) || r.keyJustPressed(ebiten.KeyEnter) {
			s.FastAITurns = !s.FastAITurns
		}
	case 3: // Müzik
		if r.keyJustPressed(ebiten.KeyArrowLeft) || r.keyJustPressed(ebiten.KeyArrowRight) || r.keyJustPressed(ebiten.KeyEnter) {
			s.MusicOn = !s.MusicOn
			applyAudioSettings(*s)
		}
	case 4: // Müzik seviyesi
		if r.keyJustPressed(ebiten.KeyArrowRight) {
			s.MusicVolume = clampVolume(s.MusicVolume + 5)
			applyAudioSettings(*s)
		}
		if r.keyJustPressed(ebiten.KeyArrowLeft) {
			s.MusicVolume = clampVolume(s.MusicVolume - 5)
			applyAudioSettings(*s)
		}
	case 5: // Ses efektleri
		if r.keyJustPressed(ebiten.KeyArrowLeft) || r.keyJustPressed(ebiten.KeyArrowRight) || r.keyJustPressed(ebiten.KeyEnter) {
			s.SoundOn = !s.SoundOn
			applyAudioSettings(*s)
		}
	case 6: // Ses seviyesi
		if r.keyJustPressed(ebiten.KeyArrowRight) {
			s.SoundVolume = clampVolume(s.SoundVolume + 5)
			applyAudioSettings(*s)
		}
		if r.keyJustPressed(ebiten.KeyArrowLeft) {
			s.SoundVolume = clampVolume(s.SoundVolume - 5)
			applyAudioSettings(*s)
		}
	case 7: // Geri dön
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
		hover := r.settingsHoverIndex(float64(mx), float64(my))
		if hover < 0 {
			return InputAction{}
		}
		r.factionCursor = hover
		switch hover {
		case 0:
			s.Difficulty++
			if s.Difficulty > 3 {
				s.Difficulty = 1
			}
		case 1:
			s.Fullscreen = !s.Fullscreen
			ApplyDisplaySettings(*s)
		case 2:
			s.FastAITurns = !s.FastAITurns
		case 3:
			s.MusicOn = !s.MusicOn
			applyAudioSettings(*s)
		case 4:
			s.MusicVolume += 10
			if s.MusicVolume > 100 {
				s.MusicVolume = 0
			}
			applyAudioSettings(*s)
		case 5:
			s.SoundOn = !s.SoundOn
			applyAudioSettings(*s)
		case 6:
			s.SoundVolume += 10
			if s.SoundVolume > 100 {
				s.SoundVolume = 0
			}
			applyAudioSettings(*s)
		case 7:
			r.factionCursor = 0
			return InputAction{Kind: ActionSaveSettings}
		}
	}
	return InputAction{}
}

func (r *Renderer) settingsHoverIndex(fx, fy float64) int {
	rowCount := settingsRowCount
	for i := 0; i < rowCount; i++ {
		rect := settingsRowRect(i)
		y := rect.Y
		if fx >= rect.X && fx <= rect.X+rect.W && fy >= y-8 && fy <= y+56 {
			return i
		}
	}
	return -1
}

func clampVolume(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func applyAudioSettings(s Settings) {
	audio.SetMusicEnabled(s.MusicOn)
	audio.SetMusicVolume(s.MusicVolume)
	audio.SetSoundEnabled(s.SoundOn)
	audio.SetSoundVolume(s.SoundVolume)
}
