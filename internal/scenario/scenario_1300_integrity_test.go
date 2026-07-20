package scenario

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

type scenarioEventRelationReference struct {
	FactionID string `json:"faction_id"`
}

type scenarioEventEffectReference struct {
	AffectedFaction   string                           `json:"affected_faction"`
	CompleteTechs     []string                         `json:"complete_techs"`
	StartResearchTech string                           `json:"start_research_tech"`
	Relations         []scenarioEventRelationReference `json:"relations"`
}

type scenarioEventChoiceReference struct {
	Effect scenarioEventEffectReference `json:"effect"`
}

type scenarioEventRelationRequirementReference struct {
	FactionID string `json:"faction_id"`
}

type scenarioEventReference struct {
	ID                   string                                      `json:"id"`
	AffectedFaction      string                                      `json:"affected_faction"`
	RequiresOwnedRegions []world.RegionID                            `json:"requires_owned_regions"`
	RequiresTechs        []string                                    `json:"requires_techs"`
	BlocksTechs          []string                                    `json:"blocks_techs"`
	RelationRequirements []scenarioEventRelationRequirementReference `json:"relation_requirements"`
	Choices              []scenarioEventChoiceReference              `json:"choices"`
}

func scenario1300IntegrityPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "assets", "scenarios", "1300_ottoman_rise")
}

func read1300JSON(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s okunamadi: %v", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("%s parse edilemedi: %v", path, err)
	}
}

func Test1300EarlyEconomyTechnologyValuesAreCalibrated(t *testing.T) {
	scenarioPath := scenario1300IntegrityPath(t)
	technologies, err := tech.LoadTechnologies(filepath.Join(scenarioPath, "data", "technologies.json"))
	if err != nil {
		t.Fatalf("1300 teknolojileri yüklenemedi: %v", err)
	}
	expected := map[string]struct {
		goldPerRegion int
		marketGoldMod float64
	}{
		"trade_routes":         {2, 0},
		"banking":              {3, 0.10},
		"guilds":               {0, 0.75},
		"tax_registers":        {3, 0},
		"caravanserai_network": {2, 0.15},
		"mint_standardization": {4, 0},
	}
	for technologyID, want := range expected {
		technology := technologies[technologyID]
		if technology == nil {
			t.Errorf("erken ekonomi teknolojisi eksik: %s", technologyID)
			continue
		}
		if technology.Effects.GoldPerRegion != want.goldPerRegion || technology.Effects.MarketGoldMod != want.marketGoldMod {
			t.Errorf("ekonomi teknolojisi değeri kalibre değil: tech=%s gold_per_region=%d/%d market_gold_mod=%.2f/%.2f", technologyID, technology.Effects.GoldPerRegion, want.goldPerRegion, technology.Effects.MarketGoldMod, want.marketGoldMod)
		}
	}
}

func load1300IntegrityData(t *testing.T) (string, map[world.RegionID]*world.Region, map[faction.FactionID]*faction.Faction) {
	t.Helper()
	scenarioPath := scenario1300IntegrityPath(t)
	dataPath := filepath.Join(scenarioPath, "data")

	regions, err := world.LoadRegions(filepath.Join(dataPath, "regions.json"))
	if err != nil {
		t.Fatalf("1300 bölgeleri yüklenemedi: %v", err)
	}
	if err := world.LoadRegionSettlements(filepath.Join(dataPath, "settlements.json"), regions); err != nil {
		t.Fatalf("1300 yerleşimleri yüklenemedi: %v", err)
	}
	factions, err := faction.LoadFactions(filepath.Join(dataPath, "factions.json"))
	if err != nil {
		t.Fatalf("1300 devletleri yüklenemedi: %v", err)
	}
	return scenarioPath, regions, factions
}

