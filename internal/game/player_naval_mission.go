package game

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// executePlayerNavalMissions, oyuncunun kalıcı hareketli filo görevlerini yeni
// turda hareket havuzu yenilendikten sonra işler. Devriye ve abluka mevcut
// denizde sabit kaldığı için yalnız escort ve nakliye burada rota izler.
func (g *Game) executePlayerNavalMissions() {
	if g == nil || g.gs == nil || len(g.gs.Armies) == 0 {
		return
	}
	ids := make([]army.ArmyID, 0, len(g.gs.Armies))
	for id, fleet := range g.gs.Armies {
		if fleet != nil && fleet.OwnerID == string(g.gs.PlayerFactionID) && fleet.IsNaval && fleet.NavalMission != nil {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	for _, fleetID := range ids {
		for {
			fleet := g.gs.Armies[fleetID]
			if fleet == nil || fleet.NavalMission == nil || fleet.MovePoints <= 0 {
				break
			}
			if fleet.NavalMission.Kind == army.NavalMissionTransport && len(fleet.EmbarkedUnits) == 0 {
				fleet.NavalMission = nil
				break
			}

			target, ok := playerNavalMissionNextTarget(g.gs, fleet)
			if !ok {
				// Escort edilen filo kaldırıldıysa eski görevi sessizce temizle;
				// oyuncu panelini geçersiz bir hedefle bırakma.
				if fleet.NavalMission.Kind == army.NavalMissionEscort {
					fleet.NavalMission = nil
				}
				break
			}
			if target == fleet.RegionID && !fleet.IsDocked() {
				break
			}

			beforeRegion := fleet.RegionID
			beforeMovePoints := fleet.MovePoints
			if fleet.NavalMission.Kind == army.NavalMissionTransport {
				if targetRegion := g.gs.Regions[target]; targetRegion != nil && !targetRegion.IsSea {
					g.resolveFleetDisembarkWithStance(fleet, target, targetRegion, combat.BattleStanceBalanced)
				} else {
					g.moveArmyWithStance(fleetID, target, combat.BattleStanceBalanced)
				}
			} else {
				g.moveArmyWithStance(fleetID, target, combat.BattleStanceBalanced)
			}
			if g.gs.PendingNavalContact != nil {
				break
			}

			fleet = g.gs.Armies[fleetID]
			if fleet == nil {
				break
			}
			if fleet.RegionID == beforeRegion && fleet.MovePoints >= beforeMovePoints {
				break
			}
			if g.gs.ArmyMoveUsage == nil {
				g.gs.ArmyMoveUsage = make(map[army.ArmyID]bool)
			}
			g.gs.ArmyMoveUsage[fleetID] = true
			if fleet.NavalMission != nil && fleet.NavalMission.Kind == army.NavalMissionTransport && len(fleet.EmbarkedUnits) == 0 {
				fleet.NavalMission = nil
				break
			}
		}
	}
}

// followEscortingFleets, oyuncu tarafından hareket ettirilen açık deniz
// filosunu koruyan filoları aynı denize taşır. Escort kendi normal hareket
// doğrulamasından geçtiği için komşuluk, temas/savaş ve hareket puanı kuralları
// burada da aynen uygulanır; hareket puanı yoksa escort yerinde kalır.
func (g *Game) followEscortingFleets(targetFleetID army.ArmyID, previousLocation string) {
	if g == nil || g.gs == nil || targetFleetID == "" || previousLocation == "" || g.gs.PendingNavalContact != nil {
		return
	}
	targetFleet := g.gs.Armies[targetFleetID]
	if targetFleet == nil || !targetFleet.IsAtSea() || targetFleet.LocationID() == previousLocation {
		return
	}

	escortIDs := make([]army.ArmyID, 0)
	for fleetID, fleet := range g.gs.Armies {
		if fleet == nil || fleet.ID == targetFleetID || fleet.OwnerID != targetFleet.OwnerID || !fleet.IsAtSea() || fleet.MovePoints <= 0 || fleet.NavalMission == nil {
			continue
		}
		if fleet.NavalMission.Kind == army.NavalMissionEscort && fleet.NavalMission.TargetFleetID == targetFleetID && fleet.RegionID != targetFleet.RegionID {
			escortIDs = append(escortIDs, fleetID)
		}
	}
	sort.Slice(escortIDs, func(i, j int) bool { return escortIDs[i] < escortIDs[j] })

	for _, escortID := range escortIDs {
		if escort := g.gs.Armies[escortID]; escort == nil || escort.MovePoints <= 0 || escort.RegionID == targetFleet.RegionID {
			continue
		}
		g.escortFollowDepth++
		g.moveArmyWithStance(escortID, targetFleet.RegionID, combat.BattleStanceBalanced)
		g.escortFollowDepth--
	}
}

// playerNavalMissionNextTarget, yalnız hareket edebilen görevler için bir
// sonraki komşu deniz/kıyı adımını döndürür. Devriye ve abluka mevcut denizde
// sabittir; escort hedef filosunu, nakliye ise hedef kıyının deniz komşusunu
// takip eder.
func playerNavalMissionNextTarget(gs *state.GameState, fleet *army.Army) (world.RegionID, bool) {
	if gs == nil || fleet == nil || fleet.NavalMission == nil {
		return "", false
	}
	switch fleet.NavalMission.Kind {
	case army.NavalMissionPatrol, army.NavalMissionBlockade:
		if !fleet.IsAtSea() || fleet.RegionID == "" {
			return "", false
		}
		return fleet.RegionID, true
	case army.NavalMissionEscort:
		targetFleet := gs.Armies[fleet.NavalMission.TargetFleetID]
		if targetFleet == nil || !targetFleet.IsNaval || targetFleet.OwnerID != fleet.OwnerID {
			return "", false
		}
		return playerSeaRouteNext(gs, fleet.RegionID, targetFleet.RegionID)
	case army.NavalMissionTransport:
		targetLand := gs.Regions[fleet.NavalMission.TargetRegionID]
		if targetLand == nil || targetLand.IsSea || !targetLand.CanLandEnter() {
			return "", false
		}
		for _, neighborID := range sortedRegionNeighbors(targetLand) {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil || !neighbor.IsSea {
				continue
			}
			if fleet.RegionID == neighborID {
				return targetLand.ID, true
			}
			if next, reachable := playerSeaRouteNext(gs, fleet.RegionID, neighborID); reachable {
				return next, true
			}
		}
		return "", false
	default:
		return "", false
	}
}

func playerSeaRouteNext(gs *state.GameState, start, target world.RegionID) (world.RegionID, bool) {
	if gs == nil || start == "" || target == "" {
		return "", false
	}
	if start == target {
		return target, true
	}
	startRegion := gs.Regions[start]
	targetRegion := gs.Regions[target]
	if startRegion == nil || targetRegion == nil || !startRegion.IsSea || !targetRegion.IsSea || targetRegion.IsLocked {
		return "", false
	}
	parents := map[world.RegionID]world.RegionID{start: ""}
	queue := []world.RegionID{start}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, neighborID := range sortedRegionNeighbors(gs.Regions[current]) {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil || !neighbor.IsSea || neighbor.IsLocked {
				continue
			}
			if _, seen := parents[neighborID]; seen {
				continue
			}
			parents[neighborID] = current
			if neighborID == target {
				step := target
				for parents[step] != start {
					step = parents[step]
				}
				return step, true
			}
			queue = append(queue, neighborID)
		}
	}
	return "", false
}

func sortedRegionNeighbors(region *world.Region) []world.RegionID {
	if region == nil || len(region.Neighbors) == 0 {
		return nil
	}
	neighbors := append([]world.RegionID(nil), region.Neighbors...)
	sort.Slice(neighbors, func(i, j int) bool { return neighbors[i] < neighbors[j] })
	return neighbors
}
