package render

import (
	"image/color"
	"math"
	"sort"
	"strconv"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type tradeRouteVisual struct {
	factionA  string
	factionB  string
	goodName  string
	amount    int
	bestFlow  int
	route     *economy.TradeRoute
	routeKeys []string
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
	networkOnly bool
}

type tradeCorridorInfo struct {
	fromName  string
	toName    string
	amount    int
	factions  int
	goods     string
	sx        float64
	sy        float64
	cx        float64
	cy        float64
	dx        float64
	dy        float64
	hitWidth  float64
	dashed    bool
	route     *economy.TradeRoute
	routeKeys []string
}

var (
	playerTradeRouteColor = color.RGBA{255, 145, 42, 235}
)

const (
	tradeRouteDashLength = 12.0
	tradeRouteGapLength  = 10.0
	tradeRouteDashParts  = 72
)

func tradeRouteDisplayAmount(route *economy.TradeRoute) int {
	if route == nil || route.SuspendedTurns > 0 {
		return 0
	}
	return route.EffectiveAmountPerTurn()
}

func tradeCorridorTooltipHeight(c tradeCorridorInfo) float64 {
	if c.dashed && c.route != nil {
		return 142
	}
	return 90
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
		if !def.ActiveInYear(r.gs.Year) {
			continue
		}
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
				networkOnly: def.NetworkOnly,
			})
			continue
		}
		reg := r.gs.Regions[def.ID]
		if reg == nil || reg.IsSea || !def.NetworkOnly && reg.TradeCapacity <= 0 {
			continue
		}
		sx, sy := r.regionScreenPos(reg)
		centers = append(centers, tradeCenterVisual{
			id:          reg.ID,
			regionID:    reg.ID,
			nameTR:      chooseRegionLabel(reg),
			tier:        def.Tier,
			worldX:      float64(reg.WorldX),
			worldY:      float64(reg.WorldY),
			x:           sx,
			y:           sy,
			networkOnly: def.NetworkOnly,
		})
	}
	return centers
}

