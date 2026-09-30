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
		"ottoman_turkmen_gazi_migration_1310":      false,
		"ottoman_bithynian_campaign_muster_1321":   false,
		"ottoman_anatolian_beylik_support_1327":    false,
		"ottoman_karesi_peaceful_integration_1345": false,
		"germiyan_dowry_settlement_1381":           false,
		"germiyan_will_integration_1429":           false,
		"golden_bull_1356":                         false,
		"jan_hus_constance_1415":                   false,
		"council_of_constance_1417":                false,
		"hussite_uprising_1419":                    false,
		"first_hussite_crusade_1420":               false,
		"hussite_counteroffensive_1427":            false,
		"lipany_hussite_settlement_1434":           false,
		"baltic_crusade_against_lithuania_1345":    false,
		"lithuanian_christianization_1387":         false,
		"grunwald_battle_1410":                     false,
		"burgundian_succession_war_1477":           false,
		"burgundian_succession_settlement_1493":    false,
		"habsburg_imperial_succession_1440":        false,
		"second_kosovo_battle_1448":                false,
		"belgrade_defense_1456":                    false,
		"serbian_despotate_falls_1459":             false,
		"fall_of_trebizond_1461":                   false,
		"bosnia_conquest_1463":                     false,
		"italian_wars_begin_1494":                  false,
		"marignano_battle_1515":                    false,
		"diet_of_worms_1521":                       false,
		"pavia_battle_1525":                        false,
		"german_peasants_war_1525":                 false,
		"sack_of_rome_1527":                        false,
		"siege_of_vienna_1529":                     false,
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
		if event.ID == "swiss_confederacy_emerges_1310" {
			revival := event.SuccessorRevival
			if revival == nil || revival.FactionID != "swiss_confederacy" || revival.UnitType != "infantry" || revival.UnitCount != 6 {
				t.Fatal("İsviçre kuruluş event'i 6 piyade ile yeniden kurulmuyor")
			}
		}
		if event.ID == "ilkhanate_breakup_1335" {
			if len(event.SuccessorRevivals) != 6 {
				t.Fatalf("İlhanlı parçalanması %d ardıl devlet diriltiyor, 6 bekleniyordu", len(event.SuccessorRevivals))
			}
			for _, revival := range event.SuccessorRevivals {
				if revival.UnitType != "infantry" || revival.UnitCount <= 0 {
					t.Fatalf("%s ardılı piyade kuruluş ordusuyla tanımlanmamış", revival.FactionID)
				}
			}
			for _, revival := range event.SuccessorRevivals {
				if revival.FactionID == "eretna_state" && revival.RegionID != "nigde" {
					t.Fatalf("Eretna yanlış bölgede diriltiliyor: %s", revival.RegionID)
				}
			}
			if len(event.ArmyDefections) != 4 {
				t.Fatalf("İlhanlı parçalanması %d mevcut orduyu ardıllara aktarıyor, 4 bekleniyordu", len(event.ArmyDefections))
			}
			for _, defection := range event.ArmyDefections {
				if defection.SourceFactionID != "ilkhanate" || defection.ArmyCount != 1 ||
					len(defection.SourceRegionIDs) != 1 || defection.DestinationRegionID == "" {
					t.Fatalf("İlhanlı ordu ayrılması eksik veya belirsiz: %+v", defection)
				}
			}
		}
		if event.ID == "timurid_rise_1370" || event.ID == "karakoyunlu_rise_1375" || event.ID == "akkoyunlu_rise_1378" {
			if len(event.ArmyDefections) != 1 || event.ArmyDefections[0].SourceFactionID != "ilkhanate" ||
				event.ArmyDefections[0].ArmyCount != 1 {
				t.Fatalf("%s mevcut İlhanlı ordusunun ardıl devlete geçişini tanımlamıyor", event.ID)
			}
		}
		if event.ID == "swiss_morgarten_war_1315" {
			if event.CombatDefensePercent != 30 {
				t.Fatalf("Morgarten savunma bonusu = %d, 30 bekleniyordu", event.CombatDefensePercent)
			}
			if len(event.UnitReinforcements) != 1 || event.UnitReinforcements[0].UnitType != "infantry" || event.UnitReinforcements[0].UnitCount != 3 {
				t.Fatal("Morgarten event'i İsviçre'ye 3 piyade Waldstätte takviyesi vermiyor")
			}
		}
		if event.ID == "ottoman_turkmen_gazi_migration_1310" {
			if event.HistoricalYear != 1310 || event.AffectedFaction != "ottoman" ||
				len(event.RequiresOwnedRegions) != 2 || len(event.RelationRequirements) != 1 ||
				event.RelationRequirements[0].FactionID != "east_rome" ||
				event.RelationRequirements[0].Stance != string(faction.StanceWar) {
				t.Fatal("1310 Osmanlı göç event'inin tarih, sahiplik veya savaş koşulları eksik")
			}
			if len(event.UnitReinforcements) != 2 ||
				event.UnitReinforcements[0].UnitType != "infantry" || event.UnitReinforcements[0].UnitCount != 2 ||
				event.UnitReinforcements[1].UnitType != "light_cavalry" || event.UnitReinforcements[1].UnitCount != 2 {
				t.Fatal("1310 Osmanlı göç event'i 2 piyade ve 2 hafif süvari takviyesi vermiyor")
			}
		}
		if event.ID == "ottoman_bithynian_campaign_muster_1321" {
			if event.HistoricalYear != 1321 || event.AffectedFaction != "ottoman" ||
				len(event.RequiresOwnedRegions) != 2 || len(event.RequiresUnownedRegions) != 1 ||
				event.RequiresUnownedRegions[0] != "bursa" || len(event.RelationRequirements) != 1 ||
				event.RelationRequirements[0].FactionID != "east_rome" ||
				event.RelationRequirements[0].Stance != string(faction.StanceWar) {
				t.Fatal("1321 Bithynia seferberliği event'inin tarih, sahiplik veya savaş koşulları eksik")
			}
			if len(event.UnitReinforcements) != 2 ||
				event.UnitReinforcements[0].UnitType != "infantry" || event.UnitReinforcements[0].UnitCount != 2 ||
				event.UnitReinforcements[1].UnitType != "catapult" || event.UnitReinforcements[1].UnitCount != 1 {
				t.Fatal("1321 Bithynia seferberliği 2 piyade ve 1 mancınık takviyesi vermiyor")
			}
		}
		if event.ID == "ottoman_anatolian_beylik_support_1327" {
			if event.HistoricalYear != 1327 || event.AffectedFaction != "ottoman" ||
				len(event.RequiresOwnedRegions) != 1 || event.RequiresOwnedRegions[0] != "bursa" ||
				len(event.RelationRequirements) != 3 || len(event.Relations) != 3 {
				t.Fatal("1327 Anadolu beylikleri desteğinin tarih, Bursa veya ilişki koşulları eksik")
			}
			for _, requirement := range event.RelationRequirements {
				if len(requirement.BlocksStances) != 1 || requirement.BlocksStances[0] != string(faction.StanceWar) {
					t.Fatal("1327 Anadolu beylikleri desteği savaş koşulunu engellemiyor")
				}
			}
			if len(event.UnitReinforcements) != 3 ||
				event.UnitReinforcements[0].UnitType != "infantry" || event.UnitReinforcements[0].UnitCount != 3 ||
				event.UnitReinforcements[1].UnitType != "light_cavalry" || event.UnitReinforcements[1].UnitCount != 2 ||
				event.UnitReinforcements[2].UnitType != "catapult" || event.UnitReinforcements[2].UnitCount != 1 {
				t.Fatal("1327 Anadolu beylikleri desteği 3 piyade, 2 hafif süvari ve 1 mancınık vermiyor")
			}
		}
		if event.ID == "germiyan_dowry_settlement_1381" {
			settlement := event.DynasticSettlement
			if settlement == nil || settlement.SourceFactionID != "germiyan_bey" ||
				settlement.RecipientFactionID != "ottoman" || settlement.Mode != "dowry" ||
				len(settlement.RegionIDs) != 1 || settlement.RegionIDs[0] != "kutahya" ||
				settlement.ArmyTransferPercent != 30 || settlement.ResourceTransferPercent != 30 ||
				!event.HistoricalDateStrict {
				t.Fatal("Germiyan çeyizi Kütahya ve yüzde 30 aktarım sözleşmesini taşımıyor")
			}
		}
		if event.ID == "germiyan_will_integration_1429" {
			settlement := event.DynasticSettlement
			if settlement == nil || !settlement.AutoUnionWhenSourceEmpty ||
				settlement.UnionResultFactionID != "ottoman" || len(settlement.RegionIDs) != 1 ||
				settlement.RegionIDs[0] != "germiyan" || !event.HistoricalDateStrict ||
				!containsEventFlag(event.RequiresFlags, "germiyan_dowry_1381") {
				t.Fatal("Germiyan vasiyeti kalan bölgeyi otomatik birleşmeye bağlamıyor")
			}
		}
	}
	for id, found := range required {
		if !found {
			t.Fatalf("beklenen tarihsel event eksik: %s", id)
		}
	}
}

