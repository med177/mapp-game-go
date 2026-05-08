package render

import (
	"encoding/json"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"os"
	"sort"

	"mapp-game-go/internal/season"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	defaultWorldW = 2892
	defaultWorldH = 1440
)

const (
	defaultShapeOffX   float64 = -530
	defaultShapeOffY   float64 = -180
	defaultShapeScaleX float64 = 2.025
	defaultShapeScaleY float64 = 2.025
)

// Oyun dünyası boyutu ve shape dönüşümü aktif senaryodan gelir.
// Eksik alanlar eski sabitlerle tamamlanır.
var (
	WorldW = defaultWorldW
	WorldH = defaultWorldH

	shapeOffX   = defaultShapeOffX
	shapeOffY   = defaultShapeOffY
	shapeScaleX = defaultShapeScaleX
	shapeScaleY = defaultShapeScaleY
)

// MapScale artık shape↔world dönüşümünde kullanılmıyor;
// sadece geriye dönük uyumluluk için tutuldu.
const MapScale = 1

// WorldMap ülke shape'lerinden üretilen dünya harita dokusunu yönetir.
type WorldMap struct {
	img          *ebiten.Image
	basePixels   []byte
	dispPixels   []byte
	regionAt     []uint16         // 0 = boş/deniz, 1..N = bölge indeksi
	regionIDs    []world.RegionID // regionIDs[0] = "" (boş)
	regionIdx    map[world.RegionID]uint16
	regionPx     map[world.RegionID][]int
	regionAnchor map[world.RegionID][2]int
	seaIdx       map[uint16]bool // deniz bölgesi indeksleri
	hasBgImage   bool
	ownerDirty   bool
	selected     world.RegionID
}

type countryShapeFile struct {
	Shapes []countryShape `json:"shapes"`
}

type countryShape struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Rings [][][2]int `json:"rings"`
}

func NewWorldMap(gs *state.GameState) *WorldMap {
	applyMapConfig(gs)
	wm := &WorldMap{
		img:          ebiten.NewImage(WorldW, WorldH),
		basePixels:   make([]byte, WorldW*WorldH*4),
		dispPixels:   make([]byte, WorldW*WorldH*4),
		regionAt:     make([]uint16, WorldW*WorldH),
		regionIDs:    []world.RegionID{""}, // indeks 0 = boş
		regionIdx:    make(map[world.RegionID]uint16),
		regionPx:     make(map[world.RegionID][]int),
		regionAnchor: make(map[world.RegionID][2]int),
		seaIdx:       make(map[uint16]bool),
	}

	// Fallback: düz okyanus mavisi. Senaryo PNG'si varsa aşağıda bunun üstüne yazılır.
	const oR, oG, oB byte = 28, 88, 168
	for i := 0; i < WorldW*WorldH; i++ {
		wm.basePixels[i*4] = oR
		wm.basePixels[i*4+1] = oG
		wm.basePixels[i*4+2] = oB
		wm.basePixels[i*4+3] = 255
	}
	if bgPixels, ok := loadPNGAsBasePixels(gs.ScenarioPath + "/maps/world_map_background.png"); ok {
		copy(wm.basePixels, bgPixels)
		wm.hasBgImage = true
		// log.Println("Arka plan harita resmi yüklendi")
	}

	shapesPath := ""
	if gs.ScenarioPath != "" {
		shapesPath = gs.ScenarioPath + "/data/country_shapes.json"
	}
	wm.buildCountryShapes(gs, loadCountryShapes(shapesPath))
	wm.buildSeaRegions(gs)
	wm.computeRegionAnchors()
	wm.applyOwnership(gs, "")
	return wm
}

