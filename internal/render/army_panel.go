package render

import (
	"fmt"
	"image/color"
	"sort"

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
	unitCardSpriteH = cardW * unitSpriteAspectH
	spriteHc        = unitCardSpriteH // yetiştirme kartıyla aynı sprite yüksekliği
	nameHc          = float32(15)
	hpBarH          = float32(8)
	cardH           = spriteHc
	cardGap         = float32(5)
	maxCols         = 10

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
func DrawArmyDetailPanel(screen *ebiten.Image, gs *state.GameState, aid army.ArmyID, selectedUnitMaps ...map[int]bool) {
	if aid == "" {
		return
	}
	a, ok := gs.Armies[aid]
	if !ok {
		return
	}
	if !playerCanSeeArmyDetails(gs, a) {
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

	ensureArmySprites()

	layout := armyPanelGeometry()
	px, py, panelW, panelH := layout.panelX, layout.panelY, layout.panelW, layout.panelH
	var selectedUnits map[int]bool
	if len(selectedUnitMaps) > 0 {
		selectedUnits = selectedUnitMaps[0]
	}
	selectedCount := splitSelectedUnitCount(a, selectedUnits)

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
	mergeTargets := FindMergeTargets(gs, aid)
	canSplit := len(a.Units) >= 2 && (selectedCount == 0 || selectedCount < len(a.Units))
	hasMerge := len(mergeTargets) > 0
	hasSplitButton := canSplit || hasMerge
	actionStartX := splitButtonBlockLeft(px, panelW, len(mergeTargets), hasSplitButton)
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
	if armyCanRenderReplenishment(gs, a) {
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
	if grainNeed := gs.EffectiveArmyGrainUpkeep(a); grainNeed > 0 {
		DrawText(screen, "Tahıl ihtiyacı: "+itoa(grainNeed)+" / tur",
			float64(px)+float64(armyPanelPadX), float64(py)+float64(armyPanelInfoY+14), FaceSmall, color.RGBA{205, 185, 120, 235})
	}
	if selectedCount > 0 {
		selectedText := "Bölünecek: " + itoa(selectedCount)
		DrawText(screen, selectedText, float64(layout.gridX), float64(py)+float64(armyPanelInfoY), FaceSmall, color.RGBA{255, 205, 75, 245})
	}

	// Aksiyon butonları — BÖL ve BİRLEŞTİR
	canCommand := a.OwnerID == string(gs.PlayerFactionID)
	if canCommand && (canSplit || hasMerge) {
		drawArmyActionButton(screen, px, py, panelW, "✂ BÖL", canSplit, hasMerge, true)
	}
	if canCommand && hasMerge {
		for index, targetID := range mergeTargets {
			other := gs.Armies[targetID]
			if other == nil {
				continue
			}
			canMerge := len(other.Units) < army.MaxArmySize
			label := "->" + itoa(len(other.Units))
			bx, by, bw, bh := mergeButtonRectAt(px, py, panelW, index, len(mergeTargets), canSplit)
			drawArmyPanelButton(screen, bx, by, bw, bh, label, canMerge)
		}
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

		unitIndex := armyPanelUnitIndex(a.Units, gs.UnitTypes, i)
		if unitIndex < 0 {
			// Boş slot — silik çerçeve
			vector.FillRect(screen, cx, cy, cardW, cardH, color.RGBA{14, 12, 8, 120}, false)
			vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, color.RGBA{45, 38, 24, 130}, false)
			// Ortada soluk artı/boş işareti
			DrawTextCentered(screen, "+", float64(cx)+float64(cardW)/2, float64(cy)+float64(cardH)/2-10,
				FaceLarge, color.RGBA{40, 35, 22, 100})
			continue
		}

		u := a.Units[unitIndex]
		utype := gs.UnitTypes[u.TypeID]
		hpPct := u.HPPercent()
		isReplenishing := armyCanRenderReplenishment(gs, a) && u.CurrentHP < army.MaxUnitHP

		// Kart arka planı sabit beyaz.
		cardBg := color.RGBA{255, 255, 255, 245}
		cardBorderCol := color.RGBA{160, 160, 160, 225}
		isSelected := selectedUnits[unitIndex]
		vector.FillRect(screen, cx, cy, cardW, cardH, cardBg, false)
		vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, cardBorderCol, false)

		// Sprite 210x360 oranını korur; hedef alanın tamamını doldurur.
		if sprite := unitSpriteForFaction(gs, a.OwnerID, u.TypeID); sprite != nil && utype != nil {
			drawUnitSpriteCard(screen, sprite, cx, cy, cardW, [3]float32{1, 1, 1})
		} else if utype == nil {
			DrawTextCentered(screen, "?", float64(cx)+float64(cardW)/2, float64(cy)+20, FaceLarge, ColorGray)
		}
		drawUnitCardFooter(screen, cx, cy, cardW, cardH, unitCardFooterH)
		// Sprite tam kart alanını kaplayabildiği için rozet sprite'tan sonra
		// çizilmelidir; aksi halde zayiatlı birimdeki yeşil artı görünmez.
		if isReplenishing {
			badgeW := float32(18)
			badgeH := float32(12)
			badgeX := cx + cardW - badgeW - 3
			badgeY := cy + 3
			vector.FillRect(screen, badgeX, badgeY, badgeW, badgeH, color.RGBA{70, 150, 84, 235}, false)
			DrawTextCentered(screen, "+", float64(badgeX)+float64(badgeW)/2, float64(badgeY)-1, FaceSmall, color.RGBA{245, 255, 245, 255})
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
			float64(cy)+float64(cardH)-float64(unitCardNameOffset),
			FaceSmall, nameCol)

		// HP çubuğu
		hpY := cy + cardH - hpBarH - 2
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
		if isSelected {
			vector.StrokeRect(screen, cx+1, cy+1, cardW-2, cardH-2, 3, color.RGBA{255, 190, 45, 255}, false)
		}
	}
	drawArmyPowerFooter(screen, layout, a.TotalStrength(gs.UnitTypes), a.TotalDefense(gs.UnitTypes), "Güç", armyTransportFooterText(gs, a))
	drawMerchantRouteFooter(screen, gs, a, layout)
	drawNavalMissionFooter(screen, gs, a, layout)

}

func armyCanRenderReplenishment(gs *state.GameState, a *army.Army) bool {
	return gs != nil && a != nil && gs.CanArmyReplenishIn(a) &&
		!gs.IsArmyDefendingSiegedRegion(a) && a.HasDamagedUnits()
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
	drawEnemyArmyFooterText(screen, layout, "Bu orduya hareket emri verilemez  |  Birim detayları gizli", "Güç: Gizli")
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

// armyPanelStrength birimlerin panelde gösterilecek saldırı/savunma gücünü
// hesaplar. Düşman ordusunda HP bilgisi istihbaratla açılmadığı için bu helper
// kısmi istihbaratta birimleri tam HP kabul eder; böylece gizli hasar bilgisi
// güç metninden sızmaz.
func armyPanelStrength(units []army.Unit, unitTypes map[string]*army.UnitType, useCurrentHP bool) (attack, defense int) {
	for _, unit := range units {
		unitType := unitTypes[unit.TypeID]
		if unitType == nil {
			continue
		}
		hpRatio := 1.0
		if useCurrentHP {
			hpRatio = unit.HPPercent()
		}
		attackPower := unit.EffectiveAttack(unitTypes) + unitType.Morale/10
		if scaled := int(float64(attackPower) * hpRatio); scaled > 0 {
			attack += scaled
		} else {
			attack++
		}
		defensePower := int(float64(unit.EffectiveDefense(unitTypes)) * hpRatio)
		if defensePower > 0 {
			defense += defensePower
		} else {
			defense++
		}
	}
	return attack, defense
}

func scoutedEnemyArmyStrength(gs *state.GameState, a *army.Army, fullIntel bool, revealRatio float64) (attack, defense int) {
	if gs == nil || a == nil {
		return 0, 0
	}
	revealed := scoutedEnemyRevealCount(len(a.Units), fullIntel, revealRatio)
	for displayIndex := 0; displayIndex < revealed; displayIndex++ {
		unitIndex := armyPanelUnitIndex(a.Units, gs.UnitTypes, displayIndex)
		if unitIndex < 0 {
			break
		}
		unitAttack, unitDefense := armyPanelStrength(a.Units[unitIndex:unitIndex+1], gs.UnitTypes, false)
		attack += unitAttack
		defense += unitDefense
	}
	return attack, defense
}

func armyTransportFooterText(gs *state.GameState, a *army.Army) string {
	if gs == nil || a == nil || !a.IsNaval {
		return ""
	}
	capacity := a.TransportCapacity(gs.UnitTypes)
	if capacity <= 0 {
		return ""
	}
	return "Taşıma: " + itoa(a.EmbarkedCount()) + "/" + itoa(capacity)
}

func drawArmyPowerFooter(screen *ebiten.Image, layout armyPanelLayout, attack, defense int, label, transportText string) {
	drawArmyPanelFooterBackground(screen, layout)
	powerText := label + ": " + itoa(attack) + " / " + itoa(defense)
	footerY := layout.panelY + layout.panelH - siegeFooterH
	footerX := layout.gridX
	footerW := layout.panelX + layout.panelW - armyPanelPadX - footerX
	powerW := MeasureText(powerText, FaceSmall)
	powerX := float64(footerX+footerW) - float64(armyPanelPadX) - powerW
	if transportText != "" {
		transportW := MeasureText(transportText, FaceSmall)
		transportX := powerX - 18 - transportW
		if transportX >= float64(footerX+armyPanelPadX) {
			DrawText(screen, transportText, transportX, float64(footerY+2), FaceSmall, color.RGBA{205, 185, 140, 230})
		}
	}
	DrawText(screen, powerText, powerX, float64(footerY+2), FaceSmall, color.RGBA{220, 190, 100, 235})
}

func drawArmyPanelFooterBackground(screen *ebiten.Image, layout armyPanelLayout) {
	footerY := layout.panelY + layout.panelH - siegeFooterH
	footerX := layout.gridX
	footerW := layout.panelX + layout.panelW - armyPanelPadX - footerX
	vector.FillRect(screen, footerX, footerY, footerW, siegeFooterH, color.RGBA{28, 18, 6, 190}, false)
	vector.StrokeLine(screen, footerX, footerY, footerX+footerW, footerY, 1, panelBorder, false)
}

func drawArmyPanelFooterRight(screen *ebiten.Image, layout armyPanelLayout, text string, textColor color.Color) {
	footerY := layout.panelY + layout.panelH - siegeFooterH
	footerX := layout.gridX
	footerW := layout.panelX + layout.panelW - armyPanelPadX - footerX
	textW := MeasureText(text, FaceSmall)
	DrawText(screen, text, float64(footerX+footerW)-float64(textW)-float64(armyPanelPadX), float64(footerY+2), FaceSmall, textColor)
}

func drawEnemyArmyFooterText(screen *ebiten.Image, layout armyPanelLayout, leftText, rightText string) {
	footerY := layout.panelY + layout.panelH - siegeFooterH
	leftX := layout.panelX + armyPanelPadX
	leftW := layout.gridX - leftX - armyPanelPadX
	rightW := MeasureText(rightText, FaceSmall)
	maxLeftW := float64(leftW) - rightW - 16
	if maxLeftW < 0 {
		maxLeftW = 0
	}
	DrawText(screen, trimTextToWidth(leftText, FaceSmall, maxLeftW), float64(leftX), float64(footerY+2), FaceSmall, color.RGBA{180, 100, 90, 210})
	drawArmyPanelFooterRight(screen, layout, rightText, color.RGBA{220, 190, 100, 235})
}

func drawScoutedEnemyArmyDetailPanel(screen *ebiten.Image, gs *state.GameState, a *army.Army, fullIntel bool, revealRatio float64) {
	ensureArmySprites()

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

		unitIndex := armyPanelUnitIndex(a.Units, gs.UnitTypes, i)
		if unitIndex < 0 {
			vector.FillRect(screen, cx, cy, cardW, cardH, color.RGBA{14, 12, 8, 90}, false)
			vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, color.RGBA{45, 38, 24, 95}, false)
			continue
		}
		if i >= revealed {
			drawUnknownEnemyUnitCard(screen, cx, cy)
			continue
		}
		drawScoutedEnemyUnitCard(screen, gs, a.OwnerID, a.Units[unitIndex], cx, cy)
	}

	vector.FillRect(screen, px, py+panelH-siegeFooterH, panelW, siegeFooterH, color.RGBA{28, 18, 6, 190}, false)
	vector.StrokeLine(screen, px, py+panelH-siegeFooterH, px+panelW, py+panelH-siegeFooterH, 1, panelBorder, false)
	footer := "Bu orduya hareket emri verilemez"
	if fullIntel {
		footer += "  |  Tam istihbarat"
	} else {
		footer += "  |  Kısmi istihbarat"
	}
	attack, defense := scoutedEnemyArmyStrength(gs, a, fullIntel, revealRatio)
	drawEnemyArmyFooterText(screen, layout, footer, "Görünen güç: "+itoa(attack)+" / "+itoa(defense))
}

