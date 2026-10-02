package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestCaptureHeadlessCheckpointIncludesNavalMilitaryPower(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"venice": {ID: "venice", NameTR: "Venedik"},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Attack: 40, HP: 100},
			"warship":  {ID: "warship", Attack: 70, HP: 100},
		},
		Armies: map[army.ArmyID]*army.Army{
			"venice_land": {
				ID:      "venice_land",
				OwnerID: "venice",
				Units:   []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
			},
			"venice_fleet": {
				ID:      "venice_fleet",
				OwnerID: "venice",
				IsNaval: true,
				Units:   []army.Unit{{TypeID: "warship", CurrentHP: 100}},
			},
		},
	}

	checkpoints := captureHeadlessCheckpoint(gs, 10)
	if len(checkpoints.Rankings) != 1 {
		t.Fatalf("tek devlet checkpoint'te görünmeli: %+v", checkpoints.Rankings)
	}
	ranking := checkpoints.Rankings[0]
	landPower := gs.EffectiveArmyStrength(gs.Armies["venice_land"])
	navalPower := gs.EffectiveArmyStrength(gs.Armies["venice_fleet"])
	if ranking.MilitaryPower != landPower+navalPower {
		t.Fatalf("checkpoint toplam gücü kara ve donanmayı içermiyor: got=%d want=%d", ranking.MilitaryPower, landPower+navalPower)
	}
	if ranking.NavalUnits != 1 {
		t.Fatalf("checkpoint donanma birimi sayısını kaydetmedi: got=%d", ranking.NavalUnits)
	}
}

func TestCompleteBuildingAddsNamedPortSettlementToMinorRegion(t *testing.T) {
	minor := &world.Region{
		ID:            "sinop_castle",
		IsMinorRegion: true,
		IsPrivileged:  true,
		Neighbors:     []world.RegionID{"black_sea"},
		WorldX:        100,
		WorldY:        200,
	}
	sea := &world.Region{
		ID:     "black_sea",
		IsSea:  true,
		WorldX: 130,
		WorldY: 200,
	}
	gs := &state.GameState{
		PrivilegedBuildingMaxLevel: 1,
		Regions: map[world.RegionID]*world.Region{
			minor.ID: minor,
			sea.ID:   sea,
		},
		BuildingTypes: map[string]*city.Building{
			"port":   {ID: "port", MaxPerRegion: 3, MinorRegions: true},
			"market": {ID: "market", MaxPerRegion: 3},
		},
	}

	g := &Game{gs: gs}
	if !g.completeBuilding(minor, "port") {
		t.Fatal("minor bölgedeki liman binası tamamlanmadı")
	}
	if len(minor.Settlements) != 1 {
		t.Fatalf("liman tamamlanınca %d yerleşim oluştu, 1 bekleniyordu", len(minor.Settlements))
	}
	port := minor.Settlements[0]
	if port.Type != world.SettlementPort || port.NameTR != "Liman" || port.Name != "Port" {
		t.Fatalf("liman yerleşimi yanlış: %#v", port)
	}
	if port.X != 110 || port.Y != 200 {
		t.Fatalf("liman kıyı konumu = (%d, %d), (%d, %d) bekleniyordu", port.X, port.Y, 110, 200)
	}
	if g.completeBuilding(minor, "port") {
		t.Fatal("minor bölge ikinci liman seviyesini tamamladı; seviye 1 tavanı aşılmamalı")
	}
	if g.completeBuilding(minor, "market") {
		t.Fatal("minor_regions=false olan bina minor bölgede tamamlandı")
	}
}

