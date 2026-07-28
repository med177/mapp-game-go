package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func movementCompositionTestState() *state.GameState {
	return &state.GameState{
		Month:           6,
		PlayerFactionID: "p1",
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "p1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"cav":   {ID: "cav", MovementPoints: 3},
			"inf":   {ID: "inf", MovementPoints: 2},
			"siege": {ID: "siege", MovementPoints: 1},
		},
	}
}

func TestSplitArmyRefreshesUnmovedPartsBySlowestUnit(t *testing.T) {
	gs := movementCompositionTestState()
	gs.Armies = map[army.ArmyID]*army.Army{
		"main": {
			ID: "main", OwnerID: "p1", RegionID: "home",
			MovePoints: 3, MaxMovePoints: 3,
			Units: []army.Unit{
				{TypeID: "cav"},
				{TypeID: "inf"},
				{TypeID: "siege"},
			},
		},
	}
	game := &Game{gs: gs, renderer: &render.Renderer{}}

	game.splitArmy("main", 0)

	main := gs.Armies["main"]
	split := gs.Armies["army_p1_1"]
	if main == nil || split == nil {
		t.Fatal("split sonrası iki ordu da mevcut olmalıydı")
	}
	if main.MaxMovePoints != 1 || main.MovePoints != 1 {
		t.Fatalf("ana parça kuşatma birimine göre 1/1 hareket etmeliydi, got=%d/%d", main.MovePoints, main.MaxMovePoints)
	}
	if split.MaxMovePoints != 3 || split.MovePoints != 3 {
		t.Fatalf("ayrılan süvari parçası 3/3 hareket etmeliydi, got=%d/%d", split.MovePoints, split.MaxMovePoints)
	}
}

func TestMergeArmiesRefreshesUnmovedResultBySlowestUnit(t *testing.T) {
	gs := movementCompositionTestState()
	gs.Armies = map[army.ArmyID]*army.Army{
		"main": {
			ID: "main", OwnerID: "p1", RegionID: "home",
			MovePoints: 3, MaxMovePoints: 3,
			Units: []army.Unit{{TypeID: "cav"}},
		},
		"reinforcement": {
			ID: "reinforcement", OwnerID: "p1", RegionID: "home",
			MovePoints: 2, MaxMovePoints: 2,
			Units: []army.Unit{{TypeID: "siege"}},
		},
	}
	game := &Game{gs: gs, renderer: &render.Renderer{}}

	game.mergeArmiesManual("reinforcement")

	main := gs.Armies["main"]
	if main == nil {
		t.Fatal("birleşme sonrası hedef ordu korunmalıydı")
	}
	if main.MaxMovePoints != 1 || main.MovePoints != 1 {
		t.Fatalf("birleşen ordu en yavaş birime göre 1/1 hareket etmeliydi, got=%d/%d", main.MovePoints, main.MaxMovePoints)
	}
}

func TestMergeArmiesUsesExplicitTargetArmy(t *testing.T) {
	gs := movementCompositionTestState()
	gs.Armies = map[army.ArmyID]*army.Army{
		"source": {
			ID: "source", OwnerID: "p1", RegionID: "home",
			Units: []army.Unit{{TypeID: "inf"}, {TypeID: "inf"}},
		},
		"target_a": {
			ID: "target_a", OwnerID: "p1", RegionID: "home",
			Units: []army.Unit{{TypeID: "inf"}},
		},
		"target_b": {
			ID: "target_b", OwnerID: "p1", RegionID: "home",
			Units: []army.Unit{{TypeID: "inf"}, {TypeID: "inf"}, {TypeID: "inf"}},
		},
	}
	game := &Game{gs: gs, renderer: &render.Renderer{}}

	game.mergeArmiesManual("source", "target_b")

	if gs.Armies["source"] != nil || gs.Armies["target_b"] == nil {
		t.Fatal("açık hedefle birleşmede kaynak ordu silinip seçilen hedef korunmalıydı")
	}
	if got := len(gs.Armies["target_b"].Units); got != 5 {
		t.Fatalf("seçilen hedef orduya kaynak birimleri eklenmeliydi: got=%d", got)
	}
	if got := len(gs.Armies["target_a"].Units); got != 1 {
		t.Fatalf("seçilmeyen hedef ordu değişmemeliydi: got=%d", got)
	}
}

func TestSplitArmyDoesNotRefundMovementAlreadyUsedThisTurn(t *testing.T) {
	gs := movementCompositionTestState()
	gs.Armies = map[army.ArmyID]*army.Army{
		"main": {
			ID: "main", OwnerID: "p1", RegionID: "home",
			MovePoints: 2, MaxMovePoints: 3,
			Units: []army.Unit{{TypeID: "cav"}, {TypeID: "siege"}},
		},
	}
	game := &Game{gs: gs, renderer: &render.Renderer{}}

	game.splitArmy("main", 0)

	main := gs.Armies["main"]
	split := gs.Armies["army_p1_1"]
	if main.MovePoints != 1 || main.MaxMovePoints != 1 {
		t.Fatalf("hareket etmiş ana parça kalan puanı iade etmemeliydi, got=%d/%d", main.MovePoints, main.MaxMovePoints)
	}
	if split.MovePoints != 2 || split.MaxMovePoints != 3 {
		t.Fatalf("ayrılan parça kullanılmayan hareketi iade etmemeliydi, got=%d/%d", split.MovePoints, split.MaxMovePoints)
	}
}