// armyPanelUnitIndex, oyun state'indeki birim sırasını değiştirmeden paneldeki
// gösterim sırasındaki birimin gerçek index'ini döndürür. Aynı kategori içindeki
// sıra korunur; kartlar piyade, süvari, kuşatma ve ardından diğer kategoriler
// şeklinde gruplanır.
func armyPanelUnitIndex(units []army.Unit, unitTypes map[string]*army.UnitType, displayIndex int) int {
	if displayIndex < 0 || displayIndex >= len(units) {
		return -1
	}

	seen := 0
	for categoryRank := 0; categoryRank <= 3; categoryRank++ {
		for index := range units {
			if armyPanelUnitCategoryRank(unitTypes, units[index].TypeID) != categoryRank {
				continue
			}
			if seen == displayIndex {
				return index
			}
			seen++
		}
	}
	return -1
}

func armyPanelUnitCategoryRank(unitTypes map[string]*army.UnitType, typeID string) int {
	unitType := unitTypes[typeID]
	if unitType == nil {
		return 3
	}
	switch unitType.Category {
	case army.CategoryInfantry:
		return 0
	case army.CategoryCavalry:
		return 1
	case army.CategorySiege:
		return 2
	default:
		return 3
	}
}

func drawUnknownEnemyUnitCard(screen *ebiten.Image, cx, cy float32) {
	vector.FillRect(screen, cx, cy, cardW, cardH, color.RGBA{24, 20, 16, 220}, false)
	vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, color.RGBA{95, 75, 45, 210}, false)
	DrawTextCentered(screen, "?", float64(cx)+float64(cardW)/2, float64(cy)+20, FaceLarge, color.RGBA{210, 180, 90, 230})
	DrawTextCentered(screen, "Gizli", float64(cx)+float64(cardW)/2, float64(cy)+float64(cardH)-float64(unitCardNameOffset), FaceSmall, color.RGBA{150, 130, 90, 220})
	drawBar(screen, cx+1, cy+cardH-hpBarH-2, cardW-2, hpBarH-1, 1, color.RGBA{80, 70, 55, 180})
}

