package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestArmyLogisticsBadgeRefreshUpdatesWholeRegionAfterMovement(t *testing.T) {
	const (
		player faction.FactionID = "player"
		front  world.RegionID    = "front"
		rear   world.RegionID    = "rear"
	)
	gs := &state.GameState{
		PlayerFactionID: player,
		Regions: map[world.RegionID]*world.Region{
			front: {ID: front, OwnerID: string(player)},
			rear:  {ID: rear, OwnerID: string(player)},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			player: {ID: player},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army-1": {ID: "army-1", OwnerID: string(player), RegionID: front, Units: []army.Unit{{TypeID: "infantry"}}},
			"army-2": {ID: "army-2", OwnerID: string(player), RegionID: front, Units: []army.Unit{{TypeID: "infantry"}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", GrainUpkeep: 4},
		},
		ArmyLogistics: map[army.ArmyID]state.ArmyLogisticsStatus{
			"army-1": {ArmyID: "army-1", RegionID: front, TotalHPDamage: 1},
			"army-2": {ArmyID: "army-2", RegionID: front, TotalHPDamage: 1},
		},
	}
	r := &Renderer{gs: gs, mapMode: MapModeNormal}
	r.RefreshAllArmyLogisticsBadges()
	if !r.armyLogisticsDamageVisible("army-1") || !r.armyLogisticsDamageVisible("army-2") {
		t.Fatal("aşırı yüklü bölgedeki iki ordunun da ikmal rozeti görünür olmalı")
	}

	gs.Armies["army-2"].RegionID = rear
	r.RefreshArmyLogisticsBadgesForRegions(front, rear)
	if r.armyLogisticsDamageVisible("army-1") || r.armyLogisticsDamageVisible("army-2") {
		t.Fatal("hareket sonrası eski ve yeni bölge marker cache'i birlikte yenilenmedi")
	}
}

func TestArmyLogisticsBadgeRefreshTracksSupplyMissionConnection(t *testing.T) {
	const (
		player faction.FactionID = "player"
		land   world.RegionID    = "coastal_front"
		sea    world.RegionID    = "coastal_sea"
	)
	gs := &state.GameState{
		PlayerFactionID: player,
		Regions: map[world.RegionID]*world.Region{
			land: {
				ID: land, OwnerID: string(player), Neighbors: []world.RegionID{sea},
				Settlements: []world.Settlement{{ID: "port", Type: world.SettlementPort}},
			},
			sea: {ID: sea, IsSea: true, Neighbors: []world.RegionID{land}},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			player: {ID: player},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army-1": {ID: "army-1", OwnerID: string(player), RegionID: land, Units: []army.Unit{{TypeID: "infantry"}}},
			"army-2": {ID: "army-2", OwnerID: string(player), RegionID: land, Units: []army.Unit{{TypeID: "infantry"}}},
			"fleet":  {ID: "fleet", OwnerID: string(player), RegionID: sea, IsNaval: true, SupplyCargo: economy.ResourceCost{Grain: 4}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", GrainUpkeep: 4},
		},
	}
	r := &Renderer{gs: gs, mapMode: MapModeNormal}
	r.RefreshAllArmyLogisticsBadges()
	if !r.armyLogisticsDamageVisible("army-1") || !r.armyLogisticsDamageVisible("army-2") {
		t.Fatal("bağlantısız ikmal durumunda iki ordunun da uyarısı görünür olmalı")
	}

	gs.Armies["fleet"].NavalMission = &army.NavalMission{
		Kind:         army.NavalMissionSupplyArmy,
		TargetArmyID: "army-1",
	}
	r.RefreshArmyLogisticsBadge("army-1")
	if r.armyLogisticsDamageVisible("army-1") || r.armyLogisticsDamageVisible("army-2") {
		t.Fatal("ikmal görevi bağlanınca aynı bölgedeki eski uyarılar gizlenmedi")
	}

	gs.Armies["fleet"].NavalMission = nil
	r.RefreshArmyLogisticsBadge("army-1")
	if !r.armyLogisticsDamageVisible("army-1") || !r.armyLogisticsDamageVisible("army-2") {
		t.Fatal("ikmal bağlantısı kopunca bölgesel uyarılar yeniden gösterilmedi")
	}
}
