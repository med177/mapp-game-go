package render

import (
	"fmt"
	"image"
	"image/color"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Sprite sheet yükleyici ─────────────────────────────────────────────

var (
	armySheet       *ebiten.Image
	armySheetLoaded bool
)

func ensureArmySheet() {
	if armySheetLoaded {
		return
	}
	armySheetLoaded = true
	armySheet = tryLoadImage(ActiveScenarioPath + "/sprites/army.png")
}

// unitDisplayOrder panel slotlarının sırasını belirler (3 sütun × 4 satır).
var unitDisplayOrder = []string{
	"militia", "infantry", "elite_infantry",
	"light_cavalry", "cavalry", "heavy_cavalry",
	"catapult", "bombard", "cannon",
	"transport", "merchant_ship", "warship",
}

// unitSpriteLoc sprite sheet içindeki hücre konumunu tanımlar.
type unitSpriteLoc struct {
	row, col int
}

var unitSpriteLocs = map[string]unitSpriteLoc{
	"militia":        {0, 0},
	"infantry":       {0, 1},
	"elite_infantry": {0, 2},
	"light_cavalry":  {1, 0},
	"cavalry":        {1, 1},
	"heavy_cavalry":  {1, 2},
	"catapult":       {2, 0},
	"bombard":        {2, 1},
	"cannon":         {2, 2},
	"transport":      {3, 0},
	"merchant_ship":  {3, 1},
	"warship":        {3, 2},
}

// unitSpriteRect görüntü boyutuna göre sprite koordinatlarını döner.
// Görüntü 4 satır × 3 sütun eşit hücrelerden oluşur.
func unitSpriteRect(id string, sheet *ebiten.Image) image.Rectangle {
	loc, ok := unitSpriteLocs[id]
	if !ok {
		return image.Rectangle{}
	}
	cellW := sheet.Bounds().Dx() / 3
	cellH := sheet.Bounds().Dy() / 4
	x0 := loc.col * cellW
	y0 := loc.row * cellH
	return image.Rect(x0, y0, x0+cellW, y0+cellH)
}

// ── Panel layout sabitleri ─────────────────────────────────────────────

const (
	recruitPanelW  = infoPanelW + 170
	recruitPanelH  = infoPanelH + 230
	recruitHeaderH = float32(52)
	recruitGridH   = float32(360)
)

func recruitPanelX() float32 { return infoPanelX() + infoPanelW + 5 }
func recruitPanelY() float32 { return float32(ScreenHeight) - recruitPanelH }

type RecruitPanelActionKind int

const (
	RecruitPanelActionNone RecruitPanelActionKind = iota
	RecruitPanelActionRecruit
	RecruitPanelActionIncrease
	RecruitPanelActionDecrease
	RecruitPanelActionCancel
)

type RecruitPanelAction struct {
	Kind    RecruitPanelActionKind
	UnitID  string
	OrderID string
}

// RecruitPanelVisible oyuncunun kendi bölgesi seçiliyken true döner.
func RecruitPanelVisible(gs *state.GameState, rid world.RegionID) bool {
	if rid == "" {
		return false
	}
	r, ok := gs.Regions[rid]
	return ok && !r.IsSea && !r.IsLocked && r.OwnerID == string(gs.PlayerFactionID)
}

// RecruitPanelHitTest fare koordinatına denk gelen unit ID'sini döner; boş = tıklama yok.
func RecruitPanelHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) string {
	if !RecruitPanelVisible(gs, rid) {
		return ""
	}
	if !RecruitPanelBoundsHit(mx, my, gs, rid) {
		return ""
	}
	px := float64(recruitPanelX())
	py := float64(recruitPanelY())
	pw := float64(recruitPanelW)
	_ = float64(recruitPanelH)

	pad := panelPad
	availW := pw - pad*2
	slotW := availW / 3
	slotH := (float64(recruitGridH) - pad) / 4

	relX := mx - px - pad
	relY := my - py - float64(recruitHeaderH)
	if relX < 0 || relY < 0 {
		return ""
	}

	col := int(relX / slotW)
	row := int(relY / slotH)
	if col < 0 || col > 2 || row < 0 || row > 3 {
		return ""
	}
	idx := row*3 + col
	display := visibleUnitIDs(gs, gs.Regions[rid])
	if idx >= len(display) {
		return ""
	}
	return display[idx]
}

