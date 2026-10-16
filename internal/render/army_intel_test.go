package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

func TestEnemyArmyIntelRequiresSameOrNeighboringPlayerRegion(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			"player-region": {ID: "player-region", Neighbors: []world.RegionID{"neighbor"}},
			"neighbor":      {ID: "neighbor", Neighbors: []world.RegionID{"player-region"}},
			"far":           {ID: "far"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player-army": {ID: "player-army", OwnerID: "player", RegionID: "player-region"},
			"enemy-army":  {ID: "enemy-army", OwnerID: "enemy", RegionID: "far"},
		},
	}
	enemy := gs.Armies["enemy-army"]

	if got := playerArmyIntelLevel(gs, enemy); got != armyIntelNone {
		t.Fatalf("uzak düşman ordusu istihbaratı = %v, want none", got)
	}
	if rows := armyMarkerSummaryRows(gs, enemy); rows != nil {
		t.Fatalf("uzak düşman ordusunun hover özeti gizlenmedi: %v", rows)
	}

	enemy.RegionID = "neighbor"
	if got := playerArmyIntelLevel(gs, enemy); got != armyIntelScouted {
		t.Fatalf("komşu bölgedeki düşman ordusu istihbaratı = %v, want scouted", got)
	}
	enemy.RegionID = "player-region"
	if got := playerArmyIntelLevel(gs, enemy); got != armyIntelScouted {
		t.Fatalf("aynı bölgedeki düşman ordusu istihbaratı = %v, want scouted", got)
	}
}

func TestScoutedArmyCombatViewHidesCommanderAndUnseenUnits(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			"region": {ID: "region"},
			"near":   {ID: "near", Neighbors: []world.RegionID{"region"}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player-army": {ID: "player-army", OwnerID: "player", RegionID: "near"},
		},
	}

	commander := &army.Commander{Name: "Gizli Komutan"}
	enemy := &army.Army{
		ID:        "enemy-army",
		OwnerID:   "enemy",
		RegionID:  "region",
		Commander: commander,
		Units: []army.Unit{
			{TypeID: "infantry", CurrentHP: army.MaxUnitHP},
			{TypeID: "infantry", CurrentHP: army.MaxUnitHP},
		},
	}

	view, _, intelLimited, ok := armyPanelCombatView(gs, enemy, false)
	if !ok || !intelLimited {
		t.Fatalf("komşu ordusu combat görünümü = (ok %v, sınırlı %v), want görünür ve sınırlı", ok, intelLimited)
	}
	if len(view.Units) != 1 {
		t.Fatalf("görünen birim sayısı = %d, want 1", len(view.Units))
	}
	if view.Commander != nil || view.EmbarkedCommander != nil {
		t.Fatal("kısmi istihbarat görünümünde komutan bilgisi açığa çıktı")
	}
}

func TestNearbyArmyIntelTechnologyDoesNotRevealDistantArmies(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Research: faction.ResearchState{Completed: map[string]bool{"scouting": true}}},
			"enemy":  {ID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", Neighbors: []world.RegionID{"near"}},
			"near": {ID: "near", Neighbors: []world.RegionID{"home"}},
			"far":  {ID: "far"},
		},
		TechTypes: map[string]*tech.Technology{
			"scouting": {ID: "scouting", Effects: tech.Effects{RevealNearbyArmyIntel: true}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player-army": {ID: "player-army", OwnerID: "player", RegionID: "home"},
			"near-army":   {ID: "near-army", OwnerID: "enemy", RegionID: "near"},
			"far-army":    {ID: "far-army", OwnerID: "enemy", RegionID: "far"},
		},
	}

	if got := playerArmyIntelLevel(gs, gs.Armies["near-army"]); got != armyIntelFull {
		t.Fatalf("komşu ordusu istihbarat seviyesi = %v, want full", got)
	}
	if got := playerArmyIntelLevel(gs, gs.Armies["far-army"]); got != armyIntelNone {
		t.Fatalf("uzak ordu istihbarat seviyesi = %v, want none", got)
	}
}
