package ai

import (
	"testing"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
)

func relationshipRepairTestState(gold int) (*state.GameState, *faction.Relation) {
	actor := faction.FactionID("actor")
	target := faction.FactionID("target")
	rel := &faction.Relation{
		FactionA:  actor,
		FactionB:  target,
		ScoreAToB: 20,
		ScoreBToA: 20,
		Stance:    faction.StanceTrade,
	}
	return &state.GameState{
		Turn: 1,
		DiplomacyConfig: scenario.DiplomacyConfig{
			RelationImprovementGoldCost: 1000,
			RelationImprovementBonus:    5,
			GiftGoldCost:                2500,
			GiftReceiverGold:            1000,
			GiftRelationBonus:           12,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			actor:  {ID: actor, Gold: gold},
			target: {ID: target, Gold: 5000},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(actor, target): rel,
		},
		TradeRoutes: []*economy.TradeRoute{{
			FromFactionID: string(actor),
			ToFactionID:   string(target),
			AmountPerTurn: 1,
		}},
	}, rel
}

func TestAIGiftTreasuryGateRejectsHalfTreasuryGift(t *testing.T) {
	gs, rel := relationshipRepairTestState(5000)
	actor := gs.Factions[rel.FactionA]

	if aiGiftTreasurySafe(actor, 2500, nil) {
		t.Fatal("2.500 altınlık hediye, 5.000 altın hazinede güvenli kabul edildi")
	}

	// Deterministik başarı zarı bu turu seçmeyebilir; iki durumda da güvenli
	// olmayan hediye uygulanmamalı, en fazla heyet uygulanmalıdır.
	aiHandleRelationshipRepairWithBudget(gs, rel.FactionA, rel.FactionB, rel, nil)
	if actor.Gold == 2500 || rel.ScoreFrom(rel.FactionA) == 32 {
		t.Fatal("AI güvenli olmayan hediyeyi uyguladı")
	}
}

func TestAIGiftTreasuryGateAllowsGiftWithLargeReserve(t *testing.T) {
	actor := &faction.Faction{ID: "actor", Gold: 10000}
	if !aiGiftTreasurySafe(actor, 2500, nil) {
		t.Fatal("yeterli hazine rezervi olan AI için hediye gereksiz yere engellendi")
	}
}

func resourceNegotiationTestState(relationScore int) *state.GameState {
	actor := faction.FactionID("actor")
	target := faction.FactionID("target")
	player := faction.FactionID("player")
	return &state.GameState{
		Turn:            4,
		DecisionSeed:    17,
		PlayerFactionID: player,
		Factions: map[faction.FactionID]*faction.Faction{
			actor:  {ID: actor, NameTR: "AI", Grain: 300},
			target: {ID: target, NameTR: "Hedef", Iron: 100},
			player: {ID: player, NameTR: "Oyuncu"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(actor, target): {
				FactionA: actor, FactionB: target,
				ScoreAToB: relationScore, ScoreBToA: relationScore,
				Stance: faction.StancePeace,
			},
			faction.RelationKey(actor, player): {
				FactionA: actor, FactionB: player,
				ScoreAToB: relationScore, ScoreBToA: relationScore,
				Stance: faction.StancePeace,
			},
		},
	}
}

func TestAIResourceNegotiationQueuesOfferForPlayer(t *testing.T) {
	gs := resourceNegotiationTestState(20)
	gs.Factions["player"].Iron = 100
	var steps []TurnStep

	if !aiHandleResourceNegotiationWithSteps(gs, "actor", economy.ResourceCost{Iron: 20}, &steps) {
		t.Fatal("AI oyuncuya kaynak pazarlığı önermedi")
	}
	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("bekleyen pazarlık sayısı = %d, 1 bekleniyordu", len(gs.DiplomaticOffers))
	}
	offer := gs.DiplomaticOffers[0]
	if offer.Action != string(diplomacy.ActionProposeTransfer) || offer.ToFactionID != "player" {
		t.Fatalf("beklenmeyen pazarlık teklifi: %+v", offer)
	}
	if len(offer.RequestedTransfers) != 1 || offer.RequestedTransfers[0].ID != "iron" || offer.RequestedTransfers[0].Amount != 20 {
		t.Fatalf("AI ihtiyacını karşılayan kaynak talebi oluşturmadı: %+v", offer.RequestedTransfers)
	}
	if len(offer.OfferedTransfers) != 1 || offer.OfferedTransfers[0].ID != "grain" || offer.OfferedTransfers[0].Amount <= 0 {
		t.Fatalf("AI fazla kaynağını karşı teklif olarak sunmadı: %+v", offer.OfferedTransfers)
	}
	if gs.Factions["actor"].Grain != 300 || gs.Factions["actor"].Iron != 0 {
		t.Fatal("oyuncu karar vermeden pazarlık kaynakları aktarıldı")
	}
	if len(steps) != 1 || steps[0].Kind != TurnStepDiplomacy {
		t.Fatalf("oyuncuya pazarlık turn step'i bildirilmedi: %+v", steps)
	}
}

func TestAIResourceNegotiationResolvesAIToAIOffer(t *testing.T) {
	gs := resourceNegotiationTestState(20)
	gs.Factions["target"].Grain = 0
	var steps []TurnStep

	if !aiHandleResourceNegotiationWithSteps(gs, "actor", economy.ResourceCost{Iron: 20}, &steps) {
		t.Fatal("AI-AI kaynak pazarlığı uygulanmadı")
	}
	if got := gs.Factions["actor"].Iron; got != 20 {
		t.Fatalf("AI'nin demir stoku = %d, 20 bekleniyordu", got)
	}
	if got := gs.Factions["target"].Iron; got != 80 {
		t.Fatalf("hedef demir stoku = %d, 80 bekleniyordu", got)
	}
	if gs.Factions["actor"].Grain >= 300 || gs.Factions["target"].Grain <= 0 {
		t.Fatalf("AI-AI pazarlığında tahıl aktarımı olmadı: actor=%d target=%d", gs.Factions["actor"].Grain, gs.Factions["target"].Grain)
	}
	if len(gs.DiplomaticOffers) != 0 || len(steps) != 1 {
		t.Fatalf("AI-AI pazarlığı kuyruğu/turn step'i beklenmedik: offers=%d steps=%d", len(gs.DiplomaticOffers), len(steps))
	}
}

func TestAIResourceNegotiationDoesNotOverpayWhenTargetRelationIsLow(t *testing.T) {
	gs := resourceNegotiationTestState(0)
	var steps []TurnStep

	if aiHandleResourceNegotiationWithSteps(gs, "actor", economy.ResourceCost{Iron: 20}, &steps) {
		t.Fatal("düşük ilişki puanında AI kabul olasılığı sağlamayan pahalı bir teklif yaptı")
	}
	if len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("reddedilmesi beklenen pazarlık kuyruğa girdi: %+v", gs.DiplomaticOffers)
	}
}
