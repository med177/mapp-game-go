package ai

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func buildingInvestmentFixture() (*state.GameState, faction.FactionID) {
	fid := faction.FactionID("player")
	return &state.GameState{
		PlayerFactionID: fid,
		Factions: map[faction.FactionID]*faction.Faction{
			fid: {
				ID:     fid,
				Gold:   5000,
				Grain:  5000,
				Iron:   5000,
				Timber: 5000,
				Stone:  5000,
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {
				ID:              "capital",
				OwnerID:         string(fid),
				BaseGoldIncome:  100,
				BaseGrainOutput: 100,
				TradeCapacity:   2,
				Satisfaction:    60,
				Population:      100,
			},
		},
		RegionOrder:   []world.RegionID{"capital"},
		BuildingOrder: []string{"market", "farm"},
		BuildingTypes: map[string]*city.Building{
			"market": {
				ID:                    "market",
				NameTR:                "Pazar",
				MaxPerRegion:          3,
				TurnsRequired:         2,
				UpgradeCostMultiplier: 1.2,
				GoldCost:              100,
				GrainCost:             10,
				GoldMod:               1.2,
				TradeCapacityMod:      1.1,
			},
			"farm": {
				ID:                    "farm",
				NameTR:                "Çiftlik",
				MaxPerRegion:          3,
				TurnsRequired:         2,
				UpgradeCostMultiplier: 1.2,
				GoldCost:              100,
				GrainCost:             10,
				GrainMod:              1.2,
			},
		},
	}, fid
}

func queuedBuildingCountForTest(gs *state.GameState, regionID world.RegionID, buildingID string) int {
	count := 0
	for _, order := range gs.ProductionQueue {
		if order.Kind == aiProductionKindBuilding && order.RegionID == regionID && order.TypeID == buildingID {
			count++
		}
	}
	return count
}

func TestAIEconomyQueuesDifferentBuildingTypesButNotDuplicateType(t *testing.T) {
	gs, fid := buildingInvestmentFixture()
	ctx := &StrategicContext{FactionID: fid, gs: gs}

	aiEconomyBuildWithStrategicContextAndSteps(gs, fid, nil, ctx, nil)

	if got := len(gs.ProductionQueue); got != 2 {
		t.Fatalf("AI uygun iki bina türünü kuyruğa almalı: got=%d want=2 queue=%+v", got, gs.ProductionQueue)
	}
	for _, buildingID := range []string{"market", "farm"} {
		if got := queuedBuildingCountForTest(gs, "capital", buildingID); got != 1 {
			t.Fatalf("%s için tek emir bekleniyordu: got=%d", buildingID, got)
		}
	}
}

func TestAIEconomyDoesNotQueueSameBuildingBehindExistingOrder(t *testing.T) {
	gs, fid := buildingInvestmentFixture()
	gs.ProductionQueue = []state.ProductionOrder{{
		ID:        "existing",
		Kind:      aiProductionKindBuilding,
		FactionID: string(fid),
		RegionID:  "capital",
		TypeID:    "market",
		TurnsLeft: 2,
	}}
	ctx := &StrategicContext{FactionID: fid, gs: gs}

	aiEconomyBuildWithStrategicContextAndSteps(gs, fid, nil, ctx, nil)

	if got := queuedBuildingCountForTest(gs, "capital", "market"); got != 1 {
		t.Fatalf("mevcut Pazar emrinin arkasına ikinci emir eklenmemeli: got=%d", got)
	}
	if got := queuedBuildingCountForTest(gs, "capital", "farm"); got != 1 {
		t.Fatalf("mevcut Pazar emri varken uygun Çiftlik emri verilmeli: got=%d", got)
	}
}

func TestAIEconomyPrefersBuildingMatchingRegionalResourceProfile(t *testing.T) {
	gs, fid := buildingInvestmentFixture()
	region := gs.Regions["capital"]
	region.BaseIronOutput = 100
	region.BaseGrainOutput = 10
	gs.BuildingOrder = []string{"forge", "farm"}
	gs.BuildingTypes["forge"] = &city.Building{
		ID:                    "forge",
		NameTR:                "Demirhane",
		MaxPerRegion:          3,
		TurnsRequired:         2,
		UpgradeCostMultiplier: 1.2,
		GoldCost:              100,
		GrainCost:             10,
		GoldMod:               1.05,
		IronMod:               1.2,
	}
	ctx := &StrategicContext{FactionID: fid, gs: gs}

	aiEconomyBuildWithStrategicContextAndSteps(gs, fid, nil, ctx, nil)

	if len(gs.ProductionQueue) == 0 || gs.ProductionQueue[0].TypeID != "forge" {
		t.Fatalf("demir ağırlıklı bölgede demirhane önce seçilmeli: queue=%+v", gs.ProductionQueue)
	}
}
