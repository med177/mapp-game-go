package combat

import (
	"math/rand"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
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

func TestPreviewBattleWithModsReflectsStanceTradeoff(t *testing.T) {
	types := map[string]*army.UnitType{
		"inf": {ID: "inf", NameTR: "Piyade", Attack: 12, Defense: 10, Morale: 50},
	}
	atk := &army.Army{
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
		},
	}
	def := &army.Army{
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
		},
	}

	aggressive := PreviewBattleWithMods(atk, def, world.TerrainPlain, types, TechMods{}, TechMods{}, BattleStanceAggressive)
	defensive := PreviewBattleWithMods(atk, def, world.TerrainPlain, types, TechMods{}, TechMods{}, BattleStanceDefensive)

	if aggressive.WinChance <= defensive.WinChance {
		t.Fatalf("agresif duruş daha yüksek zafer şansı vermeli, aggressive=%d defensive=%d", aggressive.WinChance, defensive.WinChance)
	}
	if aggressive.AttackStrength <= defensive.AttackStrength {
		t.Fatalf("agresif duruş efektif saldırı gücünü artırmalı, aggressive=%d defensive=%d", aggressive.AttackStrength, defensive.AttackStrength)
	}
	if aggressive.AttackerHPExpected == 0 || aggressive.DefenderHPExpected == 0 {
		t.Fatalf("preview HP kaybı üretmeliydi, attacker=%d defender=%d", aggressive.AttackerHPExpected, aggressive.DefenderHPExpected)
	}
	if aggressive.DefenderLossMax != len(def.Units) {
		t.Fatalf("zafer senaryosunda savunucu tam silinebilmeliydi, got=%d want=%d", aggressive.DefenderLossMax, len(def.Units))
	}
}
