package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIStartsProfitableIndependentCharacterMission(t *testing.T) {
	const owner faction.FactionID = "ai"
	character := &army.Commander{
		ID: "agent", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Agent",
	}
	gs := &state.GameState{
		Year:            300,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			owner:    {ID: owner, Gold: 200},
		},
		Commanders: map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{},
			Missions: []scenario.CharacterMissionDefinition{{
				ID: "profitable", NameTR: "Kârlı Görev", TargetKind: scenario.CharacterMissionTargetNone,
				DurationTurns: 1, GoldCost: 40,
				CompletionEffects: scenario.CharacterMissionEffects{GoldReward: 100},
			}},
		},
	}

	var steps []TurnStep
	runTurnPrelude(gs, owner, &steps)

	if character.ActiveMissionID != "profitable" || gs.Factions[owner].Gold != 160 {
		t.Fatalf("mission=%q gold=%d; want profitable mission started for 40 gold", character.ActiveMissionID, gs.Factions[owner].Gold)
	}
	if len(steps) != 1 || steps[0].Kind != TurnStepInfo {
		t.Fatalf("mission dispatch steps = %#v, want one informational step", steps)
	}
}

func TestAIReservesGoldBeforeStartingCharacterMission(t *testing.T) {
	const owner faction.FactionID = "ai"
	character := &army.Commander{
		ID: "agent", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
	}
	gs := &state.GameState{
		Year:            300,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			owner:    {ID: owner, Gold: aiMinGoldReserve + 20},
		},
		Commanders: map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Missions: []scenario.CharacterMissionDefinition{{
				ID: "costly", TargetKind: scenario.CharacterMissionTargetNone, DurationTurns: 1, GoldCost: 40,
				CompletionEffects: scenario.CharacterMissionEffects{GoldReward: 1000},
			}},
		},
	}

	aiManageCharacterMissions(gs, owner, nil)

	if character.ActiveMissionID != "" || gs.Factions[owner].Gold != aiMinGoldReserve+20 {
		t.Fatalf("mission=%q gold=%d; AI must preserve its minimum reserve", character.ActiveMissionID, gs.Factions[owner].Gold)
	}
}

func TestAILeavesArmyAssignedCharacterOutOfMissions(t *testing.T) {
	const owner faction.FactionID = "ai"
	character := &army.Commander{
		ID: "commander", Type: army.CharacterTypeCommander, OwnerID: string(owner), AssignedArmyID: "army_1",
	}
	currentArmy := &army.Army{ID: "army_1", OwnerID: string(owner), Commander: character}
	gs := &state.GameState{
		Year:            300,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			owner:    {ID: owner, Gold: 200},
		},
		Commanders: map[string]*army.Commander{character.ID: character},
		Armies:     map[army.ArmyID]*army.Army{currentArmy.ID: currentArmy},
		CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{{
			ID: "recon", TargetKind: scenario.CharacterMissionTargetNone, DurationTurns: 1, GoldCost: 40,
			CompletionEffects: scenario.CharacterMissionEffects{GoldReward: 100},
		}}},
	}

	aiManageCharacterMissions(gs, owner, nil)

	if character.ActiveMissionID != "" || gs.Factions[owner].Gold != 200 || currentArmy.Commander != character {
		t.Fatal("AI mission logic assigned an army commander to a mission")
	}
}