// RecruitPanelActionHitTest panel içinde tıklanan eylemi döner.
func RecruitPanelActionHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) RecruitPanelAction {
	if orderID := recruitQueueCancelHitTest(mx, my, gs, rid); orderID != "" {
		return RecruitPanelAction{Kind: RecruitPanelActionCancel, OrderID: orderID}
	}
	uid := RecruitPanelHitTest(mx, my, gs, rid)
	if uid == "" {
		return RecruitPanelAction{}
	}
	if recruitPanelStepButtonHit(mx, my, gs, rid, uid, true) {
		return RecruitPanelAction{Kind: RecruitPanelActionIncrease, UnitID: uid}
	}
	if recruitPanelStepButtonHit(mx, my, gs, rid, uid, false) {
		return RecruitPanelAction{Kind: RecruitPanelActionDecrease, UnitID: uid}
	}
	return RecruitPanelAction{Kind: RecruitPanelActionRecruit, UnitID: uid}
}

func RecruitPanelBoundsHit(mx, my float64, gs *state.GameState, rid world.RegionID) bool {
	if !RecruitPanelVisible(gs, rid) {
		return false
	}
	px := float64(recruitPanelX())
	py := float64(recruitPanelY())
	pw := float64(recruitPanelW)
	ph := float64(recruitPanelH)
	return mx >= px && mx <= px+pw && my >= py && my <= py+ph
}

