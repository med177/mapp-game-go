package game

import (
	"math"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

func TestTreasurerBonusAppliesOnlyToNonSiegeRegionalTax(t *testing.T) {
	const fid faction.FactionID = "player"

	buildState := func(withTreasurer bool) *state.GameState {
		character := &army.Commander{
			ID: "cmd_tywin", Type: army.CharacterTypeAdministrator,
			OwnerID: string(fid), Name: "Tywin",
		}

		gs := &state.GameState{
			Turn:            1,
			Year:            300,
			Month:           6,
			PlayerFactionID: fid,
			Factions: map[faction.FactionID]*faction.Faction{
				fid: {
					ID: fid, Gold: 0, Grain: 100,
					OtherIncomePeriods: []faction.OtherIncomePeriod{{StartYear: 300, Amount: 100}},
				},
			},
			Regions: map[world.RegionID]*world.Region{
				"capital":  {ID: "capital", OwnerID: string(fid), BaseGoldIncome: 1000, TaxRate: 50, TradeCapacity: 100, Satisfaction: 50},
				"besieged": {ID: "besieged", OwnerID: string(fid), BaseGoldIncome: 2000, TaxRate: 100, Satisfaction: 50},
			},
			Commanders: map[string]*army.Commander{character.ID: character},
			Sieges:     map[world.RegionID]*state.SiegeState{"besieged": {RegionID: "besieged"}},
		}
		if withTreasurer && !gs.AssignCharacterToCouncilRole(fid, character.ID, state.CouncilRoleTreasurer) {
			t.Fatal("treasurer assignment failed")
		}
		return gs
	}

	base := buildState(false)
	withTreasurer := buildState(true)
	treasurer := withTreasurer.Commanders["cmd_tywin"]
	baseReport := applyEconomyTick(base)
	treasurerReport := applyEconomyTick(withTreasurer)
	_ = baseReport
	_ = treasurerReport

	baseStatus := base.GoldEconomy[fid]
	treasurerStatus := withTreasurer.GoldEconomy[fid]
	wantBonus := baseStatus.TaxIncome * 5 / 100
	if wantBonus <= 0 {
		t.Fatalf("fixture produced no tax income bonus: tax income = %d", baseStatus.TaxIncome)
	}
	if got, want := treasurerStatus.CouncilIncome, wantBonus; got != want {
		t.Fatalf("economy tick council income bonus with treasurer = %d, want %d", got, want)
	}
	if got, want := treasurerStatus.TaxIncome, baseStatus.TaxIncome; got != want {
		t.Fatalf("treasurer changed tax income category: got %d want %d", got, want)
	}
	if got, want := treasurerStatus.Income, baseStatus.Income; got != want {
		t.Fatalf("economy tick income with treasurer = %d, want %d", got, want)
	}
	if got, want := treasurerStatus.TradeIncome, baseStatus.TradeIncome; got != want {
		t.Fatalf("treasurer changed trade income: got %d want %d", got, want)
	}
	if got, want := treasurerStatus.OtherIncome, baseStatus.OtherIncome; got != want {
		t.Fatalf("treasurer changed other income category: got %d want %d", got, want)
	}
	if treasurerStatus.TradeIncome == 0 || treasurerStatus.OtherIncome == 0 {
		t.Fatalf("fixture must include trade and other income, got trade=%d other=%d",
			treasurerStatus.TradeIncome, treasurerStatus.OtherIncome)
	}
	if treasurer.CouncilExperience != 10 {
		t.Fatalf("treasurer experience after income bonus = %d, want 10", treasurer.CouncilExperience)
	}
}

func TestCharacterUpkeepIsPaidBeforeArmyUpkeepWithoutChangingArmyShortage(t *testing.T) {
	const fid faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_member", Type: army.CharacterTypeCharacter, OwnerID: string(fid), GoldUpkeep: 10,
	}
	gs := &state.GameState{
		Turn:            1,
		Year:            300,
		Month:           1,
		PlayerFactionID: fid,
		Factions: map[faction.FactionID]*faction.Faction{
			fid: {ID: fid, Gold: 12, Grain: 100},
		},
		Commanders: map[string]*army.Commander{character.ID: character},
		Armies: map[army.ArmyID]*army.Army{
			"army": {ID: "army", OwnerID: string(fid), Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", GoldUpkeep: 5},
		},
	}

	applyEconomyTick(gs)

	status := gs.GoldEconomy[fid]
	if status.CharacterUpkeep != 10 || status.PaidCharacterUpkeep != 10 || status.CharacterShortage != 0 {
		t.Fatalf("character upkeep status = %+v; want 10 paid in full", status)
	}
	if status.Upkeep != 5 || status.PaidUpkeep != 2 || status.Shortage != 3 {
		t.Fatalf("army upkeep status = %+v; want 2 paid and 3 shortage after character cost", status)
	}
	if gs.Factions[fid].Gold != 0 {
		t.Fatalf("remaining gold = %d, want 0", gs.Factions[fid].Gold)
	}
}

func TestCouncilAssignedCommanderUpkeepIsChargedEachTurn(t *testing.T) {
	const fid faction.FactionID = "player"
	commander := &army.Commander{
		ID: "cmd_tyrion", Type: army.CharacterTypeCommander, OwnerID: string(fid),
		AssignedCouncilRole: state.CouncilRoleTreasurer, GoldUpkeep: 7,
	}
	gs := &state.GameState{
		Turn:            1,
		Year:            300,
		Month:           1,
		PlayerFactionID: fid,
		Factions: map[faction.FactionID]*faction.Faction{
			fid: {ID: fid, Gold: 100, Grain: 100},
		},
		Commanders: map[string]*army.Commander{commander.ID: commander},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{{
				ID: state.CouncilRoleTreasurer, NameTR: "Hazinedar",
			}},
		},
	}

	applyEconomyTick(gs)

	status := gs.GoldEconomy[fid]
	if status.CharacterUpkeep != 7 || status.PaidCharacterUpkeep != 7 || status.CharacterShortage != 0 {
		t.Fatalf("commander character upkeep status = %+v; want 7 paid in full", status)
	}
	if got, want := gs.Factions[fid].Gold, 93; got != want {
		t.Fatalf("remaining gold = %d, want %d after commander upkeep", got, want)
	}
}

