package faction

import (
	"os"
	"path/filepath"
	"testing"

	"mapp-game-go/internal/religion"
)

func TestLoadRelationsUsesPeaceAndReligionDefaultsForMissingRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "relations.json")
	if err := os.WriteFile(path, []byte("[]\n"), 0o644); err != nil {
		t.Fatalf("boş relations dosyası yazılamadı: %v", err)
	}

	factions := map[FactionID]*Faction{
		"sunni_a":  {ID: "sunni_a", Religion: religion.Sunni},
		"sunni_b":  {ID: "sunni_b", Religion: religion.Sunni},
		"catholic": {ID: "catholic", Religion: religion.Catholic},
		"shia":     {ID: "shia", Religion: religion.Shia},
	}
	relations, order, err := LoadRelationsWithOrder(path, factions)
	if err != nil {
		t.Fatalf("relations yüklenemedi: %v", err)
	}
	if len(order) != 0 {
		t.Fatalf("kayıtsız varsayılan ilişkiler kaynak sırasına eklenmemeli: %v", order)
	}

	cases := []struct {
		name  string
		a     FactionID
		b     FactionID
		score int
	}{
		{name: "aynı din", a: "sunni_a", b: "sunni_b", score: 25},
		{name: "farklı din", a: "sunni_a", b: "catholic", score: -30},
		{name: "sunni şii", a: "sunni_a", b: "shia", score: -30},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rel := relations[RelationKey(tc.a, tc.b)]
			if rel == nil || rel.Score != tc.score || rel.Stance != StancePeace {
				t.Fatalf("varsayılan ilişki hatalı: %+v", rel)
			}
		})
	}
}
