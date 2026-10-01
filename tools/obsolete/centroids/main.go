// Bölge merkez koordinatlarını (world_x, world_y) gerçek polygon piksellerinden hesaplar.
// Kullanım: go run tools/centroids/main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
)

const (
	worldW = 1920
	worldH = 1080
)

type Region struct {
	ID      string `json:"id"`
	ShapeID string `json:"shape_id,omitempty"`
	IsSea   bool   `json:"is_sea"`
	WorldX  int    `json:"world_x"`
	WorldY  int    `json:"world_y"`
	Raw     []byte `json:"-"`
}

type ShapeFile struct {
	Shapes []Shape `json:"shapes"`
}

type Shape struct {
	ID    string     `json:"id"`
	Rings [][][2]int `json:"rings"`
}

func main() {
	// regions.json ham olarak oku (tüm alanları koru)
	rawData, err := os.ReadFile("assets/data/regions.json")
	if err != nil {
		log.Fatal(err)
	}
	var rawList []json.RawMessage
	if err := json.Unmarshal(rawData, &rawList); err != nil {
		log.Fatal(err)
	}

	type regionInfoLocal = RegionInfo
	regions := make([]regionInfoLocal, len(rawList))
	for i, raw := range rawList {
		if err := json.Unmarshal(raw, &regions[i]); err != nil {
			log.Fatal(err)
		}
	}

	// country_shapes.json yükle
	shapeData, err := os.ReadFile("assets/data/generated/country_shapes.json")
	if err != nil {
		log.Fatal(err)
	}
	var sf ShapeFile
	if err := json.Unmarshal(shapeData, &sf); err != nil {
		log.Fatal(err)
	}
	shapes := make(map[string]Shape, len(sf.Shapes))
	for _, s := range sf.Shapes {
		shapes[s.ID] = s
	}

	// Kara bölgeleri shape'e göre grupla
	regionsByShape := make(map[string][]RegionInfo)
	for _, r := range regions {
		if !r.IsSea && r.ShapeID != "" {
			regionsByShape[r.ShapeID] = append(regionsByShape[r.ShapeID], r)
		}
	}
	for k := range regionsByShape {
		grp := regionsByShape[k]
		sort.Slice(grp, func(i, j int) bool { return grp[i].ID < grp[j].ID })
		regionsByShape[k] = grp
	}

	// Her bölge için piksel centroid hesapla
	type centData struct {
		sumX, sumY, count int64
	}
	centroids := make(map[string]*centData)

	for shapeID, shape := range shapes {
		grp, ok := regionsByShape[shapeID]
		if !ok || len(grp) == 0 {
			continue
		}
		for k := range grp {
			centroids[grp[k].ID] = &centData{}
		}

		for _, ring := range shape.Rings {
			if len(ring) < 3 {
				continue
			}
			minX, minY, maxX, maxY := polygonBounds(ring)
			clamp(&minX, &minY, &maxX, &maxY)

			for py := minY; py <= maxY; py++ {
				for px := minX; px <= maxX; px++ {
					if !pointInPoly(float64(px)+0.5, float64(py)+0.5, ring) {
						continue
					}
					// Nearest region (Voronoi)
					rid := nearestRegion(grp, px, py)
					if rid == "" {
						continue
					}
					c := centroids[rid]
					c.sumX += int64(px)
					c.sumY += int64(py)
					c.count++
				}
			}
		}
	}

	// Centroid'leri raporla ve regions.json güncelle
	updated := 0
	for i, raw := range rawList {
		rid := regions[i].ID
		c, ok := centroids[rid]
		if !ok || c.count == 0 {
			continue
		}
		cx := int(math.Round(float64(c.sumX) / float64(c.count)))
		cy := int(math.Round(float64(c.sumY) / float64(c.count)))

		oldX, oldY := regions[i].WorldX, regions[i].WorldY
		if cx == oldX && cy == oldY {
			continue
		}

		// JSON alanını güncelle
		var m map[string]json.RawMessage
		if err := json.Unmarshal(raw, &m); err != nil {
			log.Fatal(err)
		}
		m["world_x"], _ = json.Marshal(cx)
		m["world_y"], _ = json.Marshal(cy)
		newRaw, err := json.Marshal(m)
		if err != nil {
			log.Fatal(err)
		}
		rawList[i] = newRaw
		fmt.Printf("%-25s (%3d,%3d) → (%3d,%3d)\n", rid, oldX, oldY, cx, cy)
		updated++
	}

	if updated == 0 {
		fmt.Println("Değişecek koordinat yok.")
		return
	}

	// Yaz
	out, err := json.MarshalIndent(rawList, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	// rawList içindeki elemanlar []byte olarak marshal edilir, bunu []map şeklinde yeniden oluştur
	var finalList []json.RawMessage
	if err := json.Unmarshal(out, &finalList); err != nil {
		log.Fatal(err)
	}
	finalOut, _ := json.MarshalIndent(finalList, "", "  ")
	if err := os.WriteFile("assets/data/regions.json", finalOut, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n%d bölge güncellendi → assets/data/regions.json\n", updated)
}

type RegionInfo struct {
	ID      string `json:"id"`
	ShapeID string `json:"shape_id,omitempty"`
	IsSea   bool   `json:"is_sea"`
	WorldX  int    `json:"world_x"`
	WorldY  int    `json:"world_y"`
}

func nearestRegion(grp []RegionInfo, px, py int) string {
	best := ""
	bestDist := int64(math.MaxInt64)
	for _, r := range grp {
		dx := int64(px - r.WorldX)
		dy := int64(py - r.WorldY)
		d := dx*dx + dy*dy
		if d < bestDist {
			best = r.ID
			bestDist = d
		}
	}
	return best
}

func polygonBounds(poly [][2]int) (int, int, int, int) {
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

func clamp(minX, minY, maxX, maxY *int) {
	if *minX < 0 {
		*minX = 0
	}
	if *minY < 0 {
		*minY = 0
	}
	if *maxX >= worldW {
		*maxX = worldW - 1
	}
	if *maxY >= worldH {
		*maxY = worldH - 1
	}
}

func pointInPoly(x, y float64, poly [][2]int) bool {
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