func applyMapConfig(gs *state.GameState) {
	cfg := gs.MapConfig
	WorldW = defaultWorldW
	WorldH = defaultWorldH
	shapeOffX = defaultShapeOffX
	shapeOffY = defaultShapeOffY
	shapeScaleX = defaultShapeScaleX
	shapeScaleY = defaultShapeScaleY

	if cfg.WorldWidth != nil && *cfg.WorldWidth > 0 {
		WorldW = *cfg.WorldWidth
	}
	if cfg.WorldHeight != nil && *cfg.WorldHeight > 0 {
		WorldH = *cfg.WorldHeight
	}
	if cfg.ShapeOffsetX != nil {
		shapeOffX = *cfg.ShapeOffsetX
	}
	if cfg.ShapeOffsetY != nil {
		shapeOffY = *cfg.ShapeOffsetY
	}
	if cfg.ShapeScaleX != nil && *cfg.ShapeScaleX != 0 {
		shapeScaleX = *cfg.ShapeScaleX
	}
	if cfg.ShapeScaleY != nil && *cfg.ShapeScaleY != 0 {
		shapeScaleY = *cfg.ShapeScaleY
	}
}

// loadPNGAsBasePixels, arka plan PNG'sini WorldW×WorldH piksel tamponuna yükler.
// Shape koordinat uzayı (shapeW×shapeH) üzerinden doğru coğrafi hizalama sağlar.
func loadPNGAsBasePixels(path string) ([]byte, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		log.Printf("arka plan PNG decode hatası: %v", err)
		return nil, false
	}

	b := src.Bounds()
	srcW, srcH := b.Dx(), b.Dy()

	pixels := make([]byte, WorldW*WorldH*4)
	// Varsayılan: okyanus mavisi
	for i := 0; i < WorldW*WorldH; i++ {
		pixels[i*4], pixels[i*4+1], pixels[i*4+2], pixels[i*4+3] = 28, 88, 168, 255
	}
	// Hızlı erişim için NRGBA Pix dizisini dene, aksi halde At() kullan
	type pngGetter func(x, y int) (byte, byte, byte)
	var getPixel pngGetter
	switch img := src.(type) {
	case *image.NRGBA:
		getPixel = func(x, y int) (byte, byte, byte) {
			i := img.PixOffset(x, y)
			return img.Pix[i], img.Pix[i+1], img.Pix[i+2]
		}
	case *image.RGBA:
		getPixel = func(x, y int) (byte, byte, byte) {
			i := img.PixOffset(x, y)
			return img.Pix[i], img.Pix[i+1], img.Pix[i+2]
		}
	default:
		getPixel = func(x, y int) (byte, byte, byte) {
			r, g, bl, _ := src.At(x, y).RGBA()
			return byte(r >> 8), byte(g >> 8), byte(bl >> 8)
		}
	}

	// WorldW=PNG genişliği, WorldH=PNG yüksekliği → 1:1 piksel eşleme
	for gy := 0; gy < WorldH; gy++ {
		pngY := gy
		if pngY >= srcH {
			pngY = srcH - 1
		}
		for gx := 0; gx < WorldW; gx++ {
			pngX := gx
			if pngX >= srcW {
				pngX = srcW - 1
			}
			r, g, bl := getPixel(pngX, pngY)
			i := (gy*WorldW + gx) * 4
			pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] = r, g, bl, 255
		}
	}
	return pixels, true
}

func (wm *WorldMap) MarkDirty()                            { wm.ownerDirty = true }
func (wm *WorldMap) RegionPixels(rid world.RegionID) []int { return wm.regionPx[rid] }
func (wm *WorldMap) Image() *ebiten.Image                  { return wm.img }

func (wm *WorldMap) RegionAnchor(rid world.RegionID) (int, int, bool) {
	p, ok := wm.regionAnchor[rid]
	return p[0], p[1], ok
}

func (wm *WorldMap) Refresh(gs *state.GameState, selected world.RegionID) {
	if !wm.ownerDirty && wm.selected == selected {
		return
	}
	wm.applyOwnership(gs, selected)
	wm.ownerDirty = false
	wm.selected = selected
}

func (wm *WorldMap) RegionAt(wx, wy int) world.RegionID {
	if wx < 0 || wy < 0 || wx >= WorldW || wy >= WorldH {
		return ""
	}
	idx := wm.regionAt[wy*WorldW+wx]
	if idx == 0 {
		return ""
	}
	return wm.regionIDs[idx]
}