func drawScoutedEnemyUnitCard(screen *ebiten.Image, gs *state.GameState, ownerID string, u army.Unit, cx, cy float32) {
	utype := gs.UnitTypes[u.TypeID]
	vector.FillRect(screen, cx, cy, cardW, cardH, color.RGBA{255, 255, 255, 245}, false)
	vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, color.RGBA{160, 160, 160, 225}, false)

	if sprite := unitSpriteForFaction(gs, ownerID, u.TypeID); sprite != nil && utype != nil {
		drawUnitSpriteCard(screen, sprite, cx, cy, cardW, [3]float32{0.85, 0.85, 0.85})
	}
	drawUnitCardFooter(screen, cx, cy, cardW, cardH, unitCardFooterH)

	unitName := u.TypeID
	if utype != nil {
		unitName = utype.NameTR
	}
	DrawTextCentered(screen, shortUnitName(unitName, 14),
		float64(cx)+float64(cardW)/2,
		float64(cy)+float64(cardH)-float64(unitCardNameOffset),
		FaceSmall, color.RGBA{25, 25, 25, 235})
	drawBar(screen, cx+1, cy+cardH-hpBarH-2, cardW-2, hpBarH-1, 1, color.RGBA{120, 110, 85, 210})
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
// BÖL her zaman aksiyon grubunun en sağındaki butondur.
func splitButtonRect(px, py, panelW float32, _ bool) (bx, by, bw, bh float32) {
	bw, bh = actionBtnW, actionBtnH
	by = py + armyPanelBtnY
	bx = px + panelW - armyPanelPadX - actionBtnW - actionBtnRightInset
	return
}

