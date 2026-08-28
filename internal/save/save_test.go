package save

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestLoadFromPathRehydratesScenarioRuntimeFromScenarioID(t *testing.T) {
	tmp := t.TempDir()
	oldBaseDir := scenarioBaseDir
	scenarioBaseDir = filepath.Join(tmp, "scenarios")
	defer func() { scenarioBaseDir = oldBaseDir }()

	scenarioID := "test_scenario"
	scenarioPath := filepath.Join(scenarioBaseDir, scenarioID)
	if err := os.MkdirAll(filepath.Join(scenarioPath, "data"), 0755); err != nil {
		t.Fatalf("scenario data dir olusmadi: %v", err)
	}

	writeJSONFile(t, filepath.Join(scenarioPath, "scenario.json"), scenario.Scenario{
		ID:    scenarioID,
		Year:  1453,
		Month: 4,
		MapConfig: scenario.MapConfig{
			WorldWidth: intPtr(64),
		},
		VictoryConditions: []scenario.VictoryOptionDef{
			{ID: "economic", Type: "economic", TargetGoldIncome: 500, GoldHoldTurns: 5, DeadlineYear: 1500, DeadlineMonth: 6},
			{ID: "other_only", Type: "conquer_city", RequiredRegions: []string{"r1"}, AllowedFactions: []string{"other"}},
		},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "regions.json"), []*world.Region{
		{ID: "r1", NameTR: "R1", OwnerID: "player", ShapeID: "AAA", TaxRate: 50, Satisfaction: 50},
		{ID: "r2", NameTR: "R2", OwnerID: "other", ShapeID: "AAA", TaxRate: 50, Satisfaction: 50},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "land_passages.json"), []world.LandPassage{
		{From: "r1", To: "r2", Type: world.LandPassageStrait, MoveCost: 1, DefenseBonus: 15},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "factions.json"), []*faction.Faction{
		{ID: "player", NameTR: "Oyuncu", IsPlayable: true},
		{ID: "other", NameTR: "Diger", IsPlayable: true},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "units.json"), []map[string]any{
		{"id": "militia", "name": "Militia", "name_tr": "Milis", "category": "infantry", "attack": 5, "defense": 4, "morale": 10, "hp": 100, "gold_cost": 10, "grain_upkeep": 1, "turns_required": 1},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "buildings.json"), []map[string]any{
		{"id": "market", "name": "Market", "name_tr": "Pazar", "gold_cost": 50, "turns_required": 1, "gold_mod": 1.2, "grain_mod": 1.0, "sat_bonus": 0, "def_bonus": 0, "max_per_region": 1},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "technologies.json"), []map[string]any{
		{"id": "tax", "name_tr": "Vergi", "category": "economy", "description_tr": "vergi", "gold_cost": 20, "turns_required": 1, "requires": []string{}, "effects": map[string]any{"gold_per_region": 5}},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "country_shapes.json"), map[string]any{
		"shapes": []map[string]any{
			{"id": "AAA", "name": "AAA", "rings": [][][]int{{{1, 1}, {4, 1}, {4, 4}, {1, 4}}}},
		},
	})

	savePath := filepath.Join(tmp, "slot.json")
	writeJSONFile(t, savePath, &state.GameState{
		ScenarioID:      scenarioID,
		ScenarioPath:    "",
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", NameTR: "R1", OwnerID: "player", ShapeID: "AAA", TaxRate: 50, Satisfaction: 50},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
		},
		Armies: map[army.ArmyID]*army.Army{},
	})

	gs, err := loadFromPath(savePath)
	if err != nil {
		t.Fatalf("loadFromPath hata verdi: %v", err)
	}

	if gs.ScenarioPath != scenarioPath {
		t.Fatalf("scenario path resolve olmadi: got=%q want=%q", gs.ScenarioPath, scenarioPath)
	}
	if len(gs.ScenarioVictories) != 2 {
		t.Fatalf("senaryo victory metadata geri yuklenmedi: %+v", gs.ScenarioVictories)
	}
	if len(gs.AvailableVictories) != 1 || gs.AvailableVictories[0].ID != "economic" {
		t.Fatalf("victory metadata geri yuklenmedi: %+v", gs.AvailableVictories)
	}
	if gs.MapConfig.WorldWidth == nil || *gs.MapConfig.WorldWidth != 64 {
		t.Fatalf("map config geri yuklenmedi: %+v", gs.MapConfig)
	}
	if gs.Turn != 1 || gs.Year != 1453 || gs.Month != 4 {
		t.Fatalf("legacy save eksik zaman alanlarinda senaryo varsayimi korunmadi: turn=%d year=%d month=%d", gs.Turn, gs.Year, gs.Month)
	}
	if len(gs.RegionOrder) != 2 || gs.RegionOrder[0] != "r1" || gs.RegionOrder[1] != "r2" {
		t.Fatalf("region order geri yuklenmedi: %+v", gs.RegionOrder)
	}
	if len(gs.LandPassages) != 1 || !world.HasLandPassage(gs.LandPassages, "r1", "r2") {
		t.Fatalf("land passage senaryo tabanından geri yüklenmedi: %+v", gs.LandPassages)
	}
	if len(gs.FactionOrder) != 2 || gs.FactionOrder[0] != "player" || gs.FactionOrder[1] != "other" {
		t.Fatalf("faction order geri yuklenmedi: %+v", gs.FactionOrder)
	}
	if gs.UnitTypes["militia"] == nil || gs.BuildingTypes["market"] == nil || gs.TechTypes["tax"] == nil {
		t.Fatal("runtime tipleri geri yuklenmedi")
	}
	if len(gs.ShapeData.Shapes["AAA"]) == 0 {
		t.Fatal("shape data geri yuklenmedi")
	}
}

