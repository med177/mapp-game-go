package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIReliefArmyCreatesContactWithPlayerBesieger(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
			"p2": {ID: "p2"},
		},
		Regions: map[world.RegionID]*world.Region{
			"source": {ID: "source", OwnerID: "p2", Neighbors: []world.RegionID{"target"}},
			"target": {
				ID: "target", OwnerID: "p2", Neighbors: []world.RegionID{"source"},
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fort", Type: world.SettlementFortress}},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p1", "p2"): {FactionA: "p1", FactionB: "p2", Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"besieger": {ID: "besieger", OwnerID: "p1", RegionID: "target", MovePoints: 1, MaxMovePoints: 1, Units: testContactUnits(2)},
			"relief":   {ID: "relief", OwnerID: "p2", RegionID: "source", MovePoints: 1, MaxMovePoints: 1, Units: testContactUnits(18)},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"target": {RegionID: "target", AttackerArmyID: "besieger", AttackerFactionID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 10, Defense: 10, Morale: 50},
		},
	}

	outcome := executeMoveWithNavalPatrolAndContact(gs, gs.Armies["relief"], "target", "p2", false, false)
	if gs.PendingLandContact == nil {
		t.Fatal("aktif oyuncu kuşatmasına gelen AI ordusu kara teması oluşturmalıydı")
	}
	if gs.PendingLandContact.PlayerArmyID != "besieger" {
		t.Fatalf("temas kararı oyuncu kuşatan orduya ait olmalıydı: %+v", gs.PendingLandContact)
	}
	if gs.Armies["relief"].RegionID != "target" {
		t.Fatalf("temas popup'ı açılırken AI ordusu hedef bölgeye girmiş görünmeli: %+v", gs.Armies["relief"])
	}
	if outcome.step.Kind != TurnStepBattle {
		t.Fatalf("aktif kuşatma teması doğrudan savaşa değil temas adımına dönmeli: %+v", outcome.step)
	}
	if gs.SiegeAt("target") == nil {
		t.Fatal("oyuncunun kuşatması temas kararı verilmeden kaldırılmamalıydı")
	}
}

func testContactUnits(count int) []army.Unit {
	units := make([]army.Unit, count)
	for i := range units {
		units[i] = army.Unit{TypeID: "inf", CurrentHP: 100}
	}
	return units
}
