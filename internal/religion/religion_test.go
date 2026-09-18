package religion

import "testing"

func TestAllContainsHistoricalReligionTypes(t *testing.T) {
	want := []Type{Catholic, Orthodox, Sunni, Shia, Tengri, Pagan, Protestant, Utraquist}
	got := All()
	if len(got) != len(want) {
		t.Fatalf("religion.All() length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("religion.All()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestHistoricalReligionRelations(t *testing.T) {
	tests := []struct {
		name string
		a, b Type
		want int
	}{
		{name: "catholic protestant", a: Catholic, b: Protestant, want: -20},
		{name: "protestant utraquist", a: Protestant, b: Utraquist, want: -20},
		{name: "tengri catholic", a: Tengri, b: Catholic, want: -30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Relation(tt.a, tt.b); got != tt.want {
				t.Fatalf("Relation(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
