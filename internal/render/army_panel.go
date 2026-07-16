package render

import (
	"fmt"
	"image"
	"image/color"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Kart boyutları
const (
	cardW           = float32(88)
	unitCardSpriteH = float32(90)
	spriteHc        = unitCardSpriteH // yetiştirme kartıyla aynı sprite yüksekliği
	nameHc          = float32(15)
	hpBarH          = float32(8)
	cardH           = spriteHc + nameHc + hpBarH + 8 // ≈121px
	cardGap         = float32(5)
	maxCols         = 10

	// Sprite sheet hücresini karta uniform ölçekle ve birimi yakından göster.
	// Yatay/dikey ayrı ölçek uygulanmaz; taşan kısım kart clipping alanında kalır.
	unitCardSpriteExtraW  = float32(68)
	unitCardSpriteExtraH  = float32(44)
	unitCardSpriteOffsetX = float32(-10)

	armyPanelPadX  = float32(12)
	armyPanelPadY  = float32(8)
	armyPanelHdrH  = float32(46)
	armyPanelTopY  = float32(6)
	armyPanelInfoY = float32(20)
	armyPanelBtnY  = float32(24)
	siegeFooterH   = float32(30)

	armyPanelCommanderW        = float32(260)
	armyPanelCommanderExtraH   = float32(12)
	armyPanelColumnGap         = float32(16)
	armyPanelCommanderPortrait = float32(92)
	armyPanelCommanderCardPad  = float32(10)
	armyPanelCommanderSectionY = float32(8)
)

type armyPanelLayout struct {
	panelX, panelY, panelW, panelH float32
	headerY                        float32
	gridX, gridY, gridW, gridH     float32
	commanderX, commanderY         float32
	commanderW, commanderH         float32
}

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
		fullIntel := playerHasRevealEnemyStrength(gs)
		siegeIntel := enemyUnderPlayerSiege(gs, a)
		if fullIntel || enemyArmyInPlayerMoveRange(gs, a) || siegeIntel {
			revealRatio := 0.50
			if siegeIntel {
				revealRatio = 0.75
			}
			drawScoutedEnemyArmyDetailPanel(screen, gs, a, fullIntel, revealRatio)
		} else {
			drawEnemyArmyDetailPanel(screen, gs, a)
		}
		return
	}

	ensureArmySheet()
	sheet := armySheetForFaction(gs, a.OwnerID)

	layout := armyPanelGeometry()
	px, py, panelW, panelH := layout.panelX, layout.panelY, layout.panelW, layout.panelH

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
	if siege := gs.SiegeByArmy(aid); siege != nil {
		if r, ok2 := gs.Regions[siege.RegionID]; ok2 {
			location = r.NameTR
		}
	}
	if location == "" {
		if r, ok2 := gs.Regions[a.RegionID]; ok2 {
			location = r.NameTR
		}
	}
	headerLeft := factionName
	if location != "" {
		headerLeft += "  —  " + location
	}
	if a.Commander != nil {
		headerLeft += "  |  Komutan: " + a.Commander.Name + " Lv." + itoa(a.Commander.Level)
	}
	if a.EmbarkedCommander != nil {
		headerLeft += "  |  Taşınan: " + a.EmbarkedCommander.Name + " Lv." + itoa(a.EmbarkedCommander.Level)
	}
	mpStr := "Hareket: " + itoa(a.MovePoints) + "/" + itoa(a.MaxMovePoints)
	mpW := MeasureText(mpStr, FaceSmall)
	actionStartX := splitButtonBlockLeft(px, panelW, hasArmyMergeActions(gs, aid))
	headerMaxW := float64(actionStartX - px - armyPanelPadX - 10)
	if rightLimited := float64(panelW) - float64(armyPanelPadX*2) - mpW - 12; rightLimited < headerMaxW {
		headerMaxW = rightLimited
	}
	if headerMaxW < 0 {
		headerMaxW = 0
	}
	DrawText(screen, trimTextToWidth(headerLeft, FaceSmall, headerMaxW), float64(px)+float64(armyPanelPadX), float64(py)+float64(armyPanelTopY), FaceSmall, factionCol)

	// Hareket puanı — sağ üst
	mpCol := ColorGold
	if a.MovePoints == 0 {
		mpCol = ColorRed
	}
	DrawText(screen, mpStr,
		float64(px)+float64(panelW)-float64(armyPanelPadX)-mpW,
		float64(py)+float64(armyPanelTopY), FaceSmall, mpCol)
	if a.CanReplenishIn(gs.Regions) && !gs.IsArmyDefendingSiegedRegion(a) && a.HasDamagedUnits() {
		healStr := "Takviye aktif"
		healW := MeasureText(healStr, FaceSmall)
		DrawText(screen, healStr,
			float64(px)+float64(panelW)-float64(armyPanelPadX)-healW,
			float64(py)+float64(armyPanelInfoY), FaceSmall, color.RGBA{110, 190, 120, 220})
	}
	if logistics, ok := gs.ArmyLogistics[aid]; ok && logistics.TotalHPDamage > 0 {
		DrawText(screen, "Lojistik zayiat: -"+itoa(logistics.DamagePerUnit)+" HP / birim",
			float64(px)+float64(armyPanelPadX), float64(py)+float64(armyPanelInfoY), FaceSmall, color.RGBA{210, 96, 82, 235})
	}

	// Aksiyon butonları — BÖL ve BİRLEŞTİR
	mergeTarget := FindMergeTarget(gs, aid)
	hasMerge := mergeTarget != ""
	canSplit := len(a.Units) >= 2
	if canSplit || hasMerge {
		drawArmyActionButton(screen, px, py, panelW, "✂ BÖL", canSplit, hasMerge, true)
	}
	if hasMerge {
		other := gs.Armies[mergeTarget]
		canMerge := len(other.Units) < army.MaxArmySize
		drawArmyActionButton(screen, px, py, panelW, "⊕ BİRLEŞTİR", canMerge, true, false)
	}

	// Ayırıcı
	sepY := layout.headerY
	vector.StrokeLine(screen, px+armyPanelPadX, sepY, px+panelW-armyPanelPadX, sepY, 1, panelBorder, false)
	vector.StrokeLine(screen, layout.commanderX+layout.commanderW+armyPanelColumnGap/2, sepY+armyPanelPadY/2, layout.commanderX+layout.commanderW+armyPanelColumnGap/2, py+panelH-siegeFooterH-armyPanelPadY/2, 1, color.RGBA{70, 56, 32, 180}, false)

	drawArmyCommanderCard(screen, a, layout)

	// ── Birim kartları — 20 slot, boş olanlar silik görünür ─────────────
	const totalSlots = army.MaxArmySize
	for i := 0; i < totalSlots; i++ {
		col := i % maxCols
		row := i / maxCols

		cx := layout.gridX + float32(col)*(cardW+cardGap)
		cy := layout.gridY + float32(row)*(cardH+cardGap)

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
		hpPct := u.HPPercent()
		isReplenishing := a.CanReplenishIn(gs.Regions) && !gs.IsArmyDefendingSiegedRegion(a) && u.CurrentHP < army.MaxUnitHP

		// Kart arka planı sabit beyaz.
		cardBg := color.RGBA{255, 255, 255, 245}
		cardBorderCol := color.RGBA{160, 160, 160, 225}
		vector.FillRect(screen, cx, cy, cardW, cardH, cardBg, false)
		vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, cardBorderCol, false)
		if isReplenishing {
			badgeW := float32(18)
			badgeH := float32(12)
			badgeX := cx + cardW - badgeW - 3
			badgeY := cy + 3
			vector.FillRect(screen, badgeX, badgeY, badgeW, badgeH, color.RGBA{70, 150, 84, 235}, false)
			DrawTextCentered(screen, "+", float64(badgeX)+float64(badgeW)/2, float64(badgeY)-1, FaceSmall, color.RGBA{245, 255, 245, 255})
		}

		// Sprite: en-boy oranını koruyarak karta ortala (distorsiyon yok).
		if sheet != nil && utype != nil {
			r := unitSpriteRect(u.TypeID, sheet)
			if !r.Empty() {
				sub := sheet.SubImage(r).(*ebiten.Image)
				op := &ebiten.DrawImageOptions{}
				// Recruit panel ile aynı davranış: biraz taşır, ortalar ve karta kırpar.
				fitW := float64(cardW + unitCardSpriteExtraW)
				fitH := float64(spriteHc + unitCardSpriteExtraH)
				scale := fitW / float64(r.Dx())
				if hScale := fitH / float64(r.Dy()); hScale < scale {
					scale = hScale
				}
				drawW := float64(r.Dx()) * scale
				drawH := float64(r.Dy()) * scale
				if recruitClipBuf != nil {
					clipW := int(cardW - 2)
					clipH := int(spriteHc - 2)
					if clipW > 0 && clipH > 0 && clipW <= 160 && clipH <= 120 {
						recruitClipBuf.Clear()
						op.GeoM.Scale(scale, scale)
						op.GeoM.Translate(float64(clipW)/2-drawW/2+float64(unitCardSpriteOffsetX), float64(clipH)/2-drawH/2)
						recruitClipBuf.DrawImage(sub, op)
						cropped := recruitClipBuf.SubImage(image.Rect(0, 0, clipW, clipH)).(*ebiten.Image)
						dst := &ebiten.DrawImageOptions{}
						dst.GeoM.Translate(float64(cx+1), float64(cy+1))
						screen.DrawImage(cropped, dst)
					}
				}
			}
		} else if utype == nil {
			DrawTextCentered(screen, "?", float64(cx)+float64(cardW)/2, float64(cy)+20, FaceLarge, ColorGray)
		}

		// Birim adı
		unitName := u.TypeID
		if utype != nil {
			unitName = utype.NameTR
		}
		nameCol := color.RGBA{25, 25, 25, 235}
		if hpPct <= 0.33 {
			nameCol = color.RGBA{140, 35, 35, 235}
		}
		DrawTextCentered(screen, shortUnitName(unitName, 14),
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

func drawEnemyArmyCommanderCard(screen *ebiten.Image, a *army.Army, layout armyPanelLayout) {
	commander, role := armyPanelDisplayedCommander(a)
	drawCommanderSummaryCard(screen, commander, float64(layout.commanderX), float64(layout.commanderY), float64(layout.commanderW), float64(layout.commanderH), commanderCardOptions{
		Role:            role,
		EmptySummary:    "-",
		EmptyEffectText: "-",
		ShowEffectText:  false,
		MaxTraitRows:    1,
	})
}

func armyDetailPanelRect(gs *state.GameState, aid army.ArmyID) (gameui.Rect, bool) {
	if gs == nil || aid == "" {
		return gameui.Rect{}, false
	}
	a, ok := gs.Armies[aid]
	if !ok || a == nil {
		return gameui.Rect{}, false
	}
	layout := armyPanelGeometry()
	return gameui.Rect{
		X: float64(layout.panelX),
		Y: float64(layout.panelY),
		W: float64(layout.panelW),
		H: float64(layout.panelH),
	}, true
}

func drawEnemyArmyDetailPanel(screen *ebiten.Image, gs *state.GameState, a *army.Army) {
	layout := armyPanelGeometry()
	px, py, panelW, panelH := layout.panelX, layout.panelY, layout.panelW, layout.panelH

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

	headerLeft := "Rakip Ordu: " + factionName
	if location != "" {
		headerLeft += "  —  " + location
	}
	headerRight := "Birim sayısı bilinmiyor"
	rightW := MeasureText(headerRight, FaceSmall)
	headerMaxW := float64(panelW) - float64(armyPanelPadX*2) - rightW - 12
	if headerMaxW < 0 {
		headerMaxW = 0
	}
	DrawText(screen, trimTextToWidth(headerLeft, FaceSmall, headerMaxW), float64(px)+float64(armyPanelPadX), float64(py)+float64(armyPanelTopY), FaceSmall, factionCol)
	DrawText(screen, headerRight, float64(px)+float64(panelW)-float64(armyPanelPadX)-rightW, float64(py)+float64(armyPanelTopY), FaceSmall, color.RGBA{190, 160, 90, 230})

	sepY := layout.headerY
	vector.StrokeLine(screen, px+armyPanelPadX, sepY, px+panelW-armyPanelPadX, sepY, 1, panelBorder, false)
	vector.StrokeLine(screen, layout.commanderX+layout.commanderW+armyPanelColumnGap/2, sepY+armyPanelPadY/2, layout.commanderX+layout.commanderW+armyPanelColumnGap/2, py+panelH-siegeFooterH-armyPanelPadY/2, 1, color.RGBA{70, 56, 32, 180}, false)

	drawEnemyArmyCommanderCard(screen, a, layout)

	const totalSlots = army.MaxArmySize
	for i := 0; i < totalSlots; i++ {
		col := i % maxCols
		row := i / maxCols
		cx := layout.gridX + float32(col)*(cardW+cardGap)
		cy := layout.gridY + float32(row)*(cardH+cardGap)
		drawUnknownEnemyUnitCard(screen, cx, cy)
	}

	vector.FillRect(screen, px, py+panelH-siegeFooterH, panelW, siegeFooterH, color.RGBA{28, 18, 6, 190}, false)
	vector.StrokeLine(screen, px, py+panelH-siegeFooterH, px+panelW, py+panelH-siegeFooterH, 1, panelBorder, false)
	DrawText(screen, "Bu orduya hareket emri verilemez  |  Birim detayları gizli", float64(px)+float64(armyPanelPadX), float64(py+panelH-siegeFooterH+2), FaceSmall, color.RGBA{180, 100, 90, 210})
}

func playerHasRevealEnemyStrength(gs *state.GameState) bool {
	if gs == nil || gs.TechTypes == nil || gs.PlayerFactionID == "" {
		return false
	}
	f := gs.Factions[gs.PlayerFactionID]
	if f == nil {
		return false
	}
	return tech.ComputeEffects(f.Research.Completed, gs.TechTypes).RevealEnemyStrength
}

func enemyUnderPlayerSiege(gs *state.GameState, a *army.Army) bool {
	if gs == nil || a == nil || a.IsNaval {
		return false
	}
	if siege := gs.SiegeAt(a.RegionID); siege != nil {
		if siegeArmy := gs.Armies[siege.AttackerArmyID]; siegeArmy != nil && siegeArmy.OwnerID == string(gs.PlayerFactionID) {
			return true
		}
	}
	return false
}

func scoutedEnemyRevealCount(total int, fullIntel bool, revealRatio float64) int {
	if total <= 0 {
		return 0
	}
	if fullIntel {
		return total
	}
	if revealRatio <= 0 {
		revealRatio = 0.5
	}
	revealed := int(float64(total)*revealRatio + 0.5)
	if revealed < 1 {
		revealed = 1
	}
	if revealed > total {
		revealed = total
	}
	return revealed
}

func drawScoutedEnemyArmyDetailPanel(screen *ebiten.Image, gs *state.GameState, a *army.Army, fullIntel bool, revealRatio float64) {
	ensureArmySheet()
	sheet := armySheetForFaction(gs, a.OwnerID)

	const totalSlots = army.MaxArmySize
	layout := armyPanelGeometry()
	px, py, panelW, panelH := layout.panelX, layout.panelY, layout.panelW, layout.panelH

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
	countStr := "Birim: " + itoa(len(a.Units)) + "  |  Kısmi istihbarat (%" + itoa(int(revealRatio*100+0.5)) + ")"
	if fullIntel {
		countStr = "Birim: " + itoa(len(a.Units)) + "  |  Tam istihbarat"
	}
	countW := MeasureText(countStr, FaceSmall)
	headerMaxW := float64(panelW) - float64(armyPanelPadX*2) - countW - 12
	if headerMaxW < 0 {
		headerMaxW = 0
	}
	DrawText(screen, trimTextToWidth(headerLeft, FaceSmall, headerMaxW), float64(px)+float64(armyPanelPadX), float64(py)+float64(armyPanelTopY), FaceSmall, factionCol)
	DrawText(screen, countStr,
		float64(px)+float64(panelW)-float64(armyPanelPadX)-countW,
		float64(py)+float64(armyPanelTopY), FaceSmall, color.RGBA{190, 160, 90, 230})

	sepY := layout.headerY
	vector.StrokeLine(screen, px+armyPanelPadX, sepY, px+panelW-armyPanelPadX, sepY, 1, panelBorder, false)
	vector.StrokeLine(screen, layout.commanderX+layout.commanderW+armyPanelColumnGap/2, sepY+armyPanelPadY/2, layout.commanderX+layout.commanderW+armyPanelColumnGap/2, py+panelH-siegeFooterH-armyPanelPadY/2, 1, color.RGBA{70, 56, 32, 180}, false)

	drawEnemyArmyCommanderCard(screen, a, layout)

	revealed := scoutedEnemyRevealCount(len(a.Units), fullIntel, revealRatio)
	for i := 0; i < totalSlots; i++ {
		col := i % maxCols
		row := i / maxCols
		cx := layout.gridX + float32(col)*(cardW+cardGap)
		cy := layout.gridY + float32(row)*(cardH+cardGap)

		if i >= len(a.Units) {
			vector.FillRect(screen, cx, cy, cardW, cardH, color.RGBA{14, 12, 8, 90}, false)
			vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, color.RGBA{45, 38, 24, 95}, false)
			continue
		}
		if i >= revealed {
			drawUnknownEnemyUnitCard(screen, cx, cy)
			continue
		}
		drawScoutedEnemyUnitCard(screen, gs, sheet, a.Units[i], cx, cy)
	}

	vector.FillRect(screen, px, py+panelH-siegeFooterH, panelW, siegeFooterH, color.RGBA{28, 18, 6, 190}, false)
	vector.StrokeLine(screen, px, py+panelH-siegeFooterH, px+panelW, py+panelH-siegeFooterH, 1, panelBorder, false)
	footer := "Bu orduya hareket emri verilemez"
	if fullIntel {
		footer += "  |  Tam istihbarat"
	} else {
		footer += "  |  Kısmi istihbarat"
	}
	DrawText(screen, footer, float64(px)+float64(armyPanelPadX), float64(py+panelH-siegeFooterH+2), FaceSmall, color.RGBA{180, 100, 90, 210})
}

