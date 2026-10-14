package state

import (
	"math"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/world"
)

func TestScenarioCouncilRolesTraitsAndExperienceScaleEffects(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_scholar", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		Name: "Scholar", Traits: []army.CharacterTrait{"scholar"},
	}
	gs := &GameState{
		Year:       300,
		Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Commanders: map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{
			XPPerTurn:   5,
			XPPerEffect: 10,
			Roles: []scenario.CouncilRoleDefinition{{
				ID: "steward", NameTR: "Kâhya",
				Effects:  scenario.CouncilEffects{IncomeMod: 1.03},
				PerLevel: scenario.CouncilEffects{IncomeMod: 1.02},
			}},
			Traits: map[string]scenario.CharacterTraitDefinition{
				"scholar": {
					NameTR: "Bilgin",
					CouncilEffects: map[string]scenario.CouncilEffects{
						"steward": {IncomeMod: 1.04},
					},
				},
			},
		},
	}
	if !gs.AssignCharacterToCouncilRole(owner, character.ID, "steward") {
		t.Fatal("scenario-defined council role could not be assigned")
	}
	if got := gs.CouncilIncomeBonus(owner, 1000); got != 70 {
		t.Fatalf("scenario-defined steward income effect = %d, want 70", got)
	}
	if got := gs.CouncilRoleEffects(owner, "steward").IncomeMod; math.Abs(got-1.07) > 1e-9 {
		t.Fatalf("level-one trait-adjusted effect = %.4f, want 1.07", got)
	}
	if got, want := gs.CouncilRoleEffectsForCharacter(character, "steward"), gs.CouncilRoleEffects(owner, "steward"); got != want {
		t.Fatalf("candidate role effects = %+v, assigned role effects = %+v", got, want)
	}
	gs.RecordCouncilTurnExperience()
	gs.RecordCouncilEffectExperience(owner, "income_mod")
	if character.CouncilExperience != 15 || character.CouncilLevel != 1 {
		t.Fatalf("council progress after turn and effect = %d XP, level %d", character.CouncilExperience, character.CouncilLevel)
	}
	character.RecordCouncilExperience(85)
	if character.CouncilLevel != 2 {
		t.Fatalf("council level after reaching commander level-two threshold = %d, want 2", character.CouncilLevel)
	}
	if got := gs.CouncilRoleEffects(owner, "steward").IncomeMod; math.Abs(got-1.09) > 1e-9 {
		t.Fatalf("level-two trait-adjusted effect = %.4f, want 1.09", got)
	}
}

func TestNegativeTraitCouncilModReducesRoleBonus(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_inexperienced", Type: army.CharacterTypeCharacter,
		OwnerID: string(owner), Name: "Inexperienced", Traits: []army.CharacterTrait{"inexperienced"},
	}
	gs := &GameState{
		Year:       300,
		Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Commanders: map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{{
				ID: CouncilRoleMilitaryResponsible, NameTR: "Askeri Sorumlu",
				Effects: scenario.CouncilEffects{MilitaryPowerMod: 1.03},
			}},
			Traits: map[string]scenario.CharacterTraitDefinition{
				"inexperienced": {
					NameTR: "Tecrübesiz",
					CouncilEffects: map[string]scenario.CouncilEffects{
						CouncilRoleMilitaryResponsible: {MilitaryPowerMod: 0.98},
					},
				},
			},
		},
	}
	if !gs.AssignCharacterToCouncilRole(owner, character.ID, CouncilRoleMilitaryResponsible) {
		t.Fatal("military responsible assignment failed")
	}
	if got := gs.CouncilMilitaryPowerMod(owner); math.Abs(got-1.01) > 1e-9 {
		t.Fatalf("military power mod after negative trait effect = %.4f, want 1.01", got)
	}
}