func Test1300ScenarioArmyReferencesExist(t *testing.T) {
	scenarioPath, regions, factions := load1300IntegrityData(t)
	dataPath := filepath.Join(scenarioPath, "data")
	unitTypes, err := army.LoadUnitTypes(filepath.Join(dataPath, "units.json"))
	if err != nil {
		t.Fatalf("1300 birlik tipleri yüklenemedi: %v", err)
	}
	armies, err := army.LoadArmies(filepath.Join(dataPath, "armies.json"), unitTypes)
	if err != nil {
		t.Fatalf("1300 orduları yüklenemedi: %v", err)
	}

	for armyID, candidate := range armies {
		if candidate == nil {
			t.Errorf("nil ordu kaydi: %s", armyID)
			continue
		}
		if factions[faction.FactionID(candidate.OwnerID)] == nil {
			t.Errorf("ordu bilinmeyen devlete bağlı: army=%s faction=%s", armyID, candidate.OwnerID)
		}
		if regions[candidate.RegionID] == nil {
			t.Errorf("ordu bilinmeyen bölgeye bağlı: army=%s region=%s", armyID, candidate.RegionID)
		}
		for _, unit := range candidate.Units {
			if unitTypes[unit.TypeID] == nil {
				t.Errorf("ordu bilinmeyen birlik tipi taşıyor: army=%s unit=%s", armyID, unit.TypeID)
			}
		}
	}
}

func Test1300ScenarioStartingNaviesAreDockedAtHistoricalPorts(t *testing.T) {
	scenarioPath, regions, factions := load1300IntegrityData(t)
	dataPath := filepath.Join(scenarioPath, "data")
	unitTypes, err := army.LoadUnitTypes(filepath.Join(dataPath, "units.json"))
	if err != nil {
		t.Fatalf("1300 birlik tipleri yüklenemedi: %v", err)
	}
	armies, err := army.LoadArmies(filepath.Join(dataPath, "armies.json"), unitTypes)
	if err != nil {
		t.Fatalf("1300 orduları yüklenemedi: %v", err)
	}

	expected := map[army.ArmyID]struct {
		owner string
		sea   world.RegionID
		dock  world.RegionID
	}{
		"fleet_venice_guard":      {owner: "venice", sea: "adriatic_sea_dalmatia", dock: "venice"},
		"fleet_venice_trade":      {owner: "venice", sea: "adriatic_sea_dalmatia", dock: "venice"},
		"fleet_genoa_guard":       {owner: "genoa", sea: "new_region_246", dock: "genoa"},
		"fleet_genoa_trade":       {owner: "genoa", sea: "new_region_246", dock: "genoa"},
		"fleet_east_rome_guard":   {owner: "east_rome", sea: "new_region_238", dock: "constantinople"},
		"fleet_aragon_west":       {owner: "aragon", sea: "mediterranean_open_2", dock: "sicily"},
		"fleet_england_channel":   {owner: "england", sea: "new_region_250", dock: "london"},
		"fleet_france_channel":    {owner: "france", sea: "new_region_250", dock: "normandy"},
		"fleet_portugal_atlantic": {owner: "portugal", sea: "atlantic_open_3", dock: "portugal"},
		"fleet_portugal_trade":    {owner: "portugal", sea: "atlantic_open_3", dock: "portugal"},
		"fleet_mamluk_east_med":   {owner: "mamluk", sea: "eastern_mediterranean", dock: "egypt"},
	}

	for fleetID, want := range expected {
		fleet := armies[fleetID]
		if fleet == nil {
			t.Errorf("beklenen başlangıç filosu eksik: %s", fleetID)
			continue
		}
		if !fleet.IsNaval {
			t.Errorf("başlangıç filosu naval işaretli değil: %s", fleetID)
		}
		if fleet.OwnerID != want.owner || factions[faction.FactionID(fleet.OwnerID)] == nil {
			t.Errorf("başlangıç filosu sahibi hatalı: fleet=%s owner=%s", fleetID, fleet.OwnerID)
		}
		if fleet.RegionID != want.sea {
			t.Errorf("başlangıç filosu yanlış denizde: fleet=%s got=%s want=%s", fleetID, fleet.RegionID, want.sea)
		}
		if fleet.DockedRegionID != want.dock {
			t.Errorf("başlangıç filosu yanlış limana bağlı: fleet=%s got=%s want=%s", fleetID, fleet.DockedRegionID, want.dock)
		}
		dock := regions[want.dock]
		if dock == nil || dock.IsSea || dock.OwnerID != want.owner || !dock.HasPortBuilding() {
			t.Errorf("başlangıç filosu geçerli sahip limanına bağlı değil: fleet=%s dock=%s", fleetID, want.dock)
			continue
		}
		settlementFound := false
		for _, settlement := range dock.Settlements {
			if settlement.ID == fleet.DockedSettlementID && settlement.Type == world.SettlementPort {
				settlementFound = true
				break
			}
		}
		if !settlementFound {
			t.Errorf("başlangıç filosunun dock settlement'ı geçersiz: fleet=%s settlement=%s", fleetID, fleet.DockedSettlementID)
		}
		seaAdjacent := false
		for _, neighborID := range dock.Neighbors {
			if neighborID == fleet.RegionID && regions[neighborID] != nil && regions[neighborID].IsSea {
				seaAdjacent = true
				break
			}
		}
		if !seaAdjacent {
			t.Errorf("başlangıç filosunun denizi limana komşu değil: fleet=%s dock=%s sea=%s", fleetID, want.dock, fleet.RegionID)
		}
		for _, unit := range fleet.Units {
			unitType := unitTypes[unit.TypeID]
			if unitType == nil || (unitType.Category != army.CategoryNavalWar && unitType.Category != army.CategoryNavalTrans && unitType.Category != army.CategoryNavalTrade) {
				t.Errorf("başlangıç filosunda kara birimi var: fleet=%s unit=%s", fleetID, unit.TypeID)
			}
		}
	}
	navalCount := 0
	for _, candidate := range armies {
		if candidate != nil && candidate.IsNaval {
			navalCount++
		}
	}
	if navalCount != len(expected) {
		t.Errorf("beklenmeyen başlangıç filosu sayısı: got=%d want=%d", navalCount, len(expected))
	}
}

