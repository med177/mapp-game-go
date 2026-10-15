package combat

import (
	"math"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAssignedAdventureAbilitiesAffectOnlyMatchingBattleContext(t *testing.T) {
	const owner faction.FactionID = "player"
	landCommander := &army.Commander{
		ID: "land_commander", Type: army.CharacterTypeCommander, OwnerID: string(owner), AssignedArmyID: "land",
		AdventureAbilities: []army.CharacterAbility{{
			ID: "shadowcraft", Scope: "army", Effect: "army_power", Modifier: 0.03,
		}},
	}
	navalCommander := &army.Commander{
		ID: "naval_commander", Type: army.CharacterTypeCommander, OwnerID: string(owner), AssignedArmyID: "fleet",
		AdventureAbilities: []army.CharacterAbility{{
			ID: "naval_strategy", Scope: "navy", Effect: "navy_power", Modifier: 0.04,
		}},
	}
	gs := &state.GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			owner: {ID: owner},
		},
		Commanders: map[string]*army.Commander{
			landCommander.ID:  landCommander,
			navalCommander.ID: navalCommander,
		},
	}
	unitTypes := map[string]*army.UnitType{
		"infantry": {ID: "infantry", Attack: 10, Defense: 10, Morale: 10},
	}
	land := &army.Army{
		ID: "land", OwnerID: string(owner), Commander: landCommander,
		Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
	}
	landMods := TechModsFor(gs, string(owner))
	ApplyAdventureAbilityMods(gs, land, BattleContextLand, &landMods)
	if landMods.AssignedAbilityMod != 0.03 {
		t.Fatalf("land assigned ability modifier = %.2f, want 0.03", landMods.AssignedAbilityMod)
	}
	landBreakdown := ArmyStrengthBreakdown(land, world.TerrainPlain, unitTypes, landMods, BattleContextLand, BattleStanceBalanced, true, false)
	if landBreakdown.AssignedAbilityModifier != 0.03 ||
		math.Abs(landBreakdown.FinalStrength/float64(landBreakdown.BaseStrength)-1.03) > 1e-9 {
		t.Fatalf("land combat breakdown = %+v, want +3%% from assigned ability", landBreakdown)
	}

	fleet := &army.Army{
		ID: "fleet", OwnerID: string(owner), IsNaval: true, Commander: navalCommander,
		Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
	}
	navalMods := TechModsFor(gs, string(owner))
	ApplyAdventureAbilityMods(gs, fleet, BattleContextNaval, &navalMods)
	if navalMods.AssignedAbilityMod != 0.04 {
		t.Fatalf("naval assigned ability modifier = %.2f, want 0.04", navalMods.AssignedAbilityMod)
	}
	navalBreakdown := ArmyStrengthBreakdown(fleet, world.TerrainPlain, unitTypes, navalMods, BattleContextNaval, BattleStanceBalanced, false, false)
	if navalBreakdown.AssignedAbilityModifier != 0.04 ||
		math.Abs(navalBreakdown.FinalStrength/float64(navalBreakdown.BaseStrength)-1.04) > 1e-9 {
		t.Fatalf("naval combat breakdown = %+v, want +4%% from assigned ability", navalBreakdown)
	}

	wrongContext := TechModsFor(gs, string(owner))
	ApplyAdventureAbilityMods(gs, fleet, BattleContextLand, &wrongContext)
	if wrongContext.AssignedAbilityMod != 0 {
		t.Fatalf("naval ability applied outside naval combat: %+v", wrongContext)
	}
}