func splitButtonBlockLeft(px, panelW float32, mergeCount int, hasSplit bool) float32 {
	actionCount := mergeCount
	if hasSplit {
		actionCount++
	}
	if actionCount > 0 {
		return px + panelW - armyPanelPadX - actionBtnW*float32(actionCount) - actionBtnGap*float32(actionCount-1) - actionBtnRightInset
	}
	return px + panelW - armyPanelPadX - actionBtnW - actionBtnRightInset
}

// mergeButtonRect BİRLEŞTİR butonunun piksel dikdörtgenini döner.
// hasSplit true ise BİRLEŞTİR, BÖL butonunun soluna yerleşir.
func mergeButtonRect(px, py, panelW float32, hasSplit bool) (bx, by, bw, bh float32) {
	return mergeButtonRectAt(px, py, panelW, 0, 1, hasSplit)
}

// mergeButtonRectAt birleştirme hedeflerini soldan sağa, BÖL düğümünün
// hemen öncesine yerleştirir. Her hedef kendi dikdörtgenine sahip olduğu için
// hit-test tıklanan hedef ordunun ID'sini koruyabilir.
func mergeButtonRectAt(px, py, panelW float32, index, mergeCount int, hasSplit bool) (bx, by, bw, bh float32) {
	bw, bh = actionBtnW, actionBtnH
	by = py + armyPanelBtnY
	if mergeCount < 1 {
		return
	}
	actionCount := mergeCount
	if hasSplit {
		actionCount++
	}
	groupLeft := px + panelW - armyPanelPadX - actionBtnRightInset - actionBtnW*float32(actionCount) - actionBtnGap*float32(actionCount-1)
	bx = groupLeft + actionBtnW*float32(index) + actionBtnGap*float32(index)
	return
}

