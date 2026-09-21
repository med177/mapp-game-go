package state

import (
	"sort"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

// ActiveHistoricalTradeFlows returns the scenario-defined historical flows
// active in the current year. The flow remains active independently of the
// survival of any faction; only the current owner of a real endpoint receives
// its income.
func (s *GameState) ActiveHistoricalTradeFlows() []world.HistoricalTradeFlow {
	if s == nil {
		return nil
	}
	if s.hasSourceGoods() {
		if s.historicalTradeFlowCacheValid && s.historicalTradeFlowCacheYear == s.Year {
			return s.historicalTradeFlowCache
		}
		flows := s.derivedHistoricalTradeFlows()
		s.historicalTradeFlowCacheYear = s.Year
		s.historicalTradeFlowCacheValid = true
		s.historicalTradeFlowCache = flows
		return flows
	}
	flows := make([]world.HistoricalTradeFlow, 0, len(s.TradeCenters.HistoricalFlows))
	for _, flow := range s.TradeCenters.HistoricalFlows {
		if flow.ActiveInYear(s.Year) && s.historicalFlowCentersActive(flow) {
			flows = append(flows, flow)
		}
	}
	for i := range flows {
		flows[i].AmountPerTurn = s.applyHistoricalAmountCompetition(flows[i])
		flows[i].GoldIncomePerTurn = s.applyHistoricalIncomeCompetition(flows[i])
	}
	return flows
}

func (s *GameState) hasSourceGoods() bool {
	for _, center := range s.TradeCenters.Centers {
		if len(center.SourceGoods) > 0 {
			return true
		}
	}
	return false
}

// InvalidateHistoricalTradeFlowCache forces the derived graph to be rebuilt
// after scenario/editor data changes without waiting for a new year.
func (s *GameState) InvalidateHistoricalTradeFlowCache() {
	if s == nil {
		return
	}
	s.historicalTradeFlowCacheValid = false
	s.historicalTradeFlowCache = nil
}

type historicalFlowNode struct {
	id              world.RegionID
	amount          int
	income          int
	distance        int
	source          bool
	mainRouteSource bool
}

type historicalOriginGood struct {
	good   world.HistoricalTradeGood
	source bool
}

// derivedHistoricalTradeFlows expands each source good over the trade-center
// topology. Source centers are one-way hubs: natural production cannot return
// into a source center, while a source good may pass through another source
// hub (for example, America -> Atlantic -> Portugal). Each source/good visits
// a center once, which makes cycles harmless while allowing the same edge to
// carry multiple goods.
func (s *GameState) derivedHistoricalTradeFlows() []world.HistoricalTradeFlow {
	adj := make(map[world.RegionID][]world.RegionID, len(s.TradeCenters.Centers))
	active := make(map[world.RegionID]bool, len(s.TradeCenters.Centers))
	for _, center := range s.TradeCenters.Centers {
		if !center.ActiveInYear(s.Year) {
			continue
		}
		active[center.ID] = true
	}
	// TradeAdjacency makes normal-center links bidirectional while preserving
	// the one-way direction of off-map sources. Rendererler raw Links listesini
	// kullanarak yalnız veri içinde yazılan görsel oku çizer.
	for id, links := range s.TradeCenters.TradeAdjacency() {
		if !active[id] {
			continue
		}
		for _, linked := range links {
			if active[linked] {
				adj[id] = appendUniqueHistoricalID(adj[id], linked)
			}
		}
		adj[id] = uniqueSortedHistoricalIDs(adj[id])
	}

	flows := make([]world.HistoricalTradeFlow, 0, len(active)*4)
	for _, center := range s.TradeCenters.Centers {
		if !active[center.ID] {
			continue
		}
		goods := make([]historicalOriginGood, 0, len(center.SourceGoods)+6)
		for _, good := range center.SourceGoods {
			goods = append(goods, historicalOriginGood{good: good, source: true})
		}
		for _, good := range s.regionHistoricalGoods(center.ID) {
			goods = append(goods, historicalOriginGood{good: good})
		}
		sort.Slice(goods, func(i, j int) bool {
			if goods[i].good.Good != goods[j].good.Good {
				return goods[i].good.Good < goods[j].good.Good
			}
			return goods[i].source && !goods[j].source
		})
		for _, origin := range goods {
			good := origin.good
			if good.AmountPerTurn <= 0 || good.Good == "" {
				continue
			}
			queue := []historicalFlowNode{{id: center.ID, amount: good.AmountPerTurn, income: good.GoldIncomePerTurn, source: origin.source, mainRouteSource: center.MainRoute}}
			visited := map[world.RegionID]bool{center.ID: true}
			for len(queue) > 0 {
				node := queue[0]
				queue = queue[1:]
				for _, linked := range adj[node.id] {
					if visited[linked] || node.amount <= 0 || isHistoricalFlowReturnToMainRoute(s.TradeCenters, linked, node) {
						continue
					}
					visited[linked] = true
					flows = append(flows, world.HistoricalTradeFlow{
						FromRegionID:      node.id,
						ToRegionID:        linked,
						Good:              good.Good,
						AmountPerTurn:     node.amount,
						GoldIncomePerTurn: node.income,
						SourceCenterID:    center.ID,
					})
					nextAmount := node.amount * 75 / 100
					nextIncome := node.income * 75 / 100
					if nextAmount > 0 {
						queue = append(queue, historicalFlowNode{id: linked, amount: nextAmount, income: nextIncome, distance: node.distance + 1, source: node.source, mainRouteSource: node.mainRouteSource})
					}
				}
			}
		}
	}
	for i := range flows {
		flows[i].AmountPerTurn = s.applyHistoricalAmountCompetition(flows[i])
		flows[i].GoldIncomePerTurn = s.applyHistoricalIncomeCompetition(flows[i])
	}
	return flows
}

func isHistoricalSourceCenter(config world.TradeCenterConfig, id world.RegionID) bool {
	for _, center := range config.Centers {
		if center.ID == id {
			// Scenario sources are normalized as off-map centers during load.
			// Keep SourceGoods as a fallback for synthetic/runtime fixtures.
			return center.OffMap || len(center.SourceGoods) > 0
		}
	}
	return false
}

func isHistoricalMainRoute(config world.TradeCenterConfig, id world.RegionID) bool {
	for _, center := range config.Centers {
		if center.ID == id {
			return center.MainRoute
		}
	}
	return false
}

func isHistoricalFlowReturnToMainRoute(config world.TradeCenterConfig, id world.RegionID, node historicalFlowNode) bool {
	if isHistoricalMainRoute(config, id) && !node.source {
		return true
	}
	if !isHistoricalSourceCenter(config, id) {
		return false
	}
	if !node.source {
		return true
	}
	return node.mainRouteSource && isHistoricalMainRoute(config, id)
}

func (s *GameState) regionHistoricalGoods(id world.RegionID) []world.HistoricalTradeGood {
	region := s.Regions[id]
	if region == nil || region.IsSea || region.IsTerrainArea {
		return nil
	}
	goods := make([]world.HistoricalTradeGood, 0, 6)
	appendOutput := func(good economy.GoodType, output int) {
		amount := output / 10
		if amount <= 0 && output > 0 {
			amount = 1
		}
		if amount > 0 {
			goods = append(goods, world.HistoricalTradeGood{Good: good, AmountPerTurn: amount})
		}
	}
	appendOutput(economy.GoodGrain, region.BaseGrainOutput)
	appendOutput(economy.GoodIron, region.BaseIronOutput)
	appendOutput(economy.GoodTimber, region.BaseTimberOutput)
	appendOutput(economy.GoodStone, region.BaseStoneOutput)
	appendOutput(economy.GoodSpice, region.BaseSpiceOutput)
	appendOutput(economy.GoodCloth, region.BaseClothOutput)
	return goods
}

func appendUniqueHistoricalID(ids []world.RegionID, id world.RegionID) []world.RegionID {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}

func uniqueSortedHistoricalIDs(ids []world.RegionID) []world.RegionID {
	result := append([]world.RegionID(nil), ids...)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	unique := result[:0]
	for _, id := range result {
		if len(unique) == 0 || unique[len(unique)-1] != id {
			unique = append(unique, id)
		}
	}
	return unique
}

// HistoricalTradeFlowAmount returns the active amount after competition from
// newly opened historical centers has reduced the old route's capacity.
func (s *GameState) HistoricalTradeFlowAmount(flow world.HistoricalTradeFlow) int {
	if flow.SourceCenterID != "" {
		return flow.AmountPerTurn
	}
	amount := flow.AmountPerTurn
	if amount <= 0 {
		return 0
	}
	return s.applyHistoricalAmountCompetitionValue(flow, amount)
}

func (s *GameState) applyHistoricalAmountCompetition(flow world.HistoricalTradeFlow) int {
	return s.applyHistoricalAmountCompetitionValue(flow, flow.AmountPerTurn)
}

func (s *GameState) applyHistoricalAmountCompetitionValue(flow world.HistoricalTradeFlow, amount int) int {
	for _, center := range s.TradeCenters.Centers {
		if !center.ActiveInYear(s.Year) {
			continue
		}
		for _, impact := range center.CompetitionImpacts {
			if impact.CenterID != flow.FromRegionID && impact.CenterID != flow.ToRegionID && impact.CenterID != flow.SourceCenterID {
				continue
			}
			amount = applyHistoricalPercent(amount, impact.AmountPercent)
		}
	}
	return amount
}

// HistoricalTradeFlowIncome returns the active gold income after competition
// from newly opened historical centers has reduced the old route's return.
func (s *GameState) HistoricalTradeFlowIncome(flow world.HistoricalTradeFlow) int {
	if flow.SourceCenterID != "" {
		return flow.GoldIncomePerTurn
	}
	income := flow.GoldIncomePerTurn
	if income <= 0 {
		return 0
	}
	return s.applyHistoricalIncomeCompetitionValue(flow, income)
}

func (s *GameState) applyHistoricalIncomeCompetition(flow world.HistoricalTradeFlow) int {
	return s.applyHistoricalIncomeCompetitionValue(flow, flow.GoldIncomePerTurn)
}

func (s *GameState) applyHistoricalIncomeCompetitionValue(flow world.HistoricalTradeFlow, income int) int {
	for _, center := range s.TradeCenters.Centers {
		if !center.ActiveInYear(s.Year) {
			continue
		}
		for _, impact := range center.CompetitionImpacts {
			if impact.CenterID == flow.FromRegionID || impact.CenterID == flow.ToRegionID || impact.CenterID == flow.SourceCenterID {
				income = applyHistoricalPercent(income, impact.IncomePercent)
			}
		}
	}
	return income
}

func applyHistoricalPercent(value, percent int) int {
	if percent == 0 {
		return value
	}
	if percent < -100 {
		percent = -100
	}
	if percent > 200 {
		percent = 200
	}
	value = value * (100 + percent) / 100
	if value < 0 {
		return 0
	}
	return value
}

func (s *GameState) historicalFlowCentersActive(flow world.HistoricalTradeFlow) bool {
	for _, id := range []world.RegionID{flow.FromRegionID, flow.ToRegionID} {
		for _, center := range s.TradeCenters.Centers {
			if center.ID == id {
				if !center.ActiveInYear(s.Year) {
					return false
				}
				break
			}
		}
	}
	return true
}

// HistoricalTradeIncomeForFaction returns the gold generated by active
// historical flows for the current owner of their real endpoints. A flow
// touching one real endpoint pays that endpoint in full; a flow between two
// real endpoints is split evenly, with the remainder assigned deterministically
// to the first endpoint.
func (s *GameState) HistoricalTradeIncomeForFaction(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	income := 0
	for _, flow := range s.ActiveHistoricalTradeFlows() {
		incomePerTurn := s.HistoricalTradeFlowIncome(flow)
		if incomePerTurn <= 0 {
			continue
		}
		from := s.Regions[flow.FromRegionID]
		to := s.Regions[flow.ToRegionID]
		fromOwner := historicalEndpointOwner(s, from)
		toOwner := historicalEndpointOwner(s, to)
		if fromOwner == fid && toOwner == fid {
			income += incomePerTurn
			continue
		}
		if fromOwner != "" && toOwner != "" {
			share := incomePerTurn / 2
			if incomePerTurn%2 != 0 && string(flow.FromRegionID) < string(flow.ToRegionID) {
				share++
			}
			if fromOwner == fid || toOwner == fid {
				income += share
			}
			continue
		}
		if fromOwner == fid || toOwner == fid {
			income += incomePerTurn
		}
	}
	return income
}

func historicalEndpointOwner(s *GameState, region *world.Region) faction.FactionID {
	if s == nil || region == nil || region.IsSea || region.IsTerrainArea || region.OwnerID == "" {
		return ""
	}
	fid := faction.FactionID(region.OwnerID)
	if owner := s.Factions[fid]; owner == nil || owner.IsEliminated {
		return ""
	}
	return fid
}
