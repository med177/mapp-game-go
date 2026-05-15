package ai

import (
	"math/rand"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIHandlesPeaceWhenWarPressureIsHigh(t *testing.T) {
	gs := aiTestState()
	rel := gs.Relations[faction.RelationKey("ai_1", "ai_2")]
	rel.Stance = faction.StanceWar
	rel.Score = -100
	gs.Factions["ai_2"].Gold = 30

	aiHandleDiplomacy(gs, "ai_1")

	if rel.Stance != faction.StancePeace {
		t.Fatalf("AI barış aramalıydı, got=%s", rel.Stance)
	}
}

func TestAIStartsTradeOnHealthyPeace(t *testing.T) {
	gs := aiTestState()
	rel := gs.Relations[faction.RelationKey("ai_1", "ai_2")]
	rel.Stance = faction.StancePeace
	rel.Score = 5

	aiHandleDiplomacy(gs, "ai_1")

	if rel.Stance != faction.StanceTrade {
		t.Fatalf("AI ticaret başlatmalıydı, got=%s", rel.Stance)
	}
	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("iki yönlü ticaret rotası bekleniyordu, got=%d", len(gs.TradeRoutes))
	}
}

func TestCoalitionUsesDiplomacyEngine(t *testing.T) {
	gs := aiTestState()
	gs.PlayerFactionID = "player"
	gs.Regions["p2"] = &world.Region{ID: "p2", OwnerID: "player"}
	gs.Regions["p3"] = &world.Region{ID: "p3", OwnerID: "player"}
	gs.Regions["p4"] = &world.Region{ID: "p4", OwnerID: "player"}
	gs.Regions["p5"] = &world.Region{ID: "p5", OwnerID: "player"}
	gs.Regions["p6"] = &world.Region{ID: "p6", OwnerID: "player"}
	gs.Regions["p7"] = &world.Region{ID: "p7", OwnerID: "player"}
	gs.Regions["p8"] = &world.Region{ID: "p8", OwnerID: "player"}

	FormCoalitionAgainstPlayer(gs, "ai_1")

	playerRel := gs.Relations[faction.RelationKey("ai_1", "player")]
	if playerRel.Stance != faction.StanceWar {
		t.Fatalf("koalisyon oyuncuya savaş açmalıydı, got=%s", playerRel.Stance)
	}
	allyRel := gs.Relations[faction.RelationKey("ai_1", "ai_2")]
	if allyRel.Stance != faction.StanceAllied {
		t.Fatalf("koalisyon AI ittifakı kurmalıydı, got=%s", allyRel.Stance)
	}
}

func TestAIMoveArmyEmbarksIntoFriendlyTransportFleet(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		NextArmySeq:     1,
		Regions: map[world.RegionID]*world.Region{
			"ai_land":    {ID: "ai_land", OwnerID: "ai_1", Neighbors: []world.RegionID{"sea_1"}},
			"sea_1":      {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"ai_land", "enemy_land"}},
			"enemy_land": {ID: "enemy_land", OwnerID: "player", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_army": {
				ID:            "ai_army",
				OwnerID:       "ai_1",
				RegionID:      "ai_land",
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
			},
			"ai_fleet": {
				ID:            "ai_fleet",
				OwnerID:       "ai_1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Catholic},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -30, Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":       {ID: "inf", Embarkable: true, Attack: 10, Defense: 10, Morale: 50},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans},
		},
	}

	moveArmy(gs, gs.Armies["ai_army"])

	if _, ok := gs.Armies["ai_army"]; ok {
		t.Fatalf("AI kara ordusu embark sonrası haritadan kalkmalıydı")
	}
	fleet := gs.Armies["ai_fleet"]
	if fleet == nil || len(fleet.EmbarkedUnits) != 1 {
		t.Fatalf("AI filosunda embark birimi bekleniyordu, got=%+v", fleet)
	}
}