func drawUnknownEnemyUnitCard(screen *ebiten.Image, cx, cy float32) {
	vector.FillRect(screen, cx, cy, cardW, cardH, color.RGBA{24, 20, 16, 220}, false)
	vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, color.RGBA{95, 75, 45, 210}, false)
	DrawTextCentered(screen, "?", float64(cx)+float64(cardW)/2, float64(cy)+20, FaceLarge, color.RGBA{210, 180, 90, 230})
	DrawTextCentered(screen, "Gizli", float64(cx)+float64(cardW)/2, float64(cy)+float64(spriteHc)+1, FaceSmall, color.RGBA{150, 130, 90, 220})
	drawBar(screen, cx+1, cy+spriteHc+nameHc+1, cardW-2, hpBarH-1, 1, color.RGBA{80, 70, 55, 180})
}

func drawScoutedEnemyUnitCard(screen *ebiten.Image, gs *state.GameState, sheet *ebiten.Image, u army.Unit, cx, cy float32) {
	utype := gs.UnitTypes[u.TypeID]
	vector.FillRect(screen, cx, cy, cardW, cardH, color.RGBA{255, 255, 255, 245}, false)
	vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, color.RGBA{160, 160, 160, 225}, false)

	if sheet != nil && utype != nil {
		r := unitSpriteRect(u.TypeID, sheet)
		if !r.Empty() {
			sub := sheet.SubImage(r).(*ebiten.Image)
			op := &ebiten.DrawImageOptions{}
			fitW := float64(cardW + unitCardSpriteExtraW)
			fitH := float64(spriteHc + unitCardSpriteExtraH)
			scale := fitW / float64(r.Dx())
			if hScale := fitH / float64(r.Dy()); hScale < scale {
				scale = hScale
			}
			drawW := float64(r.Dx()) * scale
			drawH := float64(r.Dy()) * scale
			if recruitClipBuf != nil {
				clipW := int(cardW - 2)
				clipH := int(spriteHc - 2)
				if clipW > 0 && clipH > 0 && clipW <= 160 && clipH <= 120 {
					recruitClipBuf.Clear()
					op.GeoM.Scale(scale, scale)
					op.GeoM.Translate(float64(clipW)/2-drawW/2+float64(unitCardSpriteOffsetX), float64(clipH)/2-drawH/2)
					op.ColorScale.Scale(0.85, 0.85, 0.85, 1)
					recruitClipBuf.DrawImage(sub, op)
					cropped := recruitClipBuf.SubImage(image.Rect(0, 0, clipW, clipH)).(*ebiten.Image)
					dst := &ebiten.DrawImageOptions{}
					dst.GeoM.Translate(float64(cx+1), float64(cy+1))
					screen.DrawImage(cropped, dst)
				}
			}
		}
	}

	unitName := u.TypeID
	if utype != nil {
		unitName = utype.NameTR
	}
	DrawTextCentered(screen, shortUnitName(unitName, 14),
		float64(cx)+float64(cardW)/2,
		float64(cy)+float64(spriteHc)+1,
		FaceSmall, color.RGBA{25, 25, 25, 235})
	drawBar(screen, cx+1, cy+spriteHc+nameHc+1, cardW-2, hpBarH-1, 1, color.RGBA{120, 110, 85, 210})
}