func TestLoadFromPathKeepsOpenSeaFleetUndocked(t *testing.T) {
	oldBaseDir := scenarioBaseDir
	scenarioBaseDir = filepath.Join("..", "..", "assets", "scenarios")
	t.Cleanup(func() { scenarioBaseDir = oldBaseDir })

	savePath := filepath.Join(t.TempDir(), "open-sea.json")
	saved := campaignSaveState{
		ScenarioID:      "1300_ottoman_rise",
		PlayerFactionID: "east_rome",
		Armies: map[army.ArmyID]armySaveState{
			"open_sea_fleet": {
				OwnerID:       "east_rome",
				RegionID:      "aegean_sea",
				IsNaval:       true,
				MovePoints:    3,
				MaxMovePoints: 3,
				Units:         []stackedUnitSaveState{{TypeID: "warship", Count: 1}},
			},
		},
	}
	payload, err := json.Marshal(saved)
	if err != nil {
		t.Fatalf("test save marshal edilemedi: %v", err)
	}
	if err := os.WriteFile(savePath, payload, 0644); err != nil {
		t.Fatalf("test save yazilamadi: %v", err)
	}

	gs, err := loadFromPath(savePath)
	if err != nil {
		t.Fatalf("loadFromPath hata verdi: %v", err)
	}

	fleet := gs.Armies["open_sea_fleet"]
	if fleet == nil {
		t.Fatal("acik deniz filosu save/load sonrasinda kayboldu")
	}
	if fleet.RegionID != "aegean_sea" {
		t.Fatalf("deniz bolgesi korunmadi: got=%q", fleet.RegionID)
	}
	if fleet.DockedRegionID != "" || fleet.DockedSettlementID != "" {
		t.Fatalf("acik deniz filosu save/load sonrasinda limana alindi: docked_region=%q docked_settlement=%q", fleet.DockedRegionID, fleet.DockedSettlementID)
	}
}

func TestInvalidDifficultyMigratesToNormal(t *testing.T) {
	restored := &state.GameState{}
	applyCampaignSaveState(restored, campaignSaveState{Difficulty: 0})
	if restored.Difficulty != 2 {
		t.Fatalf("geçersiz kayıt zorluğu normale göç etmeliydi: got=%d", restored.Difficulty)
	}
}

func TestCompactSaveRestoresVirtualRebelFactionAndArmy(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Regions: map[world.RegionID]*world.Region{
			"border": {ID: "border", NameTR: "Sınır", OwnerID: "player"},
		},
		Relations: map[string]*faction.Relation{},
	}
	rebel := faction.FactionID("rebel_border")
	warStance := encodeStance(faction.StanceWar)
	saved := campaignSaveState{
		Regions: map[world.RegionID]regionSaveState{
			"border": {OwnerID: cloneStringPtr(string(rebel))},
		},
		Factions: map[faction.FactionID]factionSaveState{
			rebel: {IsVirtual: cloneBoolPtr(true)},
		},
		Armies: map[army.ArmyID]armySaveState{
			"rebel_army": {OwnerID: string(rebel), RegionID: "border", IsRebel: true, RebelAgainstID: "player", Units: []stackedUnitSaveState{{TypeID: "militia", Count: 2}}},
		},
		Relations: map[string]relationSaveState{
			faction.RelationKey("player", rebel): {Stance: &warStance},
		},
	}
	applyCampaignSaveState(gs, saved)

	if restored := gs.Factions[rebel]; restored == nil || !restored.IsVirtual {
		t.Fatalf("sanal rebel fraksiyonu geri yüklenmedi: %+v", restored)
	}
	if current := gs.Armies["rebel_army"]; current == nil || !current.IsRebel || current.RebelAgainstID != "player" {
		t.Fatalf("rebel ordu state'i geri yüklenmedi: %+v", current)
	}
}

func TestOldSaveMigratesMissingVirtualRebelFaction(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{"player": {ID: "player"}},
		Regions:  map[world.RegionID]*world.Region{"border": {ID: "border", NameTR: "Sınır", OwnerID: "player"}},
	}
	rebel := faction.FactionID("rebel_border")
	warStance := encodeStance(faction.StanceWar)
	applyCampaignSaveState(gs, campaignSaveState{
		Regions: map[world.RegionID]regionSaveState{"border": {OwnerID: cloneStringPtr(string(rebel))}},
		Armies:  map[army.ArmyID]armySaveState{"rebel_army": {OwnerID: string(rebel), RegionID: "border"}},
		Relations: map[string]relationSaveState{
			faction.RelationKey("player", rebel): {Stance: &warStance},
		},
	})

	if restored := gs.Factions[rebel]; restored == nil || !restored.IsVirtual {
		t.Fatalf("eski save migration sanal rebel fraksiyonu oluşturmadi: %+v", restored)
	}
	if current := gs.Armies["rebel_army"]; current == nil || !current.IsRebel || current.RebelAgainstID != "player" {
		t.Fatalf("eski save migration rebel orduyu tanımadı: %+v", current)
	}
}

