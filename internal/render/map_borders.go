package render

import (
	"image/color"
	"math"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	mapBorderStyleNone uint8 = iota
	mapBorderStyleSelected
	mapBorderStylePlayerRealm
	mapBorderStyleAlly
	mapBorderStyleEnemy
	mapBorderStyleStrong
	mapBorderStyleSubtle
	mapBorderStyleTradeStrong
	mapBorderStyleTradeSubtle
	mapBorderStyleSea
	mapBorderStyleCount
)

const mapBorderStrokeWidth float32 = 1.25

// mapBorderSegment, raster regionAt ızgarasındaki tek bir sınır hattının
// birleştirilmiş yatay veya dikey parçasıdır. Geometri dünya uzayında tutulur;
// kamera değiştikçe yalnızca ekran koordinatlarına dönüştürülür.
type mapBorderSegment struct {
	x1, y1 float32
	x2, y2 float32
	a, b   uint16
}

// Mesh parçaları, vector.Path tessellation maliyetini yalnızca kamera değiştiğinde
// değil, aynı zamanda küçük ve yönetilebilir DrawTriangles32 çağrılarına bölerek
// azaltır. Bu sınır Ebiten'in tek çağrıdaki vertex sınırının altında tutulur.
const maxMapBorderMeshVertices = 60000

type mapBorderMeshChunk struct {
	vertices []ebiten.Vertex
	indices  []uint32
}

type mapBorderMeshSet struct {
	chunks [mapBorderStyleCount][]mapBorderMeshChunk
}

type mapBorderOverlayCache struct {
	image *ebiten.Image
	white *ebiten.Image
	key   mapBorderOverlayKey
	valid bool
}

type mapBorderOverlayKey struct {
	worldMap      *WorldMap
	borderVersion uint64
	camX, camY    float64
	camScale      float64
	screenWidth   int
	screenHeight  int
}

func (m *mapBorderMeshSet) reset() {
	for style := range m.chunks {
		chunks := m.chunks[style]
		for i := range chunks {
			chunks[i].vertices = chunks[i].vertices[:0]
			chunks[i].indices = chunks[i].indices[:0]
		}
		m.chunks[style] = chunks[:0]
	}
}

func (m *mapBorderMeshSet) chunkFor(style uint8) *mapBorderMeshChunk {
	chunks := &m.chunks[style]
	if len(*chunks) == 0 || len((*chunks)[len(*chunks)-1].vertices)+4 > maxMapBorderMeshVertices {
		index := len(*chunks)
		if index < cap(*chunks) {
			*chunks = (*chunks)[:index+1]
			(*chunks)[index].vertices = (*chunks)[index].vertices[:0]
			(*chunks)[index].indices = (*chunks)[index].indices[:0]
		} else {
			*chunks = append(*chunks, mapBorderMeshChunk{
				vertices: make([]ebiten.Vertex, 0, maxMapBorderMeshVertices),
				indices:  make([]uint32, 0, maxMapBorderMeshVertices*3/2),
			})
		}
	}
	return &(*chunks)[len(*chunks)-1]
}

func (wm *WorldMap) isLandRegionIndex(gs *state.GameState, idx uint16) bool {
	if wm == nil || idx == 0 || int(idx) >= len(wm.regionIDs) || gs == nil {
		return false
	}
	r := gs.Regions[wm.regionIDs[idx]]
	return r != nil && !r.IsSea
}

func (wm *WorldMap) isSeaRegionIndex(gs *state.GameState, idx uint16) bool {
	if wm == nil || idx == 0 || int(idx) >= len(wm.regionIDs) {
		return false
	}
	if wm.seaIdx[idx] {
		return true
	}
	if gs == nil {
		return false
	}
	r := gs.Regions[wm.regionIDs[idx]]
	return r != nil && r.IsSea
}

func sameBorderPair(a1, b1, a2, b2 uint16) bool {
	return a1 == a2 && b1 == b2
}

func normalizeSeaPair(a, b uint16) (uint16, uint16) {
	if a > b {
		return b, a
	}
	return a, b
}

