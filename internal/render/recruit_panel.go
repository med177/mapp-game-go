package render

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	unitSprites       map[unitSpriteKey]*ebiten.Image
	legacyUnitSprites map[string]*ebiten.Image
	legacyArmySheet   *ebiten.Image
	armySpritesLoaded bool
)

const unitSpriteAspectH = float32(360) / float32(210)

const (
	unitCardFooterH           = float32(36)
	unitCardNameOffset        = float32(29)
	unitCardSingleLabelOffset = float32(15)
)

func unitSpriteHeight(width float32) float32 {
	return width * unitSpriteAspectH
}

type armySpriteSet uint8

const (
	armySpriteSetLegacy armySpriteSet = iota
	armySpriteSetEastern
	armySpriteSetWestern
)

type unitSpriteKey struct {
	set    armySpriteSet
	unitID string
}

// Unit türü ile ayrı sprite dosyası arasındaki kanonik eşleştirme.
var unitSpriteAssetNames = map[string]string{
	"militia":        "infantry_light.png",
	"infantry":       "infantry_medium.png",
	"elite_infantry": "infantry_heavy.png",
	"light_cavalry":  "cavalry_light.png",
	"cavalry":        "cavalry_medium.png",
	"heavy_cavalry":  "cavalry_heavy.png",
	"catapult":       "siege_trebuchet.png",
	"bombard":        "siege_mortar.png",
	"cannon":         "siege_cannon.png",
	"transport":      "ship_transport.png",
	"merchant_ship":  "ship_merchant.png",
	"warship":        "ship_war_galley.png",
}

func ensureArmySprites() {
	if armySpritesLoaded {
		return
	}
	armySpritesLoaded = true
	unitSprites = make(map[unitSpriteKey]*ebiten.Image, len(unitSpriteAssetNames)*2)
	legacyUnitSprites = make(map[string]*ebiten.Image, len(unitSpriteAssetNames))
	base := filepath.Join(ActiveScenarioPath, "sprites")
	for _, set := range []struct {
		kind armySpriteSet
		dir  string
	}{
		{kind: armySpriteSetEastern, dir: "eastern_army"},
		{kind: armySpriteSetWestern, dir: "western_army"},
	} {
		for unitID, filename := range unitSpriteAssetNames {
			img := tryLoadImage(filepath.Join(base, set.dir, filename))
			if img != nil {
				unitSprites[unitSpriteKey{set: set.kind, unitID: unitID}] = img
			}
		}
	}

	// Eski senaryoların tek sheet asset'leri için geriye dönük fallback.
	// Yeni tekil sprite klasörleri bulunduğunda eski sheet hiç kullanılmaz.
	if len(unitSprites) == 0 {
		legacyArmySheet = tryLoadImage(filepath.Join(base, "army.png"))
		if legacyArmySheet != nil {
			for unitID := range unitSpriteAssetNames {
				r := legacyUnitSpriteRect(unitID, legacyArmySheet)
				if !r.Empty() {
					legacyUnitSprites[unitID] = legacyArmySheet.SubImage(r).(*ebiten.Image)
				}
			}
		}
	}
}

func armySpriteSetForFaction(gs *state.GameState, ownerID string) armySpriteSet {
	if gs == nil || ownerID == "" {
		return armySpriteSetLegacy
	}
	f := gs.Factions[faction.FactionID(ownerID)]
	if f == nil {
		return armySpriteSetLegacy
	}
	switch f.Religion {
	case religion.Sunni, religion.Shia:
		return armySpriteSetEastern
	default:
		return armySpriteSetWestern
	}
}

func unitSpriteForFaction(gs *state.GameState, ownerID, unitID string) *ebiten.Image {
	ensureArmySprites()
	set := armySpriteSetForFaction(gs, ownerID)
	if img := unitSprites[unitSpriteKey{set: set, unitID: unitID}]; img != nil {
		return img
	}
	return legacyUnitSprites[unitID]
}

// drawUnitSpriteCard birimi kart genişliğine göre tam oranında çizer.
// Görselin tamamı korunur; üst-alt kırpılmaz. Kart üzerindeki metin ve
// kontroller çağıran kod tarafından daha sonra çizilerek sprite'ın üzerine biner.
func drawUnitSpriteCard(screen *ebiten.Image, sprite *ebiten.Image, x, y, width float32, tint [3]float32) bool {
	if screen == nil || sprite == nil || width <= 0 {
		return false
	}
	source := sprite.Bounds()
	if source.Dx() <= 0 || source.Dy() <= 0 {
		return false
	}

	scale := float64(width) / float64(source.Dx())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.Scale(tint[0], tint[1], tint[2], 1.0)
	screen.DrawImage(sprite, op)
	return true
}

