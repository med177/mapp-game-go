package world

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
)

const terrainAreaRegionPrefix = "area::"

// TerrainAreaRegionID returns the runtime graph ID for a terrain area.
// Terrain areas are stored separately in scenario data; this ID only exists
// while they are materialized as movement/selection nodes.
func TerrainAreaRegionID(areaID string) RegionID {
	return RegionID(terrainAreaRegionPrefix + areaID)
}

// TerrainAreaFragmentRegionID returns the runtime ID for the portion of a
// passable terrain area that lies inside one base region.
func TerrainAreaFragmentRegionID(areaID string, baseRegionID RegionID) RegionID {
	return RegionID(terrainAreaRegionPrefix + areaID + "::" + string(baseRegionID))
}

// TerrainAreaIsPassable is the canonical passability rule for a painted area.
// Terrain is only the area's label/visual identity; a non-zero move cost makes
// the area traversable regardless of whether its terrain label is mountain,
// lake, desert, or another type.
func TerrainAreaIsPassable(area TerrainArea) bool {
	return area.MoveCost != 0
}

// SyncTerrainAreaRegions materializes painted areas as lightweight runtime
// nodes. ParentRegionID is only a legacy/runtime hint; terrain areas are
// stored as top-level polygons in terrain_areas.json.
func SyncTerrainAreaRegions(regions map[RegionID]*Region, areas []TerrainArea) {
	for _, region := range regions {
		if region == nil {
			continue
		}
		filtered := region.Neighbors[:0]
		for _, neighbor := range region.Neighbors {
			if len(neighbor) < len(terrainAreaRegionPrefix) || string(neighbor)[:len(terrainAreaRegionPrefix)] != terrainAreaRegionPrefix {
				filtered = append(filtered, neighbor)
			}
		}
		region.Neighbors = filtered
	}
	for id, region := range regions {
		if region != nil && region.IsTerrainArea {
			delete(regions, id)
		}
	}
	for i := range areas {
		area := &areas[i]
		if !area.HasGeometry() || area.ID == "" {
			continue
		}
		id := TerrainAreaRegionID(area.ID)
		cx, cy := area.Center()
		terrain := area.Terrain
		if terrain == "" {
			// Top-level terrain areas no longer require a parent region. Keep
			// an empty type usable while the editor is still configuring it.
			terrain = TerrainPlain
		}
		regions[id] = &Region{ID: id, Name: area.Name, NameTR: area.Name, Terrain: terrain,
			OwnerID: "", WorldX: cx, WorldY: cy, IsTerrainArea: true,
			ParentRegionID: area.ParentRegionID, TerrainAreaID: area.ID,
			IsLocked: area.MoveCost == 0}
		if regions[id].NameTR == "" {
			regions[id].NameTR = "Arazi Alanı"
		}
	}
	for i := range areas {
		area := &areas[i]
		id := TerrainAreaRegionID(area.ID)
		child := regions[id]
		parent := regions[area.ParentRegionID]
		if child == nil {
			continue
		}
		if TerrainAreaIsPassable(*area) {
			// Passable areas are linked after the map raster is split into
			// base-region fragments.
			continue
		}
		if parent != nil {
			child.Neighbors = append(child.Neighbors, parent.ID)
			parent.Neighbors = appendUniqueRegionID(parent.Neighbors, child.ID)
		}
		for _, extraID := range area.ExtraNeighbors {
			extra := regions[extraID]
			if extra == nil || extra.ID == child.ID {
				continue
			}
			child.Neighbors = appendUniqueRegionID(child.Neighbors, extra.ID)
			extra.Neighbors = appendUniqueRegionID(extra.Neighbors, child.ID)
		}
	}
}

func UpdateTerrainAreaRegionOwners(regions map[RegionID]*Region) {
	for _, region := range regions {
		if region == nil || !region.IsTerrainArea {
			continue
		}
		// Terrain areas are neutral movement modifiers, not owned territory.
		// ParentRegionID is retained only for geometry/runtime linkage.
		region.OwnerID = ""
	}
}

