package render

import (
	"fmt"
	"time"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// editMapBuildResult, harita rasterının worker goroutine'inde hazırlanmış
// sonucunu ana döngüye taşır. WorldMap'in img alanı worker'da oluşturulmaz;
// Ebiten nesneleri yalnızca ana thread'de finalize edilir.
type editMapBuildResult struct {
	generation uint64
	snapshot   *state.GameState
	worldMap   *WorldMap
	duration   time.Duration
	rasterTime time.Duration
	postTime   time.Duration
	err        error
}

type editMapBuildTiming struct {
	raster time.Duration
	post   time.Duration
}

func editMapBuildCancelled(done <-chan struct{}) bool {
	if done == nil {
		return false
	}
	select {
	case <-done:
		return true
	default:
		return false
	}
}

func cloneEditMapBuildState(gs *state.GameState) *state.GameState {
	if gs == nil {
		return nil
	}
	copyState := *gs
	copyState.Regions = cloneRegionMap(gs.Regions)
	copyState.TerrainAreas = cloneTerrainAreas(gs.TerrainAreas)
	copyState.LandPassages = cloneLandPassages(gs.LandPassages)
	copyState.Factions = cloneFactionMap(gs.Factions)
	copyState.Armies = cloneArmyMap(gs.Armies)
	copyState.Relations = cloneRelationMap(gs.Relations)
	copyState.ShapeData = cloneCountryShapeJSON(gs.ShapeData)
	copyState.TradeCenters = cloneTradeCenterConfig(gs.TradeCenters)
	copyState.Sieges = cloneSieges(gs.Sieges)
	copyState.RegionPaintOverrides = cloneRegionPaintOverrides(gs.RegionPaintOverrides)
	return &copyState
}

func cloneSieges(src map[world.RegionID]*state.SiegeState) map[world.RegionID]*state.SiegeState {
	if src == nil {
		return nil
	}
	dst := make(map[world.RegionID]*state.SiegeState, len(src))
	for regionID, siege := range src {
		if siege == nil {
			continue
		}
		copySiege := *siege
		dst[regionID] = &copySiege
	}
	return dst
}

func cloneWorldMapForEdit(src *WorldMap) *WorldMap {
	if src == nil {
		return nil
	}
	dst := *src
	dst.img = nil
	dst.basePixels = append([]byte(nil), src.basePixels...)
	dst.dispPixels = make([]byte, len(src.dispPixels))
	dst.baseRegionAt = append([]uint16(nil), src.baseRegionAt...)
	dst.regionAt = append([]uint16(nil), src.regionAt...)
	dst.regionIDs = append([]world.RegionID(nil), src.regionIDs...)
	dst.regionIdx = make(map[world.RegionID]uint16, len(src.regionIdx))
	for rid, idx := range src.regionIdx {
		dst.regionIdx[rid] = idx
	}
	dst.regionPx = nil
	dst.shapeRasterPixels = make(map[string][]int, len(src.shapeRasterPixels))
	for shapeID, pixels := range src.shapeRasterPixels {
		dst.shapeRasterPixels[shapeID] = append([]int(nil), pixels...)
	}
	dst.regionAnchor = make(map[world.RegionID][2]int)
	dst.settlementAnchor = make(map[settlementAnchorKey][2]int)
	dst.primarySettlement = make(map[world.RegionID][2]int)
	dst.borderSegments = nil
	dst.borderStyles = nil
	dst.blockadeFractions = nil
	dst.seaIdx = make(map[uint16]bool, len(src.seaIdx))
	for idx, isSea := range src.seaIdx {
		dst.seaIdx[idx] = isSea
	}
	return &dst
}

func buildEditMapSnapshot(done <-chan struct{}, gs *state.GameState, overrides map[int]world.RegionID, baseMap *WorldMap, targetShapeID string) (*WorldMap, editMapBuildTiming, error) {
	var timing editMapBuildTiming
	if gs == nil {
		return nil, timing, fmt.Errorf("harita snapshot state'i nil")
	}
	if editMapBuildCancelled(done) {
		return nil, timing, nil
	}

	rasterStarted := time.Now()
	var wm *WorldMap
	if targetShapeID != "" && baseMap != nil {
		// baseMap request aşamasında ayrılmış, worker'a özel snapshot'tır.
		// Burada ikinci kez büyük raster dizilerini kopyalamaya gerek yok.
		wm = baseMap
		if wm == nil || !wm.rebuildShapeRegionAssignments(gs, targetShapeID) {
			wm = nil
		}
	}
	if wm == nil {
		wm = prepareWorldMapData(gs, "", MapModeNormal, nil, false, true)
	}
	timing.raster = time.Since(rasterStarted)
	if wm == nil {
		return nil, timing, fmt.Errorf("harita verisi hazırlanamadı")
	}
	if editMapBuildCancelled(done) {
		return nil, timing, nil
	}

	postStarted := time.Now()
	// prepareWorldMapData kaynak region_shapes.json'ı da yükleyebilir. Edit
	// oturumunun override'ları ise kaynak dosyadan bağımsız olarak son katman
	// olmalıdır; bu yüzden immutable başlangıç rasterına dönüp onları uygula.
	if len(wm.baseRegionAt) == len(wm.regionAt) {
		copy(wm.regionAt, wm.baseRegionAt)
		wm.rebuildRegionPixelsFromAssignments()
		wm.applyRegionPaintOverridesToWorldMap(overrides)
		world.SyncTerrainAreaRegions(gs.Regions, gs.TerrainAreas)
		wm.applyTerrainAreaRegions(gs)
		wm.rebuildBorderSegments(gs)
		clear(wm.regionAnchor)
		wm.computeRegionAnchors()
		clear(wm.settlementAnchor)
		clear(wm.primarySettlement)
		wm.computeSettlementAnchors(gs)
		wm.applyOwnership(gs, "", MapModeNormal)
	}
	if editMapBuildCancelled(done) {
		timing.post = time.Since(postStarted)
		return nil, timing, nil
	}
	timing.post = time.Since(postStarted)
	return wm, timing, nil
}

func (r *Renderer) cancelEditMapBuild() {
	if r == nil {
		return
	}
	r.editMapBuildGeneration++
	r.editMapBuildPending = false
	r.editMapBuildResult = nil
	if r.editMapBuildCancel != nil {
		r.editMapBuildCancel()
		r.editMapBuildCancel = nil
	}
	r.editMapBuildCompletion = nil
}

// requestEditWorldMapRebuild, pahalı shape/merkez değişikliklerini input
// thread'ini bloke etmeden hesaplar. Worker sonucu yalnızca aynı generation
// hâlâ geçerliyse ana döngüde kabul edilir.
func (r *Renderer) requestEditWorldMapRebuild() {
	_ = r.requestEditWorldMapRebuildForShapeWithCompletion("", nil)
}

func (r *Renderer) requestEditWorldMapRebuildWithCompletion(completion func()) bool {
	return r.requestEditWorldMapRebuildForShapeWithCompletion("", completion)
}

func (r *Renderer) requestEditWorldMapRebuildForShape(shapeID string) {
	_ = r.requestEditWorldMapRebuildForShapeWithCompletion(shapeID, nil)
}

func (r *Renderer) requestEditWorldMapRebuildForShapeWithCompletion(shapeID string, completion func()) bool {
	if r == nil || r.gs == nil || r.editMapBuildPending {
		return false
	}

	r.editMapBuildGeneration++
	generation := r.editMapBuildGeneration
	snapshot := cloneEditMapBuildState(r.gs)
	overrides := cloneRegionPaintOverrides(r.editRegionPaintOverrides)
	baseMap := (*WorldMap)(nil)
	if shapeID != "" {
		baseMap = cloneWorldMapForEdit(r.worldMap)
	}
	resultCh := make(chan editMapBuildResult, 1)
	cancelCh := make(chan struct{})
	r.editMapBuildPending = true
	r.editMapBuildResult = resultCh
	r.editMapBuildCancel = func() { close(cancelCh) }
	r.editMapBuildCompletion = completion

	go func() {
		started := time.Now()
		var (
			worldMap *WorldMap
			timing   editMapBuildTiming
			err      error
		)
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					err = fmt.Errorf("harita worker panic: %v", recovered)
				}
			}()
			worldMap, timing, err = buildEditMapSnapshot(cancelCh, snapshot, overrides, baseMap, shapeID)
		}()
		resultCh <- editMapBuildResult{
			generation: generation,
			snapshot:   snapshot,
			worldMap:   worldMap,
			duration:   time.Since(started),
			rasterTime: timing.raster,
			postTime:   timing.post,
			err:        err,
		}
	}()
	return true
}

