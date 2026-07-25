package game

import (
	"fmt"
	"math"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	productionKindBuilding  = "building"
	productionKindUnit      = "unit"
	autoPortSettlementInset = 2.0
)

type productionResult struct {
	factionID faction.FactionID
	regionID  world.RegionID
	kind      string
	typeID    string
	delayed   bool
	canceled  bool
	reason    string
}

func (g *Game) enqueueProduction(kind string, rid world.RegionID, typeID string, turns int) state.ProductionOrder {
	if turns < 1 {
		turns = 1
	}
	g.gs.NextProductionSeq++
	order := state.ProductionOrder{
		ID:        fmt.Sprintf("prod_%d", g.gs.NextProductionSeq),
		Kind:      kind,
		FactionID: string(g.gs.PlayerFactionID),
		RegionID:  rid,
		TypeID:    typeID,
		TurnsLeft: turns,
	}
	g.gs.ProductionQueue = append(g.gs.ProductionQueue, order)
	return order
}

func (g *Game) cancelProduction(kind string, rid world.RegionID, typeID string, ownerID faction.FactionID) bool {
	for i := len(g.gs.ProductionQueue) - 1; i >= 0; i-- {
		order := g.gs.ProductionQueue[i]
		if order.Kind != kind || order.RegionID != rid || order.TypeID != typeID || order.FactionID != string(ownerID) {
			continue
		}
		copy(g.gs.ProductionQueue[i:], g.gs.ProductionQueue[i+1:])
		last := len(g.gs.ProductionQueue) - 1
		g.gs.ProductionQueue[last] = state.ProductionOrder{}
		g.gs.ProductionQueue = g.gs.ProductionQueue[:last]
		return true
	}
	return false
}

// hasProduction iptal etmeden, kuyrukta eşleşen üretim emri olup olmadığını döndürür.
func (g *Game) hasProduction(kind string, rid world.RegionID, typeID string, ownerID faction.FactionID) (state.ProductionOrder, bool) {
	for i := len(g.gs.ProductionQueue) - 1; i >= 0; i-- {
		order := g.gs.ProductionQueue[i]
		if order.Kind == kind && order.RegionID == rid && order.TypeID == typeID && order.FactionID == string(ownerID) {
			return order, true
		}
	}
	return state.ProductionOrder{}, false
}

func (g *Game) cancelProductionByID(orderID, kind string, rid world.RegionID, ownerID faction.FactionID) (state.ProductionOrder, bool) {
	for i := len(g.gs.ProductionQueue) - 1; i >= 0; i-- {
		order := g.gs.ProductionQueue[i]
		if order.ID != orderID || order.Kind != kind || order.RegionID != rid || order.FactionID != string(ownerID) {
			continue
		}
		copy(g.gs.ProductionQueue[i:], g.gs.ProductionQueue[i+1:])
		last := len(g.gs.ProductionQueue) - 1
		g.gs.ProductionQueue[last] = state.ProductionOrder{}
		g.gs.ProductionQueue = g.gs.ProductionQueue[:last]
		return order, true
	}
	return state.ProductionOrder{}, false
}

func (g *Game) applyProductionTicks() []productionResult {
	type productionCapacityKey struct {
		regionID world.RegionID
		lane     string
	}

	queue := g.gs.ProductionQueue
	remaining := queue[:0]
	var results []productionResult
	progressedUnitsByLane := make(map[productionCapacityKey]int)

	for _, order := range queue {
		result := productionResult{
			factionID: faction.FactionID(order.FactionID),
			regionID:  order.RegionID,
			kind:      order.Kind,
			typeID:    order.TypeID,
		}

		region, ok := g.gs.Regions[order.RegionID]
		if !ok || region.OwnerID != order.FactionID {
			result.canceled = true
			result.reason = "bölge artık bu fraksiyona ait değil"
			results = append(results, result)
			continue
		}

		if order.Kind == productionKindBuilding || order.Kind == productionKindUnit {
			if g.gs.SiegeAt(region.ID) != nil {
				remaining = append(remaining, order)
				continue
			}
		}

		switch order.Kind {
		case productionKindBuilding:
			order.TurnsLeft--
			if order.TurnsLeft > 0 {
				remaining = append(remaining, order)
				continue
			}
			if region.IsLocked {
				order.TurnsLeft = 1
				remaining = append(remaining, order)
				result.delayed = true
				result.reason = "bölge kilitli"
				results = append(results, result)
				continue
			}
			if !g.completeBuilding(region, order.TypeID) {
				result.canceled = true
				result.reason = "bina zaten tamamlanmış"
			}
			results = append(results, result)
		case productionKindUnit:
			capacityKey := productionCapacityKey{regionID: region.ID, lane: g.productionCapacityLane(order.TypeID)}
			capacity := g.regionUnitProductionCapacity(region, order.TypeID)
			if progressedUnitsByLane[capacityKey] >= capacity {
				remaining = append(remaining, order)
				continue
			}
			if region.IsLocked {
				remaining = append(remaining, order)
				if order.TurnsLeft <= 1 {
					result.delayed = true
					result.reason = "bölge kilitli"
					results = append(results, result)
				}
				continue
			}
			progressedUnitsByLane[capacityKey]++
			order.TurnsLeft--
			if order.TurnsLeft > 0 {
				remaining = append(remaining, order)
				continue
			}
			if reason := g.completeUnit(region, faction.FactionID(order.FactionID), order.TypeID); reason != "" {
				order.TurnsLeft = 1
				remaining = append(remaining, order)
				result.delayed = true
				result.reason = reason
			}
			results = append(results, result)
		default:
			result.canceled = true
			result.reason = "bilinmeyen üretim türü"
			results = append(results, result)
		}
	}

	for i := len(remaining); i < len(queue); i++ {
		queue[i] = state.ProductionOrder{}
	}
	g.gs.ProductionQueue = remaining
	return results
}

