package combat

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
)

// TechModsFor aktif teknoloji etkilerinden savaş modlarını çıkarır.
func TechModsFor(gs *state.GameState, ownerID string) TechMods {
	if gs == nil {
		return TechMods{}
	}
	f, ok := gs.Factions[faction.FactionID(ownerID)]
	if !ok || f == nil || gs.TechTypes == nil {
		return TechMods{}
	}
	fx := tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
	eventAttack, eventDefense := gs.FactionEventCombatModifiers(faction.FactionID(ownerID))
	return TechMods{
		AttackMod:       fx.InfantryAttackMod + fx.CavalryAttackMod + fx.SiegeAttackMod + eventAttack,
		DefenseMod:      fx.LandDefenseMod + eventDefense,
		NavalAttackMod:  fx.NavalAttackMod,
		NavalDefenseMod: fx.NavalDefenseMod,
	}
}
