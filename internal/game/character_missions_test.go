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