func buildSplitArmyButton(gs *state.GameState, aid army.ArmyID, selectedUnitMaps ...map[int]bool) (gameui.Button, bool) {
	if aid == "" {
		return gameui.Button{}, false
	}
	a, ok := gs.Armies[aid]
	if !ok || len(a.Units) < 2 {
		return gameui.Button{}, false
	}
	if len(selectedUnitMaps) > 0 && !splitSelectionCanBeApplied(a, selectedUnitMaps[0]) {
		return gameui.Button{}, false
	}
	layout := armyPanelGeometry()
	hasMerge := FindMergeTarget(gs, aid) != ""
	bx, by, bw, bh := splitButtonRect(layout.panelX, layout.panelY, layout.panelW, hasMerge)
	return gameui.NewButton(float64(bx), float64(by), float64(bw), float64(bh), "✂ BÖL"), true
}

func buildMergeArmyButton(gs *state.GameState, aid army.ArmyID) (gameui.Button, bool) {
	targets := FindMergeTargets(gs, aid)
	if len(targets) == 0 {
		return gameui.Button{}, false
	}
	return buildMergeArmyButtonForTarget(gs, aid, targets[0], 0, len(targets))
}

func buildMergeArmyButtonForTarget(gs *state.GameState, aid, targetID army.ArmyID, index, mergeCount int) (gameui.Button, bool) {
	a := gs.Armies[aid]
	target := gs.Armies[targetID]
	if a == nil || target == nil || mergeCount < 1 {
		return gameui.Button{}, false
	}
	layout := armyPanelGeometry()
	bx, by, bw, bh := mergeButtonRectAt(layout.panelX, layout.panelY, layout.panelW, index, mergeCount, len(a.Units) >= 2)
	label := "->" + itoa(len(target.Units))
	return gameui.NewButton(float64(bx), float64(by), float64(bw), float64(bh), label), true
}

