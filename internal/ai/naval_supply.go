package ai

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// aiNavalSupplyMission AI'nin bir kıyı ordusuna tahıl götürmek için bu tur
// kullandığı geçici plandır. Filo üzerindeki NavalMission ise filonun hedef
// denize vardığında state katmanında doğrulanıp kalıcılaşan görev kaydıdır.
type aiNavalSupplyMission struct {
	FleetID         army.ArmyID
	TargetArmyID    army.ArmyID
	TargetRegionID  world.RegionID
	TargetSeaID     world.RegionID
	CapitalRegionID world.RegionID
}

func buildAINavalSupplyMission(ctx *StrategicContext) *aiNavalSupplyMission {
	if ctx == nil || ctx.gs == nil || ctx.FactionID == "" {
		return nil
	}
	gs := ctx.gs

	// Daha önce bağlanmış ve kargo taşıyan filo önceliklidir. Yeni stratejik
	// planın saldırı görevi bu filonun ikmal görevini ezmemelidir.
	for _, fleet := range aiSortedArmies(gs) {
		if fleet == nil || fleet.OwnerID != string(ctx.FactionID) || !fleet.IsNaval || fleet.NavalMission == nil || fleet.NavalMission.Kind != army.NavalMissionSupplyArmy || fleet.SupplyCargo.Grain <= 0 {
			continue
		}
		target := gs.Armies[fleet.NavalMission.TargetArmyID]
		if !aiNavalSupplyTargetValid(gs, target, ctx.FactionID) {
			continue
		}
		return aiMakeNavalSupplyMission(gs, fleet, target)
	}

	var best *aiNavalSupplyMission
	bestScore := -1
	for _, target := range aiSortedArmies(gs) {
		if !aiNavalSupplyTargetValid(gs, target, ctx.FactionID) {
			continue
		}
		_, _, overload := aiRegionLogistics(gs, gs.Regions[target.RegionID], target.OwnerID)
		if overload <= 0 {
			continue
		}
		for _, fleet := range aiSortedArmies(gs) {
			if !aiNavalSupplyFleetAvailable(gs, fleet, ctx.FactionID) {
				continue
			}
			seaID, distance := aiBestNavalSupplySea(gs, fleet.RegionID, gs.Regions[target.RegionID])
			if seaID == "" || distance < 0 {
				continue
			}
			score := overload*100 - distance*5
			if fleet.SupplyCargo.Grain > 0 {
				score += 25
			}
			if best == nil || score > bestScore || (score == bestScore && (target.ID < best.TargetArmyID || (target.ID == best.TargetArmyID && fleet.ID < best.FleetID))) {
				bestScore = score
				best = &aiNavalSupplyMission{
					FleetID:        fleet.ID,
					TargetArmyID:   target.ID,
					TargetRegionID: target.RegionID,
					TargetSeaID:    seaID,
				}
				if capital, _, _, ok := gs.FactionCapital(ctx.FactionID); ok && capital != nil {
					best.CapitalRegionID = capital.ID
				}
			}
		}
	}
	if best != nil {
		return best
	}

	// Uygun mevcut filo yoksa hedef yine de seçilir; hazırlık safhası başkent
	// limanında nakliye üretimini kuyruğa alır. Böylece AI yalnızca tesadüfen
	// hazır bulunan filoları kullanmaya mahkûm kalmaz.
	capital, _, _, ok := gs.FactionCapital(ctx.FactionID)
	if !ok || capital == nil {
		return nil
	}
	capitalSea := aiSeaNeighbor(gs, capital)
	if capitalSea == "" {
		return nil
	}
	var targetCandidate *army.Army
	bestOverload := 0
	for _, target := range aiSortedArmies(gs) {
		if !aiNavalSupplyTargetValid(gs, target, ctx.FactionID) {
			continue
		}
		_, _, overload := aiRegionLogistics(gs, gs.Regions[target.RegionID], target.OwnerID)
		if overload > bestOverload {
			bestOverload = overload
			targetCandidate = target
		}
	}
	if targetCandidate == nil {
		return nil
	}
	targetSea, distance := aiBestNavalSupplySea(gs, capitalSea, gs.Regions[targetCandidate.RegionID])
	if targetSea == "" || distance < 0 {
		return nil
	}
	return &aiNavalSupplyMission{
		TargetArmyID:    targetCandidate.ID,
		TargetRegionID:  targetCandidate.RegionID,
		TargetSeaID:     targetSea,
		CapitalRegionID: capital.ID,
	}
}

