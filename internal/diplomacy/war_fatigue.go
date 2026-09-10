package diplomacy

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// WarFatigueSatisfactionPenalty kara sınırı olan bağımsız düşman realm başına
// ekonomi turunda uygulanan memnuniyet cezasıdır.
const WarFatigueSatisfactionPenalty = 3

const (
	warFatigueSeaBorderPenalty = 2
	warFatigueRemotePenalty    = 1
)

// IndependentWarCount, fid'nin savaş halinde olduğu bağımsız realm sayısını
// döner. Overlord ve vassalları aynı realm içinde tek devlet sayılır.
func IndependentWarCount(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" {
		return 0
	}
	realm := RealmRoot(gs, fid)
	if realm == "" {
		realm = fid
	}
	return len(independentWarRealms(gs, realm))
}

// IndependentWarSatisfactionPenalty, AI ve ekonomi kararlarının aynı savaş
// yorgunluğu projeksiyonunu kullanması için ortak yardımcıdır.
func IndependentWarSatisfactionPenalty(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" {
		return 0
	}
	realm := RealmRoot(gs, fid)
	if realm == "" {
		realm = fid
	}

	opponents := independentWarRealms(gs, realm)
	penalty := 0
	for opponent := range opponents {
		switch {
		case realmsShareLandBorder(gs, realm, opponent):
			penalty += WarFatigueSatisfactionPenalty
		case realmsShareSeaBorder(gs, realm, opponent):
			penalty += warFatigueSeaBorderPenalty
		default:
			penalty += warFatigueRemotePenalty
		}
	}
	return penalty
}

func independentWarRealms(gs *state.GameState, realm faction.FactionID) map[faction.FactionID]struct{} {
	opponents := make(map[faction.FactionID]struct{})
	if gs == nil || realm == "" {
		return opponents
	}
	for _, relation := range gs.Relations {
		if relation == nil || relation.Stance != faction.StanceWar {
			continue
		}
		realmA := RealmRoot(gs, relation.FactionA)
		if realmA == "" {
			realmA = relation.FactionA
		}
		realmB := RealmRoot(gs, relation.FactionB)
		if realmB == "" {
			realmB = relation.FactionB
		}
		if realmA == realmB {
			continue
		}
		if realmA == realm {
			opponents[realmB] = struct{}{}
		}
		if realmB == realm {
			opponents[realmA] = struct{}{}
		}
	}
	return opponents
}

func realmsShareLandBorder(gs *state.GameState, first, second faction.FactionID) bool {
	if gs == nil || first == "" || second == "" || first == second {
		return false
	}
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || RealmRoot(gs, faction.FactionID(region.OwnerID)) != first {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor != nil && !neighbor.IsSea && RealmRoot(gs, faction.FactionID(neighbor.OwnerID)) == second {
				return true
			}
		}
	}
	return false
}

func realmsShareSeaBorder(gs *state.GameState, first, second faction.FactionID) bool {
	if gs == nil || first == "" || second == "" || first == second {
		return false
	}
	firstSeas := make(map[world.RegionID]struct{})
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || RealmRoot(gs, faction.FactionID(region.OwnerID)) != first {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor != nil && neighbor.IsSea {
				firstSeas[neighborID] = struct{}{}
			}
		}
	}
	if len(firstSeas) == 0 {
		return false
	}
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || RealmRoot(gs, faction.FactionID(region.OwnerID)) != second {
			continue
		}
		for _, neighborID := range region.Neighbors {
			if _, shared := firstSeas[neighborID]; shared {
				return true
			}
		}
	}
	return false
}