// boundaryPairAt, iki piksel arasındaki sınırın hangi bölgeleri ayırdığını
// döndürür. Kara tarafı her zaman a olur; böylece diplomasi rengi tek bir
// kanonik taraftan hesaplanır. Deniz-deniz sınırları da ayrıca korunur.
func (wm *WorldMap) boundaryPairAt(gs *state.GameState, first, second uint16) (uint16, uint16, bool) {
	if first == second {
		return 0, 0, false
	}
	firstLand := wm.isLandRegionIndex(gs, first)
	secondLand := wm.isLandRegionIndex(gs, second)
	switch {
	case firstLand:
		return first, second, true
	case secondLand:
		return second, first, true
	case wm.isSeaRegionIndex(gs, first) && wm.isSeaRegionIndex(gs, second):
		a, b := normalizeSeaPair(first, second)
		return a, b, true
	default:
		// Harita dışındaki boş pikseller kendi başına sınır çizmez.
		return 0, 0, false
	}
}

// rebuildBorderSegments raster regionAt sonuçlarını uzun yatay/dikey
// parçalar halinde sıkıştırır. Bu işlem harita hazırlanırken veya editörde
// regionAt değiştiğinde çalışır; her frame'de piksel taraması yapılmaz.
func (wm *WorldMap) rebuildBorderSegments(gs *state.GameState) {
	if wm == nil || WorldW < 2 || WorldH < 2 {
		return
	}
	wm.borderSegments = wm.borderSegments[:0]

	// Dikey sınırlar: x sabit, y boyunca bitişik aynı bölge çifti birleştirilir.
	for x := 1; x < WorldW; x++ {
		runStart := -1
		var runA, runB uint16
		for y := 0; y <= WorldH; y++ {
			var a, b uint16
			ok := false
			if y < WorldH {
				left := wm.regionAt[y*WorldW+x-1]
				right := wm.regionAt[y*WorldW+x]
				a, b, ok = wm.boundaryPairAt(gs, left, right)
			}
			if ok && runStart >= 0 && sameBorderPair(runA, runB, a, b) {
				continue
			}
			if runStart >= 0 {
				wm.borderSegments = append(wm.borderSegments, mapBorderSegment{
					x1: float32(x), y1: float32(runStart),
					x2: float32(x), y2: float32(y),
					a: runA, b: runB,
				})
			}
			if ok {
				runStart, runA, runB = y, a, b
			} else {
				runStart = -1
			}
		}
	}

	// Yatay sınırlar: y sabit, x boyunca bitişik aynı bölge çifti birleştirilir.
	for y := 1; y < WorldH; y++ {
		runStart := -1
		var runA, runB uint16
		for x := 0; x <= WorldW; x++ {
			var a, b uint16
			ok := false
			if x < WorldW {
				up := wm.regionAt[(y-1)*WorldW+x]
				down := wm.regionAt[y*WorldW+x]
				a, b, ok = wm.boundaryPairAt(gs, up, down)
			}
			if ok && runStart >= 0 && sameBorderPair(runA, runB, a, b) {
				continue
			}
			if runStart >= 0 {
				wm.borderSegments = append(wm.borderSegments, mapBorderSegment{
					x1: float32(runStart), y1: float32(y),
					x2: float32(x), y2: float32(y),
					a: runA, b: runB,
				})
			}
			if ok {
				runStart, runA, runB = x, a, b
			} else {
				runStart = -1
			}
		}
	}

	wm.borderStyles = make([]uint8, len(wm.borderSegments))
	wm.borderVersion++
}

func borderStyleForAffiliation(affiliation uint8) uint8 {
	switch affiliation {
	case borderAffiliationPlayerRealm:
		return mapBorderStylePlayerRealm
	case borderAffiliationAlly:
		return mapBorderStyleAlly
	case borderAffiliationEnemy:
		return mapBorderStyleEnemy
	default:
		return mapBorderStyleStrong
	}
}

