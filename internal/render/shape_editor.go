package render

import (
	"image/color"
	"sort"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type editShapeBrushMode int

const (
	editShapeBrushPaint editShapeBrushMode = iota
	editShapeBrushErase
)

type shapeEditSession struct {
	ShapeID  string
	Name     string
	MinX     int
	MinY     int
	MaxX     int
	MaxY     int
	Width    int
	Height   int
	Mask     []byte
	BaseMask []byte
	DiffMask []byte
	DiffList []int
	Dirty    bool
	LastX    int
	LastY    int
	HasLast  bool
}

type gridPoint struct{ X, Y int }

func (s *shapeEditSession) inBounds(x, y int) bool {
	return s != nil && x >= s.MinX && x <= s.MaxX && y >= s.MinY && y <= s.MaxY
}

func (s *shapeEditSession) index(x, y int) int {
	return (y-s.MinY)*s.Width + (x - s.MinX)
}

func (s *shapeEditSession) filled(x, y int) bool {
	if !s.inBounds(x, y) {
		return false
	}
	return s.Mask[s.index(x, y)] != 0
}

func (r *Renderer) invalidateShapeEditSession() {
	r.editShapeSession = nil
	r.editShapePainting = false
	r.editShapeStrokeBefore = nil
}

func (r *Renderer) selectedShapeRegion() *world.Region {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil || region.IsSea || region.ShapeID == "" {
		return nil
	}
	return region
}

func (r *Renderer) canEditSelectedShape() bool {
	return r.selectedShapeRegion() != nil
}

func (r *Renderer) canRegionPaintSelected() bool {
	region := r.gs.Regions[r.editSelectedRegion]
	return region != nil && !region.IsSea
}

func (r *Renderer) ensureShapeEditSession() *shapeEditSession {
	region := r.selectedShapeRegion()
	if region == nil {
		return nil
	}
	if r.editShapeSession != nil && r.editShapeSession.ShapeID == region.ShapeID {
		return r.editShapeSession
	}
	session := newShapeEditSession(r.gs, region.ShapeID)
	r.editShapeSession = session
	return session
}

func newShapeEditSession(gs *state.GameState, shapeID string) *shapeEditSession {
	minX, minY, maxX, maxY := editableShapeCoordBounds()
	if maxX < minX || maxY < minY {
		return nil
	}
	width := maxX - minX + 1
	height := maxY - minY + 1
	session := &shapeEditSession{
		ShapeID:  shapeID,
		Name:     gs.ShapeData.Names[shapeID],
		MinX:     minX,
		MinY:     minY,
		MaxX:     maxX,
		MaxY:     maxY,
		Width:    width,
		Height:   height,
		Mask:     make([]byte, width*height),
		DiffMask: make([]byte, width*height),
		DiffList: make([]int, 0, 2048),
	}
	if session.Name == "" {
		session.Name = shapeID
	}
	for _, ring := range gs.ShapeData.Shapes[shapeID] {
		rasterizeFloatRingToMask(session, ring)
	}
	session.BaseMask = make([]byte, len(session.Mask))
	copy(session.BaseMask, session.Mask)
	return session
}

func (s *shapeEditSession) resetStrokeDiff() {
	if s == nil {
		return
	}
	for _, idx := range s.DiffList {
		s.DiffMask[idx] = 0
	}
	s.DiffList = s.DiffList[:0]
}

func (s *shapeEditSession) trackDiff(idx int) {
	if s == nil || idx < 0 || idx >= len(s.Mask) || idx >= len(s.BaseMask) {
		return
	}
	if s.Mask[idx] != s.BaseMask[idx] {
		if s.DiffMask[idx] == 0 {
			s.DiffMask[idx] = 1
			s.DiffList = append(s.DiffList, idx)
		}
		return
	}
	s.DiffMask[idx] = 0
}

func editableShapeCoordBounds() (int, int, int, int) {
	maxX := int((float64(WorldW)-shapeOffX)/shapeScaleX + 1.5)
	maxY := int((float64(WorldH)-shapeOffY)/shapeScaleY + 1.5)
	if maxX < 1 {
		maxX = 1
	}
	if maxY < 1 {
		maxY = 1
	}
	return 0, 0, maxX, maxY
}

func rasterizeFloatRingToMask(session *shapeEditSession, ring [][2]float32) {
	if session == nil || len(ring) < 3 {
		return
	}
	minX, minY := int(ring[0][0]), int(ring[0][1])
	maxX, maxY := minX, minY
	for _, pt := range ring[1:] {
		x, y := int(pt[0]), int(pt[1])
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	if minX < session.MinX {
		minX = session.MinX
	}
	if minY < session.MinY {
		minY = session.MinY
	}
	if maxX > session.MaxX {
		maxX = session.MaxX
	}
	if maxY > session.MaxY {
		maxY = session.MaxY
	}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if pointInFloatPolygon(float64(x)+0.5, float64(y)+0.5, ring) {
				session.Mask[session.index(x, y)] = 1
			}
		}
	}
}