func TestImperialStateRoundTripInCompactPayload(t *testing.T) {
	original := &state.ImperialState{
		EmpireID: "hre", EmperorID: "milan_duchy", Authority: 41,
		Members: map[faction.FactionID]*state.ImperialMember{
			"milan_duchy": {FactionID: "milan_duchy", Status: state.ImperialMemberPrince, Loyalty: 37, Autonomy: 88, MilitaryCommitment: 42},
		},
		LastWarCall:     &state.ImperialWarCall{CallerID: "hre", EnemyID: "france", StartedTurn: 7},
		PendingDecision: &state.ImperialPendingDecision{Kind: state.ImperialDecisionDiet, CreatedTurn: 12},
	}
	encoding, payload, err := encodeCompressedStatePayload(campaignSaveState{Imperial: original})
	if err != nil {
		t.Fatalf("imperial compact payload oluşturulamadı: %v", err)
	}
	raw, err := decodeCompressedStatePayload(encoding, payload)
	if err != nil {
		t.Fatalf("imperial compact payload çözülemedi: %v", err)
	}
	var saved campaignSaveState
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatalf("imperial compact payload parse edilemedi: %v", err)
	}
	restored := &state.GameState{}
	applyCampaignSaveState(restored, saved)
	if restored.Imperial == nil || restored.Imperial.EmperorID != "milan_duchy" || restored.Imperial.Authority != 41 {
		t.Fatalf("imperial state compact save/load sonrasında korunmadı: %+v", restored.Imperial)
	}
	if restored.Imperial.Members["milan_duchy"].Loyalty != 37 || restored.Imperial.LastWarCall.EnemyID != "france" {
		t.Fatalf("imperial üye/çağrı state'i korunmadı: %+v", restored.Imperial)
	}
	if restored.Imperial.PendingDecision == nil || restored.Imperial.PendingDecision.Kind != state.ImperialDecisionDiet {
		t.Fatalf("pending imparatorluk kararı compact save/load sonrasında korunmadı: %+v", restored.Imperial.PendingDecision)
	}
}

func TestRegionSuccessorRoundTripInCompactPayload(t *testing.T) {
	base := &world.Region{ID: "r1", OwnerID: "player", SuccessorFactionID: "old_state"}
	current := &world.Region{ID: "r1", OwnerID: "player", SuccessorFactionID: "old_state"}

	current.SuccessorFactionID = "restored_state"
	delta, changed := makeRegionSaveState(current, base)
	if !changed || delta.SuccessorFactionID == nil || *delta.SuccessorFactionID != "restored_state" {
		t.Fatalf("ardıl devlet save delta'sına yazılmadı: changed=%v delta=%+v", changed, delta)
	}

	encoding, payload, err := encodeCompressedStatePayload(campaignSaveState{
		Regions: map[world.RegionID]regionSaveState{"r1": delta},
	})
	if err != nil {
		t.Fatalf("ardıl devlet compact payload oluşturulamadı: %v", err)
	}
	raw, err := decodeCompressedStatePayload(encoding, payload)
	if err != nil {
		t.Fatalf("ardıl devlet compact payload çözülemedi: %v", err)
	}
	var saved campaignSaveState
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatalf("ardıl devlet compact payload parse edilemedi: %v", err)
	}

	restored := *base
	applyRegionSaveState(&restored, saved.Regions["r1"])
	if restored.SuccessorFactionID != "restored_state" {
		t.Fatalf("ardıl devlet save/load sonrasında korunmadı: got=%q", restored.SuccessorFactionID)
	}
}

func TestArmyCommanderRoundTrip(t *testing.T) {
	commander := army.NewCommander("cmd_1", "Mihri Hanım")
	commander.Experience = army.CommanderLevel3XP
	commander.Normalize()
	original := map[army.ArmyID]*army.Army{
		"army_1": {
			ID:        "army_1",
			OwnerID:   "player",
			RegionID:  "r1",
			Commander: commander,
			Units:     []army.Unit{{TypeID: "inf", CurrentHP: 90}},
		},
	}

	restored := restoreArmiesFromSaveState(convertArmiesToSaveState(original))
	got := restored["army_1"]
	if got == nil || got.Commander == nil {
		t.Fatal("komutan save/load sonrasında kayboldu")
	}
	if got.Commander.Name != "Mihri Hanım" || got.Commander.Level != 3 || !got.Commander.HasTrait(army.CommanderTraitTactician) {
		t.Fatalf("komutan state'i eksik geri yüklendi: %+v", got.Commander)
	}
	if got.Commander == commander || &got.Commander.Traits[0] == &commander.Traits[0] {
		t.Fatal("komutan save/load kopyası bağımsız olmalı")
	}
}

func TestArmyMoraleRoundTrip(t *testing.T) {
	original := map[army.ArmyID]*army.Army{
		"army_1": {
			ID: "army_1", OwnerID: "player", RegionID: "r1", Morale: 47,
			Units: []army.Unit{{TypeID: "inf", CurrentHP: 90}},
		},
	}

	restored := restoreArmiesFromSaveState(convertArmiesToSaveState(original))
	if got := restored["army_1"].Morale; got != 47 {
		t.Fatalf("ordu morali compact save/load sonrasında korunmalıydı, got=%d", got)
	}
}

func TestArmyMerchantTradeAssignmentRoundTrip(t *testing.T) {
	original := map[army.ArmyID]*army.Army{
		"merchant_1": {
			ID: "merchant_1", OwnerID: "venice", RegionID: "adriatic", IsNaval: true,
			TradeRouteKey: "venice->mamluk",
			Units:         []army.Unit{{TypeID: "merchant_ship", CurrentHP: 100}},
		},
	}
	restored := restoreArmiesFromSaveState(convertArmiesToSaveState(original))
	if got := restored["merchant_1"]; got == nil || got.TradeRouteKey != "venice->mamluk" {
		t.Fatalf("merchant rota görevi compact save/load sonrasında korunmalıydı: %+v", got)
	}
}