// updateBorderStyles, geometriyi değiştirmeden mevcut diplomasi/map-mode
// durumuna göre hangi vektör path'ine gideceğini belirler.
func (wm *WorldMap) updateBorderStyles(gs *state.GameState, selected world.RegionID, mode MapMode) {
	if wm == nil {
		return
	}
	if len(wm.borderStyles) != len(wm.borderSegments) {
		wm.borderStyles = make([]uint8, len(wm.borderSegments))
	}
	for i := range wm.borderStyles {
		wm.borderStyles[i] = mapBorderStyleNone
	}
	if gs == nil {
		return
	}

	realmByRegion, affiliationByRegion := buildBorderDiplomacyContext(gs, wm.regionIDs)
	regionTradeNode := make(map[world.RegionID]int, len(gs.Regions))
	if mode == MapModeTrade && len(gs.TradeCenters.Centers) > 0 {
		for rid, r := range gs.Regions {
			regionTradeNode[rid] = nearestTradeCenterIndex(r, gs.TradeCenters.Centers, gs.Regions, gs.Year)
		}
	}

	for i, segment := range wm.borderSegments {
		landIdx := segment.a
		otherIdx := segment.b
		if !wm.isLandRegionIndex(gs, landIdx) {
			landIdx, otherIdx = segment.b, segment.a
		}
		if !wm.isLandRegionIndex(gs, landIdx) {
			if wm.isSeaRegionIndex(gs, segment.a) && wm.isSeaRegionIndex(gs, segment.b) {
				wm.borderStyles[i] = mapBorderStyleSea
			}
			continue
		}

		landID := wm.regionIDs[landIdx]
		if landID == selected {
			wm.borderStyles[i] = mapBorderStyleSelected
			continue
		}

		if mode == MapModeTrade && len(gs.TradeCenters.Centers) > 0 {
			currentTradeNode := regionTradeNode[landID]
			otherTradeNode := -1
			if otherIdx != 0 && int(otherIdx) < len(wm.regionIDs) {
				otherTradeNode = regionTradeNode[wm.regionIDs[otherIdx]]
			}
			if otherTradeNode != -1 && otherTradeNode != currentTradeNode {
				wm.borderStyles[i] = mapBorderStyleTradeStrong
			} else {
				wm.borderStyles[i] = mapBorderStyleTradeSubtle
			}
			continue
		}

		isOutsideRealm := otherIdx == 0 || !wm.isLandRegionIndex(gs, otherIdx)
		if !isOutsideRealm && int(otherIdx) < len(realmByRegion) {
			isOutsideRealm = realmByRegion[otherIdx] != realmByRegion[landIdx]
		}
		if !isOutsideRealm {
			wm.borderStyles[i] = mapBorderStyleSubtle
			continue
		}

		affiliation := affiliationByRegion[landIdx]
		if wm.isLandRegionIndex(gs, otherIdx) && int(otherIdx) < len(affiliationByRegion) {
			affiliation = strongerBorderAffiliation(affiliation, affiliationByRegion[otherIdx])
		}
		wm.borderStyles[i] = borderStyleForAffiliation(affiliation)
	}
	wm.borderVersion++
}

func mapBorderStyleColor(style uint8) color.RGBA {
	switch style {
	case mapBorderStyleSelected:
		return color.RGBA{255, 222, 72, 245}
	case mapBorderStylePlayerRealm:
		return borderColorPlayerRealm
	case mapBorderStyleAlly:
		return borderColorAlly
	case mapBorderStyleEnemy:
		return borderColorEnemy
	case mapBorderStyleStrong:
		return color.RGBA{30, 18, 10, 230}
	case mapBorderStyleSubtle:
		return color.RGBA{35, 22, 10, 105}
	case mapBorderStyleTradeStrong:
		return color.RGBA{242, 226, 174, 230}
	case mapBorderStyleTradeSubtle:
		return color.RGBA{35, 22, 10, 90}
	case mapBorderStyleSea:
		return color.RGBA{100, 160, 220, 160}
	default:
		return color.RGBA{}
	}
}

func shouldDrawMapBorderStyle(style uint8, camScale float64) bool {
	if camScale >= 0.85 {
		return true
	}
	// Uzak görünümde idari yardımcı çizgiler ve deniz hücresi sınırları
	// bir pikselden küçük hale gelir. Bunları GPU'ya göndermemek hem görüntüyü
	// sadeleştirir hem de zoom-out sırasında tessellation/upload yükünü keser.
	switch style {
	case mapBorderStyleSubtle, mapBorderStyleTradeSubtle, mapBorderStyleSea:
		return false
	default:
		return true
	}
}