// drawUnitCardFooter sprite'ın altındaki etiket alanını opak beyazla kapatır.
// 1px içeri alınarak kart çerçevesinin üstüne taşmaz.
func drawUnitCardFooter(screen *ebiten.Image, x, y, width, height, footerH float32) {
	if screen == nil || width <= 2 || height <= 0 || footerH <= 0 {
		return
	}
	if footerH > height {
		footerH = height
	}
	vector.FillRect(screen, x, y+height-footerH, width, footerH, color.RGBA{255, 255, 255, 255}, false)
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

func legacyUnitSpriteRect(id string, sheet *ebiten.Image) image.Rectangle {
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
	recruitMaxCards       = 20
	recruitCardsPerRow    = 10
	recruitMaxRows        = 2
	recruitQueueMaxOrders = 20
	recruitCardW          = float32(88)
	recruitCardH          = recruitCardW * unitSpriteAspectH
	recruitCardGap        = float32(6)
	recruitPanelPad       = float32(14)
	recruitHeaderH        = float32(52)
	recruitSectionH       = float32(26) + recruitCardH*float32(recruitMaxRows) + recruitCardGap*float32(recruitMaxRows-1)
	recruitSectionGap     = float32(10)
	// İki satırın 210x360 oranlı kartlarını ve aralıklarını kapsar; küsurat yukarı yuvarlanır.
	recruitPanelH    = 748
	recruitBottomGap = float32(150)
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
	slots = recruitCardsPerRow
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
	RecruitPanelActionClose
)

type RecruitPanelAction struct {
	Kind    RecruitPanelActionKind
	UnitID  string
	OrderID string
}

func buildRecruitPanelCloseButton(gs *state.GameState, rid world.RegionID) (gameui.Button, bool) {
	if !RecruitPanelVisible(gs, rid) {
		return gameui.Button{}, false
	}
	slots := recruitPanelSlots()
	px := recruitPanelX(slots)
	py := recruitPanelY()
	pw := recruitPanelW(slots)
	x, y, w, h := recruitPanelCloseRect(px, py, pw)
	btn := gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "").WithIcon(gameui.IconClose)
	btn.IconSize = 12
	return btn, true
}

func buildRecruitUnitCardButtons(gs *state.GameState, rid world.RegionID) []gameui.Button {
	if !RecruitPanelVisible(gs, rid) {
		return nil
	}
	region := gs.Regions[rid]
	display := visibleUnitIDs(gs, region)
	if len(display) == 0 {
		return nil
	}
	py := recruitPanelY()
	slots := recruitPanelSlots()
	px := recruitPanelX(slots)
	topY := py + recruitHeaderH + 4
	pw := recruitPanelW(slots)
	cardW, cardH, gap := recruitCardMetrics(pw)
	maxTop := len(display)
	if maxTop > recruitMaxCards {
		maxTop = recruitMaxCards
	}
	buttons := make([]gameui.Button, 0, maxTop)
	for i := 0; i < maxTop; i++ {
		uid := display[i]
		row := i / recruitCardsPerRow
		col := i % recruitCardsPerRow
		x := px + recruitPanelPad + float32(col)*(cardW+gap)
		y := topY + float32(row)*(cardH+gap)
		buttons = append(buttons, gameui.NewButton(float64(x), float64(y), float64(cardW), float64(cardH), uid))
	}
	return buttons
}

