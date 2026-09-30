package economy

import (
	"testing"

	"mapp-game-go/internal/faction"
)

func TestRegionalTaxBaseCapSeparatesLocalTaxFromTradeInfrastructure(t *testing.T) {
	tests := []struct {
		name                      string
		population, tradeCapacity int
		hasPort, hasMarket        bool
		want                      int
	}{
		{name: "rural region", population: 300, tradeCapacity: 2, want: 210},
		{name: "port market", population: 300, tradeCapacity: 2, hasPort: true, hasMarket: true, want: 260},
		{name: "negative trade capacity", population: 128, tradeCapacity: -2, want: 104},
		{name: "empty region", population: 0, tradeCapacity: 16, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RegionalTaxBaseCap(tt.population, tt.tradeCapacity, tt.hasPort, tt.hasMarket); got != tt.want {
				t.Fatalf("RegionalTaxBaseCap() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPrivilegedMinorTradeRouteHasNoRoutePaymentOrCustoms(t *testing.T) {
	seller := &faction.Faction{ID: "operator", Grain: 10, Gold: 20}
	grantor := &faction.Faction{ID: "grantor", Gold: 7}
	route := &TradeRoute{
		FromFactionID:     "operator",
		ToFactionID:       "grantor",
		Good:              GoodGrain,
		AmountPerTurn:     3,
		GoldPerUnit:       5,
		IsPrivilegedMinor: true,
	}

	_, transfers := ApplyTradeRoutesWithTransfersAndCustoms(map[faction.FactionID]*faction.Faction{
		"operator": seller,
		"grantor":  grantor,
	}, []*TradeRoute{route}, func(faction.FactionID) int { return 20 })

	if seller.Grain != 7 || grantor.Grain != 3 {
		t.Fatalf("imtiyaz rotasında mal transferi yanlış: operator=%d grantor=%d", seller.Grain, grantor.Grain)
	}
	if seller.Gold != 20 || grantor.Gold != 7 {
		t.Fatalf("imtiyaz rotası normal ödeme/gümrük uyguladı: operator=%d grantor=%d", seller.Gold, grantor.Gold)
	}
	if route.GoldEarned() != 0 || len(transfers) != 1 || transfers[0].Amount != 0 || transfers[0].CustomsAmount != 0 {
		t.Fatalf("imtiyaz rotası finansal transferi sıfırlanmadı: earned=%d transfers=%+v", route.GoldEarned(), transfers)
	}
}
