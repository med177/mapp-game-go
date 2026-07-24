package army

import (
	"testing"

	"mapp-game-go/internal/world"
)

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

func TestNavalCanReplenishInOwnPort(t *testing.T) {
	regions := map[world.RegionID]*world.Region{
		"sea":   {ID: "sea", IsSea: true},
		"home":  {ID: "home", OwnerID: "p1", Buildings: []string{"port"}},
		"enemy": {ID: "enemy", OwnerID: "p2", Buildings: []string{"port"}},
		"coast": {ID: "coast", OwnerID: "p1"},
	}

	tests := []struct {
		name  string
		fleet *Army
		want  bool
	}{
		{name: "kendi limanı", fleet: &Army{OwnerID: "p1", IsNaval: true, DockedRegionID: "home"}, want: true},
		{name: "açık deniz", fleet: &Army{OwnerID: "p1", IsNaval: true, RegionID: "sea"}, want: false},
		{name: "düşman limanı", fleet: &Army{OwnerID: "p1", IsNaval: true, DockedRegionID: "enemy"}, want: false},
		{name: "limansız kıyı", fleet: &Army{OwnerID: "p1", IsNaval: true, DockedRegionID: "coast"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fleet.CanReplenishIn(regions); got != tt.want {
				t.Fatalf("donanma ikmal koşulu yanlış: got=%v want=%v", got, tt.want)
			}
		})
	}
}
