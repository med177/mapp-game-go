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
	cardW    = float32(60)
	spriteHc = float32(50) // kart içindeki sprite yüksekliği
	nameHc   = float32(13)
	hpBarH   = float32(7)
	cardH    = spriteHc + nameHc + hpBarH + 5 // ≈75px
	cardGap  = float32(3)
	maxCols  = 10

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
	if a.OwnerID != string(gs.PlayerFactionID) {
		if enemyArmyInPlayerMoveRange(gs, a) {
			drawScoutedEnemyArmyDetailPanel(screen, gs, a)
		} else {
			drawEnemyArmyDetailPanel(screen, gs, a)
		}
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

	// Aksiyon butonları — BÖLDÜR ve BİRLEŞTİR
	mergeTarget := FindMergeTarget(gs, aid)
	hasMerge := mergeTarget != ""
	canSplit := len(a.Units) >= 2
	if canSplit || hasMerge {
		drawArmyActionButton(screen, px, py, panelW, "✂ BÖLDÜR", canSplit, hasMerge, true)
	}
	if hasMerge {
		other := gs.Armies[mergeTarget]
		canMerge := len(other.Units) < army.MaxArmySize
		drawArmyActionButton(screen, px, py, panelW, "⊕ BİRLEŞTİR", canMerge, true, false)
	}

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

func drawEnemyArmyDetailPanel(screen *ebiten.Image, gs *state.GameState, a *army.Army) {
	panelW := float32(380)
	panelH := float32(96)
	px := float32(ScreenWidth)/2 - panelW/2
	py := bottomBarTop() - panelH - 5

	vector.FillRect(screen, px, py, panelW, panelH, panelBg, false)
	drawPanelBorder(screen, px, py, panelW, panelH)
	vector.FillRect(screen, px, py, panelW, 3, panelBorder, false)

	factionName := "Bilinmeyen Fraksiyon"
	factionCol := ColorGold
	for fid, f := range gs.Factions {
		if string(fid) == a.OwnerID {
			factionName = f.NameTR
			factionCol = color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
			break
		}
	}
	location := "Bilinmeyen konum"
	if r, ok := gs.Regions[a.RegionID]; ok {
		location = r.NameTR
	}

	DrawText(screen, "Düşman Ordu", float64(px)+14, float64(py)+10, FaceMed, factionCol)
	DrawText(screen, factionName+"  —  "+location, float64(px)+14, float64(py)+34, FaceSmall, ColorGray)
	DrawText(screen, "Birim ve hareket detayları bilinmiyor", float64(px)+14, float64(py)+56, FaceSmall, color.RGBA{160, 140, 100, 210})
	DrawText(screen, "Bu orduya hareket emri verilemez", float64(px)+14, float64(py)+74, FaceSmall, color.RGBA{180, 100, 90, 210})
}

func drawScoutedEnemyArmyDetailPanel(screen *ebiten.Image, gs *state.GameState, a *army.Army) {
	ensureArmySheet()

	const totalSlots = army.MaxArmySize
	cols := maxCols
	rows := (totalSlots + maxCols - 1) / maxCols

	panelW := float32(cols)*(cardW+cardGap) - cardGap + armyPanelPadX*2
	panelH := armyPanelHdrH + float32(rows)*(cardH+cardGap) - cardGap + armyPanelPadY*2
	px := float32(ScreenWidth)/2 - panelW/2
	py := bottomBarTop() - panelH - 5

	vector.FillRect(screen, px, py, panelW, panelH, panelBg, false)
	drawPanelBorder(screen, px, py, panelW, panelH)
	vector.FillRect(screen, px, py, panelW, 3, panelBorder, false)

	factionName := "Bilinmeyen Fraksiyon"
	factionCol := ColorGold
	for fid, f := range gs.Factions {
		if string(fid) == a.OwnerID {
			factionName = f.NameTR
			factionCol = color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
			break
		}
	}
	location := ""
	if r, ok := gs.Regions[a.RegionID]; ok {
		location = r.NameTR
	}
	headerLeft := "Keşfedilen Düşman: " + factionName
	if location != "" {
		headerLeft += "  —  " + location
	}
	DrawText(screen, headerLeft, float64(px)+float64(armyPanelPadX), float64(py)+6, FaceSmall, factionCol)

	countStr := "Birim: " + itoa(len(a.Units)) + "  |  Kısmi istihbarat"
	countW := MeasureText(countStr, FaceSmall)
	DrawText(screen, countStr,
		float64(px)+float64(panelW)-float64(armyPanelPadX)-countW,
		float64(py)+6, FaceSmall, color.RGBA{190, 160, 90, 230})

	sepY := py + armyPanelHdrH
	vector.StrokeLine(screen, px+armyPanelPadX, sepY, px+panelW-armyPanelPadX, sepY, 1, panelBorder, false)

	revealed := (len(a.Units) + 1) / 2
	if revealed < 1 && len(a.Units) > 0 {
		revealed = 1
	}
	for i := 0; i < totalSlots; i++ {
		col := i % maxCols
		row := i / maxCols
		cx := px + armyPanelPadX + float32(col)*(cardW+cardGap)
		cy := sepY + armyPanelPadY/2 + float32(row)*(cardH+cardGap)

		if i >= len(a.Units) {
			vector.FillRect(screen, cx, cy, cardW, cardH, color.RGBA{14, 12, 8, 90}, false)
			vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, color.RGBA{45, 38, 24, 95}, false)
			continue
		}
		if i >= revealed {
			drawUnknownEnemyUnitCard(screen, cx, cy)
			continue
		}
		drawScoutedEnemyUnitCard(screen, gs, a.Units[i], cx, cy)
	}
}

