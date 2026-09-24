package render

import (
	"image/color"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type tradeRouteVisual struct {
	factionA     string
	factionB     string
	goodName     string
	amount       int
	bestFlow     int
	route        *economy.TradeRoute
	routeKeys    []string
	routeDetails []tradeCorridorRouteDetail
}

type tradeCenterVisual struct {
	id          world.RegionID
	regionID    world.RegionID
	nameTR      string
	tier        world.TradeCenterTier
	worldX      float64
	worldY      float64
	x           float64
	y           float64
	labelX      float64
	labelY      float64
	labelW      float64
	labelH      float64
	offMap      bool
	endNode     bool
	mainRoute   bool
	active      bool
	unlockYear  int
	sourceGoods []world.HistoricalTradeGood
	landFocus   bool
}

type tradeCorridorInfo struct {
	fromName            string
	toName              string
	directionText       string
	amount              int
	factions            int
	goods               string
	sx                  float64
	sy                  float64
	cx                  float64
	cy                  float64
	dx                  float64
	dy                  float64
	path                []tradeOverlayPoint
	routeType           world.TradeRouteType
	hitWidth            float64
	dashed              bool
	historical          bool
	route               *economy.TradeRoute
	routeKeys           []string
	routeDetails        []tradeCorridorRouteDetail
	centerFactions      [2]string
	showAllRouteDetails bool
}

type tradeCorridorRouteDetail struct {
	key             string
	direction       string
	good            string
	amount          int
	route           *economy.TradeRoute
	historical      bool
	historicalGoods []tradeCorridorGoodTotal
}

const tradeCenterLabelGap = 20.0

const tradeSeaCenterLabelExtraGap = 20.0

type tradeCorridorGoodTotal struct {
	good   string
	amount int
}

func appendTradeCorridorRouteDetail(details []tradeCorridorRouteDetail, detail tradeCorridorRouteDetail) []tradeCorridorRouteDetail {
	if detail.historical {
		detail.key = "historical"
		detail.historicalGoods = appendTradeCorridorGoodTotal(detail.historicalGoods, detail.good, detail.amount)
		detail.good = ""
		detail.amount = 0
		for index := range details {
			if !details[index].historical && details[index].key != "historical" {
				continue
			}
			details[index].historical = true
			details[index].key = "historical"
			for _, total := range detail.historicalGoods {
				details[index].historicalGoods = appendTradeCorridorGoodTotal(details[index].historicalGoods, total.good, total.amount)
			}
			return details
		}
	}
	if detail.key == "" {
		return details
	}
	for _, existing := range details {
		if existing.key == detail.key {
			return details
		}
	}
	return append(details, detail)
}

func appendTradeCorridorGoodTotal(totals []tradeCorridorGoodTotal, good string, amount int) []tradeCorridorGoodTotal {
	if good == "" || amount == 0 {
		return totals
	}
	for index := range totals {
		if totals[index].good == good {
			totals[index].amount += amount
			return totals
		}
	}
	return append(totals, tradeCorridorGoodTotal{good: good, amount: amount})
}

func appendTradeCorridorRouteDetails(details []tradeCorridorRouteDetail, additions ...tradeCorridorRouteDetail) []tradeCorridorRouteDetail {
	for _, detail := range additions {
		details = appendTradeCorridorRouteDetail(details, detail)
	}
	return details
}

func sortTradeCorridorRouteDetails(details []tradeCorridorRouteDetail) {
	sort.SliceStable(details, func(i, j int) bool {
		return details[i].historical && !details[j].historical
	})
}

func appendTradeRouteKeys(keys []string, additions ...string) []string {
	for _, key := range additions {
		if key == "" {
			continue
		}
		seen := false
		for _, existing := range keys {
			if existing == key {
				seen = true
				break
			}
		}
		if !seen {
			keys = append(keys, key)
		}
	}
	return keys
}

func tradeCorridorDetailsForCenter(c tradeCorridorInfo) []tradeCorridorRouteDetail {
	if c.showAllRouteDetails {
		return c.routeDetails
	}
	if c.centerFactions[0] == "" && c.centerFactions[1] == "" {
		return c.routeDetails
	}
	details := make([]tradeCorridorRouteDetail, 0, len(c.routeDetails))
	for _, detail := range c.routeDetails {
		if detail.historical || detail.route == nil {
			details = append(details, detail)
			continue
		}
		fromID := detail.route.FromFactionID
		toID := detail.route.ToFactionID
		if fromID == c.centerFactions[0] || fromID == c.centerFactions[1] || toID == c.centerFactions[0] || toID == c.centerFactions[1] {
			details = append(details, detail)
		}
	}
	return details
}

type tradeOverlayPoint struct {
	x float64
	y float64
}

type tradePhysicalPathSegment struct {
	key    string
	points []tradeOverlayPoint
}

func tradePathSegmentKey(key string) string {
	return strings.SplitN(key, "#", 2)[0]
}

func splitTradePhysicalPath(points []tradeOverlayPoint, segmentKeys []string) []tradePhysicalPathSegment {
	if len(points) < 2 || len(segmentKeys) == 0 {
		return nil
	}
	segments := make([]tradePhysicalPathSegment, 0, len(points)-1)
	start := 0
	currentKey := tradePathSegmentKey(segmentKeys[0])
	for edge := 1; edge < len(points)-1; edge++ {
		key := ""
		if edge < len(segmentKeys) {
			key = tradePathSegmentKey(segmentKeys[edge])
		}
		if key == currentKey {
			continue
		}
		segments = append(segments, tradePhysicalPathSegment{
			key:    currentKey,
			points: append([]tradeOverlayPoint(nil), points[start:edge+1]...),
		})
		start = edge
		currentKey = key
	}
	segments = append(segments, tradePhysicalPathSegment{
		key:    currentKey,
		points: append([]tradeOverlayPoint(nil), points[start:]...),
	})
	return segments
}

func tradeSeaFocusPoints(points []tradeOverlayPoint, segmentKeys []string) []tradeOverlayPoint {
	if len(points) < 2 || len(segmentKeys) == 0 {
		return nil
	}
	bases := make([]string, len(segmentKeys))
	for i, key := range segmentKeys {
		bases[i] = tradePathSegmentKey(key)
	}
	firstSeaEdge := -1
	lastSeaEdge := -1
	for i, base := range bases {
		if strings.HasPrefix(base, "connector:") {
			continue
		}
		if firstSeaEdge < 0 {
			firstSeaEdge = i
		}
		lastSeaEdge = i
	}
	if firstSeaEdge >= 0 {
		result := []tradeOverlayPoint{points[firstSeaEdge]}
		lastPointIndex := lastSeaEdge + 1
		if lastPointIndex < len(points) {
			last := points[lastPointIndex]
			if last.x != result[0].x || last.y != result[0].y {
				result = append(result, last)
			}
		}
		return result
	}
	for i := 1; i < len(bases); i++ {
		if bases[i] != bases[i-1] {
			return []tradeOverlayPoint{points[i]}
		}
	}
	return nil
}

func tradeSeaMainPath(points []tradeOverlayPoint, segmentKeys []string) []tradeOverlayPoint {
	if len(points) < 2 || len(segmentKeys) == 0 {
		return nil
	}
	firstSeaEdge := -1
	lastSeaEdge := -1
	for i, key := range segmentKeys {
		if strings.HasPrefix(tradePathSegmentKey(key), "connector:") {
			continue
		}
		if firstSeaEdge < 0 {
			firstSeaEdge = i
		}
		lastSeaEdge = i
	}
	if firstSeaEdge < 0 || lastSeaEdge+1 >= len(points) {
		return nil
	}
	return points[firstSeaEdge : lastSeaEdge+2]
}

func drawTradeSeaFocusMarkers(screen *ebiten.Image, points []tradeOverlayPoint) {
	for _, point := range points {
		vector.FillCircle(screen, float32(point.x), float32(point.y), 13, color.RGBA{4, 16, 30, 245}, false)
		vector.FillCircle(screen, float32(point.x), float32(point.y), 11, color.RGBA{12, 42, 68, 255}, false)
	}
}

var (
	playerTradeRouteColor  = color.RGBA{242, 145, 52, 235}
	tradeSeaConnectorColor = color.RGBA{55, 166, 225, 235}
	tradeCenterIconOnce    sync.Once
	tradeCenterIcon        *ebiten.Image
)

const (
	tradeRouteDashLength = 12.0
	tradeRouteGapLength  = 10.0
	tradeRouteDashParts  = 72
	primaryTradeIconBox  = float32(42)
	primaryTradeIconSize = float32(34)
)

func loadTradeCenterIcon() {
	tradeCenterIconOnce.Do(func() {
		for _, path := range []string{
			"assets/ui/trade_center.png",
			filepath.Join("..", "..", "assets", "ui", "trade_center.png"),
		} {
			if icon := tryLoadImage(path); icon != nil {
				tradeCenterIcon = icon
				break
			}
		}
	})
}

func tradeRouteDisplayAmount(route *economy.TradeRoute) int {
	if route == nil || route.SuspendedTurns > 0 {
		return 0
	}
	return route.EffectiveAmountPerTurn()
}

func tradeCorridorTooltipHeight(c tradeCorridorInfo) float64 {
	details := tradeCorridorDetailsForCenter(c)
	if len(details) > 0 {
		lines := 0
		for _, detail := range details {
			if detail.historical {
				lines++
				lines += len(detail.historicalGoods)
				continue
			}
			lines += 2
		}
		height := 96.0 + float64(lines)*16
		if height < 128 {
			height = 128
		}
		return height
	}
	if c.dashed && c.route != nil {
		return 160
	}
	return 108
}

func tradeRouteTypeLabel(routeType world.TradeRouteType) string {
	switch routeType {
	case world.TradeRouteLand:
		return "Kara yolu"
	case world.TradeRouteSea:
		return "Deniz yolu"
	default:
		return "Belirtilmemiş"
	}
}

func tradeCorridorTooltipTitle(c tradeCorridorInfo) string {
	if c.dashed {
		return "Ticaret Anlaşması"
	}
	return "Ticaret Koridoru"
}

func tradeRoutePalette(routeType world.TradeRouteType) (glow, core, arrow color.RGBA) {
	if routeType == world.TradeRouteSea {
		return color.RGBA{72, 177, 232, 255}, color.RGBA{164, 231, 255, 255}, color.RGBA{55, 166, 225, 255}
	}
	return color.RGBA{255, 196, 92, 255}, color.RGBA{247, 232, 176, 255}, color.RGBA{225, 145, 42, 255}
}

func tradeSourceRoutePalette() (glow, core, arrow color.RGBA) {
	return color.RGBA{38, 112, 76, 255}, color.RGBA{92, 171, 116, 255}, color.RGBA{35, 126, 72, 255}
}

func passiveTradeRoutePalette() (glow, core, arrow color.RGBA) {
	gray := color.RGBA{145, 151, 158, 255}
	return color.RGBA{}, gray, gray
}

func tradeRoutePairKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "|" + b
}

func routeCurveOffset(key string, dist float64) float64 {
	if dist <= 0 {
		return 0
	}
	h := 0
	for i := 0; i < len(key); i++ {
		h = (h*31 + int(key[i])) & 0x7fffffff
	}
	sign := 1.0
	if h%2 == 0 {
		sign = -1.0
	}
	mag := dist * 0.11
	if mag < 18 {
		mag = 18
	}
	if mag > 96 {
		mag = 96
	}
	return sign * mag
}

func quadBezierPoint(x0, y0, cx, cy, x1, y1, t float64) (float64, float64) {
	u := 1 - t
	x := u*u*x0 + 2*u*t*cx + t*t*x1
	y := u*u*y0 + 2*u*t*cy + t*t*y1
	return x, y
}

func drawTradeFlowArrow(screen *ebiten.Image, sx, sy, cx, cy, dx, dy, t float64, reverse bool, col color.RGBA) {
	x, y := quadBezierPoint(sx, sy, cx, cy, dx, dy, t)
	step := 0.025
	if t+step > 1 {
		step = -step
	}
	tx, ty := quadBezierPoint(sx, sy, cx, cy, dx, dy, t+step)
	vx, vy := tx-x, ty-y
	if reverse {
		vx, vy = -vx, -vy
	}
	distance := math.Hypot(vx, vy)
	if distance < 0.1 {
		return
	}
	vx /= distance
	vy /= distance
	const arrowLength = 9.0
	const arrowWidth = 4.5
	baseX := x - vx*arrowLength
	baseY := y - vy*arrowLength
	leftX := baseX - vy*arrowWidth
	leftY := baseY + vx*arrowWidth
	rightX := baseX + vy*arrowWidth
	rightY := baseY - vx*arrowWidth
	vector.StrokeLine(screen, float32(leftX), float32(leftY), float32(x), float32(y), 2.4, col, false)
	vector.StrokeLine(screen, float32(rightX), float32(rightY), float32(x), float32(y), 2.4, col, false)
}

