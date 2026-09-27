package render

import (
	"testing"

	"mapp-game-go/internal/state"

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