func TestTickHistoricalStateTriggeredEventWaitsForMinimumDate(t *testing.T) {
	const ottomanID = faction.FactionID("ottoman")
	gs := &state.GameState{
		Year:  1300,
		Month: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			ottomanID: {ID: ottomanID},
		},
		Regions: map[world.RegionID]*world.Region{
			"bursa": {ID: "bursa", OwnerID: string(ottomanID)},
		},
	}
	event := &Event{
		ID:                   "bursa_conquest_1326",
		HistoricalYear:       1326,
		HistoricalMonth:      4,
		OneShot:              true,
		Target:               "specific_faction",
		AffectedFaction:      string(ottomanID),
		RequiresOwnedRegions: []world.RegionID{"bursa"},
	}

	if got := Tick(gs, []*Event{event}); got != nil {
		t.Fatalf("state koşulu gerçekleşmiş tarihsel event kendi tarihinden önce tetiklendi: %#v", got)
	}

	gs.Year = 1327
	gs.Month = 1
	if got := Tick(gs, []*Event{event}); got != event {
		t.Fatalf("state koşulu gerçekleşmiş tarihsel event minimum tarihten sonra tetiklenmedi: %#v", got)
	}
	if !gs.FiredEventIDs[event.ID] {
		t.Fatal("minimum tarihten sonra tetiklenen tek seferlik event fired olarak işaretlenmedi")
	}
}

