package combat

import (
	"math/rand"
	"testing"

	"mapp-game-go/internal/army"
)

func TestApplyCasualtiesLeavesRecoverableDamage(t *testing.T) {
	rand.Seed(1)
	a := &army.Army{
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
		},
	}

	lost := applyCasualties(a, 0.50)

	if lost != 0 {
		t.Fatalf("orta kayipta birimlerin tam silinmesi beklenmiyor, lost=%d", lost)
	}
	for i, u := range a.Units {
		if u.CurrentHP != 50 {
			t.Fatalf("birim %d beklenen 50 HP yerine %d", i, u.CurrentHP)
		}
	}
}

func TestApplyCasualtiesCanKillUnitsOnHeavyLoss(t *testing.T) {
	rand.Seed(2)
	a := &army.Army{
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
		},
	}

	lost := applyCasualties(a, 0.80)

	if lost == 0 {
		t.Fatal("agir kayipta en az bir birim silinmeli")
	}
	if len(a.Units) != 4-lost {
		t.Fatalf("ordu boyutu kayipla uyusmuyor, len=%d lost=%d", len(a.Units), lost)
	}
	for _, u := range a.Units {
		if u.CurrentHP <= 0 || u.CurrentHP >= army.MaxUnitHP {
			t.Fatalf("hayatta kalan birim beklenen aralikta degil, hp=%d", u.CurrentHP)
		}
	}
}
