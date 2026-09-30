package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestCursorShapeAtUsesSharedEditModeUILayer(t *testing.T) {
	originalHeight := ScreenHeight
	defer func() { ScreenHeight = originalHeight }()
	ScreenHeight = 900

	r := &Renderer{
		gs:               &state.GameState{Phase: state.PhaseEditMode},
		editInspectorTab: editInspectorData,
	}
	r.rebuildUILayers()

	tab := buildEditInspectorTabButton(editInspectorData, "")
	if got := r.cursorShapeAt(tab.X+tab.W/2, tab.Y+tab.H/2); got != ebiten.CursorShapePointer {
		t.Fatalf("Edit Mode sekmesi cursor şekli = %v, want pointer", got)
	}

	x, y, _, _ := editInspectorRect()
	if got := r.cursorShapeAt(float64(x)+10, float64(y)+60); got != ebiten.CursorShapeDefault {
		t.Fatalf("Edit Mode inspector boş alanı cursor şekli = %v, want default", got)
	}

	if got := r.cursorShapeAt(float64(ScreenWidth-10), float64(ScreenHeight/2)); got != ebiten.CursorShapeDefault {
		t.Fatalf("Edit Mode harita alanı cursor şekli = %v, want default", got)
	}
}

func TestInGameHoveringIgnoresHiddenRecruitPanelCards(t *testing.T) {
	rid := world.RegionID("owned")
	gs := &state.GameState{
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			rid: {ID: rid, OwnerID: "player"},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia":  {ID: "militia"},
			"infantry": {ID: "infantry"},
			"cavalry":  {ID: "cavalry"},
		},
		UnitTypeOrder: []string{"militia", "infantry", "cavalry"},
	}
	r := &Renderer{gs: gs, SelectedRegion: rid, mapMode: MapModeNormal}
	buttons := buildRecruitUnitCardButtons(gs, rid)
	if len(buttons) < 3 {
		t.Fatalf("test kartları oluşturulamadı: %d", len(buttons))
	}
	mx := buttons[2].X + buttons[2].W/2
	my := buttons[2].Y + buttons[2].H/2

	if !RecruitPanelInteractiveHit(mx, my, gs, rid) {
		t.Fatal("kart merkezi görünür Kışla paneli için hit-test vermedi")
	}
	if r.inGameHovering(mx, my) {
		t.Fatal("Kışla paneli kapalıyken görünmeyen kart cursor hover alanı oluşturdu")
	}

	r.showRecruitPanel = true
	if !r.inGameHovering(mx, my) {
		t.Fatal("Kışla paneli açıkken görünen kart cursor hover alanı oluşturmadı")
	}
}