func aiMakeNavalSupplyMission(gs *state.GameState, fleet, target *army.Army) *aiNavalSupplyMission {
	mission := &aiNavalSupplyMission{FleetID: fleet.ID, TargetArmyID: target.ID, TargetRegionID: target.RegionID}
	if targetRegion := gs.Regions[target.RegionID]; targetRegion != nil {
		mission.TargetSeaID, _ = aiBestNavalSupplySea(gs, fleet.RegionID, targetRegion)
	}
	if capital, _, _, ok := gs.FactionCapital(faction.FactionID(fleet.OwnerID)); ok && capital != nil {
		mission.CapitalRegionID = capital.ID
	}
	return mission
}

func aiNavalSupplyTargetValid(gs *state.GameState, target *army.Army, fid faction.FactionID) bool {
	if gs == nil || target == nil || target.OwnerID != string(fid) || target.IsNaval || len(target.Units) == 0 {
		return false
	}
	region := gs.Regions[target.RegionID]
	return region != nil && !region.IsSea && region.IsCoastal(gs.Regions)
}

func aiNavalSupplyFleetAvailable(gs *state.GameState, fleet *army.Army, fid faction.FactionID) bool {
	return gs != nil && fleet != nil && fleet.OwnerID == string(fid) && fleet.IsNaval && fleet.TransportCapacity(gs.UnitTypes) > 0 && len(fleet.EmbarkedUnits) == 0 && fleet.TradeRouteKey == "" && (fleet.NavalMission == nil || fleet.NavalMission.Kind == army.NavalMissionSupplyArmy)
}

func aiBestNavalSupplySea(gs *state.GameState, start world.RegionID, target *world.Region) (world.RegionID, int) {
	if gs == nil || target == nil || start == "" {
		return "", -1
	}
	bestID := world.RegionID("")
	bestDistance := -1
	for _, seaID := range aiSortedNeighborIDs(target) {
		sea := gs.Regions[seaID]
		if sea == nil || !sea.IsSea || sea.IsLocked {
			continue
		}
		distance := aiSeaRouteDistance(gs, start, seaID)
		if distance < 0 || bestDistance < 0 || distance < bestDistance || (distance == bestDistance && seaID < bestID) {
			bestID = seaID
			bestDistance = distance
		}
	}
	return bestID, bestDistance
}

// aiPrepareNavalSupplyMission kargoyu yalnız başkent limanında yükler. Filo
// henüz yoksa aynı hedef için bir nakliye üretimini kuyruğa alır; böylece AI
// mevcut gemileri kullanabildiği gibi ihtiyaç halinde yeni gemi de hazırlar.
func aiPrepareNavalSupplyMission(gs *state.GameState, fid faction.FactionID, budget *aiBudget, ctx *StrategicContext, steps *[]TurnStep) {
	if gs == nil || ctx == nil || ctx.navalSupplyMission == nil || ctx.FactionID != fid {
		return
	}
	mission := ctx.navalSupplyMission
	transportType := gs.UnitTypes["transport"]
	capital := gs.Regions[mission.CapitalRegionID]
	self := gs.Factions[fid]
	if transportType == nil || capital == nil || self == nil || self.IsEliminated {
		return
	}

	fleet := gs.Armies[mission.FleetID]
	if fleet == nil {
		if !aiCanQueueNavalUnit(gs, fid) || !aiRegionHasPortBuilding(capital) || !transportType.HasAllRequiredTechs(self.Research.Completed) || aiPendingTransportAtRegion(gs, fid, capital.ID) > 0 || !aiCanApplySupplyTransportCost(self, transportType, budget) {
			return
		}
		if !aiApplyUnitCostForBudget(self, transportType, budget, aiBudgetNaval) {
			return
		}
		aiEnqueueProduction(gs, fid, aiProductionKindUnit, capital.ID, transportType.ID, transportType.TurnsRequired)
		addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepRecruit, TargetRegion: capital.ID, FocusRegion: mission.TargetRegionID, Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, capital.ID) + " limanında kıyı ordusu için nakliye gemisi hazırlıyor."})
		return
	}

	if fleet.SupplyCargo.Grain > 0 || fleet.DockedRegionID != capital.ID {
		return
	}
	capacity := gs.SupplyCargoCapacityForTurns(fleet.ID, 5) - gs.SupplyCargoLoadedAmount(fleet.ID)
	reserve := aiMinGrainReserveForSupply(gs, fid)
	amount := minInt(capacity, self.Grain-reserve)
	if amount <= 0 {
		return
	}
	if ok, _ := gs.LoadSupplyCargoAtCapital(fleet.ID, economy.ResourceCost{Grain: amount}); !ok {
		return
	}
	addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepInfo, ArmyID: fleet.ID, TargetRegion: mission.TargetRegionID, FocusRegion: mission.TargetRegionID, Message: turnFactionName(gs, fid) + " kıyı ordusuna deniz ikmali yüklüyor."})
}