func TestRequiresUnownedRegionsBlocksCompletedTarget(t *testing.T) {
	const ottomanID = faction.FactionID("ottoman")
	event := &Event{
		Target:                 "specific_faction",
		AffectedFaction:        string(ottomanID),
		RequiresUnownedRegions: []world.RegionID{"bursa"},
	}
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			ottomanID: {ID: ottomanID},
		},
		Regions: map[world.RegionID]*world.Region{
			"bursa": {ID: "bursa", OwnerID: "east_rome"},
		},
	}
	if !ConditionsMet(gs, event) {
		t.Fatal("hedef faction Bursa'yı kontrol etmiyorken unowned koşulu sağlanmadı")
	}

	gs.Regions["bursa"].OwnerID = string(ottomanID)
	if ConditionsMet(gs, event) {
		t.Fatal("hedef faction Bursa'yı aldıktan sonra unowned koşulu sağlandı")
	}
}

func TestTickDateOnlyHistoricalEventStillWaitsForYear(t *testing.T) {
	const ottomanID = faction.FactionID("ottoman")
	gs := &state.GameState{
		Year:  1300,
		Month: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			ottomanID: {ID: ottomanID},
		},
	}
	event := &Event{
		ID:              "date_only_event",
		HistoricalYear:  1326,
		HistoricalMonth: 4,
		OneShot:         true,
		Target:          "specific_faction",
		AffectedFaction: string(ottomanID),
	}

	if got := Tick(gs, []*Event{event}); got != nil {
		t.Fatalf("yalnız tarih koşullu event erken tetiklendi: %#v", got)
	}
}

