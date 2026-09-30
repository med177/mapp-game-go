package world

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRegionSettlementsPreservesRegionPopulation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settlements.json")
	data := []byte(`[
  {
    "region_id": "bursa",
    "settlements": [
      {"id": "bursa_center", "type": "city", "population": 61}
    ]
  }
]
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	regions := map[RegionID]*Region{
		"bursa": {ID: "bursa", Population: 873, RuralPopulation: 697},
	}
	if err := LoadRegionSettlements(path, regions); err != nil {
		t.Fatalf("LoadRegionSettlements() error = %v", err)
	}
	if got := regions["bursa"].Population; got != 873 {
		t.Fatalf("settlement yükleme bölge nüfusunu değiştirdi: got %d want 873", got)
	}

	encoded, err := json.Marshal(regions["bursa"].Settlements[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "population") {
		t.Fatalf("settlement JSON'unda nüfus alanı kaldı: %s", encoded)
	}
}
