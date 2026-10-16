package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIArmyStrengthUsesEstimateUnlessArmyIsLocallyVisible(t *testing.T) {
	enemy := &army.Army{
		ID:       "enemy-army",
		OwnerID:  "enemy",
		RegionID: "far",
		Units:    []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}},
	}
	gs := &state.GameState{
		DecisionSeed:    42,
		PlayerFactionID: "observer",
		Factions: map[faction.FactionID]*faction.Faction{
			"observer": {ID: "observer"},
			"enemy":    {ID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", Neighbors: []world.RegionID{"near"}},
			"near": {ID: "near", Neighbors: []world.RegionID{"home", "far"}},
			"far":  {ID: "far", Neighbors: []world.RegionID{"near"}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Attack: 10, Defense: 5, Morale: 5, HP: army.MaxUnitHP},
		},
		Armies: map[army.ArmyID]*army.Army{
			"observer-army": {ID: "observer-army", OwnerID: "observer", RegionID: "home"},
			"enemy-army":    enemy,
		},
	}

	actual := aiArmyStrength(gs, enemy)
	baseStrength := enemy.TotalStrength(gs.UnitTypes)
	estimate, exact := state.ArmyPowerEstimate(gs, "observer", enemy)
	if exact {
		t.Fatal("uzak düşman ordusu için istihbarat tahmini exact döndü")
	}
	wantEstimate := maxInt(1, actual*estimate/baseStrength)
	if got := aiArmyStrengthAsSeenBy(gs, "observer", enemy); got != wantEstimate {
		t.Fatalf("uzak düşman gücü = %d, want tahmini %d", got, wantEstimate)
	}

	enemy.RegionID = "near"
	if got := aiArmyStrengthAsSeenBy(gs, "observer", enemy); got != actual {
		t.Fatalf("komşu bölgedeki düşman gücü = %d, want görünen gerçek güç %d", got, actual)
	}
	enemy.RegionID = "home"
	if got := aiArmyStrengthAsSeenBy(gs, "observer", enemy); got != actual {
		t.Fatalf("aynı bölgedeki düşman gücü = %d, want görünen gerçek güç %d", got, actual)
	}
}
