package render

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
)

func TestCollectActiveWarSummariesGroupsCoalitionRelations(t *testing.T) {
	gs := &state.GameState{
		Turn: 20,
		Factions: map[faction.FactionID]*faction.Faction{
			"attacker":        {ID: "attacker", NameTR: "Saldıran"},
			"attacker_ally":   {ID: "attacker_ally", NameTR: "Saldıran Müttefiki"},
			"defender":        {ID: "defender", NameTR: "Savunan"},
			"defender_ally":   {ID: "defender_ally", NameTR: "Savunan Müttefiki"},
			"separate_target": {ID: "separate_target", NameTR: "Ayrı Hedef"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("attacker", "defender"): {
				FactionA: "attacker", FactionB: "defender", Stance: faction.StanceWar,
			},
			faction.RelationKey("attacker_ally", "defender"): {
				FactionA: "attacker_ally", FactionB: "defender", Stance: faction.StanceWar,
			},
			faction.RelationKey("attacker", "defender_ally"): {
				FactionA: "attacker", FactionB: "defender_ally", Stance: faction.StanceWar,
			},
			faction.RelationKey("attacker", "separate_target"): {
				FactionA: "attacker", FactionB: "separate_target", Stance: faction.StanceWar,
			},
			faction.RelationKey("attacker", "attacker_ally"): {
				FactionA: "attacker", FactionB: "attacker_ally", Stance: faction.StanceAllied,
			},
			faction.RelationKey("defender", "defender_ally"): {
				FactionA: "defender", FactionB: "defender_ally", Stance: faction.StanceAllied,
			},
		},
		WarLedgers: map[string]*state.WarLedger{
			faction.RelationKey("attacker", "defender"): {
				FactionA: "attacker", FactionB: "defender", DeclarerFactionID: "attacker", DefenderFactionID: "defender", StartedTurn: 8,
				CasualtiesA: 3, CasualtiesArmyA: 3, CasualtiesB: 1, CasualtiesFleetB: 1,
			},
			faction.RelationKey("attacker_ally", "defender"): {
				FactionA: "attacker_ally", FactionB: "defender", DeclarerFactionID: "attacker_ally", DefenderFactionID: "defender", StartedTurn: 9,
			},
			faction.RelationKey("attacker", "defender_ally"): {
				FactionA: "attacker", FactionB: "defender_ally", DeclarerFactionID: "attacker", DefenderFactionID: "defender_ally", StartedTurn: 9,
			},
			faction.RelationKey("attacker", "separate_target"): {
				FactionA: "attacker", FactionB: "separate_target", DeclarerFactionID: "attacker", DefenderFactionID: "separate_target", StartedTurn: 12,
			},
		},
	}

	wars := collectActiveWarSummaries(gs, nil)
	if got, want := len(wars), 2; got != want {
		t.Fatalf("gruplu savaş sayısı = %d, want %d", got, want)
	}

	var coalition ActiveWarSummary
	for _, war := range wars {
		if len(war.SideA.Participants) == 2 {
			coalition = war
			break
		}
	}
	if got, want := len(coalition.SideA.Participants), 2; got != want {
		t.Fatalf("saldıran taraf katılımcı sayısı = %d, want %d", got, want)
	}
	if got, want := len(coalition.SideB.Participants), 2; got != want {
		t.Fatalf("savunan taraf katılımcı sayısı = %d, want %d", got, want)
	}
	if coalition.SideA.Participants[0].FactionID != "attacker" || coalition.SideB.Participants[0].FactionID != "defender" {
		t.Fatalf("ana savaş tarafları ilk sırada değil: A=%s B=%s", coalition.SideA.Participants[0].FactionID, coalition.SideB.Participants[0].FactionID)
	}
	if coalition.SideA.ArmyCasualties != 3 || coalition.SideA.FleetCasualties != 0 || coalition.SideB.ArmyCasualties != 0 || coalition.SideB.FleetCasualties != 1 {
		t.Fatalf("koalisyon kayıpları = A ordu/filo %d/%d B ordu/filo %d/%d", coalition.SideA.ArmyCasualties, coalition.SideA.FleetCasualties, coalition.SideB.ArmyCasualties, coalition.SideB.FleetCasualties)
	}
	if got, want := countActiveWars(gs), 2; got != want {
		t.Fatalf("HUD savaş sayısı = %d, want %d", got, want)
	}
}

func TestActiveWarParticipantsKeepOriginalSidesFirstThenSortByPower(t *testing.T) {
	participants := []ActiveWarParticipant{
		{FactionID: "ally_low", Power: 20},
		{FactionID: "original", Power: 1},
		{FactionID: "ally_high", Power: 100},
	}

	sortActiveWarParticipants(participants, "original")
	got := []faction.FactionID{participants[0].FactionID, participants[1].FactionID, participants[2].FactionID}
	want := []faction.FactionID{"original", "ally_high", "ally_low"}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("katılımcı sırası = %v, want %v", got, want)
		}
	}
}

