package world

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLandPassagesFiltersInvalidAndReverseDuplicates(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "land_passages.json")
	data := []byte(`[
  {"from":"a","to":"b","type":"strait","move_cost":1,"defense_bonus":15},
  {"from":"b","to":"a","type":"strait","move_cost":2,"defense_bonus":20},
  {"from":"a","to":"sea","type":"strait","move_cost":1},
  {"from":"unknown","to":"b","type":"strait","move_cost":1},
  {"from":"a","to":"a","type":"strait","move_cost":1}
]`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("test JSON yazılamadı: %v", err)
	}

	regions := map[RegionID]*Region{
		"a":   {ID: "a"},
		"b":   {ID: "b"},
		"sea": {ID: "sea", IsSea: true},
	}
	passages, err := LoadLandPassages(path, regions)
	if err != nil {
		t.Fatalf("geçişler yüklenemedi: %v", err)
	}
	if len(passages) != 1 {
		t.Fatalf("beklenen 1 geçiş, got=%d", len(passages))
	}
	if !HasLandPassage(passages, "b", "a") {
		t.Fatal("geçiş yön bağımsız bulunamadı")
	}
}

func TestLoadLandPassagesKeepsExplicitEndpoints(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "land_passages.json")
	data := []byte(`[
  {"from":"a","to":"b","type":"strait","move_cost":1,"defense_bonus":15,"start":[12,34],"end":[56,78]}
]`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("test JSON yazılamadı: %v", err)
	}

	passages, err := LoadLandPassages(path, map[RegionID]*Region{
		"a": {ID: "a"},
		"b": {ID: "b"},
	})
	if err != nil {
		t.Fatalf("geçişler yüklenemedi: %v", err)
	}
	if len(passages) != 1 || !passages[0].HasCustomEndpoints() {
		t.Fatalf("açık uç noktalar korunmadı: %+v", passages)
	}
	if *passages[0].Start != [2]int{12, 34} || *passages[0].End != [2]int{56, 78} {
		t.Fatalf("uç noktalar değişti: start=%v end=%v", *passages[0].Start, *passages[0].End)
	}
}

func TestLoad1300LandPassages(t *testing.T) {
	path := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data", "land_passages.json")
	regions, err := LoadRegions(filepath.Join(filepath.Dir(path), "regions.json"))
	if err != nil {
		t.Fatalf("1300 bölgeleri yüklenemedi: %v", err)
	}
	passages, err := LoadLandPassages(path, regions)
	if err != nil {
		t.Fatalf("1300 geçişleri yüklenemedi: %v", err)
	}
	if len(passages) < 6 {
		t.Fatalf("en az 6 tarihsel geçiş bekleniyordu, got=%d", len(passages))
	}
	for _, pair := range [][2]RegionID{
		{"sicily", "naples"},
		{"constantinople", "bithynia"},
		{"gallipoli", "aydinoglu"},
		{"denmark", "sweden"},
		{"ulster", "scotland"},
		{"morocco", "granada"},
	} {
		if !HasLandPassage(passages, pair[0], pair[1]) {
			t.Fatalf("geçiş bulunamadı: %s -> %s", pair[0], pair[1])
		}
	}
	for _, passage := range passages {
		if !passage.HasCustomEndpoints() {
			t.Fatalf("%s -> %s için açık çizgi uçları eksik", passage.From, passage.To)
		}
	}
}