func TestCompleteBuildingPlacesPortAwayFromExistingSettlementMarker(t *testing.T) {
	minor := &world.Region{
		ID:            "sinop_castle",
		IsMinorRegion: true,
		Neighbors:     []world.RegionID{"black_sea"},
		WorldX:        100,
		WorldY:        200,
		Settlements: []world.Settlement{{
			ID:   "castle",
			Type: world.SettlementFortress,
			X:    110,
			Y:    200,
		}},
	}
	sea := &world.Region{
		ID:     "black_sea",
		IsSea:  true,
		WorldX: 130,
		WorldY: 200,
	}
	gs := &state.GameState{
		PrivilegedBuildingMaxLevel: 1,
		Regions: map[world.RegionID]*world.Region{
			minor.ID: minor,
			sea.ID:   sea,
		},
		BuildingTypes: map[string]*city.Building{
			"port": {ID: "port", MaxPerRegion: 1, MinorRegions: true},
		},
	}

	if !(&Game{gs: gs}).completeBuilding(minor, "port") {
		t.Fatal("liman binası tamamlanmadı")
	}
	if len(minor.Settlements) != 2 {
		t.Fatalf("liman yerleşimi eklenmedi: %#v", minor.Settlements)
	}
	port := minor.Settlements[1]
	if port.X == 110 && port.Y == 200 {
		t.Fatalf("liman mevcut marker'ın üzerine yerleştirildi: (%d, %d)", port.X, port.Y)
	}
	if !(&world.Region{Settlements: minor.Settlements[:1]}).SettlementPositionClear(port.X, port.Y, portSettlementMarkerClearance) {
		t.Fatalf("liman mevcut marker'a çok yakın: (%d, %d)", port.X, port.Y)
	}
}

func TestScenarioExportNeighborsRestoresSourceAreaLinks(t *testing.T) {
	region := &world.Region{
		ID:                "parent",
		Neighbors:         []world.RegionID{"neighbor"},
		AreaNeighborOrder: []world.RegionID{"area::first", "area::second"},
	}

	got := scenarioExportNeighbors(region)
	want := []world.RegionID{"neighbor", "area::first", "area::second"}
	if len(got) != len(want) {
		t.Fatalf("neighbor sayısı = %d, %d bekleniyordu: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("neighbor[%d] = %q, %q bekleniyordu", i, got[i], want[i])
		}
	}
}

func TestWriteScenarioFileIfChangedSkipsIdenticalData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "country_shapes.json")
	data := []byte("{\"id\":\"shapes\"}\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	writtenAt := time.Unix(123, 456)
	if err := os.Chtimes(path, writtenAt, writtenAt); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := writeScenarioFileIfChanged(path, data); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(before.ModTime()) {
		t.Fatalf("aynı içerik için dosya yeniden yazıldı: modtime = %v, %v bekleniyordu", info.ModTime(), before.ModTime())
	}
}

