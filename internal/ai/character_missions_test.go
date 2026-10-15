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