const (
	actionBtnW          = float32(92)
	actionBtnH          = float32(18)
	actionBtnGap        = float32(8)
	actionBtnRightInset = float32(92)
)

// armyPanelGeometry panel px/py/panelW değerlerini hesaplar.
func armyPanelGeometry() armyPanelLayout {
	const totalSlots = army.MaxArmySize
	cols := maxCols
	rows := (totalSlots + maxCols - 1) / maxCols
	gridW := float32(cols)*(cardW+cardGap) - cardGap
	gridH := float32(rows)*(cardH+cardGap) - cardGap
	panelW := gridW + armyPanelCommanderW + armyPanelColumnGap + armyPanelPadX*2
	panelH := armyPanelHdrH + float32(rows)*(cardH+cardGap) - cardGap + armyPanelPadY*2 + siegeFooterH + armyPanelCommanderExtraH
	px := float32(ScreenWidth)/2 - panelW/2
	py := bottomBarTop() - panelH - 55
	headerY := py + armyPanelHdrH
	gridY := headerY + armyPanelPadY/2
	return armyPanelLayout{
		panelX:     px,
		panelY:     py,
		panelW:     panelW,
		panelH:     panelH,
		headerY:    headerY,
		gridX:      px + armyPanelPadX + armyPanelCommanderW + armyPanelColumnGap,
		gridY:      gridY,
		gridW:      gridW,
		gridH:      gridH,
		commanderX: px + armyPanelPadX,
		commanderY: gridY,
		commanderW: armyPanelCommanderW,
		commanderH: panelH - armyPanelHdrH - armyPanelPadY - siegeFooterH,
	}
}

