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
			content := []byte(`{"id":"test","name":"Test","period":"` + tt.period + `"}`)
			if err := os.WriteFile(filepath.Join(dir, "scenario.json"), content, 0o600); err != nil {
				t.Fatalf("scenario.json yazılamadı: %v", err)
			}

			_, err := Load(dir)
			if (err == nil) != tt.valid {
				t.Fatalf("Load() hata = %v, valid = %v", err, tt.valid)
			}
		})
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
