package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIOrderingHelpersAreStableAndNilSafe(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"zeta":  {ID: "zeta"},
			"alpha": {ID: "alpha"},
		},
		Regions: map[world.RegionID]*world.Region{
			"r2":  {ID: "r2"},
			"r1":  {ID: "r1"},
			"nil": nil,
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_b": {ID: "army_b"},
			"army_a": {ID: "army_a"},
			"nil":    nil,
		},
	}

	factions := aiSortedFactionIDs(gs)
	if len(factions) != 2 || factions[0] != "alpha" || factions[1] != "zeta" {
		t.Fatalf("faction sıralaması deterministik değil: %v", factions)
	}
	regions := aiSortedRegions(gs)
	if len(regions) != 2 || regions[0].ID != "r1" || regions[1].ID != "r2" {
		t.Fatalf("region sıralaması deterministik değil: %v", regions)
	}
	armies := aiSortedArmies(gs)
	if len(armies) != 2 || armies[0].ID != "army_a" || armies[1].ID != "army_b" {
		t.Fatalf("army sıralaması deterministik değil: %v", armies)
	}
	if aiSortedFactionIDs(nil) != nil || aiSortedRegions(nil) != nil || aiSortedArmies(nil) != nil {
		t.Fatal("nil GameState için ordering helper'ları nil dönmeli")
	}
}

func TestAISortedRegionsUsesCanonicalRegionOrder(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1"},
			"r2": {ID: "r2"},
		},
		RegionOrder: []world.RegionID{"r2", "r1"},
	}
	regions := aiSortedRegions(gs)
	if len(regions) != 2 || regions[0].ID != "r2" || regions[1].ID != "r1" {
		t.Fatalf("canonical region order kullanılmadı: %v", regions)
	}
}

func TestAISortedFactionsUsesCanonicalFactionOrder(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"alpha": {ID: "alpha"},
			"zeta":  {ID: "zeta"},
		},
		FactionOrder: []faction.FactionID{"zeta", "alpha"},
	}
	ids := aiSortedFactionIDs(gs)
	if len(ids) != 2 || ids[0] != "zeta" || ids[1] != "alpha" {
		t.Fatalf("canonical faction order kullanılmadı: %v", ids)
	}
}

func TestAISortedArmiesCachesCanonicalOrder(t *testing.T) {
	gs := &state.GameState{
		Armies: map[army.ArmyID]*army.Army{
			"army_b": {ID: "army_b"},
			"army_a": {ID: "army_a"},
		},
	}
	armies := aiSortedArmies(gs)
	if len(armies) != 2 || armies[0].ID != "army_a" || armies[1].ID != "army_b" {
		t.Fatalf("ordu sıralaması deterministik değil: %v", armies)
	}
	if len(gs.ArmyOrder) != 2 || gs.ArmyOrder[0] != "army_a" || gs.ArmyOrder[1] != "army_b" {
		t.Fatalf("ordu sırası state'e cache'lenmedi: %v", gs.ArmyOrder)
	}
	gs.ArmyOrder = []army.ArmyID{"army_b", "army_a"}
	armies = aiSortedArmies(gs)
	if armies[0].ID != "army_b" || armies[1].ID != "army_a" {
		t.Fatalf("geçerli cached ordu sırası kullanılmadı: %v", armies)
	}
}
