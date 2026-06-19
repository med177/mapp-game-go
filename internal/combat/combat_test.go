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

func TestPreviewBattleWithContextModsUsesContextSpecificModifiers(t *testing.T) {
	types := map[string]*army.UnitType{
		"inf":  {ID: "inf", NameTR: "Piyade", Attack: 12, Defense: 10, Morale: 50},
		"ship": {ID: "ship", NameTR: "Kadırga", Attack: 18, Defense: 14, Morale: 40},
	}

	landing := &army.Army{
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
		},
	}
	coastDefender := &army.Army{
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
		},
	}
	fleet := &army.Army{
		IsNaval: true,
		Units: []army.Unit{
			{TypeID: "ship", CurrentHP: 100},
			{TypeID: "ship", CurrentHP: 100},
			{TypeID: "ship", CurrentHP: 100},
		},
	}
	enemyFleet := &army.Army{
		IsNaval: true,
		Units: []army.Unit{
			{TypeID: "ship", CurrentHP: 100},
			{TypeID: "ship", CurrentHP: 100},
			{TypeID: "ship", CurrentHP: 100},
		},
	}

	amphibious := PreviewBattleWithContextMods(landing, coastDefender, world.TerrainCoast, types, TechMods{}, TechMods{}, BattleContextAmphibious, BattleStanceAggressive)
	naval := PreviewBattleWithContextMods(fleet, enemyFleet, world.TerrainSea, types, TechMods{}, TechMods{}, BattleContextNaval, BattleStanceAggressive)

	if amphibious.StanceSummaryTR == naval.StanceSummaryTR {
		t.Fatalf("savaş tipine göre stance açıklaması değişmeliydi, got=%q", amphibious.StanceSummaryTR)
	}
	if amphibious.AttackStrength == naval.AttackStrength {
		t.Fatalf("bağlam bazlı saldırı çarpanları farklı sonuç vermeliydi, amphibious=%d naval=%d", amphibious.AttackStrength, naval.AttackStrength)
	}
	if BattleContextLabelTR(BattleContextNaval) != "Deniz Muharebesi" {
		t.Fatalf("beklenmeyen deniz savaş etiketi: %q", BattleContextLabelTR(BattleContextNaval))
	}
}