func loadCountryShapes(path string) map[string]countryShape {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("country shape dosyası okunamadı: %v", err)
		return nil
	}
	var payload countryShapeFile
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("country shape JSON parse edilemedi: %v", err)
		return nil
	}
	shapes := make(map[string]countryShape, len(payload.Shapes))
	for _, s := range payload.Shapes {
		if s.ID != "" && len(s.Rings) > 0 {
			shapes[s.ID] = s
		}
	}
	return shapes
}

func (wm *WorldMap) buildCountryShapes(gs *state.GameState, shapes map[string]countryShape) bool {
	if len(shapes) == 0 {
		return false
	}

	regionsByShape := make(map[string][]*world.Region)
	for _, r := range gs.Regions {
		if r.ShapeID != "" && !r.IsSea {
			regionsByShape[r.ShapeID] = append(regionsByShape[r.ShapeID], r)
		}
	}
	if len(regionsByShape) == 0 {
		return false
	}

	shapeIDs := make([]string, 0, len(regionsByShape))
	for shapeID := range regionsByShape {
		shapeIDs = append(shapeIDs, shapeID)
	}
	sort.Strings(shapeIDs)

	for _, shapeID := range shapeIDs {
		shape, ok := shapes[shapeID]
		if !ok {
			continue
		}
		regions := regionsByShape[shapeID]
		sort.Slice(regions, func(i, j int) bool { return regions[i].ID < regions[j].ID })
		for _, ring := range shape.Rings {
			wm.rasterizeRegionRing(gs, regions, ring)
		}
	}
	return true
}

func (wm *WorldMap) rasterizeRegionRing(_ *state.GameState, regions []*world.Region, ring [][2]int) {
	if len(regions) == 0 || len(ring) < 3 {
		return
	}
	scaled := make([][2]int, len(ring))
	for i, pt := range ring {
		scaled[i] = [2]int{
			int(shapeOffX + float64(pt[0])*shapeScaleX),
			int(shapeOffY + float64(pt[1])*shapeScaleY),
		}
	}
	minX, minY, maxX, maxY := intPolygonBounds(scaled)
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= WorldW {
		maxX = WorldW - 1
	}
	if maxY >= WorldH {
		maxY = WorldH - 1
	}
	for py := minY; py <= maxY; py++ {
		for px := minX; px <= maxX; px++ {
			if !pointInIntPolygon(float64(px)+0.5, float64(py)+0.5, scaled) {
				continue
			}
			pIdx := py*WorldW + px
			if wm.regionAt[pIdx] != 0 {
				continue
			}
			r := nearestShapeRegion(regions, px, py)

			ridx, ok := wm.regionIdx[r.ID]
			if !ok {
				ridx = uint16(len(wm.regionIDs))
				wm.regionIDs = append(wm.regionIDs, r.ID)
				wm.regionIdx[r.ID] = ridx
			}
			wm.regionAt[pIdx] = ridx
			wm.regionPx[r.ID] = append(wm.regionPx[r.ID], pIdx)

			// PNG arka plan varsa terrain rengi yazma — PNG pikseli korunur.
			if !wm.hasBgImage {
				col := terrainBaseColor(r.Terrain, px, py, string(r.ID))
				wm.basePixels[pIdx*4] = col.R
				wm.basePixels[pIdx*4+1] = col.G
				wm.basePixels[pIdx*4+2] = col.B
				wm.basePixels[pIdx*4+3] = 255
			}
		}
	}
}

