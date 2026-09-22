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

// IndependentWarSatisfactionPenalties, aynı state için bütün realm'lerin
// savaş yorgunluğu cezalarını tek dünya/ilişki taramasında hesaplar. Eski
// tekil helper davranışı korunur; toplu sonuç AI ve ekonomi hesaplarında aynı
// tur snapshot'ı olarak yeniden kullanılabilir.
func IndependentWarSatisfactionPenalties(gs *state.GameState) map[faction.FactionID]int {
	penalties := make(map[faction.FactionID]int)
	if gs == nil {
		return penalties
	}

	realmByFaction := make(map[faction.FactionID]faction.FactionID, len(gs.Factions))
	for fid := range gs.Factions {
		realm := RealmRoot(gs, fid)
		if realm == "" {
			realm = fid
		}
		realmByFaction[fid] = realm
	}
	resolveRealm := func(fid faction.FactionID) faction.FactionID {
		if realm, ok := realmByFaction[fid]; ok {
			return realm
		}
		realm := RealmRoot(gs, fid)
		if realm == "" {
			return fid
		}
		return realm
	}

	warOpponents := make(map[faction.FactionID]map[faction.FactionID]struct{})
	for _, relation := range gs.Relations {
		if relation == nil || relation.Stance != faction.StanceWar {
			continue
		}
		first := resolveRealm(relation.FactionA)
		second := resolveRealm(relation.FactionB)
		if first == "" || second == "" || first == second {
			continue
		}
		if warOpponents[first] == nil {
			warOpponents[first] = make(map[faction.FactionID]struct{})
		}
		if warOpponents[second] == nil {
			warOpponents[second] = make(map[faction.FactionID]struct{})
		}
		warOpponents[first][second] = struct{}{}
		warOpponents[second][first] = struct{}{}
	}

	landBorders := make(map[faction.FactionID]map[faction.FactionID]struct{})
	seaNeighbors := make(map[world.RegionID]map[faction.FactionID]struct{})
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID == "" {
			continue
		}
		owner := resolveRealm(faction.FactionID(region.OwnerID))
		if owner == "" {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil {
				continue
			}
			if neighbor.IsSea {
				if seaNeighbors[neighbor.ID] == nil {
					seaNeighbors[neighbor.ID] = make(map[faction.FactionID]struct{})
				}
				seaNeighbors[neighbor.ID][owner] = struct{}{}
				continue
			}
			if neighbor.OwnerID == "" {
				continue
			}
			other := resolveRealm(faction.FactionID(neighbor.OwnerID))
			if other == "" || other == owner {
				continue
			}
			if landBorders[owner] == nil {
				landBorders[owner] = make(map[faction.FactionID]struct{})
			}
			landBorders[owner][other] = struct{}{}
		}
	}

	seaBorders := make(map[faction.FactionID]map[faction.FactionID]struct{})
	for _, owners := range seaNeighbors {
		ownerIDs := make([]faction.FactionID, 0, len(owners))
		for owner := range owners {
			ownerIDs = append(ownerIDs, owner)
		}
		for i := 0; i < len(ownerIDs); i++ {
			for j := i + 1; j < len(ownerIDs); j++ {
				first, second := ownerIDs[i], ownerIDs[j]
				if seaBorders[first] == nil {
					seaBorders[first] = make(map[faction.FactionID]struct{})
				}
				if seaBorders[second] == nil {
					seaBorders[second] = make(map[faction.FactionID]struct{})
				}
				seaBorders[first][second] = struct{}{}
				seaBorders[second][first] = struct{}{}
			}
		}
	}

	for realm, opponents := range warOpponents {
		penalty := 0
		for opponent := range opponents {
			if _, ok := landBorders[realm][opponent]; ok {
				penalty += WarFatigueSatisfactionPenalty
			} else if _, ok := seaBorders[realm][opponent]; ok {
				penalty += warFatigueSeaBorderPenalty
			} else {
				penalty += warFatigueRemotePenalty
			}
		}
		penalties[realm] = penalty
	}
	for fid, realm := range realmByFaction {
		if _, ok := penalties[fid]; !ok {
			penalties[fid] = penalties[realm]
		}
	}
	return penalties
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
