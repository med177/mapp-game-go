package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
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
		AIStrategies: map[string]scenario.AIFactionStrategy{
			"player": {
				FactionID: "player",
				Objectives: []scenario.AIObjectiveDef{{
					TargetRegions:    []string{"old_region"},
					ReadinessRegions: []string{"old_region"},
					AnnexRegionIDs:   []string{"old_region"},
				}},
			},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{
			{ID: "old_region", Links: []world.RegionID{"neighbor"}},
			{ID: "neighbor", Links: []world.RegionID{"old_region"}},
		}},
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
	objective := gs.AIStrategies["player"].Objectives[0]
	if objective.TargetRegions[0] != "correct_region" || objective.ReadinessRegions[0] != "correct_region" || objective.AnnexRegionIDs[0] != "correct_region" {
		t.Fatalf("AI strateji bölge referansları güncellenmedi: %+v", objective)
	}
	if gs.TradeCenters.Centers[0].ID != "correct_region" || gs.TradeCenters.Centers[0].Links[0] != "neighbor" || gs.TradeCenters.Centers[1].Links[0] != "correct_region" {
		t.Fatalf("ticaret merkezi referansları güncellenmedi: %+v", gs.TradeCenters.Centers)
	}
	if gs.RegionPaintOverrides[12] != "correct_region" || r.editSelectedRegion != "correct_region" || r.SelectedRegion != "correct_region" {
		t.Fatalf("editör seçim/paint referansları güncellenmedi")
	}
}

func TestTradeCenterVisualFollowsEditedRegionMetadata(t *testing.T) {
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
			"trade_region": {ID: "trade_region", NameTR: "Eski Ad", WorldX: 20, WorldY: 20, TradeCapacity: 10},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{{ID: "trade_region"}}},
	}
	r := New(gs)
	visuals := r.buildTradeCenters(1)
	if len(visuals) != 1 || visuals[0].nameTR != "Eski Ad" || visuals[0].worldX != 20 || visuals[0].worldY != 20 {
		t.Fatalf("başlangıç ticaret merkezi metadata'sı beklenmiyordu: %+v", visuals)
	}

	gs.Regions["trade_region"].NameTR = "Yeni Ad"
	gs.Regions["trade_region"].WorldX = 31
	gs.Regions["trade_region"].WorldY = 33
	visuals = r.buildTradeCenters(1)
	if len(visuals) != 1 || visuals[0].nameTR != "Yeni Ad" || visuals[0].worldX != 31 || visuals[0].worldY != 33 {
		t.Fatalf("edit sonrası ticaret merkezi güncel bölge metadata'sını kullanmıyor: %+v", visuals)
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

func TestEditModeSetsSelectedSettlementAsFactionCapital(t *testing.T) {
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
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {
				ID:                         "player",
				CapitalSettlementID:        "old_capital",
				PendingCapitalSettlementID: "new_capital",
				PendingCapitalTurns:        3,
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID:      "home",
				OwnerID: "player",
				WorldX:  20,
				WorldY:  20,
				Settlements: []world.Settlement{
					{ID: "old_capital", NameTR: "Eski Başkent"},
					{ID: "new_capital", NameTR: "Yeni Başkent"},
				},
			},
		},
	}
	r := New(gs)
	r.editSelectedRegion = "home"
	r.editSelectedSettlement = 1

	rect := editInspectorButtonRect(editButtonSetFactionCapital)
	if got := editMapInspectorButtonAt(rect[0]+rect[2]/2, rect[1]+rect[3]/2); got != editButtonSetFactionCapital {
		t.Fatalf("ulusal başkent butonu hit-test edilmedi: got=%v want=%v", got, editButtonSetFactionCapital)
	}
	if _, handled := r.handleEditInspectorClick(rect[0]+rect[2]/2, rect[1]+rect[3]/2); !handled {
		t.Fatal("ulusal başkent butonu tıklaması işlenmedi")
	}

	if got := gs.Factions["player"].CapitalSettlementID; got != "new_capital" {
		t.Fatalf("ulusal başkent güncellenmedi: got=%q", got)
	}
	if gs.Factions["player"].PendingCapitalSettlementID != "" || gs.Factions["player"].PendingCapitalTurns != 0 {
		t.Fatalf("eski başkent taşıma kuyruğu temizlenmedi: %+v", gs.Factions["player"])
	}

	r.undoEditCommand()
	if got := r.gs.Factions["player"].CapitalSettlementID; got != "old_capital" ||
		r.gs.Factions["player"].PendingCapitalSettlementID != "new_capital" ||
		r.gs.Factions["player"].PendingCapitalTurns != 3 {
		t.Fatalf("ulusal başkent undo ile geri alınmadı: %+v", r.gs.Factions["player"])
	}
	r.redoEditCommand()
	if got := r.gs.Factions["player"].CapitalSettlementID; got != "new_capital" {
		t.Fatalf("ulusal başkent redo ile geri uygulanmadı: got=%q", got)
	}
}

func TestEditArmyActionsUseSettlementInspector(t *testing.T) {
	actions := [...]editInspectorButton{
		editButtonAddArmy,
		editButtonAddFleet,
		editButtonDeleteArmy,
		editButtonArmyUnitType,
		editButtonArmyUnitMinus,
		editButtonArmyUnitPlus,
		editButtonArmyOwnerFromRegion,
	}
	for _, want := range actions {
		rect := editInspectorButtonRect(want)
		mx, my := rect[0]+rect[2]/2, rect[1]+rect[3]/2
		if got := editSettlementInspectorButtonAt(mx, my); got != want {
			t.Fatalf("%v Yerleşim Birimi sekmesinde bulunamadı: got=%v rect=%v", want, got, rect)
		}
		if got := editFactionInspectorButtonAt(mx, my); got == want {
			t.Fatalf("%v Devlet sekmesinde hâlâ aktif: rect=%v", want, rect)
		}
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
