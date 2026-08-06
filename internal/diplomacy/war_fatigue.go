package diplomacy

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

// WarFatigueSatisfactionPenalty bağımsız düşman realm başına ekonomi turunda
// uygulanan memnuniyet cezasıdır.
const WarFatigueSatisfactionPenalty = 2

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
	opponents := make(map[faction.FactionID]struct{})
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
	return len(opponents)
}

// IndependentWarSatisfactionPenalty, AI ve ekonomi kararlarının aynı savaş
// yorgunluğu projeksiyonunu kullanması için ortak yardımcıdır.
func IndependentWarSatisfactionPenalty(gs *state.GameState, fid faction.FactionID) int {
	return IndependentWarCount(gs, fid) * WarFatigueSatisfactionPenalty
}
