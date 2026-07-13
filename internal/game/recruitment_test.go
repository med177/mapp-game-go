package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
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

func TestApplyProductionTicksUsesBarracksLevelAsLandTurnCapacity(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     0,
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "p1", Buildings: []string{"barracks", "barracks"}},
		},
		Armies: map[army.ArmyID]*army.Army{},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia": {ID: "militia", NameTR: "Milis"},
		},
		ProductionQueue: []state.ProductionOrder{
			{ID: "prod_1", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "militia", TurnsLeft: 1},
			{ID: "prod_2", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "militia", TurnsLeft: 1},
			{ID: "prod_3", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "militia", TurnsLeft: 1},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	results := g.applyProductionTicks()

	if len(results) != 2 {
		t.Fatalf("2 üretim sonucu bekleniyordu, got=%d", len(results))
	}
	if results[0].delayed || results[1].delayed {
		t.Fatalf("ilk iki kara üretimi tamamlanmalıydı, results=%+v", results)
	}
	if got := len(gs.ProductionQueue); got != 1 {
		t.Fatalf("bir emir sırada kalmalıydı, got=%d", got)
	}
	if gs.ProductionQueue[0].TurnsLeft != 1 {
		t.Fatalf("bekleyen emir tur sayısını korumalıydı, got=%d", gs.ProductionQueue[0].TurnsLeft)
	}
	armyRef := gs.Armies["army_p1_1"]
	if armyRef == nil || len(armyRef.Units) != 2 {
		t.Fatalf("aynı turda 2 kara birimi tamamlanmalıydı, army=%+v", armyRef)
	}
}

func TestApplyProductionTicksSeparatesBarracksAndPortTurnCapacity(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     0,
		Regions: map[world.RegionID]*world.Region{
			"r1":   {ID: "r1", OwnerID: "p1", Buildings: []string{"barracks", "port", "port"}, Neighbors: []world.RegionID{"sea1"}},
			"sea1": {ID: "sea1", IsSea: true},
		},
		Armies: map[army.ArmyID]*army.Army{},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia":   {ID: "militia", NameTR: "Milis"},
			"transport": {ID: "transport", NameTR: "Nakliye Gemisi", RequiredBldg: "port"},
		},
		ProductionQueue: []state.ProductionOrder{
			{ID: "prod_1", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "militia", TurnsLeft: 1},
			{ID: "prod_2", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "transport", TurnsLeft: 1},
			{ID: "prod_3", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "transport", TurnsLeft: 1},
			{ID: "prod_4", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "transport", TurnsLeft: 1},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	results := g.applyProductionTicks()

	if len(results) != 3 {
		t.Fatalf("3 üretim sonucu bekleniyordu, got=%d", len(results))
	}
	if results[0].delayed || results[1].delayed || results[2].delayed {
		t.Fatalf("kışla 1 + liman 2 kapasitesiyle ilk 3 üretim tamamlanmalıydı, results=%+v", results)
	}
	if got := len(gs.ProductionQueue); got != 1 {
		t.Fatalf("bir deniz emri sırada kalmalıydı, got=%d", got)
	}
	if gs.ProductionQueue[0].TurnsLeft != 1 {
		t.Fatalf("bekleyen deniz emri tur sayısını korumalıydı, got=%d", gs.ProductionQueue[0].TurnsLeft)
	}
	landArmy := gs.Armies["army_p1_1"]
	if landArmy == nil || len(landArmy.Units) != 1 || landArmy.Units[0].TypeID != "militia" {
		t.Fatalf("kara birimi tamamlanmalıydı, army=%+v", landArmy)
	}
	fleet := gs.Armies["fleet_p1_2"]
	if fleet == nil || len(fleet.Units) != 2 || !fleet.IsNaval || fleet.RegionID != "sea1" {
		t.Fatalf("aynı turda 2 deniz birimi tamamlanmalıydı, fleet=%+v", fleet)
	}
}

func TestApplyProductionTicksOnlyActiveCapacityAdvancesTurns(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "p1", Buildings: []string{"barracks", "barracks", "barracks"}},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", NameTR: "Piyade", TurnsRequired: 3},
		},
		ProductionQueue: []state.ProductionOrder{
			{ID: "prod_1", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "infantry", TurnsLeft: 3},
			{ID: "prod_2", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "infantry", TurnsLeft: 3},
			{ID: "prod_3", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "infantry", TurnsLeft: 3},
			{ID: "prod_4", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "infantry", TurnsLeft: 3},
			{ID: "prod_5", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "infantry", TurnsLeft: 3},
			{ID: "prod_6", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "infantry", TurnsLeft: 3},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	results := g.applyProductionTicks()

	if len(results) != 0 {
		t.Fatalf("henüz tamamlanan üretim olmamalıydı, results=%+v", results)
	}
	if got := len(gs.ProductionQueue); got != 6 {
		t.Fatalf("tüm emirler kuyrukta kalmalıydı, got=%d", got)
	}
	for i := 0; i < 3; i++ {
		if gs.ProductionQueue[i].TurnsLeft != 2 {
			t.Fatalf("aktif slotlardaki ilk 3 emir ilerlemeliydi, index=%d got=%d", i, gs.ProductionQueue[i].TurnsLeft)
		}
	}
	for i := 3; i < 6; i++ {
		if gs.ProductionQueue[i].TurnsLeft != 3 {
			t.Fatalf("bekleyen emir tur sayısını korumalıydı, index=%d got=%d", i, gs.ProductionQueue[i].TurnsLeft)
		}
	}
}

