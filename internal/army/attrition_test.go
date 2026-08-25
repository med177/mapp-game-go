package army

import "testing"

func TestApplyAttritionPercentReducesHPAndRemovesDeadUnits(t *testing.T) {
	a := &Army{Units: []Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 1}}}
	lost := a.ApplyAttritionPercent(20)
	if lost != 1 {
		t.Fatalf("düşük HP'li birim ölmeliydi: lost=%d", lost)
	}
	if len(a.Units) != 1 || a.Units[0].CurrentHP != 80 {
		t.Fatalf("hayatta kalan birimin HP'si %%20 azalmalıydı: units=%+v", a.Units)
	}
}

func TestApplyAttritionPercentZeroIsNoOp(t *testing.T) {
	a := &Army{Units: []Unit{{TypeID: "inf", CurrentHP: 50}}}
	if lost := a.ApplyAttritionPercent(0); lost != 0 || a.Units[0].CurrentHP != 50 {
		t.Fatalf("sıfır yüzde etkisiz olmalı: lost=%d hp=%d", lost, a.Units[0].CurrentHP)
	}
}

func TestApplyWinterAttritionStillAppliesTenPercent(t *testing.T) {
	a := &Army{Units: []Unit{{TypeID: "inf", CurrentHP: 100}}}
	a.ApplyWinterAttrition()
	if a.Units[0].CurrentHP != 90 {
		t.Fatalf("kış erozyonu %%10 olmalı: got=%d", a.Units[0].CurrentHP)
	}
}
