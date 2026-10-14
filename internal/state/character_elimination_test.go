package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
)

func TestReleaseFactionCharactersToNeutralPoolKeepsLivingCharactersAvailable(t *testing.T) {
	assigned := &army.Commander{
		ID: "chr_assigned", Type: army.CharacterTypeAdministrator, OwnerID: "fallen",
		AssignedArmyID: "army_1", AssignedCouncilRole: CouncilRoleTreasurer,
	}
	captive := &army.Commander{
		ID: "chr_captive", Type: army.CharacterTypeCharacter, OwnerID: "fallen",
		CaptiveAtSettlementID: "settlement_1",
	}
	dead := &army.Commander{ID: "chr_dead", Type: army.CharacterTypeCharacter, OwnerID: "fallen"}
	expired := &army.Commander{ID: "chr_expired", Type: army.CharacterTypeCharacter, OwnerID: "fallen", EndYear: 300}
	commander := &army.Commander{ID: "commander", Type: army.CharacterTypeCommander, OwnerID: "fallen"}
	gs := &GameState{
		Year: 300,
		Commanders: map[string]*army.Commander{
			assigned.ID:  assigned,
			captive.ID:   captive,
			dead.ID:      dead,
			expired.ID:   expired,
			commander.ID: commander,
		},
		DeadCommanderIDs: map[string]bool{dead.ID: true},
		Armies: map[army.ArmyID]*army.Army{
			"army_1": {ID: "army_1", Commander: assigned},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"fallen": {ID: "fallen"},
		},
	}

	if got := gs.ReleaseFactionCharactersToNeutralPool("fallen"); got != 2 {
		t.Fatalf("released characters = %d, want 2", got)
	}
	if assigned.OwnerID != NeutralCharacterOwnerID || assigned.AssignedArmyID != "" || assigned.AssignedCouncilRole != "" {
		t.Fatalf("assigned survivor state = owner %q, army %q, role %q", assigned.OwnerID, assigned.AssignedArmyID, assigned.AssignedCouncilRole)
	}
	if gs.Armies["army_1"].Commander != nil {
		t.Fatal("surviving character remained linked to the eliminated faction's army")
	}
	if captive.OwnerID != NeutralCharacterOwnerID || captive.CaptiveAtSettlementID == "" {
		t.Fatalf("captive survivor state = owner %q, captive at %q", captive.OwnerID, captive.CaptiveAtSettlementID)
	}
	for _, character := range []*army.Commander{dead, expired, commander} {
		if character.OwnerID != "fallen" {
			t.Errorf("ineligible record %q changed owner to %q", character.ID, character.OwnerID)
		}
	}
	if candidates := gs.NeutralCharacterCandidates(); len(candidates) != 1 || candidates[0] != assigned {
		t.Fatalf("available neutral candidates = %#v, want only the non-captive survivor", candidates)
	}
}
