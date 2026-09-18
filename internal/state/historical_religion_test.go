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
				FactionA: "safavid", FactionB: "shia_state", Score: 10,
			},
			faction.RelationKey("safavid", "sunni_state"): {
				FactionA: "safavid", FactionB: "sunni_state", Score: 20,
			},
			faction.RelationKey("safavid", "catholic_state"): {
				FactionA: "safavid", FactionB: "catholic_state", Score: 10,
			},
		},
	}

	reports := s.ApplyHistoricalFactionChanges()
	if len(reports) != 1 || !reports[0].ReligionChanged || reports[0].Religion != religion.Shia {
		t.Fatalf("historical change report = %+v, want one Shia religion change", reports)
	}
	if got := s.Relations[faction.RelationKey("safavid", "shia_state")].Score; got != 40 {
		t.Fatalf("same-new-religion relation = %d, want 40", got)
	}
	if got := s.Relations[faction.RelationKey("safavid", "sunni_state")].Score; got != -20 {
		t.Fatalf("old-religion relation = %d, want -20", got)
	}
	if got := s.Relations[faction.RelationKey("safavid", "catholic_state")].Score; got != 10 {
		t.Fatalf("unrelated-religion relation = %d, want 10", got)
	}

	if reports := s.ApplyHistoricalFactionChanges(); len(reports) != 0 {
		t.Fatalf("repeated historical change reports = %+v, want none", reports)
	}
	if got := s.Relations[faction.RelationKey("safavid", "shia_state")].Score; got != 40 {
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
				FactionA: "safavid", FactionB: "shia_state", Score: 90,
			},
		},
	}

	s.ApplyHistoricalFactionChanges()
	if got := s.Relations[faction.RelationKey("safavid", "shia_state")].Score; got != 100 {
		t.Fatalf("relation score = %d, want upper clamp 100", got)
	}
}
