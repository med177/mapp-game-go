package save

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestLoadScenarioBaseStateLoadsTerrainAreas(t *testing.T) {
	gs, err := loadScenarioBaseState("1300_ottoman_rise", filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("loadScenarioBaseState() error = %v", err)
	}
	if len(gs.TerrainAreas) == 0 {
		t.Fatal("scenario terrain areas were not loaded")
	}
	for _, area := range gs.TerrainAreas {
		if gs.Regions[world.TerrainAreaRegionID(area.ID)] == nil {
			t.Fatalf("runtime terrain region for %q was not created", area.ID)
		}
	}
}

func TestCampaignSaveStateRestoresTerrainAreasAndRuntimeRegions(t *testing.T) {
	areas := []world.TerrainArea{{
		ID:       "saved_area",
		Terrain:  world.TerrainDesert,
		MoveCost: -1,
		Polygons: [][][2]int{{{1, 1}, {4, 1}, {4, 4}, {1, 4}}},
	}}
	saved := campaignSaveState{ScenarioID: "1300_ottoman_rise", TerrainAreas: areas}
	payload, err := json.Marshal(saved)
	if err != nil {
		t.Fatalf("marshal campaign save state: %v", err)
	}
	decoded, err := decodeCampaignSaveState(payload)
	if err != nil {
		t.Fatalf("decode campaign save state: %v", err)
	}
	if len(decoded.TerrainAreas) != 1 || decoded.TerrainAreas[0].ID != "saved_area" {
		t.Fatalf("decoded terrain areas = %#v", decoded.TerrainAreas)
	}

	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"base": {ID: "base", Terrain: world.TerrainPlain},
		},
		TerrainAreas: []world.TerrainArea{{ID: "old_area", MoveCost: 0}},
	}
	applyCampaignSaveState(gs, decoded)
	if len(gs.TerrainAreas) != 1 || gs.TerrainAreas[0].ID != "saved_area" {
		t.Fatalf("restored terrain areas = %#v", gs.TerrainAreas)
	}
	if gs.Regions[world.TerrainAreaRegionID("saved_area")] == nil {
		t.Fatal("restored terrain runtime region was not recreated")
	}
}

func TestCampaignSaveStateIgnoresStaleTerrainRegionLock(t *testing.T) {
	areas := []world.TerrainArea{{
		ID:       "blocked_area",
		MoveCost: 0,
		Polygons: [][][2]int{{{1, 1}, {4, 1}, {4, 4}, {1, 4}}},
	}}
	wasUnlocked := false
	saved := campaignSaveState{
		ScenarioID:   "1300_ottoman_rise",
		TerrainAreas: areas,
		Regions: map[world.RegionID]regionSaveState{
			world.TerrainAreaRegionID("blocked_area"): {IsLocked: &wasUnlocked},
		},
	}

	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"base": {ID: "base", Terrain: world.TerrainPlain},
		},
	}
	applyCampaignSaveState(gs, saved)

	terrain := gs.Regions[world.TerrainAreaRegionID("blocked_area")]
	if terrain == nil {
		t.Fatal("terrain runtime region was not recreated")
	}
	if !terrain.IsLocked {
		t.Fatal("stale saved unlock state made a move_cost=0 terrain area passable")
	}
}
