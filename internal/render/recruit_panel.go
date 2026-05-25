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

var unitDisplayOrder = []string{
	"militia", "infantry", "elite_infantry",
	"light_cavalry", "cavalry", "heavy_cavalry",
	"catapult", "bombard", "cannon",
	"transport", "merchant_ship", "warship",
}

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

const (
	recruitMaxCards   = 20
	recruitCardW      = float32(78)
	recruitCardH      = float32(112)
	recruitCardGap    = float32(6)
	recruitPanelPad   = float32(14)
	recruitHeaderH    = float32(52)
	recruitSectionH   = float32(156)
	recruitSectionGap = float32(10)
	recruitPanelH     = int(recruitHeaderH + recruitSectionH + recruitSectionGap + recruitSectionH + 18)
	recruitBottomGap  = float32(150)
)

func recruitPanelX(slots int) float32 {
	pw := recruitPanelW(slots)
	x := (float32(ScreenWidth) - pw) * 0.5
	if x < 8 {
		return 8
	}
	return x
}
func recruitPanelY() float32 {
	return float32(ScreenHeight) - float32(recruitPanelH) - recruitBottomGap
}
func recruitPanelW(slots int) float32 {
	if slots < 1 {
		slots = 1
	}
	if slots > recruitMaxCards {
		slots = recruitMaxCards
	}
	w := recruitPanelPad*2 + recruitCardW*float32(slots) + recruitCardGap*float32(slots-1)
	maxW := float32(ScreenWidth) - 16
	if w > maxW {
		w = maxW
	}
	return w
}

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

func RecruitPanelVisible(gs *state.GameState, rid world.RegionID) bool {
	if rid == "" {
		return false
	}
	r, ok := gs.Regions[rid]
	return ok && !r.IsSea && !r.IsLocked && r.OwnerID == string(gs.PlayerFactionID)
}

func RecruitPanelHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) string {
	if !RecruitPanelVisible(gs, rid) {
		return ""
	}
	return recruitUnitCardHitTest(mx, my, gs, rid)
}

func RecruitPanelActionHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) RecruitPanelAction {
	if orderID := recruitQueueCancelHitTest(mx, my, gs, rid); orderID != "" {
		return RecruitPanelAction{Kind: RecruitPanelActionCancel, OrderID: orderID}
	}
	if uid := recruitUnitCardHitTest(mx, my, gs, rid); uid != "" {
		return RecruitPanelAction{Kind: RecruitPanelActionRecruit, UnitID: uid}
	}
	return RecruitPanelAction{}
}

func RecruitPanelBoundsHit(mx, my float64, gs *state.GameState, rid world.RegionID) bool {
	if !RecruitPanelVisible(gs, rid) {
		return false
	}
	slots := recruitPanelSlots(gs, rid)
	px := float64(recruitPanelX(slots))
	py := float64(recruitPanelY())
	pw := float64(recruitPanelW(slots))
	ph := float64(recruitPanelH)
	return mx >= px && mx <= px+pw && my >= py && my <= py+ph
}

