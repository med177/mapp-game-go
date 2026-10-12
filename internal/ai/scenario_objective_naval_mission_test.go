package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIWallBreachInvalidatesBreakWallPlanAndSelectsNorthernInvasion(t *testing.T) {
	gs := &state.GameState{
		Turn: 119,
		Year: 304,
		FiredEventIDs: map[string]bool{
			"flag:others_awakened": true,
			"flag:wall_breached":   true,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"others":    {ID: "others"},
			"free_folk": {ID: "free_folk"},
			"lannister": {ID: "lannister"},
			"stark":     {ID: "stark"},
		},
		Regions: map[world.RegionID]*world.Region{
			"always_winter":     {ID: "always_winter", OwnerID: "others", Neighbors: []world.RegionID{"lannister_region"}},
			"lannister_region":  {ID: "lannister_region", OwnerID: "lannister"},
			"the_wall_west":     {ID: "the_wall_west", OwnerID: "free_folk"},
			"winterfell_region": {ID: "winterfell_region", OwnerID: "stark"},
		},
		AIStrategies: map[string]scenario.AIFactionStrategy{
			"others": {
				FactionID: "others",
				Objectives: []scenario.AIObjectiveDef{
					{
						ID:                 "awaken_the_dead",
						Kind:               "defend",
						TerritorialClaims:  []scenario.AITerritorialClaimDef{{RegionID: "always_winter", Value: 100}},
						Priority:           100,
						RequiredEventFlags: []string{"others_awakened"},
						BlockedEventFlags:  []string{"wall_breached"},
					},
					{
						ID:                 "break_the_wall",
						Kind:               "expand",
						TerritorialClaims:  []scenario.AITerritorialClaimDef{{RegionID: "the_wall_west", Value: 100}},
						Priority:           100,
						RequiredEventFlags: []string{"others_awakened"},
						BlockedEventFlags:  []string{"wall_breached"},
					},
					{
						ID:                 "descend_into_the_north",
						Kind:               "expand",
						TerritorialClaims:  []scenario.AITerritorialClaimDef{{RegionID: "winterfell_region", Value: 90}},
						Priority:           70,
						RequiredEventFlags: []string{"wall_breached"},
					},
				},
			},
		},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"others": {
				ObjectiveID:     "break_the_wall",
				Kind:            state.AIObjectiveExpand,
				TargetFactionID: "free_folk",
				TargetRegionIDs: []world.RegionID{"the_wall_west"},
				ReassessTurn:    121,
			},
		},
	}

	if aiStrategicPlanValid(gs, "others", gs.AIPlans["others"]) {
		t.Fatal("wall_breached sonrasında break_the_wall planı geçerli kaldı")
	}
	plan := chooseScenarioObjectivePlan(gs, gs.Factions["others"], buildStrategicContext(gs, "others"))
	if plan == nil || plan.ObjectiveID != "descend_into_the_north" {
		t.Fatalf("Duvar yarıldıktan sonra kuzeye iniş seçilmedi: %#v", plan)
	}
}

func TestAINavalMissionRequiresExistingOrProducibleTransport(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":    {ID: "ai", Research: faction.ResearchState{Completed: map[string]bool{}}},
			"enemy": {ID: "enemy"},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {
				ID: "infantry", Category: army.CategoryInfantry, Embarkable: true,
			},
			"transport": {
				ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 2,
				RequiredTech:      []string{"coastal_navigation"},
				RequiredBuildings: []army.BuildingRequirement{{ID: "port", Level: 1}},
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"home":       {ID: "home", OwnerID: "ai", Neighbors: []world.RegionID{"coast"}},
			"coast":      {ID: "coast", OwnerID: "ai", Neighbors: []world.RegionID{"home", "home_sea"}, Buildings: []string{"port"}},
			"home_sea":   {ID: "home_sea", IsSea: true, Neighbors: []world.RegionID{"coast", "target_sea"}},
			"target_sea": {ID: "target_sea", IsSea: true, Neighbors: []world.RegionID{"home_sea", "target"}},
			"target":     {ID: "target", OwnerID: "enemy", Neighbors: []world.RegionID{"target_sea"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"land": {
				ID: "land", OwnerID: "ai", RegionID: "home", MovePoints: 3,
				Units: []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}},
			},
		},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"ai": {
				Kind:            state.AIObjectiveExpand,
				TargetFactionID: "enemy",
				TargetRegionIDs: []world.RegionID{"target"},
			},
		},
	}
	ctx := buildStrategicContext(gs, "ai")

	if mission := buildAINavalMission(ctx); mission != nil {
		t.Fatalf("nakliye teknolojisi olmadan imkânsız deniz görevi oluştu: %#v", mission)
	}

	gs.Factions["ai"].Research.Completed["coastal_navigation"] = true
	mission := buildAINavalMission(ctx)
	if mission == nil || mission.TargetRegionID != "target" || mission.MissingCapacity != 1 {
		t.Fatalf("nakliye üretilebilirken deniz görevi oluşmadı: %#v", mission)
	}
}