func TestMilitaryResponsibleGainsExperienceWhenMilitaryBonusApplies(t *testing.T) {
	const fid faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_military", Type: army.CharacterTypeAdministrator, OwnerID: string(fid), Name: "Marshal",
	}
	gs := &state.GameState{
		Turn:            1,
		Year:            300,
		Month:           1,
		PlayerFactionID: fid,
		Factions:        map[faction.FactionID]*faction.Faction{fid: {ID: fid, Gold: 1000, Grain: 100}},
		Commanders:      map[string]*army.Commander{character.ID: character},
		CouncilConfig:   scenario.DefaultCharacterCouncilConfig(),
	}
	if !gs.AssignCharacterToCouncilRole(fid, character.ID, state.CouncilRoleMilitaryResponsible) {
		t.Fatal("military responsible assignment failed")
	}

	gs.RecordCouncilBattleExperience(fid)
	if got := gs.CouncilMilitaryPowerMod(fid); math.Abs(got-1.02) > 1e-9 {
		t.Fatalf("military power mod without military trait = %.2f, want 1.02", got)
	}
	if character.CouncilExperience != scenario.DefaultCouncilXPPerEffect {
		t.Fatalf("military responsible experience = %d, want effect XP %d", character.CouncilExperience, scenario.DefaultCouncilXPPerEffect)
	}
}

func TestPlayerResearchUsesCouncilResearchSpeed(t *testing.T) {
	const fid faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_scientist", Type: army.CharacterTypeCharacter, OwnerID: string(fid), Name: "Scientist",
	}
	gs := &state.GameState{
		Year:            1,
		PlayerFactionID: fid,
		Factions:        map[faction.FactionID]*faction.Faction{fid: {ID: fid, Gold: 100}},
		Commanders:      map[string]*army.Commander{character.ID: character},
		TechTypes: map[string]*tech.Technology{
			"writing": {ID: "writing", NameTR: "Yazı", TurnsRequired: 10},
		},
		CouncilConfig: scenario.CharacterCouncilConfig{
			Roles: []scenario.CouncilRoleDefinition{{
				ID: state.CouncilRoleResearchResponsible, NameTR: "Bilim Sorumlusu",
				Effects: scenario.CouncilEffects{ResearchSpeedMod: 2},
			}},
		},
	}
	if !gs.AssignCharacterToCouncilRole(fid, character.ID, state.CouncilRoleResearchResponsible) {
		t.Fatal("could not assign research responsible")
	}

	game := &Game{gs: gs, renderer: &render.Renderer{}}
	game.startResearch("writing")
	if got := gs.Factions[fid].Research.TurnsLeft; got != 5 {
		t.Fatalf("research turns left with council speed = %d, want 5", got)
	}
	if character.CouncilExperience != scenario.DefaultCouncilXPPerEffect {
		t.Fatalf("research responsible XP = %d, want %d", character.CouncilExperience, scenario.DefaultCouncilXPPerEffect)
	}
}
