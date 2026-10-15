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
			Missions: []scenario.CharacterMissionDefinition{
				{
					ID: "archive", NameTR: "Arşiv Seferi", TargetKind: scenario.CharacterMissionTargetNone, DurationTurns: 1,
					CompletionEffects: scenario.CharacterMissionEffects{CouncilExperience: 10},
				},
				{
					ID: "common", NameTR: "Ortak Görev", AllowCouncilRoles: true,
					BlockedFactions: []string{"others"}, TargetKind: scenario.CharacterMissionTargetNone, DurationTurns: 1,
				},
			},
		},
	}

	if canStart, reason := gs.CanStartCharacterMission(owner, free.ID, "archive", ""); !canStart {
		t.Fatalf("unassigned character could not start a standalone mission: %s", reason)
	}
	if canStart, _ := gs.CanStartCharacterMission(owner, busy.ID, "archive", ""); canStart {
		t.Fatal("character assigned to a council role started a standalone mission")
	}
	if options := gs.CharacterMissionOptions(owner, busy.ID); len(options) != 1 || options[0].ID != "common" {
		t.Fatalf("council character mission options = %+v, want common mission only", options)
	}
	if canStart, reason := gs.CanStartCharacterMission(owner, busy.ID, "common", ""); !canStart {
		t.Fatalf("character assigned to a council role could not start common mission: %s", reason)
	}
	if busy.AssignedCouncilRole != CouncilRoleHeir {
		t.Fatalf("common mission changed council assignment to %q", busy.AssignedCouncilRole)
	}
	if started, reason := gs.StartCharacterMissionTarget(owner, busy.ID, "common", ""); !started {
		t.Fatalf("common mission start failed: %s", reason)
	}
	if gs.CouncilRoleHolder(owner, CouncilRoleHeir) != nil {
		t.Fatal("council role effect remained active while its holder was on a common mission")
	}
}