func tradePolylinePoint(points []tradeOverlayPoint, t float64) (float64, float64) {
	if len(points) == 0 {
		return 0, 0
	}
	if len(points) == 1 {
		return points[0].x, points[0].y
	}
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	position := t * float64(len(points)-1)
	index := int(position)
	if index >= len(points)-1 {
		last := points[len(points)-1]
		return last.x, last.y
	}
	local := position - float64(index)
	a := points[index]
	b := points[index+1]
	return a.x + (b.x-a.x)*local, a.y + (b.y-a.y)*local
}

func tradeCorridorPoint(c tradeCorridorInfo, t float64) (float64, float64) {
	if len(c.path) >= 2 {
		return tradePolylinePoint(c.path, t)
	}
	return quadBezierPoint(c.sx, c.sy, c.cx, c.cy, c.dx, c.dy, t)
}

func drawTradeFlowArrowOnPath(screen *ebiten.Image, points []tradeOverlayPoint, t float64, reverse bool, col color.RGBA) {
	x, y := tradePolylinePoint(points, t)
	step := 0.025
	if t+step > 1 {
		step = -step
	}
	tx, ty := tradePolylinePoint(points, t+step)
	vx, vy := tx-x, ty-y
	if reverse {
		vx, vy = -vx, -vy
	}
	distance := math.Hypot(vx, vy)
	if distance < 0.1 {
		return
	}
	vx /= distance
	vy /= distance
	const arrowLength = 9.0
	const arrowWidth = 4.5
	baseX := x - vx*arrowLength
	baseY := y - vy*arrowLength
	leftX := baseX - vy*arrowWidth
	leftY := baseY + vx*arrowWidth
	rightX := baseX + vy*arrowWidth
	rightY := baseY - vx*arrowWidth
	vector.StrokeLine(screen, float32(leftX), float32(leftY), float32(x), float32(y), 2.4, col, false)
	vector.StrokeLine(screen, float32(rightX), float32(rightY), float32(x), float32(y), 2.4, col, false)
}

func drawDashedTradeLine(screen *ebiten.Image, x1, y1, x2, y2 float64, lineW float32, lineColor color.RGBA, occludes func(float64, float64, float64, float64) bool) {
	dx := x2 - x1
	dy := y2 - y1
	length := math.Hypot(dx, dy)
	if length <= 0 {
		return
	}
	const dashLength = 5.0
	const gapLength = 6.0
	for position := 0.0; position < length; position += dashLength + gapLength {
		end := position + dashLength
		if end > length {
			end = length
		}
		startRatio := position / length
		endRatio := end / length
		sx := x1 + dx*startRatio
		sy := y1 + dy*startRatio
		ex := x1 + dx*endRatio
		ey := y1 + dy*endRatio
		if occludes == nil || !occludes(sx, sy, ex, ey) {
			vector.StrokeLine(screen, float32(sx), float32(sy), float32(ex), float32(ey), lineW, lineColor, false)
		}
	}
}

func (r *Renderer) drawTradeModeBackdrop(screen *ebiten.Image) {
	w := float32(ScreenWidth)
	h := float32(ScreenHeight)
	// Trade modunda haritayı tamamen kapatmak yerine hafif tint uygula.
	vector.FillRect(screen, 0, 0, w, h, color.RGBA{18, 26, 34, 72}, false)

}

func (r *Renderer) buildTradeCenters(maxCenters int) []tradeCenterVisual {
	if maxCenters <= 0 || len(r.gs.TradeCenters.Centers) == 0 {
		return nil
	}
	centers := make([]tradeCenterVisual, 0, maxCenters)
	for _, def := range r.gs.TradeCenters.Centers {
		if len(centers) >= maxCenters {
			break
		}
		active := def.ActiveInYear(r.gs.Year)
		if def.OffMap {
			sx, sy := r.worldToScreen(float64(def.WorldX), float64(def.WorldY))
			centers = append(centers, tradeCenterVisual{
				id:          def.ID,
				nameTR:      def.NameTR,
				tier:        def.Tier,
				worldX:      float64(def.WorldX),
				worldY:      float64(def.WorldY),
				x:           sx,
				y:           sy,
				offMap:      true,
				endNode:     len(def.Links) == 0,
				mainRoute:   def.MainRoute,
				active:      active,
				unlockYear:  def.UnlockYear,
				sourceGoods: append([]world.HistoricalTradeGood(nil), def.SourceGoods...),
			})
			continue
		}
		reg := r.gs.Regions[def.ID]
		if reg == nil || reg.IsSea || reg.TradeCapacity <= 0 {
			continue
		}
		landFocus := tradeCenterHasOnlyLandRoutes(def)
		sx, sy := r.tradePortScreenPos(reg, "")
		if landFocus && r.worldMap != nil {
			if ax, ay, ok := r.worldMap.PrimarySettlementAnchor(reg.ID); ok {
				sx, sy = r.worldToScreen(float64(ax), float64(ay))
			}
		}
		centers = append(centers, tradeCenterVisual{
			id:          reg.ID,
			regionID:    reg.ID,
			nameTR:      chooseRegionLabel(reg),
			tier:        def.Tier,
			worldX:      float64(reg.WorldX),
			worldY:      float64(reg.WorldY),
			x:           sx,
			y:           sy,
			mainRoute:   def.MainRoute,
			endNode:     len(def.Links) == 0,
			active:      active,
			unlockYear:  def.UnlockYear,
			sourceGoods: append([]world.HistoricalTradeGood(nil), def.SourceGoods...),
			landFocus:   landFocus,
		})
	}
	return centers
}

func tradeCenterHasOnlyLandRoutes(def world.TradeCenterDef) bool {
	hasLand := false
	hasSea := false
	for _, link := range def.Links {
		if link.Type == world.TradeRouteSea {
			hasSea = true
		} else {
			hasLand = true
		}
	}
	return hasLand && !hasSea
}

func historicalTradeGoodsLabel(goods []world.HistoricalTradeGood) string {
	if len(goods) == 0 {
		return "-"
	}
	labels := make([]string, 0, len(goods))
	for _, good := range goods {
		labels = append(labels, economy.GoodNameTR(good.Good))
	}
	return strings.Join(labels, ", ")
}

func historicalTradeGoodsDetail(goods []world.HistoricalTradeGood) string {
	if len(goods) == 0 {
		return "-"
	}
	labels := make([]string, 0, len(goods))
	for _, good := range goods {
		label := economy.GoodNameTR(good.Good) + " " + itoa(good.AmountPerTurn) + "/tur"
		if good.GoldIncomePerTurn > 0 {
			label += " (+" + itoa(good.GoldIncomePerTurn) + " altın)"
		}
		labels = append(labels, label)
	}
	return strings.Join(labels, ", ")
}

func (r *Renderer) tradeCenterBenefits(center tradeCenterVisual) (capacityBonus, incomeBonus int) {
	if r == nil || r.gs == nil || center.offMap || center.regionID == "" {
		return 0, 0
	}
	return r.gs.TradeCenterBenefits(r.gs.Regions[center.regionID])
}

func tradeCenterOwnerFaction(gs *state.GameState, center tradeCenterVisual) string {
	if gs == nil || center.regionID == "" {
		return ""
	}
	region := gs.Regions[center.regionID]
	if region == nil {
		return ""
	}
	return region.OwnerID
}

func tradeCenterTierLabel(tier world.TradeCenterTier) string {
	if tier == world.TradeCenterPrimary {
		return "Ana merkez"
	}
	return "İkincil merkez"
}

func chooseRegionLabel(region *world.Region) string {
	if region == nil {
		return ""
	}
	if region.NameTR != "" {
		return region.NameTR
	}
	if region.Name != "" {
		return region.Name
	}
	return string(region.ID)
}

func sqDistPointSegment(px, py, ax, ay, bx, by float64) float64 {
	abx := bx - ax
	aby := by - ay
	den := abx*abx + aby*aby
	if den <= 1e-6 {
		dx := px - ax
		dy := py - ay
		return dx*dx + dy*dy
	}
	t := ((px-ax)*abx + (py-ay)*aby) / den
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	cx := ax + abx*t
	cy := ay + aby*t
	dx := px - cx
	dy := py - cy
	return dx*dx + dy*dy
}

func (r *Renderer) tradeCorridorAt(fx, fy float64) int {
	if r.tradeOverlayOccludesPoint(fx, fy) {
		return -1
	}
	bestIdx := -1
	bestD2 := math.MaxFloat64
	for i := range r.tradeCorridors {
		c := r.tradeCorridors[i]
		segments := 24
		threshold := c.hitWidth * c.hitWidth
		prevX, prevY := tradeCorridorPoint(c, 0)
		for s := 1; s <= segments; s++ {
			t := float64(s) / float64(segments)
			x, y := tradeCorridorPoint(c, t)
			d2 := sqDistPointSegment(fx, fy, prevX, prevY, x, y)
			if d2 <= threshold && d2 < bestD2 {
				bestD2 = d2
				bestIdx = i
			}
			prevX, prevY = x, y
		}
	}
	return bestIdx
}

func (r *Renderer) tradeCenterAt(fx, fy float64) int {
	if r.tradeOverlayOccludesPoint(fx, fy) {
		return -1
	}
	bestIdx := -1
	best := math.MaxFloat64
	for i := range r.tradeCenters {
		c := r.tradeCenters[i]
		if c.labelW > 0 && (gameui.Rect{X: c.labelX, Y: c.labelY, W: c.labelW, H: c.labelH}).Hit(fx, fy) {
			d := math.Hypot(fx-c.x, fy-c.y)
			if d < best {
				best = d
				bestIdx = i
			}
			continue
		}
		d := math.Hypot(fx-c.x, fy-c.y)
		if d <= 12 && d < best {
			best = d
			bestIdx = i
		}
	}
	return bestIdx
}

func (r *Renderer) updateTradeHover() {
	r.tradeHoverIdx = -1
	r.tradeCenterIdx = -1
	if r.showTrade {
		return
	}
	if r.mapMode != MapModeTrade || (len(r.tradeCorridors) == 0 && len(r.tradeCenters) == 0) {
		return
	}
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	r.tradeHoverIdx = r.tradeCorridorAt(fx, fy)
	if r.tradeHoverIdx < 0 {
		r.tradeCenterIdx = r.tradeCenterAt(fx, fy)
	}
}

