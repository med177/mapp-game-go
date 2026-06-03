package ui

import "testing"

func TestNewModalDoesNotAllocateChildrenSlice(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		_ = NewModal(1280, 720, NewPanel(100, 100, 200, 120))
	})
	if allocs != 0 {
		t.Fatalf("expected NewModal to allocate 0 times, got %.2f", allocs)
	}
}
