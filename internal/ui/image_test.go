package ui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewLinearImageDrawOptionsUsesLinearFilter(t *testing.T) {
	if options := NewLinearImageDrawOptions(); options.Filter != ebiten.FilterLinear {
		t.Fatalf("image filter = %v, want FilterLinear", options.Filter)
	}
}
