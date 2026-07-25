package render

import (
	"testing"

	"mapp-game-go/internal/world"
)

func TestRegionPopulationDisplayTextIncludesTotalAndBreakdown(t *testing.T) {
	region := &world.Region{
		Population:      1000,
		RuralPopulation: 650,
		Settlements: []world.Settlement{
			{Population: 200},
			{Population: 150},
		},
	}

	want := "Nüfus: 1000  (Kırsal: 650 / Yerleşim: 350)"
	if got := regionPopulationDisplayText(region); got != want {
		t.Fatalf("bölge nüfus satırı toplam ve kırılımı göstermeli: got=%q want=%q", got, want)
	}
}