func appendUniqueRegionID(ids []RegionID, id RegionID) []RegionID {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}

// TerrainArea is a movement modifier painted inside a parent land region.
// MoveCost is expressed as an extra cost: 0 blocks entry, -1 consumes one
// additional movement point, and -2 consumes two.
type TerrainArea struct {
	ID string `json:"id"`
	// ParentRegionID is retained only as a runtime compatibility hint for old
	// data. New terrain areas are top-level and never serialize this field.
	ParentRegionID RegionID    `json:"-"`
	Name           string      `json:"name,omitempty"`
	Terrain        TerrainType `json:"terrain"`
	MoveCost       int         `json:"move_cost"`
	Polygons       [][][2]int  `json:"polygons,omitempty"`
	// Cells is a non-persistent runtime helper retained for editor internals;
	// scenario source data is polygon-only.
	Cells              [][2]int `json:"-"`
	RuntimeCenterX     int      `json:"-"`
	RuntimeCenterY     int      `json:"-"`
	RuntimeCenterValid bool     `json:"-"`
	// AttritionCost is the percentage of unit HP (0-100) an army loses when it
	// enters this area (e.g. desert heat, mountain fatigue). 0 means no wear.
	AttritionCost int `json:"attrition_cost,omitempty"`
	// ExtraNeighbors are manually assigned links beyond the automatic
	// parent/touching-sibling adjacency (e.g. a river area linking a bridge).
	ExtraNeighbors []RegionID `json:"extra_neighbors,omitempty"`
}

func (a TerrainArea) HasGeometry() bool {
	return len(a.Polygons) > 0
}

func (a TerrainArea) Center() (int, int) {
	if a.RuntimeCenterValid {
		return a.RuntimeCenterX, a.RuntimeCenterY
	}
	if x, y, ok := a.PolygonCenter(); ok {
		return x, y
	}
	return 0, 0
}

// PolygonCenter returns the geometric centroid in polygon/raster coordinates.
// RuntimeCenter is intentionally not considered here so render code can
// convert this value to the coordinate system used by Region.WorldX/Y.
func (a TerrainArea) PolygonCenter() (int, int, bool) {
	var weightedX, weightedY, totalArea float64
	var fallbackX, fallbackY, fallbackN int
	for _, polygon := range a.Polygons {
		if len(polygon) < 3 {
			continue
		}
		var area2, cx, cy float64
		for i, point := range polygon {
			next := polygon[(i+1)%len(polygon)]
			cross := float64(point[0]*next[1] - next[0]*point[1])
			area2 += cross
			cx += float64(point[0]+next[0]) * cross
			cy += float64(point[1]+next[1]) * cross
			fallbackX += point[0]
			fallbackY += point[1]
			fallbackN++
		}
		if math.Abs(area2) > 0.0001 {
			weightedX += cx / 3
			weightedY += cy / 3
			totalArea += area2
		}
	}
	if math.Abs(totalArea) > 0.0001 {
		centerX := int(math.Round(weightedX / totalArea))
		centerY := int(math.Round(weightedY / totalArea))
		if terrainAreaPointInside(a.Polygons, centerX, centerY) {
			return centerX, centerY, true
		}

		// İçbükey poligonlarda geometrik centroid poligonun dışında kalabilir.
		// Merkez işareti ve seçim odağı mutlaka boyalı alanın içinde kalsın diye,
		// centroid'e en yakın raster hücresini seç.
		if x, y, ok := nearestTerrainAreaCell(a.Polygons, weightedX/totalArea, weightedY/totalArea); ok {
			return x, y, true
		}
		return centerX, centerY, true
	}
	if fallbackN == 0 {
		return 0, 0, false
	}
	return fallbackX / fallbackN, fallbackY / fallbackN, true
}

func terrainAreaPointInside(polygons [][][2]int, x, y int) bool {
	for _, polygon := range polygons {
		if PointInPolygon(float64(x)+0.5, float64(y)+0.5, polygon) {
			return true
		}
	}
	return false
}