func buildArmyPanelCloseButton() gameui.Button {
	layout := armyPanelGeometry()
	x, y, w, h := panelCloseRect(layout.panelX, layout.panelY, layout.panelW)
	btn := gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "").WithIcon(gameui.IconClose)
	btn.IconSize = 12
	return btn
}

func commanderPortraitHitRect(gs *state.GameState, aid army.ArmyID) (gameui.Rect, bool) {
	if gs == nil || aid == "" {
		return gameui.Rect{}, false
	}
	a := gs.Armies[aid]
	if a == nil || a.OwnerID != string(gs.PlayerFactionID) {
		return gameui.Rect{}, false
	}
	layout := armyPanelGeometry()
	return gameui.Rect{
		X: float64(layout.commanderX + armyPanelCommanderCardPad),
		Y: float64(layout.commanderY + 28),
		W: float64(armyPanelCommanderPortrait),
		H: float64(armyPanelCommanderPortrait),
	}, true
}

// splitButtonRect BÖL butonunun piksel dikdörtgenini döner.
// hasMerge true ise iki buton yan yana olacak şekilde sola kayar.
func splitButtonRect(px, py, panelW float32, hasMerge bool) (bx, by, bw, bh float32) {
	bw, bh = actionBtnW, actionBtnH
	by = py + armyPanelBtnY
	bx = splitButtonBlockLeft(px, panelW, hasMerge)
	return
}

