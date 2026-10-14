package victory

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestOtherIncomeIsIncludedInGoldIncomeCalculations(t *testing.T) {
	const fid faction.FactionID = "papal_states_f"
	gs := &state.GameState{
		Turn:  1,
		Year:  1300,
		Month: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			fid: {ID: fid, OtherIncomePeriods: []faction.OtherIncomePeriod{{StartYear: 1300, Amount: 35, Description: "Kilise gelirleri"}}},
		},
	}

	preview := GoldEconomyPreview(gs, fid)
	if preview.OtherIncome != 35 || preview.Income != 35 {
		t.Fatalf("other income preview = (%d, total %d), want (35, 35)", preview.OtherIncome, preview.Income)
	}
	if preview.OtherIncomeDescription != "Kilise gelirleri" {
		t.Fatalf("other income description = %q, want %q", preview.OtherIncomeDescription, "Kilise gelirleri")
	}
	if got := GoldIncomeForFaction(gs, fid); got != 35 {
		t.Fatalf("gross gold income = %d, want 35", got)
	}
}

func TestTreasurerBonusMatchesTaxPreviewAndGrossIncome(t *testing.T) {
	const fid faction.FactionID = "player"
	character := &army.Commander{ID: "cmd_tywin", Type: army.CharacterTypeAdministrator, OwnerID: string(fid), Name: "Tywin"}
	region := &world.Region{ID: "capital", OwnerID: string(fid), BaseGoldIncome: 1000, TaxRate: 50, Satisfaction: 50}
	gs := &state.GameState{
		Year:            300,
		Month:           6,
		PlayerFactionID: fid,
		Factions:        map[faction.FactionID]*faction.Faction{fid: {ID: fid}},
		Regions:         map[world.RegionID]*world.Region{region.ID: region},
		Commanders:      map[string]*army.Commander{character.ID: character},
	}

	basePreview := GoldEconomyPreview(gs, fid)
	baseGross := GoldIncomeForFaction(gs, fid)
	if !gs.AssignCharacterToCouncilRole(fid, character.ID, state.CouncilRoleTreasurer) {
		t.Fatal("treasurer assignment failed")
	}
	preview := GoldEconomyPreview(gs, fid)
	bonus := basePreview.TaxIncome * state.TreasurerTaxIncomeBonusPercent / 100
	if got, want := preview.TaxIncome, basePreview.TaxIncome+bonus; got != want {
		t.Fatalf("tax preview with treasurer = %d, want %d", got, want)
	}
	if got, want := GoldIncomeForFaction(gs, fid), baseGross+bonus; got != want {
		t.Fatalf("gross income with treasurer = %d, want %d", got, want)
	}
	if preview.TradeIncome != basePreview.TradeIncome || preview.OtherIncome != basePreview.OtherIncome {
		t.Fatalf("treasurer changed non-tax categories: trade %d->%d, other %d->%d", basePreview.TradeIncome, preview.TradeIncome, basePreview.OtherIncome, preview.OtherIncome)
	}
}

func TestEventOtherIncomeDeltaIsIncludedWithHistoricalPeriod(t *testing.T) {
	const fid = faction.FactionID("papal_states")
	gs := &state.GameState{
		Year:  1400,
		Month: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			fid: {
				ID:                 fid,
				OtherIncomePeriods: []faction.OtherIncomePeriod{{StartYear: 1300, Amount: 35}},
				OtherIncomeDelta:   20,
			},
		},
		Regions: map[world.RegionID]*world.Region{},
	}

	preview := GoldEconomyPreview(gs, fid)
	if preview.OtherIncome != 55 || preview.Income != 55 {
		t.Fatalf("period ve event geliri = (%d, toplam %d), (55, 55) bekleniyordu", preview.OtherIncome, preview.Income)
	}
}

func TestEconomicVictoryCountsOtherIncome(t *testing.T) {
	const fid faction.FactionID = "papal_states_f"
	gs := &state.GameState{
		Year:            1300,
		Month:           1,
		PlayerFactionID: fid,
		Factions: map[faction.FactionID]*faction.Faction{
			fid: {ID: fid, OtherIncomePeriods: []faction.OtherIncomePeriod{{StartYear: 1300, Amount: 35, Description: "Kilise gelirleri"}}},
		},
		Victory: state.VictoryCondition{
			TargetGoldIncome: 35,
			GoldHoldTurns:    1,
		},
	}

	checkEconomic(gs)
	if gs.EconomicVictoryTurns != 1 {
		t.Fatalf("economic victory turns = %d, want 1 when other income reaches target", gs.EconomicVictoryTurns)
	}
}

func TestOtherIncomeChangesAtHistoricalPeriodBoundary(t *testing.T) {
	f := &faction.Faction{
		OtherIncomePeriods: []faction.OtherIncomePeriod{
			{StartYear: 1300, EndYear: 1517, EndMonth: 9, Amount: 153},
			{StartYear: 1517, StartMonth: 10, Amount: 30},
		},
	}

	if got := f.OtherIncomeAt(1300, 0); got != 153 {
		t.Fatalf("missing month should use January, got %d before reform", got)
	}
	if got := f.OtherIncomeAt(1517, 9); got != 153 {
		t.Fatalf("income at reform boundary before October = %d, want 153", got)
	}
	if got := f.OtherIncomeAt(1517, 10); got != 30 {
		t.Fatalf("income after reform boundary = %d, want 30", got)
	}
}

func TestPrivilegedTradePreviewShowsOperatorValueAndGrantorRoyalty(t *testing.T) {
	const grantor faction.FactionID = "grantor"
	const operator faction.FactionID = "operator"
	parent := &world.Region{ID: "parent", OwnerID: string(grantor)}
	minor := &world.Region{
		ID: "minor", OwnerID: string(operator), IsMinorRegion: true,
		IsPrivileged: true, ParentRegionID: parent.ID,
	}
	gs := &state.GameState{
		Year: 1300, Month: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			grantor:  {ID: grantor, Cloth: 10, Gold: 100},
			operator: {ID: operator, Gold: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			parent.ID: parent,
			minor.ID:  minor,
		},
		TradeRoutes: []*economy.TradeRoute{{
			FromFactionID: string(grantor), ToFactionID: string(operator),
			Good: economy.GoodCloth, AmountPerTurn: 4, IsPrivilegedMinor: true,
		}},
		BaseGoldValues: map[economy.GoodType]int{economy.GoodCloth: 8},
	}

	operatorPreview := GoldEconomyPreview(gs, operator)
	if operatorPreview.PrivilegedTradeIncome != 32 {
		t.Fatalf("imtiyaz alanın ticaret geliri = %d, 32 bekleniyordu", operatorPreview.PrivilegedTradeIncome)
	}
	if operatorPreview.TradeRouteExpense != 0 || operatorPreview.TradeRouteCustomsIncome != 0 {
		t.Fatalf("imtiyaz alan için normal rota kesintisi oluştu: expense=%d customs=%d", operatorPreview.TradeRouteExpense, operatorPreview.TradeRouteCustomsIncome)
	}
	grantorPreview := GoldEconomyPreview(gs, grantor)
	if grantorPreview.PrivilegeIncome != 8 {
		t.Fatalf("imtiyaz verenin telif geliri = %d, 8 bekleniyordu", grantorPreview.PrivilegeIncome)
	}
	if grantorPreview.TradeRouteExpense != 0 || grantorPreview.TradeRouteCustomsIncome != 0 {
		t.Fatalf("imtiyaz verene normal rota kesintisi oluştu: expense=%d customs=%d", grantorPreview.TradeRouteExpense, grantorPreview.TradeRouteCustomsIncome)
	}
}