func nearestTerrainAreaCell(polygons [][][2]int, targetX, targetY float64) (int, int, bool) {
	bestDistance := math.Inf(1)
	bestX, bestY := 0, 0
	found := false
	for _, polygon := range polygons {
		if len(polygon) < 3 {
			continue
		}
		minX, minY := polygon[0][0], polygon[0][1]
		maxX, maxY := minX, minY
		for _, point := range polygon[1:] {
			if point[0] < minX {
				minX = point[0]
			}
			if point[1] < minY {
				minY = point[1]
			}
			if point[0] > maxX {
				maxX = point[0]
			}
			if point[1] > maxY {
				maxY = point[1]
			}
		}
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				if !PointInPolygon(float64(x)+0.5, float64(y)+0.5, polygon) {
					continue
				}
				dx, dy := float64(x)+0.5-targetX, float64(y)+0.5-targetY
				distance := dx*dx + dy*dy
				if distance < bestDistance {
					bestDistance = distance
					bestX, bestY = x, y
					found = true
				}
			}
		}
	}
	return bestX, bestY, found
}

func (a TerrainArea) Contains(x, y int) bool {
	if len(a.Polygons) > 0 {
		px, py := float64(x)+0.5, float64(y)+0.5
		for _, polygon := range a.Polygons {
			if PointInPolygon(px, py, polygon) {
				return true
			}
		}
		return false
	}
	return false
}

func PointInPolygon(x, y float64, polygon [][2]int) bool {
	if len(polygon) < 3 {
		return false
	}
	inside := false
	for i, j := 0, len(polygon)-1; i < len(polygon); i++ {
		xi, yi := float64(polygon[i][0]), float64(polygon[i][1])
		xj, yj := float64(polygon[j][0]), float64(polygon[j][1])
		intersects := (yi > y) != (yj > y) && x < (xj-xi)*(y-yi)/(yj-yi)+xi
		if intersects {
			inside = !inside
		}
		j = i
	}
	return inside
}

// MergeTerrainAreaPolygons, yeni poligon mevcut poligonlardan biriyle aynı
// raster hücrelerini kaplıyorsa birleşik alanın dış sınırlarını üretir.
// Kesişmeyen poligonlarda intersects false döner; çağıran mevcut listeye yeni
// halkayı ayrı olarak ekleyebilir.
func MergeTerrainAreaPolygons(existing [][][2]int, added [][2]int) (merged [][][2]int, intersects bool) {
	if len(added) < 3 {
		return nil, false
	}
	addedCells := terrainAreaPolygonCells(added)
	if len(addedCells) == 0 {
		return nil, false
	}
	existingCells := make(map[[2]int]struct{})
	for _, polygon := range existing {
		for cell := range terrainAreaPolygonCells(polygon) {
			existingCells[cell] = struct{}{}
		}
	}
	for cell := range addedCells {
		if _, ok := existingCells[cell]; ok {
			intersects = true
			break
		}
	}
	if !intersects {
		return nil, false
	}

	allCells := existingCells
	for cell := range addedCells {
		allCells[cell] = struct{}{}
	}
	return terrainAreaCellContours(allCells), true
}

func terrainAreaPolygonCells(polygon [][2]int) map[[2]int]struct{} {
	cells := make(map[[2]int]struct{})
	if len(polygon) < 3 {
		return cells
	}
	minX, minY := polygon[0][0], polygon[0][1]
	maxX, maxY := minX, minY
	for _, point := range polygon[1:] {
		if point[0] < minX {
			minX = point[0]
		}
		if point[1] < minY {
			minY = point[1]
		}
		if point[0] > maxX {
			maxX = point[0]
		}
		if point[1] > maxY {
			maxY = point[1]
		}
	}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if PointInPolygon(float64(x)+0.5, float64(y)+0.5, polygon) {
				cells[[2]int{x, y}] = struct{}{}
			}
		}
	}
	return cells
}

type terrainAreaBoundaryEdge struct {
	from [2]int
	to   [2]int
}

