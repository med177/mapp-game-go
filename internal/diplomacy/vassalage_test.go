package diplomacy

import (
	"strings"
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestAnnexationTurnsRemainingMatchesAnnexationBlock(t *testing.T) {
	gs := &state.GameState{
		Turn:            16,
		PlayerFactionID: "overlord",
		Factions: map[faction.FactionID]*faction.Faction{
			"overlord": {ID: "overlord"},
			"vassal":   {ID: "vassal", OverlordID: "overlord", VassalizedTurn: 12, NameTR: "Vassal"},
		},
	}

	if got, want := AnnexationTurnsRemaining(gs, "overlord", "vassal"), 8; got != want {
		t.Fatalf("kalan ilhak turu: got=%d want=%d", got, want)
	}
	reason := ActionBlockReason(gs, "overlord", "vassal", ActionAnnexVassal)
	if !strings.HasSuffix(reason, "(8 tur kaldı).") {
		t.Fatalf("ilhak engel nedeni kalan turu içermiyor: %q", reason)
	}

	gs.Turn = 24
	if got := AnnexationTurnsRemaining(gs, "overlord", "vassal"); got != 0 {
		t.Fatalf("süre dolduktan sonra kalan ilhak turu: got=%d", got)
	}
	if reason := ActionBlockReason(gs, "overlord", "vassal", ActionAnnexVassal); reason != "" {
		t.Fatalf("süre dolduktan sonra ilhak engellendi: %s", reason)
	}
}

func TestRelationImprovementMessageNamesSenderAndTarget(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"sender": {ID: "sender", NameTR: "Gönderen Devlet"},
			"target": {ID: "target", NameTR: "Osmanoğulları Beyliği"},
		},
		Relations: map[string]*faction.Relation{},
	}

	result := applyRelationImprovement(gs, "sender", "target", 40, 8, 0, "diplomatik heyet")
	want := "Gönderen Devlet devleti, Osmanoğulları Beyliği için diplomatik heyet gönderdi. İlişki +8."
	if result.Message != want {
		t.Fatalf("ilişki geliştirme mesajı: got=%q want=%q", result.Message, want)
	}
}

func TestCancelTradeWithVassalDoesNotRequireOverlordDiplomacy(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "outside",
		Factions: map[faction.FactionID]*faction.Faction{
			"outside":  {ID: "outside", NameTR: "Dış Devlet"},
			"overlord": {ID: "overlord", NameTR: "Sahip Devlet"},
			"vassal":   {ID: "vassal", NameTR: "Vassal", OverlordID: "overlord"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("outside", "vassal"): {
				FactionA: "outside", FactionB: "vassal", Stance: faction.StanceTrade,
			},
		},
		TradeRoutes: []*economy.TradeRoute{{FromFactionID: "outside", ToFactionID: "vassal"}},
	}

	if reason := ActionBlockReason(gs, "outside", "vassal", ActionCancelTrade); reason != "" {
		t.Fatalf("vassalla ticareti bitirme engellendi: %s", reason)
	}
	result := Execute(gs, "outside", "vassal", ActionCancelTrade)
	if !result.Applied || HasTradeRouteBetween(gs, "outside", "vassal") {
		t.Fatalf("vassalla ticaret iptali uygulanmadı: applied=%v routes=%d", result.Applied, len(gs.TradeRoutes))
	}
}

func TestNormalizeVassalagePreservesExternalVassalTrade(t *testing.T) {
	gs := &state.GameState{
		Turn:            6,
		PlayerFactionID: "venice",
		Factions: map[faction.FactionID]*faction.Faction{
			"venice":   {ID: "venice", NameTR: "Venedik"},
			"overlord": {ID: "overlord", NameTR: "Sahip Devlet"},
			"vassal":   {ID: "vassal", NameTR: "Flandre", OverlordID: "overlord", VassalizedTurn: 1},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("venice", "vassal"): {
				FactionA: "venice", FactionB: "vassal", Score: 30, Stance: faction.StanceTrade,
			},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "venice", ToFactionID: "vassal"},
			{FromFactionID: "vassal", ToFactionID: "venice"},
		},
	}

	NormalizeVassalage(gs)

	rel := Relation(gs, "venice", "vassal")
	if rel == nil || rel.Stance != faction.StanceTrade {
		t.Fatalf("vassal dış ticaret ilişkisi normalize edildi: %#v", rel)
	}
	if !HasTradeRouteBetween(gs, "venice", "vassal") {
		t.Fatal("vassal dış ticaret rotası normalize edilirken silindi")
	}
}
