package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIAdoptsOneNeutralCharacterToFillCouncilVacancy(t *testing.T) {
	const aiFaction faction.FactionID = "ai"
	first := &army.Commander{
		ID: "chr_a", Type: army.CharacterTypeCharacter, OwnerID: state.NeutralCharacterOwnerID, Name: "A",
	}
	second := &army.Commander{
		ID: "chr_b", Type: army.CharacterTypeAdministrator, OwnerID: state.NeutralCharacterOwnerID, Name: "B",
	}
	gs := &state.GameState{
		Year:            300,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":  {ID: "player"},
			aiFaction: {ID: aiFaction},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: string(aiFaction), BaseGoldIncome: 1000, TaxRate: 100, Satisfaction: 100},
		},
		Commanders: map[string]*army.Commander{
			first.ID:  first,
			second.ID: second,
		},
		CouncilConfig: scenario.DefaultCharacterCouncilConfig(),
	}

	aiManageNeutralCharacters(gs, aiFaction)

	if first.OwnerID != string(aiFaction) || first.AssignedCouncilRole != state.CouncilRoleTreasurer {
		t.Fatalf("AI-adopted character = owner %q, role %q; want AI faction and treasurer", first.OwnerID, first.AssignedCouncilRole)
	}
	if second.OwnerID != state.NeutralCharacterOwnerID || second.AssignedCouncilRole != "" {
		t.Fatalf("second candidate was also adopted: owner %q, role %q", second.OwnerID, second.AssignedCouncilRole)
	}
	if roster := gs.CharacterRoster(string(aiFaction)); len(roster) != 1 || roster[0] != first {
		t.Fatalf("AI character roster = %#v, want only adopted character", roster)
	}
}

func TestAIReassignsCouncilRoleToCharacterWithBestExpectedReturn(t *testing.T) {
	const owner faction.FactionID = "ai"
	ordinary := &army.Commander{
		ID: "chr_a", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		AssignedCouncilRole: state.CouncilRoleTreasurer,
	}
	steward := &army.Commander{
		ID: "chr_z", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		Traits: []army.CharacterTrait{"skilled_steward"},
	}
	config := scenario.CharacterCouncilConfig{
		Roles: []scenario.CouncilRoleDefinition{{
			ID: state.CouncilRoleTreasurer, NameTR: "Hazinedar",
			Effects: scenario.CouncilEffects{IncomeMod: 1.01},
		}},
		Traits: map[string]scenario.CharacterTraitDefinition{
			"skilled_steward": {
				NameTR: "Usta Yönetici",
				CouncilEffects: map[string]scenario.CouncilEffects{
					state.CouncilRoleTreasurer: {IncomeMod: 1.10},
				},
			},
		},
	}
	gs := &state.GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			owner: {ID: owner},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: string(owner), BaseGoldIncome: 1000, TaxRate: 100, Satisfaction: 100},
		},
		Commanders: map[string]*army.Commander{
			ordinary.ID: ordinary,
			steward.ID:  steward,
		},
		CouncilConfig: config,
	}
	aiManageNeutralCharacters(gs, owner)

	if ordinary.AssignedCouncilRole != "" || steward.AssignedCouncilRole != state.CouncilRoleTreasurer {
		t.Fatalf("council assignments = ordinary:%q skilled:%q; want skilled steward in treasurer role", ordinary.AssignedCouncilRole, steward.AssignedCouncilRole)
	}
}