func TestTickStrictHistoricalStateTriggeredEventWaitsForMinimumDate(t *testing.T) {
	const ottomanID = faction.FactionID("ottoman")
	gs := &state.GameState{
		Year:  1300,
		Month: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			ottomanID: {ID: ottomanID},
		},
		Regions: map[world.RegionID]*world.Region{
			"bursa": {ID: "bursa", OwnerID: string(ottomanID)},
		},
	}
	event := &Event{
		ID:                   "strict_state_event",
		HistoricalYear:       1326,
		HistoricalMonth:      4,
		HistoricalDateStrict: true,
		OneShot:              true,
		Target:               "specific_faction",
		AffectedFaction:      string(ottomanID),
		RequiresOwnedRegions: []world.RegionID{"bursa"},
	}

	if got := Tick(gs, []*Event{event}); got != nil {
		t.Fatalf("kesin tarihli state event'i tarihinden önce tetiklendi: %#v", got)
	}

	gs.Year = 1327
	gs.Month = 1
	if got := Tick(gs, []*Event{event}); got != event {
		t.Fatalf("state event'i minimum tarihten sonra tetiklenmedi: %#v", got)
	}
}

func TestTickHistoricalEventRequiresFollowUpFlag(t *testing.T) {
	const germiyanID = faction.FactionID("germiyan_bey")
	gs := &state.GameState{
		Year:  1429,
		Month: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			germiyanID: {ID: germiyanID},
		},
	}
	event := &Event{
		ID:              "germiyan_will_integration_1429",
		HistoricalYear:  1429,
		HistoricalMonth: 1,
		OneShot:         true,
		Target:          "specific_faction",
		AffectedFaction: string(germiyanID),
		RequiresFlags:   []string{"germiyan_dowry_1381"},
	}

	if got := Tick(gs, []*Event{event}); got != nil {
		t.Fatalf("follow-up flag'i olmadan Germiyan vasiyeti tetiklendi: %#v", got)
	}

	gs.FiredEventIDs = map[string]bool{"flag:germiyan_dowry_1381": true}
	if got := Tick(gs, []*Event{event}); got != event {
		t.Fatalf("çeyiz flag'i varken Germiyan vasiyeti tetiklenmedi: %#v", got)
	}
}

func TestApplyOttomanPostBursaSupportHonorsBeylikWarBlock(t *testing.T) {
	path := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data", "events.json")
	definitions, err := LoadEvents(path)
	if err != nil {
		t.Fatalf("1300 event verisi yüklenemedi: %v", err)
	}
	var support *Event
	for _, event := range definitions {
		if event != nil && event.ID == "ottoman_anatolian_beylik_support_1327" {
			support = event
			break
		}
	}
	if support == nil {
		t.Fatal("1327 Anadolu beylikleri destek event'i bulunamadı")
	}

	const ownerID = faction.FactionID("ottoman")
	bursa := &world.Region{
		ID:          "bursa",
		OwnerID:     string(ownerID),
		Settlements: []world.Settlement{{ID: "bursa_prusa"}},
	}
	beylikIDs := []faction.FactionID{
		faction.FactionID("germiyan_bey"),
		faction.FactionID("karesioglu_bey"),
		faction.FactionID("karaman_bey"),
	}
	gs := &state.GameState{
		Year:    1327,
		Month:   9,
		Regions: map[world.RegionID]*world.Region{bursa.ID: bursa},
		Factions: map[faction.FactionID]*faction.Faction{
			ownerID: {ID: ownerID, CapitalSettlementID: "bursa_prusa", Gold: 300, Grain: 500},
		},
		Armies: map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{
			"infantry":      {ID: "infantry", Category: army.CategoryInfantry},
			"light_cavalry": {ID: "light_cavalry", Category: army.CategoryCavalry},
			"catapult":      {ID: "catapult", Category: army.CategorySiege},
		},
	}
	for _, id := range beylikIDs {
		gs.Factions[id] = &faction.Faction{ID: id}
	}

	if !ConditionsMet(gs, support) {
		t.Fatal("1327 destek event'i Bursa Osmanlıdayken ve beyliklerle savaş yokken hazır görünmüyor")
	}
	if got := Tick(gs, []*Event{support}); got != support {
		t.Fatal("1327 Bursa sonrası destek event'i minimum tarihten sonra tetiklenmedi")
	}
	Apply(gs, support)
	if len(gs.Armies) != 3 {
		t.Fatalf("1327 destek event'i %d ordu oluşturdu, 3 bekleniyordu", len(gs.Armies))
	}
	counts := make(map[string]int, len(gs.Armies))
	for _, current := range gs.Armies {
		if current.RegionID != bursa.ID || len(current.Units) == 0 {
			t.Fatalf("Bursa sonrası destek yanlış bölgede oluşturuldu")
		}
		counts[current.Units[0].TypeID] = len(current.Units)
	}
	if counts["infantry"] != 3 || counts["light_cavalry"] != 2 || counts["catapult"] != 1 {
		t.Fatalf("Bursa sonrası destek yanlış birlik bileşimi oluşturdu: %#v", counts)
	}
	if got := gs.Factions[ownerID].Gold; got != 440 {
		t.Fatalf("Bursa sonrası altın desteği = %d, 440 bekleniyordu", got)
	}
	if got := gs.Factions[ownerID].Grain; got != 610 {
		t.Fatalf("Bursa sonrası tahıl desteği = %d, 610 bekleniyordu", got)
	}

	gs.Relations = map[string]*faction.Relation{
		faction.RelationKey(ownerID, beylikIDs[0]): {
			FactionA: ownerID,
			FactionB: beylikIDs[0],
			Stance:   faction.StanceWar,
		},
	}
	if ConditionsMet(gs, support) {
		t.Fatal("1327 destek event'i Germiyan ile savaş varken hazır görünmemeli")
	}
}

