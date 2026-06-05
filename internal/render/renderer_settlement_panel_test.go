package render

import (
	"testing"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestSettlementPanelHitRequiresVisiblePanel(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()

	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"morea": {
					ID: "morea",
					Settlements: []world.Settlement{
						{Name: "Mora"},
					},
				},
			},
		},
		SelectedRegion: "morea",
	}

	mx := float64(settlementPanelX() + 20)
	my := float64(settlementPanelY() + 20)

	if r.settlementPanelHit(mx, my) {
		t.Fatal("yerlesim paneli kapaliyken hit-test map alani tuketmemeli")
	}
	if r.settlementPanelCloseHit(mx, my) {
		t.Fatal("yerlesim paneli kapaliyken kapatma butonu aktif olmamali")
	}

	r.selectSettlement("morea", 0)

	if !r.settlementPanelHit(mx, my) {
		t.Fatal("yerlesim paneli acikken hit-test aktif olmali")
	}
}
