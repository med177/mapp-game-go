package save

import (
	"encoding/json"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestTemporaryMilitaryIntelPersistsInCompactAndLegacySaves(t *testing.T) {
	intel := map[faction.FactionID]map[faction.FactionID]int{
		"player": {"target": 2},
	}
	regionIntel := map[faction.FactionID]map[world.RegionID]int{
		"player": {"scouted": 2},
	}
	compact := campaignSaveState{
		TemporaryMilitaryIntel: cloneFactionAttackTargets(intel),
		RegionIntel:            cloneFactionRegionTurns(regionIntel),
	}
	encoded, err := json.Marshal(compact)
	if err != nil {
		t.Fatalf("compact save serialize edilemedi: %v", err)
	}
	var decoded campaignSaveState
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("compact save deserialize edilemedi: %v", err)
	}
	loaded := &state.GameState{}
	applyCampaignSaveState(loaded, decoded)
	if got := loaded.TemporaryMilitaryIntel["player"]["target"]; got != 2 {
		t.Fatalf("compact save/load temporary intel = %d turns, want 2", got)
	}
	if got := loaded.TemporaryRegionMilitaryIntel["player"]["scouted"]; got != 2 {
		t.Fatalf("compact save/load region intel = %d turns, want 2", got)
	}

	debug := makeDebugCampaignSaveState(&state.GameState{
		TemporaryMilitaryIntel: intel, TemporaryRegionMilitaryIntel: regionIntel,
	})
	legacy := convertLegacyCampaignSaveState(debug)
	if got := legacy.TemporaryMilitaryIntel["player"]["target"]; got != 2 {
		t.Fatalf("legacy save conversion temporary intel = %d turns, want 2", got)
	}
	if got := legacy.RegionIntel["player"]["scouted"]; got != 2 {
		t.Fatalf("legacy save conversion region intel = %d turns, want 2", got)
	}
}