func TestWriteScenarioRelationsPreservesRelationOrder(t *testing.T) {
	scenarioPath := t.TempDir()
	dataDir := filepath.Join(scenarioPath, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}

	gs := &state.GameState{
		ScenarioPath: scenarioPath,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a"},
			"b": {ID: "b"},
			"c": {ID: "c"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("b", "c"): {FactionA: "b", FactionB: "c", Stance: faction.StancePeace},
			faction.RelationKey("a", "c"): {FactionA: "a", FactionB: "c", Stance: faction.StancePeace},
			faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b", Stance: faction.StancePeace},
		},
		RelationOrder: []string{"b|c", "a|c", "a|b", "a|b"},
	}
	if err := writeScenarioRelations(gs); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dataDir, "relations.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := "[\n  {\n    \"faction_a\": \"b\",\n    \"faction_b\": \"c\",\n    \"score_a_to_b\": 0,\n    \"score_b_to_a\": 0,\n    \"stance\": \"peace\"\n  },\n  {\n    \"faction_a\": \"a\",\n    \"faction_b\": \"c\",\n    \"score_a_to_b\": 0,\n    \"score_b_to_a\": 0,\n    \"stance\": \"peace\"\n  },\n  {\n    \"faction_a\": \"a\",\n    \"faction_b\": \"b\",\n    \"score_a_to_b\": 0,\n    \"score_b_to_a\": 0,\n    \"stance\": \"peace\"\n  }\n]\n"
	if string(data) != want {
		t.Fatalf("ilişki kaynak sırası korunmadı:\n%s", data)
	}
}

func TestWriteScenarioRelationsSkipsRuntimeDefaultRelations(t *testing.T) {
	scenarioPath := t.TempDir()
	dataDir := filepath.Join(scenarioPath, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}

	gs := &state.GameState{
		ScenarioPath: scenarioPath,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a"},
			"b": {ID: "b"},
			"c": {ID: "c"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"): {
				FactionA: "a", FactionB: "b", ScoreAToB: -40, ScoreBToA: -40, Stance: faction.StancePeace,
			},
			// a|c, kaynakta özel ilişki olmayan runtime varsayılanıdır.
			faction.RelationKey("a", "c"): {
				FactionA: "a", FactionB: "c", ScoreAToB: -30, ScoreBToA: -30, Stance: faction.StancePeace,
			},
		},
		RelationOrder: []string{faction.RelationKey("a", "b")},
	}

	if err := writeScenarioRelations(gs); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dataDir, "relations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var relations []faction.Relation
	if err := json.Unmarshal(data, &relations); err != nil {
		t.Fatal(err)
	}
	if len(relations) != 1 {
		t.Fatalf("runtime varsayılan ilişkiler kaynak dosyaya yazıldı: %#v", relations)
	}
	if got := faction.RelationKey(relations[0].FactionA, relations[0].FactionB); got != faction.RelationKey("a", "b") {
		t.Fatalf("açık ilişki yerine yanlış kayıt yazıldı: %q", got)
	}
}

func TestWriteScenarioRelationsPreservesUnchangedSourceData(t *testing.T) {
	scenarioPath := t.TempDir()
	dataDir := filepath.Join(scenarioPath, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}

	source := []byte(`[
  {
    "faction_a": "b",
    "faction_b": "c",
    "score_a_to_b": 12,
    "score_b_to_a": -8,
    "stance": "trade"
  },
  {
    "faction_a": "a",
    "faction_b": "c",
    "score_a_to_b": -35,
    "score_b_to_a": -35,
    "stance": "peace"
  },
  {
    "faction_a": "c",
    "faction_b": "a",
    "score_a_to_b": 18,
    "score_b_to_a": -12,
    "stance": "trade"
  }
]
`)
	path := filepath.Join(dataDir, "relations.json")
	if err := os.WriteFile(path, source, 0o644); err != nil {
		t.Fatal(err)
	}

	factions := map[faction.FactionID]*faction.Faction{
		"a": {ID: "a"},
		"b": {ID: "b"},
		"c": {ID: "c"},
	}
	relations, order, err := faction.LoadRelationsWithOrder(path, factions)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeScenarioRelations(&state.GameState{
		ScenarioPath:  scenarioPath,
		Factions:      factions,
		Relations:     relations,
		RelationOrder: order,
	}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(source) {
		t.Fatalf("değişiklik yapılmayan relations.json yeniden biçimlendirildi:\n%s", got)
	}
}

func TestApplyProductionTicksCancelsLandUnitWithoutBarracks(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"minor": {ID: "minor", OwnerID: "player", IsMinorRegion: true},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia": {
				ID: "militia", Category: army.CategoryInfantry, TurnsRequired: 1,
				RequiredBuildings: []army.BuildingRequirement{{ID: "barracks", Level: 1}},
			},
		},
		ProductionQueue: []state.ProductionOrder{{
			Kind: productionKindUnit, FactionID: "player", RegionID: "minor", TypeID: "militia", TurnsLeft: 1,
		}},
	}

	results := (&Game{gs: gs}).applyProductionTicks()
	if len(gs.ProductionQueue) != 0 {
		t.Fatalf("kışlasız kara birimi üretim emri kuyrukta kaldı: %#v", gs.ProductionQueue)
	}
	if len(results) != 1 || !results[0].canceled {
		t.Fatalf("kışlasız üretim emri iptal sonucu üretmedi: %#v", results)
	}
}