// buildSeaRegions kara şekilleriyle atanmamış (okyanus) pikselleri
// multi-source BFS flood fill ile atar: her deniz bölgesi merkezinden başlayarak,
// komşu su piksellere yayılır (kara pikseller bariyerdir). Bu, deniz bölgelerinin
// kara engelleri geçerek diğer tarafa uzanmasını otomatik olarak önler.
func (wm *WorldMap) buildSeaRegions(gs *state.GameState) {
	var seaRegs []*world.Region
	for _, r := range gs.Regions {
		if r.IsSea && r.WorldX > 0 && r.WorldY > 0 {
			seaRegs = append(seaRegs, r)
		}
	}
	if len(seaRegs) == 0 {
		return
	}

	// ── 1. Tüm deniz bölgelerini önceden kayıt et ──────────────────
	for _, r := range seaRegs {
		if _, ok := wm.regionIdx[r.ID]; !ok {
			ridx := uint16(len(wm.regionIDs))
			wm.regionIDs = append(wm.regionIDs, r.ID)
			wm.regionIdx[r.ID] = ridx
			wm.seaIdx[ridx] = true
		}
	}

	// ── 2. BFS sırasında kullanılacak kuyruk türü ──────────────────
	type qEntry struct {
		pIdx int
		ridx uint16
	}
	queue := make([]qEntry, 0, WorldW*WorldH/2)

	// ── 3. Her deniz bölgesi için başlangıç pikseli (seed) bul ──────
	// Merkez koordinatlarına spiral arama yaparak, atanmamış (deniz) olan
	// ilk pikseli bulur ve BFS kuyruğuna ekler.
	for _, r := range seaRegs {
		ridx := wm.regionIdx[r.ID]
		seed := wm.findSeaSeed(int(shapeOffX+float64(r.WorldX)*shapeScaleX), int(shapeOffY+float64(r.WorldY)*shapeScaleY))
		if seed < 0 {
			seed = wm.findSeaSeed(r.WorldX, r.WorldY)
		}
		if seed < 0 {
			log.Printf("Deniz bölgesi seed pikseli bulunamadı: %s (wx=%d, wy=%d)",
				r.ID, r.WorldX, r.WorldY)
			continue
		}

		rid := r.ID
		wm.regionAt[seed] = ridx
		wm.regionPx[rid] = append(wm.regionPx[rid], seed)
		queue = append(queue, qEntry{seed, ridx})
	}

	// ── 4. Multi-source BFS flood fill (4-komşu): kara pikseller bariyerdir ──
	// Tüm seed pikselleri aynı anda kuyruğa girmiş olur, böylece Voronoi-benzeri
	// sınırlar oluşur, ancak kara bariyerleri geçilemez hale gelir.
	dx4 := [4]int{1, -1, 0, 0}
	dy4 := [4]int{0, 0, 1, -1}

	for i := 0; i < len(queue); i++ {
		e := queue[i]
		py, px := e.pIdx/WorldW, e.pIdx%WorldW
		rid := wm.regionIDs[e.ridx]

		for d := 0; d < 4; d++ {
			nx, ny := px+dx4[d], py+dy4[d]
			if nx < 0 || nx >= WorldW || ny < 0 || ny >= WorldH {
				continue
			}
			nIdx := ny*WorldW + nx
			if wm.regionAt[nIdx] != 0 {
				continue // Zaten atanmış (kara veya başka deniz bölgesi)
			}
			wm.regionAt[nIdx] = e.ridx
			wm.regionPx[rid] = append(wm.regionPx[rid], nIdx)
			queue = append(queue, qEntry{nIdx, e.ridx})
		}
	}

	// ── 5. Sınır tespiti ve basePixels'a bake ───────────────────────
	// Farklı deniz bölgelerine ait 4-komşu piksel çiftleri → sınır çizgisi
	const bR, bG, bB byte = 100, 160, 220 // açık mavi sınır rengi
	const bAlpha = byte(160)
	bake := func(pIdx int) {
		wm.basePixels[pIdx*4] = blend(wm.basePixels[pIdx*4], bR, bAlpha)
		wm.basePixels[pIdx*4+1] = blend(wm.basePixels[pIdx*4+1], bG, bAlpha)
		wm.basePixels[pIdx*4+2] = blend(wm.basePixels[pIdx*4+2], bB, bAlpha)
	}
	for py := 1; py < WorldH-1; py++ {
		for px := 1; px < WorldW-1; px++ {
			pIdx := py*WorldW + px
			cur := wm.regionAt[pIdx]
			if !wm.seaIdx[cur] {
				continue
			}
			// Sağ veya alt komşu farklı deniz bölgesindeyse bu piksel sınırda
			right := wm.regionAt[pIdx+1]
			down := wm.regionAt[pIdx+WorldW]
			if (wm.seaIdx[right] && right != cur) || (wm.seaIdx[down] && down != cur) {
				bake(pIdx)
				bake(pIdx + 1) // 2px genişlik için komşuyu da işaretle
				bake(pIdx + WorldW)
			}
		}
	}
}

