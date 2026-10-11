package combat

import (
	"math"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

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

func TestArmyStrengthBreakdownMatchesCombatStrengthFormula(t *testing.T) {
	units := map[string]*army.UnitType{
		"infantry": {ID: "infantry", Attack: 10, Defense: 8, Morale: 20},
	}
	a := &army.Army{
		Units:     []army.Unit{{TypeID: "infantry", CurrentHP: 50}},
		Commander: &army.Commander{IsBoss: true},
	}
	mods := TechMods{AttackMod: 0.10, DefenseMod: 0.05, LandOrganizationMod: -0.20}

	attack := ArmyStrengthBreakdown(a, world.TerrainPlain, units, mods, BattleContextLand, BattleStanceBalanced, true, false)
	defense := ArmyStrengthBreakdown(a, world.TerrainForest, units, mods, BattleContextLand, BattleStanceBalanced, false, false)

	if attack.BaseStrength != 6 || math.Abs(attack.FinalStrength-8.1) > 1e-9 {
		t.Fatalf("saldırı breakdown = base %d, final %.2f; want base 6, final 8.1", attack.BaseStrength, attack.FinalStrength)
	}
	if defense.BaseStrength != 6 || math.Abs(defense.FinalStrength-10.14) > 1e-9 {
		t.Fatalf("savunma breakdown = base %d, final %.2f; want base 6, final 10.14", defense.BaseStrength, defense.FinalStrength)
	}
	if attack.CommanderAttackModifier != 0.25 || attack.CommanderMoraleModifier != 0.20 ||
		defense.CommanderDefenseModifier != 0.25 {
		t.Fatalf("komutan çarpanları breakdown'a yansımadı: attack=%+v defense=%+v", attack, defense)
	}

	combatAttack, combatDefense := battleStrengths(a, a, world.TerrainForest, units, mods, mods, BattleContextLand, BattleStanceBalanced)
	if combatAttack != attack.FinalStrength || combatDefense != defense.FinalStrength {
		t.Fatalf("muharebe (%v/%v) ile breakdown (%v/%v) ayrıştı", combatAttack, combatDefense, attack.FinalStrength, defense.FinalStrength)
	}
}
