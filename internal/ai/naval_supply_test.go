package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAICoastalSupplyLoadsAndAssignsTransportFleet(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", CapitalSettlementID: "capital_port", Grain: 100},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 2},
			"soldier":   {ID: "soldier", Category: army.CategoryInfantry, GrainUpkeep: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {
				ID: "capital", OwnerID: "ai", Neighbors: []world.RegionID{"sea"},
				Settlements: []world.Settlement{{ID: "capital_port", Type: world.SettlementPort}},
			},
			"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"capital", "coast"}},
			"coast": {ID: "coast", OwnerID: "ai", Neighbors: []world.RegionID{"sea"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID: "fleet", OwnerID: "ai", IsNaval: true, RegionID: "sea", DockedRegionID: "capital",
				Units: []army.Unit{{TypeID: "transport", CurrentHP: army.MaxUnitHP}}, MovePoints: 3, MaxMovePoints: 3,
			},
			"army": {
				ID: "army", OwnerID: "ai", RegionID: "coast",
				Units: []army.Unit{{TypeID: "soldier", CurrentHP: army.MaxUnitHP}, {TypeID: "soldier", CurrentHP: army.MaxUnitHP}},
			},
		},
	}

	ctx := buildStrategicContext(gs, "ai")
	ctx.navalSupplyMission = buildAINavalSupplyMission(ctx)
	if ctx.navalSupplyMission == nil || ctx.navalSupplyMission.TargetArmyID != "army" {
		t.Fatalf("AI kıyı ordusu için ikmal görevi üretmedi: %#v", ctx.navalSupplyMission)
	}
	aiPrepareNavalSupplyMission(gs, "ai", nil, ctx, nil)

	fleet := gs.Armies["fleet"]
	if fleet.SupplyCargo.Grain <= 0 {
		t.Fatalf("AI başkent limanında ikmal yüklemedi: %+v", fleet.SupplyCargo)
	}
	if gs.Factions["ai"].Grain >= 100 {
		t.Fatal("AI ikmal yükü için devlet stokundan tahıl düşülmedi")
	}

	moveArmyWithStrategicContext(gs, fleet, "ai", nil, ctx)
	if fleet.NavalMission == nil || fleet.NavalMission.Kind != army.NavalMissionSupplyArmy || fleet.NavalMission.TargetArmyID != "army" {
		t.Fatalf("AI filosu hedef kıyıda ikmal görevine bağlanmadı: %#v", fleet.NavalMission)
	}
}
