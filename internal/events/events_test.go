package events

import (
	"path/filepath"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestLoad1300HistoricalEventChains(t *testing.T) {
	path := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data", "events.json")
	events, err := LoadEvents(path)
	if err != nil {
		t.Fatalf("1300 event verisi yüklenemedi: %v", err)
	}
	required := map[string]bool{
		"golden_bull_1356":                      false,
		"jan_hus_constance_1415":                false,
		"council_of_constance_1417":             false,
		"hussite_uprising_1419":                 false,
		"first_hussite_crusade_1420":            false,
		"hussite_counteroffensive_1427":         false,
		"lipany_hussite_settlement_1434":        false,
		"baltic_crusade_against_lithuania_1345": false,
		"lithuanian_christianization_1387":      false,
		"grunwald_battle_1410":                  false,
		"burgundian_succession_war_1477":        false,
		"burgundian_succession_settlement_1493": false,
		"habsburg_imperial_succession_1440":     false,
		"second_kosovo_battle_1448":             false,
		"belgrade_defense_1456":                 false,
		"serbian_despotate_falls_1459":          false,
		"fall_of_trebizond_1461":                false,
		"bosnia_conquest_1463":                  false,
		"italian_wars_begin_1494":               false,
		"marignano_battle_1515":                 false,
		"diet_of_worms_1521":                    false,
		"pavia_battle_1525":                     false,
		"german_peasants_war_1525":              false,
		"sack_of_rome_1527":                     false,
		"siege_of_vienna_1529":                  false,
	}
	for _, event := range events {
		if event == nil {
			continue
		}
		if _, ok := required[event.ID]; ok {
			required[event.ID] = true
		}
		if event.ID == "habsburg_imperial_succession_1440" {
			if len(event.Choices) != 1 || event.Choices[0].Effect.ImperialSuccession == nil ||
				event.Choices[0].Effect.ImperialSuccession.EmperorID != "austria_duchy" ||
				!event.Choices[0].Effect.ImperialSuccession.ElectionLocked {
				t.Fatal("Habsburg imparatorluk event'inin seçim kilidi etkisi eksik")
			}
		}
		if event.ID == "hussite_uprising_1419" && !containsEventFlag(event.RequiresFlags, "jan_hus_executed_1415") {
			t.Fatal("Hussit ayaklanması Jan Hus event flag'ine bağlanmamış")
		}
		if event.ID == "diet_of_worms_1521" && !containsEventFlag(event.RequiresFlags, "reformation_1517_started") {
			t.Fatal("Worms event'i Reformasyon karar flag'ine bağlanmamış")
		}
	}
	for id, found := range required {
		if !found {
			t.Fatalf("beklenen tarihsel event eksik: %s", id)
		}
	}
}

func containsEventFlag(flags []string, wanted string) bool {
	for _, flag := range flags {
		if flag == wanted {
			return true
		}
	}
	return false
}

func Test1300UntimedFlagFollowUpsFollowParent(t *testing.T) {
	path := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data", "events.json")
	events, err := LoadEvents(path)
	if err != nil {
		t.Fatalf("1300 event verisi yüklenemedi: %v", err)
	}
	indices := make(map[string]int, len(events))
	for index, event := range events {
		if event != nil {
			indices[event.ID] = index
		}
	}
	for _, pair := range [][2]string{
		{"swiss_confederacy_emerges_1310", "swiss_habsburg_final_victory"},
		{"war_of_the_roses_1455", "war_of_the_roses_early_reunification_offer"},
	} {
		parentIndex, parentOK := indices[pair[0]]
		childIndex, childOK := indices[pair[1]]
		if !parentOK || !childOK || childIndex != parentIndex+1 {
			t.Fatalf("%s event'i %s event'inin hemen altında değil", pair[1], pair[0])
		}
	}
}

func TestApplyUnitReinforcementsCreatesLandArmyAndDockedFleet(t *testing.T) {
	const ownerID = faction.FactionID("east_rome")
	capital := &world.Region{
		ID:      "constantinople",
		OwnerID: string(ownerID),
		Neighbors: []world.RegionID{
			"sea_of_marmara",
		},
		Settlements: []world.Settlement{{ID: "constantinople_main"}},
	}
	sea := &world.Region{ID: "sea_of_marmara", IsSea: true}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			capital.ID: capital,
			sea.ID:     sea,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			ownerID: {ID: ownerID, CapitalSettlementID: "constantinople_main"},
		},
		Armies: map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Category: army.CategoryInfantry},
			"warship":  {ID: "warship", Category: army.CategoryNavalWar},
		},
	}

	Apply(gs, &Event{
		Target:          "specific_faction",
		AffectedFaction: string(ownerID),
		UnitReinforcements: []UnitReinforcementEffect{
			{UnitType: "infantry", UnitCount: 5},
			{UnitType: "warship", UnitCount: 3},
		},
	})

	if len(gs.Armies) != 2 {
		t.Fatalf("%d ordu oluşturuldu, 2 bekleniyordu", len(gs.Armies))
	}
	var foundLand, foundFleet bool
	for _, current := range gs.Armies {
		switch {
		case !current.IsNaval:
			foundLand = len(current.Units) == 5 && current.Units[0].TypeID == "infantry" && current.RegionID == capital.ID
		case current.IsNaval:
			foundFleet = len(current.Units) == 3 && current.Units[0].TypeID == "warship" &&
				current.RegionID == sea.ID && current.DockedRegionID == capital.ID &&
				current.DockedSettlementID == "constantinople_main"
		}
	}
	if !foundLand || !foundFleet {
		t.Fatalf("takviyeler beklenen kara ordusu ve dock edilmiş filo olarak oluşturulmadı")
	}
}

