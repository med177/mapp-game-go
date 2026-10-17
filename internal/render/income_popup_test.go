package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestPlayerGrainEconomyStatusUsesLiveFishingProjection(t *testing.T) {
	const fid faction.FactionID = "player"
	fleet := &army.Army{
		ID: "fleet", OwnerID: string(fid), IsNaval: true, RegionID: "sea",
		NavalMission: &army.NavalMission{Kind: army.NavalMissionFishing, TargetRegionID: "sea"},
		Units: []army.Unit{
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
		},
	}
	gs := &state.GameState{
		PlayerFactionID: fid,
		GrainEconomy: map[faction.FactionID]state.GrainEconomyStatus{
			fid: {
				FactionID:      fid,
				Production:     10,
				CivilianDemand: 5,
				ArmyUpkeep:     7,
				TotalDemand:    12,
				NetChange:      -2,
				AutoExportSold: 4,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"warship": {ID: "warship", Category: army.CategoryNavalWar, FishingGrainIncome: 40, GrainUpkeep: 20},
		},
		Regions: map[world.RegionID]*world.Region{
			"coast": {ID: "coast", OwnerID: string(fid), Neighbors: []world.RegionID{"sea"}},
			"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
		},
		Armies: map[army.ArmyID]*army.Army{"fleet": fleet},
	}

	status := playerGrainEconomyStatus(gs)
	if status.Production != 240 || status.ArmyUpkeep != 120 || status.TotalDemand != 120 || status.NetChange != 120 {
		t.Fatalf("grain popup preview = production %d, upkeep %d, total demand %d, net %d; want 240, 120, 120, 120",
			status.Production, status.ArmyUpkeep, status.TotalDemand, status.NetChange)
	}
	if status.AutoExportSold != 4 {
		t.Fatalf("cached auto-export value = %d, want 4", status.AutoExportSold)
	}
}
