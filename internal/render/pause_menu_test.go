package render

import (
	"testing"

	gameui "mapp-game-go/internal/ui"
)

func TestPauseMenuLabelsAreVerticallyCenteredInRows(t *testing.T) {
	row := gameui.NewButton(20, 30, 300, 54, "")
	label := pauseMenuLabelRect(row, row.X, row.W)
	if label.Y+label.H/2 != row.Y+row.H/2 {
		t.Fatalf("label center y = %.1f, row center y = %.1f", label.Y+label.H/2, row.Y+row.H/2)
	}
}