// DrawRecruitPanel birim seçim ızgarasını bölge panelinin sağına çizer.
func DrawRecruitPanel(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, selectedUnitID string, selectedQty int) {
	if !RecruitPanelVisible(gs, rid) {
		return
	}
	region := gs.Regions[rid]
	f := gs.Factions[gs.PlayerFactionID]

	ensureArmySheet()

	px := recruitPanelX()
	py := recruitPanelY()
	pw := recruitPanelW
	ph := recruitPanelH

	// Arka plan ve çerçeve
	vector.FillRect(screen, px, py, pw, ph, panelBg, false)
	drawPanelBorder(screen, px, py, pw, ph)
	vector.FillRect(screen, px, py, pw, 3, panelBorder, false)

	// Başlık
	titleW := MeasureText("BİRİM OLUŞTUR", FaceSmall)
	DrawText(screen, "BİRİM OLUŞTUR",
		float64(px)+float64(pw)/2-titleW/2, float64(py)+8,
		FaceSmall, color.RGBA{200, 170, 90, 220})
	limit := recruitRegionProductionLimit(region)
	queuedTotal := queuedUnitTotal(gs, rid)
	infoStr := fmt.Sprintf("Tur limiti: %d  |  Sirada: %d", limit, queuedTotal)
	infoW := MeasureText(infoStr, FaceSmall)
	DrawText(screen, infoStr, float64(px)+float64(pw)/2-infoW/2, float64(py)+28, FaceSmall, color.RGBA{145, 132, 98, 220})
	sepY := py + 46
	vector.StrokeLine(screen, px+12, sepY, px+float32(pw)-12, sepY, 1, panelBorder, false)

	// Bölge binaları
	hasBarracks, hasPort := false, false
	for _, bid := range region.Buildings {
		switch bid {
		case "barracks":
			hasBarracks = true
		case "port":
			hasPort = true
		}
	}

	// Izgara boyutları
	const cols = 3
	pad := float32(panelPad + 2)
	availW := float32(pw) - pad*2
	slotW := availW / float32(cols)
	slotH := (recruitGridH - pad) / 4
	spriteH := slotH * 0.66
	nameYOff := spriteH + 3
	costYOff := nameYOff + 15
	queueYOff := costYOff + 14
	ctrlBandYOff := slotH - 20

	display := visibleUnitIDs(gs, region)
	for i, uid := range display {
		col := i % cols
		row := i / cols

		sx := px + pad + float32(col)*slotW
		sy := py + recruitHeaderH + float32(row)*slotH
		innerW := slotW - 3

		utype := gs.UnitTypes[uid]
		if utype == nil {
			continue
		}

		// Kullanılabilirlik kontrolü
		var needsBuilding bool
		switch utype.RequiredBldg {
		case "barracks":
			needsBuilding = !hasBarracks
		case "port":
			needsBuilding = !hasPort
		}

		needsTech := utype.RequiredTech != "" &&
			(f == nil || !f.Research.Completed[utype.RequiredTech])

		canAfford := f != nil && f.Gold >= utype.GoldCost
		fullyAvail := !needsBuilding && !needsTech
		queued, firstTurn := queuedUnitInfo(gs, rid, uid)

		// Slot arka plan
		slotBg := color.RGBA{20, 16, 12, 200}
		borderCol := color.RGBA{55, 45, 30, 200}
		if fullyAvail {
			slotBg = color.RGBA{38, 30, 15, 235}
			borderCol = panelBorder
		}
		vector.FillRect(screen, sx, sy, innerW, spriteH, slotBg, false)
		vector.StrokeRect(screen, sx, sy, innerW, spriteH, 1, borderCol, false)

		// Sprite
		if armySheet != nil {
			r := unitSpriteRect(uid, armySheet)
			if !r.Empty() {
				sub := armySheet.SubImage(r).(*ebiten.Image)
				op := &ebiten.DrawImageOptions{}
				fitW := float64(innerW - 6)
				fitH := float64(spriteH - 6)
				scale := fitW / float64(r.Dx())
				if hScale := fitH / float64(r.Dy()); hScale < scale {
					scale = hScale
				}
				drawW := float64(r.Dx()) * scale
				drawH := float64(r.Dy()) * scale
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(
					float64(sx)+float64(innerW)/2-drawW/2,
					float64(sy)+float64(spriteH)/2-drawH/2,
				)
				switch {
				case needsBuilding:
					op.ColorScale.Scale(0.25, 0.25, 0.25, 1.0)
				case needsTech:
					op.ColorScale.Scale(0.45, 0.45, 0.45, 1.0)
				case !canAfford:
					op.ColorScale.Scale(0.65, 0.45, 0.45, 1.0)
				}
				screen.DrawImage(sub, op)

				// Birim görseli üzerinde kalan ilk kuyruk süresi (yoksa taban süre)
				turnsShown := utype.TurnsRequired
				if queued > 0 && firstTurn > 0 {
					turnsShown = firstTurn
				}
				turnBadge := itoa(turnsShown) + " Tur"
				bx := sx + innerW - 50
				by := sy + 4
				vector.FillRect(screen, bx, by, 48, 13, color.RGBA{18, 16, 12, 230}, false)
				vector.StrokeRect(screen, bx, by, 48, 13, 1, color.RGBA{120, 98, 56, 220}, false)
				DrawTextCentered(screen, turnBadge, float64(bx)+24, float64(by)+2, FaceSmall, color.RGBA{220, 195, 120, 235})

				// Kilit ibaresi — teknoloji eksik
				if needsTech && !needsBuilding {
					DrawTextCentered(screen, "KLT",
						float64(sx)+float64(innerW)/2, float64(sy)+float64(spriteH)/2-8,
						FaceMed, color.RGBA{200, 180, 80, 220})
				}
			}
		}

		// Birim adı
		nameX := float64(sx) + float64(innerW)/2
		nameCol := ColorGold
		if !fullyAvail {
			nameCol = color.RGBA{80, 70, 55, 190}
		}
		DrawTextCentered(screen, utype.NameTR, nameX, float64(sy)+float64(nameYOff), FaceSmall, nameCol)

		// Maliyet
		costStr := itoa(utype.GoldCost) + " G  " + itoa(utype.TurnsRequired) + "T"
		costCol := color.RGBA{180, 160, 60, 220}
		if !fullyAvail {
			costCol = color.RGBA{70, 62, 48, 180}
		} else if !canAfford {
			costCol = ColorRed
		}
		DrawTextCentered(screen, costStr, nameX, float64(sy)+float64(costYOff), FaceSmall, costCol)

		// Kuyruk bilgisi (aynı birim için bekleme görünürlüğü)
		if queued > 0 {
			qStr := fmt.Sprintf("Qx%d %dT", queued, firstTurn)
			DrawTextCentered(screen, qStr, nameX, float64(sy)+float64(queueYOff), FaceSmall, color.RGBA{125, 120, 98, 215})
		}

		// Total War benzeri adet kontrolü: - xN +
		if uid == selectedUnitID {
			if selectedQty < 1 {
				selectedQty = 1
			}
			ctrlY := sy + ctrlBandYOff
			vector.FillRect(screen, sx+1, ctrlY, innerW-2, 18, color.RGBA{18, 15, 12, 235}, false)
			vector.StrokeRect(screen, sx+1, ctrlY, innerW-2, 18, 1, color.RGBA{120, 98, 56, 220}, false)
			qtyY := float64(ctrlY) + 2
			mx, my, mw, mh := recruitPanelStepButtonRect(gs, rid, uid, false)
			px, py, pw, ph := recruitPanelStepButtonRect(gs, rid, uid, true)
			drawTinyPanelButton(screen, mx, my, mw, mh, "-", true)
			drawTinyPanelButton(screen, px, py, pw, ph, "+", true)
			DrawTextCentered(screen, "x"+itoa(selectedQty), nameX, qtyY, FaceSmall, ColorGold)
		}
	}

	drawRecruitQueueSection(screen, gs, rid, px, py+recruitHeaderH+recruitGridH+8, pw, ph-(recruitHeaderH+recruitGridH+16))
}