func (g *Game) completeBuilding(region *world.Region, buildingID string) bool {
	b, ok := g.gs.BuildingTypes[buildingID]
	if !ok {
		return false
	}
	count := 0
	for _, bid := range region.Buildings {
		if bid == buildingID {
			count++
		}
	}
	if count >= b.MaxPerRegion {
		return false
	}
	region.Buildings = append(region.Buildings, buildingID)
	if buildingID == "port" && g.ensurePortSettlement(region) && g.renderer != nil {
		g.renderer.RebuildSettlementAnchors()
		g.renderer.MarkMapDirty()
	}
	return true
}

func (g *Game) ensurePortSettlement(region *world.Region) bool {
	if region == nil || region.IsSea {
		return false
	}
	for _, settlement := range region.Settlements {
		if settlement.Type == world.SettlementPort {
			return false
		}
	}
	x, y := region.WorldX, region.WorldY
	if px, py, ok := autoPortSettlementPoint(region, g.gs.Regions); ok {
		x, y = px, py
	} else {
		for _, nid := range region.Neighbors {
			sea := g.gs.Regions[nid]
			if sea == nil || !sea.IsSea {
				continue
			}
			x = (region.WorldX*2 + sea.WorldX) / 3
			y = (region.WorldY*2 + sea.WorldY) / 3
			break
		}
	}
	region.Settlements = append(region.Settlements, world.Settlement{
		ID:     nextPortSettlementID(region),
		Name:   "Port",
		NameTR: "Liman",
		X:      x,
		Y:      y,
		Type:   world.SettlementPort,
	})
	region.RecalculatePopulation()
	return true
}

func autoPortSettlementPoint(region *world.Region, regions map[world.RegionID]*world.Region) (int, int, bool) {
	if region == nil || len(region.Shape) == 0 {
		return 0, 0, false
	}
	centerX := float64(region.WorldX)
	centerY := float64(region.WorldY)
	seaX, seaY, ok := nearestSeaNeighborCenter(region, regions)
	if !ok {
		return 0, 0, false
	}
	coastX, coastY, ok := firstCoastIntersection(region.Shape, centerX, centerY, seaX, seaY)
	if !ok {
		return 0, 0, false
	}
	dx := centerX - coastX
	dy := centerY - coastY
	dist := math.Hypot(dx, dy)
	if dist > 0 {
		inset := autoPortSettlementInset
		if inset > dist {
			inset = dist / 2
		}
		coastX += dx / dist * inset
		coastY += dy / dist * inset
	}
	return int(math.Round(coastX)), int(math.Round(coastY)), true
}

func nearestSeaNeighborCenter(region *world.Region, regions map[world.RegionID]*world.Region) (float64, float64, bool) {
	if region == nil {
		return 0, 0, false
	}
	centerX := float64(region.WorldX)
	centerY := float64(region.WorldY)
	bestDist := math.MaxFloat64
	bestX := 0.0
	bestY := 0.0
	found := false
	for _, nid := range region.Neighbors {
		sea := regions[nid]
		if sea == nil || !sea.IsSea {
			continue
		}
		dx := float64(sea.WorldX) - centerX
		dy := float64(sea.WorldY) - centerY
		dist := dx*dx + dy*dy
		if dist >= bestDist {
			continue
		}
		bestDist = dist
		bestX = float64(sea.WorldX)
		bestY = float64(sea.WorldY)
		found = true
	}
	return bestX, bestY, found
}