func DrawRecruitPanel(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, selectedUnitID string, selectedQty int) {
	if !RecruitPanelVisible(gs, rid) {
		return
	}
	region := gs.Regions[rid]
	ensureArmySheet()
	slots := recruitPanelSlots(gs, rid)

	px := recruitPanelX(slots)
	py := recruitPanelY()
	pw := recruitPanelW(slots)
	ph := float32(recruitPanelH)

	vector.FillRect(screen, px, py, pw, ph, panelBg, false)
	drawPanelBorder(screen, px, py, pw, ph)
	vector.FillRect(screen, px, py, pw, 3, panelBorder, false)

	titleW := MeasureText("BIRIM OLUSTUR", FaceSmall)
	DrawText(screen, "BIRIM OLUSTUR", float64(px)+float64(pw)/2-titleW/2, float64(py)+8, FaceSmall, color.RGBA{200, 170, 90, 220})
	limit := recruitRegionProductionLimit(region)
	queuedTotal := queuedUnitTotal(gs, rid)
	infoStr := fmt.Sprintf("Tur limiti: %d  |  Sirada: %d", limit, queuedTotal)
	infoW := MeasureText(infoStr, FaceSmall)
	DrawText(screen, infoStr, float64(px)+float64(pw)/2-infoW/2, float64(py)+24, FaceSmall, color.RGBA{145, 132, 98, 220})
	sepY := py + recruitHeaderH - 2
	vector.StrokeLine(screen, px+12, sepY, px+pw-12, sepY, 1, panelBorder, false)

	hasBarracks, hasPort := false, false
	for _, bid := range region.Buildings {
		switch bid {
		case "barracks":
			hasBarracks = true
		case "port":
			hasPort = true
		}
	}

	display := visibleUnitIDs(gs, region)
	topY := py + recruitHeaderH + 4
	cardW, cardH, gap := recruitCardMetrics(slots, pw)
	x := px + recruitPanelPad
	maxTop := len(display)
	if maxTop > slots {
		maxTop = slots
	}
	for i := 0; i < maxTop; i++ {
		uid := display[i]
		drawRecruitCard(screen, gs, rid, uid, hasBarracks, hasPort, x, topY, cardW, cardH)
		x += cardW + gap
	}

	queueY := topY + recruitSectionH + recruitSectionGap
	drawRecruitQueueSection(screen, gs, rid, slots, px, queueY, pw, recruitSectionH)
}

func recruitCardMetrics(slots int, panelW float32) (cardW, cardH, gap float32) {
	if slots < 1 {
		slots = 1
	}
	gap = recruitCardGap
	avail := panelW - recruitPanelPad*2 - gap*float32(slots-1)
	cardW = avail / float32(slots)
	if cardW > recruitCardW {
		cardW = recruitCardW
	}
	if cardW < 40 {
		cardW = 40
	}
	cardH = recruitCardH
	return cardW, cardH, gap
}

func drawRecruitCard(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, uid string, hasBarracks, hasPort bool, sx, sy, cardW, cardH float32) {
	utype := gs.UnitTypes[uid]
	if utype == nil {
		return
	}
	var needsBuilding bool
	switch utype.RequiredBldg {
	case "barracks":
		needsBuilding = !hasBarracks
	case "port":
		needsBuilding = !hasPort
	}
	ff := gs.Factions[gs.PlayerFactionID]
	needsTech := utype.RequiredTech != "" && (ff == nil || !ff.Research.Completed[utype.RequiredTech])
	canAfford := ff != nil && ff.Gold >= utype.GoldCost
	fullyAvail := !needsBuilding && !needsTech
	queued, firstTurn := queuedUnitInfo(gs, rid, uid)

	slotBg := color.RGBA{20, 16, 12, 200}
	borderCol := color.RGBA{55, 45, 30, 200}
	if fullyAvail {
		slotBg = color.RGBA{38, 30, 15, 235}
		borderCol = panelBorder
	}
	vector.FillRect(screen, sx, sy, cardW, cardH, slotBg, false)
	vector.StrokeRect(screen, sx, sy, cardW, cardH, 1, borderCol, false)

	spriteH := float32(76)
	if armySheet != nil {
		r := unitSpriteRect(uid, armySheet)
		if !r.Empty() {
			sub := armySheet.SubImage(r).(*ebiten.Image)
			op := &ebiten.DrawImageOptions{}
			fitW := float64(cardW - 8)
			fitH := float64(spriteH - 6)
			scale := fitW / float64(r.Dx())
			if hScale := fitH / float64(r.Dy()); hScale < scale {
				scale = hScale
			}
			drawW := float64(r.Dx()) * scale
			drawH := float64(r.Dy()) * scale
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(sx)+float64(cardW)/2-drawW/2, float64(sy)+float64(spriteH)/2-drawH/2)
			switch {
			case needsBuilding:
				op.ColorScale.Scale(0.25, 0.25, 0.25, 1.0)
			case needsTech:
				op.ColorScale.Scale(0.45, 0.45, 0.45, 1.0)
			case !canAfford:
				op.ColorScale.Scale(0.65, 0.45, 0.45, 1.0)
			}
			screen.DrawImage(sub, op)
			turnsShown := utype.TurnsRequired
			if queued > 0 && firstTurn > 0 {
				turnsShown = firstTurn
			}
			badge := itoa(turnsShown) + " Tur"
			bx, by := sx+cardW-52, sy+4
			vector.FillRect(screen, bx, by, 42, 13, color.RGBA{18, 16, 12, 230}, false)
			vector.StrokeRect(screen, bx, by, 42, 13, 1, color.RGBA{120, 98, 56, 220}, false)
			DrawTextCentered(screen, badge, float64(bx)+21, float64(by)+2, FaceSmall, color.RGBA{220, 195, 120, 235})
		}
	}

	nameCol := ColorGold
	if !fullyAvail {
		nameCol = color.RGBA{80, 70, 55, 190}
	}
	DrawTextCentered(screen, shortUnitName(utype.NameTR, 14), float64(sx)+float64(cardW)/2, float64(sy)+80, FaceSmall, nameCol)
	cost := itoa(utype.GoldCost) + " G  " + itoa(utype.TurnsRequired) + "T"
	costCol := color.RGBA{180, 160, 60, 220}
	if !fullyAvail {
		costCol = color.RGBA{70, 62, 48, 180}
	} else if !canAfford {
		costCol = ColorRed
	}
	DrawTextCentered(screen, cost, float64(sx)+float64(cardW)/2, float64(sy)+96, FaceSmall, costCol)
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

func recruitUnitCardHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) string {
	if !RecruitPanelVisible(gs, rid) {
		return ""
	}
	region := gs.Regions[rid]
	display := visibleUnitIDs(gs, region)
	if len(display) == 0 {
		return ""
	}
	py := recruitPanelY()
	slots := recruitPanelSlots(gs, rid)
	px := recruitPanelX(slots)
	topY := py + recruitHeaderH + 4
	pw := recruitPanelW(slots)
	cardW, cardH, gap := recruitCardMetrics(slots, pw)
	x := px + recruitPanelPad
	maxTop := len(display)
	if maxTop > slots {
		maxTop = slots
	}
	for i := 0; i < maxTop; i++ {
		uid := display[i]
		if mx >= float64(x) && mx <= float64(x+cardW) && my >= float64(topY) && my <= float64(topY+cardH) {
			return uid
		}
		x += cardW + gap
	}
	return ""
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

func drawRecruitQueueSection(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, slots int, x, y, w, h float32) {
	mx, my := ebiten.CursorPosition()
	fmx, fmy := float64(mx), float64(my)
	vector.FillRect(screen, x+8, y, w-16, h, color.RGBA{14, 12, 10, 220}, false)
	vector.StrokeRect(screen, x+8, y, w-16, h, 1, color.RGBA{88, 72, 44, 220}, false)
	DrawText(screen, "ORDU + EGITIM SIRASI", float64(x)+16, float64(y)+6, FaceSmall, color.RGBA{190, 165, 100, 230})
	items := recruitQueueItems(gs, rid)
	cardW, cardH, gap := recruitCardMetrics(slots, w)
	startX := x + recruitPanelPad
	cy := y + 26
	maxItems := len(items)
	if maxItems > slots {
		maxItems = slots
	}
	for i := 0; i < maxItems; i++ {
		it := items[i]
		if startX+cardW > x+w-recruitPanelPad {
			break
		}
		vector.FillRect(screen, startX, cy, cardW, cardH, color.RGBA{24, 21, 16, 235}, false)
		vector.StrokeRect(screen, startX, cy, cardW, cardH, 1, color.RGBA{118, 97, 58, 225}, false)
		if armySheet != nil {
			r := unitSpriteRect(it.uid, armySheet)
			if !r.Empty() {
				sub := armySheet.SubImage(r).(*ebiten.Image)
				op := &ebiten.DrawImageOptions{}
				fitW := float64(cardW - 6)
				fitH := float64(76)
				scale := fitW / float64(r.Dx())
				if hScale := fitH / float64(r.Dy()); hScale < scale {
					scale = hScale
				}
				drawW := float64(r.Dx()) * scale
				drawH := float64(r.Dy()) * scale
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(float64(startX)+float64(cardW)/2-drawW/2, float64(cy)+8+fitH/2-drawH/2)
				if it.queued {
					op.ColorScale.Scale(0.82, 0.82, 0.82, 1.0)
				}
				screen.DrawImage(sub, op)
			}
		}
		label := "x" + itoa(it.count)
		if it.queued {
			label = "+" + itoa(it.turns) + "T"
			bx, by, bw, bh := startX+cardW-19, cy+2, float32(17), float32(17)
			hovered := fmx >= float64(bx) && fmx <= float64(bx+bw) && fmy >= float64(by) && fmy <= float64(by+bh)
			drawQueueCancelButton(screen, bx, by, bw, bh, hovered)
		}
		DrawTextCentered(screen, label, float64(startX)+float64(cardW)/2, float64(cy)+98, FaceSmall, color.RGBA{220, 195, 120, 235})
		startX += cardW + gap
	}
}

func drawQueueCancelButton(screen *ebiten.Image, x, y, w, h float32, hovered bool) {
	bg := color.RGBA{70, 26, 22, 235}
	border := color.RGBA{170, 88, 76, 235}
	txt := color.RGBA{255, 220, 210, 240}
	if hovered {
		bg = color.RGBA{128, 40, 30, 245}
		border = color.RGBA{240, 140, 120, 245}
		txt = color.RGBA{255, 245, 235, 255}
	}
	vector.FillRect(screen, x, y, w, h, bg, false)
	vector.StrokeRect(screen, x, y, w, h, 1, border, false)
	DrawTextCentered(screen, "X", float64(x)+float64(w)/2, float64(y)+2, FaceSmall, txt)
}

func recruitQueueCancelHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) string {
	if !RecruitPanelVisible(gs, rid) {
		return ""
	}
	py := recruitPanelY()
	slots := recruitPanelSlots(gs, rid)
	px := recruitPanelX(slots)
	pw := recruitPanelW(slots)
	queueY := py + recruitHeaderH + recruitSectionH + recruitSectionGap
	items := recruitQueueItems(gs, rid)
	cardW, _, gap := recruitCardMetrics(slots, pw)
	x := px + recruitPanelPad
	maxItems := len(items)
	if maxItems > slots {
		maxItems = slots
	}
	for i := 0; i < maxItems; i++ {
		it := items[i]
		if x+cardW > px+pw-recruitPanelPad {
			break
		}
		if it.queued && it.orderID != "" {
			bx, by, bw, bh := x+cardW-19, queueY+26+2, float32(17), float32(17)
			if mx >= float64(bx) && mx <= float64(bx+bw) && my >= float64(by) && my <= float64(by+bh) {
				return it.orderID
			}
		}
		x += cardW + gap
	}
	return ""
}

func recruitPanelSlots(gs *state.GameState, rid world.RegionID) int {
	region := gs.Regions[rid]
	displayCount := len(visibleUnitIDs(gs, region))
	queueCount := len(recruitQueueItems(gs, rid))
	slots := displayCount
	if queueCount > slots {
		slots = queueCount
	}
	if slots < 1 {
		slots = 1
	}
	if slots > recruitMaxCards {
		slots = recruitMaxCards
	}
	return slots
}

func shortUnitName(name string, maxRunes int) string {
	r := []rune(name)
	if len(r) <= maxRunes {
		return name
	}
	if maxRunes < 2 {
		return string(r[:maxRunes])
	}
	return string(r[:maxRunes-1]) + "."
}