func TestCommanderPoolRoundTripKeepsArmyAssignmentLink(t *testing.T) {
	commander := army.NewCommander("cmd_1", "Mihri Hanım")
	commander.OwnerID = "player"
	commander.AssignedArmyID = "army_1"
	commander.Experience = army.CommanderLevel3XP
	commander.Normalize()
	original := map[army.ArmyID]*army.Army{
		"army_1": {
			ID:        "army_1",
			OwnerID:   "player",
			Commander: commander,
		},
	}

	saved := campaignSaveState{
		Armies:           convertArmiesToSaveState(original),
		Commanders:       cloneCommanders(map[string]*army.Commander{"cmd_1": commander}),
		NextCommanderSeq: 2,
	}
	restored := &state.GameState{}
	applyCampaignSaveState(restored, saved)

	pooled := restored.Commanders["cmd_1"]
	if pooled == nil {
		t.Fatal("komutan havuzu save/load sonrasında kayboldu")
	}
	if restored.Armies["army_1"].Commander != pooled {
		t.Fatal("ordu komutanı havuzdaki canonical pointer'a bağlanmadı")
	}
	if pooled.AssignedArmyID != "army_1" || pooled.OwnerID != "player" {
		t.Fatalf("komutan atama bağı korunmadı: %+v", pooled)
	}
	if restored.NextCommanderSeq != 2 {
		t.Fatalf("komutan sıra sayacı korunmadı: got=%d", restored.NextCommanderSeq)
	}
}

func TestAIPlanStateRoundTripKeepsDurableIntent(t *testing.T) {
	original := &state.AIPlanState{
		ObjectiveID:        "unite_anatolian_beyliks",
		Kind:               state.AIObjectiveExpand,
		TargetFactionID:    "germiyan_bey",
		TargetRegionIDs:    []world.RegionID{"germiyan", "kutahya"},
		StartedTurn:        7,
		ReassessTurn:       13,
		RallyRegionID:      "bithynia",
		RallyDeadlineTurn:  10,
		Commitment:         62,
		AllowVassalization: true,
		Reason:             "frontier_expansion profili",
	}
	saved := campaignSaveState{
		ScenarioID: "1300_ottoman_rise",
		AIPlans:    map[faction.FactionID]*state.AIPlanState{"ottoman": original},
		AICompletedObjectives: map[faction.FactionID]map[string]bool{
			"ottoman": {"forge_anatolian_power_base": true},
		},
	}
	payload, err := json.Marshal(saved)
	if err != nil {
		t.Fatalf("AI plan compact payload yazılamadı: %v", err)
	}
	decoded, err := decodeCampaignSaveState(payload)
	if err != nil {
		t.Fatalf("AI plan compact payload okunamadı: %v", err)
	}
	restored := &state.GameState{}
	applyCampaignSaveState(restored, decoded)

	got := restored.AIPlans["ottoman"]
	if got == nil {
		t.Fatal("AI planı save/load sonrasında kayboldu")
	}
	if got.ObjectiveID != original.ObjectiveID || got.Kind != original.Kind || got.TargetFactionID != original.TargetFactionID || got.StartedTurn != 7 || got.ReassessTurn != 13 || got.RallyRegionID != "bithynia" || got.RallyDeadlineTurn != 10 || got.Commitment != 62 || got.Reason != original.Reason {
		t.Fatalf("AI plan metadata'sı eksik geri yüklendi: %+v", got)
	}
	if len(got.TargetRegionIDs) != 2 || got.TargetRegionIDs[0] != "germiyan" || got.TargetRegionIDs[1] != "kutahya" {
		t.Fatalf("AI plan bölge öncelikleri korunmadı: %+v", got.TargetRegionIDs)
	}
	if !got.AllowVassalization {
		t.Fatalf("AI plan savaş sonrası düzeni korunmadı: %+v", got)
	}
	if !restored.AICompletedObjectives["ottoman"]["forge_anatolian_power_base"] {
		t.Fatalf("tamamlanan AI objective geçmişi save/load sonrasında kayboldu: %+v", restored.AICompletedObjectives)
	}
	if got == original || &got.TargetRegionIDs[0] == &original.TargetRegionIDs[0] {
		t.Fatal("AI plan save/load kopyası bağımsız olmalı")
	}
}

func TestAutoGrainExportPreferenceRoundTrip(t *testing.T) {
	saved := campaignSaveState{ScenarioID: "test", PlayerFactionID: "player", AutoGrainExport: true}
	encoding, payload, err := encodeCompressedStatePayload(saved)
	if err != nil {
		t.Fatalf("otomatik ihracat tercihi encode edilemedi: %v", err)
	}
	decodedBytes, err := decodeCompressedStatePayload(encoding, payload)
	if err != nil {
		t.Fatalf("otomatik ihracat tercihi decode edilemedi: %v", err)
	}
	decoded, err := decodeCampaignSaveState(decodedBytes)
	if err != nil {
		t.Fatalf("otomatik ihracat tercihi save state'e parse edilemedi: %v", err)
	}
	restored := &state.GameState{}
	applyCampaignSaveState(restored, decoded)
	if !restored.AutoGrainExport {
		t.Fatal("otomatik ihracat tercihi save/load sonrasında açık kalmalıydı")
	}
}

func TestEmbarkedCommanderRoundTrip(t *testing.T) {
	commander := army.NewCommander("cmd_embarked", "Filo Komutanı")
	commander.OwnerID = "player"
	commander.AssignedArmyID = "fleet_1"
	original := map[army.ArmyID]*army.Army{
		"fleet_1": {
			ID:                "fleet_1",
			OwnerID:           "player",
			IsNaval:           true,
			EmbarkedUnits:     []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			EmbarkedCommander: commander,
		},
	}

	restored := restoreArmiesFromSaveState(convertArmiesToSaveState(original))
	got := restored["fleet_1"]
	if got == nil || got.EmbarkedCommander == nil {
		t.Fatal("filodaki taşınan komutan save/load sonrasında kayboldu")
	}
	if got.EmbarkedCommander.Name != commander.Name || got.EmbarkedCommander.AssignedArmyID != "fleet_1" {
		t.Fatalf("taşınan komutan state'i eksik: %+v", got.EmbarkedCommander)
	}
}