func TestApplyOttomanMigrationEventUsesConfiguredReinforcements(t *testing.T) {
	path := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data", "events.json")
	definitions, err := LoadEvents(path)
	if err != nil {
		t.Fatalf("1300 event verisi yüklenemedi: %v", err)
	}
	var migration *Event
	for _, event := range definitions {
		if event != nil && event.ID == "ottoman_turkmen_gazi_migration_1310" {
			migration = event
			break
		}
	}
	if migration == nil {
		t.Fatal("1310 Osmanlı göç event'i bulunamadı")
	}

	const ownerID = faction.FactionID("ottoman")
	const eastRomeID = faction.FactionID("east_rome")
	capital := &world.Region{
		ID:              "bilecik_frontier",
		OwnerID:         string(ownerID),
		Population:      100,
		RuralPopulation: 70,
		Settlements:     []world.Settlement{{ID: "sogut"}},
	}
	frontier := &world.Region{
		ID:              "bithynia",
		OwnerID:         string(ownerID),
		Population:      140,
		RuralPopulation: 100,
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			capital.ID:  capital,
			frontier.ID: frontier,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			ownerID:    {ID: ownerID, CapitalSettlementID: "sogut"},
			eastRomeID: {ID: eastRomeID},
		},
		Armies: map[army.ArmyID]*army.Army{},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(ownerID, eastRomeID): {
				FactionA: ownerID,
				FactionB: eastRomeID,
				Stance:   faction.StanceWar,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry":      {ID: "infantry", Category: army.CategoryInfantry},
			"light_cavalry": {ID: "light_cavalry", Category: army.CategoryCavalry},
		},
	}

	if !ConditionsMet(gs, migration) {
		t.Fatal("1310 göç event'i sahiplik ve savaş koşulları sağlanırken hazır görünmüyor")
	}
	gs.Relations[faction.RelationKey(ownerID, eastRomeID)].Stance = faction.StancePeace
	if ConditionsMet(gs, migration) {
		t.Fatal("1310 göç event'i Doğu Roma ile barış varken hazır görünüyor")
	}
	gs.Relations[faction.RelationKey(ownerID, eastRomeID)].Stance = faction.StanceWar

	Apply(gs, migration)

	if len(gs.Armies) != 2 {
		t.Fatalf("1310 göç event'i %d ordu oluşturdu, 2 bekleniyordu", len(gs.Armies))
	}
	counts := make(map[string]int, len(gs.Armies))
	for _, current := range gs.Armies {
		if current.RegionID != capital.ID || len(current.Units) == 0 {
			t.Fatalf("göç takviyesi Osmanlı başkenti yerine yanlış yerde oluşturuldu")
		}
		counts[current.Units[0].TypeID] = len(current.Units)
	}
	if counts["infantry"] != 2 || counts["light_cavalry"] != 2 {
		t.Fatalf("göç takviyesi yanlış birlik bileşimi oluşturdu: %#v", counts)
	}
	if capital.Population != 130 || frontier.Population != 170 {
		t.Fatalf("göç nüfusu Osmanlı bölgelerine uygulanmadı: bilecik=%d bithynia=%d", capital.Population, frontier.Population)
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

func TestApplySuccessorRevivalUsesConfiguredUnit(t *testing.T) {
	const (
		regionID    = world.RegionID("switzerland")
		successor   = faction.FactionID("swiss_confederacy")
		predecessor = faction.FactionID("hre")
	)
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			regionID: {ID: regionID, OwnerID: string(predecessor), Settlements: []world.Settlement{{ID: "switzerland_schwyz"}}},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			predecessor: {ID: predecessor},
			successor:   {ID: successor, IsEliminated: true, CapitalSettlementID: "switzerland_schwyz"},
		},
		Armies: map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Category: army.CategoryInfantry},
		},
	}

	Apply(gs, &Event{
		Target:          "specific_faction",
		AffectedFaction: string(predecessor),
		SuccessorRevival: &SuccessorRevivalEffect{
			FactionID: string(successor), RegionID: string(regionID), UnitType: "infantry", UnitCount: 6,
		},
	})

	if len(gs.Armies) != 1 {
		t.Fatalf("yeniden kuruluş %d ordu oluşturdu, 1 bekleniyordu", len(gs.Armies))
	}
	for _, current := range gs.Armies {
		if current.OwnerID != string(successor) || current.RegionID != regionID || len(current.Units) != 6 || current.Units[0].TypeID != "infantry" {
			t.Fatalf("yeniden kuruluş 6 piyade ile beklenen orduda oluşturulmadı")
		}
	}
	if !gs.Regions[regionID].HasBuilding("barracks") || !gs.Regions[regionID].HasBuilding("granary") {
		t.Fatalf("ardıl devlet kuruluş bölgesinde kışla ve ambar oluşturulmadı: %#v", gs.Regions[regionID].Buildings)
	}
}

