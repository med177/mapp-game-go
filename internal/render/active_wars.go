package render

import (
	"image"
	"image/color"
	"sort"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	activeWarsPanelW         = 600.0
	activeWarsPanelMaxH      = 760.0
	activeWarsPanelPad       = 8.0
	activeWarsPanelHeaderH   = 38.0
	activeWarRowH            = 82.0
	activeWarRowGap          = 8.0
	activeWarFlagSize        = 44.0
	activeWarFlagSidePad     = 8.0
	activeWarFlagTextGap     = 8.0
	activeWarSidePad         = 8.0
	activeWarSideGap         = 8.0
	activeWarSideHeaderH     = 38.0
	activeWarParticipantH    = 50.0
	activeWarRowBottomPad    = 4.0
	activeWarParticipantFlag = 46.0
	activeWarRowRightInset   = 1.0
	activeWarsScrollbarGap   = 10.0
	activeWarDividerShift    = 4.0
	activeWarsHudButtonGap   = 8.0
	activeWarsHudButtonSize  = 36.0
	activeWarsHudButtonStep  = 46.0
	activeWarsHudButtonTop   = 8.0
	activeWarsPanelGap       = 12.0
)

// ActiveWarSummary, panelin harita state'inden bağımsız çizilebilir snapshot'ıdır.
// Güç değerleri mevcut ordulardan, süre ve kayıplar WarLedger'dan okunur.
type ActiveWarSummary struct {
	FactionANameTR string
	FactionBNameTR string
	FactionA       faction.FactionID
	FactionB       faction.FactionID
	SideA          ActiveWarSide
	SideB          ActiveWarSide
	Turns          int
	// These aggregate fields remain available to callers that only need the
	// compact totals. The panel renders the richer SideA/SideB snapshots.
	PowerA      int
	PowerB      int
	ArmiesA     int
	ArmiesB     int
	UnitsA      int
	UnitsB      int
	CasualtiesA int
	CasualtiesB int
}

type ActiveWarParticipant struct {
	FactionID       faction.FactionID
	NameTR          string
	Power           int
	Armies          int
	Units           int
	LandArmies      int
	NavalArmies     int
	LandUnits       int
	NavalUnits      int
	Casualties      int
	ArmyCasualties  int
	FleetCasualties int
}

type ActiveWarSide struct {
	Participants    []ActiveWarParticipant
	Power           int
	Armies          int
	Units           int
	LandArmies      int
	NavalArmies     int
	LandUnits       int
	NavalUnits      int
	Casualties      int
	ArmyCasualties  int
	FleetCasualties int
}

// topHudUtilityButtonRect, müzik kartının sağındaki ortak küçük düğme
// şeridindeki slotu döndürür. Yeni HUD durum düğmeleri aynı şeritte slot
// index'i artırılarak yan yana eklenebilir.
func topHudUtilityButtonRect(slot int) [4]float32 {
	if slot < 0 {
		slot = 0
	}
	x, y, w, _ := musicHudRect()
	return [4]float32{x + w + activeWarsHudButtonGap + float32(slot)*activeWarsHudButtonStep, y + activeWarsHudButtonTop, activeWarsHudButtonSize, activeWarsHudButtonSize}
}

func activeWarsHudButtonRect() [4]float32 {
	return topHudUtilityButtonRect(0)
}

func buildActiveWarsHUDButton() gameui.Button {
	r := activeWarsHudButtonRect()
	btn := buttonFromRectF32(r, "").WithIcon(gameui.IconSword)
	btn.IconSize = 21
	return btn
}

func activeWarsHudButtonHit(mx, my float64) bool {
	return buildActiveWarsHUDButton().HitTest(mx, my)
}

func activeWarsPanelRect() gameui.Rect {
	_, musicY, _, musicH := musicHudRect()
	y := float64(musicY + musicH + 6)
	h := activeWarsPanelMaxH
	if remaining := ScreenHeight - y - 16; remaining < h {
		h = remaining
	}
	if h < 150 {
		h = 150
	}
	// Olaylar paneli ekranın sağ kenarında kalır; aktif savaşlar onun
	// solundaki ayrılmış kolona oturur ve iki panel yatayda çakışmaz.
	right := float64(evLogX()) - activeWarsPanelGap
	if right < 10 {
		right = 10
	}
	panelW := activeWarsPanelW
	if available := right - 10; available < panelW {
		panelW = available
	}
	if panelW < 0 {
		panelW = 0
	}
	return gameui.Rect{
		X: right - panelW,
		Y: y,
		W: panelW,
		H: h,
	}
}

