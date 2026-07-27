package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestEditRegionAtAllowsSeaRegions(t *testing.T) {
	r := newSeaEditRenderer()
	sea := r.gs.Regions["sea_test"]
	sx, sy := r.worldToScreen(wcX(sea.WorldX), wcY(sea.WorldY))

	if got := r.editRegionAt(sx, sy); got != sea.ID {
		t.Fatalf("sea region secilemedi: got=%q want=%q", got, sea.ID)
	}
}

func TestAddRegionFromSourcePreservesSeaFlag(t *testing.T) {
	r := newSeaEditRenderer()
	r.addRegionFromSource("sea_test", 36, 38)

	if len(r.gs.Regions) != 2 {
		t.Fatalf("beklenen 2 region, got=%d", len(r.gs.Regions))
	}

	for rid, region := range r.gs.Regions {
		if rid == "sea_test" {
			continue
		}
		if !region.IsSea {
			t.Fatalf("yeni region deniz olmali: %+v", region)
		}
		if region.Terrain != world.TerrainSea {
			t.Fatalf("yeni deniz region terrain sea olmali: got=%q", region.Terrain)
		}
		return
	}

	t.Fatal("yeni region bulunamadi")
}

func TestMoveSelectedRegionCenterToAllowsSea(t *testing.T) {
	r := newSeaEditRenderer()
	r.editSelectedRegion = "sea_test"
	sx, sy := r.worldToScreen(wcX(40), wcY(42))

	r.moveSelectedRegionCenterTo(sx, sy)

	sea := r.gs.Regions["sea_test"]
	if sea.WorldX != 40 || sea.WorldY != 42 {
		t.Fatalf("sea center tasinmadi: got=(%d,%d)", sea.WorldX, sea.WorldY)
	}
}

func TestRenameRegionIDUpdatesEditorReferences(t *testing.T) {
	worldW := 64
	worldH := 64
	offset := 0.0
	scale := 1.0
	gs := &state.GameState{
		MapConfig: scenario.MapConfig{
			WorldWidth:   &worldW,
			WorldHeight:  &worldH,
			ShapeOffsetX: &offset,
			ShapeOffsetY: &offset,
			ShapeScaleX:  &scale,
			ShapeScaleY:  &scale,
		},
		Regions: map[world.RegionID]*world.Region{
			"old_region": {ID: "old_region", NameTR: "Eski", WorldX: 20, WorldY: 20, Neighbors: []world.RegionID{"neighbor"}},
			"neighbor":   {ID: "neighbor", NameTR: "Komşu", WorldX: 40, WorldY: 40, Neighbors: []world.RegionID{"old_region"}},
		},
		RegionOrder:  []world.RegionID{"old_region", "neighbor"},
		LandPassages: []world.LandPassage{{From: "old_region", To: "neighbor"}},
		Armies: map[army.ArmyID]*army.Army{
			"land":  {ID: "land", RegionID: "old_region"},
			"fleet": {ID: "fleet", RegionID: "neighbor", DockedRegionID: "old_region", IsNaval: true},
		},
		RegionPaintOverrides: map[int]world.RegionID{12: "old_region"},
	}
	r := New(gs)
	r.editSelectedRegion = "old_region"
	r.SelectedRegion = "old_region"
	r.renameRegionID("old_region", "correct_region")

	if gs.Regions["old_region"] != nil || gs.Regions["correct_region"] == nil {
		t.Fatal("region map anahtarı güncellenmedi")
	}
	if gs.Regions["correct_region"].ID != "correct_region" {
		t.Fatalf("region.ID güncellenmedi: %q", gs.Regions["correct_region"].ID)
	}
	if gs.Regions["neighbor"].Neighbors[0] != "correct_region" || gs.Regions["correct_region"].Neighbors[0] != "neighbor" {
		t.Fatalf("komşu referansları güncellenmedi: %#v / %#v", gs.Regions["neighbor"].Neighbors, gs.Regions["correct_region"].Neighbors)
	}
	if gs.RegionOrder[0] != "correct_region" || gs.LandPassages[0].From != "correct_region" {
		t.Fatalf("statik bölge referansları güncellenmedi")
	}
	if gs.Armies["land"].RegionID != "correct_region" || gs.Armies["fleet"].DockedRegionID != "correct_region" {
		t.Fatalf("ordu bölge referansları güncellenmedi")
	}
	if gs.RegionPaintOverrides[12] != "correct_region" || r.editSelectedRegion != "correct_region" || r.SelectedRegion != "correct_region" {
		t.Fatalf("editör seçim/paint referansları güncellenmedi")
	}
}

func TestEditRegionIDButtonStartsRename(t *testing.T) {
	r := newSeaEditRenderer()
	r.editSelectedRegion = "sea_test"
	rect := editInspectorButtonRect(editButtonRegionID)
	if got := editMapInspectorButtonAt(rect[0]+rect[2]/2, rect[1]+rect[3]/2); got != editButtonRegionID {
		t.Fatalf("ID butonu hit-test edilmedi: got=%v want=%v rect=%v", got, editButtonRegionID, rect)
	}
	if _, handled := r.handleEditInspectorClick(rect[0]+rect[2]/2, rect[1]+rect[3]/2); !handled {
		t.Fatal("ID butonu tıklaması işlenmedi")
	}
	if !r.editRenaming || r.editTextTarget != editTextRegionID || string(r.editTextRunes) != "sea_test" {
		t.Fatalf("ID düzenleme başlamadı: renaming=%v target=%v text=%q", r.editRenaming, r.editTextTarget, string(r.editTextRunes))
	}
}

func newSeaEditRenderer() *Renderer {
	worldW := 64
	worldH := 64
	offset := 0.0
	scale := 1.0
	gs := &state.GameState{
		MapConfig: scenario.MapConfig{
			WorldWidth:   &worldW,
			WorldHeight:  &worldH,
			ShapeOffsetX: &offset,
			ShapeOffsetY: &offset,
			ShapeScaleX:  &scale,
			ShapeScaleY:  &scale,
		},
		Regions: map[world.RegionID]*world.Region{
			"sea_test": {
				ID:      "sea_test",
				Name:    "Sea Test",
				NameTR:  "Deniz Test",
				Terrain: world.TerrainSea,
				WorldX:  20,
				WorldY:  20,
				ShapeID: "sea_shape",
				IsSea:   true,
			},
		},
		RegionOrder: []world.RegionID{"sea_test"},
	}
	return New(gs)
}