func (r *Renderer) drawTradeHoverTooltip(screen *ebiten.Image) {
	if r.tradeHoverIdx < 0 || r.tradeHoverIdx >= len(r.tradeCorridors) {
		return
	}
	c := r.tradeCorridors[r.tradeHoverIdx]
	if c.dashed {
		drawDashedTradeCurve(screen, c.sx, c.sy, c.cx, c.cy, c.dx, c.dy, 4.0,
			playerTradeRouteColor, r.tradeOverlayOccludesSegment)
	} else {
		hoverGlow, hoverCore, _ := tradeRoutePalette(c.routeType)
		hoverGlow.A = 56
		hoverCore.A = 230
		segments := 28
		for i := 0; i < segments; i++ {
			t1 := float64(i) / float64(segments)
			t2 := float64(i+1) / float64(segments)
			x1, y1 := tradeCorridorPoint(c, t1)
			x2, y2 := tradeCorridorPoint(c, t2)
			if r.tradeOverlayOccludesSegment(x1, y1, x2, y2) {
				continue
			}
			vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 9.0, hoverGlow, false)
			vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 3.0, hoverCore, false)
		}
	}
	if !r.tradeOverlayOccludesPoint(c.sx, c.sy) {
		vector.FillCircle(screen, float32(c.sx), float32(c.sy), 6, playerTradeRouteColor, true)
	}
	if !r.tradeOverlayOccludesPoint(c.dx, c.dy) {
		vector.FillCircle(screen, float32(c.dx), float32(c.dy), 6, playerTradeRouteColor, true)
	}

	rect, ok := r.tradeHoverTooltipRect()
	if !ok {
		return
	}
	x := float32(rect.X)
	y := float32(rect.Y)
	w := float32(rect.W)
	h := float32(rect.H)
	vector.FillRect(screen, x, y, w, h, color.RGBA{10, 14, 20, 230}, false)
	vector.StrokeRect(screen, x, y, w, h, 1.2, color.RGBA{145, 120, 74, 230}, false)
	DrawText(screen, tradeCorridorTooltipTitle(c), float64(x)+10, float64(y)+8, FaceSmall, color.RGBA{242, 226, 174, 255})
	directionText := c.directionText
	if directionText == "" {
		directionText = c.fromName + " ↔ " + c.toName
	}
	DrawText(screen, directionText, float64(x)+10, float64(y)+28, FaceSmall, color.RGBA{215, 225, 236, 235})
	DrawText(screen, "Tür: "+tradeRouteTypeLabel(c.routeType), float64(x)+10, float64(y)+46, FaceSmall, color.RGBA{197, 190, 168, 230})
	details := tradeCorridorDetailsForCenter(c)
	if len(details) > 0 && r.gs != nil {
		lineY := float64(y) + 64
		for _, detail := range details {
			if detail.historical {
				routeLabel := trimTextToWidth("Rota: Tarihsel akış", FaceSmall, rect.W-20)
				DrawText(screen, routeLabel, float64(x)+10, lineY, FaceSmall, color.RGBA{225, 212, 180, 240})
				lineY += 16
				goods := append([]tradeCorridorGoodTotal(nil), detail.historicalGoods...)
				if len(goods) == 0 && detail.good != "" {
					goods = append(goods, tradeCorridorGoodTotal{good: detail.good, amount: detail.amount})
				}
				sort.Slice(goods, func(i, j int) bool {
					return goods[i].good < goods[j].good
				})
				for _, total := range goods {
					goodsLabel := trimTextToWidth(total.good+": "+itoa(total.amount)+"/tur", FaceSmall, rect.W-20)
					DrawText(screen, goodsLabel, float64(x)+10, lineY, FaceSmall, color.RGBA{187, 203, 222, 230})
					lineY += 16
				}
				continue
			}
			routeLabel := detail.direction
			if routeLabel == "" {
				routeLabel = c.directionText
			}
			routeLabel = trimTextToWidth(routeLabel, FaceSmall, rect.W-20)
			DrawText(screen, routeLabel, float64(x)+10, lineY, FaceSmall, color.RGBA{225, 212, 180, 240})

			goldLabel := "-"
			if detail.route != nil {
				gold := detail.amount * detail.route.GoldPerUnit
				goldLabel = itoa(gold) + " Altın"
			}
			goodsLabel := detail.good
			if goodsLabel == "" {
				goodsLabel = "-"
			}
			goodsLabel = trimTextToWidth(goodsLabel+" "+itoa(detail.amount)+" | "+goldLabel, FaceSmall, rect.W-20)
			DrawText(screen, goodsLabel, float64(x)+10, lineY+16, FaceSmall, color.RGBA{187, 203, 222, 230})
			lineY += 32
		}
	} else if c.dashed && c.route != nil && r.gs != nil {
		amount := tradeRouteDisplayAmount(c.route)
		gold := amount * c.route.GoldPerUnit
		goodName := economy.GoodNameTR(c.route.Good)
		playerID := string(r.gs.PlayerFactionID)
		DrawText(screen, "Hacim: "+itoa(amount)+"/tur   Emtia: "+goodName, float64(x)+10, float64(y)+64, FaceSmall, color.RGBA{187, 203, 222, 230})
		if c.route.FromFactionID == playerID {
			DrawText(screen, "Veriyoruz: "+goodName+" "+itoa(amount)+"/tur", float64(x)+10, float64(y)+82, FaceSmall, color.RGBA{225, 205, 170, 240})
			DrawText(screen, "Alıyoruz: Altın +"+itoa(gold)+"/tur", float64(x)+10, float64(y)+100, FaceSmall, color.RGBA{225, 205, 170, 240})
			DrawText(screen, "Gelir: +"+itoa(gold)+" altın/tur", float64(x)+10, float64(y)+118, FaceSmall, color.RGBA{145, 220, 155, 245})
		} else {
			DrawText(screen, "Veriyoruz: Altın "+itoa(gold)+"/tur", float64(x)+10, float64(y)+82, FaceSmall, color.RGBA{225, 205, 170, 240})
			DrawText(screen, "Alıyoruz: "+goodName+" "+itoa(amount)+"/tur", float64(x)+10, float64(y)+100, FaceSmall, color.RGBA{225, 205, 170, 240})
			DrawText(screen, "Ödeme: -"+itoa(gold)+" altın/tur", float64(x)+10, float64(y)+118, FaceSmall, color.RGBA{230, 170, 135, 240})
		}
		if c.route.SuspendedTurns > 0 {
			DrawText(screen, "Askıda: "+itoa(c.route.SuspendedTurns)+" tur", float64(x)+10, float64(y)+136, FaceSmall, color.RGBA{230, 170, 135, 240})
		}
	} else {
		label := "Devlet: " + itoa(c.factions)
		if c.historical {
			label = "Tarihsel akış"
		}
		DrawText(screen, "Hacim: "+itoa(c.amount)+"/tur   "+label, float64(x)+10, float64(y)+64, FaceSmall, color.RGBA{187, 203, 222, 230})
		DrawText(screen, "Emtia: "+c.goods, float64(x)+10, float64(y)+82, FaceSmall, color.RGBA{197, 190, 168, 230})
	}
}

func drawDashedTradeCurve(screen *ebiten.Image, sx, sy, cx, cy, dx, dy float64, lineW float32, lineColor color.RGBA, occludes func(float64, float64, float64, float64) bool) {
	// Eşit t aralıkları eğrinin farklı noktalarında farklı fiziksel uzunluklar
	// üretir. Önce eğriyi sabit sayıda küçük yay parçasına örnekleyip, dash/gap
	// paternini bu parçaların yaklaşık gerçek piksel uzunluğu üzerinden uygula.
	var xs [tradeRouteDashParts + 1]float64
	var ys [tradeRouteDashParts + 1]float64
	var distances [tradeRouteDashParts + 1]float64
	for i := 0; i <= tradeRouteDashParts; i++ {
		t := float64(i) / float64(tradeRouteDashParts)
		xs[i], ys[i] = quadBezierPoint(sx, sy, cx, cy, dx, dy, t)
		if i == 0 {
			continue
		}
		distances[i] = distances[i-1] + math.Hypot(xs[i]-xs[i-1], ys[i]-ys[i-1])
	}

	patternLength := tradeRouteDashLength + tradeRouteGapLength
	for i := 0; i < tradeRouteDashParts; i++ {
		segmentStart := distances[i]
		segmentEnd := distances[i+1]
		segmentLength := segmentEnd - segmentStart
		if segmentLength <= 0 {
			continue
		}

		position := segmentStart
		for position < segmentEnd {
			patternPosition := math.Mod(position, patternLength)
			drawing := patternPosition < tradeRouteDashLength
			chunkLength := tradeRouteDashLength - patternPosition
			if !drawing {
				chunkLength = patternLength - patternPosition
			}
			chunkEnd := position + chunkLength
			if chunkEnd > segmentEnd {
				chunkEnd = segmentEnd
			}

			if drawing {
				startRatio := (position - segmentStart) / segmentLength
				endRatio := (chunkEnd - segmentStart) / segmentLength
				x1 := xs[i] + (xs[i+1]-xs[i])*startRatio
				y1 := ys[i] + (ys[i+1]-ys[i])*startRatio
				x2 := xs[i] + (xs[i+1]-xs[i])*endRatio
				y2 := ys[i] + (ys[i+1]-ys[i])*endRatio
				if occludes == nil || !occludes(x1, y1, x2, y2) {
					vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), lineW, lineColor, false)
				}
			}
			if chunkEnd <= position {
				break
			}
			position = chunkEnd
		}
	}
}

func (r *Renderer) tradePortScreenPos(region *world.Region, settlementID string) (float64, float64) {
	if r != nil && r.worldMap != nil && region != nil {
		for index, settlement := range region.Settlements {
			if settlement.Type != world.SettlementPort {
				continue
			}
			if settlementID != "" && settlement.ID != settlementID {
				continue
			}
			if ax, ay, ok := r.worldMap.SettlementAnchor(region.ID, index); ok {
				return r.worldToScreen(float64(ax), float64(ay))
			}
		}
	}
	if settlementID != "" {
		return r.tradePortScreenPos(region, "")
	}
	return r.regionScreenPos(region)
}

func (r *Renderer) tradePortScreenPosForSea(region *world.Region, seaID world.RegionID) (float64, float64) {
	if r == nil || r.gs == nil || r.worldMap == nil || region == nil || seaID == "" {
		return r.tradePortScreenPos(region, "")
	}
	sea := r.gs.Regions[seaID]
	if sea == nil {
		return r.tradePortScreenPos(region, "")
	}
	seaX, seaY := r.worldToScreen(wcX(sea.WorldX), wcY(sea.WorldY))
	bestDistance := math.MaxFloat64
	bestX, bestY := 0.0, 0.0
	found := false
	for index, settlement := range region.Settlements {
		if settlement.Type != world.SettlementPort {
			continue
		}
		ax, ay, ok := r.worldMap.SettlementAnchor(region.ID, index)
		if !ok {
			continue
		}
		portX, portY := r.worldToScreen(float64(ax), float64(ay))
		dx := portX - seaX
		dy := portY - seaY
		distance := dx*dx + dy*dy
		if distance >= bestDistance {
			continue
		}
		bestDistance = distance
		bestX, bestY = portX, portY
		found = true
	}
	if found {
		return bestX, bestY
	}
	return r.tradePortScreenPos(region, "")
}

// merchantTradePortCorridor, merchant filosunun rota çizgisini genel ticaret
// merkezi koridorlarından bağımsız olarak canonical liman çiftinden üretir.
// Genel merkez grafiği bir faction rotasını en yakın merkezin koridoruna
// taşıyabildiği için marker connector'ı için güvenilir kaynak değildir.
func (r *Renderer) merchantTradePortCorridor(route *economy.TradeRoute, visualKey string) (tradeCorridorInfo, bool) {
	if r == nil || r.gs == nil || route == nil || route.AssignmentKey() == "" || visualKey == "" {
		return tradeCorridorInfo{}, false
	}
	if len(r.gs.MerchantTradeRouteSeaRegions(route)) == 0 {
		return tradeCorridorInfo{}, false
	}
	pairs := r.gs.MerchantTradeRoutePortPairs(route)
	if len(pairs) == 0 {
		return tradeCorridorInfo{}, false
	}
	pair := pairs[0]
	fromRegion := r.gs.Regions[pair.FromRegionID]
	toRegion := r.gs.Regions[pair.ToRegionID]
	if fromRegion == nil || toRegion == nil {
		return tradeCorridorInfo{}, false
	}
	sx, sy := r.tradePortScreenPos(fromRegion, pair.FromSettlementID)
	dx, dy := r.tradePortScreenPos(toRegion, pair.ToSettlementID)
	mx := (sx + dx) / 2
	my := (sy + dy) / 2
	vx := dx - sx
	vy := dy - sy
	dist := math.Hypot(vx, vy)
	if dist < 1 {
		return tradeCorridorInfo{}, false
	}
	curve := routeCurveOffset("player-port|"+visualKey, dist)
	return tradeCorridorInfo{
		fromName:      chooseRegionLabel(fromRegion),
		toName:        chooseRegionLabel(toRegion),
		directionText: chooseRegionLabel(fromRegion) + " → " + chooseRegionLabel(toRegion),
		amount:        tradeRouteDisplayAmount(route),
		factions:      2,
		goods:         economy.GoodNameTR(route.Good),
		routeType:     world.TradeRouteSea,
		sx:            sx,
		sy:            sy,
		cx:            mx + (-vy/dist)*curve,
		cy:            my + (vx/dist)*curve,
		dx:            dx,
		dy:            dy,
		hitWidth:      10,
		dashed:        true,
		route:         route,
		routeKeys:     []string{route.AssignmentKey()},
	}, true
}

