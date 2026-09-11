package events

import (
	"encoding/json"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestTickCatchesUpHistoricalEventsSharingOneMonth(t *testing.T) {
	gs := &state.GameState{
		Year:          1335,
		Month:         12,
		MonthsPerTurn: 1,
		FiredEventIDs: map[string]bool{},
	}
	events := []*Event{
		{
			ID:              "first",
			HistoricalYear:  1335,
			HistoricalMonth: 12,
			OneShot:         true,
		},
		{
			ID:              "second",
			HistoricalYear:  1335,
			HistoricalMonth: 12,
			OneShot:         true,
		},
	}

	if got := Tick(gs, events); got == nil || got.ID != "first" {
		t.Fatalf("ilk tarihsel event seçilmedi: got=%v", got)
	}

	gs.Year = 1336
	gs.Month = 1
	if got := Tick(gs, events); got == nil || got.ID != "second" {
		t.Fatalf("aynı ayda bekleyen ikinci event yakalanmadı: got=%v", got)
	}
	if gs.FiredEventIDs[pendingHistoricalEventKey("second")] {
		t.Fatal("oynatılan event pending kuyruğunda kalmamalı")
	}
}

func TestTickDoesNotReplayOldHistoricalEvents(t *testing.T) {
	gs := &state.GameState{
		Year:          1501,
		Month:         1,
		MonthsPerTurn: 1,
		FiredEventIDs: map[string]bool{},
	}
	events := []*Event{{
		ID:              "old",
		HistoricalYear:  1337,
		HistoricalMonth: 5,
		OneShot:         true,
	}}

	if got := Tick(gs, events); got != nil {
		t.Fatalf("eski tarihli event geriye dönük tetiklenmemeli: got=%v", got)
	}
}

func TestApplyUsesRootEventDiplomacyAndEffects(t *testing.T) {
	var evt Event
	if err := json.Unmarshal([]byte(`{
		"id": "root_effects",
		"target": "specific_faction",
		"affected_faction": "england",
		"relation_delta_all": 4,
		"complete_techs": ["navigation"],
		"relations": [{"faction_id": "france", "stance": "war", "score_delta": -30}],
		"set_flags": ["root_effect_applied"]
	}`), &evt); err != nil {
		t.Fatalf("event parse edilemedi: %v", err)
	}

	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"england": {ID: "england", Research: faction.ResearchState{Completed: map[string]bool{}}},
			"france":  {ID: "france"},
		},
		Relations:     map[string]*faction.Relation{},
		FiredEventIDs: map[string]bool{},
	}

	Apply(gs, &evt)

	if relation := diplomacy.Relation(gs, "england", "france"); relation == nil || relation.Stance != faction.StanceWar || relation.Score != -26 {
		t.Fatalf("kök relations etkisi uygulanmadı: %+v", relation)
	}
	if !gs.Factions["england"].Research.Completed["navigation"] {
		t.Fatal("kök complete_techs etkisi uygulanmadı")
	}
	if !gs.FiredEventIDs["flag:root_effect_applied"] {
		t.Fatal("kök set_flags etkisi uygulanmadı")
	}
}

func TestApplyChoiceActivatesTradeNetworkModifier(t *testing.T) {
	gs := &state.GameState{
		FiredEventIDs: map[string]bool{},
	}
	evt := &Event{
		ID:     "spice_route_monopoly",
		Target: "all_factions",
		Choices: []Choice{{Effect: Effect{
			Target: "all_factions",
			TradeNetworkModifiers: []TradeNetworkModifierEffect{{
				ID:                 "portuguese_spice_route_loss",
				CenterIDs:          []string{"egypt", "basra"},
				RegionIDs:          []string{"egypt", "basra"},
				TradeIncomePercent: -35,
				SpicePercent:       -30,
			}},
		}}},
	}

	if _, ok := ApplyChoice(gs, evt, 0); !ok {
		t.Fatal("ticaret ağı modifier'ı içeren event seçimi uygulanmadı")
	}
	if got := gs.TradeNetworkIncomeModifier("egypt"); got != -35 {
		t.Fatalf("ticaret merkezi gelir modifier'ı state'e yazılmadı: got=%d", got)
	}
	if got := gs.RegionSpiceProductionModifier("basra"); got != -30 {
		t.Fatalf("baharat üretim modifier'ı state'e yazılmadı: got=%d", got)
	}
}

