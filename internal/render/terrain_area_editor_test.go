package render

import (
	"testing"

	"mapp-game-go/internal/world"
)

func newTerrainAreaEditRenderer() *Renderer {
	r := newLandShapeEditRenderer()
	r.gs.TerrainAreas = []world.TerrainArea{
		{ID: "area_1", ParentRegionID: "land_test", MoveCost: -1, Cells: [][2]int{{14, 10}}},
	}
	// land_test bölgesinin bu pikseli daha önce açıkça boyanmış gibi davran;
	// arazi alanı senkronu bu override'ı ezmeden en üst katman olmalı.
	r.gs.RegionPaintOverrides = map[int]world.RegionID{10*WorldW + 14: "land_test"}
	r.rebuildEditWorldMap()
	return r
}

func TestTerrainAreaRemainsSelectableOverRegionPaintOverride(t *testing.T) {
	r := newTerrainAreaEditRenderer()
	if got := r.worldMap.RegionAt(14, 10); got != "terrain_area::area_1" {
		t.Fatalf("arazi alanı hücresi seçilebilir olmalı, got=%q", got)
	}
	sx, sy := r.worldToScreen(wcX(14), wcY(10))
	if got := r.editRegionAt(sx, sy); got != "terrain_area::area_1" {
		t.Fatalf("ekran koordinatından arazi alanı seçilemedi: got=%q", got)
	}
}

func TestCycleEditTerrainAreaCostUpdatesSelectedArea(t *testing.T) {
	r := newTerrainAreaEditRenderer()
	r.editSelectedRegion = "terrain_area::area_1"
	r.syncSelectedTerrainArea(r.editSelectedRegion)
	if r.editTerrainAreaMoveCost != -1 {
		t.Fatalf("seçim sonrası maliyet senkronize olmalı: got=%d", r.editTerrainAreaMoveCost)
	}
	r.cycleEditTerrainAreaCost()
	if r.gs.TerrainAreas[0].MoveCost != -2 {
		t.Fatalf("seçili alanın maliyeti güncellenmedi: got=%d", r.gs.TerrainAreas[0].MoveCost)
	}
}

func TestCycleEditTerrainAreaAttritionUpdatesSelectedArea(t *testing.T) {
	r := newTerrainAreaEditRenderer()
	r.editSelectedRegion = "terrain_area::area_1"
	r.syncSelectedTerrainArea(r.editSelectedRegion)
	if r.editTerrainAreaAttritionCost != 0 {
		t.Fatalf("seçim sonrası yıpranma senkronize olmalı: got=%d", r.editTerrainAreaAttritionCost)
	}
	r.cycleEditTerrainAreaAttrition()
	if r.gs.TerrainAreas[0].AttritionCost != 5 {
		t.Fatalf("seçili alanın yıpranması güncellenmedi: got=%d", r.gs.TerrainAreas[0].AttritionCost)
	}
	r.editSelectedRegion = ""
	r.editTerrainAreaSelected = -1
	r.editTerrainAreaAttritionCost = 20
	r.cycleEditTerrainAreaAttrition()
	if r.editTerrainAreaAttritionCost != 0 {
		t.Fatalf("seçim yokken varsayılan yıpranma 20'den sonra sıfıra dönmeli: got=%d", r.editTerrainAreaAttritionCost)
	}
}

func TestAddNeighborBetweenTerrainAreaAndRegionPersistsAfterResync(t *testing.T) {
	r := newTerrainAreaEditRenderer()
	r.gs.Regions["other"] = &world.Region{ID: "other", WorldX: 40, WorldY: 40}
	r.gs.RegionOrder = append(r.gs.RegionOrder, "other")

	r.addNeighborBetween("terrain_area::area_1", "other")

	if !regionHasNeighbor(r.gs.Regions["other"], "terrain_area::area_1") {
		t.Fatal("normal bölgeye arazi alanı komşuluğu eklenmedi")
	}
	if len(r.gs.TerrainAreas[0].ExtraNeighbors) != 1 || r.gs.TerrainAreas[0].ExtraNeighbors[0] != "other" {
		t.Fatalf("arazi alanının ExtraNeighbors listesi güncellenmedi: got=%v", r.gs.TerrainAreas[0].ExtraNeighbors)
	}

	// Yeniden senkronizasyon (ör. başka bir düzenleme sonrası) komşuluğu korumalı.
	world.SyncTerrainAreaRegions(r.gs.Regions, r.gs.TerrainAreas)
	if !regionHasNeighbor(r.gs.Regions["terrain_area::area_1"], "other") {
		t.Fatal("yeniden senkron sonrası arazi alanı komşuluğu kayboldu")
	}
	if !regionHasNeighbor(r.gs.Regions["other"], "terrain_area::area_1") {
		t.Fatal("yeniden senkron sonrası ters yön komşuluk kayboldu")
	}
}

func TestTerrainAreaPaintModeDisablesLandPassageButtons(t *testing.T) {
	r := newTerrainAreaEditRenderer()
	r.editSelectedRegion = "land_test"
	r.toggleEditTerrainAreaMode()
	if !r.editTerrainAreaMode {
		t.Fatal("arazi alanı boyama modu açılmadı")
	}

	addRect := editInspectorButtonRect(editButtonLandPassageAdd)
	if _, handled := r.handleEditShapeInspectorClick(addRect[0]+addRect[2]/2, addRect[1]+addRect[3]/2); !handled {
		t.Fatal("tıklama tüketilmeli")
	}
	if r.editLandPassageMode {
		t.Fatal("arazi boyama açıkken geçiş ekleme aktifleşmemeli")
	}

	adjustRect := editInspectorButtonRect(editButtonLandPassageAdjust)
	r.handleEditShapeInspectorClick(adjustRect[0]+adjustRect[2]/2, adjustRect[1]+adjustRect[3]/2)
	if r.editLandPassageAdjustMode {
		t.Fatal("arazi boyama açıkken geçiş düzenleme aktifleşmemeli")
	}
}

func TestLandPassageModeDisablesTerrainAreaButton(t *testing.T) {
	r := newTerrainAreaEditRenderer()
	r.editSelectedRegion = "land_test"
	r.toggleEditLandPassageMode()
	if !r.editLandPassageMode {
		t.Fatal("geçiş ekleme modu açılmadı")
	}

	areaRect := editInspectorButtonRect(editButtonTerrainArea)
	r.handleEditShapeInspectorClick(areaRect[0]+areaRect[2]/2, areaRect[1]+areaRect[3]/2)
	if r.editTerrainAreaMode {
		t.Fatal("geçiş ekleme açıkken arazi boyama aktifleşmemeli")
	}
}