func TestCollectActiveWarSummariesIncludesVassalsOnTheirRealmSide(t *testing.T) {
	gs := &state.GameState{
		Turn: 12,
		Factions: map[faction.FactionID]*faction.Faction{
			"attacker":        {ID: "attacker", NameTR: "Saldıran"},
			"attacker_vassal": {ID: "attacker_vassal", NameTR: "Saldıran Vassalı", OverlordID: "attacker"},
			"nested_vassal":   {ID: "nested_vassal", NameTR: "Alt Vassal", OverlordID: "attacker_vassal"},
			"defender":        {ID: "defender", NameTR: "Savunan"},
			"defender_vassal": {ID: "defender_vassal", NameTR: "Savunan Vassalı", OverlordID: "defender"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("attacker", "defender"): {
				FactionA: "attacker", FactionB: "defender", Stance: faction.StanceWar,
			},
		},
	}

	wars := collectActiveWarSummaries(gs, nil)
	if got, want := len(wars), 1; got != want {
		t.Fatalf("vassallar savaşı çoğalttı: got %d want %d", got, want)
	}
	war := wars[0]
	if got, want := len(war.SideA.Participants), 3; got != want {
		t.Fatalf("saldıran realm katılımcıları = %d, want %d", got, want)
	}
	if got, want := len(war.SideB.Participants), 2; got != want {
		t.Fatalf("savunan realm katılımcıları = %d, want %d", got, want)
	}
	if war.SideA.Participants[0].FactionID != "attacker" || war.SideA.Participants[1].FactionID != "attacker_vassal" {
		t.Fatalf("saldıran taraf sırası = %v", war.SideA.Participants)
	}
}

func TestActiveWarVariableRowsShareDrawAndHitGeometry(t *testing.T) {
	wars := []ActiveWarSummary{
		{SideA: ActiveWarSide{Participants: []ActiveWarParticipant{{NameTR: "A"}, {NameTR: "B"}}}, SideB: ActiveWarSide{Participants: []ActiveWarParticipant{{NameTR: "C"}}}},
		{SideA: ActiveWarSide{Participants: []ActiveWarParticipant{{NameTR: "D"}}}, SideB: ActiveWarSide{Participants: []ActiveWarParticipant{{NameTR: "E"}}}},
	}
	viewport := gameui.Rect{X: 10, Y: 20, W: 500, H: 340}
	first := activeWarRowRectForWars(viewport, wars, 0, 0)
	second := activeWarRowRectForWars(viewport, wars, 1, 0)
	if first.Y+first.H > second.Y {
		t.Fatalf("değişken satırlar çakıştı: first bottom=%v second top=%v", first.Y+first.H, second.Y)
	}
	if got := activeWarMaxScrollForWars(wars, viewport); got != 0 {
		t.Fatalf("tüm satırlar sığarken scroll = %d", got)
	}
}

func TestActiveWarRowKeepsRightBorderInsideViewportAndTrimsBottomPad(t *testing.T) {
	wars := []ActiveWarSummary{{
		SideA: ActiveWarSide{Participants: []ActiveWarParticipant{{NameTR: "A"}}},
		SideB: ActiveWarSide{Participants: []ActiveWarParticipant{{NameTR: "B"}}},
	}}
	viewport := gameui.Rect{X: 10, Y: 20, W: 500, H: 200}
	row := activeWarRowRectForWars(viewport, wars, 0, 0)
	if got, want := row.X+row.W, viewport.X+viewport.W-activeWarRowRightInset; got != want {
		t.Fatalf("satır sağ sınırı viewport içine alınmadı: got=%v want=%v", got, want)
	}
	if got, want := row.H, activeWarSideHeaderH+activeWarParticipantH+activeWarRowBottomPad; got != want {
		t.Fatalf("satır alt boşluğu beklenenden farklı: got=%v want=%v", got, want)
	}
}

func TestActiveWarViewportKeepsPartialNextRowVisible(t *testing.T) {
	wars := []ActiveWarSummary{
		{SideA: ActiveWarSide{Participants: []ActiveWarParticipant{{NameTR: "A"}}}, SideB: ActiveWarSide{Participants: []ActiveWarParticipant{{NameTR: "B"}}}},
		{SideA: ActiveWarSide{Participants: []ActiveWarParticipant{{NameTR: "C"}}}, SideB: ActiveWarSide{Participants: []ActiveWarParticipant{{NameTR: "D"}}}},
	}
	viewport := gameui.Rect{X: 10, Y: 20, W: 500, H: 150}
	if got, want := activeWarVisibleRowsForWars(viewport, wars, 0), 2; got != want {
		t.Fatalf("viewport kısmi ikinci satırı göstermedi: got=%d want=%d", got, want)
	}
	first := activeWarRowRectForWars(viewport, wars, 0, 0)
	second := activeWarRowRectForWars(viewport, wars, 1, 0)
	if second.Y >= viewport.Y+viewport.H || second.Y+second.H <= viewport.Y+viewport.H {
		t.Fatalf("ikinci satır kısmi görünür değil: y=%v bottom=%v viewportBottom=%v", second.Y, second.Y+second.H, viewport.Y+viewport.H)
	}
	track := activeWarsScrollbarRect(viewport)
	if track.X <= first.X+first.W {
		t.Fatalf("scrollbar satır alanıyla çakışıyor: trackX=%v rowRight=%v", track.X, first.X+first.W)
	}
}
