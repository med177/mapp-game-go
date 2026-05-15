package ai

import (
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
