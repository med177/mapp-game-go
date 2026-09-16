package render

import (
	"image/color"

	"mapp-game-go/internal/faction"
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
		col := color.RGBA{120, 120, 120, 90}
		if parent := r.gs.Regions[area.ParentRegionID]; parent != nil {
			if owner := r.gs.Factions[faction.FactionID(parent.OwnerID)]; owner != nil {
				col = color.RGBA{owner.Color[0], owner.Color[1], owner.Color[2], 155}
			}
		}
		switch area.Terrain {
		case world.TerrainPlain:
			col = terrainAreaTypeColor(col, color.RGBA{224, 202, 112, 255})
		case world.TerrainMountain:
			col = terrainAreaTypeColor(col, color.RGBA{98, 65, 32, 255})
		case world.TerrainDesert:
			col = terrainAreaTypeColor(col, color.RGBA{190, 154, 40, 255})
		case world.TerrainDenseForest:
			col = tintTerrainAreaColor(col, 0.62)
		case world.TerrainLake:
			col = terrainAreaTypeColor(col, color.RGBA{120, 195, 232, 255})
		case world.TerrainRiver:
			col = tintTerrainAreaColor(col, 0.82)
		case world.TerrainSwamp:
			col = terrainAreaTypeColor(col, color.RGBA{38, 98, 48, 255})
		}
		passable := world.TerrainAreaIsPassable(area)
		if passable {
			// Geçilebilir alan daha saydam çizilir; border ve altındaki harita
			// görünür kalır.
			col.A = 85
		} else {
			// Geçilemeyen alan parlak bir terrain rengi üretmesin; koyu
			// grimsi-siyah bir örtü olarak kalsın.
			col = tintTerrainAreaColor(col, 0.32)
			col.A = 165
		}
		if area.ParentRegionID == r.editSelectedRegion || area.ID == selectedAreaID {
			if passable {
				col.A = 200
			} else {
				col.A = 175
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
		if parent := gs.Regions[area.ParentRegionID]; parent != nil {
			mixString(parent.OwnerID)
			if owner := gs.Factions[faction.FactionID(parent.OwnerID)]; owner != nil {
				for _, channel := range owner.Color {
					mix(uint64(channel))
				}
			}
		}
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