func TestAIChoosesCouncilAssignmentsForBestCombinedReturn(t *testing.T) {
	const owner faction.FactionID = "ai"
	adaptable := &army.Commander{
		ID: "chr_a", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		Traits: []army.CharacterTrait{"good_treasurer", "great_diplomat"},
	}
	treasurer := &army.Commander{
		ID: "chr_b", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		Traits: []army.CharacterTrait{"great_treasurer", "poor_diplomat"},
	}
	config := scenario.CharacterCouncilConfig{
		Roles: []scenario.CouncilRoleDefinition{
			{ID: state.CouncilRoleTreasurer, NameTR: "Hazinedar", Effects: scenario.CouncilEffects{IncomeMod: 1.01}},
			{ID: state.CouncilRoleDiplomat, NameTR: "Diplomat", Effects: scenario.CouncilEffects{IncomeMod: 1.01}},
		},
		Traits: map[string]scenario.CharacterTraitDefinition{
			"good_treasurer": {
				NameTR: "İyi Hazinedar",
				CouncilEffects: map[string]scenario.CouncilEffects{
					state.CouncilRoleTreasurer: {IncomeMod: 1.10},
				},
			},
			"great_diplomat": {
				NameTR: "Usta Diplomat",
				CouncilEffects: map[string]scenario.CouncilEffects{
					state.CouncilRoleDiplomat: {IncomeMod: 1.20},
				},
			},
			"great_treasurer": {
				NameTR: "Usta Hazinedar",
				CouncilEffects: map[string]scenario.CouncilEffects{
					state.CouncilRoleTreasurer: {IncomeMod: 1.09},
				},
			},
			"poor_diplomat": {
				NameTR: "Zayıf Diplomat",
				CouncilEffects: map[string]scenario.CouncilEffects{
					state.CouncilRoleDiplomat: {IncomeMod: 1.01},
				},
			},
		},
	}
	gs := &state.GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			owner: {ID: owner},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: string(owner), BaseGoldIncome: 1000, TaxRate: 100, Satisfaction: 100},
		},
		Commanders: map[string]*army.Commander{
			adaptable.ID: adaptable,
			treasurer.ID: treasurer,
		},
		CouncilConfig: config,
	}

	aiManageNeutralCharacters(gs, owner)

	if treasurer.AssignedCouncilRole != state.CouncilRoleTreasurer ||
		adaptable.AssignedCouncilRole != state.CouncilRoleDiplomat {
		t.Fatalf("council assignments = adaptable:%q treasurer:%q; want best combined role return",
			adaptable.AssignedCouncilRole, treasurer.AssignedCouncilRole)
	}
}

func TestAIAdoptsNeutralCharacterOnlyWhenItImprovesCouncilReturn(t *testing.T) {
	const owner faction.FactionID = "ai"
	incumbent := &army.Commander{
		ID: "chr_incumbent", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		AssignedCouncilRole: state.CouncilRoleTreasurer,
	}
	candidate := &army.Commander{
		ID: "chr_candidate", Type: army.CharacterTypeCharacter, OwnerID: state.NeutralCharacterOwnerID,
		Traits: []army.CharacterTrait{"skilled_steward"},
	}
	config := scenario.CharacterCouncilConfig{
		Roles: []scenario.CouncilRoleDefinition{{
			ID: state.CouncilRoleTreasurer, NameTR: "Hazinedar",
			Effects: scenario.CouncilEffects{IncomeMod: 1.01},
		}},
		Traits: map[string]scenario.CharacterTraitDefinition{
			"skilled_steward": {
				NameTR: "Usta Yönetici",
				CouncilEffects: map[string]scenario.CouncilEffects{
					state.CouncilRoleTreasurer: {IncomeMod: 1.10},
				},
			},
		},
	}
	gs := &state.GameState{
		Year:            300,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			owner:    {ID: owner},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: string(owner), BaseGoldIncome: 1000, TaxRate: 100, Satisfaction: 100},
		},
		Commanders: map[string]*army.Commander{
			incumbent.ID: incumbent,
			candidate.ID: candidate,
		},
		CouncilConfig: config,
	}

	aiManageNeutralCharacters(gs, owner)

	if candidate.OwnerID != string(owner) || candidate.AssignedCouncilRole != state.CouncilRoleTreasurer ||
		incumbent.AssignedCouncilRole != "" {
		t.Fatalf("candidate/incumbent state = (%q, %q) / %q; want candidate adopted as treasurer and incumbent unassigned",
			candidate.OwnerID, candidate.AssignedCouncilRole, incumbent.AssignedCouncilRole)
	}
}

