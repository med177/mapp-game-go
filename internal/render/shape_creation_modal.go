package render

import (
	"image/color"
	"strings"
	"unicode"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

type editNewShapeModalState struct {
	show bool
}

const (
	editNewShapeModalW = 460.0
	editNewShapeModalH = 218.0
)

func editNewShapeModalRect() gameui.Rect {
	return gameui.AnchorRect(
		gameui.Rect{W: ScreenWidth, H: ScreenHeight},
		editNewShapeModalW, editNewShapeModalH,
		gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0,
	)
}

func editNewShapeModalInputRect() gameui.Rect {
	rect := editNewShapeModalRect()
	return gameui.Rect{X: rect.X + 24, Y: rect.Y + 92, W: rect.W - 48, H: 34}
}

func editNewShapeModalOKButton() gameui.Button {
	rect := editNewShapeModalRect()
	return gameui.NewButton(rect.X+rect.W-174, rect.Y+rect.H-48, 72, 30, "OK")
}

func editNewShapeModalCancelButton() gameui.Button {
	rect := editNewShapeModalRect()
	return gameui.NewButton(rect.X+rect.W-92, rect.Y+rect.H-48, 68, 30, "İptal")
}

func editNewShapeModalHit(mx, my float64) bool {
	return editNewShapeModalRect().Hit(mx, my)
}

func editNewShapeModalInputHit(mx, my float64) bool {
	return editNewShapeModalInputRect().Hit(mx, my)
}

func (r *Renderer) drawEditNewShapeModal(screen *ebiten.Image) {
	if r == nil || !r.editNewShapeModal.show {
		return
	}
	rect := editNewShapeModalRect()
	gameui.DrawModal(screen,
		gameui.NewModal(ScreenWidth, ScreenHeight, gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)),
		standardModalStyle, renderText,
		func() {
			stageLabel := "Shape ID"
			help := "Boşluk kullanılamaz. Enter ile ilerle."
			if r.editTextTarget == editTextShapeName {
				stageLabel = "Shape adı"
				help = "Yeni kara bölgesinin görünen adı."
			}
			DrawText(screen, "YENİ KARA SINIRI", rect.X+24, rect.Y+18, FaceLarge, ColorGold)
			DrawText(screen, stageLabel, rect.X+24, rect.Y+64, FaceSmall, ColorGray)
			box := gameui.NewTextBox(editNewShapeModalInputRect().X, editNewShapeModalInputRect().Y, editNewShapeModalInputRect().W, editNewShapeModalInputRect().H, stageLabel+" girin")
			box.Value = string(r.editTextRunes)
			box.Focused = true
			box.MaxLen = 64
			gameui.DrawTextBox(screen, box, editNewShapeModalTextBoxStyle(), renderText)
			DrawText(screen, help, rect.X+24, rect.Y+139, FaceSmall, ColorGray)
			if r.editTextError != "" {
				DrawText(screen, r.editTextError, rect.X+24, rect.Y+163, FaceSmall, ColorRed)
			}
			drawUIButtonWidget(screen, editNewShapeModalOKButton(), tinyButtonStyle)
			drawUIButtonWidget(screen, editNewShapeModalCancelButton(), tinyButtonStyle)
		},
	)
}

func editNewShapeModalTextBoxStyle() gameui.TextBoxStyle {
	return gameui.TextBoxStyle{
		BG:          color.RGBA{28, 32, 38, 245},
		Border:      color.RGBA{120, 105, 60, 210},
		Focused:     ColorGold,
		Text:        ColorWhite,
		Placeholder: ColorGray,
		BorderWidth: 1,
		TextOffsetX: 8,
		TextOffsetY: 8,
		TextVariant: gameui.TextSmall,
	}
}

func (r *Renderer) handleEditNewShapeModalInput() InputAction {
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.closeEditNewShapeModal()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyBackspace) && len(r.editTextRunes) > 0 {
		r.editTextRunes = r.editTextRunes[:len(r.editTextRunes)-1]
	}
	if (r.editTextTarget == editTextShapeID || r.editTextTarget == editTextShapeName) &&
		r.keyJustPressed(ebiten.KeyA) && editCtrlPressed() {
		r.editTextRunes = r.editTextRunes[:0]
		return InputAction{}
	}
	chars := ebiten.AppendInputChars(nil)
	if len(chars) > 0 {
		r.editTextRunes = append(r.editTextRunes, chars...)
		if len(r.editTextRunes) > 64 {
			r.editTextRunes = r.editTextRunes[:64]
		}
	}
	if r.keyJustPressed(ebiten.KeyEnter) || (r.mouseJustPressed(ebiten.MouseButtonLeft) && editNewShapeModalOKButton().HitTest(cursorX(), cursorY())) {
		r.advanceEditNewShapeModal()
		return InputAction{}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) && editNewShapeModalCancelButton().HitTest(cursorX(), cursorY()) {
		r.closeEditNewShapeModal()
	}
	return InputAction{}
}

func (r *Renderer) advanceEditNewShapeModal() {
	if r.editTextTarget == editTextShapeID {
		value := normalizeEditID(string(r.editTextRunes))
		if value == "" {
			r.editTextError = "Shape ID boş olamaz."
			return
		}
		if strings.IndexFunc(value, unicode.IsSpace) >= 0 {
			r.editTextError = "Shape ID boşluk içeremez."
			return
		}
		if _, exists := r.gs.ShapeData.Shapes[value]; exists {
			r.editTextError = "Bu Shape ID zaten var."
			return
		}
		r.editNewShapeID = value
		r.editTextTarget = editTextShapeName
		r.editTextRunes = r.editTextRunes[:0]
		r.editTextError = ""
		return
	}
	r.commitNewShapeInput()
}

func (r *Renderer) closeEditNewShapeModal() {
	r.editNewShapeModal = editNewShapeModalState{}
	r.editTextTarget = editTextNone
	r.editTextRunes = r.editTextRunes[:0]
	r.editTextError = ""
	r.editNewShapeID = ""
	r.editNewShapeRegion = ""
}

func cursorX() float64 {
	x, _ := ebiten.CursorPosition()
	return float64(x)
}

func cursorY() float64 {
	_, y := ebiten.CursorPosition()
	return float64(y)
}
