package ai

import (
	"fmt"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// raiseWightsAfterVictory, senaryo JSON'unda tanimlanan dirilme kurallarini
// kullanarak kazanan undead faction icin yeni bir ordu olusturur.
func raiseWightsAfterVictory(gs *state.GameState, winnerID string, regionID world.RegionID, fallen int) {
	if gs == nil || regionID == "" || fallen <= 0 {
		return
	}
	rules := gs.UndeadMechanics.WithDefaults()
	if winnerID != rules.FactionID || rules.FactionID == "" || rules.RaisedUnitType == "" {
		return
	}
	if rules.RequiresEventFlag != "" && !gs.FiredEventIDs["flag:"+rules.RequiresEventFlag] {
		return
	}
	unitType := gs.UnitTypes[rules.RaisedUnitType]
	if unitType == nil {
		return
	}
	if gs.Armies == nil {
		gs.Armies = make(map[army.ArmyID]*army.Army)
	}
	percent := rules.FallenToRaisedPercent
	if percent > 100 {
		percent = 100
	}
	count := fallen * percent / 100
	if count < rules.MinimumRaisedUnits {
		count = rules.MinimumRaisedUnits
	}
	if rules.MaximumRaisedUnits > 0 && count > rules.MaximumRaisedUnits {
		count = rules.MaximumRaisedUnits
	}
	gs.NextArmySeq++
	id := army.ArmyID(fmt.Sprintf("army_%s_raised_%d", rules.FactionID, gs.NextArmySeq))
	for gs.Armies[id] != nil {
		gs.NextArmySeq++
		id = army.ArmyID(fmt.Sprintf("army_%s_raised_%d", rules.FactionID, gs.NextArmySeq))
	}
	movePoints := unitType.BaseMovementPoints()
	gs.Armies[id] = &army.Army{
		ID:            id,
		OwnerID:       rules.FactionID,
		RegionID:      regionID,
		Units:         army.MakeUnits(rules.RaisedUnitType, count),
		MovePoints:    movePoints,
		MaxMovePoints: movePoints,
	}
}