func TestAIAccountsForAdoptionFeeAndCharacterUpkeep(t *testing.T) {
	const owner faction.FactionID = "ai"
	candidate := &army.Commander{
		ID: "chr_candidate", Type: army.CharacterTypeCharacter, OwnerID: state.NeutralCharacterOwnerID,
		AdoptionCostGold: 500, GoldUpkeep: 10,
		Traits: []army.CharacterTrait{"skilled_steward"},
	}
	config := scenario.CharacterCouncilConfig{
		Roles: []scenario.CouncilRoleDefinition{{
			ID: state.CouncilRoleTreasurer, NameTR: "Hazinedar",
			Effects: scenario.CouncilEffects{IncomeMod: 1.01},
		}},
		Traits: map[string]scenario.CharacterTraitDefinition{
			"skilled_steward": {
				NameTR: "Usta Yönetici",
				CouncilEffects: map[string]scenario.CouncilEffects{
					state.CouncilRoleTreasurer: {IncomeMod: 1.10},
				},
			},
		},
	}
	gs := &state.GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			owner: {ID: owner, Gold: 1000},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: string(owner), BaseGoldIncome: 1000, TaxRate: 100, Satisfaction: 100},
		},
		Commanders:    map[string]*army.Commander{candidate.ID: candidate},
		CouncilConfig: config,
	}

	gs.Factions[owner].Gold = 499
	aiManageNeutralCharacters(gs, owner)
	if candidate.OwnerID != state.NeutralCharacterOwnerID || candidate.AssignedCouncilRole != "" {
		t.Fatal("AI adopted a candidate it could not afford")
	}

	gs.Factions[owner].Gold = 1000
	aiManageNeutralCharacters(gs, owner)

	if candidate.OwnerID != string(owner) || candidate.AssignedCouncilRole != state.CouncilRoleTreasurer {
		t.Fatalf("AI failed to adopt a profitable candidate: owner=%q role=%q", candidate.OwnerID, candidate.AssignedCouncilRole)
	}
	if got := gs.Factions[owner].Gold; got != 500 {
		t.Fatalf("AI adoption left %d gold, want 500 after fee", got)
	}
}

func TestAINoCostRoleExemptsAdoptedCharacterUpkeep(t *testing.T) {
	const owner faction.FactionID = "ai"
	candidate := &army.Commander{
		ID: "chr_candidate", Type: army.CharacterTypeCharacter, OwnerID: state.NeutralCharacterOwnerID,
		AdoptionCostGold: 500, GoldUpkeep: 100,
		Traits: []army.CharacterTrait{"great_leader"},
	}
	gs := &state.GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			owner: {ID: owner, Gold: 500},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: string(owner), BaseGoldIncome: 1000, TaxRate: 100, Satisfaction: 100},
		},
		Commanders: map[string]*army.Commander{candidate.ID: candidate},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{{
				ID: state.CouncilRoleHeadOfHouse, NameTR: "Hane Lideri", NoCost: true,
				Effects: scenario.CouncilEffects{IncomeMod: 1.01},
			}},
			Traits: map[string]scenario.CharacterTraitDefinition{
				"great_leader": {
					NameTR: "Usta Lider",
					CouncilEffects: map[string]scenario.CouncilEffects{
						state.CouncilRoleHeadOfHouse: {IncomeMod: 1.10},
					},
				},
			},
		},
	}

	aiManageNeutralCharacters(gs, owner)

	if candidate.OwnerID != string(owner) || candidate.AssignedCouncilRole != state.CouncilRoleHeadOfHouse {
		t.Fatalf("AI failed to adopt profitable no-cost candidate: owner=%q role=%q", candidate.OwnerID, candidate.AssignedCouncilRole)
	}
}

func TestAIReservesUpkeepForOwnedCharactersExceptNoCostRoles(t *testing.T) {
	const owner faction.FactionID = "ai"
	character := &army.Commander{
		ID: "chr_member", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		AssignedCouncilRole: state.CouncilRoleHeadOfHouse, GoldUpkeep: 50,
	}
	gs := &state.GameState{
		Year:       300,
		Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner, Gold: 1000}},
		Commanders: map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{{
				ID: state.CouncilRoleHeadOfHouse, NameTR: "Hane Lideri", NoCost: true,
			}},
		},
	}

	exemptSnapshot := aiBuildEconomySnapshot(gs, owner)
	exemptBudget := prepareAIBudget(gs, owner, nil)
	if exemptSnapshot.GoldUpkeep != 0 {
		t.Fatalf("no-cost role upkeep = %d; want waived", exemptSnapshot.GoldUpkeep)
	}

	gs.CouncilConfig.Roles[0].NoCost = false
	regularMaintenanceSnapshot := aiBuildEconomySnapshot(gs, owner)
	regularMaintenanceBudget := prepareAIBudget(gs, owner, nil)
	if regularMaintenanceSnapshot.GoldUpkeep != 50 ||
		regularMaintenanceBudget.EmergencyGold <= exemptBudget.EmergencyGold {
		t.Fatalf("regular character upkeep = snapshot %d, emergency reserve %d; expected upkeep 50 and a larger reserve than %d",
			regularMaintenanceSnapshot.GoldUpkeep, regularMaintenanceBudget.EmergencyGold, exemptBudget.EmergencyGold)
	}
}