func (r *Renderer) drawPlayerTradePortRoutes(screen *ebiten.Image, merged map[string]tradeRouteVisual) {
	if r == nil || r.gs == nil || len(merged) == 0 {
		return
	}
	playerID := string(r.gs.PlayerFactionID)
	keys := make([]string, 0, len(merged))
	for key, route := range merged {
		if route.route == nil || (route.factionA != playerID && route.factionB != playerID) {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		route := merged[key]
		corridor, ok := r.merchantTradePortCorridor(route.route, key)
		if !ok {
			continue
		}
		drawDashedTradeCurve(screen, corridor.sx, corridor.sy, corridor.cx, corridor.cy, corridor.dx, corridor.dy, 3.0, playerTradeRouteColor, r.tradeOverlayOccludesSegment)
		drawTradeFlowArrow(screen, corridor.sx, corridor.sy, corridor.cx, corridor.cy, corridor.dx, corridor.dy, 0.5, false, playerTradeRouteColor)
		if !r.tradeOverlayOccludesPoint(corridor.sx, corridor.sy) {
			vector.FillCircle(screen, float32(corridor.sx), float32(corridor.sy), 5, playerTradeRouteColor, true)
			vector.StrokeCircle(screen, float32(corridor.sx), float32(corridor.sy), 8, 1.2, color.RGBA{92, 54, 18, 220}, true)
		}
		if !r.tradeOverlayOccludesPoint(corridor.dx, corridor.dy) {
			vector.FillCircle(screen, float32(corridor.dx), float32(corridor.dy), 5, playerTradeRouteColor, true)
			vector.StrokeCircle(screen, float32(corridor.dx), float32(corridor.dy), 8, 1.2, color.RGBA{92, 54, 18, 220}, true)
		}
		corridor.routeKeys = route.routeKeys
		corridor.routeDetails = route.routeDetails
		r.tradeCorridors = append(r.tradeCorridors, corridor)
	}
}

func tradeCorridorHasRoute(c tradeCorridorInfo, routeKey string) bool {
	if routeKey == "" {
		return false
	}
	if c.route != nil && c.route.AssignmentKey() == routeKey {
		return true
	}
	for _, key := range c.routeKeys {
		if key == routeKey {
			return true
		}
	}
	return false
}

func closestPointOnTradeSegment(px, py, ax, ay, bx, by float64) (float64, float64, float64) {
	abx := bx - ax
	aby := by - ay
	den := abx*abx + aby*aby
	t := 0.0
	if den > 1e-6 {
		t = ((px-ax)*abx + (py-ay)*aby) / den
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
	}
	x := ax + abx*t
	y := ay + aby*t
	dx := px - x
	dy := py - y
	return x, y, dx*dx + dy*dy
}

func nearestTradeCorridorPoint(c tradeCorridorInfo, px, py float64) (float64, float64, float64, bool) {
	const segments = 28
	bestD2 := math.MaxFloat64
	bestX, bestY := 0.0, 0.0
	prevX, prevY := tradeCorridorPoint(c, 0)
	for i := 1; i <= segments; i++ {
		t := float64(i) / float64(segments)
		x, y := tradeCorridorPoint(c, t)
		candidateX, candidateY, d2 := closestPointOnTradeSegment(px, py, prevX, prevY, x, y)
		if d2 < bestD2 {
			bestX, bestY, bestD2 = candidateX, candidateY, d2
		}
		prevX, prevY = x, y
	}
	if bestD2 == math.MaxFloat64 {
		return 0, 0, 0, false
	}
	return bestX, bestY, bestD2, true
}

func (r *Renderer) tradeRouteConnectionPoint(routeKey string, px, py float64) (float64, float64, bool) {
	if r == nil || routeKey == "" {
		return 0, 0, false
	}
	if route := merchantRouteForKey(r.gs, routeKey); route != nil {
		if corridor, ok := r.merchantTradePortCorridor(route, tradeRoutePairKey(route.FromFactionID, route.ToFactionID)); ok {
			// Merchant filosunun connector'ı rota eğrisinin en yakın
			// noktasında kesilmez; rota yönündeki gerçek hedef limanda biter.
			return corridor.dx, corridor.dy, true
		}
		// Rota canonical liman çifti taşıyor ancak merchant deniz rotası
		// olarak geçerli değilse genel merkez koridoruna düşme; bu, eski save'de
		// kalmış kara rota atamasını yanlış görsel hatta bağlar.
		if len(r.gs.MerchantTradeRoutePortPairs(route)) > 0 {
			return 0, 0, false
		}
	}
	bestD2 := math.MaxFloat64
	bestX, bestY := 0.0, 0.0
	for _, corridor := range r.tradeCorridors {
		if !tradeCorridorHasRoute(corridor, routeKey) {
			continue
		}
		x, y, d2, ok := nearestTradeCorridorPoint(corridor, px, py)
		if ok && d2 < bestD2 {
			bestX, bestY, bestD2 = x, y, d2
		}
	}
	return bestX, bestY, bestD2 < math.MaxFloat64
}

type tradeBonusFleetVisual struct {
	position armyIconPos
	fleet    *army.Army
	bonus    int
	routeKey string
}

func (r *Renderer) tradeBonusFleetAtPosition(position armyIconPos) (*army.Army, bool) {
	if r == nil || r.gs == nil {
		return nil, false
	}
	fleet := r.gs.Armies[position.ArmyID]
	if fleet == nil {
		return nil, false
	}
	status, ok := r.merchantTradeStatusForArmy(fleet.ID)
	if !ok || !tradeMapMerchantFleetVisible(r.gs, fleet, status) {
		return nil, false
	}
	return fleet, true
}

// tradeMapMerchantFleetVisible, ticaret haritasında oyuncu dışı merchant
// filolarının yalnızca gerçekten bonus üreten rotaları göstermesini sağlar.
// Oyuncunun bekleyen filosu ise rotaya giderken durumunu takip edebilmesi için
// görünür kalır.
func tradeMapMerchantFleetVisible(gs *state.GameState, fleet *army.Army, status state.MerchantFleetTradeStatus) bool {
	if gs == nil || fleet == nil {
		return false
	}
	if fleet.OwnerID == string(gs.PlayerFactionID) {
		return status.Bonus > 0 || status.Pending
	}
	return status.Bonus > 0
}

func (r *Renderer) tradeBonusFleetVisuals() []tradeBonusFleetVisual {
	if r == nil || r.gs == nil {
		return nil
	}
	positions := r.armyIconPositions()
	visuals := make([]tradeBonusFleetVisual, 0, len(positions))
	for _, position := range positions {
		fleet, ok := r.tradeBonusFleetAtPosition(position)
		if !ok {
			continue
		}
		bonus := r.merchantTradeBonusForArmy(fleet)
		visuals = append(visuals, tradeBonusFleetVisual{
			position: position,
			fleet:    fleet,
			bonus:    bonus,
			routeKey: fleet.TradeRouteKey,
		})
	}
	return visuals
}

func (r *Renderer) drawTradeBonusFleetMarkers(screen *ebiten.Image) {
	if r == nil || r.gs == nil {
		return
	}
	positions := r.armyIconPositions()
	if len(positions) == 0 {
		return
	}

	// Connector'lar markerların altında kalır; marker ve mevcut bonus rozeti
	// en son çizilerek rota hattı tarafından kapatılmaz.
	for _, position := range positions {
		fleet, ok := r.tradeBonusFleetAtPosition(position)
		if !ok {
			continue
		}
		fromX, fromY := float64(position.X), float64(position.Y)
		toX, toY, ok := r.tradeRouteConnectionPoint(fleet.TradeRouteKey, fromX, fromY)
		if !ok {
			continue
		}
		vx := toX - fromX
		vy := toY - fromY
		distance := math.Hypot(vx, vy)
		if distance < 18 {
			continue
		}
		fromX += vx / distance * 15
		fromY += vy / distance * 15
		toX -= vx / distance * 3
		toY -= vy / distance * 3
		if r.tradeOverlayOccludesSegment(fromX, fromY, toX, toY) {
			continue
		}
		if r.merchantTradeBonusForArmy(fleet) > 0 {
			vector.StrokeLine(screen, float32(fromX), float32(fromY), float32(toX), float32(toY), 4.5, color.RGBA{22, 25, 30, 180}, false)
			vector.StrokeLine(screen, float32(fromX), float32(fromY), float32(toX), float32(toY), 1.8, color.RGBA{244, 195, 52, 210}, false)
		} else {
			drawDashedTradeLine(screen, fromX, fromY, toX, toY, 4.5, color.RGBA{22, 25, 30, 180}, r.tradeOverlayOccludesSegment)
			drawDashedTradeLine(screen, fromX, fromY, toX, toY, 1.8, color.RGBA{145, 151, 158, 220}, r.tradeOverlayOccludesSegment)
		}
	}

	for _, position := range positions {
		fleet, ok := r.tradeBonusFleetAtPosition(position)
		if !ok {
			continue
		}
		if _, _, connected := r.tradeRouteConnectionPoint(fleet.TradeRouteKey, float64(position.X), float64(position.Y)); !connected {
			continue
		}
		unitCount := len(fleet.Units)
		fc := factionColor(r.gs, fleet.OwnerID)
		r.drawArmyIcon(screen, fleet.ID, fleet.OwnerID, position.X, position.Y, fc, unitCount, false, true, false, position.X+armyIconInnerHalf+8)
		r.drawNavalPriorityBadges(screen, fleet, position.X, position.Y)
	}
}

func (r *Renderer) nearestTradeCenterIndex(region *world.Region, centers []tradeCenterVisual) int {
	if region == nil || len(centers) == 0 {
		return -1
	}
	rx := float64(region.WorldX)
	ry := float64(region.WorldY)
	bestIdx := -1
	bestDist := math.MaxFloat64
	for i, c := range centers {
		if c.offMap || !c.active {
			continue
		}
		d := math.Hypot(rx-c.worldX, ry-c.worldY)
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	return bestIdx
}

func (r *Renderer) buildTradeCenterAdjacency(centers []tradeCenterVisual) map[int][]int {
	adj := make(map[int][]int, len(centers))
	if len(centers) == 0 {
		return adj
	}
	indexByID := make(map[world.RegionID]int, len(centers))
	for i := range centers {
		indexByID[centers[i].id] = i
	}

	// Economic center connectivity is bidirectional for normal centers, while
	// off-map source routes remain directed. Visual arrow direction is kept
	// separately from this adjacency and still comes from raw Links.
	for fromID, linkedIDs := range r.gs.TradeCenters.TradeAdjacency() {
		from, ok := indexByID[fromID]
		if !ok || !centers[from].active {
			continue
		}
		for _, linkedID := range linkedIDs {
			to, ok := indexByID[linkedID]
			if !ok || to == from || !centers[to].active {
				continue
			}
			adj[from] = append(adj[from], to)
		}
	}

	// dedup + sort
	for i := range centers {
		neighbors := adj[i]
		if len(neighbors) == 0 {
			continue
		}
		sort.Ints(neighbors)
		uniq := neighbors[:0]
		prev := -1
		for _, n := range neighbors {
			if n == prev {
				continue
			}
			prev = n
			uniq = append(uniq, n)
		}
		adj[i] = uniq
	}
	return adj
}

func tradeCenterVisualDirections(gs *state.GameState, centers []tradeCenterVisual) map[string]map[struct{ from, to int }]struct{} {
	directions := make(map[string]map[struct{ from, to int }]struct{})
	if gs == nil {
		return directions
	}
	indexByID := make(map[world.RegionID]int, len(centers))
	for i, center := range centers {
		indexByID[center.id] = i
	}
	for _, def := range gs.TradeCenters.Centers {
		from, ok := indexByID[def.ID]
		if !ok {
			continue
		}
		for _, link := range def.Links {
			to, ok := indexByID[link.RegionID]
			if !ok || from == to {
				continue
			}
			a, b := from, to
			if a > b {
				a, b = b, a
			}
			key := itoa(a) + "|" + itoa(b)
			if directions[key] == nil {
				directions[key] = make(map[struct{ from, to int }]struct{})
			}
			directions[key][struct{ from, to int }{from: from, to: to}] = struct{}{}
		}
	}
	return directions
}

func shortestCenterPath(adj map[int][]int, from, to int) []int {
	if from < 0 || to < 0 {
		return nil
	}
	if from == to {
		return []int{from}
	}
	queue := []int{from}
	prev := map[int]int{from: -1}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, nxt := range adj[cur] {
			if _, seen := prev[nxt]; seen {
				continue
			}
			prev[nxt] = cur
			if nxt == to {
				path := []int{to}
				for p := cur; p >= 0; p = prev[p] {
					path = append(path, p)
				}
				// reverse
				for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
					path[i], path[j] = path[j], path[i]
				}
				return path
			}
			queue = append(queue, nxt)
		}
	}
	return nil
}

