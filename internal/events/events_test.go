package events

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

func TestTickReturnsEventWithoutApplyingEffects(t *testing.T) {
	gs := &state.GameState{
		Turn:            7,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Gold: 100, Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "player", Satisfaction: 60},
		},
	}
	evts := []*Event{{
		ID:          "trade_boom",
		NameTR:      "Ticaret Patlaması",
		Target:      "player_faction",
		MinTurn:     6,
		Probability: 1,
		GoldDelta:   50,
		SatDelta:    10,
	}}

	evt := Tick(gs, evts)
	if evt == nil || evt.ID != "trade_boom" {
		t.Fatalf("tetiklenen event bekleniyordu, got=%v", evt)
	}
	if gs.Factions["player"].Gold != 100 {
		t.Fatalf("Tick artık effect uygulamamalıydı, gold=%d", gs.Factions["player"].Gold)
	}
	if gs.Regions["r1"].Satisfaction != 60 {
		t.Fatalf("Tick artık satisfaction değiştirmemeliydi, got=%d", gs.Regions["r1"].Satisfaction)
	}
}

func TestTickTriggersHistoricalMonthInsideQuarterlyTurn(t *testing.T) {
	gs := &state.GameState{Year: 1337, Month: 3, MonthsPerTurn: 3}
	evts := []*Event{{
		ID:              "hundred_years_war",
		HistoricalYear:  1337,
		HistoricalMonth: 5,
		OneShot:         true,
	}}

	evt := Tick(gs, evts)
	if evt == nil || evt.ID != "hundred_years_war" {
		t.Fatalf("Mayıs olayı Mart-Mayıs stratejik turunda atlanmamalıydı: %+v", evt)
	}
	if !gs.FiredEventIDs["hundred_years_war"] {
		t.Fatal("tek seferlik tarihsel olay işaretlenmeliydi")
	}
}

func TestTickCarriesSecondHistoricalEventIntoNextQuarter(t *testing.T) {
	gs := &state.GameState{Year: 1454, Month: 12, MonthsPerTurn: 3}
	evts := []*Event{
		{ID: "first", HistoricalYear: 1455, HistoricalMonth: 1, OneShot: true},
		{ID: "second", HistoricalYear: 1455, HistoricalMonth: 1, OneShot: true},
	}
	if evt := Tick(gs, evts); evt == nil || evt.ID != "first" {
		t.Fatalf("ilk çeyrek olayı bekleniyordu: %+v", evt)
	}
	gs.AdvanceTurn()
	if evt := Tick(gs, evts); evt == nil || evt.ID != "second" {
		t.Fatalf("aynı çeyrekteki ikinci olay sonraki turda atlanmamalıydı: %+v", evt)
	}
}

func TestAffectedRegionIDsAllArmiesSkipsSeaRegions(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"land_a": {ID: "land_a"},
			"land_b": {ID: "land_b"},
			"sea":    {ID: "sea", IsSea: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"land_army":      {ID: "land_army", RegionID: "land_a"},
			"duplicate_army": {ID: "duplicate_army", RegionID: "land_a"},
			"fleet_at_sea":   {ID: "fleet_at_sea", RegionID: "sea", IsNaval: true},
			"docked_fleet":   {ID: "docked_fleet", RegionID: "sea", DockedRegionID: "land_b", IsNaval: true},
			"invalid_region": {ID: "invalid_region", RegionID: "missing"},
			"fleet_bad_dock": {ID: "fleet_bad_dock", RegionID: "sea", DockedRegionID: "sea", IsNaval: true},
		},
	}

	got := affectedRegionIDs(gs, &Event{ID: "harsh_winter", Target: "all_armies"}, nil, "")
	want := map[world.RegionID]bool{"land_a": false, "land_b": false}
	if len(got) != len(want) {
		t.Fatalf("beklenen kara bolgesi sayisi %d, got=%d (%v)", len(want), len(got), got)
	}
	for _, rid := range got {
		seen, ok := want[rid]
		if !ok {
			t.Fatalf("beklenmeyen bolge: %s (%v)", rid, got)
		}
		if seen {
			t.Fatalf("bolge tekrarlanmamali: %s (%v)", rid, got)
		}
		want[rid] = true
	}
}

