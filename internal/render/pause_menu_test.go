package render

import (
	"testing"

	gameui "mapp-game-go/internal/ui"
)

func TestPauseMusicVolumeClickUsesArrowSide(t *testing.T) {
	button := gameui.NewButton(100, 200, 300, 50, "Müzik Seviyesi")

	if got := pauseMusicDeltaForClick(button, button.X+1, 10); got != -10 {
		t.Fatalf("sol ses oku azaltmalı: got=%d want=-10", got)
	}
	if got := pauseMusicDeltaForClick(button, button.X+button.W-1, 10); got != 10 {
		t.Fatalf("sağ ses oku artırmalı: got=%d want=10", got)
	}
}
