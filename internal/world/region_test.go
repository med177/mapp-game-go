package world

import (
	"encoding/json"
	"testing"
)

func TestNewMinorRegionFromParentUsesIndependentLowIncomeDefaults(t *testing.T) {
	parent := &Region{
		ID:           "anatolia",
		Terrain:      TerrainCoast,
		OwnerID:      "candar_bey",
		ShapeID:      "TUR",
		TaxRate:      28,
		Satisfaction: 60,
		Religion:     "sunni",
	}

	minor := NewMinorRegionFromParent("sinop_castle", parent, 1173, 443)
	if minor == nil {
		t.Fatal("küçük alt bölge oluşturulmadı")
	}
	if !minor.IsMinorRegion || minor.ParentRegionID != parent.ID {
		t.Fatalf("alt bölge ilişkisi = minor:%v parent:%q", minor.IsMinorRegion, minor.ParentRegionID)
	}
	if minor.OwnerID != parent.OwnerID || minor.ShapeID != parent.ShapeID {
		t.Fatalf("ana bölgeden sahip/shape kopyalanmadı: owner=%q shape=%q", minor.OwnerID, minor.ShapeID)
	}
	if minor.BaseGoldIncome != MinorRegionDefaultBaseGoldIncome || minor.TradeCapacity != 0 || minor.Population != 0 {
		t.Fatalf("küçük alt bölge varsayılan ekonomisi beklenmedik: gold=%d trade=%d population=%d", minor.BaseGoldIncome, minor.TradeCapacity, minor.Population)
	}
	if len(minor.Settlements) != 0 || len(minor.Buildings) != 0 {
		t.Fatalf("küçük alt bölge başlangıçta settlement/bina taşımamalı")
	}
}

func TestMinorRegionUsesBuildingMinorRegionsFlag(t *testing.T) {
	minor := &Region{IsMinorRegion: true}
	for _, typ := range []SettlementType{SettlementFortress, SettlementPort} {
		if !minor.AllowsSettlementType(typ) {
			t.Fatalf("küçük alt bölge %q tipini kabul etmedi", typ)
		}
	}
	for _, typ := range []SettlementType{SettlementCity, SettlementTown} {
		if minor.AllowsSettlementType(typ) {
			t.Fatalf("küçük alt bölge %q tipini kabul etti", typ)
		}
	}
	if !minor.AllowsBuilding("walls", true) || !minor.AllowsBuilding("port", true) {
		t.Fatal("minor_regions=true olan bina reddedildi")
	}
	if minor.AllowsBuilding("market", false) {
		t.Fatal("minor_regions=false olan bina reddedildi")
	}
	if !minor.AllowsBuilding("market", true) {
		t.Fatal("minor_regions=true olan bina kabul edilmedi")
	}
	minor.Settlements = []Settlement{{Type: SettlementFortress}}
	if !EnsureRequiredSettlementBuildings(minor, true) || len(minor.Buildings) != 1 || minor.Buildings[0] != "walls" {
		t.Fatalf("küçük alt bölge başkent kurallarıyla kısıt dışı bina aldı: %#v", minor.Buildings)
	}
}

func TestMinorRegionUsesScenarioPrivilegeBuildingLevelCap(t *testing.T) {
	minor := &Region{IsMinorRegion: true, IsPrivileged: true}
	ordinaryMinor := &Region{IsMinorRegion: true}
	normal := &Region{}
	privilegedMaxLevel := 3

	for _, buildingID := range []string{"walls", "port"} {
		if got := minor.BuildingLevelCap(buildingID, 5, true, privilegedMaxLevel); got != privilegedMaxLevel {
			t.Fatalf("minor %s seviye tavanı = %d, %d bekleniyordu", buildingID, got, privilegedMaxLevel)
		}
		if got := normal.BuildingLevelCap(buildingID, 5, true, privilegedMaxLevel); got != 5 {
			t.Fatalf("normal %s seviye tavanı = %d, 5 bekleniyordu", buildingID, got)
		}
		if got := ordinaryMinor.BuildingLevelCap(buildingID, 5, true, privilegedMaxLevel); got != 5 {
			t.Fatalf("imtiyazsız minor %s seviye tavanı = %d, 5 bekleniyordu", buildingID, got)
		}
	}
	if got := minor.BuildingLevelCap("market", 3, false, privilegedMaxLevel); got != 3 {
		t.Fatalf("minor market seviye tavanı = %d, tanım tavanı 3 bekleniyordu", got)
	}
	if got := minor.BuildingLevelCap("granary", 5, true, privilegedMaxLevel); got != privilegedMaxLevel {
		t.Fatalf("minor_regions=true granary seviye tavanı = %d, %d bekleniyordu", got, privilegedMaxLevel)
	}
}

func TestSettlementPositionClearRequiresMarkerClearance(t *testing.T) {
	region := &Region{Settlements: []Settlement{{X: 100, Y: 200}}}

	if region.SettlementPositionClear(108, 200, 12) {
		t.Fatal("mevcut marker'a çok yakın koordinat boş kabul edildi")
	}
	if !region.SettlementPositionClear(112, 200, 12) {
		t.Fatal("marker açıklığı kadar uzak koordinat çakışıyor kabul edildi")
	}
}

func TestMinorRegionMetadataRoundTripsThroughJSON(t *testing.T) {
	original := &Region{
		ID:             "sinop_castle",
		IsMinorRegion:  true,
		IsPrivileged:   true,
		ParentRegionID: "sinop",
		BaseGoldIncome: MinorRegionDefaultBaseGoldIncome,
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("JSON encode hatası: %v", err)
	}
	var restored Region
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("JSON decode hatası: %v", err)
	}
	if !restored.IsMinorRegion || !restored.IsPrivileged || restored.ParentRegionID != original.ParentRegionID {
		t.Fatalf("küçük alt bölge metadata round-trip kayboldu: %#v", restored)
	}
}
