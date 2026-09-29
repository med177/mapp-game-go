package render

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestEditRegionBuildingButtonHasSeparateSharedRect(t *testing.T) {
	buildings := editInspectorButtonRect(editButtonRegionBuildings)
	privilege := editInspectorButtonRect(editButtonRegionPrivilege)
	if buildings == privilege {
		t.Fatalf("binalar ve imtiyaz düğmeleri aynı geometriyi kullanıyor: buildings=%v privilege=%v", buildings, privilege)
	}
	if buildings[1] != privilege[1] || buildings[2] != privilege[2] {
		t.Fatalf("binalar ve imtiyaz düğmeleri aynı satırda olmalı: buildings=%v privilege=%v", buildings, privilege)
	}

	buildingsCenterX := buildings[0] + buildings[2]/2
	buildingsCenterY := buildings[1] + buildings[3]/2
	if got := editRegionInspectorButtonAt(buildingsCenterX, buildingsCenterY); got != editButtonRegionBuildings {
		t.Fatalf("binalar düğmesi hit-test sonucu = %v, want %v", got, editButtonRegionBuildings)
	}
	privilegeCenterX := privilege[0] + privilege[2]/2
	privilegeCenterY := privilege[1] + privilege[3]/2
	if got := editRegionInspectorButtonAt(privilegeCenterX, privilegeCenterY); got != editButtonRegionPrivilege {
		t.Fatalf("imtiyaz düğmesi hit-test sonucu = %v, want %v", got, editButtonRegionPrivilege)
	}
}

func TestEditBuildingLevelCanBeAddedAndRemovedDirectly(t *testing.T) {
	region := &world.Region{ID: "test", Buildings: []string{"granary"}}
	r := &Renderer{gs: &state.GameState{
		Regions: map[world.RegionID]*world.Region{"test": region},
		BuildingTypes: map[string]*city.Building{
			"granary": {ID: "granary", MaxPerRegion: 3},
		},
		BuildingOrder: []string{"granary"},
	}, editSelectedRegion: "test"}

	if !r.changeEditBuildingLevel("granary", 1) || region.BuildingLevel("granary") != 2 {
		t.Fatalf("bina seviyesi eklenmedi: %#v", region.Buildings)
	}
	if !r.changeEditBuildingLevel("granary", 1) || region.BuildingLevel("granary") != 3 {
		t.Fatalf("bina maksimum seviyeye çıkarılmadı: %#v", region.Buildings)
	}
	if r.changeEditBuildingLevel("granary", 1) {
		t.Fatal("bina tanımındaki maksimum seviye aşıldı")
	}
	if !r.changeEditBuildingLevel("granary", -1) || region.BuildingLevel("granary") != 2 {
		t.Fatalf("bina seviyesi çıkarılmadı: %#v", region.Buildings)
	}
	if !r.editDirty {
		t.Fatal("bina değişikliği Edit Mode dirty durumunu işaretlemedi")
	}
}

func TestEditBuildingLevelHonorsMinorRegionBuildingMetadata(t *testing.T) {
	region := &world.Region{ID: "minor", IsMinorRegion: true}
	r := &Renderer{gs: &state.GameState{
		Regions: map[world.RegionID]*world.Region{"minor": region},
		BuildingTypes: map[string]*city.Building{
			"market":  {ID: "market", MaxPerRegion: 3},
			"granary": {ID: "granary", MaxPerRegion: 3, MinorRegions: true},
		},
		BuildingOrder: []string{"market", "granary"},
	}, editSelectedRegion: "minor"}

	if r.editBuildingCanAdd(region, "market") {
		t.Fatal("minor bölgede metadata ile izin verilmeyen bina eklenebilir görünüyor")
	}
	if !r.editBuildingCanAdd(region, "granary") {
		t.Fatal("minor bölgede metadata ile izin verilen bina eklenemiyor")
	}
}
