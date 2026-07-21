package army

import "testing"

func TestUnitTypeRequiresAllTechnologies(t *testing.T) {
	unit := &UnitType{ID: "cannon", RequiredTech: []string{"gunpowder", "cast_bronze_cannon"}}

	if unit.HasAllRequiredTechs(map[string]bool{"gunpowder": true}) {
		t.Fatal("eksik zincir teknolojisi varken top üretilebilir görünmemeliydi")
	}
	if !unit.HasAllRequiredTechs(map[string]bool{"gunpowder": true, "cast_bronze_cannon": true}) {
		t.Fatal("tüm zincir teknolojileri tamamlanınca top üretilebilir görünmeliydi")
	}

	missing := unit.MissingRequiredTechs(map[string]bool{"gunpowder": true})
	if len(missing) != 1 || missing[0] != "cast_bronze_cannon" {
		t.Fatalf("beklenen eksik teknoloji cast_bronze_cannon, got=%v", missing)
	}
}

func TestUnitTypeRequiresTech(t *testing.T) {
	unit := &UnitType{RequiredTech: []string{"navigation", "naval_doctrine"}}
	if !unit.RequiresTech("navigation") || unit.RequiresTech("gunpowder") {
		t.Fatal("birim teknoloji üyeliği yanlış değerlendirildi")
	}
}
