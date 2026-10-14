package scenario

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRequiresValidPeriod(t *testing.T) {
	tests := []struct {
		name   string
		period string
		valid  bool
	}{
		{name: "missing", valid: false},
		{name: "unknown", period: "renaissance", valid: false},
		{name: "valid", period: "medieval", valid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			dataDir := filepath.Join(dir, "data")
			if err := os.Mkdir(dataDir, 0o700); err != nil {
				t.Fatalf("data klasörü oluşturulamadı: %v", err)
			}
			content := []byte(`{"id":"test","name":"Test","period":"` + tt.period + `"}`)
			if err := os.WriteFile(filepath.Join(dataDir, "scenario.json"), content, 0o600); err != nil {
				t.Fatalf("scenario.json yazılamadı: %v", err)
			}

			_, err := Load(dir)
			if (err == nil) != tt.valid {
				t.Fatalf("Load() hata = %v, valid = %v", err, tt.valid)
			}
		})
	}
}

func TestLoadReadsPrivilegedBuildingMaxLevel(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	if err := os.Mkdir(dataDir, 0o700); err != nil {
		t.Fatalf("data klasörü oluşturulamadı: %v", err)
	}
	content := []byte(`{"id":"test","name":"Test","period":"medieval","privileged_building_max_level":3,"minor_privilege_protection_turns":17,"aggressive_expansion_last_turns":12}`)
	if err := os.WriteFile(filepath.Join(dataDir, "scenario.json"), content, 0o600); err != nil {
		t.Fatalf("scenario.json yazılamadı: %v", err)
	}

	definition, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() hatası: %v", err)
	}
	if definition.PrivilegedBuildingMaxLevel != 3 {
		t.Fatalf("imtiyazlı bina seviye tavanı = %d, 3 bekleniyordu", definition.PrivilegedBuildingMaxLevel)
	}
	if definition.MinorPrivilegeProtectionTurns != 17 {
		t.Fatalf("minor imtiyaz koruma turu = %d, 17 bekleniyordu", definition.MinorPrivilegeProtectionTurns)
	}
	if definition.AggressiveExpansionLastTurns != 12 {
		t.Fatalf("aşırı genişleme süresi = %d, 12 bekleniyordu", definition.AggressiveExpansionLastTurns)
	}
}

func TestLoadReadsPeacePeriodRequiredEvent(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	if err := os.Mkdir(dataDir, 0o700); err != nil {
		t.Fatalf("data klasörü oluşturulamadı: %v", err)
	}
	content := []byte(`{"id":"test","name":"Test","period":"medieval","diplomacy":{"peace_periods":[{"war_declaration_requires_event":"death_of_robert_298"}]}}`)
	if err := os.WriteFile(filepath.Join(dataDir, "scenario.json"), content, 0o600); err != nil {
		t.Fatalf("scenario.json yazılamadı: %v", err)
	}

	definition, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() hatası: %v", err)
	}
	if len(definition.Diplomacy.PeacePeriods) != 1 ||
		definition.Diplomacy.PeacePeriods[0].WarDeclarationRequiresEvent != "death_of_robert_298" {
		t.Fatalf("barış dönemi kalıcı event ID'sini yüklemedi: %+v", definition.Diplomacy.PeacePeriods)
	}
}

func TestLoadCharacterCouncilConfigSupportsOptionalAndScenarioSpecificData(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	if err := os.Mkdir(dataDir, 0o700); err != nil {
		t.Fatalf("data klasörü oluşturulamadı: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "scenario.json"), []byte(`{"period":"medieval"}`), 0o600); err != nil {
		t.Fatalf("scenario.json yazılamadı: %v", err)
	}
	legacy, err := Load(dir)
	if err != nil {
		t.Fatalf("eski senaryo varsayılan görevlerle yüklenemedi: %v", err)
	}
	if len(legacy.CharacterCouncil.Roles) == 0 {
		t.Fatal("görev dosyası olmayan senaryo varsayılan görev listesini almadı")
	}

	config := []byte(`{
		"xp_per_turn": 7,
		"xp_per_effect": 13,
		"roles": [{"id":"steward","name_tr":"Kâhya","set_commander_army":true,"effects":{"income_mod":1.04}}]
	}`)
	if err := os.WriteFile(filepath.Join(dataDir, "character_council.json"), config, 0o600); err != nil {
		t.Fatalf("character_council.json yazılamadı: %v", err)
	}
	traits := []byte(`{
		"traits": {"scholar":{"name_tr":"Bilgin","council_effects":{"steward":{"income_mod":1.02}}}}
	}`)
	if err := os.WriteFile(filepath.Join(dataDir, "character_traits.json"), traits, 0o600); err != nil {
		t.Fatalf("character_traits.json yazılamadı: %v", err)
	}
	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("senaryo görev tanımı yüklenemedi: %v", err)
	}
	if loaded.CharacterCouncil.XPPerTurn != 7 || loaded.CharacterCouncil.XPPerEffect != 13 ||
		len(loaded.CharacterCouncil.Roles) != 1 || loaded.CharacterCouncil.Roles[0].ID != "steward" ||
		!loaded.CharacterCouncil.Roles[0].SetCommanderArmy ||
		loaded.CharacterCouncil.Traits["scholar"].CouncilEffects["steward"].IncomeMod != 1.02 {
		t.Fatalf("senaryo görev/nitelik tanımı beklenmedik: %+v", loaded.CharacterCouncil)
	}

	legacyCombined := []byte(`{
		"xp_per_turn": 7,
		"xp_per_effect": 13,
		"roles": [{"id":"steward","name_tr":"Kâhya","effects":{"income_mod":1.04}}],
		"traits": {"scholar":{"name_tr":"Bilgin","council_effects":{"steward":{"income_mod":1.02}}}}
	}`)
	if err := os.WriteFile(filepath.Join(dataDir, "character_council.json"), legacyCombined, 0o600); err != nil {
		t.Fatalf("eski birleşik görev tanımı yazılamadı: %v", err)
	}
	if err := os.Remove(filepath.Join(dataDir, "character_traits.json")); err != nil {
		t.Fatalf("ayrı nitelik tanımı kaldırılamadı: %v", err)
	}
	legacyCombinedConfig, err := Load(dir)
	if err != nil {
		t.Fatalf("eski birleşik görev/nitelik tanımı yüklenemedi: %v", err)
	}
	if got := legacyCombinedConfig.CharacterCouncil.Traits["scholar"].CouncilEffects["steward"].IncomeMod; got != 1.02 {
		t.Fatalf("eski birleşik nitelik bonusu = %.2f, want 1.02", got)
	}
}

