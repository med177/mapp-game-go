package render

import "mapp-game-go/internal/world"

func (r *Renderer) toggleEditNeighborAddMode() {
	if r.editNeighborAddMode {
		r.applyEditNeighborAddMode()
		return
	}
	r.editNeighborAddMode = !r.editNeighborAddMode
	r.editLandPassageMode = false
	r.editLandPassageAdjustMode = false
	r.editLandPassageFrom = ""
	r.editLandPassageStart = [2]int{}
	r.editLandPassageStartSet = false
	r.editLandPassageSelected = -1
	r.editLandPassageDragEndpoint = -1
	r.editLandPassageDragChanged = false
	if !r.editNeighborAddMode {
		r.editNeighborAddFrom = ""
		r.editNeighborAddTargets = nil
		r.editNeighborAddMessage = ""
		return
	}
	source := r.gs.Regions[r.editSelectedRegion]
	if source == nil || source.IsSea {
		r.editNeighborAddMode = false
		r.editNeighborAddFrom = ""
		r.editNeighborAddMessage = "önce bir kara bölgesi seç"
		return
	}
	r.editNeighborAddFrom = source.ID
	r.editNeighborAddTargets = nil
	r.editNeighborAddMessage = "hedef kara bölgesine tıkla"
}

func (r *Renderer) handleEditNeighborAddClick(fx, fy float64) {
	if !r.editNeighborAddMode || r.editNeighborAddFrom == "" {
		return
	}
	targetID, terrainHit := r.terrainAreaRegionAt(fx, fy)
	if !terrainHit {
		targetID = r.editRegionAt(fx, fy)
	}
	target := r.gs.Regions[targetID]
	if target == nil || target.IsSea {
		r.editNeighborAddMessage = "yalnızca kara bölgesi seç"
		return
	}
	if target.ID == r.editNeighborAddFrom {
		r.editNeighborAddMessage = "aynı bölge seçilemez"
		return
	}
	for _, existing := range r.editNeighborAddTargets {
		if existing == target.ID {
			r.editNeighborAddMessage = "hedef zaten seçildi"
			return
		}
	}
	r.editNeighborAddTargets = append(r.editNeighborAddTargets, target.ID)
	r.editNeighborAddMessage = "hedef eklendi; Uygula ile kaydet"
}

func (r *Renderer) applyEditNeighborAddMode() {
	if r == nil || !r.editNeighborAddMode {
		return
	}
	source := r.gs.Regions[r.editNeighborAddFrom]
	if source == nil || len(r.editNeighborAddTargets) == 0 {
		r.editNeighborAddMode = false
		r.editNeighborAddFrom = ""
		r.editNeighborAddTargets = nil
		r.editNeighborAddMessage = "komşuluk seçimi iptal edildi"
		return
	}
	changed := false
	for _, targetID := range r.editNeighborAddTargets {
		target := r.gs.Regions[targetID]
		if target == nil || target.IsSea || target.ID == source.ID {
			continue
		}
		if source.IsTerrainArea {
			changed = r.appendTerrainAreaExtraNeighbor(source.TerrainAreaID, target.ID) || changed
		} else {
			beforeLen := len(source.Neighbors)
			addNeighborID(source, target.ID)
			changed = changed || len(source.Neighbors) != beforeLen
		}
		if target.IsTerrainArea {
			changed = r.appendTerrainAreaExtraNeighbor(target.TerrainAreaID, source.ID) || changed
		} else {
			beforeLen := len(target.Neighbors)
			addNeighborID(target, source.ID)
			changed = changed || len(target.Neighbors) != beforeLen
		}
	}
	if changed {
		r.requestEditWorldMapRebuild()
		r.editDirty = true
		r.editNeighborAddMessage = "komşuluklar uygulandı"
	} else {
		r.editNeighborAddMessage = "yeni komşuluk yok"
	}
	r.editNeighborAddMode = false
	r.editNeighborAddFrom = ""
	r.editNeighborAddTargets = nil
}

func (r *Renderer) addNeighborBetween(from, to world.RegionID) {
	source := r.gs.Regions[from]
	target := r.gs.Regions[to]
	if source == nil || target == nil || source.IsSea || target.IsSea || from == to {
		r.editNeighborAddMessage = "yalnızca iki farklı kara bölgesi seç"
		return
	}
	if regionHasNeighbor(source, to) && regionHasNeighbor(target, from) {
		r.editNeighborAddMessage = "komşuluk zaten var"
		return
	}
	if source.IsTerrainArea || target.IsTerrainArea {
		r.addTerrainAreaNeighbor(source, target)
		return
	}
	addNeighborID(source, to)
	addNeighborID(target, from)
	r.editDirty = true
	r.editNeighborAddMessage = "komşuluk eklendi"
}

// addTerrainAreaNeighbor bir arazi alanının katıldığı komşuluğu kalıcı kılar.
// Arazi alanı komşu listesi her senkronizasyonda yeniden üretildiği için
// manuel eklemeler TerrainArea.ExtraNeighbors üzerinden saklanır.
func (r *Renderer) addTerrainAreaNeighbor(source, target *world.Region) {
	changed := false
	if source.IsTerrainArea {
		if r.appendTerrainAreaExtraNeighbor(source.TerrainAreaID, target.ID) {
			changed = true
		}
	} else {
		addNeighborID(source, target.ID)
		changed = true
	}
	if target.IsTerrainArea {
		if r.appendTerrainAreaExtraNeighbor(target.TerrainAreaID, source.ID) {
			changed = true
		}
	} else {
		addNeighborID(target, source.ID)
		changed = true
	}
	if !changed {
		return
	}
	r.requestEditWorldMapRebuild()
	r.editSelectedRegion = source.ID
	r.editSelectedSettlement = -1
	r.editNeighborAddMessage = "komşuluk eklendi"
	r.editDirty = true
}

// appendTerrainAreaExtraNeighbor verilen arazi alanına manuel komşu ekler;
// zaten eklenmişse false döner.
func (r *Renderer) appendTerrainAreaExtraNeighbor(areaID string, neighborID world.RegionID) bool {
	for i := range r.gs.TerrainAreas {
		if r.gs.TerrainAreas[i].ID != areaID {
			continue
		}
		for _, existing := range r.gs.TerrainAreas[i].ExtraNeighbors {
			if existing == neighborID {
				return false
			}
		}
		r.gs.TerrainAreas[i].ExtraNeighbors = append(r.gs.TerrainAreas[i].ExtraNeighbors, neighborID)
		return true
	}
	return false
}