func splitButtonBlockLeft(px, panelW float32, hasMerge bool) float32 {
	if hasMerge {
		return px + panelW - armyPanelPadX - actionBtnW*2 - actionBtnGap - actionBtnRightInset
	}
	return px + panelW - armyPanelPadX - actionBtnW - actionBtnRightInset
}

// mergeButtonRect BİRLEŞTİR butonunun piksel dikdörtgenini döner.
func mergeButtonRect(px, py, panelW float32) (bx, by, bw, bh float32) {
	bw, bh = actionBtnW, actionBtnH
	by = py + armyPanelBtnY
	bx = px + panelW - armyPanelPadX - actionBtnW - actionBtnRightInset
	return
}

func buildSplitArmyButton(gs *state.GameState, aid army.ArmyID) (gameui.Button, bool) {
	if aid == "" {
		return gameui.Button{}, false
	}
	a, ok := gs.Armies[aid]
	if !ok || len(a.Units) < 2 {
		return gameui.Button{}, false
	}
	layout := armyPanelGeometry()
	hasMerge := FindMergeTarget(gs, aid) != ""
	bx, by, bw, bh := splitButtonRect(layout.panelX, layout.panelY, layout.panelW, hasMerge)
	return gameui.NewButton(float64(bx), float64(by), float64(bw), float64(bh), "✂ BÖL"), true
}

