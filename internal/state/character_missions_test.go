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
	armyID := army.ArmyID("army_available")
	gs := &GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			owner:  {ID: owner, Gold: 500},
			target: {ID: target},
		},
		Commanders: map[string]*army.Commander{diplomat.ID: diplomat},
		Armies:     map[army.ArmyID]*army.Army{armyID: {ID: armyID, OwnerID: string(owner)}},
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
	if gs.AssignCommanderToArmy(diplomat.ID, armyID) || gs.AssignCharacterToCouncilRole(owner, diplomat.ID, CouncilRoleDiplomat) {
		t.Fatal("active mission character was reassigned to another duty")
	}
	if gs.Armies[armyID].Commander != nil || diplomat.AssignedArmyID != "" {
		t.Fatal("active mission character was assigned to an army")
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

func TestArmyAssignedCharacterCannotStartMission(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "cmd_assigned", Type: army.CharacterTypeCommander, OwnerID: string(owner), AssignedArmyID: "army_1",
	}
	currentArmy := &army.Army{ID: "army_1", OwnerID: string(owner), Commander: character}
	gs := &GameState{
		Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner, Gold: 100}},
		Commanders: map[string]*army.Commander{character.ID: character},
		Armies:     map[army.ArmyID]*army.Army{currentArmy.ID: currentArmy},
		CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{{
			ID: "archive", TargetKind: scenario.CharacterMissionTargetNone, DurationTurns: 1, GoldCost: 25,
		}}},
	}

	if options := gs.CharacterMissionOptions(owner, character.ID); len(options) != 0 {
		t.Fatalf("army-assigned character mission options = %+v, want none", options)
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(owner, character.ID, "archive", ""); canStart || reason == "" {
		t.Fatalf("army-assigned character mission validation = (%v, %q), want rejection with reason", canStart, reason)
	}
	if started, reason := gs.StartCharacterMissionTarget(owner, character.ID, "archive", ""); started || reason == "" {
		t.Fatalf("army-assigned character mission start = (%v, %q), want rejection with reason", started, reason)
	}
	if gs.Factions[owner].Gold != 100 || character.ActiveMissionID != "" || currentArmy.Commander != character {
		t.Fatal("rejected mission changed the army assignment, character, or faction gold")
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

func TestCharacterAdventureChancesOutcomeAndAbilitySelection(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_arya", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		Traits: []army.CharacterTrait{"assassin"}, CouncilLevel: 3,
	}
	adventure := &scenario.CharacterAdventureDefinition{
		SuccessChancePercent: 40, PartialSuccessChancePercent: 30,
		RelevantTraits: []string{"assassin"}, TraitBonusPercent: 20, LevelBonusPercent: 5,
	}
	success, partial, failure := (&GameState{}).CharacterAdventureChances(character, adventure)
	if success != 75 || partial != 25 || failure != 0 {
		t.Fatalf("trait/level adventure chances = %d/%d/%d, want 75/25/0", success, partial, failure)
	}
	unrelated := &army.Commander{CouncilLevel: 3, Traits: []army.CharacterTrait{"scholar"}}
	success, partial, failure = (&GameState{}).CharacterAdventureChances(unrelated, adventure)
	if success != 55 || partial != 30 || failure != 15 {
		t.Fatalf("unrelated-trait adventure chances = %d/%d/%d, want 55/30/15", success, partial, failure)
	}

	gs := &GameState{
		Turn: 10, Year: 300, PlayerFactionID: owner,
		Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner, Gold: 500}},
		Commanders: map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{{
			ID: "journey", NameTR: "Macera", Type: scenario.CharacterMissionTypeAdventure,
			TargetKind: scenario.CharacterMissionTargetNone, DurationTurns: 1, GoldCost: 50,
			Adventure: &scenario.CharacterAdventureDefinition{
				SuccessChancePercent: 100,
				SuccessAbilityChoices: []scenario.AdventureAbilityChoice{
					{ID: "shadowcraft", NameTR: "Gölgede Ustalık", Scope: "army", Effect: "army_power", Modifier: 0.03},
					{ID: "fieldcraft", NameTR: "Saha Keşfi", Scope: "faction", Effect: "military_power", Modifier: 0.02},
				},
			},
		}}},
	}
	if started, reason := gs.StartCharacterMissionTarget(owner, character.ID, "journey", ""); !started {
		t.Fatalf("adventure mission start failed: %s", reason)
	}
	roll := character.ActiveMissionOutcomeRoll
	if roll < 0 || roll > 99 {
		t.Fatalf("stored adventure roll = %d, want [0,99]", roll)
	}
	if character.ActiveMissionOutcome != scenario.CharacterAdventureSuccess {
		t.Fatalf("stored outcome = %q, want success", character.ActiveMissionOutcome)
	}
	character.Traits = nil
	character.CouncilLevel = 0
	gs.CouncilConfig.Missions[0].Adventure.SuccessChancePercent = 0
	completed := gs.TickCharacterMissions()
	if len(completed) != 1 || completed[0].AdventureOutcome != scenario.CharacterAdventureSuccess ||
		character.ActiveMissionOutcomeRoll != 0 || len(character.PendingAdventureAbilities) != 2 {
		t.Fatalf("adventure completion/result = %+v, character=%+v", completed, character)
	}
	if options := gs.CharacterMissionOptions(owner, character.ID); len(options) != 0 {
		t.Fatalf("missions were offered before choosing the pending ability: %+v", options)
	}
	if chosen, reason := gs.ChooseCharacterAdventureAbility(owner, character.ID, "fieldcraft"); !chosen {
		t.Fatalf("earned ability could not be chosen: %s", reason)
	}
	if len(character.AdventureAbilities) != 1 || character.AdventureAbilities[0].ID != "fieldcraft" ||
		character.AdventureAbilities[0].Scope != "faction" ||
		character.AdventureAbilities[0].Effect != "military_power" ||
		character.AdventureAbilities[0].Modifier != 0.02 ||
		len(character.PendingAdventureAbilities) != 0 {
		t.Fatalf("learned/pending adventure abilities = %+v / %+v", character.AdventureAbilities, character.PendingAdventureAbilities)
	}
	if chosen, _ := gs.ChooseCharacterAdventureAbility(owner, character.ID, "fieldcraft"); chosen {
		t.Fatal("same adventure ability was granted more than once")
	}
}
