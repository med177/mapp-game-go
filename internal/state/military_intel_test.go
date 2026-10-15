package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
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

func TestTemporaryMilitaryIntelIsTargetedAndExpires(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"observer": {ID: "observer"},
			"target":   {ID: "target"},
			"other":    {ID: "other"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"target_army": {ID: "target_army", OwnerID: "target", Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Attack: 100, Defense: 80, HP: 100},
		},
	}
	gs.GrantTemporaryMilitaryIntel("observer", "target", 2)
	gs.GrantTemporaryMilitaryIntel("observer", "target", 1)
	if !HasTemporaryMilitaryIntel(gs, "observer", "target") {
		t.Fatal("targeted military intel was not granted")
	}
	if _, _, exact := MilitaryPowerEstimateBreakdown(gs, "observer", "target"); !exact {
		t.Fatal("temporary military intel did not reveal exact target strength")
	}
	if _, _, exact := MilitaryPowerEstimateBreakdown(gs, "observer", "other"); exact {
		t.Fatal("targeted military intel revealed an unrelated faction")
	}
	targetArmy := gs.Armies["target_army"]
	if _, exact := ArmyPowerEstimate(gs, "observer", targetArmy); !exact {
		t.Fatal("temporary military intel did not reveal exact target army strength")
	}
	if _, exact := ArmyDefenseEstimate(gs, "observer", targetArmy); !exact {
		t.Fatal("temporary military intel did not reveal exact target army defense")
	}
	otherArmy := &army.Army{ID: "other_army", OwnerID: "other", Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}}}
	if _, exact := ArmyPowerEstimate(gs, "observer", otherArmy); exact {
		t.Fatal("targeted military intel revealed an unrelated army's strength")
	}

	gs.AdvanceTurn()
	if turns := gs.TemporaryMilitaryIntel["observer"]["target"]; turns != 1 {
		t.Fatalf("temporary intel after first turn = %d turns, want 1", turns)
	}
	gs.AdvanceTurn()
	if HasTemporaryMilitaryIntel(gs, "observer", "target") {
		t.Fatal("temporary military intel did not expire")
	}
}

func TestTemporaryRegionMilitaryIntelIsScopedAndExpires(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"observer": {ID: "observer"},
			"enemy":    {ID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			"scouted": {ID: "scouted", OwnerID: "enemy"},
			"other":   {ID: "other", OwnerID: "enemy"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"scouted_army": {ID: "scouted_army", OwnerID: "enemy", RegionID: "scouted", Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}}},
			"other_army":   {ID: "other_army", OwnerID: "enemy", RegionID: "other", Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Attack: 100, Defense: 80, HP: 100},
		},
	}
	gs.GrantTemporaryRegionMilitaryIntel("observer", "scouted", 2)
	if _, exact := ArmyPowerEstimate(gs, "observer", gs.Armies["scouted_army"]); !exact {
		t.Fatal("region military intel did not reveal the scouted army's exact strength")
	}
	if _, exact := ArmyDefenseEstimate(gs, "observer", gs.Armies["scouted_army"]); !exact {
		t.Fatal("region military intel did not reveal the scouted army's exact defense")
	}
	if _, exact := ArmyPowerEstimate(gs, "observer", gs.Armies["other_army"]); exact {
		t.Fatal("region military intel revealed an army in another region")
	}
	gs.AdvanceTurn()
	if !HasTemporaryRegionMilitaryIntel(gs, "observer", "scouted") {
		t.Fatal("region intel expired before its final turn")
	}
	gs.AdvanceTurn()
	if HasTemporaryRegionMilitaryIntel(gs, "observer", "scouted") {
		t.Fatal("region intel did not expire")
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
