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
