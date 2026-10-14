package diplomacy

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
)

func diplomatHeyetTestState() (*state.GameState, *army.Commander) {
	const actor faction.FactionID = "player"
	const target faction.FactionID = "target"
	diplomat := &army.Commander{
		ID: "chr_varys", Type: army.CharacterTypeCharacter,
		OwnerID: string(actor), Name: "Varys",
	}
	gs := &state.GameState{
		Turn:            1,
		Year:            300,
		PlayerFactionID: actor,
		DiplomacyConfig: scenario.DiplomacyConfig{
			RelationImprovementGoldCost: 100,
			RelationImprovementBonus:    5,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			actor:  {ID: actor, Gold: 500},
			target: {ID: target, Gold: 200},
		},
		Commanders: map[string]*army.Commander{diplomat.ID: diplomat},
	}
	if !gs.AssignCharacterToCouncilRole(actor, diplomat.ID, state.CouncilRoleDiplomat) {
		panic("diplomat fixture assignment failed")
	}
	return gs, diplomat
}

func TestExecuteImproveRelationsAddsAssignedDiplomatBonus(t *testing.T) {
	gs, _ := diplomatHeyetTestState()
	result := Execute(gs, gs.PlayerFactionID, "target", ActionImproveRelations)
	if !result.Applied {
		t.Fatalf("heyet was not applied: %s", result.Message)
	}
	if got := RelationScore(gs, gs.PlayerFactionID, "target"); got != 8 {
		t.Fatalf("heyet relation score = %d, want base 5 + diplomat 3 = 8", got)
	}
	if got := gs.Factions[gs.PlayerFactionID].Gold; got != 400 {
		t.Fatalf("heyet gold after cost = %d, want 400", got)
	}
}

func TestResolveImproveRelationsNotificationAddsOriginDiplomatBonus(t *testing.T) {
	gs, _ := diplomatHeyetTestState()
	gs.DiplomaticOffers = []state.DiplomaticOffer{{
		FromFactionID: gs.PlayerFactionID,
		ToFactionID:   "target",
		Action:        string(ActionImproveRelations),
		CreatedTurn:   gs.Turn,
	}}
	result := ResolveOffer(gs, 0, true)
	if !result.Applied {
		t.Fatalf("queued heyet notification was not applied: %s", result.Message)
	}
	if got := RelationScore(gs, gs.PlayerFactionID, "target"); got != 8 {
		t.Fatalf("queued heyet relation score = %d, want base 5 + diplomat 3 = 8", got)
	}
}