func activeWarsPanelCloseButton() gameui.Button {
	panel := activeWarsPanelRect()
	btn := gameui.NewButton(panel.X+panel.W-38, panel.Y+9, 26, 24, "").WithIcon(gameui.IconClose)
	btn.IconSize = 13
	return btn
}

func activeWarsPanelViewport() gameui.Rect {
	panel := activeWarsPanelRect()
	return gameui.Rect{
		X: panel.X + activeWarsPanelPad,
		Y: panel.Y + activeWarsPanelHeaderH,
		W: panel.W - activeWarsPanelPad*2 - activeWarsScrollbarGap,
		H: panel.H - activeWarsPanelHeaderH - activeWarsPanelPad,
	}
}

func activeWarsScrollbarRect(viewport gameui.Rect) gameui.Rect {
	return gameui.Rect{
		X: viewport.X + viewport.W + activeWarsScrollbarGap/2 - 1.5,
		Y: viewport.Y,
		W: 3,
		H: viewport.H,
	}
}

func activeWarsPanelHit(mx, my float64) bool {
	return activeWarsPanelRect().Hit(mx, my)
}

// activeWarsPanelInteractiveHit yalnızca gerçekten tıklanabilen savaş satırları
// ve ortak kapatma düğmesi için pointer cursor üretir. Panelin geri kalanı
// bilgi/scroll yüzeyidir; dışı ise harita input'una bırakılır.
func activeWarsPanelInteractiveHit(r *Renderer, mx, my float64) bool {
	if r == nil || !r.showActiveWars {
		return false
	}
	if activeWarsPanelCloseButton().HitTest(mx, my) {
		return true
	}
	return activeWarRowAt(mx, my, r.activeWarsBuf, r.activeWarsScroll) >= 0
}

func activeWarVisibleRows(viewport gameui.Rect) int {
	rows := int((viewport.H + activeWarRowGap) / (activeWarRowH + activeWarRowGap))
	if rows < 1 {
		return 1
	}
	return rows
}

