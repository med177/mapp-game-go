package game

import (
	"math/rand"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestMoveArmyWithStanceChangesBattleResolution(t *testing.T) {
	newGame := func() *Game {
		types := map[string]*army.UnitType{
			"inf": {ID: "inf", NameTR: "Piyade", Attack: 12, Defense: 10, Morale: 50},
		}
		gs := &state.GameState{
			PlayerFactionID: "p1",
			Regions: map[world.RegionID]*world.Region{
				"src": {ID: "src", OwnerID: "p1", Neighbors: []world.RegionID{"dst"}},
				"dst": {ID: "dst", OwnerID: "p2", Neighbors: []world.RegionID{"src"}},
			},
			Armies: map[army.ArmyID]*army.Army{
				"atk": {
					ID:            "atk",
					OwnerID:       "p1",
					RegionID:      "src",
					MovePoints:    2,
					MaxMovePoints: 2,
					Units: []army.Unit{
						{TypeID: "inf", CurrentHP: 100},
						{TypeID: "inf", CurrentHP: 100},
						{TypeID: "inf", CurrentHP: 100},
						{TypeID: "inf", CurrentHP: 100},
						{TypeID: "inf", CurrentHP: 100},
						{TypeID: "inf", CurrentHP: 100},
					},
				},
				"def": {
					ID:            "def",
					OwnerID:       "p2",
					RegionID:      "dst",
					MovePoints:    2,
					MaxMovePoints: 2,
					Units: []army.Unit{
						{TypeID: "inf", CurrentHP: 100},
						{TypeID: "inf", CurrentHP: 100},
						{TypeID: "inf", CurrentHP: 100},
						{TypeID: "inf", CurrentHP: 100},
					},
				},
			},
			Factions: map[faction.FactionID]*faction.Faction{
				"p1": {ID: "p1"},
				"p2": {ID: "p2"},
			},
			Relations: map[string]*faction.Relation{
				faction.RelationKey("p1", "p2"): {FactionA: "p1", FactionB: "p2", Stance: faction.StanceWar},
			},
			UnitTypes: types,
		}
		return &Game{gs: gs, renderer: &render.Renderer{}}
	}

	rand.Seed(7)
	aggressiveGame := newGame()
	aggressiveGame.moveArmyWithStance("atk", "dst", combat.BattleStanceAggressive)

	rand.Seed(7)
	defensiveGame := newGame()
	defensiveGame.moveArmyWithStance("atk", "dst", combat.BattleStanceDefensive)

	aggArmy := aggressiveGame.gs.Armies["atk"]
	defArmy := defensiveGame.gs.Armies["atk"]
	if aggArmy == nil || defArmy == nil {
		t.Fatal("saldıran orduların savaş sonrası hayatta kalması bekleniyordu")
	}
	if aggArmy.RegionID != "dst" || defArmy.RegionID != "dst" {
		t.Fatalf("bu senaryoda her iki duruşta da bölgenin ele geçirilmesi bekleniyordu, aggressive=%s defensive=%s", aggArmy.RegionID, defArmy.RegionID)
	}
	totalHP := func(a *army.Army) int {
		total := 0
		for _, unit := range a.Units {
			total += unit.CurrentHP
		}
		return total
	}
	if totalHP(aggArmy) == totalHP(defArmy) {
		t.Fatalf("duruş seçimi aynı zar altında farklı savaş izi bırakmalıydı, aggressive=%d defensive=%d", totalHP(aggArmy), totalHP(defArmy))
	}
}