func TestApplyChoiceAppliesChoiceReinforcement(t *testing.T) {
	const ownerID = faction.FactionID("ottoman")
	capital := &world.Region{
		ID:          "bithynia",
		OwnerID:     string(ownerID),
		Settlements: []world.Settlement{{ID: "sogut"}},
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{capital.ID: capital},
		Factions: map[faction.FactionID]*faction.Faction{
			ownerID: {ID: ownerID, CapitalSettlementID: "sogut"},
		},
		Armies: map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Category: army.CategoryInfantry},
		},
	}

	_, ok := ApplyChoice(gs, &Event{
		Target:          "specific_faction",
		AffectedFaction: string(ownerID),
		Choices:         []Choice{{Effect: Effect{UnitReinforcements: []UnitReinforcementEffect{{UnitType: "infantry", UnitCount: 3}}}}},
	}, 0)
	if !ok || len(gs.Armies) != 1 {
		t.Fatalf("seçim takviyesi uygulanmadı")
	}
	for _, current := range gs.Armies {
		if len(current.Units) != 3 || current.Units[0].TypeID != "infantry" {
			t.Fatalf("seçim sonrası yanlış birlik oluşturuldu")
		}
	}
}

func TestApplyBaseEventReinforcement(t *testing.T) {
	const ownerID = faction.FactionID("ottoman")
	capital := &world.Region{
		ID:          "bilecik_frontier",
		OwnerID:     string(ownerID),
		Settlements: []world.Settlement{{ID: "sogut"}},
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{capital.ID: capital},
		Factions: map[faction.FactionID]*faction.Faction{
			ownerID: {ID: ownerID, CapitalSettlementID: "sogut"},
		},
		Armies: map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Category: army.CategoryInfantry},
		},
	}

	Apply(gs, &Event{
		Target:          "specific_faction",
		AffectedFaction: string(ownerID),
		UnitReinforcements: []UnitReinforcementEffect{
			{UnitType: "infantry", UnitCount: 3},
		},
	})

	if len(gs.Armies) != 1 {
		t.Fatalf("temel event takviyesi %d ordu oluşturdu, 1 bekleniyordu", len(gs.Armies))
	}
	for _, current := range gs.Armies {
		if current.RegionID != capital.ID || len(current.Units) != 3 || current.Units[0].TypeID != "infantry" {
			t.Fatalf("temel event takviyesi başkentte 3 piyade oluşturmadı")
		}
	}
}
