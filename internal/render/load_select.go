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

const (
	slotCardW                  = 480.0
	slotCardH                  = 88.0
	slotPendingDeleteNameInset = 18.0
	slotPendingDeleteNameY     = 12.0
	slotPendingDeletePromptY   = 36.0
)

func slotCardsStackRect() gameui.Rect {
	return centeredStackRect(len(SaveSlots), slotCardW, slotCardH, 14, 0)
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
		drawUILabel(screen, gameui.Rect{X: 0, Y: ScreenHeight / 2, W: ScreenWidth}, "Kayıt bulunamadı.", ColorGray, gameui.TextMedium, gameui.TextAlignCenter)
		drawUILabel(screen, gameui.Rect{X: 0, Y: ScreenHeight - 40, W: ScreenWidth}, "Geri düğmesiyle ana menüye dön", ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
		return
	}

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
		drawUICardRect(screen, gameui.Rect{X: cx, Y: cy, W: slotCardW, H: slotCardH}, bg, borderCol, 1.5)

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

		if slot.Exists {
			detailCol := color.RGBA{180, 165, 120, 200}
			if disabled {
				detailCol = color.RGBA{50, 50, 50, 160}
			}
			if isPendingDelete {
				// Silme onayında başlık ve soru ayrı bantlarda tutulur ki üst üste binmesin.
				drawUILabel(screen, gameui.Rect{X: cx + slotPendingDeleteNameInset, Y: cy + slotPendingDeleteNameY}, slot.DisplayName, nameCol, gameui.TextMedium, gameui.TextAlignStart)
				drawUILabel(screen, gameui.Rect{X: cx, Y: cy + slotPendingDeletePromptY, W: slotCardW}, "Silinecek! Emin misiniz?", color.RGBA{255, 100, 100, 255}, gameui.TextMedium, gameui.TextAlignCenter)
				yes, no := slotDeleteConfirmRects(cx, cy)
				drawSlotMiniButton(screen, yes, "Sil", color.RGBA{130, 35, 35, 230})
				drawSlotMiniButton(screen, no, "İptal", color.RGBA{45, 45, 45, 230})
			} else {
				drawUILabel(screen, gameui.Rect{X: cx + 18, Y: cy + 14}, prefix+slot.DisplayName, nameCol, gameui.TextLarge, gameui.TextAlignStart)
				faction := slot.FactionName
				if faction == "" {
					faction = "Bilinmiyor"
				}
				drawUILabel(screen, gameui.Rect{X: cx + 18, Y: cy + 44}, "Fraksiyon: "+faction, detailCol, gameui.TextSmall, gameui.TextAlignStart)
				drawUILabel(screen, gameui.Rect{X: cx + slotCardW/2, Y: cy + 44}, "Tur: "+itoa(slot.Turn)+"  |  "+itoa(slot.Year), detailCol, gameui.TextSmall, gameui.TextAlignCenter)

				modStr := slot.ModTime.Format("02.01.2006 15:04")
				drawUILabel(screen, gameui.Rect{X: cx + 18, Y: cy + 14, W: slotCardW - 36}, modStr, color.RGBA{110, 100, 70, 200}, gameui.TextSmall, gameui.TextAlignEnd)

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
			drawUILabel(screen, gameui.Rect{X: cx + 18, Y: cy + 14}, prefix+slot.DisplayName, nameCol, gameui.TextLarge, gameui.TextAlignStart)
			drawUILabel(screen, gameui.Rect{X: cx, Y: cy + slotCardH/2 - 8, W: slotCardW}, "- Bos Slot -", emptyCol, gameui.TextMedium, gameui.TextAlignCenter)
		}
	}

	hint := "Slotu seçmek veya silmek için tıkla"
	if pendingDelete != "" {
		hint = "Sil veya İptal düğmesine tıkla"
	}
	drawUILabel(screen, gameui.Rect{X: 0, Y: ScreenHeight - 30, W: ScreenWidth}, hint, color.RGBA{80, 80, 80, 200}, gameui.TextSmall, gameui.TextAlignCenter)
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

func slotPendingDeleteTextBounds(cy float64) (nameTop, nameBottom, promptTop, promptBottom float64) {
	nameTop = cy + slotPendingDeleteNameY
	nameBottom = nameTop + FaceMed.Size
	promptTop = cy + slotPendingDeletePromptY
	promptBottom = promptTop + FaceMed.Size
	return nameTop, nameBottom, promptTop, promptBottom
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
