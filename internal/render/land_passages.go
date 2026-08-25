package render

import (
	"image/color"
	"math"

	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	landPassageLineWidth = float32(3)
	landPassageDash      = float64(12)
	landPassageGap       = float64(8)
)

var landPassageColor = color.RGBA{255, 204, 82, 235}

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
		drawDashedLandPassage(screen, x1, y1, x2, y2, landPassageLineWidth, landPassageColor)
		if r.editLandPassageAdjustMode && i == r.editLandPassageSelected {
			selectedColor := color.RGBA{90, 240, 255, 220}
			drawDashedLandPassage(screen, x1, y1, x2, y2, landPassageLineWidth+2, selectedColor)
			vector.StrokeCircle(screen, float32(x1), float32(y1), 8, 2, selectedColor, true)
			vector.StrokeCircle(screen, float32(x2), float32(y2), 8, 2, selectedColor, true)
		}
		vector.FillCircle(screen, float32(x1), float32(y1), 3.5, landPassageColor, true)
		vector.FillCircle(screen, float32(x2), float32(y2), 3.5, landPassageColor, true)
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
	r.editLandPassageDragBefore = nil
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
	r.editLandPassageDragBefore = nil
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
	r.editLandPassageDragBefore = cloneLandPassages(r.gs.LandPassages)
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
	before := r.editLandPassageDragBefore
	after := cloneLandPassages(r.gs.LandPassages)
	index := r.editLandPassageSelected
	changed := r.editLandPassageDragChanged
	r.editLandPassageDragEndpoint = -1
	r.editLandPassageDragBefore = nil
	r.editLandPassageDragChanged = false
	if !changed {
		r.gs.LandPassages = cloneLandPassages(before)
		r.editLandPassageMessage = ""
		return
	}
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.gs.LandPassages = cloneLandPassages(before)
			rr.editLandPassageSelected = index
			rr.editLandPassageMessage = "uç noktası geri alındı"
		},
		redo: func(rr *Renderer) {
			rr.gs.LandPassages = cloneLandPassages(after)
			rr.editLandPassageSelected = index
			rr.editLandPassageMessage = "uç noktası taşındı"
		},
	})
	r.editLandPassageMessage = "uç noktası taşındı"
}

func (r *Renderer) deleteSelectedLandPassage() {
	index := r.editLandPassageSelected
	if index < 0 || index >= len(r.gs.LandPassages) {
		r.editLandPassageMessage = "önce bir geçiş seç"
		return
	}
	before := cloneLandPassages(r.gs.LandPassages)
	after := make([]world.LandPassage, 0, len(r.gs.LandPassages)-1)
	after = append(after, r.gs.LandPassages[:index]...)
	after = append(after, r.gs.LandPassages[index+1:]...)
	r.gs.LandPassages = after
	r.editLandPassageSelected = -1
	r.editLandPassageDragEndpoint = -1
	r.editLandPassageDragBefore = nil
	r.editLandPassageDragChanged = false
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.gs.LandPassages = cloneLandPassages(before)
			rr.editLandPassageSelected = index
			rr.editLandPassageMessage = "geçiş geri alındı"
		},
		redo: func(rr *Renderer) {
			rr.gs.LandPassages = cloneLandPassages(after)
			rr.editLandPassageSelected = -1
			rr.editLandPassageMessage = "geçiş silindi"
		},
	})
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
	before := cloneLandPassages(r.gs.LandPassages)
	r.gs.LandPassages = append(r.gs.LandPassages, world.LandPassage{
		From:         from,
		To:           rid,
		Type:         world.LandPassageStrait,
		MoveCost:     1,
		DefenseBonus: 15,
		Start:        &start,
		End:          &[2]int{endX, endY},
	})
	after := cloneLandPassages(r.gs.LandPassages)
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.gs.LandPassages = cloneLandPassages(before)
			rr.editLandPassageMessage = "geçiş geri alındı"
		},
		redo: func(rr *Renderer) {
			rr.gs.LandPassages = cloneLandPassages(after)
			rr.editLandPassageMessage = "geçiş eklendi"
		},
	})
	r.editLandPassageMessage = "geçiş eklendi"
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
