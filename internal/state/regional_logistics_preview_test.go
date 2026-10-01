package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestPreviewRegionalLogisticsIncludesSettlementGranaryAndReserveSupport(t *testing.T) {
	const (
		factionID = faction.FactionID("ottoman")
		regionID  = world.RegionID("bilecik_frontier")
	)

	gs := &GameState{
		PlayerFactionID: factionID,
		Regions: map[world.RegionID]*world.Region{
			regionID: {
				ID:      regionID,
				OwnerID: string(factionID),
				Settlements: []world.Settlement{
					{ID: "bilecik", Type: world.SettlementTown},
				},
				Buildings: []string{"granary"},
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			factionID: {ID: factionID, Grain: 1000},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army-1": {
				ID:       "army-1",
				OwnerID:  string(factionID),
				RegionID: regionID,
				Units:    []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", GrainUpkeep: 10},
		},
		BuildingTypes: map[string]*city.Building{
			"granary": {ID: "granary", StorageCapacity: 100},
		},
	}

	status, ok := gs.PreviewRegionalLogisticsStatus(regionID)
	if !ok {
		t.Fatal("ikmal önizlemesi bölgedeki ordu için bulunamadı")
	}
	if status.Demand != 10 {
		t.Fatalf("talep = %d, 10 olmalı", status.Demand)
	}
	if status.LocalProduction != 0 {
		t.Fatalf("yerel üretim = %d, 0 olmalı", status.LocalProduction)
	}
	if status.SettlementBuffer != 5 {
		t.Fatalf("yerleşim tamponu = %d, 5 olmalı", status.SettlementBuffer)
	}
	if status.GranarySupport != 100 {
		t.Fatalf("ambar desteği = %d, 100 olmalı", status.GranarySupport)
	}
	if status.ReserveSupport != 6 {
		t.Fatalf("merkez rezerv desteği = %d, 6 olmalı", status.ReserveSupport)
	}
	if status.Capacity != 111 {
		t.Fatalf("toplam kapasite = %d, 111 olmalı", status.Capacity)
	}
	if status.Overload != -101 {
		t.Fatalf("aşım = %d, -101 olmalı", status.Overload)
	}
}

func TestPreviewRegionalLogisticsStatusesUsesInvalidatableCache(t *testing.T) {
	const regionID = world.RegionID("front")
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			regionID: {ID: regionID, OwnerID: "player"},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army": {ID: "army", OwnerID: "player", RegionID: regionID, Units: []army.Unit{{TypeID: "infantry"}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", GrainUpkeep: 10},
		},
	}

	first := gs.PreviewRegionalLogisticsStatuses()[regionID]
	gs.Armies["army"].Units = append(gs.Armies["army"].Units, army.Unit{TypeID: "infantry"})
	withoutInvalidation := gs.PreviewRegionalLogisticsStatuses()[regionID]
	if withoutInvalidation.Demand != first.Demand {
		t.Fatalf("cache invalid olmadan talep değişti: ilk=%d, sonra=%d", first.Demand, withoutInvalidation.Demand)
	}

	gs.InvalidateRegionalLogisticsPreview()
	afterInvalidation := gs.PreviewRegionalLogisticsStatuses()[regionID]
	if afterInvalidation.Demand <= first.Demand {
		t.Fatalf("cache invalidation sonrası yeni talep hesaplanmadı: ilk=%d, yeni=%d", first.Demand, afterInvalidation.Demand)
	}
}

func TestArmyLogisticsDamageVisibleUsesCurrentNavalSupplyPreview(t *testing.T) {
	const factionID = faction.FactionID("player")
	const landID = world.RegionID("coastal_front")
	const seaID = world.RegionID("coastal_sea")

	buildState := func(cargo int) *GameState {
		mission := &army.NavalMission{Kind: army.NavalMissionSupplyArmy, TargetArmyID: "army"}
		return &GameState{
			PlayerFactionID: factionID,
			Regions: map[world.RegionID]*world.Region{
				landID: {
					ID: landID, OwnerID: string(factionID), Neighbors: []world.RegionID{seaID},
					Settlements: []world.Settlement{{ID: "port", Type: world.SettlementPort}},
				},
				seaID: {ID: seaID, IsSea: true, Neighbors: []world.RegionID{landID}},
			},
			Factions: map[faction.FactionID]*faction.Faction{
				factionID: {ID: factionID},
			},
			Armies: map[army.ArmyID]*army.Army{
				"army": {
					ID: "army", OwnerID: string(factionID), RegionID: landID,
					Units: []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}},
				},
				"fleet": {
					ID: "fleet", OwnerID: string(factionID), RegionID: seaID, IsNaval: true,
					SupplyCargo: economy.ResourceCost{Grain: cargo}, NavalMission: mission,
				},
			},
			UnitTypes: map[string]*army.UnitType{
				"infantry": {ID: "infantry", GrainUpkeep: 10},
			},
			ArmyLogistics: map[army.ArmyID]ArmyLogisticsStatus{
				"army": {ArmyID: "army", RegionID: landID, TotalHPDamage: 5},
			},
		}
	}

	if got := buildState(4).ArmyLogisticsDamageVisible("army"); got {
		t.Fatal("açığı kapatan deniz ikmali varken kırmızı zayiat rozeti görünür kaldı")
	}
	if got := buildState(2).ArmyLogisticsDamageVisible("army"); !got {
		t.Fatal("yetersiz deniz ikmalinde kırmızı zayiat rozeti gizlendi")
	}
}