func Test1300ScenarioHistoricalDiplomacyAndFlandersVassalage(t *testing.T) {
	scenarioPath, regions, factions := load1300IntegrityData(t)
	if flanders := factions["flanders_county"]; flanders == nil || flanders.OverlordID != "hre" {
		t.Fatalf("Flandre HRE vassalı olarak yüklenmeli: faction=%+v", flanders)
	}
	if flandersRegion := regions["flanders"]; flandersRegion == nil || flandersRegion.OwnerID != "flanders_county" {
		t.Fatalf("Flandre bölgesi vassal fraksiyona ait olmalı: region=%+v", flandersRegion)
	}

	var definitions []*faction.Relation
	read1300JSON(t, filepath.Join(scenarioPath, "data", "relations.json"), &definitions)
	relations := make(map[string]*faction.Relation, len(definitions))
	for _, relation := range definitions {
		if relation != nil {
			relations[faction.RelationKey(relation.FactionA, relation.FactionB)] = relation
		}
	}

	wars := [][2]faction.FactionID{
		{"ottoman", "east_rome"},
		{"mamluk", "ilkhanate"},
		{"aragon", "castile_kingdom"},
		{"aragon", "naples_kingdom"},
		{"england", "scotland_kingdom"},
		{"france", "hre"},
		{"france", "flanders_county"},
	}
	for _, pair := range wars {
		relation := relations[faction.RelationKey(pair[0], pair[1])]
		if relation == nil || relation.Stance != faction.StanceWar {
			t.Errorf("beklenen tarihsel savaş yok: %s-%s relation=%+v", pair[0], pair[1], relation)
		}
	}
	englandFrance := relations[faction.RelationKey("england", "france")]
	if englandFrance == nil || englandFrance.Stance != faction.StancePeace || englandFrance.Score != -20 {
		t.Errorf("1300 açılışında İngiltere-Fransa savaşı henüz başlamamalı: relation=%+v", englandFrance)
	}
	aragonGranada := relations[faction.RelationKey("aragon", "granada_emirate")]
	if aragonGranada == nil || aragonGranada.Stance != faction.StanceAllied {
		t.Errorf("Aragon-Gırnata başlangıçta müttefik olmalı: relation=%+v", aragonGranada)
	}
}

