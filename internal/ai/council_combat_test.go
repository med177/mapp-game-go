package ai

import (
	"math"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
)

func TestAICombatModifiersIncludeCouncilMilitaryBonusesWithoutTechs(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_marshal", Type: army.CharacterTypeAdministrator, OwnerID: string(owner),
	}
	gs := &state.GameState{
		Year:          1,
		Factions:      map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Commanders:    map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.DefaultCharacterCouncilConfig(),
	}
	if !gs.AssignCharacterToCouncilRole(owner, character.ID, state.CouncilRoleMilitaryResponsible) {
		t.Fatal("could not assign military responsible")
	}

	mods := aiTechMods(gs, string(owner))
	if math.Abs(mods.CouncilPowerMod-0.02) > 1e-9 || math.Abs(mods.CommanderPowerMod-0.03) > 1e-9 {
		t.Fatalf("AI council modifiers = (%v, %v), want (0.02, 0.03)", mods.CouncilPowerMod, mods.CommanderPowerMod)
	}
}
