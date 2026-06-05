package render

import (
	"image/color"

	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

func buildVictoryCardButtons(gs *state.GameState) []gameui.Button {
	opts := gs.AvailableVictories
	cardW, cardH := 520.0, 100.0
	gap := 12.0
	headerH := 80.0
	stack := centeredStackRect(len(opts), cardW, cardH, gap, headerH)
	buttons := make([]gameui.Button, 0, len(opts))
	for i, opt := range opts {
		r := stackItemRect(stack, cardH, gap, i)
		buttons = append(buttons, gameui.NewButton(r.X, r.Y, r.W, r.H, opt.Title))
	}
	return buttons
}

// DrawVictorySelect zafer koşulu seçim ekranını çizer.
// Seçenekler gs.AvailableVictories'ten okunur — hardcode değil.
func DrawVictorySelect(screen *ebiten.Image, gs *state.GameState, cursor int) {
	opts := gs.AvailableVictories
	drawUIScreenChrome(screen, color.RGBA{10, 10, 20, 255}, "ZAFER KOŞULUNU SEÇ", "Nasıl kazanmak istiyorsun?")

	cardW, cardH := 520.0, 100.0
	gap := 12.0
	headerH := 80.0
	stack := centeredStackRect(len(opts), cardW, cardH, gap, headerH)

	drawBackButton(screen)

	if len(opts) == 0 {
		drawUILabel(screen, gameui.Rect{X: 0, Y: ScreenHeight / 2, W: ScreenWidth}, "Bu senaryo için zafer koşulu tanımlanmamış.", ColorGray, gameui.TextMedium, gameui.TextAlignCenter)
		return
	}

	for i, opt := range opts {
		rect := stackItemRect(stack, cardH, gap, i)
		y := rect.Y

		bg := color.RGBA{25, 25, 45, 220}
		border := color.RGBA{80, 80, 120, 200}
		if i == cursor {
			bg = color.RGBA{50, 45, 90, 240}
			border = color.RGBA{200, 160, 60, 255}
		}

		drawUICardRect(screen, rect, bg, border, 2)

		titleCol := ColorWhite
		if i == cursor {
			titleCol = ColorYellow
		}
		drawUILabel(screen, gameui.Rect{X: rect.X + 18, Y: y + 14}, opt.Title, titleCol, gameui.TextLarge, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: rect.X + 18, Y: y + 38}, opt.Description, ColorGray, gameui.TextMedium, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: rect.X + 18, Y: y + 60}, opt.Detail, color.RGBA{140, 120, 80, 220}, gameui.TextSmall, gameui.TextAlignStart)
	}

	drawUILabel(screen, gameui.Rect{X: 0, Y: stack.Y + stack.H + 20, W: ScreenWidth}, "Zafer koşulunu seçmek için tıkla", ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
}

// handleVictorySelectInput zafer seçim ekranı girişini işler.
func (r *Renderer) handleVictorySelectInput(input gameui.InputState) InputAction {
	opts := r.gs.AvailableVictories
	n := len(opts)
	if n == 0 {
		if r.keyJustPressed(ebiten.KeyEscape) {
			return InputAction{Kind: ActionBack}
		}
		return InputAction{}
	}
	buttons := buildVictoryCardButtons(r.gs)

	for i, btn := range buttons {
		if btn.HitTest(input.MouseX, input.MouseY) {
			r.factionCursor = i
			break
		}
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) {
		r.factionCursor = (r.factionCursor + 1) % n
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) {
		r.factionCursor = (r.factionCursor - 1 + n) % n
	}
	if r.keyJustPressed(ebiten.KeyTab) {
		next := focusButtonIndex(buttons, r.factionCursor, ebiten.IsKeyPressed(ebiten.KeyShift))
		if next >= 0 && next < n {
			r.factionCursor = next
		}
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		return InputAction{Kind: ActionSelectVictory, BuildingID: opts[r.factionCursor].ID}
	}
	if input.LeftJustPressed {
		if buildBackButton().HandleInput(input) {
			r.factionCursor = 0
			return InputAction{Kind: ActionBack}
		}
		for i, btn := range buttons {
			if btn.HandleInput(input) {
				return InputAction{Kind: ActionSelectVictory, BuildingID: opts[i].ID}
			}
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
