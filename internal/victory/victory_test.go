package victory

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestIsRegionControlledForVictoryAllowsDirectVassalWhenConfigured(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "ottoman",
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman"},
			"mamluk":  {ID: "mamluk", OverlordID: "ottoman"},
		},
		Victory: state.VictoryCondition{AllowVassalControl: true},
	}

	if !IsRegionControlledForVictory(gs, &world.Region{OwnerID: "ottoman"}) {
		t.Fatal("oyuncu sahibi olduğu bölgeyi kontrol ediyor sayılmalı")
	}
	if !IsRegionControlledForVictory(gs, &world.Region{OwnerID: "mamluk"}) {
		t.Fatal("doğrudan vassal bölgesi kontrol ediliyor sayılmalı")
	}
}

func TestIsRegionControlledForVictoryDoesNotAllowVassalByDefault(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "ottoman",
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman"},
			"mamluk":  {ID: "mamluk", OverlordID: "ottoman"},
		},
	}

	if IsRegionControlledForVictory(gs, &world.Region{OwnerID: "mamluk"}) {
		t.Fatal("vassal kontrolü açık değilken bölge oyuncununmuş sayılmamalı")
	}
}

func TestControlledRegionCountExcludesSeaAndTerrainAreas(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "ottoman",
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman"},
		},
		Victory: state.VictoryCondition{AllowVassalControl: true},
		Regions: map[world.RegionID]*world.Region{
			"land": {ID: "land", OwnerID: "ottoman"},
			"sea":  {ID: "sea", OwnerID: "ottoman", IsSea: true},
			"area": {ID: "area", OwnerID: "ottoman", IsTerrainArea: true},
		},
	}

	if got := ControlledRegionCount(gs); got != 1 {
		t.Fatalf("ControlledRegionCount() = %d, 1 bekleniyordu", got)
	}
}