func TestApplyChoiceUpdatesFactionArmyAndRelations(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Gold: 200, Grain: 120, Religion: religion.Catholic},
			"ally":   {ID: "ally", Gold: 50, Grain: 40, Religion: religion.Catholic},
			"rival":  {ID: "rival", Gold: 80, Grain: 50, Religion: religion.Sunni},
		},
		Regions: map[world.RegionID]*world.Region{
			"cap": {ID: "cap", OwnerID: "player", Satisfaction: 55},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "player", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"):  {FactionA: "ally", FactionB: "player", Score: 10, Stance: faction.StancePeace},
			faction.RelationKey("player", "rival"): {FactionA: "player", FactionB: "rival", Score: -20, Stance: faction.StanceWar},
		},
		TechTypes: map[string]*tech.Technology{
			"trade_routes": {ID: "trade_routes"},
			"banking":      {ID: "banking", TurnsRequired: 4},
		},
	}
	evt := &Event{
		ID:     "policy",
		Target: "player_faction",
		Choices: []Choice{{
			LabelTR: "Karantina",
			Effect: Effect{
				Target:            "player_faction",
				SatDelta:          8,
				GoldDelta:         -30,
				GrainDelta:        -10,
				ArmyHPMod:         0.9,
				RelationDeltaAll:  -5,
				CompleteTechs:     []string{"trade_routes"},
				StartResearchTech: "banking",
				Relations: []RelationEffect{{
					FactionID:  "ally",
					Stance:     "trade",
					ScoreDelta: 7,
				}},
				SetFlags: []string{"plague_quarantine"},
			},
		}},
	}

	choice, ok := ApplyChoice(gs, evt, 0)
	if !ok || choice.LabelTR != "Karantina" {
		t.Fatalf("choice apply başarısız, choice=%+v ok=%v", choice, ok)
	}
	if gs.Factions["player"].Gold != 170 || gs.Factions["player"].Grain != 110 {
		t.Fatalf("ekonomi etkisi uygulanmadı, gold=%d grain=%d", gs.Factions["player"].Gold, gs.Factions["player"].Grain)
	}
	if gs.Regions["cap"].Satisfaction != 63 {
		t.Fatalf("memnuniyet 63 olmalıydı, got=%d", gs.Regions["cap"].Satisfaction)
	}
	if hp := gs.Armies["a1"].Units[0].CurrentHP; hp != 90 {
		t.Fatalf("ordu hp 90 olmalıydı, got=%d", hp)
	}
	if score := gs.Relations[faction.RelationKey("player", "ally")].Score; score != 12 {
		t.Fatalf("ally relation 12 olmalıydı, got=%d", score)
	}
	if score := gs.Relations[faction.RelationKey("player", "rival")].Score; score != -25 {
		t.Fatalf("rival relation -25 olmalıydı, got=%d", score)
	}
	if stance := gs.Relations[faction.RelationKey("player", "ally")].Stance; stance != faction.StanceTrade {
		t.Fatalf("ally stance trade olmalıydı, got=%s", stance)
	}
	if !gs.FiredEventIDs["flag:plague_quarantine"] {
		t.Fatalf("choice flag set edilmeliydi")
	}
	if !gs.Factions["player"].Research.Completed["trade_routes"] {
		t.Fatalf("trade_routes tech tamamlanmalıydı")
	}
	if gs.Factions["player"].Research.ActiveID != "banking" || gs.Factions["player"].Research.TurnsLeft != 4 {
		t.Fatalf("banking aktif araştırma başlamalıydı, got=%s/%d", gs.Factions["player"].Research.ActiveID, gs.Factions["player"].Research.TurnsLeft)
	}
}

func TestTickRequiresAndBlocksFlags(t *testing.T) {
	gs := &state.GameState{
		Year:  1454,
		Month: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman", Research: faction.ResearchState{Completed: map[string]bool{}}},
		},
		Regions: map[world.RegionID]*world.Region{
			"constantinople": {ID: "constantinople", OwnerID: "ottoman"},
		},
		FiredEventIDs: map[string]bool{
			"flag:capital_rebuild": true,
		},
	}
	evts := []*Event{
		{
			ID:              "blocked_event",
			NameTR:          "Blocked",
			HistoricalYear:  1454,
			HistoricalMonth: 1,
			RequiresFlags:   []string{"arsenal_push"},
		},
		{
			ID:                   "allowed_event",
			NameTR:               "Allowed",
			HistoricalYear:       1454,
			HistoricalMonth:      1,
			RequiresFlags:        []string{"capital_rebuild"},
			BlocksFlags:          []string{"arsenal_push"},
			Target:               "specific_faction",
			AffectedFaction:      "ottoman",
			RequiresOwnedRegions: []world.RegionID{"constantinople"},
		},
	}

	evt := Tick(gs, evts)
	if evt == nil || evt.ID != "allowed_event" {
		t.Fatalf("flag koşullarına göre allowed_event bekleniyordu, got=%v", evt)
	}
}