func activeWarMaxScroll(entryCount int, viewport gameui.Rect) int {
	maxScroll := entryCount - activeWarVisibleRows(viewport)
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func clampActiveWarScroll(entryCount int, viewport gameui.Rect, scroll int) int {
	if scroll < 0 {
		return 0
	}
	if maxScroll := activeWarMaxScroll(entryCount, viewport); scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func activeWarRowRect(viewport gameui.Rect, visibleIndex int) gameui.Rect {
	return gameui.Rect{
		X: viewport.X,
		Y: viewport.Y + float64(visibleIndex)*(activeWarRowH+activeWarRowGap),
		W: viewport.W,
		H: activeWarRowH,
	}
}

func activeWarRowHeight(war ActiveWarSummary) float64 {
	participantCount := len(war.SideA.Participants)
	if len(war.SideB.Participants) > participantCount {
		participantCount = len(war.SideB.Participants)
	}
	height := activeWarSideHeaderH + float64(participantCount)*activeWarParticipantH + activeWarRowBottomPad
	if height < activeWarRowH {
		return activeWarRowH
	}
	return height
}

func activeWarRowsHeight(wars []ActiveWarSummary, start, end int) float64 {
	if start < 0 {
		start = 0
	}
	if end > len(wars) {
		end = len(wars)
	}
	if start >= end {
		return 0
	}
	height := 0.0
	for index := start; index < end; index++ {
		if index > start {
			height += activeWarRowGap
		}
		height += activeWarRowHeight(wars[index])
	}
	return height
}

func activeWarMaxScrollForWars(wars []ActiveWarSummary, viewport gameui.Rect) int {
	for start := 0; start < len(wars); start++ {
		if activeWarRowsHeight(wars, start, len(wars)) <= viewport.H {
			return start
		}
	}
	if len(wars) == 0 {
		return 0
	}
	return len(wars) - 1
}

func clampActiveWarScrollForWars(wars []ActiveWarSummary, viewport gameui.Rect, scroll int) int {
	if scroll < 0 {
		return 0
	}
	if maxScroll := activeWarMaxScrollForWars(wars, viewport); scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func activeWarRowRectForWars(viewport gameui.Rect, wars []ActiveWarSummary, index, scroll int) gameui.Rect {
	y := viewport.Y
	for current := scroll; current < index; current++ {
		y += activeWarRowHeight(wars[current]) + activeWarRowGap
	}
	return gameui.Rect{
		X: viewport.X,
		Y: y,
		// StrokeRect merkezini rect sınırına çizer. Viewport'un sağ sınırında
		// bırakılırsa border'ın dış yarısı SubImage tarafından kırpılır.
		W: viewport.W - activeWarRowRightInset,
		H: activeWarRowHeight(wars[index]),
	}
}

func activeWarVisibleRowsForWars(viewport gameui.Rect, wars []ActiveWarSummary, scroll int) int {
	visible := 0
	used := 0.0
	for index := scroll; index < len(wars); index++ {
		height := activeWarRowHeight(wars[index])
		if visible > 0 {
			height += activeWarRowGap
		}
		if used >= viewport.H {
			break
		}
		used += height
		visible++
	}
	if visible < 1 && len(wars) > 0 {
		return 1
	}
	return visible
}

func activeWarRowAt(mx, my float64, wars []ActiveWarSummary, scroll int) int {
	viewport := activeWarsPanelViewport()
	if !viewport.Hit(mx, my) {
		return -1
	}
	scroll = clampActiveWarScrollForWars(wars, viewport, scroll)
	end := scroll + activeWarVisibleRowsForWars(viewport, wars, scroll)
	if end > len(wars) {
		end = len(wars)
	}
	for index := scroll; index < end; index++ {
		if activeWarRowRectForWars(viewport, wars, index, scroll).Hit(mx, my) {
			return index
		}
	}
	return -1
}

func activeWarFactionName(gs *state.GameState, id faction.FactionID) string {
	if gs != nil {
		if f := gs.Factions[id]; f != nil {
			if f.NameTR != "" {
				return f.NameTR
			}
			if f.Name != "" {
				return f.Name
			}
		}
	}
	return string(id)
}

func activeWarArmyStats(gs *state.GameState, owner faction.FactionID) (armies, units int) {
	landArmies, navalArmies, landUnits, navalUnits := activeWarArmyBreakdown(gs, owner)
	return landArmies + navalArmies, landUnits + navalUnits
}

func activeWarArmyBreakdown(gs *state.GameState, owner faction.FactionID) (landArmies, navalArmies, landUnits, navalUnits int) {
	if gs == nil {
		return 0, 0, 0, 0
	}
	for _, a := range gs.Armies {
		if a == nil || a.OwnerID != string(owner) {
			continue
		}
		if a.IsNaval {
			navalArmies++
			navalUnits += len(a.Units)
			continue
		}
		landArmies++
		landUnits += len(a.Units) + len(a.EmbarkedUnits)
	}
	return landArmies, navalArmies, landUnits, navalUnits
}

func activeWarFactionIsVassal(gs *state.GameState, id faction.FactionID) bool {
	if gs == nil {
		return false
	}
	f := gs.Factions[id]
	return f != nil && !f.IsEliminated && f.OverlordID != ""
}

type activeWarRelation struct {
	a                faction.FactionID
	b                faction.FactionID
	startedTurn      int
	casualtiesA      int
	casualtiesB      int
	armyCasualtiesA  int
	armyCasualtiesB  int
	fleetCasualtiesA int
	fleetCasualtiesB int
}

func collectActiveWarRelations(gs *state.GameState) []activeWarRelation {
	if gs == nil {
		return nil
	}
	capacity := len(gs.Relations)
	if capacity > 16 {
		capacity = 16
	}
	relations := make([]activeWarRelation, 0, capacity)
	for _, rel := range gs.Relations {
		if rel == nil || rel.Stance != faction.StanceWar || rel.FactionA == "" || rel.FactionB == "" || rel.FactionA == rel.FactionB {
			continue
		}
		a, b := faction.FactionID(rel.FactionA), faction.FactionID(rel.FactionB)
		if factionA := gs.Factions[a]; factionA == nil || factionA.IsEliminated {
			continue
		}
		if factionB := gs.Factions[b]; factionB == nil || factionB.IsEliminated {
			continue
		}
		if activeWarFactionIsVassal(gs, a) || activeWarFactionIsVassal(gs, b) {
			continue
		}

		startedTurn := gs.Turn
		casualtiesA, casualtiesB := 0, 0
		armyCasualtiesA, armyCasualtiesB := 0, 0
		fleetCasualtiesA, fleetCasualtiesB := 0, 0
		ledger := gs.WarLedgerFor(a, b)
		if ledger != nil {
			startedTurn = ledger.StartedTurn
			casualtiesA = ledger.CasualtiesA
			casualtiesB = ledger.CasualtiesB
			armyCasualtiesA = ledger.CasualtiesArmyA
			armyCasualtiesB = ledger.CasualtiesArmyB
			fleetCasualtiesA = ledger.CasualtiesFleetA
			fleetCasualtiesB = ledger.CasualtiesFleetB
			// Eski kayıtlarda yalnız toplam kayıp alanı vardır. Bu kayıpları
			// görünürlük kaybı yaşamaması için kara ordusu kaybı kabul et.
			if casualtiesA > 0 && armyCasualtiesA+fleetCasualtiesA == 0 {
				armyCasualtiesA = casualtiesA
			}
			if casualtiesB > 0 && armyCasualtiesB+fleetCasualtiesB == 0 {
				armyCasualtiesB = casualtiesB
			}
			if ledger.DeclarerFactionID != "" && ledger.DefenderFactionID != "" &&
				ledger.DeclarerFactionID != ledger.DefenderFactionID &&
				((ledger.DeclarerFactionID == a && ledger.DefenderFactionID == b) ||
					(ledger.DeclarerFactionID == b && ledger.DefenderFactionID == a)) {
				a, b = ledger.DeclarerFactionID, ledger.DefenderFactionID
			}
			if a != ledger.FactionA {
				casualtiesA, casualtiesB = casualtiesB, casualtiesA
				armyCasualtiesA, armyCasualtiesB = armyCasualtiesB, armyCasualtiesA
				fleetCasualtiesA, fleetCasualtiesB = fleetCasualtiesB, fleetCasualtiesA
			}
		}
		relations = append(relations, activeWarRelation{
			a:                a,
			b:                b,
			startedTurn:      startedTurn,
			casualtiesA:      casualtiesA,
			casualtiesB:      casualtiesB,
			armyCasualtiesA:  armyCasualtiesA,
			armyCasualtiesB:  armyCasualtiesB,
			fleetCasualtiesA: fleetCasualtiesA,
			fleetCasualtiesB: fleetCasualtiesB,
		})
	}
	sort.SliceStable(relations, func(i, j int) bool {
		if relations[i].a != relations[j].a {
			return relations[i].a < relations[j].a
		}
		return relations[i].b < relations[j].b
	})
	return relations
}

func activeWarSameSide(gs *state.GameState, a, b faction.FactionID) bool {
	if a == b || diplomacy.SameRealm(gs, a, b) {
		return true
	}
	rel := diplomacy.Relation(gs, a, b)
	return rel != nil && rel.Stance == faction.StanceAllied
}

func activeWarRelationsSameGroup(gs *state.GameState, first, second activeWarRelation) bool {
	return activeWarSameSide(gs, first.a, second.a) && activeWarSameSide(gs, first.b, second.b)
}

func groupActiveWarRelations(gs *state.GameState, relations []activeWarRelation) [][]activeWarRelation {
	if len(relations) == 0 {
		return nil
	}
	parent := make([]int, len(relations))
	for index := range parent {
		parent[index] = index
	}
	var find func(int) int
	find = func(index int) int {
		if parent[index] != index {
			parent[index] = find(parent[index])
		}
		return parent[index]
	}
	union := func(left, right int) {
		leftRoot, rightRoot := find(left), find(right)
		if leftRoot != rightRoot {
			parent[rightRoot] = leftRoot
		}
	}
	for left := 0; left < len(relations); left++ {
		for right := left + 1; right < len(relations); right++ {
			if activeWarRelationsSameGroup(gs, relations[left], relations[right]) {
				union(left, right)
			}
		}
	}

	groupsByRoot := make(map[int][]activeWarRelation, len(relations))
	for index, relation := range relations {
		root := find(index)
		groupsByRoot[root] = append(groupsByRoot[root], relation)
	}
	groups := make([][]activeWarRelation, 0, len(groupsByRoot))
	for _, group := range groupsByRoot {
		sort.SliceStable(group, func(i, j int) bool {
			if group[i].startedTurn != group[j].startedTurn {
				return group[i].startedTurn < group[j].startedTurn
			}
			if group[i].a != group[j].a {
				return group[i].a < group[j].a
			}
			return group[i].b < group[j].b
		})
		groups = append(groups, group)
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i][0].startedTurn != groups[j][0].startedTurn {
			return groups[i][0].startedTurn < groups[j][0].startedTurn
		}
		if groups[i][0].a != groups[j][0].a {
			return groups[i][0].a < groups[j][0].a
		}
		return groups[i][0].b < groups[j][0].b
	})
	return groups
}

func addActiveWarParticipant(gs *state.GameState, side *ActiveWarSide, seen map[faction.FactionID]int, id faction.FactionID, casualties, armyCasualties, fleetCasualties int) {
	if index, exists := seen[id]; exists {
		side.Participants[index].Casualties += casualties
		side.Participants[index].ArmyCasualties += armyCasualties
		side.Participants[index].FleetCasualties += fleetCasualties
		side.Casualties += casualties
		side.ArmyCasualties += armyCasualties
		side.FleetCasualties += fleetCasualties
		return
	}
	landArmies, navalArmies, landUnits, navalUnits := activeWarArmyBreakdown(gs, id)
	participant := ActiveWarParticipant{
		FactionID:       id,
		NameTR:          activeWarFactionName(gs, id),
		Power:           diplomacy.MilitaryPower(gs, id),
		Armies:          landArmies + navalArmies,
		Units:           landUnits + navalUnits,
		LandArmies:      landArmies,
		NavalArmies:     navalArmies,
		LandUnits:       landUnits,
		NavalUnits:      navalUnits,
		Casualties:      casualties,
		ArmyCasualties:  armyCasualties,
		FleetCasualties: fleetCasualties,
	}
	seen[id] = len(side.Participants)
	side.Participants = append(side.Participants, participant)
	side.Power += participant.Power
	side.Armies += participant.Armies
	side.Units += participant.Units
	side.LandArmies += participant.LandArmies
	side.NavalArmies += participant.NavalArmies
	side.LandUnits += participant.LandUnits
	side.NavalUnits += participant.NavalUnits
	side.Casualties += participant.Casualties
	side.ArmyCasualties += participant.ArmyCasualties
	side.FleetCasualties += participant.FleetCasualties
}

func addActiveWarVassalParticipants(gs *state.GameState, side *ActiveWarSide, seen map[faction.FactionID]int) {
	if gs == nil || side == nil {
		return
	}
	// Vassallar ayrı bir savaş ilişkisi olarak tutulmayabilir; overlord'un
	// savaşa girmesiyle aynı cepheye kesin katılırlar. Bu nedenle onları
	// relation listesinden değil, mevcut taraf katılımcılarının realm'inden
	// genişletiyoruz.
	baseParticipants := append([]ActiveWarParticipant(nil), side.Participants...)
	for _, participant := range baseParticipants {
		for _, vassal := range diplomacy.VassalsOf(gs, participant.FactionID) {
			addActiveWarParticipant(gs, side, seen, vassal, 0, 0, 0)
		}
	}
}

func sortActiveWarParticipants(participants []ActiveWarParticipant, primary faction.FactionID) {
	sort.SliceStable(participants, func(i, j int) bool {
		iPrimary := participants[i].FactionID == primary
		jPrimary := participants[j].FactionID == primary
		if iPrimary != jPrimary {
			return iPrimary
		}
		if participants[i].Power != participants[j].Power {
			return participants[i].Power > participants[j].Power
		}
		if participants[i].FactionID != participants[j].FactionID {
			return participants[i].FactionID < participants[j].FactionID
		}
		return participants[i].NameTR < participants[j].NameTR
	})
}

func buildActiveWarSummary(gs *state.GameState, relations []activeWarRelation) ActiveWarSummary {
	first := relations[0]
	summary := ActiveWarSummary{
		FactionA:       first.a,
		FactionB:       first.b,
		FactionANameTR: activeWarFactionName(gs, first.a),
		FactionBNameTR: activeWarFactionName(gs, first.b),
	}
	if gs != nil {
		summary.Turns = gs.Turn - first.startedTurn
		if summary.Turns < 0 {
			summary.Turns = 0
		}
	}
	seenA := make(map[faction.FactionID]int, len(relations))
	seenB := make(map[faction.FactionID]int, len(relations))
	for _, relation := range relations {
		addActiveWarParticipant(gs, &summary.SideA, seenA, relation.a, relation.casualtiesA, relation.armyCasualtiesA, relation.fleetCasualtiesA)
		addActiveWarParticipant(gs, &summary.SideB, seenB, relation.b, relation.casualtiesB, relation.armyCasualtiesB, relation.fleetCasualtiesB)
	}
	addActiveWarVassalParticipants(gs, &summary.SideA, seenA)
	addActiveWarVassalParticipants(gs, &summary.SideB, seenB)
	sortActiveWarParticipants(summary.SideA.Participants, summary.FactionA)
	sortActiveWarParticipants(summary.SideB.Participants, summary.FactionB)
	summary.PowerA = summary.SideA.Power
	summary.PowerB = summary.SideB.Power
	summary.ArmiesA = summary.SideA.Armies
	summary.ArmiesB = summary.SideB.Armies
	summary.UnitsA = summary.SideA.Units
	summary.UnitsB = summary.SideB.Units
	summary.CasualtiesA = summary.SideA.Casualties
	summary.CasualtiesB = summary.SideB.Casualties
	return summary
}

// collectActiveWarSummaries ilişkileri önce ortak cephe üyeliğine göre
// gruplar; böylece bir koalisyonun devlet-devlet kombinasyonları tek satırda
// görünür. dst Renderer tarafından yeniden kullanıldığı için normal Draw
// akışında özet slice'ı yeniden ayrıştırılmaz.
func collectActiveWarSummaries(gs *state.GameState, dst []ActiveWarSummary) []ActiveWarSummary {
	dst = dst[:0]
	groups := groupActiveWarRelations(gs, collectActiveWarRelations(gs))
	for _, group := range groups {
		dst = append(dst, buildActiveWarSummary(gs, group))
	}
	return dst
}

func countActiveWars(gs *state.GameState) int {
	return len(groupActiveWarRelations(gs, collectActiveWarRelations(gs)))
}

func activeWarRepresentativeRegion(gs *state.GameState, fid faction.FactionID) *world.Region {
	if gs == nil || fid == "" {
		return nil
	}
	if capital, _, _, ok := gs.FactionCapital(fid); ok && capital != nil {
		return capital
	}
	var representative *world.Region
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		if representative == nil || region.ID < representative.ID {
			representative = region
		}
	}
	return representative
}