func TestApplyChoiceRevivesSuccessorFromEventMetadata(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"overlord":  {ID: "overlord"},
			"successor": {ID: "successor", IsEliminated: true},
		},
		Regions: map[world.RegionID]*world.Region{
			"successor_capital": {ID: "successor_capital", OwnerID: "overlord"},
		},
		Armies:        map[army.ArmyID]*army.Army{},
		Relations:     map[string]*faction.Relation{},
		FiredEventIDs: map[string]bool{},
	}
	evt := &Event{
		ID:              "successor_revival",
		Target:          "specific_faction",
		AffectedFaction: "overlord",
		Choices: []Choice{{Effect: Effect{
			Target:          "specific_faction",
			AffectedFaction: "overlord",
			SuccessorRevival: &SuccessorRevivalEffect{
				FactionID: "successor", RegionID: "successor_capital", Mode: "vassal",
			},
		}}},
	}

	if _, ok := ApplyChoice(gs, evt, 0); !ok {
		t.Fatal("ardıl diriltme event seçimi uygulanmadı")
	}
	if gs.Factions["successor"].IsEliminated || gs.Factions["successor"].OverlordID != "overlord" {
		t.Fatalf("ardıl event ile vassal olarak dirilmedi: %+v", gs.Factions["successor"])
	}
	if gs.Regions["successor_capital"].OwnerID != "successor" || len(gs.Armies) != 1 {
		t.Fatalf("ardıl toprağı/kuruluş ordusu oluşmadı: region=%s armies=%d", gs.Regions["successor_capital"].OwnerID, len(gs.Armies))
	}
}

func TestApplyRevivesMultipleSuccessorsFromRootEventMetadata(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ilkhanate":   {ID: "ilkhanate"},
			"jelayirids":  {ID: "jelayirids", IsEliminated: true},
			"muzaffarids": {ID: "muzaffarids", IsEliminated: true},
		},
		Regions: map[world.RegionID]*world.Region{
			"baghdad":        {ID: "baghdad", OwnerID: "ilkhanate"},
			"western_persia": {ID: "western_persia", OwnerID: "ilkhanate"},
		},
		Armies:        map[army.ArmyID]*army.Army{},
		Relations:     map[string]*faction.Relation{},
		FiredEventIDs: map[string]bool{},
	}
	evt := &Event{
		ID:              "ilkhanate_breakup",
		Target:          "specific_faction",
		AffectedFaction: "ilkhanate",
		SuccessorRevivals: []SuccessorRevivalEffect{
			{FactionID: "jelayirids", RegionID: "baghdad", Mode: "independent"},
			{FactionID: "muzaffarids", RegionID: "western_persia", Mode: "independent"},
		},
	}

	Apply(gs, evt)

	for _, id := range []faction.FactionID{"jelayirids", "muzaffarids"} {
		if gs.Factions[id].IsEliminated {
			t.Fatalf("çoklu ardıl diriltmede faction etkinleşmedi: %s", id)
		}
	}
	if gs.Regions["baghdad"].OwnerID != "jelayirids" || gs.Regions["western_persia"].OwnerID != "muzaffarids" {
		t.Fatalf("çoklu ardıl bölgeleri aktarılmadı: baghdad=%s western_persia=%s", gs.Regions["baghdad"].OwnerID, gs.Regions["western_persia"].OwnerID)
	}
}

func TestApplyCoalitionCreatesAlliedMembersAndWarAgainstOpponent(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"serbian_empire":   {ID: "serbian_empire"},
			"bulgarian_empire": {ID: "bulgarian_empire"},
			"wallachia_prince": {ID: "wallachia_prince"},
			"ottoman":          {ID: "ottoman"},
		},
		Relations:     map[string]*faction.Relation{},
		FiredEventIDs: map[string]bool{},
	}
	evt := &Event{
		ID:     "balkan_coalition",
		Target: "specific_faction",
		Coalition: &CoalitionEffect{
			Members:            []string{"serbian_empire", "bulgarian_empire", "wallachia_prince"},
			Opponents:          []string{"ottoman"},
			MemberStance:       "allied",
			OpponentStance:     "war",
			MemberScoreDelta:   40,
			OpponentScoreDelta: -70,
		},
	}

	Apply(gs, evt)

	if rel := diplomacy.Relation(gs, "serbian_empire", "bulgarian_empire"); rel == nil || rel.Stance != faction.StanceAllied {
		t.Fatalf("koalisyon üyeleri müttefik olmadı: %+v", rel)
	}
	if rel := diplomacy.Relation(gs, "serbian_empire", "ottoman"); rel == nil || rel.Stance != faction.StanceWar {
		t.Fatalf("koalisyon rakibe savaş açmadı: %+v", rel)
	}
}