func TestAICharacterMissionArmyStrengthIndexSelectsStrongestRegion(t *testing.T) {
	const owner faction.FactionID = "owner"
	const enemy faction.FactionID = "enemy"
	const other faction.FactionID = "other"
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			owner: {ID: owner},
			enemy: {ID: enemy},
			other: {ID: other},
		},
		Regions: map[world.RegionID]*world.Region{
			"weak":   {ID: "weak", OwnerID: string(enemy)},
			"strong": {ID: "strong", OwnerID: string(other)},
		},
		Armies: map[army.ArmyID]*army.Army{
			"enemy_land": {
				ID: "enemy_land", OwnerID: string(enemy), RegionID: "weak",
				Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
			},
			"other_land": {
				ID: "other_land", OwnerID: string(other), RegionID: "strong",
				Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
			},
			"other_navy": {
				ID: "other_navy", OwnerID: string(other), RegionID: "strong", IsNaval: true,
				Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Attack: 10, HP: 100},
		},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Missions: []scenario.CharacterMissionDefinition{{
				ID: "region_recon", TargetKind: scenario.CharacterMissionTargetRegion,
			}},
		},
	}
	strengths := newAICharacterMissionArmyStrengths(gs)

	if got, want := strengths.factionStrength(enemy), 10.0; got != want {
		t.Fatalf("enemy army strength = %v, want %v", got, want)
	}
	if got, want := strengths.factionStrengthByType(other, false), 10.0; got != want {
		t.Fatalf("other land strength = %v, want %v", got, want)
	}
	if got, want := strengths.factionStrengthByType(other, true), 10.0; got != want {
		t.Fatalf("other naval strength = %v, want %v", got, want)
	}
	if got, want := strengths.regionStrength("strong"), 20.0; got != want {
		t.Fatalf("strong region strength = %v, want %v", got, want)
	}
	if got, want := aiCharacterMissionTarget(gs, owner, gs.CouncilConfig.Missions[0], strengths), "strong"; got != want {
		t.Fatalf("AI chose mission target %q, want strongest region %q", got, want)
	}
}

func TestAICommanderManagementLeavesMissionCharacterUnassigned(t *testing.T) {
	const owner faction.FactionID = "ai"
	character := &army.Commander{
		ID: "commander", Type: army.CharacterTypeCommander, OwnerID: string(owner), ActiveMissionID: "recon",
	}
	currentArmy := &army.Army{ID: "army_1", OwnerID: string(owner)}
	gs := &state.GameState{
		Year:       300,
		Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Commanders: map[string]*army.Commander{character.ID: character},
		Armies:     map[army.ArmyID]*army.Army{currentArmy.ID: currentArmy},
	}

	gs.EnsureFactionCommanders(string(owner))

	if currentArmy.Commander == character || character.AssignedArmyID != "" {
		t.Fatal("AI commander management assigned a character that is on a mission")
	}
}

func TestAIKeepsValuableAdvisorOutOfLowValueMission(t *testing.T) {
	const owner faction.FactionID = "ai"
	character := &army.Commander{
		ID: "advisor", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		AssignedCouncilRole: state.CouncilRoleTreasurer,
	}
	gs := &state.GameState{
		Year:            300,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			owner:    {ID: owner, Gold: 1000},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: string(owner), BaseGoldIncome: 1000, TaxRate: 100, Satisfaction: 100},
		},
		Commanders: map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{{
				ID: state.CouncilRoleTreasurer, Effects: scenario.CouncilEffects{IncomeMod: 2},
			}},
			Missions: []scenario.CharacterMissionDefinition{{
				ID: "low_value", RoleID: state.CouncilRoleTreasurer, TargetKind: scenario.CharacterMissionTargetNone,
				DurationTurns: 2, CompletionEffects: scenario.CharacterMissionEffects{GoldReward: 1},
			}},
		},
	}

	aiManageCharacterMissions(gs, owner, nil)

	if character.ActiveMissionID != "" || character.AssignedCouncilRole != state.CouncilRoleTreasurer {
		t.Fatalf("AI sent valuable advisor on a low-value mission: mission=%q role=%q",
			character.ActiveMissionID, character.AssignedCouncilRole)
	}
}

func TestAIChoosesPendingAdventureAbility(t *testing.T) {
	const owner faction.FactionID = "ai"
	character := &army.Commander{
		ID: "adventurer", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		PendingAdventureAbilities: []army.CharacterAbility{{ID: "fieldcraft", NameTR: "Saha Keşfi"}},
	}
	gs := &state.GameState{
		Year:            300,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			owner:    {ID: owner},
		},
		Commanders:    map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{},
	}

	aiManageCharacterMissions(gs, owner, nil)

	if len(character.PendingAdventureAbilities) != 0 || len(character.AdventureAbilities) != 1 ||
		character.AdventureAbilities[0].ID != "fieldcraft" {
		t.Fatalf("AI ability selection pending=%#v learned=%#v", character.PendingAdventureAbilities, character.AdventureAbilities)
	}
}