// pollEditMapBuild yalnız ana oyun döngüsünde çağrılır; Ebiten image'ı ve
// canlı state burada güncellenir.
func (r *Renderer) pollEditMapBuild() {
	if r == nil || !r.editMapBuildPending || r.editMapBuildResult == nil {
		return
	}
	select {
	case result := <-r.editMapBuildResult:
		r.editMapBuildPending = false
		r.editMapBuildResult = nil
		r.editMapBuildCancel = nil
		completion := r.editMapBuildCompletion
		r.editMapBuildCompletion = nil
		if result.generation != r.editMapBuildGeneration {
			return
		}
		if result.err != nil {
			r.ShowCombatResult("Harita hazırlanamadı: " + result.err.Error())
			return
		}
		if result.snapshot == nil || result.worldMap == nil {
			return
		}
		r.gs.Regions = result.snapshot.Regions
		r.gs.TerrainAreas = result.snapshot.TerrainAreas
		r.worldMap = FinalizePreparedWorldMap(result.worldMap)
		r.editMapLastBuildDuration = result.duration
		r.editMapLastRasterDuration = result.rasterTime
		r.editMapLastPostProcessDuration = result.postTime
		r.invalidateEditRegionCenterMarkers()
		r.buildRegionPaintBaseline()
		r.syncLandPassageRegionsFromMap()
		if completion != nil {
			completion()
		}
	default:
	}
}
