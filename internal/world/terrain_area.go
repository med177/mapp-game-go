package world

import (
	"encoding/json"
	"fmt"
	"os"
)

const terrainAreaRegionPrefix = "terrain_area::"

// SyncTerrainAreaRegions materializes painted areas as lightweight child nodes.
// They remain runtime-only; the source of truth is TerrainAreas JSON data.
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
		parent := regions[area.ParentRegionID]
		if parent == nil || parent.IsSea || len(area.Cells) == 0 || area.ID == "" {
			continue
		}
		id := RegionID(terrainAreaRegionPrefix + area.ID)
		cx, cy := 0, 0
		for _, cell := range area.Cells {
			cx += cell[0]
			cy += cell[1]
		}
		cx /= len(area.Cells)
		cy /= len(area.Cells)
		terrain := area.Terrain
		if terrain == "" {
			terrain = parent.Terrain
		}
		props := TerrainData[terrain]
		regions[id] = &Region{ID: id, Name: area.Name, NameTR: area.Name, Terrain: terrain,
			OwnerID: parent.OwnerID, WorldX: cx, WorldY: cy, IsTerrainArea: true,
			ParentRegionID: parent.ID, TerrainAreaID: area.ID, IsLocked: !props.Passable}
		if regions[id].NameTR == "" {
			regions[id].NameTR = "Arazi Alanı"
		}
	}
	for i := range areas {
		area := &areas[i]
		id := RegionID(terrainAreaRegionPrefix + area.ID)
		child := regions[id]
		parent := regions[area.ParentRegionID]
		if child == nil || parent == nil {
			continue
		}
		child.Neighbors = append(child.Neighbors, parent.ID)
		for j := range areas {
			if i == j || areas[j].ParentRegionID != area.ParentRegionID {
				continue
			}
			if terrainAreasTouch(area, &areas[j]) {
				child.Neighbors = append(child.Neighbors, RegionID(terrainAreaRegionPrefix+areas[j].ID))
			}
		}
		for _, nid := range parent.Neighbors {
			if n := regions[nid]; n != nil && !n.IsTerrainArea {
				child.Neighbors = append(child.Neighbors, n.ID)
			}
		}
		for _, extraID := range area.ExtraNeighbors {
			extra := regions[extraID]
			if extra == nil || extra.ID == child.ID {
				continue
			}
			child.Neighbors = appendUniqueRegionID(child.Neighbors, extra.ID)
			extra.Neighbors = appendUniqueRegionID(extra.Neighbors, child.ID)
		}
		parent.Neighbors = appendUniqueRegionID(parent.Neighbors, child.ID)
	}
}

func UpdateTerrainAreaRegionOwners(regions map[RegionID]*Region) {
	for _, region := range regions {
		if region == nil || !region.IsTerrainArea {
			continue
		}
		if parent := regions[region.ParentRegionID]; parent != nil {
			region.OwnerID = parent.OwnerID
		}
	}
}

func terrainAreasTouch(a, b *TerrainArea) bool {
	for _, ca := range a.Cells {
		for _, cb := range b.Cells {
			dx, dy := ca[0]-cb[0], ca[1]-cb[1]
			if dx*dx+dy*dy == 1 {
				return true
			}
		}
	}
	return false
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
	ID             string      `json:"id"`
	ParentRegionID RegionID    `json:"parent_region_id"`
	Name           string      `json:"name,omitempty"`
	Terrain        TerrainType `json:"terrain"`
	MoveCost       int         `json:"move_cost"`
	Cells          [][2]int    `json:"cells"`
	// AttritionCost is the percentage of unit HP (0-100) an army loses when it
	// enters this area (e.g. desert heat, mountain fatigue). 0 means no wear.
	AttritionCost int `json:"attrition_cost,omitempty"`
	// ExtraNeighbors are manually assigned links beyond the automatic
	// parent/touching-sibling adjacency (e.g. a river area linking a bridge).
	ExtraNeighbors []RegionID `json:"extra_neighbors,omitempty"`
}

func (a TerrainArea) Contains(x, y int) bool {
	for _, cell := range a.Cells {
		if cell[0] == x && cell[1] == y {
			return true
		}
	}
	return false
}

// TerrainAreaMovementCost returns the additional movement points at a point.
// The most restrictive overlapping area wins; a zero-cost area always blocks.
func TerrainAreaMovementCost(areas []TerrainArea, parent RegionID, x, y int) (cost int, blocked bool) {
	for _, area := range areas {
		if area.ParentRegionID != parent || !area.Contains(x, y) {
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
	var areas []TerrainArea
	if err := json.Unmarshal(data, &areas); err != nil {
		return nil, fmt.Errorf("terrain_areas JSON parse hatası: %w", err)
	}
	valid := areas[:0]
	seen := make(map[string]bool, len(areas))
	for _, area := range areas {
		parent := regions[area.ParentRegionID]
		if area.ID == "" || seen[area.ID] || parent == nil || parent.IsSea || len(area.Cells) == 0 {
			continue
		}
		seen[area.ID] = true
		valid = append(valid, area)
	}
	return valid, nil
}
