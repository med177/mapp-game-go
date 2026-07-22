package render

import "testing"

func TestFormatResourceHUDValue(t *testing.T) {
	tests := []struct {
		name    string
		current int
		change  int
		want    string
	}{
		{name: "positive", current: 405, change: 55, want: "+55/405"},
		{name: "negative", current: 405, change: -12, want: "-12/405"},
		{name: "zero", current: 405, change: 0, want: "0/405"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatResourceHUDValue(tt.current, tt.change); got != tt.want {
				t.Fatalf("HUD kaynak değeri yanlış: got=%q want=%q", got, tt.want)
			}
		})
	}
}