func TestTickRequiresTechAndRelationConditions(t *testing.T) {
	gs := &state.GameState{
		Year:  1494,
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"aragon":          {ID: "aragon", Research: faction.ResearchState{Completed: map[string]bool{"navigation": true}}},
			"castile_kingdom": {ID: "castile_kingdom"},
			"portugal":        {ID: "portugal"},
		},
		Regions: map[world.RegionID]*world.Region{
			"granada": {ID: "granada", OwnerID: "aragon"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("aragon", "castile_kingdom"): {
				FactionA: "aragon",
				FactionB: "castile_kingdom",
				Score:    22,
				Stance:   faction.StanceAllied,
			},
			faction.RelationKey("aragon", "portugal"): {
				FactionA: "aragon",
				FactionB: "portugal",
				Score:    -60,
				Stance:   faction.StanceWar,
			},
		},
		FiredEventIDs: map[string]bool{
			"flag:atlantic_expedition": true,
			"flag:iberian_market":      true,
		},
	}
	evts := []*Event{
		{
			ID:                   "blocked_by_tech",
			NameTR:               "BlockedByTech",
			HistoricalYear:       1494,
			HistoricalMonth:      4,
			Target:               "specific_faction",
			AffectedFaction:      "aragon",
			RequiresFlags:        []string{"atlantic_expedition"},
			RequiresOwnedRegions: []world.RegionID{"granada"},
			BlocksTechs:          []string{"navigation"},
		},
		{
			ID:                   "blocked_by_relation",
			NameTR:               "BlockedByRelation",
			HistoricalYear:       1494,
			HistoricalMonth:      4,
			Target:               "specific_faction",
			AffectedFaction:      "aragon",
			RequiresFlags:        []string{"iberian_market"},
			RequiresOwnedRegions: []world.RegionID{"granada"},
			RelationRequirements: []RelationRequirement{{
				FactionID:     "portugal",
				BlocksStances: []string{string(faction.StanceWar)},
			}},
		},
		{
			ID:                   "allowed_with_relation_and_tech",
			NameTR:               "Allowed",
			HistoricalYear:       1494,
			HistoricalMonth:      4,
			Target:               "specific_faction",
			AffectedFaction:      "aragon",
			RequiresFlags:        []string{"atlantic_expedition"},
			RequiresOwnedRegions: []world.RegionID{"granada"},
			RequiresTechs:        []string{"navigation"},
			RelationRequirements: []RelationRequirement{{
				FactionID:    "castile_kingdom",
				AnyOfStances: []string{string(faction.StancePeace), string(faction.StanceTrade), string(faction.StanceAllied)},
				MinScore:     10,
			}},
		},
	}

	evt := Tick(gs, evts)
	if evt == nil || evt.ID != "allowed_with_relation_and_tech" {
		t.Fatalf("tech ve relation koşullarına göre allowed event bekleniyordu, got=%v", evt)
	}
}