func TestAIMoveArmyDisembarksFromFleet(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		NextArmySeq:     10,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":   {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"ai_land"}},
			"ai_land": {ID: "ai_land", OwnerID: "ai_1", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_fleet": {
				ID:            "ai_fleet",
				OwnerID:       "ai_1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Catholic},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":       {ID: "inf", Embarkable: true, Attack: 10, Defense: 10, Morale: 50},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans},
		},
	}

	moveArmy(gs, gs.Armies["ai_fleet"])

	if len(gs.Armies["ai_fleet"].EmbarkedUnits) != 0 {
		t.Fatalf("AI çıkarma sonrası filonun embarked birimleri boş olmalı")
	}
	if _, ok := gs.Armies["army_ai_1_11"]; !ok {
		t.Fatalf("çıkarma sonrası yeni kara ordusu bekleniyordu")
	}
}

func TestAIMoveArmyDisembarksToEnemyCoastWhenAtWar(t *testing.T) {
	rand.Seed(1)
	gs := &state.GameState{
		PlayerFactionID: "player",
		NextArmySeq:     40,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":      {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"enemy_land"}},
			"enemy_land": {ID: "enemy_land", OwnerID: "player", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_fleet": {
				ID:            "ai_fleet",
				OwnerID:       "ai_1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "elite", CurrentHP: 100}, {TypeID: "elite", CurrentHP: 100}, {TypeID: "elite", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -70, Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"elite":     {ID: "elite", Embarkable: true, Attack: 120, Defense: 90, Morale: 90},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans},
		},
	}

	moveArmy(gs, gs.Armies["ai_fleet"])

	if gs.Regions["enemy_land"].OwnerID != "ai_1" {
		t.Fatalf("savaşta düşman kıyı çıkarma sonrası sahiplik değişmeli, got=%s", gs.Regions["enemy_land"].OwnerID)
	}
	if _, ok := gs.Armies["army_ai_1_41"]; !ok {
		t.Fatalf("savaşta düşman kıyı çıkarma sonrası kara ordusu oluşmalı")
	}
}

func TestAIMoveArmyDoesNotDisembarkToEnemyCoastAtPeace(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		NextArmySeq:     50,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":      {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"enemy_land"}},
			"enemy_land": {ID: "enemy_land", OwnerID: "player", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_fleet": {
				ID:            "ai_fleet",
				OwnerID:       "ai_1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: 5, Stance: faction.StancePeace},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":       {ID: "inf", Embarkable: true, Attack: 10, Defense: 10, Morale: 50},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans},
		},
	}

	moveArmy(gs, gs.Armies["ai_fleet"])

	if gs.Regions["enemy_land"].OwnerID != "player" {
		t.Fatalf("barışta düşman kıyıya çıkarma olmamalı, got=%s", gs.Regions["enemy_land"].OwnerID)
	}
	if _, ok := gs.Armies["army_ai_1_51"]; ok {
		t.Fatalf("barışta düşman kıyıya yeni kara ordusu oluşmamalı")
	}
	if len(gs.Armies["ai_fleet"].EmbarkedUnits) != 1 {
		t.Fatalf("barışta çıkarma olmamalı, cargo korunmalı")
	}
}

func aiTestState() *state.GameState {
	return &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic, Grain: 100, Gold: 100},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Catholic, Grain: 100, Gold: 100},
			"ai_2":   {ID: "ai_2", NameTR: "AI 2", Religion: religion.Catholic, Grain: 100, Gold: 100},
		},
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"p1": {ID: "p1", OwnerID: "player", TradeCapacity: 3},
			"a1": {ID: "a1", OwnerID: "ai_1", TradeCapacity: 3},
			"b1": {ID: "b1", OwnerID: "ai_2", TradeCapacity: 3},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "ai_2"):   {FactionA: "ai_1", FactionB: "ai_2", Score: 25, Stance: faction.StancePeace},
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -10, Stance: faction.StancePeace},
			faction.RelationKey("ai_2", "player"): {FactionA: "ai_2", FactionB: "player", Score: -10, Stance: faction.StancePeace},
		},
		TradeRoutes: []*economy.TradeRoute{},
		Armies: map[army.ArmyID]*army.Army{
			"ai1_army": {ID: "ai1_army", OwnerID: "ai_1", RegionID: "a1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"ai2_army": {ID: "ai2_army", OwnerID: "ai_2", RegionID: "b1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 12, Defense: 10, Morale: 60},
		},
	}
}