func visibleUnitIDs(gs *state.GameState, region *world.Region) []string {
	showNaval := region != nil && region.IsCoastal(gs.Regions)
	ids := make([]string, 0, len(unitDisplayOrder))
	for _, uid := range unitDisplayOrder {
		utype := gs.UnitTypes[uid]
		if utype == nil {
			continue
		}
		if utype.RequiredBldg == "port" && !showNaval {
			continue
		}
		ids = append(ids, uid)
	}
	return ids
}

func queuedUnitInfo(gs *state.GameState, rid world.RegionID, uid string) (count int, firstTurn int) {
	firstTurn = 0
	for _, order := range gs.ProductionQueue {
		if order.Kind != "unit" || order.RegionID != rid || order.TypeID != uid || order.FactionID != string(gs.PlayerFactionID) {
			continue
		}
		count++
		if firstTurn == 0 || order.TurnsLeft < firstTurn {
			firstTurn = order.TurnsLeft
		}
	}
	return count, firstTurn
}

func queuedUnitTotal(gs *state.GameState, rid world.RegionID) int {
	total := 0
	for _, order := range gs.ProductionQueue {
		if order.Kind == "unit" && order.RegionID == rid && order.FactionID == string(gs.PlayerFactionID) {
			total++
		}
	}
	return total
}

func recruitRegionProductionLimit(region *world.Region) int {
	if region == nil || region.IsSea {
		return 0
	}
	limit := region.Population / 100
	if limit < 1 {
		limit = 1
	}
	for _, bid := range region.Buildings {
		if bid == "barracks" {
			limit++
		}
	}
	return limit
}

func recruitPanelStepButtonRect(gs *state.GameState, rid world.RegionID, uid string, plus bool) (x, y, w, h float32) {
	px := recruitPanelX()
	py := recruitPanelY()
	pw := recruitPanelW

	const cols = 3
	pad := float32(panelPad)
	availW := float32(pw) - pad*2
	slotW := availW / float32(cols)
	slotH := (recruitGridH - pad) / 4

	display := visibleUnitIDs(gs, gs.Regions[rid])
	idx := -1
	for i := range display {
		if display[i] == uid {
			idx = i
			break
		}
	}
	if idx < 0 {
		return 0, 0, 0, 0
	}

	col := idx % cols
	row := idx / cols
	sx := px + pad + float32(col)*slotW
	sy := py + recruitHeaderH + float32(row)*slotH
	innerW := slotW - 3
	btnW, btnH := float32(18), float32(14)
	btnY := sy + slotH - 18
	if plus {
		return sx + innerW - btnW - 2, btnY, btnW, btnH
	}
	return sx + 2, btnY, btnW, btnH
}

func recruitPanelStepButtonHit(mx, my float64, gs *state.GameState, rid world.RegionID, uid string, plus bool) bool {
	x, y, w, h := recruitPanelStepButtonRect(gs, rid, uid, plus)
	return mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h)
}