// drawArmyActionButton tek bir aksiyon butonunu çizer.
// isSplit true → sağ buton (BÖL), false → BÖL varsa onun solundaki BİRLEŞTİR.
func drawArmyActionButton(screen *ebiten.Image, px, py, panelW float32, label string, active, hasOtherAction, isSplit bool) {
	var bx, by, bw, bh float32
	if isSplit {
		bx, by, bw, bh = splitButtonRect(px, py, panelW, hasOtherAction)
	} else {
		bx, by, bw, bh = mergeButtonRect(px, py, panelW, hasOtherAction)
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
	return len(FindMergeTargets(gs, aid)) > 0
}

func splitSelectionCanBeApplied(a *army.Army, selected map[int]bool) bool {
	if a == nil || len(a.Units) < 2 {
		return false
	}
	selectedCount := splitSelectedUnitCount(a, selected)
	return selectedCount == 0 || selectedCount < len(a.Units)
}

func splitSelectedUnitCount(a *army.Army, selected map[int]bool) int {
	if a == nil {
		return 0
	}
	selectedCount := 0
	for index, isSelected := range selected {
		if isSelected && index >= 0 && index < len(a.Units) {
			selectedCount++
		}
	}
	return selectedCount
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

// FindMergeTargets aynı bölgede aynı türde (naval/kara) başka dost orduları
// deterministik ID sırasıyla döner.
func FindMergeTargets(gs *state.GameState, aid army.ArmyID) []army.ArmyID {
	if gs == nil {
		return nil
	}
	a := gs.Armies[aid]
	if a == nil {
		return nil
	}
	targets := make([]army.ArmyID, 0)
	for otherID, other := range gs.Armies {
		if otherID == aid || other == nil || other.LocationID() != a.LocationID() ||
			other.OwnerID != a.OwnerID || other.IsNaval != a.IsNaval {
			continue
		}
		targets = append(targets, otherID)
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i] < targets[j] })
	return targets
}

// FindMergeTarget aynı bölgede aynı türde (naval/kara) ilk dost ordu varsa ID'sini döner.
func FindMergeTarget(gs *state.GameState, aid army.ArmyID) army.ArmyID {
	targets := FindMergeTargets(gs, aid)
	if len(targets) == 0 {
		return ""
	}
	return targets[0]
}

// SplitButtonHitTest fare BÖL butonuna denk geliyorsa true döner.
func SplitButtonHitTest(fx, fy float64, gs *state.GameState, aid army.ArmyID, selectedUnitMaps ...map[int]bool) bool {
	btn, ok := buildSplitArmyButton(gs, aid, selectedUnitMaps...)
	return ok && btn.HitTest(fx, fy)
}

// MergeButtonHitTest fare BİRLEŞTİR butonuna denk geliyorsa true döner.
func MergeButtonHitTest(fx, fy float64, gs *state.GameState, aid army.ArmyID) bool {
	_, ok := MergeButtonTargetAt(fx, fy, gs, aid)
	return ok
}

// MergeButtonTargetAt fare konumundaki BİRLEŞTİR düğmesinin hedef ordusunu
// döner. Hedef ordunun birim sayısı buton etiketinde gösterilen sonuçla aynı
// state'ten hesaplanır.
func MergeButtonTargetAt(fx, fy float64, gs *state.GameState, aid army.ArmyID) (army.ArmyID, bool) {
	if gs == nil {
		return "", false
	}
	source := gs.Armies[aid]
	if source == nil || source.OwnerID != string(gs.PlayerFactionID) {
		return "", false
	}
	targets := FindMergeTargets(gs, aid)
	for index, targetID := range targets {
		btn, ok := buildMergeArmyButtonForTarget(gs, aid, targetID, index, len(targets))
		if ok && btn.HitTest(fx, fy) {
			return targetID, true
		}
	}
	return "", false
}

func mergeResultUnitCount(source, target *army.Army) int {
	if source == nil || target == nil {
		return 0
	}
	count := len(source.Units) + len(target.Units)
	if count > army.MaxArmySize {
		return army.MaxArmySize
	}
	return count
}

func CommanderPortraitHitTest(fx, fy float64, gs *state.GameState, aid army.ArmyID) bool {
	rect, ok := commanderPortraitHitRect(gs, aid)
	return ok && rect.Hit(fx, fy)
}

func ArmyPanelBoundsHit(fx, fy float64, gs *state.GameState, aid army.ArmyID) bool {
	rect, ok := armyDetailPanelRect(gs, aid)
	return ok && rect.Hit(fx, fy)
}