func Test1300ScenarioHistoricalFrontArmiesAndTechnology(t *testing.T) {
	scenarioPath, regions, factions := load1300IntegrityData(t)
	dataPath := filepath.Join(scenarioPath, "data")
	unitTypes, err := army.LoadUnitTypes(filepath.Join(dataPath, "units.json"))
	if err != nil {
		t.Fatalf("1300 birlik tipleri yüklenemedi: %v", err)
	}
	armies, err := army.LoadArmies(filepath.Join(dataPath, "armies.json"), unitTypes)
	if err != nil {
		t.Fatalf("1300 orduları yüklenemedi: %v", err)
	}

	expectedArmies := map[army.ArmyID]struct {
		owner  string
		region world.RegionID
		units  map[string]int
	}{
		"army_east_rome_anatolia_guard": {"east_rome", "nicomedia", map[string]int{"infantry": 2, "light_cavalry": 1, "militia": 2}},
		"army_mamluk_aleppo_relief":     {"mamluk", "aleppo", map[string]int{"cavalry": 2, "infantry": 3, "light_cavalry": 1, "militia": 2}},
		"army_ilkhanate_mosul_front":    {"ilkhanate", "mosul", map[string]int{"cavalry": 2, "infantry": 2, "light_cavalry": 1, "militia": 2}},
		"army_aragon_sicily":            {"aragon", "sicily", map[string]int{"cavalry": 1, "infantry": 3, "militia": 2}},
		"army_naples_sicily_front":      {"naples_kingdom", "naples", map[string]int{"cavalry": 1, "infantry": 3, "light_cavalry": 1, "militia": 2}},
		"army_france_normandy":          {"france", "normandy", map[string]int{"cavalry": 1, "infantry": 3, "light_cavalry": 1, "militia": 2}},
		"army_england_scotland_front":   {"england", "lancashire", map[string]int{"cavalry": 1, "infantry": 3, "militia": 2}},
		"army_scotland_border":          {"scotland_kingdom", "scotland", map[string]int{"infantry": 2, "light_cavalry": 1, "militia": 3}},
		"army_castile_murcia_front":     {"castile_kingdom", "new_region_234", map[string]int{"cavalry": 1, "infantry": 3, "light_cavalry": 1, "militia": 2}},
		"army_flanders_bruges_guard":    {"flanders_county", "flanders", map[string]int{"infantry": 2, "light_cavalry": 1, "militia": 3}},
		"army_hre_flanders_relief":      {"hre", "holland", map[string]int{"cavalry": 2, "infantry": 3, "militia": 2}},
	}
	for armyID, want := range expectedArmies {
		candidate := armies[armyID]
		if candidate == nil {
			t.Errorf("beklenen tarihsel cephe ordusu eksik: %s", armyID)
			continue
		}
		if candidate.OwnerID != want.owner || candidate.RegionID != want.region || regions[candidate.RegionID] == nil {
			t.Errorf("cephe ordusu konumu/sahibi hatalı: army=%s owner=%s region=%s", armyID, candidate.OwnerID, candidate.RegionID)
		}
		gotUnits := make(map[string]int, len(candidate.Units))
		for _, unit := range candidate.Units {
			gotUnits[unit.TypeID]++
		}
		for typeID, count := range want.units {
			if gotUnits[typeID] != count {
				t.Errorf("cephe ordusu bileşimi hatalı: army=%s unit=%s got=%d want=%d", armyID, typeID, gotUnits[typeID], count)
			}
		}
	}

	expectedTech := map[faction.FactionID][]string{
		"east_rome":       {"navigation", "harbor_administration"},
		"mamluk":          {"cavalry_tactics", "horse_breeding"},
		"ilkhanate":       {"composite_bow_drills", "engineering_corps"},
		"naples_kingdom":  {"navigation", "trade_routes"},
		"flanders_county": {"banking", "harbor_administration", "navigation", "trade_routes"},
	}
	for factionID, technologies := range expectedTech {
		definition := factions[factionID]
		if definition == nil {
			t.Errorf("teknoloji kontrolü için devlet eksik: %s", factionID)
			continue
		}
		for _, technologyID := range technologies {
			if !definition.Research.Completed[technologyID] {
				t.Errorf("tarihsel başlangıç teknolojisi eksik: faction=%s technology=%s", factionID, technologyID)
			}
		}
	}
}

