package render

import (
	"image/color"
	"math"

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
	rebuild := r.terrainAreaImage == nil || r.terrainAreaImage.Bounds().Dx() != WorldW || r.terrainAreaImage.Bounds().Dy() != WorldH || r.terrainAreaImageDirty
	if rebuild {
		r.terrainAreaImage = ebiten.NewImage(WorldW, WorldH)
		r.terrainAreaImage.Clear()
		r.terrainAreaImageDirty = false
	}

	areaColor := func(area world.TerrainArea) color.RGBA {
		col := terrainAreaColor(area.Terrain)
		passable := world.TerrainAreaIsPassable(area)
		if passable {
			// Geçilebilir alan daha saydam çizilir; border ve altındaki harita
			// görünür kalır.
			col.A = terrainAreaAlpha(r.gs.MapConfig.TerrainAlpha, true)
		} else {
			col.A = terrainAreaAlpha(r.gs.MapConfig.TerrainAlpha, false)
		}
		return col
	}

	drawArea := func(area world.TerrainArea) {
		col := areaColor(area)
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

	// Poligon dolguları yukarıdaki world-space raster image içinde hazırdır.
	// Kamerayı sürüklerken her alanı tekrar ekran koordinatına çevirip
	// FillPath yapmak yerine yalnızca tek image draw çağrısı kullanılır.
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

func terrainAreaAlpha(configured *float64, passable bool) uint8 {
	alpha := 1.0
	if configured != nil {
		alpha = *configured
	}
	if math.IsNaN(alpha) || math.IsInf(alpha, 0) {
		alpha = 1.0
	}
	if passable {
		alpha *= 0.8
	}
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	return uint8(math.Round(255 * alpha))
}

func terrainAreaColor(terrain world.TerrainType) color.RGBA {
	switch terrain {
	case world.TerrainPlain:
		return rgb(149, 135, 79)
	case world.TerrainForest:
		return rgb(29, 61, 21)
	case world.TerrainDenseForest:
		return rgb(15, 44, 13)
	case world.TerrainMountain:
		return rgb(49, 20, 4)
	case world.TerrainDesert:
		return rgb(129, 107, 36)
	case world.TerrainLake:
		return rgb(14, 40, 80)
	case world.TerrainRiver:
		return rgb(52, 100, 176)
	case world.TerrainSwamp:
		return rgb(53, 90, 5)
	case world.TerrainPass:
		return rgb(13, 74, 75)
	case world.TerrainCoast:
		return rgb(170, 190, 120)
	default:
		return rgb(74, 74, 74)
	}
}

func rgb(r, g, b uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func rgba(r, g, b uint8, a float64) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: uint8(math.Round(255 * a))}
}

func terrainAreaTypeColor(col, terrainColor color.RGBA) color.RGBA {
	return color.RGBA{R: terrainColor.R, G: terrainColor.G, B: terrainColor.B, A: col.A}
}
