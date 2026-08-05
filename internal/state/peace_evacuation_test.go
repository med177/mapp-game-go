package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func navalLandingPeaceState(withFleet bool) *GameState {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"target":   {ID: "target", OwnerID: "b", WorldX: 100, WorldY: 0},
			"home":     {ID: "home", OwnerID: "a", WorldX: 0, WorldY: 100},
			"nearland": {ID: "nearland", OwnerID: "a", WorldX: 110, WorldY: 0},
			"farland":  {ID: "farland", OwnerID: "a", WorldX: 1000, WorldY: 0},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a"},
			"b": {ID: "b"},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":       {ID: "inf", Embarkable: true},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10},
		},
		Armies: map[army.ArmyID]*army.Army{
			"landing": {
				ID: "landing", OwnerID: "a", RegionID: "target",
				Units: []army.Unit{{TypeID: "inf"}, {TypeID: "inf"}},
			},
		},
		Sieges: map[world.RegionID]*SiegeState{
			"target": {RegionID: "target", AttackerArmyID: "landing", AttackerFactionID: "a", NavalLanding: true},
		},
	}
	if withFleet {
		gs.Regions["nearsea"] = &world.Region{ID: "nearsea", IsSea: true, WorldX: 100, WorldY: 20}
		gs.Regions["farsea"] = &world.Region{ID: "farsea", IsSea: true, WorldX: 0, WorldY: 1000}
		gs.Armies["near-fleet"] = &army.Army{
			ID: "near-fleet", OwnerID: "a", RegionID: "nearsea", IsNaval: true,
			Units: []army.Unit{{TypeID: "transport"}},
		}
		gs.Armies["far-fleet"] = &army.Army{
			ID: "far-fleet", OwnerID: "a", RegionID: "farsea", IsNaval: true,
			Units: []army.Unit{{TypeID: "transport"}},
		}
	}
	return gs
}

func TestEvacuateNavalLandingSiegeReembarksAtNearestTransportFleet(t *testing.T) {
	gs := navalLandingPeaceState(true)
	commander := army.NewCommander("landing-commander", "Çıkarma Komutanı")
	gs.Armies["landing"].Commander = commander
	commander.AssignedArmyID = "landing"

	if got := gs.EvacuateNavalLandingSiegesAfterPeace([]faction.FactionID{"a"}, []faction.FactionID{"b"}); got != 1 {
		t.Fatalf("tek kuşatma tahliye edilmeli, got=%d", got)
	}
	if _, ok := gs.Armies["landing"]; ok {
		t.Fatal("yeniden gemiye binen kara ordusu state'ten kaldırılmalı")
	}
	nearFleet := gs.Armies["near-fleet"]
	if len(nearFleet.EmbarkedUnits) != 2 || nearFleet.EmbarkedCommander != commander {
		t.Fatalf("en yakın nakliye filosuna birlik ve kara komutanı aktarılmalıydı: %+v", nearFleet)
	}
	if len(gs.Armies["far-fleet"].EmbarkedUnits) != 0 {
		t.Fatal("uzak nakliye filosu tercih edilmemeliydi")
	}
	if gs.SiegeAt("target") != nil {
		t.Fatal("barış sonrası denizden kurulan kuşatma kaldırılmalı")
	}
}

func TestEvacuateNavalLandingSiegeRetreatsToNearestOwnedRegionWithoutFleet(t *testing.T) {
	gs := navalLandingPeaceState(false)

	if got := gs.EvacuateNavalLandingSiegesAfterPeace([]faction.FactionID{"a"}, []faction.FactionID{"b"}); got != 1 {
		t.Fatalf("tek kuşatma geri çekilmeli, got=%d", got)
	}
	if army := gs.Armies["landing"]; army == nil || army.RegionID != "nearland" {
		t.Fatalf("nakliye yoksa ordu en yakın kendi bölgesine çekilmeli: %+v", army)
	}
	if gs.SiegeAt("target") != nil {
		t.Fatal("geri çekilmede kuşatma kaldırılmalı")
	}
}

func TestEvacuateArmiesFromPeaceTerritoryUsesNearestAlliedLand(t *testing.T) {
	gs := navalLandingPeaceState(false)
	gs.Regions["target"].WorldX = 100
	gs.Regions["target"].WorldY = 0
	gs.Regions["home"].WorldX = 0
	gs.Regions["home"].WorldY = 0
	gs.Regions["ally"] = &world.Region{ID: "ally", OwnerID: "ally", WorldX: 90, WorldY: 0}
	gs.Factions["ally"] = &faction.Faction{ID: "ally"}
	gs.Relations = map[string]*faction.Relation{
		faction.RelationKey("a", "ally"): {FactionA: "a", FactionB: "ally", Stance: faction.StanceAllied},
	}
	gs.Armies["landing"].MovePoints = 0

	if got := gs.EvacuateArmiesFromPeaceTerritory([]faction.FactionID{"a"}, []faction.FactionID{"b"}); got != 1 {
		t.Fatalf("tek düşman toprağındaki ordu tahliye edilmeli, got=%d", got)
	}
	if got := gs.Armies["landing"].RegionID; got != "ally" {
		t.Fatalf("en yakın müttefik bölge seçilmeli: %s", got)
	}
}

func TestEvacuateArmiesWithoutLandAccessRepairsStalePeaceOccupation(t *testing.T) {
	gs := navalLandingPeaceState(false)
	gs.Regions["target"].WorldX = 100
	gs.Regions["target"].WorldY = 0
	gs.Regions["nearland"].WorldX = 90
	gs.Regions["nearland"].WorldY = 0
	gs.Armies["landing"].MovePoints = 0
	gs.Sieges["target"].NavalLanding = false

	if got := gs.EvacuateArmiesWithoutLandAccess(); got != 1 {
		t.Fatalf("izinsiz eski-save ordusu tahliye edilmeli, got=%d", got)
	}
	if got := gs.Armies["landing"].RegionID; got != "nearland" {
		t.Fatalf("ordu en yakın güvenli bölgeye çekilmeli: %s", got)
	}
	if gs.SiegeAt("target") != nil {
		t.Fatal("tahliye edilen ordunun kuşatması temizlenmeli")
	}
}

func TestEvacuateArmiesWithoutLandAccessKeepsWarAndAlliedTransit(t *testing.T) {
	for _, stance := range []faction.DiplomaticStance{faction.StanceWar, faction.StanceAllied} {
		t.Run(string(stance), func(t *testing.T) {
			gs := navalLandingPeaceState(false)
			gs.Relations = map[string]*faction.Relation{
				faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b", Stance: stance},
			}

			if got := gs.EvacuateArmiesWithoutLandAccess(); got != 0 {
				t.Fatalf("%s durumunda geçerli ordular çekilmemeli, got=%d", stance, got)
			}
			if got := gs.Armies["landing"].RegionID; got != "target" {
				t.Fatalf("%s durumunda ordu mevcut bölgede kalmalı: %s", stance, got)
			}
		})
	}
}

func TestEvacuateArmiesWithoutLandAccessKeepsNeutralOccupation(t *testing.T) {
	gs := navalLandingPeaceState(false)
	gs.Regions["target"].OwnerID = ""

	if got := gs.EvacuateArmiesWithoutLandAccess(); got != 0 {
		t.Fatalf("sahipsiz bölge yabancı toprak sayılmamalı, got=%d", got)
	}
	if got := gs.Armies["landing"].RegionID; got != "target" {
		t.Fatalf("sahipsiz bölgedeki ordu geri çekilmemeli: %s", got)
	}
}
