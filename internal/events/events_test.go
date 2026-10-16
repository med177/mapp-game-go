package events

import (
	"encoding/json"
	"testing"

	"mapp-game-go/internal/army"
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

func TestTickAllReturnsMultipleEligibleEventsInOrder(t *testing.T) {
	gs := &state.GameState{Turn: 1, FiredEventIDs: map[string]bool{}}
	first := &Event{ID: "event_a", RandomProbability: 1, MinTurn: 1, OneShot: true}
	second := &Event{ID: "event_b", RandomProbability: 1, MinTurn: 1, OneShot: true}

	got := TickAll(gs, []*Event{first, second})
	if len(got) != 2 {
		t.Fatalf("uygun olay sayısı yanlış: got %d, want 2", len(got))
	}
	if got[0] != first || got[1] != second {
		t.Fatalf("olay sırası yanlış: got %v, want [%v %v]", []string{got[0].ID, got[1].ID}, first.ID, second.ID)
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

func TestAutoChooseForEventFactionRespectsChoiceAudience(t *testing.T) {
	e := &Event{
		Target:          "specific_faction",
		AffectedFaction: "nights_watch",
		Choices: []Choice{
			{ID: "free_folk_offer", AIWeight: 10, AvailableToFactions: []string{"free_folk"}},
			{ID: "watch_acceptance", AIWeight: 5, AvailableToFactions: []string{"nights_watch"}},
		},
	}

	if got := AutoChooseForEventFaction(e); got != 1 {
		t.Fatalf("hedef fraksiyon kendi seçimine göre karar vermedi: got %d, want 1", got)
	}
}

func TestChoiceCanEnableFactionTrade(t *testing.T) {
	tradeDisabled := true
	canTrade := false
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"free_folk": {ID: "free_folk", TradeDisabled: tradeDisabled},
		},
	}
	e := &Event{
		Target:          "specific_faction",
		AffectedFaction: "free_folk",
		Choices: []Choice{{Effect: Effect{
			Target:          "specific_faction",
			AffectedFaction: "free_folk",
			TradeDisabled:   &canTrade,
		}}},
	}

	if _, ok := ApplyChoice(gs, e, 0); !ok {
		t.Fatal("yerleşim seçimi uygulanmadı")
	}
	if gs.Factions["free_folk"].TradeDisabled {
		t.Fatal("yerleşimden sonra Free Folk'un ticareti hâlâ kapalı")
	}
}

func TestEventCanToggleFactionDiplomacyDisabled(t *testing.T) {
	var enableDiplomacy Event
	if err := json.Unmarshal([]byte(`{
		"target":"specific_faction",
		"affected_faction":"others",
		"diplomacy_disabled":false
	}`), &enableDiplomacy); err != nil {
		t.Fatalf("event JSON'u çözülemedi: %v", err)
	}
	if enableDiplomacy.DiplomacyDisabled == nil || *enableDiplomacy.DiplomacyDisabled {
		t.Fatal("event kökündeki diplomacy_disabled=false değeri korunmadı")
	}

	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"others": {ID: "others", DiplomacyDisabled: true},
		},
	}
	Apply(gs, &enableDiplomacy)
	if gs.Factions["others"].DiplomacyDisabled {
		t.Fatal("event diplomacy_disabled=false değerini uygulamadı")
	}

	var disableDiplomacy Event
	if err := json.Unmarshal([]byte(`{
		"target":"specific_faction",
		"affected_faction":"others",
		"choices":[{"label_tr":"Kilitle","desc_tr":"","effect":{"diplomacy_disabled":true}}]
	}`), &disableDiplomacy); err != nil {
		t.Fatalf("choice event JSON'u çözülemedi: %v", err)
	}
	if _, ok := ApplyChoice(gs, &disableDiplomacy, 0); !ok {
		t.Fatal("diplomasi kilidi choice etkisi uygulanmadı")
	}
	if !gs.Factions["others"].DiplomacyDisabled {
		t.Fatal("choice etkisindeki diplomacy_disabled=true uygulanmadı")
	}
}