func TestAIRejectsCandidateWhenExpectedGainDoesNotCoverItsCosts(t *testing.T) {
	const owner faction.FactionID = "ai"
	candidate := &army.Commander{
		ID: "chr_candidate", Type: army.CharacterTypeCharacter, OwnerID: state.NeutralCharacterOwnerID,
		AdoptionCostGold: 500, GoldUpkeep: 1,
		Traits: []army.CharacterTrait{"modest_steward"},
	}
	gs := &state.GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			owner: {ID: owner, Gold: 1000},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: string(owner), BaseGoldIncome: 1000, TaxRate: 100, Satisfaction: 100},
		},
		Commanders: map[string]*army.Commander{candidate.ID: candidate},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{{
				ID: state.CouncilRoleTreasurer, NameTR: "Hazinedar",
				Effects: scenario.CouncilEffects{IncomeMod: 1.01},
			}},
			Traits: map[string]scenario.CharacterTraitDefinition{
				"modest_steward": {
					NameTR: "Mütevazı Yönetici",
					CouncilEffects: map[string]scenario.CouncilEffects{
						state.CouncilRoleTreasurer: {IncomeMod: 1.02},
					},
				},
			},
		},
	}

	aiManageNeutralCharacters(gs, owner)

	if candidate.OwnerID != state.NeutralCharacterOwnerID || candidate.AssignedCouncilRole != "" {
		t.Fatalf("AI adopted an unprofitable candidate: owner=%q role=%q", candidate.OwnerID, candidate.AssignedCouncilRole)
	}
	if got := gs.Factions[owner].Gold; got != 1000 {
		t.Fatalf("rejected candidate changed treasury to %d, want 1000", got)
	}
}

func TestAIDoesNotAdoptNeutralCharacterForRoleWithNoReturn(t *testing.T) {
	const owner faction.FactionID = "ai"
	candidate := &army.Commander{
		ID: "chr_candidate", Type: army.CharacterTypeCharacter, OwnerID: state.NeutralCharacterOwnerID,
	}
	gs := &state.GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			owner: {ID: owner},
		},
		Commanders: map[string]*army.Commander{candidate.ID: candidate},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{{ID: "empty_role", NameTR: "Getirisiz Görev"}},
		},
	}

	aiManageNeutralCharacters(gs, owner)

	if candidate.OwnerID != state.NeutralCharacterOwnerID || candidate.AssignedCouncilRole != "" {
		t.Fatalf("AI adopted a candidate for a role without return: owner %q, role %q", candidate.OwnerID, candidate.AssignedCouncilRole)
	}
}

func TestAILeavesNeutralCharactersForPlayerControlledFaction(t *testing.T) {
	character := &army.Commander{
		ID: "chr_a", Type: army.CharacterTypeCharacter, OwnerID: state.NeutralCharacterOwnerID, Name: "A",
	}
	gs := &state.GameState{
		Year:            300,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Commanders:    map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.DefaultCharacterCouncilConfig(),
	}

	aiManageNeutralCharacters(gs, faction.FactionID("player"))

	if character.OwnerID != state.NeutralCharacterOwnerID || character.AssignedCouncilRole != "" {
		t.Fatalf("AI changed player candidate: owner %q, role %q", character.OwnerID, character.AssignedCouncilRole)
	}
}

func TestAIFactionEliminationReleasesLivingHouseholdCharacters(t *testing.T) {
	character := &army.Commander{
		ID: "chr_survivor", Type: army.CharacterTypeCharacter, OwnerID: "fallen",
		AssignedCouncilRole: state.CouncilRoleTreasurer,
	}
	gs := &state.GameState{
		Commanders: map[string]*army.Commander{character.ID: character},
		Factions: map[faction.FactionID]*faction.Faction{
			"fallen": {ID: "fallen"},
		},
		Armies:      make(map[army.ArmyID]*army.Army),
		TradeRoutes: nil,
	}

	aiEliminateFactionAfterPrivilegeRevoke(gs, "fallen")

	if !gs.Factions["fallen"].IsEliminated {
		t.Fatal("faction was not eliminated")
	}
	if character.OwnerID != state.NeutralCharacterOwnerID || character.AssignedCouncilRole != "" {
		t.Fatalf("survivor state = owner %q, role %q; want neutral and unassigned", character.OwnerID, character.AssignedCouncilRole)
	}
	if candidates := gs.NeutralCharacterCandidates(); len(candidates) != 1 || candidates[0] != character {
		t.Fatalf("neutral candidates = %#v, want the eliminated faction's survivor", candidates)
	}
}
