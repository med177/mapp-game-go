package render

import (
	"fmt"
	"image/color"
	"math"

	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	landPassageLineWidth       = float32(3)
	landPassageDash            = float64(12)
	landPassageGap             = float64(8)
	landPassageLabelZoomOffset = 0.3
)

type landPassageVisualStyle struct {
	width float32
	dash  float64
	gap   float64
	color color.RGBA
}

func landPassageStyle(passageType world.LandPassageType) landPassageVisualStyle {
	switch passageType {
	case world.LandPassageBridge:
		return landPassageVisualStyle{width: 5, color: color.RGBA{92, 205, 126, 240}}
	case world.LandPassageMountainPass:
		return landPassageVisualStyle{width: 4, dash: 5, gap: 7, color: color.RGBA{238, 157, 62, 240}}
	case world.LandPassageFortifiedCrossing:
		return landPassageVisualStyle{width: 5, dash: 13, gap: 5, color: color.RGBA{211, 91, 81, 245}}
	default:
		return landPassageVisualStyle{width: 3, dash: 8, gap: 10, color: color.RGBA{86, 194, 220, 235}}
	}
}

// drawLandPassages özel karasal geçişleri kalın kesikli çizgi olarak gösterir.
// start/end verilmişse çizgi doğrudan bu senaryo koordinatları arasında çizilir;
// eski kayıtlarda bölge anchor'ına geri dönülür.
func (r *Renderer) drawLandPassages(screen *ebiten.Image) {
	if r == nil || r.gs == nil {
		return
	}
	for i := range r.gs.LandPassages {
		passage := &r.gs.LandPassages[i]
		from := r.gs.Regions[passage.From]
		to := r.gs.Regions[passage.To]
		if from == nil || to == nil || from.IsSea || to.IsSea {
			continue
		}
		x1, y1, x2, y2 := r.landPassageScreenEndpoints(passage, from, to)
		style := landPassageStyle(passage.Type)
		drawLandPassageLine(screen, x1, y1, x2, y2, style)
		if passage.Name != "" && r.camScale >= r.landPassageLabelZoomThreshold() {
			drawLandPassageLabel(screen, passage.Name, (x1+x2)/2, (y1+y2)/2)
		}
		if r.editLandPassageAdjustMode && i == r.editLandPassageSelected {
			selectedColor := color.RGBA{90, 240, 255, 220}
			drawDashedLandPassage(screen, x1, y1, x2, y2, landPassageLineWidth+2, selectedColor)
			vector.StrokeCircle(screen, float32(x1), float32(y1), 8, 2, selectedColor, true)
			vector.StrokeCircle(screen, float32(x2), float32(y2), 8, 2, selectedColor, true)
		}
		vector.FillCircle(screen, float32(x1), float32(y1), 3.5, style.color, true)
		vector.FillCircle(screen, float32(x2), float32(y2), 3.5, style.color, true)
	}

	if !r.editLandPassageMode || r.editLandPassageFrom == "" {
		return
	}
	from := r.gs.Regions[r.editLandPassageFrom]
	if from == nil || from.IsSea {
		return
	}
	mx, my := ebiten.CursorPosition()
	targetID := r.editRegionAt(float64(mx), float64(my))
	to := r.gs.Regions[targetID]
	if !r.editLandPassageStartSet {
		return
	}
	startX, startY := r.worldToScreen(wcX(r.editLandPassageStart[0]), wcY(r.editLandPassageStart[1]))
	if to == nil || to.IsSea || to.ID == from.ID {
		vector.StrokeCircle(screen, float32(startX), float32(startY), 8, 2, color.RGBA{90, 240, 255, 220}, true)
		return
	}
	drawDashedLandPassage(screen, startX, startY, float64(mx), float64(my), 2.5, color.RGBA{90, 240, 255, 220})
	vector.StrokeCircle(screen, float32(startX), float32(startY), 8, 2, color.RGBA{90, 240, 255, 220}, true)
}

