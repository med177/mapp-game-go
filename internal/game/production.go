package game

import (
	"fmt"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	productionKindBuilding = "building"
	productionKindUnit     = "unit"
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

func (g *Game) applyProductionTicks() []productionResult {
	queue := g.gs.ProductionQueue
	remaining := queue[:0]
	var results []productionResult

	for _, order := range queue {
		order.TurnsLeft--
		if order.TurnsLeft > 0 {
			remaining = append(remaining, order)
			continue
		}

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

		switch order.Kind {
		case productionKindBuilding:
			if !g.completeBuilding(region, order.TypeID) {
				result.canceled = true
				result.reason = "bina zaten tamamlanmış"
			}
			results = append(results, result)
		case productionKindUnit:
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
	return true
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
	var fleet *army.Army
	for _, a := range g.gs.Armies {
		if a.RegionID == seaRegion && a.OwnerID == string(ownerID) && a.IsNaval {
			fleet = a
			break
		}
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
		ID:            newID,
		OwnerID:       string(ownerID),
		RegionID:      seaRegion,
		Units:         []army.Unit{{TypeID: unitTypeID, CurrentHP: 100}},
		MovePoints:    3,
		MaxMovePoints: 3,
		IsNaval:       true,
	}
	return ""
}

func (g *Game) completeLandUnit(region *world.Region, ownerID faction.FactionID, unitTypeID string) string {
	var targetArmy *army.Army
	for _, a := range g.gs.Armies {
		if a.RegionID == region.ID && a.OwnerID == string(ownerID) && !a.IsNaval {
			targetArmy = a
			break
		}
	}
	if targetArmy != nil {
		if len(targetArmy.Units) >= army.MaxArmySize {
			return "ordu dolu"
		}
		targetArmy.Units = append(targetArmy.Units, army.Unit{TypeID: unitTypeID, CurrentHP: 100})
		return ""
	}
	if g.gs.CurrentLandArmies(ownerID) >= g.gs.MaxLandArmies(ownerID) {
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
