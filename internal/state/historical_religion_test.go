package state

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
)

func TestApplyHistoricalFactionChangesAdjustsRelationsOnceForReligionChange(t *testing.T) {
	s := &GameState{
		Year: 1501,
		Factions: map[faction.FactionID]*faction.Faction{
			"safavid": {
				ID:       "safavid",
				Religion: religion.Sunni,
				HistoricalChanges: []faction.HistoricalChange{
					{Year: 1501, Religion: religion.Shia},
				},
			},
			"shia_state":     {ID: "shia_state", Religion: religion.Shia},
			"sunni_state":    {ID: "sunni_state", Religion: religion.Sunni},
			"catholic_state": {ID: "catholic_state", Religion: religion.Catholic},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("safavid", "shia_state"): {
				FactionA: "safavid", FactionB: "shia_state", ScoreAToB: 10, ScoreBToA: 20,
			},
			faction.RelationKey("safavid", "sunni_state"): {
				FactionA: "safavid", FactionB: "sunni_state", ScoreAToB: 20, ScoreBToA: 30,
			},
			faction.RelationKey("safavid", "catholic_state"): {
				FactionA: "safavid", FactionB: "catholic_state", ScoreAToB: 10, ScoreBToA: 15,
			},
		},
	}

	reports := s.ApplyHistoricalFactionChanges()
	if len(reports) != 1 || !reports[0].ReligionChanged || reports[0].Religion != religion.Shia {
		t.Fatalf("historical change report = %+v, want one Shia religion change", reports)
	}
	if got := s.Relations[faction.RelationKey("safavid", "shia_state")].ScoreFrom("safavid"); got != 40 {
		t.Fatalf("same-new-religion relation = %d, want 40", got)
	}
	if got := s.Relations[faction.RelationKey("safavid", "shia_state")].ScoreFrom("shia_state"); got != 50 {
		t.Fatalf("karşı yön aynı-din ilişkisi = %d, want 50", got)
	}
	if got := s.Relations[faction.RelationKey("safavid", "sunni_state")].ScoreFrom("safavid"); got != -20 {
		t.Fatalf("old-religion relation = %d, want -20", got)
	}
	if got := s.Relations[faction.RelationKey("safavid", "sunni_state")].ScoreFrom("sunni_state"); got != -10 {
		t.Fatalf("karşı yön eski-din ilişkisi = %d, want -10", got)
	}
	if got := s.Relations[faction.RelationKey("safavid", "catholic_state")].ScoreFrom("safavid"); got != 10 {
		t.Fatalf("unrelated-religion relation = %d, want 10", got)
	}

	if reports := s.ApplyHistoricalFactionChanges(); len(reports) != 0 {
		t.Fatalf("repeated historical change reports = %+v, want none", reports)
	}
	if got := s.Relations[faction.RelationKey("safavid", "shia_state")].ScoreFrom("safavid"); got != 40 {
		t.Fatalf("relation was adjusted more than once: %d", got)
	}
}

func TestApplyHistoricalFactionChangesClampsReligionRelationAdjustment(t *testing.T) {
	s := &GameState{
		Year: 1501,
		Factions: map[faction.FactionID]*faction.Faction{
			"safavid": {
				ID: "safavid", Religion: religion.Sunni,
				HistoricalChanges: []faction.HistoricalChange{{Year: 1501, Religion: religion.Shia}},
			},
			"shia_state": {ID: "shia_state", Religion: religion.Shia},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("safavid", "shia_state"): {
				FactionA: "safavid", FactionB: "shia_state", ScoreAToB: 90, ScoreBToA: 90,
			},
		},
	}

	s.ApplyHistoricalFactionChanges()
	if got := s.Relations[faction.RelationKey("safavid", "shia_state")].ScoreFrom("safavid"); got != 100 {
		t.Fatalf("relation score = %d, want upper clamp 100", got)
	}
}
