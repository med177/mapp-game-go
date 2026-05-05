package render

import (
	"encoding/json"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"os"
	"sort"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

// shape koordinat uzayı boyutları (Python araçlarıyla eşleşmeli)
const (
	shapeW = 1828
	shapeH = 997
)

// PNG arka plan kalibrasyon sabitleri.
// PNG'nin gösterdiği coğrafi alan shape koordinat formülüyle farklıysa
// bu 4 değeri ayarlayın: lon_to_px=20*lon+476, lat_to_py=-18.98*lat+1234.8
const (
	// Shape (0,0) noktasının karşılık geldiği PNG pikseli (sol-üst)
	bgPx0 float64 = 22
	bgPy0 float64 = 12
	// Shape (shapeW, shapeH) noktasının karşılık geldiği PNG pikseli (sağ-alt)
	bgPx1 float64 = 2794
	bgPy1 float64 = 1524
)

const (
	MapScale = 2
	WorldW   = 1920 * MapScale
	WorldH   = 1080 * MapScale
)

// WorldMap ülke shape'lerinden üretilen dünya harita dokusunu yönetir.
type WorldMap struct {
	img        *ebiten.Image
	basePixels []byte
	dispPixels []byte
	regionAt   []uint16         // 0 = boş/deniz, 1..N = bölge indeksi
	regionIDs  []world.RegionID // regionIDs[0] = "" (boş)
	regionIdx  map[world.RegionID]uint16
	regionPx   map[world.RegionID][]int
	hasBgImage bool
	ownerDirty bool
	selected   world.RegionID
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
	wm := &WorldMap{
		img:        ebiten.NewImage(WorldW, WorldH),
		basePixels: make([]byte, WorldW*WorldH*4),
		dispPixels: make([]byte, WorldW*WorldH*4),
		regionAt:   make([]uint16, WorldW*WorldH),
		regionIDs:  []world.RegionID{""}, // indeks 0 = boş
		regionIdx:  make(map[world.RegionID]uint16),
		regionPx:   make(map[world.RegionID][]int),
	}

	if true { // arka plan resmi devre dışı
		if bgPixels, ok := loadPNGAsBasePixels("assets/maps/world_map_background.png"); ok {
			copy(wm.basePixels, bgPixels)
			wm.hasBgImage = true
			log.Println("Arka plan harita resmi yüklendi")
		}
	} else {
		// Fallback: düz okyanus mavisi
		const oR, oG, oB byte = 28, 88, 168
		for i := 0; i < WorldW*WorldH; i++ {
			wm.basePixels[i*4] = oR
			wm.basePixels[i*4+1] = oG
			wm.basePixels[i*4+2] = oB
			wm.basePixels[i*4+3] = 255
		}
	}

	wm.buildCountryShapes(gs, loadCountryShapes("assets/data/generated/country_shapes.json"))
	wm.applyOwnership(gs, "")
	return wm
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

	// game piksel (gx,gy) → shape koordinat (sx,sy) → PNG piksel via kalibrasyon
	scaleX := (bgPx1 - bgPx0) / float64(shapeW)
	scaleY := (bgPy1 - bgPy0) / float64(shapeH)

	for gy := 0; gy < WorldH; gy++ {
		sy := gy / MapScale
		if sy >= shapeH {
			continue
		}
		pngY := int(bgPy0 + float64(sy)*scaleY)
		if pngY < 0 {
			pngY = 0
		} else if pngY >= srcH {
			pngY = srcH - 1
		}
		for gx := 0; gx < WorldW; gx++ {
			sx := gx / MapScale
			if sx >= shapeW {
				continue
			}
			pngX := int(bgPx0 + float64(sx)*scaleX)
			if pngX < 0 {
				pngX = 0
			} else if pngX >= srcW {
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
	// Koordinatları MapScale ile büyüt
	scaled := make([][2]int, len(ring))
	for i, pt := range ring {
		scaled[i] = [2]int{pt[0] * MapScale, pt[1] * MapScale}
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

// nearestShapeRegion piksel koordinatına en yakın bölgeyi döner.
// WorldX/WorldY orijinal 1920×1080 uzayında olduğundan MapScale ile çarpılır.
func nearestShapeRegion(regions []*world.Region, px, py int) *world.Region {
	best := regions[0]
	bestDist := int64(1<<63 - 1)
	for _, r := range regions {
		dx := int64(px - r.WorldX*MapScale)
		dy := int64(py - r.WorldY*MapScale)
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

	for rid, r := range gs.Regions {
		if r.IsSea {
			continue
		}
		fc, ok := factionColors[r.OwnerID]
		if !ok && rid != selected {
			continue
		}
		alpha := byte(36)
		if rid == selected {
			alpha = 96
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
				wm.regionAt[pIdx-1] != curIdx ||
				wm.regionAt[pIdx+WorldW] != curIdx ||
				wm.regionAt[pIdx-WorldW] != curIdx
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

// wc bölge WorldX/WorldY koordinatlarını harita dokusunun ölçeğine dönüştürür.
func wc(v int) float64 { return float64(v * MapScale) }