func drawRecruitQueueSection(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, x, y, w, h float32) {
	vector.FillRect(screen, x+8, y, w-16, h, color.RGBA{14, 12, 10, 220}, false)
	vector.StrokeRect(screen, x+8, y, w-16, h, 1, color.RGBA{88, 72, 44, 220}, false)
	DrawText(screen, "ORDU + EGITIM SIRASI", float64(x)+16, float64(y)+6, FaceSmall, color.RGBA{190, 165, 100, 230})

	type queueItem struct {
		uid     string
		count   int
		queued  bool
		turns   int
		orderID string
	}
	items := make([]queueItem, 0, 32)
	for _, it := range recruitQueueItems(gs, rid) {
		items = append(items, queueItem{uid: it.uid, count: it.count, queued: it.queued, turns: it.turns, orderID: it.orderID})
	}

	cardW, cardH := float32(52), float32(68)
	gap := float32(6)
	startX := x + 14
	startY := y + 20
	maxCols := int((w - 28 + gap) / (cardW + gap))
	if maxCols < 1 {
		maxCols = 1
	}
	for i, it := range items {
		col := i % maxCols
		row := i / maxCols
		cx := startX + float32(col)*(cardW+gap)
		cy := startY + float32(row)*(cardH+gap)
		if cy+cardH > y+h-4 {
			break
		}
		vector.FillRect(screen, cx, cy, cardW, cardH, color.RGBA{24, 21, 16, 235}, false)
		vector.StrokeRect(screen, cx, cy, cardW, cardH, 1, color.RGBA{118, 97, 58, 225}, false)
		if armySheet != nil {
			r := unitSpriteRect(it.uid, armySheet)
			if !r.Empty() {
				sub := armySheet.SubImage(r).(*ebiten.Image)
				op := &ebiten.DrawImageOptions{}
				fitW := float64(cardW - 6)
				fitH := float64(40)
				scale := fitW / float64(r.Dx())
				if hScale := fitH / float64(r.Dy()); hScale < scale {
					scale = hScale
				}
				drawW := float64(r.Dx()) * scale
				drawH := float64(r.Dy()) * scale
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(float64(cx)+float64(cardW)/2-drawW/2, float64(cy)+3+fitH/2-drawH/2)
				if it.queued {
					op.ColorScale.Scale(0.82, 0.82, 0.82, 1.0)
				}
				screen.DrawImage(sub, op)
			}
		}
		label := "x" + itoa(it.count)
		if it.queued {
			label = "+" + itoa(it.turns) + "T"
			drawTinyPanelButton(screen, cx+cardW-13, cy+2, 11, 11, "X", true)
		}
		DrawTextCentered(screen, label, float64(cx)+float64(cardW)/2, float64(cy)+48, FaceSmall, color.RGBA{220, 195, 120, 235})
	}
}

type recruitQueueItem struct {
	uid     string
	count   int
	queued  bool
	turns   int
	orderID string
}

func recruitQueueItems(gs *state.GameState, rid world.RegionID) []recruitQueueItem {
	items := make([]recruitQueueItem, 0, 32)
	existingCounts := make(map[string]int)
	for _, a := range gs.Armies {
		if a.OwnerID != string(gs.PlayerFactionID) || a.RegionID != rid || a.IsNaval {
			continue
		}
		for _, u := range a.Units {
			existingCounts[u.TypeID]++
		}
	}
	for _, uid := range unitDisplayOrder {
		if c := existingCounts[uid]; c > 0 {
			items = append(items, recruitQueueItem{uid: uid, count: c})
		}
	}
	for _, order := range gs.ProductionQueue {
		if order.Kind != "unit" || order.RegionID != rid || order.FactionID != string(gs.PlayerFactionID) {
			continue
		}
		items = append(items, recruitQueueItem{uid: order.TypeID, count: 1, queued: true, turns: order.TurnsLeft, orderID: order.ID})
	}
	return items
}

func recruitQueueCancelHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) string {
	if !RecruitPanelVisible(gs, rid) {
		return ""
	}
	px := recruitPanelX()
	py := recruitPanelY()
	pw := recruitPanelW
	ph := recruitPanelH
	x := px + 8
	y := py + recruitHeaderH + recruitGridH + 8
	w := pw - 16
	h := ph - (recruitHeaderH + recruitGridH + 16)
	cardW, cardH := float32(52), float32(68)
	gap := float32(6)
	startX := x + 14
	startY := y + 20
	maxCols := int((w - 28 + gap) / (cardW + gap))
	if maxCols < 1 {
		maxCols = 1
	}
	items := recruitQueueItems(gs, rid)
	for i, it := range items {
		if !it.queued || it.orderID == "" {
			continue
		}
		col := i % maxCols
		row := i / maxCols
		cx := startX + float32(col)*(cardW+gap)
		cy := startY + float32(row)*(cardH+gap)
		if cy+cardH > y+h-4 {
			break
		}
		bx, by, bw, bh := cx+cardW-13, cy+2, float32(11), float32(11)
		if mx >= float64(bx) && mx <= float64(bx+bw) && my >= float64(by) && my <= float64(by+bh) {
			return it.orderID
		}
	}
	return ""
}
