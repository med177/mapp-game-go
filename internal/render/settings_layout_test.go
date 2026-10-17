package render

import "testing"

func TestSettingsVolumeButtonsAreCenteredInSelectedRow(t *testing.T) {
	highlight := settingsRowHighlightRect(settingsRowRect(4))
	minus, plus := settingsVolumeButtons(4, 50)
	wantCenterY := highlight.Y + highlight.H/2

	for name, button := range map[string]struct {
		Y float64
		H float64
	}{
		"minus": {Y: minus.Y, H: minus.H},
		"plus":  {Y: plus.Y, H: plus.H},
	} {
		if got := button.Y + button.H/2; got != wantCenterY {
			t.Errorf("%s button center y = %.1f, selected row center y = %.1f", name, got, wantCenterY)
		}
	}
}
