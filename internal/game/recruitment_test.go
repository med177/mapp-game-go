package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestRecruitSpecificAllowsQueueWhenExistingArmyIsFull(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "p1"},
			"r2": {ID: "r2", OwnerID: "p1"},
			"r3": {ID: "r3", OwnerID: "p1"},
			"r4": {ID: "r4", OwnerID: "p1"},
			"r5": {ID: "r5", OwnerID: "p1"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_p1_1": {
				ID:            "army_p1_1",
				OwnerID:       "p1",
				RegionID:      "r1",
				Units:         filledUnits(army.MaxArmySize, "militia"),
				MovePoints:    2,
				MaxMovePoints: 2,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia": {ID: "militia", NameTR: "Milis", TurnsRequired: 1},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.recruitSpecific("r1", "militia", 1)

	if got := len(gs.ProductionQueue); got != 1 {
		t.Fatalf("full ordu varken recruit kuyruğu oluşmalıydı, got=%d", got)
	}
	order := gs.ProductionQueue[0]
	if order.RegionID != "r1" || order.TypeID != "militia" {
		t.Fatalf("beklenmeyen üretim emri: %+v", order)
	}
}

func TestApplyProductionTicksCreatesNewArmyWhenExistingArmyIsFull(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     1,
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "p1"},
			"r2": {ID: "r2", OwnerID: "p1"},
			"r3": {ID: "r3", OwnerID: "p1"},
			"r4": {ID: "r4", OwnerID: "p1"},
			"r5": {ID: "r5", OwnerID: "p1"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_p1_1": {
				ID:            "army_p1_1",
				OwnerID:       "p1",
				RegionID:      "r1",
				Units:         filledUnits(army.MaxArmySize, "militia"),
				MovePoints:    2,
				MaxMovePoints: 2,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia": {ID: "militia", NameTR: "Milis"},
		},
		ProductionQueue: []state.ProductionOrder{
			{
				ID:        "prod_1",
				Kind:      productionKindUnit,
				FactionID: "p1",
				RegionID:  "r1",
				TypeID:    "militia",
				TurnsLeft: 1,
			},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	results := g.applyProductionTicks()

	if len(results) != 1 || results[0].delayed || results[0].canceled {
		t.Fatalf("üretim tamamlanmalıydı, results=%+v", results)
	}
	if got := len(gs.ProductionQueue); got != 0 {
		t.Fatalf("tamamlanan üretim kuyruktan düşmeliydi, got=%d", got)
	}
	newArmy, ok := gs.Armies["army_p1_2"]
	if !ok {
		t.Fatalf("full ordu yanında yeni ordu spawn edilmeliydi")
	}
	if newArmy.RegionID != "r1" || len(newArmy.Units) != 1 || newArmy.Units[0].TypeID != "militia" {
		t.Fatalf("yeni ordu hatalı oluşturuldu: %+v", newArmy)
	}
}

func TestApplyProductionTicksCompletesAIUnitOrder(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		NextArmySeq:     3,
		Regions: map[world.RegionID]*world.Region{
			"ai_home": {ID: "ai_home", OwnerID: "ai_1"},
		},
		Armies: map[army.ArmyID]*army.Army{},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ai_1":   {ID: "ai_1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia": {ID: "militia", NameTR: "Milis"},
		},
		ProductionQueue: []state.ProductionOrder{
			{
				ID:        "prod_ai_1",
				Kind:      productionKindUnit,
				FactionID: "ai_1",
				RegionID:  "ai_home",
				TypeID:    "militia",
				TurnsLeft: 1,
			},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	results := g.applyProductionTicks()

	if len(results) != 1 || results[0].factionID != "ai_1" || results[0].delayed || results[0].canceled {
		t.Fatalf("AI üretim emri tamamlanmalıydı, results=%+v", results)
	}
	if got := len(gs.ProductionQueue); got != 0 {
		t.Fatalf("AI üretimi kuyruktan düşmeli, got=%d", got)
	}
	armyRef := gs.Armies["army_ai_1_4"]
	if armyRef == nil || armyRef.OwnerID != "ai_1" || armyRef.RegionID != "ai_home" || len(armyRef.Units) != 1 {
		t.Fatalf("AI üretimi AI ordusu oluşturmalıydı, armies=%+v", gs.Armies)
	}
}

func filledUnits(count int, unitTypeID string) []army.Unit {
	units := make([]army.Unit, count)
	for i := range units {
		units[i] = army.Unit{TypeID: unitTypeID, CurrentHP: 100}
	}
	return units
}