func (r *Renderer) landPassageHoverAt(fx, fy float64) int {
	if r == nil || r.gs == nil {
		return -1
	}
	for i := range r.gs.LandPassages {
		passage := &r.gs.LandPassages[i]
		from := r.gs.Regions[passage.From]
		to := r.gs.Regions[passage.To]
		if from == nil || to == nil || from.IsSea || to.IsSea {
			continue
		}
		x1, y1, x2, y2 := r.landPassageScreenEndpoints(passage, from, to)
		if pointSegmentDistanceSquared(fx, fy, x1, y1, x2, y2) <= 12*12 {
			return i
		}
		if passage.Name != "" && r.camScale >= r.landPassageLabelZoomThreshold() {
			if landPassageLabelRect(passage.Name, (x1+x2)/2, (y1+y2)/2).Hit(fx, fy) {
				return i
			}
		}
	}
	return -1
}

func (r *Renderer) landPassageLabelZoomThreshold() float64 {
	threshold := r.maxCameraZoomScale() - landPassageLabelZoomOffset
	minZoom := minCameraScale()
	if threshold < minZoom {
		return minZoom
	}
	return threshold
}

func drawLandPassageLabel(screen *ebiten.Image, name string, centerX, centerY float64) {
	rect := landPassageLabelRect(name, centerX, centerY)
	vector.FillRect(screen, float32(rect.X), float32(rect.Y), float32(rect.W), float32(rect.H), color.RGBA{18, 22, 28, 220}, true)
	vector.StrokeRect(screen, float32(rect.X), float32(rect.Y), float32(rect.W), float32(rect.H), 1.2, color.RGBA{255, 204, 82, 225}, true)
	drawUILabel(screen, gameui.Rect{X: rect.X + 6, Y: rect.Y + 5, W: rect.W - 12, H: rect.H - 8}, trimTextToWidth(name, FaceSmall, rect.W-12), ColorGold, gameui.TextSmall, gameui.TextAlignCenter)
}

func landPassageLabelRect(name string, centerX, centerY float64) gameui.Rect {
	const (
		labelMinW = 64.0
		labelMaxW = 260.0
		labelH    = 28.0
	)
	labelW := MeasureText(name, FaceSmall) + 20
	if labelW < labelMinW {
		labelW = labelMinW
	}
	if labelW > labelMaxW {
		labelW = labelMaxW
	}
	return gameui.Rect{X: centerX - labelW/2, Y: centerY - labelH/2, W: labelW, H: labelH}
}

func (r *Renderer) drawLandPassageHoverTooltip(screen *ebiten.Image) {
	if r == nil || r.gs == nil || r.editLandPassageForm.show {
		return
	}
	mx, my := ebiten.CursorPosition()
	index := r.landPassageHoverAt(float64(mx), float64(my))
	if index < 0 || index >= len(r.gs.LandPassages) {
		return
	}
	passage := r.gs.LandPassages[index]
	from := r.gs.Regions[passage.From]
	to := r.gs.Regions[passage.To]
	if from == nil || to == nil {
		return
	}
	name := passage.Name
	if name == "" {
		name = landPassageTypeLabel(passage.Type)
	}
	const (
		width  = 330.0
		height = 154.0
	)
	x, y, w, h := tooltipRect(float64(mx), float64(my), width, height)
	drawTooltipBox(screen, x, y, w, h)
	DrawText(screen, trimTextToWidth(name, FaceMed, w-20), x+10, y+10, FaceMed, ColorGold)
	DrawText(screen, "Güzergâh: "+landPassageRegionName(from), x+10, y+36, FaceSmall, ColorWhite)
	DrawText(screen, "→ "+landPassageRegionName(to), x+10, y+54, FaceSmall, ColorWhite)
	DrawText(screen, fmt.Sprintf("Tip: %s", landPassageTypeLabel(passage.Type)), x+10, y+78, FaceSmall, ColorGray)
	DrawText(screen, fmt.Sprintf("Hareket maliyeti: %d", passage.MoveCost), x+10, y+98, FaceSmall, ColorGray)
	DrawText(screen, fmt.Sprintf("Savunma bonusu: +%d%%", passage.DefenseBonus), x+10, y+118, FaceSmall, ColorGray)
}

func landPassageRegionName(region *world.Region) string {
	if region == nil {
		return "-"
	}
	if region.NameTR != "" {
		return region.NameTR
	}
	if region.Name != "" {
		return region.Name
	}
	return string(region.ID)
}