func TestApplySuccessorRevivalAssignsStrongestActiveCommander(t *testing.T) {
	const (
		regionID  = world.RegionID("baghdad")
		successor = faction.FactionID("jelayirids")
	)
	tweak := func(id string, level int) *army.Commander {
		return &army.Commander{ID: id, OwnerID: string(successor), Name: id, Level: level, StartYear: 1300}
	}
	strong := tweak("commander_strong", 4)
	weakCommander := tweak("commander_weak", 2)
	gs := &state.GameState{
		Year: 1335,
		Regions: map[world.RegionID]*world.Region{
			regionID: {ID: regionID, OwnerID: "ilkhanate", Settlements: []world.Settlement{{ID: "baghdad_main"}}},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ilkhanate": {ID: "ilkhanate"},
			successor:   {ID: successor, IsEliminated: true, CapitalSettlementID: "baghdad_main"},
		},
		Armies:     map[army.ArmyID]*army.Army{},
		Commanders: map[string]*army.Commander{},
		CommanderTemplates: map[string][]*army.Commander{
			string(successor): {weakCommander, strong},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Category: army.CategoryInfantry},
		},
	}

	Apply(gs, &Event{
		Target:          "specific_faction",
		AffectedFaction: "ilkhanate",
		SuccessorRevival: &SuccessorRevivalEffect{
			FactionID: string(successor), RegionID: string(regionID), UnitType: "infantry", UnitCount: 6,
		},
	})

	for _, current := range gs.Armies {
		if current.Commander == nil || current.Commander.ID != strong.ID {
			t.Fatalf("dirilen orduya en güçlü aktif komutan atanmadı: %#v", current.Commander)
		}
		if current.Commander.AssignedArmyID != current.ID {
			t.Fatalf("komutanın ordu bağlantısı kurulmadı: %q != %q", current.Commander.AssignedArmyID, current.ID)
		}
	}
	if !gs.Regions[regionID].HasBuilding("barracks") || !gs.Regions[regionID].HasBuilding("granary") {
		t.Fatalf("komutanlı ardıl kuruluşunda kışla ve ambar oluşturulmadı")
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

func TestApplyOtherIncomeDeltaPersistsOnFaction(t *testing.T) {
	const ownerID = faction.FactionID("portugal")
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			ownerID: {ID: ownerID},
		},
	}

	Apply(gs, &Event{
		Target:           "specific_faction",
		AffectedFaction:  string(ownerID),
		OtherIncomeDelta: 25,
	})
	ApplyChoice(gs, &Event{
		Target:          "specific_faction",
		AffectedFaction: string(ownerID),
		Choices: []Choice{{Effect: Effect{
			OtherIncomeDelta: -5,
		}}},
	}, 0)

	if got := gs.Factions[ownerID].OtherIncomeDelta; got != 20 {
		t.Fatalf("kalıcı diğer gelir deltası = %d, 20 bekleniyordu", got)
	}
}