func (r *Renderer) tradeCenterBenefits(center tradeCenterVisual) (capacityBonus, incomeBonus int) {
	if r == nil || r.gs == nil || center.offMap || center.regionID == "" {
		return 0, 0
	}
	return r.gs.TradeCenterBenefits(r.gs.Regions[center.regionID])
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
		prevX, prevY := quadBezierPoint(c.sx, c.sy, c.cx, c.cy, c.dx, c.dy, 0)
		for s := 1; s <= segments; s++ {
			t := float64(s) / float64(segments)
			x, y := quadBezierPoint(c.sx, c.sy, c.cx, c.cy, c.dx, c.dy, t)
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
		segments := 28
		for i := 0; i < segments; i++ {
			t1 := float64(i) / float64(segments)
			t2 := float64(i+1) / float64(segments)
			x1, y1 := quadBezierPoint(c.sx, c.sy, c.cx, c.cy, c.dx, c.dy, t1)
			x2, y2 := quadBezierPoint(c.sx, c.sy, c.cx, c.cy, c.dx, c.dy, t2)
			if r.tradeOverlayOccludesSegment(x1, y1, x2, y2) {
				continue
			}
			vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 9.0, color.RGBA{255, 228, 144, 56}, false)
			vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 3.0, color.RGBA{255, 241, 192, 230}, false)
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
	DrawText(screen, "Ticaret Koridoru", float64(x)+10, float64(y)+8, FaceSmall, color.RGBA{242, 226, 174, 255})
	DrawText(screen, c.fromName+" ↔ "+c.toName, float64(x)+10, float64(y)+28, FaceSmall, color.RGBA{215, 225, 236, 235})
	if c.dashed && c.route != nil && r.gs != nil {
		amount := tradeRouteDisplayAmount(c.route)
		gold := amount * c.route.GoldPerUnit
		goodName := economy.GoodNameTR(c.route.Good)
		playerID := string(r.gs.PlayerFactionID)
		DrawText(screen, "Hacim: "+itoa(amount)+"/tur   Emtia: "+goodName, float64(x)+10, float64(y)+46, FaceSmall, color.RGBA{187, 203, 222, 230})
		if c.route.FromFactionID == playerID {
			DrawText(screen, "Veriyoruz: "+goodName+" "+itoa(amount)+"/tur", float64(x)+10, float64(y)+64, FaceSmall, color.RGBA{225, 205, 170, 240})
			DrawText(screen, "Alıyoruz: Altın +"+itoa(gold)+"/tur", float64(x)+10, float64(y)+82, FaceSmall, color.RGBA{225, 205, 170, 240})
			DrawText(screen, "Gelir: +"+itoa(gold)+" altın/tur", float64(x)+10, float64(y)+100, FaceSmall, color.RGBA{145, 220, 155, 245})
		} else {
			DrawText(screen, "Veriyoruz: Altın "+itoa(gold)+"/tur", float64(x)+10, float64(y)+64, FaceSmall, color.RGBA{225, 205, 170, 240})
			DrawText(screen, "Alıyoruz: "+goodName+" "+itoa(amount)+"/tur", float64(x)+10, float64(y)+82, FaceSmall, color.RGBA{225, 205, 170, 240})
			DrawText(screen, "Ödeme: -"+itoa(gold)+" altın/tur", float64(x)+10, float64(y)+100, FaceSmall, color.RGBA{230, 170, 135, 240})
		}
		if c.route.SuspendedTurns > 0 {
			DrawText(screen, "Askıda: "+itoa(c.route.SuspendedTurns)+" tur", float64(x)+10, float64(y)+118, FaceSmall, color.RGBA{230, 170, 135, 240})
		}
	} else {
		DrawText(screen, "Hacim: "+itoa(c.amount)+"/tur   Fraksiyon: "+itoa(c.factions), float64(x)+10, float64(y)+46, FaceSmall, color.RGBA{187, 203, 222, 230})
		DrawText(screen, "Emtia: "+c.goods, float64(x)+10, float64(y)+64, FaceSmall, color.RGBA{197, 190, 168, 230})
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

func (r *Renderer) tradePortScreenPos(region *world.Region) (float64, float64) {
	if r != nil && r.worldMap != nil && region != nil {
		for index, settlement := range region.Settlements {
			if settlement.Type != world.SettlementPort {
				continue
			}
			if ax, ay, ok := r.worldMap.SettlementAnchor(region.ID, index); ok {
				return r.worldToScreen(float64(ax), float64(ay))
			}
		}
	}
	return r.regionScreenPos(region)
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
		pairs := r.gs.MerchantTradeRoutePortPairs(route.route)
		if len(pairs) == 0 {
			continue
		}
		pair := pairs[0]
		fromRegion := r.gs.Regions[pair.FromRegionID]
		toRegion := r.gs.Regions[pair.ToRegionID]
		if fromRegion == nil || toRegion == nil {
			continue
		}
		sx, sy := r.tradePortScreenPos(fromRegion)
		dx, dy := r.tradePortScreenPos(toRegion)
		mx := (sx + dx) / 2
		my := (sy + dy) / 2
		vx := dx - sx
		vy := dy - sy
		dist := math.Hypot(vx, vy)
		if dist < 1 {
			continue
		}
		curve := routeCurveOffset("player-port|"+key, dist)
		cx := mx + (-vy/dist)*curve
		cy := my + (vx/dist)*curve
		drawDashedTradeCurve(screen, sx, sy, cx, cy, dx, dy, 3.0, playerTradeRouteColor, r.tradeOverlayOccludesSegment)
		if !r.tradeOverlayOccludesPoint(sx, sy) {
			vector.FillCircle(screen, float32(sx), float32(sy), 5, playerTradeRouteColor, true)
			vector.StrokeCircle(screen, float32(sx), float32(sy), 8, 1.2, color.RGBA{92, 54, 18, 220}, true)
		}
		if !r.tradeOverlayOccludesPoint(dx, dy) {
			vector.FillCircle(screen, float32(dx), float32(dy), 5, playerTradeRouteColor, true)
			vector.StrokeCircle(screen, float32(dx), float32(dy), 8, 1.2, color.RGBA{92, 54, 18, 220}, true)
		}
		if r.camScale >= 1.05 {
			label := itoa(tradeRouteDisplayAmount(route.route)) + "/tur"
			labelW := MeasureText(label, FaceSmall)
			if !r.tradeOverlayOccludesPoint(mx, my) {
				DrawText(screen, label, mx-labelW/2, my-8, FaceSmall, playerTradeRouteColor)
			}
		}
		r.tradeCorridors = append(r.tradeCorridors, tradeCorridorInfo{
			fromName:  chooseRegionLabel(fromRegion),
			toName:    chooseRegionLabel(toRegion),
			amount:    tradeRouteDisplayAmount(route.route),
			factions:  2,
			goods:     route.goodName,
			sx:        sx,
			sy:        sy,
			cx:        cx,
			cy:        cy,
			dx:        dx,
			dy:        dy,
			hitWidth:  10,
			dashed:    true,
			route:     route.route,
			routeKeys: route.routeKeys,
		})
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
	prevX, prevY := c.sx, c.sy
	for i := 1; i <= segments; i++ {
		t := float64(i) / float64(segments)
		x, y := quadBezierPoint(c.sx, c.sy, c.cx, c.cy, c.dx, c.dy, t)
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
	return fleet, fleet != nil && r.merchantTradeBonusForArmy(fleet) > 0
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
		vector.StrokeLine(screen, float32(fromX), float32(fromY), float32(toX), float32(toY), 4.5, color.RGBA{22, 25, 30, 180}, false)
		vector.StrokeLine(screen, float32(fromX), float32(fromY), float32(toX), float32(toY), 1.8, color.RGBA{244, 195, 52, 210}, false)
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
		r.drawArmyIcon(screen, fleet.ID, fleet.OwnerID, position.X, position.Y, fc, unitCount, false, true, position.X+armyIconInnerHalf+8)
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
		if c.offMap {
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

	// explicit links from scenario data
	for _, def := range r.gs.TradeCenters.Centers {
		from, ok := indexByID[def.ID]
		if !ok {
			continue
		}
		for _, lid := range def.Links {
			to, ok := indexByID[lid]
			if !ok || to == from {
				continue
			}
			adj[from] = append(adj[from], to)
			adj[to] = append(adj[to], from)
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
	showLabels := r.camScale >= 1.05

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
	factionHub := make(map[string]*world.Region, len(merged)*2)
	factionCenter := make(map[string]int, len(merged)*2)
	type linkAgg struct {
		flow      int
		factions  map[string]struct{}
		goods     map[string]int
		routeKeys []string
	}
	centerLinkFlow := map[string]*linkAgg{}
	centerSpokeFlow := map[string]int{}
	mergedKeys := make([]string, 0, len(merged))
	for key := range merged {
		mergedKeys = append(mergedKeys, key)
	}
	sort.Strings(mergedKeys)
	for _, key := range mergedKeys {
		route := merged[key]
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
		if ca >= 0 {
			centerSpokeFlow[route.factionA] += route.amount
		}
		if cb >= 0 {
			centerSpokeFlow[route.factionB] += route.amount
		}
		if ca < 0 || cb < 0 || ca == cb {
			continue
		}
		path := shortestCenterPath(adj, ca, cb)
		if len(path) < 2 {
			continue
		}
		for pi := 0; pi < len(path)-1; pi++ {
			ka, kb := path[pi], path[pi+1]
			if ka > kb {
				ka, kb = kb, ka
			}
			key := itoa(ka) + "|" + itoa(kb)
			agg := centerLinkFlow[key]
			if agg == nil {
				agg = &linkAgg{
					factions: make(map[string]struct{}, 4),
					goods:    make(map[string]int, 4),
				}
				centerLinkFlow[key] = agg
			}
			agg.flow += route.amount
			agg.factions[route.factionA] = struct{}{}
			agg.factions[route.factionB] = struct{}{}
			agg.routeKeys = append(agg.routeKeys, route.routeKeys...)
			if route.goodName != "" {
				agg.goods[route.goodName] += route.amount
			}
		}
	}

	// Faction -> trade center spokes (çok hafif)
	if r.camScale >= 0.95 {
		factionIDs := make([]string, 0, len(factionHub))
		for fid := range factionHub {
			factionIDs = append(factionIDs, fid)
		}
		sort.Strings(factionIDs)
		for _, fid := range factionIDs {
			hub := factionHub[fid]
			if hub == nil {
				continue
			}
			centerIdx := factionCenter[fid]
			if centerIdx < 0 || centerIdx >= len(centers) {
				continue
			}
			flow := centerSpokeFlow[fid]
			if flow <= 0 {
				continue
			}
			hx, hy := r.regionScreenPos(hub)
			c := centers[centerIdx]
			w := float32(0.8)
			if flow >= 12 {
				w = 1.2
			}
			col := color.RGBA{180, 195, 220, 62}
			if preFocusCenter >= 0 && centerIdx != preFocusCenter {
				col = color.RGBA{120, 135, 160, 18}
			}
			if !r.tradeOverlayOccludesSegment(hx, hy, c.x, c.y) {
				vector.StrokeLine(screen, float32(hx), float32(hy), float32(c.x), float32(c.y), w, col, false)
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
	linkKeys := make([]string, 0, len(linkKeySet))
	for key := range linkKeySet {
		linkKeys = append(linkKeys, key)
	}
	sort.Strings(linkKeys)
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

		glow := color.RGBA{120, 108, 86, 18}
		core := color.RGBA{165, 150, 118, 42}
		coreW := float32(1.0)
		glowW := float32(2.8)
		if amount > 0 {
			alphaScale := min(uint8(80+(amount*12)), 255)
			glow = color.RGBA{255, 224, 138, alphaScale}
			core = color.RGBA{247, 232, 176, alphaScale}
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

		if preFocusCenter >= 0 && i != preFocusCenter && j != preFocusCenter {
			if amount > 0 {
				glow = color.RGBA{180, 170, 130, 10}
				core = color.RGBA{170, 165, 140, 34}
				coreW = 1.1
				glowW = 3.4
			} else {
				glow = color.RGBA{100, 94, 82, 8}
				core = color.RGBA{120, 116, 104, 22}
			}
		}
		segments := 22
		for i := 0; i < segments; i++ {
			t1 := float64(i) / float64(segments)
			t2 := float64(i+1) / float64(segments)
			x1, y1 := quadBezierPoint(sx, sy, cx, cy, dx, dy, t1)
			x2, y2 := quadBezierPoint(sx, sy, cx, cy, dx, dy, t2)
			if r.tradeOverlayOccludesSegment(x1, y1, x2, y2) {
				continue
			}
			vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), glowW, glow, false)
			vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), coreW, core, false)
		}
		if amount > 0 && showLabels {
			lx, ly := quadBezierPoint(sx, sy, cx, cy, dx, dy, 0.5)
			if !r.tradeOverlayOccludesPoint(lx, ly) {
				qtyStr := itoa(amount) + "/tur"
				tw2 := MeasureText(qtyStr, FaceSmall)
				DrawText(screen, qtyStr, lx-tw2/2, ly-8, FaceSmall, color.RGBA{225, 204, 144, 230})
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
		r.tradeCorridors = append(r.tradeCorridors, tradeCorridorInfo{
			fromName: centers[i].nameTR,
			toName:   centers[j].nameTR,
			amount:   amount,
			factions: factionCount,
			goods:    goodsSummary,
			sx:       sx,
			sy:       sy,
			cx:       cx,
			cy:       cy,
			dx:       dx,
			dy:       dy,
			hitWidth: float64(glowW) + 4,
			routeKeys: func() []string {
				if agg == nil {
					return nil
				}
				return agg.routeKeys
			}(),
		})
	}
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
			vol += r.gs.EffectiveRegionTradeCapacity(reg)
		}
		for _, tr := range r.gs.TradeRoutes {
			if tr.ToFactionID != "" && tr.FromFactionID != "" {
				fromHub := r.factionPrimaryRegion(tr.FromFactionID)
				toHub := r.factionPrimaryRegion(tr.ToFactionID)
				if fromHub != nil && toHub != nil {
					ca := r.nearestTradeCenterIndex(fromHub, centers)
					cb := r.nearestTradeCenterIndex(toHub, centers)
					if ca == idx || cb == idx {
						vol += tr.AmountPerTurn
					}
				}
			}
		}
		centerVolume[idx] = vol
	}

	for i := range centers {
		alphaBg := uint8(235)
		alphaBorder := uint8(235)
		alphaText := uint8(255)
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

		nameW := float32(MeasureText(centers[i].nameTR, FaceSmall))
		contentW := nameW
		volStr := ""
		if !centers[i].offMap {
			capacityBonus, incomeBonus := r.tradeCenterBenefits(centers[i])
			if centers[i].networkOnly {
				volStr = "Ağ geçidi | Hacim: " + itoa(centerVolume[i])
			} else {
				volStr = "Hacim: " + itoa(centerVolume[i])
			}
			if !centers[i].networkOnly && (capacityBonus != 0 || incomeBonus != 0) {
				volStr += " | +" + itoa(capacityBonus) + " kap. | +" + itoa(incomeBonus) + " altın"
			}
			volW := float32(MeasureText(volStr, FaceSmall))
			if volW > contentW {
				contentW = volW
			}
		}
		w := contentW + 40 // yatay padding + ikon/kenar payı
		if w < 116 {
			w = 116
		}
		h := float32(38)
		if centers[i].offMap {
			h = 22
		}
		x := float32(centers[i].x) - w/2
		y := float32(centers[i].y) - h/2
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
		}
		vector.FillRect(screen, x, y, w, h, bgColor, false)

		// Border: gold for primary, bronze for secondary
		borderColor := color.RGBA{197, 160, 89, alphaBorder}
		if centers[i].tier == world.TradeCenterPrimary {
			vector.StrokeRect(screen, x-1, y-1, w+2, h+2, 1.2, color.RGBA{235, 200, 110, alphaBorder}, false)
			vector.StrokeRect(screen, x+1, y+1, w-2, h-2, 0.8, color.RGBA{150, 110, 50, alphaBorder}, false)
		} else {
			borderColor = color.RGBA{160, 130, 90, alphaBorder}
			if centers[i].offMap {
				borderColor = color.RGBA{118, 156, 188, alphaBorder}
			}
		}
		vector.StrokeRect(screen, x, y, w, h, 1.0, borderColor, false)

		// Center Name
		nameCol := color.RGBA{242, 226, 174, alphaText}
		if centers[i].tier == world.TradeCenterPrimary {
			nameCol = color.RGBA{255, 235, 170, alphaText}
		}
		if centers[i].offMap {
			nameCol = color.RGBA{210, 228, 245, alphaText}
		}
		DrawText(screen, centers[i].nameTR, float64(x)+20, float64(y)+4, FaceSmall, nameCol)

		// Volume indicator
		if !centers[i].offMap {
			DrawText(screen, volStr, float64(x)+20, float64(y)+20, FaceSmall, color.RGBA{180, 180, 170, alphaText})
		}
	}
	r.drawTradeBonusFleetMarkers(screen)
}

// factionPrimaryRegion bir fraksiyonun görsel temsili için ana bölgesini döner.
// Önce başkent settlement'ı olan bölgeyi, yoksa ilk bulunan bölgeyi döner.
func (r *Renderer) factionPrimaryRegion(factionID string) *world.Region {
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
