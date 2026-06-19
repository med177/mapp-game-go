package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestNavalShowsFriendlyDisembark(t *testing.T) {
	gs := &state.GameState{}
	fleet := &army.Army{
		ID:            "fleet",
		OwnerID:       "p1",
		RegionID:      "sea_1",
		IsNaval:       true,
		EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
	}

	if !navalShowsFriendlyDisembark(gs, fleet, &world.Region{ID: "land_a", OwnerID: "p1"}) {
		t.Fatal("kendi kara bölgesi için IN davranışı bekleniyordu")
	}
	if !navalShowsFriendlyDisembark(gs, fleet, &world.Region{ID: "land_b"}) {
		t.Fatal("boş kara bölgesi için IN davranışı bekleniyordu")
	}
	if navalShowsFriendlyDisembark(gs, fleet, &world.Region{ID: "land_e", OwnerID: "p2"}) {
		t.Fatal("düşman kara bölgesi için friendly IN davranışı olmamalı")
	}
	if navalShowsFriendlyDisembark(gs, fleet, &world.Region{ID: "sea_1", IsSea: true}) {
		t.Fatal("deniz bölgesi için friendly IN davranışı olmamalı")
	}
}
