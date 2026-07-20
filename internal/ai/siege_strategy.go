package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiEnsureSiegeMap(gs *state.GameState) {
	if gs != nil && gs.Sieges == nil {
		gs.Sieges = make(map[world.RegionID]*state.SiegeState)
	}
}

func aiClearSiege(gs *state.GameState, regionID world.RegionID) {
	if gs == nil || gs.Sieges == nil || regionID == "" {
		return
	}
	delete(gs.Sieges, regionID)
}

func aiClearSiegesByArmy(gs *state.GameState, armyID army.ArmyID) {
	if gs == nil || gs.Sieges == nil || armyID == "" {
		return
	}
	for rid, siege := range gs.Sieges {
		if siege != nil && siege.AttackerArmyID == armyID {
			delete(gs.Sieges, rid)
		}
	}
}

func aiSiegeDefenseBonus(fortLevel, breachLevel int) float64 {
	if fortLevel <= 0 {
		return 0
	}
	base := float64(fortLevel) * 0.14
	switch breachLevel {
	case 2:
		return base * 0.25
	case 1:
		return base * 0.55
	default:
		return base + 0.18
	}
}

func aiVirtualSiegeGarrison(gs *state.GameState, target *world.Region) *army.Army {
	if gs == nil || target == nil {
		return nil
	}
	unitTypeID := aiMilitiaID
	if _, ok := gs.UnitTypes[unitTypeID]; !ok {
		unitTypeIDs := make([]string, 0, len(gs.UnitTypes))
		for id := range gs.UnitTypes {
			unitTypeIDs = append(unitTypeIDs, id)
		}
		sort.Strings(unitTypeIDs)
		for _, id := range unitTypeIDs {
			ut := gs.UnitTypes[id]
			if ut != nil && ut.Category == army.CategoryInfantry {
				unitTypeID = id
				break
			}
		}
	}
	unitCount := 1 + target.FortificationLevel()
	if unitCount > 6 {
		unitCount = 6
	}
	units := make([]army.Unit, 0, unitCount)
	for i := 0; i < unitCount; i++ {
		units = append(units, army.Unit{TypeID: unitTypeID, CurrentHP: army.MaxUnitHP})
	}
	return &army.Army{OwnerID: target.OwnerID, RegionID: target.ID, Units: units, MovePoints: 0}
}

func aiCanStartSiege(gs *state.GameState, a *army.Army, target *world.Region) bool {
	if gs == nil || a == nil || target == nil || a.IsNaval || !target.CanLandEnter() {
		return false
	}
	if target.OwnerID == "" || target.OwnerID == a.OwnerID || !target.IsFortified() {
		return false
	}
	_, stance := relationScore(gs, a.OwnerID, target.OwnerID)
	if stance != faction.StanceWar {
		return false
	}
	if active := gs.SiegeByArmy(a.ID); active != nil && active.RegionID != target.ID {
		return false
	}
	if siege := gs.SiegeAt(target.ID); siege != nil && siege.AttackerArmyID != a.ID {
		return false
	}
	return true
}

func aiStartSiege(gs *state.GameState, a *army.Army, target *world.Region, defender *army.Army) {
	if gs == nil || a == nil || target == nil {
		return
	}
	aiEnsureSiegeMap(gs)
	siege := &state.SiegeState{RegionID: target.ID, AttackerArmyID: a.ID, AttackerHomeRegionID: a.RegionID, AttackerFactionID: a.OwnerID, StartedTurn: gs.Turn, FortLevel: target.FortificationLevel()}
	if defender != nil {
		siege.DefenderArmyID = defender.ID
	}
	gs.Sieges[target.ID] = siege
	a.MovePoints = 0
}
