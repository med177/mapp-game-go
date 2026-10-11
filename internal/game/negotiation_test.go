package game

import (
	"testing"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
)

func TestAITurnCounterOfferIsSubmittedAndClosesOfferorTurn(t *testing.T) {
	gs := &state.GameState{
		Phase:           state.PhaseAITurn,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":     {ID: "ai", Grain: 100},
			"player": {ID: "player", Gold: 100},
		},
		DiplomaticOffers: []state.DiplomaticOffer{{
			FromFactionID:      "ai",
			ToFactionID:        "player",
			Action:             string(diplomacy.ActionProposeTransfer),
			RequestedTransfers: []state.DiplomaticTransfer{{Kind: "resource", ID: "gold", Amount: 90}},
			OfferedTransfers:   []state.DiplomaticTransfer{{Kind: "resource", ID: "grain", Amount: 1}},
		}},
	}
	diplomacy.EnsureRelation(gs, "ai", "player")
	g := &Game{
		gs: gs,
		aiTurn: &aiTurnState{
			order: []faction.FactionID{"ai", "another-ai"},
			index: 0,
		},
	}

	g.handleTransferNegotiationAction(render.InputAction{
		Kind:          render.ActionNegotiateTransfer,
		TargetFaction: "ai",
		OfferIndex:    0,
		CounterOffer:  true,
		NegotiationRequested: []state.DiplomaticTransfer{{
			Kind: "resource", ID: "grain", Amount: 1,
		}},
		NegotiationOffered: []state.DiplomaticTransfer{{
			Kind: "resource", ID: "gold", Amount: 90,
		}},
	})

	if len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("kabul edilen karşı tekliften sonra bekleyen teklif sayısı = %d, want 0", len(gs.DiplomaticOffers))
	}
	if gs.Factions["ai"].Grain != 99 || gs.Factions["player"].Grain != 1 ||
		gs.Factions["player"].Gold != 10 || gs.Factions["ai"].Gold != 90 {
		t.Fatalf("karşı teklif transferleri uygulanmadı: ai=%+v player=%+v", gs.Factions["ai"], gs.Factions["player"])
	}
	if g.aiTurn.index != 1 || g.aiTurn.stepper != nil || g.aiTurn.waitFrames != 0 {
		t.Fatalf("kabul edilen AI karşı teklifinden sonra AI sırası ilerlemedi: %+v", g.aiTurn)
	}
}