func appendMapBorderQuad(meshes *mapBorderMeshSet, style uint8, x1, y1, x2, y2 float64) {
	dx, dy := x2-x1, y2-y1
	length := math.Hypot(dx, dy)
	if length <= 0.001 {
		return
	}

	// Uçları yarım çizgi genişliği kadar uzatmak, komşu yatay/dikey parçaların
	// köşelerde boşluk bırakmasını önler. Kenar yine gerçek vektör geometri
	// olarak kalır; yalnızca ekrana basılan mesh dikdörtgendir.
	halfWidth := float64(mapBorderStrokeWidth) * 0.5
	ux, uy := dx/length, dy/length
	x1 -= ux * halfWidth
	y1 -= uy * halfWidth
	x2 += ux * halfWidth
	y2 += uy * halfWidth
	nx, ny := -uy*halfWidth, ux*halfWidth

	col := mapBorderStyleColor(style)
	cr := float32(col.R) / 255
	cg := float32(col.G) / 255
	cb := float32(col.B) / 255
	ca := float32(col.A) / 255
	chunk := meshes.chunkFor(style)
	base := uint32(len(chunk.vertices))
	chunk.vertices = append(chunk.vertices,
		ebiten.Vertex{DstX: float32(x1 + nx), DstY: float32(y1 + ny), SrcX: 0.5, SrcY: 0.5, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca},
		ebiten.Vertex{DstX: float32(x1 - nx), DstY: float32(y1 - ny), SrcX: 0.5, SrcY: 0.5, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca},
		ebiten.Vertex{DstX: float32(x2 - nx), DstY: float32(y2 - ny), SrcX: 0.5, SrcY: 0.5, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca},
		ebiten.Vertex{DstX: float32(x2 + nx), DstY: float32(y2 + ny), SrcX: 0.5, SrcY: 0.5, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca},
	)
	chunk.indices = append(chunk.indices, base, base+1, base+2, base+2, base+3, base)
}

func (r *Renderer) drawVectorMapBorders(screen *ebiten.Image) {
	if r == nil || r.worldMap == nil || len(r.worldMap.borderSegments) == 0 {
		return
	}
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()
	key := mapBorderOverlayKey{
		worldMap:      r.worldMap,
		borderVersion: r.worldMap.borderVersion,
		camX:          r.camX,
		camY:          r.camY,
		camScale:      r.camScale,
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
	}
	if r.mapBorderCache.image == nil || r.mapBorderCache.image.Bounds().Dx() != screenWidth || r.mapBorderCache.image.Bounds().Dy() != screenHeight {
		r.mapBorderCache.image = ebiten.NewImage(screenWidth, screenHeight)
		r.mapBorderCache.valid = false
	}
	if r.mapBorderCache.white == nil {
		r.mapBorderCache.white = ebiten.NewImage(1, 1)
		r.mapBorderCache.white.WritePixels([]byte{255, 255, 255, 255})
	}
	if r.mapBorderCache.valid && r.mapBorderCache.key == key {
		screen.DrawImage(r.mapBorderCache.image, nil)
		return
	}

	r.mapBorderMeshes.reset()
	for i, segment := range r.worldMap.borderSegments {
		if i >= len(r.worldMap.borderStyles) {
			break
		}
		style := r.worldMap.borderStyles[i]
		if style == mapBorderStyleNone || !shouldDrawMapBorderStyle(style, r.camScale) {
			continue
		}
		x1, y1 := r.worldToScreen(float64(segment.x1), float64(segment.y1))
		x2, y2 := r.worldToScreen(float64(segment.x2), float64(segment.y2))
		if r.camScale < 1 && math.Hypot(x2-x1, y2-y1) < 0.85 {
			continue
		}
		if math.Max(x1, x2) < -2 || math.Min(x1, x2) > float64(screenWidth)+2 ||
			math.Max(y1, y2) < -2 || math.Min(y1, y2) > float64(screenHeight)+2 {
			continue
		}
		appendMapBorderQuad(&r.mapBorderMeshes, style, x1, y1, x2, y2)
	}

	// Path tessellation yerine ekran uzayında hazırlanmış mesh kullanılır.
	// Bu çağrılar yalnız kamera/harita değiştiğinde çalışır; sabit frame'lerde
	// yalnızca tek bir cache image ekrana kopyalanır.
	r.mapBorderCache.image.Clear()
	drawOptions := &ebiten.DrawTrianglesOptions{
		AntiAlias:      true,
		ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha,
		Address:        ebiten.AddressClampToZero,
		Filter:         ebiten.FilterLinear,
	}
	for style := uint8(mapBorderStyleSelected); style < mapBorderStyleCount; style++ {
		for _, chunk := range r.mapBorderMeshes.chunks[style] {
			if len(chunk.vertices) == 0 {
				continue
			}
			r.mapBorderCache.image.DrawTriangles32(chunk.vertices, chunk.indices, r.mapBorderCache.white, drawOptions)
		}
	}
	r.mapBorderCache.key = key
	r.mapBorderCache.valid = true
	screen.DrawImage(r.mapBorderCache.image, nil)
}
