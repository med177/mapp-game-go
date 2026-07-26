package render

import (
	"testing"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestEditShapeInspectorLandPassageButtonsSwitchTools(t *testing.T) {
	r := New(&state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"a": {ID: "a"},
			"b": {ID: "b"},
		},
	})

	addRect := editInspectorButtonRect(editButtonLandPassageAdd)
	if _, handled := r.handleEditShapeInspectorClick(addRect[0]+addRect[2]/2, addRect[1]+addRect[3]/2); !handled {
		t.Fatal("geçiş ekle butonu tıklanmadı")
	}
	if !r.editLandPassageMode || r.editLandPassageAdjustMode {
		t.Fatalf("ekleme modu açılmadı: add=%v adjust=%v", r.editLandPassageMode, r.editLandPassageAdjustMode)
	}

	adjustRect := editInspectorButtonRect(editButtonLandPassageAdjust)
	if _, handled := r.handleEditShapeInspectorClick(adjustRect[0]+adjustRect[2]/2, adjustRect[1]+adjustRect[3]/2); !handled {
		t.Fatal("geçiş düzenle butonu tıklanmadı")
	}
	if r.editLandPassageMode || !r.editLandPassageAdjustMode {
		t.Fatalf("düzenleme modu açılmadı: add=%v adjust=%v", r.editLandPassageMode, r.editLandPassageAdjustMode)
	}
}

func TestEditLandPassageDeleteButtonAndUndo(t *testing.T) {
	start := [2]int{10, 10}
	end := [2]int{30, 10}
	r := New(&state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"a": {ID: "a"},
			"b": {ID: "b"},
		},
		LandPassages: []world.LandPassage{{From: "a", To: "b", Start: &start, End: &end}},
	})
	r.editLandPassageSelected = 0
	r.editLandPassageAdjustMode = true

	deleteRect := editInspectorButtonRect(editButtonLandPassageDelete)
	if _, handled := r.handleEditShapeInspectorClick(deleteRect[0]+deleteRect[2]/2, deleteRect[1]+deleteRect[3]/2); !handled {
		t.Fatal("geçiş sil butonu tıklanmadı")
	}
	if len(r.gs.LandPassages) != 0 {
		t.Fatalf("geçiş silinmedi: %d", len(r.gs.LandPassages))
	}
	r.undoEditCommand()
	if len(r.gs.LandPassages) != 1 {
		t.Fatal("silinen geçiş undo ile geri gelmedi")
	}
	r.redoEditCommand()
	if len(r.gs.LandPassages) != 0 {
		t.Fatal("geçiş silme redo ile tekrarlanmadı")
	}
}

func TestEditShapeInspectorAddNeighborCreatesSymmetricUndoableLink(t *testing.T) {
	r := New(&state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"a": {ID: "a"},
			"b": {ID: "b"},
		},
	})
	r.editSelectedRegion = "a"
	addRect := editInspectorButtonRect(editButtonAddNeighbor)
	if _, handled := r.handleEditShapeInspectorClick(addRect[0]+addRect[2]/2, addRect[1]+addRect[3]/2); !handled {
		t.Fatal("komşu ekle butonu tıklanmadı")
	}
	if !r.editNeighborAddMode || r.editNeighborAddFrom != "a" {
		t.Fatalf("komşu ekleme modu açılmadı: mode=%v from=%q", r.editNeighborAddMode, r.editNeighborAddFrom)
	}

	r.addNeighborBetween("a", "b")
	if !regionHasNeighbor(r.gs.Regions["a"], "b") || !regionHasNeighbor(r.gs.Regions["b"], "a") {
		t.Fatal("komşuluk iki yönlü eklenmedi")
	}
	r.undoEditCommand()
	if regionHasNeighbor(r.gs.Regions["a"], "b") || regionHasNeighbor(r.gs.Regions["b"], "a") {
		t.Fatal("komşuluk undo ile kaldırılmadı")
	}
	r.redoEditCommand()
	if !regionHasNeighbor(r.gs.Regions["a"], "b") || !regionHasNeighbor(r.gs.Regions["b"], "a") {
		t.Fatal("komşuluk redo ile geri gelmedi")
	}
}

func TestEditLandPassageDragUpdatesEndpointAndSupportsUndo(t *testing.T) {
	start := [2]int{10, 10}
	originalStart := start
	end := [2]int{30, 10}
	r := New(&state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"a": {ID: "a", WorldX: 10, WorldY: 10},
			"b": {ID: "b", WorldX: 30, WorldY: 10},
		},
		LandPassages: []world.LandPassage{{
			From: "a", To: "b", Type: world.LandPassageStrait,
			MoveCost: 1, DefenseBonus: 15, Start: &start, End: &end,
		}},
	})
	r.editLandPassageAdjustMode = true

	startX, startY := r.worldToScreen(wcX(start[0]), wcY(start[1]))
	r.handleEditLandPassageAdjustClick(startX, startY)
	if r.editLandPassageSelected != 0 || r.editLandPassageDragEndpoint != 0 {
		t.Fatalf("başlangıç ucu seçilmedi: passage=%d endpoint=%d", r.editLandPassageSelected, r.editLandPassageDragEndpoint)
	}

	movedX, movedY := r.worldToScreen(wcX(12), wcY(10))
	r.updateEditLandPassageDrag(movedX, movedY)
	r.finishEditLandPassageDrag()
	if got := *r.gs.LandPassages[0].Start; got != [2]int{12, 10} {
		t.Fatalf("uç noktası taşınmadı: got=%v", got)
	}
	if !r.editDirty || len(r.editUndoStack) == 0 {
		t.Fatal("uç noktası düzenlemesi edit geçmişine eklenmedi")
	}

	r.undoEditCommand()
	if got := *r.gs.LandPassages[0].Start; got != originalStart {
		t.Fatalf("undo başlangıç ucunu geri almadı: got=%v", got)
	}
}
