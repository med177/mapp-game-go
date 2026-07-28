package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func newNavalBattleGame(attacker, defender *army.Army) *Game {
	return &Game{
		gs: &state.GameState{
			PlayerFactionID: "p1",
			Regions: map[world.RegionID]*world.Region{
				"sea_a": {ID: "sea_a", IsSea: true, Neighbors: []world.RegionID{"sea_b"}},
				"sea_b": {ID: "sea_b", IsSea: true, Neighbors: []world.RegionID{"sea_a"}},
			},
			Armies: map[army.ArmyID]*army.Army{
				attacker.ID: attacker,
				defender.ID: defender,
			},
			Factions: map[faction.FactionID]*faction.Faction{
				"p1": {ID: "p1"},
				"p2": {ID: "p2"},
			},
			Relations: map[string]*faction.Relation{
				faction.RelationKey("p1", "p2"): {
					FactionA: "p1",
					FactionB: "p2",
					Stance:   faction.StanceWar,
				},
			},
			UnitTypes: map[string]*army.UnitType{
				"warship":   {ID: "warship", Category: army.CategoryNavalWar, Attack: 100, Defense: 100, Morale: 100},
				"transport": {ID: "transport", Category: army.CategoryNavalTrans, Attack: 1, Defense: 1, Morale: 1},
			},
		},
		renderer: &render.Renderer{},
	}
}

func TestNavalBattleLossSinksFleetAndEmbarkedArmy(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{
			ID:            "attacker_fleet",
			OwnerID:       "p1",
			RegionID:      "sea_a",
			IsNaval:       true,
			MovePoints:    1,
			MaxMovePoints: 1,
			Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
			EmbarkedUnits: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
		},
		&army.Army{
			ID:            "defender_fleet",
			OwnerID:       "p2",
			RegionID:      "sea_b",
			IsNaval:       true,
			MovePoints:    1,
			MaxMovePoints: 1,
			Units:         []army.Unit{{TypeID: "warship", CurrentHP: 100}},
		},
	)

	g.moveArmy("attacker_fleet", "sea_b")

	if _, ok := g.gs.Armies["attacker_fleet"]; ok {
		t.Fatal("deniz savaşını kaybeden filo state'ten kaldırılmalıydı")
	}
	if _, ok := g.gs.Armies["defender_fleet"]; !ok {
		t.Fatal("deniz savaşını kazanan filo korunmalıydı")
	}
}

func TestNavalBattleDefeatSinksDefenderFleetAndEmbarkedArmy(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{
			ID:            "attacker_fleet",
			OwnerID:       "p1",
			RegionID:      "sea_a",
			IsNaval:       true,
			MovePoints:    1,
			MaxMovePoints: 1,
			Units:         []army.Unit{{TypeID: "warship", CurrentHP: 100}},
		},
		&army.Army{
			ID:            "defender_fleet",
			OwnerID:       "p2",
			RegionID:      "sea_b",
			IsNaval:       true,
			MovePoints:    1,
			MaxMovePoints: 1,
			Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
			EmbarkedUnits: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
		},
	)

	g.moveArmy("attacker_fleet", "sea_b")

	if _, ok := g.gs.Armies["defender_fleet"]; ok {
		t.Fatal("deniz savaşını kaybeden savunma filosu state'ten kaldırılmalıydı")
	}
	if _, ok := g.gs.Armies["attacker_fleet"]; !ok {
		t.Fatal("deniz savaşını kazanan saldırı filosu korunmalıydı")
	}
}