func (r *Renderer) landPassageScreenEndpoints(passage *world.LandPassage, from, to *world.Region) (float64, float64, float64, float64) {
	if passage != nil && passage.HasCustomEndpoints() {
		start := passage.Start
		end := passage.End
		x1, y1 := r.worldToScreen(wcX(start[0]), wcY(start[1]))
		x2, y2 := r.worldToScreen(wcX(end[0]), wcY(end[1]))
		return x1, y1, x2, y2
	}
	x1, y1 := r.regionScreenPos(from)
	x2, y2 := r.regionScreenPos(to)
	return x1, y1, x2, y2
}

// syncLandPassageRegionsFromMap, Edit Mode haritasında geçiş uçlarının
// altındaki güncel bölge kimliklerini geçiş kaydına yansıtır. Geçişin çizgisi
// sabit koordinatlarda kaldığı için bölge boyama veya şekil değişikliği sonrası
// From/To değerleri eski bölgede kalmamalıdır.
func (r *Renderer) syncLandPassageRegionsFromMap() bool {
	if r == nil || r.gs == nil || r.worldMap == nil {
		return false
	}

	changed := false
	for i := range r.gs.LandPassages {
		passage := &r.gs.LandPassages[i]
		if !passage.HasCustomEndpoints() {
			continue
		}

		startX, startY := shapeRasterWorldPoint([2]float32{
			float32(passage.Start[0]), float32(passage.Start[1]),
		})
		endX, endY := shapeRasterWorldPoint([2]float32{
			float32(passage.End[0]), float32(passage.End[1]),
		})
		fromID := r.worldMap.RegionAt(int(math.Round(startX)), int(math.Round(startY)))
		toID := r.worldMap.RegionAt(int(math.Round(endX)), int(math.Round(endY)))
		from := r.gs.Regions[fromID]
		to := r.gs.Regions[toID]
		if from == nil || to == nil || from.IsSea || to.IsSea || fromID == toID {
			continue
		}
		if passage.From != fromID {
			passage.From = fromID
			changed = true
		}
		if passage.To != toID {
			passage.To = toID
			changed = true
		}
	}
	return changed
}

func drawDashedLandPassage(screen *ebiten.Image, x1, y1, x2, y2 float64, width float32, col color.RGBA) {
	dx := x2 - x1
	dy := y2 - y1
	distance := math.Hypot(dx, dy)
	if distance <= 0.01 {
		return
	}
	ux := dx / distance
	uy := dy / distance
	for offset := float64(0); offset < distance; offset += landPassageDash + landPassageGap {
		end := offset + landPassageDash
		if end > distance {
			end = distance
		}
		vector.StrokeLine(
			screen,
			float32(x1+ux*offset), float32(y1+uy*offset),
			float32(x1+ux*end), float32(y1+uy*end),
			width, col, true,
		)
	}
}

func drawLandPassageLine(screen *ebiten.Image, x1, y1, x2, y2 float64, style landPassageVisualStyle) {
	if style.dash <= 0 || style.gap <= 0 {
		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), style.width, style.color, true)
		return
	}
	drawDashedLandPassageWithPattern(screen, x1, y1, x2, y2, style.width, style.color, style.dash, style.gap)
}

func drawDashedLandPassageWithPattern(screen *ebiten.Image, x1, y1, x2, y2 float64, width float32, col color.RGBA, dash, gap float64) {
	dx := x2 - x1
	dy := y2 - y1
	distance := math.Hypot(dx, dy)
	if distance <= 0.01 {
		return
	}
	ux := dx / distance
	uy := dy / distance
	for offset := float64(0); offset < distance; offset += dash + gap {
		end := math.Min(offset+dash, distance)
		vector.StrokeLine(screen, float32(x1+ux*offset), float32(y1+uy*offset), float32(x1+ux*end), float32(y1+uy*end), width, col, true)
	}
}

