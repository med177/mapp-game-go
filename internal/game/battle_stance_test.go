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
	if aggressiveGame.gs.PendingLandContact != nil {
		aggressiveGame.gs.ClearLandContact()
		aggressiveGame.moveArmyToSettlementWithStanceAndContactResolved("atk", "dst", "", combat.BattleStanceAggressive, false, true)
	}

	rand.Seed(7)
	defensiveGame := newGame()
	defensiveGame.moveArmyWithStance("atk", "dst", combat.BattleStanceDefensive)
	if defensiveGame.gs.PendingLandContact != nil {
		defensiveGame.gs.ClearLandContact()
		defensiveGame.moveArmyToSettlementWithStanceAndContactResolved("atk", "dst", "", combat.BattleStanceDefensive, false, true)
	}

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

func TestMoveArmyWithStanceDoesNotAutoMergeFriendlyArmy(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"},
		},
		Regions: map[world.RegionID]*world.Region{
			"src": {ID: "src", OwnerID: "p1", Neighbors: []world.RegionID{"dst"}},
			"dst": {ID: "dst", OwnerID: "p1", Neighbors: []world.RegionID{"src"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"moving": {
				ID: "moving", OwnerID: "p1", RegionID: "src", MovePoints: 2, MaxMovePoints: 2,
				Units: repeatedUnits("inf", 2, 100),
			},
			"stationed": {
				ID: "stationed", OwnerID: "p1", RegionID: "dst", MovePoints: 2, MaxMovePoints: 2,
				Units: repeatedUnits("inf", 3, 100),
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.moveArmyWithStance("moving", "dst", combat.BattleStanceBalanced)

	if len(gs.Armies) != 2 {
		t.Fatalf("dost bölgeye hareket otomatik birleşmemeli, ordu sayısı=%d", len(gs.Armies))
	}
	if got := gs.Armies["moving"]; got == nil || got.RegionID != "dst" || len(got.Units) != 2 {
		t.Fatalf("hareket eden ordu ayrı ve 2 birimle hedefte kalmalıydı, got=%+v", gs.Armies["moving"])
	}
	if got := gs.Armies["stationed"]; got == nil || len(got.Units) != 3 {
		t.Fatalf("hedefteki ordu otomatik olarak değişmemeliydi, got=%+v", gs.Armies["stationed"])
	}

	// Aynı iki ordu, oyuncu açıkça BİRLEŞTİR aksiyonunu verdiğinde birleşmeye devam eder.
	g.mergeArmiesManual("moving")
	if gs.Armies["moving"] != nil || len(gs.Armies["stationed"].Units) != 5 {
		t.Fatalf("manuel birleştirme çalışmamalıydı: moving=%+v stationed=%+v", gs.Armies["moving"], gs.Armies["stationed"])
	}
}