func Test1300ScenarioEventReferencesExist(t *testing.T) {
	scenarioPath, regions, factions := load1300IntegrityData(t)
	dataPath := filepath.Join(scenarioPath, "data")
	var eventsList []*scenarioEventReference
	read1300JSON(t, filepath.Join(dataPath, "events.json"), &eventsList)
	technologies, err := tech.LoadTechnologies(filepath.Join(dataPath, "technologies.json"))
	if err != nil {
		t.Fatalf("1300 teknolojileri yüklenemedi: %v", err)
	}

	assertFaction := func(eventID, field string, id string) {
		t.Helper()
		if id != "" && factions[faction.FactionID(id)] == nil {
			t.Errorf("event bilinmeyen fraksiyona bağlı: event=%s field=%s faction=%s", eventID, field, id)
		}
	}
	assertRegion := func(eventID, field string, id world.RegionID) {
		t.Helper()
		if id != "" && regions[id] == nil {
			t.Errorf("event bilinmeyen bölgeye bağlı: event=%s field=%s region=%s", eventID, field, id)
		}
	}
	assertTech := func(eventID, field, id string) {
		t.Helper()
		if id != "" && technologies[id] == nil {
			t.Errorf("event bilinmeyen teknolojiye bağlı: event=%s field=%s technology=%s", eventID, field, id)
		}
	}
	for _, event := range eventsList {
		if event == nil {
			continue
		}
		assertFaction(event.ID, "affected_faction", event.AffectedFaction)
		for _, regionID := range event.RequiresOwnedRegions {
			assertRegion(event.ID, "requires_owned_regions", regionID)
		}
		for _, technologyID := range event.RequiresTechs {
			assertTech(event.ID, "requires_techs", technologyID)
		}
		for _, technologyID := range event.BlocksTechs {
			assertTech(event.ID, "blocks_techs", technologyID)
		}
		for _, requirement := range event.RelationRequirements {
			assertFaction(event.ID, "relation_requirements.faction_id", requirement.FactionID)
		}
		for _, choice := range event.Choices {
			for _, relation := range choice.Effect.Relations {
				assertFaction(event.ID, "choice.effect.relations.faction_id", relation.FactionID)
			}
			for _, technologyID := range choice.Effect.CompleteTechs {
				assertTech(event.ID, "choice.effect.complete_techs", technologyID)
			}
			assertTech(event.ID, "choice.effect.start_research_tech", choice.Effect.StartResearchTech)
			assertFaction(event.ID, "choice.effect.affected_faction", choice.Effect.AffectedFaction)
		}
	}
}

func Test1300ScenarioVictoryRegionReferencesExist(t *testing.T) {
	scenarioPath, regions, _ := load1300IntegrityData(t)
	var definition Scenario
	read1300JSON(t, filepath.Join(scenarioPath, "scenario.json"), &definition)

	for _, option := range definition.VictoryConditions {
		for _, regionID := range option.RegionTargets() {
			if regions[world.RegionID(regionID)] == nil {
				t.Errorf("zafer hedefi bilinmeyen bölgeye bağlı: victory=%s region=%s", option.ID, regionID)
			}
		}
	}
}

func Test1300ScenarioFactionReferencesExist(t *testing.T) {
	scenarioPath, _, factions := load1300IntegrityData(t)

	for factionID, definition := range factions {
		for _, targetID := range definition.AIExpansionTargets {
			if factions[targetID] == nil {
				t.Errorf("AI genişleme hedefi bilinmeyen devlete bağlı: faction=%s target=%s", factionID, targetID)
			}
		}
	}

	var relations []*faction.Relation
	read1300JSON(t, filepath.Join(scenarioPath, "data", "relations.json"), &relations)
	for index, relation := range relations {
		if relation == nil {
			t.Errorf("nil ilişki kaydı: index=%d", index)
			continue
		}
		if factions[relation.FactionA] == nil {
			t.Errorf("ilişki bilinmeyen devlete bağlı: index=%d faction_a=%s", index, relation.FactionA)
		}
		if factions[relation.FactionB] == nil {
			t.Errorf("ilişki bilinmeyen devlete bağlı: index=%d faction_b=%s", index, relation.FactionB)
		}
		if relation.FactionA == relation.FactionB {
			t.Errorf("ilişki aynı devleti iki tarafta kullanıyor: index=%d faction=%s", index, relation.FactionA)
		}
	}
}