func terrainAreaCellContours(cells map[[2]int]struct{}) [][][2]int {
	if len(cells) == 0 {
		return nil
	}
	cellList := make([][2]int, 0, len(cells))
	for cell := range cells {
		cellList = append(cellList, cell)
	}
	sort.Slice(cellList, func(i, j int) bool {
		if cellList[i][1] == cellList[j][1] {
			return cellList[i][0] < cellList[j][0]
		}
		return cellList[i][1] < cellList[j][1]
	})

	edges := make([]terrainAreaBoundaryEdge, 0, len(cells)*2)
	addEdge := func(cell, neighbor [2]int, from, to [2]int) {
		if _, exists := cells[neighbor]; !exists {
			edges = append(edges, terrainAreaBoundaryEdge{from: from, to: to})
		}
	}
	for _, cell := range cellList {
		x, y := cell[0], cell[1]
		addEdge(cell, [2]int{x, y - 1}, [2]int{x, y}, [2]int{x + 1, y})
		addEdge(cell, [2]int{x + 1, y}, [2]int{x + 1, y}, [2]int{x + 1, y + 1})
		addEdge(cell, [2]int{x, y + 1}, [2]int{x + 1, y + 1}, [2]int{x, y + 1})
		addEdge(cell, [2]int{x - 1, y}, [2]int{x, y + 1}, [2]int{x, y})
	}

	used := make([]bool, len(edges))
	contours := make([][][2]int, 0, 1)
	for startIndex := range edges {
		if used[startIndex] {
			continue
		}
		start := edges[startIndex].from
		current := startIndex
		contour := make([][2]int, 0, 16)
		for {
			if used[current] {
				break
			}
			used[current] = true
			if len(contour) == 0 {
				contour = append(contour, edges[current].from)
			}
			contour = append(contour, edges[current].to)
			end := edges[current].to
			if end == start {
				break
			}
			next := -1
			for i := range edges {
				if !used[i] && edges[i].from == end {
					next = i
					break
				}
			}
			if next < 0 {
				break
			}
			current = next
		}
		if len(contour) >= 4 && contour[len(contour)-1] == start {
			contour = contour[:len(contour)-1]
			contour = simplifyTerrainAreaContour(contour)
			if len(contour) >= 3 {
				contours = append(contours, contour)
			}
		}
	}
	return contours
}

func simplifyTerrainAreaContour(contour [][2]int) [][2]int {
	if len(contour) < 3 {
		return contour
	}
	result := make([][2]int, 0, len(contour))
	for _, point := range contour {
		result = append(result, point)
		for len(result) >= 3 {
			a := result[len(result)-3]
			b := result[len(result)-2]
			c := result[len(result)-1]
			if (b[0]-a[0])*(c[1]-b[1]) != (b[1]-a[1])*(c[0]-b[0]) {
				break
			}
			result = append(result[:len(result)-2], c)
		}
	}
	return result
}

// TerrainAreaMovementCost returns the additional movement points at a point.
// The most restrictive overlapping area wins; a zero-cost area always blocks.
func TerrainAreaMovementCost(areas []TerrainArea, parent RegionID, x, y int) (cost int, blocked bool) {
	for _, area := range areas {
		if area.ParentRegionID != "" && area.ParentRegionID != parent || !area.Contains(x, y) {
			continue
		}
		if area.MoveCost == 0 {
			return 0, true
		}
		if area.MoveCost < cost {
			cost = area.MoveCost
		}
	}
	return cost, false
}

func LoadTerrainAreas(path string, regions map[RegionID]*Region) ([]TerrainArea, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("terrain_areas dosyası okunamadı: %w", err)
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) || bytes.Equal(trimmed, []byte("[]")) {
		return nil, nil
	}
	var areas []TerrainArea
	if err := json.Unmarshal(data, &areas); err != nil {
		return nil, fmt.Errorf("terrain_areas JSON parse hatası: %w", err)
	}
	valid := areas[:0]
	seen := make(map[string]bool, len(areas))
	for _, area := range areas {
		if area.ID == "" || seen[area.ID] || !area.HasGeometry() {
			continue
		}
		seen[area.ID] = true
		valid = append(valid, area)
	}
	return valid, nil
}