func (wm *WorldMap) findSeaSeed(cx, cy int) int {
	if cx < 0 {
		cx = 0
	}
	if cx >= WorldW {
		cx = WorldW - 1
	}
	if cy < 0 {
		cy = 0
	}
	if cy >= WorldH {
		cy = WorldH - 1
	}

	// Mesafeye göre en yakın boş deniz pikselini bul
	seed := -1
	bestDist := int64(1<<63 - 1)
	const maxRadius = 200
	for dy := -maxRadius; dy <= maxRadius; dy++ {
		for dx := -maxRadius; dx <= maxRadius; dx++ {
			d := int64(dx*dx + dy*dy)
			if d > int64(maxRadius*maxRadius) {
				continue
			}
			nx, ny := cx+dx, cy+dy
			if nx < 0 || nx >= WorldW || ny < 0 || ny >= WorldH {
				continue
			}
			nIdx := ny*WorldW + nx
			if wm.regionAt[nIdx] != 0 {
				continue
			}
			if d < bestDist {
				bestDist = d
				seed = nIdx
			}
		}
	}
	return seed
}

func (wm *WorldMap) computeRegionAnchors() {
	for rid, pixels := range wm.regionPx {
		if len(pixels) == 0 {
			continue
		}

		sumX, sumY := int64(0), int64(0)
		for _, pIdx := range pixels {
			sumX += int64(pIdx % WorldW)
			sumY += int64(pIdx / WorldW)
		}
		cx := int(sumX / int64(len(pixels)))
		cy := int(sumY / int64(len(pixels)))

		best := pixels[0]
		bestDist := int64(1<<63 - 1)
		for _, pIdx := range pixels {
			px, py := pIdx%WorldW, pIdx/WorldW
			dx := int64(px - cx)
			dy := int64(py - cy)
			dist := dx*dx + dy*dy
			if dist < bestDist {
				best = pIdx
				bestDist = dist
			}
		}
		wm.regionAnchor[rid] = [2]int{best % WorldW, best / WorldW}
	}
}

// nearestShapeRegion piksel koordinatına en yakın bölgeyi döner.
// WorldX/WorldY orijinal 1920×1080 uzayında olduğundan MapScale ile çarpılır.
func nearestShapeRegion(regions []*world.Region, px, py int) *world.Region {
	best := regions[0]
	bestDist := int64(1<<63 - 1)
	for _, r := range regions {
		dx := int64(px) - int64(shapeOffX+float64(r.WorldX)*shapeScaleX)
		dy := int64(py) - int64(shapeOffY+float64(r.WorldY)*shapeScaleY)
		dist := dx*dx + dy*dy
		if dist < bestDist {
			best = r
			bestDist = dist
		}
	}
	return best
}

