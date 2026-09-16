package render

import (
	"image/color"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	shortcutsPanelW = 760.0
	shortcutsPanelH = 500.0
)

type shortcutEntry struct {
	key   string
	label string
}

var shortcutEntries = []shortcutEntry{
	{key: "Enter / Space", label: "Turu bitir"},
	{key: "Esc", label: "Paneli kapatır; seçim yoksa duraklama menüsünü açar"},
	{key: "Tab", label: "Diplomasi paneli"},
	{key: "C", label: "Pazar paneli"},
	{key: "T", label: "Teknoloji paneli"},
	{key: "M", label: "Harita görünümünü değiştirir"},
	{key: "I", label: "İmparatorluk paneli (uygunsa)"},
	{key: "Q / F1", label: "Kısayollar penceresi"},
	{key: "R", label: "Kara birliği alma (seçili bölge)"},
	{key: "N", label: "Gemi inşa et (seçili bölge)"},
	{key: "1 - 9", label: "Bina inşa et (seçili bölge)"},
	{key: "Ctrl + S / F5", label: "Hızlı kaydet"},
	{key: "F9", label: "Kayıttan yükleme menüsü"},
	{key: "+ / -", label: "Vergiyi artırır / azaltır"},
	{key: "F11", label: "Tam ekranı açar / kapatır"},
	{key: "F12", label: "AI teşhis paneli (geliştirme modu)"},
}

func shortcutsPanelRect() gameui.Rect {
	return gameui.Rect{
		X: ScreenWidth/2 - shortcutsPanelW/2,
		Y: ScreenHeight/2 - shortcutsPanelH/2,
		W: shortcutsPanelW,
		H: shortcutsPanelH,
	}
}

func shortcutsCloseButton() gameui.Button {
	panel := shortcutsPanelRect()
	return gameui.NewButton(panel.X+panel.W-132, panel.Y+panel.H-54, 108, 30, "Kapat")
}

func (r *Renderer) OpenShortcuts() {
	if r != nil {
		r.showShortcuts = true
	}
}

func (r *Renderer) handleShortcutsInput() InputAction {
	if r.keyJustPressed(ebiten.KeyQ) || r.keyJustPressed(ebiten.KeyF1) || r.keyJustPressed(ebiten.KeyEscape) {
		r.showShortcuts = false
		return InputAction{}
	}
	mx, my := ebiten.CursorPosition()
	input := gameui.InputState{
		MouseX:          float64(mx),
		MouseY:          float64(my),
		LeftJustPressed: r.mouseJustPressed(ebiten.MouseButtonLeft),
	}
	if input.LeftJustPressed {
		if shortcutsCloseButton().HandleInput(input) || !shortcutsPanelRect().Hit(input.MouseX, input.MouseY) {
			r.showShortcuts = false
		}
	}
	return InputAction{}
}

func (r *Renderer) drawShortcutsOverlay(screen *ebiten.Image) {
	drawUIOverlay(screen, color.RGBA{0, 0, 0, 175})
	panel := shortcutsPanelRect()
	drawUICardRect(screen, panel, color.RGBA{12, 14, 23, 248}, color.RGBA{205, 165, 65, 255}, 2)
	vector.FillRect(screen, float32(panel.X), float32(panel.Y), float32(panel.W), 4, color.RGBA{220, 175, 65, 255}, false)
	drawUILabel(screen, gameui.Rect{X: panel.X, Y: panel.Y + 22, W: panel.W}, "[ KLAVYE KISAYOLLARI ]", ColorYellow, gameui.TextLarge, gameui.TextAlignCenter)

	columnW := (panel.W - 56) / 2
	startY := panel.Y + 78
	rowH := 40.0
	for i, entry := range shortcutEntries {
		column := i / 8
		row := i % 8
		x := panel.X + 20 + float64(column)*(columnW+16)
		y := startY + float64(row)*rowH
		drawUILabel(screen, gameui.Rect{X: x, Y: y, W: 112}, entry.key, ColorGold, gameui.TextSmall, gameui.TextAlignStart)
		label := trimTextToWidth(entry.label, FaceSmall, columnW-116)
		drawUILabel(screen, gameui.Rect{X: x + 116, Y: y, W: columnW - 116}, label, ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
	}

	drawUILabel(screen, gameui.Rect{X: panel.X + 20, Y: panel.Y + panel.H - 49, W: panel.W - 180}, "Q / F1 / ESC: kapat", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	closeButton := shortcutsCloseButton()
	gameui.DrawButton(screen, closeButton, gameui.ButtonStyle{
		BG: color.RGBA{50, 38, 16, 240}, Border: color.RGBA{190, 145, 50, 255},
		Text: ColorGold, BorderWidth: 1, TextVariant: gameui.TextSmall,
	}, sharedTextRenderer{})
}
