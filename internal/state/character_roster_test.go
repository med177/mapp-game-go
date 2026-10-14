package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
)

func TestCharacterRosterFiltersTypesAndSortsDeterministically(t *testing.T) {
	gs := &GameState{
		Year: 300,
		Commanders: map[string]*army.Commander{
			"administrator": {ID: "administrator", Type: army.CharacterTypeAdministrator, OwnerID: "player", Name: "Tywin"},
			"character":     {ID: "character", Type: army.CharacterTypeCharacter, OwnerID: "player", Name: "Arya"},
			"commander":     {ID: "commander", Type: army.CharacterTypeCommander, OwnerID: "player", Name: "Jaime"},
			"future":        {ID: "future", Type: army.CharacterTypeCharacter, OwnerID: "player", Name: "Future", StartYear: 301},
		},
	}

	roster := gs.CharacterRoster("player")
	if len(roster) != 2 || roster[0].ID != "character" || roster[1].ID != "administrator" {
		t.Fatalf("roster = %#v, want Arya then Tywin", roster)
	}
	if got := gs.CharacterRoster("missing"); len(got) != 0 {
		t.Fatalf("missing faction roster = %#v, want empty", got)
	}
}

func TestCharacterRosterSortsByAssignedDutyBeforeName(t *testing.T) {
	gs := &GameState{
		Year: 300,
		Commanders: map[string]*army.Commander{
			"unassigned": {ID: "unassigned", Type: army.CharacterTypeCharacter, OwnerID: "player", Name: "A"},
			"diplomat":   {ID: "diplomat", Type: army.CharacterTypeCharacter, OwnerID: "player", Name: "B", AssignedCouncilRole: CouncilRoleDiplomat},
			"treasurer":  {ID: "treasurer", Type: army.CharacterTypeCharacter, OwnerID: "player", Name: "C", AssignedCouncilRole: CouncilRoleTreasurer},
			"head":       {ID: "head", Type: army.CharacterTypeCharacter, OwnerID: "player", Name: "D", AssignedCouncilRole: CouncilRoleHeadOfHouse},
			"heir":       {ID: "heir", Type: army.CharacterTypeCharacter, OwnerID: "player", Name: "E", AssignedCouncilRole: CouncilRoleHeir},
			"military":   {ID: "military", Type: army.CharacterTypeCharacter, OwnerID: "player", Name: "F", AssignedCouncilRole: CouncilRoleMilitaryResponsible},
			"army":       {ID: "army", Type: army.CharacterTypeAdministrator, OwnerID: "player", Name: "E", AssignedArmyID: "army_1"},
		},
	}

	roster := gs.CharacterRoster("player")
	want := []string{"head", "treasurer", "diplomat", "heir", "military", "army", "unassigned"}
	if len(roster) != len(want) {
		t.Fatalf("roster length = %d, want %d", len(roster), len(want))
	}
	for i, id := range want {
		if roster[i].ID != id {
			t.Errorf("roster[%d] = %q, want %q", i, roster[i].ID, id)
		}
	}
}

func TestAdoptNeutralCharacterMovesOwnershipAndRemovesCandidate(t *testing.T) {
	character := &army.Commander{
		ID: "chr_bronn", Type: army.CharacterTypeCharacter, OwnerID: NeutralCharacterOwnerID,
		Name: "Bronn", Traits: []army.CharacterTrait{"mercenary"},
	}
	gs := &GameState{
		Year: 300, PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Commanders: map[string]*army.Commander{character.ID: character},
	}

	if candidates := gs.NeutralCharacterCandidates(); len(candidates) != 1 || candidates[0] != character {
		t.Fatalf("neutral candidates = %#v, want Bronn", candidates)
	}
	if !gs.AdoptNeutralCharacter(character.ID) {
		t.Fatal("AdoptNeutralCharacter() rejected an eligible neutral character")
	}
	if character.OwnerID != "player" || len(gs.NeutralCharacterCandidates()) != 0 || len(gs.CharacterRoster("player")) != 1 {
		t.Fatalf("adopted character owner/roster mismatch: owner=%q", character.OwnerID)
	}
	if len(character.Traits) != 1 || character.Traits[0] != "mercenary" {
		t.Fatalf("adopting character changed traits: %v", character.Traits)
	}
	if gs.AdoptNeutralCharacter(character.ID) {
		t.Fatal("AdoptNeutralCharacter() adopted the character a second time")
	}
}

