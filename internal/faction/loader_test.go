package faction

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRelationsWithOrderReadsDirectionalScores(t *testing.T) {
	path := filepath.Join(t.TempDir(), "relations.json")
	data := []byte(`[
  {"faction_a":"a","faction_b":"b","score_a_to_b":-35,"score_b_to_a":20,"stance":"peace"}
]`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	relations, order, err := LoadRelationsWithOrder(path, map[FactionID]*Faction{
		"a": {ID: "a"},
		"b": {ID: "b"},
	})
	if err != nil {
		t.Fatal(err)
	}
	relation := relations[RelationKey("a", "b")]
	if relation == nil {
		t.Fatalf("ilişki yüklenmedi: %#v", relations)
	}
	if relation.ScoreAToB != -35 || relation.ScoreBToA != 20 {
		t.Fatalf("yönlü ilişki puanları korunmadı: a->b=%d b->a=%d", relation.ScoreAToB, relation.ScoreBToA)
	}
	if len(order) != 1 || order[0] != RelationKey("a", "b") {
		t.Fatalf("ilişki sırası korunmadı: %#v", order)
	}
}

func TestLoadRelationsWithOrderKeepsLegacySharedScore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "relations.json")
	data := []byte(`[
  {"faction_a":"a","faction_b":"b","score":15,"stance":"trade"}
]`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	relations, _, err := LoadRelationsWithOrder(path, map[FactionID]*Faction{
		"a": {ID: "a"},
		"b": {ID: "b"},
	})
	if err != nil {
		t.Fatal(err)
	}
	relation := relations[RelationKey("a", "b")]
	if relation == nil || relation.ScoreAToB != 15 || relation.ScoreBToA != 15 {
		t.Fatalf("eski ortak score formatı korunmadı: %#v", relation)
	}
}