func (r *Renderer) focusActiveWar(war ActiveWarSummary) bool {
	if r == nil || r.gs == nil {
		return false
	}
	regionA := activeWarRepresentativeRegion(r.gs, war.FactionA)
	regionB := activeWarRepresentativeRegion(r.gs, war.FactionB)
	if regionA == nil && regionB == nil {
		return false
	}

	focusX, focusY := 0.0, 0.0
	switch {
	case regionA != nil && regionB != nil:
		focusX = (wcX(regionA.WorldX) + wcX(regionB.WorldX)) / 2
		focusY = (wcY(regionA.WorldY) + wcY(regionB.WorldY)) / 2
	case regionA != nil:
		focusX = wcX(regionA.WorldX)
		focusY = wcY(regionA.WorldY)
	default:
		focusX = wcX(regionB.WorldX)
		focusY = wcY(regionB.WorldY)
	}
	r.camX, r.camY = clampCameraCenter(focusX, focusY, r.camScale)
	return true
}

func drawActiveWarsHUDButton(screen *ebiten.Image, gs *state.GameState, open bool) {
	btn := buildActiveWarsHUDButton()
	centerX := float32(btn.X + btn.W/2)
	centerY := float32(btn.Y + btn.H/2)
	fill := color.RGBA{58, 38, 24, 235}
	border := color.RGBA{166, 124, 60, 255}
	if open {
		fill = color.RGBA{144, 76, 36, 250}
		border = color.RGBA{255, 208, 104, 255}
	} else if countActiveWars(gs) == 0 {
		fill = color.RGBA{38, 34, 28, 220}
		border = color.RGBA{110, 100, 82, 220}
	}
	vector.FillCircle(screen, centerX, centerY, float32(btn.W/2), fill, true)
	vector.StrokeCircle(screen, centerX, centerY, float32(btn.W/2-1), 1.5, border, true)
	gameui.DrawIcon(screen, btn.Icon, btn.X+(btn.W-btn.IconSize)/2, btn.Y+(btn.H-btn.IconSize)/2, btn.IconSize, ColorWhite)

	count := countActiveWars(gs)
	if count > 0 {
		badgeX, badgeY := centerX+8, centerY-12
		vector.FillCircle(screen, badgeX, badgeY, 8, color.RGBA{170, 42, 32, 255}, true)
		drawUILabel(screen, gameui.Rect{X: float64(badgeX - 8), Y: float64(badgeY - 7), W: 16, H: 14}, itoa(count), ColorWhite, gameui.TextSmall, gameui.TextAlignCenter)
	}
}