func TestApplyDynasticSettlementTransfersDowryAndFinalUnion(t *testing.T) {
	const (
		ottomanID  = faction.FactionID("ottoman")
		germiyanID = faction.FactionID("germiyan_bey")
	)
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"kutahya":  {ID: "kutahya", OwnerID: string(germiyanID)},
			"germiyan": {ID: "germiyan", OwnerID: string(germiyanID)},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			ottomanID:  {ID: ottomanID},
			germiyanID: {ID: germiyanID, Gold: 100, Grain: 100, Iron: 10, Timber: 10, Stone: 10, Spice: 10, Cloth: 10},
		},
		Relations: map[string]*faction.Relation{},
		Armies: map[army.ArmyID]*army.Army{
			"army_a": {ID: "army_a", OwnerID: string(germiyanID)},
			"army_b": {ID: "army_b", OwnerID: string(germiyanID)},
			"army_c": {ID: "army_c", OwnerID: string(germiyanID)},
			"fleet":  {ID: "fleet", OwnerID: string(germiyanID), IsNaval: true},
		},
	}

	Apply(gs, &Event{
		Target:          "specific_faction",
		AffectedFaction: string(germiyanID),
		DynasticSettlement: &DynasticSettlementEffect{
			SourceFactionID:         string(germiyanID),
			RecipientFactionID:      string(ottomanID),
			RegionIDs:               []world.RegionID{"kutahya"},
			ArmyTransferPercent:     30,
			ResourceTransferPercent: 30,
			RelationStance:          "allied",
			RelationScoreDelta:      50,
		},
	})

	if gs.Regions["kutahya"].OwnerID != string(ottomanID) || gs.Regions["germiyan"].OwnerID != string(germiyanID) {
		t.Fatal("çeyiz yalnız Kütahya'yı Osmanlı'ya devretmedi")
	}
	if got := gs.Factions[germiyanID].Gold; got != 70 || gs.Factions[ottomanID].Gold != 30 {
		t.Fatalf("çeyiz altın aktarımı = Germiyan %d, Osmanlı %d; 70/30 bekleniyordu", got, gs.Factions[ottomanID].Gold)
	}
	if gs.Armies["army_a"].OwnerID != string(ottomanID) || gs.Armies["army_b"].OwnerID != string(germiyanID) || gs.Armies["fleet"].OwnerID != string(germiyanID) {
		t.Fatal("çeyiz yüzde 30 kara ordusu aktarımını deterministik uygulamadı veya filoyu aktardı")
	}
	relation := gs.Relations[faction.RelationKey(ottomanID, germiyanID)]
	if relation == nil || relation.Stance != faction.StanceAllied || relation.ScoreFrom(relation.FactionA) != 50 {
		t.Fatalf("çeyiz ilişkisi müttefik olarak kurulmadı: %+v", relation)
	}

	Apply(gs, &Event{
		Target:          "specific_faction",
		AffectedFaction: string(germiyanID),
		DynasticSettlement: &DynasticSettlementEffect{
			SourceFactionID:          string(germiyanID),
			RecipientFactionID:       string(ottomanID),
			RegionIDs:                []world.RegionID{"germiyan"},
			AutoUnionWhenSourceEmpty: true,
			UnionResultFactionID:     string(ottomanID),
		},
	})

	if gs.Regions["germiyan"].OwnerID != string(ottomanID) || !gs.Factions[germiyanID].IsEliminated {
		t.Fatal("vasiyet sonrası Germiyan tamamen Osmanlı'ya bağlanmadı")
	}
	if got := gs.Factions[ottomanID].Gold; got != 100 {
		t.Fatalf("vasiyet sonrası kalan Germiyan kaynağı Osmanlı'ya aktarılmadı: %d", got)
	}
}

