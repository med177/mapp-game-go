package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
)

func TestMilitaryPowerEstimateIsStableAndBounded(t *testing.T) {
	gs := &GameState{
		Turn:         8,
		DecisionSeed: 42,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"enemy_army": {ID: "enemy_army", OwnerID: "enemy", Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Attack: 100, Defense: 80, HP: 100},
		},
	}

	actual := MilitaryPowerEstimate(gs, "enemy", "enemy")
	estimate := MilitaryPowerEstimate(gs, "player", "enemy")
	if actual <= 0 {
		t.Fatalf("gerçek güç hesaplanmadı: %d", actual)
	}
	if estimate*100 < actual*80 || estimate*100 > actual*120 {
		t.Fatalf("tahmin %%20 sınırının dışında: actual=%d estimate=%d", actual, estimate)
	}
	if again := MilitaryPowerEstimate(gs, "player", "enemy"); again != estimate {
		t.Fatalf("aynı turda istihbarat tahmini değişti: %d != %d", estimate, again)
	}
	if own := MilitaryPowerEstimate(gs, "player", "player"); own != 0 {
		t.Fatalf("ordusuz devletin gücü sıfır olmalı: %d", own)
	}
}

func TestMilitaryPowerEstimateTechnologyRevealsExactValue(t *testing.T) {
	gs := &GameState{
		DecisionSeed: 9,
		Factions: map[faction.FactionID]*faction.Faction{
			"observer": {ID: "observer", Research: faction.ResearchState{Completed: map[string]bool{"intel": true}}},
			"enemy":    {ID: "enemy"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"enemy_army": {ID: "enemy_army", OwnerID: "enemy", Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Attack: 50, HP: 100},
		},
		TechTypes: map[string]*tech.Technology{
			"intel": {ID: "intel", Effects: tech.Effects{RevealEnemyStrength: true}},
		},
	}

	exact := MilitaryPowerEstimate(gs, "enemy", "enemy")
	seen := MilitaryPowerEstimate(gs, "observer", "enemy")
	if exact != seen {
		t.Fatalf("istihbarat teknolojisi gerçek gücü göstermedi: actual=%d seen=%d", exact, seen)
	}
}

func TestDebugMilitaryPowerRevealRequiresDevelopmentMode(t *testing.T) {
	gs := &GameState{
		DevelopmentMode:          true,
		DebugRevealMilitaryPower: true,
		Factions: map[faction.FactionID]*faction.Faction{
			"observer": {ID: "observer"},
			"enemy":    {ID: "enemy"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"enemy_army": {ID: "enemy_army", OwnerID: "enemy", Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Attack: 50, HP: 100},
		},
	}

	actual := MilitaryPowerEstimate(gs, "enemy", "enemy")
	if got := MilitaryPowerEstimate(gs, "observer", "enemy"); got != actual {
		t.Fatalf("geliştirme güç açığı gerçek değeri göstermedi: actual=%d got=%d", actual, got)
	}
	gs.DevelopmentMode = false
	if got := MilitaryPowerEstimate(gs, "observer", "enemy"); got == actual {
		t.Fatalf("geliştirme modu kapalıyken gerçek güç açığı çalışmaya devam ediyor: %d", got)
	}
}
