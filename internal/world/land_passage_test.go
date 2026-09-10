package world

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLandPassagesPreservesJSONOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "land_passages.json")
	data := []byte(`[
  {"from":"zeta","to":"eta","type":"strait","move_cost":1},
  {"from":"alpha","to":"beta","type":"strait","move_cost":1},
  {"from":"gamma","to":"delta","type":"strait","move_cost":1}
]`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("test land_passages dosyası yazılamadı: %v", err)
	}

	regions := map[RegionID]*Region{
		"zeta":  {ID: "zeta"},
		"eta":   {ID: "eta"},
		"alpha": {ID: "alpha"},
		"beta":  {ID: "beta"},
		"gamma": {ID: "gamma"},
		"delta": {ID: "delta"},
	}
	passages, err := LoadLandPassages(path, regions)
	if err != nil {
		t.Fatalf("land_passages yüklenemedi: %v", err)
	}

	if len(passages) != 3 {
		t.Fatalf("beklenen geçiş sayısı 3, alınan %d", len(passages))
	}
	want := [][2]RegionID{{"zeta", "eta"}, {"alpha", "beta"}, {"gamma", "delta"}}
	for i, passage := range passages {
		if passage.From != want[i][0] || passage.To != want[i][1] {
			t.Fatalf("%d. geçiş sırası değişti: %s -> %s", i, passage.From, passage.To)
		}
	}
}
