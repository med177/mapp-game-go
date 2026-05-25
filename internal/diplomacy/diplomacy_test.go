package diplomacy

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestProposePeaceRejectedOutsideWar(t *testing.T) {
	gs := testGameState()

	result := Execute(gs, "a", "b", ActionProposePeace)

	if result.Accepted || result.Applied {
		t.Fatalf("barış teklifi savaş dışındayken uygulanmamalı: %+v", result)
	}
}

func TestProposeAllianceRejectedOnLowScore(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 10

	result := Execute(gs, "a", "b", ActionProposeAlliance)

	if result.Accepted || result.Applied {
		t.Fatalf("düşük skorlu ittifak reddedilmeliydi: %+v", result)
	}
	if rel.Stance != faction.StancePeace {
		t.Fatalf("stance peace kalmalı, got=%s", rel.Stance)
	}
}

func TestProposeTradeRejectedDuringWar(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceWar
	rel.Score = -80

	result := Execute(gs, "a", "b", ActionProposeTrade)

	if result.Accepted || result.Applied {
		t.Fatalf("savaşta ticaret reddedilmeliydi: %+v", result)
	}
}

func TestTradeCreatesUniqueRoutesAndWarRemovesThem(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Score = 5

	result := Execute(gs, "a", "b", ActionProposeTrade)
	if !result.Accepted || !result.Applied {
		t.Fatalf("ticaret kabul edilmeliydi: %+v", result)
	}
	if rel.Stance != faction.StanceTrade {
		t.Fatalf("stance trade olmalı, got=%s", rel.Stance)
	}
	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("iki yönlü 2 rota bekleniyordu, got=%d", len(gs.TradeRoutes))
	}

	result = Execute(gs, "a", "b", ActionProposeTrade)
	if result.Accepted || result.Applied {
		t.Fatalf("tekrar ticaret aynı rotaları çoğaltmamalı: %+v", result)
	}
	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("rota sayısı 2 kalmalı, got=%d", len(gs.TradeRoutes))
	}

	result = Execute(gs, "a", "b", ActionDeclareWar)
	if !result.Accepted || !result.Applied {
		t.Fatalf("savaş ilanı uygulanmalı: %+v", result)
	}
	if len(gs.TradeRoutes) != 0 {
		t.Fatalf("savaşta ticaret yolları kapanmalı, got=%d", len(gs.TradeRoutes))
	}
}

func TestProposePeaceAcceptedUnderWarPressure(t *testing.T) {
	gs := testGameState()
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceWar
	rel.Score = -100
	gs.Factions["b"].Gold = 40

	result := Execute(gs, "a", "b", ActionProposePeace)

	if !result.Accepted || !result.Applied {
		t.Fatalf("yüksek savaş baskısında barış kabul edilmeliydi: %+v", result)
	}
	if rel.Stance != faction.StancePeace || rel.Score != -20 {
		t.Fatalf("barış sonrası ilişki güncellenmedi: %+v", rel)
	}
}

func TestQueueAndResolveOfferForPlayer(t *testing.T) {
	gs := testGameState()
	gs.PlayerFactionID = "b"
	rel := EnsureRelation(gs, "a", "b")
	rel.Stance = faction.StanceWar
	rel.Score = -100
	gs.Factions["b"].Gold = 40

	if !QueueOffer(gs, "a", "b", ActionProposePeace) {
		t.Fatal("teklif kuyruğa alınmalıydı")
	}
	if QueueOffer(gs, "a", "b", ActionProposePeace) {
		t.Fatal("aynı teklif ikinci kez kuyruğa alınmamalı")
	}
	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("tek bekleyen teklif bekleniyordu, got=%d", len(gs.DiplomaticOffers))
	}

	result := ResolveOffer(gs, 0, true)
	if !result.Accepted || !result.Applied {
		t.Fatalf("kabul edilen teklif uygulanmalıydı: %+v", result)
	}
	if len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("teklif kabul sonrası kuyruktan düşmeli, got=%d", len(gs.DiplomaticOffers))
	}
	if rel.Stance != faction.StancePeace {
		t.Fatalf("barış sonrası stance peace olmalı, got=%s", rel.Stance)
	}
}

func testGameState() *state.GameState {
	return &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", NameTR: "A", Religion: religion.Catholic, Grain: 120, Iron: 40, Spice: 10},
			"b": {ID: "b", NameTR: "B", Religion: religion.Catholic, Grain: 80, Cloth: 15},
		},
		Regions: map[world.RegionID]*world.Region{
			"a_cap": {ID: "a_cap", OwnerID: "a", TaxRate: 50, Satisfaction: 50, TradeCapacity: 4},
			"b_cap": {ID: "b_cap", OwnerID: "b", TaxRate: 50, Satisfaction: 50, TradeCapacity: 3},
		},
		Relations:   map[string]*faction.Relation{},
		TradeRoutes: []*economy.TradeRoute{},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "a", RegionID: "a_cap", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"b1": {ID: "b1", OwnerID: "b", RegionID: "b_cap", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 12, Defense: 10, Morale: 60},
		},
	}
}
