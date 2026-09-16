package diplomacy

import (
	"strings"
	"testing"

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
