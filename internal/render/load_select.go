package render

import (
	"image/color"

	"mapp-game-go/internal/save"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// SaveSlots kayıt seçim ekranında gösterilecek slot listesidir.
// game.go tarafından ekrana girilirken doldurulur.
var SaveSlots []save.SaveSlot

// DrawSlotSelectScreen yükleme veya kaydetme için slot seçim ekranını çizer.
// saveMode=true ise kaydetme, false ise yükleme ekranı başlığı gösterilir.
// pendingDelete dolu ise o slot için onay diyalogu gösterilir.
func DrawSlotSelectScreen(screen *ebiten.Image, cursor int, saveMode bool, pendingDelete string) {
	screen.Fill(color.RGBA{6, 8, 14, 255})

	// Üst/alt dekoratif çizgi
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)
	vector.FillRect(screen, 0, float32(ScreenHeight)-3, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)

	title := "KAYIT YÜKLE"
	if saveMode {
		title = "KAYIT YER"
	}
	DrawTextCentered(screen, title, ScreenWidth/2, 50, FaceLarge, ColorYellow)
	DrawTextCentered(screen, "Bir slot seçin", ScreenWidth/2, 84, FaceSmall, color.RGBA{160, 140, 90, 200})

	if len(SaveSlots) == 0 {
		DrawTextCentered(screen, "Kayıt bulunamadı.", ScreenWidth/2, ScreenHeight/2, FaceMed, ColorGray)
		DrawTextCentered(screen, "[Esc] Geri", ScreenWidth/2, ScreenHeight-40, FaceSmall, ColorGray)
		return
	}

	cardW := float64(480)
	cardH := float64(88)
	gap := float64(14)
	totalH := float64(len(SaveSlots))*cardH + float64(len(SaveSlots)-1)*gap
	startY := ScreenHeight/2 - totalH/2

	for i, slot := range SaveSlots {
		cx := ScreenWidth/2 - cardW/2
		cy := startY + float64(i)*(cardH+gap)

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
		vector.FillRect(screen, float32(cx), float32(cy), float32(cardW), float32(cardH), bg, false)
		vector.StrokeRect(screen, float32(cx), float32(cy), float32(cardW), float32(cardH), 1.5, borderCol, false)

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
				DrawTextCentered(screen, "[Enter] Evet   [Esc] Hayır", cx+cardW/2, cy+62,
					FaceSmall, color.RGBA{200, 160, 160, 220})
			} else {
				faction := slot.FactionName
				if faction == "" {
					faction = "Bilinmiyor"
				}
				DrawText(screen, "Fraksiyon: "+faction, cx+18, cy+44, FaceSmall, detailCol)
				DrawText(screen, "Tur: "+itoa(slot.Turn)+"  │  "+itoa(slot.Year),
					cx+cardW/2, cy+44, FaceSmall, detailCol)

				modStr := slot.ModTime.Format("02.01.2006 15:04")
				tw := MeasureText(modStr, FaceSmall)
				DrawText(screen, modStr, cx+cardW-tw-18, cy+14, FaceSmall,
					color.RGBA{110, 100, 70, 200})

				// Sil butonu göstergesi (sadece dolu ve seçili slotta)
				if isSelected {
					delW := MeasureText("[Del] Sil", FaceSmall)
					DrawText(screen, "[Del] Sil", cx+cardW-delW-18, cy+44, FaceSmall,
						color.RGBA{180, 70, 70, 220})
				}
			}
		} else {
			emptyCol := color.RGBA{45, 40, 30, 180}
			if saveMode && isSelected {
				emptyCol = color.RGBA{140, 160, 80, 220}
			}
			DrawTextCentered(screen, "— Boş Slot —", cx+cardW/2, cy+cardH/2-8, FaceMed, emptyCol)
		}
	}

	hint := "[↑↓] Seç   [Enter] Onayla   [Del] Sil   [Esc] Geri"
	if pendingDelete != "" {
		hint = "[Enter] Silmeyi Onayla   [Esc] İptal"
	}
	DrawTextCentered(screen, hint, ScreenWidth/2, ScreenHeight-30, FaceSmall, color.RGBA{80, 80, 80, 200})
}

// handleSlotSelectInput slot seçim ekranının girişini işler.
func (r *Renderer) handleSlotSelectInput(saveMode bool) InputAction {
	n := len(SaveSlots)
	if n == 0 {
		if r.keyJustPressed(ebiten.KeyEscape) {
			return InputAction{Kind: ActionBack}
		}
		return InputAction{}
	}

	// Onay bekleniyor: sadece Enter (onayla) ve Esc (iptal) çalışır
	if r.pendingDeleteSlot != "" {
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

	mx, my := ebiten.CursorPosition()
	if i := r.slotHoverIndex(float64(mx), float64(my)); i >= 0 {
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
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if i := r.slotHoverIndex(float64(mx), float64(my)); i >= 0 {
			slot := SaveSlots[i]
			if saveMode || slot.Exists {
				return InputAction{Kind: ActionSelectSave, BuildingID: slot.Name}
			}
		}
	}
	return InputAction{}
}

// slotHoverIndex fareye göre hangi slot kartının üzerinde olduğunu döner.
func (r *Renderer) slotHoverIndex(mx, my float64) int {
	cardW := 480.0
	cardH := 88.0
	gap := 14.0
	totalH := float64(len(SaveSlots))*cardH + float64(len(SaveSlots)-1)*gap
	startY := ScreenHeight/2 - totalH/2
	cx := ScreenWidth/2 - cardW/2

	for i := range SaveSlots {
		cy := startY + float64(i)*(cardH+gap)
		if mx >= cx && mx <= cx+cardW && my >= cy && my <= cy+cardH {
			return i
		}
	}
	return -1
}
