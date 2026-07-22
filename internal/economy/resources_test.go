package economy

import (
	"testing"

	"mapp-game-go/internal/faction"
)

func TestGoodNameTRUsesCentralResourceDefs(t *testing.T) {
	tests := []struct {
		good GoodType
		want string
	}{
		{good: GoodGrain, want: "Tahıl"},
		{good: GoodIron, want: "Demir"},
		{good: GoodTimber, want: "Kereste"},
		{good: GoodStone, want: "Taş"},
		{good: GoodSpice, want: "Baharat"},
		{good: GoodCloth, want: "Kumaş"},
	}

	for _, tt := range tests {
		if got := GoodNameTR(tt.good); got != tt.want {
			t.Fatalf("%s için ad eşleşmedi: got=%q want=%q", tt.good, got, tt.want)
		}
	}
}

func TestFactionResourceHelpers(t *testing.T) {
	f := &faction.Faction{
		Gold:   10,
		Grain:  20,
		Iron:   30,
		Timber: 40,
		Stone:  50,
		Spice:  60,
		Cloth:  70,
	}

	if got := FactionResourceAmount(f, ResourceStone); got != 50 {
		t.Fatalf("stone amount mismatch: got=%d want=50", got)
	}

	AddFactionResource(f, ResourceSpice, -15)
	if got := FactionResourceAmount(f, ResourceSpice); got != 45 {
		t.Fatalf("spice update mismatch: got=%d want=45", got)
	}
}

func TestResourceCostShortTRUsesCentralNames(t *testing.T) {
	cost := ResourceCost{
		Gold:   100,
		Grain:  5,
		Timber: 2,
	}

	got := cost.ShortTR()
	want := "100 Altın, 5 Tahıl, 2 Kereste"
	if got != want {
		t.Fatalf("ShortTR mismatch: got=%q want=%q", got, want)
	}
}

func TestEmergencySaleUnitPriceAppliesDiscountAndMinimum(t *testing.T) {
	if got := EmergencySaleUnitPrice(10); got != 7 {
		t.Fatalf("10 altınlık fiyat acil satışta 7 olmalıydı, got=%d", got)
	}
	if got := EmergencySaleUnitPrice(1); got != 1 {
		t.Fatalf("1 altınlık fiyat acil satışta 1'e clamp edilmeli, got=%d", got)
	}
	if got := EmergencySaleUnitPrice(0); got != 0 {
		t.Fatalf("geçersiz fiyat 0 dönmeli, got=%d", got)
	}
}

func TestAutomaticExportUnitPriceUsesLowerDiscountedRate(t *testing.T) {
	if got := AutomaticExportUnitPrice(10); got != 6 {
		t.Fatalf("10 altınlık fiyat otomatik ihracatta 6 olmalıydı, got=%d", got)
	}
	if got := AutomaticExportUnitPrice(1); got != 1 {
		t.Fatalf("minimum otomatik ihracat fiyatı 1 olmalıydı, got=%d", got)
	}
}

func TestStrategicGrainDemandRaisesMarketPrice(t *testing.T) {
	factions := map[faction.FactionID]*faction.Faction{
		"player": {ID: "player", Grain: 20},
	}
	withoutDemand := ComputeMarketPrices(factions)[GoodGrain]
	withDemand := ComputeMarketPricesWithStrategicDemand(factions, map[faction.FactionID]int{"player": 100})[GoodGrain]
	if withoutDemand != 1 || withDemand != 3 {
		t.Fatalf("stratejik tahıl talebi piyasa fiyatını artırmalıydı, without=%d with=%d", withoutDemand, withDemand)
	}
}
