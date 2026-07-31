package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

func navalMissionStateFixture() *GameState {
	return &GameState{
		PlayerFactionID: "player",
		UnitTypes: map[string]*army.UnitType{
			"warship":   {ID: "warship", Category: army.CategoryNavalWar},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10},
			"infantry":  {ID: "infantry", Category: army.CategoryInfantry, Embarkable: true},
		},
		Regions: map[world.RegionID]*world.Region{
			"sea":    {ID: "sea", IsSea: true},
			"coast":  {ID: "coast"},
			"inland": {ID: "inland"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"war-fleet": {
				ID: "war-fleet", OwnerID: "player", IsNaval: true,
				Units: []army.Unit{{TypeID: "warship"}},
			},
			"transport-fleet": {
				ID: "transport-fleet", OwnerID: "player", IsNaval: true,
				Units: []army.Unit{{TypeID: "transport"}},
			},
		},
	}
}

func TestCanAssignNavalMissionValidatesPlayerFleetRoles(t *testing.T) {
	gs := navalMissionStateFixture()
	gs.Regions["coast"].Neighbors = []world.RegionID{"sea"}
	gs.Regions["sea"].Neighbors = []world.RegionID{"coast"}
	gs.Armies["transport-fleet"].EmbarkedUnits = []army.Unit{{TypeID: "infantry"}}

	cases := []struct {
		name    string
		mission army.NavalMission
	}{
		{name: "devriye", mission: army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "sea"}},
		{name: "abluka", mission: army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "sea"}},
		{name: "escort", mission: army.NavalMission{Kind: army.NavalMissionEscort, TargetFleetID: "transport-fleet"}},
		{name: "nakliye", mission: army.NavalMission{Kind: army.NavalMissionTransport, TargetRegionID: "coast"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if ok, reason := gs.AssignNavalMission(map[bool]army.ArmyID{true: "war-fleet", false: "transport-fleet"}[tc.name == "devriye" || tc.name == "abluka" || tc.name == "escort"], tc.mission); !ok {
				t.Fatalf("görev atanamadı: %s", reason)
			}
		})
	}
	if gs.Armies["war-fleet"].NavalMission == nil || gs.Armies["transport-fleet"].NavalMission == nil {
		t.Fatal("geçerli görevler filolarda saklanmadı")
	}
}

func TestCanAssignNavalMissionRejectsInvalidTargetsAndRoles(t *testing.T) {
	gs := navalMissionStateFixture()
	gs.Regions["coast"].Neighbors = []world.RegionID{"sea"}

	invalid := []struct {
		name    string
		fleetID army.ArmyID
		mission army.NavalMission
	}{
		{name: "savaş gemisi olmayan filo devriye atamaz", fleetID: "transport-fleet", mission: army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "sea"}},
		{name: "kara bölgesi devriye hedefi olamaz", fleetID: "war-fleet", mission: army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "inland"}},
		{name: "nakliye için taşınan ordu gerekir", fleetID: "transport-fleet", mission: army.NavalMission{Kind: army.NavalMissionTransport, TargetRegionID: "coast"}},
		{name: "nakliye iç bölgeye atanamaz", fleetID: "transport-fleet", mission: army.NavalMission{Kind: army.NavalMissionTransport, TargetRegionID: "inland"}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if ok, reason := gs.CanAssignNavalMission(tc.fleetID, tc.mission); ok {
				t.Fatalf("geçersiz görev kabul edildi: %s", reason)
			}
		})
	}
}

func TestClearNavalMissionRemovesPlayerTask(t *testing.T) {
	gs := navalMissionStateFixture()
	gs.Regions["sea"].IsSea = true
	if ok, reason := gs.AssignNavalMission("war-fleet", army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "sea"}); !ok {
		t.Fatalf("test görevi atanamadı: %s", reason)
	}
	if !gs.ClearNavalMission("war-fleet") || gs.Armies["war-fleet"].NavalMission != nil {
		t.Fatal("görev temizlenmedi")
	}
}
