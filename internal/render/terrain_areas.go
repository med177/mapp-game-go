package render

import (
	"image/color"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (r *Renderer) drawTerrainAreas(screen *ebiten.Image) {
	if r == nil || r.gs == nil {
		return
	}
	selectedAreaID := ""
	if selected := r.gs.Regions[r.editSelectedRegion]; selected != nil && selected.IsTerrainArea {
		selectedAreaID = selected.TerrainAreaID
	}
	for _, area := range r.gs.TerrainAreas {
		col := color.RGBA{120, 120, 120, 110}
		if parent := r.gs.Regions[area.ParentRegionID]; parent != nil {
			if owner := r.gs.Factions[faction.FactionID(parent.OwnerID)]; owner != nil {
				col = color.RGBA{owner.Color[0], owner.Color[1], owner.Color[2], 145}
			}
		}
		switch area.Terrain {
		case world.TerrainMountain:
			col = tintTerrainAreaColor(col, 0.78)
		case world.TerrainDesert:
			col = tintTerrainAreaColor(col, 1.18)
		case world.TerrainDenseForest:
			col = tintTerrainAreaColor(col, 0.62)
		case world.TerrainLake:
			col = tintTerrainAreaColor(col, 0.75)
		case world.TerrainRiver:
			col = tintTerrainAreaColor(col, 0.82)
		case world.TerrainSwamp:
			col = tintTerrainAreaColor(col, 0.7)
		}
		if area.MoveCost == 0 {
			col.A = 190
		}
		if area.ParentRegionID == r.editSelectedRegion || area.ID == selectedAreaID {
			col.A = 180
		}
		if len(area.Polygons) > 0 {
			var path vector.Path
			hasPolygon := false
			for _, polygon := range area.Polygons {
				if len(polygon) < 3 {
					continue
				}
				hasPolygon = true
				x0, y0 := r.worldToScreen(float64(polygon[0][0]), float64(polygon[0][1]))
				path.MoveTo(float32(x0), float32(y0))
				for _, point := range polygon[1:] {
					x, y := r.worldToScreen(float64(point[0]), float64(point[1]))
					path.LineTo(float32(x), float32(y))
				}
				path.Close()
			}
			if hasPolygon {
				var options vector.DrawPathOptions
				options.ColorScale.ScaleWithColor(col)
				vector.FillPath(screen, &path, nil, &options)
			}
			continue
		}
		for _, cell := range area.Cells {
			x0, y0 := r.worldToScreen(float64(cell[0]), float64(cell[1]))
			x1, y1 := r.worldToScreen(float64(cell[0]+1), float64(cell[1]+1))
			vector.FillRect(screen, float32(x0), float32(y0), float32(x1-x0), float32(y1-y0), col, true)
		}
	}
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