func TestRevokeMinorPrivilegeTransfersRegionAndEvictsForces(t *testing.T) {
	gs := privilegeRevocationFixture(true)
	g := &Game{gs: gs}

	g.revokeMinorPrivilege("galata")

	minor := gs.Regions["galata"]
	if minor.IsPrivileged {
		t.Fatal("minor bölgenin imtiyazı kaldırılmadı")
	}
	if minor.OwnerID != "east_rome" {
		t.Fatalf("minor bölge sahibi = %q, east_rome bekleniyordu", minor.OwnerID)
	}
	if len(gs.ProductionQueue) != 0 {
		t.Fatalf("bölgenin üretim emirleri temizlenmedi: %#v", gs.ProductionQueue)
	}
	if got := gs.Armies["genoa_army"].RegionID; got != "genoa_home" {
		t.Fatalf("kara ordusu = %q, en yakın Ceneviz bölgesi bekleniyordu", got)
	}
	fleet := gs.Armies["genoa_fleet"]
	if fleet.RegionID != "black_sea" || fleet.DockedRegionID != "" || fleet.DockedSettlementID != "" {
		t.Fatalf("filo denize çıkarılmadı: %#v", fleet)
	}
	if got := gs.Armies["rome_garrison"].RegionID; got != "galata" {
		t.Fatalf("egemen devletin garnizonu dışlandı: %q", got)
	}
	if gs.Factions["genoa"].IsEliminated {
		t.Fatal("hala toprağı olan kullanım sahibi devlet elendi")
	}
}

func TestRevokeMinorPrivilegeEliminatesOwnerWithoutLand(t *testing.T) {
	gs := privilegeRevocationFixture(false)
	g := &Game{gs: gs}

	g.revokeMinorPrivilege("galata")

	if got := gs.Regions["galata"].OwnerID; got != "east_rome" {
		t.Fatalf("minor bölge sahibi = %q, east_rome bekleniyordu", got)
	}
	if !gs.Factions["genoa"].IsEliminated {
		t.Fatal("son toprağını kaybeden kullanım sahibi devlet elenmedi")
	}
	for id, currentArmy := range gs.Armies {
		if currentArmy != nil && currentArmy.OwnerID == "genoa" {
			t.Fatalf("eliminasyon sonrası Ceneviz ordusu kaldı: %s", id)
		}
	}
}

func privilegeRevocationFixture(includeOtherLand bool) *state.GameState {
	regions := map[world.RegionID]*world.Region{
		"constantinople": {ID: "constantinople", OwnerID: "east_rome", WorldX: 0, WorldY: 0},
		"galata": {
			ID: "galata", NameTR: "Galata", OwnerID: "genoa", IsMinorRegion: true,
			IsPrivileged: true, ParentRegionID: "constantinople", Neighbors: []world.RegionID{"black_sea"},
			WorldX: 100, WorldY: 100,
		},
		"black_sea": {ID: "black_sea", IsSea: true, WorldX: 130, WorldY: 100},
	}
	if includeOtherLand {
		regions["genoa_home"] = &world.Region{ID: "genoa_home", OwnerID: "genoa", WorldX: 120, WorldY: 100}
	}

	return &state.GameState{
		PlayerFactionID: "east_rome",
		Regions:         regions,
		Factions: map[faction.FactionID]*faction.Faction{
			"east_rome": {ID: "east_rome", NameTR: "Doğu Roma"},
			"genoa":     {ID: "genoa", NameTR: "Ceneviz"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"genoa_army": {
				ID: "genoa_army", OwnerID: "genoa", RegionID: "galata",
				Units: []army.Unit{{TypeID: "infantry"}},
			},
			"genoa_fleet": {
				ID: "genoa_fleet", OwnerID: "genoa", RegionID: "black_sea",
				DockedRegionID: "galata", DockedSettlementID: "galata_port", IsNaval: true,
				Units: []army.Unit{{TypeID: "galley"}},
			},
			"rome_garrison": {
				ID: "rome_garrison", OwnerID: "east_rome", RegionID: "galata",
				Units: []army.Unit{{TypeID: "infantry"}},
			},
		},
		ProductionQueue: []state.ProductionOrder{{ID: "build_galata", RegionID: "galata", FactionID: "genoa"}},
	}
}