func TestReleaseCharacterReturnsToNeutralAndClearsAssignments(t *testing.T) {
	character := &army.Commander{
		ID: "chr_bronn", Type: army.CharacterTypeCharacter, OwnerID: "player", Name: "Bronn",
		AssignedArmyID: "army_land", AssignedCouncilRole: CouncilRoleTreasurer,
	}
	embarkedCharacter := &army.Commander{
		ID: "chr_shae", Type: army.CharacterTypeCharacter, OwnerID: "player", Name: "Shae",
		AssignedArmyID: "army_fleet",
	}
	gs := &GameState{
		Year: 300, PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Commanders: map[string]*army.Commander{
			character.ID:         character,
			embarkedCharacter.ID: embarkedCharacter,
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_land":  {ID: "army_land", Commander: character},
			"army_fleet": {ID: "army_fleet", EmbarkedCommander: embarkedCharacter},
		},
	}

	if !gs.ReleaseCharacterToNeutralPool("player", character.ID) {
		t.Fatal("ReleaseCharacterToNeutralPool() rejected an owned active character")
	}
	if character.OwnerID != NeutralCharacterOwnerID || character.AssignedArmyID != "" || character.AssignedCouncilRole != "" {
		t.Fatalf("released character state = owner %q, army %q, council %q", character.OwnerID, character.AssignedArmyID, character.AssignedCouncilRole)
	}
	if gs.Armies["army_land"].Commander != nil {
		t.Fatal("released land commander still linked to army")
	}
	if len(gs.NeutralCharacterCandidates()) != 1 || gs.NeutralCharacterCandidates()[0] != character {
		t.Fatalf("neutral candidates = %#v, want released Bronn", gs.NeutralCharacterCandidates())
	}
	if !gs.ReleaseCharacterToNeutralPool("player", embarkedCharacter.ID) || gs.Armies["army_fleet"].EmbarkedCommander != nil {
		t.Fatal("release did not clear the embarked commander assignment")
	}
}

func TestReleaseCharacterRejectsNonOwner(t *testing.T) {
	character := &army.Commander{
		ID: "chr_bronn", Type: army.CharacterTypeCharacter, OwnerID: "other", Name: "Bronn",
	}
	gs := &GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"other":  {ID: "other"},
		},
		Commanders: map[string]*army.Commander{character.ID: character},
	}
	if gs.ReleaseCharacterToNeutralPool("player", character.ID) {
		t.Fatal("release succeeded for a character owned by another faction")
	}
	if character.OwnerID != "other" {
		t.Fatalf("rejected release changed owner to %q", character.OwnerID)
	}
}

func TestNeutralCharacterCandidatesExcludeUnavailableAndArmyAssigned(t *testing.T) {
	gs := &GameState{
		Year: 300,
		Commanders: map[string]*army.Commander{
			"ready":    {ID: "ready", Type: army.CharacterTypeCharacter, OwnerID: NeutralCharacterOwnerID, Name: "Ready"},
			"assigned": {ID: "assigned", Type: army.CharacterTypeAdministrator, OwnerID: NeutralCharacterOwnerID, Name: "Assigned", AssignedArmyID: "army_1"},
			"exiled":   {ID: "exiled", Type: army.CharacterTypeCharacter, OwnerID: NeutralCharacterOwnerID, Name: "Exiled"},
		},
		ExiledCommanderIDs: map[string]bool{"exiled": true},
	}

	candidates := gs.NeutralCharacterCandidates()
	if len(candidates) != 1 || candidates[0].ID != "ready" {
		t.Fatalf("neutral candidates = %#v, want only Ready", candidates)
	}
}

func TestCharacterTemplatesDoNotCreateCommanderArrivalNotices(t *testing.T) {
	administrator := &army.Commander{
		ID: "administrator", Type: army.CharacterTypeAdministrator, OwnerID: "player", Name: "Yönetici",
	}
	character := &army.Commander{
		ID: "character", Type: army.CharacterTypeCharacter, OwnerID: "player", Name: "Karakter",
	}
	gs := &GameState{
		Year:               300,
		PlayerFactionID:    "player",
		CommanderTemplates: map[string][]*army.Commander{"player": {administrator, character}},
	}

	arrivals := gs.SyncCommanderAvailability()
	if len(arrivals) != 1 || arrivals[0].ID != administrator.ID {
		t.Fatalf("commander arrivals = %#v, want only administrator", arrivals)
	}
	if gs.Commanders[character.ID] == nil || gs.CommanderArrivalNotices[character.ID] {
		t.Fatal("character was not loaded independently of military arrival notices")
	}
}