func aiPendingTransportAtRegion(gs *state.GameState, fid faction.FactionID, regionID world.RegionID) int {
	count := 0
	for _, order := range gs.ProductionQueue {
		if order.Kind == aiProductionKindUnit && order.FactionID == string(fid) && order.RegionID == regionID && order.TypeID == "transport" {
			count++
		}
	}
	return count
}

func aiCanApplySupplyTransportCost(self *faction.Faction, transportType *army.UnitType, budget *aiBudget) bool {
	return self != nil && transportType != nil && aiCanAffordForBudget(self, economy.ResourceCost{Gold: transportType.GoldCost, Grain: transportType.GrainCost, Iron: transportType.IronCost, Timber: transportType.TimberCost, Stone: transportType.StoneCost, Spice: transportType.SpiceCost, Cloth: transportType.ClothCost}, budget, aiBudgetNaval)
}

func aiMinGrainReserveForSupply(gs *state.GameState, fid faction.FactionID) int {
	reserve := aiMinGoldReserve / 4
	if reserve < 20 {
		reserve = 20
	}
	if status, ok := gs.GrainEconomy[fid]; ok && status.TotalDemand > reserve {
		reserve = status.TotalDemand
	}
	return reserve
}

// aiNavalSupplyMissionMove, kargo taşıyan filoyu hedef kıyıya götürür ve
// hedef denize vardığında ortak state doğrulamasından geçen ikmal görevini
// atar. Kargosuz filo önce başkent limanına döner.
func aiNavalSupplyMissionMove(gs *state.GameState, fleet *army.Army, ctx *StrategicContext) (world.RegionID, bool) {
	if gs == nil || fleet == nil || ctx == nil || ctx.navalSupplyMission == nil || ctx.navalSupplyMission.FleetID != fleet.ID || fleet.OwnerID != string(ctx.FactionID) {
		return "", false
	}
	mission := ctx.navalSupplyMission
	target := gs.Armies[mission.TargetArmyID]
	if !aiNavalSupplyTargetValid(gs, target, ctx.FactionID) {
		if fleet.NavalMission != nil && fleet.NavalMission.Kind == army.NavalMissionSupplyArmy {
			fleet.NavalMission = nil
		}
		return "", true
	}
	mission.TargetRegionID = target.RegionID
	mission.TargetSeaID, _ = aiBestNavalSupplySea(gs, fleet.RegionID, gs.Regions[target.RegionID])
	if mission.TargetSeaID == "" {
		return "", true
	}

	if fleet.NavalMission != nil && fleet.NavalMission.Kind == army.NavalMissionSupplyArmy {
		if fleet.SupplyCargo.Grain <= 0 {
			fleet.NavalMission = nil
			return "", true
		}
		if seaNeighborOfSupplyTarget(gs, target, fleet.RegionID) {
			return "", true
		}
		return aiThreatAwareSeaNextStep(gs, ctx.FactionID, fleet.RegionID, mission.TargetSeaID), true
	}

	if fleet.SupplyCargo.Grain <= 0 {
		capital := gs.Regions[mission.CapitalRegionID]
		if capital == nil {
			return "", true
		}
		capitalSea := aiSeaNeighbor(gs, capital)
		if capitalSea == "" {
			return "", true
		}
		if fleet.IsDocked() {
			if fleet.DockedRegionID == capital.ID {
				return fleet.RegionID, true
			}
			return fleet.RegionID, true
		}
		if fleet.RegionID == capitalSea {
			return capital.ID, true
		}
		next := aiThreatAwareSeaNextStep(gs, ctx.FactionID, fleet.RegionID, capitalSea)
		return next, true
	}

	if fleet.IsDocked() {
		return fleet.RegionID, true
	}
	if seaNeighborOfSupplyTarget(gs, target, fleet.RegionID) {
		missionState := army.NavalMission{Kind: army.NavalMissionSupplyArmy, TargetArmyID: target.ID}
		if ok, _ := gs.AssignNavalMission(fleet.ID, missionState); !ok {
			return "", true
		}
		return "", true
	}
	next := aiThreatAwareSeaNextStep(gs, ctx.FactionID, fleet.RegionID, mission.TargetSeaID)
	return next, true
}

func seaNeighborOfSupplyTarget(gs *state.GameState, target *army.Army, seaID world.RegionID) bool {
	if gs == nil || target == nil {
		return false
	}
	region := gs.Regions[target.RegionID]
	if region == nil || region.IsSea {
		return false
	}
	for _, neighborID := range region.Neighbors {
		if neighborID == seaID && gs.Regions[neighborID] != nil && gs.Regions[neighborID].IsSea {
			return true
		}
	}
	return false
}