func buildMergeArmyButton(gs *state.GameState, aid army.ArmyID) (gameui.Button, bool) {
	if FindMergeTarget(gs, aid) == "" {
		return gameui.Button{}, false
	}
	layout := armyPanelGeometry()
	bx, by, bw, bh := mergeButtonRect(layout.panelX, layout.panelY, layout.panelW)
	return gameui.NewButton(float64(bx), float64(by), float64(bw), float64(bh), "⊕ BİRLEŞTİR"), true
}

// drawArmyActionButton tek bir aksiyon butonunu çizer.
// isSplit true → sol buton (BÖL), false → sağ buton (BİRLEŞTİR).
func drawArmyActionButton(screen *ebiten.Image, px, py, panelW float32, label string, active, hasMerge, isSplit bool) {
	var bx, by, bw, bh float32
	if isSplit {
		bx, by, bw, bh = splitButtonRect(px, py, panelW, hasMerge)
	} else {
		bx, by, bw, bh = mergeButtonRect(px, py, panelW)
	}
	drawArmyPanelButton(screen, bx, by, bw, bh, label, active)
}

func drawArmyPanelButton(screen *ebiten.Image, x, y, w, h float32, label string, active bool) {
	bg := color.RGBA{50, 35, 12, 220}
	border := color.RGBA{160, 120, 40, 200}
	txt := color.RGBA{220, 185, 70, 255}
	if !active {
		bg = color.RGBA{30, 25, 18, 140}
		border = color.RGBA{55, 45, 28, 120}
		txt = color.RGBA{90, 80, 55, 160}
	}
	vector.FillRect(screen, x, y, w, h, bg, false)
	vector.FillRect(screen, x, y, w, 2, color.RGBA{208, 170, 72, 230}, false)
	vector.StrokeRect(screen, x, y, w, h, 1, border, false)
	tw := float32(MeasureText(label, FaceSmall))
	DrawText(screen, label, float64(x)+float64(w)/2-float64(tw)/2, float64(y)+3, FaceSmall, txt)
}