func TestAllowedCharactersRestrictsMissionOptionsAndStart(t *testing.T) {
	const owner faction.FactionID = "player"
	const otherOwner faction.FactionID = "other"
	allowed := &army.Commander{ID: "chr_arya_stark", Type: army.CharacterTypeCharacter, OwnerID: string(owner)}
	blocked := &army.Commander{ID: "chr_sansa_stark", Type: army.CharacterTypeCharacter, OwnerID: string(otherOwner)}
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			owner:      {ID: owner, Gold: 100},
			otherOwner: {ID: otherOwner, Gold: 100},
		},
		Commanders: map[string]*army.Commander{
			allowed.ID: allowed,
			blocked.ID: blocked,
		},
		CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{
			{
				ID: "faceless_journey", NameTR: "Yüzsüzlerin Yolculuğu", TargetKind: scenario.CharacterMissionTargetNone,
				DurationTurns: 1, AllowedCharacters: []string{allowed.ID},
			},
			{
				ID: "faction_archive", NameTR: "Fraksiyona Özel Arşiv", TargetKind: scenario.CharacterMissionTargetNone,
				DurationTurns: 1, AllowedFactions: []string{string(owner)},
			},
			{
				ID: "blocked_faction_archive", NameTR: "Fraksiyona Kapalı Arşiv", TargetKind: scenario.CharacterMissionTargetNone,
				DurationTurns: 1, BlockedFactions: []string{string(owner)},
			},
			{
				ID: "deny_overrides_allow", NameTR: "Engel Öncelikli Arşiv", TargetKind: scenario.CharacterMissionTargetNone,
				DurationTurns: 1, AllowedFactions: []string{string(owner)}, BlockedFactions: []string{string(owner)},
			},
			{
				ID: "flag_gated_mission", NameTR: "Bayrak Koşullu Görev", TargetKind: scenario.CharacterMissionTargetNone,
				DurationTurns: 1, RequiresFlags: []string{"journey_unlocked", "second_prerequisite"},
				BlocksFlags: []string{"journey_closed"},
			},
		}},
	}

	if options := gs.CharacterMissionOptions(owner, allowed.ID); len(options) != 2 ||
		options[0].ID != "faceless_journey" || options[1].ID != "faction_archive" {
		t.Fatalf("allowed faction mission options = %+v, want only allowed missions", options)
	}
	if options := gs.CharacterMissionOptions(otherOwner, blocked.ID); len(options) != 1 ||
		options[0].ID != "blocked_faction_archive" {
		t.Fatalf("other faction mission options = %+v, want blocked_faction_archive only", options)
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(owner, allowed.ID, "faceless_journey", ""); !canStart {
		t.Fatalf("allowed character could not start mission: %s", reason)
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(owner, blocked.ID, "faceless_journey", ""); canStart || reason == "" {
		t.Fatalf("blocked character mission validation = (%v, %q), want rejection with reason", canStart, reason)
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(owner, allowed.ID, "faction_archive", ""); !canStart {
		t.Fatalf("allowed faction could not start mission: %s", reason)
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(otherOwner, blocked.ID, "faction_archive", ""); canStart || reason == "" {
		t.Fatalf("blocked faction mission validation = (%v, %q), want rejection with reason", canStart, reason)
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(owner, allowed.ID, "blocked_faction_archive", ""); canStart || reason == "" {
		t.Fatalf("blocked faction mission validation = (%v, %q), want rejection with reason", canStart, reason)
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(otherOwner, blocked.ID, "blocked_faction_archive", ""); !canStart {
		t.Fatalf("non-blocked faction could not start mission: %s", reason)
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(owner, allowed.ID, "deny_overrides_allow", ""); canStart || reason == "" {
		t.Fatalf("blocked allowed faction validation = (%v, %q), want blocked_factions to override allowed_factions", canStart, reason)
	}
	if options := gs.CharacterMissionOptions(owner, allowed.ID); containsCharacterMission(options, "flag_gated_mission") {
		t.Fatal("mission with an unmet required event flag was offered")
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(owner, allowed.ID, "flag_gated_mission", ""); canStart || reason == "" {
		t.Fatalf("mission without required event flag validation = (%v, %q), want rejection", canStart, reason)
	}
	gs.FiredEventIDs = map[string]bool{"flag:journey_unlocked": true}
	if options := gs.CharacterMissionOptions(owner, allowed.ID); containsCharacterMission(options, "flag_gated_mission") {
		t.Fatal("mission was offered before all required event flags were set")
	}
	gs.FiredEventIDs["flag:second_prerequisite"] = true
	if options := gs.CharacterMissionOptions(owner, allowed.ID); !containsCharacterMission(options, "flag_gated_mission") {
		t.Fatal("mission was not offered after all required event flags were set")
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(owner, allowed.ID, "flag_gated_mission", ""); !canStart {
		t.Fatalf("mission with required event flag could not start: %s", reason)
	}
	gs.FiredEventIDs["flag:journey_closed"] = true
	if options := gs.CharacterMissionOptions(owner, allowed.ID); containsCharacterMission(options, "flag_gated_mission") {
		t.Fatal("mission with a blocking event flag was offered")
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(owner, allowed.ID, "flag_gated_mission", ""); canStart || reason == "" {
		t.Fatalf("mission with blocking event flag validation = (%v, %q), want rejection", canStart, reason)
	}
}

func containsCharacterMission(missions []scenario.CharacterMissionDefinition, missionID string) bool {
	for _, mission := range missions {
		if mission.ID == missionID {
			return true
		}
	}
	return false
}

func TestCharacterMissionMinYearRestrictsOptionsAndStart(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{ID: "chr_arya", Type: army.CharacterTypeCharacter, OwnerID: string(owner)}
	gs := &GameState{
		Year:       299,
		Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner, Gold: 100}},
		Commanders: map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{{
			ID: "historical_mission", NameTR: "Tarihli Görev", TargetKind: scenario.CharacterMissionTargetNone,
			MinYear: 300, DurationTurns: 1,
		}}},
	}

	if options := gs.CharacterMissionOptions(owner, character.ID); len(options) != 0 {
		t.Fatalf("mission offered before min_year: %+v", options)
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(owner, character.ID, "historical_mission", ""); canStart || reason == "" {
		t.Fatalf("mission start before min_year = (%v, %q), want rejection", canStart, reason)
	}

	gs.Year = 300
	if options := gs.CharacterMissionOptions(owner, character.ID); len(options) != 1 || options[0].ID != "historical_mission" {
		t.Fatalf("mission options at min_year = %+v, want historical_mission", options)
	}
	if canStart, reason := gs.CanStartCharacterMissionTarget(owner, character.ID, "historical_mission", ""); !canStart {
		t.Fatalf("mission could not start at min_year: %s", reason)
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
