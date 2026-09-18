package ai

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
)

func relationshipRepairTestState(gold int) (*state.GameState, *faction.Relation) {
	actor := faction.FactionID("actor")
	target := faction.FactionID("target")
	rel := &faction.Relation{
		FactionA: actor,
		FactionB: target,
		Score:    20,
		Stance:   faction.StanceTrade,
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
	if actor.Gold == 2500 || rel.Score == 32 {
		t.Fatal("AI güvenli olmayan hediyeyi uyguladı")
	}
}

func TestAIGiftTreasuryGateAllowsGiftWithLargeReserve(t *testing.T) {
	actor := &faction.Faction{ID: "actor", Gold: 10000}
	if !aiGiftTreasurySafe(actor, 2500, nil) {
		t.Fatal("yeterli hazine rezervi olan AI için hediye gereksiz yere engellendi")
	}
}
