package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestTreasurerBonusAppliedByEconomyTickOnlyToTaxIncome(t *testing.T) {
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
			Factions:        map[faction.FactionID]*faction.Faction{fid: {ID: fid, Gold: 0, Grain: 100}},
			Regions: map[world.RegionID]*world.Region{
				"capital": {ID: "capital", OwnerID: string(fid), BaseGoldIncome: 1000, TaxRate: 50, Satisfaction: 50},
			},
			Commanders: map[string]*army.Commander{character.ID: character},
		}
		if withTreasurer && !gs.AssignCharacterToCouncilRole(fid, character.ID, state.CouncilRoleTreasurer) {
			t.Fatal("treasurer assignment failed")
		}
		return gs
	}

	base := buildState(false)
	withTreasurer := buildState(true)
	baseReport := applyEconomyTick(base)
	treasurerReport := applyEconomyTick(withTreasurer)
	_ = baseReport
	_ = treasurerReport

	baseStatus := base.GoldEconomy[fid]
	treasurerStatus := withTreasurer.GoldEconomy[fid]
	wantBonus := baseStatus.TaxIncome * state.TreasurerTaxIncomeBonusPercent / 100
	if wantBonus <= 0 {
		t.Fatalf("fixture produced no tax bonus: base tax income = %d", baseStatus.TaxIncome)
	}
	if got, want := treasurerStatus.TaxIncome, baseStatus.TaxIncome+wantBonus; got != want {
		t.Fatalf("economy tick tax income with treasurer = %d, want %d", got, want)
	}
	if got, want := treasurerStatus.Income, baseStatus.Income+wantBonus; got != want {
		t.Fatalf("economy tick income with treasurer = %d, want %d", got, want)
	}
	if got, want := treasurerStatus.TradeIncome, baseStatus.TradeIncome; got != want {
		t.Fatalf("treasurer changed trade income: got %d want %d", got, want)
	}
}