// shortestCenterPathForTradeRoute, vassal hedefi için aynı realm içindeki
// ticaret merkezlerini transit aday olarak tercih eder. Böylece örneğin
// Flandre-HRE ilişkisindeki rota, eşit uzunluktaki rastgele bir batı yoluna
// sapmak yerine HRE'nin Palatine merkezi üzerinden görünür.
func shortestCenterPathForTradeRoute(gs *state.GameState, adj map[int][]int, centers []tradeCenterVisual, from, to int, factionIDs ...string) []int {
	basePath := shortestCenterPath(adj, from, to)
	if gs == nil || len(basePath) < 2 || len(factionIDs) == 0 {
		return basePath
	}

	preferredOwners := make(map[string]struct{}, len(factionIDs))
	for _, factionID := range factionIDs {
		if factionID == "" {
			continue
		}
		current := gs.Factions[faction.FactionID(factionID)]
		if current == nil || current.OverlordID == "" {
			continue
		}
		preferredOwners[string(current.OverlordID)] = struct{}{}
	}
	if len(preferredOwners) == 0 {
		return basePath
	}

	bestPath := basePath
	bestTransitID := ""
	for transitIndex, center := range centers {
		if transitIndex == from || transitIndex == to || center.regionID == "" {
			continue
		}
		if _, ok := preferredOwners[tradeCenterOwnerFaction(gs, center)]; !ok {
			continue
		}
		left := shortestCenterPath(adj, from, transitIndex)
		right := shortestCenterPath(adj, transitIndex, to)
		if len(left) < 2 || len(right) < 2 {
			continue
		}
		candidate := append(append([]int(nil), left...), right[1:]...)
		if len(candidate) < len(bestPath) || len(candidate) == len(bestPath) && (bestTransitID == "" || center.id < world.RegionID(bestTransitID)) {
			bestPath = candidate
			bestTransitID = string(center.id)
		}
	}
	return bestPath
}

// shortestTradeRegionPath, ticaret merkezi link'inin fiziksel türüne uygun
// harita bölgeleri arasında yol arar. Böylece merkez grafındaki soyut kenar
// doğrudan iki merkez arasında çizilmek yerine kara veya deniz düğümlerini
// izler.
func shortestTradeRegionPath(regions map[world.RegionID]*world.Region, from, to world.RegionID, sea bool) []world.RegionID {
	if from == "" || to == "" {
		return nil
	}
	fromRegion := regions[from]
	toRegion := regions[to]
	if fromRegion == nil || toRegion == nil || fromRegion.IsSea != sea || toRegion.IsSea != sea {
		return nil
	}
	if from == to {
		return []world.RegionID{from}
	}
	queue := []world.RegionID{from}
	previous := map[world.RegionID]world.RegionID{from: ""}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		currentRegion := regions[current]
		if currentRegion == nil {
			continue
		}
		neighbors := append([]world.RegionID(nil), currentRegion.Neighbors...)
		sort.Slice(neighbors, func(i, j int) bool { return neighbors[i] < neighbors[j] })
		for _, next := range neighbors {
			if _, seen := previous[next]; seen {
				continue
			}
			nextRegion := regions[next]
			if nextRegion == nil || nextRegion.IsSea != sea {
				continue
			}
			previous[next] = current
			if next == to {
				path := []world.RegionID{to}
				for currentID := current; currentID != ""; currentID = previous[currentID] {
					path = append(path, currentID)
				}
				for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
					path[i], path[j] = path[j], path[i]
				}
				return path
			}
			queue = append(queue, next)
		}
	}
	return nil
}

func nearestTradeRegionID(regions map[world.RegionID]*world.Region, x, y float64, sea bool) world.RegionID {
	bestID := world.RegionID("")
	bestDistance := math.MaxFloat64
	for id, region := range regions {
		if region == nil || region.IsSea != sea {
			continue
		}
		dx := float64(region.WorldX) - x
		dy := float64(region.WorldY) - y
		distance := dx*dx + dy*dy
		if distance < bestDistance || distance == bestDistance && id < bestID {
			bestID = id
			bestDistance = distance
		}
	}
	return bestID
}

func tradeCenterEndpointRegions(center tradeCenterVisual, routeType world.TradeRouteType, regions map[world.RegionID]*world.Region) []world.RegionID {
	sea := routeType == world.TradeRouteSea
	if center.regionID == "" {
		if id := nearestTradeRegionID(regions, center.worldX, center.worldY, sea); id != "" {
			return []world.RegionID{id}
		}
		return nil
	}
	region := regions[center.regionID]
	if region == nil {
		return nil
	}
	if region.IsSea == sea {
		return []world.RegionID{region.ID}
	}
	if !sea {
		return nil
	}
	result := make([]world.RegionID, 0, len(region.Neighbors))
	for _, neighborID := range region.Neighbors {
		neighbor := regions[neighborID]
		if neighbor != nil && neighbor.IsSea {
			result = append(result, neighborID)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		left := regions[result[i]]
		right := regions[result[j]]
		leftDX := float64(left.WorldX) - center.worldX
		leftDY := float64(left.WorldY) - center.worldY
		rightDX := float64(right.WorldX) - center.worldX
		rightDY := float64(right.WorldY) - center.worldY
		leftDistance := leftDX*leftDX + leftDY*leftDY
		rightDistance := rightDX*rightDX + rightDY*rightDY
		if leftDistance != rightDistance {
			return leftDistance < rightDistance
		}
		return result[i] < result[j]
	})
	return result[:1]
}

func curvedTradeSegmentPoints(start, end tradeOverlayPoint, key string) []tradeOverlayPoint {
	dx := end.x - start.x
	dy := end.y - start.y
	distance := math.Hypot(dx, dy)
	if distance < 0.5 {
		return []tradeOverlayPoint{start, end}
	}
	curve := routeCurveOffset(key, distance) * 0.18
	if math.Abs(curve) < 4 {
		if curve < 0 {
			curve = -4
		} else {
			curve = 4
		}
	}
	midX := (start.x + end.x) / 2
	midY := (start.y + end.y) / 2
	controlX := midX - dy/distance*curve
	controlY := midY + dx/distance*curve
	const samples = 6
	points := make([]tradeOverlayPoint, samples+1)
	for i := 0; i <= samples; i++ {
		t := float64(i) / samples
		points[i].x, points[i].y = quadBezierPoint(start.x, start.y, controlX, controlY, end.x, end.y, t)
	}
	return points
}

func smoothTradePathSegmentPoints(points []tradeOverlayPoint, segment int) []tradeOverlayPoint {
	if segment < 0 || segment+1 >= len(points) {
		return nil
	}
	p0 := points[segment]
	p1 := points[segment]
	if segment > 0 {
		p0 = points[segment-1]
	}
	p2 := points[segment+1]
	p3 := p2
	if segment+2 < len(points) {
		p3 = points[segment+2]
	}

	// Catmull-Rom kontrol noktaları, segment uçlarından geçerken komşu
	// deniz odaklarının dönüş yönünü de tangent olarak taşır.
	c1 := tradeOverlayPoint{
		x: p1.x + (p2.x-p0.x)/6,
		y: p1.y + (p2.y-p0.y)/6,
	}
	c2 := tradeOverlayPoint{
		x: p2.x - (p3.x-p1.x)/6,
		y: p2.y - (p3.y-p1.y)/6,
	}
	const samples = 8
	result := make([]tradeOverlayPoint, samples+1)
	for i := 0; i <= samples; i++ {
		t := float64(i) / samples
		u := 1 - t
		result[i] = tradeOverlayPoint{
			x: u*u*u*p1.x + 3*u*u*t*c1.x + 3*u*t*t*c2.x + t*t*t*p2.x,
			y: u*u*u*p1.y + 3*u*u*t*c1.y + 3*u*t*t*c2.y + t*t*t*p2.y,
		}
	}
	return result
}

func (r *Renderer) tradeCenterLinkPath(from, to tradeCenterVisual, routeType world.TradeRouteType) ([]tradeOverlayPoint, []string) {
	if r == nil || r.gs == nil || routeType != world.TradeRouteSea {
		return nil, nil
	}
	fromCandidates := tradeCenterEndpointRegions(from, routeType, r.gs.Regions)
	toCandidates := tradeCenterEndpointRegions(to, routeType, r.gs.Regions)
	if len(fromCandidates) == 0 || len(toCandidates) == 0 {
		return nil, nil
	}
	sea := routeType == world.TradeRouteSea
	var best []world.RegionID
	for _, fromID := range fromCandidates {
		for _, toID := range toCandidates {
			candidate := shortestTradeRegionPath(r.gs.Regions, fromID, toID, sea)
			if len(candidate) == 0 || len(best) > 0 && len(candidate) >= len(best) {
				continue
			}
			best = candidate
		}
	}
	if len(best) == 0 {
		return nil, nil
	}
	type node struct {
		id    string
		point tradeOverlayPoint
	}
	nodes := make([]node, 0, len(best)+2)
	fromPoint := tradeOverlayPoint{x: from.x, y: from.y}
	if from.regionID != "" {
		if region := r.gs.Regions[from.regionID]; region != nil {
			fromPoint.x, fromPoint.y = r.tradePortScreenPosForSea(region, best[0])
		}
	}
	toPoint := tradeOverlayPoint{x: to.x, y: to.y}
	if to.regionID != "" {
		if region := r.gs.Regions[to.regionID]; region != nil {
			toPoint.x, toPoint.y = r.tradePortScreenPosForSea(region, best[len(best)-1])
		}
	}
	nodes = append(nodes, node{id: "center:" + string(from.id), point: fromPoint})
	for _, regionID := range best {
		region := r.gs.Regions[regionID]
		if region == nil {
			continue
		}
		var x, y float64
		if region.IsSea && r.worldMap != nil {
			// Deniz rotası, raster bölgesinin ortalama anchor'ından değil,
			// Edit Mode'da belirlenen Region.WorldX/WorldY odağından geçer.
			x, y = r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
		} else {
			x, y = r.worldToScreen(float64(region.WorldX), float64(region.WorldY))
		}
		nodes = append(nodes, node{id: "sea:" + string(regionID), point: tradeOverlayPoint{x: x, y: y}})
	}
	nodes = append(nodes, node{id: "center:" + string(to.id), point: toPoint})
	nodePoints := make([]tradeOverlayPoint, len(nodes))
	for i := range nodes {
		nodePoints[i] = nodes[i].point
	}
	points := make([]tradeOverlayPoint, 0, len(nodes)*8)
	segmentKeys := make([]string, 0, len(nodes)*8)
	for i := 0; i < len(nodes)-1; i++ {
		physicalKey := "connector:" + nodes[i].id + "|" + nodes[i+1].id
		if strings.HasPrefix(nodes[i].id, "sea:") && strings.HasPrefix(nodes[i+1].id, "sea:") {
			a, b := nodes[i].id, nodes[i+1].id
			if a > b {
				a, b = b, a
			}
			physicalKey = "sea_segment:" + a + "|" + b
		}
		segmentPoints := smoothTradePathSegmentPoints(nodePoints, i)
		if len(points) == 0 {
			points = append(points, segmentPoints[0])
		}
		for sample := 1; sample < len(segmentPoints); sample++ {
			points = append(points, segmentPoints[sample])
			segmentKeys = append(segmentKeys, physicalKey+"#"+itoa(sample-1))
		}
	}
	if len(points) < 2 {
		return nil, nil
	}
	return points, segmentKeys
}

