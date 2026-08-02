package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestLandContactOutmatchedAIWithdraws(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"source": {ID: "source", OwnerID: "player", Neighbors: []world.RegionID{"front"}},
			"front":  {ID: "front", OwnerID: "ai", Neighbors: []world.RegionID{"source", "rear"}},
			"rear":   {ID: "rear", OwnerID: "ai", Neighbors: []world.RegionID{"source"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player": {ID: "player", OwnerID: "player", RegionID: "source", Units: []army.Unit{{TypeID: "inf"}, {TypeID: "inf"}, {TypeID: "inf"}}},
			"ai":     {ID: "ai", OwnerID: "ai", RegionID: "front", MovePoints: 1, Units: []army.Unit{{TypeID: "inf"}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "player"): {FactionA: "ai", FactionB: "player", Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50},
		},
	}
	contact := gs.BeginLandContact(gs.Armies["player"], gs.Armies["ai"], "front", "source", state.LandContactMovement)
	ResolveLandContactDecision(gs, contact)
	if contact.DefenderDecision != state.LandContactWithdraw {
		t.Fatalf("gücü düşük AI kara temasında geri çekilmeli: %+v", contact)
	}
	ResolveLandContactWithoutBattle(gs, contact, gs.Armies["ai"])
	if gs.Armies["ai"].RegionID != "rear" || gs.Armies["ai"].MovePoints != 0 {
		t.Fatalf("geri çekilen kara savunucusu güvenli komşuya gitmeli: %+v", gs.Armies["ai"])
	}
}