func TestTickBlocksCapitalSettlementWhenAlreadyCurrentOrPending(t *testing.T) {
	gs := &state.GameState{
		Year:  1453,
		Month: 5,
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {
				ID:                         "ottoman",
				CapitalSettlementID:        "edirne",
				PendingCapitalSettlementID: "ghost",
				PendingCapitalTurns:        2,
				Research:                   faction.ResearchState{Completed: map[string]bool{}},
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"constantinople": {ID: "constantinople", OwnerID: "ottoman"},
		},
	}
	evts := []*Event{
		{
			ID:                        "blocked_current",
			NameTR:                    "BlockedCurrent",
			HistoricalYear:            1453,
			HistoricalMonth:           5,
			Target:                    "specific_faction",
			AffectedFaction:           "ottoman",
			RequiresOwnedRegions:      []world.RegionID{"constantinople"},
			BlocksCapitalSettlementID: "edirne",
		},
		{
			ID:                        "blocked_pending",
			NameTR:                    "BlockedPending",
			HistoricalYear:            1453,
			HistoricalMonth:           5,
			Target:                    "specific_faction",
			AffectedFaction:           "ottoman",
			RequiresOwnedRegions:      []world.RegionID{"constantinople"},
			BlocksCapitalSettlementID: "ghost",
		},
		{
			ID:                        "allowed",
			NameTR:                    "Allowed",
			HistoricalYear:            1453,
			HistoricalMonth:           5,
			Target:                    "specific_faction",
			AffectedFaction:           "ottoman",
			RequiresOwnedRegions:      []world.RegionID{"constantinople"},
			BlocksCapitalSettlementID: "constantinople_main",
		},
	}

	evt := Tick(gs, evts)
	if evt == nil || evt.ID != "allowed" {
		t.Fatalf("capital block koşullarına göre allowed event bekleniyordu, got=%v", evt)
	}
}

func TestApplyChoiceStartsCapitalMove(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", CapitalSettlementID: "old_cap"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID:      "home",
				OwnerID: "player",
				Settlements: []world.Settlement{
					{ID: "old_cap", NameTR: "Eski Başkent", IsCenter: true},
					{ID: "new_cap", NameTR: "Yeni Başkent"},
				},
			},
		},
	}
	evt := &Event{
		ID:     "move_capital",
		Target: "player_faction",
		Choices: []Choice{{
			LabelTR: "Taşı",
			Effect: Effect{
				Target:              "player_faction",
				CapitalSettlementID: "new_cap",
				CapitalMoveTurns:    3,
			},
		}},
	}

	if _, ok := ApplyChoice(gs, evt, 0); !ok {
		t.Fatal("choice uygulanmalıydı")
	}
	f := gs.Factions["player"]
	if f.PendingCapitalSettlementID != "new_cap" || f.PendingCapitalTurns != 3 {
		t.Fatalf("pending başkent taşıma başlamalıydı, got=%+v", f)
	}
}

func TestApplyStoresTemporaryGrainEventModifiers(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "player"},
		},
	}
	evt := &Event{
		ID:                     "drought",
		NameTR:                 "Kuraklık",
		Target:                 "player_faction",
		GrainProductionPercent: -40,
		GrainDemandPercent:     20,
	}

	Apply(gs, evt)
	if len(gs.ActiveRegionEvents) != 1 {
		t.Fatalf("olay bölgeye aktif status eklemeliydi, got=%d", len(gs.ActiveRegionEvents))
	}
	status := gs.ActiveRegionEvents[0]
	if status.Type != "famine" || status.GrainProductionPercent != -40 || status.GrainDemandPercent != 20 {
		t.Fatalf("geçici tahıl modifiyerleri status'a taşınmalıydı, got=%+v", status)
	}
}

func TestApplyStoresRandomRegionGrainModifiersOnSelectedRegion(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Regions: map[world.RegionID]*world.Region{
			"only_region": {ID: "only_region", OwnerID: "player"},
		},
	}
	Apply(gs, &Event{
		ID:                     "drought_random",
		Target:                 "random_region",
		GrainProductionPercent: -50,
		GrainDemandPercent:     10,
	})

	if len(gs.ActiveRegionEvents) != 1 || gs.ActiveRegionEvents[0].RegionID != "only_region" {
		t.Fatalf("random bölge seçimi aktif tahıl status'una taşınmalıydı, got=%+v", gs.ActiveRegionEvents)
	}
}

func TestApplyChoiceStoresChoiceGrainModifiers(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "player"},
		},
	}
	evt := &Event{
		ID:     "harvest_choice",
		NameTR: "Hasat Kararı",
		Target: "player_faction",
		Choices: []Choice{{
			LabelTR: "Ambarları doldur",
			Effect: Effect{
				GrainProductionPercent: 25,
				GrainDemandPercent:     -10,
			},
		}},
	}

	if _, ok := ApplyChoice(gs, evt, 0); !ok {
		t.Fatal("seçim uygulanmalıydı")
	}
	status := gs.ActiveRegionEvents[0]
	if status.Type != "blessing" || status.GrainProductionPercent != 25 || status.GrainDemandPercent != -10 {
		t.Fatalf("choice tahıl modifiyerleri status'a taşınmalıydı, got=%+v", status)
	}
}