func drawUnknownEnemyUnitCard(screen *ebiten.Image, cx, cy float32) {
	vector.FillRect(screen, cx, cy, cardW, cardH, color.RGBA{24, 20, 16, 220}, false)
	vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, color.RGBA{95, 75, 45, 210}, false)
	DrawTextCentered(screen, "?", float64(cx)+float64(cardW)/2, float64(cy)+20, FaceLarge, color.RGBA{210, 180, 90, 230})
	DrawTextCentered(screen, "Gizli", float64(cx)+float64(cardW)/2, float64(cy)+float64(spriteHc)+1, FaceSmall, color.RGBA{150, 130, 90, 220})
	drawBar(screen, cx+1, cy+spriteHc+nameHc+1, cardW-2, hpBarH-1, 1, color.RGBA{80, 70, 55, 180})
}

func drawScoutedEnemyUnitCard(screen *ebiten.Image, gs *state.GameState, u army.Unit, cx, cy float32) {
	utype := gs.UnitTypes[u.TypeID]
	vector.FillRect(screen, cx, cy, cardW, cardH, color.RGBA{28, 22, 14, 225}, false)
	vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, color.RGBA{115, 85, 45, 220}, false)

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
			op.ColorScale.Scale(0.85, 0.85, 0.85, 1)
			screen.DrawImage(sub, op)
		}
	}

	unitName := u.TypeID
	if utype != nil {
		unitName = utype.NameTR
	}
	DrawTextCentered(screen, unitName,
		float64(cx)+float64(cardW)/2,
		float64(cy)+float64(spriteHc)+1,
		FaceSmall, ColorWhite)
	drawBar(screen, cx+1, cy+spriteHc+nameHc+1, cardW-2, hpBarH-1, 1, color.RGBA{120, 110, 85, 210})
}

const (
	actionBtnW   = float32(92)
	actionBtnH   = float32(18)
	actionBtnGap = float32(8)
)

// armyPanelGeometry panel px/py/panelW değerlerini hesaplar.
func armyPanelGeometry() (px, py, panelW float32) {
	const totalSlots = army.MaxArmySize
	cols := maxCols
	rows := (totalSlots + maxCols - 1) / maxCols
	panelW = float32(cols)*(cardW+cardGap) - cardGap + armyPanelPadX*2
	panelH := armyPanelHdrH + float32(rows)*(cardH+cardGap) - cardGap + armyPanelPadY*2
	px = float32(ScreenWidth)/2 - panelW/2
	py = bottomBarTop() - panelH - 5
	return
}