func drawActiveWarsScrollbar(screen *ebiten.Image, viewport gameui.Rect, entryCount, scroll int) {
	maxScroll := activeWarMaxScroll(entryCount, viewport)
	if maxScroll <= 0 {
		return
	}
	track := activeWarsScrollbarRect(viewport)
	drawUICardRect(screen, track, color.RGBA{42, 35, 25, 210}, color.RGBA{90, 72, 44, 180}, 1)
	thumbH := track.H * float64(activeWarVisibleRows(viewport)) / float64(entryCount)
	if thumbH < 24 {
		thumbH = 24
	}
	scroll = clampActiveWarScroll(entryCount, viewport, scroll)
	thumbY := track.Y
	if track.H > thumbH {
		thumbY += (track.H - thumbH) * float64(scroll) / float64(maxScroll)
	}
	drawUICardRect(screen, gameui.Rect{X: track.X, Y: thumbY, W: track.W, H: thumbH}, color.RGBA{190, 148, 74, 235}, color.RGBA{238, 206, 130, 220}, 1)
}

func drawActiveWarsScrollbarForWars(screen *ebiten.Image, viewport gameui.Rect, wars []ActiveWarSummary, scroll int) {
	maxScroll := activeWarMaxScrollForWars(wars, viewport)
	if maxScroll <= 0 {
		return
	}
	scroll = clampActiveWarScrollForWars(wars, viewport, scroll)
	visibleRows := activeWarVisibleRowsForWars(viewport, wars, scroll)
	track := activeWarsScrollbarRect(viewport)
	drawUICardRect(screen, track, color.RGBA{42, 35, 25, 210}, color.RGBA{90, 72, 44, 180}, 1)
	thumbH := track.H * float64(visibleRows) / float64(len(wars))
	if thumbH < 24 {
		thumbH = 24
	}
	thumbY := track.Y
	if track.H > thumbH {
		thumbY += (track.H - thumbH) * float64(scroll) / float64(maxScroll)
	}
	drawUICardRect(screen, gameui.Rect{X: track.X, Y: thumbY, W: track.W, H: thumbH}, color.RGBA{190, 148, 74, 235}, color.RGBA{238, 206, 130, 220}, 1)
}