func pointInFloatPolygon(x, y float64, poly [][2]float32) bool {
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

func (r *Renderer) drawEditShapeInspector(screen *ebiten.Image, ly float64) {
	x, _, _, _ := editInspectorRect()
	region := r.selectedShapeRegion()
	if region == nil {
		DrawText(screen, "Shape editor icin kara bolgesi sec.", float64(x)+14, ly, FaceSmall, ColorGray)
		drawEditInspectorButton(screen, editButtonShapePaint, "Boya", false)
		drawEditInspectorButton(screen, editButtonShapeErase, "Sil", false)
		drawEditInspectorButton(screen, editButtonShapeRegionPaint, "Bolge Boya", false)
		drawEditInspectorButton(screen, editButtonShapeRegionErase, "Bolge Sil", false)
		drawEditInspectorButton(screen, editButtonShapeBrushMinus, "Firca -", false)
		drawEditInspectorButton(screen, editButtonShapeBrushPlus, "Firca +", false)
		drawEditInspectorButton(screen, editButtonSaveScenario, "Kaydet", true)
		return
	}
	session := r.ensureShapeEditSession()
	shapeID := region.ShapeID
	name := shapeID
	if session != nil && session.Name != "" {
		name = session.Name
	}
	DrawText(screen, "Shape ID: "+shapeID, float64(x)+14, ly, FaceSmall, ColorWhite)
	ly += 18
	DrawText(screen, "Ad: "+name, float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 18
	DrawText(screen, "Ring: "+itoa(len(r.gs.ShapeData.Shapes[shapeID]))+"   Firca: "+itoa(r.editShapeBrushRadius), float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 18
	toolLabel := "Shape"
	if r.editShapeTool == editShapeToolRegion {
		toolLabel = "Bolge"
	}
	modeLabel := "Boya"
	if r.editShapeBrushMode == editShapeBrushErase {
		modeLabel = "Sil"
	}
	DrawText(screen, "Arac: "+toolLabel+"  Mod: "+modeLabel+"   Girdi: sag mouse drag", float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 18
	DrawText(screen, "Canli preview acik. Yesil ekler, kirmizi siler.", float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 18
	strokeLabel := "Bekliyor"
	if r.editShapePainting {
		strokeLabel = "Boyaniyor"
	}
	DrawText(screen, "Durum: "+strokeLabel+"   Mouse birakinca uygula+undo", float64(x)+14, ly, FaceSmall, ColorGray)

	shapePaintLabel := "Shape Boya"
	shapeEraseLabel := "Shape Sil"
	regionPaintLabel := "Bolge Boya"
	regionEraseLabel := "Bolge Sil"
	if r.editShapeTool == editShapeToolShape && r.editShapeBrushMode == editShapeBrushPaint {
		shapePaintLabel = "> Shape Boya"
	} else if r.editShapeTool == editShapeToolShape && r.editShapeBrushMode == editShapeBrushErase {
		shapeEraseLabel = "> Shape Sil"
	}
	if r.editShapeTool == editShapeToolRegion && r.editShapeBrushMode == editShapeBrushPaint {
		regionPaintLabel = "> Bolge Boya"
	} else if r.editShapeTool == editShapeToolRegion && r.editShapeBrushMode == editShapeBrushErase {
		regionEraseLabel = "> Bolge Sil"
	}
	drawEditInspectorButton(screen, editButtonShapePaint, shapePaintLabel, true)
	drawEditInspectorButton(screen, editButtonShapeErase, shapeEraseLabel, true)
	drawEditInspectorButton(screen, editButtonShapeRegionPaint, regionPaintLabel, true)
	drawEditInspectorButton(screen, editButtonShapeRegionErase, regionEraseLabel, true)
	drawEditInspectorButton(screen, editButtonShapeBrushMinus, "Firca -", r.editShapeBrushRadius > 1)
	drawEditInspectorButton(screen, editButtonShapeBrushPlus, "Firca +", r.editShapeBrushRadius < 64)
	drawEditInspectorButton(screen, editButtonSaveScenario, "Kaydet", true)
}

func (r *Renderer) handleEditShapeInspectorClick(fx, fy float64) (InputAction, bool) {
	switch editShapeInspectorButtonAt(fx, fy) {
	case editButtonShapePaint:
		r.editShapeTool = editShapeToolShape
		r.editShapeBrushMode = editShapeBrushPaint
	case editButtonShapeErase:
		r.editShapeTool = editShapeToolShape
		r.editShapeBrushMode = editShapeBrushErase
	case editButtonShapeRegionPaint:
		r.editShapeTool = editShapeToolRegion
		r.editShapeBrushMode = editShapeBrushPaint
	case editButtonShapeRegionErase:
		r.editShapeTool = editShapeToolRegion
		r.editShapeBrushMode = editShapeBrushErase
	case editButtonShapeBrushMinus:
		if r.editShapeBrushRadius > 1 {
			r.editShapeBrushRadius--
		}
	case editButtonShapeBrushPlus:
		if r.editShapeBrushRadius < 64 {
			r.editShapeBrushRadius++
		}
	case editButtonSaveScenario:
		return InputAction{Kind: ActionSaveScenario}, true
	}
	return InputAction{}, true
}

func (r *Renderer) drawEditShapeOverlay(screen *ebiten.Image) {
	if r.gs.Phase != state.PhaseEditMode || r.editInspectorTab != editInspectorShape {
		return
	}
	region := r.selectedShapeRegion()
	if region == nil {
		return
	}
	for _, ring := range r.gs.ShapeData.Shapes[region.ShapeID] {
		if len(ring) < 2 {
			continue
		}
		for i := range ring {
			a := ring[i]
			b := ring[(i+1)%len(ring)]
			ax, ay := r.worldToScreen(wcX(int(a[0]+0.5)), wcY(int(a[1]+0.5)))
			bx, by := r.worldToScreen(wcX(int(b[0]+0.5)), wcY(int(b[1]+0.5)))
			vector.StrokeLine(screen, float32(ax), float32(ay), float32(bx), float32(by), 2, color.RGBA{60, 235, 255, 215}, true)
		}
	}
	if session := r.ensureShapeEditSession(); session != nil {
		r.drawEditShapeStrokePreview(screen, session)
		r.drawEditShapeHelp(screen, session)
	}
	if !r.canEditSelectedShape() {
		return
	}
	mx, my := ebiten.CursorPosition()
	if editInspectorHit(float64(mx), float64(my)) {
		return
	}
	wx, wy := r.screenToWorld(float64(mx), float64(my))
	sx, sy := r.worldToScreen(wx, wy)
	radius := float32(maxF(6, float64(r.editShapeBrushRadius)*maxF(shapeScaleX, shapeScaleY)*r.camScale))
	brushCol := color.RGBA{80, 235, 255, 180}
	if r.editShapeBrushMode == editShapeBrushErase {
		brushCol = color.RGBA{255, 110, 110, 185}
	}
	vector.StrokeCircle(screen, float32(sx), float32(sy), radius, 2, brushCol, true)
}

func (r *Renderer) drawEditShapeStrokePreview(screen *ebiten.Image, session *shapeEditSession) {
	if session == nil || len(session.DiffList) == 0 {
		return
	}
	size := float32(maxF(2, maxF(shapeScaleX, shapeScaleY)*r.camScale))
	for _, idx := range session.DiffList {
		if idx < 0 || idx >= len(session.DiffMask) || session.DiffMask[idx] == 0 {
			continue
		}
		x := idx%session.Width + session.MinX
		y := idx/session.Width + session.MinY
		sx, sy := r.worldToScreen(wcX(x), wcY(y))
		if sx < -8 || sx > ScreenWidth+8 || sy < -8 || sy > ScreenHeight+8 {
			continue
		}
		col := color.RGBA{80, 235, 120, 165}
		if session.Mask[idx] == 0 {
			col = color.RGBA{255, 90, 90, 170}
		}
		drawPixelRect(screen, float32(sx)-size/2, float32(sy)-size/2, size, col)
	}
}

func (r *Renderer) drawEditShapeHelp(screen *ebiten.Image, session *shapeEditSession) {
	if session == nil {
		return
	}
	const panelW, panelH = float32(290), float32(92)
	x := float32(ScreenWidth) - panelW - 18
	y := float32(18)
	drawRoundedRect(screen, x, y, panelW, panelH, 8, color.RGBA{16, 20, 24, 218})
	drawPanelBorder(screen, x, y, panelW, panelH)
	DrawText(screen, "SHAPE YARDIM", float64(x)+12, float64(y)+10, FaceSmall, ColorGold)
	DrawText(screen, "Secili: "+session.ShapeID+"  Firca: "+itoa(r.editShapeBrushRadius), float64(x)+12, float64(y)+30, FaceSmall, ColorWhite)
	mode := "Boya"
	if r.editShapeBrushMode == editShapeBrushErase {
		mode = "Sil"
	}
	DrawText(screen, "Mod: "+mode+"  Sag mouse drag  Birakinca uygula", float64(x)+12, float64(y)+48, FaceSmall, ColorGray)
	DrawText(screen, "Yesil=ekle  Kirmizi=sil  Sol tik=secim", float64(x)+12, float64(y)+66, FaceSmall, ColorGray)
}

func (r *Renderer) beginShapePaintStroke(fx, fy float64) bool {
	if editInspectorHit(fx, fy) {
		return false
	}
	switch r.editShapeTool {
	case editShapeToolShape:
		if !r.canEditSelectedShape() {
			return false
		}
	case editShapeToolRegion:
		if !r.canRegionPaintSelected() {
			return false
		}
	}
	session := r.ensureShapeEditSession()
	if session == nil && r.editShapeTool == editShapeToolShape {
		return false
	}
	wx, wy := r.screenToWorld(fx, fy)
	var sx, sy int
	if r.editShapeTool == editShapeToolRegion {
		sx, sy = int(wx+0.5), int(wy+0.5)
	} else {
		sx, sy = scenarioCoordsFromWorld(wx, wy)
		if !session.inBounds(sx, sy) {
			return false
		}
	}
	before := r.worldSnapshot()
	r.editShapeStrokeBefore = &before
	r.editShapePainting = true
	r.editShapeStrokeHasLast = false
	r.editShapeStrokeDirty = false
	if session != nil {
		session.Dirty = false
		session.HasLast = false
		session.resetStrokeDiff()
	}
	r.applyShapeBrushAt(session, sx, sy)
	return true
}

func (r *Renderer) continueShapePaintStroke(fx, fy float64) {
	if !r.editShapePainting {
		return
	}
	if r.editShapeTool == editShapeToolShape {
		session := r.ensureShapeEditSession()
		if session == nil {
			return
		}
		wx, wy := r.screenToWorld(fx, fy)
		sx, sy := scenarioCoordsFromWorld(wx, wy)
		r.applyShapeBrushAt(session, sx, sy)
		return
	}
	if r.editShapeTool == editShapeToolRegion {
		if !r.canRegionPaintSelected() {
			return
		}
		wx, wy := r.screenToWorld(fx, fy)
		sx, sy := int(wx+0.5), int(wy+0.5)
		r.applyShapeBrushAt(nil, sx, sy)
	}
}

func (r *Renderer) finishShapePaintStroke() {
	before := r.editShapeStrokeBefore
	session := r.editShapeSession
	r.editShapePainting = false
	r.editShapeStrokeBefore = nil
	if before == nil {
		return
	}
	if r.editShapeTool == editShapeToolShape {
		if session == nil || !session.Dirty {
			return
		}
		rings := shapeMaskToFloatRings(session)
		applyShapeRingsToState(r.gs, session.ShapeID, rings)
		r.rebuildEditWorldMap()
		after := r.worldSnapshot()
		r.pushWorldSnapshotCommand(*before, after)
		r.editDirty = true
		if len(rings) == 0 {
			r.ShowCombatResult("Shape tamamen silindi.")
		}
		return
	}
	if r.editShapeTool == editShapeToolRegion {
		if !r.editShapeStrokeDirty {
			return
		}
		// Region paint overrides'ı oyun durumuna kaydet
		if len(r.editRegionPaintOverrides) > 0 {
			if r.gs.RegionPaintOverrides == nil {
				r.gs.RegionPaintOverrides = make(map[int]world.RegionID)
			}
			for pIdx, rid := range r.editRegionPaintOverrides {
				r.gs.RegionPaintOverrides[pIdx] = rid
			}
		}
		r.rebuildEditWorldMap()
		after := r.worldSnapshot()
		r.pushWorldSnapshotCommand(*before, after)
		r.editDirty = true
	}
}

func (r *Renderer) applyShapeBrushAt(session *shapeEditSession, x, y int) {
	if r.editShapeTool == editShapeToolRegion {
		if !r.canRegionPaintSelected() {
			return
		}
		if !r.editShapeStrokeHasLast {
			if r.applyRegionBrushCircle(x, y, r.editShapeBrushRadius, r.editShapeBrushMode == editShapeBrushPaint) {
				if session != nil {
					session.Dirty = true
				}
				r.editShapeStrokeDirty = true
			}
			r.editShapeStrokeLastX, r.editShapeStrokeLastY, r.editShapeStrokeHasLast = x, y, true
			return
		}
		if r.applyRegionBrushLine(r.editShapeStrokeLastX, r.editShapeStrokeLastY, x, y, r.editShapeBrushRadius, r.editShapeBrushMode == editShapeBrushPaint) {
			if session != nil {
				session.Dirty = true
			}
			r.editShapeStrokeDirty = true
		}
		r.editShapeStrokeLastX, r.editShapeStrokeLastY = x, y
		return
	}
	if session == nil {
		return
	}
	if !session.HasLast {
		if applyShapeBrushCircle(session, x, y, r.editShapeBrushRadius, r.editShapeBrushMode == editShapeBrushPaint) {
			session.Dirty = true
		}
		session.LastX, session.LastY, session.HasLast = x, y, true
		return
	}
	if applyShapeBrushLine(session, session.LastX, session.LastY, x, y, r.editShapeBrushRadius, r.editShapeBrushMode == editShapeBrushPaint) {
		session.Dirty = true
	}
	session.LastX, session.LastY = x, y
}

func applyShapeBrushLine(session *shapeEditSession, x0, y0, x1, y1, radius int, fill bool) bool {
	steps := maxInt(absInt(x1-x0), absInt(y1-y0))
	if steps == 0 {
		return applyShapeBrushCircle(session, x0, y0, radius, fill)
	}
	changed := false
	for i := 0; i <= steps; i++ {
		x := x0 + (x1-x0)*i/steps
		y := y0 + (y1-y0)*i/steps
		if applyShapeBrushCircle(session, x, y, radius, fill) {
			changed = true
		}
	}
	return changed
}

func applyShapeBrushCircle(session *shapeEditSession, cx, cy, radius int, fill bool) bool {
	if session == nil || radius < 0 {
		return false
	}
	changed := false
	r2 := radius * radius
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if !session.inBounds(x, y) {
				continue
			}
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy > r2 {
				continue
			}
			idx := session.index(x, y)
			want := byte(0)
			if fill {
				want = 1
			}
			if session.Mask[idx] != want {
				session.Mask[idx] = want
				session.trackDiff(idx)
				changed = true
			}
		}
	}
	return changed
}

func (r *Renderer) applyRegionBrushLine(x0, y0, x1, y1, radius int, fill bool) bool {
	steps := maxInt(absInt(x1-x0), absInt(y1-y0))
	if steps == 0 {
		return r.applyRegionBrushCircle(x0, y0, radius, fill)
	}
	changed := false
	for i := 0; i <= steps; i++ {
		x := x0 + (x1-x0)*i/steps
		y := y0 + (y1-y0)*i/steps
		if r.applyRegionBrushCircle(x, y, radius, fill) {
			changed = true
		}
	}
	return changed
}

func (r *Renderer) applyRegionBrushCircle(cx, cy, radius int, fill bool) bool {
	if r.worldMap == nil || radius < 0 || !r.canRegionPaintSelected() {
		return false
	}
	regionID := r.editSelectedRegion
	changed := false
	r2 := radius * radius
	for y := cy - radius; y <= cy+radius; y++ {
		if y < 0 || y >= WorldH {
			continue
		}
		for x := cx - radius; x <= cx+radius; x++ {
			if x < 0 || x >= WorldW {
				continue
			}
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy > r2 {
				continue
			}
			pIdx := y*WorldW + x
			baselineIdx := uint16(0)
			if len(r.editRegionPaintBaseline) == len(r.worldMap.regionAt) {
				baselineIdx = r.editRegionPaintBaseline[pIdx]
			}
			if fill {
				if baselineIdx == r.worldMap.regionIdx[regionID] {
					delete(r.editRegionPaintOverrides, pIdx)
					continue
				}
				r.editRegionPaintOverrides[pIdx] = regionID
				r.applyRegionOverride(pIdx, regionID)
				changed = true
				continue
			}
			if _, ok := r.editRegionPaintOverrides[pIdx]; !ok {
				continue
			}
			delete(r.editRegionPaintOverrides, pIdx)
			if baselineIdx != 0 {
				oldID := r.worldMap.regionIDs[r.worldMap.regionAt[pIdx]]
				r.worldMap.regionPx[oldID] = removePixelIndex(r.worldMap.regionPx[oldID], pIdx)
				r.worldMap.regionAt[pIdx] = baselineIdx
				baselineID := r.worldMap.regionIDs[baselineIdx]
				r.worldMap.regionPx[baselineID] = append(r.worldMap.regionPx[baselineID], pIdx)
			} else {
				oldID := r.worldMap.regionIDs[r.worldMap.regionAt[pIdx]]
				r.worldMap.regionPx[oldID] = removePixelIndex(r.worldMap.regionPx[oldID], pIdx)
				r.worldMap.regionAt[pIdx] = 0
			}
			changed = true
		}
	}
	return changed
}

func applyShapeRingsToState(gs *state.GameState, shapeID string, rings [][][2]float32) {
	if gs.ShapeData.Shapes == nil {
		gs.ShapeData.Shapes = make(map[string][][][2]float32)
	}
	if gs.ShapeData.Names == nil {
		gs.ShapeData.Names = make(map[string]string)
	}
	gs.ShapeData.Shapes[shapeID] = cloneFloatRings(rings)
	if gs.ShapeData.Names[shapeID] == "" {
		gs.ShapeData.Names[shapeID] = shapeID
	}
	recalculateCountryShapeBounds(&gs.ShapeData)
	for _, region := range gs.Regions {
		if region == nil || region.ShapeID != shapeID {
			continue
		}
		region.Shape = cloneFloatRings(rings)
	}
}

func recalculateCountryShapeBounds(shapeData *world.CountryShapeJSON) {
	first := true
	for _, rings := range shapeData.Shapes {
		for _, ring := range rings {
			for _, pt := range ring {
				if first {
					shapeData.Bounds = world.ShapeBounds{MinX: pt[0], MinY: pt[1], MaxX: pt[0], MaxY: pt[1]}
					first = false
					continue
				}
				if pt[0] < shapeData.Bounds.MinX {
					shapeData.Bounds.MinX = pt[0]
				}
				if pt[0] > shapeData.Bounds.MaxX {
					shapeData.Bounds.MaxX = pt[0]
				}
				if pt[1] < shapeData.Bounds.MinY {
					shapeData.Bounds.MinY = pt[1]
				}
				if pt[1] > shapeData.Bounds.MaxY {
					shapeData.Bounds.MaxY = pt[1]
				}
			}
		}
	}
	if first {
		shapeData.Bounds = world.ShapeBounds{}
	}
}

func cloneCountryShapeJSON(src world.CountryShapeJSON) world.CountryShapeJSON {
	dst := world.CountryShapeJSON{Bounds: src.Bounds}
	if src.Shapes != nil {
		dst.Shapes = make(map[string][][][2]float32, len(src.Shapes))
		for id, rings := range src.Shapes {
			dst.Shapes[id] = cloneFloatRings(rings)
		}
	}
	if src.Names != nil {
		dst.Names = make(map[string]string, len(src.Names))
		for id, name := range src.Names {
			dst.Names[id] = name
		}
	}
	return dst
}

func cloneFloatRings(src [][][2]float32) [][][2]float32 {
	if src == nil {
		return nil
	}
	dst := make([][][2]float32, len(src))
	for i := range src {
		dst[i] = make([][2]float32, len(src[i]))
		copy(dst[i], src[i])
	}
	return dst
}

func shapeMaskToFloatRings(session *shapeEditSession) [][][2]float32 {
	intRings := shapeMaskToIntRings(session)
	floatRings := make([][][2]float32, 0, len(intRings))
	for _, ring := range intRings {
		floatRing := make([][2]float32, len(ring))
		for i, pt := range ring {
			floatRing[i] = [2]float32{float32(pt[0]), float32(pt[1])}
		}
		floatRings = append(floatRings, floatRing)
	}
	return floatRings
}

func shapeMaskToIntRings(session *shapeEditSession) [][][2]int {
	if session == nil {
		return nil
	}
	edges := buildMaskEdges(session)
	if len(edges) == 0 {
		return nil
	}
	loops := make([][][2]int, 0, 8)
	for {
		start, ok := nextEdgeStart(edges)
		if !ok {
			break
		}
		loop := traceEdgeLoop(edges, start)
		loop = simplifyIntRing(loop)
		if len(loop) >= 3 {
			loops = append(loops, loop)
		}
	}
	sort.SliceStable(loops, func(i, j int) bool {
		return absInt(ringSignedArea2(loops[i])) > absInt(ringSignedArea2(loops[j]))
	})
	return loops
}

func buildMaskEdges(session *shapeEditSession) map[gridPoint][]gridPoint {
	edges := make(map[gridPoint][]gridPoint)
	add := func(a, b gridPoint) { edges[a] = append(edges[a], b) }
	for y := session.MinY; y <= session.MaxY; y++ {
		for x := session.MinX; x <= session.MaxX; x++ {
			if !session.filled(x, y) {
				continue
			}
			if !session.filled(x, y-1) {
				add(gridPoint{x, y}, gridPoint{x + 1, y})
			}
			if !session.filled(x+1, y) {
				add(gridPoint{x + 1, y}, gridPoint{x + 1, y + 1})
			}
			if !session.filled(x, y+1) {
				add(gridPoint{x + 1, y + 1}, gridPoint{x, y + 1})
			}
			if !session.filled(x-1, y) {
				add(gridPoint{x, y + 1}, gridPoint{x, y})
			}
		}
	}
	return edges
}

func nextEdgeStart(edges map[gridPoint][]gridPoint) (gridPoint, bool) {
	var best gridPoint
	found := false
	for p, next := range edges {
		if len(next) == 0 {
			continue
		}
		if !found || p.Y < best.Y || (p.Y == best.Y && p.X < best.X) {
			best = p
			found = true
		}
	}
	return best, found
}

func traceEdgeLoop(edges map[gridPoint][]gridPoint, start gridPoint) [][2]int {
	options := edges[start]
	if len(options) == 0 {
		return nil
	}
	first := options[0]
	edges[start] = options[1:]
	loop := [][2]int{{start.X, start.Y}, {first.X, first.Y}}
	prev := start
	cur := first
	for !(cur == start) {
		nexts := edges[cur]
		if len(nexts) == 0 {
			break
		}
		next, idx := pickNextEdge(prev, cur, nexts)
		edges[cur] = append(nexts[:idx], nexts[idx+1:]...)
		prev, cur = cur, next
		if cur != start {
			loop = append(loop, [2]int{cur.X, cur.Y})
		}
	}
	return loop
}

func pickNextEdge(prev, cur gridPoint, nexts []gridPoint) (gridPoint, int) {
	if len(nexts) == 1 {
		return nexts[0], 0
	}
	prevDir := edgeDir(prev, cur)
	bestIdx := 0
	bestScore := 99
	for i, next := range nexts {
		turn := (edgeDir(cur, next) - prevDir + 4) % 4
		score := mapTurnScore(turn)
		if score < bestScore {
			bestIdx = i
			bestScore = score
		}
	}
	return nexts[bestIdx], bestIdx
}

func edgeDir(a, b gridPoint) int {
	switch {
	case b.X > a.X:
		return 0
	case b.Y > a.Y:
		return 1
	case b.X < a.X:
		return 2
	default:
		return 3
	}
}

func mapTurnScore(turn int) int {
	switch turn {
	case 1:
		return 0
	case 0:
		return 1
	case 3:
		return 2
	default:
		return 3
	}
}

func simplifyIntRing(ring [][2]int) [][2]int {
	if len(ring) < 3 {
		return ring
	}
	changed := true
	for changed && len(ring) >= 3 {
		changed = false
		out := make([][2]int, 0, len(ring))
		for i := range ring {
			prev := ring[(i+len(ring)-1)%len(ring)]
			cur := ring[i]
			next := ring[(i+1)%len(ring)]
			if isCollinear(prev, cur, next) {
				changed = true
				continue
			}
			out = append(out, cur)
		}
		ring = out
	}
	return ring
}

func isCollinear(a, b, c [2]int) bool {
	return (b[0]-a[0])*(c[1]-b[1]) == (b[1]-a[1])*(c[0]-b[0])
}

func ringSignedArea2(ring [][2]int) int {
	area2 := 0
	for i := range ring {
		a := ring[i]
		b := ring[(i+1)%len(ring)]
		area2 += a[0]*b[1] - b[0]*a[1]
	}
	return area2
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