// ArmyPanelUnitHoverID, oyuncu ordusunun panelde görünen birim kartının
// gerçek birim türünü döndürür. Kart sırası çizimdeki armyPanelUnitIndex ile
// aynı tutulur; böylece hover, state içindeki fiziksel sıradan bağımsız olarak
// ekranda görünen karta bağlanır.
func ArmyPanelUnitHoverID(mx, my float64, gs *state.GameState, aid army.ArmyID) string {
	unit, _, ok := armyPanelUnitHover(mx, my, gs, aid)
	if !ok {
		return ""
	}
	return unit.TypeID
}

// armyPanelUnitHover, hover edilen kartın gerçek Unit örneğini ve aynı tipin
// seçili ordudaki adetini döndürür. Popup'ın mevcut can ve adet bilgisini
// recruit kartındaki UnitType tanımından ayırmak için bu state verisi gerekir.
func armyPanelUnitHover(mx, my float64, gs *state.GameState, aid army.ArmyID) (army.Unit, int, bool) {
	if gs == nil || aid == "" {
		return army.Unit{}, 0, false
	}
	a := gs.Armies[aid]
	if !playerCanSeeArmyDetails(gs, a) {
		return army.Unit{}, 0, false
	}

	layout := armyPanelGeometry()
	for displayIndex := 0; displayIndex < army.MaxArmySize; displayIndex++ {
		unitIndex := armyPanelUnitIndex(a.Units, gs.UnitTypes, displayIndex)
		if unitIndex < 0 {
			continue
		}
		col := displayIndex % maxCols
		row := displayIndex / maxCols
		cx := layout.gridX + float32(col)*(cardW+cardGap)
		cy := layout.gridY + float32(row)*(cardH+cardGap)
		if mx >= float64(cx) && mx <= float64(cx+cardW) &&
			my >= float64(cy) && my <= float64(cy+cardH) {
			unit := a.Units[unitIndex]
			count := 0
			for _, candidate := range a.Units {
				if candidate.TypeID == unit.TypeID {
					count++
				}
			}
			return unit, count, true
		}
	}
	return army.Unit{}, 0, false
}

func ArmyPanelInteractiveHit(fx, fy float64, gs *state.GameState, aid army.ArmyID, selectedUnitMaps ...map[int]bool) bool {
	if buildArmyPanelCloseButton().HitTest(fx, fy) {
		return true
	}
	if SplitButtonHitTest(fx, fy, gs, aid, selectedUnitMaps...) || MergeButtonHitTest(fx, fy, gs, aid) {
		return true
	}
	if merchantRouteButtonHit(fx, fy, gs, aid) {
		return true
	}
	if navalMissionButtonHit(fx, fy, gs, aid) {
		return true
	}
	if CommanderPortraitHitTest(fx, fy, gs, aid) {
		return true
	}
	if _, ok := ArmyPanelUnitHit(fx, fy, gs, aid); ok {
		return true
	}
	return false
}

// ArmyPanelUnitHit panelde tıklanan kartın oyun state'indeki fiziksel index'ini
// döndürür. Kartlar kategoriye göre çizildiği için bu index doğrudan paneldeki
// display index'i değildir.
func ArmyPanelUnitHit(mx, my float64, gs *state.GameState, aid army.ArmyID) (int, bool) {
	return armyPanelUnitIndexAt(mx, my, gs, aid)
}

func armyPanelUnitIndexAt(mx, my float64, gs *state.GameState, aid army.ArmyID) (int, bool) {
	if gs == nil || aid == "" {
		return -1, false
	}
	a := gs.Armies[aid]
	if !playerCanSeeArmyDetails(gs, a) {
		return -1, false
	}
	layout := armyPanelGeometry()
	for displayIndex := 0; displayIndex < army.MaxArmySize; displayIndex++ {
		unitIndex := armyPanelUnitIndex(a.Units, gs.UnitTypes, displayIndex)
		if unitIndex < 0 {
			continue
		}
		col := displayIndex % maxCols
		row := displayIndex / maxCols
		cx := layout.gridX + float32(col)*(cardW+cardGap)
		cy := layout.gridY + float32(row)*(cardH+cardGap)
		if mx >= float64(cx) && mx <= float64(cx+cardW) &&
			my >= float64(cy) && my <= float64(cy+cardH) {
			return unitIndex, true
		}
	}
	return -1, false
}
