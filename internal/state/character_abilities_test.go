package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
)

func TestAdventureAbilityEffectsRespectScopeAssignmentAndAvailability(t *testing.T) {
	const owner faction.FactionID = "player"
	factionCharacter := &army.Commander{
		ID: "faction_character", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		AdventureAbilities: []army.CharacterAbility{
			{ID: "fieldcraft", Scope: "faction", Effect: "military_power", Modifier: 0.02},
			{ID: "fieldcraft", Scope: "faction", Effect: "military_power", Modifier: 0.02},
		},
	}
	secondFactionCharacter := &army.Commander{
		ID: "second_faction_character", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		AdventureAbilities: []army.CharacterAbility{
			{ID: "battle_counsel", Scope: "faction", Effect: "military_power", Modifier: 0.02},
		},
	}
	landCommander := &army.Commander{
		ID: "land_commander", Type: army.CharacterTypeCommander, OwnerID: string(owner), AssignedArmyID: "land",
		AdventureAbilities: []army.CharacterAbility{
			{ID: "shadowcraft", Scope: "army", Effect: "army_power", Modifier: 0.03},
			{ID: "shadowcraft", Scope: "army", Effect: "army_power", Modifier: 0.03},
			{ID: "naval_strategy", Scope: "navy", Effect: "navy_power", Modifier: 0.50},
		},
	}
	navalCommander := &army.Commander{
		ID: "naval_commander", Type: army.CharacterTypeCommander, OwnerID: string(owner), AssignedArmyID: "fleet",
		AdventureAbilities: []army.CharacterAbility{
			{ID: "naval_strategy", Scope: "navy", Effect: "navy_power", Modifier: 0.04},
		},
	}
	deadFactionCharacter := &army.Commander{
		ID: "dead_faction_character", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		AdventureAbilities: []army.CharacterAbility{
			{ID: "dead_fieldcraft", Scope: "faction", Effect: "military_power", Modifier: 0.50},
		},
	}
	gs := &GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			owner: {ID: owner},
		},
		DeadCommanderIDs: map[string]bool{deadFactionCharacter.ID: true},
		Commanders: map[string]*army.Commander{
			factionCharacter.ID:       factionCharacter,
			secondFactionCharacter.ID: secondFactionCharacter,
			landCommander.ID:          landCommander,
			navalCommander.ID:         navalCommander,
			deadFactionCharacter.ID:   deadFactionCharacter,
		},
	}

	if got := gs.CouncilMilitaryPowerMod(owner); got != 1.04 {
		t.Fatalf("faction ability modifier = %.2f, want 1.04 (deduplicated per character)", got)
	}
	if got := gs.AdventureAbilityEffectsForArmy(&army.Army{ID: "land", OwnerID: string(owner), Commander: landCommander}); got.ArmyPowerMod != 0.03 || got.NavalPowerMod != 0 {
		t.Fatalf("land assignment effects = %+v, want only +3%% army power", got)
	}
	if got := gs.AdventureAbilityEffectsForArmy(&army.Army{ID: "fleet", OwnerID: string(owner), IsNaval: true, Commander: navalCommander}); got.NavalPowerMod != 0.04 || got.ArmyPowerMod != 0 {
		t.Fatalf("naval assignment effects = %+v, want only +4%% navy power", got)
	}
	if got := gs.AdventureAbilityEffectsForArmy(&army.Army{ID: "other", OwnerID: string(owner), Commander: landCommander}); got != (AdventureAbilityEffects{}) {
		t.Fatalf("unassigned army received ability effects: %+v", got)
	}

	factionCharacter.ActiveMissionID = "active_adventure"
	landCommander.ActiveMissionID = "active_adventure"
	if got := gs.CouncilMilitaryPowerMod(owner); got != 1.02 {
		t.Fatalf("faction modifier while ability holder is away = %.2f, want 1.02", got)
	}
	if got := gs.AdventureAbilityEffectsForArmy(&army.Army{ID: "land", OwnerID: string(owner), Commander: landCommander}); got != (AdventureAbilityEffects{}) {
		t.Fatalf("assigned character on an active mission still contributes: %+v", got)
	}
}
