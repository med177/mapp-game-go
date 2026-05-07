package render

import (
	"image/color"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Kart boyutları
const (
	cardW   = float32(60)
	spriteHc = float32(50) // kart içindeki sprite yüksekliği
	nameHc  = float32(13)
	hpBarH  = float32(7)
	cardH   = spriteHc + nameHc + hpBarH + 5 // ≈75px
	cardGap = float32(3)
	maxCols = 10

	armyPanelPadX = float32(12)
	armyPanelPadY = float32(8)
	armyPanelHdrH = float32(26)
)

// DrawArmyDetailPanel seçili ordunun birimlerini Total War stilinde ekranın alt
// orta kısmında birim kart ızgarası olarak gösterir.
// Her zaman 20 slot gösterilir; dolu slotlar normal, boş slotlar silik çerçeve ile.
func DrawArmyDetailPanel(screen *ebiten.Image, gs *state.GameState, aid army.ArmyID) {
	if aid == "" {
		return
	}
	a, ok := gs.Armies[aid]
	if !ok {
		return
	}

	ensureArmySheet()

	const totalSlots = army.MaxArmySize
	cols := maxCols
	rows := (totalSlots + maxCols - 1) / maxCols

	panelW := float32(cols)*(cardW+cardGap) - cardGap + armyPanelPadX*2
	panelH := armyPanelHdrH + float32(rows)*(cardH+cardGap) - cardGap + armyPanelPadY*2

	px := float32(ScreenWidth)/2 - panelW/2
	py := bottomBarTop() - panelH - 5

	// ── Arka plan ve çerçeve ──────────────────────────────────────────
	vector.FillRect(screen, px, py, panelW, panelH, panelBg, false)
	drawPanelBorder(screen, px, py, panelW, panelH)
	vector.FillRect(screen, px, py, panelW, 3, panelBorder, false)

	// ── Başlık satırı ─────────────────────────────────────────────────
	var factionName string
	factionCol := ColorGold
	for fid, f := range gs.Factions {
		if string(fid) == a.OwnerID {
			factionName = f.NameTR
			factionCol = color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
			break
		}
	}
	location := ""
	if r, ok2 := gs.Regions[a.RegionID]; ok2 {
		location = r.NameTR
	}
	headerLeft := factionName
	if location != "" {
		headerLeft += "  —  " + location
	}
	DrawText(screen, headerLeft, float64(px)+float64(armyPanelPadX), float64(py)+6, FaceSmall, factionCol)

	// Hareket puanı — sağ üst
	mpStr := "Hareket: " + itoa(a.MovePoints) + "/" + itoa(a.MaxMovePoints)
	mpCol := ColorGold
	if a.MovePoints == 0 {
		mpCol = ColorRed
	}
	mpW := MeasureText(mpStr, FaceSmall)
	DrawText(screen, mpStr,
		float64(px)+float64(panelW)-float64(armyPanelPadX)-mpW,
		float64(py)+6, FaceSmall, mpCol)

	// Ayırıcı
	sepY := py + armyPanelHdrH
	vector.StrokeLine(screen, px+armyPanelPadX, sepY, px+panelW-armyPanelPadX, sepY, 1, panelBorder, false)

	// ── Birim kartları — 20 slot, boş olanlar silik görünür ─────────────
	for i := 0; i < totalSlots; i++ {
		col := i % maxCols
		row := i / maxCols

		cx := px + armyPanelPadX + float32(col)*(cardW+cardGap)
		cy := sepY + armyPanelPadY/2 + float32(row)*(cardH+cardGap)

		if i >= len(a.Units) {
			// Boş slot — silik çerçeve
			vector.FillRect(screen, cx, cy, cardW, cardH, color.RGBA{14, 12, 8, 120}, false)
			vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, color.RGBA{45, 38, 24, 130}, false)
			// Ortada soluk artı/boş işareti
			DrawTextCentered(screen, "+", float64(cx)+float64(cardW)/2, float64(cy)+float64(cardH)/2-10,
				FaceLarge, color.RGBA{40, 35, 22, 100})
			continue
		}

		u := a.Units[i]
		utype := gs.UnitTypes[u.TypeID]
		hpPct := float64(u.CurrentHP) / 100.0

		// Kart arka planı — HP'ye göre hafif renk tonu
		cardBg := color.RGBA{28, 22, 14, 225}
		cardBorderCol := color.RGBA{72, 58, 36, 210}
		if hpPct <= 0.33 {
			cardBg = color.RGBA{40, 16, 14, 225}
			cardBorderCol = color.RGBA{140, 50, 40, 220}
		} else if hpPct <= 0.66 {
			cardBg = color.RGBA{36, 30, 14, 225}
			cardBorderCol = color.RGBA{130, 100, 30, 210}
		}
		vector.FillRect(screen, cx, cy, cardW, cardH, cardBg, false)
		vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, cardBorderCol, false)

		// Sprite
		if armySheet != nil && utype != nil {
			r := unitSpriteRect(u.TypeID, armySheet)
			if !r.Empty() {
				sub := armySheet.SubImage(r).(*ebiten.Image)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(
					float64(cardW-2)/float64(r.Dx()),
					float64(spriteHc-2)/float64(r.Dy()),
				)
				op.GeoM.Translate(float64(cx+1), float64(cy+1))
				screen.DrawImage(sub, op)
			}
		} else if utype == nil {
			DrawTextCentered(screen, "?", float64(cx)+float64(cardW)/2, float64(cy)+20, FaceLarge, ColorGray)
		}

		// Birim adı
		unitName := u.TypeID
		if utype != nil {
			unitName = utype.NameTR
		}
		nameCol := ColorWhite
		if hpPct <= 0.33 {
			nameCol = color.RGBA{220, 120, 100, 230}
		}
		DrawTextCentered(screen, unitName,
			float64(cx)+float64(cardW)/2,
			float64(cy)+float64(spriteHc)+1,
			FaceSmall, nameCol)

		// HP çubuğu
		hpY := cy + spriteHc + nameHc + 1
		var hpCol color.Color
		switch {
		case hpPct > 0.66:
			hpCol = color.RGBA{55, 195, 55, 255}
		case hpPct > 0.33:
			hpCol = color.RGBA{215, 175, 35, 255}
		default:
			hpCol = color.RGBA{210, 55, 55, 255}
		}
		drawBar(screen, cx+1, hpY, cardW-2, hpBarH-1, hpPct, hpCol)

		// Deneyim noktaları
		if u.Experience > 0 {
			xpPct := float64(u.Experience) / 100.0
			xpW := float32(xpPct * float64(cardW-2))
			vector.FillRect(screen, cx+1, hpY+hpBarH, xpW, 2, color.RGBA{80, 160, 255, 180}, false)
		}
	}
}