func TestEventCanStartWarWithDiplomacyDisabledFaction(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"others":       {ID: "others", DiplomacyDisabled: true},
			"nights_watch": {ID: "nights_watch"},
		},
	}
	e := &Event{Choices: []Choice{
		{Effect: Effect{
			Target:          "specific_faction",
			AffectedFaction: "others",
			Relations:       []RelationEffect{{FactionID: "nights_watch", Stance: string(faction.StanceWar)}},
		}},
		{Effect: Effect{
			Target:          "specific_faction",
			AffectedFaction: "others",
			Relations:       []RelationEffect{{FactionID: "stark", Stance: string(faction.StanceAllied)}},
		}},
	}}

	if _, ok := ApplyChoice(gs, e, 0); !ok {
		t.Fatal("event seçimi uygulanmadı")
	}
	relation := gs.Relations[faction.RelationKey("others", "nights_watch")]
	if relation == nil || relation.Stance != faction.StanceWar {
		t.Fatalf("event üzerinden savaş ilişkisi kurulmadı: %+v", relation)
	}
	if _, ok := ApplyChoice(gs, e, 1); !ok {
		t.Fatal("ikinci event seçimi uygulanmadı")
	}
	if relation := gs.Relations[faction.RelationKey("others", "stark")]; relation != nil {
		t.Fatalf("event üzerinden diplomasiye kapalı faction için ittifak ilişkisi kuruldu: %+v", relation)
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

func TestOneShotHistoricalEventDoesNotFireAgain(t *testing.T) {
	gs := &state.GameState{
		Year:          319,
		Month:         1,
		FiredEventIDs: map[string]bool{},
		Factions: map[faction.FactionID]*faction.Faction{
			"baratheon_stormsend": {ID: "baratheon_stormsend"},
		},
	}
	e := &Event{
		ID:              "renly_southern_alliance_297",
		Target:          "specific_faction",
		AffectedFaction: "baratheon_stormsend",
		HistoricalYear:  297,
		OneShot:         true,
	}

	if got := Tick(gs, []*Event{e}); got != e {
		t.Fatal("geçmiş tarihli event uygun olduğunda ilk kez tetiklenmedi")
	}
	if got := Tick(gs, []*Event{e}); got != nil {
		t.Fatalf("tek seferlik event tekrar tetiklendi: %s", got.ID)
	}
}

func TestRequiresAliveCommanderRejectsDeadOrExpiredCommander(t *testing.T) {
	gs := &state.GameState{
		Year:          299,
		Month:         5,
		FiredEventIDs: map[string]bool{},
		Factions: map[faction.FactionID]*faction.Faction{
			"baratheon_stormsend": {ID: "baratheon_stormsend"},
		},
		Commanders: map[string]*army.Commander{
			"cmd_renly_baratheon": {ID: "cmd_renly_baratheon", OwnerID: "baratheon_stormsend", StartYear: 290, EndYear: 300},
		},
		CommanderTemplates: map[string][]*army.Commander{
			"baratheon_stormsend": {
				{ID: "cmd_renly_baratheon", StartYear: 290, EndYear: 300},
			},
		},
	}
	e := &Event{
		Target:                  "specific_faction",
		AffectedFaction:         "baratheon_stormsend",
		RequiresAliveCommanders: []string{"cmd_renly_baratheon"},
	}

	if !ConditionsMet(gs, e) {
		t.Fatal("komutanın son aktif yılında yaşam koşulu sağlanmadı")
	}
	if !gs.MarkCommanderDead("cmd_renly_baratheon") {
		t.Fatal("komutan ölümü state'e yazılamadı")
	}
	if ConditionsMet(gs, e) {
		t.Fatal("ölüm state'i varken yaşam koşulu sağlandı")
	}
	if !gs.IsCommanderDead("cmd_renly_baratheon") || gs.Commanders["cmd_renly_baratheon"] != nil {
		t.Fatal("ölüm state'i runtime commander kaydını kaldırmadı")
	}

	expiredGS := &state.GameState{
		Year:          300,
		Month:         5,
		FiredEventIDs: map[string]bool{},
		Factions:      gs.Factions,
		Commanders: map[string]*army.Commander{
			"cmd_renly_baratheon": {ID: "cmd_renly_baratheon", OwnerID: "baratheon_stormsend", StartYear: 290, EndYear: 300},
		},
		CommanderTemplates: gs.CommanderTemplates,
	}
	if ConditionsMet(expiredGS, e) {
		t.Fatal("end_year sınırında doğal ölüm yaşam koşulunu kapatmadı")
	}
}

func TestChoiceCommanderDeathsRecordAndGateCommander(t *testing.T) {
	commander := &army.Commander{ID: "cmd_ned_stark", OwnerID: "stark", StartYear: 280, EndYear: 299}
	gs := &state.GameState{
		Year:          298,
		FiredEventIDs: map[string]bool{},
		Factions: map[faction.FactionID]*faction.Faction{
			"stark": {ID: "stark"},
		},
		Commanders:         map[string]*army.Commander{commander.ID: commander},
		CommanderTemplates: map[string][]*army.Commander{"stark": {commander}},
	}
	deathEvent := &Event{
		ID:              "ned_execution",
		Target:          "specific_faction",
		AffectedFaction: "stark",
		CommanderDeaths: []string{commander.ID},
		Choices: []Choice{{
			LabelTR: "İnfaz et",
		}},
	}
	aliveEvent := &Event{
		Target:                  "specific_faction",
		AffectedFaction:         "stark",
		RequiresAliveCommanders: []string{commander.ID},
	}

	if !ConditionsMet(gs, aliveEvent) {
		t.Fatal("ölümden önce alive commander koşulu sağlanmadı")
	}
	if _, ok := ApplyChoice(gs, deathEvent, 0); !ok {
		t.Fatal("komutan ölüm event effect’i uygulanmadı")
	}
	if !gs.DeadCommanderIDs[commander.ID] || gs.Commanders[commander.ID] != nil {
		t.Fatal("ölüm effect’i state’e yazılmadı veya runtime commander kaldırılmadı")
	}
	if ConditionsMet(gs, aliveEvent) {
		t.Fatal("ölüm sonrası requires_alive_commanders koşulu sağlandı")
	}
}

func TestChoiceCommanderCaptureMakesCharacterPassive(t *testing.T) {
	commander := &army.Commander{
		ID: "cmd_tyrion_lannister", Type: army.CharacterTypeCommander, OwnerID: "lannister",
		AssignedCouncilRole: state.CouncilRoleTreasurer, ActiveMissionID: "mission",
		ActiveMissionTurnsRemaining: 2,
	}
	gs := &state.GameState{
		Year:       298,
		Factions:   map[faction.FactionID]*faction.Faction{"lannister": {ID: "lannister"}},
		Commanders: map[string]*army.Commander{commander.ID: commander},
	}
	e := &Event{
		ID: "capture_tyrion",
		Choices: []Choice{{
			LabelTR: "Tutsak et",
			Effect: Effect{CommanderCaptures: []CommanderCaptureEffect{{
				CommanderID: "cmd_tyrion_lannister", HoldingSettlementID: "the_eyrie_main",
			}}},
		}},
	}

	if _, ok := ApplyChoice(gs, e, 0); !ok {
		t.Fatal("tutsaklık seçeneği uygulanmadı")
	}
	if commander.CaptiveAtSettlementID != "the_eyrie_main" {
		t.Fatalf("tutsaklık konumu = %q, want the_eyrie_main", commander.CaptiveAtSettlementID)
	}
	if commander.AssignedCouncilRole != "" || commander.ActiveMissionID != "" ||
		commander.ActiveMissionTurnsRemaining != 0 {
		t.Fatalf("tutsak karakterin görevleri temizlenmedi: %+v", commander)
	}
	if got := gs.CommanderUnavailableReason(commander); got != "Tutsak" {
		t.Fatalf("karakterin pasiflik nedeni = %q, want Tutsak", got)
	}
	if got := gs.CharacterRoster("lannister"); len(got) != 1 || got[0] != commander {
		t.Fatalf("tutsak karakter hane roster'ında görünmüyor: %#v", got)
	}
}

func TestCommanderExileEffectPreservesAliveCommanderRequirement(t *testing.T) {
	commander := &army.Commander{ID: "cmd_tyrion_lannister", OwnerID: "lannister", StartYear: 298, EndYear: 330}
	gs := &state.GameState{
		Year:          300,
		FiredEventIDs: map[string]bool{},
		Factions: map[faction.FactionID]*faction.Faction{
			"lannister": {ID: "lannister"},
		},
		Commanders:         map[string]*army.Commander{commander.ID: commander},
		CommanderTemplates: map[string][]*army.Commander{"lannister": {commander}},
	}
	e := &Event{
		Target:                  "specific_faction",
		AffectedFaction:         "lannister",
		CommanderExiles:         []string{commander.ID},
		RequiresAliveCommanders: []string{commander.ID},
	}

	Apply(gs, e)
	if !gs.ExiledCommanderIDs[commander.ID] || gs.IsCommanderDead(commander.ID) {
		t.Fatal("event exile state'i ölümü veya sürgünü doğru ayıramadı")
	}
	if !ConditionsMet(gs, e) {
		t.Fatal("hayatta olan sürgün komutan alive requirement'ı sağlayamadı")
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

func TestPlayerEventCodexShowsOnlyDirectPlayerTargets(t *testing.T) {
	gs := &state.GameState{PlayerFactionID: "free_folk"}
	var globalEvent Event
	if err := json.Unmarshal([]byte(`{"target":"all_factions"}`), &globalEvent); err != nil {
		t.Fatalf("global event JSON'dan okunamadı: %v", err)
	}
	if !IsPlayerRelevant(gs, &globalEvent) {
		t.Fatal("global event normal oyuncu bildirim uygunluğunu kaybetti")
	}
	if IsPlayerEventCodexRelevant(gs, &globalEvent) {
		t.Fatal("all_factions event'i özel Kodex/HUD listesine girdi")
	}
	if !IsPlayerEventCodexRelevant(gs, &Event{Target: "specific_faction", AffectedFaction: "free_folk"}) {
		t.Fatal("doğrudan oyuncu faction'ına atanmış event özel listede görünmedi")
	}
	if !IsPlayerEventCodexRelevant(gs, &Event{Target: "player_faction"}) {
		t.Fatal("player_faction event'i özel listede görünmedi")
	}
	if IsPlayerEventCodexRelevant(gs, &Event{
		Target:               "specific_faction",
		AffectedFaction:      "baratheon_king",
		PlayerChoiceFactions: []string{"free_folk"},
	}) {
		t.Fatal("başka faction'a atanmış, oyuncuya seçim veren event özel listeye girdi")
	}
}
