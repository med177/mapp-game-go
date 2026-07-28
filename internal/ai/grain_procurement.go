package ai

import (
	"sort"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

const (
	aiGrainReserveMonths   = 3
	aiGrainPurchaseWindow  = 2
	aiGrainPurchaseMinimum = 12
)

type aiGrainSupplier struct {
	id      faction.FactionID
	surplus int
}

// aiProcureGrain, aktif ticaret ağı üzerinden stratejik rezervi eksik AI
// fraksiyonuna tahıl alır. Kaynakta da güvenli rezerv bırakılır; böylece büyük
// devletler küçük devletleri satın alırken kendi ordularını kıtlığa sokmaz.
func aiProcureGrain(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || fid == "" {
		return 0
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return 0
	}

	totalDemand := aiFactionGrainDemand(gs, fid)
	if totalDemand <= 0 {
		return 0
	}
	capacity := gs.GrainStorageCapacityForFaction(fid)
	target := maxInt(capacity, totalDemand*aiGrainReserveMonths)
	shortfall := target - self.Grain
	if shortfall <= 0 {
		return 0
	}

	amount := minInt(shortfall, maxInt(aiGrainPurchaseMinimum, totalDemand*aiGrainPurchaseWindow))
	price := gs.MarketPrices[economy.GoodGrain]
	if price <= 0 {
		price = economy.BaseGoldValue[economy.GoodGrain]
	}
	if price <= 0 {
		return 0
	}
	if gs.MarketPrices == nil {
		gs.MarketPrices = make(map[economy.GoodType]int)
	}
	if gs.MarketPrices[economy.GoodGrain] <= 0 {
		gs.MarketPrices[economy.GoodGrain] = price
	}

	goldBudget := self.Gold - aiMinGoldReserve
	if goldBudget <= 0 {
		return 0
	}
	if oneThird := self.Gold / 3; oneThird > 0 && goldBudget > oneThird {
		goldBudget = oneThird
	}
	amount = minInt(amount, goldBudget/price)
	if amount <= 0 {
		return 0
	}

	connected := aiTradeNetworkMembers(gs, fid)
	suppliers := make([]aiGrainSupplier, 0, len(connected))
	for supplierID := range connected {
		if supplierID == fid {
			continue
		}
		supplier := gs.Factions[supplierID]
		if supplier == nil || supplier.IsEliminated {
			continue
		}
		surplus := gs.StrategicGrainSurplus(supplierID)
		if surplus <= 0 {
			continue
		}
		suppliers = append(suppliers, aiGrainSupplier{id: supplierID, surplus: surplus})
	}
	sort.Slice(suppliers, func(i, j int) bool {
		if suppliers[i].surplus != suppliers[j].surplus {
			return suppliers[i].surplus > suppliers[j].surplus
		}
		return suppliers[i].id < suppliers[j].id
	})

	purchased := 0
	for _, candidate := range suppliers {
		if purchased >= amount {
			break
		}
		buy := minInt(amount-purchased, candidate.surplus)
		if buy <= 0 || !economy.TransferGoods(gs.Factions, candidate.id, fid, economy.GoodGrain, buy, gs.MarketPrices) {
			continue
		}
		purchased += buy
	}
	return purchased
}

func aiFactionGrainDemand(gs *state.GameState, fid faction.FactionID) int {
	if status, ok := gs.GrainEconomy[fid]; ok && status.TotalDemand > 0 {
		return status.TotalDemand
	}
	demand := 0
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		demand += gs.CivilianGrainDemandForRegion(region)
	}
	for _, armyRef := range gs.Armies {
		if armyRef != nil && armyRef.OwnerID == string(fid) {
			demand += gs.EffectiveArmyGrainUpkeep(armyRef)
		}
	}
	return demand
}

// aiTradeNetworkMembers returns factions reachable through active trade
// routes. AI procurement follows the same connected-network rule as the
// player's one-time market trade, without requiring a direct bilateral route.
func aiTradeNetworkMembers(gs *state.GameState, fid faction.FactionID) map[faction.FactionID]struct{} {
	connected := map[faction.FactionID]struct{}{fid: {}}
	queue := []faction.FactionID{fid}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, route := range gs.TradeRoutes {
			if route == nil || route.SuspendedTurns > 0 {
				continue
			}
			var next faction.FactionID
			switch faction.FactionID(route.FromFactionID) {
			case current:
				next = faction.FactionID(route.ToFactionID)
			default:
				if faction.FactionID(route.ToFactionID) == current {
					next = faction.FactionID(route.FromFactionID)
				}
			}
			if next == "" {
				continue
			}
			if _, seen := connected[next]; seen {
				continue
			}
			connected[next] = struct{}{}
			queue = append(queue, next)
		}
	}
	return connected
}
