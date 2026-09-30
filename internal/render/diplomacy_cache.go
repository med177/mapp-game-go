package render

import (
	"sort"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/victory"
)

// diplomacyFactionMetrics, diplomasi listesindeki satırların her frame'de
// tekrar hesaplanmasını önleyen state-bağımlı özet değerleridir.
type diplomacyFactionMetrics struct {
	RegionCount  int
	LandPower    int
	NavalPower   int
	PowerRank    int
	FactionCount int
	PowerExact   bool
	TreasuryText string
}

type diplomacyActionPresentation struct {
	Action         ActionKind
	Chance         int
	Status         string
	DisabledReason string
}

type diplomacyRenderCache struct {
	gs       *state.GameState
	turn     int
	year     int
	month    int
	playerID faction.FactionID

	sorted      [4][]faction.FactionID
	sortedValid [4]bool

	powerReady  bool
	powers      map[faction.FactionID]int
	powerLand   map[faction.FactionID]int
	powerNaval  map[faction.FactionID]int
	powerExact  map[faction.FactionID]bool
	powerRanks  map[faction.FactionID]int
	powerCount  int
	metrics     map[faction.FactionID]diplomacyFactionMetrics
	actionCache map[faction.FactionID][]diplomacyActionPresentation
	actionValid map[faction.FactionID]bool
}

func (r *Renderer) invalidateDiplomacyCache() {
	if r == nil {
		return
	}
	r.diplomacyCache = diplomacyRenderCache{}
}

func (r *Renderer) ensureDiplomacyCache(gs *state.GameState) *diplomacyRenderCache {
	if r == nil || gs == nil {
		return nil
	}
	cache := &r.diplomacyCache
	if cache.gs != gs || cache.turn != gs.Turn || cache.year != gs.Year || cache.month != gs.Month || cache.playerID != gs.PlayerFactionID {
		*cache = diplomacyRenderCache{
			gs:          gs,
			turn:        gs.Turn,
			year:        gs.Year,
			month:       gs.Month,
			playerID:    gs.PlayerFactionID,
			metrics:     make(map[faction.FactionID]diplomacyFactionMetrics),
			actionCache: make(map[faction.FactionID][]diplomacyActionPresentation),
			actionValid: make(map[faction.FactionID]bool),
		}
	}
	return cache
}

func (r *Renderer) cachedDiplomacyFactions(sortMode diplomacyListSort) []faction.FactionID {
	cache := r.ensureDiplomacyCache(r.gs)
	if cache == nil {
		return nil
	}
	idx := int(sortMode)
	if idx < 0 || idx >= len(cache.sorted) {
		idx = int(diplomacyListSortAlphabetical)
	}
	if !cache.sortedValid[idx] {
		cache.sorted[idx] = buildCachedDiplomacyFactionOrder(r.gs, sortMode, cache)
		cache.sortedValid[idx] = true
	}
	return cache.sorted[idx]
}

func buildCachedDiplomacyFactionOrder(gs *state.GameState, sortMode diplomacyListSort, cache *diplomacyRenderCache) []faction.FactionID {
	if gs == nil {
		return nil
	}
	fids := make([]faction.FactionID, 0, len(gs.Factions))
	for fid, f := range gs.Factions {
		if f == nil || f.IsEliminated || f.IsVirtual {
			continue
		}
		fids = append(fids, fid)
	}

	relationScores := make(map[faction.FactionID]int)
	adjacentToPlayer := make(map[faction.FactionID]bool)
	economicIncome := make(map[faction.FactionID]int)
	economicGold := make(map[faction.FactionID]int)
	switch sortMode {
	case diplomacyListSortRelation:
		for _, fid := range fids {
			if rel := diplomacy.Relation(gs, gs.PlayerFactionID, fid); rel != nil {
				relationScores[fid] = diplomacy.RelationScore(gs, gs.PlayerFactionID, fid)
			}
			adjacentToPlayer[fid] = factionsShareLandBorder(gs, gs.PlayerFactionID, fid)
		}
	case diplomacyListSortPowerRanking:
		ensureDiplomacyPowerCache(gs, cache)
	case diplomacyListSortEconomicRanking:
		for _, fid := range fids {
			economicIncome[fid] = victoryGoldIncomeForDiplomacy(gs, fid)
			if f := gs.Factions[fid]; f != nil {
				economicGold[fid] = f.Gold
			}
		}
	}

	sort.Slice(fids, func(i, j int) bool {
		leftID, rightID := fids[i], fids[j]
		switch sortMode {
		case diplomacyListSortRelation:
			if relationScores[leftID] != relationScores[rightID] {
				return relationScores[leftID] > relationScores[rightID]
			}
			if adjacentToPlayer[leftID] != adjacentToPlayer[rightID] {
				return adjacentToPlayer[leftID]
			}
		case diplomacyListSortPowerRanking:
			if cache.powers[leftID] != cache.powers[rightID] {
				return cache.powers[leftID] > cache.powers[rightID]
			}
		case diplomacyListSortEconomicRanking:
			if economicIncome[leftID] != economicIncome[rightID] {
				return economicIncome[leftID] > economicIncome[rightID]
			}
			if economicGold[leftID] != economicGold[rightID] {
				return economicGold[leftID] > economicGold[rightID]
			}
		}
		return leftID < rightID
	})
	return fids
}