func (wm *WorldMap) applyOwnership(gs *state.GameState, selected world.RegionID) {
	copy(wm.dispPixels, wm.basePixels)

	factionColors := make(map[string][3]byte)
	for fid, f := range gs.Factions {
		factionColors[string(fid)] = f.Color
	}

	currentSeason := season.FromMonth(gs.Month)

	for rid, r := range gs.Regions {
		if r.IsSea {
			// Seçili deniz bölgesini belirgin açık mavi tintleyle vurgula
			if rid == selected {
				for _, pIdx := range wm.regionPx[rid] {
					wm.dispPixels[pIdx*4] = blend(wm.dispPixels[pIdx*4], 80, 120)
					wm.dispPixels[pIdx*4+1] = blend(wm.dispPixels[pIdx*4+1], 180, 120)
					wm.dispPixels[pIdx*4+2] = blend(wm.dispPixels[pIdx*4+2], 255, 120)
					wm.dispPixels[pIdx*4+3] = 255
				}
			}
			continue
		}

		// Mevsim tint'i uygula (sadece karalarda)
		var sr, sg, sb, sAlpha byte
		switch currentSeason {
		case season.SeasonWinter:
			sr, sg, sb, sAlpha = 240, 240, 255, 60 // Kar/buz
		case season.SeasonAutumn:
			sr, sg, sb, sAlpha = 200, 140, 60, 40 // Sararmış yapraklar
		case season.SeasonSpring:
			sr, sg, sb, sAlpha = 100, 200, 100, 30 // Canlı yeşil
		}

		if sAlpha > 0 {
			for _, pIdx := range wm.regionPx[rid] {
				wm.dispPixels[pIdx*4] = blend(wm.dispPixels[pIdx*4], sr, sAlpha)
				wm.dispPixels[pIdx*4+1] = blend(wm.dispPixels[pIdx*4+1], sg, sAlpha)
				wm.dispPixels[pIdx*4+2] = blend(wm.dispPixels[pIdx*4+2], sb, sAlpha)
			}
		}

		fc, ok := factionColors[r.OwnerID]
		if !ok && rid != selected {
			continue
		}
		alpha := byte(80)
		if rid == selected {
			alpha = 140
			if !ok {
				fc = [3]byte{245, 205, 80}
			}
		}
		for _, pIdx := range wm.regionPx[rid] {
			wm.dispPixels[pIdx*4] = blend(wm.dispPixels[pIdx*4], fc[0], alpha)
			wm.dispPixels[pIdx*4+1] = blend(wm.dispPixels[pIdx*4+1], fc[1], alpha)
			wm.dispPixels[pIdx*4+2] = blend(wm.dispPixels[pIdx*4+2], fc[2], alpha)
			wm.dispPixels[pIdx*4+3] = 255
		}
	}

	wm.drawRegionBorders(gs, selected)
	wm.img.WritePixels(wm.dispPixels)
}

func (wm *WorldMap) drawRegionBorders(gs *state.GameState, selected world.RegionID) {
	for py := 1; py < WorldH-1; py++ {
		for px := 1; px < WorldW-1; px++ {
			pIdx := py*WorldW + px
			curIdx := wm.regionAt[pIdx]
			if curIdx == 0 {
				continue
			}
			cur := wm.regionIDs[curIdx]
			curRegion := gs.Regions[cur]
			if curRegion == nil || curRegion.IsSea {
				continue
			}
			isBorder := wm.regionAt[pIdx+1] != curIdx ||
				wm.regionAt[pIdx+WorldW] != curIdx
			if !isBorder {
				continue
			}

			if cur == selected {
				wm.setOverlayPixel(pIdx, 255, 222, 72, 245)
				continue
			}
			wm.setOverlayPixel(pIdx, 35, 22, 10, 170)
		}
	}
}

func (wm *WorldMap) setOverlayPixel(pIdx int, r, g, b, a byte) {
	wm.dispPixels[pIdx*4] = r
	wm.dispPixels[pIdx*4+1] = g
	wm.dispPixels[pIdx*4+2] = b
	wm.dispPixels[pIdx*4+3] = a
}

func intPolygonBounds(poly [][2]int) (int, int, int, int) {
	minX, minY := poly[0][0], poly[0][1]
	maxX, maxY := minX, minY
	for _, p := range poly[1:] {
		if p[0] < minX {
			minX = p[0]
		}
		if p[1] < minY {
			minY = p[1]
		}
		if p[0] > maxX {
			maxX = p[0]
		}
		if p[1] > maxY {
			maxY = p[1]
		}
	}
	return minX, minY, maxX, maxY
}