func TestMilitaryRoleFallbackPenaltyAppliesOnlyWithoutPositiveMilitaryTrait(t *testing.T) {
	const owner faction.FactionID = "player"
	buildState := func(character *army.Commander) *GameState {
		config := scenario.DefaultCharacterCouncilConfig()
		config.Traits = map[string]scenario.CharacterTraitDefinition{
			"fighter": {
				NameTR: "Savaşçı",
				CouncilEffects: map[string]scenario.CouncilEffects{
					CouncilRoleMilitaryResponsible: {MilitaryPowerMod: 1.01},
				},
			},
			"scholar": {
				NameTR: "Bilgin",
				CouncilEffects: map[string]scenario.CouncilEffects{
					"treasurer": {IncomeMod: 1.01},
				},
			},
		}
		return &GameState{
			Year:          300,
			Factions:      map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
			Commanders:    map[string]*army.Commander{character.ID: character},
			CouncilConfig: config,
		}
	}

	unfit := &army.Commander{
		ID: "chr_unfit", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		Traits: []army.CharacterTrait{"scholar", "scholar"},
	}
	unfitState := buildState(unfit)
	if !unfitState.AssignCharacterToCouncilRole(owner, unfit.ID, CouncilRoleMilitaryResponsible) {
		t.Fatal("could not assign character without military traits")
	}
	if got := unfitState.CouncilMilitaryPowerMod(owner); math.Abs(got-1.02) > 1e-9 {
		t.Fatalf("unmatched-trait military mod = %.4f, want 1.02 (single fallback penalty)", got)
	}

	fit := &army.Commander{
		ID: "chr_fit", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
		Traits: []army.CharacterTrait{"fighter", "scholar"},
	}
	fitState := buildState(fit)
	if !fitState.AssignCharacterToCouncilRole(owner, fit.ID, CouncilRoleMilitaryResponsible) {
		t.Fatal("could not assign character with military trait")
	}
	if got := fitState.CouncilMilitaryPowerMod(owner); math.Abs(got-1.04) > 1e-9 {
		t.Fatalf("matched-trait military mod = %.4f, want 1.04 without fallback penalty", got)
	}
}

func TestTreasurerCouncilSeatAddsFactionIncomeBonusAndBlocksArmyAssignment(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "cmd_tywin", Type: army.CharacterTypeAdministrator,
		OwnerID: string(owner), Name: "Tywin Lannister",
	}
	currentArmy := &army.Army{ID: "army_player", OwnerID: string(owner)}
	gs := &GameState{
		Year:            300,
		PlayerFactionID: owner,
		Factions:        map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Commanders:      map[string]*army.Commander{character.ID: character},
		Armies:          map[army.ArmyID]*army.Army{currentArmy.ID: currentArmy},
	}

	if got := gs.CouncilIncomeBonus(owner, 1000); got != 0 {
		t.Fatalf("unassigned council income bonus = %d, want 0", got)
	}
	if !gs.AssignCharacterToCouncilRole(owner, character.ID, CouncilRoleTreasurer) {
		t.Fatal("eligible character could not be assigned as treasurer")
	}
	if got := gs.CouncilIncomeBonus(owner, 1000); got != 50 {
		t.Fatalf("treasurer income bonus = %d, want 50", got)
	}
	if gs.AssignCommanderToArmy(character.ID, currentArmy.ID) {
		t.Fatal("council character was also assigned to an army")
	}
	if len(gs.CommanderRoster(string(owner))) != 0 {
		t.Fatal("council character remained in the available commander roster")
	}
	if !gs.UnassignCharacterFromCouncilRole(owner, character.ID) || character.AssignedCouncilRole != "" {
		t.Fatal("treasurer could not be returned to the character pool")
	}
	if !gs.AssignCommanderToArmy(character.ID, currentArmy.ID) {
		t.Fatal("released council character could not be assigned to an army")
	}
}

func TestAssigningTreasurerReplacesIncumbentAndRejectsDoubleRole(t *testing.T) {
	const owner faction.FactionID = "player"
	first := &army.Commander{ID: "chr_first", Type: army.CharacterTypeAdministrator, OwnerID: string(owner), Name: "First"}
	second := &army.Commander{ID: "chr_second", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Second"}
	gs := &GameState{
		Year:       300,
		Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Commanders: map[string]*army.Commander{first.ID: first, second.ID: second},
	}

	if !gs.AssignCharacterToCouncilRole(owner, first.ID, CouncilRoleTreasurer) {
		t.Fatal("first treasurer assignment failed")
	}
	if !gs.AssignCharacterToCouncilRole(owner, second.ID, CouncilRoleTreasurer) {
		t.Fatal("replacement treasurer assignment failed")
	}
	if first.AssignedCouncilRole != "" || second.AssignedCouncilRole != CouncilRoleTreasurer ||
		gs.CouncilRoleHolder(owner, CouncilRoleTreasurer) != second {
		t.Fatalf("replacement did not produce one incumbent: first=%q second=%q", first.AssignedCouncilRole, second.AssignedCouncilRole)
	}
	if gs.AssignCharacterToCouncilRole("missing", first.ID, CouncilRoleTreasurer) {
		t.Fatal("character was assigned to a nonexistent faction")
	}
	if gs.CouncilIncomeBonus(owner, 1000) != 50 {
		t.Fatal("replacing treasurer stacked the income bonus")
	}
}

func TestCharacterCanBeReassignedBetweenCouncilDuties(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_sansa", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Sansa",
	}
	previousHead := &army.Commander{
		ID: "chr_arya", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Arya",
		AssignedCouncilRole: CouncilRoleHeadOfHouse,
	}
	gs := &GameState{
		Year:       300,
		Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Commanders: map[string]*army.Commander{character.ID: character, previousHead.ID: previousHead},
	}

	if !gs.CanAssignCharacterToCouncilRole(owner, previousHead.ID, CouncilRoleDiplomat) {
		t.Fatal("character with an existing council duty could not be reassigned")
	}
	if !gs.AssignCharacterToCouncilRole(owner, previousHead.ID, CouncilRoleDiplomat) {
		t.Fatal("reassigning a character to diplomat failed")
	}
	if previousHead.AssignedCouncilRole != CouncilRoleDiplomat ||
		gs.CouncilRoleHolder(owner, CouncilRoleHeadOfHouse) != nil ||
		gs.CouncilRoleHolder(owner, CouncilRoleDiplomat) != previousHead {
		t.Fatalf("reassignment did not move the character to diplomat: role=%q", previousHead.AssignedCouncilRole)
	}
	if !gs.AssignCharacterToCouncilRole(owner, character.ID, CouncilRoleHeadOfHouse) {
		t.Fatal("assigning a new head of house failed")
	}
	if gs.CouncilRoleHolder(owner, CouncilRoleHeadOfHouse) != character ||
		previousHead.AssignedCouncilRole != CouncilRoleDiplomat {
		t.Fatal("assigning head of house incorrectly changed the diplomat")
	}
}

