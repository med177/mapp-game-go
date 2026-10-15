package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestResolveRegionCharacterMissionGrantsRegionIntel(t *testing.T) {
	const owner faction.FactionID = "player"
	regionID := world.RegionID("scouted")
	character := &army.Commander{
		ID: "chr_scout", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Scout",
		ActiveMissionID: "recon", ActiveMissionTargetID: string(regionID), ActiveMissionTurnsRemaining: 1,
	}
	gs := &state.GameState{
		PlayerFactionID: owner,
		Factions:        map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Regions:         map[world.RegionID]*world.Region{regionID: {ID: regionID, OwnerID: "target", NameTR: "Sınır Bölgesi"}},
		Commanders:      map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{{
			ID: "recon", NameTR: "Bölgesel Keşif", TargetKind: scenario.CharacterMissionTargetRegion, DurationTurns: 1,
			CompletionEffects: scenario.CharacterMissionEffects{RegionIntelTurns: 2},
		}}},
	}
	game := &Game{gs: gs}
	game.resolveCharacterMissionCompletions()
	if !state.HasTemporaryRegionMilitaryIntel(gs, owner, regionID) ||
		gs.TemporaryRegionMilitaryIntel[owner][regionID] != 2 {
		t.Fatalf("region mission intel = %v turns, want 2", gs.TemporaryRegionMilitaryIntel[owner][regionID])
	}
}

func TestResolveCharacterMissionCompletionsAppliesConfiguredRewards(t *testing.T) {
	const owner faction.FactionID = "player"
	const target faction.FactionID = "target"
	character := &army.Commander{
		ID: "chr_diplomat", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Diplomat",
		ActiveMissionID: "envoy", ActiveMissionTargetID: string(target), ActiveMissionTurnsRemaining: 1,
		CouncilExperience: 90, CouncilLevel: 1,
	}

	gs := &state.GameState{
		PlayerFactionID: owner,
		Factions: map[faction.FactionID]*faction.Faction{
			owner:  {ID: owner, Gold: 200},
			target: {ID: target},
		},
		Commanders: map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{{
			ID: "envoy", NameTR: "Diplomatik Elçilik", TargetKind: scenario.CharacterMissionTargetFaction, DurationTurns: 1,
			CompletionEffects: scenario.CharacterMissionEffects{
				RelationBonusPoints: 15,
				GoldReward:          75,
				CouncilExperience:   20,
				MilitaryIntelTurns:  2,
			},
		}}},
	}

	game := &Game{gs: gs}
	game.resolveCharacterMissionCompletions()

	if got := gs.Factions[owner].Gold; got != 275 {
		t.Fatalf("gold after mission completion = %d, want 275", got)
	}
	if character.CouncilExperience != 110 || character.CouncilLevel != 2 {
		t.Fatalf("council progression after completion = %d XP, level %d; want 110 XP, level 2", character.CouncilExperience, character.CouncilLevel)
	}
	relation := gs.Relations[faction.RelationKey(owner, target)]
	if relation == nil || relation.ScoreFrom(owner) != 15 || relation.ScoreFrom(target) != 15 {
		t.Fatalf("mission relation reward = %+v, want both directions at 15", relation)
	}
	if !state.HasTemporaryMilitaryIntel(gs, owner, target) || gs.TemporaryMilitaryIntel[owner][target] != 2 {
		t.Fatalf("mission military intel = %v turns, want 2", gs.TemporaryMilitaryIntel[owner][target])
	}
	if character.ActiveMissionID != "" || character.ActiveMissionTargetID != "" || character.ActiveMissionTurnsRemaining != 0 {
		t.Fatalf("completed mission state was not cleared: %+v", character)
	}
}

func TestResolveAdventureMissionSuccessPartialAndFailure(t *testing.T) {
	tests := []struct {
		name        string
		adventure   scenario.CharacterAdventureDefinition
		wantOutcome string
		wantGold    int
		wantXP      int
		wantInjury  bool
	}{
		{
			name: "success",
			adventure: scenario.CharacterAdventureDefinition{
				SuccessChancePercent: 100, SuccessGoldReward: 100,
				SuccessCouncilExperience: 15,
				SuccessAbilityChoices:    []scenario.AdventureAbilityChoice{{ID: "fieldcraft", NameTR: "Saha Keşfi"}},
			},
			wantOutcome: scenario.CharacterAdventureSuccess, wantGold: 100, wantXP: 15,
		},
		{
			name: "partial",
			adventure: scenario.CharacterAdventureDefinition{
				PartialSuccessChancePercent: 100, PartialGoldReward: 40, PartialCouncilExperience: 5,
			},
			wantOutcome: scenario.CharacterAdventurePartial, wantGold: 40, wantXP: 5,
		},
		{
			name: "failure",
			adventure: scenario.CharacterAdventureDefinition{
				FailureUnavailableTurns: 2,
			},
			wantOutcome: scenario.CharacterAdventureFailure, wantInjury: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const owner faction.FactionID = "player"
			character := &army.Commander{
				ID: "chr_arya", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
				ActiveMissionID: "journey", ActiveMissionTurnsRemaining: 1,
			}
			gs := &state.GameState{
				Turn: 10, PlayerFactionID: owner,
				Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
				Commanders: map[string]*army.Commander{character.ID: character},
				CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{{
					ID: "journey", NameTR: "Macera", Type: scenario.CharacterMissionTypeAdventure, DurationTurns: 1,
					Adventure: &test.adventure,
				}}},
			}
			(&Game{gs: gs}).resolveCharacterMissionCompletions()
			if got := gs.Factions[owner].Gold; got != test.wantGold {
				t.Errorf("adventure gold = %d, want %d", got, test.wantGold)
			}
			if character.CouncilExperience != test.wantXP {
				t.Errorf("adventure XP = %d, want %d", character.CouncilExperience, test.wantXP)
			}
			if test.wantInjury {
				if !gs.CommanderUnavailable(character) || character.UnavailableUntilTurn != gs.Turn+2 ||
					character.UnavailableReasonTR != "Macera sırasında yaralandı" {
					t.Errorf("failure injury state = %+v, want two-turn temporary injury", character)
				}
			} else if test.name == "success" && len(character.PendingAdventureAbilities) != 1 {
				t.Errorf("success pending abilities = %+v, want one selection", character.PendingAdventureAbilities)
			}
		})
	}
}