func TestLoadFromPathNormalizesLegacyGarrisonArmy(t *testing.T) {
	tmp := t.TempDir()
	oldBaseDir := scenarioBaseDir
	scenarioBaseDir = filepath.Join(tmp, "scenarios")
	defer func() { scenarioBaseDir = oldBaseDir }()

	scenarioID := "test_scenario"
	scenarioPath := filepath.Join(scenarioBaseDir, scenarioID)
	if err := os.MkdirAll(filepath.Join(scenarioPath, "data"), 0755); err != nil {
		t.Fatalf("scenario data dir olusmadi: %v", err)
	}

	writeJSONFile(t, filepath.Join(scenarioPath, "scenario.json"), scenario.Scenario{ID: scenarioID, Year: 1453, Month: 4})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "regions.json"), []*world.Region{
		{ID: "r1", NameTR: "R1", OwnerID: "player", ShapeID: "AAA", TaxRate: 50, Satisfaction: 50},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "settlements.json"), []map[string]any{
		{
			"region_id": "r1",
			"settlements": []map[string]any{
				{"id": "r1_city", "name_tr": "Sehir", "x": 10, "y": 10, "type": "city"},
			},
		},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "factions.json"), []*faction.Faction{
		{ID: "player", NameTR: "Oyuncu", IsPlayable: true},
		{ID: "other", NameTR: "Diger", IsPlayable: true},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "units.json"), []map[string]any{
		{"id": "militia", "name": "Militia", "name_tr": "Milis", "category": "infantry", "attack": 5, "defense": 4, "morale": 10, "hp": 100, "gold_cost": 10, "grain_upkeep": 1, "turns_required": 1},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "buildings.json"), []map[string]any{})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "technologies.json"), []map[string]any{})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "country_shapes.json"), map[string]any{
		"shapes": []map[string]any{
			{"id": "AAA", "name": "AAA", "rings": [][][]int{{{1, 1}, {4, 1}, {4, 4}, {1, 4}}}},
		},
	})

	savePath := filepath.Join(tmp, "legacy-slot.json")
	raw := `{
  "scenario_id": "test_scenario",
  "player_faction_id": "player",
  "regions": {
    "r1": {"id": "r1", "name_tr": "R1", "owner_id": "player", "shape_id": "AAA", "tax_rate": 50, "satisfaction": 50}
  },
  "factions": {
    "player": {"id": "player", "name_tr": "Oyuncu"}
  },
  "armies": {
    "army_garrison_209": {
      "id": "army_garrison_209",
      "owner_id": "player",
      "region_id": "r1",
      "units": [{"type_id": "militia", "current_hp": 100}],
      "move_points": 2,
      "max_move_points": 2
    }
  }
}`
	if err := os.WriteFile(savePath, []byte(raw), 0644); err != nil {
		t.Fatalf("legacy save yazilamadi: %v", err)
	}

	gs, err := loadFromPath(savePath)
	if err != nil {
		t.Fatalf("legacy save yuklenemedi: %v", err)
	}

	garrison := gs.Armies["army_garrison_209"]
	if garrison == nil || !garrison.IsGarrison {
		t.Fatalf("legacy garnizon normalize olmadi: %+v", garrison)
	}
}