func firstCoastIntersection(shape [][][2]float32, centerX, centerY, seaX, seaY float64) (float64, float64, bool) {
	bestT := math.MaxFloat64
	bestX := 0.0
	bestY := 0.0
	found := false
	for _, ring := range shape {
		if len(ring) < 2 {
			continue
		}
		for i := 0; i < len(ring); i++ {
			a := ring[i]
			b := ring[(i+1)%len(ring)]
			ix, iy, t, ok := raySegmentIntersection(
				centerX, centerY,
				seaX-centerX, seaY-centerY,
				float64(a[0]), float64(a[1]),
				float64(b[0])-float64(a[0]), float64(b[1])-float64(a[1]),
			)
			if !ok || t < 0 || t > 1 || t >= bestT {
				continue
			}
			bestT = t
			bestX = ix
			bestY = iy
			found = true
		}
	}
	return bestX, bestY, found
}

func raySegmentIntersection(px, py, rx, ry, qx, qy, sx, sy float64) (float64, float64, float64, bool) {
	const eps = 1e-6
	denom := rx*sy - ry*sx
	if math.Abs(denom) < eps {
		return 0, 0, 0, false
	}
	qpx := qx - px
	qpy := qy - py
	t := (qpx*sy - qpy*sx) / denom
	u := (qpx*ry - qpy*rx) / denom
	if t < -eps || t > 1+eps || u < -eps || u > 1+eps {
		return 0, 0, 0, false
	}
	return px + t*rx, py + t*ry, t, true
}

func nextPortSettlementID(region *world.Region) string {
	if region == nil {
		return "port"
	}
	base := string(region.ID) + "_port"
	for n := 1; ; n++ {
		candidate := base
		if n > 1 {
			candidate = fmt.Sprintf("%s_%d", base, n)
		}
		exists := false
		for _, settlement := range region.Settlements {
			if settlement.ID == candidate {
				exists = true
				break
			}
		}
		if !exists {
			return candidate
		}
	}
}

func (g *Game) completeUnit(region *world.Region, ownerID faction.FactionID, unitTypeID string) string {
	utype, ok := g.gs.UnitTypes[unitTypeID]
	if !ok {
		return "birim tanımı bulunamadı"
	}
	if utype.RequiredBldg == "port" {
		return g.completeNavalUnit(region, ownerID, unitTypeID)
	}
	return g.completeLandUnit(region, ownerID, unitTypeID)
}

func (g *Game) completeNavalUnit(region *world.Region, ownerID faction.FactionID, unitTypeID string) string {
	unitType := g.gs.UnitTypes[unitTypeID]
	var seaRegion world.RegionID
	for _, nid := range region.Neighbors {
		if n, ok := g.gs.Regions[nid]; ok && n.IsSea {
			seaRegion = nid
			break
		}
	}
	if seaRegion == "" {
		return "komşu deniz bölgesi bulunamadı"
	}
	portSettlementID := preferredDockSettlementID(region)
	var fleet *army.Army
	for _, a := range g.gs.Armies {
		if !a.IsDocked() || a.DockedRegionID != region.ID ||
			a.OwnerID != string(ownerID) ||
			(portSettlementID != "" && a.DockedSettlementID != portSettlementID) ||
			!navalFleetAcceptsCompletedUnit(a, unitType, g.gs.UnitTypes) {
			continue
		}
		fleet = a
		break
	}
	if fleet != nil {
		if len(fleet.Units) >= army.MaxArmySize {
			return "filo dolu"
		}
		fleet.Units = append(fleet.Units, army.Unit{TypeID: unitTypeID, CurrentHP: 100})
		return ""
	}
	g.gs.NextArmySeq++
	newID := army.ArmyID(fmt.Sprintf("fleet_%s_%d", string(ownerID), g.gs.NextArmySeq))
	g.gs.Armies[newID] = &army.Army{
		ID:                 newID,
		OwnerID:            string(ownerID),
		RegionID:           seaRegion,
		DockedRegionID:     region.ID,
		DockedSettlementID: preferredDockSettlementID(region),
		Units:              []army.Unit{{TypeID: unitTypeID, CurrentHP: 100}},
		MovePoints:         3,
		MaxMovePoints:      3,
		IsNaval:            true,
	}
	return ""
}

