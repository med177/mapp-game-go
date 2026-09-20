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

func TestUnitTypeBuildingRequirementsUseAllListedBuildings(t *testing.T) {
	unit := &UnitType{
		RequiredBuildings: []BuildingRequirement{
			{ID: "barracks", Level: 2},
			{ID: "forge", Level: 2},
		},
	}

	if unit.HasBuildingRequirements(map[string]int{"barracks": 2, "forge": 1}) {
		t.Fatal("demirhane seviyesi eksikken birlik üretilebilir görünmemeliydi")
	}
	if !unit.HasBuildingRequirements(map[string]int{"barracks": 2, "forge": 2}) {
		t.Fatal("iki bina şartı da sağlandığında birlik üretilebilir görünmeliydi")
	}
}

func TestLoadUnitTypesKeepsPrimaryBuildingFromNewRequirements(t *testing.T) {
	path := filepath.Join(t.TempDir(), "units.json")
	data := []byte(`[
  {"id":"elite","required_buildings":[{"id":"barracks","level":2},{"id":"forge","level":1}]}
]`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("test units.json yazılamadı: %v", err)
	}

	types, err := LoadUnitTypes(path)
	if err != nil {
		t.Fatalf("yeni bina gereksinimi yüklenemedi: %v", err)
	}
	unit := types["elite"]
	if unit.PrimaryBuildingID() != "barracks" || unit.PrimaryBuildingLevel() != 2 {
		t.Fatalf("ana bina yeni gereksinimlerden okunmalıydı: %+v", unit.BuildingRequirements())
	}
	if !unit.HasBuildingRequirements(map[string]int{"barracks": 2, "forge": 1}) {
		t.Fatal("yeni listedeki tüm bina şartları korunmalıydı")
	}
}
