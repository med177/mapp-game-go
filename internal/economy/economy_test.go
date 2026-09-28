package economy

import "testing"

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