func tradeCenterLinkTypes(gs *state.GameState, centers []tradeCenterVisual) map[string]world.TradeRouteType {
	types := make(map[string]world.TradeRouteType)
	if gs == nil {
		return types
	}
	indexByID := make(map[world.RegionID]int, len(centers))
	for i, center := range centers {
		indexByID[center.id] = i
	}
	for _, def := range gs.TradeCenters.Centers {
		from, ok := indexByID[def.ID]
		if !ok {
			continue
		}
		for _, link := range def.Links {
			to, ok := indexByID[link.RegionID]
			if !ok || from == to {
				continue
			}
			linkType := link.Type
			if linkType == "" {
				linkType = world.TradeRouteLand
			}
			if def.OffMap {
				linkType = world.TradeRouteLand
			}
			a, b := from, to
			if a > b {
				a, b = b, a
			}
			key := itoa(a) + "|" + itoa(b)
			if def.OffMap || types[key] == "" {
				types[key] = linkType
			}
		}
	}
	return types
}

func tradeCenterSourceLinks(gs *state.GameState, centers []tradeCenterVisual) map[string]bool {
	sources := make(map[string]bool)
	if gs == nil {
		return sources
	}
	indexByID := make(map[world.RegionID]int, len(centers))
	for i, center := range centers {
		indexByID[center.id] = i
	}
	for _, def := range gs.TradeCenters.Centers {
		if !def.OffMap {
			continue
		}
		from, ok := indexByID[def.ID]
		if !ok {
			continue
		}
		for _, link := range def.Links {
			to, ok := indexByID[link.RegionID]
			if !ok || from == to {
				continue
			}
			a, b := from, to
			if a > b {
				a, b = b, a
			}
			sources[itoa(a)+"|"+itoa(b)] = true
		}
	}
	return sources
}