func activeWarFactionFlagColor(gs *state.GameState, id faction.FactionID) color.RGBA {
	if gs != nil {
		if f := gs.Factions[id]; f != nil {
			return color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		}
	}
	return color.RGBA{62, 52, 38, 255}
}

func activeWarRowContentRects(row gameui.Rect) (leftFlag, center, rightFlag gameui.Rect) {
	leftFlag = gameui.Rect{
		X: row.X + activeWarFlagSidePad,
		Y: row.Y + (row.H-activeWarFlagSize)/2,
		W: activeWarFlagSize,
		H: activeWarFlagSize,
	}
	rightFlag = gameui.Rect{
		X: row.X + row.W - activeWarFlagSidePad - activeWarFlagSize,
		Y: leftFlag.Y,
		W: activeWarFlagSize,
		H: activeWarFlagSize,
	}
	center = gameui.Rect{
		X: leftFlag.X + leftFlag.W + activeWarFlagTextGap,
		Y: row.Y + 4,
		W: rightFlag.X - (leftFlag.X + leftFlag.W) - activeWarFlagTextGap*2,
		H: row.H - 8,
	}
	return leftFlag, center, rightFlag
}

func activeWarSideRect(row gameui.Rect, sideIndex int) gameui.Rect {
	width := (row.W - activeWarSidePad*2 - activeWarSideGap) / 2
	return gameui.Rect{
		X: row.X + activeWarSidePad + float64(sideIndex)*(width+activeWarSideGap),
		Y: row.Y,
		W: width,
		H: row.H,
	}
}

