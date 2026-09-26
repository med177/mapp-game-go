package ui

import "testing"

func TestLayerStackUsesTopmostVisibleLayer(t *testing.T) {
	stack := NewLayerStack(2)
	stack.AddRect("alt", Rect{X: 10, Y: 10, W: 100, H: 100})
	stack.AddRect("ust", Rect{X: 40, Y: 40, W: 100, H: 100})

	layer, ok := stack.TopAt(60, 60)
	if !ok || layer.ID != "ust" {
		t.Fatalf("üst katman = (%q, %v), want (%q, true)", layer.ID, ok, "ust")
	}
}

func TestLayerStackBlocksOnlyCoveredCoordinates(t *testing.T) {
	stack := NewLayerStack(1)
	stack.AddRect("panel", Rect{X: 10, Y: 10, W: 100, H: 100})

	if !stack.BlocksAt(50, 50) {
		t.Fatal("panel içindeki koordinat input'u tüketmedi")
	}
	if stack.BlocksAt(150, 50) {
		t.Fatal("panel dışındaki koordinat input'u tüketildi")
	}
}

func TestLayerStackResetPreservesUsableStack(t *testing.T) {
	stack := NewLayerStack(1)
	stack.AddRect("old", Rect{W: 10, H: 10})
	stack.Reset()
	if stack.Len() != 0 {
		t.Fatalf("reset sonrası katman sayısı = %d, want 0", stack.Len())
	}
	stack.AddRect("new", Rect{W: 10, H: 10})
	if layer, ok := stack.TopAt(1, 1); !ok || layer.ID != "new" {
		t.Fatalf("reset sonrası yeni katman = (%q, %v)", layer.ID, ok)
	}
}