func buildRecruitQueueCancelButtons(gs *state.GameState, rid world.RegionID) map[string]gameui.Button {
	if !RecruitPanelVisible(gs, rid) {
		return nil
	}
	py := recruitPanelY()
	slots := recruitPanelSlots()
	px := recruitPanelX(slots)
	pw := recruitPanelW(slots)
	queueY := py + recruitHeaderH + recruitSectionH + recruitSectionGap
	items := recruitQueueItems(gs, rid)
	cardW, cardH, gap := recruitCardMetrics(pw)
	maxItems := len(items)
	if maxItems > recruitQueueMaxOrders {
		maxItems = recruitQueueMaxOrders
	}
	buttons := make(map[string]gameui.Button, maxItems)
	for i := 0; i < maxItems; i++ {
		it := items[i]
		if !it.queued || it.orderID == "" {
			continue
		}
		row := i / recruitCardsPerRow
		col := i % recruitCardsPerRow
		x := px + recruitPanelPad + float32(col)*(cardW+gap)
		y := queueY + 26 + float32(row)*(cardH+gap)
		bx, by, bw, bh := x+cardW-19, y+2, float32(17), float32(17)
		btn := gameui.NewButton(float64(bx), float64(by), float64(bw), float64(bh), "").WithIcon(gameui.IconClose)
		btn.IconSize = 11
		buttons[it.orderID] = btn
	}
	return buttons
}

func recruitQueueIsFull(gs *state.GameState, rid world.RegionID) bool {
	return len(recruitQueueItems(gs, rid)) >= recruitQueueMaxOrders
}

func RecruitPanelButtonEnabled(gs *state.GameState, rid world.RegionID) bool {
	if !RecruitPanelVisible(gs, rid) || recruitQueueIsFull(gs, rid) {
		return false
	}
	region := gs.Regions[rid]
	ff := gs.Factions[gs.PlayerFactionID]
	if region == nil || ff == nil {
		return false
	}
	barracksLevel, portLevel := 0, 0
	for _, bid := range region.Buildings {
		switch bid {
		case "barracks":
			barracksLevel++
		case "port":
			portLevel++
		}
	}
	for _, uid := range visibleUnitIDs(gs, region) {
		utype := gs.UnitTypes[uid]
		if utype == nil {
			continue
		}
		requiredLevel := utype.RequiredBldgLevel
		if utype.RequiredBldg != "" && requiredLevel <= 0 {
			requiredLevel = 1
		}
		switch utype.RequiredBldg {
		case "barracks":
			if barracksLevel < requiredLevel {
				continue
			}
		case "port":
			if portLevel < requiredLevel {
				continue
			}
		}
		if utype.RequiredTech != "" && !ff.Research.Completed[utype.RequiredTech] {
			continue
		}
		if !unitCost(utype).CanAfford(ff) {
			continue
		}
		return true
	}
	return false
}

func recruitPanelDisabledReason(gs *state.GameState, rid world.RegionID) string {
	if gs == nil {
		return "Bölge Uygun Değil"
	}
	if rid == "" {
		return "Bölge Seç"
	}
	region := gs.Regions[rid]
	ff := gs.Factions[gs.PlayerFactionID]
	if region == nil || ff == nil {
		return "Bölge Uygun Değil"
	}
	if region.IsSea || region.IsLocked || region.OwnerID != string(gs.PlayerFactionID) {
		return "Bölge Uygun Değil"
	}
	if recruitQueueIsFull(gs, rid) {
		return "Sıra Dolu"
	}

	barracksLevel, portLevel := 0, 0
	for _, bid := range region.Buildings {
		switch bid {
		case "barracks":
			barracksLevel++
		case "port":
			portLevel++
		}
	}
	for _, uid := range visibleUnitIDs(gs, region) {
		utype := gs.UnitTypes[uid]
		if utype == nil {
			continue
		}
		requiredLevel := utype.RequiredBldgLevel
		if utype.RequiredBldg != "" && requiredLevel <= 0 {
			requiredLevel = 1
		}
		switch utype.RequiredBldg {
		case "barracks":
			if barracksLevel < requiredLevel {
				return "Kışla Eksik"
			}
		case "port":
			if portLevel < requiredLevel {
				return "Liman Eksik"
			}
		}
		if utype.RequiredTech != "" && !ff.Research.Completed[utype.RequiredTech] {
			return "Teknoloji Eksik"
		}
		if shortage := unitCostShortageReason(ff, unitCost(utype)); shortage != "" {
			return shortage
		}
		return ""
	}
	return "Uygun Birim Yok"
}

func unitCostShortageReason(ff *faction.Faction, cost economy.ResourceCost) string {
	if ff == nil {
		return "Bölge Uygun Değil"
	}
	if ff.Gold < cost.Gold {
		return "Yetersiz Altın"
	}
	if ff.Grain < cost.Grain {
		return "Yetersiz Tahıl"
	}
	if ff.Iron < cost.Iron {
		return "Yetersiz Demir"
	}
	if ff.Timber < cost.Timber {
		return "Yetersiz Kereste"
	}
	if ff.Stone < cost.Stone {
		return "Yetersiz Taş"
	}
	return ""
}