func pointInIntPolygon(x, y float64, poly [][2]int) bool {
	inside := false
	j := len(poly) - 1
	for i := range poly {
		xi, yi := float64(poly[i][0]), float64(poly[i][1])
		xj, yj := float64(poly[j][0]), float64(poly[j][1])
		if ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}
	return inside
}

// --- Arazi rengi ---

func terrainBaseColor(terrain world.TerrainType, px, py int, regionID string) color.RGBA {
	regionHash := 0
	for i, c := range regionID {
		regionHash += int(c) * (i + 7)
	}
	// Hafif piksel bazlı doku gürültüsü
	micro := float64(((px*3+py*7+regionHash*13)&0xFF)-128) / 255.0 * 0.07

	switch terrain {
	case world.TerrainSea:
		return color.RGBA{30, 92, 174, 255}
	}

	// Geçici kampanya görünümünde her kara bölgesi net ayırt edilsin.
	palette := []color.RGBA{
		{172, 91, 70, 255},
		{190, 142, 54, 255},
		{112, 151, 78, 255},
		{74, 145, 125, 255},
		{76, 123, 177, 255},
		{131, 104, 180, 255},
		{176, 96, 144, 255},
		{178, 116, 62, 255},
		{136, 153, 69, 255},
		{82, 152, 157, 255},
		{107, 129, 190, 255},
		{158, 107, 164, 255},
		{194, 105, 88, 255},
		{160, 160, 84, 255},
		{88, 138, 102, 255},
		{150, 118, 76, 255},
	}
	base := palette[regionHash%len(palette)]

	switch terrain {
	case world.TerrainForest:
		base = mixColor(base, color.RGBA{34, 96, 48, 255}, 70)
	case world.TerrainMountain:
		base = mixColor(base, color.RGBA{144, 134, 116, 255}, 92)
	case world.TerrainPass:
		base = mixColor(base, color.RGBA{178, 150, 84, 255}, 82)
	case world.TerrainCoast:
		base = mixColor(base, color.RGBA{112, 172, 158, 255}, 78)
	default:
		base = mixColor(base, color.RGBA{154, 150, 92, 255}, 34)
	}

	return color.RGBA{
		R: clampByte(float64(base.R) * (1 + micro)),
		G: clampByte(float64(base.G) * (1 + micro)),
		B: clampByte(float64(base.B) * (1 + micro)),
		A: 255,
	}
}

// --- Yardımcılar ---

func mixColor(a, b color.RGBA, t uint8) color.RGBA {
	return color.RGBA{
		R: blend(a.R, b.R, t),
		G: blend(a.G, b.G, t),
		B: blend(a.B, b.B, t),
		A: 255,
	}
}

func blend(a, b byte, t uint8) byte {
	tF := float64(t) / 255.0
	return byte(float64(a)*(1-tF) + float64(b)*tF)
}

func clampByte(v float64) byte {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return byte(v)
}

var whitePixelImg *ebiten.Image

func drawPixelRect(screen *ebiten.Image, x, y, size float32, col color.RGBA) {
	if whitePixelImg == nil {
		whitePixelImg = ebiten.NewImage(1, 1)
		whitePixelImg.Fill(color.White)
	}
	r := float32(col.R) / 255
	g := float32(col.G) / 255
	b := float32(col.B) / 255
	a := float32(col.A) / 255
	vs := []ebiten.Vertex{
		{DstX: x, DstY: y, SrcX: 0.5, SrcY: 0.5, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: x + size, DstY: y, SrcX: 0.5, SrcY: 0.5, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: x + size, DstY: y + size, SrcX: 0.5, SrcY: 0.5, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: x, DstY: y + size, SrcX: 0.5, SrcY: 0.5, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
	}
	screen.DrawTriangles(vs, []uint16{0, 1, 2, 0, 2, 3}, whitePixelImg, nil)
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func wcX(v int) float64 { return shapeOffX + float64(v)*shapeScaleX }
func wcY(v int) float64 { return shapeOffY + float64(v)*shapeScaleY }