func TestAllScenarioCouncilConfigurationsLoad(t *testing.T) {
	root := filepath.Join("..", "..", "assets", "scenarios")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("senaryo dizini okunamadı: %v", err)
	}
	loadedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		scenarioPath := filepath.Join(root, entry.Name())
		if _, err := os.Stat(filepath.Join(scenarioPath, "data", "scenario.json")); err != nil {
			continue
		}
		definition, err := Load(scenarioPath)
		if err != nil {
			t.Errorf("%s yüklenemedi: %v", entry.Name(), err)
			continue
		}
		councilPath := filepath.Join(scenarioPath, "data", "character_council.json")
		if _, err := os.Stat(councilPath); err == nil {
			charactersPath := filepath.Join(scenarioPath, "data", "characters.json")
			if data, err := os.ReadFile(charactersPath); err == nil {
				var characters []struct {
					Traits []string `json:"traits"`
				}
				if err := json.Unmarshal(data, &characters); err != nil {
					t.Errorf("%s karakter dosyası parse edilemedi: %v", entry.Name(), err)
				} else {
					for _, character := range characters {
						for _, traitID := range character.Traits {
							if _, ok := definition.CharacterCouncil.Traits[traitID]; !ok {
								t.Errorf("%s konsey tanımında karakter niteliği %q eksik", entry.Name(), traitID)
							}
						}
					}
				}
			} else if !os.IsNotExist(err) {
				t.Errorf("%s karakter dosyası okunamadı: %v", entry.Name(), err)
			}
		} else if !os.IsNotExist(err) {
			t.Errorf("%s konsey tanımı incelenemedi: %v", entry.Name(), err)
		}
		loadedCount++
	}
	if loadedCount == 0 {
		t.Fatal("doğrulanacak senaryo bulunamadı")
	}
}

func TestDiplomacyWithDefaultsPreservesPeacePeriods(t *testing.T) {
	periods := []PeacePeriod{{
		MinTurns:                           10,
		WarDeclarationRequiresEvent:        "death_of_robert_298",
		WarDeclarationRequiresEventFlag:    "war_started",
		WarDeclarationUnblockedByEventFlag: "war_resolved",
		BlockedFactions:                    []string{"a", "b"},
	}}

	got := (DiplomacyConfig{PeacePeriods: periods}).WithDefaults()
	if len(got.PeacePeriods) != 1 ||
		got.PeacePeriods[0].WarDeclarationRequiresEvent != "death_of_robert_298" ||
		got.PeacePeriods[0].WarDeclarationRequiresEventFlag != "war_started" ||
		got.PeacePeriods[0].WarDeclarationUnblockedByEventFlag != "war_resolved" {
		t.Fatalf("barış dönemleri varsayılanlarla kayboldu: %+v", got.PeacePeriods)
	}
}

func TestLoadPoliticalTransformationsValidatesUnionMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "political_transformations.json")
	content := []byte(`[
		{"id":"test_union","type":"dynastic_union","members":["a","b"],"result_faction":"result","transfer":{"regions":true,"armies":true,"navies":true,"resources":true,"relations":"merge"}}
	]`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("dönüşüm fixture'ı yazılamadı: %v", err)
	}

	transformations, err := LoadPoliticalTransformations(path)
	if err != nil {
		t.Fatalf("dönüşüm metadata'sı yüklenemedi: %v", err)
	}
	if len(transformations) != 1 || transformations[0].ResultFaction != "result" {
		t.Fatalf("dönüşüm metadata'sı beklenmedik: %+v", transformations)
	}
}

func TestLoadPoliticalTransformationsAllowsMissingOptionalFile(t *testing.T) {
	transformations, err := LoadPoliticalTransformations(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("eksik opsiyonel dosya hata vermemeli: %v", err)
	}
	if transformations != nil {
		t.Fatalf("eksik dosyada boş nil liste bekleniyordu: %+v", transformations)
	}
}