func hasArmyMergeActions(gs *state.GameState, aid army.ArmyID) bool {
	return FindMergeTarget(gs, aid) != ""
}

func armyPanelDisplayedCommander(a *army.Army) (*army.Commander, string) {
	if a == nil {
		return nil, "Komutan"
	}
	if a.Commander != nil {
		if a.IsNaval {
			return a.Commander, "Filo Komutanı"
		}
		return a.Commander, "Komutan"
	}
	if a.EmbarkedCommander != nil {
		return a.EmbarkedCommander, "Taşınan Komutan"
	}
	return nil, "Komutan"
}

func drawArmyCommanderCard(screen *ebiten.Image, a *army.Army, layout armyPanelLayout) {
	commander, role := armyPanelDisplayedCommander(a)
	extra := ""
	if a != nil && a.Commander != nil && a.EmbarkedCommander != nil && a.Commander != a.EmbarkedCommander {
		extra = fmt.Sprintf("Taşınan: %s Lv.%d", a.EmbarkedCommander.Name, a.EmbarkedCommander.Level)
	}
	drawCommanderSummaryCard(screen, commander, float64(layout.commanderX), float64(layout.commanderY), float64(layout.commanderW), float64(layout.commanderH), commanderCardOptions{
		Role:            role,
		EmptySummary:    "Komutan atayarak savaş, hareket ve kuşatma bonusu kazan.",
		EmptyEffectText: "Katkı: atanmış komutan yok.",
		ExtraLine:       extra,
		ShowEffectText:  false,
		MaxTraitRows:    1,
		BottomInset:     8,
	})
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

// SplitButtonHitTest fare BÖL butonuna denk geliyorsa true döner.
func SplitButtonHitTest(fx, fy float64, gs *state.GameState, aid army.ArmyID) bool {
	btn, ok := buildSplitArmyButton(gs, aid)
	return ok && btn.HitTest(fx, fy)
}

// MergeButtonHitTest fare BİRLEŞTİR butonuna denk geliyorsa true döner.
func MergeButtonHitTest(fx, fy float64, gs *state.GameState, aid army.ArmyID) bool {
	btn, ok := buildMergeArmyButton(gs, aid)
	return ok && btn.HitTest(fx, fy)
}

func CommanderPortraitHitTest(fx, fy float64, gs *state.GameState, aid army.ArmyID) bool {
	rect, ok := commanderPortraitHitRect(gs, aid)
	return ok && rect.Hit(fx, fy)
}

func ArmyPanelBoundsHit(fx, fy float64, gs *state.GameState, aid army.ArmyID) bool {
	rect, ok := armyDetailPanelRect(gs, aid)
	return ok && rect.Hit(fx, fy)
}

func ArmyPanelInteractiveHit(fx, fy float64, gs *state.GameState, aid army.ArmyID) bool {
	if buildArmyPanelCloseButton().HitTest(fx, fy) {
		return true
	}
	if SplitButtonHitTest(fx, fy, gs, aid) || MergeButtonHitTest(fx, fy, gs, aid) {
		return true
	}
	if CommanderPortraitHitTest(fx, fy, gs, aid) {
		return true
	}
	return false
}