// drawTradeRoutes tüm aktif ticaret rotalarını harita üzerinde sade koridorlar olarak çizer.
// Çift yönlü rotalar (A->B ve B->A) tek bir görsel hatta birleştirilir.
// Uzak zoom'da yalnızca oyuncuyla ilgili rotalar gösterilerek çizgi karmaşası azaltılır.
func (r *Renderer) drawTradeRoutes(screen *ebiten.Image) {
	r.animationTick += 12
	if r.camScale < 0.6 {
		return
	}
	playerID := string(r.gs.PlayerFactionID)
	onlyPlayerRoutes := r.camScale < 0.85

	merged := make(map[string]tradeRouteVisual, len(r.gs.TradeRoutes))
	for _, tr := range r.gs.TradeRoutes {
		if tr == nil || tr.FromFactionID == "" || tr.ToFactionID == "" || tr.FromFactionID == tr.ToFactionID {
			continue
		}
		if onlyPlayerRoutes && tr.FromFactionID != playerID && tr.ToFactionID != playerID {
			continue
		}
		key := tradeRoutePairKey(tr.FromFactionID, tr.ToFactionID)
		route := merged[key]
		if route.factionA == "" {
			if tr.FromFactionID < tr.ToFactionID {
				route.factionA = tr.FromFactionID
				route.factionB = tr.ToFactionID
			} else {
				route.factionA = tr.ToFactionID
				route.factionB = tr.FromFactionID
			}
		}
		route.amount += tr.AmountPerTurn
		if routeKey := tr.AssignmentKey(); routeKey != "" {
			route.routeKeys = append(route.routeKeys, routeKey)
			route.routeDetails = appendTradeCorridorRouteDetail(route.routeDetails, tradeCorridorRouteDetail{
				key:       routeKey,
				direction: factionDisplayName(r.gs, tr.FromFactionID) + " → " + factionDisplayName(r.gs, tr.ToFactionID),
				good:      economy.GoodNameTR(tr.Good),
				amount:    tradeRouteDisplayAmount(tr),
				route:     tr,
			})
		}
		candidateGood := economy.GoodNameTR(tr.Good)
		if route.goodName == "" || tr.AmountPerTurn > route.bestFlow {
			route.goodName = candidateGood
			route.bestFlow = tr.AmountPerTurn
			route.route = tr
		}
		merged[key] = route
	}
	r.tradeCorridors = r.tradeCorridors[:0]
	centers := r.buildTradeCenters(len(r.gs.TradeCenters.Centers))
	r.tradeCenters = append(r.tradeCenters[:0], centers...)
	r.drawPlayerTradePortRoutes(screen, merged)
	if len(centers) == 0 {
		r.drawTradeBonusFleetMarkers(screen)
		r.tradeHoverIdx = -1
		r.tradeCenterIdx = -1
		r.updateTradeHover()
		return
	}
	mx, my := ebiten.CursorPosition()
	preFocusCenter := -1
	bestD := 13.0
	for i := range centers {
		d := math.Hypot(float64(mx)-centers[i].x, float64(my)-centers[i].y)
		if d < bestD {
			bestD = d
			preFocusCenter = i
		}
	}
	adj := r.buildTradeCenterAdjacency(centers)
	centerIndexByID := make(map[world.RegionID]int, len(centers))
	for i, center := range centers {
		centerIndexByID[center.id] = i
	}
	linkTypes := tradeCenterLinkTypes(r.gs, centers)
	sourceLinks := tradeCenterSourceLinks(r.gs, centers)
	visualDirections := tradeCenterVisualDirections(r.gs, centers)
	factionHub := make(map[string]*world.Region, len(merged)*2)
	factionCenter := make(map[string]int, len(merged)*2)
	type linkAgg struct {
		flow         int
		factions     map[string]struct{}
		goods        map[string]int
		directions   map[struct{ from, to int }]int
		routeKeys    []string
		routeDetails []tradeCorridorRouteDetail
		historical   bool
	}
	centerLinkFlow := map[string]*linkAgg{}
	mergedKeys := make([]string, 0, len(merged))
	for key := range merged {
		mergedKeys = append(mergedKeys, key)
	}
	sort.Strings(mergedKeys)
	for _, key := range mergedKeys {
		route := merged[key]
		routeAmount := tradeRouteDisplayAmount(route.route)
		routeDetails := route.routeDetails
		if len(routeDetails) == 0 && route.route != nil {
			routeDetails = []tradeCorridorRouteDetail{{
				key: key, direction: factionDisplayName(r.gs, route.factionA) + " → " + factionDisplayName(r.gs, route.factionB),
				good: route.goodName, amount: routeAmount, route: route.route,
			}}
		}
		if factionHub[route.factionA] == nil {
			factionHub[route.factionA] = r.factionPrimaryRegion(route.factionA)
		}
		if factionHub[route.factionB] == nil {
			factionHub[route.factionB] = r.factionPrimaryRegion(route.factionB)
		}
		ca, ok := factionCenter[route.factionA]
		if !ok {
			ca = r.nearestTradeCenterIndex(factionHub[route.factionA], centers)
			factionCenter[route.factionA] = ca
		}
		cb, ok := factionCenter[route.factionB]
		if !ok {
			cb = r.nearestTradeCenterIndex(factionHub[route.factionB], centers)
			factionCenter[route.factionB] = cb
		}
		if ca < 0 || cb < 0 || ca == cb {
			continue
		}
		path := shortestCenterPathForTradeRoute(r.gs, adj, centers, ca, cb, route.factionA, route.factionB)
		if len(path) < 2 {
			continue
		}
		for pi := 0; pi < len(path)-1; pi++ {
			ka, kb := path[pi], path[pi+1]
			flowFrom, flowTo := ka, kb
			if ka > kb {
				ka, kb = kb, ka
			}
			key := itoa(ka) + "|" + itoa(kb)
			agg := centerLinkFlow[key]
			if agg == nil {
				agg = &linkAgg{
					factions:   make(map[string]struct{}, 4),
					goods:      make(map[string]int, 4),
					directions: make(map[struct{ from, to int }]int, 2),
				}
				centerLinkFlow[key] = agg
			}
			agg.flow += route.amount
			agg.directions[struct{ from, to int }{from: flowFrom, to: flowTo}] += route.amount
			agg.factions[route.factionA] = struct{}{}
			agg.factions[route.factionB] = struct{}{}
			agg.routeKeys = append(agg.routeKeys, route.routeKeys...)
			agg.routeDetails = appendTradeCorridorRouteDetails(agg.routeDetails, routeDetails...)
			if route.goodName != "" {
				agg.goods[route.goodName] += route.amount
			}
		}
	}

	// Senaryo tarafından tanımlanan tarihsel akışlar, diplomatik faction
	// rotası değildir; ancak aynı merkez grafiğinde gerçek hacim olarak görünür.
	for _, flow := range r.gs.ActiveHistoricalTradeFlows() {
		from, fromOK := centerIndexByID[flow.FromRegionID]
		to, toOK := centerIndexByID[flow.ToRegionID]
		if !fromOK || !toOK || from == to {
			continue
		}
		path := shortestCenterPath(adj, from, to)
		if len(path) < 2 {
			continue
		}
		for pi := 0; pi < len(path)-1; pi++ {
			a, b := path[pi], path[pi+1]
			flowFrom, flowTo := a, b
			if a > b {
				a, b = b, a
			}
			key := itoa(a) + "|" + itoa(b)
			agg := centerLinkFlow[key]
			if agg == nil {
				agg = &linkAgg{
					factions:   make(map[string]struct{}, 1),
					goods:      make(map[string]int, 2),
					directions: make(map[struct{ from, to int }]int, 2),
				}
				centerLinkFlow[key] = agg
			}
			amount := r.gs.HistoricalTradeFlowAmount(flow)
			if amount <= 0 {
				continue
			}
			agg.flow += amount
			agg.directions[struct{ from, to int }{from: flowFrom, to: flowTo}] += amount
			agg.goods[economy.GoodNameTR(flow.Good)] += amount
			agg.routeDetails = appendTradeCorridorRouteDetail(agg.routeDetails, tradeCorridorRouteDetail{
				key: "historical", good: economy.GoodNameTR(flow.Good), amount: amount, historical: true,
			})
			agg.historical = true
		}
	}

	// Başkent -> en yakın ticaret merkezi bağlantıları. Bunlar işlem hacminden
	// bağımsız, her devlet için sabit ve ince görünür; merkezi grafiği büyütmez.
	if r.camScale >= 0.95 {
		factionIDs := make([]string, 0, len(r.gs.Factions))
		for factionID, currentFaction := range r.gs.Factions {
			if currentFaction != nil && !currentFaction.IsEliminated && !currentFaction.IsVirtual {
				factionIDs = append(factionIDs, string(factionID))
			}
		}
		sort.Strings(factionIDs)
		for _, fid := range factionIDs {
			hub := r.factionPrimaryRegion(fid)
			if hub == nil {
				continue
			}
			centerIdx := r.nearestTradeCenterIndex(hub, centers)
			if centerIdx < 0 || centerIdx >= len(centers) {
				continue
			}
			hx, hy := r.regionScreenPos(hub)
			c := centers[centerIdx]
			if hub.ID == c.regionID {
				continue
			}
			col := color.RGBA{168, 192, 220, 72}
			if preFocusCenter >= 0 && centerIdx != preFocusCenter {
				col = color.RGBA{120, 135, 160, 28}
			}
			if !r.tradeOverlayOccludesSegment(hx, hy, c.x, c.y) {
				vector.StrokeLine(screen, float32(hx), float32(hy), float32(c.x), float32(c.y), 0.75, col, false)
			}
		}
	}

	// Trade center <-> trade center corridors (ana ağ)
	linkKeySet := make(map[string]struct{}, len(centerLinkFlow))
	for key := range centerLinkFlow {
		linkKeySet[key] = struct{}{}
	}
	for fromIdx, list := range adj {
		for _, toIdx := range list {
			a, b := fromIdx, toIdx
			if a > b {
				a, b = b, a
			}
			linkKeySet[itoa(a)+"|"+itoa(b)] = struct{}{}
		}
	}
	// Pasif merkezlerin bağlantıları da haritada görünür; bunlar ağ çözümüne
	// katılmadığı için yalnızca görsel koridor olarak eklenir.
	indexByID := make(map[world.RegionID]int, len(centers))
	for idx, center := range centers {
		indexByID[center.id] = idx
	}
	for _, def := range r.gs.TradeCenters.Centers {
		from, ok := indexByID[def.ID]
		if !ok {
			continue
		}
		for _, link := range def.Links {
			linkedID := link.RegionID
			to, ok := indexByID[linkedID]
			if !ok || from == to {
				continue
			}
			a, b := from, to
			if a > b {
				a, b = b, a
			}
			linkKeySet[itoa(a)+"|"+itoa(b)] = struct{}{}
		}
	}
	linkKeys := make([]string, 0, len(linkKeySet))
	for key := range linkKeySet {
		linkKeys = append(linkKeys, key)
	}
	// Ortak bir deniz segmenti birden fazla merkez bağlantısının parçası
	// olabilir. En güçlü/aktif bağlantıyı önce çizerek o fiziksel segmentin
	// görünümünü belirlemesini sağlarız; aşağıda aynı segment tekrar çizilmez.
	sort.SliceStable(linkKeys, func(a, b int) bool {
		flowFor := func(key string) (int, bool) {
			agg := centerLinkFlow[key]
			amount := 0
			if agg != nil {
				amount = agg.flow
			}
			parts := strings.Split(key, "|")
			if len(parts) != 2 {
				return amount, false
			}
			i, errI := strconv.Atoi(parts[0])
			j, errJ := strconv.Atoi(parts[1])
			if errI != nil || errJ != nil || i < 0 || j < 0 || i >= len(centers) || j >= len(centers) {
				return amount, false
			}
			return amount, centers[i].active && centers[j].active
		}
		amountA, activeA := flowFor(linkKeys[a])
		amountB, activeB := flowFor(linkKeys[b])
		if amountA != amountB {
			return amountA > amountB
		}
		if activeA != activeB {
			return activeA
		}
		return linkKeys[a] < linkKeys[b]
	})
	corridorPaths := make(map[string][]tradeOverlayPoint, len(linkKeys))
	corridorPathKeys := make(map[string][]string, len(linkKeys))
	seaFocusMarkers := make([]tradeOverlayPoint, 0, len(linkKeys)*2)
	seaFocusByCenter := make(map[int]tradeOverlayPoint, len(centers))
	for _, key := range linkKeys {
		if linkTypes[key] != world.TradeRouteSea {
			continue
		}
		parts := strings.Split(key, "|")
		if len(parts) != 2 {
			continue
		}
		i, errI := strconv.Atoi(parts[0])
		j, errJ := strconv.Atoi(parts[1])
		if errI != nil || errJ != nil || i < 0 || j < 0 || i >= len(centers) || j >= len(centers) || i == j {
			continue
		}
		corridorPaths[key], corridorPathKeys[key] = r.tradeCenterLinkPath(centers[i], centers[j], world.TradeRouteSea)
		focusPoints := tradeSeaFocusPoints(corridorPaths[key], corridorPathKeys[key])
		if len(focusPoints) > 0 {
			seaFocusByCenter[i] = focusPoints[0]
			seaFocusByCenter[j] = focusPoints[len(focusPoints)-1]
		}
		for _, point := range focusPoints {
			duplicate := false
			for _, existing := range seaFocusMarkers {
				if existing.x == point.x && existing.y == point.y {
					duplicate = true
					break
				}
			}
			if !duplicate {
				seaFocusMarkers = append(seaFocusMarkers, point)
			}
		}
	}
	drawnPathSegments := make(map[string]struct{})
	pathCorridorIndex := make(map[string]int)
	for _, key := range linkKeys {
		agg := centerLinkFlow[key]
		parts := strings.Split(key, "|")
		if len(parts) != 2 {
			continue
		}
		i, errI := strconv.Atoi(parts[0])
		j, errJ := strconv.Atoi(parts[1])
		if errI != nil || errJ != nil {
			continue
		}
		if i < 0 || j < 0 || i >= len(centers) || j >= len(centers) || i == j {
			continue
		}
		amount := 0
		if agg != nil {
			amount = agg.flow
		}
		sx, sy := centers[i].x, centers[i].y
		dx, dy := centers[j].x, centers[j].y
		mx := (sx + dx) / 2
		my := (sy + dy) / 2
		vx := dx - sx
		vy := dy - sy
		dist := math.Hypot(vx, vy)
		if dist < 1 {
			continue
		}
		px := -vy / dist
		py := vx / dist
		curve := routeCurveOffset(key, dist)
		cx := mx + px*curve
		cy := my + py*curve
		corridorPath := corridorPaths[key]
		corridorSegmentKeys := corridorPathKeys[key]
		pathSegments := splitTradePhysicalPath(corridorPath, corridorSegmentKeys)
		arrowPath := corridorPath
		if linkTypes[key] == world.TradeRouteSea {
			arrowPath = tradeSeaMainPath(corridorPath, corridorSegmentKeys)
		}
		pointAt := func(t float64) (float64, float64) {
			if len(corridorPath) >= 2 {
				return tradePolylinePoint(corridorPath, t)
			}
			return quadBezierPoint(sx, sy, cx, cy, dx, dy, t)
		}

		idleLink := amount == 0
		passiveLink := !centers[i].active || !centers[j].active
		glow, core, routeArrowColor := tradeRoutePalette(linkTypes[key])
		if sourceLinks[key] {
			glow, core, routeArrowColor = tradeSourceRoutePalette()
		}
		glow.A = 18
		core.A = 42
		coreW := float32(1.0)
		glowW := float32(2.8)
		if idleLink {
			glow = color.RGBA{}
			if sourceLinks[key] {
				core = color.RGBA{58, 112, 78, 150}
			} else if linkTypes[key] == world.TradeRouteSea {
				core = color.RGBA{86, 154, 190, 150}
			} else {
				core = color.RGBA{160, 139, 102, 150}
			}
			coreW = 0.9
			glowW = 0
		}
		if passiveLink {
			glow, core, routeArrowColor = passiveTradeRoutePalette()
			coreW = 2.0
			glowW = 0
		}
		if amount > 0 && !passiveLink {
			alphaScale := min(uint8(80+(amount*12)), 255)
			glow.A = alphaScale
			core.A = alphaScale
			coreW = 1.5
			glowW = 5.0
			if amount >= 14 {
				coreW = 2.1
				glowW = 7.0
			} else if amount >= 8 {
				coreW = 1.8
				glowW = 6.0
			}
		}

		if preFocusCenter >= 0 && i != preFocusCenter && j != preFocusCenter && !passiveLink {
			if amount > 0 {
				glow = routeArrowColor
				glow.A = 10
				core = routeArrowColor
				core.A = 34
				coreW = 1.1
				glowW = 3.4
			} else {
				core = routeArrowColor
				core.A = 46
			}
		}
		if len(corridorPath) >= 2 {
			// Aynı deniz bölgesi kenarı birden fazla rota tarafından
			// kullanılıyorsa yalnızca ilk rotanın çizgisi görünür. Her fiziksel
			// kenarın örnek noktaları birlikte ele alındığı için kıvrım da korunur.
			for segment := 0; segment < len(corridorPath)-1; {
				baseKey := tradePathSegmentKey(corridorSegmentKeys[segment])
				end := segment + 1
				for end < len(corridorPath)-1 && end < len(corridorSegmentKeys) {
					endKey := tradePathSegmentKey(corridorSegmentKeys[end])
					if endKey != baseKey {
						break
					}
					end++
				}
				if _, drawn := drawnPathSegments[baseKey]; drawn {
					segment = end
					continue
				}
				drawnPathSegments[baseKey] = struct{}{}
				if strings.HasPrefix(baseKey, "connector:") {
					start := corridorPath[segment]
					endPoint := corridorPath[end]
					drawDashedTradeLine(screen, start.x, start.y, endPoint.x, endPoint.y, coreW, tradeSeaConnectorColor, r.tradeOverlayOccludesSegment)
					segment = end
					continue
				}
				for sub := segment; sub < end; sub++ {
					if passiveLink && sub%3 == 2 {
						continue
					}
					x1, y1 := corridorPath[sub].x, corridorPath[sub].y
					x2, y2 := corridorPath[sub+1].x, corridorPath[sub+1].y
					if r.tradeOverlayOccludesSegment(x1, y1, x2, y2) {
						continue
					}
					drawGlowW, drawCoreW := glowW, coreW
					drawGlow, drawCore := glow, core
					if drawGlowW > 0 {
						vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), drawGlowW, drawGlow, false)
					}
					vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), drawCoreW, drawCore, false)
				}
				segment = end
			}
		} else {
			segments := 36
			for segment := 0; segment < segments; segment++ {
				if passiveLink && segment%3 == 2 {
					continue
				}
				t1 := float64(segment) / float64(segments)
				t2 := float64(segment+1) / float64(segments)
				x1, y1 := pointAt(t1)
				x2, y2 := pointAt(t2)
				if r.tradeOverlayOccludesSegment(x1, y1, x2, y2) {
					continue
				}
				if glowW > 0 {
					vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), glowW, glow, false)
				}
				vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), coreW, core, false)
			}
		}
		if agg != nil && amount > 0 {
			forward := struct{ from, to int }{from: i, to: j}
			reverse := struct{ from, to int }{from: j, to: i}
			hasReverse := agg.directions[reverse] > 0
			for direction, directionAmount := range agg.directions {
				if directionAmount <= 0 || (direction != forward && direction != reverse) {
					continue
				}
				if directions := visualDirections[key]; len(directions) > 0 {
					if _, ok := directions[direction]; !ok {
						continue
					}
				}
				isReverse := direction == reverse
				arrowT := 0.5
				if hasReverse {
					if isReverse {
						arrowT = 0.58
					} else {
						arrowT = 0.42
					}
				}
				arrowColor := routeArrowColor
				arrowColor.A = 245
				if len(arrowPath) >= 2 {
					drawTradeFlowArrowOnPath(screen, arrowPath, arrowT, isReverse, arrowColor)
				} else {
					drawTradeFlowArrow(screen, sx, sy, cx, cy, dx, dy, arrowT, isReverse, arrowColor)
				}
			}
		}
		goodsList := make([]struct {
			name string
			flow int
		}, 0)
		if agg != nil {
			goodsList = make([]struct {
				name string
				flow int
			}, 0, len(agg.goods))
			for name, flow := range agg.goods {
				goodsList = append(goodsList, struct {
					name string
					flow int
				}{name: name, flow: flow})
			}
		}
		sort.Slice(goodsList, func(a, b int) bool {
			if goodsList[a].flow != goodsList[b].flow {
				return goodsList[a].flow > goodsList[b].flow
			}
			return goodsList[a].name < goodsList[b].name
		})
		goodsSummary := "-"
		if len(goodsList) > 0 {
			goodsSummary = goodsList[0].name
			if len(goodsList) > 1 {
				goodsSummary += ", " + goodsList[1].name
			}
		}
		factionCount := 0
		if agg != nil {
			factionCount = len(agg.factions)
		}
		directionText := centers[i].nameTR + " ↔ " + centers[j].nameTR
		if agg != nil {
			forward := agg.directions[struct{ from, to int }{from: i, to: j}]
			reverse := agg.directions[struct{ from, to int }{from: j, to: i}]
			switch {
			case forward > 0 && reverse == 0:
				directionText = centers[i].nameTR + " → " + centers[j].nameTR
			case reverse > 0 && forward == 0:
				directionText = centers[j].nameTR + " → " + centers[i].nameTR
			case forward > 0 && reverse > 0:
				directionText = centers[i].nameTR + " ↔ " + centers[j].nameTR
			}
		}
		corridor := tradeCorridorInfo{
			fromName:      centers[i].nameTR,
			toName:        centers[j].nameTR,
			directionText: directionText,
			amount:        amount,
			factions:      factionCount,
			goods:         goodsSummary,
			sx:            sx,
			sy:            sy,
			cx:            cx,
			cy:            cy,
			dx:            dx,
			dy:            dy,
			path:          corridorPath,
			routeType:     linkTypes[key],
			hitWidth:      float64(glowW) + 4,
			dashed:        false,
			historical:    agg != nil && agg.historical,
			routeKeys: func() []string {
				if agg == nil {
					return nil
				}
				return agg.routeKeys
			}(),
			centerFactions: [2]string{
				tradeCenterOwnerFaction(r.gs, centers[i]),
				tradeCenterOwnerFaction(r.gs, centers[j]),
			},
		}
		if agg != nil {
			corridor.routeDetails = append(corridor.routeDetails, agg.routeDetails...)
		}
		if len(pathSegments) == 0 {
			r.tradeCorridors = append(r.tradeCorridors, corridor)
			continue
		}
		for _, pathSegment := range pathSegments {
			if pathSegment.key == "" || len(pathSegment.points) < 2 {
				continue
			}
			if strings.HasPrefix(pathSegment.key, "connector:") {
				continue
			}
			segmentCorridor := corridor
			// Bu fiziksel segment, rotanın gerçek yolunun parçasıdır. Uç
			// merkezlerden biri üçüncü bir faction'a ait olsa bile segmenti
			// kullanan rota ayrıntıları tooltip'te görünmelidir.
			segmentCorridor.showAllRouteDetails = true
			segmentCorridor.path = pathSegment.points
			segmentCorridor.sx = pathSegment.points[0].x
			segmentCorridor.sy = pathSegment.points[0].y
			last := pathSegment.points[len(pathSegment.points)-1]
			segmentCorridor.dx = last.x
			segmentCorridor.dy = last.y
			segmentCorridor.cx = (segmentCorridor.sx + segmentCorridor.dx) / 2
			segmentCorridor.cy = (segmentCorridor.sy + segmentCorridor.dy) / 2
			if existing, ok := pathCorridorIndex[pathSegment.key]; ok {
				// Ortak fiziksel segment daha önce başka bir merkez rotası
				// tarafından oluşturulmuş olabilir. Transit rota ayrıntıları
				// için segmentin ortak-koridor davranışını koru.
				r.tradeCorridors[existing].showAllRouteDetails = true
				r.tradeCorridors[existing].routeDetails = appendTradeCorridorRouteDetails(r.tradeCorridors[existing].routeDetails, segmentCorridor.routeDetails...)
				r.tradeCorridors[existing].routeKeys = appendTradeRouteKeys(r.tradeCorridors[existing].routeKeys, segmentCorridor.routeKeys...)
				continue
			}
			pathCorridorIndex[pathSegment.key] = len(r.tradeCorridors)
			r.tradeCorridors = append(r.tradeCorridors, segmentCorridor)
		}
	}
	for index := range r.tradeCorridors {
		sortTradeCorridorRouteDetails(r.tradeCorridors[index].routeDetails)
	}
	drawTradeSeaFocusMarkers(screen, seaFocusMarkers)
	r.updateTradeHover()

	focusCenter := r.tradeCenterIdx
	if r.tradeHoverIdx >= 0 && r.tradeHoverIdx < len(r.tradeCorridors) {
		c := r.tradeCorridors[r.tradeHoverIdx]
		for i := range centers {
			if centers[i].nameTR == c.fromName || centers[i].nameTR == c.toName {
				focusCenter = i
				break
			}
		}
	}

	// Compute active volume for each trade center (local capacity + trade route activity)
	centerVolume := make([]int, len(centers))
	for idx, c := range centers {
		vol := 0
		reg := r.gs.Regions[c.regionID]
		if reg != nil && !c.offMap {
			vol = r.gs.TradeCenterVolume(reg)
		}
		centerVolume[idx] = vol
	}

	for i := range centers {
		primary := centers[i].tier == world.TradeCenterPrimary && !centers[i].offMap
		mainRoute := centers[i].mainRoute
		alphaBg := uint8(170)
		alphaBorder := uint8(155)
		alphaText := uint8(205)
		if primary {
			alphaBg = 238
			alphaBorder = 242
			alphaText = 255
		}
		if mainRoute {
			alphaBg = 220
			alphaBorder = 245
			alphaText = 245
		}
		if !centers[i].active {
			alphaBg = 78
			alphaBorder = 92
			alphaText = 135
		}
		if focusCenter >= 0 {
			isFocus := false
			if i == focusCenter {
				isFocus = true
			} else {
				for _, c := range r.tradeCorridors {
					if c.fromName == centers[i].nameTR || c.toName == centers[i].nameTR {
						if c.fromName == centers[focusCenter].nameTR || c.toName == centers[focusCenter].nameTR {
							isFocus = true
							break
						}
					}
				}
			}
			if !isFocus {
				alphaBg = 100
				alphaBorder = 100
				alphaText = 140
			}
		}

		nameFace := FaceSmall
		if primary {
			nameFace = FaceMed
		}
		nameW := float32(MeasureText(centers[i].nameTR, nameFace))
		contentW := nameW
		inactiveStatus := ""
		inactiveGoods := ""
		if !centers[i].active {
			if centers[i].unlockYear > 0 {
				inactiveStatus = "Pasif - " + itoa(centers[i].unlockYear) + " yılında açılır"
			} else {
				inactiveStatus = "Pasif"
			}
			inactiveGoods = "Mal: " + historicalTradeGoodsLabel(centers[i].sourceGoods) + " | 0/tur"
			for _, text := range []string{inactiveStatus, inactiveGoods} {
				textW := float32(MeasureText(text, FaceSmall))
				if textW > contentW {
					contentW = textW
				}
			}
		}
		volStr := ""
		if !centers[i].offMap {
			capacityBonus, incomeBonus := r.tradeCenterBenefits(centers[i])
			volStr = "Hacim: " + itoa(centerVolume[i])
			if capacityBonus != 0 || incomeBonus != 0 {
				volStr += " | +" + itoa(capacityBonus) + " kap. | +" + itoa(incomeBonus) + " altın"
			}
			volW := float32(MeasureText(volStr, FaceSmall))
			if volW > contentW {
				contentW = volW
			}
		}
		w := contentW + 28
		h := float32(36)
		if primary {
			w = contentW + primaryTradeIconBox + 30
			if w < 178 {
				w = 178
			}
			h = 54
		} else if !centers[i].active {
			w = contentW + 28
			h = 48
		} else if mainRoute {
			w = contentW + 34
			h = 26
		} else if w < 102 {
			w = 102
		}
		if centers[i].offMap && !mainRoute && centers[i].active {
			h = 22
		}
		x := float32(centers[i].x) - w/2
		y := float32(centers[i].y) - tradeCenterLabelGap - h
		if focus, ok := seaFocusByCenter[i]; ok {
			// Deniz odağı marker'ı donanma marker'ı yarıçapı olan 13 px'tir;
			// liman tabelası odak noktasının 36 px üstünde kalır.
			x = float32(focus.x) - w/2
			y = float32(focus.y) - tradeCenterLabelGap - tradeSeaCenterLabelExtraGap - h
		} else if centers[i].landFocus {
			// Kara rotasında odak, liman değil merkez yerleşim marker'ıdır.
			y = float32(centers[i].y) - tradeCenterLabelGap - h
		}
		labelRect := gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}
		centers[i].labelX = labelRect.X
		centers[i].labelY = labelRect.Y
		centers[i].labelW = labelRect.W
		centers[i].labelH = labelRect.H
		r.tradeCenters[i].labelX = labelRect.X
		r.tradeCenters[i].labelY = labelRect.Y
		r.tradeCenters[i].labelW = labelRect.W
		r.tradeCenters[i].labelH = labelRect.H
		if topStatusPanelHit(labelRect.X+labelRect.W/2, labelRect.Y+labelRect.H/2) ||
			topDateHudHit(labelRect.X+labelRect.W/2, labelRect.Y+labelRect.H/2) ||
			musicHudHit(labelRect.X+labelRect.W/2, labelRect.Y+labelRect.H/2) ||
			bottomActionHudHit(labelRect.X+labelRect.W/2, labelRect.Y+labelRect.H/2) ||
			eventLogPanelHit(labelRect.X+labelRect.W/2, labelRect.Y+labelRect.H/2, r.eventLogCollapsed) ||
			minimapHit(labelRect.X+labelRect.W/2, labelRect.Y+labelRect.H/2) {
			continue
		}

		// semi-transparent dark wood background
		bgColor := color.RGBA{18, 14, 10, alphaBg}
		if centers[i].offMap {
			bgColor = color.RGBA{14, 18, 22, alphaBg}
		} else if !primary {
			bgColor = color.RGBA{16, 13, 10, alphaBg}
		}
		vector.FillRect(screen, x, y, w, h, bgColor, false)

		// Border: gold for primary, bronze for secondary
		borderColor := color.RGBA{122, 101, 75, alphaBorder}
		if mainRoute {
			borderColor = color.RGBA{232, 180, 70, alphaBorder}
		}
		if !centers[i].active {
			borderColor = color.RGBA{120, 130, 140, alphaBorder}
			if mainRoute {
				borderColor = color.RGBA{170, 145, 80, alphaBorder}
			}
		}
		if primary {
			vector.StrokeRect(screen, x-1, y-1, w+2, h+2, 1.2, color.RGBA{235, 200, 110, alphaBorder}, false)
			vector.StrokeRect(screen, x+1, y+1, w-2, h-2, 0.8, color.RGBA{150, 110, 50, alphaBorder}, false)
		} else {
			if centers[i].offMap {
				borderColor = color.RGBA{118, 156, 188, alphaBorder}
			}
		}
		vector.StrokeRect(screen, x, y, w, h, 1.0, borderColor, false)
		if centers[i].endNode {
			vector.StrokeRect(screen, x-2, y-2, w+4, h+4, 1.0, color.RGBA{210, 70, 70, alphaBorder}, false)
		}

		// Center Name
		nameCol := color.RGBA{204, 196, 172, alphaText}
		if primary {
			nameCol = color.RGBA{255, 235, 170, alphaText}
		}
		if centers[i].offMap {
			nameCol = color.RGBA{210, 228, 245, alphaText}
		}
		if mainRoute {
			nameCol = color.RGBA{255, 225, 130, alphaText}
		}
		textX := float64(x) + 14
		nameY := float64(y) + 4
		volumeY := float64(y) + 20
		if primary {
			iconX := x + 6
			iconY := y + (h-primaryTradeIconBox)/2
			vector.FillRect(screen, iconX, iconY, primaryTradeIconBox, primaryTradeIconBox, color.RGBA{44, 34, 19, alphaBg}, false)
			vector.StrokeRect(screen, iconX, iconY, primaryTradeIconBox, primaryTradeIconBox, 1, color.RGBA{224, 190, 102, alphaBorder}, false)
			r.drawSettlementLabelSprite(screen, tradeCenterIcon, iconX+(primaryTradeIconBox-primaryTradeIconSize)/2, iconY+(primaryTradeIconBox-primaryTradeIconSize)/2, primaryTradeIconSize)
			textX = float64(iconX + primaryTradeIconBox + 10)
			nameY = float64(y) + 9
			volumeY = float64(y) + 30
		}
		DrawText(screen, centers[i].nameTR, textX, nameY, nameFace, nameCol)
		if !centers[i].active {
			DrawText(screen, inactiveStatus, textX, float64(y)+19, FaceSmall, color.RGBA{180, 184, 190, alphaText})
			DrawText(screen, inactiveGoods, textX, float64(y)+34, FaceSmall, color.RGBA{170, 174, 180, alphaText})
		}

		// Volume indicator
		if !centers[i].offMap {
			DrawText(screen, volStr, textX, volumeY, FaceSmall, color.RGBA{170, 170, 160, alphaText})
		}
	}
	r.drawTradeBonusFleetMarkers(screen)
}