func RecruitPanelVisible(gs *state.GameState, rid world.RegionID) bool {
	if rid == "" {
		return false
	}
	r, ok := gs.Regions[rid]
	return ok && !r.IsSea && !r.IsLocked && r.OwnerID == string(gs.PlayerFactionID)
}

func RecruitPanelHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) string {
	return recruitUnitCardHitTest(mx, my, gs, rid)
}

func RecruitPanelActionHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) RecruitPanelAction {
	if recruitPanelCloseHitTest(mx, my, gs, rid) {
		return RecruitPanelAction{Kind: RecruitPanelActionClose}
	}
	if orderID := recruitQueueCancelHitTest(mx, my, gs, rid); orderID != "" {
		return RecruitPanelAction{Kind: RecruitPanelActionCancel, OrderID: orderID}
	}
	if recruitQueueIsFull(gs, rid) {
		return RecruitPanelAction{}
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
	slots := recruitPanelSlots()
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
	ensureArmySprites()
	slots := recruitPanelSlots()

	px := recruitPanelX(slots)
	py := recruitPanelY()
	pw := recruitPanelW(slots)
	ph := float32(recruitPanelH)
	panelRect := gameui.Rect{X: float64(px), Y: float64(py), W: float64(pw), H: float64(ph)}

	drawUIPanelFrame(screen, panelRect, panelBg, panelBorder, 1.5, 3)
	drawRecruitPanelCloseButton(screen, px, py, pw)

	DrawTextCentered(screen, "BİRİM OLUŞTUR", float64(px)+float64(pw)/2, float64(py)+8, FaceSmall, ColorGold)
	queuedTotal := queuedUnitTotal(gs, rid)
	landLimit := state.LandUnitProductionLimit(region)
	infoStr := fmt.Sprintf("Kışla limiti: %d  |  Sırada: %d", landLimit, queuedTotal)
	if region.IsCoastal(gs.Regions) {
		infoStr = fmt.Sprintf("Kışla: %d  |  Liman: %d  |  Sırada: %d", landLimit, state.NavalUnitProductionLimit(region), queuedTotal)
	}
	infoW := MeasureText(infoStr, FaceSmall)
	drawUIMutedText(screen, float64(px)+float64(pw)/2-infoW/2, float64(py)+24, infoStr)
	sepY := py + recruitHeaderH - 2
	drawUISeparator(screen, px+12, sepY, px+pw-12, 1, panelBorder)

	barracksLevel, portLevel := 0, 0
	for _, bid := range region.Buildings {
		switch bid {
		case "barracks":
			barracksLevel++
		case "port":
			portLevel++
		}
	}

	display := visibleUnitIDs(gs, region)
	topY := py + recruitHeaderH + 4
	cardW, cardH, gap := recruitCardMetrics(pw)
	maxTop := len(display)
	if maxTop > recruitMaxCards {
		maxTop = recruitMaxCards
	}
	for i := 0; i < maxTop; i++ {
		uid := display[i]
		row := i / recruitCardsPerRow
		col := i % recruitCardsPerRow
		x := px + recruitPanelPad + float32(col)*(cardW+gap)
		y := topY + float32(row)*(cardH+gap)
		drawRecruitCard(screen, gs, uid, barracksLevel, portLevel, x, y, cardW, cardH)
	}

	queueY := topY + recruitSectionH + recruitSectionGap
	drawRecruitQueueSection(screen, gs, rid, px, queueY, pw, recruitSectionH)
}

func recruitPanelCloseRect(px, py, pw float32) (x, y, w, h float32) {
	w = 18
	h = 18
	x = px + pw - w - 8
	y = py + 7
	return x, y, w, h
}

func drawRecruitPanelCloseButton(screen *ebiten.Image, px, py, pw float32) {
	x, y, w, h := recruitPanelCloseRect(px, py, pw)
	mx, my := ebiten.CursorPosition()
	hovered := float64(mx) >= float64(x) && float64(mx) <= float64(x+w) && float64(my) >= float64(y) && float64(my) <= float64(y+h)
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

func recruitPanelCloseHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) bool {
	btn, ok := buildRecruitPanelCloseButton(gs, rid)
	return ok && btn.HitTest(mx, my)
}

func recruitCardMetrics(panelW float32) (cardW, cardH, gap float32) {
	slots := recruitCardsPerRow
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

func drawRecruitCard(screen *ebiten.Image, gs *state.GameState, uid string, barracksLevel, portLevel int, sx, sy, cardW, cardH float32) {
	utype := gs.UnitTypes[uid]
	if utype == nil {
		return
	}
	requiredLevel := utype.RequiredBldgLevel
	if utype.RequiredBldg != "" && requiredLevel <= 0 {
		requiredLevel = 1
	}
	var needsBuilding bool
	switch utype.RequiredBldg {
	case "barracks":
		needsBuilding = barracksLevel < requiredLevel
	case "port":
		needsBuilding = portLevel < requiredLevel
	}
	ff := gs.Factions[gs.PlayerFactionID]
	playerOwnerID := string(gs.PlayerFactionID)
	needsTech := utype.RequiredTech != "" && (ff == nil || !ff.Research.Completed[utype.RequiredTech])
	canAfford := ff != nil && unitCost(utype).CanAfford(ff)
	fullyAvail := !needsBuilding && !needsTech && canAfford
	slotBg := color.RGBA{250, 250, 250, 240}
	borderCol := color.RGBA{160, 160, 160, 220}
	if fullyAvail {
		slotBg = color.RGBA{255, 255, 255, 245}
		borderCol = color.RGBA{145, 145, 145, 225}
	}
	drawUICardRect(screen, gameui.Rect{X: float64(sx), Y: float64(sy), W: float64(cardW), H: float64(cardH)}, slotBg, borderCol, 1)

	if sprite := unitSpriteForFaction(gs, playerOwnerID, uid); sprite != nil {
		tint := [3]float32{1, 1, 1}
		switch {
		case needsBuilding:
			tint = [3]float32{0.25, 0.25, 0.25}
		case needsTech:
			tint = [3]float32{0.45, 0.45, 0.45}
		case !canAfford:
			tint = [3]float32{0.65, 0.45, 0.45}
		}
		drawUnitSpriteCard(screen, sprite, sx, sy, cardW, tint)
	}
	drawUnitCardFooter(screen, sx, sy, cardW, cardH, unitCardFooterH)

	nameCol := color.RGBA{70, 60, 42, 235}
	if !fullyAvail {
		nameCol = color.RGBA{110, 105, 95, 210}
	}
	DrawTextCentered(screen, shortUnitName(utype.NameTR, 14), float64(sx)+float64(cardW)/2, float64(sy)+float64(cardH)-float64(unitCardNameOffset), FaceSmall, nameCol)
	DrawTextCentered(screen, itoa(utype.TurnsRequired)+"T", float64(sx)+float64(cardW)/2, float64(sy)+float64(cardH)-float64(unitCardSingleLabelOffset), FaceSmall, color.RGBA{110, 100, 86, 220})
}

func unitCost(utype *army.UnitType) economy.ResourceCost {
	if utype == nil {
		return economy.ResourceCost{}
	}
	return economy.ResourceCost{
		Gold:   utype.GoldCost,
		Grain:  utype.GrainCost,
		Iron:   utype.IronCost,
		Timber: utype.TimberCost,
		Stone:  utype.StoneCost,
	}
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

func recruitUnitCardHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) string {
	for _, btn := range buildRecruitUnitCardButtons(gs, rid) {
		if btn.HitTest(mx, my) {
			return btn.Label
		}
	}
	return ""
}

type recruitQueueItem struct {
	uid                string
	count              int
	queued             bool
	turns              int
	orderID            string
	progressesThisTurn bool
}

func recruitQueueItems(gs *state.GameState, rid world.RegionID) []recruitQueueItem {
	region := gs.Regions[rid]
	items := make([]recruitQueueItem, 0, recruitQueueMaxOrders)
	progressingByLane := make(map[string]int, 2)
	for _, order := range gs.ProductionQueue {
		if order.Kind != "unit" || order.RegionID != rid || order.FactionID != string(gs.PlayerFactionID) {
			continue
		}
		if len(items) >= recruitQueueMaxOrders {
			break
		}
		utype := gs.UnitTypes[order.TypeID]
		lane := recruitQueueLane(utype)
		progressesThisTurn := false
		capacity := state.UnitProductionLimit(region, utype)
		if progressingByLane[lane] < capacity {
			progressesThisTurn = true
			progressingByLane[lane]++
		}
		items = append(items, recruitQueueItem{
			uid:                order.TypeID,
			count:              1,
			queued:             true,
			turns:              order.TurnsLeft,
			orderID:            order.ID,
			progressesThisTurn: progressesThisTurn,
		})
	}
	return items
}

func recruitQueueLane(utype *army.UnitType) string {
	if utype != nil && utype.RequiredBldg == "port" {
		return "port"
	}
	return "barracks"
}

func drawRecruitQueueSection(screen *ebiten.Image, gs *state.GameState, rid world.RegionID, x, y, w, h float32) {
	mx, my := ebiten.CursorPosition()
	fmx, fmy := float64(mx), float64(my)
	queueRect := gameui.Rect{X: float64(x + 8), Y: float64(y), W: float64(w - 16), H: float64(h)}
	drawUICardRect(screen, queueRect, color.RGBA{14, 12, 10, 220}, color.RGBA{88, 72, 44, 220}, 1)
	drawUISectionLabel(screen, float64(x)+16, float64(y)+6, "EGİTİM SIRASI")
	items := recruitQueueItems(gs, rid)
	cardW, cardH, gap := recruitCardMetrics(w)
	cy := y + 26
	maxItems := len(items)
	if maxItems > recruitQueueMaxOrders {
		maxItems = recruitQueueMaxOrders
	}
	for i := 0; i < maxItems; i++ {
		it := items[i]
		row := i / recruitCardsPerRow
		col := i % recruitCardsPerRow
		startX := x + recruitPanelPad + float32(col)*(cardW+gap)
		cardY := cy + float32(row)*(cardH+gap)
		cardBg := color.RGBA{244, 237, 216, 245}
		cardBorder := color.RGBA{184, 150, 86, 232}
		spriteTint := [3]float32{1.0, 1.0, 1.0}
		labelColor := color.RGBA{85, 75, 50, 240}
		if !it.progressesThisTurn {
			cardBg = color.RGBA{202, 198, 192, 220}
			cardBorder = color.RGBA{122, 118, 112, 210}
			spriteTint = [3]float32{0.56, 0.56, 0.56}
			labelColor = color.RGBA{110, 104, 96, 220}
		}
		drawUICardRect(screen, gameui.Rect{X: float64(startX), Y: float64(cardY), W: float64(cardW), H: float64(cardH)}, cardBg, cardBorder, 1)
		if sprite := unitSpriteForFaction(gs, string(gs.PlayerFactionID), it.uid); sprite != nil {
			drawUnitSpriteCard(screen, sprite, startX, cardY, cardW, spriteTint)
		}
		drawUnitCardFooter(screen, startX, cardY, cardW, cardH, unitCardFooterH)
		unitName := it.uid
		if utype := gs.UnitTypes[it.uid]; utype != nil {
			unitName = utype.NameTR
		}
		DrawTextCentered(screen, shortUnitName(unitName, 14), float64(startX)+float64(cardW)/2, float64(cardY)+float64(cardH)-float64(unitCardNameOffset), FaceSmall, labelColor)
		label := "x" + itoa(it.count)
		if it.queued {
			label = "+" + itoa(it.turns) + "T"
			bx, by, bw, bh := startX+cardW-19, cardY+2, float32(17), float32(17)
			hovered := fmx >= float64(bx) && fmx <= float64(bx+bw) && fmy >= float64(by) && fmy <= float64(by+bh)
			drawQueueCancelButton(screen, bx, by, bw, bh, hovered)
		}
		DrawTextCentered(screen, label, float64(startX)+float64(cardW)/2, float64(cardY)+float64(cardH)-float64(unitCardSingleLabelOffset), FaceSmall, labelColor)
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
	for orderID, btn := range buildRecruitQueueCancelButtons(gs, rid) {
		if btn.HitTest(mx, my) {
			return orderID
		}
	}
	return ""
}

func RecruitPanelInteractiveHit(mx, my float64, gs *state.GameState, rid world.RegionID) bool {
	if recruitPanelCloseHitTest(mx, my, gs, rid) {
		return true
	}
	if recruitQueueCancelHitTest(mx, my, gs, rid) != "" {
		return true
	}
	return RecruitPanelHitTest(mx, my, gs, rid) != ""
}

func recruitPanelSlots() int {
	return recruitCardsPerRow
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