func TestNewCouncilRolesCanBeAssignedAndReplaceTheirIncumbents(t *testing.T) {
	const owner faction.FactionID = "player"
	roles := []string{CouncilRoleHeir, CouncilRoleMilitaryResponsible}
	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			first := &army.Commander{ID: "first", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "First"}
			second := &army.Commander{ID: "second", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Second"}
			gs := &GameState{
				Year:       300,
				Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
				Commanders: map[string]*army.Commander{first.ID: first, second.ID: second},
			}
			if !gs.CanAssignCharacterToCouncilRole(owner, first.ID, role) ||
				!gs.AssignCharacterToCouncilRole(owner, first.ID, role) {
				t.Fatalf("could not assign first character to %q", role)
			}
			if gs.CouncilRoleHolder(owner, role) != first {
				t.Fatalf("role holder for %q = %v, want first character", role, gs.CouncilRoleHolder(owner, role))
			}
			if !gs.AssignCharacterToCouncilRole(owner, second.ID, role) {
				t.Fatalf("could not replace incumbent for %q", role)
			}
			if first.AssignedCouncilRole != "" || gs.CouncilRoleHolder(owner, role) != second {
				t.Fatalf("replacement for %q left incorrect incumbents: first=%q holder=%v", role, first.AssignedCouncilRole, gs.CouncilRoleHolder(owner, role))
			}
			if !gs.UnassignCharacterFromCouncilRole(owner, second.ID) || second.AssignedCouncilRole != "" {
				t.Fatalf("could not unassign %q", role)
			}
		})
	}
}

func TestTaxIncomeForFactionRemainsSeparateFromCouncilIncomeBonus(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{ID: "chr_sansa", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Sansa"}
	region := &world.Region{ID: "capital", OwnerID: string(owner), BaseGoldIncome: 1000, TaxRate: 50, Satisfaction: 50}
	gs := &GameState{
		Year:     300,
		Factions: map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Regions:  map[world.RegionID]*world.Region{region.ID: region},
		Commanders: map[string]*army.Commander{
			character.ID: character,
		},
	}

	if got := gs.TaxIncomeForFaction(owner); got != 500 {
		t.Fatalf("tax income without treasurer = %d, want 500", got)
	}
	if !gs.AssignCharacterToCouncilRole(owner, character.ID, CouncilRoleTreasurer) {
		t.Fatal("treasurer assignment failed")
	}
	if got := gs.TaxIncomeForFaction(owner); got != 500 {
		t.Fatalf("council income bonus changed tax-only budget = %d, want 500", got)
	}
}

func TestDiplomatCouncilSeatAddsBonusToHeyetOnly(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_varys", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Varys",
	}
	armyUnit := &army.Army{ID: "army_player", OwnerID: string(owner)}
	gs := &GameState{
		Year:            300,
		PlayerFactionID: owner,
		Factions:        map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Commanders:      map[string]*army.Commander{character.ID: character},
		Armies:          map[army.ArmyID]*army.Army{armyUnit.ID: armyUnit},
	}

	if !gs.AssignCharacterToCouncilRole(owner, character.ID, CouncilRoleDiplomat) {
		t.Fatal("diplomat assignment failed")
	}
	if got := gs.DiplomatRelationBonus(owner); got != 3 {
		t.Fatalf("diplomat bonus = %d, want 3", got)
	}
	if gs.AssignCommanderToArmy(character.ID, armyUnit.ID) {
		t.Fatal("diplomat was also assigned to an army")
	}
	if !gs.UnassignCharacterFromCouncilRole(owner, character.ID) || gs.DiplomatRelationBonus(owner) != 0 {
		t.Fatal("unassigned diplomat still contributed a bonus")
	}
}
