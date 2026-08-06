package economy

import "testing"

func TestTaxSatisfactionDeltaUsesThirtyPercentBaseline(t *testing.T) {
	tests := []struct {
		taxRate int
		want    int
	}{
		{taxRate: 0, want: 15},
		{taxRate: 10, want: 10},
		{taxRate: 20, want: 5},
		{taxRate: 29, want: 0},
		{taxRate: 30, want: 0},
		{taxRate: 39, want: 0},
		{taxRate: 40, want: -10},
		{taxRate: 50, want: -20},
		{taxRate: 100, want: -70},
	}

	for _, tt := range tests {
		if got := TaxSatisfactionDelta(tt.taxRate); got != tt.want {
			t.Errorf("TaxSatisfactionDelta(%d) = %d, want %d", tt.taxRate, got, tt.want)
		}
	}
}
