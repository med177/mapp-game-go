package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestSplitFactionForcesUsesRegionsAndFallbackForFleets(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"england":         {},
			"house_lancaster": {},
			"house_york":      {},
		},
		Armies: map[army.ArmyID]*army.Army{
			"lancashire_army": {ID: "lancashire_army", OwnerID: "england", RegionID: "lancashire"},
			"yorkshire_army":  {ID: "yorkshire_army", OwnerID: "england", RegionID: "yorkshire"},
			"channel_fleet":   {ID: "channel_fleet", OwnerID: "england", RegionID: "channel", IsNaval: true},
		},
		ArmyOrder: []army.ArmyID{"lancashire_army", "yorkshire_army", "channel_fleet"},
	}
	gs.Armies["lancashire_army"].Commander = army.NewCommander("lancaster_commander", "Lancaster Komutanı")
	gs.Armies["lancashire_army"].Commander.OwnerID = "england"

	report := gs.SplitFactionForces(
		[]faction.FactionID{"england"},
		[]faction.FactionID{"house_lancaster", "house_york"},
		map[world.RegionID]faction.FactionID{
			"lancashire": "house_lancaster",
			"yorkshire":  "house_york",
		},
	)
	if report.TransferredArmies != 2 || report.TransferredFleets != 1 {
		t.Fatalf("kuvvet raporu beklenmedik: %+v", report)
	}
	if gs.Armies["lancashire_army"].OwnerID != "house_lancaster" || gs.Armies["yorkshire_army"].OwnerID != "house_york" {
		t.Fatalf("kara orduları bölgelere göre dağılmadı: %+v", gs.Armies)
	}
	if got := gs.Armies["lancashire_army"].Commander.OwnerID; got != "house_lancaster" {
		t.Fatalf("komutan sahipliği aktarılmadı: %s", got)
	}
	if got := gs.Armies["channel_fleet"].OwnerID; got != "house_lancaster" {
		t.Fatalf("açık deniz filosu deterministik dağıtılmadı: %s", got)
	}
}

func TestMergeFactionForcesPreservesArmyObjectsAndUpdatesCommanders(t *testing.T) {
	commander := army.NewCommander("york_commander", "York Komutanı")
	commander.OwnerID = "house_york"
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{"england": {}, "house_york": {}},
		Armies: map[army.ArmyID]*army.Army{
			"york_army": {ID: "york_army", OwnerID: "house_york", RegionID: "yorkshire", Commander: commander},
		},
	}

	report := gs.MergeFactionForces([]faction.FactionID{"house_york"}, "england")
	if report.TransferredArmies != 1 || gs.Armies["york_army"].OwnerID != "england" {
		t.Fatalf("ordu birleşmesi beklenmedik: report=%+v army=%+v", report, gs.Armies["york_army"])
	}
	if commander.OwnerID != "england" || commander.AssignedArmyID != "york_army" {
		t.Fatalf("komutan birleşme sonrası tutarsız: %+v", commander)
	}
}
