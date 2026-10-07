package events

import (
	"testing"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestVictoryConditionsRequireKeyRegionsAndMinimumOwnership(t *testing.T) {
	gs := &state.GameState{
		FiredEventIDs: map[string]bool{
			"flag:war_of_five_kings_active": true,
		},
		Regions: map[world.RegionID]*world.Region{
			"kings_landing": {OwnerID: "lannister"},
			"casterly_rock": {OwnerID: "lannister"},
			"riverlands":    {OwnerID: "lannister"},
		},
	}
	e := &Event{
		Target:        "all_factions",
		RequiresFlags: []string{"war_of_five_kings_active"},
		VictoryConditions: []FactionVictoryCondition{{
			FactionID:            "lannister",
			RequiredOwnedRegions: []world.RegionID{"kings_landing", "casterly_rock"},
			MinimumOwnedRegions:  4,
		}},
	}

	if eventConditionsSatisfied(gs, e) {
		t.Fatal("minimum sahiplik koşulu sağlanmadan zafer koşulu başarılı oldu")
	}

	gs.Regions["lannisport"] = &world.Region{OwnerID: "lannister"}
	if !eventConditionsSatisfied(gs, e) {
		t.Fatal("kilit bölgeler ve minimum sahiplik sağlandığında zafer koşulu başarısız oldu")
	}
}
