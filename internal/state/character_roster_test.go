package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
)

func TestCharacterRosterIncludesHouseholdAndMilitaryCharactersSortedDeterministically(t *testing.T) {
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
	if len(roster) != 3 || roster[0].ID != "character" || roster[1].ID != "commander" || roster[2].ID != "administrator" {
		t.Fatalf("roster = %#v, want Arya, Jaime, then Tywin", roster)
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

func TestAdoptNeutralCharacterChargesFeeOnlyWhenAffordable(t *testing.T) {
	character := &army.Commander{
		ID: "chr_bronn", Type: army.CharacterTypeCharacter, OwnerID: NeutralCharacterOwnerID,
		AdoptionCostGold: 500, GoldUpkeep: 6,
	}
	owner := &faction.Faction{ID: "player", Gold: 499}
	gs := &GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": owner,
		},
		Commanders: map[string]*army.Commander{character.ID: character},
	}

	if gs.AdoptNeutralCharacterForFaction("player", character.ID) {
		t.Fatal("adoption succeeded without enough gold")
	}
	if owner.Gold != 499 || character.OwnerID != NeutralCharacterOwnerID {
		t.Fatalf("failed adoption changed gold or ownership: gold=%d owner=%q", owner.Gold, character.OwnerID)
	}
	owner.Gold++
	if !gs.AdoptNeutralCharacterForFaction("player", character.ID) {
		t.Fatal("adoption failed when the faction had exactly enough gold")
	}
	if owner.Gold != 0 || character.OwnerID != "player" {
		t.Fatalf("successful adoption state = gold %d, owner %q; want 0 and player", owner.Gold, character.OwnerID)
	}
}

func TestFactionCharacterGoldUpkeepIncludesOwnedHouseholdCharacters(t *testing.T) {
	leader := &army.Commander{
		ID: "leader", Type: army.CharacterTypeCharacter, OwnerID: "player",
		AssignedCouncilRole: CouncilRoleHeadOfHouse, GoldUpkeep: 50,
	}
	heir := &army.Commander{
		ID: "heir", Type: army.CharacterTypeCharacter, OwnerID: "player",
		AssignedCouncilRole: CouncilRoleHeir, GoldUpkeep: 60,
	}
	gs := &GameState{
		Year: 300,
		Commanders: map[string]*army.Commander{
			"owned":       {ID: "owned", Type: army.CharacterTypeCharacter, OwnerID: "player", GoldUpkeep: 6},
			leader.ID:     leader,
			heir.ID:       heir,
			"unavailable": {ID: "unavailable", Type: army.CharacterTypeAdministrator, OwnerID: "player", GoldUpkeep: 4, CaptiveAtSettlementID: "keep"},
			"neutral":     {ID: "neutral", Type: army.CharacterTypeCharacter, OwnerID: NeutralCharacterOwnerID, GoldUpkeep: 20},
			"commander":   {ID: "commander", Type: army.CharacterTypeCommander, OwnerID: "player", GoldUpkeep: 30},
			"dead":        {ID: "dead", Type: army.CharacterTypeCharacter, OwnerID: "player", GoldUpkeep: 10},
		},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{
				{ID: CouncilRoleHeadOfHouse, NameTR: "Lider", NoCost: true},
				{ID: CouncilRoleHeir, NameTR: "Veliaht", NoCost: true},
			},
		},
		DeadCommanderIDs: map[string]bool{"dead": true},
	}
	if got, want := gs.FactionCharacterGoldUpkeep("player"), 10; got != want {
		t.Fatalf("character upkeep = %d, want %d", got, want)
	}
}

func TestFactionCharacterGoldUpkeepIncludesCouncilAssignedCommanders(t *testing.T) {
	assigned := &army.Commander{
		ID: "assigned", Type: army.CharacterTypeCommander, OwnerID: "player",
		AssignedCouncilRole: CouncilRoleTreasurer, GoldUpkeep: 8,
	}
	unassigned := &army.Commander{
		ID: "unassigned", Type: army.CharacterTypeCommander, OwnerID: "player", GoldUpkeep: 12,
	}
	freeRoleCommander := &army.Commander{
		ID: "free_role", Type: army.CharacterTypeCommander, OwnerID: "player",
		AssignedCouncilRole: CouncilRoleHeadOfHouse, GoldUpkeep: 15,
	}
	gs := &GameState{
		Year: 300,
		Commanders: map[string]*army.Commander{
			assigned.ID: assigned, unassigned.ID: unassigned, freeRoleCommander.ID: freeRoleCommander,
		},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{
				{ID: CouncilRoleTreasurer, NameTR: "Hazinedar"},
				{ID: CouncilRoleHeadOfHouse, NameTR: "Lider", NoCost: true},
			},
		},
	}
	if got, want := gs.FactionCharacterGoldUpkeep("player"), 8; got != want {
		t.Fatalf("commander household upkeep = %d, want %d", got, want)
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

func TestMilitaryCommanderAppearsInHouseholdButCannotBeReleased(t *testing.T) {
	commander := &army.Commander{
		ID: "cmd_jaime", Type: army.CharacterTypeCommander, OwnerID: "player",
		Name: "Jaime", AssignedArmyID: "army_1",
	}
	gs := &GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Commanders: map[string]*army.Commander{commander.ID: commander},
	}

	roster := gs.CharacterRoster("player")
	if len(roster) != 1 || roster[0] != commander {
		t.Fatalf("household roster = %#v, want Jaime", roster)
	}
	if gs.ReleaseCharacterToNeutralPool("player", commander.ID) {
		t.Fatal("military commander was released through household character action")
	}
	if commander.OwnerID != "player" || commander.AssignedArmyID != "army_1" {
		t.Fatalf("rejected release changed commander state: owner=%q army=%q", commander.OwnerID, commander.AssignedArmyID)
	}
}

func TestNeutralCharacterCandidatesExcludeUnavailableAndArmyAssigned(t *testing.T) {
	gs := &GameState{
		Year: 300,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Commanders: map[string]*army.Commander{
			"ready":     {ID: "ready", Type: army.CharacterTypeCharacter, OwnerID: NeutralCharacterOwnerID, Name: "Ready"},
			"assigned":  {ID: "assigned", Type: army.CharacterTypeAdministrator, OwnerID: NeutralCharacterOwnerID, Name: "Assigned", AssignedArmyID: "army_1"},
			"exiled":    {ID: "exiled", Type: army.CharacterTypeCharacter, OwnerID: NeutralCharacterOwnerID, Name: "Exiled"},
			"commander": {ID: "commander", Type: army.CharacterTypeCommander, OwnerID: NeutralCharacterOwnerID, Name: "Commander"},
		},
		ExiledCommanderIDs: map[string]bool{"exiled": true},
	}

	candidates := gs.NeutralCharacterCandidates()
	if len(candidates) != 1 || candidates[0].ID != "ready" {
		t.Fatalf("neutral candidates = %#v, want only Ready", candidates)
	}
	if gs.AdoptNeutralCharacterForFaction("player", "commander") {
		t.Fatal("military commander was adopted through household candidate flow")
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
