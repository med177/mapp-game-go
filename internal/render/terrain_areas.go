package render

import (
	"image/color"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (r *Renderer) drawTerrainAreas(screen *ebiten.Image) {
	if r == nil || r.gs == nil {
		return
	}
	if len(r.gs.TerrainAreas) == 0 {
		return
	}
	selectedAreaID := ""
	if selected := r.gs.Regions[r.editSelectedRegion]; selected != nil && selected.IsTerrainArea {
		selectedAreaID = selected.TerrainAreaID
	}
	key := terrainAreaRenderKey(r.gs, selectedAreaID, r.editSelectedRegion)
	rebuild := r.terrainAreaImage == nil || r.terrainAreaImage.Bounds().Dx() != WorldW || r.terrainAreaImage.Bounds().Dy() != WorldH || r.terrainAreaKey != key
	if rebuild {
		r.terrainAreaImage = ebiten.NewImage(WorldW, WorldH)
		r.terrainAreaImage.Clear()
		r.terrainAreaKey = key
	}

	drawArea := func(area world.TerrainArea) {
		col := terrainAreaColor(area.Terrain)
		passable := world.TerrainAreaIsPassable(area)
		if passable {
			// Geçilebilir alan daha saydam çizilir; border ve altındaki harita
			// görünür kalır.
			col.A = terrainAreaPassableAlpha
		} else {
			// Geçilemeyen alan parlak bir terrain rengi üretmesin; koyu
			// grimsi-siyah bir örtü olarak kalsın.
			col = tintTerrainAreaColor(col, 0.32)
			col.A = terrainAreaBlockedAlpha
		}
		if area.ParentRegionID == r.editSelectedRegion || area.ID == selectedAreaID {
			if passable {
				col.A = terrainAreaSelectedPassableAlpha
			} else {
				col.A = terrainAreaSelectedBlockedAlpha
			}
		}
		if len(area.Polygons) > 0 {
			var path vector.Path
			hasPolygon := false
			for _, polygon := range area.Polygons {
				if len(polygon) < 3 {
					continue
				}
				hasPolygon = true
				x0, y0 := float64(polygon[0][0]), float64(polygon[0][1])
				path.MoveTo(float32(x0), float32(y0))
				for _, point := range polygon[1:] {
					x, y := float64(point[0]), float64(point[1])
					path.LineTo(float32(x), float32(y))
				}
				path.Close()
			}
			if hasPolygon {
				var options vector.DrawPathOptions
				options.ColorScale.ScaleWithColor(col)
				vector.FillPath(r.terrainAreaImage, &path, nil, &options)
			}
			return
		}
		for _, cell := range area.Cells {
			vector.FillRect(r.terrainAreaImage, float32(cell[0]), float32(cell[1]), 1, 1, col, true)
		}
	}
	// Geçilemeyen alanları önce, geçilebilir alanları sonra çiz. Böylece
	// geçilebilir alanın dolgusu ve üstteki border'ı komşu engelli alanın
	// altında kalmaz.
	if rebuild {
		// Geçilemeyen alanları önce, geçilebilir alanları sonra çiz. Böylece
		// geçilebilir alanın dolgusu ve üstteki border'ı komşu engelli alanın altında kalmaz.
		for _, area := range r.gs.TerrainAreas {
			if !world.TerrainAreaIsPassable(area) {
				drawArea(area)
			}
		}
		for _, area := range r.gs.TerrainAreas {
			if world.TerrainAreaIsPassable(area) {
				drawArea(area)
			}
		}
	}
	mapOp := &ebiten.DrawImageOptions{}
	r.applyMapGeoM(mapOp, float64(WorldW), float64(WorldH))
	screen.DrawImage(r.terrainAreaImage, mapOp)
	if selectedAreaID != "" {
		for _, area := range r.gs.TerrainAreas {
			if area.ID != selectedAreaID {
				continue
			}
			for _, polygon := range area.Polygons {
				if len(polygon) < 2 {
					continue
				}
				for i, point := range polygon {
					next := polygon[(i+1)%len(polygon)]
					x1, y1 := r.worldToScreen(float64(point[0]), float64(point[1]))
					x2, y2 := r.worldToScreen(float64(next[0]), float64(next[1]))
					vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 3, color.RGBA{255, 220, 70, 245}, true)
				}
			}
			break
		}
	}
}

func terrainAreaRenderKey(gs *state.GameState, selectedAreaID string, selectedRegionID world.RegionID) uint64 {
	if gs == nil {
		return 0
	}
	key := borderHashString(selectedAreaID)
	key ^= borderHashString(string(selectedRegionID))
	mix := func(value uint64) {
		key ^= value + 0x9e3779b97f4a7c15 + (key << 6) + (key >> 2)
	}
	mixString := func(value string) { mix(borderHashString(value)) }
	mixInt := func(value int) { mix(uint64(int64(value))) }
	for _, area := range gs.TerrainAreas {
		mixString(area.ID)
		mixString(string(area.Terrain))
		mixString(string(area.ParentRegionID))
		mixInt(area.MoveCost)
		mixInt(area.AttritionCost)
		for _, polygon := range area.Polygons {
			for _, point := range polygon {
				mixInt(point[0])
				mixInt(point[1])
			}
		}
		for _, cell := range area.Cells {
			mixInt(cell[0])
			mixInt(cell[1])
		}
	}
	return key
}

const (
	// Alfa değerleri düşük tutulur; terrain overlay haritanın altında kalan
	// dokuyu ve sınırları kapatmadan yalnızca arazi tipini belirtir.
	terrainAreaPassableAlpha         uint8 = 55
	terrainAreaBlockedAlpha          uint8 = 110
	terrainAreaSelectedPassableAlpha uint8 = 125
	terrainAreaSelectedBlockedAlpha  uint8 = 135
)

func terrainAreaColor(terrain world.TerrainType) color.RGBA {
	switch terrain {
	case world.TerrainPlain:
		return color.RGBA{224, 202, 112, 255}
	case world.TerrainForest:
		return color.RGBA{72, 112, 62, 255}
	case world.TerrainDenseForest:
		return color.RGBA{42, 78, 40, 255}
	case world.TerrainMountain:
		return color.RGBA{98, 65, 32, 255}
	case world.TerrainDesert:
		return color.RGBA{190, 154, 40, 255}
	case world.TerrainLake:
		return color.RGBA{120, 195, 232, 255}
	case world.TerrainRiver:
		return color.RGBA{80, 150, 205, 255}
	case world.TerrainSwamp:
		return color.RGBA{38, 98, 48, 255}
	case world.TerrainPass:
		return color.RGBA{155, 115, 70, 255}
	case world.TerrainCoast:
		return color.RGBA{170, 190, 120, 255}
	default:
		return color.RGBA{120, 120, 120, 255}
	}
}

func tintTerrainAreaColor(col color.RGBA, factor float64) color.RGBA {
	clamp := func(v float64) uint8 {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	return color.RGBA{clamp(float64(col.R) * factor), clamp(float64(col.G) * factor), clamp(float64(col.B) * factor), col.A}
}

func terrainAreaTypeColor(col, terrainColor color.RGBA) color.RGBA {
	return color.RGBA{R: terrainColor.R, G: terrainColor.G, B: terrainColor.B, A: col.A}
}
