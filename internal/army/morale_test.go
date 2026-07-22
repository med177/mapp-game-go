package army

import "testing"

func TestCurrentMoraleKeepsLegacyArmyAtFullMorale(t *testing.T) {
	a := &Army{Units: []Unit{{TypeID: "inf"}}}

	if got := a.CurrentMorale(); got != DefaultArmyMorale {
		t.Fatalf("legacy/default ordu tam moralden başlamalıydı, got=%d", got)
	}
	if got := a.MoraleStrengthMultiplier(); got != 1 {
		t.Fatalf("tam moral savaş gücünü değiştirmemeli, multiplier=%f", got)
	}
}

func TestApplyMoraleDeltaClampsAndReturnsActualChange(t *testing.T) {
	a := &Army{Morale: 40, Units: []Unit{{TypeID: "inf"}}}

	if got := a.ApplyMoraleDelta(-100); got != -39 || a.Morale != MinArmyMorale {
		t.Fatalf("moral alt sınıra doğru kırpılmalı, delta=%d morale=%d", got, a.Morale)
	}
	if got := a.ApplyMoraleDelta(200); got != MaxArmyMorale-MinArmyMorale || a.Morale != MaxArmyMorale {
		t.Fatalf("moral üst sınıra doğru kırpılmalı, delta=%d morale=%d", got, a.Morale)
	}
}

func TestTotalStrengthUsesArmyMorale(t *testing.T) {
	types := map[string]*UnitType{
		"inf": {ID: "inf", Attack: 100, Morale: 0},
	}
	full := &Army{Morale: 100, Units: []Unit{{TypeID: "inf", CurrentHP: 100}}}
	low := &Army{Morale: 50, Units: []Unit{{TypeID: "inf", CurrentHP: 100}}}

	if got := full.TotalStrength(types); got != 100 {
		t.Fatalf("tam moral temel gücü değiştirmemeli, got=%d", got)
	}
	if got := low.TotalStrength(types); got != 85 {
		t.Fatalf("50 moral yaklaşık %%15 güç kaybı vermeli, got=%d", got)
	}
}
