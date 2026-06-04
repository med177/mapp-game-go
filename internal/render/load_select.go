package render

import (
	"image/color"

	"mapp-game-go/internal/save"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

// SaveSlots kayıt seçim ekranında gösterilecek slot listesidir.
// game.go tarafından ekrana girilirken doldurulur.
var SaveSlots []save.SaveSlot

type slotCardLayout struct {
	X float64
	Y float64
	W float64
	H float64
}

func slotCardsStackRect() gameui.Rect {
	return centeredStackRect(len(SaveSlots), 480, 88, 14, 0)
}

func slotCardLayoutAt(i int) slotCardLayout {
	stack := slotCardsStackRect()
	rect := stackItemRect(stack, 88, 14, i)
	return slotCardLayout{
		X: rect.X,
		Y: rect.Y,
		W: rect.W,
		H: rect.H,
	}
}

func buildSlotBackButton() gameui.Button {
	r := backButtonRect()
	return gameui.NewButton(r[0], r[1], r[2], r[3], "Geri")
}

func buildSlotCardButton(i int) gameui.Button {
	card := slotCardLayoutAt(i)
	return gameui.NewButton(card.X, card.Y, card.W, card.H, SaveSlots[i].DisplayName)
}

func buildSlotDeleteButton(i int) gameui.Button {
	rect := deleteButtonRectForSlot(i)
	return gameui.NewButton(rect[0], rect[1], rect[2], rect[3], "Sil")
}

func buildSlotFocusButtons(saveMode bool) []gameui.Button {
	buttons := make([]gameui.Button, 0, len(SaveSlots))
	for i, slot := range SaveSlots {
		btn := buildSlotCardButton(i)
		btn.Enabled = saveMode || slot.Exists
		buttons = append(buttons, btn)
	}
	return buttons
}

func buildSlotConfirmButtons(pendingSlot string) (gameui.Button, gameui.Button, bool) {
	yes, no := pendingDeleteConfirmRects(pendingSlot)
	if yes == (slotRect{}) || no == (slotRect{}) {
		return gameui.Button{}, gameui.Button{}, false
	}
	return gameui.NewButton(yes[0], yes[1], yes[2], yes[3], "Sil"),
		gameui.NewButton(no[0], no[1], no[2], no[3], "İptal"), true
}

// DrawSlotSelectScreen yükleme veya kaydetme için slot seçim ekranını çizer.
// saveMode=true ise kaydetme, false ise yükleme ekranı başlığı gösterilir.
// pendingDelete dolu ise o slot için onay diyalogu gösterilir.
func DrawSlotSelectScreen(screen *ebiten.Image, cursor int, saveMode bool, pendingDelete string) {
	title := "KAYIT YÜKLE"
	if saveMode {
		title = "KAYIT ET"
	}
	drawUIScreenChrome(screen, color.RGBA{6, 8, 14, 255}, title, "Bir slot seçin")
	drawBackButton(screen)

	if len(SaveSlots) == 0 {
		DrawTextCentered(screen, "Kayıt bulunamadı.", ScreenWidth/2, ScreenHeight/2, FaceMed, ColorGray)
		DrawTextCentered(screen, "Geri düğmesiyle ana menüye dön", ScreenWidth/2, ScreenHeight-40, FaceSmall, ColorGray)
		return
	}

	cardW := float64(480)
	cardH := float64(88)

	for i, slot := range SaveSlots {
		card := slotCardLayoutAt(i)
		cx := card.X
		cy := card.Y

		isSelected := i == cursor
		disabled := !saveMode && !slot.Exists
		isPendingDelete := slot.Name == pendingDelete

		// Kart arka planı
		bg := color.RGBA{14, 12, 8, 220}
		borderCol := color.RGBA{55, 45, 28, 200}
		if isPendingDelete {
			bg = color.RGBA{50, 10, 10, 240}
			borderCol = color.RGBA{200, 50, 50, 255}
		} else if isSelected && !disabled {
			bg = color.RGBA{45, 36, 14, 240}
			borderCol = color.RGBA{200, 160, 50, 240}
		} else if disabled {
			bg = color.RGBA{10, 10, 10, 160}
			borderCol = color.RGBA{35, 30, 20, 150}
		}
		drawUICardRect(screen, gameui.Rect{X: cx, Y: cy, W: cardW, H: cardH}, bg, borderCol, 1.5)

		// Sol: slot adı
		nameCol := ColorGold
		if isPendingDelete {
			nameCol = ColorRed
		} else if disabled {
			nameCol = color.RGBA{50, 50, 50, 180}
		} else if isSelected {
			nameCol = color.RGBA{255, 220, 80, 255}
		}
		prefix := "  "
		if isSelected && !disabled && !isPendingDelete {
			prefix = "► "
		}
		DrawText(screen, prefix+slot.DisplayName, cx+18, cy+14, FaceLarge, nameCol)

		if slot.Exists {
			detailCol := color.RGBA{180, 165, 120, 200}
			if disabled {
				detailCol = color.RGBA{50, 50, 50, 160}
			}
			if isPendingDelete {
				// Onay sorusu kartın içine yerleşir
				DrawTextCentered(screen, "Silinecek! Emin misiniz?", cx+cardW/2, cy+40,
					FaceMed, color.RGBA{255, 100, 100, 255})
				yes, no := slotDeleteConfirmRects(cx, cy)
				drawSlotMiniButton(screen, yes, "Sil", color.RGBA{130, 35, 35, 230})
				drawSlotMiniButton(screen, no, "İptal", color.RGBA{45, 45, 45, 230})
			} else {
				faction := slot.FactionName
				if faction == "" {
					faction = "Bilinmiyor"
				}
				DrawText(screen, "Fraksiyon: "+faction, cx+18, cy+44, FaceSmall, detailCol)
				DrawText(screen, "Tur: "+itoa(slot.Turn)+"  |  "+itoa(slot.Year),
					cx+cardW/2, cy+44, FaceSmall, detailCol)

				modStr := slot.ModTime.Format("02.01.2006 15:04")
				tw := MeasureText(modStr, FaceSmall)
				DrawText(screen, modStr, cx+cardW-tw-18, cy+14, FaceSmall,
					color.RGBA{110, 100, 70, 200})

				// Sil butonu göstergesi (sadece dolu ve seçili slotta)
				if isSelected {
					del := slotDeleteButtonRect(cx, cy)
					drawSlotMiniButton(screen, del, "Sil", color.RGBA{95, 35, 35, 220})
				}
			}
		} else {
			emptyCol := color.RGBA{45, 40, 30, 180}
			if saveMode && isSelected {
				emptyCol = color.RGBA{140, 160, 80, 220}
			}
			DrawTextCentered(screen, "- Bos Slot -", cx+cardW/2, cy+cardH/2-8, FaceMed, emptyCol)
		}
	}

	hint := "Slotu seçmek veya silmek için tıkla"
	if pendingDelete != "" {
		hint = "Sil veya İptal düğmesine tıkla"
	}
	DrawTextCentered(screen, hint, ScreenWidth/2, ScreenHeight-30, FaceSmall, color.RGBA{80, 80, 80, 200})
}

// handleSlotSelectInput slot seçim ekranının girişini işler.
func (r *Renderer) handleSlotSelectInput(saveMode bool, input gameui.InputState) InputAction {
	n := len(SaveSlots)
	if n == 0 {
		if buildSlotBackButton().HandleInput(input) {
			return InputAction{Kind: ActionBack}
		}
		if r.keyJustPressed(ebiten.KeyEscape) {
			return InputAction{Kind: ActionBack}
		}
		return InputAction{}
	}

	// Onay bekleniyor: sadece Enter (onayla) ve Esc (iptal) çalışır
	if r.pendingDeleteSlot != "" {
		if i := r.slotHoverIndex(input.MouseX, input.MouseY); i >= 0 {
			r.slotCursor = i
		}
		if yesBtn, noBtn, ok := buildSlotConfirmButtons(r.pendingDeleteSlot); ok {
			if yesBtn.HandleInput(input) {
				slot := r.pendingDeleteSlot
				r.pendingDeleteSlot = ""
				return InputAction{Kind: ActionDeleteSave, BuildingID: slot}
			}
			if noBtn.HandleInput(input) {
				r.pendingDeleteSlot = ""
				return InputAction{}
			}
		}
		if r.keyJustPressed(ebiten.KeyEnter) {
			slot := r.pendingDeleteSlot
			r.pendingDeleteSlot = ""
			return InputAction{Kind: ActionDeleteSave, BuildingID: slot}
		}
		if r.keyJustPressed(ebiten.KeyEscape) {
			r.pendingDeleteSlot = ""
		}
		return InputAction{}
	}

	if i := r.slotHoverIndex(input.MouseX, input.MouseY); i >= 0 {
		r.slotCursor = i
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) {
		r.slotCursor = (r.slotCursor + 1) % n
		if !saveMode {
			for !SaveSlots[r.slotCursor].Exists {
				r.slotCursor = (r.slotCursor + 1) % n
			}
		}
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) {
		r.slotCursor = (r.slotCursor - 1 + n) % n
		if !saveMode {
			for !SaveSlots[r.slotCursor].Exists {
				r.slotCursor = (r.slotCursor - 1 + n) % n
			}
		}
	}
	if r.keyJustPressed(ebiten.KeyTab) {
		next := focusButtonIndex(buildSlotFocusButtons(saveMode), r.slotCursor, ebiten.IsKeyPressed(ebiten.KeyShift))
		if next >= 0 && next < n {
			r.slotCursor = next
		}
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		return InputAction{Kind: ActionBack}
	}

	// Del veya Backspace: dolu slotu silme onayına al
	if r.keyJustPressed(ebiten.KeyDelete) || r.keyJustPressed(ebiten.KeyBackspace) {
		if r.slotCursor < len(SaveSlots) && SaveSlots[r.slotCursor].Exists {
			r.pendingDeleteSlot = SaveSlots[r.slotCursor].Name
		}
		return InputAction{}
	}

	if r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace) {
		slot := SaveSlots[r.slotCursor]
		if saveMode || slot.Exists {
			return InputAction{Kind: ActionSelectSave, BuildingID: slot.Name}
		}
	}
	if input.LeftJustPressed {
		if buildSlotBackButton().HandleInput(input) {
			return InputAction{Kind: ActionBack}
		}
		for i := range SaveSlots {
			cardBtn := buildSlotCardButton(i)
			slot := SaveSlots[i]
			if slot.Exists && buildSlotDeleteButton(i).HandleInput(input) {
				r.pendingDeleteSlot = slot.Name
				return InputAction{}
			}
			if cardBtn.HandleInput(input) && (saveMode || slot.Exists) {
				return InputAction{Kind: ActionSelectSave, BuildingID: slot.Name}
			}
		}
	}
	return InputAction{}
}

