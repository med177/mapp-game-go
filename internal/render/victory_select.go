package render

import (
	"image/color"

	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawVictorySelect zafer koşulu seçim ekranını çizer.
// Seçenekler gs.AvailableVictories'ten okunur — hardcode değil.
func DrawVictorySelect(screen *ebiten.Image, gs *state.GameState, cursor int) {
	opts := gs.AvailableVictories
	screen.Fill(color.RGBA{10, 10, 20, 255})

	cardW, cardH := 520.0, 100.0
	gap := 12.0
	n := float64(len(opts))
	totalH := n*cardH + (n-1)*gap
	headerH := 80.0

	startY := (ScreenHeight-(totalH+headerH))/2 + headerH
	cx := ScreenWidth/2 - cardW/2

	DrawTextCentered(screen, "ZAFER KOŞULUNU SEÇ", ScreenWidth/2, startY-headerH+10, FaceLarge, ColorYellow)
	DrawTextCentered(screen, "Nasıl kazanmak istiyorsun?", ScreenWidth/2, startY-headerH+38, FaceSmall, ColorGray)

	if len(opts) == 0 {
		DrawTextCentered(screen, "Bu senaryo için zafer koşulu tanımlanmamış.", ScreenWidth/2, ScreenHeight/2, FaceMed, ColorGray)
		return
	}

	for i, opt := range opts {
		y := startY + float64(i)*(cardH+gap)

		bg := color.RGBA{25, 25, 45, 220}
		border := color.RGBA{80, 80, 120, 200}
		if i == cursor {
			bg = color.RGBA{50, 45, 90, 240}
			border = color.RGBA{200, 160, 60, 255}
		}

		vector.FillRect(screen, float32(cx), float32(y), float32(cardW), float32(cardH), bg, false)
		vector.StrokeRect(screen, float32(cx), float32(y), float32(cardW), float32(cardH), 2, border, false)

		titleCol := ColorWhite
		if i == cursor {
			titleCol = ColorYellow
		}
		DrawText(screen, opt.Title, cx+18, y+14, FaceLarge, titleCol)
		DrawText(screen, opt.Description, cx+18, y+38, FaceMed, ColorGray)
		DrawText(screen, opt.Detail, cx+18, y+60, FaceSmall, color.RGBA{140, 120, 80, 220})

		if i == cursor {
			DrawText(screen, "← SEÇİLİ", cx+cardW-110, y+14, FaceSmall, ColorGold)
		}
	}

	DrawTextCentered(screen, "[↑↓] Seç   [Enter] Onayla   [Esc] Geri", ScreenWidth/2, startY+totalH+20, FaceSmall, ColorGray)
}

// handleVictorySelectInput zafer seçim ekranı girişini işler.
func (r *Renderer) handleVictorySelectInput() InputAction {
	opts := r.gs.AvailableVictories
	n := len(opts)
	if n == 0 {
		if r.keyJustPressed(ebiten.KeyEscape) {
			return InputAction{Kind: ActionBack}
		}
		return InputAction{}
	}

	mx, my := ebiten.CursorPosition()
	if i := r.victoryCardHoverIndex(float64(mx), float64(my)); i >= 0 {
		r.factionCursor = i
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) {
		r.factionCursor = (r.factionCursor + 1) % n
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) {
		r.factionCursor = (r.factionCursor - 1 + n) % n
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		return InputAction{Kind: ActionSelectVictory, BuildingID: opts[r.factionCursor].ID}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if i := r.victoryCardHoverIndex(float64(mx), float64(my)); i >= 0 {
			return InputAction{Kind: ActionSelectVictory, BuildingID: opts[i].ID}
		}
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.factionCursor = 0
		return InputAction{Kind: ActionBack}
	}
	return InputAction{}
}

// VictoryOptionByID gs.AvailableVictories içinden ID'ye göre seçenek bulur.
func VictoryOptionByID(gs *state.GameState, id string) (scenario.VictoryOptionDef, bool) {
	for _, v := range gs.AvailableVictories {
		if v.ID == id {
			return v, true
		}
	}
	return scenario.VictoryOptionDef{}, false
}
