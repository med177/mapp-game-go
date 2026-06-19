package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestNavalShowsFriendlyDisembark(t *testing.T) {
	gs := &state.GameState{}
	fleet := &army.Army{
		ID:            "fleet",
		OwnerID:       "p1",
		RegionID:      "sea_1",
		IsNaval:       true,
		EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
	}

	if !navalShowsFriendlyDisembark(gs, fleet, &world.Region{ID: "land_a", OwnerID: "p1"}) {
		t.Fatal("kendi kara bölgesi için IN davranışı bekleniyordu")
	}
	if !navalShowsFriendlyDisembark(gs, fleet, &world.Region{ID: "land_b"}) {
		t.Fatal("boş kara bölgesi için IN davranışı bekleniyordu")
	}
	if navalShowsFriendlyDisembark(gs, fleet, &world.Region{ID: "land_e", OwnerID: "p2"}) {
		t.Fatal("düşman kara bölgesi için friendly IN davranışı olmamalı")
	}
	if navalShowsFriendlyDisembark(gs, fleet, &world.Region{ID: "sea_1", IsSea: true}) {
		t.Fatal("deniz bölgesi için friendly IN davranışı olmamalı")
	}
}

func TestBattlePlanIntentCoversNavalAndAmphibiousCombat(t *testing.T) {
	r := &Renderer{}

	landingFleet := &army.Army{
		ID:            "fleet",
		OwnerID:       "p1",
		RegionID:      "sea_1",
		IsNaval:       true,
		EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
	}
	enemyArmy := &army.Army{
		ID:       "enemy_land",
		OwnerID:  "p2",
		RegionID: "land_a",
	}
	action, context, ok := r.battlePlanIntent(landingFleet, &world.Region{ID: "land_a", OwnerID: "p2"}, enemyArmy)
	if !ok {
		t.Fatal("düşman kıyıya çıkarma için savaş planı bekleniyordu")
	}
	if action != ActionDisembarkArmy {
		t.Fatalf("çıkarma için beklenen aksiyon disembark olmalı, got=%s", action)
	}
	if context != combat.BattleContextAmphibious {
		t.Fatalf("çıkarma için beklenen context amphibious olmalı, got=%s", context)
	}

	navalFleet := &army.Army{
		ID:       "fleet",
		OwnerID:  "p1",
		RegionID: "sea_1",
		IsNaval:  true,
		Units:    []army.Unit{{TypeID: "ship", CurrentHP: 100}},
	}
	enemyFleet := &army.Army{
		ID:       "enemy_fleet",
		OwnerID:  "p2",
		RegionID: "sea_2",
		IsNaval:  true,
		Units:    []army.Unit{{TypeID: "ship", CurrentHP: 100}},
	}
	action, context, ok = r.battlePlanIntent(navalFleet, &world.Region{ID: "sea_2", IsSea: true}, enemyFleet)
	if !ok {
		t.Fatal("düşman donanmaya karşı deniz savaş planı bekleniyordu")
	}
	if action != ActionMoveArmy {
		t.Fatalf("deniz savaşı için beklenen aksiyon move olmalı, got=%s", action)
	}
	if context != combat.BattleContextNaval {
		t.Fatalf("deniz savaşı için beklenen context naval olmalı, got=%s", context)
	}
}

func TestOpenBattlePlanUsesEmbarkedUnitsForAmphibiousPreview(t *testing.T) {
	gs := &state.GameState{
		UnitTypes: map[string]*army.UnitType{
			"inf":  {ID: "inf", Attack: 12, Defense: 10, Morale: 50},
			"ship": {ID: "ship", Attack: 30, Defense: 20, Morale: 50},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1", NameTR: "Ahiler"},
			"p2": {ID: "p2", NameTR: "Düşman"},
		},
	}
	r := &Renderer{gs: gs}
	fleet := &army.Army{
		ID:            "fleet",
		OwnerID:       "p1",
		IsNaval:       true,
		Units:         []army.Unit{{TypeID: "ship", CurrentHP: 100}},
		EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
	}
	defender := &army.Army{
		ID:      "enemy",
		OwnerID: "p2",
		Units:   []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
	}
	target := &world.Region{ID: "land_a", NameTR: "Liman", Terrain: world.TerrainCoast}

	r.openBattlePlan(fleet, target, defender, ActionDisembarkArmy, combat.BattleContextAmphibious)

	landing := &army.Army{
		OwnerID: fleet.OwnerID,
		Units:   fleet.EmbarkedUnits,
	}
	expected := combat.PreviewBattleWithContextMods(landing, defender, target.Terrain, gs.UnitTypes, combat.TechMods{}, combat.TechMods{}, combat.BattleContextAmphibious, combat.BattleStanceBalanced)
	shipBased := combat.PreviewBattleWithContextMods(fleet, defender, target.Terrain, gs.UnitTypes, combat.TechMods{}, combat.TechMods{}, combat.BattleContextAmphibious, combat.BattleStanceBalanced)

	if r.battlePlan.previews[1].AttackStrength != expected.AttackStrength {
		t.Fatalf("çıkarma preview gücü embarked birliklerden gelmeli, got=%d want=%d", r.battlePlan.previews[1].AttackStrength, expected.AttackStrength)
	}
	if expected.AttackStrength == shipBased.AttackStrength {
		t.Fatalf("test kurulumu gemi ve çıkarma gücünü ayırt etmeliydi")
	}
}