func (r *Renderer) toggleEditLandPassageMode() {
	if r.editTerrainAreaMode {
		return
	}
	r.editLandPassageMode = !r.editLandPassageMode
	r.editLandPassageAdjustMode = false
	r.editNeighborAddMode = false
	r.editNeighborAddFrom = ""
	r.editNeighborAddMessage = ""
	r.editLandPassageFrom = ""
	r.editLandPassageStart = [2]int{}
	r.editLandPassageStartSet = false
	r.editLandPassageSelected = -1
	r.editLandPassageDragEndpoint = -1
	r.editLandPassageDragChanged = false
	r.editLandPassageMessage = ""
	if r.editLandPassageMode {
		r.editLandPassageMessage = "strait varsayılanı"
	}
}

func (r *Renderer) toggleEditLandPassageAdjustMode() {
	if r.editTerrainAreaMode {
		return
	}
	r.editLandPassageAdjustMode = !r.editLandPassageAdjustMode
	r.editLandPassageMode = false
	r.editNeighborAddMode = false
	r.editNeighborAddFrom = ""
	r.editNeighborAddMessage = ""
	r.editLandPassageFrom = ""
	r.editLandPassageStart = [2]int{}
	r.editLandPassageStartSet = false
	r.editLandPassageSelected = -1
	r.editLandPassageDragEndpoint = -1
	r.editLandPassageDragChanged = false
	if r.editLandPassageAdjustMode {
		r.editLandPassageMessage = "çizgiye veya uç noktasına tıkla"
	} else {
		r.editLandPassageMessage = ""
	}
}

func (r *Renderer) handleEditLandPassageAdjustClick(fx, fy float64) {
	index, endpoint := r.landPassageHitAt(fx, fy)
	if index < 0 {
		r.editLandPassageSelected = -1
		r.editLandPassageDragEndpoint = -1
		r.editLandPassageMessage = "geçiş çizgisi bulunamadı"
		return
	}
	r.editLandPassageSelected = index
	r.editLandPassageDragEndpoint = endpoint
	if endpoint < 0 {
		r.editLandPassageMessage = "seçildi; uç noktasını sürükle"
		return
	}
	r.editLandPassageDragChanged = false
	r.editLandPassageMessage = "uç noktası taşınıyor"
}

func (r *Renderer) updateEditLandPassageDrag(fx, fy float64) {
	if r.editLandPassageSelected < 0 || r.editLandPassageSelected >= len(r.gs.LandPassages) || r.editLandPassageDragEndpoint < 0 {
		return
	}
	passage := &r.gs.LandPassages[r.editLandPassageSelected]
	if !passage.HasCustomEndpoints() {
		from := r.gs.Regions[passage.From]
		to := r.gs.Regions[passage.To]
		if from == nil || to == nil {
			return
		}
		x1, y1, x2, y2 := r.landPassageScreenEndpoints(passage, from, to)
		wx1, wy1 := r.screenToWorld(x1, y1)
		wx2, wy2 := r.screenToWorld(x2, y2)
		startX, startY := scenarioCoordsFromWorld(wx1, wy1)
		endX, endY := scenarioCoordsFromWorld(wx2, wy2)
		passage.Start = &[2]int{startX, startY}
		passage.End = &[2]int{endX, endY}
	}
	wx, wy := r.screenToWorld(fx, fy)
	x, y := scenarioCoordsFromWorld(wx, wy)
	if r.editLandPassageDragEndpoint == 0 {
		if passage.Start[0] != x || passage.Start[1] != y {
			passage.Start[0], passage.Start[1] = x, y
			r.editLandPassageDragChanged = true
		}
	} else {
		if passage.End[0] != x || passage.End[1] != y {
			passage.End[0], passage.End[1] = x, y
			r.editLandPassageDragChanged = true
		}
	}
}

func (r *Renderer) finishEditLandPassageDrag() {
	if r.editLandPassageDragEndpoint < 0 {
		return
	}
	index := r.editLandPassageSelected
	changed := r.editLandPassageDragChanged
	r.editLandPassageDragEndpoint = -1
	r.editLandPassageDragChanged = false
	if !changed {
		r.editLandPassageMessage = ""
		return
	}
	r.editDirty = true
	r.editLandPassageSelected = index
	r.editLandPassageMessage = "uç noktası taşındı"
}