func Test1300ScenarioCapitalSettlementsExist(t *testing.T) {
	_, regions, factions := load1300IntegrityData(t)
	settlementRegions := make(map[string]world.RegionID)
	for regionID, region := range regions {
		if region == nil {
			continue
		}
		for _, settlement := range region.Settlements {
			if previous, duplicate := settlementRegions[settlement.ID]; duplicate {
				t.Errorf("settlement ID birden fazla bölgede kullanılıyor: settlement=%s first=%s second=%s", settlement.ID, previous, regionID)
				continue
			}
			settlementRegions[settlement.ID] = regionID
		}
	}

	for factionID, definition := range factions {
		if definition.CapitalSettlementID == "" {
			t.Errorf("devletin başkent settlement kimliği boş: faction=%s", factionID)
			continue
		}
		regionID, ok := settlementRegions[definition.CapitalSettlementID]
		if !ok {
			t.Errorf("devletin başkenti bilinmeyen settlement'a bağlı: faction=%s settlement=%s", factionID, definition.CapitalSettlementID)
			continue
		}
		if region := regions[regionID]; region == nil || region.OwnerID != string(factionID) {
			ownerID := ""
			if region != nil {
				ownerID = region.OwnerID
			}
			t.Errorf("başkent settlement'ı farklı devletin bölgesinde: faction=%s settlement=%s region=%s owner=%s", factionID, definition.CapitalSettlementID, regionID, ownerID)
		}
	}
}

func Test1300ScenarioTradeCenterReferencesExist(t *testing.T) {
	scenarioPath, regions, _ := load1300IntegrityData(t)
	var config world.TradeCenterConfig
	read1300JSON(t, filepath.Join(scenarioPath, "data", "trade_centers.json"), &config)

	centers := make(map[world.RegionID]world.TradeCenterDef, len(config.Centers))
	for _, center := range config.Centers {
		if center.ID == "" {
			t.Error("ticaret merkezi kimliği boş")
			continue
		}
		if _, duplicate := centers[center.ID]; duplicate {
			t.Errorf("ticaret merkezi kimliği tekrarlanıyor: center=%s", center.ID)
			continue
		}
		centers[center.ID] = center
		if !center.OffMap && regions[center.ID] == nil {
			t.Errorf("ticaret merkezi bilinmeyen bölgeye bağlı: center=%s", center.ID)
		}
	}

	for _, center := range config.Centers {
		for _, linkID := range center.Links {
			if _, ok := centers[linkID]; !ok {
				t.Errorf("ticaret merkezi bilinmeyen linke bağlı: center=%s link=%s", center.ID, linkID)
			}
			if linkID == center.ID {
				t.Errorf("ticaret merkezi kendisine bağlı: center=%s", center.ID)
			}
		}
	}
}

func Test1300ScenarioAIStrategyReferencesExist(t *testing.T) {
	scenarioPath, regions, factions := load1300IntegrityData(t)
	strategies, err := LoadAIStrategies(filepath.Join(scenarioPath, "data", "ai_strategies.json"))
	if err != nil {
		t.Fatalf("1300 AI stratejileri yüklenemedi: %v", err)
	}
	if len(strategies) == 0 {
		t.Fatal("1300 AI strateji profilleri boş")
	}
	for factionID, strategy := range strategies {
		if factions[faction.FactionID(factionID)] == nil {
			t.Errorf("AI profili bilinmeyen devlete bağlı: faction=%s", factionID)
		}
		for _, objective := range strategy.Objectives {
			for _, targetID := range objective.TargetFactions {
				if factions[faction.FactionID(targetID)] == nil {
					t.Errorf("AI objective bilinmeyen devleti hedefliyor: objective=%s target=%s", objective.ID, targetID)
				}
			}
			for _, regionIDs := range [][]string{objective.TargetRegions, objective.ReadinessRegions, objective.AnnexRegionIDs} {
				for _, regionID := range regionIDs {
					if regions[world.RegionID(regionID)] == nil {
						t.Errorf("AI objective bilinmeyen bölgeye bağlı: objective=%s region=%s", objective.ID, regionID)
					}
				}
			}
		}
	}
}

