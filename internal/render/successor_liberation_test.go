package render

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestLiberationButtonOnlyAppearsForEliminatedSuccessor(t *testing.T) {
	regionID := world.RegionID("former_capital")
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":    {ID: "player"},
			"successor": {ID: "successor", IsEliminated: true},
		},
		Regions: map[world.RegionID]*world.Region{
			regionID: {ID: regionID, OwnerID: "player", SuccessorFactionID: "successor"},
		},
	}

	barY := regionPanelActionBarY(gs, gs.Regions[regionID], regionPanelTabBuildings)
	button := buildRegionLiberateButton(infoPanelX()+float32(panelPad), float32(barY), infoPanelW-float32(panelPad*2), regionPanelActionBarHeight)
	if !regionLiberateButtonHitForTab(button.X+button.W/2, button.Y+button.H/2, gs, regionID, regionPanelTabBuildings) {
		t.Fatal("elenmiş ardıl devlet için özgürleştir düğmesi tıklanabilir olmalı")
	}

	gs.Factions["successor"].IsEliminated = false
	if regionLiberateButtonHitForTab(button.X+button.W/2, button.Y+button.H/2, gs, regionID, regionPanelTabBuildings) {
		t.Fatal("aktif ardıl devlet için özgürleştir düğmesi görünmemeli")
	}
}

func TestEditModeAssignsSuccessorWithUndoRedo(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"old": {ID: "old"},
			"new": {ID: "new"},
		},
		Regions: map[world.RegionID]*world.Region{
			"region": {ID: "region", OwnerID: "old", Terrain: world.TerrainPlain},
		},
	}
	r := New(gs)
	r.editSelectedRegion = "region"
	r.setSelectedRegionSuccessor("new")
	if got := gs.Regions["region"].SuccessorFactionID; got != "new" {
		t.Fatalf("ardıl ataması yapılmadı: %q", got)
	}
	r.undoEditCommand()
	if got := gs.Regions["region"].SuccessorFactionID; got != "" {
		t.Fatalf("ardıl ataması undo ile temizlenmedi: %q", got)
	}
	r.redoEditCommand()
	if got := gs.Regions["region"].SuccessorFactionID; got != "new" {
		t.Fatalf("ardıl ataması redo ile geri gelmedi: %q", got)
	}
}

func TestEditModeSettlementCapitalAssignsOwnerAsSuccessorWithUndoRedo(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"owner": {ID: "owner"},
		},
		Regions: map[world.RegionID]*world.Region{
			"region": {
				ID:                 "region",
				OwnerID:            "owner",
				SuccessorFactionID: "previous_successor",
				Terrain:            world.TerrainPlain,
				Settlements: []world.Settlement{
					{ID: "old_center", IsCenter: true},
					{ID: "new_center"},
				},
			},
		},
	}
	r := New(gs)
	r.editSelectedRegion = "region"
	r.editSelectedSettlement = 1
	r.setSelectedSettlementCapital()

	region := gs.Regions["region"]
	if !region.Settlements[1].IsCenter || region.Settlements[0].IsCenter {
		t.Fatal("seçili settlement bölgesel merkez yapılmadı")
	}
	if got := region.SuccessorFactionID; got != "owner" {
		t.Fatalf("bölge sahibi ardıl devlet olarak yazılmadı: %q", got)
	}

	r.undoEditCommand()
	if !region.Settlements[0].IsCenter || region.Settlements[1].IsCenter {
		t.Fatal("settlement başkent değişikliği undo ile geri alınmadı")
	}
	if got := region.SuccessorFactionID; got != "previous_successor" {
		t.Fatalf("ardıl devlet undo ile geri alınmadı: %q", got)
	}

	r.redoEditCommand()
	if got := region.SuccessorFactionID; got != "owner" {
		t.Fatalf("ardıl devlet redo ile geri gelmedi: %q", got)
	}
}
