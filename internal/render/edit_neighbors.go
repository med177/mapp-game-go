package render

import "mapp-game-go/internal/world"

func (r *Renderer) toggleEditNeighborAddMode() {
	r.editNeighborAddMode = !r.editNeighborAddMode
	r.editLandPassageMode = false
	r.editLandPassageAdjustMode = false
	r.editLandPassageFrom = ""
	r.editLandPassageStart = [2]int{}
	r.editLandPassageStartSet = false
	r.editLandPassageSelected = -1
	r.editLandPassageDragEndpoint = -1
	r.editLandPassageDragBefore = nil
	r.editLandPassageDragChanged = false
	if !r.editNeighborAddMode {
		r.editNeighborAddFrom = ""
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
	r.editNeighborAddMessage = "hedef kara bölgesine tıkla"
}

func (r *Renderer) handleEditNeighborAddClick(fx, fy float64) {
	if !r.editNeighborAddMode || r.editNeighborAddFrom == "" {
		return
	}
	targetID := r.editRegionAt(fx, fy)
	target := r.gs.Regions[targetID]
	if target == nil || target.IsSea {
		r.editNeighborAddMessage = "yalnızca kara bölgesi seç"
		return
	}
	if target.ID == r.editNeighborAddFrom {
		r.editNeighborAddMessage = "aynı bölge seçilemez"
		return
	}
	r.addNeighborBetween(r.editNeighborAddFrom, target.ID)
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
	before := r.neighborSnapshot(from, []world.RegionID{to})
	addNeighborID(source, to)
	addNeighborID(target, from)
	after := r.neighborSnapshot(from, []world.RegionID{to})
	if neighborSnapshotsEqual(before, after) {
		return
	}
	beforeCopy := cloneNeighborSnapshots(before)
	afterCopy := cloneNeighborSnapshots(after)
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.restoreNeighborSnapshots(beforeCopy)
			rr.editSelectedRegion = from
			rr.editSelectedSettlement = -1
			rr.editNeighborAddMessage = "komşuluk geri alındı"
		},
		redo: func(rr *Renderer) {
			rr.restoreNeighborSnapshots(afterCopy)
			rr.editSelectedRegion = from
			rr.editSelectedSettlement = -1
			rr.editNeighborAddMessage = "komşuluk eklendi"
		},
	})
	r.editNeighborAddMessage = "komşuluk eklendi"
}
