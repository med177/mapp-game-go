package victory

import (
	"testing"

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
