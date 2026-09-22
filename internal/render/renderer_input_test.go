package render

import (
	"testing"

	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"
)

func TestInfoPopupClickDismissesOnlyInsidePopup(t *testing.T) {
	popup := gameui.Rect{X: 100, Y: 100, W: 200, H: 80}

	tests := []struct {
		name  string
		timer int
		mx    float64
		my    float64
		want  bool
	}{
		{name: "popup ici", timer: 10, mx: 150, my: 120, want: true},
		{name: "harita disi", timer: 10, mx: 50, my: 120, want: false},
		{name: "sure bitmis", timer: 0, mx: 150, my: 120, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := infoPopupClickDismisses(tt.timer, popup, tt.mx, tt.my, true, false)
			if got != tt.want {
				t.Fatalf("popup kapatma = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfoPopupClickDismissesRequiresNewClick(t *testing.T) {
	popup := gameui.Rect{X: 100, Y: 100, W: 200, H: 80}
	if infoPopupClickDismisses(10, popup, 150, 120, true, true) {
		t.Fatal("basili tutulmus eski tiklama popup'i kapatti")
	}
}

func TestMapRegionDoubleClickUsesTerrainParentIdentity(t *testing.T) {
	parentID := world.RegionID("parent")
	fragmentID := world.RegionID("area::forest::parent")
	r := &Renderer{gs: &state.GameState{Regions: map[world.RegionID]*world.Region{
		parentID:   {ID: parentID, OwnerID: "target"},
		fragmentID: {ID: fragmentID, IsTerrainArea: true, ParentRegionID: parentID},
	}}}

	if r.mapRegionDoubleClicked(fragmentID) {
		t.Fatal("first terrain click unexpectedly counted as double click")
	}
	if !r.mapRegionDoubleClicked(parentID) {
		t.Fatal("terrain fragment and its parent were not treated as the same map click target")
	}
}
