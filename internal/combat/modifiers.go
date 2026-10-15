package combat

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
)

// TechModsFor aktif teknoloji etkilerinden savaş modlarını çıkarır.
func TechModsFor(gs *state.GameState, ownerID string) TechMods {
	if gs == nil {
		return TechMods{}
	}
	mods := TechMods{
		LandOrganizationMod: -float64(gs.ArmyOrganizationPenaltyPercent(faction.FactionID(ownerID))) / 100.0,
	}
	mods.CouncilPowerMod = gs.CouncilMilitaryPowerMod(faction.FactionID(ownerID)) - 1
	mods.CommanderPowerMod = gs.CouncilCommanderPowerMod(faction.FactionID(ownerID)) - 1
	f, ok := gs.Factions[faction.FactionID(ownerID)]
	if !ok || f == nil || gs.TechTypes == nil {
		return mods
	}
	fx := tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
	eventAttack, eventDefense := gs.FactionEventCombatModifiers(faction.FactionID(ownerID))
	mods.AttackMod = fx.InfantryAttackMod + fx.CavalryAttackMod + fx.SiegeAttackMod + eventAttack
	mods.DefenseMod = fx.LandDefenseMod + eventDefense
	mods.NavalAttackMod = fx.NavalAttackMod
	mods.NavalDefenseMod = fx.NavalDefenseMod
	return mods
}

// ApplyAdventureAbilityMods adds the assigned character's context-specific
// ability. Faction-wide abilities already flow through the council modifier.
func ApplyAdventureAbilityMods(gs *state.GameState, a *army.Army, context BattleContext, mods *TechMods) {
	if gs == nil || a == nil || mods == nil {
		return
	}
	effects := gs.AdventureAbilityEffectsForArmy(a)
	switch NormalizeBattleContext(context) {
	case BattleContextNaval:
		if a.IsNaval {
			mods.AssignedAbilityMod += effects.NavalPowerMod
		}
	default:
		if !a.IsNaval {
			mods.AssignedAbilityMod += effects.ArmyPowerMod
		}
	}
}