// factionPrimaryRegion bir fraksiyonun görsel temsili için başkent bölgesini,
// başkent tanımlı değilse deterministik ekonomik fallback bölgesini döner.
func (r *Renderer) factionPrimaryRegion(factionID string) *world.Region {
	if r == nil || r.gs == nil {
		return nil
	}
	if capital, _, _, ok := r.gs.FactionCapital(faction.FactionID(factionID)); ok && capital != nil {
		return capital
	}
	candidates := make([]*world.Region, 0, 16)
	if len(r.gs.RegionOrder) > 0 {
		for _, rid := range r.gs.RegionOrder {
			region := r.gs.Regions[rid]
			if region == nil || region.OwnerID != factionID || region.IsSea {
				continue
			}
			candidates = append(candidates, region)
		}
	} else {
		ids := make([]string, 0, len(r.gs.Regions))
		for rid := range r.gs.Regions {
			ids = append(ids, string(rid))
		}
		sort.Strings(ids)
		for _, id := range ids {
			region := r.gs.Regions[world.RegionID(id)]
			if region == nil || region.OwnerID != factionID || region.IsSea {
				continue
			}
			candidates = append(candidates, region)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	best := candidates[0]
	bestCapital := false
	bestScore := -1
	for _, region := range candidates {
		capital := false
		for _, settlement := range region.Settlements {
			if settlement.IsCenter {
				capital = true
				break
			}
		}
		if capital && !bestCapital {
			best = region
			bestCapital = true
			bestScore = r.gs.EffectiveRegionTradeCapacity(region)
			continue
		}
		if capital == bestCapital && r.gs.EffectiveRegionTradeCapacity(region) > bestScore {
			best = region
			bestScore = r.gs.EffectiveRegionTradeCapacity(region)
		}
	}
	return best
}
