package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestConvertInvalidNavalBlockadesToPatrol(t *testing.T) {
	gs := navalMissionTransitionState()
	blockade := gs.Armies["blockade"]

	if got := gs.ConvertInvalidNavalBlockadesToPatrol(); got != 0 {
		t.Fatalf("geçerli abluka için dönüşüm sayısı = %d, want 0", got)
	}
	if blockade.NavalMission.Kind != army.NavalMissionBlockade {
		t.Fatalf("geçerli abluka görevi değişmemeli, got %q", blockade.NavalMission.Kind)
	}

	gs.Regions["enemy_port"].OwnerID = "player"
	if got := gs.ConvertInvalidNavalBlockadesToPatrol(); got != 1 {
		t.Fatalf("geçersiz abluka için dönüşüm sayısı = %d, want 1", got)
	}
	if blockade.NavalMission.Kind != army.NavalMissionPatrol {
		t.Fatalf("geçersiz abluka devriyeye dönüşmeli, got %q", blockade.NavalMission.Kind)
	}
	if blockade.NavalMission.TargetRegionID != "sea" {
		t.Fatalf("devriye hedefi mevcut deniz olmalı, got %q", blockade.NavalMission.TargetRegionID)
	}
}

func TestSupplyCargoLoadsAssignsAndUnloadsAtCapital(t *testing.T) {
	gs := &GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", CapitalSettlementID: "capital_port", Grain: 100},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 5},
			"soldier":   {ID: "soldier", Category: army.CategoryInfantry, GrainUpkeep: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: "player", Neighbors: []world.RegionID{"sea"}, Settlements: []world.Settlement{{ID: "capital_port", Type: world.SettlementPort}}},
			"sea":     {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"capital", "coast"}},
			"coast":   {ID: "coast", OwnerID: "player", Neighbors: []world.RegionID{"sea"}, Settlements: []world.Settlement{{ID: "coast_port", Type: world.SettlementPort}}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {ID: "fleet", OwnerID: "player", IsNaval: true, DockedRegionID: "capital", RegionID: "sea", Units: []army.Unit{{TypeID: "transport"}}},
			"army":  {ID: "army", OwnerID: "player", RegionID: "coast", Units: []army.Unit{{TypeID: "soldier", CurrentHP: 100}}},
		},
	}
	if got := gs.SupplyCargoCapacityForTurns("fleet", 5); got != 250 {
		t.Fatalf("5 turluk ikmal kapasitesi = %d, want 250", got)
	}

	ok, reason := gs.LoadSupplyCargoAtCapital("fleet", economy.ResourceCost{Grain: 20})
	if !ok || reason != "" {
		t.Fatalf("yükleme sonucu = (%v, %q)", ok, reason)
	}
	if got := gs.Factions["player"].Grain; got != 80 {
		t.Fatalf("yükleme sonrası tahıl = %d, want 80", got)
	}

	fleet := gs.Armies["fleet"]
	fleet.DockedRegionID = ""
	fleet.NavalMission = &army.NavalMission{Kind: army.NavalMissionSupplyArmy, TargetArmyID: "army"}
	if ok, reason := gs.CanAssignNavalMission("fleet", *fleet.NavalMission); !ok {
		t.Fatalf("ikmal görevi reddedildi: %s", reason)
	}

	fleet.DockedRegionID = "capital"
	if cargo, ok, reason := gs.UnloadSupplyCargoAtCapital("fleet"); !ok || cargo.Grain != 20 || reason != "" {
		t.Fatalf("boşaltma sonucu = (%+v, %v, %q)", cargo, ok, reason)
	}
	if fleet.NavalMission != nil {
		t.Fatal("ikmal yükü boşaltılınca ordu bağı temizlenmeli")
	}
	if got := gs.Factions["player"].Grain; got != 100 {
		t.Fatalf("boşaltma sonrası tahıl = %d, want 100", got)
	}
}

func TestConvertInvalidNavalBlockadesKeepsUnrelatedMissions(t *testing.T) {
	gs := navalMissionTransitionState()
	gs.Armies["patrol"] = &army.Army{
		ID:       "patrol",
		OwnerID:  "player",
		RegionID: "sea",
		IsNaval:  true,
		NavalMission: &army.NavalMission{
			Kind:           army.NavalMissionPatrol,
			TargetRegionID: "sea",
		},
	}
	gs.Regions["enemy_port"].OwnerID = "player"

	if got := gs.ConvertInvalidNavalBlockadesToPatrol(); got != 1 {
		t.Fatalf("dönüşüm sayısı = %d, want 1", got)
	}
	if got := gs.Armies["patrol"].NavalMission.Kind; got != army.NavalMissionPatrol {
		t.Fatalf("mevcut devriye görevi korunmalı, got %q", got)
	}
}

func navalMissionTransitionState() *GameState {
	return &GameState{
		Regions: map[world.RegionID]*world.Region{
			"enemy_port": {
				ID:          "enemy_port",
				OwnerID:     "enemy",
				Neighbors:   []world.RegionID{"sea"},
				Settlements: []world.Settlement{{ID: "enemy_harbor", Type: world.SettlementPort}},
			},
			"sea": {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"enemy_port"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"blockade": {
				ID:       "blockade",
				OwnerID:  "player",
				RegionID: "sea",
				IsNaval:  true,
				NavalMission: &army.NavalMission{
					Kind:           army.NavalMissionBlockade,
					TargetRegionID: "sea",
				},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {
				FactionA: "player",
				FactionB: "enemy",
				Stance:   faction.StanceWar,
			},
		},
	}
}
