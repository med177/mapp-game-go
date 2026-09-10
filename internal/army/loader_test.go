package army

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadUnitTypesWithOrderPreservesJSONOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "units.json")
	data := []byte(`[
  {"id":"zeta","turns_required":2},
  {"id":"alpha","turns_required":3}
]`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("test units.json yazılamadı: %v", err)
	}

	types, order, err := LoadUnitTypesWithOrder(path)
	if err != nil {
		t.Fatalf("birim tipleri yüklenemedi: %v", err)
	}
	if _, ok := types["zeta"]; !ok {
		t.Fatal("zeta birimi yüklenmeliydi")
	}
	if got, want := order, []string{"zeta", "alpha"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("JSON birim sırası korunmalıydı: got=%v want=%v", got, want)
	}
}
