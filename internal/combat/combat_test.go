package combat

import "testing"

func TestApplyCoreCombatBonusesUsesStrongerDefenseBonus(t *testing.T) {
	atkMods := TechMods{}
	defMods := TechMods{}

	ApplyCoreCombatBonuses(&atkMods, &defMods, true, true)

	if atkMods.CoreAttackMod != CoreAttackBonus {
		t.Fatalf("core saldırı bonusu = %v, want %v", atkMods.CoreAttackMod, CoreAttackBonus)
	}
	if defMods.CoreDefenseMod != CoreDefenseBonus {
		t.Fatalf("core savunma bonusu = %v, want %v", defMods.CoreDefenseMod, CoreDefenseBonus)
	}
	if defMods.CoreDefenseMod <= atkMods.CoreAttackMod {
		t.Fatal("core savunma bonusu saldırı bonusundan yüksek olmalı")
	}
}
