package diplomacy

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestIsRelationshipNotificationOnlyMatchesNonBlockingOffers(t *testing.T) {
	for _, action := range []Action{ActionImproveRelations, ActionSendGift} {
		if !IsRelationshipNotification(action) {
			t.Fatalf("%q ilişki bildirimi olarak işaretlenmedi", action)
		}
	}
	for _, action := range []Action{ActionProposePeace, ActionProposeTrade, ActionProposeAlliance, ActionOfferVassalization} {
		if IsRelationshipNotification(action) {
			t.Fatalf("%q oyuncu kararı isteyen teklif olmasına rağmen bildirim sayıldı", action)
		}
	}
}

func TestResolveAcceptedTradeOfferDoesNotReassessOffer(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":     {ID: "ai"},
			"player": {ID: "player"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "player"): {
				FactionA:  "ai",
				FactionB:  "player",
				ScoreAToB: -80,
				ScoreBToA: -80,
				Stance:    faction.StancePeace,
			},
		},
		DiplomaticOffers: []state.DiplomaticOffer{{
			FromFactionID: "ai",
			ToFactionID:   "player",
			Action:        string(ActionProposeTrade),
		}},
	}

	result := ResolveOffer(gs, 0, true)
	if !result.Applied {
		t.Fatalf("kabul edilen ticaret uygulanmadı: %#v", result)
	}
	rel := Relation(gs, "ai", "player")
	if rel == nil || rel.Stance != faction.StanceTrade {
		t.Fatalf("kabul edilen teklif yeniden değerlendirilip engellendi: %#v", rel)
	}
	if !HasTradeRouteBetween(gs, "ai", "player") {
		t.Fatal("kabul edilen ticaret için rota kurulmadı")
	}
}

func TestResolveAcceptedAllianceOfferIgnoresAllianceLimit(t *testing.T) {
	factions := map[faction.FactionID]*faction.Faction{
		"ai":     {ID: "ai"},
		"player": {ID: "player"},
	}
	relations := map[string]*faction.Relation{
		faction.RelationKey("ai", "player"): {
			FactionA: "ai", FactionB: "player", Stance: faction.StancePeace,
		},
	}
	for i := 0; i < 5; i++ {
		allyID := faction.FactionID("ally_" + string(rune('a'+i)))
		factions[allyID] = &faction.Faction{ID: allyID}
		relations[faction.RelationKey("ai", allyID)] = &faction.Relation{
			FactionA: "ai", FactionB: allyID, Stance: faction.StanceAllied,
		}
	}
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions:        factions,
		Relations:       relations,
		DiplomaticOffers: []state.DiplomaticOffer{{
			FromFactionID: "ai",
			ToFactionID:   "player",
			Action:        string(ActionProposeAlliance),
		}},
	}

	result := ResolveOffer(gs, 0, true)
	if !result.Applied || Relation(gs, "ai", "player").Stance != faction.StanceAllied {
		t.Fatalf("kabul edilen ittifak kotadan dolayı uygulanmadı: %#v", result)
	}
}

func TestResolveAcceptedWarJoinOfferSkipsWarCallAssessment(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"caller": {ID: "caller"},
			"enemy":  {ID: "enemy"},
			"player": {ID: "player"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("caller", "player"): {
				FactionA: "caller", FactionB: "player", Stance: faction.StanceAllied,
			},
			faction.RelationKey("caller", "enemy"): {
				FactionA: "caller", FactionB: "enemy", Stance: faction.StancePeace,
			},
			faction.RelationKey("player", "enemy"): {
				FactionA: "player", FactionB: "enemy", Stance: faction.StancePeace,
			},
		},
		DiplomaticOffers: []state.DiplomaticOffer{{
			FromFactionID:        "caller",
			ToFactionID:          "player",
			Action:               string(ActionJoinWarCall),
			WarDeclarerFactionID: "caller",
			WarEnemyFactionID:    "enemy",
		}},
	}

	result := ResolveOffer(gs, 0, true)
	if !result.Applied || Relation(gs, "player", "enemy").Stance != faction.StanceWar {
		t.Fatalf("kabul edilen savaş çağrısı assessment ile engellendi: %#v", result)
	}
}