func Test1300ScenarioProfilesCoverRegionalObjectives(t *testing.T) {
	scenarioPath, _, factions := load1300IntegrityData(t)
	strategies, err := LoadAIStrategies(filepath.Join(scenarioPath, "data", "ai_strategies.json"))
	if err != nil {
		t.Fatalf("1300 bölgesel AI profilleri yüklenemedi: %v", err)
	}

	expected := map[string]struct {
		profile       string
		objectiveID   string
		objectiveKind string
	}{
		"ahiler":           {"central_buffer_survival", "hold_sivrihisar_buffer", "defend"},
		"aydin_bey":        {"aegean_maritime_competition", "contest_saruhan_coast", "expand"},
		"candar_bey":       {"black_sea_frontier", "control_pontic_corridor", "expand"},
		"canik_bey":        {"pontic_buffer_survival", "resist_candar_pressure", "defend"},
		"dulkadir_bey":     {"levant_buffer_survival", "hold_dulkadir_buffer", "defend"},
		"esrefoglu_bey":    {"central_anatolian_survival", "hold_beysehir_pass", "defend"},
		"germiyan_bey":     {"western_anatolian_rival", "contest_hamid_frontier", "expand"},
		"hamid_bey":        {"taurus_frontier_competition", "secure_mentese_passes", "expand"},
		"karaman_bey":      {"central_anatolian_expansion", "press_cilician_frontier", "expand"},
		"karesioglu_bey":   {"marmara_buffer_survival", "hold_marmara_bridgehead", "defend"},
		"mentese_bey":      {"aegean_coastal_survival", "contest_aydin_coast", "expand"},
		"ramazan_bey":      {"cilician_buffer_survival", "hold_cilician_gate", "defend"},
		"saruhan_bey":      {"aegean_interior_survival", "contest_aydinoglu_coast", "expand"},
		"venice":           {"adriatic_merchant_thalassocracy", "protect_adriatic_and_island_trade", "defend"},
		"genoa":            {"western_merchant_network", "protect_ligurian_and_black_sea_trade", "defend"},
		"mamluk":           {"levant_sultanate_frontier", "break_ilkhanate_mesopotamian_front", "expand"},
		"ilkhanate":        {"eastern_imperial_frontier", "press_levant_frontier", "expand"},
		"serbian_empire":   {"balkan_hegemony_buffer", "hold_serbian_mountain_core", "defend"},
		"bulgarian_empire": {"danubian_balkan_defense", "hold_danube_balkan_line", "defend"},
		"epir":             {"epirus_survival", "hold_epirus_thessaly", "defend"},
		"arnavut_des":      {"albanian_mountain_survival", "hold_albanian_mountains", "defend"},
		"athena_duk":       {"aegean_city_state_survival", "hold_athens_coast", "defend"},
		"wallachia_prince": {"danube_buffer_survival", "hold_wallachian_buffer", "defend"},
		"russia":           {"moscow_consolidation", "consolidate_eastern_rus_frontier", "expand"},
		"golden_horde":     {"steppe_hegemony", "press_rus_steppe", "expand"},
		"teutonic_order":   {"baltic_crusader_frontier", "press_lithuanian_frontier", "expand"},
		"novgorod_rep":     {"northern_trade_survival", "hold_novgorod_trade_gate", "defend"},
		"lithuanian_gd":    {"eastern_baltic_expansion", "contest_kievan_steppe", "expand"},
		"england":          {"continental_claim_awakening", "secure_english_channel_and_isles", "consolidate"},
		"france":           {"royal_recovery_after_1337", "protect_french_royal_core", "consolidate"},
		"safavid":          {"ardabil_survival_and_awakening", "hold_southern_persian_core", "consolidate"},
	}
	for factionID, want := range expected {
		if factions[faction.FactionID(factionID)] == nil {
			t.Errorf("bölgesel profil bilinmeyen devlete bağlı: faction=%s", factionID)
			continue
		}
		strategy, ok := strategies[factionID]
		if !ok {
			t.Errorf("bölgesel AI profili eksik: faction=%s", factionID)
			continue
		}
		if strategy.Profile != want.profile {
			t.Errorf("bölgesel profil beklenmiyor: faction=%s got=%s want=%s", factionID, strategy.Profile, want.profile)
		}
		if len(strategy.Objectives) == 0 || strategy.Objectives[0].ID != want.objectiveID || strategy.Objectives[0].Kind != want.objectiveKind {
			t.Errorf("bölgesel profilinin ilk objective'i yanlış: faction=%s got=%+v", factionID, strategy.Objectives)
		}
	}
}
