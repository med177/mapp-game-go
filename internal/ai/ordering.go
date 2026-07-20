package ai

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// aiSortedFactionIDs returns faction IDs in stable order so map iteration cannot
// alter diplomacy or turn replay decisions.
func aiSortedFactionIDs(gs *state.GameState) []faction.FactionID {
	if gs == nil {
		return nil
	}
	if len(gs.FactionOrder) == len(gs.Factions) && len(gs.FactionOrder) > 0 {
		ids := make([]faction.FactionID, 0, len(gs.Factions))
		valid := true
		for _, fid := range gs.FactionOrder {
			if gs.Factions[fid] == nil {
				valid = false
				break
			}
			ids = append(ids, fid)
		}
		if valid {
			return ids
		}
	}
	ids := make([]faction.FactionID, 0, len(gs.Factions))
	for id := range gs.Factions {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// aiSortedRegions returns non-nil regions in stable ID order.
func aiSortedRegions(gs *state.GameState) []*world.Region {
	if gs == nil {
		return nil
	}
	// Scenario/save loading already establishes a canonical RegionOrder. Reuse
	// it on the hot path instead of sorting the same map for every AI decision.
	if len(gs.RegionOrder) == len(gs.Regions) && len(gs.RegionOrder) > 0 {
		regions := make([]*world.Region, 0, len(gs.Regions))
		valid := true
		for _, rid := range gs.RegionOrder {
			region := gs.Regions[rid]
			if region == nil {
				valid = false
				break
			}
			regions = append(regions, region)
		}
		if valid {
			return regions
		}
	}
	regions := make([]*world.Region, 0, len(gs.Regions))
	for _, region := range gs.Regions {
		if region != nil {
			regions = append(regions, region)
		}
	}
	sort.Slice(regions, func(i, j int) bool { return regions[i].ID < regions[j].ID })
	return regions
}

// aiSortedArmies returns non-nil armies in stable ID order.
func aiSortedArmies(gs *state.GameState) []*army.Army {
	if gs == nil {
		return nil
	}
	if len(gs.ArmyOrder) == len(gs.Armies) && len(gs.ArmyOrder) > 0 {
		armies := make([]*army.Army, 0, len(gs.Armies))
		valid := true
		for _, aid := range gs.ArmyOrder {
			candidate := gs.Armies[aid]
			if candidate == nil {
				valid = false
				break
			}
			armies = append(armies, candidate)
		}
		if valid {
			return armies
		}
	}
	armies := make([]*army.Army, 0, len(gs.Armies))
	for _, candidate := range gs.Armies {
		if candidate != nil {
			armies = append(armies, candidate)
		}
	}
	sort.Slice(armies, func(i, j int) bool { return armies[i].ID < armies[j].ID })
	gs.ArmyOrder = gs.ArmyOrder[:0]
	for _, candidate := range armies {
		gs.ArmyOrder = append(gs.ArmyOrder, candidate.ID)
	}
	return armies
}
