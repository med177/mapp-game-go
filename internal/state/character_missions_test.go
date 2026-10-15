package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/world"
)

func TestRegionMissionTargetsAndCompletion(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{ID: "chr_scout", Type: army.CharacterTypeCharacter, OwnerID: string(owner)}
	gs := &GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			owner: {ID: owner, Gold: 200},
			"foe": {ID: "foe"},
		},
		Regions: map[world.RegionID]*world.Region{
			"owned":    {ID: "owned", OwnerID: string(owner), NameTR: "Oyuncu Bölgesi"},
			"region_z": {ID: "region_z", OwnerID: "foe", NameTR: "Z Bölgesi"},
			"region_a": {ID: "region_a", OwnerID: "foe", NameTR: "A Bölgesi"},
			"sea":      {ID: "sea", OwnerID: "foe", IsSea: true, NameTR: "Deniz"},
			"locked":   {ID: "locked", OwnerID: "foe", IsLocked: true, NameTR: "Kilitli"},
			"terrain":  {ID: "terrain", OwnerID: "foe", IsTerrainArea: true, NameTR: "Arazi"},
		},
		Commanders: map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{{
			ID: "recon", NameTR: "Bölgesel Keşif", TargetKind: scenario.CharacterMissionTargetRegion,
			DurationTurns: 1, GoldCost: 75, CompletionEffects: scenario.CharacterMissionEffects{RegionIntelTurns: 2},
		}}},
	}

	targets := gs.CharacterMissionTargetIDs(owner, "recon")
	if len(targets) != 2 || targets[0] != "region_a" || targets[1] != "region_z" {
		t.Fatalf("region mission targets = %v, want ordered foreign land regions", targets)
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(owner, character.ID, "recon", "sea"); canStart {
		t.Fatal("mission accepted a sea region target")
	} else if reason == "" {
		t.Fatal("invalid region mission target returned no reason")
	}
	if started, reason := gs.StartCharacterMissionTarget(owner, character.ID, "recon", "region_a"); !started {
		t.Fatalf("region mission start failed: %s", reason)
	}
	if gs.Factions[owner].Gold != 125 || character.ActiveMissionTargetID != "region_a" {
		t.Fatalf("region mission cost/target = %d gold, %q", gs.Factions[owner].Gold, character.ActiveMissionTargetID)
	}
	completed := gs.TickCharacterMissions()
	if len(completed) != 1 || completed[0].TargetRegion != "region_a" || completed[0].TargetFaction != "" {
		t.Fatalf("region mission completion target = %+v", completed)
	}
}

func TestRoleLinkedMissionPausesCouncilEffectAndCompletesAfterConfiguredTurns(t *testing.T) {
	const owner faction.FactionID = "player"
	const target faction.FactionID = "target"
	diplomat := &army.Commander{
		ID: "chr_varys", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		Name: "Varys", AssignedCouncilRole: CouncilRoleDiplomat,
	}
	gs := &GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			owner:  {ID: owner, Gold: 500},
			target: {ID: target},
		},
		Commanders: map[string]*army.Commander{diplomat.ID: diplomat},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{{
				ID: CouncilRoleDiplomat, NameTR: "Diplomat",
				Effects: scenario.CouncilEffects{RelationBonusPoints: 3},
			}},
			Missions: []scenario.CharacterMissionDefinition{{
				ID: "envoy", NameTR: "Elçilik", RoleID: CouncilRoleDiplomat,
				TargetKind: scenario.CharacterMissionTargetFaction, DurationTurns: 2, GoldCost: 100,
				CompletionEffects: scenario.CharacterMissionEffects{RelationBonusPoints: 15},
			}},
		},
	}

	if got := gs.CouncilRoleEffects(owner, CouncilRoleDiplomat).RelationBonusPoints; got != 3 {
		t.Fatalf("idle diplomat council effect = %d, want 3", got)
	}
	if canStart, reason := gs.CanStartCharacterMission(owner, diplomat.ID, "envoy", target); !canStart {
		t.Fatalf("role holder could not start its linked mission: %s", reason)
	}
	if started, reason := gs.StartCharacterMission(owner, diplomat.ID, "envoy", target); !started {
		t.Fatalf("mission start failed: %s", reason)
	}
	if gs.Factions[owner].Gold != 400 || diplomat.ActiveMissionTurnsRemaining != 2 {
		t.Fatalf("mission start cost/progress = gold %d, turns %d", gs.Factions[owner].Gold, diplomat.ActiveMissionTurnsRemaining)
	}
	if got := gs.CouncilRoleEffects(owner, CouncilRoleDiplomat).RelationBonusPoints; got != 0 {
		t.Fatalf("diplomat council effect while away = %d, want 0", got)
	}
	if gs.AssignCommanderToArmy(diplomat.ID, "army_missing") || gs.AssignCharacterToCouncilRole(owner, diplomat.ID, CouncilRoleDiplomat) {
		t.Fatal("active mission character was reassigned to another duty")
	}
	if completed := gs.TickCharacterMissions(); len(completed) != 0 || diplomat.ActiveMissionTurnsRemaining != 1 {
		t.Fatalf("first mission tick = %+v, remaining turns %d", completed, diplomat.ActiveMissionTurnsRemaining)
	}
	completed := gs.TickCharacterMissions()
	if len(completed) != 1 || completed[0].Mission.ID != "envoy" || completed[0].TargetFaction != target ||
		completed[0].Mission.CompletionEffects.RelationBonusPoints != 15 {
		t.Fatalf("mission completion = %+v", completed)
	}
	if diplomat.ActiveMissionID != "" || diplomat.ActiveMissionTargetID != "" || diplomat.ActiveMissionTurnsRemaining != 0 {
		t.Fatalf("mission state remained after completion: %+v", diplomat)
	}
	if gs.CouncilRoleHolder(owner, CouncilRoleDiplomat) != diplomat ||
		gs.CouncilRoleEffects(owner, CouncilRoleDiplomat).RelationBonusPoints != 3 {
		t.Fatal("diplomat role did not resume after mission completion")
	}
}

func TestStandaloneMissionRequiresCharacterWithoutCouncilRole(t *testing.T) {
	const owner faction.FactionID = "player"
	free := &army.Commander{ID: "chr_sam", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Samwell"}
	busy := &army.Commander{ID: "chr_sansa", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Sansa", AssignedCouncilRole: CouncilRoleHeir}
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{owner: {ID: owner, Gold: 100}},
		Commanders: map[string]*army.Commander{
			free.ID: free,
			busy.ID: busy,
		},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{{ID: CouncilRoleHeir, NameTR: "Veliaht"}},
			Missions: []scenario.CharacterMissionDefinition{{
				ID: "archive", NameTR: "Arşiv Seferi", TargetKind: scenario.CharacterMissionTargetNone, DurationTurns: 1,
				CompletionEffects: scenario.CharacterMissionEffects{CouncilExperience: 10},
			}},
		},
	}

	if canStart, reason := gs.CanStartCharacterMission(owner, free.ID, "archive", ""); !canStart {
		t.Fatalf("unassigned character could not start a standalone mission: %s", reason)
	}
	if canStart, _ := gs.CanStartCharacterMission(owner, busy.ID, "archive", ""); canStart {
		t.Fatal("character assigned to a council role started a standalone mission")
	}
}
