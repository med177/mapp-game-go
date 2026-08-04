package army

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCommanderTemplatesIndexesTraitsAndPortrait(t *testing.T) {
	path := filepath.Join(t.TempDir(), "commanders.json")
	data := []byte(`[
  {
    "id": "commander_test",
    "owner_id": "ottoman",
    "name": "Test Komutan",
    "portrait_asset": "commanders/test.png",
    "level": 1,
    "traits": ["defender"],
    "start_year": 1308,
    "end_year": 1334
  }
]`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("test commanders dosyası yazılamadı: %v", err)
	}

	got, err := LoadCommanderTemplates(path)
	if err != nil {
		t.Fatalf("komutan şablonları yüklenemedi: %v", err)
	}
	templates := got["ottoman"]
	if len(templates) != 1 || templates[0].Name != "Test Komutan" {
		t.Fatalf("komutan şablonu indekslenmedi: %+v", got)
	}
	if templates[0].PortraitAsset != "commanders/test.png" || !templates[0].HasTrait(CommanderTraitDefender) {
		t.Fatalf("portre veya trait kayboldu: %+v", templates[0])
	}
	if templates[0].StartYear != 1308 || templates[0].EndYear != 1334 || !templates[0].ActiveInYear(1308) || templates[0].ActiveInYear(1334) {
		t.Fatalf("komutan tarih aralığı yanlış yüklendi: %+v", templates[0])
	}
}
