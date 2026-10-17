package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestEconomyTickIncludesNavalFishingInGrainProduction(t *testing.T) {
	const fid faction.FactionID = "player"
	gs := &state.GameState{
		Turn:            1,
		Year:            299,
		Month:           1,
		PlayerFactionID: fid,
		Factions: map[faction.FactionID]*faction.Faction{
			fid: {ID: fid, Grain: 90},
		},
		Regions: map[world.RegionID]*world.Region{
			"coast": {ID: "coast", OwnerID: string(fid), Neighbors: []world.RegionID{"sea"}},
			"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
		},
		UnitTypes: map[string]*army.UnitType{
			"warship": {
				ID: "warship", Category: army.CategoryNavalWar,
				GrainUpkeep: 20, FishingGrainIncome: 30,
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID: "fleet", OwnerID: string(fid), IsNaval: true, RegionID: "sea",
				Units:        []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}},
				NavalMission: &army.NavalMission{Kind: army.NavalMissionFishing, TargetRegionID: "sea"},
			},
		},
	}

	applyEconomyTick(gs)

	status := gs.GrainEconomy[fid]
	if got, want := status.Production, 30; got != want {
		t.Fatalf("grain production = %d, want %d including fishing", got, want)
	}
	if got, want := status.ArmyUpkeep, 20; got != want {
		t.Fatalf("army grain upkeep = %d, want %d", got, want)
	}
	if got, want := status.NetChange, 10; got != want {
		t.Fatalf("net grain change = %d, want %d after ship upkeep", got, want)
	}
	if got, want := gs.Factions[fid].Grain, 100; got != want {
		t.Fatalf("grain stockpile = %d, want %d after fishing", got, want)
	}
}
