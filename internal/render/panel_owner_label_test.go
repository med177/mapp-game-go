package render

import (
	"image/color"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestOwnerLabelOutlineColorUsesOppositeContrast(t *testing.T) {
	darkOutline := ownerLabelOutlineColor(color.RGBA{210, 210, 210, 255})
	if darkOutline != (color.RGBA{18, 16, 12, 220}) {
		t.Fatalf("acik renk icin koyu outline bekleniyordu, alinan=%v", darkOutline)
	}

	lightOutline := ownerLabelOutlineColor(color.RGBA{30, 70, 145, 255})
	if lightOutline != (color.RGBA{245, 240, 230, 210}) {
		t.Fatalf("koyu renk icin acik outline bekleniyordu, alinan=%v", lightOutline)
	}
}

func TestOwnerDisplayUsesSingleColor(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
			"ven": {ID: "ven", NameTR: "Venedik"},
		},
	}

	if _, col := ownerDisplay(gs, "osm"); col != ColorWhite {
		t.Fatalf("oyuncu devleti tek renk cikmali, alinan=%v", col)
	}
	if _, col := ownerDisplay(gs, "ven"); col != ColorWhite {
		t.Fatalf("ai devleti tek renk cikmali, alinan=%v", col)
	}
}
