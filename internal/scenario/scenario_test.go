package scenario

import (
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

func TestDiplomacyWithDefaultsPreservesPeacePeriods(t *testing.T) {
	periods := []PeacePeriod{{
		MinTurns:                        10,
		WarDeclarationRequiresEventFlag: "war_started",
		BlockedFactions:                 []string{"a", "b"},
	}}

	got := (DiplomacyConfig{PeacePeriods: periods}).WithDefaults()
	if len(got.PeacePeriods) != 1 || got.PeacePeriods[0].WarDeclarationRequiresEventFlag != "war_started" {
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