func ensureDiplomacyPowerCache(gs *state.GameState, cache *diplomacyRenderCache) {
	if cache == nil || cache.powerReady || gs == nil {
		return
	}
	cache.powers = make(map[faction.FactionID]int, len(gs.Factions))
	cache.powerLand = make(map[faction.FactionID]int, len(gs.Factions))
	cache.powerNaval = make(map[faction.FactionID]int, len(gs.Factions))
	cache.powerExact = make(map[faction.FactionID]bool, len(gs.Factions))
	cache.powerRanks = make(map[faction.FactionID]int, len(gs.Factions))
	for fid, f := range gs.Factions {
		if f == nil || f.IsEliminated {
			continue
		}
		land, naval, exact := displayedFactionPowerBreakdown(gs, fid)
		cache.powerLand[fid] = land
		cache.powerNaval[fid] = naval
		cache.powerExact[fid] = exact
		cache.powers[fid] = land + naval
		cache.powerCount++
	}
	for fid, power := range cache.powers {
		rank := 1
		for candidateID, candidatePower := range cache.powers {
			if candidateID == fid {
				continue
			}
			if candidatePower > power || (candidatePower == power && candidateID < fid) {
				rank++
			}
		}
		cache.powerRanks[fid] = rank
	}
	cache.powerReady = true
}

func victoryGoldIncomeForDiplomacy(gs *state.GameState, fid faction.FactionID) int {
	return victory.GoldIncomeForFaction(gs, fid)
}

func (r *Renderer) diplomacyMetrics(fid faction.FactionID) diplomacyFactionMetrics {
	cache := r.ensureDiplomacyCache(r.gs)
	return diplomacyFactionMetricsFor(r.gs, fid, cache)
}

func diplomacyFactionMetricsFor(gs *state.GameState, fid faction.FactionID, cache *diplomacyRenderCache) diplomacyFactionMetrics {
	if gs == nil || fid == "" {
		return diplomacyFactionMetrics{}
	}
	if cache == nil || cache.gs != gs {
		land, naval, exact := displayedFactionPowerBreakdown(gs, fid)
		return diplomacyFactionMetrics{
			RegionCount:  len(gs.LandRegionsOwnedBy(fid)),
			LandPower:    land,
			NavalPower:   naval,
			PowerRank:    0,
			FactionCount: 0,
			PowerExact:   exact,
			TreasuryText: factionTreasuryLabel(gs, fid),
		}
	}
	if metrics, ok := cache.metrics[fid]; ok {
		return metrics
	}
	ensureDiplomacyPowerCache(gs, cache)
	metrics := diplomacyFactionMetrics{
		RegionCount:  len(gs.LandRegionsOwnedBy(fid)),
		LandPower:    cache.powerLand[fid],
		NavalPower:   cache.powerNaval[fid],
		PowerRank:    cache.powerRanks[fid],
		FactionCount: cache.powerCount,
		PowerExact:   cache.powerExact[fid],
		TreasuryText: factionTreasuryLabel(gs, fid),
	}
	cache.metrics[fid] = metrics
	return metrics
}

func (r *Renderer) diplomacyActionPresentation(target faction.FactionID, index int) diplomacyActionPresentation {
	cache := r.ensureDiplomacyCache(r.gs)
	return diplomacyActionPresentationFor(r.gs, target, index, cache)
}

func diplomacyActionPresentationFor(gs *state.GameState, target faction.FactionID, index int, cache *diplomacyRenderCache) diplomacyActionPresentation {
	if gs == nil || target == "" || index < 0 || index >= len(diplomActions) {
		return diplomacyActionPresentation{}
	}
	if cache == nil || cache.gs != gs {
		action := diplomacyActionForTarget(gs, target, index)
		chance, status := estimateDiplomacyChance(gs, target, action)
		return diplomacyActionPresentation{Action: action, Chance: chance, Status: status, DisabledReason: diplomacyActionDisabledReason(gs, target, action)}
	}
	if !cache.actionValid[target] {
		presentations := make([]diplomacyActionPresentation, len(diplomActions))
		for i := range diplomActions {
			action := diplomacyActionForTarget(gs, target, i)
			chance, status := estimateDiplomacyChance(gs, target, action)
			presentations[i] = diplomacyActionPresentation{
				Action:         action,
				Chance:         chance,
				Status:         status,
				DisabledReason: diplomacyActionDisabledReason(gs, target, action),
			}
		}
		cache.actionCache[target] = presentations
		cache.actionValid[target] = true
	}
	return cache.actionCache[target][index]
}