func TestIsPlayerRelevantIncludesDynasticSettlementRecipient(t *testing.T) {
	const (
		ottomanID  = faction.FactionID("ottoman")
		germiyanID = faction.FactionID("germiyan_bey")
	)
	gs := &state.GameState{PlayerFactionID: ottomanID}
	event := &Event{
		Target:          "specific_faction",
		AffectedFaction: string(germiyanID),
		DynasticSettlement: &DynasticSettlementEffect{
			SourceFactionID:    string(germiyanID),
			RecipientFactionID: string(ottomanID),
		},
	}
	if !IsPlayerRelevant(gs, event) {
		t.Fatal("hanedan aktarımının alıcısı olan oyuncu event'i ilgili görmüyor")
	}
}

func TestIsPlayerRelevantIncludesArmyDefectionRecipient(t *testing.T) {
	const recipientID = faction.FactionID("jelayirids")
	gs := &state.GameState{PlayerFactionID: recipientID}
	event := &Event{
		Target:          "specific_faction",
		AffectedFaction: "ilkhanate",
		ArmyDefections: []ArmyDefectionEffect{{
			SourceFactionID:    "ilkhanate",
			RecipientFactionID: string(recipientID),
			ArmyCount:          1,
		}},
	}
	if !IsPlayerRelevant(gs, event) {
		t.Fatal("ordu saf değişiminin alıcısı olan oyuncu event'i ilgili görmüyor")
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

func TestApplyArmyDefectionTransfersDeterministicLandArmyAndCommander(t *testing.T) {
	const (
		sourceID    = faction.FactionID("ilkhanate")
		recipientID = faction.FactionID("jelayirids")
	)
	commander := &army.Commander{
		ID:             "commander_defector",
		OwnerID:        string(sourceID),
		Name:           "Sınır Komutanı",
		AssignedArmyID: "army_a",
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"sivas":   {ID: "sivas", OwnerID: string(sourceID)},
			"baghdad": {ID: "baghdad", OwnerID: string(recipientID)},
			"sea":     {ID: "sea", IsSea: true},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			sourceID:    {ID: sourceID},
			recipientID: {ID: recipientID},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_b": {
				ID:       "army_b",
				OwnerID:  string(sourceID),
				RegionID: "sivas",
			},
			"army_a": {
				ID:         "army_a",
				OwnerID:    string(sourceID),
				RegionID:   "sivas",
				MovePoints: 2,
				Commander:  commander,
			},
			"fleet": {
				ID:       "fleet",
				OwnerID:  string(sourceID),
				RegionID: "sea",
				IsNaval:  true,
			},
		},
	}

	Apply(gs, &Event{
		Target:          "specific_faction",
		AffectedFaction: string(sourceID),
		ArmyDefections: []ArmyDefectionEffect{{
			SourceFactionID:     string(sourceID),
			RecipientFactionID:  string(recipientID),
			SourceRegionIDs:     []world.RegionID{"sivas"},
			DestinationRegionID: "baghdad",
			ArmyCount:           1,
		}},
	})

	defected := gs.Armies["army_a"]
	if defected.OwnerID != string(recipientID) || defected.RegionID != "baghdad" || defected.PreviousRegionID != "sivas" || defected.MovePoints != 0 {
		t.Fatalf("deterministik seçilen ordu doğru aktarılmadı: %+v", defected)
	}
	if commander.OwnerID != string(recipientID) || commander.AssignedArmyID != "army_a" {
		t.Fatalf("orduyla birlikte komutan sahipliği aktarılmadı: %+v", commander)
	}
	if gs.Armies["army_b"].OwnerID != string(sourceID) {
		t.Fatal("seçilmeyen kara ordusu aktarılmış")
	}
	if gs.Armies["fleet"].OwnerID != string(sourceID) {
		t.Fatal("include_naval kapalıyken filo aktarılmış")
	}
}
