package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestCollectActiveWarSummariesShowsTurnsStrengthAndArmyCounts(t *testing.T) {
	gs := &state.GameState{
		Turn: 9,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", NameTR: "A Devleti"},
			"b": {ID: "b", NameTR: "B Devleti"},
		},
		Regions: map[world.RegionID]*world.Region{
			"a-region": {ID: "a-region", OwnerID: "a"},
			"b-region": {ID: "b-region", OwnerID: "b"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a-army": {ID: "a-army", OwnerID: "a", Units: []army.Unit{{TypeID: "inf"}, {TypeID: "inf"}}},
			"b-army": {ID: "b-army", OwnerID: "b", Units: []army.Unit{{TypeID: "inf"}}, EmbarkedUnits: []army.Unit{{TypeID: "inf"}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("b", "a"): {FactionA: "b", FactionB: "a", Stance: faction.StanceWar},
		},
		WarLedgers: map[string]*state.WarLedger{
			faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b", StartedTurn: 4, CasualtiesA: 3, CasualtiesB: 5},
		},
	}

	wars := collectActiveWarSummaries(gs, nil)
	if len(wars) != 1 {
		t.Fatalf("tek aktif savaş bekleniyordu, got=%d", len(wars))
	}
	war := wars[0]
	if war.FactionANameTR != "A Devleti" || war.FactionBNameTR != "B Devleti" {
		t.Fatalf("taraf adları yanlış: %+v", war)
	}
	if war.Turns != 5 || war.ArmiesA != 1 || war.ArmiesB != 1 || war.UnitsA != 2 || war.UnitsB != 2 {
		t.Fatalf("savaş süresi/ordu sayıları yanlış: %+v", war)
	}
	if war.CasualtiesA != 3 || war.CasualtiesB != 5 {
		t.Fatalf("kayıplar yanlış: %+v", war)
	}
}

func TestActiveWarsPanelDoesNotCoverWarHUDButton(t *testing.T) {
	button := buildActiveWarsHUDButton()
	panel := activeWarsPanelRect()
	if panel.Hit(button.X+button.W/2, button.Y+button.H/2) {
		t.Fatal("aktif savaş paneli açılmadan önce HUD savaş düğmesinin üzerine taşmamalı")
	}
	if !activeWarsHudButtonHit(button.X+button.W/2, button.Y+button.H/2) {
		t.Fatal("aktif savaş HUD düğmesi kendi merkezinde tıklanabilir olmalı")
	}
}

func TestActiveWarsPanelLeavesOutsideMapPointAvailable(t *testing.T) {
	panel := activeWarsPanelRect()
	mapX := panel.X - 24
	mapY := panel.Y + 120
	if activeWarsPanelHit(mapX, mapY) {
		t.Fatal("panel dışındaki harita noktası aktif savaş paneli tarafından tüketilmemeli")
	}
}
