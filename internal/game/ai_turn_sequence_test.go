package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestOrderedAIFactionsUsesFactionOrder(t *testing.T) {
	g := &Game{
		gs: &state.GameState{
			PlayerFactionID: "player",
			FactionOrder:    []faction.FactionID{"venice", "player", "mamluk"},
			Factions: map[faction.FactionID]*faction.Faction{
				"player":  {ID: "player"},
				"venice":  {ID: "venice"},
				"mamluk":  {ID: "mamluk"},
				"england": {ID: "england"},
			},
		},
	}

	got := g.orderedAIFactions()
	want := []faction.FactionID{"venice", "mamluk", "england"}
	if len(got) != len(want) {
		t.Fatalf("beklenen %v, got=%v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("beklenen %v, got=%v", want, got)
		}
	}
}

func TestRegionNearPlayerUsesOwnedRegionsAndArmies(t *testing.T) {
	g := &Game{
		gs: &state.GameState{
			PlayerFactionID: "player",
			Regions: map[world.RegionID]*world.Region{
				"home":     {ID: "home", OwnerID: "player", Neighbors: []world.RegionID{"frontier"}},
				"frontier": {ID: "frontier", Neighbors: []world.RegionID{"home", "outer"}},
				"outer":    {ID: "outer", Neighbors: []world.RegionID{"frontier", "deep"}},
				"deep":     {ID: "deep", Neighbors: []world.RegionID{"outer"}},
				"fleet":    {ID: "fleet", IsSea: true, Neighbors: []world.RegionID{"deep"}},
			},
			Armies: map[army.ArmyID]*army.Army{
				"player_stack": {ID: "player_stack", OwnerID: "player", RegionID: "fleet"},
			},
		},
	}

	if !g.regionNearPlayer("outer", 2) {
		t.Fatal("outer bölgesi oyuncu toprağından iki sıçrama içinde görünür olmalı")
	}
	if !g.regionNearPlayer("deep", 1) {
		t.Fatal("deep bölgesi oyuncu ordusunun bulunduğu denizden bir sıçrama içinde görünür olmalı")
	}
	if g.regionNearPlayer("deep", 0) {
		t.Fatal("sıfır derinlikte yalnız başlangıç bölgeleri görünür olmalı")
	}
}