type slotRect [4]float64

func drawSlotMiniButton(screen *ebiten.Image, r slotRect, label string, bg color.RGBA) {
	style := slotMiniButtonStyle
	style.BG = bg
	drawUIButton(screen, r[0], r[1], r[2], r[3], label, true, style)
}

func slotDeleteButtonRect(cx, cy float64) slotRect {
	return slotRect{cx + 396, cy + 50, 58, 24}
}

func slotDeleteConfirmRects(cx, cy float64) (slotRect, slotRect) {
	return slotRect{cx + 166, cy + 54, 70, 24}, slotRect{cx + 244, cy + 54, 70, 24}
}

func deleteButtonRectForSlot(i int) slotRect {
	card := slotCardLayoutAt(i)
	cx := card.X
	cy := card.Y
	return slotDeleteButtonRect(cx, cy)
}

func pendingDeleteConfirmRects(pendingSlot string) (slotRect, slotRect) {
	for i, slot := range SaveSlots {
		if slot.Name != pendingSlot {
			continue
		}
		card := slotCardLayoutAt(i)
		cx := card.X
		cy := card.Y
		return slotDeleteConfirmRects(cx, cy)
	}
	return slotRect{}, slotRect{}
}

func (r *Renderer) slotSelectHovering(mx, my float64, saveMode bool) bool {
	if r.pendingDeleteSlot != "" {
		yesBtn, noBtn, ok := buildSlotConfirmButtons(r.pendingDeleteSlot)
		return ok && (yesBtn.HitTest(mx, my) || noBtn.HitTest(mx, my))
	}
	if buildSlotBackButton().HitTest(mx, my) {
		return true
	}
	i := r.slotHoverIndex(mx, my)
	if i < 0 || i >= len(SaveSlots) {
		return false
	}
	slot := SaveSlots[i]
	if slot.Exists && buildSlotDeleteButton(i).HitTest(mx, my) {
		return true
	}
	return buildSlotCardButton(i).HitTest(mx, my) && (saveMode || slot.Exists)
}

// slotHoverIndex fareye göre hangi slot kartının üzerinde olduğunu döner.
func (r *Renderer) slotHoverIndex(mx, my float64) int {
	for i := range SaveSlots {
		if buildSlotCardButton(i).HitTest(mx, my) {
			return i
		}
	}
	return -1
}