func TestSaveToSlotWritesEnvelopeAndLoadSlotReadsIt(t *testing.T) {
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cwd okunamadi: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("temp dir'e gecilemedi: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	oldVersion := GameVersion
	GameVersion = "1.2.3-test"
	t.Cleanup(func() {
		GameVersion = oldVersion
	})

	oldBaseDir := scenarioBaseDir
	scenarioBaseDir = filepath.Join(tmp, "scenarios")
	t.Cleanup(func() { scenarioBaseDir = oldBaseDir })

	scenarioID := "test_scenario"
	scenarioPath := filepath.Join(scenarioBaseDir, scenarioID)
	if err := os.MkdirAll(filepath.Join(scenarioPath, "data"), 0755); err != nil {
		t.Fatalf("scenario data dir olusmadi: %v", err)
	}

	writeJSONFile(t, filepath.Join(scenarioPath, "scenario.json"), scenario.Scenario{
		ID:    scenarioID,
		Year:  1453,
		Month: 4,
		MapConfig: scenario.MapConfig{
			WorldWidth: intPtr(64),
		},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "regions.json"), []*world.Region{
		{ID: "r1", NameTR: "R1", OwnerID: "player", ShapeID: "AAA", TaxRate: 50, Satisfaction: 50},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "factions.json"), []*faction.Faction{
		{ID: "player", NameTR: "Oyuncu", IsPlayable: true},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "units.json"), []map[string]any{
		{"id": "militia", "name": "Militia", "name_tr": "Milis", "category": "infantry", "attack": 5, "defense": 4, "morale": 10, "hp": 100, "gold_cost": 10, "grain_upkeep": 1, "turns_required": 1},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "buildings.json"), []map[string]any{
		{"id": "market", "name": "Market", "name_tr": "Pazar", "gold_cost": 50, "turns_required": 1, "gold_mod": 1.2, "grain_mod": 1.0, "sat_bonus": 0, "def_bonus": 0, "max_per_region": 1},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "technologies.json"), []map[string]any{
		{"id": "tax", "name_tr": "Vergi", "category": "economy", "description_tr": "vergi", "gold_cost": 20, "turns_required": 1, "requires": []string{}, "effects": map[string]any{"gold_per_region": 5}},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "country_shapes.json"), map[string]any{
		"shapes": []map[string]any{
			{"id": "AAA", "name": "AAA", "rings": [][][]int{{{1, 1}, {4, 1}, {4, 4}, {1, 4}}}},
		},
	})

	if err := SaveToSlot(&state.GameState{
		ScenarioID:      scenarioID,
		ScenarioPath:    "",
		Turn:            17,
		Year:            1455,
		Month:           4,
		DevelopmentMode: true,
		AIDiagnosticHistory: []state.AIDiagnosticHistoryEntry{{
			Turn: 17, FactionID: "player", FrontCount: 1,
		}},
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"r1": {
				ID:              "r1",
				NameTR:          "R1",
				OwnerID:         "player",
				ShapeID:         "AAA",
				TaxRate:         50,
				Satisfaction:    50,
				Population:      1234,
				RuralPopulation: 900,
				Settlements: []world.Settlement{
					{ID: "r1_city", NameTR: "Sehir", X: 10, Y: 10, Type: world.SettlementCity, Population: 200},
					{ID: "r1_port", NameTR: "Liman", X: 12, Y: 10, Type: world.SettlementPort, Population: 134},
				},
				Buildings: []string{"market"},
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Gold: 321},
			"other":  {ID: "other", NameTR: "Diger"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_1": {
				ID:            "army_1",
				OwnerID:       "player",
				RegionID:      "r1",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units: []army.Unit{
					{TypeID: "militia", CurrentHP: 100, Experience: 0},
					{TypeID: "militia", CurrentHP: 100, Experience: 0},
					{TypeID: "militia", CurrentHP: 100, Experience: 0},
					{TypeID: "militia", CurrentHP: 82, Experience: 5},
				},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "other"): {
				FactionA: "other",
				FactionB: "player",
				Score:    -40,
				Stance:   faction.StanceWar,
			},
		},
	}, "quicksave"); err != nil {
		t.Fatalf("quicksave yazilamadi: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("saves", "quicksave.json"))
	if err != nil {
		t.Fatalf("quicksave okunamadi: %v", err)
	}
	var saved struct {
		Kind          SaveKind        `json:"kind"`
		GameVersion   string          `json:"game_version"`
		Meta          saveMetadata    `json:"meta"`
		StateEncoding string          `json:"state_encoding"`
		StateZstd     string          `json:"state_zstd"`
		State         json.RawMessage `json:"state"`
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("save envelope okunamadi: %v", err)
	}
	if saved.Kind != SaveKindQuick {
		t.Fatalf("save kind quick olmaliydi: got=%q", saved.Kind)
	}
	if saved.GameVersion != "1.2.3-test" {
		t.Fatalf("game version yazilmadi: got=%q", saved.GameVersion)
	}
	if saved.Meta.Turn != 17 || saved.Meta.Year != 1455 || saved.Meta.FactionName != "Oyuncu" {
		t.Fatalf("save meta beklenen metadata'yi tasimiyor: %+v", saved.Meta)
	}
	if got := ScenarioPathForSlot("quicksave"); got != scenarioPath {
		t.Fatalf("loading ekrani icin scenario path cozulmedi: got=%q want=%q", got, scenarioPath)
	}
	if saved.StateEncoding != saveStateEncodingZstdBase64 || saved.StateZstd == "" {
		t.Fatalf("save zstd formatinda yazilmadi: encoding=%q len=%d", saved.StateEncoding, len(saved.StateZstd))
	}
	if len(saved.State) != 0 {
		t.Fatal("yeni save envelope ham state tasimamalı")
	}
	debugData, err := os.ReadFile(filepath.Join("saves", "quicksave.debug.json"))
	if err != nil {
		t.Fatalf("debug sidecar okunamadi: %v", err)
	}
	var debugSaved struct {
		Kind        SaveKind                `json:"kind"`
		GameVersion string                  `json:"game_version"`
		Meta        saveMetadata            `json:"meta"`
		State       legacyCampaignSaveState `json:"state"`
	}
	if err := json.Unmarshal(debugData, &debugSaved); err != nil {
		t.Fatalf("debug sidecar json okunamadi: %v", err)
	}
	if debugSaved.Kind != SaveKindQuick || debugSaved.GameVersion != "1.2.3-test" {
		t.Fatalf("debug sidecar envelope beklenen degil: %+v", debugSaved)
	}
	if debugSaved.State.ScenarioID != scenarioID || debugSaved.State.DevelopmentMode != true {
		t.Fatalf("debug sidecar state eksik: %+v", debugSaved.State)
	}
	if debugSaved.State.Regions["r1"].Population != 1234 {
		t.Fatalf("debug sidecar region durumu eksik: %+v", debugSaved.State.Regions["r1"])
	}
	if diagnostic := debugSaved.State.AIDiagnostics["player"]; diagnostic == nil || diagnostic.FactionID != "player" {
		t.Fatalf("debug sidecar AI teşhis snapshot'ı eksik: %+v", debugSaved.State.AIDiagnostics)
	}
	if len(debugSaved.State.AIDiagnosticHistory) != 1 || debugSaved.State.AIDiagnosticHistory[0].Turn != 17 {
		t.Fatalf("debug sidecar AI teşhis geçmişi eksik: %+v", debugSaved.State.AIDiagnosticHistory)
	}
	if armyState := debugSaved.State.Armies["army_1"]; armyState == nil || len(armyState.Units) != 4 {
		t.Fatalf("debug sidecar army state eksik: %+v", armyState)
	}
	payload, _, hasEnvelope, err := splitSavePayload(data)
	if err != nil {
		t.Fatalf("state payload ayrıştırılamadı: %v", err)
	}
	if !hasEnvelope {
		t.Fatal("save envelope bekleniyordu")
	}
	compactSaved, err := decodeCampaignSaveState(payload)
	if err != nil {
		t.Fatalf("compact state decode edilemedi: %v", err)
	}
	if compactSaved.ScenarioPath != "" {
		t.Fatalf("varsayılan scenario path payload'a yazılmamalı: %q", compactSaved.ScenarioPath)
	}
	if len(compactSaved.Relations) != 1 {
		t.Fatalf("relation delta bekleniyordu: %+v", compactSaved.Relations)
	}
	relDelta := compactSaved.Relations[faction.RelationKey("player", "other")]
	if relDelta.Score == nil || *relDelta.Score != -40 || relDelta.Stance == nil || decodeStance(*relDelta.Stance) != faction.StanceWar {
		t.Fatalf("relation delta beklenen compact formatta değil: %+v", relDelta)
	}
	rawState := map[string]any{}
	if err := json.Unmarshal(payload, &rawState); err != nil {
		t.Fatalf("decompressed state map okunamadi: %v", err)
	}
	if _, exists := rawState["scenario_path"]; exists {
		t.Fatal("compact state icinde legacy scenario_path yazilmamali")
	}
	if _, exists := rawState["map"]; exists {
		t.Fatal("compact state icinde map yazilmamali")
	}
	if _, exists := rawState["trade_centers"]; exists {
		t.Fatal("compact state icinde trade_centers yazilmamali")
	}
	if _, exists := rawState["region_paint_overrides"]; exists {
		t.Fatal("compact state icinde region_paint_overrides yazilmamali")
	}
	if _, exists := rawState["relations"]; exists {
		t.Fatal("compact state legacy relations key kullanmamalı")
	}
	rawRegions, ok := rawState["rg"].(map[string]any)
	if !ok {
		t.Fatalf("compact regions payload map degil: %#v", rawState["rg"])
	}
	rawRegion, ok := rawRegions["r1"].(map[string]any)
	if !ok {
		t.Fatalf("r1 payload map degil: %#v", rawRegions["r1"])
	}
	if _, exists := rawRegion["o"]; exists {
		t.Fatal("degismeyen owner compact region delta'ya yazilmamali")
	}
	if _, exists := rawRegion["name_tr"]; exists {
		t.Fatal("region static name compact save icinde yazilmamali")
	}
	rawSettlementPatch, ok := rawRegion["sp"].(map[string]any)
	if !ok {
		t.Fatalf("settlement patch bekleniyordu: %#v", rawRegion["sp"])
	}
	if _, hasAdd := rawSettlementPatch["a"]; !hasAdd {
		t.Fatalf("settlement patch add kaydı yok: %#v", rawSettlementPatch)
	}
	rawFactions, ok := rawState["fx"].(map[string]any)
	if !ok {
		t.Fatalf("compact factions payload map degil: %#v", rawState["fx"])
	}
	rawFaction, ok := rawFactions["player"].(map[string]any)
	if !ok {
		t.Fatalf("player faction payload map degil: %#v", rawFactions["player"])
	}
	if _, exists := rawFaction["name_tr"]; exists {
		t.Fatal("faction static name save icinde yazilmamali")
	}
	rawArmies, ok := rawState["ar"].(map[string]any)
	if !ok {
		t.Fatalf("compact armies payload map degil: %#v", rawState["ar"])
	}
	rawArmy, ok := rawArmies["army_1"].(map[string]any)
	if !ok {
		t.Fatalf("army_1 payload map degil: %#v", rawArmies["army_1"])
	}
	rawUnits, ok := rawArmy["u"].([]any)
	if !ok || len(rawUnits) != 2 {
		t.Fatalf("stacked unit payload beklenmiyordu: %#v", rawArmy["u"])
	}

	slots := ListSlots()
	var quick SaveSlot
	found := false
	for _, slot := range slots {
		if slot.Name == "quicksave" {
			quick = slot
			found = true
			break
		}
	}
	if !found {
		t.Fatal("quicksave listede bulunamadi")
	}
	if quick.Kind != SaveKindQuick {
		t.Fatalf("slot kind quick olmaliydi: got=%q", quick.Kind)
	}
	if quick.GameVersion != "1.2.3-test" {
		t.Fatalf("slot game version okunamadi: got=%q", quick.GameVersion)
	}
	if quick.Turn != 17 || quick.Year != 1455 || quick.FactionName != "Oyuncu" {
		t.Fatalf("slot metadata beklenen degerleri tasimiyor: %+v", quick)
	}

	loaded, err := LoadSlot("quicksave")
	if err != nil {
		t.Fatalf("wrapper save yuklenemedi: %v", err)
	}
	if loaded.ScenarioPath != scenarioPath {
		t.Fatalf("scenario path resolve olmadi: got=%q want=%q", loaded.ScenarioPath, scenarioPath)
	}
	if loaded.Turn != 17 || loaded.Year != 1455 {
		t.Fatalf("wrapper save state geri yuklenemedi: turn=%d year=%d", loaded.Turn, loaded.Year)
	}
	if loaded.Regions["r1"] == nil || loaded.Regions["r1"].NameTR != "R1" || loaded.Regions["r1"].ShapeID != "AAA" {
		t.Fatalf("senaryo static region verisi geri kurulmedi: %+v", loaded.Regions["r1"])
	}
	if loaded.Regions["r1"].Population != 1234 || len(loaded.Regions["r1"].Buildings) != 1 || loaded.Regions["r1"].Buildings[0] != "market" {
		t.Fatalf("region mutable save state geri yuklenemedi: %+v", loaded.Regions["r1"])
	}
	if loaded.Regions["r1"].RuralPopulation != 900 || loaded.Regions["r1"].SettlementPopulation() != 334 {
		t.Fatalf("nüfus bileşenleri save/load sonrası korunmalıydı: rural=%d settlement=%d", loaded.Regions["r1"].RuralPopulation, loaded.Regions["r1"].SettlementPopulation())
	}
	if len(loaded.Regions["r1"].Settlements) != 2 || loaded.Regions["r1"].Settlements[1].ID != "r1_port" {
		t.Fatalf("settlement patch geri yuklenemedi: %+v", loaded.Regions["r1"].Settlements)
	}
	if loaded.Factions["player"] == nil || loaded.Factions["player"].NameTR != "Oyuncu" {
		t.Fatalf("senaryo static faction verisi geri kurulmedi: %+v", loaded.Factions["player"])
	}
	if loaded.Factions["player"].Gold != 321 {
		t.Fatalf("faction mutable save state geri yuklenemedi: %+v", loaded.Factions["player"])
	}
	if rel := loaded.Relations[faction.RelationKey("player", "other")]; rel == nil || rel.Score != -40 || rel.Stance != faction.StanceWar {
		t.Fatalf("relation delta geri yuklenemedi: %+v", rel)
	}
	if a := loaded.Armies["army_1"]; a == nil || len(a.Units) != 4 {
		t.Fatalf("stacked army geri yuklenemedi: %+v", a)
	}
}

func TestLatestContinueSlotPrefersNewestAutosaveOrQuicksave(t *testing.T) {
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cwd okunamadi: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("temp dir'e gecilemedi: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	if err := os.MkdirAll("saves", 0755); err != nil {
		t.Fatalf("saves dizini olusmadi: %v", err)
	}

	autosavePath := filepath.Join("saves", "autosave.json")
	quicksavePath := filepath.Join("saves", "quicksave.json")
	if err := os.WriteFile(autosavePath, []byte("{}"), 0644); err != nil {
		t.Fatalf("autosave yazilamadi: %v", err)
	}
	if err := os.WriteFile(quicksavePath, []byte("{}"), 0644); err != nil {
		t.Fatalf("quicksave yazilamadi: %v", err)
	}

	oldTime := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 7, 12, 10, 5, 0, 0, time.UTC)

	if err := os.Chtimes(autosavePath, oldTime, oldTime); err != nil {
		t.Fatalf("autosave modtime ayarlanamadi: %v", err)
	}
	if err := os.Chtimes(quicksavePath, newTime, newTime); err != nil {
		t.Fatalf("quicksave modtime ayarlanamadi: %v", err)
	}

	slot, ok := LatestContinueSlot()
	if !ok {
		t.Fatal("beklenen continue save bulunamadi")
	}
	if slot != "quicksave" {
		t.Fatalf("en yeni slot quicksave olmaliydi: got=%q want=%q", slot, "quicksave")
	}
	if !ContinueSaveExists() {
		t.Fatal("continue save var olmaliydi")
	}

	if err := os.Chtimes(autosavePath, newTime, newTime); err != nil {
		t.Fatalf("autosave modtime yeniden ayarlanamadi: %v", err)
	}
	if err := os.Chtimes(quicksavePath, oldTime, oldTime); err != nil {
		t.Fatalf("quicksave modtime yeniden ayarlanamadi: %v", err)
	}

	slot, ok = LatestContinueSlot()
	if !ok {
		t.Fatal("beklenen continue save ikinci senaryoda bulunamadi")
	}
	if slot != "autosave" {
		t.Fatalf("en yeni slot autosave olmaliydi: got=%q want=%q", slot, "autosave")
	}
}

func TestSaveToSlotNonDevRemovesDebugSidecar(t *testing.T) {
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cwd okunamadi: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("temp dir'e gecilemedi: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	oldBaseDir := scenarioBaseDir
	scenarioBaseDir = filepath.Join(tmp, "scenarios")
	t.Cleanup(func() { scenarioBaseDir = oldBaseDir })

	scenarioID := "devless_scenario"
	scenarioPath := filepath.Join(scenarioBaseDir, scenarioID)
	if err := os.MkdirAll(filepath.Join(scenarioPath, "data"), 0755); err != nil {
		t.Fatalf("scenario data dir olusmadi: %v", err)
	}
	writeJSONFile(t, filepath.Join(scenarioPath, "scenario.json"), scenario.Scenario{ID: scenarioID, Year: 1300, Month: 3})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "regions.json"), []*world.Region{})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "factions.json"), []*faction.Faction{
		{ID: "player", NameTR: "Oyuncu", IsPlayable: true},
	})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "units.json"), []map[string]any{})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "buildings.json"), []map[string]any{})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "technologies.json"), []map[string]any{})
	writeJSONFile(t, filepath.Join(scenarioPath, "data", "country_shapes.json"), map[string]any{"shapes": []map[string]any{}})

	if err := os.MkdirAll("saves", 0755); err != nil {
		t.Fatalf("saves dizini olusmadi: %v", err)
	}
	debugPath := filepath.Join("saves", "quicksave.debug.json")
	if err := os.WriteFile(debugPath, []byte(`{"stale":true}`), 0644); err != nil {
		t.Fatalf("stale debug sidecar yazilamadi: %v", err)
	}
	if err := SaveToSlot(&state.GameState{
		ScenarioID:      scenarioID,
		Turn:            3,
		Year:            1300,
		Month:           3,
		PlayerFactionID: "player",
		Victory:         state.VictoryCondition{},
		Regions:         map[world.RegionID]*world.Region{},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
		},
		Armies: map[army.ArmyID]*army.Army{},
	}, "quicksave"); err != nil {
		t.Fatalf("non-dev quicksave yazilamadi: %v", err)
	}

	if _, err := os.Stat(debugPath); !os.IsNotExist(err) {
		t.Fatalf("non-dev save debug sidecar birakmamaliydi: err=%v", err)
	}
}

func writeJSONFile(t *testing.T, path string, payload any) {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal hatasi: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("json dosyasi yazilamadi: %v", err)
	}
}

func intPtr(v int) *int {
	return &v
}
