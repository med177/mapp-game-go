package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAITurnPreludeAssignsCommandersToFieldArmies(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", NameTR: "AI", Gold: 1000, Grain: 200},
		},
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "ai"},
			"r2": {ID: "r2", OwnerID: "ai"},
			"r3": {ID: "r3", OwnerID: "ai"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_b":   {ID: "army_b", OwnerID: "ai", RegionID: "r2", Units: []army.Unit{{TypeID: "inf"}}},
			"army_a":   {ID: "army_a", OwnerID: "ai", RegionID: "r1", Units: []army.Unit{{TypeID: "inf"}}},
			"garrison": {ID: "garrison", OwnerID: "ai", RegionID: "r3", IsGarrison: true},
		},
	}

	runTurnPrelude(gs, "ai", nil)

	if gs.Armies["army_a"] == nil || gs.Armies["army_b"] == nil {
		t.Fatalf("AI prelude saha ordularından birini beklenmedik şekilde sildi: %+v", gs.Armies)
	}
	if gs.Armies["army_a"].Commander == nil || gs.Armies["army_b"].Commander == nil {
		t.Fatal("AI prelude saha ordularına komutan atamadı")
	}
	if gs.Armies["army_a"].Commander == gs.Armies["army_b"].Commander {
		t.Fatal("AI prelude aynı komutanı iki orduya atadı")
	}
	if gs.Armies["garrison"].Commander != nil {
		t.Fatal("AI prelude garnizona otomatik komutan atadı")
	}
}