func drawActiveWarSide(screen *ebiten.Image, gs *state.GameState, rect gameui.Rect, label string, side ActiveWarSide) {
	drawUILabel(screen, gameui.Rect{X: rect.X, Y: rect.Y + 5, W: rect.W}, label, color.RGBA{255, 220, 118, 255}, gameui.TextSmall, gameui.TextAlignStart)
	summary := "Güç " + itoa(side.Power) + " • " + itoa(side.LandArmies) + " Ordu • " + itoa(side.NavalArmies) + " Filo • " + itoa(side.ArmyCasualties) + "/" + itoa(side.FleetCasualties) + " Kayıp"
	drawUILabel(screen, gameui.Rect{X: rect.X, Y: rect.Y + 20, W: rect.W}, trimTextToWidth(summary, FaceSmall, rect.W), color.RGBA{210, 194, 160, 255}, gameui.TextSmall, gameui.TextAlignStart)
	for index, participant := range side.Participants {
		y := rect.Y + activeWarSideHeaderH + float64(index)*activeWarParticipantH
		flag := gameui.Rect{X: rect.X, Y: y + 3, W: activeWarParticipantFlag, H: activeWarParticipantFlag}
		drawFactionFlagBadge(screen, participant.FactionID, factionInitial(participant.NameTR), flag.X, flag.Y, flag.W, activeWarFactionFlagColor(gs, participant.FactionID), panelBorder)
		textX := flag.X + flag.W + 6
		textW := rect.W - (textX - rect.X)
		power := "Güç " + itoa(participant.Power)
		drawUILabel(screen, gameui.Rect{X: textX, Y: y + 1, W: textW}, trimTextToWidth(power+" • "+participant.NameTR, FaceSmall, textW), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
		army := itoa(participant.LandArmies) + " Ordu • " + itoa(participant.LandUnits) + " birim • " + itoa(participant.ArmyCasualties) + " Kayıp"
		navy := itoa(participant.NavalArmies) + " Filo • " + itoa(participant.NavalUnits) + " birim • " + itoa(participant.FleetCasualties) + " Kayıp"
		drawUILabel(screen, gameui.Rect{X: textX, Y: y + 15, W: textW}, trimTextToWidth(army, FaceSmall, textW), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: textX, Y: y + 30, W: textW}, trimTextToWidth(navy, FaceSmall, textW), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	}
}

func drawActiveWarRow(screen *ebiten.Image, gs *state.GameState, row gameui.Rect, war ActiveWarSummary) {
	drawUICardRect(screen, row, color.RGBA{28, 22, 16, 235}, color.RGBA{92, 68, 38, 215}, 1)
	left := activeWarSideRect(row, 0)
	right := activeWarSideRect(row, 1)
	dividerX := float32(row.X + row.W/2 - activeWarDividerShift)
	vector.StrokeLine(screen, dividerX, float32(row.Y+8), dividerX, float32(row.Y+row.H-8), 1, color.RGBA{92, 68, 38, 180}, true)
	drawActiveWarSide(screen, gs, left, "Saldıran taraf • "+itoa(war.Turns)+" tur", war.SideA)
	drawActiveWarSide(screen, gs, right, "Savunan taraf", war.SideB)
}

func drawActiveWarsPanel(screen *ebiten.Image, gs *state.GameState, wars []ActiveWarSummary, scroll int) {
	panel := activeWarsPanelRect()
	drawUIPanelFrame(screen, panel, color.RGBA{12, 10, 8, 244}, color.RGBA{154, 112, 54, 255}, 1.5, 5)
	drawUILabel(screen, gameui.Rect{X: panel.X + activeWarsPanelPad, Y: panel.Y + 9, W: panel.W - 60}, "Aktif Savaşlar ("+itoa(len(wars))+")", color.RGBA{255, 220, 118, 255}, gameui.TextMedium, gameui.TextAlignStart)
	drawUIButtonWidget(screen, activeWarsPanelCloseButton(), eventLogButtonStyle(ColorGray))

	viewport := activeWarsPanelViewport()
	if len(wars) == 0 {
		drawUILabel(screen, viewport, "Aktif savaş bulunmuyor.", ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
		return
	}
	scroll = clampActiveWarScrollForWars(wars, viewport, scroll)
	visibleRows := activeWarVisibleRowsForWars(viewport, wars, scroll)
	end := scroll + visibleRows
	if end > len(wars) {
		end = len(wars)
	}
	left, top := int(viewport.X), int(viewport.Y)
	right, bottom := int(viewport.X+viewport.W), int(viewport.Y+viewport.H)
	if right <= left || bottom <= top {
		return
	}
	body := screen.SubImage(image.Rect(left, top, right, bottom)).(*ebiten.Image)
	for i := scroll; i < end; i++ {
		row := activeWarRowRectForWars(viewport, wars, i, scroll)
		if row.Y >= viewport.Y+viewport.H {
			break
		}
		drawActiveWarRow(body, gs, row, wars[i])
	}
	drawActiveWarsScrollbarForWars(screen, viewport, wars, scroll)
}

// handleActiveWarsOverlayInput yalnız panelin kendi yüzeyindeki input'u tüketir.
// Panel dışındaki sol/sağ tıklama ve orta tuş sürüklemesi ana harita akışına kalır.
func (r *Renderer) handleActiveWarsOverlayInput() bool {
	if r == nil || !r.showActiveWars {
		return false
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.showActiveWars = false
		r.activeWarsScroll = 0
		return true
	}
	mx, my := ebiten.CursorPosition()
	if !activeWarsPanelHit(float64(mx), float64(my)) {
		return false
	}
	if _, wheelY := ebiten.Wheel(); wheelY != 0 {
		viewport := activeWarsPanelViewport()
		r.activeWarsScroll = clampActiveWarScrollForWars(r.activeWarsBuf, viewport, r.activeWarsScroll-int(wheelY))
		return true
	}
	if activeWarsPanelCloseButton().HitTest(float64(mx), float64(my)) && r.mouseJustPressed(ebiten.MouseButtonLeft) {
		r.showActiveWars = false
		r.activeWarsScroll = 0
		return true
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if index := activeWarRowAt(float64(mx), float64(my), r.activeWarsBuf, r.activeWarsScroll); index >= 0 {
			r.focusActiveWar(r.activeWarsBuf[index])
			return true
		}
	}
	leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	leftWasPressed := r.prevMouse[ebiten.MouseButtonLeft]
	if leftPressed || leftWasPressed {
		r.prevMouse[ebiten.MouseButtonLeft] = leftPressed
		return true
	}
	rightPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	rightWasPressed := r.prevMouse[ebiten.MouseButtonRight]
	if rightPressed || rightWasPressed {
		r.prevMouse[ebiten.MouseButtonRight] = rightPressed
		return true
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		r.isDragging = false
		return true
	}
	return false
}
