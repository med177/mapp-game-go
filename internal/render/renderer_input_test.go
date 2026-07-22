package render

import (
	"testing"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestSelectMapRegionDoesNotOpenRecruitPanel(t *testing.T) {
	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"bursa": {ID: "bursa", OwnerID: "player"},
			},
		},
		SelectedRegion:   "ankara",
		showRecruitPanel: true,
		recruitUnitID:    "militia",
		recruitQty:       3,
	}

	r.selectMapRegion("bursa")
	r.selectMapRegion("bursa")

	if r.showRecruitPanel {
		t.Fatal("bölge seçimi recruit panelini açmamalı")
	}
	if r.recruitUnitID != "" || r.recruitQty != 1 {
		t.Fatalf("bölge seçimi recruit seçimini temizlemeli: unit=%q qty=%d", r.recruitUnitID, r.recruitQty)
	}
}
