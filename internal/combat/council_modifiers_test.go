package combat

import (
	"math"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
)

func TestCouncilMilitaryPowerAppliesFactionAndCommanderBonuses(t *testing.T) {
	const owner faction.FactionID = "player"
	commander := &army.Commander{
		ID: "chr_marshal", Type: army.CharacterTypeCharacter, OwnerID: string(owner),
	}
	gs := &state.GameState{
		Year:          1,
		Factions:      map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Commanders:    map[string]*army.Commander{commander.ID: commander},
		CouncilConfig: scenario.DefaultCharacterCouncilConfig(),
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Attack: 100},
		},
	}
	if !gs.AssignCharacterToCouncilRole(owner, commander.ID, state.CouncilRoleMilitaryResponsible) {
		t.Fatal("could not assign military responsible")
	}

	mods := TechModsFor(gs, string(owner))
	if math.Abs(mods.CouncilPowerMod-0.02) > 1e-9 {
		t.Fatalf("faction-wide council modifier without military trait = %v, want 0.02", mods.CouncilPowerMod)
	}
	if math.Abs(mods.CommanderPowerMod-0.03) > 1e-9 {
		t.Fatalf("commander council modifier = %v, want 0.03", mods.CommanderPowerMod)
	}

	commandedArmy := &army.Army{
		OwnerID: string(owner), Commander: commander,
		Units: []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}},
	}
	ordinaryArmy := &army.Army{
		OwnerID: string(owner),
		Units:   []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}},
	}
	commanded := ArmyStrengthBreakdown(commandedArmy, "", gs.UnitTypes, mods, BattleContextLand, BattleStanceBalanced, true, false)
	ordinary := ArmyStrengthBreakdown(ordinaryArmy, "", gs.UnitTypes, mods, BattleContextLand, BattleStanceBalanced, true, false)
	if math.Abs(commanded.FinalStrength-ordinary.FinalStrength*(1.05/1.02)) > 1e-6 {
		t.Fatalf("commanded army strength = %v, ordinary = %v; expected commander-only bonus", commanded.FinalStrength, ordinary.FinalStrength)
	}
	if math.Abs(commanded.CouncilPowerModifier-0.02) > 1e-9 ||
		math.Abs(commanded.CommanderCouncilModifier-0.03) > 1e-9 ||
		ordinary.CommanderCouncilModifier != 0 {
		t.Fatalf("council breakdown faction=%v commanded=%v ordinary=%v",
			commanded.CouncilPowerModifier, commanded.CommanderCouncilModifier, ordinary.CommanderCouncilModifier)
	}
}
