package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
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