// splitButtonRect BÖLDÜR butonunun piksel dikdörtgenini döner.
// hasMerge true ise iki buton yan yana olacak şekilde sola kayar.
func splitButtonRect(px, py, panelW float32, hasMerge bool) (bx, by, bw, bh float32) {
	bw, bh = actionBtnW, actionBtnH
	by = py + 4
	if hasMerge {
		bx = px + panelW/2 - actionBtnGap/2 - bw
	} else {
		bx = px + panelW/2 - bw/2
	}
	return
}

// mergeButtonRect BİRLEŞTİR butonunun piksel dikdörtgenini döner.
func mergeButtonRect(px, py, panelW float32) (bx, by, bw, bh float32) {
	bw, bh = actionBtnW, actionBtnH
	by = py + 4
	bx = px + panelW/2 + actionBtnGap/2
	return
}

// drawArmyActionButton tek bir aksiyon butonunu çizer.
// isSplit true → sol buton (BÖLDÜR), false → sağ buton (BİRLEŞTİR).
func drawArmyActionButton(screen *ebiten.Image, px, py, panelW float32, label string, active, hasMerge, isSplit bool) {
	var bx, by, bw, bh float32
	if isSplit {
		bx, by, bw, bh = splitButtonRect(px, py, panelW, hasMerge)
	} else {
		bx, by, bw, bh = mergeButtonRect(px, py, panelW)
	}
	bg := color.RGBA{50, 35, 12, 220}
	border := color.RGBA{160, 120, 40, 200}
	txt := color.RGBA{220, 185, 70, 255}
	if !active {
		bg = color.RGBA{30, 25, 18, 140}
		border = color.RGBA{55, 45, 28, 120}
		txt = color.RGBA{90, 80, 55, 160}
	}
	vector.FillRect(screen, bx, by, bw, bh, bg, false)
	vector.StrokeRect(screen, bx, by, bw, bh, 1, border, false)
	tw := float32(MeasureText(label, FaceSmall))
	DrawText(screen, label, float64(bx)+float64(bw)/2-float64(tw)/2, float64(by)+3, FaceSmall, txt)
}

// FindMergeTarget aynı bölgede aynı türde (naval/kara) başka dost ordu varsa ID'sini döner.
func FindMergeTarget(gs *state.GameState, aid army.ArmyID) army.ArmyID {
	a, ok := gs.Armies[aid]
	if !ok {
		return ""
	}
	for otherID, other := range gs.Armies {
		if otherID == aid || other.RegionID != a.RegionID ||
			other.OwnerID != a.OwnerID || other.IsNaval != a.IsNaval {
			continue
		}
		return otherID
	}
	return ""
}

// SplitButtonHitTest fare BÖLDÜR butonuna denk geliyorsa true döner.
func SplitButtonHitTest(fx, fy float64, gs *state.GameState, aid army.ArmyID) bool {
	if aid == "" {
		return false
	}
	a, ok := gs.Armies[aid]
	if !ok || len(a.Units) < 2 {
		return false
	}
	px, py, panelW := armyPanelGeometry()
	hasMerge := FindMergeTarget(gs, aid) != ""
	bx, by, bw, bh := splitButtonRect(px, py, panelW, hasMerge)
	return fx >= float64(bx) && fx <= float64(bx+bw) && fy >= float64(by) && fy <= float64(by+bh)
}

// MergeButtonHitTest fare BİRLEŞTİR butonuna denk geliyorsa true döner.
func MergeButtonHitTest(fx, fy float64, gs *state.GameState, aid army.ArmyID) bool {
	if FindMergeTarget(gs, aid) == "" {
		return false
	}
	px, py, panelW := armyPanelGeometry()
	bx, by, bw, bh := mergeButtonRect(px, py, panelW)
	return fx >= float64(bx) && fx <= float64(bx+bw) && fy >= float64(by) && fy <= float64(by+bh)
}