func TestApplyProductionTicksPausesSiegedRegionProductionAndResumesAfterLift(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "p1"},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia": {ID: "militia", NameTR: "Milis"},
		},
		ProductionQueue: []state.ProductionOrder{
			{ID: "prod_1", Kind: productionKindBuilding, FactionID: "p1", RegionID: "r1", TypeID: "walls", TurnsLeft: 2},
			{ID: "prod_2", Kind: productionKindUnit, FactionID: "p1", RegionID: "r1", TypeID: "militia", TurnsLeft: 2},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"r1": {RegionID: "r1", AttackerArmyID: "atk", AttackerFactionID: "p2"},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	results := g.applyProductionTicks()

	if len(results) != 0 {
		t.Fatalf("kuşatma altındaki üretimler ilerlememeliydi, results=%+v", results)
	}
	if got := len(gs.ProductionQueue); got != 2 {
		t.Fatalf("kuşatma altındayken kuyruk korunmalıydı, got=%d", got)
	}
	if gs.ProductionQueue[0].TurnsLeft != 2 || gs.ProductionQueue[1].TurnsLeft != 2 {
		t.Fatalf("kuşatma altındayken TurnsLeft sabit kalmalıydı, queue=%+v", gs.ProductionQueue)
	}

	delete(gs.Sieges, "r1")

	results = g.applyProductionTicks()

	if len(results) != 0 {
		t.Fatalf("kuşatma kalkınca hemen tamamlanmamalıydı, results=%+v", results)
	}
	if got := len(gs.ProductionQueue); got != 2 {
		t.Fatalf("kuşatma kalkınca kuyruk korunmalıydı, got=%d", got)
	}
	if gs.ProductionQueue[0].TurnsLeft != 1 || gs.ProductionQueue[1].TurnsLeft != 1 {
		t.Fatalf("kuşatma kalkınca üretim devam etmeliydi, queue=%+v", gs.ProductionQueue)
	}
}

func TestRecruitSpecificIgnoresGarrisonArmyForLimitAndCompletion(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		NextArmySeq:     2,
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "p1"},
			"r2": {ID: "r2", OwnerID: "p1"},
			"r3": {ID: "r3", OwnerID: "p1"},
			"r4": {ID: "r4", OwnerID: "p1"},
			"r5": {ID: "r5", OwnerID: "p1"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_p1_1":         {ID: "army_p1_1", OwnerID: "p1", RegionID: "r2", Units: filledUnits(5, "militia")},
			"army_p1_2":         {ID: "army_p1_2", OwnerID: "p1", RegionID: "r3", Units: filledUnits(4, "militia")},
			"army_garrison_209": {ID: "army_garrison_209", OwnerID: "p1", RegionID: "r1", Units: filledUnits(2, "militia"), IsGarrison: true},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia": {ID: "militia", NameTR: "Milis", TurnsRequired: 1},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if got := gs.CurrentLandArmies("p1"); got != 2 {
		t.Fatalf("garnizon kara ordu limitinde sayilmamaliydi, got=%d", got)
	}

	g.recruitSpecific("r1", "militia", 1)
	results := g.applyProductionTicks()

	if len(results) != 1 || results[0].delayed || results[0].canceled {
		t.Fatalf("garnizon yaninda yeni saha ordusu uretilmeliydi, results=%+v", results)
	}
	if got := len(gs.ProductionQueue); got != 0 {
		t.Fatalf("tamamlanan emir kuyruktan dusmeliydi, got=%d", got)
	}
	garrison := gs.Armies["army_garrison_209"]
	if garrison == nil || !garrison.IsGarrison || len(garrison.Units) != 2 {
		t.Fatalf("garnizon recruit hedefi olmamaliydi, got=%+v", garrison)
	}
	newArmy := gs.Armies["army_p1_3"]
	if newArmy == nil || newArmy.IsGarrison || newArmy.RegionID != "r1" || len(newArmy.Units) != 1 {
		t.Fatalf("yeni saha ordusu olusmadi: %+v", newArmy)
	}
}

func TestMoveArmyDeploysGarrisonIntoFieldArmy(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "p1", Neighbors: []world.RegionID{"r2"}},
			"r2": {ID: "r2", OwnerID: "p1", Neighbors: []world.RegionID{"r1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_garrison_209": {
				ID:            "army_garrison_209",
				OwnerID:       "p1",
				RegionID:      "r1",
				Units:         filledUnits(2, "militia"),
				MovePoints:    2,
				MaxMovePoints: 2,
				IsGarrison:    true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmyWithStance("army_garrison_209", "r2", combat.BattleStanceBalanced)

	if gs.Armies["army_garrison_209"] != nil {
		t.Fatalf("deploy edilen garnizon eski ID ile kalmamaliydi")
	}
	fieldArmy := gs.Armies["army_p1_1"]
	if fieldArmy == nil {
		t.Fatalf("garnizon saha ordusuna donusmeliydi, armies=%+v", gs.Armies)
	}
	if fieldArmy.IsGarrison {
		t.Fatalf("hareket eden garnizon saha ordusuna donusmeliydi")
	}
	if fieldArmy.RegionID != "r2" {
		t.Fatalf("ordu hedef bolgeye gitmeliydi, got=%s", fieldArmy.RegionID)
	}
	if got := gs.CurrentLandArmies("p1"); got != 1 {
		t.Fatalf("deploy sonrasi saha ordu sayisi 1 olmaliydi, got=%d", got)
	}
}

func filledUnits(count int, unitTypeID string) []army.Unit {
	units := make([]army.Unit, count)
	for i := range units {
		units[i] = army.Unit{TypeID: unitTypeID, CurrentHP: 100}
	}
	return units
}
