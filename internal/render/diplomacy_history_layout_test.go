package render

import (
	"testing"

	gameui "mapp-game-go/internal/ui"
)

func TestDiplomacyHistoryActionFiltersUseSpaciousCenteredRows(t *testing.T) {
	panel := gameui.Rect{X: 100, Y: 200, W: diplomHistoryPanelW, H: diplomHistoryPanelH}
	buttons := buildDiplomacyHistoryFilterButtons(panel, diplomacyHistoryDirectionAll, ActionNone)

	for i := 3; i < len(buttons); i++ {
		button := buttons[i].Button
		if button.W < 75 {
			t.Errorf("action filter %q width = %.1f, want at least 75", button.Label, button.W)
		}
		if button.X < panel.X || button.X+button.W > panel.X+panel.W {
			t.Errorf("action filter %q extends beyond panel: %+v", button.Label, button)
		}
	}

	if got, want := buttons[7].Button.X, buttons[6].Button.X+buttons[6].Button.W+6; got != want {
		t.Errorf("second action row is not centered/contiguous: last button starts at %.1f, want %.1f", got, want)
	}
	if buttons[7].Button.Y != buttons[6].Button.Y {
		t.Errorf("second action row buttons have different Y positions: %.1f and %.1f", buttons[6].Button.Y, buttons[7].Button.Y)
	}
}

func TestDiplomacyHistoryVisibleEntriesStayInsidePanel(t *testing.T) {
	for _, tc := range []struct {
		name       string
		panelH     float64
		maxEntries int
		want       int
	}{
		{name: "legacy height clips fourth card", panelH: 324, maxEntries: 4, want: 3},
		{name: "expanded panel fits four cards", panelH: 400, maxEntries: 4, want: 4},
		{name: "respects caller limit", panelH: 400, maxEntries: 3, want: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			panel := gameui.Rect{X: 40, Y: 60, W: diplomHistoryPanelW, H: tc.panelH}
			if got := diplomacyHistoryVisibleEntries(panel, tc.maxEntries); got != tc.want {
				t.Fatalf("visible entries = %d, want %d", got, tc.want)
			}
			count := diplomacyHistoryVisibleEntries(panel, tc.maxEntries)
			if count == 0 {
				return
			}
			last := diplomacyOfferHistoryCardRect(panel, count-1)
			if bottom, limit := last.Y+last.H, panel.Y+panel.H-8; bottom > limit {
				t.Errorf("last history card ends at %.1f, beyond inset panel bottom %.1f", bottom, limit)
			}
		})
	}
}