func TestSpecificFactionEventRequiresActiveTargetFaction(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"england": {ID: "england", IsEliminated: true},
		},
		FiredEventIDs: map[string]bool{},
	}
	evt := &Event{
		ID:              "war_of_the_roses",
		Target:          "specific_faction",
		AffectedFaction: "england",
	}

	if ConditionsMet(gs, evt) {
		t.Fatal("elenmiş İngiltere için Güller Savaşı eventi uygun kabul edilmemeli")
	}
	if reasons := ConditionFailureReasons(gs, evt); len(reasons) != 1 || reasons[0] != "hedef faction aktif değil" {
		t.Fatalf("elenmiş hedef için açıklayıcı koşul hatası bekleniyordu: %v", reasons)
	}

	gs.Factions["england"].IsEliminated = false
	if !ConditionsMet(gs, evt) {
		t.Fatal("aktif İngiltere için event uygun kabul edilmeli")
	}
}

func TestSpecificFactionEventOffersChoiceToListedPlayerFaction(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "house_york",
		Factions: map[faction.FactionID]*faction.Faction{
			"england":    {ID: "england"},
			"house_york": {ID: "house_york"},
		},
	}
	evt := &Event{
		ID:                   "war_of_the_roses_resolution",
		Target:               "specific_faction",
		AffectedFaction:      "england",
		PlayerChoiceFactions: []string{"house_lancaster", "house_york"},
		Choices:              []Choice{{ID: "york_victory"}},
	}

	if !RequiresPlayerChoice(gs, evt) {
		t.Fatal("hanedan oyuncusu Güller Savaşı sonuç event'inde seçim yapabilmeli")
	}
}

func TestEventRequiresActiveSupportingFaction(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"castile_kingdom": {ID: "castile_kingdom"},
			"granada_emirate": {ID: "granada_emirate", IsEliminated: true},
		},
		FiredEventIDs: map[string]bool{},
	}
	evt := &Event{
		ID:                     "reconquista_granada_war",
		Target:                 "specific_faction",
		AffectedFaction:        "castile_kingdom",
		RequiresActiveFactions: []string{"granada_emirate"},
	}

	if ConditionsMet(gs, evt) {
		t.Fatal("Granada elenmişken Reconquista savaşı event'i uygun kabul edilmemeli")
	}
	gs.Factions["granada_emirate"].IsEliminated = false
	if !ConditionsMet(gs, evt) {
		t.Fatal("Granada aktifken destek faction koşulu sağlanmalı")
	}
}

func TestTickTriggersFactionSubjugationEventForPlayer(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID:         "house_lancaster",
		LastSubjugationActorID:  "house_lancaster",
		LastSubjugatedFactionID: "house_york",
		Factions: map[faction.FactionID]*faction.Faction{
			"england":         {ID: "england"},
			"house_lancaster": {ID: "house_lancaster"},
			"house_york":      {ID: "house_york"},
		},
		FiredEventIDs: map[string]bool{},
	}
	evt := &Event{
		ID:              "early_reunification",
		Target:          "specific_faction",
		AffectedFaction: "england",
		OneShot:         true,
		RequiresFlags:   []string{"war_started"},
		FactionSubjugationTrigger: &FactionSubjugationTrigger{
			FactionIDs:         []string{"house_lancaster", "house_york"},
			RequirePlayerActor: true,
		},
		Choices: []Choice{{ID: "reunify"}},
	}
	gs.FiredEventIDs["flag:war_started"] = true

	if got := Tick(gs, []*Event{evt}); got != evt {
		t.Fatalf("oyuncu hanedanı rakibi yendiğinde erken birleşme eventi tetiklenmedi: got=%v", got)
	}
	if !gs.FiredEventIDs[evt.ID] {
		t.Fatal("erken birleşme event'i tek seferlik olarak işaretlenmedi")
	}
	if gs.LastSubjugationActorID != "" || gs.LastSubjugatedFactionID != "" {
		t.Fatal("siyasi üstünlük bağlamı event taramasından sonra temizlenmedi")
	}
}
