package events

import (
	"encoding/json"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestEventRandomProbabilitySupportsLegacyJSONName(t *testing.T) {
	var events []*Event
	if err := json.Unmarshal([]byte(`[
		{"random_probability": 0.25},
		{"probability": 0.5},
		{"random_probability": 0, "probability": 0.75}
	]`), &events); err != nil {
		t.Fatalf("event JSON parse edilemedi: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("event sayısı yanlış: got %d, want 3", len(events))
	}
	if events[0].RandomProbability != 0.25 || events[1].RandomProbability != 0.5 || events[2].RandomProbability != 0 {
		t.Fatalf("yeni/eski alanlar beklenmedik çözümlendi: %v, %v, %v", events[0].RandomProbability, events[1].RandomProbability, events[2].RandomProbability)
	}

	data, err := json.Marshal(events[1])
	if err != nil {
		t.Fatalf("event JSON'a yazılamadı: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("yazılan event JSON parse edilemedi: %v", err)
	}
	if _, ok := fields["random_probability"]; !ok {
		t.Fatal("yazılan JSON'da random_probability alanı yok")
	}
	if _, ok := fields["probability"]; ok {
		t.Fatal("yazılan JSON'da eski probability alanı var")
	}
}

func TestRandomProbabilityOneTriggersEvent(t *testing.T) {
	e := &Event{ID: "certain_random_event", RandomProbability: 1}
	if got := Tick(&state.GameState{}, []*Event{e}); got != e {
		t.Fatal("random_probability=1 olan event tetiklenmedi")
	}
}

func TestChoiceIndicesForFactionFiltersOnlyRestrictedChoices(t *testing.T) {
	e := &Event{Choices: []Choice{
		{ID: "bolton", AvailableToFactions: []string{"bolton"}},
		{ID: "stark", AvailableToFactions: []string{"stark"}},
		{ID: "shared"},
	}}

	got := ChoiceIndicesForFaction(e, "stark")
	want := []int{1, 2}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("stark seçimleri yanlış filtrelendi: got %v, want %v", got, want)
	}

	if got := AutoChooseForFaction(&Event{Choices: []Choice{
		{AIWeight: 10, AvailableToFactions: []string{"bolton"}},
		{AIWeight: 2, AvailableToFactions: []string{"stark"}},
	}}, "stark"); got != 1 {
		t.Fatalf("faction seçimi AI ağırlığıyla filtrelenmedi: got %d, want 1", got)
	}
}

func TestVictoryConditionsRequireKeyRegionsAndMinimumOwnership(t *testing.T) {
	gs := &state.GameState{
		FiredEventIDs: map[string]bool{
			"flag:war_of_five_kings_active": true,
		},
		Regions: map[world.RegionID]*world.Region{
			"kings_landing": {OwnerID: "lannister"},
			"casterly_rock": {OwnerID: "lannister"},
			"riverlands":    {OwnerID: "lannister"},
		},
	}
	e := &Event{
		Target:        "all_factions",
		RequiresFlags: []string{"war_of_five_kings_active"},
		VictoryConditions: []FactionVictoryCondition{{
			FactionID:            "lannister",
			RequiredOwnedRegions: []world.RegionID{"kings_landing", "casterly_rock"},
			MinimumOwnedRegions:  4,
		}},
	}

	if eventConditionsSatisfied(gs, e) {
		t.Fatal("minimum sahiplik koşulu sağlanmadan zafer koşulu başarılı oldu")
	}

	gs.Regions["lannisport"] = &world.Region{OwnerID: "lannister"}
	if !eventConditionsSatisfied(gs, e) {
		t.Fatal("kilit bölgeler ve minimum sahiplik sağlandığında zafer koşulu başarısız oldu")
	}
}

func TestStateTriggeredVictoryEventIgnoresCalendarDate(t *testing.T) {
	gs := &state.GameState{
		Year:  320,
		Month: 6,
		FiredEventIDs: map[string]bool{
			"flag:war_of_five_kings_active": true,
		},
		Regions: map[world.RegionID]*world.Region{
			"kings_landing": {OwnerID: "lannister"},
		},
	}
	e := &Event{
		ID:             "war_resolution_state_triggered",
		Target:         "all_factions",
		StateTriggered: true,
		OneShot:        true,
		RequiresFlags:  []string{"war_of_five_kings_active"},
		VictoryConditions: []FactionVictoryCondition{{
			FactionID:            "lannister",
			RequiredOwnedRegions: []world.RegionID{"kings_landing"},
		}},
	}

	if got := Tick(gs, []*Event{e}); got != e {
		t.Fatal("state koşullu zafer eventi takvim tarihi olmadan tetiklenmedi")
	}
	if !gs.FiredEventIDs[e.ID] {
		t.Fatal("tek seferlik state koşullu event işaretlenmedi")
	}
}

func TestStateTriggeredEventUsesPriorityForSameTurnCandidates(t *testing.T) {
	gs := &state.GameState{}
	low := &Event{
		ID:                   "resolution_low",
		Target:               "all_factions",
		StateTriggered:       true,
		StateTriggerGroup:    "war_resolution",
		StateTriggerPriority: 10,
		VictoryConditions:    []FactionVictoryCondition{{FactionID: "lannister", MinimumOwnedRegions: 0}},
	}
	high := &Event{
		ID:                   "resolution_high",
		Target:               "all_factions",
		StateTriggered:       true,
		StateTriggerGroup:    "war_resolution",
		StateTriggerPriority: 20,
		VictoryConditions:    []FactionVictoryCondition{{FactionID: "stark", MinimumOwnedRegions: 0}},
	}

	if got := Tick(gs, []*Event{low, high}); got != high {
		t.Fatalf("state event önceliği kullanılmadı: got %v, want %v", got.ID, high.ID)
	}
}

func TestStateTriggeredFallbackRunsWhenVictoryCandidatesFail(t *testing.T) {
	gs := &state.GameState{
		FiredEventIDs: map[string]bool{
			"flag:war_of_five_kings_active": true,
		},
		Regions: map[world.RegionID]*world.Region{
			"kings_landing": {OwnerID: "stark"},
		},
	}
	winner := &Event{
		ID:                   "resolution_winner",
		Target:               "all_factions",
		StateTriggered:       true,
		StateTriggerGroup:    "war_resolution",
		StateTriggerPriority: 60,
		RequiresFlags:        []string{"war_of_five_kings_active"},
		VictoryConditions: []FactionVictoryCondition{{
			FactionID:            "lannister",
			RequiredOwnedRegions: []world.RegionID{"kings_landing"},
		}},
	}
	fallback := &Event{
		ID:                   "resolution_stalemate",
		Target:               "all_factions",
		StateTriggered:       true,
		StateTriggerGroup:    "war_resolution",
		StateTriggerPriority: 1,
		RequiresFlags:        []string{"war_of_five_kings_active"},
	}

	if got := Tick(gs, []*Event{winner, fallback}); got != fallback {
		t.Fatalf("fallback event seçilmedi: got %v, want %v", got.ID, fallback.ID)
	}
}

func TestRequiresOwnedRegionsAnyAcceptsRemainingAnchor(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"stark": {IsEliminated: false},
		},
		Regions: map[world.RegionID]*world.Region{
			"winterfell_region": {OwnerID: "bolton"},
			"the_neck":          {OwnerID: "stark"},
			"moat_cailin":       {OwnerID: "bolton"},
		},
	}
	e := &Event{
		Target:                  "specific_faction",
		AffectedFaction:         "stark",
		RequiresOwnedRegionsAny: []world.RegionID{"winterfell_region", "the_neck", "moat_cailin"},
	}
	if !eventConditionsSatisfied(gs, e) {
		t.Fatal("kalan tek kuzey dayanağı event koşulunu sağlamadı")
	}
}

func TestRequiresInactiveFaction(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"free_folk": {IsEliminated: false},
		},
	}
	e := &Event{Target: "all_factions", RequiresInactiveFactions: []string{"free_folk"}}
	if eventConditionsSatisfied(gs, e) {
		t.Fatal("aktif Özgür Halk varken eliminasyon koşulu sağlandı")
	}
	gs.Factions["free_folk"].IsEliminated = true
	if !eventConditionsSatisfied(gs, e) {
		t.Fatal("Özgür Halk elendiğinde eliminasyon koşulu sağlanmadı")
	}
}