// navalFleetAcceptsCompletedUnit merchant görev filolarını taşıma/savaş
// filolarından ayrı tutar. Böylece rota başına iki gemilik ekonomik görev,
// başka bir deniz görevi yüzünden aynı stack içinde sürüklenmez.
func navalFleetAcceptsCompletedUnit(fleet *army.Army, completed *army.UnitType, unitTypes map[string]*army.UnitType) bool {
	if fleet == nil || completed == nil || len(fleet.Units) >= army.MaxArmySize {
		return false
	}
	completedMerchant := completed.Category == army.CategoryNavalTrade
	merchantCount := 0
	nonMerchantCount := 0
	for _, unit := range fleet.Units {
		unitType := unitTypes[unit.TypeID]
		if unit.TypeID == "merchant_ship" || unitType != nil && unitType.Category == army.CategoryNavalTrade {
			merchantCount++
		} else {
			nonMerchantCount++
		}
	}
	if completedMerchant {
		return nonMerchantCount == 0 && merchantCount < 2
	}
	return merchantCount == 0 && fleet.TradeRouteKey == ""
}

func preferredDockSettlementID(region *world.Region) string {
	if region == nil {
		return ""
	}
	for _, settlement := range region.Settlements {
		if settlement.Type == world.SettlementPort {
			return settlement.ID
		}
	}
	if len(region.Settlements) > 0 {
		return region.Settlements[0].ID
	}
	return ""
}

func (g *Game) completeLandUnit(region *world.Region, ownerID faction.FactionID, unitTypeID string) string {
	targetArmy, canCreateNew := g.findRecruitableLandArmy(region.ID, ownerID)
	if targetArmy != nil {
		targetArmy.Units = append(targetArmy.Units, army.Unit{TypeID: unitTypeID, CurrentHP: 100})
		return ""
	}
	if !canCreateNew {
		return "maksimum ordu sayısına ulaşıldı"
	}
	g.gs.NextArmySeq++
	newID := army.ArmyID(fmt.Sprintf("army_%s_%d", string(ownerID), g.gs.NextArmySeq))
	g.gs.Armies[newID] = &army.Army{
		ID:            newID,
		OwnerID:       string(ownerID),
		RegionID:      region.ID,
		Units:         []army.Unit{{TypeID: unitTypeID, CurrentHP: 100}},
		MovePoints:    2,
		MaxMovePoints: 2,
	}
	return ""
}

func (g *Game) findRecruitableLandArmy(regionID world.RegionID, ownerID faction.FactionID) (*army.Army, bool) {
	for _, a := range g.gs.Armies {
		if a.RegionID != regionID || a.OwnerID != string(ownerID) || a.IsNaval || a.IsGarrison {
			continue
		}
		if len(a.Units) < army.MaxArmySize {
			return a, true
		}
	}
	return nil, g.gs.CurrentLandArmies(ownerID) < g.gs.MaxLandArmies(ownerID)
}

func (g *Game) queuedBuildingCount(rid world.RegionID, buildingID string) int {
	count := 0
	for _, order := range g.gs.ProductionQueue {
		if order.Kind == productionKindBuilding && order.RegionID == rid && order.TypeID == buildingID {
			count++
		}
	}
	return count
}

func (g *Game) pendingLandUnitCount(fid faction.FactionID) int {
	count := 0
	for _, order := range g.gs.ProductionQueue {
		if order.Kind != productionKindUnit || order.FactionID != string(fid) {
			continue
		}
		if utype, ok := g.gs.UnitTypes[order.TypeID]; ok && utype.RequiredBldg != "port" {
			count++
		}
	}
	return count
}

func (g *Game) pendingUnitCountByRegion(rid world.RegionID, fid faction.FactionID) int {
	count := 0
	for _, order := range g.gs.ProductionQueue {
		if order.Kind == productionKindUnit && order.RegionID == rid && order.FactionID == string(fid) {
			count++
		}
	}
	return count
}

func (g *Game) pendingNavalUnitCount(seaRegion world.RegionID, fid faction.FactionID) int {
	count := 0
	for _, order := range g.gs.ProductionQueue {
		if order.Kind != productionKindUnit || order.FactionID != string(fid) {
			continue
		}
		utype, ok := g.gs.UnitTypes[order.TypeID]
		if !ok || utype.RequiredBldg != "port" {
			continue
		}
		region, ok := g.gs.Regions[order.RegionID]
		if !ok {
			continue
		}
		for _, nid := range region.Neighbors {
			if nid == seaRegion {
				count++
				break
			}
		}
	}
	return count
}

func (g *Game) productionName(result productionResult) string {
	switch result.kind {
	case productionKindBuilding:
		if b, ok := g.gs.BuildingTypes[result.typeID]; ok {
			return b.NameTR
		}
	case productionKindUnit:
		if u, ok := g.gs.UnitTypes[result.typeID]; ok {
			return u.NameTR
		}
	}
	return result.typeID
}