func (r *Renderer) deleteSelectedLandPassage() {
	index := r.editLandPassageSelected
	if index < 0 || index >= len(r.gs.LandPassages) {
		r.editLandPassageMessage = "önce bir geçiş seç"
		return
	}
	r.gs.LandPassages = append(r.gs.LandPassages[:index], r.gs.LandPassages[index+1:]...)
	r.editLandPassageSelected = -1
	r.editLandPassageDragEndpoint = -1
	r.editLandPassageDragChanged = false
	r.editDirty = true
	r.editLandPassageMessage = "geçiş silindi"
}

func (r *Renderer) landPassageHitAt(fx, fy float64) (int, int) {
	const endpointRadius = 12.0
	const lineRadius = 9.0
	bestEndpointDist := endpointRadius * endpointRadius
	bestLineDist := lineRadius * lineRadius
	bestEndpointIndex, bestEndpoint := -1, -1
	bestLineIndex := -1
	for i := range r.gs.LandPassages {
		passage := &r.gs.LandPassages[i]
		from := r.gs.Regions[passage.From]
		to := r.gs.Regions[passage.To]
		if from == nil || to == nil || from.IsSea || to.IsSea {
			continue
		}
		x1, y1, x2, y2 := r.landPassageScreenEndpoints(passage, from, to)
		for endpoint, point := range [][2]float64{{x1, y1}, {x2, y2}} {
			dx, dy := fx-point[0], fy-point[1]
			distance := dx*dx + dy*dy
			if distance <= bestEndpointDist {
				bestEndpointDist = distance
				bestEndpointIndex = i
				bestEndpoint = endpoint
			}
		}
		distance := pointSegmentDistanceSquared(fx, fy, x1, y1, x2, y2)
		if distance <= bestLineDist {
			bestLineDist = distance
			bestLineIndex = i
		}
	}
	if bestEndpointIndex >= 0 {
		return bestEndpointIndex, bestEndpoint
	}
	return bestLineIndex, -1
}

func pointSegmentDistanceSquared(px, py, x1, y1, x2, y2 float64) float64 {
	dx, dy := x2-x1, y2-y1
	if dx == 0 && dy == 0 {
		dx, dy = px-x1, py-y1
		return dx*dx + dy*dy
	}
	t := ((px-x1)*dx + (py-y1)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	closestX, closestY := x1+t*dx, y1+t*dy
	dx, dy = px-closestX, py-closestY
	return dx*dx + dy*dy
}

func (r *Renderer) handleEditLandPassageClick(fx, fy float64) {
	rid := r.editRegionAt(fx, fy)
	region := r.gs.Regions[rid]
	if region == nil || region.IsSea {
		r.editLandPassageMessage = "yalnızca kara bölgesi"
		return
	}
	if r.editLandPassageFrom == "" {
		r.editLandPassageFrom = rid
		wx, wy := r.screenToWorld(fx, fy)
		r.editLandPassageStart[0], r.editLandPassageStart[1] = scenarioCoordsFromWorld(wx, wy)
		r.editLandPassageStartSet = true
		r.editLandPassageMessage = ""
		return
	}
	from := r.editLandPassageFrom
	r.editLandPassageFrom = ""
	start := r.editLandPassageStart
	r.editLandPassageStart = [2]int{}
	r.editLandPassageStartSet = false
	if from == rid {
		r.editLandPassageMessage = "aynı bölge seçilemez"
		return
	}
	if world.HasLandPassage(r.gs.LandPassages, from, rid) {
		r.editLandPassageMessage = "geçiş zaten var"
		return
	}

	wx, wy := r.screenToWorld(fx, fy)
	endX, endY := scenarioCoordsFromWorld(wx, wy)
	r.openNewLandPassageForm(from, rid, start, [2]int{endX, endY})
	r.editLandPassageMessage = "özellikleri gir"
}

func cloneLandPassages(src []world.LandPassage) []world.LandPassage {
	if src == nil {
		return nil
	}
	dst := make([]world.LandPassage, len(src))
	copy(dst, src)
	for i := range dst {
		if src[i].Start != nil {
			start := *src[i].Start
			dst[i].Start = &start
		}
		if src[i].End != nil {
			end := *src[i].End
			dst[i].End = &end
		}
	}
	return dst
}
